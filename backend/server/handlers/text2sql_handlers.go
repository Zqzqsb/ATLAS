package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/grounding"
	"lucid/server/services"
)

// Text2SQLRequest represents the input for text2sql conversion.
type Text2SQLRequest struct {
	Question         string          `json:"question" binding:"required"`
	DatabaseID       string          `json:"database_id" binding:"required"`
	Database         string          `json:"database"`
	Options          Text2SQLOptions `json:"options"`
	FieldDescription string          `json:"field_description"`
}

// Text2SQLOptions holds optional parameters.
type Text2SQLOptions struct {
	UseRichContext bool `json:"use_rich_context"`
	UseReact       bool `json:"use_react"`
	UseGrounding   bool `json:"use_grounding"`
	MaxIterations  int  `json:"max_iterations"`
	Stream         bool `json:"stream"`
}

// ReactStep represents a single step in ReAct reasoning.
type ReactStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Phase       string      `json:"phase"`
}

// GroundingInfo represents grounding result in response.
type GroundingInfo struct {
	Tables          []GroundedTableInfo  `json:"tables"`
	Columns         []GroundedColumnInfo `json:"columns"`
	JoinPaths       []JoinPathInfo       `json:"join_paths,omitempty"`
	ExecutionTimeMs int64                `json:"execution_time_ms"`
	ExecutionLogs   []ExecutionLogInfo   `json:"execution_logs,omitempty"`
	Reasoning       string               `json:"reasoning,omitempty"`
	Mode            string               `json:"mode,omitempty"`
}

// ExecutionLogInfo represents SQL execution log for frontend transparency.
type ExecutionLogInfo struct {
	Phase       string `json:"phase"`
	SQL         string `json:"sql"`
	ResultCount int    `json:"result_count"`
	DurationMs  int64  `json:"duration_ms"`
	Summary     string `json:"summary"`
}

// GroundedTableInfo represents a grounded table in response.
type GroundedTableInfo struct {
	Name       string  `json:"name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GroundedColumnInfo represents a grounded column in response.
type GroundedColumnInfo struct {
	TableName  string  `json:"table_name"`
	ColumnName string  `json:"column_name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// JoinPathInfo represents a join path in response.
type JoinPathInfo struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	Reason     string `json:"reason,omitempty"`
}

// Text2SQLResponse represents the output.
type Text2SQLResponse struct {
	SQL             string      `json:"sql"`
	ExecutionResult interface{} `json:"execution_result,omitempty"`
	Metadata        struct {
		SelectedTables     []string       `json:"selected_tables"`
		Iterations         int            `json:"iterations"`
		ReactTrace         []ReactStep    `json:"react_trace"`
		RichContextUpdated bool           `json:"rich_context_updated"`
		ExecutionTimeMs    int64          `json:"execution_time_ms"`
		GroundingResult    *GroundingInfo `json:"grounding_result,omitempty"`
	} `json:"metadata"`
}

// Text2SQL handles synchronous text2sql conversion.
func (h *Handler) Text2SQL(c *gin.Context) {
	var req Text2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Options.MaxIterations == 0 {
		req.Options.MaxIterations = h.config.React.MaxIterations
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	groundingInfo := h.performGrounding(ctx, &req)

	inferReq := &services.Text2SQLRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.Options.UseRichContext,
		UseReact:         req.Options.UseReact,
		MaxIterations:    req.Options.MaxIterations,
		FieldDescription: req.FieldDescription,
	}

	result, err := h.inferenceService.Execute(ctx, inferReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := Text2SQLResponse{
		SQL:             result.SQL,
		ExecutionResult: result.ExecutionResult,
	}
	response.Metadata.SelectedTables = result.Metadata.SelectedTables
	response.Metadata.Iterations = result.Metadata.Iterations
	response.Metadata.ReactTrace = convertReactSteps(result.Metadata.ReactTrace)
	response.Metadata.RichContextUpdated = result.Metadata.RichContextUpdated
	response.Metadata.ExecutionTimeMs = result.Metadata.ExecutionTimeMs
	response.Metadata.GroundingResult = groundingInfo

	c.JSON(http.StatusOK, response)
}

// Text2SQLStream handles streaming text2sql conversion with SSE.
func (h *Handler) Text2SQLStream(c *gin.Context) {
	var req Text2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	if req.Options.MaxIterations == 0 {
		req.Options.MaxIterations = h.config.React.MaxIterations
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Perform grounding with SSE progress
	var groundingInfo *GroundingInfo
	if req.Options.UseGrounding && h.groundingService != nil {
		SendSSE(c.Writer, "grounding_start", map[string]string{"message": "Starting semantic grounding..."})
		flusher.Flush()

		if h.lakebaseService != nil {
			datasources, err := h.lakebaseService.ListDatasources(ctx)
			if err == nil && len(datasources) > 0 {
				h.groundingService.SetDatasourceID(datasources[0].ID)
			}
		}

		SendSSE(c.Writer, "grounding_progress", map[string]string{"stage": "analyzing", "message": "Analyzing query and schema..."})
		flusher.Flush()

		result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
		if err != nil {
			fmt.Printf("Grounding failed (continuing without): %v\n", err)
			SendSSE(c.Writer, "grounding_error", map[string]string{"error": err.Error()})
			flusher.Flush()
		} else {
			groundingInfo = convertGroundingResult(result)
			SendSSE(c.Writer, "grounding_complete", groundingInfo)
			flusher.Flush()
		}
	}

	events := make(chan services.StreamEvent, 100)

	go func() {
		defer close(events)

		inferReq := &services.Text2SQLRequest{
			Question:         req.Question,
			DatabaseID:       req.DatabaseID,
			Database:         req.Database,
			UseRichContext:   req.Options.UseRichContext,
			UseReact:         req.Options.UseReact,
			MaxIterations:    req.Options.MaxIterations,
			FieldDescription: req.FieldDescription,
		}

		if err := h.inferenceService.ExecuteStream(ctx, inferReq, events); err != nil {
			events <- services.StreamEvent{
				Type:      services.EventError,
				Data:      services.ErrorEventData{Error: err.Error()},
				Timestamp: time.Now().UnixMilli(),
			}
		}
	}()

	for event := range events {
		select {
		case <-ctx.Done():
			return
		default:
		}
		SendSSE(c.Writer, string(event.Type), event.Data)
		flusher.Flush()
	}
}

// SuggestFields analyzes the question and suggests output fields.
func (h *Handler) SuggestFields(c *gin.Context) {
	var req struct {
		Question   string `json:"question" binding:"required"`
		DatabaseID string `json:"database_id" binding:"required"`
		Database   string `json:"database"`
		Language   string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	result, err := h.inferenceService.SuggestFields(ctx, req.Question, req.DatabaseID, req.Database, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// --- Helper functions ---

// performGrounding runs semantic grounding if enabled.
func (h *Handler) performGrounding(ctx context.Context, req *Text2SQLRequest) *GroundingInfo {
	if !req.Options.UseGrounding || h.groundingService == nil {
		return nil
	}

	if h.lakebaseService != nil {
		datasources, err := h.lakebaseService.ListDatasources(ctx)
		if err == nil && len(datasources) > 0 {
			h.groundingService.SetDatasourceID(datasources[0].ID)
		}
	}

	result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
	if err != nil {
		fmt.Printf("Grounding failed (continuing without): %v\n", err)
		return nil
	}
	return convertGroundingResult(result)
}

func convertGroundingResult(result *grounding.GroundingResult) *GroundingInfo {
	if result == nil || result.Context == nil {
		return nil
	}

	info := &GroundingInfo{
		ExecutionTimeMs: result.TotalDuration.Milliseconds(),
		Mode:            result.Mode,
	}

	for _, t := range result.Context.Tables {
		info.Tables = append(info.Tables, GroundedTableInfo{
			Name:       t.Name,
			Reason:     t.Reason,
			Confidence: float64(t.Relevance),
		})
	}
	for _, col := range result.Context.Columns {
		info.Columns = append(info.Columns, GroundedColumnInfo{
			TableName:  col.TableName,
			ColumnName: col.ColumnName,
			Reason:     col.Reason,
			Confidence: float64(col.Relevance),
		})
	}
	for _, rel := range result.Context.Relationships {
		info.JoinPaths = append(info.JoinPaths, JoinPathInfo{
			FromTable:  rel.FromTable,
			FromColumn: rel.FromColumn,
			ToTable:    rel.ToTable,
			ToColumn:   rel.ToColumn,
			Reason:     rel.Type,
		})
	}
	for _, log := range result.ExecutionLogs {
		info.ExecutionLogs = append(info.ExecutionLogs, ExecutionLogInfo{
			Phase:       log.Phase,
			SQL:         log.SQL,
			ResultCount: log.ResultCount,
			DurationMs:  log.Duration.Milliseconds(),
			Summary:     log.Summary,
		})
	}
	if result.Context.Reasoning != "" {
		info.Reasoning = result.Context.Reasoning
	}
	return info
}


func convertReactSteps(steps []services.ReActStep) []ReactStep {
	result := make([]ReactStep, len(steps))
	for i, s := range steps {
		result[i] = ReactStep{
			Step:        i + 1,
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionInput,
			Observation: s.Observation,
			Phase:       s.Phase,
		}
	}
	return result
}

