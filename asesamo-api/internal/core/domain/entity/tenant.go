package entity

import (
	"time"

	"github.com/google/uuid"
)

// TenantPlan represents the subscription tier of a tenant.
type TenantPlan string

const (
	TenantPlanStarter    TenantPlan = "starter"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// Tenant represents the paying customer entity.
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
