// Package agent provides self-maintenance capabilities for LUCID.
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/lakebase"
	"lucid/internal/logger"
	"lucid/internal/react"
	"lucid/internal/react/scenarios"
	reacttools "lucid/internal/react/tools"
)

// AgentService orchestrates the self-maintenance pipeline.
// It receives signals and runs Coordinator + Executor ReAct agents.
type AgentService struct {
	pool       *lakebase.ConnectionPool
	repo       *lakebase.MySQLRepository
	vectorRepo *lakebase.MySQLVectorRepository
	changeLog  *ChangeLogger
	llmModel   llms.Model
}

// NewAgentService creates a new agent service.
func NewAgentService(pool *lakebase.ConnectionPool) *AgentService {
	repo := lakebase.NewMySQLRepository(pool)
	vectorRepo := lakebase.NewMySQLVectorRepository(pool)
	return &AgentService{
		pool:       pool,
		repo:       repo,
		vectorRepo: vectorRepo,
		changeLog:  NewChangeLogger(repo),
	}
}

// SetLLMModel sets the LLM model for the ReAct agents.
func (s *AgentService) SetLLMModel(model llms.Model) {
	s.llmModel = model
}

// GetRepository returns the repository for external use.
func (s *AgentService) GetRepository() *lakebase.MySQLRepository {
	return s.repo
}

// GetVectorRepository returns the vector repository for external use.
func (s *AgentService) GetVectorRepository() *lakebase.MySQLVectorRepository {
	return s.vectorRepo
}

// GetChangeLogger returns the change logger for external use.
func (s *AgentService) GetChangeLogger() *ChangeLogger {
	return s.changeLog
}

// ProcessSignalResult holds the outcome of processing a maintenance signal.
type ProcessSignalResult struct {
	CoordinatorResult *react.Result         `json:"coordinator_result,omitempty"`
	ExecutorResult    *react.Result         `json:"executor_result,omitempty"`
	Tasks             []MaintenanceTask     `json:"tasks"`
	Success           bool                  `json:"success"`
	Error             string                `json:"error,omitempty"`
}

// ProcessSignal runs the Coordinator → Executor pipeline for a maintenance signal.
// stepCallback receives real-time SSE events from both agents.
func (s *AgentService) ProcessSignal(
	ctx context.Context,
	signal *MaintenanceSignal,
	businessDB adapter.DBAdapter,
	stepCallback react.StepCallback,
) (*ProcessSignalResult, error) {
	log := logger.With("component", "agent_service", "dsID", signal.DatasourceID)
	result := &ProcessSignalResult{}

	if s.llmModel == nil {
		return nil, fmt.Errorf("LLM model not available for agent")
	}

	// Phase 1: Run Coordinator Agent
	log.Info("running maintenance coordinator", "signal_type", signal.Type, "changes", len(signal.Changes))

	coordCfg, registerTaskTool := scenarios.BuildMaintainCoordinatorEngine(
		s.repo,
		scenarios.MaintainCoordinatorConfig{
			DatasourceID: signal.DatasourceID,
			SignalJSON:   SignalToJSON(signal),
			StepCallback: wrapCallback(stepCallback, "coordinator"),
		},
	)

	coordEngine := react.New(s.llmModel, coordCfg)
	coordResult, err := coordEngine.Execute(ctx, "")
	if err != nil {
		log.Error("coordinator agent failed", "error", err)
		result.Error = fmt.Sprintf("coordinator failed: %v", err)
		return result, err
	}
	result.CoordinatorResult = coordResult

	// Collect tasks from the register_task tool
	taskCount := registerTaskTool.GetTaskCount()
	tasksJSON := registerTaskTool.GetTasksJSON()

	// Unmarshal tasks for the result (best-effort)
	var tasks []MaintenanceTask
	_ = json.Unmarshal([]byte(tasksJSON), &tasks)
	result.Tasks = tasks

	log.Info("coordinator completed", "iterations", coordResult.Iterations, "tasks_registered", taskCount)

	if taskCount == 0 {
		log.Info("no maintenance tasks registered, skipping executor")
		result.Success = true
		return result, nil
	}

	// Phase 2: Run Executor Agent
	log.Info("running maintenance executor", "tasks", taskCount)

	// Load current schema for the prompt
	tables, _ := s.repo.GetTablesByDatasource(ctx, signal.DatasourceID)
	columns, _ := s.repo.GetColumnsByDatasource(ctx, signal.DatasourceID)

	// Get datasource info for DB name
	ds, _ := s.repo.GetDatasource(ctx, signal.DatasourceID)
	dbName := ""
	if ds != nil && ds.DatabaseName.Valid {
		dbName = ds.DatabaseName.String
	}

	// Build RC writer with OnWrite callback to mark embeddings stale
	rcWriter := reacttools.NewLakebaseRCWriter(s.repo)
	rcWriter.SetOnWrite(func(contextType, tableName, columnName string) {
		// Mark related embeddings as stale
		switch contextType {
		case "table_description":
			for _, t := range tables {
				if t.TableName == tableName {
					_ = s.vectorRepo.MarkEmbeddingStale(ctx, signal.DatasourceID, lakebase.EntityTypeTable, t.ID)
				}
			}
		case "column_description", "column_sample_values", "column_synonyms":
			cols, _ := s.repo.GetColumnsByTable(ctx, signal.DatasourceID, tableName)
			for _, c := range cols {
				if c.ColumnName == columnName {
					_ = s.vectorRepo.MarkEmbeddingStale(ctx, signal.DatasourceID, lakebase.EntityTypeColumn, c.ID)
				}
			}
		}
	})

	execCfg := scenarios.BuildMaintainExecutorEngine(
		businessDB,
		rcWriter,
		s.repo,
		s.vectorRepo,
		scenarios.MaintainExecutorConfig{
			DatasourceID: signal.DatasourceID,
			DBType:       "mysql",
			DBName:       dbName,
			TasksJSON:    tasksJSON,
			TaskCount:    taskCount,
			Tables:       tables,
			Columns:      columns,
			StepCallback: wrapCallback(stepCallback, "executor"),
		},
	)

	execEngine := react.New(s.llmModel, execCfg)
	execResult, err := execEngine.Execute(ctx, "")
	if err != nil {
		log.Error("executor agent failed", "error", err)
		result.Error = fmt.Sprintf("executor failed: %v", err)
		return result, err
	}
	result.ExecutorResult = execResult
	result.Success = true

	log.Info("executor completed", "iterations", execResult.Iterations)

	return result, nil
}

// wrapCallback wraps a StepCallback to add a phase prefix.
func wrapCallback(cb react.StepCallback, phase string) react.StepCallback {
	if cb == nil {
		return nil
	}
	return func(step react.Step, eventType string) {
		cb(step, phase+"_"+eventType)
	}
}

// GetMaintenanceStatus returns the current maintenance status.
func (s *AgentService) GetMaintenanceStatus() map[string]interface{} {
	return map[string]interface{}{
		"llm_available": s.llmModel != nil,
	}
}

// GetChangeLogSummary returns a summary of changes for a datasource.
func (s *AgentService) GetChangeLogSummary(ctx context.Context, dsID int64, limit int) (*ChangeLogSummary, error) {
	return s.changeLog.GetChangeLogSummary(ctx, dsID, limit)
}

// GetRecentChanges returns recent changes for a datasource.
func (s *AgentService) GetRecentChanges(ctx context.Context, dsID int64, limit int) ([]*lakebase.ChangeLog, error) {
	return s.changeLog.GetRecentChanges(ctx, dsID, limit)
}
