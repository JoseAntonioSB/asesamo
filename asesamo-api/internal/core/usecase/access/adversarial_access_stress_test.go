package access_test

import (
	"context"
	"strings"
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

// TestAdversarial_ComplianceAndAccessStress executes exhaustive adversarial challenge scenarios
// against the access verification engine, check-out calculations, and log filtering.
func TestAdversarial_ComplianceAndAccessStress(t *testing.T) {
	tenantID := uuid.New()
	plantID := uuid.New()
	apID := uuid.New()
	workerID := uuid.New()
	contractorID := uuid.New()

	baseAP := &entity.AccessPoint{
		ID:       apID,
		TenantID: tenantID,
		PlantID:  plantID,
		Name:     "Caseta Principal",
		IsActive: true,
	}

	basePlant := &entity.Plant{
		ID:                 plantID,
		TenantID:           tenantID,
		Name:               "Planta Toluca",
		BlockOnExpiredDocs: true,
	}

	baseAff := &entity.WorkerAffiliation{
		WorkerID:           workerID,
		FullName:           "Juan Pérez",
		IsActive:           true,
		UserIsActive:       true,
		WorkerIsActive:     true,
		ContractorIsActive: true,
		ContractorID:       contractorID,
		ContractorName:     "Servicios Especializados SA",
		PlantStatus:        "active",
	}

	reqDC3 := entity.RequiredDocumentRule{
		DocumentTypeID:   uuid.New(),
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia de Competencias Laborales DC-3",
		AppliesTo:        "worker",
		IsRequired:       true,
	}

	reqREPSE := entity.RequiredDocumentRule{
		DocumentTypeID:   uuid.New(),
		DocumentTypeCode: "REPSE",
		DocumentName:     "Registro REPSE Vigente",
		AppliesTo:        "contractor",
		IsRequired:       true,
	}

	reqOptional := entity.RequiredDocumentRule{
		DocumentTypeID:   uuid.New(),
		DocumentTypeCode: "CURSO_ALTURAS",
		DocumentName:     "Certificación en Alturas Opcional",
		AppliesTo:        "worker",
		IsRequired:       false,
	}

	validFuture := time.Now().UTC().AddDate(1, 0, 0)
	docDC3Valid := entity.DocumentWithTypeCode{
		ID:               uuid.New(),
		DocumentTypeCode: "DC3",
		DocumentName:     "Constancia DC-3",
		AppliesTo:        "worker",
		Status:           entity.DocumentStatusApproved,
		ExpiresAt:        &validFuture,
	}

	docREPSEValid := entity.DocumentWithTypeCode{
		ID:               uuid.New(),
		DocumentTypeCode: "REPSE",
		DocumentName:     "Registro REPSE",
		AppliesTo:        "contractor",
		Status:           entity.DocumentStatusApproved,
		ExpiresAt:        &validFuture,
	}

	// =========================================================================
	// 1. Check-In Compliance Scenarios
	// =========================================================================

	t.Run("1.1 Fully compliant worker -> granted", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3, reqREPSE, reqOptional}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{docDC3Valid}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{docREPSEValid}, nil)

		var capturedLog *entity.AccessLog
		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			capturedLog = l
			return l.Result == entity.AccessResultGranted && l.DenialReason == nil
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, entity.AccessResultGranted, out.Result)
		assert.Nil(t, out.DenialReason)
		require.NotNil(t, capturedLog)
		assert.Equal(t, entity.AccessResultGranted, capturedLog.Result)
		assert.Nil(t, capturedLog.DenialReason)

		// Audit snapshot validation
		assert.Len(t, out.DocumentSnapshot, 3)
		assert.Equal(t, out.DocumentSnapshot, out.Documents)
		assert.Equal(t, entity.DocumentSnapshot(out.DocumentSnapshot), capturedLog.DocumentSnapshot)
	})

	t.Run("1.2 Inactive worker variations -> denied", func(t *testing.T) {
		scenarios := []struct {
			name         string
			userActive   bool
			workerActive bool
			affActive    bool
		}{
			{name: "users.is_active = false, cw.is_active = true", userActive: false, workerActive: true, affActive: false},
			{name: "users.is_active = true, cw.is_active = false", userActive: true, workerActive: false, affActive: true},
			{name: "users.is_active = false, cw.is_active = false", userActive: false, workerActive: false, affActive: false},
			{name: "aff.is_active = false flag only", userActive: false, workerActive: false, affActive: false},
		}

		for _, sc := range scenarios {
			t.Run(sc.name, func(t *testing.T) {
				apRepo := new(mockAPRepo)
				plantRepo := new(mockPlantRepo)
				compRepo := new(mockCompRepo)
				logRepo := new(mockLogRepo)

				testAff := *baseAff
				testAff.UserIsActive = sc.userActive
				testAff.WorkerIsActive = sc.workerActive
				testAff.IsActive = sc.affActive

				apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
				plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
				compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&testAff, nil)
				compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
				compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{docDC3Valid}, nil)
				compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{docREPSEValid}, nil)

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
		}
	})

	t.Run("1.3 Unauthorized contractor variations -> denied", func(t *testing.T) {
		statuses := []struct {
			name               string
			contractorIsActive bool
			plantStatus        string
			expectedReason     string
		}{
			{name: "contractors.is_active = false", contractorIsActive: false, plantStatus: "active", expectedReason: "inactiva"},
			{name: "plant_contractors.status = pending", contractorIsActive: true, plantStatus: "pending", expectedReason: "pending"},
			{name: "plant_contractors.status = suspended", contractorIsActive: true, plantStatus: "suspended", expectedReason: "suspended"},
			{name: "plant_contractors.status = blocked", contractorIsActive: true, plantStatus: "blocked", expectedReason: "blocked"},
		}

		for _, st := range statuses {
			t.Run(st.name, func(t *testing.T) {
				apRepo := new(mockAPRepo)
				plantRepo := new(mockPlantRepo)
				compRepo := new(mockCompRepo)
				logRepo := new(mockLogRepo)

				testAff := *baseAff
				testAff.ContractorIsActive = st.contractorIsActive
				testAff.PlantStatus = st.plantStatus

				apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
				plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
				compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(&testAff, nil)
				compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
				compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{docDC3Valid}, nil)
				compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{docREPSEValid}, nil)

				logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
					return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, st.expectedReason)
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
				assert.Contains(t, *out.DenialReason, st.expectedReason)
			})
		}
	})

	t.Run("1.4 Missing required document -> denied with Spanish reason", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		// Plant requires DC3, but worker has no documents uploaded
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
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
		assert.Contains(t, *out.DenialReason, reqDC3.DocumentName)
	})

	t.Run("1.5 Expired required document when block_on_expired_docs = true -> denied", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		expiredTime := time.Now().UTC().AddDate(0, 0, -10)
		expiredDoc := docDC3Valid
		expiredDoc.ExpiresAt = &expiredTime

		strictPlant := *basePlant
		strictPlant.BlockOnExpiredDocs = true

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(&strictPlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{expiredDoc}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultDenied && l.DenialReason != nil && strings.Contains(*l.DenialReason, "vencido")
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
		assert.Contains(t, *out.DenialReason, "vencido")
	})

	t.Run("1.6 Expired required document when block_on_expired_docs = false -> granted_with_warning", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		expiredTime := time.Now().UTC().AddDate(0, 0, -5)
		expiredDoc := docDC3Valid
		expiredDoc.ExpiresAt = &expiredTime

		lenientPlant := *basePlant
		lenientPlant.BlockOnExpiredDocs = false

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(&lenientPlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{expiredDoc}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultGrantedWithWarning && l.DenialReason == nil
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

	t.Run("1.7 Expiring soon document (within 30 days) -> granted_with_warning", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		expiringSoonTime := time.Now().UTC().AddDate(0, 0, 15) // 15 days left
		expiringSoonDoc := docDC3Valid
		expiringSoonDoc.ExpiresAt = &expiringSoonTime

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{expiringSoonDoc}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{}, nil)

		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			return l.Result == entity.AccessResultGrantedWithWarning && l.DenialReason == nil
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

	t.Run("1.8 Document snapshot array contains valid entries and audit metadata", func(t *testing.T) {
		apRepo := new(mockAPRepo)
		plantRepo := new(mockPlantRepo)
		compRepo := new(mockCompRepo)
		logRepo := new(mockLogRepo)

		apRepo.On("GetByID", mock.Anything, apID).Return(baseAP, nil)
		plantRepo.On("GetByID", mock.Anything, plantID).Return(basePlant, nil)
		compRepo.On("GetWorkerAffiliation", mock.Anything, plantID, workerID).Return(baseAff, nil)
		compRepo.On("GetPlantDocumentRequirements", mock.Anything, tenantID, plantID).Return([]entity.RequiredDocumentRule{reqDC3, reqREPSE, reqOptional}, nil)
		compRepo.On("GetWorkerDocuments", mock.Anything, workerID).Return([]entity.DocumentWithTypeCode{docDC3Valid}, nil)
		compRepo.On("GetContractorDocuments", mock.Anything, contractorID).Return([]entity.DocumentWithTypeCode{docREPSEValid}, nil)

		var capturedLog *entity.AccessLog
		logRepo.On("Create", mock.Anything, mock.MatchedBy(func(l *entity.AccessLog) bool {
			capturedLog = l
			return true
		})).Return(nil)

		uc := access.NewCheckInUseCase(apRepo, plantRepo, compRepo, logRepo)
		out, err := uc.Execute(context.Background(), access.CheckInInput{
			TenantID:      tenantID,
			WorkerUserID:  workerID,
			AccessPointID: apID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		require.Len(t, out.DocumentSnapshot, 3)

		// Validate each item has required metadata
		for _, item := range out.DocumentSnapshot {
			assert.NotEmpty(t, item.DocumentTypeCode)
			assert.NotEmpty(t, item.DocumentName)
			assert.NotEmpty(t, item.AppliesTo)
			assert.NotEmpty(t, item.Status)
			if item.DocumentTypeCode == "CURSO_ALTURAS" {
				assert.Equal(t, "missing", item.Status)
				assert.False(t, item.IsValid)
				assert.False(t, item.IsRequired)
				require.NotNil(t, item.FailureReason)
			} else {
				assert.Equal(t, "approved", item.Status)
				assert.True(t, item.IsValid)
				assert.True(t, item.IsRequired)
				assert.Nil(t, item.FailureReason)
				require.NotNil(t, item.ExpiresAt)
			}
		}

		assert.Equal(t, out.DocumentSnapshot, out.Documents)
		assert.Equal(t, entity.DocumentSnapshot(out.DocumentSnapshot), capturedLog.DocumentSnapshot)
	})

	// =========================================================================
	// 2. Check-Out Logic Scenarios
	// =========================================================================

	t.Run("2.1 Check-out duration calculation in minutes", func(t *testing.T) {
		durations := []struct {
			name             string
			checkInOffset    time.Duration
			checkOutOffset   time.Duration
			expectedDuration int
		}{
			{name: "Checked in 120 minutes ago", checkInOffset: -120 * time.Minute, checkOutOffset: 0, expectedDuration: 120},
			{name: "Checked in 45m 30s ago (fraction truncated)", checkInOffset: -(45*time.Minute + 30*time.Second), checkOutOffset: 0, expectedDuration: 45},
			{name: "Checked in 30 seconds ago", checkInOffset: -30 * time.Second, checkOutOffset: 0, expectedDuration: 0},
			{name: "Clock drift: checked in in future", checkInOffset: 10 * time.Minute, checkOutOffset: 0, expectedDuration: 0},
		}

		for _, d := range durations {
			t.Run(d.name, func(t *testing.T) {
				logRepo := new(mockLogRepo)
				apRepo := new(mockAPRepo)

				now := time.Now().UTC()
				inTime := now.Add(d.checkInOffset)
				outTime := now.Add(d.checkOutOffset)

				mockLog := &entity.AccessLog{
					ID:           uuid.New(),
					TenantID:     tenantID,
					PlantID:      plantID,
					UserID:       workerID,
					CheckedInAt:  inTime,
					CheckedOutAt: &outTime,
					Result:       entity.AccessResultGranted,
				}

				logRepo.On("CheckOutLatest", mock.Anything, tenantID, plantID, workerID, mock.Anything).Return(mockLog, nil)

				uc := access.NewCheckOutUseCase(logRepo, apRepo)
				out, err := uc.Execute(context.Background(), access.CheckOutInput{
					TenantID:     tenantID,
					PlantID:      plantID,
					WorkerUserID: workerID,
				})

				require.NoError(t, err)
				require.NotNil(t, out)
				assert.Equal(t, d.expectedDuration, out.DurationMinutes)
			})
		}
	})

	t.Run("2.2 Optional plant ID resolution from access point", func(t *testing.T) {
		logRepo := new(mockLogRepo)
		apRepo := new(mockAPRepo)

		resolvedPlantID := uuid.New()
		apRepo.On("GetByID", mock.Anything, apID).Return(&entity.AccessPoint{
			ID:       apID,
			TenantID: tenantID,
			PlantID:  resolvedPlantID,
		}, nil)

		inTime := time.Now().UTC().Add(-60 * time.Minute)
		outTime := time.Now().UTC()

		// Notice: CheckOutLatest must receive resolvedPlantID, not uuid.Nil!
		logRepo.On("CheckOutLatest", mock.Anything, tenantID, resolvedPlantID, workerID, mock.Anything).Return(&entity.AccessLog{
			ID:           uuid.New(),
			TenantID:     tenantID,
			PlantID:      resolvedPlantID,
			UserID:       workerID,
			CheckedInAt:  inTime,
			CheckedOutAt: &outTime,
			Result:       entity.AccessResultGranted,
		}, nil)

		uc := access.NewCheckOutUseCase(logRepo, apRepo)
		out, err := uc.Execute(context.Background(), access.CheckOutInput{
			TenantID:      tenantID,
			PlantID:       uuid.Nil, // Omitted plant ID!
			AccessPointID: apID,     // Access point provided
			WorkerUserID:  workerID,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, 60, out.DurationMinutes)
		apRepo.AssertExpectations(t)
		logRepo.AssertExpectations(t)
	})

	// =========================================================================
	// 3. Access Logs Multi-Filter & Pagination Bounds
	// =========================================================================

	t.Run("3.1 Multi-filter parameter propagation", func(t *testing.T) {
		logRepo := new(mockLogRepo)

		filterPlantID := uuid.New()
		filterAPID := uuid.New()
		filterWorkerID := uuid.New()
		filterResult := entity.AccessResultDenied
		fromTime := time.Now().UTC().AddDate(0, 0, -7)
		toTime := time.Now().UTC()

		logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
			return f.TenantID == tenantID &&
				f.PlantID == filterPlantID &&
				f.AccessPointID == filterAPID &&
				f.WorkerUserID == filterWorkerID &&
				f.Result != nil && *f.Result == filterResult &&
				f.From != nil && f.From.Equal(fromTime) &&
				f.To != nil && f.To.Equal(toTime) &&
				f.Limit == 20 &&
				f.Offset == 0
		})).Return([]entity.AccessLog{
			{ID: uuid.New(), TenantID: tenantID, Result: entity.AccessResultDenied},
		}, 1, nil)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID:      tenantID,
			PlantID:       filterPlantID,
			AccessPointID: filterAPID,
			WorkerUserID:  filterWorkerID,
			Result:        &filterResult,
			From:          &fromTime,
			To:            &toTime,
			Page:          1,
			PerPage:       20,
		})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, 1, out.Total)
		assert.Len(t, out.Logs, 1)
		logRepo.AssertExpectations(t)
	})

	t.Run("3.2 Pagination bounds verification", func(t *testing.T) {
		boundsCases := []struct {
			name           string
			inputPage      int
			inputPerPage   int
			expectedLimit  int
			expectedOffset int
			expectedPage   int
			expectedPP     int
		}{
			{name: "Zero values default to Page 1, PerPage 20", inputPage: 0, inputPerPage: 0, expectedLimit: 20, expectedOffset: 0, expectedPage: 1, expectedPP: 20},
			{name: "Negative values default to Page 1, PerPage 20", inputPage: -5, inputPerPage: -10, expectedLimit: 20, expectedOffset: 0, expectedPage: 1, expectedPP: 20},
			{name: "PerPage > 100 is capped at 100", inputPage: 1, inputPerPage: 500, expectedLimit: 100, expectedOffset: 0, expectedPage: 1, expectedPP: 100},
			{name: "Page 4 with PerPage 25 calculates offset 75", inputPage: 4, inputPerPage: 25, expectedLimit: 25, expectedOffset: 75, expectedPage: 4, expectedPP: 25},
			{name: "Page 3 with PerPage 100 calculates offset 200", inputPage: 3, inputPerPage: 200, expectedLimit: 100, expectedOffset: 200, expectedPage: 3, expectedPP: 100},
		}

		for _, bc := range boundsCases {
			t.Run(bc.name, func(t *testing.T) {
				logRepo := new(mockLogRepo)

				logRepo.On("ListWithFilters", mock.Anything, mock.MatchedBy(func(f repository.AccessLogFilter) bool {
					return f.Limit == bc.expectedLimit && f.Offset == bc.expectedOffset
				})).Return([]entity.AccessLog{}, 0, nil)

				uc := access.NewListAccessLogsUseCase(logRepo)
				out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
					TenantID: tenantID,
					Page:     bc.inputPage,
					PerPage:  bc.inputPerPage,
				})

				require.NoError(t, err)
				require.NotNil(t, out)
				assert.Equal(t, bc.expectedPage, out.Page)
				assert.Equal(t, bc.expectedPP, out.PerPage)
				logRepo.AssertExpectations(t)
			})
		}
	})
}
