# BRIEFING — 2026-10-01T23:00:00Z

## Mission
Analyze and formulate JWT authentication, tenant middleware, auth use case and handler, and comprehensive unit test suite specifications for milestone M2 (asesamo-api).

## 🔒 My Identity
- Archetype: explorer
- Roles: explorer, synthesis
- Working directory: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_3
- Original parent: ccf64bbc-f322-4ca8-b0be-0ee59db22a99
- Milestone: M2: Backend Core & Authentication (asesamo-api)

## 🔒 Key Constraints
- Read-only investigation — do NOT implement in source tree
- Analyze and formulate JWT auth, tenant middleware, auth usecase/handler, and unit test suite
- Adhere strictly to PROJECT.md, DEVELOPMENT_RULES.md, 000001_init.up.sql, and ORIGINAL_REQUEST.md
- Standard JSON envelope: `{"data": { "token": "...", "user": { ... } } }`
- Spanish error message on 401 Unauthorized for invalid/expired tokens
- HS256, 1 hour expiration, Claims: `sub` (user UUID), `tenant_id` (tenant UUID), `role` (user role string), `exp`
- Helper functions: `pkg/tenant/tenant.go`: `WithTenant(ctx, tenantID)` and `FromContext(ctx) (uuid.UUID, error)`
- Mock repository interfaces with `testify/mock`, >=80% coverage, zero real I/O
- Output `analysis.md` and `handoff.md` in own directory

## Current Parent
- Conversation ID: ccf64bbc-f322-4ca8-b0be-0ee59db22a99
- Updated: 2026-10-01T23:00:00Z

## Investigation State
- **Explored paths**: `ORIGINAL_REQUEST.md`, `PROJECT.md`, `DEVELOPMENT_RULES.md`, `db/migrations/000001_init.up.sql`, `db/seeds/000001_seed_dev_data.sql`, `test/e2e/tier1_feature_test.sh`, `test/e2e/tier2_boundary_test.sh`, `teamwork_preview_explorer_m2_1/analysis.md`, `teamwork_preview_explorer_m2_2/analysis.md`.
- **Key findings**: Complete consensus across all 3 M2 explorers on entity schemas, `UserRepository` interface, `pkg/tenant` contracts, `pkg/jwt` contracts, error envelopes, and `testify/mock` zero-IO testing.
- **Unexplored areas**: None for M2 auth scope; M3 access check-in engine follows.

## Key Decisions Made
- `pkg/jwt`: Claims include `sub`, `tenant_id`, `role`, `exp` (1 hr), `iat`. Alg check strictly enforces `HS256` HMAC.
- `pkg/tenant`: Unexported context key, `WithTenant(ctx, tenantID)`, and `FromContext(ctx) (uuid.UUID, error)` with fail-closed return.
- `internal/adapter/http/middleware/auth.go`: Extracts `Bearer <token>`, rejects invalid prefixes with 401, validates token, injects `user_id`, `tenant_id`, `role` into context.
- `internal/core/usecase/auth/login.go`: Protects against user enumeration by returning uniform `INVALID_CREDENTIALS` on missing user or incorrect password. Verifies active status and bcrypt hash.
- Unit Test Suite: Fully specified zero-IO test suite using `testify/mock` and `httptest.NewRecorder()`, covering success, wrong password, inactive user, non-existent user, expired/tampered tokens, and invalid header formats.

## Artifact Index
- DISPATCH.md — Initial dispatch message
- BRIEFING.md — Situational awareness working memory
- progress.md — Heartbeat and status tracker
- analysis.md — Full technical analysis and code implementations
- handoff.md — 5-component handoff report
