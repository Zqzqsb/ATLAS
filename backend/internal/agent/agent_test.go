package agent

import (
	"testing"
	"time"
)

func TestParseDDLStatement_AlterTableAddColumn(t *testing.T) {
	tests := []struct {
		name       string
		sql        string
		wantType   SchemaChangeType
		wantTable  string
		wantColumn string
	}{
		{
			name:       "ADD COLUMN basic",
			sql:        "ALTER TABLE users ADD COLUMN email VARCHAR(255)",
			wantType:   ChangeTypeColumnAdded,
			wantTable:  "users",
			wantColumn: "email",
		},
		{
			name:       "ADD COLUMN with COLUMN keyword",
			sql:        "ALTER TABLE orders ADD COLUMN status INT",
			wantType:   ChangeTypeColumnAdded,
			wantTable:  "orders",
			wantColumn: "status",
		},
		{
			name:       "DROP COLUMN",
			sql:        "ALTER TABLE users DROP COLUMN temp_field",
			wantType:   ChangeTypeColumnDropped,
			wantTable:  "users",
			wantColumn: "temp_field",
		},
		{
			name:       "MODIFY COLUMN",
			sql:        "ALTER TABLE products MODIFY name VARCHAR(500)",
			wantType:   ChangeTypeColumnModified,
			wantTable:  "products",
			wantColumn: "name",
		},
		{
			name:       "ADD CONSTRAINT FOREIGN KEY",
			sql:        "ALTER TABLE orders ADD CONSTRAINT fk_order_product FOREIGN KEY (product_id) REFERENCES products(id)",
			wantType:   ChangeTypeForeignKeyAdded,
			wantTable:  "orders",
			wantColumn: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := ParseDDLStatement(tt.sql)
			if change == nil {
				t.Fatal("Expected change, got nil")
			}
			if change.ChangeType != tt.wantType {
				t.Errorf("ChangeType = %v, want %v", change.ChangeType, tt.wantType)
			}
			if change.TableName != tt.wantTable {
				t.Errorf("TableName = %v, want %v", change.TableName, tt.wantTable)
			}
			if change.ColumnName != tt.wantColumn {
				t.Errorf("ColumnName = %v, want %v", change.ColumnName, tt.wantColumn)
			}
		})
	}
}

func TestParseDDLStatement_CreateDropTable(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantType  SchemaChangeType
		wantTable string
	}{
		{
			name:      "CREATE TABLE",
			sql:       "CREATE TABLE new_table (id INT)",
			wantType:  ChangeTypeTableAdded,
			wantTable: "new_table",
		},
		{
			name:      "DROP TABLE",
			sql:       "DROP TABLE old_table",
			wantType:  ChangeTypeTableDropped,
			wantTable: "old_table",
		},
		{
			name:      "DROP TABLE IF EXISTS",
			sql:       "DROP TABLE IF EXISTS temp_table",
			wantType:  ChangeTypeTableDropped,
			wantTable: "temp_table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := ParseDDLStatement(tt.sql)
			if change == nil {
				t.Fatal("Expected change, got nil")
			}
			if change.ChangeType != tt.wantType {
				t.Errorf("ChangeType = %v, want %v", change.ChangeType, tt.wantType)
			}
			if change.TableName != tt.wantTable {
				t.Errorf("TableName = %v, want %v", change.TableName, tt.wantTable)
			}
		})
	}
}

func TestParseDDLStatement_NonDDL(t *testing.T) {
	tests := []string{
		"SELECT * FROM users",
		"INSERT INTO users VALUES (1, 'test')",
		"UPDATE users SET name = 'test'",
		"DELETE FROM users WHERE id = 1",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			change := ParseDDLStatement(sql)
			if change != nil {
				t.Errorf("Expected nil for non-DDL statement, got %v", change)
			}
		})
	}
}

func TestSchemaChangeTypes(t *testing.T) {
	tests := []struct {
		changeType SchemaChangeType
		expected   string
	}{
		{ChangeTypeTableAdded, "table_added"},
		{ChangeTypeTableDropped, "table_dropped"},
		{ChangeTypeColumnAdded, "column_added"},
		{ChangeTypeColumnDropped, "column_dropped"},
		{ChangeTypeColumnModified, "column_modified"},
		{ChangeTypeIndexAdded, "index_added"},
		{ChangeTypeIndexDropped, "index_dropped"},
		{ChangeTypeForeignKeyAdded, "fk_added"},
		{ChangeTypeForeignKeyDropped, "fk_dropped"},
	}

	for _, tt := range tests {
		t.Run(string(tt.changeType), func(t *testing.T) {
			if string(tt.changeType) != tt.expected {
				t.Errorf("SchemaChangeType = %v, want %v", tt.changeType, tt.expected)
			}
		})
	}
}

func TestMaintenanceResult(t *testing.T) {
	result := NewMaintenanceResult(1)

	if result.DatasourceID != 1 {
		t.Errorf("DatasourceID = %d, want 1", result.DatasourceID)
	}

	if result.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}

	result.SchemaChangesFound = 5
	result.ContextExpired = 3
	result.AddError("test error")

	// Add small delay to ensure DurationMs is positive
	time.Sleep(time.Millisecond)

	result.Complete()

	if result.EndTime.IsZero() {
		t.Error("EndTime should not be zero after Complete()")
	}

	if result.DurationMs < 0 {
		t.Error("DurationMs should be non-negative after Complete()")
	}

	if result.Success {
		t.Error("Success should be false when there are errors")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Errors count = %d, want 1", len(result.Errors))
	}
}

func TestSignalToJSON(t *testing.T) {
	signal := &MaintenanceSignal{
		Type:         SignalDDL,
		DatasourceID: 1,
		Changes: []SchemaChange{
			{ChangeType: ChangeTypeColumnAdded, TableName: "users", ColumnName: "phone"},
		},
		TriggeredBy: "test",
	}
	json := SignalToJSON(signal)
	if json == "" {
		t.Error("Expected non-empty JSON")
	}
	if !contains(json, "users") {
		t.Error("JSON should contain table name")
	}
}

func TestTasksToJSON(t *testing.T) {
	tasks := []MaintenanceTask{
		{ID: "t1", Action: TaskActionCreate, Target: TaskTargetColumn, TableName: "users", ColumnName: "phone"},
	}
	json := TasksToJSON(tasks)
	if json == "" {
		t.Error("Expected non-empty JSON")
	}
	if !contains(json, "phone") {
		t.Error("JSON should contain column name")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
