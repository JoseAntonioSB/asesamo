package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

type workPermitRepository struct {
	db *sql.DB
}

// NewWorkPermitRepository constructs a new PostgreSQL work permit repository.
func NewWorkPermitRepository(db *sql.DB) repository.WorkPermitRepository {
	return &workPermitRepository{db: db}
}

// NextPermitNumber atomically computes the next sequential permit number for the tenant and year.
func (r *workPermitRepository) NextPermitNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	if tenantID == uuid.Nil {
		return "", domainErrors.ErrTenantNotFound
	}
	prefix := fmt.Sprintf("PTAR-%d-%%", year)
	query := `
		SELECT COALESCE(MAX(
			CASE 
				WHEN permit_number ~ '^PTAR-[0-9]{4}-[0-9]+$' 
				THEN SUBSTRING(permit_number FROM 'PTAR-[0-9]{4}-([0-9]+)')::integer 
				ELSE 0 
			END
		), 0) + 1
		FROM work_permits
		WHERE tenant_id = $1 AND permit_number LIKE $2
	`
	var nextSeq int
	err := r.db.QueryRowContext(ctx, query, tenantID, prefix).Scan(&nextSeq)
	if err != nil {
		return "", fmt.Errorf("postgres/permit: failed to get next permit number: %w", err)
	}
	return fmt.Sprintf("PTAR-%d-%04d", year, nextSeq), nil
}

// Create persists a new work permit record.
func (r *workPermitRepository) Create(ctx context.Context, permit *entity.WorkPermit) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	permit.TenantID = tenantID
	if permit.ID == uuid.Nil {
		permit.ID = uuid.New()
	}
	now := time.Now().UTC()
	if permit.CreatedAt.IsZero() {
		permit.CreatedAt = now
	}
	if permit.UpdatedAt.IsZero() {
		permit.UpdatedAt = now
	}
	if permit.Status == "" {
		permit.Status = entity.PermitStatusDraft
	}

	query := `
		INSERT INTO work_permits (
			id, tenant_id, plant_id, permit_number, permit_type, status,
			contractor_id, work_area, description, hazard_description,
			start_time, end_time, created_by_user_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5::permit_type, $6::permit_status, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`
	_, err = r.db.ExecContext(ctx, query,
		permit.ID, permit.TenantID, permit.PlantID, permit.PermitNumber,
		string(permit.PermitType), string(permit.Status), permit.ContractorID,
		permit.WorkArea, permit.Description, permit.HazardDescription,
		permit.StartTime, permit.EndTime, permit.CreatedByUserID,
		permit.CreatedAt, permit.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to insert work permit: %w", err)
	}
	return nil
}

// GetByID loads a work permit by ID with all its relations (crew, checklist, signatures).
func (r *workPermitRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.WorkPermit, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT 
			wp.id, wp.tenant_id, wp.plant_id, COALESCE(p.name, '') AS plant_name,
			wp.permit_number, wp.permit_type, wp.status,
			wp.contractor_id, COALESCE(c.name, '') AS contractor_name,
			wp.work_area, wp.description, wp.hazard_description,
			wp.start_time, wp.end_time,
			wp.created_by_user_id, COALESCE(u_cr.full_name, '') AS created_by_name,
			wp.approved_by_user_id, COALESCE(u_ap.full_name, '') AS approved_by_name,
			wp.approved_at, wp.suspended_at, wp.closed_at, wp.cancelled_at,
			wp.status_reason, wp.created_at, wp.updated_at, wp.deleted_at
		FROM work_permits wp
		LEFT JOIN plants p ON wp.plant_id = p.id
		LEFT JOIN contractors c ON wp.contractor_id = c.id
		LEFT JOIN users u_cr ON wp.created_by_user_id = u_cr.id
		LEFT JOIN users u_ap ON wp.approved_by_user_id = u_ap.id
		WHERE wp.tenant_id = $1 AND wp.id = $2 AND wp.deleted_at IS NULL
	`

	var (
		wp            entity.WorkPermit
		permitTypeStr string
		statusStr     string
	)

	err = r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&wp.ID, &wp.TenantID, &wp.PlantID, &wp.PlantName,
		&wp.PermitNumber, &permitTypeStr, &statusStr,
		&wp.ContractorID, &wp.ContractorName,
		&wp.WorkArea, &wp.Description, &wp.HazardDescription,
		&wp.StartTime, &wp.EndTime,
		&wp.CreatedByUserID, &wp.CreatedByName,
		&wp.ApprovedByUserID, &wp.ApprovedByName,
		&wp.ApprovedAt, &wp.SuspendedAt, &wp.ClosedAt, &wp.CancelledAt,
		&wp.StatusReason, &wp.CreatedAt, &wp.UpdatedAt, &wp.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrPermitNotFound
		}
		return nil, fmt.Errorf("postgres/permit: failed to scan work permit: %w", err)
	}

	wp.PermitType = entity.PermitType(permitTypeStr)
	wp.Status = entity.PermitStatus(statusStr)

	// Populate workers, checklist, and signatures
	workers, err := r.ListWorkersByPermit(ctx, id)
	if err != nil {
		return nil, err
	}
	wp.Workers = workers

	checklist, err := r.ListChecklistByPermit(ctx, id)
	if err != nil {
		return nil, err
	}
	wp.Checklist = checklist

	signatures, err := r.ListSignaturesByPermit(ctx, id)
	if err != nil {
		return nil, err
	}
	wp.Signatures = signatures

	return &wp, nil
}

// ListWithFilters retrieves permits matching query parameters.
func (r *workPermitRepository) ListWithFilters(ctx context.Context, filter repository.PermitFilter) ([]entity.WorkPermit, int, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, 0, domainErrors.ErrTenantNotFound
	}

	var (
		conditions []string
		args       []any
		argIdx     = 1
	)

	conditions = append(conditions, fmt.Sprintf("wp.tenant_id = $%d", argIdx))
	args = append(args, tenantID)
	argIdx++

	conditions = append(conditions, "wp.deleted_at IS NULL")

	if filter.PlantID != nil && *filter.PlantID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("wp.plant_id = $%d", argIdx))
		args = append(args, *filter.PlantID)
		argIdx++
	}

	if filter.ContractorID != nil && *filter.ContractorID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("wp.contractor_id = $%d", argIdx))
		args = append(args, *filter.ContractorID)
		argIdx++
	}

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("wp.status = $%d::permit_status", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.PermitType != nil && *filter.PermitType != "" {
		conditions = append(conditions, fmt.Sprintf("wp.permit_type = $%d::permit_type", argIdx))
		args = append(args, string(*filter.PermitType))
		argIdx++
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("wp.start_time >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("wp.end_time <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM work_permits wp WHERE %s", whereClause)
	var total int
	err = r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres/permit: count query failed: %w", err)
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	selectQuery := fmt.Sprintf(`
		SELECT 
			wp.id, wp.tenant_id, wp.plant_id, COALESCE(p.name, '') AS plant_name,
			wp.permit_number, wp.permit_type, wp.status,
			wp.contractor_id, COALESCE(c.name, '') AS contractor_name,
			wp.work_area, wp.description, wp.hazard_description,
			wp.start_time, wp.end_time,
			wp.created_by_user_id, COALESCE(u_cr.full_name, '') AS created_by_name,
			wp.approved_by_user_id, COALESCE(u_ap.full_name, '') AS approved_by_name,
			wp.approved_at, wp.suspended_at, wp.closed_at, wp.cancelled_at,
			wp.status_reason, wp.created_at, wp.updated_at, wp.deleted_at
		FROM work_permits wp
		LEFT JOIN plants p ON wp.plant_id = p.id
		LEFT JOIN contractors c ON wp.contractor_id = c.id
		LEFT JOIN users u_cr ON wp.created_by_user_id = u_cr.id
		LEFT JOIN users u_ap ON wp.approved_by_user_id = u_ap.id
		WHERE %s
		ORDER BY wp.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres/permit: select permits query failed: %w", err)
	}
	defer rows.Close()

	var permits []entity.WorkPermit
	for rows.Next() {
		var (
			wp            entity.WorkPermit
			permitTypeStr string
			statusStr     string
		)
		err := rows.Scan(
			&wp.ID, &wp.TenantID, &wp.PlantID, &wp.PlantName,
			&wp.PermitNumber, &permitTypeStr, &statusStr,
			&wp.ContractorID, &wp.ContractorName,
			&wp.WorkArea, &wp.Description, &wp.HazardDescription,
			&wp.StartTime, &wp.EndTime,
			&wp.CreatedByUserID, &wp.CreatedByName,
			&wp.ApprovedByUserID, &wp.ApprovedByName,
			&wp.ApprovedAt, &wp.SuspendedAt, &wp.ClosedAt, &wp.CancelledAt,
			&wp.StatusReason, &wp.CreatedAt, &wp.UpdatedAt, &wp.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres/permit: scan permit row failed: %w", err)
		}
		wp.PermitType = entity.PermitType(permitTypeStr)
		wp.Status = entity.PermitStatus(statusStr)
		permits = append(permits, wp)
	}

	return permits, total, nil
}

// UpdateStatus changes the state of a work permit and records relevant audit timestamps.
func (r *workPermitRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PermitStatus, reason *string, byUserID *uuid.UUID) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	now := time.Now().UTC()

	var (
		approvedAt  *time.Time
		suspendedAt *time.Time
		closedAt    *time.Time
		cancelledAt *time.Time
	)

	switch status {
	case entity.PermitStatusApprovedActive:
		approvedAt = &now
	case entity.PermitStatusSuspended:
		suspendedAt = &now
	case entity.PermitStatusClosed:
		closedAt = &now
	case entity.PermitStatusCancelled:
		cancelledAt = &now
	}

	query := `
		UPDATE work_permits
		SET status = $1::permit_status,
			status_reason = COALESCE($2, status_reason),
			approved_by_user_id = CASE WHEN $1::text = 'approved_active' THEN $3 ELSE approved_by_user_id END,
			approved_at = CASE WHEN $1::text = 'approved_active' THEN $4 ELSE approved_at END,
			suspended_at = CASE WHEN $1::text = 'suspended' THEN $5 ELSE suspended_at END,
			closed_at = CASE WHEN $1::text = 'closed' THEN $6 ELSE closed_at END,
			cancelled_at = CASE WHEN $1::text = 'cancelled' THEN $7 ELSE cancelled_at END,
			updated_at = $8
		WHERE tenant_id = $9 AND id = $10 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query,
		string(status), reason, byUserID, approvedAt, suspendedAt, closedAt, cancelledAt, now, tenantID, id,
	)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to update status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres/permit: check rows affected failed: %w", err)
	}
	if rows == 0 {
		return domainErrors.ErrPermitNotFound
	}
	return nil
}

// AssignWorker inserts or updates a worker assigned to a permit with their compliance snapshot.
func (r *workPermitRepository) AssignWorker(ctx context.Context, worker *entity.PermitWorker) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	worker.TenantID = tenantID
	if worker.ID == uuid.Nil {
		worker.ID = uuid.New()
	}
	now := time.Now().UTC()
	if worker.AssignedAt.IsZero() {
		worker.AssignedAt = now
	}
	if worker.CreatedAt.IsZero() {
		worker.CreatedAt = now
	}
	if worker.ComplianceStatus == "" {
		if worker.ComplianceSnapshot.IsCompliant {
			worker.ComplianceStatus = "compliant"
		} else {
			worker.ComplianceStatus = "non_compliant"
		}
	}

	snapshotJSON, err := json.Marshal(worker.ComplianceSnapshot)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to marshal compliance snapshot: %w", err)
	}

	query := `
		INSERT INTO permit_workers (
			id, tenant_id, permit_id, worker_id, compliance_status,
			compliance_snapshot, assigned_by, assigned_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (permit_id, worker_id) DO UPDATE SET
			compliance_status = EXCLUDED.compliance_status,
			compliance_snapshot = EXCLUDED.compliance_snapshot,
			assigned_by = EXCLUDED.assigned_by,
			assigned_at = EXCLUDED.assigned_at
	`
	_, err = r.db.ExecContext(ctx, query,
		worker.ID, worker.TenantID, worker.PermitID, worker.WorkerID,
		worker.ComplianceStatus, snapshotJSON, worker.AssignedBy,
		worker.AssignedAt, worker.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to assign worker: %w", err)
	}
	return nil
}

// RemoveWorker detaches a worker from a permit.
func (r *workPermitRepository) RemoveWorker(ctx context.Context, permitID, workerID uuid.UUID) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	query := `DELETE FROM permit_workers WHERE tenant_id = $1 AND permit_id = $2 AND worker_id = $3`
	_, err = r.db.ExecContext(ctx, query, tenantID, permitID, workerID)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to remove worker: %w", err)
	}
	return nil
}

// ListWorkersByPermit retrieves all workers assigned to a permit with profile enrichment.
func (r *workPermitRepository) ListWorkersByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitWorker, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT 
			pw.id, pw.tenant_id, pw.permit_id, pw.worker_id,
			pw.compliance_status, pw.compliance_snapshot,
			pw.assigned_by, pw.assigned_at, pw.created_at,
			COALESCE(u.full_name, '') AS full_name,
			u.avatar_r2_key,
			cw.job_title, cw.curp, cw.imss_number
		FROM permit_workers pw
		JOIN users u ON pw.worker_id = u.id
		LEFT JOIN contractor_workers cw ON pw.worker_id = cw.user_id AND cw.deleted_at IS NULL
		WHERE pw.tenant_id = $1 AND pw.permit_id = $2
		ORDER BY pw.assigned_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, permitID)
	if err != nil {
		return nil, fmt.Errorf("postgres/permit: failed to list permit workers: %w", err)
	}
	defer rows.Close()

	var workers []entity.PermitWorker
	for rows.Next() {
		var (
			w            entity.PermitWorker
			snapshotJSON []byte
		)
		err := rows.Scan(
			&w.ID, &w.TenantID, &w.PermitID, &w.WorkerID,
			&w.ComplianceStatus, &snapshotJSON,
			&w.AssignedBy, &w.AssignedAt, &w.CreatedAt,
			&w.FullName, &w.AvatarR2Key,
			&w.JobTitle, &w.CURP, &w.IMSSNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/permit: failed to scan permit worker: %w", err)
		}

		if len(snapshotJSON) > 0 {
			_ = json.Unmarshal(snapshotJSON, &w.ComplianceSnapshot)
		}
		workers = append(workers, w)
	}

	return workers, nil
}

// InitChecklistItems initializes the checklist items for a new permit.
func (r *workPermitRepository) InitChecklistItems(ctx context.Context, permitID uuid.UUID, items []entity.PermitChecklistItem) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	if len(items) == 0 {
		return nil
	}

	now := time.Now().UTC()
	query := `
		INSERT INTO permit_checklists (
			id, tenant_id, permit_id, item_key, nom_standard, title,
			description, category, is_mandatory, is_verified, display_order,
			measurements, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (permit_id, item_key) DO NOTHING
	`

	for _, item := range items {
		itemID := item.ID
		if itemID == uuid.Nil {
			itemID = uuid.New()
		}
		measurementsJSON, _ := json.Marshal(item.Measurements)
		if len(measurementsJSON) == 0 || string(measurementsJSON) == "null" {
			measurementsJSON = []byte("{}")
		}

		_, err := r.db.ExecContext(ctx, query,
			itemID, tenantID, permitID, item.ItemKey, item.NOMStandard, item.Title,
			item.Description, item.Category, item.IsMandatory, item.IsVerified, item.DisplayOrder,
			measurementsJSON, now, now,
		)
		if err != nil {
			return fmt.Errorf("postgres/permit: failed to insert checklist item %s: %w", item.ItemKey, err)
		}
	}

	return nil
}

// ListChecklistByPermit retrieves all checklist items for a permit.
func (r *workPermitRepository) ListChecklistByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitChecklistItem, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT 
			pch.id, pch.tenant_id, pch.permit_id, pch.item_key, pch.nom_standard,
			pch.title, pch.description, pch.category, pch.is_mandatory, pch.is_verified,
			pch.verified_by, COALESCE(u.full_name, '') AS verified_by_name,
			pch.verified_at, pch.measurements, pch.notes, pch.display_order,
			pch.created_at, pch.updated_at
		FROM permit_checklists pch
		LEFT JOIN users u ON pch.verified_by = u.id
		WHERE pch.tenant_id = $1 AND pch.permit_id = $2
		ORDER BY pch.display_order ASC, pch.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, permitID)
	if err != nil {
		return nil, fmt.Errorf("postgres/permit: failed to list checklist items: %w", err)
	}
	defer rows.Close()

	var items []entity.PermitChecklistItem
	for rows.Next() {
		var (
			it               entity.PermitChecklistItem
			measurementsJSON []byte
		)
		err := rows.Scan(
			&it.ID, &it.TenantID, &it.PermitID, &it.ItemKey, &it.NOMStandard,
			&it.Title, &it.Description, &it.Category, &it.IsMandatory, &it.IsVerified,
			&it.VerifiedBy, &it.VerifiedByName,
			&it.VerifiedAt, &measurementsJSON, &it.Notes, &it.DisplayOrder,
			&it.CreatedAt, &it.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/permit: failed to scan checklist item: %w", err)
		}

		if len(measurementsJSON) > 0 {
			_ = json.Unmarshal(measurementsJSON, &it.Measurements)
		}
		items = append(items, it)
	}

	return items, nil
}

// UpdateChecklistItem updates verification status and measurements for a checklist item.
func (r *workPermitRepository) UpdateChecklistItem(
	ctx context.Context,
	permitID uuid.UUID,
	itemKey string,
	isVerified bool,
	verifiedBy *uuid.UUID,
	notes *string,
	measurements map[string]any,
) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	now := time.Now().UTC()
	var verifiedAt *time.Time
	if isVerified {
		verifiedAt = &now
	}

	measurementsJSON, err := json.Marshal(measurements)
	if err != nil || len(measurementsJSON) == 0 || string(measurementsJSON) == "null" {
		measurementsJSON = []byte("{}")
	}

	query := `
		UPDATE permit_checklists
		SET is_verified = $1, verified_by = $2, verified_at = $3,
			notes = COALESCE($4, notes), measurements = $5, updated_at = $6
		WHERE tenant_id = $7 AND permit_id = $8 AND item_key = $9
	`
	_, err = r.db.ExecContext(ctx, query,
		isVerified, verifiedBy, verifiedAt, notes, measurementsJSON, now, tenantID, permitID, itemKey,
	)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to update checklist item: %w", err)
	}
	return nil
}

// AddSignature inserts or updates a digital sign-off record for a permit.
func (r *workPermitRepository) AddSignature(ctx context.Context, sig *entity.PermitSignature) error {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return domainErrors.ErrTenantNotFound
	}
	sig.TenantID = tenantID
	if sig.ID == uuid.Nil {
		sig.ID = uuid.New()
	}
	now := time.Now().UTC()
	if sig.SignedAt.IsZero() {
		sig.SignedAt = now
	}
	if sig.CreatedAt.IsZero() {
		sig.CreatedAt = now
	}

	query := `
		INSERT INTO permit_signatures (
			id, tenant_id, permit_id, user_id, role,
			signature_data, ip_address, user_agent, notes, signed_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5::permit_signature_role, $6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (permit_id, role) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			signature_data = EXCLUDED.signature_data,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent,
			notes = EXCLUDED.notes,
			signed_at = EXCLUDED.signed_at
	`
	_, err = r.db.ExecContext(ctx, query,
		sig.ID, sig.TenantID, sig.PermitID, sig.UserID, string(sig.Role),
		sig.SignatureData, sig.IPAddress, sig.UserAgent, sig.Notes,
		sig.SignedAt, sig.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/permit: failed to insert permit signature: %w", err)
	}
	return nil
}

// ListSignaturesByPermit retrieves all signatures logged for a permit.
func (r *workPermitRepository) ListSignaturesByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitSignature, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, domainErrors.ErrTenantNotFound
	}

	query := `
		SELECT 
			ps.id, ps.tenant_id, ps.permit_id, ps.user_id, ps.role,
			ps.signature_data, ps.ip_address, ps.user_agent, ps.notes,
			ps.signed_at, ps.created_at,
			COALESCE(u.full_name, '') AS signer_full_name
		FROM permit_signatures ps
		JOIN users u ON ps.user_id = u.id
		WHERE ps.tenant_id = $1 AND ps.permit_id = $2
		ORDER BY ps.signed_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, permitID)
	if err != nil {
		return nil, fmt.Errorf("postgres/permit: failed to list signatures: %w", err)
	}
	defer rows.Close()

	var signatures []entity.PermitSignature
	for rows.Next() {
		var (
			s       entity.PermitSignature
			roleStr string
		)
		err := rows.Scan(
			&s.ID, &s.TenantID, &s.PermitID, &s.UserID, &roleStr,
			&s.SignatureData, &s.IPAddress, &s.UserAgent, &s.Notes,
			&s.SignedAt, &s.CreatedAt, &s.SignerFullName,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/permit: failed to scan signature row: %w", err)
		}
		s.Role = entity.PermitSignatureRole(roleStr)
		signatures = append(signatures, s)
	}

	return signatures, nil
}
