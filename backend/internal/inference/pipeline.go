package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	contextpkg "lucid/internal/context"
)

// Config 推理管线配置
type Config struct {
	UseRichContext bool
	UseReact       bool
	ReactLinking   bool // Schema Linking 是否使用 ReAct 模式
	UseDryRun      bool
	MaxIterations  int
	ContextFile    string

	// 澄清功能配置
	ClarifyMode             string   // 澄清模式: "off" (不启用) | "on" (agent主动询问) | "force" (强制给出)
	LogMode                 string   // 日志模式: "simple" (简洁) | "full" (完整输出所有交互)
	ResultFields            []string // 期望的结果字段列表
	ResultFieldsDescription string   // 结果字段的描述

	// 校对模式配置
	EnableProofread bool   // 是否启用校对模式（允许 LLM 修正 Rich Context）
	DBName          string // 数据库名称
	DBType          string // 数据库类型
}

// StepCallback is called for each ReAct step update during streaming
// eventType: "thought" | "action" | "observation" | "finish"
type StepCallback func(step ReActStep, eventType string)

// Pipeline 推理管线
type Pipeline struct {
	llm          llms.Model
	adapter      adapter.DBAdapter
	config       *Config
	context      *contextpkg.SharedContext
	schemaLinker SchemaLinker
	tokenizer    *tiktoken.Tiktoken

	// Token 统计累积器
	promptTexts   []string
	responseTexts []string

	// Streaming callback
	stepCallback StepCallback
}

// Result 推理结果
type Result struct {
	Query           string
	GeneratedSQL    string
	ExecutionResult interface{}

	// 统计信息
	TotalTime     time.Duration
	LLMCalls      int
	SQLExecutions int
	TotalTokens   int
	ClarifyCount  int // 澄清次数

	// 中间结果
	SelectedTables []string
	ReActSteps     []ReActStep
}

// ReActStep ReAct 步骤
type ReActStep struct {
	Step        int         `json:"step,omitempty"`              // Step number for streaming
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"` // 支持 string 和 map[string]interface{}
	Observation string      `json:"observation,omitempty"`
	Phase       string      `json:"phase,omitempty"` // "schema_linking" or "sql_generation"
}

// Reset 清理累积的统计数据，防止内存泄漏
func (p *Pipeline) Reset() {
	p.promptTexts = nil
	p.responseTexts = nil
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

// NewPipeline 创建推理管线
func NewPipeline(llm llms.Model, adapter adapter.DBAdapter, config *Config) *Pipeline {
	// 初始化 tokenizer (使用 cl100k_base，适用于 GPT-3.5/GPT-4/DeepSeek)
	tokenizer, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		// 如果失败，使用 nil，后续会跳过 token 统计
		tokenizer = nil
	}

	// Schema Linking 使用 ReAct 模式（由 ReactLinking 配置控制）
	linker := NewLLMSchemaLinker(llm, adapter, config.ReactLinking)

	p := &Pipeline{
		llm:          llm,
		adapter:      adapter,
		config:       config,
		schemaLinker: linker,
		tokenizer:    tokenizer,
	}

	// 设置 token recorder
	linker.tokenRecorder = func(prompt, response string) {
		p.promptTexts = append(p.promptTexts, prompt)
		p.responseTexts = append(p.responseTexts, response)
	}

	// 加载 Context 文件（如果提供）
	// 注意：context 总是加载用于 Schema Linking
	// UseRichContext 只控制是否在 SQL Generation 中使用 rich_context
	if config.ContextFile != "" {
		if ctx, err := p.loadContext(config.ContextFile); err == nil {
			p.context = ctx
		}
	}

	return p
}

// SetContext sets the shared context directly (alternative to loading from file)
func (p *Pipeline) SetContext(ctx *contextpkg.SharedContext) {
	p.context = ctx
}

// countTokens 统计文本的 token 数量
func (p *Pipeline) countTokens(text string) int {
	if p.tokenizer == nil {
		return 0
	}
	tokens := p.tokenizer.Encode(text, nil, nil)
	return len(tokens)
}

// Execute 执行推理
func (p *Pipeline) Execute(ctx context.Context, query string) (*Result, error) {
	startTime := time.Now()

	// 重置 token 统计累积器
	p.promptTexts = []string{}
	p.responseTexts = []string{}

	result := &Result{
		Query:      query,
		ReActSteps: []ReActStep{},
	}

	// 1. Schema Linking (总是执行，识别相关表)
	var allTableInfo map[string]*TableInfo
	var err error
	if p.context != nil {
		// 从 Rich Context 提取表信息
		allTableInfo = ExtractTableInfo(p.context)
	} else {
		// 从数据库查询表信息
		allTableInfo, err = p.extractTableInfoFromDB(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to extract table info: %w", err)
		}
	}

	// Notify that schema linking is starting
	p.notifyStep(ReActStep{
		Step:    0,
		Thought: "Starting Schema Linking to identify relevant tables...",
		Phase:   "schema_linking",
	}, "thought")

	tables, schemaLinkingSteps, err := p.schemaLinker.Link(ctx, query, allTableInfo)
	if err != nil {
		return nil, fmt.Errorf("schema linking failed: %w", err)
	}
	result.SelectedTables = tables
	result.LLMCalls++

	// Add Schema Linking ReAct steps to result and notify via streaming
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

		// Send streaming notification for each schema linking step
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

	// Notify schema linking completion with selected tables
	p.notifyStep(ReActStep{
		Step:        len(schemaLinkingSteps) + 1,
		Thought:     fmt.Sprintf("Schema Linking completed. Selected %d tables: %v", len(tables), tables),
		Observation: fmt.Sprintf("Selected tables: %s", strings.Join(tables, ", ")),
		Phase:       "schema_linking",
	}, "finish")

	fmt.Printf("📋 Selected Tables: %v\n\n", tables)

	// 2. 构建 Schema Context (基础表结构信息，总是提供)
	var contextPrompt string

	if p.config.UseRichContext && p.context != nil {
		// 使用 Rich Context (详细信息)
		opts := &contextpkg.ExportOptions{
			Tables:             tables,
			IncludeColumns:     true,
			IncludeIndexes:     true,
			IncludeRichContext: true,
			IncludeStats:       true,
		}
		contextPrompt = p.context.ExportToCompactPrompt(opts)
		// 不打印完整的 Rich Context，只打印简要信息
		fmt.Printf("📚 Using Rich Context for %d tables\n", len(tables))
	} else {
		// 使用基础 Schema (仅表名+列名)
		contextPrompt = p.buildBasicSchema(ctx, tables)
		// 不打印完整的 Basic Schema
		fmt.Printf("📋 Using Basic Schema for %d tables\n", len(tables))
	}

	// 3. Generate SQL
	var sql string
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

	// 4. 统计 tokens（从累积器中统计所有 prompts 和 responses）
	// 暂时禁用 token 统计，避免潜在的问题
	fmt.Printf("[DEBUG] Token counting disabled (would count %d prompts, %d responses)\n", len(p.promptTexts), len(p.responseTexts))
	result.TotalTokens = 0 // 暂时设为 0

	// if p.tokenizer != nil {
	// 	for i, prompt := range p.promptTexts {
	// 		fmt.Printf("[DEBUG] Counting prompt %d/%d (length: %d)\n", i+1, len(p.promptTexts), len(prompt))
	// 		result.TotalTokens += p.countTokens(prompt)
	// 	}
	// 	for i, response := range p.responseTexts {
	// 		fmt.Printf("[DEBUG] Counting response %d/%d (length: %d)\n", i+1, len(p.responseTexts), len(response))
	// 		result.TotalTokens += p.countTokens(response)
	// 	}
	// }

	// 5. Execute SQL (optional)
	if sql != "" {
		execResult, err := p.adapter.ExecuteQuery(ctx, sql)
		if err == nil {
			result.ExecutionResult = execResult
			result.SQLExecutions++
		}
	}

	return result, nil
}

// loadContext 加载 Rich Context
func (p *Pipeline) loadContext(path string) (*contextpkg.SharedContext, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ctx contextpkg.SharedContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// extractTableInfoFromDB 从数据库提取表信息
func (p *Pipeline) extractTableInfoFromDB(ctx context.Context) (map[string]*TableInfo, error) {
	// 获取所有表名
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

	// 对每个表查询列信息
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

		// 查询列信息
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

// buildBasicSchema 构建基础 Schema（从数据库查询表结构）
func (p *Pipeline) buildBasicSchema(ctx context.Context, tables []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Database: %s\n\n", p.adapter.GetDatabaseType()))

	for _, tableName := range tables {
		// 查询表结构
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

		// 格式化表结构
		sb.WriteString(fmt.Sprintf("Table %s:\n", tableName))

		for _, row := range result.Rows {
			var colName, colType string

			// 根据数据库类型提取列名和类型
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
