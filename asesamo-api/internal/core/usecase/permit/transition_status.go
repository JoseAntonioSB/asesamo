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

// TransitionStatusInput encapsulates state transition parameters.
type TransitionStatusInput struct {
	PermitID      uuid.UUID
	Action        string // "submit", "suspend", "resume", "close", "cancel"
	Reason        *string
	Notes         *string
	SignatureData *string
	IPAddress     *string
	UserAgent     *string
}

// TransitionStatusUseCase enforces the PTAR state machine lifecycle rules:
// - submit:  draft -> pending_approval (requires crew)
// - suspend: approved_active -> suspended (requires reason)
// - resume:  suspended -> approved_active
// - close:   approved_active -> closed
// - cancel:  any non-terminal -> cancelled (requires reason)
type TransitionStatusUseCase struct {
	permitRepo repository.WorkPermitRepository
}

// NewTransitionStatusUseCase constructs a new TransitionStatusUseCase.
func NewTransitionStatusUseCase(permitRepo repository.WorkPermitRepository) *TransitionStatusUseCase {
	return &TransitionStatusUseCase{permitRepo: permitRepo}
}

// Execute validates and performs the requested state transition.
func (uc *TransitionStatusUseCase) Execute(ctx context.Context, input TransitionStatusInput) (*entity.WorkPermit, error) {
	if input.PermitID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El ID del permiso es obligatorio.")
	}

	permit, err := uc.permitRepo.GetByID(ctx, input.PermitID)
	if err != nil {
		return nil, err
	}

	userID, _ := tenant.UserIDFromContext(ctx)
	var userIDPtr *uuid.UUID
	if userID != uuid.Nil {
		userIDPtr = &userID
	}

	action := strings.ToLower(strings.TrimSpace(input.Action))
	var targetStatus entity.PermitStatus

	switch action {
	case "submit":
		if permit.Status != entity.PermitStatusDraft {
			return nil, domainErrors.ErrPermitInvalidState.WithMessage("Solo los permisos en estatus 'draft' pueden enviarse a aprobación.")
		}
		if len(permit.Workers) == 0 {
			return nil, domainErrors.ErrPermitNoWorkers
		}
		targetStatus = entity.PermitStatusPendingApproval

	case "suspend":
		if permit.Status != entity.PermitStatusApprovedActive {
			return nil, domainErrors.ErrPermitInvalidState.WithMessage("Solo los permisos activos pueden ser suspendidos.")
		}
		if input.Reason == nil || strings.TrimSpace(*input.Reason) == "" {
			return nil, domainErrors.ErrInvalidInput.WithMessage("Se requiere especificar el motivo de la suspensión del permiso.")
		}
		targetStatus = entity.PermitStatusSuspended

	case "resume":
		if permit.Status != entity.PermitStatusSuspended {
			return nil, domainErrors.ErrPermitInvalidState.WithMessage("Solo los permisos suspendidos pueden ser reanudados.")
		}
		targetStatus = entity.PermitStatusApprovedActive

	case "close":
		if permit.Status != entity.PermitStatusApprovedActive && permit.Status != entity.PermitStatusSuspended {
			return nil, domainErrors.ErrPermitInvalidState.WithMessage("Solo los permisos activos o suspendidos pueden ser concluidos.")
		}
		targetStatus = entity.PermitStatusClosed

	case "cancel":
		if permit.Status.IsTerminal() {
			return nil, domainErrors.ErrPermitInvalidState.WithMessage("No se puede cancelar un permiso que ya ha concluido o sido cancelado.")
		}
		if input.Reason == nil || strings.TrimSpace(*input.Reason) == "" {
			return nil, domainErrors.ErrInvalidInput.WithMessage("Se requiere especificar el motivo de la cancelación del permiso.")
		}
		targetStatus = entity.PermitStatusCancelled

	default:
		return nil, domainErrors.ErrInvalidInput.WithMessage("Acción de transición de estado desconocida.")
	}

	// Update status in repository
	if err := uc.permitRepo.UpdateStatus(ctx, permit.ID, targetStatus, input.Reason, userIDPtr); err != nil {
		return nil, err
	}

	// If submit action, log contractor supervisor digital signature
	if action == "submit" && userID != uuid.Nil {
		sig := &entity.PermitSignature{
			ID:            uuid.New(),
			TenantID:      permit.TenantID,
			PermitID:      permit.ID,
			UserID:        userID,
			Role:          entity.PermitSignerContractorSupervisor,
			SignatureData: input.SignatureData,
			IPAddress:     input.IPAddress,
			UserAgent:     input.UserAgent,
			Notes:         input.Notes,
			SignedAt:      time.Now().UTC(),
		}
		_ = uc.permitRepo.AddSignature(ctx, sig)
	}

	return uc.permitRepo.GetByID(ctx, permit.ID)
}
