import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// Agent state interface
export interface AgentState {
  id: string
  table?: string
  status: 'pending' | 'running' | 'success' | 'error'
  phase: string
  progress: number
  message: string
}

// Log entry interface
export interface LogEntry {
  id: number
  timestamp: string
  agent: string
  message: string
  type: 'info' | 'success' | 'error' | 'storage'
  data?: any
}

// Storage stats interface
export interface StorageStats {
  tablesTotal: number
  tablesUpdated: number
  columnsTotal: number
  columnsUpdated: number
  embeddingsGenerated: number
}

// Task state interface
export interface GenerationTask {
  id: string
  datasourceId: string
  datasourceName: string
  status: 'running' | 'completed' | 'error' | 'cancelled'
  startTime: number
  endTime?: number
  config: {
    concurrency: number
    force: boolean
    minIterations: number
    maxIterations: number
  }
  coordinatorState: AgentState
  workerStates: Map<string, AgentState>
  storageStats: StorageStats
  logs: LogEntry[]
  abortController?: AbortController
}

export const useContextGenerationStore = defineStore('contextGeneration', () => {
  // Active tasks (can have multiple running in background)
  const tasks = ref<Map<string, GenerationTask>>(new Map())
  
  // Currently focused task (for UI display)
  const focusedTaskId = ref<string | null>(null)
  
  // Log ID counter
  let logIdCounter = 0

  // Computed
  const activeTasks = computed(() => 
    Array.from(tasks.value.values()).filter(t => t.status === 'running')
  )
  
  const hasRunningTasks = computed(() => activeTasks.value.length > 0)
  
  const focusedTask = computed(() => 
    focusedTaskId.value ? tasks.value.get(focusedTaskId.value) : null
  )

  // Create a new generation task
  function createTask(datasourceId: string, datasourceName: string, config: GenerationTask['config']): string {
    const taskId = `gen-${Date.now()}`
    const task: GenerationTask = {
      id: taskId,
      datasourceId,
      datasourceName,
      status: 'running',
      startTime: Date.now(),
      config,
      coordinatorState: {
        id: 'coordinator',
        status: 'pending',
        phase: '',
        progress: 0,
        message: ''
      },
      workerStates: new Map(),
      storageStats: {
        tablesTotal: 0,
        tablesUpdated: 0,
        columnsTotal: 0,
        columnsUpdated: 0,
        embeddingsGenerated: 0
      },
      logs: [],
      abortController: new AbortController()
    }
    tasks.value.set(taskId, task)
    focusedTaskId.value = taskId
    return taskId
  }

  // Add log to task
  function addLog(taskId: string, agent: string, message: string, type: LogEntry['type'], data?: any) {
    const task = tasks.value.get(taskId)
    if (!task) return

    const timestamp = new Date().toLocaleTimeString('en-US', { hour12: false })
    task.logs.push({
      id: logIdCounter++,
      timestamp,
      agent,
      message,
      type,
      data
    })

    // Keep last 500 logs
    if (task.logs.length > 500) {
      task.logs = task.logs.slice(-500)
    }
  }

  // Update coordinator state
  function updateCoordinator(taskId: string, update: Partial<AgentState>) {
    const task = tasks.value.get(taskId)
    if (!task) return
    Object.assign(task.coordinatorState, update)
  }

  // Update worker state
  function updateWorker(taskId: string, workerId: string, update: Partial<AgentState>) {
    const task = tasks.value.get(taskId)
    if (!task) return
    
    if (!task.workerStates.has(workerId)) {
      task.workerStates.set(workerId, {
        id: workerId,
        status: 'pending',
        phase: '',
        progress: 0,
        message: ''
      })
    }
    const worker = task.workerStates.get(workerId)!
    Object.assign(worker, update)
  }

  // Update storage stats
  function updateStorageStats(taskId: string, update: Partial<StorageStats>) {
    const task = tasks.value.get(taskId)
    if (!task) return
    Object.assign(task.storageStats, update)
  }

  // Complete task
  function completeTask(taskId: string, status: 'completed' | 'error' | 'cancelled' = 'completed') {
    const task = tasks.value.get(taskId)
    if (!task) return
    task.status = status
    task.endTime = Date.now()
  }

  // Cancel task
  function cancelTask(taskId: string) {
    const task = tasks.value.get(taskId)
    if (!task) return
    task.abortController?.abort()
    completeTask(taskId, 'cancelled')
  }

  // Focus on a task
  function focusTask(taskId: string | null) {
    focusedTaskId.value = taskId
  }

  // Remove completed task
  function removeTask(taskId: string) {
    const task = tasks.value.get(taskId)
    if (task?.status === 'running') {
      cancelTask(taskId)
    }
    tasks.value.delete(taskId)
    if (focusedTaskId.value === taskId) {
      focusedTaskId.value = null
    }
  }

  // Clear all completed tasks
  function clearCompletedTasks() {
    for (const [id, task] of tasks.value) {
      if (task.status !== 'running') {
        tasks.value.delete(id)
      }
    }
  }

  return {
    tasks,
    focusedTaskId,
    activeTasks,
    hasRunningTasks,
    focusedTask,
    createTask,
    addLog,
    updateCoordinator,
    updateWorker,
    updateStorageStats,
    completeTask,
    cancelTask,
    focusTask,
    removeTask,
    clearCompletedTasks
  }
})
