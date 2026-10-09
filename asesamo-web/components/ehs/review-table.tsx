'use client';

import React from 'react';
import { DocumentItem } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  FileText,
  Download,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Building,
  User,
  Check,
  Ban,
  ExternalLink,
} from 'lucide-react';

interface ReviewTableProps {
  documents: DocumentItem[];
  onApprove: (doc: DocumentItem) => void;
  onReject: (doc: DocumentItem) => void;
  processingId: string | null;
  isLoading?: boolean;
}

export function ReviewTable({
  documents,
  onApprove,
  onReject,
  processingId,
  isLoading,
}: ReviewTableProps) {
  const getStatusBadge = (doc: DocumentItem) => {
    switch (doc.status) {
      case 'approved':
        return (
          <Badge className="bg-emerald-950/80 text-emerald-400 border border-emerald-800/80 flex items-center gap-1 text-[11px]">
            <CheckCircle2 className="w-3 h-3" />
            Aprobado
          </Badge>
        );
      case 'pending_review':
        return (
          <Badge className="bg-amber-950/80 text-amber-300 border border-amber-800/80 flex items-center gap-1 text-[11px] font-medium">
            <Clock className="w-3 h-3 text-amber-400" />
            Por Revisar
          </Badge>
        );
      case 'rejected':
        return (
          <Badge className="bg-rose-950/80 text-rose-400 border border-rose-800/80 flex items-center gap-1 text-[11px]">
            <XCircle className="w-3 h-3" />
            Rechazado
          </Badge>
        );
      case 'expired':
        return (
          <Badge className="bg-rose-950/80 text-rose-300 border border-rose-800/80 flex items-center gap-1 text-[11px]">
            <AlertTriangle className="w-3 h-3" />
            Vencido
          </Badge>
        );
      default:
        return (
          <Badge className="bg-slate-800 text-slate-300 text-[11px]">
            {doc.status}
          </Badge>
        );
    }
  };

  const formatFileSize = (bytes?: number): string => {
    if (!bytes) return '—';
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const handleDownload = (doc: DocumentItem) => {
    const downloadUrl = `/api/v1/documents/${doc.id}/download`;
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.target = '_blank';
    link.download = doc.file_name;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (isLoading) {
    return (
      <div className="p-12 text-center bg-slate-900 border border-slate-800 rounded-xl">
        <div className="w-8 h-8 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
        <p className="text-sm text-slate-400">Consultando bandeja de revisión EHS...</p>
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div className="p-12 text-center bg-slate-900 border border-slate-800 rounded-xl">
        <CheckCircle2 className="w-10 h-10 text-emerald-400 mx-auto mb-3 opacity-60" />
        <h3 className="text-base font-semibold text-white">No hay documentos en esta vista</h3>
        <p className="text-xs text-slate-400 mt-1 max-w-md mx-auto">
          Todos los documentos correspondientes han sido dictaminados o no hay registros con el filtro seleccionado.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-xl">
      <div className="overflow-x-auto">
        <Table>
          <TableHeader className="bg-slate-950/80 border-b border-slate-800">
            <TableRow className="border-slate-800 hover:bg-transparent">
              <TableHead className="text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Documento
              </TableHead>
              <TableHead className="text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Titular / Empresa
              </TableHead>
              <TableHead className="text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Vigencia
              </TableHead>
              <TableHead className="text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Estado
              </TableHead>
              <TableHead className="text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Dictamen / Auditoría
              </TableHead>
              <TableHead className="text-right text-slate-400 font-mono text-xs uppercase tracking-wider py-3.5 px-4">
                Acciones
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {documents.map((doc) => {
              const isProcessing = processingId === doc.id;

              return (
                <TableRow
                  key={doc.id}
                  className="border-slate-800 hover:bg-slate-800/40 transition-colors"
                >
                  {/* Document Type & File */}
                  <TableCell className="py-3.5 px-4">
                    <div className="flex items-start gap-2.5">
                      <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 shrink-0 mt-0.5">
                        <FileText className="w-4 h-4" />
                      </div>
                      <div className="space-y-0.5">
                        <div className="flex items-center gap-1.5">
                          <span className="font-semibold text-white text-xs">
                            {doc.document_type_name || doc.document_type_code || 'Documento'}
                          </span>
                          {doc.document_type_code && (
                            <span className="px-1.5 py-0.2 rounded bg-slate-800 text-[10px] font-mono font-medium text-slate-300 border border-slate-700">
                              {doc.document_type_code}
                            </span>
                          )}
                        </div>
                        <div className="flex items-center gap-2 text-[11px] text-slate-400">
                          <span className="truncate max-w-[180px] font-mono" title={doc.file_name}>
                            {doc.file_name}
                          </span>
                          <span>•</span>
                          <span>{formatFileSize(doc.file_size_bytes)}</span>
                        </div>
                      </div>
                    </div>
                  </TableCell>

                  {/* Owner & Contractor */}
                  <TableCell className="py-3.5 px-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-1.5 text-xs font-medium text-slate-200">
                        <User className="w-3.5 h-3.5 text-slate-400" />
                        <span>{doc.owner_name || 'Sin titular especificado'}</span>
                        {doc.owner_type && (
                          <span className="text-[10px] px-1 rounded bg-slate-800 text-slate-400 font-mono">
                            {doc.owner_type === 'worker' ? 'Trabajador' : 'Contratista'}
                          </span>
                        )}
                      </div>
                      {doc.contractor_name && (
                        <div className="flex items-center gap-1.5 text-[11px] text-slate-400">
                          <Building className="w-3 h-3 text-slate-500" />
                          <span>{doc.contractor_name}</span>
                        </div>
                      )}
                    </div>
                  </TableCell>

                  {/* Validity Dates */}
                  <TableCell className="py-3.5 px-4 text-xs font-mono">
                    <div className="space-y-0.5 text-slate-300">
                      <div>
                        <span className="text-slate-500 text-[10px]">Emisión: </span>
                        {doc.issued_at ? doc.issued_at.split('T')[0] : '—'}
                      </div>
                      <div>
                        <span className="text-slate-500 text-[10px]">Vence: </span>
                        <span className={doc.status === 'expired' ? 'text-rose-400 font-bold' : ''}>
                          {doc.expires_at ? doc.expires_at.split('T')[0] : 'Indefinido'}
                        </span>
                      </div>
                    </div>
                  </TableCell>

                  {/* Status */}
                  <TableCell className="py-3.5 px-4">
                    {getStatusBadge(doc)}
                  </TableCell>

                  {/* Review Audit / Rejection Reason */}
                  <TableCell className="py-3.5 px-4 text-xs max-w-xs">
                    {doc.status === 'rejected' && doc.rejection_reason ? (
                      <div className="p-2 rounded bg-rose-950/40 border border-rose-900/60 text-rose-300 text-[11px] leading-tight">
                        <span className="font-semibold block text-rose-400 mb-0.5">Motivo:</span>
                        {doc.rejection_reason}
                      </div>
                    ) : doc.reviewed_by_name ? (
                      <div className="text-[11px] text-slate-400">
                        <span className="text-slate-500">Por: </span>
                        <span className="text-slate-300 font-medium">{doc.reviewed_by_name}</span>
                        {doc.reviewed_at && (
                          <div className="text-[10px] text-slate-500 font-mono">
                            {new Date(doc.reviewed_at).toLocaleDateString('es-MX')}
                          </div>
                        )}
                      </div>
                    ) : (
                      <span className="text-slate-500 text-[11px] italic">Sin dictaminar</span>
                    )}
                  </TableCell>

                  {/* Action Buttons */}
                  <TableCell className="py-3.5 px-4 text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleDownload(doc)}
                        className="h-8 px-2 text-slate-400 hover:text-white hover:bg-slate-800"
                        title="Ver / Descargar documento"
                      >
                        <ExternalLink className="w-3.5 h-3.5" />
                      </Button>

                      {doc.status === 'pending_review' ? (
                        <>
                          <Button
                            size="sm"
                            disabled={isProcessing}
                            onClick={() => onApprove(doc)}
                            className="h-8 px-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-xs shadow-md shadow-emerald-950/50 gap-1"
                          >
                            <Check className="w-3.5 h-3.5" />
                            Aprobar
                          </Button>
                          <Button
                            size="sm"
                            variant="destructive"
                            disabled={isProcessing}
                            onClick={() => onReject(doc)}
                            className="h-8 px-2.5 bg-rose-700/80 hover:bg-rose-600 text-white font-semibold text-xs gap-1"
                          >
                            <Ban className="w-3.5 h-3.5" />
                            Rechazar
                          </Button>
                        </>
                      ) : (
                        <div className="flex items-center gap-1">
                          {doc.status === 'approved' ? (
                            <Button
                              size="sm"
                              variant="ghost"
                              disabled={isProcessing}
                              onClick={() => onReject(doc)}
                              className="h-7 px-2 text-rose-400 hover:text-rose-300 hover:bg-rose-950/30 text-[11px]"
                            >
                              Cambiar a Rechazar
                            </Button>
                          ) : (
                            <Button
                              size="sm"
                              variant="ghost"
                              disabled={isProcessing}
                              onClick={() => onApprove(doc)}
                              className="h-7 px-2 text-emerald-400 hover:text-emerald-300 hover:bg-emerald-950/30 text-[11px]"
                            >
                              Cambiar a Aprobar
                            </Button>
                          )}
                        </div>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
