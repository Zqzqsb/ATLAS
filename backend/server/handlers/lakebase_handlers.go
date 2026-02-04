package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/lakebase"
)

// ===========================================
// Lake-base Storage API Handlers
// ===========================================

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
		contexts, _ := h.lakebaseService.GetContextByDatasource(ctx, ds.ID)

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
			"context_count": len(contexts),
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

	// Get context for this datasource
	contexts, _ := h.lakebaseService.GetContextByDatasource(ctx, id)

	// Get embedding count
	embeddingCount, _ := h.lakebaseService.CountEmbeddings(ctx, id)

	// Build column count per table
	columnCountMap := make(map[string]int)
	for _, c := range columnInfos {
		columnCountMap[c.TableName]++
	}

	// Build table summary
	tables := make([]map[string]interface{}, 0, len(tableInfos))
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
	}

	// Build column details
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
	}

	// Build context details
	contextList := make([]map[string]interface{}, 0, len(contexts))
	for _, ctx := range contexts {
		contextList = append(contextList, map[string]interface{}{
			"id":           ctx.ID,
			"table_name":   ctx.TableName,
			"column_name":  ctx.ColumnName,
			"context_type": ctx.ContextType,
			"content":      ctx.Content,
			"created_at":   ctx.CreatedAt,
		})
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
		"tables_count":     len(tableInfos),
		"columns_count":    len(columnInfos),
		"context_count":    len(contexts),
		"embeddings_count": embeddingCount,
	})
}

// GetLakebaseTableContext returns context for a specific table
func (h *Handler) GetLakebaseTableContext(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	dsIDStr := c.Param("id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	tableName := c.Param("table")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get schema for this table
	schemas, err := h.lakebaseService.GetTableSchema(ctx, dsID, tableName)
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
	columns := make([]map[string]interface{}, len(schemas))
	for i, s := range schemas {
		columns[i] = map[string]interface{}{
			"name":           s.ColumnName,
			"data_type":      s.DataType,
			"is_primary_key": s.IsPrimaryKey,
			"is_foreign_key": s.IsForeignKey,
			"nullable":       s.Nullable,
			"comment":        s.Comment,
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

	dsIDStr := c.Param("id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
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

	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
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

// GenerateRichContext generates Rich Context descriptions for tables and columns using LLM
// POST /api/v1/lakebase/datasources/:id/generate-context
func (h *Handler) GenerateRichContext(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not configured",
		})
		return
	}

	// Get LLM model from inference service
	var model llms.Model
	if h.inferenceService != nil {
		if m := h.inferenceService.GetLLMModel(); m != nil {
			if llmModel, ok := m.(llms.Model); ok {
				model = llmModel
			}
		}
	}

	if model == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "LLM service not configured",
		})
		return
	}

	idStr := c.Param("id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	// Support both numeric ID and name
	var dsID int64
	var err error
	dsID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		// Try finding by name
		ds, lookupErr := h.lakebaseService.GetDatasourceByName(ctx, idStr)
		if lookupErr != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Datasource not found: " + idStr,
			})
			return
		}
		dsID = ds.ID
	}

	// Get tables and columns
	tables, err := h.lakebaseService.GetTablesByDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tables: " + err.Error(),
		})
		return
	}

	columns, err := h.lakebaseService.GetColumnsByDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get columns: " + err.Error(),
		})
		return
	}

	// Group columns by table
	columnsByTable := make(map[string][]string)
	for _, col := range columns {
		columnsByTable[col.TableName] = append(columnsByTable[col.TableName], col.ColumnName)
	}

	// Generate descriptions using LLM
	tablesUpdated := 0
	columnsUpdated := 0

	for _, table := range tables {
		// Skip if already has description
		if table.Description.Valid && table.Description.String != "" {
			continue
		}

		// Build schema info for prompt
		cols := columnsByTable[table.TableName]
		colList := ""
		for _, col := range cols {
			colList += col + ", "
		}
		if len(colList) > 2 {
			colList = colList[:len(colList)-2]
		}

		// Generate table description
		prompt := fmt.Sprintf(`Analyze this database table and provide a concise description.

Table: %s
Columns: %s
Row Count: %d

Generate a one-sentence description of what this table stores and its business purpose.
Output only the description text, no JSON or formatting.`, table.TableName, colList, table.RowCount)

		response, err := llms.GenerateFromSinglePrompt(ctx, model, prompt)
		if err != nil {
			continue // Skip on error, don't fail entire operation
		}

		description := strings.TrimSpace(response)
		if description != "" {
			err = h.lakebaseService.UpdateTableDescription(ctx, dsID, table.TableName, description, "llm", 0.85)
			if err == nil {
				tablesUpdated++
			}
		}
	}

	// Generate column descriptions
	for _, col := range columns {
		// Skip if already has description
		if col.Description.Valid && col.Description.String != "" {
			continue
		}

		prompt := fmt.Sprintf(`Analyze this database column and provide a concise description.

Table: %s
Column: %s
Data Type: %s
Is Primary Key: %v
Is Foreign Key: %v
Is Nullable: %v

Generate a one-sentence description of what this column represents.
Output only the description text, no JSON or formatting.`, 
			col.TableName, col.ColumnName, 
			col.DataType.String, col.IsPrimaryKey, col.IsForeignKey, col.IsNullable)

		response, err := llms.GenerateFromSinglePrompt(ctx, model, prompt)
		if err != nil {
			continue
		}

		description := strings.TrimSpace(response)
		if description != "" {
			err = h.lakebaseService.UpdateColumnDescription(ctx, dsID, col.TableName, col.ColumnName, description, "llm", 0.85)
			if err == nil {
				columnsUpdated++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"datasource_id":   dsID,
		"tables_updated":  tablesUpdated,
		"columns_updated": columnsUpdated,
		"total_tables":    len(tables),
		"total_columns":   len(columns),
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

// GenerateRichContextStream generates Rich Context with SSE progress streaming
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
		Concurrency int  `json:"concurrency"` // Number of parallel workers
		Force       bool `json:"force"`       // Force regenerate even if exists
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Concurrency = 3 // default
		req.Force = false
	}
	if req.Concurrency < 1 {
		req.Concurrency = 1
	}
	if req.Concurrency > 10 {
		req.Concurrency = 10
	}

	// Resolve datasource ID
	idStr := c.Param("id")
	ctx := c.Request.Context()

	var dsID int64
	var err error
	dsID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ds, lookupErr := h.lakebaseService.GetDatasourceByName(ctx, idStr)
		if lookupErr != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found: " + idStr})
			return
		}
		dsID = ds.ID
	}

	// Get tables and columns
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

	// Group columns by table
	columnsByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, col := range columns {
		columnsByTable[col.TableName] = append(columnsByTable[col.TableName], col)
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.Flush()

	// Event channel for thread-safe sending
	eventChan := make(chan GenerateContextEvent, 100)
	done := make(chan struct{})

	// Event sender goroutine
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

	// Phase 1: Coordinator - Discovery
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "coordinator",
		Phase:   "discovery",
		Status:  "running",
		Message: "Starting table discovery...",
	})

	tableNames := make([]string, len(tables))
	for i, t := range tables {
		tableNames[i] = t.TableName
	}

	sendEvent("agent_step", GenerateContextEvent{
		Agent:   "coordinator",
		Phase:   "discovery",
		Message: fmt.Sprintf("Found %d tables", len(tables)),
		Data:    map[string]interface{}{"tables": tableNames, "total_columns": len(columns)},
	})

	sendEvent("agent_done", GenerateContextEvent{
		Agent:   "coordinator",
		Phase:   "discovery",
		Status:  "success",
		Message: "Discovery complete",
	})

	// Phase 2: Worker Agents - Process tables with concurrency control
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Concurrency)

	var mu sync.Mutex
	tablesUpdated := 0
	columnsUpdated := 0

	for i, table := range tables {
		wg.Add(1)
		workerID := fmt.Sprintf("worker-%d", i+1)

		go func(wID string, tbl *lakebase.TableInfo, cols []*lakebase.ColumnInfo) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Start worker
			sendEvent("agent_start", GenerateContextEvent{
				Agent:   wID,
				Table:   tbl.TableName,
				Phase:   "schema",
				Status:  "running",
				Message: fmt.Sprintf("Analyzing %s...", tbl.TableName),
			})

			// Check if needs update
			needsTableUpdate := req.Force || !tbl.Description.Valid || tbl.Description.String == ""

			// Phase: Schema analysis
			sendEvent("agent_step", GenerateContextEvent{
				Agent:   wID,
				Table:   tbl.TableName,
				Phase:   "schema",
				Message: fmt.Sprintf("Found %d columns", len(cols)),
				Data:    map[string]interface{}{"column_count": len(cols)},
			})

			// Phase: Rich Context generation
			if needsTableUpdate {
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   wID,
					Table:   tbl.TableName,
					Phase:   "rich_context",
					Message: "Generating description via LLM...",
				})

				// Build column list
				colNames := make([]string, len(cols))
				for j, col := range cols {
					colNames[j] = col.ColumnName
				}
				colList := strings.Join(colNames, ", ")

				// Generate table description
				prompt := fmt.Sprintf(`Analyze this database table and provide a concise description.

Table: %s
Columns: %s
Row Count: %d

Generate a one-sentence description of what this table stores and its business purpose.
Output only the description text, no JSON or formatting.`, tbl.TableName, colList, tbl.RowCount)

				response, err := llms.GenerateFromSinglePrompt(ctx, model, prompt)
				if err != nil {
					sendEvent("agent_step", GenerateContextEvent{
						Agent:   wID,
						Table:   tbl.TableName,
						Phase:   "rich_context",
						Status:  "error",
						Message: fmt.Sprintf("LLM error: %v", err),
					})
				} else {
					description := strings.TrimSpace(response)
					if description != "" {
						// Save to database
						sendEvent("agent_step", GenerateContextEvent{
							Agent:   wID,
							Table:   tbl.TableName,
							Phase:   "storage",
							Message: "Saving to rc_tables...",
						})

						err = h.lakebaseService.UpdateTableDescription(ctx, dsID, tbl.TableName, description, "llm", 0.85)
						if err == nil {
							mu.Lock()
							tablesUpdated++
							mu.Unlock()

							sendEvent("storage", GenerateContextEvent{
								Agent:   wID,
								Table:   tbl.TableName,
								Phase:   "storage",
								Status:  "success",
								Message: fmt.Sprintf("Saved table: %s", tbl.TableName),
								Data:    map[string]interface{}{"target": "rc_tables", "description": description},
							})
						}
					}
				}
			} else {
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   wID,
					Table:   tbl.TableName,
					Phase:   "rich_context",
					Message: "Already has description, skipping",
				})
			}

			// Process columns
			for _, col := range cols {
				needsColUpdate := req.Force || !col.Description.Valid || col.Description.String == ""
				if !needsColUpdate {
					continue
				}

				prompt := fmt.Sprintf(`Analyze this database column and provide a concise description.

Table: %s
Column: %s
Data Type: %s
Is Primary Key: %v
Is Foreign Key: %v

Generate a one-sentence description. Output only the description text.`,
					tbl.TableName, col.ColumnName,
					col.DataType.String, col.IsPrimaryKey, col.IsForeignKey)

				response, err := llms.GenerateFromSinglePrompt(ctx, model, prompt)
				if err != nil {
					continue
				}

				description := strings.TrimSpace(response)
				if description != "" {
					err = h.lakebaseService.UpdateColumnDescription(ctx, dsID, tbl.TableName, col.ColumnName, description, "llm", 0.85)
					if err == nil {
						mu.Lock()
						columnsUpdated++
						mu.Unlock()

						sendEvent("storage", GenerateContextEvent{
							Agent:   wID,
							Table:   tbl.TableName,
							Phase:   "storage",
							Status:  "success",
							Message: fmt.Sprintf("Saved column: %s.%s", tbl.TableName, col.ColumnName),
							Data:    map[string]interface{}{"target": "rc_columns", "column": col.ColumnName},
						})
					}
				}
			}

			// Worker done
			sendEvent("agent_done", GenerateContextEvent{
				Agent:   wID,
				Table:   tbl.TableName,
				Status:  "success",
				Message: fmt.Sprintf("Completed %s", tbl.TableName),
			})
		}(workerID, table, columnsByTable[table.TableName])
	}

	// Wait for all workers
	wg.Wait()

	// Complete
	duration := time.Since(startTime)
	sendEvent("complete", GenerateContextEvent{
		Status:  "success",
		Message: "Generation complete",
		Data: map[string]interface{}{
			"tables_updated":  tablesUpdated,
			"columns_updated": columnsUpdated,
			"total_tables":    len(tables),
			"total_columns":   len(columns),
			"duration_ms":     duration.Milliseconds(),
		},
	})

	// Close event channel and wait for sender to finish
	close(eventChan)
	<-done
}
