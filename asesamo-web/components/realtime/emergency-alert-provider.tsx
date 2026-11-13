'use client';

import React, { createContext, useContext, useEffect, useState, useCallback, useRef } from 'react';
import Link from 'next/link';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { RealtimeEvacuationStartedEvent, RealtimeRollCallCheckInEvent, RealtimeEvacuationEndedEvent } from '@/lib/api/types';
import { casetaAudio } from '@/lib/utils/audio';
import { Siren, AlertTriangle, ArrowRight, X, Volume2, VolumeX } from 'lucide-react';
import { toast } from 'sonner';

interface EmergencyAlertContextType {
  activeEmergency: RealtimeEvacuationStartedEvent | null;
  isConnected: boolean;
  dismissBanner: () => void;
  requestNotificationPermission: () => Promise<NotificationPermission>;
}

const EmergencyAlertContext = createContext<EmergencyAlertContextType>({
  activeEmergency: null,
  isConnected: false,
  dismissBanner: () => {},
  requestNotificationPermission: async () => 'default',
});

export const useEmergencyAlerts = () => useContext(EmergencyAlertContext);

export function EmergencyAlertProvider({ children }: { children: React.ReactNode }) {
  const { user, token } = useAuth();
  const [activeEmergency, setActiveEmergency] = useState<RealtimeEvacuationStartedEvent | null>(null);
  const [isDismissed, setIsDismissed] = useState<boolean>(false);
  const [isConnected, setIsConnected] = useState<boolean>(false);
  const eventSourceRef = useRef<EventSource | null>(null);

  // Request browser push notification permission
  const requestNotificationPermission = useCallback(async (): Promise<NotificationPermission> => {
    if (typeof window === 'undefined' || !('Notification' in window)) {
      return 'denied';
    }
    if (Notification.permission === 'granted') {
      return 'granted';
    }
    return await Notification.requestPermission();
  }, []);

  // Format evacuation type to Spanish
  const formatEvacType = (type: string): string => {
    switch (type) {
      case 'fire':
        return 'Alarma de Incendio (NOM-002)';
      case 'gas_leak':
        return 'Fuga de Gas / Nube Tóxica';
      case 'earthquake':
        return 'Sismo / Daño Estructural';
      case 'spill':
        return 'Derrame Químico Mayor';
      case 'drill':
        return 'Simulacro Programado (Protección Civil)';
      default:
        return 'Contingencia Operativa';
    }
  };

  useEffect(() => {
    if (!token || !user) {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
        setIsConnected(false);
      }
      return;
    }

    const sseUrl = apiClient.getRealtimeEventsUrl(token);
    const es = new EventSource(sseUrl);
    eventSourceRef.current = es;

    es.addEventListener('open', () => {
      setIsConnected(true);
    });

    es.addEventListener('connected', () => {
      setIsConnected(true);
    });

    es.addEventListener('evacuation.started', (event: MessageEvent) => {
      try {
        const payload: RealtimeEvacuationStartedEvent = JSON.parse(event.data);
        setActiveEmergency(payload);
        setIsDismissed(false);

        // 1. Play siren
        casetaAudio.playEmergencySiren();

        // 2. Haptic alert
        if (typeof navigator !== 'undefined' && 'vibrate' in navigator) {
          navigator.vibrate([300, 150, 300, 150, 600]);
        }

        // 3. Desktop Notification
        if (typeof window !== 'undefined' && 'Notification' in window && Notification.permission === 'granted') {
          try {
            new Notification('¡ALARMA DE EVACUACIÓN EN PLANTA!', {
              body: `${formatEvacType(payload.evacuation_type)}: Folio ${payload.evacuation_number}. Desalojo hacia puntos de reunión.`,
              icon: '/icons/icon-192x192.png',
              tag: 'emergency-alarm',
              requireInteraction: true,
            });
          } catch {
            // Notification creation fails gracefully
          }
        }

        // 4. Dispatch custom event for /evacuacion page
        window.dispatchEvent(new CustomEvent('asesamo:evacuation:started', { detail: payload }));

        toast.error('¡ALARMA DE EVACUACIÓN ACTIVADA!', {
          description: `Folio ${payload.evacuation_number} • ${formatEvacType(payload.evacuation_type)}`,
          duration: 10000,
        });
      } catch (err) {
        console.error('Error handling evacuation.started event:', err);
      }
    });

    es.addEventListener('evacuation.roll_call_checkin', (event: MessageEvent) => {
      try {
        const payload: RealtimeRollCallCheckInEvent = JSON.parse(event.data);
        window.dispatchEvent(new CustomEvent('asesamo:evacuation:roll_call', { detail: payload }));
      } catch (err) {
        console.error('Error handling evacuation.roll_call_checkin event:', err);
      }
    });

    es.addEventListener('evacuation.ended', (event: MessageEvent) => {
      try {
        const payload: RealtimeEvacuationEndedEvent = JSON.parse(event.data);
        setActiveEmergency(null);
        setIsDismissed(false);
        window.dispatchEvent(new CustomEvent('asesamo:evacuation:ended', { detail: payload }));
        toast.success('All Clear: Fin de Contingencia Declarado', {
          description: 'La planta ha retornado a operación normal.',
        });
      } catch (err) {
        console.error('Error handling evacuation.ended event:', err);
      }
    });

    es.onerror = () => {
      setIsConnected(false);
    };

    return () => {
      es.close();
      eventSourceRef.current = null;
      setIsConnected(false);
    };
  }, [token, user]);

  const dismissBanner = () => {
    setIsDismissed(true);
  };

  return (
    <EmergencyAlertContext.Provider
      value={{
        activeEmergency,
        isConnected,
        dismissBanner,
        requestNotificationPermission,
      }}
    >
      {/* Realtime Industrial Emergency Top Banner */}
      {activeEmergency && !isDismissed && (
        <div className="bg-red-600 text-white px-4 py-3 sticky top-0 z-50 shadow-2xl border-b-2 border-red-400 flex flex-col sm:flex-row items-center justify-between gap-3 animate-pulse">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-white/20 shrink-0">
              <Siren className="w-6 h-6 animate-spin text-white" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-xs px-2 py-0.5 rounded-full bg-black/40 font-bold uppercase tracking-wider">
                  EMERGENCIA NOM-002 EN CURSO
                </span>
                <span className="text-xs font-mono font-bold">{activeEmergency.evacuation_number}</span>
              </div>
              <p className="text-sm font-bold text-white mt-0.5">
                {formatEvacType(activeEmergency.evacuation_type)}: Desalojo inmediato de áreas de trabajo hacia Puntos de Reunión.
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2.5 shrink-0 w-full sm:w-auto justify-end">
            <Link
              href="/evacuacion"
              className="px-4 py-1.5 rounded-lg bg-white text-red-700 font-bold text-xs hover:bg-red-50 flex items-center gap-1.5 shadow-md transition-colors"
            >
              <span>Ir al Centro de Mando</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>

            <button
              onClick={dismissBanner}
              title="Ocultar banner temporalmente"
              className="p-1.5 rounded-lg text-white/80 hover:text-white hover:bg-white/10 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
      )}

      {children}
    </EmergencyAlertContext.Provider>
  );
}
