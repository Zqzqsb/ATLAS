import type { StageArch } from './comm'

export const contextArch: StageArch = {
  id: 'context',
  abstract:
    '"Context Layer" 是 4 个独立子决策的组合：以什么形态定义语义？怎么构建？存哪？怎么取回？各家差异基本都落在这 4 维。',
  principles: [
    {
      name: '语义 = 可评审的契约',
      desc: '不是隐藏在 LLM 提示词里的字符串——是 wiki / YAML / 知识图谱里可读、可 diff、可评审的资产。',
    },
    {
      name: '构建至少要"半自动"',
      desc: '纯手工建模在大 schema 下不可持续；纯自动会捕获错误概念。从 introspect 起步，让人评审 + LLM 增量补全。',
    },
    {
      name: '存储决定可移植性',
      desc: '存在产品 metadata DB → 锁死；存 git 化 YAML / Markdown → 可迁移。',
    },
    {
      name: '召回必须自适应',
      desc: '小 schema 全量塞进 prompt 最简单可靠；大 schema 才上向量；对短查询要兜底关键词。',
    },
  ],
  subQuestions: [
    /* ─────── Q1: 语义以什么形态定义？ ─────── */
    {
      id: 'define-form',
      question: '"语义"以什么形态定义？',
      why: '形态决定了表达力上限、可评审性、被 LLM 消费的成本。',
      steps: [
        {
          id: 'q1-step-1',
          name: '骨架：实体 / 字段',
          desc: '声明"哪些表 / 列要被语义化"——是契约的最小单位。',
          icon: 'i-lucide-table-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL `models[]`：每个 model 含 `tableReference` + `columns[]`，列上挂 `expression` / `type` / `description`。',
              detail: {
                summary:
                  'WrenAI 把"语义层"和"物理表"分开放——MDL 里的 model 是逻辑实体，可以重命名 / 隐藏列 / 加计算列；它通过 `table_reference` 三段式定位到源库的真实物理表。所以 model 不是物理表的镜像——而是物理表上的一层"业务视图 + 解释 + 受控暴露"。',
                bullets: [
                  {
                    label: '重命名列',
                    icon: 'i-lucide-replace',
                    accent: 'violet',
                    body: '物理列叫 `usr_id`，model 暴露成 `customer_id`（用 `expression: usr_id`）；下游 SQL 永远只看到业务名。',
                  },
                  {
                    label: '选择性暴露',
                    icon: 'i-lucide-shield',
                    accent: 'rose',
                    body: '物理表有 50 列、敏感列 `ssn` 不写进 columns[] 就完全隐藏，LLM 看不到、引擎拒查。',
                  },
                  {
                    label: '计算列',
                    icon: 'i-lucide-calculator',
                    accent: 'amber',
                    body: '`is_calculated: true` + SQL 表达式，物理表没有的派生字段（"客单价"、"近 30 日下单数"）作为一等公民出现。',
                  },
                  {
                    label: '关系列',
                    icon: 'i-lucide-git-merge',
                    accent: 'emerald',
                    body: '一个 `customer` 列把 `customer_id` 包装成 join 句柄，查询里写 `orders.customer.first_name` 引擎自动展开 JOIN。',
                  },
                ],
                closing:
                  '一句话：MDL model = 物理表上的"业务适配器"——它决定 LLM 看到什么、看不到什么、看到的叫什么。',
              },
              example: {
                lang: 'yaml',
                caption: 'models/customers/metadata.yml · 物理表 ↔ 逻辑 model 映射',
                code: `# 物理表：jaffle_shop.main.customers (50 列, 名字风格不一致)
#   usr_id INTEGER PK, fname VARCHAR, lname VARCHAR, ssn VARCHAR, ...
#
# 逻辑 model：暴露 4 列、重命名、加 1 个计算列、隐藏敏感列
name: customers
table_reference:
  catalog: jaffle_shop          # ← 源库 catalog
  schema: main                  # ← 源库 schema
  table: customers              # ← 源库表名（物理）
primary_key: customer_id

columns:
  - name: customer_id           # 业务名
    type: INTEGER
    expression: usr_id          #   ↑ 物理列名 → 简单重命名
    is_primary_key: true
    not_null: true

  - name: first_name
    type: VARCHAR
    expression: fname           # 同上：重命名

  - name: last_name
    type: VARCHAR
    expression: lname

  - name: number_of_orders      # ★ 计算列：物理表里不存在
    type: BIGINT
    is_calculated: true
    expression: |
      (SELECT COUNT(*) FROM orders
       WHERE orders.customer_id = customers.customer_id)

# 注意：物理列 ssn 没写进 columns[] → LLM / SQL 都查不到（受控暴露）`,
              },
              code: [
                { repo: 'wrenai', path: 'docs/core/reference/mdl.md', label: 'docs · mdl.md' },
              ],
              diagram: {
                kind: 'adapter',
                caption: 'WrenAI MDL · 物理表 → 逻辑 model 的局部暴露 + 转写',
                physical: {
                  label: '物理表',
                  sublabel: 'jaffle_shop.main.customers',
                  columns: [
                    { name: 'usr_id', type: 'INT PK' },
                    { name: 'fname', type: 'VARCHAR' },
                    { name: 'lname', type: 'VARCHAR' },
                    { name: 'ssn', type: 'VARCHAR', hidden: true, sensitive: true },
                    { name: 'created_at', type: 'TIMESTAMP', hidden: true },
                    { name: '… 45 列', hidden: true },
                  ],
                },
                logical: {
                  label: '逻辑 model',
                  sublabel: 'customers',
                  columns: [
                    { name: 'customer_id', kind: 'rename', from: 'usr_id', expr: 'usr_id', note: 'PK · 重命名为业务名' },
                    { name: 'first_name', kind: 'rename', from: 'fname', expr: 'fname' },
                    { name: 'last_name', kind: 'rename', from: 'lname', expr: 'lname' },
                    { name: 'number_of_orders', kind: 'computed', expr: 'COUNT(orders WHERE …)', note: '派生字段 · 物理表里不存在' },
                    { name: 'orders', kind: 'relation', expr: 'JOIN orders ON …', note: '关系列 · 写 orders.customer.* 自动展开 JOIN' },
                  ],
                },
              },
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Semantic Model YAML `tables[]`：每个 table 含 `base_table` + `dimensions[] / time_dimensions[] / facts[]`，并强制 `synonyms` / `description`。',
              detail: {
                summary:
                  'Cortex Analyst 的 semantic model 是物理表上面的一层"业务视图 + 同义词词典"——它本身不存数据，只解释怎么读物理表。',
                bullets: [
                  {
                    label: '三段式定位',
                    icon: 'i-lucide-database',
                    accent: 'emerald',
                    body: '`base_table: { database, schema, table }` 直接指向 Snowflake 真实表（例：`SALES_DB.PUBLIC.ORDERS`）。',
                  },
                  {
                    label: '物理列三桶分类',
                    icon: 'i-lucide-grid-2x2',
                    accent: 'blue',
                    body: '`dimensions` = 分类列（GROUP BY / WHERE）；`time_dimensions` = 日期/时间列（带 day/month/year grain）；`facts` = 数值列（参与聚合，但还不是"指标"——指标在 metrics 里组合 facts 得到）。',
                  },
                  {
                    label: '业务名 ↔ 物理列',
                    icon: 'i-lucide-replace',
                    accent: 'violet',
                    body: '每列的 `expr:` 引用源表物理列名（也可以是 SQL 表达式），`name:` 是业务名——和 WrenAI 的 column rename 一个意思。',
                  },
                  {
                    label: 'synonyms 是核心信号',
                    icon: 'i-lucide-bookmark',
                    accent: 'amber',
                    body: '`synonyms[]` 和 `description` 不是装饰——是 Cortex Analyst 用来匹配自然语言问题的关键。"销售额"想找到 `revenue` metric，全靠 synonyms 词典。',
                  },
                  {
                    label: '和 WrenAI 的差异',
                    icon: 'i-lucide-git-compare',
                    accent: 'rose',
                    body: 'WrenAI 把指标定义在 cubes 独立文件里；Cortex 把 metrics 直接挂在 semantic model 文件中。WrenAI 的列可以是计算列（is_calculated）；Cortex 的 fact 也能用 SQL 表达式写在 expr 里。Cortex 强制 description / synonyms 必填——这是托管产品的"质量保险"。',
                  },
                ],
                closing:
                  '一句话：Cortex semantic model = 物理表 + "业务名词典 + 三桶分类 + 同义词索引"，由托管引擎执法。',
              },
              example: {
                lang: 'yaml',
                caption: 'semantic_model.yaml · 物理表 ↔ semantic model 映射',
                code: `# 物理表：SALES_DB.PUBLIC.ORDERS
#   ORDER_ID INTEGER, USR_ID INTEGER, STATUS TEXT,
#   TOTAL_AMOUNT NUMBER, ORDER_TS TIMESTAMP, CANCELLED BOOLEAN

name: revenue_model
tables:
  - name: orders                       # 业务名（不必和物理表名一致）
    description: "客户购买订单，每行一条订单。"
    base_table:
      database: SALES_DB               # ← 物理 database
      schema: PUBLIC                   # ← 物理 schema
      table: ORDERS                    # ← 物理表名

    # ─── 分桶 1：dimensions（分类列，用于 GROUP BY / WHERE）───
    dimensions:
      - name: status
        synonyms: ["state", "订单状态", "order_state"]
        description: "订单生命周期状态。P=待付款 S=已发货 C=已取消"
        expr: status                   # ← 物理列名（这里恰好同名）
        data_type: TEXT

    # ─── 分桶 2：time_dimensions（带粒度的日期 / 时间）───
    time_dimensions:
      - name: order_date
        expr: order_ts                 # ← 物理列名 ORDER_TS
        data_type: DATE
        # Cortex 自动支持 DAY / MONTH / YEAR 粒度切换

    # ─── 分桶 3：facts（数值，参与聚合）───
    facts:
      - name: amount
        synonyms: ["金额", "revenue per order"]
        expr: total_amount             # ← 物理列名 TOTAL_AMOUNT
        data_type: NUMBER

# ─── 物理列 CANCELLED 不写进任何分桶 → semantic model 里看不到 ───
# 想用它，要么用在 metric 的 filter 里，要么显式加进 dimensions

metrics:
  - name: total_revenue
    description: "已付款订单总销售额"
    synonyms: ["销售额", "GMV", "revenue"]
    expr: orders.amount                # ← 引用上面声明的 fact
    default_aggregation: sum
    filter: "orders.status = 'paid'"   # ← 业务规则下沉到 metric 定义`,
              },
              refs: ['sf-semantic-yaml', 'sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Rich Context 表 `rc_business_context`：以 (table, column) 为 key 写 JSON 注释；不强制 schema 形态。',
              detail: {
                summary:
                  'ATLAS 走的是另一条路——不试图重新定义"语义层"，而是给原物理表打"业务注释"。物理表本身不变（不需要建语义视图、不需要 YAML），只是在另一张表里挂 5 类业务上下文。',
                bullets: [
                  {
                    label: '物理表零侵入',
                    icon: 'i-lucide-shield-check',
                    accent: 'emerald',
                    body: '物理表（例：`prod.orders`）原封不动——查询时还是 `SELECT * FROM prod.orders`，没有视图层、没有重命名、没有计算列。',
                  },
                  {
                    label: '元数据表挂注释',
                    icon: 'i-lucide-database',
                    accent: 'blue',
                    body: 'Rich Context 表 `rc_business_context` 是一张元数据表，每行一条注释：`(table_name, column_name, rc_type, rc_payload_json)`。',
                  },
                  {
                    label: '5 类业务上下文',
                    icon: 'i-lucide-tags',
                    accent: 'violet',
                    body: '`rc_type` 5 类：同义词 / enum 含义 / 单位 / 业务规则 / 关联主题——这是 ATLAS 判断"哪些业务知识值得记忆"得出的最小集。',
                  },
                  {
                    label: 'LLM 看到的是带注释的物理表',
                    icon: 'i-lucide-message-square-quote',
                    accent: 'amber',
                    body: 'LLM 写 SQL 前，linking pipeline 把相关行的 rc_payload 一起塞进 prompt——它看到的还是物理列名 `usr_id` / `total_amount`，但带上了"用户编号 / 客户ID"和"金额（人民币元）"的注释。',
                  },
                  {
                    label: '没有重命名 / 没有指标层',
                    icon: 'i-lucide-x-circle',
                    accent: 'rose',
                    body: 'LLM 直接用物理列名写 SQL，不做映射；同义词只是检索 / 消歧的桥。measures 当前还是靠 prompt + 业务规则文本表达，没形式化。',
                  },
                  {
                    label: '权衡',
                    icon: 'i-lucide-scale',
                    accent: 'slate',
                    body: '✅ 优势：零迁移成本，对已有 SQL / BI 工具完全透明。\n❌ 劣势：表达力依赖文本质量，没有引擎"执法"的硬约束。\n🎯 适合：已有大量物理表、不想重建语义层、希望快速给 LLM 注入业务知识的场景。',
                  },
                ],
                closing:
                  '一句话：ATLAS Rich Context = 给物理表"贴便利贴"——不动表、不动 SQL，只用 LLM 看得懂的注释把业务知识灌进 prompt。',
              },
              example: {
                lang: 'json',
                caption: 'rc_business_context · 一行 = 一条业务注释',
                code: `// 物理表：prod.orders（不动）
//   usr_id BIGINT, total_amount DECIMAL, status CHAR(1), ...

// Rich Context 表（一行一条）
{
  "table_name": "orders",
  "column_name": "status",
  "rc_type": "enum_meaning",
  "rc_payload": {
    "values": {
      "P": "pending — 用户已下单未付款",
      "S": "shipped — 已发货",
      "C": "cancelled — 已取消（不计入 GMV）"
    },
    "note": "C 状态的订单要从所有指标里排除"
  }
}

// 同表另一列另一类：
{
  "table_name": "orders",
  "column_name": "total_amount",
  "rc_type": "unit",
  "rc_payload": {
    "currency": "CNY",
    "scale": 2,
    "note": "已扣优惠后的金额；含税；分单位时为整数 × 100"
  }
}

// 一个表级注释（column_name = NULL）：
{
  "table_name": "orders",
  "column_name": null,
  "rc_type": "business_rule",
  "rc_payload": {
    "rule": "每个 user_id 在同一秒可能下多条订单（重复点击）；去重要按 user_id + order_ts 取首条"
  }
}

// LLM 看到的 prompt（节选）：
//   table prod.orders:
//     - usr_id BIGINT  (synonyms: 用户编号, 客户ID)
//     - total_amount DECIMAL  (单位: CNY 元；含税；C 状态需排除)
//     - status CHAR(1)  (enum: P=pending, S=shipped, C=cancelled)`,
              },
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '`semantic_models:` YAML：复用 dbt model 作为 entity；`entities` / `dimensions` / `measures` 三段式。',
              detail: {
                summary:
                  'dbt SL 是"在 dbt 已经管好物理建模的前提下，再叠一层语义层"——所以它和物理表的关系经过 dbt model 二次抽象，是唯一一家"两层映射"。',
                bullets: [
                  {
                    label: '物理层',
                    icon: 'i-lucide-database',
                    accent: 'slate',
                    body: '源库的真实表（例：`raw.orders_app`、`raw.payments_stripe`）——dbt 通过 sources.yml 声明它们存在。',
                  },
                  {
                    label: 'dbt model 层（中间适配）',
                    icon: 'i-lucide-arrow-right',
                    accent: 'blue',
                    body: '用 SQL（`models/marts/fct_orders.sql`）把物理表清洗 / 聚合 / 重命名 / join 后产出"分析表"。这一步 dbt 已经做了大量列重命名、口径统一、单位归一、剔除测试单。',
                  },
                  {
                    label: 'semantic_models 层',
                    icon: 'i-lucide-layers',
                    accent: 'violet',
                    body: '直接 `model: ref(\'fct_orders\')` 引用 dbt model——这一层不再涉及物理表，只在 dbt 的"分析表"上声明 entities / dimensions / measures。',
                  },
                  {
                    label: 'entities 是 dbt SL 特有概念',
                    icon: 'i-lucide-link',
                    accent: 'emerald',
                    body: '把 join key 提升为一等公民——一个 entity 名字（例 `customer_id`）只要在多个 semantic_model 里被声明，MetricFlow 就能自动 join，不用每次手写 condition。',
                  },
                  {
                    label: '和 WrenAI / Cortex 的关键区别',
                    icon: 'i-lucide-git-compare',
                    accent: 'amber',
                    body: 'WrenAI 和 Cortex 都是"semantic model 直接挂在物理表上"——一层映射；dbt SL 是"semantic_model → dbt model → 物理表"——两层映射，多一层 SQL 转换。',
                  },
                  {
                    label: '权衡',
                    icon: 'i-lucide-scale',
                    accent: 'rose',
                    body: '✅ 优势：dbt 已经把脏活（列重命名 / NULL 处理 / 单位 / 时区）做完了，semantic 层很纯粹。\n❌ 劣势：要先有 dbt 项目（建模税前置）；新表上线必须先写 dbt model 才能进 semantic 层。',
                  },
                ],
                closing:
                  '一句话：dbt SL 不直接看物理表——它假设你已经用 dbt 把物理表"洗干净"成 model，再在 model 上盖语义层。',
              },
              example: {
                lang: 'yaml',
                caption: 'semantic_models.yml · 物理表 → dbt model → semantic_model',
                code: `# 物理层（不写在这——在 dbt 项目的 sources.yml 声明）：
#   raw.orders_app, raw.payments_stripe, raw.users_csv ...
#
# dbt model 层 (models/marts/fct_orders.sql)：
#   SELECT
#     o.order_id,
#     o.user_id              AS customer_id,    -- 重命名
#     CAST(o.amt AS DECIMAL) AS amount,         -- 类型修正
#     o.created_at           AS ordered_at,
#     CASE p.status WHEN 'OK' THEN 'paid' END   AS order_status
#   FROM raw.orders_app o
#   LEFT JOIN raw.payments_stripe p USING (order_id)
#   WHERE o.test_flag = false                   -- 业务规则：剔除测试单
#
# semantic_models 层 ↓（不再涉及物理列名 / 业务规则）

semantic_models:
  - name: orders
    model: ref('fct_orders')              # ← 引用 dbt model（不是物理表）
    entities:
      - name: order_id
        type: primary
      - name: customer_id                 # ← 在 customers semantic_model 里也声明
        type: foreign                     #    → MetricFlow 自动 join
    dimensions:
      - name: order_status
        type: categorical
      - name: ordered_at
        type: time
        type_params:
          time_granularity: day
    measures:
      - name: order_count
        agg: count
      - name: order_total
        agg: sum
        expr: amount

# metrics.yml（独立文件，引用上面的 measures）
metrics:
  - name: total_revenue
    description: "已付款订单总销售额"
    type: simple
    type_params:
      measure: order_total`,
              },
              code: [
                { repo: 'dbt-sl', path: 'dbt_semantic_interfaces/protocols/semantic_model.py', label: 'semantic_model.py' },
              ],
              refs: ['dbt-sl-overview'],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: '`CREATE METRIC VIEW` SQL DSL：底层 `source` 查询 + `dimensions` / `measures` 写在视图定义里。',
              detail: {
                summary:
                  'Databricks Metric View 的特点是——语义层就是一个 SQL 对象，不是 YAML 文件，而是 Unity Catalog 里和表 / 视图同级的"一等公民"。',
                bullets: [
                  {
                    label: 'source 子句 = mini dbt model',
                    icon: 'i-lucide-arrow-right',
                    accent: 'blue',
                    body: '`source = ( ... )` 子句里写一段 SELECT，从物理表（例：`catalog.sales.orders_raw`）读数据，可以做清洗（WHERE 过滤 cancelled）、join、列重命名、计算列——和写一个 dbt model 几乎等价。',
                  },
                  {
                    label: 'dimensions 子句',
                    icon: 'i-lucide-grid-2x2',
                    accent: 'emerald',
                    body: '`dimensions = ( ... )` 列出可被 GROUP BY 的列（也可以现场用 SQL 表达式产出，例 `DATE_TRUNC(order_ts) AS order_date`）。',
                  },
                  {
                    label: 'measures 子句',
                    icon: 'i-lucide-sigma',
                    accent: 'violet',
                    body: '`measures = ( ... )` 列出聚合指标（`SUM(amount) AS order_total`、`AVG(amount) AS aov`）。',
                  },
                  {
                    label: 'UC 一等公民',
                    icon: 'i-lucide-shield-check',
                    accent: 'amber',
                    body: 'Metric View 注册在 UC 里（`catalog.schema.view_name`），可以被授权 / 审计 / 分享——和表共享 UC 治理。BI / Genie 看到的是 metric view 而不是物理表，用 `SELECT total_revenue, order_status FROM catalog.sales.orders_metrics` 这种"半 SQL 半语义"的查法。',
                  },
                  {
                    label: '和其他 vendor 的对比',
                    icon: 'i-lucide-git-compare',
                    accent: 'rose',
                    body: 'WrenAI = "YAML + 自家引擎"；Cortex = "YAML + 托管 metric"；dbt SL = "YAML + dbt model"；Databricks Metric View = "SQL DDL + UC object"——没有独立 YAML 层，语义就长在 SQL 里。',
                  },
                  {
                    label: '权衡',
                    icon: 'i-lucide-scale',
                    accent: 'slate',
                    body: '✅ 优势：原生 SQL 兼容，所有 BI 工具不改一行就能用；UC 治理统一。\n❌ 劣势：表达力受 SQL 约束（没有 entity / synonyms 这种语义层概念），同义词 / 业务规则得靠 column / table comment 补。',
                  },
                ],
                closing:
                  '一句话：Databricks Metric View = 把"语义层"做成一段 CREATE 语句，让它和表共享同一套 UC 权限 / 审计 / SQL 入口。',
              },
              example: {
                lang: 'sql',
                caption: 'CREATE METRIC VIEW · 物理表 → metric view',
                code: `-- 物理表：catalog.sales.orders_raw（不动）
--   order_id INT, user_id INT, status STRING,
--   amt DECIMAL, ordered_at TIMESTAMP, test_flag BOOLEAN

CREATE OR REPLACE METRIC VIEW catalog.sales.orders_metrics
WITH (
  -- ─── source 子句：从物理表读 + 清洗 + 重命名 ───
  source = (
    SELECT
      order_id,
      user_id    AS customer_id,        -- 重命名
      status,
      amt        AS amount,             -- 重命名
      ordered_at,
      DATE_TRUNC('day', ordered_at) AS order_date
    FROM catalog.sales.orders_raw
    WHERE status != 'cancelled'         -- 业务规则下沉
      AND test_flag = false             -- 剔除测试单
  ),
  -- ─── dimensions：可 GROUP BY 的维度 ───
  dimensions = (
    customer_id,
    status,
    order_date
  ),
  -- ─── measures：聚合指标 ───
  measures = (
    COUNT(order_id) AS order_count,
    SUM(amount)     AS order_total,
    AVG(amount)     AS aov
  )
);

-- 注册后的用法（SQL / BI / Genie 都用这个）：
SELECT customer_id, order_total
FROM catalog.sales.orders_metrics
WHERE order_date >= '2026-01-01'
GROUP BY customer_id
ORDER BY order_total DESC LIMIT 10;
-- ↑ 引擎自动展开为对底层物理表的 SUM(amount) GROUP BY customer_id`,
              },
              refs: ['dbx-mv-ref'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '`semantic-layer/<connection-id>/` YAML：CLI 自动 introspect 生成 model + 列 + join 图；同仓还有 `wiki/` 存自由文本业务知识（双载体）。',
              detail: {
                summary:
                  'ktx 是开源 / 本地的"上下文层"——和 wren / dbt SL 同走 YAML + git 路线，但额外强调"语义层 + wiki 双载体"：结构化指标走 semantic-layer YAML，自由文本业务规则走 wiki Markdown，两者由 MCP 统一供给 agent。',
                bullets: [
                  {
                    label: 'YAML 自动产出',
                    icon: 'i-lucide-zap',
                    accent: 'emerald',
                    body: '`ktx ingest` 跑完后 `semantic-layer/<connection-id>/` 下每个 model 一个 YAML——CLI 直接 introspect 仓库 + 检测 join 列 + 解决 chasm/fan trap，不用手写。',
                  },
                  {
                    label: '语义层 + wiki 双载体',
                    icon: 'i-lucide-layers-2',
                    accent: 'blue',
                    body: '`semantic-layer/` 装结构化（entity / metric / join）；`wiki/global/` + `wiki/user/<id>/` 装自由文本业务规则——和 ATLAS Rich Context 一个思路，但走文件而不是 DB 表。',
                  },
                  {
                    label: 'git 友好',
                    icon: 'i-lucide-git-branch',
                    accent: 'violet',
                    body: '`ktx.yaml` + `semantic-layer/` + `wiki/` 全部入 git；`.ktx/` 是本地状态 / 索引（gitignored）——和 wren / dbt SL 一样享受 PR 评审。',
                  },
                  {
                    label: 'BYO LLM · 本地优先',
                    icon: 'i-lucide-shield',
                    accent: 'amber',
                    body: '只把数据传给你配的 LLM provider（Anthropic / Vertex / Claude Code / Codex 等），ktx 自己不托管任何东西——和 Cortex / Databricks 的"云内一体"形成鲜明对比。',
                  },
                  {
                    label: '面向 Agent',
                    icon: 'i-lucide-bot',
                    accent: 'rose',
                    body: '`ktx mcp start` 起一个本地 MCP server，把 wiki + semantic-layer 暴露成工具——Claude Code / Codex / Cursor / OpenCode 直接通过 MCP 调用，agent 能拿到 approved metric 而不是临场写 SQL。',
                  },
                ],
                closing:
                  '一句话：ktx = WrenAI + ATLAS RC 的开源 + 本地版——YAML 给引擎，wiki 给 LLM，MCP 给 agent。',
              },
              example: {
                lang: 'text',
                caption: 'ktx 项目布局（典型）',
                code: `my-project/
├── ktx.yaml                          # 项目配置（providers / connections）
├── semantic-layer/
│   └── warehouse/                    # connection-id
│       ├── orders.yaml               # 一个 model 一份 YAML
│       ├── customers.yaml
│       └── relationships.yaml        # 自动检测的 join 图
├── wiki/
│   ├── global/                       # 团队共享业务规则 / 名词解释
│   │   └── refund_policy.md
│   └── user/<user-id>/               # 个人 scratch
├── raw-sources/<connection-id>/      # introspect 产物 / 报告
└── .ktx/                             # 本地状态 / 索引（gitignored）

# 用法
$ npm install -g @kaelio/ktx
$ ktx setup                # 配 LLM + DB + 拉 dbt / Notion 等已有源
$ ktx ingest               # 自动 introspect → 写 YAML + wiki
$ ktx sl "revenue"         # 语义层全文搜
$ ktx wiki "refund policy" # wiki 搜
$ ktx mcp start            # 起 MCP，给 agent 用`,
              },
              code: [
                { repo: 'ktx', path: 'README.md', label: 'README' },
                { repo: 'ktx', path: 'packages/cli/src/context', label: 'context engine' },
              ],
            },
            {
              vendor: 'OKF · GoogleCloudPlatform',
              school: 'open-context',
              primary: false,
              desc: 'OKF 把"语义"定义为 **一个 .md 文件 = 一个 concept**（YAML frontmatter + Markdown body）；无 schema 强约束，type/title/desc/resource/tags 由 producer 自由约定。',
              detail: {
                summary:
                  'OKF (Open Knowledge Format) 是 Google 在 knowledge-catalog 仓库里推出的 vendor-neutral 元数据目录格式。它故意不立 schema，而是用"一个 .md 一个 concept"做最小公约数：frontmatter `type` 字段是路由键，其余字段 KV 自由扩展；body 用标准 Markdown 加 `# Schema` / `# Examples` / `# Citations` 等约定 heading。bundle = 一棵 .md 目录树，可 git 化、可 tarball 分发。',
                bullets: [
                  {
                    label: '一个 .md = 一个 entity',
                    icon: 'i-lucide-file-text',
                    accent: 'slate',
                    body: '最小知识单元就是文件——`type: BigQuery Table` 的 .md 描述一张表，`type: Playbook` 的 .md 描述一次应急流程。`type` 值不注册，consumers MUST 容忍未知 type。',
                  },
                  {
                    label: 'Path-as-ID',
                    icon: 'i-lucide-route',
                    accent: 'violet',
                    body: '概念 ID = 文件相对 bundle 路径去 .md 后缀——git mv / refactor 全部天然支持，不会出现"rename 改 ID 然后所有 link 断"的灾难。',
                  },
                  {
                    label: 'Frontmatter = 小 schema',
                    icon: 'i-lucide-list-tree',
                    accent: 'emerald',
                    body: '`type` 必填（消费方路由键）；`title` / `description` / `resource` / `tags` 推荐；任何其它字段都能自由加，consumers 不能因 schema 升级拒绝旧 bundle。',
                  },
                  {
                    label: 'No type registry',
                    icon: 'i-lucide-unplug',
                    accent: 'amber',
                    body: '跟 MDL / Cube DSL / dbt SL 不同——OKF 不规定 type 集合，也不规定"指标" / "维度" / "关系" 字段。MetricView、Playbook、Incident 都可以是 type 之一。',
                  },
                ],
                closing:
                  '一句话：OKF 是"元数据目录的 Markdown 化"——牺牲结构化换取 vendor-neutral / Git 化 / 人和 Agent 都能读。比 MDL 弱、比 Atlas RC 强，正好占生态空位。',
              },
              example: {
                lang: 'yaml',
                caption: 'okf/bundles/ga4/tables/events_.md · 真实 OKF concept',
                code: `---
type: BigQuery Table
title: GA4 Events
description: One row per GA4 event (pageview, purchase, add_to_cart, ...) from the obfuscated e-commerce sample.
resource: bigquery://ga4-obfuscated-sample-ecommerce.analytics_249147898.events_
tags: [ga4, events, web-analytics]
timestamp: 2026-05-28T14:30:00Z
---

# Schema

| Column         | Type      | Description                                  |
|----------------|-----------|----------------------------------------------|
| event_name     | STRING    | e.g. "page_view", "purchase", "add_to_cart". |
| event_date     | DATE      | UTC date the event was logged.               |
| user_pseudo_id | STRING    | GA4 client id (pseudonymous).                |

# Joins

Joined with [users](/tables/users.md) on \`user_pseudo_id\`.

# Citations

[1] [GA4 BigQuery export schema](https://support.google.com/analytics/answer/7029846)`,
              },
              code: [
                { repo: 'okf', path: 'okf/SPEC.md', label: 'OKF SPEC v0.1' },
                { repo: 'okf', path: 'okf/bundles/ga4/tables/events_.md', label: 'bundles/ga4/tables/events_.md' },
                { repo: 'okf', path: 'okf/README.md', label: 'README (Why OKF?)' },
              ],
            },
          ],
        },
        {
          id: 'q1-step-2',
          name: '关系：表与表怎么连',
          desc: '声明 join key、cardinality、是否安全 fan-out。决定多表查询能不能正确聚合。',
          icon: 'i-lucide-git-merge',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL `relationships[]`：`models` 二元组 + `joinType` (ONE_TO_ONE / ONE_TO_MANY / MANY_TO_ONE) + `condition` SQL。引擎用它生成 join。',
              detail: {
                summary:
                  'WrenAI 用一个独立的 `relationships.yml` 顶层数组声明所有表间关系，每条关系一行；声明后还能在 columns 里把它"包装成 join 句柄"——查询里写 `orders.customer.first_name` 引擎自动展开成 JOIN，业务方完全不用懂 SQL join。',
                bullets: [
                  {
                    label: '关系本体（独立文件）',
                    icon: 'i-lucide-file-text',
                    accent: 'violet',
                    body: '`relationships.yml` 顶层一个 `relationships:` 数组，每个 entry 写 `name` + 二元 `models` + `join_type` (ONE_TO_ONE / ONE_TO_MANY / MANY_TO_ONE) + `condition` SQL 片段。',
                  },
                  {
                    label: '关系列（join 句柄）',
                    icon: 'i-lucide-link',
                    accent: 'emerald',
                    body: '在 model 的 columns[] 里加一列 `relationship: orders_customers` —— 这个列就成了"指向另一张 model 的指针"，SQL / NL2SQL 里能直接 `orders.customer.first_name` 一路点过去。',
                  },
                  {
                    label: '查询时引擎自动展开',
                    icon: 'i-lucide-shuffle',
                    accent: 'amber',
                    body: '`SELECT customer.first_name FROM orders` 由 Wren Engine 看到关系列后自动 rewrite 成 `JOIN customers ON orders.customer_id = customers.customer_id`——业务方不需要写 join condition。',
                  },
                ],
                closing:
                  '一句话：关系是一等公民——声明一次，所有 NL2SQL / SQL 都能享受"点路径"语法糖，少写 90% 的 join。',
              },
              example: {
                lang: 'yaml',
                caption: 'relationships.yml + 关系列引用',
                code: `# relationships.yml
relationships:
  - name: orders_customers
    models: [orders, customers]
    join_type: MANY_TO_ONE
    condition: orders.customer_id = customers.customer_id

# models/orders/metadata.yml (节选)
columns:
  - name: customer        # 关系列：把 customer_id 变成 join 句柄
    type: customers
    relationship: orders_customers

# 查询时：SELECT customer.first_name FROM orders   ← 引擎自动展开 JOIN`,
              },
              code: [
                { repo: 'wrenai', path: 'docs/core/reference/mdl.md', label: 'mdl · relationships' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'YAML `relationships[]`：`name` / `left_table` / `right_table` + `relationship_columns[]` + `join_type`，VQR 也可补充。',
              detail: {
                summary:
                  'Cortex Analyst 把 relationships 与 tables 同级声明在 semantic_model.yaml 里，描述两表如何 join。生成 SQL 时优先复用这里的关系，业务方不用写 join condition。',
                bullets: [
                  {
                    label: '同级声明',
                    icon: 'i-lucide-list-tree',
                    accent: 'blue',
                    body: 'relationships 和 tables 在同一份 YAML 里——一个 model 文件就是完整的"表+关系+指标"包，便于审阅。',
                  },
                  {
                    label: '复合 key 支持',
                    icon: 'i-lucide-key-round',
                    accent: 'violet',
                    body: '`relationship_columns[]` 列出 (left_column, right_column) 对，可多列复合 key——电商场景的 (tenant_id, user_id) 这种合成主键也能正确 join。',
                  },
                  {
                    label: 'join_type 细分',
                    icon: 'i-lucide-git-merge',
                    accent: 'emerald',
                    body: '支持 `inner` / `left_outer` / `many_to_one` / `one_to_one` 等——和 SQL 标准一致；`relationship_type` 还可以单独标记基数（many_to_one），用于聚合优化。',
                  },
                  {
                    label: '生成 SQL 优先复用',
                    icon: 'i-lucide-sparkles',
                    accent: 'amber',
                    body: '问"上月销售额 top 客户"时 Cortex Analyst 不会临时拼 join——而是按这里声明的 orders_to_customers 直接展开，确保 join 方向 / 类型一致。',
                  },
                ],
                closing:
                  '一句话：Cortex 的 relationships 是"约束 + 优化提示"——把语义层的 join 行为提前固化，避免 LLM 自由发挥。',
              },
              example: {
                lang: 'yaml',
                caption: 'semantic_model.yaml · relationships',
                code: `relationships:
  - name: orders_to_customers
    left_table: orders
    right_table: customers
    relationship_columns:
      - left_column: customer_id
        right_column: customer_id
    join_type: left_outer
    relationship_type: many_to_one`,
              },
              refs: ['sf-semantic-yaml'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '从 information_schema 推 FK + 命名启发式补全；关系存在 `rc_table_relations`。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '`entities`：每个 model 声明 primary / foreign / unique entity，框架按 entity 名字自动 join。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Databricks Metric View 不引入独立的 `relationships` 概念——join 子句直接内联写在 Metric View 的 `source` query 里。',
              notSupported: '不形式化"关系"为一等公民；每个 metric view 各自带 join，跨 view 共享 join 图需要重复 SQL。',
              refs: ['dbx-mv-ref'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'Context engine 自动检测 joinable columns + 解决 chasm/fan trap，关系挂在 `semantic-layer/<connection>/relationships.yaml`。',
              code: [
                { repo: 'ktx', path: 'packages/cli/src/context', label: 'context engine' },
              ],
            },
          ],
        },
        {
          id: 'q1-step-3',
          name: '指标 / 度量',
          desc: '把"销售额 / GMV / 留存率"形式化——分子分母、过滤、单位。',
          icon: 'i-lucide-sigma',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'Cubes：`cubes[]` 节点持有 `measures[]` 和 `dimensions[]`；measure 有 `expression` (SUM / COUNT / 自定义 SQL)。',
              detail: {
                summary:
                  'WrenAI 把指标抽到独立的 cubes 节点——每个 cube 是一个"分析视角"，引用一个或多个 model，把 measures（聚合）和 dimensions（切片）在同一处声明。和 model（描述实体）解耦——一个 model 可以服务多个 cube。',
                bullets: [
                  {
                    label: 'cube = 分析视角',
                    icon: 'i-lucide-layers',
                    accent: 'violet',
                    body: '`cubes/<name>/metadata.yml` 一个 cube 文件。`ref_models` 列出本 cube 用到哪些 model；measures / dimensions 写在这一处。',
                  },
                  {
                    label: 'measure 是 SQL 片段',
                    icon: 'i-lucide-sigma',
                    accent: 'amber',
                    body: 'measure 的 `expression` 可以是聚合（`SUM(orders.amount)`）也可以是带逻辑的表达式（`SUM(CASE WHEN status="paid" THEN amount END)`）——条件指标 / 复合指标都能写。',
                  },
                  {
                    label: 'dimension 也是 SQL 片段',
                    icon: 'i-lucide-grid-2x2',
                    accent: 'emerald',
                    body: 'dimensions 不限于物理列——可以是 `DATE_TRUNC(\'month\', orders.ordered_at)` 这种派生维度，也可以跨 model 引用（`customers.segment`）。',
                  },
                  {
                    label: '和 model 解耦',
                    icon: 'i-lucide-unlink',
                    accent: 'blue',
                    body: '同一个 model（如 orders）可以被多个 cube（revenue / retention / fraud）引用——指标定义不污染 model 本体，迭代更自由。',
                  },
                ],
                closing:
                  '一句话：cube 把"业务问题"变成形式化 SQL 片段——LLM 不再"临场猜聚合方式"，而是引用预先约定好的 measure 名字。',
              },
              example: {
                lang: 'yaml',
                caption: 'cubes/revenue/metadata.yml',
                code: `name: revenue
ref_models: [orders, customers]
measures:
  - name: total_revenue
    expression: "SUM(orders.amount)"
    description: "总销售额（不含已取消）"
  - name: paid_revenue
    expression: "SUM(CASE WHEN orders.status = 'paid' THEN orders.amount ELSE 0 END)"
  - name: order_count
    expression: "COUNT(DISTINCT orders.order_id)"
dimensions:
  - name: order_month
    expression: "DATE_TRUNC('month', orders.ordered_at)"
  - name: customer_segment
    expression: "customers.segment"`,
              },
              code: [
                { repo: 'wrenai', path: 'docs/core/reference/mdl.md', label: 'mdl · cubes' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: '`facts[]` / `metrics[]`：metric 通过 `expr` (SQL fragment) 引用 facts，加 `default_aggregation`。',
              detail: {
                summary:
                  'Cortex 把"指标"分两层：facts 是表上的"可聚合数值列"，metrics 才是"业务真正关心的 KPI"——metrics 通过 expr 引用 facts 并指定聚合方式。这种分层让同一个 fact（amount）能衍生多个 metric（total_revenue / paid_revenue / aov），不重复物理列。',
                bullets: [
                  {
                    label: 'facts = 可聚合的数值列',
                    icon: 'i-lucide-database',
                    accent: 'blue',
                    body: '在 table 的 `facts[]` 里声明数值物理列（`amount`、`quantity`）——还不是"指标"，只是"可以参与聚合的原料"。',
                  },
                  {
                    label: 'metrics = 业务 KPI',
                    icon: 'i-lucide-sigma',
                    accent: 'amber',
                    body: '在顶层 `metrics[]` 里写指标，用 `expr` 引用 fact（`orders.amount`），加 `default_aggregation`（sum / avg / count_distinct / median …）——LLM 看到"销售额"会精确命中这个 metric。',
                  },
                  {
                    label: 'filter 把业务规则下沉',
                    icon: 'i-lucide-filter',
                    accent: 'emerald',
                    body: '`filter: "orders.status = \'paid\'"` 让 metric 自带过滤——业务方问"销售额"，自动只算已付款订单，不需要在每个 prompt 里都重复说一遍。',
                  },
                  {
                    label: 'synonyms 喂给 NL2SQL',
                    icon: 'i-lucide-bookmark',
                    accent: 'violet',
                    body: 'metric 上挂 `synonyms: ["revenue", "GMV", "销售额"]`——多语言、多写法都能命中同一个指标。',
                  },
                ],
                closing:
                  '一句话：Cortex 的 facts 是"原材料"，metrics 是"成品菜"——预先把成品菜定义好，LLM 只需要点单不需要做菜。',
              },
              example: {
                lang: 'yaml',
                caption: 'semantic_model.yaml · metrics',
                code: `metrics:
  - name: total_revenue
    description: "总销售额（已排除已取消）"
    expr: orders.amount
    default_aggregation: sum
    synonyms: ["revenue", "sales", "GMV"]
  - name: paid_revenue
    expr: orders.amount
    default_aggregation: sum
    filter: "orders.status = 'paid'"
  - name: aov
    description: "客单价"
    expr: orders.amount
    default_aggregation: average`,
              },
              refs: ['sf-semantic-yaml'],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '`measures` (基础) → `metrics:` (派生 / ratio / cumulative)。MetricFlow 在 query 时编译成 SQL。',
              code: [
                { repo: 'metricflow', path: 'metricflow/specs/specs.py', label: 'specs.py' },
              ],
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: '`measures` 子句直接在 `CREATE METRIC VIEW` 里：`MEASURE(expr) AS name`。一处定义，全平台共享。',
              refs: ['dbx-mv-ref'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '当前 ATLAS 不形式化指标 / measure——指标含义靠 RC `business_context` 自由文本承载，依赖 LLM 在 prompt 里"读懂"。',
              notSupported: '当前不形式化 measure 节点；同义指标走 prompt + 业务规则文本表达，没有引擎"执法"的硬约束。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'Approved metric definitions 进 `semantic-layer/` YAML——agent 通过 MCP 拿"已审过的指标"，不再每次重写聚合。',
              code: [
                { repo: 'ktx', path: 'python/ktx-sl', label: 'semantic-layer planner' },
              ],
            },
          ],
        },
        {
          id: 'q1-step-4',
          name: '同义词 / 业务规则',
          desc: '"客户" = customer = client = 买家——这种自由文本规则不进语义层就只能塞 prompt。',
          icon: 'i-lucide-book-open',
          takes: [
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'YAML 强制每个字段有 `synonyms[]`；UI 引导补 `description`。Cortex Analyst 用它消歧。',
              detail: {
                summary:
                  'Cortex Analyst 把 synonyms 和 description 提到"必填字段"的高度——这是托管产品最有特色的"质量约束"。每列、每 metric 都得写同义词词典 + 业务说明，LLM 才能精准消歧。',
                bullets: [
                  {
                    label: 'synonyms = 召回信号',
                    icon: 'i-lucide-bookmark',
                    accent: 'violet',
                    body: '用户问"销售额"，Cortex Analyst 在 synonyms 词典里命中 `revenue / sales / GMV` → 直接选中 metric `total_revenue`，不用 LLM 猜。',
                  },
                  {
                    label: 'description = 业务规则提示',
                    icon: 'i-lucide-message-square-quote',
                    accent: 'amber',
                    body: '`description: "已付款订单总销售额（人民币元，不含已取消）"`——这段话喂给 LLM，让它知道单位 / 排除条件 / 含义。',
                  },
                  {
                    label: 'Studio 强校验',
                    icon: 'i-lucide-shield-check',
                    accent: 'emerald',
                    body: 'Studio UI 在保存前对空 description / 空 synonyms 给红字提示——漏写无法上线。这是托管产品的"质量保险"。',
                  },
                  {
                    label: '中英双语友好',
                    icon: 'i-lucide-languages',
                    accent: 'blue',
                    body: 'synonyms 是字符串列表，无脑塞中英文混合（`["销售额", "revenue", "GMV"]`）——多语言场景一行代价。',
                  },
                ],
                closing:
                  '一句话：synonyms + description 是 Cortex 的"防线"——逼用户把业务知识写进 schema，而不是寄希望于 LLM 自己理解。',
              },
              example: {
                lang: 'yaml',
                caption: 'synonyms / description（节选）',
                code: `dimensions:
  - name: customer_tier
    description: "客户分层：依据近 12 个月 GMV 分 P0/P1/P2/P3 四档；P0 = 头部 1%。"
    synonyms: ["客户等级", "tier", "VIP 等级"]
    expr: tier_code
metrics:
  - name: total_revenue
    description: "已付款订单总销售额（人民币元，不含已取消）"
    synonyms: ["销售额", "revenue", "sales", "GMV"]
    expr: orders.amount
    default_aggregation: sum`,
              },
              refs: ['sf-semantic-yaml'],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'enrich-context skill：扫 raw/ 文档自动补 enum 含义 / 单位 / 同义词到 MDL `description`。',
              detail: {
                summary:
                  '`enrich-context` 是 wren 的 skill——在项目已有 MDL 的基础上，扫 raw/ 里的业务文档自动补 enum 含义 / 单位 / 同义词，所有变更走 git diff 评审，不会偷偷改 MDL。',
                bullets: [
                  {
                    label: '阶段 1 · scan',
                    icon: 'i-lucide-scan-search',
                    accent: 'blue',
                    body: '扫 raw/ 目录（业务文档、wiki 导出、历史 SQL）→ 产 `gap_catalog.md` 列出哪里缺了 enum_meaning / unit / synonym / cube。',
                  },
                  {
                    label: '阶段 2 · propose',
                    icon: 'i-lucide-sparkles',
                    accent: 'amber',
                    body: 'LLM 读文档片段 + 现有 MDL，生成补全提议——具体是哪个文件、哪个字段、加什么 description。',
                  },
                  {
                    label: '阶段 3 · apply (--apply)',
                    icon: 'i-lucide-pencil',
                    accent: 'violet',
                    body: '提议被接受后，写回 MDL 的 `properties.description` / `enum_meaning` / `cubes_proposals.md`——所有变更都是文件 diff。',
                  },
                  {
                    label: '阶段 4 · review',
                    icon: 'i-lucide-git-pull-request',
                    accent: 'emerald',
                    body: 'PR 上看 git diff——"YAML 多了什么注释、改了什么 description"一目了然，不像托管产品那样"黑箱补"。',
                  },
                ],
                closing:
                  '一句话：enrich-context 把"补语义肉"变成 git PR 流程——可评审、可回滚、可追责。',
              },
              example: {
                lang: 'bash',
                caption: 'enrich-context 工作流',
                code: `# 1) 扫 raw/ 文档，找出 MDL 里缺的语义点
$ wren skill enrich-context scan
  → 产出 gap_catalog.md：
    - models/orders/columns/status: 缺 enum_meaning（5 个值未解释）
    - models/orders/columns/amount: 缺 unit（货币 / 单位 / 精度）
    - cubes/revenue/measures/aov: 缺 description（"客单价"具体口径）

# 2) LLM 提议补全 → 写回 MDL
$ wren skill enrich-context propose --apply
  → models/orders/metadata.yml diff:
    + properties:
    +   description: "订单状态枚举：P=pending,S=shipped,C=cancelled..."

# 3) git diff 评审 → 合并`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/enrich-context/SKILL.md', label: 'enrich-context · SKILL.md' },
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/enrich-context/references/gap_catalog.md', label: 'gap_catalog.md' },
              ],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Rich Context 5 类（含 enum 含义 / 单位 / 同义词）；自由文本 + 结构化字段混存。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'AI-generated comments 自动补 table / column comment；可在 catalog 内编辑。',
              refs: ['dbx-uc-ai'],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 没给"同义词"一等位置——`description` 字段写自由文本，社区惯例是塞进 `meta` tag。',
              notSupported: '不形式化 synonyms；上游消费方（NL2SQL 工具）需要自行从 description / meta 里抽取，没有 schema 强约束。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'wiki Markdown 直接写自由文本业务规则；`ktx wiki "refund policy"` 全文 + 语义双路检索，agent 通过 MCP 拿到。',
              code: [
                { repo: 'ktx', path: 'README.md', label: 'wiki layout' },
              ],
            },
          ],
        },
      ],
      commonSense:
        '正式语义层精度最高但"建模税"也最高；wiki + 列描述上手最快但精度依赖描述质量。**理想形态 = 语义层（指标 / 关系）+ wiki（业务规则、enum、单位）双载体**——结构化的留给引擎、自由文本的留给人。',
    },

    /* ─────── Q2: 怎么"建"出来？ ─────── */
    {
      id: 'build-path',
      question: '怎么"建"出来？',
      why: '建模成本是上下文层落地的最大阻力点；纯手工 / 纯自动两端都不可行。',
      steps: [
        {
          id: 'q2-step-1',
          name: 'Introspect 起骨架',
          desc: '从源库 information_schema / catalog 抓表 + 列 + 类型 → 自动产出最小契约骨架。',
          icon: 'i-lucide-scan-search',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '`generate-mdl` skill：introspect 库 schema → parse-type 类型归一化 → 写 MDL YAML；FK 推关系，无 FK 按命名约定推断。',
              detail: {
                summary:
                  'agent 跑 `wren skill generate-mdl`——把"从源库 schema 起骨架"做成五阶段流水线：连接 → 抓表列 → 类型归一化 → 写 YAML → build。FK 优先，无 FK 按命名约定推断关系，最后让用户在 PR 评审。',
                bullets: [
                  {
                    label: 'Phase 1 · 建连接',
                    icon: 'i-lucide-plug',
                    accent: 'blue',
                    body: '`wren profile add` + `wren context init` 建数据源 profile 和项目骨架。',
                  },
                  {
                    label: 'Phase 2 · introspect schema',
                    icon: 'i-lucide-scan-search',
                    accent: 'emerald',
                    body: '用 SQLAlchemy / driver / raw SQL 抓 information_schema → 拿到 tables / columns / types / fks。',
                  },
                  {
                    label: 'Phase 3 · 类型归一化',
                    icon: 'i-lucide-arrow-right-left',
                    accent: 'violet',
                    body: '`wren parse-type` 把数据库原生类型（int4 / numeric / timestamptz）映射成 MDL 标准类型（INTEGER / DECIMAL / TIMESTAMP）——一个 model 跨多个源库时类型不冲突。',
                  },
                  {
                    label: 'Phase 4 · 写 MDL + 推关系',
                    icon: 'i-lucide-file-output',
                    accent: 'amber',
                    body: '每个 dbt model 转一个 MDL `models/<name>/metadata.yml`；FK 优先推 relationships，没 FK 时按命名约定（`*_id` → 同名 model）推断 + 标记 confidence。',
                  },
                  {
                    label: 'Phase 5 · build & validate',
                    icon: 'i-lucide-check-circle',
                    accent: 'rose',
                    body: '`wren context build` 把 source-of-truth YAML 编译成 `target/mdl.json`（引擎用的 camelCase）；不通过的 schema 引用会报错——CI 卡线。',
                  },
                ],
                closing:
                  '一句话：introspect 起骨架的核心是"全自动产 schema、半自动推关系、人评审落地"——不让人手抄 50 张表，但也不放任 AI 自由发挥。',
              },
              example: {
                lang: 'bash',
                caption: 'generate-mdl 五阶段流水线',
                code: `# Phase 1: 建连接 + 选 schema
$ wren profile add --name my-pg --type postgres ...
$ wren context init

# Phase 2: introspect schema
$ python -c "from sqlalchemy import inspect; ..."
  → tables=[orders, customers, ...]; columns=[...]; fks=[...]

# Phase 3: 类型归一化
$ wren parse-type --datasource postgres --types '["int4","numeric","timestamptz"]'
  → ["INTEGER", "DECIMAL", "TIMESTAMP"]

# Phase 4: 写 MDL
  models/orders/metadata.yml      ← 物理表 + columns + primary_key
  models/customers/metadata.yml
  relationships.yml               ← 从 FK 推；无 FK 用命名约定

# Phase 5: build & validate
$ wren context build               → target/mdl.json (引擎用的 camelCase)`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/generate-mdl/SKILL.md', label: 'generate-mdl · SKILL.md' },
                { repo: 'wrenai', path: 'core/wren/src/wren/type_mapping.py', label: 'type_mapping.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Studio UI：选 `database.schema.table` → 自动生成初版 YAML（dimensions/facts 默认产出，预留同义词）。',
              detail: {
                summary:
                  'Cortex Analyst Studio 是一个向导式 UI——点几下就能从"连库选表"走到"YAML + 试问验证"，过程中按列类型自动分桶 + 强制每列填 description / synonyms。',
                bullets: [
                  {
                    label: '1) 选库 / schema / 表',
                    icon: 'i-lucide-database',
                    accent: 'blue',
                    body: 'Studio 列出当前账户可访问的 SALES_DB.PUBLIC 等三段——选中一张表就开始建模。',
                  },
                  {
                    label: '2) 自动列分桶',
                    icon: 'i-lucide-grid-2x2',
                    accent: 'emerald',
                    body: '按列类型分配："文本 → dimension / 时间戳 → time_dimension / 数值 → fact"。三桶分类的初稿一秒生成。',
                  },
                  {
                    label: '3) 强制 description + synonyms',
                    icon: 'i-lucide-shield-check',
                    accent: 'amber',
                    body: 'description 必填 / synonyms 推荐——保存前 Studio 校验空值，不让漏写上线。这是托管产品的"质量保险"。',
                  },
                  {
                    label: '4) 加 metrics / relationships',
                    icon: 'i-lucide-sigma',
                    accent: 'violet',
                    body: '在向导后续步骤里挂 metric expression / 跨表关系——也可以跳过先保存基础表。',
                  },
                  {
                    label: '5) 试问验证',
                    icon: 'i-lucide-message-circle-question',
                    accent: 'rose',
                    body: '保存前必跑——用一组业务问题（"上月销售额 top 5 客户"）试跑模型，看是否选对 metric + 维度。失败可回头改。',
                  },
                  {
                    label: '6) Save → YAML 入 git',
                    icon: 'i-lucide-save',
                    accent: 'slate',
                    body: '存为 Stage 文件 / 下载 YAML 入 git——后续维护和 git 工作流接得上，不会被 UI 锁死。',
                  },
                ],
                closing:
                  '一句话：Studio = 向导式建模 + 强校验 + 试问 + git 出口——把"业务方建语义层"的成本压到最低。',
              },
              example: {
                lang: 'text',
                caption: 'Studio 向导（流程要点）',
                code: `[1] Connect → 选 SALES_DB.PUBLIC.ORDERS
[2] 自动列分类：
      status (TEXT)        → dimension
      ordered_at (TIMESTAMP)→ time_dimension
      amount (NUMBER)      → fact
[3] 表 / 列 强制填写：
      description           (必填，校验非空)
      synonyms              (推荐，给 NL2SQL 用)
[4] 加 metrics / relationships （可与其它表 join）
[5] 试问："上个月销售额 top 5 客户"
      → 模型试跑 → 看是否选到正确 metric / 维度
[6] Save → YAML 入 Stage / 下载入 git`,
              },
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Onboarding agent：introspect → forest 分簇 → 逐表生成 RC + 关系。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'AI-generated comments：UC 后台扫表 + 元数据生成 description；用户在 Catalog Explorer 复核。',
              refs: ['dbx-uc-ai'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '`ktx ingest` 一键完成 introspect 表 / 列 + 检测 join + 生成 YAML；同时拉 dbt / Looker / Notion 已有产物，自动去重 + 标注矛盾。',
              code: [
                { repo: 'ktx', path: 'packages/cli/src/connectors', label: 'connectors' },
              ],
            },
          ],
        },
        {
          id: 'q2-step-2',
          name: 'LLM 补语义肉',
          desc: '骨架有了——同义词、enum 含义、单位、业务规则需要 LLM 扫文档 / 历史 SQL 增量补。',
          icon: 'i-lucide-sparkles',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '`enrich-context` skill：扫 raw/ 文档 → 产 gap_catalog → 写回 MDL `description` / 提议 cubes（auto-pilot）。',
              detail: {
                summary:
                  'Q2-step2 用同一个 enrich-context skill，但侧重"补语义肉"流程——把 raw/ 里散落的业务规则抽出来回写到 MDL 的 properties.description，并能从历史 SQL 反推该建哪些 cube。',
                bullets: [
                  {
                    label: 'gap_catalog 列出"哪里缺"',
                    icon: 'i-lucide-list-x',
                    accent: 'rose',
                    body: '扫 raw/ → 比对 MDL → 产 `gap_catalog.md` 标出 "models/orders/columns/status: 缺 enum_meaning（5 个值未解释）"。',
                  },
                  {
                    label: 'LLM propose 补全',
                    icon: 'i-lucide-sparkles',
                    accent: 'amber',
                    body: 'LLM 读文档 + 现有 description → 提议补"P=pending,S=shipped,C=cancelled..."，写回 properties.description。',
                  },
                  {
                    label: 'auto-pilot 提议 cube',
                    icon: 'i-lucide-bot',
                    accent: 'violet',
                    body: '识别"两个表频繁 join 算同一指标" → 自动出 `cube_proposals.md`：例如 "orders ⨝ customers 出现 47 次，建议建 revenue cube" + 可执行的 measures / dimensions。',
                  },
                  {
                    label: '用户审：accept / modify / reject',
                    icon: 'i-lucide-git-pull-request',
                    accent: 'emerald',
                    body: '所有变更都是 git diff——PR 上看 YAML 多了什么、改了什么 description，决定 accept / 修改 / reject。',
                  },
                ],
                closing:
                  '一句话：补"语义肉"不是 LLM 偷偷改 schema——而是产文档化提议 → 走 PR → 人决策。',
              },
              example: {
                lang: 'markdown',
                caption: 'cube_proposals.md（auto-pilot 产物）',
                code: `# Cube proposals

## revenue (suggested)
**Why:** 历史 SQL 查询里 \`orders\` ⨝ \`customers\` 出现 47 次，
        都是为了算 SUM(amount) by customer_segment。

ref_models: [orders, customers]
measures:
  - total_revenue: SUM(orders.amount)
  - paid_revenue: SUM(CASE WHEN status='paid' THEN amount END)
dimensions:
  - customer_segment: customers.segment
  - order_month: DATE_TRUNC('month', orders.ordered_at)

→ 用户审：accept / modify / reject`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/enrich-context/SKILL.md', label: 'enrich-context · SKILL.md' },
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/enrich-context/references/cube_proposals.md', label: 'cube_proposals.md' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Studio 内 LLM 提议 synonyms / description；可结合 query history 学常用过滤值。',
              detail: {
                summary:
                  'Studio 在保存前调 LLM 给每列 / metric 自动提议 description + synonyms（用户可改）。同时 Cortex Analyst 会扫该账户的 query history，提取常用 filter 值 / time grain，作为 metric 的 default_aggregation 提示。',
                bullets: [
                  {
                    label: 'LLM 提议描述 / 同义词',
                    icon: 'i-lucide-sparkles',
                    accent: 'violet',
                    body: '保存前 Studio 调 LLM——给每列、每 metric 提议 description + synonyms 草稿；用户审改后才能保存。',
                  },
                  {
                    label: 'query history 反推',
                    icon: 'i-lucide-history',
                    accent: 'amber',
                    body: '扫该账户最近的 SQL 查询 → 哪些 filter 值高频（status=\'paid\' 出现 9k 次）→ 提议 metric default_aggregation 和 default filter。',
                  },
                  {
                    label: '强校验前置',
                    icon: 'i-lucide-shield-check',
                    accent: 'emerald',
                    body: '空 description / 空 synonyms → 红字阻断保存——逼业务方"补完再走"。',
                  },
                ],
                closing:
                  '一句话：Studio 把 LLM 当"草稿员 + 历史侦察员"——产物给人审，不偷偷写。',
              },
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'Rich Context Generation：5 类 RC 由 sub-agent 各自负责，写回 `rc_business_context`。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'UC AI 自动写 comment；Genie 在对话上下文里二次推断业务含义。',
              refs: ['dbx-uc-ai', 'dbx-genie'],
            },
            {
              vendor: 'Oracle Select AI',
              school: 'managed-cloud',
              desc: 'AI Catalog enrichment：自动给资产打描述 / 标签。',
              refs: ['oracle-ai-enrich'],
            },
          ],
        },
        {
          id: 'q2-step-3',
          name: '复用已有产出',
          desc: '团队往往已有 dbt / LookML / SQL view——别让人"再写一遍"。',
          icon: 'i-lucide-import',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '`dbt` skill：读 manifest.json + catalog.json，按 adapter→datasource 映射转 MDL；保留 model 名字 / 关系。',
              detail: {
                summary:
                  '团队往往已经有 dbt 项目——别让人重写。`wren skill dbt` 直接读 dbt 已经产好的 manifest.json + catalog.json，按 adapter 转成 MDL，dbt model 的名字 / 关系全部保留。',
                bullets: [
                  {
                    label: 'dbt 原生产物',
                    icon: 'i-lucide-package',
                    accent: 'blue',
                    body: '`dbt parse` 产 target/manifest.json（model + 列 + 类型 + 引用关系）；`dbt docs generate` 产 catalog.json（实际数据库列类型）。两份 JSON 都是 dbt 标准。',
                  },
                  {
                    label: 'adapter 映射',
                    icon: 'i-lucide-arrow-right-left',
                    accent: 'violet',
                    body: 'dbt adapter (postgres / bigquery / snowflake) → MDL data_source；adapter 决定类型映射规则（adapter 内置类型字典）。',
                  },
                  {
                    label: 'model + relationships 保留',
                    icon: 'i-lucide-link',
                    accent: 'emerald',
                    body: '每个 dbt model 转一个 MDL `models/<name>/metadata.yml`（保名字）；dbt schema test (`relationships: ...`) 转成 MDL relationships。',
                  },
                  {
                    label: '一行命令',
                    icon: 'i-lucide-terminal',
                    accent: 'amber',
                    body: '`wren skill dbt --project ./my_dbt_project --output ./my_wren_project`——不到 10 秒一个完整 wren 项目就出来了。',
                  },
                ],
                closing:
                  '一句话：复用 dbt 是性价比最高的入口——dbt 团队建模税已经付过了，wren 顺手把它升级成语义层。',
              },
              example: {
                lang: 'bash',
                caption: 'dbt → MDL 一行流',
                code: `# 1) 在 dbt 项目里 build manifest
$ cd my_dbt_project
$ dbt parse                       # 产 target/manifest.json
$ dbt docs generate               # 产 target/catalog.json

# 2) 让 wren 读这两份 → MDL
$ wren skill dbt --project ./my_dbt_project --output ./my_wren_project

  → models/dim_customer/metadata.yml      ← 来自 dbt model dim_customer
  → models/fct_orders/metadata.yml        ← 来自 dbt model fct_orders
  → relationships.yml                     ← 来自 dbt relationships test`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/dbt.py', label: 'dbt.py' },
              ],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: '`dlt-connector` skill：HubSpot / Stripe / Salesforce → DuckDB → introspect → MDL（SaaS 也能建模）。',
              detail: {
                summary:
                  '`wren skill dlt-connector` 把 SaaS 数据（HubSpot / Stripe / Salesforce）也纳入语义层——四阶段流水线：dlt 抽到 DuckDB → introspect → 自动产 MDL → build。',
                bullets: [
                  {
                    label: 'Phase 1 · dlt extract',
                    icon: 'i-lucide-download',
                    accent: 'blue',
                    body: '`pip install dlt[duckdb,hubspot]` + 一个 pipeline.py → 把 HubSpot REST API 数据 ETL 进本地 hubspot.duckdb（dlt 自带增量 / schema evolution）。',
                  },
                  {
                    label: 'Phase 2 · introspect DuckDB',
                    icon: 'i-lucide-scan-search',
                    accent: 'emerald',
                    body: '`introspect_dlt.py` 扫 DuckDB schema → 19 tables / 234 columns / 12 inferred FKs（dlt 自动建的 _dlt_load_id 等系统列也能识别）。',
                  },
                  {
                    label: 'Phase 3 · auto-generate MDL',
                    icon: 'i-lucide-file-output',
                    accent: 'violet',
                    body: '`wren skill dlt-connector --duckdb hubspot.duckdb --output ./wren_hubspot` → 完整 wren 项目（model + relationships + profile）一键生成。',
                  },
                  {
                    label: 'Phase 4 · build & verify',
                    icon: 'i-lucide-play-circle',
                    accent: 'amber',
                    body: '跑一条真实 SQL（`SELECT COUNT(*) FROM contacts WHERE country = \'JP\'`）验证 → 通过就能进入 NL2SQL。',
                  },
                ],
                closing:
                  '一句话：SaaS 数据也能上语义层——dlt 解决"取数 + 落地"，wren 解决"语义 + NL2SQL"。',
              },
              example: {
                lang: 'bash',
                caption: 'dlt-connector 四阶段',
                code: `# Phase 1: dlt extract
$ pip install dlt[duckdb,hubspot]
$ python pipeline.py              # HubSpot → ./hubspot.duckdb

# Phase 2: introspect DuckDB
$ python introspect_dlt.py ./hubspot.duckdb
  → 19 tables, 234 columns, 12 inferred FKs

# Phase 3: auto-generate Wren project
$ wren skill dlt-connector --duckdb ./hubspot.duckdb --output ./wren_hubspot

# Phase 4: build & verify
$ cd wren_hubspot && wren context build
$ wren --sql "SELECT COUNT(*) FROM contacts WHERE country = 'JP'"`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/dlt-connector/SKILL.md', label: 'dlt-connector · SKILL.md' },
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/dlt-connector/scripts/introspect_dlt.py', label: 'introspect_dlt.py' },
              ],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              primary: true,
              desc: '原生：`semantic_models` 直接挂在 dbt model 之上，零额外建模税。',
              detail: {
                summary:
                  'dbt SL 的"复用"是天然的——semantic_models: 节点本来就和 dbt model 在同一个项目里，`model: ref(\'fct_orders\')` 直接挂上去，零额外建模税。代价是要先有 dbt 项目。',
                bullets: [
                  {
                    label: '同仓 zero-import',
                    icon: 'i-lucide-link',
                    accent: 'emerald',
                    body: '不需要"转换"——`semantic_models:` 在同一个 dbt 项目里，直接引用 `ref(\'fct_orders\')`。',
                  },
                  {
                    label: '复用 dbt model',
                    icon: 'i-lucide-recycle',
                    accent: 'blue',
                    body: 'entities / dimensions / measures 复用 dbt model 的列——物理表清洗已经被 dbt 管好了，semantic 层只声明语义。',
                  },
                  {
                    label: '前提：要有 dbt 项目',
                    icon: 'i-lucide-alert-triangle',
                    accent: 'amber',
                    body: '没 dbt 项目就走不到这一步——dbt SL 只服务 dbt 用户，不像 wren 那样从 raw 库 / SaaS 起步。',
                  },
                ],
                closing:
                  '一句话：dbt SL 的复用 = "我已经管好物理建模了，再叠一层语义"——如果你已经在 dbt 生态，零额外成本。',
              },
              code: [
                { repo: 'dbt-sl', path: 'dbt_semantic_interfaces/parsing/dir_to_model.py', label: 'dir_to_model.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              desc: '可从已有 view / 物化视图起，YAML 引用 `base_table`；VQR 复用历史业务 SQL。',
              detail: {
                summary:
                  'Cortex 复用既有 SQL 资产有两条路：`base_table` 可以指向已存在的 view / materialized view（避免复制数据），VQR 把团队已审过的"问题-SQL"对入库（避免每次重新生成）。',
                bullets: [
                  {
                    label: 'base_table 指向 view',
                    icon: 'i-lucide-eye',
                    accent: 'blue',
                    body: '`base_table.table` 不限于物理表——可以是 view / materialized view，让 ETL 团队已经清洗好的资产被 semantic model 直接引用。',
                  },
                  {
                    label: 'VQR · 问题-SQL 对',
                    icon: 'i-lucide-bookmark-check',
                    accent: 'amber',
                    body: 'Verified Query Repository = 团队审核过的"问题→SQL"集合。新问题来时先检索 VQR——命中就直接复用历史 SQL，命中不到才进入生成流程。',
                  },
                  {
                    label: '两层精度保险',
                    icon: 'i-lucide-shield-check',
                    accent: 'emerald',
                    body: '高频问题（"上月销售额"）走 VQR 有人工审过保证；长尾问题走 LLM 生成——VQR 命中率随时间增长，整体精度向 100% 收敛。',
                  },
                ],
                closing:
                  '一句话：base_table = 复用 SQL 资产；VQR = 复用历史问答——双管齐下把"重复造轮子"的成本压到零。',
              },
              refs: ['sf-vqr'],
            },
          ],
        },
        {
          id: 'q2-step-4',
          name: '人工评审落地',
          desc: '半自动产物必须过人——这是最后的精度闸门，决定语义可不可信。',
          icon: 'i-lucide-git-pull-request',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL 走 git PR；`onboarding` skill 把 generate → enrich → review 串成一条 wizard。',
              detail: {
                summary:
                  '`wren skill onboarding` 是一条端到端 wizard——把"从零到能问问题"压成 8 个 phase。每个 phase 产物都是文件，PR 评审就是评审 git diff，团队 / 个人都能跑。',
                bullets: [
                  {
                    label: 'Phase 0–2 · 准备',
                    icon: 'i-lucide-check-circle',
                    accent: 'slate',
                    body: '环境检查（python / git / DB driver）→ `wren profile add`（数据源连接）→ `wren context init`（项目骨架）。',
                  },
                  {
                    label: 'Phase 3–5 · 自动建模',
                    icon: 'i-lucide-cog',
                    accent: 'blue',
                    body: 'generate-mdl（introspect → MDL）→ context build（产 target/mdl.json）→ memory index（写 LanceDB 向量）。',
                  },
                  {
                    label: 'Phase 6 · 补语义肉（可选）',
                    icon: 'i-lucide-sparkles',
                    accent: 'amber',
                    body: 'enrich-context 扫 raw/ 文档补 enum / 单位 / 同义词——可在 onboarding 内完成或之后单独跑。',
                  },
                  {
                    label: 'Phase 7 · 试问验证',
                    icon: 'i-lucide-message-circle-question',
                    accent: 'violet',
                    body: '问 "上月销售额前 5 客户" → 看是否生成正确 SQL → 不通过就回头改 description / synonyms。',
                  },
                  {
                    label: 'Git 与否两条路',
                    icon: 'i-lucide-git-pull-request',
                    accent: 'emerald',
                    body: '开 git = 团队协作 + 历史可追溯（PR 评审 MDL 改动）；纯本地 = 个人 / 试用，文件直接在工作目录。',
                  },
                ],
                closing:
                  '一句话：onboarding 把"从零到第一条 SQL"的所有步骤标准化——每一步可看 / 可改 / 可回滚。',
              },
              example: {
                lang: 'bash',
                caption: 'onboarding 一条龙',
                code: `$ wren skill onboarding

  Phase 0: 环境检查 ✓
  Phase 1: profile 添加 → my-pg ✓
  Phase 2: context init → ./my_project ✓
  Phase 3: generate-mdl
            → models/orders/metadata.yml
            → models/customers/metadata.yml
            → relationships.yml
  Phase 4: context build → target/mdl.json ✓
  Phase 5: memory index → .wren/memory/ ✓
  Phase 6: enrich-context (可选)
  Phase 7: ask "上个月销售额前 5 客户" → SQL 验证 ✓

# 用户在 git diff 上评审：
$ git diff --stat
  models/orders/metadata.yml      | 42 +++++++++
  models/customers/metadata.yml   | 28 ++++++
  relationships.yml               |  9 +++`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/skills_content/onboarding/SKILL.md', label: 'onboarding · SKILL.md' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Studio UI 审批 + 试问验证；YAML 也可下载走 git。',
              detail: {
                summary:
                  'Cortex Analyst 的人评审入口是 Studio "试问"——保存前必须用真实业务问题验证模型是否选对 metric / 维度，通过后存成 Semantic View 或下载 YAML 入 git。',
                bullets: [
                  {
                    label: '试问 = 强校验',
                    icon: 'i-lucide-message-circle-question',
                    accent: 'amber',
                    body: '保存前先问一组业务问题（"销售额 top 5 客户" / "上月新增订单"）——看 Cortex 是否选对 metric / 维度 / filter。失败可回头改。',
                  },
                  {
                    label: 'Semantic View 落地',
                    icon: 'i-lucide-database',
                    accent: 'blue',
                    body: '通过后存为 Snowflake `CREATE SEMANTIC VIEW` object——SQL / BI 都能直接 `SELECT FROM SEMANTIC_VIEW(...)`，UC 治理统一。',
                  },
                  {
                    label: 'YAML 入 git',
                    icon: 'i-lucide-git-pull-request',
                    accent: 'emerald',
                    body: '下载 YAML 进 git——和 dbt SL / Cube 一致的"语义即代码"工作流，团队 PR 评审、CI 校验都能接上。',
                  },
                ],
                closing:
                  '一句话：Studio 的"试问"是托管产品里少有的硬关卡——逼用户在保存前用真实问题对自己的语义模型进行 e2e 验证。',
              },
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: '走 dbt 项目本身的 PR / CI 流程；`dbt parse` + 单测保证不破。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Genie / Catalog UI 内编辑；缺乏 git diff 路径。',
              refs: ['dbx-genie'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '管理 UI 审批 RC 草稿；批准后写入 `rc_business_context`。',
            },
          ],
        },
      ],
      commonSense:
        '**先 introspect 起骨架，再用 LLM 补语义肉，再让人在 PR 上评审**——三段式落地阻力最小。任何一段省略都会出大问题（纯自动 → 概念错；纯手工 → 跑不通；不评审 → 错也没人发现）。',
    },

    /* ─────── Q3: 上下文存在哪里？ ─────── */
    {
      id: 'storage',
      question: '上下文存在哪里？',
      why: '存储介质决定可迁移性、可审计性、备份恢复策略、以及语义谁说了算。',
      steps: [
        {
          id: 'q3-step-1',
          name: '契约本体',
          desc: 'YAML / DSL / SQL 视图——这是"被评审的那份"。',
          icon: 'i-lucide-file-code-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'MDL YAML 文件，按 model / cube 拆，进 git；wren CLI 直接读文件系统。',
              detail: {
                summary:
                  'wren 项目就是一个目录——按子目录拆分（每个 model / cube 一个 metadata.yml），source-of-truth YAML 入 git，runtime / 编译产物 gitignore 掉。这种组织方式直接享受 git 的 PR / blame / revert / branch 能力。',
                bullets: [
                  {
                    label: 'source 进 git',
                    icon: 'i-lucide-git-branch',
                    accent: 'emerald',
                    body: '`models/<name>/metadata.yml` / `cubes/<name>/metadata.yml` / `relationships.yml` / `instructions.md` / `queries.yml` —— 这些是契约本体，必须入 git。',
                  },
                  {
                    label: 'runtime 不进 git',
                    icon: 'i-lucide-folder-x',
                    accent: 'rose',
                    body: '`.wren/` 是运行时状态（LanceDB 索引、cache）；`target/mdl.json` 是 build 产物（camelCase 给引擎用）——都 gitignored，可重建。',
                  },
                  {
                    label: '子目录拆分',
                    icon: 'i-lucide-folder-tree',
                    accent: 'violet',
                    body: '每个 model / cube 一个独立子目录——大项目 50+ 表也不会变成单一巨型 YAML。文件级粒度让 PR 评审能聚焦。',
                  },
                  {
                    label: '人机两份格式',
                    icon: 'i-lucide-arrow-left-right',
                    accent: 'amber',
                    body: 'YAML 给人写（snake_case + 注释）；`target/mdl.json` 给引擎读（camelCase + 紧凑）。Build 是单向编译，不要直接改 mdl.json。',
                  },
                ],
                closing:
                  '一句话：wren 项目 = 一个 git repo——契约 / 解释 / few-shot / 文档全在一起，团队协作和工程基础设施完全打通。',
              },
              example: {
                lang: 'text',
                caption: 'wren 项目目录',
                code: `my_project/
├── wren_project.yml             # 项目元数据（schema_version: 3, ...）
├── models/
│   ├── orders/metadata.yml      # table_reference 模式
│   ├── customers/metadata.yml
│   └── revenue_summary/
│       ├── metadata.yml         # ref_sql 模式
│       └── ref_sql.sql
├── views/
│   └── monthly_revenue/metadata.yml
├── cubes/
│   └── revenue/metadata.yml
├── relationships.yml
├── instructions.md              # 业务/操作指南，给 LLM 读
├── queries.yml                  # NL-SQL 例子（few-shot）
├── .wren/                       # gitignored — LanceDB 索引
└── target/mdl.json              # gitignored — 编译产物`,
              },
              code: [
                { repo: 'wrenai', path: 'docs/core/reference/mdl.md', label: 'mdl · file layout' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Semantic Model YAML 存在 Snowflake Stage；也作为 Semantic View object 落库。',
              detail: {
                summary:
                  'Cortex Analyst 的契约存储有两种形态——Stage 上的 YAML 文件 + 数据库 object 化的 Semantic View，可以单独用也可以双轨并行。生产推荐"YAML 走 git + Semantic View 让 SQL 也能用"双管齐下。',
                bullets: [
                  {
                    label: 'a) Stage 存 YAML',
                    icon: 'i-lucide-file-code',
                    accent: 'blue',
                    body: '`PUT file://semantic_model.yaml @MY_DB.PUBLIC.SEM_STAGE`——文件存在 stage 里，REST API 通过 `semantic_model_file: "@stage/path/model.yaml"` 引用。CI / git 同步友好。',
                  },
                  {
                    label: 'b) Semantic View object',
                    icon: 'i-lucide-database',
                    accent: 'violet',
                    body: '`CREATE SEMANTIC VIEW MY_DB.PUBLIC.REVENUE_MODEL ...` 把语义提升为 UC 一等公民——SQL 可以 `SELECT FROM SEMANTIC_VIEW(MY_DB.PUBLIC.REVENUE_MODEL METRICS revenue ...)`。',
                  },
                  {
                    label: '双轨并行',
                    icon: 'i-lucide-git-merge',
                    accent: 'amber',
                    body: '生产推荐 (a)+(b)——YAML 进 git 走 PR 工程；Semantic View 让 SQL / BI 也能直接消费语义层。两份本质同源，build 时一致。',
                  },
                  {
                    label: 'UC 治理统一',
                    icon: 'i-lucide-shield-check',
                    accent: 'emerald',
                    body: 'Stage / Semantic View 都受 Snowflake UC 权限 / 审计 / data masking——和表共享治理面。',
                  },
                ],
                closing:
                  '一句话：Cortex 把"语义层文件"和"语义层 object"做成等价两面——选其一也行，双轨更稳。',
              },
              example: {
                lang: 'sql',
                caption: '把 YAML 落地为 Semantic View object',
                code: `-- 1) YAML 放在 stage
PUT file://semantic_model.yaml @MY_DB.PUBLIC.SEM_STAGE;

-- 2) Cortex Analyst REST 调用时引用：
--    semantic_model_file: "@MY_DB.PUBLIC.SEM_STAGE/semantic_model.yaml"

-- 3) 同时 / 或者 提升为 Semantic View
CREATE OR REPLACE SEMANTIC VIEW MY_DB.PUBLIC.REVENUE_MODEL
  TABLES ( orders AS SALES_DB.PUBLIC.ORDERS PRIMARY KEY (order_id) )
  DIMENSIONS ( orders.status, orders.order_date AS DATE_TRUNC('day', ordered_at) )
  METRICS ( orders.revenue AS SUM(amount) );

-- 4) 然后 SQL 也能直接查：
SELECT * FROM SEMANTIC_VIEW(
  MY_DB.PUBLIC.REVENUE_MODEL
  METRICS revenue
  DIMENSIONS status, order_date
);`,
              },
              refs: ['sf-semantic-yaml', 'sf-semantic-views'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '`rc_business_context` MariaDB 表（仓内表，非 git）。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'YAML 跟 dbt project 一起进 git；编译产物存 dbt artifacts。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Metric View 是 UC 内一等 object（不是文件）；CREATE 语句可以纳 git。',
              refs: ['dbx-mv-ref'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '`semantic-layer/<connection-id>/*.yaml` + `wiki/global/*.md` 都进 git；`.ktx/` 装本地索引 / 状态（gitignored）。',
              code: [
                { repo: 'ktx', path: 'README.md#L168-L184', label: 'project layout' },
              ],
            },
          ],
        },
        {
          id: 'q3-step-2',
          name: '检索索引',
          desc: '召回需要 embeddings / FTS——索引是契约的"派生物"，可以重建。',
          icon: 'i-lucide-database',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'memory store：MDL → schema_indexer → embedding；持久化在 Qdrant / 本地。',
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/schema_indexer.py', label: 'schema_indexer.py' },
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/store.py', label: 'memory/store.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Cortex Search service：托管的混合检索，索引存仓内。',
              refs: ['sf-cortex-search'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'MariaDB 原生 VECTOR 列 + HNSW；索引和契约同一仓。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: '内置向量服务（Vector Search）+ Genie 的检索层。',
              refs: ['dbx-genie'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '本地 `.ktx/` 装 wiki + semantic-layer 的全文 + embedding 双索引；命中合并，跑在 ktx daemon 上。',
              code: [
                { repo: 'ktx', path: 'python/ktx-daemon', label: 'ktx-daemon' },
              ],
            },
          ],
        },
        {
          id: 'q3-step-3',
          name: '运行期 cache / 临时数据',
          desc: 'session 状态、推理 trace、近期 query history——次要资产，丢了能重建。',
          icon: 'i-lucide-archive',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'memory store 也吃 query_history / few-shot；按 namespace 隔离。',
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/store.py', label: 'memory/store.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'VQR (Verified Query Repository) 存仓内 stage；conversation state 在 Cortex 服务侧。',
              refs: ['sf-vqr'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'session / trace 存 MariaDB；planner cache 同库。',
            },
          ],
        },
      ],
      commonSense:
        '**契约存 git，索引和 cache 存仓内 / 本地**——前者保证可评审、可移植、可审计；后者保证检索快、可增量重建。把契约存进产品 metadata DB 是技术债（迁移要导数）。',
    },

    /* ─────── Q4: 怎么"取回"上下文？ ─────── */
    {
      id: 'recall',
      question: '怎么"取回"上下文？',
      why: '召回策略直接影响 SQL 生成质量和延迟；错的策略会让大 schema 完全不可用。',
      steps: [
        {
          id: 'q4-step-1',
          name: '体量自适应：能全塞就全塞',
          desc: 'schema 小（≤30k token）整个塞进 prompt——最简单、最可靠、最便宜。',
          icon: 'i-lucide-package-2',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: '默认行为：MDL 体量小就直接全量注入 prompt；超阈值才走向量召回。',
              detail: {
                summary:
                  'WrenAI 在召回上的核心智慧是"能全塞就全塞"——memory.schema_indexer 用一个 30,000 字符（约 8K token）的字符数阈值来判断 schema 是否能整段注入 prompt，低于阈值就跳过向量检索。',
                bullets: [
                  {
                    label: '阈值：30K 字符',
                    icon: 'i-lucide-ruler',
                    accent: 'blue',
                    body: '`SCHEMA_DESCRIBE_THRESHOLD = 30_000`——大约 8K token，比一般 LLM context window 占比小，留足空间给问题 + few-shot。',
                  },
                  {
                    label: '为什么用字符不用 token',
                    icon: 'i-lucide-text-cursor-input',
                    accent: 'amber',
                    body: '字符数和具体 tokenizer 无关——换 LLM / 换 tokenizer 不用重调阈值；token 估算是字符数 ÷ 3.5 这种粗估。',
                  },
                  {
                    label: '为什么全塞优于向量',
                    icon: 'i-lucide-sparkles',
                    accent: 'emerald',
                    body: 'LLM 能看到完整结构（join 路径 / 关系 / cube）——不是被切成的 fragment。简单可靠，对小中型项目精度更高。',
                  },
                  {
                    label: '超阈值的退路',
                    icon: 'i-lucide-arrow-down-circle',
                    accent: 'violet',
                    body: '超 30K 字符时调用 `retrieve_top_k(query, k=20)` 走 LanceDB 向量召回——只在必要时才用复杂检索。',
                  },
                ],
                closing:
                  '一句话：能塞就塞 = 工程上的"奥卡姆剃刀"——不要为大 schema 才需要的复杂召回方案让小 schema 也付代价。',
              },
              example: {
                lang: 'python',
                caption: 'wren · memory/schema_indexer.py（节选）',
                code: `# ~30K chars ≈ ~8K tokens. 低于这个阈值整段塞 prompt
# 优于 embedding search —— LLM 能看到完整结构（join 路径、关系）
# 而不是被切成的 fragment。
SCHEMA_DESCRIBE_THRESHOLD = 30_000

def should_full_inject(manifest: dict) -> bool:
    text = describe_schema(manifest)
    return len(text) < SCHEMA_DESCRIBE_THRESHOLD

# 调用方：
if should_full_inject(mdl):
    ctx = describe_schema(mdl)        # 全量
else:
    ctx = retrieve_top_k(query, k=20) # 向量召回`,
              },
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/schema_indexer.py', lines: [26, 50], label: 'schema_indexer.py · L26' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Semantic Model 不大时整体注入；用户在 Studio 里能看到上下文 token 占比。',
              refs: ['sf-cortex-overview'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: '小集群 forest 直接全量；大 schema 才选择性 ground。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: 'CLI 同时建全文 + embedding 索引；当 wiki + semantic-layer 体量小时，MCP 直接整段返给 agent。',
              code: [
                { repo: 'ktx', path: 'packages/cli/src/context', label: 'context engine' },
              ],
            },
          ],
        },
        {
          id: 'q4-step-2',
          name: '向量召回',
          desc: 'schema_items embed → cosine top-k——大 schema 的主力召回方式。',
          icon: 'i-lucide-radar',
          takes: [
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'memory.embeddings 把 model / column / cube 各自 embed；top-k 取相邻。',
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/embeddings.py', label: 'memory/embeddings.py' },
              ],
            },
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Cortex Search service：托管语义检索，按需混合 BM25。',
              refs: ['sf-cortex-search'],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'MariaDB VECTOR + HNSW；linking 阶段并发召回多种 RC。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Vector Search service；Genie 内部混合检索。',
              refs: ['dbx-genie'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '`ktx sl "revenue"` / `ktx wiki "..."` 走 ktx-daemon：BM25 + 向量 hybrid，落地到本地 `.ktx/` 索引。',
              code: [
                { repo: 'ktx', path: 'python/ktx-daemon', label: 'ktx-daemon' },
                { repo: 'ktx', path: 'README.md#L154-L166', label: 'sl / wiki commands' },
              ],
            },
          ],
        },
        {
          id: 'q4-step-3',
          name: '关键词 / FTS 兜底',
          desc: '短查询 ("销售额") 向量会退化——名字精确匹配反而要靠 BM25 / FTS。',
          icon: 'i-lucide-search',
          takes: [
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Cortex Search 自带混合（dense + lexical）；同义词 YAML 喂给词表。',
              refs: ['sf-cortex-search'],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              desc: '社区 / 自定义：可在 retrieval pipeline 加 BM25 节点；默认以 dense 为主。',
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'MariaDB FULLTEXT 索引并行召回；与向量结果合并。',
            },
            {
              vendor: 'Databricks UC',
              school: 'managed-cloud',
              desc: 'Vector Search 支持 hybrid (text + dense)；Genie 内部融合。',
              refs: ['dbx-genie'],
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '`ktx sl` / `ktx wiki` 默认就是 FTS + 语义双路并行——短查询有 BM25 兜底；MCP 工具同步暴露给 agent。',
              code: [
                { repo: 'ktx', path: 'README.md#L154-L166', label: 'first commands' },
              ],
            },
          ],
        },
        {
          id: 'q4-step-4',
          name: '验证集召回（VQR / few-shot）',
          desc: '历史已审过的"问题→SQL"对，比 schema 召回精度更高，是 measure 复杂场景的保险。',
          icon: 'i-lucide-bookmark-check',
          takes: [
            {
              vendor: 'Snowflake Cortex Analyst',
              school: 'managed-cloud',
              primary: true,
              desc: 'Verified Query Repository (VQR)：审核过的 SQL 入库；新问题先检索 VQR 再决定是否生成。',
              refs: ['sf-vqr'],
            },
            {
              vendor: 'WrenAI',
              school: 'semantic-layer',
              primary: true,
              desc: 'seed_queries：把 NL→SQL 例子 embed 进 memory；retrieval 阶段优先 few-shot。',
              code: [
                { repo: 'wrenai', path: 'core/wren/src/wren/memory/seed_queries.py', label: 'memory/seed_queries.py' },
              ],
            },
            {
              vendor: 'ATLAS',
              school: 'agentic',
              desc: 'session / 历史 SQL 入库；linking 时召回相似历史 SQL 作 few-shot。',
            },
            {
              vendor: 'dbt SL',
              school: 'semantic-layer',
              desc: 'dbt SL 自身不做 VQR / few-shot——把"问题→SQL"的 example 库交给消费方（Hex / Tableau / 自建 BI）。',
              notSupported: 'dbt SL 自身不维护 NL→SQL 的 example 库；这层职责被外包给上游消费方，是"语义层只管语义、不管自然语言"的取舍。',
            },
            {
              vendor: 'ktx',
              school: 'open-context',
              desc: '没有专门的 VQR——但 `semantic-layer/` 的 approved metric 起类似作用：agent 通过 MCP 拿"已审过的指标"，本质上是"不再写聚合 SQL"而非"复用 SQL"。',
              code: [
                { repo: 'ktx', path: 'python/ktx-sl', label: 'semantic-layer planner' },
              ],
            },
          ],
        },
      ],
      commonSense:
        '**体量自适应 + 多路兜底**：小 schema 全量、大 schema 走 hybrid（FTS + 向量），对短查询尤其要让 FTS 兜底——单一向量召回在表名精确匹配时反而退化。**VQR / few-shot 是大杀器**——把"做对过一次的"重用，比"每次都重新生成"准 10 倍。',
    },
  ],
  insights: [
    {
      icon: 'i-lucide-git-pull-request',
      title: '语义即代码 = 上下文层成熟度的标志',
      body: '能否走 git diff 评审 / 复用 dbt 产出 / 在不同环境间迁移——这些"工程基本面"决定了上下文层的天花板，远大于"用了什么 embedding"。',
    },
    {
      icon: 'i-lucide-layers',
      title: '语义层 + Wiki 双载体',
      body: '指标 / 关系 / measure 形式化进语义层（让引擎执法）；enum / 单位 / 业务规则进 wiki（让 LLM 看懂）。强迫所有东西进同一种载体，会要么过度工程，要么写不出。',
    },
    {
      icon: 'i-lucide-recycle',
      title: '召回失败比 SQL 错误更危险',
      body: '召回错了，LLM 在错的上下文里写出"看起来对"的 SQL——这种错最难发现。检索召回率应被作为一级指标监控，而不是只看 SQL 准确率。',
    },
  ],
  matrix: {
    cols: ['契约形态', '构建主线', '存储', '召回'],
    rows: [
      { vendor: 'WrenAI', school: 'semantic-layer', cells: ['MDL YAML', '手工/LLM/dbt/dlt', 'Git', '体量自适应+向量'] },
      { vendor: 'Snowflake Cortex', school: 'managed-cloud', cells: ['Semantic Views', '云内 + UI', '仓内', '内置 Cortex Search'] },
      { vendor: 'ATLAS', school: 'agentic', cells: ['Rich Context (JSON)', 'introspect+LLM', '仓内表 / 向量', '向量+精确'] },
      { vendor: 'dbt SL', school: 'semantic-layer', cells: ['SL YAML', '手工', 'Git', '由消费方决定'] },
      { vendor: 'Cube', school: 'semantic-layer', cells: ['Cube DSL', '手工', 'Git / Cloud DB', 'AI API'] },
      { vendor: 'Databricks UC', school: 'managed-cloud', cells: ['Metric Views', '云内', '仓内', '内置'] },
      { vendor: 'Fabric DA', school: 'managed-cloud', cells: ['Business Semantics', '云内', '仓内', '内置'] },
      { vendor: 'ktx', school: 'open-context', cells: ['SL YAML + wiki MD', 'CLI introspect + 已有源', 'Git + .ktx 索引', 'BM25 + 向量 hybrid'] },
      { vendor: 'OKF · GoogleCloudPlatform', school: 'open-context', cells: ['OKF .md (YAML+MD)', 'enrichment_agent (BQ/Web)', 'Git (文件即资产)', 'FTS / embed / link 任意'] },
    ],
  },
}
