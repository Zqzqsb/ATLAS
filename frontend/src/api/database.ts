import client from './client'
import type { Database, DatabaseConfig, SchemaInfo } from '@/types'

const DISPLAY_NAME_MAP: Record<string, string> = {
  'lucid_evolution': 'Atlas Evolution',
}
const HOST_DISPLAY_MAP: Record<string, string> = {
  'lucid-mariadb': 'atlas-mariadb',
}

// Transform lakebase datasource to Database type
function transformDatasource(ds: any): Database {
  const dbName = ds.database_name?.String || ds.database_name || ds.name
  const rawHost = ds.host?.String || ds.host || 'localhost'
  return {
    id: ds.name, // Use name as ID (e.g., "spider_tvshow")
    name: ds.name,
    displayName: DISPLAY_NAME_MAP[ds.name] || dbName,
    type: ds.db_type || 'mariadb',
    host: HOST_DISPLAY_MAP[rawHost] || rawHost,
    port: ds.port?.Int32 || ds.port || 3306,
    status: ds.status === 'active' ? 'connected' : 'disconnected',
    tableCount: ds.tables_count || 0,
    hasRichContext: (ds.context_count || 0) > 0,
    contextCount: ds.context_count || 0,
    lastConnected: ds.updated_at,
    description: ds.description?.String || ds.description || '',
    tags: ['lakebase'],
    metadata: {
      lakebaseId: ds.id // Numeric ID for lakebase API calls
    }
  }
}

export const databaseApi = {
  /**
   * Get all connected databases from lakebase
   */
  list: async (): Promise<Database[]> => {
    const response = await client.get<{ datasources: any[] }>('/lakebase/datasources')
    if (response.data.datasources && response.data.datasources.length > 0) {
      return response.data.datasources.map(transformDatasource)
    }
    return []
  },

  /**
   * Get database details by lakebase datasource ID
   */
  get: async (id: string): Promise<Database | null> => {
    try {
      const response = await client.get<any>(`/lakebase/datasources`)
      const ds = response.data.datasources?.find((d: any) => d.name === id)
      return ds ? transformDatasource(ds) : null
    } catch {
      return null
    }
  },

  /**
   * Add new database connection via POST /connections
   * Backend expects: { id, name, type, host, port, user, password, database }
   */
  create: async (config: DatabaseConfig): Promise<any> => {
    const payload = {
      id: config.name,           // Use name as connection ID
      name: config.name,
      type: config.type,
      host: config.host,
      port: config.port,
      user: config.username,
      password: config.password,
      database: config.database,
    }
    const response = await client.post('/connections', payload)
    return response.data
  },

  /**
   * Delete database connection
   */
  delete: async (id: string): Promise<void> => {
    await client.delete(`/databases/${id}`)
  },

  /**
   * Test database connection
   */
  testConnection: async (id: string): Promise<{ success: boolean; message?: string }> => {
    const response = await client.post<{ success: boolean; message?: string }>(`/databases/${id}/test`)
    return response.data
  },

  /**
   * Get database schema from lakebase (rc_tables + rc_columns)
   */
  getSchema: async (id: string): Promise<SchemaInfo> => {
    // id here is the datasource name; need to resolve to lakebaseId first
    try {
      const response = await client.get<any>(`/lakebase/datasources`)
      const ds = response.data.datasources?.find((d: any) => d.name === id)
      if (!ds) {
        return { tables: [], relationships: [], lastUpdated: new Date().toISOString() }
      }
      const detailRes = await client.get<any>(`/lakebase/datasources/${ds.id}`)
      const data = detailRes.data
      return {
        tables: (data.tables || []).map((t: any) => ({
          name: t.table_name,
          rowCount: t.row_count || 0,
          hasContext: !!(t.description?.String || t.description),
          description: t.description?.String || t.description || '',
          columns: (data.columns || [])
            .filter((c: any) => c.table_name === t.table_name)
            .map((c: any) => ({
              name: c.column_name,
              type: c.data_type,
              isPrimaryKey: c.is_primary_key,
              isForeignKey: c.is_foreign_key,
              hasContext: !!(c.description?.String || c.description),
              description: c.description?.String || c.description || ''
            }))
        })),
        relationships: (data.relations || []).map((r: any) => ({
          from: { table: r.from_table, column: r.from_column },
          to: { table: r.to_table, column: r.to_column },
          type: r.relation_type || 'foreign_key'
        })),
        lastUpdated: data.updated_at || new Date().toISOString()
      }
    } catch {
      return { tables: [], relationships: [], lastUpdated: new Date().toISOString() }
    }
  },

  /**
   * Refresh database schema
   */
  refreshSchema: async (id: string): Promise<SchemaInfo> => {
    const response = await client.post<SchemaInfo>(`/databases/${id}/schema/refresh`)
    return response.data
  },

  /**
   * Get database statistics
   */
  getStats: async (id: string): Promise<{
    queryCount: number
    avgDuration: number
    successRate: number
    contextUsageRate: number
  }> => {
    const response = await client.get(`/databases/${id}/stats`)
    return response.data
  },

  /**
   * Sync schema for a connection by connection ID (string).
   * This creates/updates rc_datasources and syncs physical schema from information_schema.
   * @param connectionId - The connection ID (e.g., "spider_tvshow")
   */
  syncConnectionSchema: async (connectionId: string): Promise<{ success: boolean; datasource_id: number; tables: number; columns: number; relations: number }> => {
    const response = await client.post<{ success: boolean; datasource_id: number; tables: number; columns: number; relations: number }>(
      `/connections/${connectionId}/sync-schema`
    )
    return response.data
  },

  /**
   * Sync schema from the target business database into rc_tables/rc_columns/rc_relations
   * @param datasourceId - The lakebase datasource ID (numeric)
   */
  syncSchema: async (datasourceId: number): Promise<{ success: boolean; tables: number; columns: number; relations: number }> => {
    const response = await client.post<{ success: boolean; tables: number; columns: number; relations: number }>(
      `/lakebase/datasources/${datasourceId}/sync-schema`
    )
    return response.data
  },

  /**
   * Delete a datasource and all its associated RC data
   * @param datasourceId - The lakebase datasource ID (numeric)
   */
  deleteDatasource: async (datasourceId: number): Promise<{ success: boolean; message: string }> => {
    const response = await client.delete<{ success: boolean; message: string }>(
      `/lakebase/datasources/${datasourceId}`
    )
    return response.data
  },

  /**
   * Prune all rich context for a datasource
   * @param datasourceId - The lakebase datasource ID (numeric)
   */
  pruneContext: async (datasourceId: number): Promise<{ success: boolean; message: string }> => {
    const response = await client.delete<{ success: boolean; message: string; datasource: string }>(
      `/lakebase/datasources/${datasourceId}/prune`
    )
    return response.data
  }
}
