package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"atlas/internal/logger"
)

// RegisterTask is a Coordinator tool that registers maintenance tasks.
// Tasks are accumulated internally and can be retrieved after agent execution.
type RegisterTask struct {
	tasks []json.RawMessage
	mu    sync.Mutex
	seq   int
}

func NewRegisterTask() *RegisterTask {
	return &RegisterTask{}
}

func (t *RegisterTask) Name() string { return "register_task" }
func (t *RegisterTask) Description() string {
	return `Register a maintenance task for the Executor agent.
Input: JSON object with fields:
  - "action": one of "create", "refresh", "delete"
  - "target": one of "table", "column"
  - "table_name": the target table name
  - "column_name": the target column name (required for column targets)
  - "context": reason/details explaining why this task is needed
Output: confirmation with task ID.`
}

type registerTaskInput struct {
	Action     string `json:"action"`
	Target     string `json:"target"`
	TableName  string `json:"table_name"`
	ColumnName string `json:"column_name,omitempty"`
	Context    string `json:"context"`
}

func (t *RegisterTask) Call(ctx context.Context, input string) (string, error) {
	log := logger.With("component", "register_task")

	var inp registerTaskInput
	if err := json.Unmarshal([]byte(strings.TrimSpace(input)), &inp); err != nil {
		return fmt.Sprintf("Error: invalid JSON input: %v", err), nil
	}

	if inp.Action == "" || inp.Target == "" || inp.TableName == "" {
		return "Error: 'action', 'target', and 'table_name' are required.", nil
	}

	// Validate action
	switch inp.Action {
	case "create", "refresh", "delete":
		// ok
	default:
		return fmt.Sprintf("Error: invalid action '%s'. Must be create, refresh, or delete.", inp.Action), nil
	}

	t.mu.Lock()
	t.seq++
	taskID := fmt.Sprintf("task_%d", t.seq)

	// Store as raw JSON with injected ID
	taskObj := map[string]interface{}{
		"id":          taskID,
		"action":      inp.Action,
		"target":      inp.Target,
		"table_name":  inp.TableName,
		"column_name": inp.ColumnName,
		"context":     inp.Context,
	}
	raw, _ := json.Marshal(taskObj)
	t.tasks = append(t.tasks, raw)
	t.mu.Unlock()

	log.Info("registered maintenance task", "id", taskID, "action", inp.Action, "target", inp.Target,
		"table", inp.TableName, "column", inp.ColumnName)

	return fmt.Sprintf("Task registered: %s (action=%s, target=%s.%s)", taskID, inp.Action, inp.TableName, inp.ColumnName), nil
}

// GetTasksJSON returns all registered tasks as a JSON array string.
// Call after agent execution to pass to the executor.
func (t *RegisterTask) GetTasksJSON() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.tasks) == 0 {
		return "[]"
	}
	arr, _ := json.MarshalIndent(t.tasks, "", "  ")
	return string(arr)
}

// GetTaskCount returns the number of registered tasks.
func (t *RegisterTask) GetTaskCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.tasks)
}
