package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapterHTTP "asesamo-api/internal/adapter/http"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type RouterMockUserRepo struct{ mock.Mock }

func (m *RouterMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	return nil, args.Error(1)
}
func (m *RouterMockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *RouterMockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	return nil
}

type RouterMockAccessPointRepo struct{ mock.Mock }

func (m *RouterMockAccessPointRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	return nil, nil
}
func (m *RouterMockAccessPointRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}
func (m *RouterMockAccessPointRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	return nil, nil
}

const routerTestSecret = "super_secret_jwt_key_min_32_characters_long_12345"

func TestRouter(t *testing.T) {
	userRepo := new(RouterMockUserRepo)
	apRepo := new(RouterMockAccessPointRepo)

	healthHandler := handler.NewHealthHandler(nil, "1.0.0")
	authHandler := handler.NewAuthHandler(auth.NewLoginUseCase(userRepo, routerTestSecret))
	apHandler := handler.NewAccessPointHandler(access.NewListAccessPointsUseCase(apRepo))

	r := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:      healthHandler,
		AuthHandler:        authHandler,
		AccessPointHandler: apHandler,
		JWTSecret:          routerTestSecret,
		CORSOrigins:        []string{"*"},
	})

	t.Run("public GET /health succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("public GET /api/v1/health succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("unauthenticated GET /api/v1/access-points returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("authenticated GET /api/v1/access-points succeeds", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		token, err := jwt.SignToken(userID, tenantID, "guard", []byte(routerTestSecret))
		require.NoError(t, err)

		apRepo.On("ListByTenant", mock.Anything, tenantID).Return([]entity.AccessPoint{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
