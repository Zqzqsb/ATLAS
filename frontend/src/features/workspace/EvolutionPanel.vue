<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { NButton, NTag, NProgress, NSwitch, NScrollbar, NEmpty, NTooltip, useMessage } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { evolutionApi } from '@/api/evolution'
import { agentApi } from '@/api/agent'
import type { EvolutionStatus, EvolutionStage, StageExecution, ContextAction } from '@/api/evolution'
import type { ChangeLog } from '@/api/agent'

const message = useMessage()
const workspaceStore = useWorkspaceStore()

// ========================================
// State
// ========================================
const status = ref<EvolutionStatus | null>(null)
const loading = ref(false)
const executing = ref(false)
const resetting = ref(false)

// Event log for real-time streaming display
interface EventLogEntry {
  id: number
  type: string
  phase: string
  message: string
  timestamp: Date
  data?: any
}
const eventLog = ref<EventLogEntry[]>([])
let eventCounter = 0

// Change logs from agent
const changeLogs = ref<ChangeLog[]>([])

// Auto-scroll ref
const logScrollRef = ref<any>(null)

// Datasource ID from workspace store (dynamic, not hardcoded)
const datasourceId = computed(() => {
  return workspaceStore.currentDatabase?.metadata?.lakebaseId ?? 1
})

// ========================================
// Computed
// ========================================
const currentStage = computed(() => status.value?.current_stage ?? 0)
const totalStages = computed(() => status.value?.total_stages ?? 5)
const stages = computed(() => status.value?.stages ?? [])
const history = computed(() => status.value?.history ?? [])
const nextStageId = computed(() => currentStage.value + 1)
const canExecuteNext = computed(() => nextStageId.value <= totalStages.value && !executing.value)
const isComplete = computed(() => currentStage.value >= totalStages.value)
const progressPercent = computed(() => Math.round((currentStage.value / totalStages.value) * 100))

// Stage icon mapping
function getStageIcon(stage: EvolutionStage): string {
  switch (stage.id) {
    case 1: return 'i-lucide-smartphone'
    case 2: return 'i-lucide-shopping-cart'
    case 3: return 'i-lucide-git-merge'
    case 4: return 'i-lucide-trending-up'
    case 5: return 'i-lucide-trash-2'
    default: return 'i-lucide-table-2'
  }
}

function getChangeTypeLabel(type: string): string {
  const map: Record<string, string> = {
    'column_added': 'Column Added',
    'column_dropped': 'Column Dropped',
    'column_modified': 'Column Modified',
    'table_added': 'Table Added',
    'table_dropped': 'Table Dropped',
    'fk_added': 'FK Added',
    'fk_dropped': 'FK Dropped',
  }
  return map[type] || type
}

function getChangeTypeColor(type: string): string {
  if (type.includes('added')) return 'text-green-600'
  if (type.includes('dropped') || type.includes('deleted')) return 'text-red-600'
  if (type.includes('modified') || type.includes('refreshed')) return 'text-amber-600'
  return 'text-blue-600'
}

function getActionIcon(type: string): string {
  switch (type) {
    case 'created': return 'i-lucide-plus-filled'
    case 'expired': return 'i-lucide-alert-triangle-alt-filled'
    case 'refreshed': return 'i-lucide-refresh-cw'
    case 'deleted': return 'i-lucide-x-filled'
    default: return 'i-lucide-info'
  }
}

function getActionColor(type: string): string {
  switch (type) {
    case 'created': return 'text-green-500'
    case 'expired': return 'text-amber-500'
    case 'refreshed': return 'text-blue-500'
    case 'deleted': return 'text-red-500'
    default: return 'text-gray-500'
  }
}

function getEventIcon(type: string): string {
  switch (type) {
    case 'stage_start': return 'i-lucide-play-filled-alt'
    case 'ddl_executing': return 'i-lucide-terminal'
    case 'ddl_complete': return 'i-lucide-check'
    case 'data_inserting': return 'i-lucide-table-2'
    case 'data_complete': return 'i-lucide-check'
    case 'detecting': return 'i-lucide-search'
    case 'changes_detected': return 'i-lucide-alert-triangle-alt'
    case 'syncing_schema': return 'i-lucide-database'
    case 'schema_synced': return 'i-lucide-check'
    case 'agent_start': return 'i-lucide-bot'
    case 'agent_step': return 'i-lucide-cpu'
    case 'agent_complete': return 'i-lucide-check-filled'
    case 'agent_error': return 'i-lucide-x-circle'
    case 'marking_expired': return 'i-lucide-clock'
    case 'context_expired': return 'i-lucide-alert-triangle-alt-filled'
    case 'creating_context': return 'i-lucide-cpu'
    case 'context_created': return 'i-lucide-plus-filled'
    case 'context_refreshed': return 'i-lucide-refresh-cw'
    case 'context_deleted': return 'i-lucide-x-filled'
    case 'refreshing_context': return 'i-lucide-cpu'
    case 'context_refreshed_complete': return 'i-lucide-check-filled'
    case 'updating_embeddings': return 'i-lucide-circle-dot'
    case 'embedding_update': return 'i-lucide-circle-dot'
    case 'embedding_complete': return 'i-lucide-check-filled'
    case 'stage_complete': return 'i-lucide-check-filled'
    case 'error': return 'i-lucide-x-circle'
    case 'execution_complete': return 'i-lucide-trophy'
    case 'reset_step': return 'i-lucide-rotate-ccw'
    case 'reset_start': return 'i-lucide-rotate-ccw'
    case 'reset_complete': return 'i-lucide-check-filled'
    default: return 'i-lucide-info'
  }
}

function getEventColor(type: string): string {
  if (type.includes('error')) return 'text-red-500'
  if (type.includes('complete') || type.includes('created') || type === 'ddl_complete' || type === 'data_complete' || type === 'schema_synced') return 'text-green-500'
  if (type.includes('expired') || type.includes('warning') || type.includes('detecting')) return 'text-amber-500'
  if (type.includes('executing') || type.includes('refreshing') || type.includes('creating') || type.includes('updating') || type.includes('syncing')) return 'text-blue-500'
  if (type === 'stage_start') return 'text-indigo-500'
  if (type === 'agent_start') return 'text-violet-500'
  if (type === 'agent_step') return 'text-cyan-500'
  return 'text-gray-500'
}

// Phase labels for separator bars
function getPhaseLabel(phase: string): string {
  const map: Record<string, string> = {
    'announce': '🚀 Stage Start',
    'reset': '🔄 Auto Reset',
    'ddl': '⚡ DDL Execution',
    'data': '📦 Sample Data',
    'detect': '🔍 Change Detection',
    'sync': '🔄 Schema Sync',
    'maintain': '🤖 Agent Maintenance',
    'embed': '📐 Embedding Update',
    'done': '✅ Complete',
  }
  return map[phase] || phase
}

// Step type specific styling for agent steps
function getStepTypeIcon(stepType: string): string {
  switch (stepType) {
    case 'thought': return 'i-lucide-brain'
    case 'action': return 'i-lucide-wrench'
    case 'observation': return 'i-lucide-eye'
    case 'finish': return 'i-lucide-check-circle-2'
    default: return 'i-lucide-info'
  }
}

function getStepTypeColor(stepType: string): string {
  switch (stepType) {
    case 'thought': return 'text-purple-400'
    case 'action': return 'text-cyan-400'
    case 'observation': return 'text-gray-500'
    case 'finish': return 'text-emerald-400'
    default: return 'text-gray-400'
  }
}

function getAgentRoleLabel(role: string): string {
  switch (role) {
    case 'coordinator': return 'COORD'
    case 'executor': return 'EXEC'
    default: return role.toUpperCase()
  }
}

function getAgentRoleBgClass(role: string): string {
  switch (role) {
    case 'coordinator': return 'bg-violet-500/20 text-violet-300 border-violet-500/30'
    case 'executor': return 'bg-teal-500/20 text-teal-300 border-teal-500/30'
    default: return 'bg-gray-500/20 text-gray-300 border-gray-500/30'
  }
}

// Track expanded thoughts
const expandedThoughts = ref<Set<number>>(new Set())
function toggleThought(id: number) {
  if (expandedThoughts.value.has(id)) {
    expandedThoughts.value.delete(id)
  } else {
    expandedThoughts.value.add(id)
  }
}

// Check if an event is a phase boundary
function isNewPhase(event: EventLogEntry, index: number): boolean {
  if (index === 0) return true
  const prev = eventLog.value[index - 1]
  return prev ? prev.phase !== event.phase : true
}

// ========================================
// Actions
// ========================================

async function fetchStatus() {
  loading.value = true
  try {
    status.value = await evolutionApi.getStatus()
  } catch (e: any) {
    console.error('Failed to fetch evolution status:', e)
  } finally {
    loading.value = false
  }
}

async function fetchChangeLogs() {
  try {
    const result = await agentApi.getChangeLogs(datasourceId.value, 30)
    changeLogs.value = result.logs || []
  } catch (e) {
    // ignore
  }
}

function addEvent(type: string, phase: string, msg: string, data?: any) {
  eventLog.value.push({
    id: eventCounter++,
    type,
    phase,
    message: msg,
    timestamp: new Date(),
    data
  })
  // Auto-scroll to bottom
  nextTick(() => {
    if (logScrollRef.value) {
      const el = logScrollRef.value.$el || logScrollRef.value
      if (el && el.scrollTo) {
        el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
      }
    }
  })
}

async function executeNextStage() {
  if (!canExecuteNext.value) return

  executing.value = true
  const stageId = nextStageId.value

  addEvent('stage_start', 'announce', `Starting Stage ${stageId}...`)

  const abort = evolutionApi.executeStageStream(
    datasourceId.value,
    stageId,
    (event) => {
      // Add event to log
      const data = event.data as any
      if (data && data.message) {
        addEvent(event.type, data.phase || '', data.message, data.data)
      }

      // Handle completion
      if (event.type === 'execution_complete') {
        executing.value = false
        fetchStatus()
        fetchChangeLogs()
        message.success(`Stage ${stageId} completed!`)
      }
      if (event.type === 'error') {
        executing.value = false
        message.error(data?.error || data?.message || 'Stage execution failed')
      }
    },
    (err) => {
      executing.value = false
      addEvent('error', 'system', `Connection error: ${err.message}`)
      message.error('Connection failed: ' + err.message)
    },
    () => {
      executing.value = false
    }
  )
}

async function resetToInitial() {
  resetting.value = true
  eventLog.value = []

  addEvent('reset_step', 'reset', 'Resetting to initial state...')

  evolutionApi.resetStream(
    datasourceId.value,
    (event) => {
      const data = event.data as any
      if (data && data.message) {
        addEvent(event.type, data.phase || 'reset', data.message, data.data)
      }

      if (event.type === 'reset_complete') {
        resetting.value = false
        message.success('Reset to initial state')
        fetchStatus()
        fetchChangeLogs()
      }
      if (event.type === 'error') {
        resetting.value = false
        message.error(data?.error || data?.message || 'Reset failed')
      }
    },
    (err) => {
      resetting.value = false
      addEvent('error', 'reset', `Connection error: ${err.message}`)
      message.error('Reset failed: ' + err.message)
    },
    () => {
      resetting.value = false
    }
  )
}

// ========================================
// Lifecycle
// ========================================
onMounted(async () => {
  await Promise.all([fetchStatus(), fetchChangeLogs()])
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="bg-indigo-50 rounded-lg p-5 border border-indigo-100">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="w-11 h-11 rounded-lg bg-indigo-600 flex items-center justify-center">
            <div class="i-lucide-bot text-3xl text-white" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-gray-900">Schema Evolution</h2>
            <p class="text-sm text-gray-500 mt-0.5">
              Watch the Agent detect schema changes and maintain Rich Context automatically
            </p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <NTag :type="isComplete ? 'success' : 'info'" size="large" round>
            Stage {{ currentStage }} / {{ totalStages }}
          </NTag>
        </div>
      </div>

      <!-- Progress bar -->
      <div class="mt-5">
        <div class="flex items-center justify-between text-xs font-medium text-gray-500 mb-2">
          <span>Evolution Progress</span>
          <span>{{ progressPercent }}%</span>
        </div>
        <div class="h-2.5 bg-white/80 rounded-full overflow-hidden border border-gray-200/50">
          <div 
            class="h-full rounded-full transition-all duration-700 ease-out"
            :class="isComplete ? 'bg-emerald-500' : 'bg-indigo-500'"
            :style="{ width: `${progressPercent}%` }"
          />
        </div>
      </div>
    </div>

    <!-- Stage Timeline -->
    <div class="card p-6">
      <h3 class="font-bold text-gray-900 mb-5 flex items-center gap-2">
        <span class="i-lucide-flag text-indigo-500" />
        Evolution Stages
      </h3>

      <div class="relative">
        <!-- Horizontal timeline line -->
        <div class="absolute top-8 left-0 right-0 h-0.5 bg-gray-200" />
        
        <div class="grid grid-cols-5 gap-2">
          <div 
            v-for="stage in stages" 
            :key="stage.id"
            class="relative flex flex-col items-center"
          >
            <!-- Stage dot -->
            <div 
              class="relative z-10 w-16 h-16 rounded-2xl flex items-center justify-center transition-all duration-300 border-2"
              :class="{
                'bg-green-50 border-green-400 shadow-sm shadow-green-100': stage.executed,
                'bg-indigo-50 border-indigo-400 shadow-md shadow-indigo-100 animate-pulse': stage.is_next && !executing,
                'bg-indigo-100 border-indigo-500 shadow-lg shadow-indigo-200': stage.is_next && executing,
                'bg-gray-50 border-gray-200': !stage.executed && !stage.is_next,
              }"
            >
              <div 
                v-if="stage.executed"
                class="i-lucide-check-filled text-2xl text-green-500"
              />
              <div 
                v-else-if="stage.is_next && executing"
                class="i-lucide-loader-2 text-2xl text-indigo-500 animate-spin"
              />
              <div 
                v-else
                :class="getStageIcon(stage)"
                class="text-2xl"
                :style="{ color: stage.is_next ? '#6366f1' : '#9ca3af' }"
              />
            </div>

            <!-- Stage label -->
            <div class="mt-3 text-center">
              <div 
                class="text-xs font-bold uppercase tracking-wider"
                :class="{
                  'text-green-600': stage.executed,
                  'text-indigo-600': stage.is_next,
                  'text-gray-400': !stage.executed && !stage.is_next,
                }"
              >
                Stage {{ stage.id }}
              </div>
              <div 
                class="text-xs mt-1 leading-tight max-w-[120px]"
                :class="{
                  'text-gray-700 font-medium': stage.executed || stage.is_next,
                  'text-gray-400': !stage.executed && !stage.is_next,
                }"
              >
                {{ stage.name }}
              </div>
            </div>

            <!-- Expected changes badges -->
            <div class="mt-2 flex flex-wrap justify-center gap-1">
              <span 
                v-for="change in stage.expected_changes" 
                :key="change"
                class="text-[10px] px-1.5 py-0.5 rounded-full"
                :class="{
                  'bg-green-100 text-green-700': stage.executed,
                  'bg-indigo-100 text-indigo-600': stage.is_next,
                  'bg-gray-100 text-gray-500': !stage.executed && !stage.is_next,
                }"
              >
                {{ getChangeTypeLabel(change) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Control Panel + Event Log -->
    <div class="grid grid-cols-12 gap-6">
      <!-- Left: Controls + Stage Details -->
      <div class="col-span-5 space-y-5">
        <!-- Action Buttons -->
        <div class="card p-5">
          <h3 class="font-bold text-gray-900 mb-4 flex items-center gap-2">
            <span class="i-lucide-play-filled-alt text-indigo-500" />
            Controls
          </h3>

          <div class="space-y-3">
            <NButton 
              type="primary" 
              size="large"
              block
              :loading="executing"
              :disabled="!canExecuteNext"
              @click="executeNextStage"
            >
              <template #icon>
                <div class="i-lucide-play" />
              </template>
              {{ executing ? `Executing Stage ${nextStageId}...` : isComplete ? 'All Stages Complete' : `Execute Stage ${nextStageId}` }}
            </NButton>

            <NButton
              size="large"
              block
              :loading="resetting"
              :disabled="executing || currentStage === 0"
              @click="resetToInitial"
              quaternary
              type="warning"
            >
              <template #icon>
                <div class="i-lucide-rotate-ccw" />
              </template>
              Reset to Initial State
            </NButton>
          </div>

          <!-- Next stage preview -->
          <div v-if="!isComplete && stages[currentStage]" class="mt-5 p-4 bg-indigo-50/50 rounded-xl border border-indigo-100">
            <div class="text-xs font-bold text-indigo-600 uppercase tracking-wider mb-2">
              Next: Stage {{ nextStageId }}
            </div>
            <div class="text-sm font-medium text-gray-800 mb-2">
              {{ stages[currentStage]?.name }}
            </div>
            <div class="text-xs text-gray-500 mb-3">
              {{ stages[currentStage]?.description }}
            </div>
            <div class="space-y-1.5">
              <div 
                v-for="(ddl, i) in stages[currentStage]?.ddls || []" 
                :key="i"
                class="font-mono text-xs bg-gray-900 text-green-400 px-3 py-2 rounded-lg overflow-x-auto"
              >
                {{ ddl }}
              </div>
            </div>
          </div>

          <!-- Complete state -->
          <div v-else-if="isComplete" class="mt-5 p-4 bg-green-50 rounded-xl border border-green-200">
            <div class="flex items-center gap-2 text-green-700 font-bold mb-1">
              <span class="i-lucide-trophy text-lg" />
              All Stages Complete!
            </div>
            <p class="text-xs text-green-600">
              The database has evolved through {{ totalStages }} stages. Click Reset to start over.
            </p>
          </div>
        </div>

        <!-- Execution History (compact) -->
        <div v-if="history.length > 0" class="card p-5">
          <h3 class="font-bold text-gray-900 mb-4 flex items-center gap-2">
            <span class="i-lucide-history text-blue-500" />
            Execution History
          </h3>

          <div class="space-y-3">
            <div 
              v-for="exec in [...history].reverse()" 
              :key="exec.stage_id"
              class="p-3 rounded-xl border"
              :class="exec.success ? 'bg-green-50/50 border-green-200' : 'bg-red-50/50 border-red-200'"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-bold text-gray-800">
                  Stage {{ exec.stage_id }}: {{ exec.stage_name }}
                </span>
                <span class="text-xs text-gray-500">
                  {{ exec.duration_ms }}ms
                </span>
              </div>
              
              <!-- Context Actions -->
              <div v-if="exec.context_actions?.length" class="space-y-1">
                <div 
                  v-for="(action, i) in exec.context_actions" 
                  :key="i"
                  class="flex items-center gap-2 text-xs"
                >
                  <span :class="[getActionIcon(action.action_type), getActionColor(action.action_type)]" />
                  <span class="text-gray-600">{{ action.description }}</span>
                </div>
              </div>

              <!-- Schema Changes -->
              <div v-if="exec.changes_detected?.length" class="mt-2 flex flex-wrap gap-1">
                <span 
                  v-for="(change, i) in exec.changes_detected" 
                  :key="i"
                  class="text-[10px] font-mono px-2 py-0.5 rounded-full"
                  :class="{
                    'bg-green-100 text-green-700': change.change_type.includes('added'),
                    'bg-red-100 text-red-700': change.change_type.includes('dropped'),
                    'bg-amber-100 text-amber-700': change.change_type.includes('modified'),
                  }"
                >
                  {{ change.change_type }}: {{ change.table_name }}{{ change.column_name ? '.' + change.column_name : '' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Real-time Event Log -->
      <div class="col-span-7">
        <div class="card p-5 h-full flex flex-col">
          <div class="flex items-center justify-between mb-4">
            <h3 class="font-bold text-gray-900 flex items-center gap-2">
              <span class="i-lucide-activity text-green-500" />
              Real-time Event Log
              <span v-if="executing" class="flex items-center gap-1 text-xs font-normal text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                <span class="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse" />
                LIVE
              </span>
            </h3>
            <NButton quaternary size="tiny" @click="eventLog = []" :disabled="executing">
              <template #icon><div class="i-lucide-trash-2" /></template>
              Clear
            </NButton>
          </div>

          <div 
            ref="logScrollRef"
            class="flex-1 overflow-y-auto min-h-[400px] max-h-[600px] bg-gray-950 rounded-xl p-4 font-mono text-xs"
          >
            <div v-if="eventLog.length === 0" class="flex items-center justify-center h-full text-gray-500">
              <div class="text-center">
                <div class="i-lucide-terminal text-4xl mb-3 opacity-50" />
                <div class="text-sm">Execute a stage to see real-time events here</div>
              </div>
            </div>

            <div v-else class="space-y-0.5">
              <template v-for="(event, idx) in eventLog" :key="event.id">
                <!-- Phase separator bar -->
                <div 
                  v-if="isNewPhase(event, idx)"
                  class="flex items-center gap-2 pt-2 pb-1"
                  :class="idx > 0 ? 'mt-2 border-t border-gray-700/40' : ''"
                >
                  <span class="text-[10px] font-bold uppercase tracking-widest px-2 py-0.5 rounded-full"
                    :class="{
                      'bg-indigo-500/15 text-indigo-400': event.phase === 'announce',
                      'bg-amber-500/15 text-amber-400': event.phase === 'reset',
                      'bg-blue-500/15 text-blue-400': event.phase === 'ddl',
                      'bg-cyan-500/15 text-cyan-400': event.phase === 'data',
                      'bg-yellow-500/15 text-yellow-400': event.phase === 'detect',
                      'bg-orange-500/15 text-orange-400': event.phase === 'sync',
                      'bg-violet-500/15 text-violet-400': event.phase === 'maintain',
                      'bg-teal-500/15 text-teal-400': event.phase === 'embed',
                      'bg-emerald-500/15 text-emerald-400': event.phase === 'done',
                    }"
                  >{{ getPhaseLabel(event.phase) }}</span>
                  <div class="flex-1 h-px bg-gray-700/30" />
                </div>

                <!-- Agent step: special rendering -->
                <div 
                  v-if="event.type === 'agent_step' && event.data"
                  class="flex items-start gap-2 py-0.5 pl-2"
                >
                  <!-- Timestamp -->
                  <span class="text-gray-600 shrink-0 w-14 mt-0.5">
                    {{ event.timestamp.toLocaleTimeString('en', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }) }}
                  </span>

                  <!-- Agent role badge -->
                  <span 
                    class="shrink-0 text-[9px] font-bold px-1.5 py-0.5 rounded border mt-px"
                    :class="getAgentRoleBgClass(event.data?.agent_role)"
                  >{{ getAgentRoleLabel(event.data?.agent_role) }}</span>

                  <!-- Step type icon -->
                  <span 
                    :class="[getStepTypeIcon(event.data?.step_type), getStepTypeColor(event.data?.step_type)]"
                    class="shrink-0 mt-0.5"
                  />

                  <!-- Content based on step type -->
                  <div class="flex-1 min-w-0">
                    <!-- Thought: italic, collapsible if long -->
                    <template v-if="event.data?.step_type === 'thought'">
                      <div 
                        class="text-purple-300/80 italic leading-relaxed cursor-pointer"
                        @click="event.message.length > 150 ? toggleThought(event.id) : null"
                      >
                        <span v-if="event.message.length > 150 && !expandedThoughts.has(event.id)">
                          {{ event.message.slice(0, 150) }}...
                          <span class="text-purple-500 text-[10px] ml-1 not-italic">[expand]</span>
                        </span>
                        <span v-else>{{ event.message }}</span>
                      </div>
                    </template>

                    <!-- Tool call: code-style badge -->
                    <template v-else-if="event.data?.step_type === 'action'">
                      <div class="flex items-center gap-1.5 flex-wrap">
                        <span class="bg-cyan-500/15 text-cyan-300 px-1.5 py-0.5 rounded font-bold text-[10px] border border-cyan-500/20">
                          {{ event.data?.tool_name || 'tool' }}
                        </span>
                        <span class="text-gray-400 break-all">{{ event.message.replace(/^\[.*?\]\s*/, '') }}</span>
                      </div>
                    </template>

                    <!-- Observation: dimmed -->
                    <template v-else-if="event.data?.step_type === 'observation'">
                      <div class="text-gray-500 leading-relaxed">
                        <span v-if="event.message.length > 200">
                          {{ event.message.slice(0, 200) }}...
                        </span>
                        <span v-else>{{ event.message }}</span>
                      </div>
                    </template>

                    <!-- Final Answer: green bold -->
                    <template v-else-if="event.data?.step_type === 'finish'">
                      <div class="text-emerald-400 font-semibold leading-relaxed">
                        {{ event.message }}
                      </div>
                    </template>

                    <!-- Fallback -->
                    <template v-else>
                      <span class="text-gray-300 leading-relaxed">{{ event.message }}</span>
                    </template>
                  </div>
                </div>

                <!-- Normal event (non-agent) -->
                <div 
                  v-else
                  class="flex items-start gap-2 py-0.5"
                >
                  <!-- Timestamp -->
                  <span class="text-gray-600 shrink-0 w-14 mt-0.5">
                    {{ event.timestamp.toLocaleTimeString('en', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }) }}
                  </span>

                  <!-- Icon -->
                  <span 
                    :class="[getEventIcon(event.type), getEventColor(event.type)]" 
                    class="shrink-0 mt-0.5"
                  />

                  <!-- Message -->
                  <span 
                    class="leading-relaxed"
                    :class="{
                      'text-green-400 font-bold': event.type === 'stage_complete' || event.type === 'execution_complete' || event.type === 'reset_complete',
                      'text-red-400 font-bold': event.type === 'error',
                      'text-amber-300': event.type.includes('expired') || event.type.includes('detecting') || event.type.includes('changes_detected'),
                      'text-blue-300': event.type.includes('executing') || event.type.includes('refreshing') || event.type.includes('creating') || event.type.includes('updating') || event.type.includes('syncing'),
                      'text-indigo-300 font-semibold': event.type === 'stage_start',
                      'text-violet-300 font-semibold': event.type === 'agent_start' || event.type === 'agent_complete',
                      'text-green-300': event.type.includes('created') || event.type === 'ddl_complete' || event.type === 'data_complete' || event.type === 'schema_synced' || event.type === 'embedding_complete',
                      'text-gray-300': !['stage_complete', 'execution_complete', 'reset_complete', 'error', 'stage_start', 'agent_start', 'agent_complete'].includes(event.type) && !event.type.includes('expired') && !event.type.includes('executing') && !event.type.includes('created') && !event.type.includes('syncing') && !event.type.includes('detecting'),
                    }"
                  >
                    {{ event.message }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Change Logs -->
    <div v-if="changeLogs.length > 0" class="card p-6">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-gray-900 flex items-center gap-2">
          <span class="i-lucide-folder-open text-purple-500" />
          Agent Change Logs
        </h3>
        <NButton quaternary size="tiny" @click="fetchChangeLogs">
          <template #icon><div class="i-lucide-refresh-cw" /></template>
          Refresh
        </NButton>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200">
              <th class="text-left py-2 px-3 text-xs font-bold text-gray-500 uppercase">Time</th>
              <th class="text-left py-2 px-3 text-xs font-bold text-gray-500 uppercase">Type</th>
              <th class="text-left py-2 px-3 text-xs font-bold text-gray-500 uppercase">Table</th>
              <th class="text-left py-2 px-3 text-xs font-bold text-gray-500 uppercase">Source</th>
              <th class="text-left py-2 px-3 text-xs font-bold text-gray-500 uppercase">Reason</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="log in changeLogs.slice(0, 15)" 
              :key="log.id"
              class="border-b border-gray-100 hover:bg-gray-50"
            >
              <td class="py-2 px-3 text-xs text-gray-500 font-mono">
                {{ new Date(log.created_at).toLocaleTimeString() }}
              </td>
              <td class="py-2 px-3">
                <span 
                  class="text-xs px-2 py-0.5 rounded-full font-medium"
                  :class="{
                    'bg-amber-100 text-amber-700': log.change_type === 'schema_change',
                    'bg-green-100 text-green-700': log.change_type === 'context_update',
                    'bg-red-100 text-red-700': log.change_type === 'context_expire',
                  }"
                >
                  {{ log.change_type === 'schema_change' ? 'Schema' : log.change_type === 'context_update' ? 'Context Update' : 'Expired' }}
                </span>
              </td>
              <td class="py-2 px-3 font-mono text-xs text-gray-800">
                {{ log.table_name }}
              </td>
              <td class="py-2 px-3 text-xs text-gray-500">
                {{ log.trigger_source }}
              </td>
              <td class="py-2 px-3 text-xs text-gray-600 max-w-[300px] truncate">
                {{ log.change_reason }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  @apply bg-white rounded-xl border border-gray-200 shadow-sm;
}

/* Terminal scrollbar */
.font-mono::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
.font-mono::-webkit-scrollbar-track {
  background: transparent;
}
.font-mono::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 3px;
}
</style>
