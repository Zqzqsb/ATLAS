package inference

import (
	"context"
	"fmt"
	"strings"

	"atlas/internal/logger"
	"atlas/internal/react"
	"atlas/internal/react/scenarios"
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
	prompt := p.buildOneShotPrompt(query, contextPrompt)

	response, err := llmCallWithRetry(ctx, p.llm, prompt, 2)
	if err != nil {
		return "", err
	}

	sql := p.extractSQL(response)
	logger.L().Info("One-shot SQL generated", "sql", truncateLog(sql, 200))
	return sql, nil
}

// reactLoop runs the SQL generation ReAct loop using the unified react.Engine.
func (p *Pipeline) reactLoop(ctx context.Context, query string, contextPrompt string, result *Result) (string, error) {
	// Build step callback that adapts react.Step → inference.ReActStep
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

	// Build inference engine config via scenario
	engineCfg, sqlTool := scenarios.BuildInferenceEngine(p.adapter, scenarios.InferenceConfig{
		DBType:                  p.config.DBType,
		ContextPrompt:           contextPrompt,
		UseRichContext:          p.config.UseRichContext,
		MaxIterations:           p.config.MaxIterations,
		ClarifyMode:            p.config.ClarifyMode,
		ResultFields:            p.config.ResultFields,
		ResultFieldsDescription: p.config.ResultFieldsDescription,
		StepCallback:            stepCB,
	})

	engine := react.New(p.llm, engineCfg)

	logger.L().Info("ReAct loop starting",
		"max_iterations", p.config.MaxIterations,
		"actual_max", engineCfg.ActualMaxOverride,
	)

	// Execute with the question as input
	engineResult, err := engine.Execute(ctx, fmt.Sprintf("Question: %s", query))
	if err != nil {
		return "", fmt.Errorf("ReAct loop failed: %w", err)
	}

	logger.L().Info("ReAct loop completed",
		"iterations", engineResult.Iterations,
		"duration", engineResult.Duration,
	)

	// Collect ReAct steps
	for _, step := range engineResult.Steps {
		result.ReActSteps = append(result.ReActSteps, ReActStep{
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Phase:       "sql_generation",
		})
	}

	// Update stats
	result.LLMCalls += engineResult.Iterations
	result.SQLExecutions += sqlTool.ExecutionCount

	// Extract SQL from Final Answer
	sql := p.extractSQL(engineResult.Output)
	if sql == "" {
		return "", fmt.Errorf("no SQL generated")
	}
	return sql, nil
}

// buildOneShotPrompt builds the prompt for one-shot SQL generation.
func (p *Pipeline) buildOneShotPrompt(query string, contextPrompt string) string {
	var sb strings.Builder

	sb.WriteString("You are a SQL expert. Generate SQL to answer the question.\n\n")

	if p.config.DBType != "" {
		sb.WriteString(fmt.Sprintf("**Database Type: %s**\n", p.config.DBType))
		sb.WriteString(fmt.Sprintf("CRITICAL: Write SQL that strictly follows %s syntax rules.\n", p.config.DBType))
		sb.WriteString("Common syntax notes:\n")
		sb.WriteString("- Use backticks for identifiers, single quotes for strings\n")
		sb.WriteString("- LIMIT syntax: LIMIT offset, count\n")
		sb.WriteString("- Use CONCAT() for string concatenation\n")
		sb.WriteString("\n")
	}

	if contextPrompt != "" {
		sb.WriteString("Database Schema:\n")
		sb.WriteString(contextPrompt)
		sb.WriteString("\n\n")
	}

	if p.config.UseRichContext {
		sb.WriteString(richContextBestPractices)
	}

	sb.WriteString(fmt.Sprintf("Question: %s\n\n", query))

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

	sb.WriteString(`Task: Generate SQL directly.
Output ONLY the SQL query (no explanations, no markdown).

Format:
SELECT ...`)

	return sb.String()
}

// richContextBestPractices is the SQL best practices prompt fragment for Rich Context mode.
const richContextBestPractices = `IMPORTANT: Rich Context may be outdated or incorrect. When Rich Context conflicts with actual database data, trust the database.

SQL Best Practices:
1. TEXT fields storing numbers: Use CAST(field AS INTEGER/REAL) for comparisons and sorting
2. NULL handling:
   - NULL means "unknown/uncertain", not zero.
   - When doing aggregations on numeric data stored in TEXT fields (like 'MPG' or 'Horsepower'), be aware of non-numeric string values like 'null'.
   - Filter both SQL NULLs and string NULLs: WHERE field IS NOT NULL AND field != 'null'
3. String matching:
   - Use exact values from Rich Context when available (e.g., if Rich Context lists "USA, UK, France", use these exact strings)
   - If no exact values in Rich Context: use case-insensitive matching (LOWER(field) = LOWER('value'))
4. Duplicates: When the question asks for a list of items (e.g., names, cities), duplicates are often undesirable. Consider using DISTINCT.
5. Zero values:
   - Zero (0) means "business non-existence" (e.g., population=0 means no people)
   - Zero is different from NULL (NULL = unknown, 0 = known to be zero)
6. Extreme values (MIN/MAX/TOP/LIMIT):
   - ALWAYS return ALL rows with the extreme value (handle ties properly)
   - Use subquery pattern: WHERE column = (SELECT MIN/MAX(column) FROM table)
   - AVOID: ORDER BY ... LIMIT 1 (only returns one arbitrary row when there are ties)
   - Exception: If the question explicitly asks for "one" or "any one", then LIMIT 1 is acceptable
7. Value Mapping: Verify which column contains a specific text value before using it in WHERE.
8. Data format conflicts: Always verify actual data format with execute_sql when encountering unexpected empty results.
9. Whitespace: Use TRIM() for TEXT field comparisons when suspecting formatting issues.

`

// extractSQL extracts SQL from LLM response.
func (p *Pipeline) extractSQL(response string) string {
	if idx := strings.Index(response, "Final Answer:"); idx >= 0 {
		response = response[idx+13:]
	}

	response = strings.TrimSpace(response)

	// Remove markdown code fences
	response = strings.TrimPrefix(response, "```sql")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Extract backtick-wrapped SQL
	if strings.Contains(response, "`SELECT") || strings.Contains(response, "`select") {
		start := strings.Index(response, "`")
		if start >= 0 {
			end := strings.Index(response[start+1:], "`")
			if end >= 0 {
				response = response[start+1 : start+1+end]
			}
		}
	}

	// Multi-line SQL: stop at explanatory text
	lines := strings.Split(response, "\n")
	if len(lines) > 1 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(strings.ToUpper(firstLine), "SELECT") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "WITH") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "INSERT") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "UPDATE") ||
			strings.HasPrefix(strings.ToUpper(firstLine), "DELETE") {
			var sqlLines []string
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
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
