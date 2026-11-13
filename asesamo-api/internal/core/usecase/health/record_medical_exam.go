package health

import (
	"context"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type RecordMedicalExamInput struct {
	TenantID            uuid.UUID
	WorkerID            uuid.UUID
	ContractorID        *uuid.UUID
	ExamType            entity.MedicalExamType
	FitnessStatus       entity.MedicalFitnessStatus
	ExamDate            time.Time
	DoctorName          string
	DoctorLicense       string
	ClinicOrInstitution string
	RestrictionsNotes   string
	Recommendations     string
	VisualAcuityResult  string
	AudiometryResult    string
	SpirometryResult    string
	AntidopingResult    string
	BloodPressure       string
}

type RecordMedicalExamOutput struct {
	Exam *entity.OccupationalMedicalExam
}

type RecordMedicalExamUseCase struct {
	repo repository.HealthRepository
}

func NewRecordMedicalExamUseCase(repo repository.HealthRepository) *RecordMedicalExamUseCase {
	return &RecordMedicalExamUseCase{repo: repo}
}

func (uc *RecordMedicalExamUseCase) Execute(ctx context.Context, input RecordMedicalExamInput) (*RecordMedicalExamOutput, error) {
	now := time.Now().UTC()
	examDate := input.ExamDate
	if examDate.IsZero() {
		examDate = now
	}

	// Calculate expiration date by exam type
	var expirationDate time.Time
	switch input.ExamType {
	case entity.MedicalExamTypeHeightClearance, entity.MedicalExamTypeConfinedSpace:
		// Specialized high risk clearance is valid for 6 months
		expirationDate = examDate.AddDate(0, 6, 0)
	case entity.MedicalExamTypeAntidoping:
		// Antidoping valid for 3 months
		expirationDate = examDate.AddDate(0, 3, 0)
	default:
		// General periodic annual exam valid for 1 year
		expirationDate = examDate.AddDate(1, 0, 0)
	}

	examNumber := fmt.Sprintf("MED-%d-%s", now.Year(), uuid.New().String()[:8])

	exam, err := entity.NewOccupationalMedicalExam(
		input.TenantID,
		input.WorkerID,
		examNumber,
		input.ExamType,
		input.FitnessStatus,
		examDate,
		expirationDate,
		input.DoctorName,
		input.DoctorLicense,
		input.ClinicOrInstitution,
	)
	if err != nil {
		return nil, err
	}
	exam.ContractorID = input.ContractorID
	exam.RestrictionsNotes = input.RestrictionsNotes
	exam.Recommendations = input.Recommendations
	exam.VisualAcuityResult = input.VisualAcuityResult
	exam.AudiometryResult = input.AudiometryResult
	exam.SpirometryResult = input.SpirometryResult
	exam.AntidopingResult = input.AntidopingResult
	exam.BloodPressure = input.BloodPressure

	if err := uc.repo.CreateMedicalExam(ctx, exam); err != nil {
		return nil, fmt.Errorf("error al guardar examen médico: %w", err)
	}

	return &RecordMedicalExamOutput{Exam: exam}, nil
}
