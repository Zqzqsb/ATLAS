// Package bridge provides adapters between internal implementations and interfaces
package bridge

import (
	"context"
	"fmt"

	"lucid/interfaces"
	"lucid/internal/adapter"
)

// AdapterBridge 将内部 adapter.DBAdapter 适配为 interfaces.DBAdapter
type AdapterBridge struct {
	internal adapter.DBAdapter
}

// NewAdapterBridge 创建适配器桥接
func NewAdapterBridge(internal adapter.DBAdapter) *AdapterBridge {
	return &AdapterBridge{internal: internal}
}

// Connect 连接数据库
func (b *AdapterBridge) Connect(ctx context.Context) error {
	return b.internal.Connect(ctx)
}

// Close 关闭连接
func (b *AdapterBridge) Close() error {
	return b.internal.Close()
}

// ExecuteQuery 执行查询
func (b *AdapterBridge) ExecuteQuery(ctx context.Context, query string) (*interfaces.QueryResult, error) {
	result, err := b.internal.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	// 转换结果
	return &interfaces.QueryResult{
		Columns:       result.Columns,
		Rows:          result.Rows,
		RowCount:      result.RowCount,
		ExecutionTime: result.ExecutionTime,
		Error:         result.Error,
	}, nil
}

// GetDatabaseType 获取数据库类型
func (b *AdapterBridge) GetDatabaseType() string {
	return b.internal.GetDatabaseType()
}

// GetDatabaseVersion 获取数据库版本
func (b *AdapterBridge) GetDatabaseVersion(ctx context.Context) (string, error) {
	return b.internal.GetDatabaseVersion(ctx)
}

// DryRunSQL 验证 SQL 语法
func (b *AdapterBridge) DryRunSQL(ctx context.Context, sql string) error {
	return b.internal.DryRunSQL(ctx, sql)
}

// AdapterFactory manages database configurations and creates adapters
type AdapterFactory struct {
	configs map[string]*interfaces.DBConfig
}

// NewAdapterFactory creates a new adapter factory
func NewAdapterFactory() *AdapterFactory {
	return &AdapterFactory{
		configs: make(map[string]*interfaces.DBConfig),
	}
}

// RegisterConfig registers a database configuration
func (f *AdapterFactory) RegisterConfig(id string, cfg *interfaces.DBConfig) {
	f.configs[id] = cfg
}

// GetConfig returns a database configuration by ID
func (f *AdapterFactory) GetConfig(id string) (*interfaces.DBConfig, error) {
	cfg, ok := f.configs[id]
	if !ok {
		return nil, fmt.Errorf("database config not found: %s", id)
	}
	return cfg, nil
}

// Create creates a new database adapter from config
func (f *AdapterFactory) Create(cfg *interfaces.DBConfig) (interfaces.DBAdapter, error) {
	// Convert to internal config
	internalCfg := &adapter.DBConfig{
		Type:     cfg.Type,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.Database,
		User:     cfg.User,
		Password: cfg.Password,
		FilePath: cfg.FilePath,
	}

	// Create internal adapter
	internalAdapter, err := adapter.NewAdapter(internalCfg)
	if err != nil {
		return nil, err
	}

	// Wrap as interface adapter
	return NewAdapterBridge(internalAdapter), nil
}

// CreateByID creates adapter using registered config
func (f *AdapterFactory) CreateByID(id string) (interfaces.DBAdapter, error) {
	cfg, err := f.GetConfig(id)
	if err != nil {
		return nil, err
	}
	return f.Create(cfg)
}
