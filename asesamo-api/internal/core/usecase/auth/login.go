// Package auth implements authentication use cases.
package auth

import (
	"context"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// LoginInput defines the credentials for authenticating a user.
type LoginInput struct {
	Email    string
	Password string
}

// LoginOutput defines the result of a successful authentication.
type LoginOutput struct {
	Token string
	User  *entity.User
}

// LoginUseCase coordinates user authentication and JWT generation.
type LoginUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

// NewLoginUseCase instantiates a LoginUseCase.
func NewLoginUseCase(userRepo repository.UserRepository, jwtSecret string) *LoginUseCase {
	return &LoginUseCase{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Execute validates credentials, verifies account status, and returns a signed JWT.
func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || input.Password == "" {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El correo y la contraseña son obligatorios.")
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if domainErrors.IsNotFound(err) {
			// Prevent user enumeration attack (OWASP A07): return same error as wrong password
			return nil, domainErrors.ErrInvalidCredentials
		}
		return nil, domainErrors.ErrDatabase.WithErr(err)
	}

	if user == nil {
		return nil, domainErrors.ErrInvalidCredentials
	}

	// Verify account status
	if !user.IsActive {
		return nil, domainErrors.ErrUserInactive
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domainErrors.ErrInvalidCredentials
	}

	// Determine tenant UUID
	var tenantID uuid.UUID
	if user.TenantID != nil {
		tenantID = *user.TenantID
	}

	// Sign JWT
	token, err := jwt.SignToken(user.ID, tenantID, string(user.Role), uc.jwtSecret)
	if err != nil {
		return nil, domainErrors.ErrInternal.WithErr(err)
	}

	// Update last login timestamp asynchronously / best-effort
	now := time.Now().UTC()
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID, now)
	user.LastLoginAt = &now

	return &LoginOutput{
		Token: token,
		User:  user,
	}, nil
}
