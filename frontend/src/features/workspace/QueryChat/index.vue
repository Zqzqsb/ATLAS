<script setup lang="ts">
import { ref, computed } from 'vue'
import { NButton, NInput, NInputNumber, NSwitch, NSelect, useMessage } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import QueryResult from './QueryResult.vue'
import RealtimeCard from './RealtimeCard.vue'

const workspaceStore = useWorkspaceStore()
const message = useMessage()

// Query input
const question = ref('')
const showExamples = ref(true)

// Use store's isQuerying for execution state
const isExecuting = computed(() => workspaceStore.isQuerying)

// Query options
const maxIterations = ref(5)
const useFieldAlignment = ref(true)
const selectedModel = ref('deepseek_v3')
const useRichContext = ref(true)
const useReact = ref(true)
const useGrounding = ref(true)

// Model options
const modelOptions = [
  { label: 'DeepSeek V3', value: 'deepseek_v3' },
  { label: 'Qwen 2.5', value: 'qwen_2.5' },
  { label: 'GPT-4', value: 'gpt4' }
]

// Example questions for spider_tvshow database
const exampleQuestions = computed(() => {
  const dbName = workspaceStore.currentDatabase?.name || ''
  
  if (dbName === 'spider_tvshow') {
    return [
      'List all TV channels with their countries',
      'Which TV channel has the highest package price?',
      'Show all cartoons and their corresponding TV channels',
      'How many TV series are broadcasted in each country?',
      'Find all channels that offer High Definition TV',
      'List all cartoons directed by Ben Jones'
    ]
  }
  
  // Default examples
  return [
    '查询所有电视频道',
    '查找收视率最高的节目',
    '统计每个国家的频道数量',
    '查询所有动画片及其播出频道'
  ]
})

// Execution stages
const vectorSearchStage = computed(() => ({
  active: isExecuting.value && workspaceStore.groundingStage !== 'idle',
  data: workspaceStore.groundingResult,
  duration: 0
}))

const schemaLinkingStage = computed(() => ({
  active: isExecuting.value && workspaceStore.reactSteps.some(s => s.phase === 'schema_linking'),
  steps: workspaceStore.reactSteps.filter(s => s.phase === 'schema_linking'),
  contexts: workspaceStore.usedContexts
}))

const sqlGenerationStage = computed(() => ({
  active: isExecuting.value || !!workspaceStore.generatedSql,
  steps: workspaceStore.reactSteps.filter(s => s.phase === 'sql_generation'),
  sql: workspaceStore.generatedSql
}))

async function handleExecute() {
  if (!question.value.trim()) {
    message.warning('请输入问题')
    return
  }
  
  // Update query options
  workspaceStore.queryOptions.maxIterations = maxIterations.value
  workspaceStore.queryOptions.useRichContext = useRichContext.value
  workspaceStore.queryOptions.useReact = useReact.value
  workspaceStore.queryOptions.useGrounding = useGrounding.value

  try {
    await workspaceStore.executeQuery(question.value)
  } catch (e: any) {
    message.error(e.message || '执行失败')
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
  workspaceStore.resetQueryState()
}
</script>

<template>
  <div class="query-chat min-h-full bg-gradient-to-br from-gray-900 via-slate-900 to-gray-950 p-6">
    <!-- Control Panel -->
    <div class="control-panel mb-6 p-6 rounded-2xl bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-md border border-white/10">
      <!-- Question Input -->
      <div class="mb-6">
        <div class="flex items-center gap-2 mb-3">
          <div class="i-carbon-chat text-xl text-blue-400" />
          <h3 class="text-lg font-semibold text-white">Natural Language Query</h3>
        </div>
        
        <NInput
          v-model:value="question"
          type="textarea"
          :autosize="{ minRows: 2, maxRows: 4 }"
          placeholder="输入自然语言问题，例如：查询所有电视频道..."
          :disabled="isExecuting"
          class="query-input"
          @keydown.ctrl.enter="handleExecute"
        />

        <!-- Example questions -->
        <div v-if="showExamples" class="mt-3 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-400">示例问题 (来自 Spider Benchmark):</span>
            <button
              class="text-xs text-gray-500 hover:text-gray-300 transition-colors"
              @click="showExamples = false"
            >
              收起 ↑
            </button>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="example in exampleQuestions"
              :key="example"
              class="text-xs px-3 py-1.5 rounded-lg bg-gradient-to-r from-blue-500/10 to-cyan-500/10 text-blue-300 hover:from-blue-500/20 hover:to-cyan-500/20 transition-all border border-blue-500/20 hover:border-blue-400/40 hover:shadow-lg hover:shadow-blue-500/10"
              @click="useExample(example)"
            >
              {{ example }}
            </button>
          </div>
        </div>
        <div v-else class="mt-3">
          <button
            class="text-xs text-gray-500 hover:text-gray-300 transition-colors"
            @click="showExamples = true"
          >
            显示示例问题 ↓
          </button>
        </div>
      </div>

      <!-- Parameters -->
      <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
        <!-- Model Selection -->
        <div class="param-item">
          <label class="text-xs text-gray-400 mb-2 block">Model</label>
          <NSelect
            v-model:value="selectedModel"
            :options="modelOptions"
            :disabled="isExecuting"
            size="small"
          />
        </div>

        <!-- Max Iterations -->
        <div class="param-item">
          <label class="text-xs text-gray-400 mb-2 block">Max Iterations</label>
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
          <div class="flex items-center gap-2 h-8">
            <NSwitch v-model:value="useRichContext" :disabled="isExecuting" size="small" />
            <span class="text-xs text-gray-300">Rich Context</span>
          </div>
        </div>

        <div class="param-item flex items-end">
          <div class="flex items-center gap-2 h-8">
            <NSwitch v-model:value="useReact" :disabled="isExecuting" size="small" />
            <span class="text-xs text-gray-300">ReAct Reasoning</span>
          </div>
        </div>

        <div class="param-item flex items-end">
          <div class="flex items-center gap-2 h-8">
            <NSwitch v-model:value="useFieldAlignment" :disabled="isExecuting" size="small" />
            <span class="text-xs text-gray-300">Field Alignment</span>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-3 mt-6">
        <NButton
          type="primary"
          size="large"
          :loading="isExecuting"
          :disabled="!question.trim()"
          @click="handleExecute"
        >
          <template #icon>
            <div class="i-carbon-play" />
          </template>
          Execute Query
        </NButton>

        <NButton
          v-if="isExecuting"
          type="error"
          size="large"
          @click="handleStop"
        >
          <template #icon>
            <div class="i-carbon-stop" />
          </template>
          Stop
        </NButton>

        <NButton
          quaternary
          size="large"
          :disabled="isExecuting"
          @click="handleClear"
        >
          <template #icon>
            <div class="i-carbon-clean" />
          </template>
          Clear
        </NButton>
      </div>
    </div>

    <!-- Real-time Execution Cards -->
    <div class="execution-pipeline grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
      <!-- Stage 1: Vector Search -->
      <RealtimeCard
        title="Vector Search"
        icon="i-carbon-search"
        :active="vectorSearchStage.active"
        :stage="workspaceStore.groundingStage"
        color="blue"
      >
        <template #content>
          <div v-if="workspaceStore.groundingResult" class="space-y-3">
            <div v-if="workspaceStore.groundingResult.tables?.length">
              <div class="text-xs text-gray-400 mb-2">Retrieved Tables:</div>
              <div class="flex flex-wrap gap-2">
                <div
                  v-for="table in workspaceStore.groundingResult.tables"
                  :key="table.name"
                  class="px-2 py-1 rounded bg-blue-500/10 border border-blue-500/30 text-xs text-blue-300"
                >
                  {{ table.name }}
                  <span class="text-gray-500 ml-1">{{ (table.confidence * 100).toFixed(0) }}%</span>
                </div>
              </div>
            </div>

            <div v-if="workspaceStore.groundingResult.columns?.length">
              <div class="text-xs text-gray-400 mb-2">Retrieved Columns:</div>
              <div class="flex flex-wrap gap-1">
                <div
                  v-for="col in workspaceStore.groundingResult.columns.slice(0, 8)"
                  :key="`${col.table}.${col.column}`"
                  class="px-2 py-1 rounded bg-blue-500/5 text-xs text-gray-400"
                >
                  {{ col.table }}.{{ col.column }}
                </div>
                <div v-if="workspaceStore.groundingResult.columns.length > 8" class="text-xs text-gray-500 px-2">
                  +{{ workspaceStore.groundingResult.columns.length - 8 }} more
                </div>
              </div>
            </div>
          </div>
          <div v-else-if="vectorSearchStage.active" class="text-sm text-gray-500">
            Searching vector database...
          </div>
          <div v-else class="text-sm text-gray-600">
            No data yet
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 2: Schema Linking -->
      <RealtimeCard
        title="ReAct Schema Linking"
        icon="i-carbon-connection-signal"
        :active="schemaLinkingStage.active"
        color="cyan"
      >
        <template #content>
          <div v-if="schemaLinkingStage.steps.length" class="space-y-2">
            <div
              v-for="step in schemaLinkingStage.steps"
              :key="step.step"
              class="react-step p-3 rounded-lg bg-cyan-500/5 border border-cyan-500/20"
            >
              <div class="flex items-start gap-2">
                <div class="w-5 h-5 rounded-full bg-cyan-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                  <div class="i-carbon-idea text-xs text-cyan-400" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs text-gray-400 mb-1">Step {{ step.step }}</div>
                  <p class="text-sm text-gray-300 leading-relaxed">{{ step.thought || step.observation }}</p>
                </div>
              </div>
            </div>
          </div>
          <div v-else-if="schemaLinkingStage.active" class="text-sm text-gray-500">
            Analyzing schema...
          </div>
          <div v-else class="text-sm text-gray-600">
            No data yet
          </div>
        </template>
      </RealtimeCard>

      <!-- Stage 3: SQL Generation -->
      <RealtimeCard
        title="ReAct SQL Generation"
        icon="i-carbon-code"
        :active="sqlGenerationStage.active"
        color="purple"
      >
        <template #content>
          <div v-if="sqlGenerationStage.steps.length" class="space-y-2">
            <div
              v-for="step in sqlGenerationStage.steps"
              :key="step.step"
              class="react-step p-3 rounded-lg bg-purple-500/5 border border-purple-500/20"
            >
              <div class="flex items-start gap-2">
                <div class="w-5 h-5 rounded-full bg-purple-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                  <div class="i-carbon-code text-xs text-purple-400" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-xs text-gray-400 mb-1">Step {{ step.step }}</div>
                  <p class="text-sm text-gray-300 leading-relaxed">{{ step.thought || step.observation }}</p>
                </div>
              </div>
            </div>
          </div>
          <div v-else-if="sqlGenerationStage.active" class="text-sm text-gray-500">
            Generating SQL...
          </div>
          <div v-else class="text-sm text-gray-600">
            No data yet
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
    />
  </div>
</template>

<style scoped>
.query-input :deep(.n-input__textarea-el) {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #e5e7eb;
  font-size: 0.95rem;
  font-weight: 400;
}

.query-input :deep(.n-input__textarea-el::placeholder) {
  color: rgba(255, 255, 255, 0.3);
}

.query-input :deep(.n-input__textarea-el):focus {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(59, 130, 246, 0.6);
  color: white;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.param-item :deep(.n-input-number),
.param-item :deep(.n-select) {
  background: rgba(255, 255, 255, 0.05);
}

.param-item :deep(.n-input-number .n-input__input-el),
.param-item :deep(.n-select .n-base-selection) {
  color: #e5e7eb;
}

.react-step {
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
