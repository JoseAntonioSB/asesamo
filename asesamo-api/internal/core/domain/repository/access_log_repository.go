package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// AccessLogFilter specifies criteria for querying access logs with multi-column filtering.
type AccessLogFilter struct {
	TenantID      uuid.UUID
	PlantID       uuid.UUID
	AccessPointID uuid.UUID
	WorkerUserID  uuid.UUID
	Result        *entity.AccessResult
	From          *time.Time
	To            *time.Time
	Limit         int
	Offset        int
}

// AccessLogRepository defines the persistence boundary for AccessLog entities.
type AccessLogRepository interface {
	Create(ctx context.Context, log *entity.AccessLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error)
	ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error)
	ListWithFilters(ctx context.Context, filter AccessLogFilter) ([]entity.AccessLog, int, error)
	CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error)
}
