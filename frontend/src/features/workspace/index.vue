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

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()
const databaseStore = useDatabaseStore()

const tabs: { key: WorkspaceTab; label: string; icon: string }[] = [
  { key: 'query', label: '对话查询', icon: 'i-carbon-chat' },
  { key: 'schema', label: 'Schema', icon: 'i-carbon-data-table' },
  { key: 'context', label: 'Context', icon: 'i-carbon-document' },
  { key: 'monitor', label: '监控', icon: 'i-carbon-analytics' }
]

// Spider 库名称模式
const SPIDER_PATTERNS = ['spider_tvshow', 'spider_flight', 'spider_wta']

// 判断是否是 Spider 库
function isSpiderDatabase(name: string): boolean {
  return name.toLowerCase().startsWith('spider_')
}

// 当前是否在 Spider 模式
const isSpiderMode = computed(() => {
  return workspaceStore.currentDatabase && isSpiderDatabase(workspaceStore.currentDatabase.name)
})

// Spider 场景列表
const spiderScenarios = computed(() => {
  if (!isSpiderMode.value) return []
  
  // 获取所有 Spider 库
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

// 当前选中的 Spider 场景
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
  // 所有 Spider 库都是脏库场景，不做区分
  return ''
}

// 切换 Spider 场景
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
  <div class="workspace-page min-h-screen bg-gradient-to-br from-slate-100 via-gray-50 to-blue-50/50">
    <!-- Loading state -->
    <div v-if="workspaceStore.loadingSchema" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-16 h-16 rounded-2xl bg-blue-50 flex items-center justify-center mx-auto mb-4">
          <div class="i-carbon-data-base text-3xl text-primary-600 animate-pulse" />
        </div>
        <p class="text-gray-500 font-medium">Loading database schema...</p>
      </div>
    </div>

    <!-- Database not found -->
    <div v-else-if="!workspaceStore.currentDatabase" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-16 h-16 rounded-2xl bg-red-50 flex items-center justify-center mx-auto mb-4">
          <div class="i-carbon-warning text-3xl text-red-500" />
        </div>
        <p class="text-xl text-gray-900 font-bold mb-2">Database not found</p>
        <p class="text-gray-500 mb-6">The database may not exist or is not connected</p>
        <NButton type="primary" @click="goBack">
          <template #icon>
            <div class="i-carbon-arrow-left" />
          </template>
          Back to Home
        </NButton>
      </div>
    </div>

    <!-- Workspace content -->
    <template v-else>
      <!-- Database header - modernized -->
      <div class="database-header bg-gradient-to-r from-white via-white to-slate-50/80 border-b border-gray-200/80 px-8 py-5 sticky top-0 z-20 backdrop-blur-sm">
        <div class="max-w-[1800px] mx-auto">
          <div class="flex items-center gap-5">
            <button 
              class="group w-11 h-11 rounded-xl bg-gradient-to-br from-gray-100 to-slate-200 flex items-center justify-center shadow-md hover:shadow-lg hover:from-primary-50 hover:to-blue-100 hover:-translate-y-0.5 transition-all duration-200"
              @click="goBack"
            >
              <div class="i-carbon-arrow-left text-xl text-gray-600 group-hover:text-primary-600 transition-colors" />
            </button>
            
            <div class="flex items-center gap-4 flex-1">
              <!-- Spider 模式显示特殊图标 -->
              <div 
                class="w-12 h-12 rounded-xl flex items-center justify-center shadow-lg"
                :class="isSpiderMode 
                  ? 'bg-gradient-to-br from-violet-500 to-purple-600 shadow-violet-500/30' 
                  : 'bg-gradient-to-br from-primary-500 to-blue-600 shadow-primary-500/30'"
              >
                <span v-if="isSpiderMode" class="text-2xl">🕷️</span>
                <div v-else class="i-carbon-data-base text-2xl text-white" />
              </div>

              <div class="flex-1">
                <div class="flex items-center gap-4">
                  <h1 class="text-2xl font-bold text-gray-900 leading-tight">
                    <template v-if="isSpiderMode">
                      Spider Dataset
                    </template>
                    <template v-else>
                      {{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}
                    </template>
                  </h1>
                  
                  <!-- Spider 场景切换 - 放在标题旁边 -->
                  <div v-if="isSpiderMode && spiderScenarios.length > 1" class="flex items-center gap-2">
                    <button
                      v-for="scenario in spiderScenarios"
                      :key="scenario.id"
                      class="scenario-btn flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all duration-200"
                      :class="scenario.id === currentScenarioId 
                        ? 'bg-violet-100 text-violet-700 ring-1 ring-violet-300' 
                        : 'bg-gray-100 text-gray-600 hover:bg-violet-50 hover:text-violet-600'"
                      @click="switchSpiderScenario(scenario.id)"
                    >
                      <span>{{ scenario.icon }}</span>
                      <span>{{ scenario.name }}</span>
                    </button>
                  </div>
                </div>
                
                <div class="flex items-center gap-3 mt-1.5">
                  <span 
                    class="px-2.5 py-1 rounded-lg text-xs font-bold uppercase tracking-wide shadow-sm"
                    :class="isSpiderMode 
                      ? 'bg-gradient-to-r from-violet-100 to-purple-100 text-violet-700' 
                      : 'bg-gradient-to-r from-gray-100 to-slate-200 text-gray-700'"
                  >
                    {{ isSpiderMode ? 'Text-to-SQL Benchmark' : workspaceStore.currentDatabase.type }}
                  </span>
                  <span v-if="!isSpiderMode && workspaceStore.currentDatabase.host" class="text-sm font-medium text-gray-500 flex items-center gap-1.5">
                    <div class="i-carbon-ibm-cloud-citrix-daas text-gray-400" />
                    {{ workspaceStore.currentDatabase.host }}
                  </span>
                  <div class="w-1.5 h-1.5 rounded-full bg-gray-300"></div>
                  <span class="text-sm font-semibold text-gray-600">
                    {{ workspaceStore.currentDatabase.tableCount }} tables
                  </span>
                  <template v-if="workspaceStore.hasRichContext">
                    <div class="w-1.5 h-1.5 rounded-full bg-primary-400"></div>
                    <span class="text-sm font-bold text-primary-600 flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-primary-50">
                      <div class="i-carbon-magic-wand" />
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
      <div class="tab-navigation bg-white border-b border-gray-200 px-8 sticky top-[105px] z-10">
        <div class="max-w-[1800px] mx-auto">
          <NTabs 
            type="line"
            size="large"
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
                <div class="flex items-center gap-2.5 py-1">
                  <div :class="[tab.icon, 'text-lg']" />
                  <span class="text-base font-medium">{{ tab.label }}</span>
                </div>
              </template>
            </NTabPane>
          </NTabs>
        </div>
      </div>

      <!-- Tab content -->
      <div class="workspace-content max-w-[1800px] mx-auto p-8">
        <QueryChat v-if="workspaceStore.activeTab === 'query'" />
        <SchemaBrowser v-else-if="workspaceStore.activeTab === 'schema'" />
        <ContextManager v-else-if="workspaceStore.activeTab === 'context'" />
        <Monitor v-else-if="workspaceStore.activeTab === 'monitor'" />
      </div>
    </template>
  </div>
</template>

<style scoped>
.workspace-content {
  min-height: calc(100vh - 200px);
}
</style>
