package notification_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"testing"

	"asesamo-api/internal/infra/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestWebhookSender_SendSuccess(t *testing.T) {
	secret := "webhook_super_secret_key_12345"
	event := "document.rejected"
	payload := map[string]interface{}{
		"document_id": "doc-123",
		"reason":      "Falta firma de la comisión mixta STPS",
		"worker_name": "Juan Perez",
	}

	var receivedHeaders http.Header
	var receivedBody []byte

	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			receivedHeaders = req.Header.Clone()
			receivedBody, _ = io.ReadAll(req.Body)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"received":true}`)),
				Header:     make(http.Header),
			}
		}),
	}

	sender := notification.NewHTTPWebhookSenderWithClient(mockClient)
	err := sender.SendWebhook(context.Background(), "https://webhook.example.com/events", secret, event, payload)
	require.NoError(t, err)

	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	assert.Equal(t, event, receivedHeaders.Get("X-Asesamo-Event"))
	assert.NotEmpty(t, receivedHeaders.Get("X-Asesamo-Timestamp"))

	// Verify HMAC signature
	sigHex := receivedHeaders.Get("X-Asesamo-Signature")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(receivedBody)
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	assert.Equal(t, expectedSig, sigHex)
}

func TestWebhookSender_TargetError(t *testing.T) {
	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"server error"}`)),
				Header:     make(http.Header),
			}
		}),
	}

	sender := notification.NewHTTPWebhookSenderWithClient(mockClient)
	err := sender.SendWebhook(context.Background(), "https://webhook.example.com/events", "secret", "test.event", map[string]string{})
	assert.Error(t, err)
}
