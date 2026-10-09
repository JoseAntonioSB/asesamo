package contractor

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ListPlantContractorsUseCase lists all contractor companies affiliated with a plant.
type ListPlantContractorsUseCase struct {
	contractorRepo repository.ContractorRepository
	plantRepo      repository.PlantRepository
}

// NewListPlantContractorsUseCase constructs ListPlantContractorsUseCase.
func NewListPlantContractorsUseCase(
	contractorRepo repository.ContractorRepository,
	plantRepo repository.PlantRepository,
) *ListPlantContractorsUseCase {
	return &ListPlantContractorsUseCase{
		contractorRepo: contractorRepo,
		plantRepo:      plantRepo,
	}
}

// Execute returns all contractors affiliated with the given plant.
func (uc *ListPlantContractorsUseCase) Execute(ctx context.Context, plantID uuid.UUID) ([]repository.ContractorWithDetails, error) {
	if plantID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador de planta es obligatorio.")
	}

	return uc.contractorRepo.ListByPlant(ctx, plantID)
}

// UpdateContractorStatusUseCase allows plant management to change a contractor's status (active, suspended, blocked).
type UpdateContractorStatusUseCase struct {
	contractorRepo repository.ContractorRepository
}

// NewUpdateContractorStatusUseCase constructs UpdateContractorStatusUseCase.
func NewUpdateContractorStatusUseCase(contractorRepo repository.ContractorRepository) *UpdateContractorStatusUseCase {
	return &UpdateContractorStatusUseCase{contractorRepo: contractorRepo}
}

// Execute changes the contractor's status in a plant.
func (uc *UpdateContractorStatusUseCase) Execute(ctx context.Context, plantID, contractorID uuid.UUID, statusStr string) error {
	if plantID == uuid.Nil || contractorID == uuid.Nil {
		return domainErrors.ErrInvalidInput.WithMessage("La planta y el contratista son obligatorios.")
	}

	var status entity.PlantContractorStatus
	switch statusStr {
	case "active":
		status = entity.PlantContractorActive
	case "pending":
		status = entity.PlantContractorPending
	case "suspended":
		status = entity.PlantContractorSuspended
	case "blocked":
		status = entity.PlantContractorBlocked
	default:
		return domainErrors.ErrInvalidInput.WithMessage("Estatus inválido. Valores permitidos: active, pending, suspended, blocked.")
	}

	return uc.contractorRepo.UpdatePlantContractorStatus(ctx, plantID, contractorID, status)
}
