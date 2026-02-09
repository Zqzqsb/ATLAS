package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/lakebase"
	"lucid/internal/logger"
	"lucid/internal/react"
	"lucid/internal/react/scenarios"
	reacttools "lucid/internal/react/tools"
)

// ===========================================
// Lake-base Storage API Handlers
// ===========================================

// resolveDatasource resolves a datasource from the :id route param.
// Accepts either a numeric ID or a datasource name.
func (h *Handler) resolveDatasource(c *gin.Context) (*lakebase.Datasource, int64, bool) {
	idStr := c.Param("id")
	ctx := c.Request.Context()

	if dsID, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		ds, err := h.lakebaseService.GetDatasource(ctx, dsID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found"})
			return nil, 0, false
		}
		return ds, dsID, true
	}

	ds, err := h.lakebaseService.GetDatasourceByName(ctx, idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found: " + idStr})
		return nil, 0, false
	}
	return ds, ds.ID, true
}

// GetLakebaseStatus returns the status of lake-base storage
func (h *Handler) GetLakebaseStatus(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":     "Lake-base service not configured",
			"connected": false,
		})
		return
	}

	connected := h.lakebaseService.IsConnected()
	if !connected {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":     "Lake-base service not connected",
			"connected": false,
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.lakebaseService.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"connected": true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":         true,
		"datasources_count": stats.DatasourcesCount,
		"tables_count":      stats.TablesCount,
		"columns_count":     stats.ColumnsCount,
		"context_count":     stats.ContextCount,
		"embeddings_count":  stats.EmbeddingsCount,
		"change_logs_count": stats.ChangeLogsCount,
		"last_updated":      stats.LastUpdated,
	})
}

// ListLakebaseDatasources lists all datasources in lake-base storage
func (h *Handler) ListLakebaseDatasources(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	datasources, err := h.lakebaseService.ListDatasources(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert to safe response with table counts
	result := make([]map[string]interface{}, len(datasources))
	for i, ds := range datasources {
		// Get table count for this datasource
		tables, _ := h.lakebaseService.GetTablesByDatasource(ctx, ds.ID)
		columns, _ := h.lakebaseService.GetColumnsByDatasource(ctx, ds.ID)

		// Count contexts from table and column descriptions (consistent with detail API)
		contextCount := 0
		for _, t := range tables {
			if t.Description.Valid && t.Description.String != "" {
				contextCount++
			}
		}
		for _, c := range columns {
			if c.Description.Valid && c.Description.String != "" {
				contextCount++
			}
		}

		result[i] = map[string]interface{}{
			"id":            ds.ID,
			"name":          ds.Name,
			"db_type":       ds.DBType,
			"host":          ds.Host,
			"port":          ds.Port,
			"database_name": ds.DatabaseName,
			"status":        ds.Status,
			"last_sync_at":  ds.LastSyncAt,
			"created_at":    ds.CreatedAt,
			"updated_at":    ds.UpdatedAt,
			"tables_count":  len(tables),
			"columns_count": len(columns),
			"context_count": contextCount,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"datasources": result,
		"count":       len(result),
	})
}

// GetLakebaseDatasource returns details for a specific datasource
// Supports both numeric ID and name as identifier
func (h *Handler) GetLakebaseDatasource(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	idStr := c.Param("id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var id int64
	var err error

	// Try parsing as numeric ID first, otherwise lookup by name
	id, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// Not a number, try finding by name
		dsObj, lookupErr := h.lakebaseService.GetDatasourceByName(ctx, idStr)
		if lookupErr != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Datasource not found: " + idStr,
			})
			return
		}
		id = dsObj.ID
	}

	// Now fetch datasource by ID (ensures consistent data)
	ds, err := h.lakebaseService.GetDatasource(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get tables from rc_tables
	tableInfos, _ := h.lakebaseService.GetTablesByDatasource(ctx, id)

	// Get columns from rc_columns
	columnInfos, _ := h.lakebaseService.GetColumnsByDatasource(ctx, id)

	// Get relations from rc_relations
	relationInfos, _ := h.lakebaseService.GetRelationsByDatasource(ctx, id)

	// Get embedding count
	embeddingCount, _ := h.lakebaseService.CountEmbeddings(ctx, id)

	// Build column count per table
	columnCountMap := make(map[string]int)
	for _, c := range columnInfos {
		columnCountMap[c.TableName]++
	}

	// Build table summary and collect contexts from descriptions
	tables := make([]map[string]interface{}, 0, len(tableInfos))
	contextList := make([]map[string]interface{}, 0)
	contextID := 1

	for _, t := range tableInfos {
		desc := ""
		if t.Description.Valid {
			desc = t.Description.String
		}
		tables = append(tables, map[string]interface{}{
			"table_name":   t.TableName,
			"description":  desc,
			"row_count":    t.RowCount,
			"column_count": columnCountMap[t.TableName],
			"is_expired":   t.IsExpired,
			"confidence":   t.Confidence,
		})

		// Add table description as context
		if desc != "" {
			contextList = append(contextList, map[string]interface{}{
				"id":           contextID,
				"table_name":   t.TableName,
				"column_name":  nil,
				"context_type": "description",
				"content":      desc,
				"source":       t.Source,
				"confidence":   t.Confidence,
				"created_at":   t.UpdatedAt,
			})
			contextID++
		}
	}

	// Build column details and collect contexts from descriptions, sample_values, synonyms
	columns := make([]map[string]interface{}, 0, len(columnInfos))
	for _, c := range columnInfos {
		desc := ""
		if c.Description.Valid {
			desc = c.Description.String
		}
		dataType := ""
		if c.DataType.Valid {
			dataType = c.DataType.String
		}
		columns = append(columns, map[string]interface{}{
			"table_name":   c.TableName,
			"column_name":  c.ColumnName,
			"data_type":    dataType,
			"description":  desc,
			"is_pk":        c.IsPrimaryKey,
			"is_fk":        c.IsForeignKey,
			"is_nullable":  c.IsNullable,
			"is_expired":   c.IsExpired,
			"confidence":   c.Confidence,
		})

		// Add column description as context
		if desc != "" {
			contextList = append(contextList, map[string]interface{}{
				"id":           contextID,
				"table_name":   c.TableName,
				"column_name":  c.ColumnName,
				"context_type": "description",
				"content":      desc,
				"source":       c.Source,
				"confidence":   c.Confidence,
				"created_at":   c.UpdatedAt,
			})
			contextID++
		}

		// Add sample_values as "example" context
		if c.SampleValues.Valid && c.SampleValues.String != "" {
			contextList = append(contextList, map[string]interface{}{
				"id":           contextID,
				"table_name":   c.TableName,
				"column_name":  c.ColumnName,
				"context_type": "example",
				"content":      c.SampleValues.String,
				"source":       c.Source,
				"confidence":   c.Confidence,
				"created_at":   c.UpdatedAt,
			})
			contextID++
		}

		// Add synonyms as "synonym" context
		if c.Synonyms.Valid && c.Synonyms.String != "" {
			contextList = append(contextList, map[string]interface{}{
				"id":           contextID,
				"table_name":   c.TableName,
				"column_name":  c.ColumnName,
				"context_type": "synonym",
				"content":      c.Synonyms.String,
				"source":       c.Source,
				"confidence":   c.Confidence,
				"created_at":   c.UpdatedAt,
			})
			contextID++
		}
	}

	// Add business terms as "business_rule" context
	termInfos, _ := h.lakebaseService.GetTermsByDatasource(ctx, id)
	for _, t := range termInfos {
		contextList = append(contextList, map[string]interface{}{
			"id":           contextID,
			"table_name":   "",
			"column_name":  nil,
			"context_type": "business_rule",
			"content":      t.Term + ": " + t.Definition,
			"source":       "llm",
			"confidence":   0.80,
			"created_at":   t.CreatedAt,
		})
		contextID++
	}

	c.JSON(http.StatusOK, gin.H{
		"datasource": map[string]interface{}{
			"id":            ds.ID,
			"name":          ds.Name,
			"db_type":       ds.DBType,
			"host":          ds.Host,
			"port":          ds.Port,
			"database_name": ds.DatabaseName,
			"status":        ds.Status,
			"last_sync_at":  ds.LastSyncAt,
			"created_at":    ds.CreatedAt,
			"updated_at":    ds.UpdatedAt,
		},
		"tables":           tables,
		"columns":          columns,
		"contexts":         contextList,
		"relations":        buildRelationsList(relationInfos),
		"tables_count":     len(tableInfos),
		"columns_count":    len(columnInfos),
		"context_count":    len(contextList),
		"relations_count":  len(relationInfos),
		"embeddings_count": embeddingCount,
	})
}

// buildRelationsList converts relations to response format
func buildRelationsList(relations []*lakebase.Relation) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(relations))
	for _, rel := range relations {
		desc := ""
		if rel.Description.Valid {
			desc = rel.Description.String
		}
		result = append(result, map[string]interface{}{
			"id":            rel.ID,
			"from_table":    rel.FromTable,
			"from_column":   rel.FromColumn,
			"to_table":      rel.ToTable,
			"to_column":     rel.ToColumn,
			"relation_type": rel.RelationType,
			"description":   desc,
		})
	}
	return result
}

// GetLakebaseTableContext returns context for a specific table
func (h *Handler) GetLakebaseTableContext(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	_, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	tableName := c.Param("table")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get columns for this table from rc_columns
	colInfos, err := h.lakebaseService.GetColumnsByTable(ctx, dsID, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get context for this table
	contexts, err := h.lakebaseService.GetContextByTable(ctx, dsID, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Format columns
	columns := make([]map[string]interface{}, len(colInfos))
	for i, col := range colInfos {
		columns[i] = map[string]interface{}{
			"name":           col.ColumnName,
			"data_type":      col.DataType.String,
			"is_primary_key": col.IsPrimaryKey,
			"is_foreign_key": col.IsForeignKey,
			"nullable":       col.IsNullable,
			"description":    col.Description.String,
		}
	}

	// Format contexts by type
	contextsByType := make(map[string][]map[string]interface{})
	for _, ctx := range contexts {
		ctxMap := map[string]interface{}{
			"id":          ctx.ID,
			"column_name": ctx.ColumnName.String,
			"content":     ctx.Content,
			"source":      ctx.Source,
			"confidence":  ctx.Confidence,
			"version":     ctx.Version,
			"is_expired":  ctx.IsExpired,
			"created_at":  ctx.CreatedAt,
		}
		contextsByType[string(ctx.ContextType)] = append(contextsByType[string(ctx.ContextType)], ctxMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"datasource_id":    dsID,
		"table_name":       tableName,
		"columns":          columns,
		"column_count":     len(columns),
		"contexts":         contextsByType,
		"total_context":    len(contexts),
	})
}

// GetLakebaseChangeLogs returns change logs for a datasource
func (h *Handler) GetLakebaseChangeLogs(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	_, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	logs, err := h.lakebaseService.GetChangeLogsByDatasource(ctx, dsID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Format logs
	result := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		result[i] = map[string]interface{}{
			"id":             log.ID,
			"table_name":     log.TableName,
			"change_type":    log.ChangeType,
			"change_detail":  log.ChangeDetail,
			"trigger_source": log.TriggerSource,
			"change_reason":  log.ChangeReason,
			"created_at":     log.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"datasource_id": dsID,
		"logs":          result,
		"count":         len(result),
	})
}

// ConnectLakebase manually connects to lake-base storage
func (h *Handler) ConnectLakebase(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not configured",
		})
		return
	}

	if h.lakebaseService.IsConnected() {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Already connected",
			"connected": true,
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.lakebaseService.Connect(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     err.Error(),
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Connected successfully",
		"connected": true,
	})
}

// GenerateEmbeddings generates embeddings for a datasource
// POST /api/v1/lakebase/datasources/:id/embeddings
func (h *Handler) GenerateEmbeddings(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not configured",
		})
		return
	}

	_, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	result, err := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"datasource_id":      result.DatasourceID,
		"tables_processed":   result.TablesProcessed,
		"columns_processed":  result.ColumnsProcessed,
		"contexts_processed": result.ContextsProcessed,
		"total_embeddings":   result.TotalEmbeddings,
	})
}

// GenerateContextEvent represents an SSE event for context generation
type GenerateContextEvent struct {
	Type      string      `json:"type"`      // agent_start, agent_step, agent_done, storage, complete, error
	Agent     string      `json:"agent"`     // coordinator, worker-1, worker-2...
	Table     string      `json:"table"`     // table name (for workers)
	Phase     string      `json:"phase"`     // discovery, schema, rich_context, storage
	Status    string      `json:"status"`    // running, success, error, pending
	Message   string      `json:"message"`   // detailed message
	Data      interface{} `json:"data"`      // structured data
	Timestamp int64       `json:"timestamp"` // unix timestamp ms
}

// GenerateRichContextStream generates Rich Context using a ReAct agent with SSE progress streaming.
// The agent autonomously explores the database via execute_sql and writes context via set_rich_context.
// POST /api/v1/lakebase/datasources/:id/generate-context/stream
func (h *Handler) GenerateRichContextStream(c *gin.Context) {
	// Validate services
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not configured"})
		return
	}

	// Get LLM model
	var model llms.Model
	if h.inferenceService != nil {
		if m := h.inferenceService.GetLLMModel(); m != nil {
			if llmModel, ok := m.(llms.Model); ok {
				model = llmModel
			}
		}
	}
	if model == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "LLM service not configured"})
		return
	}

	// Parse request
	var req struct {
		Force         bool `json:"force"`          // Force regenerate even if exists
		MinIterations int  `json:"min_iterations"` // Min exploration iterations
		MaxIterations int  `json:"max_iterations"` // Max exploration iterations
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
		req.MinIterations = 3
		req.MaxIterations = 20
	}
	if req.MinIterations < 1 {
		req.MinIterations = 3
	}
	if req.MaxIterations < req.MinIterations {
		req.MaxIterations = req.MinIterations * 3
	}

	// Resolve datasource
	ds, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()

	// Get business database adapter
	var businessDB adapter.DBAdapter
	if h.dbService != nil && ds.Name != "" {
		if adp, adpErr := h.dbService.GetAdapter(ds.Name); adpErr == nil {
			businessDB = adp
		}
	}
	if businessDB == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot connect to business database '%s'", ds.Name)})
		return
	}

	// Load schema from lakebase
	tables, err := h.lakebaseService.GetTablesByDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tables: " + err.Error()})
		return
	}
	columns, err := h.lakebaseService.GetColumnsByDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get columns: " + err.Error()})
		return
	}
	relations, err := h.lakebaseService.GetRelationsByDatasource(ctx, dsID)
	if err != nil {
		relations = nil // non-fatal
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Event channel for thread-safe SSE sending
	eventChan := make(chan GenerateContextEvent, 100)
	done := make(chan struct{})

	go func() {
		for event := range eventChan {
			event.Timestamp = time.Now().UnixMilli()
			data, _ := json.Marshal(event)
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event.Type, string(data))
			c.Writer.Flush()
		}
		close(done)
	}()

	sendEvent := func(eventType string, event GenerateContextEvent) {
		event.Type = eventType
		eventChan <- event
	}

	startTime := time.Now()
	log := logger.With("component", "generate_context", "datasource", ds.Name, "dsID", dsID)
	log.Info("starting Rich Context generation",
		"tables", len(tables),
		"columns", len(columns),
		"max_iterations", req.MaxIterations,
		"min_iterations", req.MinIterations,
		"force", req.Force,
	)

	// Phase 1: Announce start with totals
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "rc_gen",
		Phase:   "init",
		Status:  "running",
		Message: fmt.Sprintf("Starting ReAct Rich Context generation for %s (%d tables, %d columns)", ds.Name, len(tables), len(columns)),
		Data: map[string]interface{}{
			"tables_total":  len(tables),
			"columns_total": len(columns),
		},
	})

	// Phase 2: Run ReAct agent
	// Create RC writer with storage callback to notify frontend on each write
	// and trigger incremental embedding immediately
	rcWriter := reacttools.NewLakebaseRCWriter(h.lakebaseService.GetRepository())

	// Track incremental embedding count
	var embeddingCount int
	var embeddingMu sync.Mutex
	embeddingsTotal := len(tables) + len(columns)

	// Announce embedding agent is running (starts alongside RC gen)
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "embedding",
		Phase:   "embedding",
		Status:  "running",
		Message: "Streaming embeddings — each RC write triggers immediate embedding...",
	})

	rcWriter.SetOnWrite(func(contextType, tableName, columnName string) {
		target := "rc_tables"
		detail := tableName
		if columnName != "" {
			target = "rc_columns"
			detail = tableName + "." + columnName
		}
		if contextType == "business_term" {
			target = "rc_terms"
		}
		sendEvent("storage", GenerateContextEvent{
			Agent:   "storage",
			Phase:   contextType,
			Message: fmt.Sprintf("Saved %s: %s", contextType, detail),
			Data: map[string]interface{}{
				"target":       target,
				"context_type": contextType,
				"table":        tableName,
				"column":       columnName,
			},
		})

		// Incremental embedding: embed this entity immediately after write
		go func(ct, tn, cn string) {
			if embErr := h.lakebaseService.EmbedEntityByName(ctx, dsID, ct, tn, cn); embErr != nil {
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   "embedding",
					Phase:   "embedding",
					Status:  "error",
					Message: fmt.Sprintf("⚠️ Embed failed for %s %s: %v", ct, tn, embErr),
				})
				return
			}
			embeddingMu.Lock()
			embeddingCount++
			cnt := embeddingCount
			embeddingMu.Unlock()
			sendEvent("agent_step", GenerateContextEvent{
				Agent:   "embedding",
				Phase:   "embedding",
				Message: fmt.Sprintf("🧬 Embedded %s: %s (%d/%d)", ct, detail, cnt, embeddingsTotal),
				Data: map[string]interface{}{
					"context_type":        ct,
					"table":               tn,
					"column":              cn,
					"embeddings_so_far":   cnt,
					"embeddings_total":    embeddingsTotal,
				},
			})
		}(contextType, tableName, columnName)
	})

	engineCfg := scenarios.BuildRCGenEngine(businessDB, rcWriter, scenarios.RCGenConfig{
		DatasourceID:  dsID,
		Tables:        tables,
		Columns:       columns,
		Relations:     relations,
		MaxIterations: req.MaxIterations,
		MinIterations: req.MinIterations,
		Force:         req.Force,
		StepCallback: func(step react.Step, eventType string) {
			// Send distinct SSE events for each ReAct step type
			switch eventType {
			case "thought":
				if step.Thought != "" {
					sendEvent("agent_step", GenerateContextEvent{
						Agent:   "rc_gen",
						Phase:   "thought",
						Message: step.Thought,
						Data: map[string]interface{}{
							"iteration": step.Iteration,
						},
					})
				}
			case "action":
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   "rc_gen",
					Phase:   "action",
					Message: fmt.Sprintf("🔧 %s", step.Action),
					Data: map[string]interface{}{
						"iteration":    step.Iteration,
						"action":       step.Action,
						"action_input": step.ActionInput,
					},
				})
			case "observation":
				obs := step.Observation
				if len(obs) > 500 {
					obs = obs[:500] + "..."
				}
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   "rc_gen",
					Phase:   "observation",
					Message: obs,
					Data: map[string]interface{}{
						"iteration":    step.Iteration,
						"action":       step.Action,
						"observation":  step.Observation,
					},
				})
			case "finish":
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   "rc_gen",
					Phase:   "finish",
					Message: step.Thought,
					Data: map[string]interface{}{
						"iteration": step.Iteration,
					},
				})
			}
		},
	})

	engine := react.New(model, engineCfg)
	result, execErr := engine.Execute(ctx, "")

	if execErr != nil {
		log.Error("ReAct agent execution failed", "error", execErr)
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "rc_gen",
			Status:  "error",
			Message: fmt.Sprintf("ReAct agent error: %v", execErr),
		})
	} else {
		log.Info("ReAct agent completed",
			"iterations", result.Iterations,
			"duration_s", result.Duration.Seconds(),
			"output_length", len(result.Output),
		)
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "rc_gen",
			Status:  "success",
			Message: fmt.Sprintf("ReAct agent completed in %d iterations (%.1fs)", result.Iterations, result.Duration.Seconds()),
			Data: map[string]interface{}{
				"iterations": result.Iterations,
				"duration":   result.Duration.Seconds(),
				"output":     result.Output,
			},
		})
	}

	// Phase 3: Catch-up embeddings for any entities that may have been missed
	// (e.g., entities that existed before but had no embedding yet)
	embeddingMu.Lock()
	streamEmbedded := embeddingCount
	embeddingMu.Unlock()

	var catchupEmbeddings int
	embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, dsID)
	if embErr != nil {
		log.Error("catch-up embedding failed", "stream_embedded", streamEmbedded, "error", embErr)
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "embedding",
			Phase:   "embedding",
			Status:  "error",
			Message: fmt.Sprintf("Catch-up embedding error: %v", embErr),
		})
	} else if embResult != nil {
		catchupEmbeddings = embResult.TotalEmbeddings
		log.Info("embeddings complete",
			"stream_embedded", streamEmbedded,
			"catchup_embedded", catchupEmbeddings,
			"total_embeddings", catchupEmbeddings,
		)
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "embedding",
			Phase:   "embedding",
			Status:  "success",
			Message: fmt.Sprintf("Embeddings complete: %d streamed + %d catch-up = %d total", streamEmbedded, catchupEmbeddings, catchupEmbeddings),
			Data: map[string]interface{}{
				"stream_embedded":  streamEmbedded,
				"catchup_embedded": catchupEmbeddings,
				"total_embeddings": catchupEmbeddings,
			},
		})
	}

	// Complete
	duration := time.Since(startTime)
	iterations := 0
	if result != nil {
		iterations = result.Iterations
	}
	totalEmb := catchupEmbeddings
	if totalEmb == 0 {
		totalEmb = streamEmbedded
	}
	sendEvent("complete", GenerateContextEvent{
		Status:  "success",
		Message: "Generation complete",
		Data: map[string]interface{}{
			"total_tables":         len(tables),
			"total_columns":        len(columns),
			"react_iterations":     iterations,
			"embeddings_generated": totalEmb,
			"stream_embedded":      streamEmbedded,
			"duration_ms":          duration.Milliseconds(),
		},
	})

	close(eventChan)
	<-done
}

// ===========================================
// Prune Context API
// ===========================================

// PruneContext deletes all rich context data for a datasource
// DELETE /api/lakebase/datasources/:id/prune
func (h *Handler) PruneContext(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not configured",
		})
		return
	}

	ds, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Prune all context
	if err := h.lakebaseService.PruneAllContext(ctx, dsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to prune context: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    fmt.Sprintf("All rich context pruned for datasource '%s'", ds.Name),
		"datasource": ds.Name,
	})
}

// ===========================================
// Schema Sync API
// ===========================================

// SyncSchema discovers schema from the target business database and upserts into rc_tables/rc_columns/rc_relations
// POST /api/v1/lakebase/datasources/:id/sync-schema
func (h *Handler) SyncSchema(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	ds, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Get adapter for the target database
	adapter, err := h.dbService.GetAdapter(ds.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Cannot connect to database '%s': %v. Make sure the connection is configured.", ds.Name, err),
		})
		return
	}

	result, err := h.lakebaseService.SyncSchema(ctx, dsID, adapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Schema sync failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"datasource": ds.Name,
		"tables":     result.TablesCount,
		"columns":    result.ColumnsCount,
		"relations":  result.RelationsCount,
	})
}

// ===========================================
// Delete Datasource API
// ===========================================

// DeleteDatasource removes a datasource and all its associated RC data
// DELETE /api/v1/lakebase/datasources/:id
func (h *Handler) DeleteDatasource(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	ds, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Delete the datasource record (CASCADE will remove all associated rc_* data)
	if err := h.lakebaseService.DeleteDatasource(ctx, dsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete datasource: %v", err),
		})
		return
	}

	// Also remove from in-memory connection config so it can be re-added
	h.dbService.RemoveDatabase(ds.Name)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    fmt.Sprintf("Datasource '%s' and all associated data deleted", ds.Name),
		"datasource": ds.Name,
	})
}
