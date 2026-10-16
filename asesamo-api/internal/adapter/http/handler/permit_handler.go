package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/permit"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PermitHandler handles HTTP requests for High-Risk Work Permits (PTAR Digital).
type PermitHandler struct {
	createPermitUC     *permit.CreatePermitUseCase
	listPermitsUC      *permit.ListPermitsUseCase
	getPermitUC        *permit.GetPermitUseCase
	assignWorkersUC    *permit.AssignWorkersUseCase
	updateChecklistUC  *permit.UpdateChecklistUseCase
	approvePermitUC    *permit.ApprovePermitUseCase
	transitionStatusUC *permit.TransitionStatusUseCase
}

// NewPermitHandler constructs a new PermitHandler with injected use cases.
func NewPermitHandler(
	createPermitUC *permit.CreatePermitUseCase,
	listPermitsUC *permit.ListPermitsUseCase,
	getPermitUC *permit.GetPermitUseCase,
	assignWorkersUC *permit.AssignWorkersUseCase,
	updateChecklistUC *permit.UpdateChecklistUseCase,
	approvePermitUC *permit.ApprovePermitUseCase,
	transitionStatusUC *permit.TransitionStatusUseCase,
) *PermitHandler {
	return &PermitHandler{
		createPermitUC:     createPermitUC,
		listPermitsUC:      listPermitsUC,
		getPermitUC:        getPermitUC,
		assignWorkersUC:    assignWorkersUC,
		updateChecklistUC:  updateChecklistUC,
		approvePermitUC:    approvePermitUC,
		transitionStatusUC: transitionStatusUC,
	}
}

// CreatePermit handles POST /api/v1/permits.
func (h *PermitHandler) CreatePermit(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePermitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON inválido."))
		return
	}

	res, err := h.createPermitUC.Execute(r.Context(), permit.CreatePermitInput{
		PlantID:           req.PlantID,
		ContractorID:      req.ContractorID,
		PermitType:        entity.PermitType(req.PermitType),
		WorkArea:          req.WorkArea,
		Description:       req.Description,
		HazardDescription: req.HazardDescription,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		RequesterNotes:    req.Notes,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusCreated, res)
}

// ListPermits handles GET /api/v1/permits.
func (h *PermitHandler) ListPermits(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var filter repository.PermitFilter

	if plantIDStr := q.Get("plant_id"); plantIDStr != "" {
		if pid, err := uuid.Parse(plantIDStr); err == nil {
			filter.PlantID = &pid
		}
	}

	if contractorIDStr := q.Get("contractor_id"); contractorIDStr != "" {
		if cid, err := uuid.Parse(contractorIDStr); err == nil {
			filter.ContractorID = &cid
		}
	}

	if statusStr := q.Get("status"); statusStr != "" {
		st := entity.PermitStatus(statusStr)
		filter.Status = &st
	}

	if typeStr := q.Get("permit_type"); typeStr != "" {
		pt := entity.PermitType(typeStr)
		filter.PermitType = &pt
	}

	if startStr := q.Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartDate = &t
		} else if t, err := time.Parse("2006-01-02", startStr); err == nil {
			filter.StartDate = &t
		}
	}

	if endStr := q.Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndDate = &t
		} else if t, err := time.Parse("2006-01-02", endStr); err == nil {
			filter.EndDate = &t
		}
	}

	page := 1
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	filter.Page = page

	perPage := 20
	if perPageStr := q.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}
	filter.PerPage = perPage

	permits, total, err := h.listPermitsUC.Execute(r.Context(), filter)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	if permits == nil {
		permits = []entity.WorkPermit{}
	}

	dto.WriteSuccessWithMeta(w, http.StatusOK, permits, dto.PaginationMeta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// GetPermit handles GET /api/v1/permits/{id}.
func (h *PermitHandler) GetPermit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de permiso inválido."))
		return
	}

	res, err := h.getPermitUC.Execute(r.Context(), id)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// AssignWorkers handles POST /api/v1/permits/{id}/workers.
func (h *PermitHandler) AssignWorkers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de permiso inválido."))
		return
	}

	var req dto.AssignWorkersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON inválido."))
		return
	}

	res, err := h.assignWorkersUC.Execute(r.Context(), permit.AssignWorkersInput{
		PermitID:  id,
		WorkerIDs: req.WorkerIDs,
		Action:    req.Action,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// UpdateChecklist handles PATCH /api/v1/permits/{id}/checklist.
func (h *PermitHandler) UpdateChecklist(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de permiso inválido."))
		return
	}

	var req dto.UpdateChecklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON inválido."))
		return
	}

	var items []permit.ChecklistItemUpdate
	for _, it := range req.Items {
		items = append(items, permit.ChecklistItemUpdate{
			ItemKey:      it.ItemKey,
			ItemCode:     it.ItemCode,
			IsVerified:   it.IsVerified,
			IsChecked:    it.IsChecked,
			Notes:        it.Notes,
			Measurements: it.Measurements,
		})
	}

	res, err := h.updateChecklistUC.Execute(r.Context(), permit.UpdateChecklistInput{
		PermitID: id,
		Items:    items,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// ApprovePermit handles POST /api/v1/permits/{id}/approve.
func (h *PermitHandler) ApprovePermit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de permiso inválido."))
		return
	}

	var req dto.ApprovePermitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON inválido."))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	res, err := h.approvePermitUC.Execute(r.Context(), permit.ApprovePermitInput{
		PermitID:      id,
		Notes:         req.Notes,
		SignatureData: req.SignatureData,
		IPAddress:     &ip,
		UserAgent:     &ua,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}

// TransitionStatus handles POST /api/v1/permits/{id}/status.
func (h *PermitHandler) TransitionStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("ID de permiso inválido."))
		return
	}

	var req dto.TransitionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Cuerpo de solicitud JSON inválido."))
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	res, err := h.transitionStatusUC.Execute(r.Context(), permit.TransitionStatusInput{
		PermitID:      id,
		Action:        req.Action,
		Reason:        req.Reason,
		Notes:         req.Notes,
		SignatureData: req.SignatureData,
		IPAddress:     &ip,
		UserAgent:     &ua,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, res)
}
