package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"asesamo-api/internal/adapter/postgres"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkPermitRepository_Regression_EnumCastingInQueries statically verifies that all SQL statements
// manipulating custom PostgreSQL enum columns contain explicit type casts, preventing type inference errors.
func TestWorkPermitRepository_Regression_EnumCastingInQueries(t *testing.T) {
	content, err := os.ReadFile("work_permit_repository.go")
	require.NoError(t, err, "Must be able to read repository source code")
	src := string(content)

	t.Run("UpdateStatus query must explicitly cast status parameter to permit_status and disambiguate CASE", func(t *testing.T) {
		assert.True(t,
			strings.Contains(src, "SET status = $1::permit_status") ||
				strings.Contains(src, "status = $1::permit_status"),
			"UpdateStatus must cast $1 to ::permit_status",
		)
		assert.True(t,
			strings.Contains(src, "$1::text = 'approved_active'"),
			"UpdateStatus CASE clauses must explicitly disambiguate $1 comparison as ::text",
		)
		assert.True(t,
			strings.Contains(src, "$1::text = 'suspended'"),
			"UpdateStatus suspended CASE clause must cast $1::text",
		)
		assert.True(t,
			strings.Contains(src, "$1::text = 'closed'"),
			"UpdateStatus closed CASE clause must cast $1::text",
		)
		assert.True(t,
			strings.Contains(src, "$1::text = 'cancelled'"),
			"UpdateStatus cancelled CASE clause must cast $1::text",
		)
	})

	t.Run("Create query must explicitly cast permit_type and status parameters", func(t *testing.T) {
		assert.True(t,
			strings.Contains(src, "$5::permit_type") && strings.Contains(src, "$6::permit_status"),
			"Create query must cast $5::permit_type and $6::permit_status",
		)
	})

	t.Run("ListWithFilters must explicitly cast permit_status and permit_type in WHERE clause", func(t *testing.T) {
		assert.True(t,
			strings.Contains(src, "wp.status = $%d::permit_status"),
			"ListWithFilters must cast status filter to ::permit_status",
		)
		assert.True(t,
			strings.Contains(src, "wp.permit_type = $%d::permit_type"),
			"ListWithFilters must cast permit_type filter to ::permit_type",
		)
	})

	t.Run("AddSignature must explicitly cast role parameter to permit_signature_role", func(t *testing.T) {
		assert.True(t,
			strings.Contains(src, "$5::permit_signature_role"),
			"AddSignature query must cast $5::permit_signature_role",
		)
	})
}

// TestWorkPermitRepository_LiveDatabaseLifecycle executes the real PostgreSQL state machine
// and repository operations against the database when available.
func TestWorkPermitRepository_LiveDatabaseLifecycle(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "postgres://asesamo_admin:asesamo_secret_dev@localhost:5432/asesamo_db?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Skipf("Skipping live database test: failed to open driver: %v", err)
	}
	defer db.Close()

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelPing()
	if err := db.PingContext(ctxPing); err != nil {
		t.Skipf("Skipping live database test: database ping failed: %v", err)
	}

	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var plantID, contractorID, creatorID, approverID uuid.UUID

	err = db.QueryRowContext(ctxPing, "SELECT id FROM plants WHERE tenant_id = $1 LIMIT 1", tenantID).Scan(&plantID)
	if err != nil {
		t.Skipf("Skipping live database test: no plants found for tenant: %v", err)
	}
	err = db.QueryRowContext(ctxPing, "SELECT id FROM contractors WHERE is_active = true LIMIT 1").Scan(&contractorID)
	if err != nil {
		t.Skipf("Skipping live database test: no contractors found for tenant: %v", err)
	}
	err = db.QueryRowContext(ctxPing, "SELECT id FROM users WHERE tenant_id = $1 AND role = 'worker' LIMIT 1", tenantID).Scan(&creatorID)
	if err != nil {
		t.Skipf("Skipping live database test: no worker user found: %v", err)
	}
	err = db.QueryRowContext(ctxPing, "SELECT id FROM users WHERE tenant_id = $1 AND role IN ('plant_admin', 'super_admin') LIMIT 1", tenantID).Scan(&approverID)
	if err != nil {
		t.Skipf("Skipping live database test: no approver user found: %v", err)
	}

	repo := postgres.NewWorkPermitRepository(db)
	ctx := tenant.WithTenant(context.Background(), tenantID)

	permitID := uuid.New()
	permitNum := "PTAR-2026-" + uuid.New().String()[:8]
	now := time.Now().UTC()
	hazardDesc := "Riesgo de caída a distinto nivel (> 1.80m)"

	permit := &entity.WorkPermit{
		ID:                permitID,
		TenantID:          tenantID,
		PlantID:           plantID,
		PermitNumber:      permitNum,
		PermitType:        entity.PermitTypeHeights,
		Status:            entity.PermitStatusDraft,
		ContractorID:      contractorID,
		WorkArea:          "Área de Tanques de Mantenimiento",
		Description:       "Trabajo de pintura y soldadura en altura",
		HazardDescription: &hazardDesc,
		StartTime:         now,
		EndTime:           now.Add(4 * time.Hour),
		CreatedByUserID:   creatorID,
	}

	// Defer cleanup of test permit
	defer func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM work_permits WHERE id = $1", permitID)
	}()

	// 1. Create permit (Verifies $5::permit_type, $6::permit_status)
	err = repo.Create(ctx, permit)
	require.NoError(t, err, "Create permit must succeed without enum casting error")

	// 2. Transition to pending_approval (Verifies $1::permit_status)
	submitReason := "Completado por supervisor de contratista"
	err = repo.UpdateStatus(ctx, permitID, entity.PermitStatusPendingApproval, &submitReason, nil)
	require.NoError(t, err, "Transition to pending_approval must succeed")

	// 3. Transition to approved_active (REGRESSION GATE: verifies CASE WHEN $1::text = 'approved_active')
	err = repo.UpdateStatus(ctx, permitID, entity.PermitStatusApprovedActive, nil, &approverID)
	require.NoError(t, err, "Transition to approved_active must succeed without enum type deduction error")

	// 4. Verify in DB via GetByID
	loaded, err := repo.GetByID(ctx, permitID)
	require.NoError(t, err, "GetByID must succeed")
	assert.Equal(t, entity.PermitStatusApprovedActive, loaded.Status)
	assert.NotNil(t, loaded.ApprovedAt, "ApprovedAt must be populated on approved_active")
	assert.Equal(t, &approverID, loaded.ApprovedByUserID, "ApprovedByUserID must match authorizer")

	// 5. Add signature (Verifies $5::permit_signature_role)
	sigData := "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciPjwvc3ZnPg=="
	sigNotes := "Aprobado tras verificar cumplimiento de EPP y línea de vida"
	sig := &entity.PermitSignature{
		ID:            uuid.New(),
		TenantID:      tenantID,
		PermitID:      permitID,
		UserID:        approverID,
		Role:          entity.PermitSignerEHSAuthorizer,
		SignatureData: &sigData,
		Notes:         &sigNotes,
		SignedAt:      time.Now().UTC(),
	}
	err = repo.AddSignature(ctx, sig)
	require.NoError(t, err, "AddSignature must succeed with ::permit_signature_role cast")

	// 6. ListWithFilters (Verifies ::permit_status and ::permit_type in WHERE clause)
	statusFilter := entity.PermitStatusApprovedActive
	typeFilter := entity.PermitTypeHeights
	permits, total, err := repo.ListWithFilters(ctx, repository.PermitFilter{
		Status:     &statusFilter,
		PermitType: &typeFilter,
		Page:       1,
		PerPage:    10,
	})
	require.NoError(t, err, "ListWithFilters must succeed with enum filters")
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, permits)

	found := false
	for _, p := range permits {
		if p.ID == permitID {
			found = true
			break
		}
	}
	assert.True(t, found, "Newly created and approved permit must be found in filtered list")
}
