'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { SafetyIncidentItemDto, IncidentMetricsDto } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  AlertOctagon,
  PlusCircle,
  RefreshCw,
  LogOut,
  Flame,
  AlertTriangle,
  Activity,
  Calendar,
  MapPin,
  Clock,
  ArrowRight,
} from 'lucide-react';

function IncidentTypeBadge({ type }: { type: string }) {
  switch (type) {
    case 'fatality':
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-red-600 text-white">Fatalidad</span>;
    case 'lost_time_injury':
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-red-500/20 text-red-400 border border-red-500/40">Incapacidad (LTI)</span>;
    case 'medical_treatment':
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-orange-500/20 text-orange-400 border border-orange-500/40">Atención Médica</span>;
    case 'near_miss':
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-amber-500/20 text-amber-400 border border-amber-500/40">Cuasi-accidente</span>;
    case 'unsafe_condition':
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-yellow-500/20 text-yellow-400 border border-yellow-500/40">Condición Insegura</span>;
    case 'unsafe_act':
    default:
      return <span className="px-2.5 py-1 rounded-full text-xs font-bold bg-blue-500/20 text-blue-400 border border-blue-500/40">Acto Inseguro</span>;
  }
}

function IncidentesDashboardContent() {
  const { user, logout } = useAuth();
  const [incidents, setIncidents] = useState<SafetyIncidentItemDto[]>([]);
  const [metrics, setMetrics] = useState<IncidentMetricsDto | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const [incidentsData, metricsData] = await Promise.all([
        apiClient.getIncidents(),
        apiClient.getIncidentMetrics(),
      ]);

      setIncidents(incidentsData || []);
      setMetrics(metricsData || null);

      if (showToast) {
        toast.success('Incidentes actualizados');
      }
    } catch (err: any) {
      toast.error('Error al cargar datos de incidentes', {
        description: err.message || 'No se pudieron consultar los incidentes.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-emerald-500 selection:text-white">
      {/* Header Frosted Glass */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center justify-center text-red-400">
              <AlertOctagon className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Incidentes & Near-Misses (NOM-021-STPS)</div>
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
            <Link href="/incidentes" className="px-3 py-1.5 rounded-lg bg-zinc-800 text-white font-semibold">
              Incidentes
            </Link>
            <Link href="/reportes" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Reportes
            </Link>
          </nav>

          <div className="flex items-center gap-3">
            <ThemeToggle />
            <Link href="/incidentes/nuevo">
              <Button size="sm" className="bg-red-600 hover:bg-red-500 text-white font-semibold gap-1.5 shadow-sm">
                <PlusCircle className="w-4 h-4" />
                Reportar Incidente
              </Button>
            </Link>
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
        {/* Bento Grid de Métricas de Siniestralidad */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-6 rounded-2xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-zinc-400 font-medium">Total Reportados</span>
              <div className="text-3xl font-mono font-black text-white">{metrics?.total_incidents ?? 0}</div>
            </div>
            <div className="p-3 rounded-xl bg-zinc-800 text-zinc-300">
              <Activity className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-amber-950/30 border border-amber-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-amber-400 font-medium">Cuasi-Accidentes</span>
              <div className="text-3xl font-mono font-black text-amber-300">{metrics?.near_misses ?? 0}</div>
            </div>
            <div className="p-3 rounded-xl bg-amber-500/10 text-amber-400">
              <AlertTriangle className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-red-950/30 border border-red-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-red-400 font-medium">Con Tiempo Perdido (LTI)</span>
              <div className="text-3xl font-mono font-black text-red-300">{metrics?.lost_time_injuries ?? 0}</div>
            </div>
            <div className="p-3 rounded-xl bg-red-500/10 text-red-400">
              <Flame className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-zinc-400 font-medium">Días Perdidos Acumulados</span>
              <div className="text-3xl font-mono font-black text-white">{metrics?.total_days_lost ?? 0}</div>
            </div>
            <div className="p-3 rounded-xl bg-zinc-800 text-zinc-300">
              <Clock className="w-6 h-6" />
            </div>
          </div>
        </div>

        {/* Tabla de Eventos */}
        <div className="bg-zinc-900/70 border border-zinc-800 rounded-2xl overflow-hidden shadow-xl">
          <div className="p-6 border-b border-zinc-800/80 flex items-center justify-between">
            <div className="space-y-1">
              <h2 className="text-lg font-bold text-white">Bitácora de Eventos y Condiciones Inseguras</h2>
              <p className="text-xs text-zinc-400">Trazabilidad para investigación de causa raíz e índice de gravedad STPS.</p>
            </div>
            <button
              onClick={() => loadData(true)}
              disabled={isRefreshing}
              className="p-2 text-zinc-400 hover:text-white rounded-lg hover:bg-zinc-800 transition-colors"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            </button>
          </div>

          {isLoading ? (
            <div className="p-12 text-center text-zinc-500 text-sm">Cargando incidentes...</div>
          ) : incidents.length === 0 ? (
            <div className="p-12 text-center space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-zinc-800/80 text-zinc-400 mx-auto flex items-center justify-center">
                <AlertOctagon className="w-6 h-6 text-emerald-400" />
              </div>
              <p className="text-zinc-400 text-sm font-medium">Cero incidentes activos registrados. Operación segura.</p>
              <Link href="/incidentes/nuevo">
                <Button size="sm" className="bg-red-600 hover:bg-red-500 text-white">
                  Reportar Evento o Desvío
                </Button>
              </Link>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-zinc-950/50 text-zinc-400 font-mono uppercase tracking-wider border-b border-zinc-800">
                  <tr>
                    <th className="px-6 py-3.5">Folio</th>
                    <th className="px-6 py-3.5">Clasificación</th>
                    <th className="px-6 py-3.5">Título / Hallazgo</th>
                    <th className="px-6 py-3.5">Ubicación</th>
                    <th className="px-6 py-3.5">Días Perdidos</th>
                    <th className="px-6 py-3.5">Fecha</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60">
                  {incidents.map((inc) => (
                    <tr key={inc.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="px-6 py-4 font-mono font-bold text-white">{inc.incident_number}</td>
                      <td className="px-6 py-4">
                        <IncidentTypeBadge type={inc.incident_type} />
                      </td>
                      <td className="px-6 py-4 text-zinc-200 font-medium">{inc.title}</td>
                      <td className="px-6 py-4 text-zinc-400">{inc.location}</td>
                      <td className="px-6 py-4 font-mono font-bold text-red-400">
                        {inc.days_lost > 0 ? `${inc.days_lost} días` : '0'}
                      </td>
                      <td className="px-6 py-4 text-zinc-400">
                        {new Date(inc.event_timestamp).toLocaleDateString('es-MX')}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

export default function IncidentesPage() {
  return (
    <ProtectedRoute>
      <IncidentesDashboardContent />
    </ProtectedRoute>
  );
}
