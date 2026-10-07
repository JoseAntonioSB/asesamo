// asesamo-web/components/caseta/quick-fill-presets.tsx
'use client';

import React, { useState } from 'react';
import { Sparkles, CheckCircle2, AlertTriangle, XCircle, Database } from 'lucide-react';

export interface QuickFillPresetsProps {
  onSelectPreset: (workerId: string) => void;
  disabled?: boolean;
}

interface PresetOption {
  id: string;
  name: string;
  role: string;
  uuid: string;
  tier: 'granted' | 'warning' | 'denied';
  tierLabel: string;
  summary: string;
  badgeClass: string;
  borderClass: string;
}

export function QuickFillPresets({ onSelectPreset, disabled = false }: QuickFillPresetsProps) {
  const [showSeedMappings, setShowSeedMappings] = useState(false);

  // 1. Primary demo presets as defined in dispatch requirements
  const primaryPresets: PresetOption[] = [
    {
      id: 'carlos',
      name: 'Carlos Hernández',
      role: 'Soldador Industrial',
      uuid: '66666666-6666-6666-6666-666666666661',
      tier: 'granted',
      tierLabel: 'Aprobado',
      summary: 'DC-3, INE y Examen Médico Vigentes (100% Cumplido)',
      badgeClass: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
      borderClass: 'border-emerald-500/30 hover:border-emerald-500 hover:bg-emerald-950/20',
    },
    {
      id: 'maria',
      name: 'María Rodríguez',
      role: 'Técnico Mecánico',
      uuid: '66666666-6666-6666-6666-666666666662',
      tier: 'warning',
      tierLabel: 'Advertencia',
      summary: 'Constancia DC-3 próxima a vencer (< 30 días)',
      badgeClass: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
      borderClass: 'border-amber-500/30 hover:border-amber-500 hover:bg-amber-950/20',
    },
    {
      id: 'jorge',
      name: 'Jorge Treviño',
      role: 'Electricista Industrial',
      uuid: '66666666-6666-6666-6666-666666666663',
      tier: 'denied',
      tierLabel: 'Denegado',
      summary: 'DC-3 vencido hace 35 días y Médico rechazado',
      badgeClass: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
      borderClass: 'border-rose-500/30 hover:border-rose-500 hover:bg-rose-950/20',
    },
  ];

  // 2. Exact database seed matches (supporting live Postgres evaluations)
  const seedPresets: PresetOption[] = [
    {
      id: 'seed-carlos',
      name: 'Carlos Mendoza (DB)',
      role: 'Soldador Industrial',
      uuid: '66666666-6666-6666-6666-666666666661',
      tier: 'granted',
      tierLabel: 'DB: Aprobado',
      summary: 'Todos los documentos en regla',
      badgeClass: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
      borderClass: 'border-emerald-500/30 hover:border-emerald-500 hover:bg-emerald-950/20',
    },
    {
      id: 'seed-roberto',
      name: 'Roberto García (DB)',
      role: 'Electricista Industrial',
      uuid: '66666666-6666-6666-6666-666666666663',
      tier: 'warning',
      tierLabel: 'DB: Advertencia (5d)',
      summary: 'DC-3 por vencer en 5 días',
      badgeClass: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
      borderClass: 'border-amber-500/30 hover:border-amber-500 hover:bg-amber-950/20',
    },
    {
      id: 'seed-juan',
      name: 'Juan Pérez (DB)',
      role: 'Técnico Mecánico',
      uuid: '66666666-6666-6666-6666-666666666662',
      tier: 'denied',
      tierLabel: 'DB: Denegado',
      summary: 'DC-3 vencido y Médico rechazado',
      badgeClass: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
      borderClass: 'border-rose-500/30 hover:border-rose-500 hover:bg-rose-950/20',
    },
  ];

  const activeList = showSeedMappings ? seedPresets : primaryPresets;

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-slate-900/60 p-4 shadow-lg backdrop-blur">
      <div className="flex flex-wrap items-center justify-between gap-2 pb-3 mb-3 border-b border-slate-800">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-amber-400" />
          <span className="text-xs font-bold text-slate-200 uppercase tracking-wider">
            Casos de Prueba Rápidos (Quick-Fill Demo)
          </span>
        </div>
        <button
          type="button"
          onClick={() => setShowSeedMappings(!showSeedMappings)}
          className="inline-flex items-center gap-1 text-[11px] text-slate-400 hover:text-cyan-300 transition-colors cursor-pointer"
        >
          <Database className="h-3 w-3" />
          <span>{showSeedMappings ? 'Ver nombres demo' : 'Ver nombres de BD semilla'}</span>
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {activeList.map((preset) => (
          <button
            key={preset.id}
            type="button"
            disabled={disabled}
            onClick={() => onSelectPreset(preset.uuid)}
            className={`group flex flex-col justify-between rounded-xl border ${preset.borderClass} bg-slate-950/80 p-3.5 text-left transition-all active:scale-[0.98] disabled:opacity-40 shadow-sm cursor-pointer`}
          >
            <div>
              <div className="flex items-center justify-between gap-1 mb-1.5">
                <span className="text-sm font-bold text-white group-hover:text-cyan-200 transition-colors">
                  {preset.name}
                </span>
                <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${preset.badgeClass}`}>
                  {preset.tier === 'granted' && <CheckCircle2 className="h-3 w-3" />}
                  {preset.tier === 'warning' && <AlertTriangle className="h-3 w-3" />}
                  {preset.tier === 'denied' && <XCircle className="h-3 w-3" />}
                  {preset.tierLabel}
                </span>
              </div>
              <p className="text-xs text-slate-400 line-clamp-1">
                {preset.summary}
              </p>
            </div>

            <div className="mt-3 pt-2 border-t border-slate-800/80 flex items-center justify-between text-[11px] font-mono text-slate-400">
              <span className="truncate">...{preset.uuid.slice(-12)}</span>
              <span className="text-cyan-400 font-sans font-semibold group-hover:translate-x-0.5 transition-transform">
                Escanear →
              </span>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
