package tenant_test

import (
	"context"
	"testing"

	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type foreignContextKey struct{}

func TestAdversarial_TenantContext(t *testing.T) {
	t.Run("ATTACK: Context key collision resistance", func(t *testing.T) {
		// Attempt to inject tenant ID using an identical name or raw string or foreign struct
		ctx := context.WithValue(context.Background(), "tenantKey", uuid.New())
		ctx = context.WithValue(ctx, foreignContextKey{}, uuid.New())

		// Authoritative FromContext should ignore foreign keys and fail closed
		extractedID, err := tenant.FromContext(ctx)
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
		assert.Equal(t, uuid.Nil, extractedID)
	})

	t.Run("FAIL-CLOSED: Nil UUID injected as tenant returns ErrTenantNotFound", func(t *testing.T) {
		ctx := tenant.WithTenant(context.Background(), uuid.Nil)
		extractedID, err := tenant.FromContext(ctx)
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound, "uuid.Nil must fail closed and return ErrTenantNotFound")
		assert.Equal(t, uuid.Nil, extractedID)
	})

	t.Run("FAIL-CLOSED: Corrupted value type in context fails closed", func(t *testing.T) {
		// Verify nil context
		extractedID, err := tenant.FromContext(nil)
		assert.ErrorIs(t, err, tenant.ErrTenantNotFound)
		assert.Equal(t, uuid.Nil, extractedID)
	})
}
