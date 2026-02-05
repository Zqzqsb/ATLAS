package grounding

import (
	"context"
	"fmt"
	"sync"
	"time"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
	"lucid/internal/llm"
)

// Pipeline orchestrates the complete Semantic Grounding process
// combining coarse retrieval and fine selection stages
type Pipeline struct {
	coarseRetriever *CoarseRetriever
	fineSelector    *FineSelector
	config          *GroundingConfig
	mu              sync.RWMutex
}

// NewPipeline creates a new grounding pipeline
func NewPipeline(
	vectorRepo *lakebase.MySQLVectorRepository,
	embedder embedding.EmbeddingProvider,
	llmClient llm.Client,
	config *GroundingConfig,
) *Pipeline {
	if config == nil {
		config = DefaultGroundingConfig()
	}

	return &Pipeline{
		coarseRetriever: NewCoarseRetriever(vectorRepo, embedder, config.CoarseRetrieval),
		fineSelector:    NewFineSelector(llmClient, config.FineSelection),
		config:          config,
	}
}

// GroundingRequest represents a grounding pipeline request
type GroundingRequest struct {
	Query        string
	DatasourceID int64
	// Optional: specify signal types to search
	SignalTypes []SignalType
	// Optional: skip fine selection (return all coarse results)
	SkipFineSelection bool
}

// GroundingResult represents the complete grounding result
type GroundingResult struct {
	Context           *GroundedContext
	CoarseSignals     []*RetrievalSignal
	CoarseDuration    time.Duration
	SelectionDuration time.Duration
	TotalDuration     time.Duration
	Mode              string // "sequential", "parallel", "coarse_only"
	// Execution logs for transparency
	ExecutionLogs     []ExecutionLog `json:"execution_logs,omitempty"`
}

// ExecutionLog represents a SQL execution record for transparency
type ExecutionLog struct {
	Phase       string        `json:"phase"`        // "vector_search", "fine_selection"
	SQL         string        `json:"sql"`          // The SQL query executed
	Params      []interface{} `json:"params"`       // Query parameters
	ResultCount int           `json:"result_count"` // Number of results
	Duration    time.Duration `json:"duration"`     // Execution time
	Summary     string        `json:"summary"`      // Human-readable summary
}

// Ground performs the complete semantic grounding process
func (p *Pipeline) Ground(ctx context.Context, req *GroundingRequest) (*GroundingResult, error) {
	start := time.Now()

	// Stage 1: Coarse-grained retrieval
	coarseResult, err := p.coarseRetriever.Retrieve(ctx, &RetrievalRequest{
		Query:        req.Query,
		DatasourceID: req.DatasourceID,
		SignalTypes:  req.SignalTypes,
	})
	if err != nil {
		return nil, fmt.Errorf("coarse retrieval failed: %w", err)
	}

	result := &GroundingResult{
		CoarseSignals:  coarseResult.Signals,
		CoarseDuration: coarseResult.Duration,
		Mode:           "sequential",
		ExecutionLogs:  coarseResult.ExecutionLogs, // Pass through execution logs
	}

	// If skip fine selection, return coarse results directly
	if req.SkipFineSelection {
		result.Context = p.signalsToContext(coarseResult.Signals, req.Query)
		result.Mode = "coarse_only"
		result.TotalDuration = time.Since(start)
		return result, nil
	}

	// Stage 2: Fine-grained selection
	selectResult, err := p.fineSelector.Select(ctx, &SelectionRequest{
		Query:   req.Query,
		Signals: coarseResult.Signals,
	})
	if err != nil {
		// Fallback to coarse results on fine selection failure
		result.Context = p.signalsToContext(coarseResult.Signals, req.Query)
		result.TotalDuration = time.Since(start)
		return result, nil
	}

	result.Context = selectResult.Context
	result.SelectionDuration = selectResult.Duration
	result.TotalDuration = time.Since(start)

	return result, nil
}

// GroundParallel performs coarse retrieval and starts fine selection in parallel
// Returns coarse results immediately, then updates with fine selection
func (p *Pipeline) GroundParallel(ctx context.Context, req *GroundingRequest) (<-chan *GroundingResult, error) {
	resultCh := make(chan *GroundingResult, 2)

	go func() {
		defer close(resultCh)
		start := time.Now()

		// Stage 1: Coarse retrieval (emit immediately)
		coarseResult, err := p.coarseRetriever.Retrieve(ctx, &RetrievalRequest{
			Query:        req.Query,
			DatasourceID: req.DatasourceID,
			SignalTypes:  req.SignalTypes,
		})
		if err != nil {
			return
		}

		// Emit coarse result immediately with execution logs
		resultCh <- &GroundingResult{
			Context:        p.signalsToContext(coarseResult.Signals, req.Query),
			CoarseSignals:  coarseResult.Signals,
			CoarseDuration: coarseResult.Duration,
			TotalDuration:  time.Since(start),
			Mode:           "parallel_coarse",
			ExecutionLogs:  coarseResult.ExecutionLogs,
		}

		if req.SkipFineSelection {
			return
		}

		// Stage 2: Fine selection (emit when complete)
		selectResult, err := p.fineSelector.Select(ctx, &SelectionRequest{
			Query:   req.Query,
			Signals: coarseResult.Signals,
		})
		if err != nil {
			return
		}

		resultCh <- &GroundingResult{
			Context:           selectResult.Context,
			CoarseSignals:     coarseResult.Signals,
			CoarseDuration:    coarseResult.Duration,
			SelectionDuration: selectResult.Duration,
			TotalDuration:     time.Since(start),
			Mode:              "parallel_fine",
			ExecutionLogs:     coarseResult.ExecutionLogs,
		}
	}()

	return resultCh, nil
}

// signalsToContext converts raw signals to GroundedContext without LLM selection
func (p *Pipeline) signalsToContext(signals []*RetrievalSignal, query string) *GroundedContext {
	ctx := &GroundedContext{
		Query:          query,
		Tables:         make([]TableContext, 0),
		Columns:        make([]ColumnContext, 0),
		SignalsProbed:  len(signals),
	}

	// Group by type and deduplicate
	seenTables := make(map[string]bool)
	seenColumns := make(map[string]bool)

	for _, sig := range signals {
		switch sig.SignalType {
		case SignalTypeTable:
			if !seenTables[sig.EntityName] {
				seenTables[sig.EntityName] = true
				ctx.Tables = append(ctx.Tables, TableContext{
					Name:        sig.EntityName,
					Description: sig.Content,
					Relevance:   sig.Score,
				})
			}
		case SignalTypeColumn:
			if !seenColumns[sig.EntityName] {
				seenColumns[sig.EntityName] = true
				ctx.Columns = append(ctx.Columns, ColumnContext{
					ColumnName:  sig.EntityName,
					Description: sig.Content,
					Relevance:   sig.Score,
				})
			}
		}
	}

	ctx.SignalsSelected = len(ctx.Tables) + len(ctx.Columns)
	return ctx
}

// GetConfig returns the current configuration
func (p *Pipeline) GetConfig() *GroundingConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// UpdateConfig updates the pipeline configuration
func (p *Pipeline) UpdateConfig(config *GroundingConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
	p.coarseRetriever.config = config.CoarseRetrieval
	p.fineSelector.config = config.FineSelection
}
