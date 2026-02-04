// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in LUCID system.
package lakebase

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Datasource represents a data source configuration
type Datasource struct {
	ID           int64          `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	DBType       string         `json:"db_type" db:"db_type"`
	Host         sql.NullString `json:"host" db:"host"`
	Port         sql.NullInt32  `json:"port" db:"port"`
	Username     sql.NullString `json:"username" db:"username"`
	DatabaseName sql.NullString `json:"database_name" db:"db_name"`
	Description  sql.NullString `json:"description" db:"description"`
	Status       string         `json:"status" db:"status"`
	LastSyncAt   sql.NullTime   `json:"last_sync_at" db:"last_sync_at"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// DatasourceStatus constants
const (
	DatasourceStatusActive   = "active"
	DatasourceStatusInactive = "inactive"
	DatasourceStatusError    = "error"
	// Legacy aliases
	DatasourceStatusEnabled  = DatasourceStatusActive
	DatasourceStatusDisabled = DatasourceStatusInactive
)

// TableInfo represents table-level Rich Context from rc_tables
type TableInfo struct {
	ID           int64          `json:"id" db:"id"`
	DatasourceID int64          `json:"datasource_id" db:"datasource_id"`
	TableName    string         `json:"table_name" db:"table_name"`
	Description  sql.NullString `json:"description" db:"description"`
	RowCount     int64          `json:"row_count" db:"row_count"`
	IsExpired    bool           `json:"is_expired" db:"is_expired"`
	Source       string         `json:"source" db:"source"`
	Confidence   float64        `json:"confidence" db:"confidence"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// ColumnInfo represents column-level Rich Context from rc_columns
type ColumnInfo struct {
	ID           int64          `json:"id" db:"id"`
	DatasourceID int64          `json:"datasource_id" db:"datasource_id"`
	TableName    string         `json:"table_name" db:"table_name"`
	ColumnName   string         `json:"column_name" db:"column_name"`
	DataType     sql.NullString `json:"data_type" db:"data_type"`
	Description  sql.NullString `json:"description" db:"description"`
	SampleValues sql.NullString `json:"sample_values" db:"sample_values"`
	Synonyms     sql.NullString `json:"synonyms" db:"synonyms"`
	IsNullable   bool           `json:"is_nullable" db:"is_nullable"`
	IsPrimaryKey bool           `json:"is_primary_key" db:"is_primary_key"`
	IsForeignKey bool           `json:"is_foreign_key" db:"is_foreign_key"`
	IsExpired    bool           `json:"is_expired" db:"is_expired"`
	Source       string         `json:"source" db:"source"`
	Confidence   float64        `json:"confidence" db:"confidence"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// SchemaMetadata represents schema structure metadata (legacy)
type SchemaMetadata struct {
	ID           int64     `json:"id" db:"id"`
	DatasourceID int64     `json:"datasource_id" db:"datasource_id"`
	TableName    string    `json:"table_name" db:"table_name"`
	ColumnName   string    `json:"column_name" db:"column_name"`
	DataType     string    `json:"data_type" db:"data_type"`
	IsPrimaryKey bool      `json:"is_primary_key" db:"is_primary_key"`
	IsForeignKey bool      `json:"is_foreign_key" db:"is_foreign_key"`
	FKRefTable   string    `json:"fk_ref_table" db:"fk_ref_table"`
	FKRefColumn  string    `json:"fk_ref_column" db:"fk_ref_column"`
	Nullable     bool      `json:"nullable" db:"nullable"`
	DefaultValue string    `json:"default_value" db:"default_value"`
	Comment      string    `json:"comment" db:"comment"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ContextType represents the type of business context
type ContextType string

const (
	ContextTypeEnumMeaning  ContextType = "enum_meaning"
	ContextTypeBusinessRule ContextType = "business_rule"
	ContextTypeJoinHint     ContextType = "join_hint"
	ContextTypeDataQuality  ContextType = "data_quality"
	ContextTypeSemantic     ContextType = "semantic"
)

// ContextSource represents the source of context
type ContextSource string

const (
	SourceLLM           ContextSource = "llm"
	SourceCatalog       ContextSource = "catalog"
	SourceUser          ContextSource = "user"
	SourceAutoCorrected ContextSource = "auto_corrected"
)

// BusinessContext represents semantic business context
type BusinessContext struct {
	ID           int64           `json:"id" db:"id"`
	DatasourceID int64           `json:"datasource_id" db:"datasource_id"`
	TableName    string          `json:"table_name" db:"table_name"`
	ColumnName   sql.NullString  `json:"column_name" db:"column_name"`
	ContextType  ContextType     `json:"context_type" db:"context_type"`
	Content      json.RawMessage `json:"content" db:"content"`
	Source       ContextSource   `json:"source" db:"source"`
	Confidence   float64         `json:"confidence" db:"confidence"`
	IsExpired    bool            `json:"is_expired" db:"is_expired"`
	ExpiresAt    sql.NullTime    `json:"expires_at" db:"expires_at"`
	Version      int             `json:"version" db:"version"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	CreatedBy    string          `json:"created_by" db:"created_by"`
	UpdatedBy    string          `json:"updated_by" db:"updated_by"`
	UpdateReason string          `json:"update_reason" db:"update_reason"`
}

// EntityType represents the type of entity for embeddings
type EntityType string

const (
	EntityTypeTable        EntityType = "table"
	EntityTypeColumn       EntityType = "column"
	EntityTypeContext      EntityType = "context"
	EntityTypeQuery        EntityType = "query"
	EntityTypeRelationship EntityType = "relationship"
)

// Embedding represents vector embedding
type Embedding struct {
	ID             int64      `json:"id" db:"id"`
	DatasourceID   int64      `json:"datasource_id" db:"datasource_id"`
	EntityType     EntityType `json:"entity_type" db:"entity_type"`
	EntityID       int64      `json:"entity_id" db:"entity_id"`
	EntityText     string     `json:"entity_text" db:"entity_text"`
	Embedding      []float32  `json:"-" db:"embedding"` // Vector data, hidden in JSON
	EmbeddingModel string     `json:"embedding_model" db:"embedding_model"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// EmbeddingWithDistance represents an embedding result with distance score
type EmbeddingWithDistance struct {
	Embedding
	Distance float64 `json:"distance" db:"distance"`
}

// DefaultEmbeddingModel is the default embedding model
const DefaultEmbeddingModel = "text-embedding-3-small"

// DefaultEmbeddingDimension is the dimension for OpenAI text-embedding-3-small
const DefaultEmbeddingDimension = 1536

// ChangeType represents the type of change in change log
type ChangeType string

const (
	ChangeTypeSchemaChange  ChangeType = "schema_change"
	ChangeTypeContextUpdate ChangeType = "context_update"
	ChangeTypeContextExpire ChangeType = "context_expire"
)

// TriggerSource represents the source that triggered a change
type TriggerSource string

const (
	TriggerSourceAgent  TriggerSource = "agent"
	TriggerSourceUser   TriggerSource = "user"
	TriggerSourceSystem TriggerSource = "system"
)

// ChangeLog represents change audit log
type ChangeLog struct {
	ID            int64           `json:"id" db:"id"`
	DatasourceID  int64           `json:"datasource_id" db:"datasource_id"`
	TableName     string          `json:"table_name" db:"table_name"`
	ChangeType    ChangeType      `json:"change_type" db:"change_type"`
	ChangeDetail  json.RawMessage `json:"change_detail" db:"change_detail"`
	OldValue      json.RawMessage `json:"old_value" db:"old_value"`
	NewValue      json.RawMessage `json:"new_value" db:"new_value"`
	TriggerSource TriggerSource   `json:"trigger_source" db:"trigger_source"`
	ChangeReason  string          `json:"change_reason" db:"change_reason"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// StatType represents the type of statistics
type StatType string

const (
	StatTypeRowCount      StatType = "row_count"
	StatTypeDistinctCount StatType = "distinct_count"
	StatTypeNullRatio     StatType = "null_ratio"
	StatTypeHotValues     StatType = "hot_values"
)

// Statistics represents data statistics
type Statistics struct {
	ID           int64           `json:"id" db:"id"`
	DatasourceID int64           `json:"datasource_id" db:"datasource_id"`
	TableName    string          `json:"table_name" db:"table_name"`
	ColumnName   sql.NullString  `json:"column_name" db:"column_name"`
	StatType     StatType        `json:"stat_type" db:"stat_type"`
	StatValue    json.RawMessage `json:"stat_value" db:"stat_value"`
	CollectedAt  time.Time       `json:"collected_at" db:"collected_at"`
}

// JoinPath represents pre-computed JOIN paths
type JoinPath struct {
	ID             int64           `json:"id" db:"id"`
	DatasourceID   int64           `json:"datasource_id" db:"datasource_id"`
	FromTable      string          `json:"from_table" db:"from_table"`
	ToTable        string          `json:"to_table" db:"to_table"`
	PathTables     json.RawMessage `json:"path_tables" db:"path_tables"`
	JoinConditions json.RawMessage `json:"join_conditions" db:"join_conditions"`
	PathLength     int             `json:"path_length" db:"path_length"`
	Confidence     float64         `json:"confidence" db:"confidence"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// JoinCondition represents a single JOIN condition
type JoinCondition struct {
	LeftTable   string `json:"left_table"`
	LeftColumn  string `json:"left_column"`
	RightTable  string `json:"right_table"`
	RightColumn string `json:"right_column"`
	JoinType    string `json:"join_type"` // INNER, LEFT, RIGHT
}

// EnumMeaningContent represents content for enum_meaning context type
type EnumMeaningContent struct {
	Values map[string]string `json:"values"` // enum value -> meaning
}

// BusinessRuleContent represents content for business_rule context type
type BusinessRuleContent struct {
	Rules       []string `json:"rules"`
	Constraints []string `json:"constraints,omitempty"`
}

// JoinHintContent represents content for join_hint context type
type JoinHintContent struct {
	RelatedTables []string `json:"related_tables"`
	JoinKeys      []string `json:"join_keys"`
	Description   string   `json:"description,omitempty"`
}

// DataQualityContent represents content for data_quality context type
type DataQualityContent struct {
	NullRatio     float64  `json:"null_ratio,omitempty"`
	DistinctRatio float64  `json:"distinct_ratio,omitempty"`
	Anomalies     []string `json:"anomalies,omitempty"`
}

// SemanticContent represents content for semantic context type
type SemanticContent struct {
	Description   string   `json:"description"`
	Synonyms      []string `json:"synonyms,omitempty"`
	BusinessTerms []string `json:"business_terms,omitempty"`
}
