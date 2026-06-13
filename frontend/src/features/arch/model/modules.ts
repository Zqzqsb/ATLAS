/**
 * Per-module internal architecture — the deepest (non-expandable) level.
 *
 * Each module renders as ONE architecture diagram with clear module boundaries.
 * Long lists (prompt rules, RC types, storage tables) live here as data; the
 * <XxxDetail> composition wires them into the diagram primitives (ArchBox /
 * Connector / PeekPanel). Ground every value in the real backend code.
 */
import type { AccentKey } from './architecture'

export interface NamedItem {
  name: string
  desc: string
}
export interface PromptBlock {
  label: string
  desc: string
}
export interface StorageItem {
  table: string
  label: string
  spec: string
  note: string
}
export interface Insight {
  icon: string
  title: string
  body: string
}

/** Onboarding internal architecture: Coordinator → Worker(×N) → Storage. */
export interface OnboardingArch {
  id: string
  input: { table: string; label: string; note: string }
  coordinator: {
    title: string
    role: string
    points: string[]
    /** small-DB shortcut annotation — the unified path's degenerate case */
    note: string
  }
  worker: {
    title: string
    role: string
    /** how the Coordinator fans out to workers */
    dispatch: string
    prompt: {
      engine: string
      blocks: PromptBlock[]
      /** peek-on-demand: key constraints/tricks baked into the system prompt */
      rules: string[]
    }
    tools: NamedItem[]
    loop: string
    /** peek-on-demand: the Rich Context the worker produces */
    output: { label: string; store: string; types: NamedItem[] }
  }
  storage: {
    title: string
    items: StorageItem[]
  }
  /** side-lane annotations explaining the engineering design, aligned to stages */
  insights: {
    input: string
    process: Insight[]
    storage: Insight[]
  }
}

/** Inference internal architecture: Grounding → SQL Generation → Execute. */
export interface InferenceArch {
  id: string
  input: { label: string; note: string; example: string }
  /** Adaptive Grounding (Schema Linking): dispatcher + retriever + agent. */
  grounding: {
    title: string
    role: string
    /** strategy detection (small/large) + unified-path note */
    dispatcher: { points: string[]; note: string }
    /** CoarseRetriever — 4 parallel HNSW signals (peek) */
    retriever: { title: string; desc: string; signals: NamedItem[] }
    /** LinkingAgent — 3 linking modes (peek) + concurrency timing note */
    agent: { title: string; engine: string; modes: NamedItem[]; concurrency: string }
    output: { label: string; parts: string[] }
  }
  /** SQL Generation: ReAct SQLGen agent with verify gate. */
  sqlgen: {
    title: string
    role: string
    prompt: { engine: string; blocks: PromptBlock[]; rules: string[] }
    tools: NamedItem[]
    loop: string
    /** verify_sql gate annotation */
    verify: string
    output: { label: string; note: string }
  }
  execute: {
    title: string
    role: string
    steps: string[]
    output: string
  }
  /** side dependency — storage tables this pipeline reads from */
  reads: { label: string; items: { table: string; use: string }[] }
  insights: {
    input: string
    grounding: Insight[]
    sqlgen: Insight[]
    execute: Insight[]
  }
}

/** Self-Maintenance internal architecture: Signal → Coordinator → Executor → Re-embed. */
export interface MaintainArch {
  id: string
  /** schema-change signal that kicks off maintenance */
  trigger: {
    label: string
    note: string
    /** DDL change types ParseDDLStatement recognizes */
    changeTypes: NamedItem[]
    /** how the change is detected + catalog synced (not a diff) */
    detect: string
  }
  /** Coordinator ReAct agent: decide tasks + invalidate stale RC */
  coordinator: {
    title: string
    role: string
    engine: string
    /** decision guide points */
    points: string[]
    /** tools (peek) */
    tools: NamedItem[]
    budget: string
    output: { label: string; parts: string[] }
  }
  /** Executor ReAct agent: heal (regenerate / delete) per task */
  executor: {
    title: string
    role: string
    engine: string
    /** how tasks are handed over */
    dispatch: string
    steps: string[]
    /** tools (peek) */
    tools: NamedItem[]
    budget: string
    /** skip-when-no-task annotation */
    note: string
  }
  /** invalidation soft-flags + re-embed at storage layer */
  storage: {
    title: string
    /** the three soft-delete / staleness flags (peek) */
    flags: NamedItem[]
    items: StorageItem[]
    /** full re-embed annotation */
    embed: string
  }
  insights: {
    trigger: string
    coordinator: Insight[]
    executor: Insight[]
    storage: Insight[]
  }
}

/** ReAct Kernel internal architecture: Scenario config → Engine loop → Tool Belt. */
export interface KernelArch {
  id: string
  /** scenarios parameterize the shared kernel via EngineConfig */
  scenarios: {
    title: string
    role: string
    /** EngineConfig fields a scenario sets */
    config: NamedItem[]
    /** the scenario builders + their tool subset / budget */
    list: { name: string; tools: string; budget: string }[]
    note: string
  }
  /** the ReAct loop itself (langchaingo ZeroShotReactDescription) */
  engine: {
    title: string
    role: string
    base: string
    /** Reason / Act / Observe phases */
    loop: NamedItem[]
    format: string
    parser: string
    budget: string
  }
  /** a sample transcript (right demo) */
  transcript: { kind: 'thought' | 'action' | 'input' | 'observation' | 'final'; text: string }[]
  /** the Tool Belt, grouped */
  toolGroups: { name: string; icon: string; accent: AccentKey; items: NamedItem[] }[]
  /** LLM provider + step/SSE output */
  llm: { provider: string; encoding: string; note: string }
  output: { label: string; parts: string[] }
  insights: {
    scenarios: Insight[]
    engine: Insight[]
    tools: Insight[]
  }
}

/** Lakebase Storage internal architecture: datasource root → RC tables → vectors. */
export interface StorageArch {
  id: string
  /** the FK-cascade root */
  root: { table: string; label: string; note: string; cascade: string }
  /** RC metadata + semantic tables */
  tables: {
    title: string
    role: string
    items: { table: string; label: string; cols: string; flag?: string; accent: AccentKey }[]
  }
  /** simple ER mini-map (right demo) */
  er: { root: string; edges: { table: string; rel: string }[] }
  /** the vector layer */
  vector: {
    title: string
    role: string
    spec: NamedItem[]
    ddl: string
    search: string
    searchNote: string
  }
  /** embedding write path */
  embed: { provider: string; dim: string; model: string; paths: string[]; upsert: string }
  /** change log + invalidation flags */
  changelog: { table: string; types: NamedItem[]; note: string }
  flags: NamedItem[]
  insights: {
    root: string
    tables: Insight[]
    vector: Insight[]
  }
}

export interface ModuleData {
  id: string
  accent: AccentKey
  onboarding?: OnboardingArch
  inference?: InferenceArch
  maintain?: MaintainArch
  kernel?: KernelArch
  storage?: StorageArch
}

const onboardingArch: OnboardingArch = {
  id: 'onboarding',
  input: {
    table: 'rc_tables · rc_columns · rc_relations',
    label: '物理 Schema',
    note: '从 INFORMATION_SCHEMA 同步的表 / 列 / 外键关系',
  },
  coordinator: {
    title: 'Coordinator',
    role: '切分 + 分发',
    points: [
      'Forest Decompose：外键无向图 → BFS 连通分量（表簇）',
      '孤立无 FK 表按 15 张/批合并',
      '逐簇计算迭代预算（tables×3+10，上限 60→500）',
      '每个表簇打包成一个 task 下发',
    ],
    note: '大小库统一路径：小库（≤30 表）退化为 1 簇 = 1 Worker，跳过分解直接下发',
  },
  worker: {
    title: 'Worker · ReAct Agent',
    role: '只管执行',
    dispatch: 'dispatch(cluster task) × N',
    prompt: {
      engine: 'Onboarding ReAct Engine',
      blocks: [
        { label: 'Mission', desc: '数据库类型与目标' },
        { label: 'Workflow', desc: '理解→采样→分布→质量→关系→写' },
        { label: 'Schema', desc: '注入本簇表/列/PK·FK/行数' },
        { label: 'Budget', desc: 'min/max 迭代预算' },
      ],
      rules: [
        '先 execute_sql 探查真实数据，再写 context —— 绝不臆测',
        'set_rich_context 用批量数组模式，一次写多条省迭代',
        '每表约 3 次迭代：探查 → 批量描述+同义词 → 样例值',
        'enum 类列（distinct < 20）必须记录样例值',
        'Sweep Check：收尾前核对每一列都已有 description',
        '不询问「是否继续」，处理完所有表才输出 Final Answer',
      ],
    },
    tools: [
      { name: 'execute_sql', desc: 'SELECT/SHOW/DESCRIBE 探查真实数据' },
      { name: 'set_rich_context', desc: '批量写入 Rich Context' },
    ],
    loop: 'Reason → Act → Observe 迭代循环',
    output: {
      label: 'Rich Context',
      store: 'rc_business_context',
      types: [
        { name: 'table_description', desc: '表的业务用途（2-3 句）' },
        { name: 'column_description', desc: '列的语义含义' },
        { name: 'column_sample_values', desc: 'text/enum 列的真实取值' },
        { name: 'column_synonyms', desc: '业务用户的自然语言别名' },
        { name: 'business_term', desc: '数据隐含的领域术语' },
      ],
    },
  },
  storage: {
    title: 'Embedding → Storage',
    items: [
      { table: 'rc_business_context', label: 'Rich Context', spec: '5 类 · is_expired 软失效', note: 'Worker 产出的语义上下文' },
      { table: 'rc_embeddings', label: '向量 Catalog', spec: 'VECTOR(2048) · HNSW · COSINE · is_deleted', note: 'Doubao 嵌入，VEC_FromText 写入，亚毫秒召回' },
    ],
  },
  insights: {
    input: '从 INFORMATION_SCHEMA 全量抽取表 / 列 / 外键，作为后续一切处理的骨架。',
    process: [
      { icon: 'i-lucide-git-fork', title: 'Forest 分簇控成本', body: '按外键连通性切簇，单 Worker 的上下文 / Token 随簇规模线性增长，而非全库平方膨胀；孤立表批量化减少 Agent 启停开销。' },
      { icon: 'i-lucide-search-check', title: 'Explore-before-write', body: 'Worker 先 execute_sql 看真实数据再写 context，从源头杜绝凭表名臆造元数据的幻觉。' },
      { icon: 'i-lucide-layers', title: '统一路径', body: '大库小库同一套 Coordinator → Worker；小库退化为单 Worker，无需另写分支逻辑。' },
    ],
    storage: [
      { icon: 'i-lucide-box', title: '原生向量零依赖', body: 'VECTOR(2048) + HNSW 直接落 MariaDB，省掉独立向量库；结构、语义、向量同库，强一致、易运维。' },
      { icon: 'i-lucide-recycle', title: '为演进留钩子', body: 'is_expired / is_deleted 软标记，为 Self-Maintenance 的失效标记与重嵌入预留接缝。' },
    ],
  },
}

const inferenceArch: InferenceArch = {
  id: 'inference',
  input: {
    label: '自然语言 Query',
    note: '用户问题 + datasource_id',
    example: '"上月 VIP 用户的订单总额"',
  },
  grounding: {
    title: 'Adaptive Grounding · Schema Linking',
    role: '定位相关表 / 列',
    dispatcher: {
      points: [
        'detectStrategy：按表数与阈值（30）选择策略',
        '≤30 小库：全量 schema 直接注入，免向量检索',
        '>30 大库：先向量召回候选，再交 Agent 精选',
      ],
      note: '大小库统一走 LinkAsync 并发架构：小库把 schema 预写入共享 slot，退化为「免检索」分支，无需另写代码路径',
    },
    retriever: {
      title: 'CoarseRetriever · 向量召回',
      desc: '对 rc_embeddings 做 HNSW 近邻检索，提取候选表',
      signals: [
        { name: 'TABLE', desc: '表实体名 → 命中相关表' },
        { name: 'COLUMN', desc: '列实体 → 回溯其所属表' },
        { name: 'CONTEXT', desc: '业务规则 / 术语语义' },
        { name: 'SQL_TEMPLATE', desc: '历史 SQL 模式匹配' },
      ],
    },
    agent: {
      title: 'LinkingAgent · LLM 精选',
      engine: 'LinkAsync',
      modes: [
        { name: 'off', desc: '跳过 LLM，直接用向量召回结果' },
        { name: 'one-shot', desc: 'LinkDirect 单次 LLM 调用（默认）' },
        { name: 'react', desc: '多步 ReAct + execute_sql 现场探查' },
      ],
      concurrency: 'react 大库模式：Agent 在 T0 即启动并与召回 goroutine 并发；召回完成把 schema 写入共享 slot，Agent 轮询 get_candidate_schema 取到后开始推理 —— 检索与推理重叠，端到端 ≈ max 而非相加',
    },
    output: {
      label: 'GroundedContext',
      parts: [
        'SelectedTables（含理由 / 置信度）',
        '相关列 + Rich Context（描述 / 样例 / 同义词）',
        'Relationships（外键 join 路径）',
      ],
    },
  },
  sqlgen: {
    title: 'SQL Generation · ReAct',
    role: '生成并自校验 SQL',
    prompt: {
      engine: 'Inference ReAct Engine',
      blocks: [
        { label: 'DB Type', desc: '目标库方言与语法约束' },
        { label: 'Rich Context', desc: '精选表的描述 / 样例 / 同义词 / FK' },
        { label: 'Best Practices', desc: 'NULL / 类型 / 极值 / 格式 等 9 条' },
        { label: 'Workflow', desc: '分析 → 探查 → 写 SQL → verify → Final' },
      ],
      rules: [
        'TEXT 存数字：比较 / 排序前 CAST，避免按字典序错排',
        'NULL = 未知 ≠ 0；过滤需兼顾 SQL NULL 与字符串 "null"',
        '字符串匹配优先用 Rich Context 精确取值；不确定先 execute_sql 探查',
        '可能产生重复时（join 一对多）用 DISTINCT',
        'Zero = 业务不存在，与 NULL（未知）含义不同',
        '极值用子查询 WHERE x=(SELECT MAX(x)…) 取全部并列，忌 LIMIT 1 漏并列',
        'Value Mapping：用值前先确认它在哪一列，不在相似列间臆测',
        '格式冲突（如 2 位 vs 4 位年份）：返回为空时换格式重试',
        '空格 / 特殊字符：WHERE 落空疑似格式问题，用 TRIM / LIKE 探查',
      ],
    },
    tools: [
      { name: 'inference_sql', desc: '执行 SQL 看结果（内置 DryRun 保护）' },
      { name: 'verify_sql', desc: 'EXPLAIN 校验语法 + 计划 + 性能告警' },
    ],
    loop: 'Reason → Act → Observe（claimed 5，实际 +3 缓冲）',
    verify: 'Final Answer 前必须 verify_sql ✅；失败则修正重试，绝不输出未通过校验的 SQL',
    output: { label: '通过校验的 SQL', note: 'Final Answer 即 verify_sql 通过的那条' },
  },
  execute: {
    title: 'Execute',
    role: '运行并返回结果',
    steps: [
      'adapter.ExecuteQuery 执行最终 SQL',
      '回收 rows / columns / 耗时',
    ],
    output: '结果集（rows · columns · execution_time）',
  },
  reads: {
    label: 'reads · Lakebase',
    items: [
      { table: 'rc_embeddings', use: '向量召回候选表（HNSW · COSINE）' },
      { table: 'rc_business_context', use: 'Rich Context 注入 SQL 生成提示' },
    ],
  },
  insights: {
    input: '入口只有一句自然语言 + 数据源，后续全靠 Grounding 把它对齐到具体表 / 列。',
    grounding: [
      { icon: 'i-lucide-layers', title: '统一并发架构', body: '大库小库共用 LinkAsync：小库预写 schema 退化为免检索分支，省去两套代码与维护成本。' },
      { icon: 'i-lucide-zap', title: '检索 ∥ 推理 重叠', body: 'react 模式下 Agent 与向量召回在 T0 并发启动，schema 经共享 slot 交付；端到端 ≈ max(检索, 推理) 而非两者相加。' },
      { icon: 'i-lucide-radar', title: '4 路信号召回', body: '表名 / 列 / 业务规则 / 历史 SQL 四路并发 HNSW，互补盲区，比单一向量召回覆盖更全。' },
    ],
    sqlgen: [
      { icon: 'i-lucide-shield-check', title: 'verify 守门', body: 'verify_sql 用 EXPLAIN 做语法 + 计划校验，未通过不许进 Final Answer，从源头拦截错误 SQL。' },
      { icon: 'i-lucide-database', title: '冲突时信任数据库', body: 'Rich Context 可能过期；与实测数据冲突时以 execute_sql 现场探查为准，而非盲信元数据。' },
    ],
    execute: [
      { icon: 'i-lucide-shield', title: 'DryRun 零副作用', body: 'inference_sql / 适配器内置 DryRun，探查与校验阶段只读，杜绝误写目标库。' },
    ],
  },
}

const maintainArch: MaintainArch = {
  id: 'maintain',
  trigger: {
    label: 'Schema Change Signal',
    note: 'MaintenanceSignal · SignalDDL（DDLStatements + Changes[]）',
    changeTypes: [
      { name: 'table_added / dropped', desc: 'CREATE / DROP TABLE' },
      { name: 'column_added / dropped', desc: 'ALTER TABLE … ADD / DROP COLUMN' },
      { name: 'column_modified', desc: 'MODIFY / CHANGE 列类型或定义' },
      { name: 'fk_added / dropped', desc: 'ADD / DROP FOREIGN KEY' },
      { name: 'index_added / dropped', desc: 'ADD / DROP INDEX / KEY' },
    ],
    detect: 'Evolution 演示：执行预定义 DDL → ParseDDLStatement 解析出变更类型 → SyncSchemaToLakebase 把 INFORMATION_SCHEMA upsert 进 catalog（仅同步物理结构，不覆盖语义 RC）。生产侧自动监测（cron / webhook）为预留接缝。',
  },
  coordinator: {
    title: 'Coordinator · ReAct Agent',
    role: '决策 + 失效标记',
    engine: 'Maintain Coordinator Engine · max 15 iter',
    points: [
      '逐条变更判定动作：新增 → create、改动 → refresh、删除 → delete',
      'inspect_schema_change 查现有 RC，确认是否真受影响（避免空转）',
      'mark_expired 把受影响表 / 列的 RC 置 is_expired（标表级联所有列）',
      'register_task 在内存累积任务，GetTasksJSON 一次性交给 Executor',
    ],
    tools: [
      { name: 'inspect_schema_change', desc: '按变更类型查 rc_tables / rc_columns 现有 RC' },
      { name: 'read_current_context', desc: '读单表 / 列的 description · 样例 · 同义词 · is_expired' },
      { name: 'mark_expired', desc: '标记表 / 列 RC 过期（标表时级联列）' },
      { name: 'register_task', desc: '内存累积 create / refresh / delete 任务' },
      { name: 'get_table_columns', desc: '列出某表全部列及其 RC 元数据' },
    ],
    budget: 'prompt max 15 / min 3（langchaingo 实际上限 ×3）',
    output: {
      label: 'Tasks JSON + 失效标记',
      parts: [
        'create / refresh / delete 任务清单',
        '受影响 RC 已置 is_expired（挂起待愈）',
      ],
    },
  },
  executor: {
    title: 'Executor · ReAct Agent',
    role: '只管自愈',
    engine: 'Maintain Executor Engine',
    dispatch: 'tasks JSON',
    steps: [
      'execute_sql 探查新列 / 改动列的真实数据分布',
      'set_rich_context 批量重写 description / 样例 / 同义词',
      'clear_expired 清除 is_expired —— 该 RC 自愈完成',
      'delete 任务 → delete_rich_context 删 RC 行 + 软删向量',
    ],
    tools: [
      { name: 'execute_sql', desc: '业务库只读探查（采样 / 分布），绝不臆测' },
      { name: 'set_rich_context', desc: '批量写 RC；写入即自动标 is_stale' },
      { name: 'delete_rich_context', desc: '删 rc_tables / rc_columns 行 + SoftDeleteEmbedding' },
      { name: 'clear_expired', desc: '清除 is_expired，标记该项自愈完成' },
    ],
    budget: 'taskCount × 5，clamp [10, 30]（实际上限 ×3）',
    note: 'Coordinator 判定无实质影响、未注册任务时，直接跳过 Executor —— 零 LLM 开销',
  },
  storage: {
    title: 'Invalidation & Re-embed · Lakebase',
    flags: [
      { name: 'is_expired', desc: 'rc_tables / rc_columns · RC 语义过期，待 refresh' },
      { name: 'is_stale', desc: 'rc_embeddings · RC 改动后向量过期，待重嵌入' },
      { name: 'is_deleted', desc: 'rc_embeddings · 实体删除软删，待 purge' },
    ],
    items: [
      { table: 'rc_tables · rc_columns', label: 'Rich Context 主存', spec: 'is_expired 软失效', note: 'mark_expired 挂起 / clear_expired 自愈' },
      { table: 'rc_embeddings', label: '向量 Catalog', spec: 'VECTOR(2048) · is_stale / is_deleted', note: '写时标 stale，收尾批量重算' },
      { table: 'rc_change_log', label: 'Change Log', spec: 'schema_change · trigger=agent', note: '每条变更落审计日志' },
    ],
    embed: '阶段收尾 GenerateAndSaveEmbeddings 全量重算向量（batch 100）并刷新 HNSW —— 当前为全量重嵌入而非 stale-only 增量，保证召回命中最新语义。',
  },
  insights: {
    trigger: 'Schema 一变，旧 Rich Context 即可能与现实不符；先用 DDL 解析 + catalog upsert 把「哪里变了」对齐出来，再交给 Agent 处理。',
    coordinator: [
      { icon: 'i-lucide-split', title: '决策与执行解耦', body: 'Coordinator 只判定 + 标记失效 + 注册任务，从不碰数据；判定逻辑全在 LLM + prompt 指南，无硬编码变更矩阵，新增变更类型零改码。' },
      { icon: 'i-lucide-flag', title: '失效即标记，不急删', body: 'mark_expired 把受影响 RC 置 is_expired（标表级联列），先「挂起」而非立即删，给 Executor 留出按需重生的窗口。' },
    ],
    executor: [
      { icon: 'i-lucide-search-check', title: '自愈也先探查', body: '与 Onboarding 一致：先 execute_sql 看真实数据再 set_rich_context，杜绝凭 DDL 文本臆测新列语义。' },
      { icon: 'i-lucide-skip-forward', title: '无任务即跳过', body: 'Coordinator 判定无实质影响时不注册任务，Executor 整段跳过，无谓的 LLM 调用与开销归零。' },
      { icon: 'i-lucide-gauge', title: '预算随任务缩放', body: 'Executor 迭代预算 = taskCount×5 并 clamp[10,30]：任务多给更多步、少则收紧，避免空转又不至于截断。' },
    ],
    storage: [
      { icon: 'i-lucide-layers', title: '三标志分层失效', body: 'is_expired 管 RC 语义、is_stale 管向量新鲜度、is_deleted 管软删；语义层与向量层各自独立失效与回收，互不阻塞。' },
      { icon: 'i-lucide-recycle', title: '写即标脏，收尾重嵌', body: 'set_rich_context 写入即 MarkEmbeddingStale；阶段末 GenerateAndSaveEmbeddings 统一重算向量，让检索召回的始终是最新语义。' },
    ],
  },
}

const kernelArch: KernelArch = {
  id: 'kernel',
  scenarios: {
    title: 'Scenario Config',
    role: '场景即配置',
    config: [
      { name: 'SystemPrompt', desc: '场景专属的任务 / 工作流 / 约束提示' },
      { name: 'Tools', desc: '本场景挂载的工具子集（最小权限）' },
      { name: 'Max/Min Iterations', desc: '写进 prompt 的迭代预算声明' },
      { name: 'ActualMaxOverride', desc: '真实硬上限（langchaingo 实际执行）' },
      { name: 'StepCallback', desc: '步骤回调 → SSE 流式推送' },
    ],
    list: [
      { name: 'onboarding', tools: 'execute_sql · set_rich_context', budget: 'tables×3+10' },
      { name: 'rc_gen', tools: 'execute_sql · set_rich_context', budget: '按簇' },
      { name: 'inference', tools: 'execute_sql · verify_sql', budget: 'claimed 5 / +3' },
      { name: 'schema_linking', tools: 'execute_sql', budget: 'actual 15' },
      { name: 'maintain_coordinator', tools: 'inspect · mark_expired · register_task · read · get_columns', budget: 'max 15 / min 3' },
      { name: 'maintain_executor', tools: 'execute_sql · set_rich_context · delete · clear_expired', budget: 'tasks×5 [10,30]' },
    ],
    note: 'grounding 的 LinkAsync 直接构造 EngineConfig（含 get_candidate_schema），是 scenarios 包外的第 7 种用法',
  },
  engine: {
    title: 'ReAct Engine',
    role: 'Reason → Act → Observe',
    base: 'langchaingo · agents.Initialize · ZeroShotReactDescription（ATLAS 薄封装，不自造循环）',
    loop: [
      { name: 'Reason', desc: 'LLM 生成 Thought：下一步该探查什么 / 写什么' },
      { name: 'Act', desc: '解析 Action + Action Input → tools.Tool.Call() 顺序执行' },
      { name: 'Observe', desc: '工具结果写回 scratchpad 作为 Observation，进入下一轮' },
    ],
    format: '纯文本 Thought / Action / Action Input / Observation（非 JSON function-call）—— 工具名取 Tool.Name()，Input 为 SQL 或 JSON 文本',
    parser: 'ParserErrorHandler：解析失败注入格式纠错提示让 LLM 重试，不终止循环',
    budget: '真实上限 = ActualMaxOverride 或 MaxIterations×3（且 ≥15）；prompt 里的 Max/Min 仅是「声称」',
  },
  transcript: [
    { kind: 'thought', text: '先确认 orders 表的 status 有哪些取值' },
    { kind: 'action', text: 'execute_sql' },
    { kind: 'input', text: 'SELECT DISTINCT status FROM orders LIMIT 20' },
    { kind: 'observation', text: '5 rows → paid, refunded, pending, shipped, cancelled' },
    { kind: 'thought', text: '取值已确认，写聚合 SQL 并送校验' },
    { kind: 'action', text: 'verify_sql' },
    { kind: 'input', text: 'SELECT customer_id, SUM(amount) … WHERE status=\'paid\'' },
    { kind: 'observation', text: 'PASS · EXPLAIN 计划正常，无全表扫描告警' },
    { kind: 'final', text: '校验通过，返回该 SQL' },
  ],
  toolGroups: [
    {
      name: 'SQL 执行 / 校验',
      icon: 'i-lucide-terminal',
      accent: 'blue',
      items: [
        { name: 'execute_sql', desc: '只读 SELECT/SHOW/DESCRIBE 探查（含推理版 inference_sql · DryRun）' },
        { name: 'verify_sql', desc: 'EXPLAIN 校验语法 + 计划 + 性能告警，返回 PASS/FAIL' },
      ],
    },
    {
      name: 'Rich Context 读写',
      icon: 'i-lucide-book-text',
      accent: 'emerald',
      items: [
        { name: 'set_rich_context', desc: '写 RC（单条或 JSON 数组 batch）' },
        { name: 'read_current_context', desc: '读表 / 列当前 RC' },
        { name: 'get_table_columns', desc: '列出某表全部列 + RC 元数据' },
      ],
    },
    {
      name: '失效 / 维护',
      icon: 'i-lucide-wrench',
      accent: 'amber',
      items: [
        { name: 'inspect_schema_change', desc: '查变更影响的实体及现有 RC' },
        { name: 'mark_expired', desc: '标记表 / 列 RC 过期（标表级联列）' },
        { name: 'clear_expired', desc: '刷新后清除 expired 标记' },
        { name: 'register_task', desc: 'Coordinator 注册 create/refresh/delete 任务' },
        { name: 'delete_rich_context', desc: '删 RC 行 + 软删向量' },
      ],
    },
    {
      name: '并发协同',
      icon: 'i-lucide-share-2',
      accent: 'violet',
      items: [
        { name: 'get_candidate_schema', desc: '从 atomic 共享槽取候选 schema（空则轮询等待召回）' },
      ],
    },
  ],
  llm: {
    provider: 'OpenAI 兼容 · langchaingo llms.Model（llm.CreateLLMByKey 注入 BaseURL/Token/Model）',
    encoding: '文本 ReAct',
    note: '不依赖模型的 function-calling 能力，任意 OpenAI 兼容端点可用',
  },
  output: {
    label: 'Result · Step[] · SSE',
    parts: ['Final Answer', 'Steps[]（Iteration/Thought/Action/Observation）', 'SSE: action / observation / finish', 'Iterations · Duration'],
  },
  insights: {
    scenarios: [
      { icon: 'i-lucide-layers', title: '同一内核、场景即配置', body: '6 个 Build*Engine 只产出 EngineConfig（prompt + 工具子集 + 预算 + 回调），复用同一条 ReAct 循环；新增场景零改内核，agentic 能力靠配置组合而非分叉代码。' },
      { icon: 'i-lucide-shield', title: '工具子集即权限边界', body: '每个场景只挂它需要的工具：onboarding 只能探查+写 RC，coordinator 只能标记+注册任务，executor 才能删——最小权限把 Agent 的能力约束在职责内。' },
    ],
    engine: [
      { icon: 'i-lucide-type', title: '文本 ReAct 而非 JSON tool-call', body: '用 langchaingo ZeroShotReact 的纯文本 Thought/Action/Observation 协议，不依赖模型 function-calling，任意 OpenAI 兼容模型都能驱动；代价是要靠 ParserErrorHandler 兜格式。' },
      { icon: 'i-lucide-life-buoy', title: 'ParserErrorHandler 韧性', body: 'LLM 偶尔输出跑格式时不崩溃：注入一段格式纠错说明让它重写，循环继续——把 LLM 的不确定性挡在引擎内部。' },
      { icon: 'i-lucide-radio', title: '旁路采集 → 实时 SSE', body: 'Handler 把 langchaingo 回调映射成 Step 并经 SSE 推送 action/observation/finish，前端能实时看到 Agent 的 Reason→Act→Observe，而非黑盒等结果。' },
    ],
    tools: [
      { icon: 'i-lucide-boxes', title: '12 个工具 = 能力积木', body: '执行/校验、RC 读写、失效维护、并发协同四类工具构成内核的全部「手脚」；三大流程都是这套积木的不同编排。' },
      { icon: 'i-lucide-git-merge', title: '唯一并发是检索∥推理', body: '内核本身单线程顺序执行工具；唯一的重叠发生在 grounding：召回 goroutine 把 schema 写入 atomic 槽，Agent 用 get_candidate_schema 轮询取用，让检索与推理时间重叠。' },
    ],
  },
}

const storageArch: StorageArch = {
  id: 'storage',
  root: {
    table: 'rc_datasources',
    label: '数据源注册表',
    note: 'MySQL / PG / SQLite / MariaDB 业务库登记（name 唯一 · status · last_sync_at）',
    cascade: '所有 rc_* 业务表 FK datasource_id → rc_datasources ON DELETE CASCADE，删源即级联清空',
  },
  tables: {
    title: 'Rich Context 表',
    role: 'MariaDB 12 · AutoMigrate',
    items: [
      { table: 'rc_tables', label: '表级 RC + 物理元数据', cols: 'description · row_count · source · confidence', flag: 'is_expired', accent: 'indigo' },
      { table: 'rc_columns', label: '列级 RC + schema', cols: 'data_type · sample_values · synonyms · value_mapping · pk/fk', flag: 'is_expired', accent: 'indigo' },
      { table: 'rc_business_context', label: '结构化业务上下文', cols: 'context_type(枚举) · content · version', flag: 'is_expired', accent: 'indigo' },
      { table: 'rc_relations', label: '表间关系（join 边）', cols: 'from/to table·column · relation_type', accent: 'blue' },
      { table: 'rc_terms', label: '业务术语词典', cols: 'term · definition · synonyms · FULLTEXT', accent: 'blue' },
    ],
  },
  er: {
    root: 'rc_datasources',
    edges: [
      { table: 'rc_tables', rel: '1-N · is_expired' },
      { table: 'rc_columns', rel: '1-N · is_expired' },
      { table: 'rc_business_context', rel: '1-N · is_expired' },
      { table: 'rc_relations', rel: '1-N · FK 图' },
      { table: 'rc_terms', rel: '1-N · FULLTEXT' },
      { table: 'rc_embeddings', rel: '1-N · VECTOR(2048)' },
      { table: 'rc_change_log', rel: '1-N · 审计' },
    ],
  },
  vector: {
    title: 'rc_embeddings · 向量层',
    role: 'VECTOR(2048) · HNSW',
    spec: [
      { name: 'embedding VECTOR(2048)', desc: 'MariaDB 12 原生向量列，VEC_FromText 写入' },
      { name: 'VECTOR INDEX … DISTANCE=COSINE', desc: 'HNSW 索引（idx_embedding_hnsw）' },
      { name: 'uq_entity', desc: 'UNIQUE(datasource_id, entity_type, entity_id) 去重' },
      { name: 'is_stale / is_deleted', desc: 'RC 改动待重嵌 / 软删待 purge' },
    ],
    ddl: `CREATE TABLE rc_embeddings (
  id INT AUTO_INCREMENT PRIMARY KEY,
  datasource_id INT NOT NULL,
  entity_type ENUM('table','column','term','query'),
  entity_id   INT NOT NULL,
  entity_text TEXT NOT NULL,
  embedding   VECTOR(2048) NOT NULL,
  is_stale    TINYINT(1) DEFAULT 0,
  is_deleted  TINYINT(1) DEFAULT 0,
  UNIQUE KEY uq_entity (datasource_id, entity_type, entity_id),
  VECTOR INDEX idx_embedding_hnsw (embedding) DISTANCE=COSINE
)`,
    search: `SELECT id, entity_type, entity_id, entity_text,
  VEC_DISTANCE_COSINE(embedding, VEC_FromText(?)) AS distance
FROM rc_embeddings IGNORE INDEX (idx_embedding_hnsw)
WHERE datasource_id = ? AND is_deleted = 0
ORDER BY distance ASC
LIMIT ?;   -- score = 1.0 - distance`,
    searchNote: '刻意 IGNORE INDEX：HNSW 是全局图，按 datasource_id 后过滤会召回不足；故对单数据源做 scoped 暴力扫描（数据量可控，精度优先）。',
  },
  embed: {
    provider: 'Doubao Embedding（OpenAI 兼容 / Volcengine Ark）',
    dim: '2048d',
    model: 'doubao-embedding',
    paths: [
      'on-write 异步：RC 写入触发 goroutine 单条 EmbedEntityByName',
      'batch catch-up：阶段末全量扫 tables/columns/terms，EmbedBatch（100/批）',
      '维护：写 RC 即 MarkEmbeddingStale，后续 UpsertEmbedding 把 is_stale=0',
    ],
    upsert: 'INSERT … VALUES(VEC_FromText(?)) ON DUPLICATE KEY UPDATE … is_stale=0',
  },
  changelog: {
    table: 'rc_change_log',
    types: [
      { name: 'schema_change', desc: 'DDL 检测（table/column/fk add·drop·modify）' },
      { name: 'context_expire', desc: '上下文被标记过期' },
      { name: 'context_update', desc: '上下文创建 / 更新 / 维护运行' },
    ],
    note: '记 change_detail(JSON) · old/new_value · trigger_source(agent/user/system) · change_reason，供自维护审计',
  },
  flags: [
    { name: 'is_expired', desc: 'rc_tables / rc_columns / rc_business_context · RC 语义过期待 refresh' },
    { name: 'is_stale', desc: 'rc_embeddings · RC 改动后向量过期待重嵌' },
    { name: 'is_deleted', desc: 'rc_embeddings · 实体删除软删待 purge' },
  ],
  insights: {
    root: '一切以数据源为根：所有 rc_* 表都 FK 到 rc_datasources 且 ON DELETE CASCADE，多数据源天然隔离，删源即一键清空其全部语义与向量。',
    tables: [
      { icon: 'i-lucide-database', title: '结构 / 语义 / 向量同库', body: 'MariaDB 12 一库装下物理 schema、Rich Context 与向量，省掉独立向量库；启动 AutoMigrate 对比嵌入 DDL 增量加列，强一致、易运维。' },
      { icon: 'i-lucide-link', title: '逻辑关联而非硬 FK', body: 'rc_business_context / rc_embeddings 用 (datasource_id, table_name[, column_name]) 或 (entity_type, entity_id) 关联，不与 rc_tables 建硬 FK——便于跨实体类型统一存储与去重。' },
    ],
    vector: [
      { icon: 'i-lucide-radar', title: '原生向量、零外部依赖', body: 'VECTOR(2048) + VEC_DISTANCE_COSINE 直接在 SQL 里算相似度，向量召回和结构查询同一条连接、同一事务，不必把数据同步到外部向量库。' },
      { icon: 'i-lucide-crosshair', title: 'scoped 暴力胜过全局 HNSW', body: 'HNSW 建了却 IGNORE INDEX：因为它是跨数据源的全局图，按 datasource_id 过滤后近邻会被「挤掉」导致召回不足；改对单源做暴力扫描，数据量可控下精度更稳。' },
      { icon: 'i-lucide-recycle', title: '三标志驱动重嵌', body: 'is_stale 标向量过期、is_deleted 标软删；写 RC 即标 stale，收尾 UpsertEmbedding 重算并归零——语义与向量的新鲜度各自独立管理。' },
    ],
  },
}

export const MODULES: Record<string, ModuleData> = {
  onboarding: { id: 'onboarding', accent: 'emerald', onboarding: onboardingArch },
  inference: { id: 'inference', accent: 'blue', inference: inferenceArch },
  maintain: { id: 'maintain', accent: 'amber', maintain: maintainArch },
  kernel: { id: 'kernel', accent: 'violet', kernel: kernelArch },
  storage: { id: 'storage', accent: 'indigo', storage: storageArch },
}

export function getModule(id: string | null): ModuleData | null {
  if (!id) return null
  return MODULES[id] ?? null
}
