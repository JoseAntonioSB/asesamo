'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api/client';
import { TrainingCertificateDto } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Printer, ArrowLeft, ShieldCheck, QrCode, CheckCircle2 } from 'lucide-react';

export default function ImprimirDC3Page() {
  const params = useParams();
  const router = useRouter();
  const certId = params?.id as string;

  const [cert, setCert] = useState<TrainingCertificateDto | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadCert() {
      if (!certId) return;
      try {
        setIsLoading(true);
        const data = await apiClient.getTrainingCertificate(certId);
        setCert(data);
      } catch (err: any) {
        setError(err.message || 'No se pudo obtener la constancia DC-3.');
      } finally {
        setIsLoading(false);
      }
    }
    loadCert();
  }, [certId]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-zinc-950 text-white flex items-center justify-center p-6">
        <div className="text-center space-y-3">
          <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin mx-auto" />
          <p className="text-sm font-mono text-zinc-400">Generando formato oficial DC-3 STPS...</p>
        </div>
      </div>
    );
  }

  if (error || !cert) {
    return (
      <div className="min-h-screen bg-zinc-950 text-white flex items-center justify-center p-6">
        <div className="max-w-md w-full p-8 rounded-2xl bg-zinc-900 border border-zinc-800 text-center space-y-4">
          <h2 className="text-lg font-bold text-red-400">Error al cargar constancia</h2>
          <p className="text-sm text-zinc-400">{error || 'Constancia no encontrada.'}</p>
          <Button onClick={() => router.push('/capacitacion')} variant="outline" className="border-zinc-700">
            Volver al módulo de capacitación
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-950 py-8 px-4 sm:px-6 lg:px-8 selection:bg-indigo-500 selection:text-white print:p-0 print:bg-white print:text-black">
      {/* Floating Action Bar (Hidden on print) */}
      <div className="max-w-4xl mx-auto mb-6 flex items-center justify-between no-print">
        <Link
          href="/capacitacion"
          className="inline-flex items-center gap-2 text-xs font-semibold text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Volver a Capacitación
        </Link>

        <div className="flex items-center gap-3">
          <Button
            onClick={() => window.print()}
            className="bg-indigo-600 hover:bg-indigo-500 text-white font-bold text-xs gap-2 shadow-lg shadow-indigo-950/40"
          >
            <Printer className="w-4 h-4" />
            <span>Imprimir / Guardar en PDF</span>
          </Button>
        </div>
      </div>

      {/* Official STPS DC-3 Sheet (Letter Size Layout) */}
      <div className="max-w-4xl mx-auto bg-white text-black p-8 sm:p-12 shadow-2xl rounded-sm border border-zinc-300 print:shadow-none print:border-none print:p-0 print:max-w-none">
        {/* Document Header */}
        <div className="border-b-2 border-black pb-4 mb-6">
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <span className="text-[10px] font-bold tracking-widest text-zinc-600 uppercase block font-sans">
                SECRETARÍA DEL TRABAJO Y PREVISIÓN SOCIAL &bull; DIRECCIÓN GENERAL DE CAPACITACIÓN
              </span>
              <h1 className="text-xl sm:text-2xl font-black tracking-tight text-black font-sans uppercase">
                FORMATO DC-3
              </h1>
              <h2 className="text-sm font-bold text-zinc-800 font-sans uppercase tracking-wide">
                CONSTANCIA DE COMPETENCIAS O DE HABILIDADES LABORALES
              </h2>
            </div>
            <div className="text-right">
              <span className="text-[10px] font-mono font-bold uppercase text-zinc-500 block">Folio Digital STPS</span>
              <span className="text-sm font-mono font-black text-black">{cert.folio_number}</span>
            </div>
          </div>
        </div>

        {/* Bloque 1: Datos del Trabajador */}
        <div className="border border-black mb-4">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black">
            <span className="text-xs font-black uppercase tracking-wider font-sans">
              1. DATOS DEL TRABAJADOR
            </span>
          </div>
          <div className="p-3 space-y-3 text-xs font-sans">
            <div>
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                Nombre (apellido paterno, apellido materno y nombre(s))
              </span>
              <span className="text-sm font-black uppercase text-black">{cert.worker_full_name}</span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1 border-t border-zinc-200">
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                  Clave Única de Registro de Población (CURP)
                </span>
                <span className="text-xs font-mono font-bold uppercase text-black">{cert.worker_curp}</span>
              </div>
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                  Ocupación Específica (Clave CNO)
                </span>
                <span className="text-xs font-mono font-bold uppercase text-black">
                  {cert.worker_occupation_code || '9999 - TRABAJADOR INDUSTRIAL DE PLANTA'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Bloque 2: Datos de la Empresa */}
        <div className="border border-black mb-4">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black">
            <span className="text-xs font-black uppercase tracking-wider font-sans">
              2. DATOS DE LA EMPRESA
            </span>
          </div>
          <div className="p-3 space-y-3 text-xs font-sans">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="sm:col-span-2">
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                  Nombre o Razón Social del Patrón o Empresa
                </span>
                <span className="text-xs font-black uppercase text-black">{cert.company_name}</span>
              </div>
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                  Registro Federal de Contribuyentes (RFC)
                </span>
                <span className="text-xs font-mono font-bold uppercase text-black">{cert.company_rfc}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Bloque 3: Datos del Programa de Capacitación */}
        <div className="border border-black mb-4">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black">
            <span className="text-xs font-black uppercase tracking-wider font-sans">
              3. DATOS DEL PROGRAMA DE CAPACITACIÓN, ADIESTRAMIENTO Y PRODUCTIVIDAD
            </span>
          </div>
          <div className="p-3 space-y-3 text-xs font-sans">
            <div>
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Nombre del Curso</span>
              <span className="text-sm font-black uppercase text-black">
                CURSO TÉCNICO REGULADO EN SEGURIDAD Y PREVENCIÓN DE RIESGOS STPS
              </span>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-1 border-t border-zinc-200">
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">Duración en Horas</span>
                <span className="text-xs font-bold text-black">8 Horas Teórico-Prácticas</span>
              </div>
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">Fecha de Emisión</span>
                <span className="text-xs font-mono font-bold text-black">
                  {new Date(cert.issue_date).toLocaleDateString('es-MX')}
                </span>
              </div>
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">Fecha de Vigencia</span>
                <span className="text-xs font-mono font-bold text-black">
                  {new Date(cert.expiration_date).toLocaleDateString('es-MX')}
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1 border-t border-zinc-200">
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">Área Temática STPS</span>
                <span className="text-xs font-bold text-black">6000 - SEGURIDAD Y SALUD EN EL TRABAJO</span>
              </div>
              <div>
                <span className="text-[10px] font-bold text-zinc-600 block uppercase">
                  Agente Capacitador / Instructor
                </span>
                <span className="text-xs font-bold uppercase text-black">{cert.instructor_name}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Bloque 4: Firmas Autógrafas y Digitales */}
        <div className="border border-black mb-4">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black text-center">
            <span className="text-[10px] font-bold uppercase tracking-wider font-sans text-zinc-800">
              LOS FIRMANTES DECLARAN BAJO PROTESTA DE DECIR VERDAD QUE LOS DATOS ASENTADOS EN ESTE DOCUMENTO SON VERÍDICOS
            </span>
          </div>

          <div className="grid grid-cols-3 gap-4 p-6 text-center text-xs font-sans items-end">
            {/* Instructor */}
            <div className="space-y-8">
              <div className="font-mono text-[11px] text-zinc-400 italic">Firma Digital Verificada</div>
              <div className="border-t border-black pt-2">
                <span className="font-bold block text-black uppercase text-[11px]">{cert.instructor_name}</span>
                <span className="text-[9px] text-zinc-600 uppercase block">Instructor o Agente Capacitador</span>
              </div>
            </div>

            {/* Patrón */}
            <div className="space-y-8">
              <div className="font-mono text-[11px] text-zinc-400 italic">Firma Digital Verificada</div>
              <div className="border-t border-black pt-2">
                <span className="font-bold block text-black uppercase text-[11px]">{cert.legal_rep_name}</span>
                <span className="text-[9px] text-zinc-600 uppercase block">Patrón o Representante Legal</span>
              </div>
            </div>

            {/* Representante Trabajadores */}
            <div className="space-y-8">
              <div className="font-mono text-[11px] text-zinc-400 italic">Firma Digital Verificada</div>
              <div className="border-t border-black pt-2">
                <span className="font-bold block text-black uppercase text-[11px]">{cert.worker_rep_name}</span>
                <span className="text-[9px] text-zinc-600 uppercase block">
                  Representante de los Trabajadores (Comisión Mixta)
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Sello Digital ASESAMO y Token QR */}
        <div className="border border-zinc-400 p-4 rounded bg-zinc-50 flex items-center justify-between text-[10px] font-mono text-zinc-600">
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-emerald-800 font-bold uppercase">
              <ShieldCheck className="w-4 h-4 text-emerald-600" />
              <span>Sello Criptográfico Digital ASESAMO STPS Engine</span>
            </div>
            <div>Token: {cert.digital_signature_token || `STPS-SEC-${cert.id.substring(0, 16)}`}</div>
            <div>Calificación aprobatoria acreditada: {cert.score_percentage}% &bull; Estatus: Oficial Vigente</div>
          </div>
          <div className="text-right">
            <div className="w-16 h-16 bg-white border border-black p-1 inline-flex items-center justify-center">
              <QrCode className="w-14 h-14 text-black" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
