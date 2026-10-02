package access

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// CheckOutInput contains parameters for gate exit.
type CheckOutInput struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	PlantID       uuid.UUID `json:"plant_id"`
	AccessPointID uuid.UUID `json:"access_point_id"`
	WorkerUserID  uuid.UUID `json:"worker_user_id"`
}

// CheckOutOutput contains the check-out outcome.
type CheckOutOutput struct {
	AccessLogID  uuid.UUID           `json:"access_log_id"`
	Result       entity.AccessResult `json:"result"`
	CheckedInAt  time.Time           `json:"checked_in_at"`
	CheckedOutAt time.Time           `json:"checked_out_at"`
}

// CheckOutUseCase manages gate exit check-out recording.
type CheckOutUseCase struct {
	accessLogRepo repository.AccessLogRepository
}

// NewCheckOutUseCase creates a new CheckOutUseCase.
func NewCheckOutUseCase(logRepo repository.AccessLogRepository) *CheckOutUseCase {
	return &CheckOutUseCase{accessLogRepo: logRepo}
}

// Execute records worker exit.
func (uc *CheckOutUseCase) Execute(ctx context.Context, in CheckOutInput) (*CheckOutOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}

	now := time.Now().UTC()
	updatedLog, err := uc.accessLogRepo.CheckOutLatest(ctx, in.TenantID, in.PlantID, in.WorkerUserID, now)
	if err != nil {
		return nil, err
	}

	outTime := now
	if updatedLog.CheckedOutAt != nil {
		outTime = *updatedLog.CheckedOutAt
	}

	return &CheckOutOutput{
		AccessLogID:  updatedLog.ID,
		Result:       updatedLog.Result,
		CheckedInAt:  updatedLog.CheckedInAt,
		CheckedOutAt: outTime,
	}, nil
}
