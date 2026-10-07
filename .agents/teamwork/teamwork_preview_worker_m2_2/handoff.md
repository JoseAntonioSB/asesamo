# Handoff Report — Milestone M2 Remediation Implementation

## 1. Observation
1. **Mock Failure & Coverage Deficit in `usecase/access`**:
   - `internal/core/usecase/access/check_in_test.go`: `mockAPRepo` lacked `ListByPlant` method implementation, causing mock panics or incomplete contract implementation when use cases expanded. Slice/pointer type assertions were missing nil guards (`args.Get(0).([]entity.AccessPoint)` vs `if args.Get(0) == nil { return nil, args.Error(1) }`).
   - Prior to remediation, `check_out_test.go` and `list_access_logs_test.go` were missing, leaving overall access statement coverage below the 80% project threshold.
2. **Timing Side-Channel & Account Status Enumeration in `usecase/auth`**:
   - In `internal/core/usecase/auth/login.go`: When user lookup returned `domainErrors.ErrUserNotFound` (or nil), the function immediately returned `domainErrors.ErrInvalidCredentials` in sub-millisecond time (~200µs), whereas existing users triggered a cost-10 `bcrypt.CompareHashAndPassword` (~80ms), exposing an empirical timing side-channel (CWE-208 / OWASP A07).
   - In `internal/core/usecase/auth/login.go`: Inactive check `if !user.IsActive` occurred prior to `bcrypt.CompareHashAndPassword`, allowing attackers to enumerate inactive accounts without supplying valid passwords.
3. **Brute Force Rate Limiting Deficit in `adapter/http`**:
   - In `internal/adapter/http/router.go`: `POST /api/v1/auth/login` lacked IP-based throttling. In `router_adversarial_test.go`, 50 rapid login requests returned `map[401:50]`, with 0 requests rate-limited (HTTP 429).
   - Missing error envelope code `TOO_MANY_REQUESTS` in `internal/core/domain/errors/errors.go`.
4. **Multi-Tenant Query Isolation Missing in PostgreSQL Repositories**:
   - `internal/adapter/postgres/access_point_repository.go`: `GetByID` and `ListByPlant` omitted tenant scoping (`WHERE id = $1 AND deleted_at IS NULL` and `WHERE plant_id = $1 AND ...`), allowing cross-tenant object access if an ID was known. Neither method called `tenant.FromContext(ctx)`.
   - `internal/adapter/postgres/plant_repository.go`: `GetByID` did not inspect `ctx` for tenant context or filter by `tenant_id`.
   - `internal/adapter/postgres/access_log_repository.go`: `GetByID` did not validate tenant context or filter by `tenant_id`.
   - `internal/core/usecase/access/check_in.go`: In direct service invocations, tenant context might not be present in `ctx`.
5. **Toolchain & Environment Verification Results**:
   - Executing `go test -v -race -cover ./...` passes across all packages with 0 data races.
   - `internal/core/usecase/access`: 98.5% statement coverage.
   - `internal/core/usecase/auth`: 100.0% statement coverage.
   - `internal/adapter/http`: 100.0% statement coverage.
   - `internal/adapter/http/middleware`: 92.5% statement coverage.
   - `pkg/tenant`: 100.0% statement coverage.
   - `pkg/jwt`: 91.2% statement coverage.

## 2. Logic Chain
1. **Addressing Mock & Access Coverage (Observation 1)**:
   - Added `ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)` to `mockAPRepo` in `internal/core/usecase/access/check_in_test.go`.
   - Added nil-safe type assertions in all mock repository methods across `check_in_test.go`.
   - Expanded `check_in_test.go` with 14 comprehensive unit test cases covering all edge and error branches (missing fields, access point not found, cross-tenant AP mismatch, plant not found, plant inactive, worker not found, worker inactive, plant mismatch, document expiration, duplicate check-in, DB error).
   - Created `internal/core/usecase/access/check_out_test.go` (4 tests) covering successful checkout, missing worker ID, missing plant ID, and repository failure.
   - Created `internal/core/usecase/access/list_access_logs_test.go` (6 tests) covering plant queries, tenant queries, pagination boundary clamping, missing plant fallback, and error returns.
   - Verified that `usecase/access` achieved 98.5% statement coverage.
2. **Remediating Timing Attacks, User Enumeration & Rate Limiting (Observations 2 & 3)**:
   - Defined `const dummyBcryptHash = "$2a$10$xkateHF30gEES5pFuIga4egG9Mb9C7/ziFxqTnUSc03OlhgZKH9f."` in `internal/core/usecase/auth/login.go`.
   - When user is not found or nil, `login.go` invokes `bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(input.Password))` before returning `domainErrors.ErrInvalidCredentials`, equalizing response latency.
   - Moved `bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))` before `if !user.IsActive`. If password does not match, `ErrInvalidCredentials` is returned regardless of account status.
   - Added `domainErrors.ErrTooManyRequests` in `internal/core/domain/errors/errors.go` (HTTP 429, Code: `TOO_MANY_REQUESTS`, Message: `"Demasiadas solicitudes. Intente más tarde."`).
   - Integrated `github.com/go-chi/httprate v0.16.0` into `asesamo-api/go.mod` and `asesamo-api/go.sum`.
   - Mounted `httprate.LimitBy(10, 1*time.Minute, httprate.KeyByIP, httprate.WithLimitHandler(...))` on `r.With(loginLimiter).Post("/auth/login", cfg.AuthHandler.Login)` in `internal/adapter/http/router.go`.
   - Updated `internal/adapter/http/router_adversarial_test.go`: verified 50 burst requests yield 10 status 401s and 40 status 429s with full envelope structure verification.
   - Verified `usecase/auth` achieved 100.0% statement coverage and balanced timing ratio (1.00x).
3. **Remediating Multi-Tenant Isolation (Observation 4)**:
   - In `internal/adapter/postgres/access_point_repository.go`:
     - In `GetByID`: Added `tenantID, err := tenant.FromContext(ctx); if err != nil { return nil, domainErrors.ErrTenantNotFound }` and updated query to `WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`.
     - In `ListByPlant`: Added `tenantID, err := tenant.FromContext(ctx); if err != nil { return nil, domainErrors.ErrTenantNotFound }` and updated query to `WHERE tenant_id = $1 AND plant_id = $2 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`.
   - In `internal/adapter/postgres/plant_repository.go`:
     - In `GetByID`: Added fail-closed check on `tenant.FromContext(ctx)` returning `domainErrors.ErrTenantNotFound`, and updated query to `WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`.
   - In `internal/adapter/postgres/access_log_repository.go`:
     - In `GetByID`: Added fail-closed check on `tenant.FromContext(ctx)` returning `domainErrors.ErrTenantNotFound`, and updated query to `WHERE tenant_id = $1 AND id = $2`.
   - In `internal/core/usecase/access/check_in.go`:
     - Added fallback propagation: `if _, err := tenant.FromContext(ctx); err != nil && in.TenantID != uuid.Nil { ctx = tenant.WithTenant(ctx, in.TenantID) }`.

## 3. Caveats
- Integration tests requiring live PostgreSQL (`-tags=integration`) depend on external container connectivity (`make docker-up`); all unit and adversarial tests run hermetically with 100% mocked DB interfaces.
- The `dummyBcryptHash` uses standard bcrypt cost 10 matching the production hashing cost configured in `User` entity creation.
- No caveats regarding completeness or code regression; all test suites pass with `-race`.

## 4. Conclusion
All Milestone M2 Remediation requirements have been implemented and verified:
1. `mockAPRepo` implements full interface contract with nil guards; `usecase/access` achieved 98.5% statement coverage.
2. CWE-208 timing attack and account enumeration vulnerabilities remediated in `login.go` and verified empirically.
3. Brute force rate limiting (10 req/min/IP) implemented via `httprate` with standard ASESAMO error envelope and verified via adversarial test.
4. Multi-tenant PostgreSQL queries fail closed upon missing tenant context and explicitly filter by `tenant_id`.
5. Full codebase passes `go test -v -race -cover ./...` with zero race conditions.

## 5. Verification Method
To independently verify the implementation:
1. Run the test suite with race detector and coverage in `asesamo-api`:
   ```bash
   cd asesamo-api && go test -v -race -cover ./...
   ```
2. Verify rate limiting and adversarial protections:
   ```bash
   cd asesamo-api && go test -v -race ./internal/adapter/http -run TestAdversarial_RouterAndBruteForce
   cd asesamo-api && go test -v -race ./internal/core/usecase/auth -run TestLogin_TimingSideChannel
   ```
3. Inspect repository multi-tenant queries:
   - `asesamo-api/internal/adapter/postgres/access_point_repository.go` (GetByID, ListByPlant)
   - `asesamo-api/internal/adapter/postgres/plant_repository.go` (GetByID)
   - `asesamo-api/internal/adapter/postgres/access_log_repository.go` (GetByID)
4. Verify statement coverage on use cases:
   - `asesamo-api/internal/core/usecase/access`: >= 80% (actual: 98.5%)
   - `asesamo-api/internal/core/usecase/auth`: >= 80% (actual: 100.0%)
