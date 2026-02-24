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

// MaintainExecutorConfig holds parameters for the maintenance executor scenario.
type MaintainExecutorConfig struct {
	DatasourceID int64
	DBType       string
	DBName       string
	TasksJSON    string // pre-serialized []MaintenanceTask JSON
	TaskCount    int    // number of tasks (for iteration calculation)
	Tables       []*lakebase.TableInfo
	Columns      []*lakebase.ColumnInfo
	StepCallback react.StepCallback
}

// BuildMaintainExecutorEngine creates a ReAct EngineConfig for the maintenance executor.
func BuildMaintainExecutorEngine(
	businessDB adapter.DBAdapter,
	rcWriter reacttools.RCWriter,
	repo *lakebase.MySQLRepository,
	vectorRepo *lakebase.MySQLVectorRepository,
	cfg MaintainExecutorConfig,
) *react.EngineConfig {
	sqlTool := reacttools.NewExecuteSQL(businessDB)
	rcTool := reacttools.NewSetRichContext(rcWriter, cfg.DatasourceID)
	deleteTool := reacttools.NewDeleteRichContext(repo, vectorRepo, cfg.DatasourceID)
	clearTool := reacttools.NewClearExpired(repo, cfg.DatasourceID)

	toolsList := []lctools.Tool{sqlTool, rcTool, deleteTool, clearTool}

	prompt := buildExecutorPrompt(cfg)

	maxIter := cfg.TaskCount * 5
	if maxIter < 10 {
		maxIter = 10
	}
	if maxIter > 30 {
		maxIter = 30
	}

	minIter := cfg.TaskCount
	if minIter < 1 {
		minIter = 1
	}

	return &react.EngineConfig{
		MaxIterations: maxIter,
		MinIterations: minIter,
		SystemPrompt:  prompt,
		Tools:         toolsList,
		StepCallback:  cfg.StepCallback,
		LogMode:       "simple",
		Verbose:       true,
	}
}

func buildExecutorPrompt(cfg MaintainExecutorConfig) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are the Maintenance Executor for a database Rich Context system.
Database: %s "%s"

You have a list of maintenance tasks to complete. For each task:

## Task Processing Rules

### "create" tasks — entity is new (just added via DDL)
1. Use execute_sql to sample data: SELECT * FROM table LIMIT 5
2. For text/enum columns, check distributions: SELECT col, COUNT(*) FROM table GROUP BY col ORDER BY cnt DESC LIMIT 20
3. Generate appropriate description, sample_values, synonyms
4. Use set_rich_context to save them (table_description, column_description, column_sample_values, column_synonyms)
5. Use clear_expired to mark the entity as no longer stale

### "refresh" tasks — entity's schema or data changed
1. The current context (if any) is provided in the task context field
2. Use execute_sql to check if current data is still valid (sample new values, check distributions)
3. Update description/sample_values/synonyms via set_rich_context
4. Use clear_expired to mark the entity as no longer stale

### "delete" tasks — entity was dropped
1. Use delete_rich_context to clean up the dropped table/column
2. No need to call clear_expired for deleted entities

## Rich Context Types (for set_rich_context)
- Table description: {"type":"table_description","table":"...","value":"2-3 sentence description"}
- Column description: {"type":"column_description","table":"...","column":"...","value":"semantic meaning"}
- Sample values: {"type":"column_sample_values","table":"...","column":"...","value":"val1, val2, val3"}
- Synonyms: {"type":"column_synonyms","table":"...","column":"...","value":"synonym1, synonym2"}
- Business term: {"type":"business_term","value":"term","definition":"...","category":"..."}

## Important
- Process ALL tasks. Do not skip any.
- ALWAYS explore with execute_sql BEFORE writing context for create/refresh tasks.
- For primary key columns, brief descriptions are fine ("Primary key, auto-increment").
- Focus effort on semantically interesting columns.

`, cfg.DBType, cfg.DBName))

	// Add schema summary
	sb.WriteString("## Current Schema\n")
	colByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, c := range cfg.Columns {
		colByTable[c.TableName] = append(colByTable[c.TableName], c)
	}
	for _, t := range cfg.Tables {
		sb.WriteString(fmt.Sprintf("### %s", t.TableName))
		if t.Description.Valid && t.Description.String != "" {
			sb.WriteString(fmt.Sprintf(" — %s", t.Description.String))
		}
		sb.WriteString("\n")
		for _, c := range colByTable[t.TableName] {
			flags := ""
			if c.IsPrimaryKey {
				flags += " [PK]"
			}
			if c.IsForeignKey {
				flags += " [FK]"
			}
			sb.WriteString(fmt.Sprintf("  - %s %s%s\n", c.ColumnName, c.DataType.String, flags))
		}
	}

	sb.WriteString(fmt.Sprintf("\n## Tasks\n%s\n\nBegin processing tasks now.\n", cfg.TasksJSON))

	return sb.String()
}
