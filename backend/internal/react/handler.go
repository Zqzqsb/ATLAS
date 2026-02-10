package react

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	"lucid/internal/logger"
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

// NewHandler creates a Handler with an optional step callback.
func NewHandler(callback StepCallback) *Handler {
	return &Handler{
		logMode:      "simple",
		stepCallback: callback,
	}
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
	log := logger.With("component", "react_handler")
	for i, p := range prompts {
		log.Debug("[LLM] Prompt sent", "index", i, "prompt_length", len(p), "prompt_preview", truncate(p, 800))
	}
}

func (h *Handler) HandleLLMGenerateContentStart(_ context.Context, _ []llms.MessageContent) {}

func (h *Handler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	log := logger.With("component", "react_handler")
	for i, c := range res.Choices {
		log.Debug("[LLM] Response received", "choice", i, "content_length", len(c.Content), "content_preview", truncate(c.Content, 800))
	}
}

func (h *Handler) HandleLLMError(_ context.Context, err error) {
	log := logger.With("component", "react_handler")
	log.Error("LLM error", "error", err)
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

	log := logger.With("component", "react_handler")
	log.Info("ReAct iteration started", "iteration", h.iterationCount)
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

		if thought != "" {
			log := logger.With("component", "react_handler")
			log.Info("ReAct thought", "iteration", h.currentStep.Iteration, "thought", truncate(thought, 200))
		}
	}
}

func (h *Handler) HandleChainError(_ context.Context, err error) {
	log := logger.With("component", "react_handler")
	log.Error("chain error", "error", err)
}

func (h *Handler) HandleToolStart(_ context.Context, _ string) {}

func (h *Handler) HandleToolEnd(_ context.Context, output string) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = output
	}
	h.notify("observation")
	h.mu.Unlock()

	log := logger.With("component", "react_handler")
	log.Debug("Tool observation (HandleToolEnd)", "output_length", len(output), "output_preview", truncate(output, 500))
}

// InjectObservation pushes a tool observation into the current step and fires
// the SSE callback. Call this from tools directly when langchaingo's
// HandleToolEnd is not reliably invoked.
func (h *Handler) InjectObservation(output string) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = output
	}
	h.notify("observation")
	h.mu.Unlock()

	log := logger.With("component", "react_handler")
	log.Debug("Tool observation (injected)", "output_length", len(output), "output_preview", truncate(output, 500))
}

func (h *Handler) HandleToolError(_ context.Context, err error) {
	h.mu.Lock()
	if h.currentStep != nil {
		h.currentStep.Observation = fmt.Sprintf("Error: %v", err)
	}
	h.mu.Unlock()
	log := logger.With("component", "react_handler")
	log.Error("tool error", "error", err)
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

	log := logger.With("component", "react_handler")
	log.Info("ReAct action",
		"tool", action.Tool,
		"input", truncate(action.ToolInput, 200),
	)
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

		log := logger.With("component", "react_handler")
		if strings.Contains(output, "agent not finished") {
			log.Warn("max iterations reached")
		} else {
			log.Info("final answer", "output", truncate(output, 200))
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
