package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
	_ "modernc.org/sqlite"
)

var timezone *time.Location

func init() {
	// For now, hardcode to America/Los_Angeles to ensure consistency
	// This can be made configurable later once we fix the timezone loading issues
	var err error
	timezone, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Warning: Failed to load America/Los_Angeles timezone, using UTC: %v", err)
		timezone = time.UTC
	}
	
	log.Printf("Using timezone: %s", timezone.String())
	
	// Debug timezone information
	now := time.Now()
	log.Printf("Current UTC time: %s", now.UTC().Format("2006-01-02 15:04:05"))
	log.Printf("Current time in configured timezone: %s", now.In(timezone).Format("2006-01-02 15:04:05"))
	log.Printf("Current time in local timezone: %s", now.Local().Format("2006-01-02 15:04:05"))
	log.Printf("Timezone offset from UTC: %s", now.In(timezone).Format("-07:00"))
}

//go:embed static/*
var rawFSFromEmbed embed.FS

// spaFileSystem wraps an http.FileSystem to implement SPA routing.
// If a requested file is not found, it serves 'index.html' instead.
type spaFileSystem struct {
	contentRoot http.FileSystem
}

// Open implements the http.FileSystem interface.
func (sfs spaFileSystem) Open(name string) (http.File, error) {
	f, err := sfs.contentRoot.Open(name)
	// If the file exists, serve it.
	if err == nil {
		return f, nil
	}
	// If the file does not exist, check if it's a static asset
	if os.IsNotExist(err) {
		// If it's a static asset (JS, CSS, etc.), return the error
		if isStaticAsset(name) {
			return nil, err
		}
		// For other routes, this is the SPA fallback case.
		// Serve the index.html from the root of the content filesystem.
		log.Printf("SPA Fallback: Requested path '%s' not found. Serving 'index.html'.", name)
		return sfs.contentRoot.Open("index.html")
	}
	// For any other errors, return them.
	return nil, err
}

// isStaticAsset checks if the requested path is a static asset
func isStaticAsset(name string) bool {
	staticExtensions := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	// Also check for paths that start with _app (SvelteKit assets)
	if strings.HasPrefix(name, "_app/") {
		return true
	}
	return false
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Completed   bool      `json:"completed"`
	DayOfWeek   string    `json:"dayOfWeek"`
	WeekDate    string    `json:"weekDate"`    // ISO date string for the week (Sunday of the week)
	Tags        string    `json:"tags"`        // Comma-separated tags
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateTaskRequest struct {
	Title     string `json:"title"`
	DayOfWeek string `json:"dayOfWeek"`
	WeekDate  string `json:"weekDate"`  // ISO date string for the week (Sunday of the week)
	Tags      string `json:"tags"`      // Comma-separated tags
}

type UpdateTaskRequest struct {
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	DayOfWeek string `json:"dayOfWeek"`
	WeekDate  string `json:"weekDate"`  // ISO date string for the week (Sunday of the week)
	Tags      string `json:"tags"`      // Comma-separated tags
}

var db *sql.DB

func main() {
	log.Println("=== Starting Zendo server ===")
	log.Println("Time:", time.Now().Format("2006-01-02 15:04:05"))

	// Create storage directory if it doesn't exist
	err := os.MkdirAll("./storage", 0755)
	if err != nil {
		log.Fatal("Failed to create storage directory:", err)
	}

	// Initialize database
	db, err = sql.Open("sqlite", "./storage/zendo.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	log.Println("Database connection established successfully")

	// Create tasks table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed BOOLEAN DEFAULT FALSE,
		day_of_week TEXT NOT NULL,
		week_date TEXT NOT NULL,
		tags TEXT, -- Comma-separated tags
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create tasks table:", err)
	}

	log.Println("Tasks table created/verified successfully")

	// Run migration to add week_date column if it doesn't exist
	err = runMigration()
	if err != nil {
		log.Fatal("Failed to run migration:", err)
	}

	// --- Frontend File Server Setup ---
	// Create an fs.FS that is rooted at the "static" directory
	// within the raw embedded filesystem. This makes 'index.html' available at the root.
	contentFS, err := fs.Sub(rawFSFromEmbed, "static")
	if err != nil {
		log.Fatalf("Failed to create sub FS for embedded assets: %v", err)
	}

	// Wrap the correctly rooted contentFS with our SPA handler logic.
	spaFS := spaFileSystem{contentRoot: http.FS(contentFS)}
	fileServer := http.FileServer(spaFS)

	// --- HTTP Route Handling ---
	// Create a master router for the application.
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/tasks", getTasks)
	mux.HandleFunc("GET /api/tasks/week/{weekDate}", getTasksForWeek)
	mux.HandleFunc("GET /api/tasks/today", getTasksForToday)
	mux.HandleFunc("GET /api/tasks/today/week", getTasksForTodayWeek)
	mux.HandleFunc("POST /api/tasks", createTask)
	mux.HandleFunc("PUT /api/tasks/{id}", updateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", deleteTask)
	mux.HandleFunc("GET /api/debug/timezone", debugTimezone)
	mux.HandleFunc("GET /api/timezone", getTimezoneInfo)
	mux.HandleFunc("GET /api/debug/timezones", listTimezones)

	// The root handler serves the frontend SPA.
	// This must be registered after all other routes to act as a catch-all.
	mux.Handle("/", fileServer)

	// --- CORS and Server Setup ---
	// Setup CORS middleware
	// Get custom app URL from environment variable
	customAppURL := os.Getenv("APP_URL")
	
	// Build allowed origins list
	allowedOrigins := []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:8080"}
	if customAppURL != "" {
		allowedOrigins = append(allowedOrigins, customAppURL)
		log.Printf("Added custom app URL to CORS: %s", customAppURL)
	}
	
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-Requested-With", "Upgrade", "Connection"},
		AllowCredentials: true,
	})

	// Apply CORS for all routes
	handler := c.Handler(mux)

	log.Println("Server starting on http://localhost:8080")
	log.Println("CORS allowed origins:", allowedOrigins)
	log.Println("API endpoints available:")
	log.Println("  GET  /api/tasks")
	log.Println("  GET  /api/tasks/week/{weekDate}")
	log.Println("  GET  /api/tasks/today")
	log.Println("  GET  /api/tasks/today/week")
	log.Println("  POST /api/tasks")
	log.Println("  PUT  /api/tasks/{id}")
	log.Println("  DELETE /api/tasks/{id}")
	log.Println("=== Server ready ===")
	
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/tasks - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())
	
	startTime := time.Now()
	
	rows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks ORDER BY day_of_week, created_at")
	if err != nil {
		log.Printf("ERROR: Database query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	taskCount := 0
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			log.Printf("ERROR: Row scan failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
		taskCount++
		log.Printf("  Task %d: %s (completed: %v, day: %s)", task.ID, task.Title, task.Completed, task.DayOfWeek)
	}

	// Always return an array, even if empty
	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	
	duration := time.Since(startTime)
	log.Printf("=== GET /api/tasks - Response sent ===")
	log.Printf("Returned %d tasks in %v", taskCount, duration)
}

func getTasksForWeek(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/tasks/week - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())
	
	startTime := time.Now()
	
	// Extract weekDate from URL path
	path := r.URL.Path
	weekDate := path[len("/api/tasks/week/"):]
	
	log.Printf("Fetching tasks for week: %s", weekDate)
	
	rows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE week_date = ? ORDER BY day_of_week, created_at", weekDate)
	if err != nil {
		log.Printf("ERROR: Database query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	taskCount := 0
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			log.Printf("ERROR: Row scan failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
		taskCount++
		log.Printf("  Task %d: %s (completed: %v, day: %s, week: %s)", task.ID, task.Title, task.Completed, task.DayOfWeek, task.WeekDate)
	}

	// Always return an array, even if empty
	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	
	duration := time.Since(startTime)
	log.Printf("=== GET /api/tasks/week - Response sent ===")
	log.Printf("Returned %d tasks for week %s in %v", taskCount, weekDate, duration)
}

func getTasksForToday(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/tasks/today - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())
	
	startTime := time.Now()
	
	// Get today's date in YYYY-MM-DD format in the configured timezone
	now := time.Now().In(timezone)
	today := now.Format("2006-01-02")
	todayWeekStart := getWeekStart(now).Format("2006-01-02")
	
	log.Printf("Server timezone: %s", timezone.String())
	log.Printf("Current UTC time: %s", time.Now().UTC().Format("2006-01-02 15:04:05"))
	log.Printf("Current time in configured timezone: %s", now.Format("2006-01-02 15:04:05"))
	log.Printf("Fetching tasks for today: %s (week: %s)", today, todayWeekStart)
	
	// Get today's day of week (convert to lowercase to match frontend)
	todayDayOfWeek := strings.ToLower(now.Weekday().String())
	
	log.Printf("Today's day of week: %s", todayDayOfWeek)
	
	// Debug: Let's see what tasks exist in the database
	var allTasks []Task
	debugRows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks ORDER BY week_date, day_of_week")
	if err == nil {
		defer debugRows.Close()
		for debugRows.Next() {
			var task Task
			if err := debugRows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt); err == nil {
				allTasks = append(allTasks, task)
				log.Printf("  DB Task %d: %s (day: %s, week: %s)", task.ID, task.Title, task.DayOfWeek, task.WeekDate)
			}
		}
	}
	
	rows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE week_date = ? AND day_of_week = ? ORDER BY created_at", todayWeekStart, todayDayOfWeek)
	if err != nil {
		log.Printf("ERROR: Database query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	taskCount := 0
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			log.Printf("ERROR: Row scan failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
		taskCount++
		log.Printf("  Task %d: %s (completed: %v, day: %s, week: %s)", task.ID, task.Title, task.Completed, task.DayOfWeek, task.WeekDate)
	}

	// Always return an array, even if empty
	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	
	duration := time.Since(startTime)
	log.Printf("=== GET /api/tasks/today - Response sent ===")
	log.Printf("Returned %d tasks for today in %v", taskCount, duration)
}

func getTasksForTodayWeek(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/tasks/today/week - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())
	
	startTime := time.Now()
	
	// Get today's week start date in the configured timezone
	now := time.Now().In(timezone)
	todayWeekStart := getWeekStart(now).Format("2006-01-02")
	
	log.Printf("Server timezone: %s", timezone.String())
	log.Printf("Current UTC time: %s", time.Now().UTC().Format("2006-01-02 15:04:05"))
	log.Printf("Current time in configured timezone: %s", now.Format("2006-01-02 15:04:05"))
	log.Printf("Fetching tasks for today's week: %s", todayWeekStart)
	
	// Debug: Let's see what tasks exist in the database
	var allTasks []Task
	debugRows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks ORDER BY week_date, day_of_week")
	if err == nil {
		defer debugRows.Close()
		for debugRows.Next() {
			var task Task
			if err := debugRows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt); err == nil {
				allTasks = append(allTasks, task)
				log.Printf("  DB Task %d: %s (day: %s, week: %s)", task.ID, task.Title, task.DayOfWeek, task.WeekDate)
			}
		}
	}
	
	rows, err := db.Query("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE week_date = ? ORDER BY day_of_week, created_at", todayWeekStart)
	if err != nil {
		log.Printf("ERROR: Database query failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	taskCount := 0
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			log.Printf("ERROR: Row scan failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
		taskCount++
		log.Printf("  Task %d: %s (completed: %v, day: %s, week: %s)", task.ID, task.Title, task.Completed, task.DayOfWeek, task.WeekDate)
	}

	// Always return an array, even if empty
	if tasks == nil {
		tasks = []Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
	
	duration := time.Since(startTime)
	log.Printf("=== GET /api/tasks/today/week - Response sent ===")
	log.Printf("Returned %d tasks for today's week in %v", taskCount, duration)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== POST /api/tasks - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
	
	startTime := time.Now()
	
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: JSON decode failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request body: title='%s', dayOfWeek='%s', weekDate='%s', tags='%s'", req.Title, req.DayOfWeek, req.WeekDate, req.Tags)

	if req.Title == "" || req.DayOfWeek == "" || req.WeekDate == "" {
		log.Printf("ERROR: Missing required fields")
		http.Error(w, "Title, dayOfWeek, and weekDate are required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO tasks (title, day_of_week, week_date, tags) VALUES (?, ?, ?, ?)", req.Title, req.DayOfWeek, req.WeekDate, req.Tags)
	if err != nil {
		log.Printf("ERROR: Database insert failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("ERROR: Failed to get last insert ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Task inserted with ID: %d", id)

	// Fetch the created task
	var task Task
	err = db.QueryRow("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		log.Printf("ERROR: Failed to fetch created task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Created task: ID=%d, Title='%s', Completed=%v, DayOfWeek='%s', WeekDate='%s', Tags='%s'", 
		task.ID, task.Title, task.Completed, task.DayOfWeek, task.WeekDate, task.Tags)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
	
	duration := time.Since(startTime)
	log.Printf("=== POST /api/tasks - Response sent ===")
	log.Printf("Task created successfully in %v", duration)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== PUT /api/tasks - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
	
	startTime := time.Now()
	
	// Extract ID from URL path
	path := r.URL.Path
	idStr := path[len("/api/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("ERROR: Invalid task ID '%s': %v", idStr, err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	log.Printf("Updating task ID: %d", id)

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: JSON decode failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Request body: title='%s', completed=%v, dayOfWeek='%s', weekDate='%s', tags='%s'", 
		req.Title, req.Completed, req.DayOfWeek, req.WeekDate, req.Tags)

	// First, let's check what the current state is
	var currentTask Task
	err = db.QueryRow("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE id = ?", id).
		Scan(&currentTask.ID, &currentTask.Title, &currentTask.Completed, &currentTask.DayOfWeek, &currentTask.WeekDate, &currentTask.Tags, &currentTask.CreatedAt, &currentTask.UpdatedAt)
	if err != nil {
		log.Printf("ERROR: Failed to fetch current task state: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Current task state: ID=%d, Title='%s', Completed=%v, DayOfWeek='%s', WeekDate='%s', Tags='%s'", 
		currentTask.ID, currentTask.Title, currentTask.Completed, currentTask.DayOfWeek, currentTask.WeekDate, currentTask.Tags)

	// Perform the update
	_, err = db.Exec("UPDATE tasks SET title = ?, completed = ?, day_of_week = ?, week_date = ?, tags = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", 
		req.Title, req.Completed, req.DayOfWeek, req.WeekDate, req.Tags, id)
	if err != nil {
		log.Printf("ERROR: Database update failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Database update completed successfully")

	// Fetch the updated task
	var task Task
	err = db.QueryRow("SELECT id, title, completed, day_of_week, week_date, tags, created_at, updated_at FROM tasks WHERE id = ?", id).
		Scan(&task.ID, &task.Title, &task.Completed, &task.DayOfWeek, &task.WeekDate, &task.Tags, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		log.Printf("ERROR: Failed to fetch updated task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Updated task: ID=%d, Title='%s', Completed=%v, DayOfWeek='%s', WeekDate='%s', Tags='%s'", 
		task.ID, task.Title, task.Completed, task.DayOfWeek, task.WeekDate, task.Tags)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
	
	duration := time.Since(startTime)
	log.Printf("=== PUT /api/tasks - Response sent ===")
	log.Printf("Task updated successfully in %v", duration)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== DELETE /api/tasks - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	
	startTime := time.Now()
	
	// Extract ID from URL path
	path := r.URL.Path
	idStr := path[len("/api/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("ERROR: Invalid task ID '%s': %v", idStr, err)
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	log.Printf("Deleting task ID: %d", id)

	result, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		log.Printf("ERROR: Database delete failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("ERROR: Failed to get rows affected: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("ERROR: Task not found (ID: %d)", id)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	log.Printf("Task deleted successfully (ID: %d, rows affected: %d)", id, rowsAffected)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
	
	duration := time.Since(startTime)
	log.Printf("=== DELETE /api/tasks - Response sent ===")
	log.Printf("Task deleted successfully in %v", duration)
}

// runMigration handles database schema migrations
func runMigration() error {
	log.Println("=== Running Database Migration ===")

	// Check if week_date column exists
	var columnExists int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='week_date'").Scan(&columnExists)
	if err != nil {
		return err
	}

	if columnExists == 0 {
		log.Println("Adding week_date column to tasks table...")
		
		// Add the new column
		_, err = db.Exec("ALTER TABLE tasks ADD COLUMN week_date TEXT")
		if err != nil {
			return err
		}

		// Update existing tasks with default week date (current week)
		now := time.Now().In(timezone)
		weekStart := getWeekStart(now)
		weekDate := weekStart.Format("2006-01-02")
		
		log.Printf("Updating existing tasks with default week date: %s", weekDate)
		
		_, err = db.Exec("UPDATE tasks SET week_date = ? WHERE week_date IS NULL", weekDate)
		if err != nil {
			return err
		}

		log.Println("Migration completed successfully!")
	} else {
		log.Println("week_date column already exists. No migration needed.")
	}

	// Check if tags column exists
	var tagsColumnExists int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='tags'").Scan(&tagsColumnExists)
	if err != nil {
		return err
	}

	if tagsColumnExists == 0 {
		log.Println("Adding tags column to tasks table...")
		
		// Add the new column
		_, err = db.Exec("ALTER TABLE tasks ADD COLUMN tags TEXT")
		if err != nil {
			return err
		}

		// Update existing tasks with default tags (empty string)
		log.Printf("Updating existing tasks with default tags: ''")
		
		_, err = db.Exec("UPDATE tasks SET tags = '' WHERE tags IS NULL")
		if err != nil {
			return err
		}

		log.Println("Migration completed successfully!")
	} else {
		log.Println("tags column already exists. No migration needed.")
	}

	return nil
}

// getWeekStart returns the Sunday of the current week
func getWeekStart(date time.Time) time.Time {
	weekday := date.Weekday()
	return date.AddDate(0, 0, -int(weekday))
}

func debugTimezone(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/debug/timezone - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())

	startTime := time.Now()

	now := time.Now()
	
	log.Printf("Current UTC time: %s", now.UTC().Format("2006-01-02 15:04:05"))
	log.Printf("Current time in configured timezone: %s", now.In(timezone).Format("2006-01-02 15:04:05"))
	log.Printf("Current time in local timezone: %s", now.Local().Format("2006-01-02 15:04:05"))
	log.Printf("Timezone offset from UTC: %s", now.In(timezone).Format("-07:00"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"current_utc_time": now.UTC().Format("2006-01-02 15:04:05"),
		"current_local_time": now.In(timezone).Format("2006-01-02 15:04:05"),
		"timezone_offset": now.In(timezone).Format("-07:00"),
	})

	duration := time.Since(startTime)
	log.Printf("=== GET /api/debug/timezone - Response sent ===")
	log.Printf("Returned debug info in %v", duration)
}

func getTimezoneInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/timezone - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())

	startTime := time.Now()

	now := time.Now()
	
	log.Printf("Current UTC time: %s", now.UTC().Format("2006-01-02 15:04:05"))
	log.Printf("Current time in configured timezone: %s", now.In(timezone).Format("2006-01-02 15:04:05"))
	log.Printf("Timezone offset from UTC: %s", now.In(timezone).Format("-07:00"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"current_utc_time": now.UTC().Format("2006-01-02 15:04:05"),
		"current_local_time": now.In(timezone).Format("2006-01-02 15:04:05"),
		"timezone_offset": now.In(timezone).Format("-07:00"),
	})

	duration := time.Since(startTime)
	log.Printf("=== GET /api/timezone - Response sent ===")
	log.Printf("Returned timezone info in %v", duration)
}

func listTimezones(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== GET /api/debug/timezones - Request received ===")
	log.Printf("Request from: %s", r.RemoteAddr)
	log.Printf("User-Agent: %s", r.UserAgent())

	startTime := time.Now()

	timezones := time.Now().Location().String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"timezones": timezones})

	duration := time.Since(startTime)
	log.Printf("=== GET /api/debug/timezones - Response sent ===")
	log.Printf("Returned timezones in %v", duration)
}