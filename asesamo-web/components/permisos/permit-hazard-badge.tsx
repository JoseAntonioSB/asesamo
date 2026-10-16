'use client';

import React from 'react';
import { PermitType } from '@/lib/api/types';
import { Mountain, Flame, Zap, ShieldAlert } from 'lucide-react';

interface PermitHazardBadgeProps {
  hazardType: PermitType | string;
  size?: 'sm' | 'md' | 'lg';
  showNom?: boolean;
  className?: string;
}

interface HazardConfig {
  label: string;
  nom: string;
  bgColor: string;
  textColor: string;
  borderColor: string;
  icon: React.ComponentType<{ className?: string }>;
}

const hazardConfigs: Record<string, HazardConfig> = {
  heights: {
    label: 'Trabajos en Alturas',
    nom: 'NOM-009-STPS-2011',
    bgColor: 'bg-sky-950/60',
    textColor: 'text-sky-300',
    borderColor: 'border-sky-700/60',
    icon: Mountain,
  },
  confined_space: {
    label: 'Espacios Confinados',
    nom: 'NOM-033-STPS-2015',
    bgColor: 'bg-amber-950/60',
    textColor: 'text-amber-300',
    borderColor: 'border-amber-700/60',
    icon: ShieldAlert,
  },
  hot_work: {
    label: 'Corte y Soldadura',
    nom: 'NOM-027-STPS-2008',
    bgColor: 'bg-rose-950/60',
    textColor: 'text-rose-300',
    borderColor: 'border-rose-700/60',
    icon: Flame,
  },
  electrical_loto: {
    label: 'LOTO Eléctrico',
    nom: 'NOM-029-STPS-2011',
    bgColor: 'bg-yellow-950/60',
    textColor: 'text-yellow-300',
    borderColor: 'border-yellow-700/60',
    icon: Zap,
  },
};

export function PermitHazardBadge({
  hazardType,
  size = 'md',
  showNom = true,
  className = '',
}: PermitHazardBadgeProps) {
  const normalizedKey = (hazardType || '').toLowerCase();
  const config = hazardConfigs[normalizedKey] || {
    label: hazardType,
    nom: 'STPS',
    bgColor: 'bg-slate-800/80',
    textColor: 'text-slate-300',
    borderColor: 'border-slate-700',
    icon: ShieldAlert,
  };

  const IconComponent = config.icon;

  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5 gap-1',
    md: 'text-xs px-2.5 py-1 gap-1.5',
    lg: 'text-sm px-3.5 py-1.5 gap-2 font-medium',
  }[size];

  const iconSizes = {
    sm: 'w-3 h-3',
    md: 'w-3.5 h-3.5',
    lg: 'w-4 h-4',
  }[size];

  return (
    <span
      className={`inline-flex items-center rounded-md border font-medium ${config.bgColor} ${config.textColor} ${config.borderColor} ${sizeClasses} ${className}`}
    >
      <IconComponent className={`${iconSizes} shrink-0`} />
      <span className="font-semibold">{config.label}</span>
      {showNom && (
        <span className="text-[10px] opacity-75 font-mono px-1 py-0.2 bg-black/30 rounded border border-white/10">
          {config.nom}
        </span>
      )}
    </span>
  );
}
