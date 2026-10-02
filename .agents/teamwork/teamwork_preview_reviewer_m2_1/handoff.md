# Milestone M2 Gate Review Report: Go Backend Core & Authentication

- **Reviewer**: `teamwork_preview_reviewer_m2_1` (Roles: `reviewer`, `critic`)
- **Milestone**: M2 (Backend Core & Authentication)
- **Target Component**: `asesamo-api`
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_reviewer_m2_1`
- **Date**: 2026-10-01T23:13:00Z
- **Verdict**: **APPROVE**

---

## 1. Observation

### 1.1 Direct Inspection of Clean Architecture Decoupling
- **Handlers**:
  - `internal/adapter/http/handler/auth_handler.go` lines 13-22: `AuthHandler` depends solely on `*auth.LoginUseCase`. Line 43 calls `h.loginUseCase.Execute(r.Context(), ...)`. There are zero direct SQL, database pool, or repository imports.
  - `internal/adapter/http/handler/access_point_handler.go` lines 13-22: `AccessPointHandler` depends solely on `*access.ListAccessPointsUseCase`. Line 32 calls `h.listUseCase.Execute(r.Context(), tenantID)`. There are zero database imports.
  - `internal/adapter/http/handler/health_handler.go` lines 21-35: Takes a `database.Pinger` interface rather than concrete DB types.
- **Use Cases**:
  - `internal/core/usecase/auth/login.go` lines 4-16: Imports standard library (`context`, `strings`, `time`), domain packages (`entity`, `domainErrors`, `repository`), `pkg/jwt`, `uuid`, and `golang.org/x/crypto/bcrypt`. **Zero HTTP imports (`net/http`, `chi`) exist in the package.**
  - `internal/core/usecase/access/list_access_points.go` lines 4-12: Imports only `context`, domain packages (`entity`, `domainErrors`, `repository`), and `uuid`. **Zero HTTP imports.**
- **Domain Entities**:
  - All files in `internal/core/domain/entity/` (`tenant.go`, `plant.go`, `access_point.go`, `user.go`, `contractor.go`, `contractor_worker.go`, `plant_contractor.go`, `document_type.go`, `document.go`, `access_log.go`, `document_snapshot.go`) are pure domain structs. They depend solely on standard library (`time`, `database/sql/driver`, `encoding/json`, `fmt`) and `github.com/google/uuid`. No HTTP, framework, or database engine drivers are imported.

### 1.2 Entity Completeness Verification
- `Tenant` (`tenant.go`:19-29): Contains `ID`, `Name`, `RFC`, `LogoR2Key`, `Plan`, `IsActive`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `Plant` (`plant.go`:10-21): Contains `ID`, `TenantID`, `Name`, `Address`, `City`, `State`, `BlockOnExpiredDocs`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `AccessPoint` (`access_point.go`:10-20): Contains `ID`, `PlantID`, `TenantID`, `Name`, `Description`, `IsActive`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `User` (`user.go`:22-36): Contains `ID`, `TenantID`, `Email`, `PasswordHash`, `FullName`, `Phone`, `Role`, `AvatarR2Key`, `IsActive`, `LastLoginAt`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `Contractor` (`contractor.go`:18-29): Contains `ID`, `Name`, `RFC`, `REPSENumber`, `ContractorType`, `AdminUserID`, `IsActive`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `Worker` (`contractor_worker.go`:10-21): Implemented as `ContractorWorker` with `ID`, `ContractorID`, `UserID`, `JobTitle`, `CURP`, `IMSSNumber`, `IsActive`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `Document` (`document.go`:20-39): Contains `ID`, `UserID`, `ContractorID`, `DocumentTypeID`, `R2Key`, `FileName`, `FileSizeBytes`, `MimeType`, `IssuedAt`, `ExpiresAt`, `Status`, `ReviewedBy`, `ReviewedAt`, `RejectionReason`, `AIConfidence`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
- `AccessLog` (`access_log.go`:19-34): Contains `ID`, `TenantID`, `PlantID`, `AccessPointID`, `UserID`, `ContractorID`, `ProcessedBy`, `CheckedInAt`, `CheckedOutAt`, `Result`, `DenialReason`, `DocumentSnapshot`, `Notes`, `CreatedAt`.
- `DocumentSnapshot` (`document_snapshot.go`:24-67): Implements `driver.Valuer` (`Value()` lines 27-36) and `sql.Scanner` (`Scan()` lines 39-67) for PostgreSQL JSONB serialization/deserialization.
- Supplementary entities `DocumentType` (`document_type.go`) and `PlantContractor` (`plant_contractor.go`) are also fully implemented.

### 1.3 Domain Repository Interfaces
- `internal/core/domain/repository/user_repository.go`:
  ```go
  type UserRepository interface {
      GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
      GetByEmail(ctx context.Context, email string) (*entity.User, error)
      UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error
  }
  ```
- `internal/core/domain/repository/access_point_repository.go`:
  ```go
  type AccessPointRepository interface {
      GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error)
      ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error)
      ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)
  }
  ```
- `internal/core/domain/repository/plant_repository.go`:
  ```go
  type PlantRepository interface {
      GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error)
      ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error)
  }
  ```
- `internal/core/domain/repository/document_repository.go`:
  ```go
  type DocumentRepository interface {
      GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error)
      ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error)
      ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error)
  }
  ```
- `internal/core/domain/repository/access_log_repository.go`:
  ```go
  type AccessLogRepository interface {
      Create(ctx context.Context, log *entity.AccessLog) error
      GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error)
      ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error)
  }
  ```
All methods take `ctx context.Context` as their first parameter, enforce strong typing with `uuid.UUID` and domain entity pointers/slices, and return standard Go errors.

### 1.4 Usecase Implementation & Test Coverage
- `internal/core/usecase/auth/login.go`:
  - Enforces non-empty trimmed email and password.
  - Queries `userRepo.GetByEmail(ctx, email)`.
  - Replaces `sql.ErrNoRows` / not-found with generic `ErrInvalidCredentials` to mitigate user enumeration.
  - Verifies `user.IsActive` (returns `ErrUserInactive` if false).
  - Uses `bcrypt.CompareHashAndPassword` for actual hash verification.
  - Generates signed JWT using `jwt.SignToken(user.ID, tenantID, string(user.Role), uc.jwtSecret)`.
  - Best-effort updates `LastLoginAt` via `userRepo.UpdateLastLogin`.
  - `login_test.go`: 8 test cases using `testify/mock`, achieving **100.0% statement coverage**.
- `internal/core/usecase/access/list_access_points.go`:
  - Validates `tenantID != uuid.Nil` (returns `ErrTenantNotFound` if nil).
  - Calls `accessPointRepo.ListByTenant(ctx, tenantID)`.
  - Normalizes `nil` slice to empty slice `[]entity.AccessPoint{}` so JSON responses serialize as `[]` rather than `null`.
  - `list_access_points_test.go`: 4 test cases using `testify/mock`, achieving **100.0% statement coverage**.

### 1.5 Independent Command Verification
1. `cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race ./internal/core/...`:
   - `asesamo-api/internal/core/domain/entity`: PASS (81.8% coverage)
   - `asesamo-api/internal/core/domain/errors`: PASS (95.0% coverage)
   - `asesamo-api/internal/core/usecase/access`: PASS (100.0% coverage)
   - `asesamo-api/internal/core/usecase/auth`: PASS (100.0% coverage)
   - Zero data races detected. Exit code: 0.
2. `make test` from repository root:
   - All unit tests across all packages pass with zero failures. Exit code: 0.
3. `make lint && make build` from repository root:
   - `golangci-lint`: 0 warnings, 0 errors. Exit code: 0.
   - `make build`: Generated `asesamo-api/bin/api` (15 MB binary). Exit code: 0.

### 1.6 Integrity Violation Audit
- No hardcoded test credentials or bypass paths exist in use cases or adapters.
- Real `bcrypt.CompareHashAndPassword` and `jwt.SignToken` / `jwt.ValidateToken` algorithms are invoked.
- Repositories in `internal/adapter/postgres/` contain genuine SQL queries with parameter binding (`$1`).
- Verification outputs and logs were independently executed and confirmed by this reviewer.
- **Integrity Violation Status: CLEAN (0 violations detected)**.

---

## 2. Logic Chain

1. **Clean Architecture Boundary Adherence** (Obs 1.1, 1.3, 1.4):
   - The Dependency Rule mandates that source code dependencies point inward toward business logic and domain entities.
   - Handlers in `internal/adapter/http/handler/` only reference use case structs, fulfilling isolation from persistence.
   - Use cases in `internal/core/usecase/` depend only on abstract repository interfaces in `internal/core/domain/repository/` and domain entities, fulfilling isolation from delivery mechanisms (HTTP).
   - Domain entities and repositories have zero external dependencies other than standard library and UUID, fulfilling purity.

2. **Completeness & Interface Contract Conformance** (Obs 1.2, 1.3):
   - All 9 required domain entities plus JSONB scanner/valuer for `DocumentSnapshot` are fully defined with fields mapping to the database schema (`db/migrations/000001_init.up.sql`).
   - Repository interfaces cover all required persistence operations for User, AccessPoint, Plant, Document, and AccessLog, providing the contract needed for both Milestone M2 and the upcoming M3 verification engine.

3. **Security, Quality & Multi-Tenancy Invariants** (Obs 1.1, 1.4, 1.5):
   - `pkg/tenant` provides unexported context key types and strictly validates non-nil UUIDs, eliminating cross-tenant key collisions or nil-tenant queries.
   - `internal/adapter/postgres/access_point_repository.go` explicitly enforces `WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL` on list queries.
   - Password comparison uses standard `bcrypt` and mitigates user enumeration attacks.
   - The test suite runs with `-race` enabled and passes 100% of cases, with 100% statement coverage on both `auth` and `access` use cases, exceeding the 80% threshold specified in `DEVELOPMENT_RULES.md` line 224.

---

## 3. Adversarial Review & Critic Stress-Testing

| # | Challenge Dimension | Scenario / Stress Test | Predicted / Observed Behavior | Assessment |
|---|---------------------|------------------------|-------------------------------|------------|
| 1 | **JWT Algorithm Confusion** | Attacker crafts a token with `alg: none` or an HMAC token signed with an RSA public key. | `pkg/jwt/jwt.go` enforces HMAC validation (`jwtlib.WithValidMethods([]string{"HS256"})`) and checks `t.Method.(*jwtlib.SigningMethodHMAC)`. Token rejected with `ErrInvalidToken`. | **PASS (Robust)** |
| 2 | **Weak JWT Secret** | Operator configures a secret key shorter than 32 characters in production. | `config.Validate()` fails fast on startup with `secret_too_short`; `jwt.SignToken` and `jwt.ValidateToken` return `ErrSecretTooShort`. | **PASS (Robust)** |
| 3 | **Account Enumeration via Timing / Status** | Attacker queries login endpoint with invalid email vs valid email + wrong password. | Both branches return identical `domainErrors.ErrInvalidCredentials` mapped to HTTP 401 with identical Spanish message `"Credenciales inválidas. Verifique su correo y contraseña."`. | **PASS (Robust)** |
| 4 | **Nil Tenant Context Leakage** | An unauthenticated or malformed request attempts to call `ListAccessPointsUseCase` with a nil UUID. | `ListAccessPointsUseCase.Execute` checks `tenantID == uuid.Nil` and immediately aborts with `ErrTenantNotFound`. Middleware returns 401 Unauthorized. | **PASS (Robust)** |
| 5 | **SQL Injection via String Inputs** | Attacker sends malicious SQL payloads in login email or query parameters. | `user_repository.go` and `access_point_repository.go` use parameterized queries with `$1` via `database/sql`, preventing SQL injection. | **PASS (Robust)** |
| 6 | **JSON Serialization of Empty Lists** | Tenant has zero access points in database. | `ListAccessPointsUseCase.Execute` normalizes `nil` to `[]entity.AccessPoint{}` so the client receives `"data": []` rather than `"data": null`. | **PASS (Robust)** |

---

## 4. Caveats

1. **Super Admin Without Tenant ID**:
   - In `internal/core/usecase/auth/login.go`, if a `super_admin` user has `user.TenantID == nil`, the token claim `tenant_id` is set to `uuid.Nil`. Subsequent requests to tenant-scoped endpoints (`/api/v1/access-points`) by this token will fail with 401/TenantNotFound because `tenant.FromContext` rejects `uuid.Nil`.
   - *Assessment*: This is acceptable for Phase 1 MVP, where authentication and check-in flows are scoped to caseta guards belonging to a specific customer tenant. Multi-tenant administrative browsing for super-admins can be addressed in future phases.
2. **No other caveats.**

---

## 5. Conclusion

The deliverable for Milestone M2 strictly adheres to all architectural, security, entity completeness, and testing requirements specified in `DEVELOPMENT_RULES.md`, `PROJECT.md`, and `ORIGINAL_REQUEST.md`:
- **Clean Architecture Conformance**: Full decoupling verified (handlers do not touch DB; use cases do not touch HTTP; domain entities are pure).
- **Entity Completeness**: All required entities (`Tenant`, `Plant`, `AccessPoint`, `User`, `Contractor`, `Worker`, `Document`, `AccessLog`, `DocumentSnapshot`) and catalog entities are implemented with complete fields and JSONB support.
- **Repository Interfaces**: Cleanly typed and context-aware.
- **Use Cases & Test Coverage**: 100.0% coverage achieved on `LoginUseCase` and `ListAccessPointsUseCase` using `testify/mock`.
- **Integrity Audit**: Clean; zero integrity violations detected.
- **Verification Commands**: All tests, race detector, linter, and binary compilation pass with zero errors.

**Final Verdict**: **APPROVE**

---

## 6. Verification Method

To independently reproduce this verification:

1. **Run Core Unit Tests with Race Detection**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race ./internal/core/...
   ```
   *Expected result*: All entity, error, and use case tests pass without race conditions; use case coverage is 100%.

2. **Run Full Test Suite with Race Detection & Coverage**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race -cover ./...
   ```
   *Expected result*: All packages pass with zero failures.

3. **Run Project Lint and Build**:
   ```bash
   cd /Users/jasbotello/devel/asesamo
   make lint
   make build
   ```
   *Expected result*: Linter passes with 0 warnings/errors; binary compiles to `asesamo-api/bin/api`.

4. **Invalidation Conditions**:
   - Any failure or race reported by `go test -v -race ./...`.
   - Any usecase coverage falling below 80%.
   - Compilation failure in `make build` or linter errors in `make lint`.
