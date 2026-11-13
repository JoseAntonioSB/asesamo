package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type EmergencyEvacuationRepository struct {
	db *sql.DB
}

func NewEmergencyEvacuationRepository(db *sql.DB) *EmergencyEvacuationRepository {
	return &EmergencyEvacuationRepository{db: db}
}

func (r *EmergencyEvacuationRepository) CreateMusterPoint(ctx context.Context, point *entity.MusterPoint) error {
	query := `
		INSERT INTO muster_points (id, tenant_id, code, name, location_description, capacity, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		point.ID, point.TenantID, point.Code, point.Name, point.LocationDescription, point.Capacity, point.IsActive, point.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al registrar punto de reunión: %w", err)
	}
	return nil
}

func (r *EmergencyEvacuationRepository) ListMusterPoints(ctx context.Context, tenantID uuid.UUID) ([]entity.MusterPoint, error) {
	query := `
		SELECT id, tenant_id, code, name, location_description, capacity, is_active, created_at
		FROM muster_points
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY code ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar puntos de reunión: %w", err)
	}
	defer rows.Close()

	var list []entity.MusterPoint
	for rows.Next() {
		var p entity.MusterPoint
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Code, &p.Name, &desc, &p.Capacity, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.LocationDescription = desc.String
		list = append(list, p)
	}
	return list, nil
}

func (r *EmergencyEvacuationRepository) StartEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryEvac := `
		INSERT INTO emergency_evacuations (
			id, tenant_id, evacuation_number, evacuation_type, status, description, started_at, initiated_by_user_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4::evacuation_type, $5::evacuation_status, $6, $7, $8, $9, $10
		)
	`
	_, err = tx.ExecContext(ctx, queryEvac,
		evac.ID, evac.TenantID, evac.EvacuationNumber, string(evac.EvacuationType), string(evac.Status),
		evac.Description, evac.StartedAt, evac.InitiatedByUserID, evac.CreatedAt, evac.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al iniciar evento de evacuación: %w", err)
	}

	queryRollCall := `
		INSERT INTO evacuation_roll_call (id, evacuation_id, worker_id, status)
		VALUES ($1, $2, $3, 'missing'::roll_call_status)
	`
	for _, rc := range evac.RollCall {
		_, err = tx.ExecContext(ctx, queryRollCall, rc.ID, evac.ID, rc.WorkerID)
		if err != nil {
			return fmt.Errorf("error al insertar pase de lista: %w", err)
		}
	}

	return tx.Commit()
}

func (r *EmergencyEvacuationRepository) GetActiveEvacuation(ctx context.Context, tenantID uuid.UUID) (*entity.EmergencyEvacuation, error) {
	query := `
		SELECT id, tenant_id, evacuation_number, evacuation_type, status, description, started_at, initiated_by_user_id, created_at, updated_at
		FROM emergency_evacuations
		WHERE tenant_id = $1 AND status = 'active'
		ORDER BY started_at DESC
		LIMIT 1
	`
	evac := &entity.EmergencyEvacuation{}
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&evac.ID, &evac.TenantID, &evac.EvacuationNumber, &evac.EvacuationType, &evac.Status,
		&desc, &evac.StartedAt, &evac.InitiatedByUserID, &evac.CreatedAt, &evac.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener evacuación activa: %w", err)
	}
	evac.Description = desc.String

	// Contadores en vivo
	countQuery := `
		SELECT COUNT(*),
		       COALESCE(COUNT(*) FILTER (WHERE status = 'safe'), 0),
		       COALESCE(COUNT(*) FILTER (WHERE status = 'missing'), 0)
		FROM evacuation_roll_call
		WHERE evacuation_id = $1
	`
	_ = r.db.QueryRowContext(ctx, countQuery, evac.ID).Scan(&evac.TotalHeadcount, &evac.SafeCount, &evac.MissingCount)

	return evac, nil
}

func (r *EmergencyEvacuationRepository) GetEvacuationByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EmergencyEvacuation, error) {
	query := `
		SELECT id, tenant_id, evacuation_number, evacuation_type, status, description, started_at, ended_at, evacuation_time_seconds, initiated_by_user_id, created_at, updated_at
		FROM emergency_evacuations
		WHERE tenant_id = $1 AND id = $2
	`
	evac := &entity.EmergencyEvacuation{}
	var desc sql.NullString
	var timeSec sql.NullInt32
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&evac.ID, &evac.TenantID, &evac.EvacuationNumber, &evac.EvacuationType, &evac.Status,
		&desc, &evac.StartedAt, &evac.EndedAt, &timeSec, &evac.InitiatedByUserID, &evac.CreatedAt, &evac.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener evacuación: %w", err)
	}
	evac.Description = desc.String
	if timeSec.Valid {
		evac.EvacuationTimeSeconds = int(timeSec.Int32)
	}

	return evac, nil
}

func (r *EmergencyEvacuationRepository) ListEvacuations(ctx context.Context, tenantID uuid.UUID) ([]entity.EmergencyEvacuation, error) {
	query := `
		SELECT id, tenant_id, evacuation_number, evacuation_type, status, description, started_at, ended_at, evacuation_time_seconds, initiated_by_user_id, created_at, updated_at
		FROM emergency_evacuations
		WHERE tenant_id = $1
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al listar evacuaciones: %w", err)
	}
	defer rows.Close()

	var list []entity.EmergencyEvacuation
	for rows.Next() {
		var evac entity.EmergencyEvacuation
		var desc sql.NullString
		var timeSec sql.NullInt32
		if err := rows.Scan(
			&evac.ID, &evac.TenantID, &evac.EvacuationNumber, &evac.EvacuationType, &evac.Status,
			&desc, &evac.StartedAt, &evac.EndedAt, &timeSec, &evac.InitiatedByUserID, &evac.CreatedAt, &evac.UpdatedAt,
		); err != nil {
			return nil, err
		}
		evac.Description = desc.String
		if timeSec.Valid {
			evac.EvacuationTimeSeconds = int(timeSec.Int32)
		}
		list = append(list, evac)
	}
	return list, nil
}

func (r *EmergencyEvacuationRepository) UpdateRollCallCheckIn(ctx context.Context, evacID, workerID, musterPointID, checkedInBy uuid.UUID, notes string) error {
	query := `
		UPDATE evacuation_roll_call
		SET status = 'safe'::roll_call_status,
		    muster_point_id = $1,
		    checked_in_at = NOW(),
		    checked_in_by_user_id = $2,
		    notes = $3
		WHERE evacuation_id = $4 AND worker_id = $5
	`
	res, err := r.db.ExecContext(ctx, query, musterPointID, checkedInBy, notes, evacID, workerID)
	if err != nil {
		return fmt.Errorf("error al registrar trabajador a salvo en punto de reunión: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("registro de pase de lista no encontrado")
	}
	return nil
}

func (r *EmergencyEvacuationRepository) EndEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error {
	query := `
		UPDATE emergency_evacuations
		SET status = $1::evacuation_status,
		    ended_at = $2,
		    evacuation_time_seconds = $3,
		    updated_at = NOW()
		WHERE tenant_id = $4 AND id = $5
	`
	_, err := r.db.ExecContext(ctx, query, string(evac.Status), evac.EndedAt, evac.EvacuationTimeSeconds, evac.TenantID, evac.ID)
	if err != nil {
		return fmt.Errorf("error al cerrar evento de evacuación: %w", err)
	}
	return nil
}

func (r *EmergencyEvacuationRepository) GetPresentWorkersInPlant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	// Trabajadores con Check-In en access_logs sin checkout posterior hoy
	query := `
		SELECT DISTINCT worker_id
		FROM access_logs
		WHERE tenant_id = $1
		  AND action = 'check_in'
		  AND timestamp >= CURRENT_DATE
		  AND worker_id NOT IN (
		      SELECT worker_id FROM access_logs
		      WHERE tenant_id = $1 AND action = 'check_out' AND timestamp >= CURRENT_DATE
		  )
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al consultar trabajadores en planta: %w", err)
	}
	defer rows.Close()

	var workerIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			workerIDs = append(workerIDs, id)
		}
	}
	return workerIDs, nil
}
