package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// AccessLogRepository defines the persistence boundary for AccessLog entities.
type AccessLogRepository interface {
	Create(ctx context.Context, log *entity.AccessLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error)
	ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error)
	CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error)
}
