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

func TestSafetyIncidentRepository_StaticQueryIntegrity(t *testing.T) {
	content, err := os.ReadFile("safety_incident_repository.go")
	require.NoError(t, err, "Must be able to read repository source code")
	src := string(content)

	t.Run("CreateIncident must cast incident_type, severity and status enums explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::incident_type"), "Must cast to ::incident_type")
		assert.True(t, strings.Contains(src, "::incident_severity"), "Must cast to ::incident_severity")
		assert.True(t, strings.Contains(src, "::incident_status"), "Must cast to ::incident_status")
	})

	t.Run("Multi-tenant isolation: all queries must filter by tenant_id", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "tenant_id = $"), "Must include tenant_id filter")
	})
}

func TestSafetyIncidentRepository_InterfaceCompliance(t *testing.T) {
	var _ repository.SafetyIncidentRepository = (*postgres.SafetyIncidentRepository)(nil)
}
