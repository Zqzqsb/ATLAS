package tools

import (
	"context"
	"fmt"
	"strings"

	"atlas/internal/adapter"
)

// ObservationInjector pushes a tool observation directly into the ReAct handler,
// bypassing langchaingo's unreliable HandleToolEnd callback.
// Implemented by react.Handler.
type ObservationInjector interface {
	InjectObservation(output string)
}

// InferenceSQLTool is the execute_sql tool for the inference pipeline.
// Unlike the generic ExecuteSQL tool, it supports DryRun validation and
// observation injection for reliable SSE streaming.
type InferenceSQLTool struct {
	adapter        adapter.DBAdapter
	useDryRun      bool
	ExecutionCount int
	injector       ObservationInjector
}

func NewInferenceSQLTool(dbAdapter adapter.DBAdapter, useDryRun bool) *InferenceSQLTool {
	return &InferenceSQLTool{
		adapter:   dbAdapter,
		useDryRun: useDryRun,
	}
}

// SetObservationInjector implements ObservationInjectable.
func (t *InferenceSQLTool) SetObservationInjector(injector ObservationInjector) {
	t.injector = injector
}

func (t *InferenceSQLTool) Name() string { return "execute_sql" }

func (t *InferenceSQLTool) Description() string {
	if t.useDryRun {
		return `Execute SQL query with dry run validation first.
Input: SQL query string
Output: Query results or validation errors`
	}
	return `Execute SQL query and return results.
Input: SQL query string
Output: Query results`
}

func (t *InferenceSQLTool) Call(ctx context.Context, input string) (string, error) {
	t.ExecutionCount++
	sql := stripCodeFence(strings.TrimSpace(input))

	if t.useDryRun {
		if _, err := t.adapter.DryRunSQL(ctx, sql); err != nil {
			out := fmt.Sprintf("SQL validation failed: %v", err)
			t.emitResult(out)
			return out, nil
		}
	}

	result, err := t.adapter.ExecuteQuery(ctx, sql)
	if err != nil {
		out := fmt.Sprintf("SQL execution failed: %v", err)
		t.emitResult(out)
		return out, nil
	}

	output := fmt.Sprintf("Query executed successfully!\nRows: %d\n", result.RowCount)
	if result.RowCount > 0 {
		sampleStr := fmt.Sprintf("%v", result.Rows)
		const maxSampleLength = 1000
		if len(sampleStr) <= maxSampleLength {
			output += fmt.Sprintf("Sample results: %s\n", sampleStr)
		} else {
			output += fmt.Sprintf("Sample results: %s... (truncated)\n", sampleStr[:maxSampleLength])
		}
	}

	t.emitResult(output)
	return output, nil
}

func (t *InferenceSQLTool) emitResult(output string) {
	if t.injector != nil {
		t.injector.InjectObservation(output)
	}
}

// stripCodeFence removes markdown code fences from LLM output.
func stripCodeFence(s string) string {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "```") {
		idx := strings.Index(trimmed, "\n")
		if idx >= 0 {
			trimmed = trimmed[idx+1:]
		} else {
			trimmed = strings.TrimPrefix(trimmed, "```sql")
			trimmed = strings.TrimPrefix(trimmed, "```")
		}
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	return trimmed
}
