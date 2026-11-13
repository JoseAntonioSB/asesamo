package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type InspectionFilter struct {
	Status       *entity.InspectionStatus
	ContractorID *uuid.UUID
	InspectorID  *uuid.UUID
	Limit        int
	Offset       int
}

type SafetyInspectionRepository interface {
	CreateTemplate(ctx context.Context, template *entity.InspectionTemplate, items []entity.InspectionTemplateItem) error
	GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.InspectionTemplate, []entity.InspectionTemplateItem, error)
	ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]entity.InspectionTemplate, error)

	CreateInspection(ctx context.Context, inspection *entity.SafetyInspection) error
	GetInspectionByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyInspection, error)
	ListInspections(ctx context.Context, tenantID uuid.UUID, filter InspectionFilter) ([]entity.SafetyInspection, int, error)
	SaveInspectionAnswers(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []entity.InspectionItemAnswer) error
	CompleteInspection(ctx context.Context, inspection *entity.SafetyInspection, capas []entity.InspectionActionPlan) error

	ListActionPlans(ctx context.Context, tenantID uuid.UUID, inspectionID *uuid.UUID, status *entity.CAPAStatus) ([]entity.InspectionActionPlan, error)
	UpdateActionPlanStatus(ctx context.Context, tenantID, capaID uuid.UUID, status entity.CAPAStatus, notes string) error
}
