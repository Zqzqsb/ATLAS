package inference

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lucid/interfaces"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/tools"
)

// oneShotGeneration One-shot SQL 生成
func (p *Pipeline) oneShotGeneration(ctx context.Context, query string, contextPrompt string) (string, error) {
	prompt := p.buildPrompt(query, contextPrompt, false)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(" SQL Generation (One-shot) - Prompt to LLM:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(prompt)
	fmt.Println()

	// 调用 LLM，带退避重试机制
	var response string
	var err error
	maxRetries := 2
	backoffDelays := []time.Duration{1 * time.Second, 3 * time.Second}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err = p.llm.Call(ctx, prompt)
		if err == nil {
			break
		}

		// 如果还有重试机会，等待后重试
		if attempt < maxRetries {
			delay := backoffDelays[attempt]
			fmt.Printf("⚠️  SQL Generation failed (attempt %d/%d): %v\n", attempt+1, maxRetries+1, err)
			fmt.Printf("⏳ Retrying after %v...\n\n", delay)
			time.Sleep(delay)
		}
	}

	if err != nil {
		return "", fmt.Errorf("LLM call failed after %d attempts: %w", maxRetries+1, err)
	}

	// 记录 tokens
	p.promptTexts = append(p.promptTexts, prompt)
	p.responseTexts = append(p.responseTexts, response)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("💡 SQL Generation - LLM Response:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(response)
	fmt.Println()

	// 提取 SQL
	sql := p.extractSQL(response)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(" Extracted SQL:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(sql)
	fmt.Println()

	return sql, nil
}

// reactLoop ReAct 循环
func (p *Pipeline) reactLoop(ctx context.Context, query string, contextPrompt string, result *Result) (string, error) {
	// 创建工具
	sqlTool := &SQLTool{
		adapter:   p.adapter,
		useDryRun: p.config.UseDryRun,
	}

	clarifyTool := &ClarifyTool{
		resultFields:            p.config.ResultFields,
		resultFieldsDescription: p.config.ResultFieldsDescription,
	}

	// 创建 verify_sql 工具
	verifySQLTool := NewVerifySQLTool(p.adapter, p.config.DBType)

	// 创建 ReAct Agent
	var toolsList []tools.Tool
	toolsList = []tools.Tool{sqlTool, verifySQLTool}

	if p.config.ClarifyMode == "on" {
		toolsList = append(toolsList, clarifyTool)
	}

	if p.config.EnableProofread {
		updateTool := NewUpdateRichContextTool(p.config.DBName, p.config.DBType)
		toolsList = append(toolsList, updateTool)
	}

	// Create handler to collect ReAct steps
	reactHandler := &PrettyReActHandler{logMode: p.config.LogMode}

	// Set up streaming callback if available (for real-time step notifications)
	if p.stepCallback != nil {
		reactHandler.SetStepNotifier(func(step CollectedStep, eventType string) {
			p.stepCallback(ReActStep{
				Step:        step.Step,
				Thought:     step.Thought,
				Action:      step.Action,
				ActionInput: step.ActionInput,
				Observation: step.Observation,
				Phase:       "sql_generation",
			}, eventType)
		})
	}

	// Use higher actual iterations than what we show in prompt
	// This gives the model more chances to complete while not overwhelming the prompt
	actualMaxIterations := p.config.MaxIterations * 4 // e.g., user sets 5, we allow 20
	claimedMaxIterations := p.config.MaxIterations     // what we tell the model

	executor, err := agents.Initialize(
		p.llm,
		toolsList,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(actualMaxIterations),
		agents.WithCallbacksHandler(reactHandler),
	)
	if err != nil {
		return "", err
	}

	// 构建 Prompt - pass claimed iterations to prompt
	prompt := p.buildPrompt(query, contextPrompt, true)

	// 只打印关键信息，不打印完整 prompt（避免重复的 Best Practices 等）
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🔄 Starting ReAct Loop (Claimed %d, Actual Max %d iterations)\n", claimedMaxIterations, actualMaxIterations)
	fmt.Printf("Question: %s\n", query)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	agentResult, err := executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		fmt.Printf("\n❌ ReAct Loop failed: %v\n\n", err)
		return "", err
	}

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ ReAct Loop completed successfully")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Collect ReAct steps from handler
	collectedSteps := reactHandler.GetCollectedSteps()
	for _, step := range collectedSteps {
		result.ReActSteps = append(result.ReActSteps, ReActStep{
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Phase:       "sql_generation",
		})
	}

	// 更新统计信息
	result.LLMCalls += len(collectedSteps) // Use actual iteration count
	result.SQLExecutions += sqlTool.ExecutionCount
	result.ClarifyCount = clarifyTool.ClarifyCount

	// 提取最终 SQL
	if output, ok := agentResult["output"].(string); ok {
		sql := p.extractSQL(output)
		return sql, nil
	}

	return "", fmt.Errorf("no SQL generated")
}

// buildPrompt 构建 Prompt
func (p *Pipeline) buildPrompt(query string, contextPrompt string, isReact bool) string {
	var sb strings.Builder

	sb.WriteString("You are a SQL expert. Generate SQL to answer the question.\n\n")

	// 数据库类型信息
	if p.config.DBType != "" {
		sb.WriteString(fmt.Sprintf("**Database Type: %s**\n", p.config.DBType))
		sb.WriteString(fmt.Sprintf("CRITICAL: Write SQL that strictly follows %s syntax rules.\n", p.config.DBType))
		sb.WriteString("Common syntax differences to watch:\n")
		switch p.config.DBType {
		case "SQLite":
			sb.WriteString("- Use double quotes for identifiers if needed, single quotes for strings\n")
			sb.WriteString("- No LIMIT offset without LIMIT clause\n")
			sb.WriteString("- Use || for string concatenation\n")
		case "MySQL":
			sb.WriteString("- Use backticks for identifiers, single quotes for strings\n")
			sb.WriteString("- LIMIT syntax: LIMIT offset, count\n")
			sb.WriteString("- Use CONCAT() for string concatenation\n")
		case "PostgreSQL":
			sb.WriteString("- Use double quotes for identifiers, single quotes for strings\n")
			sb.WriteString("- LIMIT syntax: LIMIT count OFFSET offset\n")
			sb.WriteString("- Use || for string concatenation\n")
		}
		sb.WriteString("\n")
	}

	// Rich Context
	if contextPrompt != "" {
		sb.WriteString("Database Schema:\n")
		sb.WriteString(contextPrompt)
		sb.WriteString("\n\n")
	}

	// SQL Best Practices（仅在使用 Rich Context 时添加）
	// 这些是在 onboarding 阶段分析出的增强提示，不应该在 baseline 中使用
	if p.config.UseRichContext {
		// JOIN 路径和字段语义信息（仅在 Rich Context 模式下）
		if p.context != nil {
			if joinPathsPrompt := p.context.FormatJoinPathsForPrompt(); joinPathsPrompt != "" {
				sb.WriteString(joinPathsPrompt)
			}
			if fieldSemanticsPrompt := p.context.FormatFieldSemanticsForPrompt(); fieldSemanticsPrompt != "" {
				sb.WriteString(fieldSemanticsPrompt)
			}
		}

		sb.WriteString(`IMPORTANT: Rich Context may be outdated or incorrect. When Rich Context conflicts with actual database data, trust the database.

SQL Best Practices:
1. TEXT fields storing numbers: Use CAST(field AS INTEGER/REAL) for comparisons and sorting
2. NULL handling:
   - NULL means "unknown/uncertain", not zero.
   - When doing aggregations on numeric data stored in TEXT fields (like 'MPG' or 'Horsepower'), be aware of non-numeric string values like 'null'.
   - Filter both SQL NULLs and string NULLs: WHERE field IS NOT NULL AND field != 'null'
3. String matching:
   - Use exact values from Rich Context when available (e.g., if Rich Context lists "USA, UK, France", use these exact strings)
   - If no exact values in Rich Context and NOT in ReAct mode: use case-insensitive matching (LOWER(field) = LOWER('value'))
   - If no exact values in Rich Context and IN ReAct mode: explore with execute_sql to find exact values first
4. Duplicates: When the question asks for a list of items (e.g., names, cities), duplicates are often undesirable. If your query joins tables in a way that might create duplicates (e.g., one student has multiple pets), consider using DISTINCT to ensure unique results.
5. Zero values:
   - Zero (0) means "business non-existence" (e.g., population=0 means no people)
   - Zero is different from NULL (NULL = unknown, 0 = known to be zero)
   - Check Rich Context for specific meaning of zero in each field
6. Extreme values (MIN/MAX/TOP/LIMIT):
   - When finding extreme values (youngest, oldest, highest, lowest, etc.):
     * ALWAYS return ALL rows with the extreme value (handle ties properly)
     * Use subquery pattern: WHERE column = (SELECT MIN/MAX(column) FROM table)
     * Example: SELECT * FROM table WHERE value = (SELECT MAX(value) FROM table)
   - AVOID: ORDER BY ... LIMIT 1 (only returns one arbitrary row when there are ties)
   - Exception: If the question explicitly asks for "one" or "any one", then LIMIT 1 is acceptable
7. Value Mapping: When the question contains specific text values (e.g., "amc hornet sportabout (sw)"), you MUST verify which column contains this value before using it in a WHERE clause. DO NOT GUESS between similar columns (e.g., 'Make' vs 'Model'). Use 'execute_sql' with a 'WHERE' clause to check for the value's existence.
8. Data format conflicts:
   - If Rich Context says "2-digit year (70=1970)" but query returns 0 results, try 4-digit year (1970)
   - Always verify actual data format with execute_sql when encountering unexpected empty results
9. Data Formatting and Whitespace:
   - Be cautious of hidden characters or formatting that can cause 'WHERE' clause mismatches, especially in 'TEXT' fields.
   - **Leading/Trailing Spaces:** Values might have extra spaces (e.g., '' USA '' instead of ''USA''). Use 'TRIM()' (e.g., 'WHERE TRIM(Country) = ''USA''') to handle this.
   - **Special Characters:** Data might be enclosed in quotes or other characters (e.g., '''"France"''').
   - If a query with a 'WHERE' clause on a 'TEXT' field unexpectedly returns no results, suspect a formatting issue. Use 'execute_sql' with 'LIKE ''%value%''' to investigate the actual data format.

`)
	}

	sb.WriteString(fmt.Sprintf("Question: %s\n\n", query))

	// force 模式：强制在 prompt 中给出字段信息
	if p.config.ClarifyMode == "force" && len(p.config.ResultFields) > 0 {
		sb.WriteString("⚠️ REQUIRED OUTPUT FIELDS:\n")
		fieldsStr := strings.Join(p.config.ResultFields, ", ")
		sb.WriteString(fmt.Sprintf("Your SQL query MUST return EXACTLY these fields in this EXACT ORDER: %s\n", fieldsStr))
		if p.config.ResultFieldsDescription != "" {
			sb.WriteString(fmt.Sprintf("Field descriptions: %s\n", p.config.ResultFieldsDescription))
		}
		sb.WriteString("\nCRITICAL: Use these field names WITHOUT table prefixes (e.g., 'Name' not 'singer.Name').\n")
		sb.WriteString("Any deviation from this field list will be considered INCORRECT.\n\n")
	}

	if isReact {
		// Tools available
		sb.WriteString(`Available Tools:
- execute_sql: Execute SQL and see results`)
		if p.config.ClarifyMode == "on" {
			sb.WriteString(`
- clarify_fields: Ask which fields to return (when question doesn't specify)`)
		}
		if p.config.EnableProofread {
			sb.WriteString(`
- update_rich_context: Update expired/incorrect Rich Context`)
		}

		// Workflow
		sb.WriteString(`

Workflow:
1. Analyze question and schema`)
		if p.config.ClarifyMode == "on" {
			sb.WriteString(`
2. If unclear which columns needed → use clarify_fields
3. If string values missing from Rich Context → use execute_sql to find them`)
		} else {
			sb.WriteString(`
2. If string values missing from Rich Context → use execute_sql to find them`)
		}
		if p.config.EnableProofread {
			sb.WriteString(`
4. If Rich Context conflicts with actual data → use update_rich_context`)
		}
		sb.WriteString(`
5. Write SQL following best practices
6. If uncertain → validate with execute_sql (use LIMIT/COUNT for large results)
7. Provide Final Answer

`)

		// Output format
		sb.WriteString(`Output Format (choose ONE):
A) Use tool:
   Thought: [reasoning]
   Action: [tool_name]
   Action Input: [input]

B) Give answer:
   Thought: [reasoning]
   Final Answer: [SQL only, no markdown]

⚠️ NEVER write "Action: None"! If no tool needed, use option B.

`)

		// Critical rules
		sb.WriteString(`Critical Rules:
1. Field Order: SELECT fields MUST match expected order exactly (no table prefixes)
2. Iterations: 5 max (update_rich_context doesn't count). Track: "Iteration X/5"
3. Efficiency: Only use execute_sql when truly uncertain
4. No repetition: If stuck, try different approach
5. Final Answer: SQL only, no explanations

`)

		// 在 ReAct 模式下，再次强调字段要求（防止长程注意力丢失）
		if p.config.ClarifyMode == "force" && len(p.config.ResultFields) > 0 {
			sb.WriteString(`
⚠️ REMINDER - REQUIRED OUTPUT FIELDS ⚠️
Before Final Answer, verify your SQL returns these EXACT fields in EXACT order:
`)
			fieldsStr := strings.Join(p.config.ResultFields, ", ")
			sb.WriteString(fmt.Sprintf("Required: %s\n", fieldsStr))
			if p.config.ResultFieldsDescription != "" {
				sb.WriteString(fmt.Sprintf("(%s)\n", p.config.ResultFieldsDescription))
			}
			sb.WriteString(`If field is a name/description, JOIN the referenced table. Do NOT return IDs when names are required.
`)
		}

		if p.config.ClarifyMode == "on" {
			sb.WriteString(`
6. Clarify: Follow field names/descriptions from clarify_fields precisely
`)
		}
	} else {
		sb.WriteString(`Task: Generate SQL directly.
Output ONLY the SQL query (no explanations, no markdown).

Format:
SELECT ...`)
	}

	return sb.String()
}

// extractSQL 从响应中提取 SQL
func (p *Pipeline) extractSQL(response string) string {
	// 尝试提取 Final Answer
	if idx := strings.Index(response, "Final Answer:"); idx >= 0 {
		response = response[idx+13:]
	}

	// 清理
	response = strings.TrimSpace(response)

	// 移除 markdown 代码块
	response = strings.TrimPrefix(response, "```sql")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// 如果包含反引号包裹的 SQL，提取它
	if strings.Contains(response, "`SELECT") || strings.Contains(response, "`select") {
		start := strings.Index(response, "`")
		if start >= 0 {
			end := strings.Index(response[start+1:], "`")
			if end >= 0 {
				response = response[start+1 : start+1+end]
			}
		}
	}

	// 如果响应包含多行，且第一行是 SELECT，只取第一行
	lines := strings.Split(response, "\n")
	if len(lines) > 1 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(strings.ToUpper(firstLine), "SELECT") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "WITH") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "INSERT") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "UPDATE") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "DELETE") {
			// 找到 SQL 语句的结束位置（遇到非 SQL 内容）
			var sqlLines []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				// 如果遇到解释性文本（如 "This query"），停止
				if strings.HasPrefix(trimmed, "This ") ||
					strings.HasPrefix(trimmed, "The ") ||
					strings.HasPrefix(trimmed, "Since ") ||
					strings.HasPrefix(trimmed, "Note:") {
					break
				}
				sqlLines = append(sqlLines, line)
			}
			response = strings.Join(sqlLines, "\n")
		}
	}

	return strings.TrimSpace(response)
}

// SQLTool SQL 执行工具
type SQLTool struct {
	adapter        interfaces.DBAdapter
	useDryRun      bool
	ExecutionCount int
}

func (t *SQLTool) Name() string {
	return "execute_sql"
}

func (t *SQLTool) Description() string {
	if t.useDryRun {
		return `Execute SQL query with dry run validation first.
Input: SQL query string
Output: Query results or validation errors`
	}
	return `Execute SQL query and return results.
Input: SQL query string
Output: Query results`
}

func (t *SQLTool) Call(ctx context.Context, input string) (string, error) {
	t.ExecutionCount++

	fmt.Printf("\n🔧 Tool Call [execute_sql] #%d:\n", t.ExecutionCount)
	fmt.Printf("Input SQL: %s\n", input)

	sql := strings.TrimSpace(input)

	// Dry Run 验证（如果启用）
	if t.useDryRun {
		if err := t.adapter.DryRunSQL(ctx, sql); err != nil {
			return fmt.Sprintf("SQL validation failed: %v", err), nil
		}
	}

	// 执行 SQL
	result, err := t.adapter.ExecuteQuery(ctx, sql)
	if err != nil {
		return fmt.Sprintf("SQL execution failed: %v", err), nil
	}

	// 格式化结果
	output := fmt.Sprintf("Query executed successfully!\nRows: %d\n", result.RowCount)

	// 基于字符长度而非行数来决定是否显示 sample
	// 序列化结果并检查长度
	if result.RowCount > 0 {
		sampleStr := fmt.Sprintf("%v", result.Rows)
		const maxSampleLength = 1000 // 最大显示 1000 字符

		if len(sampleStr) <= maxSampleLength {
			// 完整显示
			output += fmt.Sprintf("Sample results: %s\n", sampleStr)
		} else {
			// 截断并添加省略标志
			truncated := sampleStr[:maxSampleLength]
			output += fmt.Sprintf("Sample results: %s... (truncated, showing first %d chars of %d total)\n",
				truncated, maxSampleLength, len(sampleStr))
		}
	}

	fmt.Printf("Output: %s\n", output)

	return output, nil
}

// ClarifyTool 澄清工具 - 用于询问需要返回哪些字段
type ClarifyTool struct {
	resultFields            []string
	resultFieldsDescription string
	ClarifyCount            int
}

func (t *ClarifyTool) Name() string {
	return "clarify_fields"
}

func (t *ClarifyTool) Description() string {
	return `Ask for clarification about which fields should be returned in the query result.
Use this when the question doesn't specify which columns to return.
Input: Your question about which fields to return (e.g., "Which fields should I return?")
Output: List of required fields or description of required fields`
}

func (t *ClarifyTool) Call(ctx context.Context, input string) (string, error) {
	t.ClarifyCount++

	fmt.Printf("\n🔔 Clarification requested: %s\n", input)

	// 统一返回字段列表 + 描述
	fieldsStr := strings.Join(t.resultFields, ", ")
	response := fmt.Sprintf("Required fields in EXACT ORDER: %s\n\nField descriptions: %s\n\nIMPORTANT: Use these field names WITHOUT table prefixes (e.g., 'Name' not 'singer.Name')",
		fieldsStr,
		t.resultFieldsDescription)

	fmt.Printf("📋 Clarification response: %s\n\n", response)

	return response, nil
}
