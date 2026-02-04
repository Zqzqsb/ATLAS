// Package lakebase provides lake-base multi-modal storage operations
// for Rich Context management in LUCID system.
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
	ErrSchemaNotFound     = errors.New("lakebase: schema not found")
	ErrContextNotFound    = errors.New("lakebase: context not found")
	ErrContextExpired     = errors.New("lakebase: context expired")
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
	GetTablesByDatasource(ctx context.Context, dsID int64) ([]*TableInfo, error)
	GetColumnsByDatasource(ctx context.Context, dsID int64) ([]*ColumnInfo, error)
	GetColumnsByTable(ctx context.Context, dsID int64, tableName string) ([]*ColumnInfo, error)

	// Schema metadata operations (legacy)
	SaveSchemaMetadata(ctx context.Context, metas []*SchemaMetadata) error
	GetSchemaByDatasource(ctx context.Context, dsID int64) ([]*SchemaMetadata, error)
	GetTableSchema(ctx context.Context, dsID int64, tableName string) ([]*SchemaMetadata, error)
	GetTableNames(ctx context.Context, dsID int64) ([]string, error)
	DeleteSchemaByDatasource(ctx context.Context, dsID int64) error
	DeleteTableSchema(ctx context.Context, dsID int64, tableName string) error

	// Business context operations
	SaveBusinessContext(ctx context.Context, bc *BusinessContext) (int64, error)
	SaveBusinessContextBatch(ctx context.Context, contexts []*BusinessContext) error
	GetContextByID(ctx context.Context, id int64) (*BusinessContext, error)
	GetContextByDatasource(ctx context.Context, dsID int64) ([]*BusinessContext, error)
	GetContextByTable(ctx context.Context, dsID int64, tableName string) ([]*BusinessContext, error)
	GetContextByColumn(ctx context.Context, dsID int64, tableName, columnName string) ([]*BusinessContext, error)
	GetContextByType(ctx context.Context, dsID int64, contextType ContextType) ([]*BusinessContext, error)
	MarkContextExpired(ctx context.Context, ids []int64, reason string) error
	UpdateContextVersion(ctx context.Context, id int64, content json.RawMessage, updatedBy string, reason string) error
	DeleteExpiredContext(ctx context.Context, dsID int64) error

	// Change log operations
	CreateChangeLog(ctx context.Context, log *ChangeLog) (int64, error)
	GetChangeLogsByDatasource(ctx context.Context, dsID int64, limit int) ([]*ChangeLog, error)
	GetChangeLogsByTable(ctx context.Context, dsID int64, tableName string, limit int) ([]*ChangeLog, error)
	GetChangeLogsByType(ctx context.Context, dsID int64, changeType ChangeType, limit int) ([]*ChangeLog, error)

	// Statistics operations
	SaveStatistics(ctx context.Context, stat *Statistics) (int64, error)
	GetStatisticsByTable(ctx context.Context, dsID int64, tableName string) ([]*Statistics, error)
	DeleteStatisticsByDatasource(ctx context.Context, dsID int64) error

	// JOIN path operations
	SaveJoinPath(ctx context.Context, path *JoinPath) (int64, error)
	SaveJoinPathBatch(ctx context.Context, paths []*JoinPath) error
	GetJoinPath(ctx context.Context, dsID int64, fromTable, toTable string) (*JoinPath, error)
	GetJoinPathsFromTable(ctx context.Context, dsID int64, fromTable string) ([]*JoinPath, error)
	DeleteJoinPathsByDatasource(ctx context.Context, dsID int64) error
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

// ===========================================
// Schema metadata operations (legacy)
// ===========================================

func (r *MySQLRepository) SaveSchemaMetadata(ctx context.Context, metas []*SchemaMetadata) error {
	if len(metas) == 0 {
		return nil
	}

	query := `
		INSERT INTO rc_schema_metadata
		(datasource_id, table_name, column_name, data_type, is_primary_key, is_foreign_key,
		 fk_ref_table, fk_ref_column, nullable, default_value, comment)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		data_type = VALUES(data_type),
		is_primary_key = VALUES(is_primary_key),
		is_foreign_key = VALUES(is_foreign_key),
		fk_ref_table = VALUES(fk_ref_table),
		fk_ref_column = VALUES(fk_ref_column),
		nullable = VALUES(nullable),
		default_value = VALUES(default_value),
		comment = VALUES(comment)
	`

	return r.pool.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("lakebase: failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, meta := range metas {
			_, err := stmt.ExecContext(ctx,
				meta.DatasourceID, meta.TableName, meta.ColumnName, meta.DataType,
				meta.IsPrimaryKey, meta.IsForeignKey, meta.FKRefTable, meta.FKRefColumn,
				meta.Nullable, meta.DefaultValue, meta.Comment)
			if err != nil {
				return fmt.Errorf("lakebase: failed to insert schema metadata: %w", err)
			}
		}
		return nil
	})
}

func (r *MySQLRepository) GetSchemaByDatasource(ctx context.Context, dsID int64) ([]*SchemaMetadata, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, data_type, is_primary_key, is_foreign_key,
		       fk_ref_table, fk_ref_column, nullable, default_value, comment, created_at, updated_at
		FROM rc_schema_metadata WHERE datasource_id = ?
		ORDER BY table_name, id
	`
	return r.querySchemaMetadata(ctx, query, dsID)
}

func (r *MySQLRepository) GetTableSchema(ctx context.Context, dsID int64, tableName string) ([]*SchemaMetadata, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, data_type, is_primary_key, is_foreign_key,
		       fk_ref_table, fk_ref_column, nullable, default_value, comment, created_at, updated_at
		FROM rc_schema_metadata WHERE datasource_id = ? AND table_name = ?
		ORDER BY id
	`
	return r.querySchemaMetadata(ctx, query, dsID, tableName)
}

func (r *MySQLRepository) GetTableNames(ctx context.Context, dsID int64) ([]string, error) {
	query := `SELECT DISTINCT table_name FROM rc_schema_metadata WHERE datasource_id = ? ORDER BY table_name`
	rows, err := r.pool.QueryContext(ctx, query, dsID)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get table names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan table name: %w", err)
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (r *MySQLRepository) DeleteSchemaByDatasource(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_schema_metadata WHERE datasource_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

func (r *MySQLRepository) DeleteTableSchema(ctx context.Context, dsID int64, tableName string) error {
	query := `DELETE FROM rc_schema_metadata WHERE datasource_id = ? AND table_name = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID, tableName)
	return err
}

func (r *MySQLRepository) querySchemaMetadata(ctx context.Context, query string, args ...interface{}) ([]*SchemaMetadata, error) {
	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query schema metadata: %w", err)
	}
	defer rows.Close()

	var metas []*SchemaMetadata
	for rows.Next() {
		m := &SchemaMetadata{}
		if err := rows.Scan(
			&m.ID, &m.DatasourceID, &m.TableName, &m.ColumnName, &m.DataType,
			&m.IsPrimaryKey, &m.IsForeignKey, &m.FKRefTable, &m.FKRefColumn,
			&m.Nullable, &m.DefaultValue, &m.Comment, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan schema metadata: %w", err)
		}
		metas = append(metas, m)
	}
	return metas, rows.Err()
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

func (r *MySQLRepository) GetContextByTable(ctx context.Context, dsID int64, tableName string) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE datasource_id = ? AND table_name = ? AND is_expired = 0
		ORDER BY column_name, context_type
	`
	return r.queryBusinessContext(ctx, query, dsID, tableName)
}

func (r *MySQLRepository) GetContextByColumn(ctx context.Context, dsID int64, tableName, columnName string) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context
		WHERE datasource_id = ? AND table_name = ? AND column_name = ? AND is_expired = 0
		ORDER BY context_type
	`
	return r.queryBusinessContext(ctx, query, dsID, tableName, columnName)
}

func (r *MySQLRepository) GetContextByType(ctx context.Context, dsID int64, contextType ContextType) ([]*BusinessContext, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, context_type, content, source, confidence,
		       is_expired, expires_at, version, created_at, updated_at, created_by, updated_by, update_reason
		FROM rc_business_context WHERE datasource_id = ? AND context_type = ? AND is_expired = 0
		ORDER BY table_name, column_name
	`
	return r.queryBusinessContext(ctx, query, dsID, contextType)
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

func (r *MySQLRepository) GetChangeLogsByTable(ctx context.Context, dsID int64, tableName string, limit int) ([]*ChangeLog, error) {
	query := `
		SELECT id, datasource_id, table_name, change_type, change_detail, old_value, new_value,
		       trigger_source, change_reason, created_at
		FROM rc_change_log WHERE datasource_id = ? AND table_name = ?
		ORDER BY created_at DESC LIMIT ?
	`
	return r.queryChangeLogs(ctx, query, dsID, tableName, limit)
}

func (r *MySQLRepository) GetChangeLogsByType(ctx context.Context, dsID int64, changeType ChangeType, limit int) ([]*ChangeLog, error) {
	query := `
		SELECT id, datasource_id, table_name, change_type, change_detail, old_value, new_value,
		       trigger_source, change_reason, created_at
		FROM rc_change_log WHERE datasource_id = ? AND change_type = ?
		ORDER BY created_at DESC LIMIT ?
	`
	return r.queryChangeLogs(ctx, query, dsID, changeType, limit)
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

// ===========================================
// Statistics operations
// ===========================================

func (r *MySQLRepository) SaveStatistics(ctx context.Context, stat *Statistics) (int64, error) {
	query := `
		INSERT INTO rc_statistics (datasource_id, table_name, column_name, stat_type, stat_value)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		stat.DatasourceID, stat.TableName, stat.ColumnName, stat.StatType, stat.StatValue)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to save statistics: %w", err)
	}
	return result.LastInsertId()
}

func (r *MySQLRepository) GetStatisticsByTable(ctx context.Context, dsID int64, tableName string) ([]*Statistics, error) {
	query := `
		SELECT id, datasource_id, table_name, column_name, stat_type, stat_value, collected_at
		FROM rc_statistics WHERE datasource_id = ? AND table_name = ?
		ORDER BY column_name, stat_type
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID, tableName)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query statistics: %w", err)
	}
	defer rows.Close()

	var stats []*Statistics
	for rows.Next() {
		s := &Statistics{}
		if err := rows.Scan(
			&s.ID, &s.DatasourceID, &s.TableName, &s.ColumnName,
			&s.StatType, &s.StatValue, &s.CollectedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan statistics: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *MySQLRepository) DeleteStatisticsByDatasource(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_statistics WHERE datasource_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

// ===========================================
// JOIN path operations
// ===========================================

func (r *MySQLRepository) SaveJoinPath(ctx context.Context, path *JoinPath) (int64, error) {
	query := `
		INSERT INTO rc_join_paths
		(datasource_id, from_table, to_table, path_tables, join_conditions, path_length, confidence)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.pool.ExecContext(ctx, query,
		path.DatasourceID, path.FromTable, path.ToTable, path.PathTables,
		path.JoinConditions, path.PathLength, path.Confidence)
	if err != nil {
		return 0, fmt.Errorf("lakebase: failed to save join path: %w", err)
	}
	return result.LastInsertId()
}

func (r *MySQLRepository) SaveJoinPathBatch(ctx context.Context, paths []*JoinPath) error {
	if len(paths) == 0 {
		return nil
	}

	query := `
		INSERT INTO rc_join_paths
		(datasource_id, from_table, to_table, path_tables, join_conditions, path_length, confidence)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	return r.pool.WithTransaction(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("lakebase: failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, path := range paths {
			_, err := stmt.ExecContext(ctx,
				path.DatasourceID, path.FromTable, path.ToTable, path.PathTables,
				path.JoinConditions, path.PathLength, path.Confidence)
			if err != nil {
				return fmt.Errorf("lakebase: failed to insert join path: %w", err)
			}
		}
		return nil
	})
}

func (r *MySQLRepository) GetJoinPath(ctx context.Context, dsID int64, fromTable, toTable string) (*JoinPath, error) {
	query := `
		SELECT id, datasource_id, from_table, to_table, path_tables, join_conditions, path_length, confidence, created_at
		FROM rc_join_paths WHERE datasource_id = ? AND from_table = ? AND to_table = ?
		ORDER BY path_length ASC LIMIT 1
	`
	path := &JoinPath{}
	err := r.pool.QueryRowContext(ctx, query, dsID, fromTable, toTable).Scan(
		&path.ID, &path.DatasourceID, &path.FromTable, &path.ToTable,
		&path.PathTables, &path.JoinConditions, &path.PathLength, &path.Confidence, &path.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // No path found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to get join path: %w", err)
	}
	return path, nil
}

func (r *MySQLRepository) GetJoinPathsFromTable(ctx context.Context, dsID int64, fromTable string) ([]*JoinPath, error) {
	query := `
		SELECT id, datasource_id, from_table, to_table, path_tables, join_conditions, path_length, confidence, created_at
		FROM rc_join_paths WHERE datasource_id = ? AND from_table = ?
		ORDER BY to_table, path_length
	`
	rows, err := r.pool.QueryContext(ctx, query, dsID, fromTable)
	if err != nil {
		return nil, fmt.Errorf("lakebase: failed to query join paths: %w", err)
	}
	defer rows.Close()

	var paths []*JoinPath
	for rows.Next() {
		path := &JoinPath{}
		if err := rows.Scan(
			&path.ID, &path.DatasourceID, &path.FromTable, &path.ToTable,
			&path.PathTables, &path.JoinConditions, &path.PathLength, &path.Confidence, &path.CreatedAt); err != nil {
			return nil, fmt.Errorf("lakebase: failed to scan join path: %w", err)
		}
		paths = append(paths, path)
	}
	return paths, rows.Err()
}

func (r *MySQLRepository) DeleteJoinPathsByDatasource(ctx context.Context, dsID int64) error {
	query := `DELETE FROM rc_join_paths WHERE datasource_id = ?`
	_, err := r.pool.ExecContext(ctx, query, dsID)
	return err
}

// ===========================================
// Helper methods for content parsing
// ===========================================

// ParseEnumMeaningContent parses content for enum_meaning context type
func ParseEnumMeaningContent(content json.RawMessage) (*EnumMeaningContent, error) {
	var result EnumMeaningContent
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse enum meaning content: %w", err)
	}
	return &result, nil
}

// ParseBusinessRuleContent parses content for business_rule context type
func ParseBusinessRuleContent(content json.RawMessage) (*BusinessRuleContent, error) {
	var result BusinessRuleContent
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse business rule content: %w", err)
	}
	return &result, nil
}

// ParseJoinHintContent parses content for join_hint context type
func ParseJoinHintContent(content json.RawMessage) (*JoinHintContent, error) {
	var result JoinHintContent
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse join hint content: %w", err)
	}
	return &result, nil
}

// ParseSemanticContent parses content for semantic context type
func ParseSemanticContent(content json.RawMessage) (*SemanticContent, error) {
	var result SemanticContent
	if err := json.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("lakebase: failed to parse semantic content: %w", err)
	}
	return &result, nil
}

// NewEnumMeaningContent creates JSON content for enum meanings
func NewEnumMeaningContent(values map[string]string) (json.RawMessage, error) {
	content := EnumMeaningContent{Values: values}
	return json.Marshal(content)
}

// NewBusinessRuleContent creates JSON content for business rules
func NewBusinessRuleContent(rules []string, constraints []string) (json.RawMessage, error) {
	content := BusinessRuleContent{Rules: rules, Constraints: constraints}
	return json.Marshal(content)
}

// NewJoinHintContent creates JSON content for join hints
func NewJoinHintContent(relatedTables, joinKeys []string, description string) (json.RawMessage, error) {
	content := JoinHintContent{RelatedTables: relatedTables, JoinKeys: joinKeys, Description: description}
	return json.Marshal(content)
}

// NewSemanticContent creates JSON content for semantic descriptions
func NewSemanticContent(description string, synonyms, businessTerms []string) (json.RawMessage, error) {
	content := SemanticContent{Description: description, Synonyms: synonyms, BusinessTerms: businessTerms}
	return json.Marshal(content)
}

// Ensure MySQLRepository implements Repository interface
var _ Repository = (*MySQLRepository)(nil)
