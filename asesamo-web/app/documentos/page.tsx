'use client';

import React, { useState, useEffect, useCallback, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { DocumentItem, DocumentTypeItem } from '@/lib/api/types';
import { DocumentUploadModal, WorkerOption } from '@/components/documentos/document-upload-modal';
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
  Building2,
  BarChart3,
  UserCheck,
  Users,
} from 'lucide-react';

function DocumentosContent() {
  const { user, logout } = useAuth();
  const searchParams = useSearchParams();
  const workerIdFilter = searchParams.get('worker_id') || undefined;

  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [documentTypes, setDocumentTypes] = useState<DocumentTypeItem[]>([]);
  const [workers, setWorkers] = useState<WorkerOption[]>([]);
  const [selectedWorker, setSelectedWorker] = useState<WorkerOption | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [uploadModalOpen, setUploadModalOpen] = useState<boolean>(false);

  const isContractorAdmin = user?.role === 'contractor_admin' || user?.role === 'contractor';

  const loadData = useCallback(async (showToast = false) => {
    try {
      if (showToast) setIsRefreshing(true);
      else setIsLoading(true);

      const promises: [Promise<DocumentItem[]>, Promise<DocumentTypeItem[]>] = [
        apiClient.getMyDocuments(workerIdFilter ? { worker_id: workerIdFilter } : undefined),
        apiClient.getDocumentTypes(),
      ];

      const [docs, types] = await Promise.all(promises);
      setDocuments(docs);
      setDocumentTypes(types);

      // If user is contractor admin, fetch their crew to populate worker options
      if (isContractorAdmin) {
        try {
          const company = await apiClient.getMyContractorCompany();
          if (company?.id) {
            const crew = await apiClient.listContractorWorkers(company.id);
            const formattedWorkers: WorkerOption[] = crew.map((w) => ({
              id: w.worker_id,
              name: w.full_name,
              curp: w.curp,
              jobTitle: w.job_title,
            }));
            setWorkers(formattedWorkers);

            if (workerIdFilter) {
              const matched = formattedWorkers.find((w) => w.id === workerIdFilter);
              setSelectedWorker(matched || null);
            } else {
              setSelectedWorker(null);
            }
          }
        } catch {
          // Non-blocking if company query fails
        }
      }

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
  }, [workerIdFilter, isContractorAdmin]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleUploadSuccess = (newDoc: DocumentItem) => {
    setDocuments((prev) => [newDoc, ...prev]);
  };

  const isGuard = user?.role === 'guard';
  const isPlantAdmin =
    user?.role === 'plant_admin' ||
    user?.role === 'tenant_admin' ||
    user?.role === 'super_admin';
  const isEhsManager = user?.role === 'ehs_manager';
  const isEhsOrAdmin = isPlantAdmin || isEhsManager;

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
            <ThemeToggle />

            {isGuard && (
              <Link href="/caseta">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-cyan-500/30 text-cyan-400 hover:bg-cyan-950/30 h-9 px-3 gap-1.5"
                >
                  <QrCode className="w-4 h-4" />
                  <span className="hidden sm:inline">Caseta</span>
                </Button>
              </Link>
            )}

            {isEhsOrAdmin && (
              <Link href="/ehs/aprobaciones">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-emerald-500/30 text-emerald-400 hover:bg-emerald-950/30 h-9 px-3 gap-1.5"
                >
                  <FileStack className="w-4 h-4" />
                  <span className="hidden sm:inline">Mesa EHS</span>
                </Button>
              </Link>
            )}

            <Link href="/contratistas">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 h-9 px-3 gap-1.5"
              >
                <Building2 className="w-4 h-4" />
                <span className="hidden sm:inline">Contratistas</span>
              </Button>
            </Link>

            <Link href="/perfil" className="flex items-center gap-2 p-1 rounded-lg hover:bg-slate-800/60 transition-colors">
              <div className="w-8 h-8 rounded-full bg-emerald-500/20 border border-emerald-500/40 text-emerald-400 flex items-center justify-center font-bold text-xs shrink-0 overflow-hidden">
                {user?.id && (
                  <img
                    src={apiClient.getUserAvatarUrl(user.id)}
                    alt=""
                    className="w-full h-full object-cover"
                    onError={(e) => {
                      (e.target as HTMLElement).style.display = 'none';
                    }}
                  />
                )}
                <span>{(user?.full_name || 'US').slice(0, 2).toUpperCase()}</span>
              </div>
              <div className="hidden sm:block text-left">
                <p className="text-xs font-medium text-white leading-tight max-w-[120px] truncate">{user?.full_name || user?.email}</p>
                <p className="text-[10px] text-emerald-400 font-mono uppercase leading-tight">{user?.role}</p>
              </div>
            </Link>

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
        {/* Active Worker Filter Banner */}
        {workerIdFilter && (
          <div className="p-4 rounded-xl border border-emerald-800/40 bg-emerald-950/20 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-lg">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shrink-0">
                <UserCheck className="w-5 h-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-white text-sm">
                    Expediente del Trabajador: {selectedWorker ? selectedWorker.name : 'Personal Asignado'}
                  </h3>
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 font-mono border border-emerald-500/20">
                    Filtro Individual
                  </span>
                </div>
                <p className="text-xs text-slate-400 mt-0.5">
                  {selectedWorker?.jobTitle && `${selectedWorker.jobTitle} • `}
                  {selectedWorker?.curp && `CURP: ${selectedWorker.curp}`}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Link href="/documentos">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs h-8"
                >
                  Ver todos los documentos
                </Button>
              </Link>
              <Link href="/contratistas">
                <Button
                  size="sm"
                  variant="outline"
                  className="border-slate-700 text-slate-300 hover:text-white hover:bg-slate-800 text-xs h-8 gap-1"
                >
                  <Users className="w-3.5 h-3.5" />
                  Cuadrilla
                </Button>
              </Link>
            </div>
          </div>
        )}

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
        defaultWorkerId={workerIdFilter}
        workers={workers}
      />
    </div>
  );
}

export default function DocumentosPage() {
  return (
    <ProtectedRoute>
      <Suspense
        fallback={
          <div className="min-h-screen bg-slate-950 flex items-center justify-center text-slate-400 font-mono text-xs">
            Cargando expediente documental...
          </div>
        }
      >
        <DocumentosContent />
      </Suspense>
    </ProtectedRoute>
  );
}
