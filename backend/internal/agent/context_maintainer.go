// Package agent provides self-maintenance capabilities for LUCID.
package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/lakebase"
)

// ContextMaintainer handles context expiration detection and automatic updates
type ContextMaintainer struct {
	repo       *lakebase.MySQLRepository
	llmModel   llms.Model
	logger     *ChangeLogger
}

// NewContextMaintainer creates a new context maintainer
func NewContextMaintainer(repo *lakebase.MySQLRepository, llmModel llms.Model, logger *ChangeLogger) *ContextMaintainer {
	return &ContextMaintainer{
		repo:     repo,
		llmModel: llmModel,
		logger:   logger,
	}
}

// SetLLMModel sets the LLM model for context regeneration
func (m *ContextMaintainer) SetLLMModel(model llms.Model) {
	m.llmModel = model
}

// ExpiredContextInfo holds information about expired context
type ExpiredContextInfo struct {
	Context      *lakebase.BusinessContext
	ExpiryReason string
	RelatedChange *SchemaChange
}

// ContextUpdateResult holds the result of a context update
type ContextUpdateResult struct {
	ContextID   int64                  `json:"context_id"`
	TableName   string                 `json:"table_name"`
	ColumnName  string                 `json:"column_name,omitempty"`
	ContextType lakebase.ContextType   `json:"context_type"`
	OldContent  json.RawMessage        `json:"old_content"`
	NewContent  json.RawMessage        `json:"new_content"`
	NewVersion  int                    `json:"new_version"`
	UpdatedAt   time.Time              `json:"updated_at"`
	UpdatedBy   string                 `json:"updated_by"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
}

// MarkContextExpiredByChanges marks context as expired based on schema changes
func (m *ContextMaintainer) MarkContextExpiredByChanges(ctx context.Context, dsID int64, changes []SchemaChange) (int, error) {
	if len(changes) == 0 {
		return 0, nil
	}

	detector := NewDDLDetector(m.repo)
	affectedIDs, err := detector.GetAffectedContextIDs(ctx, dsID, changes)
	if err != nil {
		return 0, err
	}

	if len(affectedIDs) == 0 {
		return 0, nil
	}

	// Build reason from changes
	reason := buildExpiryReason(changes)

	// Mark as expired
	if err := m.repo.MarkContextExpired(ctx, affectedIDs, reason); err != nil {
		return 0, err
	}

	// Log the expiration
	if m.logger != nil {
		for _, id := range affectedIDs {
			bc, _ := m.repo.GetContextByID(ctx, id)
			if bc != nil {
				m.logger.LogContextExpired(ctx, dsID, bc.TableName, bc.ColumnName.String, string(bc.ContextType), reason)
			}
		}
	}

	return len(affectedIDs), nil
}

// buildExpiryReason creates a reason string from schema changes
func buildExpiryReason(changes []SchemaChange) string {
	if len(changes) == 0 {
		return "Unknown reason"
	}

	reasons := make([]string, 0, len(changes))
	for _, change := range changes {
		switch change.ChangeType {
		case ChangeTypeTableDropped:
			reasons = append(reasons, fmt.Sprintf("Table '%s' was dropped", change.TableName))
		case ChangeTypeColumnDropped:
			reasons = append(reasons, fmt.Sprintf("Column '%s.%s' was dropped", change.TableName, change.ColumnName))
		case ChangeTypeColumnAdded:
			reasons = append(reasons, fmt.Sprintf("Column '%s.%s' was added", change.TableName, change.ColumnName))
		case ChangeTypeColumnModified:
			reasons = append(reasons, fmt.Sprintf("Column '%s.%s' was modified", change.TableName, change.ColumnName))
		case ChangeTypeForeignKeyAdded, ChangeTypeForeignKeyDropped:
			reasons = append(reasons, fmt.Sprintf("Foreign key changed on '%s.%s'", change.TableName, change.ColumnName))
		default:
			reasons = append(reasons, fmt.Sprintf("Schema change on '%s'", change.TableName))
		}
	}

	return strings.Join(reasons, "; ")
}

// GetExpiredContext retrieves all expired context for a datasource
func (m *ContextMaintainer) GetExpiredContext(ctx context.Context, dsID int64) ([]*lakebase.BusinessContext, error) {
	return m.repo.GetExpiredContextByDatasource(ctx, dsID)
}

// RefreshContext regenerates context for a specific entry using LLM
func (m *ContextMaintainer) RefreshContext(ctx context.Context, dsID int64, contextID int64, tableSchema []*lakebase.ColumnInfo) (*ContextUpdateResult, error) {
	result := &ContextUpdateResult{
		ContextID: contextID,
		UpdatedBy: "agent",
		UpdatedAt: time.Now(),
	}

	// Get the expired context
	bc, err := m.repo.GetContextByID(ctx, contextID)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.TableName = bc.TableName
	result.ColumnName = bc.ColumnName.String
	result.ContextType = bc.ContextType
	result.OldContent = bc.Content

	// Check if LLM is available
	if m.llmModel == nil {
		result.Error = "LLM model not available"
		return result, fmt.Errorf("LLM model not available")
	}

	// Generate new content based on context type
	newContent, err := m.regenerateContent(ctx, bc, tableSchema)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.NewContent = newContent

	// Update the context
	if err := m.repo.UpdateContextVersion(ctx, contextID, newContent, "agent", "Auto-refreshed after schema change"); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.NewVersion = bc.Version + 1
	result.Success = true

	// Log the update
	if m.logger != nil {
		m.logger.LogContextUpdated(ctx, dsID, bc.TableName, bc.ColumnName.String, string(bc.ContextType), bc.Content, newContent)
	}

	return result, nil
}

// regenerateContent generates new content for a context entry
func (m *ContextMaintainer) regenerateContent(ctx context.Context, bc *lakebase.BusinessContext, tableSchema []*lakebase.ColumnInfo) (json.RawMessage, error) {
	// Build schema description
	var schemaDesc strings.Builder
	for _, col := range tableSchema {
		schemaDesc.WriteString(fmt.Sprintf("- %s: %s", col.ColumnName, col.DataType.String))
		if col.IsPrimaryKey {
			schemaDesc.WriteString(" (PK)")
		}
		if col.IsForeignKey && col.ForeignKeyInfo != nil {
			schemaDesc.WriteString(fmt.Sprintf(" -> %s.%s", col.ForeignKeyInfo.RefTableName, col.ForeignKeyInfo.RefColumnName))
		}
		schemaDesc.WriteString("\n")
	}

	// Build prompt based on context type
	var prompt string
	switch bc.ContextType {
	case lakebase.ContextTypeSemantic:
		prompt = m.buildSemanticPrompt(bc, schemaDesc.String())
	case lakebase.ContextTypeEnumMeaning:
		prompt = m.buildEnumPrompt(bc, schemaDesc.String())
	case lakebase.ContextTypeBusinessRule:
		prompt = m.buildBusinessRulePrompt(bc, schemaDesc.String())
	case lakebase.ContextTypeDataQuality:
		prompt = m.buildDataQualityPrompt(bc, schemaDesc.String())
	case lakebase.ContextTypeJoinHint:
		prompt = m.buildJoinHintPrompt(bc, schemaDesc.String())
	default:
		prompt = m.buildGenericPrompt(bc, schemaDesc.String())
	}

	// Call LLM
	response, err := llms.GenerateFromSinglePrompt(ctx, m.llmModel, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse response based on context type
	return m.parseContextResponse(bc.ContextType, response)
}

func (m *ContextMaintainer) buildSemanticPrompt(bc *lakebase.BusinessContext, schema string) string {
	targetDesc := fmt.Sprintf("table '%s'", bc.TableName)
	if bc.ColumnName.Valid {
		targetDesc = fmt.Sprintf("column '%s.%s'", bc.TableName, bc.ColumnName.String)
	}

	return fmt.Sprintf(`Analyze the %s and provide a semantic description.

Current Schema:
%s

Previous description (may be outdated):
%s

Generate an updated semantic description in JSON format:
{"description": "...", "synonyms": ["..."], "business_terms": ["..."]}

Only output the JSON, no explanation.`, targetDesc, schema, string(bc.Content))
}

func (m *ContextMaintainer) buildEnumPrompt(bc *lakebase.BusinessContext, schema string) string {
	return fmt.Sprintf(`Analyze the enum values for column '%s.%s'.

Current Schema:
%s

Previous enum mapping (may be outdated):
%s

If the schema has changed, suggest updated enum value meanings in JSON format:
{"values": {"value1": "meaning1", "value2": "meaning2"}}

Only output the JSON, no explanation.`, bc.TableName, bc.ColumnName.String, schema, string(bc.Content))
}

func (m *ContextMaintainer) buildBusinessRulePrompt(bc *lakebase.BusinessContext, schema string) string {
	return fmt.Sprintf(`Analyze business rules for table '%s'.

Current Schema:
%s

Previous business rules (may be outdated):
%s

Generate updated business rules in JSON format:
{"rules": ["rule1", "rule2"], "constraints": ["constraint1"]}

Only output the JSON, no explanation.`, bc.TableName, schema, string(bc.Content))
}

func (m *ContextMaintainer) buildDataQualityPrompt(bc *lakebase.BusinessContext, schema string) string {
	targetDesc := fmt.Sprintf("table '%s'", bc.TableName)
	if bc.ColumnName.Valid {
		targetDesc = fmt.Sprintf("column '%s.%s'", bc.TableName, bc.ColumnName.String)
	}

	return fmt.Sprintf(`Review data quality considerations for %s.

Current Schema:
%s

Previous data quality notes (may be outdated):
%s

Generate updated data quality notes in JSON format:
{"anomalies": ["issue1"], "null_ratio": 0.0, "distinct_ratio": 0.0}

Only output the JSON, no explanation.`, targetDesc, schema, string(bc.Content))
}

func (m *ContextMaintainer) buildJoinHintPrompt(bc *lakebase.BusinessContext, schema string) string {
	return fmt.Sprintf(`Analyze JOIN relationships for table '%s'.

Current Schema:
%s

Previous JOIN hints (may be outdated):
%s

Generate updated JOIN hints in JSON format:
{"related_tables": ["table1"], "join_keys": ["key1"], "description": "..."}

Only output the JSON, no explanation.`, bc.TableName, schema, string(bc.Content))
}

func (m *ContextMaintainer) buildGenericPrompt(bc *lakebase.BusinessContext, schema string) string {
	return fmt.Sprintf(`Analyze and update the context for table '%s'.

Current Schema:
%s

Previous context (type: %s, may be outdated):
%s

Generate updated context in JSON format appropriate for the context type.
Only output the JSON, no explanation.`, bc.TableName, string(bc.ContextType), schema, string(bc.Content))
}

// parseContextResponse parses LLM response into JSON
func (m *ContextMaintainer) parseContextResponse(contextType lakebase.ContextType, response string) (json.RawMessage, error) {
	response = strings.TrimSpace(response)

	// Try to extract JSON from response
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")

	if startIdx >= 0 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	// Validate JSON
	var js json.RawMessage
	if err := json.Unmarshal([]byte(response), &js); err != nil {
		// If invalid JSON, wrap the response
		wrapped := map[string]string{"raw_response": response}
		return json.Marshal(wrapped)
	}

	return []byte(response), nil
}

// RefreshAllExpiredContext refreshes all expired context for a datasource
func (m *ContextMaintainer) RefreshAllExpiredContext(ctx context.Context, dsID int64) ([]*ContextUpdateResult, error) {
	var results []*ContextUpdateResult

	// Get expired context
	expired, err := m.GetExpiredContext(ctx, dsID)
	if err != nil {
		return nil, err
	}

	// Get schema for reference
	schemaMap := make(map[string][]*lakebase.ColumnInfo)

	for _, bc := range expired {
		// Get table schema if not cached
		if _, exists := schemaMap[bc.TableName]; !exists {
			tableSchema, err := m.repo.GetColumnsByTable(ctx, dsID, bc.TableName)
			if err != nil {
				continue
			}
			schemaMap[bc.TableName] = tableSchema
		}

		// Refresh this context
		result, _ := m.RefreshContext(ctx, dsID, bc.ID, schemaMap[bc.TableName])
		results = append(results, result)
	}

	return results, nil
}

// CheckAndMarkExpiredByTime marks context as expired if it exceeds the expiration time
func (m *ContextMaintainer) CheckAndMarkExpiredByTime(ctx context.Context, dsID int64) (int, error) {
	contexts, err := m.repo.GetContextByDatasource(ctx, dsID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var expiredIDs []int64

	for _, bc := range contexts {
		if bc.ExpiresAt.Valid && bc.ExpiresAt.Time.Before(now) {
			expiredIDs = append(expiredIDs, bc.ID)
		}
	}

	if len(expiredIDs) == 0 {
		return 0, nil
	}

	if err := m.repo.MarkContextExpired(ctx, expiredIDs, "Expired by time"); err != nil {
		return 0, err
	}

	return len(expiredIDs), nil
}

// CreateContextForNewColumn creates context for a newly added column
func (m *ContextMaintainer) CreateContextForNewColumn(ctx context.Context, dsID int64, tableName, columnName, dataType string) (*lakebase.BusinessContext, error) {
	if m.llmModel == nil {
		// Create basic context without LLM
		content, _ := lakebase.NewSemanticContent(
			fmt.Sprintf("Column %s of type %s", columnName, dataType),
			nil, nil,
		)
		bc := &lakebase.BusinessContext{
			DatasourceID: dsID,
			TableName:    tableName,
			ColumnName:   sql.NullString{String: columnName, Valid: true},
			ContextType:  lakebase.ContextTypeSemantic,
			Content:      content,
			Source:       lakebase.SourceLLM,
			Confidence:   0.5, // Lower confidence without LLM
			Version:      1,
			CreatedBy:    "agent",
			UpdatedBy:    "agent",
		}
		_, err := m.repo.SaveBusinessContext(ctx, bc)
		return bc, err
	}

	// Generate context using LLM
	prompt := fmt.Sprintf(`A new column '%s' of type '%s' was added to table '%s'.
Generate a semantic description in JSON format:
{"description": "...", "synonyms": [], "business_terms": []}
Only output the JSON.`, columnName, dataType, tableName)

	response, err := llms.GenerateFromSinglePrompt(ctx, m.llmModel, prompt)
	if err != nil {
		return nil, err
	}

	content, err := m.parseContextResponse(lakebase.ContextTypeSemantic, response)
	if err != nil {
		return nil, err
	}

	bc := &lakebase.BusinessContext{
		DatasourceID: dsID,
		TableName:    tableName,
		ColumnName:   sql.NullString{String: columnName, Valid: true},
		ContextType:  lakebase.ContextTypeSemantic,
		Content:      content,
		Source:       lakebase.SourceLLM,
		Confidence:   0.8,
		Version:      1,
		CreatedBy:    "agent",
		UpdatedBy:    "agent",
	}

	_, err = m.repo.SaveBusinessContext(ctx, bc)
	if err != nil {
		return nil, err
	}

	// Log the creation
	if m.logger != nil {
		m.logger.LogContextCreated(ctx, dsID, tableName, columnName, string(lakebase.ContextTypeSemantic))
	}

	return bc, nil
}
