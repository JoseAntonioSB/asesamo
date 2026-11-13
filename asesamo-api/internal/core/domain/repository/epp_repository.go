package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type EPPDeliveryFilter struct {
	WorkerID     *uuid.UUID
	ContractorID *uuid.UUID
	Category     *entity.EPPCategory
	Status       *entity.EPPDeliveryStatus
	Limit        int
	Offset       int
}

type EPPRepository interface {
	CreateItem(ctx context.Context, item *entity.EPPItem) error
	ListItemCatalog(ctx context.Context, tenantID uuid.UUID) ([]entity.EPPItem, error)
	GetItemByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EPPItem, error)

	RecordDelivery(ctx context.Context, delivery *entity.EPPDelivery) error
	ListDeliveries(ctx context.Context, tenantID uuid.UUID, filter EPPDeliveryFilter) ([]entity.EPPDelivery, int, error)
	UpdateDeliveryStatus(ctx context.Context, tenantID, deliveryID uuid.UUID, status entity.EPPDeliveryStatus, notes string) error
}
