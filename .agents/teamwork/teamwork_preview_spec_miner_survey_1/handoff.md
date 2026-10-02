# Handoff Report — Backend Specification Survey (Phase 1 MVP)

**Agent:** `teamwork_preview_spec_miner_survey_1`  
**Date:** 2026-10-01  
**Status:** Complete (Hard Handoff)  
**Report File:** `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_1/survey_report.md`

---

## 1. Observation

Direct observations from authoritative project sources:

1. **`ORIGINAL_REQUEST.md` (lines 16-22, 38-43):**
   - R1 mandates a Go backend service (`asesamo-api`) using Clean Architecture and `DEVELOPMENT_RULES.md`.
   - Requires connecting to PostgreSQL and applying `db/migrations/000001_init.up.sql`.
   - Requires authentication endpoint (`POST /api/v1/auth/login`) returning a JWT.
   - Requires access point check-in verification (`POST /api/v1/access/check-in`) receiving a worker UUID and access point ID, validating worker and documents against plant rules, and recording an immutable log in `access_logs` with a JSONB `document_snapshot`.
   - Acceptance criteria require `granted` on valid documents, `denied` (or warning according to policy) with denial reason in `access_logs` on missing/expired documents, and zero failures on `go test ./...`.

2. **`DEVELOPMENT_RULES.md`:**
   - Lines 11-23: Backend stack is Go (`chi/v5`), Clean Architecture, PostgreSQL, `sqlc` (no GORM), `golang-migrate` (`000001_description.up.sql`), JWT stateless (HS256/RS256), Cloudflare R2 via AWS SDK v2, Resend for email.
   - Lines 51-110: Clean Architecture directory layout with `cmd/api/main.go`, `internal/domain`, `internal/usecase`, `internal/adapter/http`, `internal/adapter/postgres`, `internal/infra`, `db/migrations`, `db/queries`, `db/sqlc.yaml`, `pkg/`. Strict rule: handlers never query DB; use cases never import HTTP packages; repos never import use cases.
   - Lines 114-120: Multi-tenancy rules: `tenant_id UUID NOT NULL` on tenant tables; all queries must include `WHERE tenant_id = $1` as first condition; `tenant_id` extracted from JWT in auth middleware into Go context; central helper `tenant.FromContext(ctx)` is the only way to obtain it.
   - Lines 144-149: Code, comments, docstrings in English; final user error messages in Spanish; domain acronyms (REPSE, SUA, DC3, NOM) kept as-is.
   - Lines 172-188: REST API conventions: `/api/v1/...`, kebab-case, success envelope `{"data": ..., "meta": ...}`, error envelope `{"error": {"code": "...", "message": "..."}}`, internal error codes in `UPPER_SNAKE_CASE` in English, error messages in Spanish.
   - Lines 191-205: JWT rules: `JWT_SECRET` min 32 chars; payload `{"sub", "tenant_id", "role", "exp"}`; access token expires in 1 hour; refresh token expires in 7 days (stored hashed in `refresh_tokens`); roles: `super_admin`, `plant_admin`, `ehs_manager`, `guard`, `contractor`.
   - Lines 218-227: Testing rules: Unit tests mock interfaces (`testify/mock`, zero real IO, min 80% coverage for use cases); Integration tests use real PostgreSQL container (`testcontainers-go`), running each test in a transaction with rollback.

3. **`db/migrations/000001_init.up.sql`:**
   - 13 tables defined: `tenants`, `plants`, `access_points`, `users`, `refresh_tokens`, `contractors`, `plant_contractors`, `contractor_workers`, `document_types`, `plant_document_requirements`, `documents`, `access_logs`, `notifications`.
   - `plants.block_on_expired_docs BOOLEAN NOT NULL DEFAULT TRUE` (line 60).
   - `access_points` has `plant_id`, `tenant_id`, `is_active` (lines 77-87).
   - `users.role` enum: `'super_admin'`, `'plant_admin'`, `'ehs_manager'`, `'guard'`, `'contractor_admin'`, `'worker'` (lines 101-108).
   - `plant_contractors.status` enum: `'pending'`, `'active'`, `'suspended'`, `'blocked'` (lines 188-195).
   - `document_types`: seeded with DC3 (worker, 365d), REPSE (contractor, 365d), SUA (contractor, 30d), SAT_32D (contractor, 30d), INE (worker, no expiry), MEDICAL (worker, 365d) (lines 271-277).
   - `documents.status` enum: `'pending_review'`, `'approved'`, `'rejected'`, `'expired'` (lines 308-313).
   - `access_logs.result` enum: `'granted'`, `'granted_with_warning'`, `'denied'` (lines 363-367).
   - `access_logs.document_snapshot JSONB NOT NULL DEFAULT '[]'` (line 387).
   - `access_logs` does not have an `updated_at` column or trigger, enforcing immutability (line 393).

---

## 2. Logic Chain

1. **Endpoint Design Logic:**
   - From `ORIGINAL_REQUEST.md` R1 and `DEVELOPMENT_RULES.md` §6, the system requires `POST /api/v1/auth/login` and `POST /api/v1/access/check-in`.
   - To support caseta operations cleanly, the web UI must populate the guard's access point (`GET /api/v1/access-points`), record departures (`POST /api/v1/access/check-out`), and provide audit visibility (`GET /api/v1/access-logs`).
   - All responses must wrap payloads in `{ "data": ... }` and errors in `{ "error": { "code": "...", "message": "..." } }` with Spanish messages.

2. **Check-In Verification Logic:**
   - Access verification depends on three layers of validation:
     a. Identity & Assignment: Worker must exist (`users.role = 'worker'`), be active, and be assigned to an active contractor (`contractor_workers`).
     b. Plant Authorization: The contractor must be authorized at this specific plant (`plant_contractors.status = 'active'`). If status is `pending`, `suspended`, or `blocked`, access is immediately `denied`.
     c. Document Compliance: All document types marked required by the plant policy (`plant_document_requirements` or default `document_types`) must have an `approved` document for the worker (or contractor) with `expires_at >= current_date`.
   - If any document is expired:
     - When `plant.block_on_expired_docs = true` -> access is `denied`.
     - When `plant.block_on_expired_docs = false` (and no documents are missing or rejected) -> access is `granted_with_warning`.
   - When all documents are approved and valid -> access is `granted`.
   - Every verification creates a row in `access_logs` with a JSONB snapshot of all evaluated documents.

3. **Tenant Isolation Logic:**
   - From `DEVELOPMENT_RULES.md` §3, multi-tenant queries must include `WHERE tenant_id = $1`.
   - The guard's `tenant_id` is established via JWT claim and injected into context.
   - The query for `access_points` filters by `WHERE id = $1 AND tenant_id = $2`. Even if an attacker passes another tenant's access point UUID, the database returns no record (404), completely preventing cross-tenant access.

4. **Testing Logic:**
   - From `DEVELOPMENT_RULES.md` §9, use cases must have unit tests with mocked interfaces (`testify/mock`) achieving >= 80% coverage.
   - All permutations of access outcomes (`granted`, `granted_with_warning`, `denied` for missing doc, `denied` for expired doc, `denied` for blocked contractor) must be verified without real I/O.
   - Integration tests must verify database query constraints, schema validity, and tenant isolation using PostgreSQL containers.

---

## 3. Caveats

1. **Storage (Cloudflare R2):** Phase 1 MVP focuses on access check-in verification using existing records in PostgreSQL. Cloudflare R2 pre-signed URL generation for document viewing is referenced in `DEVELOPMENT_RULES.md` but is secondary to the caseta check-in engine.
2. **AI Document OCR:** The schema includes `ai_extracted_data` and `ai_confidence` in `documents`, but automatic OCR is a Phase 2 feature; Phase 1 evaluates pre-existing document statuses (`approved`, `expired`, `pending_review`).
3. **Database Migration Tooling:** Both `golang-migrate` Docker container and in-process Go migration runner are viable. The Docker compose specification includes a `migrate` container service as mandated by `DEVELOPMENT_RULES.md` §14.

---

## 4. Conclusion

The Phase 1 MVP backend specification is fully enumerated, coherent, and verified against all authoritative project sources. 
- All endpoints, DTOs, status codes, and Spanish error messages are documented.
- The Clean Architecture directory structure is mapped with strict boundaries.
- The 13 database tables and exact sqlc queries are specified.
- The access verification engine logic handles all edge cases (inactive worker, blocked contractor, missing docs, expired docs with plant policy toggle, audit immutability).
- Multi-tenancy and JWT auth enforcement mechanisms are designed to prevent security bypasses.
- The complete survey is preserved in `survey_report.md`.

---

## 5. Verification Method

1. **Inspect Survey Report:**
   ```bash
   cat /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_1/survey_report.md
   ```
2. **Verify Schema Alignment:**
   Examine table names and enum definitions between `000001_init.up.sql` and `survey_report.md` Section 7.
3. **Verify Development Rules Adherence:**
   Cross-reference `DEVELOPMENT_RULES.md` §2, §3, §6, §7 with `survey_report.md` Sections 4, 5, 8, and 9.
4. **Invalidation Conditions:**
   - Any modification to `000001_init.up.sql` adding or dropping foreign keys or enums.
   - Any modification to `DEVELOPMENT_RULES.md` changing error envelope formats or language requirements.
