package inspection

import (
	"context"
	"errors"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type SubmitAnswerInput struct {
	TemplateItemID   uuid.UUID                   `json:"template_item_id"`
	Status           entity.ItemComplianceStatus `json:"status"`
	FindingNotes     string                      `json:"finding_notes,omitempty"`
	EvidencePhotoURL string                      `json:"evidence_photo_url,omitempty"`
}

type CompleteInspectionResult struct {
	InspectionID    uuid.UUID                     `json:"inspection_id"`
	ScorePercentage float64                       `json:"score_percentage"`
	Status          entity.InspectionStatus       `json:"status"`
	RequiresCAPA    bool                          `json:"requires_capa"`
	GeneratedCAPAs  []entity.InspectionActionPlan `json:"generated_capas"`
}

type CompleteInspectionUseCase struct {
	repo repository.SafetyInspectionRepository
}

func NewCompleteInspectionUseCase(repo repository.SafetyInspectionRepository) *CompleteInspectionUseCase {
	return &CompleteInspectionUseCase{repo: repo}
}

func (uc *CompleteInspectionUseCase) Execute(ctx context.Context, tenantID, inspectionID uuid.UUID, answers []SubmitAnswerInput) (*CompleteInspectionResult, error) {
	insp, err := uc.repo.GetInspectionByID(ctx, tenantID, inspectionID)
	if err != nil {
		return nil, err
	}
	if insp == nil {
		return nil, errors.New("inspección no encontrada")
	}

	_, templateItems, err := uc.repo.GetTemplateByID(ctx, tenantID, insp.TemplateID)
	if err != nil {
		return nil, err
	}

	itemMap := make(map[uuid.UUID]entity.InspectionTemplateItem)
	for _, itm := range templateItems {
		itemMap[itm.ID] = itm
	}

	insp.Answers = make([]entity.InspectionItemAnswer, 0, len(answers))
	for _, ans := range answers {
		tmplItem, exists := itemMap[ans.TemplateItemID]
		weight := 1
		critical := false
		if exists {
			weight = tmplItem.Weight
			critical = tmplItem.CriticalRisk
		}

		err := insp.AddAnswer(entity.InspectionItemAnswer{
			TemplateItemID:   ans.TemplateItemID,
			Status:           ans.Status,
			Weight:           weight,
			CriticalRisk:     critical,
			FindingNotes:     ans.FindingNotes,
			EvidencePhotoURL: ans.EvidencePhotoURL,
		})
		if err != nil {
			return nil, err
		}
	}

	score, requiresCAPA := insp.CalculateScoreAndEvaluate()

	var capas []entity.InspectionActionPlan
	if requiresCAPA {
		capas = insp.GenerateRequiredCAPAs(time.Now().AddDate(0, 0, 7)) // 7 días límite default
	}

	if err := uc.repo.CompleteInspection(ctx, insp, capas); err != nil {
		return nil, err
	}

	return &CompleteInspectionResult{
		InspectionID:    insp.ID,
		ScorePercentage: score,
		Status:          insp.Status,
		RequiresCAPA:    requiresCAPA,
		GeneratedCAPAs:  capas,
	}, nil
}
