<script setup lang="ts">
import { onMounted, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NTabs, NTabPane, NButton } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { useDatabaseStore } from '@/stores/database'
import type { WorkspaceTab } from '@/types'

// Child components - lazy loaded
import QueryChat from './QueryChat/index.vue'
import SchemaBrowser from './SchemaBrowser/index.vue'
import ContextManager from './ContextManager/index.vue'
import Monitor from './Monitor/index.vue'
import EvolutionPanel from './EvolutionPanel.vue'

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const databaseStore = useDatabaseStore()

const baseTabs: { key: WorkspaceTab; label: string; icon: string }[] = [
  { key: 'query', label: 'Query', icon: 'i-lucide-message-square' },
  { key: 'schema', label: 'Schema', icon: 'i-lucide-table-2' },
  { key: 'context', label: 'Context', icon: 'i-lucide-file-text' },
  { key: 'monitor', label: 'Monitor', icon: 'i-lucide-bar-chart-3' }
]

// Whether current database is the evolution demo DB
const isEvolutionDb = computed(() => {
  return workspaceStore.currentDatabase?.name === 'lucid_evolution'
})

// Tabs — include Evolution tab only for evolution demo DB
const tabs = computed(() => {
  if (isEvolutionDb.value) {
    return [
      ...baseTabs,
      { key: 'evolution' as WorkspaceTab, label: 'Evolution', icon: 'i-lucide-git-branch' }
    ]
  }
  return baseTabs
})

// Spider database name patterns
const SPIDER_PATTERNS = ['spider_tvshow', 'spider_flight', 'spider_wta']

// Check if database belongs to Spider dataset
function isSpiderDatabase(name: string): boolean {
  return name.toLowerCase().startsWith('spider_')
}

// Whether currently in Spider mode
const isSpiderMode = computed(() => {
  return workspaceStore.currentDatabase && isSpiderDatabase(workspaceStore.currentDatabase.name)
})

// Spider scenario list
const spiderScenarios = computed(() => {
  if (!isSpiderMode.value) return []
  
  // Get all Spider databases
  return databaseStore.databases
    .filter(db => isSpiderDatabase(db.name))
    .map(db => ({
      id: db.id,
      name: getScenarioName(db.name),
      icon: getScenarioIcon(db.name),
      tag: getScenarioTag(db.name),
      dbName: db.name
    }))
})

// Currently selected Spider scenario
const currentScenarioId = computed(() => workspaceStore.currentDatabaseId)

function getScenarioName(dbName: string): string {
  if (dbName.includes('tvshow') || dbName.includes('tv_show')) return 'TV Show'
  if (dbName.includes('flight')) return 'Flight'
  if (dbName.includes('wta')) return 'WTA Tennis'
  return dbName
}

function getScenarioIcon(dbName: string): string {
  if (dbName.includes('tvshow') || dbName.includes('tv_show')) return '📺'
  if (dbName.includes('flight')) return '✈️'
  if (dbName.includes('wta')) return '🎾'
  return '📊'
}

function getScenarioTag(_dbName: string): string {
  // All Spider databases are dirty-data scenarios, no distinction needed
  return ''
}

// Switch Spider scenario
function switchSpiderScenario(scenarioId: string) {
  if (scenarioId !== workspaceStore.currentDatabaseId) {
    router.push(`/workspace/${scenarioId}`)
  }
}

onMounted(async () => {
  // Ensure databases are loaded
  if (databaseStore.databases.length === 0) {
    await databaseStore.fetchDatabases()
  }
  
  // Select database from route
  const dbId = route.params.databaseId as string
  if (dbId && dbId !== workspaceStore.currentDatabaseId) {
    await workspaceStore.selectDatabase(dbId)
  }
})

// Watch route changes
watch(
  () => route.params.databaseId,
  async (newId) => {
    if (newId && newId !== workspaceStore.currentDatabaseId) {
      await workspaceStore.selectDatabase(newId as string)
    }
  }
)

function handleTabChange(tab: string | number) {
  workspaceStore.setActiveTab(tab as WorkspaceTab)
}

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="workspace-page min-h-screen bg-slate-50">
    <!-- Loading state -->
    <div v-if="workspaceStore.loadingSchema" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-12 h-12 rounded-lg bg-primary-50 flex items-center justify-center mx-auto mb-3">
          <div class="i-lucide-database text-2xl text-primary-600 animate-pulse" />
        </div>
        <p class="text-gray-500 text-sm">Loading database schema...</p>
      </div>
    </div>

    <!-- Database not found -->
    <div v-else-if="!workspaceStore.currentDatabase" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-12 h-12 rounded-lg bg-red-50 flex items-center justify-center mx-auto mb-3">
          <div class="i-lucide-alert-triangle text-2xl text-red-500" />
        </div>
        <p class="text-lg text-gray-900 font-medium mb-1">Database not found</p>
        <p class="text-gray-400 text-sm mb-5">The database may not exist or is not connected</p>
        <NButton type="primary" size="small" @click="goBack">
          <template #icon>
            <div class="i-lucide-arrow-left" />
          </template>
          Back to Home
        </NButton>
      </div>
    </div>

    <!-- Workspace content -->
    <template v-else>
      <!-- Database header -->
      <div class="database-header bg-white border-b border-gray-200 px-6 py-4 sticky top-0 z-20">
        <div class="max-w-[1800px] mx-auto">
          <div class="flex items-center gap-4">
            <button 
              class="group w-9 h-9 rounded-lg bg-gray-100 flex items-center justify-center hover:bg-primary-50 transition-colors"
              @click="goBack"
            >
              <div class="i-lucide-arrow-left text-lg text-gray-500 group-hover:text-primary-600 transition-colors" />
            </button>
            
            <div class="flex items-center gap-3 flex-1">
              <!-- Spider mode: special icon -->
              <div 
                class="w-10 h-10 rounded-lg flex items-center justify-center"
                :class="isSpiderMode 
                  ? 'bg-violet-100 text-violet-600' 
                  : isEvolutionDb 
                    ? 'bg-amber-100 text-amber-600'
                    : 'bg-primary-50 text-primary-600'"
              >
                <span v-if="isSpiderMode" class="text-xl">🕷️</span>
                <span v-else-if="isEvolutionDb" class="text-xl">🧬</span>
                <div v-else class="i-lucide-database text-xl" />
              </div>

              <div class="flex-1">
                <div class="flex items-center gap-3">
                  <h1 class="text-lg font-semibold text-gray-900">
                    <template v-if="isSpiderMode">
                      Spider Dataset
                    </template>
                    <template v-else>
                      {{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}
                    </template>
                  </h1>
                  
                  <!-- Spider scenario switcher -->
                  <div v-if="isSpiderMode && spiderScenarios.length > 1" class="flex items-center gap-1.5">
                    <button
                      v-for="scenario in spiderScenarios"
                      :key="scenario.id"
                      class="flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium transition-colors"
                      :class="scenario.id === currentScenarioId 
                        ? 'bg-violet-100 text-violet-700' 
                        : 'bg-gray-100 text-gray-500 hover:bg-violet-50 hover:text-violet-600'"
                      @click="switchSpiderScenario(scenario.id)"
                    >
                      <span>{{ scenario.icon }}</span>
                      <span>{{ scenario.name }}</span>
                    </button>
                  </div>
                </div>
                
                <div class="flex items-center gap-2.5 mt-1">
                  <span 
                    class="px-2 py-0.5 rounded text-xs font-medium uppercase tracking-wide"
                    :class="isSpiderMode 
                      ? 'bg-violet-50 text-violet-600' 
                      : 'bg-gray-100 text-gray-600'"
                  >
                    {{ isSpiderMode ? 'Text-to-SQL Benchmark' : workspaceStore.currentDatabase.type }}
                  </span>
                  <span v-if="!isSpiderMode && workspaceStore.currentDatabase.host" class="text-sm text-gray-400 flex items-center gap-1">
                    <div class="i-lucide-server text-gray-400 text-xs" />
                    {{ workspaceStore.currentDatabase.host }}
                  </span>
                  <div class="w-1 h-1 rounded-full bg-gray-300"></div>
                  <span class="text-sm text-gray-500">
                    {{ workspaceStore.currentDatabase.tableCount }} tables
                  </span>
                  <template v-if="workspaceStore.hasRichContext">
                    <div class="w-1 h-1 rounded-full bg-primary-400"></div>
                    <span class="text-sm text-primary-600 flex items-center gap-1 px-2 py-0.5 rounded bg-primary-50">
                      <div class="i-lucide-sparkles text-xs" />
                      {{ workspaceStore.contexts.length }} contexts
                    </span>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="tab-navigation bg-white border-b border-gray-200 px-6 sticky top-[73px] z-10">
        <div class="max-w-[1800px] mx-auto">
          <NTabs 
            type="line"
            size="medium"
            :value="workspaceStore.activeTab"
            @update:value="handleTabChange"
          >
            <NTabPane 
              v-for="tab in tabs" 
              :key="tab.key" 
              :name="tab.key"
              :tab="tab.label"
            >
              <template #tab>
                <div class="flex items-center gap-2 py-0.5">
                  <div :class="[tab.icon, 'text-base']" />
                  <span class="text-sm font-medium">{{ tab.label }}</span>
                </div>
              </template>
            </NTabPane>
          </NTabs>
        </div>
      </div>

      <!-- Tab content -->
      <div class="workspace-content max-w-[1800px] mx-auto p-6">
        <QueryChat v-if="workspaceStore.activeTab === 'query'" />
        <SchemaBrowser v-else-if="workspaceStore.activeTab === 'schema'" />
        <ContextManager v-else-if="workspaceStore.activeTab === 'context'" />
        <Monitor v-else-if="workspaceStore.activeTab === 'monitor'" />
        <EvolutionPanel v-else-if="workspaceStore.activeTab === 'evolution' && isEvolutionDb" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.workspace-content {
  min-height: calc(100vh - 180px);
}
</style>
