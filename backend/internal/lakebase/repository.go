// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in ATLAS system.
package lakebase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Repository errors
var (
	ErrDatasourceNotFound = errors.New("lakebase: datasource not found")
	ErrContextNotFound    = errors.New("lakebase: context not found")
	ErrInvalidInput       = errors.New("lakebase: invalid input")
)

// Repository defines the interface for lake-base storage operations
type Repository interface {
	// Datasource operations
	CreateDatasource(ctx context.Context, ds *Datasource) (int64, error)
	GetDatasource(ctx context.Context, id int64) (*Datasource, error)
	GetDatasourceByName(ctx context.Context, name string) (*Datasource, error)
	ListDatasources(ctx context.Context) ([]*Datasource, error)
	UpdateDatasource(ctx context.Context, ds *Datasource) error
	DeleteDatasource(ctx context.Context, id int64) error
	UpdateDatasourceLastSync(ctx context.Context, id int64) error

	// Rich Context tables (rc_tables, rc_columns)
	UpsertTable(ctx context.Context, dsID int64, tableName string, rowCount int64) error
	UpsertColumn(ctx context.Context, dsID int64, tableName, columnName, dataType string, isNullable, isPK, isFK bool) error
	GetTablesByDatasource(ctx context.Context, dsID int64) ([]*TableInfo, error)
	GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*ColumnInfo, error)
	GetColumnsByTable(ctx context.Context, dsID int64, tableName string) ([]*ColumnInfo, error)

	// Business context operations
	SaveBusinessContext(ctx context.Context, bc *BusinessContext) (int64, error)
	SaveBusinessContextBatch(ctx context.Context, contexts []*BusinessContext) error
	GetContextByID(ctx context.Context, id int64) (*BusinessContext, error)
	GetContextByDatasource(ctx context.Context, dsID int64) ([]*BusinessContext, error)
	GetContextByTable(ctx context.Context, dsID int64, tableName string) ([]*BusinessContext, error)
	MarkContextExpired(ctx context.Context, ids []int64, reason string) error
	UpdateContextVersion(ctx context.Context, id int64, content json.RawMessage, updatedBy string, reason string) error
	DeleteExpiredContext(ctx context.Context, dsID int64) error
	PruneAllContext(ctx context.Context, dsID int64) error

	// Change log operations
	CreateChangeLog(ctx context.Context, log *ChangeLog) (int64, error)
	GetChangeLogsByDatasource(ctx context.Context, dsID int64, limit int) ([]*ChangeLog, error)
}

// MySQLRepository implements Repository for MySQL/MariaDB
type MySQLRepository struct {
	pool *ConnectionPool
}

// NewMySQLRepository creates a new MySQL repository
func NewMySQLRepository(pool *ConnectionPool) *MySQLRepository {
	return &MySQLRepository{pool: pool}
}

// ===========================================
// Datasource operations
// ===========================================

func (r *MySQLRepository) CreateDatasource(ctx context.Context, ds *Datasource) (int64, error) {
	query := `
		INSERT INTO rc_datasources (name, db_type, host, port, username, db_name, description, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		ds.Name, ds.DBType, ds.Host, ds.Port, ds.Username, ds.DatabaseName, ds.Description, ds.Status)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to create datasource: %w", err)
	}
	return result.LastInsertId()
}

func (r *MySQLRepository) GetDatasource(ctx context.Context, id int64) (*Datasource, error) {
	query := `
		SELECT id, name, db_type, host, port, username, db_name, description, status,
		       last_sync_at, created_at, updated_at
		FROM rc_datasources WHERE id = ?
	`
	ds := &Datasource{}
	err := r.pool.QueryRowContext(ctx, query, id).Scan(
		&ds.ID, &ds.Name, &ds.DBType, &ds.Host, &ds.Port, &ds.Username,
		&ds.DatabaseName, &ds.Description, &ds.Status, &ds.LastSyncAt, &ds.CreatedAt, &ds.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrDatasourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get datasource: %w", err)
	}
	return ds, nil
}

func (r *MySQLRepository) GetDatasourceByName(ctx context.Context, name string) (*Datasource, error) {
	query := `
		SELECT id, name, db_type, host, port, username, db_name, description, status,
		       last_sync_at, created_at, updated_at
		FROM rc_datasources WHERE name = ?
	`
	ds := &Datasource{}
	err := r.pool.QueryRowContext(ctx, query, name).Scan(
		&ds.ID, &ds.Name, &ds.DBType, &ds.Host, &ds.Port, &ds.Username,
		&ds.DatabaseName, &ds.Description, &ds.Status, &ds.LastSyncAt, &ds.CreatedAt, &ds.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrDatasourceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get datasource by name: %w", err)
	}
	return ds, nil
}

func (r *MySQLRepository) ListDatasources(ctx context.Context) ([]*Datasource, error) {
	query := `
		SELECT id, name, db_type, host, port, username, db_name, description, status,
		       last_sync_at, created_at, updated_at
		FROM rc_datasources ORDER BY created_at DESC
	`
	rows, err := r.pool.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to list datasources: %w", err)
	}
	defer rows.Close()

	var datasources []*Datasource
	for rows.Next() {
		ds := &Datasource{}
		if err := rows.Scan(
			&ds.ID, &ds.Name, &ds.DBType, &ds.Host, &ds.Port, &ds.Username,
			&ds.DatabaseName, &ds.Description, &ds.Status, &ds.LastSyncAt, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan datasource: %w", err)
		}
		datasources = append(datasources, ds)
	}
	return datasources, rows.Err()
}

func (r *MySQLRepository) UpdateDatasource(ctx context.Context, ds *Datasource) error {
	query := `
		UPDATE rc_datasources
		SET name = ?, db_type = ?, host = ?, port = ?, username = ?,
		    db_name = ?, description = ?, status = ?
		WHERE id = ?
	`
	result, err := r.pool.ExecContext(ctx, query,
		ds.Name, ds.DBType, ds.Host, ds.Port, ds.Username, ds.DatabaseName, ds.Description, ds.Status, ds.ID)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update datasource: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrDatasourceNotFound
	}
	return nil
}

func (r *MySQLRepository) DeleteDatasource(ctx context.Context, id int64) error {
	query := `DELETE FROM rc_datasources WHERE id = ?`
	result, err := r.pool.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("lakebase: failed to delete datasource: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrDatasourceNotFound
	}
	return nil
}

func (r *MySQLRepository) UpdateDatasourceLastSync(ctx context.Context, id int64) error {
	query := `UPDATE rc_datasources SET last_sync_at = NOW() WHERE id = ?`
	_, err := r.pool.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update last sync: %w", err)
	}
	return nil
}

// ===========================================
// Rich Context tables (rc_tables, rc_columns)
// ===========================================

func (r *MySQLRepository) GetTablesByDatasource(ctx context.Context, dsID int64) ([]*TableInfo, error) {
	query := `
		SELECT id, datasource_id, table_name, description, row_count, is_expired,
		       source, confidence, created_at, updated_at
		FROM rc_tables WHERE datasource_id = ?
		ORDER BY table_name
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []*TableInfo
	for rows.Next() {
		t := &TableInfo{}
		if err := rows.Scan(
			&t.ID, &t.DatasourceID, &t.TableName, &t.Description, &t.RowCount,
			&t.IsExpired, &t.Source, &t.Confidence, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (r *MySQLRepository) GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*ColumnInfo, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, data_type, description,
		       sample_values, synonyms, is_nullable, is_primary_key, is_foreign_key,
		       is_expired, source, confidence, created_at, updated_at
		FROM rc_columns WHERE datasource_id = ?
		ORDER BY table_name, id
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []*ColumnInfo
	for rows.Next() {
		c := &ColumnInfo{}
		if err := rows.Scan(
			&c.ID, &c.DatasourceID, &c.TableName, &c.ColumnName, &c.DataType,
			&c.Description, &c.SampleValues, &c.Synonyms, &c.IsNullable,
			&c.IsPrimaryKey, &c.IsForeignKey, &c.IsExpired, &c.Source,
			&c.Confidence, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan column: %w", err)
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (r *MySQLRepository) GetColumnsByTable(ctx context.Context, dsID int64, tableName string) ([]*ColumnInfo, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, data_type, description,
		       sample_values, synonyms, is_nullable, is_primary_key, is_foreign_key,
		       is_expired, source, confidence, created_at, updated_at
		FROM rc_columns WHERE datasource_id = ? AND table_name = ?
		ORDER BY id
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID, tableName)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get columns for table: %w", err)
	}
	defer rows.Close()

	var columns []*ColumnInfo
	for rows.Next() {
		c := &ColumnInfo{}
		if err := rows.Scan(
			&c.ID, &c.DatasourceID, &c.TableName, &c.ColumnName, &c.DataType,
			&c.Description, &c.SampleValues, &c.Synonyms, &c.IsNullable,
			&c.IsPrimaryKey, &c.IsForeignKey, &c.IsExpired, &c.Source,
			&c.Confidence, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan column: %w", err)
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

// GetTermsByDatasource retrieves all business terms for a datasource
func (r *MySQLRepository) GetTermsByDatasource(ctx context.Context, dsID int64) ([]*TermInfo, error) {
	query := `
		SELECT id, datasource_id, term, definition, synonyms, examples, category, created_at, updated_at
		FROM rc_terms WHERE datasource_id = ?
		ORDER BY term
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get terms: %w", err)
	}
	defer rows.Close()

	var terms []*TermInfo
	for rows.Next() {
		t := &TermInfo{}
		if err := rows.Scan(
			&t.ID, &t.DatasourceID, &t.Term, &t.Definition, &t.Synonyms,
			&t.Examples, &t.Category, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan term: %w", err)
		}
		terms = append(terms, t)
	}
	return terms, rows.Err()
}

// ===========================================
// Schema Sync (INSERT / UPSERT for rc_tables, rc_columns, rc_relations)
// ===========================================

// UpsertTable inserts a table record or updates row_count if it already exists.
// description is NOT overwritten on conflict — it is managed by Context generation.
func (r *MySQLRepository) UpsertTable(ctx context.Context, dsID int64, tableName string, rowCount int64) error {
	query := `
		INSERT INTO rc_tables (datasource_id, table_name, row_count, source, confidence, created_at, updated_at)
		VALUES (?, ?, ?, 'catalog', 0.0, NOW(), NOW())
		ON DUPLICATE KEY UPDATE row_count = VALUES(row_count), updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query, dsID, tableName, rowCount)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert table: %w", err)
	}
	return nil
}

// UpsertColumn inserts a column record or updates its physical schema fields.
// description is NOT overwritten on conflict — it is managed by Context generation.
func (r *MySQLRepository) UpsertColumn(ctx context.Context, dsID int64, tableName, columnName, dataType string, isNullable, isPK, isFK bool) error {
	query := `
		INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, is_nullable, is_primary_key, is_foreign_key, source, confidence, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'catalog', 0.0, NOW(), NOW())
		ON DUPLICATE KEY UPDATE data_type = VALUES(data_type), is_nullable = VALUES(is_nullable),
			is_primary_key = VALUES(is_primary_key), is_foreign_key = VALUES(is_foreign_key), updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query, dsID, tableName, columnName, dataType, isNullable, isPK, isFK)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert column: %w", err)
	}
	return nil
}

// UpsertRelation inserts a foreign key relation or updates on conflict.
func (r *MySQLRepository) UpsertRelation(ctx context.Context, dsID int64, fromTable, fromColumn, toTable, toColumn string) error {
	query := `
		INSERT INTO rc_relations (datasource_id, from_table, from_column, to_table, to_column, relation_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'foreign_key', NOW(), NOW())
		ON DUPLICATE KEY UPDATE updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query, dsID, fromTable, fromColumn, toTable, toColumn)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert relation: %w", err)
	}
	return nil
}

// UpdateTableDescription updates the description for a table
func (r *MySQLRepository) UpdateTableDescription(ctx context.Context, dsID int64, tableName, description, source string, confidence float64) error {
	query := `
		UPDATE rc_tables 
		SET description = ?, source = ?, confidence = ?, updated_at = NOW()
		WHERE datasource_id = ? AND table_name = ?
	`
	_, err := r.pool.ExecContext(ctx, query, description, source, confidence, dsID, tableName)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update table description: %w", err)
	}
	return nil
}

// UpdateColumnDescription updates the description for a column
func (r *MySQLRepository) UpdateColumnDescription(ctx context.Context, dsID int64, tableName, columnName, description, source string, confidence float64) error {
	query := `
		UPDATE rc_columns 
		SET description = ?, source = ?, confidence = ?, updated_at = NOW()
		WHERE datasource_id = ? AND table_name = ? AND column_name = ?
	`
	_, err := r.pool.ExecContext(ctx, query, description, source, confidence, dsID, tableName, columnName)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update column description: %w", err)
	}
	return nil
}

// UpdateColumnSynonyms updates synonyms for a column
func (r *MySQLRepository) UpdateColumnSynonyms(ctx context.Context, dsID int64, tableName, columnName, synonyms string) error {
	query := `
		UPDATE rc_columns 
		SET synonyms = ?, updated_at = NOW()
		WHERE datasource_id = ? AND table_name = ? AND column_name = ?
	`
	_, err := r.pool.ExecContext(ctx, query, synonyms, dsID, tableName, columnName)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update column synonyms: %w", err)
	}
	return nil
}

// UpdateColumnSampleValues updates sample values for a column
func (r *MySQLRepository) UpdateColumnSampleValues(ctx context.Context, dsID int64, tableName, columnName, sampleValues string) error {
	query := `
		UPDATE rc_columns 
		SET sample_values = ?, updated_at = NOW()
		WHERE datasource_id = ? AND table_name = ? AND column_name = ?
	`
	_, err := r.pool.ExecContext(ctx, query, sampleValues, dsID, tableName, columnName)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update column sample_values: %w", err)
	}
	return nil
}

// UpsertTerm inserts or updates a business term
func (r *MySQLRepository) UpsertTerm(ctx context.Context, dsID int64, term, definition, synonyms, examples, category string) error {
	query := `
		INSERT INTO rc_terms (datasource_id, term, definition, synonyms, examples, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE definition = VALUES(definition), synonyms = VALUES(synonyms),
			examples = VALUES(examples), category = VALUES(category), updated_at = NOW()
	`
	_, err := r.pool.ExecContext(ctx, query, dsID, term, definition, synonyms, examples, category)
	if err != nil {
		return fmt.Errorf("lakebase: failed to upsert term: %w", err)
	}
	return nil
}

// ===========================================
// Business context operations
// ===========================================

func (r *MySQLRepository) SaveBusinessContext(ctx context.Context, bc *BusinessContext) (int64, error) {
	query := `
		INSERT INTO rc_business_context
		(datasource_id, table_name, column_name, context_type, content, source, confidence,
		 is_expired, expires_at, version, created_by, updated_by, update_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		bc.DatasourceID, bc.TableName, bc.ColumnName, bc.ContextType, bc.Content,
		bc.Source, bc.Confidence, bc.IsExpired, bc.ExpiresAt, bc.Version,
		bc.CreatedBy, bc.UpdatedBy, bc.UpdateReason)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to save business context: %w", err)
	}
	return result.LastInsertId()
}

func (r *MySQLRepository) SaveBusinessContextBatch(ctx context.Context, contexts []*BusinessContext) error {
	if len(contexts) == 0 {
		return nil
	}

	query := `
		INSERT INTO rc_business_context
		(datasource_id, table_name, column_name, context_type, content, source, confidence,
		 is_expired, expires_at, version, created_by, updated_by, update_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	return r.pool.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("lakebase: failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, bc := range contexts {
			_, err := stmt.ExecContext(ctx,
				bc.DatasourceID, bc.TableName, bc.ColumnName, bc.ContextType, bc.Content,
				bc.Source, bc.Confidence, bc.IsExpired, bc.ExpiresAt, bc.Version,
				bc.CreatedBy, bc.UpdatedBy, bc.UpdateReason)
			if err != nil {
				return fmt.Errorf("lakebase: failed to insert business context: %w", err)
			}
		}
		return nil
	})
}

func (r *MySQLRepository) GetContextByID(ctx context.Context, id int64) (*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE id = ?
	`
	bc := &BusinessContext{}
	err := r.pool.QueryRowContext(ctx, query, id).Scan(
		&bc.ID, &bc.DatasourceID, &bc.TableName, &bc.ColumnName, &bc.ContextType, &bc.Content,
		&bc.Source, &bc.Confidence, &bc.IsExpired, &bc.ExpiresAt, &bc.Version,
		&bc.CreatedAt, &bc.UpdatedAt, &bc.CreatedBy, &bc.UpdatedBy, &bc.UpdateReason)
	if err == sql.ErrNoRows {
		return nil, ErrContextNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get context by id: %w", err)
	}
	return bc, nil
}

func (r *MySQLRepository) GetContextByDatasource(ctx context.Context, dsID int64) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE datasource_id = ? AND is_expired = 0
		ORDER BY table_name, column_name, context_type
	`
	return r.queryBusinessContext(ctx, query, dsID)
}

// GetExpiredContextByDatasource returns only expired context for a datasource.
func (r *MySQLRepository) GetExpiredContextByDatasource(ctx context.Context, dsID int64) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE datasource_id = ? AND is_expired = 1
		ORDER BY table_name, column_name, context_type
	`
	return r.queryBusinessContext(ctx, query, dsID)
}

func (r *MySQLRepository) GetContextByTable(ctx context.Context, dsID int64, tableName string) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE datasource_id = ? AND table_name = ? AND is_expired = 0
		ORDER BY column_name, context_type
	`
	return r.queryBusinessContext(ctx, query, dsID, tableName)
}

func (r *MySQLRepository) MarkContextExpired(ctx context.Context, ids []int64, reason string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = reason
	for i, id := range ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE rc_business_context
		SET is_expired = 1, update_reason = ?
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("lakebase: failed to mark context expired: %w", err)
	}
	return nil
}

func (r *MySQLRepository) UpdateContextVersion(ctx context.Context, id int64, content json.RawMessage, updatedBy string, reason string) error {
	query := `
		UPDATE rc_business_context
		SET content = ?, version = version + 1, updated_by = ?, update_reason = ?, is_expired = 0
		WHERE id = ?
	`
	result, err := r.pool.ExecContext(ctx, query, content, updatedBy, reason, id)
	if err != nil {
		return fmt.Errorf("lakebase: failed to update context version: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrContextNotFound
	}
	return nil
}

func (r *MySQLRepository) DeleteExpiredContext(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_business_context WHERE datasource_id = ? AND is_expired = 1`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

// PruneAllContext clears all AI-generated rich context data for a datasource
// This clears descriptions in rc_tables and rc_columns, and deletes rc_terms, rc_relations, rc_business_context, rc_change_log
// It preserves the basic schema metadata (table names, column names, data types) for regeneration
func (r *MySQLRepository) PruneAllContext(ctx context.Context, dsID int64) error {
	// Use transaction to ensure atomicity
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("lakebase: failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear descriptions in rc_tables (keep schema metadata)
	if _, err := tx.ExecContext(ctx, "UPDATE rc_tables SET description = NULL, updated_at = NOW() WHERE datasource_id = ?", dsID); err != nil {
		return fmt.Errorf("lakebase: failed to clear rc_tables descriptions: %w", err)
	}

	// Clear all RC data in rc_columns (keep schema metadata: column_name, data_type, etc.)
	if _, err := tx.ExecContext(ctx, "UPDATE rc_columns SET description = NULL, sample_values = NULL, synonyms = NULL, value_mapping = NULL, updated_at = NOW() WHERE datasource_id = ?", dsID); err != nil {
		return fmt.Errorf("lakebase: failed to clear rc_columns context: %w", err)
	}

	// Delete generated context data (these can be fully regenerated)
	// NOTE: rc_relations is NOT pruned — it contains physical FK metadata from information_schema,
	// not AI-generated context. Deleting it would break ForestDecompose (all tables become isolated).
	deleteTables := []string{
		"rc_business_context",
		"rc_change_log",
		"rc_terms",
	}

	for _, table := range deleteTables {
		query := fmt.Sprintf("DELETE FROM %s WHERE datasource_id = ?", table)
		if _, err := tx.ExecContext(ctx, query, dsID); err != nil {
			return fmt.Errorf("lakebase: failed to delete from %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("lakebase: failed to commit transaction: %w", err)
	}
	return nil
}

func (r *MySQLRepository) queryBusinessContext(ctx context.Context, query string, args ...interface{}) ([]*BusinessContext, error) {
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query business context: %w", err)
	}
	defer rows.Close()

	var contexts []*BusinessContext
	for rows.Next() {
		bc := &BusinessContext{}
		if err := rows.Scan(
			&bc.ID, &bc.DatasourceID, &bc.TableName, &bc.ColumnName, &bc.ContextType, &bc.Content,
			&bc.Source, &bc.Confidence, &bc.IsExpired, &bc.ExpiresAt, &bc.Version,
			&bc.CreatedAt, &bc.UpdatedAt, &bc.CreatedBy, &bc.UpdatedBy, &bc.UpdateReason); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan business context: %w", err)
		}
		contexts = append(contexts, bc)
	}
	return contexts, rows.Err()
}

// ===========================================
// Change log operations
// ===========================================

func (r *MySQLRepository) CreateChangeLog(ctx context.Context, log *ChangeLog) (int64, error) {
	query := `
		INSERT INTO rc_change_log
		(datasource_id, table_name, change_type, change_detail, old_value, new_value, trigger_source, change_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		log.DatasourceID, log.TableName, log.ChangeType, log.ChangeDetail,
		log.OldValue, log.NewValue, log.TriggerSource, log.ChangeReason)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to create change log: %w", err)
	}
	return result.LastInsertId()
}

func (r *MySQLRepository) GetChangeLogsByDatasource(ctx context.Context, dsID int64, limit int) ([]*ChangeLog, error) {
	query := `
		SELECT id, datasource_id, table_name, change_type, change_detail, old_value, new_value,
		       trigger_source, change_reason, created_at
		FROM rc_change_log WHERE datasource_id = ?
		ORDER BY created_at DESC LIMIT ?
	`
	return r.queryChangeLogs(ctx, query, dsID, limit)
}

func (r *MySQLRepository) queryChangeLogs(ctx context.Context, query string, args ...interface{}) ([]*ChangeLog, error) {
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query change logs: %w", err)
	}
	defer rows.Close()

	var logs []*ChangeLog
	for rows.Next() {
		log := &ChangeLog{}
		if err := rows.Scan(
			&log.ID, &log.DatasourceID, &log.TableName, &log.ChangeType, &log.ChangeDetail,
			&log.OldValue, &log.NewValue, &log.TriggerSource, &log.ChangeReason, &log.CreatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan change log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// GetRelationsByDatasource retrieves all relations for a datasource
func (r *MySQLRepository) GetRelationsByDatasource(ctx context.Context, dsID int64) ([]*Relation, error) {
	query := `
		SELECT id, datasource_id, from_table, from_column, to_table, to_column, 
		       relation_type, description, created_at, updated_at
		FROM rc_relations WHERE datasource_id = ?
		ORDER BY from_table, to_table
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query relations: %w", err)
	}
	defer rows.Close()

	var relations []*Relation
	for rows.Next() {
		var rel Relation
		if err := rows.Scan(
			&rel.ID, &rel.DatasourceID, &rel.FromTable, &rel.FromColumn,
			&rel.ToTable, &rel.ToColumn, &rel.RelationType, &rel.Description,
			&rel.CreatedAt, &rel.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan relation: %w", err)
		}
		relations = append(relations, &rel)
	}
	return relations, nil
}

// ===========================================
// Expire flag operations (for self-maintaining agent)
// ===========================================

// MarkTablesExpired marks tables as expired by name
func (r *MySQLRepository) MarkTablesExpired(ctx context.Context, dsID int64, tableNames []string) (int64, error) {
	if len(tableNames) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(tableNames))
	args := make([]interface{}, 0, len(tableNames)+1)
	args = append(args, dsID)
	for i, name := range tableNames {
		placeholders[i] = "?"
		args = append(args, name)
	}
	query := fmt.Sprintf(`UPDATE rc_tables SET is_expired = 1, updated_at = NOW() WHERE datasource_id = ? AND table_name IN (%s)`,
		strings.Join(placeholders, ","))
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to mark tables expired: %w", err)
	}
	return result.RowsAffected()
}

// MarkColumnsExpired marks columns as expired by table+column name pairs
func (r *MySQLRepository) MarkColumnsExpired(ctx context.Context, dsID int64, columns []TableColumn) (int64, error) {
	if len(columns) == 0 {
		return 0, nil
	}
	// Build OR conditions: (table_name = ? AND column_name = ?) OR ...
	conditions := make([]string, len(columns))
	args := make([]interface{}, 0, len(columns)*2+1)
	args = append(args, dsID)
	for i, col := range columns {
		conditions[i] = "(table_name = ? AND column_name = ?)"
		args = append(args, col.TableName, col.ColumnName)
	}
	query := fmt.Sprintf(`UPDATE rc_columns SET is_expired = 1, updated_at = NOW() WHERE datasource_id = ? AND (%s)`,
		strings.Join(conditions, " OR "))
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to mark columns expired: %w", err)
	}
	return result.RowsAffected()
}

// MarkAllColumnsExpiredByTable marks all columns of specified tables as expired
func (r *MySQLRepository) MarkAllColumnsExpiredByTable(ctx context.Context, dsID int64, tableNames []string) (int64, error) {
	if len(tableNames) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(tableNames))
	args := make([]interface{}, 0, len(tableNames)+1)
	args = append(args, dsID)
	for i, name := range tableNames {
		placeholders[i] = "?"
		args = append(args, name)
	}
	query := fmt.Sprintf(`UPDATE rc_columns SET is_expired = 1, updated_at = NOW() WHERE datasource_id = ? AND table_name IN (%s)`,
		strings.Join(placeholders, ","))
	result, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to mark columns expired by table: %w", err)
	}
	return result.RowsAffected()
}

// ClearTableExpired clears the expired flag on tables
func (r *MySQLRepository) ClearTableExpired(ctx context.Context, dsID int64, tableNames []string) error {
	if len(tableNames) == 0 {
		return nil
	}
	placeholders := make([]string, len(tableNames))
	args := make([]interface{}, 0, len(tableNames)+1)
	args = append(args, dsID)
	for i, name := range tableNames {
		placeholders[i] = "?"
		args = append(args, name)
	}
	query := fmt.Sprintf(`UPDATE rc_tables SET is_expired = 0, updated_at = NOW() WHERE datasource_id = ? AND table_name IN (%s)`,
		strings.Join(placeholders, ","))
	_, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("lakebase: failed to clear table expired: %w", err)
	}
	return nil
}

// ClearColumnExpired clears the expired flag on columns
func (r *MySQLRepository) ClearColumnExpired(ctx context.Context, dsID int64, columns []TableColumn) error {
	if len(columns) == 0 {
		return nil
	}
	conditions := make([]string, len(columns))
	args := make([]interface{}, 0, len(columns)*2+1)
	args = append(args, dsID)
	for i, col := range columns {
		conditions[i] = "(table_name = ? AND column_name = ?)"
		args = append(args, col.TableName, col.ColumnName)
	}
	query := fmt.Sprintf(`UPDATE rc_columns SET is_expired = 0, updated_at = NOW() WHERE datasource_id = ? AND (%s)`,
		strings.Join(conditions, " OR "))
	_, err := r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("lakebase: failed to clear column expired: %w", err)
	}
	return nil
}

// GetExpiredTables returns tables with is_expired = 1
func (r *MySQLRepository) GetExpiredTables(ctx context.Context, dsID int64) ([]*TableInfo, error) {
	query := `
		SELECT id, datasource_id, table_name, description, row_count, is_expired,
		       source, confidence, created_at, updated_at
		FROM rc_tables WHERE datasource_id = ? AND is_expired = 1
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get expired tables: %w", err)
	}
	defer rows.Close()

	var tables []*TableInfo
	for rows.Next() {
		t := &TableInfo{}
		if err := rows.Scan(&t.ID, &t.DatasourceID, &t.TableName, &t.Description, &t.RowCount,
			&t.IsExpired, &t.Source, &t.Confidence, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan expired table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// GetExpiredColumns returns columns with is_expired = 1
func (r *MySQLRepository) GetExpiredColumns(ctx context.Context, dsID int64) ([]*ColumnInfo, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, data_type, description,
		       sample_values, synonyms, is_nullable, is_primary_key, is_foreign_key,
		       is_expired, source, confidence, created_at, updated_at
		FROM rc_columns WHERE datasource_id = ? AND is_expired = 1
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get expired columns: %w", err)
	}
	defer rows.Close()

	var columns []*ColumnInfo
	for rows.Next() {
		c := &ColumnInfo{}
		if err := rows.Scan(&c.ID, &c.DatasourceID, &c.TableName, &c.ColumnName, &c.DataType,
			&c.Description, &c.SampleValues, &c.Synonyms, &c.IsNullable,
			&c.IsPrimaryKey, &c.IsForeignKey, &c.IsExpired, &c.Source,
			&c.Confidence, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan expired column: %w", err)
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

// DeleteTableByName deletes a table record and its columns from rc_tables/rc_columns
func (r *MySQLRepository) DeleteTableByName(ctx context.Context, dsID int64, tableName string) error {
	return r.pool.WithTransaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "DELETE FROM rc_columns WHERE datasource_id = ? AND table_name = ?", dsID, tableName); err != nil {
			return fmt.Errorf("lakebase: failed to delete columns for table %s: %w", tableName, err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM rc_terms WHERE datasource_id = ? AND category = ?", dsID, tableName); err != nil {
			// terms are optional, don't fail
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM rc_relations WHERE datasource_id = ? AND (from_table = ? OR to_table = ?)", dsID, tableName, tableName); err != nil {
			// relations are optional
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM rc_tables WHERE datasource_id = ? AND table_name = ?", dsID, tableName); err != nil {
			return fmt.Errorf("lakebase: failed to delete table %s: %w", tableName, err)
		}
		return nil
	})
}

// DeleteColumnByName deletes a column record from rc_columns
func (r *MySQLRepository) DeleteColumnByName(ctx context.Context, dsID int64, tableName, columnName string) error {
	query := `DELETE FROM rc_columns WHERE datasource_id = ? AND table_name = ? AND column_name = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, tableName, columnName)
	if err != nil {
		return fmt.Errorf("lakebase: failed to delete column %s.%s: %w", tableName, columnName, err)
	}
	return nil
}

// Ensure MySQLRepository implements Repository interface
var _ Repository = (*MySQLRepository)(nil)
