package report

import (
	"context"
	"time"

	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ReportMetricsOutput contains all aggregated indicators for the analytics dashboard.
type ReportMetricsOutput struct {
	Summary            repository.AccessMetricsSummary          `json:"summary"`
	Timeline           []repository.DailyAccessDistribution     `json:"timeline"`
	RejectionReasons   []repository.RejectionReasonSummary      `json:"rejection_reasons"`
	ContractorRanking  []repository.ContractorComplianceSummary `json:"contractor_ranking"`
	GeneratedAt        time.Time                                `json:"generated_at"`
}

// GetReportMetricsUseCase coordinates retrieval of analytical indicators for management and EHS.
type GetReportMetricsUseCase struct {
	reportRepo repository.ReportRepository
}

// NewGetReportMetricsUseCase constructs GetReportMetricsUseCase.
func NewGetReportMetricsUseCase(reportRepo repository.ReportRepository) *GetReportMetricsUseCase {
	return &GetReportMetricsUseCase{reportRepo: reportRepo}
}

// Execute retrieves and aggregates all metrics for the requested tenant, plant, and date window.
func (uc *GetReportMetricsUseCase) Execute(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) (*ReportMetricsOutput, error) {
	if from.IsZero() {
		from = time.Now().UTC().AddDate(0, 0, -30) // default last 30 days
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}

	summary, err := uc.reportRepo.GetAccessMetrics(ctx, tenantID, plantID, from, to)
	if err != nil {
		return nil, err
	}

	timeline, err := uc.reportRepo.GetAccessTimeline(ctx, tenantID, plantID, from, to)
	if err != nil {
		return nil, err
	}
	if timeline == nil {
		timeline = []repository.DailyAccessDistribution{}
	}

	reasons, err := uc.reportRepo.GetTopRejectionReasons(ctx, tenantID, plantID, from, to, 5)
	if err != nil {
		return nil, err
	}
	if reasons == nil {
		reasons = []repository.RejectionReasonSummary{}
	}

	ranking, err := uc.reportRepo.GetContractorComplianceRanking(ctx, tenantID, plantID, from, to)
	if err != nil {
		return nil, err
	}
	if ranking == nil {
		ranking = []repository.ContractorComplianceSummary{}
	}

	return &ReportMetricsOutput{
		Summary:           *summary,
		Timeline:          timeline,
		RejectionReasons:  reasons,
		ContractorRanking: ranking,
		GeneratedAt:       time.Now().UTC(),
	}, nil
}
