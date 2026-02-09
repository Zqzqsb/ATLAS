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

	"strings"

	"lucid/config"
	"lucid/internal/adapter"
	"lucid/internal/lakebase"
	"lucid/internal/react"
	"lucid/internal/react/scenarios"
	reacttools "lucid/internal/react/tools"
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

	// Auto-generate embeddings for all context (one-stop pipeline)
	var embeddingsGenerated int
	embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, dsID)
	if embErr == nil && embResult != nil {
		embeddingsGenerated = embResult.TotalEmbeddings
	}

	c.JSON(http.StatusOK, gin.H{
		"success":              true,
		"datasource_id":        dsID,
		"tables_updated":       tablesUpdated,
		"columns_updated":      columnsUpdated,
		"total_tables":         len(tables),
		"total_columns":        len(columns),
		"embeddings_generated": embeddingsGenerated,
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
	idStr := c.Param("id")
	ctx := c.Request.Context()

	var dsID int64
	var ds *lakebase.Datasource
	var err error
	dsID, err = strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ds, err = h.lakebaseService.GetDatasourceByName(ctx, idStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found: " + idStr})
			return
		}
		dsID = ds.ID
	} else {
		ds, err = h.lakebaseService.GetDatasource(ctx, dsID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found: " + idStr})
			return
		}
	}

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

	// Phase 1: Announce start
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "rc_gen",
		Phase:   "init",
		Status:  "running",
		Message: fmt.Sprintf("Starting ReAct Rich Context generation for %s (%d tables, %d columns)", ds.Name, len(tables), len(columns)),
	})

	// Phase 2: Run ReAct agent
	rcWriter := reacttools.NewLakebaseRCWriter(h.lakebaseService.GetRepository())
	engineCfg := scenarios.BuildRCGenEngine(businessDB, rcWriter, scenarios.RCGenConfig{
		DatasourceID:  dsID,
		Tables:        tables,
		Columns:       columns,
		Relations:     relations,
		MaxIterations: req.MaxIterations,
		MinIterations: req.MinIterations,
		Force:         req.Force,
		StepCallback: func(step react.Step, eventType string) {
			// Map ReAct steps to SSE events
			sendEvent("agent_step", GenerateContextEvent{
				Agent:   "rc_gen",
				Phase:   eventType,
				Message: step.Thought,
				Data: map[string]interface{}{
					"iteration":    step.Iteration,
					"action":       step.Action,
					"action_input": step.ActionInput,
					"observation":  step.Observation,
				},
			})
		},
	})

	engine := react.New(model, engineCfg)
	result, execErr := engine.Execute(ctx, "")

	if execErr != nil {
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "rc_gen",
			Status:  "error",
			Message: fmt.Sprintf("ReAct agent error: %v", execErr),
		})
	} else {
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

	// Phase 3: Generate embeddings
	sendEvent("agent_start", GenerateContextEvent{
		Agent:   "embedding",
		Phase:   "embedding",
		Status:  "running",
		Message: "Generating embeddings for semantic search...",
	})

	var embeddingsGenerated int
	embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, dsID)
	if embErr != nil {
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "embedding",
			Phase:   "embedding",
			Status:  "error",
			Message: fmt.Sprintf("Embedding error: %v", embErr),
		})
	} else if embResult != nil {
		embeddingsGenerated = embResult.TotalEmbeddings
		sendEvent("agent_done", GenerateContextEvent{
			Agent:   "embedding",
			Phase:   "embedding",
			Status:  "success",
			Message: fmt.Sprintf("Generated %d embeddings", embeddingsGenerated),
			Data: map[string]interface{}{
				"tables_embedded":  embResult.TablesProcessed,
				"columns_embedded": embResult.ColumnsProcessed,
				"total_embeddings": embeddingsGenerated,
			},
		})
	}

	// Complete
	duration := time.Since(startTime)
	sendEvent("complete", GenerateContextEvent{
		Status:  "success",
		Message: "Generation complete",
		Data: map[string]interface{}{
			"total_tables":         len(tables),
			"total_columns":        len(columns),
			"react_iterations":     0,
			"embeddings_generated": embeddingsGenerated,
			"duration_ms":          duration.Milliseconds(),
		},
	})
	if result != nil {
		// Update with actual iterations
		sendEvent("complete", GenerateContextEvent{
			Status: "success",
			Data:   map[string]interface{}{"react_iterations": result.Iterations},
		})
	}

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

	dsIDStr := c.Param("id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Verify datasource exists
	ds, err := h.lakebaseService.GetDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Datasource not found",
		})
		return
	}

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

	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Get datasource to find its name (used as adapter key)
	ds, err := h.lakebaseService.GetDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found"})
		return
	}

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

	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	ds, err := h.lakebaseService.GetDatasource(ctx, dsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Datasource not found"})
		return
	}

	// Delete the datasource record (CASCADE will remove all associated rc_* data)
	if err := h.lakebaseService.DeleteDatasource(ctx, dsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete datasource: %v", err),
		})
		return
	}

	// Also remove from in-memory connection config so it can be re-added
	connID := ds.Name
	h.dbService.CloseAdapter(connID)
	newDatabases := make([]config.DatabaseConfig, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
		if db.ID != connID {
			newDatabases = append(newDatabases, db)
		}
	}
	h.config.Databases = newDatabases

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    fmt.Sprintf("Datasource '%s' and all associated data deleted", ds.Name),
		"datasource": ds.Name,
	})
}
