package dto

import (
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// InviteContractorRequest represents the payload for POST /api/v1/contractors/invite.
type InviteContractorRequest struct {
	PlantID        uuid.UUID `json:"plant_id"`
	Name           string    `json:"name"`
	RFC            string    `json:"rfc"`
	REPSENumber    string    `json:"repse_number"`
	ContractorType string    `json:"contractor_type"` // company | individual
	AdminEmail     string    `json:"admin_email"`
	AdminFullName  string    `json:"admin_full_name"`
	AdminPhone     string    `json:"admin_phone"`
	InitialStatus  string    `json:"initial_status"` // active | pending
}

// UpdateContractorStatusRequest represents the payload for PATCH /api/v1/contractors/{id}/status.
type UpdateContractorStatusRequest struct {
	PlantID uuid.UUID `json:"plant_id"`
	Status  string    `json:"status"` // active, pending, suspended, blocked
}

// AddWorkerRequest represents the payload for POST /api/v1/contractors/workers.
type AddWorkerRequest struct {
	ContractorID *uuid.UUID `json:"contractor_id,omitempty"` // optional if user is contractor_admin
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	JobTitle     string     `json:"job_title"`
	CURP         string     `json:"curp"`
	IMSSNumber   string     `json:"imss_number"`
}

// ToggleWorkerStatusRequest represents the payload for PATCH /api/v1/contractors/workers/{id}/status.
type ToggleWorkerStatusRequest struct {
	ContractorID *uuid.UUID `json:"contractor_id,omitempty"` // optional if user is contractor_admin
	IsActive     bool       `json:"is_active"`
}

// ContractorResponse represents a contractor company in API responses.
type ContractorResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	RFC            *string   `json:"rfc,omitempty"`
	REPSENumber    *string   `json:"repse_number,omitempty"`
	ContractorType string    `json:"contractor_type"`
	AdminUserID    *uuid.UUID `json:"admin_user_id,omitempty"`
	AdminFullName  *string   `json:"admin_full_name,omitempty"`
	AdminEmail     *string   `json:"admin_email,omitempty"`
	AdminPhone     *string   `json:"admin_phone,omitempty"`
	PlantStatus    string    `json:"plant_status,omitempty"`
	WorkersCount   int       `json:"workers_count"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// ContractorWorkerResponse represents a worker in API responses.
type ContractorWorkerResponse struct {
	WorkerID       uuid.UUID `json:"worker_id"`
	ContractorID   uuid.UUID `json:"contractor_id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Phone          *string   `json:"phone,omitempty"`
	JobTitle       *string   `json:"job_title,omitempty"`
	CURP           *string   `json:"curp,omitempty"`
	IMSSNumber     *string   `json:"imss_number,omitempty"`
	IsActive       bool      `json:"is_active"`
	UserIsActive   bool      `json:"user_is_active"`
	DocumentsCount int       `json:"documents_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// ToContractorResponse converts repository.ContractorWithDetails to ContractorResponse.
func ToContractorResponse(c repository.ContractorWithDetails) ContractorResponse {
	return ContractorResponse{
		ID:             c.ID,
		Name:           c.Name,
		RFC:            c.RFC,
		REPSENumber:    c.REPSENumber,
		ContractorType: string(c.ContractorType),
		AdminUserID:    c.AdminUserID,
		AdminFullName:  c.AdminFullName,
		AdminEmail:     c.AdminEmail,
		AdminPhone:     c.AdminPhone,
		PlantStatus:    string(c.PlantStatus),
		WorkersCount:   c.WorkersCount,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt,
	}
}

// ToContractorFromEntity converts entity.Contractor to ContractorResponse.
func ToContractorFromEntity(c *entity.Contractor) ContractorResponse {
	return ContractorResponse{
		ID:             c.ID,
		Name:           c.Name,
		RFC:            c.RFC,
		REPSENumber:    c.REPSENumber,
		ContractorType: string(c.ContractorType),
		AdminUserID:    c.AdminUserID,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt,
	}
}

// ToContractorWorkerResponse converts repository.ContractorWorkerDetail to ContractorWorkerResponse.
func ToContractorWorkerResponse(w repository.ContractorWorkerDetail) ContractorWorkerResponse {
	return ContractorWorkerResponse{
		WorkerID:       w.WorkerID,
		ContractorID:   w.ContractorID,
		FullName:       w.FullName,
		Email:          w.Email,
		Phone:          w.Phone,
		JobTitle:       w.JobTitle,
		CURP:           w.CURP,
		IMSSNumber:     w.IMSSNumber,
		IsActive:       w.IsActive,
		UserIsActive:   w.UserIsActive,
		DocumentsCount: w.DocumentsCount,
		CreatedAt:      w.CreatedAt,
	}
}
