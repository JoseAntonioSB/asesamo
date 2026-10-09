package worker_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/usecase/worker"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID, lastLoginAt time.Time) error {
	args := m.Called(ctx, id, lastLoginAt)
	return args.Error(0)
}

func (m *mockUserRepo) UpdateThemePreference(ctx context.Context, id uuid.UUID, theme string) error {
	args := m.Called(ctx, id, theme)
	return args.Error(0)
}

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

type mockComplianceRepo struct{ mock.Mock }

func (m *mockComplianceRepo) GetWorkerAffiliation(ctx context.Context, plantID, workerUserID uuid.UUID) (*entity.WorkerAffiliation, error) {
	args := m.Called(ctx, plantID, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkerAffiliation), args.Error(1)
}

func (m *mockComplianceRepo) GetPlantDocumentRequirements(ctx context.Context, tenantID, plantID uuid.UUID) ([]entity.RequiredDocumentRule, error) {
	args := m.Called(ctx, tenantID, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.RequiredDocumentRule), args.Error(1)
}

func (m *mockComplianceRepo) GetWorkerDocuments(ctx context.Context, workerUserID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, workerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func (m *mockComplianceRepo) GetContractorDocuments(ctx context.Context, contractorID uuid.UUID) ([]entity.DocumentWithTypeCode, error) {
	args := m.Called(ctx, contractorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)
}

func TestGetCredentialUseCase_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := new(mockUserRepo)
	plantRepo := new(mockPlantRepo)
	compRepo := new(mockComplianceRepo)

	workerID := uuid.New()
	plantID := uuid.New()
	tenantID := uuid.New()
	contractorID := uuid.New()

	user := &entity.User{
		ID:       workerID,
		TenantID: &tenantID,
		FullName: "Carlos Mendoza",
		Role:     entity.RoleWorker,
		IsActive: true,
	}
	userRepo.On("GetByID", ctx, workerID).Return(user, nil)

	plant := &entity.Plant{
		ID:                 plantID,
		TenantID:           tenantID,
		Name:               "Planta Querétaro Norte",
		BlockOnExpiredDocs: true,
	}
	plantRepo.On("ListByTenant", ctx, tenantID).Return([]entity.Plant{*plant}, nil)

	curp := "MENC850101HDFR01"
	nss := "12345678901"
	job := "Soldador Industrial"
	aff := &entity.WorkerAffiliation{
		WorkerID:           workerID,
		FullName:           "Carlos Mendoza",
		CURP:               &curp,
		IMSSNumber:         &nss,
		JobTitle:           &job,
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       contractorID,
		ContractorName:     "Mantenimiento Industrial del Norte",
		PlantStatus:        "active",
	}
	compRepo.On("GetWorkerAffiliation", ctx, plantID, workerID).Return(aff, nil)

	rules := []entity.RequiredDocumentRule{
		{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia de Competencias Laborales DC-3",
			AppliesTo:        "worker",
			IsRequired:       true,
		},
	}
	compRepo.On("GetPlantDocumentRequirements", ctx, tenantID, plantID).Return(rules, nil)

	future := time.Now().Add(90 * 24 * time.Hour)
	workerDocs := []entity.DocumentWithTypeCode{
		{
			DocumentTypeCode: "DC3",
			DocumentName:     "Constancia de Competencias Laborales DC-3",
			Status:           entity.DocumentStatusApproved,
			ExpiresAt:        &future,
		},
	}
	compRepo.On("GetWorkerDocuments", ctx, workerID).Return(workerDocs, nil)
	compRepo.On("GetContractorDocuments", ctx, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

	uc := worker.NewGetCredentialUseCase(userRepo, plantRepo, compRepo)

	out, err := uc.Execute(ctx, worker.GetCredentialInput{
		WorkerUserID: workerID,
	})

	require.NoError(t, err)
	assert.Equal(t, workerID, out.Worker.UserID)
	assert.Equal(t, "Carlos Mendoza", out.Worker.FullName)
	assert.Equal(t, workerID.String(), out.QRPayload)
	assert.Equal(t, entity.AccessResultGranted, out.ComplianceStatus)
	assert.Empty(t, out.DenialReasons)
	assert.Len(t, out.Documents, 1)
}

func TestGetCredentialUseCase_NilWorkerFails(t *testing.T) {
	ctx := context.Background()
	uc := worker.NewGetCredentialUseCase(new(mockUserRepo), new(mockPlantRepo), new(mockComplianceRepo))

	_, err := uc.Execute(ctx, worker.GetCredentialInput{
		WorkerUserID: uuid.Nil,
	})

	require.Error(t, err)
	assert.Equal(t, domainErrors.ErrInvalidInput.Code, err.(*domainErrors.AppError).Code)
}
