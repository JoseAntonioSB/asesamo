// asesamo-web/components/caseta/worker-card.tsx
'use client';

import React, { useState } from 'react';
import {
  Building2,
  CreditCard,
  Briefcase,
  Copy,
  Check,
  ShieldCheck,
} from 'lucide-react';
import { WorkerSummary, ContractorSummary, PlantSummary, AccessResult } from '@/lib/types/caseta';

export interface WorkerCardProps {
  worker: WorkerSummary;
  contractor: ContractorSummary;
  plant?: PlantSummary;
  result?: AccessResult;
}

export function WorkerCard({ worker, contractor, plant, result = 'granted' }: WorkerCardProps) {
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const copyToClipboard = (text: string, fieldName: string) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedField(fieldName);
    setTimeout(() => setCopiedField(null), 2000);
  };

  // Status ring color
  const ringColor =
    result === 'granted'
      ? 'border-emerald-500 shadow-emerald-500/20'
      : result === 'granted_with_warning'
      ? 'border-amber-500 shadow-amber-500/20'
      : 'border-red-500 shadow-red-500/20';

  // Compute initials
  const initials = worker.full_name
    ? worker.full_name
        .split(' ')
        .filter(Boolean)
        .slice(0, 2)
        .map((n) => n[0])
        .join('')
        .toUpperCase()
    : 'TR';

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-slate-900/95 p-6 text-white shadow-xl">
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-5 pb-6 border-b border-slate-800">
        {/* Worker Avatar / Photo placeholder */}
        <div className="relative">
          <div
            className={`flex h-20 w-20 md:h-24 md:w-24 shrink-0 items-center justify-center rounded-2xl border-4 ${ringColor} bg-slate-800 text-2xl md:text-3xl font-black text-slate-100 shadow-xl`}
          >
            {initials}
          </div>
          <div className="absolute -bottom-2 -right-2 rounded-full border-2 border-slate-900 bg-emerald-600 px-2 py-0.5 text-[10px] font-bold text-white uppercase tracking-wider">
            Activo
          </div>
        </div>

        {/* Worker Primary Identity */}
        <div className="flex-1 space-y-1">
          <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-800 text-slate-300 border border-slate-700">
            <Briefcase className="h-3 w-3 text-cyan-400" />
            {worker.job_title || 'Trabajador Especializado'}
          </div>
          <h2 className="text-2xl md:text-3xl font-extrabold text-white tracking-tight">
            {worker.full_name}
          </h2>
          <div className="flex flex-wrap items-center gap-2 text-xs text-slate-400">
            <span className="flex items-center gap-1 text-slate-300 font-medium">
              <Building2 className="h-3.5 w-3.5 text-slate-400" />
              {contractor.name}
            </span>
            {plant && (
              <>
                <span className="text-slate-600">•</span>
                <span>{plant.name}</span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Identifiers Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5 pt-6">
        {/* CURP */}
        <div className="flex flex-col justify-between rounded-xl border border-slate-800 bg-slate-950/60 p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="font-medium uppercase tracking-wider">CURP</span>
            <button
              onClick={() => copyToClipboard(worker.curp || '', 'curp')}
              className="text-slate-400 hover:text-white transition-colors cursor-pointer"
              title="Copiar CURP"
            >
              {copiedField === 'curp' ? (
                <Check className="h-3.5 w-3.5 text-emerald-400" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
            </button>
          </div>
          <p className="font-mono text-sm sm:text-base font-bold text-slate-100 tracking-wider">
            {worker.curp || 'NO REGISTRADO'}
          </p>
        </div>

        {/* NSS / IMSS */}
        <div className="flex flex-col justify-between rounded-xl border border-slate-800 bg-slate-950/60 p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="font-medium uppercase tracking-wider">NSS (IMSS)</span>
            <button
              onClick={() => copyToClipboard(worker.imss_number || '', 'imss')}
              className="text-slate-400 hover:text-white transition-colors cursor-pointer"
              title="Copiar NSS"
            >
              {copiedField === 'imss' ? (
                <Check className="h-3.5 w-3.5 text-emerald-400" />
              ) : (
                <Copy className="h-3.5 w-3.5" />
              )}
            </button>
          </div>
          <p className="font-mono text-sm sm:text-base font-bold text-slate-100 tracking-wider">
            {worker.imss_number || 'NO REGISTRADO'}
          </p>
        </div>

        {/* RFC Contratista */}
        <div className="flex flex-col justify-between rounded-xl border border-slate-800 bg-slate-950/60 p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="font-medium uppercase tracking-wider">RFC Empresa</span>
            <CreditCard className="h-3.5 w-3.5 text-slate-500" />
          </div>
          <p className="font-mono text-sm font-semibold text-slate-200">
            {contractor.rfc || 'NO REGISTRADO'}
          </p>
        </div>

        {/* REPSE Contratista */}
        <div className="flex flex-col justify-between rounded-xl border border-slate-800 bg-slate-950/60 p-3.5">
          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
            <span className="font-medium uppercase tracking-wider">Registro REPSE</span>
            <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
          </div>
          <p className="font-mono text-sm font-semibold text-emerald-300">
            {contractor.repse_number || 'EN TRÁMITE'}
          </p>
        </div>
      </div>

      {/* UUID Footer Bar */}
      <div className="mt-4 flex items-center justify-between rounded-lg bg-slate-950/40 px-3 py-2 text-[11px] text-slate-400 border border-slate-800/60">
        <span className="truncate">UUID: {worker.user_id}</span>
        <button
          onClick={() => copyToClipboard(worker.user_id, 'uuid')}
          className="ml-2 inline-flex items-center gap-1 font-semibold text-slate-300 hover:text-white cursor-pointer"
        >
          {copiedField === 'uuid' ? 'Copiado' : 'Copiar'}
        </button>
      </div>
    </div>
  );
}
