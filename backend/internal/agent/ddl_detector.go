// Package agent provides self-maintenance capabilities for LUCID.
// It includes DDL change detection, context expiration, and automatic updates.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"lucid/internal/lakebase"
)

// SchemaChange represents a detected schema change
type SchemaChange struct {
	ChangeType    SchemaChangeType `json:"change_type"`
	TableName     string           `json:"table_name"`
	ColumnName    string           `json:"column_name,omitempty"`
	OldDefinition string           `json:"old_definition,omitempty"`
	NewDefinition string           `json:"new_definition,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	DetectedAt    time.Time        `json:"detected_at"`
}

// SchemaChangeType represents the type of schema change
type SchemaChangeType string

const (
	ChangeTypeTableAdded    SchemaChangeType = "table_added"
	ChangeTypeTableDropped  SchemaChangeType = "table_dropped"
	ChangeTypeColumnAdded   SchemaChangeType = "column_added"
	ChangeTypeColumnDropped SchemaChangeType = "column_dropped"
	ChangeTypeColumnModified SchemaChangeType = "column_modified"
	ChangeTypeIndexAdded    SchemaChangeType = "index_added"
	ChangeTypeIndexDropped  SchemaChangeType = "index_dropped"
	ChangeTypeForeignKeyAdded SchemaChangeType = "fk_added"
	ChangeTypeForeignKeyDropped SchemaChangeType = "fk_dropped"
)

// DDLDetector detects schema changes by comparing current schema with stored metadata
type DDLDetector struct {
	repo *lakebase.MySQLRepository
}

// NewDDLDetector creates a new DDL detector
func NewDDLDetector(repo *lakebase.MySQLRepository) *DDLDetector {
	return &DDLDetector{repo: repo}
}

// SchemaSnapshot represents the current state of a table's schema
type SchemaSnapshot struct {
	TableName  string
	Columns    map[string]*ColumnSnapshot
	FetchedAt  time.Time
}

// ColumnSnapshot represents the current state of a column
type ColumnSnapshot struct {
	Name         string
	DataType     string
	IsPrimaryKey bool
	IsForeignKey bool
	FKRefTable   string
	FKRefColumn  string
	Nullable     bool
	DefaultValue string
}

// DetectChanges compares current schema with stored metadata and returns changes
func (d *DDLDetector) DetectChanges(ctx context.Context, dsID int64, currentSchema map[string]*SchemaSnapshot) ([]SchemaChange, error) {
	var changes []SchemaChange

	// Get stored schema from lake-base (rc_columns)
	storedColumns, err := d.repo.GetColumnsByDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored schema: %w", err)
	}

	// Build stored schema map
	storedSchema := buildStoredSchemaMapFromColumns(storedColumns)

	// Detect changes
	now := time.Now()

	// Check for new tables and modified/dropped columns
	for tableName, currentTable := range currentSchema {
		storedTable, exists := storedSchema[tableName]
		if !exists {
			// New table detected
			changes = append(changes, SchemaChange{
				ChangeType: ChangeTypeTableAdded,
				TableName:  tableName,
				Details: map[string]interface{}{
					"column_count": len(currentTable.Columns),
				},
				DetectedAt: now,
			})
			continue
		}

		// Check for column changes
		columnChanges := d.detectColumnChanges(tableName, storedTable, currentTable, now)
		changes = append(changes, columnChanges...)
	}

	// Check for dropped tables
	for tableName := range storedSchema {
		if _, exists := currentSchema[tableName]; !exists {
			changes = append(changes, SchemaChange{
				ChangeType: ChangeTypeTableDropped,
				TableName:  tableName,
				DetectedAt: now,
			})
		}
	}

	return changes, nil
}

// detectColumnChanges detects changes between stored and current columns
func (d *DDLDetector) detectColumnChanges(tableName string, stored, current *SchemaSnapshot, now time.Time) []SchemaChange {
	var changes []SchemaChange

	// Check for new and modified columns
	for colName, currentCol := range current.Columns {
		storedCol, exists := stored.Columns[colName]
		if !exists {
			// New column detected
			changes = append(changes, SchemaChange{
				ChangeType:    ChangeTypeColumnAdded,
				TableName:     tableName,
				ColumnName:    colName,
				NewDefinition: formatColumnDefinition(currentCol),
				Details: map[string]interface{}{
					"data_type":      currentCol.DataType,
					"nullable":       currentCol.Nullable,
					"is_primary_key": currentCol.IsPrimaryKey,
				},
				DetectedAt: now,
			})
			continue
		}

		// Check for modifications
		if isColumnModified(storedCol, currentCol) {
			changes = append(changes, SchemaChange{
				ChangeType:    ChangeTypeColumnModified,
				TableName:     tableName,
				ColumnName:    colName,
				OldDefinition: formatColumnDefinition(storedCol),
				NewDefinition: formatColumnDefinition(currentCol),
				Details: map[string]interface{}{
					"old_type": storedCol.DataType,
					"new_type": currentCol.DataType,
					"old_nullable": storedCol.Nullable,
					"new_nullable": currentCol.Nullable,
				},
				DetectedAt: now,
			})
		}

		// Check for FK changes
		if storedCol.IsForeignKey != currentCol.IsForeignKey {
			if currentCol.IsForeignKey {
				changes = append(changes, SchemaChange{
					ChangeType: ChangeTypeForeignKeyAdded,
					TableName:  tableName,
					ColumnName: colName,
					Details: map[string]interface{}{
						"ref_table":  currentCol.FKRefTable,
						"ref_column": currentCol.FKRefColumn,
					},
					DetectedAt: now,
				})
			} else {
				changes = append(changes, SchemaChange{
					ChangeType: ChangeTypeForeignKeyDropped,
					TableName:  tableName,
					ColumnName: colName,
					Details: map[string]interface{}{
						"old_ref_table":  storedCol.FKRefTable,
						"old_ref_column": storedCol.FKRefColumn,
					},
					DetectedAt: now,
				})
			}
		}
	}

	// Check for dropped columns
	for colName := range stored.Columns {
		if _, exists := current.Columns[colName]; !exists {
			changes = append(changes, SchemaChange{
				ChangeType:    ChangeTypeColumnDropped,
				TableName:     tableName,
				ColumnName:    colName,
				OldDefinition: formatColumnDefinition(stored.Columns[colName]),
				DetectedAt:    now,
			})
		}
	}

	return changes
}

// buildStoredSchemaMapFromColumns builds a map from rc_columns data
func buildStoredSchemaMapFromColumns(columns []*lakebase.ColumnInfo) map[string]*SchemaSnapshot {
	result := make(map[string]*SchemaSnapshot)

	for _, col := range columns {
		if _, exists := result[col.TableName]; !exists {
			result[col.TableName] = &SchemaSnapshot{
				TableName: col.TableName,
				Columns:   make(map[string]*ColumnSnapshot),
			}
		}

		snapshot := &ColumnSnapshot{
			Name:         col.ColumnName,
			DataType:     col.DataType.String,
			IsPrimaryKey: col.IsPrimaryKey,
			IsForeignKey: col.IsForeignKey,
			Nullable:     col.IsNullable,
		}
		if col.ForeignKeyInfo != nil {
			snapshot.FKRefTable = col.ForeignKeyInfo.RefTableName
			snapshot.FKRefColumn = col.ForeignKeyInfo.RefColumnName
		}

		result[col.TableName].Columns[col.ColumnName] = snapshot
	}

	return result
}

// isColumnModified checks if a column has been modified
func isColumnModified(stored, current *ColumnSnapshot) bool {
	// Normalize data types for comparison
	storedType := normalizeDataType(stored.DataType)
	currentType := normalizeDataType(current.DataType)

	if storedType != currentType {
		return true
	}
	if stored.Nullable != current.Nullable {
		return true
	}
	if stored.IsPrimaryKey != current.IsPrimaryKey {
		return true
	}
	if stored.DefaultValue != current.DefaultValue {
		return true
	}
	return false
}

// normalizeDataType normalizes data type for comparison
func normalizeDataType(dt string) string {
	dt = strings.ToUpper(strings.TrimSpace(dt))
	// Remove length specifications for comparison
	if idx := strings.Index(dt, "("); idx > 0 {
		dt = dt[:idx]
	}
	return dt
}

// formatColumnDefinition formats column definition for logging
func formatColumnDefinition(col *ColumnSnapshot) string {
	def := fmt.Sprintf("%s %s", col.Name, col.DataType)
	if col.IsPrimaryKey {
		def += " PRIMARY KEY"
	}
	if !col.Nullable {
		def += " NOT NULL"
	}
	if col.DefaultValue != "" {
		def += fmt.Sprintf(" DEFAULT %s", col.DefaultValue)
	}
	return def
}

// GetAffectedContextIDs returns context IDs that should be marked as expired due to schema changes
func (d *DDLDetector) GetAffectedContextIDs(ctx context.Context, dsID int64, changes []SchemaChange) ([]int64, error) {
	var affectedIDs []int64

	for _, change := range changes {
		// Get context entries for the affected table
		contexts, err := d.repo.GetContextByTable(ctx, dsID, change.TableName)
		if err != nil {
			continue
		}

		for _, bc := range contexts {
			// Check if this context is affected by the change
			if d.isContextAffected(bc, change) {
				affectedIDs = append(affectedIDs, bc.ID)
			}
		}
	}

	return affectedIDs, nil
}

// isContextAffected determines if a context entry is affected by a schema change
func (d *DDLDetector) isContextAffected(bc *lakebase.BusinessContext, change SchemaChange) bool {
	switch change.ChangeType {
	case ChangeTypeTableDropped:
		// All context for dropped table is affected
		return true

	case ChangeTypeColumnDropped:
		// Context for dropped column is affected
		if bc.ColumnName.Valid && bc.ColumnName.String == change.ColumnName {
			return true
		}
		// Table-level context might also need review
		if !bc.ColumnName.Valid {
			return true
		}

	case ChangeTypeColumnAdded:
		// Table-level context should be reviewed for new columns
		if !bc.ColumnName.Valid {
			return true
		}

	case ChangeTypeColumnModified:
		// Context for modified column needs review
		if bc.ColumnName.Valid && bc.ColumnName.String == change.ColumnName {
			return true
		}
		// Data quality context might need review for type changes
		if bc.ContextType == lakebase.ContextTypeDataQuality {
			return true
		}

	case ChangeTypeForeignKeyAdded, ChangeTypeForeignKeyDropped:
		// Join hints should be reviewed
		if bc.ContextType == lakebase.ContextTypeJoinHint {
			return true
		}
		if bc.ColumnName.Valid && bc.ColumnName.String == change.ColumnName {
			return true
		}
	}

	return false
}

// SchemaChangeToJSON converts a schema change to JSON for logging
func SchemaChangeToJSON(change SchemaChange) json.RawMessage {
	data, _ := json.Marshal(change)
	return data
}

// ParseDDLStatement attempts to parse a DDL statement and detect the type of change
// This is useful for immediate detection when DDL is executed through the system
func ParseDDLStatement(sql string) *SchemaChange {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	// ALTER TABLE detection
	if strings.HasPrefix(sql, "ALTER TABLE") {
		return parseDDLAlterTable(sql)
	}

	// CREATE TABLE detection
	if strings.HasPrefix(sql, "CREATE TABLE") {
		tableName := extractTableName(sql, "CREATE TABLE")
		if tableName != "" {
			return &SchemaChange{
				ChangeType: ChangeTypeTableAdded,
				TableName:  tableName,
				DetectedAt: time.Now(),
			}
		}
	}

	// DROP TABLE detection
	if strings.HasPrefix(sql, "DROP TABLE") {
		tableName := extractTableName(sql, "DROP TABLE")
		if tableName != "" {
			return &SchemaChange{
				ChangeType: ChangeTypeTableDropped,
				TableName:  tableName,
				DetectedAt: time.Now(),
			}
		}
	}

	return nil
}

// parseDDLAlterTable parses ALTER TABLE statements
func parseDDLAlterTable(sql string) *SchemaChange {
	// Extract table name
	parts := strings.Fields(sql)
	if len(parts) < 4 {
		return nil
	}

	tableName := strings.Trim(parts[2], "`\"'")
	action := parts[3]

	change := &SchemaChange{
		TableName:  tableName,
		DetectedAt: time.Now(),
	}

	switch action {
	case "ADD":
		if len(parts) > 4 {
			if parts[4] == "COLUMN" && len(parts) > 5 {
				change.ChangeType = ChangeTypeColumnAdded
				change.ColumnName = strings.Trim(parts[5], "`\"'")
			} else if parts[4] == "INDEX" || parts[4] == "KEY" {
				change.ChangeType = ChangeTypeIndexAdded
			} else if parts[4] == "FOREIGN" {
				change.ChangeType = ChangeTypeForeignKeyAdded
			} else {
				// Might be ADD COLUMN without COLUMN keyword
				change.ChangeType = ChangeTypeColumnAdded
				change.ColumnName = strings.Trim(parts[4], "`\"'")
			}
		}

	case "DROP":
		if len(parts) > 4 {
			if parts[4] == "COLUMN" && len(parts) > 5 {
				change.ChangeType = ChangeTypeColumnDropped
				change.ColumnName = strings.Trim(parts[5], "`\"'")
			} else if parts[4] == "INDEX" || parts[4] == "KEY" {
				change.ChangeType = ChangeTypeIndexDropped
			} else if parts[4] == "FOREIGN" {
				change.ChangeType = ChangeTypeForeignKeyDropped
			} else {
				change.ChangeType = ChangeTypeColumnDropped
				change.ColumnName = strings.Trim(parts[4], "`\"'")
			}
		}

	case "MODIFY", "CHANGE":
		change.ChangeType = ChangeTypeColumnModified
		if len(parts) > 4 {
			change.ColumnName = strings.Trim(parts[4], "`\"'")
		}
	}

	if change.ChangeType == "" {
		return nil
	}

	return change
}

// extractTableName extracts table name from DDL statement
func extractTableName(sql, prefix string) string {
	sql = strings.TrimPrefix(sql, prefix)
	sql = strings.TrimSpace(sql)

	// Handle IF EXISTS / IF NOT EXISTS
	if strings.HasPrefix(sql, "IF") {
		parts := strings.Fields(sql)
		if len(parts) >= 3 {
			sql = strings.Join(parts[2:], " ")
		}
	}

	parts := strings.Fields(sql)
	if len(parts) > 0 {
		return strings.Trim(parts[0], "`\"'(")
	}

	return ""
}
