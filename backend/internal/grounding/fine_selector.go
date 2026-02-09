package grounding

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// FineSelector performs LLM-based fine-grained selection from coarse signals
type FineSelector struct {
	llm    llms.Model
	config FineSelectionConfig
}

// NewFineSelector creates a new fine selector
func NewFineSelector(llm llms.Model, config FineSelectionConfig) *FineSelector {
	return &FineSelector{
		llm:    llm,
		config: config,
	}
}

// SelectionRequest represents a selection request
type SelectionRequest struct {
	Query   string
	Signals []*RetrievalSignal
}

// SelectionResult represents the fine selection result
type SelectionResult struct {
	Context       *GroundedContext
	ReasoningLog  string
	Duration      time.Duration
}

// Select performs LLM-based fine-grained selection
func (s *FineSelector) Select(ctx context.Context, req *SelectionRequest) (*SelectionResult, error) {
	start := time.Now()

	// Limit candidates
	candidates := req.Signals
	if len(candidates) > s.config.MaxCandidates {
		candidates = candidates[:s.config.MaxCandidates]
	}

	// Build the selection prompt
	prompt := s.buildSelectionPrompt(req.Query, candidates)

	// Call LLM for selection via langchaingo GenerateContent
	messages := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: selectionSystemPrompt}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: prompt}}},
	}
	resp, err := s.llm.GenerateContent(ctx, messages,
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(2000),
	)
	if err != nil {
		return nil, fmt.Errorf("LLM selection failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM selection returned no choices")
	}

	// Parse LLM response
	groundedCtx, reasoning, err := s.parseSelectionResponse(resp.Choices[0].Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selection response: %w", err)
	}

	// Enrich with metadata from signals
	s.enrichContext(groundedCtx, candidates)

	groundedCtx.Query = req.Query
	groundedCtx.GroundingTime = time.Since(start)
	groundedCtx.SignalsProbed = len(req.Signals)
	groundedCtx.SignalsSelected = len(groundedCtx.Tables) + len(groundedCtx.Columns)
	groundedCtx.Reasoning = reasoning // Store reasoning in context for transparency

	return &SelectionResult{
		Context:      groundedCtx,
		ReasoningLog: reasoning,
		Duration:     time.Since(start),
	}, nil
}

const selectionSystemPrompt = `You are an expert database schema analyst. Your task is to select the most relevant database elements for a natural language query.

Given a user query and a list of candidate database elements (tables, columns, relationships, business context), select ONLY the elements that are directly relevant to answering the query.

For each selected element, provide:
1. The element identifier
2. A brief reason why it's relevant
3. A confidence score (0.0-1.0)

Be conservative - only select elements you're confident are needed. Missing an element is better than including irrelevant ones.

Respond in JSON format:
{
  "tables": [{"name": "...", "reason": "...", "confidence": 0.9}],
  "columns": [{"table": "...", "column": "...", "reason": "...", "confidence": 0.8}],
  "relationships": [{"from": "table.column", "to": "table.column", "reason": "..."}],
  "reasoning": "Overall reasoning about the selection..."
}`

func (s *FineSelector) buildSelectionPrompt(query string, signals []*RetrievalSignal) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## User Query\n%s\n\n", query))
	sb.WriteString("## Candidate Database Elements\n\n")

	// Group signals by type
	tables := make(map[string][]string)
	columns := make(map[string][]string)
	contexts := make(map[string][]string)

	for _, sig := range signals {
		switch sig.SignalType {
		case SignalTypeTable:
			tables[sig.EntityName] = append(tables[sig.EntityName], 
				fmt.Sprintf("(score: %.2f) %s", sig.Score, sig.Content))
		case SignalTypeColumn:
			columns[sig.EntityName] = append(columns[sig.EntityName],
				fmt.Sprintf("(score: %.2f) %s", sig.Score, sig.Content))
		case SignalTypeContext:
			contexts[sig.EntityName] = append(contexts[sig.EntityName],
				fmt.Sprintf("(score: %.2f) %s", sig.Score, sig.Content))
		}
	}

	if len(tables) > 0 {
		sb.WriteString("### Tables\n")
		for name, descs := range tables {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, strings.Join(descs, "; ")))
		}
		sb.WriteString("\n")
	}

	if len(columns) > 0 {
		sb.WriteString("### Columns\n")
		for name, descs := range columns {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, strings.Join(descs, "; ")))
		}
		sb.WriteString("\n")
	}

	if len(contexts) > 0 {
		sb.WriteString("### Business Context\n")
		for name, descs := range contexts {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, strings.Join(descs, "; ")))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Select the relevant elements for the query. Be precise and conservative.")

	return sb.String()
}

// SelectionResponse represents the LLM selection response
type selectionResponse struct {
	Tables []struct {
		Name       string  `json:"name"`
		Reason     string  `json:"reason"`
		Confidence float32 `json:"confidence"`
	} `json:"tables"`
	Columns []struct {
		Table      string  `json:"table"`
		Column     string  `json:"column"`
		Reason     string  `json:"reason"`
		Confidence float32 `json:"confidence"`
	} `json:"columns"`
	Relationships []struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Reason string `json:"reason"`
	} `json:"relationships"`
	Reasoning string `json:"reasoning"`
}

func (s *FineSelector) parseSelectionResponse(content string) (*GroundedContext, string, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := content
	if idx := strings.Index(content, "```json"); idx != -1 {
		start := idx + 7
		end := strings.Index(content[start:], "```")
		if end != -1 {
			jsonStr = content[start : start+end]
		}
	} else if idx := strings.Index(content, "{"); idx != -1 {
		jsonStr = content[idx:]
		if endIdx := strings.LastIndex(jsonStr, "}"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
		}
	}

	var resp selectionResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		// Fallback: create empty context
		return &GroundedContext{}, "", nil
	}

	ctx := &GroundedContext{
		Tables:  make([]TableContext, 0, len(resp.Tables)),
		Columns: make([]ColumnContext, 0, len(resp.Columns)),
	}

	for _, t := range resp.Tables {
		if t.Confidence >= s.config.ConfidenceThreshold {
			ctx.Tables = append(ctx.Tables, TableContext{
				Name:      t.Name,
				Reason:    t.Reason,
				Relevance: t.Confidence,
			})
		}
	}

	for _, c := range resp.Columns {
		if c.Confidence >= s.config.ConfidenceThreshold {
			ctx.Columns = append(ctx.Columns, ColumnContext{
				TableName:  c.Table,
				ColumnName: c.Column,
				Reason:     c.Reason,
				Relevance:  c.Confidence,
			})
		}
	}

	for _, r := range resp.Relationships {
		parts := strings.Split(r.From, ".")
		toParts := strings.Split(r.To, ".")
		if len(parts) == 2 && len(toParts) == 2 {
			ctx.Relationships = append(ctx.Relationships, RelationshipContext{
				FromTable:  parts[0],
				FromColumn: parts[1],
				ToTable:    toParts[0],
				ToColumn:   toParts[1],
				Type:       "semantic",
				Confidence: 0.8,
			})
		}
	}

	return ctx, resp.Reasoning, nil
}

func (s *FineSelector) enrichContext(ctx *GroundedContext, signals []*RetrievalSignal) {
	// Create lookup maps for quick access
	tableDescs := make(map[string]string)
	columnDescs := make(map[string]string)

	for _, sig := range signals {
		switch sig.SignalType {
		case SignalTypeTable:
			tableDescs[sig.EntityName] = sig.Content
		case SignalTypeColumn:
			columnDescs[sig.EntityName] = sig.Content
		}
	}

	// Enrich tables
	for i := range ctx.Tables {
		if desc, ok := tableDescs[ctx.Tables[i].Name]; ok {
			ctx.Tables[i].Description = desc
		}
	}

	// Enrich columns
	for i := range ctx.Columns {
		key := fmt.Sprintf("%s.%s", ctx.Columns[i].TableName, ctx.Columns[i].ColumnName)
		if desc, ok := columnDescs[key]; ok {
			ctx.Columns[i].Description = desc
		}
	}
}
