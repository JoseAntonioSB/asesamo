package entity_test

import (
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

func TestTrainingCourse_Validation(t *testing.T) {
	tenantID := uuid.New()

	t.Run("valid course", func(t *testing.T) {
		course, err := entity.NewTrainingCourse(
			tenantID,
			"CUR-ALT-001",
			"Seguridad en Trabajos en Altura",
			entity.CourseAreaSafety,
			"6000",
			8,
			1,
			"NOM-009-STPS-2011",
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if course.Code != "CUR-ALT-001" {
			t.Errorf("expected code CUR-ALT-001, got %s", course.Code)
		}
		if !course.IsActive {
			t.Errorf("expected course to be active by default")
		}
	})

	t.Run("empty title returns error", func(t *testing.T) {
		_, err := entity.NewTrainingCourse(
			tenantID,
			"CUR-ALT-001",
			"",
			entity.CourseAreaSafety,
			"6000",
			8,
			1,
			"NOM-009-STPS-2011",
		)
		if err == nil {
			t.Fatalf("expected error for empty title")
		}
	})

	t.Run("invalid duration returns error", func(t *testing.T) {
		_, err := entity.NewTrainingCourse(
			tenantID,
			"CUR-ALT-001",
			"Trabajos en Altura",
			entity.CourseAreaSafety,
			"6000",
			0, // 0 hours
			1,
			"NOM-009-STPS-2011",
		)
		if err == nil {
			t.Fatalf("expected error for zero duration")
		}
	})
}

func TestTrainingCertificate_DC3Validation(t *testing.T) {
	tenantID := uuid.New()
	courseID := uuid.New()
	workerID := uuid.New()

	t.Run("valid DC-3 certificate issue", func(t *testing.T) {
		cert, err := entity.NewTrainingCertificate(
			tenantID,
			courseID,
			workerID,
			"DC3-2026-0001",
			"MENC850101HDFR01",
			"Carlos Mendoza",
			"8100",
			"Grupo Industrial Norteño SA de CV",
			"GIN950418KP3",
			"Ing. Laura Sanchez",
			"Carlos Mendoza",
			"Ing. Fernando Capacitador",
			time.Now().AddDate(0, 0, -10),
			1, // 1 year validity
			95.0,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cert.Status != entity.CertificateStatusIssued {
			t.Errorf("expected status issued, got %s", cert.Status)
		}
		if cert.IsExpired(time.Now()) {
			t.Errorf("certificate should not be expired")
		}
	})

	t.Run("failing score (< 80) cannot issue valid DC-3", func(t *testing.T) {
		_, err := entity.NewTrainingCertificate(
			tenantID,
			courseID,
			workerID,
			"DC3-2026-0002",
			"MENC850101HDFR01",
			"Carlos Mendoza",
			"8100",
			"Grupo Industrial Norteño SA de CV",
			"GIN950418KP3",
			"Ing. Laura Sanchez",
			"Carlos Mendoza",
			"Ing. Fernando Capacitador",
			time.Now(),
			1,
			75.0, // Score below STPS passing standard (80%)
		)
		if err == nil {
			t.Fatalf("expected error for failing score (< 80)")
		}
	})

	t.Run("expired certificate check", func(t *testing.T) {
		issueDate := time.Now().AddDate(-2, 0, 0) // 2 years ago
		cert, err := entity.NewTrainingCertificate(
			tenantID,
			courseID,
			workerID,
			"DC3-2024-0001",
			"MENC850101HDFR01",
			"Carlos Mendoza",
			"8100",
			"Grupo Industrial Norteño SA de CV",
			"GIN950418KP3",
			"Ing. Laura Sanchez",
			"Carlos Mendoza",
			"Ing. Fernando Capacitador",
			issueDate,
			1, // 1 year validity -> expired 1 year ago
			100.0,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !cert.IsExpired(time.Now()) {
			t.Errorf("certificate should be expired")
		}
	})
}
