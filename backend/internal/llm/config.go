package llm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// ModelConfig holds configuration for a single LLM model endpoint.
type ModelConfig struct {
	ModelName string `json:"model_name"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url"`
}

// Config holds all available model configurations, keyed by logical name
// (e.g. "deepseek_v3", "qwen_max").
// The JSON file is simply map[string]ModelConfig.
type Config struct {
	Models       map[string]ModelConfig
	DefaultModel string // logical name of the default model
}

// LoadConfig loads LLM configuration from a JSON file.
// The file format is: { "model_key": { "model_name": "...", "token": "...", "base_url": "..." }, ... }
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("llm: failed to read config %s: %w", path, err)
	}

	models := make(map[string]ModelConfig)
	if err := json.Unmarshal(data, &models); err != nil {
		return nil, fmt.Errorf("llm: failed to parse config %s: %w", path, err)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("llm: config %s contains no models", path)
	}

	// Pick first key as default (caller can override)
	var defaultKey string
	for k := range models {
		defaultKey = k
		break
	}

	return &Config{
		Models:       models,
		DefaultModel: defaultKey,
	}, nil
}

// FindConfig searches common paths for the LLM config file.
// Tries: LLM_CONFIG_PATH env, then relative paths.
func FindConfig() (*Config, error) {
	var paths []string
	if envPath := os.Getenv("LLM_CONFIG_PATH"); envPath != "" {
		paths = append(paths, envPath)
	}
	paths = append(paths,
		"llm_config.json",
		"../llm_config.json",
		"../../llm_config.json",
		"../../../llm_config.json",
	)

	var lastErr error
	for _, p := range paths {
		cfg, err := LoadConfig(p)
		if err != nil {
			lastErr = err
			continue
		}
		return cfg, nil
	}
	return nil, fmt.Errorf("llm: no config found (last error: %w)", lastErr)
}

// GetModel returns the ModelConfig for the given key, or an error if not found.
func (c *Config) GetModel(key string) (ModelConfig, error) {
	m, ok := c.Models[key]
	if !ok {
		return ModelConfig{}, fmt.Errorf("llm: model %q not found in config", key)
	}
	return m, nil
}

// GetDefaultModel returns the default ModelConfig.
func (c *Config) GetDefaultModel() (ModelConfig, error) {
	return c.GetModel(c.DefaultModel)
}

// ListModels returns all available model keys.
func (c *Config) ListModels() []string {
	keys := make([]string, 0, len(c.Models))
	for k := range c.Models {
		keys = append(keys, k)
	}
	return keys
}

// CreateLLM creates a langchaingo llms.Model from a ModelConfig.
func CreateLLM(mc ModelConfig) (llms.Model, error) {
	return openai.New(
		openai.WithModel(mc.ModelName),
		openai.WithToken(mc.Token),
		openai.WithBaseURL(mc.BaseURL),
	)
}

// CreateLLMByKey creates a langchaingo llms.Model by config key.
func (c *Config) CreateLLMByKey(key string) (llms.Model, error) {
	mc, err := c.GetModel(key)
	if err != nil {
		return nil, err
	}
	return CreateLLM(mc)
}

// CreateDefaultLLM creates a langchaingo llms.Model using the default model.
func (c *Config) CreateDefaultLLM() (llms.Model, error) {
	return c.CreateLLMByKey(c.DefaultModel)
}
