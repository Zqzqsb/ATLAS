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
  <div class="workspace-page min-h-screen bg-gray-50 dark:bg-gray-950">
    <!-- Loading state -->
    <div v-if="workspaceStore.loadingSchema" class="flex items-center justify-center h-screen">
      <NSpin size="large" description="加载数据库信息..." />
    </div>

    <!-- Database not found -->
    <div v-else-if="!workspaceStore.currentDatabase" class="flex items-center justify-center h-screen">
      <NEmpty description="数据库不存在或未连接">
        <template #extra>
          <NButton type="primary" @click="goBack">
            返回首页
          </NButton>
        </template>
      </NEmpty>
    </div>

    <!-- Workspace content -->
    <template v-else>
      <!-- Database header -->
      <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 px-6 py-4">
        <div class="flex items-center gap-4">
          <NButton quaternary circle @click="goBack">
            <div class="i-carbon-arrow-left" />
          </NButton>
          
          <div>
            <h1 class="text-xl font-semibold text-gray-800 dark:text-gray-100">
              {{ workspaceStore.currentDatabase.displayName || workspaceStore.currentDatabase.name }}
            </h1>
            <p class="text-sm text-gray-500">
              {{ workspaceStore.currentDatabase.type.toUpperCase() }}
              <span v-if="workspaceStore.currentDatabase.host">
                · {{ workspaceStore.currentDatabase.host }}
              </span>
              · {{ workspaceStore.currentDatabase.tableCount }} 张表
              <span v-if="workspaceStore.hasRichContext" class="text-blue-500">
                · {{ workspaceStore.contexts.length }} 条 Context
              </span>
            </p>
          </div>
        </div>
      </div>

      <!-- Tab navigation -->
      <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 px-6">
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
            <div class="flex items-center gap-2">
              <div :class="tab.icon" />
              <span>{{ tab.label }}</span>
            </div>
          </NTab>
        </NTabs>
      </div>

      <!-- Tab content -->
      <div class="workspace-content">
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
  min-height: calc(100vh - 140px);
}
</style>
