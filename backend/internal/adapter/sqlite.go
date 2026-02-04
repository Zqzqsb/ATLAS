package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteAdapter SQLite适配器
type SQLiteAdapter struct {
	db     *sql.DB
	config *SQLiteConfig
}

// SQLiteConfig SQLite连接配置
type SQLiteConfig struct {
	FilePath string // 数据库文件路径，":memory:" 表示内存数据库
}

// NewSQLiteAdapter 创建SQLite适配器
func NewSQLiteAdapter(config *SQLiteConfig) *SQLiteAdapter {
	return &SQLiteAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *SQLiteAdapter) Connect(ctx context.Context) error {
	db, err := sql.Open("sqlite3", a.config.FilePath)
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
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// ExecuteQuery 执行查询
func (a *SQLiteAdapter) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	start := time.Now()

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return &QueryResult{
			Error:         err.Error(),
			ExecutionTime: time.Since(start).Milliseconds(),
		}, err // 返回错误，而不是 nil
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
func (a *SQLiteAdapter) GetDatabaseType() string {
	return "SQLite"
}

// GetDatabaseVersion 获取数据库版本
func (a *SQLiteAdapter) GetDatabaseVersion(ctx context.Context) (string, error) {
	result, err := a.ExecuteQuery(ctx, "SELECT sqlite_version() as version")
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
