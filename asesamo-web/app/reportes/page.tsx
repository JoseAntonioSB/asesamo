'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import {
  ReportMetricsResponse,
  AccessMetricsSummary,
  DailyAccessDistribution,
  RejectionReasonSummary,
  ContractorComplianceSummary,
  AccessLogItem,
} from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  BarChart3,
  TrendingUp,
  Download,
  Calendar,
  Building2,
  FileCheck2,
  ShieldCheck,
  ShieldAlert,
  Clock,
  Users,
  LogOut,
  ArrowLeft,
  FileText,
  FileStack,
  QrCode,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  FileSpreadsheet,
  DoorOpen,
  History,
  Copy,
  Check,
} from 'lucide-react';

const DEFAULT_PLANT_ID = '22222222-2222-2222-2222-222222222222'; // Planta Querétaro Norte

type DatePreset = 'today' | '7d' | '30d' | 'month';

function ReportesContent() {
  const { user, logout } = useAuth();
  const isGuard = user?.role === 'guard';

  const [datePreset, setDatePreset] = useState<DatePreset>('30d');
  const [fromDate, setFromDate] = useState<string>('');
  const [toDate, setToDate] = useState<string>('');
  const [metricsData, setMetricsData] = useState<ReportMetricsResponse | null>(null);
  const [accessLogs, setAccessLogs] = useState<AccessLogItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [isDownloadingAccess, setIsDownloadingAccess] = useState<boolean>(false);
  const [isDownloadingCompliance, setIsDownloadingCompliance] = useState<boolean>(false);
  const [logResultFilter, setLogResultFilter] = useState<string>('all');
  const [copiedLogId, setCopiedLogId] = useState<string | null>(null);

  // Initialize dates based on preset
  const applyPreset = useCallback((preset: DatePreset) => {
    setDatePreset(preset);
    const now = new Date();
    const toStr = now.toISOString().split('T')[0];
    let from = new Date();

    switch (preset) {
      case 'today':
        from = now;
        break;
      case '7d':
        from.setDate(now.getDate() - 7);
        break;
      case '30d':
        from.setDate(now.getDate() - 30);
        break;
      case 'month':
        from = new Date(now.getFullYear(), now.getMonth(), 1);
        break;
    }

    const fromStr = from.toISOString().split('T')[0];
    setFromDate(fromStr);
    setToDate(toStr);
  }, []);

  useEffect(() => {
    applyPreset('30d');
  }, [applyPreset]);

  // Load metrics & live access logs from API
  const loadMetrics = useCallback(
    async (showToast = false) => {
      try {
        if (showToast) setIsRefreshing(true);
        else setIsLoading(true);

        const [metrics, logsRes] = await Promise.allSettled([
          apiClient.getReportMetrics({
            plant_id: DEFAULT_PLANT_ID,
            from: fromDate || undefined,
            to: toDate || undefined,
          }),
          apiClient.getAccessLogs({
            plant_id: DEFAULT_PLANT_ID,
            from: fromDate || undefined,
            to: toDate || undefined,
            per_page: 50,
          }),
        ]);

        if (metrics.status === 'fulfilled') {
          setMetricsData(metrics.value);
        } else {
          toast.error('Error al consultar métricas', {
            description: 'No se pudieron recuperar los indicadores de acceso y auditoría.',
          });
        }

        if (logsRes.status === 'fulfilled') {
          setAccessLogs(logsRes.value.logs || []);
        } else {
          setAccessLogs([]);
        }

        if (showToast && metrics.status === 'fulfilled') {
          toast.success('Métricas y bitácora de accesos actualizadas');
        }
      } catch {
        toast.error('Error al actualizar datos');
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
      }
    },
    [fromDate, toDate]
  );

  useEffect(() => {
    if (fromDate && toDate) {
      loadMetrics();
    }
  }, [fromDate, toDate, loadMetrics]);

  const copyToClipboard = (text: string, id: string) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedLogId(id);
    setTimeout(() => setCopiedLogId(null), 2000);
  };

  // Download Access Logs CSV
  const handleDownloadAccessLogs = async () => {
    try {
      setIsDownloadingAccess(true);
      const blob = await apiClient.downloadAccessLogsCSV({
        plant_id: DEFAULT_PLANT_ID,
        from: fromDate || undefined,
        to: toDate || undefined,
      });

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `reporte_accesos_asesamo_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success('Bitácora de accesos descargada', {
        description: 'Archivo CSV generado con codificación UTF-8 compatible con Microsoft Excel.',
      });
    } catch {
      toast.error('Error al exportar reporte de accesos');
    } finally {
      setIsDownloadingAccess(false);
    }
  };

  // Download Compliance Matrix CSV
  const handleDownloadCompliance = async () => {
    try {
      setIsDownloadingCompliance(true);
      const blob = await apiClient.downloadComplianceCSV(DEFAULT_PLANT_ID);

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `matriz_cumplimiento_stps_${new Date().toISOString().split('T')[0]}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success('Matriz STPS descargada', {
        description: 'Archivo CSV con dictamen general (DC-3, SUA, IMSS, Médico) exportado exitosamente.',
      });
    } catch {
      toast.error('Error al exportar matriz de cumplimiento');
    } finally {
      setIsDownloadingCompliance(false);
    }
  };

  const summary = metricsData?.summary || {
    total_accesses: 0,
    granted_count: 0,
    warning_count: 0,
    denied_count: 0,
    compliance_rate: 0,
    avg_duration_minutes: 0,
  };

  const timeline = metricsData?.timeline || [];
  const reasons = metricsData?.rejection_reasons || [];
  const ranking = metricsData?.contractor_ranking || [];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-emerald-500 selection:text-white">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <BarChart3 className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 font-mono border border-indigo-500/30">
                  Auditoría & Métricas
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Reportes STPS, Índices de Cumplimiento y Control de Incidencias
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            {isGuard ? (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                >
                  <ArrowLeft className="w-3.5 h-3.5" />
                  <span className="hidden sm:inline">Volver a Caseta</span>
                </Button>
              </Link>
            ) : (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                  title="Abrir terminal de Caseta de Vigilancia"
                >
                  <DoorOpen className="w-3.5 h-3.5 text-emerald-400" />
                  <span className="hidden md:inline">Caseta</span>
                </Button>
              </Link>
            )}

            <Link href="/contratistas">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-blue-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Padrón de Contratistas y Cuadrillas"
              >
                <Building2 className="w-3.5 h-3.5 text-blue-400" />
                <span className="hidden sm:inline">Contratistas</span>
              </Button>
            </Link>

            <Link href="/documentos">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-emerald-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Expediente de Documentos"
              >
                <FileText className="w-3.5 h-3.5 text-emerald-400" />
                <span className="hidden sm:inline">Documentos</span>
              </Button>
            </Link>

            <Link href="/ehs/aprobaciones">
              <Button
                size="sm"
                variant="outline"
                className="border-amber-800/60 bg-amber-950/30 text-amber-300 hover:text-white hover:bg-amber-900/50 text-xs gap-1.5"
                title="Bandeja EHS"
              >
                <FileStack className="w-3.5 h-3.5 text-amber-400" />
                <span className="hidden md:inline">Bandeja EHS</span>
              </Button>
            </Link>

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

            {/* User Profile Pill */}
            <Link
              href="/perfil"
              className="flex items-center gap-2 px-2.5 py-1 rounded-xl border border-slate-800 bg-slate-900 hover:bg-slate-800 hover:border-slate-700 transition-colors"
              title="Ver Mi Perfil"
            >
              <div className="h-6 w-6 rounded-lg bg-slate-800 text-slate-200 font-bold text-[10px] flex items-center justify-center border border-slate-700 overflow-hidden relative">
                {user?.id && (
                  <img
                    src={`/api/v1/users/${user.id}/avatar`}
                    alt={user.full_name || 'Usuario'}
                    className="w-full h-full object-cover absolute inset-0"
                    onError={(e) => {
                      (e.target as HTMLElement).style.display = 'none';
                    }}
                  />
                )}
                <span>{(user?.full_name || 'US').slice(0, 2).toUpperCase()}</span>
              </div>
              <span className="text-xs font-semibold text-slate-200 hidden md:inline max-w-[120px] truncate">
                {user?.full_name || 'Perfil'}
              </span>
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
        {/* Title and Audit Export Actions Area */}
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 pb-2 border-b border-slate-800/80">
          <div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white flex items-center gap-2">
              Tablero de Cumplimiento y Auditoría Laboral
            </h1>
            <p className="text-xs text-slate-400 mt-1">
              Monitoreo analítico de accesos en caseta, validación normativa STPS (DC-3, REPSE, SUA) e historial inmutable.
            </p>
          </div>

          {/* Export Action Buttons */}
          <div className="flex items-center gap-2 flex-wrap">
            <Button
              size="sm"
              variant="outline"
              disabled={isDownloadingAccess}
              onClick={handleDownloadAccessLogs}
              className="border-slate-700 bg-slate-900 text-slate-200 hover:text-white hover:bg-slate-800 text-xs gap-1.5 shadow-sm"
              title="Descargar bitácora completa de ingresos y salidas con sellos temporales y dictamen"
            >
              <Download className={`w-3.5 h-3.5 text-indigo-400 ${isDownloadingAccess ? 'animate-bounce' : ''}`} />
              <span>Bitácora de Accesos (CSV)</span>
            </Button>

            <Button
              size="sm"
              disabled={isDownloadingCompliance}
              onClick={handleDownloadCompliance}
              className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs gap-1.5 shadow-md shadow-emerald-900/30"
              title="Descargar matriz de estatus documental de cuadrillas según lineamientos de inspección STPS"
            >
              <FileSpreadsheet className={`w-3.5 h-3.5 ${isDownloadingCompliance ? 'animate-bounce' : ''}`} />
              <span>Matriz STPS (CSV)</span>
            </Button>

            <Button
              size="sm"
              variant="outline"
              onClick={() => loadMetrics(true)}
              disabled={isLoading || isRefreshing}
              className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs px-2"
              title="Refrescar métricas"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isRefreshing ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* Date Filter & Preset Bar */}
        <div className="p-3 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur-sm flex flex-col md:flex-row md:items-center justify-between gap-3">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className="text-xs text-slate-400 font-medium mr-1 flex items-center gap-1">
              <Calendar className="w-3.5 h-3.5 text-slate-500" />
              Periodo:
            </span>
            <Button
              size="sm"
              variant={datePreset === 'today' ? 'default' : 'outline'}
              onClick={() => applyPreset('today')}
              className={`text-xs h-7 px-2.5 ${datePreset === 'today' ? 'bg-indigo-600 text-white' : 'border-slate-700 text-slate-300'}`}
            >
              Hoy
            </Button>
            <Button
              size="sm"
              variant={datePreset === '7d' ? 'default' : 'outline'}
              onClick={() => applyPreset('7d')}
              className={`text-xs h-7 px-2.5 ${datePreset === '7d' ? 'bg-indigo-600 text-white' : 'border-slate-700 text-slate-300'}`}
            >
              Últimos 7 días
            </Button>
            <Button
              size="sm"
              variant={datePreset === '30d' ? 'default' : 'outline'}
              onClick={() => applyPreset('30d')}
              className={`text-xs h-7 px-2.5 ${datePreset === '30d' ? 'bg-indigo-600 text-white' : 'border-slate-700 text-slate-300'}`}
            >
              Últimos 30 días
            </Button>
            <Button
              size="sm"
              variant={datePreset === 'month' ? 'default' : 'outline'}
              onClick={() => applyPreset('month')}
              className={`text-xs h-7 px-2.5 ${datePreset === 'month' ? 'bg-indigo-600 text-white' : 'border-slate-700 text-slate-300'}`}
            >
              Mes en Curso
            </Button>
          </div>

          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1 text-xs">
              <span className="text-slate-400 text-[11px]">Desde:</span>
              <Input
                type="date"
                value={fromDate}
                onChange={(e) => {
                  setFromDate(e.target.value);
                  setDatePreset('30d');
                }}
                className="h-7 w-32 bg-slate-900 border-slate-700 text-slate-200 text-xs px-2"
              />
            </div>
            <div className="flex items-center gap-1 text-xs">
              <span className="text-slate-400 text-[11px]">Hasta:</span>
              <Input
                type="date"
                value={toDate}
                onChange={(e) => {
                  setToDate(e.target.value);
                  setDatePreset('30d');
                }}
                className="h-7 w-32 bg-slate-900 border-slate-700 text-slate-200 text-xs px-2"
              />
            </div>
          </div>
        </div>

        {/* KPI Cards Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Total Accesses */}
          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/50 flex items-center justify-between">
            <div>
              <p className="text-xs text-slate-400 font-medium">Accesos Totales</p>
              <p className="text-2xl font-bold text-white mt-1">
                {isLoading ? '...' : summary.total_accesses}
              </p>
              <p className="text-[11px] text-slate-500 mt-0.5">
                {summary.granted_count + summary.warning_count} permitidos / {summary.denied_count} denegados
              </p>
            </div>
            <div className="p-3 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <TrendingUp className="w-6 h-6" />
            </div>
          </div>

          {/* Compliance Rate */}
          <div className="p-4 rounded-xl border border-emerald-900/40 bg-emerald-950/20 flex items-center justify-between">
            <div>
              <p className="text-xs text-emerald-300 font-medium">Tasa de Cumplimiento</p>
              <p className="text-2xl font-bold text-emerald-400 mt-1">
                {isLoading ? '...' : `${summary.compliance_rate}%`}
              </p>
              <p className="text-[11px] text-emerald-500/80 mt-0.5">Efectividad de acceso en caseta</p>
            </div>
            <div className="p-3 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <CheckCircle2 className="w-6 h-6" />
            </div>
          </div>

          {/* Denied Incidents */}
          <div className="p-4 rounded-xl border border-red-900/40 bg-red-950/20 flex items-center justify-between">
            <div>
              <p className="text-xs text-red-300 font-medium">Accesos Denegados</p>
              <p className="text-2xl font-bold text-red-400 mt-1">
                {isLoading ? '...' : summary.denied_count}
              </p>
              <p className="text-[11px] text-red-500/80 mt-0.5">
                {summary.total_accesses > 0
                  ? `${((summary.denied_count / summary.total_accesses) * 100).toFixed(1)}% del total`
                  : 'Sin incidencias'}
              </p>
            </div>
            <div className="p-3 rounded-lg bg-red-500/10 text-red-400 border border-red-500/20">
              <XCircle className="w-6 h-6" />
            </div>
          </div>

          {/* Average Duration */}
          <div className="p-4 rounded-xl border border-slate-800 bg-slate-900/50 flex items-center justify-between">
            <div>
              <p className="text-xs text-slate-400 font-medium">Estancia Promedio</p>
              <p className="text-2xl font-bold text-slate-200 mt-1">
                {isLoading
                  ? '...'
                  : summary.avg_duration_minutes > 60
                  ? `${(summary.avg_duration_minutes / 60).toFixed(1)} h`
                  : `${Math.round(summary.avg_duration_minutes)} min`}
              </p>
              <p className="text-[11px] text-slate-500 mt-0.5">Tiempo dentro de instalaciones</p>
            </div>
            <div className="p-3 rounded-lg bg-slate-800 text-slate-400 border border-slate-700">
              <Clock className="w-6 h-6" />
            </div>
          </div>
        </div>

        {/* Section 2: Timeline Distribution & Top Rejections */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Daily Timeline (2 cols) */}
          <div className="lg:col-span-2 p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur-sm space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                  <BarChart3 className="w-4 h-4 text-indigo-400" />
                  <span>Distribución Cronológica de Accesos</span>
                </h3>
                <p className="text-[11px] text-slate-400">
                  Volumen diario de entradas aprobadas frente a denegaciones registradas.
                </p>
              </div>
              <div className="flex items-center gap-3 text-[11px]">
                <span className="flex items-center gap-1.5 text-emerald-400">
                  <span className="w-2.5 h-2.5 rounded-sm bg-emerald-500" />
                  Permitidos
                </span>
                <span className="flex items-center gap-1.5 text-rose-400">
                  <span className="w-2.5 h-2.5 rounded-sm bg-rose-500" />
                  Denegados
                </span>
              </div>
            </div>

            {isLoading ? (
              <div className="h-48 flex items-center justify-center text-xs text-slate-400">
                Cargando serie cronológica...
              </div>
            ) : timeline.length === 0 ? (
              <div className="h-48 flex flex-col items-center justify-center text-slate-500 text-xs">
                <Calendar className="w-8 h-8 mb-1.5 text-slate-600" />
                <p>No se encontraron registros de acceso en el periodo seleccionado.</p>
              </div>
            ) : (
              <div className="space-y-2 pt-2">
                {timeline.map((item) => {
                  const maxVal = Math.max(...timeline.map((t) => t.total), 1);
                  const grantedPct = Math.round((item.granted / maxVal) * 100);
                  const deniedPct = Math.round((item.denied / maxVal) * 100);

                  return (
                    <div key={item.date} className="space-y-1">
                      <div className="flex justify-between text-[11px] font-mono text-slate-300">
                        <span>{item.date}</span>
                        <span>
                          <strong className="text-white">{item.total}</strong> accesos (
                          <span className="text-emerald-400">{item.granted} ok</span>,{' '}
                          <span className="text-rose-400">{item.denied} no</span>)
                        </span>
                      </div>
                      <div className="h-2.5 w-full bg-slate-800 rounded-full overflow-hidden flex">
                        <div
                          style={{ width: `${grantedPct}%` }}
                          className="bg-emerald-500 h-full transition-all"
                          title={`${item.granted} permitidos`}
                        />
                        <div
                          style={{ width: `${deniedPct}%` }}
                          className="bg-rose-500 h-full transition-all"
                          title={`${item.denied} denegados`}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Top Rejection Reasons (1 col) */}
          <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur-sm space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <ShieldAlert className="w-4 h-4 text-rose-400" />
                <span>Principales Causas de Rechazo</span>
              </h3>
              <p className="text-[11px] text-slate-400">
                Motivos más comunes de denegación en torniquete de acceso.
              </p>
            </div>

            {isLoading ? (
              <div className="h-48 flex items-center justify-center text-xs text-slate-400">
                Cargando causas...
              </div>
            ) : reasons.length === 0 ? (
              <div className="h-48 flex flex-col items-center justify-center text-emerald-400/80 text-xs">
                <CheckCircle2 className="w-8 h-8 mb-1.5 text-emerald-500" />
                <p className="text-white font-medium">100% de Cumplimiento</p>
                <p className="text-slate-400 text-[11px]">No hay denegaciones en el rango elegido.</p>
              </div>
            ) : (
              <div className="space-y-3 pt-1">
                {reasons.map((r, i) => (
                  <div key={r.reason} className="space-y-1">
                    <div className="flex justify-between items-center text-xs">
                      <span className="text-slate-200 truncate pr-2 max-w-[200px]" title={r.reason}>
                        <span className="font-mono text-slate-500 mr-1.5">#{i + 1}</span>
                        {r.reason}
                      </span>
                      <span className="font-mono font-semibold text-rose-400 whitespace-nowrap text-[11px]">
                        {r.count} ({r.percentage}%)
                      </span>
                    </div>
                    <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
                      <div
                        style={{ width: `${r.percentage}%` }}
                        className="bg-rose-500 h-full rounded-full"
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Section 3: Contractor Compliance Ranking Table */}
        <div className="rounded-xl border border-slate-800 bg-slate-900/60 overflow-hidden shadow-lg backdrop-blur-sm space-y-3 p-5">
          <div>
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <Building2 className="w-4 h-4 text-cyan-400" />
              <span>Índice de Cumplimiento por Empresa Contratista</span>
            </h3>
            <p className="text-[11px] text-slate-400">
              Desempeño y porcentaje de efectividad de ingreso de cada proveedor industrial.
            </p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-300">
              <thead className="bg-slate-950/80 text-[11px] font-semibold uppercase tracking-wider text-slate-400 border-b border-slate-800">
                <tr>
                  <th scope="col" className="px-4 py-3">Empresa Contratista</th>
                  <th scope="col" className="px-4 py-3 text-center">Intentos de Acceso</th>
                  <th scope="col" className="px-4 py-3 text-center">Permitidos</th>
                  <th scope="col" className="px-4 py-3 text-center">Denegados</th>
                  <th scope="col" className="px-4 py-3 text-right">Efectividad / Cumplimiento</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/80">
                {isLoading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-slate-400">
                      Cargando ranking de contratistas...
                    </td>
                  </tr>
                ) : ranking.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-slate-500">
                      No hay registros de contratistas en este rango.
                    </td>
                  </tr>
                ) : (
                  ranking.map((c) => {
                    const isHigh = c.compliance_rate >= 90;
                    const isMedium = c.compliance_rate >= 75 && c.compliance_rate < 90;

                    return (
                      <tr key={c.contractor_id} className="hover:bg-slate-800/40 transition-colors">
                        <td className="px-4 py-3 font-medium text-white">
                          <div className="flex items-center gap-2">
                            <Building2 className="w-3.5 h-3.5 text-slate-400" />
                            <span>{c.contractor_name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-center font-mono font-semibold">
                          {c.total_attempts}
                        </td>
                        <td className="px-4 py-3 text-center font-mono text-emerald-400">
                          {c.granted_count}
                        </td>
                        <td className="px-4 py-3 text-center font-mono text-rose-400">
                          {c.denied_count}
                        </td>
                        <td className="px-4 py-3 text-right font-mono">
                          <span
                            className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full font-bold text-xs ${
                              isHigh
                                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                : isMedium
                                ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                                : 'bg-red-500/10 text-red-400 border border-red-500/20'
                            }`}
                          >
                            {c.compliance_rate}%
                          </span>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Section 4: Live Access Logs Table (Bitácora de Accesos en Vivo) */}
        <div className="rounded-xl border border-slate-800 bg-slate-900/60 overflow-hidden shadow-lg backdrop-blur-sm space-y-4 p-5">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h3 className="text-sm font-semibold text-white flex items-center gap-2">
                <History className="w-4 h-4 text-emerald-400" />
                <span>Bitácora de Accesos en Vivo (Auditoría NOM-STPS)</span>
                <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700">
                  {accessLogs.length} eventos
                </span>
              </h3>
              <p className="text-[11px] text-slate-400">
                Registro inmutable de entradas y salidas procesadas en torniquetes y puntos de acceso industrial.
              </p>
            </div>

            {/* Filter pills */}
            <div className="flex items-center gap-1.5 overflow-x-auto text-xs">
              <button
                type="button"
                onClick={() => setLogResultFilter('all')}
                className={`px-2.5 py-1 rounded-lg font-medium transition-colors cursor-pointer text-xs ${
                  logResultFilter === 'all'
                    ? 'bg-slate-800 text-white border border-slate-700'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
                }`}
              >
                Todos ({accessLogs.length})
              </button>
              <button
                type="button"
                onClick={() => setLogResultFilter('granted')}
                className={`px-2.5 py-1 rounded-lg font-medium transition-colors cursor-pointer text-xs flex items-center gap-1 ${
                  logResultFilter === 'granted'
                    ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                    : 'text-slate-400 hover:text-emerald-400 hover:bg-slate-800/60'
                }`}
              >
                <CheckCircle2 className="w-3 h-3 text-emerald-400" />
                Permitidos ({accessLogs.filter((l) => l.result === 'granted').length})
              </button>
              <button
                type="button"
                onClick={() => setLogResultFilter('granted_with_warning')}
                className={`px-2.5 py-1 rounded-lg font-medium transition-colors cursor-pointer text-xs flex items-center gap-1 ${
                  logResultFilter === 'granted_with_warning'
                    ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                    : 'text-slate-400 hover:text-amber-400 hover:bg-slate-800/60'
                }`}
              >
                <AlertTriangle className="w-3 h-3 text-amber-400" />
                Advertencias ({accessLogs.filter((l) => l.result === 'granted_with_warning').length})
              </button>
              <button
                type="button"
                onClick={() => setLogResultFilter('denied')}
                className={`px-2.5 py-1 rounded-lg font-medium transition-colors cursor-pointer text-xs flex items-center gap-1 ${
                  logResultFilter === 'denied'
                    ? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
                    : 'text-slate-400 hover:text-rose-400 hover:bg-slate-800/60'
                }`}
              >
                <XCircle className="w-3 h-3 text-rose-400" />
                Denegados ({accessLogs.filter((l) => l.result === 'denied').length})
              </button>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-300">
              <thead className="bg-slate-950/80 text-[11px] font-semibold uppercase tracking-wider text-slate-400 border-b border-slate-800">
                <tr>
                  <th scope="col" className="px-4 py-3">Fecha y Hora</th>
                  <th scope="col" className="px-4 py-3">ID Trabajador (UUID)</th>
                  <th scope="col" className="px-4 py-3">Punto de Acceso</th>
                  <th scope="col" className="px-4 py-3 text-center">Dictamen</th>
                  <th scope="col" className="px-4 py-3 text-center">Permanencia</th>
                  <th scope="col" className="px-4 py-3">Motivo / Observaciones</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/80">
                {isLoading ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                      <div className="flex items-center justify-center gap-2">
                        <div className="w-4 h-4 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
                        <span>Cargando bitácora de accesos...</span>
                      </div>
                    </td>
                  </tr>
                ) : accessLogs.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="px-4 py-8 text-center text-slate-500">
                      No hay registros de acceso para el rango seleccionado.
                    </td>
                  </tr>
                ) : (
                  accessLogs
                    .filter((log) => logResultFilter === 'all' || log.result === logResultFilter)
                    .map((log) => {
                      const isGranted = log.result === 'granted';
                      const isWarning = log.result === 'granted_with_warning';
                      const isDenied = log.result === 'denied';

                      const dateStr = log.checked_in_at
                        ? new Date(log.checked_in_at).toLocaleString('es-MX', {
                            dateStyle: 'short',
                            timeStyle: 'medium',
                          })
                        : '—';

                      return (
                        <tr key={log.id} className="hover:bg-slate-800/40 transition-colors">
                          {/* Timestamp */}
                          <td className="px-4 py-3 font-mono text-[11px] text-slate-300 whitespace-nowrap">
                            <div className="flex items-center gap-1.5">
                              <Clock className="w-3.5 h-3.5 text-slate-400" />
                              <span>{dateStr}</span>
                            </div>
                          </td>

                          {/* Worker User ID */}
                          <td className="px-4 py-3 font-mono text-[11px]">
                            <div className="flex items-center gap-1.5">
                              <span className="text-slate-200">
                                ...{log.user_id ? log.user_id.slice(-12) : '—'}
                              </span>
                              <button
                                type="button"
                                onClick={() => copyToClipboard(log.user_id, log.id)}
                                className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
                                title="Copiar UUID"
                              >
                                {copiedLogId === log.id ? (
                                  <Check className="w-3 h-3 text-emerald-400" />
                                ) : (
                                  <Copy className="w-3 h-3" />
                                )}
                              </button>
                            </div>
                          </td>

                          {/* Access Point */}
                          <td className="px-4 py-3 font-mono text-[11px] text-slate-400 whitespace-nowrap">
                            <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-slate-800/60 border border-slate-700/60 text-slate-300">
                              <DoorOpen className="w-3 h-3 text-slate-400" />
                              {log.access_point_id ? `...${log.access_point_id.slice(-8)}` : 'Torniquete'}
                            </span>
                          </td>

                          {/* Result */}
                          <td className="px-4 py-3 text-center whitespace-nowrap">
                            {isGranted && (
                              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                                <CheckCircle2 className="w-3 h-3" />
                                Permitido
                              </span>
                            )}
                            {isWarning && (
                              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                                <AlertTriangle className="w-3 h-3" />
                                Advertencia
                              </span>
                            )}
                            {isDenied && (
                              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">
                                <XCircle className="w-3 h-3" />
                                Denegado
                              </span>
                            )}
                          </td>

                          {/* Permanence / Check-out */}
                          <td className="px-4 py-3 text-center whitespace-nowrap font-mono text-[11px]">
                            {log.checked_out_at ? (
                              <span className="inline-flex items-center gap-1 text-slate-400">
                                Salida: {new Date(log.checked_out_at).toLocaleTimeString('es-MX', { hour: '2-digit', minute: '2-digit' })}
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20 text-[10px] font-semibold">
                                <span className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse" />
                                En Planta
                              </span>
                            )}
                          </td>

                          {/* Denial Reason / Notes */}
                          <td className="px-4 py-3 text-xs">
                            {log.denial_reason ? (
                              <span className="text-rose-400 font-medium">
                                {log.denial_reason}
                              </span>
                            ) : log.notes ? (
                              <span className="text-slate-300 italic">{log.notes}</span>
                            ) : (
                              <span className="text-emerald-400/90 flex items-center gap-1">
                                <ShieldCheck className="w-3 h-3" />
                                Documentación 100% conforme
                              </span>
                            )}
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

export default function ReportesPage() {
  return (
    <ProtectedRoute>
      <ReportesContent />
    </ProtectedRoute>
  );
}
