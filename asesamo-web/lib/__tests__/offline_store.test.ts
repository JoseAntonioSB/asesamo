import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  InMemoryOfflineStore,
  SyncManager,
  type OfflineQueueItem,
} from '../offline/offline-store';

test('InMemoryOfflineStore: enqueues items, lists pending, and removes synced items', async () => {
  const store = new InMemoryOfflineStore();

  const item1: OfflineQueueItem = {
    id: 'queue-1',
    endpoint: '/api/v1/inspections/123/complete',
    method: 'POST',
    payload: { answers: [{ status: 'compliant' }] },
    created_at: new Date().toISOString(),
    retry_count: 0,
    description: 'Completar Inspección NOM-002',
  };

  const item2: OfflineQueueItem = {
    id: 'queue-2',
    endpoint: '/api/v1/evacuations/456/check-in',
    method: 'POST',
    payload: { worker_id: 'w1', muster_point_id: 'mp1' },
    created_at: new Date().toISOString(),
    retry_count: 0,
    description: 'Pase de Lista - Carlos Mendoza',
  };

  // 1. Enqueue
  await store.enqueue(item1);
  await store.enqueue(item2);

  // 2. Verify pending count
  let pending = await store.getPendingItems();
  assert.strictEqual(pending.length, 2);
  assert.strictEqual(pending[0].description, 'Completar Inspección NOM-002');

  // 3. Increment retry
  await store.incrementRetry('queue-1');
  pending = await store.getPendingItems();
  assert.strictEqual(pending[0].retry_count, 1);

  // 4. Remove synced item
  await store.removeItem('queue-1');
  pending = await store.getPendingItems();
  assert.strictEqual(pending.length, 1);
  assert.strictEqual(pending[0].id, 'queue-2');
});

test('SyncManager: replays queued items against HTTP handler and clears successful items', async () => {
  const store = new InMemoryOfflineStore();

  await store.enqueue({
    id: 'sync-1',
    endpoint: '/api/v1/inspections/abc/complete',
    method: 'POST',
    payload: { score: 100 },
    created_at: new Date().toISOString(),
    retry_count: 0,
    description: 'Inspección de Extintores',
  });

  await store.enqueue({
    id: 'sync-2',
    endpoint: '/api/v1/evacuations/def/check-in',
    method: 'POST',
    payload: { safe: true },
    created_at: new Date().toISOString(),
    retry_count: 0,
    description: 'Pase de lista brigada',
  });

  const executedRequests: string[] = [];
  const mockHttpExecutor = async (item: OfflineQueueItem) => {
    executedRequests.push(item.endpoint);
    return { success: true };
  };

  const syncManager = new SyncManager(store, mockHttpExecutor);
  const result = await syncManager.flush();

  assert.strictEqual(result.synced, 2);
  assert.strictEqual(result.failed, 0);
  assert.strictEqual(executedRequests.length, 2);
  assert.strictEqual(executedRequests[0], '/api/v1/inspections/abc/complete');

  // Store should now be empty
  const pending = await store.getPendingItems();
  assert.strictEqual(pending.length, 0);
});

test('SyncManager: keeps failed items in outbox and increments retry count', async () => {
  const store = new InMemoryOfflineStore();

  await store.enqueue({
    id: 'sync-fail-1',
    endpoint: '/api/v1/inspections/fail/complete',
    method: 'POST',
    payload: {},
    created_at: new Date().toISOString(),
    retry_count: 0,
    description: 'Inspección Fallida',
  });

  const mockHttpExecutor = async () => {
    throw new Error('Servidor no disponible 503');
  };

  const syncManager = new SyncManager(store, mockHttpExecutor);
  const result = await syncManager.flush();

  assert.strictEqual(result.synced, 0);
  assert.strictEqual(result.failed, 1);

  const pending = await store.getPendingItems();
  assert.strictEqual(pending.length, 1);
  assert.strictEqual(pending[0].retry_count, 1);
});
