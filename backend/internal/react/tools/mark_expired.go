package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// MarkExpired is a Coordinator tool that marks tables/columns as expired.
type MarkExpired struct {
	repo *lakebase.MySQLRepository
	dsID int64
}

func NewMarkExpired(repo *lakebase.MySQLRepository, dsID int64) *MarkExpired {
	return &MarkExpired{repo: repo, dsID: dsID}
}

func (t *MarkExpired) Name() string { return "mark_expired" }
func (t *MarkExpired) Description() string {
	return `Mark tables or columns as expired (their Rich Context is stale and needs refresh).
Input: JSON object with fields:
  - "tables": array of table names to mark as expired (optional)
  - "columns": array of {"table":"...","column":"..."} to mark as expired (optional)
At least one of "tables" or "columns" must be provided.
Output: count of marked entities.`
}

type markExpiredInput struct {
	Tables  []string               `json:"tables,omitempty"`
	Columns []lakebase.TableColumn `json:"columns,omitempty"`
}

func (t *MarkExpired) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "mark_expired", "dsID", t.dsID)

	var inp markExpiredInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if len(inp.Tables) == 0 && len(inp.Columns) == 0 {
		return "Error: at least one of 'tables' or 'columns' must be provided.", nil
	}

	var totalMarked int64

	if len(inp.Tables) > 0 {
		n, err := t.repo.MarkTablesExpired(ctx, t.dsID, inp.Tables)
		if err != nil {
			return fmt.Sprintf("Error marking tables expired: %v", err), nil
		}
		totalMarked += n
		// Also mark all columns of those tables
		n2, err := t.repo.MarkAllColumnsExpiredByTable(ctx, t.dsID, inp.Tables)
		if err != nil {
			log.Warn("failed to mark columns by table", "error", err)
		}
		totalMarked += n2
		log.Info("marked tables expired", "tables", inp.Tables, "count", n+n2)
	}

	if len(inp.Columns) > 0 {
		n, err := t.repo.MarkColumnsExpired(ctx, t.dsID, inp.Columns)
		if err != nil {
			return fmt.Sprintf("Error marking columns expired: %v", err), nil
		}
		totalMarked += n
		log.Info("marked columns expired", "count", n)
	}

	return fmt.Sprintf("Marked %d entities as expired.", totalMarked), nil
}
