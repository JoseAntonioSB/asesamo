package access_test

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

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

// ThreadSafeMockAccessLogRepo protects calls under heavy concurrent testing.
type ThreadSafeMockAccessLogRepo struct {
	mu   sync.Mutex
	mock.Mock
}

func (m *ThreadSafeMockAccessLogRepo) Create(ctx context.Context, log *entity.AccessLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *ThreadSafeMockAccessLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func (m *ThreadSafeMockAccessLogRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessLog), args.Error(1)
}

func (m *ThreadSafeMockAccessLogRepo) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, tenantID, plantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *ThreadSafeMockAccessLogRepo) ListWithFilters(ctx context.Context, filter repository.AccessLogFilter) ([]entity.AccessLog, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *ThreadSafeMockAccessLogRepo) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func TestAdversarial_AccessVerificationEngine(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerUserID := uuid.New()
	contractorID := uuid.New()

	baseAP := &entity.AccessPoint{
		ID:       apID,
		TenantID: tenantID,
		PlantID:  plantID,
		Name:     "Caseta Norte",
		IsActive: true,
	}

	basePlant := &entity.Plant{
		ID:                  plantID,
		TenantID:            tenantID,
		Name:                "Planta Querétaro",
		BlockOnExpiredDocs:  true,
	}

	baseAffiliation := &entity.WorkerAffiliation{
		WorkerID:           workerUserID,
		FullName:           "Juan Pérez",
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       contractorID,
		ContractorName:     "Contratista Alfa",
		PlantStatus:        "active",
	}

	dc3Req := entity.RequiredDocumentRule{
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		IsRequired:       true,
	}

	repseReq := entity.RequiredDocumentRule{
		DocumentTypeCode: "REPSE",
		DocumentName:     "Registro REPSE",
		AppliesTo:        "contractor",
		IsRequired:       true,
	}

	futureDate := time.Now().UTC().AddDate(1, 0, 0)
	validDC3 := entity.DocumentWithTypeCode{
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		Status:           entity.DocumentStatusApproved,
		ExpiresAt:        &futureDate,
	}
	validREPSE := entity.DocumentWithTypeCode{
		DocumentTypeCode: "REPSE",
		DocumentName:     "Registro REPSE",
		AppliesTo:        "contractor",
		Status:           entity.DocumentStatusApproved,
		ExpiresAt:        &futureDate,
	}

	t.Run("CONCURRENCY_STRESS: 50 concurrent check-ins under race detector", func(t *testing.T) {
		mockAP := new(mockAPRepo)
		mockPlant := new(mockPlantRepo)
		mockComp := new(mockCompRepo)
		mockLog := new(ThreadSafeMockAccessLogRepo)

		mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAffiliation, nil)
		mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req, repseReq}, nil)
		mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{validDC3}, nil)
		mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{validREPSE}, nil)
		mockLog.On("Create", mock.Anything, mock.AnythingOfType("*entity.AccessLog")).Return(nil)

		uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)

		const workers = 50
		var wg sync.WaitGroup
		errs := make(chan error, workers)

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(iteration int) {
				defer wg.Done()
				// Add tiny random jitter to maximize concurrent interleaving
				time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
				ctx := tenant.WithTenant(context.Background(), tenantID)
				out, err := uc.Execute(ctx, access.CheckInInput{
					TenantID:      tenantID,
					WorkerUserID:  workerUserID,
					AccessPointID: apID,
				})
				if err != nil {
					errs <- err
					return
				}
				if out.Result != entity.AccessResultGranted {
					errs <- fmt.Errorf("iteration %d: expected granted, got %s", iteration, out.Result)
					return
				}
			}(i)
		}

		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
	})

	t.Run("CROSS_TENANT_ATTACK: Access point from foreign tenant is rejected", func(t *testing.T) {
		foreignTenantID := uuid.New()
		foreignAP := &entity.AccessPoint{
			ID:       apID,
			TenantID: foreignTenantID, // foreign!
			PlantID:  plantID,
			Name:     "Caseta Ajena",
			IsActive: true,
		}

		mockAP := new(mockAPRepo)
		mockPlant := new(mockPlantRepo)
		mockComp := new(mockCompRepo)
		mockLog := new(ThreadSafeMockAccessLogRepo)

		mockAP.On("GetByID", mock.Anything, apID).Return(foreignAP, nil)

		uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
		ctx := tenant.WithTenant(context.Background(), tenantID)

		out, err := uc.Execute(ctx, access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerUserID,
			AccessPointID: apID,
		})

		assert.Nil(t, out)
		assert.ErrorIs(t, err, domainErrors.ErrAccessPointNotFound)
	})

	t.Run("EXPIRATION_BOUNDARY: Document expiring 1 second in past vs future", func(t *testing.T) {
		now := time.Now().UTC()
		pastTime := now.Add(-1 * time.Second)
		futureTime := now.Add(40 * 24 * time.Hour) // > 30 days ahead

		// Test Past: Expired -> Denied when BlockOnExpiredDocs = true
		{
			expiredDC3 := entity.DocumentWithTypeCode{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				AppliesTo:        "worker",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastTime,
			}

			mockAP := new(mockAPRepo)
			mockPlant := new(mockPlantRepo)
			mockComp := new(mockCompRepo)
			mockLog := new(ThreadSafeMockAccessLogRepo)

			mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
			mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
			mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAffiliation, nil)
			mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
			mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{expiredDC3}, nil)
			mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
			mockLog.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AccessLog) bool {
				return log.Result == entity.AccessResultDenied && log.DenialReason != nil
			})).Return(nil)

			uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
			ctx := tenant.WithTenant(context.Background(), tenantID)

			out, err := uc.Execute(ctx, access.CheckInInput{
				TenantID:      tenantID,
				WorkerUserID:  workerUserID,
				AccessPointID: apID,
			})
			require.NoError(t, err)
			assert.Equal(t, entity.AccessResultDenied, out.Result)
			assert.NotNil(t, out.DenialReason)
		}

		// Test Future (> 30 days): Granted without warning
		{
			freshDC3 := entity.DocumentWithTypeCode{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				AppliesTo:        "worker",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureTime,
			}

			mockAP := new(mockAPRepo)
			mockPlant := new(mockPlantRepo)
			mockComp := new(mockCompRepo)
			mockLog := new(ThreadSafeMockAccessLogRepo)

			mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
			mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
			mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAffiliation, nil)
			mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
			mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{freshDC3}, nil)
			mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
			mockLog.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AccessLog) bool {
				return log.Result == entity.AccessResultGranted && log.DenialReason == nil
			})).Return(nil)

			uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
			ctx := tenant.WithTenant(context.Background(), tenantID)

			out, err := uc.Execute(ctx, access.CheckInInput{
				TenantID:      tenantID,
				WorkerUserID:  workerUserID,
				AccessPointID: apID,
			})
			require.NoError(t, err)
			assert.Equal(t, entity.AccessResultGranted, out.Result)
			assert.Nil(t, out.DenialReason)
		}
	})

	t.Run("WARNING_WINDOW_BOUNDARY: Document expiring in 15 days yields granted_with_warning", func(t *testing.T) {
		warningExp := time.Now().UTC().Add(15 * 24 * time.Hour)
		warningDC3 := entity.DocumentWithTypeCode{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia DC-3",
			AppliesTo:        "worker",
			Status:           entity.DocumentStatusApproved,
			ExpiresAt:        &warningExp,
		}

		mockAP := new(mockAPRepo)
		mockPlant := new(mockPlantRepo)
		mockComp := new(mockCompRepo)
		mockLog := new(ThreadSafeMockAccessLogRepo)

		mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAffiliation, nil)
		mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{warningDC3}, nil)
		mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		mockLog.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AccessLog) bool {
			return log.Result == entity.AccessResultGrantedWithWarning
		})).Return(nil)

		uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
		ctx := tenant.WithTenant(context.Background(), tenantID)

		out, err := uc.Execute(ctx, access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerUserID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		assert.Equal(t, entity.AccessResultGrantedWithWarning, out.Result)
		assert.Nil(t, out.DenialReason)
	})

	t.Run("POLICY_PERMUTATION: Missing document denies even if block_on_expired_docs = false", func(t *testing.T) {
		lenientPlant := &entity.Plant{
			ID:                 plantID,
			TenantID:           tenantID,
			Name:               "Planta Tolerante",
			BlockOnExpiredDocs: false, // Expired docs are allowed with warning
		}

		// REPSE is missing
		mockAP := new(mockAPRepo)
		mockPlant := new(mockPlantRepo)
		mockComp := new(mockCompRepo)
		mockLog := new(ThreadSafeMockAccessLogRepo)

		mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		mockPlant.On("GetByID", mock.Anything, plantID).Return(lenientPlant, nil)
		mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(baseAffiliation, nil)
		mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req, repseReq}, nil)
		mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{validDC3}, nil)
		mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		mockLog.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AccessLog) bool {
			return log.Result == entity.AccessResultDenied
		})).Return(nil)

		uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
		ctx := tenant.WithTenant(context.Background(), tenantID)

		out, err := uc.Execute(ctx, access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerUserID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		assert.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "Registro REPSE")
	})

	t.Run("CONTRACTOR_STATUS_PERMUTATIONS: All non-active statuses deny access", func(t *testing.T) {
		disallowedStatuses := []string{"pending", "suspended", "blocked", "revoked", "under_audit"}

		for _, status := range disallowedStatuses {
			t.Run("status="+status, func(t *testing.T) {
				badAffiliation := *baseAffiliation
				badAffiliation.PlantStatus = status

				mockAP := new(mockAPRepo)
				mockPlant := new(mockPlantRepo)
				mockComp := new(mockCompRepo)
				mockLog := new(ThreadSafeMockAccessLogRepo)

				mockAP.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
				mockPlant.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
				mockComp.On("GetWorkerAffiliation", mock.Anything, plantID, workerUserID).Return(&badAffiliation, nil)
				mockComp.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
				mockComp.On("GetWorkerDocuments", mock.Anything, workerUserID).Return([]entity.DocumentWithTypeCode{}, nil)
				mockComp.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
				mockLog.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AccessLog) bool {
					return log.Result == entity.AccessResultDenied
				})).Return(nil)

				uc := access.NewCheckInUseCase(mockAP, mockPlant, mockComp, mockLog)
				ctx := tenant.WithTenant(context.Background(), tenantID)

				out, err := uc.Execute(ctx, access.CheckInInput{
					TenantID:      tenantID,
					WorkerUserID:  workerUserID,
					AccessPointID: apID,
				})
				require.NoError(t, err)
				assert.Equal(t, entity.AccessResultDenied, out.Result)
				assert.NotNil(t, out.DenialReason)
			})
		}
	})

	t.Run("CHECKOUT_CLOCK_DRIFT: CheckedOutAt before CheckedInAt never produces negative duration", func(t *testing.T) {
		checkInTime := time.Now().UTC()
		checkOutTime := checkInTime.Add(-10 * time.Minute) // 10 minutes in past (skew)

		logRecord := &entity.AccessLog{
			ID:           uuid.New(),
			TenantID:     tenantID,
			PlantID:      plantID,
			UserID:       workerUserID,
			CheckedInAt:  checkInTime,
			CheckedOutAt: &checkOutTime,
			Result:       entity.AccessResultGranted,
		}

		mockLog := new(ThreadSafeMockAccessLogRepo)
		mockLog.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerUserID, mock.AnythingOfType("time.Time")).Return(logRecord, nil)

		uc := access.NewCheckOutUseCase(mockLog)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:     tenantID,
			PlantID:      plantID,
			WorkerUserID: workerUserID,
		})
		require.NoError(t, err)
		assert.Equal(t, 0, out.DurationMinutes, "Duration should be clamped to 0 under clock drift")
	})
}
