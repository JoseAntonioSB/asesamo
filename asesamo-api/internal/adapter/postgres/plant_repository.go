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

type plantRepository struct {
	db *sql.DB
}

// NewPlantRepository creates a new PostgreSQL plant repository.
func NewPlantRepository(db *sql.DB) repository.PlantRepository {
	return &plantRepository{db: db}
}

func (r *plantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	query := `
		SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
		FROM plants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		p         entity.Plant
		address   *string
		city      *string
		state     *string
		deletedAt *time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.TenantID,
		&p.Name,
		&address,
		&city,
		&state,
		&p.BlockOnExpiredDocs,
		&p.CreatedAt,
		&p.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrTenantNotFound.WithMessage("Planta no encontrada o inactiva.")
		}
		return nil, fmt.Errorf("postgres/plant: GetByID failed: %w", err)
	}

	p.Address = address
	p.City = city
	p.State = state
	p.DeletedAt = deletedAt
	return &p, nil
}

func (r *plantRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error) {
	query := `
		SELECT id, tenant_id, name, address, city, state, block_on_expired_docs, created_at, updated_at, deleted_at
		FROM plants
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/plant: ListByTenant query failed: %w", err)
	}
	defer rows.Close()

	var plants []entity.Plant
	for rows.Next() {
		var (
			p         entity.Plant
			address   *string
			city      *string
			state     *string
			deletedAt *time.Time
		)

		err := rows.Scan(
			&p.ID,
			&p.TenantID,
			&p.Name,
			&address,
			&city,
			&state,
			&p.BlockOnExpiredDocs,
			&p.CreatedAt,
			&p.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/plant: scan plant failed: %w", err)
		}
		p.Address = address
		p.City = city
		p.State = state
		p.DeletedAt = deletedAt
		plants = append(plants, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/plant: rows iteration error: %w", err)
	}

	return plants, nil
}
