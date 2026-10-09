package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"asesamo-api/internal/core/domain/repository"

	"github.com/google/uuid"
)

type reportRepository struct {
	db *sql.DB
}

// NewReportRepository constructs a PostgreSQL implementation of repository.ReportRepository.
func NewReportRepository(db *sql.DB) repository.ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) GetAccessMetrics(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) (*repository.AccessMetricsSummary, error) {
	query := `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE result = 'granted') AS granted,
			COUNT(*) FILTER (WHERE result = 'granted_with_warning') AS warning,
			COUNT(*) FILTER (WHERE result = 'denied') AS denied,
			COALESCE(AVG(EXTRACT(EPOCH FROM (checked_out_at - checked_in_at)) / 60) FILTER (WHERE checked_out_at IS NOT NULL), 0) AS avg_duration
		FROM access_logs
		WHERE tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR plant_id = $2)
		  AND checked_in_at >= $3 AND checked_in_at <= $4
	`

	var (
		total, granted, warning, denied int
		avgDuration                     float64
	)

	err := r.db.QueryRowContext(ctx, query, tenantID, plantID, from, to).Scan(
		&total,
		&granted,
		&warning,
		&denied,
		&avgDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query access metrics: %w", err)
	}

	complianceRate := 0.0
	if total > 0 {
		complianceRate = math.Round((float64(granted+warning)/float64(total))*1000) / 10
	}

	return &repository.AccessMetricsSummary{
		TotalAccesses:      total,
		GrantedCount:       granted,
		WarningCount:       warning,
		DeniedCount:        denied,
		ComplianceRate:     complianceRate,
		AvgDurationMinutes: math.Round(avgDuration*10) / 10,
	}, nil
}

func (r *reportRepository) GetAccessTimeline(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) ([]repository.DailyAccessDistribution, error) {
	query := `
		SELECT
			TO_CHAR(checked_in_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS log_date,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE result IN ('granted', 'granted_with_warning')) AS granted,
			COUNT(*) FILTER (WHERE result = 'denied') AS denied
		FROM access_logs
		WHERE tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR plant_id = $2)
		  AND checked_in_at >= $3 AND checked_in_at <= $4
		GROUP BY log_date
		ORDER BY log_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query access timeline: %w", err)
	}
	defer rows.Close()

	var list []repository.DailyAccessDistribution
	for rows.Next() {
		var d repository.DailyAccessDistribution
		if err := rows.Scan(&d.Date, &d.Total, &d.Granted, &d.Denied); err != nil {
			return nil, fmt.Errorf("postgres/report: failed to scan timeline row: %w", err)
		}
		list = append(list, d)
	}

	return list, nil
}

func (r *reportRepository) GetTopRejectionReasons(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
	limit int,
) ([]repository.RejectionReasonSummary, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `
		WITH total_denied AS (
			SELECT COUNT(*) AS total
			FROM access_logs
			WHERE tenant_id = $1
			  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR plant_id = $2)
			  AND result = 'denied'
			  AND checked_in_at >= $3 AND checked_in_at <= $4
		)
		SELECT
			COALESCE(NULLIF(TRIM(denial_reason), ''), 'Sin motivo especificado') AS reason,
			COUNT(*) AS cnt,
			COALESCE(ROUND((COUNT(*)::numeric / NULLIF((SELECT total FROM total_denied), 0)::numeric) * 100, 1), 0) AS pct
		FROM access_logs
		WHERE tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR plant_id = $2)
		  AND result = 'denied'
		  AND checked_in_at >= $3 AND checked_in_at <= $4
		GROUP BY reason
		ORDER BY cnt DESC
		LIMIT $5
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query rejection reasons: %w", err)
	}
	defer rows.Close()

	var list []repository.RejectionReasonSummary
	for rows.Next() {
		var s repository.RejectionReasonSummary
		if err := rows.Scan(&s.Reason, &s.Count, &s.Percentage); err != nil {
			return nil, fmt.Errorf("postgres/report: failed to scan rejection reason: %w", err)
		}
		list = append(list, s)
	}

	return list, nil
}

func (r *reportRepository) GetContractorComplianceRanking(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) ([]repository.ContractorComplianceSummary, error) {
	query := `
		SELECT
			c.id,
			c.name,
			COUNT(al.id) AS total_attempts,
			COUNT(al.id) FILTER (WHERE al.result IN ('granted', 'granted_with_warning')) AS granted_count,
			COUNT(al.id) FILTER (WHERE al.result = 'denied') AS denied_count
		FROM access_logs al
		JOIN contractors c ON al.contractor_id = c.id
		WHERE al.tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR al.plant_id = $2)
		  AND al.checked_in_at >= $3 AND al.checked_in_at <= $4
		GROUP BY c.id, c.name
		ORDER BY total_attempts DESC
		LIMIT 15
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query contractor compliance: %w", err)
	}
	defer rows.Close()

	var list []repository.ContractorComplianceSummary
	for rows.Next() {
		var s repository.ContractorComplianceSummary
		if err := rows.Scan(&s.ContractorID, &s.ContractorName, &s.TotalAttempts, &s.GrantedCount, &s.DeniedCount); err != nil {
			return nil, fmt.Errorf("postgres/report: failed to scan contractor ranking: %w", err)
		}
		if s.TotalAttempts > 0 {
			s.ComplianceRate = math.Round((float64(s.GrantedCount)/float64(s.TotalAttempts))*1000) / 10
		}
		list = append(list, s)
	}

	return list, nil
}

func (r *reportRepository) ListAuditAccessRecords(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
	from, to time.Time,
) ([]repository.AuditAccessRecord, error) {
	query := `
		SELECT
			al.id,
			al.checked_in_at,
			al.checked_out_at,
			CASE WHEN al.checked_out_at IS NOT NULL
			     THEN ROUND(EXTRACT(EPOCH FROM (al.checked_out_at - al.checked_in_at)) / 60)::int
			     ELSE NULL END AS duration_minutes,
			p.name AS plant_name,
			ap.name AS access_point_name,
			u.full_name AS worker_name,
			u.email AS worker_email,
			cw.curp AS worker_curp,
			cw.imss_number AS worker_imss,
			cw.job_title AS worker_job_title,
			COALESCE(c.name, 'Sin Contratista') AS contractor_name,
			c.rfc AS contractor_rfc,
			c.repse_number AS contractor_repse,
			al.result::text,
			al.denial_reason,
			al.notes
		FROM access_logs al
		JOIN plants p ON al.plant_id = p.id
		JOIN access_points ap ON al.access_point_id = ap.id
		JOIN users u ON al.user_id = u.id
		LEFT JOIN contractors c ON al.contractor_id = c.id
		LEFT JOIN contractor_workers cw ON (cw.contractor_id = al.contractor_id AND cw.user_id = al.user_id)
		WHERE al.tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR al.plant_id = $2)
		  AND al.checked_in_at >= $3 AND al.checked_in_at <= $4
		ORDER BY al.checked_in_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID, from, to)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query audit access records: %w", err)
	}
	defer rows.Close()

	var list []repository.AuditAccessRecord
	for rows.Next() {
		var rec repository.AuditAccessRecord
		err := rows.Scan(
			&rec.AccessLogID,
			&rec.CheckedInAt,
			&rec.CheckedOutAt,
			&rec.DurationMinutes,
			&rec.PlantName,
			&rec.AccessPointName,
			&rec.WorkerName,
			&rec.WorkerEmail,
			&rec.WorkerCURP,
			&rec.WorkerIMSS,
			&rec.WorkerJobTitle,
			&rec.ContractorName,
			&rec.ContractorRFC,
			&rec.ContractorREPSE,
			&rec.Result,
			&rec.DenialReason,
			&rec.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/report: failed to scan audit record: %w", err)
		}
		list = append(list, rec)
	}

	return list, nil
}

func (r *reportRepository) ListAuditComplianceRecords(
	ctx context.Context,
	tenantID, plantID uuid.UUID,
) ([]repository.AuditComplianceRecord, error) {
	query := `
		SELECT
			u.id AS worker_id,
			u.full_name,
			u.email,
			cw.curp,
			cw.imss_number,
			cw.job_title,
			c.name AS contractor_name,
			c.rfc AS contractor_rfc,
			c.repse_number AS contractor_repse,
			COALESCE((SELECT d.status::text FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'DC3' ORDER BY d.created_at DESC LIMIT 1), 'missing') AS dc3_status,
			(SELECT TO_CHAR(d.expires_at, 'YYYY-MM-DD') FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'DC3' ORDER BY d.created_at DESC LIMIT 1) AS dc3_expires_at,
			COALESCE((SELECT d.status::text FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE (d.user_id = u.id OR d.contractor_id = c.id) AND dt.code = 'SUA' ORDER BY d.created_at DESC LIMIT 1), 'missing') AS sua_status,
			(SELECT TO_CHAR(d.expires_at, 'YYYY-MM-DD') FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE (d.user_id = u.id OR d.contractor_id = c.id) AND dt.code = 'SUA' ORDER BY d.created_at DESC LIMIT 1) AS sua_expires_at,
			COALESCE((SELECT d.status::text FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'MED' ORDER BY d.created_at DESC LIMIT 1), 'missing') AS med_status,
			(SELECT TO_CHAR(d.expires_at, 'YYYY-MM-DD') FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'MED' ORDER BY d.created_at DESC LIMIT 1) AS med_expires_at,
			COALESCE((SELECT d.status::text FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'INE' ORDER BY d.created_at DESC LIMIT 1), 'missing') AS ine_status,
			(SELECT TO_CHAR(d.expires_at, 'YYYY-MM-DD') FROM documents d JOIN document_types dt ON d.document_type_id = dt.id WHERE d.user_id = u.id AND dt.code = 'INE' ORDER BY d.created_at DESC LIMIT 1) AS ine_expires_at
		FROM contractor_workers cw
		JOIN users u ON cw.user_id = u.id
		JOIN contractors c ON cw.contractor_id = c.id
		JOIN plant_contractors pc ON pc.contractor_id = c.id
		WHERE pc.tenant_id = $1
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR pc.plant_id = $2)
		  AND cw.is_active = TRUE
		ORDER BY c.name ASC, u.full_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, plantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/report: failed to query compliance records: %w", err)
	}
	defer rows.Close()

	var list []repository.AuditComplianceRecord
	for rows.Next() {
		var rec repository.AuditComplianceRecord
		err := rows.Scan(
			&rec.WorkerID,
			&rec.WorkerName,
			&rec.WorkerEmail,
			&rec.WorkerCURP,
			&rec.WorkerIMSS,
			&rec.WorkerJobTitle,
			&rec.ContractorName,
			&rec.ContractorRFC,
			&rec.ContractorREPSE,
			&rec.DC3Status,
			&rec.DC3ExpiresAt,
			&rec.SUAStatus,
			&rec.SUAExpiresAt,
			&rec.MedicalStatus,
			&rec.MedicalExpiresAt,
			&rec.INEStatus,
			&rec.INEExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/report: failed to scan compliance record: %w", err)
		}

		// Calculate overall compliance
		if rec.DC3Status == "approved" && (rec.SUAStatus == "approved" || rec.SUAStatus == "pending_review") && rec.MedicalStatus == "approved" && rec.INEStatus == "approved" {
			rec.OverallCompliance = "APTO"
		} else if rec.DC3Status == "rejected" || rec.MedicalStatus == "rejected" || rec.INEStatus == "rejected" {
			rec.OverallCompliance = "NO_APTO"
		} else {
			rec.OverallCompliance = "EN_REVISION"
		}

		list = append(list, rec)
	}

	return list, nil
}
