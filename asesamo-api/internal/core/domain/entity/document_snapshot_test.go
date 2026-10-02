package entity_test

import (
	"testing"

	"asesamo-api/internal/core/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentSnapshot_ValueAndScan(t *testing.T) {
	expDate := "2027-01-15"
	items := entity.DocumentSnapshot{
		{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia de Capacitación DC-3",
			AppliesTo:        "worker",
			Status:           "approved",
			IsRequired:       true,
			IsValid:          true,
			ExpiresAt:        &expDate,
			FailureReason:    nil,
		},
	}

	t.Run("Value serialization", func(t *testing.T) {
		val, err := items.Value()
		require.NoError(t, err)
		strVal, ok := val.(string)
		require.True(t, ok)
		assert.Contains(t, strVal, "DC3")
		assert.Contains(t, strVal, "Constancia de Capacitación DC-3")
	})

	t.Run("Scan from byte slice", func(t *testing.T) {
		val, err := items.Value()
		require.NoError(t, err)
		raw := []byte(val.(string))

		var target entity.DocumentSnapshot
		err = target.Scan(raw)
		require.NoError(t, err)
		assert.Len(t, target, 1)
		assert.Equal(t, "DC3", target[0].DocumentTypeCode)
		assert.Equal(t, true, target[0].IsValid)
	})

	t.Run("Scan from string", func(t *testing.T) {
		val, err := items.Value()
		require.NoError(t, err)
		raw := val.(string)

		var target entity.DocumentSnapshot
		err = target.Scan(raw)
		require.NoError(t, err)
		assert.Len(t, target, 1)
		assert.Equal(t, "DC3", target[0].DocumentTypeCode)
	})

	t.Run("Nil Value produces empty JSON array string", func(t *testing.T) {
		var empty entity.DocumentSnapshot
		val, err := empty.Value()
		require.NoError(t, err)
		assert.Equal(t, "[]", val)
	})

	t.Run("Scan from nil", func(t *testing.T) {
		var target entity.DocumentSnapshot
		err := target.Scan(nil)
		require.NoError(t, err)
		assert.Empty(t, target)
	})

	t.Run("Scan unsupported type produces error", func(t *testing.T) {
		var target entity.DocumentSnapshot
		err := target.Scan(12345)
		assert.Error(t, err)
	})
}
