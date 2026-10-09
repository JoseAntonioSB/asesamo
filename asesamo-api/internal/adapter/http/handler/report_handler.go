package handler

import (
	"fmt"
	"net/http"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/report"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

// ReportHandler coordinates metrics and compliance reporting HTTP requests.
type ReportHandler struct {
	getMetricsUseCase       *report.GetReportMetricsUseCase
	exportAccessLogsUseCase *report.ExportAccessLogsCSVUseCase
	exportComplianceUseCase *report.ExportComplianceCSVUseCase
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(
	getMetricsUseCase *report.GetReportMetricsUseCase,
	exportAccessLogsUseCase *report.ExportAccessLogsCSVUseCase,
	exportComplianceUseCase *report.ExportComplianceCSVUseCase,
) *ReportHandler {
	return &ReportHandler{
		getMetricsUseCase:       getMetricsUseCase,
		exportAccessLogsUseCase: exportAccessLogsUseCase,
		exportComplianceUseCase: exportComplianceUseCase,
	}
}

// GetMetrics handles GET /api/v1/reports/metrics.
func (h *ReportHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	plantID, from, to, err := h.parseFilterParams(r)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	out, err := h.getMetricsUseCase.Execute(r.Context(), tenantID, plantID, from, to)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, dto.ToReportMetricsResponse(out))
}

// ExportAccessLogs handles GET /api/v1/reports/export/access-logs.
func (h *ReportHandler) ExportAccessLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	plantID, from, to, err := h.parseFilterParams(r)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	csvData, err := h.exportAccessLogsUseCase.Execute(r.Context(), tenantID, plantID, from, to)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	filename := fmt.Sprintf("reporte_accesos_asesamo_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

// ExportCompliance handles GET /api/v1/reports/export/compliance.
func (h *ReportHandler) ExportCompliance(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	plantID, _, _, err := h.parseFilterParams(r)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	csvData, err := h.exportComplianceUseCase.Execute(r.Context(), tenantID, plantID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	filename := fmt.Sprintf("matriz_cumplimiento_stps_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(csvData)
}

func (h *ReportHandler) parseFilterParams(r *http.Request) (plantID uuid.UUID, from time.Time, to time.Time, err error) {
	plantIDStr := r.URL.Query().Get("plant_id")
	if plantIDStr != "" {
		parsedPlant, pErr := uuid.Parse(plantIDStr)
		if pErr == nil {
			plantID = parsedPlant
		}
	}

	fromStr := r.URL.Query().Get("from")
	if fromStr != "" {
		if t, tErr := time.Parse(time.RFC3339, fromStr); tErr == nil {
			from = t
		} else if t, tErr := time.Parse("2006-01-02", fromStr); tErr == nil {
			from = t
		}
	}

	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		if t, tErr := time.Parse(time.RFC3339, toStr); tErr == nil {
			to = t
		} else if t, tErr := time.Parse("2006-01-02", toStr); tErr == nil {
			to = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	return plantID, from, to, nil
}
