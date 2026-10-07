# M4 Architectural Specification: Auth Flow, API Client Layer & Route Protection

**Agent:** `teamwork_preview_explorer_m4_2`  
**Milestone:** M4 (Frontend Web Interface — `asesamo-web`)  
**Scope:** API Client Layer (`lib/api.ts`, `lib/api/`), Guard Authentication (`app/login/page.tsx`), Route Protection (`middleware.ts`), and Auth State Management (`contexts/auth-context.tsx`, `lib/auth.ts`).

---

## 1. Executive Summary & Architectural Overview

The ASESAMO frontend web application (`asesamo-web`) is a Next.js 14 App Router application dedicated to industrial gate security personnel and EHS compliance managers. This document provides the complete, production-ready specification and code implementation blueprints for:

1. **Centralized HTTP Client (`lib/api/client.ts`, `lib/api/types.ts`, `lib/api.ts`)**:
   - Base URL normalization supporting both `http://localhost:8080` and `http://localhost:8080/api/v1` via `NEXT_PUBLIC_API_URL`.
   - Envelope unwrapping for backend standard responses: `{ data: T, meta?: M }`.
   - Structured domain error handling for backend error envelopes: `{ error: { code: string, message: string } }`.
   - Automatic `Authorization: Bearer <JWT>` header injection from secure cookie / localStorage token storage.
   - Core API methods: `login`, `getAccessPoints`, `checkIn`, `checkOut`, and `getAccessLogs`.

2. **Guard Authentication Screen (`app/login/page.tsx`)**:
   - Industrial security UI themed with deep slate backgrounds and emerald accents.
   - Email and password form validation with immediate feedback.
   - Interactive Quick-Fill Demo button populating seeded guard credentials (`guard@asesamo.com` / `Password123!`).
   - High-visibility Spanish error alert banner mapping backend error codes and network failures.
   - Secure token storage in browser cookies (`asesamo_token`) and `localStorage` with automated redirect to `/caseta`.

3. **Edge & Client Route Protection (`middleware.ts`, `contexts/auth-context.tsx`)**:
   - Next.js Edge Middleware intercepting all navigation requests.
   - Unauthenticated guards navigating to `/caseta` or `/dashboard` are redirected to `/login?redirect=...`.
   - Authenticated guards navigating to `/login` are automatically redirected to `/caseta`.
   - Global `AuthProvider` and `useAuth()` hook providing reactive user state, login orchestration, and secure logout.

---

## 2. API Contract Alignment with `asesamo-api`

The API client directly aligns with the Go backend routes defined in `asesamo-api/internal/adapter/http/router.go`:

| HTTP Method | Route | Auth Required | Backend DTO / Entity | Purpose |
|---|---|---|---|---|
| `POST` | `/api/v1/auth/login` | No (Rate-limited: 10 req/min) | `dto.LoginRequest` $\to$ `dto.LoginResponseData` | Security guard authentication |
| `GET` | `/api/v1/access-points` | Yes (`Bearer <token>`) | `dto.AccessPointResponse[]` | Active caseta checkpoints catalog |
| `POST` | `/api/v1/access/check-in` | Yes (`Bearer <token>`) | `dto.CheckInRequest` $\to$ `access.CheckInOutput` | Gate compliance evaluation & logging |
| `POST` | `/api/v1/access/check-out` | Yes (`Bearer <token>`) | `dto.CheckOutRequest` $\to$ `access.CheckOutOutput` | Worker plant egress registration |
| `GET` | `/api/v1/access-logs` | Yes (`Bearer <token>`) | `dto.PaginationMeta`, `[]entity.AccessLog` | Audit log retrieval |

### Standard Envelope Handling
1. **Success**:
   ```json
   {
     "data": { ... },
     "meta": { "page": 1, "per_page": 20, "total": 1 }
   }
   ```
   The client extracts `data` and exposes `meta` when pagination is present.
2. **Error**:
   ```json
   {
     "error": {
       "code": "INVALID_CREDENTIALS",
       "message": "Credenciales inválidas. Verifique su correo y contraseña."
     }
   }
   ```
   The client parses the error payload into an instance of `ApiError(code, message, status)`.

---

## 3. Concrete Code Specifications

### 3.1 Data Types & Contracts (`lib/api/types.ts`)

```typescript
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

export type UserRole = 'guard' | 'plant_admin' | 'ehs_manager' | 'super_admin' | 'contractor';

export interface UserProfile {
  id: string; // UUID
  tenant_id?: string; // UUID
  email: string;
  full_name: string;
  role: UserRole | string;
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
```

---

### 3.2 Token & Cookie Storage Utility (`lib/auth.ts`)

```typescript
/**
 * ASESAMO — Token & Session Persistence Helper
 * Location: asesamo-web/lib/auth.ts
 */

import { UserProfile } from './api/types';

export const TOKEN_COOKIE_NAME = 'asesamo_token';
export const USER_STORAGE_KEY = 'asesamo_user';

/**
 * Reads token from cookie or localStorage in browser environments.
 */
export function getStoredToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }

  // 1. Check document.cookie
  const match = document.cookie.match(new RegExp('(^|;\\s*)' + TOKEN_COOKIE_NAME + '=([^;]*)'));
  if (match && match[2]) {
    try {
      return decodeURIComponent(match[2]);
    } catch {
      return match[2];
    }
  }

  // 2. Fallback to localStorage
  try {
    return localStorage.getItem(TOKEN_COOKIE_NAME);
  } catch {
    return null;
  }
}

/**
 * Persists token to both browser cookie (for Next.js Middleware) and localStorage.
 */
export function setStoredToken(token: string, daysValid: number = 7): void {
  if (typeof window === 'undefined') return;

  const maxAge = daysValid * 24 * 60 * 60;
  const isSecure = window.location.protocol === 'https:';
  
  // Set cookie with SameSite=Lax
  document.cookie = `${TOKEN_COOKIE_NAME}=${encodeURIComponent(token)}; path=/; max-age=${maxAge}; SameSite=Lax${
    isSecure ? '; Secure' : ''
  }`;

  try {
    localStorage.setItem(TOKEN_COOKIE_NAME, token);
  } catch {
    // LocalStorage quota or privacy sandbox exception
  }
}

/**
 * Removes token from both cookies and localStorage.
 */
export function removeStoredToken(): void {
  if (typeof window === 'undefined') return;

  document.cookie = `${TOKEN_COOKIE_NAME}=; path=/; max-age=0; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax`;

  try {
    localStorage.removeItem(TOKEN_COOKIE_NAME);
  } catch {
    // LocalStorage exception
  }
}

/**
 * Retrieves cached user profile from localStorage.
 */
export function getStoredUser(): UserProfile | null {
  if (typeof window === 'undefined') return null;

  try {
    const raw = localStorage.getItem(USER_STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as UserProfile;
  } catch {
    return null;
  }
}

/**
 * Caches user profile in localStorage.
 */
export function setStoredUser(user: UserProfile): void {
  if (typeof window === 'undefined') return;

  try {
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
  } catch {
    // Ignore storage errors
  }
}

/**
 * Removes cached user profile.
 */
export function removeStoredUser(): void {
  if (typeof window === 'undefined') return;

  try {
    localStorage.removeItem(USER_STORAGE_KEY);
  } catch {
    // Ignore storage errors
  }
}

/**
 * Complete session cleanup.
 */
export function clearAuthSession(): void {
  removeStoredToken();
  removeStoredUser();
}
```

---

### 3.3 API Client Implementation (`lib/api/client.ts`)

```typescript
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
    } catch (err: unknown) {
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
```

---

### 3.4 Convenience Façade (`lib/api.ts`)

```typescript
/**
 * ASESAMO — API Convenience Façade
 * Location: asesamo-web/lib/api.ts
 * Re-exports singleton methods and types for standard import paths (@/lib/api).
 */

import { apiClient } from './api/client';

export * from './api/types';
export * from './api/client';

export const login = (email: string, pass: string) => apiClient.login(email, pass);
export const getAccessPoints = () => apiClient.getAccessPoints();
export const checkIn = (workerId: string, accessPointId: string, notes?: string) =>
  apiClient.checkIn(workerId, accessPointId, notes);
export const checkOut = (workerId: string, accessPointId: string, plantId?: string) =>
  apiClient.checkOut(workerId, accessPointId, plantId);
export const getAccessLogs = (params?: any) => apiClient.getAccessLogs(params);
export const logout = () => apiClient.logout();

export default apiClient;
```

---

### 3.5 Auth State Management & React Context (`contexts/auth-context.tsx`)

```tsx
/**
 * ASESAMO — Global Authentication Context & State
 * Location: asesamo-web/contexts/auth-context.tsx
 */

'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { UserProfile, LoginResponseData } from '@/lib/api/types';
import { apiClient, ApiError } from '@/lib/api/client';
import { getStoredToken, getStoredUser, clearAuthSession } from '@/lib/auth';

interface AuthContextType {
  user: UserProfile | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, pass: string) => Promise<LoginResponseData>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Initialize from cookies / localStorage on mount
  useEffect(() => {
    try {
      const storedToken = getStoredToken();
      const storedUser = getStoredUser();

      if (storedToken) {
        setToken(storedToken);
        setUser(storedUser);
      }
    } catch (e) {
      console.error('Error hydrating auth state:', e);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleLogin = useCallback(
    async (email: string, pass: string): Promise<LoginResponseData> => {
      const data = await apiClient.login(email, pass);
      setToken(data.token);
      setUser(data.user);
      return data;
    },
    []
  );

  const handleLogout = useCallback(() => {
    clearAuthSession();
    setToken(null);
    setUser(null);
    router.push('/login');
  }, [router]);

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token,
        isLoading,
        login: handleLogin,
        logout: handleLogout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
```

---

### 3.6 Edge Route Protection (`middleware.ts`)

```typescript
/**
 * ASESAMO — Edge Route Protection Middleware
 * Location: asesamo-web/middleware.ts
 */

import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import { TOKEN_COOKIE_NAME } from './lib/auth';

export function middleware(request: NextRequest) {
  const token = request.cookies.get(TOKEN_COOKIE_NAME)?.value;
  const { pathname } = request.nextUrl;

  const isAuthRoute = pathname === '/login' || pathname.startsWith('/login/');
  const isProtectedRoute =
    pathname === '/caseta' ||
    pathname.startsWith('/caseta/') ||
    pathname === '/dashboard' ||
    pathname.startsWith('/dashboard/');

  // 1. Root route '/' redirect
  if (pathname === '/') {
    if (token) {
      return NextResponse.redirect(new URL('/caseta', request.url));
    } else {
      return NextResponse.redirect(new URL('/login', request.url));
    }
  }

  // 2. Unauthenticated user trying to access protected gate/caseta routes
  if (isProtectedRoute && !token) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // 3. Authenticated user visiting /login
  if (isAuthRoute && token) {
    return NextResponse.redirect(new URL('/caseta', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * 1. /api/* (backend or internal API routes)
     * 2. /_next/static (static JS, CSS chunks)
     * 3. /_next/image (image optimization files)
     * 4. /favicon.ico and static assets (.svg, .png, .jpg)
     */
    '/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
```

---

### 3.7 Guard Login Screen (`app/login/page.tsx`)

```tsx
/**
 * ASESAMO — Security Guard Login Screen
 * Location: asesamo-web/app/login/page.tsx
 */

'use client';

import React, { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { ApiError } from '@/lib/api/client';
import { ShieldCheck, AlertTriangle, KeyRound, UserCheck, Loader2 } from 'lucide-react';

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isAuthenticated } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorBanner, setErrorBanner] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // If already authenticated, redirect immediately to /caseta
  useEffect(() => {
    if (isAuthenticated) {
      const redirect = searchParams.get('redirect') || '/caseta';
      router.replace(redirect);
    }
  }, [isAuthenticated, router, searchParams]);

  // Handle URL error param (e.g. session expired)
  useEffect(() => {
    const err = searchParams.get('error');
    if (err === 'session_expired') {
      setErrorBanner('Su sesión ha vencido. Por favor ingrese sus credenciales nuevamente.');
    }
  }, [searchParams]);

  // Quick-fill demo credentials
  const handleQuickFill = () => {
    setEmail('guard@asesamo.com');
    setPassword('Password123!');
    setErrorBanner(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorBanner(null);

    // Client-side validation
    if (!email.trim() || !password) {
      setErrorBanner('Por favor complete todos los campos obligatorios.');
      return;
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email.trim())) {
      setErrorBanner('Ingrese un correo electrónico con formato válido.');
      return;
    }

    try {
      setIsSubmitting(true);
      await login(email.trim(), password);
      
      const destination = searchParams.get('redirect') || '/caseta';
      router.push(destination);
    } catch (err: unknown) {
      setIsSubmitting(false);
      if (err instanceof ApiError) {
        if (err.code === 'INVALID_CREDENTIALS') {
          setErrorBanner('Credenciales inválidas. Verifique su correo y contraseña.');
        } else if (err.code === 'TOO_MANY_REQUESTS') {
          setErrorBanner('Demasiados intentos fallidos. Espere un momento antes de volver a intentar.');
        } else if (err.code === 'USER_INACTIVE') {
          setErrorBanner('Su cuenta de oficial de caseta se encuentra inactiva.');
        } else {
          setErrorBanner(err.message || 'Error al iniciar sesión.');
        }
      } else {
        setErrorBanner('No se pudo establecer conexión con el servidor. Intente más tarde.');
      }
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col justify-center items-center px-4 sm:px-6 lg:px-8 selection:bg-emerald-500 selection:text-white">
      {/* Background radial accent */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_var(--tw-gradient-stops))] from-slate-900 via-slate-950 to-slate-950 pointer-events-none" />

      <div className="relative w-full max-w-md space-y-6">
        {/* Brand Header */}
        <div className="text-center space-y-2">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mb-2 shadow-lg shadow-emerald-950/40">
            <ShieldCheck className="w-8 h-8" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white uppercase sm:text-3xl font-mono">
            ASESAMO
          </h1>
          <p className="text-sm text-slate-400">
            Control de Acceso Industrial y Verificación de Cumplimiento
          </p>
          <div className="inline-block px-3 py-1 text-xs font-semibold uppercase tracking-wider text-emerald-400 bg-emerald-950/60 border border-emerald-800/60 rounded-full">
            Módulo de Caseta de Vigilancia
          </div>
        </div>

        {/* Login Card */}
        <div className="bg-slate-900/90 border border-slate-800 shadow-2xl rounded-xl p-6 sm:p-8 backdrop-blur-sm space-y-6">
          {/* Spanish Error Alert Banner */}
          {errorBanner && (
            <div
              role="alert"
              className="bg-red-950/50 border border-red-500/50 rounded-lg p-3.5 flex items-start gap-3 text-red-200 text-sm animate-in fade-in duration-200"
            >
              <AlertTriangle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />
              <div className="flex-1 font-medium">{errorBanner}</div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {/* Email Field */}
            <div className="space-y-1.5">
              <label
                htmlFor="email"
                className="block text-xs font-semibold text-slate-300 uppercase tracking-wider"
              >
                Correo Electrónico
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                disabled={isSubmitting}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="guard@asesamo.com"
                className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-700/80 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/60 focus:border-emerald-500 transition-colors text-sm"
              />
            </div>

            {/* Password Field */}
            <div className="space-y-1.5">
              <label
                htmlFor="password"
                className="block text-xs font-semibold text-slate-300 uppercase tracking-wider"
              >
                Contraseña
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                disabled={isSubmitting}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••••••"
                className="w-full px-3.5 py-2.5 bg-slate-950 border border-slate-700/80 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/60 focus:border-emerald-500 transition-colors text-sm"
              />
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full mt-2 py-3 px-4 rounded-lg bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-semibold text-sm shadow-md shadow-emerald-950/50 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin text-white" />
                  <span>Verificando credenciales...</span>
                </>
              ) : (
                <>
                  <KeyRound className="w-4 h-4" />
                  <span>Ingresar a Caseta</span>
                </>
              )}
            </button>
          </form>

          {/* Quick-Fill Demo Helpers */}
          <div className="pt-4 border-t border-slate-800/80 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-slate-400 font-medium flex items-center gap-1.5">
                <UserCheck className="w-3.5 h-3.5 text-emerald-400" />
                Credenciales de demostración:
              </span>
              <button
                type="button"
                onClick={handleQuickFill}
                disabled={isSubmitting}
                className="text-xs text-emerald-400 hover:text-emerald-300 font-medium underline underline-offset-4 transition-colors cursor-pointer"
              >
                Autocompletar
              </button>
            </div>

            <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-2.5 text-xs font-mono text-slate-400 space-y-1">
              <div className="flex justify-between">
                <span className="text-slate-500">Usuario:</span>
                <span className="text-slate-200">guard@asesamo.com</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Contraseña:</span>
                <span className="text-slate-200">Password123!</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-500">Rol:</span>
                <span className="text-emerald-400">Oficial de Caseta</span>
              </div>
            </div>
          </div>
        </div>

        {/* Industrial Footer Note */}
        <p className="text-center text-xs text-slate-400 font-mono">
          ASESAMO v1.0 • Planta Querétaro Norte • NOM-STPS
        </p>
      </div>
    </div>
  );
}
```

---

### 3.8 Client-Side Guard Component (`components/auth/ProtectedRoute.tsx`)

```tsx
/**
 * ASESAMO — Client Protected Route Wrapper
 * Location: asesamo-web/components/auth/ProtectedRoute.tsx
 */

'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { Loader2 } from 'lucide-react';

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.replace('/login');
    }
  }, [isLoading, isAuthenticated, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center text-slate-400">
        <Loader2 className="w-8 h-8 animate-spin text-emerald-500 mb-2" />
        <p className="text-sm font-mono">Verificando sesión de seguridad...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
```

---

## 4. Verification & Testing Matrix

| Test Scenario | Input / Action | Expected Result | Implementation Path Verified |
|---|---|---|---|
| **V1: Valid Guard Login** | `guard@asesamo.com` / `Password123!` | HTTP 200, JWT token returned, cookie `asesamo_token` set, redirect to `/caseta`. | `api.login` $\to$ `setStoredToken` $\to$ `router.push('/caseta')` |
| **V2: Invalid Credentials** | `guard@asesamo.com` / `WrongPass!` | HTTP 401 `INVALID_CREDENTIALS`. Red banner: *"Credenciales inválidas. Verifique su correo y contraseña."* | `ApiError` catch in `LoginPage` |
| **V3: Rate Limiting Handling** | >10 rapid login attempts | HTTP 429 `TOO_MANY_REQUESTS`. Red banner: *"Demasiados intentos fallidos. Espere un momento..."* | `loginLimiter` handler in Go chi router |
| **V4: Edge Route Protection** | Navigating to `/caseta` with no cookie | HTTP 307 redirect to `/login?redirect=%2Fcaseta`. | Next.js Edge `middleware.ts` |
| **V5: Authenticated Redirection** | Navigating to `/login` with valid cookie | HTTP 307 redirect to `/caseta`. | Next.js Edge `middleware.ts` |
| **V6: Envelope Unwrapping** | Calling `getAccessPoints()` | Returns array of `AccessPointResponse[]` stripped of outer `{ data: ... }`. | `apiClient.request` unwrapping |
| **V7: Bearer Injection** | Calling `checkIn(workerId, apId)` | Request includes `Authorization: Bearer <token>` in headers. | `apiClient.request` header injection |
| **V8: Quick-Fill Interaction** | Click "Autocompletar" button | Form inputs populate with demo credentials instantly. | `handleQuickFill` in `app/login/page.tsx` |
| **V9: Session Expiration** | API returns 401 on protected call | Cookies cleared, user redirected to `/login?error=session_expired`. | Global 401 interceptor in `ApiClient` |
