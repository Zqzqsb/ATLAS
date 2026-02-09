package services

import (
	"context"
	"fmt"
	"time"

	"lucid/internal/adapter"
	"lucid/internal/config"
)

// InferenceService handles Text2SQL inference.
type InferenceService struct {
	config       *config.Config
	dbService    *DatabaseService
	engine       InferenceEngineInterface
	fieldSuggest FieldSuggesterInterface
}

// NewInferenceService creates a new inference service.
func NewInferenceService(
	cfg *config.Config,
	dbService *DatabaseService,
	engine InferenceEngineInterface,
	fieldSuggester FieldSuggesterInterface,
) *InferenceService {
	return &InferenceService{
		config:       cfg,
		dbService:    dbService,
		engine:       engine,
		fieldSuggest: fieldSuggester,
	}
}

// GetAvailableModels returns list of available models
func (s *InferenceService) GetAvailableModels() []ModelInfo {
	if s.engine == nil {
		return []ModelInfo{}
	}
	return s.engine.GetAvailableModels()
}

// SwitchModel switches the LLM model
func (s *InferenceService) SwitchModel(modelID string) error {
	if s.engine == nil {
		return fmt.Errorf("inference engine not initialized")
	}
	return s.engine.SwitchModel(modelID)
}

// GetCurrentModel returns the current model ID
func (s *InferenceService) GetCurrentModel() string {
	if s.engine == nil {
		return ""
	}
	return s.engine.GetCurrentModel()
}

// GetLLMModel returns the LLM model if available.
func (s *InferenceService) GetLLMModel() interface{} {
	if s.engine == nil {
		return nil
	}
	if getter, ok := s.engine.(interface{ GetLLMModelInterface() interface{} }); ok {
		return getter.GetLLMModelInterface()
	}
	return nil
}

// Text2SQLRequest represents a text2sql request.
type Text2SQLRequest struct {
	Question         string `json:"question"`
	DatabaseID       string `json:"database_id"`
	Database         string `json:"database"`
	UseRichContext   bool   `json:"use_rich_context"`
	UseReact         bool   `json:"use_react"`
	MaxIterations    int    `json:"max_iterations"`
	FieldDescription string `json:"field_description"`
}

// Text2SQLResult represents the result of text2sql
type Text2SQLResult struct {
	SQL             string                 `json:"sql"`
	ExecutionResult *adapter.QueryResult `json:"execution_result,omitempty"`
	Metadata        Text2SQLMetadata       `json:"metadata"`
}

// Text2SQLMetadata contains metadata about the inference
type Text2SQLMetadata struct {
	SelectedTables     []string    `json:"selected_tables"`
	Iterations         int         `json:"iterations"`
	ReactTrace         []ReActStep `json:"react_trace"`
	RichContextUpdated bool        `json:"rich_context_updated"`
	ExecutionTimeMs    int64       `json:"execution_time_ms"`
	LLMCalls           int         `json:"llm_calls"`
	Model              string      `json:"model"`
}

// Execute performs synchronous text2sql inference
func (s *InferenceService) Execute(ctx context.Context, req *Text2SQLRequest) (*Text2SQLResult, error) {
	if s.engine == nil {
		return nil, fmt.Errorf("inference engine not initialized")
	}

	startTime := time.Now()

	inferReq := &InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	result, err := s.engine.Execute(ctx, inferReq)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	response := &Text2SQLResult{
		SQL:             result.SQL,
		ExecutionResult: result.ExecutionResult,
		Metadata: Text2SQLMetadata{
			SelectedTables:     result.Metadata.SelectedTables,
			Iterations:         result.Metadata.Iterations,
			ReactTrace:         result.Metadata.ReactTrace,
			RichContextUpdated: result.Metadata.RichContextUpdated,
			ExecutionTimeMs:    time.Since(startTime).Milliseconds(),
			LLMCalls:           result.Metadata.LLMCalls,
			Model:              result.Metadata.Model,
		},
	}

	return response, nil
}

// ExecuteStream performs streaming text2sql inference
func (s *InferenceService) ExecuteStream(ctx context.Context, req *Text2SQLRequest, events chan<- StreamEvent) error {
	if s.engine == nil {
		return fmt.Errorf("inference engine not initialized")
	}

	inferReq := &InferenceRequest{
		Question:         req.Question,
		DatabaseID:       req.DatabaseID,
		Database:         req.Database,
		UseRichContext:   req.UseRichContext,
		UseReact:         req.UseReact,
		MaxIterations:    req.MaxIterations,
		FieldDescription: req.FieldDescription,
	}

	if inferReq.MaxIterations == 0 {
		inferReq.MaxIterations = s.config.React.MaxIterations
	}

	return s.engine.ExecuteStream(ctx, inferReq, events)
}

// SuggestFields suggests output fields based on the question.
func (s *InferenceService) SuggestFields(ctx context.Context, question, dbID, database, language string) (*SuggestFieldsResult, error) {
	if s.fieldSuggest == nil {
		return nil, fmt.Errorf("field suggester not available")
	}

	req := &SuggestFieldsRequest{
		Question:   question,
		DatabaseID: dbID,
		Database:   database,
		Language:   language,
	}

	return s.fieldSuggest.SuggestFields(ctx, req)
}

