package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EvacuationType string

const (
	EvacuationTypeDrill      EvacuationType = "drill"
	EvacuationTypeFire       EvacuationType = "fire"
	EvacuationTypeGasLeak    EvacuationType = "gas_leak"
	EvacuationTypeEarthquake EvacuationType = "earthquake"
	EvacuationTypeSpill      EvacuationType = "spill"
	EvacuationTypeOther      EvacuationType = "other"
)

type EvacuationStatus string

const (
	EvacuationStatusActive    EvacuationStatus = "active"
	EvacuationStatusAllClear  EvacuationStatus = "all_clear"
	EvacuationStatusCancelled EvacuationStatus = "cancelled"
)

type RollCallStatus string

const (
	RollCallStatusSafe         RollCallStatus = "safe"
	RollCallStatusMissing      RollCallStatus = "missing"
	RollCallStatusHospitalized RollCallStatus = "hospitalized"
)

type MusterPoint struct {
	ID                  uuid.UUID `json:"id"`
	TenantID            uuid.UUID `json:"tenant_id"`
	Code                string    `json:"code"`
	Name                string    `json:"name"`
	LocationDescription string    `json:"location_description,omitempty"`
	Capacity            int       `json:"capacity"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
}

type EvacuationRollCallRecord struct {
	ID                 uuid.UUID      `json:"id"`
	EvacuationID       uuid.UUID      `json:"evacuation_id"`
	WorkerID           uuid.UUID      `json:"worker_id"`
	MusterPointID      *uuid.UUID     `json:"muster_point_id,omitempty"`
	Status             RollCallStatus `json:"status"`
	CheckedInAt        *time.Time     `json:"checked_in_at,omitempty"`
	CheckedInByUserID  *uuid.UUID     `json:"checked_in_by_user_id,omitempty"`
	Notes              string         `json:"notes,omitempty"`
}

type EmergencyEvacuation struct {
	ID                    uuid.UUID                  `json:"id"`
	TenantID              uuid.UUID                  `json:"tenant_id"`
	EvacuationNumber      string                     `json:"evacuation_number"`
	EvacuationType        EvacuationType             `json:"evacuation_type"`
	Status                EvacuationStatus           `json:"status"`
	Description           string                     `json:"description,omitempty"`
	StartedAt             time.Time                  `json:"started_at"`
	EndedAt               *time.Time                 `json:"ended_at,omitempty"`
	EvacuationTimeSeconds int                        `json:"evacuation_time_seconds"`
	InitiatedByUserID     uuid.UUID                  `json:"initiated_by_user_id"`
	TotalHeadcount        int                        `json:"total_headcount"`
	SafeCount             int                        `json:"safe_count"`
	MissingCount          int                        `json:"missing_count"`
	RollCall              []EvacuationRollCallRecord `json:"roll_call,omitempty"`
	CreatedAt             time.Time                  `json:"created_at"`
	UpdatedAt             time.Time                  `json:"updated_at"`
}

func NewEmergencyEvacuation(
	tenantID uuid.UUID,
	evacNumber string,
	evacType EvacuationType,
	desc string,
	initiatorID uuid.UUID,
) *EmergencyEvacuation {
	now := time.Now()
	return &EmergencyEvacuation{
		ID:                uuid.New(),
		TenantID:          tenantID,
		EvacuationNumber:  evacNumber,
		EvacuationType:    evacType,
		Status:            EvacuationStatusActive,
		Description:       desc,
		StartedAt:         now,
		InitiatedByUserID: initiatorID,
		RollCall:          make([]EvacuationRollCallRecord, 0),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (e *EmergencyEvacuation) InitializeRollCall(workerIDs []uuid.UUID) error {
	e.RollCall = make([]EvacuationRollCallRecord, 0, len(workerIDs))
	for _, wid := range workerIDs {
		e.RollCall = append(e.RollCall, EvacuationRollCallRecord{
			ID:           uuid.New(),
			EvacuationID: e.ID,
			WorkerID:     wid,
			Status:       RollCallStatusMissing,
		})
	}
	e.TotalHeadcount = len(workerIDs)
	e.MissingCount = len(workerIDs)
	e.SafeCount = 0
	return nil
}

func (e *EmergencyEvacuation) CheckInWorker(workerID, musterPointID, checkedInBy uuid.UUID, notes string) error {
	found := false
	now := time.Now()

	for i := range e.RollCall {
		if e.RollCall[i].WorkerID == workerID {
			if e.RollCall[i].Status == RollCallStatusMissing {
				e.MissingCount--
				e.SafeCount++
			}
			e.RollCall[i].Status = RollCallStatusSafe
			e.RollCall[i].MusterPointID = &musterPointID
			e.RollCall[i].CheckedInAt = &now
			e.RollCall[i].CheckedInByUserID = &checkedInBy
			e.RollCall[i].Notes = notes
			found = true
			break
		}
	}

	if !found {
		// Trabajador que no estaba en el headcount inicial pero llegó al punto de reunión
		e.RollCall = append(e.RollCall, EvacuationRollCallRecord{
			ID:                uuid.New(),
			EvacuationID:      e.ID,
			WorkerID:          workerID,
			MusterPointID:     &musterPointID,
			Status:            RollCallStatusSafe,
			CheckedInAt:       &now,
			CheckedInByUserID: &checkedInBy,
			Notes:             notes,
		})
		e.TotalHeadcount++
		e.SafeCount++
	}

	e.UpdatedAt = now
	return nil
}

func (e *EmergencyEvacuation) CanDeclareAllClear() bool {
	return e.MissingCount == 0 && e.TotalHeadcount > 0
}

func (e *EmergencyEvacuation) DeclareAllClear() error {
	if !e.CanDeclareAllClear() {
		return errors.New("aún hay personal clasificado como faltante o atrapado en planta")
	}

	now := time.Now()
	e.Status = EvacuationStatusAllClear
	e.EndedAt = &now
	e.EvacuationTimeSeconds = int(now.Sub(e.StartedAt).Seconds())
	e.UpdatedAt = now
	return nil
}
