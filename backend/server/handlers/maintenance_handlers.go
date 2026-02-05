package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/agent"
	ctx "lucid/internal/context"
	"lucid/internal/lakebase"
)

// GetMaintenanceReport returns the maintenance status of Rich Context
// GET /api/context/maintenance/:connection_id
func (h *Handler) GetMaintenanceReport(c *gin.Context) {
	connectionID := c.Param("connection_id")

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rich Context not found: " + err.Error(),
		})
		return
	}

	maintainer := ctx.NewRichContextMaintainer(richContext, ctx.MaintainerConfig{})
	report := maintainer.CheckExpiration()

	c.JSON(http.StatusOK, report)
}

// GetExpiredEntries returns only expired Rich Context entries
// GET /api/context/expired/:connection_id
func (h *Handler) GetExpiredEntries(c *gin.Context) {
	connectionID := c.Param("connection_id")

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rich Context not found: " + err.Error(),
		})
		return
	}

	maintainer := ctx.NewRichContextMaintainer(richContext, ctx.MaintainerConfig{})
	expired := maintainer.GetExpiredEntries()

	c.JSON(http.StatusOK, gin.H{
		"count":   len(expired),
		"entries": expired,
	})
}

// UpdateRichContextEntryRequest represents a request to update a Rich Context entry
type UpdateRichContextEntryRequest struct {
	TableName string `json:"table_name" binding:"required"`
	Key       string `json:"key" binding:"required"`
	Content   string `json:"content" binding:"required"`
	Source    string `json:"source"` // catalog, user, llm, auto_corrected
	Reason    string `json:"reason"` // Optional reason for the update
}

// UpdateRichContextEntry updates a single Rich Context entry with maintenance tracking
// POST /api/context/update/:connection_id
func (h *Handler) UpdateRichContextEntry(c *gin.Context) {
	connectionID := c.Param("connection_id")

	var req UpdateRichContextEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rich Context not found: " + err.Error(),
		})
		return
	}

	// Map source string to CatalogSource
	source := ctx.SourceUser
	switch req.Source {
	case "catalog":
		source = ctx.SourceCatalog
	case "llm":
		source = ctx.SourceLLM
	case "auto_corrected":
		source = ctx.SourceAutoCorrected
	case "analysis":
		source = ctx.SourceAnalysis
	}

	maintainer := ctx.NewRichContextMaintainer(richContext, ctx.MaintainerConfig{})
	if err := maintainer.UpdateEntry(req.TableName, req.Key, req.Content, source, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update entry: " + err.Error(),
		})
		return
	}

	// Save the updated context
	if err := h.dbService.SaveRichContext(connectionID, richContext); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save Rich Context: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rich Context entry updated",
		"table":   req.TableName,
		"key":     req.Key,
		"source":  string(source),
	})
}

// BatchUpdateRequest represents a batch update request
type BatchUpdateRequest struct {
	Updates []UpdateRichContextEntryRequest `json:"updates" binding:"required"`
}

// BatchUpdateRichContext updates multiple Rich Context entries at once
// POST /api/context/batch-update/:connection_id
func (h *Handler) BatchUpdateRichContext(c *gin.Context) {
	connectionID := c.Param("connection_id")

	var req BatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rich Context not found: " + err.Error(),
		})
		return
	}

	maintainer := ctx.NewRichContextMaintainer(richContext, ctx.MaintainerConfig{})
	
	successCount := 0
	var errors []string

	for _, update := range req.Updates {
		source := ctx.SourceUser
		switch update.Source {
		case "catalog":
			source = ctx.SourceCatalog
		case "llm":
			source = ctx.SourceLLM
		case "auto_corrected":
			source = ctx.SourceAutoCorrected
		case "analysis":
			source = ctx.SourceAnalysis
		}

		if err := maintainer.UpdateEntry(update.TableName, update.Key, update.Content, source, update.Reason); err != nil {
			errors = append(errors, err.Error())
		} else {
			successCount++
		}
	}

	// Save the updated context
	if err := h.dbService.SaveRichContext(connectionID, richContext); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save Rich Context: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       len(errors) == 0,
		"updated_count": successCount,
		"error_count":   len(errors),
		"errors":        errors,
	})
}

// RefreshExpiredEntries triggers re-analysis of expired entries
// POST /api/context/refresh/:connection_id
func (h *Handler) RefreshExpiredEntries(c *gin.Context) {
	connectionID := c.Param("connection_id")

	richContext, err := h.dbService.GetRichContext(connectionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rich Context not found: " + err.Error(),
		})
		return
	}

	maintainer := ctx.NewRichContextMaintainer(richContext, ctx.MaintainerConfig{})
	expired := maintainer.GetExpiredEntries()

	// In a real implementation, this would trigger re-analysis using LLM
	// For now, we just return the list of entries that need refresh

	c.JSON(http.StatusOK, gin.H{
		"message":        "The following entries are expired and should be refreshed",
		"expired_count":  len(expired),
		"expired_entries": expired,
		"action":         "Run onboarding analysis to refresh these entries",
	})
}

// ===========================================
// Lake-base Agent Maintenance Handlers
// ===========================================

// agentService holds the singleton agent service
var agentService *agent.AgentService

// InitAgentService initializes the agent service
func InitAgentService(pool *lakebase.ConnectionPool, config *agent.AgentConfig) {
	agentService = agent.NewAgentService(pool, config)
}

// GetAgentService returns the agent service
func GetAgentService() *agent.AgentService {
	return agentService
}

// GetAgentStatus returns the status of the agent maintenance service
// GET /api/v1/agent/status
func (h *Handler) GetAgentStatus(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Agent service not initialized",
			"running": false,
		})
		return
	}

	status := agentService.GetMaintenanceStatus()
	c.JSON(http.StatusOK, status)
}

// StartAgentService starts the background maintenance loop
// POST /api/v1/agent/start
func (h *Handler) StartAgentService(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	if err := agentService.Start(); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent service started",
		"running": true,
	})
}

// StopAgentService stops the background maintenance loop
// POST /api/v1/agent/stop
func (h *Handler) StopAgentService(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	agentService.Stop()

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent service stopped",
		"running": false,
	})
}

// RunAgentMaintenance triggers maintenance for a specific datasource
// POST /api/v1/agent/maintenance/:datasource_id
func (h *Handler) RunAgentMaintenance(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsIDStr := c.Param("datasource_id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	result, err := agentService.RunMaintenance(reqCtx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": result,
		})
		return
	}

	// Auto-regenerate embeddings if context was updated
	var embeddingsUpdated int
	if result.ContextRefreshed > 0 || result.ContextCreated > 0 {
		if h.lakebaseService != nil {
			embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(reqCtx, dsID)
			if embErr == nil && embResult != nil {
				embeddingsUpdated = embResult.TotalEmbeddings
				result.EmbeddingsUpdated = embeddingsUpdated
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Maintenance completed",
		"result":  result,
	})
}

// TriggerContextRefresh triggers context refresh for a datasource
// POST /api/v1/agent/refresh/:datasource_id
func (h *Handler) TriggerContextRefresh(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsIDStr := c.Param("datasource_id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	results, err := agentService.TriggerContextRefresh(reqCtx, dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	// Auto-regenerate embeddings after context refresh
	var embeddingsUpdated int
	if successCount > 0 && h.lakebaseService != nil {
		embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(reqCtx, dsID)
		if embErr == nil && embResult != nil {
			embeddingsUpdated = embResult.TotalEmbeddings
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Context refresh completed",
		"total":              len(results),
		"success_count":      successCount,
		"embeddings_updated": embeddingsUpdated,
		"results":            results,
	})
}

// SimulateDDLChange simulates a DDL change for demo purposes
// POST /api/v1/agent/simulate-ddl/:datasource_id
func (h *Handler) SimulateDDLChange(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsIDStr := c.Param("datasource_id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	var req struct {
		SQL string `json:"sql" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Parse and process the DDL
	result, err := agentService.ProcessDDLStatement(reqCtx, dsID, req.SQL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": result,
		})
		return
	}

	// Get the parsed change for response
	change := agent.ParseDDLStatement(req.SQL)

	// Auto-regenerate embeddings after DDL change processing
	var embeddingsUpdated int
	if (result.ContextRefreshed > 0 || result.ContextCreated > 0) && h.lakebaseService != nil {
		embResult, embErr := h.lakebaseService.GenerateAndSaveEmbeddings(reqCtx, dsID)
		if embErr == nil && embResult != nil {
			embeddingsUpdated = embResult.TotalEmbeddings
			result.EmbeddingsUpdated = embeddingsUpdated
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "DDL change processed",
		"sql":           req.SQL,
		"parsed_change": change,
		"result":        result,
	})
}

// GetAgentChangeLogs returns change logs for a datasource
// GET /api/v1/agent/logs/:datasource_id
func (h *Handler) GetAgentChangeLogs(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsIDStr := c.Param("datasource_id")
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

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	logs, err := agentService.GetRecentChanges(reqCtx, dsID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"datasource_id": dsID,
		"logs":          logs,
		"count":         len(logs),
	})
}

// GetAgentChangeLogSummary returns a summary of change logs
// GET /api/v1/agent/logs/:datasource_id/summary
func (h *Handler) GetAgentChangeLogSummary(c *gin.Context) {
	if agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsIDStr := c.Param("datasource_id")
	dsID, err := strconv.ParseInt(dsIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid datasource ID",
		})
		return
	}

	reqCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	summary, err := agentService.GetChangeLogSummary(reqCtx, dsID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}
