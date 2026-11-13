package training

import (
	"context"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type IssueDC3Input struct {
	TenantID       uuid.UUID
	CourseID       uuid.UUID
	WorkerID       uuid.UUID
	ContractorID   *uuid.UUID
	WorkerCURP     string
	WorkerFullName string
	OccupationCode string
	CompanyName    string
	CompanyRFC     string
	LegalRepName   string
	WorkerRepName  string
	InstructorName string
	IssueDate      time.Time
	Score          float64
}

type IssueDC3Output struct {
	Certificate *entity.TrainingCertificate
}

type IssueDC3CertificateUseCase struct {
	repo repository.TrainingRepository
}

func NewIssueDC3CertificateUseCase(repo repository.TrainingRepository) *IssueDC3CertificateUseCase {
	return &IssueDC3CertificateUseCase{repo: repo}
}

func (uc *IssueDC3CertificateUseCase) Execute(ctx context.Context, input IssueDC3Input) (*IssueDC3Output, error) {
	// 1. Verify course exists
	course, err := uc.repo.GetCourseByID(ctx, input.TenantID, input.CourseID)
	if err != nil {
		return nil, fmt.Errorf("error al verificar curso: %w", err)
	}

	// 2. Generate folio
	now := time.Now().UTC()
	folio := fmt.Sprintf("DC3-%d-%s", now.Year(), uuid.New().String()[:8])

	occupationCode := input.OccupationCode
	if occupationCode == "" {
		occupationCode = "8100"
	}

	issueDate := input.IssueDate
	if issueDate.IsZero() {
		issueDate = now
	}

	// 3. Create domain entity
	cert, err := entity.NewTrainingCertificate(
		input.TenantID,
		course.ID,
		input.WorkerID,
		folio,
		input.WorkerCURP,
		input.WorkerFullName,
		occupationCode,
		input.CompanyName,
		input.CompanyRFC,
		input.LegalRepName,
		input.WorkerRepName,
		input.InstructorName,
		issueDate,
		course.ValidityYears,
		input.Score,
	)
	if err != nil {
		return nil, err
	}
	cert.ContractorID = input.ContractorID

	// 4. Persist in database
	if err := uc.repo.CreateCertificate(ctx, cert); err != nil {
		return nil, fmt.Errorf("error al guardar constancia DC-3: %w", err)
	}

	return &IssueDC3Output{Certificate: cert}, nil
}
