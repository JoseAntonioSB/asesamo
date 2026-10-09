'use client';

import React, { useState, useRef, DragEvent, ChangeEvent } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { apiClient, ApiError } from '@/lib/api/client';
import { DocumentTypeItem, DocumentItem } from '@/lib/api/types';
import { toast } from 'sonner';
import { UploadCloud, FileText, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react';

interface DocumentUploadModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  documentTypes: DocumentTypeItem[];
  onUploadSuccess: (newDoc: DocumentItem) => void;
  defaultDocTypeCode?: string;
}

const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10 MB
const ALLOWED_MIME_TYPES = ['application/pdf', 'image/jpeg', 'image/png'];
const ALLOWED_EXTENSIONS = ['.pdf', '.jpg', '.jpeg', '.png'];

export function DocumentUploadModal({
  open,
  onOpenChange,
  documentTypes,
  onUploadSuccess,
  defaultDocTypeCode,
}: DocumentUploadModalProps) {
  const [selectedTypeCode, setSelectedTypeCode] = useState<string>(
    defaultDocTypeCode || (documentTypes[0]?.code ?? 'DC3')
  );
  const [issuedAt, setIssuedAt] = useState<string>('');
  const [expiresAt, setExpiresAt] = useState<string>('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dragActive, setDragActive] = useState<boolean>(false);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // Validate file client-side
  const validateFile = (file: File): boolean => {
    setErrorMessage(null);

    // 1. Size check
    if (file.size > MAX_FILE_SIZE) {
      const msg = `El archivo excede el tamaño máximo permitido de 10 MB (${(file.size / (1024 * 1024)).toFixed(1)} MB).`;
      setErrorMessage(msg);
      toast.error('Archivo demasiado grande', { description: msg });
      return false;
    }

    if (file.size === 0) {
      const msg = 'El archivo está vacío.';
      setErrorMessage(msg);
      toast.error('Archivo no válido', { description: msg });
      return false;
    }

    // 2. MIME and extension check
    const fileName = file.name.toLowerCase();
    const hasValidExt = ALLOWED_EXTENSIONS.some((ext) => fileName.endsWith(ext));
    const hasValidMime = ALLOWED_MIME_TYPES.includes(file.type);

    if (!hasValidExt && !hasValidMime) {
      const msg = 'Formato no permitido. Solo se aceptan archivos PDF e imágenes JPG o PNG.';
      setErrorMessage(msg);
      toast.error('Tipo de archivo no permitido', { description: msg });
      return false;
    }

    return true;
  };

  const handleDrag = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];
      if (validateFile(file)) {
        setSelectedFile(file);
      }
    }
  };

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      if (validateFile(file)) {
        setSelectedFile(file);
      }
    }
  };

  const resetForm = () => {
    setSelectedFile(null);
    setErrorMessage(null);
    setIssuedAt('');
    setExpiresAt('');
    setIsUploading(false);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!selectedFile) {
      setErrorMessage('Por favor seleccione o arrastre un archivo para subir.');
      toast.error('Falta el archivo', { description: 'Seleccione un documento válido.' });
      return;
    }

    if (!selectedTypeCode) {
      setErrorMessage('Debe seleccionar el tipo de documento.');
      return;
    }

    try {
      setIsUploading(true);

      const createdDoc = await apiClient.uploadDocument({
        file: selectedFile,
        documentTypeCode: selectedTypeCode,
        issuedAt: issuedAt || undefined,
        expiresAt: expiresAt || undefined,
      });

      toast.success('Documento subido correctamente', {
        description: `${selectedFile.name} está ahora en estado "Pendiente de revisión".`,
      });

      onUploadSuccess(createdDoc);
      resetForm();
      onOpenChange(false);
    } catch (err: unknown) {
      setIsUploading(false);
      let msg = 'Ocurrió un error al subir el documento. Intente nuevamente.';
      if (err instanceof ApiError) {
        msg = err.message;
      }
      setErrorMessage(msg);
      toast.error('Error al subir documento', { description: msg });
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => {
      if (!isUploading) {
        if (!isOpen) resetForm();
        onOpenChange(isOpen);
      }
    }}>
      <DialogContent className="sm:max-w-[540px] bg-slate-900 border-slate-800 text-slate-100">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold text-white flex items-center gap-2">
            <UploadCloud className="w-5 h-5 text-emerald-400" />
            Carga de Documento de Cumplimiento
          </DialogTitle>
          <DialogDescription className="text-slate-400 text-sm">
            Suba constancias oficiales (DC-3, REPSE, SUA, INE, Médicos) para validación de acceso industrial.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          {/* Document Type Selector */}
          <div className="space-y-1.5">
            <label htmlFor="docType" className="block text-xs font-semibold text-slate-300 uppercase tracking-wider">
              Tipo de Documento <span className="text-red-400">*</span>
            </label>
            <select
              id="docType"
              value={selectedTypeCode}
              onChange={(e) => setSelectedTypeCode(e.target.value)}
              disabled={isUploading}
              className="w-full px-3 py-2 bg-slate-950 border border-slate-700 rounded-lg text-white text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/60 focus:border-emerald-500"
            >
              {documentTypes.map((dt) => (
                <option key={dt.code} value={dt.code}>
                  {dt.code} — {dt.name} {dt.applies_to ? `(${dt.applies_to})` : ''}
                </option>
              ))}
            </select>
          </div>

          {/* Drag & Drop Upload Zone */}
          <div className="space-y-1.5">
            <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider">
              Archivo Digital <span className="text-red-400">*</span>
            </label>
            <div
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`relative border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors ${
                dragActive
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : selectedFile
                  ? 'border-emerald-600/60 bg-emerald-950/20'
                  : 'border-slate-700 bg-slate-950 hover:border-slate-500'
              }`}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.jpg,.jpeg,.png,application/pdf,image/jpeg,image/png"
                onChange={handleFileChange}
                disabled={isUploading}
                className="hidden"
              />

              {selectedFile ? (
                <div className="flex items-center justify-center gap-3 text-left">
                  <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-400 shrink-0">
                    <FileText className="w-6 h-6" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-white truncate">{selectedFile.name}</p>
                    <p className="text-xs text-slate-400 font-mono">
                      {formatFileSize(selectedFile.size)} • {selectedFile.type || 'Documento'}
                    </p>
                  </div>
                  <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
                </div>
              ) : (
                <div className="space-y-2">
                  <div className="inline-flex p-3 rounded-full bg-slate-800 text-slate-300">
                    <UploadCloud className="w-6 h-6 text-emerald-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-slate-200">
                      Haga clic para seleccionar o arrastre el archivo aquí
                    </p>
                    <p className="text-xs text-slate-400 mt-1">
                      PDF, JPG o PNG (Máximo 10 MB por archivo)
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Dates: Issued & Expiration */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label htmlFor="issuedAt" className="block text-xs font-semibold text-slate-300 uppercase tracking-wider">
                Fecha de Emisión
              </label>
              <Input
                id="issuedAt"
                type="date"
                value={issuedAt}
                onChange={(e) => setIssuedAt(e.target.value)}
                disabled={isUploading}
                className="bg-slate-950 border-slate-700 text-white"
              />
            </div>
            <div className="space-y-1.5">
              <label htmlFor="expiresAt" className="block text-xs font-semibold text-slate-300 uppercase tracking-wider">
                Fecha de Vencimiento
              </label>
              <Input
                id="expiresAt"
                type="date"
                value={expiresAt}
                onChange={(e) => setExpiresAt(e.target.value)}
                disabled={isUploading}
                className="bg-slate-950 border-slate-700 text-white"
              />
            </div>
          </div>

          {/* Inline Error Message */}
          {errorMessage && (
            <div className="bg-red-950/60 border border-red-500/50 rounded-lg p-3 flex items-start gap-2.5 text-xs text-red-200">
              <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
              <span>{errorMessage}</span>
            </div>
          )}

          <DialogFooter className="pt-2">
            <Button
              type="button"
              variant="outline"
              disabled={isUploading}
              onClick={() => onOpenChange(false)}
              className="border-slate-700 text-slate-300 hover:bg-slate-800"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={isUploading || !selectedFile}
              className="bg-emerald-600 hover:bg-emerald-500 text-white flex items-center gap-2"
            >
              {isUploading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Subiendo documento...
                </>
              ) : (
                <>
                  <UploadCloud className="w-4 h-4" />
                  Subir Documento
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
