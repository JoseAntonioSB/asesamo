# Project: ASESAMO Phase 1 MVP

## Architecture
ASESAMO is an industrial access control and document compliance system. Phase 1 MVP comprises an end-to-end stack:
1. **Database & Storage Layer (`db/`)**: PostgreSQL 16 with `pgcrypto` and `citext` extensions, schema managed via `000001_init.up.sql`, native bcrypt password hashing, and test seed data.
2. **Backend Service (`asesamo-api/`)**: Go REST API following Clean Architecture:
   - `cmd/api/main.go`: Entry point, dependency wiring.
   - `internal/core/domain/entity/`: Domain models (Tenant, Plant, AccessPoint, User, Contractor, Worker, Document, AccessLog).
   - `internal/core/domain/repository/`: Repository interfaces.
   - `internal/core/usecase/`: Pure business logic (Auth login, Access check-in/out, Access points).
   - `internal/adapter/http/`: Chi router v5, HTTP handlers, middlewares (JWT auth, tenant context, logging, CORS), request/response DTOs with `{ "data": ... }` and `{ "error": { "code", "message" } }` envelopes.
   - `internal/adapter/postgres/`: SQL queries and repository implementations.
   - `internal/infra/`: Configuration loader, database connection pool, logger.
3. **Frontend Web (`asesamo-web/`)**: Next.js 14+ App Router:
   - `app/(auth)/login/`: Security guard authentication screen.
   - `app/(dashboard)/caseta/`: Gate check-in interface with auto-focus scanner input, quick-fill demo presets, and access point selection.
   - `components/caseta/`: Visual status indicators (Green/Granted, Yellow/Warning, Red/Denied), Worker Profile Card, and Document Validation Breakdown Table.
   - `lib/api/`: Centralized HTTP client encapsulating base URL, JWT tokens, envelope unwrapping, and Spanish error handling.
4. **Local Orchestration & Environment**:
   - `docker-compose.yml`: Multi-container orchestration booting `db`, `migrate`, `seeder`, `api` (port 8080), and `web` (port 3000) with healthcheck dependencies.
   - `.env.example` and `.env`: Standardized environment configuration.
   - `scripts/verify_acceptance.sh`: Automated smoke script verifying all 7 acceptance criteria.

---

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | DB Schema & Migrations | PostgreSQL 16 schema with 13 tables, enums, triggers, pgcrypto, citext via 000001_init.up.sql | M1 | Survey 1 & 3 |
| 2 | Seed Data with Native Bcrypt | Seed script with test tenant, plant, caseta, guard credentials, contractor, and 3 test worker cases | M1 | Survey 3 |
| 3 | Clean Architecture Backend Scaffolding | Go module, directory structure, config loader, db connection pool, healthcheck endpoint | M2 | Survey 1 |
| 4 | Guard Authentication (Backend) | POST /api/v1/auth/login returning signed JWT with sub, tenant_id, role, exp claims | M2 | Survey 1 |
| 5 | Auth & Tenant Context Middlewares | JWT validation middleware and tenant.FromContext(ctx) injection with WHERE tenant_id = $1 filtering | M2 | Survey 1 |
| 6 | Access Points Catalog Endpoint | GET /api/v1/access-points returning active access points for guard's plant/tenant | M2 | Survey 1 |
| 7 | Document Compliance Evaluation Engine | Logic evaluating worker/contractor documents against plant policy (block_on_expired_docs) | M3 | Survey 1 |
| 8 | Access Point Check-In Verification | POST /api/v1/access/check-in returning granted, granted_with_warning, or denied | M3 | Survey 1 |
| 9 | Immutable Access Logging & JSONB Snapshot | Write check-in result, denial reason, and JSONB document_snapshot to access_logs table | M3 | Survey 1 |
| 10 | Access Check-Out & Logs Retrieval | POST /api/v1/access/check-out and GET /api/v1/access-logs endpoints | M3 | Survey 1 |
| 11 | Backend Unit & Integration Tests | Tests covering permitted, denied, warning conditions and tenant isolation with 0 failures | M3 | Survey 1 |
| 12 | Frontend Next.js 14 App Router Setup | TypeScript strict, Tailwind CSS, shadcn/ui component scaffolding, TanStack Query | M4 | Survey 2 |
| 13 | Frontend API Client Layer | lib/api encapsulating base URL, JWT cookies/headers, envelope unwrapping, error handling | M4 | Survey 2 |
| 14 | Security Guard Login Screen | Form with validation, login action, cookie JWT storage, redirect to /caseta, Spanish errors | M4 | Survey 2 |
| 15 | Caseta Check-In Interface | Auto-focused scanner input field, HID wedge support, manual UUID entry, quick-fill demo buttons | M4 | Survey 2 |
| 16 | 3-Tier Visual Status Feedback | High-visibility Green (Granted), Yellow (Warning), Red (Denied) banner cards | M4 | Survey 2 |
| 17 | Worker Profile & Document Breakdown Card | Display worker photo, name, CURP, NSS, contractor name, and breakdown table of documents | M4 | Survey 2 |
| 18 | Multi-Stage Dockerfiles | Dockerfile for asesamo-api (Go multi-stage) and Dockerfile for asesamo-web (Next.js standalone) | M5 | Survey 3 |
| 19 | Root Docker Compose Orchestration | Root docker-compose.yml wiring db, migrate, seeder, api, web with healthchecks and env vars | M5 | Survey 3 |
| 20 | Automated Acceptance Verification Script | scripts/verify_acceptance.sh asserting all 7 criteria from ORIGINAL_REQUEST.md | M5 | Survey 3 |
| 21 | Comprehensive E2E Test Suite | Opaque-box requirement-driven test suite (Tiers 1-4) covering all features and boundaries | E2E | Survey 1, 2, 3 |
| 22 | Final E2E Validation & Adversarial Hardening | Running 100% of E2E test suite + Tier 5 adversarial testing with challengers and auditors | M6 | Survey 1, 2, 3 |

---

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| E2E | E2E Testing Suite | Requirements-driven opaque-box test runner & Tiers 1-4 test cases -> TEST_READY.md | none | DONE |
| M1 | Database & Environment Scaffolding | .env.example, .env, docker-compose base, seed script with native bcrypt & deterministic test records | none | DONE |
| M2 | Backend Core & Authentication | asesamo-api Clean Architecture, config, DB pool, JWT auth, tenant middleware, access-points API, unit tests | M1 | DONE |
| M3 | Access Verification Engine & Logging | Check-in/out logic, compliance evaluation, immutable JSONB access_logs, unit + integration tests | M2 | DONE |
| M4 | Frontend Web Interface (asesamo-web) | Next.js 14, Tailwind, shadcn/ui, Guard login, Caseta check-in, visual status cards, worker & doc tables | M2 | PLANNED |
| M5 | Local Orchestration & Smoke Verification | Multi-stage Dockerfiles, full docker-compose.yml, scripts/verify_acceptance.sh | M3, M4 | PLANNED |
| M6 | Final E2E Test Pass & Hardening | Run 100% E2E test suite (Tiers 1-4) + Tier 5 adversarial coverage hardening | M5, E2E | PLANNED |

---

## Interface Contracts

### 1. HTTP Response Envelopes
All responses from `asesamo-api` adhere to:
```json
// Success
{
  "data": { ... },
  "meta": { "page": 1, "per_page": 20, "total": 1 } // optional
}

// Error
{
  "error": {
    "code": "DOCUMENT_EXPIRED",
    "message": "El documento Constancia DC-3 ha vencido."
  }
}
```

### 2. Authentication Contract
- Endpoint: `POST /api/v1/auth/login`
- Request:
```json
{
  "email": "guard@asesamo.com",
  "password": "Password123!"
}
```
- Response (200 OK):
```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "44444444-4444-4444-4444-444444444444",
      "tenant_id": "11111111-1111-1111-1111-111111111111",
      "email": "guard@asesamo.com",
      "full_name": "Oficial Roberto Caseta",
      "role": "guard"
    }
  }
}
```

### 3. Check-In Verification Contract
- Endpoint: `POST /api/v1/access/check-in`
- Headers: `Authorization: Bearer <JWT>`
- Request:
```json
{
  "worker_user_id": "66666666-6666-6666-6666-666666666661",
  "access_point_id": "33333333-3333-3333-3333-333333333331"
}
```
- Response (200 OK):
```json
{
  "data": {
    "access_log_id": "99999999-9999-9999-9999-999999999991",
    "result": "granted", // 'granted' | 'granted_with_warning' | 'denied'
    "denial_reason": null,
    "checked_in_at": "2026-10-01T22:20:00Z",
    "worker": {
      "user_id": "66666666-6666-6666-6666-666666666661",
      "full_name": "Carlos Mendoza",
      "curp": "MENC850101HDFR01",
      "imss_number": "12345678901",
      "job_title": "Soldador Industrial",
      "avatar_r2_key": null
    },
    "contractor": {
      "id": "55555555-5555-5555-5555-555555555555",
      "name": "Mantenimiento Industrial del Norte SA de CV",
      "rfc": "MIN120304AB1",
      "repse_number": "REPSE-NL-2023-8841",
      "status": "active"
    },
    "plant": {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "Planta Querétaro Norte"
    },
    "access_point": {
      "id": "33333333-3333-3333-3333-333333333331",
      "name": "Caseta Principal Norte"
    },
    "document_snapshot": [
      {
        "document_type_code": "DC3",
        "document_name": "Constancia de Capacitación DC-3",
        "applies_to": "worker",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2027-08-01",
        "failure_reason": null
      },
      {
        "document_type_code": "INE",
        "document_name": "Identificación Oficial (INE/IFE)",
        "applies_to": "worker",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": null,
        "failure_reason": null
      },
      {
        "document_type_code": "REPSE",
        "document_name": "Registro REPSE",
        "applies_to": "contractor",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2027-04-15",
        "failure_reason": null
      }
    ]
  }
}
```

---

## Code Layout
```
/Users/jasbotello/devel/asesamo/
├── docker-compose.yml
├── .env.example
├── .env
├── Makefile
├── db/
│   ├── migrations/
│   │   ├── 000001_init.up.sql
│   │   └── 000001_init.down.sql
│   └── seeds/
│       └── 000001_seed_dev_data.sql
├── scripts/
│   ├── verify_acceptance.sh
│   └── run_e2e_tests.sh
├── test/
│   └── e2e/
│       ├── test_runner.sh
│       ├── tier1_feature_test.sh
│       ├── tier2_boundary_test.sh
│       ├── tier3_combination_test.sh
│       └── tier4_workload_test.sh
├── asesamo-api/
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── core/
│   │   │   ├── domain/
│   │   │   │   ├── entity/
│   │   │   │   └── repository/
│   │   │   └── usecase/
│   │   │       ├── auth/
│   │   │       └── access/
│   │   ├── adapter/
│   │   │   ├── http/
│   │   │   │   ├── handler/
│   │   │   │   ├── middleware/
│   │   │   │   ├── dto/
│   │   │   │   └── router.go
│   │   │   └── postgres/
│   │   └── infra/
│   │       ├── config/
│   │       └── database/
│   ├── pkg/
│   │   ├── jwt/
│   │   └── tenant/
│   ├── Dockerfile
│   └── go.mod
└── asesamo-web/
    ├── app/
    │   ├── layout.tsx
    │   ├── page.tsx
    │   ├── (auth)/login/page.tsx
    │   └── (dashboard)/caseta/page.tsx
    ├── components/
    │   ├── ui/
    │   └── caseta/
    │       ├── StatusBanner.tsx
    │       ├── WorkerProfileCard.tsx
    │       ├── DocumentTable.tsx
    │       └── QuickFillPresets.tsx
    ├── lib/
    │   ├── api/
    │   │   ├── client.ts
    │   │   └── types.ts
    │   └── utils.ts
    ├── Dockerfile
    ├── package.json
    ├── tailwind.config.ts
    └── tsconfig.json
```
