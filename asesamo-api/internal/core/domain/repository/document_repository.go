package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// DocumentRepository defines the persistence boundary for Document entities.
type DocumentRepository interface {
	Create(ctx context.Context, doc *entity.Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error)
	ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error)
	ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error)
	GetDocumentTypeByCode(ctx context.Context, code string) (*entity.DocumentType, error)
	ListRequiredDocumentTypes(ctx context.Context) ([]entity.DocumentType, error)
	GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error)
	GetContractorByWorkerUserID(ctx context.Context, workerUserID uuid.UUID) (*entity.Contractor, error)
	ListPendingReview(ctx context.Context, tenantID *uuid.UUID, statusFilter *entity.DocumentStatus) ([]entity.Document, error)
	UpdateReviewStatus(ctx context.Context, id uuid.UUID, status entity.DocumentStatus, reviewedBy uuid.UUID, rejectionReason *string) (*entity.Document, error)
}
