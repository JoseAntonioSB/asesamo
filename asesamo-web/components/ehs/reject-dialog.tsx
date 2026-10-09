'use client';

import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { DocumentItem } from '@/lib/api/types';
import { AlertCircle, XCircle, FileText, Check } from 'lucide-react';

interface RejectDialogProps {
  isOpen: boolean;
  onClose: () => void;
  documentItem: DocumentItem | null;
  onConfirmReject: (reason: string) => Promise<void>;
  isSubmitting: boolean;
}

const COMMON_REJECTION_REASONS = [
  'Documento ilegible o baja resolución',
  'Fecha de vigencia vencida',
  'Falta firma o sello oficial (STPS / IMSS)',
  'No corresponde a la identidad del trabajador',
  'Formato no oficial o incompleto',
];

export function RejectDialog({
  isOpen,
  onClose,
  documentItem,
  onConfirmReject,
  isSubmitting,
}: RejectDialogProps) {
  const [reason, setReason] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setReason('');
      setError(null);
    }
  }, [isOpen, documentItem]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = reason.trim();
    if (!trimmed) {
      setError('Debe especificar un motivo de rechazo claro.');
      return;
    }
    setError(null);
    await onConfirmReject(trimmed);
  };

  const handleSelectPreset = (preset: string) => {
    setReason(preset);
    setError(null);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && !isSubmitting && onClose()}>
      <DialogContent className="max-w-lg bg-slate-900 border border-slate-800 text-slate-100 p-6 sm:rounded-xl">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400">
              <XCircle className="w-5 h-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold text-white">
                Rechazar Documento
              </DialogTitle>
              <DialogDescription className="text-xs text-slate-400">
                Especifique la causa técnica o de cumplimiento para notificar al contratista.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {documentItem && (
          <div className="p-3 bg-slate-950 rounded-lg border border-slate-800/80 space-y-1 text-xs">
            <div className="flex justify-between items-center text-slate-300">
              <span className="font-semibold text-white flex items-center gap-1.5">
                <FileText className="w-3.5 h-3.5 text-emerald-400" />
                {documentItem.document_type_name || documentItem.document_type_code || 'Documento'}
              </span>
              <span className="font-mono text-[11px] text-slate-400">
                {documentItem.file_name}
              </span>
            </div>
            <div className="text-slate-400 flex flex-wrap gap-x-4 gap-y-1 pt-1 border-t border-slate-900">
              {documentItem.owner_name && (
                <div>
                  <span className="text-slate-500">Titular: </span>
                  <span className="text-slate-200">{documentItem.owner_name}</span>
                </div>
              )}
              {documentItem.contractor_name && (
                <div>
                  <span className="text-slate-500">Contratista: </span>
                  <span className="text-slate-200">{documentItem.contractor_name}</span>
                </div>
              )}
            </div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-2">
              Motivos frecuentes:
            </label>
            <div className="flex flex-wrap gap-1.5">
              {COMMON_REJECTION_REASONS.map((preset) => {
                const isSelected = reason === preset;
                return (
                  <button
                    key={preset}
                    type="button"
                    onClick={() => handleSelectPreset(preset)}
                    className={`text-[11px] px-2.5 py-1 rounded-md transition-all text-left flex items-center gap-1 border ${
                      isSelected
                        ? 'bg-rose-950/70 border-rose-600 text-rose-200 font-medium'
                        : 'bg-slate-950 hover:bg-slate-800/80 border-slate-800 text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    {isSelected && <Check className="w-3 h-3 text-rose-400" />}
                    {preset}
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <label htmlFor="rejection-reason" className="block text-xs font-semibold text-slate-300 mb-1.5">
              Motivo de rechazo detallado <span className="text-rose-400">*</span>
            </label>
            <textarea
              id="rejection-reason"
              rows={3}
              value={reason}
              onChange={(e) => {
                setReason(e.target.value);
                if (error) setError(null);
              }}
              placeholder="Ej. La fecha de vigencia del DC-3 expiró el mes pasado o el sello no es legible..."
              className="w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-xs text-white placeholder-slate-500 focus:border-rose-500 focus:outline-none focus:ring-1 focus:ring-rose-500 transition-colors resize-none"
              disabled={isSubmitting}
            />
            {error && (
              <p className="mt-1.5 text-xs text-rose-400 flex items-center gap-1">
                <AlertCircle className="w-3.5 h-3.5" />
                {error}
              </p>
            )}
          </div>

          <DialogFooter className="gap-2 sm:gap-0 pt-2 border-t border-slate-800/80">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              disabled={isSubmitting}
              className="border-slate-800 text-slate-300 hover:bg-slate-800 text-xs"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              size="sm"
              disabled={isSubmitting || !reason.trim()}
              className="bg-rose-600 hover:bg-rose-500 text-white font-semibold text-xs shadow-lg shadow-rose-950/50"
            >
              {isSubmitting ? 'Registrando rechazo...' : 'Confirmar Rechazo'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
