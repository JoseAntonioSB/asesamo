-- ============================================================
-- ASESAMO — Phase 1 MVP Deterministic Development Seed Data
-- File: db/seeds/000001_seed_dev_data.sql
-- Description: Seeds deterministic records for end-to-end testing,
--              guard authentication, and Mexican labor compliance
--              verification (DC-3, REPSE, SUA, INE, Medical).
-- ============================================================

BEGIN;

-- ------------------------------------------------------------
-- 1. TEST TENANT
-- Deterministic UUID: 11111111-1111-1111-1111-111111111111
-- Customer: Grupo Industrial Norteño SA de CV
-- ------------------------------------------------------------
INSERT INTO tenants (
  id,
  name,
  rfc,
  logo_r2_key,
  plan,
  is_active
) VALUES (
  '11111111-1111-1111-1111-111111111111',
  'Grupo Industrial Norteño SA de CV',
  'GIN950418KP3',
  'tenants/11111111-1111-1111-1111-111111111111/logo.png',
  'pro',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 2. TEST PLANT
-- Deterministic UUID: 22222222-2222-2222-2222-222222222222
-- Facility: Planta Querétaro Norte (block_on_expired_docs = TRUE)
-- ------------------------------------------------------------
INSERT INTO plants (
  id,
  tenant_id,
  name,
  address,
  city,
  state,
  block_on_expired_docs
) VALUES (
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'Planta Querétaro Norte',
  'Av. de la Industria 100, Parque Industrial Benito Juárez',
  'Querétaro',
  'Querétaro',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 3. TEST ACCESS POINT (CASETA)
-- Deterministic UUID: 33333333-3333-3333-3333-333333333331
-- Checkpoint: Caseta Principal Norte
-- ------------------------------------------------------------
INSERT INTO access_points (
  id,
  plant_id,
  tenant_id,
  name,
  description,
  is_active
) VALUES (
  '33333333-3333-3333-3333-333333333331',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'Caseta Principal Norte',
  'Caseta de vigilancia y torniquete de acceso peatonal principal',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 4. USERS
-- Password for all seed users: Password123!
-- Bcrypt Hash via native pgcrypto: crypt('Password123!', gen_salt('bf', 10))
-- ------------------------------------------------------------

-- Guard User (Deterministic UUID: 44444444-4444-4444-4444-444444444444)
INSERT INTO users (
  id,
  tenant_id,
  email,
  password_hash,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  '44444444-4444-4444-4444-444444444444',
  '11111111-1111-1111-1111-111111111111',
  'guard@asesamo.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Oficial Roberto Caseta',
  '4425551234',
  'guard',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- Plant Admin User (for management tests)
INSERT INTO users (
  id,
  tenant_id,
  email,
  password_hash,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  '44444444-4444-4444-4444-444444444441',
  '11111111-1111-1111-1111-111111111111',
  'admin@asesamo.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Administrador de Planta',
  '4429998877',
  'plant_admin',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- Worker 1 User (Deterministic UUID: 66666666-6666-6666-6666-666666666661)
INSERT INTO users (
  id,
  tenant_id,
  email,
  password_hash,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  '66666666-6666-6666-6666-666666666661',
  '11111111-1111-1111-1111-111111111111',
  'carlos.mendoza@mantenimientonorte.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Carlos Mendoza',
  '4421112233',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- Worker 2 User (Deterministic UUID: 66666666-6666-6666-6666-666666666662)
INSERT INTO users (
  id,
  tenant_id,
  email,
  password_hash,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  '66666666-6666-6666-6666-666666666662',
  '11111111-1111-1111-1111-111111111111',
  'juan.perez@mantenimientonorte.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Juan Perez',
  '4422223344',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- Worker 3 User (Deterministic UUID: 66666666-6666-6666-6666-666666666663)
INSERT INTO users (
  id,
  tenant_id,
  email,
  password_hash,
  full_name,
  phone,
  role,
  is_active
) VALUES (
  '66666666-6666-6666-6666-666666666663',
  '11111111-1111-1111-1111-111111111111',
  'roberto.garcia@mantenimientonorte.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Roberto Garcia',
  '4423334455',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 5. CONTRACTOR COMPANY
-- Deterministic UUID: 55555555-5555-5555-5555-555555555555
-- Name: Mantenimiento Industrial del Norte SA de CV
-- ------------------------------------------------------------
INSERT INTO contractors (
  id,
  name,
  rfc,
  repse_number,
  contractor_type,
  is_active
) VALUES (
  '55555555-5555-5555-5555-555555555555',
  'Mantenimiento Industrial del Norte SA de CV',
  'MIN120304AB1',
  'REPSE-NL-2023-8841',
  'company',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 6. PLANT ↔ CONTRACTOR RELATIONSHIP (ACTIVE)
-- Plant: 22222222-2222-2222-2222-222222222222
-- Contractor: 55555555-5555-5555-5555-555555555555
-- ------------------------------------------------------------
INSERT INTO plant_contractors (
  id,
  plant_id,
  tenant_id,
  contractor_id,
  status,
  invited_by,
  invited_at,
  activated_at
) VALUES (
  '55555555-5555-5555-5555-aaaaaaaaaaaa',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  '55555555-5555-5555-5555-555555555555',
  'active',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '30 days',
  now() - INTERVAL '29 days'
) ON CONFLICT (plant_id, contractor_id) DO NOTHING;

-- ------------------------------------------------------------
-- 7. CONTRACTOR ↔ WORKER AFFILIATIONS
-- Links Workers 1, 2, 3 to Contractor 55555555-5555-5555-5555-555555555555
-- ------------------------------------------------------------
INSERT INTO contractor_workers (
  id,
  contractor_id,
  user_id,
  job_title,
  curp,
  imss_number,
  is_active
) VALUES 
  (
    '66666666-6666-6666-6666-aaaaaaaaaaa1',
    '55555555-5555-5555-5555-555555555555',
    '66666666-6666-6666-6666-666666666661',
    'Soldador Industrial',
    'MENC850101HDFR01',
    '12345678901',
    TRUE
  ),
  (
    '66666666-6666-6666-6666-aaaaaaaaaaa2',
    '55555555-5555-5555-5555-555555555555',
    '66666666-6666-6666-6666-666666666662',
    'Técnico Mecánico',
    'PEPJ900215HDFR02',
    '12345678902',
    TRUE
  ),
  (
    '66666666-6666-6666-6666-aaaaaaaaaaa3',
    '55555555-5555-5555-5555-555555555555',
    '66666666-6666-6666-6666-666666666663',
    'Electricista Industrial',
    'GARR880520HDFR03',
    '12345678903',
    TRUE
  )
ON CONFLICT (contractor_id, user_id) DO NOTHING;

-- ------------------------------------------------------------
-- 8. PLANT DOCUMENT REQUIREMENTS
-- Establishes mandatory compliance documents for Planta Querétaro Norte:
-- Worker: DC3, INE, MEDICAL
-- Contractor: REPSE, SUA
-- Non-mandatory: SAT_32D
-- ------------------------------------------------------------
INSERT INTO plant_document_requirements (
  id,
  plant_id,
  tenant_id,
  document_type_id,
  is_required,
  validity_days
)
SELECT 
  gen_random_uuid(),
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  dt.id,
  TRUE,
  dt.default_validity_days
FROM document_types dt
WHERE dt.tenant_id IS NULL AND dt.code IN ('DC3', 'REPSE', 'SUA', 'INE', 'MEDICAL')
ON CONFLICT (plant_id, document_type_id) DO NOTHING;

INSERT INTO plant_document_requirements (
  id,
  plant_id,
  tenant_id,
  document_type_id,
  is_required,
  validity_days
)
SELECT 
  gen_random_uuid(),
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  dt.id,
  FALSE,
  dt.default_validity_days
FROM document_types dt
WHERE dt.tenant_id IS NULL AND dt.code = 'SAT_32D'
ON CONFLICT (plant_id, document_type_id) DO NOTHING;

-- ------------------------------------------------------------
-- 9. CONTRACTOR DOCUMENTS (Approved REPSE and SUA)
-- ------------------------------------------------------------

-- Contractor REPSE (Approved, valid for 305 more days)
INSERT INTO documents (
  id,
  contractor_id,
  user_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '55555555-5555-5555-5555-bbbbbbbbbb01',
  '55555555-5555-5555-5555-555555555555',
  NULL,
  dt.id,
  'contractors/55555555-5555-5555-5555-555555555555/repse_vigente_2026.pdf',
  'repse_vigente_2026.pdf',
  245000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '60 days',
  CURRENT_DATE + INTERVAL '305 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '59 days'
FROM document_types dt
WHERE dt.code = 'REPSE' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Contractor SUA (Approved, valid for 25 more days)
INSERT INTO documents (
  id,
  contractor_id,
  user_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '55555555-5555-5555-5555-bbbbbbbbbb02',
  '55555555-5555-5555-5555-555555555555',
  NULL,
  dt.id,
  'contractors/55555555-5555-5555-5555-555555555555/sua_pago_vigente.pdf',
  'sua_pago_vigente.pdf',
  180000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '5 days',
  CURRENT_DATE + INTERVAL '25 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '4 days'
FROM document_types dt
WHERE dt.code = 'SUA' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Contractor SAT 32-D (Approved, supplementary)
INSERT INTO documents (
  id,
  contractor_id,
  user_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '55555555-5555-5555-5555-bbbbbbbbbb03',
  '55555555-5555-5555-5555-555555555555',
  NULL,
  dt.id,
  'contractors/55555555-5555-5555-5555-555555555555/sat_32d_opinion_positiva.pdf',
  'sat_32d_opinion_positiva.pdf',
  120000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '10 days',
  CURRENT_DATE + INTERVAL '20 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '9 days'
FROM document_types dt
WHERE dt.code = 'SAT_32D' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 10. WORKER DOCUMENTS
-- ------------------------------------------------------------

-- ============================================================
-- WORKER 1: COMPLIANT WORKER -> Access Result: GRANTED
-- UUID: 66666666-6666-6666-6666-666666666661 (Carlos Mendoza)
-- Documents: Approved DC-3, Approved INE, Approved Medical
-- ============================================================

-- Worker 1: Approved DC-3 (valid for 300 days)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000101',
  '66666666-6666-6666-6666-666666666661',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666661/dc3_seguridad_alturas.pdf',
  'dc3_seguridad_alturas.pdf',
  312000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '65 days',
  CURRENT_DATE + INTERVAL '300 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '60 days'
FROM document_types dt
WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 1: Approved INE (Official ID, no expiry)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000102',
  '66666666-6666-6666-6666-666666666661',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666661/ine_carlos_mendoza.pdf',
  'ine_carlos_mendoza.pdf',
  154000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '300 days',
  NULL,
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '60 days'
FROM document_types dt
WHERE dt.code = 'INE' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 1: Approved Medical Aptitude Certificate (valid for 180 days)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000103',
  '66666666-6666-6666-6666-666666666661',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666661/examen_medico_aptitud.pdf',
  'examen_medico_aptitud.pdf',
  210000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '185 days',
  CURRENT_DATE + INTERVAL '180 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '60 days'
FROM document_types dt
WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- WORKER 2: NON-COMPLIANT WORKER -> Access Result: DENIED
-- UUID: 66666666-6666-6666-6666-666666666662 (Juan Perez)
-- Documents: Expired DC-3, Rejected Medical, Approved INE
-- Denial reasons: DC-3 expired 35 days ago, Medical rejected
-- ============================================================

-- Worker 2: Expired DC-3 (expired 35 days ago)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000201',
  '66666666-6666-6666-6666-666666666662',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666662/dc3_capacitacion_vencida.pdf',
  'dc3_capacitacion_vencida.pdf',
  290000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '400 days',
  CURRENT_DATE - INTERVAL '35 days',
  'expired',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '390 days'
FROM document_types dt
WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 2: Approved INE (Valid)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000202',
  '66666666-6666-6666-6666-666666666662',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666662/ine_juan_perez.pdf',
  'ine_juan_perez.pdf',
  160000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '200 days',
  NULL,
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '190 days'
FROM document_types dt
WHERE dt.code = 'INE' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 2: Rejected Medical Certificate
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at,
  rejection_reason
)
SELECT
  '66666666-6666-6666-6666-000000000203',
  '66666666-6666-6666-6666-666666666662',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666662/certificado_medico_rechazado.pdf',
  'certificado_medico_rechazado.pdf',
  195000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '10 days',
  CURRENT_DATE + INTERVAL '355 days',
  'rejected',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '5 days',
  'Certificado médico sin cédula profesional del médico dictaminador ni prueba antidoping vigente'
FROM document_types dt
WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- WORKER 3: WARNING-CASE WORKER -> Access Result: GRANTED_WITH_WARNING
-- UUID: 66666666-6666-6666-6666-666666666663 (Roberto Garcia)
-- Documents: DC-3 approved but expiring in 5 days, Approved INE, Approved Medical
-- ============================================================

-- Worker 3: Approved DC-3 Expiring Soon (5 days remaining)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000301',
  '66666666-6666-6666-6666-666666666663',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666663/dc3_por_vencer.pdf',
  'dc3_por_vencer.pdf',
  305000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '360 days',
  CURRENT_DATE + INTERVAL '5 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '350 days'
FROM document_types dt
WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 3: Approved INE (Valid)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000302',
  '66666666-6666-6666-6666-666666666663',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666663/ine_roberto_garcia.pdf',
  'ine_roberto_garcia.pdf',
  158000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '150 days',
  NULL,
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '140 days'
FROM document_types dt
WHERE dt.code = 'INE' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Worker 3: Approved Medical Certificate (Valid for 90 days)
INSERT INTO documents (
  id,
  user_id,
  contractor_id,
  document_type_id,
  r2_key,
  file_name,
  file_size_bytes,
  mime_type,
  issued_at,
  expires_at,
  status,
  reviewed_by,
  reviewed_at
)
SELECT
  '66666666-6666-6666-6666-000000000303',
  '66666666-6666-6666-6666-666666666663',
  NULL,
  dt.id,
  'workers/66666666-6666-6666-6666-666666666663/examen_medico_roberto.pdf',
  'examen_medico_roberto.pdf',
  215000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '275 days',
  CURRENT_DATE + INTERVAL '90 days',
  'approved',
  '44444444-4444-4444-4444-444444444444',
  now() - INTERVAL '270 days'
FROM document_types dt
WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

COMMIT;
