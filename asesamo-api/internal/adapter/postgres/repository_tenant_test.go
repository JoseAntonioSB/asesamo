package postgres_test

import (
	"context"
	"testing"

	"asesamo-api/internal/adapter/postgres"
	"asesamo-api/internal/core/domain/entity"
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
	permitRepo := postgres.NewWorkPermitRepository(nil)

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

		t.Run("WorkPermitRepository.GetByID fails closed on "+tc.name, func(t *testing.T) {
			wp, err := permitRepo.GetByID(tc.ctx, testID)
			assert.Nil(t, wp)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.ListWithFilters fails closed on "+tc.name, func(t *testing.T) {
			wps, total, err := permitRepo.ListWithFilters(tc.ctx, repository.PermitFilter{})
			assert.Nil(t, wps)
			assert.Equal(t, 0, total)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.Create fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.Create(tc.ctx, &entity.WorkPermit{ID: testID})
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.NextPermitNumber fails closed on "+tc.name, func(t *testing.T) {
			tenantID, _ := tenant.FromContext(tc.ctx)
			_, err := permitRepo.NextPermitNumber(tc.ctx, tenantID, 2026)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.UpdateStatus fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.UpdateStatus(tc.ctx, testID, entity.PermitStatusClosed, nil, nil)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.AssignWorker fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.AssignWorker(tc.ctx, &entity.PermitWorker{PermitID: testID, WorkerID: uuid.New()})
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.RemoveWorker fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.RemoveWorker(tc.ctx, testID, uuid.New())
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.ListWorkersByPermit fails closed on "+tc.name, func(t *testing.T) {
			workers, err := permitRepo.ListWorkersByPermit(tc.ctx, testID)
			assert.Nil(t, workers)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.InitChecklistItems fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.InitChecklistItems(tc.ctx, testID, []entity.PermitChecklistItem{{ItemKey: "chk_test"}})
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.ListChecklistByPermit fails closed on "+tc.name, func(t *testing.T) {
			items, err := permitRepo.ListChecklistByPermit(tc.ctx, testID)
			assert.Nil(t, items)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.UpdateChecklistItem fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.UpdateChecklistItem(tc.ctx, testID, "chk_test", true, nil, nil, nil)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.AddSignature fails closed on "+tc.name, func(t *testing.T) {
			err := permitRepo.AddSignature(tc.ctx, &entity.PermitSignature{PermitID: testID, UserID: uuid.New(), Role: entity.PermitSignerRequester})
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})

		t.Run("WorkPermitRepository.ListSignaturesByPermit fails closed on "+tc.name, func(t *testing.T) {
			sigs, err := permitRepo.ListSignaturesByPermit(tc.ctx, testID)
			assert.Nil(t, sigs)
			assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		})
	}
}
