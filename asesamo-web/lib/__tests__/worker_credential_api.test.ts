import test from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient.getWorkerCredential queries /workers/:id/credential and unwraps payload', async (t) => {
  setStoredToken('test-token-xyz');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  let capturedUrl = '';
  let capturedHeaders: any;

  global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
    capturedUrl = url.toString();
    capturedHeaders = options?.headers;

    return new Response(
      JSON.stringify({
        data: {
          worker: {
            user_id: '66666666-6666-6666-6666-666666666661',
            full_name: 'Carlos Mendoza',
            curp: 'MENC850101HDFR01',
            imss_number: '12345678901',
            job_title: 'Soldador Industrial',
            is_active: true,
          },
          contractor: {
            id: '55555555-5555-5555-5555-555555555555',
            name: 'Mantenimiento Industrial del Norte SA de CV',
            rfc: 'MIN120304AB1',
            repse_number: 'REPSE-NL-2023-8841',
            status: 'active',
          },
          plant: {
            id: '22222222-2222-2222-2222-222222222222',
            name: 'Planta Querétaro Norte',
          },
          qr_payload: '66666666-6666-6666-6666-666666666661',
          compliance_status: 'granted',
          documents: [
            {
              document_type_code: 'DC3',
              document_name: 'Constancia DC-3',
              applies_to: 'worker',
              status: 'approved',
              is_required: true,
              is_valid: true,
              expires_at: '2027-01-01',
            },
          ],
          generated_at: new Date().toISOString(),
        },
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  };

  const credential = await client.getWorkerCredential('66666666-6666-6666-6666-666666666661');

  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/workers/66666666-6666-6666-6666-666666666661/credential');
  assert.equal(credential.worker.full_name, 'Carlos Mendoza');
  assert.equal(credential.qr_payload, '66666666-6666-6666-6666-666666666661');
  assert.equal(credential.compliance_status, 'granted');
  assert.equal(credential.documents.length, 1);
});

test('ApiClient.getMyCredential queries /workers/me/credential', async (t) => {
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
        data: {
          worker: {
            user_id: '66666666-6666-6666-6666-666666666661',
            full_name: 'Carlos Mendoza',
            is_active: true,
          },
          contractor: {
            id: '55555555-5555-5555-5555-555555555555',
            name: 'Mantenimiento Industrial del Norte SA de CV',
            status: 'active',
          },
          plant: {
            id: '22222222-2222-2222-2222-222222222222',
            name: 'Planta Querétaro Norte',
          },
          qr_payload: '66666666-6666-6666-6666-666666666661',
          compliance_status: 'granted',
          documents: [],
          generated_at: new Date().toISOString(),
        },
      }),
      {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }
    );
  };

  const credential = await client.getMyCredential();
  assert.equal(capturedUrl, 'http://localhost:8080/api/v1/workers/me/credential');
  assert.equal(credential.worker.full_name, 'Carlos Mendoza');
});
