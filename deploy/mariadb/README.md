# MySQL Demo Database (mydb)

MySQL 示例数据库，用于 ReActSQL 的 Text-to-SQL 接入演示。

## 快速启动

```bash
cd system/mydb
docker compose up -d
```

## 连接信息

| 参数 | 值 |
|------|-----|
| Host | `localhost:3307` (外部) 或 `mydb:3306` (Docker 网络内) |
| Database | `ecommerce` |
| Username | `demo` |
| Password | `demo123` |

### 在 ReActSQL 中配置

1. 打开 ReActSQL 前端界面
2. 进入 "数据库连接" 页面
3. 点击 "添加连接"
4. 填入以下信息：
   - **连接 ID**: `mydb_demo`
   - **类型**: MySQL
   - **主机**: `mydb:3306` (如果 ReActSQL 也在 Docker 中运行)
   - **数据库**: `ecommerce`
   - **用户名**: `demo`
   - **密码**: `demo123`

## 数据库结构

### 表结构

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| `customers` | 客户信息 | id, name, email, city, vip_level |
| `products` | 商品目录 | id, name, category, price, stock, status |
| `orders` | 订单信息 | id, order_no, customer_id, total_amount, status |
| `order_items` | 订单明细 | id, order_id, product_id, quantity, unit_price |
| `reviews` | 商品评价 | id, product_id, customer_id, rating, comment |

### 视图

| 视图名 | 说明 |
|--------|------|
| `v_order_stats` | 客户订单统计 |
| `v_product_sales` | 商品销售统计 |

### 枚举字段说明

| 字段 | 枚举值 | 中文含义 |
|------|--------|----------|
| `customers.vip_level` | normal, silver, gold, platinum | 普通/白银/黄金/白金会员 |
| `products.category` | electronics, clothing, food, books, home | 电子/服装/食品/图书/家居 |
| `products.status` | active, inactive, discontinued | 在售/下架/停产 |
| `orders.status` | pending, paid, shipped, delivered, cancelled, refunded | 待支付/已支付/已发货/已签收/已取消/已退款 |
| `orders.payment_method` | alipay, wechat, credit_card, bank_transfer | 支付宝/微信/信用卡/银行转账 |

## 示例问题 (Demo Questions)

以下是可以用于演示的自然语言查询问题：

### 基础查询
1. **查询所有客户** - "列出所有客户信息"
2. **商品列表** - "显示所有在售的电子产品"
3. **订单查询** - "查看已完成的订单"

### 聚合统计
4. **销售统计** - "统计每个商品类别的销售总额"
5. **客户消费** - "查询每个客户的总消费金额"
6. **商品评分** - "计算每个商品的平均评分"

### 复杂查询
7. **VIP 客户分析** - "查询黄金和白金会员的订单数量和总消费"
8. **热销商品** - "找出销量前5的商品"
9. **城市分析** - "统计各城市的客户数量和订单总额"
10. **退款订单** - "查询所有退款的订单及客户信息"

### 业务场景
11. **库存预警** - "找出库存少于100的商品"
12. **支付方式分析** - "统计各支付方式的使用次数和金额"
13. **月度趋势** - "按月统计订单数量和销售额"
14. **高价值客户** - "找出消费超过10000元的客户"
15. **差评商品** - "查询评分低于3分的商品评价"

## 常用 SQL 示例

```sql
-- 1. 查询 VIP 客户及其订单统计
SELECT c.name, c.vip_level, COUNT(o.id) as order_count, SUM(o.total_amount) as total_spent
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
WHERE c.vip_level IN ('gold', 'platinum')
GROUP BY c.id, c.name, c.vip_level;

-- 2. 商品销售排行
SELECT p.name, p.category, SUM(oi.quantity) as sold_qty, SUM(oi.quantity * oi.unit_price) as revenue
FROM products p
JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.id, p.name, p.category
ORDER BY revenue DESC
LIMIT 10;

-- 3. 各城市客户消费分析
SELECT c.city, COUNT(DISTINCT c.id) as customers, COUNT(o.id) as orders, SUM(o.total_amount) as total_amount
FROM customers c
LEFT JOIN orders o ON c.id = o.customer_id
GROUP BY c.city
ORDER BY total_amount DESC;

-- 4. 订单状态分布
SELECT status, COUNT(*) as count, SUM(total_amount) as amount
FROM orders
GROUP BY status;
```

## 数据量

| 表 | 记录数 |
|----|--------|
| customers | 10 |
| products | 20 |
| orders | 15 |
| order_items | ~20 |
| reviews | 12 |

## 停止服务

```bash
cd system/mydb
docker compose down

# 如需清除数据
docker compose down -v
```
