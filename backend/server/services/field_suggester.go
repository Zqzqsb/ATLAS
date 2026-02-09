package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"

	"lucid/config"
	"lucid/interfaces"
)

// FieldSuggester suggests output fields based on LLM analysis.
// Previously lived in bridge/context_bridge.go as FieldSuggesterBridge.
type FieldSuggester struct {
	llmModel       llms.Model
	adapterFactory interfaces.AdapterFactory
	config         *config.Config
}

// NewFieldSuggester creates a field suggester.
func NewFieldSuggester(llmModel llms.Model, factory interfaces.AdapterFactory, cfg *config.Config) *FieldSuggester {
	return &FieldSuggester{
		llmModel:       llmModel,
		adapterFactory: factory,
		config:         cfg,
	}
}

func (s *FieldSuggester) SuggestFields(c context.Context, req *interfaces.SuggestFieldsRequest) (*interfaces.SuggestFieldsResult, error) {
	if s.llmModel == nil {
		return &interfaces.SuggestFieldsResult{
			SuggestedFields: []interfaces.SuggestedField{},
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

	var adapterCfg *interfaces.DBConfig
	if s.config != nil {
		for _, db := range s.config.Databases {
			if db.ID == dbID {
				adapterCfg = &interfaces.DBConfig{
					Type:     db.Type,
					Host:     db.Host,
					Port:     db.Port,
					Database: database,
					User:     db.User,
					Password: db.Password,
					FilePath: db.Path,
				}
				break
			}
		}
	}

	if adapterCfg == nil {
		adapterCfg = &interfaces.DBConfig{
			Type:     "mariadb",
			Database: database,
		}
	}

	adp, err := s.adapterFactory(adapterCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create adapter: %w", err)
	}
	defer adp.Close()

	if err := adp.Connect(c); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	var schemaBuilder strings.Builder
	schemaBuilder.WriteString("Database Schema:\n\n")

	tablesQuery := `
		SELECT TABLE_NAME 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_TYPE = 'BASE TABLE'
	`
	tablesResult, err := adp.ExecuteQuery(c, tablesQuery)
	if err != nil {
		return "", fmt.Errorf("failed to get tables: %w", err)
	}

	for _, row := range tablesResult.Rows {
		tableName, ok := row["TABLE_NAME"].(string)
		if !ok {
			continue
		}

		schemaBuilder.WriteString(fmt.Sprintf("Table: %s\n", tableName))

		columnsQuery := fmt.Sprintf(`
			SELECT COLUMN_NAME, DATA_TYPE, COLUMN_COMMENT
			FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION
		`, tableName)

		columnsResult, err := adp.ExecuteQuery(c, columnsQuery)
		if err != nil {
			continue
		}

		schemaBuilder.WriteString("  Columns:\n")
		for _, colRow := range columnsResult.Rows {
			colName, _ := colRow["COLUMN_NAME"].(string)
			dataType, _ := colRow["DATA_TYPE"].(string)
			comment, _ := colRow["COLUMN_COMMENT"].(string)

			if comment != "" {
				schemaBuilder.WriteString(fmt.Sprintf("    - %s (%s): %s\n", colName, dataType, comment))
			} else {
				schemaBuilder.WriteString(fmt.Sprintf("    - %s (%s)\n", colName, dataType))
			}
		}
		schemaBuilder.WriteString("\n")
	}

	return schemaBuilder.String(), nil
}

func (s *FieldSuggester) buildFieldSuggestionPrompt(question, schema, language string) string {
	langNote := "English"
	if language == "Chinese" || language == "zh" {
		langNote = "Chinese"
	}

	return fmt.Sprintf(`You are a database expert. Analyze the user's question and the database schema to suggest which fields (columns) should appear in the SELECT statement output.

User Question: %s

%s

Task: Based on the question, identify the most relevant output fields that the user would want to see in the query result.

Instructions:
1. Consider what information the user is asking for
2. Include fields that directly answer the question
3. Include helpful context fields (e.g., names, descriptions) that make results meaningful
4. Don't include internal IDs unless specifically requested
5. Provide the response in %s

Output Format (JSON):
{
  "suggested_fields": [
    {"table": "table_name", "column": "column_name", "reason": "why this field is relevant"},
    ...
  ],
  "analysis_note": "Brief explanation of the analysis"
}

Only output valid JSON, no other text.`, question, schema, langNote)
}

func (s *FieldSuggester) parseFieldSuggestionResponse(response string) *interfaces.SuggestFieldsResult {
	result := &interfaces.SuggestFieldsResult{
		SuggestedFields: []interfaces.SuggestedField{},
	}

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
		result.SuggestedFields = append(result.SuggestedFields, interfaces.SuggestedField{
			Name:        fmt.Sprintf("%s.%s", field.Table, field.Column),
			Description: field.Reason,
			Selected:    true,
			Source:      field.Table,
		})
	}
	result.AnalysisNote = llmResp.AnalysisNote
	return result
}
