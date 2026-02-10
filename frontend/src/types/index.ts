// API Response types
export interface ApiResponse<T> {
  success: boolean
  data: T
  error?: string
}

// ============================================
// Database & Connection Types
// ============================================

export interface Database {
  id: string
  name: string
  displayName: string
  type: 'mariadb' | 'sqlite' | 'mysql' | 'postgresql'
  host?: string
  port?: number
  status: 'connected' | 'disconnected' | 'error'
  tableCount: number
  hasRichContext: boolean
  contextCount: number
  lastConnected?: string
  description?: string
  tags?: string[]
  metadata?: {
    lakebaseId?: number
    [key: string]: any
  }
}

export interface DatabaseConfig {
  name: string
  type: 'mariadb' | 'sqlite' | 'mysql' | 'postgresql'
  host?: string
  port?: number
  username?: string
  password?: string
  database?: string
  path?: string // for sqlite
}

export interface DatabaseInfo {
  id: string
  name: string
  type: 'sqlite' | 'mysql' | 'postgresql' | 'mariadb'
  tables: TableInfo[]
}

export interface TableInfo {
  name: string
  columns: ColumnInfo[]
  rowCount?: number
  hasContext?: boolean
  description?: string
}

export interface ColumnInfo {
  name: string
  type: string
  isPrimaryKey?: boolean
  isForeignKey?: boolean
  isNullable?: boolean
  defaultValue?: string
  hasContext?: boolean
  references?: { table: string; column: string }
}

export interface SchemaInfo {
  databaseId?: string
  databaseName?: string
  tables: TableInfo[]
  relationships?: Relationship[]
  lastUpdated?: string
}

export interface Relationship {
  from: { table: string; column: string }
  to: { table: string; column: string }
  type: 'one-to-one' | 'one-to-many' | 'many-to-many'
}

// ============================================
// Rich Context Types
// ============================================

export type ContextType =
  | 'description'      // 字段/表描述
  | 'example'          // 示例值
  | 'constraint'       // 约束条件
  | 'synonym'          // 同义词/缩写
  | 'value_mapping'    // 枚举值映射
  | 'business_rule'    // 业务规则
  | 'calculation'      // 计算规则

export type ContextSource = 'auto' | 'manual' | 'feedback' | 'schema_change'

export interface RichContext {
  id: string
  databaseId: string
  tableId: string
  tableName: string
  columnName?: string
  type: ContextType
  content: string
  embedding?: number[]
  createdAt: string
  updatedAt?: string
  source: ContextSource
  confidence?: number
  usageCount?: number
}

export interface ContextFilter {
  databaseId?: string
  tableName?: string
  columnName?: string
  type?: ContextType
  source?: ContextSource
  search?: string
}

// ============================================
// Grounding Types
// ============================================

export interface GroundingResult {
  tables: GroundingTable[]
  columns: GroundingColumn[]
  joinPaths: JoinPath[]
  suggestedFields: SuggestedFieldFromLinking[]
  duration: number
  stage1Duration?: number
  stage2Duration?: number
  executionLogs?: ExecutionLog[] // SQL execution transparency
  reasoning?: string              // LLM reasoning for fine selection
  mode?: string                   // "sequential", "parallel", "coarse_only"
}

// SuggestedFieldFromLinking represents a field suggested by the linking agent
// (zero extra LLM cost — part of the schema linking step)
export interface SuggestedFieldFromLinking {
  tableName: string
  columnName: string
  reason: string
  selected: boolean
}

// ExecutionLog for grounding transparency
export interface ExecutionLog {
  phase: string        // "vector_search", "fine_selection"
  sql: string          // SQL query executed
  result_count: number // Number of results
  duration_ms: number  // Execution time in milliseconds
  summary: string      // Human-readable summary
}

export interface GroundingTable {
  name: string
  description?: string
  confidence: number
  matchedTerms: string[]
  contextUsed?: string[]
}

export interface GroundingColumn {
  table: string
  column: string
  dataType?: string
  description?: string
  confidence: number
  matchedTerms: string[]
  contextUsed?: string[]
}

export interface JoinPath {
  from: { table: string; column: string }
  to: { table: string; column: string }
  confidence?: number
}

// ============================================
// ReAct Types
// ============================================

export type ReActStepType = 'thought' | 'action' | 'observation' | 'answer' | 'error'

export interface ReActStep {
  step: number
  type: ReActStepType
  content: string
  thought?: string
  action?: string
  actionInput?: any
  observation?: string
  phase?: 'schema_linking' | 'sql_generation'
  timestamp?: number
  metadata?: Record<string, any>
}

// ============================================
// Text2SQL Types
// ============================================

export interface Text2SQLRequest {
  question: string
  databaseId: string
  database: string
  options: Text2SQLOptions
  fieldDescription?: string // Optional field alignment description
  injectedGrounding?: any   // Reuse previous grounding result for Phase 2
}

export interface Text2SQLOptions {
  useRichContext: boolean
  useReact: boolean
  useGrounding: boolean
  maxIterations: number
  temperature?: number
  model?: string
  groundingOnly?: boolean // When true, stop after grounding (for field alignment)
  skipGrounding?: boolean // When true, skip grounding and use injectedGrounding
}

export interface Text2SQLResult {
  sql: string
  executionResult?: any[]
  reactSteps: ReActStep[]
  groundingResult?: GroundingResult
  usedContexts?: RichContext[]
  duration: number
  error?: string
}

// ============================================
// Query & Execution Types
// ============================================

export interface QueryRecord {
  id: string
  databaseId: string
  question: string
  sql: string
  executionResult?: any[]
  isCorrect?: boolean
  feedback?: 'positive' | 'negative'
  feedbackNote?: string
  duration: number
  timestamp: string
  usedContexts?: RichContext[]
}

export interface ExecutionResult {
  columns: string[]
  rows: any[][]
  rowCount: number
  duration: number
  error?: string
}

// ============================================
// Comparison Types (Demo)
// ============================================

export type ComparisonCategory = 'dirty_data' | 'complex_schema' | 'business_rule'

export interface ComparisonCase {
  id: string
  name: string
  category: ComparisonCategory
  question: string
  description: string
  expectedSql: string
  hint?: string
  difficulty?: 'easy' | 'medium' | 'hard'
}

export interface ComparisonResult {
  withContext: {
    sql: string
    isCorrect: boolean
    executionResult?: any[]
    usedContexts: RichContext[]
    duration: number
    explanation?: string
  }
  withoutContext: {
    sql: string
    isCorrect: boolean
    executionResult?: any[]
    duration: number
    errorReason?: string
  }
}

// ============================================
// Self-Maintain Types
// ============================================

export type MaintenanceType =
  | 'error_feedback'    // SQL执行错误反馈
  | 'user_correction'   // 用户手动纠正
  | 'schema_change'     // Schema变更检测
  | 'pattern_learning'  // 查询模式学习

export type MaintenanceStatus =
  | 'pending'           // 待处理
  | 'analyzing'         // 分析中
  | 'applied'           // 已应用
  | 'verified'          // 已验证
  | 'rejected'          // 已拒绝

export interface MaintenanceLog {
  id: string
  type: MaintenanceType
  trigger: string
  action: string
  contextBefore?: RichContext
  contextAfter?: RichContext
  status: MaintenanceStatus
  timestamp: string
  verifiedBy?: string
}

export interface MaintenanceTrigger {
  type: MaintenanceType
  source: string
  data: any
}

// ============================================
// SSE Event Types
// ============================================

export type SSEEventType =
  | 'grounding_start'
  | 'grounding_stage1'
  | 'grounding_stage2'
  | 'grounding_complete'
  | 'retrieval_signal'
  | 'retrieval_complete'
  | 'linking_complete'
  | 'field_suggestions'
  | 'react_step'
  | 'context_retrieved'
  | 'sql_generated'
  | 'execution_start'
  | 'execution_complete'
  | 'complete'
  | 'error'

export interface SSEEvent<T = any> {
  type: SSEEventType
  data: T
  timestamp: number
}

// ============================================
// Workspace Types
// ============================================

export type WorkspaceTab = 'query' | 'schema' | 'context' | 'monitor'

export interface WorkspaceState {
  databaseId: string
  activeTab: WorkspaceTab
  queryHistory: QueryRecord[]
  selectedTable?: string
}

// ============================================
// UI Types
// ============================================

export interface Toast {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  title: string
  message?: string
  duration?: number
}

export interface ConfirmOptions {
  title: string
  content: string
  positiveText?: string
  negativeText?: string
  type?: 'info' | 'warning' | 'error'
}
