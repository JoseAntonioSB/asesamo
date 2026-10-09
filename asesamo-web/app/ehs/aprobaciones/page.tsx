'use client';

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { useAuth } from '@/contexts/auth-context';
import { apiClient } from '@/lib/api/client';
import { DocumentItem } from '@/lib/api/types';
import { ReviewTable } from '@/components/ehs/review-table';
import { RejectDialog } from '@/components/ehs/reject-dialog';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import { toast } from 'sonner';
import {
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  Clock,
  XCircle,
  RefreshCw,
  Search,
  Filter,
  FileStack,
  DoorOpen,
  FolderOpen,
  LogOut,
  AlertTriangle,
  Printer,
  QrCode,
} from 'lucide-react';

type FilterTab = 'pending_review' | 'all' | 'approved' | 'rejected';

function EHSApprovalTrayContent() {
  const { user, logout } = useAuth();
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isRefreshing, setIsRefreshing] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<FilterTab>('pending_review');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Processing state for approval / rejection
  const [processingId, setProcessingId] = useState<string | null>(null);
  const [rejectingDoc, setRejectingDoc] = useState<DocumentItem | null>(null);
  const [isRejectModalOpen, setIsRejectModalOpen] = useState<boolean>(false);
  const [isSubmittingReject, setIsSubmittingReject] = useState<boolean>(false);

  const isAuthorizedRole =
    user?.role === 'ehs_manager' ||
    user?.role === 'plant_admin' ||
    user?.role === 'super_admin';

  const loadDocuments = useCallback(
    async (showToast = false) => {
      try {
        if (showToast) setIsRefreshing(true);
        else setIsLoading(true);

        // Fetch documents from backend: if activeTab is 'all', pass undefined
        const statusParam = activeTab === 'all' ? undefined : activeTab;
        const docs = await apiClient.getPendingDocuments(statusParam);
        setDocuments(docs);

        if (showToast) {
          toast.success('Bandeja de aprobación actualizada');
        }
      } catch (err: any) {
        toast.error('Error al cargar bandeja EHS', {
          description: err.message || 'No se pudieron consultar los documentos.',
        });
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
      }
    },
    [activeTab]
  );

  useEffect(() => {
    if (isAuthorizedRole) {
      loadDocuments();
    } else {
      setIsLoading(false);
    }
  }, [loadDocuments, isAuthorizedRole]);

  // Handle direct Approve
  const handleApprove = async (doc: DocumentItem) => {
    try {
      setProcessingId(doc.id);
      const updated = await apiClient.reviewDocument(doc.id, 'approve');

      toast.success('Documento aprobado con éxito', {
        description: `${doc.document_type_name || doc.document_type_code} para ${doc.owner_name || 'titular'} ha sido validado.`,
      });

      // Update local state
      setDocuments((prev) =>
        prev.map((d) => (d.id === doc.id ? { ...d, ...updated, status: 'approved' } : d))
      );
    } catch (err: any) {
      toast.error('Error al aprobar documento', {
        description: err.message || 'No se pudo registrar la aprobación en el servidor.',
      });
    } finally {
      setProcessingId(null);
    }
  };

  // Open Reject Modal
  const handleOpenReject = (doc: DocumentItem) => {
    setRejectingDoc(doc);
    setIsRejectModalOpen(true);
  };

  // Confirm Reject with reason
  const handleConfirmReject = async (reason: string) => {
    if (!rejectingDoc) return;

    try {
      setIsSubmittingReject(true);
      const updated = await apiClient.reviewDocument(rejectingDoc.id, 'reject', reason);

      toast.warning('Documento rechazado', {
        description: `Se notificó el motivo: "${reason}".`,
      });

      // Update local state
      setDocuments((prev) =>
        prev.map((d) =>
          d.id === rejectingDoc.id
            ? { ...d, ...updated, status: 'rejected', rejection_reason: reason }
            : d
        )
      );

      setIsRejectModalOpen(false);
      setRejectingDoc(null);
    } catch (err: any) {
      toast.error('Error al rechazar documento', {
        description: err.message || 'No se pudo guardar el rechazo en el servidor.',
      });
    } finally {
      setIsSubmittingReject(false);
    }
  };

  // Filtered documents by search query
  const filteredDocuments = useMemo(() => {
    if (!searchQuery.trim()) return documents;
    const q = searchQuery.toLowerCase().trim();
    return documents.filter((d) => {
      const matchOwner = d.owner_name?.toLowerCase().includes(q);
      const matchContractor = d.contractor_name?.toLowerCase().includes(q);
      const matchType = d.document_type_name?.toLowerCase().includes(q) || d.document_type_code?.toLowerCase().includes(q);
      const matchFile = d.file_name?.toLowerCase().includes(q);
      return matchOwner || matchContractor || matchType || matchFile;
    });
  }, [documents, searchQuery]);

  // Aggregate counters
  const counters = useMemo(() => {
    const pending = documents.filter((d) => d.status === 'pending_review').length;
    const approved = documents.filter((d) => d.status === 'approved').length;
    const rejected = documents.filter((d) => d.status === 'rejected').length;
    return { pending, approved, rejected, total: documents.length };
  }, [documents]);

  if (!isAuthorizedRole) {
    return (
      <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
        <header className="border-b border-slate-800 bg-slate-900/90 h-16 flex items-center px-6 justify-between">
          <div className="flex items-center gap-2">
            <span className="font-mono font-bold text-white uppercase">ASESAMO</span>
          </div>
          <Button variant="ghost" size="sm" onClick={logout} className="text-slate-400">
            <LogOut className="w-4 h-4" />
          </Button>
        </header>
        <div className="flex-1 flex items-center justify-center p-6">
          <Card className="max-w-md w-full bg-slate-900 border-slate-800 text-center p-6 space-y-4">
            <div className="w-12 h-12 rounded-full bg-rose-500/10 border border-rose-500/20 text-rose-400 mx-auto flex items-center justify-center">
              <ShieldAlert className="w-6 h-6" />
            </div>
            <h2 className="text-lg font-bold text-white">Acceso Restringido</h2>
            <p className="text-xs text-slate-400">
              La Bandeja de Aprobación y Dictamen EHS está reservada para personal con rol de Responsable de Seguridad (EHS Manager) o Administrador de Planta.
            </p>
            <div className="pt-2 flex justify-center gap-3">
              <Link href="/caseta">
                <Button size="sm" variant="outline" className="text-xs border-slate-700">
                  Ir a Caseta
                </Button>
              </Link>
              <Link href="/documentos">
                <Button size="sm" className="text-xs bg-emerald-600 hover:bg-emerald-500">
                  Mi Expediente
                </Button>
              </Link>
            </div>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col selection:bg-emerald-500 selection:text-white">
      {/* Top Navigation */}
      <header className="border-b border-slate-800 bg-slate-900/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-mono text-base font-bold tracking-tight text-white uppercase">
                  ASESAMO
                </span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 font-mono border border-amber-500/30">
                  Bandeja EHS
                </span>
              </div>
              <p className="text-xs text-slate-400">Dictamen y Aprobación de Documentos de Seguridad</p>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            <Link href="/caseta">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Control de Accesos en Caseta"
              >
                <DoorOpen className="w-3.5 h-3.5 text-cyan-400" />
                <span className="hidden sm:inline">Caseta</span>
              </Button>
            </Link>

            <Link href="/documentos">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-800 text-slate-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Ver Expediente"
              >
                <FolderOpen className="w-3.5 h-3.5 text-emerald-400" />
                <span className="hidden sm:inline">Expediente</span>
              </Button>
            </Link>

            <Link href="/credencial">
              <Button
                size="sm"
                variant="outline"
                className="border-slate-800 text-cyan-300 hover:text-white hover:bg-slate-800 text-xs gap-1.5"
                title="Ver Credencial Móvil del Trabajador"
              >
                <QrCode className="w-3.5 h-3.5 text-cyan-400" />
                <span className="hidden sm:inline">Credencial</span>
              </Button>
            </Link>

            {/* Theme Toggle */}
            <ThemeToggle />

            <div className="hidden md:block text-right pl-2 border-l border-slate-800">
              <p className="text-xs font-medium text-white">{user?.full_name || user?.email}</p>
              <p className="text-[11px] text-amber-400 font-mono uppercase">{user?.role}</p>
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

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Header Title & Refresh */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight flex items-center gap-2.5">
              <FileStack className="w-7 h-7 text-amber-400" />
              Dictamen y Validación de Cumplimiento
            </h1>
            <p className="text-sm text-slate-400 mt-1">
              Revise constancias DC-3, afiliación IMSS, SUA, registros REPSE y certificados de aptitud médica enviados por contratistas.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => window.print()}
              className="border-slate-800 text-slate-300 hover:text-white hover:bg-slate-900 h-10 px-3.5 gap-1.5 no-print"
              title="Imprimir reporte para auditoría"
            >
              <Printer className="w-4 h-4 text-slate-400" />
              <span>Imprimir Ficha</span>
            </Button>

            <Button
              size="sm"
              variant="outline"
              disabled={isLoading || isRefreshing}
              onClick={() => loadDocuments(true)}
              className="border-slate-800 text-slate-300 hover:bg-slate-900 h-10 px-3.5 gap-1.5 no-print"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span>Actualizar Bandeja</span>
            </Button>
          </div>
        </div>

        {/* Aggregate Counters */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4" role="region" aria-label="Resumen de auditoría EHS">
          <Card className="bg-slate-900 border-slate-800 shadow-md">
            <CardContent className="p-4 flex items-center justify-between" role="status" aria-atomic="true">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-amber-400">
                  Pendientes de Revisión
                </p>
                <p className="text-2xl font-bold font-mono text-amber-400 mt-1">
                  {counters.pending}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-amber-950/60 border border-amber-900/40 text-amber-400">
                <Clock className="w-5 h-5" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-slate-900 border-slate-800 shadow-md">
            <CardContent className="p-4 flex items-center justify-between" role="status" aria-atomic="true">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-emerald-400">
                  Aprobados
                </p>
                <p className="text-2xl font-bold font-mono text-emerald-400 mt-1">
                  {counters.approved}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-emerald-950/60 border border-emerald-900/40 text-emerald-400">
                <CheckCircle2 className="w-5 h-5" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-slate-900 border-slate-800 shadow-md">
            <CardContent className="p-4 flex items-center justify-between" role="status" aria-atomic="true">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-rose-400">
                  Rechazados
                </p>
                <p className="text-2xl font-bold font-mono text-rose-400 mt-1">
                  {counters.rejected}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-rose-950/60 border border-rose-900/40 text-rose-400">
                <XCircle className="w-5 h-5" />
              </div>
            </CardContent>
          </Card>

          <Card className="bg-slate-900 border-slate-800 shadow-md">
            <CardContent className="p-4 flex items-center justify-between" role="status" aria-atomic="true">
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-slate-300">
                  Total en Vista
                </p>
                <p className="text-2xl font-bold font-mono text-white mt-1">
                  {counters.total}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-slate-800 text-slate-300">
                <FileStack className="w-5 h-5 text-sky-400" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Filter Tabs & Search Bar */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-2">
          {/* Status Tabs */}
          <div className="flex items-center gap-1.5 p-1 bg-slate-900 border border-slate-800 rounded-xl overflow-x-auto">
            <button
              onClick={() => setActiveTab('pending_review')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
                activeTab === 'pending_review'
                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Clock className="w-3.5 h-3.5" />
              Por Revisar
            </button>

            <button
              onClick={() => setActiveTab('all')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
                activeTab === 'all'
                  ? 'bg-slate-800 text-white border border-slate-700 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Filter className="w-3.5 h-3.5" />
              Todos
            </button>

            <button
              onClick={() => setActiveTab('approved')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
                activeTab === 'approved'
                  ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <CheckCircle2 className="w-3.5 h-3.5" />
              Aprobados
            </button>

            <button
              onClick={() => setActiveTab('rejected')}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all flex items-center gap-1.5 whitespace-nowrap ${
                activeTab === 'rejected'
                  ? 'bg-rose-500/20 text-rose-300 border border-rose-500/40 shadow-sm'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <XCircle className="w-3.5 h-3.5" />
              Rechazados
            </button>
          </div>

          {/* Search Bar */}
          <div className="relative w-full sm:w-72">
            <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
            <Input
              type="text"
              placeholder="Buscar por titular, contratista o tipo..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 h-10 bg-slate-900 border-slate-800 text-xs text-white placeholder-slate-500 rounded-xl focus:border-amber-500"
            />
          </div>
        </div>

        {/* Review Table */}
        <ReviewTable
          documents={filteredDocuments}
          onApprove={handleApprove}
          onReject={handleOpenReject}
          processingId={processingId}
          isLoading={isLoading}
        />
      </main>

      {/* Reject Modal */}
      <RejectDialog
        isOpen={isRejectModalOpen}
        onClose={() => setIsRejectModalOpen(false)}
        documentItem={rejectingDoc}
        onConfirmReject={handleConfirmReject}
        isSubmitting={isSubmittingReject}
      />
    </div>
  );
}

export default function EHSApprovalPage() {
  return (
    <ProtectedRoute>
      <EHSApprovalTrayContent />
    </ProtectedRoute>
  );
}
