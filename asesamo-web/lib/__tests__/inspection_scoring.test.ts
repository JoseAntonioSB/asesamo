import { test } from 'node:test';
import assert from 'node:assert/strict';
import { calculateInspectionScore, evaluateInspectionStatus } from '../inspection-scoring';
import { InspectionAnswerPayload } from '../api/types';

test('calculateInspectionScore computes exact weighted percentage and flags critical failures', () => {
  const items = [
    { id: 'item-1', weight: 3, critical_risk: true },
    { id: 'item-2', weight: 1, critical_risk: false },
    { id: 'item-3', weight: 2, critical_risk: false },
  ];

  // Caso 1: Todo conforme -> 100% y status completed
  const answers1: InspectionAnswerPayload[] = [
    { template_item_id: 'item-1', status: 'compliant' },
    { template_item_id: 'item-2', status: 'compliant' },
    { template_item_id: 'item-3', status: 'compliant' },
  ];
  const res1 = calculateInspectionScore(items, answers1);
  assert.equal(res1.scorePercentage, 100);
  assert.equal(res1.compliantCount, 3);
  assert.equal(res1.nonCompliantCount, 0);
  assert.equal(evaluateInspectionStatus(res1), 'completed');

  // Caso 2: Falla el ítem crítico (peso 3) -> 3 de 6 perdidos -> 50% y status requires_capa
  const answers2: InspectionAnswerPayload[] = [
    { template_item_id: 'item-1', status: 'non_compliant', finding_notes: 'Extintor sin carga' },
    { template_item_id: 'item-2', status: 'compliant' },
    { template_item_id: 'item-3', status: 'compliant' },
  ];
  const res2 = calculateInspectionScore(items, answers2);
  assert.equal(res2.scorePercentage, 50);
  assert.equal(res2.compliantCount, 2);
  assert.equal(res2.nonCompliantCount, 1);
  assert.equal(res2.hasCriticalFailure, true);
  assert.equal(evaluateInspectionStatus(res2), 'requires_capa');

  // Caso 3: Ítem no aplicable no penaliza el total
  const answers3: InspectionAnswerPayload[] = [
    { template_item_id: 'item-1', status: 'compliant' },
    { template_item_id: 'item-2', status: 'not_applicable' },
    { template_item_id: 'item-3', status: 'compliant' },
  ];
  const res3 = calculateInspectionScore(items, answers3);
  // Total peso aplicable: 3 + 2 = 5. Puntos obtenidos: 5 -> 100%
  assert.equal(res3.scorePercentage, 100);
  assert.equal(res3.naCount, 1);
  assert.equal(evaluateInspectionStatus(res3), 'completed');
});
