import test from 'node:test';
import assert from 'node:assert/strict';
import { cn, formatDate } from '../utils';
import { normalizeApiUrl, ApiError, ApiClient } from '../api/client';
import { TOKEN_COOKIE_NAME, USER_STORAGE_KEY } from '../auth';

test('cn helper resolves conditional and conflicting tailwind classes', () => {
  const result = cn('px-2 py-1', true && 'text-red-500', false && 'hidden', 'px-4');
  assert.equal(result, 'py-1 text-red-500 px-4');
});

test('formatDate formats valid dates in Mexican Spanish locale', () => {
  const formatted = formatDate('2026-10-01T22:00:00Z');
  assert.match(formatted, /2026/);
  assert.equal(formatDate(null), 'N/A');
  assert.equal(formatDate(undefined), 'N/A');
  assert.equal(formatDate('invalid-date'), 'invalid-date');
});

test('normalizeApiUrl normalizes various base URL inputs properly', () => {
  assert.equal(normalizeApiUrl('http://localhost:8080'), 'http://localhost:8080/api/v1');
  assert.equal(normalizeApiUrl('http://localhost:8080/'), 'http://localhost:8080/api/v1');
  assert.equal(normalizeApiUrl('http://localhost:8080/api/v1'), 'http://localhost:8080/api/v1');
  assert.equal(normalizeApiUrl('http://localhost:8080/api/v1/'), 'http://localhost:8080/api/v1');
  assert.equal(normalizeApiUrl(''), 'http://localhost:8080/api/v1');
  assert.equal(normalizeApiUrl(undefined), 'http://localhost:8080/api/v1');
});

test('ApiError correctly maps payload and network errors', () => {
  const payloadError = ApiError.fromPayload(
    {
      error: {
        code: 'INVALID_CREDENTIALS',
        message: 'Credenciales inválidas.',
      },
    },
    401
  );
  assert.equal(payloadError.code, 'INVALID_CREDENTIALS');
  assert.equal(payloadError.message, 'Credenciales inválidas.');
  assert.equal(payloadError.status, 401);

  const netError = ApiError.network(0, 'Fallo de red');
  assert.equal(netError.code, 'NETWORK_ERROR');
  assert.equal(netError.status, 0);
  assert.equal(netError.message, 'Fallo de red');
});

test('auth persistence keys match contract', () => {
  assert.equal(TOKEN_COOKIE_NAME, 'asesamo_token');
  assert.equal(USER_STORAGE_KEY, 'asesamo_user');
});

test('ApiClient unwrap envelope successfully on 200 response', async () => {
  const client = new ApiClient('http://localhost:8080');

  // Mock global.fetch
  const originalFetch = global.fetch;
  try {
    global.fetch = async () => {
      return {
        ok: true,
        status: 200,
        text: async () =>
          JSON.stringify({
            data: { access_log_id: 'test-123', result: 'granted' },
            meta: { page: 1 },
          }),
      } as unknown as Response;
    };

    const res = await client.request('/test');
    assert.deepEqual(res.data, { access_log_id: 'test-123', result: 'granted' });
    assert.deepEqual(res.meta, { page: 1 });
  } finally {
    global.fetch = originalFetch;
  }
});

test('ApiClient throws structured ApiError on 400 error envelope', async () => {
  const client = new ApiClient('http://localhost:8080');

  const originalFetch = global.fetch;
  try {
    global.fetch = async () => {
      return {
        ok: false,
        status: 400,
        text: async () =>
          JSON.stringify({
            error: {
              code: 'DOCUMENT_EXPIRED',
              message: 'El documento DC-3 ha vencido.',
            },
          }),
      } as unknown as Response;
    };

    await assert.rejects(
      async () => {
        await client.request('/access/check-in');
      },
      (err: any) => {
        assert(err instanceof ApiError);
        assert.equal(err.code, 'DOCUMENT_EXPIRED');
        assert.equal(err.message, 'El documento DC-3 ha vencido.');
        assert.equal(err.status, 400);
        return true;
      }
    );
  } finally {
    global.fetch = originalFetch;
  }
});
