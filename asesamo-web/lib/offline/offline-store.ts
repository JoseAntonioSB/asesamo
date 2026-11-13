/**
 * ASESAMO — Offline Outbox & IndexedDB Store
 * Location: asesamo-web/lib/offline/offline-store.ts
 */

export interface OfflineQueueItem {
  id: string;
  endpoint: string;
  method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  payload: any;
  created_at: string;
  retry_count: number;
  description: string;
}

export interface IOfflineStore {
  enqueue(item: OfflineQueueItem): Promise<void>;
  getPendingItems(): Promise<OfflineQueueItem[]>;
  removeItem(id: string): Promise<void>;
  incrementRetry(id: string): Promise<void>;
  clear(): Promise<void>;
}

/**
 * In-Memory fallback implementation (ideal for unit testing and SSR environments)
 */
export class InMemoryOfflineStore implements IOfflineStore {
  private items: Map<string, OfflineQueueItem> = new Map();

  public async enqueue(item: OfflineQueueItem): Promise<void> {
    this.items.set(item.id, { ...item });
  }

  public async getPendingItems(): Promise<OfflineQueueItem[]> {
    return Array.from(this.items.values()).sort(
      (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
    );
  }

  public async removeItem(id: string): Promise<void> {
    this.items.delete(id);
  }

  public async incrementRetry(id: string): Promise<void> {
    const item = this.items.get(id);
    if (item) {
      item.retry_count += 1;
      this.items.set(id, item);
    }
  }

  public async clear(): Promise<void> {
    this.items.clear();
  }
}

const DB_NAME = 'asesamo_offline_db';
const DB_VERSION = 1;
const STORE_NAME = 'sync_outbox';

/**
 * IndexedDB implementation for persistent browser client outbox storage
 */
export class IndexedDBOfflineStore implements IOfflineStore {
  private dbPromise: Promise<IDBDatabase> | null = null;

  private getDB(): Promise<IDBDatabase> {
    if (this.dbPromise) return this.dbPromise;

    this.dbPromise = new Promise((resolve, reject) => {
      if (typeof window === 'undefined' || !window.indexedDB) {
        return reject(new Error('IndexedDB no está disponible en este entorno'));
      }

      const req = window.indexedDB.open(DB_NAME, DB_VERSION);

      req.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          db.createObjectStore(STORE_NAME, { keyPath: 'id' });
        }
      };

      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });

    return this.dbPromise;
  }

  public async enqueue(item: OfflineQueueItem): Promise<void> {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const req = store.put(item);

      req.onsuccess = () => resolve();
      req.onerror = () => reject(req.error);
    });
  }

  public async getPendingItems(): Promise<OfflineQueueItem[]> {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readonly');
      const store = tx.objectStore(STORE_NAME);
      const req = store.getAll();

      req.onsuccess = () => {
        const items = (req.result as OfflineQueueItem[]) || [];
        items.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
        resolve(items);
      };
      req.onerror = () => reject(req.error);
    });
  }

  public async removeItem(id: string): Promise<void> {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const req = store.delete(id);

      req.onsuccess = () => resolve();
      req.onerror = () => reject(req.error);
    });
  }

  public async incrementRetry(id: string): Promise<void> {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const getReq = store.get(id);

      getReq.onsuccess = () => {
        const item = getReq.result as OfflineQueueItem | undefined;
        if (item) {
          item.retry_count = (item.retry_count || 0) + 1;
          const putReq = store.put(item);
          putReq.onsuccess = () => resolve();
          putReq.onerror = () => reject(putReq.error);
        } else {
          resolve();
        }
      };
      getReq.onerror = () => reject(getReq.error);
    });
  }

  public async clear(): Promise<void> {
    const db = await this.getDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const req = store.clear();

      req.onsuccess = () => resolve();
      req.onerror = () => reject(req.error);
    });
  }
}

/**
 * Returns the active offline store implementation based on environment.
 */
export function getDefaultOfflineStore(): IOfflineStore {
  if (typeof window !== 'undefined' && window.indexedDB) {
    return new IndexedDBOfflineStore();
  }
  return new InMemoryOfflineStore();
}

export type HttpExecutor = (item: OfflineQueueItem) => Promise<any>;

/**
 * SyncManager coordinates flushing the offline outbox once connectivity returns.
 */
export class SyncManager {
  private isFlushing = false;

  constructor(
    private store: IOfflineStore,
    private executor: HttpExecutor
  ) {}

  public async flush(): Promise<{ synced: number; failed: number }> {
    if (this.isFlushing) {
      return { synced: 0, failed: 0 };
    }

    this.isFlushing = true;
    let synced = 0;
    let failed = 0;

    try {
      const pending = await this.store.getPendingItems();

      for (const item of pending) {
        try {
          await this.executor(item);
          await this.store.removeItem(item.id);
          synced += 1;
        } catch {
          await this.store.incrementRetry(item.id);
          failed += 1;
        }
      }
    } finally {
      this.isFlushing = false;
    }

    return { synced, failed };
  }
}
