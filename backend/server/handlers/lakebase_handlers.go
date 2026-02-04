package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

	// Convert to safe response (hide passwords)
	result := make([]map[string]interface{}, len(datasources))
	for i, ds := range datasources {
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
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"datasources": result,
		"count":       len(result),
	})
}

// GetLakebaseDatasource returns details for a specific datasource
func (h *Handler) GetLakebaseDatasource(c *gin.Context) {
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Lake-base service not available",
		})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ds, err := h.lakebaseService.GetDatasource(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get schema for this datasource
	schemas, _ := h.lakebaseService.GetSchemaByDatasource(ctx, id)

	// Get context for this datasource
	contexts, _ := h.lakebaseService.GetContextByDatasource(ctx, id)

	// Get embedding count
	embeddingCount, _ := h.lakebaseService.CountEmbeddings(ctx, id)

	// Build table summary
	tableMap := make(map[string]int)
	for _, s := range schemas {
		tableMap[s.TableName]++
	}

	tables := make([]map[string]interface{}, 0, len(tableMap))
	for tableName, colCount := range tableMap {
		tables = append(tables, map[string]interface{}{
			"name":         tableName,
			"column_count": colCount,
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
		"tables_count":     len(tableMap),
		"columns_count":    len(schemas),
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
