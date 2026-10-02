package access

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ListAccessLogsInput contains parameters for listing logs.
type ListAccessLogsInput struct {
	TenantID uuid.UUID `json:"tenant_id"`
	PlantID  uuid.UUID `json:"plant_id"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
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

// Execute retrieves paginated access logs for a plant.
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

	var (
		logs  []entity.AccessLog
		total int
		err   error
	)

	if in.PlantID != uuid.Nil {
		logs, total, err = uc.accessLogRepo.ListByPlant(ctx, in.TenantID, in.PlantID, in.PerPage, offset)
	} else {
		logs, err = uc.accessLogRepo.ListByTenant(ctx, in.TenantID, in.PerPage, offset)
		total = len(logs)
	}
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
