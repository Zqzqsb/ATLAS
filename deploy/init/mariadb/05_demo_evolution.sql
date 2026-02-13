-- =============================================================
-- LUCID Evolution Demo Database
-- 用于展示 Agent 自维持机制的 Schema 演进演示库
-- =============================================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS lucid_evolution DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE lucid_evolution;

-- =============================================================
-- Stage 0: 初始状态 — 简单的用户订单系统
-- =============================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '用户姓名',
    email VARCHAR(255) UNIQUE COMMENT '用户邮箱'
) COMMENT='用户信息表';

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '关联用户ID',
    amount DECIMAL(10,2) COMMENT '订单金额',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (user_id) REFERENCES users(id)
) COMMENT='订单信息表';

-- 初始示例数据
INSERT INTO users (name, email) VALUES 
    ('张三', 'zhang@example.com'),
    ('李四', 'li@example.com'),
    ('王五', 'wang@example.com');

INSERT INTO orders (user_id, amount) VALUES 
    (1, 99.00),
    (1, 199.00),
    (2, 59.00),
    (3, 299.00);

-- Grant lucid user full access to this database (needed for DDL evolution stages)
GRANT ALL PRIVILEGES ON lucid_evolution.* TO 'lucid'@'%';
FLUSH PRIVILEGES;

-- =============================================================
-- 注册到 LUCID Lake-Base (rc_datasources)
-- =============================================================
USE lucid;

INSERT INTO rc_datasources (name, db_type, host, port, db_name, username, description, status)
VALUES ('lucid_evolution', 'mariadb', 'lucid-mariadb', 3306, 'lucid_evolution', 'lucid',
        'Evolution Demo - Agent self-maintenance showcase (DDL evolution stages)', 'active')
ON DUPLICATE KEY UPDATE status = 'active', description = VALUES(description);

-- 获取刚插入的 datasource ID (使用变量)
SET @evo_ds_id = (SELECT id FROM rc_datasources WHERE name = 'lucid_evolution' LIMIT 1);

-- 预注册 schema 元数据到 rc_tables
INSERT INTO rc_tables (datasource_id, table_name, row_count)
VALUES (@evo_ds_id, 'users', 3),
       (@evo_ds_id, 'orders', 4)
ON DUPLICATE KEY UPDATE row_count = VALUES(row_count);

-- 预注册列元数据到 rc_columns
-- users 表
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, is_nullable, is_primary_key, is_foreign_key)
VALUES (@evo_ds_id, 'users', 'id', 'int', 0, 1, 0),
       (@evo_ds_id, 'users', 'name', 'varchar', 0, 0, 0),
       (@evo_ds_id, 'users', 'email', 'varchar', 1, 0, 0)
ON DUPLICATE KEY UPDATE data_type = VALUES(data_type);

-- orders 表
INSERT INTO rc_columns (datasource_id, table_name, column_name, data_type, is_nullable, is_primary_key, is_foreign_key)
VALUES (@evo_ds_id, 'orders', 'id', 'int', 0, 1, 0),
       (@evo_ds_id, 'orders', 'user_id', 'int', 0, 0, 1),
       (@evo_ds_id, 'orders', 'amount', 'decimal', 1, 0, 0),
       (@evo_ds_id, 'orders', 'created_at', 'datetime', 1, 0, 0)
ON DUPLICATE KEY UPDATE data_type = VALUES(data_type);

-- =============================================================
-- 演进阶段说明 (不执行，仅供参考)
-- =============================================================

-- Stage 1: ALTER TABLE users ADD COLUMN phone VARCHAR(20);
--   → 检测 column_added，为 phone 列生成 Rich Context

-- Stage 2: CREATE TABLE products (...);
--   → 检测 table_added，为 products 表生成全套 Rich Context

-- Stage 3: ALTER TABLE orders ADD COLUMN product_id INT;
--          ALTER TABLE orders ADD FOREIGN KEY (product_id) REFERENCES products(id);
--   → 检测 column_added + fk_added，更新关系图谱

-- Stage 4: ALTER TABLE orders MODIFY amount DECIMAL(15,4);
--   → 检测 column_modified，刷新 amount 列 Context

-- Stage 5: ALTER TABLE users DROP COLUMN email;
--   → 检测 column_dropped，清理 email 相关 Context
