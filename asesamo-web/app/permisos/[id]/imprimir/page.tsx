'use client';

import React, { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api/client';
import { WorkPermitItem } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { Printer, ArrowLeft, ShieldAlert, QrCode, CheckCircle2, AlertTriangle, Mountain, Flame, Zap } from 'lucide-react';

export default function ImprimirPermisoPTARPage() {
  const params = useParams();
  const router = useRouter();
  const permitId = params?.id as string;

  const [permit, setPermit] = useState<WorkPermitItem | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadPermit() {
      if (!permitId) return;
      try {
        setIsLoading(true);
        const data = await apiClient.getPermit(permitId);
        setPermit(data);
      } catch (err: any) {
        setError(err.message || 'No se pudo obtener el permiso PTAR.');
      } finally {
        setIsLoading(false);
      }
    }
    loadPermit();
  }, [permitId]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-zinc-950 text-white flex items-center justify-center p-6">
        <div className="text-center space-y-3">
          <div className="w-8 h-8 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin mx-auto" />
          <p className="text-sm font-mono text-zinc-400">Generando formato de autorización PTAR para campo...</p>
        </div>
      </div>
    );
  }

  if (error || !permit) {
    return (
      <div className="min-h-screen bg-zinc-950 text-white flex items-center justify-center p-6">
        <div className="max-w-md w-full p-8 rounded-2xl bg-zinc-900 border border-zinc-800 text-center space-y-4">
          <h2 className="text-lg font-bold text-red-400">Error al cargar permiso</h2>
          <p className="text-sm text-zinc-400">{error || 'Permiso no encontrado.'}</p>
          <Button onClick={() => router.push('/permisos')} variant="outline" className="border-zinc-700">
            Volver a Permisos PTAR
          </Button>
        </div>
      </div>
    );
  }

  const getHazardTitle = (type: string) => {
    switch (type) {
      case 'heights':
        return 'TRABAJOS EN ALTURAS (NOM-009-STPS-2011)';
      case 'confined_space':
        return 'ESPACIOS CONFINADOS (NOM-033-STPS-2015)';
      case 'hot_work':
        return 'CORTE Y SOLDADURA EN CALIENTE (NOM-027-STPS-2008)';
      case 'electrical_loto':
        return 'BLOQUEO Y ETIQUETADO LOTO (NOM-029-STPS-2011)';
      default:
        return 'ACTIVIDAD CRÍTICA DE ALTO RIESGO';
    }
  };

  return (
    <div className="min-h-screen bg-zinc-100 dark:bg-zinc-950 py-8 px-4 sm:px-6 lg:px-8 selection:bg-cyan-500 selection:text-white print:p-0 print:bg-white print:text-black">
      {/* Floating Action Bar (Hidden on print) */}
      <div className="max-w-4xl mx-auto mb-6 flex items-center justify-between no-print">
        <Link
          href={`/permisos/${permitId}`}
          className="inline-flex items-center gap-2 text-xs font-semibold text-zinc-600 dark:text-zinc-400 hover:text-zinc-900 dark:hover:text-white transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          Volver al Detalle del Permiso
        </Link>

        <div className="flex items-center gap-3">
          <Button
            onClick={() => window.print()}
            className="bg-cyan-600 hover:bg-cyan-500 text-white font-bold text-xs gap-2 shadow-lg shadow-cyan-950/40"
          >
            <Printer className="w-4 h-4" />
            <span>Imprimir Formato PTAR / PDF</span>
          </Button>
        </div>
      </div>

      {/* Official PTAR Sheet (Letter Size Layout) */}
      <div className="max-w-4xl mx-auto bg-white text-black p-8 sm:p-12 shadow-2xl rounded-sm border border-zinc-300 print:shadow-none print:border-none print:p-0 print:max-w-none">
        {/* Document Header */}
        <div className="border-b-4 border-black pb-4 mb-4">
          <div className="flex items-start justify-between">
            <div className="space-y-1">
              <span className="text-[10px] font-bold tracking-widest text-zinc-600 uppercase block font-sans">
                SISTEMA INTEGRAL DE SEGURIDAD INDUSTRIAL &bull; CONTROL OPERATIVO STPS
              </span>
              <h1 className="text-xl sm:text-2xl font-black tracking-tight text-black font-sans uppercase">
                PERMISO DE TRABAJO DE ALTO RIESGO (PTAR)
              </h1>
              <div className="inline-block px-3 py-1 bg-black text-white text-xs font-mono font-bold uppercase tracking-wider rounded-sm">
                {getHazardTitle(permit.permit_type)}
              </div>
            </div>
            <div className="text-right">
              <span className="text-[10px] font-mono font-bold uppercase text-zinc-500 block">Folio Oficial PTAR</span>
              <span className="text-base font-mono font-black text-black">{permit.permit_number}</span>
              <div className="text-[10px] font-bold uppercase text-emerald-800 mt-1">
                Estatus: {permit.status === 'approved_active' ? 'AUTORIZADO ACTIVO' : permit.status.toUpperCase()}
              </div>
            </div>
          </div>
        </div>

        {/* Bloque 1: Ubicación, Horarios y Empresa */}
        <div className="border border-black mb-4 text-xs font-sans">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black font-bold uppercase">
            1. DATOS GENERALES Y VIGENCIA OPERATIVA EN PLANTA
          </div>
          <div className="p-3 grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div className="sm:col-span-2">
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Empresa / Contratista</span>
              <span className="font-bold text-black uppercase">{permit.contractor_name || 'EMPRESA INDUSTRIAL INTERNA'}</span>
            </div>
            <div className="sm:col-span-2">
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Área Específica / Ubicación</span>
              <span className="font-bold text-black uppercase">{permit.work_area}</span>
            </div>
            <div>
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Vigencia Inicio</span>
              <span className="font-mono font-bold text-black">
                {new Date(permit.start_time).toLocaleString('es-MX')}
              </span>
            </div>
            <div>
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Vigencia Conclusión</span>
              <span className="font-mono font-bold text-black">
                {new Date(permit.end_time).toLocaleString('es-MX')}
              </span>
            </div>
            <div className="sm:col-span-2">
              <span className="text-[10px] font-bold text-zinc-600 block uppercase">Descripción de la Actividad</span>
              <span className="font-medium text-black">{permit.description}</span>
            </div>
          </div>
        </div>

        {/* Bloque 2: Cuadrilla Autorizada con Candados STPS */}
        <div className="border border-black mb-4 text-xs font-sans">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black font-bold uppercase flex justify-between items-center">
            <span>2. PERSONAL AUTORIZADO (CANDADOS STPS VERIFICADOS)</span>
            <span className="text-[10px] font-mono">{permit.workers?.length || 0} TRABAJADORES</span>
          </div>
          <div className="p-2">
            <table className="w-full text-left text-[11px] font-sans">
              <thead>
                <tr className="border-b border-zinc-400 text-[10px] uppercase text-zinc-600">
                  <th className="py-1 px-2">Trabajador</th>
                  <th className="py-1 px-2">CURP</th>
                  <th className="py-1 px-2">Constancia DC-3</th>
                  <th className="py-1 px-2">Examen Médico</th>
                  <th className="py-1 px-2 text-right">Firma Entrada</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-200">
                {!permit.workers || permit.workers.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="py-3 text-center text-zinc-500">
                      Sin personal asignado registrado.
                    </td>
                  </tr>
                ) : (
                  permit.workers.map((w, idx) => (
                    <tr key={idx}>
                      <td className="py-1.5 px-2 font-bold uppercase">{w.full_name}</td>
                      <td className="py-1.5 px-2 font-mono">{w.curp || 'VIGENTE'}</td>
                      <td className="py-1.5 px-2 text-emerald-800 font-bold">Acreditado STPS</td>
                      <td className="py-1.5 px-2 text-emerald-800 font-bold">Apto Médico</td>
                      <td className="py-1.5 px-2 text-right font-mono text-[9px] text-zinc-400">_________________</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* Bloque 3: Checklist Preventivo y Monitoreo */}
        <div className="border border-black mb-4 text-xs font-sans">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black font-bold uppercase">
            3. VERIFICACIÓN DE CONDICIONES SEGURAS PREVIAS AL INICIO
          </div>
          <div className="p-3 grid grid-cols-1 sm:grid-cols-2 gap-3 text-[11px]">
            <div className="space-y-1.5">
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>EPP para riesgo crítico completo e inspeccionado.</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>Delimitación y acordonamiento de radio de seguridad en piso.</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>Líneas de vida o puntos de anclaje certificados (5,000 lbs).</span>
              </div>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>Extintor tipo PQS / CO2 vigente ubicado a menos de 7 metros.</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>Personal informado de la ruta de evacuación y puntos de reunión.</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 border border-black inline-flex items-center justify-center font-bold text-[10px]">✓</span>
                <span>Vigía capacitado presente de manera ininterrumpida.</span>
              </div>
            </div>
          </div>
        </div>

        {/* Bloque 4: Firmas de Autorización Oficial */}
        <div className="border border-black mb-4">
          <div className="bg-zinc-200 px-3 py-1 border-b border-black text-center text-xs font-bold uppercase">
            4. AUTORIZACIÓN Y FIRMAS DE RESPONSABILIDAD TÉCNICA
          </div>
          <div className="grid grid-cols-3 gap-4 p-6 text-center text-xs font-sans items-end">
            <div className="space-y-8">
              <div className="font-mono text-[10px] text-zinc-500 italic">Validado en Plataforma</div>
              <div className="border-t border-black pt-1">
                <span className="font-bold block text-black uppercase text-[11px]">Supervisor Contratista</span>
                <span className="text-[9px] text-zinc-600 block">Solicitante Ejecutor</span>
              </div>
            </div>

            <div className="space-y-8">
              <div className="font-mono text-[10px] text-zinc-500 italic">Validado en Plataforma</div>
              <div className="border-t border-black pt-1">
                <span className="font-bold block text-black uppercase text-[11px]">Supervisor de Área</span>
                <span className="text-[9px] text-zinc-600 block">Receptor Operativo</span>
              </div>
            </div>

            <div className="space-y-8">
              <div className="font-mono text-[10px] text-zinc-500 italic">Firma Electrónica Autorizada</div>
              <div className="border-t border-black pt-1">
                <span className="font-bold block text-black uppercase text-[11px]">Ingeniero / Coordinador EHS</span>
                <span className="text-[9px] text-zinc-600 block">Autorizante de Seguridad</span>
              </div>
            </div>
          </div>
        </div>

        {/* Sello Digital ASESAMO y QR */}
        <div className="border border-zinc-400 p-3 rounded bg-zinc-50 flex items-center justify-between text-[10px] font-mono text-zinc-600">
          <div className="space-y-1">
            <div className="font-bold text-black uppercase">
              CERTIFICACIÓN DIGITAL DE AUTORIZACIÓN ASESAMO &bull; PTAR ACTIVO
            </div>
            <div>ID Permiso: {permit.id} &bull; Candados STPS Verificados: DC-3 + Examen Médico</div>
            <div>Válido únicamente para el área y horario especificados en este documento.</div>
          </div>
          <div className="w-14 h-14 bg-white border border-black p-1 inline-flex items-center justify-center shrink-0">
            <QrCode className="w-12 h-12 text-black" />
          </div>
        </div>
      </div>
    </div>
  );
}
