package user

import (
	"bytes"
	"context"
	"fmt"
	"io"

	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

const (
	MaxAvatarFileSize = 5 * 1024 * 1024 // 5 MB
)

// UploadAvatarInput defines inputs for uploading user / employee profile picture.
type UploadAvatarInput struct {
	TargetUserID uuid.UUID
	CallerID     uuid.UUID
	CallerRole   string
	FileName     string
	File         io.Reader
	FileSize     int64
}

// UploadAvatarUseCase handles avatar upload, image inspection, and storage.
type UploadAvatarUseCase struct {
	userRepo       repository.UserRepository
	storageService repository.StorageService
}

// NewUploadAvatarUseCase constructs UploadAvatarUseCase.
func NewUploadAvatarUseCase(
	userRepo repository.UserRepository,
	storageService repository.StorageService,
) *UploadAvatarUseCase {
	return &UploadAvatarUseCase{
		userRepo:       userRepo,
		storageService: storageService,
	}
}

// ValidateImageMagicBytes checks whether the header bytes belong to JPEG, PNG, or WebP.
func ValidateImageMagicBytes(header []byte) (string, string, error) {
	if len(header) < 12 {
		return "", "", domainErrors.ErrInvalidInput.WithMessage("Archivo de imagen demasiado pequeño o dañado.")
	}

	// JPEG: FF D8 FF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image/jpeg", "jpg", nil
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 &&
		header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A {
		return "image/png", "png", nil
	}

	// WebP: 'RIFF' .... 'WEBP'
	if string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return "image/webp", "webp", nil
	}

	return "", "", domainErrors.ErrInvalidFileType.WithMessage("Formato de imagen no permitido. Solo se admiten archivos JPG, PNG o WebP.")
}

// Execute validates permissions, inspects image stream, and persists avatar.
func (uc *UploadAvatarUseCase) Execute(ctx context.Context, in UploadAvatarInput) (string, error) {
	if in.TargetUserID == uuid.Nil {
		return "", domainErrors.ErrInvalidInput.WithMessage("El usuario objetivo es obligatorio.")
	}

	// 1. Permissions check
	isSelf := in.CallerID == in.TargetUserID
	isManager := in.CallerRole == "contractor_admin" || in.CallerRole == "plant_admin" || in.CallerRole == "ehs_manager" || in.CallerRole == "super_admin"

	if !isSelf && !isManager {
		return "", domainErrors.ErrForbidden.WithMessage("No cuenta con privilegios para modificar la fotografía de este usuario.")
	}

	// Verify target user exists
	targetUser, err := uc.userRepo.GetByID(ctx, in.TargetUserID)
	if err != nil {
		return "", err
	}
	if targetUser == nil {
		return "", domainErrors.ErrUserNotFound
	}

	// 2. Validate file size
	if in.FileSize > MaxAvatarFileSize {
		return "", domainErrors.ErrFileTooLarge.WithMessage("La fotografía excede el tamaño máximo permitido de 5 MB.")
	}

	// 3. Inspect magic bytes
	headerBuf := make([]byte, 512)
	n, readErr := io.ReadFull(in.File, headerBuf)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", domainErrors.ErrInvalidInput.WithMessage("No se pudo leer la imagen proporcionada.")
	}
	headerBuf = headerBuf[:n]

	mimeType, ext, err := ValidateImageMagicBytes(headerBuf)
	if err != nil {
		return "", err
	}

	// Reconstruct full stream
	fullReader := io.MultiReader(bytes.NewReader(headerBuf), in.File)

	// 4. Generate sanitized storage key: avatars/{user_id}/{random_uuid}.{ext}
	key := fmt.Sprintf("avatars/%s/%s.%s", in.TargetUserID.String(), uuid.New().String(), ext)

	// 5. Persist to storage
	if err := uc.storageService.Upload(ctx, key, fullReader, in.FileSize, mimeType); err != nil {
		return "", fmt.Errorf("usecase/user: failed to upload avatar to storage: %w", err)
	}

	// 6. Delete previous avatar if existed (best effort)
	if targetUser.AvatarR2Key != nil && *targetUser.AvatarR2Key != "" && *targetUser.AvatarR2Key != key {
		_ = uc.storageService.Delete(ctx, *targetUser.AvatarR2Key)
	}

	// 7. Update user record
	if err := uc.userRepo.UpdateAvatar(ctx, in.TargetUserID, key); err != nil {
		return "", fmt.Errorf("usecase/user: failed to update user avatar key: %w", err)
	}

	return key, nil
}
