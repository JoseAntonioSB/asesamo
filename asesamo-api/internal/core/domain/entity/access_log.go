package entity

import (
	"time"

	"github.com/google/uuid"
)

// AccessResult represents the access gate decision outcome.
type AccessResult string

const (
	AccessResultGranted            AccessResult = "granted"
	AccessResultGrantedWithWarning AccessResult = "granted_with_warning"
	AccessResultDenied             AccessResult = "denied"
)

// AccessLog represents an immutable record of a worker check-in / check-out attempt at a gate.
type AccessLog struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	PlantID          uuid.UUID        `json:"plant_id"`
	AccessPointID    uuid.UUID        `json:"access_point_id"`
	UserID           uuid.UUID        `json:"user_id"`
	ContractorID     *uuid.UUID       `json:"contractor_id,omitempty"`
	ProcessedBy      *uuid.UUID       `json:"processed_by,omitempty"`
	CheckedInAt      time.Time        `json:"checked_in_at"`
	CheckedOutAt     *time.Time       `json:"checked_out_at,omitempty"`
	Result           AccessResult     `json:"result"`
	DenialReason     *string          `json:"denial_reason,omitempty"`
	DocumentSnapshot DocumentSnapshot `json:"document_snapshot"`
	Notes            *string          `json:"notes,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}
