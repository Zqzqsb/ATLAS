package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"atlas/internal/logger"
)

// MySQLAdapter implements DBAdapter for MySQL / MariaDB.
type MySQLAdapter struct {
	db     *sql.DB
	config *MySQLConfig
}

// MySQLConfig holds MySQL connection parameters.
type MySQLConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// NewMySQLAdapter creates a new MySQLAdapter.
func NewMySQLAdapter(config *MySQLConfig) *MySQLAdapter {
	return &MySQLAdapter{
		config: config,
	}
}

// Connect establishes a connection to the database.
func (a *MySQLAdapter) Connect(ctx context.Context) error {
	log := logger.With("component", "mysql_adapter")
	log.Info("[Connect] Connecting to database",
		"host", a.config.Host,
		"port", a.config.Port,
		"database", a.config.Database,
		"user", a.config.User,
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		a.config.User,
		a.config.Password,
		a.config.Host,
		a.config.Port,
		a.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("[Connect] Failed to open database", "error", err)
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connectivity
	if err := db.PingContext(ctx); err != nil {
		log.Error("[Connect] Failed to ping database", "error", err)
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("[Connect] Database connected successfully",
		"database", a.config.Database,
	)
	a.db = db
	return nil
}

// Close closes the database connection.
func (a *MySQLAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// ExecuteQuery runs an SQL query and returns the result.
func (a *MySQLAdapter) ExecuteQuery(ctx context.Context, query string) (*QueryResult, error) {
	log := logger.With("component", "mysql_adapter")
	start := time.Now()

	log.Info("[ExecuteQuery] Executing SQL",
		"database", a.config.Database,
		"query", query,
	)

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		elapsed := time.Since(start)
		log.Error("[ExecuteQuery] Query failed",
			"database", a.config.Database,
			"query", query,
			"error", err,
			"duration", elapsed.Round(time.Millisecond),
		)
		return &QueryResult{
			Error:         err.Error(),
			ExecutionTime: elapsed.Milliseconds(),
		}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Error("[ExecuteQuery] Failed to get columns", "error", err)
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Error("[ExecuteQuery] Failed to scan row", "error", err)
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
		log.Error("[ExecuteQuery] Row iteration error", "error", err)
		return nil, err
	}

	elapsed := time.Since(start)
	log.Info("[ExecuteQuery] Query completed",
		"database", a.config.Database,
		"row_count", len(result),
		"column_count", len(columns),
		"columns", columns,
		"duration", elapsed.Round(time.Millisecond),
	)

	return &QueryResult{
		Columns:       columns,
		Rows:          result,
		RowCount:      len(result),
		ExecutionTime: elapsed.Milliseconds(),
	}, nil
}

// GetDatabaseType returns the database engine name.
func (a *MySQLAdapter) GetDatabaseType() string {
	return "MySQL"
}

// GetDatabaseVersion returns the server version string.
func (a *MySQLAdapter) GetDatabaseVersion(ctx context.Context) (string, error) {
	result, err := a.ExecuteQuery(ctx, "SELECT VERSION() as version")
	if err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("%s", result.Error)
	}
	if len(result.Rows) > 0 {
		if version, ok := result.Rows[0]["version"].(string); ok {
			return version, nil
		}
	}
	return "unknown", nil
}
