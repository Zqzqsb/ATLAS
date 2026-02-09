package inference

import (
	"context"
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
	iterationCount          int
	effectiveIterationCount int
	lastAction              string

	mu             sync.Mutex
	collectedSteps []CollectedStep
	currentStep    *CollectedStep
	stepNotifier   StepNotifier
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

func (h *PrettyReActHandler) HandleLLMStart(_ context.Context, _ []string) {}

func (h *PrettyReActHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
}

func (h *PrettyReActHandler) HandleLLMGenerateContentEnd(_ context.Context, _ *llms.ContentResponse) {}

func (h *PrettyReActHandler) HandleLLMError(_ context.Context, _ error) {}

func (h *PrettyReActHandler) HandleChainStart(_ context.Context, _ map[string]any) {
	h.iterationCount++
	if h.lastAction != "" {
		h.effectiveIterationCount++
	}

	h.mu.Lock()
	h.finalizeCurrentStep()
	h.currentStep = &CollectedStep{
		Step:      h.iterationCount,
		Timestamp: time.Now(),
	}
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	if text, ok := outputs["text"].(string); ok {
		thought := extractThought(text)
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Thought = thought
		}
		h.notifyStepUpdate("thought")
		h.mu.Unlock()
	}
}

func (h *PrettyReActHandler) HandleChainError(_ context.Context, _ error) {}

func (h *PrettyReActHandler) HandleToolStart(_ context.Context, input string) {}

func (h *PrettyReActHandler) HandleToolEnd(_ context.Context, output string) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = output
	}
	h.notifyStepUpdate("observation")
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleToolError(_ context.Context, err error) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = "Error: " + err.Error()
	}
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	h.lastAction = action.Tool

	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Action = action.Tool
		h.currentStep.ActionInput = action.ToolInput
	}
	h.notifyStepUpdate("action")
	h.mu.Unlock()
}

func (h *PrettyReActHandler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	if output, ok := finish.ReturnValues["output"].(string); ok {
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Action = "Final Answer"
			h.currentStep.ActionInput = output
		}
		h.notifyStepUpdate("finish")
		h.mu.Unlock()
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
