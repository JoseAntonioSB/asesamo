package postgres_test

import (
	"context"
	"testing"

	"asesamo-api/internal/adapter/postgres"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type foreignTenantKey struct{}

func TestRepositories_TenantContextFailClosed(t *testing.T) {
	apRepo := postgres.NewAccessPointRepository(nil)
	plantRepo := postgres.NewPlantRepository(nil)
	accessLogRepo := postgres.NewAccessLogRepository(nil)

	testID := uuid.New()

	contextsToTest := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "empty background context",
			ctx:  context.Background(),
		},
		{
			name: "nil context",
			ctx:  nil,
		},
		{
			name: "context with uuid.Nil tenant",
			ctx:  tenant.WithTenant(context.Background(), uuid.Nil),
		},
		{
			name: "context with foreign string key",
			ctx:  context.WithValue(context.Background(), "tenantKey", uuid.New()),
		},
		{
			name: "context with foreign struct key",
			ctx:  context.WithValue(context.Background(), foreignTenantKey{}, uuid.New()),
		},
	}

	for _, tc := range contextsToTest {
		t.Run("AccessPointRepository.GetByID fails closed on "+tc.name, func(t *testing.T) {
			ap, err := apRepo.GetByID(tc.ctx, testID)
			assert.Nil(t, ap)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("AccessPointRepository.ListByPlant fails closed on "+tc.name, func(t *testing.T) {
			aps, err := apRepo.ListByPlant(tc.ctx, testID)
			assert.Nil(t, aps)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("PlantRepository.GetByID fails closed on "+tc.name, func(t *testing.T) {
			p, err := plantRepo.GetByID(tc.ctx, testID)
			assert.Nil(t, p)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("AccessLogRepository.GetByID fails closed on "+tc.name, func(t *testing.T) {
			al, err := accessLogRepo.GetByID(tc.ctx, testID)
			assert.Nil(t, al)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("AccessLogRepository.ListWithFilters fails closed on "+tc.name, func(t *testing.T) {
			logs, total, err := accessLogRepo.ListWithFilters(tc.ctx, repository.AccessLogFilter{})
			assert.Nil(t, logs)
			assert.Equal(t, 0, total)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})
	}
}
