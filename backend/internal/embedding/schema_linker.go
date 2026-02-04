package embedding

import (
	"context"
	"fmt"
	"sort"
	"strings"

	richctx "lucid/internal/context"
)

// TwoStageSchemaLinker implements retrieval-augmented schema linking
// Stage 1: Vector retrieval (coarse filtering) - Fast, recall-focused
// Stage 2: LLM re-ranking (fine selection) - Slow, precision-focused
type TwoStageSchemaLinker struct {
	index    *SchemaLinkingIndex
	provider EmbeddingProvider

	// Configuration
	stage1TopK int     // Number of candidates from retrieval (default: 20)
	stage2TopK int     // Number of final results (default: 5)
	minScore   float32 // Minimum similarity score threshold (default: 0.3)
}

// TwoStageConfig holds configuration for two-stage schema linker
type TwoStageConfig struct {
	Provider   EmbeddingProvider
	Stage1TopK int     // Retrieval candidates (default: 20)
	Stage2TopK int     // Final results (default: 5)
	MinScore   float32 // Minimum score threshold (default: 0.3)
}

// NewTwoStageSchemaLinker creates a new two-stage schema linker
func NewTwoStageSchemaLinker(config TwoStageConfig) *TwoStageSchemaLinker {
	if config.Stage1TopK == 0 {
		config.Stage1TopK = 20
	}
	if config.Stage2TopK == 0 {
		config.Stage2TopK = 5
	}
	if config.MinScore == 0 {
		config.MinScore = 0.3
	}

	return &TwoStageSchemaLinker{
		index:      NewSchemaLinkingIndex(config.Provider),
		provider:   config.Provider,
		stage1TopK: config.Stage1TopK,
		stage2TopK: config.Stage2TopK,
		minScore:   config.MinScore,
	}
}

// BuildIndex builds the schema linking index from SharedContext
func (sl *TwoStageSchemaLinker) BuildIndex(ctx context.Context, sharedCtx *richctx.SharedContext) error {
	if sl.provider == nil {
		return fmt.Errorf("embedding provider not configured")
	}

	sl.index.Clear()
	dbName := sharedCtx.DatabaseName

	for tableName, table := range sharedCtx.Tables {
		// Index table
		if err := sl.index.IndexTable(ctx, dbName, tableName, table.Description); err != nil {
			return fmt.Errorf("failed to index table %s: %w", tableName, err)
		}

		// Index columns
		for _, col := range table.Columns {
			desc := fmt.Sprintf("%s %s", col.Type, col.Name)
			if err := sl.index.IndexColumn(ctx, dbName, tableName, col.Name, desc); err != nil {
				return fmt.Errorf("failed to index column %s.%s: %w", tableName, col.Name, err)
			}
		}

		// Index Rich Context entries
		for key, value := range table.RichContext {
			if err := sl.index.IndexRichContext(ctx, dbName, tableName, key, value.Content); err != nil {
				return fmt.Errorf("failed to index context %s.%s: %w", tableName, key, err)
			}
		}
	}

	return nil
}

// LinkingResult represents the result of schema linking
type LinkingResult struct {
	Query          string               `json:"query"`
	RelevantTables []TableLinkingResult `json:"relevant_tables"`
	Stage1Count    int                  `json:"stage1_count"` // Candidates from retrieval
	Stage2Count    int                  `json:"stage2_count"` // Final results
}

// TableLinkingResult represents a linked table with its relevant columns and context
type TableLinkingResult struct {
	TableName       string                 `json:"table_name"`
	TableScore      float32                `json:"table_score"`
	Description     string                 `json:"description,omitempty"`
	RelevantColumns []ColumnLinkingResult  `json:"relevant_columns,omitempty"`
	RelevantContext []ContextLinkingResult `json:"relevant_context,omitempty"`
}

// ColumnLinkingResult represents a linked column
type ColumnLinkingResult struct {
	ColumnName string  `json:"column_name"`
	Score      float32 `json:"score"`
}

// ContextLinkingResult represents a linked Rich Context entry
type ContextLinkingResult struct {
	Key     string  `json:"key"`
	Content string  `json:"content"`
	Score   float32 `json:"score"`
}

// Link performs two-stage schema linking for a query
// Stage 1: Retrieve top-K candidates using vector similarity
// Stage 2: Re-rank using more sophisticated scoring (can be extended with LLM)
func (sl *TwoStageSchemaLinker) Link(ctx context.Context, query string) (*LinkingResult, error) {
	if sl.provider == nil {
		return &LinkingResult{Query: query}, nil
	}

	// Stage 1: Vector retrieval
	searchResult, err := sl.index.SearchRelevantSchema(ctx, query, sl.stage1TopK)
	if err != nil {
		return nil, fmt.Errorf("stage 1 retrieval failed: %w", err)
	}

	// Aggregate results by table
	tableScores := make(map[string]*aggregatedResult)

	// Process table results
	for _, tr := range searchResult.Tables {
		tableName := tr.Metadata["table"]
		if tableName == "" {
			continue
		}
		if _, ok := tableScores[tableName]; !ok {
			tableScores[tableName] = &aggregatedResult{
				tableName: tableName,
				columns:   make(map[string]float32),
				contexts:  make(map[string]contextEntry),
			}
		}
		// Table match boosts score significantly
		tableScores[tableName].tableScore = tr.Score * 2.0
	}

	// Process column results
	for _, cr := range searchResult.Columns {
		tableName := cr.Metadata["table"]
		colName := cr.Metadata["column"]
		if tableName == "" || colName == "" {
			continue
		}
		if _, ok := tableScores[tableName]; !ok {
			tableScores[tableName] = &aggregatedResult{
				tableName: tableName,
				columns:   make(map[string]float32),
				contexts:  make(map[string]contextEntry),
			}
		}
		tableScores[tableName].columns[colName] = cr.Score
	}

	// Process Rich Context results
	for _, ctx := range searchResult.RichContexts {
		tableName := ctx.Metadata["table"]
		key := ctx.Metadata["key"]
		if tableName == "" || key == "" {
			continue
		}
		if _, ok := tableScores[tableName]; !ok {
			tableScores[tableName] = &aggregatedResult{
				tableName: tableName,
				columns:   make(map[string]float32),
				contexts:  make(map[string]contextEntry),
			}
		}
		// Rich Context matches are valuable
		tableScores[tableName].contexts[key] = contextEntry{
			content: "", // Would need to fetch actual content
			score:   ctx.Score * 1.5,
		}
	}

	// Stage 2: Re-rank and select top results
	stage1Count := len(tableScores)
	rankedTables := sl.rerank(tableScores)

	// Filter by minimum score
	var finalResults []TableLinkingResult
	for _, rt := range rankedTables {
		if rt.TableScore < sl.minScore {
			continue
		}
		finalResults = append(finalResults, rt)
		if len(finalResults) >= sl.stage2TopK {
			break
		}
	}

	return &LinkingResult{
		Query:          query,
		RelevantTables: finalResults,
		Stage1Count:    stage1Count,
		Stage2Count:    len(finalResults),
	}, nil
}

type aggregatedResult struct {
	tableName  string
	tableScore float32
	columns    map[string]float32
	contexts   map[string]contextEntry
}

type contextEntry struct {
	content string
	score   float32
}

// rerank performs re-ranking of aggregated results
func (sl *TwoStageSchemaLinker) rerank(tableScores map[string]*aggregatedResult) []TableLinkingResult {
	// Calculate composite scores
	type scoredTable struct {
		result aggregatedResult
		score  float32
	}

	var scored []scoredTable
	for _, agg := range tableScores {
		// Composite score: table match + column matches + context matches
		compositeScore := agg.tableScore

		// Add column contributions (diminishing returns)
		colCount := 0
		for _, cs := range agg.columns {
			compositeScore += cs * (1.0 / float32(colCount+1))
			colCount++
		}

		// Add context contributions (valuable signals)
		for _, ctx := range agg.contexts {
			compositeScore += ctx.score
		}

		scored = append(scored, scoredTable{
			result: *agg,
			score:  compositeScore,
		})
	}

	// Sort by composite score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Convert to results
	var results []TableLinkingResult
	for _, st := range scored {
		result := TableLinkingResult{
			TableName:  st.result.tableName,
			TableScore: st.score,
		}

		// Add relevant columns
		for colName, score := range st.result.columns {
			result.RelevantColumns = append(result.RelevantColumns, ColumnLinkingResult{
				ColumnName: colName,
				Score:      score,
			})
		}
		// Sort columns by score
		sort.Slice(result.RelevantColumns, func(i, j int) bool {
			return result.RelevantColumns[i].Score > result.RelevantColumns[j].Score
		})

		// Add relevant context
		for key, ctx := range st.result.contexts {
			result.RelevantContext = append(result.RelevantContext, ContextLinkingResult{
				Key:     key,
				Content: ctx.content,
				Score:   ctx.score,
			})
		}
		// Sort contexts by score
		sort.Slice(result.RelevantContext, func(i, j int) bool {
			return result.RelevantContext[i].Score > result.RelevantContext[j].Score
		})

		results = append(results, result)
	}

	return results
}

// Stats returns statistics about the linker
func (sl *TwoStageSchemaLinker) Stats() map[string]interface{} {
	indexStats := sl.index.Stats()
	return map[string]interface{}{
		"provider":    sl.provider.Name(),
		"stage1_topk": sl.stage1TopK,
		"stage2_topk": sl.stage2TopK,
		"min_score":   sl.minScore,
		"index":       indexStats,
	}
}

// FormatLinkingResultForPrompt formats the linking result for LLM prompt
func FormatLinkingResultForPrompt(result *LinkingResult) string {
	if result == nil || len(result.RelevantTables) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Relevant Schema (from retrieval)\n\n")

	for _, table := range result.RelevantTables {
		sb.WriteString(fmt.Sprintf("### Table: %s (relevance: %.2f)\n", table.TableName, table.TableScore))

		if len(table.RelevantColumns) > 0 {
			sb.WriteString("Relevant columns: ")
			cols := make([]string, 0, len(table.RelevantColumns))
			for _, col := range table.RelevantColumns[:min(5, len(table.RelevantColumns))] {
				cols = append(cols, col.ColumnName)
			}
			sb.WriteString(strings.Join(cols, ", "))
			sb.WriteString("\n")
		}

		if len(table.RelevantContext) > 0 {
			sb.WriteString("Business context:\n")
			for _, ctx := range table.RelevantContext[:min(3, len(table.RelevantContext))] {
				sb.WriteString(fmt.Sprintf("- %s\n", ctx.Key))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
