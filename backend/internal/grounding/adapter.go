package grounding

import (
	"context"
	"fmt"
)

// SchemaLinkerAdapter adapts the Semantic Grounding service to the SchemaLinker interface
// This allows the grounding service to be used in the existing inference pipeline
type SchemaLinkerAdapter struct {
	service *Service
	mode    GroundingMode
}

// TableInfo represents table information for schema linking
// This mirrors the inference.TableInfo to avoid circular imports
type TableInfo struct {
	Name        string
	Columns     []string
	Description string
}

// ReActStep represents a ReAct reasoning step
// This mirrors inference.ReActStep
type ReActStep struct {
	Step        int         `json:"step,omitempty"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Phase       string      `json:"phase,omitempty"`
}

// NewSchemaLinkerAdapter creates a new adapter
func NewSchemaLinkerAdapter(service *Service, mode GroundingMode) *SchemaLinkerAdapter {
	return &SchemaLinkerAdapter{
		service: service,
		mode:    mode,
	}
}

// Link performs semantic grounding and returns selected tables
// This implements the SchemaLinker interface pattern
func (a *SchemaLinkerAdapter) Link(ctx context.Context, query string, allTables map[string]*TableInfo) ([]string, []ReActStep, error) {
	// Perform semantic grounding
	result, err := a.service.Ground(ctx, query, a.mode)
	if err != nil {
		return nil, nil, fmt.Errorf("semantic grounding failed: %w", err)
	}

	// Extract table names
	tables := a.service.GetSelectedTables(result.Context)

	// If no tables selected by grounding, fall back to checking all tables
	if len(tables) == 0 && len(allTables) > 0 {
		// Return all tables as fallback
		for name := range allTables {
			tables = append(tables, name)
		}
	}

	// Convert grounding result to ReAct steps for transparency
	steps := a.convertToReActSteps(result)

	return tables, steps, nil
}

// convertToReActSteps converts grounding result to ReAct-style steps
func (a *SchemaLinkerAdapter) convertToReActSteps(result *GroundingResult) []ReActStep {
	steps := make([]ReActStep, 0)

	// Step 1: Coarse retrieval
	step1 := ReActStep{
		Step:    1,
		Thought: fmt.Sprintf("Performing coarse-grained vector retrieval across %d signals", len(result.CoarseSignals)),
		Action:  "semantic_search",
		ActionInput: map[string]interface{}{
			"mode":     string(a.mode),
			"duration": result.CoarseDuration.String(),
		},
		Phase: "semantic_grounding",
	}
	
	// Build observation from coarse signals
	if len(result.CoarseSignals) > 0 {
		signalTypes := make(map[SignalType]int)
		for _, sig := range result.CoarseSignals {
			signalTypes[sig.SignalType]++
		}
		var typeStrs []string
		for st, count := range signalTypes {
			typeStrs = append(typeStrs, fmt.Sprintf("%s: %d", st, count))
		}
		step1.Observation = fmt.Sprintf("Retrieved %d signals: %s", len(result.CoarseSignals), typeStrs)
	}
	steps = append(steps, step1)

	// Step 2: Fine selection (if performed)
	if result.SelectionDuration > 0 && result.Context != nil {
		step2 := ReActStep{
			Step:    2,
			Thought: fmt.Sprintf("LLM-based fine-grained selection from %d candidates", len(result.CoarseSignals)),
			Action:  "llm_selection",
			ActionInput: map[string]interface{}{
				"duration": result.SelectionDuration.String(),
			},
			Phase: "semantic_grounding",
		}

		// Build observation from selected context
		ctx := result.Context
		var selectedItems []string
		for _, t := range ctx.Tables {
			selectedItems = append(selectedItems, fmt.Sprintf("table:%s(%.2f)", t.Name, t.Relevance))
		}
		for _, c := range ctx.Columns {
			selectedItems = append(selectedItems, fmt.Sprintf("column:%s.%s(%.2f)", c.TableName, c.ColumnName, c.Relevance))
		}
		step2.Observation = fmt.Sprintf("Selected: %v", selectedItems)
		steps = append(steps, step2)
	}

	// Final step: Summary
	finalStep := ReActStep{
		Step:    len(steps) + 1,
		Thought: fmt.Sprintf("Semantic grounding completed in %s", result.TotalDuration),
		Action:  "final_answer",
		ActionInput: map[string]interface{}{
			"tables": a.service.GetSelectedTables(result.Context),
			"mode":   result.Mode,
		},
		Observation: fmt.Sprintf("Grounded %d tables, %d columns",
			len(result.Context.Tables), len(result.Context.Columns)),
		Phase: "semantic_grounding",
	}
	steps = append(steps, finalStep)

	return steps
}

// GetGroundedContext returns the full grounded context from the last grounding operation
func (a *SchemaLinkerAdapter) GetGroundedContext(ctx context.Context, query string) (*GroundedContext, error) {
	result, err := a.service.Ground(ctx, query, a.mode)
	if err != nil {
		return nil, err
	}
	return result.Context, nil
}

// FormatAsPrompt formats the grounded context as a prompt string
func (a *SchemaLinkerAdapter) FormatAsPrompt(ctx *GroundedContext) string {
	return a.service.FormatContextPrompt(ctx)
}
