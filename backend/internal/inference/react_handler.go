package inference

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// StepNotifier is called when a ReAct step is updated (can be partial)
type StepNotifier func(step CollectedStep, eventType string)

// PrettyReActHandler enhanced ReAct log handler with step collection
type PrettyReActHandler struct {
	iterationCount          int // total iterations (including all tool calls)
	effectiveIterationCount int // effective iterations (excluding update_rich_context)
	lastAction              string
	logMode                 string // log mode: "simple" or "full"

	// Step collection for frontend visualization
	mu            sync.Mutex
	collectedSteps []CollectedStep
	currentStep   *CollectedStep

	// Streaming callback for real-time step notifications
	stepNotifier StepNotifier
}

// SetStepNotifier sets the callback for streaming step notifications
func (h *PrettyReActHandler) SetStepNotifier(notifier StepNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stepNotifier = notifier
}

// CollectedStep represents a collected ReAct step
type CollectedStep struct {
	Step        int         `json:"step"`
	Thought     string      `json:"thought"`
	Action      string      `json:"action"`
	ActionInput interface{} `json:"action_input"` // 改为 interface{} 以支持多种类型
	Observation string      `json:"observation"`
	Phase       string      `json:"phase,omitempty"` // "schema_linking" or "sql_generation"
	Timestamp   time.Time   `json:"timestamp"`
}

// GetCollectedSteps returns all collected steps
func (h *PrettyReActHandler) GetCollectedSteps() []CollectedStep {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// If there's a pending step with content, finalize it
	if h.currentStep != nil && (h.currentStep.Action != "" || h.currentStep.Thought != "") {
		h.collectedSteps = append(h.collectedSteps, *h.currentStep)
		h.currentStep = nil
	}
	
	return h.collectedSteps
}

// finalizeCurrentStep finalizes the current step and notifies
func (h *PrettyReActHandler) finalizeCurrentStep() {
	if h.currentStep != nil && (h.currentStep.Action != "" || h.currentStep.Thought != "") {
		h.collectedSteps = append(h.collectedSteps, *h.currentStep)
		// Note: Don't notify here as we already notified during the step
	}
}

// notifyStepUpdate sends a real-time notification about step update
func (h *PrettyReActHandler) notifyStepUpdate(eventType string) {
	if h.stepNotifier != nil && h.currentStep != nil {
		h.stepNotifier(*h.currentStep, eventType)
	}
}

var _ interface {
	HandleText(ctx context.Context, text string)
	HandleLLMStart(ctx context.Context, prompts []string)
	HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent)
	HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse)
	HandleLLMError(ctx context.Context, err error)
	HandleChainStart(ctx context.Context, inputs map[string]any)
	HandleChainEnd(ctx context.Context, outputs map[string]any)
	HandleChainError(ctx context.Context, err error)
	HandleToolStart(ctx context.Context, input string)
	HandleToolEnd(ctx context.Context, output string)
	HandleToolError(ctx context.Context, err error)
	HandleAgentAction(ctx context.Context, action schema.AgentAction)
	HandleAgentFinish(ctx context.Context, finish schema.AgentFinish)
	HandleRetrieverStart(ctx context.Context, query string)
	HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document)
	HandleStreamingFunc(ctx context.Context, chunk []byte)
} = &PrettyReActHandler{}

func (h *PrettyReActHandler) HandleText(_ context.Context, text string) {}

func (h *PrettyReActHandler) HandleLLMStart(_ context.Context, prompts []string) {
	if h.logMode == "full" {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("📤 LLM Prompt (Full)")
		fmt.Println(strings.Repeat("=", 80))
		for i, prompt := range prompts {
			fmt.Printf("Prompt %d:\n%s\n", i+1, prompt)
		}
		fmt.Println(strings.Repeat("=", 80))
	}
}

func (h *PrettyReActHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
}

func (h *PrettyReActHandler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	if h.logMode == "full" {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("📥 LLM Response (Full)")
		fmt.Println(strings.Repeat("=", 80))
		for i, choice := range res.Choices {
			fmt.Printf("Choice %d:\n%s\n", i+1, choice.Content)
		}
		fmt.Println(strings.Repeat("=", 80))
	}
}

func (h *PrettyReActHandler) HandleLLMError(_ context.Context, err error) {
	fmt.Printf("❌ LLM Error: %v\n", err)
}

func (h *PrettyReActHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	// Each chain start represents a new iteration
	h.iterationCount++

	// If last action was not update_rich_context, increment effective count
	if h.lastAction != "" && h.lastAction != "update_rich_context" {
		h.effectiveIterationCount++
	}

	// Display iteration info
	if h.effectiveIterationCount > 0 {
		fmt.Printf("\n┌─ Iteration %d (Effective: %d/5) ─────────────────────────────────────────────\n", h.iterationCount, h.effectiveIterationCount)
	} else {
		fmt.Printf("\n┌─ Iteration %d ─────────────────────────────────────────────\n", h.iterationCount)
	}

	// Start collecting a new step
	h.mu.Lock()
	// Save previous step if exists and has content
	h.finalizeCurrentStep()
	h.currentStep = &CollectedStep{
		Step:      h.iterationCount,
		Timestamp: time.Now(),
	}
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	// Extract LLM response
	if text, ok := outputs["text"].(string); ok {
		// Extract thought from response
		thought := extractThought(text)
		
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Thought = thought
		}
		// Notify about thought immediately (real-time streaming)
		h.notifyStepUpdate("thought")
		h.mu.Unlock()

		// full mode: output full response
		if h.logMode == "full" {
			fmt.Printf("│ 📝 Full Response:\n")
			for _, line := range strings.Split(text, "\n") {
				fmt.Printf("│   %s\n", line)
			}
		} else {
			// simple mode: show Thought summary only
			if thought != "" {
				fmt.Printf("│ 💭 Thought: %s\n", truncate(thought, 120))
			}
		}
	}
}

func (h *PrettyReActHandler) HandleChainError(_ context.Context, err error) {
	fmt.Printf("│ ❌ Chain Error: %v\n", err)
}

func (h *PrettyReActHandler) HandleToolStart(_ context.Context, input string) {}

func (h *PrettyReActHandler) HandleToolEnd(_ context.Context, output string) {
	// Collect observation
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = output
	}
	// Notify about observation immediately (real-time streaming)
	h.notifyStepUpdate("observation")
	h.mu.Unlock()

	// full mode: show tool output
	if h.logMode == "full" {
		fmt.Printf("│ 📤 Tool Output (Full):\n")
		for _, line := range strings.Split(output, "\n") {
			fmt.Printf("│   %s\n", line)
		}
	}
}

func (h *PrettyReActHandler) HandleToolError(_ context.Context, err error) {
	fmt.Printf("│ ❌ Tool Error: %v\n", err)
	
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = fmt.Sprintf("Error: %v", err)
	}
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	// Record current action
	h.lastAction = action.Tool

	// Collect action info
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Action = action.Tool
		h.currentStep.ActionInput = action.ToolInput
	}
	// Notify about action immediately (real-time streaming)
	h.notifyStepUpdate("action")
	h.mu.Unlock()

	fmt.Printf("│ 🎯 Action: %s\n", action.Tool)

	// full mode: show full input
	if h.logMode == "full" {
		fmt.Printf("│ 📥 Input (Full):\n")
		for _, line := range strings.Split(action.ToolInput, "\n") {
			fmt.Printf("│   %s\n", line)
		}
	} else {
		// simple mode: truncated display
		fmt.Printf("│ 📥 Input: %s\n", truncate(action.ToolInput, 100))
	}

	// Mark update_rich_context as not counting towards limit
	if action.Tool == "update_rich_context" {
		fmt.Printf("│ ℹ️  Note: This action does NOT count towards the 5-iteration limit\n")
	}
}

func (h *PrettyReActHandler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	if output, ok := finish.ReturnValues["output"].(string); ok {
		// Set the final answer as action
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Action = "Final Answer"
			h.currentStep.ActionInput = output
		}
		// Notify about final answer immediately (real-time streaming)
		h.notifyStepUpdate("finish")
		h.mu.Unlock()

		if strings.Contains(output, "agent not finished") {
			fmt.Printf("└─ ⚠️  Max iterations reached ────────────────────────────\n")
		} else {
			fmt.Printf("└─ ✅ Final Answer ──────────────────────────────────────\n")
			fmt.Printf("   %s\n", truncate(output, 150))
		}
	}
}

func (h *PrettyReActHandler) HandleRetrieverStart(_ context.Context, query string) {}

func (h *PrettyReActHandler) HandleRetrieverEnd(_ context.Context, query string, documents []schema.Document) {
}

func (h *PrettyReActHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {}

// extractThought extracts thought from LLM response text
func extractThought(text string) string {
	if idx := strings.Index(text, "Thought:"); idx >= 0 {
		thought := text[idx+8:]
		// Find Action or Final Answer position
		if actionIdx := strings.Index(thought, "Action:"); actionIdx >= 0 {
			return strings.TrimSpace(thought[:actionIdx])
		} else if finalIdx := strings.Index(thought, "Final Answer:"); finalIdx >= 0 {
			return strings.TrimSpace(thought[:finalIdx])
		}
		return strings.TrimSpace(thought)
	}
	return ""
}

// truncate truncates long text
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
