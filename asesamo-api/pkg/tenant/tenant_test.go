package tenant_test

import (
	"context"
	"testing"

	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantContext(t *testing.T) {
	t.Run("successful injection and retrieval of tenant_id", func(t *testing.T) {
		expectedID := uuid.New()
		ctx := tenant.WithTenant(context.Background(), expectedID)

		retrievedID, err := tenant.FromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedID, retrievedID)
	})

	t.Run("missing tenant_id returns ErrTenantNotFound", func(t *testing.T) {
		retrievedID, err := tenant.FromContext(context.Background())
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})

	t.Run("nil context returns ErrTenantNotFound", func(t *testing.T) {
		var nilCtx context.Context
		retrievedID, err := tenant.FromContext(nilCtx)
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})

	t.Run("nil UUID returns ErrTenantNotFound", func(t *testing.T) {
		ctx := tenant.WithTenant(context.Background(), uuid.Nil)
		retrievedID, err := tenant.FromContext(ctx)
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})
}

func TestUserContext(t *testing.T) {
	t.Run("successful injection and retrieval of user_id", func(t *testing.T) {
		expectedID := uuid.New()
		ctx := tenant.WithUserID(context.Background(), expectedID)

		retrievedID, err := tenant.UserIDFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedID, retrievedID)
	})

	t.Run("missing user_id returns ErrUserIDNotFound", func(t *testing.T) {
		retrievedID, err := tenant.UserIDFromContext(context.Background())
		assert.ErrorIs(t, err, tenant.ErrUserIDNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})

	t.Run("nil context returns ErrUserIDNotFound", func(t *testing.T) {
		var nilCtx context.Context
		retrievedID, err := tenant.UserIDFromContext(nilCtx)
		assert.ErrorIs(t, err, tenant.ErrUserIDNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})

	t.Run("nil UUID returns ErrUserIDNotFound", func(t *testing.T) {
		ctx := tenant.WithUserID(context.Background(), uuid.Nil)
		retrievedID, err := tenant.UserIDFromContext(ctx)
		assert.ErrorIs(t, err, tenant.ErrUserIDNotFound)
		assert.Equal(t, uuid.Nil, retrievedID)
	})
}

func TestRoleContext(t *testing.T) {
	t.Run("successful injection and retrieval of role", func(t *testing.T) {
		expectedRole := "guard"
		ctx := tenant.WithRole(context.Background(), expectedRole)

		retrievedRole, err := tenant.RoleFromContext(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedRole, retrievedRole)
	})

	t.Run("missing role returns ErrRoleNotFound", func(t *testing.T) {
		retrievedRole, err := tenant.RoleFromContext(context.Background())
		assert.ErrorIs(t, err, tenant.ErrRoleNotFound)
		assert.Empty(t, retrievedRole)
	})

	t.Run("nil context returns ErrRoleNotFound", func(t *testing.T) {
		var nilCtx context.Context
		retrievedRole, err := tenant.RoleFromContext(nilCtx)
		assert.ErrorIs(t, err, tenant.ErrRoleNotFound)
		assert.Empty(t, retrievedRole)
	})

	t.Run("empty string returns ErrRoleNotFound", func(t *testing.T) {
		ctx := tenant.WithRole(context.Background(), "")
		retrievedRole, err := tenant.RoleFromContext(ctx)
		assert.ErrorIs(t, err, tenant.ErrRoleNotFound)
		assert.Empty(t, retrievedRole)
	})
}
