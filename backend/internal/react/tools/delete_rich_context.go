package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// DeleteRichContext is an Executor tool that deletes Rich Context for dropped entities.
type DeleteRichContext struct {
	repo       *lakebase.MySQLRepository
	vectorRepo *lakebase.MySQLVectorRepository
	dsID       int64
}

func NewDeleteRichContext(repo *lakebase.MySQLRepository, vectorRepo *lakebase.MySQLVectorRepository, dsID int64) *DeleteRichContext {
	return &DeleteRichContext{repo: repo, vectorRepo: vectorRepo, dsID: dsID}
}

func (t *DeleteRichContext) Name() string { return "delete_rich_context" }
func (t *DeleteRichContext) Description() string {
	return `Delete Rich Context for a dropped table or column.
Input: JSON object with fields:
  - "target": "table" or "column"
  - "table_name": the table name
  - "column_name": the column name (required for column targets)
Cleans up rc_tables/rc_columns records and soft-deletes related embeddings.
Output: confirmation message.`
}

type deleteRCInput struct {
	Target     string `json:"target"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name,omitempty"`
}

func (t *DeleteRichContext) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "delete_rich_context", "dsID", t.dsID)

	var inp deleteRCInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if inp.Target == "" || inp.TableName == "" {
		return "Error: 'target' and 'table_name' are required.", nil
	}

	log.Info("deleting rich context", "target", inp.Target, "table", inp.TableName, "column", inp.ColumnName)

	switch inp.Target {
	case "table":
		// Delete table + all columns + soft-delete embeddings
		// First, find the table ID for embedding soft-delete
		tables, _ := t.repo.GetTablesByDatasource(ctx, t.dsID)
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				if t.vectorRepo != nil {
					_ = t.vectorRepo.SoftDeleteEmbedding(ctx, t.dsID, lakebase.EntityTypeTable, tbl.ID)
				}
			}
		}
		// Soft-delete column embeddings
		cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
		for _, col := range cols {
			if t.vectorRepo != nil {
				_ = t.vectorRepo.SoftDeleteEmbedding(ctx, t.dsID, lakebase.EntityTypeColumn, col.ID)
			}
		}
		// Delete from rc_tables + rc_columns
		if err := t.repo.DeleteTableByName(ctx, t.dsID, inp.TableName); err != nil {
			return fmt.Sprintf("Error deleting table: %v", err), nil
		}
		return fmt.Sprintf("Deleted Rich Context for table '%s' and all its columns. Embeddings soft-deleted.", inp.TableName), nil

	case "column":
		if inp.ColumnName == "" {
			return "Error: 'column_name' is required for column target.", nil
		}
		// Find column ID for embedding soft-delete
		cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
		for _, col := range cols {
			if col.ColumnName == inp.ColumnName {
				if t.vectorRepo != nil {
					_ = t.vectorRepo.SoftDeleteEmbedding(ctx, t.dsID, lakebase.EntityTypeColumn, col.ID)
				}
			}
		}
		if err := t.repo.DeleteColumnByName(ctx, t.dsID, inp.TableName, inp.ColumnName); err != nil {
			return fmt.Sprintf("Error deleting column: %v", err), nil
		}
		return fmt.Sprintf("Deleted Rich Context for column '%s.%s'. Embedding soft-deleted.", inp.TableName, inp.ColumnName), nil

	default:
		return fmt.Sprintf("Error: invalid target '%s'. Must be 'table' or 'column'.", inp.Target), nil
	}
}
