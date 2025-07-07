import { writable } from 'svelte/store';

export interface Task {
  id: number;
  title: string;
  completed: boolean;
  dayOfWeek: string;
  weekDate: string;
  tags: string; // Comma-separated tags
  createdAt: string;
  updatedAt: string;
}

const API_BASE = typeof window !== 'undefined'
  ? `${window.location.origin}/api`
  : 'http://localhost:8080/api';

// IndexedDB helpers
const DB_NAME = 'ZendoOfflineData';
const STORE_NAME = 'cachedTasks';

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, 1);
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME, { keyPath: 'id' });
      }
    };
  });
}

async function getAllFromDB(): Promise<Task[]> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_NAME], 'readonly');
    const store = tx.objectStore(STORE_NAME);
    const req = store.getAll();
    req.onsuccess = () => resolve(req.result as Task[]);
    req.onerror = () => reject(req.error);
  });
}

async function setAllToDB(tasks: Task[]): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_NAME], 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const clearReq = store.clear();
    clearReq.onsuccess = () => {
      let completed = 0;
      if (tasks.length === 0) return resolve();
      tasks.forEach(task => {
        const addReq = store.add(task);
        addReq.onsuccess = () => {
          completed++;
          if (completed === tasks.length) resolve();
        };
        addReq.onerror = () => reject(addReq.error);
      });
    };
    clearReq.onerror = () => reject(clearReq.error);
  });
}

async function upsertToDB(task: Task): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_NAME], 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.put(task);
    req.onsuccess = () => resolve();
    req.onerror = () => reject(req.error);
  });
}

async function deleteFromDB(id: number): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction([STORE_NAME], 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.delete(id);
    req.onsuccess = () => resolve();
    req.onerror = () => reject(req.error);
  });
}

function createTaskStore() {
  const { subscribe, set, update } = writable<Task[]>([]);

  // Load from IndexedDB on startup
  async function loadFromDB() {
    const tasks = await getAllFromDB();
    set(tasks);
  }

  // Fetch from API and update both store and DB
  async function fetchFromAPI() {
    try {
      const res = await fetch(`${API_BASE}/tasks`);
      if (!res.ok) throw new Error('Failed to fetch tasks from API');
      const tasks: Task[] = await res.json();
      set(tasks);
      await setAllToDB(tasks);
    } catch (e) {
      // If API fails, just use local data
      await loadFromDB();
    }
  }

  // On startup: load from DB, then try API if online
  if (typeof window !== 'undefined') {
    loadFromDB().then(() => {
      if (navigator.onLine) fetchFromAPI();
    });
    window.addEventListener('online', fetchFromAPI);
  }

  // CRUD methods
  return {
    subscribe,
    fetchFromAPI,
    async add(task: Omit<Task, 'id' | 'createdAt' | 'updatedAt'>) {
      // Try API first if online
      if (navigator.onLine) {
        const res = await fetch(`${API_BASE}/tasks`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(task)
        });
        if (!res.ok) throw new Error('Failed to create task');
        const newTask: Task = await res.json();
        update(tasks => [...tasks, newTask]);
        await upsertToDB(newTask);
        return newTask;
      } else {
        // Offline: create a local-only task with a negative ID
        const now = new Date().toISOString();
        const tempTask: Task = {
          ...task,
          id: Date.now() * -1,
          createdAt: now,
          updatedAt: now,
          completed: false
        };
        update(tasks => [...tasks, tempTask]);
        await upsertToDB(tempTask);
        return tempTask;
      }
    },
    async updateTask(task: Task) {
      if (navigator.onLine && task.id > 0) {
        const res = await fetch(`${API_BASE}/tasks/${task.id}`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(task)
        });
        if (!res.ok) throw new Error('Failed to update task');
        const updated: Task = await res.json();
        update(tasks => tasks.map(t => t.id === updated.id ? updated : t));
        await upsertToDB(updated);
        return updated;
      } else {
        // Offline or temp task
        const now = new Date().toISOString();
        const updated = { ...task, updatedAt: now };
        update(tasks => tasks.map(t => t.id === updated.id ? updated : t));
        await upsertToDB(updated);
        return updated;
      }
    },
    async deleteTask(id: number) {
      if (navigator.onLine && id > 0) {
        const res = await fetch(`${API_BASE}/tasks/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('Failed to delete task');
      }
      update(tasks => tasks.filter(t => t.id !== id));
      await deleteFromDB(id);
    },
    // Forcibly reload from API (e.g. after reconnect)
    reload: fetchFromAPI,
    // Forcibly reload from DB
    reloadFromDB: loadFromDB
  };
}

export const taskStore = createTaskStore(); 