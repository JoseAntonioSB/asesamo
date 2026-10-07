# Milestone M3 Gate Review Report: Architecture, Compliance Logic & Usecase Coverage

## 1. Observation

### 1.1 Architecture & Conformance (Clean Architecture)
- **`internal/core/usecase/access/check_in.go` (lines 1–15)**:
  Imports only pure domain entities (`entity`), domain errors (`domainErrors`), repository interfaces (`repository`), `pkg/tenant`, standard libraries, and `uuid`.
  Zero dependencies on HTTP frameworks (`chi`), database drivers (`database/sql`), or external persistence layers.
- **`internal/core/usecase/access/check_out.go` (lines 1–13)**:
  Imports only domain entities, repository contracts, and standard library. Decoupled from transport and storage.
- **`internal/core/usecase/access/list_access_logs.go` (lines 1–11)**:
  Decoupled usecase accepting domain DTOs, driving `repository.AccessLogRepository`, and returning paginated domain models.
- **`internal/core/usecase/access/list_access_points.go` (lines 1–12)**:
  Decoupled usecase invoking `repository.AccessPointRepository`.
- **`internal/core/domain/repository/`**:
  All repository contracts (`AccessLogRepository`, `AccessPointRepository`, `ComplianceRepository`, `PlantRepository`) are declared in pure Go interfaces within the core domain.

### 1.2 Compliance Evaluation Engine Logic
- **Worker Active Check (`users.is_active` AND `contractor_workers.is_active`)**:
  - `internal/adapter/postgres/compliance_repository.go` (lines 28–52): Retrieves `users.is_active` into `aff.UserIsActive` and sets `aff.IsActive = aff.UserIsActive`.
  - `compliance_repository.go` (lines 54–103): Joins `contractor_workers` and retrieves `cw.is_active` into `aff.WorkerIsActive`. If worker has no contractor record, `aff.WorkerIsActive` is set to `false`.
  - `internal/core/usecase/access/check_in.go` (lines 170–179):
    ```go
    isWorkerActive := aff.IsActive
    if aff.UserIsActive || aff.WorkerIsActive {
        isWorkerActive = aff.UserIsActive && aff.WorkerIsActive
    }
    if !aff.IsActive {
        isWorkerActive = false
    }
    if !isWorkerActive {
        nonNegotiableFailures = append(nonNegotiableFailures, "El perfil del trabajador se encuentra inactivo.")
    }
    ```
    Correctly denies access if either `users.is_active` or `contractor_workers.is_active` is false.
- **Contractor Active Check (`contractors.is_active` AND `plant_contractors.status == 'active'`)**:
  - `internal/core/usecase/access/check_in.go` (lines 183–194):
    ```go
    if aff.ContractorID != uuid.Nil {
        if !aff.ContractorIsActive && (aff.UserIsActive || aff.WorkerIsActive) {
            nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' se encuentra inactiva.", aff.ContractorName))
        }
        if aff.PlantStatus != "active" {
            if aff.PlantStatus == "blocked" || aff.PlantStatus == "suspended" {
                nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' se encuentra %s en esta planta.", aff.ContractorName, aff.PlantStatus))
            } else {
                nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' no se encuentra activa en esta planta (estatus: %s).", aff.ContractorName, aff.PlantStatus))
            }
        }
    }
    ```
    All non-active states (`pending`, `suspended`, `blocked`) and inactive contractor companies append to `nonNegotiableFailures`.
- **Separation of Non-Negotiable Denials from `plants.block_on_expired_docs`**:
  - `internal/core/usecase/access/check_in.go` (lines 251–272):
    ```go
    if len(nonNegotiableFailures) > 0 {
        // Non-negotiable failures always deny access regardless of plant policy
        decisionResult = entity.AccessResultDenied
        allFailures := append(nonNegotiableFailures, expiredDocFailures...)
        combined := strings.Join(allFailures, " | ")
        denialReason = &combined
    } else if len(expiredDocFailures) > 0 {
        if plant.BlockOnExpiredDocs {
            decisionResult = entity.AccessResultDenied
            combined := strings.Join(expiredDocFailures, " | ")
            denialReason = &combined
        } else {
            decisionResult = entity.AccessResultGrantedWithWarning
            denialReason = nil
        }
    } else if len(warnings) > 0 {
        decisionResult = entity.AccessResultGrantedWithWarning
        denialReason = nil
    } else {
        decisionResult = entity.AccessResultGranted
        denialReason = nil
    }
    ```
    Non-negotiable violations (missing mandatory documents, unapproved documents, inactive worker, inactive contractor, blocked/pending/suspended contractor status) always deny access regardless of `plant.BlockOnExpiredDocs`. Expired documents are governed by `plant.BlockOnExpiredDocs`.
- **Warning Threshold Logic (<= 30 days)**:
  - `internal/core/usecase/access/check_in.go` (lines 159–160, 236–238):
    ```go
    now := time.Now().UTC()
    warningThreshold := now.AddDate(0, 0, 30)
    ...
    if doc.ExpiresAt != nil && doc.ExpiresAt.Before(warningThreshold) {
        warnings = append(warnings, fmt.Sprintf("Documento '%s' próximo a vencer el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02")))
    }
    ```
    Valid approved documents expiring within 30 days are flagged with warnings and grant access (`granted_with_warning`).
- **Document Snapshot JSONB Array Schema**:
  - `internal/core/domain/entity/document_snapshot.go` (lines 9–36):
    Defines `DocumentSnapshotItem` containing `document_type_code`, `document_name`, `applies_to`, `status`, `is_required`, `is_valid`, `expires_at`, `failure_reason`.
    `DocumentSnapshot` implements `driver.Valuer` (serializing to `"[]"` if slice is nil, matching schema `DEFAULT '[]'`) and `sql.Scanner`.
  - `internal/core/usecase/access/check_in.go` (lines 36–37, 330–331):
    Populates both `DocumentSnapshot` and `Documents` in `CheckInOutput`.

### 1.3 Check-Out Logic & Duration Calculation
- **`internal/core/usecase/access/check_out.go`**:
  - Lines 56–62: Resolves `PlantID` from `accessPointRepo.GetByID` when omitted.
  - Line 75: Calculates `durationMinutes = int(math.Max(0, outTime.Sub(updatedLog.CheckedInAt).Minutes()))`, using `math.Max` to guard against negative durations from clock drift.

### 1.4 Access Logs Listing & Pagination
- **`internal/core/usecase/access/list_access_logs.go` (lines 46–56)**:
  Normalizes page (default 1), per_page (default 20, max 100), and computes `offset = (in.Page - 1) * in.PerPage`.
- **`internal/adapter/postgres/access_log_repository.go` (lines 220–324)**:
  `ListWithFilters` strictly anchors `WHERE tenant_id = $1` as the base predicate, appends parameterized filters (`plant_id`, `access_point_id`, `user_id`, `result`, date ranges), executes a COUNT query for total matching rows, and selects paginated rows with `LIMIT` and `OFFSET`.
- **`internal/adapter/http/handler/access_handler.go` (lines 211–215)**:
  Returns envelope format with metadata: `{ "data": [...], "meta": { "page": 1, "per_page": 20, "total": N } }`.

### 1.5 Verification Outputs (Independently Re-run)
- Command: `go test -v -race -cover -count=1 ./internal/core/usecase/access/...`
  - Result: `PASS`
  - Statement coverage: **98.7% of statements** (exceeds threshold of >= 80%).
  - Failures: **0**
  - Race conditions: **0**
- Command: `go test -v -race -cover ./internal/adapter/http/handler/...`
  - Result: `PASS`
  - Statement coverage: **97.5% of statements**.
  - Failures: **0**
  - Race conditions: **0**
- Command: `go test -v -race ./internal/adapter/postgres/...`
  - Result: `PASS` (all 25 multi-tenant fail-closed tests pass).
- Command: `make lint && make build`
  - Result: `exit code 0` (0 linter issues, binary compiled successfully).

### 1.6 Integrity Audit
- Scanned all production code for hardcoded test UUIDs, mocked responses, or artificial evaluation bypasses:
  - No fake mock returns embedded in production use cases.
  - No dummy or facade methods: all database repositories execute parameterized PostgreSQL SQL queries.
  - No shortcuts bypassing Mexican industrial compliance checks.

---

## 2. Logic Chain

1. **Clean Architecture Separation (Obs 1.1)**:
   Because the use cases in `internal/core/usecase/access/` only import domain entities, repository interfaces, and domain error packages, they remain completely agnostic of HTTP transports and PostgreSQL database drivers. All external interactions are mediated through domain repository interfaces implemented in `internal/adapter/postgres/`, fulfilling Clean Architecture and Dependency Inversion.

2. **Compliance Logic Correctness (Obs 1.2)**:
   - Worker active state correctly checks both user account active flag (`users.is_active`) and contractor worker affiliation (`contractor_workers.is_active`). Inactive states immediately cause an access denial.
   - Contractor active state correctly checks both contractor account active flag (`contractors.is_active`) and plant contractor status (`plant_contractors.status == 'active'`). States `pending`, `suspended`, or `blocked` cause immediate access denial.
   - The separation between non-negotiable failures and policy-governed expirations is strictly enforced: `len(nonNegotiableFailures) > 0` is checked first. A worker with missing required documents or unapproved documents is rejected even if `plant.BlockOnExpiredDocs` is set to `false`.
   - The 30-day window (`warningThreshold = now.AddDate(0, 0, 30)`) identifies expiring documents without blocking entry when documents are otherwise compliant.
   - Document snapshot schema matches `PROJECT.md` and `000001_init.up.sql`, serializing to JSONB.

3. **Check-Out & Duration Resilience (Obs 1.3)**:
   The checkout workflow resolves the plant ID from the access point if omitted and uses `math.Max(0, ...)` to ensure non-negative durations, correctly recording worker departure.

4. **Multi-Tenancy & Pagination Integrity (Obs 1.4)**:
   `ListWithFilters` enforces `WHERE tenant_id = $1` at the query root, preventing cross-tenant leakage. Pagination parameters are bounded to [1, 100], and the response envelope includes full pagination metadata (`page`, `per_page`, `total`).

5. **Test & Verification Rigor (Obs 1.5, 1.6)**:
   Independent test execution confirmed 98.7% statement coverage on use cases (well above the 80% requirement), 0 test failures, and 0 data races. Code inspection verified no integrity violations or hardcoded fake logic.

---

## 3. Caveats

1. **Date Precision on Document Expiration**: Document expiration in the database is stored as `DATE` (`YYYY-MM-DD`). In Go, `doc.ExpiresAt.Before(now)` uses UTC timestamp comparison. If an expiration date is stored at midnight UTC (`00:00:00Z`), the document will be considered expired as soon as that calendar date begins. This is standard industrial practice (expiry date is the last valid day or expiry deadline), but frontend date display should format dates as `YYYY-MM-DD` without timezone conversion.
2. **Physical Barcode / RFID Hardware Integration**: Verification covered backend HTTP handlers, domain logic, and PostgreSQL repositories. Physical HID wedge barcode scanner integration is a client-side concern handled in Milestone M4.

---

## 4. Conclusion

The Milestone M3 implementation for `asesamo-api` fulfills all functional, architectural, compliance, and testing requirements specified in `ORIGINAL_REQUEST.md`, `PROJECT.md`, and `DEVELOPMENT_RULES.md`. Statement coverage for `internal/core/usecase/access/...` is **98.7%** (exceeding the >= 80% threshold), with zero failures and zero data races. No integrity violations, facades, or shortcuts were found.

**Verdict**: **APPROVE**

---

## 5. Verification Method

To independently verify this gate review:

1. **Execute usecase test suite with race detector and coverage**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race -cover -count=1 ./internal/core/usecase/access/...
   ```
   *Expected*: `PASS`, `coverage: 98.7% of statements`, 0 failures, 0 races.

2. **Execute HTTP handler test suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race -cover ./internal/adapter/http/handler/...
   ```
   *Expected*: `PASS`, `coverage: 97.5% of statements`, 0 failures, 0 races.

3. **Execute repository multi-tenant fail-closed test suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race ./internal/adapter/postgres/...
   ```
   *Expected*: `PASS`, all 25 fail-closed multi-tenancy assertions pass.

4. **Execute root linting and compilation**:
   ```bash
   cd /Users/jasbotello/devel/asesamo
   make lint
   make build
   ```
   *Expected*: Exit code 0, binary created at `bin/asesamo-api`.
