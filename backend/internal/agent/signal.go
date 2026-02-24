// Package agent provides self-maintenance capabilities for LUCID.
package agent

import (
	"encoding/json"
	"time"
)

// SignalType represents the type of maintenance signal.
type SignalType string

const (
	SignalDDL        SignalType = "ddl"         // DDL change (ALTER/CREATE/DROP)
	SignalDataChange SignalType = "data_change" // Bulk data change (INSERT/UPDATE/TRUNCATE)
)

// MaintenanceSignal is a pure value object describing what changed.
type MaintenanceSignal struct {
	Type          SignalType     `json:"type"`
	DatasourceID  int64          `json:"datasource_id"`
	DDLStatements []string       `json:"ddl_statements,omitempty"`
	Changes       []SchemaChange `json:"changes,omitempty"`
	DataChanges   []DataChange   `json:"data_changes,omitempty"`
	TriggeredBy   string         `json:"triggered_by"`
	TriggeredAt   time.Time      `json:"triggered_at"`
}

// DataChange describes a bulk data change event.
type DataChange struct {
	TableName    string `json:"table_name"`
	ChangeType   string `json:"change_type"` // "bulk_insert" | "bulk_update" | "truncate"
	AffectedRows int64  `json:"affected_rows"`
	Description  string `json:"description"`
}

// MaintenanceTaskAction represents the type of maintenance task.
type MaintenanceTaskAction string

const (
	TaskActionCreate  MaintenanceTaskAction = "create"
	TaskActionRefresh MaintenanceTaskAction = "refresh"
	TaskActionDelete  MaintenanceTaskAction = "delete"
)

// MaintenanceTaskTarget represents what entity the task applies to.
type MaintenanceTaskTarget string

const (
	TaskTargetTable  MaintenanceTaskTarget = "table"
	TaskTargetColumn MaintenanceTaskTarget = "column"
)

// MaintenanceTask represents a single maintenance unit to be executed.
type MaintenanceTask struct {
	ID         string                `json:"id"`
	Action     MaintenanceTaskAction `json:"action"`
	Target     MaintenanceTaskTarget `json:"target"`
	TableName  string                `json:"table_name"`
	ColumnName string                `json:"column_name,omitempty"`
	Context    string                `json:"context"` // reason / details for the executor
}

// SignalToJSON serializes a signal for use as agent input.
func SignalToJSON(s *MaintenanceSignal) string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

// TasksToJSON serializes tasks for use as agent input.
func TasksToJSON(tasks []MaintenanceTask) string {
	b, _ := json.MarshalIndent(tasks, "", "  ")
	return string(b)
}
