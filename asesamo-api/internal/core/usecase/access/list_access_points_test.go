package access_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/access"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAccessPointRepository implements repository.AccessPointRepository for testing.
type MockAccessPointRepository struct {
	mock.Mock
}

func (m *MockAccessPointRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *MockAccessPointRepository) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func TestListAccessPointsUseCase(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	desc := "Caseta de vigilancia principal"

	testPoints := []entity.AccessPoint{
		{
			ID:          uuid.New(),
			PlantID:     plantID,
			TenantID:    tenantID,
			Name:        "Caseta Principal Norte",
			Description: &desc,
			IsActive:    true,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	t.Run("successful list retrieval", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepository)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(testPoints, nil)

		uc := access.NewListAccessPointsUseCase(mockRepo)
		result, err := uc.Execute(context.Background(), tenantID)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Caseta Principal Norte", result[0].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty repository returns empty slice instead of nil", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepository)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(([]entity.AccessPoint)(nil), nil)

		uc := access.NewListAccessPointsUseCase(mockRepo)
		result, err := uc.Execute(context.Background(), tenantID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nil tenantID returns ErrTenantNotFound", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepository)
		uc := access.NewListAccessPointsUseCase(mockRepo)

		result, err := uc.Execute(context.Background(), uuid.Nil)
		assert.ErrorIs(t, err, domainErrors.ErrTenantNotFound)
		assert.Nil(t, result)
	})

	t.Run("database error returns ErrDatabase", func(t *testing.T) {
		mockRepo := new(MockAccessPointRepository)
		mockRepo.On("ListByTenant", mock.Anything, tenantID).Return(([]entity.AccessPoint)(nil), errors.New("db connection failure"))

		uc := access.NewListAccessPointsUseCase(mockRepo)
		result, err := uc.Execute(context.Background(), tenantID)

		assert.ErrorIs(t, err, domainErrors.ErrDatabase)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
