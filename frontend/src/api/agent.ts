import { request } from './client'

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

export const agentApi = {
  /**
   * Get change logs for a datasource
   */
  getChangeLogs: (datasourceId: number, limit = 50) =>
    request<{ datasource_id: number; logs: ChangeLog[]; count: number }>({
      url: `/agent/logs/${datasourceId}`,
      method: 'GET',
      params: { limit }
    })
}
