package permit

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// GetPermitUseCase retrieves a single work permit with its crew, checklist items, and signatures.
type GetPermitUseCase struct {
	permitRepo     repository.WorkPermitRepository
	contractorRepo repository.ContractorRepository
}

// NewGetPermitUseCase constructs a new GetPermitUseCase.
func NewGetPermitUseCase(permitRepo repository.WorkPermitRepository, contractorRepo repository.ContractorRepository) *GetPermitUseCase {
	return &GetPermitUseCase{
		permitRepo:     permitRepo,
		contractorRepo: contractorRepo,
	}
}

// Execute retrieves and validates access to the requested permit.
func (uc *GetPermitUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.WorkPermit, error) {
	if id == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del permiso es obligatorio.")
	}

	permit, err := uc.permitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify contractor scoping if role is contractor_admin
	role, _ := tenant.RoleFromContext(ctx)
	userID, _ := tenant.UserIDFromContext(ctx)
	if role == "contractor_admin" && userID != uuid.Nil && uc.contractorRepo != nil {
		c, err := uc.contractorRepo.GetContractorByAdminUserID(ctx, userID)
		if err == nil && c != nil && permit.ContractorID != c.ID {
			return nil, domainErrors.ErrForbidden.WithMessage("No tiene autorización para visualizar este permiso de trabajo.")
		}
	}

	return permit, nil
}
