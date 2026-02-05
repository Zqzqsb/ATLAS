package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
)

// LakebaseService provides high-level operations for lake-base storage
type LakebaseService struct {
	pool              *lakebase.ConnectionPool
	repo              *lakebase.MySQLRepository
	vectorRepo        *lakebase.MySQLVectorRepository
	embeddingProvider embedding.EmbeddingProvider
	config            *lakebase.LakebaseConfig
	mu                sync.RWMutex
	connected         bool
}

// NewLakebaseService creates a new lake-base service
func NewLakebaseService(configPath string) (*LakebaseService, error) {
	cfg, err := lakebase.LoadConfig(configPath)
	if err != nil {
		// Use default config if file not found
		cfg = lakebase.DefaultLakebaseConfig()
	}

	return &LakebaseService{
		config: cfg,
	}, nil
}

// NewLakebaseServiceWithConfig creates a new lake-base service with explicit config
func NewLakebaseServiceWithConfig(cfg *lakebase.LakebaseConfig) *LakebaseService {
	if cfg == nil {
		cfg = lakebase.DefaultLakebaseConfig()
	}
	return &LakebaseService{
		config: cfg,
	}
}

// Connect establishes connection to the lake-base database
func (s *LakebaseService) Connect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.connected {
		return nil
	}

	connCfg := s.config.Lakebase.ToConnectionConfig()
	s.pool = lakebase.NewConnectionPool(connCfg)

	if err := s.pool.Connect(ctx); err != nil {
		return fmt.Errorf("lakebase service: failed to connect: %w", err)
	}

	s.repo = lakebase.NewMySQLRepository(s.pool)
	s.vectorRepo = lakebase.NewMySQLVectorRepository(s.pool)

	// Initialize embedding provider if configured
	if s.config.Embedding.Enabled && s.config.Embedding.APIKey != "" {
		provider := embedding.NewOpenAIProvider(embedding.OpenAIConfig{
			APIKey:    s.config.Embedding.APIKey,
			BaseURL:   s.config.Embedding.BaseURL,
			Model:     s.config.Embedding.Model,
			Dimension: s.config.Embedding.Dimension,
		})
		s.embeddingProvider = provider
	} else {
		// Use mock embedder for development/demo purposes
		// This allows grounding to work without a real embedding API
		dimension := s.config.Embedding.Dimension
		if dimension <= 0 {
			dimension = 768 // default dimension for mock
		}
		s.embeddingProvider = embedding.NewMockEmbeddingProvider(dimension)
	}

	s.connected = true

	return nil
}

// Close closes the lake-base connection
func (s *LakebaseService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil
	}

	if s.pool != nil {
		err := s.pool.Close()
		s.connected = false
		return err
	}

	return nil
}

// IsConnected returns whether the service is connected
func (s *LakebaseService) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// GetConfig returns the lake-base configuration
func (s *LakebaseService) GetConfig() *lakebase.LakebaseConfig {
	return s.config
}

// GetPool returns the connection pool
func (s *LakebaseService) GetPool() *lakebase.ConnectionPool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pool
}

// GetEmbeddingProvider returns the embedding provider
func (s *LakebaseService) GetEmbeddingProvider() embedding.EmbeddingProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.embeddingProvider
}

// SetEmbeddingProvider sets the embedding provider
func (s *LakebaseService) SetEmbeddingProvider(provider embedding.EmbeddingProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.embeddingProvider = provider
}

// GetVectorRepository returns the vector repository for semantic grounding
func (s *LakebaseService) GetVectorRepository() *lakebase.MySQLVectorRepository {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.vectorRepo
}

// ===========================================
// Datasource Operations
// ===========================================

// CreateDatasource creates a new datasource entry
func (s *LakebaseService) CreateDatasource(ctx context.Context, ds *lakebase.Datasource) (int64, error) {
	if !s.connected {
		return 0, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.CreateDatasource(ctx, ds)
}

// GetDatasource retrieves a datasource by ID
func (s *LakebaseService) GetDatasource(ctx context.Context, id int64) (*lakebase.Datasource, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetDatasource(ctx, id)
}

// GetDatasourceByName retrieves a datasource by name
func (s *LakebaseService) GetDatasourceByName(ctx context.Context, name string) (*lakebase.Datasource, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetDatasourceByName(ctx, name)
}

// GetOrCreateDatasource gets existing datasource or creates a new one
func (s *LakebaseService) GetOrCreateDatasource(ctx context.Context, ds *lakebase.Datasource) (*lakebase.Datasource, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}

	existing, err := s.repo.GetDatasourceByName(ctx, ds.Name)
	if err == nil {
		return existing, nil
	}

	if err != lakebase.ErrDatasourceNotFound {
		return nil, err
	}

	// Create new datasource
	id, err := s.repo.CreateDatasource(ctx, ds)
	if err != nil {
		return nil, err
	}

	ds.ID = id
	return ds, nil
}

// ListDatasources lists all datasources
func (s *LakebaseService) ListDatasources(ctx context.Context) ([]*lakebase.Datasource, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.ListDatasources(ctx)
}

// ===========================================
// Schema Operations
// ===========================================

// SaveSchemaMetadata saves schema metadata for a datasource
func (s *LakebaseService) SaveSchemaMetadata(ctx context.Context, metas []*lakebase.SchemaMetadata) error {
	if !s.connected {
		return fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.SaveSchemaMetadata(ctx, metas)
}

// GetTablesByDatasource retrieves all tables from rc_tables
func (s *LakebaseService) GetTablesByDatasource(ctx context.Context, dsID int64) ([]*lakebase.TableInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetTablesByDatasource(ctx, dsID)
}

// GetColumnsByDatasource retrieves all columns from rc_columns
func (s *LakebaseService) GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*lakebase.ColumnInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetColumnsByDatasource(ctx, dsID)
}

// GetColumnsByTable retrieves columns for a specific table
func (s *LakebaseService) GetColumnsByTable(ctx context.Context, dsID int64, tableName string) ([]*lakebase.ColumnInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetColumnsByTable(ctx, dsID, tableName)
}

// GetTermsByDatasource retrieves all business terms for a datasource
func (s *LakebaseService) GetTermsByDatasource(ctx context.Context, dsID int64) ([]*lakebase.TermInfo, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetTermsByDatasource(ctx, dsID)
}

// PruneAllContext deletes all rich context data for a datasource
// This clears rc_tables, rc_columns, rc_terms, rc_relations, rc_business_context, rc_change_log
func (s *LakebaseService) PruneAllContext(ctx context.Context, dsID int64) error {
	if !s.connected {
		return fmt.Errorf("lakebase service: not connected")
	}

	// First delete embeddings associated with this datasource
	if err := s.vectorRepo.DeleteEmbeddingsByDatasource(ctx, dsID); err != nil {
		return fmt.Errorf("failed to delete embeddings: %w", err)
	}

	// Then delete all rich context data
	if err := s.repo.PruneAllContext(ctx, dsID); err != nil {
		return fmt.Errorf("failed to prune context: %w", err)
	}

	return nil
}

// GetSchemaByDatasource retrieves all schema metadata for a datasource (legacy)
func (s *LakebaseService) GetSchemaByDatasource(ctx context.Context, dsID int64) ([]*lakebase.SchemaMetadata, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetSchemaByDatasource(ctx, dsID)
}

// GetTableSchema retrieves schema for a specific table
func (s *LakebaseService) GetTableSchema(ctx context.Context, dsID int64, tableName string) ([]*lakebase.SchemaMetadata, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetTableSchema(ctx, dsID, tableName)
}

// GetTableNames retrieves all table names for a datasource
func (s *LakebaseService) GetTableNames(ctx context.Context, dsID int64) ([]string, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetTableNames(ctx, dsID)
}

// ===========================================
// Business Context Operations
// ===========================================

// SaveBusinessContext saves a single business context entry
func (s *LakebaseService) SaveBusinessContext(ctx context.Context, bc *lakebase.BusinessContext) (int64, error) {
	if !s.connected {
		return 0, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.SaveBusinessContext(ctx, bc)
}

// SaveBusinessContextBatch saves multiple business context entries
func (s *LakebaseService) SaveBusinessContextBatch(ctx context.Context, contexts []*lakebase.BusinessContext) error {
	if !s.connected {
		return fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.SaveBusinessContextBatch(ctx, contexts)
}

// GetContextByDatasource retrieves all context for a datasource
func (s *LakebaseService) GetContextByDatasource(ctx context.Context, dsID int64) ([]*lakebase.BusinessContext, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetContextByDatasource(ctx, dsID)
}

// GetContextByTable retrieves context for a specific table
func (s *LakebaseService) GetContextByTable(ctx context.Context, dsID int64, tableName string) ([]*lakebase.BusinessContext, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetContextByTable(ctx, dsID, tableName)
}

// MarkContextExpired marks context entries as expired
func (s *LakebaseService) MarkContextExpired(ctx context.Context, ids []int64, reason string) error {
	if !s.connected {
		return fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.MarkContextExpired(ctx, ids, reason)
}

// ===========================================
// Relation Operations
// ===========================================

// GetRelationsByDatasource retrieves all relations for a datasource
func (s *LakebaseService) GetRelationsByDatasource(ctx context.Context, dsID int64) ([]*lakebase.Relation, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetRelationsByDatasource(ctx, dsID)
}

// ===========================================
// Vector/Embedding Operations
// ===========================================

// SaveEmbedding saves a single embedding
func (s *LakebaseService) SaveEmbedding(ctx context.Context, emb *lakebase.Embedding) (int64, error) {
	if !s.connected {
		return 0, fmt.Errorf("lakebase service: not connected")
	}
	return s.vectorRepo.SaveEmbedding(ctx, emb)
}

// SaveEmbeddingBatch saves multiple embeddings
func (s *LakebaseService) SaveEmbeddingBatch(ctx context.Context, embeddings []*lakebase.Embedding) error {
	if !s.connected {
		return fmt.Errorf("lakebase service: not connected")
	}
	return s.vectorRepo.SaveEmbeddingBatch(ctx, embeddings)
}

// SearchSimilar performs vector similarity search
func (s *LakebaseService) SearchSimilar(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.vectorRepo.SearchSimilar(ctx, dsID, queryVector, topK)
}

// SearchSimilarByType performs vector search filtered by entity type
func (s *LakebaseService) SearchSimilarByType(ctx context.Context, dsID int64, entityType lakebase.EntityType, queryVector []float32, topK int) ([]*lakebase.EmbeddingWithDistance, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.vectorRepo.SearchSimilarByType(ctx, dsID, entityType, queryVector, topK)
}

// GenerateAndSaveEmbeddings generates embeddings for schema and context, then saves to database
func (s *LakebaseService) GenerateAndSaveEmbeddings(ctx context.Context, dsID int64) (*EmbeddingGenerationResult, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}

	if s.embeddingProvider == nil {
		return nil, fmt.Errorf("embedding provider not configured")
	}

	log.Printf("[Embedding] Starting embedding generation for datasource %d using provider: %s", dsID, s.embeddingProvider.Name())

	result := &EmbeddingGenerationResult{
		DatasourceID: dsID,
	}

	var embeddings []*lakebase.Embedding

	// 1. Generate embeddings for tables from rc_tables
	tables, err := s.repo.GetTablesByDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	for _, table := range tables {
		// Build embedding text from table description
		var embText string
		if table.Description.Valid && table.Description.String != "" {
			embText = fmt.Sprintf("Table %s: %s", table.TableName, table.Description.String)
		} else {
			embText = fmt.Sprintf("Table %s", table.TableName)
		}

		tableVector, err := s.embeddingProvider.Embed(ctx, embText)
		if err == nil {
			embeddings = append(embeddings, &lakebase.Embedding{
				DatasourceID:   dsID,
				EntityType:     lakebase.EntityTypeTable,
				EntityID:       table.ID,
				EntityText:     embText,
				Embedding:      tableVector,
				EmbeddingModel: s.config.Embedding.Model,
			})
			result.TablesProcessed++
		}
	}

	// 2. Generate embeddings for columns from rc_columns
	columns, err := s.repo.GetColumnsByDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	for _, col := range columns {
		// Build embedding text from column info
		var embText string
		if col.Description.Valid && col.Description.String != "" {
			embText = fmt.Sprintf("Column %s.%s (%s): %s", col.TableName, col.ColumnName, col.DataType.String, col.Description.String)
		} else {
			embText = fmt.Sprintf("Column %s.%s (%s)", col.TableName, col.ColumnName, col.DataType.String)
		}

		// Add sample values if available
		if col.SampleValues.Valid && col.SampleValues.String != "" {
			embText += fmt.Sprintf(". Sample values: %s", col.SampleValues.String)
		}

		colVector, err := s.embeddingProvider.Embed(ctx, embText)
		if err == nil {
			embeddings = append(embeddings, &lakebase.Embedding{
				DatasourceID:   dsID,
				EntityType:     lakebase.EntityTypeColumn,
				EntityID:       col.ID,
				EntityText:     embText,
				Embedding:      colVector,
				EmbeddingModel: s.config.Embedding.Model,
			})
			result.ColumnsProcessed++
		}
	}

	// 3. Generate embeddings for business terms from rc_terms
	terms, err := s.repo.GetTermsByDatasource(ctx, dsID)
	if err == nil {
		for _, term := range terms {
			embText := fmt.Sprintf("Term '%s': %s", term.Term, term.Definition)
			if term.Synonyms.Valid && term.Synonyms.String != "" {
				embText += fmt.Sprintf(". Synonyms: %s", term.Synonyms.String)
			}

			termVector, err := s.embeddingProvider.Embed(ctx, embText)
			if err == nil {
				embeddings = append(embeddings, &lakebase.Embedding{
					DatasourceID:   dsID,
					EntityType:     lakebase.EntityTypeTerm,
					EntityID:       term.ID,
					EntityText:     embText,
					Embedding:      termVector,
					EmbeddingModel: s.config.Embedding.Model,
				})
				result.ContextsProcessed++
			}
		}
	}

	// 4. Generate embeddings for business context from rc_business_context (if any)
	businessContexts, err := s.repo.GetContextByDatasource(ctx, dsID)
	if err == nil {
		for _, bc := range businessContexts {
			var contextText string
			switch bc.ContextType {
			case lakebase.ContextTypeSemantic:
				var content lakebase.SemanticContent
				if json.Unmarshal(bc.Content, &content) == nil {
					contextText = content.Description
				}
			case lakebase.ContextTypeBusinessRule:
				var content lakebase.BusinessRuleContent
				if json.Unmarshal(bc.Content, &content) == nil {
					contextText = fmt.Sprintf("Business rules: %v", content.Rules)
				}
			case lakebase.ContextTypeEnumMeaning:
				var content lakebase.EnumMeaningContent
				if json.Unmarshal(bc.Content, &content) == nil {
					contextText = fmt.Sprintf("Enum values for %s.%s: %v", bc.TableName, bc.ColumnName.String, content.Values)
				}
			}

			if contextText != "" {
				bcVector, err := s.embeddingProvider.Embed(ctx, contextText)
				if err == nil {
					embeddings = append(embeddings, &lakebase.Embedding{
						DatasourceID:   dsID,
						EntityType:     lakebase.EntityTypeContext,
						EntityID:       bc.ID,
						EntityText:     contextText,
						Embedding:      bcVector,
						EmbeddingModel: s.config.Embedding.Model,
					})
					result.ContextsProcessed++
				}
			}
		}
	}

	// Clear old embeddings and save new ones
	log.Printf("[Embedding] Prepared %d embeddings (tables: %d, columns: %d, contexts: %d)",
		len(embeddings), result.TablesProcessed, result.ColumnsProcessed, result.ContextsProcessed)

	if len(embeddings) > 0 {
		// Delete existing embeddings for this datasource first
		if err := s.vectorRepo.DeleteEmbeddingsByDatasource(ctx, dsID); err != nil {
			log.Printf("[Embedding] Warning: failed to delete old embeddings: %v", err)
		}

		if err := s.vectorRepo.SaveEmbeddingBatch(ctx, embeddings); err != nil {
			log.Printf("[Embedding] Error: failed to save embeddings: %v", err)
			return nil, fmt.Errorf("failed to save embeddings: %w", err)
		}
		result.TotalEmbeddings = len(embeddings)
		log.Printf("[Embedding] Successfully saved %d embeddings for datasource %d", len(embeddings), dsID)
	} else {
		log.Printf("[Embedding] Warning: No embeddings to save for datasource %d", dsID)
	}

	return result, nil
}

// EmbeddingGenerationResult holds the result of embedding generation
type EmbeddingGenerationResult struct {
	DatasourceID      int64 `json:"datasource_id"`
	TablesProcessed   int   `json:"tables_processed"`
	ColumnsProcessed  int   `json:"columns_processed"`
	ContextsProcessed int   `json:"contexts_processed"`
	TotalEmbeddings   int   `json:"total_embeddings"`
}

// CountEmbeddings returns the count of embeddings for a datasource
func (s *LakebaseService) CountEmbeddings(ctx context.Context, dsID int64) (int64, error) {
	if !s.connected {
		return 0, fmt.Errorf("lakebase service: not connected")
	}
	return s.vectorRepo.CountEmbeddingsByDatasource(ctx, dsID)
}

// ===========================================
// Change Log Operations
// ===========================================

// CreateChangeLog creates a change log entry
func (s *LakebaseService) CreateChangeLog(ctx context.Context, log *lakebase.ChangeLog) (int64, error) {
	if !s.connected {
		return 0, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.CreateChangeLog(ctx, log)
}

// GetChangeLogsByDatasource retrieves change logs for a datasource
func (s *LakebaseService) GetChangeLogsByDatasource(ctx context.Context, dsID int64, limit int) ([]*lakebase.ChangeLog, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}
	return s.repo.GetChangeLogsByDatasource(ctx, dsID, limit)
}

// ===========================================
// High-Level Onboarding Operations
// ===========================================

// OnboardingResult holds the result of an onboarding operation
type OnboardingResult struct {
	DatasourceID    int64  `json:"datasource_id"`
	DatasourceName  string `json:"datasource_name"`
	TablesCount     int    `json:"tables_count"`
	ColumnsCount    int    `json:"columns_count"`
	ContextCount    int    `json:"context_count"`
	EmbeddingsCount int    `json:"embeddings_count"`
}

// SaveOnboardingData saves all onboarding data to lake-base storage
func (s *LakebaseService) SaveOnboardingData(
	ctx context.Context,
	datasource *lakebase.Datasource,
	schemas []*lakebase.SchemaMetadata,
	contexts []*lakebase.BusinessContext,
	embeddings []*lakebase.Embedding,
) (*OnboardingResult, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}

	result := &OnboardingResult{
		DatasourceName: datasource.Name,
	}

	// 1. Get or create datasource
	ds, err := s.GetOrCreateDatasource(ctx, datasource)
	if err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}
	result.DatasourceID = ds.ID

	// Update datasource_id in all related records
	for _, schema := range schemas {
		schema.DatasourceID = ds.ID
	}
	for _, bc := range contexts {
		bc.DatasourceID = ds.ID
	}
	for _, emb := range embeddings {
		emb.DatasourceID = ds.ID
	}

	// 2. Save schema metadata
	if len(schemas) > 0 {
		if err := s.repo.SaveSchemaMetadata(ctx, schemas); err != nil {
			return nil, fmt.Errorf("failed to save schema metadata: %w", err)
		}
		result.ColumnsCount = len(schemas)
		// Count unique tables
		tableSet := make(map[string]bool)
		for _, schema := range schemas {
			tableSet[schema.TableName] = true
		}
		result.TablesCount = len(tableSet)
	}

	// 3. Save business context
	if len(contexts) > 0 {
		if err := s.repo.SaveBusinessContextBatch(ctx, contexts); err != nil {
			return nil, fmt.Errorf("failed to save business context: %w", err)
		}
		result.ContextCount = len(contexts)
	}

	// 4. Save embeddings
	if len(embeddings) > 0 {
		if err := s.vectorRepo.SaveEmbeddingBatch(ctx, embeddings); err != nil {
			return nil, fmt.Errorf("failed to save embeddings: %w", err)
		}
		result.EmbeddingsCount = len(embeddings)
	}

	// 5. Update last sync time
	_ = s.repo.UpdateDatasourceLastSync(ctx, ds.ID)

	// 6. Create change log for onboarding
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"tables":     result.TablesCount,
		"columns":    result.ColumnsCount,
		"contexts":   result.ContextCount,
		"embeddings": result.EmbeddingsCount,
	})

	_, _ = s.repo.CreateChangeLog(ctx, &lakebase.ChangeLog{
		DatasourceID:  ds.ID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceSystem,
		ChangeReason:  "Initial onboarding",
	})

	return result, nil
}

// ===========================================
// Statistics Operations
// ===========================================

// LakebaseStats holds statistics about lake-base storage
type LakebaseStats struct {
	DatasourcesCount int64            `json:"datasources_count"`
	TablesCount      int64            `json:"tables_count"`
	ColumnsCount     int64            `json:"columns_count"`
	ContextCount     int64            `json:"context_count"`
	EmbeddingsCount  int64            `json:"embeddings_count"`
	ChangeLogsCount  int64            `json:"change_logs_count"`
	LastUpdated      time.Time        `json:"last_updated"`
	ByDatasource     map[string]int64 `json:"by_datasource,omitempty"`
}

// GetStats retrieves statistics about lake-base storage
func (s *LakebaseService) GetStats(ctx context.Context) (*LakebaseStats, error) {
	if !s.connected {
		return nil, fmt.Errorf("lakebase service: not connected")
	}

	stats := &LakebaseStats{
		LastUpdated:  time.Now(),
		ByDatasource: make(map[string]int64),
	}

	db, err := s.pool.DB()
	if err != nil {
		return nil, err
	}

	// Count datasources
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rc_datasource")
	row.Scan(&stats.DatasourcesCount)

	// Count distinct tables
	row = db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT table_name) FROM rc_schema_metadata")
	row.Scan(&stats.TablesCount)

	// Count columns
	row = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rc_schema_metadata")
	row.Scan(&stats.ColumnsCount)

	// Count context entries
	row = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rc_business_context WHERE is_expired = 0")
	row.Scan(&stats.ContextCount)

	// Count embeddings
	row = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rc_embeddings")
	row.Scan(&stats.EmbeddingsCount)

	// Count change logs
	row = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rc_change_log")
	row.Scan(&stats.ChangeLogsCount)

	return stats, nil
}

// ===========================================
// Helper Methods for Context Building
// ===========================================

// BuildRichContextContent creates JSON content for rich context from analysis results
func BuildRichContextContent(contextType lakebase.ContextType, data map[string]interface{}) (json.RawMessage, error) {
	return json.Marshal(data)
}

// CreateEnumContext creates an enum_meaning context entry
func CreateEnumContext(dsID int64, tableName, columnName string, values map[string]string) *lakebase.BusinessContext {
	content, _ := lakebase.NewEnumMeaningContent(values)
	return &lakebase.BusinessContext{
		DatasourceID: dsID,
		TableName:    tableName,
		ColumnName:   sql.NullString{String: columnName, Valid: true},
		ContextType:  lakebase.ContextTypeEnumMeaning,
		Content:      content,
		Source:       lakebase.SourceLLM,
		Confidence:   0.8,
		Version:      1,
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}
}

// CreateSemanticContext creates a semantic context entry
func CreateSemanticContext(dsID int64, tableName, columnName, description string, synonyms []string) *lakebase.BusinessContext {
	content, _ := lakebase.NewSemanticContent(description, synonyms, nil)
	colName := sql.NullString{}
	if columnName != "" {
		colName = sql.NullString{String: columnName, Valid: true}
	}
	return &lakebase.BusinessContext{
		DatasourceID: dsID,
		TableName:    tableName,
		ColumnName:   colName,
		ContextType:  lakebase.ContextTypeSemantic,
		Content:      content,
		Source:       lakebase.SourceLLM,
		Confidence:   0.8,
		Version:      1,
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}
}

// CreateBusinessRuleContext creates a business_rule context entry
func CreateBusinessRuleContext(dsID int64, tableName string, rules []string) *lakebase.BusinessContext {
	content, _ := lakebase.NewBusinessRuleContent(rules, nil)
	return &lakebase.BusinessContext{
		DatasourceID: dsID,
		TableName:    tableName,
		ColumnName:   sql.NullString{},
		ContextType:  lakebase.ContextTypeBusinessRule,
		Content:      content,
		Source:       lakebase.SourceLLM,
		Confidence:   0.8,
		Version:      1,
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}
}

// CreateDataQualityContext creates a data_quality context entry
func CreateDataQualityContext(dsID int64, tableName, columnName string, issues []string) *lakebase.BusinessContext {
	content, _ := json.Marshal(lakebase.DataQualityContent{Anomalies: issues})
	return &lakebase.BusinessContext{
		DatasourceID: dsID,
		TableName:    tableName,
		ColumnName:   sql.NullString{String: columnName, Valid: columnName != ""},
		ContextType:  lakebase.ContextTypeDataQuality,
		Content:      content,
		Source:       lakebase.SourceLLM,
		Confidence:   0.9,
		Version:      1,
		CreatedBy:    "system",
		UpdatedBy:    "system",
	}
}

// ===========================================
// Rich Context Update Operations
// ===========================================

// UpdateTableDescription updates the description for a specific table
func (s *LakebaseService) UpdateTableDescription(ctx context.Context, dsID int64, tableName, description, source string, confidence float64) error {
	if !s.IsConnected() {
		return fmt.Errorf("lakebase: service not connected")
	}
	return s.repo.UpdateTableDescription(ctx, dsID, tableName, description, source, confidence)
}

// UpdateColumnDescription updates the description for a specific column
func (s *LakebaseService) UpdateColumnDescription(ctx context.Context, dsID int64, tableName, columnName, description, source string, confidence float64) error {
	if !s.IsConnected() {
		return fmt.Errorf("lakebase: service not connected")
	}
	return s.repo.UpdateColumnDescription(ctx, dsID, tableName, columnName, description, source, confidence)
}

// UpdateColumnSynonyms updates synonyms for a column
func (s *LakebaseService) UpdateColumnSynonyms(ctx context.Context, dsID int64, tableName, columnName, synonyms string) error {
	if !s.IsConnected() {
		return fmt.Errorf("lakebase: service not connected")
	}
	return s.repo.UpdateColumnSynonyms(ctx, dsID, tableName, columnName, synonyms)
}

// Global singleton for lakebase service
var (
	globalLakebaseService     *LakebaseService
	globalLakebaseServiceOnce sync.Once
	globalLakebaseServiceMu   sync.RWMutex
)

// InitGlobalLakebaseService initializes the global lakebase service
func InitGlobalLakebaseService(configPath string) error {
	globalLakebaseServiceMu.Lock()
	defer globalLakebaseServiceMu.Unlock()

	if globalLakebaseService != nil {
		return nil
	}

	svc, err := NewLakebaseService(configPath)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := svc.Connect(ctx); err != nil {
		return err
	}

	globalLakebaseService = svc
	return nil
}

// GetGlobalLakebaseService returns the global lakebase service
func GetGlobalLakebaseService() *LakebaseService {
	globalLakebaseServiceMu.RLock()
	defer globalLakebaseServiceMu.RUnlock()
	return globalLakebaseService
}

// CloseGlobalLakebaseService closes the global lakebase service
func CloseGlobalLakebaseService() error {
	globalLakebaseServiceMu.Lock()
	defer globalLakebaseServiceMu.Unlock()

	if globalLakebaseService == nil {
		return nil
	}

	err := globalLakebaseService.Close()
	globalLakebaseService = nil
	return err
}
