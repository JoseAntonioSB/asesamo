package entity

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

type InspectionType string

const (
	InspectionTypeRoutine    InspectionType = "routine"
	InspectionTypePlanned    InspectionType = "planned"
	InspectionTypeSpecial    InspectionType = "special"
	InspectionTypeRegulatory InspectionType = "regulatory"
)

type InspectionStatus string

const (
	InspectionStatusDraft        InspectionStatus = "draft"
	InspectionStatusInProgress   InspectionStatus = "in_progress"
	InspectionStatusCompleted    InspectionStatus = "completed"
	InspectionStatusRequiresCAPA InspectionStatus = "requires_capa"
	InspectionStatusCancelled    InspectionStatus = "cancelled"
)

type ItemComplianceStatus string

const (
	ItemComplianceCompliant     ItemComplianceStatus = "compliant"
	ItemComplianceNonCompliant  ItemComplianceStatus = "non_compliant"
	ItemComplianceNotApplicable ItemComplianceStatus = "not_applicable"
)

type FindingSeverity string

const (
	FindingSeverityLow      FindingSeverity = "low"
	FindingSeverityMedium   FindingSeverity = "medium"
	FindingSeverityHigh     FindingSeverity = "high"
	FindingSeverityCritical FindingSeverity = "critical"
)

type CAPAStatus string

const (
	CAPAStatusOpen       CAPAStatus = "open"
	CAPAStatusInProgress CAPAStatus = "in_progress"
	CAPAStatusResolved   CAPAStatus = "resolved"
	CAPAStatusVerified   CAPAStatus = "verified"
	CAPAStatusOverdue    CAPAStatus = "overdue"
)

type InspectionTemplate struct {
	ID                 uuid.UUID `json:"id"`
	TenantID           uuid.UUID `json:"tenant_id"`
	Title              string    `json:"title"`
	Description        string    `json:"description,omitempty"`
	Category           string    `json:"category"`
	NormativeReference string    `json:"normative_reference,omitempty"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type InspectionTemplateItem struct {
	ID            uuid.UUID `json:"id"`
	TemplateID    uuid.UUID `json:"template_id"`
	Category      string    `json:"category"`
	Question      string    `json:"question"`
	GuidanceNotes string    `json:"guidance_notes,omitempty"`
	Weight        int       `json:"weight"`
	CriticalRisk  bool      `json:"critical_risk"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
}

type InspectionItemAnswer struct {
	TemplateItemID   uuid.UUID            `json:"template_item_id"`
	Status           ItemComplianceStatus `json:"status"`
	Weight           int                  `json:"weight"`
	CriticalRisk     bool                 `json:"critical_risk"`
	FindingNotes     string               `json:"finding_notes,omitempty"`
	EvidencePhotoURL string               `json:"evidence_photo_url,omitempty"`
}

type InspectionActionPlan struct {
	ID                 uuid.UUID       `json:"id"`
	TenantID           uuid.UUID       `json:"tenant_id"`
	InspectionID       uuid.UUID       `json:"inspection_id"`
	TemplateItemID     uuid.UUID       `json:"template_item_id"`
	Description        string          `json:"description"`
	Severity           FindingSeverity `json:"severity"`
	Status             CAPAStatus      `json:"status"`
	AssignedToUserID   *uuid.UUID      `json:"assigned_to_user_id,omitempty"`
	ContractorID       *uuid.UUID      `json:"contractor_id,omitempty"`
	DueDate            time.Time       `json:"due_date"`
	ResolvedAt         *time.Time      `json:"resolved_at,omitempty"`
	VerifiedAt         *time.Time      `json:"verified_at,omitempty"`
	ResolutionNotes    string          `json:"resolution_notes,omitempty"`
	ResolutionPhotoURL string          `json:"resolution_photo_url,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type SafetyInspection struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	TemplateID        uuid.UUID              `json:"template_id"`
	InspectionNumber  string                 `json:"inspection_number"`
	InspectionType    InspectionType         `json:"inspection_type"`
	Status            InspectionStatus       `json:"status"`
	AreaOrLocation    string                 `json:"area_or_location"`
	ContractorID      *uuid.UUID             `json:"contractor_id,omitempty"`
	InspectorID       uuid.UUID              `json:"inspector_id"`
	ScorePercentage   *float64               `json:"score_percentage,omitempty"`
	TotalItems        int                    `json:"total_items"`
	CompliantItems    int                    `json:"compliant_items"`
	NonCompliantItems int                    `json:"non_compliant_items"`
	NAItems           int                    `json:"na_items"`
	Notes             string                 `json:"notes,omitempty"`
	Answers           []InspectionItemAnswer `json:"answers,omitempty"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

func NewSafetyInspection(tenantID, templateID uuid.UUID, number string, inspType InspectionType, location string, inspectorID uuid.UUID) *SafetyInspection {
	now := time.Now()
	return &SafetyInspection{
		ID:               uuid.New(),
		TenantID:         tenantID,
		TemplateID:       templateID,
		InspectionNumber: number,
		InspectionType:   inspType,
		Status:           InspectionStatusDraft,
		AreaOrLocation:   location,
		InspectorID:      inspectorID,
		Answers:          make([]InspectionItemAnswer, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (s *SafetyInspection) AddAnswer(ans InspectionItemAnswer) error {
	if ans.TemplateItemID == uuid.Nil {
		return errors.New("template_item_id no puede ser nulo")
	}
	if ans.Weight <= 0 {
		ans.Weight = 1
	}
	s.Answers = append(s.Answers, ans)
	return nil
}

// CalculateScoreAndEvaluate computa el porcentaje obtenido ponderado y define si requiere CAPA
func (s *SafetyInspection) CalculateScoreAndEvaluate() (float64, bool) {
	var totalWeightApplicable int
	var compliantWeight int
	var hasCriticalFailure bool

	s.TotalItems = len(s.Answers)
	s.CompliantItems = 0
	s.NonCompliantItems = 0
	s.NAItems = 0

	for _, ans := range s.Answers {
		switch ans.Status {
		case ItemComplianceCompliant:
			s.CompliantItems++
			totalWeightApplicable += ans.Weight
			compliantWeight += ans.Weight
		case ItemComplianceNonCompliant:
			s.NonCompliantItems++
			totalWeightApplicable += ans.Weight
			if ans.CriticalRisk {
				hasCriticalFailure = true
			}
		case ItemComplianceNotApplicable:
			s.NAItems++
		}
	}

	var score float64
	if totalWeightApplicable > 0 {
		score = (float64(compliantWeight) / float64(totalWeightApplicable)) * 100.0
		score = math.Round(score*100) / 100 // 2 decimales
	} else {
		score = 100.0
	}

	s.ScorePercentage = &score

	requiresCAPA := hasCriticalFailure || s.NonCompliantItems > 0
	if requiresCAPA {
		s.Status = InspectionStatusRequiresCAPA
	} else {
		s.Status = InspectionStatusCompleted
	}

	now := time.Now()
	s.CompletedAt = &now
	s.UpdatedAt = now

	return score, requiresCAPA
}

// GenerateRequiredCAPAs crea las acciones correctivas iniciales para cada ítem No Conforme
func (s *SafetyInspection) GenerateRequiredCAPAs(defaultDueDate time.Time) []InspectionActionPlan {
	capas := make([]InspectionActionPlan, 0)
	now := time.Now()

	for _, ans := range s.Answers {
		if ans.Status != ItemComplianceNonCompliant {
			continue
		}

		severity := FindingSeverityMedium
		if ans.CriticalRisk {
			severity = FindingSeverityCritical
		}

		desc := ans.FindingNotes
		if desc == "" {
			desc = "Desvío detectado durante inspección"
		}

		capas = append(capas, InspectionActionPlan{
			ID:             uuid.New(),
			TenantID:       s.TenantID,
			InspectionID:   s.ID,
			TemplateItemID: ans.TemplateItemID,
			Description:    desc,
			Severity:       severity,
			Status:         CAPAStatusOpen,
			ContractorID:   s.ContractorID,
			DueDate:        defaultDueDate,
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}

	return capas
}
