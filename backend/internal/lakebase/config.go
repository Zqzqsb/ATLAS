// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in LUCID system.
package lakebase

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LakebaseConfig holds the complete lake-base configuration
type LakebaseConfig struct {
	Lakebase     LakebaseDBConfig     `yaml:"lakebase"`
	Embedding    EmbeddingConfig      `yaml:"embedding"`
	VectorSearch VectorSearchConfig   `yaml:"vector_search"`
	Agent        AgentConfig          `yaml:"agent"`
}

// LakebaseDBConfig holds database connection settings
type LakebaseDBConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// EmbeddingConfig holds embedding generation settings
type EmbeddingConfig struct {
	Enabled    bool   `yaml:"enabled"`
	APIKey     string `yaml:"api_key"`
	BaseURL    string `yaml:"base_url"`
	Model      string `yaml:"model"`
	Dimension  int    `yaml:"dimension"`
	BatchSize  int    `yaml:"batch_size"`
	Multimodal bool   `yaml:"multimodal"` // use /embeddings/multimodal endpoint (Volcengine)
}

// VectorSearchConfig holds vector search settings
type VectorSearchConfig struct {
	TopK        int     `yaml:"top_k"`
	MinScore    float64 `yaml:"min_score"`
	MaxDistance float64 `yaml:"max_distance"`
}

// AgentConfig holds agent maintenance settings
type AgentConfig struct {
	EnableDDLDetection bool `yaml:"enable_ddl_detection"`
	CheckInterval      int  `yaml:"check_interval"`
	AutoRefreshContext bool `yaml:"auto_refresh_context"`
}

// LoadConfig loads lakebase configuration from a YAML file
func LoadConfig(path string) (*LakebaseConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &LakebaseConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Apply defaults
	config.applyDefaults()

	// Override with environment variables
	config.overrideFromEnv()

	return config, nil
}

// applyDefaults sets default values for missing configurations
func (c *LakebaseConfig) applyDefaults() {
	// Database defaults
	if c.Lakebase.Host == "" {
		c.Lakebase.Host = "127.0.0.1"
	}
	if c.Lakebase.Port == 0 {
		c.Lakebase.Port = 3310
	}
	if c.Lakebase.User == "" {
		c.Lakebase.User = "root"
	}
	if c.Lakebase.Database == "" {
		c.Lakebase.Database = "lucid"
	}
	if c.Lakebase.MaxOpenConns == 0 {
		c.Lakebase.MaxOpenConns = 20
	}
	if c.Lakebase.MaxIdleConns == 0 {
		c.Lakebase.MaxIdleConns = 10
	}
	if c.Lakebase.ConnMaxLifetime == 0 {
		c.Lakebase.ConnMaxLifetime = 300 * time.Second
	}

	// Embedding defaults
	if c.Embedding.Model == "" {
		c.Embedding.Model = DefaultEmbeddingModel
	}
	if c.Embedding.Dimension == 0 {
		c.Embedding.Dimension = DefaultEmbeddingDimension
	}
	if c.Embedding.BatchSize == 0 {
		c.Embedding.BatchSize = 100
	}

	// Vector search defaults
	if c.VectorSearch.TopK == 0 {
		c.VectorSearch.TopK = 10
	}
	if c.VectorSearch.MinScore == 0 {
		c.VectorSearch.MinScore = 0.7
	}
	if c.VectorSearch.MaxDistance == 0 {
		c.VectorSearch.MaxDistance = 0.3
	}

	// Agent defaults
	if c.Agent.CheckInterval == 0 {
		c.Agent.CheckInterval = 60
	}
}

// overrideFromEnv overrides configuration from environment variables
func (c *LakebaseConfig) overrideFromEnv() {
	if v := os.Getenv("LAKEBASE_HOST"); v != "" {
		c.Lakebase.Host = v
	}
	if v := os.Getenv("LAKEBASE_PASSWORD"); v != "" {
		c.Lakebase.Password = v
	}
	// Also accept MARIADB_PASSWORD as fallback for lakebase password
	if c.Lakebase.Password == "" {
		if v := os.Getenv("MARIADB_PASSWORD"); v != "" {
			c.Lakebase.Password = v
		}
	}
	if v := os.Getenv("LAKEBASE_DATABASE"); v != "" {
		c.Lakebase.Database = v
	}
	// Embedding configuration from environment
	if v := os.Getenv("EMBEDDING_API_KEY"); v != "" {
		c.Embedding.APIKey = v
		c.Embedding.Enabled = true
	}
	// Resolve API key placeholder
	if c.Embedding.APIKey == "${EMBEDDING_API_KEY}" {
		c.Embedding.APIKey = os.Getenv("EMBEDDING_API_KEY")
	}
	if v := os.Getenv("EMBEDDING_BASE_URL"); v != "" {
		c.Embedding.BaseURL = v
	}
	if v := os.Getenv("EMBEDDING_MODEL"); v != "" {
		c.Embedding.Model = v
	}
}

// ToConnectionConfig converts LakebaseDBConfig to ConnectionPool Config
func (c *LakebaseDBConfig) ToConnectionConfig() *Config {
	return &Config{
		Host:            c.Host,
		Port:            c.Port,
		User:            c.User,
		Password:        c.Password,
		Database:        c.Database,
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

// DefaultLakebaseConfig returns the default configuration
func DefaultLakebaseConfig() *LakebaseConfig {
	config := &LakebaseConfig{
		Lakebase: LakebaseDBConfig{
			Host:            "127.0.0.1",
			Port:            3310,
			User:            "root",
			Password:        "your_strong_password",
			Database:        "lucid",
			MaxOpenConns:    20,
			MaxIdleConns:    10,
			ConnMaxLifetime: 300 * time.Second,
		},
		Embedding: EmbeddingConfig{
			Model:     DefaultEmbeddingModel,
			Dimension: DefaultEmbeddingDimension,
			BatchSize: 100,
		},
		VectorSearch: VectorSearchConfig{
			TopK:        10,
			MinScore:    0.7,
			MaxDistance: 0.3,
		},
		Agent: AgentConfig{
			EnableDDLDetection: true,
			CheckInterval:      60,
			AutoRefreshContext: true,
		},
	}
	return config
}
