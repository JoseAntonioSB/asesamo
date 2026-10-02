package handler

import (
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"
)

// AccessPointHandler handles access point catalog endpoints.
type AccessPointHandler struct {
	listUseCase *access.ListAccessPointsUseCase
}

// NewAccessPointHandler creates an AccessPointHandler.
func NewAccessPointHandler(listUseCase *access.ListAccessPointsUseCase) *AccessPointHandler {
	return &AccessPointHandler{
		listUseCase: listUseCase,
	}
}

// List handles GET /api/v1/access-points.
func (h *AccessPointHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		dto.WriteError(w, domainErrors.ErrUnauthorized)
		return
	}

	points, err := h.listUseCase.Execute(r.Context(), tenantID)
	if err != nil {
		dto.WriteError(w, err)
		return
	}

	resps := make([]dto.AccessPointResponse, 0, len(points))
	for _, p := range points {
		resps = append(resps, dto.AccessPointResponse{
			ID:          p.ID,
			PlantID:     p.PlantID,
			TenantID:    p.TenantID,
			Name:        p.Name,
			Description: p.Description,
			IsActive:    p.IsActive,
		})
	}

	dto.WriteSuccess(w, http.StatusOK, resps)
}
