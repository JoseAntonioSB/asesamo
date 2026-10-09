package document

import (
	"context"
	"io"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// DownloadDocumentUseCase handles authorized document retrieval and stream reading.
type DownloadDocumentUseCase struct {
	docRepo        repository.DocumentRepository
	storageService repository.StorageService
}

// NewDownloadDocumentUseCase constructs a new DownloadDocumentUseCase.
func NewDownloadDocumentUseCase(
	docRepo repository.DocumentRepository,
	storageService repository.StorageService,
) *DownloadDocumentUseCase {
	return &DownloadDocumentUseCase{
		docRepo:        docRepo,
		storageService: storageService,
	}
}

// Execute retrieves document metadata and opens the file stream.
func (uc *DownloadDocumentUseCase) Execute(ctx context.Context, docID uuid.UUID) (io.ReadCloser, *entity.Document, error) {
	authUserID, err := tenant.UserIDFromContext(ctx)
	if err != nil {
		return nil, nil, domainErrors.ErrUnauthorized
	}
	role, err := tenant.RoleFromContext(ctx)
	if err != nil {
		return nil, nil, domainErrors.ErrUnauthorized
	}

	doc, err := uc.docRepo.GetByID(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	// Authorization check
	isAuthorized := false

	switch role {
	case "worker":
		if doc.UserID != nil && *doc.UserID == authUserID {
			isAuthorized = true
		}
	case "contractor_admin":
		contractor, err := uc.docRepo.GetContractorByAdminUserID(ctx, authUserID)
		if err == nil && contractor != nil {
			if doc.ContractorID != nil && *doc.ContractorID == contractor.ID {
				isAuthorized = true
			} else if doc.UserID != nil {
				// Check if worker belongs to this contractor
				workerContractor, err := uc.docRepo.GetContractorByWorkerUserID(ctx, *doc.UserID)
				if err == nil && workerContractor != nil && workerContractor.ID == contractor.ID {
					isAuthorized = true
				}
			}
		}
	case "plant_admin", "super_admin", "ehs_manager", "guard":
		isAuthorized = true
	}

	if !isAuthorized {
		return nil, nil, domainErrors.ErrForbidden.WithMessage("No tiene autorización para descargar este documento.")
	}

	stream, err := uc.storageService.Download(ctx, doc.R2Key)
	if err != nil {
		return nil, nil, err
	}

	return stream, doc, nil
}
