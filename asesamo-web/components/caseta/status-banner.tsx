// asesamo-web/components/caseta/status-banner.tsx
'use client';

import React from 'react';
import { CheckCircle2, AlertTriangle, XCircle, ShieldAlert, Clock } from 'lucide-react';
import { AccessResult } from '@/lib/types/caseta';

export interface StatusBannerProps {
  result: AccessResult;
  denialReason?: string | null;
  checkedInAt?: string;
  isCheckOut?: boolean;
}

export function StatusBanner({
  result,
  denialReason,
  checkedInAt,
  isCheckOut = false,
}: StatusBannerProps) {
  // Format check-in timestamp
  const formattedTime = checkedInAt
    ? new Date(checkedInAt).toLocaleTimeString('es-MX', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    : '';

  if (isCheckOut) {
    return (
      <div
        role="alert"
        aria-live="polite"
        className="w-full rounded-2xl border-2 border-blue-500/80 bg-gradient-to-r from-blue-950/90 via-slate-900 to-blue-950/90 p-6 md:p-8 text-white shadow-2xl shadow-blue-950/50 transition-all duration-300"
      >
        <div className="flex flex-col sm:flex-row items-center sm:items-start gap-5 text-center sm:text-left">
          <div className="flex h-16 w-16 md:h-20 md:w-20 shrink-0 items-center justify-center rounded-2xl bg-blue-500/20 text-blue-400 border border-blue-500/40">
            <Clock className="h-10 w-10 md:h-12 md:w-12 animate-pulse" />
          </div>
          <div className="flex-1 space-y-1">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-300 border border-blue-500/30">
              SALIDA REGISTRADA
            </div>
            <h1 className="text-3xl md:text-5xl font-black tracking-tight text-blue-200">
              SALIDA AUTORIZADA
            </h1>
            <p className="text-base md:text-lg font-medium text-slate-300">
              Se ha completado el turno y cerrado el registro de permanencia en planta.
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (result === 'granted') {
    return (
      <div
        role="alert"
        aria-live="assertive"
        className="w-full rounded-2xl border-2 border-emerald-500 bg-gradient-to-r from-emerald-950/90 via-slate-950 to-emerald-950/90 p-6 md:p-8 text-white shadow-2xl shadow-emerald-950/60 transition-all duration-300"
      >
        <div className="flex flex-col sm:flex-row items-center sm:items-start gap-5 text-center sm:text-left">
          <div className="flex h-16 w-16 md:h-20 md:w-20 shrink-0 items-center justify-center rounded-2xl bg-emerald-500/20 text-emerald-400 border border-emerald-500/40 shadow-inner shadow-emerald-500/20">
            <CheckCircle2 className="h-10 w-10 md:h-12 md:w-12 animate-in zoom-in-95 duration-200" />
          </div>
          <div className="flex-1 space-y-2">
            <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
              <span className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-emerald-500/20 text-emerald-300 border border-emerald-500/40">
                Verificación Exitosa
              </span>
              {formattedTime && (
                <span className="text-xs text-slate-400 font-mono">
                  Hora: {formattedTime} hrs
                </span>
              )}
            </div>
            <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black tracking-wider text-emerald-300 drop-shadow-[0_2px_12px_rgba(16,185,129,0.3)]">
              ACCESO PERMITIDO
            </h1>
            <p className="text-base sm:text-lg font-semibold text-emerald-100/90">
              TRABAJADOR AUTORIZADO — DOCUMENTACIÓN VIGENTE AL 100%
            </p>
            <p className="text-xs sm:text-sm text-slate-300">
              Todas las certificaciones obligatorias (STPS DC-3, INE, Examen Médico y registros patronales REPSE/SUA) se encuentran en regla.
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (result === 'granted_with_warning') {
    return (
      <div
        role="alert"
        aria-live="assertive"
        className="w-full rounded-2xl border-2 border-amber-500 bg-gradient-to-r from-amber-950/90 via-slate-950 to-amber-950/90 p-6 md:p-8 text-white shadow-2xl shadow-amber-950/60 transition-all duration-300"
      >
        <div className="flex flex-col sm:flex-row items-center sm:items-start gap-5 text-center sm:text-left">
          <div className="flex h-16 w-16 md:h-20 md:w-20 shrink-0 items-center justify-center rounded-2xl bg-amber-500/20 text-amber-400 border border-amber-500/40 shadow-inner shadow-amber-500/20">
            <AlertTriangle className="h-10 w-10 md:h-12 md:w-12 animate-in zoom-in-95 duration-200" />
          </div>
          <div className="flex-1 space-y-2">
            <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
              <span className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-amber-500/20 text-amber-300 border border-amber-500/40">
                Aviso Preventivo
              </span>
              {formattedTime && (
                <span className="text-xs text-slate-400 font-mono">
                  Hora: {formattedTime} hrs
                </span>
              )}
            </div>
            <h1 className="text-2xl sm:text-3xl lg:text-4xl font-black tracking-wide text-amber-300 drop-shadow-[0_2px_12px_rgba(245,158,11,0.3)]">
              ACCESO CON ADVERTENCIA - DOCUMENTOS PRÓXIMOS A VENCER
            </h1>
            <p className="text-base sm:text-lg font-semibold text-amber-100/90">
              PERMITIR EL PASO Y NOTIFICAR AL TRABAJADOR
            </p>
            <div className="rounded-xl border border-amber-500/40 bg-amber-950/40 p-3 text-xs sm:text-sm text-amber-200">
              <span className="font-bold">Acción requerida:</span> El trabajador tiene constancias que vencerán en los próximos 30 días. Recuerde a su contratista renovar la documentación en la plataforma ASESAMO para evitar futuros bloqueos.
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Denied
  return (
    <div
      role="alert"
      aria-live="assertive"
      className="w-full rounded-2xl border-2 border-red-500 bg-gradient-to-r from-red-950/95 via-slate-950 to-red-950/95 p-6 md:p-8 text-white shadow-2xl shadow-red-950/70 transition-all duration-300"
    >
      <div className="flex flex-col sm:flex-row items-center sm:items-start gap-5 text-center sm:text-left">
        <div className="flex h-16 w-16 md:h-20 md:w-20 shrink-0 items-center justify-center rounded-2xl bg-red-500/20 text-red-400 border border-red-500/40 shadow-inner shadow-red-500/20">
          <XCircle className="h-10 w-10 md:h-12 md:w-12 animate-in zoom-in-95 duration-200" />
        </div>
        <div className="flex-1 space-y-3">
          <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
            <span className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-red-500/20 text-red-300 border border-red-500/40">
              Bloqueo Estricto
            </span>
            {formattedTime && (
              <span className="text-xs text-slate-400 font-mono">
                Hora: {formattedTime} hrs
              </span>
            )}
          </div>
          <h1 className="text-3xl sm:text-4xl lg:text-5xl font-black tracking-wider text-red-400 drop-shadow-[0_2px_12px_rgba(239,68,68,0.4)]">
            ACCESO DENEGADO
          </h1>
          <p className="text-base sm:text-lg font-semibold text-red-200">
            PROHIBIDO EL INGRESO — NO PERMITIR EL PASO AL TRABAJADOR
          </p>

          {/* Prominent Spanish Denial Reason Box */}
          <div className="rounded-xl border-2 border-red-500/50 bg-red-950/60 p-4 text-left shadow-lg">
            <div className="flex items-center gap-2 text-xs font-bold text-red-300 uppercase tracking-wide mb-1">
              <ShieldAlert className="h-4 w-4 text-red-400 shrink-0" />
              Motivo del Rechazo:
            </div>
            <p className="text-base sm:text-lg font-bold text-white break-words">
              {denialReason || 'Documentos obligatorios vencidos o no acreditados según la política de seguridad.'}
            </p>
            <p className="mt-2 text-xs text-red-300/80">
              Normativa de la planta: Por política estricta de control de acceso, los trabajadores sin constancias STPS DC-3 vigentes o certificados médicos aprobados no pueden ingresar al inmueble.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
