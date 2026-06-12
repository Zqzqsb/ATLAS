package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// DefaultModelFirstInJSON is the sentinel for system.yaml default_model:
// use the first LLM entry in llm_config.json (JSON key order).
const DefaultModelFirstInJSON = "first_in_json"

// ModelConfig holds configuration for a single LLM model endpoint.
type ModelConfig struct {
	ModelName string `json:"model_name"`
	Token     string `json:"token"`
	BaseURL   string `json:"base_url"`
}

// EmbeddingConfig holds embedding API configuration within llm_config.json.
// Placed under the "_embedding" key so all API keys live in one file.
type EmbeddingConfig struct {
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	Dimension int    `json:"dimension"`
}

// Config holds all available model configurations, keyed by logical name
// (e.g. "deepseek_v3", "qwen_max"), plus optional _embedding section.
type Config struct {
	Models       map[string]ModelConfig
	Embedding    *EmbeddingConfig // from "_embedding" key, may be nil
	DefaultModel string          // logical name of the default model
}

// LoadConfig loads LLM configuration from a JSON file.
// The file format is: { "model_key": { "model_name": "...", "token": "...", "base_url": "..." }, ... }
// An optional "_embedding" key provides embedding API config in the same file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("llm: failed to read config %s: %w", path, err)
	}
	return parseConfigData(data, path)
}

func parseConfigData(data []byte, label string) (*Config, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("llm: failed to parse config %s: %w", label, err)
	}

	// Extract _embedding if present
	var emb *EmbeddingConfig
	if embRaw, ok := raw["_embedding"]; ok {
		emb = &EmbeddingConfig{}
		if err := json.Unmarshal(embRaw, emb); err != nil {
			return nil, fmt.Errorf("llm: failed to parse _embedding in %s: %w", label, err)
		}
		delete(raw, "_embedding")
	}

	// Parse remaining keys as model configs
	models := make(map[string]ModelConfig)
	for k, v := range raw {
		var mc ModelConfig
		if err := json.Unmarshal(v, &mc); err != nil {
			continue // skip keys that aren't valid model configs (e.g. _embedding already removed)
		}
		models[k] = mc
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("llm: config %s contains no models", label)
	}

	orderedKeys, err := orderedTopLevelKeys(data)
	if err != nil {
		return nil, fmt.Errorf("llm: failed to read model key order in %s: %w", label, err)
	}
	defaultKey := firstModelKey(orderedKeys, models)
	if defaultKey == "" {
		return nil, fmt.Errorf("llm: config %s contains no models", label)
	}

	return &Config{
		Models:       models,
		Embedding:    emb,
		DefaultModel: defaultKey,
	}, nil
}

func orderedTopLevelKeys(data []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("expected JSON object at root")
	}

	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if !ok {
			return nil, fmt.Errorf("expected string object key")
		}
		keys = append(keys, key)
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, err
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return keys, nil
}

func firstModelKey(ordered []string, models map[string]ModelConfig) string {
	for _, k := range ordered {
		if k == "_embedding" {
			continue
		}
		if _, ok := models[k]; ok {
			return k
		}
	}
	for k := range models {
		return k
	}
	return ""
}

// ResolveDefaultModel picks the runtime default model key.
// preferred may be empty, DefaultModelFirstInJSON, or an explicit llm_config.json key.
// The second return value is true when preferred was explicit but missing from config.
func ResolveDefaultModel(preferred string, cfg *Config) (string, bool) {
	if preferred == "" || preferred == DefaultModelFirstInJSON {
		return cfg.DefaultModel, false
	}
	if _, err := cfg.GetModel(preferred); err != nil {
		return cfg.DefaultModel, true
	}
	return preferred, false
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
