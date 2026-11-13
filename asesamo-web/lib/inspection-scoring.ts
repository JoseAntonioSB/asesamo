import { InspectionAnswerPayload, InspectionStatus } from './api/types';

export interface InspectionScoringItem {
  id: string;
  weight: number;
  critical_risk: boolean;
}

export interface InspectionScoreSummary {
  scorePercentage: number;
  totalItems: number;
  compliantCount: number;
  nonCompliantCount: number;
  naCount: number;
  hasCriticalFailure: boolean;
}

export function calculateInspectionScore(
  items: InspectionScoringItem[],
  answers: InspectionAnswerPayload[]
): InspectionScoreSummary {
  const itemMap = new Map<string, InspectionScoringItem>();
  for (const itm of items) {
    itemMap.set(itm.id, itm);
  }

  let totalWeightApplicable = 0;
  let compliantWeight = 0;
  let compliantCount = 0;
  let nonCompliantCount = 0;
  let naCount = 0;
  let hasCriticalFailure = false;

  for (const ans of answers) {
    const itm = itemMap.get(ans.template_item_id);
    const weight = itm?.weight ?? 1;
    const isCritical = itm?.critical_risk ?? false;

    switch (ans.status) {
      case 'compliant':
        compliantCount++;
        compliantWeight += weight;
        totalWeightApplicable += weight;
        break;
      case 'non_compliant':
        nonCompliantCount++;
        totalWeightApplicable += weight;
        if (isCritical) {
          hasCriticalFailure = true;
        }
        break;
      case 'not_applicable':
        naCount++;
        break;
    }
  }

  let scorePercentage = 100;
  if (totalWeightApplicable > 0) {
    scorePercentage = Math.round((compliantWeight / totalWeightApplicable) * 10000) / 100;
  }

  return {
    scorePercentage,
    totalItems: answers.length,
    compliantCount,
    nonCompliantCount,
    naCount,
    hasCriticalFailure,
  };
}

export function evaluateInspectionStatus(summary: InspectionScoreSummary): InspectionStatus {
  if (summary.hasCriticalFailure || summary.nonCompliantCount > 0) {
    return 'requires_capa';
  }
  return 'completed';
}
