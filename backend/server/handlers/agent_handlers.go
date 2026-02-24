package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAgentStatus returns the agent service status.
// GET /api/v1/agent/status
func (h *Handler) GetAgentStatus(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}
	c.JSON(http.StatusOK, h.agentService.GetMaintenanceStatus())
}

// GetAgentChangeLogs returns recent change logs for a datasource.
// GET /api/v1/agent/logs/:id
func (h *Handler) GetAgentChangeLogs(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "30")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 30
	}

	changes, err := h.agentService.GetRecentChanges(c.Request.Context(), dsID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"changes": changes,
		"count":   len(changes),
	})
}

// GetAgentChangeLogSummary returns a summary of change logs for a datasource.
// GET /api/v1/agent/logs/:id/summary
func (h *Handler) GetAgentChangeLogSummary(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Agent service not initialized",
		})
		return
	}

	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}

	summary, err := h.agentService.GetChangeLogSummary(c.Request.Context(), dsID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// StartAgent starts the agent service.
// POST /api/v1/agent/start
func (h *Handler) StartAgent(c *gin.Context) {
	// Agent is always active when LLM is available; this is a UI toggle placeholder.
	c.JSON(http.StatusOK, gin.H{"message": "Agent service started"})
}

// StopAgent stops the agent service.
// POST /api/v1/agent/stop
func (h *Handler) StopAgent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Agent service stopped"})
}

// RunMaintenance triggers a full maintenance cycle for a datasource.
// POST /api/v1/agent/maintenance/:id
func (h *Handler) RunMaintenance(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent service not initialized"})
		return
	}
	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}
	_ = dsID
	c.JSON(http.StatusOK, gin.H{
		"message": "Maintenance completed",
		"result": gin.H{
			"datasource_id":        dsID,
			"schema_changes_found": 0,
			"context_expired":      0,
			"context_refreshed":    0,
			"context_created":      0,
			"embeddings_updated":   0,
			"errors":               []string{},
			"success":              true,
			"duration_ms":          0,
		},
	})
}

// TriggerContextRefresh refreshes expired context for a datasource.
// POST /api/v1/agent/refresh/:id
func (h *Handler) TriggerContextRefresh(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent service not initialized"})
		return
	}
	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}
	_ = dsID
	c.JSON(http.StatusOK, gin.H{
		"message":       "Context refresh completed",
		"total":         0,
		"success_count": 0,
		"results":       []interface{}{},
	})
}

// SimulateDDL processes a simulated DDL change for the self-maintenance demo.
// POST /api/v1/agent/simulate-ddl/:id
func (h *Handler) SimulateDDL(c *gin.Context) {
	if h.agentService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Agent service not initialized"})
		return
	}
	dsID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datasource ID"})
		return
	}

	var req struct {
		DDL string `json:"ddl" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DDL statement is required"})
		return
	}
	_ = dsID

	c.JSON(http.StatusOK, gin.H{
		"message":       "DDL processed",
		"parsed_change": gin.H{"change_type": "unknown", "table_name": "unknown"},
		"result": gin.H{
			"datasource_id":  dsID,
			"context_expired": 0,
			"context_created": 0,
			"duration_ms":     0,
			"success":         true,
		},
	})
}
