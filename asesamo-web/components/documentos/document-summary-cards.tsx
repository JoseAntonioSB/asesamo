'use client';

import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { DocumentItem, DocumentTypeItem } from '@/lib/api/types';
import { ShieldCheck, Clock, AlertTriangle, FileCheck } from 'lucide-react';

interface DocumentSummaryCardsProps {
  documents: DocumentItem[];
  requiredTypes: DocumentTypeItem[];
}

export function DocumentSummaryCards({ documents, requiredTypes }: DocumentSummaryCardsProps) {
  const approvedCount = documents.filter((d) => d.status === 'approved').length;
  const pendingCount = documents.filter((d) => d.status === 'pending_review').length;
  const rejectedOrExpiredCount = documents.filter(
    (d) => d.status === 'rejected' || d.status === 'expired'
  ).length;

  // Track which required types are covered
  const uploadedTypeCodes = new Set(documents.map((d) => d.document_type_code));
  const missingCount = requiredTypes.filter(
    (rt) => rt.is_required_default && !uploadedTypeCodes.has(rt.code)
  ).length;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card className="bg-slate-900 border-slate-800 shadow-md">
        <CardContent className="p-4 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Registrados</p>
            <p className="text-2xl font-bold font-mono text-white mt-1">{documents.length}</p>
          </div>
          <div className="p-3 rounded-xl bg-slate-800 text-slate-300">
            <FileCheck className="w-5 h-5 text-sky-400" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-slate-900 border-slate-800 shadow-md">
        <CardContent className="p-4 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-emerald-400">Aprobados (Vigentes)</p>
            <p className="text-2xl font-bold font-mono text-emerald-400 mt-1">{approvedCount}</p>
          </div>
          <div className="p-3 rounded-xl bg-emerald-950/60 border border-emerald-900/40 text-emerald-400">
            <ShieldCheck className="w-5 h-5" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-slate-900 border-slate-800 shadow-md">
        <CardContent className="p-4 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-amber-400">En Revisión</p>
            <p className="text-2xl font-bold font-mono text-amber-400 mt-1">{pendingCount}</p>
          </div>
          <div className="p-3 rounded-xl bg-amber-950/60 border border-amber-900/40 text-amber-400">
            <Clock className="w-5 h-5" />
          </div>
        </CardContent>
      </Card>

      <Card className="bg-slate-900 border-slate-800 shadow-md">
        <CardContent className="p-4 flex items-center justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-rose-400">Faltantes o Vencidos</p>
            <p className="text-2xl font-bold font-mono text-rose-400 mt-1">
              {missingCount + rejectedOrExpiredCount}
            </p>
          </div>
          <div className="p-3 rounded-xl bg-rose-950/60 border border-rose-900/40 text-rose-400">
            <AlertTriangle className="w-5 h-5" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
