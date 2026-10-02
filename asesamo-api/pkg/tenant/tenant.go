// Package tenant provides context-based tenant isolation helpers.
// In accordance with DEVELOPMENT_RULES.md (lines 117-118), tenant.FromContext(ctx)
// is the single, authoritative method to obtain tenant_id in downstream queries.
package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Sentinel context errors.
var (
	ErrTenantNotFound = errors.New("tenant: tenant_id not found in context")
	ErrUserIDNotFound = errors.New("tenant: user_id not found in context")
	ErrRoleNotFound   = errors.New("tenant: role not found in context")
)

// Unexported context key types to avoid key collisions.
type tenantKey struct{}
type userKey struct{}
type roleKey struct{}

// WithTenant injects the tenant UUID into the context.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// FromContext extracts the tenant UUID from the context.
// Returns ErrTenantNotFound if the key is missing or is uuid.Nil.
func FromContext(ctx context.Context) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, ErrTenantNotFound
	}
	val := ctx.Value(tenantKey{})
	if val == nil {
		return uuid.Nil, ErrTenantNotFound
	}
	tenantID, ok := val.(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, ErrTenantNotFound
	}
	return tenantID, nil
}

// WithUserID injects the authenticated user's UUID into the context.
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey{}, userID)
}

// UserIDFromContext extracts the user UUID from the context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, ErrUserIDNotFound
	}
	val := ctx.Value(userKey{})
	if val == nil {
		return uuid.Nil, ErrUserIDNotFound
	}
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, ErrUserIDNotFound
	}
	return userID, nil
}

// WithRole injects the user's role into the context.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey{}, role)
}

// RoleFromContext extracts the user's role from the context.
func RoleFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrRoleNotFound
	}
	val := ctx.Value(roleKey{})
	if val == nil {
		return "", ErrRoleNotFound
	}
	role, ok := val.(string)
	if !ok || role == "" {
		return "", ErrRoleNotFound
	}
	return role, nil
}
