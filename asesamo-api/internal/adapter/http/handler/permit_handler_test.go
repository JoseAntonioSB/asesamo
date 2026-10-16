package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/internal/core/usecase/permit"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPermitRepoForHandler struct {
	mock.Mock
}

func (m *mockPermitRepoForHandler) Create(ctx context.Context, p *entity.WorkPermit) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPermitRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.WorkPermit, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkPermit), args.Error(1)
}

func (m *mockPermitRepoForHandler) ListWithFilters(ctx context.Context, filter repository.PermitFilter) ([]entity.WorkPermit, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.WorkPermit), args.Int(1), args.Error(2)
}

func (m *mockPermitRepoForHandler) NextPermitNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	args := m.Called(ctx, tenantID, year)
	return args.String(0), args.Error(1)
}

func (m *mockPermitRepoForHandler) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PermitStatus, reason *string, byUserID *uuid.UUID) error {
	return m.Called(ctx, id, status, reason, byUserID).Error(0)
}

func (m *mockPermitRepoForHandler) AssignWorker(ctx context.Context, worker *entity.PermitWorker) error {
	return m.Called(ctx, worker).Error(0)
}

func (m *mockPermitRepoForHandler) RemoveWorker(ctx context.Context, permitID, workerID uuid.UUID) error {
	return m.Called(ctx, permitID, workerID).Error(0)
}

func (m *mockPermitRepoForHandler) ListWorkersByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitWorker, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitWorker), args.Error(1)
}

func (m *mockPermitRepoForHandler) InitChecklistItems(ctx context.Context, permitID uuid.UUID, items []entity.PermitChecklistItem) error {
	return m.Called(ctx, permitID, items).Error(0)
}

func (m *mockPermitRepoForHandler) ListChecklistByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitChecklistItem, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitChecklistItem), args.Error(1)
}

func (m *mockPermitRepoForHandler) UpdateChecklistItem(ctx context.Context, permitID uuid.UUID, itemKey string, isVerified bool, verifiedBy *uuid.UUID, notes *string, measurements map[string]any) error {
	return m.Called(ctx, permitID, itemKey, isVerified, verifiedBy, notes, measurements).Error(0)
}

func (m *mockPermitRepoForHandler) AddSignature(ctx context.Context, sig *entity.PermitSignature) error {
	return m.Called(ctx, sig).Error(0)
}

func (m *mockPermitRepoForHandler) ListSignaturesByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitSignature, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitSignature), args.Error(1)
}

type mockComplianceRepoForHandler struct {
	mock.Mock
}

func (m *mockComplianceRepoForHandler) GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error) {
	args := m.Called(ctx, plantID, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkerAffiliation), args.Error(1)
}

func (m *mockComplianceRepoForHandler) GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error) {
	args := m.Called(ctx, tenantID, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.RequiredDocumentRule), args.Error(1)
}

func (m *mockComplianceRepoForHandler) GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func (m *mockComplianceRepoForHandler) GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, contractorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func TestPermitHandler(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()
	contractorUserID := uuid.New()
	futureDate := time.Now().Add(30 * 24 * time.Hour)

	permitRepo := new(mockPermitRepoForHandler)
	compRepo := new(mockComplianceRepoForHandler)
	engine := service.NewComplianceEngine(compRepo)

	createUC := permit.NewCreatePermitUseCase(permitRepo)
	listUC := permit.NewListPermitsUseCase(permitRepo, nil)
	getUC := permit.NewGetPermitUseCase(permitRepo, nil)
	assignUC := permit.NewAssignWorkersUseCase(permitRepo, engine)
	checklistUC := permit.NewUpdateChecklistUseCase(permitRepo)
	approveUC := permit.NewApprovePermitUseCase(permitRepo, engine)
	statusUC := permit.NewTransitionStatusUseCase(permitRepo)

	h := handler.NewPermitHandler(
		createUC,
		listUC,
		getUC,
		assignUC,
		checklistUC,
		approveUC,
		statusUC,
	)

	r := chi.NewRouter()
	r.Post("/api/v1/permits", h.CreatePermit)
	r.Get("/api/v1/permits", h.ListPermits)
	r.Get("/api/v1/permits/{id}", h.GetPermit)
	r.Post("/api/v1/permits/{id}/workers", h.AssignWorkers)
	r.Patch("/api/v1/permits/{id}/checklist", h.UpdateChecklist)
	r.Post("/api/v1/permits/{id}/approve", h.ApprovePermit)
	r.Post("/api/v1/permits/{id}/status", h.TransitionStatus)

	t.Run("POST /api/v1/permits creates permit draft (201 Created)", func(t *testing.T) {
		permitRepo.On("NextPermitNumber", mock.Anything, tenantID, mock.Anything).Return("PTAR-2026-0001", nil).Once()
		permitRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
		permitRepo.On("InitChecklistItems", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		permitRepo.On("AddSignature", mock.Anything, mock.Anything).Return(nil).Once()

		createdPermit := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PermitNumber: "PTAR-2026-0001",
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		permitRepo.On("GetByID", mock.Anything, mock.Anything).Return(createdPermit, nil).Once()

		payload := map[string]any{
			"plant_id":      plantID.String(),
			"contractor_id": contractorID.String(),
			"permit_type":   "heights",
			"work_area":     "Nave Industrial 4",
			"description":   "Mantenimiento de luminarias",
			"start_time":    time.Now().Add(time.Hour).Format(time.RFC3339),
			"end_time":      time.Now().Add(4 * time.Hour).Format(time.RFC3339),
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/permits", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		data := resp["data"].(map[string]any)
		assert.Equal(t, "PTAR-2026-0001", data["permit_number"])
	})

	t.Run("GET /api/v1/permits/{id} returns permit details (200 OK)", func(t *testing.T) {
		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PermitNumber: "PTAR-2026-0001",
			Status:       entity.PermitStatusDraft,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/permits/"+permitID.String(), nil)
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST /api/v1/permits/{id}/workers blocks non-compliant worker with 422", func(t *testing.T) {
		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Trabajador Sin DC3",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		// Documents without DC-3
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

		payload := map[string]any{
			"worker_ids": []string{workerID.String()},
			"action":     "assign",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/workers", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "WORKER_MISSING_DC3", errObj["code"])
		assert.Contains(t, errObj["message"], "NOM-009")
	})

	t.Run("POST /api/v1/permits/{id}/approve forbidden for contractor_admin (403)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("POST /api/v1/permits/{id}/approve authorized for ehs_manager (200 OK)", func(t *testing.T) {
		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Valido",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

		permitRepo.On("AssignWorker", mock.Anything, mock.Anything).Return(nil).Once()
		permitRepo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &ehsUserID).Return(nil).Once()
		permitRepo.On("AddSignature", mock.Anything, mock.Anything).Return(nil).Once()

		approvedWP := &entity.WorkPermit{
			ID:     permitID,
			Status: entity.PermitStatusApprovedActive,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(approvedWP, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte(`{"notes":"Aprobado en sitio"}`)))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
