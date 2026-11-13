package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/epp"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
)

type EPPHandler struct {
	repo      repository.EPPRepository
	deliverUC *usecase.DeliverEPPUseCase
}

func NewEPPHandler(repo repository.EPPRepository, deliverUC *usecase.DeliverEPPUseCase) *EPPHandler {
	return &EPPHandler{
		repo:      repo,
		deliverUC: deliverUC,
	}
}

func (h *EPPHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	items, err := h.repo.ListItemCatalog(r.Context(), tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": items,
	})
}

type CreateEPPItemRequestBody struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	Category          string `json:"category"`
	NormativeStandard string `json:"normative_standard,omitempty"`
	LifespanDays      int    `json:"lifespan_days"`
	StockQuantity     int    `json:"stock_quantity"`
}

func (h *EPPHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	var req CreateEPPItemRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	item := &entity.EPPItem{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Code:              req.Code,
		Name:              req.Name,
		Category:          entity.EPPCategory(req.Category),
		NormativeStandard: req.NormativeStandard,
		LifespanDays:      req.LifespanDays,
		StockQuantity:     req.StockQuantity,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := h.repo.CreateItem(r.Context(), item); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": item,
	})
}

type DeliverEPPRequestBody struct {
	EPPItemID     string    `json:"epp_item_id"`
	WorkerID      string    `json:"worker_id"`
	ContractorID  string    `json:"contractor_id,omitempty"`
	DeliveryDate  time.Time `json:"delivery_date"`
	SignatureData string    `json:"signature_data,omitempty"`
	Notes         string    `json:"notes,omitempty"`
}

func (h *EPPHandler) DeliverEPP(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}
	userID, _ := tenant.UserIDFromContext(r.Context())

	var req DeliverEPPRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	eppID, err := uuid.Parse(req.EPPItemID)
	if err != nil {
		http.Error(w, `{"error":"ID de artículo EPP inválido"}`, http.StatusBadRequest)
		return
	}
	workerID, err := uuid.Parse(req.WorkerID)
	if err != nil {
		http.Error(w, `{"error":"ID de trabajador inválido"}`, http.StatusBadRequest)
		return
	}

	input := usecase.DeliverEPPInput{
		EPPItemID:     eppID,
		WorkerID:      workerID,
		DeliveryDate:  req.DeliveryDate,
		SignatureData: req.SignatureData,
		Notes:         req.Notes,
	}

	if req.ContractorID != "" {
		if cid, err := uuid.Parse(req.ContractorID); err == nil {
			input.ContractorID = &cid
		}
	}

	delivery, err := h.deliverUC.Execute(r.Context(), tenantID, userID, input)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": delivery,
	})
}

func (h *EPPHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		http.Error(w, `{"error":"No se encontró información del tenant"}`, http.StatusUnauthorized)
		return
	}

	deliveries, total, err := h.repo.ListDeliveries(r.Context(), tenantID, repository.EPPDeliveryFilter{})
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  deliveries,
		"total": total,
	})
}
