// Re-export all API modules
export { default as client } from './client'
export { createSSEStream } from './client'
export { databaseApi } from './database'
export { contextApi } from './context'
export { queryApi } from './query'
export { agentApi } from './agent'
export type { AgentStatus, MaintenanceResult, ChangeLog, ChangeLogSummary, SchemaChange } from './agent'
export { evolutionApi } from './evolution'
export type { EvolutionStatus, EvolutionStage, StageExecution, ContextAction, EvolutionEvent } from './evolution'

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
        name: 'Dirty Data: Abbreviation Recognition',
        category: 'dirty_data',
        question: 'Find all customers in CA state',
        description: 'CA is an abbreviation for California; Rich Context provides abbreviation mapping',
        expectedSql: "SELECT * FROM customers WHERE state = 'California'",
        difficulty: 'easy'
      },
      {
        id: 'D2',
        name: 'Dirty Data: Format Variants',
        category: 'dirty_data',
        question: 'Find the user with phone number 138-0000-1234',
        description: 'Phone formats are inconsistent in the database; Rich Context provides format specifications',
        expectedSql: "SELECT * FROM users WHERE phone LIKE '%13800001234%'",
        difficulty: 'medium'
      },
      {
        id: 'S1',
        name: 'Complex Schema: Same-Name Column Disambiguation',
        category: 'complex_schema',
        question: 'Count the number of employees in each department',
        description: 'Multiple tables have a "name" column; Rich Context clarifies each table\'s purpose',
        expectedSql: 'SELECT d.name, COUNT(e.id) FROM departments d LEFT JOIN employees e ON d.id = e.dept_id GROUP BY d.id',
        difficulty: 'medium'
      },
      {
        id: 'S2',
        name: 'Complex Schema: Implicit Relationships',
        category: 'complex_schema',
        question: 'Find the total order amount for user Alice',
        description: 'Users and orders are linked through a junction table; Rich Context provides relationship details',
        expectedSql: 'SELECT SUM(o.amount) FROM orders o JOIN user_orders uo ON o.id = uo.order_id JOIN users u ON uo.user_id = u.id WHERE u.name = "Alice"',
        difficulty: 'hard'
      },
      {
        id: 'B1',
        name: 'Business Rule: Status Enum',
        category: 'business_rule',
        question: 'Find all active orders',
        description: 'status=1 means active; Rich Context provides business rule explanations',
        expectedSql: 'SELECT * FROM orders WHERE status = 1',
        difficulty: 'easy'
      },
      {
        id: 'B2',
        name: 'Business Rule: Calculation Logic',
        category: 'business_rule',
        question: 'Find the actual selling price of each product',
        description: 'Actual price = original price * (1 - discount); Rich Context provides calculation rules',
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
              content: 'CA is the standard abbreviation for California',
              createdAt: '2024-01-01',
              source: 'auto'
            }
          ],
          duration: 1200,
          explanation: 'Rich Context identified CA as an abbreviation for California'
        },
        withoutContext: {
          sql: "SELECT * FROM customers WHERE state = 'CA'",
          isCorrect: false,
          duration: 800,
          errorReason: 'Failed to resolve CA abbreviation, used raw value leading to empty results'
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
              content: 'status order status enum: 0=Pending, 1=Paid (active), 2=Shipped, 3=Completed, 4=Cancelled',
              createdAt: '2024-01-01',
              source: 'manual'
            }
          ],
          duration: 1100,
          explanation: 'Rich Context mapped "active orders" to status=1'
        },
        withoutContext: {
          sql: "SELECT * FROM orders WHERE status = 'active'",
          isCorrect: false,
          duration: 750,
          errorReason: 'Unknown status value for "active", guessed a string value causing type mismatch'
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
        trigger: 'SQL execution failed: column "stat" not found in table "orders"',
        action: 'Auto-added column name correction mapping: stat -> status',
        status: 'verified',
        timestamp: '2024-01-15 10:30:00',
        contextAfter: {
          id: 'ctx-fix-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'synonym',
          content: '"stat" is a common typo or abbreviation for "status"',
          createdAt: '2024-01-15T10:30:00Z',
          source: 'feedback'
        }
      },
      {
        id: 'log-2',
        type: 'user_correction',
        trigger: 'User feedback: "active orders" should be status IN (1,2,3) not status=1',
        action: 'Updated business rule context, expanded the definition of active orders',
        status: 'applied',
        timestamp: '2024-01-15 11:00:00',
        contextBefore: {
          id: 'ctx-old-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'business_rule',
          content: '"Active orders" means status = 1',
          createdAt: '2024-01-10T00:00:00Z',
          source: 'manual'
        },
        contextAfter: {
          id: 'ctx-new-1',
          databaseId: 'ecommerce',
          tableId: 'orders',
          tableName: 'orders',
          type: 'business_rule',
          content: '"Active orders" means status IN (1, 2, 3) — Paid, Shipped, and Completed',
          createdAt: '2024-01-15T11:00:00Z',
          source: 'feedback'
        }
      },
      {
        id: 'log-3',
        type: 'schema_change',
        trigger: 'Detected new column in customers table: vip_expire_date',
        action: 'Auto-generated column description context',
        status: 'pending',
        timestamp: '2024-01-16 09:00:00'
      },
      {
        id: 'log-4',
        type: 'pattern_learning',
        trigger: 'Detected frequent query pattern: "orders in the last N days"',
        action: 'Learned time-range query pattern, optimized date handling',
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
      trigger: data?.trigger || 'Manually triggered maintenance check',
      action: 'Analyzing...',
      status: 'analyzing',
      timestamp: new Date().toISOString()
    }
  },

  applyMaintenance: async (logId: string): Promise<MaintenanceLog> => {
    await new Promise(r => setTimeout(r, 1000))
    return {
      id: logId,
      type: 'user_correction',
      trigger: 'User confirmed apply',
      action: 'Update applied',
      status: 'applied',
      timestamp: new Date().toISOString()
    }
  },

  rejectMaintenance: async (logId: string, reason?: string): Promise<void> => {
    await new Promise(r => setTimeout(r, 500))
  }
}
