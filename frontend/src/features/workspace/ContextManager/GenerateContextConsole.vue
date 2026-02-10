<script setup lang="ts">
import { computed, onUnmounted } from 'vue'
import { NModal, NButton, NInputNumber, NProgress, NSwitch, useMessage } from 'naive-ui'
import { useContextGenerationStore } from '@/stores/contextGeneration'

const props = defineProps<{
  show: boolean
  databaseId: string
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
  (e: 'complete'): void
  (e: 'minimize'): void
}>()

const store = useContextGenerationStore()
const message = useMessage()

// Expose state for parent component
defineExpose({
  isRunning: computed(() => store.isRunning),
  isComplete: computed(() => store.isComplete),
  progress: computed(() => store.overallProgress)
})

// Computed
const showModal = computed({
  get: () => props.show,
  set: (v) => emit('update:show', v)
})

// Helper functions (purely presentational)
function getPhaseIcon(phase: string): string {
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

function getPhaseColor(phase: string): string {
  switch (phase) {
    case 'thought': return 'text-gray-200'
    case 'action': return 'text-blue-300'
    case 'observation': return 'text-cyan-200'
    case 'storage': return 'text-emerald-300'
    case 'success': return 'text-green-300'
    case 'error': return 'text-red-300'
    case 'finish': return 'text-yellow-200'
    default: return 'text-gray-300'
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

// Actions
async function startGeneration() {
  store.startGeneration(props.databaseId)
}

function handleMinimize() {
  store.minimize()
  showModal.value = false
  emit('minimize')
  message.info('Task running in background.')
}

function handleCancel() {
  store.cancelGeneration()
  showModal.value = false
}

function handleClose() {
  store.closeConsole()
  showModal.value = false
  if (store.isComplete) {
    emit('complete')
  }
}
</script>

<template>
  <NModal
    v-model:show="showModal"
    preset="card"
    :closable="true"
    :mask-closable="!store.isRunning"
    :on-close="store.isRunning ? handleMinimize : undefined"
    style="width: 850px; max-width: 90vw;"
    class="generate-console-modal"
  >
    <template #header>
      <div class="flex items-center gap-2">
        <span>Rich Context Generation</span>
        <span v-if="store.isRunning" class="px-2 py-0.5 text-xs rounded-full bg-blue-500/20 text-blue-400 animate-pulse">
          Running
        </span>
        <span v-else-if="store.isComplete" class="px-2 py-0.5 text-xs rounded-full bg-green-500/20 text-green-400">
          Complete
        </span>
      </div>
    </template>

    <div class="generate-console">
      <!-- Configuration (shown before start) -->
      <div v-if="!store.isRunning && !store.isComplete" class="config-section mb-6">
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Min Iterations</label>
            <NInputNumber v-model:value="store.config.minIterations" :min="1" :max="store.config.maxIterations - 1" size="small" />
          </div>
          <div>
            <label class="text-sm text-gray-400 mb-2 block">Max Iterations</label>
            <NInputNumber v-model:value="store.config.maxIterations" :min="store.config.minIterations + 1" :max="50" size="small" />
          </div>
        </div>
        <div class="flex items-center gap-4 mb-4">
          <div class="flex items-center gap-2">
            <NSwitch v-model:value="store.config.force" size="small" />
            <span class="text-sm text-gray-300">Force Regenerate</span>
          </div>
        </div>
        <p class="text-xs text-gray-500 mb-4">
          More iterations = deeper analysis. The agent explores tables, discovers data patterns, and saves descriptions, sample values, synonyms, and business terms.
        </p>
        <NButton type="primary" size="large" class="w-full" @click="startGeneration">
          <template #icon><div class="i-lucide-play" /></template>
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
                <span :class="['text-xs', getStatusColor(store.agentState.status)]">
                  {{ getStatusIcon(store.agentState.status) }} {{ store.agentState.status }}
                </span>
              </div>
              <NProgress
                :percentage="store.agentState.progress"
                :color="store.agentState.status === 'success' ? '#22c55e' : store.agentState.status === 'error' ? '#ef4444' : '#3b82f6'"
                :height="6"
              />
              <div class="text-xs text-gray-500 mt-1">
                <span v-if="store.agentState.iteration > 0">Iteration {{ store.agentState.iteration }}</span>
                <span v-if="store.agentState.phase" class="ml-2">· {{ store.agentState.phase }}</span>
              </div>
            </div>

            <!-- Embedding Agent -->
            <div class="agent-card p-3 rounded-lg bg-purple-500/10 border border-purple-500/30">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-lg">🧬</span>
                  <span class="font-medium text-purple-300 text-sm">Embeddings</span>
                </div>
                <span :class="['text-xs', getStatusColor(store.embeddingState.status)]">
                  {{ getStatusIcon(store.embeddingState.status) }} {{ store.embeddingState.status }}
                </span>
              </div>
              <NProgress
                :percentage="store.embeddingState.progress"
                :color="store.embeddingState.status === 'success' ? '#22c55e' : store.embeddingState.status === 'error' ? '#ef4444' : '#a855f7'"
                :height="6"
              />
              <div class="text-xs text-gray-500 mt-1 truncate">
                {{ store.embeddingState.message }}
                <span v-if="store.storageStats.embeddingsTotal > 0" class="text-gray-600 ml-1">
                  ({{ store.storageStats.embeddingsStreamed }}/{{ store.storageStats.embeddingsTotal }} entities)
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Storage Stats -->
        <div class="storage-section mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30">
          <h4 class="text-sm text-emerald-300 mb-2 flex items-center gap-2">
            <span>💾</span> Storage Activity
            <span class="text-xs text-gray-500 ml-auto">{{ store.totalContextWrites }} writes</span>
          </h4>
          <div class="grid grid-cols-2 gap-4 mb-2">
            <div>
              <div class="text-xs text-gray-400 mb-1">Table Descriptions</div>
              <NProgress
                :percentage="store.storageStats.tablesTotal ? (store.storageStats.tablesUpdated / store.storageStats.tablesTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ store.storageStats.tablesUpdated }}/{{ store.storageStats.tablesTotal }}</span>
              </NProgress>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">Column Descriptions</div>
              <NProgress
                :percentage="store.storageStats.columnsTotal ? (store.storageStats.columnsUpdated / store.storageStats.columnsTotal) * 100 : 0"
                color="#10b981"
                :height="8"
              >
                <span class="text-xs">{{ store.storageStats.columnsUpdated }}/{{ store.storageStats.columnsTotal }}</span>
              </NProgress>
            </div>
          </div>
          <div class="flex gap-4 text-xs text-gray-500">
            <span v-if="store.storageStats.sampleValuesAdded > 0">📋 Sample Values: {{ store.storageStats.sampleValuesAdded }}</span>
            <span v-if="store.storageStats.synonymsAdded > 0">🔗 Synonyms: {{ store.storageStats.synonymsAdded }}</span>
            <span v-if="store.storageStats.termsAdded > 0">📖 Terms: {{ store.storageStats.termsAdded }}</span>
            <span v-if="store.storageStats.embeddingsStreamed > 0">🧬 Embeddings: {{ store.storageStats.embeddingsStreamed }}</span>
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
          <div class="console-log-area h-80 overflow-y-auto bg-gray-950 rounded-lg p-3 font-mono text-xs leading-relaxed">
            <div v-for="log in store.logs" :key="log.id" class="log-entry py-0.5">
              <span class="text-gray-500">[{{ log.timestamp }}]</span>
              <span class="ml-1">{{ getPhaseIcon(log.phase) }}</span>
              <span :class="['ml-1', getPhaseColor(log.phase)]">{{ log.message }}</span>
              <div v-if="log.detail" class="ml-8 text-gray-400 break-all">{{ log.detail }}</div>
            </div>
            <div v-if="store.isRunning" class="cursor-blink inline-block w-2 h-4 bg-cyan-400 ml-1" />
          </div>
        </div>

        <!-- Footer Stats -->
        <div class="footer-stats mt-4 flex items-center justify-between text-sm text-gray-400">
          <span>⏱️ {{ store.formattedElapsed }}</span>
          <span>📊 Tables: {{ store.storageStats.tablesUpdated }}/{{ store.storageStats.tablesTotal }}</span>
          <span>📝 Columns: {{ store.storageStats.columnsUpdated }}/{{ store.storageStats.columnsTotal }}</span>
          <span>💾 Writes: {{ store.totalContextWrites }}</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-between">
        <div>
          <NButton v-if="store.isRunning" quaternary size="small" @click="handleMinimize">
            <template #icon><span class="i-lucide-minimize-2" /></template>
            Minimize to Background
          </NButton>
        </div>
        <div class="flex gap-2">
          <NButton v-if="store.isRunning" type="error" @click="handleCancel">Cancel</NButton>
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
