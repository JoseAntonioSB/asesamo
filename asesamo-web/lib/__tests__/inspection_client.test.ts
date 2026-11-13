import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';

test('ApiClient.completeInspection sends answers and receives scored evaluation with CAPA indicators', async (t) => {
  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
  });

  const inspectionId = 'b83d8e5e-63f2-4e89-913a-a5f7823f9901';
  const mockPayload = {
    data: {
      inspection_id: inspectionId,
      score_percentage: 66.67,
      status: 'requires_capa',
      requires_capa: true,
      generated_capas: [
        {
          id: 'c94e9f5e-74f3-4f90-824b-b6f8934f0012',
          description: 'Falta barandal de seguridad en fosa',
          severity: 'critical',
          status: 'open',
          due_date: '2026-11-19T00:00:00Z',
        },
      ],
    },
  };

  let capturedUrl = '';
  let capturedMethod = '';
  let capturedBody = '';

  global.fetch = (async (url: string | URL | Request, init?: RequestInit) => {
    capturedUrl = url.toString();
    capturedMethod = init?.method || 'GET';
    capturedBody = (init?.body as string) || '';

    return {
      ok: true,
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
      text: async () => JSON.stringify(mockPayload),
      json: async () => mockPayload,
    } as unknown as Response;
  }) as typeof fetch;

  const client = new ApiClient('http://localhost:8080/api/v1');

  const result = await client.completeInspection(inspectionId, {
    answers: [
      {
        template_item_id: 'a12d8e5e-63f2-4e89-913a-a5f7823f9900',
        status: 'non_compliant',
        finding_notes: 'Falta barandal de seguridad en fosa',
      },
    ],
  });

  assert.equal(capturedUrl, `http://localhost:8080/api/v1/inspections/${inspectionId}/complete`);
  assert.equal(capturedMethod, 'POST');
  assert.match(capturedBody, /Falta barandal de seguridad en fosa/);
  assert.equal(result.score_percentage, 66.67);
  assert.equal(result.status, 'requires_capa');
  assert.equal(result.requires_capa, true);
  assert.equal(result.generated_capas.length, 1);
});
