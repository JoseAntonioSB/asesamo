package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"time"
)

// AccessHandler handles access verification, check-out, and audit log endpoints.
type AccessHandler struct {
	checkInUseCase  *access.CheckInUseCase
	checkOutUseCase *access.CheckOutUseCase
	listLogsUseCase *access.ListAccessLogsUseCase
}

// NewAccessHandler creates an AccessHandler.
func NewAccessHandler(
	checkInUC *access.CheckInUseCase,
	checkOutUC *access.CheckOutUseCase,
	listLogsUC *access.ListAccessLogsUseCase,
) *AccessHandler {
	return &AccessHandler{
		checkInUseCase:  checkInUC,
		checkOutUseCase: checkOutUC,
		listLogsUseCase: listLogsUC,
	}
}

// CheckIn handles POST /api/v1/access/check-in.
func (h *AccessHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	guardUserID, _ := tenant.UserIDFromContext(r.Context())

	var req dto.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El cuerpo de la solicitud no es un JSON válido o está vacío."))
		return
	}

	workerUUID, apUUID, err := req.Validate()
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	var guardIDPtr *uuid.UUID
	if guardUserID != uuid.Nil {
		guardIDPtr = &guardUserID
	}

	out, err := h.checkInUseCase.Execute(r.Context(), access.CheckInInput{
		TenantID:      tenantID,
		WorkerUserID:  workerUUID,
		AccessPointID: apUUID,
		ProcessedByID: guardIDPtr,
		Notes:         req.Notes,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, out)
}

// CheckOut handles POST /api/v1/access/check-out.
func (h *AccessHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	var req dto.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El cuerpo de la solicitud no es un JSON válido o está vacío."))
		return
	}

	workerUUID, plantUUID, apUUID, err := req.Validate()
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	out, err := h.checkOutUseCase.Execute(r.Context(), access.CheckOutInput{
		TenantID:      tenantID,
		PlantID:       plantUUID,
		AccessPointID: apUUID,
		WorkerUserID:  workerUUID,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, out)
}

// ListLogs handles GET /api/v1/access-logs.
func (h *AccessHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	query := r.URL.Query()

	var plantUUID uuid.UUID
	if plantIDStr := strings.TrimSpace(query.Get("plant_id")); plantIDStr != "" {
		parsed, err := uuid.Parse(plantIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de plant_id no es válido."))
			return
		}
		plantUUID = parsed
	}

	var apUUID uuid.UUID
	if apIDStr := strings.TrimSpace(query.Get("access_point_id")); apIDStr != "" {
		parsed, err := uuid.Parse(apIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de access_point_id no es válido."))
			return
		}
		apUUID = parsed
	}

	var workerUUID uuid.UUID
	workerIDStr := strings.TrimSpace(query.Get("worker_user_id"))
	if workerIDStr == "" {
		workerIDStr = strings.TrimSpace(query.Get("worker_id"))
	}
	if workerIDStr != "" {
		parsed, err := uuid.Parse(workerIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de worker_user_id no es válido."))
			return
		}
		workerUUID = parsed
	}

	var resultPtr *entity.AccessResult
	if resultStr := strings.TrimSpace(query.Get("result")); resultStr != "" {
		res := entity.AccessResult(resultStr)
		if res != entity.AccessResultGranted && res != entity.AccessResultGrantedWithWarning && res != entity.AccessResultDenied {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El parámetro result debe ser granted, granted_with_warning o denied."))
			return
		}
		resultPtr = &res
	}

	fromTime, err := parseDateTime(query.Get("from"), false)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	toTime, err := parseDateTime(query.Get("to"), true)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 20
	if perPageStr := query.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}

	out, err := h.listLogsUseCase.Execute(r.Context(), access.ListAccessLogsInput{
		TenantID:      tenantID,
		PlantID:       plantUUID,
		AccessPointID: apUUID,
		WorkerUserID:  workerUUID,
		Result:        resultPtr,
		From:          fromTime,
		To:            toTime,
		Page:          page,
		PerPage:       perPage,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccessWithMeta(w, http.StatusOK, out.Logs, dto.PaginationMeta{
		Page:    out.Page,
		PerPage: out.PerPage,
		Total:   out.Total,
	})
}

// parseDateTime parses date strings in RFC3339 or YYYY-MM-DD format.
func parseDateTime(val string, isEndOfDay bool) (*time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", val); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		var utc time.Time
		if isEndOfDay {
			utc = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
		} else {
			utc = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		}
		return &utc, nil
	}
	return nil, domainErrors.ErrInvalidInput.WithMessage("El formato de fecha no es válido. Use YYYY-MM-DD o RFC3339.")
}
