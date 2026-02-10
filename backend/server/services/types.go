package services

import (
	"context"
	"time"

	"lucid/internal/adapter"
)

// ============================================
// Inference Engine
// ============================================

// InferenceEngineInterface 推理引擎接口
type InferenceEngineInterface interface {
	Execute(ctx context.Context, req *InferenceRequest) (*InferenceResult, error)
	ExecuteStream(ctx context.Context, req *InferenceRequest, events chan<- StreamEvent) error
	GetAvailableModels() []ModelInfo
	SwitchModel(modelID string) error
	GetCurrentModel() string
	GetLLMModel() interface{}
}

// InferenceRequest 推理请求
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
}

// InferenceResult 推理结果
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

// ModelInfo 模型信息
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

// ReActStep ReAct 推理步骤
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

// EventType 事件类型
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

// StreamEvent 流式事件
type StreamEvent struct {
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorEventData 错误事件数据
type ErrorEventData struct {
	Error string `json:"error"`
}

// ============================================
// Field Suggester
// ============================================

// FieldSuggesterInterface 字段建议器接口
type FieldSuggesterInterface interface {
	SuggestFields(ctx context.Context, req *SuggestFieldsRequest) (*SuggestFieldsResult, error)
}

// SuggestFieldsRequest 字段建议请求
type SuggestFieldsRequest struct {
	Question   string `json:"question"`
	DatabaseID string `json:"database_id"`
	Database   string `json:"database"`
	Language   string `json:"language"`
}

// SuggestFieldsResult 字段建议结果
type SuggestFieldsResult struct {
	SuggestedFields []SuggestedField `json:"suggested_fields"`
	AnalysisNote    string           `json:"analysis_note"`
}

// SuggestedField 建议的字段
type SuggestedField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
	Source      string `json:"source"`
}
