package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
