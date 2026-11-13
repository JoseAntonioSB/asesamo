-- ============================================================
-- ASESAMO — Phase 4 Training (DC-3) & Occupational Health (NOM-030) Seeds
-- File: db/seeds/000003_phase4_training_and_health_seeds.sql
-- Modules:
--   1. Training Courses Catalog & STPS DC-3 Certificates
--   2. Worker Medical Profiles & Periodic Health Exams (NOM-030)
-- ============================================================

BEGIN;

-- Common IDs:
-- Tenant:     11111111-1111-1111-1111-111111111111
-- Plant:      22222222-2222-2222-2222-222222222222
-- EHS:        44444444-4444-4444-4444-444444444443 (Ing. Laura Sanchez)
-- Contractor: 55555555-5555-5555-5555-555555555555 (Mantenimiento Industrial del Norte)
-- Worker 1:   66666666-6666-6666-6666-666666666661 (Carlos Mendoza - CURP: MENC850101HDFR01)
-- Worker 2:   66666666-6666-6666-6666-666666666662 (Juan Perez - CURP: PEZJ900202HDFR02)
-- Worker 3:   66666666-6666-6666-6666-666666666663 (Roberto Garcia - CURP: GARR880303HDFR03)

-- ------------------------------------------------------------
-- 1. TRAINING COURSES CATALOG
-- ------------------------------------------------------------
INSERT INTO training_courses (
  id, tenant_id, code, title, description, thematic_area, stps_thematic_code,
  duration_hours, validity_years, normative_reference, is_active
) VALUES
  (
    'bbbbbbbb-1111-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'CUR-ALT-001',
    'Seguridad Integral para Trabajos en Altura y Andamios',
    'Capacitación teórico-práctica para prevención de caídas, inspección de arnés, líneas de vida y armado seguro de andamios tubulares.',
    'specialized',
    '6000',
    8, 1,
    'NOM-009-STPS-2011',
    TRUE
  ),
  (
    'bbbbbbbb-1111-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    'CUR-CON-001',
    'Entrada Segura y Rescate en Espacios Confinados',
    'Monitoreo multigás de atmósferas peligrosas (O2, LEL, CO, H2S), ventilación forzada y protocolo vigía de seguridad.',
    'specialized',
    '6000',
    16, 1,
    'NOM-033-STPS-2015',
    TRUE
  ),
  (
    'bbbbbbbb-1111-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    'CUR-LOT-001',
    'Bloqueo, Candadeo y Etiquetado de Energías Peligrosas (LOTO)',
    'Desconexión y disipación de energía residual (eléctrica, neumática, hidráulica, mecánica) con colocación de candados y tarjetas personales.',
    'specialized',
    '6000',
    8, 1,
    'NOM-029-STPS-2011',
    TRUE
  ),
  (
    'bbbbbbbb-1111-1111-1111-000000000004',
    '11111111-1111-1111-1111-111111111111',
    'CUR-EXT-001',
    'Combate de Incendios y Operación de Extintores Portátiles',
    'Identificación de química del fuego, uso de extintores PQS y CO2, y evacuación ordenada de personal.',
    'safety',
    '6000',
    8, 1,
    'NOM-002-STPS-2010',
    TRUE
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 2. DC-3 CERTIFICATES (STPS COMPETENCIES)
-- ------------------------------------------------------------
INSERT INTO training_certificates (
  id, tenant_id, course_id, worker_id, contractor_id, folio_number, status,
  worker_occupation_code, worker_curp, worker_full_name,
  company_name, company_rfc, legal_rep_name, worker_rep_name, instructor_name,
  issue_date, expiration_date, score_percentage, digital_signature_token
) VALUES
  (
    'bbbbbbbb-2222-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-1111-1111-1111-000000000001', -- Alturas
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    '55555555-5555-5555-5555-555555555555',
    'DC3-2026-00001',
    'issued',
    '8100',
    'MENC850101HDFR01',
    'Carlos Mendoza',
    'Mantenimiento Industrial del Norte SA de CV',
    'MIN120304AB1',
    'Ing. Fernando Garza',
    'Carlos Mendoza',
    'Ing. Roberto Capacitador STPS',
    CURRENT_DATE - INTERVAL '20 days',
    CURRENT_DATE + INTERVAL '345 days',
    96.50,
    'STPS-SIG-VERIFIED-981240192'
  ),
  (
    'bbbbbbbb-2222-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-1111-1111-1111-000000000003', -- LOTO
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    '55555555-5555-5555-5555-555555555555',
    'DC3-2026-00002',
    'issued',
    '8100',
    'MENC850101HDFR01',
    'Carlos Mendoza',
    'Mantenimiento Industrial del Norte SA de CV',
    'MIN120304AB1',
    'Ing. Fernando Garza',
    'Carlos Mendoza',
    'Ing. Roberto Capacitador STPS',
    CURRENT_DATE - INTERVAL '15 days',
    CURRENT_DATE + INTERVAL '350 days',
    100.00,
    'STPS-SIG-VERIFIED-981240193'
  ),
  (
    'bbbbbbbb-2222-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    'bbbbbbbb-1111-1111-1111-000000000002', -- Espacios Confinados
    '66666666-6666-6666-6666-666666666662', -- Juan Perez
    '55555555-5555-5555-5555-555555555555',
    'DC3-2026-00003',
    'issued',
    '8100',
    'PEZJ900202HDFR02',
    'Juan Perez',
    'Mantenimiento Industrial del Norte SA de CV',
    'MIN120304AB1',
    'Ing. Fernando Garza',
    'Juan Perez',
    'Ing. Roberto Capacitador STPS',
    CURRENT_DATE - INTERVAL '10 days',
    CURRENT_DATE + INTERVAL '355 days',
    92.00,
    'STPS-SIG-VERIFIED-981240194'
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 3. WORKER MEDICAL PROFILES (NOM-030-STPS)
-- ------------------------------------------------------------
INSERT INTO worker_medical_profiles (
  id, tenant_id, worker_id, blood_type, allergies, chronic_conditions,
  emergency_contact_name, emergency_contact_phone, emergency_contact_relationship
) VALUES
  (
    'cccccccc-1111-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    'O+',
    'Penicilina y derivados',
    'Ninguna conocida',
    'María Mendoza López',
    '4421112233',
    'Esposa'
  ),
  (
    'cccccccc-1111-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666662', -- Juan Perez
    'A+',
    'Ninguna',
    'Hipertensión controlada',
    'Ana Pérez Gómez',
    '4422223344',
    'Hermana'
  ),
  (
    'cccccccc-1111-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666663', -- Roberto Garcia
    'O+',
    'Polvo y polen',
    'Ninguna',
    'Sofía García',
    '4423334455',
    'Madre'
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 4. OCCUPATIONAL MEDICAL EXAMS (PERIODIC & CLEARANCES)
-- ------------------------------------------------------------
INSERT INTO occupational_medical_exams (
  id, tenant_id, worker_id, contractor_id, exam_number, exam_type, fitness_status,
  exam_date, expiration_date, doctor_name, doctor_license, clinic_or_institution,
  restrictions_notes, recommendations, visual_acuity_result, audiometry_result,
  spirometry_result, antidoping_result, blood_pressure
) VALUES
  (
    'cccccccc-2222-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    '55555555-5555-5555-5555-555555555555',
    'MED-2026-00001',
    'periodic_annual',
    'fit',
    CURRENT_DATE - INTERVAL '40 days',
    CURRENT_DATE + INTERVAL '325 days',
    'Dr. Alejandro Gómez Morales',
    'CED-PROF-8849201',
    'Servicio Médico de Salud Ocupacional Planta Querétaro',
    'Sin restricciones laborales para trabajos pesados ni de altura.',
    'Mantener hidratación adecuada y rotación ergonómica.',
    '20/20 bilateral',
    'Normal sin pérdida auditiva',
    'Normal FVC 92%',
    'Negativo 5 paneles',
    '118/78 mmHg'
  ),
  (
    'cccccccc-2222-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666662', -- Juan Perez
    '55555555-5555-5555-5555-555555555555',
    'MED-2026-00002',
    'height_clearance',
    'fit_with_restrictions',
    CURRENT_DATE - INTERVAL '15 days',
    CURRENT_DATE + INTERVAL '165 days',
    'Dr. Alejandro Gómez Morales',
    'CED-PROF-8849201',
    'Servicio Médico de Salud Ocupacional Planta Querétaro',
    'Uso obligatorio permanente de lentes de seguridad con graduación visual prescrita.',
    'Monitoreo trimestral de presión arterial.',
    '20/25 corregida con lentes',
    'Normal',
    'Normal FVC 88%',
    'Negativo 5 paneles',
    '125/82 mmHg'
  ),
  (
    'cccccccc-2222-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666663', -- Roberto Garcia
    '55555555-5555-5555-5555-555555555555',
    'MED-2026-00003',
    'periodic_annual',
    'fit',
    CURRENT_DATE - INTERVAL '10 days',
    CURRENT_DATE + INTERVAL '355 days',
    'Dra. María Elena Sánchez',
    'CED-PROF-9923184',
    'Servicio Médico de Salud Ocupacional Planta Querétaro',
    'Sin restricciones laborales.',
    'Uso estricto de protección auditiva en naves con > 85 dB.',
    '20/20 bilateral',
    'Normal bilateral',
    'Normal FVC 94%',
    'Negativo 5 paneles',
    '120/80 mmHg'
  )
ON CONFLICT (id) DO NOTHING;

COMMIT;
