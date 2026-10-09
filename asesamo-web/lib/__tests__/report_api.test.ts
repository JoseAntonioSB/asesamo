import test from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';
import { setStoredToken, clearAuthSession } from '../auth';

test('ApiClient report methods work properly', async (t) => {
  setStoredToken('test-report-token');
  const client = new ApiClient('http://localhost:8080/api/v1');

  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
    clearAuthSession();
  });

  await t.test('getReportMetrics queries /reports/metrics with query params', async () => {
    let capturedUrl = '';

    global.fetch = async (url: string | URL | Request) => {
      capturedUrl = url.toString();
      return new Response(
        JSON.stringify({
          data: {
            summary: {
              total_accesses: 100,
              granted_count: 95,
              warning_count: 3,
              denied_count: 2,
              compliance_rate: 98.0,
              avg_duration_minutes: 240.5,
            },
            timeline: [
              { date: '2026-10-09', total: 100, granted: 98, denied: 2 },
            ],
            rejection_reasons: [
              { reason: 'Constancia DC-3 Vencida', count: 2, percentage: 100 },
            ],
            contractor_ranking: [
              {
                contractor_id: 'c-1',
                contractor_name: 'Mantenimiento del Norte',
                total_attempts: 100,
                granted_count: 98,
                denied_count: 2,
                compliance_rate: 98.0,
              },
            ],
            generated_at: new Date().toISOString(),
          },
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } }
      );
    };

    const res = await client.getReportMetrics({
      plant_id: 'plant-123',
      from: '2026-10-01',
      to: '2026-10-09',
    });

    assert.ok(capturedUrl.includes('/reports/metrics?'));
    assert.ok(capturedUrl.includes('plant_id=plant-123'));
    assert.ok(capturedUrl.includes('from=2026-10-01'));
    assert.ok(capturedUrl.includes('to=2026-10-09'));
    assert.equal(res.summary.total_accesses, 100);
    assert.equal(res.summary.compliance_rate, 98.0);
    assert.equal(res.rejection_reasons.length, 1);
  });

  await t.test('getExportAccessLogsUrl and getExportComplianceUrl format correct paths', () => {
    const accessUrl = client.getExportAccessLogsUrl({
      plant_id: 'plant-123',
      from: '2026-10-01',
    });
    assert.equal(
      accessUrl,
      'http://localhost:8080/api/v1/reports/export/access-logs?plant_id=plant-123&from=2026-10-01'
    );

    const complUrl = client.getExportComplianceUrl('plant-123');
    assert.equal(
      complUrl,
      'http://localhost:8080/api/v1/reports/export/compliance?plant_id=plant-123'
    );
  });
});
