// Package database manages PostgreSQL database connections using pgx/v5.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Pinger defines an interface for components capable of checking database connectivity.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DB wraps standard *sql.DB with connection pooling parameters and helpers.
type DB struct {
	*sql.DB
}

// Config specifies connection pool parameters.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPoolConfig returns standard production connection pool settings.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}
}

// NewPostgresPool opens and configures a PostgreSQL connection pool.
// It verifies connectivity with an initial 5-second ping.
func NewPostgresPool(databaseURL string, poolCfg PoolConfig) (*DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(poolCfg.MaxOpenConns)
	db.SetMaxIdleConns(poolCfg.MaxIdleConns)
	db.SetConnMaxLifetime(poolCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(poolCfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return &DB{DB: db}, nil
}

// Ping checks whether the database is responsive.
func (d *DB) Ping(ctx context.Context) error {
	if d == nil || d.DB == nil {
		return fmt.Errorf("database: connection is nil")
	}
	return d.PingContext(ctx)
}

// Close gracefully closes all database connections in the pool.
func (d *DB) Close() error {
	if d == nil || d.DB == nil {
		return nil
	}
	return d.DB.Close()
}
