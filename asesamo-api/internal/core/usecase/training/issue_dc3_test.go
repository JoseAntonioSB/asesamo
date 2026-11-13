package training_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	usecase "asesamo-api/internal/core/usecase/training"
	"github.com/google/uuid"
)

type mockTrainingRepo struct {
	courses      map[uuid.UUID]*entity.TrainingCourse
	certificates map[uuid.UUID]*entity.TrainingCertificate
}

func newMockTrainingRepo() *mockTrainingRepo {
	return &mockTrainingRepo{
		courses:      make(map[uuid.UUID]*entity.TrainingCourse),
		certificates: make(map[uuid.UUID]*entity.TrainingCertificate),
	}
}

func (m *mockTrainingRepo) CreateCourse(ctx context.Context, course *entity.TrainingCourse) error {
	m.courses[course.ID] = course
	return nil
}

func (m *mockTrainingRepo) GetCourseByID(ctx context.Context, tenantID, courseID uuid.UUID) (*entity.TrainingCourse, error) {
	c, ok := m.courses[courseID]
	if !ok || c.TenantID != tenantID {
		return nil, errors.New("curso no encontrado")
	}
	return c, nil
}

func (m *mockTrainingRepo) ListCourses(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCourse, error) {
	var list []entity.TrainingCourse
	for _, c := range m.courses {
		if c.TenantID == tenantID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *mockTrainingRepo) CreateCertificate(ctx context.Context, cert *entity.TrainingCertificate) error {
	m.certificates[cert.ID] = cert
	return nil
}

func (m *mockTrainingRepo) GetCertificateByID(ctx context.Context, tenantID, certID uuid.UUID) (*entity.TrainingCertificate, error) {
	c, ok := m.certificates[certID]
	if !ok || c.TenantID != tenantID {
		return nil, errors.New("constancia no encontrada")
	}
	return c, nil
}

func (m *mockTrainingRepo) ListCertificatesByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.TrainingCertificate, error) {
	var list []entity.TrainingCertificate
	for _, c := range m.certificates {
		if c.TenantID == tenantID && c.WorkerID == workerID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *mockTrainingRepo) ListCertificates(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCertificate, error) {
	var list []entity.TrainingCertificate
	for _, c := range m.certificates {
		if c.TenantID == tenantID {
			list = append(list, *c)
		}
	}
	return list, nil
}

func TestIssueDC3CertificateUseCase(t *testing.T) {
	repo := newMockTrainingRepo()
	uc := usecase.NewIssueDC3CertificateUseCase(repo)

	tenantID := uuid.New()
	workerID := uuid.New()

	// Seed course
	course, _ := entity.NewTrainingCourse(
		tenantID,
		"CUR-ALT-001",
		"Trabajos en Altura",
		entity.CourseAreaSafety,
		"6000",
		8,
		1,
		"NOM-009-STPS-2011",
	)
	repo.CreateCourse(context.Background(), course)

	t.Run("successfully issue DC-3 certificate", func(t *testing.T) {
		input := usecase.IssueDC3Input{
			TenantID:       tenantID,
			CourseID:       course.ID,
			WorkerID:       workerID,
			WorkerCURP:     "MENC850101HDFR01",
			WorkerFullName: "Carlos Mendoza",
			OccupationCode: "8100",
			CompanyName:    "Grupo Industrial Norteño SA de CV",
			CompanyRFC:     "GIN950418KP3",
			LegalRepName:   "Ing. Laura Sanchez",
			WorkerRepName:  "Carlos Mendoza",
			InstructorName: "Ing. Capacitador Certificado",
			IssueDate:      time.Now(),
			Score:          95.0,
		}

		output, err := uc.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if output.Certificate == nil {
			t.Fatalf("expected certificate to be returned")
		}
		if output.Certificate.FolioNumber == "" {
			t.Errorf("expected generated folio number")
		}
		if output.Certificate.Status != entity.CertificateStatusIssued {
			t.Errorf("expected status issued, got %s", output.Certificate.Status)
		}
	})

	t.Run("fails when course does not exist", func(t *testing.T) {
		input := usecase.IssueDC3Input{
			TenantID:       tenantID,
			CourseID:       uuid.New(), // Random non-existent course
			WorkerID:       workerID,
			WorkerCURP:     "MENC850101HDFR01",
			WorkerFullName: "Carlos Mendoza",
			CompanyName:    "Empresa",
			CompanyRFC:     "RFC123456789",
			Score:          90.0,
		}

		_, err := uc.Execute(context.Background(), input)
		if err == nil {
			t.Fatalf("expected error for non-existent course")
		}
	})

	t.Run("fails when score is below 80", func(t *testing.T) {
		input := usecase.IssueDC3Input{
			TenantID:       tenantID,
			CourseID:       course.ID,
			WorkerID:       workerID,
			WorkerCURP:     "MENC850101HDFR01",
			WorkerFullName: "Carlos Mendoza",
			CompanyName:    "Empresa",
			CompanyRFC:     "RFC123456789",
			Score:          70.0,
		}

		_, err := uc.Execute(context.Background(), input)
		if err == nil {
			t.Fatalf("expected error for failing score")
		}
	})
}
