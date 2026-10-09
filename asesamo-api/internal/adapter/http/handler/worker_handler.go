package handler

import (
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/worker"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WorkerHandler handles worker credential and badge requests.
type WorkerHandler struct {
	getCredentialUseCase *worker.GetCredentialUseCase
}

// NewWorkerHandler instantiates a WorkerHandler.
func NewWorkerHandler(getCredentialUseCase *worker.GetCredentialUseCase) *WorkerHandler {
	return &WorkerHandler{
		getCredentialUseCase: getCredentialUseCase,
	}
}

// GetMyCredential handles GET /api/v1/workers/me/credential.
func (h *WorkerHandler) GetMyCredential(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	h.handleGetCredential(w, r, userID)
}

// GetWorkerCredential handles GET /api/v1/workers/{id}/credential.
func (h *WorkerHandler) GetWorkerCredential(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es requerido."))
		return
	}

	workerID, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador no es un UUID válido."))
		return
	}

	h.handleGetCredential(w, r, workerID)
}

func (h *WorkerHandler) handleGetCredential(w http.ResponseWriter, r *http.Request, workerID uuid.UUID) {
	var plantID *uuid.UUID
	plantIDStr := r.URL.Query().Get("plant_id")
	if plantIDStr != "" {
		parsed, err := uuid.Parse(plantIDStr)
		if err == nil {
			plantID = &parsed
		}
	}

	out, err := h.getCredentialUseCase.Execute(r.Context(), worker.GetCredentialInput{
		WorkerUserID: workerID,
		PlantID:      plantID,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, out)
}
