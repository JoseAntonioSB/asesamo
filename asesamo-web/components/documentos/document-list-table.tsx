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
  UploadCloud,
} from 'lucide-react';

interface DocumentListTableProps {
  documents: DocumentItem[];
  onOpenUpload: () => void;
  isLoading?: boolean;
}

export function DocumentListTable({ documents, onOpenUpload, isLoading }: DocumentListTableProps) {
  const getStatusBadge = (doc: DocumentItem) => {
    switch (doc.status) {
      case 'approved':
        return (
          <Badge className="bg-emerald-950/80 text-emerald-400 border border-emerald-800/80 flex items-center gap-1">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Aprobado
          </Badge>
        );
      case 'pending_review':
        return (
          <Badge className="bg-amber-950/80 text-amber-400 border border-amber-800/80 flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            Pendiente de revisión
          </Badge>
        );
      case 'rejected':
        return (
          <Badge className="bg-red-950/80 text-red-400 border border-red-800/80 flex items-center gap-1">
            <XCircle className="w-3.5 h-3.5" />
            Rechazado
          </Badge>
        );
      case 'expired':
        return (
          <Badge className="bg-rose-950/80 text-rose-300 border border-rose-800/80 flex items-center gap-1">
            <AlertTriangle className="w-3.5 h-3.5" />
            Vencido
          </Badge>
        );
      default:
        return (
          <Badge className="bg-slate-800 text-slate-300">
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
    // Trigger browser download via anchor or new window
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
      <div className="p-8 text-center bg-slate-900 border border-slate-800 rounded-xl">
        <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin mx-auto mb-3" />
        <p className="text-sm text-slate-400">Cargando documentos de cumplimiento...</p>
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div className="p-12 text-center bg-slate-900/60 border border-slate-800 rounded-xl space-y-4">
        <div className="inline-flex p-3 rounded-full bg-slate-800 text-slate-400">
          <FileText className="w-8 h-8" />
        </div>
        <div className="max-w-md mx-auto space-y-1">
          <h3 className="text-base font-semibold text-white">No hay documentos registrados</h3>
          <p className="text-sm text-slate-400">
            Aún no ha subido documentos obligatorios de seguridad (DC-3, REPSE, IMSS o INE).
          </p>
        </div>
        <Button
          onClick={onOpenUpload}
          className="bg-emerald-600 hover:bg-emerald-500 text-white inline-flex items-center gap-2"
        >
          <UploadCloud className="w-4 h-4" />
          Subir Primer Documento
        </Button>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-slate-800 bg-slate-900/80 overflow-hidden shadow-xl">
      <Table>
        <TableHeader className="bg-slate-950/60 border-b border-slate-800">
          <TableRow className="border-slate-800 hover:bg-transparent">
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Tipo</TableHead>
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Archivo</TableHead>
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Tamaño</TableHead>
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Emisión</TableHead>
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Vencimiento</TableHead>
            <TableHead className="text-slate-400 text-xs font-semibold uppercase tracking-wider">Estado</TableHead>
            <TableHead className="text-right text-slate-400 text-xs font-semibold uppercase tracking-wider">Acciones</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {documents.map((doc) => (
            <TableRow key={doc.id} className="border-slate-800/80 hover:bg-slate-800/40 transition-colors">
              <TableCell className="font-medium text-white">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs px-2 py-0.5 rounded bg-slate-800 text-emerald-400 border border-slate-700">
                    {doc.document_type_code || 'DOC'}
                  </span>
                  <span className="text-sm">{doc.document_type_name || 'Documento'}</span>
                </div>
              </TableCell>
              <TableCell className="text-slate-300 text-sm max-w-[200px] truncate" title={doc.file_name}>
                <div className="flex items-center gap-1.5 truncate">
                  <FileText className="w-4 h-4 text-slate-500 shrink-0" />
                  <span className="truncate">{doc.file_name}</span>
                </div>
              </TableCell>
              <TableCell className="text-slate-400 text-xs font-mono">
                {formatFileSize(doc.file_size_bytes)}
              </TableCell>
              <TableCell className="text-slate-300 text-sm font-mono">
                {doc.issued_at || '—'}
              </TableCell>
              <TableCell className="text-slate-300 text-sm font-mono">
                {doc.expires_at || <span className="text-slate-500">Sin vencimiento</span>}
              </TableCell>
              <TableCell>
                <div className="flex flex-col gap-1">
                  {getStatusBadge(doc)}
                  {doc.status === 'rejected' && doc.rejection_reason && (
                    <span className="text-xs text-red-400 italic">
                      Motivo: {doc.rejection_reason}
                    </span>
                  )}
                </div>
              </TableCell>
              <TableCell className="text-right">
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleDownload(doc)}
                  className="text-slate-300 hover:text-white hover:bg-slate-800 h-8 px-2.5 gap-1.5"
                >
                  <Download className="w-3.5 h-3.5" />
                  Descargar
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
