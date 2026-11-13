package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	usecase "asesamo-api/internal/core/usecase/health"
	"github.com/google/uuid"
)

type mockHealthRepo struct {
	profiles map[uuid.UUID]*entity.WorkerMedicalProfile
	exams    map[uuid.UUID]*entity.OccupationalMedicalExam
}

func newMockHealthRepo() *mockHealthRepo {
	return &mockHealthRepo{
		profiles: make(map[uuid.UUID]*entity.WorkerMedicalProfile),
		exams:    make(map[uuid.UUID]*entity.OccupationalMedicalExam),
	}
}

func (m *mockHealthRepo) UpsertMedicalProfile(ctx context.Context, profile *entity.WorkerMedicalProfile) error {
	m.profiles[profile.WorkerID] = profile
	return nil
}

func (m *mockHealthRepo) GetMedicalProfileByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.WorkerMedicalProfile, error) {
	p, ok := m.profiles[workerID]
	if !ok || p.TenantID != tenantID {
		return nil, errors.New("perfil no encontrado")
	}
	return p, nil
}

func (m *mockHealthRepo) CreateMedicalExam(ctx context.Context, exam *entity.OccupationalMedicalExam) error {
	m.exams[exam.ID] = exam
	return nil
}

func (m *mockHealthRepo) GetMedicalExamByID(ctx context.Context, tenantID, examID uuid.UUID) (*entity.OccupationalMedicalExam, error) {
	e, ok := m.exams[examID]
	if !ok || e.TenantID != tenantID {
		return nil, errors.New("examen no encontrado")
	}
	return e, nil
}

func (m *mockHealthRepo) ListExamsByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.OccupationalMedicalExam, error) {
	var list []entity.OccupationalMedicalExam
	for _, e := range m.exams {
		if e.TenantID == tenantID && e.WorkerID == workerID {
			list = append(list, *e)
		}
	}
	return list, nil
}

func (m *mockHealthRepo) GetLatestExamByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.OccupationalMedicalExam, error) {
	var latest *entity.OccupationalMedicalExam
	for _, e := range m.exams {
		if e.TenantID == tenantID && e.WorkerID == workerID {
			if latest == nil || e.ExamDate.After(latest.ExamDate) {
				latest = e
			}
		}
	}
	if latest == nil {
		return nil, errors.New("no hay exámenes registrados")
	}
	return latest, nil
}

func (m *mockHealthRepo) ListAllExams(ctx context.Context, tenantID uuid.UUID) ([]entity.OccupationalMedicalExam, error) {
	var list []entity.OccupationalMedicalExam
	for _, e := range m.exams {
		if e.TenantID == tenantID {
			list = append(list, *e)
		}
	}
	return list, nil
}

func TestRecordMedicalExamUseCase(t *testing.T) {
	repo := newMockHealthRepo()
	uc := usecase.NewRecordMedicalExamUseCase(repo)

	tenantID := uuid.New()
	workerID := uuid.New()

	t.Run("records valid periodic exam and calculates expiration", func(t *testing.T) {
		input := usecase.RecordMedicalExamInput{
			TenantID:            tenantID,
			WorkerID:            workerID,
			ExamType:            entity.MedicalExamTypeAnnual,
			FitnessStatus:       entity.MedicalFitnessFit,
			ExamDate:            time.Now(),
			DoctorName:          "Dra. Maria Elena Sanchez",
			DoctorLicense:       "CED-8849102",
			ClinicOrInstitution: "Centro Médico Ocupacional",
			VisualAcuityResult:  "20/20",
			AudiometryResult:    "Normal bilateral",
			BloodPressure:       "118/78",
		}

		output, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if output.Exam == nil {
			t.Fatalf("expected exam to be returned")
		}
		if output.Exam.ExamNumber == "" {
			t.Errorf("expected generated exam number")
		}
		// Annual exam should expire in ~365 days
		expectedExp := input.ExamDate.AddDate(1, 0, 0)
		if output.Exam.ExpirationDate.Year() != expectedExp.Year() {
			t.Errorf("expected expiration year %d, got %d", expectedExp.Year(), output.Exam.ExpirationDate.Year())
		}
	})

	t.Run("calculates 6 months expiration for height clearance exam", func(t *testing.T) {
		input := usecase.RecordMedicalExamInput{
			TenantID:            tenantID,
			WorkerID:            workerID,
			ExamType:            entity.MedicalExamTypeHeightClearance,
			FitnessStatus:       entity.MedicalFitnessFit,
			ExamDate:            time.Now(),
			DoctorName:          "Dra. Maria Elena Sanchez",
			DoctorLicense:       "CED-8849102",
			ClinicOrInstitution: "Centro Médico Ocupacional",
		}

		output, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expectedExp := input.ExamDate.AddDate(0, 6, 0)
		if output.Exam.ExpirationDate.Month() != expectedExp.Month() {
			t.Errorf("expected expiration month %d, got %d", expectedExp.Month(), output.Exam.ExpirationDate.Month())
		}
	})
}
