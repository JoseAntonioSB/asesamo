'use client';

import React from 'react';
import { WorkerQRCode } from './worker-qr-code';
import { WorkerCredentialResponse } from '@/lib/api/types';
import {
  ShieldCheck,
  Building,
  HardHat,
  Briefcase,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  FileBadge,
  Calendar,
  Sparkles,
  Award,
} from 'lucide-react';

interface CredentialBadgeProps {
  data: WorkerCredentialResponse;
  className?: string;
}

export function CredentialBadge({ data, className = '' }: CredentialBadgeProps) {
  const { worker, contractor, plant, compliance_status, documents, qr_payload, generated_at } = data;

  const isGranted = compliance_status === 'granted';
  const isWarning = compliance_status === 'granted_with_warning';
  const isDenied = compliance_status === 'denied';

  const getStatusBadge = () => {
    if (isGranted) {
      return {
        bg: 'bg-emerald-600/90 text-white border-emerald-400/40',
        icon: <CheckCircle2 className="w-4 h-4 text-emerald-200 shrink-0" />,
        text: 'APTO PARA INGRESO',
        subtext: 'Cumplimiento Normativo STPS / EHS al 100%',
        glow: 'shadow-emerald-500/20',
      };
    }
    if (isWarning) {
      return {
        bg: 'bg-amber-600/90 text-white border-amber-400/40',
        icon: <AlertTriangle className="w-4 h-4 text-amber-200 shrink-0" />,
        text: 'ACCESO CON ADVERTENCIA',
        subtext: 'Documentación próxima a vencer (requiere renovación)',
        glow: 'shadow-amber-500/20',
      };
    }
    return {
      bg: 'bg-rose-600/90 text-white border-rose-400/40',
      icon: <XCircle className="w-4 h-4 text-rose-200 shrink-0" />,
      text: 'ACCESO DENEGADO',
      subtext: 'Documentación faltante o vencida (visite caseta/EHS)',
      glow: 'shadow-rose-500/20',
    };
  };

  const statusInfo = getStatusBadge();

  return (
    <div
      className={`relative w-full max-w-md mx-auto rounded-3xl border-2 border-slate-700/80 bg-gradient-to-b from-slate-900 via-slate-900 to-slate-950 p-5 sm:p-6 shadow-2xl text-slate-100 overflow-hidden ${statusInfo.glow} ${className}`}
    >
      {/* 1. Lanyard Slot Cutout Visual */}
      <div className="flex justify-center mb-4">
        <div className="w-14 h-2.5 rounded-full bg-slate-950 border border-slate-700/60 shadow-inner" />
      </div>

      {/* 2. Industrial Security Card Header */}
      <div className="flex items-center justify-between border-b border-slate-800 pb-4 mb-4">
        <div className="flex items-center gap-2.5">
          <div className="w-9 h-9 rounded-xl bg-emerald-600 text-white flex items-center justify-center font-bold shadow-md shadow-emerald-600/30">
            <ShieldCheck className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="font-mono text-xs font-black tracking-widest text-white">ASESAMO</span>
              <span className="text-[10px] px-1.5 py-0.2 rounded font-mono font-bold bg-slate-800 text-emerald-400 border border-slate-700">
                PASS
              </span>
            </div>
            <p className="text-[10px] text-slate-400 uppercase tracking-wider font-semibold">
              Control de Acceso Industrial
            </p>
          </div>
        </div>

        {/* Plant Badge */}
        <div className="text-right">
          <div className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-slate-800 text-[11px] font-bold text-slate-300 border border-slate-700">
            <Building className="w-3 h-3 text-cyan-400" />
            <span className="truncate max-w-[120px]">{plant.name || 'Planta Querétaro'}</span>
          </div>
          <p className="text-[9px] text-slate-500 uppercase font-mono mt-0.5">Vigencia Oficial</p>
        </div>
      </div>

      {/* 3. Worker Identity Banner */}
      <div className="flex items-start gap-4 mb-5">
        {/* Worker Avatar with Industrial Hard Hat Indicator */}
        <div className="relative shrink-0">
          <div className="w-16 h-16 sm:w-18 sm:h-18 rounded-2xl bg-gradient-to-tr from-slate-800 to-slate-700 border-2 border-emerald-500/60 text-white flex items-center justify-center font-black text-xl shadow-lg">
            {worker.full_name
              .split(' ')
              .map((n) => n[0])
              .slice(0, 2)
              .join('')
              .toUpperCase()}
          </div>
          <div className="absolute -bottom-1 -right-1 p-1 rounded-full bg-slate-900 border border-slate-700 text-amber-400 shadow">
            <HardHat className="w-3.5 h-3.5" />
          </div>
        </div>

        {/* Worker Name, Job Title & ID Details */}
        <div className="flex-1 min-w-0">
          <h2 className="text-lg sm:text-xl font-bold text-white tracking-tight truncate leading-tight">
            {worker.full_name}
          </h2>
          <div className="flex items-center gap-1 text-xs text-emerald-400 font-semibold mt-0.5">
            <Briefcase className="w-3 h-3 shrink-0" />
            <span className="truncate">{worker.job_title || 'Técnico Especialista'}</span>
          </div>

          <div className="mt-2 grid grid-cols-2 gap-1 text-[11px] font-mono text-slate-300">
            <div>
              <span className="text-[9px] uppercase text-slate-500 block">CURP</span>
              <span className="font-bold truncate block">{worker.curp || 'N/A'}</span>
            </div>
            <div>
              <span className="text-[9px] uppercase text-slate-500 block">NSS / IMSS</span>
              <span className="font-bold truncate block">{worker.imss_number || 'N/A'}</span>
            </div>
          </div>
        </div>
      </div>

      {/* 4. Contractor Company Info */}
      <div className="p-3 rounded-xl bg-slate-950/60 border border-slate-800/80 mb-5 text-xs">
        <div className="text-[10px] uppercase font-bold text-slate-500 tracking-wider">
          Empresa Contratista
        </div>
        <div className="text-white font-bold text-sm truncate mt-0.5">{contractor.name}</div>
        <div className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1 text-[10px] font-mono text-slate-400">
          {contractor.rfc && (
            <span>
              RFC: <strong className="text-slate-300">{contractor.rfc}</strong>
            </span>
          )}
          {contractor.repse_number && (
            <span className="inline-flex items-center gap-1 text-amber-400">
              <Award className="w-2.5 h-2.5" />
              REPSE: <strong>{contractor.repse_number}</strong>
            </span>
          )}
        </div>
      </div>

      {/* 5. Live EHS Traffic Light Banner */}
      <div className={`p-3 rounded-xl border flex items-center gap-3 mb-5 shadow-md ${statusInfo.bg}`}>
        {statusInfo.icon}
        <div className="flex-1 min-w-0">
          <div className="text-xs font-black tracking-wider uppercase">{statusInfo.text}</div>
          <div className="text-[11px] opacity-90 truncate">{statusInfo.subtext}</div>
        </div>
      </div>

      {/* 6. Dynamic High-Contrast QR Code */}
      <div className="py-2">
        <WorkerQRCode
          payload={qr_payload}
          workerName={worker.full_name}
          complianceStatus={compliance_status}
        />
      </div>

      {/* 7. Document Compliance Chips Summary */}
      {documents && documents.length > 0 && (
        <div className="mt-5 pt-4 border-t border-slate-800">
          <div className="flex items-center justify-between mb-2">
            <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1">
              <FileBadge className="w-3 h-3 text-emerald-400" />
              Expediente de Seguridad
            </span>
            <span className="text-[9px] font-mono text-slate-500">
              {documents.filter((d) => d.is_valid).length}/{documents.length} Vigentes
            </span>
          </div>

          <div className="grid grid-cols-2 gap-1.5">
            {documents.slice(0, 4).map((doc, idx) => (
              <div
                key={idx}
                className="p-1.5 rounded-lg bg-slate-950/80 border border-slate-800 flex items-center justify-between text-[10px]"
              >
                <div className="truncate mr-1">
                  <span className="font-bold text-white font-mono block">
                    {doc.document_type_code}
                  </span>
                  <span className="text-slate-400 text-[9px] truncate block">
                    {doc.expires_at ? `Vence: ${doc.expires_at}` : 'Sin vencimiento'}
                  </span>
                </div>
                {doc.is_valid ? (
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                ) : (
                  <XCircle className="w-3.5 h-3.5 text-rose-400 shrink-0" />
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 8. Timestamp Footer */}
      <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-[9px] text-slate-500 font-mono">
        <span>GENERADO: {new Date(generated_at).toLocaleDateString('es-MX')}</span>
        <span>ID: {qr_payload.slice(0, 8)}</span>
      </div>
    </div>
  );
}
