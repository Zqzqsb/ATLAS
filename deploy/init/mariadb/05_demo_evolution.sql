-- =============================================================
-- LUCID Evolution Demo Database
-- Demonstrates Agent self-maintenance via Schema evolution stages
-- =============================================================

-- Create database
CREATE DATABASE IF NOT EXISTS lucid_evolution DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE lucid_evolution;

-- =============================================================
-- Stage 0: Initial state — simple user-order system
-- =============================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT 'User name',
    email VARCHAR(255) UNIQUE COMMENT 'User email'
) COMMENT='User information table';

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT 'Associated user ID',
    amount DECIMAL(10,2) COMMENT 'Order amount',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    FOREIGN KEY (user_id) REFERENCES users(id)
) COMMENT='Order information table';

-- Initial sample data
INSERT INTO users (name, email) VALUES 
    ('Alice', 'alice@example.com'),
    ('Bob', 'bob@example.com'),
    ('Charlie', 'charlie@example.com');

INSERT INTO orders (user_id, amount) VALUES 
    (1, 99.00),
    (1, 199.00),
    (2, 59.00),
    (3, 299.00);

-- Grant lucid user full access to this database (needed for DDL evolution stages)
GRANT ALL PRIVILEGES ON lucid_evolution.* TO 'lucid'@'%';
FLUSH PRIVILEGES;

-- =============================================================
-- Register in LUCID Lake-Base (rc_datasources)
-- =============================================================
USE lucid;

INSERT INTO rc_datasources (name, db_type, host, port, db_name, username, description, status)
VALUES ('lucid_evolution', 'mariadb', 'lucid-mariadb', 3306, 'lucid_evolution', 'lucid',
        'Evolution Demo - Agent self-maintenance showcase (DDL evolution stages)', 'active')
ON DUPLICATE KEY UPDATE status = 'active', description = VALUES(description);

-- Get the inserted datasource ID (using variable)
SET @evo_ds_id = (SELECT id FROM rc_datasources WHERE name = 'lucid_evolution' LIMIT 1);

-- Pre-register schema metadata into rc_tables
INSERT INTO rc_tables (datasource_id, table_name, row_count)
VALUES (@evo_ds_id, 'users', 3),
       (@evo_ds_id, 'orders', 4)
ON DUPLICATE KEY UPDATE row_count = VALUES(row_count);

-- Pre-register column metadata into rc_columns
-- users table
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, is_nullable, is_primary_key, is_foreign_key)
VALUES (@evo_ds_id, 'users', 'id', 'int', 0, 1, 0),
       (@evo_ds_id, 'users', 'name', 'varchar', 0, 0, 0),
       (@evo_ds_id, 'users', 'email', 'varchar', 1, 0, 0)
ON DUPLICATE KEY UPDATE data_type = VALUES(data_type);

-- orders table
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, is_nullable, is_primary_key, is_foreign_key)
VALUES (@evo_ds_id, 'orders', 'id', 'int', 0, 1, 0),
       (@evo_ds_id, 'orders', 'user_id', 'int', 0, 0, 1),
       (@evo_ds_id, 'orders', 'amount', 'decimal', 1, 0, 0),
       (@evo_ds_id, 'orders', 'created_at', 'datetime', 1, 0, 0)
ON DUPLICATE KEY UPDATE data_type = VALUES(data_type);

-- =============================================================
-- Evolution stage descriptions (not executed, for reference only)
-- =============================================================

-- Stage 1: ALTER TABLE users ADD COLUMN phone VARCHAR(20);
--   -> Detect column_added, generate Rich Context for phone column

-- Stage 2: CREATE TABLE products (...);
--   -> Detect table_added, generate full Rich Context suite for products table

-- Stage 3: ALTER TABLE orders ADD COLUMN product_id INT;
--          ALTER TABLE orders ADD FOREIGN KEY (product_id) REFERENCES products(id);
--   -> Detect column_added + fk_added, update relationship graph

-- Stage 4: ALTER TABLE orders MODIFY amount DECIMAL(15,4);
--   -> Detect column_modified, refresh amount column Context

-- Stage 5: ALTER TABLE users DROP COLUMN email;
--   -> Detect column_dropped, clean up email-related Context
