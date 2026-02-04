package context

import (
	"fmt"
	"strings"
)

// getString 从 map 中安全获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// getInt 从 map 中安全获取整数值
func getInt(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// getInt64 从 map 中安全获取 int64 值
func getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

// getBool 从 map 中安全获取布尔值
func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// formatKey 格式化 key 为更友好的显示
// 将 snake_case 转换为 Title Case
// 例如: "user_name" -> "User Name"
func formatKey(key string) string {
	words := strings.Split(key, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// normalizeTableName 标准化表名（转小写）
func normalizeTableName(name string) string {
	return strings.ToLower(name)
}

// normalizeColumnName 标准化列名（转小写）
func normalizeColumnName(name string) string {
	return strings.ToLower(name)
}
