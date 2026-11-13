import { test } from 'node:test';
import assert from 'node:assert/strict';
import { ApiClient } from '../api/client';
import type { IssueDC3Payload, RecordMedicalExamPayload } from '../api/types';

test('ApiClient.listTrainingCourses calls GET /training/courses and unwraps data', async (t) => {
  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
  });

  const mockPayload = {
    data: [
      { id: '1', code: 'CUR-ALT-001', title: 'Trabajos en Altura', duration_hours: 8 },
    ],
  };

  let capturedUrl = '';
  global.fetch = (async (url: string | URL | Request) => {
    capturedUrl = url.toString();
    return {
      ok: true,
      status: 200,
      headers: new Headers({ 'content-type': 'application/json' }),
      text: async () => JSON.stringify(mockPayload),
      json: async () => mockPayload,
    } as unknown as Response;
  }) as typeof fetch;

  const client = new ApiClient('http://localhost:8080/api/v1');
  const courses = await client.listTrainingCourses();
  assert.ok(capturedUrl.includes('/training/courses'));
  assert.strictEqual(courses.length, 1);
  assert.strictEqual(courses[0].code, 'CUR-ALT-001');
});

test('ApiClient.issueDC3 sends POST /training/dc3 with proper payload', async (t) => {
  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
  });

  const mockPayload = {
    data: {
      id: 'cert-1',
      folio_number: 'DC3-2026-0001',
      status: 'issued',
      score_percentage: 95,
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
      status: 201,
      headers: new Headers({ 'content-type': 'application/json' }),
      text: async () => JSON.stringify(mockPayload),
      json: async () => mockPayload,
    } as unknown as Response;
  }) as typeof fetch;

  const payload: IssueDC3Payload = {
    course_id: 'c1',
    worker_id: 'w1',
    worker_curp: 'MENC850101HDFR01',
    worker_full_name: 'Carlos Mendoza',
    company_name: 'Grupo Industrial Norteño SA de CV',
    company_rfc: 'GIN950418KP3',
    legal_rep_name: 'Ing. Laura Sanchez',
    worker_rep_name: 'Carlos Mendoza',
    instructor_name: 'Instructor STPS',
    score: 95,
  };

  const client = new ApiClient('http://localhost:8080/api/v1');
  const cert = await client.issueDC3(payload);

  assert.ok(capturedUrl.includes('/training/dc3'));
  assert.strictEqual(capturedMethod, 'POST');
  const parsed = JSON.parse(capturedBody);
  assert.strictEqual(parsed.worker_curp, 'MENC850101HDFR01');
  assert.strictEqual(cert.folio_number, 'DC3-2026-0001');
});

test('ApiClient.recordMedicalExam sends POST /health/exams and unwraps exam result', async (t) => {
  const originalFetch = global.fetch;
  t.after(() => {
    global.fetch = originalFetch;
  });

  const mockPayload = {
    data: {
      id: 'exam-1',
      exam_number: 'MED-2026-0001',
      fitness_status: 'fit',
      expiration_date: '2027-11-13',
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
      status: 201,
      headers: new Headers({ 'content-type': 'application/json' }),
      text: async () => JSON.stringify(mockPayload),
      json: async () => mockPayload,
    } as unknown as Response;
  }) as typeof fetch;

  const payload: RecordMedicalExamPayload = {
    worker_id: 'w1',
    exam_type: 'periodic_annual',
    fitness_status: 'fit',
    exam_date: '2026-11-13',
    doctor_name: 'Dr. Alejandro Gomez',
    doctor_license: 'CED-8849201',
    visual_acuity_result: '20/20',
    blood_pressure: '120/80',
  };

  const client = new ApiClient('http://localhost:8080/api/v1');
  const exam = await client.recordMedicalExam(payload);

  assert.ok(capturedUrl.includes('/health/exams'));
  assert.strictEqual(capturedMethod, 'POST');
  const parsed = JSON.parse(capturedBody);
  assert.strictEqual(parsed.doctor_name, 'Dr. Alejandro Gomez');
  assert.strictEqual(exam.exam_number, 'MED-2026-0001');
  assert.strictEqual(exam.fitness_status, 'fit');
});
