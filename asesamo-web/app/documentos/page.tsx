'use client';

import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { DocumentItem, DocumentTypeItem } from '@/lib/api/types';
import { DocumentUploadModal } from '@/components/documentos/document-upload-modal';
import { DocumentListTable } from '@/components/documentos/document-list-table';
import { DocumentSummaryCards } from '@/components/documentos/document-summary-cards';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ShieldCheck,
  UploadCloud,
  RefreshCw,
  LogOut,
  ArrowLeft,
  FileCheck2,
  FileStack,
  QrCode,
} from 'lucide-react';

function DocumentosContent() {
  const { user, logout } = useAuth();
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [documentTypes, setDocumentTypes] = useState<DocumentTypeItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [uploadModalOpen, setUploadModalOpen] = useState<boolean>(false);

  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const [docs, types] = await Promise.all([
        apiClient.getMyDocuments(),
        apiClient.getDocumentTypes(),
      ]);

      setDocuments(docs);
      setDocumentTypes(types);

      if (showToast) {
        toast.success('Lista de documentos actualizada');
      }
    } catch {
      toast.error('Error al cargar documentos', {
        description: 'No se pudo obtener la lista de documentos del servidor.',
      });
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleUploadSuccess = (newDoc: DocumentItem) => {
    setDocuments((prev) => [newDoc, ...prev]);
  };

  const isGuardOrAdmin =
    user?.role === 'guard' ||
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin' ||
    user?.role === 'ehs_manager';

  const isEhsOrAdmin =
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin' ||
    user?.role === 'ehs_manager';

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-emerald-500 selection:text-white">
      {/* Top Navbar */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 font-mono border border-slate-700">
                  Cumplimiento
                </span>
              </div>
              <p className="text-xs text-slate-400">Gestión de Documentos Laborales y Seguridad</p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-4">
            {isGuardOrAdmin && (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                >
                  <ArrowLeft className="w-3.5 h-3.5" />
                  Volver a Caseta
                </Button>
              </Link>
            )}

            {isEhsOrAdmin && (
              <Link href="/ehs/aprobaciones">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-amber-800/60 bg-amber-950/30 text-amber-300 hover:text-white hover:bg-amber-900/50 text-xs gap-1.5"
                >
                  <FileStack className="w-3.5 h-3.5 text-amber-400" />
                  Bandeja EHS
                </Button>
              </Link>
            )}

            <Link href="/credencial">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-cyan-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Ver Credencial Móvil con Código QR"
              >
                <QrCode className="w-3.5 h-3.5 text-cyan-400" />
                <span className="hidden sm:inline">Mi Credencial</span>
              </Button>
            </Link>

            {/* Theme Toggle */}
            <ThemeToggle />

            <div className="hidden sm:block text-right">
              <p className="text-xs font-medium text-white">{user?.full_name || user?.email}</p>
              <p className="text-[11px] text-emerald-400 font-mono uppercase">{user?.role}</p>
            </div>

            <Button
              size="sm"
              variant="ghost"
              onClick={logout}
              className="text-slate-400 hover:text-rose-400 hover:bg-rose-950/20 h-9 px-2.5"
              title="Cerrar sesión"
            >
              <LogOut className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Page Title & Actions */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight flex items-center gap-2.5">
              <FileCheck2 className="w-7 h-7 text-emerald-400" />
              Expediente Documental
            </h1>
            <p className="text-sm text-slate-400 mt-1">
              Verifique el estado de constancias DC-3, afiliación IMSS, REPSE y dictámenes de aptitud.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              disabled={isLoading || isRefreshing}
              onClick={() => loadData(true)}
              className="border-slate-800 text-slate-300 hover:bg-slate-900 h-10 px-3 gap-1.5"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span className="hidden sm:inline">Actualizar</span>
            </Button>

            <Button
              size="sm"
              onClick={() => setUploadModalOpen(true)}
              className="bg-emerald-600 hover:bg-emerald-500 text-white h-10 px-4 font-semibold shadow-lg shadow-emerald-950/50 gap-2"
            >
              <UploadCloud className="w-4 h-4" />
              Subir Documento
            </Button>
          </div>
        </div>

        {/* Metrics Summary Cards */}
        <DocumentSummaryCards documents={documents} requiredTypes={documentTypes} />

        {/* Document List Table */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-300">
              Documentos Registrados ({documents.length})
            </h2>
            <span className="text-xs text-slate-500 font-mono">
              Revisión oficial por Seguridad e Higiene (EHS)
            </span>
          </div>

          <DocumentListTable
            documents={documents}
            isLoading={isLoading}
            onOpenUpload={() => setUploadModalOpen(true)}
          />
        </div>
      </main>

      {/* Upload Dialog Modal */}
      <DocumentUploadModal
        open={uploadModalOpen}
        onOpenChange={setUploadModalOpen}
        documentTypes={documentTypes}
        onUploadSuccess={handleUploadSuccess}
      />
    </div>
  );
}

export default function DocumentosPage() {
  return (
    <ProtectedRoute>
      <DocumentosContent />
    </ProtectedRoute>
  );
}
