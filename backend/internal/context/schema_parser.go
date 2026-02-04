package context

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SchemaParser 解析 schema.sql 文件
type SchemaParser struct {
	filePath string
}

// NewSchemaParser 创建解析器
func NewSchemaParser(filePath string) *SchemaParser {
	return &SchemaParser{filePath: filePath}
}

// ParsedTable 解析后的表结构
type ParsedTable struct {
	Name        string
	Columns     map[string]string // column_name -> type
	PrimaryKeys []string
	ForeignKeys []ParsedForeignKey
}

// ParsedForeignKey 解析后的外键
type ParsedForeignKey struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
}

// Parse 解析 schema.sql 文件
func (p *SchemaParser) Parse() (map[string]*ParsedTable, error) {
	content, err := os.ReadFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	sql := string(content)

	// 移除注释
	sql = removeComments(sql)

	// 提取所有 CREATE TABLE 语句
	tables := make(map[string]*ParsedTable)

	// 正则匹配 CREATE TABLE 语句
	// 支持带引号和不带引号的表名
	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["'\x60]?(\w+)["'\x60]?\s*\(((?:[^()]|\([^)]*\))*)\)`)

	matches := createTableRegex.FindAllStringSubmatch(sql, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		tableName := match[1] // 保留原始大小写
		tableBody := match[2]

		table := &ParsedTable{
			Name:        strings.ToLower(tableName), // 表名统一小写用于查找
			Columns:     make(map[string]string),
			PrimaryKeys: []string{},
			ForeignKeys: []ParsedForeignKey{},
		}

		// 解析表体
		parseTableBody(table, tableBody)

		tables[tableName] = table
	}

	return tables, nil
}

// removeComments 移除 SQL 注释
func removeComments(sql string) string {
	// 移除单行注释 --
	lineCommentRegex := regexp.MustCompile(`--[^\n]*`)
	sql = lineCommentRegex.ReplaceAllString(sql, "")

	// 移除多行注释 /* */
	blockCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	sql = blockCommentRegex.ReplaceAllString(sql, "")

	return sql
}

// parseTableBody 解析表定义体
func parseTableBody(table *ParsedTable, body string) {
	// 分割成各个定义项（列定义、约束等）
	items := splitTableItems(body)

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		itemLower := strings.ToLower(item)

		// 检查是否是主键约束
		if strings.HasPrefix(itemLower, "primary key") {
			parsePrimaryKey(table, item)
			continue
		}

		// 检查是否是外键约束
		if strings.HasPrefix(itemLower, "foreign key") {
			parseForeignKey(table, item)
			continue
		}

		// 否则是列定义
		parseColumnDefinition(table, item)
	}
}

// splitTableItems 分割表定义项（处理括号嵌套）
func splitTableItems(body string) []string {
	var items []string
	var current strings.Builder
	parenDepth := 0

	for _, ch := range body {
		switch ch {
		case '(':
			parenDepth++
			current.WriteRune(ch)
		case ')':
			parenDepth--
			current.WriteRune(ch)
		case ',':
			if parenDepth == 0 {
				items = append(items, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		items = append(items, current.String())
	}

	return items
}

// parseColumnDefinition 解析列定义
func parseColumnDefinition(table *ParsedTable, def string) {
	// 移除引号
	def = strings.Trim(def, `"'\x60`)

	// 分割列名和类型
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return
	}

	columnName := strings.Trim(parts[0], `"'\x60`) // 保留原始大小写
	columnType := strings.ToUpper(parts[1])

	// 检查是否有 primary key 关键字（列级约束）
	defLower := strings.ToLower(def)
	if strings.Contains(defLower, "primary key") {
		table.PrimaryKeys = append(table.PrimaryKeys, columnName)
	}

	// 检查是否有外键（列级约束）
	// 格式: REFERENCES table(column)
	if strings.Contains(defLower, "references") {
		parseForeignKeyInline(table, columnName, def)
	}

	table.Columns[columnName] = columnType
}

// parsePrimaryKey 解析主键约束
func parsePrimaryKey(table *ParsedTable, constraint string) {
	// 提取括号内的列名
	// 格式: primary key ("col1", "col2")
	re := regexp.MustCompile(`\((.*?)\)`)
	matches := re.FindStringSubmatch(constraint)
	if len(matches) < 2 {
		return
	}

	columnsStr := matches[1]
	columns := strings.Split(columnsStr, ",")

	for _, col := range columns {
		col = strings.TrimSpace(col)
		col = strings.Trim(col, `"'\x60`)
		// 保留原始大小写
		if col != "" {
			table.PrimaryKeys = append(table.PrimaryKeys, col)
		}
	}
}

// parseForeignKey 解析外键约束（表级）
func parseForeignKey(table *ParsedTable, constraint string) {
	// 格式: foreign key("col") references `table`("ref_col")

	// 提取本表列名（保留大小写）
	colRe := regexp.MustCompile(`(?i)foreign\s+key\s*\(\s*["'\x60]?(\w+)["'\x60]?\s*\)`)
	colMatches := colRe.FindStringSubmatch(constraint)
	if len(colMatches) < 2 {
		return
	}
	columnName := colMatches[1]

	// 提取引用的表和列（保留大小写）
	refRe := regexp.MustCompile(`(?i)references\s+["'\x60]?(\w+)["'\x60]?\s*\(\s*["'\x60]?(\w+)["'\x60]?\s*\)`)
	refMatches := refRe.FindStringSubmatch(constraint)
	if len(refMatches) < 3 {
		return
	}

	table.ForeignKeys = append(table.ForeignKeys, ParsedForeignKey{
		ColumnName:       columnName,
		ReferencedTable:  strings.ToLower(refMatches[1]), // 表名统一小写
		ReferencedColumn: refMatches[2],                  // 列名保留原始大小写
	})
}

// parseForeignKeyInline 解析内联外键（列级）
func parseForeignKeyInline(table *ParsedTable, columnName, def string) {
	// 格式: REFERENCES table(column)
	refRe := regexp.MustCompile(`(?i)references\s+["'\x60]?(\w+)["'\x60]?\s*\(\s*["'\x60]?(\w+)["'\x60]?\s*\)`)
	refMatches := refRe.FindStringSubmatch(def)
	if len(refMatches) < 3 {
		return
	}

	table.ForeignKeys = append(table.ForeignKeys, ParsedForeignKey{
		ColumnName:       columnName,
		ReferencedTable:  strings.ToLower(refMatches[1]), // 表名统一小写
		ReferencedColumn: refMatches[2],                  // 列名保留原始大小写
	})
}
