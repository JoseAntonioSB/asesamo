package dto

import (
	"time"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// UploadDocumentResponse defines the HTTP response payload for POST /api/v1/documents/upload.
type UploadDocumentResponse struct {
	ID               uuid.UUID  `json:"id"`
	UserID           *uuid.UUID `json:"user_id"`
	ContractorID     *uuid.UUID `json:"contractor_id"`
	DocumentTypeID   uuid.UUID  `json:"document_type_id"`
	DocumentTypeCode string     `json:"document_type_code"`
	DocumentTypeName string     `json:"document_type_name,omitempty"`
	FileName         string     `json:"file_name"`
	FileSizeBytes    *int       `json:"file_size_bytes,omitempty"`
	MimeType         *string    `json:"mime_type,omitempty"`
	IssuedAt         *string    `json:"issued_at,omitempty"`
	ExpiresAt        *string    `json:"expires_at,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
}

// MyDocumentItemResponse defines an item in GET /api/v1/documents/my-documents.
type MyDocumentItemResponse struct {
	ID               uuid.UUID `json:"id"`
	DocumentTypeCode string    `json:"document_type_code"`
	DocumentTypeName string    `json:"document_type_name"`
	FileName         string    `json:"file_name"`
	Status           string    `json:"status"`
	IssuedAt         *string   `json:"issued_at,omitempty"`
	ExpiresAt        *string   `json:"expires_at,omitempty"`
	RejectionReason  *string   `json:"rejection_reason"`
	CreatedAt        time.Time `json:"created_at"`
}

// DocumentTypeResponse represents a catalog entry for document types.
type DocumentTypeResponse struct {
	ID                  uuid.UUID `json:"id"`
	Code                string    `json:"code"`
	Name                string    `json:"name"`
	Description         *string   `json:"description,omitempty"`
	AppliesTo           string    `json:"applies_to"`
	DefaultValidityDays *int      `json:"default_validity_days,omitempty"`
	IsRequiredDefault   bool      `json:"is_required_default"`
}

// ReviewDocumentRequest represents the JSON body for POST /api/v1/documents/{id}/review.
type ReviewDocumentRequest struct {
	Action          string `json:"action"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// DocumentResponse represents document details returned in API responses.
type DocumentResponse struct {
	ID               uuid.UUID             `json:"id"`
	UserID           *uuid.UUID            `json:"user_id,omitempty"`
	ContractorID     *uuid.UUID            `json:"contractor_id,omitempty"`
	DocumentTypeID   uuid.UUID             `json:"document_type_id"`
	DocumentTypeCode string                `json:"document_type_code,omitempty"`
	DocumentTypeName string                `json:"document_type_name,omitempty"`
	OwnerName        string                `json:"owner_name,omitempty"`
	OwnerType        string                `json:"owner_type,omitempty"`
	ContractorName   string                `json:"contractor_name,omitempty"`
	ReviewedByName   string                `json:"reviewed_by_name,omitempty"`
	ReviewedAt       *string               `json:"reviewed_at,omitempty"`
	FileName         string                `json:"file_name"`
	FileSizeBytes    *int                  `json:"file_size_bytes,omitempty"`
	MimeType         *string               `json:"mime_type,omitempty"`
	IssuedAt         *string               `json:"issued_at,omitempty"`
	ExpiresAt        *string               `json:"expires_at,omitempty"`
	Status           entity.DocumentStatus `json:"status"`
	RejectionReason  *string               `json:"rejection_reason,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	DownloadURL      string                `json:"download_url"`
}

// FormatDate converts *time.Time to a "YYYY-MM-DD" string pointer.
func FormatDate(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// ToUploadDocumentResponse maps a domain Document entity to UploadDocumentResponse.
func ToUploadDocumentResponse(d *entity.Document) UploadDocumentResponse {
	return UploadDocumentResponse{
		ID:               d.ID,
		UserID:           d.UserID,
		ContractorID:     d.ContractorID,
		DocumentTypeID:   d.DocumentTypeID,
		DocumentTypeCode: d.DocumentTypeCode,
		DocumentTypeName: d.DocumentTypeName,
		FileName:         d.FileName,
		FileSizeBytes:    d.FileSizeBytes,
		MimeType:         d.MimeType,
		IssuedAt:         FormatDate(d.IssuedAt),
		ExpiresAt:        FormatDate(d.ExpiresAt),
		Status:           string(d.Status),
		CreatedAt:        d.CreatedAt,
	}
}

// ToMyDocumentItemResponse maps a domain Document entity to MyDocumentItemResponse.
func ToMyDocumentItemResponse(d entity.Document) MyDocumentItemResponse {
	return MyDocumentItemResponse{
		ID:               d.ID,
		DocumentTypeCode: d.DocumentTypeCode,
		DocumentTypeName: d.DocumentTypeName,
		FileName:         d.FileName,
		Status:           string(d.Status),
		IssuedAt:         FormatDate(d.IssuedAt),
		ExpiresAt:        FormatDate(d.ExpiresAt),
		RejectionReason:  d.RejectionReason,
		CreatedAt:        d.CreatedAt,
	}
}

// ToMyDocumentItemResponses converts a slice of Document entities into MyDocumentItemResponse.
func ToMyDocumentItemResponses(docs []entity.Document) []MyDocumentItemResponse {
	if len(docs) == 0 {
		return []MyDocumentItemResponse{}
	}
	res := make([]MyDocumentItemResponse, len(docs))
	for i, d := range docs {
		res[i] = ToMyDocumentItemResponse(d)
	}
	return res
}

// ToDocumentResponse converts an entity.Document to DocumentResponse.
func ToDocumentResponse(doc *entity.Document) DocumentResponse {
	var issuedStr, expiresStr *string
	if doc.IssuedAt != nil {
		s := doc.IssuedAt.Format("2006-01-02")
		issuedStr = &s
	}
	if doc.ExpiresAt != nil {
		s := doc.ExpiresAt.Format("2006-01-02")
		expiresStr = &s
	}

	var reviewedStr *string
	if doc.ReviewedAt != nil {
		s := doc.ReviewedAt.Format(time.RFC3339)
		reviewedStr = &s
	}

	return DocumentResponse{
		ID:               doc.ID,
		UserID:           doc.UserID,
		ContractorID:     doc.ContractorID,
		DocumentTypeID:   doc.DocumentTypeID,
		DocumentTypeCode: doc.DocumentTypeCode,
		DocumentTypeName: doc.DocumentTypeName,
		OwnerName:        doc.OwnerName,
		OwnerType:        doc.OwnerType,
		ContractorName:   doc.ContractorName,
		ReviewedByName:   doc.ReviewedByName,
		ReviewedAt:       reviewedStr,
		FileName:         doc.FileName,
		FileSizeBytes:    doc.FileSizeBytes,
		MimeType:         doc.MimeType,
		IssuedAt:         issuedStr,
		ExpiresAt:        expiresStr,
		Status:           doc.Status,
		RejectionReason:  doc.RejectionReason,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
		DownloadURL:      "/api/v1/documents/" + doc.ID.String() + "/download",
	}
}

// ToDocumentTypeResponse converts an entity.DocumentType to DocumentTypeResponse.
func ToDocumentTypeResponse(dt *entity.DocumentType) DocumentTypeResponse {
	return DocumentTypeResponse{
		ID:                  dt.ID,
		Code:                dt.Code,
		Name:                dt.Name,
		Description:         dt.Description,
		AppliesTo:           dt.AppliesTo,
		DefaultValidityDays: dt.DefaultValidityDays,
		IsRequiredDefault:   dt.IsRequiredDefault,
	}
}
