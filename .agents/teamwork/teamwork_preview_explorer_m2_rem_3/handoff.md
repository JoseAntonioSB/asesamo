# Handoff Report: Milestone M2 Remediation — Multi-Tenant Repository Query Isolation

- **Author Agent**: `teamwork_preview_explorer_m2_rem_3`
- **Archetype**: `explorer` (Investigation & Synthesis)
- **Recipient Agent**: `parent` (`ccf64bbc-f322-4ca8-b0be-0ee59db22a99`)
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_3`
- **Date**: 2026-10-02T05:14:30Z
- **Type**: Hard Handoff (Investigation & Patch Specifications Complete)

---

## 1. Observation

1. **Compilation Failure (`make test` and `make lint`)**:
   Executing `cd asesamo-api && go test ./...` produced:
   ```text
   # asesamo-api/internal/core/usecase/access_test [asesamo-api/internal/core/usecase/access.test]
   internal/core/usecase/access/check_in_test.go:175:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
   internal/core/usecase/access/check_in_test.go:217:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
   internal/core/usecase/access/check_in_test.go:264:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
   internal/core/usecase/access/check_in_test.go:285:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
   FAIL asesamo-api/internal/core/usecase/access [build failed]
   ```

2. **`access_point_repository.go` Missing Tenant Filter**:
   - `internal/adapter/postgres/access_point_repository.go` lines 61–65:
     ```go
     func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
         query := accessPointBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
         row := r.db.QueryRowContext(ctx, query, id)
         return r.scanAccessPoint(row)
     }
     ```
   - `internal/adapter/postgres/access_point_repository.go` lines 91–95:
     ```go
     func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
         query := accessPointBaseQuery + ` WHERE plant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
         rows, err := r.db.QueryContext(ctx, query, plantID)
     ```
   Both queries omit `tenant_id` filtering.

3. **`plant_repository.go` Missing Tenant Filter**:
   - `internal/adapter/postgres/plant_repository.go` lines 26–32:
     ```go
     func (r *plantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
         query := `
             SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
             FROM plants
             WHERE id = $1 AND deleted_at IS NULL
         `
     ```
   `GetByID` omits `WHERE tenant_id = $1`.

4. **`access_log_repository.go` Missing Tenant Filter in `GetByID`**:
   - `internal/adapter/postgres/access_log_repository.go` lines 134–143:
     ```go
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
     ```
   `GetByID` omits `WHERE tenant_id = $1`, whereas `ListByTenant`, `ListByPlant`, and `CheckOutLatest` in the same file do filter by `tenant_id`.

5. **`DEVELOPMENT_RULES.md` Explicit Multi-Tenancy Directive**:
   - Line 116: *"Todo query a estas tablas DEBE incluir `WHERE tenant_id = $1` como primera condición."*
   - Line 118: *"Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."*
   - Line 119: *"Nunca se permite un query sin filtro de tenant en tablas multi-tenant."*

6. **Authentication Context Flow**:
   - In `internal/adapter/http/middleware/auth_middleware.go` lines 50–56, `claims.TenantID` is extracted from the JWT and injected into the request context via `ctx = tenant.WithTenant(ctx, claims.TenantID)`.
   - In `pkg/tenant/tenant.go` line 32, `tenant.FromContext(ctx)` returns `(uuid.UUID, error)` and returns `ErrTenantNotFound` if missing or `uuid.Nil`.

7. **`user_repository.go` Global Scope Justification**:
   - In `internal/adapter/postgres/user_repository.go` line 81:
     `SELECT ... FROM users WHERE email = $1 AND deleted_at IS NULL`
   - In `db/migrations/000001_init.up.sql` line 114: `email CITEXT NOT NULL UNIQUE`.
   - In `db/migrations/000001_init.up.sql` line 112: `tenant_id UUID REFERENCES tenants(id) -- NULL only for super_admin`.
   - Users are unauthenticated at the time of calling `POST /api/v1/auth/login`.

---

## 2. Logic Chain

1. **Rule Violation**:
   - From *Observation 2*, `access_point_repository.go` queries `access_points` in `GetByID` and `ListByPlant` without `WHERE tenant_id = $1`.
   - From *Observation 5*, `DEVELOPMENT_RULES.md` Section 3 mandates that *every query* on multi-tenant tables must include `WHERE tenant_id = $1` as the first condition.
   - Therefore, `access_point_repository.go` violates the foundational multi-tenancy rules and allows cross-tenant query execution.

2. **Systemic Inconsistency Across Repositories**:
   - From *Observations 3 and 4*, the same vulnerability exists in `plant_repository.go` (`GetByID`) and `access_log_repository.go` (`GetByID`).
   - Remediation cannot be isolated to `access_point_repository.go` alone; all multi-tenant entity repositories must enforce `tenant_id = $1` on single-entity lookups.

3. **Context vs. Explicit Parameter Reconciliation**:
   - Challenger 1 recommended making `tenantID uuid.UUID` an explicit parameter on `GetByID` and `ListByPlant`.
   - However, from *Observation 5*, `DEVELOPMENT_RULES.md` line 118 explicitly states: *"Un helper centralizado tenant.FromContext(ctx) es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."*
   - Furthermore, from *Observation 6*, `middleware.Auth` already injects `claims.TenantID` into `ctx` for every authenticated HTTP request.
   - If repository implementations extract `tenantID` from `ctx` using `tenant.FromContext(ctx)`:
     a) Zero repository interfaces in `internal/core/domain/repository/` need signature breaking changes.
     b) Repositories fail closed immediately if an unauthenticated context is passed.
     c) Rule 118 is 100% honored.
     d) To support direct unit test calls (e.g. `check_in_test.go`), `CheckInUseCase.Execute` ensures `ctx = tenant.WithTenant(ctx, in.TenantID)` if not already set.

4. **Build Restoration**:
   - From *Observation 1*, `check_in_test.go` fails compilation because `*mockAPRepo` lacks `ListByPlant`.
   - Adding `ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)` to `mockAPRepo` restores interface implementation and resolves the blocking build error.

---

## 3. Caveats

1. **Docker Integration Testing**: Live execution against Dockerized PostgreSQL was not performed in this step as this task is read-only investigation. All findings are verified through static analysis, code trace, AST compilation errors, and SQL schema auditing.
2. **Dual Paradigm Flexibility**: While Approach A (`tenant.FromContext(ctx)`) is strongly recommended based on Rule 118, Approach B (explicit parameter) has also been completely documented with full patch specifications in `analysis.md` should the orchestrator prefer explicit signatures.

---

## 4. Conclusion

1. **Multi-Tenant Isolation Remediations**:
   - `internal/adapter/postgres/access_point_repository.go`: Update `GetByID` and `ListByPlant` to call `tenant.FromContext(ctx)` and execute with `WHERE tenant_id = $1`.
   - `internal/adapter/postgres/plant_repository.go`: Update `GetByID` to call `tenant.FromContext(ctx)` and execute with `WHERE tenant_id = $1`.
   - `internal/adapter/postgres/access_log_repository.go`: Update `GetByID` to call `tenant.FromContext(ctx)` and execute with `WHERE tenant_id = $1`.
   - `internal/core/usecase/access/check_in.go`: Propagate `in.TenantID` into `ctx` if not already present.
2. **Build Restoration**:
   - `internal/core/usecase/access/check_in_test.go`: Implement `ListByPlant` on `*mockAPRepo`.
3. **`user_repository.go` Status**:
   - Retain `GetByEmail` as globally scoped without `tenant_id` to support unauthenticated login and `super_admin` accounts.

---

## 5. Verification Method

1. **Compile & Run Test Suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race -cover ./...
   ```
   *Expected result*: Zero compilation errors; all packages PASS.
2. **Lint Verification**:
   ```bash
   cd /Users/jasbotello/devel/asesamo && make lint
   ```
   *Expected result*: Clean exit (code 0).
3. **Repository Query Verification**:
   Inspect `internal/adapter/postgres/access_point_repository.go`, `plant_repository.go`, and `access_log_repository.go` to confirm every query to `access_points`, `plants`, and `access_logs` includes `WHERE tenant_id = $1`.
4. **Detailed Reference**:
   Inspect `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_3/analysis.md` for complete diffs, comparative matrix, and alternative patch specs.
