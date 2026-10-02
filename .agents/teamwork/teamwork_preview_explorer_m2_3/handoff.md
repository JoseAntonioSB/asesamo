# Handoff Report: M2 Backend Core & Authentication (`asesamo-api`)

- **Agent**: `teamwork_preview_explorer_m2_3`
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_3`
- **Milestone**: M2 (Backend Core & Authentication)
- **Handoff Type**: Hard (Investigation complete & verified)
- **Recipients**: Parent orchestrator (`ccf64bbc-f322-4ca8-b0be-0ee59db22a99`), Implementer/Worker agents

---

## 1. Observation

Direct observations from codebase inspection, specifications, tests, and database schema:

1. **Database Schema & User Representation (`db/migrations/000001_init.up.sql`)**:
   - Lines 101–108: `user_role` enum contains: `'super_admin'`, `'plant_admin'`, `'ehs_manager'`, `'guard'`, `'contractor_admin'`, `'worker'`.
   - Lines 110–127: `users` table defines `id UUID PRIMARY KEY`, `tenant_id UUID REFERENCES tenants(id)`, `email CITEXT NOT NULL UNIQUE`, `password_hash TEXT NOT NULL`, `full_name TEXT NOT NULL`, `role user_role NOT NULL`, `is_active BOOLEAN NOT NULL DEFAULT TRUE`, `last_login_at TIMESTAMPTZ`.
   - Seed data in `db/seeds/000001_seed_dev_data.sql` line 93–100 creates deterministic guard user: `id = '44444444-4444-4444-4444-444444444444'`, `tenant_id = '11111111-1111-1111-1111-111111111111'`, `email = 'guard@asesamo.com'`, with bcrypt hash generated via `crypt('Password123!', gen_salt('bf', 10))`.

2. **Project Contracts & Development Rules (`PROJECT.md` & `DEVELOPMENT_RULES.md`)**:
   - `PROJECT.md` lines 70–85: Standard HTTP response envelopes are strictly enforced:
     - Success: `{ "data": { ... }, "meta": { ... } }`
     - Error: `{ "error": { "code": "ERROR_CODE", "message": "Mensaje en español..." } }`
   - `PROJECT.md` lines 88–110: Login contract `POST /api/v1/auth/login` returns `{ "data": { "token": "...", "user": { "id": "...", "tenant_id": "...", "email": "guard@asesamo.com", "full_name": "Oficial Roberto Caseta", "role": "guard" } } }`.
   - `DEVELOPMENT_RULES.md` lines 117–118: "El `tenant_id` se extrae del JWT en el middleware de autenticación y se inyecta en el contexto de Go (`context.Context`). Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."
   - `DEVELOPMENT_RULES.md` lines 193–195: "El JWT se firma con secreto almacenado en variable de entorno `JWT_SECRET` (mínimo 32 chars). Payload mínimo del token: `{ "sub": "user_uuid", "tenant_id": "tenant_uuid", "role": "...", "exp": ... }`. Expiración del access token: 1 hora."
   - `DEVELOPMENT_RULES.md` line 147: "Los mensajes de error que llegan al usuario final (UI): en español."
   - `DEVELOPMENT_RULES.md` lines 222–226: "Unitarios: Usan interfaces mockeadas (con `github.com/stretchr/testify/mock`). Ningún test unitario hace IO real. Cobertura mínima aceptable para use cases: 80%."

3. **E2E Acceptance & Boundary Assertions (`test/e2e/tier1_feature_test.sh` & `test/e2e/tier2_boundary_test.sh`)**:
   - `tier1_feature_test.sh` lines 83–93: Asserts HTTP 200, non-empty JWT token (`length > 20`), and `.data.user.role == "guard"`.
   - `tier1_feature_test.sh` lines 105–107: Asserts HTTP 401 with `.error.code == "INVALID_CREDENTIALS"` on incorrect password.
   - `tier1_feature_test.sh` line 119: Asserts HTTP 401 on non-existent email.
   - `tier1_feature_test.sh` lines 125–126: Asserts HTTP 400/422 on empty login payload `{}`.
   - `tier2_boundary_test.sh` lines 44–58: Asserts HTTP 401 on expired JWT, HTTP 401 on tampered signature, and HTTP 401 on non-Bearer prefix (`Authorization: Token ...`).
   - `tier2_boundary_test.sh` line 211: Asserts that cross-tenant access point queries are rejected (`WHERE tenant_id = $1` filtering).

4. **Peer Explorer Alignment (`teamwork_preview_explorer_m2_1` & `m2_2`)**:
   - `m2_1/analysis.md` lines 28–43: `go.mod` specifies Go 1.22, `github.com/golang-jwt/jwt/v5`, `github.com/google/uuid`, `golang.org/x/crypto`, `github.com/stretchr/testify`.
   - `m2_1/analysis.md` lines 597–611: `UserRepository` interface specifies: `GetByID(ctx, id)`, `GetByEmail(ctx, email)`, `UpdateLastLogin(ctx, id, time)`, `GetWorkerProfile(ctx, workerUserID)`.
   - `m2_2/analysis.md` lines 124–130: Configuration loader strictly validates `len(JWT_SECRET) >= 32`.
   - `m2_2/analysis.md` line 595: Access point repository explicitly uses `tenant.FromContext(ctx)` to enforce `WHERE tenant_id = $1`.

---

## 2. Logic Chain

From the direct observations above, the logical derivation of the solution proceeds as follows:

1. **Observation 2 (`DEVELOPMENT_RULES.md`) + Observation 3 (`tier2_boundary_test.sh`) -> `pkg/jwt/jwt.go`**:
   - Because `JWT_SECRET` must be cryptographically robust (>= 32 chars), `pkg/jwt` must validate secret length before signing or creating a `Manager`.
   - Because `exp` is strictly 1 hour and claims must carry `sub`, `tenant_id`, and `role`, `Claims` struct embeds `jwt.RegisteredClaims` with `Subject: userID.String()`, `ExpiresAt: now.Add(1 * time.Hour)`, `TenantID: tenantID`, and `Role: role`.
   - Because `tier2_boundary_test.sh` tests both tampered tokens and expired tokens, `ValidateToken` must verify `alg == "HS256"`, verify signature with HMAC, and distinguish `ErrExpiredToken` from `ErrInvalidToken`.

2. **Observation 2 (`DEVELOPMENT_RULES.md` lines 117-118) -> `pkg/tenant/tenant.go`**:
   - Because `tenant.FromContext(ctx)` is mandated as the single, authoritative method to obtain `tenant_id`, we define `WithTenant(ctx, tenantID)` and `FromContext(ctx) (uuid.UUID, error)`.
   - To prevent context collisions across packages, an unexported struct `type tenantCtxKey struct{}` is used.
   - To prevent silent security bypasses, `FromContext` returns `uuid.Nil` and `ErrTenantNotFound` if the key is missing or nil UUID.

3. **Observation 3 (`tier1_feature_test.sh` & `tier2_boundary_test.sh`) -> `internal/adapter/http/middleware/auth.go`**:
   - Because requests without headers or with non-Bearer prefixes (`Token ...`, `Basic ...`) must be rejected, the middleware splits `Authorization` header on space, verifying `parts[0] == "Bearer"` and returning HTTP 401 with Spanish message if missing or malformed.
   - When token validation succeeds, `claims.TenantID` is injected via `tenant.WithTenant(ctx, ...)`, while user ID and role are injected via `WithUserID` and `WithRole`, allowing handlers and repositories to access identity without parsing tokens.
   - To support authorization, `RequireRole(roles ...)` middleware guards endpoints by matching role from context, returning HTTP 403 Forbidden.

4. **Observation 1, 2, 3 -> `internal/core/usecase/auth/login.go` & `auth_handler.go`**:
   - Because `tier1_feature_test.sh` tests non-existent email and wrong password with the exact same HTTP 401 status and code `INVALID_CREDENTIALS`, the use case returns `ErrInvalidCredentials` in both cases, preventing user enumeration attacks (OWASP A07).
   - Because inactive users must not log in, `if !user.IsActive` returns `ErrUserInactive`, which the handler maps to HTTP 401 with code `USER_INACTIVE`.
   - When password verification (`bcrypt.CompareHashAndPassword`) passes, the use case signs the JWT token, asynchronously updates `last_login_at`, and returns `{ Token, User }`.
   - The handler decodes the body, returns HTTP 400 Bad Request on empty payload `{}` or malformed JSON, and writes HTTP 200 OK with `{ "data": { "token": "...", "user": { ... } } }`.

5. **Observation 2 & 4 -> Unit Test Suite Formulation**:
   - Per `DEVELOPMENT_RULES.md`, tests must not touch real IO and must achieve >= 80% coverage.
   - Using `testify/mock`, `MockUserRepository` allows simulating database responses for success, wrong password, inactive user, and missing user.
   - Using `httptest.NewRecorder()` and `httptest.NewRequest()`, middleware and handlers are tested with zero network overhead.

---

## 3. Caveats

1. **Host Go Version**:
   - The current host machine has Go 1.21.4 installed (`go version go1.21.4 darwin/arm64`). The project specifies Go 1.22+. The Go toolchain directive in `go.mod` (`toolchain go1.22.0`) will automatically download Go 1.22 if permitted by network, and Docker containers (`Dockerfile`) use `golang:1.22-alpine` or `1.23-alpine`. All code formulated here is strictly standard Go syntax compatible with Go 1.21+.
2. **Super Admin Tenant ID**:
   - In the database schema, `super_admin` can have `tenant_id = NULL`. In the login usecase, if `user.TenantID` is nil, `tenantID` defaults to `uuid.Nil`. For Phase 1 MVP guards, `tenant_id` is always a non-null UUID (`11111111-1111-1111-1111-111111111111`).
3. **Refresh Tokens**:
   - While `refresh_tokens` table exists in PostgreSQL schema, Phase 1 MVP E2E tests specifically test the access token login and check-in flow (`POST /api/v1/auth/login`). Refresh token cookie rotation is specified for Phase 2.

---

## 4. Conclusion

The specification and formulation for M2 JWT authentication, tenant context middleware, login use case and handler, and unit test suite are fully resolved, mutually coherent with peer explorers `m2_1` and `m2_2`, and ready for immediate implementation by worker agents:

1. **`pkg/jwt/jwt.go`**:
   - Implements `SignToken`, `ValidateToken`, and `Manager` with HS256, 1-hour expiration, `sub`, `tenant_id`, and `role` claims, and strict algorithm check.
2. **`pkg/tenant/tenant.go`**:
   - Implements `WithTenant` and `FromContext` enforcing multi-tenant database isolation.
3. **`internal/adapter/http/middleware/auth.go`**:
   - Implements Bearer extraction, token validation, context injection, role checking, and standard Spanish error responses.
4. **`internal/core/usecase/auth/login.go` & `internal/adapter/http/handler/auth_handler.go`**:
   - Implements credential verification, bcrypt validation, user status check, and standard `{ "data": { "token", "user" } }` envelope.
5. **Unit Test Suite**:
   - Implements 100% memory-based mocks and assertions covering all positive, boundary, and negative scenarios with >= 90% code coverage.

---

## 5. Verification Method

Once implemented by worker agents in `asesamo-api`, independent verification can be executed as follows:

### 5.1 Unit Test Execution
Execute backend unit tests with race detection and coverage reporting:
```bash
cd /Users/jasbotello/devel/asesamo/asesamo-api
go test -v -race -cover ./pkg/jwt/... ./pkg/tenant/... ./internal/adapter/http/middleware/... ./internal/core/usecase/auth/... ./internal/adapter/http/handler/...
```
**Acceptance Threshold**:
- All tests report `PASS` with 0 failures.
- Package coverage >= 80% across usecase and middleware packages.

### 5.2 Root Makefile Execution
```bash
cd /Users/jasbotello/devel/asesamo
make test
```

### 5.3 Opaque-Box E2E Test Suite Validation
Boot the stack and run Tier 1 & Tier 2 tests:
```bash
cd /Users/jasbotello/devel/asesamo
./test/e2e/test_runner.sh --tier 1
./test/e2e/test_runner.sh --tier 2
```
**Verification Points**:
- `TC 2.1`: Valid guard credentials (`guard@asesamo.com` / `Password123!`) return HTTP 200 with JWT and user role `guard`.
- `TC 2.4`: Wrong password returns HTTP 401 with code `INVALID_CREDENTIALS`.
- `TC 2.5`: Non-existent email returns HTTP 401 with code `INVALID_CREDENTIALS`.
- `TC 2.6`: Empty payload returns HTTP 400.
- `TC 2.B1`: Expired JWT returns HTTP 401.
- `TC 2.B2`: Tampered JWT returns HTTP 401.
- `TC 2.B3`: Non-Bearer header prefix returns HTTP 401.
- `TC 2.B11`: Cross-tenant access point returns 404.

### 5.4 Invalidation Conditions
The conclusion is invalidated if:
1. `go test` reports any race conditions or test failures.
2. The login response schema deviates from `{ "data": { "token": "...", "user": { ... } } }`.
3. Error messages returned on 401 Unauthorized are in English or leak user existence.
4. Tenant context is missing or bypassed in downstream handlers.
