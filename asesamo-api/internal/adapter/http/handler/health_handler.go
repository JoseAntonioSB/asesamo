// Package handler provides HTTP request handlers for ASESAMO endpoints.
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/infra/database"
)

// HealthResponse represents the health status of the service and its dependencies.
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// HealthHandler manages liveness and readiness probe requests.
type HealthHandler struct {
	pinger  database.Pinger
	version string
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(pinger database.Pinger, version string) *HealthHandler {
	if version == "" {
		version = "1.0.0"
	}
	return &HealthHandler{
		pinger:  pinger,
		version: version,
	}
}

// Check handles GET /health and GET /api/v1/health.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbStatus := "connected"
	httpStatus := http.StatusOK
	appStatus := "ok"

	if h.pinger != nil {
		if err := h.pinger.Ping(ctx); err != nil {
			dbStatus = "disconnected"
			appStatus = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}
	}

	resp := HealthResponse{
		Status:    appStatus,
		Database:  dbStatus,
		Version:   h.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(resp)
}
