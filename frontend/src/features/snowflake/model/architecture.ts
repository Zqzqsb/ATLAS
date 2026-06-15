/**
 * Snowflake deck — L0 panorama (Cortex Analyst + Semantic Views + VQR + Cortex Search).
 *
 * Black-box product → every node carries `refs: ['Sx']` pointing into
 * `sources.ts` (mirroring the WiseCat yaml evaluation evidence). Drillable
 * nodes set `flow: <id>` matching a `SnowFlowDef.id`.
 */
import { ACCENTS, type Accent, type AccentKey, type ArchLayer, type ArchNode } from '../../arch/model/architecture'
import type { Insight, NamedItem } from '../../arch/model/modules'

export { ACCENTS }
export type { Accent, AccentKey, ArchLayer, ArchNode, Insight, NamedItem }

export const SNOW_LAYERS: ArchLayer[] = [
  {
    id: 'agent-surface',
    title: 'Agent / Query Surface',
    subtitle: 'Cortex Analyst REST API + Streamlit demo + Snowsight Suggestions UI · Snowflake 不内置执行',
    icon: 'i-lucide-radio',
    accent: 'slate',
    cols: 3,
    nodes: [
      {
        id: 'analyst-rest',
        label: 'Cortex Analyst REST',
        sublabel: '/api/v2/cortex/analyst · 仅生成 SQL，不执行 · feedback endpoint',
        icon: 'i-lucide-radio',
        accent: 'slate',
        flow: 'analyst-flow',
        span: 1,
        refs: ['S1', 'S8'],
      },
      {
        id: 'streamlit',
        label: 'Streamlit / Verify UI',
        sublabel: '示例 UI · verified query 交互验证后保存',
        icon: 'i-lucide-monitor',
        accent: 'slate',
        span: 1,
        refs: ['S5'],
      },
      {
        id: 'snowsight-monitor',
        label: 'Snowsight Monitoring',
        sublabel: '管理员看用户问题 / SQL / warnings · 反馈源',
        icon: 'i-lucide-activity',
        accent: 'slate',
        span: 1,
        refs: ['S16'],
      },
    ],
  },
  {
    id: 'semantics',
    title: 'Semantic Layer',
    subtitle: '两种形态：原生 Semantic View（DDL 对象，推荐）+ stage 上 YAML semantic_model（向后兼容）',
    icon: 'i-lucide-shapes',
    accent: 'amber',
    cols: 2,
    nodes: [
      {
        id: 'semantic-view',
        label: 'Semantic View (DDL)',
        sublabel: 'CREATE SEMANTIC VIEW · TABLES / RELATIONSHIPS / FACTS / DIMENSIONS / METRICS',
        icon: 'i-lucide-shapes',
        accent: 'amber',
        flow: 'semantic-view',
        span: 1,
        refs: ['S2', 'S3', 'S4'],
      },
      {
        id: 'yaml-model',
        label: 'Stage YAML semantic_model',
        sublabel: 'legacy · 文件存 stage 或字符串注入 · 仍向后兼容',
        icon: 'i-lucide-file-text',
        accent: 'amber',
        flow: 'semantic-view',
        span: 1,
        refs: ['S1', 'S17'],
      },
    ],
  },
  {
    id: 'rich-context',
    title: 'Rich Context',
    subtitle: 'verified queries · custom instructions · synonyms · sample_values（语义模型富化）',
    icon: 'i-lucide-list-checks',
    accent: 'violet',
    cols: 3,
    nodes: [
      {
        id: 'vqr',
        label: 'Verified Query Repository',
        sublabel: 'NL-SQL 对 + verified_by/verified_at · 命中即 verified answer',
        icon: 'i-lucide-shield-check',
        accent: 'violet',
        flow: 'vqr',
        span: 1,
        refs: ['S5', 'S8'],
      },
      {
        id: 'custom-instr',
        label: 'Custom Instructions',
        sublabel: '自然语言规则约束 SQL 形状（默认时间过滤器、过滤传播等）',
        icon: 'i-lucide-pencil',
        accent: 'violet',
        flow: 'vqr',
        span: 1,
        refs: ['S14'],
      },
      {
        id: 'autopilot',
        label: 'Autopilot · Suggestions',
        sublabel: 'AI 辅助生成 / 候选建议 · 全部需人工审核',
        icon: 'i-lucide-sparkles',
        accent: 'violet',
        flow: 'autopilot',
        span: 1,
        refs: ['S11', 'S12'],
      },
    ],
  },
  {
    id: 'retrieval',
    title: 'Retrieval Layer',
    subtitle: '语义模型本体全量注入；高基数 dimension 接 Cortex Search 做 literal 召回',
    icon: 'i-lucide-search',
    accent: 'blue',
    cols: 2,
    nodes: [
      {
        id: 'cortex-search',
        label: 'Cortex Search Service',
        sublabel: '向量检索 + 关键词 + rerank（混合）· 服务于 literal / VQR 召回',
        icon: 'i-lucide-search',
        accent: 'blue',
        flow: 'cortex-search',
        span: 1,
        refs: ['S6', 'S7'],
      },
      {
        id: 'static-inject',
        label: 'Static Full Context',
        sublabel: '语义模型本体不做 schema 裁剪，整体注入 · 富上下文按需召回',
        icon: 'i-lucide-square-stack',
        accent: 'blue',
        span: 1,
        refs: ['S1', 'S9'],
      },
    ],
  },
  {
    id: 'runtime',
    title: 'Query Runtime',
    subtitle: '调用方在自有 warehouse 用自身角色执行 · masking / row access policy 在底表强制',
    icon: 'i-lucide-cpu',
    accent: 'emerald',
    cols: 3,
    nodes: [
      {
        id: 'user-warehouse',
        label: 'User Warehouse',
        sublabel: '调用方执行 SQL · Cortex Analyst 不执行 · 受 RBAC + 行列策略约束',
        icon: 'i-lucide-warehouse',
        accent: 'emerald',
        span: 1,
        refs: ['S1', 'S8'],
      },
      {
        id: 'sql-compiler',
        label: 'SQL Compiler · Error Loop',
        sublabel: 'error correction agent 用 compiler 检查 · 迭代修复',
        icon: 'i-lucide-bug-off',
        accent: 'emerald',
        flow: 'analyst-flow',
        span: 1,
        refs: ['S9'],
      },
      {
        id: 'rbac-mask',
        label: 'RBAC + Masking',
        sublabel: '底表 row access / column mask 在执行时传播到 semantic view',
        icon: 'i-lucide-shield',
        accent: 'emerald',
        flow: 'policy-runtime',
        span: 1,
        refs: ['S10'],
      },
    ],
  },
]
