package services

import (
	"context"
	"fmt"
	"sync"

	"atlas/internal/adapter"
	"atlas/internal/config"
)

// DatabaseService manages multiple database connections
type DatabaseService struct {
	config         *config.Config
	adapters       map[string]adapter.DBAdapter
	adapterFactory adapter.AdapterFactory
	mu             sync.RWMutex
}

// NewDatabaseService creates a new database service
func NewDatabaseService(cfg *config.Config, factory adapter.AdapterFactory) *DatabaseService {
	return &DatabaseService{
		config:         cfg,
		adapters:       make(map[string]adapter.DBAdapter),
		adapterFactory: factory,
	}
}

// GetAdapter returns adapter for a specific database ID
func (s *DatabaseService) GetAdapter(dbID string) (adapter.DBAdapter, error) {
	s.mu.RLock()
	if a, ok := s.adapters[dbID]; ok {
		s.mu.RUnlock()
		return a, nil
	}
	s.mu.RUnlock()

	// Create new adapter
	return s.createAdapter(dbID)
}

// createAdapter creates a new database adapter
func (s *DatabaseService) createAdapter(dbID string) (adapter.DBAdapter, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check after acquiring lock
	if a, ok := s.adapters[dbID]; ok {
		return a, nil
	}

	// Find database config
	var dbConfig *config.DatabaseConfig
	for _, db := range s.config.Databases {
		if db.ID == dbID {
			dbConfig = &db
			break
		}
	}

	if dbConfig == nil {
		return nil, fmt.Errorf("database not found: %s", dbID)
	}

	// Create adapter using factory
	adapterCfg := &adapter.DBConfig{
		Type:     dbConfig.Type,
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		Database: dbConfig.Database,
		User:     dbConfig.User,
		Password: dbConfig.Password,
	}

	adp, err := s.adapterFactory(adapterCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Connect
	if err := adp.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database %s: %w", dbID, err)
	}

	s.adapters[dbID] = adp
	return adp, nil
}

// CreateCustomAdapter creates an adapter from custom configuration (not stored in config)
func (s *DatabaseService) CreateCustomAdapter(cfg *AdapterConfig) (adapter.DBAdapter, error) {
	dbConfig := &adapter.DBConfig{
		Type:     cfg.Type,
		Host:     cfg.Host,
		Port:     cfg.Port,
		Database: cfg.Database,
		User:     cfg.User,
		Password: cfg.Password,
	}

	return s.adapterFactory(dbConfig)
}

// GetSchema returns schema information for a database
func (s *DatabaseService) GetSchema(ctx context.Context, dbID string) (*SchemaInfo, error) {
	adp, err := s.GetAdapter(dbID)
	if err != nil {
		return nil, err
	}

	schema := &SchemaInfo{
		DatabaseID:   dbID,
		DatabaseType: adp.GetDatabaseType(),
		Tables:       []TableSchema{},
	}

	// Get all tables
	tables, err := s.getTables(ctx, adp)
	if err != nil {
		return nil, err
	}

	for _, tableName := range tables {
		tableSchema, err := s.getTableSchema(ctx, adp, tableName)
		if err != nil {
			continue // Skip tables with errors
		}
		schema.Tables = append(schema.Tables, *tableSchema)
	}

	return schema, nil
}

// getTables returns list of all tables in the database
func (s *DatabaseService) getTables(ctx context.Context, adp adapter.DBAdapter) ([]string, error) {
	query := "SHOW TABLES"

	result, err := adp.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	var tables []string
	for _, row := range result.Rows {
		for _, val := range row {
			if name, ok := val.(string); ok {
				tables = append(tables, name)
				break
			}
		}
	}

	return tables, nil
}

// getTableSchema returns schema for a single table
func (s *DatabaseService) getTableSchema(ctx context.Context, adp adapter.DBAdapter, tableName string) (*TableSchema, error) {
	schema := &TableSchema{
		Name:    tableName,
		Columns: []ColumnInfo{},
	}

	query := fmt.Sprintf("DESCRIBE %s", tableName)

	result, err := adp.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, row := range result.Rows {
		col := parseColumnInfo(row, adp.GetDatabaseType())
		schema.Columns = append(schema.Columns, col)
	}

	// Get row count
	countResult, err := adp.ExecuteQuery(ctx, fmt.Sprintf("SELECT COUNT(*) as cnt FROM %s", tableName))
	if err == nil && len(countResult.Rows) > 0 {
		for _, val := range countResult.Rows[0] {
			switch v := val.(type) {
			case int64:
				schema.RowCount = v
			case float64:
				schema.RowCount = int64(v)
			}
			break
		}
	}

	return schema, nil
}

// parseColumnInfo parses column info from query result
func parseColumnInfo(row map[string]interface{}, dbType string) ColumnInfo {
	col := ColumnInfo{}
	col.Name = getString(row, "Field")
	col.Type = getString(row, "Type")
	col.Nullable = getString(row, "Null") == "YES"
	col.IsPrimaryKey = getString(row, "Key") == "PRI"
	return col
}

// ExecuteSQL executes a SQL query
func (s *DatabaseService) ExecuteSQL(ctx context.Context, dbID, sql string) (*adapter.QueryResult, error) {
	adp, err := s.GetAdapter(dbID)
	if err != nil {
		return nil, err
	}

	return adp.ExecuteQuery(ctx, sql)
}

// Close closes all database connections
func (s *DatabaseService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, adp := range s.adapters {
		adp.Close()
	}
	s.adapters = make(map[string]adapter.DBAdapter)
}

// CloseAdapter closes a specific database adapter
func (s *DatabaseService) CloseAdapter(dbID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if adp, ok := s.adapters[dbID]; ok {
		adp.Close()
		delete(s.adapters, dbID)
	}
}

// AddDatabase adds a database configuration and validates connectivity.
// Returns an error if the ID already exists or the connection fails.
func (s *DatabaseService) AddDatabase(db config.DatabaseConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.config.Databases {
		if existing.ID == db.ID {
			return fmt.Errorf("connection ID already exists: %s", db.ID)
		}
	}

	// Validate connectivity before committing
	adapterCfg := &adapter.DBConfig{
		Type:     db.Type,
		Host:     db.Host,
		Port:     db.Port,
		Database: db.Database,
		User:     db.User,
		Password: db.Password,
	}
	adp, err := s.adapterFactory(adapterCfg)
	if err != nil {
		return fmt.Errorf("failed to create adapter: %w", err)
	}
	if err := adp.Connect(context.Background()); err != nil {
		adp.Close()
		return fmt.Errorf("failed to connect: %w", err)
	}
	s.adapters[db.ID] = adp

	s.config.Databases = append(s.config.Databases, db)
	return nil
}

// RemoveDatabase removes a database configuration and closes its adapter.
// Returns false if the ID was not found.
func (s *DatabaseService) RemoveDatabase(dbID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	newDatabases := make([]config.DatabaseConfig, 0, len(s.config.Databases))
	for _, db := range s.config.Databases {
		if db.ID == dbID {
			found = true
			if adp, ok := s.adapters[dbID]; ok {
				adp.Close()
				delete(s.adapters, dbID)
			}
		} else {
			newDatabases = append(newDatabases, db)
		}
	}
	if found {
		s.config.Databases = newDatabases
	}
	return found
}

// FindDatabase returns the database config for the given ID, or nil.
func (s *DatabaseService) FindDatabase(dbID string) *config.DatabaseConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, db := range s.config.Databases {
		if db.ID == dbID {
			cpy := db
			return &cpy
		}
	}
	return nil
}

// ListDatabases returns a copy of all database configurations.
func (s *DatabaseService) ListDatabases() []config.DatabaseConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]config.DatabaseConfig, len(s.config.Databases))
	copy(out, s.config.Databases)
	return out
}

// NewIsolatedAdapter creates a new, independent adapter for the given database ID.
// Unlike GetAdapter, the returned adapter is NOT cached and must be closed by the caller.
func (s *DatabaseService) NewIsolatedAdapter(dbID string) (adapter.DBAdapter, error) {
	dbCfg := s.FindDatabase(dbID)
	if dbCfg == nil {
		return nil, fmt.Errorf("database config not found: %s", dbID)
	}
	return s.adapterFactory(&adapter.DBConfig{
		Type:     dbCfg.Type,
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		Database: dbCfg.Database,
		User:     dbCfg.User,
		Password: dbCfg.Password,
	})
}

// AdapterConfig represents configuration for creating a custom adapter
type AdapterConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int64:
			return n
		case float64:
			return int64(n)
		case int:
			return int64(n)
		}
	}
	return 0
}

// SchemaInfo represents database schema
type SchemaInfo struct {
	DatabaseID   string        `json:"database_id"`
	DatabaseType string        `json:"database_type"`
	Tables       []TableSchema `json:"tables"`
}

// TableSchema represents table schema
type TableSchema struct {
	Name     string       `json:"name"`
	RowCount int64        `json:"row_count"`
	Columns  []ColumnInfo `json:"columns"`
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	IsPrimaryKey bool   `json:"is_primary_key,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
}

