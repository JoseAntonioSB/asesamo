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
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

type accessPointRepository struct {
	db *sql.DB
}

// NewAccessPointRepository creates a PostgreSQL repository for AccessPoint entities.
func NewAccessPointRepository(db *sql.DB) repository.AccessPointRepository {
	return &accessPointRepository{db: db}
}

const accessPointBaseQuery = `
	SELECT id, plant_id, tenant_id, name, description, is_active, created_at, updated_at, deleted_at
	FROM access_points
`

func (r *accessPointRepository) scanAccessPoint(row interface{ Scan(dest ...any) error }) (*entity.AccessPoint, error) {
	var (
		ap          entity.AccessPoint
		description *string
		deletedAt   *time.Time
	)

	err := row.Scan(
		&ap.ID,
		&ap.PlantID,
		&ap.TenantID,
		&ap.Name,
		&description,
		&ap.IsActive,
		&ap.CreatedAt,
		&ap.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrAccessPointNotFound
		}
		return nil, fmt.Errorf("postgres/access_point: failed to scan access point: %w", err)
	}

	ap.Description = description
	ap.DeletedAt = deletedAt
	return &ap, nil
}

func (r *accessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}
	query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, query, tenantID, id)
	return r.scanAccessPoint(row)
}

func (r *accessPointRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/access_point: query failed: %w", err)
	}
	defer rows.Close()

	points := make([]entity.AccessPoint, 0)
	for rows.Next() {
		ap, err := r.scanAccessPoint(rows)
		if err != nil {
			return nil, err
		}
		points = append(points, *ap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/access_point: rows iteration error: %w", err)
	}

	return points, nil
}

func (r *accessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}
	query := accessPointBaseQuery + ` WHERE tenant_id = $1 AND plant_id = $2 AND is_active = true AND deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/access_point: query failed: %w", err)
	}
	defer rows.Close()

	points := make([]entity.AccessPoint, 0)
	for rows.Next() {
		ap, err := r.scanAccessPoint(rows)
		if err != nil {
			return nil, err
		}
		points = append(points, *ap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/access_point: rows iteration error: %w", err)
	}

	return points, nil
}
