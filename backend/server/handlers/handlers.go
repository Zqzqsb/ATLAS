package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/config"
	"lucid/internal/grounding"
	"lucid/internal/lakebase"
	"lucid/server/services"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	config           *config.Config
	dbService        *services.DatabaseService
	inferenceService *services.InferenceService
	lakebaseService  *services.LakebaseService
	groundingService *grounding.Service // Optional: semantic grounding service
}

// HandlerDependencies holds all dependencies needed to create handlers
type HandlerDependencies struct {
	Config           *config.Config
	DBService        *services.DatabaseService
	InferenceService *services.InferenceService
	LakebaseService  *services.LakebaseService // Optional: lake-base storage service
	GroundingService *grounding.Service        // Optional: semantic grounding service
}

// New creates a new Handler instance from dependencies
func New(deps *HandlerDependencies) (*Handler, error) {
	if deps.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if deps.DBService == nil {
		return nil, fmt.Errorf("database service is required")
	}
	if deps.InferenceService == nil {
		return nil, fmt.Errorf("inference service is required")
	}

	return &Handler{
		config:           deps.Config,
		dbService:        deps.DBService,
		inferenceService: deps.InferenceService,
		lakebaseService:  deps.LakebaseService,
		groundingService: deps.GroundingService,
	}, nil
}

// Close cleans up resources
func (h *Handler) Close() {
	if h.dbService != nil {
		h.dbService.Close()
	}
	if h.lakebaseService != nil {
		h.lakebaseService.Close()
	}
}

// ============================================
// System Handlers
// ============================================

// GetSystemInfo returns system information
func (h *Handler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"llm": gin.H{
			"default_model":    h.config.LLM.DefaultModel,
			"available_models": h.inferenceService.GetAvailableModels(),
		},
		"react": gin.H{
			"max_iterations": h.config.React.MaxIterations,
		},
	})
}

// GetModels returns list of available models
func (h *Handler) GetModels(c *gin.Context) {
	models := h.inferenceService.GetAvailableModels()
	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// SwitchModel switches the current LLM model
func (h *Handler) SwitchModel(c *gin.Context) {
	var req struct {
		ModelID string `json:"model_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.inferenceService.SwitchModel(req.ModelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model switched successfully",
		"model":   req.ModelID,
	})
}

// ============================================
// Database Handlers
// ============================================

// DatabaseInfo represents database information for API response
type DatabaseInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ListDatabases returns all configured databases
func (h *Handler) ListDatabases(c *gin.Context) {
	databases := make([]DatabaseInfo, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
		databases = append(databases, DatabaseInfo{
			ID:   db.ID,
			Name: db.Name,
			Type: db.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"databases": databases,
	})
}

// GetDatabaseSchema returns schema information for a database
func (h *Handler) GetDatabaseSchema(c *gin.Context) {
	dbID := c.Param("id")
	database := c.Query("database") // Optional: specific Spider database

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// If specific database requested, use inference service
	if database != "" {
		// Get adapter for specific database
		adp, err := h.inferenceService.CreateSpiderAdapter(dbID, database)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer adp.Close()

		if err := adp.Connect(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to connect: " + err.Error(),
			})
			return
		}

		// Build schema info manually
		c.JSON(http.StatusOK, gin.H{
			"database_id":   dbID,
			"database":      database,
			"database_type": adp.GetDatabaseType(),
			"message":       "Use /api/v1/spider/databases/:name/schema for full schema",
		})
		return
	}

	schema, err := h.dbService.GetSchema(ctx, dbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// GetDatabaseTables returns list of tables for a database
func (h *Handler) GetDatabaseTables(c *gin.Context) {
	dbID := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	schema, err := h.dbService.GetSchema(ctx, dbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	tables := make([]string, len(schema.Tables))
	for i, t := range schema.Tables {
		tables[i] = t.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"database_id": dbID,
		"tables":      tables,
	})
}

// GetRichContext returns rich context for a database
func (h *Handler) GetRichContext(c *gin.Context) {
	dbID := c.Param("id")
	database := c.Query("database")

	if database == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "database parameter required",
		})
		return
	}

	info, err := h.inferenceService.GetRichContext(dbID, database)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ============================================
// Text2SQL Handlers
// ============================================

// Text2SQLRequest represents the input for text2sql conversion
type Text2SQLRequest struct {
	Question         string          `json:"question" binding:"required"`
	DatabaseID       string          `json:"database_id" binding:"required"`
	Database         string          `json:"database"` // Specific database name (for Spider)
	Options          Text2SQLOptions `json:"options"`
	FieldDescription string          `json:"field_description"` // Optional field clarification description
}

// Text2SQLOptions holds optional parameters
type Text2SQLOptions struct {
	UseRichContext bool `json:"use_rich_context"`
	UseReact       bool `json:"use_react"`
	UseGrounding   bool `json:"use_grounding"` // Use semantic grounding for schema linking
	MaxIterations  int  `json:"max_iterations"`
	Stream         bool `json:"stream"`
}

// ReactStep represents a single step in ReAct reasoning
type ReactStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Phase       string      `json:"phase"` // "schema_linking" or "sql_generation"
}

// GroundingInfo represents grounding result in response
type GroundingInfo struct {
	Tables          []GroundedTableInfo  `json:"tables"`
	Columns         []GroundedColumnInfo `json:"columns"`
	JoinPaths       []JoinPathInfo       `json:"join_paths,omitempty"`
	ExecutionTimeMs int64                `json:"execution_time_ms"`
	ExecutionLogs   []ExecutionLogInfo   `json:"execution_logs,omitempty"` // SQL execution transparency
	Reasoning       string               `json:"reasoning,omitempty"`       // LLM reasoning for fine selection
	Mode            string               `json:"mode,omitempty"`            // "sequential", "parallel", "coarse_only"
}

// ExecutionLogInfo represents SQL execution log for frontend transparency
type ExecutionLogInfo struct {
	Phase       string `json:"phase"`        // "vector_search", "fine_selection"
	SQL         string `json:"sql"`          // SQL query executed
	ResultCount int    `json:"result_count"` // Number of results
	DurationMs  int64  `json:"duration_ms"`  // Execution time in milliseconds
	Summary     string `json:"summary"`      // Human-readable summary
}

// GroundedTableInfo represents a grounded table in response
type GroundedTableInfo struct {
	Name       string  `json:"name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GroundedColumnInfo represents a grounded column in response
type GroundedColumnInfo struct {
	TableName  string  `json:"table_name"`
	ColumnName string  `json:"column_name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// JoinPathInfo represents a join path in response
type JoinPathInfo struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	Reason     string `json:"reason,omitempty"`
}

// Text2SQLResponse represents the output
type Text2SQLResponse struct {
	SQL             string      `json:"sql"`
	ExecutionResult interface{} `json:"execution_result,omitempty"`
	Metadata        struct {
		SelectedTables     []string       `json:"selected_tables"`
		Iterations         int            `json:"iterations"`
		ReactTrace         []ReactStep    `json:"react_trace"`
		RichContextUpdated bool           `json:"rich_context_updated"`
		ExecutionTimeMs    int64          `json:"execution_time_ms"`
		GroundingResult    *GroundingInfo `json:"grounding_result,omitempty"` // Semantic grounding result
	} `json:"metadata"`
}

// Text2SQL handles synchronous text2sql conversion
func (h *Handler) Text2SQL(c *gin.Context) {
	var req Text2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set defaults
	if req.Options.MaxIterations == 0 {
		req.Options.MaxIterations = h.config.React.MaxIterations
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	// Perform semantic grounding if enabled and service is available
	var groundingInfo *GroundingInfo
	var groundingResult *services.GroundingResult
	if req.Options.UseGrounding && h.groundingService != nil {
		// Set datasource ID if lakebase service is available
		if h.lakebaseService != nil {
			// Try to get datasource ID from database name
			datasources, err := h.lakebaseService.ListDatasources(ctx)
			if err == nil && len(datasources) > 0 {
				// Use first available datasource
				h.groundingService.SetDatasourceID(datasources[0].ID)
			}
		}

		result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
		if err != nil {
			// Log but don't fail - grounding is optional enhancement
			fmt.Printf("Grounding failed (continuing without): %v\n", err)
		} else {
			// Convert grounding result to handler format
			groundingInfo = convertGroundingResult(result)
			// Also convert to interfaces format for inference engine
			groundingResult = convertToInterfaceGrounding(result)
		}
	}

	// Build inference request
	inferReq := &services.Text2SQLRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.Options.UseRichContext,
		UseReact:         req.Options.UseReact,
		MaxIterations:    req.Options.MaxIterations,
		FieldDescription: req.FieldDescription,
	}

	// Inject grounding result if available
	if groundingResult != nil {
		inferReq.GroundingResult = groundingResult
	}

	// Execute inference
	result, err := h.inferenceService.Execute(ctx, inferReq)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Build response
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

// convertGroundingResult converts grounding.GroundingResult to GroundingInfo
func convertGroundingResult(result *grounding.GroundingResult) *GroundingInfo {
	if result == nil || result.Context == nil {
		return nil
	}

	info := &GroundingInfo{
		ExecutionTimeMs: result.TotalDuration.Milliseconds(),
		Mode:            result.Mode,
	}

	// Convert tables
	for _, t := range result.Context.Tables {
		info.Tables = append(info.Tables, GroundedTableInfo{
			Name:       t.Name,
			Reason:     t.Reason,
			Confidence: float64(t.Relevance),
		})
	}

	// Convert columns
	for _, col := range result.Context.Columns {
		info.Columns = append(info.Columns, GroundedColumnInfo{
			TableName:  col.TableName,
			ColumnName: col.ColumnName,
			Reason:     col.Reason,
			Confidence: float64(col.Relevance),
		})
	}

	// Convert relationships as join paths
	for _, rel := range result.Context.Relationships {
		info.JoinPaths = append(info.JoinPaths, JoinPathInfo{
			FromTable:  rel.FromTable,
			FromColumn: rel.FromColumn,
			ToTable:    rel.ToTable,
			ToColumn:   rel.ToColumn,
			Reason:     rel.Type,
		})
	}

	// Convert execution logs for transparency
	for _, log := range result.ExecutionLogs {
		info.ExecutionLogs = append(info.ExecutionLogs, ExecutionLogInfo{
			Phase:       log.Phase,
			SQL:         log.SQL,
			ResultCount: log.ResultCount,
			DurationMs:  log.Duration.Milliseconds(),
			Summary:     log.Summary,
		})
	}

	// Add LLM reasoning if available
	if result.Context.Reasoning != "" {
		info.Reasoning = result.Context.Reasoning
	}

	return info
}

// convertToInterfaceGrounding converts grounding result to interfaces format
func convertToInterfaceGrounding(result *grounding.GroundingResult) *services.GroundingResult {
	if result == nil || result.Context == nil {
		return nil
	}

	gr := &services.GroundingResult{
		ExecutionTimeMs: result.TotalDuration.Milliseconds(),
	}

	for _, t := range result.Context.Tables {
		gr.Tables = append(gr.Tables, services.GroundedTable{
			Name:       t.Name,
			Reason:     t.Reason,
			Confidence: float64(t.Relevance),
		})
	}

	for _, col := range result.Context.Columns {
		gr.Columns = append(gr.Columns, services.GroundedColumn{
			TableName:  col.TableName,
			ColumnName: col.ColumnName,
			Reason:     col.Reason,
			Confidence: float64(col.Relevance),
		})
	}

	for _, rel := range result.Context.Relationships {
		gr.JoinPaths = append(gr.JoinPaths, services.JoinPath{
			FromTable:  rel.FromTable,
			FromColumn: rel.FromColumn,
			ToTable:    rel.ToTable,
			ToColumn:   rel.ToColumn,
			Reason:     rel.Type,
		})
	}

	return gr
}

// convertReactSteps converts service ReactSteps to handler ReactSteps
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

// Text2SQLStream handles streaming text2sql conversion with SSE (TRUE STREAMING)
func (h *Handler) Text2SQLStream(c *gin.Context) {
	var req Text2SQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Set defaults
	if req.Options.MaxIterations == 0 {
		req.Options.MaxIterations = h.config.React.MaxIterations
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 180*time.Second)
	defer cancel()

	// Get flusher early
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "streaming not supported",
		})
		return
	}

	// Perform semantic grounding if enabled and service is available
	var groundingInfo *GroundingInfo
	var groundingResult *services.GroundingResult
	if req.Options.UseGrounding && h.groundingService != nil {
		// Send grounding_start event
		sendSSEEvent(c.Writer, "grounding_start", map[string]string{
			"message": "Starting semantic grounding...",
		})
		flusher.Flush()

		// Set datasource ID if lakebase service is available
		if h.lakebaseService != nil {
			datasources, err := h.lakebaseService.ListDatasources(ctx)
			if err == nil && len(datasources) > 0 {
				h.groundingService.SetDatasourceID(datasources[0].ID)
			}
		}

		// Send grounding_progress event
		sendSSEEvent(c.Writer, "grounding_progress", map[string]string{
			"stage":   "analyzing",
			"message": "Analyzing query and schema...",
		})
		flusher.Flush()

		result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
		if err != nil {
			// Log but don't fail - grounding is optional enhancement
			fmt.Printf("Grounding failed (continuing without): %v\n", err)
			sendSSEEvent(c.Writer, "grounding_error", map[string]string{
				"error": err.Error(),
			})
			flusher.Flush()
		} else {
			// Convert grounding result to handler format
			groundingInfo = convertGroundingResult(result)
			// Also convert to interfaces format for inference engine
			groundingResult = convertToInterfaceGrounding(result)

			// Send grounding_complete event with results
			sendSSEEvent(c.Writer, "grounding_complete", groundingInfo)
			flusher.Flush()
		}
	}

	// Create event channel for streaming
	events := make(chan services.StreamEvent, 100)

	// Start streaming inference in goroutine
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

		// Inject grounding result if available
		if groundingResult != nil {
			inferReq.GroundingResult = groundingResult
		}

		err := h.inferenceService.ExecuteStream(ctx, inferReq, events)

		if err != nil {
			// Send error event
			events <- services.StreamEvent{
				Type:      services.EventError,
				Data:      services.ErrorEventData{Error: err.Error()},
				Timestamp: time.Now().UnixMilli(),
			}
		}
	}()

	// Process events from channel
	for event := range events {
		// Check if client disconnected
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Send SSE event
		sendSSEEvent(c.Writer, string(event.Type), event.Data)
		flusher.Flush()
	}
}

// sendSSEEvent sends a Server-Sent Event
func sendSSEEvent(w http.ResponseWriter, eventType string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}

// ExecuteSQL executes SQL on a database
func (h *Handler) ExecuteSQL(c *gin.Context) {
	dbID := c.Param("id")

	var req struct {
		SQL      string `json:"sql" binding:"required"`
		Database string `json:"database"` // Optional: specific Spider database
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var result interface{}
	var err error

	if req.Database != "" {
		// Execute on specific Spider database
		adp, adpErr := h.inferenceService.CreateSpiderAdapter(dbID, req.Database)
		if adpErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": adpErr.Error(),
			})
			return
		}
		defer adp.Close()

		if err := adp.Connect(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to connect: " + err.Error(),
			})
			return
		}

		result, err = adp.ExecuteQuery(ctx, req.SQL)
	} else {
		result, err = h.dbService.ExecuteSQL(ctx, dbID, req.SQL)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database_id": dbID,
		"database":    req.Database,
		"sql":         req.SQL,
		"result":      result,
	})
}

// ============================================
// Spider Dataset Handlers
// ============================================

// ListSpiderDatabases lists all available Spider databases
func (h *Handler) ListSpiderDatabases(c *gin.Context) {
	dbID := c.DefaultQuery("source", "spider_sqlite")

	databases, err := h.inferenceService.ListSpiderDatabases(dbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source":    dbID,
		"count":     len(databases),
		"databases": databases,
	})
}

// GetSpiderQuestions returns questions for a Spider database
func (h *Handler) GetSpiderQuestions(c *gin.Context) {
	database := c.Param("database")

	questions, err := h.inferenceService.LoadSpiderQuestions(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database":  database,
		"count":     len(questions),
		"questions": questions,
	})
}

// ============================================
// Demo Handlers
// ============================================

// DemoScenario represents a demo scenario
type DemoScenario struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Questions   []string `json:"questions"`
	DatabaseID  string   `json:"database_id"`
	Database    string   `json:"database"` // Specific Spider database
	Highlights  []string `json:"highlights"`
}

// ListDemoScenarios returns predefined demo scenarios
func (h *Handler) ListDemoScenarios(c *gin.Context) {
	scenarios := []DemoScenario{
		{
			ID:          "schema_linking",
			Name:        "Schema Linking Demo",
			Description: "Demonstrates automatic table and column selection",
			Questions: []string{
				"What are the names of all cars made by Toyota?",
				"Show me the average horsepower by country",
			},
			DatabaseID: "spider_sqlite",
			Database:   "car_1",
			Highlights: []string{"Schema Linking", "Table Selection"},
		},
		{
			ID:          "rich_context",
			Name:        "Rich Context Demo",
			Description: "Demonstrates Auto-Renewable Rich Context mechanism",
			Questions: []string{
				"How many concerts are there?",
				"Which stadium has the most concerts?",
			},
			DatabaseID: "spider_sqlite",
			Database:   "concert_singer",
			Highlights: []string{"Rich Context", "Auto-Renewal", "[EXPIRED] Detection"},
		},
		{
			ID:          "react_loop",
			Name:        "ReAct Loop Demo",
			Description: "Demonstrates the ReAct reasoning loop with tool usage",
			Questions: []string{
				"Find the singer who has performed in the most concerts",
			},
			DatabaseID: "spider_sqlite",
			Database:   "concert_singer",
			Highlights: []string{"ReAct Loop", "Self-Correction", "Tool Usage"},
		},
		{
			ID:          "multi_db",
			Name:        "Multi-Database Demo",
			Description: "Same query executed on SQLite, MySQL, and PostgreSQL",
			Questions: []string{
				"List all tables in the database",
			},
			DatabaseID: "spider_sqlite",
			Database:   "concert_singer",
			Highlights: []string{"Database Abstraction", "Cross-DB Compatibility"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"scenarios": scenarios,
	})
}

// GetDemoScenario returns a specific demo scenario
func (h *Handler) GetDemoScenario(c *gin.Context) {
	scenarioID := c.Param("id")

	scenarios := map[string]DemoScenario{
		"schema_linking": {
			ID:          "schema_linking",
			Name:        "Schema Linking Demo",
			Description: "Demonstrates automatic table and column selection based on natural language question",
			Questions: []string{
				"What are the names of all cars made by Toyota?",
				"Show me the average horsepower by country",
				"Find the car with the highest MPG",
			},
			DatabaseID: "spider_sqlite",
			Database:   "car_1",
			Highlights: []string{"Schema Linking", "Table Selection", "Column Mapping"},
		},
		"rich_context": {
			ID:          "rich_context",
			Name:        "Rich Context Demo",
			Description: "Demonstrates the Auto-Renewable Rich Context mechanism that enhances SQL generation with business context",
			Questions: []string{
				"How many concerts are there?",
				"Which stadium has the most concerts?",
				"List singers who performed at more than 2 concerts",
			},
			DatabaseID: "spider_sqlite",
			Database:   "concert_singer",
			Highlights: []string{"Rich Context", "Business Rules", "Auto-Renewal", "[EXPIRED] Detection"},
		},
		"react_loop": {
			ID:          "react_loop",
			Name:        "ReAct Loop Demo",
			Description: "Demonstrates the ReAct (Reasoning + Acting) loop with iterative refinement",
			Questions: []string{
				"Find the singer who has performed in the most concerts",
				"What is the total capacity of stadiums that hosted concerts in 2020?",
			},
			DatabaseID: "spider_sqlite",
			Database:   "concert_singer",
			Highlights: []string{"ReAct Loop", "Thought Process", "Self-Correction", "Tool Usage"},
		},
	}

	scenario, exists := scenarios[scenarioID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "scenario not found",
		})
		return
	}

	c.JSON(http.StatusOK, scenario)
}

// ============================================
// Database Connection Management Handlers
// ============================================

// ConnectionConfig represents database connection configuration
type ConnectionConfig struct {
	ID       string `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"` // mysql, postgresql, sqlite
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Path     string `json:"path"` // For SQLite
}

// ConnectionStatus represents connection test result
type ConnectionStatus struct {
	ID        string `json:"id"`
	Connected bool   `json:"connected"`
	Message   string `json:"message"`
	Version   string `json:"version,omitempty"`
	Latency   int64  `json:"latency_ms"`
}

// AddConnection adds a new database connection
func (h *Handler) AddConnection(c *gin.Context) {
	var conn ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate connection type
	if conn.Type != "mysql" && conn.Type != "mariadb" && conn.Type != "postgresql" && conn.Type != "sqlite" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database type. Must be mysql, mariadb, postgresql, or sqlite",
		})
		return
	}

	// Check if connection ID already exists
	for _, db := range h.config.Databases {
		if db.ID == conn.ID {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Connection ID already exists",
			})
			return
		}
	}

	// Normalize type: "mariadb" -> "mysql" for adapter
	adapterType := conn.Type
	if adapterType == "mariadb" {
		adapterType = "mysql"
	}

	// Step 1: Add to config so GetAdapter can find it
	newDB := config.DatabaseConfig{
		ID:       conn.ID,
		Name:     conn.Name,
		Type:     adapterType,
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
		Path:     conn.Path,
	}
	h.config.Databases = append(h.config.Databases, newDB)

	// Step 2: Actually connect — this is synchronous, must succeed
	if _, err := h.dbService.GetAdapter(conn.ID); err != nil {
		// Rollback: remove from config since connection failed
		h.config.Databases = h.config.Databases[:len(h.config.Databases)-1]
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to connect to database: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Connection added successfully",
		"connection": conn,
	})
}

// RemoveConnection removes a database connection
func (h *Handler) RemoveConnection(c *gin.Context) {
	connID := c.Param("id")

	// Find and remove connection
	found := false
	newDatabases := make([]config.DatabaseConfig, 0, len(h.config.Databases))
	for _, db := range h.config.Databases {
		if db.ID == connID {
			found = true
			// Close any existing adapter
			h.dbService.CloseAdapter(connID)
		} else {
			newDatabases = append(newDatabases, db)
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Connection not found",
		})
		return
	}

	h.config.Databases = newDatabases

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection removed successfully",
		"id":      connID,
	})
}

// SyncConnectionSchema creates/updates an rc_datasources record for a connection and syncs its physical schema.
// POST /connections/:id/sync-schema
// This is the bridge between connection management and RC — called explicitly by the frontend, not implicitly.
func (h *Handler) SyncConnectionSchema(c *gin.Context) {
	connID := c.Param("id")

	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Lake-base service not available"})
		return
	}

	// Find connection config
	var dbCfg *config.DatabaseConfig
	for _, db := range h.config.Databases {
		if db.ID == connID {
			dbCfg = &db
			break
		}
	}
	if dbCfg == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Get or create rc_datasources record
	ds, err := h.lakebaseService.GetOrCreateDatasource(ctx, &lakebase.Datasource{
		Name:         dbCfg.ID,
		DBType:       dbCfg.Type,
		Host:         sql.NullString{String: dbCfg.Host, Valid: dbCfg.Host != ""},
		Port:         sql.NullInt32{Int32: int32(dbCfg.Port), Valid: dbCfg.Port > 0},
		Username:     sql.NullString{String: dbCfg.User, Valid: dbCfg.User != ""},
		DatabaseName: sql.NullString{String: dbCfg.Database, Valid: dbCfg.Database != ""},
		Status:       "active",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register datasource: %v", err)})
		return
	}

	// Get adapter and sync schema
	adapter, err := h.dbService.GetAdapter(connID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cannot connect to database: %v", err)})
		return
	}

	result, err := h.lakebaseService.SyncSchema(ctx, ds.ID, adapter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Schema sync failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"datasource_id": ds.ID,
		"tables":        result.TablesCount,
		"columns":       result.ColumnsCount,
		"relations":     result.RelationsCount,
	})
}

// ReleaseAllDemoConnections releases all demo database connections
func (h *Handler) ReleaseAllDemoConnections(c *gin.Context) {
	// Get list of demo database IDs from config file
	demoConfigPath := os.Getenv("DEMO_DATABASES_PATH")
	if demoConfigPath == "" {
		demoConfigPath = "demo_databases.json"
	}

	var demoIDs []string

	// Try to load demo config to get IDs
	data, err := os.ReadFile(demoConfigPath)
	if err == nil {
		var demoConfig struct {
			Databases []struct {
				ID          string                       `json:"id"`
				Connections map[string]map[string]string `json:"connections"`
			} `json:"databases"`
		}
		if json.Unmarshal(data, &demoConfig) == nil {
			for _, db := range demoConfig.Databases {
				for connType := range db.Connections {
					demoIDs = append(demoIDs, fmt.Sprintf("%s_%s", db.ID, connType))
				}
			}
		}
	}

	// If no demo config, release all connections with known demo patterns
	if len(demoIDs) == 0 {
		// Fallback: identify demo connections by known patterns
		for _, db := range h.config.Databases {
			// Demo connections typically have patterns like xxx_mysql, xxx_sqlite, xxx_postgres
			if strings.HasSuffix(db.ID, "_mysql") ||
				strings.HasSuffix(db.ID, "_sqlite") ||
				strings.HasSuffix(db.ID, "_postgres") ||
				strings.HasSuffix(db.ID, "_postgresql") {
				demoIDs = append(demoIDs, db.ID)
			}
		}
	}

	// Remove all demo connections
	releasedCount := 0
	releasedIDs := []string{}
	newDatabases := make([]config.DatabaseConfig, 0, len(h.config.Databases))

	for _, db := range h.config.Databases {
		isDemo := false
		for _, demoID := range demoIDs {
			if db.ID == demoID {
				isDemo = true
				break
			}
		}

		if isDemo {
			// Close the adapter
			h.dbService.CloseAdapter(db.ID)
			releasedCount++
			releasedIDs = append(releasedIDs, db.ID)
		} else {
			newDatabases = append(newDatabases, db)
		}
	}

	h.config.Databases = newDatabases

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Released %d demo database connections", releasedCount),
		"released": releasedCount,
		"ids":      releasedIDs,
	})
}

// TestConnection tests a database connection
func (h *Handler) TestConnection(c *gin.Context) {
	var conn ConnectionConfig
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	startTime := time.Now()

	// Create adapter based on type
	status := ConnectionStatus{
		ID: conn.ID,
	}

	adapterConfig := &services.AdapterConfig{
		Type:     conn.Type,
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
		Path:     conn.Path,
	}

	adapter, err := h.dbService.CreateCustomAdapter(adapterConfig)
	if err != nil {
		status.Connected = false
		status.Message = fmt.Sprintf("Failed to create adapter: %v", err)
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}
	defer adapter.Close()

	// Test connection
	if err := adapter.Connect(ctx); err != nil {
		status.Connected = false
		status.Message = fmt.Sprintf("Failed to connect: %v", err)
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}

	// Get version
	version, err := adapter.GetDatabaseVersion(ctx)
	if err != nil {
		status.Connected = true
		status.Message = "Connected successfully (version query failed)"
		status.Latency = time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, status)
		return
	}

	status.Connected = true
	status.Message = "Connected successfully"
	status.Version = version
	status.Latency = time.Since(startTime).Milliseconds()

	c.JSON(http.StatusOK, status)
}

// ListConnections returns all configured connections with status
func (h *Handler) ListConnections(c *gin.Context) {
	connections := make([]map[string]interface{}, 0, len(h.config.Databases))

	for _, db := range h.config.Databases {
		conn := map[string]interface{}{
			"id":       db.ID,
			"name":     db.Name,
			"type":     db.Type,
			"host":     db.Host,
			"port":     db.Port,
			"database": db.Database,
			"user":     db.User,
		}
		// Don't expose password
		if db.Type == "sqlite" {
			conn["path"] = db.Path
		}
		connections = append(connections, conn)
	}

	c.JSON(http.StatusOK, gin.H{
		"connections": connections,
		"count":       len(connections),
	})
}

// ListAvailableConnections returns available demo databases that haven't been connected yet
func (h *Handler) ListAvailableConnections(c *gin.Context) {
	demoConfig := h.inferenceService.GetDemoDatabases()
	if demoConfig == nil {
		c.JSON(http.StatusOK, gin.H{
			"available": []interface{}{},
			"count":     0,
			"message":   "Demo database configuration not found",
		})
		return
	}

	// Build a set of existing connection IDs
	existingIDs := make(map[string]bool)
	for _, db := range h.config.Databases {
		existingIDs[db.ID] = true
	}

	// External port mapping for display (what users see as localhost:15001)
	displayPortMapping := map[string]int{
		"car_1_mysql":                         15001,
		"car_1_postgresql":                    15002,
		"flight_2_mysql":                      15011,
		"flight_2_postgresql":                 15012,
		"college_2_mysql":                     15021,
		"college_2_postgresql":                15022,
		"employee_hire_evaluation_mysql":      15031,
		"employee_hire_evaluation_postgresql": 15032,
	}

	// Internal Docker network container names (backend connects to these directly)
	internalHostMapping := map[string]string{
		"car_1_mysql":                         "lucid_demo_car_1_mysql",
		"car_1_postgresql":                    "lucid_demo_car_1_postgres",
		"flight_2_mysql":                      "lucid_demo_flight_2_mysql",
		"flight_2_postgresql":                 "lucid_demo_flight_2_postgres",
		"college_2_mysql":                     "lucid_demo_college_2_mysql",
		"college_2_postgresql":                "lucid_demo_college_2_postgres",
		"employee_hire_evaluation_mysql":      "lucid_demo_employee_mysql",
		"employee_hire_evaluation_postgresql": "lucid_demo_employee_postgres",
	}

	// Internal Docker ports (MySQL=3306, PostgreSQL=5432)
	internalPortMapping := map[string]int{
		"car_1_mysql":                         3306,
		"car_1_postgresql":                    5432,
		"flight_2_mysql":                      3306,
		"flight_2_postgresql":                 5432,
		"college_2_mysql":                     3306,
		"college_2_postgresql":                5432,
		"employee_hire_evaluation_mysql":      3306,
		"employee_hire_evaluation_postgresql": 5432,
	}

	// Build list of available connections
	available := make([]map[string]interface{}, 0)
	for _, db := range demoConfig.Databases {
		for connType, conn := range db.Connections {
			connID := fmt.Sprintf("%s_%s", db.ID, connType)

			// Skip if already connected
			if existingIDs[connID] {
				continue
			}

			// Get display port for UI (external port like 15001)
			displayPort := conn.Port
			if port, ok := displayPortMapping[connID]; ok {
				displayPort = port
			}

			// Get internal Docker host (container name) for actual connection
			internalHost := conn.Host
			if host, ok := internalHostMapping[connID]; ok {
				internalHost = host
			}

			// Get internal port (3306 for MySQL, 5432 for PostgreSQL)
			internalPort := conn.Port
			if port, ok := internalPortMapping[connID]; ok {
				internalPort = port
			}

			// Backend connects via Docker internal network (container_name:internal_port)
			// Frontend displays localhost:external_port
			available = append(available, map[string]interface{}{
				"id":               connID,
				"name":             fmt.Sprintf("%s (%s)", db.Name, strings.ToUpper(connType)),
				"description":      db.Description,
				"domain":           db.Domain,
				"data_scale":       db.DataScale,
				"type":             conn.Type,
				"host":             internalHost, // Docker container name for backend connection
				"port":             internalPort, // Internal port (3306/5432)
				"display_host":     "localhost",  // For UI display only
				"display_port":     displayPort,  // External port for UI display (15001, etc.)
				"database":         conn.Database,
				"user":             conn.User,
				"password":         conn.Password,
				"sample_questions": db.SampleQuestions,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"available": available,
		"count":     len(available),
	})
}

// LoadDemoDatabases loads database connections from demo_databases.json
func (h *Handler) LoadDemoDatabases(c *gin.Context) {
	// Get demo database config
	demoConfig := h.inferenceService.GetDemoDatabases()
	if demoConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load demo databases: config file not found",
		})
		return
	}

	// Find the base directory for SQLite databases
	// This is needed because the paths in demo_databases.json are relative
	sqliteBaseDirs := []string{
		"/app",                         // Container path
		"/root/workspace/lucid/system", // Dev path (system dir)
		"/root/workspace/lucid",        // Project root
		".",                            // Current dir
	}

	// Find which base directory contains the SQLite files
	var sqliteBaseDir string
	for _, baseDir := range sqliteBaseDirs {
		testPath := filepath.Join(baseDir, "data/spider_data/database/car_1/car_1.sqlite")
		if _, err := os.Stat(testPath); err == nil {
			sqliteBaseDir = baseDir
			break
		}
	}

	// Add connections from demo config
	addedCount := 0
	updatedCount := 0
	for _, db := range demoConfig.Databases {
		for connType, conn := range db.Connections {
			connID := fmt.Sprintf("%s_%s", db.ID, connType)

			// Handle SQLite path resolution
			dbPath := conn.Path
			if conn.Type == "sqlite" && dbPath != "" && sqliteBaseDir != "" {
				// Convert relative path to absolute path
				if strings.HasPrefix(dbPath, "./") {
					dbPath = filepath.Join(sqliteBaseDir, dbPath[2:])
				} else if !filepath.IsAbs(dbPath) {
					dbPath = filepath.Join(sqliteBaseDir, dbPath)
				}
			}

			// Check if already exists
			existsIdx := -1
			for i, existing := range h.config.Databases {
				if existing.ID == connID {
					existsIdx = i
					break
				}
			}

			newDB := config.DatabaseConfig{
				ID:       connID,
				Name:     fmt.Sprintf("%s (%s)", db.Name, strings.ToUpper(connType)),
				Type:     conn.Type,
				Host:     conn.Host,
				Port:     conn.Port,
				User:     conn.User,
				Password: conn.Password,
				Database: conn.Database,
				Path:     dbPath,
			}

			if existsIdx >= 0 {
				// Update existing connection (in case path or other config changed)
				h.config.Databases[existsIdx] = newDB
				updatedCount++
			} else {
				// Add new connection
				h.config.Databases = append(h.config.Databases, newDB)
				addedCount++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Loaded %d new, updated %d database connections", addedCount, updatedCount),
		"added":   addedCount,
		"updated": updatedCount,
	})
}

// ============================================
// Translation Handler
// ============================================

// TranslateRequest represents the input for translation
type TranslateRequest struct {
	Texts      []string `json:"texts" binding:"required"`
	TargetLang string   `json:"target_lang"` // Default: "Chinese"
}

// TranslateResponse represents the translation result
type TranslateResponse struct {
	Translations map[string]string `json:"translations"`
}

// TranslateTexts translates texts using LLM
func (h *Handler) TranslateTexts(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Default target language
	if req.TargetLang == "" {
		req.TargetLang = "Chinese"
	}

	// Limit the number of texts to translate at once
	if len(req.Texts) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Too many texts to translate (max 50)",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Call inference service to translate
	translations, err := h.inferenceService.TranslateTexts(ctx, req.Texts, req.TargetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TranslateResponse{
		Translations: translations,
	})
}

// ============================================
// Field Suggestion Handler (for field clarification)
// ============================================

// SuggestFieldsRequest represents the input for field suggestion
type SuggestFieldsRequest struct {
	Question   string `json:"question" binding:"required"`
	DatabaseID string `json:"database_id" binding:"required"`
	Database   string `json:"database"` // Specific database name (for Spider)
	Language   string `json:"language"` // Language for suggestions: "Chinese" or "English"
}

// SuggestFieldsResponse represents the suggested fields
type SuggestFieldsResponse struct {
	SuggestedFields []SuggestedField `json:"suggested_fields"`
	AnalysisNote    string           `json:"analysis_note"`
}

// SuggestedField represents a single suggested output field
type SuggestedField struct {
	Name        string `json:"name"`        // Field name (e.g., "FullName", "count")
	Description string `json:"description"` // Description of what this field represents
	Selected    bool   `json:"selected"`    // Whether this field is selected by default
	Source      string `json:"source"`      // Source table or expression
}

// SuggestFields analyzes the question and suggests output fields
func (h *Handler) SuggestFields(c *gin.Context) {
	var req SuggestFieldsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Call inference service to suggest fields
	result, err := h.inferenceService.SuggestFields(ctx, req.Question, req.DatabaseID, req.Database, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
