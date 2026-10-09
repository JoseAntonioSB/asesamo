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
import { ContractorItem } from '@/lib/api/types';
import { toast } from 'sonner';
import { Building2, UserPlus, ShieldAlert, Loader2, CheckCircle2 } from 'lucide-react';

interface ContractorInviteModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  plantId: string;
  onInviteSuccess: (newContractor: ContractorItem) => void;
}

export function ContractorInviteModal({
  open,
  onOpenChange,
  plantId,
  onInviteSuccess,
}: ContractorInviteModalProps) {
  const [name, setName] = useState('');
  const [rfc, setRfc] = useState('');
  const [repseNumber, setRepseNumber] = useState('');
  const [contractorType, setContractorType] = useState<'company' | 'individual'>('company');
  const [adminEmail, setAdminEmail] = useState('');
  const [adminFullName, setAdminFullName] = useState('');
  const [adminPhone, setAdminPhone] = useState('');
  const [initialStatus, setInitialStatus] = useState<'active' | 'pending'>('active');

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const resetForm = () => {
    setName('');
    setRfc('');
    setRepseNumber('');
    setContractorType('company');
    setAdminEmail('');
    setAdminFullName('');
    setAdminPhone('');
    setInitialStatus('active');
    setErrorMsg(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg(null);

    const trimmedName = name.trim();
    const trimmedRfc = rfc.trim().toUpperCase();
    const trimmedEmail = adminEmail.trim().toLowerCase();

    if (!trimmedName) {
      setErrorMsg('La razón social de la empresa es obligatoria.');
      return;
    }
    if (!trimmedRfc || trimmedRfc.length < 12 || trimmedRfc.length > 13) {
      setErrorMsg('El RFC debe tener entre 12 (Moral) y 13 (Física) caracteres alfanuméricos.');
      return;
    }
    if (!trimmedEmail || !trimmedEmail.includes('@')) {
      setErrorMsg('Proporcione un correo electrónico válido para el Administrador.');
      return;
    }

    try {
      setIsSubmitting(true);
      const newContractor = await apiClient.inviteContractor({
        plant_id: plantId,
        name: trimmedName,
        rfc: trimmedRfc,
        repse_number: repseNumber.trim() || undefined,
        contractor_type: contractorType,
        admin_email: trimmedEmail,
        admin_full_name: adminFullName.trim() || undefined,
        admin_phone: adminPhone.trim() || undefined,
        initial_status: initialStatus,
      });

      toast.success('Empresa contratista registrada exitosamente', {
        description: `Se dio de alta a ${newContractor.name} y se generó el usuario administrador (${trimmedEmail}).`,
      });

      onInviteSuccess(newContractor);
      resetForm();
      onOpenChange(false);
    } catch (err: unknown) {
      if (err instanceof ApiError) {
        setErrorMsg(err.message);
        toast.error('Error al afiliar contratista', { description: err.message });
      } else {
        const fallback = 'No se pudo registrar al contratista. Verifique los datos o conexión.';
        setErrorMsg(fallback);
        toast.error('Error de comunicación', { description: fallback });
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(val) => { if (!isSubmitting) { onOpenChange(val); resetForm(); } }}>
      <DialogContent className="sm:max-w-xl bg-slate-900 border-slate-800 text-slate-100 p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader className="space-y-1">
          <div className="flex items-center gap-2.5 text-emerald-400">
            <Building2 className="w-5 h-5" />
            <DialogTitle className="text-lg font-bold text-white tracking-tight">
              Afiliar Nueva Empresa Contratista
            </DialogTitle>
          </div>
          <DialogDescription className="text-xs text-slate-400">
            Registre la razón social conforme al SAT y STPS (REPSE). Se generará automáticamente el usuario administrador para gestionar la cuadrilla y subir expedientes laborales.
          </DialogDescription>
        </DialogHeader>

        {errorMsg && (
          <div className="flex items-start gap-2.5 p-3 rounded-lg bg-red-950/60 border border-red-500/40 text-red-200 text-xs">
            <ShieldAlert className="w-4 h-4 text-red-400 flex-shrink-0 mt-0.5" />
            <span>{errorMsg}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4 mt-2">
          {/* Datos Fiscales */}
          <div className="space-y-3 p-3.5 rounded-lg bg-slate-950/60 border border-slate-800">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <span>1. Datos Fiscales y STPS</span>
            </h4>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">
                Razón Social / Denominación Comercial <span className="text-emerald-400">*</span>
              </label>
              <Input
                type="text"
                placeholder="Ej. Servicios Industriales y Montajes del Bajío SA de CV"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isSubmitting}
                className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                required
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  RFC (SAT) <span className="text-emerald-400">*</span>
                </label>
                <Input
                  type="text"
                  placeholder="SIM180214XYZ"
                  value={rfc}
                  maxLength={13}
                  onChange={(e) => setRfc(e.target.value.toUpperCase())}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 uppercase font-mono text-xs h-9"
                  required
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  Folio REPSE (STPS)
                </label>
                <Input
                  type="text"
                  placeholder="REPSE-2024-XXXX"
                  value={repseNumber}
                  onChange={(e) => setRepseNumber(e.target.value)}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 font-mono text-xs h-9"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Tipo de Persona</label>
                <select
                  value={contractorType}
                  onChange={(e) => setContractorType(e.target.value as 'company' | 'individual')}
                  disabled={isSubmitting}
                  className="w-full h-9 rounded-md bg-slate-900 border border-slate-700 text-slate-100 text-xs px-3 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                >
                  <option value="company">Persona Moral (Empresa)</option>
                  <option value="individual">Persona Física con Actividad Empresarial</option>
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">Estatus Inicial en Planta</label>
                <select
                  value={initialStatus}
                  onChange={(e) => setInitialStatus(e.target.value as 'active' | 'pending')}
                  disabled={isSubmitting}
                  className="w-full h-9 rounded-md bg-slate-900 border border-slate-700 text-slate-100 text-xs px-3 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                >
                  <option value="active">Activo (Autorizado para ingresar)</option>
                  <option value="pending">Pendiente (Requiere revisión EHS previa)</option>
                </select>
              </div>
            </div>
          </div>

          {/* Administrador Contratista */}
          <div className="space-y-3 p-3.5 rounded-lg bg-slate-950/60 border border-slate-800">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
              <span>2. Administrador / Contacto de la Empresa</span>
            </h4>

            <div className="space-y-1.5">
              <label className="text-xs font-medium text-slate-300">
                Correo Electrónico de Acceso <span className="text-emerald-400">*</span>
              </label>
              <Input
                type="email"
                placeholder="contacto@serviciosbajio.com"
                value={adminEmail}
                onChange={(e) => setAdminEmail(e.target.value)}
                disabled={isSubmitting}
                className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                required
              />
              <p className="text-[10px] text-slate-500">
                Se configurará con contraseña provisional <code>Password123!</code> para primer acceso.
              </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  Nombre Completo
                </label>
                <Input
                  type="text"
                  placeholder="Ing. Fernando Gómez"
                  value={adminFullName}
                  onChange={(e) => setAdminFullName(e.target.value)}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-slate-300">
                  Teléfono / Móvil
                </label>
                <Input
                  type="tel"
                  placeholder="4421234567"
                  value={adminPhone}
                  onChange={(e) => setAdminPhone(e.target.value)}
                  disabled={isSubmitting}
                  className="bg-slate-900 border-slate-700 text-slate-100 placeholder:text-slate-500 text-xs h-9"
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
              className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs gap-1.5"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  Registrando...
                </>
              ) : (
                <>
                  <UserPlus className="w-3.5 h-3.5" />
                  Afiliar y Guardar
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
