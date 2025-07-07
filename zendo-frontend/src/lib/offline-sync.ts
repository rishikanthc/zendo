// Simple Offline Data Manager - Focus on caching only
interface CachedTask {
  id: number;
  title: string;
  completed: boolean;
  dayOfWeek: string;
  weekDate: string;
  createdAt: string;
  updatedAt: string;
}

class OfflineDataManager {
  private db: IDBDatabase | null = null;
  private isOnline = navigator.onLine;

  constructor() {
    this.initDatabase();
    this.setupNetworkListeners();
  }

  private async initDatabase() {
    return new Promise<void>((resolve, reject) => {
      const request = indexedDB.open('ZendoOfflineData', 1);
      
      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };
      
      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        
        // Create cached tasks store
        if (!db.objectStoreNames.contains('cachedTasks')) {
          const store = db.createObjectStore('cachedTasks', { keyPath: 'id' });
          store.createIndex('weekDate', 'weekDate', { unique: false });
          store.createIndex('dayOfWeek', 'dayOfWeek', { unique: false });
        }
      };
    });
  }

  private setupNetworkListeners() {
    window.addEventListener('online', () => {
      this.isOnline = true;
    });

    window.addEventListener('offline', () => {
      this.isOnline = false;
    });
  }

  // Cache tasks for offline access
  async cacheTasks(tasks: CachedTask[]): Promise<void> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readwrite');
      const store = transaction.objectStore('cachedTasks');
      
      // Clear existing cached tasks
      const clearRequest = store.clear();
      clearRequest.onsuccess = () => {
        // Add new tasks
        let completed = 0;
        const total = tasks.length;
        
        if (total === 0) {
          resolve();
          return;
        }
        
        tasks.forEach(task => {
          const request = store.add(task);
          request.onsuccess = () => {
            completed++;
            if (completed === total) resolve();
          };
          request.onerror = () => reject(request.error);
        });
      };
      clearRequest.onerror = () => reject(clearRequest.error);
    });
  }

  // Get all cached tasks
  async getCachedTasks(): Promise<CachedTask[]> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readonly');
      const store = transaction.objectStore('cachedTasks');
      const request = store.getAll();
      
      request.onsuccess = () => resolve(request.result || []);
      request.onerror = () => reject(request.error);
    });
  }

  // Get cached tasks by week date
  async getCachedTasksByWeek(weekDate: string): Promise<CachedTask[]> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readonly');
      const store = transaction.objectStore('cachedTasks');
      const index = store.index('weekDate');
      const request = index.getAll(weekDate);
      
      request.onsuccess = () => resolve(request.result || []);
      request.onerror = () => reject(request.error);
    });
  }

  // Add a single task to cache
  async addCachedTask(task: CachedTask): Promise<void> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readwrite');
      const store = transaction.objectStore('cachedTasks');
      const request = store.add(task);
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  // Update a cached task
  async updateCachedTask(task: CachedTask): Promise<void> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readwrite');
      const store = transaction.objectStore('cachedTasks');
      const request = store.put(task);
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  // Delete a cached task
  async deleteCachedTask(taskId: number): Promise<void> {
    if (!this.db) await this.initDatabase();
    
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction(['cachedTasks'], 'readwrite');
      const store = transaction.objectStore('cachedTasks');
      const request = store.delete(taskId);
      
      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  // Check if we're online
  get isOnlineStatus(): boolean {
    return this.isOnline;
  }

  // Simple fetch wrapper that caches successful responses
  async fetchWithCaching(url: string, options: RequestInit = {}): Promise<Response> {
    try {
      const response = await fetch(url, options);
      
      // Cache successful GET responses for offline access
      if (response.ok && options.method === 'GET' && url.includes('/api/tasks')) {
        try {
          const data = await response.clone().json();
          if (Array.isArray(data)) {
            await this.cacheTasks(data);
          }
        } catch (e) {
          // Ignore caching errors
          console.warn('Failed to cache tasks:', e);
        }
      }
      
      return response;
    } catch (error) {
      // Re-throw the error for handling by the caller
      throw error;
    }
  }
}

// Export singleton instance
export const offlineData = new OfflineDataManager(); 