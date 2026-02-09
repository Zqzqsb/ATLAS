<script setup lang="ts">
import { ref, computed, onUnmounted, nextTick } from 'vue'
import { NModal, NButton, NInputNumber, NProgress, NSwitch, useMessage } from 'naive-ui'

const props = defineProps<{
  show: boolean
  databaseId: string
}>()

// Expose state for parent component
defineExpose({
  isRunning: computed(() => isRunning.value),
  isComplete: computed(() => isComplete.value),
  progress: computed(() => overallProgress.value)
})

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'complete'): void
  (e: 'minimize'): void
}>()

const message = useMessage()

// Configuration
const concurrency = ref(3)
const forceRegenerate = ref(false)
const minIterations = ref(3)
const maxIterations = ref(15)

// State
const isRunning = ref(false)
const isComplete = ref(false)
const startTime = ref(0)
const elapsedTime = ref(0)
let elapsedTimer: number | null = null

// Agent state (single rc_gen agent, not coordinator/worker pattern)
interface AgentState {
  id: string
  status: 'pending' | 'running' | 'success' | 'error'
  phase: string
  progress: number
  iteration: number
  message: string
}

const agentState = ref<AgentState>({
  id: 'rc_gen',
  status: 'pending',
  phase: '',
  progress: 0,
  iteration: 0,
  message: ''
})

// Embedding agent state
const embeddingState = ref<AgentState>({
  id: 'embedding',
  status: 'pending',
  phase: '',
  progress: 0,
  iteration: 0,
  message: ''
})

// Storage stats
const storageStats = ref({
  tablesTotal: 0,
  tablesUpdated: 0,
  columnsTotal: 0,
  columnsUpdated: 0,
  termsAdded: 0,
  sampleValuesAdded: 0,
  synonymsAdded: 0,
  embeddingsStreamed: 0
})

// Console logs
interface LogEntry {
  id: number
  timestamp: string
  phase: 'thought' | 'action' | 'observation' | 'storage' | 'info' | 'success' | 'error' | 'finish'
  agent: string
  message: string
  detail?: string  // e.g. action_input or truncated observation
}

const logs = ref<LogEntry[]>([])
let logId = 0

// SSE connection
let eventSource: EventSource | null = null

// Computed
const showModal = computed({
  get: () => props.show,
  set: (v) => emit('update:show', v)
})

const overallProgress = computed(() => {
  if (!isRunning.value && !isComplete.value) return 0
  // Weighted: agent 70%, embedding 30%
  const agentPct = agentState.value.progress * 0.7
  const embPct = embeddingState.value.progress * 0.3
  return Math.round(agentPct + embPct)
})

const formattedElapsed = computed(() => {
  const secs = Math.floor(elapsedTime.value / 1000)
  const mins = Math.floor(secs / 60)
  const remainSecs = secs % 60
  if (mins > 0) return `${mins}m ${remainSecs}s`
  return `${secs}s`
})

const totalContextWrites = computed(() => {
  return storageStats.value.tablesUpdated +
    storageStats.value.columnsUpdated +
    storageStats.value.termsAdded +
    storageStats.value.sampleValuesAdded +
    storageStats.value.synonymsAdded
})

// Methods
function addLog(phase: LogEntry['phase'], agent: string, msg: string, detail?: string) {
  if (!msg) return
  const now = new Date()
  const ts = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  logs.value.push({
    id: ++logId,
    timestamp: ts,
    phase,
    agent,
    message: msg,
    detail
  })
  // Auto scroll
  nextTick(() => {
    const logArea = document.querySelector('.console-log-area')
    if (logArea) logArea.scrollTop = logArea.scrollHeight
  })
}

function getPhaseIcon(phase: LogEntry['phase']): string {
  switch (phase) {
    case 'thought': return '💭'
    case 'action': return '🔧'
    case 'observation': return '📊'
    case 'storage': return '💾'
    case 'success': return '✅'
    case 'error': return '❌'
    case 'finish': return '🏁'
    default: return '📝'
  }
}

function getPhaseColor(phase: LogEntry['phase']): string {
  switch (phase) {
    case 'thought': return 'text-gray-300'
    case 'action': return 'text-blue-400'
    case 'observation': return 'text-cyan-300'
    case 'storage': return 'text-emerald-400'
    case 'success': return 'text-green-400'
    case 'error': return 'text-red-400'
    case 'finish': return 'text-yellow-300'
    default: return 'text-gray-400'
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'running': return 'text-blue-400'
    case 'success': return 'text-green-400'
    case 'error': return 'text-red-400'
    default: return 'text-gray-500'
  }
}

function getStatusIcon(status: string): string {
  switch (status) {
    case 'running': return '🔄'
    case 'success': return '✓'
    case 'error': return '✗'
    default: return '⏳'
  }
}

async function startGeneration() {
  isRunning.value = true
  isComplete.value = false
  logs.value = []
  agentState.value = { id: 'rc_gen', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
  embeddingState.value = { id: 'embedding', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
  storageStats.value = { tablesTotal: 0, tablesUpdated: 0, columnsTotal: 0, columnsUpdated: 0, termsAdded: 0, sampleValuesAdded: 0, synonymsAdded: 0, embeddingsStreamed: 0 }
  
  startTime.value = Date.now()
  elapsedTimer = window.setInterval(() => {
    elapsedTime.value = Date.now() - startTime.value
  }, 100)

  addLog('info', 'system', `Starting generation, iterations: ${minIterations.value}-${maxIterations.value}...`)

  const url = `/api/v1/lakebase/datasources/${props.databaseId}/generate-context`
  
  try {
    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        concurrency: concurrency.value,
        force: forceRegenerate.value,
        min_iterations: minIterations.value,
        max_iterations: maxIterations.value
      })
    })

    if (!response.ok) throw new Error(`HTTP ${response.status}`)

    const reader = response.body?.getReader()
    if (!reader) throw new Error('No response body')

    const decoder = new TextDecoder()
    let buffer = ''
    let currentEventType = 'message'

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('event: ')) {
          currentEventType = line.slice(7).trim()
        } else if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6))
            handleEvent(currentEventType, data)
          } catch {
            // ignore parse errors
          }
          currentEventType = 'message'
        }
      }
    }
  } catch (e: any) {
    addLog('error', 'system', `Error: ${e.message}`)
    message.error('Generation failed: ' + e.message)
  } finally {
    isRunning.value = false
    if (elapsedTimer) {
      clearInterval(elapsedTimer)
      elapsedTimer = null
    }
  }
}

function handleEvent(eventType: string, data: any) {
  const agent = data.agent || 'system'
  
  switch (eventType) {
    case 'agent_start':
      if (agent === 'rc_gen') {
        agentState.value.status = 'running'
        agentState.value.phase = data.phase || 'init'
        agentState.value.message = data.message || ''
        // Set totals from initial event
        if (data.data?.tables_total) {
          storageStats.value.tablesTotal = data.data.tables_total
        }
        if (data.data?.columns_total) {
          storageStats.value.columnsTotal = data.data.columns_total
        }
      } else if (agent === 'embedding') {
        embeddingState.value.status = 'running'
        embeddingState.value.message = data.message || ''
      }
      addLog('info', agent, data.message || 'Started')
      break

    case 'agent_step': {
      const phase = data.phase || 'thought'
      const iter = data.data?.iteration || 0
      
      if (agent === 'rc_gen') {
        agentState.value.iteration = iter
        agentState.value.phase = phase
        agentState.value.message = data.message || ''
        // Update progress based on iteration (rough estimate)
        if (maxIterations.value > 0) {
          agentState.value.progress = Math.min(Math.round((iter / maxIterations.value) * 100), 95)
        }
      } else if (agent === 'embedding' && phase === 'embedding') {
        // Incremental embedding event
        const embSoFar = data.data?.embeddings_so_far || 0
        if (embSoFar > 0) {
          storageStats.value.embeddingsStreamed = embSoFar
          // Update embedding progress based on total writes
          const totalWrites = totalContextWrites.value
          if (totalWrites > 0) {
            embeddingState.value.progress = Math.min(Math.round((embSoFar / totalWrites) * 100), 95)
          }
          embeddingState.value.message = `Streamed ${embSoFar} embeddings`
        }
        addLog('storage', 'embedding', data.message || '')
        break
      }

      if (phase === 'thought' && data.message) {
        addLog('thought', agent, data.message)
      } else if (phase === 'action') {
        const actionName = data.data?.action || ''
        const actionInput = data.data?.action_input || ''
        // Truncate action_input for display
        const inputPreview = actionInput.length > 120 ? actionInput.slice(0, 120) + '...' : actionInput
        addLog('action', agent, `${actionName}`, inputPreview)
      } else if (phase === 'observation') {
        addLog('observation', agent, data.message || '')
      } else if (phase === 'finish') {
        addLog('finish', agent, data.message || 'Agent finished')
      }
      break
    }

    case 'agent_done':
      if (agent === 'rc_gen') {
        agentState.value.status = data.status === 'error' ? 'error' : 'success'
        agentState.value.progress = 100
        agentState.value.message = data.message || 'Done'
      } else if (agent === 'embedding') {
        embeddingState.value.status = data.status === 'error' ? 'error' : 'success'
        embeddingState.value.progress = 100
        embeddingState.value.message = data.message || 'Done'
      }
      addLog(data.status === 'error' ? 'error' : 'success', agent, data.message || 'Done')
      break

    case 'storage': {
      const contextType = data.data?.context_type || ''
      const target = data.data?.target || ''
      
      if (target === 'rc_tables') {
        storageStats.value.tablesUpdated++
      } else if (target === 'rc_columns') {
        // Track by sub-type
        if (contextType === 'column_sample_values') {
          storageStats.value.sampleValuesAdded++
        } else if (contextType === 'column_synonyms') {
          storageStats.value.synonymsAdded++
        } else {
          storageStats.value.columnsUpdated++
        }
      } else if (target === 'rc_terms') {
        storageStats.value.termsAdded++
      }
      addLog('storage', 'storage', data.message || 'Saved')
      break
    }

    case 'complete':
      isComplete.value = true
      addLog('success', 'system',
        `Complete! Iterations: ${data.data?.react_iterations || 0}, Embeddings: ${data.data?.embeddings_generated || 0}, Duration: ${Math.round((data.data?.duration_ms || 0) / 1000)}s`)
      message.success('Rich Context generation complete!')
      emit('complete')
      break

    case 'error':
      addLog('error', 'system', data.message || 'Error')
      break
  }
}

function handleMinimize() {
  showModal.value = false
  emit('minimize')
  message.info('Task running in background.')
}

function handleCancel() {
  if (isRunning.value) {
    isRunning.value = false
    if (elapsedTimer) {
      clearInterval(elapsedTimer)
      elapsedTimer = null
    }
  }
  showModal.value = false
}

function handleClose() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  showModal.value = false
}

onUnmounted(() => {
  if (eventSource) eventSource.close()
  if (elapsedTimer) clearInterval(elapsedTimer)
})
</script>

<template>
  <NModal
    v-model:show="showModal"
    preset="card"
    :closable="true"
    :mask-closable="!isRunning"
    :on-close="isRunning ? handleMinimize : undefined"
    style="width: 850px; max-width: 90vw;"
    class="generate-console-modal"
  >
    <template #header>
      <div class="flex items-center gap-2">
        <span>Rich Context Generation</span>
        <span v-if="isRunning" class="px-2 py-0.5 text-xs rounded-full bg-blue-500/20 text-blue-400 animate-pulse">
          Running
        </span>
        <span v-else-if="isComplete" class="px-2 py-0.5 text-xs rounded-full bg-green-500/20 text-green-400">
          Complete
        </span>
      </div>
    </template>

    <div class="generate-console">
      <!-- Configuration (shown before start) -->
      <div v-if="!isRunning && !isComplete" class="config-section mb-6">
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Min Iterations</label>
            <NInputNumber v-model:value="minIterations" :min="1" :max="maxIterations - 1" size="small" />
          </div>
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Max Iterations</label>
            <NInputNumber v-model:value="maxIterations" :min="minIterations + 1" :max="50" size="small" />
          </div>
        </div>
        <div class="flex items-center gap-4 mb-4">
          <div class="flex items-center gap-2">
            <NSwitch v-model:value="forceRegenerate" size="small" />
            <span class="text-sm text-gray-300">Force Regenerate</span>
          </div>
        </div>
        <p class="text-xs text-gray-500 mb-4">
          More iterations = deeper analysis. The agent explores tables, discovers data patterns, and saves descriptions, sample values, synonyms, and business terms.
        </p>
        <NButton type="primary" size="large" class="w-full" @click="startGeneration">
          <template #icon><div class="i-carbon-play" /></template>
          Start Generation
        </NButton>
      </div>

      <!-- Running/Complete view -->
      <div v-else>
        <!-- Agent Status -->
        <div class="agent-status-section mb-4">
          <h4 class="text-sm text-gray-400 mb-3">Agent Status</h4>
          <div class="grid grid-cols-2 gap-3">
            <!-- RC Gen Agent -->
            <div class="agent-card p-3 rounded-lg bg-blue-500/10 border border-blue-500/30">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-lg">🤖</span>
                  <span class="font-medium text-blue-300 text-sm">RC Generator</span>
                </div>
                <span :class="['text-xs', getStatusColor(agentState.status)]">
                  {{ getStatusIcon(agentState.status) }} {{ agentState.status }}
                </span>
              </div>
              <NProgress
                :percentage="agentState.progress"
                :color="agentState.status === 'success' ? '#22c55e' : agentState.status === 'error' ? '#ef4444' : '#3b82f6'"
                :height="6"
              />
              <div class="text-xs text-gray-500 mt-1">
                <span v-if="agentState.iteration > 0">Iteration {{ agentState.iteration }}</span>
                <span v-if="agentState.phase" class="ml-2">· {{ agentState.phase }}</span>
              </div>
            </div>

            <!-- Embedding Agent -->
            <div class="agent-card p-3 rounded-lg bg-purple-500/10 border border-purple-500/30">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-lg">🧬</span>
                  <span class="font-medium text-purple-300 text-sm">Embeddings</span>
                </div>
                <span :class="['text-xs', getStatusColor(embeddingState.status)]">
                  {{ getStatusIcon(embeddingState.status) }} {{ embeddingState.status }}
                </span>
              </div>
              <NProgress
                :percentage="embeddingState.progress"
                :color="embeddingState.status === 'success' ? '#22c55e' : embeddingState.status === 'error' ? '#ef4444' : '#a855f7'"
                :height="6"
              />
              <div class="text-xs text-gray-500 mt-1 truncate">{{ embeddingState.message }}</div>
            </div>
          </div>
        </div>

        <!-- Storage Stats -->
        <div class="storage-section mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30">
          <h4 class="text-sm text-emerald-300 mb-2 flex items-center gap-2">
            <span>💾</span> Storage Activity
            <span class="text-xs text-gray-500 ml-auto">{{ totalContextWrites }} writes</span>
          </h4>
          <div class="grid grid-cols-2 gap-4 mb-2">
            <div>
              <div class="text-xs text-gray-400 mb-1">Table Descriptions</div>
              <NProgress
                :percentage="storageStats.tablesTotal ? (storageStats.tablesUpdated / storageStats.tablesTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ storageStats.tablesUpdated }}/{{ storageStats.tablesTotal }}</span>
              </NProgress>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">Column Descriptions</div>
              <NProgress
                :percentage="storageStats.columnsTotal ? (storageStats.columnsUpdated / storageStats.columnsTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ storageStats.columnsUpdated }}/{{ storageStats.columnsTotal }}</span>
              </NProgress>
            </div>
          </div>
          <div class="flex gap-4 text-xs text-gray-500">
            <span v-if="storageStats.sampleValuesAdded > 0">📋 Sample Values: {{ storageStats.sampleValuesAdded }}</span>
            <span v-if="storageStats.synonymsAdded > 0">🔗 Synonyms: {{ storageStats.synonymsAdded }}</span>
            <span v-if="storageStats.termsAdded > 0">📖 Terms: {{ storageStats.termsAdded }}</span>
            <span v-if="storageStats.embeddingsStreamed > 0">🧬 Embeddings: {{ storageStats.embeddingsStreamed }}</span>
          </div>
        </div>

        <!-- Console Output -->
        <div class="console-section">
          <h4 class="text-sm text-gray-400 mb-2 flex items-center gap-2">
            <span>📋</span> Console Output
            <!-- Legend -->
            <span class="ml-auto flex items-center gap-3 text-xs text-gray-500">
              <span class="flex items-center gap-1"><span>💭</span><span class="text-gray-400">Thought</span></span>
              <span class="flex items-center gap-1"><span>🔧</span><span class="text-blue-400">Action</span></span>
              <span class="flex items-center gap-1"><span>📊</span><span class="text-cyan-400">Result</span></span>
              <span class="flex items-center gap-1"><span>💾</span><span class="text-emerald-400">Saved</span></span>
            </span>
          </h4>
          <div class="console-log-area h-56 overflow-y-auto bg-gray-900/80 rounded-lg p-3 font-mono text-xs leading-relaxed">
            <div v-for="log in logs" :key="log.id" class="log-entry py-0.5">
              <span class="text-gray-600">[{{ log.timestamp }}]</span>
              <span class="ml-1">{{ getPhaseIcon(log.phase) }}</span>
              <span :class="['ml-1', getPhaseColor(log.phase)]">{{ log.message }}</span>
              <div v-if="log.detail" class="ml-8 text-gray-600 break-all">{{ log.detail }}</div>
            </div>
            <div v-if="isRunning" class="cursor-blink inline-block w-2 h-4 bg-cyan-400 ml-1" />
          </div>
        </div>

        <!-- Footer Stats -->
        <div class="footer-stats mt-4 flex items-center justify-between text-sm text-gray-400">
          <span>⏱️ {{ formattedElapsed }}</span>
          <span>📊 Tables: {{ storageStats.tablesUpdated }}/{{ storageStats.tablesTotal }}</span>
          <span>📝 Columns: {{ storageStats.columnsUpdated }}/{{ storageStats.columnsTotal }}</span>
          <span>💾 Writes: {{ totalContextWrites }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-between">
        <div>
          <NButton v-if="isRunning" quaternary size="small" @click="handleMinimize">
            <template #icon><span class="i-carbon-minimize" /></template>
            Minimize to Background
          </NButton>
        </div>
        <div class="flex gap-2">
          <NButton v-if="isRunning" type="error" @click="handleCancel">Cancel</NButton>
          <NButton v-else @click="handleClose">Close</NButton>
        </div>
      </div>
    </template>
  </NModal>
</template>

<style scoped>
.generate-console-modal :deep(.n-card) {
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.95), rgba(30, 41, 59, 0.95));
  border: 1px solid rgba(100, 116, 139, 0.3);
}

.console-log-area::-webkit-scrollbar {
  width: 6px;
}
.console-log-area::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}
.console-log-area::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.cursor-blink {
  animation: blink 1s infinite;
}
@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}
</style>
