// Package agent provides self-maintenance capabilities for LUCID.
package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"

	"lucid/internal/lakebase"
)

// AgentService orchestrates DDL detection and context maintenance
type AgentService struct {
	pool         *lakebase.ConnectionPool
	repo         *lakebase.MySQLRepository
	vectorRepo   *lakebase.MySQLVectorRepository
	detector     *DDLDetector
	maintainer   *ContextMaintainer
	logger       *ChangeLogger
	llmModel     llms.Model
	config       *AgentConfig

	// Background task management
	stopChan     chan struct{}
	running      bool
	mu           sync.RWMutex
	lastRun      time.Time
	lastResult   *MaintenanceResult
}

// AgentConfig holds configuration for the agent service
type AgentConfig struct {
	EnableDDLDetection  bool          `yaml:"enable_ddl_detection"`
	CheckInterval       time.Duration `yaml:"check_interval"`
	AutoRefreshContext  bool          `yaml:"auto_refresh_context"`
	MaxConcurrentTasks  int           `yaml:"max_concurrent_tasks"`
}

// DefaultAgentConfig returns default agent configuration
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		EnableDDLDetection:  true,
		CheckInterval:       60 * time.Second,
		AutoRefreshContext:  true,
		MaxConcurrentTasks:  2,
	}
}

// NewAgentService creates a new agent service
func NewAgentService(pool *lakebase.ConnectionPool, config *AgentConfig) *AgentService {
	if config == nil {
		config = DefaultAgentConfig()
	}

	repo := lakebase.NewMySQLRepository(pool)
	vectorRepo := lakebase.NewMySQLVectorRepository(pool)
	logger := NewChangeLogger(repo)

	return &AgentService{
		pool:       pool,
		repo:       repo,
		vectorRepo: vectorRepo,
		detector:   NewDDLDetector(repo),
		logger:     logger,
		config:     config,
		stopChan:   make(chan struct{}),
	}
}

// SetLLMModel sets the LLM model for context regeneration
func (s *AgentService) SetLLMModel(model llms.Model) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.llmModel = model
	if s.maintainer != nil {
		s.maintainer.SetLLMModel(model)
	} else {
		s.maintainer = NewContextMaintainer(s.repo, model, s.logger)
	}
}

// Start starts the background maintenance loop
func (s *AgentService) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("agent service already running")
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	go s.maintenanceLoop()
	return nil
}

// Stop stops the background maintenance loop
func (s *AgentService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()
}

// IsRunning returns whether the service is running
func (s *AgentService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetLastRun returns the last maintenance run time and result
func (s *AgentService) GetLastRun() (time.Time, *MaintenanceResult) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRun, s.lastResult
}

// maintenanceLoop runs periodic maintenance tasks
func (s *AgentService) maintenanceLoop() {
	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.runMaintenanceForAllDatasources()
		}
	}
}

// runMaintenanceForAllDatasources runs maintenance for all datasources
func (s *AgentService) runMaintenanceForAllDatasources() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	datasources, err := s.repo.ListDatasources(ctx)
	if err != nil {
		return
	}

	for _, ds := range datasources {
		if ds.Status != lakebase.DatasourceStatusEnabled {
			continue
		}
		_, _ = s.RunMaintenance(ctx, ds.ID)
	}
}

// RunMaintenance runs maintenance for a specific datasource
func (s *AgentService) RunMaintenance(ctx context.Context, dsID int64) (*MaintenanceResult, error) {
	result := NewMaintenanceResult(dsID)
	defer result.Complete()

	s.mu.Lock()
	s.lastRun = time.Now()
	s.mu.Unlock()

	// 1. Check and mark expired context by time
	if s.maintainer != nil {
		expired, err := s.maintainer.CheckAndMarkExpiredByTime(ctx, dsID)
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to check time expiry: %v", err))
		} else {
			result.ContextExpired += expired
		}
	}

	// 2. Auto-refresh expired context if enabled
	if s.config.AutoRefreshContext && s.maintainer != nil {
		refreshResults, err := s.maintainer.RefreshAllExpiredContext(ctx, dsID)
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to refresh context: %v", err))
		} else {
			for _, r := range refreshResults {
				if r.Success {
					result.ContextRefreshed++
				}
			}
		}
	}

	// 3. Log the maintenance run
	if err := s.logger.LogMaintenanceRun(ctx, dsID, result); err != nil {
		result.AddError(fmt.Sprintf("Failed to log maintenance: %v", err))
	}

	s.mu.Lock()
	s.lastResult = result
	s.mu.Unlock()

	return result, nil
}

// DetectAndProcessChanges detects schema changes and processes them
func (s *AgentService) DetectAndProcessChanges(ctx context.Context, dsID int64, currentSchema map[string]*SchemaSnapshot) (*MaintenanceResult, error) {
	result := NewMaintenanceResult(dsID)
	defer result.Complete()

	// 1. Detect schema changes
	changes, err := s.detector.DetectChanges(ctx, dsID, currentSchema)
	if err != nil {
		result.AddError(fmt.Sprintf("Failed to detect changes: %v", err))
		return result, err
	}
	result.SchemaChangesFound = len(changes)

	if len(changes) == 0 {
		return result, nil
	}

	// 2. Log schema changes
	if err := s.logger.LogSchemaChanges(ctx, dsID, changes); err != nil {
		result.AddError(fmt.Sprintf("Failed to log schema changes: %v", err))
	}

	// 3. Mark affected context as expired
	if s.maintainer != nil {
		expired, err := s.maintainer.MarkContextExpiredByChanges(ctx, dsID, changes)
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to mark context expired: %v", err))
		} else {
			result.ContextExpired = expired
		}
	}

	// 4. Create context for new columns if LLM is available
	for _, change := range changes {
		if change.ChangeType == ChangeTypeColumnAdded && s.maintainer != nil && s.llmModel != nil {
			_, err := s.maintainer.CreateContextForNewColumn(ctx, dsID, change.TableName, change.ColumnName,
				getDataTypeFromDetails(change.Details))
			if err != nil {
				result.AddError(fmt.Sprintf("Failed to create context for new column: %v", err))
			} else {
				result.ContextCreated++
			}
		}
	}

	// 5. Auto-refresh if enabled
	if s.config.AutoRefreshContext && s.maintainer != nil {
		refreshResults, err := s.maintainer.RefreshAllExpiredContext(ctx, dsID)
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to refresh context: %v", err))
		} else {
			for _, r := range refreshResults {
				if r.Success {
					result.ContextRefreshed++
				}
			}
		}
	}

	return result, nil
}

// getDataTypeFromDetails extracts data type from change details
func getDataTypeFromDetails(details map[string]interface{}) string {
	if dt, ok := details["data_type"].(string); ok {
		return dt
	}
	return "UNKNOWN"
}

// ProcessDDLStatement processes a DDL statement and triggers maintenance
func (s *AgentService) ProcessDDLStatement(ctx context.Context, dsID int64, sql string) (*MaintenanceResult, error) {
	result := NewMaintenanceResult(dsID)
	defer result.Complete()

	// Parse the DDL statement
	change := ParseDDLStatement(sql)
	if change == nil {
		return result, nil
	}

	// Log the change
	if err := s.logger.LogSchemaChange(ctx, dsID, *change); err != nil {
		result.AddError(fmt.Sprintf("Failed to log DDL: %v", err))
	}
	result.SchemaChangesFound = 1

	// Mark affected context as expired
	if s.maintainer != nil {
		expired, err := s.maintainer.MarkContextExpiredByChanges(ctx, dsID, []SchemaChange{*change})
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to mark context expired: %v", err))
		} else {
			result.ContextExpired = expired
		}
	}

	// Handle new column case
	if change.ChangeType == ChangeTypeColumnAdded && s.maintainer != nil {
		dataType := ""
		if change.Details != nil {
			if dt, ok := change.Details["data_type"].(string); ok {
				dataType = dt
			}
		}
		_, err := s.maintainer.CreateContextForNewColumn(ctx, dsID, change.TableName, change.ColumnName, dataType)
		if err != nil {
			result.AddError(fmt.Sprintf("Failed to create context: %v", err))
		} else {
			result.ContextCreated++
		}
	}

	return result, nil
}

// TriggerContextRefresh manually triggers context refresh for a datasource
func (s *AgentService) TriggerContextRefresh(ctx context.Context, dsID int64) ([]*ContextUpdateResult, error) {
	if s.maintainer == nil {
		return nil, fmt.Errorf("maintainer not initialized")
	}

	results, err := s.maintainer.RefreshAllExpiredContext(ctx, dsID)
	if err != nil {
		return nil, err
	}

	// Log user-triggered maintenance
	s.logger.LogUserTriggeredMaintenance(ctx, dsID, "context_refresh", map[string]int{
		"total":   len(results),
		"success": countSuccessful(results),
	})

	return results, nil
}

// countSuccessful counts successful context updates
func countSuccessful(results []*ContextUpdateResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}

// GetMaintenanceStatus returns the current maintenance status
func (s *AgentService) GetMaintenanceStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"running":              s.running,
		"config":               s.config,
		"llm_available":        s.llmModel != nil,
		"last_run":             s.lastRun,
		"auto_refresh_enabled": s.config.AutoRefreshContext,
	}

	if s.lastResult != nil {
		status["last_result"] = s.lastResult
	}

	return status
}

// GetChangeLogSummary returns a summary of changes for a datasource
func (s *AgentService) GetChangeLogSummary(ctx context.Context, dsID int64, limit int) (*ChangeLogSummary, error) {
	return s.logger.GetChangeLogSummary(ctx, dsID, limit)
}

// GetRecentChanges returns recent changes for a datasource
func (s *AgentService) GetRecentChanges(ctx context.Context, dsID int64, limit int) ([]*lakebase.ChangeLog, error) {
	return s.logger.GetRecentChanges(ctx, dsID, limit)
}


