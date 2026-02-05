import client from './client'
import type { Database, DatabaseConfig, DatabaseInfo, SchemaInfo, TableInfo } from '@/types'

// Mock databases for demo
const mockDatabases: Database[] = [
  {
    id: 'tvshow',
    name: 'tvshow',
    displayName: 'TV Show Database',
    type: 'sqlite',
    status: 'connected',
    tableCount: 5,
    hasRichContext: true,
    contextCount: 12,
    lastConnected: new Date().toISOString(),
    description: 'Spider dataset - TV Show domain',
    tags: ['demo', 'spider']
  },
  {
    id: 'ecommerce',
    name: 'ecommerce',
    displayName: '电商数据库',
    type: 'mariadb',
    host: 'localhost',
    port: 3306,
    status: 'connected',
    tableCount: 8,
    hasRichContext: true,
    contextCount: 24,
    lastConnected: new Date().toISOString(),
    description: '电商场景精品演示库 - 包含用户、订单、商品等表',
    tags: ['demo', 'ecommerce']
  },
  {
    id: 'finance',
    name: 'finance',
    displayName: '金融数据库',
    type: 'mariadb',
    host: 'localhost',
    port: 3306,
    status: 'disconnected',
    tableCount: 15,
    hasRichContext: false,
    contextCount: 0,
    description: '金融场景演示库 - 待配置',
    tags: ['demo', 'finance']
  }
]

// Mock schema for ecommerce
const mockEcommerceSchema: SchemaInfo = {
  tables: [
    {
      name: 'customers',
      rowCount: 1500,
      hasContext: true,
      description: '客户信息表',
      columns: [
        { name: 'id', type: 'INT', isPrimaryKey: true },
        { name: 'name', type: 'VARCHAR(100)' },
        { name: 'email', type: 'VARCHAR(255)' },
        { name: 'phone', type: 'VARCHAR(20)', hasContext: true },
        { name: 'cust_lvl', type: 'INT', hasContext: true },
        { name: 'state', type: 'VARCHAR(50)', hasContext: true },
        { name: 'created_at', type: 'DATETIME' }
      ]
    },
    {
      name: 'orders',
      rowCount: 5000,
      hasContext: true,
      description: '订单表',
      columns: [
        { name: 'id', type: 'INT', isPrimaryKey: true },
        { name: 'customer_id', type: 'INT', isForeignKey: true, references: { table: 'customers', column: 'id' } },
        { name: 'order_date', type: 'DATE' },
        { name: 'status', type: 'INT', hasContext: true },
        { name: 'total_amount', type: 'DECIMAL(10,2)', hasContext: true },
        { name: 'discount', type: 'DECIMAL(5,2)' }
      ]
    },
    {
      name: 'products',
      rowCount: 500,
      hasContext: false,
      description: '商品表',
      columns: [
        { name: 'id', type: 'INT', isPrimaryKey: true },
        { name: 'name', type: 'VARCHAR(200)' },
        { name: 'category', type: 'VARCHAR(50)' },
        { name: 'price', type: 'DECIMAL(10,2)' },
        { name: 'stock', type: 'INT' }
      ]
    },
    {
      name: 'order_items',
      rowCount: 12000,
      hasContext: false,
      description: '订单明细表',
      columns: [
        { name: 'id', type: 'INT', isPrimaryKey: true },
        { name: 'order_id', type: 'INT', isForeignKey: true, references: { table: 'orders', column: 'id' } },
        { name: 'product_id', type: 'INT', isForeignKey: true, references: { table: 'products', column: 'id' } },
        { name: 'quantity', type: 'INT' },
        { name: 'unit_price', type: 'DECIMAL(10,2)' }
      ]
    }
  ],
  relationships: [
    { from: { table: 'orders', column: 'customer_id' }, to: { table: 'customers', column: 'id' }, type: 'many-to-many' },
    { from: { table: 'order_items', column: 'order_id' }, to: { table: 'orders', column: 'id' }, type: 'many-to-many' },
    { from: { table: 'order_items', column: 'product_id' }, to: { table: 'products', column: 'id' }, type: 'many-to-many' }
  ],
  lastUpdated: new Date().toISOString()
}

// Transform lakebase datasource to Database type
function transformDatasource(ds: any): Database {
  const dbName = ds.database_name?.String || ds.database_name || ds.name
  return {
    id: ds.name, // Use name as ID (e.g., "spider_tvshow")
    name: ds.name,
    displayName: dbName,
    type: ds.db_type || 'mariadb',
    host: ds.host?.String || ds.host || 'localhost',
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

// Check if running in development mode for verbose logging
const isDev = import.meta.env.DEV

export const databaseApi = {
  /**
   * Get all connected databases from lakebase
   */
  list: async (): Promise<Database[]> => {
    try {
      // First try lakebase datasources API
      const response = await client.get<{ datasources: any[] }>('/lakebase/datasources')
      if (response.data.datasources && response.data.datasources.length > 0) {
        return response.data.datasources.map(transformDatasource)
      }
      // Fallback to legacy databases API
      const legacyResponse = await client.get<Database[]>('/databases')
      return legacyResponse.data
    } catch (error) {
      // Return mock data for demo (with dev warning)
      if (isDev) {
        console.warn('[API] database.list failed, using mock data:', error)
      }
      return mockDatabases
    }
  },

  /**
   * Get database details
   */
  get: async (id: string): Promise<Database | null> => {
    try {
      const response = await client.get<Database>(`/databases/${id}`)
      return response.data
    } catch (error) {
      if (isDev) {
        console.warn(`[API] database.get(${id}) failed, using mock data:`, error)
      }
      return mockDatabases.find(db => db.id === id) || null
    }
  },

  /**
   * Add new database connection
   */
  create: async (config: DatabaseConfig): Promise<Database> => {
    const response = await client.post<Database>('/databases', config)
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
    try {
      const response = await client.post<{ success: boolean; message?: string }>(`/databases/${id}/test`)
      return response.data
    } catch (error) {
      // Mock response (with dev warning)
      if (isDev) {
        console.warn(`[API] database.testConnection(${id}) failed, using mock:`, error)
      }
      const db = mockDatabases.find(d => d.id === id)
      return {
        success: db?.status === 'connected',
        message: db?.status === 'connected' ? 'Connection successful' : 'Connection failed'
      }
    }
  },

  /**
   * Get database schema
   */
  getSchema: async (id: string): Promise<SchemaInfo> => {
    try {
      const response = await client.get<SchemaInfo>(`/databases/${id}/schema`)
      return response.data
    } catch (error) {
      // Return mock schema (with dev warning)
      if (isDev) {
        console.warn(`[API] database.getSchema(${id}) failed, using mock:`, error)
      }
      if (id === 'ecommerce') {
        return mockEcommerceSchema
      }
      return {
        tables: [],
        relationships: [],
        lastUpdated: new Date().toISOString()
      }
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
    try {
      const response = await client.get(`/databases/${id}/stats`)
      return response.data
    } catch (error) {
      if (isDev) {
        console.warn(`[API] database.getStats(${id}) failed, using mock:`, error)
      }
      return {
        queryCount: 156,
        avgDuration: 1.2,
        successRate: 0.89,
        contextUsageRate: 0.75
      }
    }
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
