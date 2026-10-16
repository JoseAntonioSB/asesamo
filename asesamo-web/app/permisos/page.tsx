'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { WorkPermitItem } from '@/lib/api/types';
import { PermitSummaryCards } from '@/components/permisos/permit-summary-cards';
import { PermitListTable } from '@/components/permisos/permit-list-table';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ShieldAlert,
  PlusCircle,
  RefreshCw,
  LogOut,
  Building2,
  FileCheck2,
  QrCode,
  ShieldCheck,
  Flame,
  FileText,
} from 'lucide-react';

function PermisosDashboardContent() {
  const { user, logout } = useAuth();
  const [permits, setPermits] = useState<WorkPermitItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [selectedStatus, setSelectedStatus] = useState<string>('');

  const loadPermits = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const res = await apiClient.getPermits();
      setPermits(res.permits || []);

      if (showToast) {
        toast.success('Permisos actualizados con éxito');
      }
    } catch (err: any) {
      toast.error('Error al cargar permisos', {
        description: err.message || 'No se pudo obtener la lista de permisos de alto riesgo.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadPermits();
  }, [loadPermits]);

  const canCreatePermit =
    user?.role === 'contractor_admin' ||
    user?.role === 'plant_admin' ||
    user?.role === 'ehs_manager' ||
    user?.role === 'tenant_admin' ||
    user?.role === 'super_admin';

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-cyan-500 selection:text-white">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
              <ShieldAlert className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-cyan-950 text-cyan-300 font-mono border border-cyan-800">
                  PTAR Digital
                </span>
              </div>
              <p className="text-xs text-slate-400">Permisos de Trabajo de Alto Riesgo STPS</p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {/* Navigation links */}
            <nav className="hidden lg:flex items-center gap-1 text-xs">
              <Link
                href="/caseta"
                className="px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
              >
                Caseta
              </Link>
              <Link
                href="/documentos"
                className="px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
              >
                Documentos
              </Link>
              <Link
                href="/contratistas"
                className="px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
              >
                Contratistas
              </Link>
              <Link
                href="/permisos"
                className="px-3 py-1.5 rounded-lg bg-slate-800 text-cyan-400 font-semibold border border-slate-700 transition-colors"
              >
                Permisos PTAR
              </Link>
              {(user?.role === 'ehs_manager' ||
                user?.role === 'plant_admin' ||
                user?.role === 'tenant_admin' ||
                user?.role === 'super_admin') && (
                <Link
                  href="/ehs/aprobaciones"
                  className="px-3 py-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
                >
                  Bandeja EHS
                </Link>
              )}
            </nav>

            <ThemeToggle />

            {/* Logout button */}
            <button
              type="button"
              onClick={logout}
              title="Cerrar sesión"
              className="p-2 text-slate-400 hover:text-rose-400 hover:bg-slate-800 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Body Content */}
      <main className="flex-1 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 w-full space-y-6">
        {/* Page Header */}
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2.5">
              <h1 className="text-2xl font-bold tracking-tight text-white">
                Permisos de Trabajo de Alto Riesgo
              </h1>
              <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-950/80 text-cyan-300 font-mono font-medium border border-cyan-800">
                NOM-009 / 033 / 027 / 029
              </span>
            </div>
            <p className="text-xs sm:text-sm text-slate-400 mt-1">
              Monitoreo operativo, candados de cumplimiento DC-3 / médico y autorización digital en tiempo real.
            </p>
          </div>

          <div className="flex items-center gap-2.5 self-stretch sm:self-auto">
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadPermits(true)}
              disabled={isLoading || isRefreshing}
              className="border-slate-800 text-slate-300 hover:bg-slate-800 gap-1.5"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span>Actualizar</span>
            </Button>

            {canCreatePermit && (
              <Link href="/permisos/nuevo">
                <Button className="bg-cyan-600 hover:bg-cyan-500 text-white font-medium text-xs gap-1.5 shadow-lg shadow-cyan-950/50">
                  <PlusCircle className="w-4 h-4" />
                  <span>Nuevo Permiso PTAR</span>
                </Button>
              </Link>
            )}
          </div>
        </div>

        {/* KPI Metrics Summary Cards */}
        <PermitSummaryCards
          permits={permits}
          selectedStatus={selectedStatus}
          onSelectStatus={(st) => setSelectedStatus(st)}
        />

        {/* Filterable Table */}
        <PermitListTable
          permits={permits}
          isLoading={isLoading}
          selectedStatus={selectedStatus}
          onStatusChange={(st) => setSelectedStatus(st)}
          onRefresh={() => loadPermits(true)}
        />
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-4 text-center text-xs text-slate-600">
        ASESAMO Industrial Platform &bull; Candado Inteligente STPS Digital Compliance Engine &bull; {new Date().getFullYear()}
      </footer>
    </div>
  );
}

export default function PermisosPage() {
  return (
    <ProtectedRoute>
      <PermisosDashboardContent />
    </ProtectedRoute>
  );
}
