package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type TrainingRepository struct {
	db *sql.DB
}

func NewTrainingRepository(db *sql.DB) *TrainingRepository {
	return &TrainingRepository{db: db}
}

func (r *TrainingRepository) CreateCourse(ctx context.Context, course *entity.TrainingCourse) error {
	query := `
		INSERT INTO training_courses (
			id, tenant_id, code, title, description, thematic_area,
			stps_thematic_code, duration_hours, validity_years, normative_reference,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6::course_thematic_area,
			$7, $8, $9, $10,
			$11, $12, $13
		)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		course.ID,
		course.TenantID,
		course.Code,
		course.Title,
		course.Description,
		string(course.ThematicArea),
		course.STPSThematicCode,
		course.DurationHours,
		course.ValidityYears,
		course.NormativeReference,
		course.IsActive,
		course.CreatedAt,
		course.UpdatedAt,
	)
	return err
}

func (r *TrainingRepository) GetCourseByID(ctx context.Context, tenantID, courseID uuid.UUID) (*entity.TrainingCourse, error) {
	query := `
		SELECT id, tenant_id, code, title, description, thematic_area,
		       stps_thematic_code, duration_hours, validity_years, normative_reference,
		       is_active, created_at, updated_at
		FROM training_courses
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, courseID, tenantID)

	var c entity.TrainingCourse
	var desc, normRef sql.NullString
	var themArea string

	err := row.Scan(
		&c.ID,
		&c.TenantID,
		&c.Code,
		&c.Title,
		&desc,
		&themArea,
		&c.STPSThematicCode,
		&c.DurationHours,
		&c.ValidityYears,
		&normRef,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("curso no encontrado: %w", err)
		}
		return nil, err
	}
	c.Description = desc.String
	c.NormativeReference = normRef.String
	c.ThematicArea = entity.CourseThematicArea(themArea)

	return &c, nil
}

func (r *TrainingRepository) ListCourses(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCourse, error) {
	query := `
		SELECT id, tenant_id, code, title, description, thematic_area,
		       stps_thematic_code, duration_hours, validity_years, normative_reference,
		       is_active, created_at, updated_at
		FROM training_courses
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []entity.TrainingCourse
	for rows.Next() {
		var c entity.TrainingCourse
		var desc, normRef sql.NullString
		var themArea string

		err := rows.Scan(
			&c.ID,
			&c.TenantID,
			&c.Code,
			&c.Title,
			&desc,
			&themArea,
			&c.STPSThematicCode,
			&c.DurationHours,
			&c.ValidityYears,
			&normRef,
			&c.IsActive,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.Description = desc.String
		c.NormativeReference = normRef.String
		c.ThematicArea = entity.CourseThematicArea(themArea)
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func (r *TrainingRepository) CreateCertificate(ctx context.Context, cert *entity.TrainingCertificate) error {
	query := `
		INSERT INTO training_certificates (
			id, tenant_id, session_id, course_id, worker_id, contractor_id,
			folio_number, status, worker_occupation_code, worker_curp, worker_full_name,
			company_name, company_rfc, legal_rep_name, worker_rep_name, instructor_name,
			issue_date, expiration_date, score_percentage, digital_signature_token, document_r2_key,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8::certificate_status, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21,
			$22, $23
		)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		cert.ID,
		cert.TenantID,
		cert.SessionID,
		cert.CourseID,
		cert.WorkerID,
		cert.ContractorID,
		cert.FolioNumber,
		string(cert.Status),
		cert.WorkerOccupationCode,
		cert.WorkerCURP,
		cert.WorkerFullName,
		cert.CompanyName,
		cert.CompanyRFC,
		cert.LegalRepName,
		cert.WorkerRepName,
		cert.InstructorName,
		cert.IssueDate,
		cert.ExpirationDate,
		cert.ScorePercentage,
		cert.DigitalSignatureToken,
		cert.DocumentR2Key,
		cert.CreatedAt,
		cert.UpdatedAt,
	)
	return err
}

func (r *TrainingRepository) GetCertificateByID(ctx context.Context, tenantID, certID uuid.UUID) (*entity.TrainingCertificate, error) {
	query := `
		SELECT id, tenant_id, session_id, course_id, worker_id, contractor_id,
		       folio_number, status, worker_occupation_code, worker_curp, worker_full_name,
		       company_name, company_rfc, legal_rep_name, worker_rep_name, instructor_name,
		       issue_date, expiration_date, score_percentage, digital_signature_token, document_r2_key,
		       created_at, updated_at
		FROM training_certificates
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, certID, tenantID)

	var cert entity.TrainingCertificate
	var sessionID, contractorID *uuid.UUID
	var statusStr, sigToken, r2Key sql.NullString

	err := row.Scan(
		&cert.ID,
		&cert.TenantID,
		&sessionID,
		&cert.CourseID,
		&cert.WorkerID,
		&contractorID,
		&cert.FolioNumber,
		&statusStr,
		&cert.WorkerOccupationCode,
		&cert.WorkerCURP,
		&cert.WorkerFullName,
		&cert.CompanyName,
		&cert.CompanyRFC,
		&cert.LegalRepName,
		&cert.WorkerRepName,
		&cert.InstructorName,
		&cert.IssueDate,
		&cert.ExpirationDate,
		&cert.ScorePercentage,
		&sigToken,
		&r2Key,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("constancia DC-3 no encontrada: %w", err)
		}
		return nil, err
	}
	cert.SessionID = sessionID
	cert.ContractorID = contractorID
	cert.Status = entity.CertificateStatus(statusStr.String)
	cert.DigitalSignatureToken = sigToken.String
	cert.DocumentR2Key = r2Key.String

	return &cert, nil
}

func (r *TrainingRepository) ListCertificatesByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.TrainingCertificate, error) {
	query := `
		SELECT id, tenant_id, session_id, course_id, worker_id, contractor_id,
		       folio_number, status, worker_occupation_code, worker_curp, worker_full_name,
		       company_name, company_rfc, legal_rep_name, worker_rep_name, instructor_name,
		       issue_date, expiration_date, score_percentage, digital_signature_token, document_r2_key,
		       created_at, updated_at
		FROM training_certificates
		WHERE tenant_id = $1 AND worker_id = $2
		ORDER BY issue_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, workerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []entity.TrainingCertificate
	for rows.Next() {
		var cert entity.TrainingCertificate
		var sessionID, contractorID *uuid.UUID
		var statusStr, sigToken, r2Key sql.NullString

		err := rows.Scan(
			&cert.ID,
			&cert.TenantID,
			&sessionID,
			&cert.CourseID,
			&cert.WorkerID,
			&contractorID,
			&cert.FolioNumber,
			&statusStr,
			&cert.WorkerOccupationCode,
			&cert.WorkerCURP,
			&cert.WorkerFullName,
			&cert.CompanyName,
			&cert.CompanyRFC,
			&cert.LegalRepName,
			&cert.WorkerRepName,
			&cert.InstructorName,
			&cert.IssueDate,
			&cert.ExpirationDate,
			&cert.ScorePercentage,
			&sigToken,
			&r2Key,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cert.SessionID = sessionID
		cert.ContractorID = contractorID
		cert.Status = entity.CertificateStatus(statusStr.String)
		cert.DigitalSignatureToken = sigToken.String
		cert.DocumentR2Key = r2Key.String
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}

func (r *TrainingRepository) ListCertificates(ctx context.Context, tenantID uuid.UUID) ([]entity.TrainingCertificate, error) {
	query := `
		SELECT id, tenant_id, session_id, course_id, worker_id, contractor_id,
		       folio_number, status, worker_occupation_code, worker_curp, worker_full_name,
		       company_name, company_rfc, legal_rep_name, worker_rep_name, instructor_name,
		       issue_date, expiration_date, score_percentage, digital_signature_token, document_r2_key,
		       created_at, updated_at
		FROM training_certificates
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []entity.TrainingCertificate
	for rows.Next() {
		var cert entity.TrainingCertificate
		var sessionID, contractorID *uuid.UUID
		var statusStr, sigToken, r2Key sql.NullString

		err := rows.Scan(
			&cert.ID,
			&cert.TenantID,
			&sessionID,
			&cert.CourseID,
			&cert.WorkerID,
			&contractorID,
			&cert.FolioNumber,
			&statusStr,
			&cert.WorkerOccupationCode,
			&cert.WorkerCURP,
			&cert.WorkerFullName,
			&cert.CompanyName,
			&cert.CompanyRFC,
			&cert.LegalRepName,
			&cert.WorkerRepName,
			&cert.InstructorName,
			&cert.IssueDate,
			&cert.ExpirationDate,
			&cert.ScorePercentage,
			&sigToken,
			&r2Key,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cert.SessionID = sessionID
		cert.ContractorID = contractorID
		cert.Status = entity.CertificateStatus(statusStr.String)
		cert.DigitalSignatureToken = sigToken.String
		cert.DocumentR2Key = r2Key.String
		certs = append(certs, cert)
	}
	return certs, rows.Err()
}
