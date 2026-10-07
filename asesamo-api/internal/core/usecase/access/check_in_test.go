package access_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/access"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAccessPointRepository
type mockAPRepo struct{ mock.Mock }

func (m *mockAPRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

// MockPlantRepository
type mockPlantRepo struct{ mock.Mock }

func (m *mockPlantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Plant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plant), args.Error(1)
}

func (m *mockPlantRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.Plant, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Plant), args.Error(1)
}

// MockComplianceRepository
type mockCompRepo struct{ mock.Mock }

func (m *mockCompRepo) GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error) {
	args := m.Called(ctx, plantID, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkerAffiliation), args.Error(1)
}

func (m *mockCompRepo) GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error) {
	args := m.Called(ctx, tenantID, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.RequiredDocumentRule), args.Error(1)
}

func (m *mockCompRepo) GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func (m *mockCompRepo) GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, contractorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

// MockAccessLogRepository
type mockLogRepo struct{ mock.Mock }

func (m *mockLogRepo) Create(ctx context.Context, log *entity.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func (m *mockLogRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessLog), args.Error(1)
}

func (m *mockLogRepo) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, tenantID, plantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockLogRepo) ListWithFilters(ctx context.Context, filter repository.AccessLogFilter) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockLogRepo) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func TestCheckInUseCase_Execute(t *testing.T) {
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
		Name:               "Planta Principal",
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

	t.Run("compliant worker is granted access", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)

		future := time.Now().Add(180 * 24 * time.Hour)
		workerDocs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &future,
			},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(workerDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultGranted && l.DenialReason == nil
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.NoError(t, err)
		assert.Equal(t, entity.AccessResultGranted, out.Result)
		assert.Nil(t, out.DenialReason)
		assert.Equal(t, "Carlos Mendoza", out.Worker.FullName)
		assert.Len(t, out.DocumentSnapshot, 1)
		assert.True(t, out.DocumentSnapshot[0].IsValid)
	})

	t.Run("expired document with block_on_expired_docs = true is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)

		past := time.Now().Add(-30 * 24 * time.Hour)
		workerDocs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &past,
			},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(workerDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.NoError(t, err)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "vencido")
	})

	t.Run("expired document with block_on_expired_docs = false yields warning", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		lenientPlant := &entity.Plant{
			ID:                 plantID,
			TenantID:           tenantID,
			Name:               "Planta Tolerante",
			BlockOnExpiredDocs: false,
		}

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(lenientPlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)

		past := time.Now().Add(-10 * 24 * time.Hour)
		workerDocs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &past,
			},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(workerDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultGrantedWithWarning
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.NoError(t, err)
		assert.Equal(t, entity.AccessResultGrantedWithWarning, out.Result)
	})

	t.Run("nonexistent worker returns error", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(nil, domainErrors.ErrWorkerNotFound)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		_, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.Error(t, err)
		assert.True(t, domainErrors.IsNotFound(err))
	})

	t.Run("nil worker_user_id returns ErrInvalidInput", func(t *testing.T) {
		uc := access.NewCheckInUseCase(new(mockAPRepo), new(mockPlantRepo), new(mockCompRepo), new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  uuid.Nil,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, domainErrors.IsInvalidInput(err))
	})

	t.Run("nil access_point_id returns ErrInvalidInput", func(t *testing.T) {
		uc := access.NewCheckInUseCase(new(mockAPRepo), new(mockPlantRepo), new(mockCompRepo), new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: uuid.Nil,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, domainErrors.IsInvalidInput(err))
	})

	t.Run("access point lookup error is returned", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		expectedErr := errors.New("ap lookup error")
		apRepo.On("GetByID", mock.Anything, apID).Return(nil, expectedErr)

		uc := access.NewCheckInUseCase(apRepo, new(mockPlantRepo), new(mockCompRepo), new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("access point tenant mismatch returns ErrAccessPointNotFound", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		foreignAP := *baseAP
		foreignAP.TenantID = uuid.New()
		apRepo.On("GetByID", mock.Anything, apID).Return(&foreignAP, nil)

		uc := access.NewCheckInUseCase(apRepo, new(mockPlantRepo), new(mockCompRepo), new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, domainErrors.IsNotFound(err))
	})

	t.Run("plant lookup error is returned", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		expectedErr := errors.New("plant lookup error")
		plantRepo.On("GetByID", mock.Anything, plantID).Return(nil, expectedErr)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, new(mockCompRepo), new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("requirements retrieval error is returned", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		expectedErr := errors.New("requirements error")
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return(nil, expectedErr)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "failed to get document requirements")
	})

	t.Run("worker documents retrieval error is returned", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		expectedErr := errors.New("worker docs error")
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(nil, expectedErr)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "failed to get worker documents")
	})

	t.Run("contractor documents retrieval error is returned", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		expectedErr := errors.New("contractor docs error")
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return(nil, expectedErr)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, new(mockLogRepo))
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "failed to get contractor documents")
	})

	t.Run("inactive worker affiliation is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		inactiveAff := *baseAff
		inactiveAff.IsActive = false

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&inactiveAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "inactivo")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "inactivo")
	})

	t.Run("blocked contractor in plant is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		blockedAff := *baseAff
		blockedAff.PlantStatus = "blocked"

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&blockedAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "blocked")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "blocked")
	})

	t.Run("missing required document is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "faltante")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "faltante")
	})

	t.Run("unapproved document status is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		future := time.Now().Add(180 * 24 * time.Hour)
		pendingDoc := entity.DocumentWithTypeCode{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia DC-3",
			Status:           entity.DocumentStatusPendingReview,
			ExpiresAt:        &future,
		}

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{pendingDoc}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "no aprobado")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "no aprobado")
	})

	t.Run("document expiring within 30 days yields warning without fatal failure", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		soonExpiring := time.Now().Add(15 * 24 * time.Hour) // 15 days < 30 days
		expiringDoc := entity.DocumentWithTypeCode{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia DC-3",
			Status:           entity.DocumentStatusApproved,
			ExpiresAt:        &soonExpiring,
		}

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{expiringDoc}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultGrantedWithWarning
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultGrantedWithWarning, out.Result)
		assert.Nil(t, out.DenialReason)
	})

	t.Run("access log creation failure returns wrapped error", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		expectedErr := errors.New("disk full")
		logRepo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.Error(t, err)
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "failed to create access log")
	})

	t.Run("worker inactive via contractor_workers.is_active=false is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		inactiveAff := *baseAff
		inactiveAff.WorkerIsActive = false

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&inactiveAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "inactivo")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "inactivo")
	})

	t.Run("contractor inactive via contractors.is_active=false is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		inactiveContractorAff := *baseAff
		inactiveContractorAff.ContractorIsActive = false

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&inactiveContractorAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "inactiva")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "inactiva")
	})

	t.Run("contractor with pending plant_status is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		pendingAff := *baseAff
		pendingAff.PlantStatus = "pending"

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&pendingAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "pending")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "pending")
	})

	t.Run("contractor with suspended plant_status is denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		suspendedAff := *baseAff
		suspendedAff.PlantStatus = "suspended"

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&suspendedAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "suspended")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "suspended")
	})

	t.Run("missing required document is denied even when block_on_expired_docs is false", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		lenientPlant := *basePlant
		lenientPlant.BlockOnExpiredDocs = false // Non-negotiable missing doc must STILL be denied!

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(&lenientPlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "faltante")
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultDenied, out.Result)
		require.NotNil(t, out.DenialReason)
		assert.Contains(t, *out.DenialReason, "faltante")
	})

	t.Run("documents and document_snapshot arrays match in CheckInOutput", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{dc3Req}, nil)

		future := time.Now().Add(180 * 24 * time.Hour)
		workerDocs := []entity.DocumentWithTypeCode{
			{
				DocumentTypeCode: "DC3",
				DocumentName:     "Constancia DC-3",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &future,
			},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(workerDocs, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)
		logRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})
		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultGranted, out.Result)
		assert.NotEmpty(t, out.DocumentSnapshot)
		assert.Equal(t, out.DocumentSnapshot, out.Documents)
	})
}
