package access

import (
	"context"
	"fmt"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// CheckInInput contains parameters for gate verification.
type CheckInInput struct {
	TenantID      uuid.UUID  `json:"tenant_id"`
	WorkerUserID  uuid.UUID  `json:"worker_user_id"`
	AccessPointID uuid.UUID  `json:"access_point_id"`
	ProcessedByID *uuid.UUID `json:"processed_by_id,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

// CheckInOutput contains the gate verification result and full compliance breakdown.
type CheckInOutput struct {
	AccessLogID      uuid.UUID                     `json:"access_log_id"`
	Result           entity.AccessResult           `json:"result"`
	DenialReason     *string                       `json:"denial_reason"`
	CheckedInAt      time.Time                     `json:"checked_in_at"`
	Worker           WorkerSummary                 `json:"worker"`
	Contractor       ContractorSummary             `json:"contractor"`
	Plant            PlantSummary                  `json:"plant"`
	AccessPoint      AccessPointSummary            `json:"access_point"`
	DocumentSnapshot []entity.DocumentSnapshotItem `json:"document_snapshot"`
}

type WorkerSummary struct {
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	CURP        *string   `json:"curp,omitempty"`
	IMSSNumber  *string   `json:"imss_number,omitempty"`
	JobTitle    *string   `json:"job_title,omitempty"`
	AvatarR2Key *string   `json:"avatar_r2_key,omitempty"`
}

type ContractorSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RFC         *string   `json:"rfc,omitempty"`
	REPSENumber *string   `json:"repse_number,omitempty"`
	Status      string    `json:"status"`
}

type PlantSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AccessPointSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// CheckInUseCase coordinates access gate check-in verification.
type CheckInUseCase struct {
	accessPointRepo repository.AccessPointRepository
	plantRepo       repository.PlantRepository
	complianceRepo  repository.ComplianceRepository
	accessLogRepo   repository.AccessLogRepository
}

// NewCheckInUseCase creates a new CheckInUseCase.
func NewCheckInUseCase(
	apRepo repository.AccessPointRepository,
	plantRepo repository.PlantRepository,
	compRepo repository.ComplianceRepository,
	logRepo repository.AccessLogRepository,
) *CheckInUseCase {
	return &CheckInUseCase{
		accessPointRepo: apRepo,
		plantRepo:       plantRepo,
		complianceRepo:  compRepo,
		accessLogRepo:   logRepo,
	}
}

// Execute evaluates compliance and records the check-in event.
func (uc *CheckInUseCase) Execute(ctx context.Context, in CheckInInput) (*CheckInOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}
	if in.AccessPointID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El punto de acceso es obligatorio.")
	}

	// 1. Verify Access Point
	ap, err := uc.accessPointRepo.GetByID(ctx, in.AccessPointID)
	if err != nil {
		return nil, err
	}
	if ap.TenantID != in.TenantID {
		return nil, domainErrors.ErrAccessPointNotFound.WithMessage("El punto de acceso no pertenece al tenant.")
	}

	// 2. Retrieve Plant
	plant, err := uc.plantRepo.GetByID(ctx, ap.PlantID)
	if err != nil {
		return nil, err
	}

	// 3. Retrieve Worker Affiliation & Profile
	aff, err := uc.complianceRepo.GetWorkerAffiliation(ctx, plant.ID, in.WorkerUserID)
	if err != nil {
		return nil, err
	}

	// 4. Retrieve Document Requirements
	requirements, err := uc.complianceRepo.GetPlantDocumentRequirements(ctx, in.TenantID, plant.ID)
	if err != nil {
		return nil, fmt.Errorf("usecase/check_in: failed to get document requirements: %w", err)
	}

	// 5. Retrieve Uploaded Documents
	workerDocs, err := uc.complianceRepo.GetWorkerDocuments(ctx, in.WorkerUserID)
	if err != nil {
		return nil, fmt.Errorf("usecase/check_in: failed to get worker documents: %w", err)
	}

	var contractorDocs []entity.DocumentWithTypeCode
	if aff.ContractorID != uuid.Nil {
		contractorDocs, err = uc.complianceRepo.GetContractorDocuments(ctx, aff.ContractorID)
		if err != nil {
			return nil, fmt.Errorf("usecase/check_in: failed to get contractor documents: %w", err)
		}
	}

	// Index docs by document_type_code
	docsByCode := make(map[string]entity.DocumentWithTypeCode)
	for _, d := range workerDocs {
		if _, exists := docsByCode[d.DocumentTypeCode]; !exists {
			docsByCode[d.DocumentTypeCode] = d
		}
	}
	for _, d := range contractorDocs {
		if _, exists := docsByCode[d.DocumentTypeCode]; !exists {
			docsByCode[d.DocumentTypeCode] = d
		}
	}

	// 6. Evaluate Compliance
	now := time.Now().UTC()
	warningThreshold := now.AddDate(0, 0, 30)

	var (
		snapshot        entity.DocumentSnapshot
		failures        []string
		warnings        []string
		hasFatalFailure bool
		hasWarning      bool
	)

	// Check worker active state
	if !aff.IsActive {
		hasFatalFailure = true
		failures = append(failures, "El perfil del trabajador se encuentra inactivo.")
	}

	// Check contractor status in this plant
	if aff.PlantStatus == "blocked" || aff.PlantStatus == "suspended" {
		hasFatalFailure = true
		failures = append(failures, fmt.Sprintf("La empresa contratista '%s' se encuentra %s en esta planta.", aff.ContractorName, aff.PlantStatus))
	}

	for _, req := range requirements {
		item := entity.DocumentSnapshotItem{
			DocumentTypeCode: req.DocumentTypeCode,
			DocumentName:     req.DocumentName,
			AppliesTo:        req.AppliesTo,
			IsRequired:       req.IsRequired,
		}

		doc, found := docsByCode[req.DocumentTypeCode]
		if !found {
			item.Status = "missing"
			item.IsValid = false
			reason := fmt.Sprintf("Documento '%s' faltante.", req.DocumentName)
			item.FailureReason = &reason
			if req.IsRequired {
				hasFatalFailure = true
				failures = append(failures, reason)
			}
		} else {
			item.Status = string(doc.Status)
			if doc.ExpiresAt != nil {
				expStr := doc.ExpiresAt.Format("2006-01-02")
				item.ExpiresAt = &expStr
			}

			if doc.Status != entity.DocumentStatusApproved {
				item.IsValid = false
				reason := fmt.Sprintf("Documento '%s' no aprobado (estatus: %s).", req.DocumentName, doc.Status)
				item.FailureReason = &reason
				if req.IsRequired {
					hasFatalFailure = true
					failures = append(failures, reason)
				}
			} else if doc.ExpiresAt != nil && doc.ExpiresAt.Before(now) {
				item.IsValid = false
				reason := fmt.Sprintf("Documento '%s' vencido el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02"))
				item.FailureReason = &reason
				if req.IsRequired {
					hasFatalFailure = true
					failures = append(failures, reason)
				}
			} else {
				item.IsValid = true
				if doc.ExpiresAt != nil && doc.ExpiresAt.Before(warningThreshold) {
					hasWarning = true
					warnings = append(warnings, fmt.Sprintf("Documento '%s' próximo a vencer el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02")))
				}
			}
		}

		snapshot = append(snapshot, item)
	}

	// 7. Determine Result
	var (
		decisionResult entity.AccessResult
		denialReason   *string
	)

	if hasFatalFailure {
		if plant.BlockOnExpiredDocs {
			decisionResult = entity.AccessResultDenied
			combined := strings.Join(failures, " | ")
			denialReason = &combined
		} else {
			decisionResult = entity.AccessResultGrantedWithWarning
			combined := strings.Join(failures, " | ")
			denialReason = &combined
		}
	} else if hasWarning {
		decisionResult = entity.AccessResultGrantedWithWarning
	} else {
		decisionResult = entity.AccessResultGranted
	}

	// 8. Create Access Log Record
	logID := uuid.New()
	var contractorIDPtr *uuid.UUID
	if aff.ContractorID != uuid.Nil {
		contractorIDPtr = &aff.ContractorID
	}

	logRecord := &entity.AccessLog{
		ID:               logID,
		TenantID:         in.TenantID,
		PlantID:          plant.ID,
		AccessPointID:    ap.ID,
		UserID:           in.WorkerUserID,
		ContractorID:     contractorIDPtr,
		ProcessedBy:      in.ProcessedByID,
		CheckedInAt:      now,
		Result:           decisionResult,
		DenialReason:     denialReason,
		DocumentSnapshot: snapshot,
		Notes:            in.Notes,
		CreatedAt:        now,
	}

	if err := uc.accessLogRepo.Create(ctx, logRecord); err != nil {
		return nil, fmt.Errorf("usecase/check_in: failed to create access log: %w", err)
	}

	// 9. Return Response
	return &CheckInOutput{
		AccessLogID:  logID,
		Result:       decisionResult,
		DenialReason: denialReason,
		CheckedInAt:  now,
		Worker: WorkerSummary{
			UserID:      aff.WorkerID,
			FullName:    aff.FullName,
			CURP:        aff.CURP,
			IMSSNumber:  aff.IMSSNumber,
			JobTitle:    aff.JobTitle,
			AvatarR2Key: aff.AvatarR2Key,
		},
		Contractor: ContractorSummary{
			ID:          aff.ContractorID,
			Name:        aff.ContractorName,
			RFC:         aff.ContractorRFC,
			REPSENumber: aff.REPSENumber,
			Status:      aff.PlantStatus,
		},
		Plant: PlantSummary{
			ID:   plant.ID,
			Name: plant.Name,
		},
		AccessPoint: AccessPointSummary{
			ID:   ap.ID,
			Name: ap.Name,
		},
		DocumentSnapshot: snapshot,
	}, nil
}
