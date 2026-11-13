package evacuation_test

import (
	"context"
	"testing"

	"asesamo-api/internal/core/domain/entity"
	usecase "asesamo-api/internal/core/usecase/evacuation"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockEvacuationRepo struct {
	mock.Mock
}

func (m *MockEvacuationRepo) CreateMusterPoint(ctx context.Context, point *entity.MusterPoint) error {
	return nil
}
func (m *MockEvacuationRepo) ListMusterPoints(ctx context.Context, tenantID uuid.UUID) ([]entity.MusterPoint, error) {
	return nil, nil
}
func (m *MockEvacuationRepo) StartEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error {
	args := m.Called(ctx, evac)
	return args.Error(0)
}
func (m *MockEvacuationRepo) GetActiveEvacuation(ctx context.Context, tenantID uuid.UUID) (*entity.EmergencyEvacuation, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EmergencyEvacuation), args.Error(1)
}
func (m *MockEvacuationRepo) GetEvacuationByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EmergencyEvacuation, error) {
	return nil, nil
}
func (m *MockEvacuationRepo) ListEvacuations(ctx context.Context, tenantID uuid.UUID) ([]entity.EmergencyEvacuation, error) {
	return nil, nil
}
func (m *MockEvacuationRepo) UpdateRollCallCheckIn(ctx context.Context, evacID, workerID, musterPointID, checkedInBy uuid.UUID, notes string) error {
	return nil
}
func (m *MockEvacuationRepo) EndEvacuation(ctx context.Context, evac *entity.EmergencyEvacuation) error {
	return nil
}
func (m *MockEvacuationRepo) GetPresentWorkersInPlant(ctx context.Context, tenantID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func TestStartEvacuationUseCase_Execute(t *testing.T) {
	mockRepo := new(MockEvacuationRepo)
	uc := usecase.NewStartEvacuationUseCase(mockRepo)

	tenantID := uuid.New()
	initiatorID := uuid.New()
	worker1 := uuid.New()
	worker2 := uuid.New()

	mockRepo.On("GetActiveEvacuation", mock.Anything, tenantID).Return(nil, nil)
	mockRepo.On("GetPresentWorkersInPlant", mock.Anything, tenantID).Return([]uuid.UUID{worker1, worker2}, nil)
	mockRepo.On("StartEvacuation", mock.Anything, mock.MatchedBy(func(evac *entity.EmergencyEvacuation) bool {
		return evac.TotalHeadcount == 2 && evac.MissingCount == 2 && evac.Status == entity.EvacuationStatusActive
	})).Return(nil)

	input := usecase.StartEvacuationInput{
		Type:        entity.EvacuationTypeDrill,
		Description: "Simulacro General de Evacuación",
	}

	evac, err := uc.Execute(context.Background(), tenantID, initiatorID, input)
	require.NoError(t, err)
	assert.Equal(t, 2, evac.TotalHeadcount)
	assert.Equal(t, 2, evac.MissingCount)
	assert.Equal(t, entity.EvacuationStatusActive, evac.Status)

	mockRepo.AssertExpectations(t)
}
