// Package repository handles data access layer for the inventory management system
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a PostgreSQL connection pool with utility methods
type DB struct {
	Pool *pgxpool.Pool
}

// Config contains database connection configuration
type Config struct {
	// DatabaseURL is the PostgreSQL connection string (e.g., postgres://user:pass@host:port/dbname)
	DatabaseURL string

	// MaxConnections is the maximum number of connections in the pool (default: 25)
	MaxConnections int32

	// MinConnections is the minimum number of connections to maintain (default: 5)
	MinConnections int32

	// MaxConnLifetime is the maximum lifetime of a connection (default: 1 hour)
	MaxConnLifetime time.Duration

	// MaxConnIdleTime is the maximum idle time before closing a connection (default: 30 minutes)
	MaxConnIdleTime time.Duration

	// HealthCheckPeriod is the interval for connection health checks (default: 1 minute)
	HealthCheckPeriod time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig(databaseURL string) *Config {
	return &Config{
		DatabaseURL:       databaseURL,
		MaxConnections:    25,
		MinConnections:    5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}
}

// NewDB creates a new database connection pool with the given configuration
func NewDB(ctx context.Context, cfg *Config) (*DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	// Parse the database URL and create pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connectivity with a ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping checks if the database is reachable
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics for monitoring
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
