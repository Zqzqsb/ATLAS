package adapter

import (
	"context"
)

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLite     DatabaseType = "sqlite"
)

// DBAdapter 数据库适配器接口
// 轻量级设计：只负责连接和执行SQL，不做ORM
type DBAdapter interface {
	// Connect 连接数据库
	Connect(ctx context.Context) error

	// Close 关闭连接
	Close() error

	// ExecuteQuery 执行查询
	// 返回统一的QueryResult结构，包含列名、数据行、执行时间等
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)

	// GetDatabaseType 获取数据库类型
	// 返回值: "MySQL", "PostgreSQL", "SQLite" 等
	GetDatabaseType() string

	// GetDatabaseVersion 获取数据库版本（可选）
	GetDatabaseVersion(ctx context.Context) (string, error)

	// DryRunSQL 验证 SQL 语法（不执行）
	DryRunSQL(ctx context.Context, sql string) error
}

// QueryResult 查询结果（统一结构）
type QueryResult struct {
	Columns       []string                 // 列名
	Rows          []map[string]interface{} // 数据行（统一为map格式）
	RowCount      int                      // 行数
	ExecutionTime int64                    // 执行时间（毫秒）
	Error         string                   // 错误信息（如果有）
}

// DBConfig 数据库连接配置（通用）
type DBConfig struct {
	Type     string // 数据库类型: "mysql", "postgresql", "sqlite"
	Host     string // 主机地址
	Port     int    // 端口
	Database string // 数据库名
	User     string // 用户名
	Password string // 密码

	// SQLite特有
	FilePath string // SQLite文件路径

	// 连接池配置（可选）
	MaxOpenConns int // 最大打开连接数
	MaxIdleConns int // 最大空闲连接数
}

// NewAdapter 工厂函数：根据配置创建对应的适配器
func NewAdapter(config *DBConfig) (DBAdapter, error) {
	switch config.Type {
	case "mysql":
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
