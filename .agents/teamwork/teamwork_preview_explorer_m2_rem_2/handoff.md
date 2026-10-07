# Handoff Report: Milestone M2 Remediation (Auth Security, Timing Side-Channel & Rate Limiting)

- **Agent**: `teamwork_preview_explorer_m2_rem_2`
- **Role**: Explorer (Investigation, Synthesis)
- **Milestone**: M2 Remediation
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2`
- **Date**: 2026-10-02T05:18:00Z
- **Recipient**: `teamwork_preview_orchestrator` / `parent` (`ccf64bbc-f322-4ca8-b0be-0ee59db22a99`)

---

## 1. Observation

### 1.1 Timing Disparity in `login.go` (CWE-208)
- File: `/Users/jasbotello/devel/asesamo/asesamo-api/internal/core/usecase/auth/login.go`
- Lines 51-62:
```go
user, err := uc.userRepo.GetByEmail(ctx, email)
if err != nil {
	if domainErrors.IsNotFound(err) {
		// Prevent user enumeration attack (OWASP A07): return same error as wrong password
		return nil, domainErrors.ErrInvalidCredentials
	}
	return nil, domainErrors.ErrDatabase.WithErr(err)
}

if user == nil {
	return nil, domainErrors.ErrInvalidCredentials
}
```
- Direct tool command: `go test -v ./internal/core/usecase/auth/...` (in `asesamo-api`)
- Output verbatim:
```
=== RUN   TestAdversarial_LoginUseCase/EMPIRICAL_TIMING_ANALYSIS:_Compare_latency_of_nonexistent_user_vs_wrong_password
    login_adversarial_test.go:137: Timing side-channel empirical measurement:
    login_adversarial_test.go:138:   Avg latency for non-existent user: 10.125µs
    login_adversarial_test.go:139:   Avg latency for wrong password (bcrypt): 76.965941ms
    login_adversarial_test.go:140:   Timing ratio (wrong password / nonexistent): 7601.6x
```
Execution immediately aborts on user not found, yielding an empirical 7,601.6x latency disparity.

### 1.2 Inactive Account Enumeration in `login.go` (OWASP A07)
- File: `/Users/jasbotello/devel/asesamo/asesamo-api/internal/core/usecase/auth/login.go`
- Lines 64-72:
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
- Direct tool command: `go test -v ./internal/core/usecase/auth/...` (in `asesamo-api`)
- Output verbatim:
```
=== RUN   TestAdversarial_LoginUseCase/VULNERABILITY:_Inactive_user_enumeration_before_password_validation
    login_adversarial_test.go:104: Observed error code for inactive account with WRONG password: USER_INACTIVE (La cuenta de usuario se encuentra inactiva.)
```
Checking `!user.IsActive` before `bcrypt.CompareHashAndPassword` allows unauthenticated attackers to supply any arbitrary password and confirm whether an email address exists and is deactivated.

### 1.3 Missing Rate Limiting on Login in `router.go` (CWE-307)
- File: `/Users/jasbotello/devel/asesamo/asesamo-api/internal/adapter/http/router.go`
- Lines 39-42:
```go
	// API v1 group
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth
		r.Post("/auth/login", cfg.AuthHandler.Login)
```
- Direct tool command: `go test -v ./internal/adapter/http/...` (in `asesamo-api`)
- Output verbatim:
```
    router_adversarial_test.go:129: Brute force results over 50 burst requests: map[401:50]
    router_adversarial_test.go:135: Rate-limited requests (HTTP 429): 0
    router_adversarial_test.go:136: Processed unauthorized attempts (HTTP 401): 50
```
- Rule: `DEVELOPMENT_RULES.md` line 323 states:
> *"Rate Limiting: El router chi debe incluir un middleware de rate limiting (ej. go-chi/httprate) en endpoints públicos y de auth."*
No rate limiting middleware was mounted, processing 50 burst login attempts in ~40ms without throttling.

---

## 2. Logic Chain

1. **Timing Side-Channel Elimination**:
   - *From Observation 1.1*, nonexistent users return in ~10µs because no bcrypt key expansion is performed, whereas existing accounts execute bcrypt cost 10 taking ~77ms.
   - Performing a dummy bcrypt comparison using a valid precomputed cost-10 hash (`$2a$10$xkateHF30gEES5pFuIga4egG9Mb9C7/ziFxqTnUSc03OlhgZKH9f.`) when the user is not found causes both code branches to execute 2^10 Eksblowfish rounds.
   - Verified empirically in `scratch/test_integration.go`: nonexistent user latency (76.52ms) vs wrong password latency (76.53ms) achieves a balanced 1.00x ratio, neutralizing statistical network analysis.

2. **Account Status Enumeration Elimination**:
   - *From Observation 1.2*, evaluating `!user.IsActive` before `bcrypt.CompareHashAndPassword` leaks `USER_INACTIVE` to attackers who do not possess the password.
   - Moving `bcrypt.CompareHashAndPassword` ahead of `!user.IsActive` ensures that all incorrect passwords yield `INVALID_CREDENTIALS` regardless of account state.
   - Only when the password is verified as correct will `!user.IsActive` trigger `USER_INACTIVE`, preserving legitimate employee UX while blocking external enumeration. Existing unit test `login_test.go:141` continues to pass without regression.

3. **Rate Limiting Enforcement**:
   - *From Observation 1.3*, `DEVELOPMENT_RULES.md` line 323 requires `httprate` on auth endpoints.
   - Using `github.com/go-chi/httprate` (v0.16.0) with `httprate.Limit(10, 1*time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP), httprate.WithLimitHandler(...))` enforces a sliding window of 10 requests per minute per IP.
   - Combining this with `chimiddleware.RealIP` ensures IP addresses behind reverse proxies are accurately tracked and IPv6 addresses are subnet-canonicalized.
   - Formatting throttled responses via `dto.WriteError(w, domainErrors.ErrTooManyRequests)` strictly adheres to the standard `{ "error": { "code": "TOO_MANY_REQUESTS", "message": "..." } }` JSON envelope contract with HTTP status 429.
   - Verified empirically in `scratch/test_integration.go`: over 50 rapid requests, exactly 10 are allowed (200/401) and 40 are throttled with HTTP 429.

---

## 3. Caveats

1. **Repository-Wide Test Failure (`check_in_test.go`)**:
   - `asesamo-api/internal/core/usecase/access/check_in_test.go:175:34` currently fails to compile because `*mockAPRepo` does not implement `ListByPlant`. While this is outside the `auth` and `router` scope of this task (it belongs to the M2 AccessPoint / Check-In mock remediation assigned to peer workers), it currently blocks `make test` across the whole repository. The auth and http adapter packages themselves compile and pass tests cleanly.
2. **Reverse Proxy IP Trust Model in Production**:
   - `chimiddleware.RealIP` reads `X-Forwarded-For` and `X-Real-IP`. In production, this requires an upstream reverse proxy (Nginx, Traefik, Cloudflare) that strips or replaces untrusted client-supplied headers. In local Docker and staging environments, `r.RemoteAddr` will default to client or container IPs.
3. **In-Memory Rate Limiting Scope**:
   - `httprate` stores rate counters in process RAM. If multiple backend replicas are deployed in production, rate limiting operates per-instance. For multi-replica deployments in Phase 2, `httprate-redis` can be swapped in without modifying route definitions.

---

## 4. Conclusion

All three vulnerabilities are completely characterized, tested, and resolved through precise code changes:
1. **Timing Side-Channel**: Mitigated via `dummyBcryptHash` comparison on `domainErrors.IsNotFound(err)` and `user == nil`, reducing the timing disparity from 7,601x to 1.00x.
2. **Account Status Enumeration**: Mitigated by reordering verification so `bcrypt.CompareHashAndPassword` executes before `!user.IsActive`.
3. **Rate Limiting**: Mitigated by mounting `httprate.Limit(10, 1*time.Minute, ...)` on `POST /api/v1/auth/login`, returning HTTP 429 and standard JSON error envelope on excess attempts.

All patch specifications are detailed in `analysis.md` and machine-readable in `remediation.patch`.

---

## 5. Verification Method

Independent verification steps:
1. **Inspect Patch and Implementation Details**:
   - Review `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2/analysis.md`
   - Review `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2/remediation.patch`
2. **Execute Empirical Scratch Test Suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2/scratch
   go run test_integration.go
   ```
   *Expected Output*:
   ```
   [TEST 1 PASSED] Inactive user with wrong password returned ErrInvalidCredentials.
   [TEST 2 PASSED] Inactive user with correct password returned ErrUserInactive.
   [TEST 3] Latency non-existent: ~76ms, wrong pass: ~76ms, ratio: ~1.00x
   [TEST 3 PASSED] Latencies are balanced and indistinguishable!
   [TEST 4] Status counts over 50 requests: map[200:10 429:40]
   [TEST 4 PASSED] Exactly 10 requests allowed, 40 requests throttled with 429!
   ALL VERIFICATIONS PASSED SUCCESSFULLY.
   ```
3. **Verify Auth Unit Test Suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v ./internal/core/usecase/auth/...
   ```
4. **Invalidation Conditions**:
   - If non-existent user latency remains sub-millisecond while existing user takes >50ms, the dummy bcrypt comparison was bypassed.
   - If an inactive account returns `USER_INACTIVE` when supplied with an invalid password, password verification was not placed first.
   - If 50 rapid login requests to `POST /api/v1/auth/login` from one IP all return 401 without receiving HTTP 429, rate limiting is not active on the router.
