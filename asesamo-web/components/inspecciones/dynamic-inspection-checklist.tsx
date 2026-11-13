'use client';

import React, { useState } from 'react';
import { InspectionAnswerPayload, ItemComplianceStatus } from '@/lib/api/types';
import { calculateInspectionScore, InspectionScoringItem } from '@/lib/inspection-scoring';
import { CheckCircle2, XCircle, MinusCircle, AlertTriangle, Camera } from 'lucide-react';

export interface InspectionChecklistItem extends InspectionScoringItem {
  category: string;
  question: string;
  guidance_notes?: string;
}

interface DynamicInspectionChecklistProps {
  items: InspectionChecklistItem[];
  onComplete: (answers: InspectionAnswerPayload[]) => void;
  isSubmitting?: boolean;
}

export function DynamicInspectionChecklist({
  items,
  onComplete,
  isSubmitting = false,
}: DynamicInspectionChecklistProps) {
  const [answers, setAnswers] = useState<Record<string, { status: ItemComplianceStatus; notes?: string; photo?: string }>>({});

  const handleStatusChange = (itemId: string, status: ItemComplianceStatus) => {
    setAnswers((prev) => ({
      ...prev,
      [itemId]: {
        ...prev[itemId],
        status,
      },
    }));
  };

  const handleNotesChange = (itemId: string, notes: string) => {
    setAnswers((prev) => ({
      ...prev,
      [itemId]: {
        ...prev[itemId],
        notes,
      },
    }));
  };

  // Convertir a array de answers para scoring en tiempo real
  const answersList: InspectionAnswerPayload[] = items.map((itm) => {
    const ans = answers[itm.id];
    return {
      template_item_id: itm.id,
      status: ans?.status || 'compliant',
      finding_notes: ans?.notes,
      evidence_photo_url: ans?.photo,
    };
  });

  const scoreSummary = calculateInspectionScore(items, answersList);
  const allAnswered = items.every((itm) => !!answers[itm.id]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onComplete(answersList);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Barra de Score en Vivo */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-sm">
        <div>
          <span className="text-xs uppercase tracking-wider text-zinc-400 font-medium">Índice de Seguridad en Tiempo Real</span>
          <div className="flex items-baseline gap-2 mt-0.5">
            <span
              className={`text-3xl font-extrabold ${
                scoreSummary.scorePercentage >= 85
                  ? 'text-emerald-400'
                  : scoreSummary.scorePercentage >= 70
                  ? 'text-amber-400'
                  : 'text-red-400'
              }`}
            >
              {scoreSummary.scorePercentage.toFixed(1)}%
            </span>
            <span className="text-xs text-zinc-500">
              ({scoreSummary.compliantCount} conformes / {scoreSummary.nonCompliantCount} hallazgos / {scoreSummary.naCount} N/A)
            </span>
          </div>
        </div>

        {scoreSummary.hasCriticalFailure && (
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-semibold">
            <AlertTriangle className="w-4 h-4 text-red-400 animate-pulse" />
            <span>Alerta: Requiere Plan de Acción Inmediato (CAPA)</span>
          </div>
        )}
      </div>

      {/* Lista de Ítems */}
      <div className="space-y-4">
        {items.map((item, idx) => {
          const currentStatus = answers[item.id]?.status;
          const isNonCompliant = currentStatus === 'non_compliant';

          return (
            <div
              key={item.id}
              className={`p-4 rounded-xl border transition-all ${
                isNonCompliant
                  ? 'bg-red-950/20 border-red-500/40 shadow-sm'
                  : 'bg-zinc-950/60 border-zinc-800/80 hover:border-zinc-700'
              }`}
            >
              <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-zinc-500">#{idx + 1}</span>
                    <span className="text-xs font-medium px-2 py-0.5 rounded bg-zinc-800/80 text-zinc-300">
                      {item.category}
                    </span>
                    {item.critical_risk && (
                      <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-red-500/20 text-red-400 border border-red-500/30">
                        Riesgo Crítico
                      </span>
                    )}
                    <span className="text-[10px] text-zinc-500">Peso: {item.weight}x</span>
                  </div>
                  <p className="text-sm font-semibold text-zinc-200 mt-1">{item.question}</p>
                  {item.guidance_notes && (
                    <p className="text-xs text-zinc-400 italic mt-0.5">{item.guidance_notes}</p>
                  )}
                </div>

                {/* Botones de Selección Táctil */}
                <div className="flex items-center gap-1.5 self-end sm:self-center shrink-0">
                  <button
                    type="button"
                    onClick={() => handleStatusChange(item.id, 'compliant')}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                      currentStatus === 'compliant'
                        ? 'bg-emerald-600 text-white shadow'
                        : 'bg-zinc-900 text-zinc-400 hover:text-zinc-200 border border-zinc-800'
                    }`}
                  >
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    Conforme
                  </button>

                  <button
                    type="button"
                    onClick={() => handleStatusChange(item.id, 'non_compliant')}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                      currentStatus === 'non_compliant'
                        ? 'bg-red-600 text-white shadow'
                        : 'bg-zinc-900 text-zinc-400 hover:text-zinc-200 border border-zinc-800'
                    }`}
                  >
                    <XCircle className="w-3.5 h-3.5" />
                    No Conforme
                  </button>

                  <button
                    type="button"
                    onClick={() => handleStatusChange(item.id, 'not_applicable')}
                    className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                      currentStatus === 'not_applicable'
                        ? 'bg-zinc-700 text-zinc-100 shadow'
                        : 'bg-zinc-900 text-zinc-400 hover:text-zinc-200 border border-zinc-800'
                    }`}
                  >
                    <MinusCircle className="w-3.5 h-3.5" />
                    N/A
                  </button>
                </div>
              </div>

              {/* Si es No Conforme, desplegar captura de hallazgo y evidencia */}
              {isNonCompliant && (
                <div className="mt-3 pt-3 border-t border-red-500/20 space-y-2">
                  <label className="block text-xs font-semibold text-red-300">
                    Descripción del Hallazgo / Desvío (Obligatorio para CAPA):
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Ej. Extintor con manómetro en zona roja o sin sello de inspección"
                    value={answers[item.id]?.notes || ''}
                    onChange={(e) => handleNotesChange(item.id, e.target.value)}
                    className="w-full px-3 py-1.5 text-xs bg-zinc-900/90 border border-red-500/40 rounded-lg text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-1 focus:ring-red-500"
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Botón de Finalización */}
      <div className="flex justify-end pt-4">
        <button
          type="submit"
          disabled={!allAnswered || isSubmitting}
          className="px-6 py-2.5 rounded-xl text-sm font-semibold text-white bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-md flex items-center gap-2"
        >
          {isSubmitting ? 'Finalizando y Generando CAPA...' : 'Finalizar Inspección y Calcular Score'}
        </button>
      </div>
    </form>
  );
}
