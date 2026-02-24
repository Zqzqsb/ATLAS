package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// ClearExpired is an Executor tool that clears the expired flag after RC has been refreshed.
type ClearExpired struct {
	repo *lakebase.MySQLRepository
	dsID int64
}

func NewClearExpired(repo *lakebase.MySQLRepository, dsID int64) *ClearExpired {
	return &ClearExpired{repo: repo, dsID: dsID}
}

func (t *ClearExpired) Name() string { return "clear_expired" }
func (t *ClearExpired) Description() string {
	return `Clear the expired flag on tables and columns after their Rich Context has been refreshed.
Input: JSON object with fields:
  - "tables": array of table names to clear (optional)
  - "columns": array of {"table":"...","column":"..."} to clear (optional)
At least one of "tables" or "columns" must be provided.
Output: confirmation message.`
}

type clearExpiredInput struct {
	Tables  []string               `json:"tables,omitempty"`
	Columns []lakebase.TableColumn `json:"columns,omitempty"`
}

func (t *ClearExpired) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "clear_expired", "dsID", t.dsID)

	var inp clearExpiredInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if len(inp.Tables) == 0 && len(inp.Columns) == 0 {
		return "Error: at least one of 'tables' or 'columns' must be provided.", nil
	}

	var cleared int

	if len(inp.Tables) > 0 {
		if err := t.repo.ClearTableExpired(ctx, t.dsID, inp.Tables); err != nil {
			return fmt.Sprintf("Error clearing table expired: %v", err), nil
		}
		cleared += len(inp.Tables)
		log.Info("cleared table expired flags", "tables", inp.Tables)
	}

	if len(inp.Columns) > 0 {
		if err := t.repo.ClearColumnExpired(ctx, t.dsID, inp.Columns); err != nil {
			return fmt.Sprintf("Error clearing column expired: %v", err), nil
		}
		cleared += len(inp.Columns)
		log.Info("cleared column expired flags", "count", len(inp.Columns))
	}

	return fmt.Sprintf("Cleared expired flag on %d entities.", cleared), nil
}
