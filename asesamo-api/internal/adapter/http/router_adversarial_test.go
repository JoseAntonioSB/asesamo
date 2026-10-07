package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapterHTTP "asesamo-api/internal/adapter/http"
	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserRepoForRouterTest struct {
	mock.Mock
}

func (m *MockUserRepoForRouterTest) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepoForRouterTest) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepoForRouterTest) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	args := m.Called(ctx, id, lastLoginAt)
	return args.Error(0)
}

type MockAccessPointRepoForRouterTest struct {
	mock.Mock
}

func (m *MockAccessPointRepoForRouterTest) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepoForRouterTest) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepoForRouterTest) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func setupTestRouter() (http.Handler, *MockUserRepoForRouterTest, *MockAccessPointRepoForRouterTest, string) {
	jwtSecret := "super_secret_jwt_key_min_32_characters_long_12345"

	userRepo := new(MockUserRepoForRouterTest)
	apRepo := new(MockAccessPointRepoForRouterTest)

	loginUC := auth.NewLoginUseCase(userRepo, jwtSecret)
	listAPUC := access.NewListAccessPointsUseCase(apRepo)

	healthH := handler.NewHealthHandler(nil, "1.0.0")
	authH := handler.NewAuthHandler(loginUC)
	apH := handler.NewAccessPointHandler(listAPUC)

	router := adapterHTTP.NewRouter(adapterHTTP.RouterConfig{
		HealthHandler:      healthH,
		AuthHandler:        authH,
		AccessPointHandler: apH,
		JWTSecret:          jwtSecret,
		CORSOrigins:        []string{"*"},
	})

	return router, userRepo, apRepo, jwtSecret
}

func TestAdversarial_RouterAndBruteForce(t *testing.T) {
	t.Run("EMPIRICAL TEST: Brute force resilience against POST /api/v1/auth/login", func(t *testing.T) {
		router, userRepo, _, _ := setupTestRouter()

		// Mock user not found for all attempts
		userRepo.On("GetByEmail", mock.Anything, "victim@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)

		const attempts = 50
		statusCodeCounts := make(map[int]int)

		for i := 0; i < attempts; i++ {
			body, _ := json.Marshal(dto.LoginRequest{
				Email:    "victim@asesamo.com",
				Password: "password123",
			})
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "203.0.113.195:12345" // Simulated single attacker IP

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			statusCodeCounts[rec.Code]++
		}

		t.Logf("Brute force results over %d burst requests: %+v", attempts, statusCodeCounts)

		// Check rate limiting enforcement (HTTP 429 Too Many Requests)
		rateLimitedCount := statusCodeCounts[http.StatusTooManyRequests]
		unauthorizedCount := statusCodeCounts[http.StatusUnauthorized]

		t.Logf("Rate-limited requests (HTTP 429): %d", rateLimitedCount)
		t.Logf("Processed unauthorized attempts (HTTP 401): %d", unauthorizedCount)

		assert.Equal(t, 10, unauthorizedCount, "First 10 attempts should be processed by auth handler")
		assert.Equal(t, 40, rateLimitedCount, "Subsequent 40 attempts must be rate-limited with HTTP 429")

		// Verify 429 response structure conforms to standard error envelope
		req429 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"email":"victim@asesamo.com","password":"pwd"}`)))
		req429.Header.Set("Content-Type", "application/json")
		req429.RemoteAddr = "203.0.113.195:12345"
		rec429 := httptest.NewRecorder()
		router.ServeHTTP(rec429, req429)

		assert.Equal(t, http.StatusTooManyRequests, rec429.Code)
		var errResp dto.ErrorResponse
		err := json.Unmarshal(rec429.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "TOO_MANY_REQUESTS", errResp.Error.Code)
		assert.Equal(t, "Demasiadas solicitudes. Intente más tarde.", errResp.Error.Message)
	})

	t.Run("FAIL-CLOSED TEST: Missing or nil tenant in JWT prevents cross-tenant access", func(t *testing.T) {
		router, _, _, jwtSecret := setupTestRouter()

		// Forge a validly signed JWT with nil tenantID (uuid.Nil: 00000000-0000-0000-0000-000000000000)
		userID := uuid.New()
		nilTenantID := uuid.Nil
		tokenWithNilTenant, err := jwt.SignToken(userID, nilTenantID, "guard", []byte(jwtSecret))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		req.Header.Set("Authorization", "Bearer "+tokenWithNilTenant)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		t.Logf("Response status for token with nil tenant: %d (Body: %s)", rec.Code, rec.Body.String())

		// Must fail closed: AccessPointHandler checks tenant.FromContext, which rejects uuid.Nil with ErrTenantNotFound -> 401 UNAUTHORIZED
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Endpoint must fail closed and return 401 when tenant is nil/missing")
		var resp dto.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})
}
