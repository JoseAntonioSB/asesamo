// Package middleware provides HTTP middlewares for routing, security, and context enrichment.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/pkg/jwt"
	"asesamo-api/pkg/tenant"
)

// Auth returns a middleware that extracts, validates, and injects JWT identity into the request context.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	secretBytes := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				dto.WriteError(w, domainErrors.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				dto.WriteError(w, domainErrors.ErrUnauthorized)
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			claims, err := jwt.ValidateToken(tokenStr, secretBytes)
			if err != nil {
				if errors.Is(err, jwt.ErrExpiredToken) {
					dto.WriteError(w, domainErrors.ErrExpiredToken)
					return
				}
				dto.WriteError(w, domainErrors.ErrInvalidToken)
				return
			}

			userID, err := claims.UserID()
			if err != nil {
				dto.WriteError(w, domainErrors.ErrInvalidToken)
				return
			}

			// Inject identity into context
			ctx := r.Context()
			ctx = tenant.WithTenant(ctx, claims.TenantID)
			ctx = tenant.WithUserID(ctx, userID)
			ctx = tenant.WithRole(ctx, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole guards endpoints to ensure only users with specific roles are permitted.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := tenant.RoleFromContext(r.Context())
			if err != nil {
				dto.WriteError(w, domainErrors.ErrUnauthorized)
				return
			}

			for _, allowed := range allowedRoles {
				if strings.EqualFold(role, allowed) {
					next.ServeHTTP(w, r)
					return
				}
			}

			dto.WriteError(w, domainErrors.ErrForbidden)
		})
	}
}
