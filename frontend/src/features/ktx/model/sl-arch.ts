import type { SlArch } from './ktx'

export const slArch: SlArch = {
  id: 'sl',
  input: {
    label: 'sl_query · measures + dimensions + filters',
    note: 'Agent 通过 MCP sl_query 或 ktx sl 提交声明式查询',
    example: '{ measures: ["orders.total_revenue"], dimensions: ["customers.region"] }',
  },
  loader: {
    title: 'SourceLoader · 双层加载',
    items: [
      { name: '_schema/*.yaml', desc: 'manifest shards → project_manifest_entry' },
      { name: '*.yaml (standalone)', desc: '有 sql/table 字段 → 独立 source' },
      { name: 'overlay merge', desc: '无 sql/table → overlay 合并到 manifest + validate_overlay' },
      { name: 'descriptions 优先级', desc: 'user → ai → dbt → db' },
    ],
  },
  joinGraph: {
    title: 'JoinGraph · Dijkstra 最小连接树',
    algo: 'resolve_join_tree(source_refs, root=anchor)',
    note: '邻接表 + 别名表；支持复合键 ("a, b" on-clause)；自动选 anchor source（基于 measure/dim grain）',
  },
  plannerSteps: [
    { name: '列可见性校验', desc: 'dimensions 自动 _qualify_bare_column 限定 source' },
    { name: 'measures 解析', desc: '预定义 / ad-hoc / derived → 拓扑排序处理 derived' },
    { name: 'segments 应用', desc: 'query-time segments + 列引用校验' },
    { name: 'anchor 选择', desc: '基于 measure/dim grain + include_empty 选根 source' },
    { name: 'join 树解析', desc: 'JoinGraph root=anchor 解析最小连接树' },
    { name: 'fan-out / chasm 检测', desc: 'measures 分组成 MeasureGroup + aggregate_locality' },
    { name: 'filter 分类', desc: 'WHERE / HAVING 分离' },
    { name: 'ResolvedPlan 输出', desc: 'ResolvedJoin · ResolvedColumn · OrderByClause' },
  ],
  generator: {
    paths: [
      { name: 'Path A · simple', desc: '单层 SELECT/JOIN/WHERE/GROUP BY/HAVING/ORDER BY' },
      { name: 'Path B · locality', desc: 'fan-out 时按 measure_group 拆 CTE 后 union' },
      { name: 'native sql:', desc: '原生 SQL source 用 CTE 内嵌、保持原样' },
    ],
    transpile: 'sqlglot _transpile(dialect)：postgres scaffold → bigquery/snowflake/mysql/duckdb',
    note: 'ExpressionParser 按连接原生方言解析 expr（BigQuery INTERVAL / Snowflake DATEADD 不丢 token）',
  },
  validate: [
    { name: 'orphan join', desc: 'join 边两端 source 都存在' },
    { name: 'grain 合法性', desc: 'measure grain 与 dimension grain 兼容' },
    { name: 'join 列覆盖', desc: 'SQL join 覆盖所有声明关系' },
    { name: '连通分量', desc: '所有 source 在 join graph 中连通' },
    { name: 'duplicate measure', desc: '同 source 上 expr+filter+segments 完全等价报错' },
  ],
  sqlBefore: `-- 声明式查询（Agent 不写物理表名）
SELECT customers.region,
       orders.total_revenue
FROM orders
WHERE orders.order_date >= '2025-01-01'`,
  sqlAfter: `-- planner + generator 展开后（fan-out 安全 · 自动 join · 方言转译）
WITH orders AS (
  SELECT o.order_id, o.total, c.region
  FROM warehouse.public.orders o
  JOIN warehouse.public.customers c
    ON o.customer_id = c.customer_id
  WHERE o.order_date >= DATE '2025-01-01'
)
SELECT region,
       SUM(total) AS total_revenue
FROM orders
GROUP BY region`,
  insights: {
    input: 'Agent 只提交 measures + dimensions + filters——不写 join、不写物理表名、不处理 fan-out。语义层引擎是"正确 SQL"的唯一真相源。',
    plan: [
      { icon: 'i-lucide-git-merge', title: 'JoinGraph 自动解 join', body: '声明的关系 + join key 在 JoinGraph 上 Dijkstra 找最小连接树；Agent 不需要知道 customers→orders 怎么连。' },
      { icon: 'i-lucide-scissors', title: '最小切片', body: '只加载本次查询命中的 sources + 关系，不把整个 semantic-layer/ 塞进 planner——大 schema 下依然轻。' },
    ],
    fanout: [
      { icon: 'i-lucide-alert-triangle', title: 'fan-out / chasm trap', body: 'planner 检测 fan-out 和 chasm trap，把 measures 分组成 MeasureGroup；generator Path B 按 group 拆 CTE 再 union——避免经典 BI 多对多重复计数。' },
    ],
  },
}
