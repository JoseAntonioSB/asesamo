package service_test

import (
	"context"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
	domainErrors "asesamo-api/internal/core/domain/errors"
	"asesamo-api/internal/core/domain/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestComplianceEngine_EvaluateWorker_Success(t *testing.T) {
	plantID := uuid.New()
	contractorID := uuid.New()
	workerID := uuid.New()
	futureDate := time.Now().Add(60 * 24 * time.Hour)

	tests := []struct {
		name       string
		permitType entity.PermitType
		dc3Code    string
		dc3Name    string
	}{
		{
			name:       "Heights with DC3_ALTURAS",
			permitType: entity.PermitTypeHeights,
			dc3Code:    "DC3_ALTURAS",
			dc3Name:    "DC-3 Alturas NOM-009",
		},
		{
			name:       "Confined Space with DC3_ESPACIOS_CONFINADOS",
			permitType: entity.PermitTypeConfinedSpace,
			dc3Code:    "DC3_ESPACIOS_CONFINADOS",
			dc3Name:    "DC-3 Espacios Confinados",
		},
		{
			name:       "Hot Work with DC3_CORTE_SOLDADURA",
			permitType: entity.PermitTypeHotWork,
			dc3Code:    "DC3_CORTE_SOLDADURA",
			dc3Name:    "DC-3 Corte y Soldadura",
		},
		{
			name:       "Electrical LOTO with DC3_LOTO",
			permitType: entity.PermitTypeElectricalLOTO,
			dc3Code:    "DC3_LOTO",
			dc3Name:    "DC-3 LOTO NOM-029",
		},
		{
			name:       "Heights with legacy generic DC3 filename matching alturas",
			permitType: entity.PermitTypeHeights,
			dc3Code:    "DC3",
			dc3Name:    "dc3_seguridad_alturas_2026.pdf",
		},
		{
			name:       "Generic DC3 fallback without conflict",
			permitType: entity.PermitTypeHeights,
			dc3Code:    "DC3",
			dc3Name:    "constancia_dc3_general.pdf",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(mockComplianceRepo)
			engine := service.NewComplianceEngine(repo)

			aff := &entity.WorkerAffiliation{
				WorkerID:           workerID,
				FullName:           "Juan Pérez",
				UserIsActive:       true,
				WorkerIsActive:     true,
				ContractorID:       contractorID,
				ContractorName:     "Mantenimiento Industrial SA",
				ContractorIsActive: true,
				PlantStatus:        "active",
			}
			repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

			docs := []entity.DocumentWithTypeCode{
				{
					ID:               uuid.New(),
					DocumentTypeCode: "MEDICAL",
					DocumentName:     "Certificado Medico 2026.pdf",
					Status:           entity.DocumentStatusApproved,
					ExpiresAt:        &futureDate,
				},
				{
					ID:               uuid.New(),
					DocumentTypeCode: tc.dc3Code,
					DocumentName:     tc.dc3Name,
					Status:           entity.DocumentStatusApproved,
					ExpiresAt:        &futureDate,
				},
			}
			repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

			snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, tc.permitType, workerID)
			assert.NoError(t, err)
			assert.NotNil(t, snapshot)
			assert.True(t, snapshot.IsCompliant)
			assert.Nil(t, snapshot.BlockReason)
			assert.Equal(t, "approved", snapshot.MedicalStatus)
			assert.Equal(t, "approved", snapshot.DC3Status)
			assert.Equal(t, "active", snapshot.ContractorStatus)
		})
	}
}

func TestComplianceEngine_EvaluateWorker_CandadoLocks(t *testing.T) {
	plantID := uuid.New()
	contractorID := uuid.New()
	workerID := uuid.New()
	futureDate := time.Now().Add(60 * 24 * time.Hour)
	pastDate := time.Now().Add(-24 * time.Hour)

	t.Run("Lock 1: Inactive Worker Profile", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Pedro Inactivo",
			UserIsActive:       false,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "El perfil del trabajador se encuentra inactivo")
	})

	t.Run("Lock 1: Wrong Contractor", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		otherContractorID := uuid.New()
		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Pedro Cruz",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       otherContractorID,
			ContractorName:     "Otra Empresa SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "El trabajador no pertenece a la empresa contratista asignada al permiso")
	})

	t.Run("Lock 1: Suspended Contractor in Plant", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Pedro Cruz",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Empresa Suspendida SA",
			ContractorIsActive: true,
			PlantStatus:        "suspended",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrContractorInactive)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "no se encuentra activa o está suspendida en esta planta")
	})

	t.Run("Lock 2: Missing Medical Certificate", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Carlos Sin Medico",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingMedical)
		assert.False(t, snapshot.IsCompliant)
		assert.Equal(t, "missing", snapshot.MedicalStatus)
		assert.Contains(t, *snapshot.BlockReason, "El trabajador no cuenta con certificado médico de aptitud (documento faltante)")
	})

	t.Run("Lock 2: Expired Medical Certificate", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Carlos Vencido",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico_viejo.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastDate,
			},
			{
				ID:               uuid.New(),
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingMedical)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "El certificado médico de aptitud se encuentra vencido desde")
	})

	t.Run("Lock 2: Rejected Medical Certificate", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Carlos Rechazado",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusRejected,
				ExpiresAt:        &futureDate,
			},
			{
				ID:               uuid.New(),
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "El certificado médico de aptitud no está aprobado (estatus: rejected)")
	})

	t.Run("Lock 3: Missing Hazard DC-3 (Heights)", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Mario Gomez",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHeights, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingDC3)
		assert.False(t, snapshot.IsCompliant)
		assert.Equal(t, "missing", snapshot.DC3Status)
		assert.Contains(t, *snapshot.BlockReason, "El trabajador no cuenta con DC-3 vigente para trabajos en alturas (NOM-009)")
	})

	t.Run("Lock 3: Conflicting Hazard DC-3 (Has Alturas DC3 but work is Confined Space)", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Mario Gomez",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
			{
				ID:               uuid.New(),
				DocumentTypeCode: "DC3_ALTURAS",
				DocumentName:     "dc3_alturas.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeConfinedSpace, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingDC3)
		assert.False(t, snapshot.IsCompliant)
		assert.Equal(t, "missing", snapshot.DC3Status)
		assert.Contains(t, *snapshot.BlockReason, "El trabajador no cuenta con DC-3 vigente para espacios confinados (NOM-033)")
	})

	t.Run("Lock 3: Expired Hazard DC-3", func(t *testing.T) {
		repo := new(mockComplianceRepo)
		engine := service.NewComplianceEngine(repo)

		aff := &entity.WorkerAffiliation{
			WorkerID:           workerID,
			FullName:           "Mario Gomez",
			UserIsActive:       true,
			WorkerIsActive:     true,
			ContractorID:       contractorID,
			ContractorName:     "Contratista SA",
			ContractorIsActive: true,
			PlantStatus:        "active",
		}
		repo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(aff, nil)

		docs := []entity.DocumentWithTypeCode{
			{
				ID:               uuid.New(),
				DocumentTypeCode: "MEDICAL",
				DocumentName:     "medico.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &futureDate,
			},
			{
				ID:               uuid.New(),
				DocumentTypeCode: "DC3_CORTE_SOLDADURA",
				DocumentName:     "dc3_soldadura.pdf",
				Status:           entity.DocumentStatusApproved,
				ExpiresAt:        &pastDate,
			},
		}
		repo.On("GetWorkerDocuments", mock.Anything, workerID).Return(docs, nil)

		snapshot, err := engine.EvaluateWorker(context.Background(), plantID, contractorID, entity.PermitTypeHotWork, workerID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domainErrors.ErrWorkerMissingDC3)
		assert.False(t, snapshot.IsCompliant)
		assert.Contains(t, *snapshot.BlockReason, "La constancia DC-3 para este riesgo se encuentra vencida desde")
	})
}
