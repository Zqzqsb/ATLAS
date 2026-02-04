package embedding

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"math"
)

// MockEmbeddingProvider provides deterministic mock embeddings for testing
// It generates consistent embeddings based on text hash
type MockEmbeddingProvider struct {
	dim int
}

// NewMockEmbeddingProvider creates a new mock embedding provider
func NewMockEmbeddingProvider(dimension int) *MockEmbeddingProvider {
	if dimension <= 0 {
		dimension = 768 // default dimension
	}
	return &MockEmbeddingProvider{dim: dimension}
}

// Embed generates a deterministic mock embedding based on text hash
// Similar texts will have similar embeddings due to hash locality
func (p *MockEmbeddingProvider) Embed(ctx context.Context, text string) (Vector, error) {
	// Generate MD5 hash of text
	hash := md5.Sum([]byte(text))

	// Use hash bytes to seed pseudo-random embedding
	embedding := make(Vector, p.dim)

	// Generate embedding values from hash
	for i := 0; i < p.dim; i++ {
		// Use different parts of hash for different dimensions
		byteIdx := i % 16
		// Combine multiple hash bytes for variation
		seed := float64(hash[byteIdx]) + float64(hash[(byteIdx+1)%16])/256.0

		// Generate value between -1 and 1
		value := math.Sin(seed*float64(i+1)*0.1) * 0.5

		// Add some variation based on text length
		value += math.Cos(float64(len(text))*float64(i)*0.01) * 0.3

		embedding[i] = float32(value)
	}

	// Normalize the embedding
	var norm float32
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding, nil
}

// EmbedBatch generates mock embeddings for multiple texts
func (p *MockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([]Vector, error) {
	embeddings := make([]Vector, len(texts))
	for i, text := range texts {
		emb, err := p.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// Dimension returns the embedding dimension
func (p *MockEmbeddingProvider) Dimension() int {
	return p.dim
}

// Name returns the provider name
func (p *MockEmbeddingProvider) Name() string {
	return "mock"
}

// hashToFloat converts a byte slice to a float32 value between -1 and 1
func hashToFloat(data []byte) float32 {
	if len(data) < 4 {
		return 0
	}
	u := binary.LittleEndian.Uint32(data)
	// Convert to float in range [-1, 1]
	return float32(u)/float32(math.MaxUint32)*2 - 1
}
