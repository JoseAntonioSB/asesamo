'use client';

import React, { useState, useMemo } from 'react';
import { WorkerComplianceStatus, PermitType } from '@/lib/api/types';
import { CheckCircle2, XCircle, AlertTriangle, Search, ShieldCheck, ShieldAlert, User, FileCheck } from 'lucide-react';

interface CrewComplianceSelectorProps {
  workers: WorkerComplianceStatus[];
  selectedWorkerIds: string[];
  permitType: PermitType;
  onChangeSelected: (selectedIds: string[]) => void;
  isLoading?: boolean;
}

export function CrewComplianceSelector({
  workers,
  selectedWorkerIds,
  permitType,
  onChangeSelected,
  isLoading = false,
}: CrewComplianceSelectorProps) {
  const [search, setSearch] = useState('');

  const hazardName = {
    heights: 'Trabajos en Alturas (NOM-009)',
    confined_space: 'Espacios Confinados (NOM-033)',
    hot_work: 'Corte y Soldadura (NOM-027)',
    electrical_loto: 'LOTO Eléctrico (NOM-029)',
  }[permitType] || 'Trabajos de Alto Riesgo';

  const filteredWorkers = useMemo(() => {
    if (!search.trim()) return workers;
    const q = search.toLowerCase().trim();
    return workers.filter((w) => {
      const matchName = (w.full_name || '').toLowerCase().includes(q);
      const matchCurp = (w.curp || '').toLowerCase().includes(q);
      const matchImss = (w.imss_number || '').toLowerCase().includes(q);
      const matchJob = (w.job_title || '').toLowerCase().includes(q);
      return matchName || matchCurp || matchImss || matchJob;
    });
  }, [workers, search]);

  const compliantCount = workers.filter((w) => w.is_compliant).length;
  const nonCompliantCount = workers.length - compliantCount;

  const toggleWorker = (workerId: string, isCompliant: boolean) => {
    if (!isCompliant) return; // Strict lock: Candado Inteligente blocks selection

    if (selectedWorkerIds.includes(workerId)) {
      onChangeSelected(selectedWorkerIds.filter((id) => id !== workerId));
    } else {
      onChangeSelected([...selectedWorkerIds, workerId]);
    }
  };

  const selectAllCompliant = () => {
    const compliantIds = workers.filter((w) => w.is_compliant).map((w) => w.worker_id);
    onChangeSelected(compliantIds);
  };

  const clearSelection = () => {
    onChangeSelected([]);
  };

  return (
    <div className="space-y-4">
      {/* Candado Inteligente banner */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-3 shadow-md">
        <div className="flex items-center gap-3">
          <div className="p-2.5 rounded-lg bg-cyan-950/70 border border-cyan-800/80 text-cyan-400">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-white">Candado Inteligente STPS</h4>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-cyan-950 text-cyan-300 border border-cyan-800">
                Auditoría en Tiempo Real
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">
              Validando DC-3 aplicable para <span className="text-cyan-300 font-medium">{hazardName}</span> y Aptitud Médica Vigente.
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 text-xs font-mono">
          <span className="px-2.5 py-1 bg-emerald-950/60 border border-emerald-700/60 text-emerald-300 rounded-lg">
            {compliantCount} Aptos
          </span>
          <span className="px-2.5 py-1 bg-rose-950/60 border border-rose-700/60 text-rose-300 rounded-lg">
            {nonCompliantCount} Bloqueados
          </span>
        </div>
      </div>

      {/* Search and Quick Selection Actions */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Buscar trabajador por nombre, CURP, NSS o puesto..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-slate-950/80 border border-slate-800 rounded-lg pl-9 pr-4 py-2 text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>

        <div className="flex items-center gap-2 shrink-0">
          <button
            type="button"
            onClick={selectAllCompliant}
            disabled={compliantCount === 0}
            className="text-xs font-medium text-emerald-400 hover:text-emerald-300 bg-emerald-950/40 border border-emerald-800/60 hover:bg-emerald-950/80 px-3 py-2 rounded-lg transition-colors disabled:opacity-50"
          >
            Seleccionar Todos los Aptos
          </button>
          {selectedWorkerIds.length > 0 && (
            <button
              type="button"
              onClick={clearSelection}
              className="text-xs font-medium text-slate-400 hover:text-slate-200 bg-slate-900 border border-slate-800 px-3 py-2 rounded-lg transition-colors"
            >
              Limpiar ({selectedWorkerIds.length})
            </button>
          )}
        </div>
      </div>

      {/* Workers List / Cards */}
      <div className="space-y-2.5">
        {isLoading ? (
          <div className="py-12 text-center text-slate-500 text-xs">
            <div className="inline-flex items-center gap-2">
              <div className="w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin" />
              <span>Verificando constancias DC-3 y exámenes médicos...</span>
            </div>
          </div>
        ) : filteredWorkers.length === 0 ? (
          <div className="py-8 text-center text-slate-500 text-xs bg-slate-900/40 rounded-xl border border-slate-800">
            No se encontraron trabajadores en la plantilla de la empresa.
          </div>
        ) : (
          filteredWorkers.map((worker) => {
            const isSelected = selectedWorkerIds.includes(worker.worker_id);
            const isCompliant = worker.is_compliant;

            return (
              <div
                key={worker.worker_id}
                onClick={() => toggleWorker(worker.worker_id, isCompliant)}
                className={`p-3.5 rounded-xl border transition-all duration-200 flex flex-col md:flex-row items-start md:items-center justify-between gap-3 ${
                  isCompliant
                    ? isSelected
                      ? 'bg-emerald-950/30 border-emerald-500/80 ring-1 ring-emerald-500 cursor-pointer shadow-md'
                      : 'bg-slate-900/70 border-slate-800 hover:border-slate-700 cursor-pointer'
                    : 'bg-rose-950/15 border-rose-900/40 opacity-80 cursor-not-allowed'
                }`}
              >
                {/* Worker Identity & Badges */}
                <div className="flex items-start gap-3">
                  <div className="pt-0.5">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      disabled={!isCompliant}
                      onChange={() => toggleWorker(worker.worker_id, isCompliant)}
                      className={`w-4 h-4 rounded border-slate-700 text-emerald-500 focus:ring-emerald-500/20 bg-slate-900 transition-colors ${
                        !isCompliant ? 'opacity-30 cursor-not-allowed' : 'cursor-pointer'
                      }`}
                    />
                  </div>

                  <div className="space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="text-sm font-semibold text-white">{worker.full_name}</span>
                      {worker.job_title && (
                        <span className="text-[11px] text-slate-400 px-2 py-0.5 rounded bg-slate-800 border border-slate-700">
                          {worker.job_title}
                        </span>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-3 text-[11px] text-slate-400 font-mono">
                      {worker.curp && <span>CURP: {worker.curp}</span>}
                      {worker.imss_number && <span>NSS: {worker.imss_number}</span>}
                    </div>

                    {/* Non-compliance explicit reasons */}
                    {!isCompliant && worker.blocking_reasons.length > 0 && (
                      <div className="mt-2 space-y-1 pt-1 border-t border-rose-900/30">
                        {worker.blocking_reasons.map((reason, idx) => (
                          <div
                            key={idx}
                            className="text-[11px] text-rose-300 flex items-center gap-1.5 font-medium"
                          >
                            <XCircle className="w-3.5 h-3.5 text-rose-400 shrink-0" />
                            <span>{reason}</span>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>

                {/* Right side compliance flag */}
                <div className="shrink-0 flex items-center gap-2 self-end md:self-center">
                  {isCompliant ? (
                    <div className="text-right">
                      <span className="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-300 bg-emerald-950/70 border border-emerald-700/70 px-2.5 py-1 rounded-full">
                        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                        <span>Apto para {permitType === 'heights' ? 'Alturas' : permitType === 'confined_space' ? 'Espacios Conf.' : permitType === 'hot_work' ? 'Soldadura' : 'LOTO'}</span>
                      </span>
                      <div className="text-[10px] text-emerald-400/80 mt-1 flex items-center justify-end gap-1">
                        <FileCheck className="w-3 h-3" />
                        <span>DC-3 + Examen Médico Acreditados</span>
                      </div>
                    </div>
                  ) : (
                    <div className="text-right">
                      <span className="inline-flex items-center gap-1.5 text-xs font-semibold text-rose-300 bg-rose-950/80 border border-rose-800/80 px-2.5 py-1 rounded-full">
                        <ShieldAlert className="w-3.5 h-3.5 text-rose-400" />
                        <span>Bloqueado por Candado STPS</span>
                      </span>
                      <div className="text-[10px] text-rose-400/70 mt-1">
                        Asignación Prohibida por Normativa
                      </div>
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Selected Crew Summary Footer */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-3 flex items-center justify-between text-xs">
        <span className="text-slate-400">
          Cuadrilla conformada:{' '}
          <strong className="text-emerald-400 font-mono text-sm">{selectedWorkerIds.length}</strong> operarios seleccionados
        </span>
        {selectedWorkerIds.length === 0 && (
          <span className="text-amber-400 flex items-center gap-1 text-[11px]">
            <AlertTriangle className="w-3.5 h-3.5" />
            Se requiere al menos 1 trabajador apto
          </span>
        )}
      </div>
    </div>
  );
}
