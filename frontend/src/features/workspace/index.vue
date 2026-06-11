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
import EvolutionPanel from './EvolutionPanel.vue'

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const databaseStore = useDatabaseStore()

const baseTabs: { key: WorkspaceTab; label: string; icon: string }[] = [
  { key: 'query', label: 'Query', icon: 'i-atlas-message-square' },
  { key: 'schema', label: 'Schema', icon: 'i-atlas-table-2' },
  { key: 'context', label: 'Context', icon: 'i-atlas-file-text' }
]

// Whether current database is the evolution demo DB
const isEvolutionDb = computed(() => {
  return workspaceStore.currentDatabase?.name === 'atlas_evolution'
})

// Database description for the banner
const dbDescription = computed(() => {
  const db = workspaceStore.currentDatabase
  if (db?.description) return db.description
  if (isSpiderMode.value) return 'Spider benchmark dataset — standardized Text-to-SQL evaluation with curated query scenarios'
  return ''
})

// Tabs — include Evolution tab only for evolution demo DB
const tabs = computed(() => {
  if (isEvolutionDb.value) {
    return [
      ...baseTabs,
      { key: 'evolution' as WorkspaceTab, label: 'Evolution', icon: 'i-atlas-git-branch' }
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
          <div class="i-atlas-database text-2xl text-primary-600 animate-pulse" />
        </div>
        <p class="text-gray-500 text-sm">Loading database schema...</p>
      </div>
    </div>

    <!-- Database not found -->
    <div v-else-if="!workspaceStore.currentDatabase" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-12 h-12 rounded-lg bg-red-50 flex items-center justify-center mx-auto mb-3">
          <div class="i-atlas-alert-triangle text-2xl text-red-500" />
        </div>
        <p class="text-lg text-gray-900 font-medium mb-1">Database not found</p>
        <p class="text-gray-400 text-sm mb-5">The database may not exist or is not connected</p>
        <NButton type="primary" size="small" @click="goBack">
          <template #icon>
            <div class="i-atlas-arrow-left" />
          </template>
          Back to Home
        </NButton>
      </div>
    </div>

    <!-- Workspace content -->
    <template v-else>
      <!-- Database header banner -->
      <div class="database-header bg-white border-b border-gray-200 px-6 py-5 sticky top-[56px] z-20">
        <div class="max-w-[1800px] mx-auto">
          <div class="flex items-start justify-between">
            <div class="flex items-start gap-4">
              <!-- Back button -->
              <button 
                class="group w-9 h-9 mt-0.5 rounded-lg bg-gray-100 flex items-center justify-center hover:bg-primary-50 transition-colors shrink-0"
                @click="goBack"
              >
                <div class="i-atlas-arrow-left text-lg text-gray-500 group-hover:text-primary-600 transition-colors" />
              </button>
              
              <!-- Icon -->
              <div 
                class="w-11 h-11 mt-px rounded-xl flex items-center justify-center shrink-0 shadow-sm border"
                :class="isSpiderMode 
                  ? 'bg-violet-50 text-violet-600 border-violet-100' 
                  : isEvolutionDb 
                    ? 'bg-amber-50 text-amber-600 border-amber-100'
                    : 'bg-primary-50 text-primary-600 border-primary-100'"
              >
                <span v-if="isSpiderMode" class="text-2xl">🕷️</span>
                <span v-else-if="isEvolutionDb" class="text-2xl">🧬</span>
                <div v-else class="i-atlas-database text-2xl" />
              </div>

              <!-- Info -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-3 flex-wrap">
                  <h1 class="text-xl font-bold text-gray-900 tracking-tight">
                    <template v-if="isSpiderMode">Spider Dataset</template>
                    <template v-else>{{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}</template>
                  </h1>
                  
                  <span 
                    class="px-2 py-0.5 rounded-md text-[11px] font-bold uppercase tracking-wider"
                    :class="isSpiderMode 
                      ? 'bg-violet-50 text-violet-600 border border-violet-100' 
                      : 'bg-gray-50 text-gray-500 border border-gray-200'"
                  >
                    {{ isSpiderMode ? 'Benchmark' : workspaceStore.currentDatabase.type }}
                  </span>
                  
                  <span v-if="!isSpiderMode && workspaceStore.currentDatabase.host" class="text-xs font-medium text-gray-400 flex items-center gap-1 bg-gray-50 px-2 py-0.5 rounded-md border border-gray-100">
                    <div class="i-atlas-server text-gray-400 text-[11px]" />
                    {{ workspaceStore.currentDatabase.host }}
                  </span>

                  <!-- Spider scenario switcher -->
                  <div v-if="isSpiderMode && spiderScenarios.length > 1" class="flex items-center gap-1.5 ml-2">
                    <button
                      v-for="scenario in spiderScenarios"
                      :key="scenario.id"
                      class="flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-semibold transition-all"
                      :class="scenario.id === currentScenarioId 
                        ? 'bg-violet-500 text-white shadow-sm ring-1 ring-violet-500/50 ring-offset-1' 
                        : 'bg-white text-gray-500 border border-gray-200 hover:border-violet-300 hover:text-violet-600 hover:bg-violet-50/50'"
                      @click="switchSpiderScenario(scenario.id)"
                    >
                      <span>{{ scenario.icon }}</span>
                      <span>{{ scenario.name }}</span>
                    </button>
                  </div>
                </div>
                
                <!-- Database description -->
                <p v-if="dbDescription" class="text-sm text-gray-500 mt-1.5 leading-relaxed max-w-3xl">
                  {{ dbDescription }}
                </p>
              </div>
            </div>

            <!-- Stats section on the right -->
            <div class="flex items-center gap-6 pr-2 pl-6 shrink-0 border-l border-gray-100 ml-6">
              <div class="flex flex-col items-start justify-center">
                <span class="text-[10px] uppercase tracking-widest font-bold text-gray-400 mb-0.5">Tables</span>
                <span class="text-xl font-black text-gray-700 tracking-tight flex items-center gap-1.5">
                  <div class="i-atlas-table text-sm text-gray-300" />
                  {{ workspaceStore.currentDatabase.tableCount }}
                </span>
              </div>
              
              <template v-if="workspaceStore.hasRichContext">
                <div class="w-px h-8 bg-gray-200"></div>
                <div class="flex flex-col items-start justify-center">
                  <span class="text-[10px] uppercase tracking-widest font-bold text-primary-500 mb-0.5">Rich Contexts</span>
                  <span class="text-xl font-black text-primary-600 tracking-tight flex items-center gap-1.5">
                    <div class="i-atlas-sparkles text-sm text-primary-400" />
                    {{ workspaceStore.contexts.length }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="tab-navigation bg-white border-b border-gray-200 px-6 sticky top-[56px] z-10">
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
        <EvolutionPanel v-else-if="workspaceStore.activeTab === 'evolution' && isEvolutionDb" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.workspace-content {
  min-height: calc(100vh - 160px);
}
</style>
