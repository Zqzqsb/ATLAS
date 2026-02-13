-- LUCID Lake-Base Storage Initialization
-- MariaDB 12 with VECTOR support
--
-- This script creates the rc_* (Rich Context) tables that store:
-- - Database metadata
-- - Rich Context (table/column descriptions)
-- - Vector embeddings with HNSW index
-- - Change logs for self-maintaining agent

-- ============================================================
-- 1. Datasources (registered databases)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_datasources (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE COMMENT 'Datasource identifier',
    db_type ENUM('mysql', 'postgresql', 'sqlite', 'mariadb') NOT NULL DEFAULT 'mysql',
    host VARCHAR(255) COMMENT 'Database host',
    port INT DEFAULT 3306 COMMENT 'Database port',
    db_name VARCHAR(255) COMMENT 'Database name',
    username VARCHAR(100) COMMENT 'Connection username',
    description TEXT COMMENT 'Human-readable description',
    status ENUM('active', 'inactive', 'error') DEFAULT 'active',
    last_sync_at TIMESTAMP NULL COMMENT 'Last schema sync time',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Registered data sources for LUCID system';

-- ============================================================
-- 2. Tables (table-level Rich Context)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_tables (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    description TEXT COMMENT 'Semantic description of the table',
    row_count BIGINT DEFAULT 0 COMMENT 'Estimated row count',
    is_expired TINYINT(1) DEFAULT 0 COMMENT 'Whether context needs refresh',
    source ENUM('catalog', 'llm', 'user', 'analysis') DEFAULT 'llm',
    confidence DECIMAL(3,2) DEFAULT 0.80 COMMENT 'Confidence score 0-1',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    UNIQUE KEY uk_datasource_table (datasource_id, table_name),
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_expired (is_expired)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Table-level Rich Context';

-- ============================================================
-- 3. Columns (column-level Rich Context)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_columns (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(100) COMMENT 'Column data type',
    description TEXT COMMENT 'Semantic description',
    sample_values TEXT COMMENT 'Sample values (comma separated)',
    synonyms TEXT COMMENT 'Alternative names (comma separated)',
    value_mapping TEXT COMMENT 'Enum value meanings (JSON or text)',
    is_nullable TINYINT(1) DEFAULT 1,
    is_primary_key TINYINT(1) DEFAULT 0,
    is_foreign_key TINYINT(1) DEFAULT 0,
    is_expired TINYINT(1) DEFAULT 0,
    source ENUM('catalog', 'llm', 'user', 'analysis') DEFAULT 'llm',
    confidence DECIMAL(3,2) DEFAULT 0.80,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_column (datasource_id, table_name, column_name),
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_table (datasource_id, table_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Column-level Rich Context';

-- ============================================================
-- 4. Relations (table relationships / foreign keys)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_relations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    from_table VARCHAR(255) NOT NULL,
    from_column VARCHAR(255) NOT NULL,
    to_table VARCHAR(255) NOT NULL,
    to_column VARCHAR(255) NOT NULL,
    relation_type ENUM('one_to_one', 'one_to_many', 'many_to_one', 'many_to_many', 'foreign_key', 'inferred') DEFAULT 'foreign_key',
    description TEXT COMMENT 'Relationship description',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    UNIQUE KEY uk_relation (datasource_id, from_table, from_column, to_table, to_column),
    INDEX idx_from (datasource_id, from_table),
    INDEX idx_to (datasource_id, to_table)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Table relationships and join paths';

-- ============================================================
-- 5. Terms (business terminology dictionary)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_terms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    term VARCHAR(255) NOT NULL,
    definition TEXT NOT NULL COMMENT 'Term definition',
    synonyms TEXT COMMENT 'Alternative terms (comma separated)',
    examples TEXT COMMENT 'Usage examples',
    category VARCHAR(100) COMMENT 'Term category',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_term (datasource_id, term),
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_category (category),
    FULLTEXT INDEX ft_term_def (term, definition)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Business terminology dictionary';

-- ============================================================
-- 6. Embeddings (vector embeddings with HNSW index)
-- ============================================================
-- Note: Requires MariaDB 12 with VECTOR support
-- Vector dimension: 1536 (OpenAI text-embedding-3-small)
CREATE TABLE IF NOT EXISTS rc_embeddings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    entity_type ENUM('table', 'column', 'term', 'query') NOT NULL,
    entity_id INT NOT NULL COMMENT 'ID in the corresponding rc_* table',
    entity_text TEXT NOT NULL COMMENT 'Text that was embedded',
    embedding VECTOR(2048) NOT NULL COMMENT 'Vector embedding',
    embedding_model VARCHAR(100) DEFAULT 'doubao-embedding-large-text-250515',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL DEFAULT NULL,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_entity (entity_type, entity_id),
    VECTOR INDEX idx_embedding_hnsw (embedding) DISTANCE=COSINE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Vector embeddings for semantic search';

-- ============================================================
-- 7. Business Context (unified table/column context)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_business_context (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255) NULL COMMENT 'NULL for table-level context',
    context_type ENUM('description', 'example', 'constraint', 'synonym', 'value_mapping', 'business_rule', 'calculation', 'semantic', 'enum_meaning', 'join_hint', 'data_quality') NOT NULL DEFAULT 'description',
    content TEXT NOT NULL COMMENT 'The actual context content',
    source ENUM('catalog', 'llm', 'user', 'analysis') DEFAULT 'llm',
    confidence DECIMAL(3,2) DEFAULT 0.80 COMMENT 'Confidence score 0-1',
    is_expired TINYINT(1) DEFAULT 0 COMMENT 'Whether context needs refresh',
    expires_at TIMESTAMP NULL DEFAULT NULL,
    version INT DEFAULT 1 COMMENT 'Version number for tracking updates',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by VARCHAR(100) DEFAULT 'system',
    updated_by VARCHAR(100) DEFAULT 'system',
    update_reason TEXT COMMENT 'Reason for last update',
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_table (datasource_id, table_name),
    INDEX idx_column (datasource_id, table_name, column_name),
    INDEX idx_type (context_type),
    INDEX idx_expired (is_expired)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Unified business context for tables and columns';

-- ============================================================
-- 8. Statistics (column statistics for query optimization)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_statistics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255),
    stat_type VARCHAR(50) NOT NULL,
    stat_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_table (datasource_id, table_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Column statistics for intelligent query optimization';

-- ============================================================
-- 10. Join Paths (pre-computed join paths between tables)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_join_paths (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    from_table VARCHAR(255) NOT NULL,
    to_table VARCHAR(255) NOT NULL,
    join_columns TEXT NOT NULL,
    join_type VARCHAR(50) DEFAULT 'INNER',
    path_cost DECIMAL(5,2) DEFAULT 1.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_from (datasource_id, from_table),
    INDEX idx_to (datasource_id, to_table)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Pre-computed join paths for multi-table queries';

-- ============================================================
-- 11. Change Log (audit trail for self-maintaining agent)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_change_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    change_type VARCHAR(50) NOT NULL COMMENT 'Type of change: ddl_*, context_*, init, etc.',
    entity_type VARCHAR(50) NOT NULL COMMENT 'table, column, relation, term, system',
    entity_name VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    reason TEXT COMMENT 'Reason for change',
    changed_by VARCHAR(100) DEFAULT 'agent' COMMENT 'Who made the change',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_time (datasource_id, created_at DESC),
    INDEX idx_entity (entity_type, entity_name),
    INDEX idx_type (change_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Change log for audit and self-maintenance';

-- ============================================================
-- Success message
-- ============================================================
SELECT 'LUCID Lake-Base storage initialized successfully!' AS status;
