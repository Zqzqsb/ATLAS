package inference

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/logger"
)

// Config holds inference pipeline configuration.
type Config struct {
	UseRichContext bool
	UseReact       bool
	ReactLinking   bool // Whether Schema Linking uses ReAct mode
	UseDryRun      bool
	MaxIterations  int

	// Clarify mode: "off" | "on" (agent asks) | "force" (forced fields)
	ClarifyMode             string
	ResultFields            []string // Expected result fields
	ResultFieldsDescription string

	DBName string
	DBType string
}

// SchemaContext holds rich context information for the inference pipeline.
// This replaces the dependency on internal/context.SharedContext.
type SchemaContext struct {
	DatabaseName string
	DatabaseType string
	Tables       map[string]*SchemaTable
}

// SchemaTable holds table-level schema and context information.
type SchemaTable struct {
	Name        string
	Description string
	RowCount    int64
	Columns     []SchemaColumn
	ForeignKeys []ForeignKeyRef
}

// SchemaColumn holds column-level information.
type SchemaColumn struct {
	Name         string
	Type         string
	Description  string
	IsPrimaryKey bool
	IsNullable   bool
	SampleValues string
	Synonyms     string
}

// ForeignKeyRef holds foreign key reference.
type ForeignKeyRef struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
}

// StepCallback is called for each ReAct step update during streaming.
// eventType: "thought" | "action" | "observation" | "finish"
type StepCallback func(step ReActStep, eventType string)

// PreLinkedContext holds the result of an external Linking stage (e.g. Grounding).
// When set, the inference pipeline skips its own Schema Linking and uses this directly.
type PreLinkedContext struct {
	SelectedTables []string // Table names selected by the linking agent
	ContextPrompt  string   // Rich context prompt ready to inject into SQL generation
}

// Pipeline is the Text-to-SQL inference pipeline.
type Pipeline struct {
	llm          llms.Model
	adapter      adapter.DBAdapter
	config       *Config
	schema       *SchemaContext
	schemaLinker SchemaLinker

	// External linking result — when set, skip internal Schema Linking
	preLinked *PreLinkedContext

	// Streaming callback
	stepCallback StepCallback
}

// Result holds inference output.
type Result struct {
	Query           string
	GeneratedSQL    string
	ExecutionResult interface{}

	TotalTime      time.Duration
	LLMCalls       int
	SQLExecutions  int
	SelectedTables []string
	ReActSteps     []ReActStep
}

// ReActStep represents a single ReAct reasoning step.
type ReActStep struct {
	Step        int         `json:"step,omitempty"`              // Step number for streaming
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"` // Supports string and map[string]interface{}
	Observation string      `json:"observation,omitempty"`
	Phase       string      `json:"phase,omitempty"` // "schema_linking" or "sql_generation"
}

// Reset cleans up pipeline state.
func (p *Pipeline) Reset() {
	p.stepCallback = nil
}

// SetStepCallback sets the callback function for streaming ReAct steps
func (p *Pipeline) SetStepCallback(callback StepCallback) {
	p.stepCallback = callback
}

// notifyStep notifies the callback of a ReAct step update
func (p *Pipeline) notifyStep(step ReActStep, eventType string) {
	if p.stepCallback != nil {
		p.stepCallback(step, eventType)
	}
}

// NewPipeline creates a new inference pipeline.
func NewPipeline(llm llms.Model, dbAdapter adapter.DBAdapter, config *Config) *Pipeline {
	linker := NewLLMSchemaLinker(llm, dbAdapter, config.ReactLinking)

	return &Pipeline{
		llm:          llm,
		adapter:      dbAdapter,
		config:       config,
		schemaLinker: linker,
	}
}

// SetSchemaContext sets the rich schema context (from lakebase).
func (p *Pipeline) SetSchemaContext(sc *SchemaContext) {
	p.schema = sc
}

// SetPreLinkedContext sets an externally computed linking result.
// When set, Execute() will skip its internal Schema Linking stage.
func (p *Pipeline) SetPreLinkedContext(plc *PreLinkedContext) {
	p.preLinked = plc
}

// Execute runs the inference pipeline.
func (p *Pipeline) Execute(ctx context.Context, query string) (*Result, error) {
	startTime := time.Now()
	log := logger.With("component", "inference_pipeline")

	log.Info("[Execute] Starting inference",
		"query", query,
		"use_rich_context", p.config.UseRichContext,
		"use_react", p.config.UseReact,
		"max_iterations", p.config.MaxIterations,
		"db_name", p.config.DBName,
		"has_pre_linked", p.preLinked != nil,
	)

	result := &Result{
		Query:      query,
		ReActSteps: []ReActStep{},
	}

	// 1. Schema Linking — identify relevant tables
	// If an external linking result is provided, skip internal linking entirely.
	var tables []string
	var contextPrompt string

	if p.preLinked != nil {
		// Use externally provided linking result (from Grounding / AdaptivePipeline)
		tables = p.preLinked.SelectedTables
		contextPrompt = p.preLinked.ContextPrompt
		result.SelectedTables = tables

		// If grounding selected 0 tables, fall back to all tables from schema
		// so SQL generation still has schema context to work with.
		if len(tables) == 0 && p.schema != nil {
			allTables := make([]string, 0, len(p.schema.Tables))
			for name := range p.schema.Tables {
				allTables = append(allTables, name)
			}
			tables = allTables
			result.SelectedTables = tables
			log.Warn("[Execute] Grounding selected 0 tables, falling back to all schema tables",
				"fallback_tables", len(tables),
			)
		}

		p.notifyStep(ReActStep{
			Step:        1,
			Thought:     fmt.Sprintf("Using pre-linked context from Grounding: %d tables selected", len(tables)),
			Observation: fmt.Sprintf("Pre-linked tables: %s", strings.Join(tables, ", ")),
			Phase:       "schema_linking",
		}, "finish")

		log.Info("[Execute] Using pre-linked context",
			"tables", strings.Join(tables, ", "),
			"context_prompt_length", len(contextPrompt),
		)

		// If the external prompt is empty but we have schema, build it from schema
		if contextPrompt == "" && p.schema != nil {
			contextPrompt = p.buildRichSchemaPrompt(tables)
		}
	} else {
		// Internal Schema Linking (legacy path)
		var allTableInfo map[string]*TableInfo
		var err error
		if p.schema != nil {
			allTableInfo = extractTableInfoFromSchema(p.schema)
		} else {
			allTableInfo, err = p.extractTableInfoFromDB(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to extract table info: %w", err)
			}
		}

		p.notifyStep(ReActStep{
			Step:    0,
			Thought: "Starting Schema Linking to identify relevant tables...",
			Phase:   "schema_linking",
		}, "thought")

		var schemaLinkingSteps []ReActStep
		tables, schemaLinkingSteps, err = p.schemaLinker.Link(ctx, query, allTableInfo)
		if err != nil {
			return nil, fmt.Errorf("schema linking failed: %w", err)
		}
		result.SelectedTables = tables
		result.LLMCalls++

		for i, step := range schemaLinkingSteps {
			reactStep := ReActStep{
				Step:        i + 1,
				Thought:     step.Thought,
				Action:      step.Action,
				ActionInput: step.ActionInput,
				Observation: step.Observation,
				Phase:       "schema_linking",
			}
			result.ReActSteps = append(result.ReActSteps, reactStep)
			if step.Thought != "" {
				p.notifyStep(reactStep, "thought")
			}
			if step.Action != "" {
				p.notifyStep(reactStep, "action")
			}
			if step.Observation != "" {
				p.notifyStep(reactStep, "observation")
			}
		}

		p.notifyStep(ReActStep{
			Step:        len(schemaLinkingSteps) + 1,
			Thought:     fmt.Sprintf("Schema Linking completed. Selected %d tables: %v", len(tables), tables),
			Observation: fmt.Sprintf("Selected tables: %s", strings.Join(tables, ", ")),
			Phase:       "schema_linking",
		}, "finish")

		log.Info("[Execute] Internal linking completed",
			"tables", strings.Join(tables, ", "),
		)

		// Build schema prompt for SQL generation
		if p.config.UseRichContext && p.schema != nil {
			contextPrompt = p.buildRichSchemaPrompt(tables)
			log.Info("[Execute] Using Rich Context",
				"table_count", len(tables),
				"prompt_length", len(contextPrompt),
			)
			log.Debug("[Execute] Rich Context prompt", "prompt", contextPrompt)
		} else {
			contextPrompt = p.buildBasicSchema(ctx, tables)
			log.Info("[Execute] Using Basic Schema",
				"table_count", len(tables),
				"prompt_length", len(contextPrompt),
			)
			log.Debug("[Execute] Basic Schema prompt", "prompt", contextPrompt)
		}
	}

	// 2. Generate SQL
	log.Info("[Execute] Starting SQL generation",
		"mode", func() string { if p.config.UseReact { return "ReAct" }; return "OneShot" }(),
		"context_prompt_length", len(contextPrompt),
	)
	var sql string
	var err error
	if p.config.UseReact {
		sql, err = p.reactLoop(ctx, query, contextPrompt, result)
	} else {
		sql, err = p.oneShotGeneration(ctx, query, contextPrompt)
		result.LLMCalls++
	}
	if err != nil {
		return nil, fmt.Errorf("SQL generation failed: %w", err)
	}

	result.GeneratedSQL = sql
	result.TotalTime = time.Since(startTime)

	log.Info("[Execute] SQL generated",
		"sql", sql,
		"llm_calls", result.LLMCalls,
		"total_time", result.TotalTime.Round(time.Millisecond),
	)

	// 3. Execute SQL
	if sql != "" {
		log.Info("[Execute] Executing generated SQL",
			"sql", sql,
			"database", p.config.DBName,
		)
		execResult, err := p.adapter.ExecuteQuery(ctx, sql)
		if err != nil {
			log.Error("[Execute] SQL execution failed",
				"sql", sql,
				"error", err,
			)
		} else {
			result.ExecutionResult = execResult
			result.SQLExecutions++
			log.Info("[Execute] SQL execution completed",
				"row_count", execResult.RowCount,
				"columns", execResult.Columns,
				"execution_time_ms", execResult.ExecutionTime,
			)
		}
	}

	log.Info("[Execute] Pipeline complete",
		"total_time", time.Since(startTime).Round(time.Millisecond),
		"llm_calls", result.LLMCalls,
		"sql_executions", result.SQLExecutions,
	)

	return result, nil
}

// extractTableInfoFromSchema extracts TableInfo from SchemaContext for schema linking.
func extractTableInfoFromSchema(sc *SchemaContext) map[string]*TableInfo {
	result := make(map[string]*TableInfo, len(sc.Tables))
	for name, table := range sc.Tables {
		columns := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columns[i] = col.Name
		}
		fks := make([]ForeignKeyRef, len(table.ForeignKeys))
		copy(fks, table.ForeignKeys)

		result[name] = &TableInfo{
			Name:        name,
			Columns:     columns,
			ForeignKeys: fks,
			Description: table.Description,
		}
	}
	return result
}

// buildRichSchemaPrompt builds a rich context prompt from SchemaContext for selected tables.
func (p *Pipeline) buildRichSchemaPrompt(tables []string) string {
	if p.schema == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Database: %s (%s)\n\n", p.schema.DatabaseName, p.schema.DatabaseType))

	for _, tableName := range tables {
		table, exists := p.schema.Tables[tableName]
		if !exists {
			continue
		}

		sb.WriteString(fmt.Sprintf("Table: %s", tableName))
		if table.RowCount > 0 {
			sb.WriteString(fmt.Sprintf(" (%d rows)", table.RowCount))
		}
		sb.WriteString("\n")
		if table.Description != "" {
			sb.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
		}

		for _, col := range table.Columns {
			sb.WriteString(fmt.Sprintf("  - %s: %s", col.Name, col.Type))
			if col.IsPrimaryKey {
				sb.WriteString(" [PK]")
			}
			if !col.IsNullable {
				sb.WriteString(" [NOT NULL]")
			}
			sb.WriteString("\n")
			if col.Description != "" {
				sb.WriteString(fmt.Sprintf("    Description: %s\n", col.Description))
			}
			if col.SampleValues != "" {
				sb.WriteString(fmt.Sprintf("    Examples: %s\n", col.SampleValues))
			}
			if col.Synonyms != "" {
				sb.WriteString(fmt.Sprintf("    Synonyms: %s\n", col.Synonyms))
			}
		}

		if len(table.ForeignKeys) > 0 {
			sb.WriteString("  Foreign Keys:\n")
			for _, fk := range table.ForeignKeys {
				sb.WriteString(fmt.Sprintf("    %s → %s.%s\n", fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// extractTableInfoFromDB extracts table info from the database directly.
func (p *Pipeline) extractTableInfoFromDB(ctx context.Context) (map[string]*TableInfo, error) {
	// Get all table names
	var query string
	switch p.adapter.GetDatabaseType() {
	case "MySQL":
		query = "SHOW TABLES"
	case "PostgreSQL":
		query = "SELECT tablename FROM pg_tables WHERE schemaname='public'"
	case "SQLite":
		query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	default:
		return nil, fmt.Errorf("unsupported database type")
	}

	result, err := p.adapter.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	tableInfo := make(map[string]*TableInfo)

	// Query column info for each table
	for _, row := range result.Rows {
		var tableName string
		for _, val := range row {
			if name, ok := val.(string); ok {
				tableName = name
				break
			}
		}

		if tableName == "" {
			continue
		}

		// Query column information
		var colQuery string
		switch p.adapter.GetDatabaseType() {
		case "MySQL":
			colQuery = fmt.Sprintf("DESCRIBE %s", tableName)
		case "SQLite":
			colQuery = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		case "PostgreSQL":
			colQuery = fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name='%s'", tableName)
		}

		colResult, err := p.adapter.ExecuteQuery(ctx, colQuery)
		if err != nil {
			continue
		}

		columns := make([]string, 0, len(colResult.Rows))
		for _, colRow := range colResult.Rows {
			var colName string
			switch p.adapter.GetDatabaseType() {
			case "MySQL":
				if field, ok := colRow["Field"].(string); ok {
					colName = field
				}
			case "SQLite":
				if name, ok := colRow["name"].(string); ok {
					colName = name
				}
			case "PostgreSQL":
				if name, ok := colRow["column_name"].(string); ok {
					colName = name
				}
			}

			if colName != "" {
				columns = append(columns, colName)
			}
		}

		tableInfo[tableName] = &TableInfo{
			Name:    tableName,
			Columns: columns,
		}
	}

	return tableInfo, nil
}

// buildBasicSchema builds a basic schema string from database metadata.
func (p *Pipeline) buildBasicSchema(ctx context.Context, tables []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Database: %s\n\n", p.adapter.GetDatabaseType()))

	for _, tableName := range tables {
		// Query table structure
		var query string
		switch p.adapter.GetDatabaseType() {
		case "MySQL":
			query = fmt.Sprintf("DESCRIBE %s", tableName)
		case "SQLite":
			query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		case "PostgreSQL":
			query = fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s'", tableName)
		default:
			continue
		}

		result, err := p.adapter.ExecuteQuery(ctx, query)
		if err != nil {
			continue
		}

		// Format table structure
		sb.WriteString(fmt.Sprintf("Table %s:\n", tableName))

		for _, row := range result.Rows {
			var colName, colType string

			// Extract column name and type based on database engine
			switch p.adapter.GetDatabaseType() {
			case "MySQL":
				if field, ok := row["Field"].(string); ok {
					colName = field
				}
				if typ, ok := row["Type"].(string); ok {
					colType = typ
				}
			case "SQLite":
				if name, ok := row["name"].(string); ok {
					colName = name
				}
				if typ, ok := row["type"].(string); ok {
					colType = typ
				}
			case "PostgreSQL":
				if name, ok := row["column_name"].(string); ok {
					colName = name
				}
				if typ, ok := row["data_type"].(string); ok {
					colType = typ
				}
			}

			if colName != "" {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", colName, colType))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
