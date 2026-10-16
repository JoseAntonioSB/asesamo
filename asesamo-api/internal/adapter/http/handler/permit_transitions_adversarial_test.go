package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/internal/core/usecase/permit"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupTransitionsRouter creates an HTTP router with full permit routes wired for empirical challenge.
func setupTransitionsRouter(
	pRepo *mockPermitRepoForHandler,
	cRepo *mockComplianceRepoForHandler,
) http.Handler {
	engine := service.NewComplianceEngine(cRepo)
	createUC := permit.NewCreatePermitUseCase(pRepo)
	listUC := permit.NewListPermitsUseCase(pRepo, nil)
	getUC := permit.NewGetPermitUseCase(pRepo, nil)
	assignUC := permit.NewAssignWorkersUseCase(pRepo, engine)
	checklistUC := permit.NewUpdateChecklistUseCase(pRepo)
	approveUC := permit.NewApprovePermitUseCase(pRepo, engine)
	statusUC := permit.NewTransitionStatusUseCase(pRepo)

	h := handler.NewPermitHandler(
		createUC, listUC, getUC, assignUC, checklistUC, approveUC, statusUC,
	)

	r := chi.NewRouter()
	r.Post("/api/v1/permits", h.CreatePermit)
	r.Get("/api/v1/permits", h.ListPermits)
	r.Get("/api/v1/permits/{id}", h.GetPermit)
	r.Post("/api/v1/permits/{id}/workers", h.AssignWorkers)
	r.Patch("/api/v1/permits/{id}/checklist", h.UpdateChecklist)
	r.Post("/api/v1/permits/{id}/approve", h.ApprovePermit)
	r.Post("/api/v1/permits/{id}/status", h.TransitionStatus)
	return r
}

// -----------------------------------------------------------------------------
// 1. CHALLENGE: EHS approval with incomplete checklist items -> must fail with HTTP 422
// -----------------------------------------------------------------------------
func TestEmpiricalChallenge_ApproveWithIncompleteChecklist(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()

	t.Run("EHS approval with 1 mandatory item unverified fails with HTTP 422 CHECKLIST_INCOMPLETE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

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
				{ItemKey: "chk_2", IsMandatory: true, IsVerified: false}, // UNVERIFIED MANDATORY
				{ItemKey: "chk_3", IsMandatory: false, IsVerified: false},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Approval must fail with 422 when checklist is incomplete")

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj, ok := resp["error"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "CHECKLIST_INCOMPLETE", errObj["code"])
		assert.Contains(t, errObj["message"], "faltan verificaciones obligatorias")
	})

	t.Run("EHS approval with all mandatory verified but optional unverified SUCCEEDS (200 OK)", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)
		futureDate := time.Now().Add(30 * 24 * time.Hour)

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
				{ItemKey: "chk_2", IsMandatory: false, IsVerified: false}, // optional unverified is allowed
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID: workerID, FullName: "Juan Valido", UserIsActive: true, WorkerIsActive: true,
			ContractorID: contractorID, ContractorIsActive: true, PlantStatus: "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

		pRepo.On("AssignWorker", mock.Anything, mock.Anything).Return(nil).Once()
		pRepo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &ehsUserID).Return(nil).Once()
		pRepo.On("AddSignature", mock.Anything, mock.Anything).Return(nil).Once()

		approvedWP := &entity.WorkPermit{ID: permitID, Status: entity.PermitStatusApprovedActive}
		pRepo.On("GetByID", mock.Anything, permitID).Return(approvedWP, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// -----------------------------------------------------------------------------
// 2. CHALLENGE: Attempt EHS approval without assigned workers -> must fail with HTTP 422
// -----------------------------------------------------------------------------
func TestEmpiricalChallenge_ApproveWithoutAssignedWorkers(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	ehsUserID := uuid.New()

	t.Run("EHS approval with empty workers slice fails with HTTP 422 PERMIT_NO_WORKERS", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{}, // EMPTY CREW
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Approval must fail with 422 when workers are missing")

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj, ok := resp["error"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "PERMIT_NO_WORKERS", errObj["code"])
		assert.Contains(t, errObj["message"], "al menos un trabajador asignado")
	})

	t.Run("EHS approval with nil workers slice fails with HTTP 422 PERMIT_NO_WORKERS", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      nil, // NIL CREW
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_NO_WORKERS", errObj["code"])
	})
}

// -----------------------------------------------------------------------------
// 3. CHALLENGE: Illegal State Transitions
//    - draft directly to approved_active
//    - editing closed permits (workers, checklist, approve, status transitions)
//    - other illegal state machine paths
// -----------------------------------------------------------------------------
func TestEmpiricalChallenge_IllegalStateTransitions(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()
	contractorUserID := uuid.New()

	// -------------------------------------------------------------------------
	// A. Illegal transition: draft directly to approved_active
	// -------------------------------------------------------------------------
	t.Run("Transition /status with action=approved_active on draft is rejected (400 Bad Request)", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		body := []byte(`{"action": "approved_active"}`)
		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "INVALID_INPUT", errObj["code"])
		assert.Contains(t, errObj["message"], "Acción de transición de estado desconocida")
	})

	t.Run("Transition /status with action=resume on draft is rejected with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		body := []byte(`{"action": "resume"}`)
		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Transition /status with action=close on draft is rejected with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		body := []byte(`{"action": "close"}`)
		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Transition /status with action=suspend on draft is rejected with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		body := []byte(`{"action": "suspend", "reason": "Motivo de prueba"}`)
		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Empirical probe: EHS approval on DRAFT permit via /approve endpoint", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)
		futureDate := time.Now().Add(30 * 24 * time.Hour)

		// Draft permit with complete checklist and assigned worker
		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft, // DRAFT STATUS
			Workers:      []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID: workerID, FullName: "Juan Valido", UserIsActive: true, WorkerIsActive: true,
			ContractorID: contractorID, ContractorIsActive: true, PlantStatus: "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		pRepo.On("AssignWorker", mock.Anything, mock.Anything).Return(nil)
		pRepo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &ehsUserID).Return(nil)
		pRepo.On("AddSignature", mock.Anything, mock.Anything).Return(nil)

		approvedWP := &entity.WorkPermit{ID: permitID, Status: entity.PermitStatusApprovedActive}
		pRepo.On("GetByID", mock.Anything, permitID).Return(approvedWP, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Empirical result of POST /approve on DRAFT permit: HTTP status code = %d", w.Code)
	})

	// -------------------------------------------------------------------------
	// B. Editing CLOSED permits must fail with HTTP 422
	// -------------------------------------------------------------------------
	t.Run("Attempt to assign workers to CLOSED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			Status:       entity.PermitStatusClosed, // CLOSED
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		payload := map[string]any{
			"worker_ids": []string{workerID.String()},
			"action":     "assign",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/workers", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Modifying workers on closed permit must fail with 422")
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
		assert.Contains(t, errObj["message"], "No se puede modificar la cuadrilla")
	})

	t.Run("Attempt to remove workers from CLOSED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			Status:       entity.PermitStatusClosed,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		payload := map[string]any{
			"worker_ids": []string{workerID.String()},
			"action":     "remove",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/workers", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), contractorUserID), "contractor_admin")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Attempt to update checklist on CLOSED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusClosed, // CLOSED is terminal
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		payload := map[string]any{
			"items": []map[string]any{
				{"item_key": "chk_1", "is_verified": true},
			},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("PATCH", "/api/v1/permits/"+permitID.String()+"/checklist", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Updating checklist on closed permit must fail with 422")
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
		assert.Contains(t, errObj["message"], "No se puede modificar la lista de verificación")
	})

	t.Run("Attempt to update checklist on CANCELLED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusCancelled, // CANCELLED is terminal
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		payload := map[string]any{
			"items": []map[string]any{
				{"item_key": "chk_1", "is_verified": true},
			},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("PATCH", "/api/v1/permits/"+permitID.String()+"/checklist", bytes.NewReader(body))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Attempt to approve a CLOSED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusClosed,
			Workers:  []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
		assert.Contains(t, errObj["message"], "El permiso ya ha sido procesado")
	})

	t.Run("Attempt to approve a CANCELLED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusCancelled,
			Workers:  []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	t.Run("Attempt to transition /status on CLOSED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		actions := []string{"resume", "close", "cancel", "submit", "suspend"}
		for _, act := range actions {
			pRepo := new(mockPermitRepoForHandler)
			cRepo := new(mockComplianceRepoForHandler)
			router := setupTransitionsRouter(pRepo, cRepo)

			wp := &entity.WorkPermit{
				ID:       permitID,
				TenantID: tenantID,
				Status:   entity.PermitStatusClosed,
			}
			pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

			body := []byte(`{"action": "` + act + `", "reason": "test"}`)
			req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
			ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Action %s on closed permit must return 422", act)
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			errObj := resp["error"].(map[string]any)
			assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
		}
	})

	t.Run("Attempt to transition /status on CANCELLED permit fails with HTTP 422 PERMIT_INVALID_STATE", func(t *testing.T) {
		actions := []string{"resume", "close", "cancel", "submit", "suspend"}
		for _, act := range actions {
			pRepo := new(mockPermitRepoForHandler)
			cRepo := new(mockComplianceRepoForHandler)
			router := setupTransitionsRouter(pRepo, cRepo)

			wp := &entity.WorkPermit{
				ID:       permitID,
				TenantID: tenantID,
				Status:   entity.PermitStatusCancelled,
			}
			pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

			body := []byte(`{"action": "` + act + `", "reason": "test"}`)
			req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/status", bytes.NewReader(body))
			ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Action %s on cancelled permit must return 422", act)
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			errObj := resp["error"].(map[string]any)
			assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
		}
	})
}

// -----------------------------------------------------------------------------
// 4. CHALLENGE: Tenant Isolation Fail-Closed Verification
// -----------------------------------------------------------------------------
func TestEmpiricalChallenge_TenantIsolation_Security(t *testing.T) {
	tenantA := uuid.New()
	tenantB := uuid.New()
	permitID := uuid.New()
	userID := uuid.New()

	t.Run("Handler request without tenant context returns 401 or 404 fail-closed", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		pRepo.On("GetByID", mock.Anything, permitID).Return(nil, domainErrors.ErrTenantNotFound).Once()

		req := httptest.NewRequest("GET", "/api/v1/permits/"+permitID.String(), nil)
		// Missing tenant context entirely

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "TENANT_NOT_FOUND", errObj["code"])
	})

	t.Run("Cross-tenant IDOR: Tenant B querying Tenant A's permit returns 404 PERMIT_NOT_FOUND", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupTransitionsRouter(pRepo, cRepo)

		// Repo rejects because tenant mismatch in WHERE clause returns sql.ErrNoRows -> ErrPermitNotFound
		pRepo.On("GetByID", mock.Anything, permitID).Return(nil, domainErrors.ErrPermitNotFound).Once()

		req := httptest.NewRequest("GET", "/api/v1/permits/"+permitID.String(), nil)
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantB), userID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_NOT_FOUND", errObj["code"])
		assert.Contains(t, errObj["message"], "Permiso de trabajo no encontrado o no pertenece al tenant")
		_ = tenantA
	})
}

// -----------------------------------------------------------------------------
// 5. CHALLENGE: High-Concurrency Stress & Race Condition Resistance
// -----------------------------------------------------------------------------
func TestEmpiricalChallenge_Concurrency_RaceConditions(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()
	futureDate := time.Now().Add(30 * 24 * time.Hour)

	pRepo := new(mockPermitRepoForHandler)
	cRepo := new(mockComplianceRepoForHandler)
	router := setupTransitionsRouter(pRepo, cRepo)

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
	pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

	aff := &entity.WorkerAffiliation{
		WorkerID: workerID, FullName: "Trabajador Concurrente", UserIsActive: true, WorkerIsActive: true,
		ContractorID: contractorID, ContractorIsActive: true, PlantStatus: "active",
	}
	cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)
	docs := []entity.DocumentWithTypeCode{
		{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
	}
	cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

	pRepo.On("AssignWorker", mock.Anything, mock.Anything).Return(nil)
	pRepo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &ehsUserID).Return(nil)
	pRepo.On("AddSignature", mock.Anything, mock.Anything).Return(nil)

	concurrency := 25
	var wg sync.WaitGroup
	wg.Add(concurrency)

	results := make([]int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte("{}")))
			ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results[idx] = w.Code
		}(i)
	}

	wg.Wait()

	for _, code := range results {
		assert.Equal(t, http.StatusOK, code)
	}
}
