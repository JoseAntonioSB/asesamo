package entity

import (
	"time"

	"github.com/google/uuid"
)

// PermitType defines the STPS regulated high-risk hazard classification.
type PermitType string

const (
	PermitTypeHeights        PermitType = "heights"        // NOM-009-STPS-2011: Trabajos en altura
	PermitTypeConfinedSpace  PermitType = "confined_space" // NOM-033-STPS-2015: Espacios confinados
	PermitTypeHotWork        PermitType = "hot_work"       // NOM-027-STPS-2008: Soldadura y corte
	PermitTypeElectricalLOTO PermitType = "electrical_loto" // NOM-029-STPS-2011: Electricidad y LOTO
)

// IsValid validates whether the permit type is a known enum.
func (pt PermitType) IsValid() bool {
	switch pt {
	case PermitTypeHeights, PermitTypeConfinedSpace, PermitTypeHotWork, PermitTypeElectricalLOTO:
		return true
	default:
		return false
	}
}

// PermitStatus defines the operational lifecycle states of a work permit.
type PermitStatus string

const (
	PermitStatusDraft           PermitStatus = "draft"
	PermitStatusPendingApproval PermitStatus = "pending_approval"
	PermitStatusApprovedActive  PermitStatus = "approved_active"
	PermitStatusSuspended       PermitStatus = "suspended"
	PermitStatusClosed          PermitStatus = "closed"
	PermitStatusCancelled       PermitStatus = "cancelled"
)

// IsValid validates whether the permit status is a known enum.
func (ps PermitStatus) IsValid() bool {
	switch ps {
	case PermitStatusDraft, PermitStatusPendingApproval, PermitStatusApprovedActive,
		PermitStatusSuspended, PermitStatusClosed, PermitStatusCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if the status cannot transition to another state.
func (ps PermitStatus) IsTerminal() bool {
	return ps == PermitStatusClosed || ps == PermitStatusCancelled
}

// PermitSignatureRole defines the digital sign-off role.
type PermitSignatureRole string

const (
	PermitSignerRequester            PermitSignatureRole = "requester"
	PermitSignerContractorSupervisor PermitSignatureRole = "contractor_supervisor"
	PermitSignerEHSAuthorizer        PermitSignatureRole = "ehs_authorizer"
)

// WorkPermit represents a High-Risk Work Permit (PTAR Digital).
type WorkPermit struct {
	ID                uuid.UUID             `json:"id"`
	TenantID          uuid.UUID             `json:"tenant_id"`
	PlantID           uuid.UUID             `json:"plant_id"`
	PlantName         string                `json:"plant_name,omitempty"`
	PermitNumber      string                `json:"permit_number"`
	PermitType        PermitType            `json:"permit_type"`
	Status            PermitStatus          `json:"status"`
	ContractorID      uuid.UUID             `json:"contractor_id"`
	ContractorName    string                `json:"contractor_name,omitempty"`
	WorkArea          string                `json:"work_area"`
	Description       string                `json:"description"`
	HazardDescription *string               `json:"hazard_description,omitempty"`
	StartTime         time.Time             `json:"start_time"`
	EndTime           time.Time             `json:"end_time"`
	CreatedByUserID   uuid.UUID             `json:"created_by_user_id"`
	CreatedByName     string                `json:"created_by_name,omitempty"`
	ApprovedByUserID  *uuid.UUID            `json:"approved_by_user_id,omitempty"`
	ApprovedByName    string                `json:"approved_by_name,omitempty"`
	ApprovedAt        *time.Time            `json:"approved_at,omitempty"`
	SuspendedAt       *time.Time            `json:"suspended_at,omitempty"`
	ClosedAt          *time.Time            `json:"closed_at,omitempty"`
	CancelledAt       *time.Time            `json:"cancelled_at,omitempty"`
	StatusReason      *string               `json:"status_reason,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
	DeletedAt         *time.Time            `json:"deleted_at,omitempty"`
	Workers           []PermitWorker        `json:"workers,omitempty"`
	Checklist         []PermitChecklistItem `json:"checklist,omitempty"`
	Signatures        []PermitSignature     `json:"signatures,omitempty"`
}

// PermitWorker is a crew member assigned to a high-risk work permit.
type PermitWorker struct {
	ID                 uuid.UUID                `json:"id"`
	TenantID           uuid.UUID                `json:"tenant_id"`
	PermitID           uuid.UUID                `json:"permit_id"`
	WorkerID           uuid.UUID                `json:"worker_id"`
	ComplianceStatus   string                   `json:"compliance_status"` // 'compliant', 'non_compliant', 'warning'
	ComplianceSnapshot WorkerComplianceSnapshot `json:"compliance_snapshot"`
	AssignedBy         *uuid.UUID               `json:"assigned_by,omitempty"`
	AssignedAt         time.Time                `json:"assigned_at"`
	CreatedAt          time.Time                `json:"created_at"`

	// Enriched worker details
	FullName    string  `json:"full_name,omitempty"`
	JobTitle    *string `json:"job_title,omitempty"`
	CURP        *string `json:"curp,omitempty"`
	IMSSNumber  *string `json:"imss_number,omitempty"`
	AvatarR2Key *string `json:"avatar_r2_key,omitempty"`
}

// WorkerComplianceSnapshot records the point-in-time compliance evaluation for a worker.
type WorkerComplianceSnapshot struct {
	IsCompliant      bool       `json:"is_compliant"`
	BlockReason      *string    `json:"block_reason,omitempty"`
	DC3Status        string     `json:"dc3_status"`
	DC3Code          string     `json:"dc3_code"`
	DC3ExpiresAt     *time.Time `json:"dc3_expires_at,omitempty"`
	MedicalStatus    string     `json:"medical_status"`
	MedicalExpiresAt *time.Time `json:"medical_expires_at,omitempty"`
	ContractorStatus string     `json:"contractor_status"`
	EvaluatedAt      time.Time  `json:"evaluated_at"`
}

// PermitChecklistItem is a single STPS safety verification item on a permit.
type PermitChecklistItem struct {
	ID             uuid.UUID      `json:"id"`
	TenantID       uuid.UUID      `json:"tenant_id"`
	PermitID       uuid.UUID      `json:"permit_id"`
	ItemKey        string         `json:"item_key"`
	NOMStandard    string         `json:"nom_standard"`
	Title          string         `json:"title"`
	Description    *string        `json:"description,omitempty"`
	Category       string         `json:"category"` // 'safety', 'ppe', 'equipment', 'atmospheric', 'emergency'
	IsMandatory    bool           `json:"is_mandatory"`
	IsVerified     bool           `json:"is_verified"`
	VerifiedBy     *uuid.UUID     `json:"verified_by,omitempty"`
	VerifiedByName string         `json:"verified_by_name,omitempty"`
	VerifiedAt     *time.Time     `json:"verified_at,omitempty"`
	Measurements   map[string]any `json:"measurements,omitempty"`
	Notes          *string        `json:"notes,omitempty"`
	DisplayOrder   int            `json:"display_order"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// PermitSignature represents an immutable digital signature record.
type PermitSignature struct {
	ID             uuid.UUID           `json:"id"`
	TenantID       uuid.UUID           `json:"tenant_id"`
	PermitID       uuid.UUID           `json:"permit_id"`
	UserID         uuid.UUID           `json:"user_id"`
	Role           PermitSignatureRole `json:"role"`
	SignatureData  *string             `json:"signature_data,omitempty"`
	IPAddress      *string             `json:"ip_address,omitempty"`
	UserAgent      *string             `json:"user_agent,omitempty"`
	Notes          *string             `json:"notes,omitempty"`
	SignedAt       time.Time           `json:"signed_at"`
	CreatedAt      time.Time           `json:"created_at"`
	SignerFullName string              `json:"signer_full_name,omitempty"`
}

// DefaultChecklistForType returns the mandatory regulatory checklist items per Mexican STPS NOM.
func DefaultChecklistForType(permitType PermitType) []PermitChecklistItem {
	strPtr := func(s string) *string { return &s }

	switch permitType {
	case PermitTypeHeights:
		return []PermitChecklistItem{
			{
				ItemKey:      "chk_heights_harness",
				NOMStandard:  "NOM-009-STPS-2011",
				Title:        "Inspección de arnés de cuerpo completo y línea de vida",
				Description:  strPtr("Inspección visual de costuras, herrajes, cinta de absorción de impacto y punto de anclaje certificado (>2,268 kg / 5,000 lbs)."),
				Category:     "ppe",
				IsMandatory:  true,
				DisplayOrder: 1,
			},
			{
				ItemKey:      "chk_heights_scaffold_ladder",
				NOMStandard:  "NOM-009-STPS-2011",
				Title:        "Estabilidad de andamios y plataformas de elevación",
				Description:  strPtr("Revisión de nivelación, bases sólidas, barandales de seguridad, rodapiés y tarjeta de inspección de andamio vigente."),
				Category:     "equipment",
				IsMandatory:  true,
				DisplayOrder: 2,
			},
			{
				ItemKey:      "chk_heights_barricade",
				NOMStandard:  "NOM-009-STPS-2011",
				Title:        "Delimitación y señalización del área inferior",
				Description:  strPtr("Acordonamiento con cinta o barreras rígidas en el radio de posible caída de herramientas u objetos."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 3,
			},
			{
				ItemKey:      "chk_heights_weather",
				NOMStandard:  "NOM-009-STPS-2011",
				Title:        "Condiciones climáticas seguras",
				Description:  strPtr("Ausencia de lluvia, tormentas eléctricas o vientos superiores a 40 km/h que pongan en riesgo al personal."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 4,
			},
			{
				ItemKey:      "chk_heights_epp",
				NOMStandard:  "NOM-009-STPS-2011",
				Title:        "EPP básico y complementario completo",
				Description:  strPtr("Casco con barbiquejo de 3 puntos, calzado de seguridad dieléctrico, guantes antiderrapantes y lentes de seguridad."),
				Category:     "ppe",
				IsMandatory:  true,
				DisplayOrder: 5,
			},
		}

	case PermitTypeConfinedSpace:
		return []PermitChecklistItem{
			{
				ItemKey:      "chk_confined_atm_test",
				NOMStandard:  "NOM-033-STPS-2015",
				Title:        "Monitoreo y prueba atmosférica continua",
				Description:  strPtr("Medición de O2 (19.5%-23.5%), LEL (<10%), H2S (<10 ppm), CO (<25 ppm) con equipo multigás calibrado."),
				Category:     "atmospheric",
				IsMandatory:  true,
				DisplayOrder: 1,
			},
			{
				ItemKey:      "chk_confined_ventilation",
				NOMStandard:  "NOM-033-STPS-2015",
				Title:        "Ventilación forzada mecánica",
				Description:  strPtr("Inyección o extracción de aire continuo antes y durante el tiempo de permanencia en el espacio."),
				Category:     "equipment",
				IsMandatory:  true,
				DisplayOrder: 2,
			},
			{
				ItemKey:      "chk_confined_vigia",
				NOMStandard:  "NOM-033-STPS-2015",
				Title:        "Vigía exterior permanente",
				Description:  strPtr("Designación y presencia obligatoria de vigía capacitado con comunicación bidireccional continua hacia el interior."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 3,
			},
			{
				ItemKey:      "chk_confined_rescue",
				NOMStandard:  "NOM-033-STPS-2015",
				Title:        "Equipo de rescate y extracción listo",
				Description:  strPtr("Trípode o pescante de rescate, malacate retráctil y arnés de rescate colocados para extracción inmediata sin ingreso."),
				Category:     "emergency",
				IsMandatory:  true,
				DisplayOrder: 4,
			},
			{
				ItemKey:      "chk_confined_isolation",
				NOMStandard:  "NOM-033-STPS-2015",
				Title:        "Aislamiento y cegado de líneas",
				Description:  strPtr("Cierre de válvulas, bridas ciegas colocadas y despresurización de tuberías conectadas al espacio."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 5,
			},
		}

	case PermitTypeHotWork:
		return []PermitChecklistItem{
			{
				ItemKey:      "chk_hot_extinguisher",
				NOMStandard:  "NOM-027-STPS-2008",
				Title:        "Extintor inspeccionado en el punto de trabajo",
				Description:  strPtr("Extintor de PQS o CO2 de al menos 9 kg verificado con manómetro en verde a no más de 5 metros."),
				Category:     "emergency",
				IsMandatory:  true,
				DisplayOrder: 1,
			},
			{
				ItemKey:      "chk_hot_combustibles",
				NOMStandard:  "NOM-027-STPS-2008",
				Title:        "Despeje de materiales combustibles a 11 metros",
				Description:  strPtr("Retiro de solventes, cartón, madera o protección con mantas ignífugas si no pueden moverse."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 2,
			},
			{
				ItemKey:      "chk_hot_screens",
				NOMStandard:  "NOM-027-STPS-2008",
				Title:        "Mamparas de protección contra chispas y radiación",
				Description:  strPtr("Instalación de cortinas o mamparas retardantes al fuego para proteger áreas adyacentes y personal cercano."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 3,
			},
			{
				ItemKey:      "chk_hot_grounding",
				NOMStandard:  "NOM-027-STPS-2008",
				Title:        "Conexión a tierra y estado de cables/mangueras",
				Description:  strPtr("Pinza de tierra firme a la pieza o estructura, cables sin empalmes pelados, válvulas arrestallamas en cilindros."),
				Category:     "equipment",
				IsMandatory:  true,
				DisplayOrder: 4,
			},
			{
				ItemKey:      "chk_hot_fire_watch",
				NOMStandard:  "NOM-027-STPS-2008",
				Title:        "Guardia contra incendios (Fire Watch)",
				Description:  strPtr("Vigilancia durante el trabajo y permanencia obligatoria durante al menos 30 minutos posteriores al concluir."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 5,
			},
		}

	case PermitTypeElectricalLOTO:
		return []PermitChecklistItem{
			{
				ItemKey:      "chk_loto_zero_energy",
				NOMStandard:  "NOM-029-STPS-2011",
				Title:        "Verificación de ausencia de tensión (energía cero)",
				Description:  strPtr("Prueba con multímetro calibrado o detector de tensión de rango adecuado en todas las fases antes de intervenir."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 1,
			},
			{
				ItemKey:      "chk_loto_locks_tags",
				NOMStandard:  "NOM-029-STPS-2011",
				Title:        "Colocación de candados y etiquetas LOTO personales",
				Description:  strPtr("Cada trabajador involucrado coloca su candado y tarjeta personal de peligro en el dispositivo de bloqueo."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 2,
			},
			{
				ItemKey:      "chk_loto_grounding",
				NOMStandard:  "NOM-029-STPS-2011",
				Title:        "Puesta a tierra y en cortocircuito temporal",
				Description:  strPtr("Instalación de equipo de puesta a tierra provisional y delimitación física del área de tableros o celdas."),
				Category:     "equipment",
				IsMandatory:  true,
				DisplayOrder: 3,
			},
			{
				ItemKey:      "chk_loto_ppe",
				NOMStandard:  "NOM-029-STPS-2011",
				Title:        "EPP dieléctrico y contra arco eléctrico verificado",
				Description:  strPtr("Guantes dieléctricos con prueba de inflado, careta para arco con balaclava, calzado dieléctrico y ropa de algodón."),
				Category:     "ppe",
				IsMandatory:  true,
				DisplayOrder: 4,
			},
			{
				ItemKey:      "chk_loto_stored_energy",
				NOMStandard:  "NOM-029-STPS-2011",
				Title:        "Descarga de energía residual almacenada",
				Description:  strPtr("Descarga de capacitores, purga de presión hidráulica/neumática y bloqueo mecánico de partes móviles o resortes."),
				Category:     "safety",
				IsMandatory:  true,
				DisplayOrder: 5,
			},
		}

	default:
		return nil
	}
}
