package embedding

import (
	"context"
	"sort"
	"sync"
)

// VectorIndex provides vector similarity search functionality
type VectorIndex interface {
	// Add adds a vector with associated ID and metadata
	Add(id string, vector Vector, metadata map[string]string)

	// Search finds the k most similar vectors to the query
	Search(query Vector, k int) []SearchResult

	// Remove removes a vector by ID
	Remove(id string)

	// Size returns the number of vectors in the index
	Size() int

	// Clear removes all vectors from the index
	Clear()
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	ID         string            `json:"id"`
	Score      float32           `json:"score"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Vector     Vector            `json:"-"` // Optionally include the vector
}

// ============================================
// In-Memory Vector Index (Brute-Force)
// ============================================

// MemoryIndex is a simple in-memory vector index using brute-force search
// Suitable for small to medium datasets (< 100k vectors)
type MemoryIndex struct {
	mu       sync.RWMutex
	vectors  map[string]indexEntry
	provider EmbeddingProvider // Optional: for auto-embedding text
}

type indexEntry struct {
	vector   Vector
	metadata map[string]string
}

// NewMemoryIndex creates a new in-memory vector index
func NewMemoryIndex(provider EmbeddingProvider) *MemoryIndex {
	return &MemoryIndex{
		vectors:  make(map[string]indexEntry),
		provider: provider,
	}
}

// Add adds a vector with associated ID and metadata
func (idx *MemoryIndex) Add(id string, vector Vector, metadata map[string]string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.vectors[id] = indexEntry{
		vector:   vector,
		metadata: metadata,
	}
}

// AddText embeds text and adds to the index
func (idx *MemoryIndex) AddText(ctx context.Context, id string, text string, metadata map[string]string) error {
	if idx.provider == nil {
		return nil
	}

	vector, err := idx.provider.Embed(ctx, text)
	if err != nil {
		return err
	}

	idx.Add(id, vector, metadata)
	return nil
}

// Search finds the k most similar vectors to the query
func (idx *MemoryIndex) Search(query Vector, k int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(idx.vectors) == 0 {
		return nil
	}

	// Calculate similarities
	results := make([]SearchResult, 0, len(idx.vectors))
	for id, entry := range idx.vectors {
		score := CosineSimilarity(query, entry.vector)
		results = append(results, SearchResult{
			ID:       id,
			Score:    score,
			Metadata: entry.metadata,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top k
	if k > len(results) {
		k = len(results)
	}
	return results[:k]
}

// SearchText embeds query text and searches
func (idx *MemoryIndex) SearchText(ctx context.Context, text string, k int) ([]SearchResult, error) {
	if idx.provider == nil {
		return nil, nil
	}

	query, err := idx.provider.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	return idx.Search(query, k), nil
}

// Remove removes a vector by ID
func (idx *MemoryIndex) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.vectors, id)
}

// Size returns the number of vectors in the index
func (idx *MemoryIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.vectors)
}

// Clear removes all vectors from the index
func (idx *MemoryIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.vectors = make(map[string]indexEntry)
}

// GetByID retrieves a vector by ID
func (idx *MemoryIndex) GetByID(id string) (Vector, map[string]string, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	entry, ok := idx.vectors[id]
	if !ok {
		return nil, nil, false
	}
	return entry.vector, entry.metadata, true
}

// ============================================
// Schema Linking Index
// ============================================

// SchemaLinkingIndex is a specialized index for schema linking
// It indexes table and column names with their descriptions for retrieval
type SchemaLinkingIndex struct {
	tableIndex   *MemoryIndex
	columnIndex  *MemoryIndex
	contextIndex *MemoryIndex // For Rich Context entries
}

// NewSchemaLinkingIndex creates a new schema linking index
func NewSchemaLinkingIndex(provider EmbeddingProvider) *SchemaLinkingIndex {
	return &SchemaLinkingIndex{
		tableIndex:   NewMemoryIndex(provider),
		columnIndex:  NewMemoryIndex(provider),
		contextIndex: NewMemoryIndex(provider),
	}
}

// IndexTable adds a table to the index
func (idx *SchemaLinkingIndex) IndexTable(ctx context.Context, dbName, tableName, description string) error {
	id := dbName + "." + tableName
	text := tableName + " " + description
	metadata := map[string]string{
		"database": dbName,
		"table":    tableName,
		"type":     "table",
	}
	return idx.tableIndex.AddText(ctx, id, text, metadata)
}

// IndexColumn adds a column to the index
func (idx *SchemaLinkingIndex) IndexColumn(ctx context.Context, dbName, tableName, columnName, description string) error {
	id := dbName + "." + tableName + "." + columnName
	text := columnName + " " + description
	metadata := map[string]string{
		"database": dbName,
		"table":    tableName,
		"column":   columnName,
		"type":     "column",
	}
	return idx.columnIndex.AddText(ctx, id, text, metadata)
}

// IndexRichContext adds a Rich Context entry to the index
func (idx *SchemaLinkingIndex) IndexRichContext(ctx context.Context, dbName, tableName, key, content string) error {
	id := dbName + "." + tableName + "." + key
	text := key + " " + content
	metadata := map[string]string{
		"database": dbName,
		"table":    tableName,
		"key":      key,
		"type":     "context",
	}
	return idx.contextIndex.AddText(ctx, id, text, metadata)
}

// SearchRelevantSchema searches for relevant tables, columns, and context
func (idx *SchemaLinkingIndex) SearchRelevantSchema(ctx context.Context, query string, topK int) (*SchemaSearchResult, error) {
	result := &SchemaSearchResult{}

	// Search tables
	if idx.tableIndex.provider != nil {
		tables, err := idx.tableIndex.SearchText(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.Tables = tables
	}

	// Search columns
	if idx.columnIndex.provider != nil {
		columns, err := idx.columnIndex.SearchText(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.Columns = columns
	}

	// Search Rich Context
	if idx.contextIndex.provider != nil {
		contexts, err := idx.contextIndex.SearchText(ctx, query, topK)
		if err != nil {
			return nil, err
		}
		result.RichContexts = contexts
	}

	return result, nil
}

// SchemaSearchResult contains search results for schema linking
type SchemaSearchResult struct {
	Tables       []SearchResult `json:"tables"`
	Columns      []SearchResult `json:"columns"`
	RichContexts []SearchResult `json:"rich_contexts"`
}

// Clear clears all indexes
func (idx *SchemaLinkingIndex) Clear() {
	idx.tableIndex.Clear()
	idx.columnIndex.Clear()
	idx.contextIndex.Clear()
}

// Stats returns statistics about the index
func (idx *SchemaLinkingIndex) Stats() map[string]int {
	return map[string]int{
		"tables":        idx.tableIndex.Size(),
		"columns":       idx.columnIndex.Size(),
		"rich_contexts": idx.contextIndex.Size(),
	}
}
