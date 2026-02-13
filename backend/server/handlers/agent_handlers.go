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

// RunAgentMaintenance triggers a manual maintenance run for a datasource.
// POST /api/v1/agent/maintenance/:id
func (h *Handler) RunAgentMaintenance(c *gin.Context) {
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

	result, err := h.agentService.RunMaintenance(c.Request.Context(), dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TriggerContextRefresh triggers a context refresh for a datasource.
// POST /api/v1/agent/refresh/:id
func (h *Handler) TriggerContextRefresh(c *gin.Context) {
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

	results, err := h.agentService.TriggerContextRefresh(c.Request.Context(), dsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refreshed": len(results),
		"results":   results,
	})
}
