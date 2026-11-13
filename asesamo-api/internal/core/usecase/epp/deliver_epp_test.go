package epp_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/epp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockEPPRepo struct {
	mock.Mock
}

func (m *MockEPPRepo) CreateItem(ctx context.Context, item *entity.EPPItem) error {
	return nil
}
func (m *MockEPPRepo) ListItemCatalog(ctx context.Context, tenantID uuid.UUID) ([]entity.EPPItem, error) {
	return nil, nil
}
func (m *MockEPPRepo) GetItemByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.EPPItem, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EPPItem), args.Error(1)
}
func (m *MockEPPRepo) RecordDelivery(ctx context.Context, delivery *entity.EPPDelivery) error {
	args := m.Called(ctx, delivery)
	return args.Error(0)
}
func (m *MockEPPRepo) ListDeliveries(ctx context.Context, tenantID uuid.UUID, filter repository.EPPDeliveryFilter) ([]entity.EPPDelivery, int, error) {
	return nil, 0, nil
}
func (m *MockEPPRepo) UpdateDeliveryStatus(ctx context.Context, tenantID, deliveryID uuid.UUID, status entity.EPPDeliveryStatus, notes string) error {
	return nil
}

func TestDeliverEPPUseCase_Execute(t *testing.T) {
	mockRepo := new(MockEPPRepo)
	uc := usecase.NewDeliverEPPUseCase(mockRepo)

	tenantID := uuid.New()
	workerID := uuid.New()
	delivererID := uuid.New()
	itemID := uuid.New()

	eppItem := &entity.EPPItem{
		ID:           itemID,
		TenantID:     tenantID,
		Code:         "EPP-ARN-001",
		Name:         "Arnés de Cuerpo Completo 4 Anillos",
		Category:     entity.EPPCategoryFallArrest,
		LifespanDays: 180,
		IsActive:     true,
	}

	mockRepo.On("GetItemByID", mock.Anything, tenantID, itemID).Return(eppItem, nil)
	mockRepo.On("RecordDelivery", mock.Anything, mock.MatchedBy(func(d *entity.EPPDelivery) bool {
		return d.WorkerID == workerID && d.EPPItemID == itemID && d.Status == entity.EPPDeliveryStatusDeliveredActive
	})).Return(nil)

	input := usecase.DeliverEPPInput{
		EPPItemID:     itemID,
		WorkerID:      workerID,
		DeliveryDate:  time.Now(),
		SignatureData: "data:image/png;base64,mock",
		Notes:         "Dotación para trabajo en alturas",
	}

	res, err := uc.Execute(context.Background(), tenantID, delivererID, input)
	require.NoError(t, err)
	assert.Equal(t, entity.EPPDeliveryStatusDeliveredActive, res.Status)
	assert.False(t, res.ExpirationDate.IsZero())

	mockRepo.AssertExpectations(t)
}
