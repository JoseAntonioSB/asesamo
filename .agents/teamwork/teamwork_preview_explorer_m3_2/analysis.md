# Milestone M3: HTTP Handlers, DTOs & Route Contracts — Technical Analysis

## 1. Executive Summary & Scope

Milestone M3 establishes the gate access verification, worker check-out, and audit logging engine for ASESAMO Phase 1 MVP. This investigation analyzes the HTTP presentation layer (`internal/adapter/http/`), its interaction with the core domain/usecase layer, and strict alignment with:
- The user prompt specifications.
- `PROJECT.md` interface contracts (§ Interface Contracts).
- `DEVELOPMENT_RULES.md` architectural conventions (Clean Architecture, tenant isolation `WHERE tenant_id = $1`, standard response envelopes, Spanish user messages, English error codes).
- The opaque-box E2E test suite (`test/e2e/tier1_feature_test.sh`, `tier2_boundary_test.sh`, `tier3_combination_test.sh`, `tier4_workload_test.sh`).

---

## 2. Endpoint Analysis

### 2.1. Gate Check-In Verification (`POST /api/v1/access/check-in`)

#### Contract Comparison & Requirements

| Specification Source | Request Fields | Response Fields Verified | Status Codes |
|---|---|---|---|
| **User Prompt** | `worker_user_id` (uuid), `access_point_id` (uuid) | `access_log_id`, `result` (`granted` \| `granted_with_warning` \| `denied`), `denial_reason` (null \| string), `worker`, `contractor`, `documents` (array), `checked_in_at` | 200, 400, 401, 404 |
| **PROJECT.md** (§ 3) | `worker_user_id`, `access_point_id` | `access_log_id`, `result`, `denial_reason`, `checked_in_at`, `worker` (full profile), `contractor` (company details), `plant`, `access_point`, `document_snapshot` | 200, 400, 401, 404 |
| **E2E Tests** (`tier1`, `tier2`, `tier3`, `tier4`) | `worker_id` and/or `worker_user_id`, `access_point_id`, optional `notes` | `.data.result` (`granted`, `denied`, `granted_with_warning`), `.data.access_log_id`, `.data.denial_reason`, `.data.document_snapshot` | 200 (even on denial), 400/422 on malformed UUIDs, 401 unauth, 404 nonexistent worker/ap |

#### Discrepancies & Harmonization Strategy
1. **Worker Identifier Field Name**:
   - The user prompt defines `worker_user_id`.
   - The E2E tests send `worker_id` (in `tier1`, `tier2`, `tier3`, `tier4`) and sometimes both `worker_id` and `worker_user_id`.
   - *Resolution*: The `CheckInRequest` DTO must accept both `worker_user_id` and `worker_id` with fallback logic. If either is provided, it validates successfully.
2. **Document Breakdown Field Name**:
   - The user prompt specifies `"documents": [ ... ]`.
   - `PROJECT.md` and E2E assertions specifically inspect `(.data.document_snapshot | type == "array")`.
   - *Resolution*: In the response payload, provide `document_snapshot` as the canonical audit field, AND include `documents` as an alias referencing the exact same slice. This guarantees 100% compliance with both prompt expectations and E2E test assertions.
3. **Optional Notes**:
   - E2E tests (`tier2` TC 2.B13, `tier3` Flow 1.3, `tier4` workload) send `"notes": "..."`.
   - *Resolution*: Include `Notes *string `json:"notes,omitempty"`` in `CheckInRequest`.
4. **Access Denial HTTP Status**:
   - When a worker's documents are expired or missing and the plant policy blocks entry (`block_on_expired_docs = true`), the gate evaluation transaction succeeds normally.
   - HTTP status is **`200 OK`**, returning `.data.result = "denied"` and `.data.denial_reason = "..."`.
   - The denial is a valid business decision recorded in `access_logs`, not an HTTP protocol fault.

---

### 2.2. Gate Check-Out (`POST /api/v1/access/check-out`)

#### Contract Comparison & Requirements

| Specification Source | Request Fields | Response Fields Verified | Status Codes |
|---|---|---|---|
| **User Prompt** | `worker_user_id` (uuid), `access_point_id` (uuid) | `access_log_id`, `checked_out_at`, `duration_minutes` | 200, 400, 401, 404 |
| **E2E Tests** (`tier3` Flow 3.2) | `worker_id`, `plant_id`, `access_point_id` | `.data.checked_out_at != null`, response envelope | 200, 401, 404 |
| **Existing Codebase** | `worker_id`, `worker_user_id`, `plant_id`, `access_point_id` | `access_log_id`, `result`, `checked_in_at`, `checked_out_at` (missing `duration_minutes`!) | 200, 400, 401, 404 |

#### Critical Gaps Identified & Fixes
1. **Missing `duration_minutes`**:
   - Current `CheckOutOutput` in `internal/core/usecase/access/check_out.go` lacks `duration_minutes`.
   - *Fix*: Add `DurationMinutes int `json:"duration_minutes"`` to the response DTO and compute `duration_minutes = int(checked_out_at.Sub(checked_in_at).Minutes())` (ensuring >= 0).
2. **Handling Missing `plant_id`**:
   - The user prompt request body contains only `{ "worker_user_id": "...", "access_point_id": "..." }` without `plant_id`.
   - The current repository query in `CheckOutLatest` filters by `plant_id = $3`. If `plant_id` is omitted, `plantUUID` is `uuid.Nil`, causing the query to fail to locate active entries.
   - *Fix*: When `plant_id` is not provided in the request, resolve `plant_id` from the provided `access_point_id` via `AccessPointRepository.GetByID`, OR allow `plant_id` to be optional in `CheckOutLatest` (`WHERE tenant_id = $1 AND ($2::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR plant_id = $2)`).

---

### 2.3. Access Logs Retrieval (`GET /api/v1/access-logs`)

#### Contract Comparison & Requirements

| Specification Source | Query Parameters | Response Format | Status Codes |
|---|---|---|---|
| **User Prompt** | `plant_id`, `access_point_id`, `worker_user_id`, `result`, `from`, `to`, `page`, `per_page` | `{ "data": [ ... ], "meta": { "page": 1, "per_page": 20, "total": ... } }` | 200, 400, 401 |
| **E2E Tests** (`tier1`, `tier2`, `tier3`, `tier4`) | `plant_id`, `page`, `per_page` | `.data` array, each element has `.id`, `.result`, `.document_snapshot`, `meta` pagination | 200, 401 |
| **Existing Codebase** | Parses only `plant_id`, `page`, `per_page` | Returns `SuccessResponse` with `Data: out.Logs`, `Meta: map[string]interface{}` | 200, 400, 401 |

#### Critical Gaps Identified & Fixes
1. **Missing Query Filter Parsing**:
   - The current `ListLogs` handler in `internal/adapter/http/handler/access_handler.go` parses only `plant_id`, `page`, and `per_page`.
   - Parameters `access_point_id`, `worker_user_id`, `result`, `from`, and `to` are completely ignored!
   - *Fix*: Expand `ListAccessLogsQuery` / `ListAccessLogsInput` to support all 8 query parameters.
2. **Dynamic Repository Query**:
   - `AccessLogRepository` currently has only `ListByPlant` and `ListByTenant`.
   - *Fix*: Introduce a flexible query method `List(ctx context.Context, filter AccessLogFilter) ([]entity.AccessLog, int, error)` with dynamic `WHERE` conditions that strictly enforces `tenant_id = $1` as the first condition.
3. **Date/Time Parsing Resilience**:
   - The `from` and `to` parameters should accept ISO 8601 / RFC 3339 formats (`2006-01-02T15:04:05Z07:00` or `2006-01-02`). Invalid formats should return `400 Bad Request` with a descriptive Spanish error message.
4. **Pagination Envelope Struct**:
   - Formalize `dto.PaginationMeta` and `dto.WriteSuccessWithMeta` instead of untyped `map[string]interface{}`.

---

## 3. Router Wiring & Security Architecture

### 3.1. Chi Router Configuration (`internal/adapter/http/router.go`)

The router is mounted via `github.com/go-chi/chi/v5` with the following pipeline:

```
[Incoming HTTP Request]
       │
       ▼
1. chimiddleware.RequestID       (Traceability)
2. chimiddleware.RealIP          (Accurate client IP for logs/rate limiting)
3. chimiddleware.Logger          (Structured request logging)
4. chimiddleware.Recoverer       (Fail-safe panic recovery returning 500)
5. customMiddleware.CORS         (Configurable CORS origins)
       │
       ├──► Public Endpoints:
       │    - GET /health
       │    - GET /api/v1/health
       │    - POST /api/v1/auth/login  (Protected by httprate: 10 req/min per IP)
       │
       └──► Protected Group (/api/v1):
            - customMiddleware.Auth(jwtSecret)
              * Validates Authorization: Bearer <JWT>
              * Validates HS256 signature and expiration
              * Injects tenant_id, user_id, role into request context
              * Rejects missing/malformed/expired tokens with 401 Unauthorized
              │
              ├── GET  /access-points       (Catalog for guard caseta selection)
              ├── POST /access/check-in      (Gate compliance verification & logging)
              ├── POST /access/check-out     (Worker exit logging)
              └── GET  /access-logs          (Paginated audit logs with multi-filters)
```

### 3.2. Defensive Nil-Check on Handlers
In `router.go`, route registrations for `cfg.AccessHandler` should be conditionally guarded:
```go
if cfg.AccessHandler != nil {
    r.Post("/access/check-in", cfg.AccessHandler.CheckIn)
    r.Post("/access/check-out", cfg.AccessHandler.CheckOut)
    r.Get("/access-logs", cfg.AccessHandler.ListLogs)
}
```
This ensures unit and adversarial tests configuring partial router configs do not panic during initialization.

---

## 4. Standard Response Envelope Formatting

Per `DEVELOPMENT_RULES.md` § 6 and `PROJECT.md` § 1, all responses adhere strictly to standardized envelopes:

### 4.1. Success Envelope
```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 42
  }
}
```

### 4.2. Error Envelope
```json
{
  "error": {
    "code": "WORKER_NOT_FOUND",
    "message": "Trabajador no encontrado en el sistema."
  }
}
```

### 4.3. Status Code Mapping
| Situation | HTTP Status | Error Code | Example Message |
|---|---|---|---|
| Successful check-in / check-out / log list | 200 OK | N/A | Success envelope |
| Malformed JSON body or empty payload | 400 Bad Request | `INVALID_INPUT` | "El cuerpo de la solicitud no es un JSON válido o está vacío." |
| Malformed UUID syntax | 400 Bad Request | `INVALID_INPUT` | "El formato del UUID de trabajador no es válido." |
| Invalid query date format (`from`/`to`) | 400 Bad Request | `INVALID_INPUT` | "El formato de fecha 'from' no es válido. Use YYYY-MM-DD o RFC3339." |
| Missing or malformed Authorization header | 401 Unauthorized | `UNAUTHORIZED` | "No autorizado. Se requiere token de autenticación válido." |
| Expired JWT token | 401 Unauthorized | `TOKEN_EXPIRED` | "El token de autenticación ha vencido." |
| Tampered / invalid JWT signature | 401 Unauthorized | `INVALID_TOKEN` | "Token de autenticación inválido o corrupto." |
| Forbidden role access | 403 Forbidden | `FORBIDDEN` | "Acceso denegado. Permisos insuficientes." |
| Worker not found in database | 404 Not Found | `WORKER_NOT_FOUND` | "Trabajador no encontrado en el sistema." |
| Access point not found / cross-tenant AP | 404 Not Found | `ACCESS_POINT_NOT_FOUND` | "Punto de acceso no encontrado o no pertenece al tenant." |
| Database query or pool failure | 500 Internal Error | `DATABASE_ERROR` | "Error de comunicación con la base de datos." |
| Rate limit exceeded | 429 Too Many Req | `TOO_MANY_REQUESTS` | "Demasiadas solicitudes. Intente más tarde." |

---

## 5. Proposed DTO & Code Specifications

### 5.1. `internal/adapter/http/dto/access_dto.go`
```go
package dto

import (
	"strings"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"

	"github.com/google/uuid"
)

// CheckInRequest represents the payload for POST /api/v1/access/check-in.
type CheckInRequest struct {
	WorkerID      string  `json:"worker_id,omitempty"`
	WorkerUserID  string  `json:"worker_user_id,omitempty"`
	AccessPointID string  `json:"access_point_id"`
	Notes         *string `json:"notes,omitempty"`
}

// Validate parses and validates identifiers in CheckInRequest.
func (r *CheckInRequest) Validate() (workerUUID uuid.UUID, apUUID uuid.UUID, err error) {
	idStr := strings.TrimSpace(r.WorkerUserID)
	if idStr == "" {
		idStr = strings.TrimSpace(r.WorkerID)
	}
	if idStr == "" {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es requerido.")
	}

	workerUUID, err = uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de trabajador no es válido.")
	}

	apStr := strings.TrimSpace(r.AccessPointID)
	if apStr == "" {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El punto de acceso es requerido.")
	}

	apUUID, err = uuid.Parse(apStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de punto de acceso no es válido.")
	}

	return workerUUID, apUUID, nil
}

// CheckInResponse represents the data payload for check-in response.
type CheckInResponse struct {
	AccessLogID      uuid.UUID                     `json:"access_log_id"`
	Result           entity.AccessResult           `json:"result"`
	DenialReason     *string                       `json:"denial_reason"`
	CheckedInAt      time.Time                     `json:"checked_in_at"`
	Worker           WorkerDTO                     `json:"worker"`
	Contractor       ContractorDTO                 `json:"contractor"`
	Plant            PlantDTO                      `json:"plant"`
	AccessPoint      AccessPointDTO                `json:"access_point"`
	DocumentSnapshot []entity.DocumentSnapshotItem `json:"document_snapshot"`
	Documents        []entity.DocumentSnapshotItem `json:"documents,omitempty"` // Alias for prompt compliance
}

type WorkerDTO struct {
	UserID      uuid.UUID `json:"user_id"`
	FullName    string    `json:"full_name"`
	CURP        *string   `json:"curp,omitempty"`
	IMSSNumber  *string   `json:"imss_number,omitempty"`
	JobTitle    *string   `json:"job_title,omitempty"`
	AvatarR2Key *string   `json:"avatar_r2_key,omitempty"`
}

type ContractorDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RFC         *string   `json:"rfc,omitempty"`
	REPSENumber *string   `json:"repse_number,omitempty"`
	Status      string    `json:"status"`
}

type PlantDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AccessPointDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// CheckOutRequest represents the payload for POST /api/v1/access/check-out.
type CheckOutRequest struct {
	WorkerID      string `json:"worker_id,omitempty"`
	WorkerUserID  string `json:"worker_user_id,omitempty"`
	PlantID       string `json:"plant_id,omitempty"`
	AccessPointID string `json:"access_point_id,omitempty"`
}

// Validate parses and validates identifiers in CheckOutRequest.
func (r *CheckOutRequest) Validate() (workerUUID uuid.UUID, plantUUID *uuid.UUID, apUUID *uuid.UUID, err error) {
	idStr := strings.TrimSpace(r.WorkerUserID)
	if idStr == "" {
		idStr = strings.TrimSpace(r.WorkerID)
	}
	if idStr == "" {
		return uuid.Nil, nil, nil, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador es requerido.")
	}

	workerUUID, err = uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, nil, nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de trabajador no es válido.")
	}

	if strings.TrimSpace(r.PlantID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(r.PlantID))
		if err != nil {
			return uuid.Nil, nil, nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de planta no es válido.")
		}
		plantUUID = &parsed
	}

	if strings.TrimSpace(r.AccessPointID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(r.AccessPointID))
		if err != nil {
			return uuid.Nil, nil, nil, domainErrors.ErrInvalidInput.WithMessage("El formato del UUID de punto de acceso no es válido.")
		}
		apUUID = &parsed
	}

	return workerUUID, plantUUID, apUUID, nil
}

// CheckOutResponse represents the data payload for check-out response.
type CheckOutResponse struct {
	AccessLogID     uuid.UUID           `json:"access_log_id"`
	Result          entity.AccessResult `json:"result,omitempty"`
	CheckedInAt     time.Time           `json:"checked_in_at,omitempty"`
	CheckedOutAt    time.Time           `json:"checked_out_at"`
	DurationMinutes int                 `json:"duration_minutes"`
}

// ListAccessLogsQuery contains parsed query parameters for GET /api/v1/access-logs.
type ListAccessLogsQuery struct {
	PlantID       *uuid.UUID
	AccessPointID *uuid.UUID
	WorkerUserID  *uuid.UUID
	Result        *entity.AccessResult
	From          *time.Time
	To            *time.Time
	Page          int
	PerPage       int
}
```

### 5.2. `internal/adapter/http/dto/envelope.go` Enhancements
```go
// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// WriteSuccessWithMeta wraps data and pagination metadata in standard envelope.
func WriteSuccessWithMeta(w http.ResponseWriter, status int, data interface{}, meta PaginationMeta) {
	WriteJSON(w, status, SuccessResponse{
		Data: data,
		Meta: meta,
	})
}
```

### 5.3. `internal/adapter/http/handler/access_handler.go` Query Parsing Logic
```go
func parseDateFilter(val string) (*time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	// Try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", val); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	return nil, errors.New("formato de fecha inválido")
}
```

---

## 6. Verification and Test Plan

To ensure zero regressions and robust validation, the following test suite must be verified:

1. **Unit Testing (`internal/adapter/http/handler/access_handler_test.go`)**:
   - `TestAccessHandler_CheckIn_Compliant`: Asserts HTTP 200, result="granted", non-empty `access_log_id`, valid `document_snapshot` and `documents`.
   - `TestAccessHandler_CheckIn_Denied`: Asserts HTTP 200, result="denied", descriptive Spanish `denial_reason`.
   - `TestAccessHandler_CheckIn_MalformedJSON`: Asserts HTTP 400 with `INVALID_INPUT`.
   - `TestAccessHandler_CheckIn_MalformedUUID`: Asserts HTTP 400 with `INVALID_INPUT`.
   - `TestAccessHandler_CheckIn_Unauthenticated`: Asserts HTTP 401 with `UNAUTHORIZED`.
   - `TestAccessHandler_CheckIn_WorkerNotFound`: Asserts HTTP 404 with `WORKER_NOT_FOUND`.
   - `TestAccessHandler_CheckOut_Success`: Asserts HTTP 200, non-empty `checked_out_at`, positive `duration_minutes`.
   - `TestAccessHandler_CheckOut_DurationCalculation`: Asserts accurate minutes difference between check-in and check-out.
   - `TestAccessHandler_ListLogs_DefaultPagination`: Asserts default page=1, per_page=20.
   - `TestAccessHandler_ListLogs_WithAllFilters`: Asserts `plant_id`, `access_point_id`, `worker_user_id`, `result`, `from`, `to` filter parameters correctly pass to usecase.
   - `TestAccessHandler_ListLogs_InvalidDateFilter`: Asserts HTTP 400 on malformed `from` or `to` parameter.
2. **Router Wiring Testing (`internal/adapter/http/router_test.go`)**:
   - Ensure all routes are wired under `/api/v1` and protected by `Auth` middleware.
3. **E2E Suite Compatibility**:
   - `test/e2e/tier1_feature_test.sh` (Features 4 & 5).
   - `test/e2e/tier2_boundary_test.sh` (Boundaries 2.B4–2.B14).
   - `test/e2e/tier3_combination_test.sh` (Flows 1, 2, 3, 4, 5).
   - `test/e2e/tier4_workload_test.sh` (15-worker burst, SLA < 500ms).
