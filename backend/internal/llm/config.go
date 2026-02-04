package llm

import (
	"encoding/json"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ModelConfig LLM 模型配置
type ModelConfig struct {
	ModelName string `json:"model_name"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url"`
}

// ConfigFile 配置文件结构
type ConfigFile struct {
	DeepSeekV3  ModelConfig `json:"deepseek_v3"`
	DeepSeekV32 ModelConfig `json:"deepseek_v3_2"`
	QwenMax     ModelConfig `json:"qwen_max"`
	Qwen3Max    ModelConfig `json:"qwen3_max"`
	AliDeepSeek ModelConfig `json:"ali_deepseek_v3_2"`
}

var (
	// 全局配置（从文件加载）
	config *ConfigFile
)

func init() {
	// 尝试加载配置文件
	var err error
	config, err = loadConfig()
	if err != nil {
		panic("Failed to load llm_config.json: " + err.Error() + ". Please create llm_config.json in the project root.")
	}
}

// loadConfig 加载配置文件
func loadConfig() (*ConfigFile, error) {
	// 尝试多个可能的配置文件路径
	paths := []string{
		"llm_config.json",
		"../llm_config.json",
		"../../llm_config.json",
		"../../../llm_config.json",
		"../../../../llm_config.json",
	}

	var lastErr error
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}

		var cfg ConfigFile
		if err := json.Unmarshal(data, &cfg); err != nil {
			lastErr = err
			continue
		}

		return &cfg, nil
	}

	// 如果找不到配置文件，返回错误
	return nil, lastErr
}

// GetConfig 获取当前配置
func GetConfig() *ConfigFile {
	if config == nil {
		panic("LLM config not initialized. Please ensure llm_config.json exists.")
	}
	return config
}

// GetModel 根据标志获取模型配置
func GetModel(useV32 bool) ModelConfig {
	cfg := GetConfig()
	if useV32 {
		return cfg.DeepSeekV32
	}
	return cfg.DeepSeekV3
}

// GetModelName 获取模型显示名称
func GetModelName(useV32 bool) string {
	if useV32 {
		return "DeepSeek-V3.2"
	}
	return "DeepSeek-V3"
}

// CreateLLM 创建 LLM 实例
func CreateLLM(config ModelConfig) (llms.Model, error) {
	return openai.New(
		openai.WithModel(config.ModelName),
		openai.WithToken(config.Token),
		openai.WithBaseURL(config.BaseURL),
	)
}

// CreateLLMWithFlag 根据标志创建 LLM 实例
func CreateLLMWithFlag(useV32 bool) (llms.Model, error) {
	modelConfig := GetModel(useV32)
	return CreateLLM(modelConfig)
}

// ModelType 模型类型
type ModelType string

const (
	ModelDeepSeekV3     ModelType = "deepseek-v3"
	ModelDeepSeekV32    ModelType = "deepseek-v3.2"
	ModelQwenMax        ModelType = "qwen-max"
	ModelQwen3Max       ModelType = "qwen3-max"
	ModelAliDeepSeekV32 ModelType = "ali-deepseek-v3.2"
)

// GetModelByType 根据模型类型获取配置
func GetModelByType(modelType ModelType) ModelConfig {
	cfg := GetConfig()
	switch modelType {
	case ModelDeepSeekV3:
		return cfg.DeepSeekV3
	case ModelDeepSeekV32:
		return cfg.DeepSeekV32
	case ModelQwenMax:
		return cfg.QwenMax
	case ModelQwen3Max:
		return cfg.Qwen3Max
	case ModelAliDeepSeekV32:
		return cfg.AliDeepSeek
	default:
		return cfg.DeepSeekV3
	}
}

// GetModelDisplayName 获取模型显示名称
func GetModelDisplayName(modelType ModelType) string {
	switch modelType {
	case ModelDeepSeekV3:
		return "DeepSeek-V3 (Volcano)"
	case ModelDeepSeekV32:
		return "DeepSeek-V3.2 (Volcano)"
	case ModelQwenMax:
		return "Qwen-Max (Aliyun)"
	case ModelQwen3Max:
		return "Qwen3-Max (Aliyun)"
	case ModelAliDeepSeekV32:
		return "DeepSeek-V3.2 (Aliyun)"
	default:
		return "Unknown"
	}
}

// CreateLLMByType 根据模型类型创建 LLM 实例
func CreateLLMByType(modelType ModelType) (llms.Model, error) {
	modelConfig := GetModelByType(modelType)
	return CreateLLM(modelConfig)
}
