package dto

import "github.com/google/uuid"

// AccessPointResponse represents an access point in catalog responses.
type AccessPointResponse struct {
	ID          uuid.UUID `json:"id"`
	PlantID     uuid.UUID `json:"plant_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
}
