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

export interface ModuleData {
  id: string
  accent: AccentKey
  onboarding?: OnboardingArch
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

export const MODULES: Record<string, ModuleData> = {
  onboarding: { id: 'onboarding', accent: 'emerald', onboarding: onboardingArch },
}

export function getModule(id: string | null): ModuleData | null {
  if (!id) return null
  return MODULES[id] ?? null
}
