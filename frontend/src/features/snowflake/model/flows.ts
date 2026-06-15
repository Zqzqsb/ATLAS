import type { AccentKey } from '../../arch/model/architecture'

export interface SnowFlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const snowFlows: SnowFlowDef[] = [
  {
    id: 'semantic-view',
    label: 'Semantic View',
    title: 'Semantic View · DDL 原生对象',
    subtitle: 'CREATE SEMANTIC VIEW DDL 显式建模 TABLES / RELATIONSHIPS（含 ASOF/range join）/ FACTS / DIMENSIONS / METRICS（含窗口变体），是 schema 级原生对象，由 GRANT / DDL 治理',
    icon: 'i-lucide-shapes',
    accent: 'amber',
  },
  {
    id: 'analyst-flow',
    label: 'Cortex Analyst Flow',
    title: 'Cortex Analyst · NL → SQL 流程',
    subtitle: '官方 REST API：读取 semantic view → 简化 logical schema 上生成 SQL → 后处理为物理 SQL → SQL compiler 检查 → error correction loop 迭代修复 → 返回 SQL（不执行）',
    icon: 'i-lucide-radio',
    accent: 'slate',
  },
  {
    id: 'vqr',
    label: 'VQR · Custom Instructions',
    title: 'Verified Query Repository · 反馈与经验注入',
    subtitle: 'verified_queries（NL-SQL 对 + verified_by/verified_at）+ custom_instructions（自然语言规则）作为富上下文按问题相关性检索注入；命中 VQR 即 verified answer',
    icon: 'i-lucide-shield-check',
    accent: 'violet',
  },
  {
    id: 'cortex-search',
    label: 'Cortex Search',
    title: 'Cortex Search · 高基数 literal 召回',
    subtitle: '高基数 dimension（>10 个不同值）配置 cortex_search_service，运行时对底层列实际值做向量+关键词+rerank 混合检索，避免把巨大列值放进语义模型本体',
    icon: 'i-lucide-search',
    accent: 'blue',
  },
  {
    id: 'autopilot',
    label: 'Autopilot · Suggestions',
    title: 'Semantic View Autopilot + Snowsight Suggestions',
    subtitle: 'Autopilot 基于 Query History + 表元数据 AI 生成候选 semantic view（含描述 / 关系 / verified queries 草稿）；Snowsight Suggestions 在已有模型上提候选——全部需人工审核',
    icon: 'i-lucide-sparkles',
    accent: 'violet',
  },
  {
    id: 'policy-runtime',
    label: 'Policy Runtime',
    title: 'RBAC + Masking · 底表策略传播',
    subtitle: 'masking / row access policy 不能直接挂在 semantic view 上，但底表的策略会按当前查询用户身份在执行时强制传播；样例值属元数据、不受 masking 保护',
    icon: 'i-lucide-shield',
    accent: 'emerald',
  },
]

export function getSnowFlow(id: string | null): SnowFlowDef | null {
  if (!id) return null
  return snowFlows.find((f) => f.id === id) ?? null
}
