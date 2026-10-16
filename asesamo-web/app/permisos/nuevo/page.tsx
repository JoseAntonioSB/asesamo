'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import {
  PermitType,
  WorkerComplianceStatus,
  PermitChecklistItem,
  ContractorItem,
} from '@/lib/api/types';
import { CrewComplianceSelector } from '@/components/permisos/crew-compliance-selector';
import { DynamicStpsChecklist } from '@/components/permisos/dynamic-stps-checklist';
import { PermitHazardBadge } from '@/components/permisos/permit-hazard-badge';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ShieldAlert,
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Mountain,
  Flame,
  Zap,
  Clock,
  MapPin,
  FileText,
  AlertCircle,
  Save,
  Send,
} from 'lucide-react';

const HAZARD_OPTIONS: Array<{
  type: PermitType;
  title: string;
  nom: string;
  desc: string;
  icon: React.ComponentType<{ className?: string }>;
  color: string;
}> = [
  {
    type: 'heights',
    title: 'Trabajos en Alturas',
    nom: 'NOM-009-STPS-2011',
    desc: 'Labores a más de 1.80 metros sobre el nivel del suelo, andamios, plataformas y techumbres.',
    icon: Mountain,
    color: 'border-sky-600/70 text-sky-400 bg-sky-950/20',
  },
  {
    type: 'confined_space',
    title: 'Espacios Confinados',
    nom: 'NOM-033-STPS-2015',
    desc: 'Ingreso a tanques, fosas, silos o ductos con ventilación limitada y atmósfera potencialmente peligrosa.',
    icon: ShieldAlert,
    color: 'border-amber-600/70 text-amber-400 bg-amber-950/20',
  },
  {
    type: 'hot_work',
    title: 'Corte y Soldadura',
    nom: 'NOM-027-STPS-2008',
    desc: 'Generación de chispa, flama abierta o calor intenso cerca de materiales combustibles.',
    icon: Flame,
    color: 'border-rose-600/70 text-rose-400 bg-rose-950/20',
  },
  {
    type: 'electrical_loto',
    title: 'LOTO Eléctrico',
    nom: 'NOM-029-STPS-2011',
    desc: 'Mantenimiento e intervención de circuitos eléctricos, tableros y bloqueo de energías peligrosas.',
    icon: Zap,
    color: 'border-yellow-600/70 text-yellow-400 bg-yellow-950/20',
  },
];

function NuevoPermisoWizardContent() {
  const router = useRouter();
  const { user } = useAuth();

  // Wizard Step State (1: Work Details, 2: Crew, 3: Checklist & Submit)
  const [currentStep, setCurrentStep] = useState<1 | 2 | 3>(1);

  // Form Fields
  const [contractorId, setContractorId] = useState<string>('');
  const [contractorsList, setContractorsList] = useState<ContractorItem[]>([]);
  const [permitType, setPermitType] = useState<PermitType>('heights');
  const [workArea, setWorkArea] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [hazardDescription, setHazardDescription] = useState<string>('');
  const [startTime, setStartTime] = useState<string>(() => {
    const d = new Date();
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
    return d.toISOString().slice(0, 16);
  });
  const [endTime, setEndTime] = useState<string>(() => {
    const d = new Date();
    d.setHours(d.getHours() + 8);
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
    return d.toISOString().slice(0, 16);
  });

  // Step 2: Crew Selection State
  const [availableWorkers, setAvailableWorkers] = useState<WorkerComplianceStatus[]>([]);
  const [selectedWorkerIds, setSelectedWorkerIds] = useState<string[]>([]);
  const [isLoadingWorkers, setIsLoadingWorkers] = useState<boolean>(false);

  // Step 3: Checklist State
  const [checklistItems, setChecklistItems] = useState<PermitChecklistItem[]>([]);

  // Submission State
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  // Load Contractor info
  useEffect(() => {
    async function initContractor() {
      try {
        if (user?.role === 'contractor_admin') {
          const comp = await apiClient.getMyContractorCompany();
          if (comp?.id) {
            setContractorId(comp.id);
          }
        } else {
          // If admin, load contractor companies
          let plantId = '22222222-2222-2222-2222-222222222222';
          try {
            const profile = await apiClient.getProfile();
            if (profile?.plant_id) {
              plantId = profile.plant_id;
            }
          } catch {
            // fallback
          }
          const list = await apiClient.listPlantContractors(plantId);
          setContractorsList(list || []);
          if (list && list.length > 0) {
            setContractorId(list[0].id);
          }
        }
      } catch {
        // Fallback or handle error
      }
    }
    initContractor();
  }, [user]);

  // Load Crew Compliance when moving to Step 2 or changing contractor / permitType
  useEffect(() => {
    if (currentStep === 2 && contractorId) {
      async function loadCrew() {
        try {
          setIsLoadingWorkers(true);
          const results = await apiClient.checkCrewCompliance(contractorId, permitType);
          setAvailableWorkers(results || []);
        } catch (err: any) {
          toast.error('Error al evaluar cumplimiento de cuadrilla', {
            description: err.message || 'No se pudieron verificar los certificados DC-3.',
          });
        } finally {
          setIsLoadingWorkers(false);
        }
      }
      loadCrew();
    }
  }, [currentStep, contractorId, permitType]);

  // Navigation handlers
  const handleNextToStep2 = () => {
    if (!workArea.trim()) {
      toast.error('Debe especificar el área o ubicación del trabajo.');
      return;
    }
    if (!description.trim()) {
      toast.error('Debe ingresar la descripción de las labores a realizar.');
      return;
    }
    if (!startTime || !endTime) {
      toast.error('Debe definir la ventana de horario de inicio y fin.');
      return;
    }
    if (new Date(endTime) <= new Date(startTime)) {
      toast.error('La fecha y hora de fin debe ser posterior a la de inicio.');
      return;
    }
    setCurrentStep(2);
  };

  const handleNextToStep3 = () => {
    if (selectedWorkerIds.length === 0) {
      toast.error('Debe seleccionar al menos un trabajador apto con Candado STPS.');
      return;
    }
    setCurrentStep(3);
  };

  // Submission
  const handleSubmitPermit = async (sendToApproval: boolean) => {
    try {
      setIsSubmitting(true);

      // 1. Create Permit Draft
      const startIso = new Date(startTime).toISOString();
      const endIso = new Date(endTime).toISOString();

      const created = await apiClient.createPermit({
        contractor_id: contractorId || undefined,
        permit_type: permitType,
        work_area: workArea.trim(),
        description: description.trim(),
        hazard_description: hazardDescription.trim() || undefined,
        start_time: startIso,
        end_time: endIso,
      });

      // 2. Assign Workers
      if (selectedWorkerIds.length > 0) {
        await apiClient.assignPermitWorkers(created.id, selectedWorkerIds, 'assign');
      }

      // 3. Update Checklist
      if (checklistItems.length > 0) {
        await apiClient.updatePermitChecklist(created.id, {
          items: checklistItems.map((it) => ({
            item_key: it.item_key,
            is_verified: !!(it.is_verified || it.is_checked),
            notes: it.notes || undefined,
            measurements: it.measurements || undefined,
          })),
        });
      }

      // 4. Optionally submit for approval
      if (sendToApproval) {
        await apiClient.transitionPermitStatus(created.id, { action: 'submit' });
        toast.success(`Permiso ${created.permit_number} enviado a revisión EHS.`);
      } else {
        toast.success(`Borrador del permiso ${created.permit_number} guardado con éxito.`);
      }

      router.push(`/permisos/${created.id}`);
    } catch (err: any) {
      toast.error('Error al emitir el permiso de trabajo', {
        description: err.message || 'Verifique que la cuadrilla cumpla con las normativas STPS.',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

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
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-cyan-950 text-cyan-300 font-mono border border-cyan-800">
                  PTAR Digital
                </span>
              </div>
              <p className="text-xs text-slate-400">Emisión de Permiso de Trabajo de Alto Riesgo</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <ThemeToggle />
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1 max-w-4xl mx-auto px-4 sm:px-6 py-8 w-full space-y-6">
        {/* Stepper Navigation */}
        <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-4 shadow-lg backdrop-blur-sm">
          <div className="grid grid-cols-3 gap-2 text-xs font-medium">
            {/* Step 1 */}
            <div
              className={`flex items-center gap-2.5 p-2 rounded-lg transition-all ${
                currentStep === 1
                  ? 'bg-cyan-950/70 border border-cyan-700/80 text-cyan-300'
                  : currentStep > 1
                  ? 'text-emerald-400'
                  : 'text-slate-500'
              }`}
            >
              <div
                className={`w-6 h-6 rounded-full flex items-center justify-center font-bold text-xs shrink-0 ${
                  currentStep === 1
                    ? 'bg-cyan-500 text-slate-950'
                    : currentStep > 1
                    ? 'bg-emerald-500 text-slate-950'
                    : 'bg-slate-800 text-slate-400'
                }`}
              >
                {currentStep > 1 ? '✓' : '1'}
              </div>
              <div className="truncate">
                <span className="block font-bold">Paso 1</span>
                <span className="text-[11px] opacity-80">Datos y Riesgo STPS</span>
              </div>
            </div>

            {/* Step 2 */}
            <div
              className={`flex items-center gap-2.5 p-2 rounded-lg transition-all ${
                currentStep === 2
                  ? 'bg-cyan-950/70 border border-cyan-700/80 text-cyan-300'
                  : currentStep > 2
                  ? 'text-emerald-400'
                  : 'text-slate-500'
              }`}
            >
              <div
                className={`w-6 h-6 rounded-full flex items-center justify-center font-bold text-xs shrink-0 ${
                  currentStep === 2
                    ? 'bg-cyan-500 text-slate-950'
                    : currentStep > 2
                    ? 'bg-emerald-500 text-slate-950'
                    : 'bg-slate-800 text-slate-400'
                }`}
              >
                {currentStep > 2 ? '✓' : '2'}
              </div>
              <div className="truncate">
                <span className="block font-bold">Paso 2</span>
                <span className="text-[11px] opacity-80">Cuadrilla & Candado STPS</span>
              </div>
            </div>

            {/* Step 3 */}
            <div
              className={`flex items-center gap-2.5 p-2 rounded-lg transition-all ${
                currentStep === 3
                  ? 'bg-cyan-950/70 border border-cyan-700/80 text-cyan-300'
                  : 'text-slate-500'
              }`}
            >
              <div
                className={`w-6 h-6 rounded-full flex items-center justify-center font-bold text-xs shrink-0 ${
                  currentStep === 3
                    ? 'bg-cyan-500 text-slate-950'
                    : 'bg-slate-800 text-slate-400'
                }`}
              >
                3
              </div>
              <div className="truncate">
                <span className="block font-bold">Paso 3</span>
                <span className="text-[11px] opacity-80">Checklist & Envío EHS</span>
              </div>
            </div>
          </div>
        </div>

        {/* STEP 1: Work Details and Hazard Type */}
        {currentStep === 1 && (
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-xl">
            <div>
              <h2 className="text-lg font-bold text-white">1. Selección de Riesgo Regulado STPS</h2>
              <p className="text-xs text-slate-400 mt-0.5">
                Seleccione la normativa aplicable para cargar los candados DC-3 y medidas preventivas correspondientes.
              </p>
            </div>

            {/* Hazard Selector Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {HAZARD_OPTIONS.map((opt) => {
                const IconComponent = opt.icon;
                const isSelected = permitType === opt.type;

                return (
                  <button
                    key={opt.type}
                    type="button"
                    onClick={() => setPermitType(opt.type)}
                    className={`text-left p-4 rounded-xl border transition-all duration-200 relative ${
                      isSelected
                        ? `${opt.color} ring-2 ring-cyan-500 shadow-lg scale-[1.01]`
                        : 'bg-slate-950/60 border-slate-800 text-slate-300 hover:border-slate-700'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center gap-2">
                        <IconComponent className="w-5 h-5 shrink-0" />
                        <span className="font-bold text-sm text-white">{opt.title}</span>
                      </div>
                      <span className="text-[10px] font-mono px-2 py-0.5 bg-black/40 rounded border border-white/10 text-slate-300">
                        {opt.nom}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 leading-relaxed">{opt.desc}</p>
                  </button>
                );
              })}
            </div>

            {/* Contractor selector (if admin) */}
            {contractorsList.length > 0 && (
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 block">
                  Empresa Contratista Responsable
                </label>
                <select
                  value={contractorId}
                  onChange={(e) => setContractorId(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-cyan-500 cursor-pointer"
                >
                  {contractorsList.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} {c.rfc ? `(${c.rfc})` : ''}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* Work Area Input */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300 block">
                Área o Ubicación Específica en Planta *
              </label>
              <div className="relative">
                <MapPin className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  placeholder="Ej. Nave 4, Tanque de Almacenamiento T-102, Techo Torre de Enfriamiento..."
                  value={workArea}
                  onChange={(e) => setWorkArea(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                />
              </div>
            </div>

            {/* Description Textarea */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300 block">
                Descripción Detallada del Trabajo *
              </label>
              <textarea
                rows={3}
                placeholder="Describa el objetivo de la actividad, herramientas a utilizar y método de intervención..."
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500 resize-none"
              />
            </div>

            {/* Hazard Analysis Textarea */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300 block">
                Análisis de Riesgos Específicos Identificados (Opcional)
              </label>
              <textarea
                rows={2}
                placeholder="Riesgos colindantes: líneas energizadas cercanas, tránsito de montacargas, vapores..."
                value={hazardDescription}
                onChange={(e) => setHazardDescription(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-cyan-500 resize-none"
              />
            </div>

            {/* Schedule Window */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 block">
                  Fecha y Hora de Inicio Programada *
                </label>
                <div className="relative">
                  <Clock className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="datetime-local"
                    value={startTime}
                    onChange={(e) => setStartTime(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 block">
                  Fecha y Hora de Fin Estimada *
                </label>
                <div className="relative">
                  <Clock className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
                  <input
                    type="datetime-local"
                    value={endTime}
                    onChange={(e) => setEndTime(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-2 text-xs text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                </div>
              </div>
            </div>

            {/* Step 1 Actions */}
            <div className="flex justify-end pt-4 border-t border-slate-800">
              <Button
                type="button"
                onClick={handleNextToStep2}
                className="bg-cyan-600 hover:bg-cyan-500 text-white font-medium text-xs gap-1.5 px-5"
              >
                <span>Continuar a Cuadrilla & Candado STPS</span>
                <ArrowRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}

        {/* STEP 2: Crew Assignment with Candado Inteligente */}
        {currentStep === 2 && (
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-xl">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-white">2. Cuadrilla & Candado Inteligente</h2>
                <p className="text-xs text-slate-400 mt-0.5">
                  Seleccione a los operarios asignados a este permiso. Los trabajadores no certificados son bloqueados automáticamente.
                </p>
              </div>
              <PermitHazardBadge hazardType={permitType} size="sm" />
            </div>

            <CrewComplianceSelector
              workers={availableWorkers}
              selectedWorkerIds={selectedWorkerIds}
              permitType={permitType}
              onChangeSelected={(ids) => setSelectedWorkerIds(ids)}
              isLoading={isLoadingWorkers}
            />

            {/* Step 2 Actions */}
            <div className="flex items-center justify-between pt-4 border-t border-slate-800">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(1)}
                className="border-slate-800 text-slate-300 hover:bg-slate-800 gap-1.5"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Volver a Datos de Trabajo</span>
              </Button>

              <Button
                type="button"
                onClick={handleNextToStep3}
                disabled={selectedWorkerIds.length === 0}
                className="bg-cyan-600 hover:bg-cyan-500 text-white font-medium text-xs gap-1.5 px-5 disabled:opacity-50"
              >
                <span>Continuar a Checklist STPS</span>
                <ArrowRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}

        {/* STEP 3: STPS Dynamic Checklist & Submit */}
        {currentStep === 3 && (
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-6 space-y-6 shadow-xl">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-white">3. Checklist de Seguridad & Emisión</h2>
                <p className="text-xs text-slate-400 mt-0.5">
                  Complete la verificación de controles preventivos antes de solicitar la autorización a EHS.
                </p>
              </div>
              <PermitHazardBadge hazardType={permitType} size="sm" />
            </div>

            <DynamicStpsChecklist
              permitType={permitType}
              onChangeItems={(items) => setChecklistItems(items)}
            />

            {/* Summary Review Banner */}
            <div className="p-4 bg-slate-950/80 border border-slate-800 rounded-xl space-y-2 text-xs">
              <div className="font-bold text-white flex items-center justify-between">
                <span>Resumen del Permiso a Registrar:</span>
                <span className="font-mono text-cyan-400">{selectedWorkerIds.length} operarios listos</span>
              </div>
              <div className="grid grid-cols-2 gap-2 text-slate-400 pt-1 border-t border-slate-800">
                <div>
                  <span className="text-slate-500 block text-[10px]">Ubicación:</span>
                  <span className="text-slate-200">{workArea}</span>
                </div>
                <div>
                  <span className="text-slate-500 block text-[10px]">Horario:</span>
                  <span className="text-slate-200">
                    {startTime.replace('T', ' ')} a {endTime.replace('T', ' ')}
                  </span>
                </div>
              </div>
            </div>

            {/* Step 3 Actions */}
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-4 border-t border-slate-800">
              <Button
                type="button"
                variant="outline"
                onClick={() => setCurrentStep(2)}
                disabled={isSubmitting}
                className="border-slate-800 text-slate-300 hover:bg-slate-800 gap-1.5"
              >
                <ArrowLeft className="w-4 h-4" />
                <span>Modificar Cuadrilla</span>
              </Button>

              <div className="flex items-center gap-2.5">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => handleSubmitPermit(false)}
                  disabled={isSubmitting}
                  className="border-slate-700 text-slate-300 hover:bg-slate-800 gap-1.5 text-xs"
                >
                  <Save className="w-4 h-4" />
                  <span>Guardar Borrador</span>
                </Button>

                <Button
                  type="button"
                  onClick={() => handleSubmitPermit(true)}
                  disabled={isSubmitting}
                  className="bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs gap-1.5 px-5 shadow-lg shadow-emerald-950/50"
                >
                  {isSubmitting ? (
                    <>
                      <div className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      <span>Emitiendo Permiso...</span>
                    </>
                  ) : (
                    <>
                      <Send className="w-4 h-4" />
                      <span>Enviar a Autorización EHS</span>
                    </>
                  )}
                </Button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default function NuevoPermisoPage() {
  return (
    <ProtectedRoute>
      <NuevoPermisoWizardContent />
    </ProtectedRoute>
  );
}
