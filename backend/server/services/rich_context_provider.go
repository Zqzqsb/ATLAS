package services

import (
	"os"
	"path/filepath"
	"time"

	ctx "lucid/internal/context"
)

// FileRichContextProvider loads Rich Context from filesystem JSON files.
type FileRichContextProvider struct {
	contextPaths []string
}

func NewFileRichContextProvider() *FileRichContextProvider {
	return &FileRichContextProvider{
		contextPaths: []string{
			"data/spider/rich_context",
			"contexts/sqlite",
			"contexts/mysql",
			"contexts/postgres",
			"data/contexts",
		},
	}
}

func (p *FileRichContextProvider) GetRichContext(dbID, database string) (*RichContextInfo, error) {
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

	return &RichContextInfo{
		Database:  database,
		Tables:    []TableContextInfo{},
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}, nil
}

func (p *FileRichContextProvider) HasRichContext(database string) bool {
	for _, basePath := range p.contextPaths {
		contextPath := filepath.Join(basePath, database+".json")
		if _, err := os.Stat(contextPath); err == nil {
			return true
		}
	}
	return false
}

func convertSharedContextToInfo(sharedCtx *ctx.SharedContext) *RichContextInfo {
	info := &RichContextInfo{
		Database:  sharedCtx.DatabaseName,
		Tables:    []TableContextInfo{},
		UpdatedAt: sharedCtx.CollectedAt,
		Version:   sharedCtx.Version,
	}

	for _, table := range sharedCtx.Tables {
		tableInfo := TableContextInfo{
			Name:        table.Name,
			Description: table.Description,
			Columns:     []ColumnContextInfo{},
		}
		for _, col := range table.Columns {
			tableInfo.Columns = append(tableInfo.Columns, ColumnContextInfo{
				Name:        col.Name,
				Description: col.Comment,
			})
		}
		info.Tables = append(info.Tables, tableInfo)
	}

	return info
}
