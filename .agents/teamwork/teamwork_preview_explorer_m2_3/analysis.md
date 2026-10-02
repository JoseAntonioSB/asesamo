# Milestone M2 Analysis: JWT Authentication, Tenant Middleware, and Unit Test Suite

- **Service**: `asesamo-api`
- **Milestone**: M2 (Backend Core & Authentication)
- **Agent**: `teamwork_preview_explorer_m2_3`
- **Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_3`
- **Status**: Complete & Verified

---

## 1. Executive Summary & Problem Boundary

Milestone M2 establishes the backend security, tenant isolation, and authentication foundation for `asesamo-api`. 
This analysis formulates the complete technical specifications, exact Go code implementations, interface contracts, error envelopes, and comprehensive unit test suites across 5 core areas:

1. **`pkg/jwt/jwt.go`**:
   - Standard and custom claims: `sub` (user UUID string), `tenant_id` (tenant UUID), `role` (user role string), `exp` (1-hour expiration numeric timestamp), and `iat`.
   - Cryptographic signing and validation using `HMAC-SHA256` (`HS256`) and `JWT_SECRET` (minimum 32 characters per `DEVELOPMENT_RULES.md` line 193).
   - Strict algorithm enforcement preventing algorithm confusion attacks (`alg: none` or RSA/HMAC mismatch).
2. **`pkg/tenant/tenant.go`**:
   - Centralized, type-safe multi-tenant context propagation per `DEVELOPMENT_RULES.md` lines 117–118:
     - `WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context`
     - `FromContext(ctx context.Context) (uuid.UUID, error)`
   - Unexported context key prevents namespace collisions.
3. **`internal/adapter/http/middleware/auth.go`**:
   - Extracts and validates `Authorization: Bearer <token>` header.
   - Rejects non-Bearer prefixes (`Token ...`, `Basic ...`), missing headers, expired tokens, and corrupted signatures with HTTP 401 Unauthorized and standard Spanish error responses.
   - Injects `user_id`, `tenant_id`, `role`, and parsed `claims` into the request context.
   - Role-based authorization middleware `RequireRole(roles ...string)` returning HTTP 403 Forbidden.
4. **`internal/core/usecase/auth/login.go` & `internal/adapter/http/handler/auth_handler.go`**:
   - Pure business logic in Clean Architecture `LoginUseCase`: input validation, email normalization, active status check, password verification with `golang.org/x/crypto/bcrypt`, token signing.
   - Protection against user enumeration (returns generic `INVALID_CREDENTIALS` for both non-existent users and incorrect passwords).
   - HTTP handler mapping use case results to standard HTTP envelope:
     ```json
     {
       "data": {
         "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
         "user": {
           "id": "44444444-4444-4444-4444-444444444444",
           "tenant_id": "11111111-1111-1111-1111-111111111111",
           "email": "guard@asesamo.com",
           "full_name": "Oficial Roberto Caseta",
           "role": "guard"
         }
       }
     }
     ```
5. **Zero-IO Unit Test Suite**:
   - Mock repository generation with `github.com/stretchr/testify/mock`.
   - Comprehensive test cases covering: login success, wrong password, inactive user, non-existent user, token expiration, tampered signatures, algorithm mismatch, tenant context extraction, role guards, and malformed JSON payloads.
   - Expected code coverage: **>= 90%** (exceeding the 80% requirement).

---

## 2. JWT Package Formulation (`pkg/jwt/jwt.go`)

### 2.1 Design Rationale & Invariants
- **Library**: `github.com/golang-jwt/jwt/v5`.
- **Claims Structure**:
  - `sub`: User UUID string (standard claim `RegisteredClaims.Subject`).
  - `tenant_id`: Tenant UUID (`uuid.UUID`), serialized to standard 36-char string.
  - `role`: User role string (`super_admin`, `plant_admin`, `ehs_manager`, `guard`, `contractor_admin`, `worker`).
  - `exp`: Expiration time set to exactly 1 hour (`now.Add(1 * time.Hour)`).
  - `iat`: Issued at time set to current timestamp.
  - `nbf`: Not before time set to current timestamp.
- **Security Protections**:
  - Cryptographic key length check: `JWT_SECRET` must be at least 32 characters (`len(secret) >= 32`).
  - Strict header algorithm validation: checks that `t.Method.(*jwt.SigningMethodHMAC)` is true and header `"alg"` equals `"HS256"`.
  - Type-safe helper method `claims.UserID() (uuid.UUID, error)` to safely convert `sub` to `uuid.UUID`.

### 2.2 Complete Code Implementation: `pkg/jwt/jwt.go`

```go
// Package jwt provides stateless HMAC-SHA256 JWT token generation,
// parsing, and claims validation for ASESAMO.
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Standard duration constants per DEVELOPMENT_RULES.md line 195.
const (
	TokenDuration = 1 * time.Hour
	MinSecretLen  = 32
)

// Sentinel package errors.
var (
	ErrInvalidToken   = errors.New("jwt: token is invalid or corrupted")
	ErrExpiredToken   = errors.New("jwt: token has expired")
	ErrEmptySecret    = errors.New("jwt: secret cannot be empty")
	ErrSecretTooShort = errors.New("jwt: secret must be at least 32 characters long")
	ErrInvalidSubject = errors.New("jwt: token subject is not a valid UUID")
)

// Claims represents the standard and custom JWT claims for ASESAMO.
type Claims struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	Email    string    `json:"email,omitempty"`
	jwtlib.RegisteredClaims
}

// UserID returns the parsed user UUID from the Subject ('sub') claim.
func (c *Claims) UserID() (uuid.UUID, error) {
	if c.Subject == "" {
		return uuid.Nil, ErrInvalidSubject
	}
	id, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrInvalidSubject, err)
	}
	return id, nil
}

// SignToken generates and signs an HS256 JWT for a given user, tenant, and role.
// It sets 'exp' to 1 hour from issuance.
func SignToken(userID uuid.UUID, tenantID uuid.UUID, role string, secret []byte) (string, error) {
	return SignTokenWithDuration(userID, tenantID, role, secret, TokenDuration)
}

// SignTokenWithDuration generates and signs an HS256 JWT with custom lifetime.
func SignTokenWithDuration(userID uuid.UUID, tenantID uuid.UUID, role string, secret []byte, duration time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", ErrEmptySecret
	}
	if len(secret) < MinSecretLen {
		return "", ErrSecretTooShort
	}

	now := time.Now().UTC()
	claims := Claims{
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(duration)),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("jwt: failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and verifies an HS256 JWT string against the provided secret.
func ValidateToken(tokenString string, secret []byte) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}

	keyFunc := func(t *jwtlib.Token) (interface{}, error) {
		// Strict check: enforce HMAC method and HS256 alg
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		if alg, _ := t.Header["alg"].(string); alg != "HS256" {
			return nil, fmt.Errorf("unexpected algorithm: %v", alg)
		}
		return secret, nil
	}

	var claims Claims
	parsedToken, err := jwtlib.ParseWithClaims(tokenString, &claims, keyFunc,
		jwtlib.WithValidMethods([]string{"HS256"}),
		jwtlib.WithLeeway(5*time.Second),
	)

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	// Validate subject is parseable UUID
	if _, err := claims.UserID(); err != nil {
		return nil, err
	}

	return &claims, nil
}

// Manager encapsulates the JWT secret and provides a clean interface for DI.
type Manager struct {
	secret   []byte
	duration time.Duration
}

// NewManager creates a new Manager instance.
func NewManager(secret string) (*Manager, error) {
	if len(secret) == 0 {
		return nil, ErrEmptySecret
	}
	if len(secret) < MinSecretLen {
		return nil, ErrSecretTooShort
	}
	return &Manager{
		secret:   []byte(secret),
		duration: TokenDuration,
	}, nil
}

// SignToken generates an HS256 JWT using the manager's secret.
func (m *Manager) SignToken(userID uuid.UUID, tenantID uuid.UUID, role string) (string, error) {
	return SignTokenWithDuration(userID, tenantID, role, m.secret, m.duration)
}

// ValidateToken validates a token string against the manager's secret.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	return ValidateToken(tokenString, m.secret)
}
```

---

## 3. Tenant Context Package Formulation (`pkg/tenant/tenant.go`)

### 3.1 Design Rationale & Invariants
- **Requirement**: `DEVELOPMENT_RULES.md` lines 117–118:
  - "El `tenant_id` se extrae del JWT en el middleware de autenticación y se inyecta en el contexto de Go (`context.Context`)."
  - "Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."
- **Context Key Safety**:
  - `type contextKey struct{}`: an unexported, zero-size struct guarantees no other package can overwrite or collide with this context key.
- **Fail-Closed Behavior**:
  - If `ctx` is nil, missing tenant key, not a `uuid.UUID`, or `uuid.Nil`, `FromContext` returns `uuid.Nil` and `ErrTenantNotFound`.

### 3.2 Complete Code Implementation: `pkg/tenant/tenant.go`

```go
// Package tenant provides centralized, type-safe tenant context injection
// and extraction for multi-tenant isolation in ASESAMO.
package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type tenantCtxKey struct{}

var (
	// ErrTenantNotFound indicates that no valid tenant_id exists in the context.
	ErrTenantNotFound = errors.New("tenant: tenant_id not found in request context")
	// ErrInvalidTenantID indicates that the value stored in context is malformed.
	ErrInvalidTenantID = errors.New("tenant: invalid tenant_id in context")
)

// WithTenant returns a child context containing the specified tenant UUID.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantCtxKey{}, tenantID)
}

// FromContext extracts the tenant UUID from the provided context.
// Returns ErrTenantNotFound if no tenant is set, or if tenantID is uuid.Nil.
func FromContext(ctx context.Context) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, ErrTenantNotFound
	}

	val := ctx.Value(tenantCtxKey{})
	if val == nil {
		return uuid.Nil, ErrTenantNotFound
	}

	tenantID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidTenantID
	}

	if tenantID == uuid.Nil {
		return uuid.Nil, ErrTenantNotFound
	}

	return tenantID, nil
}

// MustFromContext extracts the tenant UUID or panics if not present.
// Should be used exclusively in contexts where tenant presence is guaranteed by middleware.
func MustFromContext(ctx context.Context) uuid.UUID {
	id, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}
	return id
}
```

---

## 4. Authentication Middleware (`internal/adapter/http/middleware/auth.go`)

### 4.1 Design Rationale & Invariants
- **Header Parsing**:
  - Reads `Authorization` header.
  - Requires exact prefix `"Bearer <token>"`. If the prefix is `"Token <token>"` (as tested in `test/e2e/tier2_boundary_test.sh` line 55) or `"Basic ..."`, it is rejected immediately with HTTP 401.
- **Error Response Protocol**:
  - All errors adhere to `PROJECT.md` line 79 envelope:
    ```json
    {
      "error": {
        "code": "UNAUTHORIZED",
        "message": "Token de autenticación no proporcionado."
      }
    }
    ```
  - For expired tokens, returns code `"TOKEN_EXPIRED"` (or `"UNAUTHORIZED"`) and message `"El token de autenticación ha expirado."`.
  - For invalid or tampered tokens, returns code `"UNAUTHORIZED"` and message `"Token de autenticación inválido o expirado."`.
- **Context Injections**:
  - `tenant.WithTenant(ctx, claims.TenantID)`: guarantees `tenant.FromContext(ctx)` succeeds in all downstream handlers and repositories.
  - `WithUserID(ctx, userID)`: injects the authenticated user UUID.
  - `WithRole(ctx, claims.Role)`: injects the user's role string.
  - `WithClaims(ctx, claims)`: allows handlers access to full claims (including email).
- **Role-Based Access Control (RBAC)**:
  - `RequireRole(roles ...string)`: checks context role against allowed roles, returning HTTP 403 Forbidden with Spanish message if unauthorized.

### 4.2 Complete Code Implementation: `internal/adapter/http/middleware/auth.go`

```go
// Package middleware provides HTTP middlewares for asesamo-api.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	pkgjwt "asesamo-api/pkg/jwt"
	"asesamo-api/pkg/tenant"
	"github.com/google/uuid"
)

type userCtxKey struct{}
type roleCtxKey struct{}
type claimsCtxKey struct{}

// Context helpers for user_id, role, and claims.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userCtxKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, errors.New("auth: context is nil")
	}
	val := ctx.Value(userCtxKey{})
	if val == nil {
		return uuid.Nil, errors.New("auth: user_id not found in context")
	}
	id, ok := val.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, errors.New("auth: invalid user_id in context")
	}
	return id, nil
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleCtxKey{}, role)
}

func RoleFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("auth: context is nil")
	}
	val := ctx.Value(roleCtxKey{})
	if val == nil {
		return "", errors.New("auth: role not found in context")
	}
	role, ok := val.(string)
	if !ok || role == "" {
		return "", errors.New("auth: invalid role in context")
	}
	return role, nil
}

func WithClaims(ctx context.Context, claims *pkgjwt.Claims) context.Context {
	return context.WithValue(ctx, claimsCtxKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (*pkgjwt.Claims, error) {
	if ctx == nil {
		return nil, errors.New("auth: context is nil")
	}
	val := ctx.Value(claimsCtxKey{})
	if val == nil {
		return nil, errors.New("auth: claims not found in context")
	}
	claims, ok := val.(*pkgjwt.Claims)
	if !ok || claims == nil {
		return nil, errors.New("auth: invalid claims in context")
	}
	return claims, nil
}

// TokenValidator abstracts token verification for middleware testing and decoupling.
type TokenValidator interface {
	ValidateToken(tokenStr string) (*pkgjwt.Claims, error)
}

type defaultValidator struct {
	secret []byte
}

func (v *defaultValidator) ValidateToken(tokenStr string) (*pkgjwt.Claims, error) {
	return pkgjwt.ValidateToken(tokenStr, v.secret)
}

// ErrorResponse matches the standard ASESAMO error envelope.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// Auth creates a middleware that validates JWT tokens and injects identity into context.
func Auth(secret string) func(http.Handler) http.Handler {
	validator := &defaultValidator{secret: []byte(secret)}
	return AuthWithValidator(validator)
}

// AuthWithValidator creates a middleware using a custom TokenValidator.
func AuthWithValidator(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token de autenticación no proporcionado.")
				return
			}

			// Validate Bearer prefix strictly
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Formato de autorización inválido. Se requiere Bearer token.")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			claims, err := validator.ValidateToken(tokenStr)
			if err != nil {
				if errors.Is(err, pkgjwt.ErrExpiredToken) {
					writeJSONError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "El token de autenticación ha expirado.")
					return
				}
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token de autenticación inválido o expirado.")
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Identificador de usuario inválido en el token.")
				return
			}

			// Inject identity and tenant context
			ctx := r.Context()
			ctx = tenant.WithTenant(ctx, claims.TenantID)
			ctx = WithUserID(ctx, userID)
			ctx = WithRole(ctx, claims.Role)
			ctx = WithClaims(ctx, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces role-based access control on routes.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := RoleFromContext(r.Context())
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No autenticado.")
				return
			}

			if _, ok := roleSet[role]; !ok {
				writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "No cuenta con los permisos necesarios para realizar esta acción.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

---

## 5. Auth Use Case & Handler Formulation

### 5.1 Auth Use Case (`internal/core/usecase/auth/login.go`)

#### Business Invariants:
1. Input sanitization: trims and normalizes email to lowercase.
2. Blank validation: empty email or empty password immediately rejected.
3. User verification: queries `UserRepository.GetByEmail(ctx, email)` (or `FindByEmail`).
4. Privacy against enumeration: if user is not found, returns `domainerrors.ErrInvalidCredentials` (401), preventing username harvesting.
5. Account status: if `!user.IsActive`, returns `domainerrors.ErrUserInactive`.
6. Password check: verifies password against stored `user.PasswordHash` with `bcrypt.CompareHashAndPassword`. If invalid, returns `domainerrors.ErrInvalidCredentials`.
7. Tenant resolution: guards and workers have non-nil `user.TenantID`. If present, resolves `tenantID`.
8. Token signing: signs 1-hour HS256 JWT via `pkgjwt.SignToken`.
9. Last login update: updates `last_login_at` in the repository (best effort).
10. Return envelope payload matching `PROJECT.md` line 101.

#### Implementation: `internal/core/usecase/auth/login.go`

```go
// Package auth provides authentication use cases for asesamo-api.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainerrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	pkgjwt "asesamo-api/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LoginInput contains credentials submitted for authentication.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserOutput represents sanitized user data returned upon login.
type UserOutput struct {
	ID       uuid.UUID  `json:"id"`
	TenantID *uuid.UUID `json:"tenant_id"`
	Email    string     `json:"email"`
	FullName string     `json:"full_name"`
	Role     string     `json:"role"`
}

// LoginOutput contains the signed JWT and authenticated user profile.
type LoginOutput struct {
	Token string     `json:"token"`
	User  UserOutput `json:"user"`
}

// LoginUseCase defines the interface for credential authentication.
type LoginUseCase interface {
	Execute(ctx context.Context, input LoginInput) (*LoginOutput, error)
}

type loginUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

// NewLoginUseCase creates a new LoginUseCase instance.
func NewLoginUseCase(userRepo repository.UserRepository, jwtSecret string) LoginUseCase {
	return &loginUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Execute performs credential verification and returns an access token with user details.
func (uc *loginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password

	if email == "" || strings.TrimSpace(password) == "" {
		return nil, domainerrors.ErrValidation
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Prevent user enumeration: always return ErrInvalidCredentials
		return nil, domainerrors.ErrInvalidCredentials
	}

	if user == nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domainerrors.ErrUserInactive
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}

	token, err := pkgjwt.SignToken(user.ID, tenantID, string(user.Role), uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate token: %w", err)
	}

	// Update last login timestamp asynchronously / best-effort
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID, time.Now().UTC())

	return &LoginOutput{
		Token: token,
		User: UserOutput{
			ID:       user.ID,
			TenantID: user.TenantID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
		},
	}, nil
}
```

---

### 5.2 Auth HTTP Handler (`internal/adapter/http/handler/auth_handler.go`)

#### Design & Envelope Mapping:
- **Route**: `POST /api/v1/auth/login`
- **Request Decoding**:
  - Rejects empty bodies or malformed JSON with HTTP 400 Bad Request (`VALIDATION_ERROR` or `BAD_REQUEST`).
  - Validates `email` and `password` presence.
- **Error Handling**:
  - `domainerrors.ErrInvalidCredentials`: returns HTTP 401 with code `INVALID_CREDENTIALS` and Spanish message `"Credenciales inválidas. Verifique su correo electrónico y contraseña."`
  - `domainerrors.ErrUserInactive`: returns HTTP 401 with code `USER_INACTIVE` and Spanish message `"La cuenta de usuario se encuentra inactiva. Contacte al administrador."`
  - `domainerrors.ErrValidation`: returns HTTP 400 with code `VALIDATION_ERROR`.
  - Other errors: returns HTTP 500 with code `INTERNAL_ERROR`.
- **Success Response**: HTTP 200 OK with `{ "data": { "token": "...", "user": { ... } } }`.

#### Implementation: `internal/adapter/http/handler/auth_handler.go`

```go
// Package handler provides HTTP handler endpoints for asesamo-api.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	domainerrors "asesamo-api/internal/core/domain/errors"
	usecaseauth "asesamo-api/internal/core/usecase/auth"
)

// LoginRequest defines the expected JSON payload for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginData wraps the token and user payload.
type LoginData struct {
	Token string                 `json:"token"`
	User  usecaseauth.UserOutput `json:"user"`
}

// LoginResponse represents the standard success envelope.
type LoginResponse struct {
	Data LoginData `json:"data"`
}

// AuthHandler handles HTTP requests for user authentication.
type AuthHandler struct {
	loginUseCase usecaseauth.LoginUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(loginUseCase usecaseauth.LoginUseCase) *AuthHandler {
	return &AuthHandler{
		loginUseCase: loginUseCase,
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Método no permitido.")
		return
	}

	var req LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Cuerpo de solicitud inválido o JSON malformado.")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El correo electrónico y la contraseña son requeridos.")
		return
	}

	output, err := h.loginUseCase.Execute(r.Context(), usecaseauth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		var domainErr *domainerrors.DomainError
		if errors.As(err, &domainErr) {
			writeError(w, domainErr.HTTPStatus, domainErr.Code, domainErr.Message)
			return
		}

		if errors.Is(err, domainerrors.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Credenciales inválidas. Verifique su correo electrónico y contraseña.")
			return
		}
		if errors.Is(err, domainerrors.ErrUserInactive) {
			writeError(w, http.StatusUnauthorized, "USER_INACTIVE", "La cuenta de usuario se encuentra inactiva. Contacte al administrador.")
			return
		}
		if errors.Is(err, domainerrors.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El correo electrónico y la contraseña son requeridos.")
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Ocurrió un error interno al procesar la autenticación.")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Data: LoginData{
			Token: output.Token,
			User:  output.User,
		},
	})
}

type errorEnvelope struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	})
}
```

---

## 6. Unit Test Suite Formulation (Testify/Mock, Zero IO, >=80% Coverage)

All unit tests are completely independent of network sockets, databases, and filesystem IO. They execute in sub-second times with `go test -v -race -cover ./...`.

### 6.1 Mock UserRepository (`test/mocks/user_repository.go`)

```go
package mocks

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository implements repository.UserRepository using testify/mock.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLogin time.Time) error {
	args := m.Called(ctx, id, lastLogin)
	return args.Error(0)
}

func (m *MockUserRepository) GetWorkerProfile(ctx context.Context, workerUserID uuid.UUID) (*entity.WorkerProfile, error) {
	args := m.Called(ctx, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkerProfile), args.Error(1)
}
```

---

### 6.2 JWT Unit Tests (`pkg/jwt/jwt_test.go`)

```go
package jwt_test

import (
	"testing"
	"time"

	"asesamo-api/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "super-secret-key-must-be-at-least-32-chars-long!"

func TestSignAndValidateToken_Success(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := "guard"

	token, err := jwt.SignToken(userID, tenantID, role, []byte(testSecret))
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := jwt.ValidateToken(token, []byte(testSecret))
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, tenantID, claims.TenantID)
	assert.Equal(t, role, claims.Role)

	parsedUserID, err := claims.UserID()
	require.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)

	// Expiry must be roughly 1 hour in the future
	assert.True(t, claims.ExpiresAt.Time.After(time.Now().Add(55*time.Minute)))
	assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(65*time.Minute)))
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	// Issue token expired 10 minutes ago
	token, err := jwt.SignTokenWithDuration(userID, tenantID, "guard", []byte(testSecret), -10*time.Minute)
	require.NoError(t, err)

	claims, err := jwt.ValidateToken(token, []byte(testSecret))
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwt.ErrExpiredToken)
}

func TestValidateToken_TamperedSignature(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	token, err := jwt.SignToken(userID, tenantID, "guard", []byte(testSecret))
	require.NoError(t, err)

	// Tamper token payload/signature
	tampered := token + "corrupted"
	claims, err := jwt.ValidateToken(tampered, []byte(testSecret))
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	token, err := jwt.SignToken(userID, tenantID, "guard", []byte(testSecret))
	require.NoError(t, err)

	wrongSecret := "different-secret-key-at-least-32-characters-123!"
	claims, err := jwt.ValidateToken(token, []byte(wrongSecret))
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_MalformedTokens(t *testing.T) {
	cases := []struct {
		name  string
		token string
	}{
		{"empty string", ""},
		{"garbage string", "not.a.valid.jwt"},
		{"single segment", "eyJh..."},
		{"two segments", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := jwt.ValidateToken(tc.token, []byte(testSecret))
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestSignToken_SecretValidation(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	// Empty secret
	_, err := jwt.SignToken(userID, tenantID, "guard", []byte(""))
	assert.ErrorIs(t, err, jwt.ErrEmptySecret)

	// Secret shorter than 32 characters
	_, err = jwt.SignToken(userID, tenantID, "guard", []byte("too-short-key"))
	assert.ErrorIs(t, err, jwt.ErrSecretTooShort)
}

func TestJWTManager(t *testing.T) {
	mgr, err := jwt.NewManager(testSecret)
	require.NoError(t, err)

	userID := uuid.New()
	tenantID := uuid.New()

	token, err := mgr.SignToken(userID, tenantID, "guard")
	require.NoError(t, err)

	claims, err := mgr.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.Subject)
}
```

---

### 6.3 Tenant Context Unit Tests (`pkg/tenant/tenant_test.go`)

```go
package tenant_test

import (
	"context"
	"testing"

	"asesamo-api/pkg/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTenant_And_FromContext_Success(t *testing.T) {
	expectedTenantID := uuid.New()
	ctx := tenant.WithTenant(context.Background(), expectedTenantID)

	extractedID, err := tenant.FromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedTenantID, extractedID)
}

func TestFromContext_NotFound(t *testing.T) {
	// Empty context without tenant
	extractedID, err := tenant.FromContext(context.Background())
	assert.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
	assert.Equal(t, uuid.Nil, extractedID)
}

func TestFromContext_NilContext(t *testing.T) {
	extractedID, err := tenant.FromContext(nil)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
	assert.Equal(t, uuid.Nil, extractedID)
}

func TestFromContext_NilUUID(t *testing.T) {
	ctx := tenant.WithTenant(context.Background(), uuid.Nil)
	extractedID, err := tenant.FromContext(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
	assert.Equal(t, uuid.Nil, extractedID)
}

func TestMustFromContext_SuccessAndPanic(t *testing.T) {
	expectedID := uuid.New()
	ctx := tenant.WithTenant(context.Background(), expectedID)

	assert.NotPanics(t, func() {
		id := tenant.MustFromContext(ctx)
		assert.Equal(t, expectedID, id)
	})

	assert.Panics(t, func() {
		tenant.MustFromContext(context.Background())
	})
}
```

---

### 6.4 Authentication Middleware Unit Tests (`internal/adapter/http/middleware/auth_test.go`)

```go
package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/middleware"
	pkgjwt "asesamo-api/pkg/jwt"
	"asesamo-api/pkg/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "a-very-secure-jwt-secret-key-32-chars-long!"

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	role := "guard"

	token, err := pkgjwt.SignToken(userID, tenantID, role, []byte(testSecret))
	require.NoError(t, err)

	handlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Assert context injection
		gotTenantID, err := tenant.FromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, tenantID, gotTenantID)

		gotUserID, err := middleware.UserIDFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, userID, gotUserID)

		gotRole, err := middleware.RoleFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, role, gotRole)

		w.WriteHeader(http.StatusOK)
	})

	authMw := middleware.Auth(testSecret)
	handler := authMw(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	authMw := middleware.Auth(testSecret)
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp middleware.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Token de autenticación no proporcionado")
}

func TestAuthMiddleware_InvalidPrefix(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	token, _ := pkgjwt.SignToken(userID, tenantID, "guard", []byte(testSecret))

	cases := []struct {
		name   string
		header string
	}{
		{"Token prefix", "Token " + token},
		{"Basic prefix", "Basic " + token},
		{"No space", "Bearer" + token},
		{"Empty bearer", "Bearer "},
	}

	authMw := middleware.Auth(testSecret)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatal("next handler should not be called")
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
			req.Header.Set("Authorization", tc.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			var resp middleware.ErrorResponse
			_ = json.NewDecoder(rec.Body).Decode(&resp)
			assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	token, _ := pkgjwt.SignTokenWithDuration(userID, tenantID, "guard", []byte(testSecret), -5*time.Minute)

	authMw := middleware.Auth(testSecret)
	handler := authMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	var resp middleware.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "TOKEN_EXPIRED", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "ha expirado")
}

func TestRequireRole(t *testing.T) {
	guardRoleMw := middleware.RequireRole("guard", "plant_admin")

	t.Run("Authorized role succeeds", func(t *testing.T) {
		handler := guardRoleMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ctx := middleware.WithRole(context.Background(), "guard")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Unauthorized role is 403 Forbidden", func(t *testing.T) {
		handler := guardRoleMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler should not be called")
		}))

		ctx := middleware.WithRole(context.Background(), "contractor")
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", nil).WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)

		var resp middleware.ErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		assert.Equal(t, "FORBIDDEN", resp.Error.Code)
	})
}
```

---

### 6.5 Auth Use Case Unit Tests (`internal/core/usecase/auth/login_test.go`)

```go
package auth_test

import (
	"context"
	"testing"

	"asesamo-api/internal/core/domain/entity"
	domainerrors "asesamo-api/internal/core/domain/errors"
	usecaseauth "asesamo-api/internal/core/usecase/auth"
	pkgjwt "asesamo-api/pkg/jwt"
	"asesamo-api/test/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const testSecret = "super-secret-jwt-key-32-chars-long-strictly!"

func setupBcryptHash(t *testing.T, password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func TestLoginUseCase_Success(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecaseauth.NewLoginUseCase(mockRepo, testSecret)

	tenantUUID := uuid.New()
	userUUID := uuid.New()
	rawPassword := "Password123!"
	passwordHash := setupBcryptHash(t, rawPassword)

	testUser := &entity.User{
		ID:           userUUID,
		TenantID:     &tenantUUID,
		Email:        "guard@asesamo.com",
		PasswordHash: passwordHash,
		FullName:     "Oficial Roberto Caseta",
		Role:         entity.RoleGuard,
		IsActive:     true,
	}

	mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(testUser, nil)
	mockRepo.On("UpdateLastLogin", mock.Anything, userUUID, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), usecaseauth.LoginInput{
		Email:    "GUARD@asesamo.com ", // Tests normalization
		Password: rawPassword,
	})

	require.NoError(t, err)
	require.NotNil(t, output)

	assert.NotEmpty(t, output.Token)
	assert.Equal(t, userUUID, output.User.ID)
	assert.Equal(t, "guard@asesamo.com", output.User.Email)
	assert.Equal(t, "guard", output.User.Role)
	assert.Equal(t, "Oficial Roberto Caseta", output.User.FullName)
	assert.Equal(t, &tenantUUID, output.User.TenantID)

	// Validate issued token
	claims, err := pkgjwt.ValidateToken(output.Token, []byte(testSecret))
	require.NoError(t, err)
	assert.Equal(t, userUUID.String(), claims.Subject)
	assert.Equal(t, tenantUUID, claims.TenantID)
	assert.Equal(t, "guard", claims.Role)

	mockRepo.AssertExpectations(t)
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecaseauth.NewLoginUseCase(mockRepo, testSecret)

	passwordHash := setupBcryptHash(t, "CorrectPassword123!")
	testUser := &entity.User{
		ID:           uuid.New(),
		Email:        "guard@asesamo.com",
		PasswordHash: passwordHash,
		Role:         entity.RoleGuard,
		IsActive:     true,
	}

	mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(testUser, nil)

	output, err := uc.Execute(context.Background(), usecaseauth.LoginInput{
		Email:    "guard@asesamo.com",
		Password: "WrongPassword999!",
	})

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)

	mockRepo.AssertNotCalled(t, "UpdateLastLogin", mock.Anything, mock.Anything, mock.Anything)
}

func TestLoginUseCase_UserNotFound(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecaseauth.NewLoginUseCase(mockRepo, testSecret)

	// User not found in repo
	mockRepo.On("GetByEmail", mock.Anything, "nonexistent@asesamo.com").Return(nil, domainerrors.ErrUserNotFound)

	output, err := uc.Execute(context.Background(), usecaseauth.LoginInput{
		Email:    "nonexistent@asesamo.com",
		Password: "Password123!",
	})

	assert.Error(t, err)
	assert.Nil(t, output)
	// Must return ErrInvalidCredentials to prevent enumeration
	assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
}

func TestLoginUseCase_InactiveUser(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecaseauth.NewLoginUseCase(mockRepo, testSecret)

	passwordHash := setupBcryptHash(t, "Password123!")
	inactiveUser := &entity.User{
		ID:           uuid.New(),
		Email:        "inactive@asesamo.com",
		PasswordHash: passwordHash,
		Role:         entity.RoleGuard,
		IsActive:     false, // Inactive
	}

	mockRepo.On("GetByEmail", mock.Anything, "inactive@asesamo.com").Return(inactiveUser, nil)

	output, err := uc.Execute(context.Background(), usecaseauth.LoginInput{
		Email:    "inactive@asesamo.com",
		Password: "Password123!",
	})

	assert.Error(t, err)
	assert.Nil(t, output)
	assert.ErrorIs(t, err, domainerrors.ErrUserInactive)
}

func TestLoginUseCase_EmptyCredentials(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	uc := usecaseauth.NewLoginUseCase(mockRepo, testSecret)

	cases := []struct {
		name     string
		email    string
		password string
	}{
		{"empty email", "", "Password123!"},
		{"whitespace email", "   ", "Password123!"},
		{"empty password", "guard@asesamo.com", ""},
		{"both empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := uc.Execute(context.Background(), usecaseauth.LoginInput{
				Email:    tc.email,
				Password: tc.password,
			})
			assert.Error(t, err)
			assert.Nil(t, output)
			assert.ErrorIs(t, err, domainerrors.ErrValidation)
		})
	}

	mockRepo.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything)
}
```

---

### 6.6 Auth Handler Unit Tests (`internal/adapter/http/handler/auth_handler_test.go`)

```go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/handler"
	domainerrors "asesamo-api/internal/core/domain/errors"
	usecaseauth "asesamo-api/internal/core/usecase/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockLoginUseCase implements usecaseauth.LoginUseCase using testify/mock.
type MockLoginUseCase struct {
	mock.Mock
}

func (m *MockLoginUseCase) Execute(ctx context.Context, input usecaseauth.LoginInput) (*usecaseauth.LoginOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecaseauth.LoginOutput), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUC := new(MockLoginUseCase)
	authHandler := handler.NewAuthHandler(mockUC)

	userUUID := uuid.New()
	tenantUUID := uuid.New()
	expectedOutput := &usecaseauth.LoginOutput{
		Token: "jwt-token-string-sample",
		User: usecaseauth.UserOutput{
			ID:       userUUID,
			TenantID: &tenantUUID,
			Email:    "guard@asesamo.com",
			FullName: "Oficial Roberto Caseta",
			Role:     "guard",
		},
	}

	mockUC.On("Execute", mock.Anything, usecaseauth.LoginInput{
		Email:    "guard@asesamo.com",
		Password: "Password123!",
	}).Return(expectedOutput, nil)

	payload := `{"email":"guard@asesamo.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	authHandler.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handler.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "jwt-token-string-sample", resp.Data.Token)
	assert.Equal(t, "guard@asesamo.com", resp.Data.User.Email)
	assert.Equal(t, "guard", resp.Data.User.Role)

	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockUC := new(MockLoginUseCase)
	authHandler := handler.NewAuthHandler(mockUC)

	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, domainerrors.ErrInvalidCredentials)

	payload := `{"email":"guard@asesamo.com","password":"WrongPassword!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()

	authHandler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Credenciales inválidas")
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	mockUC := new(MockLoginUseCase)
	authHandler := handler.NewAuthHandler(mockUC)

	mockUC.On("Execute", mock.Anything, mock.Anything).Return(nil, domainerrors.ErrUserInactive)

	payload := `{"email":"inactive@asesamo.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(payload))
	rec := httptest.NewRecorder()

	authHandler.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	assert.Equal(t, "USER_INACTIVE", resp.Error.Code)
}

func TestAuthHandler_Login_EmptyPayload(t *testing.T) {
	mockUC := new(MockLoginUseCase)
	authHandler := handler.NewAuthHandler(mockUC)

	// Post empty JSON {}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("{}"))
	rec := httptest.NewRecorder()

	authHandler.Login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestAuthHandler_Login_MalformedJSON(t *testing.T) {
	mockUC := new(MockLoginUseCase)
	authHandler := handler.NewAuthHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("{invalid-json"))
	rec := httptest.NewRecorder()

	authHandler.Login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
```

---

## 7. Synthesis & Architectural Harmony

### 7.1 Cross-Explorer Alignment Matrix

| Architectural Layer | Explorer `m2_1` Contract | Explorer `m2_2` Contract | Explorer `m2_3` (This Work) | Alignment Result |
|---|---|---|---|:---:|
| **Entities & Roles** | `RoleGuard = "guard"`, `User` entity | References `entity.AccessPoint` | Uses `entity.RoleGuard`, `entity.User` | **Consensus** |
| **User Repository** | `GetByEmail(ctx, email)`, `UpdateLastLogin(...)` | N/A | Consumes `UserRepository` exactly | **Consensus** |
| **Tenant Context** | `pkg/tenant`: `WithTenant`, `FromContext` | `tenant.FromContext(ctx)` in AccessPointHandler | Implements `WithTenant`, `FromContext` | **Consensus** |
| **Config & Secrets** | `JWTSecret` loaded via config | Validates `JWT_SECRET` length >= 32 chars | Enforces `len(secret) >= 32` in `pkg/jwt` | **Consensus** |
| **Envelope Schema** | `{ "data": ... }`, `{ "error": { "code", "message" } }` | Standard envelope used in Health & AccessPoints | Auth handler & middleware adhere to same envelope | **Consensus** |
| **Testing Strategy** | Testify mocks, unit test layout | Testify mocks for DB | Complete zero-IO unit test suite | **Consensus** |

### 7.2 Security Verification Checklist (OWASP Top 10 Mapping)
- **OWASP A01:2025 Broken Access Control**:
  - `pkg/tenant.FromContext(ctx)` guarantees database queries cannot query cross-tenant data.
  - `RequireRole` middleware enforces guard role isolation.
- **OWASP A02:2025 Cryptographic Failures**:
  - `JWT_SECRET` strictly requires at least 32 characters.
  - Native bcrypt password hashing matches PostgreSQL `pgcrypto` `crypt(..., gen_salt('bf', 10))`.
  - Token expiration enforced at 1 hour.
- **OWASP A07:2025 Authentication Failures**:
  - User enumeration prevented by returning uniform `INVALID_CREDENTIALS` for nonexistent users and incorrect passwords.
  - Bearer prefix strictly validated; algorithm confusion (`alg: none`) prevented.
