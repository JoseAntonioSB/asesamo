package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/training"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TrainingHandler struct {
	repo       repository.TrainingRepository
	issueDC3UC *usecase.IssueDC3CertificateUseCase
}

func NewTrainingHandler(repo repository.TrainingRepository, issueDC3UC *usecase.IssueDC3CertificateUseCase) *TrainingHandler {
	return &TrainingHandler{
		repo:       repo,
		issueDC3UC: issueDC3UC,
	}
}

func (h *TrainingHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	courses, err := h.repo.ListCourses(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": courses})
}

type CreateCourseRequestBody struct {
	Code               string `json:"code"`
	Title              string `json:"title"`
	Description        string `json:"description,omitempty"`
	ThematicArea       string `json:"thematic_area"`
	STPSThematicCode   string `json:"stps_thematic_code"`
	DurationHours      int    `json:"duration_hours"`
	ValidityYears      int    `json:"validity_years"`
	NormativeReference string `json:"normative_reference,omitempty"`
}

func (h *TrainingHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	var req CreateCourseRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload inválido"}`, http.StatusBadRequest)
		return
	}

	course, err := entity.NewTrainingCourse(
		tenantID,
		req.Code,
		req.Title,
		entity.CourseThematicArea(req.ThematicArea),
		req.STPSThematicCode,
		req.DurationHours,
		req.ValidityYears,
		req.NormativeReference,
	)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	course.Description = req.Description

	if err := h.repo.CreateCourse(r.Context(), course); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": course})
}

type IssueDC3RequestBody struct {
	CourseID       string    `json:"course_id"`
	WorkerID       string    `json:"worker_id"`
	ContractorID   string    `json:"contractor_id,omitempty"`
	WorkerCURP     string    `json:"worker_curp"`
	WorkerFullName string    `json:"worker_full_name"`
	OccupationCode string    `json:"occupation_code,omitempty"`
	CompanyName    string    `json:"company_name"`
	CompanyRFC     string    `json:"company_rfc"`
	LegalRepName   string    `json:"legal_rep_name"`
	WorkerRepName  string    `json:"worker_rep_name"`
	InstructorName string    `json:"instructor_name"`
	IssueDate      time.Time `json:"issue_date"`
	Score          float64   `json:"score"`
}

func (h *TrainingHandler) IssueDC3(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	var req IssueDC3RequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload inválido"}`, http.StatusBadRequest)
		return
	}

	courseUUID, err := uuid.Parse(req.CourseID)
	if err != nil {
		http.Error(w, `{"error":"CourseID inválido"}`, http.StatusBadRequest)
		return
	}
	workerUUID, err := uuid.Parse(req.WorkerID)
	if err != nil {
		http.Error(w, `{"error":"WorkerID inválido"}`, http.StatusBadRequest)
		return
	}

	var contractorUUID *uuid.UUID
	if req.ContractorID != "" {
		parsed, err := uuid.Parse(req.ContractorID)
		if err == nil {
			contractorUUID = &parsed
		}
	}

	out, err := h.issueDC3UC.Execute(r.Context(), usecase.IssueDC3Input{
		TenantID:       tenantID,
		CourseID:       courseUUID,
		WorkerID:       workerUUID,
		ContractorID:   contractorUUID,
		WorkerCURP:     req.WorkerCURP,
		WorkerFullName: req.WorkerFullName,
		OccupationCode: req.OccupationCode,
		CompanyName:    req.CompanyName,
		CompanyRFC:     req.CompanyRFC,
		LegalRepName:   req.LegalRepName,
		WorkerRepName:  req.WorkerRepName,
		InstructorName: req.InstructorName,
		IssueDate:      req.IssueDate,
		Score:          req.Score,
	})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": out.Certificate})
}

func (h *TrainingHandler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	certs, err := h.repo.ListCertificates(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": certs})
}

func (h *TrainingHandler) GetCertificate(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	certIDStr := chi.URLParam(r, "id")
	certUUID, err := uuid.Parse(certIDStr)
	if err != nil {
		http.Error(w, `{"error":"ID de constancia inválido"}`, http.StatusBadRequest)
		return
	}

	cert, err := h.repo.GetCertificateByID(r.Context(), tenantID, certUUID)
	if err != nil || cert == nil {
		http.Error(w, `{"error":"Constancia no encontrada"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": cert})
}

func (h *TrainingHandler) ListWorkerCertificates(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	workerIDStr := chi.URLParam(r, "workerId")
	workerUUID, err := uuid.Parse(workerIDStr)
	if err != nil {
		http.Error(w, `{"error":"WorkerID inválido"}`, http.StatusBadRequest)
		return
	}

	certs, err := h.repo.ListCertificatesByWorker(r.Context(), tenantID, workerUUID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": certs})
}
