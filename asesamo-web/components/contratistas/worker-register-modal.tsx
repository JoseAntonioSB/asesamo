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
import { Input } from '@/components/ui/input';
import { apiClient, ApiError } from '@/lib/api/client';
import { ContractorWorkerItem } from '@/lib/api/types';
import { toast } from 'sonner';
import { UserCheck, ShieldAlert, Loader2, HardHat, IdCard } from 'lucide-react';

interface WorkerRegisterModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  contractorId?: string;
  contractorName?: string;
  onWorkerAdded: (newWorker: ContractorWorkerItem) => void;
}

export function WorkerRegisterModal({
  open,
  onOpenChange,
  contractorId,
  contractorName,
  onWorkerAdded,
}: WorkerRegisterModalProps) {
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [jobTitle, setJobTitle] = useState('');
  const [curp, setCurp] = useState('');
  const [imssNumber, setImssNumber] = useState('');

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const resetForm = () => {
    setFullName('');
    setEmail('');
    setPhone('');
    setJobTitle('');
    setCurp('');
    setImssNumber('');
    setErrorMsg(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);

    const trimmedName = fullName.trim();
    const trimmedEmail = email.trim().toLowerCase();
    const trimmedCurp = curp.trim().toUpperCase();
    const trimmedImss = imssNumber.trim();

    if (!trimmedName) {
      setErrorMsg('El nombre completo del trabajador es obligatorio.');
      return;
    }
    if (!trimmedEmail || !trimmedEmail.includes('@')) {
      setErrorMsg('Proporcione un correo electrónico válido para la cuenta del trabajador.');
      return;
    }
    if (trimmedCurp && trimmedCurp.length !== 18) {
      setErrorMsg('La CURP debe contar con exactamente 18 caracteres.');
      return;
    }
    if (trimmedImss && trimmedImss.length !== 11) {
      setErrorMsg('El Número de Seguridad Social (NSS / IMSS) consta de 11 dígitos.');
      return;
    }

    try {
      setIsSubmitting(true);
      const worker = await apiClient.addContractorWorker({
        contractor_id: contractorId,
        full_name: trimmedName,
        email: trimmedEmail,
        phone: phone.trim() || undefined,
        job_title: jobTitle.trim() || undefined,
        curp: trimmedCurp || undefined,
        imss_number: trimmedImss || undefined,
      });

      toast.success('Trabajador registrado exitosamente en la cuadrilla', {
        description: `${worker.full_name} fue dado de alta con credencial digital provisional.`,
      });

      onWorkerAdded(worker);
      resetForm();
      onOpenChange(false);
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setErrorMsg(err.message);
        toast.error('Error al registrar trabajador', { description: err.message });
      } else {
        const fallback = 'No se pudo dar de alta al trabajador. Verifique los datos ingresados.';
        setErrorMsg(fallback);
        toast.error('Error de comunicación', { description: fallback });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(val) => { if (!isSubmitting) { onOpenChange(val); resetForm(); } }}>
      <DialogContent className="sm:max-w-lg bg-slate-900 border-slate-800 text-slate-100 p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader className="space-y-1">
          <div className="flex items-center gap-2.5 text-cyan-400">
            <HardHat className="w-5 h-5" />
            <DialogTitle className="text-lg font-bold text-white tracking-tight">
              Alta de Personal en Cuadrilla
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-slate-400">
            {contractorName
              ? `Empresa asignada: ${contractorName}`
              : 'Ingrese los datos laborales y regulatorios del personal operativo.'}
          </DialogDescription>
        </DialogHeader>

        {errorMsg && (
          <div className="flex items-start gap-2.5 p-3 rounded-lg bg-red-950/60 border border-red-500/40 text-red-200 text-xs">
            <ShieldAlert className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
            <span>{errorMsg}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4 mt-2">
          {/* Identidad */}
          <div className="space-y-3 p-3.5 rounded-lg bg-slate-950/60 border border-slate-800">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <IdCard className="w-3.5 h-3.5 text-cyan-400" />
              <span>Identidad y Puesto Operativo</span>
            </h4>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">
                Nombre Completo (como en INE/CURP) <span className="text-cyan-400">*</span>
              </label>
              <Input
                type="text"
                placeholder="Ej. Juan Alberto Pérez Ramírez"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                disabled={isSubmitting}
                className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                required
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  Puesto / Especialidad
                </label>
                <Input
                  type="text"
                  placeholder="Ej. Soldador Argonero 6G"
                  value={jobTitle}
                  onChange={(e) => setJobTitle(e.target.value)}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  Teléfono Móvil
                </label>
                <Input
                  type="tel"
                  placeholder="4428889900"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">
                Correo Electrónico del Trabajador <span className="text-cyan-400">*</span>
              </label>
              <Input
                type="email"
                placeholder="juan.perez@cuadrilla.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isSubmitting}
                className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                required
              />
              <p className="text-[10px] text-slate-500">
                Permite al trabajador ingresar a ver su Credencial QR digital en su móvil.
              </p>
            </div>
          </div>

          {/* Cumplimiento Regulatorio Mexicano */}
          <div className="space-y-3 p-3.5 rounded-lg bg-slate-950/60 border border-slate-800">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <span>Cumplimiento Normativo (CURP / IMSS)</span>
            </h4>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  CURP (18 caracteres)
                </label>
                <Input
                  type="text"
                  placeholder="PERJ900101HNLNN01"
                  value={curp}
                  maxLength={18}
                  onChange={(e) => setCurp(e.target.value.toUpperCase())}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 uppercase font-mono text-xs h-9"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  NSS / IMSS (11 dígitos)
                </label>
                <Input
                  type="text"
                  placeholder="12345678901"
                  value={imssNumber}
                  maxLength={11}
                  onChange={(e) => setImssNumber(e.target.value.replace(/\D/g, ''))}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 font-mono text-xs h-9"
                />
              </div>
            </div>
          </div>

          <DialogFooter className="pt-2 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              disabled={isSubmitting}
              onClick={() => onOpenChange(false)}
              className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="bg-cyan-600 hover:bg-cyan-500 text-white text-xs gap-1.5"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  Guardando...
                </>
              ) : (
                <>
                  <UserCheck className="w-3.5 h-3.5" />
                  Registrar en Cuadrilla
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
