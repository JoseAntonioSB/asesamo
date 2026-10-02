package entity

import (
	"time"

	"github.com/google/uuid"
)

// AccessPoint represents an entry or exit gate / checkpoint at a plant.
type AccessPoint struct {
	ID          uuid.UUID  `json:"id"`
	PlantID     uuid.UUID  `json:"plant_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
