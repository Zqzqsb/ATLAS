-- LUCID Lake-Base Storage Initialization
-- MariaDB 12+ with VECTOR support
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
    name VARCHAR(255) NOT NULL UNIQUE,
    type ENUM('mysql', 'postgresql', 'sqlite') NOT NULL,
    host VARCHAR(255),
    port INT,
    database_name VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 2. Tables (table-level Rich Context)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_tables (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    description TEXT,
    row_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    source ENUM('catalog', 'llm', 'user', 'analysis') DEFAULT 'llm',
    UNIQUE KEY uk_datasource_table (datasource_id, table_name),
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 3. Columns (column-level Rich Context)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_columns (
    id INT AUTO_INCREMENT PRIMARY KEY,
    table_id INT NOT NULL,
    column_name VARCHAR(255) NOT NULL,
    data_type VARCHAR(100),
    description TEXT,
    synonyms JSON,
    examples JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    source ENUM('catalog', 'llm', 'user', 'analysis') DEFAULT 'llm',
    UNIQUE KEY uk_table_column (table_id, column_name),
    FOREIGN KEY (table_id) REFERENCES rc_tables(id) ON DELETE CASCADE,
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    relation_type ENUM('foreign_key', 'semantic', 'inferred') DEFAULT 'foreign_key',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_from_table (from_table),
    INDEX idx_to_table (to_table)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 5. Terms (business terminology dictionary)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_terms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    term VARCHAR(255) NOT NULL,
    definition TEXT NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_datasource_term (datasource_id, term),
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_term (term)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 6. Embeddings (vector embeddings with HNSW index)
-- ============================================================
-- Note: Requires MariaDB 12+ with VECTOR support
-- Vector dimension: 1536 (OpenAI text-embedding-3-small)
CREATE TABLE IF NOT EXISTS rc_embeddings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    entity_type ENUM('table', 'column', 'term') NOT NULL,
    entity_id INT NOT NULL,
    text_content TEXT NOT NULL,
    embedding VECTOR(1536) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_entity (entity_type, entity_id),
    VECTOR INDEX idx_embedding_hnsw (embedding) DISTANCE=COSINE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- 7. Change Log (audit trail for self-maintaining agent)
-- ============================================================
CREATE TABLE IF NOT EXISTS rc_change_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    datasource_id INT NOT NULL,
    change_type ENUM('ddl_add_table', 'ddl_drop_table', 'ddl_alter_column', 'context_update', 'embedding_refresh') NOT NULL,
    entity_type ENUM('table', 'column', 'relation', 'term') NOT NULL,
    entity_name VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) DEFAULT 'agent',
    FOREIGN KEY (datasource_id) REFERENCES rc_datasources(id) ON DELETE CASCADE,
    INDEX idx_datasource_time (datasource_id, created_at),
    INDEX idx_entity (entity_type, entity_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================
-- Demo: Insert a sample datasource
-- ============================================================
INSERT IGNORE INTO rc_datasources (name, type, host, port, database_name, description) 
VALUES ('demo_ecommerce', 'mysql', 'lucid-demo-mysql', 3306, 'ecommerce', 'E-commerce demo database');

-- Success message
SELECT 'LUCID Lake-Base storage initialized successfully!' AS status;
