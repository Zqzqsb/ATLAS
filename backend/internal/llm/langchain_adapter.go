package llm

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

// LangChainAdapter adapts langchaingo LLM to our Client interface
type LangChainAdapter struct {
	llm   llms.Model
	model string
}

// NewLangChainAdapter creates a new adapter for langchaingo LLM
func NewLangChainAdapter(llm llms.Model, model string) *LangChainAdapter {
	return &LangChainAdapter{
		llm:   llm,
		model: model,
	}
}

// Complete generates a completion using the underlying langchaingo LLM
func (a *LangChainAdapter) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Build prompt from messages
	var prompt string
	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			prompt += msg.Content + "\n\n"
		case "user":
			prompt += msg.Content + "\n"
		case "assistant":
			prompt += msg.Content + "\n"
		}
	}

	// Build call options
	var opts []llms.CallOption
	if req.Temperature > 0 {
		opts = append(opts, llms.WithTemperature(req.Temperature))
	}
	if req.MaxTokens > 0 {
		opts = append(opts, llms.WithMaxTokens(req.MaxTokens))
	}

	// Call the underlying LLM
	response, err := a.llm.Call(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	return &CompletionResponse{
		Content: response,
		Model:   a.model,
	}, nil
}

// GetModel returns the model name
func (a *LangChainAdapter) GetModel() string {
	return a.model
}
