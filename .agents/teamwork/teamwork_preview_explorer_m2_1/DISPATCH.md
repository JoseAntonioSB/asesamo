## 2026-10-01T22:53:29Z

[Message] timestamp=2026-10-01T22:53:29Z sender=ccf64bbc-f322-4ca8-b0be-0ee59db22a99 priority=MESSAGE_PRIORITY_HIGH content=You are teamwork_preview_explorer_m2_1.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_1

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql

Milestone M2: Backend Core & Authentication (`asesamo-api`).
Your task:
Analyze and formulate the Clean Architecture package layout and domain foundation:
1. `go.mod` definition: module name `asesamo-api`, Go version (1.22+), specific dependencies (`github.com/go-chi/chi/v5`, `github.com/golang-jwt/jwt/v5`, `github.com/google/uuid`, `github.com/jackc/pgx/v5` or `database/sql` with pgx/pq, `golang.org/x/crypto/bcrypt`, `github.com/stretchr/testify`, `github.com/go-playground/validator/v10`, `github.com/go-chi/cors`).
2. Directory structure for `asesamo-api/`:
   - `cmd/api/main.go`
   - `internal/core/domain/entity/` (User, Tenant, Plant, AccessPoint, Contractor, Document, AccessLog)
   - `internal/core/domain/repository/` (UserRepository, AccessPointRepository, AccessLogRepository, DocumentRepository)
   - `internal/core/domain/errors/` (domain error constants)
   - `internal/core/usecase/auth/`
   - `internal/core/usecase/access/`
   - `internal/adapter/http/` (router, handlers, middlewares, DTOs)
   - `internal/adapter/postgres/` (repository implementations)
   - `internal/infra/config/`
   - `internal/infra/database/`
   - `pkg/jwt/`, `pkg/tenant/`
3. Exact Go struct definitions for core domain entities and repository interfaces.

Write your analysis and proposed code structures to analysis.md and handoff.md in your working directory. Call send_message when done.
