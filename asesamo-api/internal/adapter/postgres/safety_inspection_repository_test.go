package postgres_test

import (
	"os"
	"strings"
	"testing"

	"asesamo-api/internal/adapter/postgres"
	"asesamo-api/internal/core/domain/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSafetyInspectionRepository_StaticQueryIntegrity verifies query patterns and explicit enum casts.
func TestSafetyInspectionRepository_StaticQueryIntegrity(t *testing.T) {
	content, err := os.ReadFile("safety_inspection_repository.go")
	require.NoError(t, err, "Must be able to read repository source code")
	src := string(content)

	t.Run("CreateInspection must cast inspection_type and status enums explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::inspection_type"), "Must cast to ::inspection_type")
		assert.True(t, strings.Contains(src, "::inspection_status"), "Must cast to ::inspection_status")
	})

	t.Run("SaveInspectionAnswers must cast item_compliance_status enum explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::item_compliance_status"), "Must cast to ::item_compliance_status")
	})

	t.Run("CompleteInspection must cast finding_severity and capa_status enums explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::finding_severity"), "Must cast to ::finding_severity")
		assert.True(t, strings.Contains(src, "::capa_status"), "Must cast to ::capa_status")
	})

	t.Run("Multi-tenant isolation: all queries must filter by tenant_id", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "tenant_id = $"), "Must include tenant_id filter")
	})
}

func TestSafetyInspectionRepository_InterfaceCompliance(t *testing.T) {
	// Verificar que postgres.NewSafetyInspectionRepository cumple con repository.SafetyInspectionRepository
	var _ repository.SafetyInspectionRepository = (*postgres.SafetyInspectionRepository)(nil)
}
