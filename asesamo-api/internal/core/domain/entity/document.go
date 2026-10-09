package entity

import (
	"time"

	"github.com/google/uuid"
)

// DocumentStatus represents the verification status of an uploaded document.
type DocumentStatus string

const (
	DocumentStatusPendingReview DocumentStatus = "pending_review"
	DocumentStatusApproved      DocumentStatus = "approved"
	DocumentStatusRejected      DocumentStatus = "rejected"
	DocumentStatusExpired       DocumentStatus = "expired"
)

// Document represents an uploaded compliance document.
type Document struct {
	ID              uuid.UUID      `json:"id"`
	UserID          *uuid.UUID     `json:"user_id,omitempty"`
	ContractorID    *uuid.UUID     `json:"contractor_id,omitempty"`
	DocumentTypeID   uuid.UUID      `json:"document_type_id"`
	DocumentTypeCode string         `json:"document_type_code,omitempty"`
	DocumentTypeName string         `json:"document_type_name,omitempty"`
	OwnerName        string         `json:"owner_name,omitempty"`
	OwnerType        string         `json:"owner_type,omitempty"` // "worker" | "contractor"
	ContractorName   string         `json:"contractor_name,omitempty"`
	ReviewedByName   string         `json:"reviewed_by_name,omitempty"`
	R2Key            string         `json:"r2_key"`
	FileName        string         `json:"file_name"`
	FileSizeBytes   *int           `json:"file_size_bytes,omitempty"`
	MimeType        *string        `json:"mime_type,omitempty"`
	IssuedAt        *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	Status          DocumentStatus `json:"status"`
	ReviewedBy      *uuid.UUID     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time     `json:"reviewed_at,omitempty"`
	RejectionReason *string        `json:"rejection_reason,omitempty"`
	AIConfidence    *float64       `json:"ai_confidence,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}
