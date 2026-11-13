package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/evacuation"
	"asesamo-api/internal/infra/realtime"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EvacuationHandler struct {
	repo    repository.EmergencyEvacuationRepository
	startUC *usecase.StartEvacuationUseCase
	broker  realtime.Broker
}

func NewEvacuationHandler(repo repository.EmergencyEvacuationRepository, startUC *usecase.StartEvacuationUseCase, broker realtime.Broker) *EvacuationHandler {
	return &EvacuationHandler{
		repo:    repo,
		startUC: startUC,
		broker:  broker,
	}
}

func (h *EvacuationHandler) ListMusterPoints(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	points, err := h.repo.ListMusterPoints(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": points,
	})
}

type StartEvacuationRequestBody struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func (h *EvacuationHandler) StartEvacuation(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}
	userID, _ := tenant.UserIDFromContext(r.Context())

	var req StartEvacuationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	evacType := entity.EvacuationType(req.Type)
	if evacType == "" {
		evacType = entity.EvacuationTypeDrill
	}

	evac, err := h.startUC.Execute(r.Context(), tenantID, userID, usecase.StartEvacuationInput{
		Type:        evacType,
		Description: req.Description,
	})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if h.broker != nil {
		h.broker.Publish(realtime.Event{
			Type:      "evacuation.started",
			TenantID:  tenantID.String(),
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"evacuation_id":     evac.ID.String(),
				"evacuation_number": evac.EvacuationNumber,
				"evacuation_type":   string(evac.EvacuationType),
				"description":       evac.Description,
				"total_headcount":   evac.TotalHeadcount,
				"safe_count":        evac.SafeCount,
				"missing_count":     evac.MissingCount,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": evac,
	})
}

func (h *EvacuationHandler) GetActiveEvacuation(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	evac, err := h.repo.GetActiveEvacuation(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": evac,
	})
}

type RollCallCheckInRequestBody struct {
	WorkerID      string `json:"worker_id"`
	MusterPointID string `json:"muster_point_id"`
	Notes         string `json:"notes,omitempty"`
}

func (h *EvacuationHandler) CheckInWorker(w http.ResponseWriter, r *http.Request) {
	evacIDStr := chi.URLParam(r, "id")
	evacID, err := uuid.Parse(evacIDStr)
	if err != nil {
		http.Error(w, `{"error":"ID de evacuación inválido"}`, http.StatusBadRequest)
		return
	}
	userID, _ := tenant.UserIDFromContext(r.Context())

	var req RollCallCheckInRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	workerID, err := uuid.Parse(req.WorkerID)
	if err != nil {
		http.Error(w, `{"error":"ID de trabajador inválido"}`, http.StatusBadRequest)
		return
	}
	musterID, err := uuid.Parse(req.MusterPointID)
	if err != nil {
		http.Error(w, `{"error":"ID de punto de reunión inválido"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateRollCallCheckIn(r.Context(), evacID, workerID, musterID, userID, req.Notes); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if h.broker != nil {
		tenantID, _ := tenant.FromContext(r.Context())
		activeEvac, _ := h.repo.GetActiveEvacuation(r.Context(), tenantID)
		payload := map[string]interface{}{
			"evacuation_id":   evacID.String(),
			"worker_id":       workerID.String(),
			"muster_point_id": musterID.String(),
		}
		if activeEvac != nil {
			payload["total_headcount"] = activeEvac.TotalHeadcount
			payload["safe_count"] = activeEvac.SafeCount
			payload["missing_count"] = activeEvac.MissingCount
		}
		h.broker.Publish(realtime.Event{
			Type:      "evacuation.roll_call_checkin",
			TenantID:  tenantID.String(),
			Timestamp: time.Now(),
			Payload:   payload,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Trabajador registrado a salvo",
	})
}

func (h *EvacuationHandler) EndEvacuation(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	evacIDStr := chi.URLParam(r, "id")
	evacID, err := uuid.Parse(evacIDStr)
	if err != nil {
		http.Error(w, `{"error":"ID de evacuación inválido"}`, http.StatusBadRequest)
		return
	}

	evac, err := h.repo.GetEvacuationByID(r.Context(), tenantID, evacID)
	if err != nil || evac == nil {
		http.Error(w, `{"error":"Evacuación no encontrada"}`, http.StatusNotFound)
		return
	}

	if err := evac.DeclareAllClear(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.EndEvacuation(r.Context(), evac); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if h.broker != nil {
		h.broker.Publish(realtime.Event{
			Type:      "evacuation.ended",
			TenantID:  tenantID.String(),
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"evacuation_id": evac.ID.String(),
				"status":        "closed",
				"all_clear":     true,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": evac,
	})
}
