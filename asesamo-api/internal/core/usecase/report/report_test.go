package report_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/report"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockReportRepo struct {
	metrics   *repository.AccessMetricsSummary
	timeline  []repository.DailyAccessDistribution
	reasons   []repository.RejectionReasonSummary
	ranking   []repository.ContractorComplianceSummary
	accesses  []repository.AuditAccessRecord
	compls    []repository.AuditComplianceRecord
}

func (m *mockReportRepo) GetAccessMetrics(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) (*repository.AccessMetricsSummary, error) {
	if m.metrics != nil {
		return m.metrics, nil
	}
	return &repository.AccessMetricsSummary{
		TotalAccesses:  10,
		GrantedCount:   8,
		WarningCount:   1,
		DeniedCount:    1,
		ComplianceRate: 90.0,
	}, nil
}

func (m *mockReportRepo) GetAccessTimeline(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.DailyAccessDistribution, error) {
	return m.timeline, nil
}

func (m *mockReportRepo) GetTopRejectionReasons(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time, limit int) ([]repository.RejectionReasonSummary, error) {
	return m.reasons, nil
}

func (m *mockReportRepo) GetContractorComplianceRanking(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.ContractorComplianceSummary, error) {
	return m.ranking, nil
}

func (m *mockReportRepo) ListAuditAccessRecords(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.AuditAccessRecord, error) {
	return m.accesses, nil
}

func (m *mockReportRepo) ListAuditComplianceRecords(ctx context.Context, tenantID, plantID uuid.UUID) ([]repository.AuditComplianceRecord, error) {
	return m.compls, nil
}

func TestGetReportMetricsUseCase(t *testing.T) {
	repo := &mockReportRepo{
		timeline: []repository.DailyAccessDistribution{
			{Date: "2026-10-08", Total: 5, Granted: 4, Denied: 1},
			{Date: "2026-10-09", Total: 5, Granted: 5, Denied: 0},
		},
		reasons: []repository.RejectionReasonSummary{
			{Reason: "Falta Constancia DC-3", Count: 1, Percentage: 100.0},
		},
		ranking: []repository.ContractorComplianceSummary{
			{
				ContractorID:   uuid.New(),
				ContractorName: "Mantenimiento Industrial del Norte",
				TotalAttempts:  10,
				GrantedCount:   9,
				DeniedCount:    1,
				ComplianceRate: 90.0,
			},
		},
	}

	uc := report.NewGetReportMetricsUseCase(repo)
	res, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), time.Time{}, time.Time{})

	require.NoError(t, err)
	assert.Equal(t, 10, res.Summary.TotalAccesses)
	assert.Equal(t, 90.0, res.Summary.ComplianceRate)
	assert.Len(t, res.Timeline, 2)
	assert.Len(t, res.RejectionReasons, 1)
	assert.Len(t, res.ContractorRanking, 1)
}

func TestExportAccessLogsCSVUseCase(t *testing.T) {
	dur := 45
	curp := "MENC850101HDFR01"
	imss := "12345678901"
	job := "Soldador Especializado"
	rfc := "MIN120304AB1"
	repse := "REPSE-NL-2023-8841"
	reason := "Constancia DC-3 Vencida"

	repo := &mockReportRepo{
		accesses: []repository.AuditAccessRecord{
			{
				AccessLogID:     uuid.New(),
				CheckedInAt:     time.Date(2026, 10, 9, 8, 30, 0, 0, time.UTC),
				DurationMinutes: &dur,
				PlantName:       "Planta Querétaro Norte",
				AccessPointName: "Caseta Principal",
				WorkerName:      "Carlos Mendoza",
				WorkerEmail:     "carlos@mantenimiento.com",
				WorkerCURP:      &curp,
				WorkerIMSS:      &imss,
				WorkerJobTitle:  &job,
				ContractorName:  "Mantenimiento Industrial del Norte",
				ContractorRFC:   &rfc,
				ContractorREPSE: &repse,
				Result:          "denied",
				DenialReason:    &reason,
			},
		},
	}

	uc := report.NewExportAccessLogsCSVUseCase(repo)
	csvBytes, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), time.Time{}, time.Time{})

	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(csvBytes, []byte{0xEF, 0xBB, 0xBF}), "must have UTF-8 BOM")

	csvContent := string(csvBytes)
	assert.Contains(t, csvContent, "ID Registro")
	assert.Contains(t, csvContent, "Carlos Mendoza")
	assert.Contains(t, csvContent, "MENC850101HDFR01")
	assert.Contains(t, csvContent, "DENEGADO")
	assert.Contains(t, csvContent, "Constancia DC-3 Vencida")
}

func TestExportComplianceCSVUseCase(t *testing.T) {
	curp := "MENC850101HDFR01"
	imss := "12345678901"
	job := "Soldador"
	rfc := "MIN120304AB1"
	repse := "REPSE-123"
	dc3Exp := "2027-01-15"

	repo := &mockReportRepo{
		compls: []repository.AuditComplianceRecord{
			{
				WorkerID:          uuid.New(),
				WorkerName:        "Carlos Mendoza",
				WorkerEmail:       "carlos@mantenimiento.com",
				WorkerCURP:        &curp,
				WorkerIMSS:        &imss,
				WorkerJobTitle:    &job,
				ContractorName:    "Mantenimiento Industrial del Norte",
				ContractorRFC:     &rfc,
				ContractorREPSE:   &repse,
				DC3Status:         "approved",
				DC3ExpiresAt:      &dc3Exp,
				SUAStatus:         "approved",
				MedicalStatus:     "approved",
				INEStatus:         "approved",
				OverallCompliance: "APTO",
			},
		},
	}

	uc := report.NewExportComplianceCSVUseCase(repo)
	csvBytes, err := uc.Execute(context.Background(), uuid.New(), uuid.New())

	require.NoError(t, err)
	assert.True(t, bytes.HasPrefix(csvBytes, []byte{0xEF, 0xBB, 0xBF}), "must have UTF-8 BOM")

	csvContent := string(csvBytes)
	assert.Contains(t, csvContent, "Dictamen General STPS")
	assert.Contains(t, csvContent, "Carlos Mendoza")
	assert.Contains(t, csvContent, "APTO")
	assert.Contains(t, csvContent, "2027-01-15")
}
