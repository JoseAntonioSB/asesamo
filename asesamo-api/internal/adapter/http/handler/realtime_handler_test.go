package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/infra/realtime"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

func TestRealtimeHandler_SSEStream(t *testing.T) {
	broker := realtime.NewInMemoryEventBroker(10)
	h := handler.NewRealtimeHandler(broker)

	tenantID := uuid.New()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = tenant.WithTenant(ctx, tenantID)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/realtime/events", nil)
	rec := httptest.NewRecorder()

	go func() {
		// Wait briefly for connection setup
		time.Sleep(50 * time.Millisecond)

		// Broadcast an event
		broker.Publish(realtime.Event{
			Type:      "evacuation.started",
			TenantID:  tenantID.String(),
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"evacuation_number": "EVAC-TEST-001",
			},
		})

		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	h.StreamEvents(rec, req)

	res := rec.Result()
	if res.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected text/event-stream, got %s", res.Header.Get("Content-Type"))
	}
	if res.Header.Get("Cache-Control") != "no-cache" {
		t.Fatalf("expected no-cache, got %s", res.Header.Get("Cache-Control"))
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: connected") {
		t.Fatalf("expected connected event in stream, got %s", body)
	}
	if !strings.Contains(body, "event: evacuation.started") {
		t.Fatalf("expected evacuation.started event in stream, got %s", body)
	}
	if !strings.Contains(body, "EVAC-TEST-001") {
		t.Fatalf("expected EVAC-TEST-001 in stream body, got %s", body)
	}
}
