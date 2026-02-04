<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { NTabs, NTab, NSpin, NEmpty, NButton } from 'naive-ui'
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
  <div class="workspace-page min-h-screen bg-gradient-to-br from-gray-900 via-slate-900 to-gray-950">
    <!-- Loading state -->
    <div v-if="workspaceStore.loadingSchema" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center mx-auto mb-4 border border-blue-500/30">
          <div class="i-carbon-data-base text-3xl text-blue-400 animate-pulse" />
        </div>
        <p class="text-gray-400">Loading database schema...</p>
      </div>
    </div>

    <!-- Database not found -->
    <div v-else-if="!workspaceStore.currentDatabase" class="flex items-center justify-center h-screen">
      <div class="text-center">
        <div class="w-16 h-16 rounded-2xl bg-white/5 flex items-center justify-center mx-auto mb-4 border border-white/10">
          <div class="i-carbon-warning text-3xl text-red-400" />
        </div>
        <p class="text-xl text-white mb-2">Database not found</p>
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
      <div class="database-header bg-gradient-to-r from-gray-900/95 to-gray-800/95 backdrop-blur-md border-b border-white/10 px-6 py-5">
        <div class="max-w-[1800px] mx-auto">
          <div class="flex items-center gap-4">
            <NButton quaternary circle size="large" @click="goBack">
              <div class="i-carbon-arrow-left text-xl text-gray-400 hover:text-white transition-colors" />
            </NButton>
            
            <div class="flex items-center gap-4 flex-1">
              <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center border border-blue-500/30 shadow-lg shadow-blue-500/20">
                <div class="i-carbon-data-base text-2xl text-blue-400" />
              </div>

              <div>
                <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-cyan-400 bg-clip-text text-transparent">
                  {{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}
                </h1>
                <div class="flex items-center gap-3 mt-1">
                  <span class="px-2 py-0.5 rounded text-xs bg-white/10 text-gray-400 border border-white/20">
                    {{ workspaceStore.currentDatabase.type.toUpperCase() }}
                  </span>
                  <span v-if="workspaceStore.currentDatabase.host" class="text-sm text-gray-500">
                    {{ workspaceStore.currentDatabase.host }}
                  </span>
                  <span class="text-sm text-gray-400">
                    {{ workspaceStore.currentDatabase.tableCount }} tables
                  </span>
                  <span v-if="workspaceStore.hasRichContext" class="text-sm text-blue-400 flex items-center gap-1">
                    <div class="i-carbon-magic-wand text-sm" />
                    {{ workspaceStore.contexts.length }} contexts
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="tab-navigation bg-gray-900/50 backdrop-blur-md border-b border-white/10 px-6">
        <div class="max-w-[1800px] mx-auto">
          <NTabs
            :value="workspaceStore.activeTab"
            type="line"
            animated
            @update:value="handleTabChange"
          >
            <NTab
              v-for="tab in tabs"
              :key="tab.key"
              :name="tab.key"
            >
              <div class="flex items-center gap-2 px-2 py-1">
                <div :class="[tab.icon, 'text-lg']" />
                <span class="font-medium">{{ tab.label }}</span>
              </div>
            </NTab>
          </NTabs>
        </div>
      </div>

      <!-- Tab content -->
      <div class="workspace-content max-w-[1800px] mx-auto">
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
  min-height: calc(100vh - 180px);
}

.database-header {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.tab-navigation :deep(.n-tabs-nav) {
  background: transparent;
}

.tab-navigation :deep(.n-tabs-tab) {
  color: rgba(255, 255, 255, 0.6);
}

.tab-navigation :deep(.n-tabs-tab:hover) {
  color: rgba(255, 255, 255, 0.9);
}

.tab-navigation :deep(.n-tabs-tab--active) {
  color: white;
}

.tab-navigation :deep(.n-tabs-bar) {
  background: linear-gradient(90deg, #3b82f6, #06b6d4);
}
</style>
