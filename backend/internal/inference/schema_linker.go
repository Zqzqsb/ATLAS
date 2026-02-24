package inference

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/react"
	"lucid/internal/react/scenarios"
)

// SchemaLinker defines the interface for identifying relevant tables.
type SchemaLinker interface {
	Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error)
}

// TableInfo holds table information for Schema Linking.
type TableInfo struct {
	Name        string
	Columns     []string
	ForeignKeys []ForeignKeyRef
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

// Link performs Schema Linking.
func (l *LLMSchemaLinker) Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	if l.useReact {
		return l.linkWithReact(ctx, query, allTables)
	}
	return l.linkOneShot(ctx, query, allTables)
}

// linkOneShot performs one-shot Schema Linking.
func (l *LLMSchemaLinker) linkOneShot(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	var schemaDesc strings.Builder
	for _, table := range allTables {
		schemaDesc.WriteString(fmt.Sprintf("- %s\n", table.Name))
		schemaDesc.WriteString(fmt.Sprintf("  Columns: %s\n", strings.Join(table.Columns, ", ")))
		if table.Description != "" {
			schemaDesc.WriteString(fmt.Sprintf("  Description: %s\n", table.Description))
		}
		schemaDesc.WriteString("\n")
	}

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

	if response == "all" {
		result := make([]string, 0, len(allTables))
		for name := range allTables {
			result = append(result, name)
		}
		tablesStr := strings.Join(result, ", ")
		steps := []ReActStep{{
			Thought:     fmt.Sprintf("The question '%s' requires all tables to answer.", query),
			Action:      "final_answer",
			ActionInput: map[string]interface{}{"tables": tablesStr},
			Observation: fmt.Sprintf("Selected tables: %s", tablesStr),
			Phase:       "schema_linking",
		}}
		return result, steps, nil
	}

	if response == "none" {
		steps := []ReActStep{{
			Thought:     fmt.Sprintf("The question '%s' does not require any tables to answer.", query),
			Action:      "final_answer",
			ActionInput: map[string]interface{}{"tables": "none"},
			Observation: "No tables needed",
			Phase:       "schema_linking",
		}}
		return []string{}, steps, nil
	}

	lines := strings.Split(response, "\n")
	firstLine := strings.TrimSpace(lines[0])

	tables := strings.Split(firstLine, ",")
	result := make([]string, 0, len(tables))
	for _, table := range tables {
		table = strings.TrimSpace(table)
		if table != "" {
			result = append(result, table)
		}
	}

	tablesStr := strings.Join(result, ", ")
	steps := []ReActStep{{
		Thought:     fmt.Sprintf("Analyzed the question '%s' and identified relevant tables based on their columns and descriptions.", query),
		Action:      "final_answer",
		ActionInput: map[string]interface{}{"tables": tablesStr},
		Observation: fmt.Sprintf("Selected tables: %s", tablesStr),
		Phase:       "schema_linking",
	}}

	return result, steps, nil
}

// linkWithReact performs ReAct-mode Schema Linking using the unified react.Engine.
func (l *LLMSchemaLinker) linkWithReact(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	// Build schema description
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

	// Build engine config via scenario
	engineCfg := scenarios.BuildSchemaLinkingEngine(scenarios.SchemaLinkingConfig{
		DBAdapter:  l.adapter,
		SchemaDesc: schemaDesc.String(),
		Query:      query,
	})

	engine := react.New(l.llm, engineCfg)

	// Execute
	engineResult, err := engine.Execute(ctx, "")
	if err != nil {
		return nil, []ReActStep{}, err
	}

	// Convert react.Steps → inference.ReActSteps
	schemaLinkingSteps := make([]ReActStep, 0, len(engineResult.Steps))
	for _, step := range engineResult.Steps {
		schemaLinkingSteps = append(schemaLinkingSteps, ReActStep{
			Thought:     step.Thought,
			Action:      step.Action,
			ActionInput: step.ActionInput,
			Observation: step.Observation,
			Phase:       "schema_linking",
		})
	}

	// Parse final answer
	output := engineResult.Output
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

	if len(result) == 0 {
		return nil, schemaLinkingSteps, fmt.Errorf("schema linking failed to produce a valid table list")
	}
	return result, schemaLinkingSteps, nil
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
