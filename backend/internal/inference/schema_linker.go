package inference

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"lucid/internal/adapter"
)

// SchemaLinker defines the interface for identifying relevant tables.
type SchemaLinker interface {
	Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error)
}

// TableInfo holds table information for Schema Linking.
type TableInfo struct {
	Name        string
	Columns     []string
	ForeignKeys []ForeignKeyRef // Uses inference-local type
	Description string
}

// LLMSchemaLinker performs LLM-based Schema Linking.
type LLMSchemaLinker struct {
	llm      llms.Model
	adapter  adapter.DBAdapter
	useReact bool
}

// NewLLMSchemaLinker creates a new LLM-based schema linker.
func NewLLMSchemaLinker(llm llms.Model, dbAdapter adapter.DBAdapter, useReact bool) *LLMSchemaLinker {
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


	response, err := llmCallWithRetry(ctx, l.llm, prompt, 2)
	if err != nil {
		return nil, []ReActStep{}, fmt.Errorf("schema linking failed: %w", err)
	}
	response = strings.TrimSpace(response)

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

	// 创建 SQL 工具
	sqlTool := &SQLTool{
		adapter:   l.adapter,
		useDryRun: false,
	}

	// Create handler to collect ReAct steps
	reactHandler := &PrettyReActHandler{}

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

// llmCallWithRetry calls the LLM with exponential backoff retry.
func llmCallWithRetry(ctx context.Context, model llms.Model, prompt string, maxRetries int) (string, error) {
	backoff := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := model.Call(ctx, prompt)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if attempt < maxRetries && attempt < len(backoff) {
			time.Sleep(backoff[attempt])
		}
	}
	return "", fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries+1, lastErr)
}
