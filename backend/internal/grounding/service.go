package grounding

import (
	"context"
	"fmt"
	"strings"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
	"lucid/internal/llm"
)

// GroundingMode defines the grounding execution mode
type GroundingMode string

const (
	// ModeCoarseOnly uses only vector retrieval without LLM selection
	ModeCoarseOnly GroundingMode = "coarse_only"
	// ModeSequential runs coarse retrieval then fine selection
	ModeSequential GroundingMode = "sequential"
	// ModeParallel runs coarse and fine selection with streaming results
	ModeParallel GroundingMode = "parallel"
)

// Service provides the main entry point for Semantic Grounding
type Service struct {
	pipeline     *Pipeline
	vectorRepo   *lakebase.MySQLVectorRepository
	embedder     embedding.EmbeddingProvider
	llmClient    llm.Client
	config       *GroundingConfig
	datasourceID int64
}

// ServiceConfig configures the grounding service
type ServiceConfig struct {
	DatasourceID int64
	VectorRepo   *lakebase.MySQLVectorRepository
	Embedder     embedding.EmbeddingProvider
	LLMClient    llm.Client
	Config       *GroundingConfig
}

// NewService creates a new grounding service
func NewService(cfg *ServiceConfig) *Service {
	if cfg.Config == nil {
		cfg.Config = DefaultGroundingConfig()
	}

	return &Service{
		pipeline:     NewPipeline(cfg.VectorRepo, cfg.Embedder, cfg.LLMClient, cfg.Config),
		vectorRepo:   cfg.VectorRepo,
		embedder:     cfg.Embedder,
		llmClient:    cfg.LLMClient,
		config:       cfg.Config,
		datasourceID: cfg.DatasourceID,
	}
}

// Ground performs semantic grounding for a query
// Returns selected tables, columns, and grounded context
func (s *Service) Ground(ctx context.Context, query string, mode GroundingMode) (*GroundingResult, error) {
	req := &GroundingRequest{
		Query:             query,
		DatasourceID:      s.datasourceID,
		SkipFineSelection: mode == ModeCoarseOnly,
	}

	switch mode {
	case ModeCoarseOnly, ModeSequential:
		return s.pipeline.Ground(ctx, req)
	case ModeParallel:
		// For parallel mode, collect all results and return the final one
		resultCh, err := s.pipeline.GroundParallel(ctx, req)
		if err != nil {
			return nil, err
		}
		var finalResult *GroundingResult
		for result := range resultCh {
			finalResult = result
		}
		return finalResult, nil
	default:
		return s.pipeline.Ground(ctx, req)
	}
}

// GroundWithStreaming performs grounding and streams intermediate results
func (s *Service) GroundWithStreaming(ctx context.Context, query string) (<-chan *GroundingResult, error) {
	return s.pipeline.GroundParallel(ctx, &GroundingRequest{
		Query:        query,
		DatasourceID: s.datasourceID,
	})
}

// GetSelectedTables extracts table names from grounded context
func (s *Service) GetSelectedTables(ctx *GroundedContext) []string {
	tables := make([]string, 0, len(ctx.Tables))
	for _, t := range ctx.Tables {
		tables = append(tables, t.Name)
	}
	return tables
}

// FormatContextPrompt formats the grounded context into a prompt string
// This can be used directly in SQL generation
func (s *Service) FormatContextPrompt(ctx *GroundedContext) string {
	var sb strings.Builder

	sb.WriteString("=== Semantic Grounding Results ===\n\n")

	// Tables
	if len(ctx.Tables) > 0 {
		sb.WriteString("## Relevant Tables:\n")
		for _, t := range ctx.Tables {
			sb.WriteString(fmt.Sprintf("- **%s** (relevance: %.2f)\n", t.Name, t.Relevance))
			if t.Description != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", t.Description))
			}
			if t.Reason != "" {
				sb.WriteString(fmt.Sprintf("  Reason: %s\n", t.Reason))
			}
			if len(t.Columns) > 0 {
				sb.WriteString(fmt.Sprintf("  Columns: %s\n", strings.Join(t.Columns, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	// Columns
	if len(ctx.Columns) > 0 {
		sb.WriteString("## Relevant Columns:\n")
		for _, c := range ctx.Columns {
			sb.WriteString(fmt.Sprintf("- **%s.%s** (%s, relevance: %.2f)\n",
				c.TableName, c.ColumnName, c.DataType, c.Relevance))
			if c.Description != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", c.Description))
			}
			if c.Reason != "" {
				sb.WriteString(fmt.Sprintf("  Reason: %s\n", c.Reason))
			}
		}
		sb.WriteString("\n")
	}

	// Relationships
	if len(ctx.Relationships) > 0 {
		sb.WriteString("## Join Paths:\n")
		for _, r := range ctx.Relationships {
			sb.WriteString(fmt.Sprintf("- %s.%s → %s.%s (%s, confidence: %.2f)\n",
				r.FromTable, r.FromColumn, r.ToTable, r.ToColumn, r.Type, r.Confidence))
		}
		sb.WriteString("\n")
	}

	// Business Rules
	if len(ctx.BusinessRules) > 0 {
		sb.WriteString("## Business Rules:\n")
		for _, br := range ctx.BusinessRules {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", br.TableName, br.RuleName, br.Description))
		}
		sb.WriteString("\n")
	}

	// Domain Terms
	if len(ctx.DomainTerms) > 0 {
		sb.WriteString("## Domain Terms:\n")
		for _, dt := range ctx.DomainTerms {
			sb.WriteString(fmt.Sprintf("- **%s**: %s", dt.Term, dt.Definition))
			if dt.MapsTo != "" {
				sb.WriteString(fmt.Sprintf(" (maps to: %s)", dt.MapsTo))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// SQL Templates
	if len(ctx.SQLTemplates) > 0 {
		sb.WriteString("## Similar SQL Patterns:\n")
		for _, tpl := range ctx.SQLTemplates {
			sb.WriteString(fmt.Sprintf("- %s (similarity: %.2f)\n", tpl.Description, tpl.Similarity))
			if tpl.Pattern != "" {
				sb.WriteString(fmt.Sprintf("  Pattern: %s\n", tpl.Pattern))
			}
		}
		sb.WriteString("\n")
	}

	// Metadata
	sb.WriteString(fmt.Sprintf("Grounding Stats: %d signals probed, %d selected, confidence: %.2f\n",
		ctx.SignalsProbed, ctx.SignalsSelected, ctx.Confidence))

	return sb.String()
}

// SetDatasourceID updates the datasource ID for grounding
func (s *Service) SetDatasourceID(id int64) {
	s.datasourceID = id
}

// GetConfig returns the current grounding configuration
func (s *Service) GetConfig() *GroundingConfig {
	return s.config
}

// UpdateConfig updates the grounding configuration
func (s *Service) UpdateConfig(cfg *GroundingConfig) {
	s.config = cfg
	s.pipeline.UpdateConfig(cfg)
}
