// Package scenarios provides pre-built ReAct engine configurations for different use cases.
package scenarios

import (
	"fmt"
	"strings"

	"lucid/interfaces"
	"lucid/internal/lakebase"
	"lucid/internal/react"
	reacttools "lucid/internal/react/tools"

	lctools "github.com/tmc/langchaingo/tools"
)

// RCGenConfig holds parameters for Rich Context generation.
type RCGenConfig struct {
	DatasourceID  int64
	Tables        []*lakebase.TableInfo
	Columns       []*lakebase.ColumnInfo
	Relations     []*lakebase.Relation
	MaxIterations int
	MinIterations int
	Force         bool // Force regenerate even if description already exists
	StepCallback  react.StepCallback
}

// BuildRCGenEngine creates a ReAct EngineConfig for Rich Context generation.
// The agent is given execute_sql + set_rich_context tools and told to explore the database
// then write structured context for each table and column.
func BuildRCGenEngine(
	businessDB interfaces.DBAdapter,
	rcWriter reacttools.RCWriter,
	cfg RCGenConfig,
) *react.EngineConfig {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 15
	}
	if cfg.MinIterations <= 0 {
		cfg.MinIterations = 3
	}

	// Tools
	sqlTool := reacttools.NewExecuteSQL(businessDB)
	rcTool := reacttools.NewSetRichContext(rcWriter, cfg.DatasourceID)

	toolsList := []lctools.Tool{sqlTool, rcTool}

	// Build the system prompt
	prompt := buildRCGenPrompt(cfg)

	return &react.EngineConfig{
		MaxIterations: cfg.MaxIterations,
		MinIterations: cfg.MinIterations,
		SystemPrompt:  prompt,
		Tools:         toolsList,
		StepCallback:  cfg.StepCallback,
		LogMode:       "simple",
		Verbose:       true,
	}
}

func buildRCGenPrompt(cfg RCGenConfig) string {
	var sb strings.Builder

	sb.WriteString(`You are a database analyst agent. Your mission is to explore a database and generate Rich Context metadata that will help a Text-to-SQL system understand the database better.

## Your Goal
Explore the database tables, understand their business purpose, discover data patterns, and save Rich Context using the set_rich_context tool.

## Available Tools
- execute_sql: Run SELECT/SHOW/DESCRIBE queries to explore the database
- set_rich_context: Save discovered context (table descriptions, column descriptions, sample values, synonyms, business terms)

## What Rich Context to Generate
For each table, you should discover and save:

1. **Table Description** — A 2-3 sentence summary of the table's business purpose, key columns, and any important data characteristics.
   Use: {"type": "table_description", "table": "...", "value": "..."}

2. **Column Description** — For each non-trivial column, a concise description of its semantic meaning.
   Use: {"type": "column_description", "table": "...", "column": "...", "value": "..."}

3. **Sample Values** — For text/enum-like columns with limited distinct values, list the actual values found.
   Use: {"type": "column_sample_values", "table": "...", "column": "...", "value": "val1, val2, val3"}

4. **Column Synonyms** — Alternative names someone might use to refer to this column in natural language.
   Use: {"type": "column_synonyms", "table": "...", "column": "...", "value": "synonym1, synonym2"}

5. **Business Terms** — Domain-specific terms that appear in the data (e.g., "churn rate", "GMV").
   Use: {"type": "business_term", "value": "term name", "definition": "...", "category": "...", "synonyms": "...", "examples": "..."}

## Exploration Strategy
1. Start by understanding the overall schema structure (SHOW TABLES, DESCRIBE each table)
2. For each table:
   a. Get row count: SELECT COUNT(*) FROM table
   b. Sample some rows: SELECT * FROM table LIMIT 5
   c. For text columns, check value distributions: SELECT col, COUNT(*) FROM table GROUP BY col ORDER BY COUNT(*) DESC LIMIT 15
   d. Check for data quality issues: NULLs, whitespace, type mismatches
   e. For foreign keys, understand the relationship semantics
3. After exploring, save all discovered context using set_rich_context
4. Look for cross-table patterns and business terms

## Important Rules
- ALWAYS explore with execute_sql BEFORE writing context — never guess
- Write context incrementally: save as you discover, don't wait until the end
- Be thorough but efficient: don't run redundant queries
- Primary key columns usually don't need detailed descriptions (just "Primary key, auto-increment" is fine)
- Focus on columns that are semantically interesting (enums, status fields, text with patterns, etc.)
`)

	// Add schema overview
	sb.WriteString("\n## Database Schema Overview\n")

	// Tables
	sb.WriteString(fmt.Sprintf("Total tables: %d\n\n", len(cfg.Tables)))
	for _, t := range cfg.Tables {
		hasDesc := t.Description.Valid && t.Description.String != ""
		if !cfg.Force && hasDesc {
			sb.WriteString(fmt.Sprintf("- **%s** (already has description, skip unless you find issues)\n", t.TableName))
		} else {
			sb.WriteString(fmt.Sprintf("- **%s** (needs context)\n", t.TableName))
		}
	}

	// Columns grouped by table
	colByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, c := range cfg.Columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}

	sb.WriteString("\n### Column Details\n")
	for _, t := range cfg.Tables {
		cols := colByTable[t.TableName]
		if len(cols) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n**%s** (%d columns):\n", t.TableName, len(cols)))
		for _, c := range cols {
			flags := ""
			if c.IsPrimaryKey {
				flags += " [PK]"
			}
			if c.IsForeignKey {
				flags += " [FK]"
			}
			if c.IsNullable {
				flags += " [nullable]"
			}
			needsCtx := cfg.Force || !c.Description.Valid || c.Description.String == ""
			if needsCtx {
				sb.WriteString(fmt.Sprintf("  - %s %s%s (needs context)\n", c.ColumnName, c.DataType.String, flags))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s %s%s (has description)\n", c.ColumnName, c.DataType.String, flags))
			}
		}
	}

	// Relations
	if len(cfg.Relations) > 0 {
		sb.WriteString("\n### Foreign Key Relations\n")
		for _, r := range cfg.Relations {
			sb.WriteString(fmt.Sprintf("- %s.%s → %s.%s\n", r.FromTable, r.FromColumn, r.ToTable, r.ToColumn))
		}
	}

	// Iteration guidance
	sb.WriteString(fmt.Sprintf(`

## Iteration Budget
- Minimum iterations: %d (ensure thorough exploration)
- Maximum iterations: %d
- You have enough budget to explore all tables carefully. Use it wisely.

Now begin your exploration. Start with execute_sql to understand the data, then save context as you discover it.
`, cfg.MinIterations, cfg.MaxIterations))

	return sb.String()
}
