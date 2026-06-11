// Package agent provides self-maintenance capabilities for ATLAS.
// It includes DDL change detection, context expiration, and automatic updates.
package agent

import (
	"context"
	"strings"
	"time"

	"atlas/internal/lakebase"
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

// ParseDDLStatement attempts to parse a DDL statement and detect the type of change.
// Preserves original table/column name casing for correct lake-base matching.
func ParseDDLStatement(sql string) *SchemaChange {
	sql = strings.TrimSpace(sql)
	upper := strings.ToUpper(sql)

	// ALTER TABLE detection
	if strings.HasPrefix(upper, "ALTER TABLE") {
		return parseDDLAlterTable(sql)
	}

	// CREATE TABLE detection
	if strings.HasPrefix(upper, "CREATE TABLE") {
		tableName := extractTableName(sql, upper, "CREATE TABLE")
		if tableName != "" {
			return &SchemaChange{
				ChangeType: ChangeTypeTableAdded,
				TableName:  tableName,
				DetectedAt: time.Now(),
			}
		}
	}

	// DROP TABLE detection
	if strings.HasPrefix(upper, "DROP TABLE") {
		tableName := extractTableName(sql, upper, "DROP TABLE")
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

// parseDDLAlterTable parses ALTER TABLE statements, preserving original casing.
func parseDDLAlterTable(sql string) *SchemaChange {
	parts := strings.Fields(sql)
	if len(parts) < 4 {
		return nil
	}

	tableName := strings.Trim(parts[2], "`\"'")
	action := strings.ToUpper(parts[3])

	change := &SchemaChange{
		TableName:  tableName,
		DetectedAt: time.Now(),
	}

	switch action {
	case "ADD":
		if len(parts) > 4 {
			kw := strings.ToUpper(parts[4])
			if kw == "COLUMN" && len(parts) > 5 {
				change.ChangeType = ChangeTypeColumnAdded
				change.ColumnName = strings.Trim(parts[5], "`\"'")
			} else if kw == "INDEX" || kw == "KEY" {
				change.ChangeType = ChangeTypeIndexAdded
			} else if kw == "FOREIGN" {
				change.ChangeType = ChangeTypeForeignKeyAdded
			} else if kw == "CONSTRAINT" {
				// ADD CONSTRAINT ... FOREIGN KEY (...)
				change.ChangeType = ChangeTypeForeignKeyAdded
			} else {
				// Might be ADD COLUMN without COLUMN keyword
				change.ChangeType = ChangeTypeColumnAdded
				change.ColumnName = strings.Trim(parts[4], "`\"'")
			}
		}

	case "DROP":
		if len(parts) > 4 {
			kw := strings.ToUpper(parts[4])
			if kw == "COLUMN" && len(parts) > 5 {
				change.ChangeType = ChangeTypeColumnDropped
				change.ColumnName = strings.Trim(parts[5], "`\"'")
			} else if kw == "INDEX" || kw == "KEY" {
				change.ChangeType = ChangeTypeIndexDropped
			} else if kw == "FOREIGN" {
				change.ChangeType = ChangeTypeForeignKeyDropped
			} else {
				change.ChangeType = ChangeTypeColumnDropped
				change.ColumnName = strings.Trim(parts[4], "`\"'")
			}
		}

	case "MODIFY", "CHANGE":
		change.ChangeType = ChangeTypeColumnModified
		if len(parts) > 4 {
			kw := strings.ToUpper(parts[4])
			if kw == "COLUMN" && len(parts) > 5 {
				change.ColumnName = strings.Trim(parts[5], "`\"'")
			} else {
				change.ColumnName = strings.Trim(parts[4], "`\"'")
			}
		}
	}

	if change.ChangeType == "" {
		return nil
	}

	return change
}

// extractTableName extracts table name from DDL statement, preserving original casing.
// origSQL is the original-case SQL, upperSQL is the uppercased version, prefix is the uppercased command prefix.
func extractTableName(origSQL, upperSQL, prefix string) string {
	// Trim the prefix from original by using the length of the prefix
	idx := strings.Index(upperSQL, prefix)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(origSQL[idx+len(prefix):])

	// Handle IF EXISTS / IF NOT EXISTS
	upperRest := strings.ToUpper(rest)
	if strings.HasPrefix(upperRest, "IF") {
		parts := strings.Fields(rest)
		if len(parts) >= 3 && (strings.EqualFold(parts[1], "EXISTS") || strings.EqualFold(parts[1], "NOT")) {
			if strings.EqualFold(parts[1], "NOT") && len(parts) >= 4 {
				rest = strings.Join(parts[3:], " ")
			} else {
				rest = strings.Join(parts[2:], " ")
			}
		}
	}

	parts := strings.Fields(rest)
	if len(parts) > 0 {
		return strings.Trim(parts[0], "`\"'(")
	}

	return ""
}
