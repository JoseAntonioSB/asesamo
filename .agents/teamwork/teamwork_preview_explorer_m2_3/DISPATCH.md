## 2026-10-01T22:53:29Z

You are teamwork_preview_explorer_m2_3.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_3

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql

Milestone M2: Backend Core & Authentication (`asesamo-api`).
Your task:
Analyze and formulate JWT authentication, tenant middleware, and unit test suite:
1. `pkg/jwt/jwt.go`:
   - Claims: `sub` (user UUID), `tenant_id` (tenant UUID), `role` (user role string), `exp` (1 hour expiration).
   - SignToken and ValidateToken functions using HS256 and `JWT_SECRET`.
2. `pkg/tenant/tenant.go`:
   - `WithTenant(ctx, tenantID)` and `FromContext(ctx) (uuid.UUID, error)` helper functions per DEVELOPMENT_RULES.md lines 117-118.
3. `internal/adapter/http/middleware/auth.go`:
   - Extract Bearer token from `Authorization` header.
   - Validate JWT, inject `user_id`, `tenant_id`, and `role` into context.
   - Return 401 Unauthorized with Spanish error message on invalid or expired token.
4. `internal/core/usecase/auth/login.go` and `internal/adapter/http/handler/auth_handler.go`:
   - Input: email, password.
   - Verify user exists and is active; verify password with `bcrypt.CompareHashAndPassword`.
   - Generate JWT; return standard envelope `{ "data": { "token": "...", "user": { ... } } }`.
5. Unit test suite:
   - Mock repository interfaces (`testify/mock`).
   - Test login success, wrong password, inactive user, non-existent user, token validation, tenant context extraction with zero real IO and >= 80% coverage.

Write your analysis and proposed code structures to analysis.md and handoff.md in your working directory. Call send_message when done.
