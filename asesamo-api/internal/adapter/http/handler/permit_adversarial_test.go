package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/internal/core/usecase/permit"
	"asesamo-api/pkg/tenant"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestAdversarial_CandadoInteligente_HTTPVerifications tests all empirical compliance requirements
// for Milestone 2 Candado Inteligente compliance checks via the HTTP API.
func TestAdversarial_CandadoInteligente_HTTPVerifications(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()
	contractorUserID := uuid.New()

	now := time.Now().UTC()
	futureDate := now.Add(90 * 24 * time.Hour)
	pastDate := now.Add(-48 * time.Hour)

	setupRouter := func(
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
		r.Post("/api/v1/permits/{id}/workers", h.AssignWorkers)
		r.Post("/api/v1/permits/{id}/approve", h.ApprovePermit)
		return r
	}

	// -------------------------------------------------------------------------
	// 1. Missing active DC-3 for each hazard type (NOM-009, NOM-033, NOM-027, NOM-029)
	// -------------------------------------------------------------------------
	hazardScenarios := []struct {
		name              string
		permitType        entity.PermitType
		expectedNOM       string
		presentDocCode    string
		presentDocName    string
		expectCode        string
	}{
		{
			name:           "Rejection when worker lacks active DC-3 for Heights (NOM-009)",
			permitType:     entity.PermitTypeHeights,
			expectedNOM:    "NOM-009",
			presentDocCode: "",
			expectCode:     "WORKER_MISSING_DC3",
		},
		{
			name:           "Rejection when worker lacks active DC-3 for Confined Space (NOM-033)",
			permitType:     entity.PermitTypeConfinedSpace,
			expectedNOM:    "NOM-033",
			presentDocCode: "",
			expectCode:     "WORKER_MISSING_DC3",
		},
		{
			name:           "Rejection when worker lacks active DC-3 for Hot Work (NOM-027)",
			permitType:     entity.PermitTypeHotWork,
			expectedNOM:    "NOM-027",
			presentDocCode: "",
			expectCode:     "WORKER_MISSING_DC3",
		},
		{
			name:           "Rejection when worker lacks active DC-3 for Electrical LOTO (NOM-029)",
			permitType:     entity.PermitTypeElectricalLOTO,
			expectedNOM:    "NOM-029",
			presentDocCode: "",
			expectCode:     "WORKER_MISSING_DC3",
		},
		{
			name:           "Rejection when worker has DC-3 for Alturas but work is Confined Space (NOM-033)",
			permitType:     entity.PermitTypeConfinedSpace,
			expectedNOM:    "NOM-033",
			presentDocCode: "DC3_ALTURAS",
			presentDocName: "dc3_alturas.pdf",
			expectCode:     "WORKER_MISSING_DC3",
		},
		{
			name:           "Rejection when worker has DC-3 for Soldadura but work is Electrical LOTO (NOM-029)",
			permitType:     entity.PermitTypeElectricalLOTO,
			expectedNOM:    "NOM-029",
			presentDocCode: "DC3_CORTE_SOLDADURA",
			presentDocName: "dc3_soldadura.pdf",
			expectCode:     "WORKER_MISSING_DC3",
		},
	}

	for _, tc := range hazardScenarios {
		t.Run(tc.name, func(t *testing.T) {
			pRepo := new(mockPermitRepoForHandler)
			cRepo := new(mockComplianceRepoForHandler)
			router := setupRouter(pRepo, cRepo)

			wp := &entity.WorkPermit{
				ID:           permitID,
				TenantID:     tenantID,
				PlantID:      plantID,
				ContractorID: contractorID,
				PermitType:   tc.permitType,
				Status:       entity.PermitStatusDraft,
			}
			pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

			aff := &entity.WorkerAffiliation{
				WorkerID:           workerID,
				FullName:           "Juan Valenzuela",
				UserIsActive:       true,
				WorkerIsActive:     true,
				ContractorID:       contractorID,
				ContractorName:     "Constructora Industrial SA",
				ContractorIsActive: true,
				PlantStatus:        "active",
			}
			cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

			docs := []entity.DocumentWithTypeCode{
				{
					DocumentTypeCode: "MEDICAL",
					DocumentName:     "certificado_medico.pdf",
					Status:           entity.DocumentStatusApproved,
					ExpiresAt:        &futureDate,
				},
			}
			if tc.presentDocCode != "" {
				docs = append(docs, entity.DocumentWithTypeCode{
					DocumentTypeCode: tc.presentDocCode,
					DocumentName:     tc.presentDocName,
					Status:           entity.DocumentStatusApproved,
					ExpiresAt:        &futureDate,
				})
			}
			cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected HTTP 422 for missing DC-3")

			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			errObj, ok := resp["error"].(map[string]any)
			require.True(t, ok, "Response must contain 'error' object")
			assert.Equal(t, tc.expectCode, errObj["code"])
			assert.Contains(t, errObj["message"], tc.expectedNOM)
		})
	}

	// -------------------------------------------------------------------------
	// 2. Expired DC-3 certificate
	// -------------------------------------------------------------------------
	t.Run("Rejection when worker has expired DC-3", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Roberto Vencido",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Techumbres SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		docs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas_exp.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastDate,
			},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected HTTP 422 for expired DC-3")

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "WORKER_MISSING_DC3", errObj["code"])
		assert.Contains(t, errObj["message"], "vencida desde")
	})

	// -------------------------------------------------------------------------
	// 3. Missing or expired medical certificate
	// -------------------------------------------------------------------------
	t.Run("Rejection when worker has missing medical certificate", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Luis Sin Medico",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Techumbres SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		// Only has DC-3, no MEDICAL
		docs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected HTTP 422 for missing medical certificate")

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "WORKER_MISSING_MEDICAL", errObj["code"])
		assert.Contains(t, errObj["message"], "certificado médico de aptitud")
	})

	t.Run("Rejection when worker has expired medical certificate", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Luis Medico Vencido",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Techumbres SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		docs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico_antiguo.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastDate,
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected HTTP 422 for expired medical certificate")

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "WORKER_MISSING_MEDICAL", errObj["code"])
		assert.Contains(t, errObj["message"], "vencido desde")
	})

	// -------------------------------------------------------------------------
	// 4. Contractor inactive or suspended in plant
	// -------------------------------------------------------------------------
	contractorScenarios := []struct {
		name               string
		contractorIsActive bool
		plantStatus        string
		userIsActive       bool
		workerIsActive     bool
		expectedCode       string
		expectedMsg        string
	}{
		{
			name:               "Rejection when contractor is suspended in plant",
			contractorIsActive: true,
			plantStatus:        "suspended",
			userIsActive:       true,
			workerIsActive:     true,
			expectedCode:       "CONTRACTOR_INACTIVE",
			expectedMsg:        "no se encuentra activa o está suspendida en esta planta",
		},
		{
			name:               "Rejection when contractor is inactive in plant",
			contractorIsActive: true,
			plantStatus:        "inactive",
			userIsActive:       true,
			workerIsActive:     true,
			expectedCode:       "CONTRACTOR_INACTIVE",
			expectedMsg:        "no se encuentra activa o está suspendida en esta planta",
		},
		{
			name:               "Rejection when contractor is globally inactive",
			contractorIsActive: false,
			plantStatus:        "active",
			userIsActive:       true,
			workerIsActive:     true,
			expectedCode:       "CONTRACTOR_INACTIVE",
			expectedMsg:        "no se encuentra activa o está suspendida en esta planta",
		},
		{
			name:               "Rejection when worker profile is inactive",
			contractorIsActive: true,
			plantStatus:        "active",
			userIsActive:       false,
			workerIsActive:     true,
			expectedCode:       "CANDADO_INTELIGENTE_BLOCKED",
			expectedMsg:        "El perfil del trabajador se encuentra inactivo",
		},
	}

	for _, tc := range contractorScenarios {
		t.Run(tc.name, func(t *testing.T) {
			pRepo := new(mockPermitRepoForHandler)
			cRepo := new(mockComplianceRepoForHandler)
			router := setupRouter(pRepo, cRepo)

			wp := &entity.WorkPermit{
				ID:           permitID,
				TenantID:     tenantID,
				PlantID:      plantID,
				ContractorID: contractorID,
				PermitType:   entity.PermitTypeHeights,
				Status:       entity.PermitStatusDraft,
			}
			pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

			aff := &entity.WorkerAffiliation{
				WorkerID:           workerID,
				FullName:           "Trabajador Contratista Inactivo",
				UserIsActive:       tc.userIsActive,
				WorkerIsActive:     tc.workerIsActive,
				ContractorID:       contractorID,
				ContractorName:     "Contratista Bloqueado SA",
				ContractorIsActive: tc.contractorIsActive,
				PlantStatus:        tc.plantStatus,
			}
			cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

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

			// Check HTTP status code and Spanish error message
			var resp map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			errObj, ok := resp["error"].(map[string]any)
			require.True(t, ok, "Response must contain 'error' envelope")
			assert.Equal(t, tc.expectedCode, errObj["code"])
			assert.Contains(t, errObj["message"], tc.expectedMsg)

			// Record observed HTTP status for analysis
			t.Logf("[%s] HTTP status returned: %d, Code: %s, Message: %s",
				tc.name, w.Code, errObj["code"], errObj["message"])
		})
	}

	// -------------------------------------------------------------------------
	// 5. Rejection during EHS Approval if worker compliance degrades
	// -------------------------------------------------------------------------
	t.Run("EHS Approval re-evaluates Candado and blocks permit if worker DC-3 expired in interim", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

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
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Valido En Asignacion",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		// Worker DC-3 expired before EHS approval
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &pastDate},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte(`{"notes":"Intento de aprobacion"}`)))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code, "Expected HTTP 422 when approving permit with non-compliant crew")
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "WORKER_MISSING_DC3", errObj["code"])
		assert.Contains(t, errObj["message"], "vencida desde")
	})

	// -------------------------------------------------------------------------
	// 6. Incomplete checklist blocks EHS approval with HTTP 422
	// -------------------------------------------------------------------------
	t.Run("EHS Approval blocks when mandatory checklist items are not verified (HTTP 422)", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_arnes", IsMandatory: true, IsVerified: false}, // Unverified!
			},
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/permits/"+permitID.String()+"/approve", bytes.NewReader([]byte(`{}`)))
		ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(req.Context(), tenantID), ehsUserID), "ehs_manager")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "CHECKLIST_INCOMPLETE", errObj["code"])
		assert.Contains(t, errObj["message"], "faltan verificaciones obligatorias")
	})

	// -------------------------------------------------------------------------
	// 7. Modifying crew on closed permit returns HTTP 422
	// -------------------------------------------------------------------------
	t.Run("Modifying crew on closed permit is rejected with HTTP 422", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusClosed, // Already closed!
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

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		errObj := resp["error"].(map[string]any)
		assert.Equal(t, "PERMIT_INVALID_STATE", errObj["code"])
	})

	// -------------------------------------------------------------------------
	// 8. Worker has old expired DC-3 AND a newer valid DC-3
	// -------------------------------------------------------------------------
	t.Run("Worker with multiple DC-3s (first expired, second valid)", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan MultiDoc",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Constructora SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		// Docs array has expired DC-3 first, then valid DC-3
		docs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas_old.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastDate, // Expired
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas_new.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate, // Valid!
			},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		t.Logf("[Multi-DC3 Test] HTTP Status: %d, Response: %v", w.Code, resp)
	})

	// -------------------------------------------------------------------------
	// 9. Worker with valid active DC-3 plus newer pending_review renewal
	// -------------------------------------------------------------------------
	t.Run("Worker with active DC-3 and newer pending_review renewal", func(t *testing.T) {
		pRepo := new(mockPermitRepoForHandler)
		cRepo := new(mockComplianceRepoForHandler)
		router := setupRouter(pRepo, cRepo)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		pRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Renovacion",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Constructora SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		cRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil).Once()

		// Newest uploaded doc (first in slice due to created_at DESC) is pending_review
		// Older uploaded doc (second in slice) is approved and unexpired
		docs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas_renovacion_2027.pdf",
				Status:           entity.DocumentStatusPendingReview, // Pending!
				ExpiresAt:        &futureDate,
			},
			{
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas_actual.pdf",
				Status:           entity.DocumentStatusApproved, // Active and valid!
				ExpiresAt:        &futureDate,
			},
		}
		cRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil).Once()

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

		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)

		t.Logf("[Pending Renewal Test] HTTP Status: %d, Response: %v", w.Code, resp)
	})
}
