# Milestone M3: PostgreSQL Repositories, Multi-Tenancy & Integration Tests Analysis

**Author:** teamwork_preview_explorer_m3_3  
**Date:** 2026-10-02  
**Status:** Completed Analysis  
**Target Subsystem:** `asesamo-api` (PostgreSQL Repositories, Multi-Tenant Query Layer, JSONB Serialization, and Test Specifications)

---

## 1. Executive Summary & Problem Boundary

In ASESAMO Phase 1 MVP, gate check-in verification is the central operational flow. The backend must enforce:
1. **Strict Multi-Tenancy**: Every data access on tenant-scoped tables (`plants`, `access_points`, `plant_contractors`, `plant_document_requirements`, `access_logs`) **MUST** include `WHERE tenant_id = $1` as the first filter. Fail-closed isolation must prevent any cross-tenant data leakage.
2. **Immutable Audit Trails with JSONB Snapshots**: Check-in events write an immutable `access_logs` record containing a complete `document_snapshot` in JSONB representing the exact compliance state of the worker and contractor at the moment of entry.
3. **Robust Repository Separation**: Document compliance queries, contractor affiliations, and access log persistence must adhere to Clean Architecture boundaries.
4. **Test Verification with >= 80% Coverage**: Comprehensive unit tests covering all 7 check-in and check-out scenarios, plus integration test specifications against PostgreSQL.

---

## 2. Multi-Tenant Repository Query Analysis

### 2.1 Multi-Tenant Data Model Classification

A critical finding from analyzing `db/migrations/000001_init.up.sql` is the hybrid nature of entities in ASESAMO:

| Entity / Table | Tenancy Model | Scope / Scoping Mechanism | Multi-Tenant SQL Rule |
| :--- | :--- | :--- | :--- |
| `tenants` | Tenant Root | The tenant company itself. | `WHERE id = $1` |
| `plants` | Tenant-Scoped | Owned by a tenant (`tenant_id UUID NOT NULL REFERENCES tenants(id)`). | `WHERE tenant_id = $1 AND id = $2` |
| `access_points` | Tenant-Scoped | Checkpoint at a plant (`tenant_id UUID NOT NULL`, `plant_id UUID NOT NULL`). | `WHERE tenant_id = $1 AND plant_id = $2` |
| `plant_contractors` | Tenant-Scoped Relationship | Activation of a contractor for a plant (`tenant_id UUID NOT NULL`). | `WHERE tenant_id = $1 AND plant_id = $2 AND contractor_id = $3` |
| `plant_document_requirements` | Tenant-Scoped Rules | Plant-specific document requirements (`tenant_id UUID NOT NULL`). | `WHERE tenant_id = $1 AND plant_id = $2` |
| `access_logs` | Tenant-Scoped Audit | Immutable log of check-in/out events (`tenant_id UUID NOT NULL`). | `WHERE tenant_id = $1 AND ...` |
| `contractors` | **Global** Entity | Contractors work across multiple plants/tenants. Linked via `plant_contractors`. | Global table, but accessed in tenant context via `plant_contractors`. |
| `users` (workers/guards) | Hybrid Entity | Workers belong to a contractor, but have `tenant_id` set or NULL for global. | `WHERE id = $1 AND (tenant_id = $2 OR tenant_id IS NULL)` |
| `documents` | Global to Worker/Contractor | Uploaded files owned by `user_id` or `contractor_id`. Evaluated against plant requirements. | Scoped to worker/contractor; compliance evaluated per plant rule. |
| `document_types` | System / Tenant Catalog | Global system types (`tenant_id IS NULL`) + custom tenant types. | `WHERE (tenant_id IS NULL OR tenant_id = $1)` |

---

### 2.2 Analysis of `compliance_repository.go`

`compliance_repository.go` currently implements four queries:
1. `GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error)`
2. `GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error)`
3. `GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error)`
4. `GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error)`

#### Observations & Remediation Points:

1. **Parameter Ordering in `GetPlantDocumentRequirements`**:
   - *Current SQL*:
     ```sql
     WHERE pdr.plant_id = $1 AND pdr.tenant_id = $2 AND dt.is_active = TRUE
     ```
   - *Violation of DEVELOPMENT_RULES.md Section 3*: "Todo query a estas tablas DEBE incluir `WHERE tenant_id = $1` como primera condición."
   - *Remediation*:
     ```sql
     WHERE pdr.tenant_id = $1 AND pdr.plant_id = $2 AND dt.is_active = TRUE
     ```
   - Swap parameters to pass `tenantID` as `$1` and `plantID` as `$2`.

2. **Tenant Scoping on `GetWorkerAffiliation`**:
   - *Current Implementation*: Currently queries `users` and joins `contractor_workers`, `contractors`, and `plant_contractors` with `plant_id = $1` and `user_id = $2`, but does not verify `tenant_id`.
   - *Risk*: If a guard passes an arbitrary worker user ID, a cross-tenant user could be retrieved if the contractor is shared.
   - *Remediation*: Extract `tenantID` from `ctx` (via `tenant.FromContext(ctx)`) or accept `tenantID`, and filter `pc.tenant_id = $1`:
     ```sql
     SELECT 
         c.id AS contractor_id,
         c.name AS contractor_name,
         c.rfc AS contractor_rfc,
         c.repse_number AS repse_number,
         cw.curp,
         cw.imss_number,
         cw.job_title,
         COALESCE(pc.status::text, 'pending') AS plant_status
     FROM contractor_workers cw
     JOIN contractors c ON cw.contractor_id = c.id
     LEFT JOIN plant_contractors pc ON pc.contractor_id = c.id AND pc.plant_id = $1 AND pc.tenant_id = $2
     WHERE cw.user_id = $3 AND cw.deleted_at IS NULL AND c.deleted_at IS NULL
     LIMIT 1
     ```

3. **Document Deduplication & Multi-Document Selection**:
   - When a worker uploads multiple documents of the same type (e.g. an old expired DC-3 and a newly approved DC-3):
   - Current query sorts by `ORDER BY d.created_at DESC`.
   - In `check_in.go`, the first encountered document is mapped:
     ```go
     if _, exists := docsByCode[d.DocumentTypeCode]; !exists {
         docsByCode[d.DocumentTypeCode] = d
     }
     ```
   - *Corner Case*: If a user uploaded an unapproved or rejected revision *after* an approved one, sorting only by `created_at DESC` might shadow an approved valid document, or vice-versa.
   - *Recommended Query Optimization*:
     ```sql
     SELECT DISTINCT ON (dt.code)
         d.id, dt.code, dt.name, dt.applies_to, d.status::text, d.expires_at, d.issued_at
     FROM documents d
     JOIN document_types dt ON d.document_type_id = dt.id
     WHERE d.user_id = $1 AND d.deleted_at IS NULL
     ORDER BY dt.code, 
              (CASE WHEN d.status = 'approved' THEN 1 ELSE 2 END), 
              d.expires_at DESC NULLS LAST, 
              d.created_at DESC
     ```
     This deterministic query prioritizes approved, unexpired documents, ensuring workers with valid renewed certificates are never falsely denied entry.

---

### 2.3 Analysis of `access_log_repository.go`

`access_log_repository.go` currently implements:
1. `Create(ctx context.Context, log *entity.AccessLog) error`
2. `GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error)`
3. `ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error)`
4. `ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error)`
5. `CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error)`

#### Observations & Missing Capabilities:

1. **Missing `ListWithFilters` Query**:
   - The dispatch and HTTP handler requirements call for multi-criteria filtering:
     - `plant_id` (optional UUID)
     - `access_point_id` (optional UUID)
     - `worker_user_id` (optional UUID)
     - `result` (optional enum: `granted`, `granted_with_warning`, `denied`)
     - `from` (optional timestamp)
     - `to` (optional timestamp)
     - `page` / `per_page` (pagination)
   - *Proposed Architecture*: Add `AccessLogFilter` struct and `ListWithFilters(ctx context.Context, filter AccessLogFilter) ([]entity.AccessLog, int, error)` to `repository.AccessLogRepository`.
   - Query construction must build parameter indices (`$1`, `$2`, ...) dynamically without string interpolation to maintain SQL injection immunity.
   - `WHERE tenant_id = $1` must be strictly hardcoded as condition 1.

2. **Check-Out Plant ID Handling**:
   - In `CheckOutLatest`:
     ```sql
     WHERE tenant_id = $2 AND plant_id = $3 AND user_id = $4 AND checked_out_at IS NULL
     ```
   - When a guard records check-out at an exit gate, `plant_id` might not be provided in the request payload (or might be `uuid.Nil`).
   - If `plantID == uuid.Nil`, the query fails to find the active check-in.
   - *Fix*: Allow `plant_id` to be optional or check active entries across the tenant:
     ```sql
     WHERE tenant_id = $1 
       AND ($2::uuid IS NULL OR plant_id = $2) 
       AND user_id = $3 
       AND checked_out_at IS NULL
     ```

3. **Fail-Closed Tenant Scoping**:
   - In `GetByID`, `tenant.FromContext(ctx)` is checked.
   - In `ListByTenant`, `ListByPlant`, and `CheckOutLatest`, `tenantID` is passed as an argument.
   - *Defense-in-depth enhancement*: Compare argument `tenantID` with `tenant.FromContext(ctx)`. If both exist and differ, fail closed immediately with `domainErrors.ErrTenantNotFound`.

---

### 2.4 Analysis of `contractor_repository.go` (Architectural Requirement)

Clean Architecture requires clear domain boundaries. Currently, contractor affiliation queries are blended into `compliance_repository.go`. A dedicated `ContractorRepository` interface and implementation provides:
1. `GetByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error)`
2. `GetPlantContractor(ctx context.Context, tenantID, plantID, contractorID uuid.UUID) (*entity.PlantContractor, error)`
3. `GetWorkerContractorAffiliation(ctx context.Context, tenantID, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error)`
4. `ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.Contractor, error)`

#### Proposed SQL Queries for `contractor_repository.go`:

```sql
-- 1. GetByID (Contractors are global, but check deleted_at)
SELECT id, name, rfc, repse_number, contractor_type, admin_user_id, is_active, created_at, updated_at, deleted_at
FROM contractors
WHERE id = $1 AND deleted_at IS NULL;

-- 2. GetPlantContractor (Tenant-scoped relationship)
SELECT id, plant_id, tenant_id, contractor_id, status, invited_by, invited_at, activated_at, created_at, updated_at
FROM plant_contractors
WHERE tenant_id = $1 AND plant_id = $2 AND contractor_id = $3;

-- 3. GetWorkerContractorAffiliation (Joins contractor_workers with plant_contractors)
SELECT 
    cw.user_id,
    c.id AS contractor_id,
    c.name AS contractor_name,
    c.rfc AS contractor_rfc,
    c.repse_number,
    cw.curp,
    cw.imss_number,
    cw.job_title,
    cw.is_active AS worker_is_active,
    c.is_active AS contractor_is_active,
    COALESCE(pc.status::text, 'pending') AS plant_status
FROM contractor_workers cw
JOIN contractors c ON cw.contractor_id = c.id
LEFT JOIN plant_contractors pc ON pc.contractor_id = c.id AND pc.tenant_id = $1 AND pc.plant_id = $2
WHERE cw.user_id = $3 AND cw.deleted_at IS NULL AND c.deleted_at IS NULL;
```

---

## 3. Verification of JSONB `document_snapshot` Scanning and Serialization

### 3.1 Domain Implementation Review

In `internal/core/domain/entity/document_snapshot.go`:
- `DocumentSnapshotItem`:
  ```go
  type DocumentSnapshotItem struct {
      DocumentTypeCode string  `json:"document_type_code"`
      DocumentName     string  `json:"document_name"`
      AppliesTo        string  `json:"applies_to"`
      Status           string  `json:"status"`
      IsRequired       bool    `json:"is_required"`
      IsValid          bool    `json:"is_valid"`
      ExpiresAt        *string `json:"expires_at"`
      FailureReason    *string `json:"failure_reason"`
  }
  ```
- `DocumentSnapshot`: `[]DocumentSnapshotItem`
- Implements:
  - `database/sql/driver.Valuer`: `Value() (driver.Value, error)`
    - Empty/nil returns `"[]"`.
    - Marshals slice to JSON string.
  - `database/sql.Scanner`: `Scan(src interface{}) error`
    - Accepts `[]byte`, `string`, or `nil`.
    - Handles empty inputs by resetting to `DocumentSnapshot{}`.
    - Deserializes JSON into `[]DocumentSnapshotItem`.

### 3.2 Unit Test Verification

In `internal/core/domain/entity/document_snapshot_test.go`:
All 6 test cases pass with 100% success:
1. `Value serialization`: asserts JSON contains type codes and names.
2. `Scan from byte slice`: asserts PostgreSQL `[]byte` JSONB representation scans correctly.
3. `Scan from string`: asserts string inputs parse accurately.
4. `Nil Value produces empty JSON array string`: asserts `[]` representation for SQL NOT NULL constraint.
5. `Scan from nil`: returns empty snapshot without error.
6. `Scan unsupported type produces error`: safely rejects non-byte/string types (e.g. integer).

### 3.3 Optimization in `access_log_repository.go`

Currently, `access_log_repository.go` manually calls `json.Marshal(log.DocumentSnapshot)` before `ExecContext` and manually scans into `[]byte` before `json.Unmarshal`.

Because `DocumentSnapshot` already implements `driver.Valuer` and `sql.Scanner`, `access_log_repository.go` should leverage them directly:

#### In `Create`:
```go
// Replace manual marshal:
// snapshotBytes, err := json.Marshal(log.DocumentSnapshot)
// With direct pass-through:
_, err = r.db.ExecContext(
    ctx,
    query,
    log.ID,
    log.TenantID,
    log.PlantID,
    log.AccessPointID,
    log.UserID,
    log.ContractorID,
    log.ProcessedBy,
    log.CheckedInAt,
    log.CheckedOutAt,
    string(log.Result),
    log.DenialReason,
    log.DocumentSnapshot, // driver.Valuer will be invoked automatically!
    log.Notes,
    log.CreatedAt,
)
```

#### In `scanRow`:
```go
// Replace manual unmarshal:
err := row.Scan(
    &log.ID,
    &log.TenantID,
    &log.PlantID,
    &log.AccessPointID,
    &log.UserID,
    &contractorID,
    &processedBy,
    &log.CheckedInAt,
    &checkedOutAt,
    &resultStr,
    &denialReason,
    &log.DocumentSnapshot, // sql.Scanner will be invoked automatically!
    &notes,
    &log.CreatedAt,
)
```
This guarantees that `driver.Valuer` and `sql.Scanner` are actively tested and exercised in production queries.

---

## 4. Integration and Unit Test Specifications

### 4.1 Current Test Coverage State

Running `go test -v -cover ./...` in `asesamo-api` shows:
- `internal/core/usecase/access`: **98.5% statement coverage** (exceeds >= 80% requirement).
- `internal/core/usecase/auth`: **100.0% statement coverage**.
- `internal/core/domain/entity`: **81.8% statement coverage**.
- `pkg/tenant`: **100.0% statement coverage**.
- `pkg/jwt`: **91.2% statement coverage**.
- `internal/adapter/postgres`: **5.5% statement coverage** (only tenant fail-closed tests with nil DB).

### 4.2 Check-In & Check-Out Unit Test Matrix (7 Core Scenarios)

The test suite in `internal/core/usecase/access/check_in_test.go` and `check_out_test.go` exhaustively covers all 7 required scenarios:

| # | Scenario | Inputs & Preconditions | Expected Outcome | Denial Reason / Warning | Test Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **1** | **Compliant worker** | Active worker, active contractor, approved DC-3 (valid for 180d). | `result = granted`, `HTTP 200` | `denial_reason = nil` | **PASS** (`TestCheckInUseCase_Execute/compliant_worker_is_granted_access`) |
| **2** | **Inactive worker** | Worker affiliation has `IsActive = false`. | `result = denied`, `HTTP 200` | `"El perfil del trabajador se encuentra inactivo."` | **PASS** (`TestCheckInUseCase_Execute/inactive_worker_affiliation_is_denied`) |
| **3** | **Blocked contractor** | Contractor has `PlantStatus = 'blocked'`. | `result = denied`, `HTTP 200` | `"La empresa contratista '...' se encuentra blocked en esta planta."` | **PASS** (`TestCheckInUseCase_Execute/blocked_contractor_in_plant_is_denied`) |
| **4** | **Expired doc (`block_on_expired_docs = true`)** | DC-3 expired 30 days ago, plant `block_on_expired_docs = true`. | `result = denied`, `HTTP 200` | `"Documento 'Constancia DC-3' vencido el ..."` | **PASS** (`TestCheckInUseCase_Execute/expired_document_with_block_on_expired_docs_=_true_is_denied`) |
| **5** | **Expiring soon doc** | DC-3 expires in 15 days (< 30d window), no fatal failure. | `result = granted_with_warning`, `HTTP 200` | `denial_reason = nil`, snapshot flagged | **PASS** (`TestCheckInUseCase_Execute/document_expiring_within_30_days_yields_warning_without_fatal_failure`) |
| **6** | **Check-out flow** | Worker has active check-in with `checked_out_at = nil`. | `HTTP 200`, `checked_out_at` updated to `now()`. | Duration calculated, log updated | **PASS** (`TestCheckOutUseCase_Execute/successful_check-out_with_existing_checked_out_at`) |
| **7** | **Access log listing with pagination** | Plant has 105 logs, query `page=2&per_page=150`. | `HTTP 200`, clamped to 100 per_page, total = 105. | Paginated JSON envelope | **PASS** (`TestListAccessLogsUseCase_Execute/list_by_plant_with_per_page_capped_at_100`) |

Additional boundary tests already passing:
- Missing mandatory document -> `result = denied` ("Documento '...' faltante.")
- Unapproved document status (pending review) -> `result = denied` ("Documento '...' no aprobado (estatus: pending_review).")
- Expired document with `block_on_expired_docs = false` -> `result = granted_with_warning`
- Access point tenant mismatch -> `ErrAccessPointNotFound` (404)
- Malformed / Nil UUIDs -> `ErrInvalidInput` (400)
- Database creation error -> Wrapped error propagation (500)

---

### 4.3 Proposed PostgreSQL Integration Test Suite (`//go:build integration`)

To raise repository coverage and verify live SQL execution, create `asesamo-api/internal/adapter/postgres/access_log_repository_integration_test.go` and `compliance_repository_integration_test.go`:

#### Specification:

```go
//go:build integration

package postgres_test

import (
    "context"
    "os"
    "testing"
    "time"

    "asesamo-api/internal/adapter/postgres"
    "asesamo-api/internal/core/domain/entity"
    "asesamo-api/internal/core/domain/repository"
    "asesamo-api/internal/infra/database"
    "asesamo-api/pkg/tenant"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Deterministic seed constants from db/seeds/000001_seed_dev_data.sql
var (
    seedTenantID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
    seedPlantID    = uuid.MustParse("22222222-2222-2222-2222-222222222222")
    seedCasetaID   = uuid.MustParse("33333333-3333-3333-3333-333333333331")
    seedGuardID    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
    seedContractor = uuid.MustParse("55555555-5555-5555-5555-555555555555")
    seedWorker1ID  = uuid.MustParse("66666666-6666-6666-6666-666666666661")
    seedWorker2ID  = uuid.MustParse("66666666-6666-6666-6666-666666666662")
    seedWorker3ID  = uuid.MustParse("66666666-6666-6666-6666-666666666663")
)

func getTestDB(t *testing.T) *database.PostgresPool {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable"
    }
    pool, err := database.NewPostgresPool(dbURL, database.DefaultPoolConfig())
    require.NoError(t, err, "Integration test requires running PostgreSQL with applied seeds")
    return pool
}

func TestIntegration_AccessLogRepository_Lifecycle(t *testing.T) {
    pool := getTestDB(t)
    defer pool.Close()
    repo := postgres.NewAccessLogRepository(pool.DB)

    ctx := tenant.WithTenant(context.Background(), seedTenantID)

    // 1. Create access log with JSONB DocumentSnapshot
    expDate := "2027-08-01"
    snapshot := entity.DocumentSnapshot{
        {
            DocumentTypeCode: "DC3",
            DocumentName:     "Constancia de Capacitación DC-3",
            AppliesTo:        "worker",
            Status:           "approved",
            IsRequired:       true,
            IsValid:          true,
            ExpiresAt:        &expDate,
        },
    }

    logID := uuid.New()
    newLog := &entity.AccessLog{
        ID:               logID,
        TenantID:         seedTenantID,
        PlantID:          seedPlantID,
        AccessPointID:    seedCasetaID,
        UserID:           seedWorker1ID,
        ContractorID:     &seedContractor,
        ProcessedBy:      &seedGuardID,
        CheckedInAt:      time.Now().UTC(),
        Result:           entity.AccessResultGranted,
        DocumentSnapshot: snapshot,
        CreatedAt:        time.Now().UTC(),
    }

    err := repo.Create(ctx, newLog)
    require.NoError(t, err)

    // 2. Retrieve by ID and verify JSONB scanner
    retrieved, err := repo.GetByID(ctx, logID)
    require.NoError(t, err)
    assert.Equal(t, logID, retrieved.ID)
    assert.Equal(t, entity.AccessResultGranted, retrieved.Result)
    require.Len(t, retrieved.DocumentSnapshot, 1)
    assert.Equal(t, "DC3", retrieved.DocumentSnapshot[0].DocumentTypeCode)
    assert.True(t, retrieved.DocumentSnapshot[0].IsValid)

    // 3. Test Check-Out
    checkOutTime := time.Now().UTC().Add(45 * time.Minute)
    checkedOut, err := repo.CheckOutLatest(ctx, seedTenantID, seedPlantID, seedWorker1ID, checkOutTime)
    require.NoError(t, err)
    assert.Equal(t, logID, checkedOut.ID)
    require.NotNil(t, checkedOut.CheckedOutAt)

    // 4. Test Multi-Tenant Isolation
    foreignTenantID := uuid.New()
    foreignCtx := tenant.WithTenant(context.Background(), foreignTenantID)
    _, err = repo.GetByID(foreignCtx, logID)
    assert.Error(t, err, "Accessing log with foreign tenant context MUST fail closed")
}
```

---

## 5. Summary of Proposed Changes & Recommendations

1. **`compliance_repository.go`**:
   - Fix query parameter ordering on `GetPlantDocumentRequirements`: use `WHERE pdr.tenant_id = $1 AND pdr.plant_id = $2`.
   - Implement `DISTINCT ON (dt.code)` sorting on `GetWorkerDocuments` and `GetContractorDocuments` prioritizing approved certificates over stale rejected uploads.
2. **`access_log_repository.go`**:
   - Streamline `Create` and `scanRow` to use `entity.DocumentSnapshot`'s `driver.Valuer` and `sql.Scanner` directly.
   - Add `ListWithFilters(ctx context.Context, filter AccessLogFilter) ([]entity.AccessLog, int, error)` with parameterized query building.
   - Support optional `plant_id` in `CheckOutLatest`.
3. **`contractor_repository.go`**:
   - Add dedicated contractor repository implementing `GetByID`, `GetPlantContractor`, and `GetWorkerContractorAffiliation` to satisfy Clean Architecture separation.
4. **Integration Testing**:
   - Provide integration test files with build tag `//go:build integration` that execute against Docker DB using `make test-integration`.
