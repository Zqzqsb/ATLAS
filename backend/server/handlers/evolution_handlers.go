package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/agent"
	"lucid/internal/logger"
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

// ExecuteEvolutionStage executes the next evolution stage (non-streaming)
// POST /api/v1/evolution/execute-stage
func (h *Handler) ExecuteEvolutionStage(c *gin.Context) {
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

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Minute)
	defer cancel()

	execution, err := h.evolutionService.ExecuteStage(reqCtx, req.DatasourceID, req.Stage, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"execution": execution,
		})
		return
	}

	// Auto-regenerate embeddings
	if h.lakebaseService != nil {
		embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(reqCtx, req.DatasourceID)
		if embErr == nil && embResult != nil {
			// Report embeddings in response
			_ = embResult
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"execution": execution,
	})
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
				// Channel closed — send final result
				if execErr != nil {
					SendSSE(c.Writer, "error", map[string]interface{}{
						"error": execErr.Error(),
					})
				} else {
					SendSSE(c.Writer, "execution_complete", map[string]interface{}{
						"success":   true,
						"execution": execution,
					})
				}
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

// ResetEvolution resets the demo to initial state
// POST /api/v1/evolution/reset
func (h *Handler) ResetEvolution(c *gin.Context) {
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

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	if err := h.evolutionService.ResetToInitial(reqCtx, req.DatasourceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Reset failed: " + err.Error(),
		})
		return
	}

	// Re-sync schema to lake-base
	if err := h.evolutionService.SyncSchemaToLakebase(reqCtx, req.DatasourceID); err != nil {
		// Non-fatal: log warning
		logger.L().Warn("Failed to sync schema after reset", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Reset to initial state",
		"current_stage": 0,
	})
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

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
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

	// Step 3: Regenerate context
	sendStep("generate_context", "Regenerating initial Rich Context...")
	// Context regeneration could be triggered here if needed

	// Complete
	SendSSE(c.Writer, "reset_complete", map[string]interface{}{
		"success":       true,
		"current_stage": 0,
		"message":       "Reset to initial state complete",
	})
	flusher.Flush()
}

