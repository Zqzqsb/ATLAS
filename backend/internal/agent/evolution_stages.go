package agent

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"atlas/internal/adapter"
	"atlas/internal/lakebase"
	"atlas/internal/logger"
	"atlas/internal/react"
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

// Predefined evolution stages for the demo database atlas_evolution
var DefaultEvolutionStages = []EvolutionStage{
	{
		ID:          1,
		Name:        "Add User Phone Column",
		Description: "Business need: add user phone number for contact",
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
		Description: "Business need: introduce product catalog",
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
			"INSERT INTO products (name, price, category) VALUES ('iPhone 15', 7999.00, 'Electronics')",
			"INSERT INTO products (name, price, category) VALUES ('MacBook Pro', 14999.00, 'Electronics')",
			"INSERT INTO products (name, price, category) VALUES ('AirPods Pro', 1899.00, 'Accessories')",
			"INSERT INTO products (name, price, category) VALUES ('iPad Air', 4799.00, 'Electronics')",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeTableAdded},
	},
	{
		ID:          3,
		Name:        "Add Order-Product Foreign Key",
		Description: "Business need: link orders to products",
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
		Description: "Business need: upgrade amount precision from 2 to 4 decimal places",
		DDLs: []string{
			"ALTER TABLE orders MODIFY COLUMN amount DECIMAL(15,4)",
		},
		ExpectedChanges: []SchemaChangeType{ChangeTypeColumnModified},
	},
	{
		ID:          5,
		Name:        "Drop Email Column",
		Description: "Business change: remove email column for privacy compliance",
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
	businessDB   *sql.DB // cached business DB connection
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
			DatabaseName: "atlas_evolution",
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

	// Auto-reset: when executing Stage 1, always reset to initial state first
	// This handles the case where the server restarted (status.CurrentStage=0) but
	// the database still has DDL changes from a previous session.
	if stageID == 1 {
		emit("reset_start", "reset", "Auto-resetting database to initial state before Stage 1...", nil)
		if err := s.resetToInitialLocked(ctx, dsID); err != nil {
			execution.Error = fmt.Sprintf("auto-reset failed: %v", err)
			execution.Success = false
			emit("error", "reset", execution.Error, nil)
			return execution, err
		}
		emit("reset_complete", "reset", "Database reset to initial state", nil)
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

	// Phase 4: Sync schema to lake-base (captures new tables/columns before agents run)
	emit("syncing_schema", "sync", "Syncing schema to lake-base...", nil)
	if err := s.SyncSchemaToLakebase(ctx, dsID); err != nil {
		logger.L().Warn("Schema sync failed", "error", err)
	}
	emit("schema_synced", "sync", "Schema synced to lake-base", nil)

	// Phase 5: Build signal and run maintenance agents (Coordinator → Executor)
	if s.agentService != nil && s.agentService.llmModel != nil && len(detectedChanges) > 0 {
		emit("agent_start", "maintain", "Running maintenance agents (Coordinator → Executor)...", nil)

		signal := &MaintenanceSignal{
			Type:          SignalDDL,
			DatasourceID:  dsID,
			DDLStatements: stage.DDLs,
			Changes:       detectedChanges,
			TriggeredBy:   "evolution_stage",
			TriggeredAt:   time.Now(),
		}

		// Create a DBAdapter for the business database
		businessDBAdapter, err := s.getBusinessDBAdapter(dsID)
		if err != nil {
			logger.L().Warn("Failed to create business DB adapter for agent", "error", err)
			emit("agent_error", "maintain", fmt.Sprintf("Agent skipped: %v", err), nil)
		} else {
			// Wrap evolution events channel as a react.StepCallback for agent SSE
			// eventType comes in as "coordinator_thought", "executor_action", etc.
			stepCB := func(step react.Step, eventType string) {
				// Parse agent role and step type from eventType
				agentRole := "coordinator"
				stepType := eventType
				if strings.HasPrefix(eventType, "coordinator_") {
					stepType = strings.TrimPrefix(eventType, "coordinator_")
				} else if strings.HasPrefix(eventType, "executor_") {
					agentRole = "executor"
					stepType = strings.TrimPrefix(eventType, "executor_")
				}

				msg := ""
				switch stepType {
				case "thought":
					msg = step.Thought
				case "action":
					msg = fmt.Sprintf("[%s] %v", step.Action, step.ActionInput)
				case "observation":
					msg = step.Observation
				case "finish":
					msg = step.Thought
					if msg == "" {
						msg = "Agent finished"
					}
				default:
					if step.Thought != "" {
						msg = step.Thought
					}
					if step.Action != "" {
						msg = fmt.Sprintf("[%s] %v", step.Action, step.ActionInput)
					}
				}

				if len(msg) > 500 {
					msg = msg[:500] + "..."
				}

				emit("agent_step", "maintain", msg, map[string]interface{}{
					"event_type": eventType,
					"agent_role": agentRole,
					"step_type":  stepType,
					"tool_name":  step.Action,
					"iteration":  step.Iteration,
				})
			}

			agentResult, err := s.agentService.ProcessSignal(ctx, signal, businessDBAdapter, stepCB)
			if err != nil {
				logger.L().Warn("Maintenance agent failed", "error", err)
				emit("agent_error", "maintain", fmt.Sprintf("Agent error: %v", err), nil)
			} else {
				// Record context actions from agent tasks
				for _, task := range agentResult.Tasks {
					action := ContextAction{
						ActionType:  string(task.Action),
						TableName:   task.TableName,
						ColumnName:  task.ColumnName,
						Description: fmt.Sprintf("[%s] %s.%s — %s", task.Action, task.TableName, task.ColumnName, task.Context),
					}
					execution.ContextActions = append(execution.ContextActions, action)
				}
				emit("agent_complete", "maintain", fmt.Sprintf("Maintenance agents completed: %d tasks processed", len(agentResult.Tasks)), map[string]interface{}{
					"tasks": len(agentResult.Tasks),
				})
			}
		}
	} else if len(detectedChanges) > 0 {
		emit("agent_skipped", "maintain", "LLM not available — maintenance agents skipped", nil)
	}

	// Phase 6: Log changes
	if s.agentService != nil {
		for _, change := range detectedChanges {
			_ = s.agentService.changeLog.LogSchemaChange(ctx, dsID, change)
		}
	}

	// Phase 7: Embedding update is handled by the handler layer after stage execution
	emit("updating_embeddings", "embed", "Embedding update pending (handled after stage completion)...", nil)

	// Complete
	execution.DurationMs = time.Since(startTime).Milliseconds()
	execution.Success = true

	s.status.CurrentStage = stageID
	s.status.StageHistory = append(s.status.StageHistory, *execution)

	emit("stage_complete", "done", fmt.Sprintf("Stage %d completed in %dms", stageID, execution.DurationMs), map[string]interface{}{
		"stage_id":        stageID,
		"duration_ms":     execution.DurationMs,
		"changes":         len(detectedChanges),
		"context_actions": len(execution.ContextActions),
	})

	return execution, nil
}

// ResetToInitial resets the evolution database to its initial state
func (s *EvolutionService) ResetToInitial(ctx context.Context, dsID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.resetToInitialLocked(ctx, dsID)
}

// resetToInitialLocked performs the actual reset. Caller must hold s.mu.
func (s *EvolutionService) resetToInitialLocked(ctx context.Context, dsID int64) error {
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
			name VARCHAR(100) NOT NULL COMMENT 'User name',
			email VARCHAR(255) UNIQUE COMMENT 'User email'
		) COMMENT='User information'`,
		`CREATE TABLE orders (
			id INT PRIMARY KEY AUTO_INCREMENT,
			user_id INT NOT NULL COMMENT 'Associated user ID',
			amount DECIMAL(10,2) COMMENT 'Order amount',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
			FOREIGN KEY (user_id) REFERENCES users(id)
		) COMMENT='Order information'`,
		"INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com'), ('Bob', 'bob@example.com'), ('Charlie', 'charlie@example.com')",
		"INSERT INTO orders (user_id, amount) VALUES (1, 99.00), (1, 199.00), (2, 59.00), (3, 299.00)",
	}

	for _, stmt := range resetStatements {
		_, err := businessDB.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("reset statement failed: %s — %w", truncateSQL(stmt, 60), err)
		}
	}

	// Invalidate cached business DB connection to get a clean slate
	s.closeBusinessDB()

	// Clear lake-base context for this datasource
	if s.repo != nil {
		_ = s.repo.PruneAllContext(ctx, dsID)
	}

	// Reset status
	s.status = &EvolutionStatus{
		CurrentStage: 0,
		TotalStages:  len(s.stages),
		StageHistory: []StageExecution{},
		DatabaseName: "atlas_evolution",
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

// getBusinessDB gets or creates a cached business database connection for a datasource.
func (s *EvolutionService) getBusinessDB(dsID int64) (*sql.DB, error) {
	// Return cached connection if available and alive
	if s.businessDB != nil {
		if err := s.businessDB.Ping(); err == nil {
			return s.businessDB, nil
		}
		s.businessDB.Close()
		s.businessDB = nil
	}

	if s.pool == nil {
		return nil, fmt.Errorf("connection pool not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ds, err := s.repo.GetDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Use pool config for credentials (same MariaDB instance),
	// datasource record only provides host/port/dbName.
	poolCfg := s.pool.GetConfig()
	if poolCfg == nil {
		return nil, fmt.Errorf("pool config not available")
	}

	host := poolCfg.Host
	port := poolCfg.Port
	user := poolCfg.User
	password := poolCfg.Password
	dbName := "atlas_evolution"

	if ds.Host.Valid {
		host = ds.Host.String
	}
	if ds.Port.Valid {
		port = int(ds.Port.Int32)
	}
	if ds.DatabaseName.Valid {
		dbName = ds.DatabaseName.String
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open business DB: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping business DB: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	s.businessDB = db
	return db, nil
}

// closeBusinessDB closes the cached business DB connection.
func (s *EvolutionService) closeBusinessDB() {
	if s.businessDB != nil {
		s.businessDB.Close()
		s.businessDB = nil
	}
}

// getBusinessDBAdapter creates a fresh DBAdapter for the business database.
// Used by maintenance agents which require the adapter.DBAdapter interface.
func (s *EvolutionService) getBusinessDBAdapter(dsID int64) (adapter.DBAdapter, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("connection pool not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ds, err := s.repo.GetDatasource(ctx, dsID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	poolCfg := s.pool.GetConfig()
	if poolCfg == nil {
		return nil, fmt.Errorf("pool config not available")
	}

	host := poolCfg.Host
	port := poolCfg.Port
	user := poolCfg.User
	password := poolCfg.Password
	dbName := "atlas_evolution"

	if ds.Host.Valid {
		host = ds.Host.String
	}
	if ds.Port.Valid {
		port = int(ds.Port.Int32)
	}
	if ds.DatabaseName.Valid {
		dbName = ds.DatabaseName.String
	}

	dbAdapter, err := adapter.NewAdapter(&adapter.DBConfig{
		Type:     "mysql",
		Host:     host,
		Port:     port,
		Database: dbName,
		User:     user,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	if err := dbAdapter.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect adapter: %w", err)
	}

	return dbAdapter, nil
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
