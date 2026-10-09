/**
 * ASESAMO — API Contracts and Domain Models
 * Location: asesamo-web/lib/api/types.ts
 */

// ==========================================
// 1. HTTP Envelope Contracts
// ==========================================

export interface ApiResponse<T, M = unknown> {
  data: T;
  meta?: M;
}

export interface ApiErrorDetail {
  code: string;
  message: string;
}

export interface ApiErrorPayload {
  error: ApiErrorDetail;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
}

// ==========================================
// 2. Authentication & User Profile
// ==========================================

export type UserRole = 'guard' | 'plant_admin' | 'ehs_manager' | 'super_admin' | 'contractor' | 'contractor_admin' | 'worker';

export interface UserProfile {
  id: string; // UUID
  tenant_id?: string; // UUID
  email: string;
  full_name: string;
  role: UserRole | string;
  theme_preference?: 'light' | 'dark' | 'system';
}

// ==========================================
// 2.1 Document Management Models
// ==========================================

export type DocumentStatus = 'pending_review' | 'approved' | 'rejected' | 'expired';

export interface DocumentItem {
  id: string;
  user_id?: string;
  contractor_id?: string;
  document_type_id: string;
  document_type_code?: string;
  document_type_name?: string;
  owner_name?: string;
  owner_type?: string;
  contractor_name?: string;
  file_name: string;
  file_size_bytes?: number;
  mime_type?: string;
  issued_at?: string;
  expires_at?: string;
  status: DocumentStatus;
  rejection_reason?: string;
  reviewed_by?: string;
  reviewed_by_name?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
  download_url: string;
}

export interface ReviewDocumentPayload {
  action: 'approve' | 'reject';
  rejection_reason?: string;
}

export interface DocumentTypeItem {
  id: string;
  code: string;
  name: string;
  description?: string;
  applies_to: string;
  default_validity_days?: number;
  is_required_default: boolean;
}

export interface UploadDocumentParams {
  file: File;
  documentTypeCode: string;
  issuedAt?: string;
  expiresAt?: string;
  workerId?: string;
  contractorId?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponseData {
  token: string;
  user: UserProfile;
}

// ==========================================
// 3. Access Points
// ==========================================

export interface AccessPointResponse {
  id: string; // UUID
  plant_id: string; // UUID
  tenant_id: string; // UUID
  name: string;
  description?: string | null;
  is_active: boolean;
}

// ==========================================
// 4. Check-In & Compliance Verification
// ==========================================

export type AccessResult = 'granted' | 'granted_with_warning' | 'denied';

export interface CheckInRequest {
  worker_id?: string; // UUID
  worker_user_id?: string; // UUID
  access_point_id: string; // UUID
  notes?: string;
}

export interface WorkerSummary {
  user_id: string; // UUID
  full_name: string;
  curp?: string | null;
  imss_number?: string | null;
  job_title?: string | null;
  avatar_r2_key?: string | null;
}

export interface ContractorSummary {
  id: string; // UUID
  name: string;
  rfc?: string | null;
  repse_number?: string | null;
  status: string;
}

export interface PlantSummary {
  id: string; // UUID
  name: string;
}

export interface AccessPointSummary {
  id: string; // UUID
  name: string;
}

export type DocumentComplianceStatus = 'approved' | 'rejected' | 'pending' | 'expired' | 'missing';

export interface DocumentSnapshotItem {
  document_type_code: string; // e.g. "DC3", "INE", "REPSE", "SUA"
  document_name: string;
  applies_to: 'worker' | 'contractor' | string;
  status: DocumentComplianceStatus | string;
  is_required: boolean;
  is_valid: boolean;
  expires_at?: string | null; // YYYY-MM-DD or ISO timestamp
  failure_reason?: string | null;
}

export interface CheckInOutput {
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

// ==========================================
// 5. Check-Out Egress
// ==========================================

export interface CheckOutRequest {
  worker_id?: string;
  worker_user_id?: string;
  plant_id?: string;
  access_point_id?: string;
}

export interface CheckOutOutput {
  access_log_id: string;
  result?: AccessResult;
  checked_in_at?: string;
  checked_out_at: string;
  duration_minutes: number;
}

// ==========================================
// 6. Access Logs & Audit
// ==========================================

export interface AccessLogsQueryParams {
  plant_id?: string;
  access_point_id?: string;
  worker_id?: string;
  worker_user_id?: string;
  result?: AccessResult;
  from?: string; // RFC3339 or YYYY-MM-DD
  to?: string;   // RFC3339 or YYYY-MM-DD
  page?: number;
  per_page?: number;
}

// ==========================================
// 7. Worker Credential & Mobile Badge
// ==========================================

export interface WorkerCredentialDetail {
  user_id: string;
  full_name: string;
  curp?: string | null;
  imss_number?: string | null;
  job_title?: string | null;
  avatar_r2_key?: string | null;
  is_active: boolean;
}

export interface ContractorCredentialDetail {
  id: string;
  name: string;
  rfc?: string | null;
  repse_number?: string | null;
  status: string;
}

export interface PlantCredentialDetail {
  id: string;
  name: string;
}

export interface WorkerCredentialResponse {
  worker: WorkerCredentialDetail;
  contractor: ContractorCredentialDetail;
  plant: PlantCredentialDetail;
  qr_payload: string;
  compliance_status: AccessResult;
  denial_reasons?: string[];
  warning_reasons?: string[];
  documents: DocumentSnapshotItem[];
  generated_at: string;
}

// ==========================================
// 8. Contractor & Crew Management
// ==========================================

export type PlantContractorStatus = 'active' | 'pending' | 'suspended' | 'blocked';
export type ContractorType = 'company' | 'individual';

export interface ContractorItem {
  id: string; // UUID
  name: string;
  rfc?: string | null;
  repse_number?: string | null;
  contractor_type: ContractorType | string;
  admin_user_id?: string | null;
  admin_full_name?: string | null;
  admin_email?: string | null;
  admin_phone?: string | null;
  plant_status?: PlantContractorStatus | string;
  workers_count: number;
  is_active: boolean;
  created_at: string;
}

export interface ContractorWorkerItem {
  worker_id: string; // UUID (user_id)
  contractor_id: string; // UUID
  full_name: string;
  email: string;
  phone?: string | null;
  job_title?: string | null;
  curp?: string | null;
  imss_number?: string | null;
  is_active: boolean;
  user_is_active: boolean;
  documents_count: number;
  created_at: string;
}

export interface InviteContractorRequest {
  plant_id: string;
  name: string;
  rfc: string;
  repse_number?: string;
  contractor_type: 'company' | 'individual';
  admin_email: string;
  admin_full_name?: string;
  admin_phone?: string;
  initial_status?: 'active' | 'pending';
}

export interface UpdateContractorStatusPayload {
  plant_id: string;
  status: PlantContractorStatus;
}

export interface AddContractorWorkerPayload {
  contractor_id?: string;
  full_name: string;
  email: string;
  phone?: string;
  job_title?: string;
  curp?: string;
  imss_number?: string;
}

export interface ToggleWorkerStatusPayload {
  contractor_id?: string;
  is_active: boolean;
}

// ==========================================
// 9. Metrics & STPS Audit Reports
// ==========================================

export interface AccessMetricsSummary {
  total_accesses: number;
  granted_count: number;
  warning_count: number;
  denied_count: number;
  compliance_rate: number;
  avg_duration_minutes: number;
}

export interface DailyAccessDistribution {
  date: string;
  total: number;
  granted: number;
  denied: number;
}

export interface RejectionReasonSummary {
  reason: string;
  count: number;
  percentage: number;
}

export interface ContractorComplianceSummary {
  contractor_id: string;
  contractor_name: string;
  total_attempts: number;
  granted_count: number;
  denied_count: number;
  compliance_rate: number;
}

export interface ReportMetricsResponse {
  summary: AccessMetricsSummary;
  timeline: DailyAccessDistribution[];
  rejection_reasons: RejectionReasonSummary[];
  contractor_ranking: ContractorComplianceSummary[];
  generated_at: string;
}

export interface ReportFilterParams {
  plant_id?: string;
  from?: string;
  to?: string;
}


