'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { WifiOff, Wifi, RefreshCw } from 'lucide-react';
import {
  getDefaultOfflineStore,
  SyncManager,
  type OfflineQueueItem,
  type IOfflineStore,
} from '@/lib/offline/offline-store';
import { apiClient } from '@/lib/api/client';

interface PwaContextValue {
  isOnline: boolean;
  pendingCount: number;
  enqueueOfflineAction: (
    endpoint: string,
    method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    payload: any,
    description: string
  ) => Promise<void>;
  flushPendingSync: () => Promise<void>;
}

const PwaContext = createContext<PwaContextValue | null>(null);

export function usePwa() {
  const ctx = useContext(PwaContext);
  if (!ctx) {
    throw new Error('usePwa must be used within a PwaProvider');
  }
  return ctx;
}

export function PwaProvider({ children }: { children: React.ReactNode }) {
  const [isOnline, setIsOnline] = useState<boolean>(true);
  const [pendingCount, setPendingCount] = useState<number>(0);
  const [isSyncing, setIsSyncing] = useState<boolean>(false);
  const [store] = useState<IOfflineStore>(() => getDefaultOfflineStore());

  const refreshPendingCount = useCallback(async () => {
    try {
      const items = await store.getPendingItems();
      setPendingCount(items.length);
    } catch {
      // Ignored in non-browser envs
    }
  }, [store]);

  const flushPendingSync = useCallback(async () => {
    if (!navigator.onLine || isSyncing) return;

    try {
      setIsSyncing(true);
      const syncManager = new SyncManager(store, async (item: OfflineQueueItem) => {
        // Execute request against backend
        return apiClient.request(item.endpoint, {
          method: item.method,
          body: JSON.stringify(item.payload),
        });
      });

      const res = await syncManager.flush();
      await refreshPendingCount();

      if (res.synced > 0) {
        toast.success('Sincronización PWA completada', {
          description: `Se sincronizaron ${res.synced} operación(es) registradas en campo sin conexión.`,
        });
      }
    } catch (err: any) {
      toast.error('Error durante la sincronización automática', {
        description: err.message || 'Algunos registros no pudieron enviarse al servidor.',
      });
    } finally {
      setIsSyncing(false);
    }
  }, [store, isSyncing, refreshPendingCount]);

  const enqueueOfflineAction = useCallback(
    async (
      endpoint: string,
      method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
      payload: any,
      description: string
    ) => {
      const item: OfflineQueueItem = {
        id: `offline-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
        endpoint,
        method,
        payload,
        created_at: new Date().toISOString(),
        retry_count: 0,
        description,
      };

      await store.enqueue(item);
      await refreshPendingCount();

      toast.info('Registro guardado localmente', {
        description: `"${description}" se sincronizará automáticamente al recuperar la señal de red.`,
      });
    },
    [store, refreshPendingCount]
  );

  useEffect(() => {
    // 1. Initial status check
    if (typeof window !== 'undefined') {
      setIsOnline(window.navigator.onLine);
      refreshPendingCount();

      // 2. Register Service Worker in production / supported environments
      if ('serviceWorker' in navigator && process.env.NODE_ENV === 'production') {
        navigator.serviceWorker
          .register('/sw.js')
          .then((reg) => {
            console.log('[PWA] Service Worker registrado con alcance:', reg.scope);
          })
          .catch((err) => {
            console.warn('[PWA] Fallo al registrar Service Worker:', err);
          });
      }

      // 3. Connectivity Event Listeners
      const handleOnline = () => {
        setIsOnline(true);
        toast.success('Conexión a red recuperada', {
          description: 'Reanudando sincronización con el servidor...',
        });
        flushPendingSync();
      };

      const handleOffline = () => {
        setIsOnline(false);
        toast.warning('Modo Sin Conexión activado', {
          description: 'Las auditorías y pases de lista se guardarán en almacenamiento local protegido.',
        });
      };

      window.addEventListener('online', handleOnline);
      window.addEventListener('offline', handleOffline);

      return () => {
        window.removeEventListener('online', handleOnline);
        window.removeEventListener('offline', handleOffline);
      };
    }
  }, [flushPendingSync, refreshPendingCount]);

  return (
    <PwaContext.Provider
      value={{
        isOnline,
        pendingCount,
        enqueueOfflineAction,
        flushPendingSync,
      }}
    >
      {children}

      {/* Floating Offline Indicator */}
      {!isOnline && (
        <div className="fixed bottom-4 left-4 z-50 flex items-center gap-2.5 px-3.5 py-2 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-300 backdrop-blur-md shadow-2xl shadow-amber-500/10 text-xs font-medium animate-in fade-in slide-in-from-bottom-2">
          <WifiOff className="w-4 h-4 text-amber-400 animate-pulse" />
          <span>
            Sin conexión a internet
            {pendingCount > 0 && (
              <strong className="ml-1 text-white">({pendingCount} pendientes)</strong>
            )}
          </span>
          <button
            onClick={flushPendingSync}
            disabled={isSyncing}
            className="ml-1 p-1 hover:bg-amber-500/20 rounded-md transition-colors"
            title="Reintentar sincronización"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isSyncing ? 'animate-spin' : ''}`} />
          </button>
        </div>
      )}

      {/* Syncing Toast Pill */}
      {isOnline && isSyncing && (
        <div className="fixed bottom-4 left-4 z-50 flex items-center gap-2.5 px-3.5 py-2 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 backdrop-blur-md shadow-2xl shadow-emerald-500/10 text-xs font-medium animate-in fade-in">
          <Wifi className="w-4 h-4 text-emerald-400" />
          <span>Sincronizando registros con el servidor...</span>
          <RefreshCw className="w-3.5 h-3.5 animate-spin" />
        </div>
      )}
    </PwaContext.Provider>
  );
}
