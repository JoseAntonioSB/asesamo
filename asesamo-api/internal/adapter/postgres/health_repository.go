package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

type HealthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

func (r *HealthRepository) UpsertMedicalProfile(ctx context.Context, profile *entity.WorkerMedicalProfile) error {
	query := `
		INSERT INTO worker_medical_profiles (
			id, tenant_id, worker_id, blood_type, allergies,
			chronic_conditions, emergency_contact_name, emergency_contact_phone,
			emergency_contact_relationship, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11
		)
		ON CONFLICT (tenant_id, worker_id) DO UPDATE SET
			blood_type = EXCLUDED.blood_type,
			allergies = EXCLUDED.allergies,
			chronic_conditions = EXCLUDED.chronic_conditions,
			emergency_contact_name = EXCLUDED.emergency_contact_name,
			emergency_contact_phone = EXCLUDED.emergency_contact_phone,
			emergency_contact_relationship = EXCLUDED.emergency_contact_relationship,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		profile.ID,
		profile.TenantID,
		profile.WorkerID,
		profile.BloodType,
		profile.Allergies,
		profile.ChronicConditions,
		profile.EmergencyContactName,
		profile.EmergencyContactPhone,
		profile.EmergencyContactRelationship,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	return err
}

func (r *HealthRepository) GetMedicalProfileByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.WorkerMedicalProfile, error) {
	query := `
		SELECT id, tenant_id, worker_id, blood_type, allergies,
		       chronic_conditions, emergency_contact_name, emergency_contact_phone,
		       emergency_contact_relationship, created_at, updated_at
		FROM worker_medical_profiles
		WHERE tenant_id = $1 AND worker_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, workerID)

	var p entity.WorkerMedicalProfile
	var allergies, chronic sql.NullString

	err := row.Scan(
		&p.ID,
		&p.TenantID,
		&p.WorkerID,
		&p.BloodType,
		&allergies,
		&chronic,
		&p.EmergencyContactName,
		&p.EmergencyContactPhone,
		&p.EmergencyContactRelationship,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("perfil médico no encontrado: %w", err)
		}
		return nil, err
	}
	p.Allergies = allergies.String
	p.ChronicConditions = chronic.String

	return &p, nil
}

func (r *HealthRepository) CreateMedicalExam(ctx context.Context, exam *entity.OccupationalMedicalExam) error {
	query := `
		INSERT INTO occupational_medical_exams (
			id, tenant_id, worker_id, contractor_id, exam_number,
			exam_type, fitness_status, exam_date, expiration_date,
			doctor_name, doctor_license, clinic_or_institution,
			restrictions_notes, recommendations, visual_acuity_result,
			audiometry_result, spirometry_result, antidoping_result,
			blood_pressure, certificate_r2_key, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6::medical_exam_type, $7::medical_fitness_status, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			$19, $20, $21, $22
		)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		exam.ID,
		exam.TenantID,
		exam.WorkerID,
		exam.ContractorID,
		exam.ExamNumber,
		string(exam.ExamType),
		string(exam.FitnessStatus),
		exam.ExamDate,
		exam.ExpirationDate,
		exam.DoctorName,
		exam.DoctorLicense,
		exam.ClinicOrInstitution,
		exam.RestrictionsNotes,
		exam.Recommendations,
		exam.VisualAcuityResult,
		exam.AudiometryResult,
		exam.SpirometryResult,
		exam.AntidopingResult,
		exam.BloodPressure,
		exam.CertificateR2Key,
		exam.CreatedAt,
		exam.UpdatedAt,
	)
	return err
}

func (r *HealthRepository) GetMedicalExamByID(ctx context.Context, tenantID, examID uuid.UUID) (*entity.OccupationalMedicalExam, error) {
	query := `
		SELECT id, tenant_id, worker_id, contractor_id, exam_number,
		       exam_type, fitness_status, exam_date, expiration_date,
		       doctor_name, doctor_license, clinic_or_institution,
		       restrictions_notes, recommendations, visual_acuity_result,
		       audiometry_result, spirometry_result, antidoping_result,
		       blood_pressure, certificate_r2_key, created_at, updated_at
		FROM occupational_medical_exams
		WHERE id = $1 AND tenant_id = $2
	`
	row := r.db.QueryRowContext(ctx, query, examID, tenantID)
	return r.scanExamRow(row)
}

func (r *HealthRepository) GetLatestExamByWorker(ctx context.Context, tenantID, workerID uuid.UUID) (*entity.OccupationalMedicalExam, error) {
	query := `
		SELECT id, tenant_id, worker_id, contractor_id, exam_number,
		       exam_type, fitness_status, exam_date, expiration_date,
		       doctor_name, doctor_license, clinic_or_institution,
		       restrictions_notes, recommendations, visual_acuity_result,
		       audiometry_result, spirometry_result, antidoping_result,
		       blood_pressure, certificate_r2_key, created_at, updated_at
		FROM occupational_medical_exams
		WHERE tenant_id = $1 AND worker_id = $2
		ORDER BY exam_date DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, tenantID, workerID)
	return r.scanExamRow(row)
}

func (r *HealthRepository) ListExamsByWorker(ctx context.Context, tenantID, workerID uuid.UUID) ([]entity.OccupationalMedicalExam, error) {
	query := `
		SELECT id, tenant_id, worker_id, contractor_id, exam_number,
		       exam_type, fitness_status, exam_date, expiration_date,
		       doctor_name, doctor_license, clinic_or_institution,
		       restrictions_notes, recommendations, visual_acuity_result,
		       audiometry_result, spirometry_result, antidoping_result,
		       blood_pressure, certificate_r2_key, created_at, updated_at
		FROM occupational_medical_exams
		WHERE tenant_id = $1 AND worker_id = $2
		ORDER BY exam_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID, workerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []entity.OccupationalMedicalExam
	for rows.Next() {
		exam, err := r.scanExamFromRows(rows)
		if err != nil {
			return nil, err
		}
		exams = append(exams, *exam)
	}
	return exams, rows.Err()
}

func (r *HealthRepository) ListAllExams(ctx context.Context, tenantID uuid.UUID) ([]entity.OccupationalMedicalExam, error) {
	query := `
		SELECT id, tenant_id, worker_id, contractor_id, exam_number,
		       exam_type, fitness_status, exam_date, expiration_date,
		       doctor_name, doctor_license, clinic_or_institution,
		       restrictions_notes, recommendations, visual_acuity_result,
		       audiometry_result, spirometry_result, antidoping_result,
		       blood_pressure, certificate_r2_key, created_at, updated_at
		FROM occupational_medical_exams
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []entity.OccupationalMedicalExam
	for rows.Next() {
		exam, err := r.scanExamFromRows(rows)
		if err != nil {
			return nil, err
		}
		exams = append(exams, *exam)
	}
	return exams, rows.Err()
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *HealthRepository) scanExamRow(s scannable) (*entity.OccupationalMedicalExam, error) {
	var e entity.OccupationalMedicalExam
	var contractorID *uuid.UUID
	var examType, fitness string
	var restr, recom, vis, aud, spir, antidop, bp, r2Key sql.NullString

	err := s.Scan(
		&e.ID,
		&e.TenantID,
		&e.WorkerID,
		&contractorID,
		&e.ExamNumber,
		&examType,
		&fitness,
		&e.ExamDate,
		&e.ExpirationDate,
		&e.DoctorName,
		&e.DoctorLicense,
		&e.ClinicOrInstitution,
		&restr,
		&recom,
		&vis,
		&aud,
		&spir,
		&antidop,
		&bp,
		&r2Key,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("examen médico no encontrado: %w", err)
		}
		return nil, err
	}
	e.ContractorID = contractorID
	e.ExamType = entity.MedicalExamType(examType)
	e.FitnessStatus = entity.MedicalFitnessStatus(fitness)
	e.RestrictionsNotes = restr.String
	e.Recommendations = recom.String
	e.VisualAcuityResult = vis.String
	e.AudiometryResult = aud.String
	e.SpirometryResult = spir.String
	e.AntidopingResult = antidop.String
	e.BloodPressure = bp.String
	e.CertificateR2Key = r2Key.String

	return &e, nil
}

func (r *HealthRepository) scanExamFromRows(rows *sql.Rows) (*entity.OccupationalMedicalExam, error) {
	return r.scanExamRow(rows)
}
