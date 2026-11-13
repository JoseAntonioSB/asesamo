'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { apiClient } from '@/lib/api/client';
import { InspectionTemplateDto, InspectionAnswerPayload } from '@/lib/api/types';
import { DynamicInspectionChecklist, InspectionChecklistItem } from '@/components/inspecciones/dynamic-inspection-checklist';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import { ArrowLeft, ClipboardCheck, Sparkles, AlertCircle } from 'lucide-react';

function NuevaInspeccionContent() {
  const router = useRouter();
  const [templates, setTemplates] = useState<InspectionTemplateDto[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('');
  const [areaLocation, setAreaLocation] = useState<string>('Nave 1 - Producción');
  const [inspectionType, setInspectionType] = useState<'routine' | 'planned' | 'special' | 'regulatory'>('routine');
  const [notes, setNotes] = useState<string>('');

  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [activeInspectionId, setActiveInspectionId] = useState<string | null>(null);
  const [checklistItems, setChecklistItems] = useState<InspectionChecklistItem[]>([]);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  useEffect(() => {
    async function loadData() {
      try {
        setIsLoading(true);
        const data = await apiClient.getInspectionTemplates();
        setTemplates(data || []);
        if (data && data.length > 0) {
          setSelectedTemplateId(data[0].id);
        }
      } catch (err: any) {
        toast.error('Error al cargar plantillas', {
          description: err.message || 'No se pudieron consultar los checklists disponibles.',
        });
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  const handleStartInspection = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTemplateId) {
      toast.error('Por favor seleccione una plantilla');
      return;
    }

    try {
      setIsSubmitting(true);
      const res = await apiClient.createInspection({
        template_id: selectedTemplateId,
        inspection_type: inspectionType,
        area_or_location: areaLocation,
        notes,
      });

      setActiveInspectionId(res.id);

      // Cargar ítems preconfigurados de la plantilla
      const currentTmpl = templates.find((t) => t.id === selectedTemplateId);
      if (currentTmpl?.items && currentTmpl.items.length > 0) {
        setChecklistItems(
          currentTmpl.items.map((i) => ({
            id: i.id,
            category: i.category,
            question: i.question,
            guidance_notes: i.guidance_notes,
            weight: i.weight,
            critical_risk: i.critical_risk,
          }))
        );
      } else {
        // Fallback items demo para verificación inmediata en campo
        setChecklistItems([
          {
            id: 'e1111111-1111-1111-1111-111111111111',
            category: 'Protección Contra Incendio (NOM-002)',
            question: '¿Los extintores cuentan con manómetro en rango operable, sello intacto y libre acceso?',
            guidance_notes: 'Verificar distancia mínima de 1 metro libre alrededor del equipo.',
            weight: 3,
            critical_risk: true,
          },
          {
            id: 'e2222222-2222-2222-2222-222222222222',
            category: 'Orden y Limpieza',
            question: '¿Los pasillos de tránsito peatonal y salidas de emergencia están despejados?',
            guidance_notes: 'Sin tarimas, cables ni derrames.',
            weight: 2,
            critical_risk: false,
          },
          {
            id: 'e3333333-3333-3333-3333-333333333333',
            category: 'Equipo de Protección Personal (NOM-017)',
            question: '¿El personal en el área porta casco, gafas de seguridad y calzado con casquillo dieléctrico?',
            guidance_notes: 'Revisión ocular directa en estación de trabajo.',
            weight: 2,
            critical_risk: true,
          },
        ]);
      }

      toast.success('Inspección iniciada en campo', {
        description: `Folio ${res.inspection_number} generado correctamente.`,
      });
    } catch (err: any) {
      toast.error('Error al iniciar inspección', {
        description: err.message || 'No se pudo crear el registro.',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCompleteChecklist = async (answers: InspectionAnswerPayload[]) => {
    if (!activeInspectionId) return;

    try {
      setIsSubmitting(true);
      const res = await apiClient.completeInspection(activeInspectionId, { answers });

      toast.success(`Inspección finalizada con puntaje de ${res.score_percentage}%`, {
        description: res.requires_capa
          ? 'Se han generado planes de acción correctiva (CAPA) automáticamente.'
          : 'Cumplimiento satisfactorio sin observaciones críticas.',
      });

      router.push('/inspecciones');
    } catch (err: any) {
      toast.error('Error al completar inspección', {
        description: err.message || 'No se pudieron registrar las respuestas.',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-emerald-500 selection:text-white py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto space-y-8">
        <div className="flex items-center justify-between">
          <Link
            href="/inspecciones"
            className="inline-flex items-center gap-2 text-xs text-zinc-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Volver al listado de inspecciones
          </Link>
          <span className="text-xs font-mono text-zinc-500 uppercase tracking-wider">Módulo Operativo EHS</span>
        </div>

        {!activeInspectionId ? (
          /* Paso 1: Configurar Datos Generales */
          <div className="bg-zinc-900/80 border border-zinc-800 rounded-2xl p-6 sm:p-8 space-y-6 shadow-xl">
            <div className="space-y-1">
              <h1 className="text-2xl font-bold text-white flex items-center gap-2.5">
                <ClipboardCheck className="w-6 h-6 text-emerald-400" />
                Nueva Inspección de Seguridad
              </h1>
              <p className="text-sm text-zinc-400">
                Seleccione la norma técnica y la zona de planta a auditar.
              </p>
            </div>

            <form onSubmit={handleStartInspection} className="space-y-5">
              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                  Plantilla / Checklist Normativo STPS
                </label>
                {templates.length > 0 ? (
                  <select
                    value={selectedTemplateId}
                    onChange={(e) => setSelectedTemplateId(e.target.value)}
                    className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
                  >
                    {templates.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.title} {t.normative_reference ? `(${t.normative_reference})` : ''}
                      </option>
                    ))}
                  </select>
                ) : (
                  <div className="p-4 rounded-xl bg-zinc-950/80 border border-zinc-800 text-xs text-zinc-400 flex items-center gap-2">
                    <Sparkles className="w-4 h-4 text-emerald-400 shrink-0" />
                    <span>Checklist Maestro STPS General (Extintores NOM-002, Orden y EPP NOM-017)</span>
                  </div>
                )}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                    Tipo de Inspección
                  </label>
                  <select
                    value={inspectionType}
                    onChange={(e) => setInspectionType(e.target.value as any)}
                    className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
                  >
                    <option value="routine">Rutinaria / Operación Diaria</option>
                    <option value="planned">Planeada / Programa Mensual</option>
                    <option value="special">Extraordinaria / Post-incidente</option>
                    <option value="regulatory">Auditoría Normativa STPS</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                    Área o Ubicación de Planta
                  </label>
                  <input
                    type="text"
                    required
                    value={areaLocation}
                    onChange={(e) => setAreaLocation(e.target.value)}
                    placeholder="Ej. Línea 3 de Ensamble, Bodega A"
                    className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                  Notas u Observaciones Iniciales
                </label>
                <textarea
                  rows={2}
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="Alcance particular del recorrido o condiciones meteorológicas..."
                  className="w-full px-4 py-2.5 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
                />
              </div>

              <div className="flex justify-end pt-2">
                <Button
                  type="submit"
                  disabled={isSubmitting}
                  className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold px-6 py-2.5 rounded-xl shadow-md"
                >
                  {isSubmitting ? 'Iniciando...' : 'Comenzar Recorrido en Campo'}
                </Button>
              </div>
            </form>
          </div>
        ) : (
          /* Paso 2: Ejecución del Checklist Táctil */
          <div className="space-y-6">
            <div className="p-4 rounded-xl bg-zinc-900 border border-zinc-800 flex items-center justify-between">
              <div>
                <span className="text-xs text-zinc-400 font-mono">UBICACIÓN: {areaLocation.toUpperCase()}</span>
                <h2 className="text-lg font-bold text-white">Evaluación de Cumplimiento en Tiempo Real</h2>
              </div>
              <span className="text-xs px-2.5 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 font-mono font-bold">
                AUDITORÍA ACTIVA
              </span>
            </div>

            <DynamicInspectionChecklist
              items={checklistItems}
              onComplete={handleCompleteChecklist}
              isSubmitting={isSubmitting}
            />
          </div>
        )}
      </div>
    </div>
  );
}

export default function NuevaInspeccionPage() {
  return (
    <ProtectedRoute>
      <NuevaInspeccionContent />
    </ProtectedRoute>
  );
}
