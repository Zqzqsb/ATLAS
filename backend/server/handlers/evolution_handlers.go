package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/agent"
	"lucid/internal/lakebase"
	"lucid/internal/logger"
	"lucid/internal/react"
	"lucid/internal/react/scenarios"
	reacttools "lucid/internal/react/tools"
)

// GetEvolutionStatus returns the current evolution demo state
// GET /api/v1/evolution/status
func (h *Handler) GetEvolutionStatus(c *gin.Context) {
	if h.evolutionService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Evolution service not initialized",
		})
		return
	}

	status := h.evolutionService.GetStatus()
	stages := h.evolutionService.GetStages()

	// Enrich with stage details
	stageDetails := make([]map[string]interface{}, len(stages))
	for i, stage := range stages {
		executed := i < status.CurrentStage
		isCurrent := i == status.CurrentStage
		stageDetails[i] = map[string]interface{}{
			"id":               stage.ID,
			"name":             stage.Name,
			"description":      stage.Description,
			"ddls":             stage.DDLs,
			"expected_changes": stage.ExpectedChanges,
			"executed":         executed,
			"is_next":          isCurrent,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"current_stage": status.CurrentStage,
		"total_stages":  status.TotalStages,
		"database_name": status.DatabaseName,
		"is_ready":      status.IsReady,
		"stages":        stageDetails,
		"history":       status.StageHistory,
	})
}

// GetEvolutionStagePreview returns details about what a stage will do
// GET /api/v1/evolution/stages/:stage_id
func (h *Handler) GetEvolutionStagePreview(c *gin.Context) {
	if h.evolutionService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Evolution service not initialized",
		})
		return
	}

	stageIDStr := c.Param("stage_id")
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid stage ID",
		})
		return
	}

	stage, err := h.evolutionService.GetStagePreview(stageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stage)
}

// ExecuteEvolutionStageStream executes a stage with SSE event streaming
// POST /api/v1/evolution/execute-stage/stream
func (h *Handler) ExecuteEvolutionStageStream(c *gin.Context) {
	if h.evolutionService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Evolution service not initialized",
		})
		return
	}

	var req struct {
		DatasourceID int64 `json:"datasource_id" binding:"required"`
		Stage        int   `json:"stage" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Streaming not supported",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	events := make(chan agent.EvolutionEvent, 50)

	// Execute in background
	var execution *agent.StageExecution
	var execErr error

	go func() {
		defer close(events)
		execution, execErr = h.evolutionService.ExecuteStage(reqCtx, req.DatasourceID, req.Stage, events)
	}()

	// Stream events
	for {
		select {
		case <-reqCtx.Done():
			return
		case event, ok := <-events:
			if !ok {
				// Channel closed — stage execution finished
				if execErr != nil {
					SendSSE(c.Writer, "error", map[string]interface{}{
						"error": execErr.Error(),
					})
					flusher.Flush()
					return
				}

				// Regenerate embeddings after successful stage
				if h.lakebaseService != nil {
					SendSSE(c.Writer, "embedding_update", map[string]interface{}{
						"phase":   "embed",
						"message": "Updating vector embeddings...",
					})
					flusher.Flush()

					embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(reqCtx, req.DatasourceID)
					if embErr != nil {
						logger.L().Warn("Embedding generation after stage failed", "error", embErr)
					} else if embResult != nil {
						SendSSE(c.Writer, "embedding_complete", map[string]interface{}{
							"phase":   "embed",
							"message": fmt.Sprintf("Updated %d embeddings", embResult.TotalEmbeddings),
						})
						flusher.Flush()
					}
				}

				SendSSE(c.Writer, "execution_complete", map[string]interface{}{
					"success":   true,
					"execution": execution,
				})
				flusher.Flush()
				return
			}
			SendSSE(c.Writer, event.Type, map[string]interface{}{
				"phase":   event.Phase,
				"message": event.Message,
				"data":    event.Data,
			})
			flusher.Flush()
		}
	}
}

// ResetEvolutionStream resets with SSE streaming for frontend feedback
// POST /api/v1/evolution/reset/stream
func (h *Handler) ResetEvolutionStream(c *gin.Context) {
	if h.evolutionService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Evolution service not initialized",
		})
		return
	}

	var req struct {
		DatasourceID int64 `json:"datasource_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Streaming not supported",
		})
		return
	}

	// Longer timeout — the ReAct agent may take several minutes
	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	sendStep := func(phase, msg string) {
		SendSSE(c.Writer, "reset_step", map[string]interface{}{
			"phase":   phase,
			"message": msg,
		})
		flusher.Flush()
	}

	// Step 1: Reset database
	sendStep("reset_db", "Dropping and recreating evolution database tables...")

	if err := h.evolutionService.ResetToInitial(reqCtx, req.DatasourceID); err != nil {
		SendSSE(c.Writer, "error", map[string]string{
			"error": "Reset failed: " + err.Error(),
		})
		flusher.Flush()
		return
	}
	sendStep("reset_db_done", "Database tables reset to initial state")

	// Step 2: Sync schema
	sendStep("sync_schema", "Syncing schema to lake-base...")
	if err := h.evolutionService.SyncSchemaToLakebase(reqCtx, req.DatasourceID); err != nil {
		sendStep("sync_schema_warn", "Schema sync warning: "+err.Error())
	} else {
		sendStep("sync_schema_done", "Schema synced to lake-base")
	}

	// Step 3: Run real Onboarding ReAct Agent to generate Rich Context
	sendStep("generate_context", "Running Onboarding ReAct agent to explore database and generate Rich Context...")

	onboardErr := h.runOnboardingForDatasource(reqCtx, req.DatasourceID, "lucid_evolution", func(phase, msg string) {
		sendStep(phase, msg)
	})
	if onboardErr != nil {
		sendStep("generate_context_warn", "Onboarding warning: "+onboardErr.Error())
	} else {
		sendStep("generate_context_done", "Onboarding ReAct agent completed — Rich Context generated from real data")
	}

	// Complete
	SendSSE(c.Writer, "reset_complete", map[string]interface{}{
		"success":       true,
		"current_stage": 0,
		"message":       "Reset to initial state complete",
	})
	flusher.Flush()
}

// runOnboardingForDatasource runs the real Onboarding ReAct pipeline for an
// already-registered datasource. This is the same pipeline used when a user
// clicks "Onboard" on a new database — ReAct agent queries real data, discovers
// distributions, and writes Rich Context + embeddings.
//
// connectionID must match an entry in system.yaml databases[] so we can get a
// DBAdapter that points at the live business database.
func (h *Handler) runOnboardingForDatasource(
	ctx context.Context,
	dsID int64,
	connectionID string,
	progress func(phase, msg string),
) error {
	// 1. Get a live DB adapter for the business database
	adapter, err := h.dbService.GetAdapter(connectionID)
	if err != nil {
		return fmt.Errorf("failed to get adapter for %s: %w", connectionID, err)
	}

	// 2. Ensure lakebase service is available
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		return fmt.Errorf("lakebase service not available")
	}

	// 3. Use lakebase SyncSchema to properly sync via information_schema
	// (replaces the evolution-specific SyncSchemaToLakebase which is less complete)
	syncResult, err := h.lakebaseService.SyncSchema(ctx, dsID, adapter)
	if err != nil {
		return fmt.Errorf("schema sync failed: %w", err)
	}
	progress("schema_synced", fmt.Sprintf("Schema synced: %d tables, %d columns, %d relations",
		syncResult.TablesCount, syncResult.ColumnsCount, syncResult.RelationsCount))

	// 4. Load schema metadata for the ReAct prompt
	tables, err := h.lakebaseService.GetTablesByDatasource(ctx, dsID)
	if err != nil {
		return fmt.Errorf("failed to load tables: %w", err)
	}
	columns, err := h.lakebaseService.GetColumnsByDatasource(ctx, dsID)
	if err != nil {
		return fmt.Errorf("failed to load columns: %w", err)
	}
	relations, _ := h.lakebaseService.GetRelationsByDatasource(ctx, dsID)

	// 5. Get LLM model
	llmModelInterface := h.inferenceService.GetLLMModel()
	if llmModelInterface == nil {
		return fmt.Errorf("LLM not available")
	}
	llmModel, ok := llmModelInterface.(llms.Model)
	if !ok {
		return fmt.Errorf("LLM type assertion failed")
	}

	// 6. Build and run the ReAct onboarding engine
	progress("react_start", "ReAct agent exploring database...")

	repo := h.lakebaseService.GetRepository()
	rcWriter := reacttools.NewLakebaseRCWriter(repo)

	// Compute iterations based on table count (~3 per table + overhead)
	tableCount := len(tables)
	target := tableCount*3 + 10
	maxIter := max(15, int(float64(target)*1.5))
	if maxIter > 300 {
		maxIter = 300
	}
	minIter := max(3, int(float64(target)*0.6))

	engineCfg := scenarios.BuildOnboardingEngine(adapter, rcWriter, scenarios.OnboardingConfig{
		DatasourceID:  dsID,
		DBType:        "mysql",
		Tables:        tables,
		Columns:       columns,
		Relations:     relations,
		MaxIterations: maxIter,
		MinIterations: minIter,
		StepCallback: func(step react.Step, eventType string) {
			msg := ""
			if step.Thought != "" {
				msg = step.Thought
			}
			if step.Action != "" {
				msg = fmt.Sprintf("[%s] %v", step.Action, step.ActionInput)
			}
			if len(msg) > 200 {
				msg = msg[:200] + "..."
			}
			progress("react_"+eventType, msg)
		},
	})

	engine := react.New(llmModel, engineCfg)
	result, err := engine.Execute(ctx, "")
	if err != nil {
		return fmt.Errorf("ReAct agent failed: %w", err)
	}
	progress("react_complete", fmt.Sprintf("Agent finished in %d iterations, %dms",
		result.Iterations, result.Duration.Milliseconds()))

	// 7. Generate embeddings from the newly written Rich Context
	progress("embedding_start", "Generating vector embeddings...")

	embResult, err := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, dsID)
	if err != nil {
		logger.L().Warn("Embedding generation warning", "error", err)
		progress("embedding_warn", "Embedding warning: "+err.Error())
	} else if embResult != nil {
		progress("embedding_done", fmt.Sprintf("Generated %d embeddings", embResult.TotalEmbeddings))
	}

	// 8. Log change
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"tables":     syncResult.TablesCount,
		"columns":    syncResult.ColumnsCount,
		"iterations": result.Iterations,
		"trigger":    "evolution_reset",
	})
	h.lakebaseService.CreateChangeLog(ctx, &lakebase.ChangeLog{
		DatasourceID:  dsID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceSystem,
		ChangeReason:  "Evolution reset — onboarding completed",
	})

	return nil
}
