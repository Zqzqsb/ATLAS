// Package bridge 提供 internal 实现与 interfaces 之间的桥接
package bridge

import (
	"context"

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

// AdapterFactory 创建适配器工厂函数
// 返回一个工厂函数，接受 interfaces.DBConfig 并返回 interfaces.DBAdapter
func AdapterFactory() interfaces.AdapterFactory {
	return func(cfg *interfaces.DBConfig) (interfaces.DBAdapter, error) {
		// 转换配置
		internalCfg := &adapter.DBConfig{
			Type:     cfg.Type,
			Host:     cfg.Host,
			Port:     cfg.Port,
			Database: cfg.Database,
			User:     cfg.User,
			Password: cfg.Password,
			FilePath: cfg.FilePath,
		}

		// 创建内部适配器
		internalAdapter, err := adapter.NewAdapter(internalCfg)
		if err != nil {
			return nil, err
		}

		// 包装为接口适配器
		return NewAdapterBridge(internalAdapter), nil
	}
}
