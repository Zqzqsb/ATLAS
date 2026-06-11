package scenarios

import (
	"fmt"

	"atlas/internal/lakebase"
	"atlas/internal/react"
	reacttools "atlas/internal/react/tools"

	lctools "github.com/tmc/langchaingo/tools"
)

// MaintainCoordinatorConfig holds parameters for the maintenance coordinator scenario.
type MaintainCoordinatorConfig struct {
	DatasourceID int64
	SignalJSON   string // pre-serialized MaintenanceSignal JSON
	StepCallback react.StepCallback
}

// BuildMaintainCoordinatorEngine creates a ReAct EngineConfig for the maintenance coordinator.
//
// The coordinator agent receives a maintenance signal (DDL changes or data changes),
// inspects the affected entities, marks them as expired, and registers maintenance tasks
// for the executor agent. All decisions are made by the LLM through tools — no hardcoded logic.
func BuildMaintainCoordinatorEngine(
	repo *lakebase.MySQLRepository,
	cfg MaintainCoordinatorConfig,
) (*react.EngineConfig, *reacttools.RegisterTask) {
	inspectTool := reacttools.NewInspectSchemaChange(repo, cfg.DatasourceID)
	markExpiredTool := reacttools.NewMarkExpired(repo, cfg.DatasourceID)
	registerTaskTool := reacttools.NewRegisterTask()
	readContextTool := reacttools.NewReadCurrentContext(repo, cfg.DatasourceID)
	getColumnsTool := reacttools.NewGetTableColumns(repo, cfg.DatasourceID)

	toolsList := []lctools.Tool{inspectTool, markExpiredTool, registerTaskTool, readContextTool, getColumnsTool}

	prompt := buildCoordinatorPrompt(cfg.SignalJSON)

	return &react.EngineConfig{
		MaxIterations: 15,
		MinIterations: 3,
		SystemPrompt:  prompt,
		Tools:         toolsList,
		StepCallback:  cfg.StepCallback,
		LogMode:       "simple",
		Verbose:       true,
	}, registerTaskTool
}

func buildCoordinatorPrompt(signalJSON string) string {
	return fmt.Sprintf(`You are the Maintenance Coordinator for a database Rich Context system.

Given a maintenance signal (DDL changes or data changes), your job is to:
1. Inspect which tables and columns are affected using inspect_schema_change
2. Read their current Rich Context using read_current_context to understand what exists
3. Mark affected entities as expired using mark_expired
4. Register maintenance tasks for the Executor agent using register_task

## Task Types
- "create": Generate new Rich Context for a newly added table/column (no existing context)
- "refresh": Update existing Rich Context because the schema or data changed
- "delete": Clean up Rich Context for a dropped table/column

## Decision Guidelines
- If a table is added: register "create" tasks for the table AND each of its columns
- If a column is added: mark the parent table as expired (table description may need update), register "create" for the new column and "refresh" for the table
- If a column is modified: mark it expired, register "refresh" for the column and the parent table
- If a column is dropped: register "delete" for the column, mark parent table expired and register "refresh"
- If a table is dropped: register "delete" for the table
- If a foreign key is added/dropped: mark the table expired, register "refresh"
- For data changes: mark affected tables/columns expired and register "refresh"

Be thorough but efficient. Process ALL changes in the signal.

## Signal
%s

Begin processing. Use the tools to inspect, mark, and register tasks.`, signalJSON)
}
