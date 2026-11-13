import { InspectionStatus } from './api/types';

export interface InspectionBadgeData {
  label: string;
  colorClass: string;
}

export function renderInspectionSummaryBadges(
  status: InspectionStatus,
  score: number | null
): InspectionBadgeData {
  switch (status) {
    case 'completed':
      return {
        label: score !== null ? `Conforme (${score}%)` : 'Conforme',
        colorClass: 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30',
      };
    case 'requires_capa':
      return {
        label: score !== null ? `Requiere CAPA (${score}%)` : 'Requiere CAPA',
        colorClass: 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30',
      };
    case 'in_progress':
      return {
        label: 'En Campo',
        colorClass: 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30',
      };
    case 'cancelled':
      return {
        label: 'Cancelada',
        colorClass: 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-400 border-zinc-500/30',
      };
    case 'draft':
    default:
      return {
        label: 'Borrador',
        colorClass: 'bg-zinc-500/15 text-zinc-700 dark:text-zinc-400 border-zinc-500/30',
      };
  }
}
