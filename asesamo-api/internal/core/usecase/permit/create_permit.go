package permit

import (
	"context"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// CreatePermitInput specifies the payload for issuing a new high-risk work permit draft.
type CreatePermitInput struct {
	PlantID           uuid.UUID
	ContractorID      uuid.UUID
	PermitType        entity.PermitType
	WorkArea          string
	Description       string
	HazardDescription *string
	StartTime         time.Time
	EndTime           time.Time
	CreatedByUserID   uuid.UUID
	RequesterNotes    *string
}

// CreatePermitUseCase handles creating a new permit draft, assigning sequential number,
// seeding default STPS NOM checklist items, and recording requester digital signature.
type CreatePermitUseCase struct {
	permitRepo repository.WorkPermitRepository
}

// NewCreatePermitUseCase constructs a new CreatePermitUseCase.
func NewCreatePermitUseCase(permitRepo repository.WorkPermitRepository) *CreatePermitUseCase {
	return &CreatePermitUseCase{permitRepo: permitRepo}
}

// Execute performs validation and creates the draft work permit.
func (uc *CreatePermitUseCase) Execute(ctx context.Context, input CreatePermitInput) (*entity.WorkPermit, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	// 1. Validate inputs
	if input.PlantID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID de la planta es obligatorio.")
	}
	if input.ContractorID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del contratista es obligatorio.")
	}
	if !input.PermitType.IsValid() {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El tipo de permiso de alto riesgo es inválido.")
	}
	if strings.TrimSpace(input.WorkArea) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El área o ubicación del trabajo es obligatoria.")
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("La descripción de las actividades es obligatoria.")
	}
	if input.StartTime.IsZero() || input.EndTime.IsZero() {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Las fechas de inicio y fin son obligatorias.")
	}
	if !input.StartTime.Before(input.EndTime) {
		return nil, domainErrors.ErrPermitTimeInvalid
	}

	userID := input.CreatedByUserID
	if userID == uuid.Nil {
		userID, _ = tenant.UserIDFromContext(ctx)
	}

	// 2. Generate sequential permit number for the starting year
	year := input.StartTime.Year()
	if year == 0 {
		year = time.Now().UTC().Year()
	}
	permitNumber, err := uc.permitRepo.NextPermitNumber(ctx, tenantID, year)
	if err != nil {
		return nil, err
	}

	// 3. Persist WorkPermit in draft status
	permit := &entity.WorkPermit{
		ID:                uuid.New(),
		TenantID:          tenantID,
		PlantID:           input.PlantID,
		PermitNumber:      permitNumber,
		PermitType:        input.PermitType,
		Status:            entity.PermitStatusDraft,
		ContractorID:      input.ContractorID,
		WorkArea:          strings.TrimSpace(input.WorkArea),
		Description:       strings.TrimSpace(input.Description),
		HazardDescription: input.HazardDescription,
		StartTime:         input.StartTime,
		EndTime:           input.EndTime,
		CreatedByUserID:   userID,
	}

	if err := uc.permitRepo.Create(ctx, permit); err != nil {
		return nil, err
	}

	// 4. Seed default regulatory checklist items
	defaultItems := entity.DefaultChecklistForType(permit.PermitType)
	if len(defaultItems) > 0 {
		if err := uc.permitRepo.InitChecklistItems(ctx, permit.ID, defaultItems); err != nil {
			return nil, err
		}
	}

	// 5. Add initial requester digital signature
	if userID != uuid.Nil {
		sig := &entity.PermitSignature{
			ID:       uuid.New(),
			TenantID: tenantID,
			PermitID: permit.ID,
			UserID:   userID,
			Role:     entity.PermitSignerRequester,
			Notes:    input.RequesterNotes,
			SignedAt: time.Now().UTC(),
		}
		_ = uc.permitRepo.AddSignature(ctx, sig)
	}

	// 6. Return enriched created permit
	return uc.permitRepo.GetByID(ctx, permit.ID)
}
