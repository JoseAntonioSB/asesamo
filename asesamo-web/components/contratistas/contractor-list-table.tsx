'use client';

import React, { useState } from 'react';
import { ContractorItem, PlantContractorStatus } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Building2,
  Search,
  Users,
  ShieldCheck,
  ShieldAlert,
  AlertTriangle,
  Ban,
  CheckCircle2,
  Clock,
  MoreVertical,
  ChevronRight,
} from 'lucide-react';

interface ContractorListTableProps {
  contractors: ContractorItem[];
  isLoading: boolean;
  onSelectContractor: (contractor: ContractorItem) => void;
  onUpdateStatus: (contractorId: string, newStatus: PlantContractorStatus) => Promise<void>;
  canManageStatus: boolean;
}

export function ContractorListTable({
  contractors,
  isLoading,
  onSelectContractor,
  onUpdateStatus,
  canManageStatus,
}: ContractorListTableProps) {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const filtered = contractors.filter((c) => {
    const matchesSearch =
      c.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (c.rfc && c.rfc.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (c.repse_number && c.repse_number.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (c.admin_email && c.admin_email.toLowerCase().includes(searchTerm.toLowerCase()));

    const status = c.plant_status || 'active';
    const matchesStatus = statusFilter === 'all' || status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  const handleStatusChange = async (contractorId: string, status: PlantContractorStatus) => {
    try {
      setActionLoadingId(contractorId);
      await onUpdateStatus(contractorId, status);
    } finally {
      setActionLoadingId(null);
    }
  };

  const getStatusBadge = (status?: string) => {
    switch (status) {
      case 'active':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" />
            Activo
          </span>
        );
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Clock className="w-3 h-3" />
            Pendiente
          </span>
        );
      case 'suspended':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-orange-500/10 text-orange-400 border border-orange-500/20">
            <AlertTriangle className="w-3 h-3" />
            Suspendido
          </span>
        );
      case 'blocked':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-red-500/10 text-red-400 border border-red-500/20">
            <Ban className="w-3 h-3" />
            Bloqueado
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-slate-800 text-slate-300 border border-slate-700">
            {status || 'Desconocido'}
          </span>
        );
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
            placeholder="Buscar por Razón Social, RFC o REPSE..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-9 h-9 bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs"
          />
        </div>

        <div className="flex items-center gap-2">
          <label className="text-xs text-slate-400 whitespace-nowrap">Filtrar por estatus:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="h-9 rounded-md bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          >
            <option value="all">Todos ({contractors.length})</option>
            <option value="active">Activos</option>
            <option value="pending">Pendientes</option>
            <option value="suspended">Suspendidos</option>
            <option value="blocked">Bloqueados</option>
          </select>
        </div>
      </div>

      {/* Table Container */}
      <div className="rounded-xl border border-slate-800 bg-slate-900/60 overflow-hidden shadow-lg backdrop-blur-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs text-slate-300">
            <thead className="bg-slate-950/80 text-[11px] font-semibold uppercase tracking-wider text-slate-400 border-b border-slate-800">
              <tr>
                <th scope="col" className="px-4 py-3">Empresa / Razón Social</th>
                <th scope="col" className="px-4 py-3">Identificadores Fiscales</th>
                <th scope="col" className="px-4 py-3">Contacto Administrador</th>
                <th scope="col" className="px-4 py-3 text-center">Cuadrilla</th>
                <th scope="col" className="px-4 py-3">Estatus</th>
                <th scope="col" className="px-4 py-3 text-right">Acciones</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/80">
              {isLoading ? (
                <tr>
                  <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                    <div className="flex items-center justify-center gap-2">
                      <div className="w-4 h-4 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
                      <span>Cargando padrón de contratistas...</span>
                    </div>
                  </td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-10 text-center text-slate-400">
                    <Building2 className="w-8 h-8 text-slate-600 mx-auto mb-2" />
                    <p className="font-medium text-slate-300">No se encontraron empresas contratistas</p>
                    <p className="text-[11px] text-slate-500 mt-0.5">
                      {searchTerm ? 'Pruebe ajustando los términos de búsqueda.' : 'Afile la primera empresa para esta planta.'}
                    </p>
                  </td>
                </tr>
              ) : (
                filtered.map((c) => {
                  const currentStatus = (c.plant_status as PlantContractorStatus) || 'active';
                  const isActionLoading = actionLoadingId === c.id;

                  return (
                    <tr
                      key={c.id}
                      className="hover:bg-slate-800/40 transition-colors group cursor-pointer"
                      onClick={() => onSelectContractor(c)}
                    >
                      {/* Name & Type */}
                      <td className="px-4 py-3.5">
                        <div className="flex items-start gap-2.5">
                          <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mt-0.5">
                            <Building2 className="w-4 h-4" />
                          </div>
                          <div>
                            <span className="font-semibold text-white group-hover:text-emerald-400 transition-colors block">
                              {c.name}
                            </span>
                            <span className="text-[10px] text-slate-400">
                              {c.contractor_type === 'individual'
                                ? 'Persona Física con Actividad Empresarial'
                                : 'Persona Moral (Empresa)'}
                            </span>
                          </div>
                        </div>
                      </td>

                      {/* Fiscal Info */}
                      <td className="px-4 py-3.5 font-mono">
                        <div className="space-y-0.5">
                          <div className="text-slate-200">
                            RFC: <span className="font-bold text-white">{c.rfc || 'N/A'}</span>
                          </div>
                          {c.repse_number && (
                            <div className="text-[11px] text-amber-300/90">
                              REPSE: {c.repse_number}
                            </div>
                          )}
                        </div>
                      </td>

                      {/* Contact */}
                      <td className="px-4 py-3.5">
                        <div className="space-y-0.5">
                          <div className="text-white font-medium">
                            {c.admin_full_name || 'Admin Asignado'}
                          </div>
                          <div className="text-slate-400 text-[11px]">
                            {c.admin_email || 'Sin correo'}
                          </div>
                          {c.admin_phone && (
                            <div className="text-slate-500 text-[10px]">
                              Tel: {c.admin_phone}
                            </div>
                          )}
                        </div>
                      </td>

                      {/* Workers count */}
                      <td className="px-4 py-3.5 text-center">
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-slate-800 text-slate-200 font-semibold text-xs border border-slate-700/80">
                          <Users className="w-3.5 h-3.5 text-cyan-400" />
                          {c.workers_count || 0}
                        </span>
                      </td>

                      {/* Status */}
                      <td className="px-4 py-3.5">
                        {getStatusBadge(c.plant_status)}
                      </td>

                      {/* Actions */}
                      <td
                        className="px-4 py-3.5 text-right space-x-1"
                        onClick={(e) => e.stopPropagation()} // don't trigger row click
                      >
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            size="sm"
                            variant="outline"
                            className="border-slate-700 text-slate-200 hover:text-white hover:bg-slate-800 text-xs h-7 px-2.5 gap-1"
                            onClick={() => onSelectContractor(c)}
                          >
                            <span>Cuadrilla</span>
                            <ChevronRight className="w-3 h-3 text-slate-400" />
                          </Button>

                          {canManageStatus && (
                            <select
                              value={currentStatus}
                              disabled={isActionLoading}
                              onChange={(e) =>
                                handleStatusChange(c.id, e.target.value as PlantContractorStatus)
                              }
                              className="h-7 text-[11px] rounded bg-slate-900 border border-slate-700 text-slate-300 px-2 focus:ring-1 focus:ring-emerald-500 cursor-pointer disabled:opacity-50"
                            >
                              <option value="active">Marcar Activo</option>
                              <option value="pending">Marcar Pendiente</option>
                              <option value="suspended">Suspender</option>
                              <option value="blocked">Bloquear</option>
                            </select>
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
    </div>
  );
}
