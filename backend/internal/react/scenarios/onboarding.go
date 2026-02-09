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

// OnboardingConfig holds parameters for the onboarding scenario.
type OnboardingConfig struct {
	DatasourceID  int64
	DBType        string
	Tables        []*lakebase.TableInfo
	Columns       []*lakebase.ColumnInfo
	Relations     []*lakebase.Relation
	MaxIterations int
	MinIterations int
	StepCallback  react.StepCallback
}

// BuildOnboardingEngine creates a ReAct EngineConfig for database onboarding.
//
// The onboarding agent explores the database schema, discovers data patterns,
// and writes Rich Context (table descriptions, column descriptions, sample values, etc.)
// using the same tool set as RC generation but with a prompt tuned for first-time exploration.
func BuildOnboardingEngine(
	businessDB adapter.DBAdapter,
	rcWriter reacttools.RCWriter,
	cfg OnboardingConfig,
) *react.EngineConfig {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 20
	}
	if cfg.MinIterations <= 0 {
		cfg.MinIterations = 5
	}

	sqlTool := reacttools.NewExecuteSQL(businessDB)
	rcTool := reacttools.NewSetRichContext(rcWriter, cfg.DatasourceID)

	toolsList := []lctools.Tool{sqlTool, rcTool}

	prompt := buildOnboardingPrompt(cfg)

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

func buildOnboardingPrompt(cfg OnboardingConfig) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are onboarding a new %s database into the LUCID Text-to-SQL system.
Your mission is to thoroughly explore the database and generate Rich Context metadata that helps future SQL generation.

## Available Tools
- execute_sql: Run SELECT/SHOW/DESCRIBE queries to explore the database
- set_rich_context: Save discovered context (table descriptions, column descriptions, sample values, synonyms, business terms)

## Onboarding Workflow
1. **Understand the schema**: For each table, examine its columns, types, and constraints
2. **Sample the data**: Run SELECT * FROM table LIMIT 5 to understand actual data
3. **Discover distributions**: For text/enum columns, run GROUP BY to find value distributions
4. **Check data quality**: Look for NULLs, whitespace issues, unexpected patterns
5. **Identify relationships**: Understand foreign keys and cross-table patterns
6. **Write context**: Use set_rich_context to save ALL discoveries

## Rich Context Types
Use set_rich_context with JSON input:

- Table description: {"type":"table_description","table":"...","value":"2-3 sentence description of business purpose"}
- Column description: {"type":"column_description","table":"...","column":"...","value":"semantic meaning"}
- Sample values: {"type":"column_sample_values","table":"...","column":"...","value":"val1, val2, val3"}
- Synonyms: {"type":"column_synonyms","table":"...","column":"...","value":"synonym1, synonym2"}
- Business term: {"type":"business_term","value":"term","definition":"...","category":"..."}

## Important Rules
- ALWAYS explore with execute_sql BEFORE writing context — never guess
- Write context incrementally as you discover it
- Be thorough: cover ALL tables, not just a few
- For enum-like columns (< 20 distinct values), always record sample values
- For text columns, check for leading/trailing whitespace issues
- Primary keys only need brief descriptions ("Primary key, auto-increment")
- Focus effort on semantically interesting columns (status fields, categories, codes)
`, cfg.DBType))

	// Add schema overview
	sb.WriteString("\n## Database Schema\n")
	sb.WriteString(fmt.Sprintf("Total tables: %d\n\n", len(cfg.Tables)))

	colByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, c := range cfg.Columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}

	for _, t := range cfg.Tables {
		sb.WriteString(fmt.Sprintf("### %s", t.TableName))
		if t.RowCount > 0 {
			sb.WriteString(fmt.Sprintf(" (%d rows)", t.RowCount))
		}
		sb.WriteString("\n")

		cols := colByTable[t.TableName]
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
			sb.WriteString(fmt.Sprintf("  - %s %s%s\n", c.ColumnName, c.DataType.String, flags))
		}
		sb.WriteString("\n")
	}

	// Relations
	if len(cfg.Relations) > 0 {
		sb.WriteString("### Foreign Key Relations\n")
		for _, r := range cfg.Relations {
			sb.WriteString(fmt.Sprintf("- %s.%s → %s.%s\n", r.FromTable, r.FromColumn, r.ToTable, r.ToColumn))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`## Iteration Budget
- Minimum: %d iterations (ensure thorough exploration)
- Maximum: %d iterations
- Budget is generous — explore ALL tables carefully.

Begin onboarding now. Start with execute_sql to understand the data.
`, cfg.MinIterations, cfg.MaxIterations))

	return sb.String()
}
