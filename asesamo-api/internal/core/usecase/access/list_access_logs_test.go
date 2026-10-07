package access_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/usecase/access"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockListLogsRepo struct {
	mock.Mock
}

func (m *mockListLogsRepo) Create(ctx context.Context, log *entity.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockListLogsRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func (m *mockListLogsRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessLog), args.Error(1)
}

func (m *mockListLogsRepo) ListByPlant(ctx context.Context, tenantID, plantID uuid.UUID, limit, offset int) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, tenantID, plantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockListLogsRepo) ListWithFilters(ctx context.Context, filter repository.AccessLogFilter) ([]entity.AccessLog, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)
}

func (m *mockListLogsRepo) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
}

func TestListAccessLogsUseCase_Execute(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	now := time.Now().UTC()
	resultGranted := entity.AccessResultGranted

	sampleLogs := []entity.AccessLog{
		{
			ID:          uuid.New(),
			TenantID:    tenantID,
			PlantID:     plantID,
			UserID:      workerID,
			Result:      entity.AccessResultGranted,
			CheckedInAt: now,
			CreatedAt:   now,
		},
	}

	t.Run("list by plant successfully with pagination defaults", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.TenantID == tenantID && f.PlantID == plantID && f.Limit == 20 && f.Offset == 0
		})).Return(sampleLogs, 1, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  plantID,
			Page:     0, // normalized to 1
			PerPage:  0, // normalized to 20
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Len(t, out.Logs, 1)
		assert.Equal(t, 1, out.Total)
		assert.Equal(t, 1, out.Page)
		assert.Equal(t, 20, out.PerPage)
		logRepo.AssertExpectations(t)
	})

	t.Run("list with all filters: ap, worker, result, dates", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		from := now.Add(-24 * time.Hour)
		to := now

		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.TenantID == tenantID &&
				f.PlantID == plantID &&
				f.AccessPointID == apID &&
				f.WorkerUserID == workerID &&
				f.Result != nil && *f.Result == resultGranted &&
				f.From == &from &&
				f.To == &to &&
				f.Limit == 10 && f.Offset == 10
		})).Return(sampleLogs, 42, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID:      tenantID,
			PlantID:       plantID,
			AccessPointID: apID,
			WorkerUserID:  workerID,
			Result:        &resultGranted,
			From:          &from,
			To:            &to,
			Page:          2,
			PerPage:       10,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, 42, out.Total)
		assert.Equal(t, 2, out.Page)
		assert.Equal(t, 10, out.PerPage)
		logRepo.AssertExpectations(t)
	})

	t.Run("list by plant with per_page capped at 100", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.Limit == 100 && f.Offset == 100
		})).Return(sampleLogs, 105, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  plantID,
			Page:     2,
			PerPage:  150, // clamped to 100, offset = (2-1)*100 = 100
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, 105, out.Total)
		assert.Equal(t, 2, out.Page)
		assert.Equal(t, 100, out.PerPage)
		logRepo.AssertExpectations(t)
	})

	t.Run("list by tenant when plantID is Nil", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.PlantID == uuid.Nil && f.Limit == 20 && f.Offset == 0
		})).Return(sampleLogs, 1, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  uuid.Nil,
			Page:     1,
			PerPage:  20,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Len(t, out.Logs, 1)
		assert.Equal(t, 1, out.Total)
		logRepo.AssertExpectations(t)
	})

	t.Run("returns empty slice when logs is nil", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		logRepo.On("ListWithFilters", mock.Anything, mock.Anything).
			Return(([]entity.AccessLog)(nil), 0, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  plantID,
			Page:     1,
			PerPage:  20,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.NotNil(t, out.Logs)
		assert.Empty(t, out.Logs)
		assert.Equal(t, 0, out.Total)
		logRepo.AssertExpectations(t)
	})

	t.Run("error from repository is returned", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		expectedErr := errors.New("database timeout")
		logRepo.On("ListWithFilters", mock.Anything, mock.Anything).
			Return(([]entity.AccessLog)(nil), 0, expectedErr)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  plantID,
			Page:     1,
			PerPage:  20,
		})

		require.Error(t, err)
		assert.Nil(t, out)
		assert.Equal(t, expectedErr, err)
		logRepo.AssertExpectations(t)
	})
}
