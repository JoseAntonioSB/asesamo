package notification_test

import (
	"context"
	"strings"
	"testing"

	"asesamo-api/internal/infra/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailTemplates_Formatting(t *testing.T) {
	t.Run("Document Rejection Email", func(t *testing.T) {
		subject, body := notification.FormatDocumentRejectionEmail(
			"Mantenimiento Industrial del Norte SA",
			"Constancia DC-3 Alturas",
			"Carlos Mendez",
			"El certificado presenta fecha de emisión expirada y carece del folio STPS.",
		)

		assert.Contains(t, subject, "Documento Rechazado")
		assert.Contains(t, body, "Mantenimiento Industrial del Norte SA")
		assert.Contains(t, body, "Constancia DC-3 Alturas")
		assert.Contains(t, body, "Carlos Mendez")
		assert.Contains(t, body, "El certificado presenta fecha de emisión expirada")
		assert.True(t, strings.HasPrefix(body, "<!DOCTYPE html>"))
	})

	t.Run("Document Expiration Warning Email", func(t *testing.T) {
		subject, body := notification.FormatDocumentExpirationWarningEmail(
			"Logística y Maniobras SA",
			"Opinión de Cumplimiento SAT 32-D",
			7,
		)

		assert.Contains(t, subject, "Vencimiento Próximo")
		assert.Contains(t, subject, "7 días")
		assert.Contains(t, body, "Logística y Maniobras SA")
		assert.Contains(t, body, "Opinión de Cumplimiento SAT 32-D")
	})

	t.Run("Critical Incident Email", func(t *testing.T) {
		subject, body := notification.FormatCriticalIncidentEmail(
			"INC-2026-0042",
			"critical",
			"Nave 3 - Subestación Eléctrica",
			"Conato de arco eléctrico durante maniobra de mantenimiento.",
		)

		assert.Contains(t, subject, "ALERTA EHS")
		assert.Contains(t, subject, "INC-2026-0042")
		assert.Contains(t, body, "Nave 3 - Subestación Eléctrica")
		assert.Contains(t, body, "Conato de arco eléctrico")
	})
}

func TestTransactionalEmailSender_Mock(t *testing.T) {
	sender := notification.NewMockEmailSender()
	err := sender.SendEmail(context.Background(), "contratista@empresa.com", "Prueba", "<h1>Hola</h1>")
	require.NoError(t, err)

	history := sender.GetSentEmails()
	require.Len(t, history, 1)
	assert.Equal(t, "contratista@empresa.com", history[0].To)
	assert.Equal(t, "Prueba", history[0].Subject)
}
