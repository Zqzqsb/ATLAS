package bridge

import (
	"context"
	"os"
	"path/filepath"
	"time"

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
	llmModel       interface{}
	adapterFactory interfaces.AdapterFactory
	config         interface{}
}

// NewFieldSuggesterBridge 创建字段建议器桥接
func NewFieldSuggesterBridge(llmModel interface{}, factory interfaces.AdapterFactory, cfg interface{}) *FieldSuggesterBridge {
	return &FieldSuggesterBridge{
		llmModel:       llmModel,
		adapterFactory: factory,
		config:         cfg,
	}
}

// SuggestFields 建议字段
func (s *FieldSuggesterBridge) SuggestFields(ctx context.Context, req *interfaces.SuggestFieldsRequest) (*interfaces.SuggestFieldsResult, error) {
	// TODO: 实现基于 LLM 的字段建议
	return &interfaces.SuggestFieldsResult{
		SuggestedFields: []interfaces.SuggestedField{},
		AnalysisNote:    "Field suggestion not yet implemented",
	}, nil
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
