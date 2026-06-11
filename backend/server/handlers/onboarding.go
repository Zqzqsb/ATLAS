package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"atlas/internal/lakebase"
	"atlas/internal/react"
	"atlas/internal/react/scenarios"
	reacttools "atlas/internal/react/tools"
)

// OnboardingEvent represents an event in the onboarding SSE stream.
type OnboardingEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// OnboardingStream handles SSE streaming for database onboarding.
// GET /api/v1/onboarding/stream?connection_id=xxx
func (h *Handler) OnboardingStream(c *gin.Context) {
	connectionID := c.Query("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "connection_id parameter is required"})
		return
	}

	// Find connection config
	dbCfg := h.dbService.FindDatabase(connectionID)
	if dbCfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}
	dbName, dbType := dbCfg.Name, dbCfg.Type

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	events := make(chan OnboardingEvent, 100)

	go h.runOnboarding(ctx, connectionID, dbName, dbType, events)

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			SendSSE(c.Writer, event.Type, event.Data)
			flusher.Flush()
		}
	}
}

// runOnboarding executes the onboarding process using the unified ReAct engine.
// For large schemas (>30 tables), it uses forest-based chunked onboarding:
// FK graph → connected components → per-subtree ReAct agents → merge.
func (h *Handler) runOnboarding(ctx context.Context, connectionID, dbName, dbType string, events chan<- OnboardingEvent) {
	defer close(events)
	startTime := time.Now()

	// Phase 1: Connect and discover schema
	sendEvent(events, "phase_change", map[string]string{"phase": "connecting", "message": "Connecting to database..."})

	adapter, err := h.dbService.GetAdapter(connectionID)
	if err != nil {
		sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("Failed to connect: %v", err)})
		return
	}

	// Phase 2: Ensure lakebase datasource exists and schema is synced
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		sendEvent(events, "error", map[string]string{"message": "Lake-base service not available"})
		return
	}

	sendEvent(events, "phase_change", map[string]string{"phase": "discovering", "message": "Syncing schema..."})

	ds, err := h.lakebaseService.GetOrCreateDatasource(ctx, &lakebase.Datasource{
		Name:         connectionID,
		DBType:       dbType,
		DatabaseName: sql.NullString{String: dbName, Valid: true},
		Status:       lakebase.DatasourceStatusActive,
	})
	if err != nil {
		sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("Failed to register datasource: %v", err)})
		return
	}

	// Sync physical schema into rc_tables/rc_columns
	syncResult, err := h.lakebaseService.SyncSchema(ctx, ds.ID, adapter)
	if err != nil {
		sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("Schema sync failed: %v", err)})
		return
	}

	sendEvent(events, "schema_synced", map[string]interface{}{
		"datasource_id": ds.ID,
		"tables":        syncResult.TablesCount,
		"columns":       syncResult.ColumnsCount,
		"relations":     syncResult.RelationsCount,
	})

	// Phase 3: Load schema metadata for the ReAct prompt
	tables, err := h.lakebaseService.GetTablesByDatasource(ctx, ds.ID)
	if err != nil {
		sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("Failed to load tables: %v", err)})
		return
	}
	columns, err := h.lakebaseService.GetColumnsByDatasource(ctx, ds.ID)
	if err != nil {
		sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("Failed to load columns: %v", err)})
		return
	}
	relations, _ := h.lakebaseService.GetRelationsByDatasource(ctx, ds.ID)

	tableNames := make([]string, len(tables))
	for i, t := range tables {
		tableNames[i] = t.TableName
	}
	sendEvent(events, "table_discovered", map[string]interface{}{
		"tables": tableNames,
		"count":  len(tables),
	})

	// Phase 4: Get LLM model
	llmModelInterface := h.inferenceService.GetLLMModel()
	if llmModelInterface == nil {
		sendEvent(events, "error", map[string]string{"message": "LLM not available"})
		return
	}
	llmModel, ok := llmModelInterface.(llms.Model)
	if !ok {
		sendEvent(events, "error", map[string]string{"message": "LLM type assertion failed"})
		return
	}

	// Phase 5: Run ReAct onboarding agent(s)
	repo := h.lakebaseService.GetRepository()
	rcWriter := reacttools.NewLakebaseRCWriter(repo)
	tableCount := len(tables)

	const forestThreshold = 30

	if tableCount <= forestThreshold {
		// Small schema: single-agent onboarding (original path)
		sendEvent(events, "phase_change", map[string]string{"phase": "analyzing", "message": "Agent exploring database with ReAct..."})

		minIter, maxIter := scenarios.ComputeChunkBudget(tableCount)
		engineCfg := scenarios.BuildOnboardingEngine(adapter, rcWriter, scenarios.OnboardingConfig{
			DatasourceID:  ds.ID,
			DBType:        dbType,
			Tables:        tables,
			Columns:       columns,
			Relations:     relations,
			MaxIterations: maxIter,
			MinIterations: minIter,
			StepCallback: func(step react.Step, eventType string) {
				events <- OnboardingEvent{
					Type: "react_" + eventType,
					Data: map[string]interface{}{
						"iteration":    step.Iteration,
						"thought":      step.Thought,
						"action":       step.Action,
						"action_input": step.ActionInput,
						"observation":  step.Observation,
					},
					Timestamp: time.Now().UnixMilli(),
				}
			},
		})

		engine := react.New(llmModel, engineCfg)
		result, err := engine.Execute(ctx, "")
		if err != nil {
			sendEvent(events, "error", map[string]string{"message": fmt.Sprintf("ReAct agent failed: %v", err)})
			return
		}

		sendEvent(events, "react_complete", map[string]interface{}{
			"iterations": result.Iterations,
			"duration":   result.Duration.Milliseconds(),
			"output":     result.Output,
		})
	} else {
		// Large schema: forest-based chunked onboarding
		sendEvent(events, "phase_change", map[string]string{
			"phase":   "forest_decompose",
			"message": fmt.Sprintf("Decomposing %d tables into FK-based clusters...", tableCount),
		})

		forestResult := scenarios.ForestDecompose(tables, columns, relations)

		// Merge isolated (single-table, no FK) clusters into batches of 15
		clusters := scenarios.MergeIsolatedTables(forestResult.Clusters, 15)

		sendEvent(events, "forest_result", map[string]interface{}{
			"total_tables":   forestResult.TotalTables,
			"clusters":       len(clusters),
			"largest_cluster": forestResult.LargestSize,
			"median_cluster":  forestResult.MedianSize,
			"isolated_tables": forestResult.IsolatedCount,
		})

		totalIterations := 0
		var lastErr error

		for ci, cluster := range clusters {
			clusterTableNames := make([]string, len(cluster.Tables))
			for i, t := range cluster.Tables {
				clusterTableNames[i] = t.TableName
			}

			sendEvent(events, "chunk_start", map[string]interface{}{
				"chunk_index":  ci,
				"chunk_total":  len(clusters),
				"tables":       clusterTableNames,
				"table_count":  len(cluster.Tables),
				"relation_count": len(cluster.Relations),
			})

			minIter, maxIter := scenarios.ComputeChunkBudget(len(cluster.Tables))

			engineCfg := scenarios.BuildOnboardingEngine(adapter, rcWriter, scenarios.OnboardingConfig{
				DatasourceID:  ds.ID,
				DBType:        dbType,
				Tables:        cluster.Tables,
				Columns:       cluster.Columns,
				Relations:     cluster.Relations,
				MaxIterations: maxIter,
				MinIterations: minIter,
				StepCallback: func(step react.Step, eventType string) {
					events <- OnboardingEvent{
						Type: "react_" + eventType,
						Data: map[string]interface{}{
							"chunk":        ci,
							"iteration":    step.Iteration,
							"thought":      step.Thought,
							"action":       step.Action,
							"action_input": step.ActionInput,
							"observation":  step.Observation,
						},
						Timestamp: time.Now().UnixMilli(),
					}
				},
			})

			engine := react.New(llmModel, engineCfg)
			result, err := engine.Execute(ctx, "")
			if err != nil {
				sendEvent(events, "chunk_error", map[string]interface{}{
					"chunk_index": ci,
					"error":       err.Error(),
				})
				lastErr = err
				continue // Continue with next chunk even if one fails
			}

			totalIterations += result.Iterations

			sendEvent(events, "chunk_complete", map[string]interface{}{
				"chunk_index": ci,
				"iterations":  result.Iterations,
				"duration":    result.Duration.Milliseconds(),
			})
		}

		if lastErr != nil {
			sendEvent(events, "warning", map[string]string{
				"message": fmt.Sprintf("Some chunks had errors (last: %v), but onboarding continued", lastErr),
			})
		}

		sendEvent(events, "react_complete", map[string]interface{}{
			"iterations":   totalIterations,
			"chunks":       len(clusters),
			"duration":     time.Since(startTime).Milliseconds(),
			"output":       fmt.Sprintf("Forest-based onboarding completed: %d clusters, %d total iterations", len(clusters), totalIterations),
		})
	}

	// Phase 6: Generate embeddings
	sendEvent(events, "phase_change", map[string]string{"phase": "embedding", "message": "Generating embeddings..."})

	embResult, err := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, ds.ID)
	if err != nil {
		sendEvent(events, "warning", map[string]string{"message": fmt.Sprintf("Embedding generation partial: %v", err)})
	} else if embResult != nil {
		sendEvent(events, "embeddings_complete", map[string]interface{}{
			"total": embResult.TotalEmbeddings,
		})
	}

	// Phase 7: Create change log
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"tables":  syncResult.TablesCount,
		"columns": syncResult.ColumnsCount,
		"mode":    func() string { if tableCount > forestThreshold { return "forest_chunked" }; return "single_agent" }(),
	})
	h.lakebaseService.CreateChangeLog(ctx, &lakebase.ChangeLog{
		DatasourceID:  ds.ID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceSystem,
		ChangeReason:  "Onboarding completed",
	})

	// Complete
	sendEvent(events, "complete", map[string]interface{}{
		"message":       "Onboarding completed successfully",
		"total_time_ms": time.Since(startTime).Milliseconds(),
		"datasource_id": ds.ID,
		"tables":        syncResult.TablesCount,
		"columns":       syncResult.ColumnsCount,
	})
}

// --- SSE helpers ---

func sendEvent(events chan<- OnboardingEvent, eventType string, data interface{}) {
	events <- OnboardingEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

