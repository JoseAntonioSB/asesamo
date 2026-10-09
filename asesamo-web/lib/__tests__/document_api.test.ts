import test from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient.uploadDocument serializes FormData and extracts document payload', async (t) => {
  setStoredToken('test-token-xyz');

  const client = new ApiClient('http://localhost:8080/api/v1');

  // Intercept global fetch
  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';
  let capturedOptions: RequestInit | undefined;

  global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
    capturedUrl = url.toString();
    capturedOptions = options;

    return new Response(
      JSON.stringify({
        data: {
          id: '11111111-2222-3333-4444-555555555555',
          document_type_code: 'DC3',
          file_name: 'dc3_test.pdf',
          status: 'pending_review',
          download_url: '/api/v1/documents/11111111-2222-3333-4444-555555555555/download',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
      }),
      {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  };

  const file = new File(['%PDF-1.4 sample content'], 'dc3_test.pdf', { type: 'application/pdf' });
  const result = await client.uploadDocument({
    file,
    documentTypeCode: 'DC3',
    issuedAt: '2026-01-15',
  });

  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/upload');
  assert.equal(capturedOptions?.method, 'POST');
  assert.ok(capturedOptions?.body instanceof FormData, 'body should be FormData');
  assert.equal(result.id, '11111111-2222-3333-4444-555555555555');
  assert.equal(result.status, 'pending_review');
  assert.equal(result.document_type_code, 'DC3');
});

test('ApiClient.getMyDocuments queries /documents/my-documents with proper query params', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';

  global.fetch = async (url: string | URL | Request) => {
    capturedUrl = url.toString();
    return new Response(
      JSON.stringify({
        data: [
          {
            id: 'doc-1',
            document_type_code: 'INE',
            file_name: 'ine.pdf',
            status: 'approved',
          },
        ],
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } }
    );
  };

  const docs = await client.getMyDocuments({ worker_id: 'worker-123' });
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/my-documents?worker_id=worker-123');
  assert.equal(docs.length, 1);
  assert.equal(docs[0].document_type_code, 'INE');
});

test('ApiClient.getDocumentTypes queries /documents/types', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';

  global.fetch = async (url: string | URL | Request) => {
    capturedUrl = url.toString();
    return new Response(
      JSON.stringify({
        data: [
          { code: 'DC3', name: 'Constancia DC-3', applies_to: 'worker' },
          { code: 'REPSE', name: 'Registro REPSE', applies_to: 'contractor' },
        ],
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } }
    );
  };

  const types = await client.getDocumentTypes();
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/types');
  assert.equal(types.length, 2);
  assert.equal(types[0].code, 'DC3');
});

test('ApiClient.getPendingDocuments queries /documents/pending with optional status filter', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';

  global.fetch = async (url: string | URL | Request) => {
    capturedUrl = url.toString();
    return new Response(
      JSON.stringify({
        data: [
          {
            id: 'doc-pending-1',
            document_type_code: 'DC3',
            file_name: 'dc3.pdf',
            status: 'pending_review',
            owner_name: 'Juan Perez',
            contractor_name: 'Acme Corp',
          },
        ],
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } }
    );
  };

  const docs = await client.getPendingDocuments('pending_review');
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/pending?status=pending_review');
  assert.equal(docs.length, 1);
  assert.equal(docs[0].id, 'doc-pending-1');
  assert.equal(docs[0].owner_name, 'Juan Perez');
});

test('ApiClient.reviewDocument sends POST /documents/:id/review with approve and reject actions', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';
  let capturedBody = '';

  global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
    capturedUrl = url.toString();
    capturedBody = (options?.body as string) || '';
    const parsed = JSON.parse(capturedBody);
    return new Response(
      JSON.stringify({
        data: {
          id: 'doc-999',
          status: parsed.action === 'approve' ? 'approved' : 'rejected',
          rejection_reason: parsed.rejection_reason || null,
        },
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } }
    );
  };

  // Test Approve
  const approved = await client.reviewDocument('doc-999', 'approve');
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/doc-999/review');
  assert.deepEqual(JSON.parse(capturedBody), { action: 'approve' });
  assert.equal(approved.status, 'approved');

  // Test Reject
  const rejected = await client.reviewDocument('doc-999', 'reject', 'Documento borroso o ilegible');
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/documents/doc-999/review');
  assert.deepEqual(JSON.parse(capturedBody), {
    action: 'reject',
    rejection_reason: 'Documento borroso o ilegible',
  });
  assert.equal(rejected.status, 'rejected');
  assert.equal(rejected.rejection_reason, 'Documento borroso o ilegible');
});

test('ApiClient.updateTheme sends PATCH /users/me/theme with theme choice', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';
  let capturedBody = '';
  let capturedMethod = '';

  global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
    capturedUrl = url.toString();
    capturedMethod = options?.method || '';
    capturedBody = (options?.body as string) || '';
    return new Response(
      JSON.stringify({
        data: {
          theme: JSON.parse(capturedBody).theme,
        },
      }),
      { status: 200, headers: { 'Content-Type': 'application/json' } }
    );
  };

  const res = await client.updateTheme('light');
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/users/me/theme');
  assert.equal(capturedMethod, 'PATCH');
  assert.deepEqual(JSON.parse(capturedBody), { theme: 'light' });
  assert.equal(res.theme, 'light');
});
