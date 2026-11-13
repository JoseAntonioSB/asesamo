package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookSender defines the contract for sending transactional webhooks.
type WebhookSender interface {
	SendWebhook(ctx context.Context, webhookURL string, secret string, event string, payload interface{}) error
}

// HTTPWebhookSender implements WebhookSender over HTTP POST with HMAC-SHA256 signatures.
type HTTPWebhookSender struct {
	client *http.Client
}

// NewHTTPWebhookSender constructs an HTTPWebhookSender with a configurable request timeout.
func NewHTTPWebhookSender(timeout time.Duration) *HTTPWebhookSender {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPWebhookSender{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// NewHTTPWebhookSenderWithClient constructs an HTTPWebhookSender with a custom http.Client.
func NewHTTPWebhookSenderWithClient(client *http.Client) *HTTPWebhookSender {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPWebhookSender{
		client: client,
	}
}

// SendWebhook dispatches an HTTP POST request to the webhookURL with HMAC-SHA256 signature in X-Asesamo-Signature.
func (s *HTTPWebhookSender) SendWebhook(ctx context.Context, webhookURL string, secret string, event string, payload interface{}) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializing webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating webhook request: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ASESAMO-Industrial-Webhook/1.0")
	req.Header.Set("X-Asesamo-Event", event)
	req.Header.Set("X-Asesamo-Timestamp", timestamp)

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(bodyBytes)
	sigHex := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Asesamo-Signature", sigHex)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}
