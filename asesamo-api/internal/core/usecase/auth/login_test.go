package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/auth"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository implements repository.UserRepository for testing.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	args := m.Called(ctx, id, lastLoginAt)
	return args.Error(0)
}

const validSecret = "super_secret_jwt_key_min_32_characters_long_12345"

func TestLoginUseCase(t *testing.T) {
	password := "Password123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenantID := uuid.New()
	userID := uuid.New()

	baseUser := &entity.User{
		ID:           userID,
		TenantID:     &tenantID,
		Email:        "guard@asesamo.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Oficial Roberto Caseta",
		Role:         entity.RoleGuard,
		IsActive:     true,
	}

	t.Run("successful login", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userCopy := *baseUser
		mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(&userCopy, nil)
		mockRepo.On("UpdateLastLogin", mock.Anything, userID, mock.Anything).Return(nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "guard@asesamo.com",
			Password: password,
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.NotEmpty(t, output.Token)
		assert.Equal(t, userID, output.User.ID)
		assert.Equal(t, entity.RoleGuard, output.User.Role)

		// Validate that the issued token is authentic and claims match
		claims, err := jwt.ValidateToken(output.Token, []byte(validSecret))
		require.NoError(t, err)
		assert.Equal(t, userID.String(), claims.Subject)
		assert.Equal(t, tenantID, claims.TenantID)
		assert.Equal(t, "guard", claims.Role)

		mockRepo.AssertExpectations(t)
	})

	t.Run("wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userCopy := *baseUser
		mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(&userCopy, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "guard@asesamo.com",
			Password: "WrongPassword!",
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("non-existent user returns ErrInvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByEmail", mock.Anything, "nonexistent@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "nonexistent@asesamo.com",
			Password: password,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil user returned with no error returns ErrInvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByEmail", mock.Anything, "empty@asesamo.com").Return(nil, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "empty@asesamo.com",
			Password: password,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("inactive user with correct password returns ErrUserInactive", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		inactiveUser := *baseUser
		inactiveUser.IsActive = false
		mockRepo.On("GetByEmail", mock.Anything, "inactive@asesamo.com").Return(&inactiveUser, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "inactive@asesamo.com",
			Password: password,
		})

		assert.ErrorIs(t, err, domainErrors.ErrUserInactive)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("inactive user with wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		inactiveUser := *baseUser
		inactiveUser.IsActive = false
		mockRepo.On("GetByEmail", mock.Anything, "inactive@asesamo.com").Return(&inactiveUser, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "inactive@asesamo.com",
			Password: "WrongPassword!",
		})

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("non-existent user timing branch executes dummy bcrypt and returns ErrInvalidCredentials", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByEmail", mock.Anything, "nonexistent@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		start := time.Now()
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "nonexistent@asesamo.com",
			Password: password,
		})
		elapsed := time.Since(start)

		assert.ErrorIs(t, err, domainErrors.ErrInvalidCredentials)
		assert.Nil(t, output)
		assert.True(t, elapsed > 5*time.Millisecond, "Dummy bcrypt comparison must take non-trivial time to mitigate timing attacks")
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty email or password returns ErrInvalidInput", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		uc := auth.NewLoginUseCase(mockRepo, validSecret)

		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "",
			Password: password,
		})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
		assert.Nil(t, output)

		output, err = uc.Execute(context.Background(), auth.LoginInput{
			Email:    "guard@asesamo.com",
			Password: "",
		})
		assert.ErrorIs(t, err, domainErrors.ErrInvalidInput)
		assert.Nil(t, output)
	})

	t.Run("database error returns ErrDatabase", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByEmail", mock.Anything, "db-error@asesamo.com").Return(nil, errors.New("connection reset by peer"))

		uc := auth.NewLoginUseCase(mockRepo, validSecret)
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "db-error@asesamo.com",
			Password: password,
		})

		assert.ErrorIs(t, err, domainErrors.ErrDatabase)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})

	t.Run("jwt signing failure with short secret returns ErrInternal", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userCopy := *baseUser
		mockRepo.On("GetByEmail", mock.Anything, "guard@asesamo.com").Return(&userCopy, nil)

		// Short secret (< 32 chars) triggers jwt.ErrSecretTooShort
		uc := auth.NewLoginUseCase(mockRepo, "too_short")
		output, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "guard@asesamo.com",
			Password: password,
		})

		assert.ErrorIs(t, err, domainErrors.ErrInternal)
		assert.Nil(t, output)
		mockRepo.AssertExpectations(t)
	})
}
