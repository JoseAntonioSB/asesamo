package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type TrainingRepository interface {
	CreateCourse(ctx context.Context, course *entity.TrainingCourse) error
	GetCourseByID(ctx context.Context, tenantID, courseID uuid.UUID) (*entity.TrainingCourse, error)
	ListCourses(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCourse, error)
	CreateCertificate(ctx context.Context, cert *entity.TrainingCertificate) error
	GetCertificateByID(ctx context.Context, tenantID, certID uuid.UUID) (*entity.TrainingCertificate, error)
	ListCertificatesByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.TrainingCertificate, error)
	ListCertificates(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCertificate, error)
}
