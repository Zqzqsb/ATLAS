package embedding

import (
	"context"
	"fmt"

	"lucid/internal/lakebase"
)

// LakebaseEmbedder stores embeddings directly in MariaDB using the lakebase package
type LakebaseEmbedder struct {
	vectorRepo   *lakebase.MySQLVectorRepository
	provider     EmbeddingProvider
	datasourceID int64
	model        string
}

// LakebaseEmbedderConfig holds configuration for LakebaseEmbedder
type LakebaseEmbedderConfig struct {
	Pool         *lakebase.ConnectionPool
	Provider     EmbeddingProvider // Underlying provider to generate embeddings
	DatasourceID int64
	Model        string // Default: text-embedding-3-small
}

// NewLakebaseEmbedder creates a new LakebaseEmbedder
func NewLakebaseEmbedder(cfg LakebaseEmbedderConfig) (*LakebaseEmbedder, error) {
	if cfg.Pool == nil {
		return nil, fmt.Errorf("connection pool is required")
	}
	if cfg.Provider == nil {
		return nil, fmt.Errorf("embedding provider is required")
	}
	if cfg.Model == "" {
		cfg.Model = lakebase.DefaultEmbeddingModel
	}

	return &LakebaseEmbedder{
		vectorRepo:   lakebase.NewMySQLVectorRepository(cfg.Pool),
		provider:     cfg.Provider,
		datasourceID: cfg.DatasourceID,
		model:        cfg.Model,
	}, nil
}

// SetDatasourceID sets the datasource ID for subsequent operations
func (e *LakebaseEmbedder) SetDatasourceID(dsID int64) {
	e.datasourceID = dsID
}

// EmbedAndStore generates an embedding and stores it in the database
func (e *LakebaseEmbedder) EmbedAndStore(ctx context.Context, entityType lakebase.EntityType, entityID int64, text string) (int64, error) {
	// Generate embedding using the provider
	vector, err := e.provider.Embed(ctx, text)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Ensure dimension matches
	if len(vector) != lakebase.DefaultEmbeddingDimension {
		return 0, fmt.Errorf("embedding dimension mismatch: got %d, expected %d", len(vector), lakebase.DefaultEmbeddingDimension)
	}

	// Create embedding record
	emb := &lakebase.Embedding{
		DatasourceID:   e.datasourceID,
		EntityType:     entityType,
		EntityID:       entityID,
		EntityText:     text,
		Embedding:      vector,
		EmbeddingModel: e.model,
	}

	// Store in database
	return e.vectorRepo.SaveEmbedding(ctx, emb)
}

// EmbedAndStoreBatch generates embeddings and stores them in batch
func (e *LakebaseEmbedder) EmbedAndStoreBatch(ctx context.Context, items []EmbeddingItem) error {
	if len(items) == 0 {
		return nil
	}

	// Collect all texts for batch embedding
	texts := make([]string, len(items))
	for i, item := range items {
		texts[i] = item.Text
	}

	// Generate embeddings in batch
	vectors, err := e.provider.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create embedding records
	embeddings := make([]*lakebase.Embedding, len(items))
	for i, item := range items {
		if len(vectors[i]) != lakebase.DefaultEmbeddingDimension {
			return fmt.Errorf("embedding[%d] dimension mismatch: got %d, expected %d",
				i, len(vectors[i]), lakebase.DefaultEmbeddingDimension)
		}

		embeddings[i] = &lakebase.Embedding{
			DatasourceID:   e.datasourceID,
			EntityType:     item.EntityType,
			EntityID:       item.EntityID,
			EntityText:     item.Text,
			Embedding:      vectors[i],
			EmbeddingModel: e.model,
		}
	}

	// Store in batch
	return e.vectorRepo.SaveEmbeddingBatch(ctx, embeddings)
}

// SearchSimilar performs vector similarity search
func (e *LakebaseEmbedder) SearchSimilar(ctx context.Context, query string, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	// Generate query embedding
	queryVector, err := e.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in database
	return e.vectorRepo.SearchSimilar(ctx, e.datasourceID, queryVector, topK)
}

// SearchSimilarByType performs vector search filtered by entity type
func (e *LakebaseEmbedder) SearchSimilarByType(ctx context.Context, entityType lakebase.EntityType, query string, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	// Generate query embedding
	queryVector, err := e.provider.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in database
	return e.vectorRepo.SearchSimilarByType(ctx, e.datasourceID, entityType, queryVector, topK)
}

// SearchSimilarWithVector performs search using a pre-computed vector
func (e *LakebaseEmbedder) SearchSimilarWithVector(ctx context.Context, vector []float32, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	return e.vectorRepo.SearchSimilar(ctx, e.datasourceID, vector, topK)
}

// DeleteByEntity deletes embeddings for a specific entity
func (e *LakebaseEmbedder) DeleteByEntity(ctx context.Context, entityType lakebase.EntityType, entityID int64) error {
	return e.vectorRepo.DeleteEmbeddingsByEntity(ctx, e.datasourceID, entityType, entityID)
}

// DeleteByDatasource deletes all embeddings for the current datasource
func (e *LakebaseEmbedder) DeleteByDatasource(ctx context.Context) error {
	return e.vectorRepo.DeleteEmbeddingsByDatasource(ctx, e.datasourceID)
}

// Count returns the count of embeddings for the current datasource
func (e *LakebaseEmbedder) Count(ctx context.Context) (int64, error) {
	return e.vectorRepo.CountEmbeddingsByDatasource(ctx, e.datasourceID)
}

// EmbeddingItem represents an item to be embedded
type EmbeddingItem struct {
	EntityType lakebase.EntityType
	EntityID   int64
	Text       string
}

// TableEmbeddingText generates the text to embed for a table
func TableEmbeddingText(tableName, description string, columnNames []string) string {
	text := fmt.Sprintf("Table: %s", tableName)
	if description != "" {
		text += fmt.Sprintf(". Description: %s", description)
	}
	if len(columnNames) > 0 {
		text += fmt.Sprintf(". Columns: %s", joinStrings(columnNames, ", "))
	}
	return text
}

// ColumnEmbeddingText generates the text to embed for a column
func ColumnEmbeddingText(tableName, columnName, dataType, description string) string {
	text := fmt.Sprintf("Column: %s.%s (%s)", tableName, columnName, dataType)
	if description != "" {
		text += fmt.Sprintf(". Description: %s", description)
	}
	return text
}

// ContextEmbeddingText generates the text to embed for business context
func ContextEmbeddingText(tableName, columnName, contextType, content string) string {
	if columnName != "" {
		return fmt.Sprintf("Context for %s.%s (%s): %s", tableName, columnName, contextType, content)
	}
	return fmt.Sprintf("Context for %s (%s): %s", tableName, contextType, content)
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// SchemaEmbedder provides high-level functions for embedding schema elements
type SchemaEmbedder struct {
	embedder *LakebaseEmbedder
}

// NewSchemaEmbedder creates a new SchemaEmbedder
func NewSchemaEmbedder(embedder *LakebaseEmbedder) *SchemaEmbedder {
	return &SchemaEmbedder{embedder: embedder}
}

// EmbedTable generates and stores embedding for a table
func (se *SchemaEmbedder) EmbedTable(ctx context.Context, tableID int64, tableName, description string, columnNames []string) error {
	text := TableEmbeddingText(tableName, description, columnNames)
	_, err := se.embedder.EmbedAndStore(ctx, lakebase.EntityTypeTable, tableID, text)
	return err
}

// EmbedColumn generates and stores embedding for a column
func (se *SchemaEmbedder) EmbedColumn(ctx context.Context, columnID int64, tableName, columnName, dataType, description string) error {
	text := ColumnEmbeddingText(tableName, columnName, dataType, description)
	_, err := se.embedder.EmbedAndStore(ctx, lakebase.EntityTypeColumn, columnID, text)
	return err
}

// EmbedContext generates and stores embedding for business context
func (se *SchemaEmbedder) EmbedContext(ctx context.Context, contextID int64, tableName, columnName, contextType, content string) error {
	text := ContextEmbeddingText(tableName, columnName, contextType, content)
	_, err := se.embedder.EmbedAndStore(ctx, lakebase.EntityTypeContext, contextID, text)
	return err
}

// EmbedTableBatch embeds multiple tables in batch
func (se *SchemaEmbedder) EmbedTableBatch(ctx context.Context, tables []TableEmbedInfo) error {
	items := make([]EmbeddingItem, len(tables))
	for i, t := range tables {
		items[i] = EmbeddingItem{
			EntityType: lakebase.EntityTypeTable,
			EntityID:   t.ID,
			Text:       TableEmbeddingText(t.Name, t.Description, t.ColumnNames),
		}
	}
	return se.embedder.EmbedAndStoreBatch(ctx, items)
}

// EmbedColumnBatch embeds multiple columns in batch
func (se *SchemaEmbedder) EmbedColumnBatch(ctx context.Context, columns []ColumnEmbedInfo) error {
	items := make([]EmbeddingItem, len(columns))
	for i, c := range columns {
		items[i] = EmbeddingItem{
			EntityType: lakebase.EntityTypeColumn,
			EntityID:   c.ID,
			Text:       ColumnEmbeddingText(c.TableName, c.Name, c.DataType, c.Description),
		}
	}
	return se.embedder.EmbedAndStoreBatch(ctx, items)
}

// TableEmbedInfo holds information needed to embed a table
type TableEmbedInfo struct {
	ID          int64
	Name        string
	Description string
	ColumnNames []string
}

// ColumnEmbedInfo holds information needed to embed a column
type ColumnEmbedInfo struct {
	ID          int64
	TableName   string
	Name        string
	DataType    string
	Description string
}

// SearchTables searches for tables similar to the query
func (se *SchemaEmbedder) SearchTables(ctx context.Context, query string, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	return se.embedder.SearchSimilarByType(ctx, lakebase.EntityTypeTable, query, topK)
}

// SearchColumns searches for columns similar to the query
func (se *SchemaEmbedder) SearchColumns(ctx context.Context, query string, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	return se.embedder.SearchSimilarByType(ctx, lakebase.EntityTypeColumn, query, topK)
}
