package grounding

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/embedding"
	"lucid/internal/lakebase"
)

// Service provides the main entry point for Semantic Grounding.
// Uses AdaptivePipeline for all grounding operations:
//   - Small scale (≤ threshold tables): full schema → LinkingAgent
//   - Large scale (> threshold tables): vector retrieval → LinkingAgent
type Service struct {
	pipeline     *AdaptivePipeline
	vectorRepo   *lakebase.MySQLVectorRepository
	embedder     embedding.EmbeddingProvider
	llmModel     llms.Model
	config       *GroundingConfig
	datasourceID int64
}

// ServiceConfig configures the grounding service
type ServiceConfig struct {
	DatasourceID int64
	VectorRepo   *lakebase.MySQLVectorRepository
	Embedder     embedding.EmbeddingProvider
	LLMModel     llms.Model
	Config       *GroundingConfig
}

// NewService creates a new grounding service
func NewService(cfg *ServiceConfig) *Service {
	if cfg.Config == nil {
		cfg.Config = DefaultGroundingConfig()
	}

	svc := &Service{
		vectorRepo:   cfg.VectorRepo,
		embedder:     cfg.Embedder,
		llmModel:     cfg.LLMModel,
		config:       cfg.Config,
		datasourceID: cfg.DatasourceID,
	}

	if cfg.LLMModel != nil {
		svc.pipeline = NewAdaptivePipeline(cfg.VectorRepo, cfg.Embedder, cfg.LLMModel, cfg.Config)
	}

	return svc
}

// Ground performs adaptive grounding for a query.
// Requires AllSchemas to be provided via AdaptiveGroundingRequest.
// For backward compatibility, this method builds a minimal request.
func (s *Service) Ground(ctx context.Context, req *AdaptiveGroundingRequest) (*GroundingResult, error) {
	if s.pipeline == nil {
		return nil, fmt.Errorf("grounding pipeline not available (LLM required)")
	}

	if req.DatasourceID == 0 {
		req.DatasourceID = s.datasourceID
	}

	result, err := s.pipeline.Ground(ctx, req)
	if err != nil {
		return nil, err
	}

	return s.convertResult(result), nil
}

// GroundAdaptive is an alias for Ground for explicit adaptive grounding calls.
func (s *Service) GroundAdaptive(ctx context.Context, req *AdaptiveGroundingRequest) (*AdaptiveGroundingResult, error) {
	if s.pipeline == nil {
		return nil, fmt.Errorf("grounding pipeline not available (LLM required)")
	}

	if req.DatasourceID == 0 {
		req.DatasourceID = s.datasourceID
	}

	return s.pipeline.Ground(ctx, req)
}

// IsAvailable returns true if the grounding pipeline is initialized.
func (s *Service) IsAvailable() bool {
	return s.pipeline != nil
}

// SetDatasourceID updates the datasource ID for grounding
func (s *Service) SetDatasourceID(id int64) {
	s.datasourceID = id
}

// GetConfig returns the current grounding configuration
func (s *Service) GetConfig() *GroundingConfig {
	return s.config
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
// for injection into SQL generation.
func (s *Service) FormatContextPrompt(ctx *GroundedContext) string {
	var sb strings.Builder

	sb.WriteString("=== Semantic Grounding Results ===\n\n")

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

	if len(ctx.Columns) > 0 {
		sb.WriteString("## Relevant Columns:\n")
		for _, c := range ctx.Columns {
			sb.WriteString(fmt.Sprintf("- **%s.%s** (%s, relevance: %.2f)\n",
				c.TableName, c.ColumnName, c.DataType, c.Relevance))
			if c.Description != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", c.Description))
			}
			if c.SampleValues != "" {
				sb.WriteString(fmt.Sprintf("  Sample values: %s\n", c.SampleValues))
			}
			if c.Synonyms != "" {
				sb.WriteString(fmt.Sprintf("  Synonyms: %s\n", c.Synonyms))
			}
			if c.Reason != "" {
				sb.WriteString(fmt.Sprintf("  Reason: %s\n", c.Reason))
			}
		}
		sb.WriteString("\n")
	}

	if len(ctx.Relationships) > 0 {
		sb.WriteString("## Join Paths:\n")
		for _, r := range ctx.Relationships {
			sb.WriteString(fmt.Sprintf("- %s.%s → %s.%s (%s, confidence: %.2f)\n",
				r.FromTable, r.FromColumn, r.ToTable, r.ToColumn, r.Type, r.Confidence))
		}
		sb.WriteString("\n")
	}

	if len(ctx.BusinessRules) > 0 {
		sb.WriteString("## Business Rules:\n")
		for _, br := range ctx.BusinessRules {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s\n", br.TableName, br.RuleName, br.Description))
		}
		sb.WriteString("\n")
	}

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

	sb.WriteString(fmt.Sprintf("Grounding Stats: %d signals probed, %d selected, confidence: %.2f\n",
		ctx.SignalsProbed, ctx.SignalsSelected, ctx.Confidence))

	return sb.String()
}

// convertResult converts AdaptiveGroundingResult to legacy GroundingResult
func (s *Service) convertResult(ar *AdaptiveGroundingResult) *GroundingResult {
	if ar == nil || ar.Context == nil {
		return nil
	}
	return &GroundingResult{
		Context:           ar.Context,
		TotalDuration:     ar.TotalDuration,
		Mode:              string(ar.Strategy),
		ExecutionLogs:     ar.ExecutionLogs,
		CoarseSignals:     ar.CoarseSignals,
		CoarseDuration:    ar.RetrievalTime,
		SelectionDuration: ar.LinkingTime,
	}
}
