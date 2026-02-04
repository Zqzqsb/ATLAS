package embedding

import (
	"context"
)

// Vector represents an embedding vector
type Vector []float32

// EmbeddingProvider defines the interface for text embedding services
type EmbeddingProvider interface {
	// Embed converts a single text to an embedding vector
	Embed(ctx context.Context, text string) (Vector, error)

	// EmbedBatch converts multiple texts to embedding vectors
	EmbedBatch(ctx context.Context, texts []string) ([]Vector, error)

	// Dimension returns the dimension of embedding vectors
	Dimension() int

	// Name returns the name of the embedding provider
	Name() string
}

// CosineSimilarity calculates the cosine similarity between two vectors
func CosineSimilarity(a, b Vector) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt is a simple square root implementation for float32
func sqrt(x float32) float32 {
	if x <= 0 {
		return 0
	}
	// Newton's method
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// NormalizeVector normalizes a vector to unit length
func NormalizeVector(v Vector) Vector {
	var norm float32
	for _, val := range v {
		norm += val * val
	}
	norm = sqrt(norm)
	if norm == 0 {
		return v
	}

	result := make(Vector, len(v))
	for i, val := range v {
		result[i] = val / norm
	}
	return result
}
