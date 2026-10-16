'use client';

import React from 'react';
import { WorkPermitItem } from '@/lib/api/types';
import { Activity, Clock, CheckCircle2, AlertTriangle, FileText } from 'lucide-react';

interface PermitSummaryCardsProps {
  permits: WorkPermitItem[];
  selectedStatus?: string;
  onSelectStatus?: (status: string) => void;
}

export function PermitSummaryCards({
  permits,
  selectedStatus,
  onSelectStatus,
}: PermitSummaryCardsProps) {
  const activeCount = permits.filter((p) => p.status === 'approved_active').length;
  const pendingCount = permits.filter((p) => p.status === 'pending_approval').length;
  const closedCount = permits.filter((p) => p.status === 'closed').length;
  const suspendedCount = permits.filter((p) => p.status === 'suspended').length;
  const totalCount = permits.length;

  const cards = [
    {
      id: 'all',
      title: 'Total Permisos',
      value: totalCount,
      subtitle: 'Registrados en planta',
      icon: FileText,
      bgColor: 'bg-slate-900/60',
      borderColor: 'border-slate-800',
      activeBorder: 'ring-2 ring-slate-400',
      iconColor: 'text-slate-400',
      filterValue: '',
    },
    {
      id: 'approved_active',
      title: 'Activos en Campo',
      value: activeCount,
      subtitle: 'Trabajos en ejecución',
      icon: Activity,
      bgColor: 'bg-emerald-950/30',
      borderColor: 'border-emerald-800/40',
      activeBorder: 'ring-2 ring-emerald-500',
      iconColor: 'text-emerald-400',
      pulsing: true,
      filterValue: 'approved_active',
    },
    {
      id: 'pending_approval',
      title: 'Pendiente EHS',
      value: pendingCount,
      subtitle: 'Requiere autorización',
      icon: Clock,
      bgColor: 'bg-amber-950/30',
      borderColor: 'border-amber-800/40',
      activeBorder: 'ring-2 ring-amber-500',
      iconColor: 'text-amber-400',
      pulsing: pendingCount > 0,
      filterValue: 'pending_approval',
    },
    {
      id: 'suspended',
      title: 'Suspendidos',
      value: suspendedCount,
      subtitle: 'Por condición o clima',
      icon: AlertTriangle,
      bgColor: 'bg-rose-950/30',
      borderColor: 'border-rose-800/40',
      activeBorder: 'ring-2 ring-rose-500',
      iconColor: 'text-rose-400',
      filterValue: 'suspended',
    },
    {
      id: 'closed',
      title: 'Cerrados',
      value: closedCount,
      subtitle: 'Trabajos concluidos',
      icon: CheckCircle2,
      bgColor: 'bg-blue-950/30',
      borderColor: 'border-blue-800/40',
      activeBorder: 'ring-2 ring-blue-500',
      iconColor: 'text-blue-400',
      filterValue: 'closed',
    },
  ];

  return (
    <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
      {cards.map((card) => {
        const IconComponent = card.icon;
        const isSelected = selectedStatus === card.filterValue;

        return (
          <button
            key={card.id}
            type="button"
            onClick={() => onSelectStatus?.(card.filterValue)}
            className={`text-left p-4 rounded-xl border transition-all duration-200 ${card.bgColor} ${card.borderColor} ${
              isSelected ? `${card.activeBorder} shadow-lg scale-[1.02]` : 'hover:border-slate-600'
            }`}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-medium text-slate-400">{card.title}</span>
              <div className="relative">
                {card.pulsing && (
                  <span className="absolute -top-1 -right-1 flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                  </span>
                )}
                <IconComponent className={`w-4 h-4 ${card.iconColor}`} />
              </div>
            </div>
            <div className="text-2xl font-bold font-mono text-white mb-1">{card.value}</div>
            <div className="text-[11px] text-slate-500 truncate">{card.subtitle}</div>
          </button>
        );
      })}
    </div>
  );
}
