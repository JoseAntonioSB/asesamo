package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type SafetyInspectionRepository struct {
	db *sql.DB
}

func NewSafetyInspectionRepository(db *sql.DB) *SafetyInspectionRepository {
	return &SafetyInspectionRepository{db: db}
}

func (r *SafetyInspectionRepository) CreateTemplate(ctx context.Context, template *entity.InspectionTemplate, items []entity.InspectionTemplateItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryTemplate := `
		INSERT INTO inspection_templates (id, tenant_id, title, description, category, normative_reference, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.ExecContext(ctx, queryTemplate,
		template.ID, template.TenantID, template.Title, template.Description, template.Category, template.NormativeReference, template.IsActive, template.CreatedAt, template.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al insertar plantilla de inspección: %w", err)
	}

	queryItem := `
		INSERT INTO inspection_template_items (id, template_id, category, question, guidance_notes, weight, critical_risk, sort_order, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, queryItem,
			item.ID, template.ID, item.Category, item.Question, item.GuidanceNotes, item.Weight, item.CriticalRisk, item.SortOrder, item.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("error al insertar item de plantilla: %w", err)
		}
	}

	return tx.Commit()
}

func (r *SafetyInspectionRepository) GetTemplateByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.InspectionTemplate, []entity.InspectionTemplateItem, error) {
	query := `
		SELECT id, tenant_id, title, description, category, normative_reference, is_active, created_at, updated_at
		FROM inspection_templates
		WHERE tenant_id = $1 AND id = $2
	`
	tmpl := &entity.InspectionTemplate{}
	var desc, normRef sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&tmpl.ID, &tmpl.TenantID, &tmpl.Title, &desc, &tmpl.Category, &normRef, &tmpl.IsActive, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("error al obtener plantilla: %w", err)
	}
	tmpl.Description = desc.String
	tmpl.NormativeReference = normRef.String

	itemsQuery := `
		SELECT id, template_id, category, question, guidance_notes, weight, critical_risk, sort_order, created_at
		FROM inspection_template_items
		WHERE template_id = $1
		ORDER BY sort_order ASC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, nil, fmt.Errorf("error al obtener items de plantilla: %w", err)
	}
	defer rows.Close()

	var items []entity.InspectionTemplateItem
	for rows.Next() {
		var itm entity.InspectionTemplateItem
		var guide sql.NullString
		if err := rows.Scan(&itm.ID, &itm.TemplateID, &itm.Category, &itm.Question, &guide, &itm.Weight, &itm.CriticalRisk, &itm.SortOrder, &itm.CreatedAt); err != nil {
			return nil, nil, err
		}
		itm.GuidanceNotes = guide.String
		items = append(items, itm)
	}

	return tmpl, items, nil
}

func (r *SafetyInspectionRepository) ListTemplates(ctx context.Context, tenantID uuid.UUID) ([]entity.InspectionTemplate, error) {
	query := `
		SELECT id, tenant_id, title, description, category, normative_reference, is_active, created_at, updated_at
		FROM inspection_templates
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY title ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al listar plantillas: %w", err)
	}
	defer rows.Close()

	var list []entity.InspectionTemplate
	for rows.Next() {
		var tmpl entity.InspectionTemplate
		var desc, normRef sql.NullString
		if err := rows.Scan(&tmpl.ID, &tmpl.TenantID, &tmpl.Title, &desc, &tmpl.Category, &normRef, &tmpl.IsActive, &tmpl.CreatedAt, &tmpl.UpdatedAt); err != nil {
			return nil, err
		}
		tmpl.Description = desc.String
		tmpl.NormativeReference = normRef.String
		list = append(list, tmpl)
	}
	return list, nil
}

func (r *SafetyInspectionRepository) CreateInspection(ctx context.Context, inspection *entity.SafetyInspection) error {
	query := `
		INSERT INTO safety_inspections (
			id, tenant_id, template_id, inspection_number, inspection_type, status,
			area_or_location, contractor_id, inspector_id, notes, started_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5::inspection_type, $6::inspection_status,
			$7, $8, $9, $10, $11, $12, $13
		)
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		inspection.ID, inspection.TenantID, inspection.TemplateID, inspection.InspectionNumber,
		string(inspection.InspectionType), string(inspection.Status), inspection.AreaOrLocation,
		inspection.ContractorID, inspection.InspectorID, inspection.Notes, now, inspection.CreatedAt, inspection.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear inspección: %w", err)
	}
	return nil
}

func (r *SafetyInspectionRepository) GetInspectionByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyInspection, error) {
	query := `
		SELECT id, tenant_id, template_id, inspection_number, inspection_type, status,
		       area_or_location, contractor_id, inspector_id, score_percentage,
		       total_items, compliant_items, non_compliant_items, na_items, notes,
		       started_at, completed_at, created_at, updated_at
		FROM safety_inspections
		WHERE tenant_id = $1 AND id = $2
	`
	insp := &entity.SafetyInspection{}
	var score sql.NullFloat64
	var notes sql.NullString
	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&insp.ID, &insp.TenantID, &insp.TemplateID, &insp.InspectionNumber, &insp.InspectionType, &insp.Status,
		&insp.AreaOrLocation, &insp.ContractorID, &insp.InspectorID, &score,
		&insp.TotalItems, &insp.CompliantItems, &insp.NonCompliantItems, &insp.NAItems, &notes,
		&insp.StartedAt, &insp.CompletedAt, &insp.CreatedAt, &insp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener inspección: %w", err)
	}

	if score.Valid {
		insp.ScorePercentage = &score.Float64
	}
	insp.Notes = notes.String

	// Cargar respuestas si existen
	ansQuery := `
		SELECT a.template_item_id, a.status, ti.weight, ti.critical_risk, a.finding_notes, a.evidence_photo_url
		FROM safety_inspection_items a
		JOIN inspection_template_items ti ON ti.id = a.template_item_id
		WHERE a.inspection_id = $1
	`
	rows, err := r.db.QueryContext(ctx, ansQuery, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ans entity.InspectionItemAnswer
			var fNotes, ePhoto sql.NullString
			if err := rows.Scan(&ans.TemplateItemID, &ans.Status, &ans.Weight, &ans.CriticalRisk, &fNotes, &ePhoto); err == nil {
				ans.FindingNotes = fNotes.String
				ans.EvidencePhotoURL = ePhoto.String
				insp.Answers = append(insp.Answers, ans)
			}
		}
	}

	return insp, nil
}

func (r *SafetyInspectionRepository) ListInspections(ctx context.Context, tenantID uuid.UUID, filter repository.InspectionFilter) ([]entity.SafetyInspection, int, error) {
	baseQuery := `
		SELECT id, tenant_id, template_id, inspection_number, inspection_type, status,
		       area_or_location, contractor_id, inspector_id, score_percentage,
		       total_items, compliant_items, non_compliant_items, na_items, notes,
		       started_at, completed_at, created_at, updated_at
		FROM safety_inspections
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND status = $%d::inspection_status", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.ContractorID != nil {
		baseQuery += fmt.Sprintf(" AND contractor_id = $%d", argIdx)
		args = append(args, *filter.ContractorID)
		argIdx++
	}
	if filter.InspectorID != nil {
		baseQuery += fmt.Sprintf(" AND inspector_id = $%d", argIdx)
		args = append(args, *filter.InspectorID)
		argIdx++
	}

	baseQuery += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		baseQuery += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar inspecciones: %w", err)
	}
	defer rows.Close()

	var list []entity.SafetyInspection
	for rows.Next() {
		var insp entity.SafetyInspection
		var score sql.NullFloat64
		var notes sql.NullString
		if err := rows.Scan(
			&insp.ID, &insp.TenantID, &insp.TemplateID, &insp.InspectionNumber, &insp.InspectionType, &insp.Status,
			&insp.AreaOrLocation, &insp.ContractorID, &insp.InspectorID, &score,
			&insp.TotalItems, &insp.CompliantItems, &insp.NonCompliantItems, &insp.NAItems, &notes,
			&insp.StartedAt, &insp.CompletedAt, &insp.CreatedAt, &insp.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if score.Valid {
			insp.ScorePercentage = &score.Float64
		}
		insp.Notes = notes.String
		list = append(list, insp)
	}

	return list, len(list), nil
}

func (r *SafetyInspectionRepository) SaveInspectionAnswers(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []entity.InspectionItemAnswer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Validar que pertenezca al tenant
	var count int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM safety_inspections WHERE tenant_id = $1 AND id = $2`, tenantID, inspectionID).Scan(&count)
	if err != nil || count == 0 {
		return errors.New("inspección no encontrada o no pertenece al tenant")
	}

	upsertQuery := `
		INSERT INTO safety_inspection_items (id, inspection_id, template_item_id, status, finding_notes, evidence_photo_url, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3::item_compliance_status, $4, $5, NOW())
		ON CONFLICT (inspection_id, template_item_id)
		DO UPDATE SET status = EXCLUDED.status, finding_notes = EXCLUDED.finding_notes, evidence_photo_url = EXCLUDED.evidence_photo_url, updated_at = NOW()
	`
	for _, ans := range answers {
		_, err := tx.ExecContext(ctx, upsertQuery,
			inspectionID, ans.TemplateItemID, string(ans.Status), ans.FindingNotes, ans.EvidencePhotoURL,
		)
		if err != nil {
			return fmt.Errorf("error al guardar respuesta de inspección: %w", err)
		}
	}

	return tx.Commit()
}

func (r *SafetyInspectionRepository) CompleteInspection(ctx context.Context, inspection *entity.SafetyInspection, capas []entity.InspectionActionPlan) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Actualizar cabecera de la inspección
	updateQuery := `
		UPDATE safety_inspections
		SET status = $1::inspection_status,
		    score_percentage = $2,
		    total_items = $3,
		    compliant_items = $4,
		    non_compliant_items = $5,
		    na_items = $6,
		    completed_at = $7,
		    updated_at = $8
		WHERE tenant_id = $9 AND id = $10
	`
	_, err = tx.ExecContext(ctx, updateQuery,
		string(inspection.Status), inspection.ScorePercentage, inspection.TotalItems,
		inspection.CompliantItems, inspection.NonCompliantItems, inspection.NAItems,
		inspection.CompletedAt, inspection.UpdatedAt, inspection.TenantID, inspection.ID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de inspección: %w", err)
	}

	// Insertar acciones correctivas (CAPA) si aplica
	if len(capas) > 0 {
		capaQuery := `
			INSERT INTO inspection_action_plans (
				id, tenant_id, inspection_id, template_item_id, description, severity,
				status, assigned_to_user_id, contractor_id, due_date, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6::finding_severity,
				$7::capa_status, $8, $9, $10, $11, $12
			)
		`
		for _, c := range capas {
			_, err := tx.ExecContext(ctx, capaQuery,
				c.ID, c.TenantID, c.InspectionID, c.TemplateItemID, c.Description, string(c.Severity),
				string(c.Status), c.AssignedToUserID, c.ContractorID, c.DueDate, c.CreatedAt, c.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("error al insertar acción correctiva: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *SafetyInspectionRepository) ListActionPlans(ctx context.Context, tenantID uuid.UUID, inspectionID *uuid.UUID, status *entity.CAPAStatus) ([]entity.InspectionActionPlan, error) {
	query := `
		SELECT id, tenant_id, inspection_id, template_item_id, description, severity, status,
		       assigned_to_user_id, contractor_id, due_date, resolved_at, verified_at,
		       resolution_notes, resolution_photo_url, created_at, updated_at
		FROM inspection_action_plans
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if inspectionID != nil {
		query += fmt.Sprintf(" AND inspection_id = $%d", argIdx)
		args = append(args, *inspectionID)
		argIdx++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d::capa_status", argIdx)
		args = append(args, string(*status))
		argIdx++
	}

	query += " ORDER BY due_date ASC, created_at DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al consultar acciones correctivas: %w", err)
	}
	defer rows.Close()

	var list []entity.InspectionActionPlan
	for rows.Next() {
		var c entity.InspectionActionPlan
		var rNotes, rPhoto sql.NullString
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.InspectionID, &c.TemplateItemID, &c.Description, &c.Severity, &c.Status,
			&c.AssignedToUserID, &c.ContractorID, &c.DueDate, &c.ResolvedAt, &c.VerifiedAt,
			&rNotes, &rPhoto, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.ResolutionNotes = rNotes.String
		c.ResolutionPhotoURL = rPhoto.String
		list = append(list, c)
	}

	return list, nil
}

func (r *SafetyInspectionRepository) UpdateActionPlanStatus(ctx context.Context, tenantID, capaID uuid.UUID, status entity.CAPAStatus, notes string) error {
	query := `
		UPDATE inspection_action_plans
		SET status = $1::capa_status,
		    resolution_notes = CASE WHEN $2 != '' THEN $2 ELSE resolution_notes END,
		    resolved_at = CASE WHEN $1::capa_status = 'resolved' THEN NOW() ELSE resolved_at END,
		    verified_at = CASE WHEN $1::capa_status = 'verified' THEN NOW() ELSE verified_at END,
		    updated_at = NOW()
		WHERE tenant_id = $3 AND id = $4
	`
	res, err := r.db.ExecContext(ctx, query, string(status), notes, tenantID, capaID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de CAPA: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("acción correctiva no encontrada")
	}
	return nil
}
