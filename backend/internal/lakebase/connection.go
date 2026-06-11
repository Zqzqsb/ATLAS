// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in ATLAS system.
package lakebase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config holds the configuration for lake-base storage connection
type Config struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "127.0.0.1",
		Port:            3310,
		User:            "root",
		Password:        "your_strong_password",
		Database:        "atlas",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 300 * time.Second,
	}
}

// Connection errors
var (
	ErrNotConnected     = errors.New("lakebase: not connected to database")
	ErrAlreadyConnected = errors.New("lakebase: already connected to database")
	ErrConnectionClosed = errors.New("lakebase: connection closed")
	ErrPingFailed       = errors.New("lakebase: ping failed")
)

// ConnectionPool manages database connections for lake-base storage
type ConnectionPool struct {
	config *Config
	db     *sql.DB
	mu     sync.RWMutex
	closed bool
}

// NewConnectionPool creates a new connection pool with the given configuration
func NewConnectionPool(cfg *Config) *ConnectionPool {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &ConnectionPool{
		config: cfg,
	}
}

// Connect establishes a connection to the database
func (p *ConnectionPool) Connect(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrConnectionClosed
	}

	if p.db != nil {
		return ErrAlreadyConnected
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		p.config.User,
		p.config.Password,
		p.config.Host,
		p.config.Port,
		p.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("lakebase: failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)

	// Ping to verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("%w: %v", ErrPingFailed, err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrConnectionClosed
	}

	p.closed = true

	if p.db != nil {
		err := p.db.Close()
		p.db = nil
		return err
	}

	return nil
}

// DB returns the underlying sql.DB instance
func (p *ConnectionPool) DB() (*sql.DB, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, ErrConnectionClosed
	}

	if p.db == nil {
		return nil, ErrNotConnected
	}

	return p.db, nil
}

// Ping verifies the database connection is still alive
func (p *ConnectionPool) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return ErrConnectionClosed
	}

	if p.db == nil {
		return ErrNotConnected
	}

	return p.db.PingContext(ctx)
}

// Stats returns database connection pool statistics
func (p *ConnectionPool) Stats() (sql.DBStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return sql.DBStats{}, ErrConnectionClosed
	}

	if p.db == nil {
		return sql.DBStats{}, ErrNotConnected
	}

	return p.db.Stats(), nil
}

// IsConnected returns whether the pool is connected
func (p *ConnectionPool) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db != nil && !p.closed
}

// GetConfig returns the pool's connection configuration (read-only).
func (p *ConnectionPool) GetConfig() *Config {
	return p.config
}

// ExecContext executes a query without returning any rows
func (p *ConnectionPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db, err := p.DB()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (p *ConnectionPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := p.DB()
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns at most one row
func (p *ConnectionPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db, err := p.DB()
	if err != nil {
		// Return a row that will produce an error when scanned
		return nil
	}
	return db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction
func (p *ConnectionPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	db, err := p.DB()
	if err != nil {
		return nil, err
	}
	return db.BeginTx(ctx, opts)
}

// WithTransaction executes a function within a transaction
func (p *ConnectionPool) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := p.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lakebase: failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("lakebase: transaction rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lakebase: failed to commit transaction: %w", err)
	}

	return nil
}

