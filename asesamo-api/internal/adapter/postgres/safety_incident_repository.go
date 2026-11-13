package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type SafetyIncidentRepository struct {
	db *sql.DB
}

func NewSafetyIncidentRepository(db *sql.DB) *SafetyIncidentRepository {
	return &SafetyIncidentRepository{db: db}
}

func (r *SafetyIncidentRepository) CreateIncident(ctx context.Context, incident *entity.SafetyIncident) error {
	query := `
		INSERT INTO safety_incidents (
			id, tenant_id, incident_number, incident_type, severity, status,
			title, description, location, event_timestamp, reported_by_user_id,
			involved_worker_id, contractor_id, days_lost, evidence_photo_url,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4::incident_type, $5::incident_severity, $6::incident_status,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		incident.ID, incident.TenantID, incident.IncidentNumber, string(incident.IncidentType), string(incident.Severity), string(incident.Status),
		incident.Title, incident.Description, incident.Location, incident.EventTimestamp, incident.ReportedByUserID,
		incident.InvolvedWorkerID, incident.ContractorID, incident.DaysLost, incident.EvidencePhotoURL,
		incident.CreatedAt, incident.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al registrar incidente: %w", err)
	}
	return nil
}

func (r *SafetyIncidentRepository) GetIncidentByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyIncident, error) {
	query := `
		SELECT id, tenant_id, incident_number, incident_type, severity, status,
		       title, description, location, event_timestamp, reported_by_user_id,
		       involved_worker_id, contractor_id, days_lost, evidence_photo_url,
		       root_cause_summary, immediate_cause, basic_cause,
		       investigated_by_user_id, investigated_at, closed_at,
		       created_at, updated_at
		FROM safety_incidents
		WHERE tenant_id = $1 AND id = $2
	`
	inc := &entity.SafetyIncident{}
	var photo, rootCause, immCause, basCause sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&inc.ID, &inc.TenantID, &inc.IncidentNumber, &inc.IncidentType, &inc.Severity, &inc.Status,
		&inc.Title, &inc.Description, &inc.Location, &inc.EventTimestamp, &inc.ReportedByUserID,
		&inc.InvolvedWorkerID, &inc.ContractorID, &inc.DaysLost, &photo,
		&rootCause, &immCause, &basCause,
		&inc.InvestigatedByUserID, &inc.InvestigatedAt, &inc.ClosedAt,
		&inc.CreatedAt, &inc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener incidente: %w", err)
	}
	inc.EvidencePhotoURL = photo.String
	inc.RootCauseSummary = rootCause.String
	inc.ImmediateCause = immCause.String
	inc.BasicCause = basCause.String

	return inc, nil
}

func (r *SafetyIncidentRepository) ListIncidents(ctx context.Context, tenantID uuid.UUID, filter repository.IncidentFilter) ([]entity.SafetyIncident, int, error) {
	baseQuery := `
		SELECT id, tenant_id, incident_number, incident_type, severity, status,
		       title, description, location, event_timestamp, reported_by_user_id,
		       involved_worker_id, contractor_id, days_lost, evidence_photo_url,
		       root_cause_summary, immediate_cause, basic_cause,
		       investigated_by_user_id, investigated_at, closed_at,
		       created_at, updated_at
		FROM safety_incidents
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Type != nil {
		baseQuery += fmt.Sprintf(" AND incident_type = $%d::incident_type", argIdx)
		args = append(args, string(*filter.Type))
		argIdx++
	}
	if filter.Severity != nil {
		baseQuery += fmt.Sprintf(" AND severity = $%d::incident_severity", argIdx)
		args = append(args, string(*filter.Severity))
		argIdx++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d::incident_status", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.ContractorID != nil {
		baseQuery += fmt.Sprintf(" AND contractor_id = $%d", argIdx)
		args = append(args, *filter.ContractorID)
		argIdx++
	}

	baseQuery += " ORDER BY event_timestamp DESC"
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar incidentes: %w", err)
	}
	defer rows.Close()

	var list []entity.SafetyIncident
	for rows.Next() {
		var inc entity.SafetyIncident
		var photo, rootCause, immCause, basCause sql.NullString
		if err := rows.Scan(
			&inc.ID, &inc.TenantID, &inc.IncidentNumber, &inc.IncidentType, &inc.Severity, &inc.Status,
			&inc.Title, &inc.Description, &inc.Location, &inc.EventTimestamp, &inc.ReportedByUserID,
			&inc.InvolvedWorkerID, &inc.ContractorID, &inc.DaysLost, &photo,
			&rootCause, &immCause, &basCause,
			&inc.InvestigatedByUserID, &inc.InvestigatedAt, &inc.ClosedAt,
			&inc.CreatedAt, &inc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		inc.EvidencePhotoURL = photo.String
		inc.RootCauseSummary = rootCause.String
		inc.ImmediateCause = immCause.String
		inc.BasicCause = basCause.String
		list = append(list, inc)
	}

	return list, len(list), nil
}

func (r *SafetyIncidentRepository) UpdateInvestigation(ctx context.Context, incident *entity.SafetyIncident) error {
	query := `
		UPDATE safety_incidents
		SET status = $1::incident_status,
		    root_cause_summary = $2,
		    immediate_cause = $3,
		    basic_cause = $4,
		    days_lost = $5,
		    investigated_by_user_id = $6,
		    investigated_at = $7,
		    closed_at = $8,
		    updated_at = NOW()
		WHERE tenant_id = $9 AND id = $10
	`
	_, err := r.db.ExecContext(ctx, query,
		string(incident.Status), incident.RootCauseSummary, incident.ImmediateCause, incident.BasicCause,
		incident.DaysLost, incident.InvestigatedByUserID, incident.InvestigatedAt, incident.ClosedAt,
		incident.TenantID, incident.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar investigación: %w", err)
	}
	return nil
}

func (r *SafetyIncidentRepository) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*repository.SafetyIncidentMetrics, error) {
	query := `
		SELECT COUNT(*),
		       COALESCE(COUNT(*) FILTER (WHERE incident_type = 'near_miss'), 0),
		       COALESCE(COUNT(*) FILTER (WHERE incident_type = 'unsafe_condition'), 0),
		       COALESCE(COUNT(*) FILTER (WHERE incident_type = 'unsafe_act'), 0),
		       COALESCE(COUNT(*) FILTER (WHERE incident_type = 'lost_time_injury'), 0),
		       COALESCE(SUM(days_lost), 0)
		FROM safety_incidents
		WHERE tenant_id = $1
	`
	m := &repository.SafetyIncidentMetrics{}
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&m.TotalIncidents, &m.NearMisses, &m.UnsafeConditions, &m.UnsafeActs,
		&m.LostTimeInjuries, &m.TotalDaysLost,
	)
	if err != nil {
		return nil, fmt.Errorf("error al consultar métricas de incidentes: %w", err)
	}
	return m, nil
}
