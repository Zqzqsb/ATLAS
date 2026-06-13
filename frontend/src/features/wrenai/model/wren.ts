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
        flow: 'skills',
        span: 1,
      },
      {
        id: 'skills',
        label: 'Agent Skills',
        sublabel: 'Markdown 工作流：先建模 → 取上下文 → 写 SQL → 验证 → 记忆',
        icon: 'i-lucide-scroll-text',
        accent: 'slate',
        flow: 'skills',
        span: 1,
        codeRefs: ['skills/wren/SKILL.md', 'core/wren/src/wren/skills_content/'],
      },
      {
        id: 'sdk',
        label: 'Agent SDK',
        sublabel: 'wren-langchain · wren-pydantic（把原语包装成工具）',
        icon: 'i-lucide-plug',
        accent: 'slate',
        flow: 'skills',
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
        flow: 'memory',
        span: 1,
        codeRefs: ['core/wren/src/wren/memory/schema_indexer.py'],
      },
      {
        id: 'query-history',
        label: 'query_history',
        sublabel: '确认过的 NL-SQL 对（recall 复用）',
        icon: 'i-lucide-history',
        accent: 'blue',
        flow: 'memory',
        span: 1,
        codeRefs: ['core/wren/src/wren/memory/store.py'],
      },
      {
        id: 'embed',
        label: 'Embeddings',
        sublabel: 'sentence-transformers · 多语 MiniLM · 384d',
        icon: 'i-lucide-spline',
        accent: 'blue',
        flow: 'memory',
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
        flow: 'planning',
        span: 1,
      },
      {
        id: 'cte',
        label: 'CTE Rewriter',
        sublabel: '识别引用模型 → 注入展开后的 CTE',
        icon: 'i-lucide-git-merge',
        accent: 'amber',
        flow: 'planning',
        span: 1,
        codeRefs: ['core/wren/src/wren/mdl/cte_rewriter.py'],
      },
      {
        id: 'wren-core',
        label: 'wren-core',
        sublabel: 'Rust · Apache DataFusion · MDL 语义展开',
        icon: 'i-lucide-cog',
        accent: 'amber',
        flow: 'planning',
        span: 1,
        codeRefs: ['core/wren-core/core/src/mdl/mod.rs'],
      },
      {
        id: 'policy',
        label: 'Policy Checks',
        sublabel: 'strict mode · RLAC/CLAC · denied funcs · row limit',
        icon: 'i-lucide-shield-check',
        accent: 'amber',
        flow: 'planning',
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
        flow: 'execution',
        span: 1,
        codeRefs: ['core/wren/src/wren/connector/factory.py'],
      },
      {
        id: 'dryrun',
        label: 'Dry-run',
        sublabel: 'LIMIT 0 对 live DB 校验，不返回行',
        icon: 'i-lucide-flask-conical',
        accent: 'indigo',
        flow: 'execution',
        span: 1,
      },
      {
        id: 'pyarrow',
        label: 'PyArrow Result',
        sublabel: 'table · CSV · JSON · SDK 返回值',
        icon: 'i-lucide-table',
        accent: 'indigo',
        flow: 'execution',
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
  {
    id: 'planning',
    label: 'Planning Engine',
    title: 'Planning Engine · Modeled SQL → 可执行 SQL',
    subtitle: 'wren-core（Rust / Apache DataFusion）是 MDL → SQL 的唯一真相源：解析限定 → 抽取最小切片 → 展开模型/关系/计算列 → 注入 CTE → 策略校验 → 转译目标方言',
    icon: 'i-lucide-cpu',
    accent: 'amber',
  },
  {
    id: 'memory',
    label: 'Memory',
    title: 'Memory · LanceDB 向量检索层',
    subtitle: 'index 把 MDL 工件向量化为 schema_items，把确认过的 NL-SQL 对沉淀进 query_history；查询时 fetch 做 schema linking、recall 做 few-shot，体量自适应',
    icon: 'i-lucide-database',
    accent: 'blue',
  },
  {
    id: 'skills',
    label: 'Agent Workflow',
    title: 'Agent Workflow · BYO Agent + Skills',
    subtitle: 'WrenAI 不内置 Agent：你已有的 Agent 通过 SDK 把原语当工具调用，Skills（Markdown 工作流）把"正确做法"沉淀成可发现、可约束步骤顺序的剧本',
    icon: 'i-lucide-scroll-text',
    accent: 'slate',
  },
  {
    id: 'execution',
    label: 'Execution Layer',
    title: 'Execution Layer · Connectors + Dry-run',
    subtitle: '可执行方言 SQL 经原生连接器跑到 live DB：凭据由 profiles 与项目分离，dry-run 先 LIMIT 0 校验，结果回 PyArrow（→ table / CSV / JSON / SDK）',
    icon: 'i-lucide-plug-zap',
    accent: 'indigo',
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
  /** the multiple provenance pathways an MDL can be created from */
  sourcing: {
    title: string
    paths: { name: string; icon: string; accent: AccentKey; badge?: string; desc: string }[]
  }
  /** demo project identity + scale (mocked from jaffle_shop) */
  project: { name: string; stats: string; desc: string }
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
    sourcing: string
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

/** Planning Engine — Modeled SQL → executable SQL (the semantic source of truth). */
export interface PlanningArch {
  id: string
  input: { label: string; note: string; example: string }
  /** the three collaborating engines */
  collaborators: { name: string; desc: string; icon: string; accent: AccentKey; lang: string }[]
  /** the numbered transform pipeline */
  steps: { name: string; desc: string }[]
  /** what wren-core expands (peek) */
  expands: NamedItem[]
  /** policy gate checks (peek) */
  policy: NamedItem[]
  /** target dialects */
  dialects: { count: string; list: string }
  output: { label: string; note: string }
  /** richer before/after expansion demo */
  sqlBefore: string
  sqlAfter: string
  insights: {
    input: string
    transform: Insight[]
    policy: Insight[]
  }
}

/** Memory — LanceDB vector layer (schema linking + few-shot recall). */
export interface MemoryArch {
  id: string
  input: { label: string; note: string }
  /** indexing step */
  index: { cmd: string; items: NamedItem[]; note: string }
  /** embedding config (mocked but realistic) */
  embed: { model: string; dim: string; engine: string; note: string }
  /** the two LanceDB collections, with mocked live counts */
  collections: { table: string; count: string; desc: string; use: string; icon: string }[]
  /** retrieval primitives */
  retrieve: { fetch: string; recall: string; strategy: string }
  /** seed + store lifecycle */
  seed: { cmd: string; note: string }
  store: { cmd: string; note: string }
  /** mocked query_history sample pairs */
  samples: { nl: string; sql: string; tag: string }[]
  insights: {
    input: string
    index: Insight[]
    retrieve: Insight[]
  }
}

/** Agent Workflow — BYO agent + SDK tools + Markdown skills. */
export interface SkillsArch {
  id: string
  input: { label: string; note: string; frameworks: string[] }
  /** SDK wraps primitives as callable tools */
  sdk: { title: string; note: string; tools: NamedItem[] }
  /** installed skills (mocked but grounded in skills_content) */
  skills: { name: string; file: string; when: string; steps: string[]; accent: AccentKey }[]
  /** `wren skills list` mocked terminal output */
  listCmd: string
  /** a sample skill markdown excerpt (right demo) */
  sampleSkill: { name: string; md: string }
  insights: {
    input: string
    sdk: Insight[]
    skills: Insight[]
  }
}

/** Execution Layer — connectors, credential separation, dry-run, results. */
export interface ExecutionArch {
  id: string
  input: { label: string; note: string }
  /** credential separation via profiles */
  profiles: { title: string; note: string; fields: NamedItem[]; example: string }
  /** grouped data sources (mocked realistic catalog) */
  connectors: { group: string; icon: string; items: string }[]
  /** dry-run validation */
  dryrun: { cmd: string; points: string[] }
  /** result formats + mocked preview */
  result: { formats: NamedItem[]; preview: { cols: string[]; rows: string[][] } }
  insights: {
    input: string
    connect: Insight[]
    execute: Insight[]
  }
}

export interface WrenModuleData {
  id: string
  accent: AccentKey
  mdl?: MdlArch
  query?: QueryArch
  planning?: PlanningArch
  memory?: MemoryArch
  skills?: SkillsArch
  execution?: ExecutionArch
}

const mdlArch: MdlArch = {
  id: 'mdl',
  sourcing: {
    title: 'MDL 从哪来 · 五条来源路径',
    paths: [
      {
        name: '人工编写 YAML',
        icon: 'i-lucide-file-pen',
        accent: 'emerald',
        desc: '手写 models/ · relationships.yml · cubes/ · views/ —— 完全可控、可评审',
      },
      {
        name: 'Agent 生成',
        icon: 'i-lucide-bot',
        accent: 'violet',
        badge: 'generate-mdl',
        desc: 'introspect 库 schema → parse-type 类型归一化 → 写 YAML；FK 推关系，缺 FK 按命名约定推断',
      },
      {
        name: '端到端引导',
        icon: 'i-lucide-rocket',
        accent: 'blue',
        badge: 'onboarding',
        desc: '环境检查 → profile 连接 → init → 生成 MDL → context build → memory index 一条龙',
      },
      {
        name: 'dbt 导入',
        icon: 'i-lucide-package',
        accent: 'amber',
        badge: 'dbt',
        desc: '读 dbt manifest.json / catalog.json，按 adapter→datasource 映射转成 MDL 模型与关系',
      },
      {
        name: 'dlt / SaaS 抽取',
        icon: 'i-lucide-cloud-download',
        accent: 'indigo',
        badge: 'dlt-connector',
        desc: 'HubSpot / Stripe / Salesforce → DuckDB 落地 → introspect 生成 MDL（SaaS 也能建模）',
      },
      {
        name: 'enrich-context 补全',
        icon: 'i-lucide-sparkles',
        accent: 'rose',
        badge: 'skill',
        desc: '在已有 MDL 上从 raw/ 文档补 enum 含义 / 单位 / NULL 语义 / 同义词 / cubes（grill · auto-pilot）',
      },
    ],
  },
  project: {
    name: 'jaffle_shop',
    stats: '6 models · 5 relationships · 2 cubes',
    desc: 'customers · orders · order_items · products · stores · supplies',
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
    sourcing: 'MDL 不是只有手写一条路：可以手写、可让 Agent 从库 introspect 生成、走 onboarding 端到端引导、从 dbt 项目导入、用 dlt 把 SaaS 抽到 DuckDB 再建模，最后再用 enrich-context 补业务语义。形式统一为可评审、可移植的项目 YAML。',
    model: [
      { icon: 'i-lucide-file-signature', title: '语义层即契约', body: 'MDL 是数据团队、Agent、查询引擎三方共同遵守的合约：data team 评审业务逻辑、Agent 据此选模型/join/计算、引擎据此规划 SQL。语义不是模型临场猜出来的，而是被人工评审过、可复用的稳定契约。' },
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
      { icon: 'i-lucide-bot', title: 'Agent 外置', body: 'WrenAI 不内置生成内核：写 SQL 这步交给你已有的 Agent / 模型，生成质量取决于它；平台只负责喂对上下文、给对原语、把住正确性。' },
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

const planningArch: PlanningArch = {
  id: 'planning',
  input: {
    label: 'Modeled SQL',
    note: 'Agent 针对 MDL 模型 / 列名书写的 SQL（不写物理表名）',
    example: 'SELECT customer_id, lifetime_value FROM customers',
  },
  collaborators: [
    { name: 'sqlglot', desc: '解析 / 限定引用 / 转译目标方言', icon: 'i-lucide-code-2', accent: 'amber', lang: 'Python' },
    { name: 'CTE Rewriter', desc: '识别引用模型 → 注入展开后的 CTE', icon: 'i-lucide-git-merge', accent: 'amber', lang: 'Python' },
    { name: 'wren-core', desc: 'MDL 语义展开（关系 / 计算列 / views）', icon: 'i-lucide-cog', accent: 'violet', lang: 'Rust · DataFusion 53' },
  ],
  steps: [
    { name: 'parse + qualify', desc: 'sqlglot 解析 SQL，限定表 / 列引用到 MDL 对象' },
    { name: 'ManifestExtractor', desc: '只抽取本次查询命中的最小 MDL 切片，避免整库注入' },
    { name: 'expand (wren-core)', desc: '展开 models / relationships / calculatedFields / views → DataFusion 逻辑计划' },
    { name: 'inject CTEs', desc: 'CTE Rewriter 把展开后的模型 SQL 注入为命名 CTE' },
    { name: 'policy checks', desc: 'strict mode / RLAC / CLAC / denied funcs / row limit 校验' },
    { name: 'transpile', desc: 'sqlglot 输出目标方言的可执行 SQL' },
  ],
  expands: [
    { name: 'models', desc: '逻辑表 → 物理 table_reference / ref_sql 的 CTE' },
    { name: 'relationships', desc: '声明的 join key → 自动展开 join path（无需手写 join）' },
    { name: 'calculatedFields', desc: '派生列表达式（经关系展开为聚合 / 子查询）' },
    { name: 'views', desc: 'statement 原样作为 CTE 复用（不经语义展开）' },
    { name: 'cubes', desc: 'measures / dimensions → 预聚合 SQL' },
  ],
  policy: [
    { name: 'strict_mode', desc: '禁止裸物理表 / 未建模列，强制走 MDL' },
    { name: 'rowLevelAccessControls', desc: '按 session property 注入行过滤 WHERE' },
    { name: 'columnLevelAccessControl', desc: 'operator / threshold 判定列是否可见' },
    { name: 'denied_functions', desc: '黑名单函数（如危险 UDF）直接拒绝' },
    { name: 'row limit', desc: '默认 limit 100 / 硬顶 1000' },
  ],
  dialects: {
    count: '20+',
    list: 'Postgres · MySQL · BigQuery · Snowflake · DuckDB · Trino · ClickHouse · Databricks · Redshift · Oracle · Athena · Spark …',
  },
  output: { label: '可执行方言 SQL', note: '交连接器 dry-run 校验后执行' },
  sqlBefore: `-- Agent 针对 MDL 模型写的 SQL
SELECT customer_id, lifetime_value
FROM customers
WHERE lifetime_value > 1000
ORDER BY lifetime_value DESC`,
  sqlAfter: `-- wren-core 展开后（注入 CTE · 展开计算列 · 自动 join · 注入 RLAC）
WITH customers AS (
  SELECT c.customer_id,
         SUM(o.total) AS lifetime_value   -- calculatedField 经 1-N 关系展开
  FROM jaffle_shop.main.customers c
  LEFT JOIN jaffle_shop.main.orders o
    ON c.customer_id = o.customer_id
  WHERE c.region = $session.region          -- RLAC 行级过滤注入
  GROUP BY c.customer_id
)
SELECT customer_id, lifetime_value
FROM customers
WHERE lifetime_value > 1000
ORDER BY lifetime_value DESC
LIMIT 100                                    -- row limit 策略`,
  insights: {
    input: 'Agent 只写"模型语言"的 SQL——引用 MDL 的模型名 / 列名 / 计算列，不碰物理表名、不手写 join。把"翻译成正确物理 SQL"这件事整体交给引擎。',
    transform: [
      { icon: 'i-lucide-cpu', title: 'wren-core 是唯一真相源', body: 'MDL → SQL 的展开规则（关系 join、计算列、views）集中在 Rust/DataFusion 引擎里，保证任何 Agent、任何方言下生成的 SQL 都与语义契约一致——正确性不依赖 prompt。' },
      { icon: 'i-lucide-scissors', title: '最小切片注入', body: 'ManifestExtractor 只挑出本次查询命中的模型 / 关系，而不是把整个 manifest 塞进去；这让规划在大 schema 下依然轻、可预期。' },
      { icon: 'i-lucide-route', title: 'dry-plan = 可解释', body: '展开后的 SQL（用了哪些 CTE / join / 计算列）就是生成轨迹，可以不连库直接看；把"黑盒生成"变成可检视的逻辑计划。' },
    ],
    policy: [
      { icon: 'i-lucide-shield-check', title: '建模即治理、规划即执法', body: 'RLAC / CLAC / denied funcs 不是事后在应用层补的，而是在规划阶段由引擎强制注入到 SQL 里——权限是语义层的一等公民。' },
    ],
  },
}

const memoryArch: MemoryArch = {
  id: 'memory',
  input: { label: 'MDL + instructions', note: 'wren memory index 解析项目工件作为输入' },
  index: {
    cmd: 'wren memory index',
    items: [
      { name: 'model', desc: '每个逻辑表一条（名 + 描述 + 列摘要）' },
      { name: 'column', desc: '业务名 / 类型 / 枚举 / 描述' },
      { name: 'relationship', desc: 'join 语义（1-N / N-1 …）' },
      { name: 'view / cube', desc: '复用查询与预聚合指标' },
      { name: 'measure / dimension', desc: 'cube 内度量与维度' },
      { name: 'instruction', desc: 'instructions.md 段落（业务规则）' },
    ],
    note: 'schema_indexer 把每类工件拆成可检索条目，逐条 embed 后写入 LanceDB',
  },
  embed: {
    model: 'paraphrase-multilingual-MiniLM-L12-v2',
    dim: '384d',
    engine: 'sentence-transformers（本地推理，无需外部 API）',
    note: 'LanceDB 自带 embedding registry，索引与查询用同一模型',
  },
  collections: [
    { table: 'schema_items', count: '142 条', desc: 'models / 列 / 关系 / views / cubes（+ instructions）', use: 'schema linking', icon: 'i-lucide-table-2' },
    { table: 'query_history', count: '37 对', desc: '确认过的 NL-SQL 对（含 seed + 真实使用沉淀）', use: 'few-shot 召回', icon: 'i-lucide-history' },
  ],
  retrieve: {
    fetch: 'wren memory fetch — 召回相关 schema_items 做 schema linking',
    recall: 'wren memory recall — 召回相似 NL-SQL 对做 few-shot',
    strategy: 'get_context 体量自适应：schema ≤ 30k 字符时全量注入，超出才走向量检索（无 BM25 / rerank 混合召回）',
  },
  seed: {
    cmd: 'wren memory seed-queries',
    note: '从 MDL 自动生成 canonical NL-SQL 对（打 tag source:seed），冷启动即有 few-shot',
  },
  store: {
    cmd: 'wren memory store --nl … --sql …',
    note: 'Agent 确认无误的 NL-SQL 对追加进 query_history（tag source:confirmed），使用即沉淀',
  },
  samples: [
    {
      nl: '上月每个客户的订单总额',
      sql: 'SELECT customer_id, SUM(total) AS revenue\nFROM orders\nWHERE order_date >= date_trunc(\'month\', current_date - interval \'1 month\')\nGROUP BY customer_id',
      tag: 'confirmed',
    },
    {
      nl: '销量 Top 5 的产品',
      sql: 'SELECT product_id, SUM(quantity) AS qty\nFROM order_items\nGROUP BY product_id\nORDER BY qty DESC\nLIMIT 5',
      tag: 'confirmed',
    },
    {
      nl: '每家门店本月营收',
      sql: 'SELECT store_id, SUM(total) AS revenue\nFROM orders\nWHERE order_date >= date_trunc(\'month\', current_date)\nGROUP BY store_id',
      tag: 'seed',
    },
  ],
  insights: {
    input: '记忆层的输入不是原始数据库，而是已经建模好的 MDL + instructions——记忆检索的是"语义条目"和"历史成功"，而非物理 schema。',
    index: [
      { icon: 'i-lucide-boxes', title: '工件 → 可检索条目', body: 'schema_indexer 把 MDL 拆成 model / 列 / 关系 / view / cube / instruction 等细粒度条目逐条向量化；检索精度落在"哪个模型、哪一列"，而非整张表。' },
      { icon: 'i-lucide-spline', title: '本地多语 embedding', body: '用 sentence-transformers 多语 MiniLM（384d）本地推理，索引和查询同模型；不依赖外部 embedding API，中英文混合 schema 也能召回。' },
    ],
    retrieve: [
      { icon: 'i-lucide-layers', title: '体量自适应、实现轻', body: '小 schema 直接全量注入、大 schema 才走向量检索——一个由体量驱动的简单阈值策略，没有关键词 / rerank 混合管道，行为可预期、易调试。' },
      { icon: 'i-lucide-recycle', title: '使用即沉淀的 few-shot', body: 'seed 提供冷启动样例，confirmed 把真实成功回写 query_history；recall 把"以前 work 过"的 NL-SQL 对喂给 Agent，越用越准——是显式可评审的闭环，而非自动微调。' },
    ],
  },
}

const skillsArch: SkillsArch = {
  id: 'skills',
  input: {
    label: 'BYO Agent（外置）',
    note: 'WrenAI 不内置对话 / LLM 服务；你已有的 Agent 直接驱动整个流程',
    frameworks: ['LangChain', 'LangGraph', 'Pydantic AI', 'Claude / Cursor 等 coding agent'],
  },
  sdk: {
    title: 'Agent SDK · 原语即工具',
    note: 'wren-langchain / wren-pydantic 把 CLI 原语包装成 Agent 可调用的 tool，结构化入参出参',
    tools: [
      { name: 'memory.fetch', desc: 'schema linking：取相关 schema_items' },
      { name: 'memory.recall', desc: 'few-shot：召回相似 NL-SQL 对' },
      { name: 'context.instructions', desc: '取业务规则注入 system prompt' },
      { name: 'dry_plan', desc: '不连库展开 modeled SQL，返回轨迹' },
      { name: 'dry_run', desc: 'LIMIT 0 live 校验' },
      { name: 'query', desc: '执行并返回 PyArrow 结果' },
      { name: 'memory.store', desc: '确认对回写 query_history' },
    ],
  },
  skills: [
    {
      name: 'onboarding',
      file: 'skills_content/onboarding.md',
      when: '接入一个新数据源 / 新项目时',
      steps: ['检查环境与 wren 版本', '配置 profile 连接', 'wren init 建项目', '生成 MDL', 'context build 编译', 'memory index 建索引'],
      accent: 'blue',
    },
    {
      name: 'generate-mdl',
      file: 'skills_content/generate-mdl.md',
      when: '需要从已有库 schema 起一份 MDL',
      steps: ['introspect 库表 / 列 / 主键', 'parse-type 归一化类型', 'FK → relationships（缺 FK 按命名推断）', '写 models/ · relationships.yml', '人工评审'],
      accent: 'violet',
    },
    {
      name: 'enrich-context',
      file: 'skills_content/enrich-context.md',
      when: 'MDL 已有但语义稀薄、生成不准时',
      steps: ['扫 raw/ 文档与样例查询', '补列描述 / enum 含义 / 单位 / NULL 语义', '加同义词与业务别名', '提炼可复用 cubes / metrics', 'grill 复核 · auto-pilot 批量'],
      accent: 'rose',
    },
    {
      name: 'dlt-connector',
      file: 'skills_content/dlt-connector.md',
      when: '数据在 SaaS（HubSpot / Stripe …）里',
      steps: ['配置 dlt pipeline', '抽取 SaaS → DuckDB 落地', 'introspect 生成 MDL', '增量调度刷新'],
      accent: 'indigo',
    },
    {
      name: 'query',
      file: 'skills_content/query.md',
      when: '把一个业务问题安全地变成 SQL 答案',
      steps: ['recall + fetch 取上下文', '针对 MDL 模型写 SQL', 'dry-plan 检查展开轨迹', 'dry-run 校验', 'query 执行', '确认后 memory store 沉淀'],
      accent: 'emerald',
    },
  ],
  listCmd: `$ wren skills list
  onboarding        接入新数据源的端到端引导
  generate-mdl      从库 schema 生成 MDL
  enrich-context    补全业务语义（enum/单位/同义词/cubes）
  dlt-connector     从 SaaS 抽数到 DuckDB 再建模
  query             安全跑 NL2SQL（取上下文→写→验证→记忆）
5 skills · 来源 core/wren/src/wren/skills_content/`,
  sampleSkill: {
    name: 'query.md',
    md: `# Skill: query
当用户问一个数据问题时，按序使用原语，不要跳步：

1. \`memory.recall(nl)\` — 找相似的、已确认的 SQL 当 few-shot
2. \`memory.fetch(nl)\` — 取相关模型 / 列做 schema linking
3. 针对 **MDL 模型名** 写 SQL（不要写物理表名）
4. \`dry_plan(sql)\` — 看展开轨迹；用了对的 join / 计算列吗？
5. \`dry_run(sql)\` — LIMIT 0 校验语法与权限
6. \`query(sql)\` — 执行，返回结果
7. 用户确认无误 → \`memory.store(nl, sql)\` 沉淀

> 失败时读 WrenError.phase 定位（plan / dry-run / execute）并修复重试。`,
  },
  insights: {
    input: '入口是你已有的 Agent —— WrenAI 把"对话与生成"完全外置，自己只暴露上下文与正确性原语。换任何 Agent / 模型都不动平台。',
    sdk: [
      { icon: 'i-lucide-plug', title: '原语即工具', body: 'SDK 把 fetch / recall / dry-plan / dry-run / query / store 包装成结构化 tool，Agent 用 function-calling 直接编排，无需理解 CLI 细节。' },
      { icon: 'i-lucide-boxes', title: '正确性是可组合的', body: '每个原语职责单一、可独立调用，Agent 按需拼装出"取上下文→写→验证→执行→记忆"的链路——正确性是一组积木，不是一个黑盒特性。' },
    ],
    skills: [
      { icon: 'i-lucide-scroll-text', title: 'Skills 约束步骤顺序', body: 'Markdown 工作流把"先取上下文再写、先验证再执行、成功才记忆"写成 Agent 可发现的剧本，避免它跳步乱来——把团队的最佳实践沉淀成可执行文档。' },
      { icon: 'i-lucide-git-fork', title: '覆盖建模到查询全链路', body: 'onboarding / generate-mdl / enrich-context / dlt-connector / query 分别对应接入、起模、补义、抽数、查询——同一套 skill 机制贯穿语义层的整个生命周期。' },
    ],
  },
}

const executionArch: ExecutionArch = {
  id: 'execution',
  input: { label: '可执行方言 SQL', note: '由 Planning Engine 转译输出，已是目标库方言' },
  profiles: {
    title: 'Profiles · 凭据与项目分离',
    note: '连接信息存 ~/.wren，与可 Git 化的项目工件（MDL / instructions）分离，项目可安全共享',
    fields: [
      { name: 'name', desc: 'profile 名（query 时 --profile 指定）' },
      { name: 'type', desc: 'datasource 类型（postgres / bigquery …）' },
      { name: 'connection', desc: 'host / port / database / schema' },
      { name: 'credentials', desc: '密钥引用（env / keyring，不落项目）' },
    ],
    example: `# ~/.wren/profiles.yml
jaffle_pg:
  type: postgres
  host: \${WREN_PG_HOST}
  port: 5432
  database: jaffle_shop
  user: \${WREN_PG_USER}
  password: \${WREN_PG_PASSWORD}`,
  },
  connectors: [
    { group: '数仓 / OLAP', icon: 'i-lucide-warehouse', items: 'BigQuery · Snowflake · Redshift · Databricks · ClickHouse' },
    { group: 'OLTP', icon: 'i-lucide-database', items: 'Postgres · MySQL · SQL Server · Oracle' },
    { group: 'Lake / 文件', icon: 'i-lucide-folder-tree', items: 'DuckDB · Trino · Athena · Spark · S3/Parquet' },
    { group: 'SaaS（经 dlt）', icon: 'i-lucide-cloud', items: 'HubSpot · Stripe · Salesforce · Notion' },
  ],
  dryrun: {
    cmd: 'wren dry-run',
    points: [
      'LIMIT 0 提交到 live DB：验证语法 / 列存在 / 权限，不返回行、几乎零成本',
      '校验失败抛 WrenError(phase=dry-run, code) 供 Agent 定位修复',
      '通过后才走真正执行，避免错 SQL 直接打到生产库',
    ],
  },
  result: {
    formats: [
      { name: 'table', desc: '终端表格（默认）' },
      { name: 'CSV / JSON', desc: '导出文件' },
      { name: 'PyArrow', desc: 'SDK 返回 Arrow table（零拷贝转 pandas / polars）' },
    ],
    preview: {
      cols: ['customer_id', 'lifetime_value'],
      rows: [
        ['1042', '8,930.50'],
        ['1071', '7,415.00'],
        ['1008', '6,220.75'],
        ['1135', '5,980.20'],
      ],
    },
  },
  insights: {
    input: '执行层拿到的已是某个具体方言的可执行 SQL —— 语义正确性在上游 Planning 已保证，这一层只管"连得上、跑得稳、收得回"。',
    connect: [
      { icon: 'i-lucide-key-round', title: '凭据与项目分离', body: 'MDL / instructions / queries 可 Git 化、可评审、可共享；连接凭据单独存 ~/.wren，避免密钥进仓库——语义层可移植、连接环境各自配。' },
      { icon: 'i-lucide-plug', title: '原生连接器、20+ 数据源', body: '从数仓到 OLTP 到 SaaS（经 dlt 落 DuckDB）统一接入；同一份 MDL 配不同 profile 就能切库，建模与执行解耦。' },
    ],
    execute: [
      { icon: 'i-lucide-flask-conical', title: 'dry-run 先于执行', body: 'LIMIT 0 几乎零成本地在 live DB 上验证语法 / 列 / 权限，错 SQL 不会直接打到生产；失败抛结构化 WrenError 供 Agent 修复重试。' },
      { icon: 'i-lucide-table', title: 'PyArrow 作为返回边界', body: '结果统一回 PyArrow，可零拷贝转 pandas / polars 或导出 CSV / JSON；执行层只到"返回数据"为止，可视化 / 报表交给上层。' },
    ],
  },
}

export const WREN_MODULES: Record<string, WrenModuleData> = {
  mdl: { id: 'mdl', accent: 'emerald', mdl: mdlArch },
  query: { id: 'query', accent: 'violet', query: queryArch },
  planning: { id: 'planning', accent: 'amber', planning: planningArch },
  memory: { id: 'memory', accent: 'blue', memory: memoryArch },
  skills: { id: 'skills', accent: 'slate', skills: skillsArch },
  execution: { id: 'execution', accent: 'indigo', execution: executionArch },
}

export function getWrenModule(id: string | null): WrenModuleData | null {
  if (!id) return null
  return WREN_MODULES[id] ?? null
}
