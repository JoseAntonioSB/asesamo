package document

import (
	"context"
	"strings"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ReviewDocumentInput specifies parameters for approving or rejecting a document.
type ReviewDocumentInput struct {
	DocumentID      uuid.UUID
	Action          string // "approve" | "reject"
	RejectionReason string // Mandatory if Action == "reject"
}

// ReviewDocumentUseCase handles EHS compliance review of uploaded documents.
type ReviewDocumentUseCase struct {
	docRepo repository.DocumentRepository
}

// NewReviewDocumentUseCase constructs a new ReviewDocumentUseCase.
func NewReviewDocumentUseCase(docRepo repository.DocumentRepository) *ReviewDocumentUseCase {
	return &ReviewDocumentUseCase{docRepo: docRepo}
}

// Execute performs validation and updates the review status of the document.
func (uc *ReviewDocumentUseCase) Execute(ctx context.Context, input ReviewDocumentInput) (*entity.Document, error) {
	reviewerID, err := tenant.UserIDFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}
	role, err := tenant.RoleFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}

	// 1. Role validation: only EHS managers, plant admins, and super admins can approve/reject
	switch role {
	case "ehs_manager", "plant_admin", "super_admin":
		// Allowed
	default:
		return nil, domainErrors.ErrForbidden.WithMessage("Solo personal de Seguridad e Higiene (EHS) o administradores pueden revisar documentos.")
	}

	// 2. Action validation
	action := strings.ToLower(strings.TrimSpace(input.Action))
	var targetStatus entity.DocumentStatus
	var reasonPtr *string

	switch action {
	case "approve":
		targetStatus = entity.DocumentStatusApproved
		reasonPtr = nil

	case "reject":
		targetStatus = entity.DocumentStatusRejected
		reason := strings.TrimSpace(input.RejectionReason)
		if reason == "" {
			return nil, domainErrors.ErrInvalidInput.WithMessage("Debe especificar el motivo del rechazo del documento.")
		}
		reasonPtr = &reason

	default:
		return nil, domainErrors.ErrInvalidInput.WithMessage("Acción no válida. Utilice 'approve' o 'reject'.")
	}

	// 3. Verify document exists
	existingDoc, err := uc.docRepo.GetByID(ctx, input.DocumentID)
	if err != nil {
		return nil, err
	}

	// 4. Update status in repository
	updatedDoc, err := uc.docRepo.UpdateReviewStatus(ctx, existingDoc.ID, targetStatus, reviewerID, reasonPtr)
	if err != nil {
		return nil, err
	}

	return updatedDoc, nil
}
