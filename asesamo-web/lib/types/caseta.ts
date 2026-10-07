/**
 * ASESAMO — Caseta Types and Interfaces
 * Location: asesamo-web/lib/types/caseta.ts
 */

export type AccessResult = 'granted' | 'granted_with_warning' | 'denied';

export type DocumentStatusCode = 'approved' | 'pending' | 'rejected' | 'expired' | 'missing';

export interface WorkerSummary {
  user_id: string;
  full_name: string;
  curp?: string | null;
  imss_number?: string | null;
  job_title?: string | null;
  avatar_r2_key?: string | null;
}

export interface ContractorSummary {
  id: string;
  name: string;
  rfc?: string | null;
  repse_number?: string | null;
  status: string;
}

export interface PlantSummary {
  id: string;
  name: string;
}

export interface AccessPointSummary {
  id: string;
  name: string;
}

export interface DocumentSnapshotItem {
  document_type_code: string;
  document_name: string;
  applies_to: 'worker' | 'contractor' | string;
  status: DocumentStatusCode | string;
  is_required: boolean;
  is_valid: boolean;
  expires_at?: string | null;
  failure_reason?: string | null;
}

export interface CheckInResponseData {
  access_log_id: string;
  result: AccessResult;
  denial_reason?: string | null;
  checked_in_at: string;
  worker: WorkerSummary;
  contractor: ContractorSummary;
  plant: PlantSummary;
  access_point: AccessPointSummary;
  document_snapshot: DocumentSnapshotItem[];
  documents?: DocumentSnapshotItem[];
}

export interface CheckOutResponseData {
  access_log_id: string;
  result?: AccessResult;
  checked_in_at?: string;
  checked_out_at: string;
  duration_minutes: number;
}

export interface AccessPointItem {
  id: string;
  plant_id: string;
  tenant_id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
}
