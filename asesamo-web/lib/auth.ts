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

/**
 * Safely extracts the user role from a JWT token payload without external libraries.
 * Compatible with Edge runtime, Node.js, and browser environments.
 */
export function getRoleFromJwt(token?: string | null): string | null {
  if (!token) return null;
  try {
    const parts = token.split('.');
    if (parts.length < 2) return null;
    const base64Url = parts[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    let jsonPayload = '';
    if (typeof atob === 'function') {
      jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
    } else if (typeof Buffer !== 'undefined') {
      jsonPayload = Buffer.from(base64, 'base64').toString('utf-8');
    }
    const parsed = JSON.parse(jsonPayload);
    return parsed.role || null;
  } catch {
    return null;
  }
}

/**
 * Maps user roles to their corresponding natural Home Page in ASESAMO.
 */
export function getDefaultHomeForRole(role?: string | null): string {
  switch (role) {
    case 'contractor_admin':
      return '/contratistas';
    case 'ehs_manager':
      return '/ehs/aprobaciones';
    case 'worker':
      return '/credencial';
    case 'plant_admin':
    case 'tenant_admin':
      return '/reportes';
    case 'guard':
    default:
      return '/caseta';
  }
}

