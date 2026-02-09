package tools

import (
	"context"
	"fmt"
	"strings"

	"lucid/interfaces"
)

// ExecuteSQL is a ReAct tool that lets the agent run read-only SQL on the business database.
type ExecuteSQL struct {
	adapter interfaces.DBAdapter
	callCount int
}

func NewExecuteSQL(adapter interfaces.DBAdapter) *ExecuteSQL {
	return &ExecuteSQL{adapter: adapter}
}

func (t *ExecuteSQL) Name() string        { return "execute_sql" }
func (t *ExecuteSQL) Description() string {
	return `Execute a read-only SQL query on the business database and return results.
Use this to explore data: row counts, sample values, value distributions, NULL checks, etc.
Input: a SQL query string (SELECT only).
Output: query results (rows and column names).`
}

func (t *ExecuteSQL) Call(ctx context.Context, input string) (string, error) {
	t.callCount++
	sql := strings.TrimSpace(input)

	// Safety: only allow SELECT
	upper := strings.ToUpper(sql)
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "SHOW") && !strings.HasPrefix(upper, "DESCRIBE") {
		return "Error: only SELECT / SHOW / DESCRIBE queries are allowed.", nil
	}

	result, err := t.adapter.ExecuteQuery(ctx, sql)
	if err != nil {
		return fmt.Sprintf("SQL error: %v", err), nil
	}

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Rows: %d\n", result.RowCount))
	if result.RowCount > 0 && len(result.Rows) > 0 {
		// Column names from first row
		if len(result.Columns) > 0 {
			sb.WriteString("Columns: " + strings.Join(result.Columns, ", ") + "\n")
		}
		// Show results (cap at 1000 chars)
		sampleStr := fmt.Sprintf("%v", result.Rows)
		const maxLen = 1000
		if len(sampleStr) <= maxLen {
			sb.WriteString("Results: " + sampleStr + "\n")
		} else {
			sb.WriteString("Results: " + sampleStr[:maxLen] + "... (truncated)\n")
		}
	}
	return sb.String(), nil
}

func (t *ExecuteSQL) CallCount() int { return t.callCount }
