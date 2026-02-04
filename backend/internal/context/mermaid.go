package context

import (
	"fmt"
	"strings"
)

// GenerateMermaidER 生成 Mermaid ER 图
// 注意：调用此方法时必须已经持有锁
func (c *SharedContext) GenerateMermaidER() *SchemaDiagram {
	// 不需要加锁，调用者已经持有锁

	var sb strings.Builder

	// ER 图头部
	sb.WriteString("erDiagram\n")

	// 收集所有关系
	relationships := make(map[string]bool) // 用于去重

	for _, table := range c.Tables {
		for _, fk := range table.ForeignKeys {
			// 格式: TABLE1 ||--o{ TABLE2 : "relationship"
			// ||--o{ 表示 one-to-many (一对多)
			refTable := strings.ToUpper(fk.ReferencedTable)
			currentTable := strings.ToUpper(table.Name)

			// 关系描述：使用列名作为关系标签
			relationKey := fmt.Sprintf("%s_%s_%s", refTable, currentTable, fk.ColumnName)

			if !relationships[relationKey] {
				sb.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"has\"\n",
					refTable, currentTable))
				relationships[relationKey] = true
			}
		}
	}

	sb.WriteString("\n")

	// 添加表结构定义
	for _, table := range c.Tables {
		tableName := strings.ToUpper(table.Name)
		sb.WriteString(fmt.Sprintf("    %s {\n", tableName))

		// 添加列定义
		for _, col := range table.Columns {
			// 格式: type column_name PK/FK
			var tags []string

			if col.IsPrimaryKey {
				tags = append(tags, "PK")
			}

			// 检查是否是外键
			for _, fk := range table.ForeignKeys {
				if fk.ColumnName == col.Name {
					tags = append(tags, "FK")
					break
				}
			}

			tagStr := ""
			if len(tags) > 0 {
				tagStr = " " + strings.Join(tags, ",")
			}

			// 简化类型名称
			colType := simplifyType(col.Type)

			sb.WriteString(fmt.Sprintf("        %s %s%s\n",
				colType, col.Name, tagStr))
		}

		sb.WriteString("    }\n")
	}

	return &SchemaDiagram{
		Format:      "mermaid-er",
		Description: "Entity-Relationship diagram showing database schema and relationships",
		Content:     sb.String(),
	}
}

// simplifyType 简化类型名称用于 Mermaid 显示
func simplifyType(fullType string) string {
	// 移除长度限制，只保留基本类型
	fullType = strings.ToLower(fullType)

	if strings.Contains(fullType, "int") {
		return "int"
	}
	if strings.Contains(fullType, "varchar") || strings.Contains(fullType, "char") {
		return "string"
	}
	if strings.Contains(fullType, "text") {
		return "text"
	}
	if strings.Contains(fullType, "real") || strings.Contains(fullType, "float") || strings.Contains(fullType, "double") {
		return "float"
	}
	if strings.Contains(fullType, "date") || strings.Contains(fullType, "time") {
		return "datetime"
	}
	if strings.Contains(fullType, "bool") {
		return "boolean"
	}

	return "string" // 默认
}
