// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in ATLAS system.
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
	ErrInvalidVector     = errors.New("lakebase: invalid vector format")
)

// VectorRepository defines the interface for vector operations
type VectorRepository interface {
	SaveEmbeddingBatch(ctx context.Context, embeddings []*Embedding) error
	UpsertEmbedding(ctx context.Context, emb *Embedding) error
	DeleteEmbeddingsByDatasource(ctx context.Context, dsID int64) error
	UpsertSearchDocument(ctx context.Context, doc *SearchDocument) error
	DeleteSearchDocumentsByDatasource(ctx context.Context, dsID int64) error
	SearchSimilar(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error)
	SearchSimilarByType(ctx context.Context, dsID int64, entityType EntityType, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error)
	SearchSparse(ctx context.Context, dsID int64, query string, topK int) ([]*SparseSearchResult, error)
	CountEmbeddingsByDatasource(ctx context.Context, dsID int64) (int64, error)
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

	// Use INSERT ... ON DUPLICATE KEY UPDATE for proper upsert on (datasource_id, entity_type, entity_id)
	query := `
		INSERT INTO rc_embeddings
		(datasource_id, entity_type, entity_id, entity_text, embedding, embedding_model)
		VALUES (?, ?, ?, ?, VEC_FromText(?), ?)
		ON DUPLICATE KEY UPDATE
		entity_text = VALUES(entity_text),
		embedding = VALUES(embedding),
		embedding_model = VALUES(embedding_model),
		is_stale = 0,
		is_deleted = 0,
		updated_at = NOW()
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

// DeleteEmbeddingsByDatasource deletes all embeddings for a datasource
func (r *MySQLVectorRepository) DeleteEmbeddingsByDatasource(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_embeddings WHERE datasource_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

func (r *MySQLVectorRepository) DeleteSearchDocumentsByDatasource(ctx context.Context, dsID int64) error {
	_, err := r.pool.ExecContext(ctx, `DELETE FROM rc_search_documents WHERE datasource_id = ?`, dsID)
	return err
}

func (r *MySQLVectorRepository) UpsertSearchDocument(ctx context.Context, doc *SearchDocument) error {
	query := `
		INSERT INTO rc_search_documents
		(datasource_id, entity_type, entity_id, title, body, is_deleted)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		title = VALUES(title),
		body = VALUES(body),
		is_deleted = VALUES(is_deleted),
		updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query,
		doc.DatasourceID, doc.EntityType, doc.EntityID, doc.Title, doc.Body, doc.IsDeleted)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert search document: %w", err)
	}
	return nil
}

func (r *MySQLVectorRepository) SearchSparse(ctx context.Context, dsID int64, text string, topK int) ([]*SparseSearchResult, error) {
	if topK <= 0 || strings.TrimSpace(text) == "" {
		return nil, nil
	}

	query := `
		SELECT id, datasource_id, entity_type, entity_id, title, body, is_deleted,
		       created_at, updated_at,
		       MATCH(title, body) AGAINST (? IN NATURAL LANGUAGE MODE) AS score
		FROM rc_search_documents
		WHERE datasource_id = ?
		  AND is_deleted = 0
		  AND MATCH(title, body) AGAINST (? IN NATURAL LANGUAGE MODE)
		ORDER BY score DESC
		LIMIT ?
	`
	rows, err := r.pool.QueryContext(ctx, query, text, dsID, text, topK)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to sparse search documents: %w", err)
	}
	defer rows.Close()

	var results []*SparseSearchResult
	for rows.Next() {
		res := &SparseSearchResult{}
		var updatedAt sql.NullTime
		if err := rows.Scan(&res.ID, &res.DatasourceID, &res.EntityType, &res.EntityID,
			&res.Title, &res.Body, &res.IsDeleted, &res.CreatedAt, &updatedAt, &res.Score); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan sparse result: %w", err)
		}
		if updatedAt.Valid {
			res.UpdatedAt = updatedAt.Time
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

// SearchSimilar performs hybrid vector retrieval: global HNSW overfetch first,
// then application-side datasource filtering, with scoped brute-force fallback
// when ANN candidates are insufficient after filtering.
func (r *MySQLVectorRepository) SearchSimilar(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	if len(queryVector) != DefaultEmbeddingDimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(queryVector))
	}
	if topK <= 0 {
		return nil, nil
	}

	overfetchK := topK * 10
	if overfetchK < 1000 {
		overfetchK = 1000
	}
	if overfetchK > 10000 {
		overfetchK = 10000
	}

	annResults, err := r.searchSimilarGlobalHNSW(ctx, queryVector, overfetchK)
	if err == nil {
		filtered := make([]*EmbeddingWithDistance, 0, topK)
		for _, result := range annResults {
			if result.DatasourceID != dsID || result.IsDeleted {
				continue
			}
			filtered = append(filtered, result)
			if len(filtered) >= topK {
				break
			}
		}
		minRequired := topK / 2
		if minRequired < 1 {
			minRequired = 1
		}
		if len(filtered) >= minRequired {
			return filtered, nil
		}
	}

	return r.searchSimilarScopedBruteForce(ctx, dsID, queryVector, topK)
}

func (r *MySQLVectorRepository) searchSimilarScopedBruteForce(ctx context.Context, dsID int64, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, is_stale, is_deleted, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings IGNORE INDEX (idx_embedding_hnsw)
		WHERE datasource_id = ? AND is_deleted = 0
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, dsID, topK)
}

func (r *MySQLVectorRepository) searchSimilarGlobalHNSW(ctx context.Context, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, is_stale, is_deleted, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings FORCE INDEX (idx_embedding_hnsw)
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, topK)
}

// SearchSimilarByType performs brute-force search filtered by datasource and entity type.
func (r *MySQLVectorRepository) SearchSimilarByType(ctx context.Context, dsID int64, entityType EntityType, queryVector []float32, topK int) ([]*EmbeddingWithDistance, error) {
	if len(queryVector) != DefaultEmbeddingDimension {
		return nil, fmt.Errorf("%w: expected %d, got %d", ErrVectorDimMismatch, DefaultEmbeddingDimension, len(queryVector))
	}

	vectorStr := vectorToString(queryVector)
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, is_stale, is_deleted, created_at, updated_at,
		       VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
		FROM rc_embeddings IGNORE INDEX (idx_embedding_hnsw)
		WHERE datasource_id = ? AND entity_type = ? AND is_deleted = 0
		ORDER BY distance ASC
		LIMIT ?
	`

	return r.searchEmbeddings(ctx, query, vectorStr, dsID, entityType, topK)
}

func (r *MySQLVectorRepository) searchEmbeddings(ctx context.Context, query string, args ...interface{}) ([]*EmbeddingWithDistance, error) {
	db, err := r.pool.DB()
	if err != nil {
		return nil, fmt.Errorf("lakebase: pool unavailable: %w", err)
	}

	// Acquire a single connection for query execution.
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to acquire connection: %w", err)
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to search embeddings: %w", err)
	}
	defer rows.Close()

	var results []*EmbeddingWithDistance
	for rows.Next() {
		ewd := &EmbeddingWithDistance{}
		var updatedAt sql.NullTime
		err := rows.Scan(
			&ewd.ID, &ewd.DatasourceID, &ewd.EntityType, &ewd.EntityID, &ewd.EntityText,
			&ewd.EmbeddingModel, &ewd.IsStale, &ewd.IsDeleted, &ewd.CreatedAt, &updatedAt, &ewd.Distance)
		if err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan embedding result: %w", err)
		}
		if updatedAt.Valid {
			ewd.UpdatedAt = updatedAt.Time
		}
		results = append(results, ewd)
	}
	return results, rows.Err()
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
		is_stale = 0,
		is_deleted = 0,
		updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query,
		emb.DatasourceID, emb.EntityType, emb.EntityID, emb.EntityText, vectorStr, emb.EmbeddingModel)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert embedding: %w", err)
	}
	return nil
}

// CountEmbeddingsByDatasource returns the count of embeddings for a datasource
func (r *MySQLVectorRepository) CountEmbeddingsByDatasource(ctx context.Context, dsID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM rc_embeddings WHERE datasource_id = ? AND is_deleted = 0`
	var count int64
	err := r.pool.QueryRowContext(ctx, query, dsID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to count embeddings: %w", err)
	}
	return count, nil
}

// ===========================================
// Stale / Soft-delete / Batch operations
// ===========================================

// MarkEmbeddingStale marks embeddings for a specific entity as stale (needs re-embedding)
func (r *MySQLVectorRepository) MarkEmbeddingStale(ctx context.Context, dsID int64, entityType EntityType, entityID int64) error {
	query := `UPDATE rc_embeddings SET is_stale = 1, updated_at = NOW() WHERE datasource_id = ? AND entity_type = ? AND entity_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("lakebase: failed to mark embedding stale: %w", err)
	}
	return nil
}

// MarkEmbeddingStaleByEntity marks embeddings stale by entity type and name (for table/column lookup)
func (r *MySQLVectorRepository) MarkEmbeddingStaleByEntity(ctx context.Context, dsID int64, entityType EntityType, entityText string) error {
	query := `UPDATE rc_embeddings SET is_stale = 1, updated_at = NOW() WHERE datasource_id = ? AND entity_type = ? AND entity_text LIKE ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, entityType, entityText+"%")
	if err != nil {
		return fmt.Errorf("lakebase: failed to mark embedding stale by entity: %w", err)
	}
	return nil
}

// SoftDeleteEmbedding soft-deletes embeddings for a specific entity
func (r *MySQLVectorRepository) SoftDeleteEmbedding(ctx context.Context, dsID int64, entityType EntityType, entityID int64) error {
	query := `UPDATE rc_embeddings SET is_deleted = 1, updated_at = NOW() WHERE datasource_id = ? AND entity_type = ? AND entity_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("lakebase: failed to soft delete embedding: %w", err)
	}
	return nil
}

// GetStaleEmbeddings returns all stale, non-deleted embeddings for a datasource
func (r *MySQLVectorRepository) GetStaleEmbeddings(ctx context.Context, dsID int64) ([]*Embedding, error) {
	query := `
		SELECT id, datasource_id, entity_type, entity_id, entity_text,
		       embedding_model, is_stale, is_deleted, created_at, updated_at
		FROM rc_embeddings
		WHERE datasource_id = ? AND is_stale = 1 AND is_deleted = 0
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get stale embeddings: %w", err)
	}
	defer rows.Close()

	var embeddings []*Embedding
	for rows.Next() {
		e := &Embedding{}
		var updatedAt sql.NullTime
		if err := rows.Scan(&e.ID, &e.DatasourceID, &e.EntityType, &e.EntityID, &e.EntityText,
			&e.EmbeddingModel, &e.IsStale, &e.IsDeleted, &e.CreatedAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan stale embedding: %w", err)
		}
		if updatedAt.Valid {
			e.UpdatedAt = updatedAt.Time
		}
		embeddings = append(embeddings, e)
	}
	return embeddings, rows.Err()
}

// PurgeDeletedEmbeddings permanently removes soft-deleted embeddings
func (r *MySQLVectorRepository) PurgeDeletedEmbeddings(ctx context.Context, dsID int64) (int64, error) {
	query := `DELETE FROM rc_embeddings WHERE datasource_id = ? AND is_deleted = 1`
	result, err := r.pool.ExecContext(ctx, query, dsID)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to purge deleted embeddings: %w", err)
	}
	return result.RowsAffected()
}

// ClearStaleFlag clears the stale flag on embeddings after re-embedding
func (r *MySQLVectorRepository) ClearStaleFlag(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(`UPDATE rc_embeddings SET is_stale = 0, updated_at = NOW() WHERE id IN (%s)`,
		strings.Join(placeholders, ","))
	_, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("lakebase: failed to clear stale flag: %w", err)
	}
	return nil
}

// Ensure MySQLVectorRepository implements VectorRepository interface
var _ VectorRepository = (*MySQLVectorRepository)(nil)
