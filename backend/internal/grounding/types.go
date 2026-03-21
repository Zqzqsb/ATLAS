// Package grounding implements the Semantic Grounding Pipeline
// for AI-native database context retrieval.
//
// Architecture (Adaptive Pipeline):
//   - Small scale (≤ threshold tables): full schema + RC → LinkingAgent
//   - Large scale (> threshold tables): CoarseRetriever → LinkingAgent
package grounding

import (
	"time"
)

// SignalType represents different types of retrieval signals
type SignalType string

const (
	SignalTypeTable          SignalType = "table"
	SignalTypeColumn         SignalType = "column"
	SignalTypeContext        SignalType = "context"
	SignalTypeSQLTemplate    SignalType = "sql_template"
	SignalTypeDomainKnowledge SignalType = "domain_knowledge"
	SignalTypeRelationship   SignalType = "relationship"
)

// RetrievalSignal represents a single retrieval result from vector search
type RetrievalSignal struct {
	ID           int64      `json:"id"`
	SignalType   SignalType `json:"signal_type"`
	DatasourceID int64      `json:"datasource_id"`

	EntityName string `json:"entity_name"`
	Content    string `json:"content"`
	Metadata   string `json:"metadata"`

	Embedding []float32 `json:"embedding,omitempty"`
	Distance  float32   `json:"distance"`
	Score     float32   `json:"score"` // 1 - distance

	SourceTable  string `json:"source_table,omitempty"`
	SourceColumn string `json:"source_column,omitempty"`
}

// GroundedContext represents the final grounded context after linking agent selection
type GroundedContext struct {
	Tables        []TableContext        `json:"tables"`
	Columns       []ColumnContext       `json:"columns"`
	Relationships []RelationshipContext `json:"relationships"`

	BusinessRules []BusinessRule `json:"business_rules,omitempty"`
	DomainTerms   []DomainTerm   `json:"domain_terms,omitempty"`
	SQLTemplates  []SQLTemplate  `json:"sql_templates,omitempty"`

	Query           string        `json:"query"`
	GroundingTime   time.Duration `json:"grounding_time"`
	SignalsProbed   int           `json:"signals_probed"`
	SignalsSelected int           `json:"signals_selected"`
	Confidence      float32       `json:"confidence"`
	Reasoning       string        `json:"reasoning,omitempty"`
}

// TableContext represents grounded table information
type TableContext struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Columns     []string `json:"columns"`
	Relevance   float32  `json:"relevance"`
	Reason      string   `json:"reason"`
	Hint        string   `json:"hint,omitempty"` // Query-specific usage hint from generative linking
}

// ColumnContext represents grounded column information
type ColumnContext struct {
	TableName    string  `json:"table_name"`
	ColumnName   string  `json:"column_name"`
	DataType     string  `json:"data_type"`
	Description  string  `json:"description"`
	SampleValues string  `json:"sample_values,omitempty"`
	Synonyms     string  `json:"synonyms,omitempty"`
	Relevance    float32 `json:"relevance"`
	Reason       string  `json:"reason"`
	Hint         string  `json:"hint,omitempty"` // Query-specific usage hint from generative linking
}

// RelationshipContext represents grounded table relationships
type RelationshipContext struct {
	FromTable  string  `json:"from_table"`
	FromColumn string  `json:"from_column"`
	ToTable    string  `json:"to_table"`
	ToColumn   string  `json:"to_column"`
	Type       string  `json:"type"` // "foreign_key", "implicit", "semantic"
	Confidence float32 `json:"confidence"`
}

// BusinessRule represents a grounded business rule
type BusinessRule struct {
	TableName   string  `json:"table_name"`
	RuleName    string  `json:"rule_name"`
	Description string  `json:"description"`
	Relevance   float32 `json:"relevance"`
}

// DomainTerm represents a grounded domain terminology
type DomainTerm struct {
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Synonyms   []string `json:"synonyms,omitempty"`
	MapsTo     string   `json:"maps_to"`
}

// SQLTemplate represents a historical SQL pattern
type SQLTemplate struct {
	Pattern     string  `json:"pattern"`
	Description string  `json:"description"`
	Similarity  float32 `json:"similarity"`
	Example     string  `json:"example,omitempty"`
}

// --- Configuration ---

// CoarseRetrievalConfig configures the vector retrieval stage (large scale only)
type CoarseRetrievalConfig struct {
	ProbesPerType int     `json:"probes_per_type" yaml:"probes_per_type"`
	MaxSignals    int     `json:"max_signals" yaml:"max_signals"`
	MinScore      float32 `json:"min_score" yaml:"min_score"`
	Speculative   bool    `json:"speculative" yaml:"speculative"`
	TimeoutMs     int     `json:"timeout_ms" yaml:"timeout_ms"`
}

// LinkingAgentConfig configures the LLM-based linking agent
type LinkingAgentConfig struct {
	MaxTablesInContext   int              `json:"max_tables_in_context" yaml:"max_tables_in_context"`
	IncludeColumnDetails bool             `json:"include_column_details" yaml:"include_column_details"`
	IncludeRichContext   bool             `json:"include_rich_context" yaml:"include_rich_context"`
	ConfidenceThreshold  float32          `json:"confidence_threshold" yaml:"confidence_threshold"`
	Strategy             GroundingStrategy `json:"strategy" yaml:"strategy"`
}

// GroundingStrategy defines how the grounding pipeline operates
type GroundingStrategy string

const (
	StrategyAuto       GroundingStrategy = "auto"
	StrategySmallScale GroundingStrategy = "small_scale"
	StrategyLargeScale GroundingStrategy = "large_scale"
)

// GroundingConfig holds the complete grounding configuration
type GroundingConfig struct {
	CoarseRetrieval CoarseRetrievalConfig `json:"coarse_retrieval" yaml:"coarse_retrieval"`
	ScaleThreshold  int                   `json:"scale_threshold" yaml:"scale_threshold"`
	LinkingAgent    LinkingAgentConfig    `json:"linking_agent" yaml:"linking_agent"`
}

// GroundingResult represents the final grounding result (unified)
type GroundingResult struct {
	Context           *GroundedContext
	CoarseSignals     []*RetrievalSignal
	CoarseDuration    time.Duration
	SelectionDuration time.Duration
	TotalDuration     time.Duration
	Mode              string
	ExecutionLogs     []ExecutionLog `json:"execution_logs,omitempty"`
}

// ExecutionLog represents a step in the grounding pipeline for transparency
type ExecutionLog struct {
	Phase       string        `json:"phase"`
	SQL         string        `json:"sql"`
	Params      []interface{} `json:"params"`
	ResultCount int           `json:"result_count"`
	Duration    time.Duration `json:"duration"`
	Summary     string        `json:"summary"`
}

// DefaultGroundingConfig returns sensible defaults
func DefaultGroundingConfig() *GroundingConfig {
	return &GroundingConfig{
		CoarseRetrieval: CoarseRetrievalConfig{
			ProbesPerType: 30,
			MaxSignals:    100,
			MinScore:      0.15,
			Speculative:   true,
			TimeoutMs:     5000,
		},
		ScaleThreshold: 30,
		LinkingAgent: LinkingAgentConfig{
			MaxTablesInContext:   50,
			IncludeColumnDetails: true,
			IncludeRichContext:   true,
			ConfidenceThreshold:  0.5,
			Strategy:             StrategyAuto,
		},
	}
}
