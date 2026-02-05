// Re-export all API modules
export { default as client } from './client'
export { createSSEStream } from './client'
export { databaseApi } from './database'
export { contextApi } from './context'
export { queryApi } from './query'
export { agentApi } from './agent'
export type { AgentStatus, MaintenanceResult, ChangeLog, ChangeLogSummary, SchemaChange } from './agent'

// Re-export demo APIs (preserved from original)
import type {
  ComparisonCase,
  ComparisonResult,
  MaintenanceLog,
  RichContext
} from '@/types'

// Comparison API (Mock for demo)
export const comparisonApi = {
  getCases: async (): Promise<ComparisonCase[]> => {
    return [
      {
        id: 'D1',
        name: '脏数据：缩写识别',
        category: 'dirty_data',
        question: '查询CA州的所有客户',
        description: 'CA 是 California 的缩写，Rich Context 提供缩写映射',
        expectedSql: "SELECT * FROM customers WHERE state = 'California'",
        difficulty: 'easy'
      },
      {
        id: 'D2',
        name: '脏数据：格式变体',
        category: 'dirty_data',
        question: '查询电话号码 138-0000-1234 的用户',
        description: '数据库中电话格式不一致，Rich Context 提供格式说明',
        expectedSql: "SELECT * FROM users WHERE phone LIKE '%13800001234%'",
        difficulty: 'medium'
      },
      {
        id: 'S1',
        name: '复杂Schema：同名列消歧',
        category: 'complex_schema',
        question: '查询每个部门的员工数量',
        description: '多个表都有 name 列，Rich Context 说明各表用途',
        expectedSql: 'SELECT d.name, COUNT(e.id) FROM departments d LEFT JOIN employees e ON d.id = e.dept_id GROUP BY d.id',
        difficulty: 'medium'
      },
      {
        id: 'S2',
        name: '复杂Schema：隐式关系',
        category: 'complex_schema',
        question: '查询张三的所有订单金额',
        description: '用户表和订单表通过中间表关联，Rich Context 提供关系说明',
        expectedSql: 'SELECT SUM(o.amount) FROM orders o JOIN user_orders uo ON o.id = uo.order_id JOIN users u ON uo.user_id = u.id WHERE u.name = "张三"',
        difficulty: 'hard'
      },
      {
        id: 'B1',
        name: '业务规则：状态枚举',
        category: 'business_rule',
        question: '查询所有有效订单',
        description: 'status=1 表示有效，Rich Context 提供业务规则说明',
        expectedSql: 'SELECT * FROM orders WHERE status = 1',
        difficulty: 'easy'
      },
      {
        id: 'B2',
        name: '业务规则：计算逻辑',
        category: 'business_rule',
        question: '查询每个商品的实际售价',
        description: '实际价格=原价*(1-折扣)，Rich Context 提供计算规则',
        expectedSql: 'SELECT name, price * (1 - discount) as actual_price FROM products',
        difficulty: 'medium'
      }
    ]
  },

  runComparison: async (caseId: string): Promise<ComparisonResult> => {
    await new Promise(r => setTimeout(r, 2000))

    const results: Record<string, ComparisonResult> = {
      'D1': {
        withContext: {
          sql: "SELECT * FROM customers WHERE state = 'California'",
          isCorrect: true,
          usedContexts: [
            {
              id: 'ctx-syn-1',
              databaseId: 'ecommerce',
              tableId: 'customers',
              tableName: 'customers',
              columnName: 'state',
              type: 'synonym',
              content: 'CA 是 California 的标准缩写',
              createdAt: '2024-01-01',
              source: 'auto'
            }
          ],
          duration: 1200,
          explanation: '通过 Rich Context 识别到 CA 是 California 的缩写'
        },
        withoutContext: {
          sql: "SELECT * FROM customers WHERE state = 'CA'",
          isCorrect: false,
          duration: 800,
          errorReason: '无法识别 CA 缩写，直接使用原始值导致查询结果为空'
        }
      },
      'B1': {
        withContext: {
          sql: 'SELECT * FROM orders WHERE status = 1',
          isCorrect: true,
          usedContexts: [
            {
              id: 'ctx-rule-1',
              databaseId: 'ecommerce',
              tableId: 'orders',
              tableName: 'orders',
              columnName: 'status',
              type: 'value_mapping',
              content: 'status 订单状态枚举: 0=待支付, 1=已支付(有效), 2=已发货, 3=已完成, 4=已取消',
              createdAt: '2024-01-01',
              source: 'manual'
            }
          ],
          duration: 1100,
          explanation: '通过 Rich Context 理解"有效订单"对应 status=1'
        },
        withoutContext: {
          sql: "SELECT * FROM orders WHERE status = 'active'",
          isCorrect: false,
          duration: 750,
          errorReason: '不知道"有效"对应的状态值，猜测使用字符串导致类型错误'
        }
      }
    }

    return results[caseId] ?? results['D1']!
  }
}

// Self-maintain API (Mock for demo)
export const selfMaintainApi = {
  getLogs: async (): Promise<MaintenanceLog[]> => {
    return [
      {
        id: 'log-1',
        type: 'error_feedback',
        trigger: 'SQL执行失败: column "stat" not found in table "orders"',
        action: '自动添加列名纠错映射: stat → status',
        status: 'verified',
        timestamp: '2024-01-15 10:30:00',
        contextAfter: {
          id: 'ctx-fix-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'synonym',
          content: 'stat 是 status 的常见拼写错误或缩写',
          createdAt: '2024-01-15T10:30:00Z',
          source: 'feedback'
        }
      },
      {
        id: 'log-2',
        type: 'user_correction',
        trigger: '用户反馈: "有效订单"应该是 status IN (1,2,3) 而不是 status=1',
        action: '更新业务规则 Context，扩展有效订单的定义',
        status: 'applied',
        timestamp: '2024-01-15 11:00:00',
        contextBefore: {
          id: 'ctx-old-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'business_rule',
          content: '"有效订单"指 status = 1 的订单',
          createdAt: '2024-01-10T00:00:00Z',
          source: 'manual'
        },
        contextAfter: {
          id: 'ctx-new-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'business_rule',
          content: '"有效订单"指 status IN (1, 2, 3) 的订单，即已支付、已发货、已完成状态',
          createdAt: '2024-01-15T11:00:00Z',
          source: 'feedback'
        }
      },
      {
        id: 'log-3',
        type: 'schema_change',
        trigger: '检测到 customers 表新增列: vip_expire_date',
        action: '自动生成列描述 Context',
        status: 'pending',
        timestamp: '2024-01-16 09:00:00'
      },
      {
        id: 'log-4',
        type: 'pattern_learning',
        trigger: '检测到高频查询模式: "最近N天的订单"',
        action: '学习时间范围查询模式，优化日期处理',
        status: 'analyzing',
        timestamp: '2024-01-16 14:00:00'
      }
    ]
  },

  triggerMaintenance: async (type: string, data?: any): Promise<MaintenanceLog> => {
    await new Promise(r => setTimeout(r, 1500))
    return {
      id: `log-${Date.now()}`,
      type: type as any,
      trigger: data?.trigger || '手动触发维护检查',
      action: '正在分析...',
      status: 'analyzing',
      timestamp: new Date().toISOString()
    }
  },

  applyMaintenance: async (logId: string): Promise<MaintenanceLog> => {
    await new Promise(r => setTimeout(r, 1000))
    return {
      id: logId,
      type: 'user_correction',
      trigger: '用户确认应用',
      action: '已应用更新',
      status: 'applied',
      timestamp: new Date().toISOString()
    }
  },

  rejectMaintenance: async (logId: string, reason?: string): Promise<void> => {
    await new Promise(r => setTimeout(r, 500))
  }
}
