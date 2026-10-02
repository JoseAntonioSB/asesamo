# Handoff Report: Milestone M2 Backend Core & Authentication Architecture

**Agent**: `teamwork_preview_explorer_m2_1`  
**Milestone**: M2: Backend Core & Authentication (`asesamo-api`)  
**Type**: Hard Handoff (Analysis & Architecture Formulation Complete)  
**Date**: 2026-10-01T22:58:00Z  

---

## 1. Observation

1. **Database Schema & Enums** (`db/migrations/000001_init.up.sql`):
   - Lines 31-42: `tenants` table with `id UUID PRIMARY KEY`, `plan TEXT CHECK (plan IN ('starter', 'pro', 'enterprise'))`.
   - Lines 52-64: `plants` table with `block_on_expired_docs BOOLEAN NOT NULL DEFAULT TRUE`, `tenant_id UUID REFERENCES tenants(id)`.
   - Lines 77-87: `access_points` table with `plant_id UUID`, `tenant_id UUID`.
   - Lines 101-127: `users` table with enum `user_role` (`super_admin`, `plant_admin`, `ehs_manager`, `guard`, `contractor_admin`, `worker`), `password_hash TEXT`, `avatar_r2_key TEXT`.
   - Lines 158-174: `contractors` table (global, not tenant-scoped) with `contractor_type` (`company`, `individual`).
   - Lines 188-202: `plant_contractors` table binding plant and contractor with status enum (`pending`, `active`, `suspended`, `blocked`).
   - Lines 216-230: `contractor_workers` table linking contractor to worker user (`curp`, `imss_number`, `job_title`).
   - Lines 243-261: `document_types` table (`code TEXT`, `applies_to TEXT CHECK ('worker', 'contractor')`, `default_validity_days INT`).
   - Lines 308-345: `documents` table with status enum (`pending_review`, `approved`, `rejected`, `expired`), dates `issued_at DATE`, `expires_at DATE`, check constraint `chk_document_owner`.
   - Lines 363-394: `access_logs` table with enum `access_result` (`granted`, `granted_with_warning`, `denied`), `denial_reason TEXT`, `document_snapshot JSONB NOT NULL DEFAULT '[]'`.

2. **Interface & Acceptance Contracts** (`PROJECT.md` & `test/e2e/`):
   - `PROJECT.md` Lines 69-85: Standard response envelopes `{ "data": ... }` and `{ "error": { "code", "message" } }`.
   - `PROJECT.md` Lines 87-110: `POST /api/v1/auth/login` contract returning JWT with `sub`, `tenant_id`, `role`, `email`, and `user` profile object.
   - `PROJECT.md` Lines 112-187: `POST /api/v1/access/check-in` returning `result` (`granted`, `granted_with_warning`, `denied`), `denial_reason`, `worker`, `contractor`, `plant`, `access_point`, and `document_snapshot` array.
   - `test/e2e/tier1_feature_test.sh` Lines 33-64: `GET /api/v1/health` and `GET /health` returning status 200, `{"status": "ok"}`, `{"database": "connected"}`.
   - `test/e2e/tier1_feature_test.sh` Lines 130-160: `GET /api/v1/access-points` requiring Bearer JWT, returning access point list.
   - `test/e2e/tier2_boundary_test.sh` Lines 36-58: Expired and tampered JWT rejection returning HTTP 401 with error envelope.
   - `test/e2e/tier2_boundary_test.sh` Lines 195-223: Multi-tenant boundary enforcement (`WHERE tenant_id = $1`) returning 404/403 when requesting foreign tenant's access points or logs.

3. **Development & Architectural Directives** (`DEVELOPMENT_RULES.md`):
   - Section 1 (Lines 10-23): Chi router (`github.com/go-chi/chi/v5`), Clean Architecture, PostgreSQL with native bcrypt and `pgx`.
   - Section 2 (Lines 49-111): Strict layer separation: Handlers -> Use Cases -> Repositories. Handlers never speak to DB. Use Cases never import HTTP packages.
   - Section 3 (Lines 113-121): Centralized `tenant.FromContext(ctx)` helper is mandatory; every tenant-scoped query must enforce `WHERE tenant_id = $1` as first condition.
   - Section 7 (Lines 192-205): JWT access token lifetime 1 hour, signed with `JWT_SECRET` (minimum 32 characters).
   - Section 10 (Lines 236-258): Strict environment variable validation on boot; process must terminate if required variables are missing.

4. **Current Repository Workspace State**:
   - `asesamo-api/` directory does not yet exist in the root folder (`/Users/jasbotello/devel/asesamo/`).
   - Installed Go toolchain: `go version go1.21.4 darwin/arm64`.

---

## 2. Logic Chain

1. **Module & Dependency Selection**:
   - Based on *Observation 3* (backend stack specification) and the prompt's instructions:
     - `go.mod` module name must be `asesamo-api`.
     - `github.com/go-chi/chi/v5` and `github.com/go-chi/cors` provide minimal, fast HTTP routing and standard middleware.
     - `github.com/golang-jwt/jwt/v5` cleanly handles HS256 claims encoding and decoding.
     - `github.com/google/uuid` maps directly to PostgreSQL `UUID` primary and foreign keys.
     - `github.com/jackc/pgx/v5` with `stdlib` driver allows standard `database/sql` connection pooling while supporting PostgreSQL 16 features (`citext`, `pgcrypto`, `JSONB`).
     - `golang.org/x/crypto/bcrypt` guarantees seamless password hash validation against the bcrypt hashes generated in `000001_seed_dev_data.sql` via `pgcrypto crypt(..., gen_salt('bf', 10))`.
     - `github.com/go-playground/validator/v10` and `github.com/stretchr/testify` fulfill request DTO validation and unit testing requirements.

2. **Clean Architecture Directory Layout**:
   - Following *Observation 3* (strict Clean Architecture rules), the core business layer (`internal/core/domain/` and `internal/core/usecase/`) is kept 100% free of external frameworks (chi, HTTP, SQL).
   - Domain entities (`entity/`) represent the Mexican industrial access domain (Tenants, Plants, AccessPoints, Users, Contractors, Documents, AccessLogs).
   - Repository interfaces (`repository/`) specify the persistence boundaries required by use cases.
   - Domain errors (`errors/`) declare canonical business errors with Spanish messages for end users and UPPER_SNAKE_CASE codes.
   - Infrastructure (`internal/infra/`) and Adapters (`internal/adapter/http`, `internal/adapter/postgres`) implement the external details.
   - Shared cross-cutting packages (`pkg/tenant`, `pkg/jwt`) provide decoupled helpers without circular dependencies.

3. **Multi-Tenancy and JSONB Storage**:
   - By *Observation 1* and *Observation 3*, `access_logs.document_snapshot` is stored as `JSONB`. In Go, this is implemented as `DocumentSnapshot []DocumentSnapshotItem` implementing `driver.Valuer` and `sql.Scanner`. This enables direct serialization into and scanning from PostgreSQL `JSONB` columns while preserving type safety.
   - By *Observation 3*, `pkg/tenant` provides `WithTenant(ctx, tenantID)` and `FromContext(ctx)` to eliminate manual argument passing and prevent cross-tenant data leaks.

---

## 3. Caveats

1. **Host Go Version vs Docker Target Version**:
   - The host system runs Go 1.21.4 (`go version go1.21.4 darwin/arm64`), whereas the project specification requests Go 1.22+.
   - When running on the host, Go 1.21+ toolchain auto-switching can download the 1.22 toolchain if internet access is present. However, inside Docker containers (e.g. `docker compose up`), the official `golang:1.22-alpine` or `1.23-alpine` image will run natively without host version constraints.
2. **Read-Only Scope**:
   - As an explorer subagent, no files were created or modified outside `.agents/teamwork/teamwork_preview_explorer_m2_1/`.
   - The implementation of `asesamo-api/` source files will be performed by subsequent worker agents following the blueprints in `analysis.md` and this handoff.
3. **Check-In vs Check-Out Scope**:
   - Full check-in compliance evaluation is formally planned for Milestone M3; however, the domain models, repositories, and DTOs designed in this handoff already incorporate both check-in and check-out data contracts to avoid refactoring later.

---

## 4. Conclusion

The Clean Architecture package layout, `go.mod` dependencies, domain entities, repository interfaces, domain errors, and cross-cutting packages have been fully analyzed and formulated in exact Go definitions in:
- `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_1/analysis.md`

Worker agents can now immediately scaffold `asesamo-api` and implement Milestone M2 endpoints (`/health`, `/api/v1/auth/login`, `/api/v1/access-points`) with zero ambiguity regarding data structures, database queries, or multi-tenant contracts.

---

## 5. Verification Method

Once worker agents generate the source code in `asesamo-api/`:

1. **Compilation & Go Mod Verification**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go mod tidy
   go build -v ./...
   ```
   *Expected outcome*: Compiles with zero errors.

2. **Unit Tests Run**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race ./...
   ```
   *Expected outcome*: All unit tests for `pkg/jwt`, `pkg/tenant`, `usecase/auth`, and domain entities pass with 0 failures.

3. **Acceptance & E2E Validation (Once Server is Booted)**:
   ```bash
   # From repository root with running PostgreSQL and API:
   ./test/e2e/test_runner.sh --tier1
   ```
   *Expected outcome*: Feature 1 (Health), Feature 2 (Guard Login), and Feature 3 (Access Points) pass 100% of assertions.

4. **Invalidation Conditions**:
   - Any repository query on multi-tenant tables omitting `WHERE tenant_id = $1`.
   - Any usecase directly importing `net/http` or `github.com/go-chi/chi/v5`.
   - Handlers executing raw SQL or bypassing use cases.
