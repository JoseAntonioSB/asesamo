package access

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ListAccessLogsInput contains parameters for listing logs with multi-column filtering.
type ListAccessLogsInput struct {
	TenantID      uuid.UUID            `json:"tenant_id"`
	PlantID       uuid.UUID            `json:"plant_id"`
	AccessPointID uuid.UUID            `json:"access_point_id"`
	WorkerUserID  uuid.UUID            `json:"worker_user_id"`
	Result        *entity.AccessResult `json:"result,omitempty"`
	From          *time.Time           `json:"from,omitempty"`
	To            *time.Time           `json:"to,omitempty"`
	Page          int                  `json:"page"`
	PerPage       int                  `json:"per_page"`
}

// ListAccessLogsOutput contains paginated logs.
type ListAccessLogsOutput struct {
	Logs    []entity.AccessLog `json:"logs"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
}

// ListAccessLogsUseCase handles audit log retrieval.
type ListAccessLogsUseCase struct {
	accessLogRepo repository.AccessLogRepository
}

// NewListAccessLogsUseCase creates a new ListAccessLogsUseCase.
func NewListAccessLogsUseCase(logRepo repository.AccessLogRepository) *ListAccessLogsUseCase {
	return &ListAccessLogsUseCase{accessLogRepo: logRepo}
}

// Execute retrieves paginated access logs matching multi-criteria filters.
func (uc *ListAccessLogsUseCase) Execute(ctx context.Context, in ListAccessLogsInput) (*ListAccessLogsOutput, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PerPage <= 0 {
		in.PerPage = 20
	}
	if in.PerPage > 100 {
		in.PerPage = 100
	}

	offset := (in.Page - 1) * in.PerPage

	filter := repository.AccessLogFilter{
		TenantID:      in.TenantID,
		PlantID:       in.PlantID,
		AccessPointID: in.AccessPointID,
		WorkerUserID:  in.WorkerUserID,
		Result:        in.Result,
		From:          in.From,
		To:            in.To,
		Limit:         in.PerPage,
		Offset:        offset,
	}

	logs, total, err := uc.accessLogRepo.ListWithFilters(ctx, filter)
	if err != nil {
		return nil, err
	}

	if logs == nil {
		logs = []entity.AccessLog{}
	}

	return &ListAccessLogsOutput{
		Logs:    logs,
		Total:   total,
		Page:    in.Page,
		PerPage: in.PerPage,
	}, nil
}
