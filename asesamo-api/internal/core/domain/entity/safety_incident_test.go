package entity_test

import (
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafetyIncident_LifecycleAndSTPSMetrics(t *testing.T) {
	tenantID := uuid.New()
	reporterID := uuid.New()
	investigatorID := uuid.New()

	t.Run("creates new incident with initial reported status and validates severity classification", func(t *testing.T) {
		inc, err := entity.NewSafetyIncident(
			tenantID,
			"INC-2026-0001",
			entity.IncidentTypeNearMiss,
			entity.IncidentSeverityModerate,
			"Caída de herramienta desde andamio",
			"Se desprendió llave stilson desde nivel 3 sin lesionados",
			"Nave 2 - Estructura Este",
			time.Now(),
			reporterID,
		)
		require.NoError(t, err)
		assert.Equal(t, entity.IncidentStatusReported, inc.Status)
		assert.Equal(t, 0, inc.DaysLost)
		assert.False(t, inc.IsRecordableSTPS())
	})

	t.Run("lost_time_injury requires days_lost > 0 and classifies as recordable STPS", func(t *testing.T) {
		inc, err := entity.NewSafetyIncident(
			tenantID,
			"INC-2026-0002",
			entity.IncidentTypeLostTimeInjury,
			entity.IncidentSeverityMajor,
			"Esguince de tobillo por desnivel",
			"Trabajador tropezó con cable suelto",
			"Pasillo Central",
			time.Now(),
			reporterID,
		)
		require.NoError(t, err)

		// Debe fallar al cerrar si tiene 0 días perdidos siendo LTI
		err = inc.CompleteInvestigation(
			investigatorID,
			"Falta de canalización de cables en zona de paso peatonal",
			"Cable temporal no fijado al piso",
			"Falta de supervisión de contratista",
			0, // 0 días perdidos inválido para LTI
		)
		assert.Error(t, err, "LTI debe requerir al menos 1 día de incapacidad")

		// Completar con 5 días perdidos
		err = inc.CompleteInvestigation(
			investigatorID,
			"Falta de canalización de cables en zona de paso peatonal",
			"Cable temporal no fijado al piso",
			"Falta de supervisión de contratista",
			5,
		)
		require.NoError(t, err)
		assert.Equal(t, entity.IncidentStatusInvestigating, inc.Status)
		assert.Equal(t, 5, inc.DaysLost)
		assert.True(t, inc.IsRecordableSTPS())

		// Cerrar formalmente
		err = inc.CloseIncident()
		require.NoError(t, err)
		assert.Equal(t, entity.IncidentStatusClosed, inc.Status)
		assert.NotNil(t, inc.ClosedAt)
	})
}
