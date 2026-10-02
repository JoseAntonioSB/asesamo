# ASESAMO Phase 1 MVP — Infrastructure, Environment & End-to-End Integration Survey

**Document Version:** 1.0.0  
**Author:** teamwork_preview_explorer_survey_3 (Integration & Infrastructure Explorer)  
**Date:** 2026-10-01  
**Working Directory:** `/Users/jasbotello/devel/asesamo`  
**Integrity Mode:** Demo / Phase 1 MVP  

---

## 1. Executive Summary

This survey defines the foundational infrastructure, container orchestration, database migration and seeding strategy, cross-module interface contracts, and end-to-end acceptance criteria verification procedures for the **ASESAMO Phase 1 MVP**.

ASESAMO is an industrial access control SaaS designed to enforce Mexican occupational safety regulations (STPS NOM-009, NOM-027, NOM-029, DC-3, IMSS/SUA, and labor reform REPSE) at plant entry checkpoints (*casetas*). The MVP comprises:
1. **`asesamo-api`**: High-performance Go (Golang 1.23+) REST API built on Chi, adhering to Clean Architecture and utilizing `sqlc` for type-safe PostgreSQL interactions.
2. **`asesamo-web`**: Next.js 14+ App Router frontend built with TypeScript strict mode, styled with Tailwind CSS and shadcn/ui components, providing the Guard Caseta Check-in terminal.
3. **`db`**: PostgreSQL 16 database with `pgcrypto` (UUID generation and bcrypt hashing) and `citext` (case-insensitive emails) extensions.
4. **`migrate` & `seeder`**: Containerized database migration engine (golang-migrate) and deterministic seed harness for instant local evaluation.

All services are orchestrated via a single root `docker-compose.yml` supporting instant bootstrap (`docker compose up --build`), clean dependency-ordered initialization, and zero-configuration local execution.

---

## 2. Docker Compose Infrastructure Setup (`docker-compose.yml`)

### 2.1 Orchestration Architecture

The system topology requires a clean dependency lifecycle:
1. `db` starts and initializes PostgreSQL 16.
2. `migrate` waits for `db` to pass health check, executes `db/migrations/000001_init.up.sql`, and exits with code 0.
3. `seeder` waits for `migrate` to complete successfully, injects deterministic test data, and exits with code 0.
4. `api` waits for `migrate` and `seeder` to complete and `db` to be healthy, starts on port `8080`, and exposes `/api/v1/health`.
5. `web` waits for `api` to pass health check and serves the Next.js application on port `3000`.

```
               ┌────────────────────────┐
               │    PostgreSQL 16       │
               │      (asesamo-db)      │
               └───────────┬────────────┘
                           │ (service_healthy)
               ┌───────────▼────────────┐
               │     golang-migrate     │
               │   (asesamo-migrate)    │
               └───────────┬────────────┘
                           │ (service_completed_successfully)
               ┌───────────▼────────────┐
               │       DB Seeder        │
               │    (asesamo-seeder)    │
               └───────────┬────────────┘
                           │ (service_completed_successfully)
               ┌───────────▼────────────┐
               │      Go Backend        │
               │    (asesamo-api)       │
               └───────────┬────────────┘
                           │ (service_healthy)
               ┌───────────▼────────────┐
               │   Next.js Frontend     │
               │    (asesamo-web)       │
               └────────────────────────┘
```

### 2.2 Root `docker-compose.yml` Specification

```yaml
version: "3.9"

services:
  db:
    image: postgres:16-alpine
    container_name: asesamo-db
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: ${DB_USER:-asesamo_admin}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-asesamo_secret_dev}
      POSTGRES_DB: ${DB_NAME:-asesamo_db}
    volumes:
      - asesamo_pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 5s
    networks:
      - asesamo-net

  migrate:
    image: migrate/migrate:v4.17.0
    container_name: asesamo-migrate
    restart: "no"
    depends_on:
      db:
        condition: service_healthy
    volumes:
      - ./db/migrations:/migrations:ro
    command:
      [
        "-path=/migrations",
        "-database",
        "postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable",
        "up"
      ]
    networks:
      - asesamo-net

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
    networks:
      - asesamo-net

  api:
    build:
      context: ./asesamo-api
      dockerfile: Dockerfile
    container_name: asesamo-api
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
      migrate:
        condition: service_completed_successfully
      seeder:
        condition: service_completed_successfully
    environment:
      APP_ENV: development
      APP_PORT: 8080
      DATABASE_URL: postgres://${DB_USER:-asesamo_admin}:${DB_PASSWORD:-asesamo_secret_dev}@db:5432/${DB_NAME:-asesamo_db}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET:-super_secret_jwt_key_min_32_characters_long_12345}
      CORS_ALLOWED_ORIGINS: "http://localhost:3000,http://127.0.0.1:3000,http://web:3000"
      R2_ACCOUNT_ID: ${R2_ACCOUNT_ID:-mock_account}
      R2_ACCESS_KEY_ID: ${R2_ACCESS_KEY_ID:-mock_key}
      R2_SECRET_ACCESS_KEY: ${R2_SECRET_ACCESS_KEY:-mock_secret}
      R2_BUCKET_NAME: ${R2_BUCKET_NAME:-asesamo-dev}
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - asesamo-net

  web:
    build:
      context: ./asesamo-web
      dockerfile: Dockerfile
      args:
        NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1
    container_name: asesamo-web
    restart: unless-stopped
    ports:
      - "3000:3000"
    depends_on:
      api:
        condition: service_healthy
    environment:
      NODE_ENV: production
      PORT: 3000
      NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1
      INTERNAL_API_URL: http://api:8080/api/v1
    healthcheck:
      test: ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1"]
      interval: 5s
      timeout: 5s
      retries: 5
      start_period: 15s
    networks:
      - asesamo-net

volumes:
  asesamo_pgdata:
    name: asesamo_pgdata

networks:
  asesamo-net:
    name: asesamo-net
    driver: bridge
```

### 2.3 Container Dockerfiles

#### 2.3.1 Backend Dockerfile (`asesamo-api/Dockerfile`)
```dockerfile
# Stage 1: Build binary
FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api ./cmd/api

# Stage 2: Minimal runtime
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /app/bin/api /app/bin/api

EXPOSE 8080

ENTRYPOINT ["/app/bin/api"]
```

#### 2.3.2 Frontend Dockerfile (`asesamo-web/Dockerfile`)
```dockerfile
# Stage 1: Install dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci

# Stage 2: Build Next.js application
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_TELEMETRY_DISABLED=1

RUN npm run build

# Stage 3: Minimal runner
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV NEXT_TELEMETRY_DISABLED=1

RUN apk add --no-cache wget

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

EXPOSE 3000

CMD ["npm", "run", "start"]
```

### 2.4 Environment Variables Specification

#### Configuration File: `.env.example`
```dotenv
# ============================================================
# ASESAMO — Local Environment Configuration
# ============================================================

# Database (PostgreSQL 16)
DB_USER=asesamo_admin
DB_PASSWORD=asesamo_secret_dev
DB_NAME=asesamo_db
DB_PORT=5432

# Backend Connection URLs
DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable
DOCKER_DATABASE_URL=postgres://asesamo_admin:asesamo_secret_dev@db:5432/asesamo_db?sslmode=disable

# Go Backend API
APP_ENV=development
APP_PORT=8080
JWT_SECRET=super_secret_jwt_key_min_32_characters_long_12345
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

# Cloudflare R2 Storage (Mock / Dev credentials)
R2_ACCOUNT_ID=mock_account
R2_ACCESS_KEY_ID=mock_key
R2_SECRET_ACCESS_KEY=mock_secret
R2_BUCKET_NAME=asesamo-dev

# Next.js Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
INTERNAL_API_URL=http://api:8080/api/v1
PORT=3000
```

---

## 3. Migration and Database Seeding Strategy

### 3.1 Migration Lifecycle (`000001_init.up.sql`)

The migration `db/migrations/000001_init.up.sql` establishes:
1. **Extensions**: `pgcrypto` and `citext`.
2. **Triggers**: `set_updated_at()` applied across all mutable entities.
3. **Core Tables (13 tables)**:
   - Multi-tenant roots: `tenants`, `plants`, `access_points`.
   - Security & Identity: `users`, `refresh_tokens`.
   - Contractor entities: `contractors`, `plant_contractors`, `contractor_workers`.
   - Compliance & Documents: `document_types`, `plant_document_requirements`, `documents`.
   - Operation & Audit: `access_logs` (immutable event log with JSONB snapshot), `notifications`.
4. **System Document Types**:
   - `DC3` (Constancia de Capacitación, worker)
   - `REPSE` (Registro REPSE, contractor)
   - `SUA` (Pago IMSS/INFONAVIT, contractor)
   - `SAT_32D` (Opinión 32-D, contractor)
   - `INE` (Identificación Oficial, worker)
   - `MEDICAL` (Certificado Médico, worker)

### 3.2 Deterministic Seed Script Specification (`db/seeds/000001_seed_dev_data.sql`)

To support instant end-to-end testing, reproduction of all acceptance criteria, and browser verification, the seed script uses explicit, well-known UUIDs and native `pgcrypto` bcrypt password hashing (`crypt('Password123!', gen_salt('bf', 10))`).

```sql
-- ============================================================
-- ASESAMO — Phase 1 MVP Deterministic Seed Data
-- Target: db/seeds/000001_seed_dev_data.sql
-- ============================================================

BEGIN;

-- 1. TEST TENANT
INSERT INTO tenants (id, name, rfc, plan, is_active)
VALUES (
  '11111111-1111-1111-1111-111111111111',
  'Aceros Industriales del Bajío S.A. de C.V.',
  'AIB980315TR8',
  'pro',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- 2. TEST PLANT
INSERT INTO plants (id, tenant_id, name, address, city, state, block_on_expired_docs)
VALUES (
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'Planta Matriz Querétaro (Parque Industrial Benito Juárez)',
  'Av. Industrialización 105',
  'Querétaro',
  'Querétaro',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- 3. TEST ACCESS POINT (CASETA)
INSERT INTO access_points (id, plant_id, tenant_id, name, description, is_active)
VALUES (
  '33333333-3333-3333-3333-333333333333',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  'Caseta 1 - Acceso Peatonal Principal',
  'Torniquete y módulo de control de contratistas',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- 4. GUARD USER (Credentials: guard@asesamo.com / Password123!)
INSERT INTO users (id, tenant_id, email, password_hash, full_name, phone, role, is_active)
VALUES (
  '44444444-4444-4444-4444-444444444444',
  '11111111-1111-1111-1111-111111111111',
  'guard@asesamo.com',
  crypt('Password123!', gen_salt('bf', 10)),
  'Oficial Juan Gabriel Mendoza (Guardia Caseta)',
  '4425551234',
  'guard',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- 5. CONTRACTOR COMPANY (GLOBAL ENTITY)
INSERT INTO contractors (id, name, rfc, repse_number, contractor_type, is_active)
VALUES (
  '55555555-5555-5555-5555-555555555555',
  'Mantenimiento Industrial y Soldadura Especializada S.A. de C.V.',
  'MIS090814KK2',
  'AR-55420/2023',
  'company',
  TRUE
) ON CONFLICT (id) DO NOTHING;

-- 6. PLANT ↔ CONTRACTOR RELATIONSHIP (ACTIVE)
INSERT INTO plant_contractors (id, plant_id, tenant_id, contractor_id, status, activated_at)
VALUES (
  '55555555-5555-5555-5555-aaaaaaaaaaaa',
  '22222222-2222-2222-2222-222222222222',
  '11111111-1111-1111-1111-111111111111',
  '55555555-5555-5555-5555-555555555555',
  'active',
  now()
) ON CONFLICT (plant_id, contractor_id) DO NOTHING;

-- 7. PLANT DOCUMENT REQUIREMENTS
-- Set required documents for Planta Matriz Querétaro
INSERT INTO plant_document_requirements (id, plant_id, tenant_id, document_type_id, is_required, validity_days)
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

-- 8. CONTRACTOR COMPANY DOCUMENTS (REPSE & SUA)
-- Valid REPSE document
INSERT INTO documents (
  id, contractor_id, user_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type,
  issued_at, expires_at, status
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
  'approved'
FROM document_types dt
WHERE dt.code = 'REPSE' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- Valid SUA payment receipt
INSERT INTO documents (
  id, contractor_id, user_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type,
  issued_at, expires_at, status
)
SELECT
  '55555555-5555-5555-5555-bbbbbbbbbb02',
  '55555555-5555-5555-5555-555555555555',
  NULL,
  dt.id,
  'contractors/55555555-5555-5555-5555-555555555555/sua_pago_oct2026.pdf',
  'sua_pago_oct2026.pdf',
  180000,
  'application/pdf',
  CURRENT_DATE - INTERVAL '5 days',
  CURRENT_DATE + INTERVAL '25 days',
  'approved'
FROM document_types dt
WHERE dt.code = 'SUA' AND dt.tenant_id IS NULL
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 9. TEST WORKERS & DOCUMENT CONFIGURATIONS
-- ============================================================

-- ------------------------------------------------------------
-- WORKER 1: COMPLIANT WORKER (Expected Check-in: GRANTED)
-- UUID: 66666666-6666-6666-6666-666666666661
-- ------------------------------------------------------------
INSERT INTO users (id, tenant_id, email, password_hash, full_name, phone, role, is_active)
VALUES (
  '66666666-6666-6666-6666-666666666661',
  '11111111-1111-1111-1111-111111111111',
  'carlos.mendoza@mantenimientoind.mx',
  crypt('Password123!', gen_salt('bf', 10)),
  'Carlos Mendoza Rodríguez',
  '4421112233',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO contractor_workers (contractor_id, user_id, job_title, curp, imss_number, is_active)
VALUES (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666661',
  'Técnico Soldador Especialista en Alturas',
  'MERC850101HDFR01',
  '12850123456',
  TRUE
) ON CONFLICT (contractor_id, user_id) DO NOTHING;

-- Worker 1 DC3 (Valid: expires in 300 days)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000101', '66666666-6666-6666-6666-666666666661', NULL, dt.id, 'workers/carlos_dc3.pdf', 'dc3_alturas.pdf', 312000, 'application/pdf', CURRENT_DATE - INTERVAL '65 days', CURRENT_DATE + INTERVAL '300 days', 'approved'
FROM document_types dt WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 1 INE (Valid: no expiration)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000102', '66666666-6666-6666-6666-666666666661', NULL, dt.id, 'workers/carlos_ine.pdf', 'ine_carlos.pdf', 154000, 'application/pdf', CURRENT_DATE - INTERVAL '300 days', NULL, 'approved'
FROM document_types dt WHERE dt.code = 'INE' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 1 Medical (Valid: expires in 180 days)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000103', '66666666-6666-6666-6666-666666666661', NULL, dt.id, 'workers/carlos_med.pdf', 'examen_medico_apto.pdf', 210000, 'application/pdf', CURRENT_DATE - INTERVAL '185 days', CURRENT_DATE + INTERVAL '180 days', 'approved'
FROM document_types dt WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;


-- ------------------------------------------------------------
-- WORKER 2: NON-COMPLIANT WORKER (Expected Check-in: DENIED)
-- UUID: 66666666-6666-6666-6666-666666666662
-- Reason: DC-3 Expired 45 days ago, Medical Certificate Rejected
-- ------------------------------------------------------------
INSERT INTO users (id, tenant_id, email, password_hash, full_name, phone, role, is_active)
VALUES (
  '66666666-6666-6666-6666-666666666662',
  '11111111-1111-1111-1111-111111111111',
  'roberto.morales@mantenimientoind.mx',
  crypt('Password123!', gen_salt('bf', 10)),
  'Roberto Morales Santos',
  '4423334455',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO contractor_workers (contractor_id, user_id, job_title, curp, imss_number, is_active)
VALUES (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666662',
  'Ayudante General de Mantenimiento',
  'MOSR920412HDFR02',
  '12920498765',
  TRUE
) ON CONFLICT (contractor_id, user_id) DO NOTHING;

-- Worker 2 DC3 (EXPIRED: expired 45 days ago)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000201', '66666666-6666-6666-6666-666666666662', NULL, dt.id, 'workers/roberto_dc3.pdf', 'dc3_antiguo.pdf', 290000, 'application/pdf', CURRENT_DATE - INTERVAL '410 days', CURRENT_DATE - INTERVAL '45 days', 'expired'
FROM document_types dt WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 2 INE (Valid)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000202', '66666666-6666-6666-6666-666666666662', NULL, dt.id, 'workers/roberto_ine.pdf', 'ine_roberto.pdf', 160000, 'application/pdf', CURRENT_DATE - INTERVAL '200 days', NULL, 'approved'
FROM document_types dt WHERE dt.code = 'INE' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 2 Medical (REJECTED by EHS)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status, rejection_reason)
SELECT '66666666-6666-6666-6666-000000000203', '66666666-6666-6666-6666-666666666662', NULL, dt.id, 'workers/roberto_med.pdf', 'certificado_rechazado.pdf', 195000, 'application/pdf', CURRENT_DATE - INTERVAL '10 days', CURRENT_DATE + INTERVAL '355 days', 'rejected', 'Certificado sin cédula profesional ni firma de médico dictaminador'
FROM document_types dt WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;


-- ------------------------------------------------------------
-- WORKER 3: WARNING-CASE WORKER (Expected Check-in: GRANTED_WITH_WARNING)
-- UUID: 66666666-6666-6666-6666-666666666663
-- Reason: DC-3 approved but expires in 5 days (threshold alert)
-- ------------------------------------------------------------
INSERT INTO users (id, tenant_id, email, password_hash, full_name, phone, role, is_active)
VALUES (
  '66666666-6666-6666-6666-666666666663',
  '11111111-1111-1111-1111-111111111111',
  'mateo.vargas@mantenimientoind.mx',
  crypt('Password123!', gen_salt('bf', 10)),
  'Mateo Vargas Silva',
  '4428889900',
  'worker',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO contractor_workers (contractor_id, user_id, job_title, curp, imss_number, is_active)
VALUES (
  '55555555-5555-5555-5555-555555555555',
  '66666666-6666-6666-6666-666666666663',
  'Electricista de Mantenimiento',
  'VASM880720HDFR03',
  '12880765432',
  TRUE
) ON CONFLICT (contractor_id, user_id) DO NOTHING;

-- Worker 3 DC3 (Expiring soon: 5 days remaining)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000301', '66666666-6666-6666-6666-666666666663', NULL, dt.id, 'workers/mateo_dc3.pdf', 'dc3_electricidad.pdf', 305000, 'application/pdf', CURRENT_DATE - INTERVAL '360 days', CURRENT_DATE + INTERVAL '5 days', 'approved'
FROM document_types dt WHERE dt.code = 'DC3' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 3 INE (Valid)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000302', '66666666-6666-6666-6666-666666666663', NULL, dt.id, 'workers/mateo_ine.pdf', 'ine_mateo.pdf', 158000, 'application/pdf', CURRENT_DATE - INTERVAL '150 days', NULL, 'approved'
FROM document_types dt WHERE dt.code = 'INE' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

-- Worker 3 Medical (Valid: expires in 90 days)
INSERT INTO documents (id, user_id, contractor_id, document_type_id, r2_key, file_name, file_size_bytes, mime_type, issued_at, expires_at, status)
SELECT '66666666-6666-6666-6666-000000000303', '66666666-6666-6666-6666-666666666663', NULL, dt.id, 'workers/mateo_med.pdf', 'medico_mateo.pdf', 215000, 'application/pdf', CURRENT_DATE - INTERVAL '275 days', CURRENT_DATE + INTERVAL '90 days', 'approved'
FROM document_types dt WHERE dt.code = 'MEDICAL' AND dt.tenant_id IS NULL ON CONFLICT (id) DO NOTHING;

COMMIT;
```

---

## 4. End-to-End Acceptance Criteria Verification Matrix

The following matrix documents the exact command, expected response, assertion check, and validation method for each acceptance criterion specified in `ORIGINAL_REQUEST.md`.

| Acceptance Criterion | Verification Method / Command | Expected Result | Pass Criteria |
|---|---|---|---|
| **AC-1: Docker Compose Boot** | `docker compose up --build -d` | All containers reach steady state | `asesamo-db`, `asesamo-api`, `asesamo-web` healthy; `asesamo-migrate`, `asesamo-seeder` exited with code 0. |
| **AC-2: DB Migration Clean Apply** | `docker compose exec -T db psql -U asesamo_admin -d asesamo_db -c "\dt"` | Lists 13 tables | Tables `tenants`, `plants`, `access_points`, `users`, `refresh_tokens`, `contractors`, `plant_contractors`, `contractor_workers`, `document_types`, `plant_document_requirements`, `documents`, `access_logs`, `notifications` present. |
| **AC-3: Auth Generates JWT** | `curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"email":"guard@asesamo.com","password":"Password123!"}'` | HTTP 200 with JWT envelope | Response contains `"token"`, user object with `"role": "guard"`, `"tenant_id": "11111111-1111-1111-1111-111111111111"`. |
| **AC-4: Check-in Granted** | `curl -s -X POST http://localhost:8080/api/v1/access/check-in -H "Authorization: Bearer $TOKEN" -d '{"worker_id":"66666666-6666-6666-6666-666666666661","access_point_id":"33333333-3333-3333-3333-333333333333"}'` | HTTP 200 with `result: "granted"` | `access_logs` row inserted with `result = 'granted'`, `denial_reason IS NULL`, JSONB `document_snapshot` contains approved docs. |
| **AC-5: Check-in Denied** | `curl -s -X POST http://localhost:8080/api/v1/access/check-in -H "Authorization: Bearer $TOKEN" -d '{"worker_id":"66666666-6666-6666-6666-666666666662","access_point_id":"33333333-3333-3333-3333-333333333333"}'` | HTTP 200 with `result: "denied"` | `access_logs` row inserted with `result = 'denied'`, `denial_reason` explicitly details expired DC-3 and rejected medical. |
| **AC-6: Automated Tests** | `cd asesamo-api && go test -v -race -cover ./...` | Tests pass with 0 failures | Exit code 0, 100% of unit & integration tests pass, use case coverage >= 80%. |
| **AC-7: E2E Browser Flow** | Browser navigates to `http://localhost:3000/login`, enters guard credentials, visits `/caseta`, enters worker UUIDs | Interactive UI status | Worker 1 renders Emerald/Green Granted card; Worker 2 renders Crimson/Red Denied card with detail list and Sonner toast. |

### 4.1 Automated Verification Script (`scripts/verify_acceptance.sh`)

To enable automated CI or smoke verification, the following script can be executed against the booted stack:

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "==> [1/6] Verifying Docker Services..."
docker compose ps --filter "status=running" | grep -q "asesamo-api"
docker compose ps --filter "status=running" | grep -q "asesamo-web"
echo "✔ Services are running."

echo "==> [2/6] Verifying Database Schema..."
TABLE_COUNT=$(docker compose exec -T db psql -U asesamo_admin -d asesamo_db -t -c \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';")
if [ "${TABLE_COUNT// /}" -lt 13 ]; then
  echo "✖ Missing database tables. Found: $TABLE_COUNT"
  exit 1
fi
echo "✔ All 13 schema tables verified."

echo "==> [3/6] Testing Authentication (Guard Login)..."
LOGIN_RES=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"guard@asesamo.com","password":"Password123!"}')
HTTP_STATUS=$(echo "$LOGIN_RES" | tail -n1)
BODY=$(echo "$LOGIN_RES" | sed '$d')

if [ "$HTTP_STATUS" != "200" ]; then
  echo "✖ Login failed with status $HTTP_STATUS: $BODY"
  exit 1
fi
TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
echo "✔ Guard JWT generated successfully."

echo "==> [4/6] Testing Check-In: Compliant Worker (Granted)..."
CHECKIN_GRANTED=$(curl -s -X POST http://localhost:8080/api/v1/access/check-in \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "worker_id": "66666666-6666-6666-6666-666666666661",
    "access_point_id": "33333333-3333-3333-3333-333333333333"
  }')
echo "$CHECKIN_GRANTED" | grep -q '"result":"granted"'
echo "✔ Access granted verified."

echo "==> [5/6] Testing Check-In: Non-Compliant Worker (Denied)..."
CHECKIN_DENIED=$(curl -s -X POST http://localhost:8080/api/v1/access/check-in \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "worker_id": "66666666-6666-6666-6666-666666666662",
    "access_point_id": "33333333-3333-3333-3333-333333333333"
  }')
echo "$CHECKIN_DENIED" | grep -q '"result":"denied"'
echo "✔ Access denied with documented reason verified."

echo "==> [6/6] Verifying Immutable Access Logs in Database..."
LOG_COUNT=$(docker compose exec -T db psql -U asesamo_admin -d asesamo_db -t -c \
  "SELECT count(*) FROM access_logs WHERE plant_id = '22222222-2222-2222-2222-222222222222';")
if [ "${LOG_COUNT// /}" -lt 2 ]; then
  echo "✖ Access logs not recorded in database."
  exit 1
fi
echo "✔ Access audit logs successfully persisted with JSONB snapshots."

echo "=========================================================="
echo "✔ ALL ASESAMO PHASE 1 ACCEPTANCE CRITERIA PASS CLEANLY!"
echo "=========================================================="
```

---

## 5. Dependency Ordering and Cross-Module Interface Contracts

### 5.1 Build & Implementation Topological Order

To avoid circular dependencies and integration bottlenecks, development must strictly follow this dependency order:

```
[Level 0] Database Schema (db/migrations/000001_init.up.sql)
    ↓
[Level 1] Database Seed Script (db/seeds/000001_seed_dev_data.sql)
    ↓
[Level 2] Backend Domain Entities & Repository Contracts (asesamo-api/internal/domain/...)
    ↓
[Level 3] Backend Persistence Layer with sqlc (asesamo-api/internal/adapter/postgres/...)
    ↓
[Level 4] Backend Business Use Cases (Auth, Check-In Evaluator, Document Engine)
    ↓
[Level 5] Backend HTTP Delivery Layer (Chi router, JWT middleware, DTO envelopes)
    ↓
[Level 6] Frontend API Contracts & TypeScript Types (asesamo-web/lib/api/types.ts)
    ↓
[Level 7] Frontend State & Auth Hooks (useAuth, useCasetaCheckIn)
    ↓
[Level 8] Frontend UI Components (shadcn/ui, Caseta Check-in Screen, Status Cards)
    ↓
[Level 9] Container Orchestration & E2E Verification (docker compose up, verification script)
```

### 5.2 Cross-Module API Contracts

#### Standard Response Envelopes (Mandatory per DEVELOPMENT_RULES.md §6)

**Success Envelope:**
```json
{
  "data": { ... },
  "meta": { ... }
}
```

**Error Envelope:**
```json
{
  "error": {
    "code": "STRING_CONSTANT_IN_UPPER_SNAKE_CASE",
    "message": "Mensaje legible en español dirigido al usuario final."
  }
}
```

---

#### 5.2.1 Authentication Endpoint: `POST /api/v1/auth/login`

**Public Route (Rate limited)**

**Request Headers:**  
`Content-Type: application/json`

**Request Body:**
```json
{
  "email": "guard@asesamo.com",
  "password": "Password123!"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-10-01T23:15:51Z",
    "user": {
      "id": "44444444-4444-4444-4444-444444444444",
      "email": "guard@asesamo.com",
      "full_name": "Oficial Juan Gabriel Mendoza (Guardia Caseta)",
      "role": "guard",
      "tenant_id": "11111111-1111-1111-1111-111111111111"
    }
  }
}
```

**Error Responses:**
- `400 Bad Request`: `{"error": {"code": "INVALID_REQUEST", "message": "Datos de formulario inválidos."}}`
- `401 Unauthorized`: `{"error": {"code": "INVALID_CREDENTIALS", "message": "Correo o contraseña incorrectos."}}`

---

#### 5.2.2 Check-In Verification Endpoint: `POST /api/v1/access/check-in`

**Protected Route:** Requires `Authorization: Bearer <JWT>` where role is `guard` or `plant_admin`.

**Request Headers:**  
`Authorization: Bearer <jwt_token>`  
`Content-Type: application/json`

**Request Body:**
```json
{
  "worker_id": "66666666-6666-6666-6666-666666666661",
  "access_point_id": "33333333-3333-3333-3333-333333333333"
}
```

**Response (200 OK — Granted):**
```json
{
  "data": {
    "access_log_id": "77777777-7777-7777-7777-777777777777",
    "result": "granted",
    "denial_reason": null,
    "checked_in_at": "2026-10-01T22:30:00Z",
    "worker": {
      "id": "66666666-6666-6666-6666-666666666661",
      "full_name": "Carlos Mendoza Rodríguez",
      "job_title": "Técnico Soldador Especialista en Alturas",
      "curp": "MERC850101HDFR01",
      "imss_number": "12850123456",
      "avatar_r2_key": null
    },
    "contractor": {
      "id": "55555555-5555-5555-5555-555555555555",
      "name": "Mantenimiento Industrial y Soldadura Especializada S.A. de C.V.",
      "rfc": "MIS090814KK2",
      "repse_number": "AR-55420/2023"
    },
    "plant": {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "Planta Matriz Querétaro (Parque Industrial Benito Juárez)"
    },
    "access_point": {
      "id": "33333333-3333-3333-3333-333333333333",
      "name": "Caseta 1 - Acceso Peatonal Principal"
    },
    "document_snapshot": [
      {
        "document_type_code": "DC3",
        "name": "Constancia de Capacitación DC-3",
        "applies_to": "worker",
        "status": "approved",
        "expires_at": "2027-07-28",
        "is_valid": true,
        "rejection_reason": null
      },
      {
        "document_type_code": "INE",
        "name": "Identificación Oficial (INE/IFE)",
        "applies_to": "worker",
        "status": "approved",
        "expires_at": null,
        "is_valid": true,
        "rejection_reason": null
      },
      {
        "document_type_code": "MEDICAL",
        "name": "Certificado Médico de Aptitud",
        "applies_to": "worker",
        "status": "approved",
        "expires_at": "2027-03-30",
        "is_valid": true,
        "rejection_reason": null
      },
      {
        "document_type_code": "REPSE",
        "name": "Registro REPSE",
        "applies_to": "contractor",
        "status": "approved",
        "expires_at": "2027-08-02",
        "is_valid": true,
        "rejection_reason": null
      },
      {
        "document_type_code": "SUA",
        "name": "Comprobante de Pago SUA (IMSS/INFONAVIT)",
        "applies_to": "contractor",
        "status": "approved",
        "expires_at": "2026-10-26",
        "is_valid": true,
        "rejection_reason": null
      }
    ]
  }
}
```

**Response (200 OK — Denied):**
```json
{
  "data": {
    "access_log_id": "77777777-7777-7777-7777-888888888888",
    "result": "denied",
    "denial_reason": "Constancia DC-3 vencida el 2026-08-17 y Certificado Médico rechazado",
    "checked_in_at": "2026-10-01T22:31:00Z",
    "worker": {
      "id": "66666666-6666-6666-6666-666666666662",
      "full_name": "Roberto Morales Santos",
      "job_title": "Ayudante General de Mantenimiento",
      "curp": "MOSR920412HDFR02",
      "imss_number": "12920498765",
      "avatar_r2_key": null
    },
    "contractor": {
      "id": "55555555-5555-5555-5555-555555555555",
      "name": "Mantenimiento Industrial y Soldadura Especializada S.A. de C.V.",
      "rfc": "MIS090814KK2",
      "repse_number": "AR-55420/2023"
    },
    "plant": {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "Planta Matriz Querétaro (Parque Industrial Benito Juárez)"
    },
    "access_point": {
      "id": "33333333-3333-3333-3333-333333333333",
      "name": "Caseta 1 - Acceso Peatonal Principal"
    },
    "document_snapshot": [
      {
        "document_type_code": "DC3",
        "name": "Constancia de Capacitación DC-3",
        "applies_to": "worker",
        "status": "expired",
        "expires_at": "2026-08-17",
        "is_valid": false,
        "rejection_reason": null
      },
      {
        "document_type_code": "MEDICAL",
        "name": "Certificado Médico de Aptitud",
        "applies_to": "worker",
        "status": "rejected",
        "expires_at": "2027-09-21",
        "is_valid": false,
        "rejection_reason": "Certificado sin cédula profesional ni firma de médico dictaminador"
      }
    ]
  }
}
```

**Error Responses:**
- `401 Unauthorized`: `{"error": {"code": "UNAUTHORIZED", "message": "Token de autenticación inválido o ausente."}}`
- `403 Forbidden`: `{"error": {"code": "TENANT_MISMATCH", "message": "No tiene permiso para operar en este punto de acceso."}}`
- `404 Not Found`: `{"error": {"code": "WORKER_NOT_FOUND", "message": "El trabajador especificado no existe o no pertenece a una empresa contratista registrada."}}`

---

#### 5.2.3 Access Points Listing Endpoint: `GET /api/v1/access-points`

**Protected Route:** Required to populate the Caseta selector in the Next.js header.

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "33333333-3333-3333-3333-333333333333",
      "plant_id": "22222222-2222-2222-2222-222222222222",
      "plant_name": "Planta Matriz Querétaro (Parque Industrial Benito Juárez)",
      "name": "Caseta 1 - Acceso Peatonal Principal",
      "description": "Torniquete y módulo de control de contratistas",
      "is_active": true
    }
  ]
}
```

---

#### 5.2.4 System Health Endpoint: `GET /api/v1/health`

**Public Route:** Used by Docker Compose health checks and external monitoring.

**Response (200 OK):**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "database": "connected",
  "timestamp": "2026-10-01T22:30:00Z"
}
```

---

### 5.3 Frontend TypeScript Interface Contract (`asesamo-web/lib/api/types.ts`)

```typescript
export type UserRole =
  | 'super_admin'
  | 'plant_admin'
  | 'ehs_manager'
  | 'guard'
  | 'contractor_admin'
  | 'worker';

export type AccessResult = 'granted' | 'granted_with_warning' | 'denied';

export interface UserSession {
  id: string;
  email: string;
  full_name: string;
  role: UserRole;
  tenant_id: string;
}

export interface AuthLoginResponse {
  token: string;
  expires_at: string;
  user: UserSession;
}

export interface DocumentSnapshotItem {
  document_type_code: string;
  name: string;
  applies_to: 'worker' | 'contractor';
  status: 'pending_review' | 'approved' | 'rejected' | 'expired';
  expires_at: string | null;
  is_valid: boolean;
  rejection_reason?: string | null;
}

export interface CheckInResponse {
  access_log_id: string;
  result: AccessResult;
  denial_reason: string | null;
  checked_in_at: string;
  worker: {
    id: string;
    full_name: string;
    job_title: string;
    curp: string;
    imss_number: string;
    avatar_r2_key?: string | null;
  };
  contractor: {
    id: string;
    name: string;
    rfc: string;
    repse_number: string;
  };
  plant: {
    id: string;
    name: string;
  };
  access_point: {
    id: string;
    name: string;
  };
  document_snapshot: DocumentSnapshotItem[];
}

export interface AccessPointOption {
  id: string;
  plant_id: string;
  plant_name: string;
  name: string;
  description: string;
  is_active: boolean;
}

export interface ApiEnvelope<T> {
  data: T;
  meta?: Record<string, unknown>;
}

export interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
  };
}
```

---

## 6. Implementation Action Plan for Builders

With the survey complete and contracts established, the implementation team should proceed with the following actions:

1. **Root Scaffolding**:
   - Create root `docker-compose.yml`, `.env.example`, and `.env`.
   - Create `db/seeds/000001_seed_dev_data.sql`.
   - Create `scripts/verify_acceptance.sh`.
2. **Backend Construction (`asesamo-api`)**:
   - Scaffold Clean Architecture directory structure (`cmd/api`, `internal/domain`, `internal/usecase`, `internal/adapter`, `internal/infra`, `pkg`).
   - Configure `sqlc.yaml` and generate queries from `db/migrations/000001_init.up.sql`.
   - Implement `auth` (bcrypt password verification, JWT generator/validator) and `access` use cases (compliance evaluation engine).
   - Implement Chi HTTP handlers and router with CORS and tenant context injection.
   - Add unit tests for usecase verification (`checkin_usecase_test.go`) and integration test fixtures.
   - Add `asesamo-api/Dockerfile` and `Makefile`.
3. **Frontend Construction (`asesamo-web`)**:
   - Initialize Next.js 14 App Router project with TypeScript strict mode and Tailwind CSS.
   - Install and configure shadcn/ui components (`card`, `button`, `input`, `badge`, `table`, `sonner`).
   - Implement `lib/api/client.ts` encapsulating auth headers and error handling.
   - Implement `/login` page with Guard sign-in form.
   - Implement `/caseta` page with Access Point selection, worker UUID input (keyboard/scanner input), immediate visual status indicator (Green/Yellow/Red), worker details card, and document breakdown list.
   - Add `asesamo-web/Dockerfile`.
4. **Integration & End-to-End Verification**:
   - Run `docker compose up --build -d`.
   - Execute `scripts/verify_acceptance.sh`.
   - Verify browser interaction on `http://localhost:3000`.

---
*Report completed by teamwork_preview_explorer_survey_3.*
