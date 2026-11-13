'use client';

import React from 'react';
import { InspectionStatus } from '@/lib/api/types';
import { renderInspectionSummaryBadges } from '@/lib/inspection-badge-logic';

interface InspectionStatusBadgeProps {
  status: InspectionStatus;
  score?: number | null;
  className?: string;
}

export function InspectionStatusBadge({
  status,
  score = null,
  className = '',
}: InspectionStatusBadgeProps) {
  const badge = renderInspectionSummaryBadges(status, score);

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border ${badge.colorClass} ${className}`}
    >
      <span className="w-1.5 h-1.5 rounded-full bg-current opacity-80" />
      {badge.label}
    </span>
  );
}
