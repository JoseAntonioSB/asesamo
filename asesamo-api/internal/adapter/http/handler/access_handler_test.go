package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock Repositories for AccessHandler Tests

type MockAPRepoForHandler struct{ mock.Mock }

func (m *MockAPRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *MockAPRepoForHandler) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *MockAPRepoForHandler) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

type MockPlantRepoForHandler struct{ mock.Mock }

func (m *MockPlantRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plant), args.Error(1)
}

func (m *MockPlantRepoForHandler) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Plant), args.Error(1)
}

type MockCompRepoForHandler struct{ mock.Mock }

func (m *MockCompRepoForHandler) GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error) {
	args := m.Called(ctx, plantID, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkerAffiliation), args.Error(1)
}

func (m *MockCompRepoForHandler) GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error) {
	args := m.Called(ctx, tenantID, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.RequiredDocumentRule), args.Error(1)
}

func (m *MockCompRepoForHandler) GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func (m *MockCompRepoForHandler) GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, contractorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

type MockLogRepoForHandler struct{ mock.Mock }

func (m *MockLogRepoForHandler) Create(ctx context.Context, log *entity.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockLogRepoForHandler) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func (m *MockLogRepoForHandler) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessLog), args.Error(1)
}

func (m *MockLogRepoForHandler) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, tenantID, plantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *MockLogRepoForHandler) ListWithFilters(ctx context.Context, filter repository.AccessLogFilter) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *MockLogRepoForHandler) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

// Test Suite

func TestAccessHandler_CheckIn(t *testing.T) {
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
		Name:               "Planta Queretaro",
		BlockOnExpiredDocs: true,
	}

	baseAff := &entity.WorkerAffiliation{
		WorkerID:           workerID,
		FullName:           "Carlos Mendoza",
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       contractorID,
		ContractorName:     "Mantenimiento SA",
		PlantStatus:        "active",
	}

	dc3Req := entity.RequiredDocumentRule{
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		IsRequired:       true,
	}

	future := time.Now().UTC().Add(180 * 24 * time.Hour)
	validDocs := []entity.DocumentWithTypeCode{
		{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia DC-3",
			Status:           entity.DocumentStatusApproved,
			ExpiresAt:        &future,
		},
	}

	t.Run("successful check-in with worker_user_id returns full envelope and document arrays", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(validDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
			"notes":           "Turno matutino",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		ctx := tenant.WithTenant(req.Context(), tenantID)
		ctx = tenant.WithUserID(ctx, uuid.New())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		h.CheckIn(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		dataMap, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, dataMap["access_log_id"])
		assert.Equal(t, "granted", dataMap["result"])
		assert.Nil(t, dataMap["denial_reason"])
		assert.NotEmpty(t, dataMap["checked_in_at"])

		// Verify both document_snapshot and documents arrays exist and are non-empty
		snapshots, ok := dataMap["document_snapshot"].([]interface{})
		require.True(t, ok)
		assert.Len(t, snapshots, 1)

		docs, ok := dataMap["documents"].([]interface{})
		require.True(t, ok)
		assert.Len(t, docs, 1)

		workerMap, ok := dataMap["worker"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Carlos Mendoza", workerMap["full_name"])

		contractorMap, ok := dataMap["contractor"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Mantenimiento SA", contractorMap["name"])
	})

	t.Run("successful check-in with worker_id fallback", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(validDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_id":       workerID.String(),
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non-compliant worker returns denied result with denial reason", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		dataMap := resp.Data.(map[string]interface{})
		assert.Equal(t, "denied", dataMap["result"])
		assert.NotNil(t, dataMap["denial_reason"])
	})

	t.Run("returns 400 when missing worker identifier", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 when missing access_point_id", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id": workerID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed worker UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id":  "invalid-uuid",
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed access_point UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": "not-a-uuid",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid JSON body", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader([]byte("{invalid-json")))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 401 when tenant context is missing", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns domain error when worker is not found", func(t *testing.T) {
		apRepo := new(MockAPRepoForHandler)
		plantRepo := new(MockPlantRepoForHandler)
		compRepo := new(MockCompRepoForHandler)
		logRepo := new(MockLogRepoForHandler)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(nil, domainErrors.ErrWorkerNotFound)

		checkInUC := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		h := handler.NewAccessHandler(checkInUC, nil, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-in", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckIn(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAccessHandler_CheckOut(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	logID := uuid.New()

	checkInTime := time.Now().UTC().Add(-45 * time.Minute)
	checkOutTime := time.Now().UTC()

	expectedLog := &entity.AccessLog{
		ID:           logID,
		TenantID:     tenantID,
		PlantID:      plantID,
		UserID:       workerID,
		Result:       entity.AccessResultGranted,
		CheckedInAt:  checkInTime,
		CheckedOutAt: &checkOutTime,
	}

	t.Run("successful check-out returns access_log_id, checked_out_at, duration_minutes", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(expectedLog, nil)

		checkOutUC := access.NewCheckOutUseCase(logRepo)
		h := handler.NewAccessHandler(nil, checkOutUC, nil)

		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"plant_id":        plantID.String(),
			"access_point_id": apID.String(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		dataMap, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, logID.String(), dataMap["access_log_id"])
		assert.NotEmpty(t, dataMap["checked_out_at"])
		assert.Equal(t, float64(45), dataMap["duration_minutes"])
	})

	t.Run("successful check-out with worker_id fallback", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(expectedLog, nil)

		checkOutUC := access.NewCheckOutUseCase(logRepo)
		h := handler.NewAccessHandler(nil, checkOutUC, nil)

		body := map[string]interface{}{
			"worker_id": workerID.String(),
			"plant_id":  plantID.String(),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 400 on missing worker identifier", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"plant_id": plantID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed worker UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id": "invalid-uuid",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed plant UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id": workerID.String(),
			"plant_id":       "invalid-plant",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on malformed access_point UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		body := map[string]interface{}{
			"worker_user_id":  workerID.String(),
			"access_point_id": "invalid-ap",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid JSON", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader([]byte("{invalid")))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 401 when tenant context is missing", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns error when usecase fails", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(nil, errors.New("db failure"))

		checkOutUC := access.NewCheckOutUseCase(logRepo)
		h := handler.NewAccessHandler(nil, checkOutUC, nil)

		body := map[string]interface{}{
			"worker_user_id": workerID.String(),
			"plant_id":       plantID.String(),
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/access/check-out", bytes.NewReader(bodyBytes))
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.CheckOut(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAccessHandler_ListLogs(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	now := time.Now().UTC()

	sampleLogs := []entity.AccessLog{
		{
			ID:               uuid.New(),
			TenantID:         tenantID,
			PlantID:          plantID,
			AccessPointID:    apID,
			UserID:           workerID,
			Result:           entity.AccessResultGranted,
			CheckedInAt:      now,
			DocumentSnapshot: entity.DocumentSnapshot{},
			CreatedAt:        now,
		},
	}

	t.Run("returns logs and meta with full query parameters", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		resultGranted := entity.AccessResultGranted

		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.TenantID == tenantID &&
				f.PlantID == plantID &&
				f.AccessPointID == apID &&
				f.WorkerUserID == workerID &&
				f.Result != nil && *f.Result == resultGranted &&
				f.From != nil && f.To != nil &&
				f.Limit == 5 && f.Offset == 0
		})).Return(sampleLogs, 1, nil)

		listLogsUC := access.NewListAccessLogsUseCase(logRepo)
		h := handler.NewAccessHandler(nil, nil, listLogsUC)

		url := "/api/v1/access-logs?plant_id=" + plantID.String() +
			"&access_point_id=" + apID.String() +
			"&worker_user_id=" + workerID.String() +
			"&result=granted" +
			"&from=2026-10-01T00:00:00Z" +
			"&to=2026-10-02T23:59:59Z" +
			"&page=1&per_page=5"

		req := httptest.NewRequest(http.MethodGet, url, nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp dto.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		dataList, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Len(t, dataList, 1)

		metaMap, ok := resp.Meta.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(1), metaMap["page"])
		assert.Equal(t, float64(5), metaMap["per_page"])
		assert.Equal(t, float64(1), metaMap["total"])
	})

	t.Run("supports worker_id fallback and date-only from/to", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.WorkerUserID == workerID && f.From != nil && f.To != nil
		})).Return(sampleLogs, 1, nil)

		listLogsUC := access.NewListAccessLogsUseCase(logRepo)
		h := handler.NewAccessHandler(nil, nil, listLogsUC)

		url := "/api/v1/access-logs?worker_id=" + workerID.String() +
			"&from=2026-10-01&to=2026-10-02"

		req := httptest.NewRequest(http.MethodGet, url, nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 400 on invalid plant_id UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?plant_id=invalid", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid access_point_id UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?access_point_id=invalid", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid worker_user_id UUID", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?worker_user_id=invalid", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid result param", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?result=unknown_result", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid from date format", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?from=not-a-date", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 on invalid to date format", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs?to=bad-date", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 401 when tenant context is missing", func(t *testing.T) {
		h := handler.NewAccessHandler(nil, nil, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs", nil)
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns error when usecase fails", func(t *testing.T) {
		logRepo := new(MockLogRepoForHandler)
		logRepo.On("ListWithFilters", mock.Anything, mock.Anything).
			Return(nil, 0, errors.New("timeout"))

		listLogsUC := access.NewListAccessLogsUseCase(logRepo)
		h := handler.NewAccessHandler(nil, nil, listLogsUC)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-logs", nil)
		req = req.WithContext(tenant.WithTenant(req.Context(), tenantID))
		w := httptest.NewRecorder()

		h.ListLogs(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
