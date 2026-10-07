# Analysis: M2 Remediation — Test Mock Compilation & Coverage

**Agent**: `teamwork_preview_explorer_m2_rem_1`  
**Milestone**: M2 Remediation (Backend Core & Authentication)  
**Date**: 2026-10-02T05:16:30Z  
**Target Repository**: `asesamo-api`  

---

## 1. Executive Summary

Milestone M2 was rejected at Iteration 1 primarily due to:
1. Compilation failure in `internal/core/usecase/access/check_in_test.go` lines 175, 217, 264, 285 because `*mockAPRepo` does not implement `repository.AccessPointRepository` (missing method `ListByPlant`).
2. Cascade failure in `make lint` caused by `go vet` failing on the same uncompilable test file.
3. Severe test coverage deficit in `internal/core/usecase/access`: `check_out.go` (0% coverage, no test file) and `list_access_logs.go` (0% coverage, no test file). The package as a whole fails the mandatory 80% statement coverage requirement stipulated in `DEVELOPMENT_RULES.md` Rule 9.
4. Latent runtime panic bugs discovered in `mockAPRepo`, `mockPlantRepo`, `mockCompRepo`, and `mockLogRepo`: slice return types lack `args.Get(0) == nil` checks, which causes runtime interface assertion panics when mocking error paths returning `nil`.

This analysis provides the complete architectural diagnosis, exact method implementations using `github.com/stretchr/testify/mock`, and turnkey code specifications for all affected test suites.

---

## 2. Issue 1: `check_in_test.go` Compilation Failure & `mockAPRepo`

### 2.1 Direct Observation & Root Cause
Executing `go test ./internal/core/usecase/access` terminates with exit code 1:
```text
# asesamo-api/internal/core/usecase/access_test [asesamo-api/internal/core/usecase/access.test]
internal/core/usecase/access/check_in_test.go:175:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:217:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:264:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
internal/core/usecase/access/check_in_test.go:285:34: cannot use apRepo (variable of type *mockAPRepo) as repository.AccessPointRepository value in argument to access.NewCheckInUseCase: *mockAPRepo does not implement repository.AccessPointRepository (missing method ListByPlant)
FAIL	asesamo-api/internal/core/usecase/access [build failed]
```

Executing `make lint` fails with make Error 1 because `go vet ./...` discovers the identical failure.

### 2.2 Interface Contract
In `asesamo-api/internal/core/domain/repository/access_point_repository.go:12-16`:
```go
type AccessPointRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error)
	ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error)
}
```

In `asesamo-api/internal/core/usecase/access/check_in_test.go:19-33`:
```go
type mockAPRepo struct{ mock.Mock }

func (m *mockAPRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}
```
`mockAPRepo` only implemented `GetByID` and `ListByTenant`, omitting `ListByPlant`.

### 2.3 Exact Method Implementation for `mockAPRepo`
```go
func (m *mockAPRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}
```

### 2.4 Critical Secondary Finding: Nil Type Assertion Hazards in Existing Mocks
In `check_in_test.go`, multiple mock methods perform unsafe direct type assertions without verifying if `args.Get(0) == nil`:
1. `mockAPRepo.ListByTenant`:
   `return args.Get(0).([]entity.AccessPoint), args.Error(1)`
   If mocked with `.Return(nil, err)`, `args.Get(0)` is `nil`, resulting in a runtime panic:
   `panic: interface conversion: interface {} is nil, not []entity.AccessPoint`
2. `mockPlantRepo.ListByTenant`:
   `return args.Get(0).([]entity.Plant), args.Error(1)`
3. `mockCompRepo.GetPlantDocumentRequirements`:
   `return args.Get(0).([]entity.RequiredDocumentRule), args.Error(1)`
4. `mockCompRepo.GetWorkerDocuments`:
   `return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)`
5. `mockCompRepo.GetContractorDocuments`:
   `return args.Get(0).([]entity.DocumentWithTypeCode), args.Error(1)`
6. `mockLogRepo.ListByTenant`:
   `return args.Get(0).([]entity.AccessLog), args.Error(1)`
7. `mockLogRepo.ListByPlant`:
   `return args.Get(0).([]entity.AccessLog), args.Int(1), args.Error(2)`

**Remediation Rule**: Every mock method returning a slice or pointer must guard against `nil` before casting:
```go
if args.Get(0) == nil {
    return nil, args.Error(1)
}
return args.Get(0).([]TargetType), args.Error(1)
```

---

## 3. Issue 2: Access Usecase Package Test Coverage Analysis

### 3.1 Inventory of Files in `internal/core/usecase/access/`
| File | Lines | Statements (Approx.) | Existing Tests | Current Test Coverage | Target Coverage |
|------|-------|----------------------|----------------|-----------------------|-----------------|
| `list_access_points.go` | 44 | ~8 | `list_access_points_test.go` | 100.0% | 100.0% |
| `check_out.go` | 64 | ~12 | None (`check_out_test.go` missing) | 0.0% | 100.0% |
| `list_access_logs.go` | 79 | ~20 | None (`list_access_logs_test.go` missing) | 0.0% | 100.0% |
| `check_in.go` | 309 | ~60 | `check_in_test.go` (4 tests only) | ~55% (untested branches) | > 95.0% |
| **Total Package** | **496** | **~100** | **4 tests in 1 file** | **< 35% (fails compilation)** | **> 95.0%** |

### 3.2 Coverage Gap Breakdown

#### A. `check_out.go` (0% Coverage)
Needs tests verifying:
1. `in.WorkerUserID == uuid.Nil`: returns `domainErrors.ErrInvalidInput` ("El identificador del trabajador es obligatorio.").
2. `CheckOutLatest` repository error: returns underlying error.
3. Successful checkout when `updatedLog.CheckedOutAt != nil`: returns correct `CheckOutOutput`.
4. Successful checkout when `updatedLog.CheckedOutAt == nil` (fallback to `now`): sets `CheckedOutAt` to `now`.

#### B. `list_access_logs.go` (0% Coverage)
Needs tests verifying:
1. Pagination defaults: `Page <= 0` defaults to 1, `PerPage <= 0` defaults to 20.
2. Pagination clamping: `PerPage > 100` clamps to 100 with accurate offset `(Page-1)*PerPage`.
3. Filter by plant: `PlantID != uuid.Nil` calls `ListByPlant` and returns `logs, total`.
4. Filter by tenant: `PlantID == uuid.Nil` calls `ListByTenant` and computes `total = len(logs)`.
5. Nil slice fallback: when repository returns `nil` logs, returns empty slice `[]entity.AccessLog{}` (not nil JSON).
6. Repository error handling: errors from `ListByPlant` and `ListByTenant` are returned.

#### C. `check_in.go` Missing Branches (~45% Uncovered)
Currently, `check_in_test.go` covers only 4 happy/semi-happy scenarios. The following critical execution paths are completely untested:
1. Input validation: `in.WorkerUserID == uuid.Nil` -> `ErrInvalidInput`.
2. Input validation: `in.AccessPointID == uuid.Nil` -> `ErrInvalidInput`.
3. Access Point lookup failure: `apRepo.GetByID` error -> propagated error.
4. Tenant isolation check: `ap.TenantID != in.TenantID` -> `ErrAccessPointNotFound`.
5. Plant lookup failure: `plantRepo.GetByID` error -> propagated error.
6. Plant document requirements retrieval failure -> wrapped error.
7. Worker documents retrieval failure -> wrapped error.
8. Contractor documents retrieval failure -> wrapped error.
9. Inactive worker affiliation (`!aff.IsActive`) -> access denied with `El perfil del trabajador se encuentra inactivo.`
10. Contractor blocked/suspended in plant (`aff.PlantStatus == "blocked" || "suspended"`) -> access denied.
11. Missing required document (`req.IsRequired` and doc not uploaded) -> access denied with failure reason.
12. Unapproved document status (`doc.Status != DocumentStatusApproved`, e.g., `pending` or `rejected`) -> access denied.
13. Document expiring within 30 days (`ExpiresAt.Before(warningThreshold)`) -> `AccessResultGrantedWithWarning`.
14. Access log creation failure: `logRepo.Create` error -> wrapped error.

---

## 4. Turnkey Patch & Code Proposals

### 4.1 Specification for `check_in_test.go`

#### Modifications:
1. Add `"errors"` and `"strings"` to imports.
2. Add `ListByPlant` to `mockAPRepo`.
3. Add nil guards to `mockAPRepo.ListByTenant`, `mockPlantRepo.ListByTenant`, `mockCompRepo.GetPlantDocumentRequirements`, `mockCompRepo.GetWorkerDocuments`, `mockCompRepo.GetContractorDocuments`, `mockLogRepo.ListByTenant`, and `mockLogRepo.ListByPlant`.
4. Add all 14 missing test subtests to `TestCheckInUseCase_Execute`.

#### Exact Implementation of `mockAPRepo`:
```go
// MockAccessPointRepository
type mockAPRepo struct{ mock.Mock }

func (m *mockAPRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}

func (m *mockAPRepo) ListByPlant(ctx context.Context, plantID uuid.UUID) ([]entity.AccessPoint, error) {
	args := m.Called(ctx, plantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.AccessPoint), args.Error(1)
}
```

---

### 4.2 Specification for `check_out_test.go` (New File)
Target: `asesamo-api/internal/core/usecase/access/check_out_test.go`

```go
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

func (m *mockCheckOutLogRepo) CheckOutLatest(ctx context.Context, tenantID, plantID, workerID uuid.UUID, checkedOutAt time.Time) (*entity.AccessLog, error) {
	args := m.Called(ctx, tenantID, plantID, workerID, checkedOutAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AccessLog), args.Error(1)
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
```

---

### 4.3 Specification for `list_access_logs_test.go` (New File)
Target: `asesamo-api/internal/core/usecase/access/list_access_logs_test.go`

```go
package access_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"asesamo-api/internal/core/domain/entity"
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
	workerID := uuid.New()
	now := time.Now().UTC()

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
		logRepo.On("ListByPlant", mock.Anything, tenantID, plantID, 20, 0).
			Return(sampleLogs, 1, nil)

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

	t.Run("list by plant with per_page capped at 100", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		logRepo.On("ListByPlant", mock.Anything, tenantID, plantID, 100, 100).
			Return(sampleLogs, 105, nil)

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
		logRepo.On("ListByTenant", mock.Anything, tenantID, 20, 0).
			Return(sampleLogs, nil)

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
		logRepo.On("ListByPlant", mock.Anything, tenantID, plantID, 20, 0).
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

	t.Run("error from ListByPlant is returned", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		expectedErr := errors.New("database timeout")
		logRepo.On("ListByPlant", mock.Anything, tenantID, plantID, 20, 0).
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

	t.Run("error from ListByTenant is returned", func(t *testing.T) {
		logRepo := new(mockListLogsRepo)
		expectedErr := errors.New("database query error")
		logRepo.On("ListByTenant", mock.Anything, tenantID, 20, 0).
			Return(([]entity.AccessLog)(nil), expectedErr)

		uc := access.NewListAccessLogsUseCase(logRepo)
		out, err := uc.Execute(context.Background(), access.ListAccessLogsInput{
			TenantID: tenantID,
			PlantID:  uuid.Nil,
			Page:     1,
			PerPage:  20,
		})

		require.Error(t, err)
		assert.Nil(t, out)
		assert.Equal(t, expectedErr, err)
		logRepo.AssertExpectations(t)
	})
}
```

---

## 5. Verification Method

Once the patch is applied by the worker:

1. **Usecase Package Compilation & Coverage**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race -cover ./internal/core/usecase/access/...
   ```
   **Expected Result**: All tests pass, 0 failures, statement coverage >= 80% (projected > 95%).

2. **Full Repository Test Suite**:
   ```bash
   cd /Users/jasbotello/devel/asesamo/asesamo-api
   go test -v -race -cover ./...
   ```
   **Expected Result**: All packages pass with 0 failures, 0 data races.

3. **Linter Execution**:
   ```bash
   cd /Users/jasbotello/devel/asesamo
   make lint
   ```
   **Expected Result**: Zero errors reported; exit code 0.
