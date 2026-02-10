package grounding

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
)

// AdaptivePipeline implements the adaptive grounding strategy
// - Small scale (≤ threshold tables): pass all schema directly to linking agent
// - Large scale (> threshold tables): vector retrieval → narrow candidates → linking agent
type AdaptivePipeline struct {
	coarseRetriever *CoarseRetriever
	linkingAgent    *LinkingAgent
	fineSelector    *FineSelector // Keep for backward compatibility
	config          *GroundingConfig
	vectorRepo      *lakebase.MySQLVectorRepository
	embedder        embedding.EmbeddingProvider
}

// NewAdaptivePipeline creates a new adaptive grounding pipeline
func NewAdaptivePipeline(
	vectorRepo *lakebase.MySQLVectorRepository,
	embedder embedding.EmbeddingProvider,
	llmModel llms.Model,
	config *GroundingConfig,
) *AdaptivePipeline {
	if config == nil {
		config = DefaultGroundingConfig()
	}

	return &AdaptivePipeline{
		coarseRetriever: NewCoarseRetriever(vectorRepo, embedder, config.CoarseRetrieval),
		linkingAgent:    NewLinkingAgent(llmModel, config.LinkingAgent),
		fineSelector:    NewFineSelector(llmModel, config.FineSelection),
		config:          config,
		vectorRepo:      vectorRepo,
		embedder:        embedder,
	}
}

// GroundingProgressCallback is called during adaptive grounding to report progress via SSE.
// stage: "retrieval_done" | "linking_start" | "linking_done"
type GroundingProgressCallback func(stage string, data map[string]interface{})

// AdaptiveGroundingRequest extends GroundingRequest with schema information
type AdaptiveGroundingRequest struct {
	Query        string
	DatasourceID int64
	// Full schema information (loaded from lakebase)
	AllSchemas []SchemaInfo
	// Table count for the datasource (used for strategy detection)
	TableCount int
	// Optional progress callback for SSE streaming
	ProgressCallback GroundingProgressCallback
}

// AdaptiveGroundingResult extends GroundingResult with strategy information
type AdaptiveGroundingResult struct {
	// Strategy used
	Strategy GroundingStrategy `json:"strategy"`
	// Selected tables with reasons
	SelectedTables []SelectedTable `json:"selected_tables"`
	// Full grounded context
	Context *GroundedContext `json:"context"`
	// Timing info
	TotalDuration   time.Duration `json:"total_duration"`
	RetrievalTime   time.Duration `json:"retrieval_time,omitempty"`
	LinkingTime     time.Duration `json:"linking_time"`
	// Coarse signals (only in large scale mode)
	CoarseSignals []*RetrievalSignal `json:"coarse_signals,omitempty"`
	// Execution logs for transparency
	ExecutionLogs []ExecutionLog `json:"execution_logs,omitempty"`
	// Agent reasoning
	Reasoning string `json:"reasoning,omitempty"`
}

// Ground performs adaptive grounding based on schema scale
func (p *AdaptivePipeline) Ground(ctx context.Context, req *AdaptiveGroundingRequest) (*AdaptiveGroundingResult, error) {
	start := time.Now()

	strategy := p.detectStrategy(req)
	log.Printf("[adaptive-grounding] Using strategy: %s (tables: %d, threshold: %d)",
		strategy, req.TableCount, p.config.ScaleThreshold)

	switch strategy {
	case StrategySmallScale:
		return p.groundSmallScale(ctx, req, start)
	case StrategyLargeScale:
		return p.groundLargeScale(ctx, req, start)
	default:
		return p.groundSmallScale(ctx, req, start)
	}
}

// detectStrategy determines the grounding strategy based on schema scale
func (p *AdaptivePipeline) detectStrategy(req *AdaptiveGroundingRequest) GroundingStrategy {
	// Explicit override
	if p.config.LinkingAgent.Strategy != "" && p.config.LinkingAgent.Strategy != StrategyAuto {
		return p.config.LinkingAgent.Strategy
	}

	if req.TableCount <= p.config.ScaleThreshold {
		return StrategySmallScale
	}
	return StrategyLargeScale
}

// groundSmallScale: pass all schema directly to linking agent
// No vector retrieval needed - the LLM sees everything
func (p *AdaptivePipeline) groundSmallScale(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log.Printf("[adaptive-grounding] Small-scale mode: passing all %d tables to linking agent", len(req.AllSchemas))

	// Notify: linking agent starting
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_start", map[string]interface{}{
			"message":     "Linking agent analyzing schema...",
			"table_count": len(req.AllSchemas),
		})
	}

	linkReq := &LinkingRequest{
		Query:   req.Query,
		Schemas: req.AllSchemas,
	}

	linkResult, err := p.linkingAgent.Link(ctx, linkReq)
	if err != nil {
		return nil, fmt.Errorf("linking agent failed: %w", err)
	}

	// Notify: linking agent done
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables": len(linkResult.SelectedTables),
			"reasoning":       linkResult.Reasoning,
			"duration_ms":     linkResult.Duration.Milliseconds(),
		})
	}

	// Build grounded context from linking result
	groundedCtx := p.buildGroundedContext(req.Query, linkResult, req.AllSchemas)

	return &AdaptiveGroundingResult{
		Strategy:       StrategySmallScale,
		SelectedTables: linkResult.SelectedTables,
		Context:        groundedCtx,
		TotalDuration:  time.Since(start),
		LinkingTime:    linkResult.Duration,
		Reasoning:      linkResult.Reasoning,
		ExecutionLogs: []ExecutionLog{
			{
				Phase:       "linking_agent",
				Summary:     fmt.Sprintf("Small-scale linking: %d tables → selected %d in %v", len(req.AllSchemas), len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond)),
				ResultCount: len(linkResult.SelectedTables),
				Duration:    linkResult.Duration,
			},
		},
	}, nil
}

// groundLargeScale: vector retrieval → narrow candidates → linking agent
// Vector retrieval reduces context, then linking agent makes final selection
func (p *AdaptivePipeline) groundLargeScale(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log.Printf("[adaptive-grounding] Large-scale mode: vector retrieval → linking agent (%d total tables)", len(req.AllSchemas))

	// Stage 1: Coarse retrieval to identify candidate tables
	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_start", map[string]interface{}{
			"message":     "Vector retrieval starting...",
			"table_count": len(req.AllSchemas),
		})
	}

	coarseResult, err := p.coarseRetriever.Retrieve(ctx, &RetrievalRequest{
		Query:        req.Query,
		DatasourceID: req.DatasourceID,
		SignalTypes: []SignalType{
			SignalTypeTable,
			SignalTypeColumn,
			SignalTypeContext,
			SignalTypeSQLTemplate,
		},
	})
	if err != nil {
		// Fallback to small scale if retrieval fails
		log.Printf("[adaptive-grounding] Vector retrieval failed, falling back to small-scale: %v", err)
		return p.groundSmallScale(ctx, req, start)
	}

	retrievalTime := coarseResult.Duration

	// Extract candidate table names from signals
	candidateTableNames := p.extractCandidateTableNames(coarseResult.Signals)
	log.Printf("[adaptive-grounding] Vector retrieval identified %d candidate tables", len(candidateTableNames))

	// Build candidate schema set - only include tables hit by retrieval
	// But also respect MaxTablesInContext limit
	candidateSchemas := p.filterSchemaByCandidates(req.AllSchemas, candidateTableNames)

	// Notify: retrieval done
	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_done", map[string]interface{}{
			"candidate_tables": len(candidateTableNames),
			"duration_ms":      coarseResult.Duration.Milliseconds(),
		})
	}

	// If candidates are too few, add more from full schema to avoid missing tables
	if len(candidateSchemas) < 5 && len(req.AllSchemas) > len(candidateSchemas) {
		log.Printf("[adaptive-grounding] Too few candidates (%d), expanding with full schema", len(candidateSchemas))
		candidateSchemas = p.expandCandidates(req.AllSchemas, candidateSchemas, p.config.LinkingAgent.MaxTablesInContext)
	}

	// Stage 2: Linking agent on narrowed candidate set
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_start", map[string]interface{}{
			"message":     "Linking agent analyzing candidates...",
			"table_count": len(candidateSchemas),
		})
	}

	linkReq := &LinkingRequest{
		Query:         req.Query,
		Schemas:       candidateSchemas,
		VectorSignals: coarseResult.Signals,
	}

	linkResult, err := p.linkingAgent.Link(ctx, linkReq)
	if err != nil {
		// Fallback: use coarse results directly
		log.Printf("[adaptive-grounding] Linking agent failed, using coarse results: %v", err)
		return &AdaptiveGroundingResult{
			Strategy:       StrategyLargeScale,
			SelectedTables: p.signalsToSelectedTables(coarseResult.Signals),
			Context:        p.signalsOnlyContext(coarseResult.Signals, req.Query),
			TotalDuration:  time.Since(start),
			RetrievalTime:  retrievalTime,
			CoarseSignals:  coarseResult.Signals,
			ExecutionLogs:  coarseResult.ExecutionLogs,
		}, nil
	}

	// Notify: linking agent done
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables": len(linkResult.SelectedTables),
			"reasoning":       linkResult.Reasoning,
			"duration_ms":     linkResult.Duration.Milliseconds(),
		})
	}

	// Build grounded context
	groundedCtx := p.buildGroundedContext(req.Query, linkResult, candidateSchemas)

	executionLogs := append(coarseResult.ExecutionLogs, ExecutionLog{
		Phase:       "linking_agent",
		Summary:     fmt.Sprintf("Large-scale linking: %d candidates → selected %d in %v", len(candidateSchemas), len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond)),
		ResultCount: len(linkResult.SelectedTables),
		Duration:    linkResult.Duration,
	})

	return &AdaptiveGroundingResult{
		Strategy:       StrategyLargeScale,
		SelectedTables: linkResult.SelectedTables,
		Context:        groundedCtx,
		TotalDuration:  time.Since(start),
		RetrievalTime:  retrievalTime,
		LinkingTime:    linkResult.Duration,
		CoarseSignals:  coarseResult.Signals,
		ExecutionLogs:  executionLogs,
		Reasoning:      linkResult.Reasoning,
	}, nil
}

// extractCandidateTableNames extracts unique table names from retrieval signals
func (p *AdaptivePipeline) extractCandidateTableNames(signals []*RetrievalSignal) map[string]bool {
	tables := make(map[string]bool)
	for _, sig := range signals {
		switch sig.SignalType {
		case SignalTypeTable:
			// EntityName for table signals contains the table entity text
			// We need to extract just the table name
			tables[sig.EntityName] = true
		case SignalTypeColumn:
			// Column signals reference their source table
			if sig.SourceTable != "" {
				tables[sig.SourceTable] = true
			}
		}
	}
	return tables
}

// filterSchemaByCandidates returns only schemas matching candidate table names
func (p *AdaptivePipeline) filterSchemaByCandidates(allSchemas []SchemaInfo, candidates map[string]bool) []SchemaInfo {
	var filtered []SchemaInfo
	for _, schema := range allSchemas {
		if candidates[schema.TableName] {
			filtered = append(filtered, schema)
		}
	}
	return filtered
}

// expandCandidates adds more schemas to reach the desired count
func (p *AdaptivePipeline) expandCandidates(allSchemas []SchemaInfo, existing []SchemaInfo, maxCount int) []SchemaInfo {
	existingNames := make(map[string]bool)
	for _, s := range existing {
		existingNames[s.TableName] = true
	}

	result := make([]SchemaInfo, len(existing))
	copy(result, existing)

	for _, schema := range allSchemas {
		if len(result) >= maxCount {
			break
		}
		if !existingNames[schema.TableName] {
			result = append(result, schema)
			existingNames[schema.TableName] = true
		}
	}
	return result
}

// buildGroundedContext builds GroundedContext from linking result and schemas
func (p *AdaptivePipeline) buildGroundedContext(query string, linkResult *LinkingResult, schemas []SchemaInfo) *GroundedContext {
	ctx := &GroundedContext{
		Query:          query,
		Tables:         make([]TableContext, 0, len(linkResult.SelectedTables)),
		Columns:        make([]ColumnContext, 0),
		GroundingTime:  linkResult.Duration,
		SignalsProbed:  len(schemas),
		Reasoning:      linkResult.Reasoning,
	}

	// Build schema lookup
	schemaMap := make(map[string]SchemaInfo)
	for _, s := range schemas {
		schemaMap[s.TableName] = s
	}

	for _, selected := range linkResult.SelectedTables {
		tc := TableContext{
			Name:      selected.Name,
			Reason:    selected.Reason,
			Relevance: selected.Confidence,
		}

		if schema, ok := schemaMap[selected.Name]; ok {
			tc.Description = schema.Description

			// If linking agent identified relevant columns, use those;
			// otherwise fall back to all columns from schema
			if len(selected.RelevantColumns) > 0 {
				// Use only the columns the linking agent deemed relevant
				relevantSet := make(map[string]string) // name -> reason
				for _, rc := range selected.RelevantColumns {
					relevantSet[rc.Name] = rc.Reason
				}
				colNames := make([]string, 0, len(selected.RelevantColumns))
				for _, rc := range selected.RelevantColumns {
					colNames = append(colNames, rc.Name)
				}
				tc.Columns = colNames

				// Add column contexts with reasons from linking agent
				for _, col := range schema.Columns {
					if reason, isRelevant := relevantSet[col.Name]; isRelevant {
						ctx.Columns = append(ctx.Columns, ColumnContext{
							TableName:   selected.Name,
							ColumnName:  col.Name,
							DataType:    col.Type,
							Description: col.Description,
							Relevance:   selected.Confidence,
							Reason:      reason,
						})
					}
				}
			} else {
				// Fallback: use all columns
				colNames := make([]string, len(schema.Columns))
				for i, col := range schema.Columns {
					colNames[i] = col.Name
				}
				tc.Columns = colNames

				for _, col := range schema.Columns {
					ctx.Columns = append(ctx.Columns, ColumnContext{
						TableName:   selected.Name,
						ColumnName:  col.Name,
						DataType:    col.Type,
						Description: col.Description,
						Relevance:   selected.Confidence,
					})
				}
			}
		}

		ctx.Tables = append(ctx.Tables, tc)
	}

	ctx.SignalsSelected = len(ctx.Tables)
	return ctx
}

// signalsToSelectedTables converts raw signals to SelectedTable (fallback)
func (p *AdaptivePipeline) signalsToSelectedTables(signals []*RetrievalSignal) []SelectedTable {
	seen := make(map[string]bool)
	var tables []SelectedTable
	for _, sig := range signals {
		if sig.SignalType == SignalTypeTable && !seen[sig.EntityName] {
			seen[sig.EntityName] = true
			tables = append(tables, SelectedTable{
				Name:       sig.EntityName,
				Reason:     "Retrieved by vector search",
				Confidence: sig.Score,
			})
		}
	}
	return tables
}

// signalsOnlyContext builds a basic GroundedContext from signals only (fallback)
func (p *AdaptivePipeline) signalsOnlyContext(signals []*RetrievalSignal, query string) *GroundedContext {
	ctx := &GroundedContext{
		Query:          query,
		Tables:         make([]TableContext, 0),
		Columns:        make([]ColumnContext, 0),
		SignalsProbed:  len(signals),
	}

	seenTables := make(map[string]bool)
	for _, sig := range signals {
		if sig.SignalType == SignalTypeTable && !seenTables[sig.EntityName] {
			seenTables[sig.EntityName] = true
			ctx.Tables = append(ctx.Tables, TableContext{
				Name:      sig.EntityName,
				Relevance: sig.Score,
			})
		}
	}
	ctx.SignalsSelected = len(ctx.Tables)
	return ctx
}
