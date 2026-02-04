package llm

import (
	"context"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Model       string    `json:"model,omitempty"`
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	Usage        struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Client defines the interface for LLM providers
type Client interface {
	// Complete generates a completion for the given request
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	// GetModel returns the model name
	GetModel() string
}
