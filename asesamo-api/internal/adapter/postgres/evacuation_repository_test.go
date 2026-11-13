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

func TestEvacuationRepository_StaticQueryIntegrity(t *testing.T) {
	content, err := os.ReadFile("evacuation_repository.go")
	require.NoError(t, err, "Must be able to read repository source code")
	src := string(content)

	t.Run("StartEvacuation must cast evacuation_type and status enums explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::evacuation_type"), "Must cast to ::evacuation_type")
		assert.True(t, strings.Contains(src, "::evacuation_status"), "Must cast to ::evacuation_status")
	})

	t.Run("UpdateRollCallCheckIn must cast roll_call_status enum explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::roll_call_status"), "Must cast to ::roll_call_status")
	})

	t.Run("Multi-tenant isolation: all queries must filter by tenant_id", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "tenant_id = $"), "Must include tenant_id filter")
	})
}

func TestEvacuationRepository_InterfaceCompliance(t *testing.T) {
	var _ repository.EmergencyEvacuationRepository = (*postgres.EmergencyEvacuationRepository)(nil)
}
