package bridge

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"

	"lucid/interfaces"
	"lucid/internal/adapter"
	"lucid/internal/inference"
)

// InferenceEngineBridge wraps internal/inference.Pipeline
// to implement interfaces.InferenceEngine
type InferenceEngineBridge struct {
	llm            llms.Model
	adapterFactory *AdapterFactory
	currentModel   string
	modelConfigs   map[string]interfaces.ModelInfo
}

// NewInferenceEngineBridge creates inference engine bridge
func NewInferenceEngineBridge(llm llms.Model, factory *AdapterFactory) *InferenceEngineBridge {
	return &InferenceEngineBridge{
		llm:            llm,
		adapterFactory: factory,
		currentModel:   "default",
		modelConfigs:   make(map[string]interfaces.ModelInfo),
	}
}

// SetLLM updates the LLM model
func (e *InferenceEngineBridge) SetLLM(llm llms.Model) {
	e.llm = llm
}

// Execute runs inference using internal Pipeline
func (e *InferenceEngineBridge) Execute(ctx context.Context, req *interfaces.InferenceRequest) (*interfaces.InferenceResult, error) {
	if e.llm == nil {
		return nil, fmt.Errorf("LLM not initialized")
	}

	// Get database adapter
	dbAdapter, err := e.getAdapter(req.DatabaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter: %w", err)
	}
	defer dbAdapter.Close()

	// Connect to database
	if err := dbAdapter.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Build inference config from request
	config := &inference.Config{
		UseRichContext:  req.UseRichContext,
		UseReact:        req.UseReact,
		ReactLinking:    req.UseReact,
		UseDryRun:       true,
		MaxIterations:   req.MaxIterations,
		ContextFile:     req.ContextFile,
		ClarifyMode:     "off",
		LogMode:         "simple",
		EnableProofread: false,
		DBName:          req.DatabaseID,
		DBType:          dbAdapter.GetDatabaseType(),
	}

	if config.MaxIterations == 0 {
		config.MaxIterations = 5
	}

	// Create and execute pipeline
	pipeline := inference.NewPipeline(e.llm, dbAdapter, config)
	defer pipeline.Reset()

	result, err := pipeline.Execute(ctx, req.Question)
	if err != nil {
		return nil, err
	}

	// Convert internal result to interface result
	return e.convertResult(result), nil
}

// ExecuteStream runs streaming inference
func (e *InferenceEngineBridge) ExecuteStream(ctx context.Context, req *interfaces.InferenceRequest, events chan<- interfaces.StreamEvent) error {
	if e.llm == nil {
		return fmt.Errorf("LLM not initialized")
	}

	// Get database adapter
	dbAdapter, err := e.getAdapter(req.DatabaseID)
	if err != nil {
		return fmt.Errorf("failed to get adapter: %w", err)
	}
	defer dbAdapter.Close()

	// Connect to database
	if err := dbAdapter.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Build inference config
	config := &inference.Config{
		UseRichContext:  req.UseRichContext,
		UseReact:        req.UseReact,
		ReactLinking:    req.UseReact,
		UseDryRun:       true,
		MaxIterations:   req.MaxIterations,
		ContextFile:     req.ContextFile,
		ClarifyMode:     "off",
		LogMode:         "simple",
		EnableProofread: false,
		DBName:          req.DatabaseID,
		DBType:          dbAdapter.GetDatabaseType(),
	}

	if config.MaxIterations == 0 {
		config.MaxIterations = 5
	}

	// Create pipeline with step callback for streaming
	pipeline := inference.NewPipeline(e.llm, dbAdapter, config)
	defer pipeline.Reset()

	// Set streaming callback
	pipeline.SetStepCallback(func(step inference.ReActStep, eventType string) {
		events <- interfaces.StreamEvent{
			Type: interfaces.EventType(eventType),
			Data: interfaces.ReActStep{
				Step:        step.Step,
				Thought:     step.Thought,
				Action:      step.Action,
				ActionInput: step.ActionInput,
				Observation: step.Observation,
				Phase:       step.Phase,
			},
		}
	})

	// Execute
	result, err := pipeline.Execute(ctx, req.Question)
	if err != nil {
		events <- interfaces.StreamEvent{
			Type: interfaces.EventError,
			Data: interfaces.ErrorEventData{Error: err.Error()},
		}
		return err
	}

	// Send final result
	events <- interfaces.StreamEvent{
		Type: interfaces.EventComplete,
		Data: e.convertResult(result),
	}

	return nil
}

// GetAvailableModels returns configured models
func (e *InferenceEngineBridge) GetAvailableModels() []interfaces.ModelInfo {
	models := make([]interfaces.ModelInfo, 0, len(e.modelConfigs))
	for _, m := range e.modelConfigs {
		models = append(models, m)
	}
	if len(models) == 0 {
		// Return default if no models configured
		return []interfaces.ModelInfo{
			{ID: "default", Name: "Default Model", Provider: "unknown", IsDefault: true},
		}
	}
	return models
}

// SwitchModel changes current model
func (e *InferenceEngineBridge) SwitchModel(modelID string) error {
	e.currentModel = modelID
	return nil
}

// GetCurrentModel returns current model ID
func (e *InferenceEngineBridge) GetCurrentModel() string {
	return e.currentModel
}

// GetLLMModel returns underlying LLM model
func (e *InferenceEngineBridge) GetLLMModel() interface{} {
	return e.llm
}

// SetModelConfigs sets available model configurations
func (e *InferenceEngineBridge) SetModelConfigs(configs map[string]interfaces.ModelInfo) {
	e.modelConfigs = configs
}

// getAdapter gets or creates database adapter
func (e *InferenceEngineBridge) getAdapter(databaseID string) (adapter.DBAdapter, error) {
	if e.adapterFactory == nil {
		return nil, fmt.Errorf("adapter factory not set")
	}

	// Get adapter config from factory
	cfg, err := e.adapterFactory.GetConfig(databaseID)
	if err != nil {
		return nil, err
	}

	// Create internal adapter
	return adapter.NewAdapter(&adapter.DBConfig{
		Type:     cfg.Type,
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
		FilePath: cfg.FilePath,
	})
}

// convertResult converts internal result to interface result
func (e *InferenceEngineBridge) convertResult(r *inference.Result) *interfaces.InferenceResult {
	steps := make([]interfaces.ReActStep, len(r.ReActSteps))
	for i, s := range r.ReActSteps {
		steps[i] = interfaces.ReActStep{
			Step:        s.Step,
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionInput,
			Observation: s.Observation,
			Phase:       s.Phase,
		}
	}

	result := &interfaces.InferenceResult{
		SQL: r.GeneratedSQL,
		Metadata: interfaces.InferenceMetadata{
			SelectedTables: r.SelectedTables,
			Iterations:     r.LLMCalls,
			TotalTokens:    r.TotalTokens,
			ExecutionTime:  r.TotalTime,
			ReactTrace:     steps,
			Model:          e.currentModel,
		},
	}

	// Convert execution result if available
	if r.ExecutionResult != nil {
		if qr, ok := r.ExecutionResult.(*adapter.QueryResult); ok {
			result.ExecutionResult = &interfaces.QueryResult{
				Columns:       qr.Columns,
				Rows:          qr.Rows,
				RowCount:      qr.RowCount,
				ExecutionTime: qr.ExecutionTime,
				Error:         qr.Error,
			}
		}
	}

	return result
}
