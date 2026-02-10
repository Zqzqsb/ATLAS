package agent

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"lucid/internal/lakebase"
	"lucid/internal/logger"
)

// EvolutionStage represents a single schema evolution step
type EvolutionStage struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	DDLs        []string         `json:"ddls"`         // One or more DDL statements
	SampleData  []string         `json:"sample_data"`   // Optional INSERT statements
	ExpectedChanges []SchemaChangeType `json:"expected_changes"`
}

// EvolutionStatus tracks the current state of evolution demo
type EvolutionStatus struct {
	CurrentStage int              `json:"current_stage"`
	TotalStages  int              `json:"total_stages"`
	StageHistory []StageExecution `json:"stage_history"`
	DatabaseName string           `json:"database_name"`
	IsReady      bool             `json:"is_ready"`
}

// StageExecution records a stage execution result
type StageExecution struct {
	StageID      int              `json:"stage_id"`
	StageName    string           `json:"stage_name"`
	DDLExecuted  []string         `json:"ddl_executed"`
	Changes      []SchemaChange   `json:"changes_detected"`
	ContextActions []ContextAction `json:"context_actions"`
	ExecutedAt   time.Time        `json:"executed_at"`
	DurationMs   int64            `json:"duration_ms"`
	Success      bool             `json:"success"`
	Error        string           `json:"error,omitempty"`
}

// ContextAction describes what the maintainer did in response to a change
type ContextAction struct {
	ActionType  string `json:"action_type"`  // "created", "expired", "refreshed", "deleted"
	TableName   string `json:"table_name"`
	ColumnName  string `json:"column_name,omitempty"`
	ContextType string `json:"context_type,omitempty"`
	Description string `json:"description"`
	OldContent  string `json:"old_content,omitempty"`
	NewContent  string `json:"new_content,omitempty"`
}

// EvolutionEvent is emitted via SSE during stage execution
type EvolutionEvent struct {
	Type    string      `json:"type"`
	Phase   string      `json:"phase"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Predefined evolution stages for the demo database lucid_evolution
var DefaultEvolutionStages = []EvolutionStage{
	{
		ID:          1,
		Name:        "Add User Phone Column",
		Description: "业务需求：需要用户手机号用于联系",
		DDLs: []string{
			"ALTER TABLE users ADD COLUMN phone VARCHAR(20)",
		},
		SampleData: []string{
			"UPDATE users SET phone = '13800001111' WHERE id = 1",
			"UPDATE users SET phone = '13900002222' WHERE id = 2",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeColumnAdded},
	},
	{
		ID:          2,
		Name:        "Add Products Table",
		Description: "业务需求：引入商品系统",
		DDLs: []string{
			`CREATE TABLE products (
				id INT PRIMARY KEY AUTO_INCREMENT,
				name VARCHAR(200) NOT NULL,
				price DECIMAL(10,2) NOT NULL,
				category VARCHAR(50),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
		},
		SampleData: []string{
			"INSERT INTO products (name, price, category) VALUES ('iPhone 15', 7999.00, '电子产品')",
			"INSERT INTO products (name, price, category) VALUES ('MacBook Pro', 14999.00, '电子产品')",
			"INSERT INTO products (name, price, category) VALUES ('AirPods Pro', 1899.00, '配件')",
			"INSERT INTO products (name, price, category) VALUES ('iPad Air', 4799.00, '电子产品')",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeTableAdded},
	},
	{
		ID:          3,
		Name:        "Add Order-Product Foreign Key",
		Description: "业务需求：订单需要关联商品",
		DDLs: []string{
			"ALTER TABLE orders ADD COLUMN product_id INT",
			"ALTER TABLE orders ADD CONSTRAINT fk_order_product FOREIGN KEY (product_id) REFERENCES products(id)",
		},
		SampleData: []string{
			"UPDATE orders SET product_id = 1 WHERE id = 1",
			"UPDATE orders SET product_id = 2 WHERE id = 2",
			"UPDATE orders SET product_id = 3 WHERE id = 3",
			"UPDATE orders SET product_id = 4 WHERE id = 4",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeColumnAdded, ChangeTypeForeignKeyAdded},
	},
	{
		ID:          4,
		Name:        "Modify Amount Precision",
		Description: "业务需求：金额精度从 2 位升级到 4 位，支持更精细定价",
		DDLs: []string{
			"ALTER TABLE orders MODIFY COLUMN amount DECIMAL(15,4)",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeColumnModified},
	},
	{
		ID:          5,
		Name:        "Drop Email Column",
		Description: "业务调整：用户隐私合规，移除 email 字段",
		DDLs: []string{
			"ALTER TABLE users DROP COLUMN email",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeColumnDropped},
	},
}

// EvolutionService manages the evolution demo lifecycle
type EvolutionService struct {
	pool         *lakebase.ConnectionPool
	repo         *lakebase.MySQLRepository
	agentService *AgentService
	stages       []EvolutionStage
	status       *EvolutionStatus
	mu           sync.RWMutex
}

// NewEvolutionService creates a new evolution service
func NewEvolutionService(pool *lakebase.ConnectionPool, repo *lakebase.MySQLRepository, agentSvc *AgentService) *EvolutionService {
	return &EvolutionService{
		pool:         pool,
		repo:         repo,
		agentService: agentSvc,
		stages:       DefaultEvolutionStages,
		status: &EvolutionStatus{
			CurrentStage: 0,
			TotalStages:  len(DefaultEvolutionStages),
			StageHistory: []StageExecution{},
			DatabaseName: "lucid_evolution",
			IsReady:      false,
		},
	}
}

// GetStages returns all evolution stages
func (s *EvolutionService) GetStages() []EvolutionStage {
	return s.stages
}

// GetStatus returns current evolution status
func (s *EvolutionService) GetStatus() *EvolutionStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// ExecuteStage executes a specific stage with SSE event streaming
func (s *EvolutionService) ExecuteStage(ctx context.Context, dsID int64, stageID int, events chan<- EvolutionEvent) (*StageExecution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate stage
	if stageID < 1 || stageID > len(s.stages) {
		return nil, fmt.Errorf("invalid stage ID: %d (valid: 1-%d)", stageID, len(s.stages))
	}

	// Check ordering — must execute in sequence
	if stageID != s.status.CurrentStage+1 {
		return nil, fmt.Errorf("must execute stage %d next (requested: %d)", s.status.CurrentStage+1, stageID)
	}

	stage := s.stages[stageID-1]
	startTime := time.Now()

	execution := &StageExecution{
		StageID:     stageID,
		StageName:   stage.Name,
		DDLExecuted: stage.DDLs,
		ExecutedAt:  startTime,
	}

	emit := func(eventType, phase, message string, data interface{}) {
		if events != nil {
			select {
			case events <- EvolutionEvent{Type: eventType, Phase: phase, Message: message, Data: data}:
			case <-ctx.Done():
			}
		}
	}

	// Phase 1: Announce stage
	emit("stage_start", "announce", fmt.Sprintf("Stage %d: %s — %s", stageID, stage.Name, stage.Description), map[string]interface{}{
		"stage_id":   stageID,
		"stage_name": stage.Name,
		"ddls":       stage.DDLs,
	})

	// Phase 2: Execute DDL on business database
	emit("ddl_executing", "ddl", "Executing DDL statements on business database...", nil)

	businessDB, err := s.getBusinessDB(dsID)
	if err != nil {
		execution.Error = fmt.Sprintf("failed to get business DB: %v", err)
		execution.Success = false
		emit("error", "ddl", execution.Error, nil)
		return execution, err
	}

	for i, ddl := range stage.DDLs {
		emit("ddl_executing", "ddl", fmt.Sprintf("Executing DDL %d/%d: %s", i+1, len(stage.DDLs), truncateSQL(ddl, 80)), map[string]interface{}{
			"ddl":   ddl,
			"index": i + 1,
			"total": len(stage.DDLs),
		})

		_, err := businessDB.ExecContext(ctx, ddl)
		if err != nil {
			execution.Error = fmt.Sprintf("DDL execution failed: %v", err)
			execution.Success = false
			emit("error", "ddl", execution.Error, nil)
			return execution, fmt.Errorf("DDL failed: %w", err)
		}

		emit("ddl_complete", "ddl", fmt.Sprintf("DDL %d/%d executed successfully", i+1, len(stage.DDLs)), nil)
	}

	// Phase 2.5: Execute sample data
	if len(stage.SampleData) > 0 {
		emit("data_inserting", "data", fmt.Sprintf("Inserting %d sample data statements...", len(stage.SampleData)), nil)
		for _, stmt := range stage.SampleData {
			_, err := businessDB.ExecContext(ctx, stmt)
			if err != nil {
				logger.L().Warn("Sample data insert failed", "error", err)
				// Don't fail the stage for sample data errors
			}
		}
		emit("data_complete", "data", "Sample data inserted", nil)
	}

	// Phase 3: DDL Detection
	emit("detecting", "detect", "Detecting schema changes...", nil)

	// Parse DDL changes
	var detectedChanges []SchemaChange
	for _, ddl := range stage.DDLs {
		change := ParseDDLStatement(ddl)
		if change != nil {
			detectedChanges = append(detectedChanges, *change)
		}
	}
	execution.Changes = detectedChanges

	emit("changes_detected", "detect", fmt.Sprintf("Detected %d schema changes", len(detectedChanges)), detectedChanges)

	// Phase 4: Mark affected context as expired
	emit("marking_expired", "maintain", "Marking affected Rich Context as expired...", nil)

	expiredCount := 0
	if s.agentService != nil && len(detectedChanges) > 0 {
		count, err := s.agentService.maintainer.MarkContextExpiredByChanges(ctx, dsID, detectedChanges)
		if err != nil {
			logger.L().Warn("Failed to mark context expired", "error", err)
		} else {
			expiredCount = count
		}
	}

	if expiredCount > 0 {
		execution.ContextActions = append(execution.ContextActions, ContextAction{
			ActionType:  "expired",
			Description: fmt.Sprintf("Marked %d context entries as expired", expiredCount),
		})
		emit("context_expired", "maintain", fmt.Sprintf("Marked %d context entries as expired", expiredCount), map[string]interface{}{
			"expired_count": expiredCount,
		})
	}

	// Phase 5: Create context for new columns/tables
	emit("creating_context", "maintain", "Creating Rich Context for new schema elements...", nil)

	var contextActions []ContextAction
	for _, change := range detectedChanges {
		switch change.ChangeType {
		case ChangeTypeColumnAdded:
			action := ContextAction{
				ActionType:  "created",
				TableName:   change.TableName,
				ColumnName:  change.ColumnName,
				ContextType: "semantic",
				Description: fmt.Sprintf("Creating context for new column %s.%s", change.TableName, change.ColumnName),
			}

			if s.agentService != nil && s.agentService.llmModel != nil {
				dataType := ""
				if change.Details != nil {
					if dt, ok := change.Details["data_type"]; ok {
						dataType = fmt.Sprintf("%v", dt)
					}
				}
				bc, err := s.agentService.maintainer.CreateContextForNewColumn(ctx, dsID, change.TableName, change.ColumnName, dataType)
				if err != nil {
					action.Description = fmt.Sprintf("Failed to create context for %s.%s: %v", change.TableName, change.ColumnName, err)
				} else if bc != nil {
					action.NewContent = string(bc.Content)
				}
			}

			contextActions = append(contextActions, action)
			emit("context_created", "maintain", action.Description, action)

		case ChangeTypeTableAdded:
			action := ContextAction{
				ActionType:  "created",
				TableName:   change.TableName,
				ContextType: "table",
				Description: fmt.Sprintf("New table %s detected — context will be generated during next onboarding", change.TableName),
			}
			contextActions = append(contextActions, action)
			emit("context_created", "maintain", action.Description, action)

		case ChangeTypeColumnModified:
			action := ContextAction{
				ActionType:  "refreshed",
				TableName:   change.TableName,
				ColumnName:  change.ColumnName,
				Description: fmt.Sprintf("Column %s.%s modified — context marked for refresh", change.TableName, change.ColumnName),
			}
			contextActions = append(contextActions, action)
			emit("context_refreshed", "maintain", action.Description, action)

		case ChangeTypeColumnDropped:
			action := ContextAction{
				ActionType:  "deleted",
				TableName:   change.TableName,
				ColumnName:  change.ColumnName,
				Description: fmt.Sprintf("Column %s.%s dropped — related context invalidated", change.TableName, change.ColumnName),
			}
			contextActions = append(contextActions, action)
			emit("context_deleted", "maintain", action.Description, action)

		case ChangeTypeForeignKeyAdded:
			action := ContextAction{
				ActionType:  "created",
				TableName:   change.TableName,
				ContextType: "join_hint",
				Description: fmt.Sprintf("Foreign key added on %s — join hint context created", change.TableName),
			}
			contextActions = append(contextActions, action)
			emit("context_created", "maintain", action.Description, action)
		}
	}

	execution.ContextActions = append(execution.ContextActions, contextActions...)

	// Phase 6: Refresh expired context (if agent has LLM)
	refreshedCount := 0
	if s.agentService != nil && s.agentService.llmModel != nil && expiredCount > 0 {
		emit("refreshing_context", "maintain", "Refreshing expired Rich Context with LLM...", nil)

		results, err := s.agentService.TriggerContextRefresh(ctx, dsID)
		if err != nil {
			logger.L().Warn("Context refresh failed", "error", err)
		} else {
			for _, r := range results {
				if r.Success {
					refreshedCount++
				}
			}
		}

		if refreshedCount > 0 {
			emit("context_refreshed_complete", "maintain", fmt.Sprintf("Refreshed %d context entries", refreshedCount), map[string]interface{}{
				"refreshed_count": refreshedCount,
			})
		}
	}

	// Phase 7: Log changes
	if s.agentService != nil {
		for _, change := range detectedChanges {
			_ = s.agentService.logger.LogSchemaChange(ctx, dsID, change)
		}
	}

	// Phase 8: Update embeddings
	emit("updating_embeddings", "embed", "Updating vector embeddings...", nil)
	// Embedding update is handled by the handler layer after stage execution

	// Complete
	execution.DurationMs = time.Since(startTime).Milliseconds()
	execution.Success = true

	s.status.CurrentStage = stageID
	s.status.StageHistory = append(s.status.StageHistory, *execution)

	emit("stage_complete", "done", fmt.Sprintf("Stage %d completed in %dms", stageID, execution.DurationMs), map[string]interface{}{
		"stage_id":        stageID,
		"duration_ms":     execution.DurationMs,
		"changes":         len(detectedChanges),
		"expired":         expiredCount,
		"refreshed":       refreshedCount,
		"context_actions": len(execution.ContextActions),
	})

	return execution, nil
}

// ResetToInitial resets the evolution database to its initial state
func (s *EvolutionService) ResetToInitial(ctx context.Context, dsID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	businessDB, err := s.getBusinessDB(dsID)
	if err != nil {
		return fmt.Errorf("failed to get business DB: %w", err)
	}

	// Drop and recreate the database
	resetStatements := []string{
		"DROP TABLE IF EXISTS orders",
		"DROP TABLE IF EXISTS products",
		"DROP TABLE IF EXISTS users",
		`CREATE TABLE users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(255) UNIQUE
		)`,
		`CREATE TABLE orders (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL,
			amount DECIMAL(10,2),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		"INSERT INTO users (name, email) VALUES ('张三', 'zhang@example.com'), ('李四', 'li@example.com'), ('王五', 'wang@example.com')",
		"INSERT INTO orders (user_id, amount) VALUES (1, 99.00), (1, 199.00), (2, 59.00), (3, 299.00)",
	}

	for _, stmt := range resetStatements {
		_, err := businessDB.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("reset statement failed: %s — %w", truncateSQL(stmt, 60), err)
		}
	}

	// Clear lake-base context for this datasource
	if s.repo != nil {
		_ = s.repo.PruneAllContext(ctx, dsID)
	}

	// Reset status
	s.status = &EvolutionStatus{
		CurrentStage: 0,
		TotalStages:  len(s.stages),
		StageHistory: []StageExecution{},
		DatabaseName: "lucid_evolution",
		IsReady:      true,
	}

	return nil
}

// SyncSchemaToLakebase syncs the current business DB schema to lake-base
// Uses UpsertTable/UpsertColumn to write into rc_tables and rc_columns.
func (s *EvolutionService) SyncSchemaToLakebase(ctx context.Context, dsID int64) error {
	businessDB, err := s.getBusinessDB(dsID)
	if err != nil {
		return fmt.Errorf("failed to get business DB: %w", err)
	}

	// Query current columns
	rows, err := businessDB.QueryContext(ctx, `
		SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_KEY, IS_NULLABLE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME, ORDINAL_POSITION
	`)
	if err != nil {
		return fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	if s.repo == nil {
		return nil
	}

	seenTables := make(map[string]bool)
	for rows.Next() {
		var tableName, columnName, dataType, columnKey, isNullable string
		if err := rows.Scan(&tableName, &columnName, &dataType, &columnKey, &isNullable); err != nil {
			continue
		}

		// Upsert table (once per table)
		if !seenTables[tableName] {
			seenTables[tableName] = true
			if err := s.repo.UpsertTable(ctx, dsID, tableName, 0); err != nil {
				return fmt.Errorf("failed to upsert table %s: %w", tableName, err)
			}
		}

		// Upsert column
		isPK := columnKey == "PRI"
		isNull := isNullable == "YES"
		isFK := columnKey == "MUL" // approximate FK detection
		if err := s.repo.UpsertColumn(ctx, dsID, tableName, columnName, dataType, isNull, isPK, isFK); err != nil {
			return fmt.Errorf("failed to upsert column %s.%s: %w", tableName, columnName, err)
		}
	}

	return nil
}

// getBusinessDB gets the business database connection for a datasource
func (s *EvolutionService) getBusinessDB(dsID int64) (*sql.DB, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("connection pool not available")
	}

	// Get datasource info to find the business DB connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ds, err := s.repo.GetDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Build DSN for business DB
	host := "127.0.0.1"
	port := 3310
	user := "root"
	password := "your_strong_password"
	dbName := "lucid_evolution"

	if ds.Host.Valid {
		host = ds.Host.String
	}
	if ds.Port.Valid {
		port = int(ds.Port.Int32)
	}
	if ds.Username.Valid {
		user = ds.Username.String
	}
	if ds.DatabaseName.Valid {
		dbName = ds.DatabaseName.String
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open business DB: %w", err)
	}

	return db, nil
}

// GetStagePreview returns what a stage will do without executing
func (s *EvolutionService) GetStagePreview(stageID int) (*EvolutionStage, error) {
	if stageID < 1 || stageID > len(s.stages) {
		return nil, fmt.Errorf("invalid stage ID: %d", stageID)
	}
	stage := s.stages[stageID-1]
	return &stage, nil
}

// truncateSQL truncates a SQL string for display
func truncateSQL(sql string, maxLen int) string {
	// Normalize whitespace
	s := strings.Join(strings.Fields(sql), " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// SchemaChangeToDisplay formats a schema change for frontend display
func SchemaChangeToDisplay(change SchemaChange) map[string]interface{} {
	result := map[string]interface{}{
		"type":  string(change.ChangeType),
		"table": change.TableName,
	}
	if change.ColumnName != "" {
		result["column"] = change.ColumnName
	}
	if change.OldDefinition != "" {
		result["old"] = change.OldDefinition
	}
	if change.NewDefinition != "" {
		result["new"] = change.NewDefinition
	}
	return result
}
