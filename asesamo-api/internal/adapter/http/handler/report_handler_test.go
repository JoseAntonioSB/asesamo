package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/report"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type handlerMockReportRepo struct{}

func (m *handlerMockReportRepo) GetAccessMetrics(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) (*repository.AccessMetricsSummary, error) {
	return &repository.AccessMetricsSummary{
		TotalAccesses:  50,
		GrantedCount:   45,
		WarningCount:   3,
		DeniedCount:    2,
		ComplianceRate: 96.0,
	}, nil
}

func (m *handlerMockReportRepo) GetAccessTimeline(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.DailyAccessDistribution, error) {
	return []repository.DailyAccessDistribution{
		{Date: "2026-10-09", Total: 50, Granted: 48, Denied: 2},
	}, nil
}

func (m *handlerMockReportRepo) GetTopRejectionReasons(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time, limit int) ([]repository.RejectionReasonSummary, error) {
	return []repository.RejectionReasonSummary{
		{Reason: "Constancia DC-3 Vencida", Count: 2, Percentage: 100.0},
	}, nil
}

func (m *handlerMockReportRepo) GetContractorComplianceRanking(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.ContractorComplianceSummary, error) {
	return []repository.ContractorComplianceSummary{
		{
			ContractorID:   uuid.New(),
			ContractorName: "Mantenimiento Industrial del Norte",
			TotalAttempts:  50,
			GrantedCount:   48,
			DeniedCount:    2,
			ComplianceRate: 96.0,
		},
	}, nil
}

func (m *handlerMockReportRepo) ListAuditAccessRecords(ctx context.Context, tenantID, plantID uuid.UUID, from, to time.Time) ([]repository.AuditAccessRecord, error) {
	curp := "MENC850101HDFR01"
	return []repository.AuditAccessRecord{
		{
			AccessLogID:     uuid.New(),
			CheckedInAt:     time.Now().UTC(),
			PlantName:       "Planta Querétaro Norte",
			AccessPointName: "Caseta Principal",
			WorkerName:      "Carlos Mendoza",
			WorkerEmail:     "carlos@mantenimiento.com",
			WorkerCURP:      &curp,
			ContractorName:  "Mantenimiento Industrial",
			Result:          "granted",
		},
	}, nil
}

func (m *handlerMockReportRepo) ListAuditComplianceRecords(ctx context.Context, tenantID, plantID uuid.UUID) ([]repository.AuditComplianceRecord, error) {
	curp := "MENC850101HDFR01"
	return []repository.AuditComplianceRecord{
		{
			WorkerID:          uuid.New(),
			WorkerName:        "Carlos Mendoza",
			WorkerEmail:       "carlos@mantenimiento.com",
			WorkerCURP:        &curp,
			ContractorName:    "Mantenimiento Industrial",
			DC3Status:         "approved",
			OverallCompliance: "APTO",
		},
	}, nil
}

func TestReportHandler(t *testing.T) {
	mockRepo := &handlerMockReportRepo{}
	getMetricsUC := report.NewGetReportMetricsUseCase(mockRepo)
	exportAccessUC := report.NewExportAccessLogsCSVUseCase(mockRepo)
	exportComplUC := report.NewExportComplianceCSVUseCase(mockRepo)

	h := handler.NewReportHandler(getMetricsUC, exportAccessUC, exportComplUC)

	r := chi.NewRouter()
	r.Get("/api/v1/reports/metrics", h.GetMetrics)
	r.Get("/api/v1/reports/export/access-logs", h.ExportAccessLogs)
	r.Get("/api/v1/reports/export/compliance", h.ExportCompliance)

	tenantID := uuid.New()

	t.Run("GET /api/v1/reports/metrics returns 200 and payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/metrics?from=2026-10-01&to=2026-10-09", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "compliance_rate")
		assert.Contains(t, w.Body.String(), "96")
	})

	t.Run("GET /api/v1/reports/export/access-logs returns CSV with attachment header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/export/access-logs", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=\"reporte_accesos_asesamo_")
		assert.Contains(t, w.Body.String(), "Carlos Mendoza")
	})

	t.Run("GET /api/v1/reports/export/compliance returns CSV with attachment header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/export/compliance", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=\"matriz_cumplimiento_stps_")
		assert.Contains(t, w.Body.String(), "APTO")
	})
}
