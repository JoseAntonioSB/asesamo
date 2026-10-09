import test from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient contractor methods work properly', async (t) => {
  setStoredToken('test-contractor-token');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  await t.test('inviteContractor sends POST /contractors/invite', async () => {
    let capturedUrl = '';
    let capturedMethod = '';
    let capturedBody = '';

    global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
      capturedUrl = url.toString();
      capturedMethod = options?.method || 'GET';
      capturedBody = options?.body?.toString() || '';

      return new Response(
        JSON.stringify({
          data: {
            id: 'con-123',
            name: 'Electromecánica Alfa',
            rfc: 'EMA010101XYZ',
            repse_number: 'REPSE-1234',
            contractor_type: 'company',
            admin_user_id: 'usr-999',
            is_active: true,
            created_at: new Date().toISOString(),
          },
        }),
        { status: 201, headers: { 'Content-Type': 'application/json' } }
      );
    };

    const res = await client.inviteContractor({
      plant_id: 'plant-1',
      name: 'Electromecánica Alfa',
      rfc: 'EMA010101XYZ',
      repse_number: 'REPSE-1234',
      contractor_type: 'company',
      admin_email: 'alfa@empresa.mx',
    });

    assert.equal(capturedUrl, 'http://localhost:8080/api/v1/contractors/invite');
    assert.equal(capturedMethod, 'POST');
    assert.ok(capturedBody.includes('Electromecánica Alfa'));
    assert.equal(res.name, 'Electromecánica Alfa');
    assert.equal(res.rfc, 'EMA010101XYZ');
  });

  await t.test('listPlantContractors sends GET /contractors?plant_id=...', async () => {
    let capturedUrl = '';

    global.fetch = async (url: string | URL | Request) => {
      capturedUrl = url.toString();
      return new Response(
        JSON.stringify({
          data: [
            {
              id: 'con-123',
              name: 'Electromecánica Alfa',
              plant_status: 'active',
              workers_count: 5,
              is_active: true,
              created_at: new Date().toISOString(),
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    };

    const list = await client.listPlantContractors('plant-1');
    assert.equal(capturedUrl, 'http://localhost:8080/api/v1/contractors?plant_id=plant-1');
    assert.equal(list.length, 1);
    assert.equal(list[0].name, 'Electromecánica Alfa');
    assert.equal(list[0].plant_status, 'active');
  });

  await t.test('listContractorWorkers and addContractorWorker', async () => {
    let capturedUrl = '';
    let capturedMethod = '';

    global.fetch = async (url: string | URL | Request, options?: RequestInit) => {
      capturedUrl = url.toString();
      capturedMethod = options?.method || 'GET';

      if (capturedMethod === 'POST') {
        return new Response(
          JSON.stringify({
            data: {
              worker_id: 'w-1',
              contractor_id: 'con-123',
              full_name: 'Miguel Hidalgo',
              email: 'miguel@alfa.mx',
              job_title: 'Electricista Industrial',
              curp: 'HIMM800508HDF',
              imss_number: '99887766554',
              is_active: true,
              user_is_active: true,
              documents_count: 0,
              created_at: new Date().toISOString(),
            },
          }),
          { status: 201, headers: { 'Content-Type': 'application/json' } }
        );
      }

      return new Response(
        JSON.stringify({
          data: [
            {
              worker_id: 'w-1',
              contractor_id: 'con-123',
              full_name: 'Miguel Hidalgo',
              email: 'miguel@alfa.mx',
              job_title: 'Electricista Industrial',
              is_active: true,
              user_is_active: true,
              documents_count: 0,
              created_at: new Date().toISOString(),
            },
          ],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    };

    const added = await client.addContractorWorker({
      contractor_id: 'con-123',
      full_name: 'Miguel Hidalgo',
      email: 'miguel@alfa.mx',
      job_title: 'Electricista Industrial',
      curp: 'HIMM800508HDF',
      imss_number: '99887766554',
    });

    assert.equal(capturedUrl, 'http://localhost:8080/api/v1/contractors/workers');
    assert.equal(capturedMethod, 'POST');
    assert.equal(added.full_name, 'Miguel Hidalgo');

    const workers = await client.listContractorWorkers('con-123');
    assert.equal(capturedUrl, 'http://localhost:8080/api/v1/contractors/workers?contractor_id=con-123');
    assert.equal(workers.length, 1);
  });
});
