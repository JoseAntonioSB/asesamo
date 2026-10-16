'use client';

import React, { useState, useMemo } from 'react';
import Link from 'next/link';
import { WorkPermitItem, PermitType, PermitStatus } from '@/lib/api/types';
import { PermitStatusBadge } from './permit-status-badge';
import { PermitHazardBadge } from './permit-hazard-badge';
import { PermitCountdownTimer } from './permit-countdown-timer';
import { Search, Filter, Users, MapPin, ChevronRight, AlertCircle, ArrowUpDown } from 'lucide-react';

interface PermitListTableProps {
  permits: WorkPermitItem[];
  isLoading?: boolean;
  selectedStatus?: string;
  onStatusChange?: (status: string) => void;
  onRefresh?: () => void;
}

export function PermitListTable({
  permits,
  isLoading = false,
  selectedStatus = '',
  onStatusChange,
  onRefresh,
}: PermitListTableProps) {
  const [search, setSearch] = useState('');
  const [selectedHazard, setSelectedHazard] = useState<string>('');
  const [selectedContractor, setSelectedContractor] = useState<string>('');
  const [sortField, setSortField] = useState<'created_at' | 'permit_number' | 'end_time'>('created_at');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');

  // Extract unique contractors for filter dropdown
  const uniqueContractors = useMemo(() => {
    const set = new Set<string>();
    permits.forEach((p) => {
      if (p.contractor_name) set.add(p.contractor_name);
    });
    return Array.from(set).sort();
  }, [permits]);

  // Filter and sort permits
  const filteredPermits = useMemo(() => {
    return permits
      .filter((p) => {
        // Status filter
        if (selectedStatus && p.status !== selectedStatus) {
          return false;
        }

        // Hazard filter
        if (selectedHazard && p.permit_type !== selectedHazard) {
          return false;
        }

        // Contractor filter
        if (selectedContractor && p.contractor_name !== selectedContractor) {
          return false;
        }

        // Search text
        if (search.trim()) {
          const query = search.toLowerCase().trim();
          const matchFolio = (p.permit_number || '').toLowerCase().includes(query);
          const matchContractor = (p.contractor_name || '').toLowerCase().includes(query);
          const matchArea = (p.work_area || '').toLowerCase().includes(query);
          const matchDesc = (p.description || '').toLowerCase().includes(query);
          if (!matchFolio && !matchContractor && !matchArea && !matchDesc) {
            return false;
          }
        }

        return true;
      })
      .sort((a, b) => {
        let valA = a[sortField] || '';
        let valB = b[sortField] || '';
        if (valA < valB) return sortDirection === 'asc' ? -1 : 1;
        if (valA > valB) return sortDirection === 'asc' ? 1 : -1;
        return 0;
      });
  }, [permits, selectedStatus, selectedHazard, selectedContractor, search, sortField, sortDirection]);

  const toggleSort = (field: 'created_at' | 'permit_number' | 'end_time') => {
    if (sortField === field) {
      setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  return (
    <div className="space-y-4">
      {/* Search and Filters Bar */}
      <div className="bg-slate-900/80 border border-slate-800 p-4 rounded-xl backdrop-blur-sm flex flex-col md:flex-row gap-3 items-stretch md:items-center justify-between">
        <div className="relative flex-1">
          <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Buscar por folio, contratista, área o descripción..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-slate-950/80 border border-slate-800 rounded-lg pl-9 pr-4 py-2 text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition-colors"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Hazard Filter */}
          <div className="relative min-w-[170px]">
            <select
              value={selectedHazard}
              onChange={(e) => setSelectedHazard(e.target.value)}
              className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-300 focus:outline-none focus:border-cyan-500 cursor-pointer"
            >
              <option value="">Todos los Riesgos STPS</option>
              <option value="heights">Alturas (NOM-009)</option>
              <option value="confined_space">Espacios Confinados (NOM-033)</option>
              <option value="hot_work">Corte y Soldadura (NOM-027)</option>
              <option value="electrical_loto">LOTO Eléctrico (NOM-029)</option>
            </select>
          </div>

          {/* Status Filter */}
          <div className="relative min-w-[150px]">
            <select
              value={selectedStatus}
              onChange={(e) => onStatusChange?.(e.target.value)}
              className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-300 focus:outline-none focus:border-cyan-500 cursor-pointer"
            >
              <option value="">Todos los Estatus</option>
              <option value="draft">Borrador</option>
              <option value="pending_approval">Pendiente EHS</option>
              <option value="approved_active">Activo en Campo</option>
              <option value="suspended">Suspendido</option>
              <option value="closed">Cerrado</option>
              <option value="cancelled">Cancelado</option>
            </select>
          </div>

          {/* Contractor Filter */}
          {uniqueContractors.length > 0 && (
            <div className="relative min-w-[160px]">
              <select
                value={selectedContractor}
                onChange={(e) => setSelectedContractor(e.target.value)}
                className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-300 focus:outline-none focus:border-cyan-500 cursor-pointer"
              >
                <option value="">Todas las Empresas</option>
                {uniqueContractors.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </div>
          )}

          {(search || selectedHazard || selectedStatus || selectedContractor) && (
            <button
              type="button"
              onClick={() => {
                setSearch('');
                setSelectedHazard('');
                setSelectedContractor('');
                onStatusChange?.('');
              }}
              className="text-xs text-cyan-400 hover:text-cyan-300 px-2 py-1.5 transition-colors"
            >
              Limpiar filtros
            </button>
          )}
        </div>
      </div>

      {/* Table Container */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-slate-800 bg-slate-950/70 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                <th
                  className="py-3 px-4 cursor-pointer hover:text-white transition-colors"
                  onClick={() => toggleSort('permit_number')}
                >
                  <div className="flex items-center gap-1.5">
                    <span>Folio PTAR</span>
                    <ArrowUpDown className="w-3 h-3" />
                  </div>
                </th>
                <th className="py-3 px-4">Riesgo Regulado STPS</th>
                <th className="py-3 px-4">Contratista y Área</th>
                <th
                  className="py-3 px-4 cursor-pointer hover:text-white transition-colors"
                  onClick={() => toggleSort('end_time')}
                >
                  <div className="flex items-center gap-1.5">
                    <span>Validez y Horario</span>
                    <ArrowUpDown className="w-3 h-3" />
                  </div>
                </th>
                <th className="py-3 px-4">Cuadrilla</th>
                <th className="py-3 px-4">Estatus</th>
                <th className="py-3 px-4 text-right">Acción</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 text-xs text-slate-300">
              {isLoading ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center text-slate-500">
                    <div className="inline-flex items-center gap-2">
                      <div className="w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin" />
                      <span>Cargando permisos de trabajo...</span>
                    </div>
                  </td>
                </tr>
              ) : filteredPermits.length === 0 ? (
                <tr>
                  <td colSpan={7} className="py-12 text-center">
                    <div className="flex flex-col items-center justify-center gap-2">
                      <AlertCircle className="w-8 h-8 text-slate-600" />
                      <p className="text-sm font-medium text-slate-400">No se encontraron permisos de alto riesgo</p>
                      <p className="text-xs text-slate-500 max-w-sm">
                        {search || selectedHazard || selectedStatus
                          ? 'Ajuste los filtros de búsqueda o consulte todos los estatus.'
                          : 'Comience creando un nuevo permiso de trabajo con el botón superior.'}
                      </p>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredPermits.map((permit) => {
                  const workerCount =
                    permit.worker_count !== undefined
                      ? permit.worker_count
                      : permit.workers?.length || 0;

                  return (
                    <tr
                      key={permit.id}
                      className="hover:bg-slate-800/40 transition-colors group cursor-pointer"
                    >
                      {/* Folio */}
                      <td className="py-3.5 px-4 font-mono font-bold text-white whitespace-nowrap">
                        <Link
                          href={`/permisos/${permit.id}`}
                          className="hover:text-cyan-400 flex items-center gap-1 transition-colors"
                        >
                          <span>{permit.permit_number}</span>
                        </Link>
                      </td>

                      {/* Hazard Type */}
                      <td className="py-3.5 px-4 whitespace-nowrap">
                        <PermitHazardBadge hazardType={permit.permit_type} size="sm" />
                      </td>

                      {/* Contractor & Location */}
                      <td className="py-3.5 px-4">
                        <div className="font-medium text-slate-200">
                          {permit.contractor_name || 'Empresa Contratista'}
                        </div>
                        <div className="text-[11px] text-slate-400 flex items-center gap-1 mt-0.5">
                          <MapPin className="w-3 h-3 text-slate-500 shrink-0" />
                          <span className="truncate max-w-[200px]">{permit.work_area}</span>
                        </div>
                      </td>

                      {/* Validity & Countdown */}
                      <td className="py-3.5 px-4 whitespace-nowrap">
                        <PermitCountdownTimer
                          startTime={permit.start_time}
                          endTime={permit.end_time}
                          status={permit.status}
                        />
                        <div className="text-[10px] text-slate-500 mt-1">
                          {new Intl.DateTimeFormat('es-MX', {
                            dateStyle: 'short',
                            timeStyle: 'short',
                          }).format(new Date(permit.start_time))}{' '}
                          -{' '}
                          {new Intl.DateTimeFormat('es-MX', {
                            timeStyle: 'short',
                          }).format(new Date(permit.end_time))}
                        </div>
                      </td>

                      {/* Crew Count */}
                      <td className="py-3.5 px-4 whitespace-nowrap">
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-slate-800 rounded text-slate-300 font-mono text-[11px]">
                          <Users className="w-3 h-3 text-slate-400" />
                          {workerCount} {workerCount === 1 ? 'operario' : 'operarios'}
                        </span>
                      </td>

                      {/* Status Badge */}
                      <td className="py-3.5 px-4 whitespace-nowrap">
                        <PermitStatusBadge status={permit.status} size="sm" />
                      </td>

                      {/* Actions */}
                      <td className="py-3.5 px-4 text-right whitespace-nowrap">
                        <Link
                          href={`/permisos/${permit.id}`}
                          className="inline-flex items-center gap-1 text-xs text-cyan-400 hover:text-cyan-300 bg-cyan-950/40 hover:bg-cyan-950/70 border border-cyan-800/60 px-3 py-1.5 rounded-lg font-medium transition-all group-hover:border-cyan-600"
                        >
                          <span>Ver Detalle</span>
                          <ChevronRight className="w-3.5 h-3.5" />
                        </Link>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Footer info */}
        <div className="px-4 py-3 bg-slate-950/80 border-t border-slate-800 text-[11px] text-slate-500 flex items-center justify-between">
          <span>
            Mostrando {filteredPermits.length} de {permits.length} permisos de alto riesgo
          </span>
          <span className="font-mono text-slate-600">STPS NOM Compliance Engine v1.0</span>
        </div>
      </div>
    </div>
  );
}
