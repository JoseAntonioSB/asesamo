'use client';

import React, { useState, useRef } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { ContractorWorkerItem } from '@/lib/api/types';
import { apiClient, ApiError } from '@/lib/api/client';
import { toast } from 'sonner';
import {
  Camera,
  UploadCloud,
  Loader2,
  HardHat,
  CheckCircle2,
  X,
  AlertCircle,
  FileImage,
} from 'lucide-react';

interface WorkerPhotoModalProps {
  isOpen: boolean;
  onClose: () => void;
  worker: ContractorWorkerItem | null;
  onPhotoUploaded: (workerId: string) => void;
}

export function WorkerPhotoModal({
  isOpen,
  onClose,
  worker,
  onPhotoUploaded,
}: WorkerPhotoModalProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const [dragActive, setDragActive] = useState<boolean>(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  if (!worker) return null;

  const handleClose = () => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setIsUploading(false);
    onClose();
  };

  const handleFile = (file: File) => {
    // Max 5 MB
    if (file.size > 5 * 1024 * 1024) {
      toast.error('La fotografía excede el tamaño máximo permitido de 5 MB.');
      return;
    }

    const validTypes = ['image/jpeg', 'image/png', 'image/webp'];
    if (!validTypes.includes(file.type)) {
      toast.error('Formato no válido. Solo se permiten imágenes JPEG, PNG o WebP.');
      return;
    }

    setSelectedFile(file);
    setPreviewUrl(URL.createObjectURL(file));
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  const handleSubmit = async () => {
    if (!selectedFile) {
      toast.error('Seleccione una fotografía primero.');
      return;
    }

    try {
      setIsUploading(true);
      await apiClient.uploadUserAvatar(worker.worker_id, selectedFile);
      toast.success(`Fotografía de ${worker.full_name} actualizada con éxito.`);
      onPhotoUploaded(worker.worker_id);
      handleClose();
    } catch (err) {
      const apiErr = err as ApiError;
      toast.error(apiErr.message || 'Error al subir la fotografía del trabajador.');
    } finally {
      setIsUploading(false);
    }
  };

  const currentAvatarUrl = `${apiClient.getUserAvatarUrl(worker.worker_id)}?t=${Date.now()}`;
  const initials = worker.full_name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((n) => n[0])
    .join('')
    .toUpperCase();

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogContent className="sm:max-w-md bg-slate-900 border-slate-800 text-slate-100 p-6">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold text-white flex items-center gap-2">
            <Camera className="w-5 h-5 text-emerald-400" />
            <span>Fotografía del Trabajador</span>
          </DialogTitle>
          <DialogDescription className="text-xs text-slate-400">
            Actualice la fotografía oficial de credencial e identificación en caseta para este operario.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5 pt-2">
          {/* Worker identity preview */}
          <div className="flex items-center gap-3.5 p-3.5 rounded-xl bg-slate-950/80 border border-slate-800">
            <div className="w-14 h-14 rounded-2xl bg-slate-800 border-2 border-slate-700 flex items-center justify-center font-bold text-slate-200 overflow-hidden shrink-0 relative">
              {previewUrl ? (
                <img src={previewUrl} alt="Preview" className="w-full h-full object-cover" />
              ) : (
                <>
                  <img
                    src={currentAvatarUrl}
                    alt={worker.full_name}
                    className="w-full h-full object-cover"
                    onError={(e) => {
                      (e.target as HTMLElement).style.display = 'none';
                    }}
                  />
                  <span className="text-sm font-black">{initials}</span>
                </>
              )}
            </div>
            <div className="min-w-0 flex-1">
              <h4 className="text-sm font-bold text-white truncate">{worker.full_name}</h4>
              <p className="text-xs text-emerald-400 font-medium truncate">{worker.job_title || 'Operario'}</p>
              <p className="text-[11px] text-slate-400 font-mono truncate">{worker.email}</p>
            </div>
          </div>

          {/* Upload Drop Zone */}
          <div
            onDragEnter={handleDrag}
            onDragOver={handleDrag}
            onDragLeave={handleDrag}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`border-2 border-dashed rounded-2xl p-6 text-center cursor-pointer transition-all flex flex-col items-center justify-center gap-2.5 ${
              dragActive
                ? 'border-emerald-500 bg-emerald-950/20 text-emerald-300'
                : 'border-slate-800 hover:border-slate-700 bg-slate-950/50 hover:bg-slate-950 text-slate-400'
            }`}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp"
              className="hidden"
              onChange={(e) => {
                if (e.target.files && e.target.files[0]) {
                  handleFile(e.target.files[0]);
                }
              }}
            />

            <div className="w-12 h-12 rounded-xl bg-emerald-500/10 text-emerald-400 flex items-center justify-center shadow-inner">
              <UploadCloud className="w-6 h-6" />
            </div>

            <div className="space-y-0.5">
              <p className="text-xs font-semibold text-white">
                {selectedFile ? selectedFile.name : 'Haga clic para seleccionar o arrastre una imagen'}
              </p>
              <p className="text-[11px] text-slate-500">
                JPEG, PNG o WebP (Máx. 5 MB). Recomendado relación 1:1.
              </p>
            </div>
          </div>

          {/* Dialog Action Buttons */}
          <div className="flex items-center justify-end gap-2.5 pt-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleClose}
              disabled={isUploading}
              className="border-slate-800 text-slate-400 hover:text-white hover:bg-slate-800"
            >
              Cancelar
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={!selectedFile || isUploading}
              onClick={handleSubmit}
              className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold shadow-md shadow-emerald-600/20 gap-2"
            >
              {isUploading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>Subiendo...</span>
                </>
              ) : (
                <>
                  <Camera className="w-4 h-4" />
                  <span>Guardar Fotografía</span>
                </>
              )}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
