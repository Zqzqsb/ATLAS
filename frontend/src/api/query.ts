import { createSSEStream, apiClient } from './client'
import type {
  Text2SQLRequest,
  Text2SQLResult,
  QueryRecord,
  ExecutionResult
} from '@/types'

// Field suggestion types
export interface SuggestedField {
  name: string
  description: string
  selected: boolean
  source: string
}

export interface SuggestFieldsRequest {
  question: string
  databaseId: string
  database: string
  language?: string
}

export interface SuggestFieldsResponse {
  suggested_fields: SuggestedField[]
  analysis_note: string
}

export const queryApi = {
  /**
   * Execute Text2SQL query with SSE streaming
   */
  stream: (
    request: Text2SQLRequest,
    onEvent: (event: { type: string; data: any }) => void,
    onError?: (error: Error) => void,
    onComplete?: () => void
  ): (() => void) => {
    return createSSEStream(
      '/api/v1/text2sql/stream',
      {
        question: request.question,
        database_id: request.databaseId,
        database: request.database,
        field_description: request.fieldDescription || '',
        injected_grounding: request.injectedGrounding || undefined,
        options: {
          use_rich_context: request.options.useRichContext,
          use_react: request.options.useReact,
          use_grounding: request.options.useGrounding,
          skip_linking: request.options.skipLinking || false,
          max_iterations: request.options.maxIterations,
          grounding_only: request.options.groundingOnly || false,
          skip_grounding: request.options.skipGrounding || false,
          stream: true
        }
      },
      onEvent,
      onError,
      onComplete
    )
  },

  /**
   * Execute SQL directly
   */
  executeSql: async (databaseId: string, sql: string): Promise<ExecutionResult> => {
    // Mock execution for demo
    await new Promise(r => setTimeout(r, 500))

    return {
      columns: ['id', 'name', 'amount'],
      rows: [
        [1, '张三', 12000],
        [2, '李四', 8500],
        [3, '王五', 15000]
      ],
      rowCount: 3,
      duration: 0.15
    }
  },

  /**
   * Get query history
   */
  getHistory: async (databaseId: string, limit = 50): Promise<QueryRecord[]> => {
    // Mock history for demo
    return [
      {
        id: 'q-1',
        databaseId,
        question: '查询VIP客户的总订单金额',
        sql: 'SELECT c.name, SUM(o.total_amount) FROM customers c JOIN orders o ON c.id = o.customer_id WHERE c.cust_lvl >= 4 GROUP BY c.id',
        duration: 1.2,
        timestamp: new Date(Date.now() - 3600000).toISOString(),
        isCorrect: true,
        feedback: 'positive'
      },
      {
        id: 'q-2',
        databaseId,
        question: '查询CA州最近7天的订单',
        sql: "SELECT * FROM orders o JOIN customers c ON o.customer_id = c.id WHERE c.state = 'California' AND o.order_date >= DATE_SUB(NOW(), INTERVAL 7 DAY)",
        duration: 0.8,
        timestamp: new Date(Date.now() - 7200000).toISOString(),
        isCorrect: true
      },
      {
        id: 'q-3',
        databaseId,
        question: '统计各等级客户数量',
        sql: 'SELECT cust_lvl, COUNT(*) as count FROM customers GROUP BY cust_lvl ORDER BY cust_lvl',
        duration: 0.5,
        timestamp: new Date(Date.now() - 86400000).toISOString(),
        isCorrect: true,
        feedback: 'positive'
      }
    ]
  },

  /**
   * Submit feedback for a query
   */
  submitFeedback: async (
    queryId: string,
    feedback: 'positive' | 'negative',
    note?: string
  ): Promise<void> => {
    // Mock API call
    await new Promise(r => setTimeout(r, 300))
    console.log('Feedback submitted:', { queryId, feedback, note })
  },

  /**
   * Delete query from history
   */
  deleteFromHistory: async (queryId: string): Promise<void> => {
    // Mock API call
    await new Promise(r => setTimeout(r, 200))
  },

  /**
   * Suggest output fields based on question and schema
   */
  suggestFields: async (request: SuggestFieldsRequest): Promise<SuggestFieldsResponse> => {
    const response = await apiClient.post<SuggestFieldsResponse>('/text2sql/suggest-fields', {
      question: request.question,
      database_id: request.databaseId,
      database: request.database,
      language: request.language || 'Chinese'
    })
    return response.data
  }
}
