package permit

import (
	"context"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// AssignWorkersInput specifies worker assignments or removals for a permit.
type AssignWorkersInput struct {
	PermitID  uuid.UUID
	WorkerIDs []uuid.UUID
	Action    string // "assign" (default) or "remove"
}

// AssignWorkersUseCase executes worker assignment or removal while enforcing
// Candado Inteligente compliance locks on every attached worker.
type AssignWorkersUseCase struct {
	permitRepo       repository.WorkPermitRepository
	complianceEngine *service.ComplianceEngine
}

// NewAssignWorkersUseCase constructs a new AssignWorkersUseCase.
func NewAssignWorkersUseCase(
	permitRepo repository.WorkPermitRepository,
	complianceEngine *service.ComplianceEngine,
) *AssignWorkersUseCase {
	return &AssignWorkersUseCase{
		permitRepo:       permitRepo,
		complianceEngine: complianceEngine,
	}
}

// Execute validates permissions, executes Candado Inteligente compliance evaluation,
// and updates the crew on the permit.
func (uc *AssignWorkersUseCase) Execute(ctx context.Context, input AssignWorkersInput) (*entity.WorkPermit, error) {
	if input.PermitID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del permiso es obligatorio.")
	}
	if len(input.WorkerIDs) == 0 {
		return nil, domainErrors.ErrInvalidInput.WithMessage("Debe especificar al menos un ID de trabajador.")
	}

	permit, err := uc.permitRepo.GetByID(ctx, input.PermitID)
	if err != nil {
		return nil, err
	}

	// Only draft and pending_approval permits permit crew changes
	if permit.Status != entity.PermitStatusDraft && permit.Status != entity.PermitStatusPendingApproval {
		return nil, domainErrors.ErrPermitInvalidState.WithMessage("No se puede modificar la cuadrilla en el estatus actual del permiso.")
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action == "remove" {
		for _, workerID := range input.WorkerIDs {
			if workerID == uuid.Nil {
				continue
			}
			if err := uc.permitRepo.RemoveWorker(ctx, permit.ID, workerID); err != nil {
				return nil, err
			}
		}
		return uc.permitRepo.GetByID(ctx, permit.ID)
	}

	// Action is assign (or default)
	assignedBy, _ := tenant.UserIDFromContext(ctx)
	var assignedByPtr *uuid.UUID
	if assignedBy != uuid.Nil {
		assignedByPtr = &assignedBy
	}

	for _, workerID := range input.WorkerIDs {
		if workerID == uuid.Nil {
			continue
		}

		// Evaluate Candado Inteligente compliance rules
		snapshot, err := uc.complianceEngine.EvaluateWorker(ctx, permit.PlantID, permit.ContractorID, permit.PermitType, workerID)
		if err != nil {
			return nil, err
		}
		if !snapshot.IsCompliant {
			blockMsg := "El trabajador no cumple con los requisitos de seguridad obligatorios."
			if snapshot.BlockReason != nil && *snapshot.BlockReason != "" {
				blockMsg = *snapshot.BlockReason
			}
			return nil, domainErrors.ErrCandadoComplianceFailed.WithMessage(blockMsg)
		}

		pw := &entity.PermitWorker{
			ID:                 uuid.New(),
			TenantID:           permit.TenantID,
			PermitID:           permit.ID,
			WorkerID:           workerID,
			ComplianceStatus:   "compliant",
			ComplianceSnapshot: *snapshot,
			AssignedBy:         assignedByPtr,
			AssignedAt:         time.Now().UTC(),
		}

		if err := uc.permitRepo.AssignWorker(ctx, pw); err != nil {
			return nil, err
		}
	}

	return uc.permitRepo.GetByID(ctx, permit.ID)
}
