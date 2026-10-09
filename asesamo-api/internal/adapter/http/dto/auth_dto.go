package dto

import "github.com/google/uuid"

// LoginRequest represents incoming credentials for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserProfileResponse represents sanitized user information returned upon login.
type UserProfileResponse struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Role            string     `json:"role"`
	ThemePreference string     `json:"theme_preference"`
}

// LoginResponseData represents data payload returned upon successful login.
type LoginResponseData struct {
	Token string              `json:"token"`
	User  UserProfileResponse `json:"user"`
}

// UpdateThemeRequest represents payload for PATCH /api/v1/users/me/theme.
type UpdateThemeRequest struct {
	Theme string `json:"theme"`
}

