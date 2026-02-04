package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLAdapter PostgreSQL适配器
type PostgreSQLAdapter struct {
	db     *sql.DB
	config *PostgreSQLConfig
}

// PostgreSQLConfig PostgreSQL连接配置
type PostgreSQLConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string // disable, require, verify-ca, verify-full
}

// NewPostgreSQLAdapter 创建PostgreSQL适配器
func NewPostgreSQLAdapter(config *PostgreSQLConfig) *PostgreSQLAdapter {
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	return &PostgreSQLAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *PostgreSQLAdapter) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		a.config.Host,
		a.config.Port,
		a.config.User,
		a.config.Password,
		a.config.Database,
		a.config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	a.db = db
	return nil
}

// Close 关闭连接
func (a *PostgreSQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// ExecuteQuery 执行查询
func (a *PostgreSQLAdapter) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	start := time.Now()

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return &QueryResult{
			Error:         err.Error(),
			ExecutionTime: time.Since(start).Milliseconds(),
		}, err // 返回错误，让调用方能正确处理
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 读取数据
	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &QueryResult{
		Columns:       columns,
		Rows:          result,
		RowCount:      len(result),
		ExecutionTime: time.Since(start).Milliseconds(),
	}, nil
}

// GetDatabaseType 获取数据库类型
func (a *PostgreSQLAdapter) GetDatabaseType() string {
	return "PostgreSQL"
}

// GetDatabaseVersion 获取数据库版本
func (a *PostgreSQLAdapter) GetDatabaseVersion(ctx context.Context) (string, error) {
	result, err := a.ExecuteQuery(ctx, "SELECT version() as version")
	if err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf(result.Error)
	}
	if len(result.Rows) > 0 {
		if version, ok := result.Rows[0]["version"].(string); ok {
			return version, nil
		}
	}
	return "unknown", nil
}
