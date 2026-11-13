package incident

import (
	"context"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type ReportIncidentInput struct {
	Type             entity.IncidentType     `json:"type"`
	Severity         entity.IncidentSeverity `json:"severity"`
	Title            string                  `json:"title"`
	Description      string                  `json:"description"`
	Location         string                  `json:"location"`
	EventTimestamp   time.Time               `json:"event_timestamp"`
	InvolvedWorkerID *uuid.UUID              `json:"involved_worker_id,omitempty"`
	ContractorID     *uuid.UUID              `json:"contractor_id,omitempty"`
	EvidencePhotoURL string                  `json:"evidence_photo_url,omitempty"`
}

type ReportIncidentUseCase struct {
	repo repository.SafetyIncidentRepository
}

func NewReportIncidentUseCase(repo repository.SafetyIncidentRepository) *ReportIncidentUseCase {
	return &ReportIncidentUseCase{repo: repo}
}

func (uc *ReportIncidentUseCase) Execute(ctx context.Context, tenantID, reporterID uuid.UUID, input ReportIncidentInput) (*entity.SafetyIncident, error) {
	incidentNumber := fmt.Sprintf("INC-%s-%s", time.Now().Format("2006"), uuid.New().String()[:6])

	inc, err := entity.NewSafetyIncident(
		tenantID,
		incidentNumber,
		input.Type,
		input.Severity,
		input.Title,
		input.Description,
		input.Location,
		input.EventTimestamp,
		reporterID,
	)
	if err != nil {
		return nil, err
	}

	inc.InvolvedWorkerID = input.InvolvedWorkerID
	inc.ContractorID = input.ContractorID
	inc.EvidencePhotoURL = input.EvidencePhotoURL

	if err := uc.repo.CreateIncident(ctx, inc); err != nil {
		return nil, err
	}

	return inc, nil
}
