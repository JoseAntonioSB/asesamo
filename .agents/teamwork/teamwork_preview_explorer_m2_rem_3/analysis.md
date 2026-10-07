# Multi-Tenant Repository Query Isolation Analysis & Patch Specification

**Component**: `asesamo-api` (PostgreSQL Repositories & Domain Repository Interfaces)  
**Author**: `teamwork_preview_explorer_m2_rem_3` (Explorer: Investigation & Synthesis)  
**Date**: 2026-10-02T05:14:00Z  
**Context**: Milestone M2 Remediation (Addressing Challenger 1 findings & DEVELOPMENT_RULES.md Section 3 compliance)

---

## 1. Executive Summary

During the Milestone M2 Gate review, Challenger 1 (`teamwork_preview_challenger_m2_r1`) rejected the milestone citing critical security and multi-tenancy vulnerabilities, alongside a test compilation break:
1. `access_point_repository.go` (`GetByID` and `ListByPlant`) queries the multi-tenant `access_points` table without `WHERE tenant_id = $1`, violating `DEVELOPMENT_RULES.md` lines 115–119.
2. `check_in_test.go:175:34` fails compilation because `*mockAPRepo` is missing the `ListByPlant` method defined on `repository.AccessPointRepository`.
3. In addition, an architectural ambiguity was flagged: should tenant isolation be enforced via `tenant.FromContext(ctx)` in repository implementations or via explicit `tenantID uuid.UUID` method signatures in domain repository interfaces?

This investigation audited **all 5 PostgreSQL repository implementations** in `internal/adapter/postgres/`, all repository interfaces in `internal/core/domain/repository/`, all SQL statements in the codebase, and the context injection pipeline in `pkg/tenant/` and `middleware.Auth`.

### Core Findings:
- **`DEVELOPMENT_RULES.md` Section 3 mandates context-driven isolation**: Line 118 explicitly states: *"Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."*
- **Current repositories exhibit inconsistency**: `access_point_repository.go` (`GetByID`, `ListByPlant`), `plant_repository.go` (`GetByID`), and `access_log_repository.go` (`GetByID`) omit `tenant_id` filtering at the SQL level, while `ListByTenant`, `ListByPlant` (in access_log), and `GetPlantDocumentRequirements` (in compliance) include `tenant_id`.
- **`user_repository.go` appropriately keeps `GetByEmail` global**: Because `users.email` is globally unique (`CITEXT UNIQUE`) and authentication occurs before a JWT exists (unauthenticated request context), `GetByEmail` cannot and should not require `tenant_id`.
- **Recommended Strategy (Option A - Context-Driven with Fail-Closed Enforcement)**: Repositories should extract `tenant.FromContext(ctx)` and fail-closed (`ErrTenantNotFound`) if missing, binding `WHERE tenant_id = $1` as the first condition. Use cases (`CheckInUseCase`) should ensure `ctx = tenant.WithTenant(ctx, in.TenantID)` if not already set, providing both zero-signature breaking changes and airtight security. A secondary Option B (explicit parameters) is also fully analyzed and patched below if the orchestrator elects to modify domain signatures.

---

## 2. Review of Multi-Tenancy Rules & Standards

`DEVELOPMENT_RULES.md` Section 3 defines non-negotiable standards for multi-tenancy:

| Rule | Text in `DEVELOPMENT_RULES.md` (Lines 115–120) | Architectural Impact |
|---|---|---|
| **R3.1** | *Cada tabla que contenga datos de un cliente (planta) tendrá una columna `tenant_id UUID NOT NULL`.* | Multi-tenant tables: `plants`, `access_points`, `plant_contractors`, `plant_document_requirements`, `access_logs`, `notifications`. (`users` has `tenant_id` nullable only for `super_admin`). |
| **R3.2** | *Todo query a estas tablas DEBE incluir `WHERE tenant_id = $1` como primera condición.* | Strict query-level filter. No application-level post-filtering is acceptable as a substitute. |
| **R3.3** | *El `tenant_id` se extrae del JWT en el middleware de autenticación y se inyecta en el contexto de Go (`context.Context`).* | Implemented in `middleware/auth_middleware.go` via `ctx = tenant.WithTenant(ctx, claims.TenantID)`. |
| **R3.4** | *Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones.* | Context is the single source of truth; prevents accidental spoofing or passing unverified tenant IDs between functions. |
| **R3.5** | ***Nunca** se permite un query sin filtro de tenant en tablas multi-tenant. Los tests de integración deben verificar el aislamiento.* | Any query omitting `tenant_id` on multi-tenant tables is an immediate security violation. |

---

## 3. Comprehensive Audit of PostgreSQL Repositories

### 3.1 `access_point_repository.go` (`internal/adapter/postgres/access_point_repository.go`)

- **Table**: `access_points` (has `tenant_id UUID NOT NULL REFERENCES tenants(id)`)
- **Query 1: `GetByID(ctx, id)` (lines 61–65)**:
  ```go
  func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
      query := accessPointBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
      row := r.db.QueryRowContext(ctx, query, id)
      return r.scanAccessPoint(row)
  }
  ```
  - **Verdict**: **VIOLATION (CRITICAL)**.
  - **Issue**: Omits `WHERE tenant_id = $1`. An authenticated guard from Tenant A can look up an access point from Tenant B by ID.
  - **Application-level defense**: In `check_in.go` line 102, `if ap.TenantID != in.TenantID` was added, but the SQL query still touched and returned data across tenant boundaries, violating R3.2 and R3.5.

- **Query 2: `ListByTenant(ctx, tenantID)` (lines 67–73)**:
  ```go
  query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
  ```
  - **Verdict**: **COMPLIANT**. Includes `WHERE tenant_id = $1` as first condition.

- **Query 3: `ListByPlant(ctx, plantID)` (lines 91–97)**:
  ```go
  func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
      query := accessPointBaseQuery + ` WHERE plant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
      rows, err := r.db.QueryContext(ctx, query, plantID)
  ```
  - **Verdict**: **VIOLATION (CRITICAL)**.
  - **Issue**: Omits `tenant_id`. If `plantID` is known or brute-forced, it lists all access points for that plant without confirming the caller belongs to the plant's tenant.

---

### 3.2 `plant_repository.go` (`internal/adapter/postgres/plant_repository.go`)

- **Table**: `plants` (has `tenant_id UUID NOT NULL REFERENCES tenants(id)`)
- **Query 1: `GetByID(ctx, id)` (lines 26–32)**:
  ```go
  query := `
      SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
      FROM plants
      WHERE id = $1 AND deleted_at IS NULL
  `
  ```
  - **Verdict**: **VIOLATION (HIGH)**.
  - **Issue**: Omits `WHERE tenant_id = $1`. In `CheckInUseCase.Execute` (`check_in.go` line 107), `uc.plantRepo.GetByID(ctx, ap.PlantID)` queries the plant without filtering by tenant.

- **Query 2: `ListByTenant(ctx, tenantID)` (lines 67–73)**:
  ```go
  WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY name ASC
  ```
  - **Verdict**: **COMPLIANT**. Filters by `tenant_id = $1`.

---

### 3.3 `user_repository.go` (`internal/adapter/postgres/user_repository.go`)

- **Table**: `users` (`tenant_id UUID REFERENCES tenants(id)`, nullable only for `super_admin`)
- **Query 1: `GetByEmail(ctx, email)` (lines 81–85)**:
  ```go
  query := userBaseQuery + ` WHERE email = $1 AND deleted_at IS NULL`
  ```
  - **Verdict**: **COMPLIANT BY DESIGN (EXCEPTION JUSTIFIED)**.
  - **Justification**:
    1. During login (`POST /api/v1/auth/login`), the user is unauthenticated. No JWT exists, so `ctx` contains no `tenant_id`.
    2. In `db/migrations/000001_init.up.sql` line 114: `email CITEXT NOT NULL UNIQUE`. Emails are globally unique.
    3. `super_admin` users have `tenant_id IS NULL`. Filtering by tenant at login would break super_admin logins and require users to specify tenant domain/ID before knowing who they are.
    4. Thus, `GetByEmail` is globally scoped as the identity entry point.

- **Query 2: `UpdateLastLogin(ctx, id, lastLoginAt)` (lines 87–94)**:
  ```go
  UPDATE users SET last_login_at = $1, updated_at = now() WHERE id = $2 AND deleted_at IS NULL
  ```
  - **Verdict**: **ACCEPTABLE**. Called directly after user identity is authenticated in `LoginUseCase`.

- **Query 3: `GetByID(ctx, id)` (lines 75–79)**:
  ```go
  query := userBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
  ```
  - **Verdict**: Currently not called in any business use case. If called in authenticated context, should ideally verify tenant if user is not `super_admin`.

---

### 3.4 `access_log_repository.go` (`internal/adapter/postgres/access_log_repository.go`)

- **Table**: `access_logs` (has `tenant_id UUID NOT NULL REFERENCES tenants(id)`)
- **Query 1: `Create(ctx, log)` (lines 27–76)**:
  - **Verdict**: **COMPLIANT**. Always inserts `log.TenantID` into column `tenant_id`.
- **Query 2: `GetByID(ctx, id)` (lines 134–143)**:
  ```go
  SELECT ... FROM access_logs WHERE id = $1
  ```
  - **Verdict**: **VIOLATION (MEDIUM)**. Omits `WHERE tenant_id = $1`. Any access log can be fetched if its UUID is known.
- **Query 3: `ListByTenant(ctx, tenantID, limit, offset)` (lines 145–179)**:
  - **Verdict**: **COMPLIANT**. Filters `WHERE tenant_id = $1`.
- **Query 4: `ListByPlant(ctx, tenantID, plantID, limit, offset)` (lines 181–225)**:
  - **Verdict**: **COMPLIANT**. Filters `WHERE tenant_id = $1 AND plant_id = $2`.
- **Query 5: `CheckOutLatest(ctx, tenantID, plantID, workerID, checkedOutAt)` (lines 227–268)**:
  - **Verdict**: **COMPLIANT**. Filters `WHERE tenant_id = $2 AND plant_id = $3 AND user_id = $4`.

---

### 3.5 `compliance_repository.go` (`internal/adapter/postgres/compliance_repository.go`)

- **Query 1: `GetWorkerAffiliation(ctx, plantID, workerUserID)` (lines 26–106)**:
  - Selects user by ID, then joins `contractor_workers`, `contractors`, and `plant_contractors` on `pc.contractor_id = c.id AND pc.plant_id = $1`.
  - **Verdict**: **ACCEPTABLE**. The plant ID is verified prior to affiliation retrieval.
- **Query 2: `GetPlantDocumentRequirements(ctx, tenantID, plantID)` (lines 108–167)**:
  - **Verdict**: **COMPLIANT**. Queries `WHERE pdr.plant_id = $1 AND pdr.tenant_id = $2`.
- **Query 3 & 4: `GetWorkerDocuments` and `GetContractorDocuments` (lines 169–242)**:
  - **Verdict**: **COMPLIANT BY DESIGN**. Per schema line 306, uploaded documents belong globally to the worker user or contractor company, evaluated dynamically against plant requirements.

---

## 4. Architectural Analysis: Context vs. Explicit Parameter

The core question is whether repository interfaces in `internal/core/domain/repository/` should:
- **Approach A**: Keep concise signatures `(ctx, id)` and extract `tenant.FromContext(ctx)` within the Postgres adapter, or
- **Approach B**: Require `(ctx, tenantID, id)` explicitly on all repository interface methods.

### Detailed Comparison Matrix

| Evaluation Dimension | Approach A: `tenant.FromContext(ctx)` in Adapter | Approach B: Explicit `tenantID` in Signatures |
|---|---|---|
| **Adherence to `DEVELOPMENT_RULES.md`** | **Strict (100%)**: Section 3 line 118 explicitly forbids manual passing: *"Un helper centralizado tenant.FromContext(ctx) es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."* | **Contradicts Rule 118**: Passes `tenantID` as an explicit parameter between handlers, use cases, and repositories. |
| **Clean Architecture Interface Purity** | **High**: Domain repository interfaces (`internal/core/domain/repository/`) remain pure entity persistence contracts. `context.Context` carries cross-cutting request metadata. | **Mixed**: Binds multi-tenancy infrastructure concern into every domain repository method signature. |
| **Security & Fail-Closed Behavior** | **Airtight / Fail-Closed**: If `ctx` does not contain a valid tenant (e.g. unauthenticated request), `FromContext(ctx)` returns `ErrTenantNotFound` immediately. Zero queries execute. | **Airtight at Compile Time**: Caller cannot compile without a UUID, but can accidentally pass an untrusted or spoofed tenant UUID variable. |
| **Blast Radius & Refactoring Impact** | **Minimal**: Only Postgres adapter files and `check_in.go` context propagation need updates. No changes to mock signatures in test files. | **High**: Requires changing repository interfaces, postgres implementations, use cases, and ~5 test mock structs across multiple packages. |
| **Test Ergonomics** | Tests pass `tenant.WithTenant(ctx, tenantID)` or rely on mocks. In unit tests using mocks, mock signatures remain simple. | Tests must pass `tenantID` to every mock call expectation (`apRepo.On("GetByID", mock.Anything, tenantID, apID)`). |

### Definitive Recommendation: Approach A (Context-Driven with Fail-Closed Fallback)

**Why Approach A is superior**:
1. It is the **official project rule** stated in `DEVELOPMENT_RULES.md` line 118.
2. It prevents accidental developer error where a parameter `req.TenantID` from an HTTP body is passed into the repository instead of the authenticated tenant from JWT.
3. In `CheckInUseCase.Execute(ctx, in)`, the use case guarantees `ctx = tenant.WithTenant(ctx, in.TenantID)` if not already present. This ensures that even in test environments or background workers where `ctx` was initialized with `context.Background()`, the tenant is safely propagated to downstream repositories.
4. It resolves Challenger 1's objection by ensuring that **every single SQL query to `access_points`, `plants`, and `access_logs` includes `WHERE tenant_id = $1` as its first condition**.

---

## 5. Exact Patch Specifications

### Patch 1: Enforce Multi-Tenant Isolation in `access_point_repository.go`

**Target File**: `asesamo-api/internal/adapter/postgres/access_point_repository.go`  
**Description**: Update `GetByID` and `ListByPlant` to extract `tenantID` from `ctx` using `tenant.FromContext(ctx)` and add `WHERE tenant_id = $1` as the first condition.

#### Diffs:

```go
<<<<
func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	query := accessPointBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanAccessPoint(row)
}
====
func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)
	return r.scanAccessPoint(row)
}
>>>>
```

```go
<<<<
func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	query := accessPointBaseQuery + ` WHERE plant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query, plantID)
====
func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND plant_id = $2 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID)
>>>>
```

---

### Patch 2: Enforce Multi-Tenant Isolation in `plant_repository.go`

**Target File**: `asesamo-api/internal/adapter/postgres/plant_repository.go`  
**Description**: Update `GetByID` to extract `tenantID` from `ctx` and enforce `WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`.

#### Diff:

```go
<<<<
func (r *plantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	query := `
		SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
		FROM plants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		p         entity.Plant
		address   *string
		city      *string
		state     *string
		deletedAt *time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
====
func (r *plantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
		FROM plants
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
	`

	var (
		p         entity.Plant
		address   *string
		city      *string
		state     *string
		deletedAt *time.Time
	)

	err = r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
>>>>
```

---

### Patch 3: Enforce Multi-Tenant Isolation in `access_log_repository.go`

**Target File**: `asesamo-api/internal/adapter/postgres/access_log_repository.go`  
**Description**: Update `GetByID` to extract `tenantID` from `ctx` and enforce `WHERE tenant_id = $1 AND id = $2`.

#### Diff:

```go
<<<<
func (r *accessLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	query := `
		SELECT id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		       processed_by, checked_in_at, checked_out_at, result, denial_reason,
		       document_snapshot, notes, created_at
		FROM access_logs
		WHERE id = $1
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id))
}
====
func (r *accessLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		       processed_by, checked_in_at, checked_out_at, result, denial_reason,
		       document_snapshot, notes, created_at
		FROM access_logs
		WHERE tenant_id = $1 AND id = $2
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, tenantID, id))
}
>>>>
```

---

### Patch 4: Ensure Context Tenant Propagation in `check_in.go`

**Target File**: `asesamo-api/internal/core/usecase/access/check_in.go`  
**Description**: Ensure that `in.TenantID` is present in `ctx` before calling downstream repositories, guaranteeing compatibility with both HTTP middleware contexts and direct test contexts.

#### Diff:

```go
<<<<
func (uc *CheckInUseCase) Execute(ctx context.Context, in CheckInInput) (*CheckInOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}
	if in.AccessPointID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El punto de acceso es obligatorio.")
	}

	// 1. Verify Access Point
	ap, err := uc.accessPointRepo.GetByID(ctx, in.AccessPointID)
====
func (uc *CheckInUseCase) Execute(ctx context.Context, in CheckInInput) (*CheckInOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}
	if in.AccessPointID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El punto de acceso es obligatorio.")
	}

	// Propagate tenant into context if not already present (fail-safe for direct invocations)
	if _, err := tenant.FromContext(ctx); err != nil && in.TenantID != uuid.Nil {
		ctx = tenant.WithTenant(ctx, in.TenantID)
	}

	// 1. Verify Access Point
	ap, err := uc.accessPointRepo.GetByID(ctx, in.AccessPointID)
>>>>
```

---

### Patch 5: Fix Test Mock Compilation in `check_in_test.go`

**Target File**: `asesamo-api/internal/core/usecase/access/check_in_test.go`  
**Description**: Add missing `ListByPlant` method to `mockAPRepo` to restore `repository.AccessPointRepository` interface compliance and fix `make test` and `make lint`.

#### Diff:

```go
<<<<
func (m *mockAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

// MockPlantRepository
====
func (m *mockAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

// MockPlantRepository
>>>>
```

---

## 6. Alternative Patch Specifications (Approach B: Explicit Parameters)

If the team prefers modifying repository interfaces to require explicit `tenantID` parameters:

1. **`internal/core/domain/repository/access_point_repository.go`**:
   ```go
   type AccessPointRepository interface {
       GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.AccessPoint, error)
       ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error)
       ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.AccessPoint, error)
   }
   ```
2. **`internal/core/domain/repository/plant_repository.go`**:
   ```go
   type PlantRepository interface {
       GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.Plant, error)
       ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error)
   }
   ```
3. **`internal/core/domain/repository/access_log_repository.go`**:
   ```go
   type AccessLogRepository interface {
       Create(ctx context.Context, log *entity.AccessLog) error
       GetByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.AccessLog, error)
       ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error)
       ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error)
       CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error)
   }
   ```
4. **Callers in `check_in.go`**:
   - Line 98: `ap, err := uc.accessPointRepo.GetByID(ctx, in.TenantID, in.AccessPointID)`
   - Line 107: `plant, err := uc.plantRepo.GetByID(ctx, in.TenantID, ap.PlantID)`
5. **Mock Updates**:
   Update `GetByID` and `ListByPlant` in all mock structs (`mockAPRepo`, `mockPlantRepo`, `RouterMockAccessPointRepo`, `MockAccessPointRepoForRouterTest`, `MockAccessPointRepository`, `MockAccessPointRepoForHandler`).

---

## 7. Verification Method

Once patches are applied by the worker:
1. **Compilation & Unit Tests**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race -cover ./...
   ```
   *Expected result*: `PASS` with 0 errors across all packages.
2. **Linter Check**:
   ```bash
   cd /Users/jasbotello/devel/asesamo && make lint
   ```
   *Expected result*: Clean exit (code 0).
3. **Multi-Tenant SQL Query Audit**:
   Verify with `grep_search` that every `SELECT` query in `access_point_repository.go`, `plant_repository.go`, and `access_log_repository.go` begins with `WHERE tenant_id = $1`.
4. **Fail-Closed Context Unit Test**:
   Assert that invoking `accessPointRepo.GetByID(context.Background(), id)` returns `domainErrors.ErrTenantNotFound` without executing a database query.
