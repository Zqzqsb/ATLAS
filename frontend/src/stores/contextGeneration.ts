import { defineStore } from 'pinia'
import { ref, computed, nextTick, watch } from 'vue'

// Agent state interface
export interface AgentState {
  id: string
  status: 'pending' | 'running' | 'success' | 'error'
  phase: string
  progress: number
  iteration: number
  message: string
}

// Log entry interface
export interface LogEntry {
  id: number
  timestamp: string
  phase: 'thought' | 'action' | 'observation' | 'storage' | 'info' | 'success' | 'error' | 'finish'
  agent: string
  message: string
  detail?: string
}

// Storage stats interface
export interface StorageStats {
  tablesTotal: number
  tablesUpdated: number
  columnsTotal: number
  columnsUpdated: number
  termsAdded: number
  sampleValuesAdded: number
  synonymsAdded: number
  embeddingsStreamed: number
  embeddingsTotal: number
}

// Persisted state shape (saved to sessionStorage)
interface PersistedState {
  isRunning: boolean
  isComplete: boolean
  isMinimized: boolean
  databaseId: string
  startTime: number
  agentState: AgentState
  embeddingState: AgentState
  storageStats: StorageStats
  config: { concurrency: number; force: boolean; minIterations: number; maxIterations: number }
  logs: LogEntry[]
}

const STORAGE_KEY = 'lucid_context_generation'

function loadPersistedState(): PersistedState | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch { /* ignore */ }
  return null
}

function savePersistedState(state: PersistedState) {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  } catch { /* ignore */ }
}

function clearPersistedState() {
  sessionStorage.removeItem(STORAGE_KEY)
}

export const useContextGenerationStore = defineStore('contextGeneration', () => {
  const persisted = loadPersistedState()

  // Core state
  const isRunning = ref(persisted?.isRunning ?? false)
  const isComplete = ref(persisted?.isComplete ?? false)
  const isMinimized = ref(persisted?.isMinimized ?? false)
  const showConsole = ref(false)
  const databaseId = ref(persisted?.databaseId ?? '')
  const startTime = ref(persisted?.startTime ?? 0)
  const elapsedTime = ref(0)
  let elapsedTimer: number | null = null

  // Config
  const config = ref(persisted?.config ?? {
    concurrency: 3,
    force: false,
    minIterations: 3,
    maxIterations: 15
  })

  // Agent states
  const agentState = ref<AgentState>(persisted?.agentState ?? {
    id: 'rc_gen', status: 'pending', phase: '', progress: 0, iteration: 0, message: ''
  })

  const embeddingState = ref<AgentState>(persisted?.embeddingState ?? {
    id: 'embedding', status: 'pending', phase: '', progress: 0, iteration: 0, message: ''
  })

  // Storage stats
  const storageStats = ref<StorageStats>(persisted?.storageStats ?? {
    tablesTotal: 0, tablesUpdated: 0, columnsTotal: 0, columnsUpdated: 0,
    termsAdded: 0, sampleValuesAdded: 0, synonymsAdded: 0, embeddingsStreamed: 0, embeddingsTotal: 0
  })

  // Logs
  const logs = ref<LogEntry[]>(persisted?.logs ?? [])
  let logId = logs.value.length > 0 ? Math.max(...logs.value.map(l => l.id)) : 0

  // SSE reader abort controller
  let abortController: AbortController | null = null

  // Computed
  const overallProgress = computed(() => {
    if (!isRunning.value && !isComplete.value) return 0
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

  // Persist state on every meaningful change
  function persist() {
    savePersistedState({
      isRunning: isRunning.value,
      isComplete: isComplete.value,
      isMinimized: isMinimized.value,
      databaseId: databaseId.value,
      startTime: startTime.value,
      agentState: { ...agentState.value },
      embeddingState: { ...embeddingState.value },
      storageStats: { ...storageStats.value },
      config: { ...config.value },
      // Only keep last 200 logs to avoid storage overflow
      logs: logs.value.slice(-200)
    })
  }

  // Log helper
  function addLog(phase: LogEntry['phase'], agent: string, msg: string, detail?: string) {
    if (!msg) return
    const now = new Date()
    const ts = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
    logs.value.push({ id: ++logId, timestamp: ts, phase, agent, message: msg, detail })
    // Keep last 500 in memory
    if (logs.value.length > 500) {
      logs.value = logs.value.slice(-500)
    }
    // Auto scroll console
    nextTick(() => {
      const logArea = document.querySelector('.console-log-area')
      if (logArea) logArea.scrollTop = logArea.scrollHeight
    })
  }

  // SSE Event handler
  function handleEvent(eventType: string, data: any) {
    const agent = data.agent || 'system'

    switch (eventType) {
      case 'agent_start':
        if (agent === 'rc_gen') {
          agentState.value.status = 'running'
          agentState.value.phase = data.phase || 'init'
          agentState.value.message = data.message || ''
          if (data.data?.tables_total) storageStats.value.tablesTotal = data.data.tables_total
          if (data.data?.columns_total) storageStats.value.columnsTotal = data.data.columns_total
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
          if (config.value.maxIterations > 0) {
            agentState.value.progress = Math.min(Math.round((iter / config.value.maxIterations) * 100), 95)
          }
        } else if (agent === 'embedding' && phase === 'embedding') {
          const embSoFar = data.data?.embeddings_so_far || 0
          const embTotal = data.data?.embeddings_total || 0
          if (embSoFar > 0) {
            storageStats.value.embeddingsStreamed = embSoFar
            if (embTotal > 0) {
              storageStats.value.embeddingsTotal = embTotal
              embeddingState.value.progress = Math.min(Math.round((embSoFar / embTotal) * 100), 95)
            }
            embeddingState.value.message = embTotal > 0
              ? `Embedded ${embSoFar}/${embTotal}`
              : `Streamed ${embSoFar} embeddings`
          }
          addLog('storage', 'embedding', data.message || '')
          break
        }

        if (phase === 'thought' && data.message) {
          addLog('thought', agent, data.message)
        } else if (phase === 'action') {
          const actionName = data.data?.action || ''
          const actionInput = data.data?.action_input || ''
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
          if (contextType === 'column_sample_values') storageStats.value.sampleValuesAdded++
          else if (contextType === 'column_synonyms') storageStats.value.synonymsAdded++
          else storageStats.value.columnsUpdated++
        } else if (target === 'rc_terms') {
          storageStats.value.termsAdded++
        }
        addLog('storage', 'storage', data.message || 'Saved')
        break
      }

      case 'complete':
        isComplete.value = true
        isRunning.value = false
        if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null }
        addLog('success', 'system',
          `Complete! Iterations: ${data.data?.react_iterations || 0}, Embeddings: ${data.data?.embeddings_generated || 0}, Duration: ${Math.round((data.data?.duration_ms || 0) / 1000)}s`)
        persist()
        break

      case 'error':
        addLog('error', 'system', data.message || 'Error')
        break
    }

    // Persist periodically (every event is fine since it's debounced by SSE batching)
    persist()
  }

  // Start SSE generation
  async function startGeneration(dbId: string) {
    // Reset
    databaseId.value = dbId
    isRunning.value = true
    isComplete.value = false
    isMinimized.value = false
    logs.value = []
    logId = 0
    agentState.value = { id: 'rc_gen', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
    embeddingState.value = { id: 'embedding', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
    storageStats.value = {
      tablesTotal: 0, tablesUpdated: 0, columnsTotal: 0, columnsUpdated: 0,
      termsAdded: 0, sampleValuesAdded: 0, synonymsAdded: 0, embeddingsStreamed: 0, embeddingsTotal: 0
    }

    startTime.value = Date.now()
    elapsedTimer = window.setInterval(() => {
      elapsedTime.value = Date.now() - startTime.value
    }, 100)

    addLog('info', 'system', `Starting generation, iterations: ${config.value.minIterations}-${config.value.maxIterations}...`)
    persist()

    const url = `/api/v1/lakebase/datasources/${dbId}/generate-context`
    abortController = new AbortController()

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          concurrency: config.value.concurrency,
          force: config.value.force,
          min_iterations: config.value.minIterations,
          max_iterations: config.value.maxIterations
        }),
        signal: abortController.signal
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
            } catch (parseErr) {
              console.warn('[GenerateContext] SSE parse error:', parseErr)
            }
            currentEventType = 'message'
          }
        }
      }
    } catch (e: any) {
      if (e.name !== 'AbortError') {
        addLog('error', 'system', `Error: ${e.message}`)
      }
    } finally {
      isRunning.value = false
      abortController = null
      if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null }
      persist()
    }
  }

  // Minimize to background
  function minimize() {
    isMinimized.value = true
    showConsole.value = false
    persist()
  }

  // Restore from minimized state
  function restore() {
    showConsole.value = true
    isMinimized.value = false
    // Restart elapsed timer if still running
    if (isRunning.value && !elapsedTimer) {
      elapsedTimer = window.setInterval(() => {
        elapsedTime.value = Date.now() - startTime.value
      }, 100)
    }
    persist()
  }

  // Open console (fresh or resume)
  function openConsole(dbId: string) {
    databaseId.value = dbId
    showConsole.value = true
    isMinimized.value = false
  }

  // Cancel generation
  function cancelGeneration() {
    if (abortController) {
      abortController.abort()
      abortController = null
    }
    isRunning.value = false
    isMinimized.value = false
    if (elapsedTimer) { clearInterval(elapsedTimer); elapsedTimer = null }
    persist()
  }

  // Close console
  function closeConsole() {
    showConsole.value = false
    if (!isRunning.value) {
      isMinimized.value = false
    }
  }

  // Full reset (after user acknowledges completion)
  function reset() {
    isRunning.value = false
    isComplete.value = false
    isMinimized.value = false
    showConsole.value = false
    databaseId.value = ''
    logs.value = []
    agentState.value = { id: 'rc_gen', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
    embeddingState.value = { id: 'embedding', status: 'pending', phase: '', progress: 0, iteration: 0, message: '' }
    storageStats.value = {
      tablesTotal: 0, tablesUpdated: 0, columnsTotal: 0, columnsUpdated: 0,
      termsAdded: 0, sampleValuesAdded: 0, synonymsAdded: 0, embeddingsStreamed: 0, embeddingsTotal: 0
    }
    clearPersistedState()
  }

  // On store init: if persisted state says running but we lost SSE connection (page refresh),
  // mark as interrupted
  if (persisted?.isRunning && !isComplete.value) {
    // SSE connection was lost on refresh - mark the state
    isRunning.value = false
    isMinimized.value = true  // Keep minimized indicator to show last known state
    addLog('error', 'system', 'Connection lost due to page refresh. Progress shown is last known state.')
    persist()
  }

  // Resume elapsed timer if still running (shouldn't happen after refresh, but safety)
  if (isRunning.value && startTime.value > 0 && !elapsedTimer) {
    elapsedTimer = window.setInterval(() => {
      elapsedTime.value = Date.now() - startTime.value
    }, 100)
  }

  return {
    // State
    isRunning,
    isComplete,
    isMinimized,
    showConsole,
    databaseId,
    startTime,
    elapsedTime,
    config,
    agentState,
    embeddingState,
    storageStats,
    logs,

    // Computed
    overallProgress,
    formattedElapsed,
    totalContextWrites,

    // Actions
    startGeneration,
    minimize,
    restore,
    openConsole,
    closeConsole,
    cancelGeneration,
    reset,
    addLog,
    handleEvent
  }
})
