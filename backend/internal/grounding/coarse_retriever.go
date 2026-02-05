package grounding

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
)

// CoarseRetriever performs parallel vector search across multiple signal types
type CoarseRetriever struct {
	vectorRepo *lakebase.MySQLVectorRepository
	embedder   embedding.EmbeddingProvider
	config     CoarseRetrievalConfig
}

// NewCoarseRetriever creates a new coarse retriever
func NewCoarseRetriever(
	vectorRepo *lakebase.MySQLVectorRepository,
	embedder embedding.EmbeddingProvider,
	config CoarseRetrievalConfig,
) *CoarseRetriever {
	return &CoarseRetriever{
		vectorRepo: vectorRepo,
		embedder:   embedder,
		config:     config,
	}
}

// RetrievalRequest represents a retrieval request
type RetrievalRequest struct {
	Query        string
	DatasourceID int64
	SignalTypes  []SignalType // if empty, search all types
}

// RetrievalResult represents the coarse retrieval result
type RetrievalResult struct {
	Signals       []*RetrievalSignal
	TotalProbed   int
	Duration      time.Duration
	QueryVector   []float32
	ExecutionLogs []ExecutionLog // Logs for transparency
}

// Retrieve performs speculative parallel retrieval across all signal types
func (r *CoarseRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()

	// Generate query embedding
	queryVector, err := r.embedder.Embed(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Determine which signal types to search
	signalTypes := req.SignalTypes
	if len(signalTypes) == 0 {
		signalTypes = []SignalType{
			SignalTypeTable,
			SignalTypeColumn,
			SignalTypeContext,
			SignalTypeSQLTemplate,
		}
	}

	// Speculative parallel retrieval
	var wg sync.WaitGroup
	type retrievalResultWithLog struct {
		signals []*RetrievalSignal
		log     ExecutionLog
	}
	resultCh := make(chan retrievalResultWithLog, len(signalTypes))
	errCh := make(chan error, len(signalTypes))

	for _, st := range signalTypes {
		wg.Add(1)
		go func(signalType SignalType) {
			defer wg.Done()
			
			searchStart := time.Now()
			signals, err := r.retrieveByType(ctx, req.DatasourceID, signalType, queryVector)
			searchDuration := time.Since(searchStart)
			
			if err != nil {
				errCh <- err
				return
			}
			
			// Build execution log for transparency
			entityType := mapSignalToEntityType(signalType)
			log := ExecutionLog{
				Phase: "vector_search",
				SQL: fmt.Sprintf("SELECT id, entity_text, VEC_DISTANCE_COSINE(embedding, ?) AS distance FROM rc_embeddings WHERE datasource_id = %d AND entity_type = '%s' ORDER BY distance ASC LIMIT %d",
					req.DatasourceID, entityType, r.config.ProbesPerType),
				Params:      []interface{}{"[query_vector...]"},
				ResultCount: len(signals),
				Duration:    searchDuration,
				Summary:     fmt.Sprintf("Vector search for %s: found %d results in %v", signalType, len(signals), searchDuration.Round(time.Millisecond)),
			}
			
			resultCh <- retrievalResultWithLog{signals: signals, log: log}
		}(st)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultCh)
	close(errCh)

	// Check for errors
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	default:
	}

	// Collect and merge results
	var allSignals []*RetrievalSignal
	var executionLogs []ExecutionLog
	for result := range resultCh {
		allSignals = append(allSignals, result.signals...)
		executionLogs = append(executionLogs, result.log)
	}

	// Sort by score (descending)
	sort.Slice(allSignals, func(i, j int) bool {
		return allSignals[i].Score > allSignals[j].Score
	})

	// Apply max signals limit
	if len(allSignals) > r.config.MaxSignals {
		allSignals = allSignals[:r.config.MaxSignals]
	}

	return &RetrievalResult{
		Signals:       allSignals,
		TotalProbed:   len(allSignals),
		Duration:      time.Since(start),
		QueryVector:   queryVector,
		ExecutionLogs: executionLogs,
	}, nil
}

// retrieveByType searches for signals of a specific type
func (r *CoarseRetriever) retrieveByType(
	ctx context.Context,
	dsID int64,
	signalType SignalType,
	queryVector []float32,
) ([]*RetrievalSignal, error) {
	// Map signal type to entity type
	entityType := mapSignalToEntityType(signalType)

	// Perform vector search
	results, err := r.vectorRepo.SearchSimilarByType(ctx, dsID, entityType, queryVector, r.config.ProbesPerType)
	if err != nil {
		return nil, fmt.Errorf("vector search failed for %s: %w", signalType, err)
	}

	// Convert to RetrievalSignal
	var signals []*RetrievalSignal
	for _, result := range results {
		score := float32(1.0 - result.Distance)
		if score < r.config.MinScore {
			continue
		}

		signals = append(signals, &RetrievalSignal{
			ID:           result.ID,
			SignalType:   signalType,
			DatasourceID: dsID,
			EntityName:   result.EntityText,
			Content:      result.EntityText,
			Distance:     float32(result.Distance),
			Score:        score,
		})
	}

	return signals, nil
}

// mapSignalToEntityType maps signal type to lakebase entity type
func mapSignalToEntityType(signalType SignalType) lakebase.EntityType {
	switch signalType {
	case SignalTypeTable:
		return lakebase.EntityTypeTable
	case SignalTypeColumn:
		return lakebase.EntityTypeColumn
	case SignalTypeContext:
		return lakebase.EntityTypeContext
	case SignalTypeSQLTemplate:
		return lakebase.EntityTypeQuery
	case SignalTypeDomainKnowledge:
		return lakebase.EntityTypeContext
	case SignalTypeRelationship:
		return lakebase.EntityTypeRelationship
	default:
		return lakebase.EntityTypeTable
	}
}
