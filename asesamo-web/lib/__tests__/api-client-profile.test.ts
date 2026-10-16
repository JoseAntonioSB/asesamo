import test from 'node:test';
import assert from 'node:assert/strict';
import { apiClient } from '../api/client';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient: Profile and Avatar Operations', async (t) => {
  const originalFetch = global.fetch;
  const originalWindow = (global as any).window;
  const originalDocument = (global as any).document;

  (global as any).window = { location: { protocol: 'http:' } };
  (global as any).document = { cookie: '' };

  t.after(() => {
    global.fetch = originalFetch;
    (global as any).window = originalWindow;
    (global as any).document = originalDocument;
    clearAuthSession();
  });

  t.afterEach(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  await t.test('getProfile sends GET /users/me with bearer token and unwraps data', async () => {
    setStoredToken('test-profile-jwt-token');

    let capturedUrl = '';
    let capturedMethod = '';
    let capturedAuth = '';

    global.fetch = (async (url: string | URL | Request, init?: RequestInit) => {
      capturedUrl = url.toString();
      capturedMethod = init?.method || 'GET';
      capturedAuth = (init?.headers as Record<string, string>)?.['Authorization'] || '';

      return new Response(
        JSON.stringify({
          data: {
            id: '11111111-1111-1111-1111-111111111111',
            email: 'user@asesamo.com',
            full_name: 'Juan Perez',
            phone: '5512345678',
            role: 'worker',
            theme_preference: 'dark',
            is_active: true,
            created_at: '2026-10-09T00:00:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    }) as unknown as typeof fetch;

    const profile = await apiClient.getProfile();
    assert.match(capturedUrl, /\/users\/me$/);
    assert.equal(capturedMethod, 'GET');
    assert.equal(capturedAuth, 'Bearer test-profile-jwt-token');
    assert.equal(profile.full_name, 'Juan Perez');
    assert.equal(profile.role, 'worker');
  });

  await t.test('updateProfile sends PATCH /users/me with JSON body', async () => {
    setStoredToken('test-profile-jwt-token');

    let capturedUrl = '';
    let capturedMethod = '';
    let capturedBody = '';

    global.fetch = (async (url: string | URL | Request, init?: RequestInit) => {
      capturedUrl = url.toString();
      capturedMethod = init?.method || 'GET';
      capturedBody = init?.body as string;

      return new Response(
        JSON.stringify({
          data: {
            id: '11111111-1111-1111-1111-111111111111',
            email: 'user@asesamo.com',
            full_name: 'Juan Carlos Perez',
            phone: '5598765432',
            role: 'worker',
            theme_preference: 'dark',
            is_active: true,
            created_at: '2026-10-09T00:00:00Z',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    }) as unknown as typeof fetch;

    const updated = await apiClient.updateProfile({
      full_name: 'Juan Carlos Perez',
      phone: '5598765432',
    });

    assert.match(capturedUrl, /\/users\/me$/);
    assert.equal(capturedMethod, 'PATCH');
    assert.match(capturedBody, /Juan Carlos Perez/);
    assert.equal(updated.full_name, 'Juan Carlos Perez');
    assert.equal(updated.phone, '5598765432');
  });

  await t.test('uploadMyAvatar sends POST /users/me/avatar as FormData', async () => {
    setStoredToken('test-profile-jwt-token');

    let capturedUrl = '';
    let capturedMethod = '';
    let capturedContentType: string | undefined;

    global.fetch = (async (url: string | URL | Request, init?: RequestInit) => {
      capturedUrl = url.toString();
      capturedMethod = init?.method || 'GET';
      capturedContentType = (init?.headers as Record<string, string>)?.['Content-Type'];

      return new Response(
        JSON.stringify({
          data: {
            avatar_url: '/api/v1/users/me/avatar',
            avatar_key: 'avatars/111/avatar.jpg',
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    }) as unknown as typeof fetch;

    const dummyBlob = new Blob(['image data'], { type: 'image/jpeg' });
    const dummyFile = new File([dummyBlob], 'avatar.jpg', { type: 'image/jpeg' });

    const res = await apiClient.uploadMyAvatar(dummyFile);
    assert.match(capturedUrl, /\/users\/me\/avatar$/);
    assert.equal(capturedMethod, 'POST');
    assert.equal(capturedContentType, undefined); // FormData lets browser set boundary
    assert.equal(res.avatar_key, 'avatars/111/avatar.jpg');
  });

  await t.test('getUserAvatarUrl formats correct path', () => {
    const url = apiClient.getUserAvatarUrl('user-123-uuid');
    assert.match(url, /\/users\/user-123-uuid\/avatar$/);
  });
});
