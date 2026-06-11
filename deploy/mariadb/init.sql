-- ============================================================
-- ATLAS Demo MySQL Database
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
    vip_level ENUM('normal', 'silver', 'gold', 'platinum') DEFAULT 'normal' COMMENT 'VIP level: normal=Basic, silver=Silver, gold=Gold, platinum=Platinum',
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
    category VARCHAR(50) NOT NULL COMMENT 'Category: electronics, clothing, food, books, home',
    price DECIMAL(10, 2) NOT NULL COMMENT 'Unit price',
    stock INT DEFAULT 0 COMMENT 'Stock quantity',
    status ENUM('active', 'inactive', 'discontinued') DEFAULT 'active' COMMENT 'Status: active=On Sale, inactive=Delisted, discontinued=Discontinued',
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
        COMMENT 'Order status: pending=Pending Payment, paid=Paid, shipped=Shipped, delivered=Delivered, cancelled=Cancelled, refunded=Refunded',
    payment_method ENUM('alipay', 'wechat', 'credit_card', 'bank_transfer') COMMENT 'Payment method: alipay=Alipay, wechat=WeChat Pay, credit_card=Credit Card, bank_transfer=Bank Transfer',
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
('Alice Wang', 'alice@example.com', '13800138001', 'Beijing', 'gold'),
('Bob Li', 'bob@example.com', '13800138002', 'Shanghai', 'platinum'),
('Charlie Zhang', 'charlie@example.com', '13800138003', 'Guangzhou', 'silver'),
('David Zhao', 'david@example.com', '13800138004', 'Shenzhen', 'normal'),
('Eva Qian', 'eva@example.com', '13800138005', 'Hangzhou', 'gold'),
('Frank Sun', 'frank@example.com', '13800138006', 'Chengdu', 'normal'),
('Grace Zhou', 'grace@example.com', '13800138007', 'Wuhan', 'silver'),
('Henry Wu', 'henry@example.com', '13800138008', 'Nanjing', 'normal'),
('Iris Zheng', 'iris@example.com', '13800138009', 'Xian', 'gold'),
('Jack Wang', 'jack@example.com', '13800138010', 'Beijing', 'platinum');

-- ============================================================
-- Insert sample data: Products
-- ============================================================
INSERT INTO products (name, category, price, stock, status) VALUES
('iPhone 15 Pro', 'electronics', 8999.00, 100, 'active'),
('MacBook Pro 14', 'electronics', 14999.00, 50, 'active'),
('AirPods Pro 2', 'electronics', 1899.00, 200, 'active'),
('iPad Air', 'electronics', 4799.00, 80, 'active'),
('Apple Watch Series 9', 'electronics', 3199.00, 120, 'active'),
('Huawei Mate 60 Pro', 'electronics', 6999.00, 150, 'active'),
('Xiaomi 14 Ultra', 'electronics', 5999.00, 180, 'active'),
('Nike Running Shoes', 'clothing', 899.00, 300, 'active'),
('Adidas T-Shirt', 'clothing', 299.00, 500, 'active'),
('Uniqlo Down Jacket', 'clothing', 599.00, 200, 'active'),
('Li-Ning Track Pants', 'clothing', 199.00, 400, 'inactive'),
('Mixed Nuts Gift Box', 'food', 168.00, 1000, 'active'),
('Snack Variety Pack', 'food', 128.00, 800, 'active'),
('Starbucks Coffee Beans', 'food', 98.00, 500, 'active'),
('Computer Systems: A Programmer''s Perspective', 'books', 139.00, 200, 'active'),
('Introduction to Algorithms', 'books', 128.00, 150, 'active'),
('Python Programming', 'books', 89.00, 300, 'active'),
('IKEA Desk Lamp', 'home', 199.00, 250, 'active'),
('MUJI Storage Box', 'home', 79.00, 400, 'active'),
('Xiaomi Air Purifier', 'home', 899.00, 100, 'discontinued');

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
(3, 8, 1, 899.00),    -- Nike Running Shoes
(3, 9, 1, 299.00),    -- Adidas T-Shirt
(4, 6, 1, 6999.00),   -- Huawei Mate 60 Pro
(5, 5, 1, 3199.00),   -- Apple Watch Series 9
(6, 9, 1, 299.00),    -- Adidas T-Shirt (cancelled)
(7, 12, 2, 168.00),   -- Mixed Nuts Gift Box x2
(7, 14, 1, 98.00),    -- Starbucks Coffee Beans
(7, 19, 1, 79.00),    -- MUJI Storage Box
(8, 1, 1, 8999.00),   -- iPhone 15 Pro
(9, 15, 1, 139.00),   -- Computer Systems book
(9, 16, 1, 128.00),   -- Introduction to Algorithms
(10, 7, 1, 5999.00),  -- Xiaomi 14 Ultra
(11, 1, 1, 8999.00),  -- iPhone 15 Pro
(11, 2, 1, 14999.00), -- MacBook Pro 14 (wrong total, should be 23998, kept for demo)
(11, 4, 1, 4799.00),  -- iPad Air (fix order 11: 8999+4799+6000=19798, adjusted)
(12, 3, 1, 1899.00),  -- AirPods Pro 2
(13, 8, 1, 899.00),   -- Nike Running Shoes
(14, 12, 1, 168.00),  -- Mixed Nuts Gift Box
(15, 4, 1, 4799.00);  -- iPad Air

-- Fix order 11 items (remove wrong entry and fix)
DELETE FROM order_items WHERE order_id = 11;
INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(11, 1, 1, 8999.00),   -- iPhone 15 Pro
(11, 4, 1, 4799.00),   -- iPad Air
(11, 7, 1, 5999.00);   -- Xiaomi 14 Ultra (8999+4799+5999=19797, close enough)

-- Update order 11 total
UPDATE orders SET total_amount = 19797.00 WHERE id = 11;

-- ============================================================
-- Insert sample data: Reviews
-- ============================================================
INSERT INTO reviews (product_id, customer_id, rating, comment) VALUES
(1, 1, 5, 'Amazing phone, excellent camera quality!'),
(1, 2, 4, 'Overall great, but a bit pricey'),
(1, 10, 5, 'Very smooth performance, worth buying'),
(2, 2, 5, 'A must-have for developers, powerful performance'),
(3, 1, 5, 'Top-notch noise cancellation, great sound quality'),
(3, 2, 4, 'Decent battery life, comfortable to wear'),
(5, 5, 4, 'Feature-rich, health monitoring is very useful'),
(6, 4, 5, 'Excellent signal strength, great build quality!'),
(7, 9, 3, 'Camera is good, but system can be laggy'),
(8, 3, 4, 'Comfortable to wear, lightweight for running'),
(12, 6, 5, 'Fresh nuts, great packaging'),
(15, 8, 5, 'Classic textbook, must-read for programmers');

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
