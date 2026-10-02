# Analysis & Technical Formulation: M2 Database, Config, Health Check & Access Points Catalog

- **Agent**: `teamwork_preview_explorer_m2_2`
- **Milestone**: M2: Backend Core & Authentication (`asesamo-api`)
- **Focus Area**: Infrastructure Configuration, PostgreSQL Connection Pool, Health Check Endpoints, and Access Points Catalog API

---

## 1. Executive Summary & Problem Boundary

In Milestone M2, `asesamo-api` establishes its production-grade foundation adhering strictly to Clean Architecture (`DEVELOPMENT_RULES.md` and `PROJECT.md`). This investigation formulates four critical infrastructure and domain components:

1. **`internal/infra/config/config.go`**: Environment configuration loader enforcing strict validation for required variables:
   - `DATABASE_URL`: Valid PostgreSQL connection string.
   - `APP_PORT`: Valid port (defaulting to `8080`).
   - `JWT_SECRET`: Cryptographically adequate key with minimum 32 characters (`DEVELOPMENT_RULES.md` line 193).
   - `APP_ENV`: Allowed environments (`development`, `staging`, `production`, `test`).
   - Security: Database URL credential masking for safe logging.
2. **`internal/infra/database/postgres.go`**: Connection pool initialization using modern PostgreSQL driver (`github.com/jackc/pgx/v5/stdlib` for `database/sql` compatibility or native `pgxpool`), ping verification, pool tuning limits, and graceful shutdown.
3. **Health Check Endpoints (`/health` and `/api/v1/health`)**: Dual-mount endpoints validating database liveness and returning HTTP 200 `{"status": "ok", "database": "connected", "version": "...", "timestamp": "..."}` matching both requirements and E2E test assertions (`test/e2e/tier1_feature_test.sh` lines 48-64), with 503 degraded fallback when the database is unreachable.
4. **Access Points Catalog API (`GET /api/v1/access-points`)**:
   - Repository layer filtering by `tenant_id` (`WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL`) enforcing strict multi-tenant isolation.
   - HTTP handler consuming `tenant.FromContext(ctx)`.
   - Response envelope `{ "data": [ { "id": "...", "plant_id": "...", "name": "Caseta Principal Norte", "is_active": true } ] }`.

---

## 2. Configuration Loader (`internal/infra/config/config.go`)

### 2.1 Design Rationale & Constraints
Per `DEVELOPMENT_RULES.md`:
- "Variables de entorno validadas al inicio de la aplicación (`cmd/api/main.go`). Si falta alguna requerida, el proceso termina con error claro." (Line 239)
- "El JWT se firma con secreto almacenado en variable de entorno `JWT_SECRET` (mínimo 32 chars)." (Line 193)
- "Cero secrets hardcodeados." (Line 237)
- "Nombres en `UPPER_SNAKE_CASE`." (Line 240)

To provide developer ergonomics and prevent multiple restart iterations, the loader collects all validation errors and reports them in a single comprehensive message.

### 2.2 Proposed Implementation

```go
// Package config handles environment-based configuration for asesamo-api.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all validated configuration values for the application.
type Config struct {
	AppEnv             string
	AppPort            string
	DatabaseURL        string
	JWTSecret          string
	CORSAllowedOrigins []string

	// Connection Pool Settings
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// App Metadata
	Version string
}

// Load reads and validates configuration from environment variables.
// If any required variable is missing or invalid, an aggregated error is returned.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CORSAllowedOrigins: parseCommaSeparated(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,http://web:3000")),
		DBMaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime:  getEnvDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),
		DBConnMaxIdleTime:  getEnvDuration("DB_CONN_MAX_IDLE_TIME", 15*time.Minute),
		Version:            getEnv("APP_VERSION", "1.0.0"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: validation failed: %w", err)
	}

	return cfg, nil
}

// Validate ensures all required configuration invariants are satisfied.
func (c *Config) Validate() error {
	var errs []string

	// 1. Validate APP_ENV
	switch c.AppEnv {
	case "development", "staging", "production", "test":
		// valid
	default:
		errs = append(errs, fmt.Sprintf("APP_ENV must be one of ['development', 'staging', 'production', 'test'], got %q", c.AppEnv))
	}

	// 2. Validate APP_PORT
	portNum, err := strconv.Atoi(c.AppPort)
	if err != nil || portNum < 1 || portNum > 65535 {
		errs = append(errs, fmt.Sprintf("APP_PORT must be a valid port number (1-65535), got %q", c.AppPort))
	}

	// 3. Validate DATABASE_URL
	if strings.TrimSpace(c.DatabaseURL) == "" {
		errs = append(errs, "DATABASE_URL is required and cannot be empty")
	} else {
		parsedURL, err := url.Parse(c.DatabaseURL)
		if err != nil || (parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql") {
			errs = append(errs, "DATABASE_URL must be a valid PostgreSQL connection URL (e.g. postgres://user:pass@host:5432/db)")
		}
	}

	// 4. Validate JWT_SECRET (Strictly >= 32 characters per DEVELOPMENT_RULES.md line 193)
	if strings.TrimSpace(c.JWTSecret) == "" {
		errs = append(errs, "JWT_SECRET is required and cannot be empty")
	} else if len(c.JWTSecret) < 32 {
		errs = append(errs, fmt.Sprintf("JWT_SECRET must be at least 32 characters long per security standards (current length: %d)", len(c.JWTSecret)))
	}

	// 5. Validate Database Pool Configuration
	if c.DBMaxOpenConns < 1 {
		errs = append(errs, fmt.Sprintf("DB_MAX_OPEN_CONNS must be >= 1, got %d", c.DBMaxOpenConns))
	}
	if c.DBMaxIdleConns < 0 {
		errs = append(errs, fmt.Sprintf("DB_MAX_IDLE_CONNS must be >= 0, got %d", c.DBMaxIdleConns))
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		errs = append(errs, fmt.Sprintf("DB_MAX_IDLE_CONNS (%d) cannot exceed DB_MAX_OPEN_CONNS (%d)", c.DBMaxIdleConns, c.DBMaxOpenConns))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// MaskedDatabaseURL returns the connection URL with the password redacted for secure logging.
func (c *Config) MaskedDatabaseURL() string {
	u, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return "[malformed database url]"
	}
	return u.Redacted()
}

// ServerAddress returns the address string suitable for http.Server.Addr (e.g. ":8080").
func (c *Config) ServerAddress() string {
	return fmt.Sprintf(":%s", c.AppPort)
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseCommaSeparated(val string) []string {
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
```

### 2.3 Proposed Unit Tests (`internal/infra/config/config_test.go`)
```go
package config_test

import (
	"os"
	"testing"
	"asesamo-api/internal/infra/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validation(t *testing.T) {
	t.Run("valid configuration passes", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("APP_PORT", "8080")
		t.Setenv("DATABASE_URL", "postgres://user:secret@localhost:5432/asesamo_db?sslmode=disable")
		t.Setenv("JWT_SECRET", "super_secret_jwt_key_min_32_characters_long_12345")

		cfg, err := config.Load()
		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.AppPort)
		assert.Equal(t, ":8080", cfg.ServerAddress())
		assert.Contains(t, cfg.MaskedDatabaseURL(), "user:xxxxx@localhost:5432")
	})

	t.Run("missing DATABASE_URL fails", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("APP_PORT", "8080")
		os.Unsetenv("DATABASE_URL")
		t.Setenv("JWT_SECRET", "super_secret_jwt_key_min_32_characters_long_12345")

		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
	})

	t.Run("JWT_SECRET less than 32 characters fails", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("APP_PORT", "8080")
		t.Setenv("DATABASE_URL", "postgres://user:secret@localhost:5432/asesamo_db?sslmode=disable")
		t.Setenv("JWT_SECRET", "short_secret_only_24_chars_")

		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters long")
	})

	t.Run("invalid APP_PORT fails", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		t.Setenv("APP_PORT", "invalid-port")
		t.Setenv("DATABASE_URL", "postgres://user:secret@localhost:5432/asesamo_db?sslmode=disable")
		t.Setenv("JWT_SECRET", "super_secret_jwt_key_min_32_characters_long_12345")

		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APP_PORT must be a valid port number")
	})

	t.Run("invalid APP_ENV fails", func(t *testing.T) {
		t.Setenv("APP_ENV", "invalid_env")
		t.Setenv("APP_PORT", "8080")
		t.Setenv("DATABASE_URL", "postgres://user:secret@localhost:5432/asesamo_db?sslmode=disable")
		t.Setenv("JWT_SECRET", "super_secret_jwt_key_min_32_characters_long_12345")

		_, err := config.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APP_ENV must be one of")
	})
}
```

---

## 3. PostgreSQL Connection Pool (`internal/infra/database/postgres.go`)

### 3.1 Driver Selection & Architectural Evaluation
Two options are evaluated:
1. **`database/sql` + `github.com/jackc/pgx/v5/stdlib`** (*Recommended*):
   - Standard Go interface: seamless integration with standard tools (`sqlc` standard database/sql generation, `sqlmock` for testing, `context.Context`).
   - High performance: Uses the modern `pgx/v5` engine under the hood, with full support for PostgreSQL 16 features (`citext`, `pgcrypto` binary representation).
   - Built-in connection pool management: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`, `SetConnMaxIdleTime`.
2. **`github.com/jackc/pgx/v5/pgxpool`**:
   - Native pgx pool: slightly lower overhead for heavy binary marshaling.
   - Requires pgx-specific interfaces (`pgx.Tx`, `pgxpool.Pool`).

We formulate the robust `database/sql` + `jackc/pgx/v5/stdlib` implementation as the primary solution, with interface compatibility so that repositories interact with standard `DBTX` (`*sql.DB`, `*sql.Tx`), ensuring zero coupling to driver internals.

### 3.2 Proposed Implementation

```go
// Package database provides PostgreSQL connection pooling and lifecycle management.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"asesamo-api/internal/infra/config"

	// Register pgx driver for database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps standard *sql.DB with convenience health check and lifecycle methods.
type DB struct {
	*sql.DB
}

// NewPostgresPool initializes a PostgreSQL connection pool, applies pool limits,
// and verifies connectivity with a 5-second timeout.
func NewPostgresPool(ctx context.Context, cfg *config.Config) (*DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open driver: %w", err)
	}

	// Apply connection pool limits
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	// Verify connectivity with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping failed on %s: %w", cfg.MaskedDatabaseURL(), err)
	}

	return &DB{DB: db}, nil
}

// Ping checks if the database is reachable with a short 2-second timeout.
func (d *DB) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.DB.PingContext(pingCtx)
}

// Close gracefully terminates all connections in the pool.
func (d *DB) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}
```

---

## 4. Health Check Endpoints (`/health` & `/api/v1/health`)

### 4.1 Specification & E2E Requirements
From `ORIGINAL_REQUEST.md` and `test/e2e/tier1_feature_test.sh` (lines 48-64):
- Mounted at both root `/health` and versioned `/api/v1/health`.
- Public endpoint (no JWT required).
- Returns HTTP 200 OK when database is reachable.
- Returns JSON body matching:
  - `status`: `"ok"` (asserted by TC 1.2: `(.data.status // .status) == "ok"`)
  - `database`: `"connected"` (asserted by TC 1.3: `(.data.database // .database // "connected") == "connected"`)
  - `version` / `timestamp`: asserted by TC 1.4: `(.data.version // .version // .data.timestamp // .timestamp) != null`
- Graceful degradation: Returns HTTP 503 Service Unavailable when the database is unreachable with `status: "degraded"` and `database: "disconnected"`.

### 4.2 Proposed Implementation

```go
// Package handler provides HTTP request handlers.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger defines the capability to check database liveness.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler serves liveness and readiness health checks.
type HealthHandler struct {
	pinger  Pinger
	version string
}

// NewHealthHandler constructs a HealthHandler.
func NewHealthHandler(pinger Pinger, version string) *HealthHandler {
	if version == "" {
		version = "1.0.0"
	}
	return &HealthHandler{
		pinger:  pinger,
		version: version,
	}
}

// HealthResponse represents the health status payload.
type HealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Error     string `json:"error,omitempty"`
}

// Check handles GET /health and GET /api/v1/health.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	resp := HealthResponse{
		Status:    "ok",
		Database:  "connected",
		Version:   h.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if h.pinger != nil {
		if err := h.pinger.Ping(r.Context()); err != nil {
			resp.Status = "degraded"
			resp.Database = "disconnected"
			resp.Error = "database ping failed"
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
```

### 4.3 Proposed Unit Tests (`internal/adapter/http/handler/health_handler_test.go`)
```go
package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthHandler_Check(t *testing.T) {
	t.Run("returns 200 and ok when database is connected", func(t *testing.T) {
		h := handler.NewHealthHandler(&mockPinger{err: nil}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.Check(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))

		var resp handler.HealthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.Equal(t, "connected", resp.Database)
		assert.Equal(t, "1.0.0", resp.Version)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("returns 503 and degraded when database ping fails", func(t *testing.T) {
		h := handler.NewHealthHandler(&mockPinger{err: errors.New("connection refused")}, "1.0.0")
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()

		h.Check(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		var resp handler.HealthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "degraded", resp.Status)
		assert.Equal(t, "disconnected", resp.Database)
		assert.NotEmpty(t, resp.Error)
	})
}
```

---

## 5. Access Points Catalog API (`GET /api/v1/access-points`)

### 5.1 Architecture & Flow
In compliance with Clean Architecture:
```
HTTP GET /api/v1/access-points
  → Chi Router
  → Auth Middleware (validates JWT, injects tenant_id into context)
  → AccessPointHandler
      → extracts tenantID via tenant.FromContext(ctx)
      → executes ListAccessPointsUseCase
          → calls AccessPointRepository.ListActiveByTenantID(ctx, tenantID)
              → executes SQL query: WHERE tenant_id = $1 AND is_active = TRUE AND deleted_at IS NULL
          ← returns []*entity.AccessPoint
      ← maps to DTO
  ← returns JSON { "data": [ ... ] }
```

### 5.2 Domain Entity (`internal/core/domain/entity/access_point.go`)
```go
package entity

import (
	"time"
	"github.com/google/uuid"
)

// AccessPoint represents an entry/exit checkpoint (caseta) within a plant.
type AccessPoint struct {
	ID          uuid.UUID  `json:"id"`
	PlantID     uuid.UUID  `json:"plant_id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
```

### 5.3 Repository Interface (`internal/core/domain/repository/access_point_repository.go`)
```go
package repository

import (
	"context"
	"asesamo-api/internal/core/domain/entity"
	"github.com/google/uuid"
)

// AccessPointRepository defines storage operations for access points.
type AccessPointRepository interface {
	// ListActiveByTenantID retrieves all active and non-deleted access points for a tenant.
	ListActiveByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error)
	// GetByID retrieves a specific access point with tenant isolation.
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessPoint, error)
}
```

### 5.4 Postgres Repository Adapter (`internal/adapter/postgres/access_point_repository.go`)
```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

type accessPointRepository struct {
	db *sql.DB
}

// NewAccessPointRepository instantiates an accessPointRepository backed by PostgreSQL.
func NewAccessPointRepository(db *sql.DB) repository.AccessPointRepository {
	return &accessPointRepository{db: db}
}

func (r *accessPointRepository) ListActiveByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error) {
	// Strict multi-tenant isolation: tenant_id = $1 is mandatory as the first condition.
	query := `
		SELECT 
			id,
			plant_id,
			tenant_id,
			name,
			description,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM access_points
		WHERE tenant_id = $1 
		  AND is_active = TRUE 
		  AND deleted_at IS NULL
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("postgres/access_point: query failed: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice so JSON serialization produces [] rather than null
	accessPoints := make([]*entity.AccessPoint, 0)

	for rows.Next() {
		var ap entity.AccessPoint
		var desc sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&ap.ID,
			&ap.PlantID,
			&ap.TenantID,
			&ap.Name,
			&desc,
			&ap.IsActive,
			&ap.CreatedAt,
			&ap.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres/access_point: scan failed: %w", err)
		}

		if desc.Valid {
			ap.Description = &desc.String
		}
		if deletedAt.Valid {
			ap.DeletedAt = &deletedAt.Time
		}

		accessPoints = append(accessPoints, &ap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres/access_point: rows iteration error: %w", err)
	}

	return accessPoints, nil
}

func (r *accessPointRepository) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessPoint, error) {
	query := `
		SELECT 
			id,
			plant_id,
			tenant_id,
			name,
			description,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM access_points
		WHERE tenant_id = $1 
		  AND id = $2 
		  AND deleted_at IS NULL
	`

	var ap entity.AccessPoint
	var desc sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, id).Scan(
		&ap.ID,
		&ap.PlantID,
		&ap.TenantID,
		&ap.Name,
		&desc,
		&ap.IsActive,
		&ap.CreatedAt,
		&ap.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres/access_point: get by id failed: %w", err)
	}

	if desc.Valid {
		ap.Description = &desc.String
	}
	if deletedAt.Valid {
		ap.DeletedAt = &deletedAt.Time
	}

	return &ap, nil
}
```

### 5.5 Use Case (`internal/core/usecase/access/list_access_points.go`)
```go
package access

import (
	"context"
	"errors"
	"fmt"

	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/domain/repository"
	"github.com/google/uuid"
)

// ListAccessPointsUseCase coordinates retrieving active access points for a tenant.
type ListAccessPointsUseCase struct {
	repo repository.AccessPointRepository
}

// NewListAccessPointsUseCase constructs the usecase.
func NewListAccessPointsUseCase(repo repository.AccessPointRepository) *ListAccessPointsUseCase {
	return &ListAccessPointsUseCase{repo: repo}
}

// Execute returns all active access points for the given tenant.
func (uc *ListAccessPointsUseCase) Execute(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("usecase/access: tenant id cannot be nil")
	}

	points, err := uc.repo.ListActiveByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("usecase/access: failed to list access points: %w", err)
	}

	return points, nil
}
```

### 5.6 DTOs & HTTP Envelopes (`internal/adapter/http/dto/access_point_dto.go`)
```go
package dto

// AccessPointResponse represents an access point in the API response.
type AccessPointResponse struct {
	ID          string  `json:"id"`
	PlantID     string  `json:"plant_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
}

// AccessPointListEnvelope defines the standard success envelope for access points.
type AccessPointListEnvelope struct {
	Data []AccessPointResponse `json:"data"`
}

// ErrorDetail represents the structured error response.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorEnvelope defines the standard error response envelope.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}
```

### 5.7 HTTP Handler (`internal/adapter/http/handler/access_point_handler.go`)
```go
package handler

import (
	"encoding/json"
	"net/http"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"
)

// AccessPointHandler handles HTTP operations for access points.
type AccessPointHandler struct {
	listUseCase *access.ListAccessPointsUseCase
}

// NewAccessPointHandler constructs an AccessPointHandler.
func NewAccessPointHandler(listUseCase *access.ListAccessPointsUseCase) *AccessPointHandler {
	return &AccessPointHandler{listUseCase: listUseCase}
}

// List handles GET /api/v1/access-points.
func (h *AccessPointHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 1. Extract tenant_id from context (injected by Auth/Tenant middleware)
	tenantID, err := tenant.FromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(dto.ErrorEnvelope{
			Error: dto.ErrorDetail{
				Code:    "UNAUTHORIZED",
				Message: "Token de autenticación no válido o contexto de empresa ausente.",
			},
		})
		return
	}

	// 2. Execute usecase
	points, err := h.listUseCase.Execute(r.Context(), tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(dto.ErrorEnvelope{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Error al consultar los puntos de acceso.",
			},
		})
		return
	}

	// 3. Map entities to response DTOs
	data := make([]dto.AccessPointResponse, 0, len(points))
	for _, p := range points {
		data = append(data, dto.AccessPointResponse{
			ID:          p.ID.String(),
			PlantID:     p.PlantID.String(),
			Name:        p.Name,
			Description: p.Description,
			IsActive:    p.IsActive,
		})
	}

	// 4. Return standard envelope
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto.AccessPointListEnvelope{
		Data: data,
	})
}
```

### 5.8 Proposed Unit Tests (`internal/adapter/http/handler/access_point_handler_test.go`)
```go
package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"asesamo-api/internal/adapter/http/dto"
	"asesamo-api/internal/adapter/http/handler"
	"asesamo-api/internal/core/domain/entity"
	"asesamo-api/internal/core/usecase/access"
	"asesamo-api/pkg/tenant"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAccessPointRepo struct {
	mock.Mock
}

func (m *mockAccessPointRepo) ListActiveByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID)
	if pts := args.Get(0); pts != nil {
		return pts.([]*entity.AccessPoint), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockAccessPointRepo) GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*entity.AccessPoint, error) {
	args := m.Called(ctx, tenantID, id)
	if pt := args.Get(0); pt != nil {
		return pt.(*entity.AccessPoint), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAccessPointHandler_List(t *testing.T) {
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	plantID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	pointID := uuid.MustParse("33333333-3333-3333-3333-333333333331")

	t.Run("returns 401 when tenant context is missing", func(t *testing.T) {
		repo := new(mockAccessPointRepo)
		uc := access.NewListAccessPointsUseCase(repo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		var errResp dto.ErrorEnvelope
		err := json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "UNAUTHORIZED", errResp.Error.Code)
	})

	t.Run("returns 200 with list of access points in envelope", func(t *testing.T) {
		repo := new(mockAccessPointRepo)
		expectedPoints := []*entity.AccessPoint{
			{
				ID:       pointID,
				PlantID:  plantID,
				TenantID: tenantID,
				Name:     "Caseta Principal Norte",
				IsActive: true,
			},
		}
		repo.On("ListActiveByTenantID", mock.Anything, tenantID).Return(expectedPoints, nil)

		uc := access.NewListAccessPointsUseCase(repo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp dto.AccessPointListEnvelope
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, pointID.String(), resp.Data[0].ID)
		assert.Equal(t, "Caseta Principal Norte", resp.Data[0].Name)
		assert.True(t, resp.Data[0].IsActive)
		repo.AssertExpectations(t)
	})

	t.Run("returns 200 with empty array when no access points exist", func(t *testing.T) {
		repo := new(mockAccessPointRepo)
		repo.On("ListActiveByTenantID", mock.Anything, tenantID).Return([]*entity.AccessPoint{}, nil)

		uc := access.NewListAccessPointsUseCase(repo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"data":[]`)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		repo := new(mockAccessPointRepo)
		repo.On("ListActiveByTenantID", mock.Anything, tenantID).Return(nil, errors.New("db disk failure"))

		uc := access.NewListAccessPointsUseCase(repo)
		h := handler.NewAccessPointHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/access-points", nil)
		ctx := tenant.WithTenant(req.Context(), tenantID)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		h.List(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		var errResp dto.ErrorEnvelope
		err := json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.NoError(t, err)
		assert.Equal(t, "INTERNAL_SERVER_ERROR", errResp.Error.Code)
	})
}
```

---

## 6. Wiring & Router Assembly (`internal/adapter/http/router.go`)

### 6.1 Router Layout & Middleware Ordering
In `cmd/api/main.go` and `internal/adapter/http/router.go`:
```go
package http

import (
	"net/http"

	"asesamo-api/internal/adapter/http/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type RouterDependencies struct {
	HealthHandler      *handler.HealthHandler
	AuthHandler        *handler.AuthHandler
	AccessPointHandler *handler.AccessPointHandler
	AuthMiddleware     func(http.Handler) http.Handler
	AllowedOrigins     []string
}

func NewRouter(deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	// Standard middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   deps.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 1. Root healthcheck endpoint
	r.Get("/health", deps.HealthHandler.Check)

	// 2. Versioned API routes (/api/v1)
	r.Route("/api/v1", func(v1 chi.Router) {
		// Public health check
		v1.Get("/health", deps.HealthHandler.Check)

		// Public authentication
		if deps.AuthHandler != nil {
			v1.Post("/auth/login", deps.AuthHandler.Login)
		}

		// Protected endpoints (require JWT & tenant context)
		v1.Group(func(protected chi.Router) {
			if deps.AuthMiddleware != nil {
				protected.Use(deps.AuthMiddleware)
			}
			if deps.AccessPointHandler != nil {
				protected.Get("/access-points", deps.AccessPointHandler.List)
			}
		})
	})

	return r
}
```

---

## 7. E2E & Boundary Test Compatibility Matrix

| Test Case | Method | Path | Required Assertions | Architectural Formulation |
|---|---|---|---|---|
| **TC 1.1** | GET | `/api/v1/health` & `/health` | Status 200 OK | Both mounted; responds with 200 |
| **TC 1.2** | GET | `/api/v1/health` | `(.data.status // .status) == "ok"` | Body has top-level `status: "ok"` |
| **TC 1.3** | GET | `/api/v1/health` | `(.data.database // .database) == "connected"` | Body has `database: "connected"` |
| **TC 1.4** | GET | `/api/v1/health` | Version/Timestamp != null | Body includes `version: "1.0.0"` & `timestamp: "<RFC3339>"` |
| **TC 1.5** | GET | `/api/v1/health` | Public access without auth header | Mounted outside auth middleware group |
| **TC 3.1** | GET | `/api/v1/access-points` | Authenticated request returns 200 OK | Auth middleware passes valid token; handler returns 200 |
| **TC 3.2** | GET | `/api/v1/access-points` | `.data | type == "array"` and `length > 0` | Returns `{ "data": [...] }` slice |
| **TC 3.3** | GET | `/api/v1/access-points` | Contains seed ID `33333333-3333-3333-3333-333333333331` | Queries DB for tenant `11111111-...` where active caseta exists |
| **TC 3.4** | GET | `/api/v1/access-points` | Unauthenticated request returns 401 | Auth middleware blocks request without header |
| **TC 3.5** | GET | `/api/v1/access-points` | Forged token returns 401 | Auth middleware rejects invalid signature |
| **TC 2.B1** | GET | `/api/v1/access-points` | Expired token returns 401 with error envelope | Auth middleware returns `{"error": {"code": "...", "message": "..."}}` |
| **TC 2.B3** | GET | `/api/v1/access-points` | Non-Bearer prefix returns 401 | Auth middleware requires `Bearer <token>` format |

---

## 8. Summary of Deliverables for Implementation Workers

The worker implementing M2 can directly use the specifications and code drafted above:
1. Create `internal/infra/config/config.go` with aggregated validation and URL masking.
2. Create `internal/infra/database/postgres.go` with `database/sql` + `jackc/pgx/v5/stdlib`, pool tuning, and ping checks.
3. Create `internal/adapter/http/handler/health_handler.go` with dual mounting at `/health` and `/api/v1/health`.
4. Create `internal/core/domain/entity/access_point.go`, `internal/core/domain/repository/access_point_repository.go`, `internal/adapter/postgres/access_point_repository.go`, `internal/core/usecase/access/list_access_points.go`, `internal/adapter/http/dto/access_point_dto.go`, and `internal/adapter/http/handler/access_point_handler.go`.
5. Run unit tests (`go test ./...`) achieving >= 80% coverage with zero IO dependencies.
