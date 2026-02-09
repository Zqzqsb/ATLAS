package inference

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"lucid/internal/adapter"
)

// VerifySQLTool SQL validation tool using EXPLAIN (no actual execution)
type VerifySQLTool struct {
	adapter adapter.DBAdapter
	dbType  string
}

// Name returns tool name
func (t *VerifySQLTool) Name() string {
	return "verify_sql"
}

// Description returns tool description
func (t *VerifySQLTool) Description() string {
	return `Verify SQL syntax and inspect execution plan via EXPLAIN (does NOT execute the actual query).

Input: SQL query string to verify
Output: ✅ VERIFY_PASSED or ❌ VERIFY_FAILED, with EXPLAIN execution plan and performance warnings

What it checks:
- Static syntax errors (illegal aliases, unmatched parentheses)
- Database-level validation via EXPLAIN
- Execution plan analysis (full table scans, join strategies, estimated rows)

⚠️ IMPORTANT: After receiving the result, YOU MUST act on it:
- If ❌ VERIFY_FAILED: Fix the SQL error and call verify_sql again with the corrected SQL.
- If ✅ VERIFY_PASSED but has ⚠️ Performance warnings (e.g., full table scan, filesort, temporary table):
  → Rewrite the SQL to avoid the issue (add WHERE filters, use indexed columns, restructure JOINs)
  → Then call verify_sql again to confirm the improvement
- If ✅ VERIFY_PASSED with no warnings: Safe to give Final Answer.

You may iterate up to 3 times to optimize the SQL based on EXPLAIN feedback.
Use this tool BEFORE giving your final answer.`
}

// Call executes the verification
func (t *VerifySQLTool) Call(ctx context.Context, input string) (string, error) {
	sql := strings.TrimSpace(input)

	// Step 1: Static checks (fast, no DB call)
	if err := t.quickCheck(sql); err != nil {
		return fmt.Sprintf("❌ VERIFY_FAILED\nSQL validation failed (static check):\n%v\n\nPlease fix the error and try again.", err), nil
	}

	// Step 2: EXPLAIN validation (safe, doesn't execute the actual query)
	explainResult, err := t.adapter.DryRunSQL(ctx, sql)
	if err != nil {
		return fmt.Sprintf("❌ VERIFY_FAILED\nSQL validation failed (EXPLAIN check):\n%v\n\nPlease fix the error and try again.", err), nil
	}

	// Step 3: Format EXPLAIN plan for agent review
	planSummary := t.formatExplainPlan(explainResult)
	warnings := t.analyzeExplainPlan(explainResult)

	var sb strings.Builder
	sb.WriteString("✅ VERIFY_PASSED\n")
	sb.WriteString("SQL syntax is valid. EXPLAIN execution plan:\n")
	sb.WriteString(planSummary)

	if len(warnings) > 0 {
		sb.WriteString("\n⚠️ Performance warnings:\n")
		for _, w := range warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", w))
		}
		sb.WriteString("\nConsider optimizing the SQL if these warnings are critical for the use case.")
	} else {
		sb.WriteString("\nExecution plan looks good. You can now provide the final answer.")
	}

	return sb.String(), nil
}

// formatExplainPlan formats EXPLAIN QueryResult into a readable summary
func (t *VerifySQLTool) formatExplainPlan(result *adapter.QueryResult) string {
	if result == nil || len(result.Rows) == 0 {
		return "  (empty execution plan)\n"
	}

	var sb strings.Builder

	// Format column headers
	if len(result.Columns) > 0 {
		sb.WriteString("  ")
		sb.WriteString(strings.Join(result.Columns, " | "))
		sb.WriteString("\n")
		sb.WriteString("  " + strings.Repeat("-", 60) + "\n")
	}

	// Format rows (limit to first 10 rows for readability)
	maxRows := 10
	if len(result.Rows) < maxRows {
		maxRows = len(result.Rows)
	}
	for i := 0; i < maxRows; i++ {
		row := result.Rows[i]
		var vals []string
		for _, col := range result.Columns {
			if v, ok := row[col]; ok {
				vals = append(vals, fmt.Sprintf("%v", v))
			} else {
				vals = append(vals, "")
			}
		}
		sb.WriteString("  ")
		sb.WriteString(strings.Join(vals, " | "))
		sb.WriteString("\n")
	}
	if len(result.Rows) > maxRows {
		sb.WriteString(fmt.Sprintf("  ... (%d more rows)\n", len(result.Rows)-maxRows))
	}

	return sb.String()
}

// analyzeExplainPlan checks the EXPLAIN plan for potential performance issues
func (t *VerifySQLTool) analyzeExplainPlan(result *adapter.QueryResult) []string {
	if result == nil || len(result.Rows) == 0 {
		return nil
	}

	var warnings []string

	for _, row := range result.Rows {
		// MySQL/MariaDB EXPLAIN fields
		if scanType, ok := row["type"]; ok {
			typeStr := fmt.Sprintf("%v", scanType)
			if typeStr == "ALL" {
				tableName := ""
				if t, ok := row["table"]; ok {
					tableName = fmt.Sprintf("%v", t)
				}
				rowsEst := ""
				if r, ok := row["rows"]; ok {
					rowsEst = fmt.Sprintf("%v", r)
				}
				warnings = append(warnings, fmt.Sprintf("Full table scan on '%s' (estimated %s rows). Consider adding WHERE conditions or indexes.", tableName, rowsEst))
			}
		}

		// Check for "Using filesort" or "Using temporary" in Extra
		if extra, ok := row["Extra"]; ok {
			extraStr := fmt.Sprintf("%v", extra)
			if strings.Contains(extraStr, "Using filesort") {
				warnings = append(warnings, "Using filesort — may be slow on large datasets. Consider adding an index on the ORDER BY column(s).")
			}
			if strings.Contains(extraStr, "Using temporary") {
				warnings = append(warnings, "Using temporary table — may impact performance for GROUP BY / DISTINCT operations.")
			}
		}

		// SQLite EXPLAIN QUERY PLAN: check for "SCAN TABLE" (full scan)
		if detail, ok := row["detail"]; ok {
			detailStr := fmt.Sprintf("%v", detail)
			if strings.Contains(detailStr, "SCAN TABLE") {
				warnings = append(warnings, fmt.Sprintf("Full table scan detected: %s", detailStr))
			}
		}
	}

	return warnings
}

// quickCheck performs fast static checks
func (t *VerifySQLTool) quickCheck(sql string) error {
	if err := t.checkIllegalAliases(sql); err != nil {
		return err
	}
	if err := t.checkParentheses(sql); err != nil {
		return err
	}
	return nil
}

// checkIllegalAliases checks for illegal alias patterns
func (t *VerifySQLTool) checkIllegalAliases(sql string) error {
	illegalAliasPattern := regexp.MustCompile(`(?i)\s+AS\s+([a-z_]+\s*\([^)]*\))`)

	matches := illegalAliasPattern.FindAllStringSubmatch(sql, -1)
	if len(matches) > 0 {
		aliases := make([]string, 0, len(matches))
		for _, match := range matches {
			if len(match) > 1 {
				aliases = append(aliases, match[1])
			}
		}
		return fmt.Errorf("illegal alias syntax: %v\nAliases cannot contain parentheses.\nUse simple names like 'total_count' instead of 'count(*)'", aliases)
	}
	return nil
}

// checkParentheses checks for matching parentheses
func (t *VerifySQLTool) checkParentheses(sql string) error {
	stack := 0
	for i, char := range sql {
		if char == '(' {
			stack++
		} else if char == ')' {
			stack--
			if stack < 0 {
				return fmt.Errorf("unmatched closing parenthesis at position %d", i)
			}
		}
	}
	if stack > 0 {
		return fmt.Errorf("unmatched opening parenthesis: %d unclosed", stack)
	}
	return nil
}

// NewVerifySQLTool creates a verification tool
func NewVerifySQLTool(dbAdapter adapter.DBAdapter, dbType string) *VerifySQLTool {
	return &VerifySQLTool{
		adapter: dbAdapter,
		dbType:  dbType,
	}
}

// checkDuplicateRows 检查结果中是否有重复行
func (t *VerifySQLTool) checkDuplicateRows(rows [][]string) string {
	if len(rows) <= 2 { // 没有数据行或只有一行数据
		return ""
	}

	seen := make(map[string]bool)
	dataRows := rows[1:] // 排除标题行

	for _, row := range dataRows {
		// 为行创建一个唯一的键
		rowKey := strings.Join(row, "||<SEP>||")
		if seen[rowKey] {
			// 发现重复
			return fmt.Sprintf("Warning: The query returned duplicate rows (e.g., %v). Review the question to determine if duplicates should be removed using DISTINCT.", row)
		}
		seen[rowKey] = true
	}

	return ""
}

// convertQueryResultFormat 将查询结果从 map 转换为二维字符串数组
func convertQueryResultFormat(data []map[string]interface{}) [][]string {
	if len(data) == 0 {
		return nil
	}

	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	result := make([][]string, len(data)+1)
	result[0] = headers

	for i, row := range data {
		rowValues := make([]string, len(headers))
		for j, header := range headers {
			if val, ok := row[header]; ok {
				rowValues[j] = fmt.Sprintf("%v", val)
			} else {
				rowValues[j] = ""
			}
		}
		result[i+1] = rowValues
	}

	return result
}
