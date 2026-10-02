package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// DocumentRepository defines the persistence boundary for Document entities.
type DocumentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error)
	ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error)
	ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error)
}
