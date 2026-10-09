'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient, ApiError } from '@/lib/api/client';
import { WorkerCredentialResponse } from '@/lib/api/types';
import { CredentialBadge } from '@/components/credencial/credential-badge';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import {
  ShieldCheck,
  DoorOpen,
  FolderOpen,
  FileStack,
  Printer,
  RefreshCw,
  UploadCloud,
  LogOut,
  Users,
  ChevronRight,
  AlertCircle,
} from 'lucide-react';

// Seed Presets for Instant Demo & Offline Reliability
const DEMO_PRESETS: Record<string, WorkerCredentialResponse> = {
  '66666666-6666-6666-6666-666666666661': {
    worker: {
      user_id: '66666666-6666-6666-6666-666666666661',
      full_name: 'Carlos Mendoza',
      curp: 'MENC850101HDFR01',
      imss_number: '12345678901',
      job_title: 'Soldador Industrial',
      is_active: true,
    },
    contractor: {
      id: '55555555-5555-5555-5555-555555555555',
      name: 'Mantenimiento Industrial del Norte SA de CV',
      rfc: 'MIN120304AB1',
      repse_number: 'REPSE-NL-2023-8841',
      status: 'active',
    },
    plant: {
      id: '22222222-2222-2222-2222-222222222222',
      name: 'Planta Querétaro Norte',
    },
    qr_payload: '66666666-6666-6666-6666-666666666661',
    compliance_status: 'granted',
    documents: [
      {
        document_type_code: 'DC3',
        document_name: 'Constancia de Competencias Laborales DC-3',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2027-02-15',
      },
      {
        document_type_code: 'IMSS',
        document_name: 'Comprobante de Alta IMSS / SUA',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2026-12-31',
      },
      {
        document_type_code: 'MED',
        document_name: 'Examen Médico de Aptitud Física',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2027-04-10',
      },
      {
        document_type_code: 'INE',
        document_name: 'Identificación Oficial Vigente',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2030-01-01',
      },
    ],
    generated_at: new Date().toISOString(),
  },
  '66666666-6666-6666-6666-666666666662': {
    worker: {
      user_id: '66666666-6666-6666-6666-666666666662',
      full_name: 'Juan Perez',
      curp: 'PEPJ900215HDFR02',
      imss_number: '23456789012',
      job_title: 'Técnico Mecánico',
      is_active: true,
    },
    contractor: {
      id: '55555555-5555-5555-5555-555555555555',
      name: 'Mantenimiento Industrial del Norte SA de CV',
      rfc: 'MIN120304AB1',
      repse_number: 'REPSE-NL-2023-8841',
      status: 'active',
    },
    plant: {
      id: '22222222-2222-2222-2222-222222222222',
      name: 'Planta Querétaro Norte',
    },
    qr_payload: '66666666-6666-6666-6666-666666666662',
    compliance_status: 'denied',
    denial_reasons: ['Documento DC-3 vencido el 2024-01-15.'],
    documents: [
      {
        document_type_code: 'DC3',
        document_name: 'Constancia de Competencias Laborales DC-3',
        applies_to: 'worker',
        status: 'expired',
        is_required: true,
        is_valid: false,
        expires_at: '2024-01-15',
        failure_reason: 'Constancia vencida según NOM-STPS.',
      },
      {
        document_type_code: 'IMSS',
        document_name: 'Comprobante de Alta IMSS / SUA',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2026-11-20',
      },
    ],
    generated_at: new Date().toISOString(),
  },
  '66666666-6666-6666-6666-666666666663': {
    worker: {
      user_id: '66666666-6666-6666-6666-666666666663',
      full_name: 'Roberto Gomez',
      curp: 'GOMR880320HDFR03',
      imss_number: '34567890123',
      job_title: 'Electricista Industrial',
      is_active: true,
    },
    contractor: {
      id: '55555555-5555-5555-5555-555555555555',
      name: 'Mantenimiento Industrial del Norte SA de CV',
      rfc: 'MIN120304AB1',
      repse_number: 'REPSE-NL-2023-8841',
      status: 'active',
    },
    plant: {
      id: '22222222-2222-2222-2222-222222222222',
      name: 'Planta Querétaro Norte',
    },
    qr_payload: '66666666-6666-6666-6666-666666666663',
    compliance_status: 'granted_with_warning',
    warning_reasons: ['Constancia DC-3 vence en 12 días.'],
    documents: [
      {
        document_type_code: 'DC3',
        document_name: 'Constancia de Competencias Laborales DC-3',
        applies_to: 'worker',
        status: 'approved',
        is_required: true,
        is_valid: true,
        expires_at: '2026-10-20',
      },
    ],
    generated_at: new Date().toISOString(),
  },
};

function CredencialContent() {
  const { user, logout } = useAuth();
  const searchParams = useSearchParams();

  // Selected Worker ID (from query param, user ID if worker, or default preset)
  const initialWorkerId =
    searchParams.get('worker_id') ||
    (user?.role === 'worker' ? user.id : '66666666-6666-6666-6666-666666666661');

  const [selectedWorkerId, setSelectedWorkerId] = useState<string>(initialWorkerId);
  const [credential, setCredential] = useState<WorkerCredentialResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const isGuardOrAdmin =
    user?.role === 'guard' ||
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin' ||
    user?.role === 'ehs_manager' ||
    user?.role === 'contractor_admin';

  const isEhsOrAdmin =
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin' ||
    user?.role === 'ehs_manager';

  const loadCredential = useCallback(
    async (workerId: string, showToast = false) => {
      try {
        if (showToast) setIsRefreshing(true);
        else setIsLoading(true);

        const data = await apiClient.getWorkerCredential(workerId);
        setCredential(data);

        if (showToast) {
          toast.success('Credencial sincronizada con caseta.');
        }
      } catch (err: unknown) {
        // Fallback to local demo preset if API is in demo mode
        if (DEMO_PRESETS[workerId]) {
          setCredential(DEMO_PRESETS[workerId]);
        } else {
          toast.error('No se pudo cargar la credencial del trabajador.');
        }
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
      }
    },
    []
  );

  useEffect(() => {
    loadCredential(selectedWorkerId);
  }, [selectedWorkerId, loadCredential]);

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-emerald-500 selection:text-white pb-12">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40 no-print">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 font-mono border border-emerald-500/30">
                  GAFETE DIGITAL
                </span>
              </div>
              <p className="text-[11px] text-slate-400">Credencial Móvil del Trabajador</p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {isGuardOrAdmin && (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                  title="Ir a Caseta de Vigilancia"
                >
                  <DoorOpen className="w-3.5 h-3.5 text-cyan-400" />
                  <span className="hidden sm:inline">Caseta</span>
                </Button>
              </Link>
            )}

            <Link href="/documentos">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Ver Expediente de Documentos"
              >
                <FolderOpen className="w-3.5 h-3.5 text-emerald-400" />
                <span className="hidden sm:inline">Documentos</span>
              </Button>
            </Link>

            {isEhsOrAdmin && (
              <Link href="/ehs/aprobaciones">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-amber-800/60 bg-amber-950/30 text-amber-300 hover:text-white hover:bg-amber-900/50 text-xs gap-1.5"
                  title="Bandeja EHS"
                >
                  <FileStack className="w-3.5 h-3.5 text-amber-400" />
                  <span className="hidden sm:inline">Bandeja EHS</span>
                </Button>
              </Link>
            )}

            {/* Theme Toggle Button */}
            <ThemeToggle />

            <Button
              size="sm"
              variant="ghost"
              onClick={logout}
              className="text-slate-400 hover:text-rose-400 hover:bg-rose-950/20 h-9 px-2.5"
              title="Cerrar sesión"
            >
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1 max-w-md w-full mx-auto px-4 sm:px-0 py-6 space-y-6">
        {/* Worker Selector (for Admins / Guards / Testing Demonstration) */}
        {isGuardOrAdmin && (
          <div className="p-3 rounded-2xl bg-slate-900/80 border border-slate-800 no-print space-y-2">
            <div className="flex items-center justify-between text-xs">
              <span className="font-bold text-slate-300 flex items-center gap-1.5">
                <Users className="w-3.5 h-3.5 text-emerald-400" />
                Simulador de Trabajador
              </span>
              <span className="text-[10px] text-slate-500 font-mono">Modo Demostración</span>
            </div>

            <div className="grid grid-cols-3 gap-1.5 text-xs">
              <button
                type="button"
                onClick={() => setSelectedWorkerId('66666666-6666-6666-6666-666666666661')}
                className={`py-1.5 px-2 rounded-lg border text-center transition-all cursor-pointer ${
                  selectedWorkerId === '66666666-6666-6666-6666-666666666661'
                    ? 'bg-emerald-950/60 border-emerald-500 text-emerald-300 font-bold'
                    : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                Carlos (Válido)
              </button>

              <button
                type="button"
                onClick={() => setSelectedWorkerId('66666666-6666-6666-6666-666666666662')}
                className={`py-1.5 px-2 rounded-lg border text-center transition-all cursor-pointer ${
                  selectedWorkerId === '66666666-6666-6666-6666-666666666662'
                    ? 'bg-rose-950/60 border-rose-500 text-rose-300 font-bold'
                    : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                Juan (Vencido)
              </button>

              <button
                type="button"
                onClick={() => setSelectedWorkerId('66666666-6666-6666-6666-666666666663')}
                className={`py-1.5 px-2 rounded-lg border text-center transition-all cursor-pointer ${
                  selectedWorkerId === '66666666-6666-6666-6666-666666666663'
                    ? 'bg-amber-950/60 border-amber-500 text-amber-300 font-bold'
                    : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
                }`}
              >
                Roberto (Aviso)
              </button>
            </div>
          </div>
        )}

        {/* Loading Spinner */}
        {isLoading && !credential ? (
          <div className="py-20 text-center">
            <RefreshCw className="w-8 h-8 text-emerald-400 animate-spin mx-auto mb-3" />
            <p className="text-sm text-slate-400">Generando credencial oficial...</p>
          </div>
        ) : credential ? (
          <>
            {/* The Physical Credential Badge */}
            <CredentialBadge data={credential} />

            {/* Quick Actions Bar */}
            <div className="flex items-center gap-2 pt-2 no-print">
              <Button
                variant="outline"
                size="sm"
                onClick={() => loadCredential(selectedWorkerId, true)}
                disabled={isRefreshing}
                className="flex-1 border-slate-800 bg-slate-900 text-slate-300 hover:text-white h-11 gap-1.5 rounded-xl"
              >
                <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
                <span>Actualizar</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={handlePrint}
                className="flex-1 border-slate-800 bg-slate-900 text-slate-300 hover:text-white h-11 gap-1.5 rounded-xl"
                title="Imprimir carnet de acceso en formato oficial"
              >
                <Printer className="w-4 h-4 text-cyan-400" />
                <span>Imprimir</span>
              </Button>

              {credential.compliance_status !== 'granted' && (
                <Link href="/documentos" className="flex-1">
                  <Button
                    size="sm"
                    className="w-full bg-emerald-600 hover:bg-emerald-500 text-white h-11 gap-1.5 rounded-xl shadow-lg shadow-emerald-950/40"
                  >
                    <UploadCloud className="w-4 h-4" />
                    <span>Renovar</span>
                  </Button>
                </Link>
              )}
            </div>

            {/* Prompt for Caseta Check-In Flow */}
            <div className="p-3.5 rounded-xl bg-slate-900/60 border border-slate-800/80 text-xs text-slate-400 flex items-start gap-2.5 no-print">
              <AlertCircle className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
              <div>
                <p className="font-semibold text-slate-300">Instrucciones de Ingreso:</p>
                <p className="mt-0.5 leading-relaxed">
                  Muestre el código QR al oficial de caseta o aproxímelo al lector del torniquete. Si el semáforo está en verde, el acceso se concederá automáticamente.
                </p>
              </div>
            </div>
          </>
        ) : null}
      </main>
    </div>
  );
}

export default function CredencialPage() {
  return (
    <ProtectedRoute>
      <CredencialContent />
    </ProtectedRoute>
  );
}
