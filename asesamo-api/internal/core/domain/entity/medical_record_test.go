package entity_test

import (
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

func TestWorkerMedicalProfile_Validation(t *testing.T) {
	tenantID := uuid.New()
	workerID := uuid.New()

	t.Run("valid profile", func(t *testing.T) {
		profile, err := entity.NewWorkerMedicalProfile(
			tenantID,
			workerID,
			"O+",
			"Penicilina",
			"Ninguna",
			"Maria Mendoza",
			"4421112233",
			"Esposa",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if profile.BloodType != "O+" {
			t.Errorf("expected blood type O+, got %s", profile.BloodType)
		}
	})

	t.Run("invalid blood type returns error", func(t *testing.T) {
		_, err := entity.NewWorkerMedicalProfile(
			tenantID,
			workerID,
			"INVALID",
			"",
			"",
			"Contacto",
			"1234567890",
			"Familiar",
		)
		if err == nil {
			t.Fatalf("expected error for invalid blood type")
		}
	})
}

func TestOccupationalMedicalExam_FitnessRules(t *testing.T) {
	tenantID := uuid.New()
	workerID := uuid.New()

	t.Run("valid fit exam", func(t *testing.T) {
		exam, err := entity.NewOccupationalMedicalExam(
			tenantID,
			workerID,
			"MED-2026-0001",
			entity.MedicalExamTypeAnnual,
			entity.MedicalFitnessFit,
			time.Now().AddDate(0, -1, 0),
			time.Now().AddDate(0, 11, 0),
			"Dr. Alejandro Gomez",
			"CED-MED-8849201",
			"Servicio Médico Industrial",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !exam.IsFitForWork(time.Now()) {
			t.Errorf("expected worker to be fit for work")
		}
	})

	t.Run("expired exam is not fit for work", func(t *testing.T) {
		exam, err := entity.NewOccupationalMedicalExam(
			tenantID,
			workerID,
			"MED-2025-0001",
			entity.MedicalExamTypeAnnual,
			entity.MedicalFitnessFit,
			time.Now().AddDate(-1, -1, 0),
			time.Now().AddDate(0, -1, 0), // expired 1 month ago
			"Dr. Alejandro Gomez",
			"CED-MED-8849201",
			"Servicio Médico Industrial",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exam.IsFitForWork(time.Now()) {
			t.Errorf("expected expired exam to be unfit for work")
		}
	})

	t.Run("unfit temporary status is not fit for work", func(t *testing.T) {
		exam, err := entity.NewOccupationalMedicalExam(
			tenantID,
			workerID,
			"MED-2026-0002",
			entity.MedicalExamTypeHeightClearance,
			entity.MedicalFitnessUnfitTemporary,
			time.Now(),
			time.Now().AddDate(0, 6, 0),
			"Dr. Alejandro Gomez",
			"CED-MED-8849201",
			"Servicio Médico Industrial",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if exam.IsFitForWork(time.Now()) {
			t.Errorf("expected unfit_temporary exam to return false for IsFitForWork")
		}
	})
}
