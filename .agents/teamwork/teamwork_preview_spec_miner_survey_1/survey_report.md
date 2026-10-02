# ASESAMO — Phase 1 MVP Backend Specification Survey Report

**Author:** `teamwork_preview_spec_miner_survey_1`  
**Date:** 2026-10-01  
**Scope:** Backend Architecture, REST API, Database Mappings, Access Verification Engine, Multi-Tenancy, Authentication, and Testing for `asesamo-api`.  
**Authoritative Sources:**
- `ORIGINAL_REQUEST.md` (and `.agents/teamwork/ORIGINAL_REQUEST.md`)
- `DEVELOPMENT_RULES.md`
- `db/migrations/000001_init.up.sql`
- `db/migrations/000001_init.down.sql`
- `SAS Seguridad e Higiene asesamo.md`

---

## 1. Executive Summary

ASESAMO is an industrial safety and access compliance SaaS tailored to the Mexican regulatory landscape (STPS, REPSE, IMSS, NOMs). 
The Phase 1 MVP focuses on an end-to-end access control verification pipeline:
1. Security personnel (Guards) authenticate via email and password to receive a scoped JWT.
2. A Guard at a plant gate/caseta submits a worker UUID (scanned from a QR badge or entered manually) and the access point ID.
3. The backend validates worker identity, active contractor relationship to the plant, and document compliance against the plant's policy (DC-3, REPSE, SUA, SAT 32-D, IMSS, Medical certificate, INE).
4. The system produces an immutable audit record in `access_logs` with a JSONB document snapshot.
5. The system returns an immediate status (`granted`, `granted_with_warning`, or `denied`) with detailed Mexican labor compliance rationale.

---

## 2. Features Discovered

| # | Category | Feature | Description | Inputs | Outputs | Error Behavior | Discovered Via |
|---|----------|---------|-------------|--------|---------|----------------|----------------|
| 1 | Auth | Guard Login (`POST /api/v1/auth/login`) | Authenticates a user (guard/admin) via email and password, verifies active status, updates `last_login_at`, and returns a 1-hour JWT access token plus 7-day refresh token cookie. | `{ "email": string, "password": string }` | `{ "data": { "access_token": string, "token_type": "Bearer", "expires_in": 3600, "user": { ... } } }` | 400 Bad Request if format invalid; 401 Unauthorized (`INVALID_CREDENTIALS`, `USER_INACTIVE`); 429 Too Many Requests | `ORIGINAL_REQUEST.md` R1; `DEVELOPMENT_RULES.md` §6, §7; `000001_init.up.sql` |
| 2 | Auth | Token Refresh (`POST /api/v1/auth/refresh`) | Validates refresh token cookie, rotates hash in `refresh_tokens` table, and returns a new 1-hour JWT access token. | HTTP cookie `refresh_token` | `{ "data": { "access_token": string, "expires_in": 3600 } }` | 401 Unauthorized (`INVALID_REFRESH_TOKEN`, `TOKEN_REVOKED`, `TOKEN_EXPIRED`) | `DEVELOPMENT_RULES.md` §7; `000001_init.up.sql` table `refresh_tokens` |
| 3 | Auth | Logout (`POST /api/v1/auth/logout`) | Revokes server-side refresh token in `refresh_tokens` and clears cookie. | Authenticated context + refresh cookie | `{ "data": { "message": "Sesión cerrada correctamente" } }` | 401 Unauthorized | `DEVELOPMENT_RULES.md` §7; `000001_init.up.sql` |
| 4 | Access Control | Gate Check-In (`POST /api/v1/access/check-in`) | Verifies worker identity, contractor affiliation, plant-contractor authorization, and document compliance against plant policy. Creates an immutable `access_logs` record with JSONB snapshot. | `{ "worker_id": UUID, "access_point_id": UUID, "notes": string (optional) }` | `{ "data": { "access_log_id": UUID, "result": "granted"\|"granted_with_warning"\|"denied", "worker": { ... }, "plant": { ... }, "access_point": { ... }, "checked_in_at": timestamp, "denial_reason": string\|null, "document_summary": { ... }, "document_snapshot": [ ... ] } }` | 400/404 if worker/point missing; 422 if worker inactive; 200 with `result: denied` if compliance fails | `ORIGINAL_REQUEST.md` R1; `DEVELOPMENT_RULES.md` §2, §3, §6; `000001_init.up.sql` |
| 5 | Access Control | Gate Check-Out (`POST /api/v1/access/check-out`) | Records exit timestamp for a worker currently inside the plant (`checked_out_at = now()`), updating the open `access_logs` entry. | `{ "worker_id": UUID, "plant_id": UUID }` or `{ "access_log_id": UUID }` | `{ "data": { "access_log_id": UUID, "checked_out_at": timestamp } }` | 404 (`ACTIVE_ENTRY_NOT_FOUND`) if worker is not logged as currently inside | `000001_init.up.sql` index `idx_access_logs_active_entries` |
| 6 | System | Health Check (`GET /health` or `GET /api/v1/health`) | Verifies service liveness, uptime, and database connectivity. | None | `{ "data": { "status": "ok", "database": "connected", "version": string } }` | 503 Service Unavailable if database is unreachable | Standard cloud architecture; `ORIGINAL_REQUEST.md` Acceptance Criteria |
| 7 | Caseta Operations | List Active Access Points (`GET /api/v1/access-points`) | Returns active access points for the authenticated guard's tenant/plant to configure the caseta client. | Optional query `?plant_id=UUID` | `{ "data": [ { "id": UUID, "name": string, "description": string, "plant_id": UUID } ] }` | 401 Unauthorized, 403 Forbidden | `000001_init.up.sql` table `access_points`; `ORIGINAL_REQUEST.md` R2 |
| 8 | Caseta Operations | Worker Compliance Preview (`GET /api/v1/workers/{id}/compliance`) | Inspects a worker's current compliance and document status for a plant without creating an access log entry. | Path parameter `id` (worker UUID), query `?plant_id=UUID` | `{ "data": { "worker": { ... }, "contractor": { ... }, "compliance_status": "valid"\|"warning"\|"expired"\|"missing", "documents": [ ... ] } }` | 404 (`WORKER_NOT_FOUND`), 403 (`FORBIDDEN`) | `ORIGINAL_REQUEST.md` R2; `000001_init.up.sql` |
| 9 | Audit & History | List Recent Access Logs (`GET /api/v1/access-logs`) | Paginated audit trail of gate entries and exits for a plant with immutable snapshots. | Query `?plant_id=UUID&page=1&per_page=20` | `{ "data": [ ... ], "meta": { "page": 1, "per_page": 20, "total": 100 } }` | 400 Bad Request, 401 Unauthorized | `DEVELOPMENT_RULES.md` §6; `000001_init.up.sql` table `access_logs` |
| 10 | Notifications | Access Denied Alerting | Triggers an internal system notification when an access attempt is denied at the gate. | Auto-triggered by `check_in` use case on denial | Notification row in `notifications` table (`type = 'access_denied'`) | Logged warning if notification insert fails; check-in remains denied | `000001_init.up.sql` table `notifications` |
| 11 | Compliance Engine | Plant Policy Evaluation | Dynamically resolves document requirements by combining global system types (`document_types`) and plant overrides (`plant_document_requirements`). | `plant_id`, `worker_id`, `contractor_id` | Set of evaluated document rules, validity windows, and blocking flags (`block_on_expired_docs`) | Fallback to default system catalog if no plant override exists | `000001_init.up.sql` tables `document_types`, `plant_document_requirements` |
| 12 | Tenant Isolation | Tenant Context Injection | Middleware extracts `tenant_id` from JWT claims and injects it into Go `context.Context` via `tenant.FromContext(ctx)`. All queries enforce `WHERE tenant_id = $1`. | Request header `Authorization: Bearer <token>` | Populated `context.Context` | 401 Unauthorized if missing/invalid; 403 Forbidden if tenant context mismatch | `DEVELOPMENT_RULES.md` §3 |

---

## 3. Edge Cases

| # | Feature | Input | Observed Behavior |
|---|---------|-------|-------------------|
| 1 | Check-In | Non-existent Worker UUID | Repository returns no rows. Use case halts before document query; returns 404 Not Found with error envelope `WORKER_NOT_FOUND` ("Trabajador no encontrado en el sistema"). No `access_logs` row created (cannot link foreign key `user_id`). |
| 2 | Check-In | Worker `is_active = FALSE` or `deleted_at IS NOT NULL` | Worker exists but is deactivated. Check-in records an `access_logs` entry with `result: denied`, `denial_reason: "El trabajador se encuentra inactivo en el sistema"`. Response returns 200/422 with denied status. |
| 3 | Check-In | Worker role is not `worker` (e.g. `guard` or `plant_admin` UUID submitted) | Use case rejects check-in; returns 422 Unprocessable Entity `INVALID_WORKER_ROLE` ("El usuario especificado no cuenta con rol de trabajador de contratista"). |
| 4 | Check-In | Worker has no active contractor assignment (`contractor_workers` missing or inactive) | Worker exists but is unassigned or contractor affiliation is severed. Check-in produces `result: denied`, `denial_reason: "El trabajador no tiene una empresa contratista activa asignada"`. Snapshot is empty. |
| 5 | Check-In | Contractor exists globally but has no relationship with the plant in `plant_contractors` | Contractor is not registered for this plant. Check-in denied: `denial_reason: "La empresa contratista no está vinculada a esta planta industrial"`. |
| 6 | Check-In | Contractor relationship status is `pending`, `suspended`, or `blocked` | Check-in denied immediately: `denial_reason: "El contratista se encuentra [suspendido/bloqueado/pendiente] en esta planta"`. Documents are not evaluated because contractor status is a hard blocker. |
| 7 | Check-In | Plant has `block_on_expired_docs = FALSE` and worker has an expired DC-3 | System detects expiration. Because `block_on_expired_docs = FALSE` and no mandatory documents are missing or rejected, `result: granted_with_warning`. Denial reason is `null`, warning details recorded in snapshot and notes. |
| 8 | Check-In | Plant has `block_on_expired_docs = TRUE` and worker has an expired DC-3 | System detects expiration. Because `block_on_expired_docs = TRUE`, `result: denied`. `denial_reason: "Documento Constancia DC-3 vencido el YYYY-MM-DD"`. |
| 9 | Check-In | Worker documents are valid, but Contractor's REPSE or SUA is expired | Access check evaluates both worker-level and contractor-level documents. If contractor's REPSE is expired and plant blocks expired docs, `result: denied`, `denial_reason: "Documento de la empresa contratista (Registro REPSE) vencido el YYYY-MM-DD"`. |
| 10 | Check-In | Document is uploaded but has `status = 'pending_review'` or `'rejected'` | A pending or rejected document cannot satisfy a mandatory requirement. Evaluated as invalid: `result: denied`, `denial_reason: "Documento pendiente de aprobación por Seguridad e Higiene (EHS)"`. |
| 11 | Check-In | Guard belongs to Tenant A, but tries to access Access Point or Plant belonging to Tenant B | Tenant isolation middleware and query `WHERE tenant_id = $1` filters out the access point. Query returns no rows. System responds with 404 `ACCESS_POINT_NOT_FOUND` ("Punto de acceso no encontrado"). Cross-tenant leakage is strictly prevented. |
| 12 | Check-In | Worker scans QR code immediately after already checking in (double check-in without check-out) | Check-in succeeds but logs entry; optional warning note "El trabajador ya contaba con un ingreso activo sin registro de salida". Access log is still recorded for safety and audit compliance. |
| 13 | Login | Guard user has `is_active = FALSE` | Login verification detects inactive user; returns 401 Unauthorized `USER_INACTIVE` ("La cuenta de usuario está desactivada"). |
| 14 | Login | Guard submits incorrect password | `bcrypt.CompareHashAndPassword` returns `ErrMismatchedHashAndPassword`. Returns 401 Unauthorized `INVALID_CREDENTIALS` ("Correo o contraseña incorrectos"). No distinguishing between wrong email or wrong password for security. |

---

## 4. Clean Architecture Layer Structure (`asesamo-api`)

Following `DEVELOPMENT_RULES.md` §2:

```
asesamo-api/
├── cmd/
│   └── api/
│       └── main.go                       # Application entry point, dependency injection & graceful shutdown
├── internal/
│   ├── core/                             # Core business domain and use cases
│   │   ├── domain/                       # Pure enterprise entities and repository interfaces
│   │   │   ├── entity/
│   │   │   │   ├── access_log.go         # AccessLog, DocumentSnapshot, AccessResult
│   │   │   │   ├── access_point.go       # AccessPoint
│   │   │   │   ├── contractor.go         # Contractor, ContractorWorker, PlantContractor
│   │   │   │   ├── document.go           # Document, DocumentType, PlantDocRequirement
│   │   │   │   ├── plant.go              # Plant
│   │   │   │   ├── tenant.go             # Tenant
│   │   │   │   └── user.go               # User, UserRole, RefreshToken
│   │   │   └── repository/               # Repositories contracts consumed by use cases
│   │   │       ├── access_log_repo.go
│   │   │       ├── access_point_repo.go
│   │   │       ├── contractor_repo.go
│   │   │       ├── document_repo.go
│   │   │       ├── plant_repo.go
│   │   │       └── user_repo.go
│   │   └── usecase/                      # Pure application business logic
│   │       ├── access/
│   │       │   ├── check_in_usecase.go   # Core access verification engine
│   │       │   ├── check_in_usecase_test.go
│   │       │   ├── check_out_usecase.go
│   │       │   └── access_dto.go
│   │       └── auth/
│   │           ├── login_usecase.go      # Authentication & token generation
│   │           ├── login_usecase_test.go
│   │           └── auth_dto.go
│   ├── adapter/
│   │   ├── http/                         # Chi HTTP adapters, handlers, middleware & router
│   │   │   ├── handler/
│   │   │   │   ├── access_handler.go     # POST /api/v1/access/check-in, check-out
│   │   │   │   ├── auth_handler.go       # POST /api/v1/auth/login, refresh, logout
│   │   │   │   └── health_handler.go     # GET /health
│   │   │   ├── middleware/
│   │   │   │   ├── auth_jwt.go           # JWT token verification and context extraction
│   │   │   │   ├── cors.go               # CORS headers
│   │   │   │   ├── logger.go             # Structured HTTP request logger
│   │   │   │   ├── rate_limit.go         # Rate limiter for auth routes
│   │   │   │   ├── role_guard.go         # RBAC guard (e.g. guard, plant_admin)
│   │   │   │   ├── security_headers.go   # OWASP recommended security headers
│   │   │   │   └── tenant.go             # Tenant isolation context validator
│   │   │   └── router.go                 # Chi route registration and grouping
│   │   └── postgres/                     # Repository implementations using sqlc
│   │       ├── access_log_repository.go
│   │       ├── access_point_repository.go
│   │       ├── contractor_repository.go
│   │       ├── document_repository.go
│   │       ├── plant_repository.go
│   │       ├── user_repository.go
│   │       └── sqlc/                     # Generated Go code from sqlc
│   │           ├── db.go
│   │           ├── models.go
│   │           └── querier.go
│   └── infra/
│       ├── config/                       # Env vars loading and validation
│       │   └── config.go
│       ├── database/                     # PostgreSQL connection pool (pgxpool)
│       │   └── postgres.go
│       └── logger/                       # Slog / structured logger wrapper
│           └── logger.go
├── db/
│   ├── migrations/                       # golang-migrate SQL migrations
│   │   ├── 000001_init.down.sql
│   │   └── 000001_init.up.sql
│   ├── queries/                          # Pure SQL queries for sqlc
│   │   ├── access.sql
│   │   ├── auth.sql
│   │   ├── contractors.sql
│   │   ├── documents.sql
│   │   └── plants.sql
│   └── sqlc.yaml                         # sqlc code generation configuration
├── pkg/
│   ├── jwt/                              # JWT generation and validation helper
│   │   └── jwt.go
│   ├── response/                         # Standardized JSON response envelope
│   │   └── response.go
│   ├── tenant/                           # Context tenant helpers (FromContext, WithContext)
│   │   └── context.go
│   └── validator/                        # go-playground/validator wrapper
│       └── validator.go
├── Dockerfile                            # Multi-stage production and dev Dockerfile
├── docker-compose.yml                    # Local multi-service orchestration
├── go.mod
├── go.sum
├── Makefile                              # Standard build, test, migration commands
└── .env.example                          # Environment variable specifications
```

### Strict Architectural Boundaries
1. **Domain Layer (`internal/core/domain`)**: Contains only pure Go structs and interfaces. Does NOT import Chi, SQL, sqlc, or any infrastructure package.
2. **Use Case Layer (`internal/core/usecase`)**: Contains all business rules. Depends ONLY on Domain interfaces. Does NOT know about HTTP, Chi, JSON, or SQL.
3. **HTTP Adapter Layer (`internal/adapter/http`)**: Translates HTTP requests to Use Case input DTOs, executes use cases, and maps results to standard JSON envelopes. Never executes database queries directly.
4. **Postgres Adapter Layer (`internal/adapter/postgres`)**: Implements Domain repository interfaces using `sqlc`-generated queries. Converts database models into Domain entities. Never imports use cases or HTTP handlers.

---

## 5. API Endpoints Specification

### 5.1 Envelopes and Error Format

All responses must adhere strictly to `DEVELOPMENT_RULES.md` §6:

#### Success Envelope
```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 42
  }
}
```
*(The `meta` object is included on list/paginated endpoints and omitted when not applicable).*

#### Error Envelope
```json
{
  "error": {
    "code": "ERROR_CODE_IN_UPPER_SNAKE_CASE",
    "message": "Mensaje de error claro y amigable en español."
  }
}
```

### 5.2 Error Codes and Spanish Messages Catalog

| HTTP Status | Error Code (`code`) | User Message (`message`) |
|---|---|---|
| 400 Bad Request | `INVALID_INPUT` | "Los datos proporcionados son inválidos o están incompletos." |
| 400 Bad Request | `INVALID_UUID` | "El identificador proporcionado no tiene un formato UUID válido." |
| 401 Unauthorized | `UNAUTHORIZED` | "No se proporcionó un token de autorización válido." |
| 401 Unauthorized | `INVALID_CREDENTIALS` | "Correo electrónico o contraseña incorrectos." |
| 401 Unauthorized | `TOKEN_EXPIRED` | "La sesión ha expirado. Por favor inicie sesión nuevamente." |
| 401 Unauthorized | `TOKEN_INVALID` | "El token de autenticación no es válido." |
| 401 Unauthorized | `USER_INACTIVE` | "La cuenta de usuario se encuentra desactivada." |
| 403 Forbidden | `FORBIDDEN` | "No tiene permisos suficientes para realizar esta acción." |
| 403 Forbidden | `TENANT_MISMATCH` | "No tiene autorización para acceder a los recursos de esta planta." |
| 404 Not Found | `WORKER_NOT_FOUND` | "Trabajador no encontrado en el sistema." |
| 404 Not Found | `ACCESS_POINT_NOT_FOUND` | "Punto de acceso no encontrado o inactivo." |
| 404 Not Found | `PLANT_NOT_FOUND` | "Planta industrial no encontrada." |
| 404 Not Found | `CONTRACTOR_NOT_FOUND` | "Empresa contratista no encontrada." |
| 422 Unprocessable Entity | `INVALID_WORKER_ROLE` | "El identificador no corresponde a un trabajador de contratista." |
| 422 Unprocessable Entity | `CONTRACTOR_NOT_LINKED` | "El contratista no cuenta con autorización activa en esta planta." |
| 422 Unprocessable Entity | `NO_ACTIVE_ENTRY` | "No se encontró un registro de ingreso activo para registrar la salida." |
| 429 Too Many Requests | `RATE_LIMIT_EXCEEDED` | "Demasiados intentos. Por favor espere un momento antes de reintentar." |
| 500 Internal Server Error | `INTERNAL_SERVER_ERROR` | "Ocurrió un error interno en el servidor. Por favor intente más tarde." |

---

### 5.3 Detailed Endpoint Contracts

#### 1. `POST /api/v1/auth/login`
- **Purpose:** Authenticate security guard and plant personnel.
- **Authentication:** Public (Rate limited via `go-chi/httprate`: 5 requests/min per IP).
- **Request Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "email": "guardia@plantanorte.com",
    "password": "Password123!"
  }
  ```
- **Validation Rules:**
  - `email`: required, valid email format (`citeext` in DB).
  - `password`: required, min 8 characters.
- **Response 200 OK:**
  - Sets HTTP-only cookie:
    `Set-Cookie: refresh_token=<random_token>; Path=/api/v1/auth; HttpOnly; Secure; SameSite=Strict; Max-Age=604800`
  - Body:
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 3600,
        "user": {
          "id": "e7b0e118-2436-4d1a-85d8-c92e76fae639",
          "email": "guardia@plantanorte.com",
          "full_name": "Mateo Silva",
          "role": "guard",
          "tenant_id": "a1b2c3d4-0000-0000-0000-000000000001",
          "last_login_at": "2026-10-01T22:30:00Z"
        }
      }
    }
    ```

#### 2. `POST /api/v1/access/check-in`
- **Purpose:** Verify worker & contractor credentials, validate compliance against plant policy, record immutable access log, and return entry verdict.
- **Authentication:** Protected (`Authorization: Bearer <JWT>`), Roles: `guard`, `plant_admin`, `super_admin`.
- **Request Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "worker_id": "8f3b772c-d238-4ea7-9a64-7bf3e9c40211",
    "access_point_id": "3c914bf1-a4b5-4b02-861f-132d439b1a01",
    "notes": "Ingreso para mantenimiento eléctrico en subestación 2"
  }
  ```
- **Response 200 OK (Granted):**
  ```json
  {
    "data": {
      "access_log_id": "9d8c7b6a-5432-10fe-ba98-76543210fedc",
      "result": "granted",
      "worker": {
        "id": "8f3b772c-d238-4ea7-9a64-7bf3e9c40211",
        "full_name": "Juan Carlos Pérez Gómez",
        "curp": "PEGJ850612HDFRRN09",
        "imss_number": "12345678901",
        "job_title": "Técnico Electricista",
        "avatar_url": "https://cdn.asesamo.com/presigned/worker_avatar.jpg"
      },
      "contractor": {
        "id": "11223344-5566-7788-99aa-bbccddeeff00",
        "name": "Electrificaciones del Bajío SA de CV",
        "rfc": "EBJ180415KZ8",
        "repse_number": "AR12903/2021",
        "contractor_type": "company",
        "plant_relationship_status": "active"
      },
      "plant": {
        "id": "22334455-6677-8899-aabb-ccddeeff0011",
        "name": "Planta Querétaro Norte",
        "block_on_expired_docs": true
      },
      "access_point": {
        "id": "3c914bf1-a4b5-4b02-861f-132d439b1a01",
        "name": "Caseta Principal"
      },
      "processed_by": {
        "id": "e7b0e118-2436-4d1a-85d8-c92e76fae639",
        "full_name": "Mateo Silva"
      },
      "checked_in_at": "2026-10-01T22:35:10Z",
      "denial_reason": null,
      "document_summary": {
        "total_required": 5,
        "valid": 5,
        "expired": 0,
        "missing": 0,
        "pending_review": 0
      },
      "document_snapshot": [
        {
          "document_type_code": "DC3",
          "document_type_name": "Constancia de Capacitación DC-3",
          "applies_to": "worker",
          "status": "approved",
          "is_valid": true,
          "expires_at": "2027-04-15",
          "reason": null
        },
        {
          "document_type_code": "MEDICAL",
          "document_type_name": "Certificado Médico de Aptitud",
          "applies_to": "worker",
          "status": "approved",
          "is_valid": true,
          "expires_at": "2027-02-10",
          "reason": null
        },
        {
          "document_type_code": "INE",
          "document_type_name": "Identificación Oficial (INE/IFE)",
          "applies_to": "worker",
          "status": "approved",
          "is_valid": true,
          "expires_at": null,
          "reason": null
        },
        {
          "document_type_code": "REPSE",
          "document_type_name": "Registro REPSE",
          "applies_to": "contractor",
          "status": "approved",
          "is_valid": true,
          "expires_at": "2027-08-30",
          "reason": null
        },
        {
          "document_type_code": "SUA",
          "document_type_name": "Comprobante de Pago SUA (IMSS/INFONAVIT)",
          "applies_to": "contractor",
          "status": "approved",
          "is_valid": true,
          "expires_at": "2026-10-20",
          "reason": null
        }
      ]
    }
  }
  ```

- **Response 200 OK (Denied Example):**
  ```json
  {
    "data": {
      "access_log_id": "11223344-5566-7788-99aa-bbccddeeff11",
      "result": "denied",
      "worker": {
        "id": "8f3b772c-d238-4ea7-9a64-7bf3e9c40211",
        "full_name": "Juan Carlos Pérez Gómez",
        "curp": "PEGJ850612HDFRRN09",
        "imss_number": "12345678901",
        "job_title": "Técnico Electricista"
      },
      "contractor": {
        "id": "11223344-5566-7788-99aa-bbccddeeff00",
        "name": "Electrificaciones del Bajío SA de CV",
        "rfc": "EBJ180415KZ8",
        "repse_number": "AR12903/2021",
        "contractor_type": "company",
        "plant_relationship_status": "active"
      },
      "plant": {
        "id": "22334455-6677-8899-aabb-ccddeeff0011",
        "name": "Planta Querétaro Norte",
        "block_on_expired_docs": true
      },
      "access_point": {
        "id": "3c914bf1-a4b5-4b02-861f-132d439b1a01",
        "name": "Caseta Principal"
      },
      "checked_in_at": "2026-10-01T22:36:00Z",
      "denial_reason": "Acceso denegado: Constancia DC-3 vencida el 2026-09-15; Comprobante SUA no presentado.",
      "document_summary": {
        "total_required": 5,
        "valid": 3,
        "expired": 1,
        "missing": 1,
        "pending_review": 0
      },
      "document_snapshot": [
        {
          "document_type_code": "DC3",
          "document_type_name": "Constancia de Capacitación DC-3",
          "applies_to": "worker",
          "status": "expired",
          "is_valid": false,
          "expires_at": "2026-09-15",
          "reason": "Documento vencido el 2026-09-15"
        },
        {
          "document_type_code": "SUA",
          "document_type_name": "Comprobante de Pago SUA (IMSS/INFONAVIT)",
          "applies_to": "contractor",
          "status": "missing",
          "is_valid": false,
          "expires_at": null,
          "reason": "Documento no registrado en el sistema"
        }
      ]
    }
  }
  ```

#### 3. `POST /api/v1/access/check-out`
- **Purpose:** Close open access entry when worker leaves the facility.
- **Authentication:** Protected (`guard`, `plant_admin`).
- **Request Body:**
  ```json
  {
    "worker_id": "8f3b772c-d238-4ea7-9a64-7bf3e9c40211",
    "plant_id": "22334455-6677-8899-aabb-ccddeeff0011"
  }
  ```
- **Response 200 OK:**
  ```json
  {
    "data": {
      "access_log_id": "9d8c7b6a-5432-10fe-ba98-76543210fedc",
      "worker_id": "8f3b772c-d238-4ea7-9a64-7bf3e9c40211",
      "checked_out_at": "2026-10-01T23:45:00Z",
      "message": "Salida registrada exitosamente."
    }
  }
  ```

#### 4. `GET /api/v1/access-points`
- **Purpose:** Provide list of access points in current plant/tenant for the Guard UI selector.
- **Authentication:** Protected (`guard`, `plant_admin`).
- **Response 200 OK:**
  ```json
  {
    "data": [
      {
        "id": "3c914bf1-a4b5-4b02-861f-132d439b1a01",
        "plant_id": "22334455-6677-8899-aabb-ccddeeff0011",
        "name": "Caseta Principal",
        "description": "Entrada peatonal contratistas",
        "is_active": true
      }
    ]
  }
  ```

#### 5. `GET /api/v1/access-logs`
- **Purpose:** Historical audit feed of entries and exits.
- **Authentication:** Protected.
- **Query Params:** `?plant_id=UUID&page=1&per_page=20`
- **Response 200 OK:**
  ```json
  {
    "data": [ ... ],
    "meta": {
      "page": 1,
      "per_page": 20,
      "total": 128
    }
  }
  ```

#### 6. `GET /health`
- **Purpose:** Infrastructure liveness probe.
- **Authentication:** Public.
- **Response 200 OK:**
  ```json
  {
    "data": {
      "status": "ok",
      "database": "connected",
      "version": "1.0.0"
    }
  }
  ```

---

## 6. Access Verification Engine & Business Logic Specification

The access point check-in verification is the core compliance engine of ASESAMO. It executes the following deterministic sequential decision tree:

```
[Request: worker_id, access_point_id]
                  │
                  ▼
   1. Access Point & Plant Resolution
   - Lookup Access Point WHERE id = access_point_id AND tenant_id = ctx.tenant_id
   - Check is_active == true
   - Lookup Plant WHERE id = access_point.plant_id AND tenant_id = ctx.tenant_id
   - Extract plant.block_on_expired_docs
                  │ (Valid)
                  ▼
   2. Worker Identity Check
   - Lookup User WHERE id = worker_id AND role = 'worker' AND deleted_at IS NULL
   - If user not found -> 404 WORKER_NOT_FOUND
   - If user.is_active == false -> DENIED ("Trabajador inactivo en el sistema")
                  │ (Active)
                  ▼
   3. Contractor & Plant Authorization Check
   - Lookup contractor_workers WHERE user_id = worker.id AND is_active = true
   - If not found -> DENIED ("El trabajador no está asignado a ninguna empresa contratista activa")
   - Lookup contractor in contractors WHERE id = contractor_workers.contractor_id
   - If contractor.is_active == false -> DENIED ("La empresa contratista está inactiva")
   - Lookup plant_contractors WHERE plant_id = plant.id AND contractor_id = contractor.id AND tenant_id = ctx.tenant_id
   - Check plant_contractors.status:
       * 'pending'   -> DENIED ("Relación contratista-planta pendiente de aprobación")
       * 'suspended' -> DENIED ("Contratista suspendido en esta planta")
       * 'blocked'   -> DENIED ("Contratista bloqueado en esta planta")
       * 'active'    -> PASSED
                  │ (Active)
                  ▼
   4. Document Requirement Resolution
   - Fetch effective required documents for plant:
     SELECT document_types joined with plant_document_requirements
     WHERE is_required = true
   - Split into worker-level rules (applies_to = 'worker') and contractor-level rules (applies_to = 'contractor')
                  │
                  ▼
   5. Document Evaluation against Policy
   - For each requirement:
     * Query latest document in documents WHERE owner_id = (worker.id OR contractor.id) AND document_type_id = req.id
     * If no document exists -> status = 'missing', is_valid = false
     * If document exists:
         - If doc.status != 'approved' (i.e. 'pending_review' or 'rejected') -> is_valid = false
         - If doc.expires_at IS NOT NULL AND doc.expires_at < current_date -> status = 'expired', is_valid = false
         - If approved and not expired -> is_valid = true
                  │
                  ▼
   6. Decision Aggregation
   - Any missing mandatory documents?
       -> Result = DENIED ("Documentos requeridos faltantes: ...")
   - Any rejected or pending documents?
       -> Result = DENIED ("Documentos sin aprobación vigente: ...")
   - Any expired documents?
       - If plant.block_on_expired_docs == TRUE:
           -> Result = DENIED ("Documentos vencidos: ...")
       - If plant.block_on_expired_docs == FALSE:
           -> Result = GRANTED_WITH_WARNING ("Acceso permitido con advertencia: documentos vencidos")
   - All documents valid?
       -> Result = GRANTED
                  │
                  ▼
   7. Immutable Log Creation & Notification
   - Construct JSONB document_snapshot array
   - INSERT INTO access_logs (...)
   - If result == 'denied' -> INSERT INTO notifications (type = 'access_denied', ...)
   - Return 200 OK with full compliance report
```

### Document Snapshot Schema (JSONB)
Stored in `access_logs.document_snapshot` as an immutable audit snapshot:
```json
[
  {
    "document_type_code": "DC3",
    "document_type_name": "Constancia de Capacitación DC-3",
    "applies_to": "worker",
    "status": "approved",
    "is_valid": true,
    "issued_at": "2026-04-15",
    "expires_at": "2027-04-15",
    "reason": null
  },
  {
    "document_type_code": "REPSE",
    "document_type_name": "Registro REPSE",
    "applies_to": "contractor",
    "status": "expired",
    "is_valid": false,
    "issued_at": "2023-09-01",
    "expires_at": "2026-09-01",
    "reason": "Documento vencido el 2026-09-01"
  }
]
```

---

## 7. Database Schema Mappings & sqlc Specifications

### 7.1 Key Schema Tables and Mappings
1. `tenants`: Primary isolation boundary.
2. `plants`: Facilities with access policies (`block_on_expired_docs`).
3. `access_points`: Gates / checkpoints (`Caseta Principal`, `Torniquete 1`).
4. `users`: Identity table with `role` enum (`guard`, `plant_admin`, `worker`, etc.).
5. `refresh_tokens`: Stored SHA-256 hashes of refresh tokens with revocation support.
6. `contractors`: Global contractor companies.
7. `plant_contractors`: Tenant/plant authorization status for contractors (`active`, `pending`, `suspended`, `blocked`).
8. `contractor_workers`: Worker-to-contractor assignment with CURP and IMSS/NSS numbers.
9. `document_types`: Catalog of Mexican compliance documents (DC3, REPSE, SUA, SAT 32-D, INE, MEDICAL).
10. `plant_document_requirements`: Plant-level overrides of mandatory document rules.
11. `documents`: Uploaded document metadata and review states.
12. `access_logs`: Immutable gate entry audit log with JSONB snapshot.
13. `notifications`: Alerting and in-app notifications.

### 7.2 Required sqlc Queries (`db/queries/`)

#### File: `db/queries/auth.sql`
```sql
-- name: GetUserByEmail :one
SELECT id, tenant_id, email, password_hash, full_name, phone, role, avatar_r2_key, is_active, last_login_at, created_at, updated_at
FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = now()
WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id, user_id, token_hash, expires_at
) VALUES (
    gen_random_uuid(), $1, $2, $3
) RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1;
```

#### File: `db/queries/access.sql`
```sql
-- name: GetAccessPointByIDAndTenant :one
SELECT id, plant_id, tenant_id, name, description, is_active, created_at, updated_at
FROM access_points
WHERE id = $1 AND tenant_id = $2 AND is_active = TRUE AND deleted_at IS NULL
LIMIT 1;

-- name: ListAccessPointsByTenant :many
SELECT id, plant_id, tenant_id, name, description, is_active, created_at, updated_at
FROM access_points
WHERE tenant_id = $1 AND (plant_id = $2 OR $2 IS NULL) AND is_active = TRUE AND deleted_at IS NULL
ORDER BY name ASC;

-- name: GetPlantByIDAndTenant :one
SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at
FROM plants
WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: GetWorkerDetailsForCheckIn :one
SELECT 
    u.id AS worker_id,
    u.full_name,
    u.email,
    u.phone,
    u.is_active AS user_active,
    u.avatar_r2_key,
    cw.id AS contractor_worker_id,
    cw.contractor_id,
    cw.job_title,
    cw.curp,
    cw.imss_number,
    cw.is_active AS contractor_worker_active,
    c.name AS contractor_name,
    c.rfc AS contractor_rfc,
    c.repse_number AS contractor_repse,
    c.contractor_type,
    c.is_active AS contractor_active
FROM users u
LEFT JOIN contractor_workers cw ON u.id = cw.user_id AND cw.deleted_at IS NULL
LEFT JOIN contractors c ON cw.contractor_id = c.id AND c.deleted_at IS NULL
WHERE u.id = $1 AND u.role = 'worker' AND u.deleted_at IS NULL
LIMIT 1;

-- name: GetPlantContractorStatus :one
SELECT id, plant_id, tenant_id, contractor_id, status, activated_at
FROM plant_contractors
WHERE plant_id = $1 AND contractor_id = $2 AND tenant_id = $3
LIMIT 1;

-- name: GetPlantEffectiveDocumentRequirements :many
SELECT 
    dt.id AS document_type_id,
    dt.code,
    dt.name,
    dt.applies_to,
    COALESCE(pdr.is_required, dt.is_required_default) AS is_required,
    COALESCE(pdr.validity_days, dt.default_validity_days) AS validity_days
FROM document_types dt
LEFT JOIN plant_document_requirements pdr 
    ON dt.id = pdr.document_type_id AND pdr.plant_id = $1 AND pdr.tenant_id = $2
WHERE (dt.tenant_id IS NULL OR dt.tenant_id = $2)
  AND dt.is_active = TRUE
  AND COALESCE(pdr.is_required, dt.is_required_default) = TRUE
ORDER BY dt.applies_to ASC, dt.code ASC;

-- name: GetLatestApprovedDocumentsForWorkerAndContractor :many
SELECT DISTINCT ON (d.document_type_id)
    d.id,
    d.user_id,
    d.contractor_id,
    d.document_type_id,
    dt.code AS document_type_code,
    dt.name AS document_type_name,
    dt.applies_to,
    d.status,
    d.issued_at,
    d.expires_at,
    d.r2_key,
    d.file_name,
    d.created_at
FROM documents d
JOIN document_types dt ON d.document_type_id = dt.id
WHERE (d.user_id = $1 OR d.contractor_id = $2)
  AND d.deleted_at IS NULL
ORDER BY d.document_type_id, d.created_at DESC;

-- name: CreateAccessLog :one
INSERT INTO access_logs (
    id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
    processed_by, checked_in_at, result, denial_reason, document_snapshot, notes
) VALUES (
    gen_random_uuid(), $1, $2, $3, $4, $5, $6, now(), $7, $8, $9, $10
) RETURNING *;

-- name: GetActiveCheckInForWorker :one
SELECT id, tenant_id, plant_id, access_point_id, user_id, checked_in_at
FROM access_logs
WHERE plant_id = $1 AND user_id = $2 AND checked_out_at IS NULL
ORDER BY checked_in_at DESC
LIMIT 1;

-- name: RecordCheckOut :one
UPDATE access_logs
SET checked_out_at = now()
WHERE id = $1 AND checked_out_at IS NULL
RETURNING *;

-- name: ListRecentAccessLogsByPlant :many
SELECT 
    al.id, al.tenant_id, al.plant_id, al.access_point_id, al.user_id, al.contractor_id,
    al.processed_by, al.checked_in_at, al.checked_out_at, al.result, al.denial_reason,
    al.document_snapshot, al.notes,
    u.full_name AS worker_name,
    c.name AS contractor_name,
    ap.name AS access_point_name
FROM access_logs al
JOIN users u ON al.user_id = u.id
LEFT JOIN contractors c ON al.contractor_id = c.id
JOIN access_points ap ON al.access_point_id = ap.id
WHERE al.tenant_id = $1 AND al.plant_id = $2
ORDER BY al.checked_in_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAccessLogsByPlant :one
SELECT COUNT(*) FROM access_logs
WHERE tenant_id = $1 AND plant_id = $2;
```

#### File: `db/queries/notifications.sql`
```sql
-- name: CreateNotification :one
INSERT INTO notifications (
    id, tenant_id, recipient_id, type, title, body, resource_type, resource_id, created_at
) VALUES (
    gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, now()
) RETURNING *;
```

### 7.3 sqlc Configuration (`db/sqlc.yaml`)
```yaml
version: "2"
sql:
  - schema: "db/migrations/000001_init.up.sql"
    queries: "db/queries/"
    gen:
      go:
        package: "sqlc"
        out: "internal/adapter/postgres/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_empty_slices: true
```

---

## 8. Multi-Tenancy & Tenant Isolation Enforcement

Per `DEVELOPMENT_RULES.md` §3:
1. **Context Extraction:**
   - The JWT contains `tenant_id` (UUID).
   - Middleware `jwt_auth.go` parses the claims and validates the token.
   - Middleware `tenant.go` sets `tenant_id` into context via helper package:
     ```go
     ctx = tenant.WithContext(ctx, claims.TenantID)
     ```
2. **Accessing Tenant:**
   - `tenant.FromContext(ctx)` is the sole mechanism to access the tenant ID in use cases and handlers:
     ```go
     tenantID, err := tenant.FromContext(ctx)
     if err != nil {
         return nil, ErrTenantRequired
     }
     ```
3. **Database Isolation (`WHERE tenant_id = $1`):**
   - All tenant-scoped tables (`plants`, `access_points`, `plant_contractors`, `plant_document_requirements`, `access_logs`) MUST be queried with `tenant_id` as the primary filter.
   - Handlers and use cases NEVER accept `tenant_id` as a client-provided parameter.
   - If an access point or plant ID is supplied in a request, it is queried with `WHERE id = $req_id AND tenant_id = $ctx_tenant_id`. If it belongs to a different tenant, PostgreSQL returns 0 rows, resulting in 404 `ACCESS_POINT_NOT_FOUND`.

---

## 9. Authentication & Authorization Specification

Per `DEVELOPMENT_RULES.md` §7:

### 9.1 JWT Token Claims
```go
package jwt

import (
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type Claims struct {
    jwt.RegisteredClaims
    UserID   uuid.UUID `json:"sub"`
    TenantID uuid.UUID `json:"tenant_id,omitempty"`
    Role     string    `json:"role"`
}
```
- Signing method: `HMAC-SHA256` (`jwt.SigningMethodHS256`).
- Secret: `JWT_SECRET` environment variable (minimum 32 characters validated on boot).
- Access Token Lifetime: **1 hour** (`time.Now().Add(1 * time.Hour)`).
- Refresh Token Lifetime: **7 days** (`time.Now().Add(7 * 24 * time.Hour)`), stored in database table `refresh_tokens` as SHA-256 hash.

### 9.2 Password Hashing
- Library: `golang.org/x/crypto/bcrypt`
- Cost: `bcrypt.DefaultCost` (10).
- Login comparison: `bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))`.

### 9.3 RBAC Roles
Defined in schema `user_role` enum:
- `super_admin`: Platform operator. Can bypass tenant filter if explicitly running global operations.
- `plant_admin`: Plant customer admin. Full management of their tenant and plants.
- `ehs_manager`: Health and safety manager. Approves/rejects documents, reviews compliance metrics.
- `guard`: Caseta security personnel. Permitted to call `POST /api/v1/access/check-in`, `POST /api/v1/access/check-out`, `GET /api/v1/access-points`.
- `contractor_admin`: Admin of contractor company.
- `worker`: Individual contractor worker.

---

## 10. Unit and Integration Testing Specification

Per `DEVELOPMENT_RULES.md` §9 and `ORIGINAL_REQUEST.md`:

### 10.1 Unit Testing Strategy
- **Location:** Co-located next to implementation files (e.g., `internal/core/usecase/access/check_in_usecase_test.go`).
- **Isolation:** 100% in-memory with mocked repository interfaces using `github.com/stretchr/testify/mock` or handwritten mocks.
- **Coverage Target:** Minimum 80% for use case logic.
- **Test Scenarios Matrix for `CheckInUseCase`:**
  1. *Success Case (Full Compliance):* Worker active, contractor active, all documents approved and within validity -> Returns `result: granted`, zero denial reasons, valid snapshot.
  2. *Worker Inactive:* Worker `is_active = false` -> Returns `result: denied`, reason mentions inactive worker.
  3. *Contractor Blocked:* Contractor has `plant_contractors.status = 'blocked'` -> Returns `result: denied`, reason mentions contractor blocked.
  4. *Contractor Suspended:* Contractor has `status = 'suspended'` -> Returns `result: denied`.
  5. *Contractor Pending:* Contractor has `status = 'pending'` -> Returns `result: denied`.
  6. *Missing Mandatory Document:* Worker has no DC-3 uploaded -> Returns `result: denied`, reason mentions missing DC-3.
  7. *Expired Document with `block_on_expired_docs = true`:* Worker DC-3 expired yesterday -> Returns `result: denied`, reason mentions expired DC-3.
  8. *Expired Document with `block_on_expired_docs = false`:* Worker DC-3 expired yesterday -> Returns `result: granted_with_warning`, note contains warning.
  9. *Contractor Expired REPSE:* Worker docs valid, but contractor REPSE expired and plant blocks -> Returns `result: denied`, reason mentions contractor REPSE.
  10. *Pending Document Review:* Worker uploaded new DC-3, status is `pending_review` -> Returns `result: denied`, reason indicates pending review.
  11. *Access Point Mismatch:* Access point does not belong to plant/tenant -> Returns error.

### 10.2 Integration Testing Strategy
- **Framework:** `testcontainers-go` booting PostgreSQL 16 image or local test container running `000001_init.up.sql`.
- **Isolation:** Each test runs inside a database transaction that rolls back at test completion, or uses a dynamically created test tenant.
- **Verification of Multi-Tenant Isolation:**
  - Create Tenant 1 (Planta Norte) and Tenant 2 (Planta Sur).
  - Attempt check-in using Guard 1 JWT at Tenant 2's Access Point -> Must fail with 404 / 403.
  - Query access logs as Tenant 1 -> Must return strictly Tenant 1 records, never Tenant 2 records.
- **Execution:** `make test` for unit tests; `make test-integration` for containerized integration tests.

---

## 11. Local Orchestration & Docker Compose Specification

Per `ORIGINAL_REQUEST.md` R3 and `DEVELOPMENT_RULES.md` §14, root `docker-compose.yml` must orchestrate:
1. `db`:
   - Image: `postgres:16-alpine`
   - Ports: `5432:5432`
   - Env: `POSTGRES_DB=asesamo`, `POSTGRES_USER=asesamo`, `POSTGRES_PASSWORD=asesamo_secret`
   - Healthcheck: `pg_isready -U asesamo -d asesamo`
2. `migrate`:
   - Image: `migrate/migrate`
   - Command: `-path=/migrations -database=postgres://asesamo:asesamo_secret@db:5432/asesamo?sslmode=disable up`
   - Volumes: `./db/migrations:/migrations`
   - Depends_on: `db` (healthy)
3. `api`:
   - Build: context `./asesamo-api`
   - Ports: `8080:8080`
   - Env: `APP_PORT=8080`, `DATABASE_URL=postgres://asesamo:asesamo_secret@db:5432/asesamo?sslmode=disable`, `JWT_SECRET=super_secret_asesamo_jwt_key_must_be_32_chars`
   - Depends_on: `migrate` (completed successfully)
4. `web`:
   - Build: context `./asesamo-web`
   - Ports: `3000:3000`
   - Env: `NEXT_PUBLIC_API_URL=http://localhost:8080`
   - Depends_on: `api`
