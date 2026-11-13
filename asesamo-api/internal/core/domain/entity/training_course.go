package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CourseThematicArea string

const (
	CourseAreaSafety         CourseThematicArea = "safety"
	CourseAreaTechnical      CourseThematicArea = "technical"
	CourseAreaEnvironmental  CourseThematicArea = "environmental"
	CourseAreaAdministrative CourseThematicArea = "administrative"
	CourseAreaSpecialized    CourseThematicArea = "specialized"
)

type CertificateStatus string

const (
	CertificateStatusDraft   CertificateStatus = "draft"
	CertificateStatusIssued  CertificateStatus = "issued"
	CertificateStatusRevoked CertificateStatus = "revoked"
	CertificateStatusExpired CertificateStatus = "expired"
)

// TrainingCourse represents a certified training course in the catalog.
type TrainingCourse struct {
	ID                 uuid.UUID          `json:"id"`
	TenantID           uuid.UUID          `json:"tenant_id"`
	Code               string             `json:"code"`
	Title              string             `json:"title"`
	Description        string             `json:"description,omitempty"`
	ThematicArea       CourseThematicArea `json:"thematic_area"`
	STPSThematicCode   string             `json:"stps_thematic_code"`
	DurationHours      int                `json:"duration_hours"`
	ValidityYears      int                `json:"validity_years"`
	NormativeReference string             `json:"normative_reference,omitempty"`
	IsActive           bool               `json:"is_active"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

func NewTrainingCourse(
	tenantID uuid.UUID,
	code string,
	title string,
	thematicArea CourseThematicArea,
	stpsCode string,
	durationHours int,
	validityYears int,
	normativeRef string,
) (*TrainingCourse, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("el título del curso es obligatorio")
	}
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("el código del curso es obligatorio")
	}
	if durationHours <= 0 {
		return nil, errors.New("la duración del curso debe ser mayor a 0 horas")
	}
	if validityYears <= 0 {
		validityYears = 1
	}

	now := time.Now().UTC()
	return &TrainingCourse{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		Code:               strings.TrimSpace(code),
		Title:              strings.TrimSpace(title),
		ThematicArea:       thematicArea,
		STPSThematicCode:   strings.TrimSpace(stpsCode),
		DurationHours:      durationHours,
		ValidityYears:      validityYears,
		NormativeReference: strings.TrimSpace(normativeRef),
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// TrainingCertificate represents an official STPS DC-3 certificate of labor competencies.
type TrainingCertificate struct {
	ID                    uuid.UUID         `json:"id"`
	TenantID              uuid.UUID         `json:"tenant_id"`
	SessionID             *uuid.UUID        `json:"session_id,omitempty"`
	CourseID              uuid.UUID         `json:"course_id"`
	WorkerID              uuid.UUID         `json:"worker_id"`
	ContractorID          *uuid.UUID        `json:"contractor_id,omitempty"`
	FolioNumber           string            `json:"folio_number"`
	Status                CertificateStatus `json:"status"`
	WorkerOccupationCode  string            `json:"worker_occupation_code"`
	WorkerCURP            string            `json:"worker_curp"`
	WorkerFullName        string            `json:"worker_full_name"`
	CompanyName           string            `json:"company_name"`
	CompanyRFC            string            `json:"company_rfc"`
	LegalRepName          string            `json:"legal_rep_name"`
	WorkerRepName         string            `json:"worker_rep_name"`
	InstructorName        string            `json:"instructor_name"`
	IssueDate             time.Time         `json:"issue_date"`
	ExpirationDate        time.Time         `json:"expiration_date"`
	ScorePercentage       float64           `json:"score_percentage"`
	DigitalSignatureToken string            `json:"digital_signature_token,omitempty"`
	DocumentR2Key         string            `json:"document_r2_key,omitempty"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

func NewTrainingCertificate(
	tenantID uuid.UUID,
	courseID uuid.UUID,
	workerID uuid.UUID,
	folioNumber string,
	curp string,
	workerName string,
	occupationCode string,
	companyName string,
	companyRFC string,
	legalRepName string,
	workerRepName string,
	instructorName string,
	issueDate time.Time,
	validityYears int,
	score float64,
) (*TrainingCertificate, error) {
	if strings.TrimSpace(folioNumber) == "" {
		return nil, errors.New("el número de folio de la constancia DC-3 es obligatorio")
	}
	if strings.TrimSpace(curp) == "" {
		return nil, errors.New("la CURP del trabajador es obligatoria para formato DC-3")
	}
	if score < 80.0 {
		return nil, errors.New("la calificación mínima aprobatoria para acreditación STPS DC-3 es de 80.00%")
	}
	if validityYears <= 0 {
		validityYears = 1
	}

	expDate := issueDate.AddDate(validityYears, 0, 0)
	now := time.Now().UTC()

	return &TrainingCertificate{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		CourseID:             courseID,
		WorkerID:             workerID,
		FolioNumber:          strings.TrimSpace(folioNumber),
		Status:               CertificateStatusIssued,
		WorkerOccupationCode: strings.TrimSpace(occupationCode),
		WorkerCURP:           strings.ToUpper(strings.TrimSpace(curp)),
		WorkerFullName:       strings.TrimSpace(workerName),
		CompanyName:          strings.TrimSpace(companyName),
		CompanyRFC:           strings.ToUpper(strings.TrimSpace(companyRFC)),
		LegalRepName:         strings.TrimSpace(legalRepName),
		WorkerRepName:        strings.TrimSpace(workerRepName),
		InstructorName:       strings.TrimSpace(instructorName),
		IssueDate:            issueDate,
		ExpirationDate:       expDate,
		ScorePercentage:      score,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

func (c *TrainingCertificate) IsExpired(at time.Time) bool {
	return at.After(c.ExpirationDate)
}
