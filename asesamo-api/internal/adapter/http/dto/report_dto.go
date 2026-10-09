package dto

import (
	"time"

	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/report"
)

// ReportMetricsResponse represents the payload for GET /api/v1/reports/metrics.
type ReportMetricsResponse struct {
	Summary           repository.AccessMetricsSummary          `json:"summary"`
	Timeline          []repository.DailyAccessDistribution     `json:"timeline"`
	RejectionReasons  []repository.RejectionReasonSummary      `json:"rejection_reasons"`
	ContractorRanking []repository.ContractorComplianceSummary `json:"contractor_ranking"`
	GeneratedAt       time.Time                                `json:"generated_at"`
}

// ToReportMetricsResponse converts usecase output to HTTP DTO.
func ToReportMetricsResponse(out *report.ReportMetricsOutput) ReportMetricsResponse {
	return ReportMetricsResponse{
		Summary:           out.Summary,
		Timeline:          out.Timeline,
		RejectionReasons:  out.RejectionReasons,
		ContractorRanking: out.ContractorRanking,
		GeneratedAt:       out.GeneratedAt,
	}
}
