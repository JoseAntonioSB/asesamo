# BRIEFING — 2026-10-01T23:08:00Z

## Mission
Implement the entire Go backend service (`asesamo-api`) for Milestone 2 adhering strictly to Clean Architecture, Project rules, Spanish domain errors, and >= 80% test coverage on use cases.

## 🔒 My Identity
- Archetype: worker
- Roles: implementer, qa
- Working directory: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m2_1
- Original parent: ccf64bbc-f322-4ca8-b0be-0ee59db22a99
- Milestone: m2 (Go backend service)

## 🔒 Key Constraints
- Must read ORIGINAL_REQUEST.md, PROJECT.md, DEVELOPMENT_RULES.md, migrations, and explorer handoffs.
- Integrity: DO NOT hardcode test results, create dummy/facade implementations, or circumvent genuine logic.
- Clean Architecture layout: entity, repository interfaces, domain errors, usecases, infra (config, database), adapter (http handlers, middleware, router, postgres repo implementations), pkg (jwt, tenant).
- Dependencies: chi/v5, golang-jwt/jwt/v5, google/uuid, jackc/pgx/v5, crypto/bcrypt, testify, validator/v10, cors.
- Standard response envelope: `{ "data": ... }` or `{ "error": { "code": "...", "message": "..." } }`.
- Domain errors with UPPER_SNAKE_CASE codes and Spanish user-facing messages.
- Unit tests must pass with 0 failures and >= 80% coverage on usecases.

## Current Parent
- Conversation ID: ccf64bbc-f322-4ca8-b0be-0ee59db22a99
- Updated: 2026-10-01T23:08:00Z

## Task Summary
- **What to build**: Full Go backend `asesamo-api` with Chi, PGX, JWT, Tenant context, Login and Access Points use cases, handlers, middleware, and domain entities/repos.
- **Success criteria**: `go mod tidy`, `go build ./...`, `go test -v -cover ./...` pass with 0 failures and usecases >= 80% coverage.
- **Interface contracts**: PROJECT.md, DEVELOPMENT_RULES.md, 000001_init.up.sql.
- **Code layout**: asesamo-api/

## Key Decisions Made
- Implemented `DocumentSnapshot` with SQL `driver.Valuer` and `sql.Scanner` for PostgreSQL `JSONB` support.
- Implemented `pkg/tenant` with unexported context keys and error checks to guarantee multi-tenant security.
- Enforced algorithm check `HS256` and secret length >= 32 chars in `pkg/jwt`.
- Enforced bcrypt comparison and user status validation in `LoginUseCase`.
- Use cases achieved 100.0% test coverage with `testify/mock`.
- Application boots with graceful shutdown and masked database URL logging.

## Artifact Index
- `asesamo-api/go.mod`
- `asesamo-api/cmd/api/main.go`
- `asesamo-api/pkg/jwt/jwt.go` & `jwt_test.go`
- `asesamo-api/pkg/tenant/tenant.go` & `tenant_test.go`
- `asesamo-api/internal/core/domain/entity/*.go`
- `asesamo-api/internal/core/domain/errors/errors.go` & `errors_test.go`
- `asesamo-api/internal/core/domain/repository/*.go`
- `asesamo-api/internal/core/usecase/auth/login.go` & `login_test.go`
- `asesamo-api/internal/core/usecase/access/list_access_points.go` & `list_access_points_test.go`
- `asesamo-api/internal/infra/config/config.go` & `config_test.go`
- `asesamo-api/internal/infra/database/postgres.go`
- `asesamo-api/internal/adapter/postgres/user_repository.go`
- `asesamo-api/internal/adapter/postgres/access_point_repository.go`
- `asesamo-api/internal/adapter/http/dto/envelope.go` & `envelope_test.go`
- `asesamo-api/internal/adapter/http/dto/auth_dto.go`
- `asesamo-api/internal/adapter/http/dto/access_point_dto.go`
- `asesamo-api/internal/adapter/http/middleware/cors_middleware.go`
- `asesamo-api/internal/adapter/http/middleware/auth_middleware.go` & `auth_middleware_test.go`
- `asesamo-api/internal/adapter/http/handler/health_handler.go` & `health_handler_test.go`
- `asesamo-api/internal/adapter/http/handler/auth_handler.go` & `auth_handler_test.go`
- `asesamo-api/internal/adapter/http/handler/access_point_handler.go` & `access_point_handler_test.go`
- `asesamo-api/internal/adapter/http/router.go` & `router_test.go`

## Change Tracker
- **Files modified**: All backend service files in `asesamo-api/` created and verified.
- **Build status**: `go build ./...` passed (0 errors), binary compiled at `asesamo-api/bin/api`.
- **Pending issues**: None.

## Quality Status
- **Build/test result**: All unit tests pass with race detector enabled (`go test -v -race -cover ./...`), 0 failures.
  - `internal/core/usecase/auth`: 100.0% coverage
  - `internal/core/usecase/access`: 100.0% coverage
  - `internal/core/domain/errors`: 95.0% coverage
  - `internal/core/domain/entity`: 81.8% coverage
  - `internal/infra/config`: 97.7% coverage
  - `internal/adapter/http`: 100.0% coverage
  - `internal/adapter/http/dto`: 100.0% coverage
  - `internal/adapter/http/handler`: 95.7% coverage
  - `internal/adapter/http/middleware`: 92.5% coverage
  - `pkg/tenant`: 100.0% coverage
  - `pkg/jwt`: 91.2% coverage
- **Lint status**: `make lint` clean (0 warnings).
- **Tests added/modified**: 11 unit test suites added.

## Loaded Skills
- None
