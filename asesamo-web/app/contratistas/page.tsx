'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { ContractorItem, ContractorWorkerItem, PlantContractorStatus } from '@/lib/api/types';
import { ContractorInviteModal } from '@/components/contratistas/contractor-invite-modal';
import { WorkerRegisterModal } from '@/components/contratistas/worker-register-modal';
import { ContractorListTable } from '@/components/contratistas/contractor-list-table';
import { CrewListTable } from '@/components/contratistas/crew-list-table';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ShieldCheck,
  Building2,
  HardHat,
  Users,
  UserPlus,
  RefreshCw,
  LogOut,
  ArrowLeft,
  FileStack,
  FileText,
  QrCode,
  CheckCircle2,
  AlertTriangle,
  BadgeAlert,
  Briefcase,
} from 'lucide-react';

const DEFAULT_PLANT_ID = '22222222-2222-2222-2222-222222222222'; // Planta Querétaro Norte

function ContratistasContent() {
  const { user, logout } = useAuth();

  const isContractorAdmin = user?.role === 'contractor_admin' || user?.role === 'contractor';
  const isPlantOrEhsAdmin =
    user?.role === 'plant_admin' ||
    user?.role === 'ehs_manager' ||
    user?.role === 'super_admin' ||
    user?.role === 'guard';

  // Data states
  const [contractors, setContractors] = useState<ContractorItem[]>([]);
  const [myCompany, setMyCompany] = useState<ContractorItem | null>(null);
  const [selectedContractor, setSelectedContractor] = useState<ContractorItem | null>(null);
  const [workers, setWorkers] = useState<ContractorWorkerItem[]>([]);

  // UI state
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [inviteModalOpen, setInviteModalOpen] = useState<boolean>(false);
  const [workerModalOpen, setWorkerModalOpen] = useState<boolean>(false);

  // Load data depending on user role
  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      if (isContractorAdmin) {
        // Load company & workers
        const company = await apiClient.getMyContractorCompany();
        setMyCompany(company);
        setSelectedContractor(company);

        const crew = await apiClient.listContractorWorkers(company.id);
        setWorkers(crew);
      } else {
        // Plant/EHS view
        const list = await apiClient.listPlantContractors(DEFAULT_PLANT_ID);
        setContractors(list);

        // If a contractor was already selected, refresh their workers too
        if (selectedContractor) {
          const crew = await apiClient.listContractorWorkers(selectedContractor.id);
          setWorkers(crew);
        }
      }

      if (showToast) {
        toast.success('Padrón de contratistas actualizado');
      }
    } catch {
      toast.error('Error al cargar datos de contratistas', {
        description: 'No se pudo conectar con el servicio de gestión de contratistas.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, [isContractorAdmin, selectedContractor]);

  useEffect(() => {
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // When clicking on a contractor to inspect their crew (plant view)
  const handleSelectContractor = async (contractor: ContractorItem) => {
    setSelectedContractor(contractor);
    try {
      setIsLoading(true);
      const crew = await apiClient.listContractorWorkers(contractor.id);
      setWorkers(crew);
    } catch {
      toast.error('Error al cargar la cuadrilla', {
        description: `No se pudo obtener la plantilla de ${contractor.name}.`,
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Back to contractor list (plant view)
  const handleBackToContractors = () => {
    setSelectedContractor(null);
    setWorkers([]);
  };

  // Status change handler
  const handleUpdateContractorStatus = async (contractorId: string, newStatus: PlantContractorStatus) => {
    try {
      await apiClient.updateContractorStatus(contractorId, {
        plant_id: DEFAULT_PLANT_ID,
        status: newStatus,
      });

      setContractors((prev) =>
        prev.map((c) => (c.id === contractorId ? { ...c, plant_status: newStatus } : c))
      );

      if (selectedContractor?.id === contractorId) {
        setSelectedContractor((prev) => (prev ? { ...prev, plant_status: newStatus } : null));
      }

      toast.success('Estatus actualizado correctamente');
    } catch {
      toast.error('Error al actualizar el estatus del contratista');
    }
  };

  // Worker status toggle
  const handleToggleWorkerStatus = async (workerId: string, currentActive: boolean) => {
    try {
      const activeContractorId = isContractorAdmin ? myCompany?.id : selectedContractor?.id;
      const res = await apiClient.toggleContractorWorkerStatus(
        workerId,
        !currentActive,
        activeContractorId
      );

      setWorkers((prev) =>
        prev.map((w) => (w.worker_id === workerId ? { ...w, is_active: res.is_active } : w))
      );

      toast.success(
        res.is_active ? 'Trabajador dado de alta en cuadrilla' : 'Trabajador marcado como inactivo'
      );
    } catch {
      toast.error('Error al actualizar estatus del trabajador');
    }
  };

  // Invite modal success
  const handleInviteSuccess = (newContractor: ContractorItem) => {
    setContractors((prev) => [newContractor, ...prev]);
  };

  // Worker register success
  const handleWorkerAdded = (newWorker: ContractorWorkerItem) => {
    setWorkers((prev) => [newWorker, ...prev]);
    // Increment workers_count
    if (selectedContractor) {
      setSelectedContractor((prev) =>
        prev ? { ...prev, workers_count: (prev.workers_count || 0) + 1 } : null
      );
    }
    setContractors((prev) =>
      prev.map((c) =>
        c.id === (isContractorAdmin ? myCompany?.id : selectedContractor?.id)
          ? { ...c, workers_count: (c.workers_count || 0) + 1 }
          : c
      )
    );
  };

  // Metrics calculation
  const totalWorkers = workers.length;
  const activeWorkers = workers.filter((w) => w.is_active).length;
  const inactiveWorkers = totalWorkers - activeWorkers;

  const totalContractors = contractors.length;
  const activeContractors = contractors.filter((c) => c.plant_status === 'active').length;
  const pendingContractors = contractors.filter((c) => c.plant_status === 'pending').length;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-emerald-500 selection:text-white">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Building2 className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 font-mono border border-slate-700">
                  Contratistas & Cuadrillas
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Padrón de Empresas Proveedoras y Personal Operativo
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {isPlantOrEhsAdmin && (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                >
                  <ArrowLeft className="w-3.5 h-3.5" />
                  <span className="hidden sm:inline">Caseta</span>
                </Button>
              </Link>
            )}

            <Link href="/documentos">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Gestión de Documentos"
              >
                <FileText className="w-3.5 h-3.5 text-emerald-400" />
                <span className="hidden sm:inline">Documentos</span>
              </Button>
            </Link>

            {isPlantOrEhsAdmin && (
              <Link href="/ehs/aprobaciones">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-amber-800/60 bg-amber-950/30 text-amber-300 hover:text-white hover:bg-amber-900/50 text-xs gap-1.5"
                  title="Bandeja de Aprobación EHS"
                >
                  <FileStack className="w-3.5 h-3.5 text-amber-400" />
                  <span className="hidden md:inline">Bandeja EHS</span>
                </Button>
              </Link>
            )}

            <Link href="/credencial">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-cyan-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Ver Credencial Móvil QR"
              >
                <QrCode className="w-3.5 h-3.5 text-cyan-400" />
                <span className="hidden lg:inline">Mi Credencial</span>
              </Button>
            </Link>

            {/* Theme Toggle */}
            <ThemeToggle />

            {/* Logout */}
            <Button
              size="sm"
              variant="ghost"
              onClick={logout}
              className="text-slate-400 hover:text-red-400 hover:bg-slate-800 text-xs gap-1.5 px-2"
              title="Cerrar sesión"
            >
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 flex-1 w-full space-y-6">
        {/* Banner / Header Title Area */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-slate-800/80">
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
                {isContractorAdmin
                  ? myCompany?.name || 'Portal del Contratista'
                  : selectedContractor
                  ? `Cuadrilla: ${selectedContractor.name}`
                  : 'Gestión de Contratistas y Cuadrillas'}
              </h1>
              {isContractorAdmin && myCompany && (
                <span className="px-2 py-0.5 rounded-full text-xs font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {myCompany.rfc}
                </span>
              )}
            </div>
            <p className="text-xs text-slate-400 mt-1">
              {isContractorAdmin
                ? 'Administre los trabajadores de su cuadrilla y valide su estatus de ingreso antes de acudir a planta.'
                : selectedContractor
                ? `Padrón de operarios y estatus laboral registrado para ${selectedContractor.name}.`
                : 'Padrón de empresas contratistas autorizadas para laborar en Planta Querétaro Norte.'}
            </p>
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 flex-wrap">
            <Button
              size="sm"
              variant="outline"
              onClick={() => loadData(true)}
              disabled={isLoading || isRefreshing}
              className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
              Actualizar
            </Button>

            {/* If in selected contractor view, show back button */}
            {!isContractorAdmin && selectedContractor && (
              <Button
                size="sm"
                variant="outline"
                onClick={handleBackToContractors}
                className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
              >
                <ArrowLeft className="w-3.5 h-3.5" />
                Volver a Contratistas
              </Button>
            )}

            {/* Plant view: Invite Contractor */}
            {!isContractorAdmin && !selectedContractor && (
              <Button
                size="sm"
                onClick={() => setInviteModalOpen(true)}
                className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs gap-1.5 shadow-md shadow-emerald-900/30"
              >
                <Building2 className="w-3.5 h-3.5" />
                Invitar Contratista
              </Button>
            )}

            {/* Crew view (Contractor Admin OR Plant Admin inspecting contractor): Register Worker */}
            {(isContractorAdmin || selectedContractor) && (
              <Button
                size="sm"
                onClick={() => setWorkerModalOpen(true)}
                className="bg-cyan-600 hover:bg-cyan-500 text-white text-xs gap-1.5 shadow-md shadow-cyan-900/30"
              >
                <UserPlus className="w-3.5 h-3.5" />
                Registrar Trabajador
              </Button>
            )}
          </div>
        </div>

        {/* KPI Summary Cards */}
        {isContractorAdmin || selectedContractor ? (
          /* Crew KPIs */
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/50 flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400 font-medium">Personal en Cuadrilla</p>
                <p className="text-2xl font-bold text-white mt-1">{totalWorkers}</p>
                <p className="text-[11px] text-slate-500 mt-0.5">Operarios registrados</p>
              </div>
              <div className="p-3 rounded-lg bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                <Users className="w-6 h-6" />
              </div>
            </div>

            <div className="p-4 rounded-xl border border-emerald-900/40 bg-emerald-950/20 flex items-center justify-between">
              <div>
                <p className="text-xs text-emerald-300 font-medium">Personal Activo</p>
                <p className="text-2xl font-bold text-emerald-400 mt-1">{activeWorkers}</p>
                <p className="text-[11px] text-emerald-500/80 mt-0.5">Autorizados en planta</p>
              </div>
              <div className="p-3 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <CheckCircle2 className="w-6 h-6" />
              </div>
            </div>

            <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/50 flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400 font-medium">Personal Inactivo</p>
                <p className="text-2xl font-bold text-slate-300 mt-1">{inactiveWorkers}</p>
                <p className="text-[11px] text-slate-500 mt-0.5">Bajas o suspensiones</p>
              </div>
              <div className="p-3 rounded-lg bg-slate-800 text-slate-400 border border-slate-700">
                <HardHat className="w-6 h-6" />
              </div>
            </div>
          </div>
        ) : (
          /* Plant Contractor KPIs */
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/50 flex items-center justify-between">
              <div>
                <p className="text-xs text-slate-400 font-medium">Total Empresas Afiliadas</p>
                <p className="text-2xl font-bold text-white mt-1">{totalContractors}</p>
                <p className="text-[11px] text-slate-500 mt-0.5">Padrón de proveedores</p>
              </div>
              <div className="p-3 rounded-lg bg-slate-800 text-slate-300 border border-slate-700">
                <Building2 className="w-6 h-6" />
              </div>
            </div>

            <div className="p-4 rounded-xl border border-emerald-900/40 bg-emerald-950/20 flex items-center justify-between">
              <div>
                <p className="text-xs text-emerald-300 font-medium">Contratistas Activos</p>
                <p className="text-2xl font-bold text-emerald-400 mt-1">{activeContractors}</p>
                <p className="text-[11px] text-emerald-500/80 mt-0.5">Autorizados para ingresar</p>
              </div>
              <div className="p-3 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <CheckCircle2 className="w-6 h-6" />
              </div>
            </div>

            <div className="p-4 rounded-xl border border-amber-900/40 bg-amber-950/20 flex items-center justify-between">
              <div>
                <p className="text-xs text-amber-300 font-medium">Pendientes o Revisión</p>
                <p className="text-2xl font-bold text-amber-400 mt-1">{pendingContractors}</p>
                <p className="text-[11px] text-amber-500/80 mt-0.5">En trámite de alta</p>
              </div>
              <div className="p-3 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <AlertTriangle className="w-6 h-6" />
              </div>
            </div>
          </div>
        )}

        {/* Content Table Area */}
        {isContractorAdmin || selectedContractor ? (
          /* Cuadrilla View */
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-base font-semibold text-white flex items-center gap-2">
                <HardHat className="w-4 h-4 text-cyan-400" />
                <span>Personal de la Cuadrilla</span>
              </h2>
              <span className="text-xs text-slate-400 font-mono">
                {workers.length} operarios registrados
              </span>
            </div>

            <CrewListTable
              workers={workers}
              isLoading={isLoading}
              onToggleStatus={handleToggleWorkerStatus}
              canManageWorkers={true}
              contractorName={isContractorAdmin ? myCompany?.name : selectedContractor?.name}
            />
          </div>
        ) : (
          /* Padrón de Contratistas View */
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h2 className="text-base font-semibold text-white flex items-center gap-2">
                <Building2 className="w-4 h-4 text-emerald-400" />
                <span>Empresas Contratistas Registradas</span>
              </h2>
              <span className="text-xs text-slate-400 font-mono">
                {contractors.length} empresas en padrón
              </span>
            </div>

            <ContractorListTable
              contractors={contractors}
              isLoading={isLoading}
              onSelectContractor={handleSelectContractor}
              onUpdateStatus={handleUpdateContractorStatus}
              canManageStatus={isPlantOrEhsAdmin}
            />
          </div>
        )}
      </main>

      {/* Invite Contractor Modal */}
      <ContractorInviteModal
        open={inviteModalOpen}
        onOpenChange={setInviteModalOpen}
        plantId={DEFAULT_PLANT_ID}
        onInviteSuccess={handleInviteSuccess}
      />

      {/* Register Worker Modal */}
      <WorkerRegisterModal
        open={workerModalOpen}
        onOpenChange={setWorkerModalOpen}
        contractorId={isContractorAdmin ? myCompany?.id : selectedContractor?.id}
        contractorName={isContractorAdmin ? myCompany?.name : selectedContractor?.name}
        onWorkerAdded={handleWorkerAdded}
      />
    </div>
  );
}

export default function ContratistasPage() {
  return (
    <ProtectedRoute>
      <ContratistasContent />
    </ProtectedRoute>
  );
}
