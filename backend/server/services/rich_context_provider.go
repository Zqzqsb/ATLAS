package services

import (
	"os"
	"path/filepath"
	"time"

	"lucid/interfaces"
	ctx "lucid/internal/context"
)

// FileRichContextProvider loads Rich Context from filesystem JSON files.
// Previously lived in bridge/context_bridge.go as RichContextProviderBridge.
type FileRichContextProvider struct {
	contextPaths []string
}

// NewFileRichContextProvider creates a new file-based Rich Context provider.
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

func (p *FileRichContextProvider) GetRichContext(dbID, database string) (*interfaces.RichContextInfo, error) {
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

	return &interfaces.RichContextInfo{
		Database:  database,
		Tables:    []interfaces.TableContextInfo{},
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

func convertSharedContextToInfo(sharedCtx *ctx.SharedContext) *interfaces.RichContextInfo {
	info := &interfaces.RichContextInfo{
		Database:  sharedCtx.DatabaseName,
		Tables:    []interfaces.TableContextInfo{},
		UpdatedAt: sharedCtx.CollectedAt,
		Version:   sharedCtx.Version,
	}

	for _, table := range sharedCtx.Tables {
		tableInfo := interfaces.TableContextInfo{
			Name:        table.Name,
			Description: table.Description,
			Columns:     []interfaces.ColumnContextInfo{},
		}
		for _, col := range table.Columns {
			colInfo := interfaces.ColumnContextInfo{
				Name:        col.Name,
				Description: col.Comment,
			}
			tableInfo.Columns = append(tableInfo.Columns, colInfo)
		}
		info.Tables = append(info.Tables, tableInfo)
	}

	return info
}
