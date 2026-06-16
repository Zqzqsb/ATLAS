// comm — generic Context-Layer Framework deck.
// Shape: a horizontal "system stages" pipeline (interaction → context → reason → SQL validate → human feedback → memory),
// each stage drilling into how DIFFERENT vendors realize it (variants) + our common-sense principles + sharp tradeoffs.

import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }

/** A "fake UI" fragment rendered inside the Interaction showcase TV.
 *  Each fragment is a self-contained <template> block keyed by `kind`,
 *  so the showcase can swap implementations without re-typing huge
 *  class strings. */
export type InteractionScenario =
  | {
      kind: 'chat'
      /** the user message bubble */
      userMsg: string
      /** the assistant's typed-out reply, with a highlighted SQL line + a "执行" button */
      assistantSql: string
      /** one-line clarification the bot asked (or null if it didn't need to) */
      clarification?: string
    }
  | {
      kind: 'mcp'
      /** function-calling toolbelt — agent picks tools to call */
      tools: { name: string; icon: string; status: 'idle' | 'calling' | 'done'; output?: string }[]
      /** one LLM reasoning line that prompted the tool calls */
      reasoning: string
    }
  | {
      kind: 'ide'
      /** Left pane: VSCode-style editor with an AI-assistant chat asking for the change. */
      left: {
        /** Editor file tab (path + language hint icon) */
        filePath: string
        /** Mock code lines rendered in the editor body. Mix of kept lines
         *  + new lines that the AI just wrote. Each line: { code, kind } where
         *  kind ∈ 'kept' | 'new' | 'new-active' | 'margin' */
        lines: { code: string; kind: 'kept' | 'new' | 'new-active' | 'margin' }[]
        /** Right-side AI chat panel inside the editor */
        chat: {
          /** user prompt in the chat */
          userPrompt: string
          /** AI reply (short, one or two lines) */
          aiReply: string
        }
      }
      /** Right pane: GitHub-style PR page. The diff is a *summary* of the
       *  editor's new lines + an old baseline. */
      right: {
        /** PR title (becomes the commit msg on merge) */
        title: string
        /** "+N −M" change badge */
        additions: number
        deletions: number
        /** CI status line under the title */
        ciLabel: string
        ciState: 'pass' | 'pending' | 'fail'
        /** Reviewer chips */
        reviewers: string[]
        /** Files-changed diff (just the conceptual patch; not full file) */
        patch: {
          filePath: string
          oldBlock: string[]
          newBlock: string[]
        }
      }
    }
  | {
      kind: 'bi'
      /** Left pane: a static BI dashboard (the pre-existing artifact the
       *  user is already looking at). */
      left: {
        title: string
        subtitle: string
        xLabel: string
        yLabel: string
        bars: { label: string; value: number; highlight?: boolean }[]
      }
      /** Right pane: an "Ask this chart" NL side panel that the BI surfaces
       *  inject next to an existing chart. Shows a user prompt, the SQL the
       *  Context Layer generated, and the resulting narrative answer. */
      right: {
        /** suggested NL actions shown above the input (chips) */
        suggestions: string[]
        /** the user's free-text question (typed in the input) */
        userPrompt: string
        /** the SQL the Context Layer produced to answer it */
        sql: string
        /** one-paragraph narrative answer the user reads */
        answer: string
      }
    }

/** Augment an ArchNode with a mini-UI scenario used by InteractionShowcase.
 *  (We actually add the optional field directly to ArchNode in
 *   features/arch/model/architecture.ts to avoid module-augmentation
 *   friction with TS path mapping.) */
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
        scenario: {
          kind: 'chat',
          userMsg: '上个月华东区复购用户的客单价趋势怎么样？',
          clarification: '"复购" 是按 30 天 / 90 天 / 终身?  还是说历史上下过 ≥2 单的客户?',
          assistantSql:
            'SELECT date_trunc(\'week\', o.ordered_at) AS wk,\n       avg(o.total_amount)              AS aov\nFROM   prod.orders o\nJOIN   prod.customers c USING (customer_id)\nWHERE  c.region = \'east-china\'\n  AND  c.is_repeat = true\nGROUP  BY 1 ORDER BY 1;',
        },
      },
      {
        id: 'agent-tool',
        label: 'Agent · MCP / SDK',
        sublabel: '面向外部 Agent · 工具调用 / function-calling',
        icon: 'i-lucide-bot',
        accent: 'slate',
        flow: 'ux',
        span: 1,
        scenario: {
          kind: 'mcp',
          reasoning: '"上个月客单价" 需要：① 找到 orders 表 ② 解析 "上个月" → 时间过滤 ③ 聚合 avg(total_amount)',
          tools: [
            { name: 'list_tables',         icon: 'i-lucide-database', status: 'done',   output: 'orders, customers, products' },
            { name: 'get_schema:orders',   icon: 'i-lucide-table-2',  status: 'done',   output: '21 columns' },
            { name: 'nl2sql',              icon: 'i-lucide-wand-2',   status: 'calling' },
            { name: 'dry_run_sql',         icon: 'i-lucide-flask-conical', status: 'idle' },
          ],
        },
      },
      {
        id: 'ide-cli',
        label: 'IDE / CLI',
        sublabel: '面向开发者 · 仓库化语义层 · git diff 评审',
        icon: 'i-lucide-terminal',
        accent: 'slate',
        flow: 'ux',
        span: 1,
        scenario: {
          kind: 'ide',
          left: {
            filePath: 'semantic/cubes/orders.yml',
            lines: [
              { code: '# semantic/cubes/orders.yml',                 kind: 'kept' },
              { code: 'model: orders',                               kind: 'kept' },
              { code: '',                                            kind: 'kept' },
              { code: 'measures:',                                   kind: 'kept' },
              { code: '  - name: total_revenue',                     kind: 'kept' },
              { code: '    expr: "SUM(orders.amount)"',              kind: 'kept' },
              { code: '  - name: order_count',                       kind: 'kept' },
              { code: '    expr: "COUNT(DISTINCT order_id)"',        kind: 'kept' },
              { code: '',                                            kind: 'margin' },
              { code: '  - name: aov',                               kind: 'new-active' },
              { code: '    expr: "SUM(amount) / NULLIF(COUNT(DISTINCT order_id), 0)"', kind: 'new-active' },
              { code: '',                                            kind: 'margin' },
              { code: '  - name: is_repeat',                         kind: 'new' },
              { code: '    expr: "COUNT(DISTINCT order_id) >= 2"',    kind: 'new' },
            ],
            chat: {
              userPrompt: '给 orders 加 aov + is_repeat 两个指标',
              aiReply: '已写入 cubes/orders.yml · 5 行 · commit 后推 PR',
            },
          },
          right: {
            title: 'feat(cubes): add aov + is_repeat derived measures',
            additions: 5,
            deletions: 0,
            ciLabel: '✓ Checks passed',
            ciState: 'pass',
            reviewers: ['alice', 'bob', 'carol'],
            patch: {
              filePath: 'semantic/cubes/orders.yml',
              oldBlock: [
                '  - name: order_count',
                '    expr: "COUNT(DISTINCT order_id)"',
              ],
              newBlock: [
                '  - name: aov',
                '    expr: "SUM(amount) / NULLIF(COUNT(DISTINCT order_id), 0)"',
                '',
                '  - name: is_repeat',
                '    expr: "COUNT(DISTINCT order_id) >= 2"',
              ],
            },
          },
        },
      },
      {
        id: 'bi-embed',
        label: 'BI / Notebook 嵌入',
        sublabel: '已有 BI / Notebook 拉数据 · 生成图表 / 报表',
        icon: 'i-lucide-bar-chart-3',
        accent: 'slate',
        flow: 'ux',
        span: 1,
        scenario: {
          kind: 'bi',
          left: {
            title: '上个月 · 华东区 · 复购客单价',
            subtitle: '已有 BI 看板（Looker / Tableau / Superset 等）',
            xLabel: '周',
            yLabel: 'AOV (¥)',
            bars: [
              { label: 'W1', value: 412 },
              { label: 'W2', value: 458 },
              { label: 'W3', value: 521, highlight: true },
              { label: 'W4', value: 487 },
            ],
          },
          right: {
            suggestions: [
              '解释 W3 飙升原因',
              '对比去年同周',
              '按品类拆开看',
            ],
            userPrompt: '为什么 W3 客单价比 W2 涨了 14%？',
            sql: 'SELECT reason_code, SUM(amount) AS s\nFROM   prod.orders o\nJOIN   prod.refund_reasons r USING (order_id)\nWHERE  o.region = \'east-china\'\n  AND  o.is_repeat = true\n  AND  DATE_TRUNC(\'week\', o.ordered_at) = DATE \'2026-05-12\'\nGROUP  BY 1 ORDER BY s DESC LIMIT 3;',
            answer: 'W3 客单价 521 比 W2 458 高 14% · 主要来自 ① 复购用户客单价基数本身较高 (¥638 vs ¥420) ② W3 的 5/14 "高端会员日" 拉动 ③ 退货率从 8% 降到 5%，分母缩水。',
          },
        },
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

/* ─── Code-ref system: chips that auto-link to a public github file ─── */

export type RepoKey = 'wrenai' | 'dbt-sl' | 'cube' | 'atlas' | 'ktx' | 'metricflow' | 'okf'

/** Public-repo registry — codebases we can deep-link into.
 *  Vendors that are closed-source (Snowflake, Databricks, Fabric, Oracle …)
 *  use the EvidenceChip / SourceCatalog system in arch/components/module/diagram instead. */
export const REPO_REGISTRY: Record<RepoKey, { label: string; base: string | null }> = {
  'wrenai':     { label: 'Canner/WrenAI',                 base: 'https://github.com/Canner/WrenAI/blob/main' },
  'dbt-sl':     { label: 'dbt-labs/dbt-semantic-interfaces', base: 'https://github.com/dbt-labs/dbt-semantic-interfaces/blob/main' },
  'metricflow': { label: 'dbt-labs/metricflow',           base: 'https://github.com/dbt-labs/metricflow/blob/main' },
  'cube':       { label: 'cube-js/cube',                  base: 'https://github.com/cube-js/cube/blob/master' },
  'atlas':      { label: 'ATLAS (internal)',              base: null },
  'ktx':        { label: 'Kaelio/ktx',                    base: 'https://github.com/Kaelio/ktx/blob/main' },
  'okf':        { label: 'GoogleCloudPlatform/knowledge-catalog', base: 'https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main' },
}

export interface CodeRef {
  repo: RepoKey
  /** path within the repo (e.g. `wren-ai-service/src/pipelines/generation/sql_generation.py`) */
  path: string
  /** optional line range — appended as `#L10-L20` */
  lines?: [number, number]
  /** override the displayed text (defaults to the file basename) */
  label?: string
}

/** One vendor's concrete take on a public step. */
export interface VendorTake {
  vendor: string
  school: VendorVariant['school']
  /** primary axis (WrenAI / Cortex Analyst by default) — rendered first, slightly bolder */
  primary?: boolean
  /** one-liner of how this vendor does THIS step */
  desc: string
  /** optional longer explanation. String = plain prose; object = structured
   *  summary + bullet list (rendered as a stylish "PeekPanel-style" expansion). */
  detail?: string | VendorDetail
  /** optional concrete example: code / YAML / SQL / shell snippet. Rendered as a code block when expanded. */
  example?: {
    /** language hint for syntax highlighting (yaml / sql / python / bash / json …) */
    lang?: string
    /** caption above the snippet */
    caption?: string
    /** the snippet body (newlines preserved) */
    code: string
  }
  /** white-box: github file refs (rendered as clickable chips) */
  code?: CodeRef[]
  /** black-box: external evidence ids (resolved against a SourceCatalog if provided) */
  refs?: string[]
  /** optional structured diagram (rendered above the example block when present) */
  diagram?: AdapterDiagram
  /** Explicit "this vendor does NOT do this step" marker — distinct from
   *  "we just haven't drilled down". When set, the card renders a muted
   *  "不形式化 / 不做" badge + the string as the gap explanation, no
   *  "尚未补充" hint. Falsy + missing detail/example = "drill-down can be
   *  added later". */
  notSupported?: string
  /** Explicit "one-liner is enough — no drill-down needed" marker. The
   *  card hides the gray fallback hint entirely; consumers know desc
   *  alone is the contract for this vendor's take. */
  selfContained?: boolean
}

/** Structured detail: a 1-line summary + bullet points + optional closing line. */
export interface VendorDetail {
  /** one-paragraph headline shown collapsed; clicking expands the bullets */
  summary: string
  /** ordered, expandable bullets */
  bullets: VendorBullet[]
  /** optional closing punchline (rendered emphasized below the bullets) */
  closing?: string
}

export interface VendorBullet {
  /** short label rendered bolder in front, e.g. "重命名列" / "选择性暴露" */
  label: string
  /** body of the bullet (one line preferred) */
  body: string
  /** optional small icon on the leading bullet marker */
  icon?: string
  /** accent color override (defaults to the active vendor's school accent) */
  accent?: AccentKey
}

/** "Adapter" diagram — renders a physical table on the left, a logical model on the
 *  right, and color-coded mapping lines in between.  Captures the four kinds of
 *  exposure: rename / expose / hidden (no mapping line) / computed / relation. */
export interface AdapterDiagram {
  kind: 'adapter'
  caption?: string
  physical: {
    label: string
    sublabel?: string
    columns: AdapterPhysicalCol[]
  }
  logical: {
    label: string
    sublabel?: string
    columns: AdapterLogicalCol[]
  }
}

export interface AdapterPhysicalCol {
  name: string
  type?: string
  /** when true, render struck-through with "敏感/隐藏" badge — no mapping line out */
  hidden?: boolean
  /** semantic-only flag for the "敏感" red label (still hidden=true for layout) */
  sensitive?: boolean
}

export interface AdapterLogicalCol {
  name: string
  /** how this column came to exist on the logical side */
  kind: 'rename' | 'expose' | 'computed' | 'relation'
  /** physical column name(s) it maps from (for rename/expose, single; computed/relation may be 0/many) */
  from?: string | string[]
  /** SQL expression (computed) or join hint (relation) — shown inline in monospace */
  expr?: string
  /** optional short note shown under the row */
  note?: string
}

/** A "common-sense" step in the stage. Vendor takes are listed under it. */
export interface Step {
  id: string
  /** short imperative label, e.g. `从 schema 起骨架` */
  name: string
  /** one-line description of the step */
  desc: string
  /** optional icon */
  icon?: string
  /** vendor-by-vendor takes; primaries appear first */
  takes: VendorTake[]
}

/** A sub-question within a stage: the single design choice that splits vendors apart.
 *  Two presentation modes:
 *  - `variants[]`  (legacy):  group vendors by school
 *  - `steps[]`     (preferred): public common-sense steps × per-vendor takes */
export interface SubQuestion {
  id: string
  question: string
  /** short framing of why this question matters */
  why: string
  variants?: { name: string; desc: string; vendors: string[]; accent: AccentKey }[]
  steps?: Step[]
  /** our common-sense / opinion. NOT the same as a vendor — this is the framework's stance. */
  commonSense: string
}

/** Build a github URL from a CodeRef. Returns null for closed/unknown repos. */
export function codeRefUrl(c: CodeRef): string | null {
  const base = REPO_REGISTRY[c.repo]?.base
  if (!base) return null
  const line = c.lines ? `#L${c.lines[0]}-L${c.lines[1]}` : ''
  return `${base}/${c.path}${line}`
}

/** Default chip text for a CodeRef (basename or last 2 path segments). */
export function codeRefLabel(c: CodeRef): string {
  if (c.label) return c.label
  const segs = c.path.split('/')
  return segs[segs.length - 1] ?? c.path
}

/** Axis classification: white-box (open codebase) vs black-box (managed/closed).
 *  Used by the comm StageDetail to split each step's vendor takes into two
 *  facing card decks (WrenAI vs Cortex Analyst as the two primaries). */
export type VendorAxis = 'white' | 'black'

export function vendorAxis(t: VendorTake): VendorAxis {
  return t.school === 'managed-cloud' ? 'black' : 'white'
}

/** Split + sort vendor takes into the two axis decks; primaries first. */
export function splitTakesByAxis(takes: VendorTake[]): { white: VendorTake[]; black: VendorTake[] } {
  const white: VendorTake[] = []
  const black: VendorTake[] = []
  for (const t of takes) {
    if (vendorAxis(t) === 'white') white.push(t)
    else black.push(t)
  }
  const sortPrimaryFirst = (arr: VendorTake[]) =>
    arr.sort((a, b) => Number(!!b.primary) - Number(!!a.primary))
  return { white: sortPrimaryFirst(white), black: sortPrimaryFirst(black) }
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
