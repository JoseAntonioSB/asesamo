'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { SafetyInspectionItemDto } from '@/lib/api/types';
import { InspectionStatusBadge } from '@/components/inspecciones/inspection-status-badge';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ClipboardCheck,
  PlusCircle,
  RefreshCw,
  LogOut,
  Building2,
  FileCheck2,
  QrCode,
  ShieldCheck,
  AlertTriangle,
  ArrowRight,
} from 'lucide-react';

function InspeccionesDashboardContent() {
  const { user, logout } = useAuth();
  const [inspections, setInspections] = useState<SafetyInspectionItemDto[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const loadInspections = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const res = await apiClient.getInspections();
      setInspections(res || []);

      if (showToast) {
        toast.success('Inspecciones actualizadas con éxito');
      }
    } catch (err: any) {
      toast.error('Error al cargar inspecciones', {
        description: err.message || 'No se pudo obtener la lista de auditorías de seguridad.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadInspections();
  }, [loadInspections]);

  const completedCount = inspections.filter((i) => i.status === 'completed').length;
  const capaCount = inspections.filter((i) => i.status === 'requires_capa').length;

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-emerald-500 selection:text-white">
      {/* Header Flotante Frosted Glass */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <ClipboardCheck className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Inspecciones & Auditorías EHS</div>
            </div>
          </div>

          <nav className="hidden md:flex items-center gap-1 bg-zinc-900/60 border border-zinc-800/80 p-1 rounded-xl text-xs">
            <Link href="/caseta" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Caseta
            </Link>
            <Link href="/permisos" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Permisos PTAR
            </Link>
            <Link href="/inspecciones" className="px-3 py-1.5 rounded-lg bg-zinc-800 text-white font-semibold">
              Inspecciones
            </Link>
            <Link href="/contratistas" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Contratistas
            </Link>
            <Link href="/reportes" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Reportes
            </Link>
          </nav>

          <div className="flex items-center gap-3">
            <ThemeToggle />
            <Link href="/inspecciones/nueva">
              <Button size="sm" className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold gap-1.5 shadow-sm">
                <PlusCircle className="w-4 h-4" />
                Nueva Inspección
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
        {/* Banner Superior & Contadores */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="p-6 rounded-2xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-zinc-400 font-medium">Total Realizadas</span>
              <div className="text-3xl font-mono font-black text-white">{inspections.length}</div>
            </div>
            <div className="p-3 rounded-xl bg-zinc-800 text-zinc-300">
              <ClipboardCheck className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-emerald-950/30 border border-emerald-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-emerald-400 font-medium">100% Conformes</span>
              <div className="text-3xl font-mono font-black text-emerald-300">{completedCount}</div>
            </div>
            <div className="p-3 rounded-xl bg-emerald-500/10 text-emerald-400">
              <ShieldCheck className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-amber-950/30 border border-amber-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-amber-400 font-medium">Con Plan de Acción (CAPA)</span>
              <div className="text-3xl font-mono font-black text-amber-300">{capaCount}</div>
            </div>
            <div className="p-3 rounded-xl bg-amber-500/10 text-amber-400">
              <AlertTriangle className="w-6 h-6" />
            </div>
          </div>
        </div>

        {/* Tabla de Inspecciones */}
        <div className="bg-zinc-900/70 border border-zinc-800 rounded-2xl overflow-hidden shadow-xl">
          <div className="p-6 border-b border-zinc-800/80 flex items-center justify-between">
            <div className="space-y-1">
              <h2 className="text-lg font-bold text-white">Registro de Auditorías e Inspecciones en Campo</h2>
              <p className="text-xs text-zinc-400">Verificaciones periódicas de condiciones y actos seguros bajo normas STPS.</p>
            </div>
            <button
              onClick={() => loadInspections(true)}
              disabled={isRefreshing}
              className="p-2 text-zinc-400 hover:text-white rounded-lg hover:bg-zinc-800 transition-colors"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            </button>
          </div>

          {isLoading ? (
            <div className="p-12 text-center text-zinc-500 text-sm">Cargando inspecciones...</div>
          ) : inspections.length === 0 ? (
            <div className="p-12 text-center space-y-4">
              <div className="w-12 h-12 rounded-2xl bg-zinc-800/80 text-zinc-400 mx-auto flex items-center justify-center">
                <ClipboardCheck className="w-6 h-6" />
              </div>
              <p className="text-zinc-400 text-sm">No hay inspecciones registradas en este centro de trabajo.</p>
              <Link href="/inspecciones/nueva">
                <Button size="sm" className="bg-emerald-600 hover:bg-emerald-500 text-white">
                  Iniciar Primera Inspección
                </Button>
              </Link>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-zinc-950/50 text-zinc-400 font-mono uppercase tracking-wider border-b border-zinc-800">
                  <tr>
                    <th className="px-6 py-3.5">Folio</th>
                    <th className="px-6 py-3.5">Tipo</th>
                    <th className="px-6 py-3.5">Ubicación / Área</th>
                    <th className="px-6 py-3.5">Puntaje</th>
                    <th className="px-6 py-3.5">Estatus</th>
                    <th className="px-6 py-3.5">Fecha</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60">
                  {inspections.map((insp) => (
                    <tr key={insp.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="px-6 py-4 font-mono font-bold text-white">{insp.inspection_number}</td>
                      <td className="px-6 py-4 uppercase font-semibold text-zinc-400">{insp.inspection_type}</td>
                      <td className="px-6 py-4 text-zinc-200">{insp.area_or_location}</td>
                      <td className="px-6 py-4 font-mono font-bold text-emerald-400">
                        {insp.score_percentage !== null && insp.score_percentage !== undefined
                          ? `${insp.score_percentage}%`
                          : '—'}
                      </td>
                      <td className="px-6 py-4">
                        <InspectionStatusBadge status={insp.status} score={insp.score_percentage} />
                      </td>
                      <td className="px-6 py-4 text-zinc-400">
                        {new Date(insp.created_at).toLocaleDateString('es-MX')}
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

export default function InspeccionesPage() {
  return (
    <ProtectedRoute>
      <InspeccionesDashboardContent />
    </ProtectedRoute>
  );
}
