// comm — generic Context-Layer Framework deck.
// Shape: a horizontal "system stages" pipeline (interaction → context → reason → SQL validate → human feedback → memory),
// each stage drilling into how DIFFERENT vendors realize it (variants) + our common-sense principles + sharp tradeoffs.

import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }
/* ─── L0: 通用 Context-Layer 构建框架 · 6 个核心环节按问答时间线串起来 ─── */
export const COMM_LAYERS: ArchLayer[] = [
  {
    id: 'ux',
    title: '① Interaction · 用户交互层',
    subtitle: '怎么把"业务问题"变成可处理输入：对话 / 半结构化查询 / 编辑器 IDE / Agent 工具调用',
    icon: 'i-lucide-message-square',
    accent: 'slate',
    cols: 4,
    nodes: [
      {
        id: 'chat-ui',
        label: 'Chat / 自然语言',
        sublabel: '面向业务人员 · 单轮 / 多轮 · 澄清回路',
        icon: 'i-lucide-messages-square',
        accent: 'slate',
        flow: 'ux',
        span: 1,
      },
      {
        id: 'agent-tool',
        label: 'Agent · MCP / SDK',
        sublabel: '面向外部 Agent · 工具调用 / function-calling',
        icon: 'i-lucide-bot',
        accent: 'slate',
        flow: 'ux',
        span: 1,
      },
      {
        id: 'ide-cli',
        label: 'IDE / CLI',
        sublabel: '面向开发者 · 仓库化语义层 · git diff 评审',
        icon: 'i-lucide-terminal',
        accent: 'slate',
        flow: 'ux',
        span: 1,
      },
      {
        id: 'bi-embed',
        label: 'BI / Notebook 嵌入',
        sublabel: '已有 BI / Notebook 拉数据 · 生成图表 / 报表',
        icon: 'i-lucide-bar-chart-3',
        accent: 'slate',
        flow: 'ux',
        span: 1,
      },
    ],
  },
  {
    id: 'context',
    title: '② Context Layer · 定义 / 构建 / 存储 / 召回',
    subtitle: '上下文是"可信数据"的载体：决定语义契约形态、构建路径、存储介质、检索策略',
    icon: 'i-lucide-database',
    accent: 'emerald',
    cols: 4,
    nodes: [
      {
        id: 'define',
        label: '定义 · 语义契约',
        sublabel: '语义层 / Wiki / 知识图谱 / 自然语言注解',
        icon: 'i-lucide-file-signature',
        accent: 'emerald',
        flow: 'context',
        span: 1,
      },
      {
        id: 'build',
        label: '构建 · 自动 vs 手工',
        sublabel: 'introspect / dbt 导入 / LLM 生成 / 人工评审',
        icon: 'i-lucide-hammer',
        accent: 'emerald',
        flow: 'context',
        span: 1,
      },
      {
        id: 'store',
        label: '存储 · Git / DB / 向量',
        sublabel: 'YAML 仓库 / 应用 metadata DB / 向量库 / FTS',
        icon: 'i-lucide-archive',
        accent: 'emerald',
        flow: 'context',
        span: 1,
      },
      {
        id: 'recall',
        label: '召回 · 检索策略',
        sublabel: '全量注入 / BM25 / 向量 / Hybrid · schema linking',
        icon: 'i-lucide-search',
        accent: 'emerald',
        flow: 'context',
        span: 1,
      },
    ],
  },
  {
    id: 'reason',
    title: '③ Reasoning · 推理过程的一般化拆解',
    subtitle: '"从问题到 SQL"的中间步骤怎么切：单次生成 vs 流水线 vs Agent 循环',
    icon: 'i-lucide-cpu',
    accent: 'violet',
    cols: 3,
    nodes: [
      {
        id: 'plan',
        label: 'Plan · 意图分解',
        sublabel: '问题改写 / 意图分类 / 任务拆分 / 多轮规划',
        icon: 'i-lucide-list-tree',
        accent: 'violet',
        flow: 'reason',
        span: 1,
      },
      {
        id: 'ground',
        label: 'Ground · schema 落地',
        sublabel: '把"客户"绑到具体表/列 · 解歧义 · 取值剖析',
        icon: 'i-lucide-target',
        accent: 'violet',
        flow: 'reason',
        span: 1,
      },
      {
        id: 'gen',
        label: 'Generate · SQL 生成',
        sublabel: '直接生成 / 模板引擎 / 语义层编译 / 自修复',
        icon: 'i-lucide-code-2',
        accent: 'violet',
        flow: 'reason',
        span: 1,
      },
    ],
  },
  {
    id: 'verify',
    title: '④ Verify · SQL 校验闭环',
    subtitle: '"看起来对" ≠ "可执行 ∧ 语义正确"：在执行前后建立多重闸门',
    icon: 'i-lucide-shield-check',
    accent: 'amber',
    cols: 4,
    nodes: [
      {
        id: 'static',
        label: '静态校验',
        sublabel: 'parse / qualify / type-check / 引用列存在',
        icon: 'i-lucide-scan-line',
        accent: 'amber',
        flow: 'verify',
        span: 1,
      },
      {
        id: 'policy',
        label: '策略校验',
        sublabel: 'read-only · RLAC/CLAC · denied funcs · row limit',
        icon: 'i-lucide-shield',
        accent: 'amber',
        flow: 'verify',
        span: 1,
      },
      {
        id: 'dryrun',
        label: '预执行 / Dry-run',
        sublabel: 'LIMIT 0 / 计划解释 · 验证语法 / 列 / 权限',
        icon: 'i-lucide-flask-conical',
        accent: 'amber',
        flow: 'verify',
        span: 1,
      },
      {
        id: 'semantic',
        label: '语义合理性',
        sublabel: 'fan-out / chasm trap · 计量单位 · NULL 语义 · 时间窗',
        icon: 'i-lucide-microscope',
        accent: 'amber',
        flow: 'verify',
        span: 1,
      },
    ],
  },
  {
    id: 'human',
    title: '⑤ Human-in-the-Loop · 反馈与校准',
    subtitle: '"自动建模 + 自学" 不是闭环——人是不可或缺的精度锚点',
    icon: 'i-lucide-users',
    accent: 'rose',
    cols: 4,
    nodes: [
      {
        id: 'review',
        label: '事前评审',
        sublabel: '语义契约 / 关系 / measure 走 git diff / Pull Request',
        icon: 'i-lucide-git-pull-request',
        accent: 'rose',
        flow: 'human',
        span: 1,
      },
      {
        id: 'thumbs',
        label: '事中点赞 / 修正',
        sublabel: '👍/👎 · 在线编辑 SQL · 标注错误样本',
        icon: 'i-lucide-thumbs-up',
        accent: 'rose',
        flow: 'human',
        span: 1,
      },
      {
        id: 'eval',
        label: '事后回归 · Eval',
        sublabel: 'NL→SQL 黄金集 · LLM-as-judge · 人工抽样',
        icon: 'i-lucide-clipboard-check',
        accent: 'rose',
        flow: 'human',
        span: 1,
      },
      {
        id: 'curate',
        label: '知识精炼',
        sublabel: '矛盾标注 / 同义词 / 冷热数据淘汰 · 半自动 grill',
        icon: 'i-lucide-sparkles',
        accent: 'rose',
        flow: 'human',
        span: 1,
      },
    ],
  },
  {
    id: 'memory',
    title: '⑥ Memory & Governance · 自学习 / 治理 / 可观测',
    subtitle: '使用即沉淀：成功对话 → 知识，失败对话 → 任务；同时保留审计、安全、解释能力',
    icon: 'i-lucide-recycle',
    accent: 'indigo',
    cols: 4,
    nodes: [
      {
        id: 'learn',
        label: '使用即沉淀',
        sublabel: '确认的 NL-SQL → query_history · 失败 → 任务卡',
        icon: 'i-lucide-history',
        accent: 'indigo',
        flow: 'memory',
        span: 1,
      },
      {
        id: 'lineage',
        label: 'Lineage / 解释',
        sublabel: '展开轨迹 / dry-plan · 用了哪些模型 / join / 列',
        icon: 'i-lucide-route',
        accent: 'indigo',
        flow: 'memory',
        span: 1,
      },
      {
        id: 'access',
        label: '权限 / 审计',
        sublabel: 'session property → RLAC/CLAC · 全 SQL 落审计日志',
        icon: 'i-lucide-shield-check',
        accent: 'indigo',
        flow: 'memory',
        span: 1,
      },
      {
        id: 'observe',
        label: '指标 / 漂移监控',
        sublabel: '准确率 / 延迟 / 成本 · schema 漂移 → RC 失效',
        icon: 'i-lucide-activity',
        accent: 'indigo',
        flow: 'memory',
        span: 1,
      },
    ],
  },
]

export interface CommFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const commFlows: CommFlowDef[] = [
  {
    id: 'ux',
    label: 'Interaction',
    title: '① 用户交互层 · 入口形态',
    subtitle: '同一个上下文层，可被自然语言 / Agent 工具 / IDE / BI 四类入口共享。每类入口对正确性、延迟、可控性的诉求差异极大；上下文层必须在最底下提供同一份契约。',
    icon: 'i-lucide-message-square',
    accent: 'slate',
  },
  {
    id: 'context',
    label: 'Context Layer',
    title: '② 上下文层 · 定义 / 构建 / 存储 / 召回',
    subtitle: '"Context Layer" 不是一个固定的东西——是 4 个独立可换的子决策的组合：以什么形态定义语义？怎么构建出来？存哪里？怎么取出来。各家差异主要在这 4 维。',
    icon: 'i-lucide-database',
    accent: 'emerald',
  },
  {
    id: 'reason',
    label: 'Reasoning',
    title: '③ 推理过程 · 一般化拆解',
    subtitle: '不论"端到端 LLM" 还是 "Agent 编排原语"，都能被切成 Plan → Ground → Generate 三段。各家把 LLM 放在哪几段、怎么循环，是核心差异。',
    icon: 'i-lucide-cpu',
    accent: 'violet',
  },
  {
    id: 'verify',
    label: 'SQL Verify',
    title: '④ SQL 校验 · 多重闸门',
    subtitle: 'SQL "看起来对"远远不够。把校验拆成静态 / 策略 / 预执行 / 语义合理性四层闸门，能在执行前堵住 95% 的错——并能在出错时给 Agent 结构化反馈。',
    icon: 'i-lucide-shield-check',
    accent: 'amber',
  },
  {
    id: 'human',
    label: 'Human Loop',
    title: '⑤ 人在回路 · 反馈与校准',
    subtitle: '完全自动建模 / 自学 = 容易跑偏。把"人参与"细化成事前 / 事中 / 事后 / 知识精炼四种形态，让人成为精度的锚点而非瓶颈。',
    icon: 'i-lucide-users',
    accent: 'rose',
  },
  {
    id: 'memory',
    label: 'Memory & Governance',
    title: '⑥ 记忆 · 治理 · 可观测',
    subtitle: '使用即沉淀（成功 → 知识 / 失败 → 任务）+ 解释（lineage / dry-plan）+ 权限（RLAC/CLAC）+ 漂移监控——四件套缺一不可。',
    icon: 'i-lucide-recycle',
    accent: 'indigo',
  },
]

export function getCommFlow(id: string | null): CommFlowDef | null {
  if (!id) return null
  return commFlows.find((f) => f.id === id) ?? null
}

/* ─── L1 module data — a generic framework drill-down shape ─── */

/** Vendor variants for a single sub-question. Keep prose short, columns aligned. */
export interface VendorVariant {
  vendor: string
  /** which "school" this vendor belongs to (drives the column color) */
  school: 'agentic' | 'semantic-layer' | 'managed-cloud' | 'open-context'
  desc: string
}

/** A sub-question within a stage: the single design choice that splits vendors apart. */
export interface SubQuestion {
  id: string
  question: string
  /** short framing of why this question matters */
  why: string
  /** one row per VARIANT (i.e. an "answer school"). Vendors get bucketed under variants. */
  variants: { name: string; desc: string; vendors: string[]; accent: AccentKey }[]
  /** our common-sense / opinion. NOT the same as a vendor — this is the framework's stance. */
  commonSense: string
}

/** A concrete stage drilled down: principles + ordered sub-questions + tradeoffs. */
export interface StageArch {
  id: string
  /** one-line abstract of the stage's role in the pipeline */
  abstract: string
  /** the irreducible commitments any system must make in this stage */
  principles: NamedItem[]
  /** the design choices that split vendors apart */
  subQuestions: SubQuestion[]
  /** sharp tradeoffs / pitfalls / our opinions */
  insights: Insight[]
  /** quick visual: a tiny "vendor matrix" preview (compact form, for the right column) */
  matrix?: {
    rows: { vendor: string; school: VendorVariant['school']; cells: string[] }[]
    cols: string[]
  }
}

export const SCHOOL_META: Record<VendorVariant['school'], { label: string; accent: AccentKey; desc: string }> = {
  'agentic': {
    label: 'Agentic',
    accent: 'violet',
    desc: 'Agent 内化所有步骤（ATLAS / Fabric Data Agent / DMS Data Agent）',
  },
  'semantic-layer': {
    label: 'Semantic Layer',
    accent: 'amber',
    desc: '把契约抽到语义层、Agent 外置（dbt SL / Cube / WrenAI）',
  },
  'managed-cloud': {
    label: 'Managed Cloud',
    accent: 'blue',
    desc: '云厂商把 SL + LLM 内置进数据平台（Snowflake Cortex / Databricks UC / Oracle）',
  },
  'open-context': {
    label: 'Open Context',
    accent: 'emerald',
    desc: '开源 / 本地 · git 化 · BYO Agent（ktx / 类似工具）',
  },
}

import { uxArch } from './ux-arch'
import { contextArch } from './context-arch'
import { reasonArch } from './reason-arch'
import { verifyArch } from './verify-arch'
import { humanArch } from './human-arch'
import { memoryArch } from './memory-arch'

export const COMM_STAGES: Record<string, StageArch> = {
  ux: uxArch,
  context: contextArch,
  reason: reasonArch,
  verify: verifyArch,
  human: humanArch,
  memory: memoryArch,
}

export function getCommStage(id: string | null): StageArch | null {
  if (!id) return null
  return COMM_STAGES[id] ?? null
}
