# Milestone M2 Gate 2 Review & Adversarial Challenge Report

## 1. Observation

1. **HTTP Layer & Chi Router Structure**:
   - In `asesamo-api/internal/adapter/http/router.go` (lines 28–72):
     ```go
     func NewRouter(cfg RouterConfig) http.Handler {
         r := chi.NewRouter()
         r.Use(chimiddleware.RequestID)
         r.Use(chimiddleware.RealIP)
         r.Use(chimiddleware.Logger)
         r.Use(chimiddleware.Recoverer)
         r.Use(customMiddleware.CORS(cfg.CORSOrigins))
         ...
     ```
     Public health checks are exposed at `GET /health` and `GET /api/v1/health` (lines 39–40).
     The `/api/v1` group mounts public and protected routes (lines 43–69).

2. **Rate Limiting on `POST /api/v1/auth/login`**:
   - In `asesamo-api/internal/adapter/http/router.go` (lines 45–55):
     ```go
     loginLimiter := httprate.LimitBy(
         10,
         1*time.Minute,
         httprate.KeyByIP,
         httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
             dto.WriteError(w, domainErrors.ErrTooManyRequests)
         }),
     )
     r.With(loginLimiter).Post("/auth/login", cfg.AuthHandler.Login)
     ```
   - In `asesamo-api/internal/core/domain/errors/errors.go` (lines 193–197):
     ```go
     ErrTooManyRequests = &AppError{
         Code:       "TOO_MANY_REQUESTS",
         Message:    "Demasiadas solicitudes. Intente más tarde.",
         HTTPStatus: http.StatusTooManyRequests,
     }
     ```
   - In `asesamo-api/internal/adapter/http/router_adversarial_test.go` (lines 104–154):
     `TestAdversarial_RouterAndBruteForce/EMPIRICAL_TEST:_Brute_force_resilience_against_POST_/api/v1/auth/login` executes 50 burst requests from IP `203.0.113.195:12345`. Result: exactly 10 requests returned HTTP 401 (`INVALID_CREDENTIALS`), and 40 requests returned HTTP 429 (`TOO_MANY_REQUESTS`) formatted in the standard `{ "error": { "code": "TOO_MANY_REQUESTS", "message": "Demasiadas solicitudes. Intente más tarde." } }` envelope.

3. **Middlewares Conformance**:
   - **Auth Middleware**: `asesamo-api/internal/adapter/http/middleware/auth_middleware.go` (lines 16–59) validates JWT via `jwt.ValidateToken(tokenStr, secretBytes)`, checks expiration, algorithm, signature, extracts `claims.TenantID`, `claims.UserID()`, `claims.Role`, and injects them into request context using `tenant.WithTenant(ctx, claims.TenantID)`, `tenant.WithUserID(ctx, userID)`, and `tenant.WithRole(ctx, claims.Role)`.
   - **Tenant Context Helper**: `asesamo-api/pkg/tenant/tenant.go` (lines 20–45) uses unexported type `tenantKey{}` to prevent collisions, and `FromContext(ctx)` fails closed returning `uuid.Nil, ErrTenantNotFound` if the key is missing, wrong type, or `uuid.Nil`.
   - **CORS Middleware**: `asesamo-api/internal/adapter/http/middleware/cors_middleware.go` (lines 10–19) wraps `github.com/go-chi/cors` with allowed methods (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`), allowed headers, and `AllowCredentials: true`.
   - **Logging & Diagnostics**: `chimiddleware.RequestID`, `chimiddleware.RealIP`, `chimiddleware.Logger`, `chimiddleware.Recoverer` are mounted globally on the Chi multiplexer.

4. **PostgreSQL Repositories Multi-Tenant Query Isolation**:
   - `asesamo-api/internal/adapter/postgres/access_point_repository.go`:
     - `GetByID` (lines 62–70):
       ```go
       tenantID, err := tenant.FromContext(ctx)
       if err != nil {
           return nil, domainErrors.ErrTenantNotFound
       }
       query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`
       row := r.db.QueryRowContext(ctx, query, tenantID, id)
       return r.scanAccessPoint(row)
       ```
     - `ListByPlant` (lines 96–122):
       ```go
       tenantID, err := tenant.FromContext(ctx)
       if err != nil {
           return nil, domainErrors.ErrTenantNotFound
       }
       query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND plant_id = $2 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
       rows, err := r.db.QueryContext(ctx, query, tenantID, plantID)
       ```
   - `asesamo-api/internal/adapter/postgres/plant_repository.go`:
     - `GetByID` (lines 27–71):
       ```go
       tenantID, err := tenant.FromContext(ctx)
       if err != nil {
           return nil, domainErrors.ErrTenantNotFound
       }
       query := `
           SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
           FROM plants
           WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL
       `
       ```
   - `asesamo-api/internal/adapter/postgres/access_log_repository.go`:
     - `GetByID` (lines 135–148):
       ```go
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
       ```
   - `asesamo-api/internal/adapter/postgres/repository_tenant_test.go`:
     `TestRepositories_TenantContextFailClosed` executes 20 subtests verifying that `AccessPointRepository.GetByID`, `AccessPointRepository.ListByPlant`, `PlantRepository.GetByID`, and `AccessLogRepository.GetByID` all fail closed (`ErrTenantNotFound`) when passed empty contexts, nil contexts, contexts with `uuid.Nil`, or contexts with foreign keys.

5. **Independent Build, Lint, and Test Execution**:
   - `go test -v -race -cover ./internal/adapter/... ./pkg/...`: Exited code 0 with 0 data races.
     - `internal/adapter/http`: 100.0% statement coverage.
     - `internal/adapter/http/middleware`: 92.5% statement coverage.
     - `pkg/tenant`: 100.0% statement coverage.
     - `pkg/jwt`: 91.2% statement coverage.
   - `go test -v -race -cover ./...`: Exited code 0 across the entire application.
     - `internal/core/usecase/access`: 98.5% statement coverage.
     - `internal/core/usecase/auth`: 100.0% statement coverage.
     - `internal/infra/config`: 97.7% statement coverage.
   - `make lint` / `go vet ./...`: Exited code 0 without any warnings.
   - `make build`: Exited code 0, compiling `bin/api` successfully.

---

## 2. Logic Chain

1. **HTTP Layer Conformance**:
   - Observation 1 demonstrates that Chi v5 router is correctly constructed in `internal/adapter/http/router.go`.
   - The route hierarchy cleanly separates unauthenticated endpoints (`/health`, `/api/v1/health`, `/api/v1/auth/login`) from authenticated endpoints (`/api/v1/access-points`, `/api/v1/access/check-in`, `/api/v1/access/check-out`, `/api/v1/access-logs`) via Chi middleware sub-routing (`r.Group`).
   - All response rendering delegates to `dto.WriteSuccess` or `dto.WriteError`, guaranteeing `{ "data": ... }` and `{ "error": { "code", "message" } }` envelope compliance per `PROJECT.md` and `DEVELOPMENT_RULES.md`.

2. **Rate Limiting Verification**:
   - Observation 2 demonstrates that `POST /api/v1/auth/login` is guarded with `httprate.LimitBy(10, 1*time.Minute, httprate.KeyByIP, WithLimitHandler(...))`.
   - When requests exceed the 10 req/min quota from an IP, `dto.WriteError(w, domainErrors.ErrTooManyRequests)` writes HTTP 429 with JSON payload:
     `{"error":{"code":"TOO_MANY_REQUESTS","message":"Demasiadas solicitudes. Intente más tarde."}}`.
   - The test `router_adversarial_test.go` empirically confirms 10 allowed attempts and 40 throttled attempts under a 50-request burst.

3. **Multi-Tenant Isolation Assurance**:
   - Observation 4 confirms that `AccessPointRepository.GetByID`, `AccessPointRepository.ListByPlant`, `PlantRepository.GetByID`, and `AccessLogRepository.GetByID` strictly call `tenant.FromContext(ctx)` as their very first action.
   - If `ctx` lacks a valid tenant UUID, they return `domainErrors.ErrTenantNotFound` without initiating a database query.
   - Every SQL query uses parameterized placeholders with `$1` bound to `tenantID` (`WHERE tenant_id = $1 ...`).
   - In `CheckInUseCase.Execute` (`check_in.go`), the retrieved access point and plant are checked against the authenticated caller's tenant context, establishing multi-layered tenant boundary enforcement.

4. **Integrity & Quality Check**:
   - Observations 4 and 5 verify that tests are genuine and execute real assertion logic.
   - No hardcoded test responses, facades, or test bypasses were discovered.
   - All race detector checks passed cleanly.

---

## 3. Caveats

1. **In-Memory Rate Limiting Scope**:
   - `httprate.LimitBy` maintains an in-memory window per process instance. In a single-instance deployment (or when deployed behind IP-sticky load balancing), this provides robust brute-force protection. If horizontal clustering with multiple API instances without session stickiness is deployed in Phase 2, a Redis-backed rate limiter backend should be adopted. For Phase 1 MVP, in-memory `httprate` satisfies all architectural rules.
2. **Reverse Proxy Real-IP Trust**:
   - `chimiddleware.RealIP` parses `X-Forwarded-For` and `X-Real-IP`. In production, the edge reverse proxy (e.g. Cloudflare or Nginx) must be configured to strip untrusted upstream headers so external attackers cannot forge client IPs to bypass rate limits.

---

## 4. Conclusion & Final Verdict

All requirements for Milestone M2 Gate 2 have been thoroughly verified, tested, and validated without any integrity violations or logic bypasses:
- Chi router is fully conformant with Clean Architecture.
- Rate limiting on `POST /api/v1/auth/login` (10 req/min/IP) properly returns HTTP 429 and standard error envelope.
- Auth, Tenant Context, CORS, and Logger middlewares operate correctly.
- PostgreSQL repositories (`AccessPointRepository`, `PlantRepository`, `AccessLogRepository`) strictly enforce multi-tenant isolation via `tenant.FromContext(ctx)` and `WHERE tenant_id = $1`.
- Tests pass with zero race conditions and high statement coverage.

**Verdict: APPROVE**

---

## 5. Verification Method

To independently reproduce this verification:
1. Run the test suite with race detector and coverage:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race -cover ./internal/adapter/... ./pkg/...
   ```
2. Verify rate limiting and fail-closed tenant tests:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race ./internal/adapter/http -run TestAdversarial_RouterAndBruteForce
   cd /Users/jasbotello/devel/asesamo/asesamo-api && go test -v -race -coverpkg=asesamo-api/internal/adapter/postgres ./internal/adapter/postgres/... -run TestRepositories_TenantContextFailClosed
   ```
3. Run linter and binary compilation:
   ```bash
   cd /Users/jasbotello/devel/asesamo && make lint && make build
   ```
