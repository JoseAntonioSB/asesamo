package epp

import (
	"context"
	"errors"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type DeliverEPPInput struct {
	EPPItemID     uuid.UUID  `json:"epp_item_id"`
	WorkerID      uuid.UUID  `json:"worker_id"`
	ContractorID  *uuid.UUID `json:"contractor_id,omitempty"`
	DeliveryDate  time.Time  `json:"delivery_date"`
	SignatureData string     `json:"signature_data,omitempty"`
	Notes         string     `json:"notes,omitempty"`
}

type DeliverEPPUseCase struct {
	repo repository.EPPRepository
}

func NewDeliverEPPUseCase(repo repository.EPPRepository) *DeliverEPPUseCase {
	return &DeliverEPPUseCase{repo: repo}
}

func (uc *DeliverEPPUseCase) Execute(ctx context.Context, tenantID, delivererID uuid.UUID, input DeliverEPPInput) (*entity.EPPDelivery, error) {
	item, err := uc.repo.GetItemByID(ctx, tenantID, input.EPPItemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("artículo EPP no encontrado")
	}

	if input.DeliveryDate.IsZero() {
		input.DeliveryDate = time.Now()
	}

	delivery, err := entity.NewEPPDelivery(
		tenantID,
		input.EPPItemID,
		input.WorkerID,
		input.ContractorID,
		input.DeliveryDate,
		item.LifespanDays,
		delivererID,
		input.SignatureData,
		input.Notes,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.RecordDelivery(ctx, delivery); err != nil {
		return nil, err
	}

	return delivery, nil
}
