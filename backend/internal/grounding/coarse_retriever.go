package grounding

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// columnEntityRe parses "Column <table>.<column> (<type>): description..." from entity_text
var columnEntityRe = regexp.MustCompile(`^Column\s+(\S+)\.(\S+)\s+\(([^)]+)\)(?::\s*(.+))?`)

// parseColumnEntity extracts (table, column, dataType, description) from a column embedding text.
// The description includes any trailing ". Sample values: ..." text.
func parseColumnEntity(entityText string) (table, column, dataType, description string, ok bool) {
	m := columnEntityRe.FindStringSubmatch(entityText)
	if len(m) < 4 {
		return "", "", "", "", false
	}
	desc := ""
	if len(m) >= 5 && m[4] != "" {
		desc = strings.TrimSpace(m[4])
	}
	return m[1], m[2], m[3], desc, true
}

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

// RetrievalProgressCallback is called when a single signal-type SQL query completes.
// signalType: which type just finished; signals: results for that type; log: execution log.
type RetrievalProgressCallback func(signalType SignalType, signals []*RetrievalSignal, log ExecutionLog)

// RetrievalRequest represents a retrieval request
type RetrievalRequest struct {
	Query        string
	DatasourceID int64
	SignalTypes  []SignalType // if empty, search all types
	// Optional: called as each signal-type SQL query completes (for incremental SSE push)
	ProgressCallback RetrievalProgressCallback
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
	log := logger.With("component", "coarse_retriever")

	log.Debug("[Retrieve] Starting vector retrieval",
		"query", req.Query,
		"datasource_id", req.DatasourceID,
		"signal_types", fmt.Sprintf("%v", req.SignalTypes),
	)

	// Generate query embedding
	queryVector, err := r.embedder.Embed(ctx, req.Query)
	if err != nil {
		log.Error("[Retrieve] Failed to embed query", "error", err)
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	log.Debug("[Retrieve] Query embedded", "vector_dim", len(queryVector))

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

	// Speculative parallel retrieval — each goroutine pushes its result as soon as it completes.
	// If a ProgressCallback is set, it fires per signal-type so the handler can stream incremental SSE.
	var wg sync.WaitGroup
	type retrievalResultWithLog struct {
		signalType SignalType
		signals    []*RetrievalSignal
		log        ExecutionLog
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
				log.Error("[Retrieve] Signal search failed",
					"signal_type", signalType,
					"error", err,
				)
				errCh <- err
				return
			}

			log.Debug("[Retrieve] Signal search completed",
				"signal_type", signalType,
				"results", len(signals),
				"duration", searchDuration.Round(time.Millisecond),
			)
			
			// Build execution log for transparency
			entityType := mapSignalToEntityType(signalType)
			log := ExecutionLog{
				Phase: "vector_search",
				SQL: fmt.Sprintf("SELECT id, entity_text, VEC_DISTANCE_COSINE(embedding, ?) AS distance FROM rc_embeddings WHERE datasource_id = %d AND entity_type = '%s' ORDER BY distance ASC LIMIT %d",
					req.DatasourceID, entityType, r.config.ProbesPerType),
				Params:      []interface{}{"[query_vector..."},
				ResultCount: len(signals),
				Duration:    searchDuration,
				Summary:     fmt.Sprintf("Vector search for %s: found %d results in %v", signalType, len(signals), searchDuration.Round(time.Millisecond)),
			}

			// Fire per-signal-type callback so handler can push incremental SSE
			if req.ProgressCallback != nil {
				req.ProgressCallback(signalType, signals, log)
			}
			
			resultCh <- retrievalResultWithLog{signalType: signalType, signals: signals, log: log}
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

	// Log top signals for debugging
	for i, sig := range allSignals {
		if i >= 10 {
			break
		}
		log.Debug("[Retrieve] Top signal",
			"rank", i+1,
			"type", sig.SignalType,
			"entity", sig.EntityName,
			"score", fmt.Sprintf("%.4f", sig.Score),
			"distance", fmt.Sprintf("%.4f", sig.Distance),
		)
	}

	// Apply max signals limit
	if len(allSignals) > r.config.MaxSignals {
		allSignals = allSignals[:r.config.MaxSignals]
	}

	log.Debug("[Retrieve] Completed",
		"total_signals", len(allSignals),
		"duration", time.Since(start).Round(time.Millisecond),
	)

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

		sig := &RetrievalSignal{
			ID:           result.ID,
			SignalType:   signalType,
			DatasourceID: dsID,
			EntityName:   result.EntityText,
			Content:      result.EntityText,
			Distance:     float32(result.Distance),
			Score:        score,
		}

		// For column signals, parse "Column table.col (type): desc..." to populate SourceTable/SourceColumn
		if signalType == SignalTypeColumn {
			if tbl, col, dataType, desc, ok := parseColumnEntity(result.EntityText); ok {
				sig.SourceTable = tbl
				sig.SourceColumn = col
				sig.Metadata = dataType // store parsed data type
				if desc != "" {
					sig.Content = desc // store RC description (without prefix)
				}
				sig.EntityName = tbl + "." + col // normalise to "table.column"
			}
		}

		signals = append(signals, sig)
	}

	// Deduplicate column signals by table.column — keep highest score
	if signalType == SignalTypeColumn {
		signals = deduplicateColumnSignals(signals)
	}

	return signals, nil
}

// deduplicateColumnSignals keeps only the highest-score signal per table.column pair.
func deduplicateColumnSignals(signals []*RetrievalSignal) []*RetrievalSignal {
	best := make(map[string]*RetrievalSignal) // key = "table.column"
	for _, sig := range signals {
		key := sig.SourceTable + "." + sig.SourceColumn
		if key == "." {
			// Cannot parse — keep as-is via a unique key
			key = fmt.Sprintf("__unparsed_%d", sig.ID)
		}
		if existing, ok := best[key]; !ok || sig.Score > existing.Score {
			best[key] = sig
		}
	}

	deduped := make([]*RetrievalSignal, 0, len(best))
	for _, sig := range best {
		deduped = append(deduped, sig)
	}
	// Sort by score descending so results are deterministic
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].Score > deduped[j].Score
	})
	return deduped
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
