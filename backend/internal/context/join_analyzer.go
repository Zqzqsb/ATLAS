package context

import (
	"fmt"
	"strings"
)

// AnalyzeJoinPaths 分析表之间的 JOIN 路径
func (c *SharedContext) AnalyzeJoinPaths() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 不再生成冗长的 JOIN 路径和字段语义信息
	// 这些信息会让 Rich Context 过于臃肿，造成注意力分散
	// 外键关系（foreign_keys）已经足够表达 JOIN 信息
	// LLM 可以根据外键关系自行推断 JOIN 路径

	// 保留数据结构初始化，避免 nil 指针
	if c.JoinPaths == nil {
		c.JoinPaths = make(map[string]*JoinPath)
	}
	if c.FieldSemantics == nil {
		c.FieldSemantics = make(map[string]*FieldSemantic)
	}
}

// buildForeignKeyGraph 构建外键关系图
func (c *SharedContext) buildForeignKeyGraph() map[string][]string {
	graph := make(map[string][]string)

	for tableName, table := range c.Tables {
		if _, exists := graph[tableName]; !exists {
			graph[tableName] = []string{}
		}

		for _, fk := range table.ForeignKeys {
			// 添加双向边（因为 JOIN 可以双向）
			graph[tableName] = append(graph[tableName], fk.ReferencedTable)
			if _, exists := graph[fk.ReferencedTable]; !exists {
				graph[fk.ReferencedTable] = []string{}
			}
			graph[fk.ReferencedTable] = append(graph[fk.ReferencedTable], tableName)
		}
	}

	return graph
}

// findShortestPath 使用 BFS 找到最短路径
func (c *SharedContext) findShortestPath(graph map[string][]string, from, to string) []string {
	if from == to {
		return []string{from}
	}

	visited := make(map[string]bool)
	queue := [][]string{{from}}
	visited[from] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]

		for _, neighbor := range graph[current] {
			if neighbor == to {
				return append(path, neighbor)
			}

			if !visited[neighbor] {
				visited[neighbor] = true
				newPath := make([]string, len(path))
				copy(newPath, path)
				newPath = append(newPath, neighbor)
				queue = append(queue, newPath)
			}
		}
	}

	return nil // 没有找到路径
}

// buildJoinPath 根据路径构建 JoinPath 对象
func (c *SharedContext) buildJoinPath(path []string) *JoinPath {
	if len(path) < 2 {
		return nil
	}

	joinClauses := []string{}
	description := ""

	for i := 0; i < len(path)-1; i++ {
		fromTable := path[i]
		toTable := path[i+1]

		// 找到连接条件
		joinClause := c.findJoinClause(fromTable, toTable)
		if joinClause != "" {
			joinClauses = append(joinClauses, joinClause)
		}
	}

	// 生成描述
	if len(path) == 2 {
		description = fmt.Sprintf("Direct join between %s and %s", path[0], path[1])
	} else {
		intermediates := path[1 : len(path)-1]
		description = fmt.Sprintf("Join through intermediate table(s): %s", strings.Join(intermediates, ", "))
	}

	return &JoinPath{
		FromTable:   path[0],
		ToTable:     path[len(path)-1],
		Path:        path,
		JoinClauses: joinClauses,
		Description: description,
	}
}

// findJoinClause 找到两个表之间的 JOIN 条件
func (c *SharedContext) findJoinClause(table1, table2 string) string {
	// 检查 table1 是否有指向 table2 的外键
	if t1, exists := c.Tables[table1]; exists {
		for _, fk := range t1.ForeignKeys {
			if fk.ReferencedTable == table2 {
				return fmt.Sprintf("%s.%s = %s.%s",
					table1, fk.ColumnName,
					table2, fk.ReferencedColumn)
			}
		}
	}

	// 检查 table2 是否有指向 table1 的外键
	if t2, exists := c.Tables[table2]; exists {
		for _, fk := range t2.ForeignKeys {
			if fk.ReferencedTable == table1 {
				return fmt.Sprintf("%s.%s = %s.%s",
					table2, fk.ColumnName,
					table1, fk.ReferencedColumn)
			}
		}
	}

	return ""
}

// reversePath 反转路径
func (c *SharedContext) reversePath(path []string) []string {
	reversed := make([]string, len(path))
	for i := 0; i < len(path); i++ {
		reversed[i] = path[len(path)-1-i]
	}
	return reversed
}

// analyzeFieldSemantics 分析字段语义
func (c *SharedContext) analyzeFieldSemantics() {
	for tableName, table := range c.Tables {
		// 分析外键字段
		for _, fk := range table.ForeignKeys {
			key := fmt.Sprintf("%s.%s", tableName, fk.ColumnName)
			c.FieldSemantics[key] = &FieldSemantic{
				TableName:   tableName,
				ColumnName:  fk.ColumnName,
				StorageType: "foreign_key",
				References:  fmt.Sprintf("%s.%s", fk.ReferencedTable, fk.ReferencedColumn),
				Note:        fmt.Sprintf("Stores %s ID, not %s name. Use JOIN to get related data.", fk.ReferencedTable, fk.ReferencedTable),
			}
		}

		// 分析可能的 ID 字段（但不是外键）
		for _, col := range table.Columns {
			colLower := strings.ToLower(col.Name)
			if strings.HasSuffix(colLower, "_id") || strings.HasSuffix(colLower, "id") {
				key := fmt.Sprintf("%s.%s", tableName, col.Name)

				// 检查是否已经作为外键处理
				if _, exists := c.FieldSemantics[key]; !exists {
					if col.IsPrimaryKey {
						c.FieldSemantics[key] = &FieldSemantic{
							TableName:   tableName,
							ColumnName:  col.Name,
							StorageType: "primary_key",
							Note:        fmt.Sprintf("Primary key of %s table", tableName),
						}
					} else {
						c.FieldSemantics[key] = &FieldSemantic{
							TableName:   tableName,
							ColumnName:  col.Name,
							StorageType: "id_field",
							Note:        "Likely an identifier field, may need special handling",
						}
					}
				}
			}
		}
	}
}

// FormatJoinPathsForPrompt 格式化 JOIN 路径信息用于 Prompt
func (c *SharedContext) FormatJoinPathsForPrompt() string {
	if len(c.JoinPaths) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## Join Path Guidelines\n")
	sb.WriteString("When joining tables, refer to these pre-analyzed join paths:\n\n")

	// 按表名排序输出
	for key, joinPath := range c.JoinPaths {
		sb.WriteString(fmt.Sprintf("**%s**:\n", key))
		sb.WriteString(fmt.Sprintf("  - Path: %s\n", strings.Join(joinPath.Path, " → ")))
		sb.WriteString(fmt.Sprintf("  - Description: %s\n", joinPath.Description))
		if len(joinPath.JoinClauses) > 0 {
			sb.WriteString("  - Join clauses:\n")
			for _, clause := range joinPath.JoinClauses {
				sb.WriteString(fmt.Sprintf("    * %s\n", clause))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatFieldSemanticsForPrompt 格式化字段语义信息用于 Prompt
func (c *SharedContext) FormatFieldSemanticsForPrompt() string {
	if len(c.FieldSemantics) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n## Field Semantics\n")
	sb.WriteString("Important field storage information:\n\n")

	// 按表名分组
	tableFields := make(map[string][]*FieldSemantic)
	for _, fs := range c.FieldSemantics {
		tableFields[fs.TableName] = append(tableFields[fs.TableName], fs)
	}

	for tableName, fields := range tableFields {
		sb.WriteString(fmt.Sprintf("**%s**:\n", tableName))
		for _, fs := range fields {
			sb.WriteString(fmt.Sprintf("  - %s (%s): %s\n", fs.ColumnName, fs.StorageType, fs.Note))
			if fs.References != "" {
				sb.WriteString(fmt.Sprintf("    References: %s\n", fs.References))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
