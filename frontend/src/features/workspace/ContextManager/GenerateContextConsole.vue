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
const minIterations = ref(1)
const maxIterations = ref(3)

// State
const isRunning = ref(false)
const isComplete = ref(false)
const startTime = ref(0)
const elapsedTime = ref(0)
let elapsedTimer: number | null = null

// Agent states
interface AgentState {
  id: string
  table?: string
  status: 'pending' | 'running' | 'success' | 'error'
  phase: string
  progress: number
  message: string
}

const coordinatorState = ref<AgentState>({
  id: 'coordinator',
  status: 'pending',
  phase: '',
  progress: 0,
  message: ''
})

const workerStates = ref<Map<string, AgentState>>(new Map())

// Storage stats
const storageStats = ref({
  tablesTotal: 0,
  tablesUpdated: 0,
  columnsTotal: 0,
  columnsUpdated: 0
})

// Console logs
interface LogEntry {
  id: number
  timestamp: string
  agent: string
  message: string
  type: 'info' | 'success' | 'error' | 'storage'
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
  if (storageStats.value.tablesTotal === 0) return 0
  const tableProgress = (storageStats.value.tablesUpdated / storageStats.value.tablesTotal) * 50
  const columnProgress = storageStats.value.columnsTotal > 0 
    ? (storageStats.value.columnsUpdated / storageStats.value.columnsTotal) * 50 
    : 50
  return Math.round(tableProgress + columnProgress)
})

const formattedElapsed = computed(() => {
  const secs = Math.floor(elapsedTime.value / 1000)
  const mins = Math.floor(secs / 60)
  const remainSecs = secs % 60
  if (mins > 0) {
    return `${mins}m ${remainSecs}s`
  }
  return `${secs}s`
})

// Methods
function addLog(agent: string, msg: string, type: LogEntry['type'] = 'info') {
  const now = new Date()
  const ts = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  logs.value.push({
    id: ++logId,
    timestamp: ts,
    agent,
    message: msg,
    type
  })
  // Auto scroll
  nextTick(() => {
    const logArea = document.querySelector('.console-log-area')
    if (logArea) {
      logArea.scrollTop = logArea.scrollHeight
    }
  })
}

function getAgentIcon(agent: string): string {
  if (agent === 'coordinator') return '🎯'
  if (agent.startsWith('worker')) return '📊'
  if (agent === 'storage') return '💾'
  return '📝'
}

function getStatusIcon(status: string): string {
  switch (status) {
    case 'running': return '🔄'
    case 'success': return '✓'
    case 'error': return '✗'
    default: return '⏳'
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

function getLogColor(type: LogEntry['type']): string {
  switch (type) {
    case 'success': return 'text-green-400'
    case 'error': return 'text-red-400'
    case 'storage': return 'text-emerald-400'
    default: return 'text-gray-300'
  }
}

function getAgentColor(agent: string): string {
  if (agent === 'coordinator') return 'text-yellow-400'
  if (agent.startsWith('worker')) return 'text-cyan-400'
  if (agent === 'storage') return 'text-emerald-400'
  return 'text-gray-400'
}

async function startGeneration() {
  isRunning.value = true
  isComplete.value = false
  logs.value = []
  workerStates.value.clear()
  coordinatorState.value = { id: 'coordinator', status: 'pending', phase: '', progress: 0, message: '' }
  storageStats.value = { tablesTotal: 0, tablesUpdated: 0, columnsTotal: 0, columnsUpdated: 0 }
  
  startTime.value = Date.now()
  elapsedTimer = window.setInterval(() => {
    elapsedTime.value = Date.now() - startTime.value
  }, 100)

  addLog('system', `Starting generation with ${concurrency.value} workers, iterations: ${minIterations.value}-${maxIterations.value}...`, 'info')

  // Create SSE connection
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

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

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
          } catch (e) {
            // ignore parse errors
          }
          currentEventType = 'message'
        }
      }
    }
  } catch (e: any) {
    addLog('system', `Error: ${e.message}`, 'error')
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
      if (agent === 'coordinator') {
        coordinatorState.value = {
          id: 'coordinator',
          status: 'running',
          phase: data.phase || '',
          progress: 0,
          message: data.message || ''
        }
      } else {
        workerStates.value.set(agent, {
          id: agent,
          table: data.table,
          status: 'running',
          phase: data.phase || '',
          progress: 0,
          message: data.message || ''
        })
      }
      addLog(agent, data.message || 'Started', 'info')
      break

    case 'agent_step':
      if (agent === 'coordinator') {
        coordinatorState.value.phase = data.phase || coordinatorState.value.phase
        coordinatorState.value.message = data.message || ''
        if (data.data?.tables) {
          storageStats.value.tablesTotal = data.data.tables.length
        }
        if (data.data?.total_columns) {
          storageStats.value.columnsTotal = data.data.total_columns
        }
      } else {
        const state = workerStates.value.get(agent)
        if (state) {
          state.phase = data.phase || state.phase
          state.message = data.message || ''
          state.progress = Math.min(state.progress + 25, 90)
        }
      }
      addLog(agent, data.message || '', 'info')
      break

    case 'agent_done':
      if (agent === 'coordinator') {
        coordinatorState.value.status = 'success'
        coordinatorState.value.progress = 100
      } else {
        const state = workerStates.value.get(agent)
        if (state) {
          state.status = data.status === 'error' ? 'error' : 'success'
          state.progress = 100
        }
      }
      addLog(agent, data.message || 'Done', data.status === 'error' ? 'error' : 'success')
      break

    case 'storage':
      if (data.data?.target === 'rc_tables') {
        storageStats.value.tablesUpdated++
      } else if (data.data?.target === 'rc_columns') {
        storageStats.value.columnsUpdated++
      }
      addLog('storage', data.message || 'Saved', 'storage')
      break

    case 'complete':
      isComplete.value = true
      addLog('system', `Complete! Tables: ${data.data?.tables_updated || 0}, Columns: ${data.data?.columns_updated || 0}`, 'success')
      message.success('Rich Context generation complete!')
      emit('complete')
      break

    case 'error':
      addLog('system', data.message || 'Error', 'error')
      break
  }
}

// Minimize to background (keep running)
function handleMinimize() {
  showModal.value = false
  emit('minimize')
  message.info('Task running in background. Click the indicator to view progress.')
}

// Cancel and close
function handleCancel() {
  // Abort fetch if running
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
  if (eventSource) {
    eventSource.close()
  }
  if (elapsedTimer) {
    clearInterval(elapsedTimer)
  }
})
</script>

<template>
  <NModal
    v-model:show="showModal"
    preset="card"
    :closable="true"
    :mask-closable="!isRunning"
    :on-close="isRunning ? handleMinimize : undefined"
    style="width: 800px; max-width: 90vw;"
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
            <label class="text-sm text-gray-400 mb-2 block">Worker Concurrency</label>
            <NInputNumber
              v-model:value="concurrency"
              :min="1"
              :max="10"
              size="small"
            />
          </div>
          <div class="flex items-end">
            <div class="flex items-center gap-2">
              <NSwitch v-model:value="forceRegenerate" size="small" />
              <span class="text-sm text-gray-300">Force Regenerate</span>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Min Iterations</label>
            <NInputNumber
              v-model:value="minIterations"
              :min="1"
              :max="maxIterations - 1"
              size="small"
            />
          </div>
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Max Iterations</label>
            <NInputNumber
              v-model:value="maxIterations"
              :min="minIterations + 1"
              size="small"
            />
          </div>
        </div>
        <p class="text-xs text-gray-500 mt-2 mb-4">
          Iterations control the depth of context analysis. Higher values produce richer context but take longer. (0 &lt; min &lt; max)
        </p>
        
        <NButton
          type="primary"
          size="large"
          class="mt-4 w-full"
          @click="startGeneration"
        >
          <template #icon>
            <div class="i-carbon-play" />
          </template>
          Start Generation
        </NButton>
      </div>

      <!-- Running/Complete view -->
      <div v-else>
        <!-- Agent Status Cards -->
        <div class="agent-status-section mb-4">
          <h4 class="text-sm text-gray-400 mb-3">Agent Status</h4>
          
          <!-- Coordinator -->
          <div class="agent-card coordinator mb-3 p-3 rounded-lg bg-yellow-500/10 border border-yellow-500/30">
            <div class="flex items-center justify-between mb-2">
              <div class="flex items-center gap-2">
                <span class="text-lg">🎯</span>
                <span class="font-medium text-yellow-300">Coordinator</span>
              </div>
              <span :class="['text-sm', getStatusColor(coordinatorState.status)]">
                {{ getStatusIcon(coordinatorState.status) }} {{ coordinatorState.status }}
              </span>
            </div>
            <NProgress
              :percentage="coordinatorState.progress"
              :color="coordinatorState.status === 'success' ? '#22c55e' : '#eab308'"
              :height="6"
            />
          </div>

          <!-- Workers Grid -->
          <div class="workers-grid grid grid-cols-3 gap-2">
            <div
              v-for="[id, state] in workerStates"
              :key="id"
              class="worker-card p-2 rounded-lg bg-cyan-500/10 border border-cyan-500/30"
            >
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs text-cyan-300 truncate">{{ state.table || id }}</span>
                <span :class="['text-xs', getStatusColor(state.status)]">
                  {{ getStatusIcon(state.status) }}
                </span>
              </div>
              <NProgress
                :percentage="state.progress"
                :color="state.status === 'success' ? '#22c55e' : '#06b6d4'"
                :height="4"
              />
              <div class="text-xs text-gray-500 mt-1 truncate">{{ state.phase }}</div>
            </div>
          </div>
        </div>

        <!-- Storage Stats -->
        <div class="storage-section mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30">
          <h4 class="text-sm text-emerald-300 mb-2 flex items-center gap-2">
            <span>💾</span> Storage Activity
          </h4>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-gray-400 mb-1">rc_tables</div>
              <NProgress
                :percentage="storageStats.tablesTotal ? (storageStats.tablesUpdated / storageStats.tablesTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ storageStats.tablesUpdated }}/{{ storageStats.tablesTotal }}</span>
              </NProgress>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">rc_columns</div>
              <NProgress
                :percentage="storageStats.columnsTotal ? (storageStats.columnsUpdated / storageStats.columnsTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ storageStats.columnsUpdated }}/{{ storageStats.columnsTotal }}</span>
              </NProgress>
            </div>
          </div>
        </div>

        <!-- Console Output -->
        <div class="console-section">
          <h4 class="text-sm text-gray-400 mb-2 flex items-center gap-2">
            <span>📋</span> Console Output
          </h4>
          <div class="console-log-area h-48 overflow-y-auto bg-gray-900/80 rounded-lg p-3 font-mono text-xs">
            <div
              v-for="log in logs"
              :key="log.id"
              class="log-entry py-0.5"
            >
              <span class="text-gray-500">[{{ log.timestamp }}]</span>
              <span :class="['ml-2', getAgentColor(log.agent)]">{{ getAgentIcon(log.agent) }}</span>
              <span :class="['ml-1', getLogColor(log.type)]">{{ log.message }}</span>
            </div>
            <div v-if="isRunning" class="cursor-blink inline-block w-2 h-4 bg-cyan-400 ml-1 animate-pulse" />
          </div>
        </div>

        <!-- Footer Stats -->
        <div class="footer-stats mt-4 flex items-center justify-between text-sm text-gray-400">
          <span>⏱️ Elapsed: {{ formattedElapsed }}</span>
          <span>📊 Tables: {{ storageStats.tablesUpdated }}/{{ storageStats.tablesTotal }}</span>
          <span>📝 Columns: {{ storageStats.columnsUpdated }}/{{ storageStats.columnsTotal }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-between">
        <div>
          <NButton v-if="isRunning" quaternary size="small" @click="handleMinimize">
            <template #icon>
              <span class="i-carbon-minimize" />
            </template>
            Minimize to Background
          </NButton>
        </div>
        <div class="flex gap-2">
          <NButton v-if="isRunning" type="error" @click="handleCancel">
            Cancel
          </NButton>
          <NButton v-else @click="handleClose">
            Close
          </NButton>
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
