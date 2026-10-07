package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdversarial_AccessHandler_HTTPBoundary(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	contractorID := uuid.New()

	baseAP := &entity.AccessPoint{
		ID:       apID,
		TenantID: tenantID,
		PlantID:  plantID,
		Name:     "Caseta Norte",
		IsActive: true,
	}

	basePlant := &entity.Plant{
		ID:                 plantID,
		TenantID:           tenantID,
		Name:               "Planta Ramos Arizpe",
		BlockOnExpiredDocs: true,
	}

	baseAff := &entity.WorkerAffiliation{
		WorkerID:           workerID,
		FullName:           "Miguel Hernández",
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       contractorID,
		ContractorName:     "Contratista Regio SA",
		PlantStatus:        "active",
	}

	reqDC3 := entity.RequiredDocumentRule{
		DocumentTypeID:   uuid.New(),
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		IsRequired:       true,
	}

	validFuture := time.Now().UTC().AddDate(1, 0, 0)
	docDC3Valid := entity.DocumentWithTypeCode{
		ID:               uuid.New(),
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		Status:           entity.DocumentStatusApproved,
		ExpiresAt:        &validFuture,
	}

	t.Run("HTTP Check-In: inactive worker returns HTTP 200 with result=denied and denial reason", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		inactiveAff := *baseAff
		inactiveAff.UserIsActive = false
		inactiveAff.IsActive = false

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&inactiveAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{docDC3Valid}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil
		})).Return(nil)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
		}
		raw, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(raw))
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CheckIn(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data access.CheckInOutput `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, entity.AccessResultDenied, resp.Data.Result)
		require.NotNil(t, resp.Data.DenialReason)
		assert.Contains(t, *resp.Data.DenialReason, "inactivo")
	})

	t.Run("HTTP Check-In: missing required document returns HTTP 200 with result=denied and Spanish reason", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "faltante")
		})).Return(nil)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
		}
		raw, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(raw))
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CheckIn(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data access.CheckInOutput `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, entity.AccessResultDenied, resp.Data.Result)
		require.NotNil(t, resp.Data.DenialReason)
		assert.Contains(t, *resp.Data.DenialReason, "faltante")
		assert.Contains(t, *resp.Data.DenialReason, "Constancia DC-3")
	})

	t.Run("HTTP Check-Out: duration calculated and optional plant_id resolved", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		apRepo := new(MockAPRepoForHandler)

		resolvedPlantID := uuid.New()
		apRepo.On("GetByID", mock.Anything, apID).Return(&entity.AccessPoint{
			ID:       apID,
			TenantID: tenantID,
			PlantID:  resolvedPlantID,
		}, nil)

		inTime := time.Now().UTC().Add(-75 * time.Minute)
		outTime := time.Now().UTC()

		logRepo.On("CheckOutLatest", mock.Anything, tenantID, resolvedPlantID, workerID, mock.Anything).Return(&entity.AccessLog{
			ID:           uuid.New(),
			TenantID:     tenantID,
			PlantID:      resolvedPlantID,
			UserID:       workerID,
			CheckedInAt:  inTime,
			CheckedOutAt: &outTime,
			Result:       entity.AccessResultGranted,
		}, nil)

		checkOutUC := access.NewCheckOutUseCase(logRepo, apRepo)
		h := handler.NewAccessHandler(nil, checkOutUC, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
			// plant_id is intentionally omitted
		}
		raw, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(raw))
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.CheckOut(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data access.CheckOutOutput `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, 75, resp.Data.DurationMinutes)
	})

	t.Run("HTTP List Logs: multi-filter, pagination meta, and malformed parameter rejections", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)

		// 1. Success with all valid filters
		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.TenantID == tenantID &&
				f.PlantID == plantID &&
				f.AccessPointID == apID &&
				f.WorkerUserID == workerID &&
				f.Result != nil && *f.Result == entity.AccessResultGranted &&
				f.Limit == 20 &&
				f.Offset == 0
		})).Return([]entity.AccessLog{
			{ID: uuid.New(), TenantID: tenantID, PlantID: plantID, Result: entity.AccessResultGranted},
		}, 1, nil)

		listUC := access.NewListAccessLogsUseCase(logRepo)
		h := handler.NewAccessHandler(nil, nil, listUC)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?plant_id="+plantID.String()+
			"&access_point_id="+apID.String()+
			"&worker_user_id="+workerID.String()+
			"&result=granted&from=2026-10-01&to=2026-10-07&page=1&per_page=20", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		h.ListLogs(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data []entity.AccessLog `json:"data"`
			Meta dto.PaginationMeta `json:"meta"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, 1, resp.Meta.Total)
		assert.Equal(t, 1, resp.Meta.Page)
		assert.Equal(t, 20, resp.Meta.PerPage)

		// 2. Malformed result parameter -> 400
		badReq := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?result=invalid_status", nil)
		badReq = badReq.WithContext(ctx)
		wBad := httptest.NewRecorder()
		h.ListLogs(wBad, badReq)
		assert.Equal(t, http.StatusBadRequest, wBad.Code)
		assert.Contains(t, wBad.Body.String(), "INVALID_INPUT")

		// 3. Malformed date format -> 400
		badDateReq := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?from=not-a-date", nil)
		badDateReq = badDateReq.WithContext(ctx)
		wBadDate := httptest.NewRecorder()
		h.ListLogs(wBadDate, badDateReq)
		assert.Equal(t, http.StatusBadRequest, wBadDate.Code)
		assert.Contains(t, wBadDate.Body.String(), "INVALID_INPUT")
	})
}
