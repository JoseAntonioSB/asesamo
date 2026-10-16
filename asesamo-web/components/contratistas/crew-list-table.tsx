'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { ContractorWorkerItem } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  HardHat,
  Search,
  CheckCircle2,
  XCircle,
  FileText,
  QrCode,
  ShieldCheck,
  UserCheck,
  UserX,
  Loader2,
  ExternalLink,
  Camera,
} from 'lucide-react';
import { apiClient } from '@/lib/api/client';
import { WorkerPhotoModal } from './worker-photo-modal';

interface CrewListTableProps {
  workers: ContractorWorkerItem[];
  isLoading: boolean;
  onToggleStatus: (workerId: string, currentActive: boolean) => Promise<void>;
  canManageWorkers: boolean;
  contractorName?: string;
  onRefresh?: () => void;
}

export function CrewListTable({
  workers,
  isLoading,
  onToggleStatus,
  canManageWorkers,
  contractorName,
  onRefresh,
}: CrewListTableProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all');
  const [togglingId, setTogglingId] = useState<string | null>(null);
  const [photoModalWorker, setPhotoModalWorker] = useState<ContractorWorkerItem | null>(null);
  const [avatarRefreshKey, setAvatarRefreshKey] = useState<number>(Date.now());

  const filtered = workers.filter((w) => {
    const matchesSearch =
      w.full_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      w.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (w.job_title && w.job_title.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (w.curp && w.curp.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (w.imss_number && w.imss_number.includes(searchTerm));

    const matchesStatus =
      statusFilter === 'all' ||
      (statusFilter === 'active' && w.is_active) ||
      (statusFilter === 'inactive' && !w.is_active);

    return matchesSearch && matchesStatus;
  });

  const handleToggle = async (workerId: string, currentActive: boolean) => {
    try {
      setTogglingId(workerId);
      await onToggleStatus(workerId, currentActive);
    } finally {
      setTogglingId(null);
    }
  };

  return (
    <div className="space-y-4">
      {/* Search and Filters */}
      <div className="flex flex-col sm:flex-row gap-3 items-stretch sm:items-center justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <Input
            type="text"
            placeholder="Buscar por Nombre, Puesto, CURP o NSS..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9 h-9 bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs"
          />
        </div>

        <div className="flex items-center gap-2">
          <label className="text-xs text-slate-400 whitespace-nowrap">Filtrar cuadrilla:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as 'all' | 'active' | 'inactive')}
            className="h-9 rounded-md bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 focus:outline-none focus:ring-1 focus:ring-cyan-500"
          >
            <option value="all">Todos ({workers.length})</option>
            <option value="active">Activos</option>
            <option value="inactive">Inactivos</option>
          </select>
        </div>
      </div>

      {/* Table Container */}
      <div className="rounded-xl border border-slate-800 bg-slate-900/60 overflow-hidden shadow-lg backdrop-blur-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs text-slate-300">
            <thead className="bg-slate-950/80 text-[11px] font-semibold uppercase tracking-wider text-slate-400 border-b border-slate-800">
              <tr>
                <th scope="col" className="px-4 py-3">Trabajador / Cuenta</th>
                <th scope="col" className="px-4 py-3">Puesto / Especialidad</th>
                <th scope="col" className="px-4 py-3">CURP y NSS (IMSS)</th>
                <th scope="col" className="px-4 py-3 text-center">Expediente</th>
                <th scope="col" className="px-4 py-3">Estatus</th>
                <th scope="col" className="px-4 py-3 text-right">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/80">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <div className="w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin" />
                      <span>Cargando trabajadores de la cuadrilla...</span>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-10 text-center text-slate-400">
                    <HardHat className="w-8 h-8 text-slate-600 mx-auto mb-2" />
                    <p className="font-medium text-slate-300">No hay trabajadores registrados en esta cuadrilla</p>
                    <p className="text-[11px] text-slate-500 mt-0.5">
                      {searchTerm
                        ? 'No se encontraron coincidencias para la búsqueda.'
                        : 'Utilice el botón "Registrar Trabajador" para incorporar personal a la plantilla.'}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((w) => {
                  const isToggling = togglingId === w.worker_id;

                  return (
                    <tr
                      key={w.worker_id}
                      className="hover:bg-slate-800/40 transition-colors group"
                    >
                      {/* Name & Account */}
                      <td className="px-4 py-3.5">
                        <div className="flex items-center gap-3">
                          <button
                            type="button"
                            onClick={() => canManageWorkers && setPhotoModalWorker(w)}
                            disabled={!canManageWorkers}
                            title={canManageWorkers ? 'Subir o cambiar fotografía del trabajador' : undefined}
                            className={`relative group/avatar w-10 h-10 rounded-xl bg-slate-800 border border-slate-700 overflow-hidden shrink-0 flex items-center justify-center font-bold text-xs text-white ${
                              canManageWorkers ? 'hover:border-emerald-500 cursor-pointer' : 'cursor-default'
                            }`}
                          >
                            <img
                              src={`${apiClient.getUserAvatarUrl(w.worker_id)}?t=${avatarRefreshKey}`}
                              alt={w.full_name}
                              className="w-full h-full object-cover"
                              onError={(e) => {
                                (e.target as HTMLElement).style.display = 'none';
                              }}
                            />
                            <span className="text-[11px] font-black text-slate-300">
                              {w.full_name
                                .split(' ')
                                .filter(Boolean)
                                .slice(0, 2)
                                .map((n) => n[0])
                                .join('')
                                .toUpperCase()}
                            </span>
                            {canManageWorkers && (
                              <div className="absolute inset-0 bg-slate-950/70 backdrop-blur-xs opacity-0 group-hover/avatar:opacity-100 flex items-center justify-center transition-opacity">
                                <Camera className="w-4 h-4 text-emerald-400" />
                              </div>
                            )}
                          </button>
                          <div>
                            <span className="font-semibold text-white group-hover:text-cyan-400 transition-colors block">
                              {w.full_name}
                            </span>
                            <span className="text-[11px] text-slate-400 font-mono">
                              {w.email}
                            </span>
                            {w.phone && (
                              <div className="text-[10px] text-slate-500">
                                Tel: {w.phone}
                              </div>
                            )}
                          </div>
                        </div>
                      </td>

                      {/* Job Title */}
                      <td className="px-4 py-3.5">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-[11px] bg-slate-800 text-slate-200 border border-slate-700">
                          {w.job_title || 'Operario General'}
                        </span>
                      </td>

                      {/* CURP & NSS */}
                      <td className="px-4 py-3.5 font-mono">
                        <div className="space-y-0.5">
                          <div className="text-slate-300 text-[11px]">
                            CURP: <span className="text-white font-medium">{w.curp || 'No registrada'}</span>
                          </div>
                          <div className="text-slate-400 text-[11px]">
                            NSS: <span className="text-cyan-300/90 font-medium">{w.imss_number || 'No registrado'}</span>
                          </div>
                        </div>
                      </td>

                      {/* Documents / Compliance */}
                      <td className="px-4 py-3.5 text-center">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-semibold ${
                            w.documents_count > 0
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                          }`}
                        >
                          <FileText className="w-3 h-3" />
                          {w.documents_count} doc{w.documents_count === 1 ? '' : 's'}
                        </span>
                      </td>

                      {/* Status */}
                      <td className="px-4 py-3.5">
                        {w.is_active ? (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                            <CheckCircle2 className="w-3 h-3" />
                            Activo
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-slate-800 text-slate-400 border border-slate-700">
                            <XCircle className="w-3 h-3" />
                            Inactivo
                          </span>
                        )}
                      </td>

                      {/* Actions */}
                      <td className="px-4 py-3.5 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          {/* Ver Credencial */}
                          <Link href={`/credencial?worker_id=${encodeURIComponent(w.worker_id)}`}>
                            <Button
                              size="sm"
                              variant="outline"
                              className="border-slate-700 text-cyan-300 hover:text-white hover:bg-slate-800 text-xs h-7 px-2 gap-1"
                              title="Ver Credencial Móvil QR"
                            >
                              <QrCode className="w-3 h-3 text-cyan-400" />
                              <span className="hidden sm:inline">Credencial</span>
                            </Button>
                          </Link>

                          {/* Subir DC-3 / Expediente */}
                          <Link href={`/documentos?worker_id=${encodeURIComponent(w.worker_id)}`}>
                            <Button
                              size="sm"
                              variant="outline"
                              className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs h-7 px-2 gap-1"
                              title="Cargar o ver documentos de este trabajador"
                            >
                              <FileText className="w-3 h-3 text-amber-400" />
                              <span className="hidden sm:inline">Expediente</span>
                            </Button>
                          </Link>

                          {/* Subir / Actualizar Fotografía */}
                          {canManageWorkers && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => setPhotoModalWorker(w)}
                              className="border-slate-700 text-slate-300 hover:text-emerald-300 hover:bg-slate-800 text-xs h-7 px-2 gap-1"
                              title="Subir o cambiar fotografía del trabajador"
                            >
                              <Camera className="w-3 h-3 text-emerald-400" />
                              <span className="hidden lg:inline">Foto</span>
                            </Button>
                          )}

                          {/* Toggle Active Switch */}
                          {canManageWorkers && (
                            <Button
                              size="sm"
                              variant="outline"
                              disabled={isToggling}
                              onClick={() => handleToggle(w.worker_id, w.is_active)}
                              className={`text-xs h-7 px-2 gap-1 border ${
                                w.is_active
                                  ? 'border-red-900/60 bg-red-950/20 text-red-300 hover:bg-red-900/40'
                                  : 'border-emerald-900/60 bg-emerald-950/20 text-emerald-300 hover:bg-emerald-900/40'
                              }`}
                              title={w.is_active ? 'Desactivar en cuadrilla' : 'Reactivar en cuadrilla'}
                            >
                              {isToggling ? (
                                <Loader2 className="w-3 h-3 animate-spin" />
                              ) : w.is_active ? (
                                <>
                                  <UserX className="w-3 h-3" />
                                  <span className="hidden md:inline">Baja</span>
                                </>
                              ) : (
                                <>
                                  <UserCheck className="w-3 h-3" />
                                  <span className="hidden md:inline">Alta</span>
                                </>
                              )}
                            </Button>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Worker Photo Upload Modal */}
      <WorkerPhotoModal
        isOpen={!!photoModalWorker}
        onClose={() => setPhotoModalWorker(null)}
        worker={photoModalWorker}
        onPhotoUploaded={(workerId) => {
          setAvatarRefreshKey(Date.now());
          onRefresh?.();
        }}
      />
    </div>
  );
}
