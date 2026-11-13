import { test } from 'node:test';
import assert from 'node:assert/strict';
import { renderInspectionSummaryBadges } from '../inspection-badge-logic';
import { InspectionStatus, FindingSeverity } from '../api/types';

test('renderInspectionSummaryBadges maps status to correct tone, label, and variant', () => {
  // Caso 1: Completada con 100% de cumplimiento
  const badgeCompleted = renderInspectionSummaryBadges('completed', 100);
  assert.equal(badgeCompleted.label, 'Conforme (100%)');
  assert.equal(badgeCompleted.colorClass, 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/30');

  // Caso 2: Requiere plan de acción (CAPA) por hallazgos no conformes
  const badgeCapa = renderInspectionSummaryBadges('requires_capa', 75);
  assert.equal(badgeCapa.label, 'Requiere CAPA (75%)');
  assert.equal(badgeCapa.colorClass, 'bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/30');

  // Caso 3: En progreso en campo
  const badgeInProgress = renderInspectionSummaryBadges('in_progress', null);
  assert.equal(badgeInProgress.label, 'En Campo');
  assert.equal(badgeInProgress.colorClass, 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/30');
});
