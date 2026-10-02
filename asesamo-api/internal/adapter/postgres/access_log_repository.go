package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

type accessLogRepository struct {
	db *sql.DB
}

// NewAccessLogRepository creates a PostgreSQL repository for AccessLog entities.
func NewAccessLogRepository(db *sql.DB) repository.AccessLogRepository {
	return &accessLogRepository{db: db}
}

func (r *accessLogRepository) Create(ctx context.Context, log *entity.AccessLog) error {
	query := `
		INSERT INTO access_logs (
			id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
			processed_by, checked_in_at, checked_out_at, result, denial_reason,
			document_snapshot, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CheckedInAt.IsZero() {
		log.CheckedInAt = time.Now().UTC()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = log.CheckedInAt
	}

	snapshotBytes, err := json.Marshal(log.DocumentSnapshot)
	if err != nil {
		return fmt.Errorf("postgres/access_log: failed to marshal document snapshot: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		log.ID,
		log.TenantID,
		log.PlantID,
		log.AccessPointID,
		log.UserID,
		log.ContractorID,
		log.ProcessedBy,
		log.CheckedInAt,
		log.CheckedOutAt,
		string(log.Result),
		log.DenialReason,
		string(snapshotBytes),
		log.Notes,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/access_log: failed to insert access log: %w", err)
	}

	return nil
}

func (r *accessLogRepository) scanRow(row interface{ Scan(dest ...any) error }) (*entity.AccessLog, error) {
	var (
		log           entity.AccessLog
		contractorID  *uuid.UUID
		processedBy   *uuid.UUID
		checkedOutAt  *time.Time
		denialReason  *string
		notes         *string
		snapshotRaw   []byte
		resultStr     string
	)

	err := row.Scan(
		&log.ID,
		&log.TenantID,
		&log.PlantID,
		&log.AccessPointID,
		&log.UserID,
		&contractorID,
		&processedBy,
		&log.CheckedInAt,
		&checkedOutAt,
		&resultStr,
		&denialReason,
		&snapshotRaw,
		&notes,
		&log.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrWorkerNotFound.WithMessage("Registro de acceso no encontrado.")
		}
		return nil, fmt.Errorf("postgres/access_log: failed to scan row: %w", err)
	}

	log.ContractorID = contractorID
	log.ProcessedBy = processedBy
	log.CheckedOutAt = checkedOutAt
	log.Result = entity.AccessResult(resultStr)
	log.DenialReason = denialReason
	log.Notes = notes

	if len(snapshotRaw) > 0 {
		var snapshot entity.DocumentSnapshot
		if err := json.Unmarshal(snapshotRaw, &snapshot); err == nil {
			log.DocumentSnapshot = snapshot
		} else {
			log.DocumentSnapshot = entity.DocumentSnapshot{}
		}
	} else {
		log.DocumentSnapshot = entity.DocumentSnapshot{}
	}

	return &log, nil
}

func (r *accessLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	query := `
		SELECT id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		       processed_by, checked_in_at, checked_out_at, result, denial_reason,
		       document_snapshot, notes, created_at
		FROM access_logs
		WHERE id = $1
	`
	return r.scanRow(r.db.QueryRowContext(ctx, query, id))
}

func (r *accessLogRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		       processed_by, checked_in_at, checked_out_at, result, denial_reason,
		       document_snapshot, notes, created_at
		FROM access_logs
		WHERE tenant_id = $1
		ORDER BY checked_in_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres/access_log: ListByTenant query failed: %w", err)
	}
	defer rows.Close()

	var logs []entity.AccessLog
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}

	return logs, nil
}

func (r *accessLogRepository) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	countQuery := `
		SELECT COUNT(*)
		FROM access_logs
		WHERE tenant_id = $1 AND plant_id = $2
	`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID, plantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("postgres/access_log: count query failed: %w", err)
	}

	query := `
		SELECT id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		       processed_by, checked_in_at, checked_out_at, result, denial_reason,
		       document_snapshot, notes, created_at
		FROM access_logs
		WHERE tenant_id = $1 AND plant_id = $2
		ORDER BY checked_in_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres/access_log: ListByPlant query failed: %w", err)
	}
	defer rows.Close()

	var logs []entity.AccessLog
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, *l)
	}

	return logs, total, nil
}

func (r *accessLogRepository) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	query := `
		UPDATE access_logs
		SET checked_out_at = $1
		WHERE id = (
			SELECT id
			FROM access_logs
			WHERE tenant_id = $2 AND plant_id = $3 AND user_id = $4 AND checked_out_at IS NULL
			ORDER BY checked_in_at DESC
			LIMIT 1
		)
		RETURNING id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
		          processed_by, checked_in_at, checked_out_at, result, denial_reason,
		          document_snapshot, notes, created_at
	`

	row := r.db.QueryRowContext(ctx, query, checkedOutAt, tenantID, plantID, workerID)
	l, err := r.scanRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || domainErrors.IsNotFound(err) {
			// If no active check-in with checked_out_at IS NULL exists, find the latest check-in and update it
			fallbackQuery := `
				UPDATE access_logs
				SET checked_out_at = $1
				WHERE id = (
					SELECT id
					FROM access_logs
					WHERE tenant_id = $2 AND plant_id = $3 AND user_id = $4
					ORDER BY checked_in_at DESC
					LIMIT 1
				)
				RETURNING id, tenant_id, plant_id, access_point_id, user_id, contractor_id,
				          processed_by, checked_in_at, checked_out_at, result, denial_reason,
				          document_snapshot, notes, created_at
			`
			fbRow := r.db.QueryRowContext(ctx, fallbackQuery, checkedOutAt, tenantID, plantID, workerID)
			return r.scanRow(fbRow)
		}
		return nil, err
	}
	return l, nil
}
