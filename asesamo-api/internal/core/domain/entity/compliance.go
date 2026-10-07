package entity

import (
	"time"

	"github.com/google/uuid"
)

// WorkerAffiliation holds worker information and affiliated contractor details.
type WorkerAffiliation struct {
	WorkerID       uuid.UUID  `json:"worker_id"`
	FullName       string     `json:"full_name"`
	CURP           *string    `json:"curp,omitempty"`
	IMSSNumber     *string    `json:"imss_number,omitempty"`
	JobTitle       *string    `json:"job_title,omitempty"`
	AvatarR2Key    *string    `json:"avatar_r2_key,omitempty"`
	IsActive           bool       `json:"is_active"`
	UserIsActive       bool       `json:"user_is_active"`
	WorkerIsActive     bool       `json:"worker_is_active"`
	ContractorIsActive bool       `json:"contractor_is_active"`
	ContractorID       uuid.UUID  `json:"contractor_id"`
	ContractorName     string     `json:"contractor_name"`
	ContractorRFC      *string    `json:"contractor_rfc,omitempty"`
	REPSENumber        *string    `json:"repse_number,omitempty"`
	PlantStatus        string     `json:"plant_status"` // 'active' | 'pending' | 'suspended' | 'blocked'
}

// DocumentWithTypeCode holds document data along with its document type code.
type DocumentWithTypeCode struct {
	ID               uuid.UUID      `json:"id"`
	DocumentTypeCode string         `json:"document_type_code"`
	DocumentName     string         `json:"document_name"`
	AppliesTo        string         `json:"applies_to"` // 'worker' | 'contractor'
	Status           DocumentStatus `json:"status"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	IssuedAt         *time.Time     `json:"issued_at,omitempty"`
}

// RequiredDocumentRule specifies a document required by plant or system defaults.
type RequiredDocumentRule struct {
	DocumentTypeID   uuid.UUID `json:"document_type_id"`
	DocumentTypeCode string    `json:"document_type_code"`
	DocumentName     string    `json:"document_name"`
	AppliesTo        string    `json:"applies_to"`
	IsRequired       bool      `json:"is_required"`
	ValidityDays     *int      `json:"validity_days,omitempty"`
}
