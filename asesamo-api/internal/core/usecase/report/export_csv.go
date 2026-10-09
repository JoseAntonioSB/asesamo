package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

// ExportAccessLogsCSVUseCase generates a CSV export of access logs compliant with STPS/OSHA audit standards.
type ExportAccessLogsCSVUseCase struct {
	reportRepo repository.ReportRepository
}

// NewExportAccessLogsCSVUseCase constructs ExportAccessLogsCSVUseCase.
func NewExportAccessLogsCSVUseCase(reportRepo repository.ReportRepository) *ExportAccessLogsCSVUseCase {
	return &ExportAccessLogsCSVUseCase{reportRepo: reportRepo}
}

// Execute queries access records and encodes them as a CSV byte slice with UTF-8 BOM.
func (uc *ExportAccessLogsCSVUseCase) Execute(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) ([]byte, error) {
	if from.IsZero() {
		from = time.Now().UTC().AddDate(0, 0, -30)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}

	records, err := uc.reportRepo.ListAuditAccessRecords(ctx, tenantID, plantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("usecase/report: failed to query audit access records: %w", err)
	}

	buf := &bytes.Buffer{}
	// Prepend UTF-8 BOM so Excel automatically recognizes UTF-8 encoded Spanish characters
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(buf)

	// CSV Header in Mexican Spanish
	header := []string{
		"ID Registro",
		"Fecha/Hora Ingreso (UTC)",
		"Fecha/Hora Salida (UTC)",
		"Estancia (Minutos)",
		"Planta",
		"Caseta / Punto de Acceso",
		"Trabajador",
		"Correo Trabajador",
		"CURP",
		"NSS (IMSS)",
		"Puesto / Especialidad",
		"Empresa Contratista",
		"RFC Contratista",
		"Folio REPSE",
		"Resultado Acceso",
		"Motivo de Rechazo",
		"Observaciones",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, r := range records {
		checkedOut := ""
		if r.CheckedOutAt != nil {
			checkedOut = r.CheckedOutAt.Format("2006-01-02 15:04:05")
		}

		durationStr := ""
		if r.DurationMinutes != nil {
			durationStr = strconv.Itoa(*r.DurationMinutes)
		}

		curp := ""
		if r.WorkerCURP != nil {
			curp = *r.WorkerCURP
		}

		imss := ""
		if r.WorkerIMSS != nil {
			imss = *r.WorkerIMSS
		}

		job := ""
		if r.WorkerJobTitle != nil {
			job = *r.WorkerJobTitle
		}

		rfc := ""
		if r.ContractorRFC != nil {
			rfc = *r.ContractorRFC
		}

		repse := ""
		if r.ContractorREPSE != nil {
			repse = *r.ContractorREPSE
		}

		resultText := "PERMITIDO"
		switch r.Result {
		case "denied":
			resultText = "DENEGADO"
		case "granted_with_warning":
			resultText = "ADVERTENCIA"
		}

		denialReason := ""
		if r.DenialReason != nil {
			denialReason = *r.DenialReason
		}

		notes := ""
		if r.Notes != nil {
			notes = *r.Notes
		}

		row := []string{
			r.AccessLogID.String(),
			r.CheckedInAt.Format("2006-01-02 15:04:05"),
			checkedOut,
			durationStr,
			r.PlantName,
			r.AccessPointName,
			r.WorkerName,
			r.WorkerEmail,
			curp,
			imss,
			job,
			r.ContractorName,
			rfc,
			repse,
			resultText,
			denialReason,
			notes,
		}

		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

// ExportComplianceCSVUseCase generates a CSV export of the compliance matrix (DC-3, SUA, IMSS, etc.).
type ExportComplianceCSVUseCase struct {
	reportRepo repository.ReportRepository
}

// NewExportComplianceCSVUseCase constructs ExportComplianceCSVUseCase.
func NewExportComplianceCSVUseCase(reportRepo repository.ReportRepository) *ExportComplianceCSVUseCase {
	return &ExportComplianceCSVUseCase{reportRepo: reportRepo}
}

// Execute queries worker compliance matrix and encodes it as CSV with UTF-8 BOM.
func (uc *ExportComplianceCSVUseCase) Execute(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
) ([]byte, error) {
	records, err := uc.reportRepo.ListAuditComplianceRecords(ctx, tenantID, plantID)
	if err != nil {
		return nil, fmt.Errorf("usecase/report: failed to query compliance records: %w", err)
	}

	buf := &bytes.Buffer{}
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(buf)

	header := []string{
		"ID Trabajador",
		"Trabajador",
		"Correo",
		"CURP",
		"NSS (IMSS)",
		"Puesto",
		"Empresa Contratista",
		"RFC",
		"REPSE STPS",
		"Estatus DC-3",
		"Vigencia DC-3",
		"Estatus IMSS (SUA)",
		"Vigencia IMSS",
		"Estatus Examen Médico",
		"Vigencia Médico",
		"Estatus INE",
		"Vigencia INE",
		"Dictamen General STPS",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, r := range records {
		curp := ""
		if r.WorkerCURP != nil {
			curp = *r.WorkerCURP
		}

		imss := ""
		if r.WorkerIMSS != nil {
			imss = *r.WorkerIMSS
		}

		job := ""
		if r.WorkerJobTitle != nil {
			job = *r.WorkerJobTitle
		}

		rfc := ""
		if r.ContractorRFC != nil {
			rfc = *r.ContractorRFC
		}

		repse := ""
		if r.ContractorREPSE != nil {
			repse = *r.ContractorREPSE
		}

		dc3Exp := ""
		if r.DC3ExpiresAt != nil {
			dc3Exp = *r.DC3ExpiresAt
		}

		suaExp := ""
		if r.SUAExpiresAt != nil {
			suaExp = *r.SUAExpiresAt
		}

		medExp := ""
		if r.MedicalExpiresAt != nil {
			medExp = *r.MedicalExpiresAt
		}

		ineExp := ""
		if r.INEExpiresAt != nil {
			ineExp = *r.INEExpiresAt
		}

		row := []string{
			r.WorkerID.String(),
			r.WorkerName,
			r.WorkerEmail,
			curp,
			imss,
			job,
			r.ContractorName,
			rfc,
			repse,
			r.DC3Status,
			dc3Exp,
			r.SUAStatus,
			suaExp,
			r.MedicalStatus,
			medExp,
			r.INEStatus,
			ineExp,
			r.OverallCompliance,
		}

		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}
