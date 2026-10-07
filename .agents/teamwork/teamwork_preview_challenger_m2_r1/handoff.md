# Milestone M2 Gate Adversarial Challenge Report: Security & Authentication

- **Challenger Agent**: `teamwork_preview_challenger_m2_r1`
- **Roles**: `critic`, `specialist`
- **Milestone**: M2 Gate (Backend Core & Authentication)
- **Target Component**: `asesamo-api`
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_challenger_m2_r1`
- **Date**: 2026-10-02T05:10:00Z
- **Final Gate Verdict**: **REJECT**

---

## Challenge Summary

- **Overall Risk Assessment**: **CRITICAL**
- **Summary**:
  While the core stateless JWT implementation in `pkg/jwt` and `middleware.Auth` correctly defends against algorithm confusion (`alg=none`, RS256/ES256/HS384/HS512), signature tampering, token expiration, and short secret keys (<32 bytes), empirical stress-testing has identified critical security vulnerabilities in authentication and multi-tenancy, alongside a broken build/test state in the workspace:
  1. **Test Suite & Linter Failure**: `make test` and `make lint` fail with compilation errors in `internal/core/usecase/access/check_in_test.go:175:34`, violating acceptance criteria from `ORIGINAL_REQUEST.md`.
  2. **Unrestricted Authentication Attempts (CWE-307 / OWASP A07)**: `POST /api/v1/auth/login` lacks rate-limiting middleware, allowing unrestricted credential stuffing and brute-force attacks (50/50 burst requests unthrottled), directly violating `DEVELOPMENT_RULES.md` line 323.
  3. **Account Status Enumeration (OWASP A07)**: `LoginUseCase.Execute` evaluates account active status (`!user.IsActive`) **before** verifying the password hash, revealing whether an account is inactive to unauthenticated attackers using arbitrary passwords.
  4. **Observable Timing Side-Channel for User Enumeration (CWE-208 / OWASP A07)**: `LoginUseCase.Execute` immediately aborts upon user-not-found without computing a dummy bcrypt hash, producing an empirical ~8.56µs vs ~76.57ms response latency differential (an **8,941.5x disparity**) that leaks user registration status.
  5. **Multi-Tenant Query Isolation Violation**: `access_point_repository.go` (`GetByID`, `ListByPlant`) queries the multi-tenant `access_points` table without `WHERE tenant_id = $1`, violating `DEVELOPMENT_RULES.md` line 115 and 119.

---

## 1. Observation

### 1.1 Test Suite & Linter Build Failure
Executing `make test` and `make lint` from repository root yielded:
```bash
$ make test
==> Running backend unit tests...
(cd asesamo-api && go test -v -race -cover ./...)
# asesamo-api/internal/core/usecase/access_test [asesamo-api/internal/core/usecase/access.test]
internal/core/usecase/access/check_in_test.go:175:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:217:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:264:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:285:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
FAIL	asesamo-api/internal/core/usecase/access [build failed]
make: *** [test] Error 1
```
And:
```bash
$ make lint
==> Linting asesamo-api...
(cd asesamo-api && if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else go vet ./...; fi)
# asesamo-api/internal/core/usecase/access_test
vet: internal/core/usecase/access/check_in_test.go:175:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
make: *** [lint] Error 1
```

### 1.2 JWT Security Stress-Test Observations
Direct execution of adversarial probes against `pkg/jwt` and `middleware.Auth`:
- **Alg 'none' attacks**: Tested 4 variants (unsigned with trailing dot, without trailing dot, cased as `None`, all-caps `NONE`). All returned non-nil error and were rejected with HTTP 401 `{"error":{"code":"INVALID_TOKEN",...}}`.
- **Algorithm confusion attacks**:
  - Validly signed 2048-bit RS256 token passed to `jwt.ValidateToken(token, secret)`: Rejected with `ErrInvalidToken`.
  - Validly signed ECDSA P-256 (ES256) token passed to `jwt.ValidateToken(token, secret)`: Rejected with `ErrInvalidToken`.
  - HMAC variants (HS384, HS512) passed to validator: Rejected because validator explicitly pins `jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Name})` (`pkg/jwt/jwt.go:104`).
- **Token expiration**: Tokens with negative duration (10s past leeway, 30 days past) return `ErrExpiredToken` and HTTP 401 `TOKEN_EXPIRED`.
- **Tampered signature & payload**:
  - Byte alteration in signature returns `ErrInvalidToken`.
  - Payload alteration elevating role from `guard` to `super_admin` returns `ErrInvalidToken`.
- **Secret key length enforcement**:
  - `SignToken` and `ValidateToken` with keys of length 0, 1, 16, 31 bytes return `ErrSecretTooShort` or `ErrEmptySecret`. Key of length 32 bytes passes.

### 1.3 Login Endpoint Adversarial Observations
Direct execution of empirical harness against `POST /api/v1/auth/login`:
- **Generic error envelope consistency**:
  - Non-existent user returns HTTP 401:
    `{"error":{"code":"INVALID_CREDENTIALS","message":"Credenciales inválidas. Verifique su correo y contraseña."}}`
  - Existing user with incorrect password returns HTTP 401:
    `{"error":{"code":"INVALID_CREDENTIALS","message":"Credenciales inválidas. Verifique su correo y contraseña."}}`
  - Byte-for-byte error payload matches.
- **Timing Side-Channel (CWE-208)**:
  - Latency benchmark across 20 iterations:
    - Non-existent user: mean **8.56 microseconds** (min: 2.50µs, max: 46.8µs)
    - Existing user with wrong password: mean **76.57 milliseconds** (min: 71.2ms, max: 84.1ms)
    - Disparity: **8,941.5x latency ratio**.
  - Code inspection of `internal/core/usecase/auth/login.go` lines 51-58:
    ```go
    user, err := uc.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if domainErrors.IsNotFound(err) {
            // Prevent user enumeration attack (OWASP A07): return same error as wrong password
            return nil, domainErrors.ErrInvalidCredentials
        }
        return nil, domainErrors.ErrDatabase.WithErr(err)
    }
    ```
    Upon `IsNotFound(err)`, execution immediately returns without invoking `bcrypt.CompareHashAndPassword`.
- **Inactive User Account Enumeration**:
  - Code inspection of `internal/core/usecase/auth/login.go` lines 64-73:
    ```go
    // Verify account status
    if !user.IsActive {
        return nil, domainErrors.ErrUserInactive
    }

    // Verify password hash
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        return nil, domainErrors.ErrInvalidCredentials
    }
    ```
  - Empirical test: Requesting `POST /api/v1/auth/login` with email `"inactive@asesamo.com"` and arbitrary password `"this_is_an_intentionally_wrong_password_xyz"` returned HTTP 401:
    `{"error":{"code":"USER_INACTIVE","message":"La cuenta de usuario se encuentra inactiva."}}`
    This leaks that the account exists and is inactive, even when the password is completely wrong.
- **Brute-Force & Rate Limiting**:
  - 50 rapid sequential HTTP POST requests to `/api/v1/auth/login` from IP `198.51.100.25:4567`:
    - Result: 50 requests returned HTTP 401.
    - Zero (0) requests returned HTTP 429 Too Many Requests.
  - Inspection of `internal/adapter/http/router.go` lines 27-56 reveals no rate limiting middleware mounted on the Chi router.

### 1.4 Tenant Isolation & Fail-Closed Context Observations
- `pkg/tenant/tenant.go`:
  - `FromContext(nil)` -> `(uuid.Nil, ErrTenantNotFound)`
  - `FromContext(context.Background())` -> `(uuid.Nil, ErrTenantNotFound)`
  - `FromContext(WithTenant(ctx, uuid.Nil))` -> `(uuid.Nil, ErrTenantNotFound)`
  - Foreign struct keys do not collide with unexported `tenantKey{}`.
- HTTP Routes:
  - Calling `GET /api/v1/access-points` with token having `tenant_id = uuid.Nil` returns HTTP 401 `UNAUTHORIZED`.
- PostgreSQL Repository Multi-Tenant Queries:
  - Inspection of `internal/adapter/postgres/access_point_repository.go` lines 61-65:
    ```go
    func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
        query := accessPointBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
        row := r.db.QueryRowContext(ctx, query, id)
        return r.scanAccessPoint(row)
    }
    ```
    And lines 91-95:
    ```go
    func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
        query := accessPointBaseQuery + ` WHERE plant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
        rows, err := r.db.QueryContext(ctx, query, plantID)
    ```
    Neither `GetByID` nor `ListByPlant` includes `WHERE tenant_id = $1`, which contradicts `DEVELOPMENT_RULES.md` lines 115 and 119.

---

## 2. Logic Chain

1. **Gate Acceptance Rule**:
   - `ORIGINAL_REQUEST.md` line 42 states: `"Automated tests run via go test ./... pass with zero failures."`
   - `DEVELOPMENT_RULES.md` line 306 states standard verification includes `make test` and `make lint`.
   - From *Observation 1.1*, `make test` and `make lint` fail with compilation error in `check_in_test.go:175:34`.
   - Therefore, the codebase currently fails the fundamental automated test gate.

2. **Login Brute-Force & Credential Stuffing Vulnerability (CWE-307)**:
   - `DEVELOPMENT_RULES.md` line 323 states: `"Rate Limiting: El router chi debe incluir un middleware de rate limiting (ej. go-chi/httprate) en endpoints públicos y de auth."`
   - From *Observation 1.3*, `internal/adapter/http/router.go` does not include any rate limiting middleware, and 50 burst requests were processed without any HTTP 429 throttling.
   - Therefore, the API is vulnerable to unrestricted credential brute-forcing.

3. **Account Status Enumeration Vulnerability (OWASP A07)**:
   - In a secure authentication flow, identity existence and account attributes must never be leaked to unauthenticated users prior to credential verification.
   - From *Observation 1.3*, `LoginUseCase.Execute` checks `!user.IsActive` before `bcrypt.CompareHashAndPassword`.
   - Supplying an invalid password to an inactive account yields `USER_INACTIVE` instead of `INVALID_CREDENTIALS`.
   - Therefore, any unauthenticated attacker can probe whether arbitrary email addresses belong to disabled users without knowing their passwords.

4. **Timing Attack User Enumeration (CWE-208)**:
   - A constant-time response pattern is required to prevent side-channel username enumeration.
   - From *Observation 1.3*, non-existent accounts resolve in ~8.56µs (DB not-found), whereas existing accounts require ~76.57ms (bcrypt execution).
   - This 8,941x timing discrepancy permits statistical discrimination of registered vs non-registered email addresses over the network.
   - Dummy bcrypt computation is missing for the not-found branch.

5. **Multi-Tenant Isolation Compliance (OWASP A01)**:
   - `DEVELOPMENT_RULES.md` line 115 states: `"Cada tabla que contenga datos de un cliente (planta) tendrá una columna tenant_id UUID NOT NULL. Todo query a estas tablas DEBE incluir WHERE tenant_id = $1 como primera condición."`
   - Line 119 states: `"Nunca se permite un query sin filtro de tenant en tablas multi-tenant."`
   - From *Observation 1.4*, `access_point_repository.go` implements `GetByID` and `ListByPlant` without `tenant_id` filtering.
   - Although `tenant.FromContext` is robust, repository queries that omit `tenant_id` allow potential cross-tenant IDOR access if invoked with cross-tenant IDs.

---

## 3. Stress Test Results

| Attack / Stress Vector | Target | Expected Behavior | Actual Behavior | Result |
|---|---|---|---|---|
| Alg "none" (trailing dot) | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Alg "none" (no dot) | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Alg "None" / "NONE" cased | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Alg Mismatch (RS256 vs HS256) | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Alg Mismatch (ES256 vs HS256) | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Alg Mismatch (HS384 / HS512) | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Expired Token Rejection | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrExpiredToken`) | **PASS** |
| Tampered Signature | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Tampered Payload Privilege Escalation | `pkg/jwt.ValidateToken` | Token rejected | Rejected (`ErrInvalidToken`) | **PASS** |
| Secret Key Length (<32 bytes) | `pkg/jwt.SignToken` & `Validate` | Rejected (<32 bytes) | Rejected (`ErrSecretTooShort`) | **PASS** |
| Secret Key Length (32 bytes) | `pkg/jwt.SignToken` & `Validate` | Accepted (==32 bytes) | Accepted (0 errors) | **PASS** |
| Generic Error Envelope Equivalence | `POST /api/v1/auth/login` | Identical payload | Byte-for-byte identical 401 | **PASS** |
| Inactive User Status Exposure | `POST /api/v1/auth/login` | Require valid password first | Leaks `USER_INACTIVE` on wrong pass | **FAIL** |
| Timing Side-Channel User Enumeration | `POST /api/v1/auth/login` | Constant-time response | 8,941.5x latency disparity (~8µs vs ~76ms) | **FAIL** |
| Auth Endpoint Burst Brute Force | `POST /api/v1/auth/login` | Rate limit with HTTP 429 | 0/50 requests throttled (no limiter) | **FAIL** |
| Tenant Context Fail-Closed | `pkg/tenant.FromContext` | Reject nil / missing context | Returns `ErrTenantNotFound` & `uuid.Nil` | **PASS** |
| Endpoint Nil Tenant Fail-Closed | `GET /api/v1/access-points` | 401 Unauthorized | Returns 401 Unauthorized | **PASS** |
| Multi-Tenant Repository Query Scoping | `access_point_repository.go` | All queries filter `tenant_id` | `GetByID` & `ListByPlant` omit `tenant_id` | **FAIL** |
| Test Suite Compilation | `make test` & `make lint` | 0 failures | Build fails in `check_in_test.go:175:34` | **FAIL** |

---

## 4. Caveats

1. **Check-In Feature Boundaries**:
   - `internal/core/usecase/access/check_in.go` and `check_in_test.go` were introduced as part of the transition to Milestone M3. The compilation error in `check_in_test.go` (`*mockAPRepo` missing `ListByPlant`) is located within the test mock, not in the production code logic of M2. However, because Go package compilation evaluates all package test files, it breaks the repository-wide gate commands (`make test` and `make lint`).
2. **PostgreSQL Live Integration Testing**:
   - Live integration tests against the PostgreSQL container (`make test-integration`) were not executed in this pass because the Docker daemon was not running during review; all empirical stress tests were executed in-memory against domain use cases, HTTP Chi router, and repository interfaces.

---

## 5. Conclusion & Actionable Mitigations

### Verdict: **REJECT**

Milestone M2 cannot be approved in its current state due to the blocking compilation failure and critical security vulnerabilities in authentication and repository querying.

### Required Remediations before M2 Re-evaluation:
1. **Fix `check_in_test.go` compilation**: Add `ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)` to `mockAPRepo` in `internal/core/usecase/access/check_in_test.go` so `make test` and `make lint` pass with 0 errors.
2. **Mount Rate Limiting Middleware**: Mount `go-chi/httprate` on `/api/v1/auth/login` (e.g. `httprate.LimitByIP(5, 1*time.Minute)`) per `DEVELOPMENT_RULES.md` line 323 to prevent brute-force attacks.
3. **Fix Inactive User Account Status Leak**: In `internal/core/usecase/auth/login.go`, verify the password hash (`bcrypt.CompareHashAndPassword`) **before** checking `if !user.IsActive`, or return generic `ErrInvalidCredentials` for inactive users until authenticated.
4. **Mitigate Timing Leak (CWE-208)**: When `userRepo.GetByEmail` returns not-found, compute a dummy bcrypt hash comparison (e.g., `_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(input.Password))`) to balance response latency and eliminate the 8,900x timing discrepancy.
5. **Enforce Tenant Filter on `AccessPointRepository`**: Update `access_point_repository.go` so `GetByID` and `ListByPlant` accept `tenantID uuid.UUID` and include `WHERE tenant_id = $1` as required by `DEVELOPMENT_RULES.md` line 115.

---

## 6. Verification Method

To verify these findings:
1. **Verify Test Build Failure**:
   ```bash
   cd /Users/jasbotello/devel/asesamo && make test
   cd /Users/jasbotello/devel/asesamo && make lint
   ```
   *Expected outcome*: Exits with code 1/2 showing `*mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)`.
2. **Verify Rate Limiting Absence**:
   Inspect `internal/adapter/http/router.go` lines 27-56; observe no rate limiting middleware is applied to `/api/v1/auth/login`.
3. **Verify Inactive Account Leak**:
   Inspect `internal/core/usecase/auth/login.go` lines 64-73; observe `!user.IsActive` check precedes `bcrypt.CompareHashAndPassword`.
4. **Verify Repository Tenant Queries**:
   Inspect `internal/adapter/postgres/access_point_repository.go` lines 61-65 and 91-95; observe absence of `WHERE tenant_id = $1`.
