package tools

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"atlas/internal/adapter"
)

// VerifySQLTool validates SQL via EXPLAIN (no actual execution).
// Performs static checks + EXPLAIN plan analysis.
type VerifySQLTool struct {
	adapter  adapter.DBAdapter
	dbType   string
	injector ObservationInjector
}

func NewVerifySQLTool(dbAdapter adapter.DBAdapter, dbType string) *VerifySQLTool {
	return &VerifySQLTool{
		adapter: dbAdapter,
		dbType:  dbType,
	}
}

// SetObservationInjector implements ObservationInjectable.
func (t *VerifySQLTool) SetObservationInjector(injector ObservationInjector) {
	t.injector = injector
}

func (t *VerifySQLTool) Name() string { return "verify_sql" }

func (t *VerifySQLTool) Description() string {
	return `Validate SQL syntax and inspect execution plan via EXPLAIN (does NOT execute the actual query).

Input: SQL query string to verify
Output: ✅ VERIFY_PASSED or ❌ VERIFY_FAILED, with EXPLAIN execution plan

IMPORTANT workflow:
- You MUST call verify_sql before giving your Final Answer.
- If ❌ VERIFY_FAILED: fix the SQL error and call verify_sql AGAIN with the corrected SQL. Repeat until it passes.
- If ✅ VERIFY_PASSED: proceed to Final Answer.
- NEVER give Final Answer with SQL that has not passed verify_sql.`
}

func (t *VerifySQLTool) Call(ctx context.Context, input string) (string, error) {
	sql := stripCodeFence(strings.TrimSpace(input))

	// Step 1: Static checks
	if err := t.quickCheck(sql); err != nil {
		slog.Warn("[VerifySQL] Static check failed",
			"component", "verify_sql",
			"error", err.Error(),
			"sql_preview", truncateStr(sql, 200),
		)
		out := fmt.Sprintf("❌ VERIFY_FAILED\nSQL validation failed (static check):\n%v\n\nYou MUST fix the error and call verify_sql again with the corrected SQL. Do NOT give Final Answer until verify_sql passes.", err)
		t.emitResult(out)
		return out, nil
	}

	// Step 2: EXPLAIN validation
	explainResult, err := t.adapter.DryRunSQL(ctx, sql)
	if err != nil {
		out := fmt.Sprintf("❌ VERIFY_FAILED\nSQL validation failed (EXPLAIN check):\n%v\n\nYou MUST fix the error and call verify_sql again with the corrected SQL. Do NOT give Final Answer until verify_sql passes.", err)
		t.emitResult(out)
		return out, nil
	}

	// Step 3: Format EXPLAIN plan
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
		sb.WriteString(`
Decision guide:
  - Full table scan on ≤1000 rows: ACCEPTABLE, proceed to Final Answer.
  - Full table scan on >1000 rows: Consider adding WHERE conditions or using indexed columns.
  - Using filesort / temporary on large tables: Consider rewriting ORDER BY or GROUP BY.
  - If optimization is not possible (e.g., no suitable index exists), proceed to Final Answer anyway.
`)
	} else {
		sb.WriteString("\nExecution plan looks good. You can now provide the final answer.")
	}

	out := sb.String()
	t.emitResult(out)
	return out, nil
}

func (t *VerifySQLTool) emitResult(output string) {
	if t.injector != nil {
		t.injector.InjectObservation(output)
	}
}

func (t *VerifySQLTool) formatExplainPlan(result *adapter.QueryResult) string {
	if result == nil || len(result.Rows) == 0 {
		return "  (empty execution plan)\n"
	}

	var sb strings.Builder
	for i, row := range result.Rows {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("  ... (%d more steps)\n", len(result.Rows)-10))
			break
		}

		tableName := valOrDash(row, "table")
		scanType := valOrDash(row, "type")
		possibleKeys := valOrDash(row, "possible_keys")
		usedKey := valOrDash(row, "key")
		rowsEst := valOrDash(row, "rows")
		extra := valOrDash(row, "Extra")

		sb.WriteString(fmt.Sprintf("  Step %d: %s\n", i+1, tableName))
		sb.WriteString(fmt.Sprintf("    scan: %-10s  key: %-20s  rows: %s\n", scanType, usedKey, rowsEst))
		if possibleKeys != "-" {
			sb.WriteString(fmt.Sprintf("    possible_keys: %s\n", possibleKeys))
		}
		if extra != "-" {
			sb.WriteString(fmt.Sprintf("    extra: %s\n", extra))
		}
	}
	return sb.String()
}

func valOrDash(row map[string]interface{}, key string) string {
	if v, ok := row[key]; ok && v != nil {
		s := fmt.Sprintf("%v", v)
		if s == "" || s == "<nil>" {
			return "-"
		}
		return s
	}
	return "-"
}

func (t *VerifySQLTool) analyzeExplainPlan(result *adapter.QueryResult) []string {
	if result == nil || len(result.Rows) == 0 {
		return nil
	}

	var warnings []string
	for _, row := range result.Rows {
		if scanType, ok := row["type"]; ok {
			typeStr := fmt.Sprintf("%v", scanType)
			if typeStr == "ALL" {
				tableName := ""
				if tn, ok := row["table"]; ok {
					tableName = fmt.Sprintf("%v", tn)
				}
				rowsEst := ""
				if r, ok := row["rows"]; ok {
					rowsEst = fmt.Sprintf("%v", r)
				}
				warnings = append(warnings, fmt.Sprintf("Full table scan on '%s' (estimated %s rows). Consider adding WHERE conditions or indexes.", tableName, rowsEst))
			}
		}

		if extra, ok := row["Extra"]; ok {
			extraStr := fmt.Sprintf("%v", extra)
			if strings.Contains(extraStr, "Using filesort") {
				warnings = append(warnings, "Using filesort — may be slow on large datasets. Consider adding an index on the ORDER BY column(s).")
			}
			if strings.Contains(extraStr, "Using temporary") {
				warnings = append(warnings, "Using temporary table — may impact performance for GROUP BY / DISTINCT operations.")
			}
		}
	}
	return warnings
}

func (t *VerifySQLTool) quickCheck(sql string) error {
	if err := checkIllegalAliases(sql); err != nil {
		return err
	}
	if err := checkParentheses(sql); err != nil {
		return err
	}
	return nil
}

var (
	castPattern         = regexp.MustCompile(`(?i)CAST\s*\([^)]*\s+AS\s+[A-Z_]+\s*\([^)]*\)\s*\)`)
	convertPattern      = regexp.MustCompile(`(?i)CONVERT\s*\([^,]+,\s*[A-Z_]+\s*\([^)]*\)\s*\)`)
	illegalAliasPattern = regexp.MustCompile(`(?i)\s+AS\s+([a-z_]+\s*\([^)]*\))`)
)

func checkIllegalAliases(sql string) error {
	cleaned := castPattern.ReplaceAllString(sql, "/*CAST*/")
	cleaned = convertPattern.ReplaceAllString(cleaned, "/*CONVERT*/")

	matches := illegalAliasPattern.FindAllStringSubmatch(cleaned, -1)
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

func checkParentheses(sql string) error {
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


