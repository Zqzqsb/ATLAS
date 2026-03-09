// Package adapter provides the database adapter interface and concrete implementations.
// This is the single source of truth for DBAdapter, DBConfig, QueryResult.
package adapter

import (
	"context"
)

// ============================================
// Core types — used across internal/ and server/
// ============================================

// DBAdapter defines the interface for database operations.
type DBAdapter interface {
	Connect(ctx context.Context) error
	Close() error
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)
	GetDatabaseType() string
	GetDatabaseVersion(ctx context.Context) (string, error)
	DryRunSQL(ctx context.Context, sql string) (*QueryResult, error)
}

// DBConfig holds database connection parameters.
type DBConfig struct {
	Type     string // "mysql", "mariadb"
	Host     string
	Port     int
	Database string
	User     string
	Password string

	MaxOpenConns int
	MaxIdleConns int
}

// QueryResult holds the result of an SQL query execution.
type QueryResult struct {
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"row_count"`
	ExecutionTime int64                    `json:"execution_time"`
	Error         string                   `json:"error,omitempty"`
}

// AdapterFactory is a constructor function type for adapters.
type AdapterFactory func(config *DBConfig) (DBAdapter, error)

// ============================================
// DatabaseType enum
// ============================================

type DatabaseType string

const (
	MySQL DatabaseType = "mysql"
)

// ============================================
// Factory
// ============================================

// NewAdapter creates a DBAdapter based on the config type.
func NewAdapter(config *DBConfig) (DBAdapter, error) {
	switch config.Type {
	case "mysql", "mariadb":
		return NewMySQLAdapter(&MySQLConfig{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     config.User,
			Password: config.Password,
		}), nil
	default:
		return nil, &UnsupportedDatabaseError{Type: config.Type}
	}
}

// UnsupportedDatabaseError indicates an unsupported database type.
type UnsupportedDatabaseError struct {
	Type string
}

func (e *UnsupportedDatabaseError) Error() string {
	return "unsupported database type: " + e.Type
}
