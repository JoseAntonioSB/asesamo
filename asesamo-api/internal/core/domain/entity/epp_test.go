package entity_test

import (
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEPP_DomainRules(t *testing.T) {
	tenantID := uuid.New()
	workerID := uuid.New()
	delivererID := uuid.New()

	t.Run("calculates correct expiration date and active status upon delivery", func(t *testing.T) {
		item := entity.EPPItem{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Code:         "EPP-CAS-001",
			Name:         "Casco Industrial Tipo 1 Clase E",
			Category:     entity.EPPCategoryHead,
			LifespanDays: 365,
			IsActive:     true,
		}

		deliveryDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		delivery, err := entity.NewEPPDelivery(
			tenantID,
			item.ID,
			workerID,
			nil,
			deliveryDate,
			item.LifespanDays,
			delivererID,
			"firma-digital-base64",
			"Entrega inicial de dotación",
		)
		require.NoError(t, err)

		assert.Equal(t, entity.EPPDeliveryStatusDeliveredActive, delivery.Status)
		expectedExp := deliveryDate.AddDate(0, 0, 365)
		assert.Equal(t, expectedExp, delivery.ExpirationDate)
		assert.False(t, delivery.IsExpiredAt(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))
		assert.True(t, delivery.IsExpiredAt(time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)))
	})

	t.Run("marks as replaced with replacement notes", func(t *testing.T) {
		delivery := &entity.EPPDelivery{
			ID:             uuid.New(),
			Status:         entity.EPPDeliveryStatusDeliveredActive,
			ExpirationDate: time.Now().AddDate(0, 0, 100),
		}

		err := delivery.MarkReplaced("Casco con impacto o fisura en obra")
		require.NoError(t, err)
		assert.Equal(t, entity.EPPDeliveryStatusReplaced, delivery.Status)
		assert.Contains(t, delivery.Notes, "Casco con impacto o fisura")
	})
}
