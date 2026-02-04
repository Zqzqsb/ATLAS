package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements EmbeddingProvider using OpenAI API
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimension  int
	httpClient *http.Client
}

// OpenAIConfig holds configuration for OpenAI embedding provider
type OpenAIConfig struct {
	APIKey    string
	BaseURL   string // Default: https://api.openai.com/v1
	Model     string // Default: text-embedding-3-small
	Dimension int    // Default: 1536
	Timeout   time.Duration
}

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "text-embedding-3-small"
	}
	if config.Dimension == 0 {
		config.Dimension = 1536
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		apiKey:    config.APIKey,
		baseURL:   config.BaseURL,
		model:     config.Model,
		dimension: config.Dimension,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// openAIEmbeddingRequest represents the request body for OpenAI embedding API
type openAIEmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

// openAIEmbeddingResponse represents the response from OpenAI embedding API
type openAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Embed converts a single text to an embedding vector
func (p *OpenAIProvider) Embed(ctx context.Context, text string) (Vector, error) {
	vectors, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return vectors[0], nil
}

// EmbedBatch converts multiple texts to embedding vectors
func (p *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := openAIEmbeddingRequest{
		Input: texts,
		Model: p.model,
	}

	// For text-embedding-3-* models, can specify dimensions
	if p.dimension > 0 && (p.model == "text-embedding-3-small" || p.model == "text-embedding-3-large") {
		reqBody.Dimensions = p.dimension
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result openAIEmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Sort by index to ensure correct order
	vectors := make([]Vector, len(texts))
	for _, item := range result.Data {
		if item.Index < len(vectors) {
			vectors[item.Index] = item.Embedding
		}
	}

	return vectors, nil
}

// Dimension returns the dimension of embedding vectors
func (p *OpenAIProvider) Dimension() int {
	return p.dimension
}

// Name returns the name of the embedding provider
func (p *OpenAIProvider) Name() string {
	return fmt.Sprintf("openai/%s", p.model)
}

// ============================================
// Local/Fallback Provider (Simple TF-IDF style)
// ============================================

// LocalProvider implements a simple local embedding provider
// Uses character n-grams for basic text similarity (no external API needed)
type LocalProvider struct {
	dimension int
	ngramSize int
}

// LocalConfig holds configuration for local embedding provider
type LocalConfig struct {
	Dimension int // Default: 384
	NgramSize int // Default: 3
}

// NewLocalProvider creates a new local embedding provider
func NewLocalProvider(config LocalConfig) *LocalProvider {
	if config.Dimension == 0 {
		config.Dimension = 384
	}
	if config.NgramSize == 0 {
		config.NgramSize = 3
	}

	return &LocalProvider{
		dimension: config.Dimension,
		ngramSize: config.NgramSize,
	}
}

// Embed converts a single text to an embedding vector using character n-grams
func (p *LocalProvider) Embed(ctx context.Context, text string) (Vector, error) {
	vector := make(Vector, p.dimension)

	// Generate character n-grams
	runes := []rune(text)
	for i := 0; i <= len(runes)-p.ngramSize; i++ {
		ngram := string(runes[i : i+p.ngramSize])
		// Hash the n-gram to get a bucket index
		hash := hashString(ngram)
		idx := int(hash % uint64(p.dimension))
		vector[idx] += 1.0
	}

	// Also add word-level features
	words := splitWords(text)
	for _, word := range words {
		hash := hashString(word)
		idx := int(hash % uint64(p.dimension))
		vector[idx] += 0.5
	}

	// Normalize
	return NormalizeVector(vector), nil
}

// EmbedBatch converts multiple texts to embedding vectors
func (p *LocalProvider) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	vectors := make([]Vector, len(texts))
	for i, text := range texts {
		v, err := p.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		vectors[i] = v
	}
	return vectors, nil
}

// Dimension returns the dimension of embedding vectors
func (p *LocalProvider) Dimension() int {
	return p.dimension
}

// Name returns the name of the embedding provider
func (p *LocalProvider) Name() string {
	return "local/ngram"
}

// hashString computes a simple hash for a string
func hashString(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// splitWords splits text into words (simple space-based split)
func splitWords(text string) []string {
	var words []string
	var current []rune

	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == ',' || r == '.' || r == '(' || r == ')' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
		} else {
			current = append(current, r)
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}
