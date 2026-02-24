import { request } from './client'

export interface AgentStatus {
  running: boolean
  config: {
    enable_ddl_detection: boolean
    check_interval: number
    auto_refresh_context: boolean
    max_concurrent_tasks: number
  }
  llm_available: boolean
  last_run: string
  auto_refresh_enabled: boolean
  last_result?: MaintenanceResult
}

export interface MaintenanceResult {
  datasource_id: number
  start_time: string
  end_time: string
  duration_ms: number
  schema_changes_found: number
  context_expired: number
  context_refreshed: number
  context_created: number
  embeddings_updated: number
  errors: string[]
  success: boolean
}

export interface ChangeLog {
  id: number
  datasource_id: number
  table_name: string
  change_type: 'schema_change' | 'context_update' | 'context_expire'
  change_detail: any
  old_value?: any
  new_value?: any
  trigger_source: 'agent' | 'user' | 'system'
  change_reason: string
  created_at: string
}

export interface ChangeLogSummary {
  total_changes: number
  schema_changes: number
  context_updates: number
  context_expiries: number
  by_table: Record<string, number>
  by_trigger_source: Record<string, number>
  oldest_change: string
  newest_change: string
}

export interface SchemaChange {
  change_type: string
  table_name: string
  column_name?: string
  old_definition?: string
  new_definition?: string
  details?: Record<string, any>
  detected_at: string
}

export const agentApi = {
  /**
   * Get agent status
   */
  getStatus: () => 
    request<AgentStatus>({ url: '/agent/status', method: 'GET' }),

  /**
   * Run maintenance for a datasource
   */
  runMaintenance: (datasourceId: number) =>
    request<{ message: string; result: MaintenanceResult }>({
      url: `/agent/maintenance/${datasourceId}`,
      method: 'POST'
    }),

  /**
   * Trigger context refresh for a datasource
   */
  triggerContextRefresh: (datasourceId: number) =>
    request<{ message: string; total: number; success_count: number; results: any[] }>({
      url: `/agent/refresh/${datasourceId}`,
      method: 'POST'
    }),

  /**
   * Get change logs for a datasource
   */
  getChangeLogs: (datasourceId: number, limit = 50) =>
    request<{ datasource_id: number; logs: ChangeLog[]; count: number }>({
      url: `/agent/logs/${datasourceId}`,
      method: 'GET',
      params: { limit }
    }),

  /**
   * Get change log summary for a datasource
   */
  getChangeLogSummary: (datasourceId: number) =>
    request<ChangeLogSummary>({
      url: `/agent/logs/${datasourceId}/summary`,
      method: 'GET'
    }),

  /**
   * Start the agent service (placeholder — agent is always active when LLM is available)
   */
  start: () =>
    request<{ message: string }>({ url: '/agent/start', method: 'POST' }),

  /**
   * Stop the agent service (placeholder)
   */
  stop: () =>
    request<{ message: string }>({ url: '/agent/stop', method: 'POST' }),

  /**
   * Simulate a DDL change for the self-maintenance demo
   */
  simulateDDL: (datasourceId: number, ddl: string) =>
    request<{ message: string; parsed_change: any; result: MaintenanceResult }>({
      url: `/agent/simulate-ddl/${datasourceId}`,
      method: 'POST',
      data: { ddl }
    })
}
