package adapter

import (
	"context"
	"fmt"
)

// DryRunSQL 验证 SQL 语法（不执行）
func (a *MySQLAdapter) DryRunSQL(ctx context.Context, sql string) error {
	// MySQL: 使用 EXPLAIN 验证语法
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)
	_, err := a.ExecuteQuery(ctx, explainSQL)
	return err
}

// DryRunSQL SQLite 的 Dry Run
func (a *SQLiteAdapter) DryRunSQL(ctx context.Context, sql string) error {
	// SQLite: 使用 EXPLAIN QUERY PLAN 验证语法
	explainSQL := fmt.Sprintf("EXPLAIN QUERY PLAN %s", sql)
	_, err := a.ExecuteQuery(ctx, explainSQL)
	return err
}

// DryRunSQL PostgreSQL 的 Dry Run
func (a *PostgreSQLAdapter) DryRunSQL(ctx context.Context, sql string) error {
	// PostgreSQL: 使用 EXPLAIN 验证语法
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)
	_, err := a.ExecuteQuery(ctx, explainSQL)
	return err
}

// DBAdapter 接口添加 DryRunSQL 方法
// 需要在 adapter.go 中添加此方法声明
