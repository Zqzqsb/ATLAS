package adapter

import (
	"context"
	"fmt"
)

// DryRunSQL validates SQL syntax via EXPLAIN (does not execute the actual query).
func (a *MySQLAdapter) DryRunSQL(ctx context.Context, sql string) (*QueryResult, error) {
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)
	result, err := a.ExecuteQuery(ctx, explainSQL)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DryRunSQL SQLite: uses EXPLAIN QUERY PLAN.
func (a *SQLiteAdapter) DryRunSQL(ctx context.Context, sql string) (*QueryResult, error) {
	explainSQL := fmt.Sprintf("EXPLAIN QUERY PLAN %s", sql)
	result, err := a.ExecuteQuery(ctx, explainSQL)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DryRunSQL PostgreSQL: uses EXPLAIN.
func (a *PostgreSQLAdapter) DryRunSQL(ctx context.Context, sql string) (*QueryResult, error) {
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)
	result, err := a.ExecuteQuery(ctx, explainSQL)
	if err != nil {
		return nil, err
	}
	return result, nil
}
