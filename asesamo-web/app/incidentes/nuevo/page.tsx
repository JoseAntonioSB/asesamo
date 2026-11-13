'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { apiClient } from '@/lib/api/client';
import { IncidentType, IncidentSeverity } from '@/lib/api/types';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';
import { ArrowLeft, AlertOctagon, Camera, AlertTriangle, ShieldAlert } from 'lucide-react';

function ReportarIncidenteContent() {
  const router = useRouter();
  const [type, setType] = useState<IncidentType>('near_miss');
  const [severity, setSeverity] = useState<IncidentSeverity>('moderate');
  const [title, setTitle] = useState<string>('');
  const [description, setDescription] = useState<string>('');
  const [location, setLocation] = useState<string>('');
  const [photoUrl, setPhotoUrl] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim() || !description.trim() || !location.trim()) {
      toast.error('Complete todos los campos obligatorios');
      return;
    }

    try {
      setIsSubmitting(true);
      const res = await apiClient.reportIncident({
        type,
        severity,
        title: title.trim(),
        description: description.trim(),
        location: location.trim(),
        evidence_photo_url: photoUrl.trim() || undefined,
      });

      toast.success('Incidente registrado exitosamente', {
        description: `Folio ${res.incident_number} generado para investigación de causa raíz.`,
      });

      router.push('/incidentes');
    } catch (err: any) {
      toast.error('Error al reportar incidente', {
        description: err.message || 'No se pudo guardar el reporte.',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 selection:bg-emerald-500 selection:text-white py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto space-y-8">
        <div className="flex items-center justify-between">
          <Link
            href="/incidentes"
            className="inline-flex items-center gap-2 text-xs text-zinc-400 hover:text-white transition-colors"
          >
            <ArrowLeft className="w-4 h-4" />
            Volver a la bitácora de incidentes
          </Link>
          <span className="text-xs font-mono text-zinc-500 uppercase tracking-wider">Protocolo NOM-021-STPS</span>
        </div>

        <div className="bg-zinc-900/80 border border-zinc-800 rounded-2xl p-6 sm:p-8 space-y-6 shadow-xl">
          <div className="space-y-1">
            <h1 className="text-2xl font-bold text-white flex items-center gap-2.5">
              <AlertOctagon className="w-6 h-6 text-red-400" />
              Notificación Inmediata de Incidente / Near-Miss
            </h1>
            <p className="text-sm text-zinc-400">
              Reporte de actos inseguros, desvíos críticos o cuasi-accidentes para investigación y acción correctiva.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                  Clasificación del Evento
                </label>
                <select
                  value={type}
                  onChange={(e) => setType(e.target.value as IncidentType)}
                  className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
                >
                  <option value="near_miss">Cuasi-accidente (Near Miss)</option>
                  <option value="unsafe_condition">Condición Insegura en Instalaciones</option>
                  <option value="unsafe_act">Acto Inseguro de Personal / Contratista</option>
                  <option value="first_aid">Primeros Auxilios (Sin incapacidad)</option>
                  <option value="medical_treatment">Atención Médica Externa</option>
                  <option value="lost_time_injury">Accidente con Tiempo Perdido (LTI)</option>
                </select>
              </div>

              <div className="space-y-1.5">
                <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                  Severidad Estimada
                </label>
                <select
                  value={severity}
                  onChange={(e) => setSeverity(e.target.value as IncidentSeverity)}
                  className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
                >
                  <option value="minor">Leve (Riesgo bajo)</option>
                  <option value="moderate">Moderado (Riesgo potencial)</option>
                  <option value="major">Grave (Afectación o daño relevante)</option>
                  <option value="catastrophic">Crítico (Peligro de vida inminente)</option>
                </select>
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                Título Resumido del Evento
              </label>
              <input
                type="text"
                required
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Ej. Fuga de aire comprimido a alta presión sin guarda de seguridad"
                className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                Ubicación Exacta en Planta
              </label>
              <input
                type="text"
                required
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="Ej. Nave 4, Cuarto de Compresores, Estación 12"
                className="w-full px-4 py-3 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                Descripción Detallada de los Hechos
              </label>
              <textarea
                rows={3}
                required
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Explique qué sucedió, equipo o personal involucrado y acción inmediata tomada..."
                className="w-full px-4 py-2.5 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-xs font-bold uppercase tracking-wider text-zinc-300">
                URL o Fotografía de Evidencia (Opcional)
              </label>
              <div className="relative">
                <input
                  type="text"
                  value={photoUrl}
                  onChange={(e) => setPhotoUrl(e.target.value)}
                  placeholder="https://storage.asesamo.com/incidents/evidencia.jpg"
                  className="w-full px-4 py-2.5 pl-10 bg-zinc-950 border border-zinc-800 rounded-xl text-white text-sm focus:outline-none focus:ring-2 focus:ring-red-500/50"
                />
                <Camera className="w-4 h-4 text-zinc-500 absolute left-3.5 top-3" />
              </div>
            </div>

            <div className="flex justify-end pt-2">
              <Button
                type="submit"
                disabled={isSubmitting}
                className="bg-red-600 hover:bg-red-500 text-white font-semibold px-6 py-2.5 rounded-xl shadow-md gap-2"
              >
                {isSubmitting ? 'Registrando...' : 'Emitir Reporte Oficial'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default function ReportarIncidentePage() {
  return (
    <ProtectedRoute>
      <ReportarIncidenteContent />
    </ProtectedRoute>
  );
}
