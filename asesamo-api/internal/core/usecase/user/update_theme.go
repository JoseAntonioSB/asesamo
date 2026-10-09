package user

import (
	"context"
	"strings"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// UpdateThemeInput holds parameters for updating a user's theme.
type UpdateThemeInput struct {
	UserID uuid.UUID
	Theme  string
}

// UpdateThemeUseCase handles updating the theme preference for an authenticated user.
type UpdateThemeUseCase struct {
	userRepo repository.UserRepository
}

// NewUpdateThemeUseCase instantiates an UpdateThemeUseCase.
func NewUpdateThemeUseCase(userRepo repository.UserRepository) *UpdateThemeUseCase {
	return &UpdateThemeUseCase{userRepo: userRepo}
}

// Execute validates and persists the new theme preference.
func (uc *UpdateThemeUseCase) Execute(ctx context.Context, input UpdateThemeInput) (string, error) {
	theme := strings.ToLower(strings.TrimSpace(input.Theme))
	if theme != "light" && theme != "dark" && theme != "system" {
		return "", domainErrors.ErrInvalidInput.WithMessage("El tema debe ser 'light', 'dark' o 'system'.")
	}

	if input.UserID == uuid.Nil {
		return "", domainErrors.ErrUnauthorized
	}

	if err := uc.userRepo.UpdateThemePreference(ctx, input.UserID, theme); err != nil {
		return "", err
	}

	return theme, nil
}
