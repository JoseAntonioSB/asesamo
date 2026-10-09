package document

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ListMyDocumentsUseCase handles querying documents for the authenticated user/contractor.
type ListMyDocumentsUseCase struct {
	docRepo repository.DocumentRepository
}

// NewListMyDocumentsUseCase constructs a new ListMyDocumentsUseCase.
func NewListMyDocumentsUseCase(docRepo repository.DocumentRepository) *ListMyDocumentsUseCase {
	return &ListMyDocumentsUseCase{docRepo: docRepo}
}

// Execute queries documents based on authenticated user context.
func (uc *ListMyDocumentsUseCase) Execute(ctx context.Context, targetWorkerID *uuid.UUID, targetContractorID *uuid.UUID) ([]entity.Document, error) {
	authUserID, err := tenant.UserIDFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}
	role, err := tenant.RoleFromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrUnauthorized
	}

	switch role {
	case "worker":
		// Workers can only see their own documents
		return uc.docRepo.ListByWorker(ctx, authUserID)

	case "contractor_admin":
		contractor, err := uc.docRepo.GetContractorByAdminUserID(ctx, authUserID)
		if err != nil {
			return nil, err
		}
		if targetWorkerID != nil {
			// Ensure worker belongs to contractor
			workerContractor, err := uc.docRepo.GetContractorByWorkerUserID(ctx, *targetWorkerID)
			if err != nil {
				return nil, err
			}
			if workerContractor.ID != contractor.ID {
				return nil, domainErrors.ErrForbidden.WithMessage("El trabajador solicitado no pertenece a su empresa contratista.")
			}
			return uc.docRepo.ListByWorker(ctx, *targetWorkerID)
		}
		// Returns company documents
		return uc.docRepo.ListByContractor(ctx, contractor.ID)

	case "plant_admin", "super_admin", "ehs_manager", "guard":
		if targetWorkerID != nil {
			return uc.docRepo.ListByWorker(ctx, *targetWorkerID)
		}
		if targetContractorID != nil {
			return uc.docRepo.ListByContractor(ctx, *targetContractorID)
		}
		// Fallback to worker list if admin user is also associated with worker profile
		docs, err := uc.docRepo.ListByWorker(ctx, authUserID)
		if err == nil && len(docs) > 0 {
			return docs, nil
		}
		return []entity.Document{}, nil

	default:
		return nil, domainErrors.ErrForbidden
	}
}
