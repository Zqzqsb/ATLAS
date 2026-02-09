package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lucid/internal/grounding"
)

// GroundRequest represents a grounding request
type GroundRequest struct {
	Query        string `json:"query" binding:"required"`
	DatasourceID int64  `json:"datasource_id"`
	Mode         string `json:"mode"` // "coarse_only", "sequential", "parallel"
}

// GroundResponse represents a grounding response
type GroundResponse struct {
	Success bool                      `json:"success"`
	Context *grounding.GroundedContext `json:"context,omitempty"`
	Tables  []string                  `json:"tables,omitempty"`
	Signals int                       `json:"signals_probed,omitempty"`
	Mode    string                    `json:"mode,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

// Ground performs semantic grounding for a query
// POST /api/v1/grounding/ground
func (h *Handler) Ground(c *gin.Context) {
	if h.groundingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Grounding service not available"})
		return
	}

	var req GroundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GroundResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if req.DatasourceID > 0 {
		h.groundingService.SetDatasourceID(req.DatasourceID)
	}

	mode := grounding.ModeSequential
	switch req.Mode {
	case "coarse_only":
		mode = grounding.ModeCoarseOnly
	case "parallel":
		mode = grounding.ModeParallel
	}

	result, err := h.groundingService.Ground(c.Request.Context(), req.Query, mode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GroundResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GroundResponse{
		Success: true,
		Context: result.Context,
		Tables:  h.groundingService.GetSelectedTables(result.Context),
		Signals: len(result.CoarseSignals),
		Mode:    result.Mode,
	})
}

// GroundStream performs grounding with SSE streaming
// GET /api/v1/grounding/stream?query=...&datasource_id=...
func (h *Handler) GroundStream(c *gin.Context) {
	if h.groundingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Grounding service not available"})
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	datasourceID, _ := strconv.ParseInt(c.Query("datasource_id"), 10, 64)
	if datasourceID > 0 {
		h.groundingService.SetDatasourceID(datasourceID)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	resultCh, err := h.groundingService.GroundWithStreaming(c.Request.Context(), query)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	for result := range resultCh {
		eventData := gin.H{
			"mode":           result.Mode,
			"signals_probed": len(result.CoarseSignals),
			"tables":         h.groundingService.GetSelectedTables(result.Context),
			"duration_ms":    result.TotalDuration.Milliseconds(),
		}

		if result.Mode == "parallel_fine" {
			eventData["context"] = result.Context
			c.SSEvent("complete", eventData)
		} else {
			c.SSEvent("progress", eventData)
		}
		c.Writer.Flush()
	}
}

// GetGroundingConfig returns the current grounding configuration
// GET /api/v1/grounding/config
func (h *Handler) GetGroundingConfig(c *gin.Context) {
	if h.groundingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Grounding service not available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  h.groundingService.GetConfig(),
	})
}

// UpdateGroundingConfig updates the grounding configuration
// PUT /api/v1/grounding/config
func (h *Handler) UpdateGroundingConfig(c *gin.Context) {
	if h.groundingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Grounding service not available"})
		return
	}

	var config grounding.GroundingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.groundingService.UpdateConfig(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated",
		"config":  h.groundingService.GetConfig(),
	})
}

// FormatGroundingPrompt formats grounded context as a prompt
// POST /api/v1/grounding/format
func (h *Handler) FormatGroundingPrompt(c *gin.Context) {
	if h.groundingService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Grounding service not available"})
		return
	}

	var ctx grounding.GroundedContext
	if err := c.ShouldBindJSON(&ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	prompt := h.groundingService.FormatContextPrompt(&ctx)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"prompt":  prompt,
	})
}
