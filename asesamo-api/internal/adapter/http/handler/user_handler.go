package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/user"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// UserHandler handles user-related endpoints.
type UserHandler struct {
	getProfileUseCase    *user.GetProfileUseCase
	updateProfileUseCase *user.UpdateProfileUseCase
	uploadAvatarUseCase  *user.UploadAvatarUseCase
	getAvatarUseCase     *user.GetAvatarUseCase
	updateThemeUseCase   *user.UpdateThemeUseCase
}

// NewUserHandler instantiates a UserHandler.
func NewUserHandler(
	getProfileUseCase *user.GetProfileUseCase,
	updateProfileUseCase *user.UpdateProfileUseCase,
	uploadAvatarUseCase *user.UploadAvatarUseCase,
	getAvatarUseCase *user.GetAvatarUseCase,
	updateThemeUseCase *user.UpdateThemeUseCase,
) *UserHandler {
	return &UserHandler{
		getProfileUseCase:    getProfileUseCase,
		updateProfileUseCase: updateProfileUseCase,
		uploadAvatarUseCase:  uploadAvatarUseCase,
		getAvatarUseCase:     getAvatarUseCase,
		updateThemeUseCase:   updateThemeUseCase,
	}
}

// GetProfile handles GET /api/v1/users/me.
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	detail, err := h.getProfileUseCase.Execute(r.Context(), userID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, dto.ToUserProfileDetailResponse(detail))
}

// UpdateProfile handles PATCH /api/v1/users/me.
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de petición JSON inválido."))
		return
	}

	err = h.updateProfileUseCase.Execute(r.Context(), user.UpdateProfileInput{
		UserID:          userID,
		FullName:        req.FullName,
		Phone:           req.Phone,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	detail, err := h.getProfileUseCase.Execute(r.Context(), userID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, dto.ToUserProfileDetailResponse(detail))
}

// UploadMyAvatar handles POST /api/v1/users/me/avatar.
func (h *UserHandler) UploadMyAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	role, _ := tenant.RoleFromContext(r.Context())
	h.handleAvatarUpload(w, r, userID, userID, role)
}

// UploadUserAvatar handles POST /api/v1/users/{id}/avatar.
func (h *UserHandler) UploadUserAvatar(w http.ResponseWriter, r *http.Request) {
	callerID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	callerRole, _ := tenant.RoleFromContext(r.Context())

	targetIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El ID de usuario especificado no es un UUID válido."))
		return
	}

	h.handleAvatarUpload(w, r, targetID, callerID, callerRole)
}

func (h *UserHandler) handleAvatarUpload(w http.ResponseWriter, r *http.Request, targetUserID, callerID uuid.UUID, callerRole string) {
	// Parse multipart form up to 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, user.MaxAvatarFileSize)
	if err := r.ParseMultipartForm(user.MaxAvatarFileSize); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			dto.WriteError(w, domainErrors.ErrFileTooLarge.WithMessage("La imagen excede el límite máximo de 5 MB."))
			return
		}
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Error al procesar el archivo de imagen."))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		// Also check "avatar" field name
		file, header, err = r.FormFile("avatar")
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Debe incluir el archivo de imagen en el campo 'file' o 'avatar'."))
			return
		}
	}
	defer file.Close()

	key, err := h.uploadAvatarUseCase.Execute(r.Context(), user.UploadAvatarInput{
		TargetUserID: targetUserID,
		CallerID:     callerID,
		CallerRole:   callerRole,
		FileName:     header.Filename,
		File:         file,
		FileSize:     header.Size,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	avatarURL := fmt.Sprintf("/api/v1/users/%s/avatar", targetUserID.String())

	dto.WriteSuccess(w, http.StatusOK, dto.AvatarUploadResponse{
		AvatarKey: key,
		AvatarURL: avatarURL,
		Message:   "Fotografía de perfil actualizada exitosamente.",
	})
}

// GetAvatar handles GET /api/v1/users/{id}/avatar.
func (h *UserHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	targetIDStr := chi.URLParam(r, "id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El ID de usuario no es válido."))
		return
	}

	stream, contentType, err := h.getAvatarUseCase.Execute(r.Context(), targetID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400, must-revalidate")
	_, _ = io.Copy(w, stream)
}

// UpdateTheme handles PATCH /api/v1/users/me/theme.
func (h *UserHandler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	if r.Body == nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El cuerpo de la petición es obligatorio."))
		return
	}
	defer r.Body.Close()

	var req dto.UpdateThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Formato JSON inválido."))
		return
	}

	theme, err := h.updateThemeUseCase.Execute(r.Context(), user.UpdateThemeInput{
		UserID: userID,
		Theme:  req.Theme,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, map[string]string{
		"theme": theme,
	})
}
