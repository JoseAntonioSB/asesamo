import { Page, expect } from '@playwright/test';

export const TEST_USERS = {
  guard: {
    email: 'guard@asesamo.com',
    password: 'Password123!',
    role: 'guard',
    expectedHome: '/caseta',
  },
  plantAdmin: {
    email: 'admin@asesamo.com',
    password: 'Password123!',
    role: 'plant_admin',
    expectedHome: '/dashboard',
  },
  ehsManager: {
    email: 'ehs@asesamo.com',
    password: 'Password123!',
    role: 'ehs_manager',
    expectedHome: '/ehs/aprobaciones',
  },
  contractorAdmin: {
    email: 'contratista@mantenimientonorte.com',
    password: 'Password123!',
    role: 'contractor_admin',
    expectedHome: '/contratistas',
  },
};

/**
 * Performs clean login on the /login page
 */
export async function loginAs(page: Page, user: typeof TEST_USERS['guard']) {
  await page.goto('/login');
  await page.fill('input[type="email"]', user.email);
  await page.fill('input[type="password"]', user.password);
  await page.click('button[type="submit"]');
}

/**
 * Injects a mock session cookie to test fast protected route access
 */
export async function mockAuthenticatedSession(page: Page, role = 'ehs_manager') {
  const userPayload = {
    id: '44444444-4444-4444-4444-444444444443',
    email: `${role}@asesamo.com`,
    first_name: 'Carlos',
    last_name: 'EHS Manager',
    role,
    tenant_id: '11111111-1111-1111-1111-111111111111',
    plant_id: '22222222-2222-2222-2222-222222222222',
  };

  // Simple payload for role identification
  const mockJwt = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.${btoa(
    JSON.stringify({
      user_id: userPayload.id,
      tenant_id: userPayload.tenant_id,
      role,
      exp: Math.floor(Date.now() / 1000) + 3600 * 24,
    })
  )}.signature`;

  await page.context().addCookies([
    {
      name: 'asesamo_token',
      value: mockJwt,
      domain: 'localhost',
      path: '/',
    },
  ]);

  await page.addInitScript(
    ({ user, token }) => {
      window.localStorage.setItem('asesamo_token', token);
      window.localStorage.setItem('asesamo_user', JSON.stringify(user));
    },
    { user: userPayload, token: mockJwt }
  );
}
