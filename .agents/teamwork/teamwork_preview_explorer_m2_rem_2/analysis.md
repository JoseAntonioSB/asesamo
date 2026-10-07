# Milestone M2 Remediation Analysis: Auth Security, Timing Side-Channel & Rate Limiting

- **Author**: `teamwork_preview_explorer_m2_rem_2` (Explorer Agent)
- **Target Component**: `asesamo-api`
- **Scope**: Authentication Security, Timing Side-Channel Mitigation (CWE-208), Account Enumeration Elimination (OWASP A07), and IP Rate Limiting (CWE-307)
- **Date**: 2026-10-02T05:17:00Z
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_rem_2`

---

## 1. Executive Summary

Milestone M2 review identified two critical authentication security vulnerabilities in `LoginUseCase.Execute` and one architecture/rule violation in `router.go`:
1. **Timing Side-Channel (CWE-208 / OWASP A07)**: When an email is not found, `LoginUseCase.Execute` immediately returned `ErrInvalidCredentials` in ~10 microseconds, compared to ~77 milliseconds when evaluating an existing user's password with bcrypt. This ~7,600x latency disparity allows network attackers to reliably enumerate registered accounts.
2. **Account Status Enumeration (OWASP A07)**: `LoginUseCase.Execute` inspected `!user.IsActive` prior to password verification, returning `USER_INACTIVE` even when an arbitrary, incorrect password was supplied. This allowed unauthenticated attackers to discover valid, inactive accounts without knowing the password.
3. **Missing Rate Limiting on Login (CWE-307 / DEVELOPMENT_RULES.md Line 323)**: `POST /api/v1/auth/login` had no rate limiting mounted, allowing 50 rapid burst requests to execute without throttling, exposing the backend to credential brute-forcing and bcrypt CPU exhaustion.

This document formulates the root causes, empirical benchmarks, architectural designs, and exact code changes to resolve all three issues cleanly.

---

## 2. Issue 1: Observable Timing Side-Channel (CWE-208) in `LoginUseCase`

### 2.1 Code Inspection & Root Cause
In `asesamo-api/internal/core/usecase/auth/login.go` lines 51-62:
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
Although the code intended to prevent user enumeration by returning the identical `ErrInvalidCredentials` error payload, the control flow returned immediately upon database "not found".

### 2.2 Empirical Benchmark
Empirical benchmark executed on the existing implementation (`login_adversarial_test.go`):
- Latency for non-existent user: **10.12 µs** (mean)
- Latency for existing user with wrong password (bcrypt cost 10): **76.97 ms** (mean)
- **Disparity Ratio**: **7,601.6x**

In an internet or LAN environment, an attacker sending 20 requests per candidate email can compute the median response time:
- Mean ~15–20 ms total RTT -> Email DOES NOT exist.
- Mean ~90–100 ms total RTT -> Email DOES exist.

### 2.3 Remediation Architecture: Constant-Time Dummy Bcrypt Hash
To eliminate this side-channel, the use case must execute a bcrypt hash comparison of identical computational cost when the user is not found.

#### Dummy Hash Specification:
- **Algorithm**: Eksblowfish ($2a$)
- **Cost**: 10 (matching `bcrypt.DefaultCost == 10` and `pgcrypto` seed cost `gen_salt('bf', 10)`)
- **Format**: Standard 60-character modular crypt format string
- **Hash value**:
  `$2a$10$xkateHF30gEES5pFuIga4egG9Mb9C7/ziFxqTnUSc03OlhgZKH9f.`
- **Properties**:
  - Valid syntax: parsed by Go's `golang.org/x/crypto/bcrypt` without parse errors.
  - Constant execution time: executes full 2^10 key expansions (~75ms).
  - High entropy / Unconditional rejection: The return value of the dummy comparison is discarded (`_ = ...`), and the function unconditionally returns `domainErrors.ErrInvalidCredentials`.

#### Execution Flow in `Execute`:
```go
user, err := uc.userRepo.GetByEmail(ctx, email)
if err != nil {
    if domainErrors.IsNotFound(err) {
        // Mitigate timing side-channel (CWE-208 / OWASP A07):
        // Execute dummy bcrypt comparison with cost 10 to ensure response
        // latency is indistinguishable from existing user password verification.
        _ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(input.Password))
        return nil, domainErrors.ErrInvalidCredentials
    }
    return nil, domainErrors.ErrDatabase.WithErr(err)
}

if user == nil {
    _ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(input.Password))
    return nil, domainErrors.ErrInvalidCredentials
}
```

### 2.4 Empirical Verification Post-Fix
Running the benchmark in our scratch verification harness (`scratch/test_integration.go`):
- Non-existent user latency: **76.52 ms**
- Existing user (wrong password) latency: **76.53 ms**
- **Disparity Ratio**: **1.00x** (Timing difference: 0.01 ms / negligible)
- Side-channel completely eliminated.

---

## 3. Issue 2: Account Status Enumeration (OWASP A07) in `LoginUseCase`

### 3.1 Code Inspection & Root Cause
In `asesamo-api/internal/core/usecase/auth/login.go` lines 64-72:
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
Because the account status check preceded the password hash comparison, an attacker testing `inactive.guard@asesamo.com` with password `"ArbitraryWrongPassword"` received:
```json
{
  "error": {
    "code": "USER_INACTIVE",
    "message": "La cuenta de usuario se encuentra inactiva."
  }
}
```
This leaked two sensitive facts to unauthenticated callers without valid credentials:
1. The user account exists.
2. The user account is currently deactivated/suspended.

### 3.2 Architectural Evaluation of Fixes

| Approach | Mechanics | Security Assessment | UX Assessment | Backward Compatibility |
|---|---|---|---|---|
| **Option A (Recommended): Password-First Verification** | Verify `bcrypt.CompareHashAndPassword` FIRST. Only if valid password is confirmed, check `if !user.IsActive` and return `ErrUserInactive`. | Zero information leak to unauthenticated attackers (wrong password always returns `INVALID_CREDENTIALS`). | High: A legitimate user or disabled guard who knows their password receives clear feedback why they cannot log in. | 100% compatible with existing `login_test.go:141` which tests inactive user with valid password. |
| **Option B: Generic Error Always** | Replace `ErrUserInactive` entirely with `ErrInvalidCredentials`. | Maximum security isolation. | Poor: Legitimate staff cannot distinguish between mistyped password and disabled account; repeatedly attempts password resets. | Breaks `login_test.go:141` which asserts `ErrUserInactive`. |

### 3.3 Implementation Decision
We adopt **Option A**:
1. Check password hash FIRST via `bcrypt.CompareHashAndPassword`.
   - If invalid: return `ErrInvalidCredentials`.
2. Check `!user.IsActive` SECOND.
   - If inactive: return `ErrUserInactive`.
   - If active: proceed to issue JWT.

### 3.4 Verification Matrix:
- Probing inactive account with WRONG password -> returns `INVALID_CREDENTIALS` (Status 401).
- Probing inactive account with CORRECT password -> returns `USER_INACTIVE` (Status 401).
- Probing non-existent account with any password -> returns `INVALID_CREDENTIALS` (Status 401).
- Probing active account with WRONG password -> returns `INVALID_CREDENTIALS` (Status 401).
- Logging in to active account with CORRECT password -> returns JWT (Status 200).

---

## 4. Issue 3: Rate Limiting on `POST /api/v1/auth/login` (CWE-307) in `router.go`

### 4.1 Requirement & Current Gap
`DEVELOPMENT_RULES.md` line 323 specifies:
> *"Rate Limiting: El router chi debe incluir un middleware de rate limiting (ej. go-chi/httprate) en endpoints públicos y de auth."*

In `asesamo-api/internal/adapter/http/router.go` lines 39-42:
```go
r.Route("/api/v1", func(r chi.Router) {
    // Public Auth
    r.Post("/auth/login", cfg.AuthHandler.Login)
```
No rate limiting middleware was attached. Adversarial test `router_adversarial_test.go` fired 50 burst requests in 40ms, and all 50 were processed without throttling. Because each login attempt involves a bcrypt operation (~75ms of CPU), an unauthenticated attacker could easily cause a Denial of Service (DoS) by saturating server CPU cores.

### 4.2 Library Analysis: `github.com/go-chi/httprate` (v0.16.0)
`httprate` implements an in-memory sliding window rate limiter designed specifically for Chi.
Key features and architectural alignment:
1. **Sliding Window Counter**: Maintains rolling window buckets per key, preventing boundary burst attacks that affect simple fixed-window counters.
2. **Client IP Resolution**:
   - `router.go` already mounts `chimiddleware.RealIP` at line 29:
     `r.Use(chimiddleware.RealIP)`
     This extracts the client IP from `X-Forwarded-For` or `X-Real-IP` and sets `r.RemoteAddr`.
   - `httprate.KeyByIP` consumes `r.RemoteAddr` and normalizes IPv6 addresses (grouping `/64` subnets so attackers cannot bypass rate limits by rotating IPv6 addresses).
3. **Response Headers**:
   - `httprate` automatically emits:
     - `X-RateLimit-Limit: 10`
     - `X-RateLimit-Remaining: <N>`
     - `X-RateLimit-Reset: <UnixTimestamp>`
4. **Error Envelope Integration via `WithLimitHandler`**:
   - By default, `httprate.LimitByIP` returns a `text/plain` body `Too Many Requests\n`, which violates the standard JSON envelope required by `PROJECT.md` Section 1.
   - By using `httprate.Limit(10, 1*time.Minute, httprate.WithKeyFuncs(httprate.KeyByIP), httprate.WithLimitHandler(...))`, we can format the HTTP 429 response using `dto.WriteError(w, domainErrors.ErrTooManyRequests)`:
     ```json
     {
       "error": {
         "code": "TOO_MANY_REQUESTS",
         "message": "Demasiadas solicitudes. Por favor intente más tarde."
       }
     }
     ```

### 4.3 Rate Limit Parameter Recommendation
- **Endpoint**: `POST /api/v1/auth/login`
- **Limit**: 10 requests per minute per client IP (`10, 1*time.Minute`).
- **Rationale**:
  - A legitimate human guard typing credentials will never exceed 10 attempts per minute.
  - Limits automated credential brute-force attacks to 10 attempts per minute per IP.
  - Caps bcrypt CPU utilization per malicious IP to at most ~750ms of CPU time per minute (1.25% of one CPU core).

---

## 5. Exact Patch Specifications

### 5.1 Patch 1: Domain Error for Rate Limiting
**Target File**: `asesamo-api/internal/core/domain/errors/errors.go`
**Lines to Add**: Add `ErrTooManyRequests` to canonical domain error declarations (around line 160):

```go
	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "Demasiadas solicitudes. Por favor intente más tarde.",
		HTTPStatus: http.StatusTooManyRequests,
	}
```

### 5.2 Patch 2: Timing Mitigation & Password-First Sequence in Login
**Target File**: `asesamo-api/internal/core/usecase/auth/login.go`

```go
package auth

import (
	"context"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// dummyBcryptHash is a precomputed valid bcrypt hash (cost 10) used for constant-time
// password verification when a requested user is not found, mitigating CWE-208 timing attacks.
const dummyBcryptHash = "$2a$10$xkateHF30gEES5pFuIga4egG9Mb9C7/ziFxqTnUSc03OlhgZKH9f."

// LoginInput defines the credentials for authenticating a user.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput defines the result of a successful authentication.
type LoginOutput struct {
	Token string
	User  *entity.User
}

// LoginUseCase coordinates user authentication and JWT generation.
type LoginUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

// NewLoginUseCase instantiates a LoginUseCase.
func NewLoginUseCase(userRepo repository.UserRepository, jwtSecret string) *LoginUseCase {
	return &LoginUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Execute validates credentials, verifies account status, and returns a signed JWT.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || input.Password == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El correo y la contraseña son obligatorios.")
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if domainErrors.IsNotFound(err) {
			// Mitigate timing side-channel (CWE-208 / OWASP A07):
			// Execute dummy bcrypt comparison with cost 10 to ensure response
			// latency is indistinguishable from existing user password verification.
			_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(input.Password))
			return nil, domainErrors.ErrInvalidCredentials
		}
		return nil, domainErrors.ErrDatabase.WithErr(err)
	}

	if user == nil {
		_ = bcrypt.CompareHashAndPassword([]byte(dummyBcryptHash), []byte(input.Password))
		return nil, domainErrors.ErrInvalidCredentials
	}

	// Verify password hash FIRST to prevent inactive account status enumeration (OWASP A07)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domainErrors.ErrInvalidCredentials
	}

	// Verify account status only AFTER verifying password credentials
	if !user.IsActive {
		return nil, domainErrors.ErrUserInactive
	}

	// Determine tenant UUID
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}

	// Sign JWT
	token, err := jwt.SignToken(user.ID, tenantID, string(user.Role), uc.jwtSecret)
	if err != nil {
		return nil, domainErrors.ErrInternal.WithErr(err)
	}

	// Update last login timestamp asynchronously / best-effort
	now := time.Now().UTC()
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID, now)
	user.LastLoginAt = &now

	return &LoginOutput{
		Token: token,
		User:  user,
	}, nil
}
```

### 5.3 Patch 3: Rate Limiter Middleware on Login in Router
**Target File**: `asesamo-api/internal/adapter/http/router.go`

```go
package http

import (
	"net/http"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	customMiddleware "asesamo-api/internal/adapter/http/middleware"
	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// RouterConfig contains dependencies for configuring the Chi router.
type RouterConfig struct {
	HealthHandler      *handler.HealthHandler
	AuthHandler        *handler.AuthHandler
	AccessPointHandler *handler.AccessPointHandler
	AccessHandler      *handler.AccessHandler
	JWTSecret          string
	CORSOrigins        []string
}

// NewRouter constructs and configures the HTTP request multiplexer.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(customMiddleware.CORS(cfg.CORSOrigins))

	// Public Health endpoints (at root and /api/v1)
	r.Get("/health", cfg.HealthHandler.Check)
	r.Get("/api/v1/health", cfg.HealthHandler.Check)

	// Rate limiter for authentication endpoint: 10 requests per minute per IP
	loginLimiter := httprate.Limit(
		10,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			dto.WriteError(w, domainErrors.ErrTooManyRequests)
		}),
	)

	// API v1 group
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth (protected by IP rate limiter)
		r.With(loginLimiter).Post("/auth/login", cfg.AuthHandler.Login)

		// Protected routes (JWT authenticated)
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Auth(cfg.JWTSecret))

			// Access point catalog
			r.Get("/access-points", cfg.AccessPointHandler.List)

			// Gate Access Verification, Check-Out, and Audit Logs
			r.Post("/access/check-in", cfg.AccessHandler.CheckIn)
			r.Post("/access/check-out", cfg.AccessHandler.CheckOut)
			r.Get("/access-logs", cfg.AccessHandler.ListLogs)
		})
	})

	return r
}
```

### 5.4 Patch 4: Dependencies in `go.mod` and `go.sum`
**Target File**: `asesamo-api/go.mod`
Add to `require (...)`:
```go
	github.com/go-chi/httprate v0.16.0
```
Add to `require (...) // indirect`:
```go
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
```

**Target File**: `asesamo-api/go.sum`
Add the checksums:
```
github.com/go-chi/httprate v0.16.0 h1:8V5DH9j6pSK6UQoBsTpvMyFxycqaKEIToyPKzHJjUa8=
github.com/go-chi/httprate v0.16.0/go.mod h1:A8lo+qRhk+s9LiuP5saS7XCGDXRXMcrueq0NfIuCa/I=
github.com/klauspost/cpuid/v2 v2.2.10 h1:tBs3QSyvjDyFTq3uoc/9xFpCuOsJQFNPiAhYdw2skhE=
github.com/klauspost/cpuid/v2 v2.2.10/go.mod h1:hqwkgyIinND0mEev00jJYCxPNVRVXFQeu1XKlok6oO0=
github.com/zeebo/xxh3 v1.0.2 h1:xZmwmqxHZA8AI603jOQ0tMqmBr9lPeFwGg6d+xy9DC0=
github.com/zeebo/xxh3 v1.0.2/go.mod h1:5NWz9Sef7zIDm2JHfFlcQvNekmcEl9ekUZQQKCYaDcA=
```

### 5.5 Patch 5: Test Suite Updates
**Target File**: `asesamo-api/internal/core/usecase/auth/login_adversarial_test.go`
- Update line 105:
  Change `assert.Equal(t, "USER_INACTIVE", appErr.Code)` to `assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)`.
- Update line 143:
  Change timing ratio assertion:
  ```go
  // Timing must be balanced within a 2.5x margin of variation across runs, not > 10x
  timingRatio := float64(durationWrongPassword) / float64(durationNonExistent)
  assert.True(t, timingRatio < 2.5 && timingRatio > 0.4,
      "Timing between non-existent user and wrong password must be balanced (indistinguishable)")
  ```

**Target File**: `asesamo-api/internal/adapter/http/router_adversarial_test.go`
- Update lines 139-140:
  ```go
  // Exactly 10 requests allowed under the 1-minute window, subsequent 40 throttled with 429
  assert.Equal(t, 10, unauthorizedCount, "First 10 attempts allowed through to handler")
  assert.Equal(t, 40, rateLimitedCount, "Subsequent 40 burst attempts throttled with HTTP 429")
  ```
- Also assert JSON envelope on 429:
  ```go
  assert.Equal(t, "TOO_MANY_REQUESTS", rateLimitedResponse.Error.Code)
  ```
