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

type contractorRepository struct {
	db *sql.DB
}

// NewContractorRepository creates a PostgreSQL repository for contractors and crew management.
func NewContractorRepository(db *sql.DB) repository.ContractorRepository {
	return &contractorRepository{db: db}
}

func (r *contractorRepository) CreateContractor(ctx context.Context, c *entity.Contractor) error {
	query := `
		INSERT INTO contractors (
			id, name, rfc, repse_number, contractor_type, admin_user_id, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	now := time.Now().UTC()
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		c.ID,
		c.Name,
		c.RFC,
		c.REPSENumber,
		string(c.ContractorType),
		c.AdminUserID,
		c.IsActive,
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/contractor: create contractor failed: %w", err)
	}
	return nil
}

func (r *contractorRepository) GetContractorByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error) {
	query := `
		SELECT id, name, rfc, repse_number, contractor_type::text, admin_user_id, is_active, created_at, updated_at, deleted_at
		FROM contractors
		WHERE id = $1 AND deleted_at IS NULL
	`
	var (
		c              entity.Contractor
		contractorType string
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.Name,
		&c.RFC,
		&c.REPSENumber,
		&contractorType,
		&c.AdminUserID,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrContractorNotFound
		}
		return nil, fmt.Errorf("postgres/contractor: get contractor by id failed: %w", err)
	}
	c.ContractorType = entity.ContractorType(contractorType)
	return &c, nil
}

func (r *contractorRepository) GetContractorByRFC(ctx context.Context, rfc string) (*entity.Contractor, error) {
	query := `
		SELECT id, name, rfc, repse_number, contractor_type::text, admin_user_id, is_active, created_at, updated_at, deleted_at
		FROM contractors
		WHERE UPPER(rfc) = UPPER($1) AND deleted_at IS NULL
		LIMIT 1
	`
	var (
		c              entity.Contractor
		contractorType string
	)

	err := r.db.QueryRowContext(ctx, query, rfc).Scan(
		&c.ID,
		&c.Name,
		&c.RFC,
		&c.REPSENumber,
		&contractorType,
		&c.AdminUserID,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrContractorNotFound
		}
		return nil, fmt.Errorf("postgres/contractor: get contractor by rfc failed: %w", err)
	}
	c.ContractorType = entity.ContractorType(contractorType)
	return &c, nil
}

func (r *contractorRepository) GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error) {
	query := `
		SELECT id, name, rfc, repse_number, contractor_type::text, admin_user_id, is_active, created_at, updated_at, deleted_at
		FROM contractors
		WHERE admin_user_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var (
		c              entity.Contractor
		contractorType string
	)

	err := r.db.QueryRowContext(ctx, query, adminUserID).Scan(
		&c.ID,
		&c.Name,
		&c.RFC,
		&c.REPSENumber,
		&contractorType,
		&c.AdminUserID,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrContractorNotFound
		}
		return nil, fmt.Errorf("postgres/contractor: get contractor by admin user id failed: %w", err)
	}
	c.ContractorType = entity.ContractorType(contractorType)
	return &c, nil
}

func (r *contractorRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]repository.ContractorWithDetails, error) {
	query := `
		SELECT 
			c.id, 
			c.name, 
			c.rfc, 
			c.repse_number, 
			c.contractor_type::text, 
			c.admin_user_id,
			u.full_name as admin_full_name,
			u.email as admin_email,
			u.phone as admin_phone,
			pc.status::text as plant_status,
			COALESCE((SELECT COUNT(*) FROM contractor_workers cw WHERE cw.contractor_id = c.id AND cw.deleted_at IS NULL), 0) as workers_count,
			c.is_active,
			c.created_at
		FROM plant_contractors pc
		JOIN contractors c ON pc.contractor_id = c.id
		LEFT JOIN users u ON c.admin_user_id = u.id
		WHERE pc.plant_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, plantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/contractor: list by plant query failed: %w", err)
	}
	defer rows.Close()

	var result []repository.ContractorWithDetails
	for rows.Next() {
		var (
			item           repository.ContractorWithDetails
			contractorType string
			plantStatus    string
		)
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.RFC,
			&item.REPSENumber,
			&contractorType,
			&item.AdminUserID,
			&item.AdminFullName,
			&item.AdminEmail,
			&item.AdminPhone,
			&plantStatus,
			&item.WorkersCount,
			&item.IsActive,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/contractor: scan contractor item failed: %w", err)
		}
		item.ContractorType = entity.ContractorType(contractorType)
		item.PlantStatus = entity.PlantContractorStatus(plantStatus)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/contractor: rows iteration failed: %w", err)
	}

	return result, nil
}

func (r *contractorRepository) CreatePlantContractor(ctx context.Context, pc *entity.PlantContractor) error {
	query := `
		INSERT INTO plant_contractors (
			id, plant_id, tenant_id, contractor_id, status, invited_by, invited_at, activated_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (plant_id, contractor_id)
		DO UPDATE SET status = EXCLUDED.status, updated_at = now()
	`
	now := time.Now().UTC()
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	pc.CreatedAt = now
	pc.UpdatedAt = now
	if pc.InvitedAt.IsZero() {
		pc.InvitedAt = now
	}

	_, err := r.db.ExecContext(ctx, query,
		pc.ID,
		pc.PlantID,
		pc.TenantID,
		pc.ContractorID,
		string(pc.Status),
		pc.InvitedBy,
		pc.InvitedAt,
		pc.ActivatedAt,
		pc.CreatedAt,
		pc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/contractor: create plant contractor failed: %w", err)
	}
	return nil
}

func (r *contractorRepository) UpdatePlantContractorStatus(ctx context.Context, plantID, contractorID uuid.UUID, status entity.PlantContractorStatus) error {
	query := `
		UPDATE plant_contractors
		SET 
			status = $1, 
			updated_at = now(),
			activated_at = CASE WHEN $1 = 'active' AND activated_at IS NULL THEN now() ELSE activated_at END
		WHERE plant_id = $2 AND contractor_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, string(status), plantID, contractorID)
	if err != nil {
		return fmt.Errorf("postgres/contractor: update status failed: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domainErrors.ErrContractorNotFound.WithMessage("La vinculación con la planta no existe.")
	}
	return nil
}

func (r *contractorRepository) ListWorkersByContractor(ctx context.Context, contractorID uuid.UUID) ([]repository.ContractorWorkerDetail, error) {
	query := `
		SELECT
			u.id as worker_id,
			cw.contractor_id,
			u.full_name,
			u.email,
			u.phone,
			cw.job_title,
			cw.curp,
			cw.imss_number,
			cw.is_active as cw_is_active,
			u.is_active as u_is_active,
			COALESCE((SELECT COUNT(*) FROM documents d WHERE d.user_id = u.id AND d.deleted_at IS NULL), 0) as docs_count,
			cw.created_at
		FROM contractor_workers cw
		JOIN users u ON cw.user_id = u.id
		WHERE cw.contractor_id = $1 AND cw.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY u.full_name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, contractorID)
	if err != nil {
		return nil, fmt.Errorf("postgres/contractor: list workers query failed: %w", err)
	}
	defer rows.Close()

	var result []repository.ContractorWorkerDetail
	for rows.Next() {
		var item repository.ContractorWorkerDetail
		err := rows.Scan(
			&item.WorkerID,
			&item.ContractorID,
			&item.FullName,
			&item.Email,
			&item.Phone,
			&item.JobTitle,
			&item.CURP,
			&item.IMSSNumber,
			&item.IsActive,
			&item.UserIsActive,
			&item.DocumentsCount,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/contractor: scan worker failed: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/contractor: rows iteration failed: %w", err)
	}

	return result, nil
}

func (r *contractorRepository) AddWorker(ctx context.Context, workerUser *entity.User, cw *entity.ContractorWorker) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres/contractor: failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if workerUser.ID == uuid.Nil {
		workerUser.ID = uuid.New()
	}
	workerUser.CreatedAt = now
	workerUser.UpdatedAt = now

	// 1. Insert into users
	userQuery := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	_, err = tx.ExecContext(ctx, userQuery,
		workerUser.ID,
		workerUser.TenantID,
		workerUser.Email,
		workerUser.PasswordHash,
		workerUser.FullName,
		workerUser.Phone,
		string(workerUser.Role),
		workerUser.IsActive,
		workerUser.CreatedAt,
		workerUser.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/contractor: insert worker user failed: %w", err)
	}

	// 2. Insert into contractor_workers
	if cw.ID == uuid.Nil {
		cw.ID = uuid.New()
	}
	cw.UserID = workerUser.ID
	cw.CreatedAt = now
	cw.UpdatedAt = now

	cwQuery := `
		INSERT INTO contractor_workers (
			id, contractor_id, user_id, job_title, curp, imss_number, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	_, err = tx.ExecContext(ctx, cwQuery,
		cw.ID,
		cw.ContractorID,
		cw.UserID,
		cw.JobTitle,
		cw.CURP,
		cw.IMSSNumber,
		cw.IsActive,
		cw.CreatedAt,
		cw.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/contractor: insert contractor_worker failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres/contractor: commit tx failed: %w", err)
	}
	return nil
}

func (r *contractorRepository) ToggleWorkerStatus(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	cwQuery := `
		UPDATE contractor_workers
		SET is_active = $1, updated_at = now()
		WHERE contractor_id = $2 AND user_id = $3 AND deleted_at IS NULL
	`
	_, err = tx.ExecContext(ctx, cwQuery, isActive, contractorID, workerUserID)
	if err != nil {
		return fmt.Errorf("postgres/contractor: toggle contractor_workers failed: %w", err)
	}

	userQuery := `
		UPDATE users
		SET is_active = $1, updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err = tx.ExecContext(ctx, userQuery, isActive, workerUserID)
	if err != nil {
		return fmt.Errorf("postgres/contractor: toggle users failed: %w", err)
	}

	return tx.Commit()
}

func (r *contractorRepository) CreateUserWithPassword(ctx context.Context, user *entity.User, passwordHash string) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, password_hash, full_name, phone, role, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`
	now := time.Now().UTC()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = now
	user.UpdatedAt = now
	user.PasswordHash = passwordHash

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.TenantID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		string(user.Role),
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres/contractor: create user failed: %w", err)
	}
	return nil
}
