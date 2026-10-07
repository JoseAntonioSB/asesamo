package access

import (
	"context"
	"fmt"
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

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
	Documents        []entity.DocumentSnapshotItem `json:"documents"`
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

	// Propagate tenant into context if not already present (fail-safe for direct invocations)
	if _, err := tenant.FromContext(ctx); err != nil && in.TenantID != uuid.Nil {
		ctx = tenant.WithTenant(ctx, in.TenantID)
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
		snapshot              entity.DocumentSnapshot
		nonNegotiableFailures []string
		expiredDocFailures    []string
		warnings              []string
	)

	// Check worker active state (both users.is_active AND contractor_workers.is_active)
	isWorkerActive := aff.IsActive
	if aff.UserIsActive || aff.WorkerIsActive {
		isWorkerActive = aff.UserIsActive && aff.WorkerIsActive
	}
	if !aff.IsActive {
		isWorkerActive = false
	}
	if !isWorkerActive {
		nonNegotiableFailures = append(nonNegotiableFailures, "El perfil del trabajador se encuentra inactivo.")
	}

	// Check contractor active state (both contractors.is_active AND plant_contractors.status == 'active')
	// Statuses 'pending', 'suspended', or 'blocked' must all trigger denied.
	if aff.ContractorID != uuid.Nil {
		if !aff.ContractorIsActive && (aff.UserIsActive || aff.WorkerIsActive) {
			nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' se encuentra inactiva.", aff.ContractorName))
		}
		if aff.PlantStatus != "active" {
			if aff.PlantStatus == "blocked" || aff.PlantStatus == "suspended" {
				nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' se encuentra %s en esta planta.", aff.ContractorName, aff.PlantStatus))
			} else {
				nonNegotiableFailures = append(nonNegotiableFailures, fmt.Sprintf("La empresa contratista '%s' no se encuentra activa en esta planta (estatus: %s).", aff.ContractorName, aff.PlantStatus))
			}
		}
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
				nonNegotiableFailures = append(nonNegotiableFailures, reason)
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
					nonNegotiableFailures = append(nonNegotiableFailures, reason)
				}
			} else if doc.ExpiresAt != nil && doc.ExpiresAt.Before(now) {
				item.IsValid = false
				reason := fmt.Sprintf("Documento '%s' vencido el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02"))
				item.FailureReason = &reason
				if req.IsRequired {
					expiredDocFailures = append(expiredDocFailures, reason)
				}
			} else {
				item.IsValid = true
				if doc.ExpiresAt != nil && doc.ExpiresAt.Before(warningThreshold) {
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

	if len(nonNegotiableFailures) > 0 {
		// Non-negotiable failures always deny access regardless of plant policy
		decisionResult = entity.AccessResultDenied
		allFailures := append(nonNegotiableFailures, expiredDocFailures...)
		combined := strings.Join(allFailures, " | ")
		denialReason = &combined
	} else if len(expiredDocFailures) > 0 {
		if plant.BlockOnExpiredDocs {
			decisionResult = entity.AccessResultDenied
			combined := strings.Join(expiredDocFailures, " | ")
			denialReason = &combined
		} else {
			decisionResult = entity.AccessResultGrantedWithWarning
			denialReason = nil
		}
	} else if len(warnings) > 0 {
		decisionResult = entity.AccessResultGrantedWithWarning
		denialReason = nil
	} else {
		decisionResult = entity.AccessResultGranted
		denialReason = nil
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
		Documents:        snapshot,
	}, nil
}
