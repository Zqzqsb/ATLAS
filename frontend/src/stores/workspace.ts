import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type {
  Database,
  SchemaInfo,
  RichContext,
  QueryRecord,
  WorkspaceTab,
  Text2SQLOptions,
  ReActStep,
  GroundingResult,
  SSEEvent
} from '@/types'
import { databaseApi, contextApi, queryApi } from '@/api'
import { useDatabaseStore } from './database'

export const useWorkspaceStore = defineStore('workspace', () => {
  // Dependencies
  const databaseStore = useDatabaseStore()

  // State
  const currentDatabaseId = ref<string | null>(null)
  const activeTab = ref<WorkspaceTab>('query')
  const schemaCache = ref<SchemaInfo | null>(null)
  const contexts = ref<RichContext[]>([])
  const queryHistory = ref<QueryRecord[]>([])

  // Loading states
  const loadingSchema = ref(false)
  const loadingContexts = ref(false)
  const loadingHistory = ref(false)

  // Query state
  const currentQuestion = ref('')
  const isQuerying = ref(false)
  const queryError = ref<string | null>(null)

  // Query options
  const queryOptions = ref<Text2SQLOptions>({
    useRichContext: true,
    useReact: true,
    useGrounding: true,
    maxIterations: 5
  })

  // Query results
  const generatedSql = ref('')
  const reactSteps = ref<ReActStep[]>([])
  const groundingResult = ref<GroundingResult | null>(null)
  const groundingStage = ref<'idle' | 'stage1' | 'stage2' | 'done'>('idle')
  const usedContexts = ref<RichContext[]>([])
  const queryDuration = ref(0)
  const executionResult = ref<any[] | null>(null)

  // Abort controller for streaming
  let abortQuery: (() => void) | null = null

  // Computed
  const currentDatabase = computed<Database | null>(() => {
    if (!currentDatabaseId.value) return null
    return databaseStore.getDatabaseById(currentDatabaseId.value) || null
  })

  const tableNames = computed(() =>
    schemaCache.value?.tables.map(t => t.name) || []
  )

  const contextsByTable = computed(() => {
    const map: Record<string, RichContext[]> = {}
    for (const ctx of contexts.value) {
      if (!map[ctx.tableName]) {
        map[ctx.tableName] = []
      }
      map[ctx.tableName]!.push(ctx)
    }
    return map
  })

  const hasRichContext = computed(() => contexts.value.length > 0)

  // Actions
  async function selectDatabase(id: string) {
    if (currentDatabaseId.value === id) return

    // Reset state
    resetQueryState()
    schemaCache.value = null
    contexts.value = []
    queryHistory.value = []

    currentDatabaseId.value = id

    // Load data in parallel
    await Promise.all([
      fetchSchema(),
      fetchContexts(),
      fetchQueryHistory()
    ])
  }

  async function fetchSchema() {
    if (!currentDatabaseId.value || !currentDatabase.value) return

    loadingSchema.value = true
    try {
      // Get lakebase numeric ID
      const lakebaseId = currentDatabase.value.metadata?.lakebaseId || currentDatabaseId.value
      
      // Use lakebase API to get schema from rc_tables and rc_columns
      const response = await fetch(`/api/v1/lakebase/datasources/${lakebaseId}`)
      const data = await response.json()
      
      if (data.tables && data.columns) {
        // Transform to SchemaInfo format
        schemaCache.value = {
          databaseId: currentDatabaseId.value,
          databaseName: currentDatabase.value.name,
          tables: data.tables.map((t: any) => ({
            name: t.table_name,
            comment: t.description || '',
            columns: data.columns
              .filter((c: any) => c.table_name === t.table_name)
              .map((c: any) => ({
                name: c.column_name,
                type: c.data_type || 'VARCHAR',
                nullable: true,
                comment: c.description || ''
              }))
          }))
        }
      }
    } catch (e: any) {
      console.error('Failed to fetch schema:', e)
    } finally {
      loadingSchema.value = false
    }
  }

  async function fetchContexts() {
    if (!currentDatabaseId.value || !currentDatabase.value) return

    loadingContexts.value = true
    try {
      // Get lakebase numeric ID
      const lakebaseId = currentDatabase.value.metadata?.lakebaseId || currentDatabaseId.value
      
      // Use lakebase API to get contexts from rc_tables and rc_columns
      const response = await fetch(`/api/v1/lakebase/datasources/${lakebaseId}`)
      const data = await response.json()
      
      if (data.contexts) {
        contexts.value = data.contexts.map((ctx: any) => ({
          id: String(ctx.id),
          databaseId: currentDatabaseId.value!,
          tableName: ctx.table_name,
          columnName: ctx.column_name,
          type: ctx.context_type || 'description',
          content: ctx.content,
          createdAt: ctx.created_at,
          usageCount: 0
        }))
      }
    } catch (e: any) {
      console.error('Failed to fetch contexts:', e)
    } finally {
      loadingContexts.value = false
    }
  }

  async function fetchQueryHistory() {
    if (!currentDatabaseId.value) return

    loadingHistory.value = true
    try {
      queryHistory.value = await queryApi.getHistory(currentDatabaseId.value)
    } catch (e: any) {
      console.error('Failed to fetch query history:', e)
    } finally {
      loadingHistory.value = false
    }
  }

  function resetQueryState() {
    generatedSql.value = ''
    reactSteps.value = []
    groundingResult.value = null
    groundingStage.value = 'idle'
    usedContexts.value = []
    queryDuration.value = 0
    executionResult.value = null
    queryError.value = null
    isQuerying.value = false
  }

  async function executeQuery(question?: string) {
    if (!currentDatabaseId.value || !currentDatabase.value) return

    const q = question || currentQuestion.value
    if (!q.trim()) return

    // Abort previous query if any
    abortCurrentQuery()

    // Reset state
    resetQueryState()
    currentQuestion.value = q
    isQuerying.value = true
    const startTime = Date.now()

    abortQuery = queryApi.stream(
      {
        question: q,
        databaseId: currentDatabaseId.value,
        database: currentDatabaseId.value, // Use ID instead of name
        options: queryOptions.value
      },
      (event: { type: string; data: any }) => {
        console.log('SSE Event:', event.type, event.data) // Debug log
        
        switch (event.type) {
          case 'thought':
          case 'action':
          case 'observation':
          case 'finish':
            // ReAct step events
            if (event.data.step !== undefined) {
              const existingIndex = reactSteps.value.findIndex(s => s.step === event.data.step)
              if (existingIndex >= 0) {
                // Update existing step
                reactSteps.value[existingIndex] = {
                  ...reactSteps.value[existingIndex],
                  ...event.data
                }
              } else {
                // Add new step
                reactSteps.value.push(event.data)
              }
            }
            break
          case 'grounding_start':
            groundingStage.value = 'stage1'
            break
          case 'grounding_stage1':
            groundingStage.value = 'stage1'
            break
          case 'grounding_stage2':
            groundingStage.value = 'stage2'
            break
          case 'grounding_complete':
            groundingStage.value = 'done'
            groundingResult.value = event.data
            break
          case 'context_retrieved':
            usedContexts.value = event.data.contexts || []
            break
          case 'sql_generated':
            generatedSql.value = event.data.sql || ''
            break
          case 'execution_complete':
            executionResult.value = event.data.result || []
            break
          case 'complete':
            generatedSql.value = event.data.sql || event.data.final_sql || generatedSql.value
            if (event.data.execution_result) {
              executionResult.value = event.data.execution_result.rows || []
            }
            queryDuration.value = Date.now() - startTime
            isQuerying.value = false

            // Add to history
            if (generatedSql.value) {
              const record: QueryRecord = {
                id: `q-${Date.now()}`,
                databaseId: currentDatabaseId.value!,
                question: q,
                sql: generatedSql.value,
                duration: queryDuration.value / 1000,
                timestamp: new Date().toISOString(),
                usedContexts: usedContexts.value
              }
              queryHistory.value.unshift(record)
            }
            break
          case 'error':
            queryError.value = event.data.message || event.data.error || 'Unknown error'
            isQuerying.value = false
            break
        }
      },
      (error) => {
        queryError.value = error.message
        isQuerying.value = false
      },
      () => {
        // Stream completed
        if (isQuerying.value) {
          isQuerying.value = false
        }
      }
    )
  }

  function abortCurrentQuery() {
    if (abortQuery) {
      abortQuery()
      abortQuery = null
    }
    isQuerying.value = false
  }

  async function addContext(context: Omit<RichContext, 'id' | 'createdAt' | 'usageCount'>) {
    try {
      const newContext = await contextApi.create(context)
      contexts.value.push(newContext)
      return newContext
    } catch (e: any) {
      console.error('Failed to add context:', e)
      return null
    }
  }

  async function updateContext(contextId: string, updates: Partial<RichContext>) {
    if (!currentDatabaseId.value) return null

    try {
      const updated = await contextApi.update(currentDatabaseId.value, contextId, updates)
      const index = contexts.value.findIndex(c => c.id === contextId)
      if (index >= 0) {
        contexts.value[index] = updated
      }
      return updated
    } catch (e: any) {
      console.error('Failed to update context:', e)
      return null
    }
  }

  async function deleteContext(contextId: string) {
    if (!currentDatabaseId.value) return false

    try {
      await contextApi.delete(currentDatabaseId.value, contextId)
      const index = contexts.value.findIndex(c => c.id === contextId)
      if (index >= 0) {
        contexts.value.splice(index, 1)
      }
      return true
    } catch (e: any) {
      console.error('Failed to delete context:', e)
      return false
    }
  }

  function setActiveTab(tab: WorkspaceTab) {
    activeTab.value = tab
  }

  return {
    // State
    currentDatabaseId,
    activeTab,
    schemaCache,
    contexts,
    queryHistory,
    loadingSchema,
    loadingContexts,
    loadingHistory,
    currentQuestion,
    isQuerying,
    queryError,
    queryOptions,
    generatedSql,
    reactSteps,
    groundingResult,
    groundingStage,
    usedContexts,
    queryDuration,
    executionResult,

    // Computed
    currentDatabase,
    tableNames,
    contextsByTable,
    hasRichContext,

    // Actions
    selectDatabase,
    fetchSchema,
    fetchContexts,
    fetchQueryHistory,
    resetQueryState,
    executeQuery,
    abortCurrentQuery,
    addContext,
    updateContext,
    deleteContext,
    setActiveTab
  }
})
