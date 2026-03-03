package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"log/slog"

	"lucid/internal/logger"
)

// RCWriter is the interface for writing Rich Context data.
// The ReAct agent calls set_rich_context with structured JSON, and RCWriter handles persistence.
type RCWriter interface {
	SetTableDescription(ctx context.Context, dsID int64, tableName, description string) error
	SetColumnDescription(ctx context.Context, dsID int64, tableName, columnName, description string) error
	SetColumnSampleValues(ctx context.Context, dsID int64, tableName, columnName, sampleValues string) error
	SetColumnSynonyms(ctx context.Context, dsID int64, tableName, columnName, synonyms string) error
	AddBusinessTerm(ctx context.Context, dsID int64, term, definition, synonyms, examples, category string) error
}

// SetRichContext is a ReAct tool that lets the agent write discovered context to the RC store.
// The agent doesn't need to know about rc_tables/rc_columns/rc_terms — it just calls this tool
// with a structured action describing what context to set.
type SetRichContext struct {
	writer    RCWriter
	dsID      int64
	callCount int
}

func NewSetRichContext(writer RCWriter, dsID int64) *SetRichContext {
	return &SetRichContext{writer: writer, dsID: dsID}
}

func (t *SetRichContext) Name() string { return "set_rich_context" }
func (t *SetRichContext) Description() string {
	return `Save discovered Rich Context to the knowledge store.
Input: a JSON object OR a JSON array of objects. Each object has these fields:
  - "type": one of "table_description", "column_description", "column_sample_values", "column_synonyms", "business_term"
  - "table": table name (required for table_* and column_* types)
  - "column": column name (required for column_* types)
  - "value": the context value (required for all types)
  - "definition": term definition (required for business_term)
  - "synonyms": comma-separated synonyms (optional, for business_term or column_synonyms)
  - "examples": usage examples (optional, for business_term)
  - "category": term category (optional, for business_term)

PREFER BATCH MODE: Pass a JSON array to save multiple items in one call (much more efficient).
Example (batch — saves 4 items in ONE call):
  [
    {"type": "table_description", "table": "orders", "value": "Contains customer purchase orders."},
    {"type": "column_description", "table": "orders", "column": "status", "value": "Order lifecycle status."},
    {"type": "column_sample_values", "table": "orders", "column": "status", "value": "pending, shipped, delivered"},
    {"type": "column_synonyms", "table": "orders", "column": "status", "value": "order state, fulfillment status"}
  ]
Example (single):
  {"type": "table_description", "table": "orders", "value": "Contains customer purchase orders."}
Output: confirmation message with count of saved items.`
}

type rcAction struct {
	Type       string `json:"type"`
	Table      string `json:"table,omitempty"`
	Column     string `json:"column,omitempty"`
	Value      string `json:"value"`
	Definition string `json:"definition,omitempty"`
	Synonyms   string `json:"synonyms,omitempty"`
	Examples   string `json:"examples,omitempty"`
	Category   string `json:"category,omitempty"`
}

func (t *SetRichContext) Call(ctx context.Context, input string) (string, error) {
	t.callCount++
	log := logger.With("component", "set_rich_context", "dsID", t.dsID, "call", t.callCount)

	trimmed := stripMarkdownCodeBlock(strings.TrimSpace(input))

	// Detect batch mode (JSON array)
	if strings.HasPrefix(trimmed, "[") {
		var actions []rcAction
		if err := json.Unmarshal([]byte(trimmed), &actions); err != nil {
			log.Warn("invalid JSON array input", "error", err, "input", truncateStr(trimmed, 200))
			return fmt.Sprintf("Error: invalid JSON array. Got: %s", truncateStr(trimmed, 200)), nil
		}
		if len(actions) == 0 {
			return "Error: empty array. Provide at least one item.", nil
		}

		var saved, failed int
		var errors []string
		for i, action := range actions {
			if result, err := t.processSingle(ctx, log, action); err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("[%d] %s", i, err.Error()))
			} else if strings.HasPrefix(result, "Error:") {
				failed++
				errors = append(errors, fmt.Sprintf("[%d] %s", i, result))
			} else {
				saved++
			}
		}

		msg := fmt.Sprintf("Batch complete: %d/%d saved.", saved, saved+failed)
		if len(errors) > 0 {
			msg += " Errors: " + strings.Join(errors, "; ")
		}
		return msg, nil
	}

	// Single mode
	var action rcAction
	if err := json.Unmarshal([]byte(trimmed), &action); err != nil {
		log.Warn("invalid JSON input", "error", err, "input", truncateStr(trimmed, 200))
		return fmt.Sprintf("Error: invalid JSON input. Expected a JSON object or array. Got: %s", truncateStr(trimmed, 200)), nil
	}

	result, _ := t.processSingle(ctx, log, action)
	return result, nil
}

// processSingle handles a single rcAction write. Returns (message, error).
func (t *SetRichContext) processSingle(ctx context.Context, log *slog.Logger, action rcAction) (string, error) {
	log.Info("writing rich context",
		"type", action.Type,
		"table", action.Table,
		"column", action.Column,
		"value_length", len(action.Value),
	)

	switch action.Type {
	case "table_description":
		if action.Table == "" || action.Value == "" {
			return "Error: 'table' and 'value' are required for table_description.", nil
		}
		if err := t.writer.SetTableDescription(ctx, t.dsID, action.Table, action.Value); err != nil {
			log.Error("failed to save table description", "table", action.Table, "error", err)
			return fmt.Sprintf("Error saving table description: %v", err), nil
		}
		return fmt.Sprintf("Saved table description for '%s'.", action.Table), nil

	case "column_description":
		if action.Table == "" || action.Column == "" || action.Value == "" {
			return "Error: 'table', 'column', and 'value' are required for column_description.", nil
		}
		if err := t.writer.SetColumnDescription(ctx, t.dsID, action.Table, action.Column, action.Value); err != nil {
			log.Error("failed to save column description", "table", action.Table, "column", action.Column, "error", err)
			return fmt.Sprintf("Error saving column description: %v", err), nil
		}
		return fmt.Sprintf("Saved column description for '%s.%s'.", action.Table, action.Column), nil

	case "column_sample_values":
		if action.Table == "" || action.Column == "" || action.Value == "" {
			return "Error: 'table', 'column', and 'value' are required for column_sample_values.", nil
		}
		if err := t.writer.SetColumnSampleValues(ctx, t.dsID, action.Table, action.Column, action.Value); err != nil {
			log.Error("failed to save sample values", "table", action.Table, "column", action.Column, "error", err)
			return fmt.Sprintf("Error saving sample values: %v", err), nil
		}
		return fmt.Sprintf("Saved sample values for '%s.%s'.", action.Table, action.Column), nil

	case "column_synonyms":
		if action.Table == "" || action.Column == "" || action.Value == "" {
			return "Error: 'table', 'column', and 'value' are required for column_synonyms.", nil
		}
		if err := t.writer.SetColumnSynonyms(ctx, t.dsID, action.Table, action.Column, action.Value); err != nil {
			log.Error("failed to save synonyms", "table", action.Table, "column", action.Column, "error", err)
			return fmt.Sprintf("Error saving synonyms: %v", err), nil
		}
		return fmt.Sprintf("Saved synonyms for '%s.%s'.", action.Table, action.Column), nil

	case "business_term":
		if action.Value == "" || action.Definition == "" {
			return "Error: 'value' (term name) and 'definition' are required for business_term.", nil
		}
		if err := t.writer.AddBusinessTerm(ctx, t.dsID, action.Value, action.Definition, action.Synonyms, action.Examples, action.Category); err != nil {
			log.Error("failed to save business term", "term", action.Value, "error", err)
			return fmt.Sprintf("Error saving business term: %v", err), nil
		}
		return fmt.Sprintf("Saved business term '%s'.", action.Value), nil

	default:
		return fmt.Sprintf("Error: unknown context type '%s'. Must be one of: table_description, column_description, column_sample_values, column_synonyms, business_term.", action.Type), nil
	}
}

func (t *SetRichContext) CallCount() int { return t.callCount }

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// stripMarkdownCodeBlock removes ```json ... ``` or ``` ... ``` wrappers that LLMs often add.
func stripMarkdownCodeBlock(s string) string {
	// Check for ```json or ``` prefix
	if strings.HasPrefix(s, "```") {
		// Remove opening ``` line (with optional language tag)
		idx := strings.Index(s, "\n")
		if idx >= 0 {
			s = s[idx+1:]
		} else {
			s = strings.TrimPrefix(s, "```json")
			s = strings.TrimPrefix(s, "```")
		}
		// Remove trailing ```
		if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
		s = strings.TrimSpace(s)
	}
	return s
}
