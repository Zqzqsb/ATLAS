// Package interfaces defines shared types used across internal/ and server/ layers.
// Only database adapter abstractions live here — they are the true cross-layer contract.
package interfaces

import (
	"context"
)

// ============================================
// Database Adapter — the only cross-layer contract
// ============================================

// DBAdapter 数据库适配器接口
type DBAdapter interface {
	Connect(ctx context.Context) error
	Close() error
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)
	GetDatabaseType() string
	GetDatabaseVersion(ctx context.Context) (string, error)
	DryRunSQL(ctx context.Context, sql string) error
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

	// 连接池配置（可选）
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
