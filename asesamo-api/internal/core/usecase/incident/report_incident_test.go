package incident_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	usecase "asesamo-api/internal/core/usecase/incident"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSafetyIncidentRepo struct {
	mock.Mock
}

func (m *MockSafetyIncidentRepo) CreateIncident(ctx context.Context, incident *entity.SafetyIncident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}
func (m *MockSafetyIncidentRepo) GetIncidentByID(ctx context.Context, tenantID, id uuid.UUID) (*entity.SafetyIncident, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.SafetyIncident), args.Error(1)
}
func (m *MockSafetyIncidentRepo) ListIncidents(ctx context.Context, tenantID uuid.UUID, filter repository.IncidentFilter) ([]entity.SafetyIncident, int, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]entity.SafetyIncident), args.Int(1), args.Error(2)
}
func (m *MockSafetyIncidentRepo) UpdateInvestigation(ctx context.Context, incident *entity.SafetyIncident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}
func (m *MockSafetyIncidentRepo) GetMetrics(ctx context.Context, tenantID uuid.UUID) (*repository.SafetyIncidentMetrics, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.SafetyIncidentMetrics), args.Error(1)
}

func TestReportIncidentUseCase_Execute(t *testing.T) {
	mockRepo := new(MockSafetyIncidentRepo)
	uc := usecase.NewReportIncidentUseCase(mockRepo)

	tenantID := uuid.New()
	reporterID := uuid.New()

	input := usecase.ReportIncidentInput{
		Type:           entity.IncidentTypeNearMiss,
		Severity:       entity.IncidentSeverityModerate,
		Title:          "Derrame menor de solvente",
		Description:    "Goteo de tambo de 200L en almacén temporal",
		Location:       "Almacén Químico",
		EventTimestamp: time.Now(),
	}

	mockRepo.On("CreateIncident", mock.Anything, mock.MatchedBy(func(inc *entity.SafetyIncident) bool {
		return inc.Title == input.Title && inc.Status == entity.IncidentStatusReported && inc.TenantID == tenantID
	})).Return(nil)

	res, err := uc.Execute(context.Background(), tenantID, reporterID, input)
	require.NoError(t, err)
	assert.NotEmpty(t, res.IncidentNumber)
	assert.Equal(t, entity.IncidentStatusReported, res.Status)

	mockRepo.AssertExpectations(t)
}
