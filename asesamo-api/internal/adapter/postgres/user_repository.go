// Package postgres provides PostgreSQL implementations of domain repositories.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a PostgreSQL repository for User entities.
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

const userBaseQuery = `
	SELECT id, tenant_id, email, password_hash, full_name, phone, role, avatar_r2_key, is_active, last_login_at, created_at, updated_at, deleted_at, COALESCE(theme_preference, 'system')
	FROM users
`

func (r *userRepository) scanUser(row interface{ Scan(dest ...any) error }) (*entity.User, error) {
	var (
		u               entity.User
		tenantID        *uuid.UUID
		phone           *string
		avatarR2Key     *string
		lastLoginAt     *time.Time
		deletedAt       *time.Time
		roleStr         string
		themePreference string
	)

	err := row.Scan(
		&u.ID,
		&tenantID,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&phone,
		&roleStr,
		&avatarR2Key,
		&u.IsActive,
		&lastLoginAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&deletedAt,
		&themePreference,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("postgres/user: failed to scan user: %w", err)
	}

	u.TenantID = tenantID
	u.Phone = phone
	u.AvatarR2Key = avatarR2Key
	u.ThemePreference = themePreference
	u.LastLoginAt = lastLoginAt
	u.DeletedAt = deletedAt
	u.Role = entity.UserRole(roleStr)

	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := userBaseQuery + ` WHERE id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanUser(row)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := userBaseQuery + ` WHERE email = $1 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanUser(row)
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	query := `UPDATE users SET last_login_at = $1, updated_at = now() WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, lastLoginAt, id)
	if err != nil {
		return fmt.Errorf("postgres/user: failed to update last login: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	query := `UPDATE users SET theme_preference = $1, updated_at = now() WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, theme, id)
	if err != nil {
		return fmt.Errorf("postgres/user: failed to update theme preference: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return domainErrors.ErrUserNotFound
	}
	return nil
}

