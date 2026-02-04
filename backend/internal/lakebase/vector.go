// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in ReActSQL system.
package lakebase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Vector operation errors
var (
	ErrVectorDimMismatch = errors.New("lakebase: vector dimension mismatch")
	ErrEmbeddingNotFound = errors.New("lakebase: embedding not found")
	ErrInvalidVector     = errors.New("lakebase: invalid vector format")
)

// VectorRepository defines the interface for vector operations
type VectorRepository interface {
	// Embedding CRUD operations
	SaveEmbedding(ctx context.Context, emb *Embedding) (int64, error)
	SaveEmbeddingBatch(ctx context.Context, embeddings []*Embedding) error
	GetEmbedding(ctx context.Context, id int64) (*Embedding, error)
	GetEmbeddingByEntity(ctx context.Context, dsID int64, entityType EntityType, entityID int64) (*Embedding, error)
	DeleteEmbedding(ctx context.Context, id int64) error
	DeleteEmbeddingsByDatasource(ctx context.Context, dsID int64) error
	DeleteEmbeddingsByEntity(ctx context.Context, dsID int64, entityType EntityType, entityID int64) error

	// Vector search operations
	SearchSimilar(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error)
	SearchSimilarByType(ctx context.Context, dsID int64, entityType EntityType, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error)
	SearchSimilarWithThreshold(ctx context.Context, dsID int64, queryVector []float32, topK int, maxDistance float64) ([]*EmbeddingWithDistance, error)
}

// MySQLVectorRepository implements VectorRepository for MariaDB with HNSW
type MySQLVectorRepository struct {
	pool *ConnectionPool
}

// NewMySQLVectorRepository creates a new vector repository
func NewMySQLVectorRepository(pool *ConnectionPool) *MySQLVectorRepository {
	return &MySQLVectorRepository{pool: pool}
}

// vectorToString converts a float32 slice to MariaDB vector string format
// Format: [0.1, 0.2, 0.3, ...]
func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}

	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = strconv.FormatFloat(float64(val), 'f', 8, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// parseVectorFromBytes parses vector bytes from MariaDB
// MariaDB returns vectors in binary format
func parseVectorFromBytes(data []byte) ([]float32, error) {
	if len(data) == 0 {
		return nil, ErrInvalidVector
	}

	// If data starts with '[', it's in text format
	if data[0] == '[' {
		return parseVectorFromText(string(data))
	}

	// Binary format: 4 bytes per float32
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("%w: invalid binary length %d", ErrInvalidVector, len(data))
	}

	numFloats := len(data) / 4
	result := make([]float32, numFloats)
	for i := 0; i < numFloats; i++ {
		offset := i * 4
		// Little-endian float32
		bits := uint32(data[offset]) |
			uint32(data[offset+1])<<8 |
			uint32(data[offset+2])<<16 |
			uint32(data[offset+3])<<24
		result[i] = float32FromBits(bits)
	}
	return result, nil
}

// float32FromBits converts uint32 to float32
func float32FromBits(bits uint32) float32 {
	return math.Float32frombits(bits)
}

// parseVectorFromText parses vector from text format [0.1, 0.2, ...]
func parseVectorFromText(s string) ([]float32, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil, fmt.Errorf("%w: missing brackets", ErrInvalidVector)
	}

	s = s[1 : len(s)-1] // Remove brackets
	if s == "" {
		return []float32{}, nil
	}

	parts := strings.Split(s, ",")
	result := make([]float32, len(parts))
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid number %q", ErrInvalidVector, part)
		}
		result[i] = float32(val)
	}
	return result, nil
}

// SaveEmbedding saves a single embedding
func (r *MySQLVectorRepository) SaveEmbedding(ctx context.Context, emb *Embedding) (int64, error) {
	if len(emb.Embedding) != DefaultEmbeddingDimension {
		return 0, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(emb.Embedding))
	}

	vectorStr := vectorToString(emb.Embedding)
	// Use REPLACE INTO for upsert behavior
	query := `
		REPLACE INTO rc_embeddings
		(datasource_id, entity_type, entity_id, entity_text, embedding, embedding_model)
		VALUES (?, ?, ?, ?, VEC_FromText(?), ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		emb.DatasourceID, emb.EntityType, emb.EntityID, emb.EntityText, vectorStr, emb.EmbeddingModel)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to save embedding: %w", err)
	}
	return result.LastInsertId()
}

// SaveEmbeddingBatch saves multiple embeddings in a transaction
func (r *MySQLVectorRepository) SaveEmbeddingBatch(ctx context.Context, embeddings []*Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	// Validate dimensions
	for i, emb := range embeddings {
		if len(emb.Embedding) != DefaultEmbeddingDimension {
			return fmt.Errorf("%w: embedding[%d] expected %d, got %d",
				ErrVectorDimMismatch, i, DefaultEmbeddingDimension, len(emb.Embedding))
		}
	}

	// Use REPLACE INTO to handle duplicate keys (upsert behavior)
	query := `
		REPLACE INTO rc_embeddings
		(datasource_id, entity_type, entity_id, entity_text, embedding, embedding_model)
		VALUES (?, ?, ?, ?, VEC_FromText(?), ?)
	`

	return r.pool.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("lakebase: failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, emb := range embeddings {
			vectorStr := vectorToString(emb.Embedding)
			_, err := stmt.ExecContext(ctx,
				emb.DatasourceID, emb.EntityType, emb.EntityID, emb.EntityText, vectorStr, emb.EmbeddingModel)
			if err != nil {
				return fmt.Errorf("lakebase: failed to insert embedding: %w", err)
			}
		}
		return nil
	})
}

// GetEmbedding retrieves an embedding by ID
func (r *MySQLVectorRepository) GetEmbedding(ctx context.Context, id int64) (*Embedding, error) {
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       VEC_ToText(embedding) as embedding, embedding_model, created_at, updated_at
		FROM rc_embeddings WHERE id = ?
	`
	emb := &Embedding{}
	var embeddingStr string
	err := r.pool.QueryRowContext(ctx, query, id).Scan(
		&emb.ID, &emb.DatasourceID, &emb.EntityType, &emb.EntityID, &emb.EntityText,
		&embeddingStr, &emb.EmbeddingModel, &emb.CreatedAt, &emb.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrEmbeddingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get embedding: %w", err)
	}

	// Parse the vector from text format
	vec, err := parseVectorFromText(embeddingStr)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse embedding: %w", err)
	}
	emb.Embedding = vec

	return emb, nil
}

// GetEmbeddingByEntity retrieves an embedding by entity
func (r *MySQLVectorRepository) GetEmbeddingByEntity(ctx context.Context, dsID int64, entityType EntityType, entityID int64) (*Embedding, error) {
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       VEC_ToText(embedding) as embedding, embedding_model, created_at, updated_at
		FROM rc_embeddings
		WHERE datasource_id = ? AND entity_type = ? AND entity_id = ?
	`
	emb := &Embedding{}
	var embeddingStr string
	err := r.pool.QueryRowContext(ctx, query, dsID, entityType, entityID).Scan(
		&emb.ID, &emb.DatasourceID, &emb.EntityType, &emb.EntityID, &emb.EntityText,
		&embeddingStr, &emb.EmbeddingModel, &emb.CreatedAt, &emb.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrEmbeddingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get embedding by entity: %w", err)
	}

	vec, err := parseVectorFromText(embeddingStr)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse embedding: %w", err)
	}
	emb.Embedding = vec

	return emb, nil
}

// DeleteEmbedding deletes an embedding by ID
func (r *MySQLVectorRepository) DeleteEmbedding(ctx context.Context, id int64) error {
	query := `DELETE FROM rc_embeddings WHERE id = ?`
	result, err := r.pool.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("lakebase: failed to delete embedding: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrEmbeddingNotFound
	}
	return nil
}

// DeleteEmbeddingsByDatasource deletes all embeddings for a datasource
func (r *MySQLVectorRepository) DeleteEmbeddingsByDatasource(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_embeddings WHERE datasource_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

// DeleteEmbeddingsByEntity deletes embeddings for a specific entity
func (r *MySQLVectorRepository) DeleteEmbeddingsByEntity(ctx context.Context, dsID int64, entityType EntityType, entityID int64) error {
	query := `DELETE FROM rc_embeddings WHERE datasource_id = ? AND entity_type = ? AND entity_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, entityType, entityID)
	return err
}

// SearchSimilar performs HNSW vector similarity search
func (r *MySQLVectorRepository) SearchSimilar(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	if len(queryVector) != DefaultEmbeddingDimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(queryVector))
	}

	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings
		WHERE datasource_id = ?
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, dsID, topK)
}

// SearchSimilarByType performs HNSW search filtered by entity type
func (r *MySQLVectorRepository) SearchSimilarByType(ctx context.Context, dsID int64, entityType EntityType, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	if len(queryVector) != DefaultEmbeddingDimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(queryVector))
	}

	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings
		WHERE datasource_id = ? AND entity_type = ?
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, dsID, entityType, topK)
}

// SearchSimilarWithThreshold performs search with a distance threshold
func (r *MySQLVectorRepository) SearchSimilarWithThreshold(ctx context.Context, dsID int64, queryVector []float32, topK int, maxDistance float64) ([]*EmbeddingWithDistance, error) {
	if len(queryVector) != DefaultEmbeddingDimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(queryVector))
	}

	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings
		WHERE datasource_id = ?
		HAVING distance <= ?
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, dsID, maxDistance, topK)
}

func (r *MySQLVectorRepository) searchEmbeddings(ctx context.Context, query string, args ...interface{}) ([]*EmbeddingWithDistance, error) {
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to search embeddings: %w", err)
	}
	defer rows.Close()

	var results []*EmbeddingWithDistance
	for rows.Next() {
		ewd := &EmbeddingWithDistance{}
		err := rows.Scan(
			&ewd.ID, &ewd.DatasourceID, &ewd.EntityType, &ewd.EntityID, &ewd.EntityText,
			&ewd.EmbeddingModel, &ewd.CreatedAt, &ewd.UpdatedAt, &ewd.Distance)
		if err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan embedding result: %w", err)
		}
		results = append(results, ewd)
	}
	return results, rows.Err()
}

// UpdateEmbedding updates an existing embedding
func (r *MySQLVectorRepository) UpdateEmbedding(ctx context.Context, id int64, embedding []float32, entityText string) error {
	if len(embedding) != DefaultEmbeddingDimension {
		return fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(embedding))
	}

	vectorStr := vectorToString(embedding)
	query := `
		UPDATE rc_embeddings
		SET embedding = VEC_FromText(?), entity_text = ?, updated_at = NOW()
		WHERE id = ?
	`
	result, err := r.pool.ExecContext(ctx, query, vectorStr, entityText, id)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update embedding: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrEmbeddingNotFound
	}
	return nil
}

// UpsertEmbedding inserts or updates an embedding by entity
func (r *MySQLVectorRepository) UpsertEmbedding(ctx context.Context, emb *Embedding) error {
	if len(emb.Embedding) != DefaultEmbeddingDimension {
		return fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(emb.Embedding))
	}

	vectorStr := vectorToString(emb.Embedding)
	query := `
		INSERT INTO rc_embeddings
		(datasource_id, entity_type, entity_id, entity_text, embedding, embedding_model)
		VALUES (?, ?, ?, ?, VEC_FromText(?), ?)
		ON DUPLICATE KEY UPDATE
		entity_text = VALUES(entity_text),
		embedding = VALUES(embedding),
		embedding_model = VALUES(embedding_model),
		updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query,
		emb.DatasourceID, emb.EntityType, emb.EntityID, emb.EntityText, vectorStr, emb.EmbeddingModel)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert embedding: %w", err)
	}
	return nil
}

// GetEmbeddingsByDatasource retrieves all embeddings for a datasource
func (r *MySQLVectorRepository) GetEmbeddingsByDatasource(ctx context.Context, dsID int64) ([]*Embedding, error) {
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, created_at, updated_at
		FROM rc_embeddings
		WHERE datasource_id = ?
		ORDER BY entity_type, entity_id
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get embeddings by datasource: %w", err)
	}
	defer rows.Close()

	var results []*Embedding
	for rows.Next() {
		emb := &Embedding{}
		err := rows.Scan(
			&emb.ID, &emb.DatasourceID, &emb.EntityType, &emb.EntityID, &emb.EntityText,
			&emb.EmbeddingModel, &emb.CreatedAt, &emb.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan embedding: %w", err)
		}
		// Note: We don't load the embedding vector here for efficiency
		results = append(results, emb)
	}
	return results, rows.Err()
}

// CountEmbeddingsByDatasource returns the count of embeddings for a datasource
func (r *MySQLVectorRepository) CountEmbeddingsByDatasource(ctx context.Context, dsID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM rc_embeddings WHERE datasource_id = ?`
	var count int64
	err := r.pool.QueryRowContext(ctx, query, dsID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to count embeddings: %w", err)
	}
	return count, nil
}

// CountEmbeddingsByType returns the count of embeddings by type for a datasource
func (r *MySQLVectorRepository) CountEmbeddingsByType(ctx context.Context, dsID int64, entityType EntityType) (int64, error) {
	query := `SELECT COUNT(*) FROM rc_embeddings WHERE datasource_id = ? AND entity_type = ?`
	var count int64
	err := r.pool.QueryRowContext(ctx, query, dsID, entityType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to count embeddings by type: %w", err)
	}
	return count, nil
}

// Ensure MySQLVectorRepository implements VectorRepository interface
var _ VectorRepository = (*MySQLVectorRepository)(nil)
