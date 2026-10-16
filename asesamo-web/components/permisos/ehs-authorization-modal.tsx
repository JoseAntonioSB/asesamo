'use client';

import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { WorkPermitItem } from '@/lib/api/types';
import { apiClient } from '@/lib/api/client';
import { PermitHazardBadge } from './permit-hazard-badge';
import { ShieldCheck, CheckCircle2, AlertTriangle, Users, FileCheck, Wind, Lock } from 'lucide-react';
import { toast } from 'sonner';

interface EhsAuthorizationModalProps {
  isOpen: boolean;
  onClose: () => void;
  permit: WorkPermitItem | null;
  onSuccess: (updatedPermit: WorkPermitItem) => void;
}

export function EhsAuthorizationModal({
  isOpen,
  onClose,
  permit,
  onSuccess,
}: EhsAuthorizationModalProps) {
  const [notes, setNotes] = useState('');
  const [signatureData, setSignatureData] = useState('EHS_DIGITAL_SIG_VALIDATED');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [confirmedCheck, setConfirmedCheck] = useState(false);

  if (!permit) return null;

  const workers = permit.workers || [];
  const checklist = permit.checklist || permit.checklists || [];
  const checkedCount = checklist.filter((i) => i.is_verified || i.is_checked).length;
  const isChecklistComplete = checklist.length > 0 && checkedCount === checklist.length;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!confirmedCheck) {
      toast.error('Debe confirmar la verificación de condiciones de seguridad antes de autorizar.');
      return;
    }

    try {
      setIsSubmitting(true);
      const updated = await apiClient.approvePermit(permit.id, notes, signatureData);
      toast.success(`Permiso ${permit.permit_number} autorizado con éxito. Estatus: Activo en Campo.`);
      onSuccess(updated);
      onClose();
    } catch (err: any) {
      toast.error(err.message || 'Error al autorizar el permiso de trabajo.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl bg-slate-900 border-slate-800 text-slate-100 p-6 shadow-2xl">
        <DialogHeader className="space-y-2">
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-emerald-950 border border-emerald-700 text-emerald-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold text-white">
                Autorización Digital EHS / Seguridad Industrial
              </DialogTitle>
              <DialogDescription className="text-xs text-slate-400">
                Firma electrónica y validación regulatoria para emisión de PTAR en campo.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 my-2">
          {/* Permit Summary Card */}
          <div className="p-3.5 bg-slate-950/80 border border-slate-800 rounded-xl space-y-2">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="font-mono text-sm font-bold text-white">{permit.permit_number}</span>
              <PermitHazardBadge hazardType={permit.permit_type} size="sm" />
            </div>

            <div className="grid grid-cols-2 gap-2 text-xs text-slate-400 pt-1 border-t border-slate-800/80">
              <div>
                <span className="text-slate-500 block text-[10px] uppercase font-semibold">Empresa</span>
                <span className="text-slate-200 font-medium">{permit.contractor_name || 'Contratista'}</span>
              </div>
              <div>
                <span className="text-slate-500 block text-[10px] uppercase font-semibold">Área / Ubicación</span>
                <span className="text-slate-200 font-medium">{permit.work_area}</span>
              </div>
            </div>

            <div className="text-xs text-slate-400">
              <span className="text-slate-500 block text-[10px] uppercase font-semibold">Descripción</span>
              <p className="text-slate-300 line-clamp-2">{permit.description}</p>
            </div>
          </div>

          {/* Validation Checklist Audit Summary */}
          <div className="grid grid-cols-2 gap-3 text-xs">
            <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg flex items-center gap-2.5">
              <Users className="w-4 h-4 text-cyan-400 shrink-0" />
              <div>
                <span className="text-slate-400 block text-[10px]">Cuadrilla Validada</span>
                <span className="text-emerald-400 font-semibold font-mono">
                  {workers.length} operarios con Candado STPS
                </span>
              </div>
            </div>

            <div className="p-3 bg-slate-950/60 border border-slate-800 rounded-lg flex items-center gap-2.5">
              <FileCheck className="w-4 h-4 text-emerald-400 shrink-0" />
              <div>
                <span className="text-slate-400 block text-[10px]">Verificación Preventiva</span>
                <span className={`font-semibold font-mono ${isChecklistComplete ? 'text-emerald-400' : 'text-amber-400'}`}>
                  {checkedCount} de {checklist.length} puntos verificados
                </span>
              </div>
            </div>
          </div>

          {/* Atmospheric Warning if Confined Space */}
          {permit.permit_type === 'confined_space' && (
            <div className="p-3 bg-amber-950/30 border border-amber-800/60 rounded-xl flex items-start gap-2 text-xs text-amber-300">
              <Wind className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
              <div>
                <strong className="block font-semibold">Prueba de Atmósfera NOM-033:</strong>
                <span>
                  Confirmar que el vigía cuenta con equipo detector multigás calibrado y comunicación ininterrumpida.
                </span>
              </div>
            </div>
          )}

          {/* Authorizer Notes Input */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-300 block">
              Observaciones del Supervisor EHS (Opcional)
            </label>
            <input
              type="text"
              placeholder="Ej. Se autoriza con vigía permanente y rescate posicionado..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500"
            />
          </div>

          {/* Mandatory Safety Confirmation Checkbox */}
          <div className="p-3 bg-emerald-950/20 border border-emerald-800/40 rounded-xl flex items-start gap-3">
            <input
              type="checkbox"
              id="confirm-ehs"
              checked={confirmedCheck}
              onChange={(e) => setConfirmedCheck(e.target.checked)}
              className="w-4 h-4 mt-0.5 rounded border-slate-700 text-emerald-500 focus:ring-emerald-500/20 bg-slate-900 cursor-pointer"
            />
            <label htmlFor="confirm-ehs" className="text-xs text-slate-300 cursor-pointer leading-relaxed">
              Doy fe como responsable de Seguridad y Salud en el Trabajo (EHS) que el personal cuenta con aptitud médica y capacitación DC-3 vigente, y que el área de trabajo cumple con todas las medidas preventivas de la normativa STPS aplicable.
            </label>
          </div>

          <DialogFooter className="gap-2 sm:gap-0 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSubmitting}
              className="border-slate-700 text-slate-300 hover:bg-slate-800"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting || !confirmedCheck}
              className="bg-emerald-600 hover:bg-emerald-500 text-white font-medium gap-1.5"
            >
              {isSubmitting ? (
                <>
                  <div className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Emitiendo Autorización...</span>
                </>
              ) : (
                <>
                  <Lock className="w-3.5 h-3.5" />
                  <span>Autorizar y Activar en Campo</span>
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
