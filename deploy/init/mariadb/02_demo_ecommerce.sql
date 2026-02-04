-- LUCID Demo Database: E-Commerce
-- Simple e-commerce schema for Text-to-SQL demonstrations

CREATE DATABASE IF NOT EXISTS demo_ecommerce
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE demo_ecommerce;

-- Customers table
CREATE TABLE customers (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    city VARCHAR(50),
    vip_level ENUM('normal', 'silver', 'gold', 'platinum') DEFAULT 'normal',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_city (city),
    INDEX idx_vip (vip_level)
) ENGINE=InnoDB;

-- Products table
CREATE TABLE products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(200) NOT NULL,
    category ENUM('electronics', 'clothing', 'food', 'books', 'home') NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    stock INT DEFAULT 0,
    status ENUM('active', 'inactive', 'discontinued') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- Orders table
CREATE TABLE orders (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    customer_id INT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    status ENUM('pending', 'paid', 'shipped', 'delivered', 'cancelled', 'refunded') DEFAULT 'pending',
    payment_method ENUM('alipay', 'wechat', 'credit_card', 'bank_transfer'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_status (status),
    INDEX idx_customer (customer_id)
) ENGINE=InnoDB;

-- Order items table
CREATE TABLE order_items (
    id INT PRIMARY KEY AUTO_INCREMENT,
    order_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id),
    INDEX idx_order (order_id),
    INDEX idx_product (product_id)
) ENGINE=InnoDB;

-- Reviews table
CREATE TABLE reviews (
    id INT PRIMARY KEY AUTO_INCREMENT,
    product_id INT NOT NULL,
    customer_id INT NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (customer_id) REFERENCES customers(id),
    INDEX idx_product (product_id),
    INDEX idx_rating (rating)
) ENGINE=InnoDB;

-- Insert sample data
INSERT INTO customers (name, email, city, vip_level) VALUES
('Alice Wang', 'alice@example.com', 'Beijing', 'platinum'),
('Bob Li', 'bob@example.com', 'Shanghai', 'gold'),
('Charlie Chen', 'charlie@example.com', 'Shenzhen', 'silver'),
('David Zhang', 'david@example.com', 'Guangzhou', 'normal'),
('Emma Liu', 'emma@example.com', 'Beijing', 'gold');

INSERT INTO products (name, category, price, stock, status) VALUES
('iPhone 15', 'electronics', 6999.00, 50, 'active'),
('MacBook Pro', 'electronics', 12999.00, 30, 'active'),
('Cotton T-Shirt', 'clothing', 99.00, 200, 'active'),
('Coffee Beans', 'food', 58.00, 500, 'active'),
('Python Programming', 'books', 89.00, 100, 'active');

INSERT INTO orders (order_no, customer_id, total_amount, status, payment_method) VALUES
('ORD001', 1, 6999.00, 'delivered', 'alipay'),
('ORD002', 2, 12999.00, 'shipped', 'credit_card'),
('ORD003', 1, 198.00, 'delivered', 'wechat');

INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES
(1, 1, 1, 6999.00),
(2, 2, 1, 12999.00),
(3, 3, 2, 99.00);

INSERT INTO reviews (product_id, customer_id, rating, comment) VALUES
(1, 1, 5, 'Excellent phone!'),
(2, 2, 4, 'Good performance but expensive'),
(3, 1, 5, 'Comfortable and good quality');

SELECT 'E-commerce demo database created successfully!' AS status;
