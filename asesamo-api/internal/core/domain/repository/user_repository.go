package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// UserRepository defines the persistence boundary for User entities.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error
	UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error
	UpdateProfile(ctx context.Context, id uuid.UUID, fullName string, phone *string) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarKey string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

