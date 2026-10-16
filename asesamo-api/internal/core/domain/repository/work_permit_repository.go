package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// PermitFilter encapsulates search criteria and pagination for work permits.
type PermitFilter struct {
	PlantID      *uuid.UUID
	ContractorID *uuid.UUID
	Status       *entity.PermitStatus
	PermitType   *entity.PermitType
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PerPage      int
}

// WorkPermitRepository defines the data access contract for High-Risk Work Permits (PTAR Digital).
type WorkPermitRepository interface {
	// Permits
	Create(ctx context.Context, permit *entity.WorkPermit) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WorkPermit, error)
	ListWithFilters(ctx context.Context, filter PermitFilter) ([]entity.WorkPermit, int, error)
	NextPermitNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PermitStatus, reason *string, byUserID *uuid.UUID) error

	// Workers
	AssignWorker(ctx context.Context, worker *entity.PermitWorker) error
	RemoveWorker(ctx context.Context, permitID, workerID uuid.UUID) error
	ListWorkersByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitWorker, error)

	// Checklists
	InitChecklistItems(ctx context.Context, permitID uuid.UUID, items []entity.PermitChecklistItem) error
	ListChecklistByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitChecklistItem, error)
	UpdateChecklistItem(ctx context.Context, permitID uuid.UUID, itemKey string, isVerified bool, verifiedBy *uuid.UUID, notes *string, measurements map[string]any) error

	// Signatures
	AddSignature(ctx context.Context, sig *entity.PermitSignature) error
	ListSignaturesByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitSignature, error)
}
