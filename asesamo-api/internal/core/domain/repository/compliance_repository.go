package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// ComplianceRepository provides access to worker affiliation, plant document rules,
// and worker/contractor uploaded documents for compliance evaluation.
type ComplianceRepository interface {
	GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error)
	GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error)
	GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error)
	GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error)
}
