package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/inspection"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockInspectionRepoForHandler struct {
	mock.Mock
}

func (m *MockInspectionRepoForHandler) CreateTemplate(ctx context.Context, template *entity.InspectionTemplate, items []entity.InspectionTemplateItem) error {
	return nil
}
func (m *MockInspectionRepoForHandler) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.InspectionTemplate, []entity.InspectionTemplateItem, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*entity.InspectionTemplate), args.Get(1).([]entity.InspectionTemplateItem), args.Error(2)
}
func (m *MockInspectionRepoForHandler) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]entity.InspectionTemplate, error) {
	return nil, nil
}
func (m *MockInspectionRepoForHandler) CreateInspection(ctx context.Context, inspection *entity.SafetyInspection) error {
	return nil
}
func (m *MockInspectionRepoForHandler) GetInspectionByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyInspection, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SafetyInspection), args.Error(1)
}
func (m *MockInspectionRepoForHandler) ListInspections(ctx context.Context, tenantID uuid.UUID, filter repository.InspectionFilter) ([]entity.SafetyInspection, int, error) {
	return nil, 0, nil
}
func (m *MockInspectionRepoForHandler) SaveInspectionAnswers(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []entity.InspectionItemAnswer) error {
	return nil
}
func (m *MockInspectionRepoForHandler) CompleteInspection(ctx context.Context, inspection *entity.SafetyInspection, capas []entity.InspectionActionPlan) error {
	args := m.Called(ctx, inspection, capas)
	return args.Error(0)
}
func (m *MockInspectionRepoForHandler) ListActionPlans(ctx context.Context, tenantID uuid.UUID, inspectionID *uuid.UUID, status *entity.CAPAStatus) ([]entity.InspectionActionPlan, error) {
	return nil, nil
}
func (m *MockInspectionRepoForHandler) UpdateActionPlanStatus(ctx context.Context, tenantID, capaID uuid.UUID, status entity.CAPAStatus, notes string) error {
	return nil
}

func TestInspectionHandler_CompleteInspection_Endpoint(t *testing.T) {
	tenantID := uuid.New()
	inspectionID := uuid.New()
	templateID := uuid.New()
	inspectorID := uuid.New()

	existingInspection := entity.NewSafetyInspection(tenantID, templateID, "INSP-2026-0010", entity.InspectionTypeRoutine, "Nave Principal", inspectorID)
	existingInspection.ID = inspectionID

	templateItems := []entity.InspectionTemplateItem{
		{
			ID:           uuid.New(),
			TemplateID:   templateID,
			Category:     "Seguridad Eléctrica",
			Question:     "¿Tableros eléctricos cerrados y señalizados?",
			Weight:       2,
			CriticalRisk: true,
		},
	}

	mockRepo := new(MockInspectionRepoForHandler)
	mockRepo.On("GetInspectionByID", mock.Anything, tenantID, inspectionID).Return(existingInspection, nil)
	mockRepo.On("GetTemplateByID", mock.Anything, tenantID, templateID).Return(&entity.InspectionTemplate{ID: templateID}, templateItems, nil)
	mockRepo.On("CompleteInspection", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	completeUC := usecase.NewCompleteInspectionUseCase(mockRepo)
	inspHandler := handler.NewInspectionHandler(mockRepo, completeUC)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := tenant.WithTenant(req.Context(), tenantID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Post("/api/v1/inspections/{id}/complete", inspHandler.CompleteInspection)

	bodyPayload := map[string]interface{}{
		"answers": []map[string]interface{}{
			{
				"template_item_id": templateItems[0].ID.String(),
				"status":           "compliant",
			},
		},
	}
	bodyBytes, _ := json.Marshal(bodyPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/inspections/"+inspectionID.String()+"/complete", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	data := res["data"].(map[string]interface{})
	assert.Equal(t, 100.0, data["score_percentage"])
	assert.Equal(t, "completed", data["status"])
	assert.False(t, data["requires_capa"].(bool))
}
