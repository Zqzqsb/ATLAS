package context

import (
	"fmt"
	"strings"
	"time"
)

// ExportOptions 导出选项
type ExportOptions struct {
	// 指定要导出的表（为空则导出所有表）
	Tables []string
	// 是否包含详细的列信息
	IncludeColumns bool
	// 是否包含索引信息
	IncludeIndexes bool
	// 是否包含 Rich Context
	IncludeRichContext bool
	// 是否包含统计信息
	IncludeStats bool
}

// DefaultExportOptions 默认导出选项
func DefaultExportOptions() *ExportOptions {
	return &ExportOptions{
		Tables:             nil, // 导出所有表
		IncludeColumns:     true,
		IncludeIndexes:     true,
		IncludeRichContext: true,
		IncludeStats:       true,
	}
}

// ExportToPrompt 导出为 LLM 友好的 Prompt 格式
func (c *SharedContext) ExportToPrompt(opts *ExportOptions) string {
	if opts == nil {
		opts = DefaultExportOptions()
	}

	var sb strings.Builder

	// 1. 数据库概览
	sb.WriteString("# Database Schema Context\n\n")
	sb.WriteString(fmt.Sprintf("**Database**: %s (%s)\n", c.DatabaseName, c.DatabaseType))

	// 过滤要导出的表
	tables := c.filterTables(opts.Tables)
	sb.WriteString(fmt.Sprintf("**Tables**: %d\n\n", len(tables)))

	// 2. 表详情
	for _, tableName := range tables {
		table, exists := c.Tables[tableName]
		if !exists {
			continue
		}

		sb.WriteString("---\n\n")
		sb.WriteString(fmt.Sprintf("## Table: `%s`\n\n", table.Name))

		// 统计信息
		if opts.IncludeStats {
			sb.WriteString(fmt.Sprintf("- **Row Count**: %d\n", table.RowCount))
			sb.WriteString(fmt.Sprintf("- **Columns**: %d\n", len(table.Columns)))
			sb.WriteString(fmt.Sprintf("- **Indexes**: %d\n\n", len(table.Indexes)))
		}

		// 表注释
		if table.Comment != "" {
			sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", table.Comment))
		}

		// Rich Context（业务信息和数据质量问题）
		if opts.IncludeRichContext && len(table.RichContext) > 0 {
			// 分离数据质量问题和业务信息
			qualityIssues := make(map[string]RichContextValue)
			businessContext := make(map[string]RichContextValue)

			for key, note := range table.RichContext {
				if strings.Contains(key, "quality_issue") || strings.Contains(key, "orphan_issue") {
					qualityIssues[key] = note
				} else {
					businessContext[key] = note
				}
			}

			// 优先显示数据质量问题（更重要）
			if len(qualityIssues) > 0 {
				sb.WriteString("### ⚠️ Data Quality Issues\n\n")
				sb.WriteString("> **CRITICAL**: These issues directly affect SQL query correctness.\n\n")
				for key, note := range qualityIssues {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", formatKey(key), note.Content))
				}
				sb.WriteString("\n")
			}

			// 然后显示业务上下文
			if len(businessContext) > 0 {
				sb.WriteString("### Business Context\n\n")
				for key, note := range businessContext {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", formatKey(key), note.Content))
				}
				sb.WriteString("\n")
			}
		}

		// 列信息
		if opts.IncludeColumns && len(table.Columns) > 0 {
			sb.WriteString("### Columns\n\n")
			sb.WriteString("| Column | Type | Nullable | Default | Key | Comment |\n")
			sb.WriteString("|--------|------|----------|---------|-----|----------|\n")

			for _, col := range table.Columns {
				nullable := "NO"
				if col.Nullable {
					nullable = "YES"
				}

				key := ""
				if col.IsPrimaryKey {
					key = "PRI"
				}

				defaultVal := col.DefaultValue
				if defaultVal == "" {
					defaultVal = "-"
				}

				comment := col.Comment
				if comment == "" {
					comment = "-"
				}

				sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s | %s |\n",
					col.Name, col.Type, nullable, defaultVal, key, comment))
			}
			sb.WriteString("\n")
		}

		// 索引信息
		if opts.IncludeIndexes && len(table.Indexes) > 0 {
			sb.WriteString("### Indexes\n\n")
			for _, idx := range table.Indexes {
				indexType := "INDEX"
				if idx.IsPrimary {
					indexType = "PRIMARY KEY"
				} else if idx.IsUnique {
					indexType = "UNIQUE INDEX"
				}

				sb.WriteString(fmt.Sprintf("- **%s** `%s` on (%s)\n",
					indexType, idx.Name, strings.Join(idx.Columns, ", ")))
			}
			sb.WriteString("\n")
		}

		// 外键关系（简洁格式）
		if len(table.ForeignKeys) > 0 {
			sb.WriteString("### Foreign Keys\n\n")
			for _, fk := range table.ForeignKeys {
				sb.WriteString(fmt.Sprintf("- `%s` → `%s.%s`\n",
					fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ExportTableToPrompt 导出单个表为 Prompt 格式
func (c *SharedContext) ExportTableToPrompt(tableName string) string {
	opts := DefaultExportOptions()
	opts.Tables = []string{tableName}
	return c.ExportToPrompt(opts)
}

// ExportToCompactPrompt 导出为紧凑的 Prompt 格式（用于 Schema Linking）
func (c *SharedContext) ExportToCompactPrompt(opts *ExportOptions) string {
	if opts == nil {
		opts = DefaultExportOptions()
	}

	var sb strings.Builder

	// 数据库信息
	sb.WriteString(fmt.Sprintf("Database: %s\n\n", c.DatabaseName))

	// 过滤要导出的表
	tables := c.filterTables(opts.Tables)

	for _, tableName := range tables {
		table, exists := c.Tables[tableName]
		if !exists {
			continue
		}

		// 表名和行数
		sb.WriteString(fmt.Sprintf("Table %s (%d rows):\n", table.Name, table.RowCount))

		// 列信息（紧凑格式）
		if opts.IncludeColumns {
			for _, col := range table.Columns {
				pk := ""
				if col.IsPrimaryKey {
					pk = " [PK]"
				}
				// 检查是否是外键
				fkInfo := ""
				for _, fk := range table.ForeignKeys {
					if fk.ColumnName == col.Name {
						fkInfo = fmt.Sprintf(" → %s.%s", fk.ReferencedTable, fk.ReferencedColumn)
						break
					}
				}
				sb.WriteString(fmt.Sprintf("  - %s: %s%s%s\n", col.Name, col.Type, pk, fkInfo))
			}
		}

		// Rich Context（关键业务信息和数据质量问题）
		if opts.IncludeRichContext && len(table.RichContext) > 0 {
			// 分离数据质量问题和业务信息
			qualityIssues := make(map[string]RichContextValue)
			businessNotes := make(map[string]RichContextValue)

			for key, note := range table.RichContext {
				if strings.Contains(key, "quality_issue") || strings.Contains(key, "orphan_issue") {
					qualityIssues[key] = note
				} else {
					businessNotes[key] = note
				}
			}

			// 优先显示数据质量问题
			if len(qualityIssues) > 0 {
				sb.WriteString("  ⚠️ Data Quality Issues:\n")
				for key, note := range qualityIssues {
					expiredTag := ""
					if note.ExpiresAt != "" {
						expiresAt, err := time.Parse(time.RFC3339, note.ExpiresAt)
						if err == nil && time.Now().After(expiresAt) {
							expiredTag = " [EXPIRED]"
						}
					}
					sb.WriteString(fmt.Sprintf("    * [%s] %s%s\n", key, note.Content, expiredTag))
				}
			}

			// 然后显示业务信息
			if len(businessNotes) > 0 {
				sb.WriteString("  Business Notes:\n")
				for key, note := range businessNotes {
					expiredTag := ""
					if note.ExpiresAt != "" {
						expiresAt, err := time.Parse(time.RFC3339, note.ExpiresAt)
						if err == nil && time.Now().After(expiresAt) {
							expiredTag = " [EXPIRED]"
						}
					}
					sb.WriteString(fmt.Sprintf("    * [%s] %s: %s%s\n", key, formatKey(key), note.Content, expiredTag))
				}
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// ExportToSchemaLinking 导出为 Schema Linking 格式（最紧凑）
func (c *SharedContext) ExportToSchemaLinking(tableNames []string) string {
	var sb strings.Builder

	tables := c.filterTables(tableNames)

	for i, tableName := range tables {
		table, exists := c.Tables[tableName]
		if !exists {
			continue
		}

		if i > 0 {
			sb.WriteString(" | ")
		}

		// 表名
		sb.WriteString(table.Name)
		sb.WriteString(" (")

		// 列名列表
		colNames := make([]string, len(table.Columns))
		for j, col := range table.Columns {
			if col.IsPrimaryKey {
				colNames[j] = col.Name + "*"
			} else {
				colNames[j] = col.Name
			}
		}
		sb.WriteString(strings.Join(colNames, ", "))
		sb.WriteString(")")
	}

	return sb.String()
}

// filterTables 过滤要导出的表
func (c *SharedContext) filterTables(tableNames []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 如果没有指定表，返回所有表
	if len(tableNames) == 0 {
		result := make([]string, 0, len(c.Tables))
		for name := range c.Tables {
			result = append(result, name)
		}
		return result
	}

	// 过滤存在的表
	result := make([]string, 0, len(tableNames))
	for _, name := range tableNames {
		if _, exists := c.Tables[name]; exists {
			result = append(result, name)
		}
	}
	return result
}
