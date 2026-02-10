package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/adapter"
	"lucid/internal/config"
)

// FieldSuggester suggests output fields based on LLM analysis.
type FieldSuggester struct {
	llmModel       llms.Model
	adapterFactory adapter.AdapterFactory
	config         *config.Config
}

func NewFieldSuggester(llmModel llms.Model, factory adapter.AdapterFactory, cfg *config.Config) *FieldSuggester {
	return &FieldSuggester{
		llmModel:       llmModel,
		adapterFactory: factory,
		config:         cfg,
	}
}

func (s *FieldSuggester) SuggestFields(c context.Context, req *SuggestFieldsRequest) (*SuggestFieldsResult, error) {
	if s.llmModel == nil {
		return &SuggestFieldsResult{
			SuggestedFields: []SuggestedField{},
			AnalysisNote:    "LLM not available for field suggestion",
		}, nil
	}

	schema, err := s.getSchemaForSuggestion(c, req.DatabaseID, req.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	prompt := s.buildFieldSuggestionPrompt(req.Question, schema, req.Language)
	response, err := s.llmModel.Call(c, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	return s.parseFieldSuggestionResponse(response), nil
}

func (s *FieldSuggester) getSchemaForSuggestion(c context.Context, dbID, database string) (string, error) {
	if s.adapterFactory == nil {
		return "", fmt.Errorf("adapter factory not available")
	}

	var adapterCfg *adapter.DBConfig
	if s.config != nil {
		for _, db := range s.config.Databases {
			if db.ID == dbID {
				adapterCfg = &adapter.DBConfig{
					Type: db.Type, Host: db.Host, Port: db.Port,
					Database: database, User: db.User, Password: db.Password,
				}
				break
			}
		}
	}
	if adapterCfg == nil {
		adapterCfg = &adapter.DBConfig{Type: "mariadb", Database: database}
	}

	adp, err := s.adapterFactory(adapterCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create adapter: %w", err)
	}
	defer adp.Close()

	if err := adp.Connect(c); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Database Schema:\n\n")

	tablesResult, err := adp.ExecuteQuery(c, `
		SELECT TABLE_NAME FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE'
	`)
	if err != nil {
		return "", fmt.Errorf("failed to get tables: %w", err)
	}

	for _, row := range tablesResult.Rows {
		tableName, ok := row["TABLE_NAME"].(string)
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("Table: %s\n  Columns:\n", tableName))

		columnsResult, err := adp.ExecuteQuery(c, fmt.Sprintf(`
			SELECT COLUMN_NAME, DATA_TYPE, COLUMN_COMMENT
			FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION
		`, tableName))
		if err != nil {
			continue
		}

		for _, colRow := range columnsResult.Rows {
			colName, _ := colRow["COLUMN_NAME"].(string)
			dataType, _ := colRow["DATA_TYPE"].(string)
			comment, _ := colRow["COLUMN_COMMENT"].(string)
			if comment != "" {
				sb.WriteString(fmt.Sprintf("    - %s (%s): %s\n", colName, dataType, comment))
			} else {
				sb.WriteString(fmt.Sprintf("    - %s (%s)\n", colName, dataType))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (s *FieldSuggester) buildFieldSuggestionPrompt(question, schema, language string) string {
	langNote := "English"
	if language == "Chinese" || language == "zh" {
		langNote = "Chinese"
	}
	return fmt.Sprintf(`You are a database expert. Analyze the user's question and the database schema to suggest which fields (columns) should appear in the SELECT statement output.

User Question: %s

%s

Task: Based on the question, identify the most relevant output fields.

Instructions:
1. Consider what information the user is asking for
2. Include fields that directly answer the question
3. Include helpful context fields that make results meaningful
4. Don't include internal IDs unless specifically requested
5. Provide the response in %s

Output Format (JSON):
{
  "suggested_fields": [
    {"table": "table_name", "column": "column_name", "reason": "why this field is relevant"}
  ],
  "analysis_note": "Brief explanation"
}

Only output valid JSON, no other text.`, question, schema, langNote)
}

func (s *FieldSuggester) parseFieldSuggestionResponse(response string) *SuggestFieldsResult {
	result := &SuggestFieldsResult{SuggestedFields: []SuggestedField{}}

	response = strings.TrimSpace(response)
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx >= 0 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	type FieldSuggestion struct {
		Table  string `json:"table"`
		Column string `json:"column"`
		Reason string `json:"reason"`
	}
	type LLMResponse struct {
		SuggestedFields []FieldSuggestion `json:"suggested_fields"`
		AnalysisNote    string            `json:"analysis_note"`
	}

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(response), &llmResp); err != nil {
		result.AnalysisNote = "Failed to parse LLM response"
		return result
	}

	for _, field := range llmResp.SuggestedFields {
		result.SuggestedFields = append(result.SuggestedFields, SuggestedField{
			Name:        fmt.Sprintf("%s.%s", field.Table, field.Column),
			Description: field.Reason,
			Selected:    true,
			Source:      field.Table,
		})
	}
	result.AnalysisNote = llmResp.AnalysisNote
	return result
}
