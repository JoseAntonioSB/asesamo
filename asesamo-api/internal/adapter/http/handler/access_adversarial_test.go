package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdversarial_AccessHandler(t *testing.T) {
	tenantID := uuid.New()
	workerUserID := uuid.New()
	apID := uuid.New()
	plantID := uuid.New()

	baseAP := &entity.AccessPoint{
		ID:       apID,
		TenantID: tenantID,
		PlantID:  plantID,
		Name:     "Caseta Principal",
		IsActive: true,
	}
	basePlant := &entity.Plant{
		ID:                 plantID,
		TenantID:           tenantID,
		Name:               "Planta Principal",
		BlockOnExpiredDocs: true,
	}
	baseAff := &entity.WorkerAffiliation{
		WorkerID:           workerUserID,
		FullName:           "Carlos Mendoza",
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       uuid.New(),
		ContractorName:     "Contratista MX",
		PlantStatus:        "active",
	}

	t.Run("INJECTION_ATTACK: SQL injection patterns in query params rejected with 400", func(t *testing.T) {
		mockLogRepo := new(MockLogRepoForHandler)
		listUC := access.NewListAccessLogsUseCase(mockLogRepo)
		h := handler.NewAccessHandler(nil, nil, listUC)

		injectionPayloads := []string{
			"'; DROP TABLE access_logs; --",
			"1' OR '1'='1",
			"00000000-0000-0000-0000-000000000000' UNION SELECT NULL--",
			"<script>alert(1)</script>",
		}

		for _, payload := range injectionPayloads {
			escaped := url.QueryEscape(payload)
			// Test plant_id
			req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?plant_id="+escaped, nil)
			req = req.WithContext(tenant.WithTenant(context.Background(), tenantID))
			rec := httptest.NewRecorder()
			h.ListLogs(rec, req)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "SQL injection in plant_id must yield 400")

			// Test access_point_id
			reqAP := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?access_point_id="+escaped, nil)
			reqAP = reqAP.WithContext(tenant.WithTenant(context.Background(), tenantID))
			recAP := httptest.NewRecorder()
			h.ListLogs(recAP, reqAP)
			assert.Equal(t, http.StatusBadRequest, recAP.Code, "SQL injection in access_point_id must yield 400")

			// Test worker_user_id
			reqWorker := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?worker_user_id="+escaped, nil)
			reqWorker = reqWorker.WithContext(tenant.WithTenant(context.Background(), tenantID))
			recWorker := httptest.NewRecorder()
			h.ListLogs(recWorker, reqWorker)
			assert.Equal(t, http.StatusBadRequest, recWorker.Code, "SQL injection in worker_user_id must yield 400")

			// Test result
			reqRes := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?result="+escaped, nil)
			reqRes = reqRes.WithContext(tenant.WithTenant(context.Background(), tenantID))
			recRes := httptest.NewRecorder()
			h.ListLogs(recRes, reqRes)
			assert.Equal(t, http.StatusBadRequest, recRes.Code, "Invalid result enum must yield 400")
		}
	})

	t.Run("DATE_PARSING_ADVERSARIAL: Malformed and invalid dates yield 400", func(t *testing.T) {
		mockLogRepo := new(MockLogRepoForHandler)
		listUC := access.NewListAccessLogsUseCase(mockLogRepo)
		h := handler.NewAccessHandler(nil, nil, listUC)

		invalidDates := []string{
			"2025-02-29",       // Not a leap year!
			"2026-13-01",       // Invalid month
			"2026-00-01",       // Month 00
			"yesterday",        // Non-ISO string
			"01/01/2026",       // US date format
			"2026-10-07 12:00", // Missing seconds / timezone
		}

		for _, dt := range invalidDates {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?from="+url.QueryEscape(dt), nil)
			req = req.WithContext(tenant.WithTenant(context.Background(), tenantID))
			rec := httptest.NewRecorder()
			h.ListLogs(rec, req)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "Malformed date "+dt+" must yield 400")
			assert.Contains(t, rec.Body.String(), "El formato de fecha no es válido")
		}
	})

	t.Run("PAGINATION_EXTREMES: Negative and massive page sizes handled cleanly", func(t *testing.T) {
		mockLogRepo := new(MockLogRepoForHandler)
		mockLogRepo.On("ListWithFilters", mock.Anything, mock.AnythingOfType("repository.AccessLogFilter")).Return([]entity.AccessLog{}, 0, nil)

		listUC := access.NewListAccessLogsUseCase(mockLogRepo)
		h := handler.NewAccessHandler(nil, nil, listUC)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?page=-5&per_page=9999", nil)
		req = req.WithContext(tenant.WithTenant(context.Background(), tenantID))
		rec := httptest.NewRecorder()
		h.ListLogs(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"page":1`)
		assert.Contains(t, rec.Body.String(), `"per_page":100`)
	})

	t.Run("CONCURRENT_HTTP_CHECKINS: 30 parallel HTTP CheckIn requests", func(t *testing.T) {
		mockAP := new(MockAPRepoForHandler)
		mockPlant := new(MockPlantRepoForHandler)
		mockComp := new(MockCompRepoForHandler)
		mockLog := new(MockLogRepoForHandler)

		mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAff, nil)
		mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{}, nil)
		mockComp.On("GetContractorDocuments", mock.Anything, mock.Anything).Return([]entity.DocumentWithTypeCode{}, nil)
		mockLog.On("Create", mock.Anything, mock.AnythingOfType("*entity.AccessLog")).Return(nil)

		checkInUC := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		const concurrentCount = 30
		var wg sync.WaitGroup
		statusCodes := make(chan int, concurrentCount)

		for i := 0; i < concurrentCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				body := fmt.Sprintf(`{"worker_user_id":"%s","access_point_id":"%s"}`, workerUserID.String(), apID.String())
				req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(tenant.WithTenant(context.Background(), tenantID))
				rec := httptest.NewRecorder()

				h.CheckIn(rec, req)
				statusCodes <- rec.Code
			}()
		}

		wg.Wait()
		close(statusCodes)

		for code := range statusCodes {
			assert.Equal(t, http.StatusOK, code)
		}
	})
}
