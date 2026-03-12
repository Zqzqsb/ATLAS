<script setup lang="ts">
import { ref, computed, watch, onMounted, Transition, TransitionGroup } from 'vue'
import { NButton, NInput, NInputNumber, NSwitch, NSelect, NCollapse, NCollapseItem, NCheckbox, NSpin, NTag, useMessage } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { queryApi } from '@/api/query'
import type { SuggestedFieldFromLinking } from '@/types'
import { apiClient } from '@/api/client'
import QueryResult from './QueryResult.vue'
import RealtimeCard from './RealtimeCard.vue'

const workspaceStore = useWorkspaceStore()
const message = useMessage()

// Query input
const question = ref('')

// Use store's isQuerying for execution state
const isExecuting = computed(() => workspaceStore.isQuerying)

// Field Alignment state
// Now populated from grounding result (zero extra LLM cost) instead of a separate API call
const suggestedFields = ref<SuggestedFieldFromLinking[]>([])
const showFieldPanel = ref(false)
const awaitingFieldConfirmation = ref(false) // True when grounding-only is done, waiting for user
const fieldPanelConsumed = ref(false) // True after user confirms/dismisses — prevents watcher re-trigger

// Stage timing
const stageTimings = ref({
  vectorSearch: { start: 0, end: 0 },
  schemaLinking: { start: 0, end: 0 },
  sqlGeneration: { start: 0, end: 0 }
})

// Watch grounding stage changes to track timing for progressive stages
watch(() => workspaceStore.groundingStage, (newStage, oldStage) => {
  if (newStage === 'stage1' && oldStage === 'idle') {
    stageTimings.value.vectorSearch.start = Date.now()
  } else if (newStage === 'retrieval_done') {
    // retrieval_complete SSE arrived → Vector search is done
    stageTimings.value.vectorSearch.end = Date.now()
  } else if (newStage === 'stage2') {
    // linking_complete arrived → Schema linking is done
    if (!stageTimings.value.vectorSearch.end) {
      stageTimings.value.vectorSearch.end = Date.now()
    }
    // Fallback: if linking_start wasn't received, mark start now
    if (!stageTimings.value.schemaLinking.start) {
      stageTimings.value.schemaLinking.start = Date.now()
    }
    // Mark end immediately — the actual duration comes from backend linkingDurationMs
    stageTimings.value.schemaLinking.end = Date.now()
  } else if (newStage === 'done') {
    if (!stageTimings.value.vectorSearch.end) {
      stageTimings.value.vectorSearch.end = Date.now()
    }
    // Mark schema linking end if it started
    if (stageTimings.value.schemaLinking.start && !stageTimings.value.schemaLinking.end) {
      stageTimings.value.schemaLinking.end = Date.now()
    }
  }
})

// Watch grounding progress for linking_start to capture the real start time
watch(() => workspaceStore.groundingProgress, (progress) => {
  if (progress?.stage === 'linking_start' && !stageTimings.value.schemaLinking.start) {
    stageTimings.value.schemaLinking.start = Date.now()
  }
})

// Watch react steps to track schema linking and sql generation timing
watch(() => workspaceStore.reactSteps, (steps) => {
  const schemaSteps = steps.filter(s => s.phase === 'schema_linking')
  const sqlSteps = steps.filter(s => s.phase === 'sql_generation')
  
  if (schemaSteps.length > 0 && !stageTimings.value.schemaLinking.start) {
    stageTimings.value.schemaLinking.start = Date.now()
  }
  
  if (sqlSteps.length > 0) {
    if (!stageTimings.value.schemaLinking.end) {
      stageTimings.value.schemaLinking.end = Date.now()
    }
    if (!stageTimings.value.sqlGeneration.start) {
      stageTimings.value.sqlGeneration.start = Date.now()
    }
  }
}, { deep: true })

// Watch query completion
watch(() => workspaceStore.generatedSql, (sql) => {
  if (sql && !stageTimings.value.sqlGeneration.end) {
    stageTimings.value.sqlGeneration.end = Date.now()
  }
})

function resetTimings() {
  stageTimings.value = {
    vectorSearch: { start: 0, end: 0 },
    schemaLinking: { start: 0, end: 0 },
    sqlGeneration: { start: 0, end: 0 }
  }
}

// Query options
const maxIterations = ref(5)
const useFieldAlignment = ref(false)
const selectedModel = ref('deepseek_v3')
const useRichContext = ref(true)
const useReact = ref(true)
const useGrounding = ref(true)
const linkingMode = ref<'off' | 'one-shot' | 'react'>('one-shot')

// Linking mode options for the NSelect
const linkingModeOptions = [
  { label: 'Off', value: 'off' },
  { label: 'One-Shot', value: 'one-shot' },
  { label: 'ReAct', value: 'react' }
]

// Model options - loaded from backend /models API
const modelOptions = ref<{ label: string; value: string }[]>([
  { label: 'DeepSeek V3', value: 'deepseek_v3' } // fallback
])

async function loadModels() {
  try {
    const resp = await apiClient.get<{ models: { id: string; name: string; is_default: boolean }[] }>('/models')
    const models = resp.data?.models
    if (models && models.length > 0) {
      modelOptions.value = models.map(m => ({ label: m.name, value: m.id }))
      // Select the default model if current selection is not in the list
      const defaultModel = models.find(m => m.is_default)
      const ids = models.map(m => m.id)
      if (!ids.includes(selectedModel.value) && defaultModel) {
        selectedModel.value = defaultModel.id
      }
    }
  } catch (e) {
    console.warn('Failed to load models from backend, using fallback', e)
  }
}

onMounted(() => {
  loadModels()
})

// Example questions for different Spider databases
const exampleQuestions = computed(() => {
  const dbName = workspaceStore.currentDatabase?.name?.toLowerCase() || ''
  
  // TV Show database
  if (dbName.includes('tvshow') || dbName.includes('tv_show')) {
    return [
      'List all TV channels',
      'Which channel has the highest package price?',
      'Show all cartoons and their broadcast channels',
      'Count the number of channels per country'
    ]
  }
  
  // Flight database
  if (dbName.includes('flight')) {
    return [
      'List all airlines',
      'What flights depart from Los Angeles?',
      'Show all airports and their cities',
      'Which airline has the most flights?'
    ]
  }
  
  // WTA Tennis database
  if (dbName.includes('wta')) {
    return [
      'List all player information',
      'Which player has won the most matches?',
      'Show all matches in 2016',
      'Count the number of players by country'
    ]
  }

  // TPC-H Enterprise (38-table large-scale demo)
  if (dbName.includes('tpch') || dbName.includes('enterprise')) {
    return [
      'Which supplier has the highest profit?',
      'Which parts in the East warehouse are low on stock?',
      'What are the lowest-rated products?',
      'Which orders have shipping delays over 3 days?',
      'Which parts have a price increase over 10%?',
    ]
  }

  // Default examples
  return [
    'List all tables',
    'Count the number of records in each table',
    'Find data containing a specific keyword'
  ]
})

// Execution stages
const hasGroundingContent = computed((): boolean => {
  const r = workspaceStore.groundingResult
  if (!r) return false
  return (r.tables?.length ?? 0) > 0 ||
    (r.columns?.length ?? 0) > 0 ||
    (r.joinPaths?.length ?? 0) > 0
})

const vectorSearchStage = computed(() => {
  const { start, end } = stageTimings.value.vectorSearch
  // Vector search is done when we have retrieval data (tables/columns), even if grounding isn't fully complete
  const hasRetrievalData = workspaceStore.groundingResult && 
    ((workspaceStore.groundingResult.tables?.length ?? 0) > 0 || (workspaceStore.groundingResult.columns?.length ?? 0) > 0)
  const stageDone = hasRetrievalData || workspaceStore.groundingStage === 'retrieval_done' || workspaceStore.groundingStage === 'done'
  // Prefer backend-reported retrieval latency T0→T1 (from linking_complete event).
  // Fall back to retrievalDurationMs (from retrieval_complete), then client-side timestamps.
  const backendLatency = workspaceStore.groundingResult?.retrievalLatencyMs
  const backendDuration = workspaceStore.groundingResult?.retrievalDurationMs
  const clientDuration = end && start ? end - start : 0
  return {
    active: isExecuting.value && workspaceStore.groundingStage !== 'idle',
    completed: stageDone && hasGroundingContent.value,
    empty: stageDone && !hasGroundingContent.value,
    data: workspaceStore.groundingResult,
    // For small_scale: no vector retrieval → use client-side round-trip time
    // For large_scale: prefer backend-reported retrieval latency T0→T1
    duration: workspaceStore.groundingResult?.strategy === 'small_scale'
      ? clientDuration
      : backendLatency != null && backendLatency > 0 ? backendLatency
        : backendDuration != null && backendDuration > 0 ? backendDuration
        : clientDuration
  }
})

// Detect small-scale strategy (no vector search, schema passed directly)
const isSmallScale = computed(() => workspaceStore.groundingResult?.strategy === 'small_scale')

// Group linking results by table for collapsible display
const linkedTableGroups = computed(() => {
  const tables = workspaceStore.groundingResult?.linkingTables || []
  const columns = workspaceStore.groundingResult?.linkingColumns || []
  if (!tables.length) return []

  // Build column lookup by table name
  const colsByTable = new Map<string, typeof columns>()
  for (const c of columns) {
    const list = colsByTable.get(c.table) || []
    list.push(c)
    colsByTable.set(c.table, list)
  }

  return tables.map(t => ({
    name: t.name,
    description: t.description || '',
    confidence: t.confidence || 0,
    hint: t.hint || '',
    columns: colsByTable.get(t.name) || []
  }))
})

const schemaLinkingStage = computed(() => {
  const { start, end } = stageTimings.value.schemaLinking
  const steps = workspaceStore.reactSteps.filter(s => s.phase === 'schema_linking')
  const hasSchemaLinkingSteps = steps.length > 0
  const hasSqlGenerationSteps = workspaceStore.reactSteps.some(s => s.phase === 'sql_generation')
  const completed = end > 0 || hasSqlGenerationSteps || !!workspaceStore.generatedSql
  
  // Active when: vector search is visually done AND (linking started/done OR has linking data)
  // Gate on vectorSearchStage to prevent Schema Linking from activating before Vector Search completes
  const vectorDone = vectorSearchStage.value.completed || vectorSearchStage.value.empty
  const groundingDone = workspaceStore.groundingStage === 'done' || workspaceStore.groundingStage === 'stage2'
  const linkingInProgress = workspaceStore.groundingProgress?.stage === 'linking_start'
  const hasLinkingData = !!workspaceStore.groundingResult?.reasoning || !!workspaceStore.groundingResult?.linkingTables?.length
  const active = awaitingFieldConfirmation.value || (vectorDone && (hasLinkingData || linkingInProgress || (isExecuting.value && groundingDone && !hasSqlGenerationSteps)))
  
  // Separate polling steps (get_candidate_schema with no result) from real analysis steps
  // Also filter out get_candidate_schema calls that returned schema data (successful fetch)
  // — those are infrastructure steps, not analysis steps worth showing.
  const pollingSteps = steps.filter(s => s.action === 'get_candidate_schema' && !s.observation)
  const schemaReceivedSteps = steps.filter(s => s.action === 'get_candidate_schema' && !!s.observation)
  const realSteps = steps.filter(s => s.action !== 'get_candidate_schema')

  return {
    active: active || (isExecuting.value && hasSchemaLinkingSteps),
    completed: completed && hasSchemaLinkingSteps,
    steps: realSteps,
    pollingCount: pollingSteps.length,
    schemaReceived: schemaReceivedSteps.length > 0,
    isPolling: pollingSteps.length > 0 && realSteps.length === 0 && isExecuting.value,
    contexts: workspaceStore.usedContexts,
    // Prefer backend-reported reasoning latency T1.1→T2 (from linking_complete event).
    // Fall back to linkingDurationMs, then client-side timestamps.
    duration: workspaceStore.groundingResult?.reasoningLatencyMs
      ? workspaceStore.groundingResult.reasoningLatencyMs
      : workspaceStore.groundingResult?.linkingDurationMs
      ? workspaceStore.groundingResult.linkingDurationMs
      : (completed && start && end ? end - start : 0)
  }
})

const sqlGenerationStage = computed(() => {
  const { start, end } = stageTimings.value.sqlGeneration
  const steps = workspaceStore.reactSteps.filter(s => s.phase === 'sql_generation')
  const hasSqlGenerationSteps = steps.length > 0
  const completed = !!workspaceStore.generatedSql && !isExecuting.value
  
  // Active when: has sql generation steps OR has generated SQL
  const active = isExecuting.value && hasSqlGenerationSteps
  
  return {
    active: active || !!workspaceStore.generatedSql,
    completed,
    steps,
    sql: workspaceStore.generatedSql,
    duration: completed && start && end ? end - start : (completed && start ? Date.now() - start : 0)
  }
})

// Field Alignment: watch grounding result's suggestedFields specifically
// When field alignment is on, show the panel only after field_suggestions event arrives.
// Only populate once per query cycle — skip if user has already been shown the panel
// or has already confirmed/dismissed (fieldPanelConsumed prevents re-trigger after Phase 2 starts).
watch(() => workspaceStore.groundingResult?.suggestedFields, (fields) => {
  if (useFieldAlignment.value && fields && fields.length > 0 && !showFieldPanel.value && !fieldPanelConsumed.value) {
    suggestedFields.value = fields.map(f => ({ ...f }))
    showFieldPanel.value = true
  }
}, { deep: true })

// Toggle field selection
function toggleField(field: SuggestedFieldFromLinking) {
  field.selected = !field.selected
}

// Build field description from selected linking agent fields
function getFieldDescription(): string {
  const selected = suggestedFields.value.filter(f => f.selected)
  if (selected.length === 0) return ''
  
  return selected.map(f => `${f.tableName}.${f.columnName} (${f.reason})`).join(', ')
}

// Convert frontend GroundingResult back to backend format for injection
// Prefer linking agent's selection (linkingTables/linkingColumns) over retrieval snapshot
function serializeGroundingForInjection(): any {
  const gr = workspaceStore.groundingResult
  if (!gr) return undefined

  // Use linking results if available, otherwise fall back to retrieval snapshot
  const tableSrc = gr.linkingTables?.length ? gr.linkingTables : gr.tables
  const colSrc = gr.linkingColumns?.length ? gr.linkingColumns : gr.columns
  const jpSrc = gr.linkingJoinPaths?.length ? gr.linkingJoinPaths : gr.joinPaths

  return {
    tables: (tableSrc || []).map(t => ({
      name: t.name,
      reason: t.matchedTerms?.[0] || '',
      confidence: t.confidence
    })),
    columns: (colSrc || []).map(c => ({
      table_name: c.table,
      column_name: c.column,
      reason: c.matchedTerms?.[0] || '',
      confidence: c.confidence
    })),
    join_paths: (jpSrc || []).map(jp => ({
      from_table: jp.from?.table,
      from_column: jp.from?.column,
      to_table: jp.to?.table,
      to_column: jp.to?.column
    })),
    suggested_fields: (gr.suggestedFields || []).map(f => ({
      table_name: f.tableName,
      column_name: f.columnName,
      reason: f.reason,
      selected: f.selected
    })),
    execution_time_ms: gr.duration || 0,
    execution_logs: gr.executionLogs || [],
    reasoning: gr.reasoning || '',
    mode: gr.mode || ''
  }
}

// Dismiss field panel and execute full pipeline without field constraints (Phase 2: inference only)
function dismissFieldPanel() {
  showFieldPanel.value = false
  fieldPanelConsumed.value = true
  if (awaitingFieldConfirmation.value) {
    awaitingFieldConfirmation.value = false
    // Phase 2: skip grounding, reuse previous grounding result, no field constraints
    const injectedGrounding = serializeGroundingForInjection()
    workspaceStore.executeQuery(question.value, undefined, false, injectedGrounding)
  }
}

// Confirm field selection and execute inference only with field constraints (Phase 2)
async function confirmFieldsAndExecute() {
  showFieldPanel.value = false
  awaitingFieldConfirmation.value = false
  fieldPanelConsumed.value = true
  const fieldDesc = getFieldDescription()
  const injectedGrounding = serializeGroundingForInjection()
  try {
    // Phase 2: skip grounding, reuse grounding result, inject field description
    await workspaceStore.executeQuery(question.value, fieldDesc, false, injectedGrounding)
  } catch (e: any) {
    message.error(e.message || 'Execution failed')
  }
}

// Re-execute with updated field selections (for post-SQL adjustment — also Phase 2)
async function reExecuteWithFields() {
  showFieldPanel.value = false
  fieldPanelConsumed.value = true
  const fieldDesc = getFieldDescription()
  const injectedGrounding = serializeGroundingForInjection()
  workspaceStore.abortCurrentQuery()
  try {
    await workspaceStore.executeQuery(question.value, fieldDesc, false, injectedGrounding)
  } catch (e: any) {
    message.error(e.message || 'Execution failed')
  }
}

async function handleExecute() {
  if (!question.value.trim()) {
    message.warning('Please enter a question')
    return
  }
  
  if (useFieldAlignment.value) {
    // Phase 1: Run grounding only, then pause for field confirmation
    await doExecuteGroundingOnly()
  } else {
    // No field alignment — execute full pipeline directly
    await doExecuteFull()
  }
}

// Execute grounding only (Phase 1 of field alignment flow)
async function doExecuteGroundingOnly() {
  resetTimings()
  suggestedFields.value = []
  showFieldPanel.value = false
  awaitingFieldConfirmation.value = true
  fieldPanelConsumed.value = false
  
  workspaceStore.queryOptions.maxIterations = maxIterations.value
  workspaceStore.queryOptions.useRichContext = useRichContext.value
  workspaceStore.queryOptions.useReact = useReact.value
  workspaceStore.queryOptions.useGrounding = useGrounding.value
  workspaceStore.queryOptions.linkingMode = linkingMode.value

  try {
    await workspaceStore.executeQuery(question.value, undefined, true) // groundingOnly=true


  } catch (e: any) {
    awaitingFieldConfirmation.value = false
    message.error(e.message || 'Grounding failed')
  }
}

// Execute full pipeline (Phase 2 after field confirmation, or direct when no field alignment)
async function doExecuteFull(fieldDescription?: string) {
  resetTimings()
  
  workspaceStore.queryOptions.maxIterations = maxIterations.value
  workspaceStore.queryOptions.useRichContext = useRichContext.value
  workspaceStore.queryOptions.useReact = useReact.value
  workspaceStore.queryOptions.useGrounding = useGrounding.value
  workspaceStore.queryOptions.linkingMode = linkingMode.value

  try {
    await workspaceStore.executeQuery(question.value, fieldDescription, false)
  } catch (e: any) {
    message.error(e.message || 'Execution failed')
  }
}

function handleStop() {
  workspaceStore.abortCurrentQuery()
}

function useExample(q: string) {
  question.value = q
}

function handleClear() {
  question.value = ''
  resetTimings()
  workspaceStore.resetQueryState()
}

// Parse verify_sql observation into structured execution plan steps
function parseExplainPlan(observation: string): { passed: boolean; steps: { table: string; scan: string; key: string; rows: string; extra: string }[]; warnings: string[]; summary: string } {
  const passed = observation?.startsWith('✅') ?? false
  const steps: { table: string; scan: string; key: string; rows: string; extra: string }[] = []
  const warnings: string[] = []
  let summary = ''

  if (!observation) return { passed, steps, warnings, summary }

  const lines = observation.split('\n')
  let currentStep: { table: string; scan: string; key: string; rows: string; extra: string } | null = null

  for (const line of lines) {
    const stepMatch = line.match(/^\s*Step \d+:\s*(.+)/)
    if (stepMatch) {
      if (currentStep) steps.push(currentStep)
      currentStep = { table: (stepMatch[1] ?? '').trim(), scan: '-', key: '-', rows: '-', extra: '-' }
      continue
    }
    if (currentStep) {
      const scanMatch = line.match(/scan:\s*(\S+)/)
      const keyMatch = line.match(/key:\s*(\S+)/)
      const rowsMatch = line.match(/rows:\s*(\S+)/)
      const extraMatch = line.match(/extra:\s*(.+)/)
      if (scanMatch?.[1]) currentStep.scan = scanMatch[1]
      if (keyMatch?.[1]) currentStep.key = keyMatch[1]
      if (rowsMatch?.[1]) currentStep.rows = rowsMatch[1]
      if (extraMatch?.[1]) currentStep.extra = extraMatch[1].trim()
    }
    if (line.match(/^\s*-\s+/) && line.includes('scan') || line.includes('table') || line.includes('filesort')) {
      warnings.push(line.replace(/^\s*-\s+/, '').trim())
    }
    if (line.includes('Execution plan looks good') || line.includes('proceed to Final Answer')) {
      summary = line.trim()
    }
  }
  if (currentStep) steps.push(currentStep)

  if (!summary) {
    summary = passed ? 'Execution plan looks good' : 'Query has issues'
  }

  return { passed, steps, warnings, summary }
}

async function handleFeedback(type: 'positive' | 'negative', note?: string) {
  // Update query history with feedback
  if (workspaceStore.queryHistory.length > 0) {
    const latestQuery = workspaceStore.queryHistory[0]
    if (latestQuery) {
      latestQuery.feedback = type
      latestQuery.feedbackNote = note
    }
  }
  
  // If negative feedback, this could trigger context update
  if (type === 'negative' && note) {
    console.log('Feedback submitted for context improvement:', {
      question: question.value,
      sql: workspaceStore.generatedSql,
      feedback: type,
      note
    })
    // In production, this would call an API to trigger context update
  }
}
</script>

<template>
  <div class="query-chat min-h-full bg-gray-50 p-6">
    <!-- Control Panel -->
    <div class="control-panel mb-6 p-5 rounded-lg bg-white border border-gray-200">
      <!-- Parameters (at top) -->
      <div class="grid grid-cols-2 lg:grid-cols-5 gap-4 p-4 bg-gray-50 rounded-lg border border-gray-100 mb-5">
        <!-- Model Selection -->
        <div class="param-item">
          <label class="text-xs font-bold text-gray-500 mb-2 block uppercase tracking-wide">Model</label>
          <NSelect
            v-model:value="selectedModel"
            :options="modelOptions"
            :disabled="isExecuting"
            size="small"
          />
        </div>

        <!-- Max Iterations -->
        <div class="param-item">
          <label class="text-xs font-bold text-gray-500 mb-2 block uppercase tracking-wide">Max Iterations</label>
          <NInputNumber
            v-model:value="maxIterations"
            :min="1"
            :max="10"
            :disabled="isExecuting"
            size="small"
            class="w-full"
          />
        </div>

        <!-- Switches -->
        <div class="param-item flex items-end">
          <div class="flex items-center gap-3 h-8">
            <NSwitch v-model:value="useRichContext" :disabled="isExecuting" size="small" />
            <span class="text-sm font-medium text-gray-700">Rich Context</span>
          </div>
        </div>

        <div class="param-item flex items-end">
          <div class="flex items-center gap-3 h-8">
            <NSwitch v-model:value="useFieldAlignment" :disabled="isExecuting" size="small" />
            <span class="text-sm font-medium text-gray-700">Field Alignment</span>
          </div>
        </div>

        <div class="param-item">
          <label class="text-xs font-bold text-gray-500 mb-2 block uppercase tracking-wide">Linking Mode</label>
          <NSelect
            v-model:value="linkingMode"
            :options="linkingModeOptions"
            :disabled="isExecuting"
            size="small"
          />
        </div>
      </div>

      <!-- Question Input -->
      <div class="mb-5">
        <div class="relative">
          <NInput
            v-model:value="question"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="e.g. List all TV channels with their countries..."
            :disabled="isExecuting"
            class="query-input !text-lg !font-medium !p-4"
            @keydown.ctrl.enter="handleExecute"
          />
          <div class="absolute right-4 bottom-4 text-xs text-gray-400 font-medium">
            Ctrl + Enter to execute
          </div>
        </div>

        <!-- Example questions - Collapsible -->
        <div class="mt-4 example-collapse">
          <NCollapse :default-expanded-names="['examples']" arrow-placement="left">
            <NCollapseItem name="examples">
              <template #header>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-600">Example Questions</span>
                  <span class="text-xs text-gray-400">(Spider Benchmark)</span>
                </div>
              </template>
              <div class="flex flex-wrap gap-3 pt-2">
                <button
                  v-for="example in exampleQuestions"
                  :key="example"
                  class="text-sm px-3 py-2 rounded-md bg-gray-100 text-gray-600 hover:bg-primary-50 hover:text-primary-600 transition-colors font-medium border border-gray-200 hover:border-primary-200"
                  @click="useExample(example)"
                >
                  {{ example }}
                </button>
              </div>
            </NCollapseItem>
          </NCollapse>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-3 mt-5">
        <button
          :disabled="!question.trim() || isExecuting"
          class="execute-btn flex items-center gap-2 px-5 py-2.5 rounded-lg text-white font-medium text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          :class="isExecuting 
            ? 'bg-amber-500 hover:bg-amber-600' 
            : 'bg-primary-600 hover:bg-primary-700'"
          @click="handleExecute"
        >
          <div v-if="isExecuting" class="i-lucide-loader-2 animate-spin" />
          <div v-else class="i-lucide-play" />
          {{ isExecuting ? 'Executing...' : 'Execute Query' }}
        </button>

        <button
          v-if="isExecuting"
          class="flex items-center gap-2 px-4 py-2.5 rounded-lg bg-red-500 text-white font-medium text-sm hover:bg-red-600 transition-colors"
          @click="handleStop"
        >
          <div class="i-lucide-square" />
          Stop
        </button>

        <button
          :disabled="isExecuting"
          class="flex items-center gap-2 px-4 py-2.5 rounded-lg bg-white text-gray-600 font-medium text-sm border border-gray-200 hover:border-gray-300 hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          @click="handleClear"
        >
          <div class="i-lucide-eraser" />
          Clear
        </button>
      </div>
    </div>

    <!-- Real-time Execution Cards -->
    <div class="execution-pipeline grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
      <!-- Stage 1: Vector Search / Schema Loaded -->
      <RealtimeCard
        :title="isSmallScale ? 'Schema Loaded' : 'Vector Search'"
        :subtitle="isSmallScale ? 'Small-scale path · full schema sent to linker' : undefined"
        :icon="isSmallScale ? 'i-lucide-database' : 'i-lucide-search'"
        :active="vectorSearchStage.active"
        :pending="isExecuting && !vectorSearchStage.active && !vectorSearchStage.completed"
        :stage="workspaceStore.groundingStage"
        :completed="vectorSearchStage.completed"
        :duration="vectorSearchStage.duration"
        color="blue"
      >
        <template #content>
          <!-- Skeleton screen: shows table names from local cache while waiting for backend -->
          <div v-if="workspaceStore.showSkeleton && workspaceStore.skeletonTables.length > 0" class="space-y-4 animate-pulse">
            <div class="flex items-center gap-2 mb-2">
            <div class="i-lucide-table-2 text-sm text-blue-400" />
              <span class="text-xs font-bold text-gray-400 uppercase tracking-wide">Analyzing {{ workspaceStore.skeletonTables.length }} tables...</span>
            </div>
            <div class="space-y-2">
              <div
                v-for="table in workspaceStore.skeletonTables.slice(0, 8)"
                :key="'sk-' + table"
                class="flex items-center justify-between px-3 py-2 rounded-lg bg-gray-50 border border-gray-100"
              >
                <span class="text-sm text-gray-400 font-medium">{{ table }}</span>
                <div class="w-16 h-1.5 rounded-full bg-gray-100" />
              </div>
              <div v-if="workspaceStore.skeletonTables.length > 8" class="text-xs text-gray-400 text-center">
                +{{ workspaceStore.skeletonTables.length - 8 }} more tables
              </div>
            </div>
          </div>
          <div v-else-if="vectorSearchStage.empty" class="flex flex-col items-center justify-center py-8 text-gray-400">
            <div class="i-lucide-folder-open text-3xl mb-2 opacity-40" />
            <span class="text-sm font-medium">No context available</span>
            <span class="text-xs mt-1 opacity-70">Generate Rich Context to enable vector retrieval</span>
          </div>
          <div v-else-if="workspaceStore.groundingResult" class="space-y-4 content-fade">
            <!-- Tables with confidence -->
            <div v-if="workspaceStore.groundingResult.tables?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-table-2 text-sm text-blue-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Retrieved Tables ({{ workspaceStore.groundingResult.tables.length }})</span>
              </div>
              <div class="space-y-2">
                <div
                  v-for="table in workspaceStore.groundingResult.tables"
                  :key="table.name"
                  class="grounding-item px-3 py-2 rounded-lg bg-blue-50 border border-blue-100 hover:bg-blue-100/80 transition-colors"
                >
                  <div class="flex items-center justify-between">
                    <span class="text-sm text-blue-800 font-medium">{{ table.name }}</span>
                    <div class="flex items-center gap-2">
                      <div class="w-16 h-1.5 rounded-full bg-blue-100 overflow-hidden">
                        <div 
                          class="h-full rounded-full bg-blue-500 transition-all duration-500"
                          :style="{ width: `${(table.confidence * 100)}%` }"
                        />
                      </div>
                      <span class="text-xs text-gray-500 font-bold w-8 text-right">{{ (table.confidence * 100).toFixed(0) }}%</span>
                    </div>
                  </div>
                  <div v-if="table.description" class="text-xs text-gray-500 mt-1 leading-relaxed truncate" :title="table.description">
                    {{ table.description }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Columns grouped by table -->
            <div v-if="workspaceStore.groundingResult.columns?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-columns-3 text-sm text-cyan-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Retrieved Columns ({{ workspaceStore.groundingResult.columns.length }})</span>
              </div>
              <div class="flex flex-wrap gap-1.5">
                <div
                  v-for="col in workspaceStore.groundingResult.columns.slice(0, 10)"
                  :key="`${col.table}.${col.column}`"
                  class="column-tag px-2 py-1 rounded bg-cyan-50 border border-cyan-100 text-xs font-medium hover:bg-cyan-100/80 transition-colors"
                  :title="[col.dataType, col.description].filter(Boolean).join(' — ')"
                >
                  <span class="text-gray-500">{{ col.table }}.</span><span class="text-cyan-700">{{ col.column }}</span>
                  <span v-if="col.dataType" class="text-gray-400 ml-1">({{ col.dataType }})</span>
                </div>
                <div v-if="workspaceStore.groundingResult.columns.length > 10" class="column-tag px-2 py-1 text-xs text-gray-400 font-medium">
                  +{{ workspaceStore.groundingResult.columns.length - 10 }} more
                </div>
              </div>
            </div>

            <!-- Join paths if any -->
            <div v-if="workspaceStore.groundingResult.joinPaths?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-git-merge text-sm text-purple-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Join Paths</span>
              </div>
              <div class="space-y-1">
                <div
                  v-for="(path, idx) in workspaceStore.groundingResult.joinPaths.slice(0, 3)"
                  :key="idx"
                  class="grounding-item flex items-center gap-2 text-xs text-gray-500 font-medium"
                >
                  <span class="text-purple-700">{{ path.from?.table }}.{{ path.from?.column }}</span>
                  <div class="i-lucide-arrow-right text-gray-400" />
                  <span class="text-purple-700">{{ path.to?.table }}.{{ path.to?.column }}</span>
                </div>
              </div>
            </div>


          </div>
          <div v-else-if="vectorSearchStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
            <div :class="isSmallScale ? 'i-lucide-database' : 'i-lucide-search'" class="animate-pulse text-blue-500 text-xl" />
            <div class="space-y-1">
              <span class="font-medium block">
                {{ isSmallScale
                  ? 'Loading schema (small scale)...'
                  : (workspaceStore.groundingProgress?.stage === 'linking_start' || workspaceStore.groundingProgress?.stage === 'linking_done'
                    ? 'Vector retrieval complete'
                    : 'Searching vector database...') }}
              </span>
              <span class="text-xs text-gray-400">
                {{ isSmallScale
                  ? 'All tables passed directly to linking agent'
                  : (workspaceStore.groundingProgress?.stage === 'retrieval_done'
                    ? `Found ${workspaceStore.groundingProgress?.data?.candidate_tables || 0} candidate tables`
                    : 'Identifying relevant tables and columns') }}
              </span>
            </div>
          </div>
          <div v-else-if="isExecuting" class="flex items-center gap-3 text-sm text-gray-500 pending-indicator">
            <div class="i-lucide-search animate-pulse text-blue-400 text-xl" />
            <div class="space-y-1">
              <span class="font-medium block text-gray-600">Preparing vector search...</span>
              <span class="text-xs text-gray-400">Waiting for pipeline to start</span>
            </div>
          </div>
          <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
            <div class="i-lucide-search text-3xl mb-2 opacity-30" />
            <span class="text-sm font-medium">Waiting for query...</span>
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 2: Schema Linking -->
      <RealtimeCard
        :title="linkingMode === 'react' ? 'ReAct Schema Linking' : linkingMode === 'off' ? 'Schema Linking (Off)' : 'One-Shot Schema Linking'"
        icon="i-lucide-link"
        :active="schemaLinkingStage.active"
        :pending="isExecuting && !schemaLinkingStage.active && !schemaLinkingStage.completed"
        :completed="schemaLinkingStage.completed"
        :duration="schemaLinkingStage.duration"
        color="cyan"
      >
        <template #content>
          <!-- Step 1: Linking Agent Result (LLM Fine Selection + Selected Tables/Columns + Execution Logs) -->
          <div v-if="workspaceStore.groundingResult?.reasoning || workspaceStore.groundingResult?.linkingTables?.length" class="space-y-3 mb-4">
            <!-- LLM Fine Selection -->
            <div v-if="workspaceStore.groundingResult.reasoning" class="stagger-item" style="--stagger: 0">
              <NCollapse :default-expanded-names="['reasoning']" arrow-placement="left">
                <NCollapseItem name="reasoning">
                  <template #header>
                    <div class="flex items-center gap-2">
                      <div class="i-lucide-brain text-sm text-indigo-500" />
                      <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">LLM Fine Selection</span>
                      <span v-if="workspaceStore.groundingResult.mode" class="px-1.5 py-0.5 text-xs bg-indigo-100 text-indigo-700 rounded">
                        {{ workspaceStore.groundingResult.mode }}
                      </span>
                    </div>
                  </template>
                  <div class="p-3 rounded-lg bg-indigo-50 border border-indigo-100 text-sm text-gray-700 leading-relaxed mt-2">
                    {{ workspaceStore.groundingResult.reasoning }}
                  </div>
                </NCollapseItem>
              </NCollapse>
            </div>

            <!-- Linking Agent Selected Tables/Columns — Grouped by Table -->
            <div v-if="linkedTableGroups.length" class="stagger-item" style="--stagger: 1">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-filter text-sm text-teal-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">
                  Linked Schema ({{ linkedTableGroups.length }} tables, {{ workspaceStore.groundingResult.linkingColumns?.length || 0 }} columns)
                </span>
              </div>
              <div class="space-y-1.5">
                <div
                  v-for="group in linkedTableGroups"
                  :key="'lg-' + group.name"
                  class="rounded-lg border border-teal-100 bg-teal-50/50 overflow-hidden"
                >
                  <!-- Table header row -->
                  <div class="flex items-center gap-2 px-3 py-1.5">
                    <div class="w-5 h-5 rounded bg-teal-100 flex items-center justify-center flex-shrink-0">
                      <div class="i-lucide-table-2 text-xs text-teal-600" />
                    </div>
                    <span class="text-xs font-bold text-teal-800">{{ group.name }}</span>
                    <span class="text-xs text-gray-400 ml-auto">{{ group.columns.length }} cols</span>
                  </div>
                  <!-- Columns row -->
                  <div v-if="group.columns.length" class="px-3 pb-2 flex flex-wrap gap-1">
                    <span
                      v-for="c in group.columns"
                      :key="'lc-' + c.table + '.' + c.column"
                      class="px-1.5 py-0.5 rounded bg-white border border-teal-100 text-xs font-medium text-teal-700"
                      :title="[c.hint, c.dataType, c.description].filter(Boolean).join(' — ')"
                    >{{ c.column }}<span v-if="c.dataType" class="text-gray-400 ml-0.5 text-[10px]">{{ c.dataType }}</span></span>
                  </div>
                  <!-- Hint row (query-specific, generated by LLM) -->
                  <div v-if="group.hint" class="px-3 pb-2">
                    <div class="flex items-start gap-1.5">
                      <div class="i-lucide-sparkles text-[10px] text-amber-500 mt-0.5 shrink-0" />
                      <p class="text-[10px] text-amber-700 leading-snug" :title="group.hint">{{ group.hint }}</p>
                    </div>
                  </div>
                  <!-- Description row (if available and no hint) -->
                  <div v-else-if="group.description" class="px-3 pb-2">
                    <p class="text-[10px] text-gray-400 leading-snug truncate" :title="group.description">{{ group.description }}</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Execution Logs -->
            <div v-if="workspaceStore.groundingResult.executionLogs?.length" class="stagger-item" style="--stagger: 2">
              <NCollapse :default-expanded-names="[]" arrow-placement="left">
                <NCollapseItem name="logs">
                  <template #header>
                    <div class="flex items-center gap-2">
                      <div class="i-lucide-terminal text-sm text-gray-500" />
                      <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Execution Log</span>
                      <span class="text-xs text-gray-400">({{ workspaceStore.groundingResult.executionLogs.length }} queries)</span>
                    </div>
                  </template>
                  <div class="space-y-2 mt-2">
                    <div
                      v-for="(log, idx) in workspaceStore.groundingResult.executionLogs"
                      :key="idx"
                      class="p-3 rounded-lg bg-gray-800 text-xs font-mono"
                    >
                      <div class="flex items-center justify-between mb-2">
                        <span class="text-green-400 font-bold">{{ log.phase }}</span>
                        <span class="text-gray-400">{{ log.duration_ms }}ms | {{ log.result_count }} results</span>
                      </div>
                      <div class="text-gray-300 overflow-x-auto whitespace-pre-wrap break-all">{{ log.sql }}</div>
                      <div class="mt-2 text-gray-500 italic">{{ log.summary }}</div>
                    </div>
                  </div>
                </NCollapseItem>
              </NCollapse>
            </div>
          </div>

          <!-- Step 2: Field Suggestions Panel — shown after linking agent result -->
          <Transition name="field-panel">
          <div v-if="showFieldPanel && suggestedFields.length > 0" class="p-4 rounded-lg bg-purple-50 border border-purple-200">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <div class="i-lucide-table-2 text-purple-500 text-sm" />
                  <span class="text-xs font-bold text-purple-700 uppercase tracking-wide">Suggested Fields</span>
                  <span class="px-1.5 py-0.5 text-xs bg-purple-100 text-purple-600 rounded font-medium">from linking</span>
                  <span v-if="awaitingFieldConfirmation" class="px-1.5 py-0.5 text-xs bg-amber-100 text-amber-700 rounded font-medium animate-pulse">
                    awaiting confirmation
                  </span>
                </div>
                <button 
                  class="text-gray-400 hover:text-gray-600 transition-colors p-0.5"
                  @click="dismissFieldPanel"
                >
                  <div class="i-lucide-x text-sm" />
                </button>
              </div>
              
              <!-- Field list (vertical) -->
              <div class="space-y-1.5 mb-3">
                <button
                  v-for="field in suggestedFields"
                  :key="`${field.tableName}.${field.columnName}`"
                  class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium transition-all cursor-pointer border"
                  :class="field.selected 
                    ? 'bg-purple-50 border-purple-200 text-purple-800' 
                    : 'bg-white border-gray-200 text-gray-500 hover:border-purple-200'"
                  :title="field.reason"
                  @click="toggleField(field)"
                >
                  <NCheckbox :checked="field.selected" size="small" @update:checked="(v: boolean) => field.selected = v" @click.stop />
                  <span class="font-mono font-bold">{{ field.columnName }}</span>
                  <span class="text-gray-400 font-normal ml-auto">{{ field.tableName }}</span>
                </button>
              </div>
              
              <!-- Action row — only shown during initial field confirmation -->
              <div v-if="awaitingFieldConfirmation" class="flex items-center justify-between pt-2 border-t border-purple-100">
                <p class="text-xs text-gray-400">
                  <span class="i-lucide-info inline-block mr-0.5 align-middle" />
                  Select the fields you want in the output, then confirm to generate SQL.
                </p>
                <div class="flex items-center gap-2">
                  <button
                    class="px-3 py-1.5 rounded-lg bg-white border border-gray-200 text-gray-600 font-medium text-xs hover:bg-gray-50 transition-all"
                    @click="dismissFieldPanel"
                  >
                    Skip
                  </button>
                  <button
                    class="px-3 py-1.5 rounded-lg text-white font-bold text-xs shadow-sm hover:-translate-y-0.5 transition-all bg-primary-600 hover:bg-primary-700"
                    :disabled="isExecuting"
                    @click="confirmFieldsAndExecute()"
                  >
                    Confirm & Generate SQL
                  </button>
                </div>
              </div>
            </div>
          </Transition>

          <!-- Schema Linking Steps (only shown when inference is running / complete) -->
          <div v-if="schemaLinkingStage.isPolling || schemaLinkingStage.steps.length" class="space-y-4" :class="{ 'mt-4': showFieldPanel }">
            <!-- Polling indicator: waiting for schema data (cold-start acceleration) -->
            <div v-if="schemaLinkingStage.pollingCount > 0 || schemaLinkingStage.schemaReceived" class="flex items-center gap-3 px-4 py-3 rounded-lg border"
              :class="schemaLinkingStage.schemaReceived ? 'bg-green-50/50 border-green-100' : 'bg-cyan-50/50 border-cyan-100'"
            >
              <div v-if="schemaLinkingStage.isPolling" class="i-lucide-loader-2 animate-spin text-cyan-500 text-lg flex-shrink-0" />
              <div v-else class="i-lucide-check-circle text-green-500 text-lg flex-shrink-0" />
              <div class="flex-1">
                <span v-if="schemaLinkingStage.isPolling" class="text-sm font-medium text-gray-600">Waiting for schema data...</span>
                <span v-else class="text-sm font-medium text-green-700">Schema data received</span>
                <span v-if="schemaLinkingStage.pollingCount > 0 && !schemaLinkingStage.isPolling" class="text-xs text-gray-400 ml-2">
                  (waited {{ schemaLinkingStage.pollingCount }} {{ schemaLinkingStage.pollingCount === 1 ? 'round' : 'rounds' }})
                </span>
              </div>
            </div>

            <!-- Step counter (only for real analysis steps) -->
            <div v-if="schemaLinkingStage.steps.length" class="flex items-center gap-2 pb-2 border-b border-gray-100">
              <div class="text-xs font-bold text-gray-500 uppercase tracking-wide">
                {{ schemaLinkingStage.steps.length }} reasoning step{{ schemaLinkingStage.steps.length > 1 ? 's' : '' }}
              </div>
            </div>
            
            <!-- Steps -->
            <TransitionGroup name="step-list" tag="div" class="space-y-4">
            <div
              v-for="(step, idx) in schemaLinkingStage.steps"
              :key="step.step || idx"
              class="react-step p-4 rounded-lg bg-cyan-50 border border-cyan-100"
            >
              <div class="flex items-start gap-3">
                <!-- Step indicator -->
                <div class="flex flex-col items-center">
                  <div class="w-6 h-6 rounded-full bg-white flex items-center justify-center flex-shrink-0 shadow-sm border border-cyan-100">
                    <span class="text-xs text-cyan-600 font-bold">{{ step.step }}</span>
                  </div>
                </div>
                
                <div class="flex-1 min-w-0 space-y-3">
                  <!-- Thought -->
                  <div v-if="step.thought" class="flex items-start gap-2">
                    <div class="i-lucide-lightbulb text-cyan-600 mt-0.5 flex-shrink-0" />
                    <p class="text-sm text-gray-700 leading-relaxed font-medium">{{ step.thought }}</p>
                  </div>
                  
                  <!-- Action -->
                  <div v-if="step.action" class="flex items-start gap-2 bg-white p-2 rounded border border-cyan-100">
                    <div class="i-lucide-play text-teal-600 mt-0.5 flex-shrink-0" />
                    <div>
                      <span class="text-xs text-teal-700 font-mono font-bold">{{ step.action }}</span>
                      <span v-if="step.actionInput" class="text-xs text-gray-500 ml-2 font-mono">
                        {{ typeof step.actionInput === 'object' ? JSON.stringify(step.actionInput) : step.actionInput }}
                      </span>
                    </div>
                  </div>
                  
                  <!-- Observation -->
                  <div v-if="step.observation" class="flex items-start gap-2">
                    <div class="i-lucide-eye text-amber-500 mt-0.5 flex-shrink-0" />
                    <p class="text-xs text-gray-500 leading-relaxed">{{ step.observation }}</p>
                  </div>
                </div>
              </div>
            </div>
            </TransitionGroup>
          </div>
          <!-- Loading/Waiting states (only when NO linking result AND field panel is NOT shown) -->
          <div v-else-if="!showFieldPanel && !(workspaceStore.groundingResult?.reasoning || workspaceStore.groundingResult?.linkingTables?.length)">
            <!-- Linking agent progress from grounding sub-stages -->
            <div v-if="workspaceStore.groundingProgress?.stage === 'linking_start'" class="space-y-3">
              <div class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
                <div class="i-lucide-brain animate-pulse text-cyan-500 text-xl" />
                <div class="space-y-1">
                  <span class="font-medium block">Linking agent analyzing schema...</span>
                  <span class="text-xs text-gray-400">
                    Selecting relevant tables from {{ workspaceStore.groundingProgress?.data?.table_count || '?' }} candidates
                  </span>
                </div>
              </div>
            </div>
            <div v-else-if="workspaceStore.groundingProgress?.stage === 'linking_done'" class="space-y-3 content-fade">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-brain text-sm text-cyan-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Linking Agent Result</span>
                <span class="text-xs text-gray-400">
                  {{ workspaceStore.groundingProgress?.data?.selected_tables || 0 }} tables selected
                  in {{ workspaceStore.groundingProgress?.data?.duration_ms || 0 }}ms
                </span>
              </div>
              <div v-if="workspaceStore.groundingProgress?.data?.reasoning" class="p-3 rounded-lg bg-cyan-50 border border-cyan-100 text-sm text-gray-700 leading-relaxed">
                {{ workspaceStore.groundingProgress?.data?.reasoning }}
              </div>
            </div>
            <div v-else-if="schemaLinkingStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
            <div class="i-lucide-link animate-pulse text-cyan-500 text-xl" />
              <div class="space-y-1">
                <span class="font-medium block">Analyzing schema structure...</span>
                <span class="text-xs text-gray-400">Identifying table relationships and join paths</span>
              </div>
            </div>
            <div v-else-if="isExecuting && workspaceStore.groundingProgress?.stage === 'linking_start'" class="flex items-center gap-3 text-sm text-gray-500 pending-indicator">
            <div class="i-lucide-brain animate-pulse text-cyan-400 text-xl" />
              <div class="space-y-1">
                <span class="font-medium block text-gray-600">Linking agent analyzing...</span>
                <span class="text-xs text-gray-400">Selecting relevant tables from {{ workspaceStore.groundingProgress?.data?.table_count || '?' }} candidates</span>
              </div>
            </div>
            <div v-else-if="isExecuting" class="flex items-center gap-3 text-sm text-gray-500 pending-indicator">
            <div class="i-lucide-link animate-pulse text-cyan-400 text-xl" />
              <div class="space-y-1">
                <span class="font-medium block text-gray-600">Waiting for schema linking...</span>
                <span class="text-xs text-gray-400">Will start after vector retrieval completes</span>
              </div>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
            <div class="i-lucide-link text-3xl mb-2 opacity-30" />
              <span class="text-sm font-medium">Waiting for schema linking...</span>
            </div>
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 3: SQL Generation -->
      <RealtimeCard
        title="ReAct SQL Generation"
        icon="i-lucide-code-2"
        :active="sqlGenerationStage.active"
        :pending="isExecuting && !sqlGenerationStage.active && !sqlGenerationStage.completed"
        :completed="sqlGenerationStage.completed"
        :duration="sqlGenerationStage.duration"
        color="purple"
      >
        <template #content>
          <div class="space-y-4">
            <!-- Steps if any -->
            <div v-if="sqlGenerationStage.steps.length" class="space-y-4">
              <div class="flex items-center gap-2 pb-2 border-b border-gray-100">
                <div class="text-xs font-bold text-gray-500 uppercase tracking-wide">
                  {{ sqlGenerationStage.steps.length }} generation step{{ sqlGenerationStage.steps.length > 1 ? 's' : '' }}
                </div>
              </div>
              
              <TransitionGroup name="step-list" tag="div" class="space-y-4">
              <div
                v-for="(step, idx) in sqlGenerationStage.steps"
                :key="step.step || idx"
                class="react-step p-4 rounded-lg border"
                :class="step.action === 'verify_sql' && step.observation
                  ? (step.observation.startsWith('✅') ? 'bg-green-50 border-green-200' : step.observation.startsWith('❌') ? 'bg-red-50 border-red-200' : 'bg-purple-50 border-purple-100')
                  : 'bg-purple-50 border-purple-100'"
              >
                <div class="flex items-start gap-3">
                  <div class="w-6 h-6 rounded-full bg-white flex items-center justify-center flex-shrink-0 shadow-sm border"
                    :class="step.action === 'verify_sql' && step.observation
                      ? (step.observation.startsWith('✅') ? 'border-green-200' : step.observation.startsWith('❌') ? 'border-red-200' : 'border-purple-100')
                      : 'border-purple-100'"
                  >
                    <span v-if="step.action === 'verify_sql' && step.observation?.startsWith('✅')" class="text-xs text-green-600 font-bold">✓</span>
                    <span v-else-if="step.action === 'verify_sql' && step.observation?.startsWith('❌')" class="text-xs text-red-600 font-bold">✗</span>
                    <span v-else class="text-xs text-purple-600 font-bold">{{ step.step }}</span>
                  </div>
                  
                  <div class="flex-1 min-w-0 space-y-3">
                    <div v-if="step.thought" class="flex items-start gap-2">
                      <div class="i-lucide-lightbulb text-purple-600 mt-0.5 flex-shrink-0" />
                      <p class="text-sm text-gray-700 leading-relaxed font-medium">{{ step.thought }}</p>
                    </div>
                    
                    <!-- Action: verify_sql with status badge + expandable execution plan -->
                    <div v-if="step.action === 'verify_sql'" class="space-y-2">
                      <div class="flex items-center gap-2 bg-white p-2 rounded border"
                        :class="step.observation?.startsWith('✅') ? 'border-green-200' : step.observation?.startsWith('❌') ? 'border-red-200' : 'border-purple-100'"
                      >
                        <div class="i-lucide-check-circle mt-0.5 flex-shrink-0"
                          :class="step.observation?.startsWith('✅') ? 'text-green-600' : step.observation?.startsWith('❌') ? 'text-red-600' : 'text-pink-600'"
                        />
                        <span class="text-xs font-mono font-bold"
                          :class="step.observation?.startsWith('✅') ? 'text-green-700' : step.observation?.startsWith('❌') ? 'text-red-700' : 'text-pink-600'"
                        >verify_sql</span>
                        <NTag v-if="step.observation?.startsWith('✅')" size="tiny" type="success" round>PASSED</NTag>
                        <NTag v-else-if="step.observation?.startsWith('❌')" size="tiny" type="error" round>FAILED</NTag>
                        <NTag v-else-if="step.observation" size="tiny" type="warning" round>CHECKING</NTag>
                      </div>
                      
                      <!-- Execution Plan (structured display) -->
                      <div v-if="step.observation" class="verify-result mt-2">
                        <div class="rounded-lg overflow-hidden border"
                          :class="step.observation?.startsWith('✅') ? 'border-green-200' : 'border-red-200'"
                        >
                          <!-- Structured EXPLAIN table -->
                          <div v-if="parseExplainPlan(step.observation).steps.length > 0" class="p-2">
                            <table class="w-full text-xs">
                              <thead>
                                <tr class="text-left text-gray-400 uppercase tracking-wider">
                                  <th class="px-2 py-1.5 font-semibold">Table</th>
                                  <th class="px-2 py-1.5 font-semibold">Scan</th>
                                  <th class="px-2 py-1.5 font-semibold">Key</th>
                                  <th class="px-2 py-1.5 font-semibold text-right">Rows</th>
                                </tr>
                              </thead>
                              <tbody>
                                <tr
                                  v-for="(planStep, pi) in parseExplainPlan(step.observation).steps"
                                  :key="pi"
                                  class="border-t border-gray-100"
                                >
                                  <td class="px-2 py-1.5 font-mono font-medium text-gray-700">{{ planStep.table }}</td>
                                  <td class="px-2 py-1.5">
                                    <span class="px-1.5 py-0.5 rounded text-xs font-medium"
                                      :class="planStep.scan === 'ALL' ? 'bg-red-100 text-red-700' : planStep.scan === 'eq_ref' || planStep.scan === 'const' ? 'bg-green-100 text-green-700' : 'bg-blue-100 text-blue-700'"
                                    >{{ planStep.scan }}</span>
                                  </td>
                                  <td class="px-2 py-1.5 font-mono text-gray-500">{{ planStep.key }}</td>
                                  <td class="px-2 py-1.5 text-right font-mono text-gray-500">{{ planStep.rows }}</td>
                                </tr>
                              </tbody>
                            </table>
                          </div>
                          <!-- Fallback: raw text if parsing fails -->
                          <pre v-else class="text-xs text-gray-600 font-mono whitespace-pre-wrap leading-relaxed p-3 max-h-48 overflow-y-auto"
                            :class="step.observation?.startsWith('✅') ? 'bg-green-50/50' : 'bg-red-50/50'"
                          >{{ step.observation }}</pre>
                          <!-- Summary footer -->
                          <div class="px-3 py-1.5 text-xs flex items-center gap-1.5 border-t"
                            :class="step.observation?.startsWith('✅') ? 'bg-green-50 text-green-600 border-green-100' : 'bg-red-50 text-red-600 border-red-100'"
                          >
                            <div :class="step.observation?.startsWith('✅') ? 'i-lucide-check-circle' : 'i-lucide-alert-triangle'" class="text-xs" />
                            <span class="font-medium">{{ parseExplainPlan(step.observation).summary }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <!-- Regular action (non-verify_sql) -->
                    <div v-else-if="step.action" class="space-y-2">
                      <div class="flex items-start gap-2 bg-white p-2 rounded border border-purple-100">
                        <div class="i-lucide-play text-pink-600 mt-0.5 flex-shrink-0" />
                        <span class="text-xs text-pink-600 font-mono font-bold">{{ step.action }}</span>
                      </div>
                      <!-- Observation for execute_sql: show as code block -->
                      <div v-if="step.observation && step.action === 'execute_sql'" class="rounded-lg overflow-hidden border border-gray-200">
                        <div class="px-3 py-1.5 bg-gray-100 text-xs font-medium text-gray-500 flex items-center gap-1.5">
                          <div class="i-lucide-terminal text-xs" />
                          Query Result
                        </div>
                        <pre class="text-xs text-gray-600 font-mono whitespace-pre-wrap leading-relaxed p-3 bg-gray-50 max-h-32 overflow-y-auto">{{ step.observation }}</pre>
                      </div>
                      <!-- Observation for other actions -->
                      <div v-else-if="step.observation" class="flex items-start gap-2">
                        <div class="i-lucide-eye text-amber-500 mt-0.5 flex-shrink-0" />
                        <p class="text-xs text-gray-500 leading-relaxed">{{ step.observation }}</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              </TransitionGroup>
            </div>
            
            <!-- Generated SQL Preview -->
            <div v-if="sqlGenerationStage.sql" class="mt-4 p-4 rounded-lg bg-gray-900 border border-gray-800 shadow-inner sql-highlight-enter">
              <div class="flex items-center gap-2 mb-3 border-b border-gray-800 pb-2">
                <div class="i-lucide-check text-green-400" />
                <span class="text-xs text-green-400 font-bold uppercase tracking-wide">SQL Generated</span>
              </div>
              <pre class="text-xs text-gray-300 font-mono overflow-x-auto whitespace-pre-wrap">{{ sqlGenerationStage.sql.substring(0, 200) }}{{ sqlGenerationStage.sql.length > 200 ? '...' : '' }}</pre>
            </div>
            
            <!-- Loading state -->
            <div v-if="!sqlGenerationStage.steps.length && !sqlGenerationStage.sql">
              <div v-if="sqlGenerationStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
            <div class="i-lucide-code-2 animate-pulse text-purple-500 text-xl" />
                <div class="space-y-1">
                  <span class="font-medium block">Generating SQL query...</span>
                  <span class="text-xs text-gray-400">Building optimized query from context</span>
                </div>
              </div>
              <div v-else-if="isExecuting" class="flex items-center gap-3 text-sm text-gray-500 pending-indicator">
            <div class="i-lucide-code-2 animate-pulse text-purple-400 text-xl" />
                <div class="space-y-1">
                  <span class="font-medium block text-gray-600">Waiting for SQL generation...</span>
                  <span class="text-xs text-gray-400">Will start after schema linking completes</span>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
            <div class="i-lucide-code-2 text-3xl mb-2 opacity-30" />
                <span class="text-sm font-medium">Waiting for SQL generation...</span>
              </div>
            </div>
          </div>
        </template>
      </RealtimeCard>
    </div>

    <!-- Grounding Error (shown in-place, NOT in Generated SQL area) -->
    <div v-if="workspaceStore.groundingError" class="rounded-xl overflow-hidden bg-white border border-red-200 shadow-sm">
      <div class="p-6 bg-red-50 border-l-4 border-red-500">
        <div class="flex items-start justify-between">
          <div class="flex items-start gap-3">
            <div class="i-lucide-alert-triangle text-xl text-red-500 flex-shrink-0 mt-1" />
            <div>
              <h4 class="text-red-700 font-bold mb-1">Grounding Error</h4>
              <p class="text-sm text-red-600">{{ workspaceStore.groundingError }}</p>
              <p class="text-xs text-red-400 mt-2">The grounding pipeline failed. SQL generation was not started.</p>
            </div>
          </div>
          <button
            class="px-3 py-1.5 text-sm font-medium text-red-600 bg-white border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
            @click="handleExecute"
          >
            <div class="flex items-center gap-1.5">
              <div class="i-lucide-rotate-ccw text-sm" />
              <span>Retry</span>
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- Query Result (only for SQL generation / execution errors, NOT grounding errors) -->
    <QueryResult
      v-if="workspaceStore.generatedSql || workspaceStore.queryError"
      :sql="workspaceStore.generatedSql"
      :error="workspaceStore.queryError"
      :duration="workspaceStore.queryDuration"
      :result="workspaceStore.executionResult"
      :loading="workspaceStore.isQuerying"
      :database-id="workspaceStore.currentDatabaseId || undefined"
      @retry="handleExecute"
      @feedback="handleFeedback"
    />
  </div>
</template>

<style scoped>
.query-input :deep(.n-input__textarea-el) {
  font-size: 1.125rem;
  line-height: 1.75rem;
}

.example-collapse :deep(.n-collapse-item__header) {
  padding: 12px 0;
}

.example-collapse :deep(.n-collapse-item__content-inner) {
  padding-top: 8px;
}

/* React step animation */
.react-step {
  animation: slideIn 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Step list transition group */
.step-list-enter-active {
  transition: all 0.45s cubic-bezier(0.16, 1, 0.3, 1);
}
.step-list-enter-from {
  opacity: 0;
  transform: translateY(-16px) scale(0.97);
}

/* Stagger animation for linking card sections */
.stagger-item {
  animation: staggerFadeIn 0.4s cubic-bezier(0.16, 1, 0.3, 1) both;
  animation-delay: calc(var(--stagger, 0) * 150ms);
}

@keyframes staggerFadeIn {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Grounding result items animation */
.grounding-item {
  animation: fadeSlideIn 0.35s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.grounding-item:nth-child(1) { animation-delay: 0ms; }
.grounding-item:nth-child(2) { animation-delay: 50ms; }
.grounding-item:nth-child(3) { animation-delay: 100ms; }
.grounding-item:nth-child(4) { animation-delay: 150ms; }
.grounding-item:nth-child(5) { animation-delay: 200ms; }
.grounding-item:nth-child(6) { animation-delay: 250ms; }
.grounding-item:nth-child(7) { animation-delay: 300ms; }
.grounding-item:nth-child(8) { animation-delay: 350ms; }

@keyframes fadeSlideIn {
  from {
    opacity: 0;
    transform: translateX(-8px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

/* Column tags animation */
.column-tag {
  animation: scaleIn 0.25s cubic-bezier(0.16, 1, 0.3, 1) both;
}

.column-tag:nth-child(1) { animation-delay: 0ms; }
.column-tag:nth-child(2) { animation-delay: 30ms; }
.column-tag:nth-child(3) { animation-delay: 60ms; }
.column-tag:nth-child(4) { animation-delay: 90ms; }
.column-tag:nth-child(5) { animation-delay: 120ms; }
.column-tag:nth-child(6) { animation-delay: 150ms; }
.column-tag:nth-child(7) { animation-delay: 180ms; }
.column-tag:nth-child(8) { animation-delay: 210ms; }
.column-tag:nth-child(9) { animation-delay: 240ms; }
.column-tag:nth-child(10) { animation-delay: 270ms; }

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

/* SQL highlight animation */
.sql-highlight-enter {
  animation: expandIn 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes expandIn {
  from {
    opacity: 0;
    max-height: 0;
    transform: scaleY(0.95);
  }
  to {
    opacity: 1;
    max-height: 300px;
    transform: scaleY(1);
  }
}

/* Content fade animation */
.content-fade {
  animation: contentFade 0.3s ease-out;
}

@keyframes contentFade {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* Processing indicator animation */
.processing-indicator {
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

/* Field panel transition */
.field-panel-enter-active {
  animation: fieldPanelIn 0.35s cubic-bezier(0.16, 1, 0.3, 1);
}

.field-panel-leave-active {
  animation: fieldPanelOut 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes fieldPanelIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
    max-height: 0;
  }
  to {
    opacity: 1;
    transform: translateY(0);
    max-height: 500px;
  }
}

@keyframes fieldPanelOut {
  from {
    opacity: 1;
    transform: translateY(0);
  }
  to {
    opacity: 0;
    transform: translateY(-10px);
  }
}
</style>
