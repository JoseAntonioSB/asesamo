package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"

	"github.com/google/uuid"
)

// ContractorWithDetails contains contractor corporate details, admin contact, and plant link status.
type ContractorWithDetails struct {
	ID             uuid.UUID                    `json:"id"`
	Name           string                       `json:"name"`
	RFC            *string                      `json:"rfc,omitempty"`
	REPSENumber    *string                      `json:"repse_number,omitempty"`
	ContractorType entity.ContractorType        `json:"contractor_type"`
	AdminUserID    *uuid.UUID                   `json:"admin_user_id,omitempty"`
	AdminFullName  *string                      `json:"admin_full_name,omitempty"`
	AdminEmail     *string                      `json:"admin_email,omitempty"`
	AdminPhone     *string                      `json:"admin_phone,omitempty"`
	PlantStatus    entity.PlantContractorStatus `json:"plant_status"`
	WorkersCount   int                          `json:"workers_count"`
	IsActive       bool                         `json:"is_active"`
	CreatedAt      time.Time                    `json:"created_at"`
}

// ContractorWorkerDetail represents a worker in a contractor's crew.
type ContractorWorkerDetail struct {
	WorkerID        uuid.UUID  `json:"worker_id"`
	ContractorID    uuid.UUID  `json:"contractor_id"`
	FullName        string     `json:"full_name"`
	Email           string     `json:"email"`
	Phone           *string    `json:"phone,omitempty"`
	JobTitle        *string    `json:"job_title,omitempty"`
	CURP            *string    `json:"curp,omitempty"`
	IMSSNumber      *string    `json:"imss_number,omitempty"`
	IsActive        bool       `json:"is_active"`
	UserIsActive    bool       `json:"user_is_active"`
	DocumentsCount  int        `json:"documents_count"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ContractorRepository defines data access methods for contractors, plant affiliations, and crews.
type ContractorRepository interface {
	CreateContractor(ctx context.Context, c *entity.Contractor) error
	GetContractorByID(ctx context.Context, id uuid.UUID) (*entity.Contractor, error)
	GetContractorByRFC(ctx context.Context, rfc string) (*entity.Contractor, error)
	GetContractorByAdminUserID(ctx context.Context, adminUserID uuid.UUID) (*entity.Contractor, error)
	ListByPlant(ctx context.Context, plantID uuid.UUID) ([]ContractorWithDetails, error)
	CreatePlantContractor(ctx context.Context, pc *entity.PlantContractor) error
	UpdatePlantContractorStatus(ctx context.Context, plantID, contractorID uuid.UUID, status entity.PlantContractorStatus) error
	ListWorkersByContractor(ctx context.Context, contractorID uuid.UUID) ([]ContractorWorkerDetail, error)
	AddWorker(ctx context.Context, workerUser *entity.User, cw *entity.ContractorWorker) error
	ToggleWorkerStatus(ctx context.Context, contractorID, workerUserID uuid.UUID, isActive bool) error
	CreateUserWithPassword(ctx context.Context, user *entity.User, passwordHash string) error
}
