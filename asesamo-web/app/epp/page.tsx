'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { EPPItemDto, EPPDeliveryDto } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  Shield,
  PlusCircle,
  RefreshCw,
  LogOut,
  Package,
  Clock,
  CheckCircle2,
  AlertTriangle,
  HardHat,
  Eye,
  Ear,
  Hand,
  Footprints,
} from 'lucide-react';

function EPPCategoryIcon({ category }: { category: string }) {
  switch (category) {
    case 'head':
      return <HardHat className="w-4 h-4 text-amber-400" />;
    case 'eyes_face':
      return <Eye className="w-4 h-4 text-blue-400" />;
    case 'hearing':
      return <Ear className="w-4 h-4 text-purple-400" />;
    case 'hands':
      return <Hand className="w-4 h-4 text-emerald-400" />;
    case 'feet':
      return <Footprints className="w-4 h-4 text-orange-400" />;
    case 'fall_arrest':
      return <Shield className="w-4 h-4 text-red-400" />;
    default:
      return <Package className="w-4 h-4 text-zinc-400" />;
  }
}

function EPPDashboardContent() {
  const { user, logout } = useAuth();
  const [catalog, setCatalog] = useState<EPPItemDto[]>([]);
  const [deliveries, setDeliveries] = useState<EPPDeliveryDto[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const [catalogData, deliveriesData] = await Promise.all([
        apiClient.getEPPCatalog(),
        apiClient.getEPPDeliveries(),
      ]);

      setCatalog(catalogData || []);
      setDeliveries(deliveriesData || []);

      if (showToast) {
        toast.success('Inventario y entregas de EPP actualizadas');
      }
    } catch (err: any) {
      toast.error('Error al cargar datos de EPP', {
        description: err.message || 'No se pudo obtener el catálogo de protección personal.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const activeDeliveriesCount = deliveries.filter((d) => d.status === 'delivered_active').length;

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-emerald-500 selection:text-white">
      {/* Header Frosted Glass */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
              <HardHat className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Control & Dotación de EPP (NOM-017-STPS)</div>
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
            <Link href="/epp" className="px-3 py-1.5 rounded-lg bg-zinc-800 text-white font-semibold">
              EPP
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
        {/* Contadores Bento */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="p-6 rounded-2xl bg-zinc-900/80 border border-zinc-800 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-zinc-400 font-medium">Artículos en Catálogo</span>
              <div className="text-3xl font-mono font-black text-white">{catalog.length}</div>
            </div>
            <div className="p-3 rounded-xl bg-zinc-800 text-zinc-300">
              <Package className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-emerald-950/30 border border-emerald-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-emerald-400 font-medium">Dotaciones Activas</span>
              <div className="text-3xl font-mono font-black text-emerald-300">{activeDeliveriesCount}</div>
            </div>
            <div className="p-3 rounded-xl bg-emerald-500/10 text-emerald-400">
              <CheckCircle2 className="w-6 h-6" />
            </div>
          </div>

          <div className="p-6 rounded-2xl bg-amber-950/30 border border-amber-800/40 flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-xs uppercase tracking-wider text-amber-400 font-medium">Normativa Oficial</span>
              <div className="text-sm font-mono font-bold text-amber-300">NOM-017-STPS-2008</div>
              <p className="text-[11px] text-zinc-400">Equipo de Protección Personal en los Centros de Trabajo</p>
            </div>
            <div className="p-3 rounded-xl bg-amber-500/10 text-amber-400">
              <Shield className="w-6 h-6" />
            </div>
          </div>
        </div>

        {/* Catálogo de EPP */}
        <div className="bg-zinc-900/70 border border-zinc-800 rounded-2xl overflow-hidden shadow-xl">
          <div className="p-6 border-b border-zinc-800/80 flex items-center justify-between">
            <div className="space-y-1">
              <h2 className="text-lg font-bold text-white">Catálogo de Protección Personal Aprobado</h2>
              <p className="text-xs text-zinc-400">Estándares técnicos y vida útil de recambio para cuadrillas y personal de planta.</p>
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
            <div className="p-12 text-center text-zinc-500 text-sm">Cargando catálogo...</div>
          ) : catalog.length === 0 ? (
            <div className="p-12 text-center text-zinc-400 text-sm">
              Catálogo inicial configurado: Casco Clase E (NOM-115), Lentes Z87.1, Calzado Dieléctrico (NOM-113) y Arnés de Alturas (NOM-009).
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-zinc-950/50 text-zinc-400 font-mono uppercase tracking-wider border-b border-zinc-800">
                  <tr>
                    <th className="px-6 py-3.5">Código</th>
                    <th className="px-6 py-3.5">Categoría</th>
                    <th className="px-6 py-3.5">Artículo / Modelo</th>
                    <th className="px-6 py-3.5">Norma Técnica</th>
                    <th className="px-6 py-3.5">Vida Útil (Días)</th>
                    <th className="px-6 py-3.5">Stock</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60">
                  {catalog.map((item) => (
                    <tr key={item.id} className="hover:bg-zinc-800/30 transition-colors">
                      <td className="px-6 py-4 font-mono font-bold text-white">{item.code}</td>
                      <td className="px-6 py-4 flex items-center gap-2">
                        <EPPCategoryIcon category={item.category} />
                        <span className="capitalize text-zinc-300 font-medium">{item.category}</span>
                      </td>
                      <td className="px-6 py-4 text-zinc-100 font-semibold">{item.name}</td>
                      <td className="px-6 py-4 text-zinc-400">{item.normative_standard || 'NOM-017-STPS'}</td>
                      <td className="px-6 py-4 font-mono text-zinc-300">{item.lifespan_days} días</td>
                      <td className="px-6 py-4 font-mono font-bold text-emerald-400">{item.stock_quantity} pzas</td>
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

export default function EPPPage() {
  return (
    <ProtectedRoute>
      <EPPDashboardContent />
    </ProtectedRoute>
  );
}
