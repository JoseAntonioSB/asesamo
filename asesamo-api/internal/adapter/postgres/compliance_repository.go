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

type complianceRepository struct {
	db *sql.DB
}

// NewComplianceRepository creates a PostgreSQL repository for compliance and affiliation queries.
func NewComplianceRepository(db *sql.DB) repository.ComplianceRepository {
	return &complianceRepository{db: db}
}

func (r *complianceRepository) GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error) {
	// First, check if the user exists
	userQuery := `
		SELECT id, full_name, avatar_r2_key, is_active
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var (
		aff         entity.WorkerAffiliation
		avatarR2Key *string
	)

	err := r.db.QueryRowContext(ctx, userQuery, workerUserID).Scan(
		&aff.WorkerID,
		&aff.FullName,
		&avatarR2Key,
		&aff.UserIsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrWorkerNotFound
		}
		return nil, fmt.Errorf("postgres/compliance: check user query failed: %w", err)
	}
	aff.AvatarR2Key = avatarR2Key
	aff.IsActive = aff.UserIsActive

	// Next, retrieve contractor affiliation and plant status
	affiliationQuery := `
		SELECT 
			c.id AS contractor_id,
			c.name AS contractor_name,
			c.rfc AS contractor_rfc,
			c.repse_number AS repse_number,
			c.is_active AS contractor_is_active,
			cw.curp,
			cw.imss_number,
			cw.job_title,
			cw.is_active AS worker_is_active,
			COALESCE(pc.status::text, 'pending') AS plant_status
		FROM contractor_workers cw
		JOIN contractors c ON cw.contractor_id = c.id
		LEFT JOIN plant_contractors pc ON pc.contractor_id = c.id AND pc.plant_id = $1
		WHERE cw.user_id = $2 AND cw.deleted_at IS NULL AND c.deleted_at IS NULL
		LIMIT 1
	`

	var (
		rfc         *string
		repseNumber *string
		curp        *string
		imssNumber  *string
		jobTitle    *string
		plantStatus string
	)

	err = r.db.QueryRowContext(ctx, affiliationQuery, plantID, workerUserID).Scan(
		&aff.ContractorID,
		&aff.ContractorName,
		&rfc,
		&repseNumber,
		&aff.ContractorIsActive,
		&curp,
		&imssNumber,
		&jobTitle,
		&aff.WorkerIsActive,
		&plantStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Worker exists but is not affiliated with a contractor in this plant
			aff.PlantStatus = "pending"
			aff.WorkerIsActive = false
			aff.ContractorIsActive = false
			return &aff, nil
		}
		return nil, fmt.Errorf("postgres/compliance: affiliation query failed: %w", err)
	}

	aff.ContractorRFC = rfc
	aff.REPSENumber = repseNumber
	aff.CURP = curp
	aff.IMSSNumber = imssNumber
	aff.JobTitle = jobTitle
	aff.PlantStatus = plantStatus

	return &aff, nil
}

func (r *complianceRepository) GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error) {
	plantReqQuery := `
		SELECT dt.id, dt.code, dt.name, dt.applies_to, pdr.is_required, COALESCE(pdr.validity_days, dt.default_validity_days)
		FROM plant_document_requirements pdr
		JOIN document_types dt ON pdr.document_type_id = dt.id
		WHERE pdr.tenant_id = $1 AND pdr.plant_id = $2 AND dt.is_active = TRUE
		ORDER BY dt.code ASC
	`

	rows, err := r.db.QueryContext(ctx, plantReqQuery, tenantID, plantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/compliance: query plant doc requirements failed: %w", err)
	}
	defer rows.Close()

	var rules []entity.RequiredDocumentRule
	for rows.Next() {
		var (
			rule         entity.RequiredDocumentRule
			validityDays *int
		)
		if err := rows.Scan(&rule.DocumentTypeID, &rule.DocumentTypeCode, &rule.DocumentName, &rule.AppliesTo, &rule.IsRequired, &validityDays); err != nil {
			return nil, fmt.Errorf("postgres/compliance: scan plant doc rule failed: %w", err)
		}
		rule.ValidityDays = validityDays
		rules = append(rules, rule)
	}

	if len(rules) > 0 {
		return rules, nil
	}

	// Fallback to active system document types
	systemQuery := `
		SELECT id, code, name, applies_to, is_required_default, default_validity_days
		FROM document_types
		WHERE is_active = TRUE AND is_system = TRUE AND (tenant_id IS NULL OR tenant_id = $1)
		ORDER BY code ASC
	`

	sysRows, err := r.db.QueryContext(ctx, systemQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/compliance: query system doc types failed: %w", err)
	}
	defer sysRows.Close()

	for sysRows.Next() {
		var (
			rule         entity.RequiredDocumentRule
			validityDays *int
		)
		if err := sysRows.Scan(&rule.DocumentTypeID, &rule.DocumentTypeCode, &rule.DocumentName, &rule.AppliesTo, &rule.IsRequired, &validityDays); err != nil {
			return nil, fmt.Errorf("postgres/compliance: scan system doc rule failed: %w", err)
		}
		rule.ValidityDays = validityDays
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *complianceRepository) GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	query := `
		SELECT d.id, dt.code, dt.name, dt.applies_to, d.status::text, d.expires_at, d.issued_at
		FROM documents d
		JOIN document_types dt ON d.document_type_id = dt.id
		WHERE d.user_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workerUserID)
	if err != nil {
		return nil, fmt.Errorf("postgres/compliance: worker documents query failed: %w", err)
	}
	defer rows.Close()

	var docs []entity.DocumentWithTypeCode
	for rows.Next() {
		var (
			d         entity.DocumentWithTypeCode
			statusStr string
			expiresAt *time.Time
			issuedAt  *time.Time
		)
		if err := rows.Scan(&d.ID, &d.DocumentTypeCode, &d.DocumentName, &d.AppliesTo, &statusStr, &expiresAt, &issuedAt); err != nil {
			return nil, fmt.Errorf("postgres/compliance: scan worker doc failed: %w", err)
		}
		d.Status = entity.DocumentStatus(statusStr)
		d.ExpiresAt = expiresAt
		d.IssuedAt = issuedAt
		docs = append(docs, d)
	}

	return docs, nil
}

func (r *complianceRepository) GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	if contractorID == uuid.Nil {
		return nil, nil
	}

	query := `
		SELECT d.id, dt.code, dt.name, dt.applies_to, d.status::text, d.expires_at, d.issued_at
		FROM documents d
		JOIN document_types dt ON d.document_type_id = dt.id
		WHERE d.contractor_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, contractorID)
	if err != nil {
		return nil, fmt.Errorf("postgres/compliance: contractor documents query failed: %w", err)
	}
	defer rows.Close()

	var docs []entity.DocumentWithTypeCode
	for rows.Next() {
		var (
			d         entity.DocumentWithTypeCode
			statusStr string
			expiresAt *time.Time
			issuedAt  *time.Time
		)
		if err := rows.Scan(&d.ID, &d.DocumentTypeCode, &d.DocumentName, &d.AppliesTo, &statusStr, &expiresAt, &issuedAt); err != nil {
			return nil, fmt.Errorf("postgres/compliance: scan contractor doc failed: %w", err)
		}
		d.Status = entity.DocumentStatus(statusStr)
		d.ExpiresAt = expiresAt
		d.IssuedAt = issuedAt
		docs = append(docs, d)
	}

	return docs, nil
}
