package handler

import (
	"encoding/json"
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/user"
	"asesamo-api/pkg/tenant"
)

// UserHandler handles user-related endpoints.
type UserHandler struct {
	updateThemeUseCase *user.UpdateThemeUseCase
}

// NewUserHandler instantiates a UserHandler.
func NewUserHandler(updateThemeUseCase *user.UpdateThemeUseCase) *UserHandler {
	return &UserHandler{
		updateThemeUseCase: updateThemeUseCase,
	}
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
