'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { TrainingCourseDto, TrainingCertificateDto } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  GraduationCap,
  Award,
  BookOpen,
  CheckCircle2,
  Calendar,
  Building2,
  RefreshCw,
  LogOut,
  FileText,
  Clock,
  Sparkles,
  Printer,
} from 'lucide-react';

function CapacitacionDashboardContent() {
  const { user, logout } = useAuth();
  const [courses, setCourses] = useState<TrainingCourseDto[]>([]);
  const [certificates, setCertificates] = useState<TrainingCertificateDto[]>([]);
  const [activeTab, setActiveTab] = useState<'certificates' | 'courses'>('certificates');
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);

  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const [coursesData, certsData] = await Promise.all([
        apiClient.listTrainingCourses().catch(() => []),
        apiClient.listTrainingCertificates().catch(() => []),
      ]);

      setCourses(coursesData);
      setCertificates(certsData);

      if (showToast) {
        toast.success('Capacitación y constancias DC-3 actualizadas');
      }
    } catch (err: any) {
      toast.error('Error al cargar datos de capacitación', {
        description: err.message || 'No se pudieron consultar los cursos o constancias STPS.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const activeCerts = certificates.filter((c) => c.status === 'issued');
  const avgScore =
    certificates.length > 0
      ? (certificates.reduce((acc, c) => acc + (c.score_percentage || 100), 0) / certificates.length).toFixed(1)
      : '100.0';

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-indigo-500 selection:text-white">
      {/* Header Flotante Frosted Glass */}
      <header className="sticky top-0 z-40 w-full border-b border-zinc-800/80 bg-zinc-950/80 backdrop-blur-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
              <GraduationCap className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-black font-mono tracking-wider text-white">ASESAMO</div>
              <div className="text-[10px] text-zinc-400 font-medium">Capacitación & Habilidades STPS (DC-3)</div>
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
            <Link href="/capacitacion" className="px-3 py-1.5 rounded-lg bg-zinc-800 font-semibold text-white">
              Capacitación
            </Link>
            <Link href="/salud-ocupacional" className="px-3 py-1.5 rounded-lg text-zinc-400 hover:text-zinc-200">
              Salud NOM-030
            </Link>
          </nav>

          <div className="flex items-center gap-2">
            <ThemeToggle />
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadData(true)}
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
            <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-medium mb-3">
              <Sparkles className="w-3.5 h-3.5" />
              Cumplimiento Normativo STPS (Formatos DC-1, DC-2, DC-3 & DC-4)
            </div>
            <h1 className="text-3xl font-black tracking-tight text-white">
              Capacitación, Adiestramiento y Habilidades Laborales
            </h1>
            <p className="text-sm text-zinc-400 mt-1 max-w-2xl">
              Control digital de cursos certificados y emisión oficial de constancias DC-3 vinculadas automáticamente
              con los candados de acceso a caseta y permisos de trabajo de alto riesgo (PTAR).
            </p>
          </div>
        </div>

        {/* Bento Grid Metrics */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Constancias DC-3 Vigentes</span>
              <Award className="w-4 h-4 text-indigo-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{activeCerts.length}</span>
              <span className="text-xs text-indigo-400 font-medium">Acreditadas</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Catálogo de Cursos Activos</span>
              <BookOpen className="w-4 h-4 text-emerald-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{courses.length}</span>
              <span className="text-xs text-emerald-400 font-medium">STPS</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Promedio de Acreditación</span>
              <CheckCircle2 className="w-4 h-4 text-amber-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">{avgScore}%</span>
              <span className="text-xs text-amber-400 font-medium">Evaluación</span>
            </div>
          </div>

          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/50 p-5 backdrop-blur-sm relative overflow-hidden">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-400">Vigencia y Recertificación</span>
              <Clock className="w-4 h-4 text-sky-400" />
            </div>
            <div className="mt-4 flex items-baseline gap-2">
              <span className="text-3xl font-black tracking-tight text-white font-mono">1 - 2 Años</span>
              <span className="text-xs text-sky-400 font-medium">Anual STPS</span>
            </div>
          </div>
        </div>

        {/* Tab Controls */}
        <div className="flex items-center gap-2 border-b border-zinc-800/80 pb-3">
          <button
            onClick={() => setActiveTab('certificates')}
            className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all ${
              activeTab === 'certificates'
                ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                : 'bg-zinc-900 text-zinc-400 hover:text-white'
            }`}
          >
            Constancias DC-3 Emitidas ({certificates.length})
          </button>
          <button
            onClick={() => setActiveTab('courses')}
            className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all ${
              activeTab === 'courses'
                ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                : 'bg-zinc-900 text-zinc-400 hover:text-white'
            }`}
          >
            Catálogo de Cursos STPS ({courses.length})
          </button>
        </div>

        {/* Table Content */}
        {activeTab === 'certificates' ? (
          <div className="rounded-2xl border border-zinc-800/80 bg-zinc-900/40 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="border-b border-zinc-800/80 bg-zinc-900/80 text-zinc-400 font-medium uppercase tracking-wider">
                  <tr>
                    <th className="py-3 px-4">Folio STPS</th>
                    <th className="py-3 px-4">Trabajador & CURP</th>
                    <th className="py-3 px-4">Empresa / Razón Social</th>
                    <th className="py-3 px-4">Instructor Registrado</th>
                    <th className="py-3 px-4">Fecha Emisión</th>
                    <th className="py-3 px-4">Vigencia</th>
                    <th className="py-3 px-4">Calificación</th>
                    <th className="py-3 px-4">Estatus</th>
                    <th className="py-3 px-4 text-right">Formato Oficial</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-zinc-800/60 font-mono text-zinc-300">
                  {certificates.length === 0 ? (
                    <tr>
                      <td colSpan={9} className="py-8 text-center text-zinc-500 font-sans">
                        No hay constancias DC-3 emitidas registradas.
                      </td>
                    </tr>
                  ) : (
                    certificates.map((cert) => (
                      <tr key={cert.id} className="hover:bg-zinc-800/30 transition-colors">
                        <td className="py-3.5 px-4 font-bold text-indigo-400">{cert.folio_number}</td>
                        <td className="py-3.5 px-4">
                          <div className="font-sans font-semibold text-white">{cert.worker_full_name}</div>
                          <div className="text-[10px] text-zinc-500">{cert.worker_curp}</div>
                        </td>
                        <td className="py-3.5 px-4 font-sans text-zinc-300">{cert.company_name}</td>
                        <td className="py-3.5 px-4 font-sans text-zinc-400">{cert.instructor_name}</td>
                        <td className="py-3.5 px-4 text-zinc-400">
                          {new Date(cert.issue_date).toLocaleDateString('es-MX')}
                        </td>
                        <td className="py-3.5 px-4 text-zinc-400">
                          {new Date(cert.expiration_date).toLocaleDateString('es-MX')}
                        </td>
                        <td className="py-3.5 px-4">
                          <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-bold">
                            {cert.score_percentage}%
                          </span>
                        </td>
                        <td className="py-3.5 px-4">
                          <span className="px-2.5 py-1 rounded-full text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                            VIGENTE STPS
                          </span>
                        </td>
                        <td className="py-3.5 px-4 text-right">
                          <Link
                            href={`/capacitacion/dc3/${cert.id}/imprimir`}
                            target="_blank"
                            className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg bg-zinc-800 text-indigo-400 hover:text-white hover:bg-indigo-600 text-xs font-semibold border border-zinc-700 transition-colors"
                          >
                            <Printer className="w-3.5 h-3.5" />
                            <span>Imprimir DC-3</span>
                          </Link>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {courses.length === 0 ? (
              <div className="col-span-full py-8 text-center text-zinc-500">
                No hay cursos de capacitación registrados en el catálogo.
              </div>
            ) : (
              courses.map((course) => (
                <div
                  key={course.id}
                  className="rounded-2xl border border-zinc-800/80 bg-zinc-900/40 p-5 hover:border-zinc-700 transition-all"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-[10px] font-mono font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                      {course.code}
                    </span>
                    <span className="text-[10px] font-mono text-zinc-400">{course.duration_hours} Horas</span>
                  </div>
                  <h3 className="font-bold text-sm text-white mb-1.5">{course.title}</h3>
                  <p className="text-xs text-zinc-400 line-clamp-2 mb-4">{course.description || 'Sin descripción'}</p>
                  <div className="flex items-center justify-between text-[11px] pt-3 border-t border-zinc-800/80 text-zinc-500">
                    <span>Norma: {course.normative_reference || 'STPS'}</span>
                    <span className="text-emerald-400">Vigencia {course.validity_years} año(s)</span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </main>
    </div>
  );
}

export default function CapacitacionPage() {
  return (
    <ProtectedRoute>
      <CapacitacionDashboardContent />
    </ProtectedRoute>
  );
}
