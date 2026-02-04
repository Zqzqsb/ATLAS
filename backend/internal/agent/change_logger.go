// Package agent provides self-maintenance capabilities for ReActSQL.
package agent

import (
	"context"
	"encoding/json"
	"time"

	"lucid/internal/lakebase"
)

// ChangeLogger records all schema and context changes to rc_change_log
type ChangeLogger struct {
	repo *lakebase.MySQLRepository
}

// NewChangeLogger creates a new change logger
func NewChangeLogger(repo *lakebase.MySQLRepository) *ChangeLogger {
	return &ChangeLogger{repo: repo}
}

// LogSchemaChange logs a schema change event
func (l *ChangeLogger) LogSchemaChange(ctx context.Context, dsID int64, change SchemaChange) error {
	changeDetail, _ := json.Marshal(change)

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		TableName:     change.TableName,
		ChangeType:    lakebase.ChangeTypeSchemaChange,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceAgent,
		ChangeReason:  string(change.ChangeType),
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// LogSchemaChanges logs multiple schema changes
func (l *ChangeLogger) LogSchemaChanges(ctx context.Context, dsID int64, changes []SchemaChange) error {
	for _, change := range changes {
		if err := l.LogSchemaChange(ctx, dsID, change); err != nil {
			return err
		}
	}
	return nil
}

// LogContextExpired logs a context expiration event
func (l *ChangeLogger) LogContextExpired(ctx context.Context, dsID int64, tableName, columnName, contextType, reason string) error {
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"table_name":   tableName,
		"column_name":  columnName,
		"context_type": contextType,
		"reason":       reason,
		"expired_at":   time.Now(),
	})

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		TableName:     tableName,
		ChangeType:    lakebase.ChangeTypeContextExpire,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceAgent,
		ChangeReason:  reason,
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// LogContextUpdated logs a context update event
func (l *ChangeLogger) LogContextUpdated(ctx context.Context, dsID int64, tableName, columnName, contextType string, oldContent, newContent json.RawMessage) error {
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"table_name":   tableName,
		"column_name":  columnName,
		"context_type": contextType,
		"updated_at":   time.Now(),
	})

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		TableName:     tableName,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		OldValue:      oldContent,
		NewValue:      newContent,
		TriggerSource: lakebase.TriggerSourceAgent,
		ChangeReason:  "Context refreshed by agent",
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// LogContextCreated logs a new context creation event
func (l *ChangeLogger) LogContextCreated(ctx context.Context, dsID int64, tableName, columnName, contextType string) error {
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"table_name":   tableName,
		"column_name":  columnName,
		"context_type": contextType,
		"created_at":   time.Now(),
	})

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		TableName:     tableName,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceAgent,
		ChangeReason:  "New context created by agent",
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// LogMaintenanceRun logs a maintenance run event
func (l *ChangeLogger) LogMaintenanceRun(ctx context.Context, dsID int64, result *MaintenanceResult) error {
	changeDetail, _ := json.Marshal(result)

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceAgent,
		ChangeReason:  "Scheduled maintenance run",
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// LogUserTriggeredMaintenance logs a user-triggered maintenance event
func (l *ChangeLogger) LogUserTriggeredMaintenance(ctx context.Context, dsID int64, action string, result interface{}) error {
	changeDetail, _ := json.Marshal(map[string]interface{}{
		"action":       action,
		"result":       result,
		"triggered_at": time.Now(),
	})

	log := &lakebase.ChangeLog{
		DatasourceID:  dsID,
		ChangeType:    lakebase.ChangeTypeContextUpdate,
		ChangeDetail:  changeDetail,
		TriggerSource: lakebase.TriggerSourceUser,
		ChangeReason:  "User triggered: " + action,
	}

	_, err := l.repo.CreateChangeLog(ctx, log)
	return err
}

// GetRecentChanges retrieves recent changes for a datasource
func (l *ChangeLogger) GetRecentChanges(ctx context.Context, dsID int64, limit int) ([]*lakebase.ChangeLog, error) {
	return l.repo.GetChangeLogsByDatasource(ctx, dsID, limit)
}

// GetChangesByType retrieves changes of a specific type
func (l *ChangeLogger) GetChangesByType(ctx context.Context, dsID int64, changeType lakebase.ChangeType, limit int) ([]*lakebase.ChangeLog, error) {
	return l.repo.GetChangeLogsByType(ctx, dsID, changeType, limit)
}

// GetChangesByTable retrieves changes for a specific table
func (l *ChangeLogger) GetChangesByTable(ctx context.Context, dsID int64, tableName string, limit int) ([]*lakebase.ChangeLog, error) {
	return l.repo.GetChangeLogsByTable(ctx, dsID, tableName, limit)
}

// MaintenanceResult holds the result of a maintenance run
type MaintenanceResult struct {
	DatasourceID        int64     `json:"datasource_id"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	DurationMs          int64     `json:"duration_ms"`
	SchemaChangesFound  int       `json:"schema_changes_found"`
	ContextExpired      int       `json:"context_expired"`
	ContextRefreshed    int       `json:"context_refreshed"`
	ContextCreated      int       `json:"context_created"`
	EmbeddingsUpdated   int       `json:"embeddings_updated"`
	Errors              []string  `json:"errors,omitempty"`
	Success             bool      `json:"success"`
}

// NewMaintenanceResult creates a new maintenance result
func NewMaintenanceResult(dsID int64) *MaintenanceResult {
	return &MaintenanceResult{
		DatasourceID: dsID,
		StartTime:    time.Now(),
		Errors:       []string{},
	}
}

// Complete marks the maintenance as complete
func (r *MaintenanceResult) Complete() {
	r.EndTime = time.Now()
	r.DurationMs = r.EndTime.Sub(r.StartTime).Milliseconds()
	r.Success = len(r.Errors) == 0
}

// AddError adds an error to the result
func (r *MaintenanceResult) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

// ChangeLogSummary provides a summary of change logs
type ChangeLogSummary struct {
	TotalChanges     int            `json:"total_changes"`
	SchemaChanges    int            `json:"schema_changes"`
	ContextUpdates   int            `json:"context_updates"`
	ContextExpiries  int            `json:"context_expiries"`
	ByTable          map[string]int `json:"by_table"`
	ByTriggerSource  map[string]int `json:"by_trigger_source"`
	OldestChange     time.Time      `json:"oldest_change"`
	NewestChange     time.Time      `json:"newest_change"`
}

// GetChangeLogSummary generates a summary of change logs
func (l *ChangeLogger) GetChangeLogSummary(ctx context.Context, dsID int64, limit int) (*ChangeLogSummary, error) {
	logs, err := l.repo.GetChangeLogsByDatasource(ctx, dsID, limit)
	if err != nil {
		return nil, err
	}

	summary := &ChangeLogSummary{
		ByTable:         make(map[string]int),
		ByTriggerSource: make(map[string]int),
	}

	for i, log := range logs {
		summary.TotalChanges++

		switch log.ChangeType {
		case lakebase.ChangeTypeSchemaChange:
			summary.SchemaChanges++
		case lakebase.ChangeTypeContextUpdate:
			summary.ContextUpdates++
		case lakebase.ChangeTypeContextExpire:
			summary.ContextExpiries++
		}

		if log.TableName != "" {
			summary.ByTable[log.TableName]++
		}

		summary.ByTriggerSource[string(log.TriggerSource)]++

		if i == 0 {
			summary.NewestChange = log.CreatedAt
		}
		summary.OldestChange = log.CreatedAt
	}

	return summary, nil
}
