package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// AccessPointRepository defines the persistence boundary for AccessPoint entities.
type AccessPointRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error)
	ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)
}
