package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// PlantRepository defines the persistence boundary for Plant entities.
type PlantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error)
}
