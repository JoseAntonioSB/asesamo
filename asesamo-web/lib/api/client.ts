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
      'Content-Type': 'application/json',
      Accept: 'application/json',
      ...((options.headers as Record<string, string>) || {}),
    };

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
  ): Promise<{ logs: any[]; meta?: PaginationMeta }> {
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
   * Clears stored credentials and logs out.
   */
  public logout(): void {
    clearAuthSession();
  }
}

// Global singleton instance
export const apiClient = new ApiClient();
