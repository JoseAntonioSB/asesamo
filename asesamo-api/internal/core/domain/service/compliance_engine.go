package service

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

// ComplianceEngine encapsulates the Mexican STPS regulatory compliance validation rules (Candado Inteligente).
type ComplianceEngine struct {
	complianceRepo repository.ComplianceRepository
}

// NewComplianceEngine constructs a new compliance engine instance.
func NewComplianceEngine(complianceRepo repository.ComplianceRepository) *ComplianceEngine {
	return &ComplianceEngine{
		complianceRepo: complianceRepo,
	}
}

// EvaluateWorker evaluates all Candado Inteligente compliance locks for a specific worker and permit type.
func (e *ComplianceEngine) EvaluateWorker(
	ctx context.Context,
	plantID uuid.UUID,
	contractorID uuid.UUID,
	permitType entity.PermitType,
	workerUserID uuid.UUID,
) (*entity.WorkerComplianceSnapshot, error) {
	now := time.Now().UTC()
	snapshot := &entity.WorkerComplianceSnapshot{
		EvaluatedAt: now,
	}

	// ------------------------------------------------------------------------
	// Lock 1: Contractor Active in Plant & Worker Active Profile
	// ------------------------------------------------------------------------
	aff, err := e.complianceRepo.GetWorkerAffiliation(ctx, plantID, workerUserID)
	if err != nil {
		return nil, fmt.Errorf("candado/evaluate: failed to query worker affiliation: %w", err)
	}

	snapshot.ContractorStatus = aff.PlantStatus
	if !aff.UserIsActive || !aff.WorkerIsActive {
		reason := "El perfil del trabajador se encuentra inactivo."
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrCandadoComplianceFailed.WithMessage(reason)
	}

	if aff.ContractorID != contractorID {
		reason := "El trabajador no pertenece a la empresa contratista asignada al permiso."
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrCandadoComplianceFailed.WithMessage(reason)
	}

	if !aff.ContractorIsActive || aff.PlantStatus != "active" {
		reason := fmt.Sprintf("La empresa contratista '%s' no se encuentra activa o está suspendida en esta planta.", aff.ContractorName)
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrContractorInactive.WithMessage(reason)
	}

	// ------------------------------------------------------------------------
	// Lock 2: Medical Fitness Certificate (Certificado Médico de Aptitud)
	// ------------------------------------------------------------------------
	workerDocs, err := e.complianceRepo.GetWorkerDocuments(ctx, workerUserID)
	if err != nil {
		return nil, fmt.Errorf("candado/evaluate: failed to query worker documents: %w", err)
	}

	var (
		medicalDoc *entity.DocumentWithTypeCode
		dc3Doc     *entity.DocumentWithTypeCode
		genericDC3 *entity.DocumentWithTypeCode
	)

	for i := range workerDocs {
		d := &workerDocs[i]
		docCode := strings.ToUpper(strings.TrimSpace(d.DocumentTypeCode))

		if docCode == "MEDICAL" && medicalDoc == nil {
			medicalDoc = d
		}

		if e.matchesHazardDC3(permitType, d) && dc3Doc == nil {
			dc3Doc = d
		} else if docCode == "DC3" && !e.hasConflictingHazard(permitType, d) && genericDC3 == nil {
			genericDC3 = d
		}
	}

	// If no specific hazard DC3 was found, use generic DC3 fallback if available
	if dc3Doc == nil && genericDC3 != nil {
		dc3Doc = genericDC3
	}

	// Validate Medical Certificate
	if medicalDoc == nil {
		reason := "El trabajador no cuenta con certificado médico de aptitud (documento faltante)."
		snapshot.MedicalStatus = "missing"
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrWorkerMissingMedical.WithMessage(reason)
	}

	snapshot.MedicalStatus = string(medicalDoc.Status)
	snapshot.MedicalExpiresAt = medicalDoc.ExpiresAt

	if medicalDoc.Status != entity.DocumentStatusApproved {
		reason := fmt.Sprintf("El certificado médico de aptitud no está aprobado (estatus: %s).", medicalDoc.Status)
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrCandadoComplianceFailed.WithMessage(reason)
	}

	if medicalDoc.ExpiresAt != nil && medicalDoc.ExpiresAt.Before(now) {
		reason := fmt.Sprintf("El certificado médico de aptitud se encuentra vencido desde el %s.", medicalDoc.ExpiresAt.Format("2006-01-02"))
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrWorkerMissingMedical.WithMessage(reason)
	}

	// ------------------------------------------------------------------------
	// Lock 3: Hazard-Specific DC-3 Certificate
	// ------------------------------------------------------------------------
	if dc3Doc == nil {
		reason := e.hazardDC3MissingMessage(permitType)
		snapshot.DC3Status = "missing"
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrWorkerMissingDC3.WithMessage(reason)
	}

	snapshot.DC3Status = string(dc3Doc.Status)
	snapshot.DC3Code = dc3Doc.DocumentTypeCode
	snapshot.DC3ExpiresAt = dc3Doc.ExpiresAt

	if dc3Doc.Status != entity.DocumentStatusApproved {
		reason := fmt.Sprintf("La constancia DC-3 para este riesgo no se encuentra aprobada (estatus: %s).", dc3Doc.Status)
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrCandadoComplianceFailed.WithMessage(reason)
	}

	if dc3Doc.ExpiresAt != nil && dc3Doc.ExpiresAt.Before(now) {
		reason := fmt.Sprintf("La constancia DC-3 para este riesgo se encuentra vencida desde el %s.", dc3Doc.ExpiresAt.Format("2006-01-02"))
		snapshot.IsCompliant = false
		snapshot.BlockReason = &reason
		return snapshot, domainErrors.ErrWorkerMissingDC3.WithMessage(reason)
	}

	// All compliance checks passed
	snapshot.IsCompliant = true
	snapshot.BlockReason = nil
	return snapshot, nil
}

// matchesHazardDC3 checks if a document corresponds to the required hazard standard.
func (e *ComplianceEngine) matchesHazardDC3(permitType entity.PermitType, doc *entity.DocumentWithTypeCode) bool {
	code := strings.ToUpper(strings.TrimSpace(doc.DocumentTypeCode))
	name := strings.ToLower(doc.DocumentName)

	switch permitType {
	case entity.PermitTypeHeights:
		return code == "DC3_ALTURAS" || code == "DC3_NOM009" ||
			(code == "DC3" && (strings.Contains(name, "altur") || strings.Contains(name, "009")))

	case entity.PermitTypeConfinedSpace:
		return code == "DC3_ESPACIOS_CONFINADOS" || code == "DC3_CONFINED" || code == "DC3_NOM033" ||
			(code == "DC3" && (strings.Contains(name, "confin") || strings.Contains(name, "033")))

	case entity.PermitTypeHotWork:
		return code == "DC3_CORTE_SOLDADURA" || code == "DC3_HOT_WORK" || code == "DC3_NOM027" ||
			(code == "DC3" && (strings.Contains(name, "solda") || strings.Contains(name, "corte") || strings.Contains(name, "calient") || strings.Contains(name, "027")))

	case entity.PermitTypeElectricalLOTO:
		return code == "DC3_LOTO" || code == "DC3_ELECTRICAL" || code == "DC3_NOM029" ||
			(code == "DC3" && (strings.Contains(name, "electr") || strings.Contains(name, "loto") || strings.Contains(name, "029")))

	default:
		return code == "DC3"
	}
}

// hasConflictingHazard checks if a generic DC3 document is labeled with a DIFFERENT hazard norm.
func (e *ComplianceEngine) hasConflictingHazard(permitType entity.PermitType, doc *entity.DocumentWithTypeCode) bool {
	name := strings.ToLower(doc.DocumentName)

	switch permitType {
	case entity.PermitTypeHeights:
		return strings.Contains(name, "confin") || strings.Contains(name, "033") ||
			strings.Contains(name, "solda") || strings.Contains(name, "027") ||
			strings.Contains(name, "electr") || strings.Contains(name, "029")
	case entity.PermitTypeConfinedSpace:
		return strings.Contains(name, "altur") || strings.Contains(name, "009") ||
			strings.Contains(name, "solda") || strings.Contains(name, "027") ||
			strings.Contains(name, "electr") || strings.Contains(name, "029")
	case entity.PermitTypeHotWork:
		return strings.Contains(name, "altur") || strings.Contains(name, "009") ||
			strings.Contains(name, "confin") || strings.Contains(name, "033") ||
			strings.Contains(name, "electr") || strings.Contains(name, "029")
	case entity.PermitTypeElectricalLOTO:
		return strings.Contains(name, "altur") || strings.Contains(name, "009") ||
			strings.Contains(name, "confin") || strings.Contains(name, "033") ||
			strings.Contains(name, "solda") || strings.Contains(name, "027")
	default:
		return false
	}
}

// hazardDC3MissingMessage generates the localized Mexican STPS regulation error message.
func (e *ComplianceEngine) hazardDC3MissingMessage(permitType entity.PermitType) string {
	switch permitType {
	case entity.PermitTypeHeights:
		return "El trabajador no cuenta con DC-3 vigente para trabajos en alturas (NOM-009)."
	case entity.PermitTypeConfinedSpace:
		return "El trabajador no cuenta con DC-3 vigente para espacios confinados (NOM-033)."
	case entity.PermitTypeHotWork:
		return "El trabajador no cuenta con DC-3 vigente para corte y soldadura (NOM-027)."
	case entity.PermitTypeElectricalLOTO:
		return "El trabajador no cuenta con DC-3 vigente para trabajos eléctricos y LOTO (NOM-029)."
	default:
		return "El trabajador no cuenta con constancia DC-3 vigente para el tipo de riesgo."
	}
}
