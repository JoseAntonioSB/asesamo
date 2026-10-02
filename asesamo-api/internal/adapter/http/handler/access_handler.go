package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
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
	plantIDStr := strings.TrimSpace(query.Get("plant_id"))
	var plantUUID uuid.UUID
	if plantIDStr != "" {
		parsed, err := uuid.Parse(plantIDStr)
		if err != nil {
			dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El formato de plant_id no es válido."))
			return
		}
		plantUUID = parsed
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
		TenantID: tenantID,
		PlantID:  plantUUID,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteJSON(w, http.StatusOK, dto.SuccessResponse{
		Data: out.Logs,
		Meta: map[string]interface{}{
			"page":     out.Page,
			"per_page": out.PerPage,
			"total":    out.Total,
		},
	})
}
