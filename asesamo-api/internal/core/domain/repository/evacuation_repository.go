package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type EmergencyEvacuationRepository interface {
	CreateMusterPoint(ctx context.Context, point *entity.MusterPoint) error
	ListMusterPoints(ctx context.Context, tenantID uuid.UUID) ([]entity.MusterPoint, error)

	StartEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error
	GetActiveEvacuation(ctx context.Context, tenantID uuid.UUID) (*entity.EmergencyEvacuation, error)
	GetEvacuationByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EmergencyEvacuation, error)
	ListEvacuations(ctx context.Context, tenantID uuid.UUID) ([]entity.EmergencyEvacuation, error)
	UpdateRollCallCheckIn(ctx context.Context, evacID, workerID, musterPointID, checkedInBy uuid.UUID, notes string) error
	EndEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error

	// Obtener trabajadores actualmente dentro de la planta según registros de caseta
	GetPresentWorkersInPlant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error)
}
