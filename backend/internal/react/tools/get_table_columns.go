package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"atlas/internal/lakebase"
	"atlas/internal/logger"
)

// GetTableColumns is a Coordinator tool that returns all columns for a table.
type GetTableColumns struct {
	repo *lakebase.MySQLRepository
	dsID int64
}

func NewGetTableColumns(repo *lakebase.MySQLRepository, dsID int64) *GetTableColumns {
	return &GetTableColumns{repo: repo, dsID: dsID}
}

func (t *GetTableColumns) Name() string { return "get_table_columns" }
func (t *GetTableColumns) Description() string {
	return `Get all columns of a table with their schema and Rich Context.
Input: JSON object with field "table_name" (required).
Output: list of columns with data_type, description, sample_values, synonyms, etc.`
}

type getColumnsInput struct {
	TableName string `json:"table_name"`
}

func (t *GetTableColumns) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "get_table_columns", "dsID", t.dsID)

	var inp getColumnsInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if inp.TableName == "" {
		return "Error: table_name is required.", nil
	}

	log.Info("getting table columns", "table", inp.TableName)

	cols, err := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	type colSummary struct {
		ColumnName   string `json:"column_name"`
		DataType     string `json:"data_type"`
		Description  string `json:"description,omitempty"`
		SampleValues string `json:"sample_values,omitempty"`
		Synonyms     string `json:"synonyms,omitempty"`
		IsPrimaryKey bool   `json:"is_primary_key"`
		IsForeignKey bool   `json:"is_foreign_key"`
		IsExpired    bool   `json:"is_expired"`
	}

	summaries := make([]colSummary, len(cols))
	for i, col := range cols {
		summaries[i] = colSummary{
			ColumnName:   col.ColumnName,
			DataType:     col.DataType.String,
			Description:  col.Description.String,
			SampleValues: col.SampleValues.String,
			Synonyms:     col.Synonyms.String,
			IsPrimaryKey: col.IsPrimaryKey,
			IsForeignKey: col.IsForeignKey,
			IsExpired:    col.IsExpired,
		}
	}

	result, _ := json.MarshalIndent(map[string]interface{}{
		"table_name": inp.TableName,
		"columns":    summaries,
		"count":      len(summaries),
	}, "", "  ")
	return string(result), nil
}
