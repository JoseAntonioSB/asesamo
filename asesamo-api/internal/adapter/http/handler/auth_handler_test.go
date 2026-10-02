package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepoForHandler struct {
	mock.Mock
}

func (m *MockUserRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepoForHandler) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepoForHandler) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	args := m.Called(ctx, id, lastLoginAt)
	return args.Error(0)
}

const testJWTSecret = "super_secret_jwt_key_min_32_characters_long_12345"

func TestAuthHandler_Login(t *testing.T) {
	password := "Password123!"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenantID := uuid.New()
	userID := uuid.New()
	testUser := &entity.User{
		ID:           userID,
		TenantID:     &tenantID,
		Email:        "guard@asesamo.com",
		PasswordHash: string(hashed),
		FullName:     "Oficial Roberto Caseta",
		Role:         entity.RoleGuard,
		IsActive:     true,
	}

	t.Run("successful login returns 200 and token with user envelope", func(t *testing.T) {
		mockRepo := new(MockUserRepoForHandler)
		userCopy := *testUser
		mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(&userCopy, nil)
		mockRepo.On("UpdateLastLogin", mock.Anything, userID, mock.Anything).Return(nil)

		uc := auth.NewLoginUseCase(mockRepo, testJWTSecret)
		h := handler.NewAuthHandler(uc)

		payload := dto.LoginRequest{
			Email:    "guard@asesamo.com",
			Password: password,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp struct {
			Data dto.LoginResponseData `json:"data"`
		}
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Data.Token)
		assert.Equal(t, userID, resp.Data.User.ID)
		assert.Equal(t, "guard@asesamo.com", resp.Data.User.Email)
		assert.Equal(t, "guard", resp.Data.User.Role)
		assert.Equal(t, "Oficial Roberto Caseta", resp.Data.User.FullName)
		assert.Equal(t, &tenantID, resp.Data.User.TenantID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password returns 401 INVALID_CREDENTIALS", func(t *testing.T) {
		mockRepo := new(MockUserRepoForHandler)
		userCopy := *testUser
		mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(&userCopy, nil)

		uc := auth.NewLoginUseCase(mockRepo, testJWTSecret)
		h := handler.NewAuthHandler(uc)

		payload := dto.LoginRequest{
			Email:    "guard@asesamo.com",
			Password: "WrongPassword!",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
		assert.Contains(t, resp.Error.Message, "Credenciales inválidas")
		mockRepo.AssertExpectations(t)
	})

	t.Run("non-existent user returns 401 INVALID_CREDENTIALS", func(t *testing.T) {
		mockRepo := new(MockUserRepoForHandler)
		mockRepo.On("GetByEmail", mock.Anything, "notfound@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)

		uc := auth.NewLoginUseCase(mockRepo, testJWTSecret)
		h := handler.NewAuthHandler(uc)

		payload := dto.LoginRequest{
			Email:    "notfound@asesamo.com",
			Password: password,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, "INVALID_CREDENTIALS", resp.Error.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty body returns 400 INVALID_INPUT", func(t *testing.T) {
		uc := auth.NewLoginUseCase(new(MockUserRepoForHandler), testJWTSecret)
		h := handler.NewAuthHandler(uc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("{}")))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp dto.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "INVALID_INPUT", resp.Error.Code)
	})

	t.Run("malformed json returns 400 INVALID_INPUT", func(t *testing.T) {
		uc := auth.NewLoginUseCase(new(MockUserRepoForHandler), testJWTSecret)
		h := handler.NewAuthHandler(uc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("not a json")))
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("nil request body returns 400 INVALID_INPUT", func(t *testing.T) {
		uc := auth.NewLoginUseCase(new(MockUserRepoForHandler), testJWTSecret)
		h := handler.NewAuthHandler(uc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		rec := httptest.NewRecorder()

		h.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
