package permit_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/internal/core/domain/service"
	"asesamo-api/internal/core/usecase/permit"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ----------------------------------------------------------------------------
// Mock Repositories
// ----------------------------------------------------------------------------

type mockPermitRepo struct {
	mock.Mock
}

func (m *mockPermitRepo) Create(ctx context.Context, p *entity.WorkPermit) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPermitRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.WorkPermit, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkPermit), args.Error(1)
}

func (m *mockPermitRepo) ListWithFilters(ctx context.Context, filter repository.PermitFilter) ([]entity.WorkPermit, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.WorkPermit), args.Int(1), args.Error(2)
}

func (m *mockPermitRepo) NextPermitNumber(ctx context.Context, tenantID uuid.UUID, year int) (string, error) {
	args := m.Called(ctx, tenantID, year)
	return args.String(0), args.Error(1)
}

func (m *mockPermitRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PermitStatus, reason *string, byUserID *uuid.UUID) error {
	return m.Called(ctx, id, status, reason, byUserID).Error(0)
}

func (m *mockPermitRepo) AssignWorker(ctx context.Context, worker *entity.PermitWorker) error {
	return m.Called(ctx, worker).Error(0)
}

func (m *mockPermitRepo) RemoveWorker(ctx context.Context, permitID, workerID uuid.UUID) error {
	return m.Called(ctx, permitID, workerID).Error(0)
}

func (m *mockPermitRepo) ListWorkersByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitWorker, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitWorker), args.Error(1)
}

func (m *mockPermitRepo) InitChecklistItems(ctx context.Context, permitID uuid.UUID, items []entity.PermitChecklistItem) error {
	return m.Called(ctx, permitID, items).Error(0)
}

func (m *mockPermitRepo) ListChecklistByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitChecklistItem, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitChecklistItem), args.Error(1)
}

func (m *mockPermitRepo) UpdateChecklistItem(ctx context.Context, permitID uuid.UUID, itemKey string, isVerified bool, verifiedBy *uuid.UUID, notes *string, measurements map[string]any) error {
	return m.Called(ctx, permitID, itemKey, isVerified, verifiedBy, notes, measurements).Error(0)
}

func (m *mockPermitRepo) AddSignature(ctx context.Context, sig *entity.PermitSignature) error {
	return m.Called(ctx, sig).Error(0)
}

func (m *mockPermitRepo) ListSignaturesByPermit(ctx context.Context, permitID uuid.UUID) ([]entity.PermitSignature, error) {
	args := m.Called(ctx, permitID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PermitSignature), args.Error(1)
}

type mockComplianceRepo struct {
	mock.Mock
}

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

// ----------------------------------------------------------------------------
// Test CreatePermitUseCase
// ----------------------------------------------------------------------------

func TestCreatePermitUseCase(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	userID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), userID), "contractor_admin")

	now := time.Now().UTC()
	startTime := now.Add(1 * time.Hour)
	endTime := now.Add(5 * time.Hour)

	t.Run("Success: creates draft permit with auto numbering and seeded checklist", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewCreatePermitUseCase(repo)

		repo.On("NextPermitNumber", mock.Anything, tenantID, startTime.Year()).Return("PTAR-2026-0001", nil)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.WorkPermit) bool {
			return p.PermitNumber == "PTAR-2026-0001" && p.Status == entity.PermitStatusDraft && p.PermitType == entity.PermitTypeHeights
		})).Return(nil)
		repo.On("InitChecklistItems", mock.Anything, mock.Anything, mock.MatchedBy(func(items []entity.PermitChecklistItem) bool {
			return len(items) == 5 // NOM-009 Heights checklist has 5 items
		})).Return(nil)
		repo.On("AddSignature", mock.Anything, mock.MatchedBy(func(s *entity.PermitSignature) bool {
			return s.Role == entity.PermitSignerRequester
		})).Return(nil)

		createdPermit := &entity.WorkPermit{
			ID:           uuid.New(),
			TenantID:     tenantID,
			PermitNumber: "PTAR-2026-0001",
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		repo.On("GetByID", mock.Anything, mock.Anything).Return(createdPermit, nil)

		res, err := uc.Execute(ctx, permit.CreatePermitInput{
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			WorkArea:     "Techumbre Nave 3",
			Description:  "Reparación de láminas traslúcidas",
			StartTime:    startTime,
			EndTime:      endTime,
		})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "PTAR-2026-0001", res.PermitNumber)
		repo.AssertExpectations(t)
	})

	t.Run("Validation: EndTime before StartTime returns ErrPermitTimeInvalid", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewCreatePermitUseCase(repo)

		_, err := uc.Execute(ctx, permit.CreatePermitInput{
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			WorkArea:     "Área 1",
			Description:  "Trabajo",
			StartTime:    endTime,
			EndTime:      startTime, // invalid
		})

		assert.ErrorIs(t, err, domainErrors.ErrPermitTimeInvalid)
	})
}

// ----------------------------------------------------------------------------
// Test AssignWorkersUseCase
// ----------------------------------------------------------------------------

func TestAssignWorkersUseCase(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	workerID := uuid.New()
	permitID := uuid.New()
	ctx := tenant.WithRole(tenant.WithTenant(context.Background(), tenantID), "contractor_admin")
	futureDate := time.Now().Add(30 * 24 * time.Hour)

	t.Run("Success: compliant worker assigned", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewAssignWorkersUseCase(permitRepo, engine)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Pérez",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		permitRepo.On("AssignWorker", mock.Anything, mock.MatchedBy(func(w *entity.PermitWorker) bool {
			return w.WorkerID == workerID && w.ComplianceStatus == "compliant"
		})).Return(nil)

		res, err := uc.Execute(ctx, permit.AssignWorkersInput{
			PermitID:  permitID,
			WorkerIDs: []uuid.UUID{workerID},
		})

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Failure: worker missing DC-3 is blocked with 422", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewAssignWorkersUseCase(permitRepo, engine)

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusDraft,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Sin DC3",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		_, err := uc.Execute(ctx, permit.AssignWorkersInput{
			PermitID:  permitID,
			WorkerIDs: []uuid.UUID{workerID},
		})

		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingDC3)
		assert.Contains(t, err.Error(), "El trabajador no cuenta con DC-3 vigente para trabajos en alturas (NOM-009)")
	})
}

// ----------------------------------------------------------------------------
// Test ApprovePermitUseCase
// ----------------------------------------------------------------------------

func TestApprovePermitUseCase(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	contractorID := uuid.New()
	permitID := uuid.New()
	workerID := uuid.New()
	ehsUserID := uuid.New()
	futureDate := time.Now().Add(30 * 24 * time.Hour)

	t.Run("Forbidden: contractor cannot authorize permit", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewApprovePermitUseCase(permitRepo, engine)

		contractorCtx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), uuid.New()), "contractor_admin")
		_, err := uc.Execute(contractorCtx, permit.ApprovePermitInput{PermitID: permitID})
		assert.ErrorIs(t, err, domainErrors.ErrForbidden)
	})

	t.Run("Blocked: no assigned workers", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewApprovePermitUseCase(permitRepo, engine)

		ehsCtx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), ehsUserID), "ehs_manager")

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{}, // empty
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		_, err := uc.Execute(ehsCtx, permit.ApprovePermitInput{PermitID: permitID})
		assert.ErrorIs(t, err, domainErrors.ErrPermitNoWorkers)
	})

	t.Run("Blocked: mandatory checklist item unverified", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewApprovePermitUseCase(permitRepo, engine)

		ehsCtx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), ehsUserID), "ehs_manager")

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
				{ItemKey: "chk_2", IsMandatory: true, IsVerified: false}, // incomplete!
			},
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		_, err := uc.Execute(ehsCtx, permit.ApprovePermitInput{PermitID: permitID})
		assert.ErrorIs(t, err, domainErrors.ErrChecklistIncomplete)
	})

	t.Run("Success: transitions to approved_active and logs EHS signature", func(t *testing.T) {
		permitRepo := new(mockPermitRepo)
		compRepo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(compRepo)
		uc := permit.NewApprovePermitUseCase(permitRepo, engine)

		ehsCtx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), ehsUserID), "ehs_manager")

		wp := &entity.WorkPermit{
			ID:           permitID,
			TenantID:     tenantID,
			PlantID:      plantID,
			ContractorID: contractorID,
			PermitType:   entity.PermitTypeHeights,
			Status:       entity.PermitStatusPendingApproval,
			Workers:      []entity.PermitWorker{{WorkerID: workerID}},
			Checklist: []entity.PermitChecklistItem{
				{ItemKey: "chk_1", IsMandatory: true, IsVerified: true},
			},
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()

		// Candado re-evaluation
		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Juan Pérez",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)
		docs := []entity.DocumentWithTypeCode{
			{DocumentTypeCode: "MEDICAL", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
			{DocumentTypeCode: "DC3_ALTURAS", Status: entity.DocumentStatusApproved, ExpiresAt: &futureDate},
		}
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		permitRepo.On("AssignWorker", mock.Anything, mock.Anything).Return(nil)
		permitRepo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &ehsUserID).Return(nil)
		permitRepo.On("AddSignature", mock.Anything, mock.MatchedBy(func(s *entity.PermitSignature) bool {
			return s.Role == entity.PermitSignerEHSAuthorizer && s.UserID == ehsUserID
		})).Return(nil)

		approvedPermit := &entity.WorkPermit{
			ID:     permitID,
			Status: entity.PermitStatusApprovedActive,
		}
		permitRepo.On("GetByID", mock.Anything, permitID).Return(approvedPermit, nil).Once()

		res, err := uc.Execute(ehsCtx, permit.ApprovePermitInput{PermitID: permitID})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, entity.PermitStatusApprovedActive, res.Status)
	})
}

// ----------------------------------------------------------------------------
// Test TransitionStatusUseCase
// ----------------------------------------------------------------------------

func TestTransitionStatusUseCase(t *testing.T) {
	tenantID := uuid.New()
	permitID := uuid.New()
	userID := uuid.New()
	ctx := tenant.WithRole(tenant.WithUserID(tenant.WithTenant(context.Background(), tenantID), userID), "ehs_manager")
	reason := "Tormenta eléctrica en la zona"

	t.Run("Suspend active permit with reason", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewTransitionStatusUseCase(repo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusApprovedActive,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()
		repo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusSuspended, &reason, &userID).Return(nil)

		suspendedPermit := &entity.WorkPermit{
			ID:     permitID,
			Status: entity.PermitStatusSuspended,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(suspendedPermit, nil).Once()

		res, err := uc.Execute(ctx, permit.TransitionStatusInput{
			PermitID: permitID,
			Action:   "suspend",
			Reason:   &reason,
		})

		assert.NoError(t, err)
		assert.Equal(t, entity.PermitStatusSuspended, res.Status)
	})

	t.Run("Suspend active permit without reason fails", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewTransitionStatusUseCase(repo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusApprovedActive,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		_, err := uc.Execute(ctx, permit.TransitionStatusInput{
			PermitID: permitID,
			Action:   "suspend",
			Reason:   nil, // empty
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Se requiere especificar el motivo de la suspensión")
	})

	t.Run("Resume suspended permit", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewTransitionStatusUseCase(repo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusSuspended,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(wp, nil).Once()
		repo.On("UpdateStatus", mock.Anything, permitID, entity.PermitStatusApprovedActive, mock.Anything, &userID).Return(nil)

		resumedPermit := &entity.WorkPermit{
			ID:     permitID,
			Status: entity.PermitStatusApprovedActive,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(resumedPermit, nil).Once()

		res, err := uc.Execute(ctx, permit.TransitionStatusInput{
			PermitID: permitID,
			Action:   "resume",
		})

		assert.NoError(t, err)
		assert.Equal(t, entity.PermitStatusApprovedActive, res.Status)
	})

	t.Run("Invalid transition: closed to approved_active fails", func(t *testing.T) {
		repo := new(mockPermitRepo)
		uc := permit.NewTransitionStatusUseCase(repo)

		wp := &entity.WorkPermit{
			ID:       permitID,
			TenantID: tenantID,
			Status:   entity.PermitStatusClosed,
		}
		repo.On("GetByID", mock.Anything, permitID).Return(wp, nil)

		_, err := uc.Execute(ctx, permit.TransitionStatusInput{
			PermitID: permitID,
			Action:   "resume",
		})

		assert.ErrorIs(t, err, domainErrors.ErrPermitInvalidState)
	})
}
