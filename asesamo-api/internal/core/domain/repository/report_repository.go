package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AccessMetricsSummary holds aggregated access KPI indicators.
type AccessMetricsSummary struct {
	TotalAccesses      int     `json:"total_accesses"`
	GrantedCount       int     `json:"granted_count"`
	WarningCount       int     `json:"warning_count"`
	DeniedCount        int     `json:"denied_count"`
	ComplianceRate     float64 `json:"compliance_rate"`      // percentage (0-100)
	AvgDurationMinutes float64 `json:"avg_duration_minutes"` // average stay length
}

// DailyAccessDistribution represents entry logs grouped by date.
type DailyAccessDistribution struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Total   int    `json:"total"`
	Granted int    `json:"granted"`
	Denied  int    `json:"denied"`
}

// RejectionReasonSummary represents an aggregated denial reason.
type RejectionReasonSummary struct {
	Reason     string  `json:"reason"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ContractorComplianceSummary aggregates performance per contractor company.
type ContractorComplianceSummary struct {
	ContractorID   uuid.UUID `json:"contractor_id"`
	ContractorName string    `json:"contractor_name"`
	TotalAttempts  int       `json:"total_attempts"`
	GrantedCount   int       `json:"granted_count"`
	DeniedCount    int       `json:"denied_count"`
	ComplianceRate float64   `json:"compliance_rate"`
}

// AuditAccessRecord represents a flattened check-in log ready for STPS audit export.
type AuditAccessRecord struct {
	AccessLogID       uuid.UUID  `json:"access_log_id"`
	CheckedInAt       time.Time  `json:"checked_in_at"`
	CheckedOutAt      *time.Time `json:"checked_out_at,omitempty"`
	DurationMinutes   *int       `json:"duration_minutes,omitempty"`
	PlantName         string     `json:"plant_name"`
	AccessPointName   string     `json:"access_point_name"`
	WorkerName        string     `json:"worker_name"`
	WorkerEmail       string     `json:"worker_email"`
	WorkerCURP        *string    `json:"worker_curp,omitempty"`
	WorkerIMSS        *string    `json:"worker_imss,omitempty"`
	WorkerJobTitle    *string    `json:"worker_job_title,omitempty"`
	ContractorName    string     `json:"contractor_name"`
	ContractorRFC     *string    `json:"contractor_rfc,omitempty"`
	ContractorREPSE   *string    `json:"contractor_repse,omitempty"`
	Result            string     `json:"result"`
	DenialReason      *string    `json:"denial_reason,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
}

// AuditComplianceRecord represents a worker's multi-document compliance status for STPS matrix export.
type AuditComplianceRecord struct {
	WorkerID          uuid.UUID `json:"worker_id"`
	WorkerName        string    `json:"worker_name"`
	WorkerEmail       string    `json:"worker_email"`
	WorkerCURP        *string   `json:"worker_curp,omitempty"`
	WorkerIMSS        *string   `json:"worker_imss,omitempty"`
	WorkerJobTitle    *string   `json:"worker_job_title,omitempty"`
	ContractorName    string    `json:"contractor_name"`
	ContractorRFC     *string   `json:"contractor_rfc,omitempty"`
	ContractorREPSE   *string   `json:"contractor_repse,omitempty"`
	DC3Status         string    `json:"dc3_status"`
	DC3ExpiresAt      *string   `json:"dc3_expires_at,omitempty"`
	SUAStatus         string    `json:"sua_status"`
	SUAExpiresAt      *string   `json:"sua_expires_at,omitempty"`
	MedicalStatus     string    `json:"medical_status"`
	MedicalExpiresAt  *string   `json:"medical_expires_at,omitempty"`
	INEStatus         string    `json:"ine_status"`
	INEExpiresAt      *string   `json:"ine_expires_at,omitempty"`
	OverallCompliance string    `json:"overall_compliance"` // "APTO" | "NO_APTO" | "EN_REVISION"
}

// ReportRepository defines data aggregation operations for analytics and STPS compliance audits.
type ReportRepository interface {
	GetAccessMetrics(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) (*AccessMetricsSummary, error)
	GetAccessTimeline(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]DailyAccessDistribution, error)
	GetTopRejectionReasons(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time, limit int) ([]RejectionReasonSummary, error)
	GetContractorComplianceRanking(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]ContractorComplianceSummary, error)
	ListAuditAccessRecords(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]AuditAccessRecord, error)
	ListAuditComplianceRecords(ctx context.Context, tenantID, plantID uuid.UUID) ([]AuditComplianceRecord, error)
}
