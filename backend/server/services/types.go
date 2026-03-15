package services

import (
	"context"
	"time"

	"lucid/internal/adapter"
)

// ============================================
// Inference Engine
// ============================================

// InferenceEngineInterface defines the inference engine contract.
type InferenceEngineInterface interface {
	Execute(ctx context.Context, req *InferenceRequest) (*InferenceResult, error)
	ExecuteStream(ctx context.Context, req *InferenceRequest, events chan<- StreamEvent) error
	GetAvailableModels() []ModelInfo
	SwitchModel(modelID string) error
	GetCurrentModel() string
	GetLLMModel() interface{}
}

// InferenceRequest holds the parameters for an inference call.
type InferenceRequest struct {
	Question         string `json:"question"`
	DatabaseID       string `json:"database_id"`
	Database         string `json:"database"`
	UseRichContext   bool   `json:"use_rich_context"`
	UseReact         bool   `json:"use_react"`
	MaxIterations    int    `json:"max_iterations"`
	FieldDescription string `json:"field_description"`

	// Pre-linked context from Grounding stage (if available)
	// When set, inference pipeline skips its own Schema Linking
	LinkedTables       []string `json:"linked_tables,omitempty"`
	LinkedContextPrompt string  `json:"linked_context_prompt,omitempty"`

	// GroundingExecuted indicates that the Grounding pipeline has already run.
	// When true, inference pipeline ALWAYS skips internal Schema Linking,
	// even if LinkedTables is empty (grounding selected 0 tables).
	// This prevents the "two schema linkers" problem where legacy inference
	// re-runs linking after grounding already did it.
	GroundingExecuted bool `json:"grounding_executed,omitempty"`
}

// InferenceResult holds the output of an inference call.
type InferenceResult struct {
	SQL             string               `json:"sql"`
	ExecutionResult *adapter.QueryResult `json:"execution_result,omitempty"`
	Metadata        InferenceMetadata    `json:"metadata"`
}

// InferenceMetadata holds metadata about an inference run.
type InferenceMetadata struct {
	SelectedTables     []string      `json:"selected_tables"`
	Iterations         int           `json:"iterations"`
	ExecutionTime      time.Duration `json:"execution_time,omitempty"`
	ReactTrace         []ReActStep   `json:"react_trace"`
	RichContextUpdated bool          `json:"rich_context_updated"`
	LLMCalls           int           `json:"llm_calls"`
	Model              string        `json:"model"`
}

// ModelInfo describes an available LLM model.
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description,omitempty"`
	IsDefault   bool   `json:"is_default"`
}

// ============================================
// ReAct
// ============================================

// ReActStep represents a single step in a ReAct reasoning trace.
type ReActStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Timestamp   int64       `json:"timestamp"`
	Phase       string      `json:"phase"`
}

// ============================================
// Streaming
// ============================================

// EventType identifies the kind of streaming event.
type EventType string

const (
	EventStart             EventType = "start"
	EventThought           EventType = "thought"
	EventAction            EventType = "action"
	EventObservation       EventType = "observation"
	EventSchemaLinking     EventType = "schema_linking"
	EventSQLGeneration     EventType = "sql_generation"
	EventSQLExecuted       EventType = "sql_executed"
	EventComplete          EventType = "complete"
	EventError             EventType = "error"
	EventRichContextUpdate EventType = "rich_context_update"
)

// StreamEvent is a single SSE event sent to the frontend.
type StreamEvent struct {
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorEventData carries error details inside a StreamEvent.
type ErrorEventData struct {
	Error string `json:"error"`
}

// ============================================
// Field Suggester
// ============================================

// FieldSuggesterInterface defines the field suggestion contract.
type FieldSuggesterInterface interface {
	SuggestFields(ctx context.Context, req *SuggestFieldsRequest) (*SuggestFieldsResult, error)
}

// SuggestFieldsRequest holds parameters for a field suggestion call.
type SuggestFieldsRequest struct {
	Question   string `json:"question"`
	DatabaseID string `json:"database_id"`
	Database   string `json:"database"`
	Language   string `json:"language"`
}

// SuggestFieldsResult holds the output of a field suggestion call.
type SuggestFieldsResult struct {
	SuggestedFields []SuggestedField `json:"suggested_fields"`
	AnalysisNote    string           `json:"analysis_note"`
}

// SuggestedField is a single field recommendation.
type SuggestedField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
	Source      string `json:"source"`
}
