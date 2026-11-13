package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MedicalFitnessStatus string

const (
	MedicalFitnessFit             MedicalFitnessStatus = "fit"
	MedicalFitnessFitRestrictions MedicalFitnessStatus = "fit_with_restrictions"
	MedicalFitnessUnfitTemporary  MedicalFitnessStatus = "unfit_temporary"
	MedicalFitnessUnfitPermanent  MedicalFitnessStatus = "unfit_permanent"
)

type MedicalExamType string

const (
	MedicalExamTypePrePlacement    MedicalExamType = "pre_placement"
	MedicalExamTypeAnnual          MedicalExamType = "periodic_annual"
	MedicalExamTypeHeightClearance MedicalExamType = "height_clearance"
	MedicalExamTypeConfinedSpace   MedicalExamType = "confined_space"
	MedicalExamTypePostIncident    MedicalExamType = "post_incident"
	MedicalExamTypeAntidoping      MedicalExamType = "antidoping"
)

var validBloodTypes = map[string]bool{
	"A+":  true,
	"A-":  true,
	"B+":  true,
	"B-":  true,
	"AB+": true,
	"AB-": true,
	"O+":  true,
	"O-":  true,
}

// WorkerMedicalProfile represents the base health profile and emergency contacts of a worker.
type WorkerMedicalProfile struct {
	ID                            uuid.UUID `json:"id"`
	TenantID                      uuid.UUID `json:"tenant_id"`
	WorkerID                      uuid.UUID `json:"worker_id"`
	BloodType                     string    `json:"blood_type"`
	Allergies                     string    `json:"allergies,omitempty"`
	ChronicConditions             string    `json:"chronic_conditions,omitempty"`
	EmergencyContactName          string    `json:"emergency_contact_name"`
	EmergencyContactPhone         string    `json:"emergency_contact_phone"`
	EmergencyContactRelationship  string    `json:"emergency_contact_relationship"`
	CreatedAt                     time.Time `json:"created_at"`
	UpdatedAt                     time.Time `json:"updated_at"`
}

func NewWorkerMedicalProfile(
	tenantID uuid.UUID,
	workerID uuid.UUID,
	bloodType string,
	allergies string,
	chronicConditions string,
	emergencyContactName string,
	emergencyContactPhone string,
	relationship string,
) (*WorkerMedicalProfile, error) {
	bt := strings.ToUpper(strings.TrimSpace(bloodType))
	if !validBloodTypes[bt] {
		return nil, errors.New("el tipo de sangre no es válido (ej. O+, A+, B+, AB-)")
	}
	if strings.TrimSpace(emergencyContactName) == "" {
		return nil, errors.New("el nombre del contacto de emergencia es obligatorio")
	}
	if strings.TrimSpace(emergencyContactPhone) == "" {
		return nil, errors.New("el teléfono de emergencia es obligatorio")
	}

	now := time.Now().UTC()
	return &WorkerMedicalProfile{
		ID:                           uuid.New(),
		TenantID:                     tenantID,
		WorkerID:                     workerID,
		BloodType:                    bt,
		Allergies:                    strings.TrimSpace(allergies),
		ChronicConditions:            strings.TrimSpace(chronicConditions),
		EmergencyContactName:         strings.TrimSpace(emergencyContactName),
		EmergencyContactPhone:        strings.TrimSpace(emergencyContactPhone),
		EmergencyContactRelationship: strings.TrimSpace(relationship),
		CreatedAt:                    now,
		UpdatedAt:                    now,
	}, nil
}

// OccupationalMedicalExam represents an official medical exam according to NOM-030-STPS.
type OccupationalMedicalExam struct {
	ID                  uuid.UUID            `json:"id"`
	TenantID            uuid.UUID            `json:"tenant_id"`
	WorkerID            uuid.UUID            `json:"worker_id"`
	ContractorID        *uuid.UUID           `json:"contractor_id,omitempty"`
	ExamNumber          string               `json:"exam_number"`
	ExamType            MedicalExamType      `json:"exam_type"`
	FitnessStatus       MedicalFitnessStatus `json:"fitness_status"`
	ExamDate            time.Time            `json:"exam_date"`
	ExpirationDate      time.Time            `json:"expiration_date"`
	DoctorName          string               `json:"doctor_name"`
	DoctorLicense       string               `json:"doctor_license"`
	ClinicOrInstitution string               `json:"clinic_or_institution"`
	RestrictionsNotes   string               `json:"restrictions_notes,omitempty"`
	Recommendations     string               `json:"recommendations,omitempty"`
	VisualAcuityResult  string               `json:"visual_acuity_result,omitempty"`
	AudiometryResult    string               `json:"audiometry_result,omitempty"`
	SpirometryResult    string               `json:"spirometry_result,omitempty"`
	AntidopingResult    string               `json:"antidoping_result,omitempty"`
	BloodPressure       string               `json:"blood_pressure,omitempty"`
	CertificateR2Key    string               `json:"certificate_r2_key,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

func NewOccupationalMedicalExam(
	tenantID uuid.UUID,
	workerID uuid.UUID,
	examNumber string,
	examType MedicalExamType,
	fitnessStatus MedicalFitnessStatus,
	examDate time.Time,
	expirationDate time.Time,
	doctorName string,
	doctorLicense string,
	clinic string,
) (*OccupationalMedicalExam, error) {
	if strings.TrimSpace(examNumber) == "" {
		return nil, errors.New("el número de folio del examen médico es obligatorio")
	}
	if strings.TrimSpace(doctorName) == "" {
		return nil, errors.New("el nombre del médico dictaminador es obligatorio")
	}
	if strings.TrimSpace(doctorLicense) == "" {
		return nil, errors.New("la cédula profesional del médico es obligatoria")
	}
	if expirationDate.Before(examDate) {
		return nil, errors.New("la fecha de vencimiento no puede ser anterior a la fecha de realización del examen")
	}

	now := time.Now().UTC()
	return &OccupationalMedicalExam{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		WorkerID:            workerID,
		ExamNumber:          strings.TrimSpace(examNumber),
		ExamType:            examType,
		FitnessStatus:       fitnessStatus,
		ExamDate:            examDate,
		ExpirationDate:      expirationDate,
		DoctorName:          strings.TrimSpace(doctorName),
		DoctorLicense:       strings.TrimSpace(doctorLicense),
		ClinicOrInstitution: strings.TrimSpace(clinic),
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

func (e *OccupationalMedicalExam) IsFitForWork(at time.Time) bool {
	if at.After(e.ExpirationDate) {
		return false
	}
	return e.FitnessStatus == MedicalFitnessFit || e.FitnessStatus == MedicalFitnessFitRestrictions
}
