package notification_test

import (
	"context"
	"net/http"
	"testing"

	"asesamo-api/internal/infra/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationService_DocumentRejection(t *testing.T) {
	mockEmail := notification.NewMockEmailSender()
	var webhookPayloadSent bool

	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			if req.Header.Get("X-Asesamo-Event") == "document.rejected" {
				webhookPayloadSent = true
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}
		}),
	}
	mockWebhook := notification.NewHTTPWebhookSenderWithClient(mockClient)

	svc := notification.NewService(mockEmail, mockWebhook, "https://webhook.example.com", "test-secret")

	err := svc.NotifyDocumentRejection(
		context.Background(),
		"contratista@empresa.com",
		"Mantenimiento Industrial SA",
		"Constancia DC-3",
		"Pedro Gomez",
		"Falta sello del agente capacitador STPS",
	)
	require.NoError(t, err)

	// Verify email
	sentEmails := mockEmail.GetSentEmails()
	require.Len(t, sentEmails, 1)
	assert.Equal(t, "contratista@empresa.com", sentEmails[0].To)
	assert.Contains(t, sentEmails[0].Subject, "Documento Rechazado")
	assert.Contains(t, sentEmails[0].HTMLBody, "Mantenimiento Industrial SA")

	// Verify webhook
	assert.True(t, webhookPayloadSent)
}

func TestNotificationService_CriticalIncident(t *testing.T) {
	mockEmail := notification.NewMockEmailSender()
	var eventHeader string

	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			eventHeader = req.Header.Get("X-Asesamo-Event")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}
		}),
	}
	mockWebhook := notification.NewHTTPWebhookSenderWithClient(mockClient)

	svc := notification.NewService(mockEmail, mockWebhook, "https://webhook.example.com", "test-secret")

	err := svc.NotifyCriticalIncident(
		context.Background(),
		"INC-2026-0099",
		"critical",
		"Nave 2 - Almacén Químico",
		"Fuga de vapores de ácido clorhídrico en área de descarga.",
	)
	require.NoError(t, err)
	assert.Equal(t, "incident.critical", eventHeader)
}
