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
import { WorkPermitItem, TransitionPermitStatusPayload } from '@/lib/api/types';
import { apiClient } from '@/lib/api/client';
import { AlertTriangle, CheckCircle2, Play, Ban, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';

interface FieldControlDialogProps {
  isOpen: boolean;
  onClose: () => void;
  permit: WorkPermitItem | null;
  actionType: 'suspend' | 'resume' | 'close' | 'cancel';
  onSuccess: (updatedPermit: WorkPermitItem) => void;
}

const COMMON_SUSPENSION_REASONS = [
  'Condiciones climáticas adversas (vientos > 40 km/h o lluvia)',
  'Alarma de gas o detección de atmósfera peligrosa',
  'Falla o deterioro detectado en equipo de protección / andamio',
  'Incidente o emergencia general en planta',
  'Interrupción por incumplimiento de medidas de seguridad',
];

export function FieldControlDialog({
  isOpen,
  onClose,
  permit,
  actionType,
  onSuccess,
}: FieldControlDialogProps) {
  const [reason, setReason] = useState('');
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (!permit) return null;

  const config = {
    suspend: {
      title: 'Suspender Permiso de Trabajo',
      desc: 'Pausa inmediata de las labores en campo por riesgo o contingencia.',
      icon: AlertTriangle,
      iconColor: 'text-rose-400',
      iconBg: 'bg-rose-950 border-rose-800',
      btnText: 'Confirmar Suspensión',
      btnColor: 'bg-rose-600 hover:bg-rose-500 text-white',
    },
    resume: {
      title: 'Reactivar Permiso en Campo',
      desc: 'Reanudación de actividades tras validar que la condición de riesgo cesó.',
      icon: Play,
      iconColor: 'text-emerald-400',
      iconBg: 'bg-emerald-950 border-emerald-800',
      btnText: 'Reactivar Permiso',
      btnColor: 'bg-emerald-600 hover:bg-emerald-500 text-white',
    },
    close: {
      title: 'Cierre Formal de Permiso (Conclusión)',
      desc: 'Fin de actividades, entrega de área limpia y desmovilización de personal.',
      icon: CheckCircle2,
      iconColor: 'text-blue-400',
      iconBg: 'bg-blue-950 border-blue-800',
      btnText: 'Cerrar Permiso',
      btnColor: 'bg-blue-600 hover:bg-blue-500 text-white',
    },
    cancel: {
      title: 'Cancelar Permiso de Trabajo',
      desc: 'Cancelación definitiva del permiso antes o durante la ejecución.',
      icon: Ban,
      iconColor: 'text-zinc-400',
      iconBg: 'bg-zinc-900 border-zinc-700',
      btnText: 'Cancelar Permiso',
      btnColor: 'bg-zinc-700 hover:bg-zinc-600 text-white',
    },
  }[actionType];

  const IconComponent = config.icon;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (actionType === 'suspend' && !reason.trim()) {
      toast.error('Debe especificar el motivo obligatorio de suspensión.');
      return;
    }

    try {
      setIsSubmitting(true);
      const payload: TransitionPermitStatusPayload = {
        action: actionType,
        reason: reason.trim() || undefined,
        notes: notes.trim() || undefined,
      };

      const updated = await apiClient.transitionPermitStatus(permit.id, payload);
      toast.success(`Estatus del permiso ${permit.permit_number} actualizado.`);
      onSuccess(updated);
      onClose();
    } catch (err: any) {
      toast.error(err.message || 'Error al actualizar el estatus del permiso.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg bg-slate-900 border-slate-800 text-slate-100 p-6 shadow-2xl">
        <DialogHeader className="space-y-2">
          <div className="flex items-center gap-2.5">
            <div className={`p-2 rounded-lg border ${config.iconBg} ${config.iconColor}`}>
              <IconComponent className="w-5 h-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold text-white">{config.title}</DialogTitle>
              <DialogDescription className="text-xs text-slate-400">{config.desc}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 my-2">
          {/* Permit quick identifier */}
          <div className="p-3 bg-slate-950 border border-slate-800 rounded-lg text-xs space-y-1">
            <div className="flex items-center justify-between font-mono">
              <span className="font-bold text-white">{permit.permit_number}</span>
              <span className="text-slate-400">{permit.work_area}</span>
            </div>
            <p className="text-slate-400 truncate">{permit.description}</p>
          </div>

          {/* Suspend Reason Selection */}
          {actionType === 'suspend' && (
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200 block">
                Motivo de Suspensión (Obligatorio)
              </label>
              <div className="space-y-1.5">
                {COMMON_SUSPENSION_REASONS.map((preset) => (
                  <button
                    key={preset}
                    type="button"
                    onClick={() => setReason(preset)}
                    className={`w-full text-left text-xs p-2 rounded border transition-colors ${
                      reason === preset
                        ? 'bg-rose-950/60 border-rose-600 text-rose-200 font-medium'
                        : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700'
                    }`}
                  >
                    {preset}
                  </button>
                ))}
              </div>
              <input
                type="text"
                placeholder="O escriba otro motivo específico..."
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-rose-500 mt-1"
              />
            </div>
          )}

          {/* Close Confirmation Details */}
          {actionType === 'close' && (
            <div className="p-3 bg-blue-950/30 border border-blue-800/40 rounded-xl space-y-2 text-xs text-blue-300">
              <div className="flex items-center gap-1.5 font-semibold">
                <CheckCircle2 className="w-4 h-4 text-blue-400" />
                <span>Verificación de Conclusión de Labores</span>
              </div>
              <p className="text-slate-300 text-[11px] leading-relaxed">
                Al cerrar este permiso confirma que las herramientas y equipos fueron retirados, el área fue aseada, se restablecieron las protecciones colectivas y el personal evacuó el punto de trabajo.
              </p>
            </div>
          )}

          {/* Additional Notes */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-slate-300 block">
              Comentarios u observaciones adicionales
            </label>
            <textarea
              rows={3}
              placeholder="Detalles sobre el estado del área, personal o seguimiento..."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2.5 text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500 resize-none"
            />
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
              disabled={isSubmitting}
              className={`${config.btnColor} font-medium`}
            >
              {isSubmitting ? 'Procesando...' : config.btnText}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
