<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NSpin, NEmpty, NButton } from 'naive-ui'
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

function handleTabChange(tab: WorkspaceTab) {
  workspaceStore.setActiveTab(tab)
}

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="workspace-page min-h-screen bg-gray-50">
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
      <!-- Database header -->
      <div class="database-header bg-white border-b border-gray-200 px-8 py-6 sticky top-0 z-20">
        <div class="max-w-[1800px] mx-auto">
          <div class="flex items-center gap-6">
            <button 
              class="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-50 to-slate-100 border border-gray-200 flex items-center justify-center hover:from-primary-50 hover:to-blue-100 hover:border-primary-200 hover:shadow-md transition-all duration-300 group"
              @click="goBack"
            >
              <div class="i-carbon-arrow-left text-xl text-gray-400 group-hover:text-primary-600 transition-colors" />
            </button>
            
            <div class="flex items-center gap-5 flex-1">
              <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-primary-50 to-blue-100 flex items-center justify-center border border-blue-100 shadow-sm">
                <div class="i-carbon-data-base text-3xl text-primary-600" />
              </div>

              <div>
                <h1 class="text-2xl font-bold text-gray-900 leading-tight">
                  {{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}
                </h1>
                <div class="flex items-center gap-3 mt-1.5">
                  <span class="px-2.5 py-0.5 rounded text-xs font-bold bg-gray-100 text-gray-600 border border-gray-200 uppercase tracking-wide">
                    {{ workspaceStore.currentDatabase.type }}
                  </span>
                  <span v-if="workspaceStore.currentDatabase.host" class="text-sm font-medium text-gray-500 flex items-center gap-1">
                    <div class="i-carbon-ibm-cloud-citrix-daas" />
                    {{ workspaceStore.currentDatabase.host }}
                  </span>
                  <div class="w-1 h-1 rounded-full bg-gray-300"></div>
                  <span class="text-sm font-medium text-gray-500">
                    {{ workspaceStore.currentDatabase.tableCount }} tables
                  </span>
                  <template v-if="workspaceStore.hasRichContext">
                    <div class="w-1 h-1 rounded-full bg-gray-300"></div>
                    <span class="text-sm font-bold text-primary-600 flex items-center gap-1">
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
      <div class="tab-navigation bg-gradient-to-r from-slate-50 via-white to-slate-50 border-b border-gray-200 px-8 sticky top-[105px] z-10">
        <div class="max-w-[1800px] mx-auto py-2">
          <div class="flex gap-2">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="tab-btn flex items-center gap-2.5 px-5 py-2.5 rounded-xl font-bold text-sm transition-all duration-300"
              :class="workspaceStore.activeTab === tab.key 
                ? 'bg-gradient-to-r from-primary-500 to-blue-600 text-white shadow-lg shadow-primary-500/30' 
                : 'text-gray-500 hover:text-gray-900 hover:bg-white hover:shadow-md'"
              @click="handleTabChange(tab.key)"
            >
              <div :class="[tab.icon, 'text-lg']" />
              <span>{{ tab.label }}</span>
            </button>
          </div>
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

.tab-btn {
  min-width: 100px;
  justify-content: center;
}
</style>
