package entity_test

import (
	"testing"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmergencyEvacuation_DomainRules(t *testing.T) {
	tenantID := uuid.New()
	initiatorID := uuid.New()
	worker1 := uuid.New()
	worker2 := uuid.New()
	musterPointID := uuid.New()

	t.Run("creates active evacuation and manages roll call accounting", func(t *testing.T) {
		evac := entity.NewEmergencyEvacuation(
			tenantID,
			"EVAC-2026-0001",
			entity.EvacuationTypeDrill,
			"Simulacro Nacional de Evacuación",
			initiatorID,
		)

		assert.Equal(t, entity.EvacuationStatusActive, evac.Status)

		// Cargar lista de trabajadores actualmente en planta
		err := evac.InitializeRollCall([]uuid.UUID{worker1, worker2})
		require.NoError(t, err)
		assert.Equal(t, 2, evac.TotalHeadcount)
		assert.Equal(t, 2, evac.MissingCount)
		assert.Equal(t, 0, evac.SafeCount)

		// Marcar trabajador 1 como a salvo en Punto de Reunión
		err = evac.CheckInWorker(worker1, musterPointID, initiatorID, "")
		require.NoError(t, err)
		assert.Equal(t, 1, evac.MissingCount)
		assert.Equal(t, 1, evac.SafeCount)

		// Intentar finalizar cuando aún hay faltantes
		canFinishAllClear := evac.CanDeclareAllClear()
		assert.False(t, canFinishAllClear, "No puede declarar All Clear si aún hay personal faltante")

		// Marcar trabajador 2
		err = evac.CheckInWorker(worker2, musterPointID, initiatorID, "")
		require.NoError(t, err)
		assert.Equal(t, 0, evac.MissingCount)
		assert.Equal(t, 2, evac.SafeCount)
		assert.True(t, evac.CanDeclareAllClear())

		// Declarar fin de contingencia
		err = evac.DeclareAllClear()
		require.NoError(t, err)
		assert.Equal(t, entity.EvacuationStatusAllClear, evac.Status)
		assert.NotNil(t, evac.EndedAt)
		assert.GreaterOrEqual(t, evac.EvacuationTimeSeconds, 0)
	})
}
