package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// OpenAIProvider implements EmbeddingProvider using OpenAI-compatible embedding APIs.
// Works with OpenAI, Volcengine Ark (Doubao/Seed), and other compatible services.
// Supports both standard /embeddings and Volcengine multimodal /embeddings/multimodal endpoints.
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimension  int
	multimodal bool // use /embeddings/multimodal endpoint (Volcengine)
	httpClient *http.Client
}

// OpenAIConfig holds configuration for OpenAI-compatible embedding provider
type OpenAIConfig struct {
	APIKey     string
	BaseURL    string // e.g. "https://ark.cn-beijing.volces.com/api/v3"
	Model      string // e.g. "doubao-embedding-vision-250615"
	Dimension  int    // e.g. 2048
	Multimodal bool   // use multimodal endpoint
	Timeout    time.Duration
}

// NewOpenAIProvider creates a new OpenAI-compatible embedding provider
func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	}
	if config.Dimension == 0 {
		config.Dimension = 2048
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		apiKey:     config.APIKey,
		baseURL:    config.BaseURL,
		model:      config.Model,
		dimension:  config.Dimension,
		multimodal: config.Multimodal,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// =============================================
// Standard OpenAI /embeddings endpoint
// =============================================

type openAIEmbeddingRequest struct {
	Input          interface{} `json:"input"`
	Model          string      `json:"model"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
}

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

// =============================================
// Volcengine Multimodal /embeddings/multimodal
// =============================================

type multimodalInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type multimodalEmbeddingRequest struct {
	Model string            `json:"model"`
	Input []multimodalInput `json:"input"`
}

type multimodalEmbeddingResponse struct {
	Object string `json:"object"`
	Data   struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
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

	if p.multimodal {
		return p.embedBatchMultimodal(ctx, texts)
	}
	return p.embedBatchStandard(ctx, texts)
}

// embedBatchStandard uses the standard OpenAI /embeddings endpoint (batch supported)
func (p *OpenAIProvider) embedBatchStandard(ctx context.Context, texts []string) ([]Vector, error) {
	reqBody := openAIEmbeddingRequest{
		Input:          texts,
		Model:          p.model,
		EncodingFormat: "float",
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

	vectors := make([]Vector, len(texts))
	for _, item := range result.Data {
		if item.Index < len(vectors) {
			vectors[item.Index] = item.Embedding
		}
	}

	return vectors, nil
}

// embedBatchMultimodal uses Volcengine /embeddings/multimodal endpoint (one text per request)
func (p *OpenAIProvider) embedBatchMultimodal(ctx context.Context, texts []string) ([]Vector, error) {
	vectors := make([]Vector, len(texts))

	for i, text := range texts {
		vec, err := p.embedOneMultimodal(ctx, text)
		if err != nil {
			log.Printf("[Embedding] Error embedding text %d/%d: %v", i+1, len(texts), err)
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		vectors[i] = vec
	}

	return vectors, nil
}

func (p *OpenAIProvider) embedOneMultimodal(ctx context.Context, text string) (Vector, error) {
	reqBody := multimodalEmbeddingRequest{
		Model: p.model,
		Input: []multimodalInput{{Type: "text", Text: text}},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings/multimodal", bytes.NewReader(jsonData))
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

	var result multimodalEmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", result.Error.Code, result.Error.Message)
	}

	if len(result.Data.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return result.Data.Embedding, nil
}

// Dimension returns the dimension of embedding vectors
func (p *OpenAIProvider) Dimension() int {
	return p.dimension
}

// Name returns the name of the embedding provider
func (p *OpenAIProvider) Name() string {
	return fmt.Sprintf("openai/%s", p.model)
}
