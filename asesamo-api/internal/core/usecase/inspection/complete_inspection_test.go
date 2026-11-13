package inspection_test

import (
	"context"
	"testing"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/inspection"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockInspectionRepo struct {
	mock.Mock
}

func (m *MockInspectionRepo) CreateTemplate(ctx context.Context, template *entity.InspectionTemplate, items []entity.InspectionTemplateItem) error {
	args := m.Called(ctx, template, items)
	return args.Error(0)
}

func (m *MockInspectionRepo) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.InspectionTemplate, []entity.InspectionTemplateItem, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.InspectionTemplate), args.Get(1).([]entity.InspectionTemplateItem), args.Error(2)
}

func (m *MockInspectionRepo) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]entity.InspectionTemplate, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]entity.InspectionTemplate), args.Error(1)
}

func (m *MockInspectionRepo) CreateInspection(ctx context.Context, inspection *entity.SafetyInspection) error {
	args := m.Called(ctx, inspection)
	return args.Error(0)
}

func (m *MockInspectionRepo) GetInspectionByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyInspection, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SafetyInspection), args.Error(1)
}

func (m *MockInspectionRepo) ListInspections(ctx context.Context, tenantID uuid.UUID, filter repository.InspectionFilter) ([]entity.SafetyInspection, int, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]entity.SafetyInspection), args.Int(1), args.Error(2)
}

func (m *MockInspectionRepo) SaveInspectionAnswers(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []entity.InspectionItemAnswer) error {
	args := m.Called(ctx, tenantID, inspectionID, answers)
	return args.Error(0)
}

func (m *MockInspectionRepo) CompleteInspection(ctx context.Context, inspection *entity.SafetyInspection, capas []entity.InspectionActionPlan) error {
	args := m.Called(ctx, inspection, capas)
	return args.Error(0)
}

func (m *MockInspectionRepo) ListActionPlans(ctx context.Context, tenantID uuid.UUID, inspectionID *uuid.UUID, status *entity.CAPAStatus) ([]entity.InspectionActionPlan, error) {
	args := m.Called(ctx, tenantID, inspectionID, status)
	return args.Get(0).([]entity.InspectionActionPlan), args.Error(1)
}

func (m *MockInspectionRepo) UpdateActionPlanStatus(ctx context.Context, tenantID, capaID uuid.UUID, status entity.CAPAStatus, notes string) error {
	args := m.Called(ctx, tenantID, capaID, status, notes)
	return args.Error(0)
}

func TestCompleteInspectionUseCase_Execute(t *testing.T) {
	mockRepo := new(MockInspectionRepo)
	uc := usecase.NewCompleteInspectionUseCase(mockRepo)

	tenantID := uuid.New()
	inspectionID := uuid.New()
	templateID := uuid.New()
	inspectorID := uuid.New()

	existingInspection := entity.NewSafetyInspection(tenantID, templateID, "INSP-2026-0005", entity.InspectionTypePlanned, "Nave 2", inspectorID)
	existingInspection.ID = inspectionID

	templateItems := []entity.InspectionTemplateItem{
		{
			ID:           uuid.New(),
			TemplateID:   templateID,
			Category:     "Extintores",
			Question:     "¿Extintor con presión adecuada y sin obstrucciones?",
			Weight:       3,
			CriticalRisk: true,
		},
		{
			ID:           uuid.New(),
			TemplateID:   templateID,
			Category:     "Orden",
			Question:     "¿Pasillos libres de obstáculos?",
			Weight:       1,
			CriticalRisk: false,
		},
	}

	mockRepo.On("GetInspectionByID", mock.Anything, tenantID, inspectionID).Return(existingInspection, nil)
	mockRepo.On("GetTemplateByID", mock.Anything, tenantID, templateID).Return(&entity.InspectionTemplate{ID: templateID}, templateItems, nil)

	answersInput := []usecase.SubmitAnswerInput{
		{
			TemplateItemID:   templateItems[0].ID,
			Status:           entity.ItemComplianceNonCompliant,
			FindingNotes:     "Extintor descargado",
			EvidencePhotoURL: "https://r2.asesamo.com/ext.jpg",
		},
		{
			TemplateItemID: templateItems[1].ID,
			Status:         entity.ItemComplianceCompliant,
		},
	}

	mockRepo.On("CompleteInspection", mock.Anything, mock.MatchedBy(func(insp *entity.SafetyInspection) bool {
		return insp.Status == entity.InspectionStatusRequiresCAPA && *insp.ScorePercentage == 25.0
	}), mock.MatchedBy(func(capas []entity.InspectionActionPlan) bool {
		return len(capas) == 1 && capas[0].Severity == entity.FindingSeverityCritical && capas[0].Description == "Extintor descargado"
	})).Return(nil)

	res, err := uc.Execute(context.Background(), tenantID, inspectionID, answersInput)
	require.NoError(t, err)
	assert.Equal(t, 25.0, res.ScorePercentage)
	assert.True(t, res.RequiresCAPA)
	assert.Equal(t, entity.InspectionStatusRequiresCAPA, res.Status)
	assert.Len(t, res.GeneratedCAPAs, 1)

	mockRepo.AssertExpectations(t)
}
