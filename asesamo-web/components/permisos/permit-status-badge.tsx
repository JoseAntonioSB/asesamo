'use client';

import React from 'react';
import { PermitStatus } from '@/lib/api/types';
import { FileEdit, Clock, Activity, AlertTriangle, CheckCircle2, Ban } from 'lucide-react';

interface PermitStatusBadgeProps {
  status: PermitStatus | string;
  size?: 'sm' | 'md' | 'lg';
  showIcon?: boolean;
  className?: string;
}

interface StatusConfig {
  label: string;
  bgColor: string;
  textColor: string;
  borderColor: string;
  dotColor?: string;
  icon: React.ComponentType<{ className?: string }>;
  pulsing?: boolean;
}

const statusConfigs: Record<string, StatusConfig> = {
  draft: {
    label: 'Borrador',
    bgColor: 'bg-slate-800/80',
    textColor: 'text-slate-300',
    borderColor: 'border-slate-700',
    icon: FileEdit,
  },
  pending_approval: {
    label: 'Pendiente EHS',
    bgColor: 'bg-amber-950/60',
    textColor: 'text-amber-400',
    borderColor: 'border-amber-700/60',
    dotColor: 'bg-amber-400',
    icon: Clock,
    pulsing: true,
  },
  approved_active: {
    label: 'Activo en Campo',
    bgColor: 'bg-emerald-950/60',
    textColor: 'text-emerald-400',
    borderColor: 'border-emerald-600/60',
    dotColor: 'bg-emerald-400',
    icon: Activity,
    pulsing: true,
  },
  suspended: {
    label: 'Suspendido',
    bgColor: 'bg-rose-950/70',
    textColor: 'text-rose-400',
    borderColor: 'border-rose-700/70',
    icon: AlertTriangle,
  },
  closed: {
    label: 'Cerrado',
    bgColor: 'bg-blue-950/50',
    textColor: 'text-blue-400',
    borderColor: 'border-blue-800/50',
    icon: CheckCircle2,
  },
  cancelled: {
    label: 'Cancelado',
    bgColor: 'bg-zinc-900/60',
    textColor: 'text-zinc-500',
    borderColor: 'border-zinc-800',
    icon: Ban,
  },
};

export function PermitStatusBadge({
  status,
  size = 'md',
  showIcon = true,
  className = '',
}: PermitStatusBadgeProps) {
  const normalizedKey = status.toLowerCase();
  const config = statusConfigs[normalizedKey] || {
    label: status,
    bgColor: 'bg-slate-800',
    textColor: 'text-slate-400',
    borderColor: 'border-slate-700',
    icon: FileEdit,
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
      className={`inline-flex items-center rounded-full border font-medium ${config.bgColor} ${config.textColor} ${config.borderColor} ${sizeClasses} ${className}`}
    >
      {config.dotColor && (
        <span className="relative flex h-2 w-2">
          {config.pulsing && (
            <span
              className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${config.dotColor}`}
            />
          )}
          <span className={`relative inline-flex rounded-full h-2 w-2 ${config.dotColor}`} />
        </span>
      )}
      {showIcon && !config.dotColor && <IconComponent className={iconSizes} />}
      <span>{config.label}</span>
    </span>
  );
}
