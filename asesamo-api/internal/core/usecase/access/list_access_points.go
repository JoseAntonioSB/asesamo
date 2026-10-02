// Package access implements access point and gate check-in use cases.
package access

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ListAccessPointsUseCase handles listing active access points for a tenant.
type ListAccessPointsUseCase struct {
	accessPointRepo repository.AccessPointRepository
}

// NewListAccessPointsUseCase instantiates a ListAccessPointsUseCase.
func NewListAccessPointsUseCase(accessPointRepo repository.AccessPointRepository) *ListAccessPointsUseCase {
	return &ListAccessPointsUseCase{
		accessPointRepo: accessPointRepo,
	}
}

// Execute retrieves all active access points for the specified tenant.
// Guarantees returning a non-nil slice (empty slice if no access points found).
func (uc *ListAccessPointsUseCase) Execute(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	if tenantID == uuid.Nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	points, err := uc.accessPointRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, domainErrors.ErrDatabase.WithErr(err)
	}

	if points == nil {
		points = []entity.AccessPoint{}
	}

	return points, nil
}
