package evacuation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type StartEvacuationInput struct {
	Type        entity.EvacuationType `json:"type"`
	Description string                `json:"description,omitempty"`
}

type StartEvacuationUseCase struct {
	repo repository.EmergencyEvacuationRepository
}

func NewStartEvacuationUseCase(repo repository.EmergencyEvacuationRepository) *StartEvacuationUseCase {
	return &StartEvacuationUseCase{repo: repo}
}

func (uc *StartEvacuationUseCase) Execute(ctx context.Context, tenantID, initiatorID uuid.UUID, input StartEvacuationInput) (*entity.EmergencyEvacuation, error) {
	active, err := uc.repo.GetActiveEvacuation(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, errors.New("ya existe un evento de evacuación activo en planta")
	}

	presentWorkers, err := uc.repo.GetPresentWorkersInPlant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	evacNumber := fmt.Sprintf("EVAC-%s-%s", time.Now().Format("2006"), uuid.New().String()[:6])
	evac := entity.NewEmergencyEvacuation(tenantID, evacNumber, input.Type, input.Description, initiatorID)

	if err := evac.InitializeRollCall(presentWorkers); err != nil {
		return nil, err
	}

	if err := uc.repo.StartEvacuation(ctx, evac); err != nil {
		return nil, err
	}

	return evac, nil
}
