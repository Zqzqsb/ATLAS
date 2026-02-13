package grounding

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/logger"
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
	// Compact mode: only output name+type+PK/FK in prompt (skip RC details).
	// Used in SmallScale where table count is small and RC is not needed for selection.
	Compact bool
}

// LinkingResult represents the linking agent's output
type LinkingResult struct {
	SelectedTables []SelectedTable
	Reasoning      string
	Duration       time.Duration
}

// SelectedTable represents a table selected by the linking agent
type SelectedTable struct {
	Name            string           `json:"name"`
	Reason          string           `json:"reason"`
	Confidence      float32          `json:"confidence"`
	RelevantColumns []RelevantColumn `json:"relevant_columns,omitempty"`
}

// RelevantColumn represents a column identified as relevant by the linking agent.
// This replaces the standalone field-alignment LLM call with zero extra cost.
type RelevantColumn struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
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
	log := logger.With("component", "linking_agent")

	log.Debug("[Link] Starting schema linking",
		"query", req.Query,
		"table_count", len(req.Schemas),
		"has_vector_signals", len(req.VectorSignals) > 0,
	)

	prompt := a.buildLinkingPrompt(req)
	log.Debug("[Link] Built prompt",
		"prompt_length", len(prompt),
		"prompt_preview", truncateLinking(prompt, 500),
	)

	messages := []llms.MessageContent{
		{Role: llms.ChatMessageTypeSystem, Parts: []llms.ContentPart{llms.TextContent{Text: linkingAgentSystemPrompt}}},
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: prompt}}},
	}

	resp, err := a.llm.GenerateContent(ctx, messages,
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(2000),
	)
	if err != nil {
		log.Error("[Link] LLM call failed", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("linking agent LLM call failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		log.Error("[Link] LLM returned no choices")
		return nil, fmt.Errorf("linking agent returned no choices")
	}

	rawResponse := resp.Choices[0].Content
	log.Info("[Link] LLM response received",
		"response_length", len(rawResponse),
	)
	log.Debug("[Link] LLM raw response",
		"response", truncateLinking(rawResponse, 1000),
	)

	selected, reasoning, err := a.parseLinkingResponse(rawResponse)
	if err != nil {
		log.Error("[Link] Failed to parse response", "error", err, "raw", truncateLinking(rawResponse, 300))
		return nil, fmt.Errorf("failed to parse linking response: %w", err)
	}

	// Filter by confidence threshold
	var filtered []SelectedTable
	for _, t := range selected {
		if t.Confidence >= a.config.ConfidenceThreshold {
			filtered = append(filtered, t)
		} else {
			log.Debug("[Link] Table filtered out (low confidence)",
				"table", t.Name,
				"confidence", fmt.Sprintf("%.2f", t.Confidence),
				"threshold", fmt.Sprintf("%.2f", a.config.ConfidenceThreshold),
			)
		}
	}

	// Log final result
	selectedNames := make([]string, len(filtered))
	for i, t := range filtered {
		selectedNames[i] = t.Name
		log.Debug("[Link] Selected table",
			"table", t.Name,
			"confidence", fmt.Sprintf("%.2f", t.Confidence),
			"reason", t.Reason,
			"relevant_columns", len(t.RelevantColumns),
		)
	}
	log.Info("[Link] Schema linking completed",
		"selected_tables", strings.Join(selectedNames, ", "),
		"reasoning", truncateLinking(reasoning, 200),
		"duration", time.Since(start).Round(time.Millisecond),
	)

	return &LinkingResult{
		SelectedTables: filtered,
		Reasoning:      reasoning,
		Duration:       time.Since(start),
	}, nil
}

const linkingAgentSystemPrompt = `You are an expert database schema analyst performing Schema Linking and Field Selection.

Your task: Given a natural language query and a database schema:
1. Identify which tables are needed to answer the query
2. For each selected table, identify which columns are relevant to the query

You will receive the COMPLETE schema of all available tables, including:
- Table names and descriptions
- Column names, types, descriptions, sample values, and synonyms
- Foreign key relationships

Analyze the query carefully. Consider:
1. Which tables contain the data being queried?
2. Which tables are needed for JOINs to connect the data?
3. Business context: Does a column description or synonym match the query intent?
4. Sample values: Do they help confirm the right table/column?
5. Which columns should appear in SELECT, WHERE, JOIN, GROUP BY, or ORDER BY?

For column selection:
- Include columns that directly answer the question (SELECT candidates)
- Include columns needed for filtering (WHERE candidates)
- Include columns needed for joining (JOIN keys)
- Include columns needed for aggregation or ordering
- Do NOT include every column — only the ones relevant to this specific query

Be thorough but precise. Missing a needed table is worse than including an extra one.

Respond in JSON format:
{
  "tables": [
    {
      "name": "table_name",
      "reason": "why this table is needed",
      "confidence": 0.9,
      "relevant_columns": [
        {"name": "column_name", "reason": "SELECT: directly answers the question"}
      ]
    }
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

		if !req.Compact && schema.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", schema.Description))
		}
		if schema.RowCount > 0 {
			sb.WriteString(fmt.Sprintf("Row count: %d\n", schema.RowCount))
		}

		// Columns — compact mode omits RC details to reduce prompt size
		includeRC := a.config.IncludeRichContext && !req.Compact
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

				if includeRC {
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

func truncateLinking(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// linkingResponse represents the LLM's linking response
type linkingResponse struct {
	Tables []struct {
		Name            string  `json:"name"`
		Reason          string  `json:"reason"`
		Confidence      float32 `json:"confidence"`
		RelevantColumns []struct {
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"relevant_columns"`
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
		st := SelectedTable{
			Name:       t.Name,
			Reason:     t.Reason,
			Confidence: t.Confidence,
		}
		for _, col := range t.RelevantColumns {
			st.RelevantColumns = append(st.RelevantColumns, RelevantColumn{
				Name:   col.Name,
				Reason: col.Reason,
			})
		}
		tables = append(tables, st)
	}

	return tables, resp.Reasoning, nil
}
