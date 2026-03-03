// Package scenarios provides pre-built ReAct engine configurations for different use cases.
package scenarios

import (
	"fmt"
	"strings"

	"lucid/internal/adapter"
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
	businessDB adapter.DBAdapter,
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
Explore each table, understand its business purpose, and save ALL types of Rich Context using the set_rich_context tool.

## Available Tools
- execute_sql: Run SELECT/SHOW/DESCRIBE queries to explore the database
- set_rich_context: Save discovered context. **SUPPORTS BATCH MODE** — pass a JSON array to save multiple items in ONE call.

## What Rich Context to Generate (ALL FIVE types required)
1. **table_description** — 2-3 sentence summary of the table's business purpose
2. **column_description** — Semantic meaning of each non-trivial column
3. **column_sample_values** — CRITICAL: actual distinct values for text/varchar/enum/categorical columns
4. **column_synonyms** — Natural language alternatives a business user might use
5. **business_term** — Domain-specific terms implied by the data

## Efficient Strategy (IMPORTANT — follow this to maximize coverage)

### For EACH table, use exactly 3 iterations:

**Iteration A — Explore:**
` + "```" + `
SELECT * FROM <table> LIMIT 5
` + "```" + `
This gives you column names, types, and sample data in one query.

**Iteration B — Batch save descriptions + synonyms:**
Call set_rich_context with a JSON ARRAY containing ALL of:
- 1 table_description
- column_descriptions for EVERY column in the table (including PK columns)
- column_synonyms for all non-PK columns
Example:
` + "```" + `json
[
  {"type": "table_description", "table": "T", "value": "..."},
  {"type": "column_description", "table": "T", "column": "C1", "value": "..."},
  {"type": "column_description", "table": "T", "column": "C2", "value": "..."},
  {"type": "column_synonyms", "table": "T", "column": "C1", "value": "..."},
  {"type": "column_synonyms", "table": "T", "column": "C2", "value": "..."}
]
` + "```" + `

**Iteration C — Sample values:**
For text/varchar/enum columns, query sample values and batch save:
` + "```" + `sql
SELECT 'col1' AS col, GROUP_CONCAT(DISTINCT col1 ORDER BY col1 SEPARATOR ', ') AS vals FROM (SELECT col1 FROM T LIMIT 100) t
UNION ALL
SELECT 'col2', GROUP_CONCAT(DISTINCT col2 ORDER BY col2 SEPARATOR ', ') FROM (SELECT col2 FROM T LIMIT 100) t
` + "```" + `
Then batch save all sample_values in one set_rich_context call.

### After all tables: save business_terms in one batch call.

## CRITICAL Rules
- **ALWAYS use batch mode** for set_rich_context — pass arrays, not single objects
- **Do NOT waste iterations** on COUNT(*) queries — SELECT * LIMIT 5 already shows if data exists
- **3 iterations per table** is the target. With ` + fmt.Sprintf("%d", len(cfg.Tables)) + ` tables, aim for ~` + fmt.Sprintf("%d", len(cfg.Tables)*3+10) + ` total iterations
- Skip primary key / auto-increment columns for sample_values and synonyms
- Explore with execute_sql BEFORE writing context — never guess values
- Write context incrementally as you go — don't wait until the end
- If a column has only numeric IDs, skip sample_values for it
- **ALL ` + fmt.Sprintf("%d", len(cfg.Columns)) + ` columns MUST have a column_description** — do NOT skip any column
- When batch-saving descriptions for a table, count them to make sure EVERY column in the table is covered (including PK columns — they need descriptions too)

## Sweep Check Phase (MANDATORY before finishing)
After processing all tables, you MUST do a sweep check:
1. Look at the column list above and compare with what you have saved
2. If any column_description is missing, batch-save the missing ones immediately
3. Only output Final Answer AFTER confirming all ` + fmt.Sprintf("%d", len(cfg.Columns)) + ` columns have descriptions
`)

	// Add schema overview
	sb.WriteString("\n## Database Schema Overview\n")

	// Tables
	sb.WriteString(fmt.Sprintf("Total tables: %d\n\n", len(cfg.Tables)))
	for _, t := range cfg.Tables {
		hasDesc := t.Description.Valid && t.Description.String != ""
		if !cfg.Force && hasDesc {
			sb.WriteString(fmt.Sprintf("- **%s** ✓ has description\n", t.TableName))
		} else {
			sb.WriteString(fmt.Sprintf("- **%s** ✗ needs description\n", t.TableName))
		}
	}

	// Columns grouped by table
	colByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, c := range cfg.Columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}

	sb.WriteString("\n### Column Details (missing context is marked with ✗)\n")
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

			// Build per-column missing list
			var missing []string
			hasDesc := !cfg.Force && c.Description.Valid && c.Description.String != ""
			hasSamples := !cfg.Force && c.SampleValues.Valid && c.SampleValues.String != ""
			hasSynonyms := !cfg.Force && c.Synonyms.Valid && c.Synonyms.String != ""

			if !hasDesc {
				missing = append(missing, "description")
			}
			// PK columns don't need sample_values or synonyms
			if !c.IsPrimaryKey {
				if !hasSamples {
					missing = append(missing, "sample_values")
				}
				if !hasSynonyms {
					missing = append(missing, "synonyms")
				}
			}

			if len(missing) == 0 {
				sb.WriteString(fmt.Sprintf("  - %s %s%s ✓ complete\n", c.ColumnName, c.DataType.String, flags))
			} else {
				sb.WriteString(fmt.Sprintf("  - %s %s%s ✗ missing: %s\n", c.ColumnName, c.DataType.String, flags, strings.Join(missing, ", ")))
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
	targetIter := len(cfg.Tables)*3 + 10
	sb.WriteString(fmt.Sprintf(`

## Iteration Budget
- Target: ~%d iterations for %d tables (3 per table + sweep check + business terms)
- Maximum iterations: %d
- ALWAYS use batch set_rich_context to stay within budget
- Do NOT stop early — process ALL %d tables
- Do NOT ask "would you like me to continue?" — just continue processing all tables
- After all tables, do the **Sweep Check** (see above), then save business terms, THEN output Final Answer

Now begin. Process each table: explore → batch save descriptions+synonyms → batch save sample values. After all tables, sweep check → business terms → Final Answer.
`, targetIter, len(cfg.Tables), cfg.MaxIterations, len(cfg.Tables)))

	return sb.String()
}
