package auth_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAdversarial_LoginUseCase(t *testing.T) {
	password := "CorrectPassword123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenantID := uuid.New()
	userID := uuid.New()

	activeUser := &entity.User{
		ID:           userID,
		TenantID:     &tenantID,
		Email:        "active.guard@asesamo.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Active Guard",
		Role:         entity.RoleGuard,
		IsActive:     true,
	}

	inactiveUser := &entity.User{
		ID:           uuid.New(),
		TenantID:     &tenantID,
		Email:        "inactive.guard@asesamo.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Inactive Guard",
		Role:         entity.RoleGuard,
		IsActive:     false,
	}

	t.Run("VERIFY: User enumeration resilience - identical error for invalid user vs invalid password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		// Setup mock for invalid user
		mockRepo.On("GetByEmail", mock.Anything, "nonexistent@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)
		// Setup mock for valid user with wrong password
		userCopy := *activeUser
		mockRepo.On("GetByEmail", mock.Anything, "active.guard@asesamo.com").Return(&userCopy, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)

		// 1. Non-existent user
		out1, err1 := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "nonexistent@asesamo.com",
			Password: "AnyPassword123!",
		})
		assert.Nil(t, out1)
		require.Error(t, err1)

		// 2. Valid user, wrong password
		out2, err2 := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "active.guard@asesamo.com",
			Password: "WrongPassword123!",
		})
		assert.Nil(t, out2)
		require.Error(t, err2)

		// Both must return identical domain error
		appErr1, ok1 := err1.(*domainErrors.AppError)
		appErr2, ok2 := err2.(*domainErrors.AppError)
		require.True(t, ok1)
		require.True(t, ok2)

		assert.Equal(t, "INVALID_CREDENTIALS", appErr1.Code)
		assert.Equal(t, "INVALID_CREDENTIALS", appErr2.Code)
		assert.Equal(t, appErr1.Message, appErr2.Message, "Error messages must be byte-for-byte identical")
		assert.Equal(t, appErr1.HTTPStatus, appErr2.HTTPStatus, "HTTP statuses must be identical (401)")
	})

	t.Run("VERIFY: Inactive user enumeration prevention - wrong password returns INVALID_CREDENTIALS", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		userCopy := *inactiveUser
		mockRepo.On("GetByEmail", mock.Anything, "inactive.guard@asesamo.com").Return(&userCopy, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)

		out, err := uc.Execute(context.Background(), auth.LoginInput{
			Email:    "inactive.guard@asesamo.com",
			Password: "WrongPasswordWithZeroKnowledge!",
		})
		assert.Nil(t, out)
		require.Error(t, err)

		appErr, ok := err.(*domainErrors.AppError)
		require.True(t, ok)

		t.Logf("Observed error code for inactive account with WRONG password: %s (%s)", appErr.Code, appErr.Message)
		assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
	})

	t.Run("EMPIRICAL TIMING ANALYSIS: Compare latency of nonexistent user vs wrong password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockRepo.On("GetByEmail", mock.Anything, "nonexistent@asesamo.com").Return(nil, domainErrors.ErrUserNotFound)
		userCopy := *activeUser
		mockRepo.On("GetByEmail", mock.Anything, "active.guard@asesamo.com").Return(&userCopy, nil)

		uc := auth.NewLoginUseCase(mockRepo, validSecret)

		// Measure nonexistent user
		startNonExistent := time.Now()
		const iterations = 5
		for i := 0; i < iterations; i++ {
			_, _ = uc.Execute(context.Background(), auth.LoginInput{
				Email:    "nonexistent@asesamo.com",
				Password: "WrongPassword!",
			})
		}
		durationNonExistent := time.Since(startNonExistent) / iterations

		// Measure existent user with wrong password
		startWrongPassword := time.Now()
		for i := 0; i < iterations; i++ {
			_, _ = uc.Execute(context.Background(), auth.LoginInput{
				Email:    "active.guard@asesamo.com",
				Password: "WrongPassword!",
			})
		}
		durationWrongPassword := time.Since(startWrongPassword) / iterations

		timingRatio := float64(durationWrongPassword) / float64(durationNonExistent)
		t.Logf("Timing side-channel empirical measurement:")
		t.Logf("  Avg latency for non-existent user: %v", durationNonExistent)
		t.Logf("  Avg latency for wrong password (bcrypt): %v", durationWrongPassword)
		t.Logf("  Timing ratio (wrong password / nonexistent): %.2fx", timingRatio)

		// Timing must be balanced within a reasonable margin, proving both run bcrypt cost 10
		assert.True(t, timingRatio < 2.5 && timingRatio > 0.4,
			"Timing between non-existent user and wrong password must be balanced (indistinguishable)")
	})
}
