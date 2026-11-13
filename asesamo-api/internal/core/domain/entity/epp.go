package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EPPCategory string

const (
	EPPCategoryHead        EPPCategory = "head"
	EPPCategoryEyesFace    EPPCategory = "eyes_face"
	EPPCategoryHearing     EPPCategory = "hearing"
	EPPCategoryRespiratory EPPCategory = "respiratory"
	EPPCategoryHands       EPPCategory = "hands"
	EPPCategoryBody        EPPCategory = "body"
	EPPCategoryFeet        EPPCategory = "feet"
	EPPCategoryFallArrest  EPPCategory = "fall_arrest"
)

type EPPDeliveryStatus string

const (
	EPPDeliveryStatusDeliveredActive EPPDeliveryStatus = "delivered_active"
	EPPDeliveryStatusReplaced        EPPDeliveryStatus = "replaced"
	EPPDeliveryStatusExpired         EPPDeliveryStatus = "expired"
	EPPDeliveryStatusReturned        EPPDeliveryStatus = "returned"
)

type EPPItem struct {
	ID                uuid.UUID   `json:"id"`
	TenantID          uuid.UUID   `json:"tenant_id"`
	Code              string      `json:"code"`
	Name              string      `json:"name"`
	Category          EPPCategory `json:"category"`
	NormativeStandard string      `json:"normative_standard,omitempty"`
	LifespanDays      int         `json:"lifespan_days"`
	StockQuantity     int         `json:"stock_quantity"`
	IsActive          bool        `json:"is_active"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type EPPDelivery struct {
	ID                uuid.UUID         `json:"id"`
	TenantID          uuid.UUID         `json:"tenant_id"`
	EPPItemID         uuid.UUID         `json:"epp_item_id"`
	WorkerID          uuid.UUID         `json:"worker_id"`
	ContractorID      *uuid.UUID        `json:"contractor_id,omitempty"`
	DeliveryDate      time.Time         `json:"delivery_date"`
	ExpirationDate    time.Time         `json:"expiration_date"`
	Status            EPPDeliveryStatus `json:"status"`
	DeliveredByUserID uuid.UUID         `json:"delivered_by_user_id"`
	SignatureData     string            `json:"signature_data,omitempty"`
	Notes             string            `json:"notes,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

func NewEPPDelivery(
	tenantID uuid.UUID,
	eppItemID uuid.UUID,
	workerID uuid.UUID,
	contractorID *uuid.UUID,
	deliveryDate time.Time,
	lifespanDays int,
	deliveredByUserID uuid.UUID,
	signature string,
	notes string,
) (*EPPDelivery, error) {
	if lifespanDays <= 0 {
		lifespanDays = 180
	}
	now := time.Now()
	expirationDate := deliveryDate.AddDate(0, 0, lifespanDays)

	return &EPPDelivery{
		ID:                uuid.New(),
		TenantID:          tenantID,
		EPPItemID:         eppItemID,
		WorkerID:          workerID,
		ContractorID:      contractorID,
		DeliveryDate:      deliveryDate,
		ExpirationDate:    expirationDate,
		Status:            EPPDeliveryStatusDeliveredActive,
		DeliveredByUserID: deliveredByUserID,
		SignatureData:     signature,
		Notes:             notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (d *EPPDelivery) IsExpiredAt(checkDate time.Time) bool {
	return checkDate.After(d.ExpirationDate)
}

func (d *EPPDelivery) MarkReplaced(reason string) error {
	if d.Status == EPPDeliveryStatusReplaced {
		return errors.New("la dotación ya se encuentra marcada como reemplazada")
	}
	d.Status = EPPDeliveryStatusReplaced
	if d.Notes != "" {
		d.Notes = d.Notes + " | Reemplazo: " + reason
	} else {
		d.Notes = "Reemplazo: " + reason
	}
	d.UpdatedAt = time.Now()
	return nil
}
