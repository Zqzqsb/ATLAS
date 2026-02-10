// Package grounding implements Semantic Grounding Pipeline
// for AI-native database context retrieval.
//
// Semantic Grounding goes beyond traditional "Schema Linking" by:
// - Multi-signal vector retrieval (tables, columns, context, SQL templates)
// - Speculative parallel queries (50-100 lightweight probes)
// - Agent-based fine-grained selection with cross-attention reasoning
package grounding

import (
	"time"
)

// SignalType represents different types of retrieval signals
type SignalType string

const (
	// SignalTypeTable represents table-level semantic embeddings
	SignalTypeTable SignalType = "table"
	// SignalTypeColumn represents column-level semantic embeddings
	SignalTypeColumn SignalType = "column"
	// SignalTypeContext represents business context embeddings
	SignalTypeContext SignalType = "context"
	// SignalTypeSQLTemplate represents historical SQL pattern embeddings
	SignalTypeSQLTemplate SignalType = "sql_template"
	// SignalTypeDomainKnowledge represents domain-specific knowledge embeddings
	SignalTypeDomainKnowledge SignalType = "domain_knowledge"
	// SignalTypeRelationship represents table relationship embeddings
	SignalTypeRelationship SignalType = "relationship"
)

// RetrievalSignal represents a single retrieval result from vector search
type RetrievalSignal struct {
	// Basic identification
	ID           int64      `json:"id"`
	SignalType   SignalType `json:"signal_type"`
	DatasourceID int64      `json:"datasource_id"`

	// Content
	EntityName string `json:"entity_name"` // table name, column name, etc.
	Content    string `json:"content"`     // the actual text content
	Metadata   string `json:"metadata"`    // JSON metadata

	// Vector search results
	Embedding []float32 `json:"embedding,omitempty"`
	Distance  float32   `json:"distance"` // cosine distance
	Score     float32   `json:"score"`    // relevance score (1 - distance)

	// Lineage
	SourceTable  string `json:"source_table,omitempty"`
	SourceColumn string `json:"source_column,omitempty"`
}

// GroundedContext represents the final grounded context after agent selection
type GroundedContext struct {
	// Identified entities
	Tables       []TableContext       `json:"tables"`
	Columns      []ColumnContext      `json:"columns"`
	Relationships []RelationshipContext `json:"relationships"`

	// Business knowledge
	BusinessRules   []BusinessRule   `json:"business_rules,omitempty"`
	DomainTerms     []DomainTerm     `json:"domain_terms,omitempty"`
	SQLTemplates    []SQLTemplate    `json:"sql_templates,omitempty"`

	// Metadata
	Query          string    `json:"query"`
	GroundingTime  time.Duration `json:"grounding_time"`
	SignalsProbed  int       `json:"signals_probed"`
	SignalsSelected int      `json:"signals_selected"`
	Confidence     float32   `json:"confidence"`
	Reasoning      string    `json:"reasoning,omitempty"` // LLM reasoning for fine selection
}

// TableContext represents grounded table information
type TableContext struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Columns     []string `json:"columns"`
	Relevance   float32  `json:"relevance"`
	Reason      string   `json:"reason"` // why this table is selected
}

// ColumnContext represents grounded column information
type ColumnContext struct {
	TableName   string  `json:"table_name"`
	ColumnName  string  `json:"column_name"`
	DataType    string  `json:"data_type"`
	Description string  `json:"description"`
	Relevance   float32 `json:"relevance"`
	Reason      string  `json:"reason"`
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
	MapsTo     string   `json:"maps_to"` // table.column mapping
}

// SQLTemplate represents a historical SQL pattern
type SQLTemplate struct {
	Pattern     string  `json:"pattern"`
	Description string  `json:"description"`
	Similarity  float32 `json:"similarity"`
	Example     string  `json:"example,omitempty"`
}

// CoarseRetrievalConfig configures the coarse-grained retrieval stage
type CoarseRetrievalConfig struct {
	// Number of parallel probes per signal type
	ProbesPerType int `json:"probes_per_type" yaml:"probes_per_type"`
	// Total maximum signals to retrieve
	MaxSignals int `json:"max_signals" yaml:"max_signals"`
	// Minimum score threshold (0-1)
	MinScore float32 `json:"min_score" yaml:"min_score"`
	// Enable speculative parallel queries
	Speculative bool `json:"speculative" yaml:"speculative"`
	// Timeout for retrieval
	TimeoutMs int `json:"timeout_ms" yaml:"timeout_ms"`
}

// FineSelectionConfig configures the agent fine-grained selection stage
type FineSelectionConfig struct {
	// Maximum signals to send to LLM for selection
	MaxCandidates int `json:"max_candidates" yaml:"max_candidates"`
	// Enable multi-hop relationship tracing
	MultiHop bool `json:"multi_hop" yaml:"multi_hop"`
	// Maximum hops for relationship tracing
	MaxHops int `json:"max_hops" yaml:"max_hops"`
	// Confidence threshold for selection
	ConfidenceThreshold float32 `json:"confidence_threshold" yaml:"confidence_threshold"`
}

// GroundingConfig holds the complete grounding pipeline configuration
type GroundingConfig struct {
	// Coarse retrieval settings
	CoarseRetrieval CoarseRetrievalConfig `json:"coarse_retrieval" yaml:"coarse_retrieval"`
	// Fine selection settings
	FineSelection FineSelectionConfig `json:"fine_selection" yaml:"fine_selection"`
	// Enable parallel execution of coarse and fine stages
	ParallelExecution bool `json:"parallel_execution" yaml:"parallel_execution"`
	// Adaptive grounding settings
	ScaleThreshold int `json:"scale_threshold" yaml:"scale_threshold"` // Table count threshold for small vs large scale
	// Linking agent settings (used in both small and large scale modes)
	LinkingAgent LinkingAgentConfig `json:"linking_agent" yaml:"linking_agent"`
}

// GroundingStrategy defines how the grounding pipeline operates
type GroundingStrategy string

const (
	// StrategyAuto automatically selects strategy based on schema scale
	StrategyAuto GroundingStrategy = "auto"
	// StrategySmallScale directly passes all schema to linking agent
	StrategySmallScale GroundingStrategy = "small_scale"
	// StrategyLargeScale uses vector retrieval to narrow candidates first
	StrategyLargeScale GroundingStrategy = "large_scale"
)

// LinkingAgentConfig configures the LLM-based linking agent
type LinkingAgentConfig struct {
	// Maximum tables to include in LLM context (for large scale fallback)
	MaxTablesInContext int `json:"max_tables_in_context" yaml:"max_tables_in_context"`
	// Whether to include full column details in linking prompt
	IncludeColumnDetails bool `json:"include_column_details" yaml:"include_column_details"`
	// Whether to include rich context (descriptions, samples, synonyms) in linking prompt
	IncludeRichContext bool `json:"include_rich_context" yaml:"include_rich_context"`
	// Confidence threshold for table selection
	ConfidenceThreshold float32 `json:"confidence_threshold" yaml:"confidence_threshold"`
	// Strategy override (empty = auto)
	Strategy GroundingStrategy `json:"strategy" yaml:"strategy"`
}

// DefaultGroundingConfig returns sensible defaults for the grounding pipeline
func DefaultGroundingConfig() *GroundingConfig {
	return &GroundingConfig{
		CoarseRetrieval: CoarseRetrievalConfig{
			ProbesPerType: 20,
			MaxSignals:    100,
			MinScore:      0.3,
			Speculative:   true,
			TimeoutMs:     5000,
		},
		FineSelection: FineSelectionConfig{
			MaxCandidates:       50,
			MultiHop:            true,
			MaxHops:             3,
			ConfidenceThreshold: 0.5,
		},
		ParallelExecution: true,
		ScaleThreshold:    30, // <= 30 tables: small scale, > 30: large scale
		LinkingAgent: LinkingAgentConfig{
			MaxTablesInContext:  50,
			IncludeColumnDetails: true,
			IncludeRichContext:   true,
			ConfidenceThreshold: 0.5,
			Strategy:            StrategyAuto,
		},
	}
}
