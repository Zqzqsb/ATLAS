import client from './client'
import type {
  RichContext,
  ContextType,
  ContextSource,
  ContextFilter
} from '@/types'

export const contextApi = {
  /**
   * Get all contexts for a database (via lakebase datasource API)
   * This is handled by workspace store directly — kept for backward compat
   */
  list: async (databaseId: string, _filter?: ContextFilter): Promise<RichContext[]> => {
    // Contexts are fetched via lakebase datasource API in workspace store
    // This is kept as a stub for backward compatibility
    return []
  },

  /**
   * Get contexts for a specific table
   */
  getByTable: async (databaseId: string, tableName: string): Promise<RichContext[]> => {
    return contextApi.list(databaseId, { tableName })
  },

  /**
   * Create a new context via lakebase API
   * Writes to rc_tables/rc_columns/rc_terms depending on type, and triggers embedding
   */
  create: async (
    lakebaseId: string | number,
    context: { tableName: string; columnName?: string; type: ContextType; content: string }
  ): Promise<{ success: boolean; message: string }> => {
    const response = await client.post(`/lakebase/datasources/${lakebaseId}/context`, {
      table_name: context.tableName,
      column_name: context.columnName || '',
      type: context.type,
      content: context.content
    })
    return response.data
  },

  /**
   * Delete a context via lakebase API
   * Clears the corresponding field in rc_tables/rc_columns or removes from rc_terms
   */
  delete: async (
    lakebaseId: string | number,
    context: { tableName: string; columnName?: string; type: string }
  ): Promise<{ success: boolean; message: string }> => {
    const response = await client.delete(`/lakebase/datasources/${lakebaseId}/context`, {
      data: {
        table_name: context.tableName,
        column_name: context.columnName || '',
        type: context.type
      }
    })
    return response.data
  }
}
