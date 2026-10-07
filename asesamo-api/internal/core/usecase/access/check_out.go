package access

import (
	"context"
	"math"
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
	AccessLogID     uuid.UUID           `json:"access_log_id"`
	Result          entity.AccessResult `json:"result,omitempty"`
	CheckedInAt     time.Time           `json:"checked_in_at,omitempty"`
	CheckedOutAt    time.Time           `json:"checked_out_at"`
	DurationMinutes int                 `json:"duration_minutes"`
}

// CheckOutUseCase manages gate exit check-out recording.
type CheckOutUseCase struct {
	accessLogRepo   repository.AccessLogRepository
	accessPointRepo repository.AccessPointRepository
}

// NewCheckOutUseCase creates a new CheckOutUseCase.
func NewCheckOutUseCase(logRepo repository.AccessLogRepository, apRepo ...repository.AccessPointRepository) *CheckOutUseCase {
	var ap repository.AccessPointRepository
	if len(apRepo) > 0 {
		ap = apRepo[0]
	}
	return &CheckOutUseCase{
		accessLogRepo:   logRepo,
		accessPointRepo: ap,
	}
}

// Execute records worker exit.
func (uc *CheckOutUseCase) Execute(ctx context.Context, in CheckOutInput) (*CheckOutOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}

	// Resolve plant_id from access point if omitted
	if in.PlantID == uuid.Nil && in.AccessPointID != uuid.Nil && uc.accessPointRepo != nil {
		ap, err := uc.accessPointRepo.GetByID(ctx, in.AccessPointID)
		if err == nil && ap != nil {
			in.PlantID = ap.PlantID
		}
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

	durationMinutes := int(math.Max(0, outTime.Sub(updatedLog.CheckedInAt).Minutes()))

	return &CheckOutOutput{
		AccessLogID:     updatedLog.ID,
		Result:          updatedLog.Result,
		CheckedInAt:     updatedLog.CheckedInAt,
		CheckedOutAt:    outTime,
		DurationMinutes: durationMinutes,
	}, nil
}
