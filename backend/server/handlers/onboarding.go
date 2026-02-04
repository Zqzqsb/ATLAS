package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/lakebase"
	"lucid/server/services"
)

// OnboardingEvent represents an event in the onboarding stream
type OnboardingEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// TableInfo holds real table information from database
type TableInfo struct {
	Name     string                `json:"name"`
	RowCount int64                 `json:"row_count"`
	Columns  []services.ColumnInfo `json:"columns"`
}

// AgentAnalysisResult holds the LLM analysis result for a table
type AgentAnalysisResult struct {
	TableName       string            `json:"table_name"`
	Description     string            `json:"description"`
	BusinessPurpose string            `json:"business_purpose"`
	ColumnNotes     map[string]string `json:"column_notes"`
	Relationships   []string          `json:"relationships"`
	SampleQueries   []string          `json:"sample_queries"`
	RichContext     map[string]string `json:"rich_context"`      // Rich context from ReAct exploration
	QualityIssues   []string          `json:"quality_issues"`    // Data quality issues discovered
	Iterations      int               `json:"iterations"`        // Number of ReAct iterations
}

// ReActIteration represents a single ReAct iteration for streaming
type ReActIteration struct {
	Index       int    `json:"index"`
	Thought     string `json:"thought"`
	Action      string `json:"action"`
	ActionInput string `json:"action_input"`
	Observation string `json:"observation"`
}

// CatalogSearchResult represents a result from catalog search (reserved for future)
type CatalogSearchResult struct {
	TableName   string   `json:"table_name"`
	ColumnName  string   `json:"column_name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Score       float64  `json:"score"`
}

// CatalogSearcher interface for future catalog vectorization
// TODO: Implement with vector database for semantic search
type CatalogSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]CatalogSearchResult, error)
	SearchByTable(ctx context.Context, tableName string, query string) ([]CatalogSearchResult, error)
}

// OnboardingStreamHandler handles SSE streaming for database onboarding
func (h *Handler) OnboardingStream(c *gin.Context) {
	connectionID := c.Query("connection_id")
	if connectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "connection_id parameter is required",
		})
		return
	}

	// Find connection config
	var connConfig *struct {
		ID   string
		Name string
		Type string
	}
	for _, db := range h.config.Databases {
		if db.ID == connectionID {
			connConfig = &struct {
				ID   string
				Name string
				Type string
			}{
				ID:   db.ID,
				Name: db.Name,
				Type: db.Type,
			}
			break
		}
	}

	if connConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Connection not found",
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
	defer cancel()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Streaming not supported",
		})
		return
	}

	// Create event channel
	events := make(chan OnboardingEvent, 100)

	// Start real onboarding process with database service
	go h.runOnboardingProcess(ctx, connectionID, connConfig.Name, connConfig.Type, events)

	// Stream events to client
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			sendOnboardingEvent(c.Writer, event)
			flusher.Flush()
		}
	}
}

// sendOnboardingEvent sends a single SSE event
func sendOnboardingEvent(w http.ResponseWriter, event OnboardingEvent) {
	jsonData, _ := json.Marshal(event.Data)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}

// runOnboardingProcess executes the real database analysis process
func (h *Handler) runOnboardingProcess(ctx context.Context, connectionID, dbName, dbType string, events chan<- OnboardingEvent) {
	defer close(events)

	startTime := time.Now()

	// Phase 1: Connecting
	sendPhaseChange(events, "connecting", fmt.Sprintf("Connecting to %s database...", dbType))

	// Try to get real schema from database
	schema, err := h.dbService.GetSchema(ctx, connectionID)
	if err != nil {
		// Connection failed - send error and fallback to demo mode
		events <- OnboardingEvent{
			Type: "error",
			Data: map[string]interface{}{
				"phase":   "connecting",
				"message": fmt.Sprintf("Failed to connect: %v. Using demo mode.", err),
			},
			Timestamp: time.Now().UnixMilli(),
		}

		// Fallback to simulated data for demo
		h.runSimulatedOnboarding(ctx, connectionID, dbName, dbType, events, startTime)
		return
	}

	time.Sleep(300 * time.Millisecond) // Brief delay for visual effect

	// Phase 2: Discovering tables (using real data)
	sendPhaseChange(events, "discovering", "Discovering database tables...")
	time.Sleep(200 * time.Millisecond)

	// Convert schema tables to TableInfo
	tables := make([]TableInfo, 0, len(schema.Tables))
	tableNames := make([]string, 0, len(schema.Tables))
	for _, t := range schema.Tables {
		tables = append(tables, TableInfo{
			Name:     t.Name,
			RowCount: t.RowCount,
			Columns:  t.Columns,
		})
		tableNames = append(tableNames, t.Name)
	}

	events <- OnboardingEvent{
		Type: "table_discovered",
		Data: map[string]interface{}{
			"tables":        tableNames,
			"count":         len(tables),
			"database_type": schema.DatabaseType,
		},
		Timestamp: time.Now().UnixMilli(),
	}
	time.Sleep(300 * time.Millisecond)

// Phase 3: Dispatching
	sendPhaseChange(events, "dispatching", "Dispatching analysis tasks to workers...")
	time.Sleep(200 * time.Millisecond)

	// Preheat LLM model (warm up the connection)
	go h.preheatLLM(ctx, events)

	// Phase 4: Agent-Driven Analyzing (with ReAct loop)
	sendPhaseChange(events, "analyzing", "Agent analyzing tables with ReAct loop...")

	// Process tables with Agent analysis
	analysisResults := make(map[string]*AgentAnalysisResult)
	var resultsMu sync.Mutex

	// Process tables in parallel using worker pool (max 2 for LLM rate limiting)
	var wg sync.WaitGroup
	workerCount := min(2, len(tables))
	if workerCount == 0 {
		workerCount = 1
	}

	tableChan := make(chan int, len(tables))
	for i := range tables {
		tableChan <- i
	}
	close(tableChan)

	for workerID := 1; workerID <= workerCount; workerID++ {
		wg.Add(1)
		go func(wID int) {
			defer wg.Done()
			for tableIdx := range tableChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := h.processTableWithReActAgent(ctx, events, wID, connectionID, tables[tableIdx], schema.DatabaseType, tableNames)
					if result != nil {
						resultsMu.Lock()
						analysisResults[tables[tableIdx].Name] = result
						resultsMu.Unlock()
					}
				}
			}
		}(workerID)
	}

	wg.Wait()

	// Phase 5: Building Rich Context
	sendPhaseChange(events, "building", "Building Rich Context from analysis...")
	time.Sleep(500 * time.Millisecond)

	// Send rich context summary
	events <- OnboardingEvent{
		Type: "rich_context_built",
		Data: map[string]interface{}{
			"tables_analyzed":  len(analysisResults),
			"analysis_results": analysisResults,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Phase 6: Save to Lake-base Storage
	var lakebaseResult *LakebaseOnboardingResult
	if h.lakebaseService != nil && h.lakebaseService.IsConnected() {
		lakebaseResult = h.saveOnboardingToLakebase(ctx, events, connectionID, dbName, schema.DatabaseType, tables, analysisResults)
	}

	// Complete
	totalTime := time.Since(startTime).Milliseconds()
	completeData := map[string]interface{}{
		"message":          "Onboarding completed successfully",
		"total_time":       totalTime,
		"database_id":      connectionID,
		"database_type":    schema.DatabaseType,
		"table_count":      len(tables),
		"agent_analyzed":   true,
		"analysis_results": analysisResults,
	}

	// Add lake-base result if available
	if lakebaseResult != nil {
		completeData["lakebase"] = lakebaseResult
	}

	events <- OnboardingEvent{
		Type:      "complete",
		Data:      completeData,
		Timestamp: time.Now().UnixMilli(),
	}
}

// processTableWithReActAgent processes a single table using ReAct Agent loop
func (h *Handler) processTableWithReActAgent(ctx context.Context, events chan<- OnboardingEvent, workerID int, connectionID string, table TableInfo, dbType string, allTables []string) *AgentAnalysisResult {
	// Worker assigned
	events <- OnboardingEvent{
		Type: "worker_assigned",
		Data: map[string]interface{}{
			"worker_id": workerID,
			"table":     table.Name,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Initialize result
	result := &AgentAnalysisResult{
		TableName:     table.Name,
		ColumnNotes:   make(map[string]string),
		Relationships: []string{},
		SampleQueries: []string{},
		RichContext:   make(map[string]string),
		QualityIssues: []string{},
	}

	// Phase 1: Extracting schema metadata (fixed queries)
	events <- OnboardingEvent{
		Type: "worker_progress",
		Data: map[string]interface{}{
			"worker_id":   workerID,
			"phase":       1,
			"total_phases": 3,
			"step":        1,
			"total_steps": 4,
			"message":     "Phase 1: Collecting basic metadata",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Execute metadata queries
	metadataQueries := getMetadataQueries(dbType, table.Name)
	for i, query := range metadataQueries {
		events <- OnboardingEvent{
			Type: "react_iteration",
			Data: map[string]interface{}{
				"worker_id":    workerID,
				"table":        table.Name,
				"phase":        1,
				"iteration":    i + 1,
				"thought":      fmt.Sprintf("Collecting metadata: %s", query.description),
				"action":       "execute_sql",
				"action_input": query.sql,
			},
			Timestamp: time.Now().UnixMilli(),
		}

		// Execute SQL
		sqlResult, err := h.dbService.ExecuteSQL(ctx, connectionID, query.sql)
		observation := ""
		if err != nil {
			observation = fmt.Sprintf("Error: %v", err)
		} else if sqlResult.Error != "" {
			observation = fmt.Sprintf("SQL Error: %s", sqlResult.Error)
		} else {
			observation = fmt.Sprintf("✓ %d rows returned", len(sqlResult.Rows))
		}

		events <- OnboardingEvent{
			Type: "react_observation",
			Data: map[string]interface{}{
				"worker_id":   workerID,
				"table":       table.Name,
				"phase":       1,
				"iteration":   i + 1,
				"observation": observation,
			},
			Timestamp: time.Now().UnixMilli(),
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Phase 2: ReAct exploration for rich context (LLM-driven)
	events <- OnboardingEvent{
		Type: "worker_progress",
		Data: map[string]interface{}{
			"worker_id":    workerID,
			"phase":        2,
			"total_phases": 3,
			"step":         1,
			"total_steps":  1,
			"message":      "Phase 2: ReAct exploring rich context...",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Send agent analysis start event
	events <- OnboardingEvent{
		Type: "agent_analysis_start",
		Data: map[string]interface{}{
			"worker_id": workerID,
			"table":     table.Name,
			"phase":     2,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Get LLM model from inference service
	llmModelInterface := h.inferenceService.GetLLMModel()
	if llmModelInterface == nil {
		events <- OnboardingEvent{
			Type: "agent_analysis_error",
			Data: map[string]interface{}{
				"worker_id": workerID,
				"table":     table.Name,
				"error":     "LLM not available, using rule-based analysis",
			},
			Timestamp: time.Now().UnixMilli(),
		}
		return createFallbackAnalysis(table)
	}

	llmModel, ok := llmModelInterface.(llms.Model)
	if !ok {
		events <- OnboardingEvent{
			Type: "agent_analysis_error",
			Data: map[string]interface{}{
				"worker_id": workerID,
				"table":     table.Name,
				"error":     "LLM type assertion failed, using rule-based analysis",
			},
			Timestamp: time.Now().UnixMilli(),
		}
		return createFallbackAnalysis(table)
	}

	// Run ReAct loop for rich context exploration
	maxIterations := 15
	iterationCount := 0
	conversationHistory := buildInitialReActPrompt(table, dbType, allTables)

	for iterationCount < maxIterations {
		iterationCount++

		// Send iteration start
		events <- OnboardingEvent{
			Type: "react_iteration_start",
			Data: map[string]interface{}{
				"worker_id": workerID,
				"table":     table.Name,
				"phase":     2,
				"iteration": iterationCount,
				"max":       maxIterations,
			},
			Timestamp: time.Now().UnixMilli(),
		}

		// Call LLM with streaming
		var fullResponse strings.Builder
		streamingContent := ""

		_, err := llms.GenerateFromSinglePrompt(ctx, llmModel, conversationHistory,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				token := string(chunk)
				fullResponse.WriteString(token)
				streamingContent += token

				// Send streaming token event
				events <- OnboardingEvent{
					Type: "agent_analysis_token",
					Data: map[string]interface{}{
						"worker_id": workerID,
						"table":     table.Name,
						"phase":     2,
						"iteration": iterationCount,
						"token":     token,
						"content":   streamingContent,
					},
					Timestamp: time.Now().UnixMilli(),
				}
				return nil
			}),
		)

		if err != nil {
			events <- OnboardingEvent{
				Type: "agent_analysis_error",
				Data: map[string]interface{}{
					"worker_id": workerID,
					"table":     table.Name,
					"error":     fmt.Sprintf("LLM error at iteration %d: %v", iterationCount, err),
				},
				Timestamp: time.Now().UnixMilli(),
			}
			break
		}

		response := fullResponse.String()

		// Parse ReAct response
		thought, action, actionInput, finalAnswer := parseReActResponse(response)

		// Send parsed iteration
		events <- OnboardingEvent{
			Type: "react_iteration",
			Data: map[string]interface{}{
				"worker_id":    workerID,
				"table":        table.Name,
				"phase":        2,
				"iteration":    iterationCount,
				"thought":      thought,
				"action":       action,
				"action_input": actionInput,
				"final_answer": finalAnswer,
			},
			Timestamp: time.Now().UnixMilli(),
		}

		// Check if agent wants to finish
		if finalAnswer != "" || action == "" {
			events <- OnboardingEvent{
				Type: "react_complete",
				Data: map[string]interface{}{
					"worker_id":    workerID,
					"table":        table.Name,
					"phase":        2,
					"iterations":   iterationCount,
					"final_answer": finalAnswer,
				},
				Timestamp: time.Now().UnixMilli(),
			}
			break
		}

		// Execute action and get observation
		observation := h.executeReActAction(ctx, events, workerID, connectionID, table.Name, action, actionInput, result)

		// Send observation
		events <- OnboardingEvent{
			Type: "react_observation",
			Data: map[string]interface{}{
				"worker_id":   workerID,
				"table":       table.Name,
				"phase":       2,
				"iteration":   iterationCount,
				"observation": observation,
			},
			Timestamp: time.Now().UnixMilli(),
		}

		// Update conversation history for next iteration
		conversationHistory += fmt.Sprintf("\n%s\nObservation: %s\n", response, observation)

		time.Sleep(200 * time.Millisecond)
	}

	result.Iterations = iterationCount

	// Phase 3: Generate final description
	events <- OnboardingEvent{
		Type: "worker_progress",
		Data: map[string]interface{}{
			"worker_id":    workerID,
			"phase":        3,
			"total_phases": 3,
			"step":         1,
			"total_steps":  1,
			"message":      "Phase 3: Generating table description...",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Generate description based on collected rich context
	result.Description, result.BusinessPurpose = h.generateTableDescriptionFromContext(ctx, llmModel, table, result.RichContext, result.QualityIssues)

	// Send agent analysis complete event
	events <- OnboardingEvent{
		Type: "agent_analysis_complete",
		Data: map[string]interface{}{
			"worker_id":       workerID,
			"table":           table.Name,
			"iterations":      iterationCount,
			"rich_context":    result.RichContext,
			"quality_issues":  result.QualityIssues,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Convert columns to map format for JSON
	columns := make([]map[string]interface{}, len(table.Columns))
	for i, col := range table.Columns {
		columns[i] = map[string]interface{}{
			"name":           col.Name,
			"type":           col.Type,
			"nullable":       col.Nullable,
			"is_primary_key": col.IsPrimaryKey,
		}
	}

	// Table completed with agent analysis
	events <- OnboardingEvent{
		Type: "table_completed",
		Data: map[string]interface{}{
			"worker_id":        workerID,
			"table":            table.Name,
			"description":      result.Description,
			"business_purpose": result.BusinessPurpose,
			"column_count":     len(table.Columns),
			"row_count":        table.RowCount,
			"columns":          columns,
			"column_notes":     result.ColumnNotes,
			"relationships":    result.Relationships,
			"sample_queries":   result.SampleQueries,
			"rich_context":     result.RichContext,
			"quality_issues":   result.QualityIssues,
			"iterations":       result.Iterations,
			"agent_analyzed":   true,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	return result
}

// MetadataQuery represents a metadata query with description
type MetadataQuery struct {
	sql         string
	description string
}

// getMetadataQueries returns fixed metadata queries based on database type
func getMetadataQueries(dbType, tableName string) []MetadataQuery {
	switch dbType {
	case "MySQL", "mysql":
		return []MetadataQuery{
			{fmt.Sprintf("DESCRIBE %s", tableName), "Column structure"},
			{fmt.Sprintf("SHOW INDEX FROM %s", tableName), "Index information"},
			{fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName), "Row count"},
		}
	case "PostgreSQL", "postgresql", "postgres":
		return []MetadataQuery{
			{fmt.Sprintf("SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name='%s'", tableName), "Column structure"},
			{fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename='%s'", tableName), "Index information"},
			{fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName), "Row count"},
		}
	case "SQLite", "sqlite":
		return []MetadataQuery{
			{fmt.Sprintf("PRAGMA table_info(%s)", tableName), "Column structure"},
			{fmt.Sprintf("PRAGMA index_list(%s)", tableName), "Index information"},
			{fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName), "Row count"},
			{fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName), "Foreign keys"},
		}
	default:
		return []MetadataQuery{
			{fmt.Sprintf("DESCRIBE %s", tableName), "Column structure"},
			{fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName), "Row count"},
		}
	}
}

// buildInitialReActPrompt builds the initial ReAct prompt for rich context exploration
func buildInitialReActPrompt(table TableInfo, dbType string, allTables []string) string {
	// Build columns description
	var columnsDesc strings.Builder
	var textColumns []string
	for _, col := range table.Columns {
		pkMarker := ""
		if col.IsPrimaryKey {
			pkMarker = " [PRIMARY KEY]"
		}
		nullableMarker := ""
		if col.Nullable {
			nullableMarker = " (nullable)"
		}
		columnsDesc.WriteString(fmt.Sprintf("  - %s: %s%s%s\n", col.Name, col.Type, pkMarker, nullableMarker))
		
		// Track TEXT/VARCHAR columns for quality checks
		colType := strings.ToUpper(col.Type)
		if strings.Contains(colType, "TEXT") || strings.Contains(colType, "VARCHAR") || strings.Contains(colType, "CHAR") {
			textColumns = append(textColumns, col.Name)
		}
	}

	// DB-specific SQL hints
	sqlHint := ""
	switch dbType {
	case "SQLite", "sqlite":
		sqlHint = "Note: This is SQLite. Use PRAGMA table_info(table_name) instead of DESCRIBE. Use TRIM() for whitespace check."
	case "MySQL", "mysql":
		sqlHint = "Note: This is MySQL. Use DESCRIBE table_name or SHOW COLUMNS FROM table_name."
	case "PostgreSQL", "postgresql", "postgres":
		sqlHint = "Note: This is PostgreSQL. Use information_schema.columns for metadata."
	}

	prompt := fmt.Sprintf(`You are analyzing table "%s" in %s database to discover RICH CONTEXT for better SQL generation.
%s

Table Info:
- Row Count: %d
- Columns:
%s
- Other tables in database: %s

**YOUR GOAL**: Discover DATA QUALITY ISSUES and BUSINESS MEANING that will help generate correct SQL queries.

**AVAILABLE TOOLS**:
1. execute_sql - Execute SQL queries to explore the data
2. set_rich_context - **IMPORTANT**: Save ALL discovered insights using key|value format. YOU MUST USE THIS TOOL to record findings.
3. search_catalog - Search business catalog for semantic meaning (reserved for future)

**CRITICAL**: Every important discovery MUST be saved using set_rich_context. Do NOT just execute SQL - you MUST also save the results.

**MANDATORY CHECKS** (do these in order):

1. For EACH TEXT column (%s), check whitespace:
   - First: Action: execute_sql, Action Input: SELECT [column] FROM %s WHERE [column] != TRIM([column]) LIMIT 3
   - If results found: Action: set_rich_context, Action Input: [column]_has_whitespace|⚠️ Column contains leading/trailing whitespace. Use TRIM() for matching.

2. Discover value distributions for enum-like columns (columns with < 20 distinct values):
   - First: Action: execute_sql, Action Input: SELECT [column], COUNT(*) FROM %s GROUP BY [column] ORDER BY COUNT(*) DESC
   - Then ALWAYS: Action: set_rich_context, Action Input: [column]_value_distribution|value1:count1, value2:count2, ...

3. Check for NULL patterns in important columns
   - Save findings: Action: set_rich_context, Action Input: [column]_null_info|X%% of values are NULL

4. For foreign key columns (columns ending with _id or _code), check orphan records

**OUTPUT FORMAT**:
Thought: What I'm thinking about doing
Action: execute_sql OR set_rich_context OR search_catalog
Action Input: The SQL query OR key|value pair OR search query

After each action, wait for Observation, then continue.

When finished exploring, output:
Final Answer: Summary of discovered rich context

**EXAMPLE**:
Thought: I should check the value distribution of the 'country' column
Action: execute_sql
Action Input: SELECT country, COUNT(*) as count FROM airports GROUP BY country ORDER BY count DESC LIMIT 10

[Wait for Observation]

Thought: Found the distribution. I need to save this as rich context.
Action: set_rich_context
Action Input: country_value_distribution|United States:5, Canada:3, Mexico:2

Begin your analysis:
`, table.Name, dbType, sqlHint, table.RowCount, columnsDesc.String(), strings.Join(allTables, ", "),
		strings.Join(textColumns, ", "), table.Name, table.Name)

	return prompt
}

// parseReActResponse parses the LLM response for Thought, Action, Action Input, or Final Answer
func parseReActResponse(response string) (thought, action, actionInput, finalAnswer string) {
	lines := strings.Split(response, "\n")
	
	var currentKey string
	var currentValue strings.Builder
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "Thought:") {
			if currentKey != "" {
				saveValue(&thought, &action, &actionInput, &finalAnswer, currentKey, currentValue.String())
			}
			currentKey = "thought"
			currentValue.Reset()
			currentValue.WriteString(strings.TrimPrefix(line, "Thought:"))
		} else if strings.HasPrefix(line, "Action:") {
			if currentKey != "" {
				saveValue(&thought, &action, &actionInput, &finalAnswer, currentKey, currentValue.String())
			}
			currentKey = "action"
			currentValue.Reset()
			currentValue.WriteString(strings.TrimPrefix(line, "Action:"))
		} else if strings.HasPrefix(line, "Action Input:") {
			if currentKey != "" {
				saveValue(&thought, &action, &actionInput, &finalAnswer, currentKey, currentValue.String())
			}
			currentKey = "action_input"
			currentValue.Reset()
			currentValue.WriteString(strings.TrimPrefix(line, "Action Input:"))
		} else if strings.HasPrefix(line, "Final Answer:") {
			if currentKey != "" {
				saveValue(&thought, &action, &actionInput, &finalAnswer, currentKey, currentValue.String())
			}
			currentKey = "final_answer"
			currentValue.Reset()
			currentValue.WriteString(strings.TrimPrefix(line, "Final Answer:"))
		} else if currentKey != "" && line != "" {
			currentValue.WriteString(" ")
			currentValue.WriteString(line)
		}
	}
	
	if currentKey != "" {
		saveValue(&thought, &action, &actionInput, &finalAnswer, currentKey, currentValue.String())
	}
	
	thought = strings.TrimSpace(thought)
	action = strings.TrimSpace(action)
	actionInput = strings.TrimSpace(actionInput)
	finalAnswer = strings.TrimSpace(finalAnswer)
	
	return
}

func saveValue(thought, action, actionInput, finalAnswer *string, key, value string) {
	value = strings.TrimSpace(value)
	switch key {
	case "thought":
		*thought = value
	case "action":
		*action = value
	case "action_input":
		*actionInput = value
	case "final_answer":
		*finalAnswer = value
	}
}

// executeReActAction executes a ReAct action and returns observation
func (h *Handler) executeReActAction(ctx context.Context, events chan<- OnboardingEvent, workerID int, connectionID, tableName, action, actionInput string, result *AgentAnalysisResult) string {
	action = strings.ToLower(strings.TrimSpace(action))
	
	switch action {
	case "execute_sql":
		return h.executeSQL(ctx, connectionID, actionInput)
	
	case "set_rich_context":
		return h.setRichContext(tableName, actionInput, result)
	
	case "search_catalog":
		// Reserved for future catalog vectorization
		return h.searchCatalog(ctx, tableName, actionInput)
	
	default:
		return fmt.Sprintf("Unknown action: %s. Available actions: execute_sql, set_rich_context, search_catalog", action)
	}
}

// executeSQL executes a SQL query and returns formatted observation
func (h *Handler) executeSQL(ctx context.Context, connectionID, sql string) string {
	sqlResult, err := h.dbService.ExecuteSQL(ctx, connectionID, sql)
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err)
	}
	if sqlResult.Error != "" {
		return fmt.Sprintf("❌ SQL Error: %s", sqlResult.Error)
	}
	
	// Format results
	rowCount := len(sqlResult.Rows)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✓ Query successful (%d rows)\n", rowCount))
	
	if rowCount > 0 && len(sqlResult.Rows) > 0 {
		// Show first few rows
		maxRows := 5
		if len(sqlResult.Rows) < maxRows {
			maxRows = len(sqlResult.Rows)
		}
		
		jsonBytes, _ := json.MarshalIndent(sqlResult.Rows[:maxRows], "", "  ")
		sb.WriteString("Results:\n")
		sb.WriteString(string(jsonBytes))
		
		if len(sqlResult.Rows) > maxRows {
			sb.WriteString(fmt.Sprintf("\n... and %d more rows", len(sqlResult.Rows)-maxRows))
		}
	} else if rowCount == 0 {
		sb.WriteString("No results returned (empty result set)")
	}
	
	return sb.String()
}

// setRichContext saves a key|value pair to rich context
func (h *Handler) setRichContext(tableName, input string, result *AgentAnalysisResult) string {
	parts := strings.SplitN(input, "|", 2)
	if len(parts) != 2 {
		return "❌ Invalid format. Use: key|value"
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	if key == "" || value == "" {
		return "❌ Key and value cannot be empty"
	}
	
	// Clean value (remove any trailing Thought/Action text)
	if idx := strings.Index(value, "\n\nThought:"); idx > 0 {
		value = value[:idx]
	}
	if idx := strings.Index(value, "\nThought:"); idx > 0 {
		value = value[:idx]
	}
	value = strings.TrimSpace(value)
	
	// Save to result
	result.RichContext[key] = value
	
	// Track quality issues separately
	if strings.Contains(key, "quality_issue") || strings.HasPrefix(value, "⚠️") {
		result.QualityIssues = append(result.QualityIssues, fmt.Sprintf("%s: %s", key, value))
	}
	
	return fmt.Sprintf("✓ Rich context saved: %s = %s", key, value)
}

// searchCatalog searches the business catalog (reserved for future vectorization)
func (h *Handler) searchCatalog(ctx context.Context, tableName, query string) string {
	// TODO: Implement with vector database for semantic catalog search
	// This will be connected to user-provided catalog data
	//
	// Future implementation:
	// 1. User uploads catalog (Excel/CSV with table/column descriptions)
	// 2. Catalog is vectorized and stored
	// 3. Agent can search catalog for semantic meaning
	// 4. Results help generate better descriptions and understand business context
	//
	// Example future usage:
	// - "search_catalog: what does customer_id mean?"
	// - "search_catalog: find tables related to orders"
	// - "search_catalog: what are valid status values?"
	
	return fmt.Sprintf("📚 Catalog search not yet implemented. Query: '%s'\nThis feature will be available when catalog vectorization is enabled.", query)
}

// generateTableDescriptionFromContext generates description based on collected rich context
func (h *Handler) generateTableDescriptionFromContext(ctx context.Context, llmModel llms.Model, table TableInfo, richContext map[string]string, qualityIssues []string) (description, businessPurpose string) {
	// Build context summary
	var contextSummary strings.Builder
	for key, value := range richContext {
		contextSummary.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
	}
	
	var issuesSummary strings.Builder
	for _, issue := range qualityIssues {
		issuesSummary.WriteString(fmt.Sprintf("- %s\n", issue))
	}

	prompt := fmt.Sprintf(`Based on the analysis of table "%s", generate a concise description.

Table: %s
Row Count: %d
Columns: %d

Rich Context Discovered:
%s

Data Quality Issues Found:
%s

Generate:
1. A one-sentence description of what this table stores
2. A brief business purpose explanation

Output JSON:
{"description": "...", "business_purpose": "..."}`,
		table.Name, table.Name, table.RowCount, len(table.Columns),
		contextSummary.String(), issuesSummary.String())

	response, err := llms.GenerateFromSinglePrompt(ctx, llmModel, prompt)
	if err != nil {
		description = fmt.Sprintf("Table '%s' with %d columns", table.Name, len(table.Columns))
		businessPurpose = "Data storage table"
		return
	}

	// Parse JSON response
	response = strings.TrimSpace(response)
	// Remove markdown code blocks if present
	response = regexp.MustCompile("(?s)```json\\s*(.*)\\s*```").ReplaceAllString(response, "$1")
	response = regexp.MustCompile("(?s)```\\s*(.*)\\s*```").ReplaceAllString(response, "$1")
	response = strings.TrimSpace(response)

	var parsed struct {
		Description     string `json:"description"`
		BusinessPurpose string `json:"business_purpose"`
	}
	
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		description = fmt.Sprintf("Table '%s' with %d columns", table.Name, len(table.Columns))
		businessPurpose = "Data storage table"
		return
	}

	description = parsed.Description
	businessPurpose = parsed.BusinessPurpose
	
	if description == "" {
		description = fmt.Sprintf("Table '%s' with %d columns", table.Name, len(table.Columns))
	}
	if businessPurpose == "" {
		businessPurpose = "Data storage table"
	}

	return
}

// createFallbackAnalysis creates a fallback analysis when LLM is not available
func createFallbackAnalysis(table TableInfo) *AgentAnalysisResult {
	columnNotes := make(map[string]string)
	for _, col := range table.Columns {
		note := fmt.Sprintf("%s column", col.Type)
		if col.IsPrimaryKey {
			note = "Primary key - " + note
		}
		columnNotes[col.Name] = note
	}

	return &AgentAnalysisResult{
		TableName:       table.Name,
		Description:     fmt.Sprintf("Table '%s' contains %d columns with %d rows", table.Name, len(table.Columns), table.RowCount),
		BusinessPurpose: "Data storage table",
		ColumnNotes:     columnNotes,
		Relationships:   []string{},
		SampleQueries:   []string{},
		RichContext:     map[string]string{},
		QualityIssues:   []string{},
		Iterations:      0,
	}
}

// runSimulatedOnboarding runs a demo simulation when real DB connection fails
func (h *Handler) runSimulatedOnboarding(ctx context.Context, connectionID, dbName, dbType string, events chan<- OnboardingEvent, startTime time.Time) {
	// Phase 2: Discovering tables (simulated)
	sendPhaseChange(events, "discovering", "Discovering database tables (demo mode)...")
	time.Sleep(500 * time.Millisecond)

	// Get simulated tables based on connection ID pattern
	tables := getSimulatedTablesForDB(connectionID)

	events <- OnboardingEvent{
		Type: "table_discovered",
		Data: map[string]interface{}{
			"tables":    tables,
			"count":     len(tables),
			"demo_mode": true,
		},
		Timestamp: time.Now().UnixMilli(),
	}
	time.Sleep(300 * time.Millisecond)

	// Phase 3: Dispatching
	sendPhaseChange(events, "dispatching", "Dispatching analysis tasks to workers...")
	time.Sleep(300 * time.Millisecond)

	// Phase 4: Analyzing (parallel workers)
	sendPhaseChange(events, "analyzing", "Workers analyzing tables...")

	var wg sync.WaitGroup
	workerCount := min(4, len(tables))
	if workerCount == 0 {
		workerCount = 1
	}

	tableChan := make(chan int, len(tables))
	for i := range tables {
		tableChan <- i
	}
	close(tableChan)

	for workerID := 1; workerID <= workerCount; workerID++ {
		wg.Add(1)
		go func(wID int) {
			defer wg.Done()
			for tableIdx := range tableChan {
				select {
				case <-ctx.Done():
					return
				default:
					processSimulatedTable(events, wID, tables[tableIdx])
				}
			}
		}(workerID)
	}

	wg.Wait()

	// Phase 5: Building Rich Context
	sendPhaseChange(events, "building", "Building Rich Context...")
	time.Sleep(800 * time.Millisecond)

	// Complete
	totalTime := time.Since(startTime).Milliseconds()
	events <- OnboardingEvent{
		Type: "complete",
		Data: map[string]interface{}{
			"message":    "Onboarding completed (demo mode)",
			"total_time": totalTime,
			"demo_mode":  true,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

// sendPhaseChange sends a phase change event
func sendPhaseChange(events chan<- OnboardingEvent, phase, message string) {
	events <- OnboardingEvent{
		Type: "phase_change",
		Data: map[string]interface{}{
			"phase":   phase,
			"message": message,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

// processSimulatedTable simulates processing a single table (for demo mode)
func processSimulatedTable(events chan<- OnboardingEvent, workerID int, tableName string) {
	// Worker assigned
	events <- OnboardingEvent{
		Type: "worker_assigned",
		Data: map[string]interface{}{
			"worker_id": workerID,
			"table":     tableName,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// Simulate analysis steps
	totalSteps := 5
	stepMessages := []string{
		"Extracting schema metadata",
		"Analyzing column types",
		"Sampling data patterns",
		"Discovering relationships",
		"Generating semantic description",
	}

	for step := 1; step <= totalSteps; step++ {
		time.Sleep(200 * time.Millisecond)
		events <- OnboardingEvent{
			Type: "worker_progress",
			Data: map[string]interface{}{
				"worker_id":   workerID,
				"step":        step,
				"total_steps": totalSteps,
				"message":     stepMessages[step-1],
			},
			Timestamp: time.Now().UnixMilli(),
		}
	}

	// Table completed with simulated data
	columns := generateSimulatedColumns(tableName)
	events <- OnboardingEvent{
		Type: "table_completed",
		Data: map[string]interface{}{
			"worker_id":    workerID,
			"table":        tableName,
			"description":  fmt.Sprintf("Table storing %s data", tableName),
			"column_count": len(columns),
			"row_count":    100 + len(tableName)*50, // Deterministic for demo
			"columns":      columns,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}

// getSimulatedTablesForDB returns appropriate table names based on database ID
func getSimulatedTablesForDB(connectionID string) []string {
	// Match based on connection ID patterns from demo_databases.json
	switch {
	case contains(connectionID, "car"):
		return []string{
			"car_makers",
			"model_list",
			"car_names",
			"cars_data",
			"continents",
			"countries",
		}
	case contains(connectionID, "flight"):
		return []string{
			"airlines",
			"airports",
			"flights",
			"routes",
		}
	case contains(connectionID, "college"):
		return []string{
			"class",
			"department",
			"enrolled_in",
			"faculty",
			"gradeconversion",
			"member",
			"student",
		}
	case contains(connectionID, "employee"):
		return []string{
			"employee",
			"shop",
			"hiring",
			"evaluation",
		}
	default:
		// Generic demo tables
		return []string{
			"users",
			"orders",
			"products",
			"categories",
		}
	}
}

// contains checks if substr is in s (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			tc := substr[j]
			// Simple ASCII lowercase comparison
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if tc >= 'A' && tc <= 'Z' {
				tc += 32
			}
			if sc != tc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// generateSimulatedColumns generates simulated column info based on table name
func generateSimulatedColumns(tableName string) []map[string]interface{} {
	baseColumns := []map[string]interface{}{
		{"name": "id", "type": "INTEGER", "is_primary_key": true, "nullable": false},
	}

	// Add table-specific columns based on common patterns
	switch tableName {
	case "car_makers":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "maker", "type": "VARCHAR(255)", "nullable": false},
			map[string]interface{}{"name": "full_name", "type": "VARCHAR(255)", "nullable": true},
			map[string]interface{}{"name": "country", "type": "VARCHAR(100)", "nullable": true},
		)
	case "cars_data":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "mpg", "type": "DECIMAL(5,2)", "nullable": true},
			map[string]interface{}{"name": "cylinders", "type": "INTEGER", "nullable": true},
			map[string]interface{}{"name": "horsepower", "type": "INTEGER", "nullable": true},
			map[string]interface{}{"name": "weight", "type": "INTEGER", "nullable": true},
			map[string]interface{}{"name": "year", "type": "INTEGER", "nullable": true},
		)
	case "airlines":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "name", "type": "VARCHAR(255)", "nullable": false},
			map[string]interface{}{"name": "iata", "type": "VARCHAR(10)", "nullable": true},
			map[string]interface{}{"name": "country", "type": "VARCHAR(100)", "nullable": true},
		)
	case "flights":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "airline_id", "type": "INTEGER", "nullable": false},
			map[string]interface{}{"name": "source_airport", "type": "VARCHAR(10)", "nullable": false},
			map[string]interface{}{"name": "dest_airport", "type": "VARCHAR(10)", "nullable": false},
			map[string]interface{}{"name": "departure_time", "type": "TIMESTAMP", "nullable": true},
		)
	case "student":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "name", "type": "VARCHAR(255)", "nullable": false},
			map[string]interface{}{"name": "dept_name", "type": "VARCHAR(100)", "nullable": true},
			map[string]interface{}{"name": "tot_cred", "type": "INTEGER", "nullable": true},
		)
	case "employee":
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "name", "type": "VARCHAR(255)", "nullable": false},
			map[string]interface{}{"name": "age", "type": "INTEGER", "nullable": true},
			map[string]interface{}{"name": "city", "type": "VARCHAR(100)", "nullable": true},
		)
	default:
		baseColumns = append(baseColumns,
			map[string]interface{}{"name": "name", "type": "VARCHAR(255)", "nullable": false},
			map[string]interface{}{"name": "description", "type": "TEXT", "nullable": true},
			map[string]interface{}{"name": "created_at", "type": "TIMESTAMP", "nullable": true},
		)
	}

	return baseColumns
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// preheatLLM warms up the LLM connection to reduce first call latency
func (h *Handler) preheatLLM(ctx context.Context, events chan<- OnboardingEvent) {
	events <- OnboardingEvent{
		Type: "llm_preheat_start",
		Data: map[string]interface{}{
			"message": "Preheating AI Agent...",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	llmModelInterface := h.inferenceService.GetLLMModel()
	if llmModelInterface == nil {
		return
	}

	llmModel, ok := llmModelInterface.(llms.Model)
	if !ok {
		return
	}

	// Simple preheat call to warm up connection
	_, err := llms.GenerateFromSinglePrompt(ctx, llmModel, "Hello", llms.WithMaxTokens(1))
	if err == nil {
		events <- OnboardingEvent{
			Type: "llm_preheat_complete",
			Data: map[string]interface{}{
				"message": "AI Agent ready",
			},
			Timestamp: time.Now().UnixMilli(),
		}
	}
}

// ===========================================
// Lake-base Storage Integration
// ===========================================

// LakebaseOnboardingResult holds the result of saving onboarding data to lake-base
type LakebaseOnboardingResult struct {
	DatasourceID    int64  `json:"datasource_id"`
	DatasourceName  string `json:"datasource_name"`
	TablesCount     int    `json:"tables_count"`
	ColumnsCount    int    `json:"columns_count"`
	ContextCount    int    `json:"context_count"`
	EmbeddingsCount int    `json:"embeddings_count"`
	Success         bool   `json:"success"`
	Error           string `json:"error,omitempty"`
}

// saveOnboardingToLakebase saves all onboarding results to lake-base storage
func (h *Handler) saveOnboardingToLakebase(
	ctx context.Context,
	events chan<- OnboardingEvent,
	connectionID string,
	dbName string,
	dbType string,
	tables []TableInfo,
	analysisResults map[string]*AgentAnalysisResult,
) *LakebaseOnboardingResult {
	result := &LakebaseOnboardingResult{
		DatasourceName: dbName,
	}

	// Check if lake-base service is available
	if h.lakebaseService == nil || !h.lakebaseService.IsConnected() {
		result.Error = "Lake-base service not available"
		return result
	}

	// Send phase change event
	events <- OnboardingEvent{
		Type: "phase_change",
		Data: map[string]interface{}{
			"phase":   "lakebase_storage",
			"message": "Saving to lake-base storage...",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// 1. Create or get datasource
	ds := &lakebase.Datasource{
		Name:         connectionID,
		DBType:       dbType,
		Host:         sql.NullString{String: "onboarding", Valid: true},
		Port:         sql.NullInt32{Int32: 0, Valid: true},
		Username:     sql.NullString{},
		DatabaseName: sql.NullString{String: dbName, Valid: true},
		Status:       lakebase.DatasourceStatusActive,
	}

	existingDS, err := h.lakebaseService.GetOrCreateDatasource(ctx, ds)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create datasource: %v", err)
		sendLakebaseError(events, result.Error)
		return result
	}
	result.DatasourceID = existingDS.ID

	// 2. Build schema metadata from tables
	var schemaMetas []*lakebase.SchemaMetadata
	for _, table := range tables {
		for _, col := range table.Columns {
			meta := &lakebase.SchemaMetadata{
				DatasourceID: existingDS.ID,
				TableName:    table.Name,
				ColumnName:   col.Name,
				DataType:     col.Type,
				IsPrimaryKey: col.IsPrimaryKey,
				Nullable:     col.Nullable,
			}
			schemaMetas = append(schemaMetas, meta)
		}
	}

	// Save schema metadata
	if len(schemaMetas) > 0 {
		if err := h.lakebaseService.SaveSchemaMetadata(ctx, schemaMetas); err != nil {
			result.Error = fmt.Sprintf("Failed to save schema metadata: %v", err)
			sendLakebaseError(events, result.Error)
			return result
		}
		result.ColumnsCount = len(schemaMetas)

		// Count unique tables
		tableSet := make(map[string]bool)
		for _, meta := range schemaMetas {
			tableSet[meta.TableName] = true
		}
		result.TablesCount = len(tableSet)
	}

	events <- OnboardingEvent{
		Type: "lakebase_progress",
		Data: map[string]interface{}{
			"step":    "schema_saved",
			"tables":  result.TablesCount,
			"columns": result.ColumnsCount,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// 3. Build business context from analysis results
	var contexts []*lakebase.BusinessContext
	for tableName, analysis := range analysisResults {
		if analysis == nil {
			continue
		}

		// Save table-level semantic description
		if analysis.Description != "" {
			content, _ := lakebase.NewSemanticContent(analysis.Description, nil, nil)
			contexts = append(contexts, &lakebase.BusinessContext{
				DatasourceID: existingDS.ID,
				TableName:    tableName,
				ColumnName:   sql.NullString{},
				ContextType:  lakebase.ContextTypeSemantic,
				Content:      content,
				Source:       lakebase.SourceLLM,
				Confidence:   0.8,
				Version:      1,
				CreatedBy:    "onboarding",
				UpdatedBy:    "onboarding",
			})
		}

		// Save business purpose as business rule
		if analysis.BusinessPurpose != "" {
			content, _ := lakebase.NewBusinessRuleContent([]string{analysis.BusinessPurpose}, nil)
			contexts = append(contexts, &lakebase.BusinessContext{
				DatasourceID: existingDS.ID,
				TableName:    tableName,
				ColumnName:   sql.NullString{},
				ContextType:  lakebase.ContextTypeBusinessRule,
				Content:      content,
				Source:       lakebase.SourceLLM,
				Confidence:   0.8,
				Version:      1,
				CreatedBy:    "onboarding",
				UpdatedBy:    "onboarding",
			})
		}

		// Save rich context entries
		for key, value := range analysis.RichContext {
			// Determine context type based on key patterns
			contextType := lakebase.ContextTypeSemantic
			if strings.Contains(key, "value_distribution") || strings.Contains(key, "enum") {
				contextType = lakebase.ContextTypeEnumMeaning
			} else if strings.Contains(key, "quality") || strings.Contains(key, "whitespace") || strings.Contains(key, "null") {
				contextType = lakebase.ContextTypeDataQuality
			} else if strings.Contains(key, "join") || strings.Contains(key, "relationship") {
				contextType = lakebase.ContextTypeJoinHint
			}

			// Extract column name from key if present (format: columnName_contextInfo)
			columnName := sql.NullString{}
			if parts := strings.SplitN(key, "_", 2); len(parts) >= 1 {
				// Check if first part looks like a column name
				for _, table := range tables {
					if table.Name == tableName {
						for _, col := range table.Columns {
							if strings.EqualFold(col.Name, parts[0]) {
								columnName = sql.NullString{String: col.Name, Valid: true}
								break
							}
						}
						break
					}
				}
			}

			content, _ := json.Marshal(map[string]string{
				"key":   key,
				"value": value,
			})
			contexts = append(contexts, &lakebase.BusinessContext{
				DatasourceID: existingDS.ID,
				TableName:    tableName,
				ColumnName:   columnName,
				ContextType:  contextType,
				Content:      content,
				Source:       lakebase.SourceLLM,
				Confidence:   0.75,
				Version:      1,
				CreatedBy:    "onboarding",
				UpdatedBy:    "onboarding",
			})
		}

		// Save data quality issues
		for _, issue := range analysis.QualityIssues {
			content, _ := json.Marshal(lakebase.DataQualityContent{
				Anomalies: []string{issue},
			})
			contexts = append(contexts, &lakebase.BusinessContext{
				DatasourceID: existingDS.ID,
				TableName:    tableName,
				ColumnName:   sql.NullString{},
				ContextType:  lakebase.ContextTypeDataQuality,
				Content:      content,
				Source:       lakebase.SourceLLM,
				Confidence:   0.9,
				Version:      1,
				CreatedBy:    "onboarding",
				UpdatedBy:    "onboarding",
			})
		}
	}

	// Save business contexts
	if len(contexts) > 0 {
		if err := h.lakebaseService.SaveBusinessContextBatch(ctx, contexts); err != nil {
			result.Error = fmt.Sprintf("Failed to save business context: %v", err)
			sendLakebaseError(events, result.Error)
			return result
		}
		result.ContextCount = len(contexts)
	}

	events <- OnboardingEvent{
		Type: "lakebase_progress",
		Data: map[string]interface{}{
			"step":          "context_saved",
			"context_count": result.ContextCount,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	// 4. Create change log entry for onboarding
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"tables":   result.TablesCount,
		"columns":  result.ColumnsCount,
		"contexts": result.ContextCount,
	})

	_, _ = h.lakebaseService.CreateChangeLog(ctx, &lakebase.ChangeLog{
		DatasourceID:  existingDS.ID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceSystem,
		ChangeReason:  "Initial onboarding completed",
	})

	// 5. Generate embeddings for semantic grounding
	events <- OnboardingEvent{
		Type: "lakebase_progress",
		Data: map[string]interface{}{
			"step":    "generating_embeddings",
			"message": "Generating embeddings for semantic grounding...",
		},
		Timestamp: time.Now().UnixMilli(),
	}

	embeddingResult, err := h.lakebaseService.GenerateAndSaveEmbeddings(ctx, existingDS.ID)
	if err != nil {
		// Log warning but don't fail onboarding
		events <- OnboardingEvent{
			Type: "lakebase_progress",
			Data: map[string]interface{}{
				"step":    "embeddings_warning",
				"message": fmt.Sprintf("Embedding generation partial: %v", err),
			},
			Timestamp: time.Now().UnixMilli(),
		}
	} else if embeddingResult != nil {
		result.EmbeddingsCount = embeddingResult.TotalEmbeddings
		events <- OnboardingEvent{
			Type: "lakebase_progress",
			Data: map[string]interface{}{
				"step":             "embeddings_generated",
				"embeddings_count": embeddingResult.TotalEmbeddings,
				"tables_processed": embeddingResult.TablesProcessed,
				"columns_processed": embeddingResult.ColumnsProcessed,
				"contexts_processed": embeddingResult.ContextsProcessed,
			},
			Timestamp: time.Now().UnixMilli(),
		}
	}

	result.Success = true

	events <- OnboardingEvent{
		Type: "lakebase_complete",
		Data: map[string]interface{}{
			"datasource_id": result.DatasourceID,
			"tables":        result.TablesCount,
			"columns":       result.ColumnsCount,
			"contexts":      result.ContextCount,
			"embeddings":    result.EmbeddingsCount,
			"success":       true,
		},
		Timestamp: time.Now().UnixMilli(),
	}

	return result
}

// sendLakebaseError sends a lake-base error event
func sendLakebaseError(events chan<- OnboardingEvent, message string) {
	events <- OnboardingEvent{
		Type: "lakebase_error",
		Data: map[string]interface{}{
			"error": message,
		},
		Timestamp: time.Now().UnixMilli(),
	}
}
