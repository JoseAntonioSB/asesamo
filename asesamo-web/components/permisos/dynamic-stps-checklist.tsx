'use client';

import React, { useState, useEffect } from 'react';
import { PermitType, PermitChecklistItem } from '@/lib/api/types';
import { ShieldCheck, CheckSquare, Square, AlertTriangle, Wind, Info } from 'lucide-react';

interface DynamicStpsChecklistProps {
  permitType: PermitType;
  items?: PermitChecklistItem[];
  readOnly?: boolean;
  onChangeItems?: (items: PermitChecklistItem[]) => void;
}

interface DefaultItemDefinition {
  item_key: string;
  nom_standard: string;
  title: string;
  description: string;
  category: string;
  is_mandatory: boolean;
}

const defaultChecklistDefinitions: Record<PermitType, DefaultItemDefinition[]> = {
  heights: [
    {
      item_key: 'chk_heights_harness',
      nom_standard: 'NOM-009-STPS-2011',
      title: 'Inspección de arnés de cuerpo completo y línea de vida',
      description: 'Inspección visual de costuras, herrajes, cinta de absorción de impacto y punto de anclaje certificado (>2,268 kg / 5,000 lbs).',
      category: 'ppe',
      is_mandatory: true,
    },
    {
      item_key: 'chk_heights_scaffold_ladder',
      nom_standard: 'NOM-009-STPS-2011',
      title: 'Estabilidad de andamios y plataformas de elevación',
      description: 'Revisión de nivelación, bases sólidas, barandales de seguridad, rodapiés y tarjeta de inspección de andamio vigente.',
      category: 'equipment',
      is_mandatory: true,
    },
    {
      item_key: 'chk_heights_barricade',
      nom_standard: 'NOM-009-STPS-2011',
      title: 'Delimitación y señalización del área inferior',
      description: 'Acordonamiento con cinta o barreras rígidas en el radio de posible caída de herramientas u objetos.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_heights_weather',
      nom_standard: 'NOM-009-STPS-2011',
      title: 'Condiciones climáticas seguras',
      description: 'Ausencia de lluvia, tormentas eléctricas o vientos superiores a 40 km/h que pongan en riesgo al personal.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_heights_epp',
      nom_standard: 'NOM-009-STPS-2011',
      title: 'EPP básico y complementario completo',
      description: 'Casco con barbiquejo de 3 puntos, calzado de seguridad dieléctrico, guantes antiderrapantes y lentes de seguridad.',
      category: 'ppe',
      is_mandatory: true,
    },
  ],
  confined_space: [
    {
      item_key: 'chk_confined_atm_test',
      nom_standard: 'NOM-033-STPS-2015',
      title: 'Monitoreo y prueba atmosférica continua',
      description: 'Medición obligatoria de Oxígeno (19.5%-23.5%), LEL (<10%), Monóxido de Carbono (<25 ppm) y Ácido Sulfhídrico (<10 ppm).',
      category: 'atmospheric',
      is_mandatory: true,
    },
    {
      item_key: 'chk_confined_ventilation',
      nom_standard: 'NOM-033-STPS-2015',
      title: 'Ventilación forzada mecánica',
      description: 'Inyección o extracción de aire continuo antes y durante el tiempo de permanencia en el espacio confinado.',
      category: 'equipment',
      is_mandatory: true,
    },
    {
      item_key: 'chk_confined_vigia',
      nom_standard: 'NOM-033-STPS-2015',
      title: 'Vigía exterior permanente',
      description: 'Designación y presencia obligatoria de vigía capacitado con comunicación bidireccional continua hacia el interior.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_confined_rescue',
      nom_standard: 'NOM-033-STPS-2015',
      title: 'Equipo de rescate y extracción listo',
      description: 'Trípode o pescante de rescate, malacate retráctil y arnés de rescate colocados para extracción inmediata sin ingreso.',
      category: 'emergency',
      is_mandatory: true,
    },
    {
      item_key: 'chk_confined_isolation',
      nom_standard: 'NOM-033-STPS-2015',
      title: 'Aislamiento y cegado de líneas',
      description: 'Cierre de válvulas, bridas ciegas colocadas y despresurización de tuberías conectadas al espacio.',
      category: 'safety',
      is_mandatory: true,
    },
  ],
  hot_work: [
    {
      item_key: 'chk_hot_extinguisher',
      nom_standard: 'NOM-027-STPS-2008',
      title: 'Extintor inspeccionado en el punto de trabajo',
      description: 'Extintor de PQS o CO2 de al menos 9 kg verificado con manómetro en verde a no más de 5 metros.',
      category: 'emergency',
      is_mandatory: true,
    },
    {
      item_key: 'chk_hot_combustibles',
      nom_standard: 'NOM-027-STPS-2008',
      title: 'Despeje de materiales combustibles a 11 metros',
      description: 'Retiro de solventes, cartón, madera o protección con mantas ignífugas si no pueden moverse.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_hot_screens',
      nom_standard: 'NOM-027-STPS-2008',
      title: 'Mamparas de protección contra chispas y radiación',
      description: 'Instalación de cortinas o mamparas retardantes al fuego para proteger áreas adyacentes y personal cercano.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_hot_grounding',
      nom_standard: 'NOM-027-STPS-2008',
      title: 'Conexión a tierra y estado de cables/mangueras',
      description: 'Pinza de tierra firme a la pieza o estructura, cables sin empalmes pelados, válvulas arrestallamas en cilindros.',
      category: 'equipment',
      is_mandatory: true,
    },
    {
      item_key: 'chk_hot_fire_watch',
      nom_standard: 'NOM-027-STPS-2008',
      title: 'Guardia contra incendios (Fire Watch)',
      description: 'Vigilancia durante el trabajo y permanencia obligatoria durante al menos 30 minutos posteriores al concluir.',
      category: 'safety',
      is_mandatory: true,
    },
  ],
  electrical_loto: [
    {
      item_key: 'chk_loto_zero_energy',
      nom_standard: 'NOM-029-STPS-2011',
      title: 'Verificación de ausencia de tensión (energía cero)',
      description: 'Prueba con multímetro calibrado o detector de tensión de rango adecuado en todas las fases antes de intervenir.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_loto_locks_tags',
      nom_standard: 'NOM-029-STPS-2011',
      title: 'Colocación de candados y etiquetas LOTO personales',
      description: 'Cada trabajador involucrado coloca su candado y tarjeta personal de peligro en el dispositivo de bloqueo.',
      category: 'safety',
      is_mandatory: true,
    },
    {
      item_key: 'chk_loto_grounding',
      nom_standard: 'NOM-029-STPS-2011',
      title: 'Puesta a tierra y en cortocircuito temporal',
      description: 'Instalación de equipo de puesta a tierra provisional y delimitación física del área de tableros o celdas.',
      category: 'equipment',
      is_mandatory: true,
    },
    {
      item_key: 'chk_loto_ppe',
      nom_standard: 'NOM-029-STPS-2011',
      title: 'EPP dieléctrico y contra arco eléctrico verificado',
      description: 'Guantes dieléctricos con prueba de inflado, careta para arco con balaclava, calzado dieléctrico y ropa de algodón.',
      category: 'ppe',
      is_mandatory: true,
    },
    {
      item_key: 'chk_loto_stored_energy',
      nom_standard: 'NOM-029-STPS-2011',
      title: 'Descarga de energía residual almacenada',
      description: 'Descarga de capacitores, purga de presión hidráulica/neumática y bloqueo mecánico de partes móviles o resortes.',
      category: 'safety',
      is_mandatory: true,
    },
  ],
};

export function DynamicStpsChecklist({
  permitType,
  items,
  readOnly = false,
  onChangeItems,
}: DynamicStpsChecklistProps) {
  // Atmospheric measurement inputs state
  const [measurements, setMeasurements] = useState<{
    o2: string;
    lel: string;
    co: string;
    h2s: string;
  }>({
    o2: '20.9',
    lel: '0',
    co: '0',
    h2s: '0',
  });

  // Local checklist items state
  const [localItems, setLocalItems] = useState<PermitChecklistItem[]>([]);

  useEffect(() => {
    if (items && items.length > 0) {
      setLocalItems(items);
      // Extract measurements if atmospheric item exists
      const atmItem = items.find((it) => it.item_key === 'chk_confined_atm_test');
      if (atmItem?.measurements) {
        setMeasurements({
          o2: String(atmItem.measurements.o2 ?? '20.9'),
          lel: String(atmItem.measurements.lel ?? '0'),
          co: String(atmItem.measurements.co ?? '0'),
          h2s: String(atmItem.measurements.h2s ?? '0'),
        });
      }
    } else {
      // Initialize with default standard items for this permit type
      const defs = defaultChecklistDefinitions[permitType] || [];
      const initialized: PermitChecklistItem[] = defs.map((d, idx) => ({
        id: `local-${d.item_key}`,
        item_key: d.item_key,
        nom_standard: d.nom_standard,
        title: d.title,
        description: d.description,
        category: d.category,
        is_mandatory: d.is_mandatory,
        is_verified: false,
        is_checked: false,
        display_order: idx + 1,
        notes: '',
        measurements: d.item_key === 'chk_confined_atm_test' ? { o2: 20.9, lel: 0, co: 0, h2s: 0 } : null,
      }));
      setLocalItems(initialized);
      onChangeItems?.(initialized);
    }
  }, [permitType, items, onChangeItems]);

  const toggleCheck = (itemKey: string) => {
    if (readOnly) return;
    const updated = localItems.map((it) => {
      if (it.item_key === itemKey) {
        const nextChecked = !(it.is_verified || it.is_checked);
        return {
          ...it,
          is_verified: nextChecked,
          is_checked: nextChecked,
        };
      }
      return it;
    });
    setLocalItems(updated);
    onChangeItems?.(updated);
  };

  const updateNotes = (itemKey: string, notes: string) => {
    if (readOnly) return;
    const updated = localItems.map((it) => {
      if (it.item_key === itemKey) {
        return { ...it, notes };
      }
      return it;
    });
    setLocalItems(updated);
    onChangeItems?.(updated);
  };

  const handleMeasurementChange = (field: 'o2' | 'lel' | 'co' | 'h2s', value: string) => {
    if (readOnly) return;
    const nextMeasurements = { ...measurements, [field]: value };
    setMeasurements(nextMeasurements);

    const updated = localItems.map((it) => {
      if (it.item_key === 'chk_confined_atm_test') {
        return {
          ...it,
          measurements: {
            o2: parseFloat(nextMeasurements.o2) || 0,
            lel: parseFloat(nextMeasurements.lel) || 0,
            co: parseFloat(nextMeasurements.co) || 0,
            h2s: parseFloat(nextMeasurements.h2s) || 0,
          },
        };
      }
      return it;
    });
    setLocalItems(updated);
    onChangeItems?.(updated);
  };

  // Evaluate Atmospheric Status for Confined Spaces
  const o2Num = parseFloat(measurements.o2) || 0;
  const lelNum = parseFloat(measurements.lel) || 0;
  const coNum = parseFloat(measurements.co) || 0;
  const h2sNum = parseFloat(measurements.h2s) || 0;

  const o2Safe = o2Num >= 19.5 && o2Num <= 23.5;
  const lelSafe = lelNum < 10;
  const coSafe = coNum < 25;
  const h2sSafe = h2sNum < 10;
  const atmosphericSafe = o2Safe && lelSafe && coSafe && h2sSafe;

  const allVerified = localItems.every((it) => it.is_verified || it.is_checked);

  return (
    <div className="space-y-4">
      {/* Header Info */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 p-3 bg-slate-900/90 border border-slate-800 rounded-xl">
        <div className="flex items-center gap-2">
          <ShieldCheck className="w-5 h-5 text-cyan-400" />
          <h4 className="text-sm font-semibold text-white">
            Lista de Verificación Preventiva (STPS)
          </h4>
        </div>
        <div className="text-xs font-mono">
          {allVerified ? (
            <span className="px-2 py-0.5 rounded bg-emerald-950 text-emerald-300 border border-emerald-700 font-semibold">
              ✓ 100% Verificado
            </span>
          ) : (
            <span className="px-2 py-0.5 rounded bg-amber-950 text-amber-300 border border-amber-700">
              {localItems.filter((i) => i.is_verified || i.is_checked).length} de {localItems.length} verificados
            </span>
          )}
        </div>
      </div>

      {/* Confined Space Atmospheric Test Box */}
      {permitType === 'confined_space' && (
        <div className="p-4 bg-slate-900/90 border border-amber-800/40 rounded-xl space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Wind className="w-4 h-4 text-amber-400" />
              <h5 className="text-xs font-bold text-white uppercase tracking-wider">
                Prueba Atmosférica Multigás (NOM-033-STPS-2015)
              </h5>
            </div>
            {atmosphericSafe ? (
              <span className="text-[11px] font-mono font-bold text-emerald-400 bg-emerald-950/80 border border-emerald-700 px-2 py-0.5 rounded">
                ✓ Atmósfera Segura
              </span>
            ) : (
              <span className="text-[11px] font-mono font-bold text-rose-400 bg-rose-950/80 border border-rose-700 px-2 py-0.5 rounded animate-pulse">
                ✕ Peligro: Atmósfera Fuera de Rango
              </span>
            )}
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {/* O2 */}
            <div className="p-2.5 rounded-lg bg-slate-950 border border-slate-800">
              <label className="text-[10px] text-slate-400 block font-medium">Oxígeno (O₂)</label>
              <div className="flex items-center gap-1.5 mt-1">
                <input
                  type="number"
                  step="0.1"
                  disabled={readOnly}
                  value={measurements.o2}
                  onChange={(e) => handleMeasurementChange('o2', e.target.value)}
                  className={`w-full bg-slate-900 border rounded px-2 py-1 text-sm font-mono font-bold text-right focus:outline-none ${
                    o2Safe ? 'border-emerald-600 text-emerald-400' : 'border-rose-600 text-rose-400'
                  }`}
                />
                <span className="text-xs text-slate-400">%</span>
              </div>
              <span className="text-[9px] text-slate-500 block mt-1 font-mono">Norma: 19.5 - 23.5%</span>
            </div>

            {/* LEL */}
            <div className="p-2.5 rounded-lg bg-slate-950 border border-slate-800">
              <label className="text-[10px] text-slate-400 block font-medium">Explosividad (LEL)</label>
              <div className="flex items-center gap-1.5 mt-1">
                <input
                  type="number"
                  step="1"
                  disabled={readOnly}
                  value={measurements.lel}
                  onChange={(e) => handleMeasurementChange('lel', e.target.value)}
                  className={`w-full bg-slate-900 border rounded px-2 py-1 text-sm font-mono font-bold text-right focus:outline-none ${
                    lelSafe ? 'border-emerald-600 text-emerald-400' : 'border-rose-600 text-rose-400'
                  }`}
                />
                <span className="text-xs text-slate-400">%</span>
              </div>
              <span className="text-[9px] text-slate-500 block mt-1 font-mono">Límite: &lt; 10% LEL</span>
            </div>

            {/* CO */}
            <div className="p-2.5 rounded-lg bg-slate-950 border border-slate-800">
              <label className="text-[10px] text-slate-400 block font-medium">Monóxido (CO)</label>
              <div className="flex items-center gap-1.5 mt-1">
                <input
                  type="number"
                  step="1"
                  disabled={readOnly}
                  value={measurements.co}
                  onChange={(e) => handleMeasurementChange('co', e.target.value)}
                  className={`w-full bg-slate-900 border rounded px-2 py-1 text-sm font-mono font-bold text-right focus:outline-none ${
                    coSafe ? 'border-emerald-600 text-emerald-400' : 'border-rose-600 text-rose-400'
                  }`}
                />
                <span className="text-xs text-slate-400">ppm</span>
              </div>
              <span className="text-[9px] text-slate-500 block mt-1 font-mono">Límite: &lt; 25 ppm</span>
            </div>

            {/* H2S */}
            <div className="p-2.5 rounded-lg bg-slate-950 border border-slate-800">
              <label className="text-[10px] text-slate-400 block font-medium">Ácido Sulfhídrico (H₂S)</label>
              <div className="flex items-center gap-1.5 mt-1">
                <input
                  type="number"
                  step="0.5"
                  disabled={readOnly}
                  value={measurements.h2s}
                  onChange={(e) => handleMeasurementChange('h2s', e.target.value)}
                  className={`w-full bg-slate-900 border rounded px-2 py-1 text-sm font-mono font-bold text-right focus:outline-none ${
                    h2sSafe ? 'border-emerald-600 text-emerald-400' : 'border-rose-600 text-rose-400'
                  }`}
                />
                <span className="text-xs text-slate-400">ppm</span>
              </div>
              <span className="text-[9px] text-slate-500 block mt-1 font-mono">Límite: &lt; 10 ppm</span>
            </div>
          </div>
        </div>
      )}

      {/* Checklist items list */}
      <div className="space-y-3">
        {localItems.map((item, index) => {
          const isChecked = !!(item.is_verified || item.is_checked);

          return (
            <div
              key={item.item_key}
              className={`p-3.5 rounded-xl border transition-all duration-200 ${
                isChecked
                  ? 'bg-slate-900/80 border-emerald-600/50'
                  : 'bg-slate-900/50 border-slate-800 hover:border-slate-700'
              }`}
            >
              <div className="flex items-start gap-3">
                <button
                  type="button"
                  disabled={readOnly}
                  onClick={() => toggleCheck(item.item_key)}
                  className={`mt-0.5 shrink-0 focus:outline-none transition-colors ${
                    readOnly ? 'cursor-default' : 'cursor-pointer hover:opacity-80'
                  }`}
                >
                  {isChecked ? (
                    <CheckSquare className="w-5 h-5 text-emerald-400 fill-emerald-950" />
                  ) : (
                    <Square className="w-5 h-5 text-slate-500 hover:text-slate-400" />
                  )}
                </button>

                <div className="flex-1 space-y-1">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span
                      onClick={() => !readOnly && toggleCheck(item.item_key)}
                      className={`text-sm font-semibold cursor-pointer ${
                        isChecked ? 'text-white' : 'text-slate-300'
                      }`}
                    >
                      {item.title}
                    </span>
                    {item.nom_standard && (
                      <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                        {item.nom_standard}
                      </span>
                    )}
                  </div>

                  {item.description && (
                    <p className="text-xs text-slate-400 leading-relaxed">{item.description}</p>
                  )}

                  {/* Optional Notes field */}
                  {!readOnly ? (
                    <div className="pt-2">
                      <input
                        type="text"
                        placeholder="Observaciones o notas de verificación (opcional)..."
                        value={item.notes || ''}
                        onChange={(e) => updateNotes(item.item_key, e.target.value)}
                        className="w-full bg-slate-950/60 border border-slate-800/80 rounded px-2.5 py-1 text-xs text-slate-300 placeholder-slate-600 focus:outline-none focus:border-cyan-500"
                      />
                    </div>
                  ) : item.notes ? (
                    <div className="pt-1 text-xs text-slate-400 italic">
                      Notas: <span className="text-slate-300">{item.notes}</span>
                    </div>
                  ) : null}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
