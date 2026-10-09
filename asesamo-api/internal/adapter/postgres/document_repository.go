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

type documentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a PostgreSQL document repository.
func NewDocumentRepository(db *sql.DB) repository.DocumentRepository {
	return &documentRepository{db: db}
}

func (r *documentRepository) Create(ctx context.Context, doc *entity.Document) error {
	query := `
		INSERT INTO documents (
			id, user_id, contractor_id, document_type_id, r2_key,
			file_name, file_size_bytes, mime_type, issued_at, expires_at,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13
		)
	`

	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	now := time.Now().UTC()
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = now
	}
	if doc.UpdatedAt.IsZero() {
		doc.UpdatedAt = now
	}
	if doc.Status == "" {
		doc.Status = entity.DocumentStatusPendingReview
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		doc.ID,
		doc.UserID,
		doc.ContractorID,
		doc.DocumentTypeID,
		doc.R2Key,
		doc.FileName,
		doc.FileSizeBytes,
		doc.MimeType,
		doc.IssuedAt,
		doc.ExpiresAt,
		string(doc.Status),
		doc.CreatedAt,
		doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/document: insert failed: %w", err)
	}

	return nil
}

func (r *documentRepository) scanDocument(row interface{ Scan(dest ...any) error }) (*entity.Document, error) {
	var (
		d               entity.Document
		userID          *uuid.UUID
		contractorID    *uuid.UUID
		fileSizeBytes   *int
		mimeType        *string
		issuedAt        *time.Time
		expiresAt       *time.Time
		statusStr       string
		reviewedBy      *uuid.UUID
		reviewedAt      *time.Time
		rejectionReason *string
		aiConfidence    *float64
		deletedAt       *time.Time
		typeCode        string
		typeName        string
		ownerName       string
		ownerType       string
		contractorName  string
		reviewerName    string
	)

	err := row.Scan(
		&d.ID,
		&userID,
		&contractorID,
		&d.DocumentTypeID,
		&d.R2Key,
		&d.FileName,
		&fileSizeBytes,
		&mimeType,
		&issuedAt,
		&expiresAt,
		&statusStr,
		&reviewedBy,
		&reviewedAt,
		&rejectionReason,
		&aiConfidence,
		&d.CreatedAt,
		&d.UpdatedAt,
		&deletedAt,
		&typeCode,
		&typeName,
		&ownerName,
		&ownerType,
		&contractorName,
		&reviewerName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("postgres/document: scan failed: %w", err)
	}

	d.UserID = userID
	d.ContractorID = contractorID
	d.FileSizeBytes = fileSizeBytes
	d.MimeType = mimeType
	d.IssuedAt = issuedAt
	d.ExpiresAt = expiresAt
	d.Status = entity.DocumentStatus(statusStr)
	d.ReviewedBy = reviewedBy
	d.ReviewedAt = reviewedAt
	d.RejectionReason = rejectionReason
	d.AIConfidence = aiConfidence
	d.DeletedAt = deletedAt
	d.DocumentTypeCode = typeCode
	d.DocumentTypeName = typeName
	d.OwnerName = ownerName
	d.OwnerType = ownerType
	d.ContractorName = contractorName
	d.ReviewedByName = reviewerName

	return &d, nil
}

func (r *documentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Document, error) {
	query := `
		SELECT d.id, d.user_id, d.contractor_id, d.document_type_id, d.r2_key,
		       d.file_name, d.file_size_bytes, d.mime_type, d.issued_at, d.expires_at,
		       d.status, d.reviewed_by, d.reviewed_at, d.rejection_reason, d.ai_confidence,
		       d.created_at, d.updated_at, d.deleted_at,
		       COALESCE(dt.code, ''), COALESCE(dt.name, ''),
		       COALESCE(u.full_name, c.name, ''),
		       CASE WHEN d.user_id IS NOT NULL THEN 'worker' ELSE 'contractor' END,
		       COALESCE(wc.name, c.name, ''),
		       COALESCE(reviewer.full_name, '')
		FROM documents d
		LEFT JOIN document_types dt ON d.document_type_id = dt.id
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN contractors c ON d.contractor_id = c.id
		LEFT JOIN contractor_workers cw ON d.user_id = cw.user_id AND cw.is_active = TRUE
		LEFT JOIN contractors wc ON cw.contractor_id = wc.id
		LEFT JOIN users reviewer ON d.reviewed_by = reviewer.id
		WHERE d.id = $1 AND d.deleted_at IS NULL
	`
	return r.scanDocument(r.db.QueryRowContext(ctx, query, id))
}

func (r *documentRepository) ListByWorker(ctx context.Context, workerUserID uuid.UUID) ([]entity.Document, error) {
	query := `
		SELECT d.id, d.user_id, d.contractor_id, d.document_type_id, d.r2_key,
		       d.file_name, d.file_size_bytes, d.mime_type, d.issued_at, d.expires_at,
		       d.status, d.reviewed_by, d.reviewed_at, d.rejection_reason, d.ai_confidence,
		       d.created_at, d.updated_at, d.deleted_at,
		       COALESCE(dt.code, ''), COALESCE(dt.name, ''),
		       COALESCE(u.full_name, c.name, ''),
		       'worker',
		       COALESCE(wc.name, c.name, ''),
		       COALESCE(reviewer.full_name, '')
		FROM documents d
		LEFT JOIN document_types dt ON d.document_type_id = dt.id
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN contractors c ON d.contractor_id = c.id
		LEFT JOIN contractor_workers cw ON d.user_id = cw.user_id AND cw.is_active = TRUE
		LEFT JOIN contractors wc ON cw.contractor_id = wc.id
		LEFT JOIN users reviewer ON d.reviewed_by = reviewer.id
		WHERE d.user_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, workerUserID)
	if err != nil {
		return nil, fmt.Errorf("postgres/document: ListByWorker query failed: %w", err)
	}
	defer rows.Close()

	var docs []entity.Document
	for rows.Next() {
		d, err := r.scanDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, *d)
	}

	return docs, nil
}

func (r *documentRepository) ListByContractor(ctx context.Context, contractorID uuid.UUID) ([]entity.Document, error) {
	query := `
		SELECT d.id, d.user_id, d.contractor_id, d.document_type_id, d.r2_key,
		       d.file_name, d.file_size_bytes, d.mime_type, d.issued_at, d.expires_at,
		       d.status, d.reviewed_by, d.reviewed_at, d.rejection_reason, d.ai_confidence,
		       d.created_at, d.updated_at, d.deleted_at,
		       COALESCE(dt.code, ''), COALESCE(dt.name, ''),
		       COALESCE(c.name, ''),
		       'contractor',
		       COALESCE(c.name, ''),
		       COALESCE(reviewer.full_name, '')
		FROM documents d
		LEFT JOIN document_types dt ON d.document_type_id = dt.id
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN contractors c ON d.contractor_id = c.id
		LEFT JOIN contractor_workers cw ON d.user_id = cw.user_id AND cw.is_active = TRUE
		LEFT JOIN contractors wc ON cw.contractor_id = wc.id
		LEFT JOIN users reviewer ON d.reviewed_by = reviewer.id
		WHERE d.contractor_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, contractorID)
	if err != nil {
		return nil, fmt.Errorf("postgres/document: ListByContractor query failed: %w", err)
	}
	defer rows.Close()

	var docs []entity.Document
	for rows.Next() {
		d, err := r.scanDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, *d)
	}

	return docs, nil
}

func (r *documentRepository) ListPendingReview(ctx context.Context, tenantID *uuid.UUID, statusFilter *entity.DocumentStatus) ([]entity.Document, error) {
	var statusParam *string
	if statusFilter != nil {
		s := string(*statusFilter)
		statusParam = &s
	}

	query := `
		SELECT d.id, d.user_id, d.contractor_id, d.document_type_id, d.r2_key,
		       d.file_name, d.file_size_bytes, d.mime_type, d.issued_at, d.expires_at,
		       d.status, d.reviewed_by, d.reviewed_at, d.rejection_reason, d.ai_confidence,
		       d.created_at, d.updated_at, d.deleted_at,
		       COALESCE(dt.code, ''), COALESCE(dt.name, ''),
		       COALESCE(u.full_name, c.name, ''),
		       CASE WHEN d.user_id IS NOT NULL THEN 'worker' ELSE 'contractor' END,
		       COALESCE(wc.name, c.name, ''),
		       COALESCE(reviewer.full_name, '')
		FROM documents d
		LEFT JOIN document_types dt ON d.document_type_id = dt.id
		LEFT JOIN users u ON d.user_id = u.id
		LEFT JOIN contractors c ON d.contractor_id = c.id
		LEFT JOIN contractor_workers cw ON d.user_id = cw.user_id AND cw.is_active = TRUE
		LEFT JOIN contractors wc ON cw.contractor_id = wc.id
		LEFT JOIN users reviewer ON d.reviewed_by = reviewer.id
		WHERE d.deleted_at IS NULL
		  AND ($1::text IS NULL OR d.status::text = $1)
		  AND (
		    $2::uuid IS NULL 
		    OR u.tenant_id = $2
		    OR EXISTS (
		      SELECT 1 FROM plant_contractors pc 
		      WHERE pc.tenant_id = $2 
		        AND (pc.contractor_id = d.contractor_id OR pc.contractor_id = cw.contractor_id)
		    )
		  )
		ORDER BY d.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, statusParam, tenantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/document: ListPendingReview failed: %w", err)
	}
	defer rows.Close()

	var docs []entity.Document
	for rows.Next() {
		d, err := r.scanDocument(rows)
		if err != nil {
			return nil, err
		}
		docs = append(docs, *d)
	}

	return docs, nil
}

func (r *documentRepository) UpdateReviewStatus(ctx context.Context, id uuid.UUID, status entity.DocumentStatus, reviewedBy uuid.UUID, rejectionReason *string) (*entity.Document, error) {
	query := `
		UPDATE documents
		SET status = $2,
		    reviewed_by = $3,
		    reviewed_at = NOW(),
		    rejection_reason = $4,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`
	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id, string(status), reviewedBy, rejectionReason).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("postgres/document: UpdateReviewStatus failed: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *documentRepository) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	query := `
		SELECT id, name, rfc, repse_number, contractor_type, admin_user_id, is_active,
		       created_at, updated_at, deleted_at
		FROM contractors
		WHERE admin_user_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var (
		c           entity.Contractor
		rfc         *string
		repseNumber *string
		adminUID    *uuid.UUID
		deletedAt   *time.Time
		cType       string
	)
	err := r.db.QueryRowContext(ctx, query, adminUserID).Scan(
		&c.ID,
		&c.Name,
		&rfc,
		&repseNumber,
		&cType,
		&adminUID,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrContractorNotFound
		}
		return nil, fmt.Errorf("postgres/document: GetContractorByAdminUserID failed: %w", err)
	}
	c.RFC = rfc
	c.REPSENumber = repseNumber
	c.ContractorType = entity.ContractorType(cType)
	c.AdminUserID = adminUID
	c.DeletedAt = deletedAt
	return &c, nil
}

func (r *documentRepository) GetContractorByWorkerUserID(ctx context.Context, workerUserID uuid.UUID) (*entity.Contractor, error) {
	query := `
		SELECT c.id, c.name, c.rfc, c.repse_number, c.contractor_type, c.admin_user_id, c.is_active,
		       c.created_at, c.updated_at, c.deleted_at
		FROM contractor_workers cw
		JOIN contractors c ON cw.contractor_id = c.id
		WHERE cw.user_id = $1 AND cw.is_active = TRUE AND c.deleted_at IS NULL
		LIMIT 1
	`
	var (
		c           entity.Contractor
		rfc         *string
		repseNumber *string
		adminUID    *uuid.UUID
		deletedAt   *time.Time
		cType       string
	)
	err := r.db.QueryRowContext(ctx, query, workerUserID).Scan(
		&c.ID,
		&c.Name,
		&rfc,
		&repseNumber,
		&cType,
		&adminUID,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrContractorNotFound
		}
		return nil, fmt.Errorf("postgres/document: GetContractorByWorkerUserID failed: %w", err)
	}
	c.RFC = rfc
	c.REPSENumber = repseNumber
	c.ContractorType = entity.ContractorType(cType)
	c.AdminUserID = adminUID
	c.DeletedAt = deletedAt
	return &c, nil
}

func (r *documentRepository) GetDocumentTypeByCode(ctx context.Context, code string) (*entity.DocumentType, error) {
	query := `
		SELECT id, tenant_id, code, name, description, applies_to,
		       default_validity_days, is_required_default, is_system, is_active,
		       created_at, updated_at
		FROM document_types
		WHERE code = $1 AND is_active = TRUE
		LIMIT 1
	`
	var (
		dt                  entity.DocumentType
		tenantID            *uuid.UUID
		desc                *string
		defaultValidityDays *int
	)

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&dt.ID,
		&tenantID,
		&dt.Code,
		&dt.Name,
		&desc,
		&dt.AppliesTo,
		&defaultValidityDays,
		&dt.IsRequiredDefault,
		&dt.IsSystem,
		&dt.IsActive,
		&dt.CreatedAt,
		&dt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrDocumentTypeNotFound.WithMessage(fmt.Sprintf("Tipo de documento '%s' no válido.", code))
		}
		return nil, fmt.Errorf("postgres/document: GetDocumentTypeByCode failed: %w", err)
	}

	dt.TenantID = tenantID
	dt.Description = desc
	dt.DefaultValidityDays = defaultValidityDays
	return &dt, nil
}

func (r *documentRepository) ListRequiredDocumentTypes(ctx context.Context) ([]entity.DocumentType, error) {
	query := `
		SELECT id, tenant_id, code, name, description, applies_to,
		       default_validity_days, is_required_default, is_system, is_active,
		       created_at, updated_at
		FROM document_types
		WHERE is_active = TRUE
		ORDER BY code ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres/document: ListRequiredDocumentTypes failed: %w", err)
	}
	defer rows.Close()

	var list []entity.DocumentType
	for rows.Next() {
		var (
			dt                  entity.DocumentType
			tenantID            *uuid.UUID
			desc                *string
			defaultValidityDays *int
		)
		err := rows.Scan(
			&dt.ID,
			&tenantID,
			&dt.Code,
			&dt.Name,
			&desc,
			&dt.AppliesTo,
			&defaultValidityDays,
			&dt.IsRequiredDefault,
			&dt.IsSystem,
			&dt.IsActive,
			&dt.CreatedAt,
			&dt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dt.TenantID = tenantID
		dt.Description = desc
		dt.DefaultValidityDays = defaultValidityDays
		list = append(list, dt)
	}

	return list, nil
}
