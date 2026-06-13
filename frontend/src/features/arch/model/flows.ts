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
]

export function getFlow(id: string | null): FlowDef | null {
  if (!id) return null
  return flows.find((f) => f.id === id) ?? null
}
