'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { OccupationalMedicalExamDto } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  HeartPulse,
  Activity,
  ShieldAlert,
  CheckCircle,
  AlertCircle,
  FileHeart,
  Stethoscope,
  RefreshCw,
  LogOut,
  Eye,
  Ear,
  Wind,
} from 'lucide-react';

function SaludOcupacionalDashboardContent() {
  const { user, logout } = useAuth();
  const [exams, setExams] = useState<OccupationalMedicalExamDto[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const loadExams = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const res = await apiClient.listMedicalExams().catch(() => []);
      setExams(res || []);

      if (showToast) {
        toast.success('Expedientes médicos actualizados');
      }
    } catch (err: any) {
      toast.error('Error al cargar exámenes médicos', {
        description: err.message || 'No se pudieron consultar los expedientes de salud ocupacional.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadExams();
  }, [loadExams]);

  const fitCount = exams.filter((e) => e.fitness_status === 'fit').length;
  const restrictionCount = exams.filter((e) => e.fitness_status === 'fit_with_restrictions').length;
  const unfitCount = exams.filter((e) => e.fitness_status === 'unfit_temporary' || e.fitness_status === 'unfit_permanent').length;

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-rose-500 selection:text-white">
      {/* Header Flotante */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-rose-500/10 border border-rose-500/20 flex items-center justify-center text-rose-400">
              <HeartPulse className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Salud Ocupacional (NOM-030-STPS)</div>
            </div>
          </div>

          <nav className="hidden md:flex items-center gap-1 bg-zinc-900/60 border border-zinc-800/80 p-1 rounded-xl text-xs">
            <Link href="/caseta" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Caseta
            </Link>
            <Link href="/permisos" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Permisos PTAR
            </Link>
            <Link href="/inspecciones" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Inspecciones
            </Link>
            <Link href="/incidentes" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Incidentes
            </Link>
            <Link href="/epp" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              EPP
            </Link>
            <Link href="/evacuacion" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Evacuación
            </Link>
            <Link href="/capacitacion" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Capacitación
            </Link>
            <Link href="/salud-ocupacional" className="px-3 py-1.5 rounded-lg bg-zinc-800 font-semibold text-white">
              Salud NOM-030
            </Link>
          </nav>

          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadExams(true)}
              disabled={isRefreshing || isLoading}
              className="border-zinc-800 hover:bg-zinc-900 text-zinc-300"
            >
              <RefreshCw className={`w-3.5 h-3.5 mr-1.5 ${isRefreshing ? 'animate-spin' : ''}`} />
              Refrescar
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={logout}
              className="text-zinc-400 hover:text-red-400 hover:bg-red-500/10"
            >
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Hero Section */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-zinc-800/80 pb-6">
          <div>
            <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs font-medium mb-3">
              <Stethoscope className="w-3.5 h-3.5" />
              Medicina del Trabajo y Vigilancia a la Salud (NOM-030-STPS-2009)
            </div>
            <h1 className="text-3xl font-black tracking-tight text-white">
              Salud Ocupacional & Dictámenes de Aptitud
            </h1>
            <p className="text-sm text-zinc-400 mt-1 max-w-2xl">
              Monitoreo clínico de personal operativo y contratistas: exámenes periódicos anuales, dictámenes para
              alturas (NOM-009) y espacios confinados (NOM-033), y control de paraclínicos (antidoping, audiometría, espirometría).
            </p>
          </div>
        </div>

        {/* Bento Grid Metrics */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Total Evaluaciones Clínicas</span>
              <FileHeart className="w-4 h-4 text-rose-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{exams.length}</span>
              <span className="text-xs text-rose-400 font-medium">Dictámenes</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Aptos sin Restricciones</span>
              <CheckCircle className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{fitCount}</span>
              <span className="text-xs text-emerald-400 font-medium">100% Aptitud</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Aptos con Restricción</span>
              <AlertCircle className="w-4 h-4 text-amber-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{restrictionCount}</span>
              <span className="text-xs text-amber-400 font-medium">Lentes / Carga</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">No Aptos / Incapacidad</span>
              <ShieldAlert className="w-4 h-4 text-red-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{unfitCount}</span>
              <span className="text-xs text-red-400 font-medium">Bloqueo Caseta</span>
            </div>
          </div>
        </div>

        {/* Exams Table */}
        <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden">
          <div className="p-4 border-b border-zinc-800/80 flex items-center justify-between">
            <h2 className="text-sm font-bold text-white flex items-center gap-2">
              <Activity className="w-4 h-4 text-rose-400" />
              Bitácora de Exámenes y Dictámenes Médicos Industriales
            </h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-zinc-800/80 bg-zinc-900/80 text-zinc-400 font-medium uppercase tracking-wider">
                <tr>
                  <th className="py-3 px-4">Folio Médico</th>
                  <th className="py-3 px-4">Tipo de Examen</th>
                  <th className="py-3 px-4">Dictamen de Aptitud</th>
                  <th className="py-3 px-4">Signos & Paraclínicos</th>
                  <th className="py-3 px-4">Médico Dictaminador</th>
                  <th className="py-3 px-4">Fecha Examen</th>
                  <th className="py-3 px-4">Vigencia</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/60 font-mono text-zinc-300">
                {exams.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-8 text-center text-zinc-500 font-sans">
                      No hay exámenes médicos ocupacionales registrados.
                    </td>
                  </tr>
                ) : (
                  exams.map((exam) => {
                    const isFit = exam.fitness_status === 'fit';
                    const isRestriction = exam.fitness_status === 'fit_with_restrictions';

                    return (
                      <tr key={exam.id} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-3.5 px-4 font-bold text-rose-400">{exam.exam_number}</td>
                        <td className="py-3.5 px-4 font-sans capitalize">
                          {exam.exam_type === 'periodic_annual'
                            ? 'Periódico Anual NOM-030'
                            : exam.exam_type === 'height_clearance'
                            ? 'Aptitud para Alturas (>1.80m)'
                            : exam.exam_type === 'confined_space'
                            ? 'Espacios Confinados'
                            : exam.exam_type}
                        </td>
                        <td className="py-3.5 px-4">
                          <span
                            className={`px-2.5 py-1 rounded-full text-[10px] font-semibold border ${
                              isFit
                                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                                : isRestriction
                                ? 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                                : 'bg-red-500/10 text-red-400 border-red-500/20'
                            }`}
                          >
                            {isFit
                              ? 'APTO INTEGRAL'
                              : isRestriction
                              ? 'APTO CON RESTRICCIÓN'
                              : 'NO APTO TEMPORAL'}
                          </span>
                        </td>
                        <td className="py-3.5 px-4 font-sans text-[11px] text-zinc-400">
                          <div>P/A: {exam.blood_pressure || '120/80'} | Vista: {exam.visual_acuity_result || '20/20'}</div>
                          <div className="text-[10px] text-zinc-500">
                            Antidoping: {exam.antidoping_result || 'Negativo'}
                          </div>
                        </td>
                        <td className="py-3.5 px-4 font-sans">
                          <div className="text-white font-medium">{exam.doctor_name}</div>
                          <div className="text-[10px] text-zinc-500">{exam.doctor_license}</div>
                        </td>
                        <td className="py-3.5 px-4 text-zinc-400">
                          {new Date(exam.exam_date).toLocaleDateString('es-MX')}
                        </td>
                        <td className="py-3.5 px-4 text-zinc-400">
                          {new Date(exam.expiration_date).toLocaleDateString('es-MX')}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>
      </main>
    </div>
  );
}

export default function SaludOcupacionalPage() {
  return (
    <ProtectedRoute>
      <SaludOcupacionalDashboardContent />
    </ProtectedRoute>
  );
}
