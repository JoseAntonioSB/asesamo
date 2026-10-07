# Milestone M3 Technical Analysis: Compliance Evaluation Engine & JSONB Snapshots

**Agent**: `teamwork_preview_explorer_m3_1`  
**Date**: October 2, 2026  
**Target Service**: `asesamo-api` (Go REST API backend)  
**Database**: PostgreSQL 16 (`db/migrations/000001_init.up.sql`)  
**Scope**: Check-In Verification Engine, Immutable `access_logs` audit trail, and JSONB `document_snapshot` specifications.

---

## 1. Executive Summary

Milestone M3 establishes the core industrial compliance decision engine of ASESAMO. When a security guard scans a worker's credential or enters their UUID at a plant gate (`POST /api/v1/access/check-in`), the system must perform multi-level compliance verification in real-time (< 500 ms) and write an immutable audit log to `access_logs` containing a cryptographic/frozen `document_snapshot` in JSONB.

Our investigation of the existing codebase (`asesamo-api/internal/core/usecase/access/check_in.go`, `asesamo-api/internal/adapter/postgres/compliance_repository.go`, `db/migrations/000001_init.up.sql`, and `db/seeds/000001_seed_dev_data.sql`) reveals that while the foundation is functioning and existing unit tests pass, there are critical compliance edge cases and security vulnerabilities in the current implementation:
1. **Worker Inactivity Blindspot**: The current repository checks `users.is_active`, but completely ignores `contractor_workers.is_active`. If a worker is dismissed or deactivated by their contractor company, they could still enter.
2. **Contractor Inactivity Blindspot**: The current repository does not query `contractors.is_active`, and only checks if `plant_contractors.status` is `'blocked'` or `'suspended'`. If a contractor is in `'pending'` status (not yet approved for that plant) or not affiliated at all, the current logic does not block them.
3. **Flawed `block_on_expired_docs` Precedence**: In the existing use case, if `plant.BlockOnExpiredDocs == false`, **all** fatal failures (including missing documents, rejected documents, inactive worker, or blocked contractor) are demoted to `granted_with_warning`. Non-negotiable rejections must never be overridden by `block_on_expired_docs`; that flag must strictly apply only to expired documents.
4. **Audit Trail Completeness**: The database schema defines `document_snapshot` as a JSONB column (`NOT NULL DEFAULT '[]'`). We formulate both the canonical array schema expected by the API and E2E tests (`[]DocumentSnapshotItem`) and the self-contained audit envelope capturing worker and contractor metadata at the exact moment of entry.

---

## 2. Check-In Verification Engine Deep Dive

### 2.1 Worker Active Status Evaluation
A worker in ASESAMO exists at the intersection of two database tables:
- `users`: Represents the authentication identity.
  - Column: `users.is_active BOOLEAN NOT NULL DEFAULT TRUE`.
  - Check: If `users.is_active == false`, the user account is disabled across the platform.
- `contractor_workers`: Represents the employment and labor affiliation between a contractor company and a worker.
  - Column: `contractor_workers.is_active BOOLEAN NOT NULL DEFAULT TRUE`.
  - Check: If `contractor_workers.is_active == false`, the worker is no longer authorized to work on behalf of that contractor.

**Evaluation Rule**:
```
worker_is_compliant = (users.is_active == TRUE) AND (contractor_workers.is_active == TRUE)
```
If either is `false`, access must be **denied** immediately with an explicit Spanish message:
- If `!users.is_active`: `"El usuario del trabajador se encuentra inactivo."`
- If `!contractor_workers.is_active`: `"El trabajador se encuentra inactivo en la relación con la empresa contratista."`

### 2.2 Contractor Active Status Evaluation
Access control at an industrial facility is predicated on the legal authorization of the contractor company:
- `contractors`: The global contractor organization.
  - Column: `contractors.is_active BOOLEAN NOT NULL DEFAULT TRUE`.
  - Check: If `contractors.is_active == false`, the contractor is globally disabled.
- `plant_contractors`: The specific authorization of a contractor company within a specific plant facility.
  - Column: `plant_contractors.status plant_contractor_status NOT NULL DEFAULT 'pending'`.
  - Allowed enum values: `'pending'`, `'active'`, `'suspended'`, `'blocked'`.

**Evaluation Rule**:
Access is **only** permitted if:
```
contractor_is_authorized = (contractors.is_active == TRUE) AND (plant_contractors.status == 'active')
```
If `plant_contractors.status != 'active'`, access must be **denied** with explicit Spanish messages:
- If `!contractors.is_active`: `"La empresa contratista '%s' se encuentra inactiva en el sistema."`
- If `status == 'blocked'`: `"La empresa contratista '%s' se encuentra bloqueada en esta planta."`
- If `status == 'suspended'`: `"La empresa contratista '%s' se encuentra suspendida en esta planta."`
- If `status == 'pending'` or record missing: `"La empresa contratista '%s' no cuenta con autorización activa en esta planta (estatus: pendiente)."`

### 2.3 Plant Policy Evaluation: `plants.block_on_expired_docs`
Each plant in `plants` defines its own document enforcement policy via:
- Column: `plants.block_on_expired_docs BOOLEAN NOT NULL DEFAULT TRUE`.

**Behavioral Semantics**:
- `block_on_expired_docs == true` (Standard industrial enforcement):
  If any mandatory document is expired (`expires_at <= now()`), access result is `denied`.
- `block_on_expired_docs == false` (Lenient / emergency plant mode):
  If a mandatory document is expired, but is otherwise approved and present, access is **not** denied; instead, the result is `granted_with_warning`, and an informational warning is logged.

**Critical Invariant**:
`block_on_expired_docs` applies **ONLY** to document expiration (`expires_at <= now()`). It must **NEVER** bypass:
1. Missing mandatory documents.
2. Rejected documents (`status == 'rejected'`).
3. Documents pending review (`status == 'pending_review'`).
4. Inactive worker accounts (`users.is_active == false` or `contractor_workers.is_active == false`).
5. Unauthorized/blocked contractors (`plant_contractors.status != 'active'`).

### 2.4 Plant Document Requirements Catalog
Requirements are evaluated from `plant_document_requirements`:
- Columns: `plant_id`, `tenant_id`, `document_type_id`, `is_required`, `validity_days`.
- Each requirement joins to `document_types` to obtain:
  - `code`: e.g. `'DC3'`, `'REPSE'`, `'SUA'`, `'INE'`, `'MEDICAL'`, `'SAT_32D'`.
  - `name`: e.g. `"Constancia de Capacitación DC-3"`.
  - `applies_to`: `'worker'` or `'contractor'`.
  - `default_validity_days`: Fallback validity days if `validity_days` is NULL.

**System Fallback**:
If a plant has zero entries in `plant_document_requirements`, the engine falls back to global system types in `document_types` where `is_system = TRUE` and `is_required_default = TRUE`.

**Seed Baseline (Planta Querétaro Norte)**:
- Mandatory Worker Documents (`applies_to = 'worker'`, `is_required = TRUE`):
  1. `DC3`: Constancia de Capacitación DC-3.
  2. `INE`: Identificación Oficial (INE/IFE) (validity: indefinite).
  3. `MEDICAL`: Certificado Médico de Aptitud.
- Mandatory Contractor Documents (`applies_to = 'contractor'`, `is_required = TRUE`):
  1. `REPSE`: Registro REPSE (Secretaría del Trabajo y Previsión Social).
  2. `SUA`: Comprobante de Pago SUA (IMSS / INFONAVIT).
- Supplementary / Optional Documents (`is_required = FALSE`):
  1. `SAT_32D`: Opinión de Cumplimiento SAT (32-D). Evaluated and recorded in snapshot, but never triggers denial.

### 2.5 Document Compliance & Expiration Evaluation Algorithm
For each requirement rule `req` in `plant_document_requirements`:
1. **Document Lookup**:
   - If `req.AppliesTo == 'worker'`: Query `documents` where `user_id = worker_user_id AND document_type_id = req.DocumentTypeID AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`.
   - If `req.AppliesTo == 'contractor'`: Query `documents` where `contractor_id = contractor_id AND document_type_id = req.DocumentTypeID AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1`.
2. **Missing Document Check**:
   - If no document record is found:
     - `item.Status = "missing"`
     - `item.IsValid = false`
     - `item.FailureReason = "Documento '%s' faltante."`
     - If `req.IsRequired == true` -> Register `nonNegotiableDenial`.
3. **Approval Status Check**:
   - If document exists:
     - `item.Status = string(doc.Status)`
     - If `doc.Status != 'approved'`:
       - `item.IsValid = false`
       - If `doc.Status == 'rejected'`:
         `item.FailureReason = "Documento '%s' rechazado%s."` (appending rejection reason if available).
       - If `doc.Status == 'pending_review'`:
         `item.FailureReason = "Documento '%s' pendiente de revisión."`
       - If `doc.Status == 'expired'`:
         `item.FailureReason = "Documento '%s' vencido el YYYY-MM-DD."`
       - If `req.IsRequired == true` -> Register `nonNegotiableDenial`.
4. **Expiration Date Check**:
   - If `doc.Status == 'approved'`:
     - If `doc.ExpiresAt == nil`:
       - Valid indefinitely (e.g. INE).
       - `item.IsValid = true`, `item.FailureReason = nil`.
     - If `doc.ExpiresAt != nil`:
       - Format date string: `YYYY-MM-DD`.
       - If `doc.ExpiresAt.Before(now)`:
         - `item.IsValid = false`
         - `item.FailureReason = "Documento '%s' vencido el %s."`
         - If `req.IsRequired == true`:
           - If `plant.BlockOnExpiredDocs == true` -> Register `expiredDocDenial`.
           - If `plant.BlockOnExpiredDocs == false` -> Register `expiredDocWarning`.
       - If `doc.ExpiresAt.After(now)`:
         - `item.IsValid = true`, `item.FailureReason = nil`.
         - Warning Window Check:
           If `doc.ExpiresAt.Before(warningThreshold)` (where `warningThreshold = now + 15 to 30 days`):
           - Register `expiringSoonWarning`: `"Documento '%s' próximo a vencer el %s."`

### 2.6 Decision Matrix

| Condition | `nonNegotiableDenials` | `expiredDocDenials` | `expiredDocWarnings` | `expiringSoonWarnings` | Result | Denial Reason |
|:---|:---:|:---:|:---:|:---:|:---:|:---|
| All docs valid, no warnings | 0 | 0 | 0 | 0 | `granted` | `null` |
| All docs valid, 1+ expiring soon | 0 | 0 | 0 | > 0 | `granted_with_warning` | `null` (or warning summary) |
| Doc expired, `block_on_expired_docs = false` | 0 | 0 | > 0 | ≥ 0 | `granted_with_warning` | Summary of expired doc warnings |
| Doc expired, `block_on_expired_docs = true` | 0 | > 0 | 0 | ≥ 0 | `denied` | Detail of expired documents |
| Missing mandatory doc | > 0 | any | any | any | `denied` | Detail of missing documents |
| Rejected mandatory doc | > 0 | any | any | any | `denied` | Detail of rejected documents |
| Inactive worker / unauthorized contractor | > 0 | any | any | any | `denied` | Worker/contractor inactivity reason |

---

## 3. Immutable Access Logs & JSONB Snapshot Design

### 3.1 Database Table Schema (`access_logs`)
From `db/migrations/000001_init.up.sql`:

```sql
CREATE TABLE access_logs (
  id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         UUID          NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  plant_id          UUID          NOT NULL REFERENCES plants(id) ON DELETE RESTRICT,
  access_point_id   UUID          NOT NULL REFERENCES access_points(id) ON DELETE RESTRICT,
  user_id           UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  contractor_id     UUID          REFERENCES contractors(id) ON DELETE RESTRICT,
  processed_by      UUID          REFERENCES users(id) ON DELETE SET NULL,
  checked_in_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
  checked_out_at    TIMESTAMPTZ,
  result            access_result NOT NULL,
  denial_reason     TEXT,
  document_snapshot JSONB         NOT NULL DEFAULT '[]',
  notes             TEXT,
  created_at        TIMESTAMPTZ   NOT NULL DEFAULT now()
);
```

**Immutability Characteristics**:
- No `updated_at` column: The row cannot be edited once inserted.
- The single mutable field post-creation is `checked_out_at`, which is set only once upon exit.
- `document_snapshot` is stored in PostgreSQL as `JSONB`, freezing the exact compliance status of all documents at the timestamp of `checked_in_at`. Even if a document is later modified, expired, or deleted, the snapshot remains historically accurate for regulatory audits.

### 3.2 Canonical JSONB Schema for `document_snapshot` (Array of Items)
In accordance with `PROJECT.md` line 153, `000001_init.up.sql` line 388, and the E2E test suites (`tier1_feature_test.sh` line 189), the `document_snapshot` stored in the database and returned under `.data.document_snapshot` is a JSON array.

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "DocumentSnapshot",
  "description": "Array of evaluated document compliance records at check-in time",
  "type": "array",
  "items": {
    "type": "object",
    "required": [
      "document_type_code",
      "document_name",
      "applies_to",
      "status",
      "is_required",
      "is_valid"
    ],
    "properties": {
      "document_type_code": {
        "type": "string",
        "example": "DC3"
      },
      "document_name": {
        "type": "string",
        "example": "Constancia de Capacitación DC-3"
      },
      "applies_to": {
        "type": "string",
        "enum": ["worker", "contractor"]
      },
      "status": {
        "type": "string",
        "enum": ["approved", "pending_review", "rejected", "expired", "missing"]
      },
      "is_required": {
        "type": "boolean"
      },
      "is_valid": {
        "type": "boolean"
      },
      "expires_at": {
        "type": ["string", "null"],
        "format": "date",
        "example": "2027-08-01"
      },
      "failure_reason": {
        "type": ["string", "null"],
        "example": "Documento 'Constancia de Capacitación DC-3' vencido el 2026-08-25."
      }
    },
    "additionalProperties": false
  }
}
```

### 3.3 Complete Check-In API Response Envelope
The HTTP response for `POST /api/v1/access/check-in` provides the guard with an immediate visual and structural breakdown:

```json
{
  "data": {
    "access_log_id": "99999999-9999-9999-9999-999999999991",
    "result": "granted",
    "denial_reason": null,
    "checked_in_at": "2026-10-02T06:00:00Z",
    "worker": {
      "user_id": "66666666-6666-6666-6666-666666666661",
      "full_name": "Carlos Mendoza",
      "curp": "MENC850101HDFR01",
      "imss_number": "12345678901",
      "job_title": "Soldador Industrial",
      "avatar_r2_key": null
    },
    "contractor": {
      "id": "55555555-5555-5555-5555-555555555555",
      "name": "Mantenimiento Industrial del Norte SA de CV",
      "rfc": "MIN120304AB1",
      "repse_number": "REPSE-NL-2023-8841",
      "status": "active"
    },
    "plant": {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "Planta Querétaro Norte"
    },
    "access_point": {
      "id": "33333333-3333-3333-3333-333333333331",
      "name": "Caseta Principal Norte"
    },
    "document_snapshot": [
      {
        "document_type_code": "DC3",
        "document_name": "Constancia de Capacitación DC-3",
        "applies_to": "worker",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2027-08-01",
        "failure_reason": null
      },
      {
        "document_type_code": "INE",
        "document_name": "Identificación Oficial (INE/IFE)",
        "applies_to": "worker",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": null,
        "failure_reason": null
      },
      {
        "document_type_code": "MEDICAL",
        "document_name": "Certificado Médico de Aptitud",
        "applies_to": "worker",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2027-03-31",
        "failure_reason": null
      },
      {
        "document_type_code": "REPSE",
        "document_name": "Registro REPSE",
        "applies_to": "contractor",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2027-08-03",
        "failure_reason": null
      },
      {
        "document_type_code": "SUA",
        "document_name": "Comprobante de Pago SUA (IMSS/INFONAVIT)",
        "applies_to": "contractor",
        "status": "approved",
        "is_required": true,
        "is_valid": true,
        "expires_at": "2026-10-27",
        "failure_reason": null
      }
    ]
  }
}
```

### 3.4 Extended Standalone Audit Envelope Schema
For long-term compliance archives, legal disputes, or STPS audits, the entire context of the check-in (including worker CURP/NSS, contractor RFC/REPSE, and plant policy at that precise moment) can be encapsulated in an extended audit envelope:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ComprehensiveAuditSnapshot",
  "type": "object",
  "required": [
    "evaluated_at",
    "decision",
    "worker_snapshot",
    "contractor_snapshot",
    "plant_policy_snapshot",
    "documents"
  ],
  "properties": {
    "evaluated_at": { "type": "string", "format": "date-time" },
    "decision": {
      "type": "object",
      "required": ["result"],
      "properties": {
        "result": { "type": "string", "enum": ["granted", "granted_with_warning", "denied"] },
        "denial_reason": { "type": ["string", "null"] }
      }
    },
    "worker_snapshot": {
      "type": "object",
      "required": ["user_id", "full_name", "user_active", "contractor_worker_active"],
      "properties": {
        "user_id": { "type": "string", "format": "uuid" },
        "full_name": { "type": "string" },
        "curp": { "type": ["string", "null"] },
        "imss_number": { "type": ["string", "null"] },
        "job_title": { "type": ["string", "null"] },
        "user_active": { "type": "boolean" },
        "contractor_worker_active": { "type": "boolean" }
      }
    },
    "contractor_snapshot": {
      "type": "object",
      "required": ["id", "name", "contractor_active", "plant_status"],
      "properties": {
        "id": { "type": "string", "format": "uuid" },
        "name": { "type": "string" },
        "rfc": { "type": ["string", "null"] },
        "repse_number": { "type": ["string", "null"] },
        "contractor_active": { "type": "boolean" },
        "plant_status": { "type": "string", "enum": ["active", "pending", "suspended", "blocked"] }
      }
    },
    "plant_policy_snapshot": {
      "type": "object",
      "required": ["plant_id", "block_on_expired_docs"],
      "properties": {
        "plant_id": { "type": "string", "format": "uuid" },
        "plant_name": { "type": "string" },
        "block_on_expired_docs": { "type": "boolean" }
      }
    },
    "documents": {
      "type": "array",
      "items": { "$ref": "#/definitions/DocumentSnapshotItem" }
    }
  }
}
```

---

## 4. Gap Analysis & Proposed Code Enhancements

### 4.1 Gap Comparison Matrix

| Component | Current Implementation | Milestone M3 Required Behavior | Severity |
|:---|:---|:---|:---:|
| **Worker Inactive** | Only checks `users.is_active`. Omits `contractor_workers.is_active`. | Must check BOTH `users.is_active` AND `contractor_workers.is_active`. | **High** |
| **Contractor Inactive** | Omits `contractors.is_active`. Only checks `plant_contractors.status == 'blocked' \|\| 'suspended'`. | Must check `contractors.is_active` AND require `plant_contractors.status == 'active'`. 'pending' must be blocked. | **High** |
| **Policy Precedence** | If `plant.BlockOnExpiredDocs == false`, all fatal failures (including missing/rejected docs and inactive users) turn into `granted_with_warning`. | `BlockOnExpiredDocs` applies ONLY to expired documents. Missing/rejected docs and inactive accounts ALWAYS cause `denied`. | **Critical** |
| **Warning Threshold** | Hardcoded to 30 days (`now.AddDate(0, 0, 30)`). | Configurable threshold (e.g. 15 to 30 days). | **Low** |
| **SQL Affiliation Query** | Missing `cw.is_active` and `c.is_active` columns in query and scan. | Query and scan `cw.is_active` and `c.is_active` in `compliance_repository.go`. | **Medium** |

### 4.2 Proposed Domain Entity Struct Updates

#### Update to `asesamo-api/internal/core/domain/entity/compliance.go`:
```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// WorkerAffiliation holds worker information and affiliated contractor details.
type WorkerAffiliation struct {
	WorkerID                 uuid.UUID `json:"worker_id"`
	FullName                 string    `json:"full_name"`
	CURP                     *string   `json:"curp,omitempty"`
	IMSSNumber               *string   `json:"imss_number,omitempty"`
	JobTitle                 *string   `json:"job_title,omitempty"`
	AvatarR2Key              *string   `json:"avatar_r2_key,omitempty"`
	UserIsActive             bool      `json:"user_is_active"`
	ContractorWorkerIsActive bool      `json:"contractor_worker_is_active"`
	IsActive                 bool      `json:"is_active"` // composite: UserIsActive && ContractorWorkerIsActive
	ContractorID             uuid.UUID `json:"contractor_id"`
	ContractorName           string    `json:"contractor_name"`
	ContractorRFC            *string   `json:"contractor_rfc,omitempty"`
	REPSENumber              *string   `json:"repse_number,omitempty"`
	ContractorIsActive       bool      `json:"contractor_is_active"`
	PlantStatus              string    `json:"plant_status"` // 'active' | 'pending' | 'suspended' | 'blocked'
}
```

### 4.3 Proposed Use Case Logic Update (`check_in.go`)

```go
	// 6. Evaluate Compliance
	now := time.Now().UTC()
	warningThreshold := now.AddDate(0, 0, 30) // or plant-configured threshold (e.g. 15-30 days)

	var (
		snapshot               entity.DocumentSnapshot
		nonNegotiableDenials   []string
		expiredDocDenials      []string
		expiredDocWarnings     []string
		expiringSoonWarnings   []string
	)

	// A. Worker Active Verification
	if !aff.UserIsActive {
		nonNegotiableDenials = append(nonNegotiableDenials, "El usuario del trabajador se encuentra inactivo.")
	}
	if !aff.ContractorWorkerIsActive {
		nonNegotiableDenials = append(nonNegotiableDenials, "El trabajador no se encuentra activo con la empresa contratista.")
	}

	// B. Contractor Active Verification
	if aff.ContractorID != uuid.Nil {
		if !aff.ContractorIsActive {
			nonNegotiableDenials = append(nonNegotiableDenials, fmt.Sprintf("La empresa contratista '%s' se encuentra inactiva en el sistema.", aff.ContractorName))
		}
		if aff.PlantStatus != "active" {
			switch aff.PlantStatus {
			case "blocked":
				nonNegotiableDenials = append(nonNegotiableDenials, fmt.Sprintf("La empresa contratista '%s' se encuentra bloqueada en esta planta.", aff.ContractorName))
			case "suspended":
				nonNegotiableDenials = append(nonNegotiableDenials, fmt.Sprintf("La empresa contratista '%s' se encuentra suspendida en esta planta.", aff.ContractorName))
			case "pending":
				nonNegotiableDenials = append(nonNegotiableDenials, fmt.Sprintf("La empresa contratista '%s' no cuenta con autorización activa en esta planta (estatus: pendiente).", aff.ContractorName))
			default:
				nonNegotiableDenials = append(nonNegotiableDenials, fmt.Sprintf("La empresa contratista '%s' no está autorizada en esta planta.", aff.ContractorName))
			}
		}
	}

	// C. Document Requirements Verification
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
				nonNegotiableDenials = append(nonNegotiableDenials, reason)
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
					nonNegotiableDenials = append(nonNegotiableDenials, reason)
				}
			} else if doc.ExpiresAt != nil && doc.ExpiresAt.Before(now) {
				item.IsValid = false
				reason := fmt.Sprintf("Documento '%s' vencido el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02"))
				item.FailureReason = &reason
				if req.IsRequired {
					if plant.BlockOnExpiredDocs {
						expiredDocDenials = append(expiredDocDenials, reason)
					} else {
						expiredDocWarnings = append(expiredDocWarnings, reason)
					}
				}
			} else {
				item.IsValid = true
				if doc.ExpiresAt != nil && doc.ExpiresAt.Before(warningThreshold) {
					expiringSoonWarnings = append(expiringSoonWarnings, fmt.Sprintf("Documento '%s' próximo a vencer el %s.", req.DocumentName, doc.ExpiresAt.Format("2006-01-02")))
				}
			}
		}

		snapshot = append(snapshot, item)
	}

	// 7. Determine Final Result
	var (
		decisionResult entity.AccessResult
		denialReason   *string
	)

	allDenials := append(nonNegotiableDenials, expiredDocDenials...)

	if len(allDenials) > 0 {
		decisionResult = entity.AccessResultDenied
		combined := strings.Join(allDenials, " | ")
		denialReason = &combined
	} else if len(expiredDocWarnings) > 0 || len(expiringSoonWarnings) > 0 {
		decisionResult = entity.AccessResultGrantedWithWarning
		if len(expiredDocWarnings) > 0 {
			combined := strings.Join(expiredDocWarnings, " | ")
			denialReason = &combined
		}
	} else {
		decisionResult = entity.AccessResultGranted
	}
```

### 4.4 Proposed PostgreSQL Repository Query (`compliance_repository.go`)
```sql
SELECT 
    c.id AS contractor_id,
    c.name AS contractor_name,
    c.rfc AS contractor_rfc,
    c.repse_number AS repse_number,
    c.is_active AS contractor_is_active,
    cw.curp,
    cw.imss_number,
    cw.job_title,
    cw.is_active AS contractor_worker_is_active,
    COALESCE(pc.status::text, 'pending') AS plant_status
FROM contractor_workers cw
JOIN contractors c ON cw.contractor_id = c.id
LEFT JOIN plant_contractors pc ON pc.contractor_id = c.id AND pc.plant_id = $1
WHERE cw.user_id = $2 AND cw.deleted_at IS NULL AND c.deleted_at IS NULL
LIMIT 1;
```

---

## 5. Verification & Test Plan

1. **Unit Testing**:
   - `internal/core/usecase/access/check_in_test.go`:
     - Test compliant worker -> `granted`.
     - Test worker with `users.is_active = false` -> `denied`.
     - Test worker with `contractor_workers.is_active = false` -> `denied`.
     - Test contractor with `contractors.is_active = false` -> `denied`.
     - Test contractor with `plant_contractors.status = 'pending'` -> `denied`.
     - Test contractor with `plant_contractors.status = 'blocked'` -> `denied`.
     - Test missing mandatory doc -> `denied`.
     - Test rejected mandatory doc -> `denied`.
     - Test expired doc with `plant.BlockOnExpiredDocs = true` -> `denied`.
     - Test expired doc with `plant.BlockOnExpiredDocs = false` -> `granted_with_warning`.
     - Test missing doc with `plant.BlockOnExpiredDocs = false` -> MUST BE `denied` (asserting separation).
     - Test doc expiring within threshold -> `granted_with_warning`.
2. **Database Integration Testing**:
   - Execute against local PostgreSQL using `db/seeds/000001_seed_dev_data.sql`:
     - Worker 1 (`Carlos Mendoza`) -> `granted`.
     - Worker 2 (`Juan Perez`) -> `denied`.
     - Worker 3 (`Roberto Garcia`) -> `granted_with_warning`.
3. **E2E Suite Verification**:
   - Run `test/e2e/tier1_feature_test.sh`: Asserts TC 4.1 to 4.7.
   - Run `test/e2e/tier2_boundary_test.sh`: Asserts TC 2.B9 to 2.B10.
   - Run `test/e2e/tier3_combination_test.sh`: Asserts audit trail snapshot persistence.
   - Run `test/e2e/tier4_workload_test.sh`: Asserts shift burst and latency <= 500ms.
