/**
 * Per-module "internal detail" data — the deepest (non-expandable) level.
 *
 * Each module is described by a set of typed sections that reusable section
 * components render. To add a new module's detail: author a ModuleData entry
 * here + a <XxxDetail> composition component, then register it in ModuleDetail.
 */
import type { AccentKey } from './architecture'

/* ─── Strategy / dispatch (e.g. how small vs large databases are handled) ─── */
export interface StrategyOption {
  id: string
  label: string
  when: string
  icon: string
  accent: AccentKey
  points: string[]
}
export interface StrategySection {
  title: string
  subtitle: string
  decision: string
  options: StrategyOption[]
}

/* ─── Prompt engineering recipe ─── */
export interface PromptBlock {
  icon: string
  label: string
  desc: string
}
export interface PromptSection {
  title: string
  subtitle: string
  engine: string
  tools: string[]
  blocks: PromptBlock[]
  rules: string[]
}

/* ─── Storage layout ─── */
export type StorageKind = 'schema' | 'context' | 'catalog' | 'log'
export interface StorageItem {
  table: string
  label: string
  kind: StorageKind
  spec?: string
  note: string
}
export interface StorageSection {
  title: string
  subtitle: string
  items: StorageItem[]
}

/* ─── Design insights ─── */
export interface Insight {
  icon: string
  title: string
  body: string
}
export interface InsightSection {
  title: string
  subtitle: string
  items: Insight[]
}

export interface ModuleData {
  /** matches FlowDef.id in flows.ts */
  id: string
  strategy: StrategySection
  prompt: PromptSection
  storage: StorageSection
  insights: InsightSection
}

export const STORAGE_KIND_META: Record<StorageKind, { label: string; icon: string; accent: AccentKey }> = {
  schema: { label: '基础物理 Schema', icon: 'i-lucide-table-2', accent: 'indigo' },
  context: { label: '语义上下文层', icon: 'i-lucide-book-text', accent: 'emerald' },
  catalog: { label: '高维向量 Catalog', icon: 'i-lucide-radar', accent: 'violet' },
  log: { label: '审计日志', icon: 'i-lucide-scroll-text', accent: 'slate' },
}

const onboarding: ModuleData = {
  id: 'onboarding',
  strategy: {
    title: '任务注册与分发',
    subtitle: '同一套 ReAct 内核，按库规模自动切换编排策略',
    decision: '阈值：表数量 > 30 触发 Forest 分簇',
    options: [
      {
        id: 'small',
        label: 'Small · 单 Agent',
        when: '≤ 30 张表',
        icon: 'i-lucide-circle-dot',
        accent: 'emerald',
        points: [
          '整库 schema 一次性注入单个 Onboarding ReAct Agent',
          '迭代预算 ComputeChunkBudget = tables×3 + 10（上限 60）',
          '无需图分解，流程最短、上下文完整',
        ],
      },
      {
        id: 'large',
        label: 'Large · Forest 分簇',
        when: '> 30 张表',
        icon: 'i-lucide-git-fork',
        accent: 'blue',
        points: [
          'ForestDecompose：外键无向图 → BFS 求连通分量（表簇）',
          'MergeIsolatedTables：无 FK 的孤立表按 15 张/批合并',
          '逐簇构建独立 Agent，各自预算（上限随规模 60→500）',
          '单簇失败不阻塞整体（continue-on-error），可并行扩展',
        ],
      },
    ],
  },
  prompt: {
    title: 'Prompt Engineering',
    subtitle: 'Onboarding 专用 system prompt 如何被构造与约束',
    engine: 'Onboarding ReAct Engine',
    tools: ['execute_sql', 'set_rich_context'],
    blocks: [
      { icon: 'i-lucide-target', label: 'Mission', desc: '告知数据库类型与目标：探查并产出 Rich Context' },
      { icon: 'i-lucide-list-checks', label: 'Workflow', desc: '理解 schema → 采样 → 取值分布 → 数据质量 → 关系 → 写 context' },
      { icon: 'i-lucide-tags', label: 'RC Types', desc: '5 类 Rich Context 的 JSON schema 与示例' },
      { icon: 'i-lucide-database', label: 'Schema Overview', desc: '注入表/列/类型/PK·FK·nullable/行数 + 外键关系' },
      { icon: 'i-lucide-gauge', label: 'Iteration Budget', desc: '注入 min/max 迭代预算，引导覆盖全部表' },
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
  storage: {
    title: '落库结构',
    subtitle: '从基础物理 schema 到高维向量 catalog 的三层沉淀',
    items: [
      { table: 'rc_tables', label: '表元数据', kind: 'schema', spec: 'name · row_count · description', note: 'INFORMATION_SCHEMA 同步的物理表清单' },
      { table: 'rc_columns', label: '列元数据', kind: 'schema', spec: 'type · PK/FK · nullable · sample · synonyms', note: '列结构 + Agent 回填的语义字段' },
      { table: 'rc_relations', label: '外键关系', kind: 'schema', spec: 'from.col → to.col', note: 'Forest 分解的图来源' },
      { table: 'rc_business_context', label: 'Rich Context', kind: 'context', spec: '5 类 · is_expired 软失效', note: '表述/列述/样例值/同义词/业务术语' },
      { table: 'rc_embeddings', label: '向量 Catalog', kind: 'catalog', spec: 'VECTOR(2048) · HNSW · COSINE · is_deleted', note: 'Doubao 嵌入，VEC_FromText 写入，亚毫秒召回' },
      { table: 'rc_change_log', label: '变更日志', kind: 'log', spec: 'change_type · reason · trigger', note: 'Onboarding 完成审计，演进可追溯' },
    ],
  },
  insights: {
    title: '关键设计 Insight',
    subtitle: '为什么这样设计',
    items: [
      { icon: 'i-lucide-search-check', title: 'Explore-before-write', body: 'Agent 必须先跑 SQL 看真实数据再落 context，从源头杜绝 LLM 凭表名臆造元数据的幻觉。' },
      { icon: 'i-lucide-git-fork', title: 'Forest 分簇控成本', body: '按外键连通性把大库切成表簇，单 Agent 的上下文与 token 成本随簇规模线性增长，而非全库平方膨胀；孤立表批量化减少 Agent 启停开销。' },
      { icon: 'i-lucide-box', title: '原生向量，零外部依赖', body: '2048 维向量与 HNSW 索引直接落在 MariaDB，省掉独立向量库；结构、语义、向量同库，强一致、易运维。' },
      { icon: 'i-lucide-recycle', title: '为演进预留钩子', body: 'rc_business_context.is_expired 与 rc_embeddings.is_deleted 软标记，为后续 Self-Maintenance 的失效标记与重嵌入留好接缝。' },
      { icon: 'i-lucide-layers', title: 'Schema 是骨架，Context 是血肉', body: '基础 schema 提供结构，Rich Context 注入业务语义，二者共同支撑 Text-to-SQL 的精准 grounding。' },
    ],
  },
}

export const MODULES: Record<string, ModuleData> = { onboarding }

export function getModule(id: string | null): ModuleData | null {
  if (!id) return null
  return MODULES[id] ?? null
}
