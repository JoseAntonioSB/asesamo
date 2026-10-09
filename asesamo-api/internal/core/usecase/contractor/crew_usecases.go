package contractor

import (
	"context"
	"fmt"
	"strings"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GetMyContractorCompanyUseCase returns the contractor entity for the logged-in contractor_admin.
type GetMyContractorCompanyUseCase struct {
	contractorRepo repository.ContractorRepository
}

func NewGetMyContractorCompanyUseCase(contractorRepo repository.ContractorRepository) *GetMyContractorCompanyUseCase {
	return &GetMyContractorCompanyUseCase{contractorRepo: contractorRepo}
}

func (uc *GetMyContractorCompanyUseCase) Execute(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	if adminUserID == uuid.Nil {
		return nil, domainErrors.ErrUnauthorized
	}

	return uc.contractorRepo.GetContractorByAdminUserID(ctx, adminUserID)
}

// ListContractorWorkersUseCase retrieves all workers in a contractor's crew.
type ListContractorWorkersUseCase struct {
	contractorRepo repository.ContractorRepository
}

func NewListContractorWorkersUseCase(contractorRepo repository.ContractorRepository) *ListContractorWorkersUseCase {
	return &ListContractorWorkersUseCase{contractorRepo: contractorRepo}
}

func (uc *ListContractorWorkersUseCase) Execute(ctx context.Context, contractorID uuid.UUID) ([]repository.ContractorWorkerDetail, error) {
	if contractorID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del contratista es obligatorio.")
	}

	return uc.contractorRepo.ListWorkersByContractor(ctx, contractorID)
}

// AddContractorWorkerInput defines inputs for registering a worker under a contractor company.
type AddContractorWorkerInput struct {
	ContractorID uuid.UUID
	FullName     string
	Email        string
	Phone        string
	JobTitle     string
	CURP         string
	IMSSNumber   string
}

// AddContractorWorkerUseCase registers an individual worker into the contractor company's crew.
type AddContractorWorkerUseCase struct {
	contractorRepo repository.ContractorRepository
	userRepo       repository.UserRepository
}

func NewAddContractorWorkerUseCase(
	contractorRepo repository.ContractorRepository,
	userRepo repository.UserRepository,
) *AddContractorWorkerUseCase {
	return &AddContractorWorkerUseCase{
		contractorRepo: contractorRepo,
		userRepo:       userRepo,
	}
}

func (uc *AddContractorWorkerUseCase) Execute(ctx context.Context, in AddContractorWorkerInput) (*repository.ContractorWorkerDetail, error) {
	fullName := strings.TrimSpace(in.FullName)
	if fullName == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El nombre completo del trabajador es obligatorio.")
	}

	email := strings.TrimSpace(strings.ToLower(in.Email))
	if email == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El correo electrónico es obligatorio.")
	}

	if in.ContractorID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("La empresa contratista es obligatoria.")
	}

	// Verify contractor exists
	_, err := uc.contractorRepo.GetContractorByID(ctx, in.ContractorID)
	if err != nil {
		return nil, err
	}

	// Check if user already exists
	existing, _ := uc.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Ya existe un usuario registrado con este correo electrónico.")
	}

	// Hash password for worker login
	hashBytes, err := bcrypt.GenerateFromPassword([]byte("Password123!"), 10)
	if err != nil {
		return nil, fmt.Errorf("usecase/contractor: failed to hash password: %w", err)
	}

	workerID := uuid.New()
	workerUser := &entity.User{
		ID:           workerID,
		TenantID:     nil, // workers are global to contractors
		Email:        email,
		FullName:     fullName,
		Phone:        &in.Phone,
		Role:         entity.RoleWorker,
		IsActive:     true,
		PasswordHash: string(hashBytes),
	}

	var curpPtr *string
	if curp := strings.TrimSpace(strings.ToUpper(in.CURP)); curp != "" {
		curpPtr = &curp
	}

	var imssPtr *string
	if imss := strings.TrimSpace(in.IMSSNumber); imss != "" {
		imssPtr = &imss
	}

	var jobPtr *string
	if job := strings.TrimSpace(in.JobTitle); job != "" {
		jobPtr = &job
	}

	cw := &entity.ContractorWorker{
		ID:           uuid.New(),
		ContractorID: in.ContractorID,
		UserID:       workerID,
		JobTitle:     jobPtr,
		CURP:         curpPtr,
		IMSSNumber:   imssPtr,
		IsActive:     true,
	}

	if err := uc.contractorRepo.AddWorker(ctx, workerUser, cw); err != nil {
		return nil, err
	}

	return &repository.ContractorWorkerDetail{
		WorkerID:       workerID,
		ContractorID:   in.ContractorID,
		FullName:       fullName,
		Email:          email,
		Phone:          &in.Phone,
		JobTitle:       jobPtr,
		CURP:           curpPtr,
		IMSSNumber:     imssPtr,
		IsActive:       true,
		UserIsActive:   true,
		DocumentsCount: 0,
	}, nil
}

// ToggleContractorWorkerStatusUseCase changes the active status of a worker in the crew.
type ToggleContractorWorkerStatusUseCase struct {
	contractorRepo repository.ContractorRepository
}

func NewToggleContractorWorkerStatusUseCase(contractorRepo repository.ContractorRepository) *ToggleContractorWorkerStatusUseCase {
	return &ToggleContractorWorkerStatusUseCase{contractorRepo: contractorRepo}
}

func (uc *ToggleContractorWorkerStatusUseCase) Execute(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error {
	if contractorID == uuid.Nil || workerUserID == uuid.Nil {
		return domainErrors.ErrInvalidInput.WithMessage("Contratista y trabajador son obligatorios.")
	}

	return uc.contractorRepo.ToggleWorkerStatus(ctx, contractorID, workerUserID, isActive)
}
