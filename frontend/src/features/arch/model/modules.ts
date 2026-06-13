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

export interface ModuleData {
  id: string
  accent: AccentKey
  onboarding?: OnboardingArch
  inference?: InferenceArch
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

export const MODULES: Record<string, ModuleData> = {
  onboarding: { id: 'onboarding', accent: 'emerald', onboarding: onboardingArch },
  inference: { id: 'inference', accent: 'blue', inference: inferenceArch },
}

export function getModule(id: string | null): ModuleData | null {
  if (!id) return null
  return MODULES[id] ?? null
}
