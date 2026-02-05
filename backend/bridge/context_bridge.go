package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/config"
	"lucid/interfaces"
	ctx "lucid/internal/context"
)

// RichContextProviderBridge Rich Context 提供者桥接
type RichContextProviderBridge struct {
	contextPaths []string
}

// NewRichContextProviderBridge 创建 Rich Context 提供者桥接
func NewRichContextProviderBridge() *RichContextProviderBridge {
	return &RichContextProviderBridge{
		contextPaths: []string{
			"data/spider/rich_context",
			"contexts/sqlite",
			"contexts/mysql",
			"contexts/postgres",
			"data/contexts",
		},
	}
}

// GetRichContext 获取 Rich Context
func (p *RichContextProviderBridge) GetRichContext(dbID, database string) (*interfaces.RichContextInfo, error) {
	// 尝试从多个路径加载
	for _, basePath := range p.contextPaths {
		contextPath := filepath.Join(basePath, database+".json")
		if _, err := os.Stat(contextPath); err == nil {
			sharedCtx, err := ctx.LoadContextFromFile(contextPath)
			if err != nil {
				continue
			}
			return convertSharedContextToInfo(sharedCtx), nil
		}
	}

	// 创建空的 context
	return &interfaces.RichContextInfo{
		Database:  database,
		Tables:    []interfaces.TableContextInfo{},
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}, nil
}

// HasRichContext 检查是否存在 Rich Context
func (p *RichContextProviderBridge) HasRichContext(database string) bool {
	for _, basePath := range p.contextPaths {
		contextPath := filepath.Join(basePath, database+".json")
		if _, err := os.Stat(contextPath); err == nil {
			return true
		}
	}
	return false
}

// convertSharedContextToInfo 将 SharedContext 转换为 RichContextInfo
func convertSharedContextToInfo(sharedCtx *ctx.SharedContext) *interfaces.RichContextInfo {
	info := &interfaces.RichContextInfo{
		Database:  sharedCtx.DatabaseName,
		Tables:    []interfaces.TableContextInfo{},
		UpdatedAt: sharedCtx.CollectedAt,
		Version:   sharedCtx.Version,
	}

	// SharedContext.Tables 是 map[string]*TableMetadata
	for _, table := range sharedCtx.Tables {
		tableInfo := interfaces.TableContextInfo{
			Name:        table.Name,
			Description: table.Description,
			Columns:     []interfaces.ColumnContextInfo{},
		}

		// 转换列信息
		for _, col := range table.Columns {
			colInfo := interfaces.ColumnContextInfo{
				Name:        col.Name,
				Description: col.Comment, // ColumnMetadata 使用 Comment 而不是 Description
			}
			tableInfo.Columns = append(tableInfo.Columns, colInfo)
		}

		info.Tables = append(info.Tables, tableInfo)
	}

	return info
}

// FieldSuggesterBridge 字段建议器桥接
type FieldSuggesterBridge struct {
	llmModel       llms.Model
	adapterFactory interfaces.AdapterFactory
	config         *config.Config
}

// NewFieldSuggesterBridge 创建字段建议器桥接
func NewFieldSuggesterBridge(llmModel interface{}, factory interfaces.AdapterFactory, cfg interface{}) *FieldSuggesterBridge {
	bridge := &FieldSuggesterBridge{
		adapterFactory: factory,
	}

	// Type assert llmModel
	if model, ok := llmModel.(llms.Model); ok {
		bridge.llmModel = model
	}

	// Type assert config
	if c, ok := cfg.(*config.Config); ok {
		bridge.config = c
	}

	return bridge
}

// SuggestFields 基于 LLM 分析问题和 schema，建议候选字段
func (s *FieldSuggesterBridge) SuggestFields(c context.Context, req *interfaces.SuggestFieldsRequest) (*interfaces.SuggestFieldsResult, error) {
	if s.llmModel == nil {
		return &interfaces.SuggestFieldsResult{
			SuggestedFields: []interfaces.SuggestedField{},
			AnalysisNote:    "LLM not available for field suggestion",
		}, nil
	}

	// Get schema from database
	schema, err := s.getSchemaForSuggestion(c, req.DatabaseID, req.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	// Build prompt for field suggestion
	prompt := s.buildFieldSuggestionPrompt(req.Question, schema, req.Language)

	// Call LLM
	response, err := s.llmModel.Call(c, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse response
	result := s.parseFieldSuggestionResponse(response, schema)
	return result, nil
}

// getSchemaForSuggestion 获取用于字段建议的 schema 信息
func (s *FieldSuggesterBridge) getSchemaForSuggestion(c context.Context, dbID, database string) (string, error) {
	if s.adapterFactory == nil {
		return "", fmt.Errorf("adapter factory not available")
	}

	// Try to get database config and create adapter
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
		// Use database ID directly as config identifier
		adapterCfg = &interfaces.DBConfig{
			Type:     "mariadb",
			Database: database,
		}
	}

	adapter, err := s.adapterFactory(adapterCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create adapter: %w", err)
	}
	defer adapter.Close()

	if err := adapter.Connect(c); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// Query schema information
	var schemaBuilder strings.Builder
	schemaBuilder.WriteString("Database Schema:\n\n")

	// Get tables
	tablesQuery := `
		SELECT TABLE_NAME 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_TYPE = 'BASE TABLE'
	`
	tablesResult, err := adapter.ExecuteQuery(c, tablesQuery)
	if err != nil {
		return "", fmt.Errorf("failed to get tables: %w", err)
	}

	for _, row := range tablesResult.Rows {
		tableName, ok := row["TABLE_NAME"].(string)
		if !ok {
			continue
		}

		schemaBuilder.WriteString(fmt.Sprintf("Table: %s\n", tableName))

		// Get columns for this table
		columnsQuery := fmt.Sprintf(`
			SELECT COLUMN_NAME, DATA_TYPE, COLUMN_COMMENT
			FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION
		`, tableName)

		columnsResult, err := adapter.ExecuteQuery(c, columnsQuery)
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

// buildFieldSuggestionPrompt 构建字段建议的 prompt
func (s *FieldSuggesterBridge) buildFieldSuggestionPrompt(question, schema, language string) string {
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

// parseFieldSuggestionResponse 解析 LLM 响应
func (s *FieldSuggesterBridge) parseFieldSuggestionResponse(response, schema string) *interfaces.SuggestFieldsResult {
	result := &interfaces.SuggestFieldsResult{
		SuggestedFields: []interfaces.SuggestedField{},
		AnalysisNote:    "",
	}

	// Try to parse JSON response
	type FieldSuggestion struct {
		Table  string `json:"table"`
		Column string `json:"column"`
		Reason string `json:"reason"`
	}

	type LLMResponse struct {
		SuggestedFields []FieldSuggestion `json:"suggested_fields"`
		AnalysisNote    string            `json:"analysis_note"`
	}

	// Clean response - find JSON object
	response = strings.TrimSpace(response)
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx >= 0 && endIdx > startIdx {
		response = response[startIdx : endIdx+1]
	}

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(response), &llmResp); err != nil {
		// If parsing fails, return empty result with note
		result.AnalysisNote = "Failed to parse LLM response"
		return result
	}

	// Convert to interface format
	for _, field := range llmResp.SuggestedFields {
		result.SuggestedFields = append(result.SuggestedFields, interfaces.SuggestedField{
			Name:        fmt.Sprintf("%s.%s", field.Table, field.Column),
			Description: field.Reason,
			Selected:    true, // Default to selected
			Source:      field.Table,
		})
	}

	result.AnalysisNote = llmResp.AnalysisNote
	return result
}

// TranslatorBridge 翻译器桥接
type TranslatorBridge struct {
	llmModel interface{}
}

// NewTranslatorBridge 创建翻译器桥接
func NewTranslatorBridge(llmModel interface{}) *TranslatorBridge {
	return &TranslatorBridge{llmModel: llmModel}
}

// TranslateTexts 翻译文本
func (t *TranslatorBridge) TranslateTexts(ctx context.Context, texts []string, targetLang string) (map[string]string, error) {
	// TODO: 实现基于 LLM 的翻译
	result := make(map[string]string)
	for _, text := range texts {
		result[text] = text // 占位：返回原文
	}
	return result, nil
}
