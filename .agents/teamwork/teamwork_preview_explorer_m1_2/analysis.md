# Detailed Technical Analysis: Database Seeding Strategy & Implementation

**Milestone:** M1 — Database & Environment Scaffolding  
**Artifact:** `db/seeds/000001_seed_dev_data.sql`  
**Author:** teamwork_preview_explorer_m1_2  
**Target Working Directory:** `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_2`  
**Date:** 2026-10-01  

---

## 1. Executive Summary

Milestone M1 establishes the local database, migration pipeline, and deterministic seed data required for the ASESAMO Phase 1 MVP. This analysis formulates the complete, exact SQL script `db/seeds/000001_seed_dev_data.sql` that populates a development PostgreSQL database with deterministic, reproducible records.

The seed script serves as the foundational data harness for:
1. **Guard Authentication (`POST /api/v1/auth/login`)**: Verifying bcrypt hashes generated via native PostgreSQL `pgcrypto`.
2. **Access Point Check-In Verification (`POST /api/v1/access/check-in`)**: Verifying the three primary real-world compliance outcomes (`granted`, `denied`, and `granted_with_warning`).
3. **Mexican Regulatory Compliance Enforcement**: Accurately reflecting Mexican industrial contractor requirements (STPS DC-3, STPS REPSE, IMSS SUA, INE, and Medical Aptitude).
4. **Automated End-to-End Acceptance Tests**: Providing predictable UUIDs for test runners, cURL verification scripts, and the Next.js Caseta UI.

---

## 2. Database Schema & Dependency Graph Analysis

A detailed inspection of `db/migrations/000001_init.up.sql` reveals 13 core relational tables, several custom Postgres types (enums), and triggers. Foreign key constraints dictate a strict insertion topology.

### 2.1 Relational Dependency DAG (Directed Acyclic Graph)

```
                       ┌──────────────┐
                       │   tenants    │ ◄─── (Level 0: Root customer entity)
                       └──────┬───────┘
                              │
            ┌─────────────────┼─────────────────┐
            │                 │                 │
            ▼                 ▼                 ▼
      ┌───────────┐     ┌───────────┐     ┌───────────┐
      │  plants   │     │   users   │     │ doc_types │ (Pre-seeded in 000001_init.up.sql)
      └─────┬─────┘     └─────┬─────┘     └─────┬─────┘
            │                 │                 │
            ├────────────┬────┴────┬────────────┤
            │            │         │            │
            ▼            ▼         ▼            │
      ┌───────────┐ ┌─────────┐ ┌─────────┐     │
      │access_pts │ │contract.│ │plant_req│ ◄───┘
      └───────────┘ └────┬────┘ └─────────┘
                         │
            ┌────────────┼────────────┐
            ▼            ▼            ▼
      ┌───────────┐ ┌─────────┐ ┌─────────┐
      │plant_contr│ │contr_wrk│ │documents│ ◄─── (Workers & Contractors docs)
      └───────────┘ └─────────┘ └─────────┘
```

### 2.2 Table-by-Table Column and Constraint Matrix

| Table | Dependency Level | Foreign Keys | Unique / Check Constraints | Seeding Strategy |
|---|---|---|---|---|
| `tenants` | 0 | None | PK `id`, `plan IN ('starter','pro','enterprise')` | Explicit UUID `11111111-1111-1111-1111-111111111111` |
| `plants` | 1 | `tenant_id -> tenants(id)` | PK `id`, `block_on_expired_docs` | Explicit UUID `22222222-2222-2222-2222-222222222222`, `block_on_expired_docs = true` |
| `access_points` | 2 | `plant_id -> plants(id)`, `tenant_id -> tenants(id)` | PK `id` | Explicit UUID `33333333-3333-3333-3333-333333333331` |
| `users` | 1 | `tenant_id -> tenants(id)` | PK `id`, `email UNIQUE (citext)` | Guard UUID `44444444-4444-4444-4444-444444444444`, Workers 1-3 UUIDs, Admin UUID |
| `contractors` | 1 | `admin_user_id -> users(id)` (nullable) | PK `id`, `contractor_type ENUM` | Explicit UUID `55555555-5555-5555-5555-555555555555` |
| `plant_contractors` | 2 | `plant_id`, `tenant_id`, `contractor_id`, `invited_by` | PK `id`, `UNIQUE (plant_id, contractor_id)` | Link contractor to plant with `status = 'active'` |
| `contractor_workers` | 2 | `contractor_id`, `user_id` | PK `id`, `UNIQUE (contractor_id, user_id)` | Link Workers 1, 2, 3 to Contractor 5555... |
| `document_types` | 0 | `tenant_id` (nullable) | PK `id`, `UNIQUE (tenant_id, code)` | **Already populated** by migration lines 271-277 with `gen_random_uuid()`! |
| `plant_document_requirements` | 2 | `plant_id`, `tenant_id`, `document_type_id` | PK `id`, `UNIQUE (plant_id, document_type_id)` | Insert requirements by querying `document_types.id` via code |
| `documents` | 3 | `user_id`, `contractor_id`, `document_type_id`, `reviewed_by` | PK `id`, `chk_document_owner CHECK ((user_id IS NOT NULL AND contractor_id IS NULL) OR (user_id IS NULL AND contractor_id IS NOT NULL))` | Insert worker docs (user_id) and contractor docs (contractor_id) with deterministic doc IDs |

---

## 3. Critical Technical Nuance: Dynamic Resolution of `document_types`

In `db/migrations/000001_init.up.sql` (lines 271-277), system document types are seeded on migration execution:
```sql
INSERT INTO document_types (id, tenant_id, code, name, applies_to, default_validity_days, is_required_default, is_system) VALUES
  (gen_random_uuid(), NULL, 'DC3',      'Constancia de Capacitación DC-3',          'worker',      365,  TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'REPSE',    'Registro REPSE',                            'contractor',  365,  TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'SUA',      'Comprobante de Pago SUA (IMSS/INFONAVIT)', 'contractor',  30,   TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'SAT_32D',  'Opinión de Cumplimiento SAT (32-D)',        'contractor',  30,   TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'INE',      'Identificación Oficial (INE/IFE)',          'worker',      NULL, TRUE,  TRUE),
  (gen_random_uuid(), NULL, 'MEDICAL',  'Certificado Médico de Aptitud',             'worker',      365,  TRUE,  TRUE);
```

Because `000001_init.up.sql` uses `gen_random_uuid()` to populate `document_types.id`, hardcoding static UUIDs in foreign key references to `document_types` in `000001_seed_dev_data.sql` would cause immediate foreign key constraint violation crashes.

### The Robust Architectural Solution
The seed script must dynamically resolve foreign keys via subqueries matching the unique `(tenant_id IS NULL AND code = '...')` tuple:
```sql
SELECT dt.id FROM document_types dt WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL
```
This guarantees that regardless of database engine instance, random seed generation in migration, or environment resets, the seed script always connects cleanly to the existing catalog.

---

## 4. Native Password Hashing with PostgreSQL `pgcrypto`

`DEVELOPMENT_RULES.md` and the dispatch specification strictly require native PostgreSQL bcrypt password hashing:
```sql
crypt('Password123!', gen_salt('bf', 10))
```

### Why this is essential:
1. `pgcrypto` is already enabled by `000001_init.up.sql` (line 12: `CREATE EXTENSION IF NOT EXISTS "pgcrypto";`).
2. Using SQL native `crypt(..., gen_salt('bf', 10))` produces a valid `$2a$10$...` Modular Crypt Format (MCF) hash conforming to the standard Golang `golang.org/x/crypto/bcrypt` algorithm.
3. When the Go backend (`asesamo-api`) executes `bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))` during `POST /api/v1/auth/login`, verification succeeds seamlessly.
4. Using an identical known password (`Password123!`) across guard, admin, and worker test accounts eliminates test ambiguity.

---

## 5. Mexican Industrial Safety Compliance Context & Seed Personas

ASESAMO enforces statutory Mexican industrial contractor compliance. The test records mirror this framework:

### 5.1 Mexican Regulatory Standards Encoded
- **STPS Formato DC-3**: "Constancia de competencias o de habilidades laborales" under Articles 153-A to 153-X of the Ley Federal del Trabajo (LFT). Validates hazardous job authorization (NOM-009-STPS for working at heights, NOM-027-STPS for welding/cutting, NOM-029-STPS for electrical maintenance).
- **STPS REPSE**: "Registro de Prestadoras de Servicios Especializados u Obras Especializadas". Mandatory legal registry for external contractors operating within a third party's plant under the 2021 Labor Subcontracting Reform.
- **IMSS SUA**: "Sistema Único de Autodeterminación". Proof of monthly employer social security and housing fund (INFONAVIT) contributions. Ensures workers on-site are actively insured against workplace accidents.
- **INE / Cédula**: Official voter/national identification confirming worker identity.
- **Examen Médico de Aptitud**: Industrial health clearance verifying physical fitness for designated hazardous work, drug screening, and audiometry/spirometry.

### 5.2 Deterministic Entity Catalog

#### 1. Tenant
- **UUID**: `11111111-1111-1111-1111-111111111111`
- **Name**: "Grupo Industrial Norteño SA de CV"
- **RFC**: `GIN950418KP3`
- **Plan**: `pro`
- **Role**: Primary manufacturing host organization.

#### 2. Plant
- **UUID**: `22222222-2222-2222-2222-222222222222`
- **Tenant ID**: `11111111-1111-1111-1111-111111111111`
- **Name**: "Planta Querétaro Norte"
- **Location**: Parque Industrial Benito Juárez, Querétaro
- **Policy**: `block_on_expired_docs = TRUE` (enforces strict blocking at the caseta).

#### 3. Access Point (Caseta)
- **UUID**: `33333333-3333-3333-3333-333333333331`
- **Plant ID**: `22222222-2222-2222-2222-222222222222`
- **Tenant ID**: `11111111-1111-1111-1111-111111111111`
- **Name**: "Caseta Principal Norte"
- **Description**: "Caseta de vigilancia y torniquete de acceso peatonal principal"

#### 4. Guard User
- **UUID**: `44444444-4444-4444-4444-444444444444`
- **Tenant ID**: `11111111-1111-1111-1111-111111111111`
- **Email**: `guard@asesamo.com`
- **Full Name**: "Oficial Roberto Caseta"
- **Role**: `guard`
- **Password**: `Password123!` (hashed via `crypt('Password123!', gen_salt('bf', 10))`)

#### 5. Contractor Company
- **UUID**: `55555555-5555-5555-5555-555555555555`
- **Name**: "Mantenimiento Industrial del Norte SA de CV"
- **RFC**: `MIN120304AB1`
- **REPSE**: `REPSE-NL-2023-8841`
- **Status in Plant**: `active` (linked via `plant_contractors`)
- **Required Documents**:
  - `REPSE`: Approved, valid for 305 days (`expires_at = CURRENT_DATE + INTERVAL '305 days'`).
  - `SUA`: Approved, valid for 25 days (`expires_at = CURRENT_DATE + INTERVAL '25 days'`).
  - `SAT_32D`: Approved, supplementary tax compliance opinion.

#### 6. Worker 1: Carlos Mendoza (Compliant -> GRANTED)
- **User UUID**: `66666666-6666-6666-6666-666666666661`
- **Full Name**: "Carlos Mendoza"
- **Email**: `carlos.mendoza@mantenimientonorte.com`
- **CURP**: `MENC850101HDFR01`
- **NSS / IMSS**: `12345678901`
- **Job Title**: "Soldador Industrial" (DC-3 in Hot Work & Heights)
- **Documents**:
  1. `DC3`: Approved, valid for 300 days (`expires_at = CURRENT_DATE + INTERVAL '300 days'`).
  2. `INE`: Approved, valid, no expiration date (`expires_at = NULL`).
  3. `MEDICAL`: Approved, valid for 180 days (`expires_at = CURRENT_DATE + INTERVAL '180 days'`).
- **Expected Check-In Result**: `granted` (HTTP 200, emerald card in UI, access granted log inserted).

#### 7. Worker 2: Juan Perez (Non-Compliant -> DENIED)
- **User UUID**: `66666666-6666-6666-6666-666666666662`
- **Full Name**: "Juan Perez"
- **Email**: `juan.perez@mantenimientonorte.com`
- **CURP**: `PEPJ900215HDFR02`
- **NSS / IMSS**: `12345678902`
- **Job Title**: "Técnico Mecánico"
- **Documents**:
  1. `DC3`: **Expired** 35 days ago (`expires_at = CURRENT_DATE - INTERVAL '35 days'`, status = `'expired'`).
  2. `INE`: Approved, valid (`expires_at = NULL`).
  3. `MEDICAL`: **Rejected** by EHS (`status = 'rejected'`, reason: `'Certificado médico sin cédula profesional del médico dictaminador ni prueba antidoping vigente'`).
- **Expected Check-In Result**: `denied` (HTTP 200, crimson card in UI, detailed denial reasons in `access_logs`).

#### 8. Worker 3: Roberto Garcia (Expiring Soon -> GRANTED_WITH_WARNING)
- **User UUID**: `66666666-6666-6666-6666-666666666663`
- **Full Name**: "Roberto Garcia"
- **Email**: `roberto.garcia@mantenimientonorte.com`
- **CURP**: `GARR880520HDFR03`
- **NSS / IMSS**: `12345678903`
- **Job Title**: "Electricista Industrial"
- **Documents**:
  1. `DC3`: Approved, but **expires in 5 days** (`expires_at = CURRENT_DATE + INTERVAL '5 days'`).
  2. `INE`: Approved, valid (`expires_at = NULL`).
  3. `MEDICAL`: Approved, valid for 90 days (`expires_at = CURRENT_DATE + INTERVAL '90 days'`).
- **Expected Check-In Result**: `granted_with_warning` (HTTP 200, amber card in UI, warning details in audit snapshot).

---

## 6. Idempotency and Time-Shift Resilience

A common failure mode in database seed scripts is hardcoding static dates (e.g. `2026-10-01`) or crashing on re-execution.

### 6.1 Idempotency via `ON CONFLICT DO NOTHING`
Every `INSERT` in `000001_seed_dev_data.sql` includes an explicit `ON CONFLICT` clause:
- For primary keys: `ON CONFLICT (id) DO NOTHING`
- For composite relationships:
  - `plant_contractors`: `ON CONFLICT (plant_id, contractor_id) DO NOTHING`
  - `contractor_workers`: `ON CONFLICT (contractor_id, user_id) DO NOTHING`
  - `plant_document_requirements`: `ON CONFLICT (plant_id, document_type_id) DO NOTHING`

Re-running `psql -f db/seeds/000001_seed_dev_data.sql` or restarting the `seeder` Docker container is completely safe and produces zero duplicate key errors.

### 6.2 Relative Date Calculations (`CURRENT_DATE +/- INTERVAL`)
Using expressions like:
- `CURRENT_DATE + INTERVAL '300 days'` (for valid documents)
- `CURRENT_DATE - INTERVAL '35 days'` (for expired documents)
- `CURRENT_DATE + INTERVAL '5 days'` (for documents expiring soon)

guarantees that test suites will pass regardless of the exact calendar day or year they are run, avoiding test rot.

---

## 7. Complete SQL Seed Script Formulation

The complete SQL formulation is saved in the agent directory as `proposed_000001_seed_dev_data.sql` and ready for destination `db/seeds/000001_seed_dev_data.sql`:

```sql
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
```

---

## 8. Alignment with Upstream and Downstream Milestones

### 8.1 Docker Compose Seeder Service Alignment (M1 Explorer 1 & 3)
The Docker Compose `seeder` service can execute this script with standard Postgres tools:
```yaml
seeder:
  image: postgres:16-alpine
  container_name: asesamo-seeder
  restart: "no"
  depends_on:
    migrate:
      condition: service_completed_successfully
  volumes:
    - ./db/seeds:/seeds:ro
  environment:
    POSTGRES_USER: ${DB_USER:-asesamo_admin}
    POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
    POSTGRES_DB: ${DB_NAME:-asesamo_db}
  entrypoint:
    [
      "sh",
      "-c",
      "PGPASSWORD=$${POSTGRES_PASSWORD} psql -h db -U $${POSTGRES_USER} -d $${POSTGRES_DB} -f /seeds/000001_seed_dev_data.sql"
    ]
```

### 8.2 Makefile Target Alignment
```makefile
seed:
	PGPASSWORD=$(DB_PASSWORD) psql -h localhost -U $(DB_USER) -d $(DB_NAME) -f db/seeds/000001_seed_dev_data.sql
```

### 8.3 Backend (`asesamo-api`) & E2E Testing Suite Alignment
- When `POST /api/v1/auth/login` receives `{"email": "guard@asesamo.com", "password": "Password123!"}`, it matches user `44444444-4444-4444-4444-444444444444` and returns a signed JWT containing claims `sub: 44444444-4444-4444-4444-444444444444`, `tenant_id: 11111111-1111-1111-1111-111111111111`, and `role: guard`.
- When `POST /api/v1/access/check-in` is called with Worker 1 `66666666-6666-6666-6666-666666666661`, all documents evaluate valid, returning `result: "granted"`.
- When called with Worker 2 `66666666-6666-6666-6666-666666666662`, `result: "denied"` is returned due to expired DC-3 and rejected medical.
- When called with Worker 3 `66666666-6666-6666-6666-666666666663`, `result: "granted_with_warning"` is returned due to DC-3 expiring in 5 days.
