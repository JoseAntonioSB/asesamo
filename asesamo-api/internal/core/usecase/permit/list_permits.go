package permit

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ListPermitsUseCase retrieves a paginated list of permits with optional status, contractor,
// plant, and date filters, automatically scoped by role.
type ListPermitsUseCase struct {
	permitRepo     repository.WorkPermitRepository
	contractorRepo repository.ContractorRepository
}

// NewListPermitsUseCase constructs a new ListPermitsUseCase.
func NewListPermitsUseCase(permitRepo repository.WorkPermitRepository, contractorRepo repository.ContractorRepository) *ListPermitsUseCase {
	return &ListPermitsUseCase{
		permitRepo:     permitRepo,
		contractorRepo: contractorRepo,
	}
}

// Execute applies role restrictions and returns the filtered permits.
func (uc *ListPermitsUseCase) Execute(ctx context.Context, filter repository.PermitFilter) ([]entity.WorkPermit, int, error) {
	role, _ := tenant.RoleFromContext(ctx)
	userID, _ := tenant.UserIDFromContext(ctx)

	// If the user is a contractor admin, scope results to their own contractor
	if role == "contractor_admin" && userID != uuid.Nil && uc.contractorRepo != nil {
		c, err := uc.contractorRepo.GetContractorByAdminUserID(ctx, userID)
		if err == nil && c != nil {
			filter.ContractorID = &c.ID
		}
	}

	return uc.permitRepo.ListWithFilters(ctx, filter)
}
