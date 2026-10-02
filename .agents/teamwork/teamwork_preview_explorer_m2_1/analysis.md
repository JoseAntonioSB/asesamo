# ASESAMO-API Clean Architecture & Domain Foundation Analysis

**Milestone**: M2: Backend Core & Authentication (`asesamo-api`)  
**Author**: `teamwork_preview_explorer_m2_1`  
**Date**: 2026-10-01  
**Working Directory**: `/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m2_1`

---

## 1. Executive Summary & Problem Boundary

The objective of Milestone M2 is to establish the backend foundation for `asesamo-api`: a robust, production-grade Go REST API adhering strictly to Clean Architecture, Mexican industrial safety compliance specifications (DC-3, REPSE, SUA, INE, Medical), multi-tenant isolation policies, and automated end-to-end testing criteria.

This analysis provides:
1. **The exact `go.mod` definition** with specified modules and version alignment.
2. **The complete directory layout** mapping layers (Domain, Use Cases, Adapters, Infrastructure, Packages).
3. **Exact Go struct definitions** for all 7+ core domain entities and complex value types (e.g. JSONB document snapshots).
4. **Complete Go repository interface contracts** enforcing tenant isolation and clean boundary separation.
5. **Standardized domain error definitions** with Spanish user-facing messages and HTTP envelope mapping.
6. **Cross-cutting packages** (`pkg/tenant`, `pkg/jwt`) and infrastructure adapters (`internal/infra/config`, `internal/infra/database`).
7. **Concrete implementation guidance** for subsequent worker agents implementing M2 and M3.

---

## 2. `go.mod` Specification & Compatibility Analysis

### 2.1 Proposed `go.mod` Content

```go
module asesamo-api

go 1.22

require (
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.23.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.32.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

### 2.2 Dependency Rationale & Compatibility Notes
- **`github.com/go-chi/chi/v5`**: Lightweight, idiomatic HTTP router fully compatible with standard `net/http` handlers and context propagation.
- **`github.com/go-chi/cors`**: Standard CORS middleware handling preflight `OPTIONS` requests seamlessly for Next.js frontend calls.
- **`github.com/golang-jwt/jwt/v5`**: Industry-standard JWT signing and verification library with standard claims support (HS256).
- **`github.com/google/uuid`**: RFC 4122 compliant UUID generation and string parsing.
- **`github.com/jackc/pgx/v5` & `github.com/jackc/pgx/v5/stdlib`**: High-performance PostgreSQL driver supporting PostgreSQL 16 features (`citext`, `pgcrypto`, `JSONB`) plugged directly into Go's standard `database/sql` connection pooling.
- **`golang.org/x/crypto/bcrypt`**: Native bcrypt password hashing and comparison matching PostgreSQL `pgcrypto` `crypt(..., gen_salt('bf', 10))`.
- **`github.com/go-playground/validator/v10`**: High-speed struct field validation for incoming HTTP DTOs.
- **`github.com/stretchr/testify`**: Assertion (`assert`, `require`) and mocking framework for unit testing domain use cases without IO.

### 2.3 Host Toolchain Observation & Caveat
- Host toolchain currently installed: `go version go1.21.4 darwin/arm64`.
- The project requirement specifies Go 1.22+. In Go 1.21+, the toolchain directive (`toolchain go1.22.0`) allows automated toolchain downloading if network is available. For containerized local orchestration (`docker compose`), the `Dockerfile` will use `golang:1.22-alpine` or `golang:1.23-alpine` ensuring pure 1.22+ execution.

---

## 3. Directory Layout for `asesamo-api/`

```
asesamo-api/
├── cmd/
│   └── api/
│       └── main.go                         # Server bootstrap, config loading, DB pooling, DI wiring, graceful shutdown
├── internal/
│   ├── core/
│   │   ├── domain/
│   │   │   ├── entity/                     # Pure business domain entities & enums (no external IO deps)
│   │   │   │   ├── tenant.go
│   │   │   │   ├── plant.go
│   │   │   │   ├── access_point.go
│   │   │   │   ├── user.go
│   │   │   │   ├── contractor.go
│   │   │   │   ├── document.go
│   │   │   │   └── access_log.go
│   │   │   ├── repository/                 # Repository interface contracts (consumed by usecases)
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── access_point_repository.go
│   │   │   │   ├── plant_repository.go
│   │   │   │   ├── document_repository.go
│   │   │   │   └── access_log_repository.go
│   │   │   └── errors/                     # Standard domain error sentinels & codes
│   │   │       └── errors.go
│   │   └── usecase/                        # Application business logic
│   │       ├── auth/
│   │       │   ├── auth_usecase.go         # Login, credential verification, JWT issuance
│   │       │   └── auth_usecase_test.go
│   │       └── access/
│   │           ├── access_usecase.go       # Check-in compliance evaluation, check-out, log retrieval
│   │           └── access_usecase_test.go
│   ├── adapter/
│   │   ├── http/                           # Inbound HTTP transport layer (Chi router, handlers, DTOs, middlewares)
│   │   │   ├── dto/
│   │   │   │   ├── envelope.go             # Standard { "data": ... } and { "error": ... } envelopes
│   │   │   │   ├── auth_dto.go             # Login request/response structs
│   │   │   │   ├── access_point_dto.go     # Access point catalog DTOs
│   │   │   │   └── access_dto.go           # Check-in, check-out, access log DTOs
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go         # POST /api/v1/auth/login
│   │   │   │   ├── access_point_handler.go # GET /api/v1/access-points
│   │   │   │   ├── access_handler.go       # POST /api/v1/access/check-in, check-out, GET /api/v1/access-logs
│   │   │   │   └── health_handler.go       # GET /health, GET /api/v1/health
│   │   │   ├── middleware/
│   │   │   │   ├── auth_middleware.go      # JWT extraction, role extraction, tenant context injection
│   │   │   │   ├── tenant_middleware.go    # Strict tenant verification
│   │   │   │   ├── logger_middleware.go    # Structured HTTP request logging
│   │   │   │   └── cors_middleware.go      # Chi CORS configuration
│   │   │   └── router.go                   # Chi router assembly & route mounting
│   │   └── postgres/                       # Outbound persistence adapter (implements domain/repository interfaces)
│   │       ├── user_repo.go
│   │       ├── access_point_repo.go
│   │       ├── plant_repo.go
│   │       ├── document_repo.go
│   │       └── access_log_repo.go
│   └── infra/
│       ├── config/                         # Environment configuration loading & validation
│       │   └── config.go
│       └── database/                       # PostgreSQL connection pooling & healthcheck
│           └── postgres.go
├── pkg/                                    # Shared foundational packages (zero domain awareness)
│   ├── jwt/                                # JWT issuance, parsing, validation
│   │   ├── jwt.go
│   │   └── jwt_test.go
│   └── tenant/                             # Tenant context key, WithTenant, FromContext helpers
│       ├── context.go
│       └── context_test.go
├── Dockerfile                              # Multi-stage production container build
├── go.mod
└── go.sum
```

---

## 4. Exact Go Struct Definitions for Core Domain Entities

### 4.1 Tenant Entity (`internal/core/domain/entity/tenant.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantPlan represents the subscription tier of a customer tenant.
type TenantPlan string

const (
	TenantPlanStarter    TenantPlan = "starter"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// Tenant represents a paying customer company operating one or more industrial plants.
type Tenant struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	RFC       *string    `json:"rfc,omitempty"`
	LogoR2Key *string    `json:"logo_r2_key,omitempty"`
	Plan      TenantPlan `json:"plan"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

### 4.2 Plant Entity (`internal/core/domain/entity/plant.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Plant represents an industrial facility belonging to a tenant.
type Plant struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	Name               string     `json:"name"`
	Address            *string    `json:"address,omitempty"`
	City               *string    `json:"city,omitempty"`
	State              *string    `json:"state,omitempty"`
	BlockOnExpiredDocs bool       `json:"block_on_expired_docs"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}
```

### 4.3 AccessPoint Entity (`internal/core/domain/entity/access_point.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// AccessPoint represents a physical checkpoint, gate, or caseta at an industrial plant.
type AccessPoint struct {
	ID          uuid.UUID  `json:"id"`
	PlantID     uuid.UUID  `json:"plant_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```

### 4.4 User & Worker Profile Entity (`internal/core/domain/entity/user.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the authorization tier of a user in ASESAMO.
type UserRole string

const (
	RoleSuperAdmin      UserRole = "super_admin"
	RolePlantAdmin      UserRole = "plant_admin"
	RoleEHSManager      UserRole = "ehs_manager"
	RoleGuard           UserRole = "guard"
	RoleContractorAdmin UserRole = "contractor_admin"
	RoleWorker          UserRole = "worker"
)

// User represents an identity with system access or worker registry.
type User struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"` // nil for super_admin
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     string     `json:"full_name"`
	Phone        *string    `json:"phone,omitempty"`
	Role         UserRole   `json:"role"`
	AvatarR2Key  *string    `json:"avatar_r2_key,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// WorkerProfile combines user credentials with contractor worker attributes (CURP, IMSS, Contractor).
type WorkerProfile struct {
	UserID        uuid.UUID  `json:"user_id"`
	FullName      string     `json:"full_name"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	AvatarR2Key   *string    `json:"avatar_r2_key,omitempty"`
	IsActive      bool       `json:"is_active"`
	ContractorID  uuid.UUID  `json:"contractor_id"`
	ContractorName string    `json:"contractor_name"`
	ContractorRFC *string    `json:"contractor_rfc,omitempty"`
	RepseNumber   *string    `json:"repse_number,omitempty"`
	JobTitle      *string    `json:"job_title,omitempty"`
	CURP          *string    `json:"curp,omitempty"`
	IMSSNumber    *string    `json:"imss_number,omitempty"`
}
```

### 4.5 Contractor Entities (`internal/core/domain/entity/contractor.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// ContractorType specifies whether contractor is a legal entity or physical person.
type ContractorType string

const (
	ContractorTypeCompany    ContractorType = "company"
	ContractorTypeIndividual ContractorType = "individual"
)

// PlantContractorStatus represents the relationship status between a plant and a contractor.
type PlantContractorStatus string

const (
	PlantContractorPending   PlantContractorStatus = "pending"
	PlantContractorActive    PlantContractorStatus = "active"
	PlantContractorSuspended PlantContractorStatus = "suspended"
	PlantContractorBlocked   PlantContractorStatus = "blocked"
)

// Contractor represents an external vendor company sending workers to plants.
type Contractor struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	RFC            *string        `json:"rfc,omitempty"`
	RepseNumber    *string        `json:"repse_number,omitempty"`
	ContractorType ContractorType `json:"contractor_type"`
	AdminUserID    *uuid.UUID     `json:"admin_user_id,omitempty"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty"`
}

// PlantContractor represents plant-level activation and authorization of a contractor.
type PlantContractor struct {
	ID           uuid.UUID             `json:"id"`
	PlantID      uuid.UUID             `json:"plant_id"`
	TenantID     uuid.UUID             `json:"tenant_id"`
	ContractorID uuid.UUID             `json:"contractor_id"`
	Status       PlantContractorStatus `json:"status"`
	InvitedBy    *uuid.UUID            `json:"invited_by,omitempty"`
	InvitedAt    time.Time             `json:"invited_at"`
	ActivatedAt  *time.Time            `json:"activated_at,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// ContractorWorker represents worker association to a contractor company.
type ContractorWorker struct {
	ID           uuid.UUID  `json:"id"`
	ContractorID uuid.UUID  `json:"contractor_id"`
	UserID       uuid.UUID  `json:"user_id"`
	JobTitle     *string    `json:"job_title,omitempty"`
	CURP         *string    `json:"curp,omitempty"`
	IMSSNumber   *string    `json:"imss_number,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
```

### 4.6 Document Entities (`internal/core/domain/entity/document.go`)

```go
package entity

import (
	"time"

	"github.com/google/uuid"
)

// DocumentAppliesTo indicates if document belongs to a worker or contractor company.
type DocumentAppliesTo string

const (
	AppliesToWorker     DocumentAppliesTo = "worker"
	AppliesToContractor DocumentAppliesTo = "contractor"
)

// DocumentStatus represents the review and validity lifecycle of an uploaded document.
type DocumentStatus string

const (
	DocStatusPendingReview DocumentStatus = "pending_review"
	DocStatusApproved      DocumentStatus = "approved"
	DocStatusRejected      DocumentStatus = "rejected"
	DocStatusExpired       DocumentStatus = "expired"
)

// DocumentType represents a regulatory or custom document definition.
type DocumentType struct {
	ID                  uuid.UUID         `json:"id"`
	TenantID            *uuid.UUID        `json:"tenant_id,omitempty"` // nil for global system types
	Code                string            `json:"code"`                // DC3, REPSE, SUA, etc.
	Name                string            `json:"name"`
	Description         *string           `json:"description,omitempty"`
	AppliesTo           DocumentAppliesTo `json:"applies_to"`
	DefaultValidityDays *int              `json:"default_validity_days,omitempty"`
	IsRequiredDefault   bool              `json:"is_required_default"`
	IsSystem            bool              `json:"is_system"`
	IsActive            bool              `json:"is_active"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// PlantDocumentRequirement represents a plant's specific required document rule.
type PlantDocumentRequirement struct {
	ID             uuid.UUID `json:"id"`
	PlantID        uuid.UUID `json:"plant_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	DocumentTypeID uuid.UUID `json:"document_type_id"`
	IsRequired     bool      `json:"is_required"`
	ValidityDays   *int      `json:"validity_days,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PlantDocumentRequirementWithDocType enriches the requirement with DocumentType metadata.
type PlantDocumentRequirementWithDocType struct {
	PlantDocumentRequirement
	DocumentTypeCode string            `json:"document_type_code"`
	DocumentTypeName string            `json:"document_name"`
	AppliesTo        DocumentAppliesTo `json:"applies_to"`
}

// Document represents an uploaded compliance document file.
type Document struct {
	ID              uuid.UUID      `json:"id"`
	UserID          *uuid.UUID     `json:"user_id,omitempty"`
	ContractorID    *uuid.UUID     `json:"contractor_id,omitempty"`
	DocumentTypeID  uuid.UUID      `json:"document_type_id"`
	R2Key           string         `json:"r2_key"`
	FileName        string         `json:"file_name"`
	FileSizeBytes   *int64         `json:"file_size_bytes,omitempty"`
	MimeType        *string        `json:"mime_type,omitempty"`
	IssuedAt        *time.Time     `json:"issued_at,omitempty"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	Status          DocumentStatus `json:"status"`
	ReviewedBy      *uuid.UUID     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time     `json:"reviewed_at,omitempty"`
	RejectionReason *string        `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       *time.Time     `json:"deleted_at,omitempty"`
}

// DocumentWithDocType enriches Document with code and name for evaluation engine.
type DocumentWithDocType struct {
	Document
	TypeCode string            `json:"type_code"`
	TypeName string            `json:"type_name"`
	AppliesTo DocumentAppliesTo `json:"applies_to"`
}
```

### 4.7 AccessLog Entity & JSONB Snapshot (`internal/core/domain/entity/access_log.go`)

```go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AccessResult represents the outcome of an access verification check-in.
type AccessResult string

const (
	AccessGranted            AccessResult = "granted"
	AccessGrantedWithWarning AccessResult = "granted_with_warning"
	AccessDenied             AccessResult = "denied"
)

// DocumentSnapshotItem represents the compliance state of a document at check-in time.
type DocumentSnapshotItem struct {
	DocumentTypeCode string  `json:"document_type_code"`
	DocumentName     string  `json:"document_name"`
	AppliesTo        string  `json:"applies_to"`
	Status           string  `json:"status"`
	IsRequired       bool    `json:"is_required"`
	IsValid          bool    `json:"is_valid"`
	ExpiresAt        *string `json:"expires_at"` // ISO date "YYYY-MM-DD" or nil
	FailureReason    *string `json:"failure_reason"`
}

// DocumentSnapshot is a slice of DocumentSnapshotItem supporting database/sql JSONB serialization.
type DocumentSnapshot []DocumentSnapshotItem

// Value implements driver.Valuer for PostgreSQL JSONB column.
func (d DocumentSnapshot) Value() (driver.Value, error) {
	if d == nil {
		return "[]", nil
	}
	return json.Marshal(d)
}

// Scan implements sql.Scanner for reading PostgreSQL JSONB column into Go struct.
func (d *DocumentSnapshot) Scan(value any) error {
	if value == nil {
		*d = DocumentSnapshot{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("entity: unsupported type for DocumentSnapshot scan: %T", value)
	}

	if len(bytes) == 0 {
		*d = DocumentSnapshot{}
		return nil
	}

	return json.Unmarshal(bytes, d)
}

// AccessLog represents an immutable record of check-in and check-out at an access point.
type AccessLog struct {
	ID               uuid.UUID        `json:"id"`
	TenantID         uuid.UUID        `json:"tenant_id"`
	PlantID          uuid.UUID        `json:"plant_id"`
	AccessPointID    uuid.UUID        `json:"access_point_id"`
	UserID           uuid.UUID        `json:"user_id"`
	ContractorID     *uuid.UUID       `json:"contractor_id,omitempty"`
	ProcessedBy      *uuid.UUID       `json:"processed_by,omitempty"`
	CheckedInAt      time.Time        `json:"checked_in_at"`
	CheckedOutAt     *time.Time       `json:"checked_out_at,omitempty"`
	Result           AccessResult     `json:"result"`
	DenialReason     *string          `json:"denial_reason,omitempty"`
	DocumentSnapshot DocumentSnapshot `json:"document_snapshot"`
	Notes            *string          `json:"notes,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`

	// Populated in joined list views
	WorkerName      *string `json:"worker_name,omitempty"`
	ContractorName  *string `json:"contractor_name,omitempty"`
	AccessPointName *string `json:"access_point_name,omitempty"`
}

// AccessLogFilter supports querying and pagination of access logs.
type AccessLogFilter struct {
	TenantID    uuid.UUID
	PlantID     *uuid.UUID
	UserID      *uuid.UUID
	Result      *AccessResult
	CheckedInAt *time.Time
	Page        int
	PerPage     int
}
```

---

## 5. Domain Repository Interfaces

All repository contracts reside in `internal/core/domain/repository/` and adhere strictly to Clean Architecture:
- Interfaces are defined where they are consumed (by use cases).
- Every query accessing tenant-scoped data explicitly receives `tenantID uuid.UUID` or extracts it from `ctx`.
- No HTTP or presentation concerns exist in repository signatures.

### 5.1 UserRepository (`internal/core/domain/repository/user_repository.go`)

```go
package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// UserRepository manages persistence and querying for users and worker profiles.
type UserRepository interface {
	// GetByID returns the user by ID.
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail returns the active user matching the specified email address (citext).
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// UpdateLastLogin updates the last_login_at timestamp for a user.
	UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLogin time.Time) error

	// GetWorkerProfile retrieves worker details including contractor information, CURP, NSS.
	GetWorkerProfile(ctx context.Context, workerUserID uuid.UUID) (*entity.WorkerProfile, error)
}
```

### 5.2 AccessPointRepository (`internal/core/domain/repository/access_point_repository.go`)

```go
package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// AccessPointRepository defines operations for caseta and gate checkpoint catalogs.
type AccessPointRepository interface {
	// GetByID retrieves an access point enforcing tenant isolation (WHERE tenant_id = $1).
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessPoint, error)

	// ListByTenant returns all active access points for the guard's tenant.
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error)

	// ListByPlant returns access points scoped to a specific plant within the tenant.
	ListByPlant(ctx context.Context, tenantID uuid.UUID, plantID uuid.UUID) ([]*entity.AccessPoint, error)
}
```

### 5.3 PlantRepository (`internal/core/domain/repository/plant_repository.go`)

```go
package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// PlantRepository defines plant configuration and contractor relationship lookups.
type PlantRepository interface {
	// GetByID retrieves a plant with its access policy configuration (e.g. block_on_expired_docs).
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.Plant, error)

	// GetPlantContractor checks whether a contractor is activated and in good standing for a plant.
	GetPlantContractor(ctx context.Context, tenantID uuid.UUID, plantID uuid.UUID, contractorID uuid.UUID) (*entity.PlantContractor, error)
}
```

### 5.4 DocumentRepository (`internal/core/domain/repository/document_repository.go`)

```go
package repository

import (
	"context"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// DocumentRepository provides access to document requirements and uploaded worker/contractor records.
type DocumentRepository interface {
	// GetRequirementsByPlant returns the required document types and validity rules configured for a plant.
	GetRequirementsByPlant(ctx context.Context, tenantID uuid.UUID, plantID uuid.UUID) ([]*entity.PlantDocumentRequirementWithDocType, error)

	// GetWorkerDocuments retrieves the latest active documents uploaded for a specific worker user.
	GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]*entity.DocumentWithDocType, error)

	// GetContractorDocuments retrieves the latest active company-level documents (e.g. REPSE, SUA).
	GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]*entity.DocumentWithDocType, error)
}
```

### 5.5 AccessLogRepository (`internal/core/domain/repository/access_log_repository.go`)

```go
package repository

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// AccessLogRepository handles immutable recording and retrieval of gate check-ins.
type AccessLogRepository interface {
	// Create persists an immutable check-in log entry with its JSONB document snapshot.
	Create(ctx context.Context, log *entity.AccessLog) error

	// GetByID retrieves a single access log entry enforcing tenant isolation.
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessLog, error)

	// List queries access logs according to filter criteria and pagination parameters.
	List(ctx context.Context, filter entity.AccessLogFilter) ([]*entity.AccessLog, int, error)

	// UpdateCheckOut stamps the checked_out_at time for an active entry.
	UpdateCheckOut(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, checkedOutAt time.Time) error

	// FindActiveEntryByWorker finds the latest entry for a worker that hasn't checked out yet.
	FindActiveEntryByWorker(ctx context.Context, tenantID uuid.UUID, plantID uuid.UUID, workerUserID uuid.UUID) (*entity.AccessLog, error)
}
```

---

## 6. Domain Errors & HTTP Envelope Mapping

Per `DEVELOPMENT_RULES.md` and `PROJECT.md`, error responses must always be enclosed in:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Mensaje en español dirigido al usuario final."
  }
}
```

### 6.1 `internal/core/domain/errors/errors.go`

```go
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// DomainError represents a structured business error with machine code, user message, and HTTP status.
type DomainError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError constructs a new DomainError instance.
func NewDomainError(code, message string, httpStatus int, err error) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Predefined Sentinel Domain Errors
var (
	ErrInvalidCredentials = &DomainError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Credenciales inválidas. Verifique su correo electrónico y contraseña.",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrUnauthorized = &DomainError{
		Code:       "UNAUTHORIZED",
		Message:    "No autenticado o token de sesión inválido.",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &DomainError{
		Code:       "FORBIDDEN",
		Message:    "No tiene permisos suficientes para realizar esta acción.",
		HTTPStatus: http.StatusForbidden,
	}

	ErrUserNotFound = &DomainError{
		Code:       "USER_NOT_FOUND",
		Message:    "El usuario no existe o se encuentra inactivo.",
		HTTPStatus: http.StatusNotFound,
	}

	ErrWorkerNotFound = &DomainError{
		Code:       "WORKER_NOT_FOUND",
		Message:    "Trabajador no encontrado en el sistema o registro inexistente.",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAccessPointNotFound = &DomainError{
		Code:       "ACCESS_POINT_NOT_FOUND",
		Message:    "Punto de acceso no encontrado o no pertenece al inquilino correspondiente.",
		HTTPStatus: http.StatusNotFound,
	}

	ErrPlantNotFound = &DomainError{
		Code:       "PLANT_NOT_FOUND",
		Message:    "Planta industrial no encontrada o no pertenece al inquilino correspondiente.",
		HTTPStatus: http.StatusNotFound,
	}

	ErrContractorNotFound = &DomainError{
		Code:       "CONTRACTOR_NOT_FOUND",
		Message:    "Empresa contratista no encontrada.",
		HTTPStatus: http.StatusNotFound,
	}

	ErrContractorBlocked = &DomainError{
		Code:       "CONTRACTOR_BLOCKED",
		Message:    "La empresa contratista se encuentra inactiva, bloqueada o suspendida para esta planta.",
		HTTPStatus: http.StatusOK, // Access check-in returns HTTP 200 with result='denied'
	}

	ErrWorkerInactive = &DomainError{
		Code:       "WORKER_INACTIVE",
		Message:    "El perfil del trabajador se encuentra inactivo.",
		HTTPStatus: http.StatusOK, // Evaluated to denied in check-in flow
	}

	ErrTenantMismatch = &DomainError{
		Code:       "TENANT_MISMATCH",
		Message:    "Violación de aislamiento de inquilino: el recurso solicitado no pertenece a su tenant.",
		HTTPStatus: http.StatusNotFound, // Prevent resource probing across tenants
	}

	ErrValidation = &DomainError{
		Code:       "VALIDATION_ERROR",
		Message:    "Los parámetros de la solicitud son inválidos o están incompletos.",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInternal = &DomainError{
		Code:       "INTERNAL_ERROR",
		Message:    "Ocurrió un error interno en el servidor. Intente más tarde.",
		HTTPStatus: http.StatusInternalServerError,
	}
)
```

---

## 7. Cross-Cutting Packages

### 7.1 `pkg/tenant/context.go`

Enforces Section 3 of `DEVELOPMENT_RULES.md`:
> "Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones."

```go
package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type tenantCtxKey struct{}

var ErrMissingTenant = errors.New("tenant: missing tenant_id in request context")

// WithTenant returns a new context containing the tenant UUID.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantCtxKey{}, tenantID)
}

// FromContext extracts the tenant UUID from context.
func FromContext(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value(tenantCtxKey{})
	if val == nil {
		return uuid.Nil, ErrMissingTenant
	}

	tenantID, ok := val.(uuid.UUID)
	if !ok || tenantID == uuid.Nil {
		return uuid.Nil, ErrMissingTenant
	}

	return tenantID, nil
}
```

### 7.2 `pkg/jwt/jwt.go`

```go
package jwt

import (
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("jwt: token is invalid or corrupted")
	ErrExpiredToken = errors.New("jwt: token has expired")
)

// Claims represents the JWT payload structure for ASESAMO users.
type Claims struct {
	jwtlib.RegisteredClaims
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	Email    string    `json:"email"`
}

// Manager handles signing and token validation.
type Manager struct {
	secret        []byte
	issuer        string
	tokenDuration time.Duration
}

// NewManager initializes a JWT manager with secret and lifetime.
func NewManager(secret string, issuer string, duration time.Duration) *Manager {
	return &Manager{
		secret:        []byte(secret),
		issuer:        issuer,
		tokenDuration: duration,
	}
}

// GenerateToken creates and signs a new JWT token.
func (m *Manager) GenerateToken(userID uuid.UUID, tenantID uuid.UUID, role string, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    m.issuer,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.tokenDuration)),
		},
		TenantID: tenantID,
		Role:     role,
		Email:    email,
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("jwt: signing failed: %w", err)
	}
	return signed, nil
}

// ValidateToken verifies signature and expiry, returning parsed Claims.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(t *jwtlib.Token) (any, error) {
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
```

---

## 8. HTTP Envelopes & DTOs

### 8.1 Standard Envelope (`internal/adapter/http/dto/envelope.go`)

```go
package dto

// SuccessEnvelope standardizes all successful API responses.
type SuccessEnvelope[T any] struct {
	Data T          `json:"data"`
	Meta *Pagination `json:"meta,omitempty"`
}

// Pagination metadata for list endpoints.
type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Total   int `json:"total"`
}

// ErrorDetail contains machine code and Spanish user-facing message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorEnvelope standardizes all error API responses.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}
```

### 8.2 Authentication DTOs (`internal/adapter/http/dto/auth_dto.go`)

```go
package dto

import "github.com/google/uuid"

// LoginRequest defines input for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserDTO represents user profile information returned on login.
type UserDTO struct {
	ID       uuid.UUID  `json:"id"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
	Email    string     `json:"email"`
	FullName string     `json:"full_name"`
	Role     string     `json:"role"`
}

// LoginResponse defines payload inside data envelope on successful authentication.
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}
```

### 8.3 Check-In & Access DTOs (`internal/adapter/http/dto/access_dto.go`)

```go
package dto

import (
	"time"

	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// CheckInRequest supports both worker_id and worker_user_id parameters for compatibility.
type CheckInRequest struct {
	WorkerID      *uuid.UUID `json:"worker_id,omitempty"`
	WorkerUserID  *uuid.UUID `json:"worker_user_id,omitempty"`
	AccessPointID uuid.UUID  `json:"access_point_id" validate:"required"`
	Notes         *string    `json:"notes,omitempty"`
}

// GetWorkerUUID safely extracts the worker UUID from either alias.
func (r *CheckInRequest) GetWorkerUUID() uuid.UUID {
	if r.WorkerUserID != nil && *r.WorkerUserID != uuid.Nil {
		return *r.WorkerUserID
	}
	if r.WorkerID != nil && *r.WorkerID != uuid.Nil {
		return *r.WorkerID
	}
	return uuid.Nil
}

// WorkerSummaryDTO represents worker metadata in check-in response.
type WorkerSummaryDTO struct {
	UserID       uuid.UUID `json:"user_id"`
	FullName     string    `json:"full_name"`
	CURP         *string   `json:"curp,omitempty"`
	IMSSNumber   *string   `json:"imss_number,omitempty"`
	JobTitle     *string   `json:"job_title,omitempty"`
	AvatarR2Key  *string   `json:"avatar_r2_key,omitempty"`
}

// ContractorSummaryDTO represents contractor info in check-in response.
type ContractorSummaryDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	RFC         *string   `json:"rfc,omitempty"`
	RepseNumber *string   `json:"repse_number,omitempty"`
	Status      string    `json:"status"`
}

// PlantSummaryDTO represents plant metadata.
type PlantSummaryDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// AccessPointSummaryDTO represents access point metadata.
type AccessPointSummaryDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// CheckInResponse defines the response structure for POST /api/v1/access/check-in.
type CheckInResponse struct {
	AccessLogID      uuid.UUID               `json:"access_log_id"`
	Result           string                  `json:"result"` // granted, granted_with_warning, denied
	DenialReason     *string                 `json:"denial_reason"`
	CheckedInAt      time.Time               `json:"checked_in_at"`
	Worker           WorkerSummaryDTO        `json:"worker"`
	Contractor       ContractorSummaryDTO    `json:"contractor"`
	Plant            PlantSummaryDTO         `json:"plant"`
	AccessPoint      AccessPointSummaryDTO   `json:"access_point"`
	DocumentSnapshot entity.DocumentSnapshot `json:"document_snapshot"`
}

// CheckOutRequest defines input for POST /api/v1/access/check-out.
type CheckOutRequest struct {
	WorkerID      *uuid.UUID `json:"worker_id,omitempty"`
	WorkerUserID  *uuid.UUID `json:"worker_user_id,omitempty"`
	PlantID       *uuid.UUID `json:"plant_id,omitempty"`
	AccessPointID *uuid.UUID `json:"access_point_id,omitempty"`
}

// CheckOutResponse defines response for POST /api/v1/access/check-out.
type CheckOutResponse struct {
	AccessLogID  uuid.UUID `json:"access_log_id"`
	CheckedOutAt time.Time `json:"checked_out_at"`
	Status       string    `json:"status"`
}
```

---

## 9. Infrastructure Layer (`config` & `database`)

### 9.1 `internal/infra/config/config.go`

Validates that environment variables are loaded and enforces `JWT_SECRET` >= 32 characters:

```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv             string
	AppPort            int
	DatabaseURL        string
	JWTSecret          string
	CORSAllowedOrigins []string
}

func Load() (*Config, error) {
	appEnv := getEnv("APP_ENV", "development")

	portStr := getEnv("APP_PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("config: invalid APP_PORT: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET must be at least 32 characters long per DEVELOPMENT_RULES.md (current length: %d)", len(jwtSecret))
	}

	corsOriginsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")
	corsOrigins := strings.Split(corsOriginsStr, ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	return &Config{
		AppEnv:             appEnv,
		AppPort:            port,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		CORSAllowedOrigins: corsOrigins,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
```

### 9.2 `internal/infra/database/postgres.go`

Uses `github.com/jackc/pgx/v5/stdlib` for clean connection pooling with `database/sql`:

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect establishes a connection pool to PostgreSQL.
func Connect(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open driver: %w", err)
	}

	// Recommended connection pooling configuration
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}
```

---

## 10. Multi-Tenant PostgreSQL Repository Patterns

Section 3 of `DEVELOPMENT_RULES.md` demands:
> "Todo query a estas tablas DEBE incluir `WHERE tenant_id = $1` como primera condición."

Here is how `AccessPointRepository` and `AccessLogRepository` implement this pattern:

```go
// AccessPointRepository implementation snippet:
func (r *postgresAccessPointRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error) {
	query := `
		SELECT id, plant_id, tenant_id, name, description, is_active, created_at, updated_at
		FROM access_points
		WHERE tenant_id = $1 AND deleted_at IS NULL AND is_active = TRUE
		ORDER BY name ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	// scan into entity slice
}

func (r *postgresAccessPointRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessPoint, error) {
	query := `
		SELECT id, plant_id, tenant_id, name, description, is_active, created_at, updated_at
		FROM access_points
		WHERE tenant_id = $1 AND id = $2 AND deleted_at IS NULL;
	`
	// returns errors.ErrAccessPointNotFound if sql.ErrNoRows
}
```

---

## 11. Synthesis & Next Actions for Milestone M2

1. **Step 1 (Scaffolding)**:
   - Initialize `asesamo-api/go.mod` with the dependencies listed above.
   - Run `go mod tidy` in `asesamo-api/`.
2. **Step 2 (Core Domain & Packages)**:
   - Write `pkg/tenant/context.go` and unit tests.
   - Write `pkg/jwt/jwt.go` and unit tests.
   - Write `internal/core/domain/entity/` structs and `internal/core/domain/errors/errors.go`.
   - Write `internal/core/domain/repository/` interfaces.
3. **Step 3 (Infrastructure & Adapters)**:
   - Implement `internal/infra/config/` and `internal/infra/database/`.
   - Implement `internal/adapter/postgres/` repositories with tenant isolation queries.
4. **Step 4 (Use Cases & Handlers)**:
   - Implement `internal/core/usecase/auth/` and `POST /api/v1/auth/login`.
   - Implement `GET /health` and `GET /api/v1/health`.
   - Implement `GET /api/v1/access-points`.
5. **Step 5 (Unit Testing & Verification)**:
   - Verify `go test ./...` in `asesamo-api/` passes cleanly with 0 errors.
