package react

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Handler implements langchaingo's callbacks interface for the ReAct loop.
// It collects steps and optionally streams them via StepCallback for SSE.
type Handler struct {
	logMode      string
	stepCallback StepCallback

	mu             sync.Mutex
	iterationCount int
	steps          []Step
	currentStep    *Step
	lastAction     string
}

// Ensure Handler implements all required callback interfaces.
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
} = &Handler{}

// GetSteps returns all collected steps. Finalizes any pending step.
func (h *Handler) GetSteps() []Step {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.finalizeCurrent()
	return h.steps
}

// --- Internal helpers ---

func (h *Handler) finalizeCurrent() {
	if h.currentStep != nil && (h.currentStep.Action != "" || h.currentStep.Thought != "") {
		h.steps = append(h.steps, *h.currentStep)
	}
	h.currentStep = nil
}

func (h *Handler) notify(eventType string) {
	if h.stepCallback != nil && h.currentStep != nil {
		h.stepCallback(*h.currentStep, eventType)
	}
}

// --- Callback implementations ---

func (h *Handler) HandleText(_ context.Context, text string) {}

func (h *Handler) HandleLLMStart(_ context.Context, prompts []string) {
	if h.logMode == "full" {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("[ReAct] LLM Prompt")
		fmt.Println(strings.Repeat("=", 60))
		for _, p := range prompts {
			fmt.Println(p)
		}
	}
}

func (h *Handler) HandleLLMGenerateContentStart(_ context.Context, _ []llms.MessageContent) {}

func (h *Handler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	if h.logMode == "full" {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("[ReAct] LLM Response")
		fmt.Println(strings.Repeat("=", 60))
		for _, c := range res.Choices {
			fmt.Println(c.Content)
		}
	}
}

func (h *Handler) HandleLLMError(_ context.Context, err error) {
	if h.logMode != "quiet" {
		fmt.Printf("[ReAct] LLM Error: %v\n", err)
	}
}

func (h *Handler) HandleChainStart(_ context.Context, _ map[string]any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.iterationCount++
	h.finalizeCurrent()
	h.currentStep = &Step{
		Iteration: h.iterationCount,
		Timestamp: time.Now(),
	}

	if h.logMode != "quiet" {
		fmt.Printf("\n┌─ ReAct Iteration %d ───────────────────────────\n", h.iterationCount)
	}
}

func (h *Handler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	if text, ok := outputs["text"].(string); ok {
		thought := extractThought(text)
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Thought = thought
		}
		h.notify("thought")
		h.mu.Unlock()

		if h.logMode != "quiet" && thought != "" {
			fmt.Printf("│ Thought: %s\n", truncate(thought, 120))
		}
	}
}

func (h *Handler) HandleChainError(_ context.Context, err error) {
	if h.logMode != "quiet" {
		fmt.Printf("│ Chain Error: %v\n", err)
	}
}

func (h *Handler) HandleToolStart(_ context.Context, _ string) {}

func (h *Handler) HandleToolEnd(_ context.Context, output string) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = output
	}
	h.notify("observation")
	h.mu.Unlock()

	if h.logMode == "full" {
		fmt.Printf("│ Observation: %s\n", truncate(output, 200))
	}
}

func (h *Handler) HandleToolError(_ context.Context, err error) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = fmt.Sprintf("Error: %v", err)
	}
	h.mu.Unlock()
	if h.logMode != "quiet" {
		fmt.Printf("│ Tool Error: %v\n", err)
	}
}

func (h *Handler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	h.lastAction = action.Tool

	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Action = action.Tool
		h.currentStep.ActionInput = action.ToolInput
	}
	h.notify("action")
	h.mu.Unlock()

	if h.logMode != "quiet" {
		fmt.Printf("│ Action: %s\n", action.Tool)
		fmt.Printf("│ Input: %s\n", truncate(action.ToolInput, 100))
	}
}

func (h *Handler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	if output, ok := finish.ReturnValues["output"].(string); ok {
		h.mu.Lock()
		if h.currentStep != nil {
			h.currentStep.Action = "Final Answer"
			h.currentStep.ActionInput = output
		}
		h.notify("finish")
		h.mu.Unlock()

		if h.logMode != "quiet" {
			if strings.Contains(output, "agent not finished") {
				fmt.Println("└─ Max iterations reached")
			} else {
				fmt.Printf("└─ Final Answer: %s\n", truncate(output, 150))
			}
		}
	}
}

func (h *Handler) HandleRetrieverStart(_ context.Context, _ string)                       {}
func (h *Handler) HandleRetrieverEnd(_ context.Context, _ string, _ []schema.Document)    {}
func (h *Handler) HandleStreamingFunc(_ context.Context, _ []byte)                        {}

// --- Utilities ---

func extractThought(text string) string {
	if idx := strings.Index(text, "Thought:"); idx >= 0 {
		thought := text[idx+8:]
		if actionIdx := strings.Index(thought, "Action:"); actionIdx >= 0 {
			return strings.TrimSpace(thought[:actionIdx])
		}
		if finalIdx := strings.Index(thought, "Final Answer:"); finalIdx >= 0 {
			return strings.TrimSpace(thought[:finalIdx])
		}
		return strings.TrimSpace(thought)
	}
	return ""
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
