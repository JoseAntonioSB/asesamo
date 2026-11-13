package handler

import (
	"encoding/json"
	"net/http"
	"time"

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

func (h *InspectionHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	templates, err := h.repo.ListTemplates(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": templates,
	})
}

func (h *InspectionHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	inspections, total, err := h.repo.ListInspections(r.Context(), tenantID, repository.InspectionFilter{})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  inspections,
		"total": total,
	})
}

type CreateInspectionRequestBody struct {
	TemplateID     string `json:"template_id"`
	InspectionType string `json:"inspection_type"`
	AreaOrLocation string `json:"area_or_location"`
	ContractorID   string `json:"contractor_id,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

func (h *InspectionHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}
	userID, _ := tenant.UserIDFromContext(r.Context())

	var req CreateInspectionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	tmplID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		http.Error(w, `{"error":"ID de plantilla inválido"}`, http.StatusBadRequest)
		return
	}

	inspType := entity.InspectionType(req.InspectionType)
	if inspType == "" {
		inspType = entity.InspectionTypeRoutine
	}

	inspNumber := "INSP-" + time.Now().Format("2006") + "-" + uuid.New().String()[:6]
	insp := entity.NewSafetyInspection(tenantID, tmplID, inspNumber, inspType, req.AreaOrLocation, userID)
	insp.Notes = req.Notes

	if req.ContractorID != "" {
		if cid, err := uuid.Parse(req.ContractorID); err == nil {
			insp.ContractorID = &cid
		}
	}

	if err := h.repo.CreateInspection(r.Context(), insp); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": insp,
	})
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
