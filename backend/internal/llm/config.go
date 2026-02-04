package llm

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ModelConfig LLM model configuration
type ModelConfig struct {
	ModelName string `json:"model_name"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url"`
}

// ConfigFile configuration file structure
type ConfigFile struct {
	DeepSeekV3  ModelConfig `json:"deepseek_v3"`
	DeepSeekV32 ModelConfig `json:"deepseek_v3_2"`
	QwenMax     ModelConfig `json:"qwen_max"`
	Qwen3Max    ModelConfig `json:"qwen3_max"`
	AliDeepSeek ModelConfig `json:"ali_deepseek_v3_2"`
}

var (
	// Global config (loaded from file or env)
	config        *ConfigFile
	configLoaded  bool
	ErrNoLLMConfig = errors.New("LLM config not available")
)

func init() {
	// Try to load config file (optional)
	var err error
	config, err = loadConfig()
	if err != nil {
		log.Printf("[LLM] Config not found: %v. LLM features disabled until configured.", err)
		config = nil
		configLoaded = false
	} else {
		configLoaded = true
		log.Printf("[LLM] Config loaded successfully")
	}
}

// IsConfigured returns true if LLM config is available
func IsConfigured() bool {
	return configLoaded && config != nil
}

// loadConfig loads config from file or environment
func loadConfig() (*ConfigFile, error) {
	// Check environment variable first
	paths := []string{}
	if envPath := os.Getenv("LLM_CONFIG_PATH"); envPath != "" {
		paths = append(paths, envPath)
	}
	// Then try common paths
	paths = append(paths,
		"llm_config.json",
		"../llm_config.json",
		"../../llm_config.json",
		"../../../llm_config.json",
		"../../../../llm_config.json",
	)

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

// GetConfig returns config or nil if not configured
func GetConfig() *ConfigFile {
	return config
}

// GetConfigOrError returns config or error if not configured
func GetConfigOrError() (*ConfigFile, error) {
	if !IsConfigured() {
		return nil, ErrNoLLMConfig
	}
	return config, nil
}

// GetModel returns model config by flag, or empty config if not configured
func GetModel(useV32 bool) ModelConfig {
	cfg := GetConfig()
	if cfg == nil {
		return ModelConfig{}
	}
	if useV32 {
		return cfg.DeepSeekV32
	}
	return cfg.DeepSeekV3
}

// GetModelName returns model display name
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

// CreateLLMWithFlag creates LLM instance by flag, returns error if not configured
func CreateLLMWithFlag(useV32 bool) (llms.Model, error) {
	if !IsConfigured() {
		return nil, ErrNoLLMConfig
	}
	modelConfig := GetModel(useV32)
	return CreateLLM(modelConfig)
}

// ModelType model type enum
type ModelType string

const (
	ModelDeepSeekV3     ModelType = "deepseek-v3"
	ModelDeepSeekV32    ModelType = "deepseek-v3.2"
	ModelQwenMax        ModelType = "qwen-max"
	ModelQwen3Max       ModelType = "qwen3-max"
	ModelAliDeepSeekV32 ModelType = "ali-deepseek-v3.2"
)

// GetModelByType returns model config by type, or empty config if not configured
func GetModelByType(modelType ModelType) ModelConfig {
	cfg := GetConfig()
	if cfg == nil {
		return ModelConfig{}
	}
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

// GetModelDisplayName returns display name for model type
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

// CreateLLMByType creates LLM instance by type, returns error if not configured
func CreateLLMByType(modelType ModelType) (llms.Model, error) {
	if !IsConfigured() {
		return nil, ErrNoLLMConfig
	}
	modelConfig := GetModelByType(modelType)
	return CreateLLM(modelConfig)
}
