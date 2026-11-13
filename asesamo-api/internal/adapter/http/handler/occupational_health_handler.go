package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/health"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OccupationalHealthHandler struct {
	repo       repository.HealthRepository
	recordExam *usecase.RecordMedicalExamUseCase
}

func NewOccupationalHealthHandler(repo repository.HealthRepository, recordExam *usecase.RecordMedicalExamUseCase) *OccupationalHealthHandler {
	return &OccupationalHealthHandler{
		repo:       repo,
		recordExam: recordExam,
	}
}

func (h *OccupationalHealthHandler) GetWorkerMedicalProfile(w http.ResponseWriter, r *http.Request) {
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

	profile, err := h.repo.GetMedicalProfileByWorker(r.Context(), tenantID, workerUUID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": profile})
}

type UpsertProfileRequestBody struct {
	BloodType                    string `json:"blood_type"`
	Allergies                    string `json:"allergies,omitempty"`
	ChronicConditions            string `json:"chronic_conditions,omitempty"`
	EmergencyContactName          string `json:"emergency_contact_name"`
	EmergencyContactPhone         string `json:"emergency_contact_phone"`
	EmergencyContactRelationship  string `json:"emergency_contact_relationship"`
}

func (h *OccupationalHealthHandler) UpsertWorkerMedicalProfile(w http.ResponseWriter, r *http.Request) {
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

	var req UpsertProfileRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload inválido"}`, http.StatusBadRequest)
		return
	}

	profile, err := entity.NewWorkerMedicalProfile(
		tenantID,
		workerUUID,
		req.BloodType,
		req.Allergies,
		req.ChronicConditions,
		req.EmergencyContactName,
		req.EmergencyContactPhone,
		req.EmergencyContactRelationship,
	)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpsertMedicalProfile(r.Context(), profile); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": profile})
}

type RecordExamRequestBody struct {
	WorkerID            string    `json:"worker_id"`
	ContractorID        string    `json:"contractor_id,omitempty"`
	ExamType            string    `json:"exam_type"`
	FitnessStatus       string    `json:"fitness_status"`
	ExamDate            time.Time `json:"exam_date"`
	DoctorName          string    `json:"doctor_name"`
	DoctorLicense       string    `json:"doctor_license"`
	ClinicOrInstitution string    `json:"clinic_or_institution,omitempty"`
	RestrictionsNotes   string    `json:"restrictions_notes,omitempty"`
	Recommendations     string    `json:"recommendations,omitempty"`
	VisualAcuityResult  string    `json:"visual_acuity_result,omitempty"`
	AudiometryResult    string    `json:"audiometry_result,omitempty"`
	SpirometryResult    string    `json:"spirometry_result,omitempty"`
	AntidopingResult    string    `json:"antidoping_result,omitempty"`
	BloodPressure       string    `json:"blood_pressure,omitempty"`
}

func (h *OccupationalHealthHandler) RecordMedicalExam(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	var req RecordExamRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload inválido"}`, http.StatusBadRequest)
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

	out, err := h.recordExam.Execute(r.Context(), usecase.RecordMedicalExamInput{
		TenantID:            tenantID,
		WorkerID:            workerUUID,
		ContractorID:        contractorUUID,
		ExamType:            entity.MedicalExamType(req.ExamType),
		FitnessStatus:       entity.MedicalFitnessStatus(req.FitnessStatus),
		ExamDate:            req.ExamDate,
		DoctorName:          req.DoctorName,
		DoctorLicense:       req.DoctorLicense,
		ClinicOrInstitution: req.ClinicOrInstitution,
		RestrictionsNotes:   req.RestrictionsNotes,
		Recommendations:     req.Recommendations,
		VisualAcuityResult:  req.VisualAcuityResult,
		AudiometryResult:    req.AudiometryResult,
		SpirometryResult:    req.SpirometryResult,
		AntidopingResult:    req.AntidopingResult,
		BloodPressure:       req.BloodPressure,
	})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"data": out.Exam})
}

func (h *OccupationalHealthHandler) ListAllExams(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	exams, err := h.repo.ListAllExams(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": exams})
}

func (h *OccupationalHealthHandler) ListWorkerExams(w http.ResponseWriter, r *http.Request) {
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

	exams, err := h.repo.ListExamsByWorker(r.Context(), tenantID, workerUUID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": exams})
}
