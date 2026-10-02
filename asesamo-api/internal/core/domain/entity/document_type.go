package entity

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType represents a catalog definition of an industrial safety or legal document.
type DocumentType struct {
	ID                  uuid.UUID  `json:"id"`
	TenantID            *uuid.UUID `json:"tenant_id,omitempty"`
	Code                string     `json:"code"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	AppliesTo           string     `json:"applies_to"`
	DefaultValidityDays *int       `json:"default_validity_days,omitempty"`
	IsRequiredDefault   bool       `json:"is_required_default"`
	IsSystem            bool       `json:"is_system"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
