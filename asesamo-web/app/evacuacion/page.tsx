'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { EmergencyEvacuationDto, MusterPointDto, EvacuationType } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  Siren,
  ShieldAlert,
  Users,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Radio,
  LogOut,
  MapPin,
  Flame,
  Activity,
  ArrowRight,
  Printer,
} from 'lucide-react';

function EvacuacionDashboardContent() {
  const { user, logout } = useAuth();
  const [activeEvac, setActiveEvac] = useState<EmergencyEvacuationDto | null>(null);
  const [musterPoints, setMusterPoints] = useState<MusterPointDto[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isTriggering, setIsTriggering] = useState<boolean>(false);
  const [drillType, setDrillType] = useState<EvacuationType>('drill');
  const [drillDesc, setDrillDesc] = useState<string>('Simulacro General de Repliegue y Evacuación');

  const loadData = useCallback(async (silent = false) => {
    try {
      if (!silent) setIsLoading(true);
      const [evacData, pointsData] = await Promise.all([
        apiClient.getActiveEvacuation(),
        apiClient.getMusterPoints(),
      ]);

      setActiveEvac(evacData || null);
      setMusterPoints(pointsData || []);
    } catch (err: any) {
      if (!silent) {
        toast.error('Error al consultar estado de contingencia', {
          description: err.message,
        });
      }
    } finally {
      if (!silent) setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();

    // Sondeo periódico cada 10 segundos como fallback
    const interval = setInterval(() => {
      loadData(true);
    }, 10000);

    // Escucha instantánea de eventos en tiempo real SSE
    const handleEvacStarted = () => {
      loadData(true);
    };
    const handleRollCall = () => {
      loadData(true);
    };
    const handleEvacEnded = () => {
      setActiveEvac(null);
      loadData(true);
    };

    window.addEventListener('asesamo:evacuation:started', handleEvacStarted);
    window.addEventListener('asesamo:evacuation:roll_call', handleRollCall);
    window.addEventListener('asesamo:evacuation:ended', handleEvacEnded);

    return () => {
      clearInterval(interval);
      window.removeEventListener('asesamo:evacuation:started', handleEvacStarted);
      window.removeEventListener('asesamo:evacuation:roll_call', handleRollCall);
      window.removeEventListener('asesamo:evacuation:ended', handleEvacEnded);
    };
  }, [loadData]);

  const handleStartEvacuation = async () => {
    try {
      setIsTriggering(true);
      const evac = await apiClient.startEvacuation({
        type: drillType,
        description: drillDesc,
      });

      setActiveEvac(evac);
      toast.error('¡ALERTA DE EVACUACIÓN ACTIVADA!', {
        description: `Se ha iniciado el pase de lista de ${evac.total_headcount} personas registradas en planta.`,
      });
    } catch (err: any) {
      toast.error('Error al activar evacuación', {
        description: err.message,
      });
    } finally {
      setIsTriggering(false);
    }
  };

  const handleEndEvacuation = async () => {
    if (!activeEvac) return;

    try {
      setIsTriggering(true);
      const res = await apiClient.endEvacuation(activeEvac.id);
      setActiveEvac(null);

      toast.success('CONTINGENCIA SUPERADA (ALL CLEAR)', {
        description: `Tiempo total de desalojo: ${res.evacuation_time_seconds} segundos. Todos a salvo.`,
      });
    } catch (err: any) {
      toast.error('No se puede cerrar la evacuación', {
        description: err.message || 'Verifique que no haya personal faltante.',
      });
    } finally {
      setIsTriggering(false);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-red-500 selection:text-white">
      {/* Header Frosted Glass */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`w-9 h-9 rounded-xl flex items-center justify-center ${
              activeEvac ? 'bg-red-500/20 text-red-400 border border-red-500/40 animate-pulse' : 'bg-zinc-800 text-zinc-300'
            }`}>
              <Siren className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Plan de Evacuación & Simulacros (NOM-002-STPS)</div>
            </div>
          </div>

          <nav className="hidden md:flex items-center gap-1 bg-zinc-900/60 border border-zinc-800/80 p-1 rounded-xl text-xs">
            <Link href="/caseta" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Caseta
            </Link>
            <Link href="/permisos" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Permisos PTAR
            </Link>
            <Link href="/inspecciones" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Inspecciones
            </Link>
            <Link href="/incidentes" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Incidentes
            </Link>
            <Link href="/evacuacion" className="px-3 py-1.5 rounded-lg bg-red-950 text-red-200 border border-red-500/40 font-semibold">
              Evacuación
            </Link>
            <Link href="/reportes" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Reportes
            </Link>
          </nav>

          <div className="flex items-center gap-3">
            <ThemeToggle />
            <button
              onClick={() => logout()}
              title="Cerrar sesión"
              className="p-2 text-zinc-400 hover:text-zinc-200 rounded-lg hover:bg-zinc-900 transition-colors"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </header>

      {/* Contenido Principal */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Banner de Estado Activo o Normal */}
        {activeEvac ? (
          <div className="p-8 rounded-3xl bg-gradient-to-r from-red-950 via-zinc-950 to-red-950 border-2 border-red-500/80 shadow-2xl shadow-red-950/60 space-y-6 animate-in fade-in">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div className="space-y-2">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-red-500 text-white text-xs font-black tracking-widest uppercase">
                  <span className="w-2 h-2 rounded-full bg-white animate-ping" />
                  EMERGENCIA EN CURSO • FOLIO {activeEvac.evacuation_number}
                </div>
                <h1 className="text-3xl font-black text-white tracking-tight uppercase">
                  Evacuación de Planta Activada
                </h1>
                <p className="text-sm text-red-200">
                  {activeEvac.description || 'Desalojo preventivo de áreas operativas hacia Puntos de Reunión seguros.'}
                </p>
              </div>

              <Button
                onClick={handleEndEvacuation}
                disabled={isTriggering || activeEvac.missing_count > 0}
                className="bg-emerald-600 hover:bg-emerald-500 text-white font-bold px-8 py-4 rounded-2xl shadow-xl shadow-emerald-950/50 disabled:opacity-50 disabled:cursor-not-allowed text-base gap-2"
              >
                <CheckCircle2 className="w-5 h-5" />
                Declarar All Clear (Fin de Contingencia)
              </Button>
            </div>

            {/* Contadores Críticos en Tiempo Real */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-4 border-t border-red-500/30">
              <div className="p-5 rounded-2xl bg-zinc-950/80 border border-red-500/30">
                <span className="text-xs uppercase tracking-wider text-zinc-400 font-bold">Censo de Planta (Caseta)</span>
                <div className="text-4xl font-mono font-black text-white mt-1">{activeEvac.total_headcount}</div>
                <span className="text-[11px] text-zinc-400">Personas en planta al sonar alarma</span>
              </div>

              <div className="p-5 rounded-2xl bg-emerald-950/40 border border-emerald-500/40">
                <span className="text-xs uppercase tracking-wider text-emerald-400 font-bold">A Salvo en Puntos de Reunión</span>
                <div className="text-4xl font-mono font-black text-emerald-400 mt-1">{activeEvac.safe_count}</div>
                <span className="text-[11px] text-emerald-300">Pase de lista verificado</span>
              </div>

              <div className="p-5 rounded-2xl bg-red-950/60 border border-red-500/60">
                <span className="text-xs uppercase tracking-wider text-red-400 font-bold">Faltantes / En Búsqueda</span>
                <div className="text-4xl font-mono font-black text-red-400 mt-1 animate-pulse">{activeEvac.missing_count}</div>
                <span className="text-[11px] text-red-300">Brigadas de rescate desplegadas</span>
              </div>
            </div>
          </div>
        ) : (
          <div className="p-8 rounded-3xl bg-zinc-900/60 border border-zinc-800 space-y-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div className="space-y-1">
                <span className="text-xs font-mono uppercase tracking-widest text-emerald-400">
                  Condición Normal • Planta en Operación
                </span>
                <h1 className="text-2xl font-bold text-white">Centro de Comando de Evacuación y Simulacros</h1>
                <p className="text-xs text-zinc-400">
                  Contrastación automática en tiempo real contra registros de caseta de acceso.
                </p>
              </div>

              <div className="flex flex-wrap items-center gap-3">
                <Link href="/evacuacion/acta" target="_blank">
                  <Button
                    variant="outline"
                    className="border-zinc-700 text-zinc-300 hover:text-white hover:bg-zinc-800 gap-2 font-bold px-4 py-3 rounded-xl cursor-pointer"
                  >
                    <Printer className="w-4 h-4 text-zinc-400" />
                    <span>Acta Oficial / PDF</span>
                  </Button>
                </Link>

                <Button
                  onClick={handleStartEvacuation}
                  disabled={isTriggering}
                  className="bg-red-600 hover:bg-red-500 text-white font-bold px-6 py-3 rounded-xl shadow-lg shadow-red-950/50 gap-2 cursor-pointer"
                >
                  <Siren className="w-4 h-4 animate-bounce" />
                  {isTriggering ? 'Iniciando...' : 'Activar Alarma de Evacuación / Simulacro'}
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4 border-t border-zinc-800/80">
              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-400">
                  Tipo de Contingencia
                </label>
                <select
                  value={drillType}
                  onChange={(e) => setDrillType(e.target.value as EvacuationType)}
                  className="w-full px-4 py-2.5 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-xs focus:outline-none focus:ring-2 focus:ring-red-500/50"
                >
                  <option value="drill">Simulacro Programado (Protección Civil)</option>
                  <option value="fire">Alarma de Incendio (NOM-002)</option>
                  <option value="gas_leak">Fuga de Gas / Nube Tóxica</option>
                  <option value="earthquake">Sismo / Daño Estructural</option>
                  <option value="spill">Derrame Químico Mayor</option>
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-400">
                  Hipótesis o Motivo del Ejercicio
                </label>
                <input
                  type="text"
                  value={drillDesc}
                  onChange={(e) => setDrillDesc(e.target.value)}
                  className="w-full px-4 py-2.5 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-xs focus:outline-none focus:ring-2 focus:ring-red-500/50"
                />
              </div>
            </div>
          </div>
        )}

        {/* Puntos de Reunión de la Planta */}
        <div className="bg-zinc-900/70 border border-zinc-800 rounded-2xl p-6 space-y-4">
          <div className="space-y-1">
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <MapPin className="w-5 h-5 text-emerald-400" />
              Puntos de Reunión y Zonas de Seguridad Verificadas
            </h2>
            <p className="text-xs text-zinc-400">
              Estaciones seguras designadas para pase de lista en caso de repliegue o desalojo de naves.
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 pt-2">
            {musterPoints.length > 0 ? (
              musterPoints.map((pt) => (
                <div key={pt.id} className="p-4 rounded-xl bg-zinc-950 border border-zinc-800/80 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="font-mono font-bold text-emerald-400 text-xs">{pt.code}</span>
                    <span className="text-[10px] font-mono text-zinc-500">Capacidad: {pt.capacity} personas</span>
                  </div>
                  <div className="font-bold text-sm text-zinc-100">{pt.name}</div>
                  {pt.location_description && (
                    <div className="text-xs text-zinc-400 italic">{pt.location_description}</div>
                  )}
                </div>
              ))
            ) : (
              <>
                <div className="p-4 rounded-xl bg-zinc-950 border border-zinc-800/80 space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-mono font-bold text-emerald-400 text-xs">PR-01</span>
                    <span className="text-[10px] font-mono text-zinc-500">Cap: 150 pax</span>
                  </div>
                  <div className="font-bold text-sm text-zinc-100">Estacionamiento Principal Norte</div>
                  <div className="text-xs text-zinc-400">Frente a Caseta 1, zona despejada de líneas eléctricas.</div>
                </div>

                <div className="p-4 rounded-xl bg-zinc-950 border border-zinc-800/80 space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-mono font-bold text-emerald-400 text-xs">PR-02</span>
                    <span className="text-[10px] font-mono text-zinc-500">Cap: 100 pax</span>
                  </div>
                  <div className="font-bold text-sm text-zinc-100">Cancha Deportiva / Área Verde Sur</div>
                  <div className="text-xs text-zinc-400">Acceso posterior por puerta de emergencia Nave 3.</div>
                </div>
              </>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}

export default function EvacuacionPage() {
  return (
    <ProtectedRoute>
      <EvacuacionDashboardContent />
    </ProtectedRoute>
  );
}
