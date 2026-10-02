package entity

import (
	"time"

	"github.com/google/uuid"
)

// PlantContractorStatus represents the activation state of a contractor for a specific plant.
type PlantContractorStatus string

const (
	PlantContractorPending   PlantContractorStatus = "pending"
	PlantContractorActive    PlantContractorStatus = "active"
	PlantContractorSuspended PlantContractorStatus = "suspended"
	PlantContractorBlocked   PlantContractorStatus = "blocked"
)

// PlantContractor represents the association between an industrial plant and a contractor company.
type PlantContractor struct {
	ID           uuid.UUID             `json:"id"`
	PlantID      uuid.UUID             `json:"plant_id"`
	TenantID     uuid.UUID             `json:"tenant_id"`
	ContractorID uuid.UUID             `json:"contractor_id"`
	Status       PlantContractorStatus `json:"status"`
	InvitedBy    *uuid.UUID            `json:"invited_by,omitempty"`
	InvitedAt    time.Time             `json:"invited_at"`
	ActivatedAt  *time.Time            `json:"activated_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}
