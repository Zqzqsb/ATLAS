package grounding

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// SchemaInfo represents full table schema for linking agent input
type SchemaInfo struct {
	TableName   string       `json:"table_name"`
	Description string       `json:"description,omitempty"`
	RowCount    int64        `json:"row_count,omitempty"`
	Columns     []ColumnInfo `json:"columns"`
	ForeignKeys []FKInfo     `json:"foreign_keys,omitempty"`
}

// ColumnInfo represents column details for linking agent
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	IsPrimaryKey bool   `json:"is_primary_key,omitempty"`
	IsNullable   bool   `json:"is_nullable,omitempty"`
	SampleValues string `json:"sample_values,omitempty"`
	Synonyms     string `json:"synonyms,omitempty"`
}

// FKInfo represents a foreign key relationship
type FKInfo struct {
	Column           string `json:"column"`
	ReferencedTable  string `json:"referenced_table"`
	ReferencedColumn string `json:"referenced_column"`
}

// LinkingRequest represents a request to the linking agent
type LinkingRequest struct {
	Query   string
	Schemas []SchemaInfo
	// Optional: signals from vector retrieval (large scale mode)
	VectorSignals []*RetrievalSignal
}

// LinkingResult represents the linking agent's output
type LinkingResult struct {
	SelectedTables []SelectedTable
	Reasoning      string
	Duration       time.Duration
}

// SelectedTable represents a table selected by the linking agent
type SelectedTable struct {
	Name       string  `json:"name"`
	Reason     string  `json:"reason"`
	Confidence float32 `json:"confidence"`
}

// LinkingAgent performs LLM-based schema linking with full schema context
type LinkingAgent struct {
	llm    llms.Model
	config LinkingAgentConfig
}

// NewLinkingAgent creates a new linking agent
func NewLinkingAgent(llm llms.Model, config LinkingAgentConfig) *LinkingAgent {
	return &LinkingAgent{
		llm:    llm,
		config: config,
	}
}

// Link performs schema linking given full schema information
func (a *LinkingAgent) Link(ctx context.Context, req *LinkingRequest) (*LinkingResult, error) {
	start := time.Now()

	prompt := a.buildLinkingPrompt(req)

	messages := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: linkingAgentSystemPrompt}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: prompt}}},
	}

	resp, err := a.llm.GenerateContent(ctx, messages,
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(2000),
	)
	if err != nil {
		return nil, fmt.Errorf("linking agent LLM call failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("linking agent returned no choices")
	}

	selected, reasoning, err := a.parseLinkingResponse(resp.Choices[0].Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse linking response: %w", err)
	}

	// Filter by confidence threshold
	var filtered []SelectedTable
	for _, t := range selected {
		if t.Confidence >= a.config.ConfidenceThreshold {
			filtered = append(filtered, t)
		}
	}

	return &LinkingResult{
		SelectedTables: filtered,
		Reasoning:      reasoning,
		Duration:       time.Since(start),
	}, nil
}

const linkingAgentSystemPrompt = `You are an expert database schema analyst performing Schema Linking.

Your task: Given a natural language query and a database schema, identify which tables are needed to answer the query.

You will receive the COMPLETE schema of all available tables, including:
- Table names and descriptions
- Column names, types, descriptions, sample values, and synonyms
- Foreign key relationships

Analyze the query carefully and select ONLY the tables that are directly needed. Consider:
1. Which tables contain the data being queried?
2. Which tables are needed for JOINs to connect the data?
3. Business context: Does a column description or synonym match the query intent?
4. Sample values: Do they help confirm the right table/column?

Be thorough but precise. Missing a needed table is worse than including an extra one.

Respond in JSON format:
{
  "tables": [
    {"name": "table_name", "reason": "why this table is needed", "confidence": 0.9}
  ],
  "reasoning": "Overall explanation of your selection logic"
}`

func (a *LinkingAgent) buildLinkingPrompt(req *LinkingRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## User Query\n%s\n\n", req.Query))
	sb.WriteString("## Database Schema\n\n")

	// If we have vector signals, annotate tables with retrieval scores
	signalScores := make(map[string]float32)
	if len(req.VectorSignals) > 0 {
		for _, sig := range req.VectorSignals {
			if sig.SignalType == SignalTypeTable {
				signalScores[sig.EntityName] = sig.Score
			}
		}
	}

	for _, schema := range req.Schemas {
		// Table header with optional retrieval score annotation
		sb.WriteString(fmt.Sprintf("### Table: `%s`", schema.TableName))
		if score, ok := signalScores[schema.TableName]; ok {
			sb.WriteString(fmt.Sprintf(" (vector relevance: %.2f)", score))
		}
		sb.WriteString("\n")

		if schema.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", schema.Description))
		}
		if schema.RowCount > 0 {
			sb.WriteString(fmt.Sprintf("Row count: %d\n", schema.RowCount))
		}

		// Columns
		if len(schema.Columns) > 0 && a.config.IncludeColumnDetails {
			sb.WriteString("Columns:\n")
			for _, col := range schema.Columns {
				sb.WriteString(fmt.Sprintf("  - `%s` (%s)", col.Name, col.Type))
				if col.IsPrimaryKey {
					sb.WriteString(" [PK]")
				}
				if !col.IsNullable {
					sb.WriteString(" [NOT NULL]")
				}
				sb.WriteString("\n")

				if a.config.IncludeRichContext {
					if col.Description != "" {
						sb.WriteString(fmt.Sprintf("    Description: %s\n", col.Description))
					}
					if col.SampleValues != "" {
						sb.WriteString(fmt.Sprintf("    Sample values: %s\n", col.SampleValues))
					}
					if col.Synonyms != "" {
						sb.WriteString(fmt.Sprintf("    Synonyms: %s\n", col.Synonyms))
					}
				}
			}
		}

		// Foreign keys
		if len(schema.ForeignKeys) > 0 {
			sb.WriteString("Foreign keys:\n")
			for _, fk := range schema.ForeignKeys {
				sb.WriteString(fmt.Sprintf("  - %s → %s.%s\n", fk.Column, fk.ReferencedTable, fk.ReferencedColumn))
			}
		}
		sb.WriteString("\n")
	}

	// If large scale mode with vector signals, add a hint
	if len(req.VectorSignals) > 0 {
		sb.WriteString("---\n")
		sb.WriteString("Note: Tables annotated with 'vector relevance' scores were identified by semantic search as potentially relevant. ")
		sb.WriteString("Use these scores as hints, but make your own judgment based on the full schema.\n")
	}

	sb.WriteString("\nSelect the tables needed to answer the query. Be thorough - missing a table is worse than including an extra one.")

	return sb.String()
}

// linkingResponse represents the LLM's linking response
type linkingResponse struct {
	Tables []struct {
		Name       string  `json:"name"`
		Reason     string  `json:"reason"`
		Confidence float32 `json:"confidence"`
	} `json:"tables"`
	Reasoning string `json:"reasoning"`
}

func (a *LinkingAgent) parseLinkingResponse(content string) ([]SelectedTable, string, error) {
	// Extract JSON from response
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

	var resp linkingResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	var tables []SelectedTable
	for _, t := range resp.Tables {
		tables = append(tables, SelectedTable{
			Name:       t.Name,
			Reason:     t.Reason,
			Confidence: t.Confidence,
		})
	}

	return tables, resp.Reasoning, nil
}
