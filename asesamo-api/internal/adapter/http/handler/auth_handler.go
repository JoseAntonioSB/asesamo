package handler

import (
	"encoding/json"
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/auth"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	loginUseCase *auth.LoginUseCase
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(loginUseCase *auth.LoginUseCase) *AuthHandler {
	return &AuthHandler{
		loginUseCase: loginUseCase,
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El cuerpo de la petición es obligatorio."))
		return
	}
	defer r.Body.Close()

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Formato JSON inválido."))
		return
	}

	if req.Email == "" || req.Password == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El correo y la contraseña son requeridos."))
		return
	}

	output, err := h.loginUseCase.Execute(r.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	resp := dto.LoginResponseData{
		Token: output.Token,
		User: dto.UserProfileResponse{
			ID:              output.User.ID,
			TenantID:        output.User.TenantID,
			Email:           output.User.Email,
			FullName:        output.User.FullName,
			Role:            string(output.User.Role),
			ThemePreference: output.User.EffectiveTheme(),
		},
	}

	dto.WriteSuccess(w, http.StatusOK, resp)
}
