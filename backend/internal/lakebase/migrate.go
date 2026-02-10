// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in LUCID system.
package lakebase

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"lucid/internal/logger"
)

//go:embed schema/01_init_lakebase.sql
var initSQL string

// columnDef represents a column definition parsed from init SQL.
type columnDef struct {
	Name         string
	TypeSQL      string // e.g. "INT AUTO_INCREMENT", "VARCHAR(255) NOT NULL"
	FullDef      string // full original line (minus trailing comma)
	AfterColumn  string // the column defined immediately before this one (for ADD COLUMN ... AFTER)
}

// tableDef represents a table definition parsed from init SQL.
type tableDef struct {
	Name    string
	Columns []columnDef
}

// renameEntry records a historical column rename so that the migration can
// detect the old name and CHANGE it to the new name instead of dropping+adding.
type renameEntry struct {
	Table   string
	OldName string
	NewName string
	TypeSQL string // target type for the CHANGE COLUMN statement
}

// knownRenames lists historical column renames. When auto-migration detects
// that the old column exists but the new one doesn't, it issues CHANGE COLUMN.
// Add entries here when you rename a column in init SQL.
var knownRenames = []renameEntry{
	{Table: "rc_embeddings", OldName: "text_content", NewName: "entity_text", TypeSQL: "TEXT NOT NULL COMMENT 'Text that was embedded'"},
	{Table: "rc_embeddings", OldName: "model", NewName: "embedding_model", TypeSQL: "VARCHAR(100) DEFAULT 'doubao-embedding-large-text-250515'"},
}

// AutoMigrate parses the embedded init SQL, compares with the live database
// schema (via information_schema), and automatically applies:
//   - ADD COLUMN   for columns in init SQL but not in DB
//   - MODIFY COLUMN for columns that exist but have a different type
//   - CHANGE COLUMN for known renames (old column exists, new doesn't)
//
// It never drops columns (safety). Each operation is idempotent.
func AutoMigrate(ctx context.Context, pool *ConnectionPool) error {
	db, err := pool.DB()
	if err != nil {
		return fmt.Errorf("auto-migrate: failed to get db: %w", err)
	}

	// 1. Parse init SQL to get target schema
	targetTables := parseInitSQL(initSQL)
	if len(targetTables) == 0 {
		logger.L().Warn("AutoMigrate: no tables parsed from init SQL")
		return nil
	}

	applied := 0

	// 2. Handle known renames first (must happen before add-column check)
	for _, r := range knownRenames {
		hasNew, _ := columnExists(ctx, db, r.Table, r.NewName)
		if hasNew {
			continue // already renamed
		}
		hasOld, _ := columnExists(ctx, db, r.Table, r.OldName)
		if !hasOld {
			continue // neither exists, table may be new (init SQL already correct)
		}
		stmt := fmt.Sprintf("ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` %s",
			r.Table, r.OldName, r.NewName, r.TypeSQL)
		logger.L().Info("AutoMigrate: renaming column", "table", r.Table, "old", r.OldName, "new", r.NewName)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("auto-migrate: rename %s.%s→%s failed: %w", r.Table, r.OldName, r.NewName, err)
		}
		applied++
	}

	// 3. For each target table, diff columns and apply ADD/MODIFY
	for _, tbl := range targetTables {
		// Check if table exists at all
		exists, err := tableExists(ctx, db, tbl.Name)
		if err != nil {
			logger.L().Warn("AutoMigrate: failed to check table", "table", tbl.Name, "error", err)
			continue
		}
		if !exists {
			// Table doesn't exist — init SQL should have created it.
			// If it didn't (e.g. the SQL was never run), we skip;
			// we don't want to create tables from migration.
			logger.L().Debug("AutoMigrate: table does not exist, skipping", "table", tbl.Name)
			continue
		}

		// Get actual columns from information_schema
		actualCols, err := getActualColumns(ctx, db, tbl.Name)
		if err != nil {
			logger.L().Warn("AutoMigrate: failed to read columns", "table", tbl.Name, "error", err)
			continue
		}

		for _, col := range tbl.Columns {
			actual, found := actualCols[strings.ToLower(col.Name)]

			if !found {
				// Column missing — ADD COLUMN
				afterClause := ""
				if col.AfterColumn != "" {
					afterClause = fmt.Sprintf(" AFTER `%s`", col.AfterColumn)
				}
				stmt := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s%s",
					tbl.Name, col.Name, col.TypeSQL, afterClause)
				logger.L().Info("AutoMigrate: adding column", "table", tbl.Name, "column", col.Name)
				if _, err := db.ExecContext(ctx, stmt); err != nil {
					logger.L().Warn("AutoMigrate: ADD COLUMN failed", "table", tbl.Name, "column", col.Name, "error", err)
				} else {
					applied++
				}
				continue
			}

			// Column exists — check if type needs modification
			// We compare the base type (simplified) to avoid false positives
			if needsModify(col.TypeSQL, actual.typeSQL) {
				stmt := fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN `%s` %s",
					tbl.Name, col.Name, col.TypeSQL)
				logger.L().Info("AutoMigrate: modifying column", "table", tbl.Name, "column", col.Name, "from", actual.typeSQL, "to", col.TypeSQL)
				if _, err := db.ExecContext(ctx, stmt); err != nil {
					logger.L().Warn("AutoMigrate: MODIFY COLUMN failed", "table", tbl.Name, "column", col.Name, "error", err)
				} else {
					applied++
				}
			}
		}
	}

	if applied > 0 {
		logger.L().Info("AutoMigrate: schema changes applied", "count", applied)
	} else {
		logger.L().Debug("AutoMigrate: schema is up-to-date")
	}

	return nil
}

// ====================================================================
// SQL Parsing
// ====================================================================

// parseInitSQL extracts CREATE TABLE definitions from the init SQL text.
func parseInitSQL(sqlText string) []tableDef {
	var tables []tableDef

	// Match CREATE TABLE blocks
	re := regexp.MustCompile(`(?is)CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\((.*?)\)\s*ENGINE`)
	matches := re.FindAllStringSubmatch(sqlText, -1)

	for _, m := range matches {
		tableName := m[1]
		body := m[2]

		cols := parseColumns(body)
		tables = append(tables, tableDef{Name: tableName, Columns: cols})
	}

	return tables
}

// parseColumns parses column definitions from a CREATE TABLE body.
// It skips constraints (FOREIGN KEY, INDEX, UNIQUE KEY, PRIMARY KEY, VECTOR INDEX, FULLTEXT INDEX).
func parseColumns(body string) []columnDef {
	var cols []columnDef
	lines := strings.Split(body, "\n")

	// Patterns for non-column lines
	constraintPrefixes := []string{
		"FOREIGN KEY", "INDEX ", "UNIQUE KEY", "PRIMARY KEY",
		"VECTOR INDEX", "FULLTEXT INDEX", "KEY ",
	}

	prevCol := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip constraint/index lines
		upper := strings.ToUpper(line)
		isConstraint := false
		for _, prefix := range constraintPrefixes {
			if strings.HasPrefix(upper, prefix) {
				isConstraint = true
				break
			}
		}
		if isConstraint {
			continue
		}

		// Try to parse as column definition: name TYPE [rest...]
		// Column names may be backtick-quoted
		colRe := regexp.MustCompile(`^` + "`?" + `(\w+)` + "`?" + `\s+(.+?)(?:,\s*)?$`)
		match := colRe.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		colName := match[1]
		typePart := strings.TrimRight(match[2], ",")

		// Skip if "name" is a SQL keyword that starts a constraint
		upperName := strings.ToUpper(colName)
		if upperName == "FOREIGN" || upperName == "INDEX" || upperName == "UNIQUE" ||
			upperName == "PRIMARY" || upperName == "VECTOR" || upperName == "FULLTEXT" ||
			upperName == "KEY" || upperName == "CONSTRAINT" {
			continue
		}

		cols = append(cols, columnDef{
			Name:        colName,
			TypeSQL:     typePart,
			FullDef:     line,
			AfterColumn: prevCol,
		})
		prevCol = colName
	}

	return cols
}

// ====================================================================
// Information Schema Queries
// ====================================================================

type actualColumn struct {
	name    string
	typeSQL string // COLUMN_TYPE from information_schema (e.g. "int(11)", "varchar(255)")
}

// tableExists checks if a table exists in the current database.
func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	query := `SELECT COUNT(*) FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
	var count int
	err := db.QueryRowContext(ctx, query, table).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check table %s: %w", table, err)
	}
	return count > 0, nil
}

// columnExists checks if a column exists in a table.
func columnExists(ctx context.Context, db *sql.DB, table, column string) (bool, error) {
	query := `SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`
	var count int
	err := db.QueryRowContext(ctx, query, table, column).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	return count > 0, nil
}

// getActualColumns returns all columns for a table from information_schema.
func getActualColumns(ctx context.Context, db *sql.DB, table string) (map[string]actualColumn, error) {
	query := `SELECT COLUMN_NAME, COLUMN_TYPE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`
	rows, err := db.QueryContext(ctx, query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]actualColumn)
	for rows.Next() {
		var name, colType string
		if err := rows.Scan(&name, &colType); err != nil {
			return nil, err
		}
		result[strings.ToLower(name)] = actualColumn{name: name, typeSQL: colType}
	}
	return result, rows.Err()
}

// ====================================================================
// Type Comparison
// ====================================================================

// needsModify compares the target type definition (from init SQL) with the
// actual column type (from information_schema). Returns true if they differ
// in a meaningful way (base type mismatch).
//
// We do a simplified comparison: extract the base type keyword + size from both
// and compare. This avoids false positives from COMMENT, DEFAULT, etc.
func needsModify(targetDef, actualType string) bool {
	targetBase := normalizeType(targetDef)
	actualBase := normalizeType(actualType)
	return targetBase != actualBase
}

// normalizeType extracts and normalizes a base type string for comparison.
// E.g. "VARCHAR(255) NOT NULL COMMENT '...'" → "varchar(255)"
//      "int(11)"                              → "int"
//      "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"  → "timestamp"
//      "ENUM('a','b','c') NOT NULL"           → "enum('a','b','c')"
//      "DECIMAL(3,2)"                         → "decimal(3,2)"
//      "VECTOR(2048)"                         → "vector(2048)"
func normalizeType(t string) string {
	t = strings.TrimSpace(t)
	if t == "" {
		return ""
	}

	// Convert to lower for comparison
	lower := strings.ToLower(t)

	// Handle ENUM specially — extract ENUM(...)
	if strings.HasPrefix(lower, "enum(") || strings.HasPrefix(lower, "enum (") {
		enumRe := regexp.MustCompile(`(?i)^enum\s*\(([^)]+)\)`)
		if m := enumRe.FindString(lower); m != "" {
			return strings.ReplaceAll(m, " ", "")
		}
	}

	// Extract TYPE or TYPE(size) pattern
	typeRe := regexp.MustCompile(`(?i)^(\w+)(?:\(([^)]+)\))?`)
	match := typeRe.FindStringSubmatch(lower)
	if match == nil {
		return lower
	}

	baseType := match[1]
	size := match[2]

	// Normalize away display widths for integer types (MariaDB specific)
	// e.g. int(11) → int, tinyint(1) → tinyint (except when it matters)
	intTypes := map[string]bool{
		"int": true, "integer": true, "bigint": true,
		"smallint": true, "mediumint": true,
	}
	if intTypes[baseType] {
		// Keep tinyint(1) as-is since it's used for BOOLEAN
		if baseType == "tinyint" && size == "1" {
			return "tinyint(1)"
		}
		return baseType
	}

	if size != "" {
		return baseType + "(" + size + ")"
	}
	return baseType
}
