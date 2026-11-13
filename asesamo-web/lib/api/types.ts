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
  phone?: string;
  role: UserRole | string;
  avatar_r2_key?: string;
  theme_preference?: 'light' | 'dark' | 'system';
}

export interface UserProfileDetail {
  id: string;
  tenant_id?: string;
  email: string;
  full_name: string;
  phone?: string;
  role: UserRole | string;
  avatar_r2_key?: string;
  theme_preference: 'light' | 'dark' | 'system';
  is_active: boolean;
  contractor_id?: string;
  contractor_name?: string;
  plant_id?: string;
  plant_name?: string;
  created_at: string;
}

export interface UpdateProfilePayload {
  full_name: string;
  phone?: string;
  current_password?: string;
  new_password?: string;
}

export interface AvatarUploadResponse {
  avatar_url: string;
  avatar_key: string;
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

export interface AccessLogItem {
  id: string;
  tenant_id: string;
  plant_id: string;
  access_point_id: string;
  user_id: string;
  contractor_id?: string;
  processed_by?: string;
  checked_in_at: string;
  checked_out_at?: string;
  result: AccessResult;
  denial_reason?: string | null;
  notes?: string | null;
  created_at: string;
}

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

// ==========================================
// 10. High-Risk Work Permits (PTAR Digital)
// ==========================================

export type PermitType = 'heights' | 'confined_space' | 'hot_work' | 'electrical_loto';

export type PermitStatus =
  | 'draft'
  | 'pending_approval'
  | 'approved_active'
  | 'suspended'
  | 'closed'
  | 'cancelled';

export type PermitSignatureRole = 'requester' | 'contractor_supervisor' | 'ehs_authorizer';

export interface WorkerComplianceSnapshot {
  is_compliant: boolean;
  block_reason?: string | null;
  dc3_status: string;
  dc3_code?: string;
  dc3_expires_at?: string | null;
  medical_status: string;
  medical_expires_at?: string | null;
  contractor_status: string;
  evaluated_at: string;
}

export interface WorkerComplianceStatus {
  worker_id: string;
  full_name: string;
  job_title?: string | null;
  curp?: string | null;
  imss_number?: string | null;
  is_compliant: boolean;
  blocking_reasons: string[];
  has_valid_dc3: boolean;
  dc3_document_id?: string | null;
  dc3_expires_at?: string | null;
  has_valid_medical: boolean;
  medical_document_id?: string | null;
  medical_expires_at?: string | null;
  contractor_active: boolean;
}

export interface PermitWorkerDetail {
  id: string; // UUID
  tenant_id?: string;
  permit_id: string;
  worker_id: string;
  compliance_status: 'compliant' | 'non_compliant' | 'warning' | string;
  compliance_snapshot?: WorkerComplianceSnapshot;
  assigned_by?: string | null;
  assigned_at?: string;
  created_at?: string;
  full_name?: string;
  worker_name?: string;
  job_title?: string | null;
  curp?: string | null;
  imss_number?: string | null;
  avatar_r2_key?: string | null;
  is_compliant?: boolean;
}

export interface PermitChecklistItem {
  id: string; // UUID
  tenant_id?: string;
  permit_id?: string;
  item_key: string;
  nom_standard?: string;
  title?: string;
  description?: string | null;
  category: 'safety' | 'ppe' | 'equipment' | 'atmospheric' | 'emergency' | string;
  is_mandatory?: boolean;
  is_verified?: boolean;
  is_checked?: boolean; // alias for is_verified
  verified_by?: string | null;
  verified_by_user_id?: string | null;
  verified_by_name?: string | null;
  verified_at?: string | null;
  measurements?: Record<string, any> | null;
  measurement_value?: string | null;
  notes?: string | null;
  display_order?: number;
  created_at?: string;
  updated_at?: string;
}

export interface PermitSignatureItem {
  id: string; // UUID
  tenant_id?: string;
  permit_id: string;
  user_id?: string;
  signer_user_id?: string;
  signer_name?: string;
  signer_full_name?: string;
  role: PermitSignatureRole;
  signature_data?: string | null;
  signature_hash?: string | null;
  notes?: string | null;
  signed_at: string;
  created_at?: string;
}

export interface WorkPermitItem {
  id: string; // UUID
  tenant_id?: string;
  plant_id?: string;
  plant_name?: string;
  permit_number: string;
  permit_type: PermitType;
  status: PermitStatus;
  contractor_id: string;
  contractor_name?: string;
  work_area: string;
  description: string;
  hazard_description?: string | null;
  start_time: string; // ISO 8601
  end_time: string; // ISO 8601
  created_by_user_id: string;
  created_by_name?: string;
  approved_by_user_id?: string | null;
  approved_by_name?: string | null;
  approved_at?: string | null;
  suspended_at?: string | null;
  closed_at?: string | null;
  cancelled_at?: string | null;
  status_reason?: string | null;
  suspension_reason?: string | null;
  closure_notes?: string | null;
  created_at: string;
  updated_at: string;
  worker_count?: number;
  workers?: PermitWorkerDetail[];
  checklist?: PermitChecklistItem[];
  checklists?: PermitChecklistItem[]; // alias for compatibility
  signatures?: PermitSignatureItem[];
}

export interface CreatePermitDraftPayload {
  plant_id?: string;
  contractor_id?: string;
  permit_type: PermitType;
  work_area: string;
  description: string;
  hazard_description?: string;
  start_time: string;
  end_time: string;
  notes?: string;
}

export interface AssignPermitWorkersPayload {
  worker_ids: string[];
  action?: 'assign' | 'remove';
}

export interface UpdatePermitChecklistItemPayload {
  item_key: string;
  item_code?: string;
  is_verified?: boolean;
  is_checked?: boolean;
  notes?: string;
  measurements?: Record<string, any>;
  measurement_value?: string;
}

export interface UpdatePermitChecklistPayload {
  items: UpdatePermitChecklistItemPayload[];
}

export interface ApprovePermitPayload {
  notes?: string;
  signature_data?: string;
}

export interface TransitionPermitStatusPayload {
  action: 'submit' | 'suspend' | 'resume' | 'close' | 'cancel';
  reason?: string;
  notes?: string;
  signature_data?: string;
}

export interface PermitFilterParams {
  plant_id?: string;
  contractor_id?: string;
  status?: PermitStatus | string;
  permit_type?: PermitType | string;
  start_date?: string;
  end_date?: string;
  from?: string;
  to?: string;
  search?: string;
  page?: number;
  per_page?: number;
}

// ---------------------------------------------------------------------------
// Safety Inspections & EHS Audits Types (TDD)
// ---------------------------------------------------------------------------
export type InspectionType = 'routine' | 'planned' | 'special' | 'regulatory';
export type InspectionStatus = 'draft' | 'in_progress' | 'completed' | 'requires_capa' | 'cancelled';
export type ItemComplianceStatus = 'compliant' | 'non_compliant' | 'not_applicable';
export type FindingSeverity = 'low' | 'medium' | 'high' | 'critical';
export type CAPAStatus = 'open' | 'in_progress' | 'resolved' | 'verified' | 'overdue';

export interface InspectionAnswerPayload {
  template_item_id: string;
  status: ItemComplianceStatus;
  finding_notes?: string;
  evidence_photo_url?: string;
}

export interface CompleteInspectionInput {
  answers: InspectionAnswerPayload[];
}

export interface InspectionActionPlanDto {
  id: string;
  description: string;
  severity: FindingSeverity;
  status: CAPAStatus;
  due_date: string;
  assigned_to_user_id?: string;
  contractor_id?: string;
}

export interface InspectionTemplateItemDto {
  id: string;
  template_id: string;
  category: string;
  question: string;
  guidance_notes?: string;
  weight: number;
  critical_risk: boolean;
  sort_order: number;
}

export interface InspectionTemplateDto {
  id: string;
  title: string;
  description?: string;
  category: string;
  normative_reference?: string;
  is_active: boolean;
  items?: InspectionTemplateItemDto[];
}

export interface SafetyInspectionItemDto {
  id: string;
  inspection_number: string;
  inspection_type: InspectionType;
  status: InspectionStatus;
  area_or_location: string;
  contractor_id?: string;
  inspector_id: string;
  score_percentage?: number | null;
  total_items: number;
  compliant_items: number;
  non_compliant_items: number;
  na_items: number;
  notes?: string;
  created_at: string;
}

export interface CreateInspectionPayload {
  template_id: string;
  inspection_type?: InspectionType;
  area_or_location: string;
  contractor_id?: string;
  notes?: string;
}

// ---------------------------------------------------------------------------
// Safety Incidents & Near-Misses Types (NOM-021-STPS)
// ---------------------------------------------------------------------------
export type IncidentType =
  | 'unsafe_act'
  | 'unsafe_condition'
  | 'near_miss'
  | 'first_aid'
  | 'medical_treatment'
  | 'lost_time_injury'
  | 'fatality';

export type IncidentSeverity = 'minor' | 'moderate' | 'major' | 'catastrophic';

export type IncidentStatus =
  | 'reported'
  | 'investigating'
  | 'actions_pending'
  | 'closed'
  | 'cancelled';

export interface SafetyIncidentItemDto {
  id: string;
  incident_number: string;
  incident_type: IncidentType;
  severity: IncidentSeverity;
  status: IncidentStatus;
  title: string;
  description: string;
  location: string;
  event_timestamp: string;
  reported_by_user_id: string;
  involved_worker_id?: string;
  contractor_id?: string;
  days_lost: number;
  evidence_photo_url?: string;
  root_cause_summary?: string;
  created_at: string;
}

export interface ReportIncidentPayload {
  type: IncidentType;
  severity: IncidentSeverity;
  title: string;
  description: string;
  location: string;
  event_timestamp?: string;
  involved_worker_id?: string;
  contractor_id?: string;
  evidence_photo_url?: string;
}

export interface IncidentMetricsDto {
  total_incidents: number;
  near_misses: number;
  unsafe_conditions: number;
  unsafe_acts: number;
  lost_time_injuries: number;
  total_days_lost: number;
}

// ---------------------------------------------------------------------------
// EPP Management & Dotation Matrix Types (NOM-017-STPS)
// ---------------------------------------------------------------------------
export type EPPCategory =
  | 'head'
  | 'eyes_face'
  | 'hearing'
  | 'respiratory'
  | 'hands'
  | 'body'
  | 'feet'
  | 'fall_arrest';

export type EPPDeliveryStatus =
  | 'delivered_active'
  | 'replaced'
  | 'expired'
  | 'returned';

export interface EPPItemDto {
  id: string;
  code: string;
  name: string;
  category: EPPCategory;
  normative_standard?: string;
  lifespan_days: number;
  stock_quantity: number;
  is_active: boolean;
}

export interface EPPDeliveryDto {
  id: string;
  epp_item_id: string;
  worker_id: string;
  contractor_id?: string;
  delivery_date: string;
  expiration_date: string;
  status: EPPDeliveryStatus;
  delivered_by_user_id: string;
  signature_data?: string;
  notes?: string;
  created_at: string;
}

export interface DeliverEPPPayload {
  epp_item_id: string;
  worker_id: string;
  contractor_id?: string;
  delivery_date?: string;
  signature_data?: string;
  notes?: string;
}

// ---------------------------------------------------------------------------
// Emergency Evacuation & Roll Call Types (NOM-002-STPS)
// ---------------------------------------------------------------------------
export type EvacuationType = 'drill' | 'fire' | 'gas_leak' | 'earthquake' | 'spill' | 'other';
export type EvacuationStatus = 'active' | 'all_clear' | 'cancelled';
export type RollCallStatus = 'safe' | 'missing' | 'hospitalized';

export interface MusterPointDto {
  id: string;
  code: string;
  name: string;
  location_description?: string;
  capacity: number;
  is_active: boolean;
}

export interface EmergencyEvacuationDto {
  id: string;
  evacuation_number: string;
  evacuation_type: EvacuationType;
  status: EvacuationStatus;
  description?: string;
  started_at: string;
  ended_at?: string;
  evacuation_time_seconds: number;
  total_headcount: number;
  safe_count: number;
  missing_count: number;
}

export interface StartEvacuationPayload {
  type: EvacuationType;
  description?: string;
}

export interface RollCallCheckInPayload {
  worker_id: string;
  muster_point_id: string;
  notes?: string;
}




export interface CompleteInspectionResult {
  inspection_id: string;
  score_percentage: number;
  status: InspectionStatus;
  requires_capa: boolean;
  generated_capas: InspectionActionPlanDto[];
}




