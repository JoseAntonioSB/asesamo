'use client';

import React from 'react';
import { WorkPermitItem } from '@/lib/api/types';
import { Activity, Clock, CheckCircle2, AlertTriangle, ShieldCheck, Flame, Zap, ArrowUpRight } from 'lucide-react';

interface PermitSummaryCardsProps {
  permits: WorkPermitItem[];
  selectedStatus?: string;
  onSelectStatus?: (status: string) => void;
}

export function PermitSummaryCards({
  permits,
  selectedStatus,
  onSelectStatus,
}: PermitSummaryCardsProps) {
  const activeCount = permits.filter((p) => p.status === 'approved_active').length;
  const pendingCount = permits.filter((p) => p.status === 'pending_approval').length;
  const closedCount = permits.filter((p) => p.status === 'closed').length;
  const suspendedCount = permits.filter((p) => p.status === 'suspended').length;
  const totalCount = permits.length;

  // Conteo por tipo de riesgo crítico STPS
  const heightsCount = permits.filter((p) => p.permit_type === 'heights').length;
  const hotWorkCount = permits.filter((p) => p.permit_type === 'hot_work').length;
  const confinedCount = permits.filter((p) => p.permit_type === 'confined_space').length;
  const lotoCount = permits.filter((p) => p.permit_type === 'electrical_loto').length;

  return (
    <div className="grid grid-cols-1 md:grid-cols-12 gap-4 auto-rows-fr [grid-auto-flow:dense]">
      {/* Celda Heroica Bento: Permisos Activos en Campo (5 cols) */}
      <div
        onClick={() => onSelectStatus?.(selectedStatus === 'approved_active' ? '' : 'approved_active')}
        className={`md:col-span-5 relative group overflow-hidden rounded-2xl p-6 border transition-all duration-300 cursor-pointer flex flex-col justify-between ${
          selectedStatus === 'approved_active'
            ? 'ring-2 ring-emerald-500 bg-gradient-to-br from-emerald-950/60 via-zinc-950 to-zinc-950 border-emerald-500/50 shadow-lg shadow-emerald-950/40'
            : 'bg-zinc-950/80 border-zinc-800/80 hover:border-emerald-500/40 hover:bg-zinc-900/40'
        }`}
      >
        <div className="absolute top-0 right-0 w-32 h-32 bg-emerald-500/10 rounded-full blur-2xl group-hover:bg-emerald-500/20 transition-all pointer-events-none" />

        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
              </span>
              <span className="text-xs font-bold uppercase tracking-wider text-emerald-400">
                En Ejecución Activa
              </span>
            </div>
            <h3 className="text-lg font-bold text-zinc-100">Trabajos de Alto Riesgo en Planta</h3>
          </div>
          <div className="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 group-hover:scale-105 transition-transform">
            <Activity className="w-5 h-5" />
          </div>
        </div>

        <div className="my-6">
          <div className="flex items-baseline gap-3">
            <span className="text-5xl font-black font-mono tracking-tight text-white group-hover:text-emerald-400 transition-colors">
              {activeCount}
            </span>
            <span className="text-xs text-zinc-400 font-medium">
              de {totalCount} permisos totales emitidos
            </span>
          </div>
        </div>

        <div className="pt-3 border-t border-zinc-800/80 flex items-center justify-between text-xs text-zinc-400">
          <span className="flex items-center gap-1.5 text-emerald-300 font-medium">
            <ShieldCheck className="w-4 h-4 text-emerald-400" />
            Candados STPS Verificados
          </span>
          <span className="text-zinc-400 group-hover:text-zinc-200 flex items-center gap-1 font-mono text-[11px]">
            Filtrar activos <ArrowUpRight className="w-3.5 h-3.5" />
          </span>
        </div>
      </div>

      {/* Celda Bento 2: Pendientes de Autorización EHS (4 cols) */}
      <div
        onClick={() => onSelectStatus?.(selectedStatus === 'pending_approval' ? '' : 'pending_approval')}
        className={`md:col-span-4 relative group overflow-hidden rounded-2xl p-6 border transition-all duration-300 cursor-pointer flex flex-col justify-between ${
          selectedStatus === 'pending_approval'
            ? 'ring-2 ring-amber-500 bg-gradient-to-br from-amber-950/60 via-zinc-950 to-zinc-950 border-amber-500/50 shadow-lg shadow-amber-950/40'
            : 'bg-zinc-950/80 border-zinc-800/80 hover:border-amber-500/40 hover:bg-zinc-900/40'
        }`}
      >
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <span className="text-xs font-bold uppercase tracking-wider text-amber-400">
              Bandeja de Autorización
            </span>
            <h3 className="text-base font-bold text-zinc-100">Pendiente Dictamen EHS</h3>
          </div>
          <div className="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400 group-hover:scale-105 transition-transform">
            <Clock className="w-5 h-5" />
          </div>
        </div>

        <div className="my-4">
          <div className="flex items-baseline gap-2">
            <span className="text-4xl font-extrabold font-mono tracking-tight text-white group-hover:text-amber-400 transition-colors">
              {pendingCount}
            </span>
            <span className="text-xs text-zinc-400">
              {pendingCount === 1 ? 'permiso en espera' : 'permisos en espera'}
            </span>
          </div>
        </div>

        <div className="pt-3 border-t border-zinc-800/80 text-xs text-zinc-400 flex items-center justify-between">
          <span>Checklist & Cuadrilla</span>
          <span className="font-mono text-amber-400 text-[11px] font-semibold">Prioridad Alta</span>
        </div>
      </div>

      {/* Celda Bento 3: Suspendidos & Cierres (3 cols) */}
      <div className="md:col-span-3 flex flex-col gap-3 justify-between">
        <div
          onClick={() => onSelectStatus?.(selectedStatus === 'suspended' ? '' : 'suspended')}
          className={`p-4 rounded-xl border transition-all cursor-pointer flex items-center justify-between ${
            selectedStatus === 'suspended'
              ? 'ring-2 ring-red-500 bg-red-950/30 border-red-500/50'
              : 'bg-zinc-950/80 border-zinc-800/80 hover:border-red-500/30'
          }`}
        >
          <div className="space-y-0.5">
            <span className="text-[11px] font-bold uppercase tracking-wider text-red-400">Suspendidos</span>
            <div className="text-2xl font-black font-mono text-white">{suspendedCount}</div>
          </div>
          <div className="p-2 rounded-lg bg-red-500/10 text-red-400">
            <AlertTriangle className="w-4 h-4" />
          </div>
        </div>

        <div
          onClick={() => onSelectStatus?.(selectedStatus === 'closed' ? '' : 'closed')}
          className={`p-4 rounded-xl border transition-all cursor-pointer flex items-center justify-between ${
            selectedStatus === 'closed'
              ? 'ring-2 ring-zinc-400 bg-zinc-900 border-zinc-500'
              : 'bg-zinc-950/80 border-zinc-800/80 hover:border-zinc-700'
          }`}
        >
          <div className="space-y-0.5">
            <span className="text-[11px] font-bold uppercase tracking-wider text-zinc-400">Cerrados / Salida</span>
            <div className="text-2xl font-black font-mono text-white">{closedCount}</div>
          </div>
          <div className="p-2 rounded-lg bg-zinc-800 text-zinc-300">
            <CheckCircle2 className="w-4 h-4" />
          </div>
        </div>
      </div>

      {/* Franja Inferior Bento: Distribución de Riesgos STPS (12 cols) */}
      <div className="md:col-span-12 p-4 rounded-xl bg-zinc-950/60 border border-zinc-800/80 flex flex-wrap items-center justify-between gap-4 text-xs">
        <div className="flex items-center gap-2 text-zinc-400 font-medium">
          <span className="w-2 h-2 rounded-full bg-amber-400" />
          <span>Matriz de Riesgo Activo (STPS México):</span>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-200">
            <Flame className="w-3.5 h-3.5 text-orange-400" />
            <span className="font-semibold text-zinc-300">Corte y Soldadura (NOM-027):</span>
            <span className="font-mono font-bold text-white ml-1">{hotWorkCount}</span>
          </div>

          <div className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-200">
            <ArrowUpRight className="w-3.5 h-3.5 text-blue-400" />
            <span className="font-semibold text-zinc-300">Alturas (NOM-009):</span>
            <span className="font-mono font-bold text-white ml-1">{heightsCount}</span>
          </div>

          <div className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-200">
            <Zap className="w-3.5 h-3.5 text-yellow-400" />
            <span className="font-semibold text-zinc-300">LOTO Eléctrico (NOM-029):</span>
            <span className="font-mono font-bold text-white ml-1">{lotoCount}</span>
          </div>

          <div className="flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-900 border border-zinc-800 text-zinc-200">
            <ShieldCheck className="w-3.5 h-3.5 text-purple-400" />
            <span className="font-semibold text-zinc-300">Espacios Confinados (NOM-033):</span>
            <span className="font-mono font-bold text-white ml-1">{confinedCount}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
