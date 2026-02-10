package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"lucid/internal/grounding"
	"lucid/internal/lakebase"
	"lucid/server/services"
)

// Text2SQLRequest represents the input for text2sql conversion.
type Text2SQLRequest struct {
	Question           string          `json:"question" binding:"required"`
	DatabaseID         string          `json:"database_id" binding:"required"`
	Database           string          `json:"database"`
	Options            Text2SQLOptions `json:"options"`
	FieldDescription   string          `json:"field_description"`
	InjectedGrounding  *GroundingInfo  `json:"injected_grounding,omitempty"` // Reuse previous grounding result
}

// Text2SQLOptions holds optional parameters.
type Text2SQLOptions struct {
	UseRichContext bool `json:"use_rich_context"`
	UseReact       bool `json:"use_react"`
	UseGrounding   bool `json:"use_grounding"`
	MaxIterations  int  `json:"max_iterations"`
	Stream         bool `json:"stream"`
	GroundingOnly  bool `json:"grounding_only"` // When true, stop after grounding (for field alignment)
	SkipGrounding  bool `json:"skip_grounding"` // When true, skip grounding and use InjectedGrounding
}

// Text2SQLRequest represents the input for text2sql conversion.
// InjectedGrounding allows the frontend to pass back previously-obtained grounding
// results, so that Phase 2 (inference) can reuse them without re-running grounding.

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
	SuggestedFields []SuggestedFieldInfo `json:"suggested_fields,omitempty"`
	ExecutionTimeMs int64                `json:"execution_time_ms"`
	ExecutionLogs   []ExecutionLogInfo   `json:"execution_logs,omitempty"`
	Reasoning       string               `json:"reasoning,omitempty"`
	Mode            string               `json:"mode,omitempty"`

	// richContextPrompt is a pre-formatted prompt containing full grounding context
	// (tables, columns, business rules, domain terms, SQL templates).
	// Not serialized to JSON — used internally by extractLinkedContext.
	richContextPrompt string
}

// SuggestedFieldInfo represents a field suggested by the linking agent.
// These are the columns the agent identified as directly relevant to the query.
type SuggestedFieldInfo struct {
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name"`
	Reason     string `json:"reason"`
	Selected   bool   `json:"selected"`
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

	var groundingInfo *GroundingInfo
	if req.Options.SkipGrounding && req.InjectedGrounding != nil {
		groundingInfo = req.InjectedGrounding
		// Regenerate richContextPrompt since it's not serialized in JSON
		if h.groundingService != nil {
			groundingInfo.richContextPrompt = h.regenerateRichContextPrompt(ctx, groundingInfo)
		}
	} else {
		groundingInfo = h.performGrounding(ctx, &req)
	}

	inferReq := &services.Text2SQLRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.Options.UseRichContext,
		UseReact:         req.Options.UseReact,
		MaxIterations:    req.Options.MaxIterations,
		FieldDescription: req.FieldDescription,
	}

	// Inject grounding result into inference to skip redundant Schema Linking
	if groundingInfo != nil {
		inferReq.LinkedTables, inferReq.LinkedContextPrompt = extractLinkedContext(groundingInfo, req.FieldDescription)
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

	// Perform grounding with SSE progress (or reuse injected grounding from Phase 1)
	var groundingInfo *GroundingInfo
	if req.Options.SkipGrounding && req.InjectedGrounding != nil {
		// Phase 2: Reuse grounding from Phase 1 — skip grounding entirely
		groundingInfo = req.InjectedGrounding
		// Regenerate richContextPrompt since it's not serialized in JSON
		if h.groundingService != nil {
			groundingInfo.richContextPrompt = h.regenerateRichContextPrompt(ctx, groundingInfo)
		}
		SendSSE(c.Writer, "grounding_complete", groundingInfo)
		flusher.Flush()
	} else if req.Options.UseGrounding && h.groundingService != nil {
		SendSSE(c.Writer, "grounding_start", map[string]string{"message": "Starting semantic grounding..."})
		flusher.Flush()

		var datasourceID int64
		if h.lakebaseService != nil {
			datasources, err := h.lakebaseService.ListDatasources(ctx)
			if err == nil && len(datasources) > 0 {
				datasourceID = datasources[0].ID
				h.groundingService.SetDatasourceID(datasourceID)
			}
		}

		SendSSE(c.Writer, "grounding_progress", map[string]string{"stage": "analyzing", "message": "Analyzing query and schema..."})
		flusher.Flush()

		// Try adaptive grounding first
		var groundingErr error
		if h.groundingService.IsAdaptiveAvailable() && h.lakebaseService != nil && datasourceID > 0 {
			schemas, err := h.loadSchemasForGrounding(ctx, datasourceID)
			if err == nil && len(schemas) > 0 {
				adaptiveReq := &grounding.AdaptiveGroundingRequest{
					Query:        req.Question,
					DatasourceID: datasourceID,
					AllSchemas:   schemas,
					TableCount:   len(schemas),
					ProgressCallback: func(stage string, data map[string]interface{}) {
						// Forward grounding sub-stage progress via SSE
						SendSSE(c.Writer, "grounding_progress", map[string]interface{}{
							"stage": stage,
							"data":  data,
						})
						flusher.Flush()
					},
				}
				result, err := h.groundingService.GroundAdaptive(ctx, adaptiveReq)
				if err == nil {
					legacyResult := h.groundingService.ConvertAdaptiveResult(result)
					groundingInfo = h.convertGroundingResultRich(legacyResult)
				} else {
					fmt.Printf("Adaptive grounding failed, falling back to legacy: %v\n", err)
					groundingErr = err
				}
			}
		}

		// Fallback to legacy grounding
		if groundingInfo == nil {
			result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
			if err != nil {
				groundingErr = err
			} else {
				groundingInfo = h.convertGroundingResultRich(result)
			}
		}

		if groundingErr != nil && groundingInfo == nil {
			fmt.Printf("Grounding failed (continuing without): %v\n", groundingErr)
			SendSSE(c.Writer, "grounding_error", map[string]string{"error": groundingErr.Error()})
			flusher.Flush()
		} else if groundingInfo != nil {
			SendSSE(c.Writer, "grounding_complete", groundingInfo)
			flusher.Flush()
		}
	}

	// If grounding_only mode, stop here — frontend will show field panel and re-call with fieldDescription
	if req.Options.GroundingOnly {
		SendSSE(c.Writer, "complete", map[string]interface{}{
			"grounding_only": true,
			"message":        "Grounding complete. Awaiting field confirmation.",
		})
		flusher.Flush()
		return
	}

	// Extract linked context from grounding for injection into inference
	var linkedTables []string
	var linkedContextPrompt string
	if groundingInfo != nil {
		linkedTables, linkedContextPrompt = extractLinkedContext(groundingInfo, req.FieldDescription)
	}

	events := make(chan services.StreamEvent, 100)

	go func() {
		defer close(events)

		inferReq := &services.Text2SQLRequest{
			Question:            req.Question,
			DatabaseID:          req.DatabaseID,
			Database:            req.Database,
			UseRichContext:      req.Options.UseRichContext,
			UseReact:            req.Options.UseReact,
			MaxIterations:       req.Options.MaxIterations,
			FieldDescription:    req.FieldDescription,
			LinkedTables:        linkedTables,
			LinkedContextPrompt: linkedContextPrompt,
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

// --- Schema cache for grounding ---

type groundingSchemaCacheEntry struct {
	schemas   []grounding.SchemaInfo
	expiresAt time.Time
}

var (
	groundingSchemaCache sync.Map // map[int64]*groundingSchemaCacheEntry
	groundingSchemaTTL   = 5 * time.Minute
)

// Warmup pre-loads schema data into caches for faster query execution.
// This endpoint can be called by the frontend when the user selects a database
// or starts typing a question, so that the actual query doesn't pay the cold-start cost.
func (h *Handler) Warmup(c *gin.Context) {
	var req struct {
		DatabaseID string `json:"database_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	warmedUp := make(map[string]bool)

	// 1. Warm up inference engine schema cache
	if h.inferenceService != nil && req.DatabaseID != "" {
		// Trigger schema load which populates the cache
		if engine, ok := h.inferenceService.GetEngine().(*services.InferenceEngine); ok {
			engine.WarmupSchema(ctx, req.DatabaseID)
			warmedUp["inference_schema"] = true
		}
	}

	// 2. Warm up grounding schema cache
	if h.lakebaseService != nil && h.groundingService != nil {
		var datasourceID int64
		datasources, err := h.lakebaseService.ListDatasources(ctx)
		if err == nil && len(datasources) > 0 {
			datasourceID = datasources[0].ID
		}
		if datasourceID > 0 {
			_, err := h.loadSchemasForGrounding(ctx, datasourceID)
			if err == nil {
				warmedUp["grounding_schema"] = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"warmed_up": warmedUp,
	})
}

// --- Helper functions ---

// performGrounding runs semantic grounding if enabled.
// Prefers adaptive grounding (with full schema from lakebase) when available,
// falls back to legacy vector-only grounding otherwise.
func (h *Handler) performGrounding(ctx context.Context, req *Text2SQLRequest) *GroundingInfo {
	if !req.Options.UseGrounding || h.groundingService == nil {
		return nil
	}

	// Resolve datasource
	var datasourceID int64
	if h.lakebaseService != nil {
		datasources, err := h.lakebaseService.ListDatasources(ctx)
		if err == nil && len(datasources) > 0 {
			datasourceID = datasources[0].ID
			h.groundingService.SetDatasourceID(datasourceID)
		}
	}

	// Try adaptive grounding first (full schema → linking agent)
	if h.groundingService.IsAdaptiveAvailable() && h.lakebaseService != nil && datasourceID > 0 {
		schemas, err := h.loadSchemasForGrounding(ctx, datasourceID)
		if err == nil && len(schemas) > 0 {
			adaptiveReq := &grounding.AdaptiveGroundingRequest{
				Query:        req.Question,
				DatasourceID: datasourceID,
				AllSchemas:   schemas,
				TableCount:   len(schemas),
			}
			result, err := h.groundingService.GroundAdaptive(ctx, adaptiveReq)
			if err == nil {
				legacyResult := h.groundingService.ConvertAdaptiveResult(result)
				return h.convertGroundingResultRich(legacyResult)
			}
			fmt.Printf("Adaptive grounding failed, falling back to legacy: %v\n", err)
		}
	}

	// Fallback: legacy vector-only grounding
	result, err := h.groundingService.Ground(ctx, req.Question, grounding.ModeParallel)
	if err != nil {
		fmt.Printf("Grounding failed (continuing without): %v\n", err)
		return nil
	}
	return h.convertGroundingResultRich(result)
}

// loadSchemasForGrounding loads all table schemas from lakebase for the adaptive grounding pipeline.
// Uses an in-memory cache with TTL to avoid repeated DB queries across requests.
func (h *Handler) loadSchemasForGrounding(ctx context.Context, datasourceID int64) ([]grounding.SchemaInfo, error) {
	// Check cache first
	if cached, ok := groundingSchemaCache.Load(datasourceID); ok {
		entry := cached.(*groundingSchemaCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.schemas, nil
		}
		groundingSchemaCache.Delete(datasourceID)
	}

	tables, err := h.lakebaseService.GetTablesByDatasource(ctx, datasourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tables: %w", err)
	}

	columns, err := h.lakebaseService.GetColumnsByDatasource(ctx, datasourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load columns: %w", err)
	}

	// Group columns by table
	columnsByTable := make(map[string][]*lakebase.ColumnInfo)
	for _, col := range columns {
		columnsByTable[col.TableName] = append(columnsByTable[col.TableName], col)
	}

	// Also load foreign key relations
	relations, _ := h.lakebaseService.GetRelationsByDatasource(ctx, datasourceID)
	fkByTable := make(map[string][]grounding.FKInfo)
	for _, rel := range relations {
		fkByTable[rel.FromTable] = append(fkByTable[rel.FromTable], grounding.FKInfo{
			Column:           rel.FromColumn,
			ReferencedTable:  rel.ToTable,
			ReferencedColumn: rel.ToColumn,
		})
	}

	var schemas []grounding.SchemaInfo
	for _, t := range tables {
		schema := grounding.SchemaInfo{
			TableName: t.TableName,
			RowCount:  t.RowCount,
		}
		if t.Description.Valid {
			schema.Description = t.Description.String
		}

		// Add columns
		if cols, ok := columnsByTable[t.TableName]; ok {
			for _, col := range cols {
				ci := grounding.ColumnInfo{
					Name:         col.ColumnName,
					IsPrimaryKey: col.IsPrimaryKey,
					IsNullable:   col.IsNullable,
				}
				if col.DataType.Valid {
					ci.Type = col.DataType.String
				}
				if col.Description.Valid {
					ci.Description = col.Description.String
				}
				if col.SampleValues.Valid {
					ci.SampleValues = col.SampleValues.String
				}
				if col.Synonyms.Valid {
					ci.Synonyms = col.Synonyms.String
				}
				schema.Columns = append(schema.Columns, ci)
			}
		}

		// Add foreign keys
		if fks, ok := fkByTable[t.TableName]; ok {
			schema.ForeignKeys = fks
		}

		schemas = append(schemas, schema)
	}

	// Store in cache
	groundingSchemaCache.Store(datasourceID, &groundingSchemaCacheEntry{
		schemas:   schemas,
		expiresAt: time.Now().Add(groundingSchemaTTL),
	})

	return schemas, nil
}

// InvalidateGroundingSchemaCache clears the grounding schema cache.
// Should be called when schema changes are detected (e.g., after sync).
func InvalidateGroundingSchemaCache(datasourceID int64) {
	if datasourceID == 0 {
		groundingSchemaCache = sync.Map{}
		return
	}
	groundingSchemaCache.Delete(datasourceID)
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
		// Columns with a Reason from the linking agent are "suggested fields"
		if col.Reason != "" {
			info.SuggestedFields = append(info.SuggestedFields, SuggestedFieldInfo{
				TableName:  col.TableName,
				ColumnName: col.ColumnName,
				Reason:     col.Reason,
				Selected:   true,
			})
		}
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

// convertGroundingResultRich converts GroundingResult and also generates a rich context prompt
// using grounding.Service.FormatContextPrompt for injection into inference.
func (h *Handler) convertGroundingResultRich(result *grounding.GroundingResult) *GroundingInfo {
	info := convertGroundingResult(result)
	if info == nil {
		return nil
	}

	// Generate rich context prompt (includes business rules, domain terms, SQL templates)
	if h.groundingService != nil && result.Context != nil {
		info.richContextPrompt = h.groundingService.FormatContextPrompt(result.Context)
	}
	return info
}

// regenerateRichContextPrompt rebuilds the richContextPrompt from GroundingInfo tables/columns.
// This is needed when GroundingInfo is deserialized from JSON (Phase 2 skip_grounding),
// because richContextPrompt is an unexported field that isn't serialized.
func (h *Handler) regenerateRichContextPrompt(ctx context.Context, info *GroundingInfo) string {
	if info == nil || len(info.Tables) == 0 {
		return ""
	}

	// Try to load full schema from lakebase and build the prompt
	if h.lakebaseService != nil {
		var datasourceID int64
		datasources, err := h.lakebaseService.ListDatasources(ctx)
		if err == nil && len(datasources) > 0 {
			datasourceID = datasources[0].ID
		}
		if datasourceID > 0 {
			schemas, err := h.loadSchemasForGrounding(ctx, datasourceID)
			if err == nil && len(schemas) > 0 {
				// Build a GroundedContext from the info + schemas for FormatContextPrompt
				selectedTableNames := make(map[string]bool)
				for _, t := range info.Tables {
					selectedTableNames[t.Name] = true
				}

				gc := &grounding.GroundedContext{
					Tables:  make([]grounding.TableContext, 0),
					Columns: make([]grounding.ColumnContext, 0),
				}
				for _, schema := range schemas {
					if !selectedTableNames[schema.TableName] {
						continue
					}
					tc := grounding.TableContext{
						Name:        schema.TableName,
						Description: schema.Description,
					}
					colNames := make([]string, len(schema.Columns))
					for i, col := range schema.Columns {
						colNames[i] = col.Name
						gc.Columns = append(gc.Columns, grounding.ColumnContext{
							TableName:   schema.TableName,
							ColumnName:  col.Name,
							DataType:    col.Type,
							Description: col.Description,
						})
					}
					tc.Columns = colNames
					gc.Tables = append(gc.Tables, tc)
				}

				return h.groundingService.FormatContextPrompt(gc)
			}
		}
	}

	return ""
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

// extractLinkedContext converts GroundingInfo into linked tables and a context prompt
// for injection into the inference pipeline, eliminating redundant Schema Linking.
//
// If fieldDescription is provided (from user's field alignment confirmation), it takes priority.
// Otherwise, if SuggestedFields are available from the linking agent, they are injected.
// If a rich context prompt was pre-generated by FormatContextPrompt, it is appended.
func extractLinkedContext(info *GroundingInfo, fieldDescription string) ([]string, string) {
	if info == nil || len(info.Tables) == 0 {
		return nil, ""
	}

	tables := make([]string, 0, len(info.Tables))
	for _, t := range info.Tables {
		tables = append(tables, t.Name)
	}

	var sb strings.Builder

	// Append rich context prompt if available (includes full schema + business rules)
	if info.richContextPrompt != "" {
		sb.WriteString(info.richContextPrompt)
	} else {
		// Fallback: build a basic context prompt from GroundingInfo fields
		sb.WriteString("=== Grounding Context ===\n\n")

		for _, t := range info.Tables {
			sb.WriteString(fmt.Sprintf("Table: %s", t.Name))
			if t.Reason != "" {
				sb.WriteString(fmt.Sprintf(" (reason: %s)", t.Reason))
			}
			sb.WriteString("\n")
		}

		if len(info.Columns) > 0 {
			sb.WriteString("\nRelevant Columns:\n")
			for _, c := range info.Columns {
				sb.WriteString(fmt.Sprintf("  - %s.%s", c.TableName, c.ColumnName))
				if c.Reason != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", c.Reason))
				}
				sb.WriteString("\n")
			}
		}

		if len(info.JoinPaths) > 0 {
			sb.WriteString("\nJoin Paths:\n")
			for _, jp := range info.JoinPaths {
				sb.WriteString(fmt.Sprintf("  - %s.%s → %s.%s\n", jp.FromTable, jp.FromColumn, jp.ToTable, jp.ToColumn))
			}
		}

		if info.Reasoning != "" {
			sb.WriteString(fmt.Sprintf("\nReasoning: %s\n", info.Reasoning))
		}
	}

	// Inject field alignment hints
	if fieldDescription != "" {
		// User-provided field description takes priority
		sb.WriteString("\n=== Output Field Constraints ===\n")
		sb.WriteString(fmt.Sprintf("The user has specified these output fields: %s\n", fieldDescription))
		sb.WriteString("Ensure the SELECT clause includes these fields.\n")
	} else if len(info.SuggestedFields) > 0 {
		// Use linking agent's field suggestions
		sb.WriteString("\n=== Suggested Output Fields (from Schema Linking) ===\n")
		for _, f := range info.SuggestedFields {
			if f.Selected {
				sb.WriteString(fmt.Sprintf("  - %s.%s: %s\n", f.TableName, f.ColumnName, f.Reason))
			}
		}
		sb.WriteString("Consider including these fields in the SELECT clause.\n")
	}

	return tables, sb.String()
}

