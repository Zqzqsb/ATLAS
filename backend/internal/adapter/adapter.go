// Package adapter provides the database adapter interface and concrete implementations.
// This is the single source of truth for DBAdapter, DBConfig, QueryResult.
package adapter

import (
	"context"
)

// ============================================
// Core types — used across internal/ and server/
// ============================================

// DBAdapter 数据库适配器接口
type DBAdapter interface {
	Connect(ctx context.Context) error
	Close() error
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)
	GetDatabaseType() string
	GetDatabaseVersion(ctx context.Context) (string, error)
	DryRunSQL(ctx context.Context, sql string) (*QueryResult, error)
}

// DBConfig 数据库连接配置
type DBConfig struct {
	Type     string // "mysql", "mariadb", "postgresql", "sqlite"
	Host     string
	Port     int
	Database string
	User     string
	Password string
	FilePath string // SQLite 文件路径

	MaxOpenConns int
	MaxIdleConns int
}

// QueryResult 查询结果
type QueryResult struct {
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"row_count"`
	ExecutionTime int64                    `json:"execution_time"`
	Error         string                   `json:"error,omitempty"`
}

// AdapterFactory 适配器工厂函数类型
type AdapterFactory func(config *DBConfig) (DBAdapter, error)

// ============================================
// DatabaseType enum
// ============================================

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLite     DatabaseType = "sqlite"
)

// ============================================
// Factory
// ============================================

// NewAdapter 工厂函数：根据配置创建对应的适配器
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
	case "postgresql":
		return NewPostgreSQLAdapter(&PostgreSQLConfig{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     config.User,
			Password: config.Password,
		}), nil
	case "sqlite":
		return NewSQLiteAdapter(&SQLiteConfig{
			FilePath: config.FilePath,
		}), nil
	default:
		return nil, &UnsupportedDatabaseError{Type: config.Type}
	}
}

// UnsupportedDatabaseError 不支持的数据库类型错误
type UnsupportedDatabaseError struct {
	Type string
}

func (e *UnsupportedDatabaseError) Error() string {
	return "unsupported database type: " + e.Type
}
