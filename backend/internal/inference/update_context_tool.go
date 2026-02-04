package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UpdateRichContextTool 更新 Rich Context 的工具
type UpdateRichContextTool struct {
	dbName      string
	contextPath string
}

// Name 工具名称
func (t *UpdateRichContextTool) Name() string {
	return "update_rich_context"
}

// Description 工具描述
func (t *UpdateRichContextTool) Description() string {
	return `Update expired or incorrect Rich Context information.
Use this tool when you find an EXPIRED insight is incorrect after verification.

Input should be a JSON object with:
- table_name: string (the table name)
- note_key: string (the Rich Context note key to update)
- new_content: string (the corrected content)
- reason: string (why you're updating it, based on SQL verification)

Example:
{
  "table_name": "cars_data",
  "note_key": "Year_business_meaning",
  "new_content": "Model year (4-digit format, 1970-1982)",
  "reason": "Verified by SQL: SELECT DISTINCT Year shows 4-digit years, not 2-digit"
}`
}

// UpdateInput 更新参数
type UpdateInput struct {
	TableName  string `json:"table_name"`
	NoteKey    string `json:"note_key"`
	NewContent string `json:"new_content"`
	Reason     string `json:"reason"`
}

// BusinessNote Rich Context 条目
type BusinessNote struct {
	Content   string `json:"content"`
	ExpiresAt string `json:"expires_at"`
}

// Call 执行更新
func (t *UpdateRichContextTool) Call(ctx context.Context, input string) (string, error) {
	// 清理输入：移除可能的 markdown 代码块标记
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```")
	input = strings.TrimSuffix(input, "```")
	input = strings.TrimSpace(input)

	// 解析输入
	var updateInput UpdateInput
	if err := json.Unmarshal([]byte(input), &updateInput); err != nil {
		// 返回友好的错误信息，但不中断推理
		return fmt.Sprintf("⚠️  Failed to parse input JSON: %v\nPlease provide valid JSON without markdown code blocks.", err), nil
	}

	// 验证参数
	if updateInput.TableName == "" {
		return "⚠️  Error: table_name is required", nil
	}
	if updateInput.NoteKey == "" {
		return "⚠️  Error: note_key is required", nil
	}
	if updateInput.NewContent == "" {
		return "⚠️  Error: new_content is required", nil
	}

	// 读取 Rich Context 文件
	data, err := os.ReadFile(t.contextPath)
	if err != nil {
		return fmt.Sprintf("⚠️  Failed to read context file: %v\nContinue with SQL generation.", err), nil
	}

	// 解析为通用 map
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return fmt.Sprintf("⚠️  Failed to parse context file: %v\nContinue with SQL generation.", err), nil
	}

	// 获取 tables
	tables, ok := rawData["tables"].(map[string]interface{})
	if !ok {
		return "⚠️  No tables field in context. Continue with SQL generation.", nil
	}

	// 获取指定表
	tableData, ok := tables[updateInput.TableName].(map[string]interface{})
	if !ok {
		return fmt.Sprintf("⚠️  Table '%s' not found in context. Continue with SQL generation.", updateInput.TableName), nil
	}

	// 获取 rich_context
	richContext, ok := tableData["rich_context"].(map[string]interface{})
	if !ok {
		return fmt.Sprintf("⚠️  No rich_context in table '%s'. Continue with SQL generation.", updateInput.TableName), nil
	}

	// 检查 note 是否存在 - 如果不存在，创建新的
	if _, exists := richContext[updateInput.NoteKey]; !exists {
		return fmt.Sprintf("⚠️  Note key '%s' not found in table '%s'.\nTip: This might be a new insight. You can continue with SQL generation based on your findings.", updateInput.NoteKey, updateInput.TableName), nil
	}

	// 更新 note
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	richContext[updateInput.NoteKey] = map[string]string{
		"content":    updateInput.NewContent,
		"expires_at": expiresAt,
	}

	// 写回文件
	output, err := json.MarshalIndent(rawData, "", "  ")
	if err != nil {
		return fmt.Sprintf("⚠️  Failed to marshal context: %v\nContinue with SQL generation.", err), nil
	}

	if err := os.WriteFile(t.contextPath, output, 0644); err != nil {
		return fmt.Sprintf("⚠️  Failed to write context file: %v\nContinue with SQL generation.", err), nil
	}

	// 返回成功信息
	result := fmt.Sprintf(
		"✓ Rich Context updated successfully!\n"+
			"Table: %s\n"+
			"Note: %s\n"+
			"New Content: %s\n"+
			"Expires At: %s\n"+
			"Reason: %s",
		updateInput.TableName,
		updateInput.NoteKey,
		updateInput.NewContent,
		expiresAt,
		updateInput.Reason,
	)

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📝 Rich Context Updated:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return result, nil
}

// NewUpdateRichContextTool 创建更新工具
func NewUpdateRichContextTool(dbName, dbType string) *UpdateRichContextTool {
	// 构建 context 文件路径
	contextPath := filepath.Join("contexts", strings.ToLower(dbType), "spider", dbName+".json")

	return &UpdateRichContextTool{
		dbName:      dbName,
		contextPath: contextPath,
	}
}
