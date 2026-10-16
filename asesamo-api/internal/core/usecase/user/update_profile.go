package user

import (
	"context"
	"strings"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileInput holds parameters for updating user profile details.
type UpdateProfileInput struct {
	UserID          uuid.UUID
	FullName        string
	Phone           *string
	CurrentPassword string
	NewPassword     string
}

// UpdateProfileUseCase handles profile information and password updates.
type UpdateProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewUpdateProfileUseCase constructs UpdateProfileUseCase.
func NewUpdateProfileUseCase(userRepo repository.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{userRepo: userRepo}
}

// Execute validates and saves profile modifications.
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, in UpdateProfileInput) error {
	if in.UserID == uuid.Nil {
		return domainErrors.ErrUnauthorized
	}

	fullName := strings.TrimSpace(in.FullName)
	if fullName == "" {
		return domainErrors.ErrInvalidInput.WithMessage("El nombre completo es obligatorio.")
	}

	u, err := uc.userRepo.GetByID(ctx, in.UserID)
	if err != nil {
		return err
	}

	var phonePtr *string
	if in.Phone != nil {
		trimmedPhone := strings.TrimSpace(*in.Phone)
		if trimmedPhone != "" {
			phonePtr = &trimmedPhone
		}
	}

	// 1. If changing password, verify current password first
	var newPasswordHash string
	if in.NewPassword != "" {
		if in.CurrentPassword == "" {
			return domainErrors.ErrInvalidInput.WithMessage("Debe ingresar su contraseña actual para establecer una nueva.")
		}
		if len(in.NewPassword) < 8 {
			return domainErrors.ErrInvalidInput.WithMessage("La nueva contraseña debe tener al menos 8 caracteres.")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.CurrentPassword)); err != nil {
			return domainErrors.ErrInvalidCredentials.WithMessage("La contraseña actual es incorrecta.")
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), 10)
		if err != nil {
			return domainErrors.ErrInternal.WithErr(err)
		}
		newPasswordHash = string(newHash)
	}

	// 2. Update basic profile
	if err := uc.userRepo.UpdateProfile(ctx, in.UserID, fullName, phonePtr); err != nil {
		return err
	}

	// 3. Update password if changed
	if newPasswordHash != "" {
		if err := uc.userRepo.UpdatePassword(ctx, in.UserID, newPasswordHash); err != nil {
			return err
		}
	}

	return nil
}
