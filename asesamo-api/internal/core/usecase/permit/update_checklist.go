package permit

import (
	"context"
	"strings"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ChecklistItemUpdate specifies an update to an individual safety verification item.
type ChecklistItemUpdate struct {
	ItemKey      string         `json:"item_key"`
	ItemCode     string         `json:"item_code,omitempty"` // Alias for ItemKey
	IsVerified   bool           `json:"is_verified"`
	IsChecked    bool           `json:"is_checked,omitempty"` // Alias for IsVerified
	Notes        *string        `json:"notes,omitempty"`
	Measurements map[string]any `json:"measurements,omitempty"`
}

// UpdateChecklistInput contains the checklist updates to apply to a permit.
type UpdateChecklistInput struct {
	PermitID uuid.UUID
	Items    []ChecklistItemUpdate
}

// UpdateChecklistUseCase updates verification flags, observations, and safety measurements.
type UpdateChecklistUseCase struct {
	permitRepo repository.WorkPermitRepository
}

// NewUpdateChecklistUseCase constructs a new UpdateChecklistUseCase.
func NewUpdateChecklistUseCase(permitRepo repository.WorkPermitRepository) *UpdateChecklistUseCase {
	return &UpdateChecklistUseCase{permitRepo: permitRepo}
}

// Execute applies checklist verifications and updates.
func (uc *UpdateChecklistUseCase) Execute(ctx context.Context, input UpdateChecklistInput) (*entity.WorkPermit, error) {
	if input.PermitID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del permiso es obligatorio.")
	}
	if len(input.Items) == 0 {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Debe enviar al menos un elemento para actualizar.")
	}

	permit, err := uc.permitRepo.GetByID(ctx, input.PermitID)
	if err != nil {
		return nil, err
	}

	// Permits cannot be modified if already closed or cancelled
	if permit.Status.IsTerminal() {
		return nil, domainErrors.ErrPermitInvalidState.WithMessage("No se puede modificar la lista de verificación de un permiso concluido o cancelado.")
	}

	userID, _ := tenant.UserIDFromContext(ctx)
	var verifiedByPtr *uuid.UUID
	if userID != uuid.Nil {
		verifiedByPtr = &userID
	}

	for _, item := range input.Items {
		key := strings.TrimSpace(item.ItemKey)
		if key == "" {
			key = strings.TrimSpace(item.ItemCode)
		}
		if key == "" {
			continue
		}

		verified := item.IsVerified || item.IsChecked
		if err := uc.permitRepo.UpdateChecklistItem(ctx, permit.ID, key, verified, verifiedByPtr, item.Notes, item.Measurements); err != nil {
			return nil, err
		}
	}

	return uc.permitRepo.GetByID(ctx, permit.ID)
}
