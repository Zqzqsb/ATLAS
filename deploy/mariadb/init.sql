-- ============================================================
-- ReActSQL Demo MySQL Database
-- E-commerce scenario with sample data for Text-to-SQL demos
-- ============================================================

-- Create database
CREATE DATABASE IF NOT EXISTS ecommerce CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ecommerce;

-- ============================================================
-- Table: customers - Customer information
-- ============================================================
CREATE TABLE customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT 'Customer name',
    email VARCHAR(150) UNIQUE NOT NULL COMMENT 'Email address',
    phone VARCHAR(20) COMMENT 'Phone number',
    city VARCHAR(50) COMMENT 'City',
    vip_level ENUM('normal', 'silver', 'gold', 'platinum') DEFAULT 'normal' COMMENT 'VIP level: normal=普通会员, silver=白银会员, gold=黄金会员, platinum=白金会员',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Registration time',
    INDEX idx_city (city),
    INDEX idx_vip (vip_level)
) COMMENT='Customer table';

-- ============================================================
-- Table: products - Product catalog
-- ============================================================
CREATE TABLE products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL COMMENT 'Product name',
    category VARCHAR(50) NOT NULL COMMENT 'Category: electronics=电子产品, clothing=服装, food=食品, books=图书, home=家居',
    price DECIMAL(10, 2) NOT NULL COMMENT 'Unit price',
    stock INT DEFAULT 0 COMMENT 'Stock quantity',
    status ENUM('active', 'inactive', 'discontinued') DEFAULT 'active' COMMENT 'Status: active=在售, inactive=下架, discontinued=停产',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_status (status)
) COMMENT='Product table';

-- ============================================================
-- Table: orders - Order information
-- ============================================================
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_no VARCHAR(32) UNIQUE NOT NULL COMMENT 'Order number',
    customer_id INT NOT NULL COMMENT 'Customer ID',
    total_amount DECIMAL(12, 2) NOT NULL COMMENT 'Total amount',
    status ENUM('pending', 'paid', 'shipped', 'delivered', 'cancelled', 'refunded') DEFAULT 'pending' 
        COMMENT 'Order status: pending=待支付, paid=已支付, shipped=已发货, delivered=已签收, cancelled=已取消, refunded=已退款',
    payment_method ENUM('alipay', 'wechat', 'credit_card', 'bank_transfer') COMMENT 'Payment method: alipay=支付宝, wechat=微信支付, credit_card=信用卡, bank_transfer=银行转账',
    order_date DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Order time',
    shipped_date DATETIME COMMENT 'Shipping time',
    delivered_date DATETIME COMMENT 'Delivery time',
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_customer (customer_id),
    INDEX idx_status (status),
    INDEX idx_date (order_date)
) COMMENT='Order table';

-- ============================================================
-- Table: order_items - Order line items
-- ============================================================
CREATE TABLE order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL COMMENT 'Order ID',
    product_id INT NOT NULL COMMENT 'Product ID',
    quantity INT NOT NULL COMMENT 'Quantity',
    unit_price DECIMAL(10, 2) NOT NULL COMMENT 'Unit price at purchase time',
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    INDEX idx_order (order_id),
    INDEX idx_product (product_id)
) COMMENT='Order items table';

-- ============================================================
-- Table: reviews - Product reviews
-- ============================================================
CREATE TABLE reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT NOT NULL COMMENT 'Product ID',
    customer_id INT NOT NULL COMMENT 'Customer ID',
    rating INT NOT NULL COMMENT 'Rating: 1-5 stars',
    comment TEXT COMMENT 'Review content',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_product (product_id),
    INDEX idx_rating (rating)
) COMMENT='Product reviews table';

-- ============================================================
-- Insert sample data: Customers
-- ============================================================
INSERT INTO customers (name, email, phone, city, vip_level) VALUES
('张三', 'zhangsan@example.com', '13800138001', '北京', 'gold'),
('李四', 'lisi@example.com', '13800138002', '上海', 'platinum'),
('王五', 'wangwu@example.com', '13800138003', '广州', 'silver'),
('赵六', 'zhaoliu@example.com', '13800138004', '深圳', 'normal'),
('钱七', 'qianqi@example.com', '13800138005', '杭州', 'gold'),
('孙八', 'sunba@example.com', '13800138006', '成都', 'normal'),
('周九', 'zhoujiu@example.com', '13800138007', '武汉', 'silver'),
('吴十', 'wushi@example.com', '13800138008', '南京', 'normal'),
('郑十一', 'zheng11@example.com', '13800138009', '西安', 'gold'),
('王小明', 'wangxm@example.com', '13800138010', '北京', 'platinum');

-- ============================================================
-- Insert sample data: Products
-- ============================================================
INSERT INTO products (name, category, price, stock, status) VALUES
('iPhone 15 Pro', 'electronics', 8999.00, 100, 'active'),
('MacBook Pro 14', 'electronics', 14999.00, 50, 'active'),
('AirPods Pro 2', 'electronics', 1899.00, 200, 'active'),
('iPad Air', 'electronics', 4799.00, 80, 'active'),
('Apple Watch Series 9', 'electronics', 3199.00, 120, 'active'),
('华为 Mate 60 Pro', 'electronics', 6999.00, 150, 'active'),
('小米 14 Ultra', 'electronics', 5999.00, 180, 'active'),
('Nike 运动鞋', 'clothing', 899.00, 300, 'active'),
('Adidas T恤', 'clothing', 299.00, 500, 'active'),
('优衣库羽绒服', 'clothing', 599.00, 200, 'active'),
('李宁运动裤', 'clothing', 199.00, 400, 'inactive'),
('三只松鼠坚果礼盒', 'food', 168.00, 1000, 'active'),
('良品铺子零食大礼包', 'food', 128.00, 800, 'active'),
('星巴克咖啡豆', 'food', 98.00, 500, 'active'),
('《深入理解计算机系统》', 'books', 139.00, 200, 'active'),
('《算法导论》', 'books', 128.00, 150, 'active'),
('《Python编程》', 'books', 89.00, 300, 'active'),
('宜家台灯', 'home', 199.00, 250, 'active'),
('无印良品收纳盒', 'home', 79.00, 400, 'active'),
('小米空气净化器', 'home', 899.00, 100, 'discontinued');

-- ============================================================
-- Insert sample data: Orders
-- ============================================================
INSERT INTO orders (order_no, customer_id, total_amount, status, payment_method, order_date, shipped_date, delivered_date) VALUES
('ORD20240101001', 1, 10898.00, 'delivered', 'alipay', '2024-01-15 10:30:00', '2024-01-16 08:00:00', '2024-01-18 14:00:00'),
('ORD20240101002', 2, 14999.00, 'delivered', 'wechat', '2024-01-16 14:20:00', '2024-01-17 09:00:00', '2024-01-19 16:00:00'),
('ORD20240101003', 3, 1198.00, 'shipped', 'credit_card', '2024-01-18 09:15:00', '2024-01-19 10:00:00', NULL),
('ORD20240101004', 4, 6999.00, 'paid', 'alipay', '2024-01-20 16:45:00', NULL, NULL),
('ORD20240101005', 5, 3199.00, 'delivered', 'wechat', '2024-01-22 11:30:00', '2024-01-23 08:00:00', '2024-01-25 12:00:00'),
('ORD20240101006', 1, 299.00, 'cancelled', 'alipay', '2024-01-23 20:00:00', NULL, NULL),
('ORD20240101007', 6, 466.00, 'delivered', 'bank_transfer', '2024-01-25 15:30:00', '2024-01-26 09:00:00', '2024-01-28 10:00:00'),
('ORD20240101008', 7, 8999.00, 'shipped', 'credit_card', '2024-02-01 10:00:00', '2024-02-02 08:00:00', NULL),
('ORD20240101009', 8, 267.00, 'pending', NULL, '2024-02-05 18:30:00', NULL, NULL),
('ORD20240101010', 9, 5999.00, 'refunded', 'wechat', '2024-02-08 12:00:00', '2024-02-09 09:00:00', '2024-02-11 14:00:00'),
('ORD20240101011', 10, 19798.00, 'delivered', 'alipay', '2024-02-10 09:00:00', '2024-02-11 08:00:00', '2024-02-13 16:00:00'),
('ORD20240101012', 2, 1899.00, 'delivered', 'wechat', '2024-02-15 14:30:00', '2024-02-16 09:00:00', '2024-02-18 11:00:00'),
('ORD20240101013', 3, 899.00, 'paid', 'alipay', '2024-02-20 16:00:00', NULL, NULL),
('ORD20240101014', 5, 168.00, 'delivered', 'wechat', '2024-02-22 11:00:00', '2024-02-23 09:00:00', '2024-02-25 10:00:00'),
('ORD20240101015', 1, 4799.00, 'shipped', 'credit_card', '2024-02-28 10:30:00', '2024-03-01 08:00:00', NULL);

-- ============================================================
-- Insert sample data: Order Items
-- ============================================================
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 1, 1, 8999.00),   -- iPhone 15 Pro
(1, 3, 1, 1899.00),   -- AirPods Pro 2
(2, 2, 1, 14999.00),  -- MacBook Pro 14
(3, 8, 1, 899.00),    -- Nike 运动鞋
(3, 9, 1, 299.00),    -- Adidas T恤
(4, 6, 1, 6999.00),   -- 华为 Mate 60 Pro
(5, 5, 1, 3199.00),   -- Apple Watch Series 9
(6, 9, 1, 299.00),    -- Adidas T恤 (cancelled)
(7, 12, 2, 168.00),   -- 三只松鼠坚果礼盒 x2
(7, 14, 1, 98.00),    -- 星巴克咖啡豆
(7, 19, 1, 79.00),    -- 无印良品收纳盒 (已停产的不会被购买,这里用其他)
(8, 1, 1, 8999.00),   -- iPhone 15 Pro
(9, 15, 1, 139.00),   -- 《深入理解计算机系统》
(9, 16, 1, 128.00),   -- 《算法导论》
(10, 7, 1, 5999.00),  -- 小米 14 Ultra
(11, 1, 1, 8999.00),  -- iPhone 15 Pro
(11, 2, 1, 14999.00), -- MacBook Pro 14 (错误的总额,应该是23998,这里保持演示)
(11, 4, 1, 4799.00),  -- iPad Air (fix order 11: 8999+4799+6000=19798, let's adjust)
(12, 3, 1, 1899.00),  -- AirPods Pro 2
(13, 8, 1, 899.00),   -- Nike 运动鞋
(14, 12, 1, 168.00),  -- 三只松鼠坚果礼盒
(15, 4, 1, 4799.00);  -- iPad Air

-- Fix order 11 items (remove wrong entry and fix)
DELETE FROM order_items WHERE order_id = 11;
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(11, 1, 1, 8999.00),   -- iPhone 15 Pro
(11, 4, 1, 4799.00),   -- iPad Air
(11, 7, 1, 5999.00);   -- 小米 14 Ultra (8999+4799+5999=19797, close enough)

-- Update order 11 total
UPDATE orders SET total_amount = 19797.00 WHERE id = 11;

-- ============================================================
-- Insert sample data: Reviews
-- ============================================================
INSERT INTO reviews (product_id, customer_id, rating, comment) VALUES
(1, 1, 5, '非常棒的手机，拍照效果很好！'),
(1, 2, 4, '整体不错，但价格有点贵'),
(1, 10, 5, '流畅度很高，值得购买'),
(2, 2, 5, '程序员必备神器，性能强悍'),
(3, 1, 5, '降噪效果一流，音质很好'),
(3, 2, 4, '续航还可以，佩戴舒适'),
(5, 5, 4, '功能丰富，健康监测很实用'),
(6, 4, 5, '华为信号真的强，支持国货！'),
(7, 9, 3, '相机不错，但系统有点卡'),
(8, 3, 4, '穿着舒服，跑步很轻便'),
(12, 6, 5, '坚果新鲜，包装也很好'),
(15, 8, 5, '经典教材，程序员必读');

-- ============================================================
-- Create view for order statistics
-- ============================================================
CREATE VIEW v_order_stats AS
SELECT 
    c.id as customer_id,
    c.name as customer_name,
    c.vip_level,
    COUNT(o.id) as order_count,
    SUM(CASE WHEN o.status = 'delivered' THEN o.total_amount ELSE 0 END) as total_spent,
    AVG(o.total_amount) as avg_order_amount
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
GROUP BY c.id, c.name, c.vip_level;

-- ============================================================
-- Create view for product sales
-- ============================================================
CREATE VIEW v_product_sales AS
SELECT 
    p.id as product_id,
    p.name as product_name,
    p.category,
    p.price as current_price,
    COUNT(oi.id) as times_sold,
    COALESCE(SUM(oi.quantity), 0) as total_quantity,
    COALESCE(SUM(oi.quantity * oi.unit_price), 0) as total_revenue
FROM products p
LEFT JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.id, p.name, p.category, p.price;

-- Grant permissions
GRANT ALL PRIVILEGES ON ecommerce.* TO 'demo'@'%';
FLUSH PRIVILEGES;
