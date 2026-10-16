/**
 * ASESAMO — Unit & Edge Tests for Smart Role-Based Routing
 * File: asesamo-web/lib/__tests__/role_routing.test.ts
 */

import test from 'node:test';
import assert from 'node:assert/strict';
import { NextRequest } from 'next/server';
import { middleware } from '../../middleware';
import {
  TOKEN_COOKIE_NAME,
  getRoleFromJwt,
  getDefaultHomeForRole,
} from '../auth';

// Helper to craft mock JWT tokens with specific roles
function createMockJwt(role: string): string {
  const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
  const payload = Buffer.from(
    JSON.stringify({
      sub: '44444444-4444-4444-4444-444444444444',
      role,
      tenant_id: '11111111-1111-1111-1111-111111111111',
      exp: Math.floor(Date.now() / 1000) + 3600,
    })
  ).toString('base64url');
  const sig = 'mock_signature_bytes';
  return `${header}.${payload}.${sig}`;
}

test('Role Routing: getDefaultHomeForRole maps each role to its dedicated home page', () => {
  assert.equal(getDefaultHomeForRole('guard'), '/caseta');
  assert.equal(getDefaultHomeForRole('plant_admin'), '/reportes');
  assert.equal(getDefaultHomeForRole('tenant_admin'), '/reportes');
  assert.equal(getDefaultHomeForRole('contractor_admin'), '/contratistas');
  assert.equal(getDefaultHomeForRole('ehs_manager'), '/ehs/aprobaciones');
  assert.equal(getDefaultHomeForRole('worker'), '/credencial');
  // Fallbacks
  assert.equal(getDefaultHomeForRole(null), '/caseta');
  assert.equal(getDefaultHomeForRole(undefined), '/caseta');
  assert.equal(getDefaultHomeForRole('unknown_role'), '/caseta');
});

test('Role Routing: getRoleFromJwt parses role cleanly from JWT payload', () => {
  const guardJwt = createMockJwt('guard');
  assert.equal(getRoleFromJwt(guardJwt), 'guard');

  const contractorJwt = createMockJwt('contractor_admin');
  assert.equal(getRoleFromJwt(contractorJwt), 'contractor_admin');

  const ehsJwt = createMockJwt('ehs_manager');
  assert.equal(getRoleFromJwt(ehsJwt), 'ehs_manager');

  const plantAdminJwt = createMockJwt('plant_admin');
  assert.equal(getRoleFromJwt(plantAdminJwt), 'plant_admin');

  // Edge cases
  assert.equal(getRoleFromJwt(''), null);
  assert.equal(getRoleFromJwt('invalid.token'), null);
  assert.equal(getRoleFromJwt(null), null);
});

test('Role Routing: Middleware redirects root / to role-specific home page for authenticated users', () => {
  // 1. Guard -> /caseta
  const guardReq = new NextRequest('http://localhost:3000/', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('guard')}` },
  });
  const guardRes = middleware(guardReq);
  assert.equal(guardRes.status, 307);
  assert.equal(new URL(guardRes.headers.get('location')!).pathname, '/caseta');

  // 2. Contractor Admin -> /contratistas
  const contReq = new NextRequest('http://localhost:3000/', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('contractor_admin')}` },
  });
  const contRes = middleware(contReq);
  assert.equal(contRes.status, 307);
  assert.equal(new URL(contRes.headers.get('location')!).pathname, '/contratistas');

  // 3. EHS Manager -> /ehs/aprobaciones
  const ehsReq = new NextRequest('http://localhost:3000/', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('ehs_manager')}` },
  });
  const ehsRes = middleware(ehsReq);
  assert.equal(ehsRes.status, 307);
  assert.equal(new URL(ehsRes.headers.get('location')!).pathname, '/ehs/aprobaciones');

  // 4. Plant Admin -> /reportes
  const adminReq = new NextRequest('http://localhost:3000/', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('plant_admin')}` },
  });
  const adminRes = middleware(adminReq);
  assert.equal(adminRes.status, 307);
  assert.equal(new URL(adminRes.headers.get('location')!).pathname, '/reportes');

  // 5. Worker -> /credencial
  const workerReq = new NextRequest('http://localhost:3000/', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('worker')}` },
  });
  const workerRes = middleware(workerReq);
  assert.equal(workerRes.status, 307);
  assert.equal(new URL(workerRes.headers.get('location')!).pathname, '/credencial');
});

test('Role Routing: Middleware redirects /login to role-specific home page when already authenticated', () => {
  const contReq = new NextRequest('http://localhost:3000/login', {
    headers: { cookie: `${TOKEN_COOKIE_NAME}=${createMockJwt('contractor_admin')}` },
  });
  const contRes = middleware(contReq);
  assert.equal(contRes.status, 307);
  assert.equal(new URL(contRes.headers.get('location')!).pathname, '/contratistas');
});

test('Role Routing: Middleware protects all administrative and operational subpaths', () => {
  const protectedRoutes = [
    '/documentos',
    '/contratistas',
    '/ehs/aprobaciones',
    '/reportes',
    '/credencial',
    '/caseta',
  ];

  for (const path of protectedRoutes) {
    const unauthReq = new NextRequest(`http://localhost:3000${path}`);
    const unauthRes = middleware(unauthReq);
    assert.equal(unauthRes.status, 307);
    const loc = new URL(unauthRes.headers.get('location')!);
    assert.equal(loc.pathname, '/login');
    assert.equal(loc.searchParams.get('redirect'), path);
  }
});
