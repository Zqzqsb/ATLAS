package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLAdapter MySQL适配器
type MySQLAdapter struct {
	db     *sql.DB
	config *MySQLConfig
}

// MySQLConfig MySQL连接配置
type MySQLConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// NewMySQLAdapter 创建MySQL适配器
func NewMySQLAdapter(config *MySQLConfig) *MySQLAdapter {
	return &MySQLAdapter{
		config: config,
	}
}

// Connect 连接数据库
func (a *MySQLAdapter) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		a.config.User,
		a.config.Password,
		a.config.Host,
		a.config.Port,
		a.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
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
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// ExecuteQuery 执行查询
func (a *MySQLAdapter) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
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
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// 扫描行
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 构建map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// 处理[]byte类型
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
func (a *MySQLAdapter) GetDatabaseType() string {
	return "MySQL"
}

// GetDatabaseVersion 获取数据库版本
func (a *MySQLAdapter) GetDatabaseVersion(ctx context.Context) (string, error) {
	result, err := a.ExecuteQuery(ctx, "SELECT VERSION() as version")
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
