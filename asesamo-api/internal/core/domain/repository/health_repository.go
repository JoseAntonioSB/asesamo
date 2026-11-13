package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type HealthRepository interface {
	UpsertMedicalProfile(ctx context.Context, profile *entity.WorkerMedicalProfile) error
	GetMedicalProfileByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.WorkerMedicalProfile, error)
	CreateMedicalExam(ctx context.Context, exam *entity.OccupationalMedicalExam) error
	GetMedicalExamByID(ctx context.Context, tenantID, examID uuid.UUID) (*entity.OccupationalMedicalExam, error)
	ListExamsByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.OccupationalMedicalExam, error)
	GetLatestExamByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.OccupationalMedicalExam, error)
	ListAllExams(ctx context.Context, tenantID uuid.UUID) ([]entity.OccupationalMedicalExam, error)
}
