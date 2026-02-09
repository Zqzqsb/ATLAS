package inference

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"lucid/interfaces"
	contextpkg "lucid/internal/context"
)

// SchemaLinker Schema Linking 模块接口
type SchemaLinker interface {
	// Link 执行 Schema Linking
	// 输入: query, 所有表的信息
	// 输出: 相关表名列表, ReAct 步骤（如果使用 ReAct 模式）
	Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error)
}

// TableInfo 表的简要信息（用于 Schema Linking）
type TableInfo struct {
	Name        string
	Columns     []string                        // 列名列表
	ForeignKeys []contextpkg.ForeignKeyMetadata // 外键关系
	Description string                          // 表描述（可选，来自 rich_context 或表注释）
}

// LLMSchemaLinker 基于 LLM 的 Schema Linking
type LLMSchemaLinker struct {
	llm           llms.Model
	adapter       interfaces.DBAdapter
	useReact      bool
	tokenRecorder func(prompt, response string)
}

// NewLLMSchemaLinker 创建 LLM Schema Linker
func NewLLMSchemaLinker(llm llms.Model, dbAdapter interfaces.DBAdapter, useReact bool) *LLMSchemaLinker {
	return &LLMSchemaLinker{
		llm:      llm,
		adapter:  dbAdapter,
		useReact: useReact,
	}
}

// Link 执行 Schema Linking
func (l *LLMSchemaLinker) Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	if l.useReact {
		return l.linkWithReact(ctx, query, allTables)
	}
	return l.linkOneShot(ctx, query, allTables)
}

// linkOneShot One-shot Schema Linking
func (l *LLMSchemaLinker) linkOneShot(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	// 构建表信息描述（格式化为易读的列表）
	var schemaDesc strings.Builder
	for _, table := range allTables {
		schemaDesc.WriteString(fmt.Sprintf("- %s\n", table.Name))
		schemaDesc.WriteString(fmt.Sprintf("  Columns: %s\n", strings.Join(table.Columns, ", ")))
		if table.Description != "" {
			schemaDesc.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
		}
		schemaDesc.WriteString("\n")
	}

	// 构建 Prompt
	prompt := fmt.Sprintf(`You are a database expert. Identify which tables are relevant to answer the question.

Available Tables:
%s

Question: %s

Task: Select the minimum set of tables needed to answer this question.
Output format: table1, table2, table3 (comma-separated, no extra text)
If all tables are needed, output: all
If no tables are needed, output: none

Output:`, schemaDesc.String(), query)

	// 不打印完整的 Schema Linking prompt
	fmt.Println("🔍 Schema Linking...")

	// 调用 LLM，带退避重试机制
	var response string
	var err error
	maxRetries := 2
	backoffDelays := []time.Duration{1 * time.Second, 3 * time.Second}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err = l.llm.Call(ctx, prompt)
		if err == nil {
			break
		}

		// 如果还有重试机会，等待后重试
		if attempt < maxRetries {
			delay := backoffDelays[attempt]
			fmt.Printf("⚠️  Schema Linking failed (attempt %d/%d): %v\n", attempt+1, maxRetries+1, err)
			fmt.Printf("⏳ Retrying after %v...\n\n", delay)
			time.Sleep(delay)
		}
	}

	if err != nil {
		return nil, []ReActStep{}, fmt.Errorf("schema linking failed after %d attempts: %w", maxRetries+1, err)
	}

	response = strings.TrimSpace(response)

	// 记录 tokens
	if l.tokenRecorder != nil {
		l.tokenRecorder(prompt, response)
	}

	// 解析响应
	if response == "all" {
		result := make([]string, 0, len(allTables))
		for name := range allTables {
			result = append(result, name)
		}
		// 创建一个简单的步骤来表示 Schema Linking 过程
		tablesStr := strings.Join(result, ", ")
		steps := []ReActStep{
			{
				Thought: fmt.Sprintf("The question '%s' requires all tables to answer.", query),
				Action:  "final_answer",
				ActionInput: map[string]interface{}{
					"tables": tablesStr,
				},
				Observation: fmt.Sprintf("Selected tables: %s", tablesStr),
				Phase:       "schema_linking",
			},
		}
		return result, steps, nil
	}

	if response == "none" {
		// 创建一个简单的步骤来表示 Schema Linking 过程
		steps := []ReActStep{
			{
				Thought: fmt.Sprintf("The question '%s' does not require any tables to answer.", query),
				Action:  "final_answer",
				ActionInput: map[string]interface{}{
					"tables": "none",
				},
				Observation: "No tables needed",
				Phase:       "schema_linking",
			},
		}
		return []string{}, steps, nil
	}

	// 只取第一行（LLM 可能会返回额外的 Explanation）
	lines := strings.Split(response, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// 解析表名列表
	tables := strings.Split(firstLine, ",")
	result := make([]string, 0, len(tables))
	for _, table := range tables {
		table = strings.TrimSpace(table)
		if table != "" {
			result = append(result, table)
		}
	}

	// 创建一个简单的步骤来表示 Schema Linking 过程
	tablesStr := strings.Join(result, ", ")
	steps := []ReActStep{
		{
			Thought: fmt.Sprintf("Analyzed the question '%s' and identified relevant tables based on their columns and descriptions.", query),
			Action:  "final_answer",
			ActionInput: map[string]interface{}{
				"tables": tablesStr,
			},
			Observation: fmt.Sprintf("Selected tables: %s", tablesStr),
			Phase:       "schema_linking",
		},
	}

	return result, steps, nil
}

// linkWithReact ReAct 模式 Schema Linking
func (l *LLMSchemaLinker) linkWithReact(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	fmt.Println("🔍 Schema Linking (ReAct mode)...")

	// 创建 SQL 工具
	sqlTool := &SQLTool{
		adapter:   l.adapter,
		useDryRun: false,
	}

	// Create handler to collect ReAct steps
	reactHandler := &PrettyReActHandler{logMode: "simple"}

	// 创建 ReAct Agent
	// 策略：告诉模型最大 5 次迭代（制造紧迫感），实际设置 15 次（保证足够空间）
	actualMaxIterations := 15
	claimedMaxIterations := 5

	executor, err := agents.Initialize(
		l.llm,
		[]tools.Tool{sqlTool},
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(actualMaxIterations),
		agents.WithCallbacksHandler(reactHandler),
	)
	if err != nil {
		return nil, []ReActStep{}, err
	}

	// 构建表信息
	var schemaDesc strings.Builder
	for _, table := range allTables {
		schemaDesc.WriteString(fmt.Sprintf("- %s\n", table.Name))
		schemaDesc.WriteString(fmt.Sprintf("  Columns: %s\n", strings.Join(table.Columns, ", ")))

		// 添加外键信息
		if len(table.ForeignKeys) > 0 {
			schemaDesc.WriteString("  Foreign Keys:\n")
			for _, fk := range table.ForeignKeys {
				schemaDesc.WriteString(fmt.Sprintf("    %s → %s.%s\n", fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn))
			}
		}

		if table.Description != "" {
			schemaDesc.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
		}
		schemaDesc.WriteString("\n")
	}

	// 构建 Prompt
	prompt := fmt.Sprintf(`You are a database expert. Identify which tables are relevant to answer the question.

⚠️  ITERATION LIMIT: You have maximum %d iterations to complete this task. Be efficient!

Available Tables:
%s

Question: %s

Foreign key relationships are shown above. Use them to:
1. Identify direct relationships between tables
2. Find intermediate junction tables for many-to-many relationships
3. Trace the join path from source to target tables

You can use execute_sql to:
- Verify data existence: SELECT COUNT(*) FROM table
- Check join validity: SELECT COUNT(*) FROM t1 JOIN t2 ON ...
- Explore sample data: SELECT * FROM table LIMIT 3
- Check column values: SELECT DISTINCT column FROM table LIMIT 5

Workflow:
1. Identify tables with columns that seem relevant to the question.
2. Use the foreign key relationships to find all necessary tables for joins.
3. If you are unsure about a table's relevance, use 'execute_sql' to sample its data.
4. Provide the final list of tables.

Output Format:
A) Use tool to explore:
   Thought: [reasoning]
   Action: execute_sql
   Action Input: [SQL query]

B) Give final answer:
   Thought: [reasoning]
   Final Answer: table1, table2, table3

IMPORTANT:
- Output comma-separated table names only in Final Answer
- Include ALL tables needed for joins (don't miss intermediate tables)
- For NOT queries, include base table
- For foreign key columns, include referenced tables
- If all tables needed, output: all
- If no tables needed, output: none

Output:`, claimedMaxIterations, schemaDesc.String(), query)

	// 执行 ReAct
	agentResult, err := executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return nil, []ReActStep{}, err
	}

	// Collect ReAct steps from handler
	collectedSteps := reactHandler.GetCollectedSteps()
	schemaLinkingSteps := make([]ReActStep, 0, len(collectedSteps))
	for _, step := range collectedSteps {
		schemaLinkingSteps = append(schemaLinkingSteps, ReActStep{
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Phase:       "schema_linking",
		})
	}

	// 提取最终结果
	if output, ok := agentResult["output"].(string); ok {
		// 只取第一行（LLM 可能会返回额外的 Explanation）
		lines := strings.Split(output, "\n")
		firstLine := strings.TrimSpace(lines[0])

		if firstLine == "all" {
			result := make([]string, 0, len(allTables))
			for name := range allTables {
				result = append(result, name)
			}
			return result, schemaLinkingSteps, nil
		}

		if firstLine == "none" {
			return []string{}, schemaLinkingSteps, nil
		}

		tables := strings.Split(firstLine, ",")
		result := make([]string, 0, len(tables))
		for _, table := range tables {
			table = strings.TrimSpace(table)
			if table != "" {
				result = append(result, table)
			}
		}
		return result, schemaLinkingSteps, nil
	}

	return nil, []ReActStep{}, fmt.Errorf("schema linking failed to produce a valid table list")
}

// ExtractTableInfo 从 Rich Context 提取表信息
func ExtractTableInfo(ctx *contextpkg.SharedContext) map[string]*TableInfo {
	result := make(map[string]*TableInfo)

	for name, table := range ctx.Tables {
		columns := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columns[i] = col.Name
		}

		// 优先使用 LLM 生成的 Description
		description := table.Description
		if description == "" {
			// 备选：使用表注释
			description = table.Comment
		}
		if description == "" && len(table.RichContext) > 0 {
			// 最后备选：使用第一个 rich_context 条目的内容
			for _, v := range table.RichContext {
				description = v.Content
				break
			}
		}

		result[name] = &TableInfo{
			Name:        name,
			Columns:     columns,
			ForeignKeys: table.ForeignKeys,
			Description: description,
		}
	}

	return result
}
