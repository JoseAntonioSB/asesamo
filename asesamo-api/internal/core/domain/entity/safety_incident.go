package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type IncidentType string

const (
	IncidentTypeUnsafeAct        IncidentType = "unsafe_act"
	IncidentTypeUnsafeCondition  IncidentType = "unsafe_condition"
	IncidentTypeNearMiss         IncidentType = "near_miss"
	IncidentTypeFirstAid         IncidentType = "first_aid"
	IncidentTypeMedicalTreatment IncidentType = "medical_treatment"
	IncidentTypeLostTimeInjury   IncidentType = "lost_time_injury"
	IncidentTypeFatality         IncidentType = "fatality"
)

type IncidentSeverity string

const (
	IncidentSeverityMinor        IncidentSeverity = "minor"
	IncidentSeverityModerate     IncidentSeverity = "moderate"
	IncidentSeverityMajor        IncidentSeverity = "major"
	IncidentSeverityCatastrophic IncidentSeverity = "catastrophic"
)

type IncidentStatus string

const (
	IncidentStatusReported       IncidentStatus = "reported"
	IncidentStatusInvestigating  IncidentStatus = "investigating"
	IncidentStatusActionsPending IncidentStatus = "actions_pending"
	IncidentStatusClosed         IncidentStatus = "closed"
	IncidentStatusCancelled      IncidentStatus = "cancelled"
)

type SafetyIncident struct {
	ID                   uuid.UUID        `json:"id"`
	TenantID             uuid.UUID        `json:"tenant_id"`
	IncidentNumber       string           `json:"incident_number"`
	IncidentType         IncidentType     `json:"incident_type"`
	Severity             IncidentSeverity `json:"severity"`
	Status               IncidentStatus   `json:"status"`
	Title                string           `json:"title"`
	Description          string           `json:"description"`
	Location             string           `json:"location"`
	EventTimestamp       time.Time        `json:"event_timestamp"`
	ReportedByUserID     uuid.UUID        `json:"reported_by_user_id"`
	InvolvedWorkerID     *uuid.UUID       `json:"involved_worker_id,omitempty"`
	ContractorID         *uuid.UUID       `json:"contractor_id,omitempty"`
	DaysLost             int              `json:"days_lost"`
	EvidencePhotoURL     string           `json:"evidence_photo_url,omitempty"`
	RootCauseSummary     string           `json:"root_cause_summary,omitempty"`
	ImmediateCause       string           `json:"immediate_cause,omitempty"`
	BasicCause           string           `json:"basic_cause,omitempty"`
	InvestigatedByUserID *uuid.UUID       `json:"investigated_by_user_id,omitempty"`
	InvestigatedAt       *time.Time       `json:"investigated_at,omitempty"`
	ClosedAt             *time.Time       `json:"closed_at,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

func NewSafetyIncident(
	tenantID uuid.UUID,
	incidentNumber string,
	incType IncidentType,
	severity IncidentSeverity,
	title string,
	desc string,
	location string,
	eventTime time.Time,
	reporterID uuid.UUID,
) (*SafetyIncident, error) {
	if title == "" {
		return nil, errors.New("el título del incidente es obligatorio")
	}
	if desc == "" {
		return nil, errors.New("la descripción del incidente es obligatoria")
	}

	now := time.Now()
	return &SafetyIncident{
		ID:               uuid.New(),
		TenantID:         tenantID,
		IncidentNumber:   incidentNumber,
		IncidentType:     incType,
		Severity:         severity,
		Status:           IncidentStatusReported,
		Title:            title,
		Description:      desc,
		Location:         location,
		EventTimestamp:   eventTime,
		ReportedByUserID: reporterID,
		DaysLost:         0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// IsRecordableSTPS determina si debe registrarse ante la STPS en el índice de siniestralidad
func (s *SafetyIncident) IsRecordableSTPS() bool {
	return s.IncidentType == IncidentTypeLostTimeInjury ||
		s.IncidentType == IncidentTypeFatality ||
		s.IncidentType == IncidentTypeMedicalTreatment ||
		s.DaysLost > 0
}

func (s *SafetyIncident) CompleteInvestigation(
	investigatorID uuid.UUID,
	rootCause string,
	immediateCause string,
	basicCause string,
	daysLost int,
) error {
	if s.IncidentType == IncidentTypeLostTimeInjury && daysLost <= 0 {
		return errors.New("un accidente con tiempo perdido (LTI) debe registrar al menos 1 día de incapacidad")
	}

	now := time.Now()
	s.InvestigatedByUserID = &investigatorID
	s.InvestigatedAt = &now
	s.RootCauseSummary = rootCause
	s.ImmediateCause = immediateCause
	s.BasicCause = basicCause
	s.DaysLost = daysLost
	s.Status = IncidentStatusInvestigating
	s.UpdatedAt = now

	return nil
}

func (s *SafetyIncident) CloseIncident() error {
	if s.Status == IncidentStatusClosed {
		return errors.New("el incidente ya se encuentra cerrado")
	}
	now := time.Now()
	s.Status = IncidentStatusClosed
	s.ClosedAt = &now
	s.UpdatedAt = now
	return nil
}
