## 2026-10-02T05:18:37Z

You are teamwork_preview_worker_m2_2.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m2_2

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_orchestrator/GATE_STATUS.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_1/handoff.md and analysis.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2/handoff.md and analysis.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_3/handoff.md and analysis.md

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A teamwork_preview_auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Scope & File Ownership:
You exclusively own and must implement the verified remediation fixes across `asesamo-api`:
1. Test Mock & Coverage:
   - In `internal/core/usecase/access/check_in_test.go`:
     * Implement missing method `ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)` on `mockAPRepo`.
     * Add nil guards on all slice and pointer type assertions in mock methods (`if args.Get(0) == nil { return nil, args.Error(1) }`).
     * Add branch tests for input validations, entity lookups, compliance failures, and log errors per Explorer 1 analysis.
   - Create `internal/core/usecase/access/check_out_test.go` and `internal/core/usecase/access/list_access_logs_test.go` using the complete implementations in Explorer 1's analysis.md.
2. Auth Security, Timing Side-Channel & Rate Limiting:
   - In `internal/core/usecase/auth/login.go`:
     * Add dummy bcrypt hash comparison when `user == nil` (using a valid cost-10 bcrypt hash constant) to prevent timing side-channel attacks (CWE-208).
     * Verify password hash with `bcrypt.CompareHashAndPassword` BEFORE checking `!user.IsActive`. Return `ErrInvalidCredentials` if password does not match. Only return `ErrUserInactive` if the password is valid!
   - Update `internal/core/usecase/auth/login_test.go` to test non-existent user timing branch, inactive user with wrong password (returns invalid credentials), and inactive user with correct password (returns user inactive).
   - In `internal/adapter/http/router.go`:
     * Add `httprate.LimitByIP(10, 1*time.Minute)` to `POST /api/v1/auth/login` per DEVELOPMENT_RULES.md line 323, returning HTTP 429 with standard ASESAMO envelope `{"error": {"code": "TOO_MANY_REQUESTS", "message": "Demasiadas solicitudes. Intente más tarde."}}`.
     * Add `github.com/go-chi/httprate` to `go.mod` and run `go mod tidy`.
3. Multi-Tenant Query Isolation:
   - In `internal/adapter/postgres/access_point_repository.go`:
     * Update `GetByID` and `ListByPlant` to call `tenant.FromContext(ctx)` (fail closed with `tenant.ErrTenantNotFound` if missing) and query with `WHERE tenant_id = $1`.
   - In `internal/adapter/postgres/plant_repository.go`:
     * Update `GetByID` to call `tenant.FromContext(ctx)` and query with `WHERE tenant_id = $1`.
   - In `internal/adapter/postgres/access_log_repository.go`:
     * Update `GetByID` to call `tenant.FromContext(ctx)` and query with `WHERE tenant_id = $1`.
   - In `internal/core/usecase/access/check_in.go`:
     * Ensure `ctx = tenant.WithTenant(ctx, in.TenantID)` if not already present.

Build and Test Verification:
Inside `/Users/jasbotello/devel/asesamo/asesamo-api`:
- Run `go mod tidy`
- Run `go test -v -race -cover ./...` (All tests must pass, 0 data races, and usecase statement coverage >= 80%).
- Run `make lint` from repository root (must pass with exit code 0).
- Run `make build` from repository root (must compile `bin/api` with exit code 0).

Document all changed files, test output, coverage percentages, and verification commands in `handoff.md` in your working directory. Send a completion message via send_message when done.
