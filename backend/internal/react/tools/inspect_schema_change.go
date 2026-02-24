package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// InspectSchemaChange is a Coordinator tool that inspects which rc_tables/rc_columns
// are affected by a schema change and returns their current Rich Context.
type InspectSchemaChange struct {
	repo *lakebase.MySQLRepository
	dsID int64
}

func NewInspectSchemaChange(repo *lakebase.MySQLRepository, dsID int64) *InspectSchemaChange {
	return &InspectSchemaChange{repo: repo, dsID: dsID}
}

func (t *InspectSchemaChange) Name() string { return "inspect_schema_change" }
func (t *InspectSchemaChange) Description() string {
	return `Inspect which tables and columns are affected by a schema change and return their current Rich Context.
Input: JSON object with fields:
  - "change_type": one of "table_added", "table_dropped", "column_added", "column_dropped", "column_modified", "fk_added", "fk_dropped"
  - "table_name": the affected table
  - "column_name": the affected column (optional, for column-level changes)
Output: JSON with affected entities and their current Rich Context content.`
}

type inspectInput struct {
	ChangeType string `json:"change_type"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name,omitempty"`
}

func (t *InspectSchemaChange) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "inspect_schema_change", "dsID", t.dsID)

	var inp inspectInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if inp.TableName == "" {
		return "Error: table_name is required.", nil
	}

	log.Info("inspecting schema change", "change_type", inp.ChangeType, "table", inp.TableName, "column", inp.ColumnName)

	type entityInfo struct {
		Type        string `json:"type"` // "table" or "column"
		TableName   string `json:"table_name"`
		ColumnName  string `json:"column_name,omitempty"`
		Description string `json:"description,omitempty"`
		SampleVals  string `json:"sample_values,omitempty"`
		Synonyms    string `json:"synonyms,omitempty"`
		DataType    string `json:"data_type,omitempty"`
	}

	var affected []entityInfo

	switch inp.ChangeType {
	case "table_added":
		// New table — return the table record (likely empty description)
		tables, _ := t.repo.GetTablesByDatasource(ctx, t.dsID)
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				affected = append(affected, entityInfo{
					Type: "table", TableName: tbl.TableName,
					Description: tbl.Description.String,
				})
			}
		}
		// Also return columns
		cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
		for _, col := range cols {
			affected = append(affected, entityInfo{
				Type: "column", TableName: col.TableName, ColumnName: col.ColumnName,
				DataType: col.DataType.String, Description: col.Description.String,
				SampleVals: col.SampleValues.String, Synonyms: col.Synonyms.String,
			})
		}

	case "table_dropped":
		// Dropped table — return what we had
		tables, _ := t.repo.GetTablesByDatasource(ctx, t.dsID)
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				affected = append(affected, entityInfo{
					Type: "table", TableName: tbl.TableName,
					Description: tbl.Description.String,
				})
			}
		}
		cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
		for _, col := range cols {
			affected = append(affected, entityInfo{
				Type: "column", TableName: col.TableName, ColumnName: col.ColumnName,
				Description: col.Description.String,
			})
		}

	case "column_added", "column_dropped", "column_modified":
		// The table itself might need refresh + the specific column
		tables, _ := t.repo.GetTablesByDatasource(ctx, t.dsID)
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				affected = append(affected, entityInfo{
					Type: "table", TableName: tbl.TableName,
					Description: tbl.Description.String,
				})
			}
		}
		if inp.ColumnName != "" {
			cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
			for _, col := range cols {
				if col.ColumnName == inp.ColumnName {
					affected = append(affected, entityInfo{
						Type: "column", TableName: col.TableName, ColumnName: col.ColumnName,
						DataType: col.DataType.String, Description: col.Description.String,
						SampleVals: col.SampleValues.String, Synonyms: col.Synonyms.String,
					})
				}
			}
		}

	case "fk_added", "fk_dropped":
		// FK changes affect the table description and column
		tables, _ := t.repo.GetTablesByDatasource(ctx, t.dsID)
		for _, tbl := range tables {
			if tbl.TableName == inp.TableName {
				affected = append(affected, entityInfo{
					Type: "table", TableName: tbl.TableName,
					Description: tbl.Description.String,
				})
			}
		}
		if inp.ColumnName != "" {
			cols, _ := t.repo.GetColumnsByTable(ctx, t.dsID, inp.TableName)
			for _, col := range cols {
				if col.ColumnName == inp.ColumnName {
					affected = append(affected, entityInfo{
						Type: "column", TableName: col.TableName, ColumnName: col.ColumnName,
						DataType: col.DataType.String, Description: col.Description.String,
					})
				}
			}
		}

	default:
		return fmt.Sprintf("Error: unknown change_type '%s'", inp.ChangeType), nil
	}

	result, _ := json.MarshalIndent(map[string]interface{}{
		"affected_count": len(affected),
		"entities":       affected,
	}, "", "  ")
	return string(result), nil
}
