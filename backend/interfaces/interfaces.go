// Package interfaces 定义系统核心接口
// 这些接口被 server 层和 internal 层共同使用，实现解耦
package interfaces

import (
	"context"
	"time"
)

// ============================================
// Database Adapter Interfaces
// ============================================

// DBAdapter 数据库适配器接口
type DBAdapter interface {
	// Connect 连接数据库
	Connect(ctx context.Context) error

	// Close 关闭连接
	Close() error

	// ExecuteQuery 执行查询
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)

	// GetDatabaseType 获取数据库类型
	GetDatabaseType() string

	// GetDatabaseVersion 获取数据库版本
	GetDatabaseVersion(ctx context.Context) (string, error)

	// DryRunSQL 验证 SQL 语法（不执行）
	DryRunSQL(ctx context.Context, sql string) error
}

// DBConfig 数据库连接配置
type DBConfig struct {
	Type     string // 数据库类型: "mysql", "postgresql", "sqlite"
	Host     string
	Port     int
	Database string
	User     string
	Password string
	FilePath string // SQLite 文件路径
}

// QueryResult 查询结果
type QueryResult struct {
	Columns       []string                 `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"row_count"`
	ExecutionTime int64                    `json:"execution_time"`
	Error         string                   `json:"error,omitempty"`
}

// AdapterFactory 适配器工厂函数类型
type AdapterFactory func(config *DBConfig) (DBAdapter, error)

// ============================================
// Inference Engine Interfaces
// ============================================

// InferenceEngine 推理引擎接口
type InferenceEngine interface {
	// Execute 执行推理
	Execute(ctx context.Context, req *InferenceRequest) (*InferenceResult, error)

	// ExecuteStream 流式执行推理
	ExecuteStream(ctx context.Context, req *InferenceRequest, events chan<- StreamEvent) error

	// GetAvailableModels 获取可用模型列表
	GetAvailableModels() []ModelInfo

	// SwitchModel 切换模型
	SwitchModel(modelID string) error

	// GetCurrentModel 获取当前模型
	GetCurrentModel() string

	// GetLLMModel 获取LLM模型实例（用于其他组件）
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
	GroundingResult  *GroundingResult `json:"grounding_result,omitempty"`
}

// InferenceResult 推理结果
type InferenceResult struct {
	SQL             string           `json:"sql"`
	ExecutionResult *QueryResult     `json:"execution_result,omitempty"`
	Metadata        InferenceMetadata `json:"metadata"`
}

// InferenceMetadata 推理元数据
type InferenceMetadata struct {
	SelectedTables     []string    `json:"selected_tables"`
	Iterations         int         `json:"iterations"`
	ReactTrace         []ReActStep `json:"react_trace"`
	RichContextUpdated bool        `json:"rich_context_updated"`
	LLMCalls           int         `json:"llm_calls"`
	Model              string      `json:"model"`
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
// ReAct Interfaces
// ============================================

// ReActStep ReAct 推理步骤
type ReActStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Timestamp   int64       `json:"timestamp"`
	Phase       string      `json:"phase"` // "schema_linking" or "sql_generation"
}

// ============================================
// Streaming Interfaces
// ============================================

// EventType 事件类型
type EventType string

const (
	EventStart           EventType = "start"
	EventThought         EventType = "thought"
	EventAction          EventType = "action"
	EventObservation     EventType = "observation"
	EventSchemaLinking   EventType = "schema_linking"
	EventSQLGeneration   EventType = "sql_generation"
	EventSQLExecuted     EventType = "sql_executed"
	EventComplete        EventType = "complete"
	EventError           EventType = "error"
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
// Grounding Interfaces
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
// Rich Context Interfaces
// ============================================

// RichContextProvider Rich Context 提供者接口
type RichContextProvider interface {
	// GetRichContext 获取 Rich Context
	GetRichContext(dbID, database string) (*RichContextInfo, error)

	// HasRichContext 检查是否存在 Rich Context
	HasRichContext(database string) bool
}

// RichContextInfo Rich Context 信息
type RichContextInfo struct {
	Database    string                 `json:"database"`
	Tables      []TableContextInfo     `json:"tables"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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
// Field Suggester Interfaces
// ============================================

// FieldSuggester 字段建议器接口
type FieldSuggester interface {
	// SuggestFields 根据问题建议输出字段
	SuggestFields(ctx context.Context, req *SuggestFieldsRequest) (*SuggestFieldsResult, error)
}

// SuggestFieldsRequest 字段建议请求
type SuggestFieldsRequest struct {
	Question   string `json:"question"`
	DatabaseID string `json:"database_id"`
	Database   string `json:"database"`
	Language   string `json:"language"` // "Chinese" or "English"
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
