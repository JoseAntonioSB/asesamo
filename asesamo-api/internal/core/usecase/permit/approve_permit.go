package permit

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ApprovePermitInput specifies the payload for EHS digital authorization.
type ApprovePermitInput struct {
	PermitID      uuid.UUID
	Notes         *string
	SignatureData *string
	IPAddress     *string
	UserAgent     *string
}

// ApprovePermitUseCase performs the EHS authorization gate:
// 1. Verifies caller role is EHS Manager, Plant Admin, or Super Admin.
// 2. Verifies permit has at least 1 assigned worker.
// 3. Verifies ALL mandatory checklist items are verified.
// 4. Re-evaluates Candado Inteligente compliance for all crew members.
// 5. Transitions status to approved_active and logs authorizer digital signature.
type ApprovePermitUseCase struct {
	permitRepo       repository.WorkPermitRepository
	complianceEngine *service.ComplianceEngine
}

// NewApprovePermitUseCase constructs a new ApprovePermitUseCase.
func NewApprovePermitUseCase(
	permitRepo repository.WorkPermitRepository,
	complianceEngine *service.ComplianceEngine,
) *ApprovePermitUseCase {
	return &ApprovePermitUseCase{
		permitRepo:       permitRepo,
		complianceEngine: complianceEngine,
	}
}

// Execute enforces the EHS authorization gate.
func (uc *ApprovePermitUseCase) Execute(ctx context.Context, input ApprovePermitInput) (*entity.WorkPermit, error) {
	if input.PermitID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del permiso es obligatorio.")
	}

	// 1. Verify caller role
	role, err := tenant.RoleFromContext(ctx)
	if err != nil || (role != "ehs_manager" && role != "plant_admin" && role != "super_admin" && role != "tenant_admin") {
		return nil, domainErrors.ErrForbidden.WithMessage("Solo personal con rol EHS o Administrador de Planta puede autorizar permisos de trabajo.")
	}

	authorizerID, err := tenant.UserIDFromContext(ctx)
	if err != nil || authorizerID == uuid.Nil {
		return nil, domainErrors.ErrUnauthorized
	}

	// 2. Load permit
	permit, err := uc.permitRepo.GetByID(ctx, input.PermitID)
	if err != nil {
		return nil, err
	}

	// 3. Verify status allows authorization
	if permit.Status != entity.PermitStatusPendingApproval && permit.Status != entity.PermitStatusDraft {
		return nil, domainErrors.ErrPermitInvalidState.WithMessage("El permiso ya ha sido procesado o se encuentra en un estado que no permite autorización.")
	}

	// 4. Crew gate: must have at least 1 assigned worker
	if len(permit.Workers) == 0 {
		return nil, domainErrors.ErrPermitNoWorkers
	}

	// 5. Checklist gate: all mandatory items must be verified
	for _, item := range permit.Checklist {
		if item.IsMandatory && !item.IsVerified {
			return nil, domainErrors.ErrChecklistIncomplete
		}
	}

	// 6. Candado Inteligente re-evaluation gate for all crew members
	for _, worker := range permit.Workers {
		snapshot, err := uc.complianceEngine.EvaluateWorker(ctx, permit.PlantID, permit.ContractorID, permit.PermitType, worker.WorkerID)
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

		// Update point-in-time compliance snapshot
		worker.ComplianceSnapshot = *snapshot
		worker.ComplianceStatus = "compliant"
		_ = uc.permitRepo.AssignWorker(ctx, &worker)
	}

	// 7. Transition status to approved_active
	if err := uc.permitRepo.UpdateStatus(ctx, permit.ID, entity.PermitStatusApprovedActive, input.Notes, &authorizerID); err != nil {
		return nil, err
	}

	// 8. Add EHS authorizer digital signature
	sig := &entity.PermitSignature{
		ID:            uuid.New(),
		TenantID:      permit.TenantID,
		PermitID:      permit.ID,
		UserID:        authorizerID,
		Role:          entity.PermitSignerEHSAuthorizer,
		SignatureData: input.SignatureData,
		IPAddress:     input.IPAddress,
		UserAgent:     input.UserAgent,
		Notes:         input.Notes,
		SignedAt:      time.Now().UTC(),
	}
	_ = uc.permitRepo.AddSignature(ctx, sig)

	return uc.permitRepo.GetByID(ctx, permit.ID)
}
