// Package react provides a unified ReAct (Reasoning + Acting) engine.
//
// The engine is scenario-agnostic: different use cases (inference, rc_gen, onboarding, maintenance)
// are expressed purely through different EngineConfig (prompt + tools + hyperparams).
// The engine handles the ReAct loop, SSE streaming callbacks, and cross-cutting concerns.
package react

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"lucid/internal/logger"
)

// EngineConfig configures a ReAct engine instance for a specific scenario.
type EngineConfig struct {
	// Hyperparams
	MaxIterations    int // Max ReAct iterations (what we tell the LLM)
	MinIterations    int // Minimum iterations before allowing early stop (informational, embedded in prompt)
	ActualMaxOverride int // If > 0, use this as the actual max instead of MaxIterations*3. Set to 0 for default behaviour.

	// Prompt
	SystemPrompt string // The full prompt including goals, workflow, examples

	// Tools
	Tools []tools.Tool // Pluggable tool set (execute_sql, set_rich_context, etc.)

	// Callbacks
	StepCallback StepCallback // Called on each Think/Act/Observe step for SSE streaming

	// Cross-cutting
	LogMode string // "simple" | "full" | "quiet"
	Verbose bool   // Extra logging
}

// ObservationInjectable is implemented by tools that need direct observation
// injection (bypassing langchaingo's unreliable HandleToolEnd).
type ObservationInjectable interface {
	SetObservationInjector(injector ObservationInjector)
}

// ObservationInjector pushes observations into the ReAct handler.
type ObservationInjector interface {
	InjectObservation(output string)
}

// StepCallback is invoked on each ReAct step for real-time streaming.
// eventType: "thought" | "action" | "observation" | "finish"
type StepCallback func(step Step, eventType string)

// Step represents a single ReAct iteration.
type Step struct {
	Iteration   int         `json:"iteration"`
	Thought     string      `json:"thought,omitempty"`
	Action      string      `json:"action,omitempty"`
	ActionInput interface{} `json:"action_input,omitempty"`
	Observation string      `json:"observation,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Result holds the outcome of a ReAct execution.
type Result struct {
	Output     string `json:"output"`      // Final Answer text
	Steps      []Step `json:"steps"`       // All collected steps
	Iterations int    `json:"iterations"`  // Total iterations used
	StartedAt  time.Time
	Duration   time.Duration
}

// Engine is a reusable ReAct execution engine.
type Engine struct {
	llm    llms.Model
	config *EngineConfig
}

// New creates a new ReAct engine.
func New(llm llms.Model, config *EngineConfig) *Engine {
	if config.MaxIterations <= 0 {
		config.MaxIterations = 10
	}
	if config.LogMode == "" {
		config.LogMode = "simple"
	}
	return &Engine{llm: llm, config: config}
}

// Execute runs the ReAct loop with the given input prompt.
// The input is appended to the SystemPrompt configured in EngineConfig.
func (e *Engine) Execute(ctx context.Context, input string) (*Result, error) {
	result := &Result{StartedAt: time.Now()}

	// Build handler for step collection + SSE callbacks
	handler := &Handler{
		logMode:      e.config.LogMode,
		stepCallback: e.config.StepCallback,
	}

	// Wire up observation injection for tools that need it
	for _, tool := range e.config.Tools {
		if injectable, ok := tool.(ObservationInjectable); ok {
			injectable.SetObservationInjector(handler)
		}
	}

	// Actual iterations: give the agent more room than what we claim in the prompt
	actualMax := e.config.ActualMaxOverride
	if actualMax <= 0 {
		actualMax = e.config.MaxIterations * 3
		if actualMax < 15 {
			actualMax = 15
		}
	}

	executor, err := agents.Initialize(
		e.llm,
		e.config.Tools,
		agents.ZeroShotReactDescription,
		agents.WithMaxIterations(actualMax),
		agents.WithCallbacksHandler(handler),
	)
	if err != nil {
		return nil, fmt.Errorf("react engine: failed to initialize agent: %w", err)
	}

	// Combine system prompt + input
	fullPrompt := e.config.SystemPrompt
	if input != "" {
		fullPrompt += "\n\n" + input
	}

	log := logger.With("component", "react_engine")
	log.Info("starting ReAct execution",
		"max_iterations", e.config.MaxIterations,
		"actual_max", actualMax,
		"tools_count", len(e.config.Tools),
		"log_mode", e.config.LogMode,
	)

	// Execute the agent
	agentResult, err := executor.Call(ctx, map[string]any{"input": fullPrompt})
	if err != nil {
		log.Error("agent execution failed", "error", err)
		return nil, fmt.Errorf("react engine: agent execution failed: %w", err)
	}

	// Collect results
	result.Steps = handler.GetSteps()
	result.Iterations = len(result.Steps)
	result.Duration = time.Since(result.StartedAt)

	if output, ok := agentResult["output"].(string); ok {
		result.Output = strings.TrimSpace(output)
	}

	log.Info("ReAct execution completed",
		"duration", result.Duration,
		"iterations", result.Iterations,
		"output_length", len(result.Output),
	)

	return result, nil
}
