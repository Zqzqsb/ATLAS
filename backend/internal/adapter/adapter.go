// Package adapter provides concrete database adapter implementations.
// All adapters implement interfaces.DBAdapter directly.
package adapter

import (
	"lucid/interfaces"
)

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
	SQLite     DatabaseType = "sqlite"
)

// NewAdapter 工厂函数：根据配置创建对应的适配器
func NewAdapter(config *interfaces.DBConfig) (interfaces.DBAdapter, error) {
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
