package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 系统配置
type Config struct {
	Server    ServerConfig     `yaml:"server" json:"server"`
	LLM       LLMConfig        `yaml:"llm" json:"llm"`
	Databases []DatabaseConfig `yaml:"databases" json:"databases"`
	React     ReactConfig      `yaml:"react" json:"react"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Mode     string `yaml:"mode" json:"mode"`           // debug | release
	LogLevel string `yaml:"log_level" json:"log_level"` // debug | info | warn | error
}

// LLMConfig LLM配置
type LLMConfig struct {
	ConfigFile   string `yaml:"config_file" json:"config_file"`
	DefaultModel string `yaml:"default_model" json:"default_model"`
}

// DatabaseConfig 数据库连接配置
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

// ReactConfig ReAct配置
type ReactConfig struct {
	MaxIterations    int  `yaml:"max_iterations" json:"max_iterations"`
	EnableRichContext bool `yaml:"enable_rich_context" json:"enable_rich_context"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	
	// 根据扩展名选择解析方式
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

	// 应用环境变量覆盖
	applyEnvOverrides(config)

	// 设置默认值
	setDefaults(config)

	return config, nil
}

// applyEnvOverrides 应用环境变量覆盖配置
func applyEnvOverrides(config *Config) {
	// LLM 默认模型
	if model := os.Getenv("LLM_DEFAULT_MODEL"); model != "" {
		config.LLM.DefaultModel = model
	}

	// 服务器端口
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		fmt.Sscanf(port, "%d", &p)
		if p > 0 {
			config.Server.Port = p
		}
	}

	// 服务器模式
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// 日志级别
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Server.LogLevel = logLevel
	}
}

// setDefaults 设置默认值
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
