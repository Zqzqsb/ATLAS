<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue'
import { 
  NStatistic, NCard, NGrid, NGridItem, NDataTable, NTag, NEmpty, NButton, 
  NSwitch, NModal, NInput, NCode, NScrollbar, NTimeline, NTimelineItem,
  useMessage
} from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { agentApi, databaseApi } from '@/api'
import type { AgentStatus, ChangeLog, MaintenanceResult } from '@/api/agent'

const workspaceStore = useWorkspaceStore()
const message = useMessage()

// Agent status
const agentStatus = ref<AgentStatus | null>(null)
const loadingAgent = ref(false)

// Change logs
const changeLogs = ref<ChangeLog[]>([])
const loadingLogs = ref(false)

// Query stats
const stats = ref({
  queryCount: 0,
  avgDuration: 0,
  successRate: 0,
  contextUsageRate: 0
})

// DDL Simulation modal
const showDDLModal = ref(false)
const ddlInput = ref('')
const simulatingDDL = ref(false)
const ddlResult = ref<any>(null)

// Computed
const datasourceId = computed(() => {
  return workspaceStore.currentDatabase?.metadata?.lakebaseId || 1
})

const agentRunning = computed(() => agentStatus.value?.running ?? false)

// Methods
async function fetchAgentStatus() {
  loadingAgent.value = true
  try {
    agentStatus.value = await agentApi.getStatus()
  } catch (e: any) {
    console.error('Failed to fetch agent status:', e)
  } finally {
    loadingAgent.value = false
  }
}

async function fetchChangeLogs() {
  loadingLogs.value = true
  try {
    const result = await agentApi.getChangeLogs(datasourceId.value, 50)
    changeLogs.value = result.logs || []
  } catch (e: any) {
    console.error('Failed to fetch change logs:', e)
  } finally {
    loadingLogs.value = false
  }
}

async function fetchStats() {
  if (!workspaceStore.currentDatabaseId) return
  try {
    stats.value = await databaseApi.getStats(workspaceStore.currentDatabaseId)
  } catch (e) {
    // Stats API might not exist, use mock data
    stats.value = {
      queryCount: workspaceStore.queryHistory.length,
      avgDuration: 1.2,
      successRate: 0.85,
      contextUsageRate: 0.72
    }
  }
}

async function toggleAgentService(running: boolean) {
  try {
    if (running) {
      await agentApi.start()
      message.success('Agent service started')
    } else {
      await agentApi.stop()
      message.success('Agent service stopped')
    }
    await fetchAgentStatus()
  } catch (e: any) {
    message.error(e.message || 'Failed to toggle agent service')
  }
}

async function runMaintenance() {
  const msgReactive = message.loading('Running maintenance...', { duration: 0 })
  try {
    const result = await agentApi.runMaintenance(datasourceId.value)
    msgReactive.destroy()
    message.success('Maintenance completed')
    await fetchAgentStatus()
    await fetchChangeLogs()
  } catch (e: any) {
    msgReactive.destroy()
    message.error(e.message || 'Maintenance failed')
  }
}

async function refreshContext() {
  const msgReactive = message.loading('Refreshing expired context...', { duration: 0 })
  try {
    const result = await agentApi.triggerContextRefresh(datasourceId.value)
    msgReactive.destroy()
    message.success(`Refreshed ${result.success_count}/${result.total} contexts`)
    await fetchChangeLogs()
  } catch (e: any) {
    msgReactive.destroy()
    message.error(e.message || 'Refresh failed')
  }
}

function openDDLModal() {
  ddlInput.value = ''
  ddlResult.value = null
  showDDLModal.value = true
}

async function simulateDDL() {
  if (!ddlInput.value.trim()) {
    message.warning('Please enter a DDL statement')
    return
  }
  
  simulatingDDL.value = true
  try {
    ddlResult.value = await agentApi.simulateDDL(datasourceId.value, ddlInput.value)
    message.success('DDL processed successfully')
    await fetchChangeLogs()
  } catch (e: any) {
    message.error(e.message || 'DDL simulation failed')
  } finally {
    simulatingDDL.value = false
  }
}

// Table columns for change logs
const logColumns = [
  {
    title: 'Time',
    key: 'created_at',
    width: 160,
    render: (row: ChangeLog) => new Date(row.created_at).toLocaleString()
  },
  {
    title: 'Type',
    key: 'change_type',
    width: 140,
    render: (row: ChangeLog) => {
      const typeMap: Record<string, { type: 'info' | 'warning' | 'success' | 'error', label: string }> = {
        'schema_change': { type: 'warning', label: 'Schema Change' },
        'context_update': { type: 'success', label: 'Context Update' },
        'context_expire': { type: 'error', label: 'Context Expired' }
      }
      const config = typeMap[row.change_type] || { type: 'info', label: row.change_type }
      return h(NTag, { type: config.type, size: 'small' }, { default: () => config.label })
    }
  },
  {
    title: 'Table',
    key: 'table_name',
    width: 140
  },
  {
    title: 'Source',
    key: 'trigger_source',
    width: 100,
    render: (row: ChangeLog) => {
      const sourceMap: Record<string, string> = {
        'agent': 'Agent',
        'user': 'User',
        'system': 'System'
      }
      return sourceMap[row.trigger_source] || row.trigger_source
    }
  },
  {
    title: 'Reason',
    key: 'change_reason',
    ellipsis: { tooltip: true }
  }
]

// Query history columns
const historyColumns = [
  {
    title: 'Time',
    key: 'timestamp',
    width: 160,
    render: (row: any) => new Date(row.timestamp).toLocaleString()
  },
  {
    title: 'Question',
    key: 'question',
    ellipsis: { tooltip: true }
  },
  {
    title: 'Duration',
    key: 'duration',
    width: 100,
    render: (row: any) => `${row.duration.toFixed(2)}s`
  },
  {
    title: 'Status',
    key: 'feedback',
    width: 100,
    render: (row: any) => {
      if (row.feedback === 'positive') {
        return h(NTag, { type: 'success', size: 'small' }, { default: () => 'Correct' })
      }
      if (row.feedback === 'negative') {
        return h(NTag, { type: 'error', size: 'small' }, { default: () => 'Wrong' })
      }
      return h(NTag, { size: 'small' }, { default: () => 'Pending' })
    }
  }
]

// Example DDL statements
const ddlExamples = [
  'ALTER TABLE tv_channel ADD COLUMN rating DECIMAL(3,1)',
  'ALTER TABLE cartoon DROP COLUMN legacy_flag',
  'ALTER TABLE tv_series MODIFY COLUMN episode_count INT NOT NULL',
  'CREATE TABLE tv_schedule (id INT PRIMARY KEY, channel_id INT, program_name VARCHAR(200))',
  'DROP TABLE temp_data'
]

onMounted(async () => {
  await Promise.all([
    fetchAgentStatus(),
    fetchChangeLogs(),
    fetchStats()
  ])
})
</script>

<template>
  <div class="monitor-page p-6 bg-gray-50 min-h-full">
    <!-- Agent Control Panel -->
    <div class="agent-panel mb-8 p-8 rounded-xl bg-white border border-gray-200 shadow-sm">
      <div class="flex items-center justify-between mb-8">
        <div class="flex items-center gap-6">
          <div class="w-14 h-14 rounded-2xl bg-primary-50 flex items-center justify-center border border-primary-100">
            <div class="i-carbon-bot text-3xl text-primary-600" />
          </div>
          <div>
            <h2 class="text-2xl font-bold text-gray-900">Agent Self-Maintenance</h2>
            <p class="text-base text-gray-500 mt-1">
              Automatic DDL detection and context synchronization
            </p>
          </div>
        </div>
        
        <div class="flex items-center gap-6">
          <div class="flex items-center gap-3">
            <span class="text-sm font-semibold text-gray-600">Agent Service</span>
            <NSwitch 
              :value="agentRunning" 
              :loading="loadingAgent"
              @update:value="toggleAgentService"
              size="large"
            />
          </div>
          
          <div 
            class="flex items-center gap-2 px-4 py-2 rounded-lg border transition-all"
            :class="agentRunning ? 'bg-green-50 border-green-200 text-green-700' : 'bg-gray-100 border-gray-200 text-gray-600'"
          >
            <div 
              class="w-2.5 h-2.5 rounded-full"
              :class="agentRunning ? 'bg-green-500 animate-pulse' : 'bg-gray-400'"
            />
            <span class="font-bold text-sm">
              {{ agentRunning ? 'Running' : 'Stopped' }}
            </span>
          </div>
        </div>
      </div>
      
      <!-- Action buttons -->
      <div class="flex items-center gap-4 border-b border-gray-100 pb-8 mb-8">
        <NButton 
          type="primary" 
          size="large" 
          @click="runMaintenance" 
          :disabled="!agentRunning"
          class="!rounded-full shadow-lg shadow-primary-500/20 hover:shadow-xl hover:shadow-primary-500/30 hover:-translate-y-0.5 transition-all duration-300 font-bold"
        >
          <template #icon>
            <div class="i-carbon-renew" />
          </template>
          Run Maintenance
        </NButton>
        
        <NButton 
          size="large" 
          @click="refreshContext" 
          :disabled="!agentRunning"
          class="!rounded-full hover:-translate-y-0.5 transition-all duration-300 font-bold"
        >
          <template #icon>
            <div class="i-carbon-reset" />
          </template>
          Refresh Expired Context
        </NButton>
        
        <NButton 
          type="warning" 
          size="large" 
          ghost 
          @click="openDDLModal"
          class="!rounded-full hover:-translate-y-0.5 transition-all duration-300 font-bold"
        >
          <template #icon>
            <div class="i-carbon-sql" />
          </template>
          Simulate DDL Change
        </NButton>
        
        <div class="flex-1"></div>
        
        <NButton 
          quaternary 
          size="large" 
          @click="fetchChangeLogs"
          class="!rounded-full hover:bg-gray-100 text-gray-500 hover:text-gray-700"
        >
          <template #icon>
            <div class="i-carbon-refresh" />
          </template>
          Refresh Logs
        </NButton>
      </div>
      
      <!-- Last maintenance result -->
      <div v-if="agentStatus?.last_result" class="p-6 rounded-xl bg-gray-50 border border-gray-200">
        <div class="flex items-center gap-2 mb-4">
          <div class="i-carbon-time text-gray-500" />
          <span class="text-sm font-medium text-gray-500">
            Last Run: {{ new Date(agentStatus.last_run).toLocaleString() }}
          </span>
        </div>
        <div class="grid grid-cols-5 gap-8">
          <div class="text-center">
            <div class="text-3xl font-bold text-gray-900">{{ agentStatus.last_result.schema_changes_found }}</div>
            <div class="text-sm font-medium text-gray-500 mt-1">Schema Changes</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-amber-500">{{ agentStatus.last_result.context_expired }}</div>
            <div class="text-sm font-medium text-gray-500 mt-1">Context Expired</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-green-600">{{ agentStatus.last_result.context_refreshed }}</div>
            <div class="text-sm font-medium text-gray-500 mt-1">Context Refreshed</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold text-primary-600">{{ agentStatus.last_result.context_created }}</div>
            <div class="text-sm font-medium text-gray-500 mt-1">Context Created</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-bold" :class="agentStatus.last_result.success ? 'text-green-600' : 'text-red-500'">
              {{ agentStatus.last_result.duration_ms }}ms
            </div>
            <div class="text-sm font-medium text-gray-500 mt-1">Duration</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Stats cards -->
    <NGrid :x-gap="24" :y-gap="24" :cols="4" class="mb-8">
      <NGridItem>
        <div class="stat-card p-6 rounded-xl bg-white border border-gray-200 hover:shadow-lg hover:border-blue-200 transition-all">
          <div class="flex items-center gap-4 mb-4">
            <div class="w-12 h-12 rounded-xl bg-blue-50 flex items-center justify-center">
              <div class="i-carbon-query text-2xl text-blue-600" />
            </div>
            <span class="text-sm font-bold text-gray-500 uppercase tracking-wide">Total Queries</span>
          </div>
          <div class="text-4xl font-bold text-gray-900">{{ stats.queryCount }}</div>
        </div>
      </NGridItem>
      
      <NGridItem>
        <div class="stat-card p-6 rounded-xl bg-white border border-gray-200 hover:shadow-lg hover:border-green-200 transition-all">
          <div class="flex items-center gap-4 mb-4">
            <div class="w-12 h-12 rounded-xl bg-green-50 flex items-center justify-center">
              <div class="i-carbon-time text-2xl text-green-600" />
            </div>
            <span class="text-sm font-bold text-gray-500 uppercase tracking-wide">Avg Duration</span>
          </div>
          <div class="text-4xl font-bold text-gray-900">{{ stats.avgDuration.toFixed(2) }}<span class="text-2xl text-gray-400 ml-1">s</span></div>
        </div>
      </NGridItem>
      
      <NGridItem>
        <div class="stat-card p-6 rounded-xl bg-white border border-gray-200 hover:shadow-lg hover:border-amber-200 transition-all">
          <div class="flex items-center gap-4 mb-4">
            <div class="w-12 h-12 rounded-xl bg-amber-50 flex items-center justify-center">
              <div class="i-carbon-checkmark-filled text-2xl text-amber-600" />
            </div>
            <span class="text-sm font-bold text-gray-500 uppercase tracking-wide">Success Rate</span>
          </div>
          <div class="text-4xl font-bold text-gray-900">{{ (stats.successRate * 100).toFixed(1) }}<span class="text-2xl text-gray-400 ml-1">%</span></div>
        </div>
      </NGridItem>
      
      <NGridItem>
        <div class="stat-card p-6 rounded-xl bg-white border border-gray-200 hover:shadow-lg hover:border-purple-200 transition-all">
          <div class="flex items-center gap-4 mb-4">
            <div class="w-12 h-12 rounded-xl bg-purple-50 flex items-center justify-center">
              <div class="i-carbon-magic-wand text-2xl text-purple-600" />
            </div>
            <span class="text-sm font-bold text-gray-500 uppercase tracking-wide">Context Usage</span>
          </div>
          <div class="text-4xl font-bold text-gray-900">{{ (stats.contextUsageRate * 100).toFixed(1) }}<span class="text-2xl text-gray-400 ml-1">%</span></div>
        </div>
      </NGridItem>
    </NGrid>

    <!-- Change logs and Query history -->
    <NGrid :x-gap="24" :cols="2">
      <!-- Change Logs -->
      <NGridItem>
        <div class="log-panel p-6 rounded-xl bg-white border border-gray-200 shadow-sm h-full">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="i-carbon-activity text-xl text-primary-600" />
              <h3 class="text-xl font-bold text-gray-900">Change Logs</h3>
            </div>
            <NTag type="info" size="small" round>{{ changeLogs.length }} records</NTag>
          </div>
          
          <NScrollbar style="max-height: 500px;">
            <NDataTable
              v-if="changeLogs.length > 0"
              :columns="logColumns"
              :data="changeLogs"
              :loading="loadingLogs"
              :bordered="false"
              size="small"
              striped
            />
            <NEmpty v-else description="No change logs yet" class="py-12" />
          </NScrollbar>
        </div>
      </NGridItem>
      
      <!-- Query History -->
      <NGridItem>
        <div class="log-panel p-6 rounded-xl bg-white border border-gray-200 shadow-sm h-full">
          <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
              <div class="i-carbon-recently-viewed text-xl text-blue-600" />
              <h3 class="text-xl font-bold text-gray-900">Query History</h3>
            </div>
            <NButton secondary size="tiny" @click="workspaceStore.fetchQueryHistory">
              <template #icon>
                <div class="i-carbon-refresh" />
              </template>
              Refresh
            </NButton>
          </div>
          
          <NScrollbar style="max-height: 500px;">
            <NDataTable
              v-if="workspaceStore.queryHistory.length > 0"
              :columns="historyColumns"
              :data="workspaceStore.queryHistory"
              :loading="workspaceStore.loadingHistory"
              :bordered="false"
              size="small"
              striped
            />
            <NEmpty v-else description="No query history" class="py-12" />
          </NScrollbar>
        </div>
      </NGridItem>
    </NGrid>

    <!-- DDL Simulation Modal -->
    <NModal v-model:show="showDDLModal" preset="card" title="Simulate DDL Change" style="width: 700px;" size="huge">
      <div class="space-y-6">
        <div class="bg-blue-50 p-4 rounded-lg border border-blue-100 flex gap-3">
          <div class="i-carbon-information text-blue-600 text-lg mt-0.5 shrink-0" />
          <p class="text-sm text-blue-800">
            Enter a DDL statement to simulate a schema change. The Agent will detect this change and automatically update the affected Rich Context without manual intervention.
          </p>
        </div>
        
        <div>
          <label class="text-sm font-bold text-gray-700 mb-2 block">DDL Statement</label>
          <NInput
            v-model:value="ddlInput"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 6 }"
            placeholder="e.g. ALTER TABLE tv_channel ADD COLUMN rating DECIMAL(3,1)"
            class="font-mono text-sm"
          />
        </div>
        
        <div>
          <label class="text-sm font-bold text-gray-700 mb-2 block">Quick Examples:</label>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="example in ddlExamples"
              :key="example"
              class="text-xs px-3 py-1.5 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200 hover:text-gray-900 transition-colors border border-gray-200 font-mono"
              @click="ddlInput = example"
            >
              {{ example.substring(0, 45) }}...
            </button>
          </div>
        </div>
        
        <!-- Result display -->
        <div v-if="ddlResult" class="p-5 rounded-lg bg-gray-50 border border-gray-200">
          <div class="flex items-center gap-2 mb-4">
            <div class="i-carbon-checkmark-filled text-green-600" />
            <span class="font-bold text-green-700">DDL Processed Successfully</span>
          </div>
          
          <div v-if="ddlResult.parsed_change" class="space-y-3 text-sm bg-white p-4 rounded border border-gray-100">
            <div class="flex items-center gap-2">
              <span class="text-gray-500 font-medium w-24">Change Type:</span>
              <NTag type="warning" size="small">{{ ddlResult.parsed_change.change_type }}</NTag>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-gray-500 font-medium w-24">Table:</span>
              <span class="font-mono text-gray-900 bg-gray-100 px-1.5 rounded">{{ ddlResult.parsed_change.table_name }}</span>
            </div>
            <div v-if="ddlResult.parsed_change.column_name" class="flex items-center gap-2">
              <span class="text-gray-500 font-medium w-24">Column:</span>
              <span class="font-mono text-gray-900 bg-gray-100 px-1.5 rounded">{{ ddlResult.parsed_change.column_name }}</span>
            </div>
          </div>
          
          <div v-if="ddlResult.result" class="mt-4 pt-4 border-t border-gray-200">
            <div class="text-xs font-bold text-gray-500 uppercase tracking-wider mb-3">Maintenance Result</div>
            <div class="grid grid-cols-3 gap-4 text-sm">
              <div class="bg-white p-3 rounded border border-gray-100 text-center">
                <span class="block text-gray-500 text-xs mb-1">Context Expired</span>
                <span class="text-xl font-bold text-amber-500">{{ ddlResult.result.context_expired }}</span>
              </div>
              <div class="bg-white p-3 rounded border border-gray-100 text-center">
                <span class="block text-gray-500 text-xs mb-1">Context Created</span>
                <span class="text-xl font-bold text-green-600">{{ ddlResult.result.context_created }}</span>
              </div>
              <div class="bg-white p-3 rounded border border-gray-100 text-center">
                <span class="block text-gray-500 text-xs mb-1">Duration</span>
                <span class="text-xl font-bold text-primary-600">{{ ddlResult.result.duration_ms }}ms</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showDDLModal = false" size="medium">Cancel</NButton>
          <NButton type="primary" :loading="simulatingDDL" @click="simulateDDL" size="medium">
            <template #icon>
              <div class="i-carbon-play" />
            </template>
            Execute Change
          </NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.stat-card {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.05);
}

:deep(.n-data-table .n-data-table-th) {
  font-weight: 700;
  color: #374151; /* Gray 700 */
}
</style>
