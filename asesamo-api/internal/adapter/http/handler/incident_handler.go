package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/incident"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

type IncidentHandler struct {
	repo     repository.SafetyIncidentRepository
	reportUC *usecase.ReportIncidentUseCase
}

func NewIncidentHandler(repo repository.SafetyIncidentRepository, reportUC *usecase.ReportIncidentUseCase) *IncidentHandler {
	return &IncidentHandler{
		repo:     repo,
		reportUC: reportUC,
	}
}

type ReportIncidentRequestBody struct {
	Type             string    `json:"type"`
	Severity         string    `json:"severity"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Location         string    `json:"location"`
	EventTimestamp   time.Time `json:"event_timestamp"`
	InvolvedWorkerID string    `json:"involved_worker_id,omitempty"`
	ContractorID     string    `json:"contractor_id,omitempty"`
	EvidencePhotoURL string    `json:"evidence_photo_url,omitempty"`
}

func (h *IncidentHandler) ReportIncident(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}
	userID, _ := tenant.UserIDFromContext(r.Context())

	var req ReportIncidentRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	if req.EventTimestamp.IsZero() {
		req.EventTimestamp = time.Now()
	}

	input := usecase.ReportIncidentInput{
		Type:             entity.IncidentType(req.Type),
		Severity:         entity.IncidentSeverity(req.Severity),
		Title:            req.Title,
		Description:      req.Description,
		Location:         req.Location,
		EventTimestamp:   req.EventTimestamp,
		EvidencePhotoURL: req.EvidencePhotoURL,
	}

	if req.InvolvedWorkerID != "" {
		if wid, err := uuid.Parse(req.InvolvedWorkerID); err == nil {
			input.InvolvedWorkerID = &wid
		}
	}
	if req.ContractorID != "" {
		if cid, err := uuid.Parse(req.ContractorID); err == nil {
			input.ContractorID = &cid
		}
	}

	inc, err := h.reportUC.Execute(r.Context(), tenantID, userID, input)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": inc,
	})
}

func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	list, total, err := h.repo.ListIncidents(r.Context(), tenantID, repository.IncidentFilter{})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  list,
		"total": total,
	})
}

func (h *IncidentHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	metrics, err := h.repo.GetMetrics(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": metrics,
	})
}
