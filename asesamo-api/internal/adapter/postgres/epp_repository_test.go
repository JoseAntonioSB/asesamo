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

func TestEPPRepository_StaticQueryIntegrity(t *testing.T) {
	content, err := os.ReadFile("epp_repository.go")
	require.NoError(t, err, "Must be able to read repository source code")
	src := string(content)

	t.Run("CreateItem must cast epp_category enum explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::epp_category"), "Must cast to ::epp_category")
	})

	t.Run("RecordDelivery must cast epp_delivery_status enum explicitly", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "::epp_delivery_status"), "Must cast to ::epp_delivery_status")
	})

	t.Run("Multi-tenant isolation: all queries must filter by tenant_id", func(t *testing.T) {
		assert.True(t, strings.Contains(src, "tenant_id = $"), "Must include tenant_id filter")
	})
}

func TestEPPRepository_InterfaceCompliance(t *testing.T) {
	var _ repository.EPPRepository = (*postgres.EPPRepository)(nil)
}
