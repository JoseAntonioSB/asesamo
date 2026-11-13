package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"asesamo-api/internal/infra/realtime"
	"asesamo-api/pkg/tenant"
)

// RealtimeHandler manages Server-Sent Events (SSE) connections for industrial real-time monitoring.
type RealtimeHandler struct {
	broker realtime.Broker
}

// NewRealtimeHandler constructs a new RealtimeHandler.
func NewRealtimeHandler(broker realtime.Broker) *RealtimeHandler {
	return &RealtimeHandler{
		broker: broker,
	}
}

// StreamEvents handles GET /api/v1/realtime/events streaming SSE to authenticated clients.
func (h *RealtimeHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"Streaming no soportado por el cliente"}`, http.StatusInternalServerError)
		return
	}

	// Set SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Send initial handshake
	_, _ = fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"tenant_id\":\"%s\"}\n\n", tenantID.String())
	flusher.Flush()

	// Subscribe to broker
	eventCh, unsubscribe := h.broker.Subscribe(tenantID.String())
	defer unsubscribe()

	// Keepalive ticker (every 15s to keep connections alive through reverse proxies)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			_, err := fmt.Fprintf(w, ": keepalive\n\n")
			if err != nil {
				return
			}
			flusher.Flush()

		case event, ok := <-eventCh:
			if !ok {
				return
			}

			payloadBytes, err := json.Marshal(event.Payload)
			if err != nil {
				continue
			}

			_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(payloadBytes))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
