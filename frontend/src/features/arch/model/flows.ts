/**
 * Module identity registry — lightweight metadata for each drillable module.
 * Used by the overview (drill target) and the ModuleDetail header.
 * The full internal architecture lives in modules.ts.
 */
import type { AccentKey } from './architecture'

export interface FlowDef {
  id: string
  label: string
  title: string
  subtitle: string
  icon: string
  accent: AccentKey
}

export const flows: FlowDef[] = [
  {
    id: 'onboarding',
    label: 'Onboarding',
    title: 'Database Onboarding',
    subtitle: '接入新库时，Coordinator 切分调度、Worker 探查执行，自动沉淀 Rich Context 与向量索引',
    icon: 'i-lucide-database-zap',
    accent: 'emerald',
  },
  {
    id: 'inference',
    label: 'Inference',
    title: 'Text-to-SQL Inference',
    subtitle: '自适应 Grounding 定位表/列，ReAct 生成并自校验 SQL，最终执行返回结果',
    icon: 'i-lucide-git-graph',
    accent: 'blue',
  },
  {
    id: 'maintain',
    label: 'Self-Maintenance',
    title: 'Self-Maintenance',
    subtitle: 'Schema 变更触发 Signal，Coordinator 标记失效并派发任务，Executor 探查重写自愈，收尾重嵌入',
    icon: 'i-lucide-bot',
    accent: 'amber',
  },
  {
    id: 'kernel',
    label: 'ReAct Kernel',
    title: 'ReAct Kernel · 共用推理内核',
    subtitle: '三大流程共用同一条 Reason→Act→Observe 循环：场景只注入 prompt / 工具子集 / 预算（EngineConfig），内核基于 langchaingo 文本 ReAct，旁路采集 Step 推 SSE',
    icon: 'i-lucide-cpu',
    accent: 'violet',
  },
  {
    id: 'storage',
    label: 'Lakebase Storage',
    title: 'Lakebase Storage · MariaDB 12',
    subtitle: '以 rc_datasources 为根的 rc_* 表族：结构 / 语义 / 向量同库，rc_embeddings 用原生 VECTOR(2048) + HNSW(COSINE)，is_expired/is_stale/is_deleted 驱动自维护',
    icon: 'i-lucide-database',
    accent: 'indigo',
  },
]

export function getFlow(id: string | null): FlowDef | null {
  if (!id) return null
  return flows.find((f) => f.id === id) ?? null
}
