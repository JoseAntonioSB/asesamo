package entity

import (
	"time"

	"github.com/google/uuid"
)

// ContractorWorker represents a worker affiliated with a contractor company.
type ContractorWorker struct {
	ID           uuid.UUID  `json:"id"`
	ContractorID uuid.UUID  `json:"contractor_id"`
	UserID       uuid.UUID  `json:"user_id"`
	JobTitle     *string    `json:"job_title,omitempty"`
	CURP         *string    `json:"curp,omitempty"`
	IMSSNumber   *string    `json:"imss_number,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
