package entity_test

import (
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafetyInspection_CalculateScoreAndCAPA(t *testing.T) {
	tenantID := uuid.New()
	inspectorID := uuid.New()
	templateID := uuid.New()

	t.Run("calculates 100% when all non-NA items are compliant", func(t *testing.T) {
		insp := entity.NewSafetyInspection(tenantID, templateID, "INSP-2026-0001", entity.InspectionTypeRoutine, "Almacén Central", inspectorID)

		item1 := entity.InspectionItemAnswer{
			TemplateItemID: uuid.New(),
			Status:         entity.ItemComplianceCompliant,
			Weight:         2,
			CriticalRisk:   false,
		}
		item2 := entity.InspectionItemAnswer{
			TemplateItemID: uuid.New(),
			Status:         entity.ItemComplianceNotApplicable,
			Weight:         3,
			CriticalRisk:   false,
		}

		err := insp.AddAnswer(item1)
		require.NoError(t, err)
		err = insp.AddAnswer(item2)
		require.NoError(t, err)

		score, requiresCAPA := insp.CalculateScoreAndEvaluate()
		assert.Equal(t, 100.0, score)
		assert.False(t, requiresCAPA)
		assert.Equal(t, entity.InspectionStatusCompleted, insp.Status)
		assert.Equal(t, 1, insp.CompliantItems)
		assert.Equal(t, 0, insp.NonCompliantItems)
		assert.Equal(t, 1, insp.NAItems)
	})

	t.Run("calculates weighted score and flags requires_capa when critical item fails", func(t *testing.T) {
		insp := entity.NewSafetyInspection(tenantID, templateID, "INSP-2026-0002", entity.InspectionTypeRegulatory, "Área de Pintura", inspectorID)

		// Item 1: Conforme (Peso 2) -> 2 puntos obtenidos de 2
		item1 := entity.InspectionItemAnswer{
			TemplateItemID: uuid.New(),
			Status:         entity.ItemComplianceCompliant,
			Weight:         2,
			CriticalRisk:   false,
		}
		// Item 2: No Conforme (Peso 3, Riesgo Crítico e.g. Extintor bloqueado o falta EPP) -> 0 puntos de 3
		item2 := entity.InspectionItemAnswer{
			TemplateItemID:   uuid.New(),
			Status:           entity.ItemComplianceNonCompliant,
			Weight:           3,
			CriticalRisk:     true,
			FindingNotes:     "Extintor PQS obstruido por tarimas de madera",
			EvidencePhotoURL: "https://storage.asesamo.com/evidence/ext1.jpg",
		}

		require.NoError(t, insp.AddAnswer(item1))
		require.NoError(t, insp.AddAnswer(item2))

		// Puntos obtenidos: 2. Puntos aplicables: 2 + 3 = 5. Score: (2/5) * 100 = 40.0%
		score, requiresCAPA := insp.CalculateScoreAndEvaluate()
		assert.Equal(t, 40.0, score)
		assert.True(t, requiresCAPA)
		assert.Equal(t, entity.InspectionStatusRequiresCAPA, insp.Status)

		// Debe generar plan de acción preliminar para los ítems no conformes
		capas := insp.GenerateRequiredCAPAs(time.Now().AddDate(0, 0, 3)) // Vencimiento en 3 días
		require.Len(t, capas, 1)
		assert.Equal(t, entity.FindingSeverityCritical, capas[0].Severity)
		assert.Equal(t, "Extintor PQS obstruido por tarimas de madera", capas[0].Description)
		assert.Equal(t, entity.CAPAStatusOpen, capas[0].Status)
	})
}
