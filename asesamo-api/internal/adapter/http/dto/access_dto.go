package dto

import (
	"strings"

	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/google/uuid"
)

// CheckInRequest represents the payload for POST /api/v1/access/check-in.
type CheckInRequest struct {
	WorkerID      string  `json:"worker_id"`
	WorkerUserID  string  `json:"worker_user_id"`
	AccessPointID string  `json:"access_point_id"`
	Notes         *string `json:"notes,omitempty"`
}

// Validate parses and validates identifiers in CheckInRequest.
func (r *CheckInRequest) Validate() (workerUUID uuid.UUID, apUUID uuid.UUID, err error) {
	idStr := strings.TrimSpace(r.WorkerUserID)
	if idStr == "" {
		idStr = strings.TrimSpace(r.WorkerID)
	}
	if idStr == "" {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es requerido.")
	}

	workerUUID, err = uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de trabajador no es válido.")
	}

	apStr := strings.TrimSpace(r.AccessPointID)
	if apStr == "" {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El punto de acceso es requerido.")
	}

	apUUID, err = uuid.Parse(apStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de punto de acceso no es válido.")
	}

	return workerUUID, apUUID, nil
}

// CheckOutRequest represents the payload for POST /api/v1/access/check-out.
type CheckOutRequest struct {
	WorkerID      string `json:"worker_id"`
	WorkerUserID  string `json:"worker_user_id"`
	PlantID       string `json:"plant_id"`
	AccessPointID string `json:"access_point_id"`
}

// Validate parses and validates identifiers in CheckOutRequest.
func (r *CheckOutRequest) Validate() (workerUUID uuid.UUID, plantUUID uuid.UUID, apUUID uuid.UUID, err error) {
	idStr := strings.TrimSpace(r.WorkerUserID)
	if idStr == "" {
		idStr = strings.TrimSpace(r.WorkerID)
	}
	if idStr == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es requerido.")
	}

	workerUUID, err = uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de trabajador no es válido.")
	}

	if strings.TrimSpace(r.PlantID) != "" {
		plantUUID, err = uuid.Parse(strings.TrimSpace(r.PlantID))
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de planta no es válido.")
		}
	}

	if strings.TrimSpace(r.AccessPointID) != "" {
		apUUID, err = uuid.Parse(strings.TrimSpace(r.AccessPointID))
		if err != nil {
			return uuid.Nil, uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de punto de acceso no es válido.")
		}
	}

	return workerUUID, plantUUID, apUUID, nil
}
