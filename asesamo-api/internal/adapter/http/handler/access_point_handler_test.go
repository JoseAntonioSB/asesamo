package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAccessPointRepoForHandler struct {
	mock.Mock
}

func (m *MockAccessPointRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepoForHandler) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepoForHandler) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func TestAccessPointHandler_List(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	desc := "Caseta de vigilancia principal"

	points := []entity.AccessPoint{
		{
			ID:          apID,
			PlantID:     plantID,
			TenantID:    tenantID,
			Name:        "Caseta Principal Norte",
			Description: &desc,
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	t.Run("returns active access points in standard envelope", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepoForHandler)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(points, nil)

		uc := access.NewListAccessPointsUseCase(mockRepo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp struct {
			Data []dto.AccessPointResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		require.Len(t, resp.Data, 1)
		assert.Equal(t, apID, resp.Data[0].ID)
		assert.Equal(t, "Caseta Principal Norte", resp.Data[0].Name)
		assert.Equal(t, true, resp.Data[0].IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("returns empty array in data envelope when no access points found", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepoForHandler)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return([]entity.AccessPoint{}, nil)

		uc := access.NewListAccessPointsUseCase(mockRepo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Data []dto.AccessPointResponse `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotNil(t, resp.Data)
		assert.Empty(t, resp.Data)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing tenant context returns 401 UNAUTHORIZED", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepoForHandler)
		uc := access.NewListAccessPointsUseCase(mockRepo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("use case error returns 500 INTERNAL_ERROR", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepoForHandler)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(([]entity.AccessPoint)(nil), errors.New("db timeout"))

		uc := access.NewListAccessPointsUseCase(mockRepo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "DATABASE_ERROR", resp.Error.Code)
		mockRepo.AssertExpectations(t)
	})
}
