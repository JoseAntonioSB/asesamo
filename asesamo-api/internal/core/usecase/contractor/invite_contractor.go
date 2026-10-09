package contractor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// InviteContractorInput holds the corporate and admin details to invite/register a contractor company.
type InviteContractorInput struct {
	TenantID       uuid.UUID
	PlantID        uuid.UUID
	InvitedByID    *uuid.UUID
	Name           string
	RFC            string
	REPSENumber    string
	ContractorType string // "company" | "individual"
	AdminEmail     string
	AdminFullName  string
	AdminPhone     string
	InitialStatus  string // "active" | "pending"
}

// InviteContractorUseCase handles onboarding of contractor companies and their administrative users.
type InviteContractorUseCase struct {
	contractorRepo repository.ContractorRepository
	userRepo       repository.UserRepository
	plantRepo      repository.PlantRepository
}

// NewInviteContractorUseCase constructs InviteContractorUseCase.
func NewInviteContractorUseCase(
	contractorRepo repository.ContractorRepository,
	userRepo repository.UserRepository,
	plantRepo repository.PlantRepository,
) *InviteContractorUseCase {
	return &InviteContractorUseCase{
		contractorRepo: contractorRepo,
		userRepo:       userRepo,
		plantRepo:      plantRepo,
	}
}

// Execute registers or links a contractor company, creates the contractor_admin if new, and associates with the plant.
func (uc *InviteContractorUseCase) Execute(ctx context.Context, in InviteContractorInput) (*entity.Contractor, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("La razón social de la empresa contratista es obligatoria.")
	}

	rfc := strings.TrimSpace(strings.ToUpper(in.RFC))
	if rfc == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El RFC de la empresa contratista es obligatorio.")
	}

	adminEmail := strings.TrimSpace(strings.ToLower(in.AdminEmail))
	if adminEmail == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El correo electrónico del Administrador Contratista es obligatorio.")
	}

	adminFullName := strings.TrimSpace(in.AdminFullName)
	if adminFullName == "" {
		adminFullName = "Admin " + name
	}

	if in.PlantID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador de la planta es obligatorio.")
	}

	// Verify plant exists
	_, err := uc.plantRepo.GetByID(ctx, in.PlantID)
	if err != nil {
		return nil, domainErrors.ErrPlantNotFound.WithMessage("La planta especificada no existe.")
	}

	// 1. Resolve or create contractor_admin user
	var adminUserID uuid.UUID
	existingUser, err := uc.userRepo.GetByEmail(ctx, adminEmail)
	if err == nil && existingUser != nil {
		adminUserID = existingUser.ID
	} else {
		// Create new user with default password 'Password123!'
		hashBytes, err := bcrypt.GenerateFromPassword([]byte("Password123!"), 10)
		if err != nil {
			return nil, fmt.Errorf("usecase/contractor: failed to hash password: %w", err)
		}

		adminUserID = uuid.New()
		newUser := &entity.User{
			ID:           adminUserID,
			TenantID:     nil, // global
			Email:        adminEmail,
			FullName:     adminFullName,
			Phone:        &in.AdminPhone,
			Role:         entity.RoleContractorAdmin,
			IsActive:     true,
			PasswordHash: string(hashBytes),
		}
		if err := uc.contractorRepo.CreateUserWithPassword(ctx, newUser, string(hashBytes)); err != nil {
			return nil, fmt.Errorf("usecase/contractor: failed to create admin user: %w", err)
		}
	}

	// 2. Resolve or create contractor company by RFC
	cType := entity.ContractorTypeCompany
	if in.ContractorType == "individual" {
		cType = entity.ContractorTypeIndividual
	}

	var repsePtr *string
	if repse := strings.TrimSpace(in.REPSENumber); repse != "" {
		repsePtr = &repse
	}

	contractor, err := uc.contractorRepo.GetContractorByRFC(ctx, rfc)
	if err != nil && domainErrors.IsNotFound(err) {
		// Create new contractor
		contractor = &entity.Contractor{
			ID:             uuid.New(),
			Name:           name,
			RFC:            &rfc,
			REPSENumber:    repsePtr,
			ContractorType: cType,
			AdminUserID:    &adminUserID,
			IsActive:       true,
		}
		if err := uc.contractorRepo.CreateContractor(ctx, contractor); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// 3. Associate with plant in plant_contractors
	status := entity.PlantContractorActive
	if in.InitialStatus == "pending" {
		status = entity.PlantContractorPending
	}

	now := time.Now().UTC()
	var activatedAt *time.Time
	if status == entity.PlantContractorActive {
		activatedAt = &now
	}

	pc := &entity.PlantContractor{
		ID:           uuid.New(),
		PlantID:      in.PlantID,
		TenantID:     in.TenantID,
		ContractorID: contractor.ID,
		Status:       status,
		InvitedBy:    in.InvitedByID,
		InvitedAt:    now,
		ActivatedAt:  activatedAt,
	}

	if err := uc.contractorRepo.CreatePlantContractor(ctx, pc); err != nil {
		return nil, err
	}

	return contractor, nil
}
