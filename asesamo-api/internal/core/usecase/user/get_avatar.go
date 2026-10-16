package user

import (
	"context"
	"io"
	"strings"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// GetAvatarUseCase retrieves the avatar stream for a user.
type GetAvatarUseCase struct {
	userRepo       repository.UserRepository
	storageService repository.StorageService
}

// NewGetAvatarUseCase constructs GetAvatarUseCase.
func NewGetAvatarUseCase(
	userRepo repository.UserRepository,
	storageService repository.StorageService,
) *GetAvatarUseCase {
	return &GetAvatarUseCase{
		userRepo:       userRepo,
		storageService: storageService,
	}
}

// Execute returns the image stream, content type, and error.
func (uc *GetAvatarUseCase) Execute(ctx context.Context, userID uuid.UUID) (io.ReadCloser, string, error) {
	if userID == uuid.Nil {
		return nil, "", domainErrors.ErrInvalidInput.WithMessage("El ID de usuario es obligatorio.")
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if u == nil {
		return nil, "", domainErrors.ErrUserNotFound
	}

	if u.AvatarR2Key == nil || *u.AvatarR2Key == "" {
		return nil, "", domainErrors.ErrDocumentNotFound.WithMessage("El usuario no cuenta con una fotografía de perfil registrada.")
	}

	key := *u.AvatarR2Key
	stream, err := uc.storageService.Download(ctx, key)
	if err != nil {
		return nil, "", err
	}

	contentType := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(key), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(key), ".webp") {
		contentType = "image/webp"
	}

	return stream, contentType, nil
}
