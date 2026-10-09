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

// ListPendingDocumentsUseCase handles querying documents pending or undergoing EHS review.
type ListPendingDocumentsUseCase struct {
	docRepo repository.DocumentRepository
}

// NewListPendingDocumentsUseCase constructs a new ListPendingDocumentsUseCase.
func NewListPendingDocumentsUseCase(docRepo repository.DocumentRepository) *ListPendingDocumentsUseCase {
	return &ListPendingDocumentsUseCase{docRepo: docRepo}
}

// Execute retrieves documents matching status filter for the tenant.
func (uc *ListPendingDocumentsUseCase) Execute(ctx context.Context, rawStatus string) ([]entity.Document, error) {
	role, err := tenant.RoleFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}

	switch role {
	case "ehs_manager", "plant_admin", "super_admin", "guard":
		// Authorized to view tray
	default:
		return nil, domainErrors.ErrForbidden.WithMessage("No tiene permisos para ver la bandeja de aprobación documental.")
	}

	// Resolve tenant if not super_admin
	var tenantPtr *uuid.UUID
	tenantID, err := tenant.FromContext(ctx)
	if err == nil && tenantID != uuid.Nil && role != "super_admin" {
		tenantPtr = &tenantID
	}

	var statusFilter *entity.DocumentStatus
	cleanStatus := strings.ToLower(strings.TrimSpace(rawStatus))
	if cleanStatus != "" && cleanStatus != "all" {
		parsedStatus := entity.DocumentStatus(cleanStatus)
		statusFilter = &parsedStatus
	}

	return uc.docRepo.ListPendingReview(ctx, tenantPtr, statusFilter)
}
