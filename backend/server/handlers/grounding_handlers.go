package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lucid/internal/grounding"
)

// GroundingHandlers handles semantic grounding API requests
type GroundingHandlers struct {
	service *grounding.Service
}

// NewGroundingHandlers creates new grounding handlers
func NewGroundingHandlers(service *grounding.Service) *GroundingHandlers {
	return &GroundingHandlers{service: service}
}

// GroundRequest represents a grounding request
type GroundRequest struct {
	Query        string `json:"query" binding:"required"`
	DatasourceID int64  `json:"datasource_id"`
	Mode         string `json:"mode"` // "coarse_only", "sequential", "parallel"
}

// GroundResponse represents a grounding response
type GroundResponse struct {
	Success bool                     `json:"success"`
	Context *grounding.GroundedContext `json:"context,omitempty"`
	Tables  []string                 `json:"tables,omitempty"`
	Signals int                      `json:"signals_probed,omitempty"`
	Mode    string                   `json:"mode,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

// Ground performs semantic grounding for a query
// POST /api/v1/grounding/ground
func (h *GroundingHandlers) Ground(c *gin.Context) {
	var req GroundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GroundResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Set datasource ID if provided
	if req.DatasourceID > 0 {
		h.service.SetDatasourceID(req.DatasourceID)
	}

	// Determine grounding mode
	mode := grounding.ModeSequential
	switch req.Mode {
	case "coarse_only":
		mode = grounding.ModeCoarseOnly
	case "parallel":
		mode = grounding.ModeParallel
	}

	// Perform grounding
	result, err := h.service.Ground(c.Request.Context(), req.Query, mode)
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
		Tables:  h.service.GetSelectedTables(result.Context),
		Signals: len(result.CoarseSignals),
		Mode:    result.Mode,
	})
}

// GroundStream performs grounding with SSE streaming
// GET /api/v1/grounding/stream?query=...&datasource_id=...
func (h *GroundingHandlers) GroundStream(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	datasourceID, _ := strconv.ParseInt(c.Query("datasource_id"), 10, 64)
	if datasourceID > 0 {
		h.service.SetDatasourceID(datasourceID)
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Get streaming channel
	resultCh, err := h.service.GroundWithStreaming(c.Request.Context(), query)
	if err != nil {
		c.SSEvent("error", gin.H{"error": err.Error()})
		return
	}

	// Stream results
	for result := range resultCh {
		eventData := gin.H{
			"mode":           result.Mode,
			"signals_probed": len(result.CoarseSignals),
			"tables":         h.service.GetSelectedTables(result.Context),
			"duration_ms":    result.TotalDuration.Milliseconds(),
		}

		if result.Mode == "parallel_fine" {
			// Final result with full context
			eventData["context"] = result.Context
			c.SSEvent("complete", eventData)
		} else {
			// Intermediate result
			c.SSEvent("progress", eventData)
		}
		c.Writer.Flush()
	}
}

// GetConfig returns the current grounding configuration
// GET /api/v1/grounding/config
func (h *GroundingHandlers) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  h.service.GetConfig(),
	})
}

// UpdateConfig updates the grounding configuration
// PUT /api/v1/grounding/config
func (h *GroundingHandlers) UpdateConfig(c *gin.Context) {
	var config grounding.GroundingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	h.service.UpdateConfig(&config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated",
		"config":  h.service.GetConfig(),
	})
}

// FormatPrompt formats grounded context as a prompt
// POST /api/v1/grounding/format
func (h *GroundingHandlers) FormatPrompt(c *gin.Context) {
	var ctx grounding.GroundedContext
	if err := c.ShouldBindJSON(&ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	prompt := h.service.FormatContextPrompt(&ctx)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"prompt":  prompt,
	})
}
