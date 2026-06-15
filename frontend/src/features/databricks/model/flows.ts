import type { AccentKey } from '../../arch/model/architecture'

export interface DbxFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const dbxFlows: DbxFlowDef[] = [
  {
    id: 'metric-view',
    label: 'Metric View',
    title: 'Metric View · 关系与指标模型',
    subtitle: 'YAML 显式建模 source / joins / dimensions / measures / filter；持久化为 UC 原生对象，写一次的指标公式可被任意 dimension 重写为正确聚合 SQL',
    icon: 'i-lucide-shapes',
    accent: 'amber',
  },
  {
    id: 'agent-metadata',
    label: 'Agent Metadata',
    subtitle: '在 metric view / table 之上挂 synonyms · display_name · format · description 等列级富上下文，Genie 与第三方 LLM 工具靠它把用户输入对到字段',
    title: 'Agent Metadata · 富上下文层',
    icon: 'i-lucide-tags',
    accent: 'amber',
  },
  {
    id: 'genie',
    label: 'Genie Space',
    title: 'Genie Space · NL2SQL Agent 编排',
    subtitle: 'NL → SQL 的 Agent 形态：选定 metric view / table 子集，靠 example SQL 关键词匹配召回，trusted asset / verified query 复用，benchmark 评估准确性',
    icon: 'i-lucide-sparkles',
    accent: 'slate',
  },
  {
    id: 'uc-federation',
    label: 'UC + Federation',
    title: 'Unity Catalog + Lakehouse Federation',
    subtitle: 'metric view 的 source 必须是 UC 表类资产；Lakehouse Federation 把 PG / MySQL / Snowflake / BigQuery 等外部源以 foreign table 形式接入，构成跨源统一语义层',
    icon: 'i-lucide-database',
    accent: 'indigo',
  },
  {
    id: 'policy-runtime',
    label: 'Policy Runtime',
    title: 'Row Filter / Column Mask 运行时',
    subtitle: '策略不能直接挂在 view 上，但写在底表的 row filter / column mask 会按当前查询用户的身份在执行时强制传播——权限是查询时的事，不是 prompt 时的事',
    icon: 'i-lucide-shield',
    accent: 'emerald',
  },
  {
    id: 'mv-materialize',
    label: 'MV Materialization',
    title: 'Metric View Materialization',
    subtitle: '增量刷新 + 自动查询重写：metric view 可被物化成实体表，命中重写后查询直接读物化结果——在不改 NL2SQL 路径的前提下影响 SQL 形状与延迟',
    icon: 'i-lucide-snowflake',
    accent: 'violet',
  },
]

export function getDbxFlow(id: string | null): DbxFlowDef | null {
  if (!id) return null
  return dbxFlows.find((f) => f.id === id) ?? null
}
