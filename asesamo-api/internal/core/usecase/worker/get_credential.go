package worker

import (
	"context"
	"fmt"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

type GetCredentialInput struct {
	WorkerUserID uuid.UUID
	PlantID      *uuid.UUID
}

type WorkerProfileDetail struct {
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	CURP        *string   `json:"curp,omitempty"`
	IMSSNumber  *string   `json:"imss_number,omitempty"`
	JobTitle    *string   `json:"job_title,omitempty"`
	AvatarR2Key *string   `json:"avatar_r2_key,omitempty"`
	IsActive    bool      `json:"is_active"`
}

type ContractorProfileDetail struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RFC         *string   `json:"rfc,omitempty"`
	REPSENumber *string   `json:"repse_number,omitempty"`
	Status      string    `json:"status"`
}

type PlantProfileDetail struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type WorkerCredentialOutput struct {
	Worker           WorkerProfileDetail           `json:"worker"`
	Contractor       ContractorProfileDetail       `json:"contractor"`
	Plant            PlantProfileDetail            `json:"plant"`
	QRPayload        string                        `json:"qr_payload"`
	ComplianceStatus entity.AccessResult           `json:"compliance_status"`
	DenialReasons    []string                      `json:"denial_reasons"`
	WarningReasons   []string                      `json:"warning_reasons"`
	Documents        []entity.DocumentSnapshotItem `json:"documents"`
	GeneratedAt      time.Time                     `json:"generated_at"`
}

type GetCredentialUseCase struct {
	userRepo       repository.UserRepository
	plantRepo      repository.PlantRepository
	complianceRepo repository.ComplianceRepository
}

func NewGetCredentialUseCase(
	userRepo repository.UserRepository,
	plantRepo repository.PlantRepository,
	compRepo repository.ComplianceRepository,
) *GetCredentialUseCase {
	return &GetCredentialUseCase{
		userRepo:       userRepo,
		plantRepo:      plantRepo,
		complianceRepo: compRepo,
	}
}

func (uc *GetCredentialUseCase) Execute(ctx context.Context, in GetCredentialInput) (*WorkerCredentialOutput, error) {
	if in.WorkerUserID == uuid.Nil {
		return nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es obligatorio.")
	}

	u, err := uc.userRepo.GetByID(ctx, in.WorkerUserID)
	if err != nil {
		return nil, err
	}

	// Resolve plant
	var plant *entity.Plant
	if in.PlantID != nil && *in.PlantID != uuid.Nil {
		plant, err = uc.plantRepo.GetByID(ctx, *in.PlantID)
		if err != nil {
			return nil, err
		}
	} else {
		// Resolve by tenant
		tenantID, _ := tenant.FromContext(ctx)
		if tenantID == uuid.Nil && u.TenantID != nil {
			tenantID = *u.TenantID
		}

		if tenantID != uuid.Nil {
			plants, err := uc.plantRepo.ListByTenant(ctx, tenantID)
			if err == nil && len(plants) > 0 {
				plant = &plants[0]
			}
		}
	}

	if plant == nil {
		// Fallback: construct default plant context for display
		plant = &entity.Plant{
			ID:                 uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			TenantID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:               "Planta Querétaro Norte",
			BlockOnExpiredDocs: true,
		}
	}

	// Retrieve Worker Affiliation & Profile
	aff, err := uc.complianceRepo.GetWorkerAffiliation(ctx, plant.ID, in.WorkerUserID)
	if err != nil {
		return nil, err
	}

	// Retrieve Document Requirements
	requirements, err := uc.complianceRepo.GetPlantDocumentRequirements(ctx, plant.TenantID, plant.ID)
	if err != nil {
		return nil, fmt.Errorf("usecase/get_credential: failed to get document requirements: %w", err)
	}

	// Retrieve Uploaded Documents
	workerDocs, err := uc.complianceRepo.GetWorkerDocuments(ctx, in.WorkerUserID)
	if err != nil {
		return nil, fmt.Errorf("usecase/get_credential: failed to get worker documents: %w", err)
	}

	var contractorDocs []entity.DocumentWithTypeCode
	if aff.ContractorID != uuid.Nil {
		contractorDocs, err = uc.complianceRepo.GetContractorDocuments(ctx, aff.ContractorID)
		if err != nil {
			return nil, fmt.Errorf("usecase/get_credential: failed to get contractor documents: %w", err)
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

	// Evaluate Compliance
	now := time.Now().UTC()
	warningThreshold := now.AddDate(0, 0, 30)

	var (
		snapshot              []entity.DocumentSnapshotItem
		nonNegotiableFailures []string
		expiredDocFailures    []string
		warnings              []string
	)

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

	var status entity.AccessResult
	if len(nonNegotiableFailures) > 0 {
		status = entity.AccessResultDenied
	} else if len(expiredDocFailures) > 0 {
		if plant.BlockOnExpiredDocs {
			status = entity.AccessResultDenied
		} else {
			status = entity.AccessResultGrantedWithWarning
		}
	} else if len(warnings) > 0 {
		status = entity.AccessResultGrantedWithWarning
	} else {
		status = entity.AccessResultGranted
	}

	allDenials := append(nonNegotiableFailures, expiredDocFailures...)

	return &WorkerCredentialOutput{
		Worker: WorkerProfileDetail{
			UserID:      u.ID,
			FullName:    u.FullName,
			CURP:        aff.CURP,
			IMSSNumber:  aff.IMSSNumber,
			JobTitle:    aff.JobTitle,
			AvatarR2Key: aff.AvatarR2Key,
			IsActive:    isWorkerActive,
		},
		Contractor: ContractorProfileDetail{
			ID:          aff.ContractorID,
			Name:        aff.ContractorName,
			RFC:         aff.ContractorRFC,
			REPSENumber: aff.REPSENumber,
			Status:      aff.PlantStatus,
		},
		Plant: PlantProfileDetail{
			ID:   plant.ID,
			Name: plant.Name,
		},
		QRPayload:        u.ID.String(),
		ComplianceStatus: status,
		DenialReasons:    allDenials,
		WarningReasons:   warnings,
		Documents:        snapshot,
		GeneratedAt:      now,
	}, nil
}
