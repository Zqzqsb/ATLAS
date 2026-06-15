/**
 * Architecture model — single source of truth for the panoramic (L0) view.
 * Layers / nodes / accents are all data-driven so the diagram can be reshaped
 * by editing this file alone, without touching components.
 */

/* ─── Accent presets ───
 * Full literal class strings (NOT interpolated) so UnoCSS can statically detect
 * and generate them. Add a new color here to make it available everywhere.
 */
export interface Accent {
  /** thin top bar on a layer/node */
  bar: string
  /** small status dot */
  dot: string
  /** soft tinted surface for a layer container */
  surface: string
  /** icon chip background + foreground */
  iconBg: string
  iconText: string
  /** hover affordance for a drillable node */
  hover: string
  /** label/text accent */
  text: string
  /** pill / chip */
  chip: string
  /** gradient for emphasis (e.g. stepper active icon) */
  gradient: string
}

export const ACCENTS = {
  slate: {
    bar: 'bg-slate-400',
    dot: 'bg-slate-400',
    surface: 'bg-slate-50/60 border-slate-200/70',
    iconBg: 'bg-slate-100',
    iconText: 'text-slate-600',
    hover: 'hover:border-slate-300 hover:shadow-slate-100',
    text: 'text-slate-600',
    chip: 'bg-slate-50 text-slate-600 border-slate-200',
    gradient: 'from-slate-500 to-gray-500',
  },
  emerald: {
    bar: 'bg-emerald-500',
    dot: 'bg-emerald-500',
    surface: 'bg-emerald-50/40 border-emerald-200/60',
    iconBg: 'bg-emerald-100',
    iconText: 'text-emerald-600',
    hover: 'hover:border-emerald-300 hover:shadow-emerald-100',
    text: 'text-emerald-600',
    chip: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    gradient: 'from-emerald-500 to-teal-500',
  },
  blue: {
    bar: 'bg-blue-500',
    dot: 'bg-blue-500',
    surface: 'bg-blue-50/40 border-blue-200/60',
    iconBg: 'bg-blue-100',
    iconText: 'text-blue-600',
    hover: 'hover:border-blue-300 hover:shadow-blue-100',
    text: 'text-blue-600',
    chip: 'bg-blue-50 text-blue-700 border-blue-200',
    gradient: 'from-blue-500 to-cyan-500',
  },
  amber: {
    bar: 'bg-amber-500',
    dot: 'bg-amber-500',
    surface: 'bg-amber-50/40 border-amber-200/60',
    iconBg: 'bg-amber-100',
    iconText: 'text-amber-600',
    hover: 'hover:border-amber-300 hover:shadow-amber-100',
    text: 'text-amber-600',
    chip: 'bg-amber-50 text-amber-700 border-amber-200',
    gradient: 'from-amber-500 to-orange-500',
  },
  violet: {
    bar: 'bg-violet-500',
    dot: 'bg-violet-500',
    surface: 'bg-violet-50/40 border-violet-200/60',
    iconBg: 'bg-violet-100',
    iconText: 'text-violet-600',
    hover: 'hover:border-violet-300 hover:shadow-violet-100',
    text: 'text-violet-600',
    chip: 'bg-violet-50 text-violet-700 border-violet-200',
    gradient: 'from-violet-500 to-purple-500',
  },
  indigo: {
    bar: 'bg-indigo-500',
    dot: 'bg-indigo-500',
    surface: 'bg-indigo-50/40 border-indigo-200/60',
    iconBg: 'bg-indigo-100',
    iconText: 'text-indigo-600',
    hover: 'hover:border-indigo-300 hover:shadow-indigo-100',
    text: 'text-indigo-600',
    chip: 'bg-indigo-50 text-indigo-700 border-indigo-200',
    gradient: 'from-indigo-500 to-blue-500',
  },
  rose: {
    bar: 'bg-rose-500',
    dot: 'bg-rose-500',
    surface: 'bg-rose-50/40 border-rose-200/60',
    iconBg: 'bg-rose-100',
    iconText: 'text-rose-600',
    hover: 'hover:border-rose-300 hover:shadow-rose-100',
    text: 'text-rose-600',
    chip: 'bg-rose-50 text-rose-700 border-rose-200',
    gradient: 'from-rose-500 to-pink-500',
  },
} satisfies Record<string, Accent>

export type AccentKey = keyof typeof ACCENTS

/* ─── Architecture nodes & layers ─── */
export interface ArchNode {
  id: string
  label: string
  sublabel?: string
  icon: string
  accent: AccentKey
  /** id of the dataflow in flows.ts; if present the node is drillable */
  flow?: string
  /** grid column span hint (out of the layer's column count) */
  span?: number
  /** backend source files this node maps to (for future "view code") */
  codeRefs?: string[]
  /** evidence source IDs (used by black-box vendor decks like Databricks/Snowflake) */
  refs?: string[]
}

export interface ArchLayer {
  id: string
  title: string
  subtitle?: string
  icon: string
  accent: AccentKey
  /** number of grid columns the layer renders */
  cols: number
  nodes: ArchNode[]
}

export const ARCH_LAYERS: ArchLayer[] = [
  {
    id: 'interface',
    title: 'Interface',
    subtitle: '用户接入与 API 网关',
    icon: 'i-lucide-monitor',
    accent: 'slate',
    cols: 2,
    nodes: [
      {
        id: 'web-console',
        label: 'Web Console',
        sublabel: 'Vue 3 · 工作台 / 上下文 / 演进面板',
        icon: 'i-lucide-layout-dashboard',
        accent: 'slate',
        span: 1,
      },
      {
        id: 'api-sse',
        label: 'REST API + SSE',
        sublabel: 'server/handlers · 流式进度推送',
        icon: 'i-lucide-radio',
        accent: 'slate',
        span: 1,
        codeRefs: ['backend/server/handlers/sse.go'],
      },
    ],
  },
  {
    id: 'pipelines',
    title: 'Pipelines',
    subtitle: '三大核心流程',
    icon: 'i-lucide-workflow',
    accent: 'violet',
    cols: 3,
    nodes: [
      {
        id: 'onboarding',
        label: 'Onboarding',
        sublabel: '接入新库 · 自动生成 Rich Context',
        icon: 'i-lucide-database-zap',
        accent: 'emerald',
        flow: 'onboarding',
        span: 1,
        codeRefs: [
          'backend/internal/react/scenarios/onboarding.go',
          'backend/internal/react/scenarios/rc_gen.go',
        ],
      },
      {
        id: 'inference',
        label: 'Inference',
        sublabel: 'Schema Linking → SQL → 校验执行',
        icon: 'i-lucide-git-graph',
        accent: 'blue',
        flow: 'inference',
        span: 1,
        codeRefs: [
          'backend/internal/inference/pipeline.go',
          'backend/internal/grounding/adaptive_pipeline.go',
        ],
      },
      {
        id: 'maintain',
        label: 'Self-Maintenance',
        sublabel: 'Signal → 失效标记 → 自愈重嵌',
        icon: 'i-lucide-bot',
        accent: 'amber',
        flow: 'maintain',
        span: 1,
        codeRefs: [
          'backend/internal/agent/agent_service.go',
          'backend/internal/react/scenarios/maintain_coordinator.go',
          'backend/internal/react/scenarios/maintain_executor.go',
        ],
      },
    ],
  },
  {
    id: 'kernel',
    title: 'ReAct Kernel',
    subtitle: '三大流程共用的推理内核',
    icon: 'i-lucide-cpu',
    accent: 'violet',
    cols: 2,
    nodes: [
      {
        id: 'react-engine',
        label: 'ReAct Engine',
        sublabel: 'Reason → Act → Observe 迭代循环',
        icon: 'i-lucide-repeat',
        accent: 'violet',
        flow: 'kernel',
        span: 1,
        codeRefs: ['backend/internal/react/engine.go'],
      },
      {
        id: 'tool-belt',
        label: 'Tool Belt',
        sublabel: 'execute_sql · set_rich_context · verify_sql …',
        icon: 'i-lucide-wrench',
        accent: 'violet',
        flow: 'kernel',
        span: 1,
        codeRefs: ['backend/internal/react/tools'],
      },
    ],
  },
  {
    id: 'storage',
    title: 'Lakebase Storage',
    subtitle: 'MariaDB 12 · 原生 VECTOR + HNSW',
    icon: 'i-lucide-database',
    accent: 'indigo',
    cols: 4,
    nodes: [
      { id: 'rc-schema', label: 'Schema Metadata', sublabel: 'rc_tables / rc_columns / rc_relations', icon: 'i-lucide-table-2', accent: 'indigo', flow: 'storage', span: 1 },
      { id: 'rc-context', label: 'Rich Context', sublabel: 'rc_business_context', icon: 'i-lucide-book-text', accent: 'indigo', flow: 'storage', span: 1 },
      { id: 'rc-embed', label: 'Vector Embeddings', sublabel: 'rc_embeddings · VECTOR(2048)', icon: 'i-lucide-radar', accent: 'indigo', flow: 'storage', span: 1 },
      { id: 'rc-log', label: 'Change Log', sublabel: 'rc_change_log', icon: 'i-lucide-scroll-text', accent: 'indigo', flow: 'storage', span: 1 },
    ],
  },
  {
    id: 'foundation',
    title: 'Foundation',
    subtitle: '底层依赖与外部资源',
    icon: 'i-lucide-layers',
    accent: 'slate',
    cols: 3,
    nodes: [
      { id: 'llm', label: 'LLM Provider', sublabel: 'OpenAI 兼容 · 推理 / 生成', icon: 'i-lucide-brain', accent: 'slate', span: 1 },
      { id: 'embedding', label: 'Embedding', sublabel: 'Doubao Embedding · 2048d', icon: 'i-lucide-spline', accent: 'slate', span: 1 },
      { id: 'adapter', label: 'Target DB Adapter', sublabel: 'MySQL · DryRun 安全执行', icon: 'i-lucide-plug', accent: 'slate', span: 1 },
    ],
  },
]
