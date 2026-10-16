'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { WorkPermitItem } from '@/lib/api/types';
import { PermitStatusBadge } from '@/components/permisos/permit-status-badge';
import { PermitHazardBadge } from '@/components/permisos/permit-hazard-badge';
import { PermitCountdownTimer } from '@/components/permisos/permit-countdown-timer';
import { DynamicStpsChecklist } from '@/components/permisos/dynamic-stps-checklist';
import { EhsAuthorizationModal } from '@/components/permisos/ehs-authorization-modal';
import { FieldControlDialog } from '@/components/permisos/field-control-dialog';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ArrowLeft,
  RefreshCw,
  MapPin,
  Clock,
  ShieldCheck,
  CheckCircle2,
  Users,
  AlertTriangle,
  Play,
  Check,
  FileCheck2,
  UserCheck,
  Building2,
  Lock,
  Ban,
  Calendar,
} from 'lucide-react';

function PermitDetailContent() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const permitId = params?.id as string;

  const [permit, setPermit] = useState<WorkPermitItem | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  // Modals state
  const [ehsModalOpen, setEhsModalOpen] = useState<boolean>(false);
  const [fieldControlAction, setFieldControlAction] = useState<
    'suspend' | 'resume' | 'close' | 'cancel' | null
  >(null);

  const loadPermit = useCallback(
    async (showToast = false) => {
      if (!permitId) return;
      try {
        if (showToast) setIsRefreshing(true);
        else setIsLoading(true);

        const data = await apiClient.getPermit(permitId);
        setPermit(data);

        if (showToast) {
          toast.success('Permiso actualizado');
        }
      } catch (err: any) {
        toast.error('Error al cargar detalle del permiso', {
          description: err.message || 'No se pudo obtener el permiso solicitado.',
        });
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
      }
    },
    [permitId]
  );

  useEffect(() => {
    loadPermit();
  }, [loadPermit]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center text-slate-400">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin" />
          <span className="text-sm font-medium">Cargando expediente del permiso PTAR...</span>
        </div>
      </div>
    );
  }

  if (!permit) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center text-slate-400 p-4">
        <div className="text-center space-y-3">
          <AlertTriangle className="w-10 h-10 text-amber-500 mx-auto" />
          <h2 className="text-lg font-bold text-white">Permiso No Encontrado</h2>
          <p className="text-xs text-slate-500">
            El folio solicitado no existe o no tiene privilegios para consultarlo.
          </p>
          <Link href="/permisos">
            <Button variant="outline" size="sm" className="border-slate-800 text-slate-300">
              Volver a Lista de Permisos
            </Button>
          </Link>
        </div>
      </div>
    );
  }

  const isEhsOrAdmin =
    user?.role === 'ehs_manager' ||
    user?.role === 'plant_admin' ||
    user?.role === 'tenant_admin' ||
    user?.role === 'super_admin';

  const workers = permit.workers || [];
  const checklist = permit.checklist || permit.checklists || [];
  const signatures = permit.signatures || [];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-cyan-500 selection:text-white">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link
              href="/permisos"
              className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold text-white uppercase">
                  {permit.permit_number}
                </span>
                <PermitStatusBadge status={permit.status} size="sm" />
              </div>
              <p className="text-xs text-slate-400">Expediente de Alto Riesgo STPS</p>
            </div>
          </div>

          <div className="flex items-center gap-2.5">
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadPermit(true)}
              disabled={isRefreshing}
              className="border-slate-800 text-slate-300 hover:bg-slate-800 gap-1.5"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span className="hidden sm:inline">Actualizar</span>
            </Button>
            <ThemeToggle />
          </div>
        </div>
      </header>

      {/* Main Body */}
      <main className="flex-1 max-w-6xl mx-auto px-4 sm:px-6 py-8 w-full space-y-6">
        {/* Banner with Status & Operational Actions */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-5 shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex flex-wrap items-center gap-2.5">
              <PermitHazardBadge hazardType={permit.permit_type} size="md" />
              <PermitCountdownTimer
                startTime={permit.start_time}
                endTime={permit.end_time}
                status={permit.status}
              />
            </div>
            <div className="flex items-center gap-2 text-xs text-slate-400 pt-0.5">
              <Building2 className="w-3.5 h-3.5 text-slate-500" />
              <span className="font-medium text-slate-200">
                {permit.contractor_name || 'Contratista'}
              </span>
              <span>&bull;</span>
              <MapPin className="w-3.5 h-3.5 text-slate-500" />
              <span>{permit.work_area}</span>
            </div>
          </div>

          {/* Operational Action Buttons */}
          <div className="flex flex-wrap items-center gap-2 self-stretch md:self-auto">
            {/* EHS Authorization Button */}
            {isEhsOrAdmin &&
              (permit.status === 'pending_approval' || permit.status === 'draft') && (
                <Button
                  onClick={() => setEhsModalOpen(true)}
                  className="bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs gap-1.5 shadow-lg shadow-emerald-950/50"
                >
                  <ShieldCheck className="w-4 h-4" />
                  <span>Revisar y Autorizar EHS</span>
                </Button>
              )}

            {/* Suspend Button (when active) */}
            {permit.status === 'approved_active' && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setFieldControlAction('suspend')}
                className="border-rose-800 text-rose-300 hover:bg-rose-950/50 gap-1.5 text-xs"
              >
                <AlertTriangle className="w-3.5 h-3.5 text-rose-400" />
                <span>Suspender</span>
              </Button>
            )}

            {/* Resume Button (when suspended) */}
            {permit.status === 'suspended' && (
              <Button
                size="sm"
                onClick={() => setFieldControlAction('resume')}
                className="bg-emerald-600 hover:bg-emerald-500 text-white gap-1.5 text-xs"
              >
                <Play className="w-3.5 h-3.5" />
                <span>Reactivar en Campo</span>
              </Button>
            )}

            {/* Close Button (when active or suspended) */}
            {(permit.status === 'approved_active' || permit.status === 'suspended') && (
              <Button
                size="sm"
                onClick={() => setFieldControlAction('close')}
                className="bg-blue-600 hover:bg-blue-500 text-white gap-1.5 text-xs"
              >
                <CheckCircle2 className="w-3.5 h-3.5" />
                <span>Concluir y Cerrar</span>
              </Button>
            )}

            {/* Cancel Button (when draft or pending) */}
            {(permit.status === 'draft' || permit.status === 'pending_approval') && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setFieldControlAction('cancel')}
                className="border-slate-800 text-slate-400 hover:text-rose-400 hover:bg-slate-800 gap-1.5 text-xs"
              >
                <Ban className="w-3.5 h-3.5" />
                <span>Cancelar</span>
              </Button>
            )}
          </div>
        </div>

        {/* Detailed Work Description & Metadata */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="md:col-span-2 space-y-6">
            {/* Description Card */}
            <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 space-y-4">
              <h3 className="text-sm font-bold text-white flex items-center gap-2">
                <span>Descripción y Alcance de las Labores</span>
              </h3>
              <p className="text-xs text-slate-300 leading-relaxed whitespace-pre-line bg-slate-950/60 p-3.5 rounded-lg border border-slate-800/80">
                {permit.description}
              </p>

              {permit.hazard_description && (
                <div className="space-y-1.5">
                  <h4 className="text-xs font-semibold text-amber-400 flex items-center gap-1.5">
                    <AlertTriangle className="w-3.5 h-3.5" />
                    <span>Análisis de Riesgos Específicos</span>
                  </h4>
                  <p className="text-xs text-slate-400 bg-amber-950/10 p-3 rounded-lg border border-amber-800/30">
                    {permit.hazard_description}
                  </p>
                </div>
              )}
            </div>

            {/* Cuadrilla / Workers Table Card */}
            <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold text-white flex items-center gap-2">
                  <Users className="w-4 h-4 text-cyan-400" />
                  <span>Cuadrilla Asignada ({workers.length} operarios)</span>
                </h3>
                <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-emerald-950 text-emerald-300 border border-emerald-700">
                  Candado STPS Acreditado
                </span>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-800 text-[11px] font-semibold text-slate-400 uppercase">
                      <th className="py-2.5 px-3">Trabajador</th>
                      <th className="py-2.5 px-3">Puesto / Registro</th>
                      <th className="py-2.5 px-3">DC-3 Especialidad</th>
                      <th className="py-2.5 px-3">Aptitud Médica</th>
                      <th className="py-2.5 px-3 text-right">Estatus Candado</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {workers.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="py-4 text-center text-slate-500">
                          Sin operarios asignados a este permiso.
                        </td>
                      </tr>
                    ) : (
                      workers.map((worker) => {
                        const snap = worker.compliance_snapshot;
                        const isCompliant =
                          worker.compliance_status === 'compliant' ||
                          worker.is_compliant ||
                          snap?.is_compliant;

                        return (
                          <tr key={worker.id || worker.worker_id} className="hover:bg-slate-800/30">
                            <td className="py-3 px-3 font-semibold text-white">
                              {worker.full_name || worker.worker_name || 'Trabajador'}
                            </td>
                            <td className="py-3 px-3 text-slate-400">
                              <div>{worker.job_title || 'Operario'}</div>
                              <div className="text-[10px] text-slate-500 font-mono">
                                {worker.curp || worker.imss_number || ''}
                              </div>
                            </td>
                            <td className="py-3 px-3">
                              <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-medium">
                                <Check className="w-3 h-3 text-emerald-400" />
                                <span>{snap?.dc3_code || 'DC-3 Vigente'}</span>
                              </span>
                            </td>
                            <td className="py-3 px-3">
                              <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-medium">
                                <Check className="w-3 h-3 text-emerald-400" />
                                <span>Apto Médico</span>
                              </span>
                            </td>
                            <td className="py-3 px-3 text-right">
                              {isCompliant ? (
                                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-950 text-emerald-300 border border-emerald-700">
                                  ✓ Autorizado
                                </span>
                              ) : (
                                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold bg-rose-950 text-rose-300 border border-rose-700">
                                  ✕ No Cumple
                                </span>
                              )}
                            </td>
                          </tr>
                        );
                      })
                    )}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Checklist Audit Section */}
            <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 space-y-4">
              <DynamicStpsChecklist
                permitType={permit.permit_type}
                items={checklist}
                readOnly={permit.status === 'closed' || permit.status === 'cancelled'}
                onChangeItems={async (updatedItems) => {
                  if (permit.status === 'draft' || permit.status === 'pending_approval') {
                    try {
                      await apiClient.updatePermitChecklist(permit.id, {
                        items: updatedItems.map((it) => ({
                          item_key: it.item_key,
                          is_verified: !!(it.is_verified || it.is_checked),
                          notes: it.notes || undefined,
                          measurements: it.measurements || undefined,
                        })),
                      });
                    } catch {
                      // Silent auto-save or handle error
                    }
                  }
                }}
              />
            </div>
          </div>

          {/* Right Column: Metadata, Time Window & Signatures Trail */}
          <div className="space-y-6">
            {/* Validity Window Card */}
            <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 space-y-3">
              <h4 className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-1.5">
                <Clock className="w-4 h-4 text-cyan-400" />
                <span>Ventana Operativa</span>
              </h4>

              <div className="space-y-2 text-xs">
                <div>
                  <span className="text-slate-500 block text-[10px]">Inicio Programado:</span>
                  <span className="font-mono text-slate-200">
                    {new Intl.DateTimeFormat('es-MX', {
                      dateStyle: 'medium',
                      timeStyle: 'short',
                    }).format(new Date(permit.start_time))}
                  </span>
                </div>
                <div>
                  <span className="text-slate-500 block text-[10px]">Término Programado:</span>
                  <span className="font-mono text-slate-200">
                    {new Intl.DateTimeFormat('es-MX', {
                      dateStyle: 'medium',
                      timeStyle: 'short',
                    }).format(new Date(permit.end_time))}
                  </span>
                </div>
              </div>
            </div>

            {/* Digital Signatures Audit Trail Card */}
            <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-1.5">
                  <Lock className="w-4 h-4 text-emerald-400" />
                  <span>Bitácora de Firmas Digitales</span>
                </h4>
              </div>

              <div className="space-y-3">
                {/* 1. Requester / Contractor Supervisor */}
                <div className="p-3 bg-slate-950/70 border border-slate-800 rounded-lg text-xs space-y-1">
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-slate-300">Solicitante Contratista</span>
                    <span className="text-[10px] text-emerald-400 font-mono">Firma Validada</span>
                  </div>
                  <div className="text-slate-400 font-mono text-[11px]">
                    {permit.created_by_name || 'Supervisor de Contratista'}
                  </div>
                  <div className="text-[10px] text-slate-500">
                    {new Intl.DateTimeFormat('es-MX', {
                      dateStyle: 'short',
                      timeStyle: 'short',
                    }).format(new Date(permit.created_at))}
                  </div>
                </div>

                {/* 2. EHS Authorizer */}
                <div
                  className={`p-3 rounded-lg text-xs space-y-1 border ${
                    permit.approved_at
                      ? 'bg-slate-950/70 border-emerald-800/60'
                      : 'bg-slate-950/40 border-slate-800/60 opacity-60'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-slate-300">Autorizador EHS / Planta</span>
                    {permit.approved_at ? (
                      <span className="text-[10px] text-emerald-400 font-mono">Sello Digital OK</span>
                    ) : (
                      <span className="text-[10px] text-amber-400 font-mono">Pendiente</span>
                    )}
                  </div>
                  <div className="text-slate-400 font-mono text-[11px]">
                    {permit.approved_by_name ||
                      (permit.approved_at ? 'Ing. Seguridad Industrial' : 'En espera de firma EHS')}
                  </div>
                  {permit.approved_at && (
                    <div className="text-[10px] text-slate-500">
                      {new Intl.DateTimeFormat('es-MX', {
                        dateStyle: 'short',
                        timeStyle: 'short',
                      }).format(new Date(permit.approved_at))}
                    </div>
                  )}
                </div>

                {/* Status reason if suspended or closed */}
                {(permit.suspension_reason || permit.status_reason) && (
                  <div className="p-3 bg-rose-950/30 border border-rose-800/50 rounded-lg text-xs space-y-1">
                    <span className="text-rose-400 font-semibold block text-[11px]">
                      Observación de Estatus:
                    </span>
                    <p className="text-slate-300 text-[11px]">
                      {permit.suspension_reason || permit.status_reason}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* EHS Authorization Modal */}
      <EhsAuthorizationModal
        isOpen={ehsModalOpen}
        onClose={() => setEhsModalOpen(false)}
        permit={permit}
        onSuccess={(updated) => setPermit(updated)}
      />

      {/* Field Control Dialog */}
      {fieldControlAction && (
        <FieldControlDialog
          isOpen={true}
          onClose={() => setFieldControlAction(null)}
          permit={permit}
          actionType={fieldControlAction}
          onSuccess={(updated) => setPermit(updated)}
        />
      )}
    </div>
  );
}

export default function PermitDetailPage() {
  return (
    <ProtectedRoute>
      <PermitDetailContent />
    </ProtectedRoute>
  );
}
