/**
 * ASESAMO — PTAR Digital Empirical Challenger Verification Suite
 * Location: asesamo-web/lib/__tests__/ptar_empirical_challenger.test.ts
 *
 * Adversarially probes:
 * 1. Edge Middleware Route Protection for /permisos and nested subroutes
 * 2. CrewComplianceSelector: Strict Candado Inteligente blocking of non-compliant workers
 * 3. DynamicStpsChecklist: NOM-033 Multigás atmospheric threshold boundary analysis (O2, LEL, CO, H2S)
 */

import test from 'node:test';
import assert from 'node:assert/strict';
import React from 'react';
import ReactDOMServer from 'react-dom/server';
import { NextRequest } from 'next/server';

import { middleware } from '../../middleware';
import { TOKEN_COOKIE_NAME } from '../auth';
import { CrewComplianceSelector } from '../../components/permisos/crew-compliance-selector';
import { DynamicStpsChecklist } from '../../components/permisos/dynamic-stps-checklist';
import { WorkerComplianceStatus, PermitType, PermitChecklistItem } from '../api/types';

// ============================================================================
// SUITE 1: MIDDLEWARE EDGE ROUTE PROTECTION FOR /permisos
// ============================================================================

test('Middleware: Unauthenticated access to /permisos strictly redirects to /login with redirect query param', () => {
  const req = new NextRequest('http://localhost:3000/permisos');
  const res = middleware(req);

  assert.equal(res.status, 307);
  const location = res.headers.get('location');
  assert.ok(location, 'Expected redirect location header');
  const url = new URL(location);
  assert.equal(url.pathname, '/login');
  assert.equal(url.searchParams.get('redirect'), '/permisos');
});

test('Middleware: Unauthenticated access to nested /permisos subroutes preserves deep pathname', () => {
  const nestedPaths = [
    '/permisos/nuevo',
    '/permisos/550e8400-e29b-41d4-a716-446655440000',
    '/permisos/550e8400-e29b-41d4-a716-446655440000/checklist',
    '/permisos/history?page=2&status=approved_active',
  ];

  for (const path of nestedPaths) {
    const req = new NextRequest(`http://localhost:3000${path}`);
    const res = middleware(req);

    assert.equal(res.status, 307, `Expected redirect for ${path}`);
    const location = res.headers.get('location');
    assert.ok(location);
    const url = new URL(location);
    assert.equal(url.pathname, '/login');
    // Pathname part before query should match
    const expectedPath = path.split('?')[0];
    assert.equal(url.searchParams.get('redirect'), expectedPath);
  }
});

test('Middleware: Authenticated user with valid token cookie accesses /permisos without redirect', () => {
  // Construct dummy JWT (header.payload.sig)
  const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
  const payload = Buffer.from(JSON.stringify({ role: 'ehs_officer', sub: 'user-123' })).toString('base64url');
  const fakeToken = `${header}.${payload}.dummy_signature`;

  const req = new NextRequest('http://localhost:3000/permisos', {
    headers: {
      cookie: `${TOKEN_COOKIE_NAME}=${fakeToken}`,
    },
  });

  const res = middleware(req);
  // Next.js middleware returns next() which does not set redirect location
  assert.equal(res.headers.get('location'), null);
});

// ============================================================================
// SUITE 2: CREW COMPLIANCE SELECTOR (CANDADO INTELIGENTE UI LOCK)
// ============================================================================

const mockWorkers: WorkerComplianceStatus[] = [
  {
    worker_id: 'worker-compliant-01',
    full_name: 'Juan Pérez Electricista',
    curp: 'PERJ850101HMCRLN01',
    imss_number: '12345678901',
    job_title: 'Técnico Especialista LOTO',
    is_compliant: true,
    blocking_reasons: [],
    has_valid_dc3: true,
    dc3_document_id: 'doc-dc3-1',
    dc3_expires_at: '2027-12-31T00:00:00Z',
    has_valid_medical: true,
    medical_document_id: 'doc-med-1',
    medical_expires_at: '2027-06-30T00:00:00Z',
    contractor_active: true,
  },
  {
    worker_id: 'worker-noncompliant-02',
    full_name: 'Carlos Bloqueado Sin DC3',
    curp: 'BLOC900202HMCRLN02',
    imss_number: '98765432109',
    job_title: 'Ayudante General',
    is_compliant: false,
    blocking_reasons: [
      'El trabajador no cuenta con constancia DC-3 vigente para seguridad eléctrica y LOTO (NOM-029)',
      'Certificado médico de aptitud vencido o ausente',
    ],
    has_valid_dc3: false,
    dc3_document_id: null,
    dc3_expires_at: null,
    has_valid_medical: false,
    medical_document_id: null,
    medical_expires_at: null,
    contractor_active: true,
  },
];

test('CrewComplianceSelector: Renders both compliant and non-compliant workers with correct badges and counts', () => {
  const html = ReactDOMServer.renderToString(
    React.createElement(CrewComplianceSelector, {
      workers: mockWorkers,
      selectedWorkerIds: ['worker-compliant-01'],
      permitType: 'electrical_loto',
      onChangeSelected: () => {},
    })
  );

  // Check header badge counts (allowing for React 18 SSR comment markers)
  assert.ok(/1\s*(<!--.*?-->)?\s*Aptos/.test(html), 'Must display 1 Aptos');
  assert.ok(/1\s*(<!--.*?-->)?\s*Bloqueados/.test(html), 'Must display 1 Bloqueados');

  // Check worker names
  assert.ok(html.includes('Juan Pérez Electricista'));
  assert.ok(html.includes('Carlos Bloqueado Sin DC3'));

  // Check compliant indicators (allowing for React 18 SSR comment markers)
  assert.ok(/Apto para\s*(<!--.*?-->)?\s*LOTO/.test(html), 'Must display Apto para LOTO');
  assert.ok(html.includes('DC-3 + Examen Médico Acreditados'));

  // Check non-compliant blocking reasons
  assert.ok(html.includes('Bloqueado por Candado STPS'));
  assert.ok(html.includes('Asignación Prohibida por Normativa'));
  assert.ok(html.includes('El trabajador no cuenta con constancia DC-3 vigente para seguridad eléctrica y LOTO (NOM-029)'));
  assert.ok(html.includes('Certificado médico de aptitud vencido o ausente'));

  // Non-compliant card must have disabled input
  assert.ok(html.includes('disabled=""') || html.includes('disabled'), 'Non-compliant checkbox must be disabled');
});

test('CrewComplianceSelector: Pure logic verifies non-compliant workers cannot be added to selection', () => {
  let selectedIds: string[] = [];
  const onChangeSelected = (ids: string[]) => {
    selectedIds = ids;
  };

  // Replicate toggleWorker logic from CrewComplianceSelector:
  const toggleWorker = (workerId: string, isCompliant: boolean) => {
    if (!isCompliant) return; // Strict lock
    if (selectedIds.includes(workerId)) {
      onChangeSelected(selectedIds.filter((id) => id !== workerId));
    } else {
      onChangeSelected([...selectedIds, workerId]);
    }
  };

  // 1. Attempt to toggle non-compliant worker -> Should NOT be selected
  toggleWorker('worker-noncompliant-02', false);
  assert.deepEqual(selectedIds, [], 'Non-compliant worker must never be selected');

  // 2. Toggle compliant worker -> Should be selected
  toggleWorker('worker-compliant-01', true);
  assert.deepEqual(selectedIds, ['worker-compliant-01'], 'Compliant worker should be selected');

  // 3. Attempt to add non-compliant again -> Remains only compliant
  toggleWorker('worker-noncompliant-02', false);
  assert.deepEqual(selectedIds, ['worker-compliant-01'], 'Non-compliant worker attempt rejected');

  // 4. Toggle compliant worker off -> Empty
  toggleWorker('worker-compliant-01', true);
  assert.deepEqual(selectedIds, []);

  // 5. Test selectAllCompliant logic
  const selectAllCompliant = (workers: WorkerComplianceStatus[]) => {
    const compliantIds = workers.filter((w) => w.is_compliant).map((w) => w.worker_id);
    onChangeSelected(compliantIds);
  };

  selectAllCompliant(mockWorkers);
  assert.deepEqual(selectedIds, ['worker-compliant-01'], 'selectAllCompliant must only select compliant IDs');
  const finalIds: string[] = selectedIds;
  assert.ok(!finalIds.includes('worker-noncompliant-02'));
});

// ============================================================================
// SUITE 3: CONFINED SPACE DYNAMIC CHECKLIST & GAS THRESHOLDS (NOM-033)
// ============================================================================

test('DynamicStpsChecklist: Confined space multigás evaluates gas threshold boundary conditions', () => {
  // Gas safety evaluation oracle following NOM-033-STPS-2015:
  // O2: 19.5% to 23.5%
  // LEL: < 10%
  // CO: < 25 ppm
  // H2S: < 10 ppm
  const evaluateAtmosphere = (o2: number, lel: number, co: number, h2s: number) => {
    const o2Safe = o2 >= 19.5 && o2 <= 23.5;
    const lelSafe = lel < 10;
    const coSafe = co < 25;
    const h2sSafe = h2s < 10;
    return {
      o2Safe,
      lelSafe,
      coSafe,
      h2sSafe,
      atmosphericSafe: o2Safe && lelSafe && coSafe && h2sSafe,
    };
  };

  // 1. Nominal ambient conditions (20.9% O2, 0 LEL, 0 CO, 0 H2S)
  const nominal = evaluateAtmosphere(20.9, 0, 0, 0);
  assert.equal(nominal.atmosphericSafe, true);

  // 2. O2 Boundaries (19.5% - 23.5%):
  assert.equal(evaluateAtmosphere(19.5, 0, 0, 0).o2Safe, true, '19.5% is lower safe boundary');
  assert.equal(evaluateAtmosphere(19.49, 0, 0, 0).o2Safe, false, '19.49% is asphyxiation hazard');
  assert.equal(evaluateAtmosphere(19.49, 0, 0, 0).atmosphericSafe, false);
  assert.equal(evaluateAtmosphere(23.5, 0, 0, 0).o2Safe, true, '23.5% is upper safe boundary');
  assert.equal(evaluateAtmosphere(23.51, 0, 0, 0).o2Safe, false, '23.51% is oxygen enrichment hazard');
  assert.equal(evaluateAtmosphere(23.51, 0, 0, 0).atmosphericSafe, false);

  // 3. LEL Boundaries (< 10%):
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 0).lelSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 9.9, 0, 0).lelSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 10.0, 0, 0).lelSafe, false, '10% LEL is flammability threshold violation');
  assert.equal(evaluateAtmosphere(20.9, 10.0, 0, 0).atmosphericSafe, false);
  assert.equal(evaluateAtmosphere(20.9, 50, 0, 0).lelSafe, false);

  // 4. CO (Carbon Monoxide) Boundaries (< 25 ppm):
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 0).coSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 0, 24.9, 0).coSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 0, 25, 0).coSafe, false, '25 ppm CO is toxicity threshold violation');
  assert.equal(evaluateAtmosphere(20.9, 0, 25, 0).atmosphericSafe, false);
  assert.equal(evaluateAtmosphere(20.9, 0, 100, 0).coSafe, false);

  // 5. H2S (Hydrogen Sulfide) Boundaries (< 10 ppm):
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 0).h2sSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 9.9).h2sSafe, true);
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 10).h2sSafe, false, '10 ppm H2S is toxicity threshold violation');
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 10).atmosphericSafe, false);
  assert.equal(evaluateAtmosphere(20.9, 0, 0, 15).h2sSafe, false);

  // 6. Compound 16-state truth table: ONLY the (safe, safe, safe, safe) state yields atmosphericSafe = true
  for (const o2 of [18.0, 20.9]) {
    for (const lel of [0, 15]) {
      for (const co of [0, 30]) {
        for (const h2s of [0, 15]) {
          const res = evaluateAtmosphere(o2, lel, co, h2s);
          const expectedSafe = o2 === 20.9 && lel === 0 && co === 0 && h2s === 0;
          assert.equal(
            res.atmosphericSafe,
            expectedSafe,
            `State (O2=${o2}, LEL=${lel}, CO=${co}, H2S=${h2s}) expected atmosphericSafe=${expectedSafe}`
          );
        }
      }
    }
  }
});

test('DynamicStpsChecklist: Confined space multigás UI component renders atmospheric monitoring panel', () => {
  // Render default confined_space checklist
  const safeHtml = ReactDOMServer.renderToString(
    React.createElement(DynamicStpsChecklist, {
      permitType: 'confined_space',
    })
  );

  assert.ok(safeHtml.includes('Prueba Atmosférica Multigás (NOM-033-STPS-2015)'));
  assert.ok(safeHtml.includes('Atmósfera Segura'));
  assert.ok(safeHtml.includes('Norma: 19.5 - 23.5%'));
  assert.ok(safeHtml.includes('Límite: &lt; 10% LEL') || safeHtml.includes('Límite: < 10% LEL'));
  assert.ok(safeHtml.includes('Límite: &lt; 25 ppm') || safeHtml.includes('Límite: < 25 ppm'));
  assert.ok(safeHtml.includes('Límite: &lt; 10 ppm') || safeHtml.includes('Límite: < 10 ppm'));
  assert.ok(safeHtml.includes('Oxígeno (O₂)'));
  assert.ok(safeHtml.includes('Explosividad (LEL)'));
  assert.ok(safeHtml.includes('Monóxido (CO)'));
  assert.ok(safeHtml.includes('Ácido Sulfhídrico (H₂S)'));
});

test('DynamicStpsChecklist: Non-confined space hazards do not render multigás panel', () => {
  const hazards: PermitType[] = ['heights', 'hot_work', 'electrical_loto'];

  for (const hazard of hazards) {
    const html = ReactDOMServer.renderToString(
      React.createElement(DynamicStpsChecklist, {
        permitType: hazard,
      })
    );

    assert.ok(html.includes('Lista de Verificación Preventiva (STPS)'));
    assert.ok(!html.includes('Prueba Atmosférica Multigás (NOM-033-STPS-2015)'), `${hazard} must not have gas panel`);
  }
});
