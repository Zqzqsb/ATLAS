package inference

import (
	"context"
	"fmt"
	"strings"

	"lucid/internal/adapter"
	"lucid/internal/logger"
	"lucid/internal/react"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/tools"
)

// truncateLog truncates a string for log output.
func truncateLog(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// oneShotGeneration performs one-shot SQL generation.
func (p *Pipeline) oneShotGeneration(ctx context.Context, query string, contextPrompt string) (string, error) {
	prompt := p.buildPrompt(query, contextPrompt, false)

	response, err := llmCallWithRetry(ctx, p.llm, prompt, 2)
	if err != nil {
		return "", err
	}

	sql := p.extractSQL(response)
	logger.L().Info("One-shot SQL generated", "sql", truncateLog(sql, 200))
	return sql, nil
}

// reactLoop ReAct 循环
func (p *Pipeline) reactLoop(ctx context.Context, query string, contextPrompt string, result *Result) (string, error) {
	// Create handler to collect ReAct steps
	var stepCB react.StepCallback
	if p.stepCallback != nil {
		stepCB = func(step react.Step, eventType string) {
			p.stepCallback(ReActStep{
				Step:        step.Iteration,
				Thought:     step.Thought,
				Action:      step.Action,
				ActionInput: step.ActionInput,
				Observation: step.Observation,
				Phase:       "sql_generation",
			}, eventType)
		}
	}
	reactHandler := react.NewHandler(stepCB)

	// Tool observation callback — bypasses langchaingo's broken HandleToolEnd.
	// Directly pushes the observation into the handler and fires the SSE event.
	toolObsCB := ToolObservationCallback(func(toolName, output string) {
		reactHandler.InjectObservation(output)
	})

	// 创建工具
	sqlTool := &SQLTool{
		adapter:   p.adapter,
		useDryRun: p.config.UseDryRun,
		OnResult:  toolObsCB,
	}
	verifySQLTool := NewVerifySQLTool(p.adapter, p.config.DBType)
	verifySQLTool.OnResult = toolObsCB

	toolsList := []tools.Tool{sqlTool, verifySQLTool}

	// Actual iterations = claimed + small buffer for error recovery.
	// No 4x inflation — agent should operate within the budget it sees in the prompt.
	claimedMaxIterations := p.config.MaxIterations
	actualMaxIterations := claimedMaxIterations + 3 // small buffer for verify_sql retry

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

	logger.L().Info("ReAct loop starting", "max_iterations", actualMaxIterations, "claimed", claimedMaxIterations)

	agentResult, err := executor.Call(ctx, map[string]any{"input": prompt})
	if err != nil {
		return "", fmt.Errorf("ReAct loop failed: %w", err)
	}

	logger.L().Info("ReAct loop completed")

	// Collect ReAct steps from handler
	for _, step := range reactHandler.GetSteps() {
		result.ReActSteps = append(result.ReActSteps, ReActStep{
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Phase:       "sql_generation",
		})
	}

	// 更新统计信息
	result.LLMCalls += len(reactHandler.GetSteps()) // Use actual iteration count
	result.SQLExecutions += sqlTool.ExecutionCount

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
		sb.WriteString("Common syntax notes:\n")
		sb.WriteString("- Use backticks for identifiers, single quotes for strings\n")
		sb.WriteString("- LIMIT syntax: LIMIT offset, count\n")
		sb.WriteString("- Use CONCAT() for string concatenation\n")
		sb.WriteString("\n")
	}

	// Rich Context
	if contextPrompt != "" {
		sb.WriteString("Database Schema:\n")
		sb.WriteString(contextPrompt)
		sb.WriteString("\n\n")
	}

	// SQL Best Practices (only in Rich Context mode)
	if p.config.UseRichContext {
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
- execute_sql: Execute SQL and see results
- verify_sql: Validate SQL via EXPLAIN before giving Final Answer
  → Returns: ✅ VERIFY_PASSED or ❌ VERIFY_FAILED, plus EXPLAIN execution plan and performance warnings

Workflow:
1. Analyze question and schema
2. If string values missing from Rich Context → use execute_sql to find them
3. Write SQL following best practices
4. ALWAYS call verify_sql to validate your SQL and inspect the execution plan
   - If ❌ FAILED: fix the SQL error and call verify_sql AGAIN. Repeat until it passes.
   - NEVER give Final Answer with SQL that has not passed verify_sql.
   - If ✅ PASSED with no warnings: proceed to Final Answer
   - If ✅ PASSED with performance warnings: evaluate the warnings:
     * Full table scan on small tables (≤1000 rows) is acceptable
     * Full table scan on large tables (>1000 rows): try to optimize (add WHERE, use indexed columns)
     * If optimization is not feasible, proceed to Final Answer with the current SQL
5. Provide Final Answer — the SQL in your Final Answer MUST be the one that passed verify_sql

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
		sb.WriteString(fmt.Sprintf(`Critical Rules:
1. Field Order: SELECT fields MUST match expected order exactly (no table prefixes)
2. Iterations: %d max. Track: "Iteration X/%d"
3. Efficiency: Only use execute_sql when truly uncertain about data values.
4. ALWAYS verify: Call verify_sql before Final Answer. Review the EXPLAIN plan — optimize if large table scans are avoidable, otherwise accept.
5. Final Answer: SQL only, no explanations

`, p.config.MaxIterations, p.config.MaxIterations))

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
	adapter        adapter.DBAdapter
	useDryRun      bool
	ExecutionCount int
	OnResult       ToolObservationCallback
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
	sql := stripCodeFence(strings.TrimSpace(input))

	if t.useDryRun {
		if _, err := t.adapter.DryRunSQL(ctx, sql); err != nil {
			out := fmt.Sprintf("SQL validation failed: %v", err)
			t.emitResult(out)
			return out, nil
		}
	}

	result, err := t.adapter.ExecuteQuery(ctx, sql)
	if err != nil {
		out := fmt.Sprintf("SQL execution failed: %v", err)
		t.emitResult(out)
		return out, nil
	}

	output := fmt.Sprintf("Query executed successfully!\nRows: %d\n", result.RowCount)
	if result.RowCount > 0 {
		sampleStr := fmt.Sprintf("%v", result.Rows)
		const maxSampleLength = 1000
		if len(sampleStr) <= maxSampleLength {
			output += fmt.Sprintf("Sample results: %s\n", sampleStr)
		} else {
			output += fmt.Sprintf("Sample results: %s... (truncated)\n", sampleStr[:maxSampleLength])
		}
	}

	t.emitResult(output)
	return output, nil
}

func (t *SQLTool) emitResult(output string) {
	if t.OnResult != nil {
		t.OnResult("execute_sql", output)
	}
}

