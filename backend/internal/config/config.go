package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level system configuration.
type Config struct {
	Server    ServerConfig     `yaml:"server" json:"server"`
	LLM       LLMConfig        `yaml:"llm" json:"llm"`
	Databases []DatabaseConfig `yaml:"databases" json:"databases"`
	React     ReactConfig      `yaml:"react" json:"react"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Mode     string `yaml:"mode" json:"mode"`           // debug | release
	LogLevel string `yaml:"log_level" json:"log_level"` // debug | info | warn | error
}

// LLMConfig holds LLM provider settings.
type LLMConfig struct {
	ConfigFile   string `yaml:"config_file" json:"config_file"`
	DefaultModel string `yaml:"default_model" json:"default_model"`
}

// DatabaseConfig holds connection details for a target database.
type DatabaseConfig struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"` // mysql, mariadb
	Host        string `yaml:"host" json:"host"`
	Port        int    `yaml:"port" json:"port"`
	User        string `yaml:"user" json:"user"`
	Password    string `yaml:"password" json:"password"`
	Database    string `yaml:"database" json:"database"`
	Description string `yaml:"description" json:"description"`
}

// ReactConfig holds ReAct agent settings.
type ReactConfig struct {
	MaxIterations    int  `yaml:"max_iterations" json:"max_iterations"`
	EnableRichContext bool `yaml:"enable_rich_context" json:"enable_rich_context"`
}

// Load reads and parses a config file (YAML or JSON).
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	
	// Choose parser based on file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Apply environment variable overrides
	applyEnvOverrides(config)

	// Set default values
	setDefaults(config)

	return config, nil
}

// applyEnvOverrides overrides config values from environment variables.
func applyEnvOverrides(config *Config) {
	// LLM default model
	if model := os.Getenv("LLM_DEFAULT_MODEL"); model != "" {
		config.LLM.DefaultModel = model
	}

	// Server port
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		fmt.Sscanf(port, "%d", &p)
		if p > 0 {
			config.Server.Port = p
		}
	}

	// Server mode
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// Log level
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Server.LogLevel = logLevel
	}

	// Database password overrides: fill empty passwords from env vars
	rootPass := os.Getenv("MARIADB_ROOT_PASSWORD")
	userPass := os.Getenv("MARIADB_PASSWORD")
	for i := range config.Databases {
		if config.Databases[i].Password == "" {
			if config.Databases[i].User == "root" && rootPass != "" {
				config.Databases[i].Password = rootPass
			} else if userPass != "" {
				config.Databases[i].Password = userPass
			}
		}
	}
}

// setDefaults fills in zero-value fields with sensible defaults.
func setDefaults(config *Config) {
	if config.Server.Port == 0 {
		config.Server.Port = 8081
	}
	if config.Server.Mode == "" {
		config.Server.Mode = "debug"
	}
	if config.Server.LogLevel == "" {
		config.Server.LogLevel = "debug"
	}
	if config.React.MaxIterations == 0 {
		config.React.MaxIterations = 5
	}
	if config.LLM.DefaultModel == "" {
		config.LLM.DefaultModel = "deepseek_v3"
	}
}
