package access_test

import (
	"context"
	"errors"
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

type mockCheckOutLogRepo struct {
	mock.Mock
}

func (m *mockCheckOutLogRepo) Create(ctx context.Context, log *entity.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockCheckOutLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func (m *mockCheckOutLogRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessLog), args.Error(1)
}

func (m *mockCheckOutLogRepo) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, tenantID, plantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockCheckOutLogRepo) ListWithFilters(ctx context.Context, filter repository.AccessLogFilter) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockCheckOutLogRepo) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

type mockCheckOutAPRepo struct {
	mock.Mock
}

func (m *mockCheckOutAPRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *mockCheckOutAPRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *mockCheckOutAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func TestCheckOutUseCase_Execute(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	logID := uuid.New()
	checkInTime := time.Now().UTC().Add(-2 * time.Hour)
	checkOutTime := time.Now().UTC()

	t.Run("successful check-out with existing checked_out_at", func(t *testing.T) {
		logRepo := new(mockCheckOutLogRepo)
		expectedLog := &entity.AccessLog{
			ID:           logID,
			TenantID:     tenantID,
			PlantID:      plantID,
			UserID:       workerID,
			Result:       entity.AccessResultGranted,
			CheckedInAt:  checkInTime,
			CheckedOutAt: &checkOutTime,
		}

		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(expectedLog, nil)

		uc := access.NewCheckOutUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       plantID,
			AccessPointID: apID,
			WorkerUserID:  workerID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, logID, out.AccessLogID)
		assert.Equal(t, entity.AccessResultGranted, out.Result)
		assert.Equal(t, checkInTime, out.CheckedInAt)
		assert.Equal(t, checkOutTime, out.CheckedOutAt)
		assert.Equal(t, 120, out.DurationMinutes)
		logRepo.AssertExpectations(t)
	})

	t.Run("resolves plant_id from access point repository when plant_id is omitted", func(t *testing.T) {
		logRepo := new(mockCheckOutLogRepo)
		apRepo := new(mockCheckOutAPRepo)
		expectedAP := &entity.AccessPoint{
			ID:       apID,
			TenantID: tenantID,
			PlantID:  plantID,
			Name:     "Caseta Norte",
		}
		expectedLog := &entity.AccessLog{
			ID:           logID,
			TenantID:     tenantID,
			PlantID:      plantID,
			UserID:       workerID,
			Result:       entity.AccessResultGranted,
			CheckedInAt:  checkInTime,
			CheckedOutAt: &checkOutTime,
		}

		apRepo.On("GetByID", mock.Anything, apID).Return(expectedAP, nil)
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(expectedLog, nil)

		uc := access.NewCheckOutUseCase(logRepo, apRepo)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       uuid.Nil, // Omitted plant_id
			AccessPointID: apID,
			WorkerUserID:  workerID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, logID, out.AccessLogID)
		assert.Equal(t, 120, out.DurationMinutes)
		apRepo.AssertExpectations(t)
		logRepo.AssertExpectations(t)
	})

	t.Run("successful check-out with nil checked_out_at fallback to now", func(t *testing.T) {
		logRepo := new(mockCheckOutLogRepo)
		expectedLog := &entity.AccessLog{
			ID:           logID,
			TenantID:     tenantID,
			PlantID:      plantID,
			UserID:       workerID,
			Result:       entity.AccessResultGranted,
			CheckedInAt:  checkInTime,
			CheckedOutAt: nil,
		}

		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(expectedLog, nil)

		uc := access.NewCheckOutUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       plantID,
			AccessPointID: apID,
			WorkerUserID:  workerID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, logID, out.AccessLogID)
		assert.False(t, out.CheckedOutAt.IsZero())
		logRepo.AssertExpectations(t)
	})

	t.Run("nil worker_user_id returns ErrInvalidInput", func(t *testing.T) {
		logRepo := new(mockCheckOutLogRepo)
		uc := access.NewCheckOutUseCase(logRepo)

		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       plantID,
			AccessPointID: apID,
			WorkerUserID:  uuid.Nil,
		})

		require.Error(t, err)
		assert.Nil(t, out)
		assert.True(t, domainErrors.IsInvalidInput(err))
	})

	t.Run("repository error is returned", func(t *testing.T) {
		logRepo := new(mockCheckOutLogRepo)
		expectedErr := errors.New("database connection failed")
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.AnythingOfType("time.Time")).
			Return(nil, expectedErr)

		uc := access.NewCheckOutUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       plantID,
			AccessPointID: apID,
			WorkerUserID:  workerID,
		})

		require.Error(t, err)
		assert.Nil(t, out)
		assert.Equal(t, expectedErr, err)
		logRepo.AssertExpectations(t)
	})
}
