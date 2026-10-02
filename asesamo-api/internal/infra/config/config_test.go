package config_test

import (
	"os"
	"testing"

	"asesamo-api/internal/infra/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setTestEnv(t *testing.T, env map[string]string) {
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoadFromEnv_Valid(t *testing.T) {
	setTestEnv(t, map[string]string{
		"DATABASE_URL":         "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
		"APP_PORT":             "8080",
		"JWT_SECRET":           "super_secret_jwt_key_min_32_characters_long_12345",
		"APP_ENV":              "development",
		"CORS_ALLOWED_ORIGINS": "http://localhost:3000,http://127.0.0.1:3000",
	})

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "postgres://user:pass@localhost:5432/dbname?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, 8080, cfg.AppPort)
	assert.Equal(t, "super_secret_jwt_key_min_32_characters_long_12345", cfg.JWTSecret)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, []string{"http://localhost:3000", "http://127.0.0.1:3000"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, "postgres://user:xxxxx@localhost:5432/dbname?sslmode=disable", cfg.MaskedDatabaseURL())
}

func TestLoadFromEnv_Defaults(t *testing.T) {
	// Unset everything, then set only mandatory
	os.Clearenv()
	setTestEnv(t, map[string]string{
		"DATABASE_URL": "postgresql://user:pass@localhost:5432/dbname",
		"JWT_SECRET":   "super_secret_jwt_key_min_32_characters_long_12345",
	})

	cfg, err := config.LoadFromEnv()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 8080, cfg.AppPort)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
}

func TestLoadFromEnv_InvalidPort(t *testing.T) {
	setTestEnv(t, map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/dbname",
		"APP_PORT":     "not-a-number",
		"JWT_SECRET":   "super_secret_jwt_key_min_32_characters_long_12345",
	})

	cfg, err := config.LoadFromEnv()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestValidate_Errors(t *testing.T) {
	t.Run("missing database url", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL: "",
			AppPort:     8080,
			JWTSecret:   "super_secret_jwt_key_min_32_characters_long_12345",
			AppEnv:      "development",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DATABASE_URL is required")
	})

	t.Run("invalid database url scheme", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL: "mysql://user:pass@localhost:3306/db",
			AppPort:     8080,
			JWTSecret:   "super_secret_jwt_key_min_32_characters_long_12345",
			AppEnv:      "development",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must start with postgres://")
	})

	t.Run("port out of range", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL: "postgres://localhost:5432/db",
			AppPort:     70000,
			JWTSecret:   "super_secret_jwt_key_min_32_characters_long_12345",
			AppEnv:      "development",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "APP_PORT must be between 1 and 65535")
	})

	t.Run("secret too short", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL: "postgres://localhost:5432/db",
			AppPort:     8080,
			JWTSecret:   "short_secret",
			AppEnv:      "development",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT_SECRET must be at least 32 characters")
	})

	t.Run("invalid environment", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseURL: "postgres://localhost:5432/db",
			AppPort:     8080,
			JWTSecret:   "super_secret_jwt_key_min_32_characters_long_12345",
			AppEnv:      "invalid_env",
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "APP_ENV must be one of")
	})
}

func TestMaskedDatabaseURL_EmptyOrInvalid(t *testing.T) {
	cfgEmpty := &config.Config{DatabaseURL: ""}
	assert.Equal(t, "", cfgEmpty.MaskedDatabaseURL())

	cfgInvalid := &config.Config{DatabaseURL: "::invalid-url"}
	assert.Equal(t, "<invalid database url>", cfgInvalid.MaskedDatabaseURL())
}
