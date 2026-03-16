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
		terms, _ := h.lakebaseService.GetTermsByDatasource(ctx, ds.ID)

		// Count ALL context types — consistent with detail API
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
			if c.SampleValues.Valid && c.SampleValues.String != "" {
				contextCount++
			}
			if c.Synonyms.Valid && c.Synonyms.String != "" {
				contextCount++
			}
		}
		contextCount += len(terms) // business_rule contexts

		// Datasource-level description
		desc := ""
		if ds.Description.Valid {
			desc = ds.Description.String
		}

		result[i] = map[string]interface{}{
			"id":            ds.ID,
			"name":          ds.Name,
			"db_type":       ds.DBType,
			"host":          ds.Host,
			"port":          ds.Port,
			"database_name": ds.DatabaseName,
			"description":   desc,
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

// PreviewForestChunks returns forest decomposition and context coverage for each cluster
// without actually running generation. Used by the frontend to show a preview before starting.
// GET /api/v1/lakebase/datasources/:id/generate-context/preview
func (h *Handler) PreviewForestChunks(c *gin.Context) {
	if h.lakebaseService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not configured"})
		return
	}

	ds, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	_ = ds

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
	relations, _ := h.lakebaseService.GetRelationsByDatasource(ctx, dsID)

	const forestThreshold = 30
	tableCount := len(tables)

	if tableCount <= forestThreshold {
		c.JSON(http.StatusOK, gin.H{
			"mode":        "single_agent",
			"table_count": tableCount,
		})
		return
	}

	forestResult := scenarios.ForestDecompose(tables, columns, relations)
	clusters := scenarios.MergeIsolatedTables(forestResult.Clusters, 15)

	type clusterPreview struct {
		Index          int      `json:"index"`
		TableCount     int      `json:"table_count"`
		RelationCount  int      `json:"relation_count"`
		Tables         []string `json:"tables"`
		TablesWithCtx  int      `json:"tables_with_context"`    // tables that have description
		ColumnsTotal   int      `json:"columns_total"`
		ColumnsWithCtx int      `json:"columns_with_context"`   // columns that have description
		WillSkip       bool     `json:"will_skip"`              // >=90% tables have context
		CoverageRatio  float64  `json:"coverage_ratio"`         // table description coverage 0.0-1.0
		MinIter        int      `json:"min_iter"`
		MaxIter        int      `json:"max_iter"`
	}

	previews := make([]clusterPreview, len(clusters))
	totalSkip := 0
	totalNeed := 0

	for i, cl := range clusters {
		tableNames := make([]string, len(cl.Tables))
		tablesWithCtx := 0
		for j, t := range cl.Tables {
			tableNames[j] = t.TableName
			if t.Description.Valid && t.Description.String != "" {
				tablesWithCtx++
			}
		}

		colsTotal := len(cl.Columns)
		colsWithCtx := 0
		for _, col := range cl.Columns {
			if col.Description.Valid && col.Description.String != "" {
				colsWithCtx++
			}
		}

		willSkip := clusterHasContext(cl)
		if willSkip {
			totalSkip++
		} else {
			totalNeed++
		}

		chunkMin, chunkMax := scenarios.ComputeChunkBudget(len(cl.Tables))

		previews[i] = clusterPreview{
			Index:          i,
			TableCount:     len(cl.Tables),
			RelationCount:  len(cl.Relations),
			Tables:         tableNames,
			TablesWithCtx:  tablesWithCtx,
			ColumnsTotal:   colsTotal,
			ColumnsWithCtx: colsWithCtx,
			WillSkip:       willSkip,
			CoverageRatio:  func() float64 { if len(cl.Tables) == 0 { return 1.0 }; return float64(tablesWithCtx) / float64(len(cl.Tables)) }(),
			MinIter:        chunkMin,
			MaxIter:        chunkMax,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"mode":            "forest_chunked",
		"table_count":     tableCount,
		"column_count":    len(columns),
		"clusters_total":  len(clusters),
		"clusters_skip":   totalSkip,
		"clusters_need":   totalNeed,
		"largest_cluster": forestResult.LargestSize,
		"median_cluster":  forestResult.MedianSize,
		"isolated_tables": forestResult.IsolatedCount,
		"clusters":        previews,
	})
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

	const forestThreshold = 30
	tableCount := len(tables)

	// Create RC writer with storage callback to notify frontend on each write
	// and trigger incremental embedding immediately
	rcWriter := reacttools.NewLakebaseRCWriter(h.lakebaseService.GetRepository())

	// Track incremental embedding progress.
	var embeddingsExpected int
	var embeddingsCompleted int
	var embeddingMu sync.Mutex

	setupEmbeddingCallback := func(rcw *reacttools.LakebaseRCWriter) {
		rcw.SetOnWrite(func(contextType, tableName, columnName string) {
			target := "rc_tables"
			detail := tableName
			if columnName != "" {
				target = "rc_columns"
				detail = tableName + "." + columnName
			}
			if contextType == "business_term" {
				target = "rc_terms"
			}

			embeddingMu.Lock()
			embeddingsExpected++
			embeddingMu.Unlock()

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
				embeddingsCompleted++
				completed := embeddingsCompleted
				expected := embeddingsExpected
				embeddingMu.Unlock()
				sendEvent("agent_step", GenerateContextEvent{
					Agent:   "embedding",
					Phase:   "embedding",
					Message: fmt.Sprintf("🧬 Embedded %s: %s (%d/%d)", ct, detail, completed, expected),
					Data: map[string]interface{}{
						"context_type":      ct,
						"table":             tn,
						"column":            cn,
						"embeddings_so_far": completed,
						"embeddings_total":  expected,
					},
				})
			}(contextType, tableName, columnName)
		})
	}

	// Announce embedding agent is running
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "embedding",
		Phase:   "embedding",
		Status:  "running",
		Message: "Streaming embeddings — each RC write triggers immediate embedding...",
	})

	var totalIterations int

	if tableCount <= forestThreshold {
		// === Small schema: single-agent path (original) ===
		// Count existing context for progress baseline
		tablesExisting := 0
		for _, t := range tables {
			if t.Description.Valid && t.Description.String != "" {
				tablesExisting++
			}
		}
		columnsExisting := 0
		for _, col := range columns {
			if col.Description.Valid && col.Description.String != "" {
				columnsExisting++
			}
		}
		sendEvent("agent_start", GenerateContextEvent{
			Agent:   "rc_gen",
			Phase:   "init",
			Status:  "running",
			Message: fmt.Sprintf("Starting ReAct Rich Context generation for %s (%d tables, %d columns)", ds.Name, len(tables), len(columns)),
			Data: map[string]interface{}{
				"tables_total":    len(tables),
				"columns_total":   len(columns),
				"tables_existing": tablesExisting,
				"columns_existing": columnsExisting,
				"mode":            "single_agent",
			},
		})

		setupEmbeddingCallback(rcWriter)

		engineCfg := scenarios.BuildRCGenEngine(businessDB, rcWriter, scenarios.RCGenConfig{
			DatasourceID:  dsID,
			Tables:        tables,
			Columns:       columns,
			Relations:     relations,
			MaxIterations: req.MaxIterations,
			MinIterations: req.MinIterations,
			Force:         req.Force,
			StepCallback: func(step react.Step, eventType string) {
				h.sendRCGenStep(sendEvent, step, eventType, -1)
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
			totalIterations = result.Iterations
			log.Info("ReAct agent completed", "iterations", result.Iterations, "duration_s", result.Duration.Seconds())
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
	} else {
		// === Large schema: forest-based chunked path ===
		forestResult := scenarios.ForestDecompose(tables, columns, relations)
		clusters := scenarios.MergeIsolatedTables(forestResult.Clusters, 15)

		// Build per-cluster metadata for treemap visualization
		clustersMeta := make([]map[string]interface{}, len(clusters))
		for i, cl := range clusters {
			tableNames := make([]string, len(cl.Tables))
			for j, t := range cl.Tables {
				tableNames[j] = t.TableName
			}
			// Compute coverage so frontend can show initial skip status in treemap
			withCtx := 0
			for _, t := range cl.Tables {
				if t.Description.Valid && t.Description.String != "" {
					withCtx++
				}
			}
			coverage := 0.0
			if len(cl.Tables) > 0 {
				coverage = float64(withCtx) / float64(len(cl.Tables))
			}
			willSkip := !req.Force && clusterHasContext(cl)
			clustersMeta[i] = map[string]interface{}{
				"index":          i,
				"table_count":    len(cl.Tables),
				"relation_count": len(cl.Relations),
				"tables":         tableNames,
				"will_skip":      willSkip,
				"coverage_ratio": coverage,
			}
		}

		// Count existing context (tables/columns that already have descriptions)
		// so frontend progress bars can include pre-existing data in their totals.
		tablesWithExistingCtx := 0
		for _, t := range tables {
			if t.Description.Valid && t.Description.String != "" {
				tablesWithExistingCtx++
			}
		}
		columnsWithExistingCtx := 0
		for _, col := range columns {
			if col.Description.Valid && col.Description.String != "" {
				columnsWithExistingCtx++
			}
		}

		sendEvent("agent_start", GenerateContextEvent{
			Agent:   "rc_gen",
			Phase:   "init",
			Status:  "running",
			Message: fmt.Sprintf("Starting forest-based chunked generation for %s (%d tables → %d clusters)", ds.Name, tableCount, len(clusters)),
			Data: map[string]interface{}{
				"tables_total":    tableCount,
				"columns_total":   len(columns),
				"tables_existing": tablesWithExistingCtx,
				"columns_existing": columnsWithExistingCtx,
				"mode":            "forest_chunked",
				"clusters_total":  len(clusters),
				"largest_cluster": forestResult.LargestSize,
				"median_cluster":  forestResult.MedianSize,
				"isolated_tables": forestResult.IsolatedCount,
				"clusters":        clustersMeta,
			},
		})

		setupEmbeddingCallback(rcWriter)

		var skippedChunks int
		for ci, cluster := range clusters {
			clusterTableNames := make([]string, len(cluster.Tables))
			for i, t := range cluster.Tables {
				clusterTableNames[i] = t.TableName
			}

		// Skip chunks where >=90% tables already have descriptions (unless Force is set)
			if !req.Force && clusterHasContext(cluster) {
				skippedChunks++
				// Count tables with context for the skip message
				tablesWithCtx := 0
				for _, t := range cluster.Tables {
					if t.Description.Valid && t.Description.String != "" {
						tablesWithCtx++
					}
				}
				sendEvent("chunk_complete", GenerateContextEvent{
					Agent:   "rc_gen",
					Phase:   "chunk",
					Status:  "skipped",
					Message: fmt.Sprintf("⏭ Chunk %d/%d skipped: %d/%d tables have context (≥90%%)", ci+1, len(clusters), tablesWithCtx, len(cluster.Tables)),
					Data: map[string]interface{}{
						"chunk_index": ci,
						"chunk_total": len(clusters),
						"skipped":     true,
						"iterations":  0,
						"duration":    0,
					},
				})
				continue
			}

			chunkMin, chunkMax := scenarios.ComputeChunkBudget(len(cluster.Tables))

			sendEvent("chunk_start", GenerateContextEvent{
				Agent:   "rc_gen",
				Phase:   "chunk",
				Status:  "running",
				Message: fmt.Sprintf("🌲 Chunk %d/%d: %d tables (%d relations), budget %d-%d iters", ci+1, len(clusters), len(cluster.Tables), len(cluster.Relations), chunkMin, chunkMax),
				Data: map[string]interface{}{
					"chunk_index":    ci,
					"chunk_total":    len(clusters),
					"tables":         clusterTableNames,
					"table_count":    len(cluster.Tables),
					"relation_count": len(cluster.Relations),
					"max_iterations": chunkMax,
					"min_iterations": chunkMin,
				},
			})

			engineCfg := scenarios.BuildRCGenEngine(businessDB, rcWriter, scenarios.RCGenConfig{
				DatasourceID:  dsID,
				Tables:        cluster.Tables,
				Columns:       cluster.Columns,
				Relations:     cluster.Relations,
				MaxIterations: chunkMax,
				MinIterations: chunkMin,
				Force:         req.Force,
				StepCallback: func(step react.Step, eventType string) {
					h.sendRCGenStep(sendEvent, step, eventType, ci)
				},
			})

			engine := react.New(model, engineCfg)
			result, execErr := engine.Execute(ctx, "")

			if execErr != nil {
				log.Error("chunk agent failed", "chunk", ci, "error", execErr)
				sendEvent("chunk_error", GenerateContextEvent{
					Agent:   "rc_gen",
					Phase:   "chunk",
					Status:  "error",
					Message: fmt.Sprintf("Chunk %d/%d error: %v", ci+1, len(clusters), execErr),
					Data: map[string]interface{}{
						"chunk_index": ci,
						"error":       execErr.Error(),
					},
				})
				continue
			}

			totalIterations += result.Iterations
			sendEvent("chunk_complete", GenerateContextEvent{
				Agent:   "rc_gen",
				Phase:   "chunk",
				Status:  "success",
				Message: fmt.Sprintf("✓ Chunk %d/%d done: %d iterations (%.1fs)", ci+1, len(clusters), result.Iterations, result.Duration.Seconds()),
				Data: map[string]interface{}{
					"chunk_index": ci,
					"iterations":  result.Iterations,
					"duration":    result.Duration.Seconds(),
				},
			})
		}

		doneMsg := fmt.Sprintf("Forest-based generation completed: %d clusters, %d total iterations", len(clusters), totalIterations)
		if skippedChunks > 0 {
			doneMsg = fmt.Sprintf("Forest-based generation completed: %d clusters (%d skipped), %d total iterations", len(clusters), skippedChunks, totalIterations)
		}
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "rc_gen",
			Status:  "success",
			Message: doneMsg,
			Data: map[string]interface{}{
				"iterations":     totalIterations,
				"clusters":       len(clusters),
				"skipped_chunks": skippedChunks,
				"duration":       time.Since(startTime).Seconds(),
			},
		})
	}

	// Phase 3: Catch-up embeddings for any entities that may have been missed
	embeddingMu.Lock()
	streamEmbedded := embeddingsCompleted
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
			"react_iterations":     totalIterations,
			"embeddings_generated": totalEmb,
			"stream_embedded":      streamEmbedded,
			"duration_ms":          duration.Milliseconds(),
		},
	})

	close(eventChan)
	<-done
}

// sendRCGenStep sends an SSE event for a ReAct step (shared by single-agent and chunked paths).
// chunkIndex < 0 means single-agent mode (no chunk info).
func (h *Handler) sendRCGenStep(
	sendEvent func(string, GenerateContextEvent),
	step react.Step,
	eventType string,
	chunkIndex int,
) {
	buildData := func(extra map[string]interface{}) map[string]interface{} {
		d := map[string]interface{}{"iteration": step.Iteration}
		if chunkIndex >= 0 {
			d["chunk"] = chunkIndex
		}
		for k, v := range extra {
			d[k] = v
		}
		return d
	}

	switch eventType {
	case "thought":
		if step.Thought != "" {
			sendEvent("agent_step", GenerateContextEvent{
				Agent:   "rc_gen",
				Phase:   "thought",
				Message: step.Thought,
				Data:    buildData(nil),
			})
		}
	case "action":
		sendEvent("agent_step", GenerateContextEvent{
			Agent:   "rc_gen",
			Phase:   "action",
			Message: fmt.Sprintf("🔧 %s", step.Action),
			Data:    buildData(map[string]interface{}{"action": step.Action, "action_input": step.ActionInput}),
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
			Data:    buildData(map[string]interface{}{"action": step.Action, "observation": step.Observation}),
		})
	case "finish":
		sendEvent("agent_step", GenerateContextEvent{
			Agent:   "rc_gen",
			Phase:   "finish",
			Message: step.Thought,
			Data:    buildData(nil),
		})
	}
}

// clusterHasContext returns true if enough tables in the cluster already have
// non-empty descriptions, indicating it can be skipped.
// Uses a 90% threshold: for a 70-table cluster, 63+ tables with context = skip.
// This accommodates LLM agents that may miss a few tables in large clusters.
func clusterHasContext(cluster *scenarios.TableCluster) bool {
	const skipThreshold = 0.9 // 90% of tables must have context

	if len(cluster.Tables) == 0 {
		return true
	}
	withCtx := 0
	for _, t := range cluster.Tables {
		if t.Description.Valid && t.Description.String != "" {
			withCtx++
		}
	}
	ratio := float64(withCtx) / float64(len(cluster.Tables))
	return ratio >= skipThreshold
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

// ===========================================
// Manual Context CRUD API
// ===========================================

// addContextRequest represents the request body for adding a manual context
type addContextRequest struct {
	TableName  string `json:"table_name" binding:"required"`
	ColumnName string `json:"column_name"`                        // empty = table-level
	Type       string `json:"type" binding:"required"`            // description, example, synonym, value_mapping, business_rule, constraint, calculation
	Content    string `json:"content" binding:"required"`
}

// AddContext manually adds a Rich Context entry for a datasource
// POST /api/v1/lakebase/datasources/:id/context
func (h *Handler) AddContext(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not available"})
		return
	}

	_, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	var req addContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	log := logger.With("component", "add_context", "dsID", dsID, "table", req.TableName, "column", req.ColumnName, "type", req.Type)
	log.Info("manually adding context")

	var contextType string // for embedding callback
	var writeErr error

	switch req.Type {
	case "description":
		if req.ColumnName == "" {
			// Table-level description → rc_tables.description
			writeErr = h.lakebaseService.UpdateTableDescription(ctx, dsID, req.TableName, req.Content, "manual", 1.0)
			contextType = "table_description"
		} else {
			// Column-level description → rc_columns.description
			writeErr = h.lakebaseService.UpdateColumnDescription(ctx, dsID, req.TableName, req.ColumnName, req.Content, "manual", 1.0)
			contextType = "column_description"
		}
	case "example":
		if req.ColumnName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Example context requires a column name"})
			return
		}
		writeErr = h.lakebaseService.UpdateColumnSampleValues(ctx, dsID, req.TableName, req.ColumnName, req.Content)
		contextType = "column_sample_values"
	case "synonym":
		if req.ColumnName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Synonym context requires a column name"})
			return
		}
		writeErr = h.lakebaseService.UpdateColumnSynonyms(ctx, dsID, req.TableName, req.ColumnName, req.Content)
		contextType = "column_synonyms"
	case "business_rule", "constraint", "calculation", "value_mapping":
		// Store as a business term in rc_terms
		term := req.TableName
		if req.ColumnName != "" {
			term = req.TableName + "." + req.ColumnName
		}
		writeErr = h.lakebaseService.UpsertTerm(ctx, dsID, term, req.Content, "", "", req.Type)
		contextType = "business_term"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported context type: " + req.Type})
		return
	}

	if writeErr != nil {
		log.Error("failed to write context", "error", writeErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save context: " + writeErr.Error()})
		return
	}

	log.Info("context saved, triggering embedding")

	// Trigger embedding asynchronously
	go func() {
		embCtx, embCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer embCancel()
		if embErr := h.lakebaseService.EmbedEntityByName(embCtx, dsID, contextType, req.TableName, req.ColumnName); embErr != nil {
			logger.With("component", "add_context").Warn("embedding after manual add failed", "error", embErr)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      fmt.Sprintf("Context added for %s", req.TableName),
		"table_name":   req.TableName,
		"column_name":  req.ColumnName,
		"context_type": req.Type,
	})
}

// deleteContextRequest represents the request body for deleting a manual context
type deleteContextRequest struct {
	TableName  string `json:"table_name" binding:"required"`
	ColumnName string `json:"column_name"`
	Type       string `json:"type" binding:"required"`
}

// DeleteContext removes a specific Rich Context entry for a datasource
// DELETE /api/v1/lakebase/datasources/:id/context
func (h *Handler) DeleteContext(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not available"})
		return
	}

	_, dsID, ok := h.resolveDatasource(c)
	if !ok {
		return
	}

	var req deleteContextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	log := logger.With("component", "delete_context", "dsID", dsID, "table", req.TableName, "column", req.ColumnName, "type", req.Type)
	log.Info("manually deleting context")

	var writeErr error

	switch req.Type {
	case "description":
		if req.ColumnName == "" {
			writeErr = h.lakebaseService.UpdateTableDescription(ctx, dsID, req.TableName, "", "manual", 0)
		} else {
			writeErr = h.lakebaseService.UpdateColumnDescription(ctx, dsID, req.TableName, req.ColumnName, "", "manual", 0)
		}
	case "example":
		writeErr = h.lakebaseService.UpdateColumnSampleValues(ctx, dsID, req.TableName, req.ColumnName, "")
	case "synonym":
		writeErr = h.lakebaseService.UpdateColumnSynonyms(ctx, dsID, req.TableName, req.ColumnName, "")
	case "business_rule", "constraint", "calculation", "value_mapping":
		// For terms: delete the term from rc_terms by setting content empty
		// The repo doesn't have a dedicated delete-term-by-name, so we use UpsertTerm with empty content
		term := req.TableName
		if req.ColumnName != "" {
			term = req.TableName + "." + req.ColumnName
		}
		writeErr = h.lakebaseService.UpsertTerm(ctx, dsID, term, "", "", "", req.Type)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported context type: " + req.Type})
		return
	}

	if writeErr != nil {
		log.Error("failed to delete context", "error", writeErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete context: " + writeErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Context deleted for %s", req.TableName),
	})
}
