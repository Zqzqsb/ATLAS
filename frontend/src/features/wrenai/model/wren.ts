/**
 * WrenAI architecture model — single source of truth for the `/wrenai` panorama.
 *
 * WrenAI (Canner) post-2026-05 = an *open context layer for agents*. Unlike ATLAS
 * (agentic: Coordinator/Worker ReAct agents own RC generation + NL2SQL end-to-end),
 * WrenAI keeps the agent EXTERNAL ("BYO agent") and ships a governed semantic layer
 * (MDL) plus correctness PRIMITIVES (memory / dry-plan / dry-run / structured errors)
 * the agent orchestrates. Grounded in the local WrenAI checkout + docs.getwren.ai.
 *
 * Reuses the generic diagram types/accents/primitives from features/arch so both
 * decks share one visual language.
 */
import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }

/* ─── L0 panorama: the open-context-layer stack (also a query path) ─── */
export const WREN_LAYERS: ArchLayer[] = [
  {
    id: 'agent',
    title: 'Agent Workflow',
    subtitle: 'BYO Agent · Skills 编排（Agent 外置，不内置 NL2SQL LLM 服务）',
    icon: 'i-lucide-bot',
    accent: 'slate',
    cols: 3,
    nodes: [
      {
        id: 'agent-fw',
        label: 'Coding / Framework Agent',
        sublabel: 'LangChain · LangGraph · Pydantic AI · 你已有的 Agent',
        icon: 'i-lucide-bot',
        accent: 'slate',
        span: 1,
      },
      {
        id: 'skills',
        label: 'Agent Skills',
        sublabel: 'Markdown 工作流：先建模 → 取上下文 → 写 SQL → 验证 → 记忆',
        icon: 'i-lucide-scroll-text',
        accent: 'slate',
        span: 1,
        codeRefs: ['skills/wren/SKILL.md', 'core/wren/src/wren/skills_content/'],
      },
      {
        id: 'sdk',
        label: 'Agent SDK',
        sublabel: 'wren-langchain · wren-pydantic（把原语包装成工具）',
        icon: 'i-lucide-plug',
        accent: 'slate',
        span: 1,
        codeRefs: ['sdk/wren-langchain/', 'sdk/wren-pydantic/'],
      },
    ],
  },
  {
    id: 'cli',
    title: 'Wren CLI / SDK',
    subtitle: '编排原语 · plan-and-execute（CLI 是 Python SDK 的薄 Typer 包装）',
    icon: 'i-lucide-terminal',
    accent: 'violet',
    cols: 2,
    nodes: [
      {
        id: 'cli-node',
        label: 'Wren CLI',
        sublabel: 'query · dry-plan · dry-run · context · profile · memory',
        icon: 'i-lucide-terminal',
        accent: 'violet',
        span: 1,
        codeRefs: ['core/wren/src/wren/cli.py'],
      },
      {
        id: 'query-flow',
        label: 'Agent 辅助查询 · NL2SQL',
        sublabel: 'recall → fetch → 写 SQL → dry-plan → 执行 → 修复 → 记忆',
        icon: 'i-lucide-workflow',
        accent: 'violet',
        flow: 'query',
        span: 1,
        codeRefs: ['core/wren/src/wren/engine.py'],
      },
    ],
  },
  {
    id: 'context',
    title: 'Project Context',
    subtitle: '可移植的开放上下文包（Git 化语义层，凭据分离）',
    icon: 'i-lucide-folder-git-2',
    accent: 'emerald',
    cols: 4,
    nodes: [
      {
        id: 'mdl',
        label: 'MDL',
        sublabel: '语义契约：models / 关系 / 计算列 / views / cubes · RLAC/CLAC',
        icon: 'i-lucide-box',
        accent: 'emerald',
        flow: 'mdl',
        span: 1,
        codeRefs: ['core/wren-mdl/mdl.schema.json', 'docs/core/reference/mdl.md'],
      },
      {
        id: 'instructions',
        label: 'instructions.md',
        sublabel: '业务 / 操作指南（注入 system prompt）',
        icon: 'i-lucide-file-text',
        accent: 'emerald',
        span: 1,
        codeRefs: ['core/wren/src/wren/context.py'],
      },
      {
        id: 'queries',
        label: 'queries.yml',
        sublabel: '审核过的 NL-SQL 样例（可 seed memory）',
        icon: 'i-lucide-list-checks',
        accent: 'emerald',
        span: 1,
      },
      {
        id: 'profiles',
        label: 'profiles.yml',
        sublabel: '连接 profile（凭据与项目分离，存 ~/.wren）',
        icon: 'i-lucide-key-round',
        accent: 'emerald',
        span: 1,
        codeRefs: ['core/wren/src/wren/profile.py'],
      },
    ],
  },
  {
    id: 'memory',
    title: 'Memory',
    subtitle: 'LanceDB 向量检索层（schema linking + few-shot 召回）',
    icon: 'i-lucide-database',
    accent: 'blue',
    cols: 3,
    nodes: [
      {
        id: 'schema-items',
        label: 'schema_items',
        sublabel: 'models / 列 / 关系 / views / cubes（+ instructions）',
        icon: 'i-lucide-table-2',
        accent: 'blue',
        span: 1,
        codeRefs: ['core/wren/src/wren/memory/schema_indexer.py'],
      },
      {
        id: 'query-history',
        label: 'query_history',
        sublabel: '确认过的 NL-SQL 对（recall 复用）',
        icon: 'i-lucide-history',
        accent: 'blue',
        span: 1,
        codeRefs: ['core/wren/src/wren/memory/store.py'],
      },
      {
        id: 'embed',
        label: 'Embeddings',
        sublabel: 'sentence-transformers · 多语 MiniLM · 384d',
        icon: 'i-lucide-spline',
        accent: 'blue',
        span: 1,
        codeRefs: ['core/wren/src/wren/memory/embeddings.py'],
      },
    ],
  },
  {
    id: 'planning',
    title: 'Planning Engine',
    subtitle: 'Modeled SQL → 可执行 SQL（语义层正确性的真相源）',
    icon: 'i-lucide-cpu',
    accent: 'amber',
    cols: 4,
    nodes: [
      {
        id: 'sqlglot',
        label: 'sqlglot',
        sublabel: 'parse · qualify · 方言转译',
        icon: 'i-lucide-code-2',
        accent: 'amber',
        span: 1,
      },
      {
        id: 'cte',
        label: 'CTE Rewriter',
        sublabel: '识别引用模型 → 注入展开后的 CTE',
        icon: 'i-lucide-git-merge',
        accent: 'amber',
        span: 1,
        codeRefs: ['core/wren/src/wren/mdl/cte_rewriter.py'],
      },
      {
        id: 'wren-core',
        label: 'wren-core',
        sublabel: 'Rust · Apache DataFusion · MDL 语义展开',
        icon: 'i-lucide-cog',
        accent: 'amber',
        span: 1,
        codeRefs: ['core/wren-core/core/src/mdl/mod.rs'],
      },
      {
        id: 'policy',
        label: 'Policy Checks',
        sublabel: 'strict mode · RLAC/CLAC · denied funcs · row limit',
        icon: 'i-lucide-shield-check',
        accent: 'amber',
        span: 1,
        codeRefs: ['core/wren/src/wren/policy.py'],
      },
    ],
  },
  {
    id: 'execution',
    title: 'Execution Layer',
    subtitle: '原生连接器 · 20+ 数据源（dry-run 校验，结果回 PyArrow）',
    icon: 'i-lucide-plug-zap',
    accent: 'indigo',
    cols: 3,
    nodes: [
      {
        id: 'connectors',
        label: 'Connectors',
        sublabel: 'Postgres / MySQL / BigQuery / Snowflake / DuckDB / Trino …',
        icon: 'i-lucide-plug',
        accent: 'indigo',
        span: 1,
        codeRefs: ['core/wren/src/wren/connector/factory.py'],
      },
      {
        id: 'dryrun',
        label: 'Dry-run',
        sublabel: 'LIMIT 0 对 live DB 校验，不返回行',
        icon: 'i-lucide-flask-conical',
        accent: 'indigo',
        span: 1,
      },
      {
        id: 'pyarrow',
        label: 'PyArrow Result',
        sublabel: 'table · CSV · JSON · SDK 返回值',
        icon: 'i-lucide-table',
        accent: 'indigo',
        span: 1,
      },
    ],
  },
]

/* ─── L1 module identity ─── */
export interface WrenFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const wrenFlows: WrenFlowDef[] = [
  {
    id: 'mdl',
    label: 'MDL',
    title: 'MDL · 语义契约',
    subtitle: '把物理 schema 建模成机器可读的业务语义层：models / 关系 / 计算列 / views / cubes / 访问控制，编译为 target/mdl.json 供引擎规划',
    icon: 'i-lucide-box',
    accent: 'emerald',
  },
  {
    id: 'query',
    label: 'NL2SQL Query Flow',
    title: 'Agent 辅助查询 · NL2SQL',
    subtitle: '正确性不是一个隐藏特性，而是一组 Agent 自行编排的原语：召回 / 取上下文 / 针对 MDL 写 SQL / dry-plan / 执行 / 修复 / 记忆',
    icon: 'i-lucide-workflow',
    accent: 'violet',
  },
]

export function getWrenFlow(id: string | null): WrenFlowDef | null {
  if (!id) return null
  return wrenFlows.find((f) => f.id === id) ?? null
}

/* ─── L1 module internal data ─── */

/** MDL — the semantic contract. */
export interface MdlArch {
  id: string
  input: { label: string; note: string }
  /** the modeled entities (each a box with a peek list) */
  entities: { title: string; desc: string; icon: string; accent: AccentKey; items: NamedItem[] }[]
  /** governance baked into the model */
  govern: { title: string; items: NamedItem[] }
  /** compile step */
  compile: { cmd: string; out: string; note: string }
  /** 5-layer context model annotation */
  contextLayers: NamedItem[]
  /** illustrative project YAML + compiled JSON */
  yamlExample: string
  jsonExample: string
  insights: {
    input: string
    model: Insight[]
    compile: Insight[]
  }
}

/** NL2SQL — correctness as a system of primitives. */
export interface QueryArch {
  id: string
  input: { label: string; example: string; note: string }
  /** memory primitives run before generation */
  retrieve: {
    title: string
    role: string
    recall: { cmd: string; desc: string }
    fetch: { cmd: string; desc: string }
    instructions: string
    note: string
  }
  /** the external agent writes SQL against MDL objects */
  generate: {
    title: string
    role: string
    points: string[]
    note: string
  }
  /** plan + validate primitives */
  plan: {
    title: string
    role: string
    steps: NamedItem[]
    sqlBefore: string
    sqlAfter: string
    note: string
  }
  execute: { title: string; role: string; steps: string[]; output: string }
  /** repair / clarify loop */
  repair: { title: string; items: NamedItem[]; note: string }
  /** closed-loop memory store */
  store: { cmd: string; note: string }
  /** the correctness primitives legend (right demo) */
  primitives: NamedItem[]
  insights: {
    input: string
    retrieve: Insight[]
    generate: Insight[]
    plan: Insight[]
    store: Insight[]
  }
}

export interface WrenModuleData {
  id: string
  accent: AccentKey
  mdl?: MdlArch
  query?: QueryArch
}

const mdlArch: MdlArch = {
  id: 'mdl',
  input: {
    label: '物理 Schema（warehouse 表）',
    note: '已有数仓 / 转换管道 / 既有语义层 —— WrenAI 叠加其上，不替换',
  },
  entities: [
    {
      title: 'Models',
      desc: '逻辑表：把仓库对象映射成业务实体',
      icon: 'i-lucide-box',
      accent: 'emerald',
      items: [
        { name: 'columns', desc: '业务名 / 类型 / 主键 / not_null / 枚举' },
        { name: 'calculatedFields', desc: '声明式派生列（经关系展开）' },
        { name: 'table_reference', desc: 'catalog · schema · table 物理绑定' },
        { name: 'ref_sql', desc: '用 SQL 定义模型来源（可选）' },
      ],
    },
    {
      title: 'Relationships',
      desc: '实体间 join 语义',
      icon: 'i-lucide-spline',
      accent: 'blue',
      items: [
        { name: 'one_to_many', desc: '1-N（如 customers → orders）' },
        { name: 'many_to_one', desc: 'N-1' },
        { name: 'one_to_one', desc: '1-1' },
        { name: 'join key', desc: '声明连接键，生成时自动展开 join path' },
      ],
    },
    {
      title: 'Views / Cubes',
      desc: '复用查询 + 预聚合指标',
      icon: 'i-lucide-layers',
      accent: 'violet',
      items: [
        { name: 'views', desc: 'statement 原样作为 CTE 复用（不经 wren-core）' },
        { name: 'cubes.measures', desc: '可复用度量（指标口径集中定义）' },
        { name: 'cubes.dimensions', desc: '维度' },
        { name: 'timeDimensions', desc: '时间维度' },
        { name: 'hierarchies', desc: '层级' },
      ],
    },
  ],
  govern: {
    title: '访问控制（建模即治理）',
    items: [
      { name: 'rowLevelAccessControls (RLAC)', desc: '模型级行过滤：condition = SQL 表达式 + requiredProperties' },
      { name: 'columnLevelAccessControl (CLAC)', desc: '列级：operator / threshold / requiredProperties，运行时按 session property 判定' },
      { name: '列暴露控制', desc: '未在模型列出的物理列对 Agent 不可见' },
    ],
  },
  compile: {
    cmd: 'wren context build',
    out: 'target/mdl.json',
    note: '源 YAML（可评审、Git diff）编译成 camelCase manifest，base64 喂给引擎；instructions.md 不进 manifest',
  },
  contextLayers: [
    { name: 'Structural', desc: '数据集 / 列 / 类型 / 键 / 关系 —— MDL 承载' },
    { name: 'Semantic', desc: '业务名 / 描述 / 计算 / views / cubes —— MDL 承载' },
    { name: 'Business', desc: '权威表 / 可复用指标 / 分析接口 —— MDL 承载' },
    { name: 'Examples', desc: 'queries.yml / memory query_history —— 由记忆层承载' },
    { name: 'Operational', desc: 'instructions.md —— 注入 system prompt' },
  ],
  yamlExample: `name: customers
table_reference:
  catalog: jaffle_shop
  schema: main
  table: customers
primary_key: customer_id
columns:
  - name: customer_id
    type: INTEGER
    is_primary_key: true
  - name: lifetime_value
    type: DOUBLE
    expression: SUM(orders.total)   # calculatedField`,
  jsonExample: `// target/mdl.json (编译后, 引擎可读)
{ "models": [ {
  "name": "customers",
  "tableReference": { "catalog": "jaffle_shop",
    "schema": "main", "table": "customers" },
  "primaryKey": "customer_id",
  "columns": [ /* … */ ],
  "rowLevelAccessControls": [ /* RLAC */ ]
} ], "relationships": [ /* … */ ] }`,
  insights: {
    input: 'WrenAI 坐在你已有的栈之上：数仓、转换管道、既有语义层都不动，只在其上叠加一层可移植的业务语义契约。',
    model: [
      { icon: 'i-lucide-file-signature', title: '语义层即契约', body: 'MDL 是数据团队、Agent、查询引擎三方共同遵守的合约：data team 评审业务逻辑、Agent 据此选模型/join/计算、引擎据此规划 SQL。与 ATLAS 用 Agent 自动生成 Rich Context 的路线相反，WrenAI 强调人工建模的稳定契约。' },
      { icon: 'i-lucide-git-pull-request', title: 'Git 化、可评审', body: '模型存为可读 YAML，契约变更走 diff 评审；这让语义层像代码一样被版本化与审查，而非藏在某个应用元数据库里。' },
      { icon: 'i-lucide-shield', title: '建模即治理', body: 'RLAC / CLAC 与列暴露直接写在模型里，权限是语义层的一等公民，规划 SQL 时由 wren-core 强制执行，而非事后在应用层补。' },
    ],
    compile: [
      { icon: 'i-lucide-box', title: '稳定接口隔离漂移', body: 'target/mdl.json 是引擎规划 SQL 的真相源；底层仓库结构变化时，模型契约作为稳定接口吸收变化，上层 Agent 无需重新理解业务。' },
    ],
  },
}

const queryArch: QueryArch = {
  id: 'query',
  input: {
    label: '用户业务问题',
    example: '"上月每个客户的订单总额"',
    note: '由外部 Agent 接收（WrenAI 不内置对话 / LLM）',
  },
  retrieve: {
    title: 'Memory · 召回 + 取上下文',
    role: 'schema linking + few-shot',
    recall: { cmd: 'wren memory recall', desc: '向量召回相似的、已确认 NL-SQL 对作为 few-shot' },
    fetch: { cmd: 'wren memory fetch', desc: '取相关 schema_items（小 schema 全量、大 schema 走向量检索）' },
    instructions: 'wren context instructions：业务规则注入 system prompt',
    note: '检索策略按 schema 体量自适应：≤30k 字符全量注入，超出才走向量检索（无 BM25 / rerank）',
  },
  generate: {
    title: 'Agent 写 SQL（外置 LLM）',
    role: '针对 MDL 模型名',
    points: [
      'Agent 用召回的样例 + schema 上下文 + instructions 直接写 SQL',
      'SQL 针对 MDL 模型 / 列名书写，而非物理表名',
      'WrenAI 本身不生成 SQL —— 生成质量取决于你选的 Agent / 模型',
    ],
    note: 'Skills（Markdown 工作流）约束 Agent 的步骤顺序：先取上下文再写、先验证再执行、成功才记忆',
  },
  plan: {
    title: 'Planning · dry-plan',
    role: 'Modeled SQL → 可执行 SQL',
    steps: [
      { name: 'sqlglot parse / qualify', desc: '解析并限定表 / 列引用' },
      { name: 'ManifestExtractor', desc: '只抽取本次查询需要的最小 MDL 切片' },
      { name: 'wren-core transform', desc: 'Rust/DataFusion 展开模型 / 关系 / 计算列' },
      { name: 'CTE Rewriter', desc: '把展开后的模型 SQL 注入为 CTE' },
      { name: 'Policy checks', desc: 'strict mode / RLAC / CLAC / denied funcs' },
      { name: 'transpile', desc: 'sqlglot 转译到目标方言' },
    ],
    sqlBefore: `-- 针对 MDL 模型写的 SQL
SELECT customer_id, lifetime_value
FROM customers`,
    sqlAfter: `-- dry-plan 展开后（注入 CTE + 展开计算列）
WITH customers AS (
  SELECT c.customer_id,
         SUM(o.total) AS lifetime_value
  FROM jaffle_shop.main.customers c
  LEFT JOIN jaffle_shop.main.orders o
    ON c.customer_id = o.customer_id
  GROUP BY 1
)
SELECT customer_id, lifetime_value FROM customers`,
    note: 'dry-plan 不连库即可展示生成轨迹（用了哪些模型 / join / CTE）—— 这就是 WrenAI 的可解释性来源',
  },
  execute: {
    title: 'Execute',
    role: '连接器执行',
    steps: [
      'dry-run：LIMIT 0 对 live DB 校验，不返回行',
      'wren query：连接器执行，默认 limit 100 / 硬顶 1000',
      '结果回 PyArrow table（→ CSV / JSON / SDK 返回）',
    ],
    output: 'PyArrow 结果集',
  },
  repair: {
    title: 'Repair / Clarify（Agent 编排）',
    items: [
      { name: 'WrenError(phase, code)', desc: '结构化错误：plan / dry-run / execute 各阶段定位' },
      { name: 'Agent retry', desc: 'Agent 据 error.phase 修正 SQL 重试（无内置 correction pipeline）' },
      { name: 'Clarify', desc: '问题欠定义时由 Skill 引导 Agent 主动澄清' },
    ],
    note: '正确性 = 一组 Agent 可编排的原语（fetch / recall / dry-plan / dry-run / repair / clarify），而非平台内一个隐藏特性',
  },
  store: {
    cmd: 'wren memory store --nl … --sql …',
    note: '确认无误的 NL-SQL 对写回 query_history，成为未来 few-shot —— 使用即沉淀（闭环，但非自动微调）',
  },
  primitives: [
    { name: 'wren memory fetch', desc: 'schema linking' },
    { name: 'wren memory recall', desc: 'few-shot 召回' },
    { name: 'wren dry-plan', desc: '展开轨迹 / 不连库' },
    { name: 'wren dry-run', desc: 'live 校验 / LIMIT 0' },
    { name: 'wren query', desc: '执行' },
    { name: 'wren memory store', desc: '记忆回流' },
  ],
  insights: {
    input: '入口只有一句业务问题，且由外部 Agent 接收 —— WrenAI 把"对话与生成"交给你已有的 Agent，自己专注上下文与正确性。',
    retrieve: [
      { icon: 'i-lucide-layers', title: '自适应注入', body: '小 schema 直接全量注入、大 schema 才走向量检索；这是体量驱动的简单策略（无关键词 / rerank 混合召回），实现轻、可预期。' },
      { icon: 'i-lucide-history', title: '召回即 few-shot', body: 'query_history 里"以前work过"的 NL-SQL 对被向量召回作示例，把历史成功直接喂给 Agent，减少重复试错。' },
    ],
    generate: [
      { icon: 'i-lucide-bot', title: 'Agent 外置', body: '与 ATLAS（Coordinator/Worker 自有 ReAct 内核）不同，WrenAI 不内置生成；生成质量取决于你接的 Agent / 模型，平台只负责喂对上下文、给对原语。' },
      { icon: 'i-lucide-scroll-text', title: 'Skills 约束顺序', body: 'Markdown Skills 教 Agent 安全地按序使用原语（先取上下文再写、先验证再执行、成功才记忆），把"正确做法"沉淀成可发现的工作流。' },
    ],
    plan: [
      { icon: 'i-lucide-cpu', title: '语义层是真相源', body: 'wren-core（Rust/DataFusion）是 MDL → SQL 的唯一真相源：关系、计算列、views 的展开规则集中在引擎，保证生成的 SQL 与语义契约一致。' },
      { icon: 'i-lucide-route', title: 'dry-plan = 可解释性', body: 'dry-plan 不连库就展开出最终 SQL（用了哪些模型 / join / CTE），既是验证也是生成轨迹的解释 —— 把"黑盒生成"变成可检视的规划。' },
      { icon: 'i-lucide-shield-check', title: '正确性即原语', body: 'schema linking / 值剖析 / 歧义检测 / 生成轨迹 / 重试修复 / eval 被拆成独立原语，由 Agent 按需编排，而非塞进一个隐藏特性。' },
    ],
    store: [
      { icon: 'i-lucide-recycle', title: '使用即沉淀', body: 'confirmed 的 NL-SQL 回写 memory 成为后续 few-shot；这是闭环复用，但不是自动 ML 微调 —— 沉淀是显式、可评审的。' },
    ],
  },
}

export const WREN_MODULES: Record<string, WrenModuleData> = {
  mdl: { id: 'mdl', accent: 'emerald', mdl: mdlArch },
  query: { id: 'query', accent: 'violet', query: queryArch },
}

export function getWrenModule(id: string | null): WrenModuleData | null {
  if (!id) return null
  return WREN_MODULES[id] ?? null
}
