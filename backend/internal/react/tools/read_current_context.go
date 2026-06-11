package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"atlas/internal/lakebase"
	"atlas/internal/logger"
)

// ReadCurrentContext is a Coordinator tool that reads current Rich Context for a table/column.
type ReadCurrentContext struct {
	repo *lakebase.MySQLRepository
	dsID int64
}

func NewReadCurrentContext(repo *lakebase.MySQLRepository, dsID int64) *ReadCurrentContext {
	return &ReadCurrentContext{repo: repo, dsID: dsID}
}

func (t *ReadCurrentContext) Name() string { return "read_current_context" }
func (t *ReadCurrentContext) Description() string {
	return `Read the current Rich Context for a table or column.
Input: JSON object with fields:
  - "table_name": the table to read (required)
  - "column_name": the column to read (optional; omit for table-level context)
Output: current description, sample_values, synonyms, and metadata.`
}

type readContextInput struct {
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name,omitempty"`
}

func (t *ReadCurrentContext) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "read_current_context", "dsID", t.dsID)

	var inp readContextInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if inp.TableName == "" {
		return "Error: table_name is required.", nil
	}

	log.Info("reading current context", "table", inp.TableName, "column", inp.ColumnName)

	if inp.ColumnName == "" {
		// Table-level context
		tables, err := t.repo.GetTablesByDatasource(ctx, t.dsID)
		if err != nil {
			return fmt.Sprintf("Error: %v", err), nil
		}
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				result, _ := json.MarshalIndent(map[string]interface{}{
					"type":        "table",
					"table_name":  tbl.TableName,
					"description": tbl.Description.String,
					"row_count":   tbl.RowCount,
					"is_expired":  tbl.IsExpired,
				}, "", "  ")
				return string(result), nil
			}
		}
		return fmt.Sprintf("Table '%s' not found in Rich Context.", inp.TableName), nil
	}

	// Column-level context
	cols, err := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}
	for _, col := range cols {
		if col.ColumnName == inp.ColumnName {
			result, _ := json.MarshalIndent(map[string]interface{}{
				"type":           "column",
				"table_name":    col.TableName,
				"column_name":   col.ColumnName,
				"data_type":     col.DataType.String,
				"description":   col.Description.String,
				"sample_values": col.SampleValues.String,
				"synonyms":      col.Synonyms.String,
				"is_expired":    col.IsExpired,
				"is_primary_key": col.IsPrimaryKey,
				"is_foreign_key": col.IsForeignKey,
			}, "", "  ")
			return string(result), nil
		}
	}
	return fmt.Sprintf("Column '%s.%s' not found in Rich Context.", inp.TableName, inp.ColumnName), nil
}
