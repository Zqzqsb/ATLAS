package grounding

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"atlas/internal/embedding"
	"atlas/internal/lakebase"
	"atlas/internal/logger"
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

// Retrieve performs unified vector retrieval and splits results by signal type.
//
// MariaDB HNSW is a global index — adding a WHERE entity_type filter causes the
// index scan to return only the N nearest neighbours overall, then post-filters
// by type.  When one type dominates, the filtered set for minority types can be
// empty.  To work around this we issue a single large SearchSimilar query
// (without type filter) and bucket the results in application code.
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

	// Build a set of desired entity types for quick lookup
	wantedEntityTypes := make(map[lakebase.EntityType]SignalType, len(signalTypes))
	for _, st := range signalTypes {
		et := mapSignalToEntityType(st)
		wantedEntityTypes[et] = st
	}

	// Unified HNSW search — no entity_type filter so the index scan is not
	// prematurely truncated.  We fetch ProbesPerType * len(types) * 2 rows to
	// ensure every type has enough candidates after bucketing.
	totalTopK := r.config.ProbesPerType * len(signalTypes) * 2
	if totalTopK < 200 {
		totalTopK = 200
	}

	searchStart := time.Now()
	results, err := r.vectorRepo.SearchSimilar(ctx, req.DatasourceID, queryVector, totalTopK)
	searchDuration := time.Since(searchStart)
	if err != nil {
		log.Error("[Retrieve] Unified vector search failed", "error", err)
		return nil, fmt.Errorf("unified vector search failed: %w", err)
	}
	log.Debug("[Retrieve] Unified search completed",
		"total_results", len(results),
		"duration", searchDuration.Round(time.Millisecond),
	)

	// Bucket results by entity type → signal type
	buckets := make(map[SignalType][]*RetrievalSignal)
	for _, result := range results {
		signalType, ok := wantedEntityTypes[lakebase.EntityType(result.EntityType)]
		if !ok {
			continue // entity_type not in the requested set
		}

		score := float32(1.0 - result.Distance)
		sig := &RetrievalSignal{
			ID:           result.ID,
			SignalType:   signalType,
			DatasourceID: req.DatasourceID,
			EntityName:   result.EntityText,
			Content:      result.EntityText,
			Distance:     float32(result.Distance),
			Score:        score,
		}

		// For column signals, parse structured entity text
		if signalType == SignalTypeColumn {
			if tbl, col, dataType, desc, ok := parseColumnEntity(result.EntityText); ok {
				sig.SourceTable = tbl
				sig.SourceColumn = col
				sig.Metadata = dataType
				if desc != "" {
					sig.Content = desc
				}
				sig.EntityName = tbl + "." + col
			}
		}

		// Cap per-type at ProbesPerType
		if len(buckets[signalType]) < r.config.ProbesPerType {
			buckets[signalType] = append(buckets[signalType], sig)
		}
	}

	// Post-process: deduplicate columns
	if cols, ok := buckets[SignalTypeColumn]; ok {
		buckets[SignalTypeColumn] = deduplicateColumnSignals(cols)
	}

	// Build per-type execution logs & fire progress callbacks
	var allSignals []*RetrievalSignal
	var executionLogs []ExecutionLog
	for _, st := range signalTypes {
		signals := buckets[st]
		entityType := mapSignalToEntityType(st)

		execLog := ExecutionLog{
			Phase: "vector_search",
			SQL: fmt.Sprintf(
				"SELECT id, entity_text, VEC_DISTANCE_COSINE(embedding, ?) AS distance "+
					"FROM rc_embeddings WHERE datasource_id = %d ORDER BY distance ASC LIMIT %d "+
					"/* bucketed by entity_type = '%s' */",
				req.DatasourceID, totalTopK, entityType),
			Params:      []interface{}{"[query_vector...]"},
			ResultCount: len(signals),
			Duration:    searchDuration / time.Duration(len(signalTypes)), // approximate per-type
			Summary: fmt.Sprintf("Vector search for %s: found %d results (unified query, %v total)",
				st, len(signals), searchDuration.Round(time.Millisecond)),
		}
		executionLogs = append(executionLogs, execLog)

		if req.ProgressCallback != nil {
			req.ProgressCallback(st, signals, execLog)
		}

		allSignals = append(allSignals, signals...)
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
		// Prefer "term" (domain glossary) which has actual embeddings;
		// fall back covers onboarding flows that store as "context".
		return lakebase.EntityTypeTerm
	case SignalTypeSQLTemplate:
		return lakebase.EntityTypeQuery
	case SignalTypeDomainKnowledge:
		return lakebase.EntityTypeTerm
	case SignalTypeRelationship:
		return lakebase.EntityTypeRelationship
	default:
		return lakebase.EntityTypeTable
	}
}
