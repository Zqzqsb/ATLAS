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

// Warmup API - pre-loads caches on the backend to reduce first-query latency
const warmupApi = {
  warmup: async (databaseId: string) => {
    try {
      await fetch('/api/v1/text2sql/warmup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ database_id: databaseId })
      })
    } catch (e) {
      // Silently ignore warmup failures - it's a best-effort optimization
      console.debug('Warmup failed (non-critical):', e)
    }
  }
}

export const useWorkspaceStore = defineStore('workspace', () => {
  // Dependencies
  const databaseStore = useDatabaseStore()

  // State
  const currentDatabaseId = ref<string | null>(null)
  const activeTab = ref<WorkspaceTab>('query')
  const schemaCache = ref<SchemaInfo | null>(null)
  const contexts = ref<RichContext[]>([])
  const relations = ref<Array<{
    id: number
    fromTable: string
    fromColumn: string
    toTable: string
    toColumn: string
    relationType: string
    description: string
  }>>([])
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

  // Skeleton state: immediately show schema info while waiting for backend
  const skeletonTables = ref<string[]>([])
  const showSkeleton = ref(false)

  // Grounding sub-stage progress for SSE streaming (linking reasoning, retrieval done, etc.)
  const groundingProgress = ref<{ stage: string; data: any } | null>(null)

  // Abort controller for streaming
  let abortQuery: (() => void) | null = null

  // Transform backend grounding result to frontend format
  function transformGroundingResult(data: any): GroundingResult | null {
    if (!data) return null

    return {
      tables: (data.tables || []).map((t: any) => ({
        name: t.name,
        confidence: t.confidence || 0,
        matchedTerms: t.reason ? [t.reason] : [],
        contextUsed: []
      })),
      columns: (data.columns || []).map((c: any) => ({
        table: c.table_name || c.table,
        column: c.column_name || c.column,
        confidence: c.confidence || 0,
        matchedTerms: c.reason ? [c.reason] : [],
        contextUsed: []
      })),
      joinPaths: (data.join_paths || []).map((jp: any) => ({
        from: { table: jp.from_table, column: jp.from_column },
        to: { table: jp.to_table, column: jp.to_column },
        confidence: jp.confidence
      })),
      duration: data.execution_time_ms || 0,
      // Execution logs for transparency
      executionLogs: (data.execution_logs || []).map((log: any) => ({
        phase: log.phase,
        sql: log.sql,
        result_count: log.result_count,
        duration_ms: log.duration_ms,
        summary: log.summary
      })),
      // LLM reasoning for fine selection
      reasoning: data.reasoning || '',
      mode: data.mode || '',
      // Suggested fields from linking agent (zero extra LLM cost)
      suggestedFields: (data.suggested_fields || []).map((f: any) => ({
        tableName: f.table_name,
        columnName: f.column_name,
        reason: f.reason || '',
        selected: f.selected !== false
      }))
    }
  }

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

    // Load data in parallel, plus trigger backend warmup
    await Promise.all([
      fetchSchema(),
      fetchContexts(),
      fetchQueryHistory(),
      warmupApi.warmup(id) // Pre-load backend caches
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
            description: t.description || '',
            rowCount: t.row_count,
            hasContext: !!t.description,
            columns: data.columns
              .filter((c: any) => c.table_name === t.table_name)
              .map((c: any) => ({
                name: c.column_name,
                type: c.data_type || 'VARCHAR',
                isPrimaryKey: c.is_pk,
                isForeignKey: c.is_fk,
                isNullable: c.is_nullable,
                hasContext: !!c.description,
                comment: c.description || ''
              }))
          }))
        }

        // Save relations
        if (data.relations) {
          relations.value = data.relations.map((r: any) => ({
            id: r.id,
            fromTable: r.from_table,
            fromColumn: r.from_column,
            toTable: r.to_table,
            toColumn: r.to_column,
            relationType: r.relation_type,
            description: r.description || ''
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
    groundingProgress.value = null
    usedContexts.value = []
    queryDuration.value = 0
    executionResult.value = null
    queryError.value = null
    isQuerying.value = false
    skeletonTables.value = []
    showSkeleton.value = false
  }

  // Reset only inference state (preserve grounding results for Phase 2)
  function resetInferenceState() {
    generatedSql.value = ''
    reactSteps.value = []
    usedContexts.value = []
    queryDuration.value = 0
    executionResult.value = null
    queryError.value = null
    // NOTE: groundingResult, groundingStage, skeletonTables are preserved
  }

  async function executeQuery(question?: string, fieldDescription?: string, groundingOnly?: boolean, injectedGrounding?: any) {
    if (!currentDatabaseId.value || !currentDatabase.value) return

    const q = question || currentQuestion.value
    if (!q.trim()) return

    // Abort previous query if any
    abortCurrentQuery()

    // If skip_grounding (Phase 2), only reset inference state — preserve grounding result
    const skipGrounding = !!injectedGrounding
    if (skipGrounding) {
      resetInferenceState()
    } else {
      resetQueryState()
    }
    currentQuestion.value = q
    isQuerying.value = true
    const startTime = Date.now()

    // Skeleton screen: immediately show table names from local schemaCache
    // so the user sees progress before the backend SSE starts streaming
    if (!skipGrounding && schemaCache.value && schemaCache.value.tables.length > 0) {
      skeletonTables.value = schemaCache.value.tables.map(t => t.name)
      showSkeleton.value = true
    }

    // Also fire warmup in parallel (non-blocking) to ensure backend caches are hot
    if (currentDatabaseId.value) {
      warmupApi.warmup(currentDatabaseId.value)
    }

    abortQuery = queryApi.stream(
      {
        question: q,
        databaseId: currentDatabaseId.value,
        database: currentDatabaseId.value, // Use ID instead of name
        options: {
          ...queryOptions.value,
          groundingOnly: groundingOnly || false,
          skipGrounding: skipGrounding
        },
        fieldDescription: fieldDescription,
        injectedGrounding: injectedGrounding
      },
      (event: { type: string; data: any }) => {
        console.log('SSE Event:', event.type, event.data) // Debug log

        switch (event.type) {
          case 'thought':
          case 'action':
          case 'observation':
          case 'finish':
            // ReAct step events - use step + phase as unique key
            if (event.data.step !== undefined) {
              const phase = event.data.phase || 'unknown'
              const stepKey = `${phase}-${event.data.step}`
              const existingIndex = reactSteps.value.findIndex(
                s => `${s.phase || 'unknown'}-${s.step}` === stepKey
              )

              // Create step data with unique ID
              const stepData = {
                ...event.data,
                id: stepKey,
                type: event.type as any
              }

              if (existingIndex >= 0) {
                // Update existing step (merge new data)
                reactSteps.value[existingIndex] = {
                  ...reactSteps.value[existingIndex],
                  ...stepData
                }
              } else {
                // Add new step
                reactSteps.value.push(stepData)
              }
            }
            break
          case 'grounding_start':
            groundingStage.value = 'stage1'
            groundingProgress.value = null
            // Hide skeleton once real grounding data starts arriving
            showSkeleton.value = false
            break
          case 'grounding_progress': {
            // Sub-stage events from adaptive pipeline: retrieval_start, retrieval_done, linking_start, linking_done
            const stage = event.data?.stage
            groundingProgress.value = { stage, data: event.data?.data || event.data }
            // Map sub-stages to groundingStage for UI
            if (stage === 'linking_start' || stage === 'linking_done') {
              groundingStage.value = 'stage2'
            }
            break
          }
          case 'grounding_stage1':
            groundingStage.value = 'stage1'
            break
          case 'grounding_stage2':
            groundingStage.value = 'stage2'
            break
          case 'grounding_complete':
            groundingStage.value = 'done'
            // Transform backend format to frontend format
            groundingResult.value = transformGroundingResult(event.data)
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
            showSkeleton.value = false // Ensure skeleton is hidden on complete

            // If this was a grounding-only request, don't add to history
            if (event.data.grounding_only) {
              break
            }

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
            showSkeleton.value = false // Ensure skeleton is hidden on error
            break
        }
      },
      (error) => {
        queryError.value = error.message
        isQuerying.value = false
        showSkeleton.value = false // Ensure skeleton is hidden on error
      },
      () => {
        // Stream completed
        if (isQuerying.value) {
          isQuerying.value = false
          showSkeleton.value = false // Ensure skeleton is hidden on completion
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
    relations,
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
    groundingProgress,
    skeletonTables,
    showSkeleton,

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
