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
Explore the database tables, understand their business purpose, discover data patterns, and save ALL types of Rich Context using the set_rich_context tool.

## Available Tools
- execute_sql: Run SELECT/SHOW/DESCRIBE queries to explore the database
- set_rich_context: Save discovered context (table descriptions, column descriptions, sample values, synonyms, business terms)

## What Rich Context to Generate

You MUST generate ALL FIVE types of context. Descriptions alone are NOT sufficient.

1. **Table Description** — A 2-3 sentence summary of the table's business purpose, key columns, and any important data characteristics.
   Use: {"type": "table_description", "table": "...", "value": "..."}

2. **Column Description** — For each non-trivial column, a concise description of its semantic meaning.
   Use: {"type": "column_description", "table": "...", "column": "...", "value": "..."}

3. **Sample Values** (CRITICAL) — For text, varchar, enum, and categorical columns, query the actual distinct values and save them. This is essential for Text-to-SQL to match user queries to real data values.
   Use: {"type": "column_sample_values", "table": "...", "column": "...", "value": "val1, val2, val3"}
   Query: SELECT DISTINCT col FROM table ORDER BY col LIMIT 20

4. **Column Synonyms** (CRITICAL) — Natural language alternatives someone might use when asking questions about this column. Think about how a business user would refer to each column.
   Use: {"type": "column_synonyms", "table": "...", "column": "...", "value": "synonym1, synonym2"}

5. **Business Terms** — Domain-specific terms that appear in the data or are implied by the schema (e.g., "churn rate", "GMV", "high definition").
   Use: {"type": "business_term", "value": "term name", "definition": "...", "category": "...", "synonyms": "...", "examples": "..."}

## Three-Phase Strategy (MUST follow this order)

### Phase 1: Explore & Describe
For each table:
1. SELECT COUNT(*) FROM table
2. SELECT * FROM table LIMIT 5
3. Save table_description and column_description for all columns

### Phase 2: Sample Values (DO NOT SKIP)
After Phase 1, revisit EVERY table and for EACH text/varchar/enum column:
1. Run: SELECT col, COUNT(*) AS cnt FROM table GROUP BY col ORDER BY cnt DESC LIMIT 20
2. Save column_sample_values with the discovered values
3. Also save for columns that look categorical even if numeric (e.g., status codes, ratings, boolean-like)

### Phase 3: Synonyms & Business Terms (DO NOT SKIP)
After Phase 2:
1. For each column, think about what a business user would call it and save column_synonyms
2. Identify domain-specific terms from the data and schema, save as business_term
3. Look for cross-table patterns and domain concepts

## Important Rules
- ALWAYS explore with execute_sql BEFORE writing context — never guess
- Write context incrementally: save as you discover, don't wait until the end
- You MUST complete all three phases — descriptions alone are worthless without sample values and synonyms
- Be thorough but efficient: don't run redundant queries
- Primary key columns usually don't need sample_values or synonyms (just "Primary key, auto-increment" is fine)
- Focus sample_values on columns that have meaningful categorical or textual data
- Every non-PK text column should get synonyms
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
	sb.WriteString(fmt.Sprintf(`

## Iteration Budget
- Minimum iterations: %d (ensure thorough exploration)
- Maximum iterations: %d
- You have enough budget to complete ALL THREE PHASES. Do not stop after descriptions.
- Checkpoint: After finishing Phase 1 (descriptions), you should be roughly 1/3 done.
- If you find yourself running low on iterations, prioritize: sample_values > synonyms > business_terms.

Now begin. Phase 1: explore each table and save descriptions. Then Phase 2: sample values. Then Phase 3: synonyms and business terms. Mark each column's ✗ missing items as you go.
`, cfg.MinIterations, cfg.MaxIterations))

	return sb.String()
}
