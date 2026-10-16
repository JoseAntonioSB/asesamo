package handler

import (
	"encoding/json"
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/contractor"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ContractorHandler coordinates HTTP requests for contractor companies and crews.
type ContractorHandler struct {
	inviteContractorUseCase   *contractor.InviteContractorUseCase
	listContractorsUseCase     *contractor.ListPlantContractorsUseCase
	updateStatusUseCase        *contractor.UpdateContractorStatusUseCase
	getMyCompanyUseCase        *contractor.GetMyContractorCompanyUseCase
	listWorkersUseCase         *contractor.ListContractorWorkersUseCase
	addWorkerUseCase           *contractor.AddContractorWorkerUseCase
	toggleWorkerStatusUseCase  *contractor.ToggleContractorWorkerStatusUseCase
}

// NewContractorHandler creates a new ContractorHandler instance.
func NewContractorHandler(
	inviteContractorUseCase *contractor.InviteContractorUseCase,
	listContractorsUseCase *contractor.ListPlantContractorsUseCase,
	updateStatusUseCase *contractor.UpdateContractorStatusUseCase,
	getMyCompanyUseCase *contractor.GetMyContractorCompanyUseCase,
	listWorkersUseCase *contractor.ListContractorWorkersUseCase,
	addWorkerUseCase *contractor.AddContractorWorkerUseCase,
	toggleWorkerStatusUseCase *contractor.ToggleContractorWorkerStatusUseCase,
) *ContractorHandler {
	return &ContractorHandler{
		inviteContractorUseCase:   inviteContractorUseCase,
		listContractorsUseCase:     listContractorsUseCase,
		updateStatusUseCase:        updateStatusUseCase,
		getMyCompanyUseCase:        getMyCompanyUseCase,
		listWorkersUseCase:         listWorkersUseCase,
		addWorkerUseCase:           addWorkerUseCase,
		toggleWorkerStatusUseCase:  toggleWorkerStatusUseCase,
	}
}

// InviteContractor handles POST /api/v1/contractors/invite.
func (h *ContractorHandler) InviteContractor(w http.ResponseWriter, r *http.Request) {
	callerID, _ := tenant.UserIDFromContext(r.Context())
	tenantID, _ := tenant.FromContext(r.Context())

	var req dto.InviteContractorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Payload JSON malformado."))
		return
	}

	var invitedByID *uuid.UUID
	if callerID != uuid.Nil {
		invitedByID = &callerID
	}

	res, err := h.inviteContractorUseCase.Execute(r.Context(), contractor.InviteContractorInput{
		TenantID:       tenantID,
		PlantID:        req.PlantID,
		InvitedByID:    invitedByID,
		Name:           req.Name,
		RFC:            req.RFC,
		REPSENumber:    req.REPSENumber,
		ContractorType: req.ContractorType,
		AdminEmail:     req.AdminEmail,
		AdminFullName:  req.AdminFullName,
		AdminPhone:     req.AdminPhone,
		InitialStatus:  req.InitialStatus,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusCreated, dto.ToContractorFromEntity(res))
}

// ListPlantContractors handles GET /api/v1/contractors.
func (h *ContractorHandler) ListPlantContractors(w http.ResponseWriter, r *http.Request) {
	plantIDStr := r.URL.Query().Get("plant_id")
	if plantIDStr == "" {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El parámetro plant_id es obligatorio."))
		return
	}

	plantID, err := uuid.Parse(plantIDStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El parámetro plant_id no es un UUID válido."))
		return
	}

	list, err := h.listContractorsUseCase.Execute(r.Context(), plantID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	resp := make([]dto.ContractorResponse, len(list))
	for i, c := range list {
		resp[i] = dto.ToContractorResponse(c)
	}

	dto.WriteSuccess(w, http.StatusOK, resp)
}

// UpdateContractorStatus handles PATCH /api/v1/contractors/{id}/status.
func (h *ContractorHandler) UpdateContractorStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	contractorID, err := uuid.Parse(idStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El identificador del contratista no es válido."))
		return
	}

	var req dto.UpdateContractorStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Payload JSON malformado."))
		return
	}

	if err := h.updateStatusUseCase.Execute(r.Context(), req.PlantID, contractorID, req.Status); err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"contractor_id": contractorID,
		"plant_id":      req.PlantID,
		"status":        req.Status,
		"message":       "Estatus del contratista actualizado correctamente.",
	})
}

// GetMyCompany handles GET /api/v1/contractors/my-company.
func (h *ContractorHandler) GetMyCompany(w http.ResponseWriter, r *http.Request) {
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	c, err := h.getMyCompanyUseCase.Execute(r.Context(), userID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, dto.ToContractorFromEntity(c))
}

// ListWorkers handles GET /api/v1/contractors/workers.
func (h *ContractorHandler) ListWorkers(w http.ResponseWriter, r *http.Request) {
	contractorID, err := h.resolveContractorID(r)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	workers, err := h.listWorkersUseCase.Execute(r.Context(), contractorID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	resp := make([]dto.ContractorWorkerResponse, len(workers))
	for i, wItem := range workers {
		resp[i] = dto.ToContractorWorkerResponse(wItem)
	}

	dto.WriteSuccess(w, http.StatusOK, resp)
}

// AddWorker handles POST /api/v1/contractors/workers.
func (h *ContractorHandler) AddWorker(w http.ResponseWriter, r *http.Request) {
	var req dto.AddWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Payload JSON malformado."))
		return
	}

	var contractorID uuid.UUID
	if req.ContractorID != nil && *req.ContractorID != uuid.Nil {
		contractorID = *req.ContractorID
	} else {
		var err error
		contractorID, err = h.resolveContractorID(r)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
	}

	tenantID, _ := tenant.FromContext(r.Context())
	workerDetail, err := h.addWorkerUseCase.Execute(r.Context(), contractor.AddContractorWorkerInput{
		ContractorID: contractorID,
		TenantID:     tenantID,
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		JobTitle:     req.JobTitle,
		CURP:         req.CURP,
		IMSSNumber:   req.IMSSNumber,
	})
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusCreated, dto.ToContractorWorkerResponse(*workerDetail))
}

// ToggleWorkerStatus handles PATCH /api/v1/contractors/workers/{id}/status.
func (h *ContractorHandler) ToggleWorkerStatus(w http.ResponseWriter, r *http.Request) {
	workerIDStr := chi.URLParam(r, "id")
	workerID, err := uuid.Parse(workerIDStr)
	if err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("El identificador del trabajador no es un UUID válido."))
		return
	}

	var req dto.ToggleWorkerStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.WriteError(w, domainErrors.ErrInvalidInput.WithMessage("Payload JSON malformado."))
		return
	}

	var contractorID uuid.UUID
	if req.ContractorID != nil && *req.ContractorID != uuid.Nil {
		contractorID = *req.ContractorID
	} else {
		var err error
		contractorID, err = h.resolveContractorID(r)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
	}

	if err := h.toggleWorkerStatusUseCase.Execute(r.Context(), contractorID, workerID, req.IsActive); err != nil {
		dto.WriteError(w, err)
		return
	}

	dto.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"worker_id":  workerID,
		"is_active":  req.IsActive,
		"message":    "Estatus del trabajador actualizado correctamente.",
	})
}

func (h *ContractorHandler) resolveContractorID(r *http.Request) (uuid.UUID, error) {
	queryID := r.URL.Query().Get("contractor_id")
	if queryID != "" {
		id, err := uuid.Parse(queryID)
		if err == nil && id != uuid.Nil {
			return id, nil
		}
	}

	// Try resolving via logged-in contractor_admin
	userID, err := tenant.UserIDFromContext(r.Context())
	if err != nil {
		return uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("El identificador de la empresa contratista es obligatorio.")
	}

	c, err := h.getMyCompanyUseCase.Execute(r.Context(), userID)
	if err != nil {
		return uuid.Nil, domainErrors.ErrInvalidInput.WithMessage("No se encontró la empresa contratista asociada a su usuario.")
	}

	return c.ID, nil
}
