import client from './client'
import type {
  RichContext,
  ContextType,
  ContextSource,
  ContextFilter
} from '@/types'

// Mock Rich Contexts
const mockContexts: RichContext[] = [
  {
    id: 'ctx-1',
    databaseId: 'ecommerce',
    tableId: 'customers',
    tableName: 'customers',
    columnName: 'cust_lvl',
    type: 'value_mapping',
    content: 'cust_lvl 客户等级枚举值: 1=普通会员, 2=银卡会员, 3=金卡会员, 4=白金VIP, 5=钻石VIP',
    createdAt: '2024-01-15T10:30:00Z',
    source: 'manual',
    confidence: 1.0,
    usageCount: 45
  },
  {
    id: 'ctx-2',
    databaseId: 'ecommerce',
    tableId: 'customers',
    tableName: 'customers',
    columnName: 'state',
    type: 'synonym',
    content: 'state 字段使用缩写: CA=California, NY=New York, TX=Texas, FL=Florida',
    createdAt: '2024-01-15T11:00:00Z',
    source: 'auto',
    confidence: 0.95,
    usageCount: 23
  },
  {
    id: 'ctx-3',
    databaseId: 'ecommerce',
    tableId: 'customers',
    tableName: 'customers',
    columnName: 'phone',
    type: 'description',
    content: 'phone 电话号码格式不统一，可能是 138-0000-1234, 13800001234, +86-138-0000-1234 等多种格式',
    createdAt: '2024-01-14T16:45:00Z',
    source: 'feedback',
    confidence: 0.9,
    usageCount: 12
  },
  {
    id: 'ctx-4',
    databaseId: 'ecommerce',
    tableId: 'orders',
    tableName: 'orders',
    columnName: 'status',
    type: 'value_mapping',
    content: 'status 订单状态枚举: 0=待支付, 1=已支付, 2=已发货, 3=已完成, 4=已取消, 5=已退款',
    createdAt: '2024-01-15T09:00:00Z',
    source: 'manual',
    confidence: 1.0,
    usageCount: 67
  },
  {
    id: 'ctx-5',
    databaseId: 'ecommerce',
    tableId: 'orders',
    tableName: 'orders',
    columnName: 'total_amount',
    type: 'business_rule',
    content: 'total_amount 单位为分(cent)，显示时需要除以100转换为元。实际支付金额 = total_amount * (1 - discount / 100)',
    createdAt: '2024-01-14T14:20:00Z',
    source: 'manual',
    confidence: 1.0,
    usageCount: 34
  },
  {
    id: 'ctx-6',
    databaseId: 'ecommerce',
    tableId: 'orders',
    tableName: 'orders',
    type: 'business_rule',
    content: '"有效订单"指 status IN (1, 2, 3) 的订单，即已支付、已发货、已完成状态',
    createdAt: '2024-01-16T10:00:00Z',
    source: 'feedback',
    confidence: 0.95,
    usageCount: 28
  }
]

export const contextApi = {
  /**
   * Get all contexts for a database
   */
  list: async (databaseId: string, filter?: ContextFilter): Promise<RichContext[]> => {
    try {
      const response = await client.get<RichContext[]>(`/databases/${databaseId}/contexts`, {
        params: filter
      })
      return response.data
    } catch {
      // Return filtered mock data
      let contexts = mockContexts.filter(c => c.databaseId === databaseId)

      if (filter?.tableName) {
        contexts = contexts.filter(c => c.tableName === filter.tableName)
      }
      if (filter?.columnName) {
        contexts = contexts.filter(c => c.columnName === filter.columnName)
      }
      if (filter?.type) {
        contexts = contexts.filter(c => c.type === filter.type)
      }
      if (filter?.search) {
        const search = filter.search.toLowerCase()
        contexts = contexts.filter(c =>
          c.content.toLowerCase().includes(search) ||
          c.tableName.toLowerCase().includes(search) ||
          c.columnName?.toLowerCase().includes(search)
        )
      }

      return contexts
    }
  },

  /**
   * Get contexts for a specific table
   */
  getByTable: async (databaseId: string, tableName: string): Promise<RichContext[]> => {
    return contextApi.list(databaseId, { tableName })
  },

  /**
   * Get a single context by ID
   */
  get: async (databaseId: string, contextId: string): Promise<RichContext | null> => {
    try {
      const response = await client.get<RichContext>(`/databases/${databaseId}/contexts/${contextId}`)
      return response.data
    } catch {
      return mockContexts.find(c => c.id === contextId) || null
    }
  },

  /**
   * Create a new context
   */
  create: async (context: Omit<RichContext, 'id' | 'createdAt' | 'usageCount'>): Promise<RichContext> => {
    try {
      const response = await client.post<RichContext>(`/databases/${context.databaseId}/contexts`, context)
      return response.data
    } catch {
      // Mock creation
      const newContext: RichContext = {
        ...context,
        id: `ctx-${Date.now()}`,
        createdAt: new Date().toISOString(),
        usageCount: 0
      }
      mockContexts.push(newContext)
      return newContext
    }
  },

  /**
   * Update an existing context
   */
  update: async (databaseId: string, contextId: string, updates: Partial<RichContext>): Promise<RichContext> => {
    try {
      const response = await client.put<RichContext>(`/databases/${databaseId}/contexts/${contextId}`, updates)
      return response.data
    } catch {
      // Mock update
      const index = mockContexts.findIndex(c => c.id === contextId)
      if (index >= 0) {
        mockContexts[index] = { ...mockContexts[index], ...updates, updatedAt: new Date().toISOString() } as RichContext
        return mockContexts[index]!
      }
      throw new Error('Context not found')
    }
  },

  /**
   * Delete a context
   */
  delete: async (databaseId: string, contextId: string): Promise<void> => {
    try {
      await client.delete(`/databases/${databaseId}/contexts/${contextId}`)
    } catch {
      // Mock deletion
      const index = mockContexts.findIndex(c => c.id === contextId)
      if (index >= 0) {
        mockContexts.splice(index, 1)
      }
    }
  },

  /**
   * Auto-generate contexts from schema analysis
   */
  generateFromSchema: async (databaseId: string, tableName?: string): Promise<RichContext[]> => {
    try {
      const response = await client.post<RichContext[]>(`/databases/${databaseId}/contexts/generate`, {
        tableName
      })
      return response.data
    } catch {
      // Mock generation - return sample generated contexts
      await new Promise(r => setTimeout(r, 1500))
      return [
        {
          id: `ctx-gen-${Date.now()}`,
          databaseId,
          tableId: tableName || 'unknown',
          tableName: tableName || 'unknown',
          type: 'description',
          content: '自动分析生成的 Schema 描述',
          createdAt: new Date().toISOString(),
          source: 'auto',
          confidence: 0.8,
          usageCount: 0
        }
      ]
    }
  },

  /**
   * Import contexts from file
   */
  import: async (databaseId: string, file: File): Promise<{ imported: number; errors: string[] }> => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await client.post(`/databases/${databaseId}/contexts/import`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    return response.data
  },

  /**
   * Export contexts to JSON
   */
  export: async (databaseId: string): Promise<Blob> => {
    const response = await client.get(`/databases/${databaseId}/contexts/export`, {
      responseType: 'blob'
    })
    return response.data
  }
}
