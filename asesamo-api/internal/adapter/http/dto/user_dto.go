package dto

import (
	"fmt"
	"time"

	"asesamo-api/internal/core/usecase/user"

	"github.com/google/uuid"
)

// UserProfileDetailResponse represents full user profile information with company details.
type UserProfileDetailResponse struct {
	ID              uuid.UUID  `json:"id"`
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Phone           *string    `json:"phone,omitempty"`
	Role            string     `json:"role"`
	AvatarR2Key     *string    `json:"avatar_r2_key,omitempty"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	ThemePreference string     `json:"theme_preference"`
	ContractorID    *uuid.UUID `json:"contractor_id,omitempty"`
	ContractorName  *string    `json:"contractor_name,omitempty"`
	CreatedAt       string     `json:"created_at"`
}

// UpdateProfileRequest represents payload for PATCH /api/v1/users/me.
type UpdateProfileRequest struct {
	FullName        string  `json:"full_name"`
	Phone           *string `json:"phone,omitempty"`
	CurrentPassword string  `json:"current_password,omitempty"`
	NewPassword     string  `json:"new_password,omitempty"`
}

// AvatarUploadResponse represents result returned upon successful photo upload.
type AvatarUploadResponse struct {
	AvatarKey string `json:"avatar_key"`
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
}

// ToUserProfileDetailResponse maps from ProfileDetail domain model.
func ToUserProfileDetailResponse(d *user.ProfileDetail) UserProfileDetailResponse {
	var avatarURL *string
	if d.User.AvatarR2Key != nil && *d.User.AvatarR2Key != "" {
		url := fmt.Sprintf("/api/v1/users/%s/avatar", d.User.ID.String())
		avatarURL = &url
	}

	return UserProfileDetailResponse{
		ID:              d.User.ID,
		TenantID:        d.User.TenantID,
		Email:           d.User.Email,
		FullName:        d.User.FullName,
		Phone:           d.User.Phone,
		Role:            string(d.User.Role),
		AvatarR2Key:     d.User.AvatarR2Key,
		AvatarURL:       avatarURL,
		ThemePreference: d.User.ThemePreference,
		ContractorID:    d.ContractorID,
		ContractorName:  d.ContractorName,
		CreatedAt:       d.User.CreatedAt.Format(time.RFC3339),
	}
}
