package user

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ProfileDetail contains comprehensive user details including corporate affiliations.
type ProfileDetail struct {
	User           *entity.User
	ContractorID   *uuid.UUID
	ContractorName *string
	PlantID        *uuid.UUID
	PlantName      *string
}

// GetProfileUseCase retrieves full user profile and associated company information.
type GetProfileUseCase struct {
	userRepo       repository.UserRepository
	contractorRepo repository.ContractorRepository
}

// NewGetProfileUseCase constructs GetProfileUseCase.
func NewGetProfileUseCase(
	userRepo repository.UserRepository,
	contractorRepo repository.ContractorRepository,
) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo:       userRepo,
		contractorRepo: contractorRepo,
	}
}

// Execute retrieves the profile for the target user.
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*ProfileDetail, error) {
	if userID == uuid.Nil {
		return nil, domainErrors.ErrUnauthorized
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	detail := &ProfileDetail{
		User: u,
	}

	// If user is contractor admin, fetch their company
	if u.Role == entity.RoleContractorAdmin && uc.contractorRepo != nil {
		c, err := uc.contractorRepo.GetContractorByAdminUserID(ctx, u.ID)
		if err == nil && c != nil {
			detail.ContractorID = &c.ID
			detail.ContractorName = &c.Name
		}
	}

	return detail, nil
}
