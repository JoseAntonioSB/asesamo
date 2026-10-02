package entity

import (
	"time"

	"github.com/google/uuid"
)

// Plant represents an industrial facility belonging to a tenant.
type Plant struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            uuid.UUID  `json:"tenant_id"`
	Name                string     `json:"name"`
	Address             *string    `json:"address,omitempty"`
	City                *string    `json:"city,omitempty"`
	State               *string    `json:"state,omitempty"`
	BlockOnExpiredDocs  bool       `json:"block_on_expired_docs"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
}
