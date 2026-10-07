# Handoff Report: Milestone M3 — Compliance Evaluation Engine & JSONB Snapshots

**Agent**: `teamwork_preview_explorer_m3_1`  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m3_1`  
**Handoff Type**: Hard (Investigation & Analysis Complete)  
**Recipient**: Parent Agent (`ccf64bbc-f322-4ca8-b0be-0ee59db22a99`)  
**Associated Analysis**: `analysis.md`

---

## 1. Observation

Direct code and database observations verified during the investigation:

1. **Database Schema (`db/migrations/000001_init.up.sql`)**:
   - `plants` (lines 59-60): `block_on_expired_docs BOOLEAN NOT NULL DEFAULT TRUE`.
   - `users` (line 122): `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
   - `contractors` (line 170): `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
   - `plant_contractors` (lines 188, 195): `status plant_contractor_status NOT NULL DEFAULT 'pending'` with enum values `'pending'`, `'active'`, `'suspended'`, `'blocked'`.
   - `contractor_workers` (line 224): `is_active BOOLEAN NOT NULL DEFAULT TRUE`.
   - `plant_document_requirements` (lines 284-294): joins `document_types` to specify mandatory documents (`is_required BOOLEAN NOT NULL DEFAULT TRUE`) and custom `validity_days`.
   - `access_logs` (lines 369-395): Immutable audit table with columns:
     `id` (UUID), `tenant_id` (UUID), `plant_id` (UUID), `access_point_id` (UUID), `user_id` (UUID), `contractor_id` (UUID), `processed_by` (UUID), `checked_in_at` (TIMESTAMPTZ), `checked_out_at` (TIMESTAMPTZ), `result` (`granted`, `granted_with_warning`, `denied`), `denial_reason` (TEXT), `document_snapshot` (`JSONB NOT NULL DEFAULT '[]'`), `notes` (TEXT), `created_at` (TIMESTAMPTZ). There is **no** `updated_at` column.

2. **Repository Implementation (`asesamo-api/internal/adapter/postgres/compliance_repository.go`)**:
   - Lines 28-44: User query checks `users.is_active` as `aff.IsActive`.
   - Lines 54-68: Affiliation query selects `c.id`, `c.name`, `c.rfc`, `c.repse_number`, `cw.curp`, `cw.imss_number`, `cw.job_title`, and `COALESCE(pc.status::text, 'pending') AS plant_status`.
   - **Crucial Observation**: Neither `cw.is_active` (`contractor_workers.is_active`) nor `c.is_active` (`contractors.is_active`) are selected or scanned.

3. **Check-In Use Case Logic (`asesamo-api/internal/core/usecase/access/check_in.go`)**:
   - Lines 170-173:
     ```go
     if !aff.IsActive {
         hasFatalFailure = true
         failures = append(failures, "El perfil del trabajador se encuentra inactivo.")
     }
     ```
   - Lines 176-179:
     ```go
     if aff.PlantStatus == "blocked" || aff.PlantStatus == "suspended" {
         hasFatalFailure = true
         failures = append(failures, fmt.Sprintf("La empresa contratista '%s' se encuentra %s en esta planta.", aff.ContractorName, aff.PlantStatus))
     }
     ```
     **Crucial Observation**: If `aff.PlantStatus == "pending"`, it is not flagged as a fatal failure, permitting access even when contractor approval is pending.
   - Lines 240-249:
     ```go
     if hasFatalFailure {
         if plant.BlockOnExpiredDocs {
             decisionResult = entity.AccessResultDenied
             combined := strings.Join(failures, " | ")
             denialReason = &combined
         } else {
             decisionResult = entity.AccessResultGrantedWithWarning
             combined := strings.Join(failures, " | ")
             denialReason = &combined
         }
     }
     ```
     **Crucial Observation**: If `plant.BlockOnExpiredDocs == false`, **any** failure (including missing documents, rejected documents, inactive worker, or blocked contractor) is demoted to `granted_with_warning`.

4. **Test Suite Verification (`go test ./...`)**:
   - Executed `go test ./...` in `/Users/jasbotello/devel/asesamo/asesamo-api`.
   - Result: Exited with code 0 across all 15 packages.
   - Existing unit tests in `check_in_test.go` test current mocks, but do not assert the boundary where `BlockOnExpiredDocs == false` should still deny missing documents.

5. **E2E Contract Requirements (`PROJECT.md` & `test/e2e/tier1_feature_test.sh`)**:
   - `tier1_feature_test.sh` line 189: `assert_json_expr "TC 4.3: Document snapshot contains valid DC3/REPSE/SUA entries" "$BODY" '(.data.document_snapshot | type == "array") and (.data.document_snapshot | length > 0)'`.
   - `tier1_feature_test.sh` lines 205-209: Non-compliant worker check-in returns `result == "denied"` and a Spanish `denial_reason` mentioning expired or rejected documents.
   - `tier1_feature_test.sh` lines 225-227: Expiring worker check-in returns `result == "granted_with_warning"`.

---

## 2. Logic Chain

1. **From Observation 1 & 2 to Worker Active Status**:
   - A worker can be deactivated globally (`users.is_active = FALSE`) OR locally by their contractor company (`contractor_workers.is_active = FALSE`).
   - Because `compliance_repository.go` currently only checks `users.is_active`, a dismissed contractor worker whose relationship is deactivated (`contractor_workers.is_active = false`) would be evaluated as active.
   - **Conclusion**: `compliance_repository.go` must select both `users.is_active` and `contractor_workers.is_active`, and `check_in.go` must deny access if either is false.

2. **From Observation 1 & 2 to Contractor Active Status**:
   - Contractors operate under a two-tier authorization model: global company standing (`contractors.is_active`) and plant facility activation (`plant_contractors.status = 'active'`).
   - The current check-in engine only blocks if status is `'blocked'` or `'suspended'`. It allows `'pending'`.
   - A contractor whose relationship is pending approval has not completed compliance checks and must not be admitted into the plant.
   - **Conclusion**: Contractor check must require `contractors.is_active == TRUE` and `plant_contractors.status == 'active'`. Statuses `'pending'`, `'suspended'`, or `'blocked'` must all trigger `denied`.

3. **From Observation 1 & 3 to Plant Policy Separation (`block_on_expired_docs`)**:
   - The purpose of `plants.block_on_expired_docs` is to govern whether workers whose approved documents have elapsed their expiration date can enter under emergency/lenient supervision (`granted_with_warning`).
   - The current code groups missing documents, rejected documents, worker deactivations, and contractor blocks into `hasFatalFailure`, and then converts all of them to `granted_with_warning` if `block_on_expired_docs == false`.
   - This creates a severe security flaw: a stranger with missing documents or an inactive worker could enter an industrial plant if `block_on_expired_docs` is toggled off.
   - **Conclusion**: Failures must be split into two strict categories:
     a) `nonNegotiableDenials`: missing required doc, rejected doc, pending review doc, inactive worker, unauthorized contractor. Always `denied`.
     b) `expiredDocDenials`: approved documents that have expired. Denied if `block_on_expired_docs == true`; warning if `false`.

4. **From Observation 1, 3, & 5 to `document_snapshot` Structure**:
   - The database table defines `access_logs.document_snapshot` as `JSONB NOT NULL DEFAULT '[]'`.
   - `PROJECT.md` and E2E test suites (`tier1_feature_test.sh` TC 4.3) mandate that `.data.document_snapshot` must be a JSON array of `DocumentSnapshotItem` objects.
   - In relational storage, `access_logs` references `user_id` and `contractor_id`.
   - In audit retrieval, the API returns the worker summary, contractor summary, plant summary, and access point summary alongside the `document_snapshot` array.
   - For complete regulatory audit archival, an extended self-contained JSON schema (Worker + Contractor + Plant Policy + Document Snapshot Items) has been formulated and documented in `analysis.md` section 3.4.

---

## 3. Caveats

- **Warning Window Threshold**: The current code sets `warningThreshold := now.AddDate(0, 0, 30)` (30 days). The seed data has Worker 3's DC-3 expiring in +5 days and SUA expiring in +25 days. Both fall within 30 days; 5 days also falls within a 15-day window. If individual plants need customized warning windows in future phases, a `warning_days` column can be added to `plant_document_requirements`.
- **Multiple Document Versions**: In PostgreSQL `compliance_repository.go`, queries order by `created_at DESC` to ensure the most recently uploaded document is evaluated. This assumes contractors/workers upload newer documents to remediate expired or rejected ones.
- **Offline / Edge Gates**: Phase 1 MVP assumes synchronous database access from `asesamo-api`. Edge gate caching (local SQLite/badging) is out of scope for Phase 1.

---

## 4. Conclusion

1. **Check-In Compliance Logic**:
   - Worker active evaluation must check `users.is_active AND contractor_workers.is_active`.
   - Contractor active evaluation must check `contractors.is_active AND plant_contractors.status == 'active'`.
   - Document evaluation must check:
     * Matching requirement by `applies_to` (`worker` vs `contractor`).
     * Status must be `'approved'`.
     * `expires_at > now()`.
     * Expiration within warning threshold triggers `granted_with_warning`.
   - Result determination:
     * `denied`: Any non-negotiable failure (missing, rejected, unapproved doc, inactive worker, unauthorized contractor) OR expired required doc when `block_on_expired_docs = true`.
     * `granted_with_warning`: All required docs approved and unexpired, but 1+ expiring soon; OR required doc expired when `block_on_expired_docs = false` (with no non-negotiable failures).
     * `granted`: All required docs approved, valid, and unexpired with no warnings.

2. **Immutable `access_logs` and JSONB Schema**:
   - `access_logs` entry is write-only upon check-in; only `checked_out_at` is updated at exit.
   - `document_snapshot` is stored as an immutable JSONB array of `DocumentSnapshotItem` objects:
     `[ { document_type_code, document_name, applies_to, status, is_required, is_valid, expires_at, failure_reason } ]`.
   - Full audit context is delivered via `CheckInOutput` and can be exported as the comprehensive audit envelope specified in `analysis.md`.

---

## 5. Verification Method

To independently verify the findings and proposed implementation:

1. **Verify Backend Compilation and Unit Tests**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v ./internal/core/usecase/access/...
   ```

2. **Verify Database Schema and Seed Data**:
   Inspect `db/migrations/000001_init.up.sql` lines 369-395 and `db/seeds/000001_seed_dev_data.sql`:
   - Confirm Worker 1 (`66666666-6666-6666-6666-666666666661`) has valid DC-3, INE, Medical -> Expect `granted`.
   - Confirm Worker 2 (`66666666-6666-6666-6666-666666666662`) has expired DC-3 and rejected Medical -> Expect `denied`.
   - Confirm Worker 3 (`66666666-6666-6666-6666-666666666663`) has DC-3 expiring in 5 days -> Expect `granted_with_warning`.

3. **Verify E2E Compliance Tests**:
   ```bash
   cd /Users/jasbotello/devel/asesamo
   bash test/e2e/tier1_feature_test.sh
   bash test/e2e/tier2_boundary_test.sh
   bash test/e2e/tier3_combination_test.sh
   ```

4. **Invalidation Conditions**:
   - If `tier1_feature_test.sh` TC 4.3 fails because `document_snapshot` is an object instead of an array, this invalidates object-based root JSONB schema proposals.
   - If a worker with `contractor_workers.is_active = false` is allowed access, the implementation fails Mexican industrial access compliance standards.
