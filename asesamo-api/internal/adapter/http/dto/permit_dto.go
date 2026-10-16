package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreatePermitRequest represents the JSON payload for POST /api/v1/permits.
type CreatePermitRequest struct {
	PlantID           uuid.UUID `json:"plant_id"`
	ContractorID      uuid.UUID `json:"contractor_id"`
	PermitType        string    `json:"permit_type"`
	WorkArea          string    `json:"work_area"`
	Description       string    `json:"description"`
	HazardDescription *string   `json:"hazard_description,omitempty"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Notes             *string   `json:"notes,omitempty"`
}

// AssignWorkersRequest represents the JSON payload for POST /api/v1/permits/{id}/workers.
type AssignWorkersRequest struct {
	WorkerIDs []uuid.UUID `json:"worker_ids"`
	Action    string      `json:"action,omitempty"` // "assign" (default) or "remove"
}

// UpdateChecklistItemDTO represents a single verification item in PATCH /api/v1/permits/{id}/checklist.
type UpdateChecklistItemDTO struct {
	ItemKey      string         `json:"item_key"`
	ItemCode     string         `json:"item_code,omitempty"` // alias for ItemKey
	IsVerified   bool           `json:"is_verified"`
	IsChecked    bool           `json:"is_checked,omitempty"` // alias for IsVerified
	Notes        *string        `json:"notes,omitempty"`
	Measurements map[string]any `json:"measurements,omitempty"`
}

// UpdateChecklistRequest represents the JSON payload for PATCH /api/v1/permits/{id}/checklist.
type UpdateChecklistRequest struct {
	Items []UpdateChecklistItemDTO `json:"items"`
}

// ApprovePermitRequest represents the JSON payload for POST /api/v1/permits/{id}/approve.
type ApprovePermitRequest struct {
	Notes         *string `json:"notes,omitempty"`
	SignatureData *string `json:"signature_data,omitempty"`
}

// TransitionStatusRequest represents the JSON payload for POST /api/v1/permits/{id}/status.
type TransitionStatusRequest struct {
	Action        string  `json:"action"` // "submit", "suspend", "resume", "close", "cancel"
	Reason        *string `json:"reason,omitempty"`
	Notes         *string `json:"notes,omitempty"`
	SignatureData *string `json:"signature_data,omitempty"`
}
