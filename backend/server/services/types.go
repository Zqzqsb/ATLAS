package services

import (
	"context"
	"time"

	"lucid/interfaces"
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
	Question         string           `json:"question"`
	DatabaseID       string           `json:"database_id"`
	Database         string           `json:"database"`
	UseRichContext   bool             `json:"use_rich_context"`
	UseReact         bool             `json:"use_react"`
	MaxIterations    int              `json:"max_iterations"`
	FieldDescription string           `json:"field_description"`
	ContextFile      string           `json:"context_file,omitempty"`
	GroundingResult  *GroundingResult `json:"grounding_result,omitempty"`
}

// InferenceResult 推理结果
type InferenceResult struct {
	SQL             string               `json:"sql"`
	ExecutionResult *interfaces.QueryResult `json:"execution_result,omitempty"`
	Metadata        InferenceMetadata    `json:"metadata"`
}

// InferenceMetadata 推理元数据
type InferenceMetadata struct {
	SelectedTables     []string      `json:"selected_tables"`
	Iterations         int           `json:"iterations"`
	TotalTokens        int           `json:"total_tokens,omitempty"`
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
// Grounding
// ============================================

// GroundingResult 语义接地结果
type GroundingResult struct {
	Tables          []GroundedTable  `json:"tables"`
	Columns         []GroundedColumn `json:"columns"`
	JoinPaths       []JoinPath       `json:"join_paths,omitempty"`
	ExecutionTimeMs int64            `json:"execution_time_ms"`
}

// GroundedTable 接地的表
type GroundedTable struct {
	Name       string  `json:"name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// GroundedColumn 接地的列
type GroundedColumn struct {
	TableName  string  `json:"table_name"`
	ColumnName string  `json:"column_name"`
	Reason     string  `json:"reason,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// JoinPath 连接路径
type JoinPath struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	Reason     string `json:"reason,omitempty"`
}

// ============================================
// Rich Context
// ============================================

// RichContextProvider Rich Context 提供者接口
type RichContextProvider interface {
	GetRichContext(dbID, database string) (*RichContextInfo, error)
	HasRichContext(database string) bool
}

// RichContextInfo Rich Context 信息
type RichContextInfo struct {
	Database  string                 `json:"database"`
	Tables    []TableContextInfo     `json:"tables"`
	UpdatedAt time.Time              `json:"updated_at"`
	Version   string                 `json:"version"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TableContextInfo 表级 Context 信息
type TableContextInfo struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Columns     []ColumnContextInfo `json:"columns"`
	IsExpired   bool                `json:"is_expired"`
}

// ColumnContextInfo 列级 Context 信息
type ColumnContextInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Synonyms    []string `json:"synonyms,omitempty"`
	Examples    []string `json:"examples,omitempty"`
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
