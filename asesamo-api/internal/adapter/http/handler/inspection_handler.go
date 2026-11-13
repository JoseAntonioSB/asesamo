package handler

import (
	"encoding/json"
	"net/http"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/inspection"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CompleteInspectionRequestBody struct {
	Answers []struct {
		TemplateItemID   string `json:"template_item_id"`
		Status           string `json:"status"`
		FindingNotes     string `json:"finding_notes,omitempty"`
		EvidencePhotoURL string `json:"evidence_photo_url,omitempty"`
	} `json:"answers"`
}

type InspectionHandler struct {
	repo       repository.SafetyInspectionRepository
	completeUC *usecase.CompleteInspectionUseCase
}

func NewInspectionHandler(repo repository.SafetyInspectionRepository, completeUC *usecase.CompleteInspectionUseCase) *InspectionHandler {
	return &InspectionHandler{
		repo:       repo,
		completeUC: completeUC,
	}
}

func (h *InspectionHandler) CompleteInspection(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	inspectionID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"ID de inspección inválido"}`, http.StatusBadRequest)
		return
	}

	var req CompleteInspectionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	answersInput := make([]usecase.SubmitAnswerInput, 0, len(req.Answers))
	for _, a := range req.Answers {
		tmplItemID, err := uuid.Parse(a.TemplateItemID)
		if err != nil {
			http.Error(w, `{"error":"ID de template_item_id inválido"}`, http.StatusBadRequest)
			return
		}
		answersInput = append(answersInput, usecase.SubmitAnswerInput{
			TemplateItemID:   tmplItemID,
			Status:           entity.ItemComplianceStatus(a.Status),
			FindingNotes:     a.FindingNotes,
			EvidencePhotoURL: a.EvidencePhotoURL,
		})
	}

	res, err := h.completeUC.Execute(r.Context(), tenantID, inspectionID, answersInput)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": res,
	})
}
