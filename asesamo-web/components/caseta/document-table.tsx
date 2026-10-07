// asesamo-web/components/caseta/document-table.tsx
'use client';

import React from 'react';
import {
  CheckCircle2,
  AlertTriangle,
  Clock,
  XCircle,
  FileQuestion,
  FileText,
  Calendar,
} from 'lucide-react';
import { DocumentSnapshotItem } from '@/lib/types/caseta';

export interface DocumentTableProps {
  documents: DocumentSnapshotItem[];
}

export function DocumentTable({ documents }: DocumentTableProps) {
  if (!documents || documents.length === 0) {
    return (
      <div className="w-full rounded-2xl border border-slate-800 bg-slate-900/95 p-6 text-center text-slate-400">
        No hay registros de documentos disponibles para este trabajador.
      </div>
    );
  }

  // Calculate summary counts
  const total = documents.length;
  const approved = documents.filter((d) => d.status === 'approved' && d.is_valid).length;

  return (
    <div className="w-full rounded-2xl border border-slate-800 bg-slate-900/95 overflow-hidden shadow-xl text-white">
      {/* Table Header with Counts */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 px-6 py-4 bg-slate-950/40">
        <div>
          <h3 className="text-lg font-bold text-white flex items-center gap-2">
            <FileText className="h-5 w-5 text-emerald-400" />
            Desglose de Documentación Normativa
          </h3>
          <p className="text-xs text-slate-400">
            Validación de constancias laborales mexicanas (STPS, IMSS, REPSE, Médicas)
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold bg-slate-800 text-slate-200 border border-slate-700">
            {approved} de {total} Conformes
          </span>
        </div>
      </div>

      {/* Responsive Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-950/80 text-xs uppercase tracking-wider text-slate-400 border-b border-slate-800">
            <tr>
              <th scope="col" className="px-5 py-3.5 font-semibold">Documento</th>
              <th scope="col" className="px-4 py-3.5 font-semibold">Ámbito</th>
              <th scope="col" className="px-4 py-3.5 font-semibold">Tipo</th>
              <th scope="col" className="px-4 py-3.5 font-semibold">Estatus</th>
              <th scope="col" className="px-4 py-3.5 font-semibold">Vigencia</th>
              <th scope="col" className="px-5 py-3.5 font-semibold">Observaciones</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/80 font-normal">
            {documents.map((doc, idx) => {
              // Determine Badge
              const isWarning =
                doc.status === 'approved' &&
                doc.expires_at &&
                new Date(doc.expires_at).getTime() - new Date().getTime() < 30 * 24 * 60 * 60 * 1000 &&
                new Date(doc.expires_at).getTime() >= new Date().getTime();

              const isExpired =
                doc.status === 'expired' ||
                (doc.expires_at && new Date(doc.expires_at).getTime() < new Date().getTime());

              const isRejected = doc.status === 'rejected';
              const isMissing = doc.status === 'missing';
              const isApproved = doc.status === 'approved' && !isExpired && !isWarning;

              return (
                <tr
                  key={`${doc.document_type_code}-${idx}`}
                  className="hover:bg-slate-800/40 transition-colors"
                >
                  {/* Document Name & Code */}
                  <td className="px-5 py-3.5 font-medium text-white whitespace-nowrap">
                    <div className="flex items-center gap-2.5">
                      <span className="font-mono text-xs font-bold px-2 py-0.5 rounded bg-slate-800 text-cyan-300 border border-slate-700">
                        {doc.document_type_code}
                      </span>
                      <span>{doc.document_name}</span>
                    </div>
                  </td>

                  {/* Applies To */}
                  <td className="px-4 py-3.5 text-xs text-slate-300">
                    <span className="capitalize">
                      {doc.applies_to === 'worker' ? 'Trabajador' : 'Contratista'}
                    </span>
                  </td>

                  {/* Required vs Optional */}
                  <td className="px-4 py-3.5 text-xs">
                    {doc.is_required ? (
                      <span className="font-semibold text-rose-400">Obligatorio</span>
                    ) : (
                      <span className="text-slate-400">Opcional</span>
                    )}
                  </td>

                  {/* Status Badge */}
                  <td className="px-4 py-3.5 whitespace-nowrap">
                    {isApproved && (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        Aprobado
                      </span>
                    )}
                    {isWarning && (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 animate-pulse">
                        <AlertTriangle className="h-3.5 w-3.5" />
                        Por vencer
                      </span>
                    )}
                    {isExpired && (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/30">
                        <Clock className="h-3.5 w-3.5" />
                        Vencido
                      </span>
                    )}
                    {isRejected && (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/30">
                        <XCircle className="h-3.5 w-3.5" />
                        Rechazado
                      </span>
                    )}
                    {isMissing && (
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-slate-500/10 text-slate-400 border border-slate-600/30">
                        <FileQuestion className="h-3.5 w-3.5" />
                        Faltante
                      </span>
                    )}
                  </td>

                  {/* Validity / Expiry Date */}
                  <td className="px-4 py-3.5 text-xs font-mono text-slate-300 whitespace-nowrap">
                    {doc.expires_at ? (
                      <div className="flex items-center gap-1.5">
                        <Calendar className="h-3.5 w-3.5 text-slate-500" />
                        <span>{doc.expires_at}</span>
                      </div>
                    ) : (
                      <span className="text-slate-500">Permanente</span>
                    )}
                  </td>

                  {/* Failure Reason / Observation */}
                  <td className="px-5 py-3.5 text-xs text-slate-300">
                    {doc.failure_reason ? (
                      <span className="text-rose-300 font-medium">
                        {doc.failure_reason}
                      </span>
                    ) : isWarning ? (
                      <span className="text-amber-300 font-medium">
                        Próximo a expirar (&lt; 30 días)
                      </span>
                    ) : (
                      <span className="text-slate-400">Vigente y validado</span>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
