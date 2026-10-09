package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the authorized system role for a user.
type UserRole string

const (
	RoleSuperAdmin      UserRole = "super_admin"
	RolePlantAdmin      UserRole = "plant_admin"
	RoleEHSManager      UserRole = "ehs_manager"
	RoleGuard           UserRole = "guard"
	RoleContractorAdmin UserRole = "contractor_admin"
	RoleWorker          UserRole = "worker"
)

// User represents an authenticated system identity.
type User struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     string     `json:"full_name"`
	Phone        *string    `json:"phone,omitempty"`
	Role         UserRole   `json:"role"`
	AvatarR2Key     *string    `json:"avatar_r2_key,omitempty"`
	ThemePreference string     `json:"theme_preference"`
	IsActive        bool       `json:"is_active"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// IsAdmin returns true if the user belongs to an administrative oversight role.
func (u *User) IsAdmin() bool {
	switch u.Role {
	case RoleSuperAdmin, RolePlantAdmin, RoleEHSManager, RoleContractorAdmin:
		return true
	default:
		return false
	}
}

// EffectiveTheme resolves theme preference based on user choice or role default.
// Administrative roles default to "light", operational roles (guard, worker) to "dark".
func (u *User) EffectiveTheme() string {
	if u.ThemePreference == "light" || u.ThemePreference == "dark" {
		return u.ThemePreference
	}
	if u.IsAdmin() {
		return "light"
	}
	return "dark"
}

