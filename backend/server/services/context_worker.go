package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"

	"lucid/interfaces"
	"lucid/internal/lakebase"
)

// ContextWorker handles multi-phase Rich Context exploration for a table
type ContextWorker struct {
	workerID      string
	tableName     string
	datasourceID  int64
	model         llms.Model
	lakebaseRepo  *lakebase.MySQLRepository
	businessDB    interfaces.DBAdapter // Business database adapter for SQL execution
	eventCallback func(event WorkerEvent)
	config        WorkerConfig
}

// WorkerConfig holds configuration for context exploration
type WorkerConfig struct {
	MinIterations int
	MaxIterations int
	Force         bool
}

// WorkerEvent represents an event emitted during exploration
type WorkerEvent struct {
	Type         string                 `json:"type"`          // step, sql, result, context, error, done
	Phase        string                 `json:"phase"`         // metadata, quality, business, description
	Message      string                 `json:"message"`
	SQL          string                 `json:"sql,omitempty"`
	Result       interface{}            `json:"result,omitempty"`
	ContextKey   string                 `json:"context_key,omitempty"`
	ContextValue string                 `json:"context_value,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// NewContextWorker creates a new context exploration worker
func NewContextWorker(
	workerID string,
	tableName string,
	datasourceID int64,
	model llms.Model,
	repo *lakebase.MySQLRepository,
	businessDB interfaces.DBAdapter,
	config WorkerConfig,
	eventCallback func(WorkerEvent),
) *ContextWorker {
	return &ContextWorker{
		workerID:      workerID,
		tableName:     tableName,
		datasourceID:  datasourceID,
		model:         model,
		lakebaseRepo:  repo,
		businessDB:    businessDB,
		eventCallback: eventCallback,
		config:        config,
	}
}

// Execute runs the multi-phase context exploration
func (w *ContextWorker) Execute(ctx context.Context, columns []*lakebase.ColumnInfo) (string, error) {
	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "start",
		Message: fmt.Sprintf("Starting analysis of table %s", w.tableName),
	})

	// Phase 1: Collect basic metadata
	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "metadata",
		Message: "Phase 1: Collecting basic metadata...",
	})
	rowCount := w.collectMetadata(ctx)

	// Phase 2: Check data quality issues
	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "quality",
		Message: "Phase 2: Checking data quality issues...",
	})
	qualityIssues := w.checkDataQuality(ctx, columns)

	// Phase 3: Discover business semantics
	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "business",
		Message: "Phase 3: Discovering business semantics...",
	})
	enumFindings := w.discoverBusinessSemantics(ctx, columns)

	// Phase 4: Generate comprehensive description
	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "description",
		Message: "Phase 4: Generating table description...",
	})
	description, err := w.generateDescription(ctx, columns, rowCount, qualityIssues, enumFindings)
	if err != nil {
		w.emit(WorkerEvent{Type: "error", Phase: "description", Message: err.Error()})
		return "", err
	}

	w.emit(WorkerEvent{
		Type:    "done",
		Phase:   "complete",
		Message: fmt.Sprintf("Completed analysis of %s", w.tableName),
	})

	return description, nil
}

func (w *ContextWorker) emit(event WorkerEvent) {
	if w.eventCallback != nil {
		w.eventCallback(event)
	}
}

// collectMetadata collects basic table metadata
func (w *ContextWorker) collectMetadata(ctx context.Context) int64 {
	if w.businessDB == nil {
		return 0
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) as cnt FROM `%s`", w.tableName)
	w.emit(WorkerEvent{
		Type:    "sql",
		Phase:   "metadata",
		Message: "Getting row count",
		SQL:     countSQL,
	})

	result, err := w.businessDB.ExecuteQuery(ctx, countSQL)
	if err == nil && len(result.Rows) > 0 {
		if cnt, ok := result.Rows[0]["cnt"]; ok {
			var count int64
			switch v := cnt.(type) {
			case int64:
				count = v
			case float64:
				count = int64(v)
			case int:
				count = int64(v)
			}
			w.emit(WorkerEvent{
				Type:    "result",
				Phase:   "metadata",
				Message: fmt.Sprintf("Row count: %d", count),
				Result:  count,
			})
			return count
		}
	}
	return 0
}

// checkDataQuality checks for common data quality issues
func (w *ContextWorker) checkDataQuality(ctx context.Context, columns []*lakebase.ColumnInfo) []string {
	var issues []string

	if w.businessDB == nil {
		return issues
	}

	for _, col := range columns {
		// Skip primary keys
		if col.IsPrimaryKey {
			continue
		}

		colType := strings.ToUpper(col.DataType.String)

		// Check text columns for whitespace issues
		if strings.Contains(colType, "VARCHAR") || strings.Contains(colType, "CHAR") || strings.Contains(colType, "TEXT") {
			issue := w.checkWhitespaceIssue(ctx, col.ColumnName)
			if issue != "" {
				issues = append(issues, issue)
			}

			// Check if text column stores numeric values
			issue = w.checkTypeMismatch(ctx, col.ColumnName)
			if issue != "" {
				issues = append(issues, issue)
			}
		}

		// Check NULL percentage
		w.checkNullPercentage(ctx, col.ColumnName)
	}

	return issues
}

func (w *ContextWorker) checkWhitespaceIssue(ctx context.Context, columnName string) string {
	sql := fmt.Sprintf("SELECT `%s` FROM `%s` WHERE `%s` != TRIM(`%s`) AND `%s` IS NOT NULL LIMIT 3",
		columnName, w.tableName, columnName, columnName, columnName)

	w.emit(WorkerEvent{
		Type:    "sql",
		Phase:   "quality",
		Message: fmt.Sprintf("Checking whitespace in %s", columnName),
		SQL:     sql,
	})

	result, err := w.businessDB.ExecuteQuery(ctx, sql)
	if err != nil {
		return ""
	}

	if len(result.Rows) > 0 {
		var samples []string
		for _, row := range result.Rows {
			if val, ok := row[columnName]; ok {
				if s, ok := val.(string); ok {
					samples = append(samples, fmt.Sprintf("'%s'", s))
				}
			}
		}
		if len(samples) > 0 {
			issue := fmt.Sprintf("Column %s has leading/trailing whitespace (found %d examples). Use TRIM(%s) for matching.",
				columnName, len(samples), columnName)
			w.emit(WorkerEvent{
				Type:    "result",
				Phase:   "quality",
				Message: issue,
				Result:  samples,
			})
			return issue
		}
	}

	return ""
}

func (w *ContextWorker) checkTypeMismatch(ctx context.Context, columnName string) string {
	// Check if text column contains only numeric values
	sql := fmt.Sprintf("SELECT `%s` FROM `%s` WHERE `%s` IS NOT NULL AND `%s` REGEXP '^[0-9]+$' LIMIT 10",
		columnName, w.tableName, columnName, columnName)

	w.emit(WorkerEvent{
		Type:    "sql",
		Phase:   "quality",
		Message: fmt.Sprintf("Checking type mismatch in %s", columnName),
		SQL:     sql,
	})

	result, err := w.businessDB.ExecuteQuery(ctx, sql)
	if err != nil {
		return ""
	}

	if len(result.Rows) >= 5 {
		var samples []string
		for i, row := range result.Rows {
			if i >= 5 {
				break
			}
			if val, ok := row[columnName]; ok {
				samples = append(samples, fmt.Sprintf("%v", val))
			}
		}
		issue := fmt.Sprintf("Column %s (TEXT) appears to store numeric values. Consider CAST(%s AS SIGNED) for arithmetic.",
			columnName, columnName)
		w.emit(WorkerEvent{
			Type:    "result",
			Phase:   "quality",
			Message: issue,
			Result:  samples,
		})
		return issue
	}

	return ""
}

func (w *ContextWorker) checkNullPercentage(ctx context.Context, columnName string) {
	sql := fmt.Sprintf("SELECT COUNT(*) as total, COUNT(`%s`) as non_null FROM `%s`",
		columnName, w.tableName)

	w.emit(WorkerEvent{
		Type:    "sql",
		Phase:   "quality",
		Message: fmt.Sprintf("Checking NULL percentage in %s", columnName),
		SQL:     sql,
	})

	result, err := w.businessDB.ExecuteQuery(ctx, sql)
	if err != nil || len(result.Rows) == 0 {
		return
	}

	var total, nonNull int64
	row := result.Rows[0]
	if v, ok := row["total"]; ok {
		switch n := v.(type) {
		case int64:
			total = n
		case float64:
			total = int64(n)
		}
	}
	if v, ok := row["non_null"]; ok {
		switch n := v.(type) {
		case int64:
			nonNull = n
		case float64:
			nonNull = int64(n)
		}
	}

	if total > 0 {
		nullPct := float64(total-nonNull) / float64(total) * 100
		if nullPct > 5 {
			w.emit(WorkerEvent{
				Type:    "result",
				Phase:   "quality",
				Message: fmt.Sprintf("%s: %.1f%% NULL values", columnName, nullPct),
				Data: map[string]interface{}{
					"column":       columnName,
					"null_percent": nullPct,
				},
			})
		}
	}
}

// EnumFinding represents discovered enum-like values
type EnumFinding struct {
	ColumnName string
	Values     []string
}

// discoverBusinessSemantics discovers business meanings
func (w *ContextWorker) discoverBusinessSemantics(ctx context.Context, columns []*lakebase.ColumnInfo) []EnumFinding {
	var findings []EnumFinding

	if w.businessDB == nil {
		return findings
	}

	for _, col := range columns {
		// Skip primary keys
		if col.IsPrimaryKey {
			continue
		}

		finding := w.checkEnumValues(ctx, col)
		if finding != nil {
			findings = append(findings, *finding)
		}
	}

	return findings
}

func (w *ContextWorker) checkEnumValues(ctx context.Context, col *lakebase.ColumnInfo) *EnumFinding {
	sql := fmt.Sprintf("SELECT `%s`, COUNT(*) as cnt FROM `%s` WHERE `%s` IS NOT NULL GROUP BY `%s` ORDER BY cnt DESC LIMIT 20",
		col.ColumnName, w.tableName, col.ColumnName, col.ColumnName)

	w.emit(WorkerEvent{
		Type:    "sql",
		Phase:   "business",
		Message: fmt.Sprintf("Checking value distribution of %s", col.ColumnName),
		SQL:     sql,
	})

	result, err := w.businessDB.ExecuteQuery(ctx, sql)
	if err != nil || len(result.Rows) == 0 {
		return nil
	}

	type valueDist struct {
		Value string
		Count int64
	}
	var values []valueDist
	var total int64

	for _, row := range result.Rows {
		var val string
		var cnt int64

		if v, ok := row[col.ColumnName]; ok {
			val = fmt.Sprintf("%v", v)
		}
		if v, ok := row["cnt"]; ok {
			switch n := v.(type) {
			case int64:
				cnt = n
			case float64:
				cnt = int64(n)
			}
		}
		values = append(values, valueDist{Value: val, Count: cnt})
		total += cnt
	}

	// If column has <= 10 distinct values, it's likely an enum
	if len(values) > 0 && len(values) <= 10 {
		var parts []string
		var enumValues []string
		for _, v := range values {
			pct := float64(v.Count) / float64(total) * 100
			parts = append(parts, fmt.Sprintf("%s (%.0f%%)", v.Value, pct))
			enumValues = append(enumValues, v.Value)
		}
		distribution := strings.Join(parts, ", ")

		w.emit(WorkerEvent{
			Type:         "context",
			Phase:        "business",
			Message:      fmt.Sprintf("Found enum-like column: %s", col.ColumnName),
			ContextKey:   fmt.Sprintf("%s_values", col.ColumnName),
			ContextValue: distribution,
			Data: map[string]interface{}{
				"distinct_count": len(values),
				"values":         values,
			},
		})

		return &EnumFinding{
			ColumnName: col.ColumnName,
			Values:     enumValues,
		}
	}

	return nil
}

// generateDescription generates comprehensive table description using LLM
func (w *ContextWorker) generateDescription(ctx context.Context, columns []*lakebase.ColumnInfo, rowCount int64, qualityIssues []string, enumFindings []EnumFinding) (string, error) {
	// Build column info
	var colInfos []string
	for _, col := range columns {
		info := fmt.Sprintf("%s (%s)", col.ColumnName, col.DataType.String)
		if col.IsPrimaryKey {
			info += " [PK]"
		}
		if col.IsForeignKey {
			info += " [FK]"
		}
		colInfos = append(colInfos, info)
	}

	// Build quality issues summary
	qualitySummary := "None found"
	if len(qualityIssues) > 0 {
		qualitySummary = strings.Join(qualityIssues, "; ")
	}

	// Build enum findings summary
	enumSummary := "None"
	if len(enumFindings) > 0 {
		var enumParts []string
		for _, f := range enumFindings {
			enumParts = append(enumParts, fmt.Sprintf("%s has %d distinct values: %s",
				f.ColumnName, len(f.Values), strings.Join(f.Values, ", ")))
		}
		enumSummary = strings.Join(enumParts, "; ")
	}

	prompt := fmt.Sprintf(`Analyze this database table and generate a comprehensive description.

## Table: %s
## Row Count: %d
## Columns: %s

## Data Quality Issues Detected:
%s

## Enum-like Columns Discovered:
%s

## Your Task:
Generate a 2-3 sentence description that:
1. Explains the business purpose of this table
2. Highlights key columns and their semantic meaning
3. Notes important data quality considerations if any

Output ONLY the description text, no JSON or extra formatting.`, w.tableName, rowCount, strings.Join(colInfos, ", "), qualitySummary, enumSummary)

	w.emit(WorkerEvent{
		Type:    "step",
		Phase:   "description",
		Message: "Generating description via LLM...",
		Data: map[string]interface{}{
			"prompt_preview": fmt.Sprintf("Analyzing %s with %d columns, %d rows", w.tableName, len(columns), rowCount),
			"quality_issues": len(qualityIssues),
			"enum_columns":   len(enumFindings),
		},
	})

	response, err := llms.GenerateFromSinglePrompt(ctx, w.model, prompt)
	if err != nil {
		return "", err
	}

	description := strings.TrimSpace(response)
	w.emit(WorkerEvent{
		Type:         "context",
		Phase:        "description",
		Message:      "Generated table description",
		ContextKey:   "description",
		ContextValue: description,
	})

	return description, nil
}
