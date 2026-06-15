/**
 * Databricks deck — L1 module data (per-flow internal architecture).
 *
 * Each module is grounded in `WiseCat/.claude/skills/research/results/databricks_uc_metric_views.yaml`
 * + the cited official docs. `refs` arrays attach evidence to specific boxes.
 */
import type { AccentKey, NamedItem, Insight } from './architecture'

export interface RefItem extends NamedItem {
  refs?: string[]
}

/** Metric View — relation + metric model. */
export interface MetricViewArch {
  id: string
  source: { label: string; note: string; refs: string[] }
  yamlSections: { name: string; desc: string; refs?: string[] }[]
  joinModes: { name: string; desc: string; refs?: string[] }[]
  modelingFeatures: RefItem[]
  /** what 'metric view as UC object' implies */
  ucProps: RefItem[]
  yamlExample: string
  rewrittenSql: string
  insights: {
    input: string
    model: Insight[]
    object: Insight[]
  }
}

/** Agent Metadata — column-level rich context. */
export interface AgentMetadataArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  fields: { name: string; desc: string; example?: string; refs?: string[] }[]
  yamlExample: string
  /** how Genie consumes the metadata */
  consumers: RefItem[]
  /** how Genie does retrieval given this metadata (rule/keyword) */
  retrievalNote: string
  insights: {
    input: string
    field: Insight[]
    consume: Insight[]
  }
}

/** Genie Space — NL2SQL agent orchestration. */
export interface GenieArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  curation: { name: string; desc: string; refs?: string[] }[]
  flow: { name: string; desc: string; refs?: string[] }[]
  guards: RefItem[]
  explain: RefItem[]
  feedback: RefItem[]
  /** undisclosed internals (D-graded) */
  undisclosed: { name: string; desc: string }[]
  sampleConvo: { user: string; interp: string; sql: string; verified?: boolean }
  insights: {
    input: string
    curate: Insight[]
    runtime: Insight[]
    feedback: Insight[]
  }
}

/** UC + Lakehouse Federation. */
export interface UcFedArch {
  id: string
  ucNamespace: { fqn: string; note: string; refs: string[] }
  /** what UC objects are valid `source` for metric view */
  sources: RefItem[]
  /** federation connectors */
  connectors: { group: string; icon: string; items: string; refs?: string[] }[]
  multiSource: { points: string[]; refs: string[] }
  insights: {
    input: string
    namespace: Insight[]
    federate: Insight[]
  }
}

/** Policy Runtime — row filter / column mask propagation. */
export interface PolicyArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  /** policies live on base tables */
  baseTablePolicies: { name: string; desc: string; refs?: string[] }[]
  /** the propagation rule (the key constraint) */
  propagation: { points: string[]; refs: string[] }
  /** object-level ACL on metric view itself */
  objectAcl: RefItem[]
  /** D-graded: policy-aware generation NOT supported */
  notPolicyAware: { points: string[]; refs?: string[] }
  ddlExample: string
  insights: {
    input: string
    base: Insight[]
    propagate: Insight[]
  }
}

/** Metric View Materialization. */
export interface MvMaterializeArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  setup: { name: string; desc: string; refs?: string[] }[]
  rewrite: { points: string[]; refs: string[] }
  refresh: RefItem[]
  cardinalityHints: RefItem[]
  insights: {
    input: string
    materialize: Insight[]
    rewrite: Insight[]
  }
}

export interface DbxModuleData {
  id: string
  accent: AccentKey
  metricView?: MetricViewArch
  agentMetadata?: AgentMetadataArch
  genie?: GenieArch
  ucFed?: UcFedArch
  policy?: PolicyArch
  mvMaterialize?: MvMaterializeArch
}

const metricViewArch: MetricViewArch = {
  id: 'metric-view',
  source: {
    label: 'CREATE METRIC VIEW · YAML',
    note: '通过 Catalog Explorer UI 或 SQL DDL 创建；语法 = YAML 规约嵌入 DDL，UC 内置 YAML validation 做语法 / 字段校验',
    refs: ['S2', 'S4'],
  },
  yamlSections: [
    { name: 'version', desc: 'metric view YAML 版本号', refs: ['S4'] },
    { name: 'source', desc: '三段式 catalog.schema.table（任意 UC 表类资产，含 foreign tables / system tables）', refs: ['S5'] },
    { name: 'joins', desc: 'star / snowflake · 嵌套 joins · using / on · cardinality + rely 提示', refs: ['S4'] },
    { name: 'dimensions', desc: '维度列（含派生 expr / display_name / synonyms / format）', refs: ['S3', 'S4'] },
    { name: 'measures', desc: '聚合公式（SUM / COUNT(DISTINCT) …）· 一次定义，任意 dimension 重写', refs: ['S2', 'S4'] },
    { name: 'filter', desc: '默认过滤条件（口径前置）', refs: ['S4'] },
  ],
  joinModes: [
    { name: 'star', desc: '事实 + 多维度直接 join（推荐：避免 fanout）', refs: ['S4'] },
    { name: 'snowflake', desc: '维度可嵌套 join（嵌套 joins 字段）', refs: ['S4'] },
    { name: 'cardinality hints', desc: 'at_most_one_match / one_to_one · 优化提示约束 SQL 形状', refs: ['S4'] },
    { name: 'rely', desc: 'rely=true 表示用户保证唯一性，引擎可消除冗余 join', refs: ['S4'] },
  ],
  modelingFeatures: [
    { name: 'metric formula', desc: '指标公式集中定义（收入 = SUM(amount)）', refs: ['S2'] },
    { name: 'dimension synonyms', desc: '同义词字段挂在 dimension 上，供 Genie 经关键词召回', refs: ['S3'] },
    { name: 'display_name / format', desc: '展示语义（Genie / BI 直接消费）', refs: ['S3'] },
    { name: 'token saving', desc: '把多表 + 口径压缩成单一语义对象，减少注入 LLM 的 schema', refs: ['S7'] },
  ],
  ucProps: [
    { name: 'CREATE / ALTER / DROP', desc: '标准 SQL DDL · 走 UC 生命周期', refs: ['S2'] },
    { name: 'GRANT / REVOKE', desc: 'metric view 自身可对象级授权', refs: ['S2'] },
    { name: 'governance', desc: '列在 UC business semantics 主入口（与 catalog 一同治理）', refs: ['S1'] },
    { name: 'cross-catalog source', desc: 'source 可跨 catalog / schema 三段式引用', refs: ['S2'] },
  ],
  yamlExample: `-- DDL: CREATE METRIC VIEW
CREATE OR REPLACE METRIC VIEW main.sales.revenue
WITH METRIC_VIEW = $$
version: 0.1
source: main.sales.fact_orders
joins:
  - name: dim_customer
    source: main.sales.dim_customer
    using: [customer_id]
    cardinality: at_most_one_match     # 提示
dimensions:
  - name: region
    expr: dim_customer.region
    synonyms: [area, territory]        # → Agent Metadata
    display_name: 销售区域
measures:
  - name: total_revenue
    expr: SUM(fact_orders.amount)      # 写一次
filter: fact_orders.is_canceled = false
$$;`,
  rewrittenSql: `-- 用户在 Genie 问 "Q1 各区域营收"
-- → metric view 引擎自动重写为：
SELECT dim_customer.region                AS 销售区域,
       SUM(fact_orders.amount)             AS total_revenue
FROM   main.sales.fact_orders
LEFT JOIN main.sales.dim_customer
       ON fact_orders.customer_id = dim_customer.customer_id
WHERE  fact_orders.is_canceled = false
  AND  fact_orders.order_date >= '2025-01-01'
  AND  fact_orders.order_date <  '2025-04-01'
GROUP BY 1`,
  insights: {
    input: 'Metric View 是 Databricks 业务语义的承载形式：YAML 是定义语法、不是存储介质——它最终是 UC 的原生对象，受 DDL / GRANT / lineage 全套治理，而不是某个本地配置文件。',
    model: [
      { icon: 'i-lucide-shapes', title: '关系 + 指标模型', body: 'measures（聚合公式）+ dimensions + joins（star/snowflake，含 cardinality / rely 优化提示）+ filter（默认口径）显式建模——查询时用户选 measure/dimension 分组，引擎自动生成正确聚合 SQL。' },
      { icon: 'i-lucide-scissors', title: 'token saving by design', body: '指标定义把多表 + 口径压缩成一个语义对象；Genie 文档明确建议用 metric view 替代裸表来"保持在上下文上限内"——上下文压缩内嵌在建模里。' },
    ],
    object: [
      { icon: 'i-lucide-database', title: 'Metric view = UC 原生对象', body: '不是 YAML 文件存 stage，也不是应用元数据库——CREATE METRIC VIEW 就是 UC DDL，权限、生命周期、跨 catalog 引用都走 UC 一套；YAML 只是定义语法。' },
      { icon: 'i-lucide-edit', title: '更新靠人工编辑', body: 'schema 漂移时未见自动检测 / 自动同步证据；语义定义需要人工 CREATE OR REPLACE 重建——把"语义稳定性"作为人为承诺。' },
    ],
  },
}

const agentMetadataArch: AgentMetadataArch = {
  id: 'agent-metadata',
  input: {
    label: 'Metric View · Table · View',
    note: 'agent metadata 挂在已存在的 UC 对象上（metric view 的 dimension/measure 或普通表的列）',
    refs: ['S3'],
  },
  fields: [
    { name: 'synonyms', desc: '同义词列表（Genie 通过用户输入的术语关键词命中字段）', example: '[area, territory]', refs: ['S3', 'S7'] },
    { name: 'display_name', desc: '展示名（Genie / BI 渲染）', example: '销售区域', refs: ['S3'] },
    { name: 'description', desc: '业务描述 · LLM 可见上下文', example: '客户所在大区', refs: ['S3'] },
    { name: 'format', desc: '数值 / 时间格式', example: '#,##0.00', refs: ['S3'] },
    { name: 'sample_values (UC tables)', desc: '可在底表加 sample / enum 注解 · 由 Genie 消费', refs: ['S3'] },
  ],
  yamlExample: `# Metric View YAML 内的 dimension（agent metadata 内联）
dimensions:
  - name: region
    expr: dim_customer.region
    synonyms: [area, territory, "大区"]
    display_name: 销售区域
    description: 客户的销售大区，按总部划分（华东/华北/华南/西部）
    format: text

  - name: order_month
    expr: date_trunc('month', fact_orders.order_date)
    synonyms: [month, 月份]
    display_name: 订单月份
    format: yyyy-MM`,
  consumers: [
    { name: 'Genie Space', desc: 'synonyms 帮 Genie 经关键词命中字段（rule_or_keyword_retrieval）', refs: ['S3', 'S7'] },
    { name: 'Catalog Explorer', desc: 'description / display_name 渲染元数据浏览', refs: ['S3'] },
    { name: '第三方 LLM 工具', desc: 'agent metadata 是 UC 元数据，外部工具可读', refs: ['S3'] },
  ],
  retrievalNote: 'Genie 的上下文召回方式 = 规则 / 关键词召回（synonyms 文本匹配 + example SQL 文本匹配）；公开文档未见对 schema / 指标做向量 embedding 检索的描述',
  insights: {
    input: '富上下文以"列级注解"形态挂在已存在的 UC 对象上——不是另起炉灶的 prompt 库，而是直接补在 metric view 与底表的 dimension/column 上，由 UC 一同治理。',
    field: [
      { icon: 'i-lucide-tags', title: 'Synonyms = 召回的载体', body: 'Genie 没有对 schema 做向量召回，靠的是 synonyms + example SQL 的关键词匹配；用户怎么说话与字段怎么命名之间的桥梁，由人工写在 synonyms 里。' },
      { icon: 'i-lucide-text', title: '人工注解为主', body: 'synonyms / display_name / description / format 全部需要人工编写——富上下文的"准"靠人工策划，没有自动学习闭环。' },
    ],
    consume: [
      { icon: 'i-lucide-share-2', title: 'UC 元数据可被复用', body: 'agent metadata 不绑死 Genie：第三方 LLM 工具同样可读 UC 元数据消费 description / synonyms——把语义注解从应用层抽到平台元数据层。' },
    ],
  },
}

const genieArch: GenieArch = {
  id: 'genie',
  input: {
    label: 'Genie Space · NL2SQL Agent',
    note: 'Space 是 Genie 的工作单元：选定一组 metric view / table 作为可见数据 + curation 资料；用户经 UI 或 Conversation API 进入',
    refs: ['S7', 'S10', 'S13'],
  },
  curation: [
    { name: 'Selected assets', desc: '一个 space 包含一组精选的 metric view / table（建议数量受限以保持上下文规模）', refs: ['S7'] },
    { name: 'Example SQL', desc: '人工策划的 NL→SQL 样例 · Genie 经文本匹配召回', refs: ['S7'] },
    { name: 'Text instructions', desc: '"未指定时间范围时请澄清" 等业务级指令（澄清行为来源）', refs: ['S7'] },
    { name: 'Trusted asset / verified query', desc: '审核后命中即标 verified answer · 形成反馈闭环', refs: ['S8'] },
    { name: 'Benchmark', desc: '一组评估问题用于离线测路由 / 生成准确性', refs: ['S8'] },
  ],
  flow: [
    { name: '路由 (route)', desc: '依 example SQL / synonyms 把问题路由到合适 metric view 或 table', refs: ['S7', 'S8'] },
    { name: '生成 SQL', desc: '直接生成最终 SQL（端到端 SQL 生成 · 未披露独立 query IR）', refs: ['S7'] },
    { name: '执行', desc: '在 SQL Warehouse、当前 session、当前用户身份下实际执行（read-only）', refs: ['S10'] },
    { name: '返回', desc: 'NL 解释 + SQL + 结果；命中 verified 标记 verified answer', refs: ['S8'] },
  ],
  guards: [
    { name: 'YAML validation', desc: 'Catalog Explorer 创建 metric view 时内置 YAML 语法 / 字段校验', refs: ['S2'] },
    { name: 'Read-only', desc: 'Genie 生成的 SQL 始终为只读（禁止 DDL/DML）', refs: ['S10'] },
    { name: 'Execution check', desc: '在当前 session 实际执行 SQL 取回结果 · benchmark 离线评估准确性', refs: ['S8'] },
    { name: 'Intent clarification', desc: '由 text instruction 配置（"询问时间区间"），属意图级澄清', refs: ['S7'] },
  ],
  explain: [
    { name: 'SQL 展示', desc: 'Genie 把生成的 SQL 与结果一并展示（SQL 级解释）', refs: ['S7'] },
    { name: 'Verified answer 标记', desc: '命中 trusted asset / verified query 时标注答案来源与可信度', refs: ['S8'] },
  ],
  feedback: [
    { name: '点赞 / 点踩', desc: '用户对回答打分 · 标记请求审核', refs: ['S9'] },
    { name: 'Manager 审核', desc: 'space manager 确认 / 给出纠正 · 决定是否注入 verified', refs: ['S9'] },
    { name: 'SQL 片段建议', desc: '正向反馈促使 Genie 向 manager 建议新 measure / join / filter 供审批', refs: ['S7'] },
    { name: 'Verified injection', desc: '审批通过的 trusted asset 沉淀进 space · 后续问答匹配复用', refs: ['S8'] },
  ],
  undisclosed: [
    { name: '向量召回路径', desc: '公开文档未提对 schema / 指标做 embedding 检索（与 Snowflake Cortex Search 不同）' },
    { name: '自动 SQL 修复', desc: '官方文档未明确描述多轮错误修复闭环；第三方博客提到"静默重试"但未公开机制' },
    { name: '统计 / 代价感知生成', desc: '生成 SQL 不利用统计 / 基数 / 分区信息（除 cardinality / rely 显式建模提示）' },
  ],
  sampleConvo: {
    user: 'Q1 各大区营收 Top 3',
    interp: 'We interpreted your question as: 2025-Q1 (2025-01-01 至 2025-03-31)，按客户大区聚合订单总额，取前 3。',
    sql: `SELECT dim_customer.region AS 销售区域,
       SUM(fact_orders.amount)  AS total_revenue
FROM   main.sales.fact_orders
LEFT JOIN main.sales.dim_customer
       ON fact_orders.customer_id = dim_customer.customer_id
WHERE  fact_orders.order_date >= '2025-01-01'
  AND  fact_orders.order_date <  '2025-04-01'
  AND  fact_orders.is_canceled = false
GROUP BY 1
ORDER BY total_revenue DESC
LIMIT 3`,
    verified: true,
  },
  insights: {
    input: 'Genie 是 Databricks 用 NL → SQL 这条路对外的 Agent 入口；Space 是它的工作单元——选哪些 metric view、放哪些 example SQL、写哪些 text instruction，全靠人工策划（curation 即上下文）。',
    curate: [
      { icon: 'i-lucide-list-checks', title: 'Curation 即上下文', body: 'Space 不是自动学出来的：精选 metric view、写 example SQL、加 text instruction、跑 benchmark 评估——这套人工流程构成了 Genie 的"上下文层"，而不是某种 RAG 自动召回。' },
      { icon: 'i-lucide-search', title: '关键词召回，非向量召回', body: 'synonyms + example SQL 都是关键词 / 文本匹配；公开文档未见对 schema / 指标做 embedding 检索——和 Snowflake Cortex Search 的"高基数 dimension 接 search service"形成明显差异。' },
    ],
    runtime: [
      { icon: 'i-lucide-shield', title: '只读执行 · 用户身份', body: 'Genie 生成的 SQL 始终只读（read-only by design），并在当前用户的 session 与 RBAC / 行列策略下执行——越权问题返回空结果，与 Cortex Analyst 的 "API 不执行" 是不同选择。' },
      { icon: 'i-lucide-message-circle-question', title: '澄清靠 text instruction', body: '澄清能力不是平台内置交互，而是 manager 在 Space 里配置的指令（如"未指定时间范围时请询问"）；属意图级澄清，需人工设计。' },
    ],
    feedback: [
      { icon: 'i-lucide-recycle', title: '反馈即 verified', body: '点赞 / 点踩 + manager 审批后，正确的 SQL 片段 / verified query 沉淀进 trusted asset，被未来问答匹配复用——闭环靠人工审核驱动，非全自动学习。' },
    ],
  },
}

const ucFedArch: UcFedArch = {
  id: 'uc-federation',
  ucNamespace: {
    fqn: 'catalog.schema.{table | view | metric_view | foreign_table | system_table}',
    note: 'metric view 的 source 必须是 UC 表类资产；命名空间是统一的三段式，跨 catalog/schema 可直接 join',
    refs: ['S2', 'S5'],
  },
  sources: [
    { name: 'managed table', desc: 'Delta 托管表（标准）', refs: ['S5'] },
    { name: 'external table', desc: '外部存储 + Delta 元数据', refs: ['S5'] },
    { name: 'view', desc: '可作为 metric view 的 source', refs: ['S5'] },
    { name: 'foreign table', desc: 'Lakehouse Federation 接入的外部源', refs: ['S5', 'S11'] },
    { name: 'system_tables', desc: 'system.* 计量 / 审计 / lineage 表', refs: ['S5'] },
  ],
  connectors: [
    { group: 'OLTP', icon: 'i-lucide-database', items: 'PostgreSQL · MySQL · SQL Server · Oracle', refs: ['S11'] },
    { group: 'Cloud DW', icon: 'i-lucide-warehouse', items: 'Snowflake · BigQuery · Redshift · Synapse', refs: ['S11'] },
    { group: 'Lake / Query', icon: 'i-lucide-folder-tree', items: 'Hive Metastore · Glue · MongoDB · Salesforce DC', refs: ['S11'] },
  ],
  multiSource: {
    points: [
      'metric view 的 source 字段使用三段式 catalog.schema.table，可在 UC 统一命名空间下跨 catalog / schema 引用与 join',
      'Lakehouse Federation 把 PG / MySQL / Snowflake / BigQuery 等外部源以 foreign table 形式接入，metric view 可直接 join 跨外部源的表',
      '跨源 join 的执行由 Photon / SQL Warehouse 处理（部分谓词下推到外部源）',
    ],
    refs: ['S2', 'S5', 'S11'],
  },
  insights: {
    input: 'UC 不只是元数据目录，它是 Databricks 整个语义层的根命名空间——metric view 的合法 source 与 Genie 的可见数据范围都被它统一治理。',
    namespace: [
      { icon: 'i-lucide-folder-tree', title: '三段式 = 统一命名', body: 'catalog.schema.{table | view | metric_view | foreign_table} 一套命名贯穿所有对象类型；metric view 跟普通 table 一样可被引用、授权、追溯 lineage。' },
      { icon: 'i-lucide-shield-check', title: '权限 / 生命周期归 UC', body: '语义对象 = UC 对象，意味着 GRANT / REVOKE / DROP / lineage / audit 全自动落到平台层；不需要再做一套语义层 ACL。' },
    ],
    federate: [
      { icon: 'i-lucide-network', title: '跨外部源构建语义层', body: '同一个 metric view 的 join 可以横跨 Delta / PG / Snowflake / BigQuery；这把"统一语义层"从 lakehouse 内部扩展到了多 vendor 数据栈，是 metric view 多源能力的关键基座。' },
    ],
  },
}

const policyArch: PolicyArch = {
  id: 'policy-runtime',
  input: {
    label: 'Base Tables (UC managed/external)',
    note: 'metric view / view 本身不能直接挂行列策略；策略写在底表，查 metric view 时由引擎按用户身份强制传播',
    refs: ['S6'],
  },
  baseTablePolicies: [
    { name: 'ROW FILTER (function)', desc: 'CREATE FUNCTION → ALTER TABLE ... SET ROW FILTER · 按 session/role 过滤行', refs: ['S6'] },
    { name: 'COLUMN MASK (function)', desc: 'CREATE FUNCTION → ALTER TABLE ... ALTER COLUMN ... SET MASK · 按身份脱敏列', refs: ['S6'] },
    { name: 'GRANT / REVOKE', desc: 'metric view / table / catalog 三层 ACL · 角色级对象访问', refs: ['S2', 'S12'] },
  ],
  propagation: {
    points: [
      '行 / 列策略 NOT 可以直接 ALTER 在 view / metric view 上',
      '当查询 metric view 时，引擎在底表层强制套用这些策略',
      '不同用户查同一 metric view 看到不同结果（行被过滤、列被遮罩）',
      '越权问题不抛错，而是返回空结果或脱敏值',
    ],
    refs: ['S6'],
  },
  objectAcl: [
    { name: 'metric view GRANT', desc: 'metric view 自身可对象级授权（SELECT 给某 role）', refs: ['S2'] },
    { name: 'catalog / schema GRANT', desc: '继承式角色级访问', refs: ['S12'] },
  ],
  notPolicyAware: {
    points: [
      'NL2SQL 生成阶段不显式感知策略（不会避免生成越权 SQL）',
      '策略是查询执行时由引擎强制，不是 prompt 时刻',
      'Genie 不会解释"因为权限受限所以无结果"',
    ],
    refs: ['S6', 'S10'],
  },
  ddlExample: `-- 在底表 fact_orders 上设行过滤
CREATE FUNCTION main.sec.region_filter(region STRING)
RETURNS BOOLEAN
RETURN region = current_user_region()
   OR is_account_group_member('admins');

ALTER TABLE main.sales.fact_orders
SET ROW FILTER main.sec.region_filter
ON (customer_region);

-- 在底表 fact_orders 上设列遮罩
CREATE FUNCTION main.sec.mask_amount(a DECIMAL(18,2))
RETURNS DECIMAL(18,2)
RETURN CASE WHEN is_account_group_member('finance') THEN a
            ELSE NULL END;

ALTER TABLE main.sales.fact_orders
ALTER COLUMN amount SET MASK main.sec.mask_amount;

-- 之后任何走 metric view main.sales.revenue 的查询，
-- 都会自动套用 region_filter + mask_amount，
-- 不同用户身份返回不同结果。`,
  insights: {
    input: 'Row Filter / Column Mask 是 Databricks 的细粒度访问控制，但它们不能直接挂在 view 或 metric view 上——这是一个看似奇怪、实则关键的工程约束。',
    base: [
      { icon: 'i-lucide-shield', title: '策略只能写在底表', body: 'ALTER TABLE ... SET ROW FILTER / SET MASK 只对 base table 生效；这是 Databricks 设计选择——把"策略"作为底层物理对象的属性，而不是语义层的属性，避免每个 view 都得复制一遍策略。' },
      { icon: 'i-lucide-key-round', title: 'ACL 与策略分层', body: '对象访问（GRANT SELECT）走 metric view / table 层，行列策略走底表层；两套体系组合实现"谁能看 metric view + 看到时只看自己有权的行 / 列"。' },
    ],
    propagate: [
      { icon: 'i-lucide-share-2', title: '策略随查询自动传播', body: '查 metric view 时，底表的 row filter / column mask 在执行时按当前用户身份生效；同一条 SQL 不同人跑得到不同结果——治理交给查询引擎，不是 NL2SQL 平台。' },
      { icon: 'i-lucide-eye-off', title: '生成阶段不感知策略', body: 'Genie 在 prompt 阶段不会避免生成"理论上越权"的 SQL，也不解释权限限制；越权问题靠运行时静默过滤——选择把"策略"做为运行时不变量，而不是生成时约束。' },
    ],
  },
}

const mvMaterializeArch: MvMaterializeArch = {
  id: 'mv-materialize',
  input: {
    label: 'Metric View (logical)',
    note: 'metric view 默认是逻辑对象 · 查询时展开为聚合 SQL；可选物化为实体表，把聚合提前到刷新时刻',
    refs: ['S2', 'S14'],
  },
  setup: [
    { name: 'CREATE MATERIALIZED METRIC VIEW', desc: '将 metric view 物化 · 持久化聚合结果到 Delta 表', refs: ['S14'] },
    { name: 'SCHEDULE / TRIGGERED', desc: '增量刷新调度 · 按 cron 或显式触发', refs: ['S14'] },
    { name: 'cardinality / rely 提示', desc: '建模阶段写在 joins 上 · 优化器据此可消除冗余 join', refs: ['S4'] },
  ],
  rewrite: {
    points: [
      'MV 命中后查询直接读物化结果，跳过运行时聚合',
      '自动查询重写：用户写 metric view 的查询时，优化器透明替换为 MV',
      '增量刷新只重算受影响的分区（而非全量）',
      'cardinality + rely 提示让优化器消除"用户保证唯一"的冗余 join',
    ],
    refs: ['S2', 'S4', 'S14'],
  },
  refresh: [
    { name: 'incremental refresh', desc: '增量算法识别 source 变更，只重算受影响行', refs: ['S14'] },
    { name: 'manual REFRESH', desc: 'REFRESH MATERIALIZED VIEW · 强制重算', refs: ['S14'] },
    { name: 'staleness 提示', desc: '查询时可见 last refresh 时间 · 无强一致保证', refs: ['S14'] },
  ],
  cardinalityHints: [
    { name: 'at_most_one_match', desc: '维度对事实最多匹配一行（避免 fanout）', refs: ['S4'] },
    { name: 'one_to_one', desc: '一对一关系', refs: ['S4'] },
    { name: 'rely', desc: 'rely=true 表示保证唯一性 · 优化器可消除冗余 join', refs: ['S4'] },
  ],
  insights: {
    input: 'MV 与 cardinality 提示是 Databricks 在 metric view 之上做"性能感知"的两条主要路径——不是 NL2SQL 阶段感知统计，而是建模阶段写规则、执行阶段重写。',
    materialize: [
      { icon: 'i-lucide-snowflake', title: 'MV = 用空间换时间', body: '物化把聚合从查询时刻挪到刷新时刻；增量刷新让代价可控、命中后查询近乎 O(1)。是 metric view 应对大数据量场景的主要加速手段。' },
      { icon: 'i-lucide-refresh-ccw', title: '增量算法 · 弱一致', body: 'MV 不是强一致 · last refresh 时间可见 · 用户在精度与延迟之间显式权衡。这点和 Snowflake 的 dynamic table 思路一致。' },
    ],
    rewrite: [
      { icon: 'i-lucide-replace', title: '自动查询重写', body: '用户写的查询不需要点名 MV——命中条件时优化器透明替换。NL2SQL 路径完全不变，性能改造留在 SQL Warehouse / Photon 这一层。' },
      { icon: 'i-lucide-fingerprint', title: 'cardinality / rely = 规则型性能指导', body: '建模时写下"维度最多匹配一行"等业务事实，优化器据此消除冗余 join——把"性能"作为建模约束而非生成阶段感知，避开了 LLM 不懂统计的问题。' },
    ],
  },
}

export const DBX_MODULES: Record<string, DbxModuleData> = {
  'metric-view': { id: 'metric-view', accent: 'amber', metricView: metricViewArch },
  'agent-metadata': { id: 'agent-metadata', accent: 'amber', agentMetadata: agentMetadataArch },
  genie: { id: 'genie', accent: 'slate', genie: genieArch },
  'uc-federation': { id: 'uc-federation', accent: 'indigo', ucFed: ucFedArch },
  'policy-runtime': { id: 'policy-runtime', accent: 'emerald', policy: policyArch },
  'mv-materialize': { id: 'mv-materialize', accent: 'violet', mvMaterialize: mvMaterializeArch },
}

export function getDbxModule(id: string | null): DbxModuleData | null {
  if (!id) return null
  return DBX_MODULES[id] ?? null
}
