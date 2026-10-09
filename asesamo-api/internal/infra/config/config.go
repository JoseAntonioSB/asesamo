// Package config loads and validates runtime environment variables for ASESAMO.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Allowed AppEnv values.
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// Config holds the validated application configuration.
type Config struct {
	DatabaseURL        string
	AppPort            int
	JWTSecret          string
	AppEnv             string
	CORSAllowedOrigins []string
	StoragePath        string
	StorageDir         string
}

// LoadFromEnv loads configuration from environment variables and validates all required fields.
func LoadFromEnv() (*Config, error) {
	portStr := os.Getenv("APP_PORT")
	port := 8080
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("config: invalid APP_PORT %q: must be an integer: %w", portStr, err)
		}
		port = p
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = EnvDevelopment
	}

	corsOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var corsOrigins []string
	if corsOriginsStr != "" {
		for _, o := range strings.Split(corsOriginsStr, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				corsOrigins = append(corsOrigins, trimmed)
			}
		}
	}
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"*"}
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = os.Getenv("STORAGE_PATH")
	}
	if storageDir == "" {
		storageDir = "./uploads"
	}

	cfg := &Config{
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		AppPort:            port,
		JWTSecret:          os.Getenv("JWT_SECRET"),
		AppEnv:             appEnv,
		CORSAllowedOrigins: corsOrigins,
		StoragePath:        storageDir,
		StorageDir:         storageDir,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifies that all configuration values adhere to project constraints.
func (c *Config) Validate() error {
	var errs []string

	// Validate DATABASE_URL
	if c.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is required")
	} else {
		parsed, err := url.Parse(c.DatabaseURL)
		if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
			errs = append(errs, fmt.Sprintf("DATABASE_URL must start with postgres:// or postgresql://: invalid URL %q", c.DatabaseURL))
		}
	}

	// Validate APP_PORT
	if c.AppPort < 1 || c.AppPort > 65535 {
		errs = append(errs, fmt.Sprintf("APP_PORT must be between 1 and 65535, got %d", c.AppPort))
	}

	// Validate JWT_SECRET (minimum 32 characters)
	if len(c.JWTSecret) < 32 {
		errs = append(errs, fmt.Sprintf("JWT_SECRET must be at least 32 characters long, got length %d", len(c.JWTSecret)))
	}

	// Validate APP_ENV
	switch c.AppEnv {
	case EnvDevelopment, EnvStaging, EnvProduction, EnvTest:
		// Valid
	default:
		errs = append(errs, fmt.Sprintf("APP_ENV must be one of ['development', 'staging', 'production', 'test'], got %q", c.AppEnv))
	}

	if len(errs) > 0 {
		return errors.New("config validation failed:\n - " + strings.Join(errs, "\n - "))
	}

	return nil
}

// MaskedDatabaseURL returns the DATABASE_URL with passwords redacted for safe logging.
func (c *Config) MaskedDatabaseURL() string {
	if c.DatabaseURL == "" {
		return ""
	}
	parsed, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return "<invalid database url>"
	}
	return parsed.Redacted()
}
