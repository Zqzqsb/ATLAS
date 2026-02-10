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

// Stage timing
const stageTimings = ref({
  vectorSearch: { start: 0, end: 0 },
  schemaLinking: { start: 0, end: 0 },
  sqlGeneration: { start: 0, end: 0 }
})

// Watch grounding stage changes to track vector search timing
watch(() => workspaceStore.groundingStage, (newStage, oldStage) => {
  if (newStage === 'stage1' && oldStage === 'idle') {
    stageTimings.value.vectorSearch.start = Date.now()
  } else if (newStage === 'done') {
    stageTimings.value.vectorSearch.end = Date.now()
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
const useFieldAlignment = ref(true)
const selectedModel = ref('deepseek_v3')
const useRichContext = ref(true)
const useReact = ref(true)
const useGrounding = ref(true)

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
  const stageDone = workspaceStore.groundingStage === 'done' || !!workspaceStore.groundingResult
  return {
    active: isExecuting.value && workspaceStore.groundingStage !== 'idle',
    completed: stageDone && hasGroundingContent.value,
    empty: stageDone && !hasGroundingContent.value,
    data: workspaceStore.groundingResult,
    duration: stageDone && start && end ? end - start : 0
  }
})

const schemaLinkingStage = computed(() => {
  const { start, end } = stageTimings.value.schemaLinking
  const steps = workspaceStore.reactSteps.filter(s => s.phase === 'schema_linking')
  const hasSchemaLinkingSteps = steps.length > 0
  const hasSqlGenerationSteps = workspaceStore.reactSteps.some(s => s.phase === 'sql_generation')
  const completed = end > 0 || hasSqlGenerationSteps || !!workspaceStore.generatedSql
  
  // Active when: grounding is done AND (no sql generation steps yet OR has schema linking steps)
  // Also active when awaiting field confirmation (field panel is shown inside this card)
  const groundingDone = workspaceStore.groundingStage === 'done' || !!workspaceStore.groundingResult
  const active = awaitingFieldConfirmation.value || (isExecuting.value && groundingDone && !hasSqlGenerationSteps)
  
  return {
    active: active || (isExecuting.value && hasSchemaLinkingSteps),
    completed: completed && hasSchemaLinkingSteps,
    steps,
    contexts: workspaceStore.usedContexts,
    duration: completed && start && end ? end - start : 0
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

// Field Alignment: watch grounding result for suggested fields
// When field alignment is on and we did grounding-only, show the panel and wait
watch(() => workspaceStore.groundingResult, (result) => {
  if (useFieldAlignment.value && result && result.suggestedFields && result.suggestedFields.length > 0) {
    suggestedFields.value = result.suggestedFields.map(f => ({ ...f }))
    showFieldPanel.value = true
    // If we're in grounding-only mode, mark that we're waiting for confirmation
    if (awaitingFieldConfirmation.value) {
      // Stream already completed (grounding_only), user needs to confirm
    }
  }
})

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

// Dismiss field panel and execute without field constraints
function dismissFieldPanel() {
  showFieldPanel.value = false
  if (awaitingFieldConfirmation.value) {
    awaitingFieldConfirmation.value = false
    // Execute full query without field constraints
    doExecuteFull()
  }
}

// Confirm field selection and execute full query with field constraints
async function confirmFieldsAndExecute() {
  showFieldPanel.value = false
  awaitingFieldConfirmation.value = false
  const fieldDesc = getFieldDescription()
  try {
    await workspaceStore.executeQuery(question.value, fieldDesc, false)
  } catch (e: any) {
    message.error(e.message || 'Execution failed')
  }
}

// Re-execute with updated field selections (for post-SQL adjustment)
async function reExecuteWithFields() {
  showFieldPanel.value = false
  const fieldDesc = getFieldDescription()
  workspaceStore.abortCurrentQuery()
  try {
    await workspaceStore.executeQuery(question.value, fieldDesc, false)
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
  
  workspaceStore.queryOptions.maxIterations = maxIterations.value
  workspaceStore.queryOptions.useRichContext = useRichContext.value
  workspaceStore.queryOptions.useReact = useReact.value
  workspaceStore.queryOptions.useGrounding = useGrounding.value

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
    <div class="control-panel mb-8 p-8 rounded-xl bg-white border border-gray-200 shadow-sm">
      <!-- Question Input -->
      <div class="mb-8">
        <div class="flex items-center gap-3 mb-4">
          <div class="w-10 h-10 rounded-xl bg-primary-50 flex items-center justify-center">
            <div class="i-carbon-chat text-xl text-primary-600" />
          </div>
          <div>
            <h3 class="text-lg font-bold text-gray-900">Natural Language Query</h3>
            <p class="text-sm text-gray-500 font-medium">Ask questions in plain English</p>
          </div>
        </div>
        
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
        <div class="mt-6 example-collapse">
          <NCollapse :default-expanded-names="['examples']" arrow-placement="left">
            <NCollapseItem name="examples">
              <template #header>
                <div class="flex items-center gap-2">
                  <span class="text-base font-semibold text-gray-700">Example Questions</span>
                  <span class="text-sm text-gray-400 font-medium">(Spider Benchmark)</span>
                </div>
              </template>
              <div class="flex flex-wrap gap-3 pt-2">
                <button
                  v-for="example in exampleQuestions"
                  :key="example"
                  class="text-base px-4 py-2.5 rounded-lg bg-gray-100 text-gray-600 hover:bg-primary-50 hover:text-primary-600 transition-all font-medium border border-gray-200 hover:border-primary-200 hover:shadow-sm"
                  @click="useExample(example)"
                >
                  {{ example }}
                </button>
              </div>
            </NCollapseItem>
          </NCollapse>
        </div>
      </div>

      <!-- Parameters -->
      <div class="grid grid-cols-2 lg:grid-cols-5 gap-6 p-6 bg-gray-50 rounded-xl border border-gray-100">
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
      </div>

      <!-- Action Buttons — no more field panel here, it moved to Schema Linking card -->
      <div class="flex items-center gap-4 mt-8">
        <button
          :disabled="!question.trim() || isExecuting"
          class="execute-btn flex items-center gap-3 px-8 py-3.5 rounded-xl text-white font-bold text-base transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
          :class="isExecuting 
            ? 'bg-gradient-to-r from-amber-500 to-orange-500 shadow-amber-500/30 animate-pulse' 
            : 'bg-gradient-to-r from-primary-500 to-blue-600 hover:from-primary-600 hover:to-blue-700 shadow-primary-500/30 hover:shadow-xl hover:shadow-primary-500/40 hover:-translate-y-0.5'"
          @click="handleExecute"
        >
          <div v-if="isExecuting" class="i-carbon-circle-dash animate-spin text-lg" />
          <div v-else class="i-carbon-play text-lg" />
          {{ isExecuting ? 'Executing...' : 'Execute Query' }}
        </button>

        <button
          v-if="isExecuting"
          class="flex items-center gap-2 px-6 py-3.5 rounded-xl bg-gradient-to-r from-red-500 to-rose-600 text-white font-bold text-base hover:from-red-600 hover:to-rose-700 shadow-lg shadow-red-500/30 hover:shadow-xl hover:-translate-y-0.5 transition-all duration-300"
          @click="handleStop"
        >
          <div class="i-carbon-stop text-lg" />
          Stop
        </button>

        <button
          :disabled="isExecuting"
          class="flex items-center gap-2 px-6 py-3.5 rounded-xl bg-white text-gray-700 font-bold text-base border-2 border-gray-200 hover:border-gray-300 hover:bg-gray-50 hover:text-gray-900 shadow-sm hover:shadow transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed"
          @click="handleClear"
        >
          <div class="i-carbon-clean text-lg" />
          Clear
        </button>
      </div>
    </div>

    <!-- Real-time Execution Cards -->
    <div class="execution-pipeline grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
      <!-- Stage 1: Vector Search -->
      <RealtimeCard
        title="Vector Search"
        icon="i-carbon-search"
        :active="vectorSearchStage.active"
        :stage="workspaceStore.groundingStage"
        :completed="vectorSearchStage.completed"
        :duration="vectorSearchStage.duration"
        color="blue"
      >
        <template #content>
          <!-- Skeleton screen: shows table names from local cache while waiting for backend -->
          <div v-if="workspaceStore.showSkeleton && workspaceStore.skeletonTables.length > 0" class="space-y-4 animate-pulse">
            <div class="flex items-center gap-2 mb-2">
              <div class="i-carbon-table-alias text-sm text-blue-400" />
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
            <div class="i-carbon-catalog text-3xl mb-2 opacity-40" />
            <span class="text-sm font-medium">No context available</span>
            <span class="text-xs mt-1 opacity-70">Generate Rich Context to enable vector retrieval</span>
          </div>
          <div v-else-if="workspaceStore.groundingResult" class="space-y-4 content-fade">
            <!-- Tables with confidence -->
            <div v-if="workspaceStore.groundingResult.tables?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-carbon-table-alias text-sm text-blue-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Retrieved Tables ({{ workspaceStore.groundingResult.tables.length }})</span>
              </div>
              <div class="space-y-2">
                <div
                  v-for="table in workspaceStore.groundingResult.tables"
                  :key="table.name"
                  class="grounding-item flex items-center justify-between px-3 py-2 rounded-lg bg-blue-50 border border-blue-100 hover:bg-blue-100/80 transition-colors"
                >
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
              </div>
            </div>

            <!-- Columns grouped by table -->
            <div v-if="workspaceStore.groundingResult.columns?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-carbon-column text-sm text-cyan-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Retrieved Columns ({{ workspaceStore.groundingResult.columns.length }})</span>
              </div>
              <div class="flex flex-wrap gap-1.5">
                <div
                  v-for="col in workspaceStore.groundingResult.columns.slice(0, 10)"
                  :key="`${col.table}.${col.column}`"
                  class="column-tag px-2 py-1 rounded bg-cyan-50 border border-cyan-100 text-xs font-medium hover:bg-cyan-100/80 transition-colors"
                >
                  <span class="text-gray-500">{{ col.table }}.</span><span class="text-cyan-700">{{ col.column }}</span>
                </div>
                <div v-if="workspaceStore.groundingResult.columns.length > 10" class="column-tag px-2 py-1 text-xs text-gray-400 font-medium">
                  +{{ workspaceStore.groundingResult.columns.length - 10 }} more
                </div>
              </div>
            </div>

            <!-- Join paths if any -->
            <div v-if="workspaceStore.groundingResult.joinPaths?.length">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-carbon-connect text-sm text-purple-600" />
                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Join Paths</span>
              </div>
              <div class="space-y-1">
                <div
                  v-for="(path, idx) in workspaceStore.groundingResult.joinPaths.slice(0, 3)"
                  :key="idx"
                  class="grounding-item flex items-center gap-2 text-xs text-gray-500 font-medium"
                >
                  <span class="text-purple-700">{{ path.from?.table }}.{{ path.from?.column }}</span>
                  <div class="i-carbon-arrow-right text-gray-400" />
                  <span class="text-purple-700">{{ path.to?.table }}.{{ path.to?.column }}</span>
                </div>
              </div>
            </div>

            <!-- LLM Reasoning (Fine Selection) -->
            <div v-if="workspaceStore.groundingResult.reasoning" class="mt-3">
              <NCollapse :default-expanded-names="['reasoning']" arrow-placement="left">
                <NCollapseItem name="reasoning">
                  <template #header>
                    <div class="flex items-center gap-2">
                      <div class="i-carbon-machine-learning-model text-sm text-indigo-500" />
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

            <!-- Execution Logs (Transparency) -->
            <div v-if="workspaceStore.groundingResult.executionLogs?.length" class="mt-3">
              <NCollapse :default-expanded-names="[]" arrow-placement="left">
                <NCollapseItem name="logs">
                  <template #header>
                    <div class="flex items-center gap-2">
                      <div class="i-carbon-terminal text-sm text-gray-500" />
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
          <div v-else-if="vectorSearchStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
            <div class="i-carbon-search animate-pulse text-blue-500 text-xl" />
            <div class="space-y-1">
              <span class="font-medium block">Searching vector database...</span>
              <span class="text-xs text-gray-400">Identifying relevant tables and columns</span>
            </div>
          </div>
          <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
            <div class="i-carbon-search text-3xl mb-2 opacity-30" />
            <span class="text-sm font-medium">Waiting for query...</span>
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 2: Schema Linking -->
      <RealtimeCard
        title="ReAct Schema Linking"
        icon="i-carbon-connection-signal"
        :active="schemaLinkingStage.active"
        :completed="schemaLinkingStage.completed"
        :duration="schemaLinkingStage.duration"
        color="cyan"
      >
        <template #content>
          <!-- Field Suggestions Panel — shown when grounding-only completes, BEFORE any linking steps -->
          <Transition name="field-panel">
            <div v-if="showFieldPanel && suggestedFields.length > 0" class="p-4 rounded-lg bg-gradient-to-br from-purple-50 to-violet-50 border border-purple-200/60">
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <div class="i-carbon-data-table text-purple-500 text-sm" />
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
                  <div class="i-carbon-close text-sm" />
                </button>
              </div>
              
              <!-- Compact field chips -->
              <div class="flex flex-wrap gap-2 mb-3">
                <button
                  v-for="field in suggestedFields"
                  :key="`${field.tableName}.${field.columnName}`"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer border"
                  :class="field.selected 
                    ? 'bg-purple-100 border-purple-300 text-purple-800 shadow-sm' 
                    : 'bg-white border-gray-200 text-gray-500 hover:border-purple-200'"
                  :title="field.reason"
                  @click="toggleField(field)"
                >
                  <NCheckbox :checked="field.selected" size="small" @update:checked="(v: boolean) => field.selected = v" @click.stop />
                  <span class="font-mono">{{ field.columnName }}</span>
                  <span class="text-gray-400 font-normal">{{ field.tableName }}</span>
                </button>
              </div>
              
              <!-- Action row — different buttons based on state -->
              <div class="flex items-center justify-between pt-2 border-t border-purple-100">
                <p class="text-xs text-gray-400">
                  <span class="i-carbon-information inline-block mr-0.5 align-middle" />
                  {{ awaitingFieldConfirmation 
                    ? 'Select the fields you want in the output, then confirm to generate SQL.' 
                    : 'Adjust fields and re-run to refine SQL output.' }}
                </p>
                <div class="flex items-center gap-2">
                  <button
                    v-if="awaitingFieldConfirmation"
                    class="px-3 py-1.5 rounded-lg bg-white border border-gray-200 text-gray-600 font-medium text-xs hover:bg-gray-50 transition-all"
                    @click="dismissFieldPanel"
                  >
                    Skip
                  </button>
                  <button
                    class="px-3 py-1.5 rounded-lg text-white font-bold text-xs shadow-sm hover:-translate-y-0.5 transition-all"
                    :class="awaitingFieldConfirmation 
                      ? 'bg-gradient-to-r from-primary-500 to-blue-600 hover:from-primary-600 hover:to-blue-700' 
                      : 'bg-purple-600 hover:bg-purple-700'"
                    :disabled="isExecuting"
                    @click="awaitingFieldConfirmation ? confirmFieldsAndExecute() : reExecuteWithFields()"
                  >
                    {{ awaitingFieldConfirmation ? 'Confirm & Generate SQL' : 'Re-run with Selection' }}
                  </button>
                </div>
              </div>
            </div>
          </Transition>

          <!-- Schema Linking Steps (only shown when inference is running / complete) -->
          <div v-if="schemaLinkingStage.steps.length" class="space-y-4" :class="{ 'mt-4': showFieldPanel }">
            <!-- Step counter -->
            <div class="flex items-center gap-2 pb-2 border-b border-gray-100">
              <div class="text-xs font-bold text-gray-500 uppercase tracking-wide">
                {{ schemaLinkingStage.steps.length }} reasoning step{{ schemaLinkingStage.steps.length > 1 ? 's' : '' }}
              </div>
            </div>
            
            <!-- Steps -->
            <div
              v-for="step in schemaLinkingStage.steps"
              :key="step.step"
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
                    <div class="i-carbon-idea text-cyan-600 mt-0.5 flex-shrink-0" />
                    <p class="text-sm text-gray-700 leading-relaxed font-medium">{{ step.thought }}</p>
                  </div>
                  
                  <!-- Action -->
                  <div v-if="step.action" class="flex items-start gap-2 bg-white p-2 rounded border border-cyan-100">
                    <div class="i-carbon-play-filled text-teal-600 mt-0.5 flex-shrink-0" />
                    <div>
                      <span class="text-xs text-teal-700 font-mono font-bold">{{ step.action }}</span>
                      <span v-if="step.actionInput" class="text-xs text-gray-500 ml-2 font-mono">
                        {{ typeof step.actionInput === 'object' ? JSON.stringify(step.actionInput) : step.actionInput }}
                      </span>
                    </div>
                  </div>
                  
                  <!-- Observation -->
                  <div v-if="step.observation" class="flex items-start gap-2">
                    <div class="i-carbon-view text-amber-500 mt-0.5 flex-shrink-0" />
                    <p class="text-xs text-gray-500 leading-relaxed">{{ step.observation }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <!-- Loading/Waiting states (only when field panel is NOT shown) -->
          <div v-else-if="!showFieldPanel">
            <div v-if="schemaLinkingStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
              <div class="i-carbon-connection-signal animate-pulse text-cyan-500 text-xl" />
              <div class="space-y-1">
                <span class="font-medium block">Analyzing schema structure...</span>
                <span class="text-xs text-gray-400">Identifying table relationships and join paths</span>
              </div>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
              <div class="i-carbon-connection-signal text-3xl mb-2 opacity-30" />
              <span class="text-sm font-medium">Waiting for schema linking...</span>
            </div>
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 3: SQL Generation -->
      <RealtimeCard
        title="ReAct SQL Generation"
        icon="i-carbon-code"
        :active="sqlGenerationStage.active"
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
              
              <div
                v-for="step in sqlGenerationStage.steps"
                :key="step.step"
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
                      <div class="i-carbon-idea text-purple-600 mt-0.5 flex-shrink-0" />
                      <p class="text-sm text-gray-700 leading-relaxed font-medium">{{ step.thought }}</p>
                    </div>
                    
                    <!-- Action: verify_sql with status badge + expandable execution plan -->
                    <div v-if="step.action === 'verify_sql'" class="space-y-2">
                      <div class="flex items-center gap-2 bg-white p-2 rounded border"
                        :class="step.observation?.startsWith('✅') ? 'border-green-200' : step.observation?.startsWith('❌') ? 'border-red-200' : 'border-purple-100'"
                      >
                        <div class="i-carbon-checkmark-outline mt-0.5 flex-shrink-0"
                          :class="step.observation?.startsWith('✅') ? 'text-green-600' : step.observation?.startsWith('❌') ? 'text-red-600' : 'text-pink-600'"
                        />
                        <span class="text-xs font-mono font-bold"
                          :class="step.observation?.startsWith('✅') ? 'text-green-700' : step.observation?.startsWith('❌') ? 'text-red-700' : 'text-pink-600'"
                        >verify_sql</span>
                        <NTag v-if="step.observation?.startsWith('✅')" size="tiny" type="success" round>PASSED</NTag>
                        <NTag v-else-if="step.observation?.startsWith('❌')" size="tiny" type="error" round>FAILED</NTag>
                        <NTag v-else-if="step.observation" size="tiny" type="warning" round>CHECKING</NTag>
                      </div>
                      
                      <!-- Expandable Execution Plan -->
                      <div v-if="step.observation" class="verify-result">
                        <NCollapse 
                          :default-expanded-names="step.observation?.startsWith('❌') ? ['explain'] : []" 
                          arrow-placement="left"
                        >
                          <NCollapseItem name="explain">
                            <template #header>
                              <div class="flex items-center gap-2">
                                <div class="i-carbon-analytics text-sm"
                                  :class="step.observation?.startsWith('✅') ? 'text-green-500' : 'text-red-500'"
                                />
                                <span class="text-xs font-bold text-gray-500 uppercase tracking-wide">Execution Plan</span>
                                <span class="text-xs text-gray-400">(click to expand)</span>
                              </div>
                            </template>
                            <div class="mt-1 rounded-lg overflow-hidden border"
                              :class="step.observation?.startsWith('✅') ? 'border-green-200' : 'border-red-200'"
                            >
                              <!-- Summary header -->
                              <div class="px-3 py-2 text-xs font-medium flex items-center gap-2"
                                :class="step.observation?.startsWith('✅') ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
                              >
                                <div :class="step.observation?.startsWith('✅') ? 'i-carbon-checkmark-filled' : 'i-carbon-warning-alt'" />
                                {{ step.observation?.startsWith('✅') ? 'Query passed verification' : 'Query has issues' }}
                              </div>
                              <!-- Full plan detail -->
                              <pre class="text-xs text-gray-600 font-mono whitespace-pre-wrap leading-relaxed p-3 max-h-64 overflow-y-auto"
                                :class="step.observation?.startsWith('✅') ? 'bg-green-50/50' : 'bg-red-50/50'"
                              >{{ step.observation }}</pre>
                            </div>
                          </NCollapseItem>
                        </NCollapse>
                      </div>
                    </div>
                    
                    <!-- Regular action (non-verify_sql) -->
                    <div v-else-if="step.action" class="flex items-start gap-2 bg-white p-2 rounded border border-purple-100">
                      <div class="i-carbon-play-filled text-pink-600 mt-0.5 flex-shrink-0" />
                      <span class="text-xs text-pink-600 font-mono font-bold">{{ step.action }}</span>
                    </div>
                    
                    <!-- Observation for non-verify_sql actions -->
                    <div v-if="step.observation && step.action !== 'verify_sql'" class="flex items-start gap-2">
                      <div class="i-carbon-view text-amber-500 mt-0.5 flex-shrink-0" />
                      <p class="text-xs text-gray-500 leading-relaxed">{{ step.observation }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- Generated SQL Preview -->
            <div v-if="sqlGenerationStage.sql" class="mt-4 p-4 rounded-lg bg-gray-900 border border-gray-800 shadow-inner sql-highlight-enter">
              <div class="flex items-center gap-2 mb-3 border-b border-gray-800 pb-2">
                <div class="i-carbon-checkmark-filled text-green-400" />
                <span class="text-xs text-green-400 font-bold uppercase tracking-wide">SQL Generated</span>
              </div>
              <pre class="text-xs text-gray-300 font-mono overflow-x-auto whitespace-pre-wrap">{{ sqlGenerationStage.sql.substring(0, 200) }}{{ sqlGenerationStage.sql.length > 200 ? '...' : '' }}</pre>
            </div>
            
            <!-- Loading state -->
            <div v-if="!sqlGenerationStage.steps.length && !sqlGenerationStage.sql">
              <div v-if="sqlGenerationStage.active" class="flex items-center gap-3 text-sm text-gray-600 processing-indicator">
                <div class="i-carbon-code animate-pulse text-purple-500 text-xl" />
                <div class="space-y-1">
                  <span class="font-medium block">Generating SQL query...</span>
                  <span class="text-xs text-gray-400">Building optimized query from context</span>
                </div>
              </div>
              <div v-else class="flex flex-col items-center justify-center py-8 text-gray-400">
                <div class="i-carbon-code text-3xl mb-2 opacity-30" />
                <span class="text-sm font-medium">Waiting for SQL generation...</span>
              </div>
            </div>
          </div>
        </template>
      </RealtimeCard>
    </div>

    <!-- Query Result -->
    <QueryResult
      v-if="workspaceStore.generatedSql || workspaceStore.queryError"
      :sql="workspaceStore.generatedSql"
      :error="workspaceStore.queryError"
      :duration="workspaceStore.queryDuration"
      :result="workspaceStore.executionResult"
      :loading="workspaceStore.isQuerying"
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
