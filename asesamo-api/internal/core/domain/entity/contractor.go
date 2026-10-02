package entity

import (
	"time"

	"github.com/google/uuid"
)

// ContractorType represents the classification of a contractor company.
type ContractorType string

const (
	ContractorTypeCompany    ContractorType = "company"
	ContractorTypeIndividual ContractorType = "individual"
)

// Contractor represents a global contractor company or independent contractor entity.
type Contractor struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	RFC            *string        `json:"rfc,omitempty"`
	REPSENumber    *string        `json:"repse_number,omitempty"`
	ContractorType ContractorType `json:"contractor_type"`
	AdminUserID    *uuid.UUID     `json:"admin_user_id,omitempty"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty"`
}
