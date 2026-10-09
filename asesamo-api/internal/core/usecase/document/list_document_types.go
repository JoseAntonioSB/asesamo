package document

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
)

// ListDocumentTypesUseCase retrieves all required document types.
type ListDocumentTypesUseCase struct {
	docRepo repository.DocumentRepository
}

// NewListDocumentTypesUseCase constructs a new ListDocumentTypesUseCase.
func NewListDocumentTypesUseCase(docRepo repository.DocumentRepository) *ListDocumentTypesUseCase {
	return &ListDocumentTypesUseCase{docRepo: docRepo}
}

// Execute returns the active document types catalog.
func (uc *ListDocumentTypesUseCase) Execute(ctx context.Context) ([]entity.DocumentType, error) {
	return uc.docRepo.ListRequiredDocumentTypes(ctx)
}
