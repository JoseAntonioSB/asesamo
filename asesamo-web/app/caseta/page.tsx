// asesamo-web/app/caseta/page.tsx
'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  Shield,
  Building,
  LogOut,
  Volume2,
  VolumeX,
  ScanBarcode,
  AlertCircle,
  DoorOpen,
  FileText,
  FileStack,
  Printer,
  QrCode,
} from 'lucide-react';

import { StatusBanner } from '@/components/caseta/status-banner';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { WorkerCard } from '@/components/caseta/worker-card';
import { DocumentTable } from '@/components/caseta/document-table';
import { ScannerInput } from '@/components/caseta/scanner-input';
import { QuickFillPresets } from '@/components/caseta/quick-fill-presets';
import { casetaAudio } from '@/lib/utils/audio';
import { apiClient, ApiError } from '@/lib/api/client';
import { useAuth } from '@/contexts/auth-context';
import {
  CheckInOutput,
  CheckOutOutput,
  AccessPointResponse,
} from '@/lib/api/types';

// Fallback seed access points if API is still booting or unseeded
const DEFAULT_ACCESS_POINTS: AccessPointResponse[] = [
  {
    id: '33333333-3333-3333-3333-333333333331',
    plant_id: '22222222-2222-2222-2222-222222222222',
    tenant_id: '11111111-1111-1111-1111-111111111111',
    name: 'Caseta Principal Norte',
    description: 'Caseta de vigilancia y torniquete de acceso peatonal principal',
    is_active: true,
  },
];

export default function CasetaPage() {
  const router = useRouter();
  const { user, logout, isAuthenticated, isLoading: authLoading } = useAuth();

  // Access point & plant state
  const [accessPoints, setAccessPoints] = useState<AccessPointResponse[]>(DEFAULT_ACCESS_POINTS);
  const [selectedApId, setSelectedApId] = useState<string>(DEFAULT_ACCESS_POINTS[0].id);

  // Check-In vs Check-Out mode
  const [mode, setMode] = useState<'check_in' | 'check_out'>('check_in');

  // Operation state
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [checkInResult, setCheckInResult] = useState<CheckInOutput | null>(null);
  const [checkOutResult, setCheckOutResult] = useState<CheckOutOutput | null>(null);

  // Live Digital Clock
  const [currentTime, setCurrentTime] = useState<string>('');
  const [currentDate, setCurrentDate] = useState<string>('');

  // Audio mute toggle
  const [isMuted, setIsMuted] = useState<boolean>(false);

  // Guard info
  const guardName = user?.full_name || 'Oficial Roberto Caseta';
  const isEhsOrAdmin =
    user?.role === 'ehs_manager' ||
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin';

  // Guard auth check
  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [authLoading, isAuthenticated, router]);

  // Clock interval
  useEffect(() => {
    const updateClock = () => {
      const now = new Date();
      setCurrentTime(
        now.toLocaleTimeString('es-MX', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        })
      );
      setCurrentDate(
        now.toLocaleDateString('es-MX', {
          weekday: 'long',
          year: 'numeric',
          month: 'long',
          day: 'numeric',
        })
      );
    };

    updateClock();
    const timer = setInterval(updateClock, 1000);
    return () => clearInterval(timer);
  }, []);

  // Sync mute state
  useEffect(() => {
    setIsMuted(casetaAudio.getMuted());
  }, []);

  const toggleMute = () => {
    const next = !isMuted;
    casetaAudio.setMuted(next);
    setIsMuted(next);
  };

  // Fetch access points catalog on mount
  useEffect(() => {
    const fetchAccessPoints = async () => {
      try {
        const apList = await apiClient.getAccessPoints();
        if (Array.isArray(apList) && apList.length > 0) {
          setAccessPoints(apList);
          setSelectedApId(apList[0].id);
        }
      } catch {
        // Fallback to default caseta
      }
    };

    fetchAccessPoints();
  }, []);

  // Handle Logout
  const handleLogout = () => {
    logout();
  };

  // Handle Scan Verification
  const handleScanWorker = useCallback(
    async (workerId: string) => {
      setIsLoading(true);
      setError(null);
      setCheckOutResult(null);

      try {
        if (mode === 'check_in') {
          const data = await apiClient.checkIn(workerId, selectedApId);
          setCheckInResult(data);

          // Trigger synthesized sound based on compliance outcome
          if (data.result === 'granted') {
            casetaAudio.playGranted();
          } else if (data.result === 'granted_with_warning') {
            casetaAudio.playWarning();
          } else {
            casetaAudio.playDenied();
          }
        } else {
          // Check-Out mode
          const data = await apiClient.checkOut(workerId, selectedApId);
          setCheckOutResult(data);
          casetaAudio.playGranted();
        }
      } catch (err: unknown) {
        if (err instanceof ApiError) {
          setError(err.message);
        } else if (err instanceof Error) {
          setError(`Fallo en el servicio: ${err.message}`);
        } else {
          setError('Error al procesar la solicitud en caseta.');
        }
        casetaAudio.playDenied();
      } finally {
        setIsLoading(false);
      }
    },
    [mode, selectedApId]
  );

  const currentAp = accessPoints.find((a) => a.id === selectedApId) || accessPoints[0];

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col antialiased">
      {/* 1. Header with Plant, Caseta selector, Clock, Guard Profile & Logout */}
      <header className="sticky top-0 z-30 border-b border-slate-800 bg-slate-950/90 backdrop-blur px-4 sm:px-6 py-3.5 shadow-md">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
          {/* Brand & Access Point */}
          <div className="flex items-center gap-3 w-full md:w-auto justify-between md:justify-start">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-600 text-white font-black shadow-lg shadow-emerald-600/30">
                <Shield className="h-6 w-6" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-base font-black tracking-tight text-white font-mono">ASESAMO</span>
                  <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                    EN LÍNEA
                  </span>
                </div>
                <div className="text-xs text-slate-400 flex items-center gap-1.5">
                  <Building className="h-3 w-3 text-slate-500" />
                  <span>Planta Querétaro Norte</span>
                </div>
              </div>
            </div>

            {/* Access Point Dropdown if multiple exist */}
            <div className="ml-0 md:ml-4">
              {accessPoints.length > 1 ? (
                <select
                  value={selectedApId}
                  onChange={(e) => setSelectedApId(e.target.value)}
                  className="rounded-lg border border-slate-700 bg-slate-900 px-3 py-1.5 text-xs font-semibold text-slate-200 outline-none focus:border-emerald-500 cursor-pointer"
                >
                  {accessPoints.map((ap) => (
                    <option key={ap.id} value={ap.id}>
                      {ap.name}
                    </option>
                  ))}
                </select>
              ) : (
                <span className="inline-flex items-center gap-1 px-3 py-1 rounded-lg text-xs font-bold bg-slate-900 text-cyan-300 border border-slate-800">
                  <DoorOpen className="h-3.5 w-3.5 text-cyan-400" />
                  {currentAp?.name || 'Caseta Principal Norte'}
                </span>
              )}
            </div>
          </div>

          {/* Center: Live Digital Clock */}
          <div className="hidden lg:flex flex-col items-center text-center">
            <div className="font-mono text-xl font-black text-white tracking-widest">
              {currentTime || '--:--:--'}
            </div>
            <div className="text-[11px] text-slate-400 capitalize">
              {currentDate || 'Cargando fecha...'}
            </div>
          </div>

          {/* Right: Sound toggle, Guard Info & Logout */}
          <div className="flex items-center gap-2 sm:gap-3 w-full md:w-auto justify-end">
            {/* Link to Documentos */}
            <Link
              href="/documentos"
              className="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-semibold text-slate-300 hover:text-white hover:bg-slate-800 transition-colors"
              title="Ir al expediente de documentos"
            >
              <FileText className="h-3.5 w-3.5 text-emerald-400" />
              <span className="hidden sm:inline">Documentos</span>
            </Link>

            {/* Link to Credencial Digital */}
            <Link
              href="/credencial"
              className="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-semibold text-cyan-300 hover:text-white hover:bg-slate-800 transition-colors"
              title="Ver Credencial Móvil / QR del Trabajador"
            >
              <QrCode className="h-3.5 w-3.5 text-cyan-400" />
              <span className="hidden sm:inline">Credencial</span>
            </Link>

            {/* Link to Bandeja EHS (if EHS or Admin) */}
            {isEhsOrAdmin && (
              <Link
                href="/ehs/aprobaciones"
                className="inline-flex items-center gap-1.5 rounded-lg border border-amber-800/60 bg-amber-950/30 px-3 py-2 text-xs font-semibold text-amber-300 hover:text-white hover:bg-amber-900/50 transition-colors"
                title="Bandeja de Aprobación y Dictamen EHS"
              >
                <FileStack className="h-3.5 w-3.5 text-amber-400" />
                <span className="hidden sm:inline">Bandeja EHS</span>
              </Link>
            )}

            {/* Audio chime toggle */}
            <button
              type="button"
              onClick={toggleMute}
              className="p-2 rounded-lg border border-slate-800 bg-slate-900 text-slate-400 hover:text-white hover:bg-slate-800 transition-colors cursor-pointer"
              title={isMuted ? 'Activar sonido' : 'Silenciar sonido'}
            >
              {isMuted ? <VolumeX className="h-4 w-4 text-rose-400" /> : <Volume2 className="h-4 w-4 text-emerald-400" />}
            </button>

            {/* Theme Toggle */}
            <ThemeToggle />

            {/* Guard Profile Badge */}
            <div className="flex items-center gap-2.5 px-3 py-1.5 rounded-xl border border-slate-800 bg-slate-900/80">
              <div className="h-7 w-7 rounded-lg bg-slate-800 text-slate-200 font-bold text-xs flex items-center justify-center border border-slate-700">
                {guardName.slice(0, 2).toUpperCase()}
              </div>
              <div className="text-left hidden sm:block">
                <p className="text-xs font-bold text-white leading-none">{guardName}</p>
                <p className="text-[10px] text-slate-400 leading-tight">Guardia de Caseta</p>
              </div>
            </div>

            {/* Logout Button */}
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1.5 rounded-xl border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-bold text-slate-300 hover:bg-rose-950/40 hover:text-rose-300 hover:border-rose-500/40 transition-colors cursor-pointer"
              title="Cerrar sesión"
            >
              <LogOut className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Salir</span>
            </button>
          </div>
        </div>
      </header>

      {/* 2. Main Workstation Area */}
      <main className="flex-1 max-w-7xl mx-auto w-full p-4 sm:p-6 lg:p-8 space-y-6">
        {/* Top Control Bar: Mode Switcher & Presets */}
        <div className="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-4">
          {/* Mode Switcher Tabs */}
          <div className="flex rounded-xl bg-slate-900 p-1 border border-slate-800 max-w-xs shrink-0">
            <button
              type="button"
              onClick={() => {
                setMode('check_in');
                setError(null);
              }}
              className={`flex-1 py-2 px-4 rounded-lg text-xs font-bold transition-all cursor-pointer ${
                mode === 'check_in'
                  ? 'bg-emerald-600 text-white shadow-lg shadow-emerald-600/30'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              🟢 Entrada (Check-In)
            </button>
            <button
              type="button"
              onClick={() => {
                setMode('check_out');
                setError(null);
              }}
              className={`flex-1 py-2 px-4 rounded-lg text-xs font-bold transition-all cursor-pointer ${
                mode === 'check_out'
                  ? 'bg-blue-600 text-white shadow-lg shadow-blue-600/30'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              🔵 Salida (Check-Out)
            </button>
          </div>

          {/* Quick instructions status */}
          <div className="text-right text-xs text-slate-400 hidden md:block">
            Punto activo: <span className="font-bold text-white">{currentAp?.name}</span>
          </div>
        </div>

        {/* Demo Quick-Fill Presets */}
        <QuickFillPresets onSelectPreset={handleScanWorker} disabled={isLoading} />

        {/* HID Wedge Barcode Scanner Input */}
        <ScannerInput onScan={handleScanWorker} isLoading={isLoading} />

        {/* Error Notification Alert */}
        {error && (
          <div
            role="alert"
            className="rounded-2xl border-2 border-red-500/80 bg-red-950/80 p-5 text-white shadow-xl flex items-start gap-4 animate-in fade-in"
          >
            <AlertCircle className="h-6 w-6 text-red-400 shrink-0 mt-0.5" />
            <div className="flex-1">
              <h4 className="text-base font-bold text-red-200">Error en Verificación de Acceso</h4>
              <p className="text-sm text-red-300 mt-1">{error}</p>
            </div>
            <button
              type="button"
              onClick={() => setError(null)}
              className="text-xs font-bold text-red-300 hover:text-white px-2 py-1 rounded bg-red-900/60 cursor-pointer"
            >
              Cerrar
            </button>
          </div>
        )}

        {/* Verification Outcome Cards */}
        {checkOutResult && (
          <div className="space-y-6 animate-in fade-in duration-300">
            <StatusBanner
              result="granted"
              isCheckOut={true}
              checkedInAt={checkOutResult.checked_out_at}
            />
          </div>
        )}

        {checkInResult && (
          <div className="space-y-6 animate-in fade-in duration-300">
            {/* 3-Tier Visual Feedback Status Banner */}
            <StatusBanner
              result={checkInResult.result}
              denialReason={checkInResult.denial_reason}
              checkedInAt={checkInResult.checked_in_at}
            />

            {/* Quick Action Bar for Audit Print */}
            <div className="flex justify-end no-print">
              <button
                type="button"
                onClick={() => window.print()}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-slate-800 bg-slate-900 text-xs text-slate-300 hover:text-white hover:bg-slate-800 active:scale-95 transition-all cursor-pointer font-medium"
                title="Imprimir o guardar dictamen de acceso para auditoría"
              >
                <Printer className="w-3.5 h-3.5 text-slate-400" />
                <span>Imprimir Ficha de Acceso</span>
              </button>
            </div>

            {/* Profile & Document Breakdown Columns */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              {/* Worker Profile Card (5 cols on lg) */}
              <div className="lg:col-span-5">
                <WorkerCard
                  worker={checkInResult.worker}
                  contractor={checkInResult.contractor}
                  plant={checkInResult.plant}
                  result={checkInResult.result}
                />
              </div>

              {/* Document Breakdown Table (7 cols on lg) */}
              <div className="lg:col-span-7">
                <DocumentTable documents={checkInResult.document_snapshot || checkInResult.documents || []} />
              </div>
            </div>
          </div>
        )}

        {/* Idle Standby Screen (When no scan has occurred yet) */}
        {!checkInResult && !checkOutResult && !error && (
          <div className="rounded-2xl border border-slate-800 bg-slate-900/40 p-12 text-center text-slate-400 space-y-4">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-slate-800/80 text-emerald-400 border border-slate-700">
              <ScanBarcode className="h-8 w-8 animate-pulse" />
            </div>
            <div className="max-w-md mx-auto space-y-1">
              <h3 className="text-lg font-bold text-slate-200">
                Listo para Escaneo en Caseta
              </h3>
              <p className="text-xs text-slate-400">
                Pase el gafete o código QR frente al lector óptico, use los botones demo superiores o ingrese el UUID manualmente.
              </p>
            </div>
          </div>
        )}
      </main>

      {/* Footer Status Bar */}
      <footer className="border-t border-slate-800 bg-slate-950/80 px-6 py-3 text-xs text-slate-400 flex flex-col sm:flex-row items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
          <span>Sistema ASESAMO v1.0 • Control de Acceso Industrial y Validación Normativa</span>
        </div>
        <div>
          <span>Normativa STPS DC-3 • IMSS SUA • Padrón REPSE • NOM-031</span>
        </div>
      </footer>
    </div>
  );
}
