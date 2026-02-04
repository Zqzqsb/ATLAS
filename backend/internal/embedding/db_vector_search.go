// Package embedding provides vector embedding and search capabilities.
package embedding

import (
	"context"
	"fmt"
	"sort"
	"time"

	"lucid/internal/lakebase"
)

// DBVectorSearch provides in-database vector search using MariaDB HNSW index
type DBVectorSearch struct {
	vectorRepo *lakebase.MySQLVectorRepository
	provider   EmbeddingProvider
	config     *DBVectorSearchConfig
}

// DBVectorSearchConfig holds configuration for DB vector search
type DBVectorSearchConfig struct {
	TopK           int     // Number of results to return (default: 10)
	MinScore       float64 // Minimum similarity score (default: 0.3)
	MaxDistance    float64 // Maximum cosine distance (default: 0.7)
	IncludeTables  bool    // Include table embeddings in search
	IncludeColumns bool    // Include column embeddings in search
	IncludeContext bool    // Include context embeddings in search
}

// DefaultDBVectorSearchConfig returns default configuration
func DefaultDBVectorSearchConfig() *DBVectorSearchConfig {
	return &DBVectorSearchConfig{
		TopK:           10,
		MinScore:       0.3,
		MaxDistance:    0.7,
		IncludeTables:  true,
		IncludeColumns: true,
		IncludeContext: true,
	}
}

// NewDBVectorSearch creates a new DB vector search instance
func NewDBVectorSearch(pool *lakebase.ConnectionPool, provider EmbeddingProvider, config *DBVectorSearchConfig) *DBVectorSearch {
	if config == nil {
		config = DefaultDBVectorSearchConfig()
	}
	return &DBVectorSearch{
		vectorRepo: lakebase.NewMySQLVectorRepository(pool),
		provider:   provider,
		config:     config,
	}
}

// DBSearchResult represents the result of a database vector search
type DBSearchResult struct {
	Query        string               `json:"query"`
	Tables       []EntitySearchResult `json:"tables"`
	Columns      []EntitySearchResult `json:"columns"`
	Contexts     []EntitySearchResult `json:"contexts"`
	SearchTimeMs int64                `json:"search_time_ms"`
	TotalResults int                  `json:"total_results"`
}

// EntitySearchResult represents a single search result
type EntitySearchResult struct {
	EntityID   int64              `json:"entity_id"`
	EntityType lakebase.EntityType `json:"entity_type"`
	EntityText string             `json:"entity_text"`
	Score      float64            `json:"score"`    // Similarity score (1 - distance)
	Distance   float64            `json:"distance"` // Cosine distance
	Metadata   map[string]string  `json:"metadata,omitempty"`
}

// Search performs vector similarity search against the lake-base embeddings
func (s *DBVectorSearch) Search(ctx context.Context, dsID int64, query string) (*DBSearchResult, error) {
	startTime := time.Now()

	result := &DBSearchResult{
		Query:   query,
		Tables:  []EntitySearchResult{},
		Columns: []EntitySearchResult{},
		Contexts: []EntitySearchResult{},
	}

	// Generate query embedding
	queryVector, err := s.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search tables
	if s.config.IncludeTables {
		tableResults, err := s.vectorRepo.SearchSimilarByType(ctx, dsID, lakebase.EntityTypeTable, queryVector, s.config.TopK)
		if err == nil {
			for _, r := range tableResults {
				if r.Distance <= s.config.MaxDistance {
					result.Tables = append(result.Tables, EntitySearchResult{
						EntityID:   r.EntityID,
						EntityType: r.EntityType,
						EntityText: r.EntityText,
						Score:      1.0 - r.Distance,
						Distance:   r.Distance,
					})
				}
			}
		}
	}

	// Search columns
	if s.config.IncludeColumns {
		columnResults, err := s.vectorRepo.SearchSimilarByType(ctx, dsID, lakebase.EntityTypeColumn, queryVector, s.config.TopK)
		if err == nil {
			for _, r := range columnResults {
				if r.Distance <= s.config.MaxDistance {
					result.Columns = append(result.Columns, EntitySearchResult{
						EntityID:   r.EntityID,
						EntityType: r.EntityType,
						EntityText: r.EntityText,
						Score:      1.0 - r.Distance,
						Distance:   r.Distance,
					})
				}
			}
		}
	}

	// Search context
	if s.config.IncludeContext {
		contextResults, err := s.vectorRepo.SearchSimilarByType(ctx, dsID, lakebase.EntityTypeContext, queryVector, s.config.TopK)
		if err == nil {
			for _, r := range contextResults {
				if r.Distance <= s.config.MaxDistance {
					result.Contexts = append(result.Contexts, EntitySearchResult{
						EntityID:   r.EntityID,
						EntityType: r.EntityType,
						EntityText: r.EntityText,
						Score:      1.0 - r.Distance,
						Distance:   r.Distance,
					})
				}
			}
		}
	}

	result.SearchTimeMs = time.Since(startTime).Milliseconds()
	result.TotalResults = len(result.Tables) + len(result.Columns) + len(result.Contexts)

	return result, nil
}

// SearchAll performs unified vector search across all entity types
func (s *DBVectorSearch) SearchAll(ctx context.Context, dsID int64, query string, topK int) ([]EntitySearchResult, error) {
	// Generate query embedding
	queryVector, err := s.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search all types
	results, err := s.vectorRepo.SearchSimilar(ctx, dsID, queryVector, topK*3) // Get more to filter later
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Convert and filter
	var searchResults []EntitySearchResult
	for _, r := range results {
		if r.Distance <= s.config.MaxDistance {
			searchResults = append(searchResults, EntitySearchResult{
				EntityID:   r.EntityID,
				EntityType: r.EntityType,
				EntityText: r.EntityText,
				Score:      1.0 - r.Distance,
				Distance:   r.Distance,
			})
		}
	}

	// Sort by score and limit
	sort.Slice(searchResults, func(i, j int) bool {
		return searchResults[i].Score > searchResults[j].Score
	})

	if len(searchResults) > topK {
		searchResults = searchResults[:topK]
	}

	return searchResults, nil
}

// SearchTables searches only table embeddings
func (s *DBVectorSearch) SearchTables(ctx context.Context, dsID int64, query string, topK int) ([]EntitySearchResult, error) {
	queryVector, err := s.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	results, err := s.vectorRepo.SearchSimilarByType(ctx, dsID, lakebase.EntityTypeTable, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("table search failed: %w", err)
	}

	var searchResults []EntitySearchResult
	for _, r := range results {
		searchResults = append(searchResults, EntitySearchResult{
			EntityID:   r.EntityID,
			EntityType: r.EntityType,
			EntityText: r.EntityText,
			Score:      1.0 - r.Distance,
			Distance:   r.Distance,
		})
	}

	return searchResults, nil
}

// SearchWithVector searches using a pre-computed vector
func (s *DBVectorSearch) SearchWithVector(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]EntitySearchResult, error) {
	results, err := s.vectorRepo.SearchSimilar(ctx, dsID, queryVector, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	var searchResults []EntitySearchResult
	for _, r := range results {
		searchResults = append(searchResults, EntitySearchResult{
			EntityID:   r.EntityID,
			EntityType: r.EntityType,
			EntityText: r.EntityText,
			Score:      1.0 - r.Distance,
			Distance:   r.Distance,
		})
	}

	return searchResults, nil
}

// TableCandidate represents a table candidate from vector search
type TableCandidate struct {
	TableName     string              `json:"table_name"`
	Score         float64             `json:"score"`
	MatchedOn     string              `json:"matched_on"` // "table", "column", "context"
	MatchedText   string              `json:"matched_text"`
	ColumnMatches []ColumnCandidate   `json:"column_matches,omitempty"`
	ContextMatches []ContextCandidate `json:"context_matches,omitempty"`
}

// ColumnCandidate represents a column match
type ColumnCandidate struct {
	ColumnName string  `json:"column_name"`
	Score      float64 `json:"score"`
	Text       string  `json:"text"`
}

// ContextCandidate represents a context match
type ContextCandidate struct {
	ContextType string  `json:"context_type"`
	Score       float64 `json:"score"`
	Text        string  `json:"text"`
}

// SearchTableCandidates searches and aggregates results by table
func (s *DBVectorSearch) SearchTableCandidates(ctx context.Context, dsID int64, query string, topK int) ([]TableCandidate, error) {
	// Perform full search
	result, err := s.Search(ctx, dsID, query)
	if err != nil {
		return nil, err
	}

	// Aggregate by table
	tableMap := make(map[string]*TableCandidate)

	// Process table matches (highest weight)
	for _, t := range result.Tables {
		tableName := extractTableNameFromText(t.EntityText)
		if tableName == "" {
			continue
		}

		if existing, ok := tableMap[tableName]; ok {
			// Update if better score
			if t.Score > existing.Score {
				existing.Score = t.Score
				existing.MatchedOn = "table"
				existing.MatchedText = t.EntityText
			}
		} else {
			tableMap[tableName] = &TableCandidate{
				TableName:   tableName,
				Score:       t.Score * 1.5, // Boost table direct matches
				MatchedOn:   "table",
				MatchedText: t.EntityText,
			}
		}
	}

	// Process column matches
	for _, c := range result.Columns {
		tableName := extractTableNameFromColumnText(c.EntityText)
		if tableName == "" {
			continue
		}

		colMatch := ColumnCandidate{
			ColumnName: extractColumnName(c.EntityText),
			Score:      c.Score,
			Text:       c.EntityText,
		}

		if existing, ok := tableMap[tableName]; ok {
			existing.ColumnMatches = append(existing.ColumnMatches, colMatch)
			// Boost score based on column matches
			existing.Score += c.Score * 0.3
		} else {
			tableMap[tableName] = &TableCandidate{
				TableName:     tableName,
				Score:         c.Score,
				MatchedOn:     "column",
				MatchedText:   c.EntityText,
				ColumnMatches: []ColumnCandidate{colMatch},
			}
		}
	}

	// Process context matches
	for _, ctx := range result.Contexts {
		tableName := extractTableNameFromContextText(ctx.EntityText)
		if tableName == "" {
			continue
		}

		ctxMatch := ContextCandidate{
			Score: ctx.Score,
			Text:  ctx.EntityText,
		}

		if existing, ok := tableMap[tableName]; ok {
			existing.ContextMatches = append(existing.ContextMatches, ctxMatch)
			// Context matches are valuable signals
			existing.Score += ctx.Score * 0.5
		} else {
			tableMap[tableName] = &TableCandidate{
				TableName:      tableName,
				Score:          ctx.Score * 1.2,
				MatchedOn:      "context",
				MatchedText:    ctx.EntityText,
				ContextMatches: []ContextCandidate{ctxMatch},
			}
		}
	}

	// Convert to slice and sort by score
	var candidates []TableCandidate
	for _, tc := range tableMap {
		candidates = append(candidates, *tc)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Limit to topK
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	return candidates, nil
}

// Helper functions to extract table/column names from entity text

func extractTableNameFromText(text string) string {
	// Format: "Table: tablename..."
	if len(text) > 7 && text[:7] == "Table: " {
		parts := splitOnFirst(text[7:], ".")
		return parts[0]
	}
	return ""
}

func extractTableNameFromColumnText(text string) string {
	// Format: "Column: tablename.columnname..."
	if len(text) > 8 && text[:8] == "Column: " {
		rest := text[8:]
		parts := splitOnFirst(rest, ".")
		return parts[0]
	}
	return ""
}

func extractColumnName(text string) string {
	// Format: "Column: tablename.columnname (type)"
	if len(text) > 8 && text[:8] == "Column: " {
		rest := text[8:]
		parts := splitOnFirst(rest, ".")
		if len(parts) > 1 {
			colPart := parts[1]
			spaceParts := splitOnFirst(colPart, " ")
			return spaceParts[0]
		}
	}
	return ""
}

func extractTableNameFromContextText(text string) string {
	// Format: "Context for tablename..." or "Context for tablename.column..."
	prefix := "Context for "
	if len(text) > len(prefix) && text[:len(prefix)] == prefix {
		rest := text[len(prefix):]
		parts := splitOnFirst(rest, ".")
		tablePart := parts[0]
		// Remove any trailing context type info
		spaceParts := splitOnFirst(tablePart, " ")
		return spaceParts[0]
	}
	return ""
}

func splitOnFirst(s, sep string) []string {
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			return []string{s[:i], s[i+len(sep):]}
		}
	}
	return []string{s}
}

// GetEmbeddingStats returns statistics about embeddings for a datasource
func (s *DBVectorSearch) GetEmbeddingStats(ctx context.Context, dsID int64) (map[string]int64, error) {
	stats := make(map[string]int64)

	tableCount, _ := s.vectorRepo.CountEmbeddingsByType(ctx, dsID, lakebase.EntityTypeTable)
	columnCount, _ := s.vectorRepo.CountEmbeddingsByType(ctx, dsID, lakebase.EntityTypeColumn)
	contextCount, _ := s.vectorRepo.CountEmbeddingsByType(ctx, dsID, lakebase.EntityTypeContext)

	stats["tables"] = tableCount
	stats["columns"] = columnCount
	stats["contexts"] = contextCount
	stats["total"] = tableCount + columnCount + contextCount

	return stats, nil
}
