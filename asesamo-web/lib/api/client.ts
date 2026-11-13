/**
 * ASESAMO — Centralized HTTP Client
 * Location: asesamo-web/lib/api/client.ts
 */

import {
  ApiResponse,
  ApiErrorPayload,
  PaginationMeta,
  LoginRequest,
  LoginResponseData,
  AccessPointResponse,
  CheckInRequest,
  CheckInOutput,
  CheckOutRequest,
  CheckOutOutput,
  AccessLogsQueryParams,
  AccessLogItem,
  DocumentItem,
  DocumentTypeItem,
  UploadDocumentParams,
  ReviewDocumentPayload,
  WorkerCredentialResponse,
  ContractorItem,
  ContractorWorkerItem,
  InviteContractorRequest,
  UpdateContractorStatusPayload,
  AddContractorWorkerPayload,
  ToggleWorkerStatusPayload,
  ReportMetricsResponse,
  ReportFilterParams,
  UserProfileDetail,
  UpdateProfilePayload,
  AvatarUploadResponse,
  WorkPermitItem,
  WorkerComplianceStatus,
  PermitChecklistItem,
  PermitSignatureItem,
  PermitType,
  PermitStatus,
  PermitWorkerDetail,
  CreatePermitDraftPayload,
  AssignPermitWorkersPayload,
  UpdatePermitChecklistPayload,
  ApprovePermitPayload,
  TransitionPermitStatusPayload,
  PermitFilterParams,
  CompleteInspectionInput,
  CompleteInspectionResult,
  InspectionTemplateDto,
  SafetyInspectionItemDto,
  CreateInspectionPayload,
  IncidentMetricsDto,
  SafetyIncidentItemDto,
  ReportIncidentPayload,
  EPPItemDto,
  EPPDeliveryDto,
  DeliverEPPPayload,
  MusterPointDto,
  EmergencyEvacuationDto,
  StartEvacuationPayload,
  RollCallCheckInPayload,
} from './types';
import { getStoredToken, setStoredToken, setStoredUser, clearAuthSession } from '../auth';

/**
 * Standard typed application error carrying backend error code and HTTP status.
 */
export class ApiError extends Error {
  public readonly code: string;
  public readonly status: number;
  public readonly details?: unknown;

  constructor(code: string, message: string, status: number = 400, details?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
    this.details = details;
    Object.setPrototypeOf(this, ApiError.prototype);
  }

  static fromPayload(payload: ApiErrorPayload, status: number): ApiError {
    return new ApiError(
      payload.error?.code || 'API_ERROR',
      payload.error?.message || 'Error en la petición al servidor.',
      status
    );
  }

  static network(status: number = 0, message?: string): ApiError {
    return new ApiError(
      'NETWORK_ERROR',
      message || 'No se pudo conectar con el servidor. Verifique su conexión o que el servicio esté activo.',
      status
    );
  }
}

/**
 * Normalizes API Base URL ensuring trailing slashes and '/api/v1' are handled correctly.
 */
export function normalizeApiUrl(rawUrl?: string): string {
  const defaultUrl = 'http://localhost:8080/api/v1';
  if (!rawUrl || !rawUrl.trim()) {
    return defaultUrl;
  }
  const clean = rawUrl.trim().replace(/\/+$/, '');
  return clean.endsWith('/api/v1') ? clean : `${clean}/api/v1`;
}

export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = normalizeApiUrl(baseUrl || process.env.NEXT_PUBLIC_API_URL);
  }

  /**
   * Internal request executor with automatic envelope unwrapping and bearer auth.
   */
  public async request<T, M = unknown>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T, M>> {
    const url = `${this.baseUrl}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`;

    // Default headers
    const headers: Record<string, string> = {
      Accept: 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    };

    if (!(options.body instanceof FormData)) {
      if (!headers['Content-Type']) {
        headers['Content-Type'] = 'application/json';
      }
    } else {
      delete headers['Content-Type'];
    }

    // Bearer token injection
    const token = getStoredToken();
    if (token && !headers['Authorization']) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    let response: Response;
    try {
      response = await fetch(url, {
        ...options,
        headers,
      });
    } catch {
      throw ApiError.network(0, `Fallo de conexión al conectar con ${url}`);
    }

    // Handle 401 Unauthorized globally
    if (response.status === 401 && !endpoint.includes('/auth/login')) {
      clearAuthSession();
      if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.href = '/login?error=session_expired';
      }
    }

    // Parse JSON
    let body: any;
    const text = await response.text();
    try {
      body = text ? JSON.parse(text) : {};
    } catch {
      if (!response.ok) {
        throw new ApiError(
          'INVALID_SERVER_RESPONSE',
          `El servidor respondió con código ${response.status} sin formato JSON válido.`,
          response.status
        );
      }
      body = {};
    }

    // Handle HTTP Error Envelopes
    if (!response.ok) {
      if (body && body.error && typeof body.error.code === 'string') {
        throw ApiError.fromPayload(body as ApiErrorPayload, response.status);
      }
      throw new ApiError(
        'HTTP_ERROR',
        `Error del servidor (${response.status}): ${response.statusText}`,
        response.status
      );
    }

    // Envelope unwrapping: returns { data, meta }
    return {
      data: body.data as T,
      meta: body.meta as M | undefined,
    };
  }

  // ==========================================
  // Public API Methods
  // ==========================================

  /**
   * POST /api/v1/auth/login
   * Authenticates security guard and automatically persists session.
   */
  public async login(email: string, password: string): Promise<LoginResponseData> {
    const payload: LoginRequest = { email: email.trim(), password };
    const res = await this.request<LoginResponseData>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    if (res.data?.token) {
      setStoredToken(res.data.token);
      if (res.data.user) {
        setStoredUser(res.data.user);
      }
    }

    return res.data;
  }

  /**
   * GET /api/v1/access-points
   * Retrieves active caseta access points for the guard's tenant/plant.
   */
  public async getAccessPoints(): Promise<AccessPointResponse[]> {
    const res = await this.request<AccessPointResponse[]>('/access-points', {
      method: 'GET',
    });
    return res.data || [];
  }

  /**
   * POST /api/v1/access/check-in
   * Evaluates worker and document compliance at an access point.
   */
  public async checkIn(
    workerId: string,
    accessPointId: string,
    notes?: string
  ): Promise<CheckInOutput> {
    const payload: CheckInRequest = {
      worker_id: workerId.trim(),
      access_point_id: accessPointId.trim(),
      notes: notes?.trim() || undefined,
    };

    const res = await this.request<CheckInOutput>('/access/check-in', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    return res.data;
  }

  /**
   * POST /api/v1/access/check-out
   * Registers worker exit from plant facility.
   */
  public async checkOut(
    workerId: string,
    accessPointId: string,
    plantId?: string
  ): Promise<CheckOutOutput> {
    const payload: CheckOutRequest = {
      worker_id: workerId.trim(),
      access_point_id: accessPointId.trim(),
      plant_id: plantId?.trim() || undefined,
    };

    const res = await this.request<CheckOutOutput>('/access/check-out', {
      method: 'POST',
      body: JSON.stringify(payload),
    });

    return res.data;
  }

  /**
   * GET /api/v1/access-logs
   * Retrieves paginated audit logs with search filters.
   */
  public async getAccessLogs(
    params: AccessLogsQueryParams = {}
  ): Promise<{ logs: AccessLogItem[]; meta?: PaginationMeta }> {
    const query = new URLSearchParams();
    if (params.plant_id) query.set('plant_id', params.plant_id);
    if (params.access_point_id) query.set('access_point_id', params.access_point_id);
    if (params.worker_id) query.set('worker_id', params.worker_id);
    if (params.result) query.set('result', params.result);
    if (params.from) query.set('from', params.from);
    if (params.to) query.set('to', params.to);
    if (params.page) query.set('page', params.page.toString());
    if (params.per_page) query.set('per_page', params.per_page.toString());

    const qs = query.toString();
    const endpoint = `/access-logs${qs ? `?${qs}` : ''}`;

    const res = await this.request<any[], PaginationMeta>(endpoint, {
      method: 'GET',
    });

    return {
      logs: res.data || [],
      meta: res.meta,
    };
  }

  /**
   * POST /api/v1/documents/upload
   * Uploads compliance document via multipart/form-data.
   */
  public async uploadDocument(params: UploadDocumentParams): Promise<DocumentItem> {
    const formData = new FormData();
    formData.append('file', params.file);
    formData.append('document_type_code', params.documentTypeCode);
    if (params.issuedAt) formData.append('issued_at', params.issuedAt);
    if (params.expiresAt) formData.append('expires_at', params.expiresAt);
    if (params.workerId) formData.append('worker_id', params.workerId);
    if (params.contractorId) formData.append('contractor_id', params.contractorId);

    const res = await this.request<DocumentItem>('/documents/upload', {
      method: 'POST',
      body: formData,
    });
    return res.data;
  }

  /**
   * GET /api/v1/documents/my-documents
   * Retrieves documents for the authenticated worker or contractor.
   */
  public async getMyDocuments(params?: { worker_id?: string; contractor_id?: string }): Promise<DocumentItem[]> {
    const query = new URLSearchParams();
    if (params?.worker_id) query.set('worker_id', params.worker_id);
    if (params?.contractor_id) query.set('contractor_id', params.contractor_id);
    const qs = query.toString();
    const endpoint = `/documents/my-documents${qs ? `?${qs}` : ''}`;

    const res = await this.request<DocumentItem[]>(endpoint, {
      method: 'GET',
    });
    return res.data || [];
  }

  /**
   * GET /api/v1/documents/types
   * Retrieves active document types catalog.
   */
  public async getDocumentTypes(): Promise<DocumentTypeItem[]> {
    const res = await this.request<DocumentTypeItem[]>('/documents/types', {
      method: 'GET',
    });
    return res.data || [];
  }

  /**
   * GET /api/v1/documents/pending
   * Retrieves documents awaiting review or filtered by status for EHS review tray.
   */
  public async getPendingDocuments(status?: string): Promise<DocumentItem[]> {
    const query = new URLSearchParams();
    if (status) query.set('status', status);
    const qs = query.toString();
    const endpoint = `/documents/pending${qs ? `?${qs}` : ''}`;

    const res = await this.request<DocumentItem[]>(endpoint, {
      method: 'GET',
    });
    return res.data || [];
  }

  /**
   * POST /api/v1/documents/:id/review
   * Approves or rejects a document with mandatory rejection reason if rejected.
   */
  public async reviewDocument(
    docId: string,
    action: 'approve' | 'reject',
    rejectionReason?: string
  ): Promise<DocumentItem> {
    const payload: ReviewDocumentPayload = {
      action,
      rejection_reason: rejectionReason?.trim() || undefined,
    };

    const res = await this.request<DocumentItem>(`/documents/${docId}/review`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * Returns full download URL for a document.
   */
  public getDownloadUrl(docId: string): string {
    return `${this.baseUrl}/documents/${docId}/download`;
  }

  /**
   * PATCH /api/v1/users/me/theme
   * Persists the user's theme preference in the database.
   */
  public async updateTheme(theme: 'light' | 'dark' | 'system'): Promise<{ theme: string }> {
    const res = await this.request<{ theme: string }>('/users/me/theme', {
      method: 'PATCH',
      body: JSON.stringify({ theme }),
    });
    return res.data;
  }

  /**
   * GET /api/v1/users/me
   * Retrieves full profile details of authenticated user.
   */
  public async getProfile(): Promise<UserProfileDetail> {
    const res = await this.request<UserProfileDetail>('/users/me', {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * PATCH /api/v1/users/me
   * Updates user name, phone, or password.
   */
  public async updateProfile(payload: UpdateProfilePayload): Promise<UserProfileDetail> {
    const res = await this.request<UserProfileDetail>('/users/me', {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/users/me/avatar
   * Uploads or updates current user's profile photo.
   */
  public async uploadMyAvatar(file: File): Promise<AvatarUploadResponse> {
    const formData = new FormData();
    formData.append('avatar', file);
    const res = await this.request<AvatarUploadResponse>('/users/me/avatar', {
      method: 'POST',
      body: formData,
    });
    return res.data;
  }

  /**
   * POST /api/v1/users/:id/avatar
   * Uploads or updates profile photo for a specific worker/user.
   */
  public async uploadUserAvatar(userId: string, file: File): Promise<AvatarUploadResponse> {
    const formData = new FormData();
    formData.append('avatar', file);
    const res = await this.request<AvatarUploadResponse>(`/users/${encodeURIComponent(userId)}/avatar`, {
      method: 'POST',
      body: formData,
    });
    return res.data;
  }

  /**
   * Returns full public URL for user's avatar image with optional cache buster.
   */
  public getUserAvatarUrl(userId: string): string {
    return `${this.baseUrl}/users/${encodeURIComponent(userId)}/avatar`;
  }

  /**
   * GET /api/v1/workers/me/credential
   * Retrieves current authenticated worker's live credential and QR code.
   */
  public async getMyCredential(plantId?: string): Promise<WorkerCredentialResponse> {
    const qs = plantId ? `?plant_id=${encodeURIComponent(plantId)}` : '';
    const res = await this.request<WorkerCredentialResponse>(`/workers/me/credential${qs}`, {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * GET /api/v1/workers/:id/credential
   * Retrieves a specific worker's live credential, badge info, and QR code payload.
   */
  public async getWorkerCredential(
    workerId: string,
    plantId?: string
  ): Promise<WorkerCredentialResponse> {
    const qs = plantId ? `?plant_id=${encodeURIComponent(plantId)}` : '';
    const res = await this.request<WorkerCredentialResponse>(
      `/workers/${encodeURIComponent(workerId)}/credential${qs}`,
      {
        method: 'GET',
      }
    );
    return res.data;
  }

  /**
   * POST /api/v1/contractors/invite
   * Invites and registers a new contractor company and its admin user.
   */
  public async inviteContractor(payload: InviteContractorRequest): Promise<ContractorItem> {
    const res = await this.request<ContractorItem>('/contractors/invite', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * GET /api/v1/contractors?plant_id=...
   * Lists all contractors affiliated with the given plant.
   */
  public async listPlantContractors(plantId: string): Promise<ContractorItem[]> {
    const res = await this.request<ContractorItem[]>(`/contractors?plant_id=${encodeURIComponent(plantId)}`, {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * PATCH /api/v1/contractors/:id/status
   * Updates contractor company status in a plant (active, suspended, blocked, pending).
   */
  public async updateContractorStatus(
    contractorId: string,
    payload: UpdateContractorStatusPayload
  ): Promise<{ message: string }> {
    const res = await this.request<{ message: string }>(`/contractors/${encodeURIComponent(contractorId)}/status`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * GET /api/v1/contractors/my-company
   * Retrieves company information for the logged-in contractor admin.
   */
  public async getMyContractorCompany(): Promise<ContractorItem> {
    const res = await this.request<ContractorItem>('/contractors/my-company', {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * GET /api/v1/contractors/workers?contractor_id=...
   * Lists all crew workers registered under a contractor.
   */
  public async listContractorWorkers(contractorId?: string): Promise<ContractorWorkerItem[]> {
    const qs = contractorId ? `?contractor_id=${encodeURIComponent(contractorId)}` : '';
    const res = await this.request<ContractorWorkerItem[]>(`/contractors/workers${qs}`, {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * POST /api/v1/contractors/workers
   * Registers a new worker in the contractor's crew.
   */
  public async addContractorWorker(payload: AddContractorWorkerPayload): Promise<ContractorWorkerItem> {
    const res = await this.request<ContractorWorkerItem>('/contractors/workers', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * PATCH /api/v1/contractors/workers/:id/status
   * Activates or deactivates a worker in the crew.
   */
  public async toggleContractorWorkerStatus(
    workerId: string,
    isActive: boolean,
    contractorId?: string
  ): Promise<{ worker_id: string; is_active: boolean; message: string }> {
    const payload: ToggleWorkerStatusPayload = {
      is_active: isActive,
      contractor_id: contractorId,
    };
    const res = await this.request<{ worker_id: string; is_active: boolean; message: string }>(
      `/contractors/workers/${encodeURIComponent(workerId)}/status`,
      {
        method: 'PATCH',
        body: JSON.stringify(payload),
      }
    );
    return res.data;
  }

  /**
   * GET /api/v1/reports/metrics
   * Retrieves aggregated KPI analytics, daily timelines, and top rejection reasons.
   */
  public async getReportMetrics(params?: ReportFilterParams): Promise<ReportMetricsResponse> {
    const query = new URLSearchParams();
    if (params?.plant_id) query.append('plant_id', params.plant_id);
    if (params?.from) query.append('from', params.from);
    if (params?.to) query.append('to', params.to);

    const qs = query.toString() ? `?${query.toString()}` : '';
    const res = await this.request<ReportMetricsResponse>(`/reports/metrics${qs}`, {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * Returns full URL for downloading access logs CSV.
   */
  public getExportAccessLogsUrl(params?: ReportFilterParams): string {
    const query = new URLSearchParams();
    if (params?.plant_id) query.append('plant_id', params.plant_id);
    if (params?.from) query.append('from', params.from);
    if (params?.to) query.append('to', params.to);

    const qs = query.toString() ? `?${query.toString()}` : '';
    return `${this.baseUrl}/reports/export/access-logs${qs}`;
  }

  /**
   * Returns full URL for downloading worker compliance matrix CSV.
   */
  public getExportComplianceUrl(plantId?: string): string {
    const qs = plantId ? `?plant_id=${encodeURIComponent(plantId)}` : '';
    return `${this.baseUrl}/reports/export/compliance${qs}`;
  }

  /**
   * Downloads access logs CSV via authenticated fetch.
   */
  public async downloadAccessLogsCSV(params?: ReportFilterParams): Promise<Blob> {
    const url = this.getExportAccessLogsUrl(params);
    const token = getStoredToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(url, { headers });
    if (!response.ok) {
      throw new ApiError('EXPORT_ERROR', 'No se pudo generar el reporte de accesos en CSV.', response.status);
    }
    return response.blob();
  }

  /**
   * Downloads compliance matrix CSV via authenticated fetch.
   */
  public async downloadComplianceCSV(plantId?: string): Promise<Blob> {
    const url = this.getExportComplianceUrl(plantId);
    const token = getStoredToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(url, { headers });
    if (!response.ok) {
      throw new ApiError('EXPORT_ERROR', 'No se pudo generar la matriz de cumplimiento STPS.', response.status);
    }
    return response.blob();
  }

  // ==========================================
  // 10. High-Risk Work Permits (PTAR Digital)
  // ==========================================

  /**
   * GET /api/v1/permits
   * Lists high-risk work permits with optional filtering.
   */
  public async getPermits(
    params: PermitFilterParams = {}
  ): Promise<{ permits: WorkPermitItem[]; meta?: PaginationMeta }> {
    const query = new URLSearchParams();
    if (params.plant_id) query.set('plant_id', params.plant_id);
    if (params.contractor_id) query.set('contractor_id', params.contractor_id);
    if (params.status) query.set('status', params.status);
    if (params.permit_type) query.set('permit_type', params.permit_type);
    if (params.start_date || params.from) query.set('start_date', params.start_date || params.from || '');
    if (params.end_date || params.to) query.set('end_date', params.end_date || params.to || '');
    if (params.page) query.set('page', params.page.toString());
    if (params.per_page) query.set('per_page', params.per_page.toString());

    const qs = query.toString();
    const res = await this.request<WorkPermitItem[], PaginationMeta>(`/permits${qs ? `?${qs}` : ''}`, {
      method: 'GET',
    });
    return {
      permits: res.data || [],
      meta: res.meta,
    };
  }

  /**
   * GET /api/v1/permits/:id
   * Retrieves single permit with details, crew, checklist, and signatures.
   */
  public async getPermit(id: string): Promise<WorkPermitItem> {
    const res = await this.request<WorkPermitItem>(`/permits/${encodeURIComponent(id)}`, {
      method: 'GET',
    });
    return res.data;
  }

  /**
   * POST /api/v1/permits
   * Creates a new draft work permit.
   */
  public async createPermit(payload: CreatePermitDraftPayload): Promise<WorkPermitItem> {
    const res = await this.request<WorkPermitItem>('/permits', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * Evaluates candidate crew workers against Candado Inteligente rules.
   * If backend dedicated endpoint /permits/workers/compliance is available, queries it.
   * Otherwise evaluates compliance using contractor workers list and document statuses.
   */
  public async checkCrewCompliance(
    contractorId: string,
    permitType: PermitType
  ): Promise<WorkerComplianceStatus[]> {
    const qs = `?contractor_id=${encodeURIComponent(contractorId)}&permit_type=${encodeURIComponent(permitType)}`;
    try {
      const res = await this.request<WorkerComplianceStatus[]>(`/permits/workers/compliance${qs}`, {
        method: 'GET',
      });
      if (res.data && Array.isArray(res.data) && res.data.length > 0) {
        return res.data;
      }
    } catch {
      // Backend dedicated compliance endpoint optional; evaluate client-side
    }

    const workers = await this.listContractorWorkers(contractorId);
    const results: WorkerComplianceStatus[] = [];

    const hazardLabels: Record<PermitType, { name: string; nom: string }> = {
      heights: { name: 'trabajos en alturas', nom: 'NOM-009-STPS-2011' },
      confined_space: { name: 'espacios confinados', nom: 'NOM-033-STPS-2015' },
      hot_work: { name: 'corte y soldadura', nom: 'NOM-027-STPS-2008' },
      electrical_loto: { name: 'seguridad eléctrica y LOTO', nom: 'NOM-029-STPS-2011' },
    };

    const hazardInfo = hazardLabels[permitType] || { name: 'trabajo de alto riesgo', nom: 'STPS' };

    for (const w of workers) {
      const blockingReasons: string[] = [];
      let hasValidDC3 = false;
      let dc3DocId: string | null = null;
      let dc3ExpiresAt: string | null = null;
      let hasValidMedical = false;
      let medicalDocId: string | null = null;
      let medicalExpiresAt: string | null = null;

      try {
        const docs = await this.getMyDocuments({ worker_id: w.worker_id });
        const now = new Date();

        for (const doc of docs) {
          const typeCode = (doc.document_type_code || '').toUpperCase();
          const isApproved = doc.status === 'approved';
          const notExpired = !doc.expires_at || new Date(doc.expires_at) > now;

          if (typeCode === 'MEDICAL') {
            medicalDocId = doc.id;
            medicalExpiresAt = doc.expires_at || null;
            if (isApproved && notExpired) {
              hasValidMedical = true;
            }
          }

          const isDC3 = typeCode.startsWith('DC3') || typeCode === 'DC-3';
          if (isDC3) {
            dc3DocId = doc.id;
            dc3ExpiresAt = doc.expires_at || null;
            if (isApproved && notExpired) {
              hasValidDC3 = true;
            }
          }
        }
      } catch {
        // Continue evaluation
      }

      if (!w.is_active || !w.user_is_active) {
        blockingReasons.push('El perfil del trabajador se encuentra inactivo.');
      }
      if (!hasValidDC3) {
        blockingReasons.push(`Falta constancia DC-3 vigente para ${hazardInfo.name} (${hazardInfo.nom}).`);
      }
      if (!hasValidMedical) {
        blockingReasons.push('Certificado de aptitud médica no acreditado o expirado.');
      }

      results.push({
        worker_id: w.worker_id,
        full_name: w.full_name,
        job_title: w.job_title || undefined,
        curp: w.curp || undefined,
        imss_number: w.imss_number || undefined,
        is_compliant: blockingReasons.length === 0,
        blocking_reasons: blockingReasons,
        has_valid_dc3: hasValidDC3,
        dc3_document_id: dc3DocId,
        dc3_expires_at: dc3ExpiresAt,
        has_valid_medical: hasValidMedical,
        medical_document_id: medicalDocId,
        medical_expires_at: medicalExpiresAt,
        contractor_active: true,
      });
    }

    return results;
  }

  /**
   * POST /api/v1/permits/:id/workers
   * Assigns or removes crew workers to/from a permit, enforcing Candado Inteligente locks.
   */
  public async assignPermitWorkers(
    permitId: string,
    workerIds: string[],
    action: 'assign' | 'remove' = 'assign'
  ): Promise<WorkPermitItem> {
    const payload: AssignPermitWorkersPayload = {
      worker_ids: workerIds,
      action,
    };
    const res = await this.request<WorkPermitItem>(`/permits/${encodeURIComponent(permitId)}/workers`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * PATCH /api/v1/permits/:id/checklist
   * Updates safety verification checklist items for a permit.
   */
  public async updatePermitChecklist(
    permitId: string,
    payload: UpdatePermitChecklistPayload
  ): Promise<WorkPermitItem> {
    const res = await this.request<WorkPermitItem>(`/permits/${encodeURIComponent(permitId)}/checklist`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/permits/:id/approve
   * Performs EHS digital authorization and transitions permit to 'approved_active'.
   */
  public async approvePermit(
    permitId: string,
    notes?: string,
    signatureData?: string
  ): Promise<WorkPermitItem> {
    const payload: ApprovePermitPayload = {
      notes: notes?.trim() || undefined,
      signature_data: signatureData?.trim() || undefined,
    };
    const res = await this.request<WorkPermitItem>(`/permits/${encodeURIComponent(permitId)}/approve`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/permits/:id/status
   * Transitions permit lifecycle status (submit, suspend, resume, close, cancel).
   */
  public async transitionPermitStatus(
    permitId: string,
    payload: TransitionPermitStatusPayload
  ): Promise<WorkPermitItem> {
    const res = await this.request<WorkPermitItem>(`/permits/${encodeURIComponent(permitId)}/status`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * GET /api/v1/inspections/templates
   * Lists active inspection checklist templates.
   */
  public async getInspectionTemplates(): Promise<InspectionTemplateDto[]> {
    const res = await this.request<InspectionTemplateDto[]>('/inspections/templates');
    return res.data;
  }

  /**
   * GET /api/v1/inspections
   * Lists safety inspections.
   */
  public async getInspections(): Promise<SafetyInspectionItemDto[]> {
    const res = await this.request<SafetyInspectionItemDto[]>('/inspections');
    return res.data;
  }

  /**
   * POST /api/v1/inspections
   * Creates a new draft inspection.
   */
  public async createInspection(payload: CreateInspectionPayload): Promise<SafetyInspectionItemDto> {
    const res = await this.request<SafetyInspectionItemDto>('/inspections', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/inspections/:id/complete
   * Submits answers for a field safety inspection and calculates score & CAPA actions.
   */
  public async completeInspection(
    inspectionId: string,
    payload: CompleteInspectionInput
  ): Promise<CompleteInspectionResult> {
    const res = await this.request<CompleteInspectionResult>(
      `/inspections/${encodeURIComponent(inspectionId)}/complete`,
      {
        method: 'POST',
        body: JSON.stringify(payload),
      }
    );
    return res.data;
  }

  /**
   * GET /api/v1/incidents
   * Lists reported safety incidents.
   */
  public async getIncidents(): Promise<SafetyIncidentItemDto[]> {
    const res = await this.request<SafetyIncidentItemDto[]>('/incidents');
    return res.data;
  }

  /**
   * POST /api/v1/incidents
   * Reports a new safety incident, near-miss, or unsafe condition.
   */
  public async reportIncident(payload: ReportIncidentPayload): Promise<SafetyIncidentItemDto> {
    const res = await this.request<SafetyIncidentItemDto>('/incidents', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * GET /api/v1/incidents/metrics
   * Retrieves STPS incident metrics.
   */
  public async getIncidentMetrics(): Promise<IncidentMetricsDto> {
    const res = await this.request<IncidentMetricsDto>('/incidents/metrics');
    return res.data;
  }

  /**
   * GET /api/v1/epp/catalog
   * Lists available EPP items.
   */
  public async getEPPCatalog(): Promise<EPPItemDto[]> {
    const res = await this.request<EPPItemDto[]>('/epp/catalog');
    return res.data;
  }

  /**
   * GET /api/v1/epp/deliveries
   * Lists EPP delivery logs.
   */
  public async getEPPDeliveries(): Promise<EPPDeliveryDto[]> {
    const res = await this.request<EPPDeliveryDto[]>('/epp/deliveries');
    return res.data;
  }

  /**
   * POST /api/v1/epp/deliveries
   * Records a new personal protective equipment delivery.
   */
  public async deliverEPP(payload: DeliverEPPPayload): Promise<EPPDeliveryDto> {
    const res = await this.request<EPPDeliveryDto>('/epp/deliveries', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * GET /api/v1/evacuations/muster-points
   * Lists active muster safety points.
   */
  public async getMusterPoints(): Promise<MusterPointDto[]> {
    const res = await this.request<MusterPointDto[]>('/evacuations/muster-points');
    return res.data;
  }

  /**
   * GET /api/v1/evacuations/active
   * Gets current active emergency evacuation or drill.
   */
  public async getActiveEvacuation(): Promise<EmergencyEvacuationDto | null> {
    const res = await this.request<EmergencyEvacuationDto | null>('/evacuations/active');
    return res.data;
  }

  /**
   * POST /api/v1/evacuations/start
   * Triggers an emergency evacuation or safety drill.
   */
  public async startEvacuation(payload: StartEvacuationPayload): Promise<EmergencyEvacuationDto> {
    const res = await this.request<EmergencyEvacuationDto>('/evacuations/start', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/evacuations/:id/check-in
   * Marks a worker as safe at a muster point during evacuation roll call.
   */
  public async checkInEvacuee(evacId: string, payload: RollCallCheckInPayload): Promise<{ success: boolean; message: string }> {
    const res = await this.request<{ success: boolean; message: string }>(`/evacuations/${encodeURIComponent(evacId)}/check-in`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    return res.data;
  }

  /**
   * POST /api/v1/evacuations/:id/end
   * Concludes evacuation and records all-clear time.
   */
  public async endEvacuation(evacId: string): Promise<EmergencyEvacuationDto> {
    const res = await this.request<EmergencyEvacuationDto>(`/evacuations/${encodeURIComponent(evacId)}/end`, {
      method: 'POST',
    });
    return res.data;
  }




  /**
   * Clears stored credentials and logs out.
   */
  public logout(): void {
    clearAuthSession();
  }
}

// Global singleton instance
export const apiClient = new ApiClient();
