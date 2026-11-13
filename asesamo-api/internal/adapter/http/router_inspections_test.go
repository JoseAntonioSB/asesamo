package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	adapterHTTP "asesamo-api/internal/adapter/http"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/inspection"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type RouterMockInspectionRepo struct {
	mock.Mock
}

func (m *RouterMockInspectionRepo) CreateTemplate(ctx context.Context, template *entity.InspectionTemplate, items []entity.InspectionTemplateItem) error {
	return nil
}
func (m *RouterMockInspectionRepo) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.InspectionTemplate, []entity.InspectionTemplateItem, error) {
	return nil, nil, nil
}
func (m *RouterMockInspectionRepo) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]entity.InspectionTemplate, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.InspectionTemplate), args.Error(1)
}
func (m *RouterMockInspectionRepo) CreateInspection(ctx context.Context, inspection *entity.SafetyInspection) error {
	return nil
}
func (m *RouterMockInspectionRepo) GetInspectionByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyInspection, error) {
	return nil, nil
}
func (m *RouterMockInspectionRepo) ListInspections(ctx context.Context, tenantID uuid.UUID, filter repository.InspectionFilter) ([]entity.SafetyInspection, int, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]entity.SafetyInspection), args.Int(1), args.Error(2)
}
func (m *RouterMockInspectionRepo) SaveInspectionAnswers(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []entity.InspectionItemAnswer) error {
	return nil
}
func (m *RouterMockInspectionRepo) CompleteInspection(ctx context.Context, inspection *entity.SafetyInspection, capas []entity.InspectionActionPlan) error {
	return nil
}
func (m *RouterMockInspectionRepo) ListActionPlans(ctx context.Context, tenantID uuid.UUID, inspectionID *uuid.UUID, status *entity.CAPAStatus) ([]entity.InspectionActionPlan, error) {
	return nil, nil
}
func (m *RouterMockInspectionRepo) UpdateActionPlanStatus(ctx context.Context, tenantID, capaID uuid.UUID, status entity.CAPAStatus, notes string) error {
	return nil
}

func TestRouter_InspectionRoutes_AuthenticatedAccess(t *testing.T) {
	jwtSecret := "super_secret_jwt_key_min_32_characters_long_12345"
	tenantID := uuid.New()
	userID := uuid.New()

	token, err := jwt.SignToken(userID, tenantID, "ehs_manager", []byte(jwtSecret))
	require.NoError(t, err)

	mockInspectionRepo := new(RouterMockInspectionRepo)
	completeUC := usecase.NewCompleteInspectionUseCase(mockInspectionRepo)
	inspHandler := handler.NewInspectionHandler(mockInspectionRepo, completeUC)

	// Inyectar en router
	cfg := adapterHTTP.RouterConfig{
		HealthHandler:     handler.NewHealthHandler(nil, "1.0.0"),
		InspectionHandler: inspHandler,
		JWTSecret:         jwtSecret,
		CORSOrigins:       []string{"*"},
	}

	mockInspectionRepo.On("ListTemplates", mock.Anything, tenantID).Return([]entity.InspectionTemplate{
		{ID: uuid.New(), Title: "NOM-002 Extintores", Category: "fire_safety"},
	}, nil)

	r := adapterHTTP.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inspections/templates", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
