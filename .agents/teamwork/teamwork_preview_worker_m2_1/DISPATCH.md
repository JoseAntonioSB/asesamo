## 2026-10-01T22:57:41Z

You are teamwork_preview_worker_m2_1.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m2_1

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_1/handoff.md and analysis.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_2/handoff.md and analysis.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_3/handoff.md and analysis.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & File Ownership:
You exclusively own and must implement the entire Go backend service in `/Users/jasbotello/devel/asesamo/asesamo-api/`:
1. `go.mod` (module `asesamo-api`, Go 1.22+, dependencies: `chi/v5`, `golang-jwt/jwt/v5`, `google/uuid`, `jackc/pgx/v5`, `crypto/bcrypt`, `testify`, `validator/v10`, `cors`).
2. `cmd/api/main.go`: Application entrypoint, environment validation, database connection pool, Clean Architecture dependency injection, Chi router mounting, graceful shutdown.
3. `internal/core/domain/`:
   - `entity/`: User, Tenant, Plant, AccessPoint, Contractor, PlantContractor, ContractorWorker, DocumentType, Document, AccessLog, DocumentSnapshot (with SQL driver/scanner for JSONB).
   - `repository/`: UserRepository, AccessPointRepository, PlantRepository, DocumentRepository, AccessLogRepository interfaces.
   - `errors/`: Domain errors with UPPER_SNAKE_CASE codes and Spanish user-facing messages.
4. `internal/infra/`:
   - `config/config.go` & `config_test.go`: Environment configuration loader validating DATABASE_URL, APP_PORT, JWT_SECRET (>= 32 chars), APP_ENV.
   - `database/postgres.go`: PostgreSQL pool using `jackc/pgx/v5/stdlib` with connection limits, health ping check, and graceful close.
5. `pkg/`:
   - `jwt/jwt.go` & `jwt_test.go`: Claims (`sub`, `tenant_id`, `role`, `exp` 1 hour), HS256 SignToken & ValidateToken.
   - `tenant/tenant.go` & `tenant_test.go`: WithTenant & FromContext helpers per rules line 117-118.
6. `internal/adapter/`:
   - `http/middleware/`: CORS middleware, Auth middleware (extract Bearer, validate JWT, inject identity), Tenant context middleware, Spanish error responses.
   - `http/handler/`:
     * Healthcheck handler: `GET /health` and `GET /api/v1/health` returning `{"status": "ok", "database": "connected", "version": "1.0.0"}`.
     * Auth handler: `POST /api/v1/auth/login` validating email/password, bcrypt compare, returning token and user profile in standard envelope `{ "data": { ... } }`.
     * Access points handler: `GET /api/v1/access-points` returning active access points for the guard's tenant/plant in `{ "data": [ ... ] }`.
   - `http/router.go`: Chi router wiring public vs protected routes.
   - `postgres/`:
     * `user_repository.go`: GetByEmail, GetByID with bcrypt password hash.
     * `access_point_repository.go`: ListByTenant with `WHERE tenant_id = $1 AND is_active = true`.
7. `internal/core/usecase/`:
   - `auth/login.go` & `login_test.go`: LoginUseCase with mocked repository (`testify/mock`), verifying login success, bad password, inactive user, non-existent user with >= 80% coverage.
   - `access/list_access_points.go` & `list_access_points_test.go`: ListAccessPointsUseCase with unit tests.

Build and Test Verification:
- Inside `/Users/jasbotello/devel/asesamo/asesamo-api`, run:
  * `go mod tidy`
  * `go build ./...`
  * `go test -v -cover ./...`
- All unit tests must pass with 0 failures and >= 80% coverage on use cases.
- Document all implemented files, build commands, and test outputs in `handoff.md` in your working directory.
- Send a completion message via send_message when done.
