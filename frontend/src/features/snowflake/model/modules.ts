/**
 * Snowflake deck — L1 module data.
 *
 * Grounded in `WiseCat/.claude/skills/research/results/snowflake_cortex_analyst_semantic_views.yaml`
 * + cited official docs. Every box carries `refs: ['Sx']` chips.
 */
import type { AccentKey, NamedItem, Insight } from './architecture'

export interface RefItem extends NamedItem {
  refs?: string[]
}

/** Semantic View — DDL-native object + YAML-on-stage form. */
export interface SemanticViewArch {
  id: string
  forms: { name: string; desc: string; recommended?: boolean; refs?: string[] }[]
  ddlSections: { name: string; desc: string; refs?: string[] }[]
  joinSupport: RefItem[]
  metricFeatures: RefItem[]
  governance: RefItem[]
  ddlExample: string
  yamlExample: string
  insights: {
    input: string
    form: Insight[]
    model: Insight[]
  }
}

/** Cortex Analyst Flow — NL → SQL pipeline (official blog disclosed). */
export interface AnalystFlowArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  pipeline: { name: string; desc: string; refs?: string[] }[]
  classification: { points: string[]; refs: string[] }
  /** suggestions response type (clarification mechanism) */
  suggestions: { points: string[]; refs: string[] }
  errorLoop: { name: string; desc: string; refs?: string[] }[]
  /** REST API content types */
  responseTypes: { name: string; desc: string; refs?: string[] }[]
  apiExample: string
  insights: {
    input: string
    pipeline: Insight[]
    verify: Insight[]
    boundary: Insight[]
  }
}

/** Verified Query Repository — feedback / knowledge injection. */
export interface VqrArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  fields: { name: string; desc: string; refs?: string[] }[]
  addPaths: { name: string; desc: string; refs?: string[] }[]
  customInstructions: RefItem[]
  feedback: { name: string; desc: string; refs?: string[] }[]
  yamlExample: string
  insights: {
    input: string
    repo: Insight[]
    feedback: Insight[]
  }
}

/** Cortex Search — high-cardinality literal retrieval. */
export interface CortexSearchArch {
  id: string
  motivation: { points: string[]; refs: string[] }
  /** the hybrid retrieval composition */
  hybrid: { name: string; desc: string; refs?: string[] }[]
  /** how it integrates with semantic view */
  integration: { points: string[]; refs: string[] }
  /** what it serves at runtime */
  serves: RefItem[]
  yamlConfig: string
  insights: {
    input: string
    hybrid: Insight[]
    integrate: Insight[]
  }
}

/** Autopilot + Suggestions — assisted update / rich context generation. */
export interface AutopilotArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  autopilot: { name: string; desc: string; refs?: string[] }[]
  suggestions: { name: string; desc: string; refs?: string[] }[]
  generator: { name: string; desc: string; refs?: string[] }[]
  /** the human review gate */
  reviewGate: { points: string[]; refs: string[] }
  insights: {
    input: string
    generate: Insight[]
    review: Insight[]
  }
}

/** Policy runtime — masking / RAP propagation through semantic view. */
export interface PolicyArch {
  id: string
  input: { label: string; note: string; refs: string[] }
  baseTablePolicies: { name: string; desc: string; refs?: string[] }[]
  propagation: { points: string[]; refs: string[] }
  objectAcl: RefItem[]
  notPolicyAware: { points: string[]; refs?: string[] }
  insights: {
    input: string
    base: Insight[]
    propagate: Insight[]
  }
}

export interface SnowModuleData {
  id: string
  accent: AccentKey
  semanticView?: SemanticViewArch
  analystFlow?: AnalystFlowArch
  vqr?: VqrArch
  cortexSearch?: CortexSearchArch
  autopilot?: AutopilotArch
  policy?: PolicyArch
}

const semanticViewArch: SemanticViewArch = {
  id: 'semantic-view',
  forms: [
    { name: 'Native Semantic View (DDL)', desc: 'CREATE / ALTER / DROP / DESCRIBE · 由 GRANT 授权 · Cortex Analyst 推荐方式', recommended: true, refs: ['S2'] },
    { name: 'Stage YAML semantic_model', desc: 'YAML 文件存 stage 或字符串随请求传入 · 仍向后兼容', refs: ['S1', 'S17'] },
  ],
  ddlSections: [
    { name: 'TABLES', desc: 'logical table 列表，三段式 base_table 完全限定名 · 含 PRIMARY KEY / UNIQUE 约束', refs: ['S3', 'S4'] },
    { name: 'RELATIONSHIPS', desc: '实体间 join 语义 · 含 ASOF / range join 等高级形态', refs: ['S4'] },
    { name: 'FACTS', desc: '事实列（可被聚合的列）', refs: ['S3'] },
    { name: 'DIMENSIONS', desc: '维度列：name / synonyms / description / expr / data_type / is_enum / sample_values / labels', refs: ['S3'] },
    { name: 'METRICS', desc: '聚合公式（含窗口函数变体）· 一次定义、任意 dimension 重写', refs: ['S3', 'S4'] },
    { name: 'WITH SYNONYMS / COMMENT', desc: '元数据扩展子句 · 含 AI_SQL_GENERATION / AI_VERIFIED_QUERIES', refs: ['S4'] },
  ],
  joinSupport: [
    { name: 'inner / left / outer', desc: '标准 join · 在 RELATIONSHIPS 段声明', refs: ['S4'] },
    { name: 'ASOF', desc: '时间序列 ASOF join · 处理时间错位场景', refs: ['S4'] },
    { name: 'range join', desc: '区间 join · 适合层级 / 时间窗口', refs: ['S4'] },
  ],
  metricFeatures: [
    { name: 'metric formula', desc: '聚合公式集中定义 (SUM / COUNT(DISTINCT) / 自定义)', refs: ['S3'] },
    { name: 'window variants', desc: 'METRICS 支持窗口函数变体（如 moving avg）', refs: ['S4'] },
    { name: 'non_additive_dimensions', desc: '声明哪些 dimension 不能直接 SUM（避免错误聚合）', refs: ['S3'] },
    { name: 'view-level metrics', desc: 'view 级 metric · 复用查询', refs: ['S3'] },
  ],
  governance: [
    { name: 'schema-level object', desc: 'semantic view 是 schema 级原生对象，被归类为 metadata', refs: ['S2'] },
    { name: 'GRANT SELECT', desc: '查询 semantic view 需对象上 SELECT 权限（底表也需 SELECT）', refs: ['S15'] },
    { name: 'YAML on stage', desc: 'stage 上的 YAML 文件受 RBAC 控制', refs: ['S15'] },
  ],
  ddlExample: `CREATE OR REPLACE SEMANTIC VIEW SALES.PUBLIC.REVENUE
TABLES (
  fact_orders   AS RAW.SALES.FACT_ORDERS  PRIMARY KEY (order_id),
  dim_customer  AS RAW.SALES.DIM_CUSTOMER PRIMARY KEY (customer_id)
)
RELATIONSHIPS (
  fact_orders.customer_id ->-< dim_customer.customer_id
)
FACTS (
  amount AS fact_orders.amount
)
DIMENSIONS (
  region AS dim_customer.region
    WITH SYNONYMS = ('area', 'territory', '大区')
    COMMENT = '客户大区',
  order_month AS DATE_TRUNC('month', fact_orders.order_date)
)
METRICS (
  total_revenue AS SUM(amount)
)
COMMENT = 'Revenue by region · Q1 2025+';`,
  yamlExample: `# stage YAML 形态（兼容路径）
name: revenue
tables:
  - name: fact_orders
    base_table:
      database: RAW
      schema: SALES
      table: FACT_ORDERS
    primary_key: { columns: [order_id] }
    facts:
      - name: amount
        expr: amount
    dimensions:
      - name: region
        expr: dim_customer.region
        synonyms: [area, territory]
        sample_values: [华东, 华北, 华南, 西部]
    metrics:
      - name: total_revenue
        expr: SUM(amount)
verified_queries:
  - name: q1_revenue_by_region
    question: Q1 各大区营收
    sql: |
      SELECT region, SUM(amount)
      FROM RAW.SALES.FACT_ORDERS f
      JOIN RAW.SALES.DIM_CUSTOMER c USING(customer_id)
      WHERE order_date >= '2025-01-01' AND order_date < '2025-04-01'
      GROUP BY 1
    verified_by: alice@acme
    verified_at: 2025-04-08`,
  insights: {
    input: 'Snowflake 的 semantic layer 提供两种形态：原生 Semantic View（DDL 对象，官方推荐）与 stage 上的 YAML semantic_model（向后兼容）；Cortex Analyst 同时支持两者。',
    form: [
      { icon: 'i-lucide-database', title: 'Semantic View = schema 级原生对象', body: 'CREATE / ALTER / DROP / GRANT 全套 DDL · 被归类为 metadata · 不是配置文件而是数据库对象——这与 Databricks Metric View 同源思路：把语义对象做成数据库一等公民。' },
      { icon: 'i-lucide-file-text', title: 'YAML 形态向后兼容', body: '老路径仍可用：YAML 存 stage 或字符串随请求传入。Cortex Analyst 可以同时支持两种来源，但官方推荐迁移到原生 semantic view。' },
    ],
    model: [
      { icon: 'i-lucide-shapes', title: '完整的关系 + 指标模型', body: 'TABLES（带 PRIMARY KEY / UNIQUE）+ RELATIONSHIPS（含 ASOF / range join）+ FACTS / DIMENSIONS / METRICS（含窗口变体 + non_additive_dimensions）—— Schema IR 表达力达到关系-指标层级。' },
      { icon: 'i-lucide-tags', title: '富上下文内联在 dimension', body: 'synonyms / description / sample_values / is_enum / labels 直接挂在 dimension 上；DDL 还支持 WITH SYNONYMS / AI_SQL_GENERATION / AI_VERIFIED_QUERIES 等 AI 扩展子句。' },
    ],
  },
}

const analystFlowArch: AnalystFlowArch = {
  id: 'analyst-flow',
  input: {
    label: 'NL question + semantic_model_file',
    note: 'POST /api/v2/cortex/analyst/message · 请求带 messages + semantic_model_file/view',
    refs: ['S1', 'S8'],
  },
  pipeline: [
    { name: '1. Read semantic model', desc: '读取 semantic view 定义或 stage 上 YAML（支持多个 model 路由）', refs: ['S1'] },
    { name: '2. Classify question', desc: 'classification agent 把歧义问题前置拒答 · 生成相似问题列表', refs: ['S9'] },
    { name: '3. Generate on logical schema', desc: '内部先在简化 logical schema 上生成 SQL（不直接生成物理 SQL）', refs: ['S9'] },
    { name: '4. Post-process to physical', desc: '后处理把 logical SQL 转成物理表 SQL（绑定 base_table）', refs: ['S9'] },
    { name: '5. Compiler check + repair', desc: 'error correction agent 用 SQL compiler 检查语法/语义 · 触发 error correction loop', refs: ['S9'] },
    { name: '6. Return SQL (no execute)', desc: '返回可直接执行的 SQL · API 不执行，调用方在自身 warehouse 执行', refs: ['S1', 'S8'] },
  ],
  classification: {
    points: [
      'classification agent 在生成 SQL 前先判分类是否歧义',
      '歧义问题 → suggestions 响应（不返回 SQL）',
      '正常问题 → 走生成路径',
    ],
    refs: ['S9'],
  },
  suggestions: {
    points: [
      'response.content.type = "suggestions" · 仅在歧义、无法返回 SQL 时返回',
      '内容是语义模型可回答的相似问题列表',
      '澄清形式 = "拒答 + 建议问题列表"，而非交互式定向提问',
      'content.type = "text"（解释）也会返回："We interpreted your question as ..."',
    ],
    refs: ['S8', 'S9', 'S1'],
  },
  errorLoop: [
    { name: 'SQL compiler check', desc: 'Snowflake SQL compiler 静态检查语法/语义/字段存在性', refs: ['S9'] },
    { name: 'hallucination handler', desc: '检测引用了不存在的列/表 · 触发 LLM 修复', refs: ['S9'] },
    { name: 'iterative repair', desc: 'error correction loop 多轮生成-检查-修复直到通过', refs: ['S9'] },
    { name: 'warnings array', desc: 'response.warnings[] 暴露生成过程问题给调用方', refs: ['S8'] },
  ],
  responseTypes: [
    { name: 'sql', desc: '生成的 SQL 语句', refs: ['S8'] },
    { name: 'text', desc: '解释文本（"We interpreted your question as ..."）', refs: ['S1', 'S8'] },
    { name: 'suggestions', desc: '歧义时返回的相似问题列表', refs: ['S8'] },
    { name: 'confidence.verified_query_used', desc: '命中 VQR 时披露 name/question/sql/verified_at/verified_by', refs: ['S5', 'S8'] },
  ],
  apiExample: `# 请求
POST /api/v2/cortex/analyst/message
{
  "messages": [{"role": "user", "content": [{"type": "text", "text": "Q1 各大区营收 Top 3"}]}],
  "semantic_view": "SALES.PUBLIC.REVENUE"
}

# 响应（成功路径）
{
  "request_id": "abc-123",
  "message": {
    "role": "analyst",
    "content": [
      {"type": "text", "text": "We interpreted your question as: 2025-Q1 (..) by region top 3."},
      {"type": "sql", "statement": "SELECT region, SUM(amount) ... LIMIT 3"}
    ]
  },
  "warnings": [],
  "confidence": {
    "verified_query_used": {
      "name": "q1_revenue_by_region",
      "question": "Q1 各大区营收",
      "verified_at": "2025-04-08",
      "verified_by": "alice@acme"
    }
  }
}`,
  insights: {
    input: 'Cortex Analyst 把 NL2SQL 拆成"读语义模型 → 分类 → 生成 → 后处理 → 校验 → 返回"六步，且 API 本身不执行——执行权完全交给调用方在自身 warehouse 完成。',
    pipeline: [
      { icon: 'i-lucide-layers', title: 'Logical → Physical 两段生成', body: '官方博客披露：先在简化 logical schema 上生成 SQL，再后处理为物理 SQL。中间产物仍是 SQL（不是结构化 query IR），因此评估为"端到端 SQL 生成"。' },
      { icon: 'i-lucide-message-circle-question', title: 'Classification 前置拒答', body: 'classification agent 把歧义问题在生成前拦截，返回 suggestions（相似问题列表）；这是 Snowflake 的澄清机制——非交互式定向提问，而是"拒答 + 建议"。' },
    ],
    verify: [
      { icon: 'i-lucide-bug-off', title: 'SQL compiler-based 验证', body: '验证基于 SQL compiler 静态检查（语法、语义、字段存在性），不是真实执行；不返回前不试跑、也不做空结果检测。' },
      { icon: 'i-lucide-recycle', title: 'Iterative repair loop', body: 'error correction agent 多轮生成-检查-修复直到通过 · hallucination 处理是关键能力（这与 Databricks Genie 未公开多轮修复形成差异）。' },
    ],
    boundary: [
      { icon: 'i-lucide-shield', title: 'API 不执行，仅返回 SQL', body: 'Cortex Analyst REST API 仅生成 SQL · 调用方在自身 warehouse 用自身角色执行——这把"安全 / 权限 / 计费"完全交给 Snowflake 现有 RBAC，与 Databricks Genie 内置执行不同。' },
      { icon: 'i-lucide-lock', title: 'Governance boundary 不外发', body: '官方承诺不在客户数据上训练 · 默认用 Snowflake 托管 LLM · 数据/元数据/提示不离开 Snowflake governance boundary——属 runtime/平台级安全边界。' },
    ],
  },
}

const vqrArch: VqrArch = {
  id: 'vqr',
  input: {
    label: 'Verified Query Repository (semantic model 内)',
    note: 'verified_queries 段落 · 每条含 name / question / sql / verified_by / verified_at',
    refs: ['S5'],
  },
  fields: [
    { name: 'name', desc: '查询标识（响应 confidence.verified_query_used.name）', refs: ['S5', 'S8'] },
    { name: 'question', desc: '自然语言问题（用作召回输入）', refs: ['S5'] },
    { name: 'sql', desc: '人工验证过的 SQL（命中即直接复用）', refs: ['S5'] },
    { name: 'verified_by', desc: '审核者标识', refs: ['S5'] },
    { name: 'verified_at', desc: '审核时间戳', refs: ['S5'] },
  ],
  addPaths: [
    { name: '人工编辑 YAML', desc: '直接在 semantic model YAML 的 verified_queries 段编辑提交', refs: ['S5'] },
    { name: 'Streamlit 工具交互', desc: '用户在 Streamlit UI 验证 SQL 后保存为 verified query', refs: ['S5'] },
    { name: 'Verified Query Suggestion', desc: 'Snowsight 基于用户行为建议候选 verified query · 人工接受/编辑/驳回', refs: ['S12'] },
  ],
  customInstructions: [
    { name: 'natural language rules', desc: '自然语言规则约束 SQL 生成形状', refs: ['S14'] },
    { name: 'default time filter', desc: '官方示例：默认时间过滤器 · 影响生成 SQL 形状', refs: ['S14'] },
    { name: 'filter propagation', desc: '关联列过滤传播规则', refs: ['S14'] },
  ],
  feedback: [
    { name: 'POST /feedback', desc: 'feedback endpoint · request_id + positive (赞/踩) + 可选 message', refs: ['S8'] },
    { name: 'verified_query_used', desc: '响应 confidence 字段披露命中的 VQR · 标 verified answer', refs: ['S8'] },
    { name: 'monitoring tab', desc: '管理员可在 Snowsight Monitoring 查看 SQL / errors / warnings 日志', refs: ['S16'] },
    { name: 'Optimization', desc: '从已有 verified queries 提炼可泛化概念 · 建议新增（人工确认）', refs: ['S5'] },
  ],
  yamlExample: `# semantic model YAML 内的 verified_queries 段
verified_queries:
  - name: q1_revenue_by_region
    question: Q1 各大区营收
    sql: |
      SELECT region, SUM(amount)
      FROM RAW.SALES.FACT_ORDERS f
      JOIN RAW.SALES.DIM_CUSTOMER c USING(customer_id)
      WHERE order_date >= '2025-01-01' AND order_date < '2025-04-01'
      GROUP BY 1
    verified_by: alice@acme
    verified_at: 2025-04-08

  - name: top_customers_lifetime
    question: 终身价值 Top 10 客户
    sql: |
      SELECT customer_id, SUM(amount) AS ltv
      FROM RAW.SALES.FACT_ORDERS GROUP BY 1
      ORDER BY ltv DESC LIMIT 10
    verified_by: bob@acme
    verified_at: 2025-04-15

custom_instructions: |
  - 默认时间范围：若用户未指定时间，使用 last_30_days
  - 关联列过滤传播：当用户按 region 过滤时，对 dim_customer 同时应用过滤`,
  insights: {
    input: 'Verified Query Repository 是 Snowflake 反馈闭环的核心载体——它把"审过的 NL-SQL 对"作为富上下文，按问题相关性检索注入 prompt，命中即可直接返回 verified answer。',
    repo: [
      { icon: 'i-lucide-shield-check', title: 'VQR = 经验注入', body: 'verified queries 是经过人工/管理员审核的 NL-SQL 对 · 三种添加路径（手写 YAML / Streamlit 验证 / Snowsight 建议）—— 把"团队最佳 SQL"沉淀成可复用上下文。' },
      { icon: 'i-lucide-pencil', title: 'Custom Instructions = 规则上下文', body: '与 VQR 并列：自然语言规则（默认时间过滤器、关联列过滤传播）也作为富上下文影响生成 SQL 形状。' },
    ],
    feedback: [
      { icon: 'i-lucide-recycle', title: '反馈回流为人工闭环', body: '点赞/点踩 + 监控日志如何自动影响后续生成未披露 · verified query 需人工确认后注入 · 未见全自动学习证据——属"经验注入"而非"闭环学习"。' },
      { icon: 'i-lucide-eye', title: '响应可解释 verified 来源', body: '命中 VQR 时响应 confidence.verified_query_used 直接返回 name/question/verified_at/verified_by · 用户可看到答案的可信来源。' },
    ],
  },
}

const cortexSearchArch: CortexSearchArch = {
  id: 'cortex-search',
  motivation: {
    points: [
      '高基数 dimension（>10 个不同值）若全量塞 sample_values 进语义模型 → 上下文爆炸',
      '官方建议：把这类 dimension 的 literal 召回外置到独立服务',
      'Cortex Search 让 schema 模型保持精简，又能在生成时用真实数据值',
      '官方：reduces data duplication and keeps your semantic model concise',
    ],
    refs: ['S6'],
  },
  hybrid: [
    { name: 'Vector retrieval', desc: 'embedding 相似度检索（语义近似）', refs: ['S7'] },
    { name: 'Keyword search', desc: '倒排索引精确词命中（容错短查询）', refs: ['S7'] },
    { name: 'Semantic rerank', desc: '检索后语义重排 · 提升相关性', refs: ['S7'] },
  ],
  integration: {
    points: [
      'dimension 上配置 cortex_search_service: <service_name>',
      '生成 SQL 时遇到该 dimension 的字面值 → 转向 search service 召回',
      '召回结果作为 literal 候选注入 prompt',
      '语义模型本体仍整体注入（无 schema 裁剪），仅富上下文（VQR / literal）按需召回',
    ],
    refs: ['S1', 'S6', 'S9'],
  },
  serves: [
    { name: 'literal lookup', desc: '高基数列字面值召回（如客户名 / SKU）', refs: ['S6'] },
    { name: 'verified_queries 召回', desc: 'VQR 也按问题相关性检索注入', refs: ['S9'] },
    { name: 'dynamic instance awareness', desc: '运行时对底层列实际值做语义检索 · 感知真实数据', refs: ['S6'] },
  ],
  yamlConfig: `# semantic model YAML：高基数 dimension 接 Cortex Search
dimensions:
  - name: customer_name
    expr: dim_customer.full_name
    description: 客户全名（来自 CRM）
    cortex_search_service: ANALYTICS.SEARCH.CUSTOMER_NAMES   # ← 关键
    # 不再写 sample_values: 全量列表（避免上下文爆炸）

  - name: product_sku
    expr: dim_product.sku
    cortex_search_service: ANALYTICS.SEARCH.PRODUCT_SKUS

# 单独建 search service（一次性）
CREATE OR REPLACE CORTEX SEARCH SERVICE ANALYTICS.SEARCH.CUSTOMER_NAMES
  ON full_name
  WAREHOUSE = SEARCH_WH
  TARGET_LAG = '1 hour'
  AS SELECT full_name FROM RAW.SALES.DIM_CUSTOMER;`,
  insights: {
    input: 'Snowflake 用 Cortex Search 解决了一个 Databricks 没明确解决的问题：当 dimension 是高基数（客户名 / SKU 等），怎么让 NL2SQL 知道真实数据值，又不把上下文塞爆。',
    hybrid: [
      { icon: 'i-lucide-blocks', title: '混合检索（向量 + 关键词 + rerank）', body: '不是纯向量也不是纯关键词：Cortex Search 把三者组合 · 向量抓语义、关键词抓精确词、rerank 抓相关性——这是 Snowflake 检索能力上比 Databricks Genie（关键词召回）更高一档的关键。' },
    ],
    integrate: [
      { icon: 'i-lucide-share-2', title: 'literal 召回外置', body: 'cortex_search_service 字段把 dimension 的字面值召回从语义模型抽出去 · 模型本体只放 schema/语义，运行时实际值由 search service 召回——是 token saving 的核心机制。' },
      { icon: 'i-lucide-radar', title: 'Dynamic instance awareness', body: '生成时实际去查底层列 · 知道真实数据分布；属"动态实例感知"——这一点比 Databricks（仅 cardinality / rely 显式提示）走得更远。' },
    ],
  },
}

const autopilotArch: AutopilotArch = {
  id: 'autopilot',
  input: {
    label: 'Query History · Table metadata · Sample queries',
    note: '辅助生成的输入：历史 SQL · 表元数据 · 示例查询',
    refs: ['S11', 'S12'],
  },
  autopilot: [
    { name: 'analyze query history', desc: '从 Query History 学习常用 join / filter / 字段', refs: ['S11'] },
    { name: 'AI generate semantic view draft', desc: '生成候选 semantic view（含描述 / 关系 / verified queries）', refs: ['S11'] },
    { name: '产物为草稿', desc: '人工审核后采用 · 不自动应用到生产', refs: ['S11'] },
  ],
  suggestions: [
    { name: 'Snowsight Suggestions', desc: '在已有 semantic model/view 上提候选修改', refs: ['S12'] },
    { name: 'metric / filter / description / synonym 候选', desc: '基于查询历史与用户提问产生', refs: ['S12'] },
    { name: 'verified query suggestion', desc: '从用户行为中识别可固化的 NL-SQL 对', refs: ['S12'] },
    { name: 'human review', desc: '所有建议需人工接受 / 编辑 / 驳回 · 才入模', refs: ['S12'] },
  ],
  generator: [
    { name: 'semantic-model-generator (Labs)', desc: '从 schema 一次性生成 YAML 草稿', refs: ['S13'] },
    { name: 'LLM 描述补全', desc: '缺注释时调用 Cortex LLM 生成 description（带后缀标记待复核）', refs: ['S13'] },
    { name: 'FILL-OUT 段', desc: '保留人工补充同义词 / 过滤器的占位段', refs: ['S13'] },
  ],
  reviewGate: {
    points: [
      '所有 AI 生成 / 候选建议都不自动应用',
      '产出的是"草稿"或"建议"，进入人工审核队列',
      '人工接受 → 入模；编辑 / 驳回 → 不入模',
      '这把"自适应更新"压回到"辅助更新"等级',
    ],
    refs: ['S11', 'S12'],
  },
  insights: {
    input: 'Snowflake 在"维护语义层 / 富上下文"上有较完整的 AI 辅助路径——但所有路径都有显式人工审核闸门，因此 Schema IR 更新仍归类为"辅助更新"，富上下文为"辅助维护"。',
    generate: [
      { icon: 'i-lucide-sparkles', title: '三条 AI 辅助路径', body: 'Autopilot（生成 semantic view 草稿）+ Snowsight Suggestions（已有模型上提候选）+ semantic-model-generator（Labs 一次性生成 YAML）—— 三套机制覆盖从 0 到 1 与从 1 到 N。' },
      { icon: 'i-lucide-history', title: 'Query History 是关键输入', body: 'Autopilot 与 Suggestions 都从 Query History 学习——这把"团队历史 SQL"作为 AI 生成 semantic 模型的语料，弥补了人工注解的稀疏。' },
    ],
    review: [
      { icon: 'i-lucide-user-check', title: 'Human review 闸门', body: '所有 AI 产物都是草稿/建议 · 人工接受/编辑/驳回后才入模——把"自适应更新"压回"辅助更新"。这是工程取舍：保证语义层稳定性 > 自动化覆盖。' },
      { icon: 'i-lucide-eye-off', title: '未见物理 schema 漂移检测', body: '官方未披露针对底层 schema 漂移的自动检测告警或自动同步机制 · semantic view 创建后不能直接增改表/列 · schema 变化需 CREATE OR REPLACE 重建。' },
    ],
  },
}

const policyArch: PolicyArch = {
  id: 'policy-runtime',
  input: {
    label: 'Base Tables (RAW · physical)',
    note: 'masking / row access policy 不能直接挂在 semantic view 上 · 必须设在底表',
    refs: ['S10'],
  },
  baseTablePolicies: [
    { name: 'MASKING POLICY', desc: '列级脱敏 · ALTER TABLE ... ALTER COLUMN ... SET MASKING POLICY', refs: ['S10'] },
    { name: 'ROW ACCESS POLICY', desc: '行级过滤 · ALTER TABLE ... ADD ROW ACCESS POLICY', refs: ['S10'] },
    { name: 'GRANT', desc: 'semantic view + 底表 SELECT 权限 · 角色级访问', refs: ['S15'] },
  ],
  propagation: {
    points: [
      '底表的 masking / row access policy 在执行时按当前查询用户身份生效',
      '即使查询走 semantic view，策略仍会传播到底表层',
      '不同用户查同一 semantic view 看到不同结果',
      '样例值（sample_values）属元数据，不受 masking 保护——避免暴露敏感值',
    ],
    refs: ['S10'],
  },
  objectAcl: [
    { name: 'semantic view GRANT', desc: 'semantic view 自身可对象级授权', refs: ['S15'] },
    { name: 'YAML stage RBAC', desc: 'stage 上的 YAML 文件受 RBAC 控制', refs: ['S15'] },
  ],
  notPolicyAware: {
    points: [
      'NL2SQL 生成阶段不显式感知策略（不会避免生成越权 SQL）',
      '策略不能直接定义在 semantic view 属性上',
      '所有越权问题靠运行时强制（用户角色 + 底表策略）',
    ],
    refs: ['S10'],
  },
  insights: {
    input: 'Snowflake 与 Databricks 在 policy 上的取舍极其相似：策略写在底表 · 查 view / semantic view 时按用户身份在执行时强制传播 · 生成阶段不感知策略。',
    base: [
      { icon: 'i-lucide-shield', title: '策略只能写在底表', body: 'masking / row access policy 不能直接挂在 view / semantic view 上；这是为了避免每个 view 都得复制策略。这与 Databricks Row Filter / Column Mask 是同一设计哲学。' },
      { icon: 'i-lucide-eye-off', title: 'sample_values 不受保护', body: '语义模型里的 sample_values 是元数据 · 不被 masking 覆盖 · 因此官方提醒不要把敏感值放进 sample_values（应改用 cortex_search_service 召回真实值）。' },
    ],
    propagate: [
      { icon: 'i-lucide-share-2', title: '运行时按身份强制', body: 'Cortex Analyst 不执行 SQL · 调用方在自身 warehouse 用自身角色执行 · 自然继承 RBAC + masking + row access 三层约束——策略与 NL2SQL 完全解耦。' },
      { icon: 'i-lucide-x', title: '生成阶段不感知策略', body: 'NL2SQL 生成不会因权限受限而调整 SQL 或解释 · 所有越权问题靠执行时静默处理（返回空结果或脱敏值）——把"安全"作为运行时不变量。' },
    ],
  },
}

export const SNOW_MODULES: Record<string, SnowModuleData> = {
  'semantic-view': { id: 'semantic-view', accent: 'amber', semanticView: semanticViewArch },
  'analyst-flow': { id: 'analyst-flow', accent: 'slate', analystFlow: analystFlowArch },
  vqr: { id: 'vqr', accent: 'violet', vqr: vqrArch },
  'cortex-search': { id: 'cortex-search', accent: 'blue', cortexSearch: cortexSearchArch },
  autopilot: { id: 'autopilot', accent: 'violet', autopilot: autopilotArch },
  'policy-runtime': { id: 'policy-runtime', accent: 'emerald', policy: policyArch },
}

export function getSnowModule(id: string | null): SnowModuleData | null {
  if (!id) return null
  return SNOW_MODULES[id] ?? null
}
