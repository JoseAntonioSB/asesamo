package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type IncidentFilter struct {
	Type         *entity.IncidentType
	Severity     *entity.IncidentSeverity
	Status       *entity.IncidentStatus
	ContractorID *uuid.UUID
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

type SafetyIncidentMetrics struct {
	TotalIncidents     int `json:"total_incidents"`
	NearMisses         int `json:"near_misses"`
	UnsafeConditions   int `json:"unsafe_conditions"`
	UnsafeActs         int `json:"unsafe_acts"`
	LostTimeInjuries   int `json:"lost_time_injuries"`
	TotalDaysLost      int `json:"total_days_lost"`
}

type SafetyIncidentRepository interface {
	CreateIncident(ctx context.Context, incident *entity.SafetyIncident) error
	GetIncidentByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyIncident, error)
	ListIncidents(ctx context.Context, tenantID uuid.UUID, filter IncidentFilter) ([]entity.SafetyIncident, int, error)
	UpdateInvestigation(ctx context.Context, incident *entity.SafetyIncident) error
	GetMetrics(ctx context.Context, tenantID uuid.UUID) (*SafetyIncidentMetrics, error)
}
