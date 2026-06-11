package grounding

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/tmc/langchaingo/llms"

	"atlas/internal/adapter"
	"atlas/internal/embedding"
	"atlas/internal/lakebase"
	"atlas/internal/logger"
)

// AdaptivePipeline implements the adaptive grounding strategy
// - Small scale (≤ threshold tables): pass all schema directly to linking agent
// - Large scale (> threshold tables): vector retrieval → narrow candidates → linking agent
type AdaptivePipeline struct {
	coarseRetriever *CoarseRetriever
	linkingAgent    *LinkingAgent
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
	// LinkingMode controls the LLM linking agent behaviour.
	// "off"      — skip linking, use vector retrieval results directly (LargeScale only)
	// "one-shot" — single LLM call (default, current behaviour)
	// "react"    — multi-step ReAct linking with execute_sql tool
	LinkingMode string
	// DBAdapter is required when LinkingMode == "react" (for the execute_sql tool).
	DBAdapter adapter.DBAdapter
	// ForceSmallScale forces SmallScale strategy regardless of table count.
	// Used for ablation study: full schema injection to LLM without vector retrieval.
	ForceSmallScale bool
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
	log := logger.With("component", "adaptive_grounding")

	strategy := p.detectStrategy(req)
	log.Info("[Ground] Strategy selected",
		"strategy", string(strategy),
		"query", req.Query,
		"datasource_id", req.DatasourceID,
		"table_count", req.TableCount,
		"threshold", p.config.ScaleThreshold,
	)

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
	// Per-request override (for ablation study)
	if req.ForceSmallScale {
		return StrategySmallScale
	}

	// Config-level override
	if p.config.LinkingAgent.Strategy != "" && p.config.LinkingAgent.Strategy != StrategyAuto {
		return p.config.LinkingAgent.Strategy
	}

	if req.TableCount <= p.config.ScaleThreshold {
		return StrategySmallScale
	}
	return StrategyLargeScale
}

// groundSmallScale: all tables pass through LinkAsync with immediate schema injection.
// No vector retrieval needed — schema data is written to the slot before the agent starts.
// This ensures SmallScale and LargeScale share the same concurrent architecture.
func (p *AdaptivePipeline) groundSmallScale(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log := logger.With("component", "adaptive_grounding")
	log.Info("[SmallScale] Passing all tables to linking agent",
		"table_count", len(req.AllSchemas),
	)

	// Log table names for debugging
	for _, s := range req.AllSchemas {
		log.Debug("[SmallScale] Table in schema",
			"table", s.TableName,
			"columns", len(s.Columns),
			"description", s.Description,
		)
	}

	// Small-scale: immediately push all tables/columns so the frontend
	// can show the "schema loaded" card without waiting for the LLM.
	if req.ProgressCallback != nil {
		req.ProgressCallback("schema_loaded", map[string]interface{}{
			"message":     "Schema loaded (small scale — no vector search needed)",
			"table_count": len(req.AllSchemas),
			"schemas":     req.AllSchemas,
			"strategy":    string(StrategySmallScale),
		})
	}

	// LinkingMode=off: return all tables without LLM call
	if req.LinkingMode == "off" {
		log.Info("[SmallScale] LinkingMode=off, returning all tables without filtering")
		allTables := make([]SelectedTable, 0, len(req.AllSchemas))
		for _, s := range req.AllSchemas {
			allTables = append(allTables, SelectedTable{
				Name:       s.TableName,
				Reason:     "All tables included (linking off)",
				Confidence: 1.0,
			})
		}
		groundedCtx := p.buildGroundedContext(req.Query, &LinkingResult{
			SelectedTables: allTables,
			Reasoning:      "Linking agent skipped — all tables included.",
			Duration:       0,
		}, req.AllSchemas)

		if req.ProgressCallback != nil {
			req.ProgressCallback("linking_done", map[string]interface{}{
				"selected_tables": allTables,
				"reasoning":       "Linking agent skipped — all tables included.",
				"duration_ms":     int64(0),
				"context":         groundedCtx,
			})
		}

		return &AdaptiveGroundingResult{
			Strategy:       StrategySmallScale,
			SelectedTables: allTables,
			Context:        groundedCtx,
			TotalDuration:  time.Since(start),
			Reasoning:      "Linking agent skipped — all tables included.",
		}, nil
	}

	// --- One-shot mode: single LLM call (no ReAct overhead) ---
	if req.LinkingMode != "react" {
		// Build indexed schema text — injected directly into prompt
		tempReq := &LinkingRequest{
			Query:   req.Query,
			Schemas: req.AllSchemas,
			Compact: false, // Full RC for quality hints
		}
		indexedResult := p.linkingAgent.buildIndexedLinkingPrompt(tempReq)

		// Notify: linking agent starting
		if req.ProgressCallback != nil {
			req.ProgressCallback("linking_start", map[string]interface{}{
				"message":     "Linking agent analyzing schema...",
				"table_count": len(req.AllSchemas),
			})
		}

		log.Info("[SmallScale] Calling LinkDirect (one-shot, single LLM call)",
			"query", req.Query,
			"schema_count", len(req.AllSchemas),
		)

		linkResult, err := p.linkingAgent.LinkDirect(ctx, indexedResult.Prompt, &indexedResult)
		if err != nil {
			log.Error("[SmallScale] LinkDirect failed", "error", err)
			return nil, fmt.Errorf("linking agent failed: %w", err)
		}

		log.Info("[SmallScale] LinkDirect completed",
			"selected_tables", len(linkResult.SelectedTables),
			"total_duration", linkResult.Duration.Round(time.Millisecond),
			"reasoning", linkResult.Reasoning,
		)

		// Build grounded context before the callback so we can push the full result
		groundedCtx := p.buildGroundedContext(req.Query, linkResult, req.AllSchemas)

		// Notify: linking agent done
		if req.ProgressCallback != nil {
			req.ProgressCallback("linking_done", map[string]interface{}{
				"selected_tables":   linkResult.SelectedTables,
				"reasoning":         linkResult.Reasoning,
				"duration_ms":       linkResult.Duration.Milliseconds(),
				"retrieval_latency": int64(0), // SmallScale: no vector retrieval
				"reasoning_latency": linkResult.ReasoningLatency.Milliseconds(),
				"context":           groundedCtx,
			})
		}

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
					Summary:     fmt.Sprintf("Small-scale direct linking: %d tables → selected %d in %v", len(req.AllSchemas), len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond)),
					ResultCount: len(linkResult.SelectedTables),
					Duration:    linkResult.Duration,
				},
			},
		}, nil
	}

	// --- React mode: full ReAct engine with tools ---
	// SmallScale: schema is available immediately (no vector retrieval needed).
	// We pre-build the schema text and write it to schemaSlot before starting LinkAsync,
	// so the agent's first get_candidate_schema call returns data instantly.

	var schemaSlot atomic.Pointer[string]
	var indexSlot atomic.Pointer[IndexedPromptResult]

	// Build indexed schema text and write to slot immediately
	tempReq := &LinkingRequest{
		Query:   req.Query,
		Schemas: req.AllSchemas,
		Compact: false, // Full RC for quality hints
	}
	indexedResult := p.linkingAgent.buildIndexedLinkingPrompt(tempReq)
	schemaSlot.Store(&indexedResult.Prompt)
	indexSlot.Store(&indexedResult)

	// Notify: linking agent starting (react mode)
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_start", map[string]interface{}{
			"message":     "ReAct linking agent analyzing schema...",
			"table_count": len(req.AllSchemas),
		})
	}

	log.Info("[SmallScale] Calling LinkAsync (react mode)",
		"query", req.Query,
		"schema_count", len(req.AllSchemas),
		"linking_mode", req.LinkingMode,
	)

	// Build step callback
	var reactStepCB func(step interface{}, eventType string)
	if req.ProgressCallback != nil {
		reactStepCB = func(step interface{}, eventType string) {
			req.ProgressCallback("linking_step", map[string]interface{}{
				"step":       step,
				"event_type": eventType,
			})
		}
	}

	linkResult, err := p.linkingAgent.LinkAsync(ctx, req.Query, req.LinkingMode, &schemaSlot, &indexSlot, req.DBAdapter, reactStepCB)
	if err != nil {
		log.Error("[SmallScale] LinkAsync failed", "error", err)
		return nil, fmt.Errorf("linking agent failed: %w", err)
	}

	log.Info("[SmallScale] LinkAsync completed",
		"selected_tables", len(linkResult.SelectedTables),
		"total_duration", linkResult.Duration.Round(time.Millisecond),
		"reasoning_latency", linkResult.ReasoningLatency.Round(time.Millisecond),
		"reasoning", linkResult.Reasoning,
	)

	// Build grounded context before the callback so we can push the full result
	groundedCtx := p.buildGroundedContext(req.Query, linkResult, req.AllSchemas)

	// Notify: linking agent done
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables":   linkResult.SelectedTables,
			"reasoning":         linkResult.Reasoning,
			"duration_ms":       linkResult.Duration.Milliseconds(),
			"retrieval_latency": int64(0), // SmallScale: no vector retrieval
			"reasoning_latency": linkResult.ReasoningLatency.Milliseconds(),
			"context":           groundedCtx,
		})
	}

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
				Summary:     fmt.Sprintf("Small-scale react linking: %d tables → selected %d in %v (reasoning: %v)", len(req.AllSchemas), len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond), linkResult.ReasoningLatency.Round(time.Millisecond)),
				ResultCount: len(linkResult.SelectedTables),
				Duration:    linkResult.Duration,
			},
		},
	}, nil
}

// vectorRetrievalResult holds the output of doVectorRetrieval, shared by one-shot and react paths.
type vectorRetrievalResult struct {
	coarseResult    *RetrievalResult
	candidateSchemas []SchemaInfo
	indexedResult   IndexedPromptResult
}

// doVectorRetrieval performs sequential vector retrieval, candidate extraction, schema building.
// Shared by groundLargeScale (one-shot) and groundLargeScaleReact (react goroutine).
func (p *AdaptivePipeline) doVectorRetrieval(
	ctx context.Context,
	req *AdaptiveGroundingRequest,
) (*vectorRetrievalResult, error) {
	log := logger.With("component", "adaptive_grounding")

	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_start", map[string]interface{}{
			"message":     "Vector retrieval starting...",
			"table_count": len(req.AllSchemas),
		})
	}

	var retrievalProgressCb RetrievalProgressCallback
	if req.ProgressCallback != nil {
		retrievalProgressCb = func(signalType SignalType, signals []*RetrievalSignal, execLog ExecutionLog) {
			req.ProgressCallback("retrieval_signal", map[string]interface{}{
				"signal_type":   string(signalType),
				"result_count":  len(signals),
				"duration_ms":   execLog.Duration.Milliseconds(),
				"execution_log": execLog,
				"signals":       signals,
			})
		}
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
		ProgressCallback: retrievalProgressCb,
	})
	if err != nil {
		return nil, err
	}

	// Extract candidates and build schema text
	candidateTableNames := p.extractCandidateTableNames(coarseResult.Signals)
	candidateSchemas := p.filterSchemaByCandidates(req.AllSchemas, candidateTableNames)

	if len(candidateSchemas) < 5 && len(req.AllSchemas) > len(candidateSchemas) {
		candidateSchemas = p.expandCandidates(req.AllSchemas, candidateSchemas, p.config.LinkingAgent.MaxTablesInContext)
	}

	// Build schema description text
	tempReq := &LinkingRequest{
		Query:         req.Query,
		Schemas:       candidateSchemas,
		VectorSignals: coarseResult.Signals,
	}
	indexedResult := p.linkingAgent.buildIndexedLinkingPrompt(tempReq)

	log.Info("[LargeScale] Retrieval complete",
		"candidate_tables", len(candidateSchemas),
		"schema_text_length", len(indexedResult.Prompt),
		"retrieval_time", coarseResult.Duration.Round(time.Millisecond),
	)

	// Notify frontend
	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_done", map[string]interface{}{
			"candidate_tables": len(candidateTableNames),
			"duration_ms":      coarseResult.Duration.Milliseconds(),
			"signals":          coarseResult.Signals,
			"execution_logs":   coarseResult.ExecutionLogs,
			"strategy":         string(StrategyLargeScale),
		})
	}

	return &vectorRetrievalResult{
		coarseResult:     coarseResult,
		candidateSchemas: candidateSchemas,
		indexedResult:    indexedResult,
	}, nil
}

// groundLargeScale dispatches large-scale grounding based on linking mode:
//   - "off"              → sequential retrieval only (no LLM)
//   - "one-shot" / ""    → sequential retrieval → LinkDirect (single LLM call)
//   - "react"            → concurrent retrieval goroutine + LinkAsync (ReAct multi-step)
func (p *AdaptivePipeline) groundLargeScale(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log := logger.With("component", "adaptive_grounding")

	// LinkingMode=off: sequential retrieval only, no LLM
	if req.LinkingMode == "off" {
		return p.groundLargeScaleRetrievalOnly(ctx, req, start)
	}

	// React mode: concurrent retrieval + LinkAsync (multi-step exploration)
	if req.LinkingMode == "react" {
		return p.groundLargeScaleReact(ctx, req, start)
	}

	// --- One-shot mode: sequential retrieval → LinkDirect (single LLM call) ---
	log.Info("[LargeScale] Starting sequential retrieval + LinkDirect (one-shot)",
		"total_tables", len(req.AllSchemas),
		"linking_mode", req.LinkingMode,
	)

	// Step 1: Sequential vector retrieval
	retResult, err := p.doVectorRetrieval(ctx, req)
	if err != nil {
		log.Warn("[LargeScale] Vector retrieval failed, falling back to small-scale", "error", err)
		return p.groundSmallScale(ctx, req, start)
	}

	retrievalLatency := retResult.coarseResult.Duration

	// Step 2: Single LLM call via LinkDirect
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_start", map[string]interface{}{
			"message":     "Linking agent analyzing candidates (one-shot)...",
			"table_count": len(retResult.candidateSchemas),
		})
	}

	log.Info("[LargeScale] Calling LinkDirect (one-shot, single LLM call)",
		"query", req.Query,
		"candidate_count", len(retResult.candidateSchemas),
	)

	linkResult, err := p.linkingAgent.LinkDirect(ctx, retResult.indexedResult.Prompt, &retResult.indexedResult)
	if err != nil {
		log.Warn("[LargeScale] LinkDirect failed, falling back to coarse results", "error", err)
		return &AdaptiveGroundingResult{
			Strategy:       StrategyLargeScale,
			SelectedTables: p.signalsToSelectedTables(retResult.coarseResult.Signals),
			Context:        p.signalsOnlyContext(retResult.coarseResult.Signals, req.Query),
			TotalDuration:  time.Since(start),
			RetrievalTime:  retrievalLatency,
			CoarseSignals:  retResult.coarseResult.Signals,
			ExecutionLogs:  retResult.coarseResult.ExecutionLogs,
		}, nil
	}

	linkResult.RetrievalLatency = retrievalLatency

	log.Info("[LargeScale] LinkDirect completed",
		"selected_tables", len(linkResult.SelectedTables),
		"total_duration", linkResult.Duration.Round(time.Millisecond),
		"reasoning", linkResult.Reasoning,
	)

	// Build grounded context
	groundedCtx := p.buildGroundedContext(req.Query, linkResult, retResult.candidateSchemas)

	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables":   linkResult.SelectedTables,
			"reasoning":         linkResult.Reasoning,
			"duration_ms":       linkResult.Duration.Milliseconds(),
			"retrieval_latency": retrievalLatency.Milliseconds(),
			"reasoning_latency": linkResult.ReasoningLatency.Milliseconds(),
			"context":           groundedCtx,
		})
	}

	executionLogs := append(retResult.coarseResult.ExecutionLogs, ExecutionLog{
		Phase:       "linking_agent_direct",
		Summary:     fmt.Sprintf("Large-scale one-shot linking: %d candidates → selected %d in %v (retrieval: %v, reasoning: %v)", len(retResult.candidateSchemas), len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond), retrievalLatency.Round(time.Millisecond), linkResult.ReasoningLatency.Round(time.Millisecond)),
		ResultCount: len(linkResult.SelectedTables),
		Duration:    linkResult.Duration,
	})

	return &AdaptiveGroundingResult{
		Strategy:       StrategyLargeScale,
		SelectedTables: linkResult.SelectedTables,
		Context:        groundedCtx,
		TotalDuration:  time.Since(start),
		RetrievalTime:  retrievalLatency,
		LinkingTime:    linkResult.Duration,
		CoarseSignals:  retResult.coarseResult.Signals,
		ExecutionLogs:  executionLogs,
		Reasoning:      linkResult.Reasoning,
	}, nil
}

// groundLargeScaleReact: concurrent vector retrieval + LinkAsync agent (react mode only).
//   - Goroutine 1: vector retrieval → builds schema text → writes to schemaSlot (T0.1 → T1)
//   - Main thread: LinkAsync starts at T0, polls get_candidate_schema until schema arrives (T1.1)
//
// Latency metrics:
//
//	retrieval_latency = T0 → T1 (vector retrieval wall-clock)
//	reasoning_latency = T1.1 → T2 (LLM reasoning after schema delivered)
//	total_latency = T0 → T2 (end-to-end)
//	overlap_savings = retrieval_latency + reasoning_latency - total_latency
func (p *AdaptivePipeline) groundLargeScaleReact(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log := logger.With("component", "adaptive_grounding")

	log.Info("[LargeScale] Starting concurrent retrieval + LinkAsync (react)",
		"total_tables", len(req.AllSchemas),
	)

	// Shared memory slot: retrieval goroutine writes schema text, LLM tool reads it
	var schemaSlot atomic.Pointer[string]
	var indexSlot atomic.Pointer[IndexedPromptResult]

	// Track retrieval results for building final output
	type retrievalResultAsync struct {
		coarseResult *RetrievalResult
		schemas      []SchemaInfo
		err          error
	}
	retrievalCh := make(chan retrievalResultAsync, 1)

	// --- Goroutine 1: Vector retrieval (T0.1 → T1) ---
	go func() {
		retResult, err := p.doVectorRetrieval(ctx, req)
		if err != nil {
			retrievalCh <- retrievalResultAsync{err: err}
			return
		}

		// Write to shared slots for LinkAsync polling
		schemaSlot.Store(&retResult.indexedResult.Prompt)
		indexSlot.Store(&retResult.indexedResult)

		retrievalCh <- retrievalResultAsync{
			coarseResult: retResult.coarseResult,
			schemas:      retResult.candidateSchemas,
		}
	}()

	// --- Main thread: LinkAsync starts immediately (T0) ---
	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_start", map[string]interface{}{
			"message":     "ReAct linking agent starting (concurrent with retrieval)...",
			"table_count": len(req.AllSchemas),
		})
	}

	var reactStepCB func(step interface{}, eventType string)
	if req.ProgressCallback != nil {
		reactStepCB = func(step interface{}, eventType string) {
			req.ProgressCallback("linking_step", map[string]interface{}{
				"step":       step,
				"event_type": eventType,
			})
		}
	}

	linkResult, linkErr := p.linkingAgent.LinkAsync(ctx, req.Query, req.LinkingMode, &schemaSlot, &indexSlot, req.DBAdapter, reactStepCB)

	// Wait for retrieval to finish (it may already be done)
	retResult := <-retrievalCh

	// Calculate retrieval latency (T0 → T1)
	var retrievalLatency time.Duration
	if retResult.coarseResult != nil {
		retrievalLatency = retResult.coarseResult.Duration
	}

	// Inject retrieval_latency into linkResult
	if linkResult != nil {
		linkResult.RetrievalLatency = retrievalLatency
	}

	// Handle retrieval failure
	if retResult.err != nil {
		log.Warn("[LargeScale] Vector retrieval failed", "error", retResult.err)
		if linkErr != nil {
			// Both failed — fall back to small scale
			return p.groundSmallScale(ctx, req, start)
		}
		// LLM succeeded despite retrieval failure — unlikely but handle gracefully
	}

	// Handle linking failure
	if linkErr != nil {
		log.Warn("[LargeScale] LinkAsync failed, falling back to coarse results", "error", linkErr)
		if retResult.coarseResult != nil {
			return &AdaptiveGroundingResult{
				Strategy:       StrategyLargeScale,
				SelectedTables: p.signalsToSelectedTables(retResult.coarseResult.Signals),
				Context:        p.signalsOnlyContext(retResult.coarseResult.Signals, req.Query),
				TotalDuration:  time.Since(start),
				RetrievalTime:  retrievalLatency,
				CoarseSignals:  retResult.coarseResult.Signals,
				ExecutionLogs:  retResult.coarseResult.ExecutionLogs,
			}, nil
		}
		return p.groundSmallScale(ctx, req, start)
	}

	// Both succeeded — build result
	var coarseSignals []*RetrievalSignal
	var executionLogs []ExecutionLog
	candidateSchemas := retResult.schemas

	if retResult.coarseResult != nil {
		coarseSignals = retResult.coarseResult.Signals
		executionLogs = retResult.coarseResult.ExecutionLogs
	}

	groundedCtx := p.buildGroundedContext(req.Query, linkResult, candidateSchemas)

	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables":    linkResult.SelectedTables,
			"reasoning":          linkResult.Reasoning,
			"duration_ms":        linkResult.Duration.Milliseconds(),
			"retrieval_latency":  retrievalLatency.Milliseconds(),
			"reasoning_latency":  linkResult.ReasoningLatency.Milliseconds(),
			"context":            groundedCtx,
		})
	}

	executionLogs = append(executionLogs, ExecutionLog{
		Phase:       "linking_agent_async",
		Summary:     fmt.Sprintf("Concurrent react linking: selected %d tables in %v (retrieval: %v, reasoning: %v)", len(linkResult.SelectedTables), linkResult.Duration.Round(time.Millisecond), retrievalLatency.Round(time.Millisecond), linkResult.ReasoningLatency.Round(time.Millisecond)),
		ResultCount: len(linkResult.SelectedTables),
		Duration:    linkResult.Duration,
	})

	return &AdaptiveGroundingResult{
		Strategy:       StrategyLargeScale,
		SelectedTables: linkResult.SelectedTables,
		Context:        groundedCtx,
		TotalDuration:  time.Since(start),
		RetrievalTime:  retrievalLatency,
		LinkingTime:    linkResult.Duration,
		CoarseSignals:  coarseSignals,
		ExecutionLogs:  executionLogs,
		Reasoning:      linkResult.Reasoning,
	}, nil
}

// groundLargeScaleRetrievalOnly handles LinkingMode=off: sequential retrieval only, no LLM.
func (p *AdaptivePipeline) groundLargeScaleRetrievalOnly(ctx context.Context, req *AdaptiveGroundingRequest, start time.Time) (*AdaptiveGroundingResult, error) {
	log := logger.With("component", "adaptive_grounding")
	log.Info("[LargeScale] LinkingMode=off, retrieval only",
		"total_tables", len(req.AllSchemas),
	)

	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_start", map[string]interface{}{
			"message":     "Vector retrieval starting...",
			"table_count": len(req.AllSchemas),
		})
	}

	var retrievalProgressCb RetrievalProgressCallback
	if req.ProgressCallback != nil {
		retrievalProgressCb = func(signalType SignalType, signals []*RetrievalSignal, execLog ExecutionLog) {
			req.ProgressCallback("retrieval_signal", map[string]interface{}{
				"signal_type":   string(signalType),
				"result_count":  len(signals),
				"duration_ms":   execLog.Duration.Milliseconds(),
				"execution_log": execLog,
				"signals":       signals,
			})
		}
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
		ProgressCallback: retrievalProgressCb,
	})
	if err != nil {
		log.Warn("[LargeScale] Vector retrieval failed, falling back to small-scale", "error", err)
		return p.groundSmallScale(ctx, req, start)
	}

	if req.ProgressCallback != nil {
		req.ProgressCallback("retrieval_done", map[string]interface{}{
			"candidate_tables": len(p.extractCandidateTableNames(coarseResult.Signals)),
			"duration_ms":      coarseResult.Duration.Milliseconds(),
			"signals":          coarseResult.Signals,
			"execution_logs":   coarseResult.ExecutionLogs,
			"strategy":         string(StrategyLargeScale),
		})
	}

	if req.ProgressCallback != nil {
		req.ProgressCallback("linking_done", map[string]interface{}{
			"selected_tables": p.signalsToSelectedTables(coarseResult.Signals),
			"reasoning":       "Linking agent skipped — using vector retrieval results directly.",
			"duration_ms":     int64(0),
			"context":         p.signalsOnlyContext(coarseResult.Signals, req.Query),
		})
	}

	return &AdaptiveGroundingResult{
		Strategy:       StrategyLargeScale,
		SelectedTables: p.signalsToSelectedTables(coarseResult.Signals),
		Context:        p.signalsOnlyContext(coarseResult.Signals, req.Query),
		TotalDuration:  time.Since(start),
		RetrievalTime:  coarseResult.Duration,
		CoarseSignals:  coarseResult.Signals,
		ExecutionLogs:  coarseResult.ExecutionLogs,
		Reasoning:      "Linking agent skipped — using vector retrieval results directly.",
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
			Hint:      selected.Hint,
		}

		if schema, ok := schemaMap[selected.Name]; ok {
			tc.Description = schema.Description

			// If linking agent identified relevant columns, use those;
			// otherwise fall back to all columns from schema
			if len(selected.RelevantColumns) > 0 {
				// Use only the columns the linking agent deemed relevant
				relevantSet := make(map[string]RelevantColumn) // name -> RelevantColumn
				for _, rc := range selected.RelevantColumns {
					relevantSet[rc.Name] = rc
				}
				colNames := make([]string, 0, len(selected.RelevantColumns))
				for _, rc := range selected.RelevantColumns {
					colNames = append(colNames, rc.Name)
				}
				tc.Columns = colNames

				// Add column contexts with reasons and hints from linking agent
				for _, col := range schema.Columns {
					if rc, isRelevant := relevantSet[col.Name]; isRelevant {
						ctx.Columns = append(ctx.Columns, ColumnContext{
							TableName:    selected.Name,
							ColumnName:   col.Name,
							DataType:     col.Type,
							Description:  col.Description,
							SampleValues: col.SampleValues,
							Synonyms:     col.Synonyms,
							Relevance:    selected.Confidence,
							Reason:       rc.Reason,
							Hint:         rc.Hint,
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
						TableName:    selected.Name,
						ColumnName:   col.Name,
						DataType:     col.Type,
						Description:  col.Description,
						SampleValues: col.SampleValues,
						Synonyms:     col.Synonyms,
						Relevance:    selected.Confidence,
					})
				}
			}
		}

		ctx.Tables = append(ctx.Tables, tc)
	}

	ctx.SignalsSelected = len(ctx.Tables)

	// Populate Relationships from foreign keys of selected tables
	selectedSet := make(map[string]bool)
	for _, t := range ctx.Tables {
		selectedSet[t.Name] = true
	}
	for _, schema := range schemas {
		if !selectedSet[schema.TableName] {
			continue
		}
		for _, fk := range schema.ForeignKeys {
			// Only include relationships where both sides are selected
			if selectedSet[fk.ReferencedTable] {
				ctx.Relationships = append(ctx.Relationships, RelationshipContext{
					FromTable:  schema.TableName,
					FromColumn: fk.Column,
					ToTable:    fk.ReferencedTable,
					ToColumn:   fk.ReferencedColumn,
					Type:       "foreign_key",
					Confidence: 1.0,
				})
			}
		}
	}

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
