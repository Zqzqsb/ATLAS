<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NButton, NDropdown, NSpace, NSwitch, NTooltip } from 'naive-ui'
import { useAppStore } from '@/stores/app'
import { useWorkspaceStore } from '@/stores/workspace'
import { useDatabaseStore } from '@/stores/database'

const router = useRouter()
const route = useRoute()
const appStore = useAppStore()
const workspaceStore = useWorkspaceStore()
const databaseStore = useDatabaseStore()

// Computed
const currentDatabase = computed(() => workspaceStore.currentDatabase)
const isInWorkspace = computed(() => route.name?.toString().startsWith('Workspace'))

const databaseOptions = computed(() => 
  databaseStore.databases
    .filter(db => db.status === 'connected')
    .map(db => ({
      label: db.displayName || db.name,
      key: db.id
    }))
)

// Methods
function goHome() {
  router.push('/')
}

function goToDemo() {
  router.push('/demo')
}

function handleDatabaseSelect(key: string) {
  router.push(`/workspace/${key}`)
}
</script>

<template>
  <header class="h-14 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 px-4 flex items-center justify-between">
    <!-- Left: Logo & Navigation -->
    <div class="flex items-center gap-4">
      <!-- Logo -->
      <div 
        class="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity"
        @click="goHome"
      >
        <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
          <span class="text-white font-bold text-sm">LC</span>
        </div>
        <span class="font-semibold text-lg text-gray-800 dark:text-gray-100">LUCID</span>
      </div>

      <!-- Database indicator (when in workspace) -->
      <template v-if="isInWorkspace && currentDatabase">
        <div class="flex items-center gap-2 text-sm">
          <span class="text-gray-400">/</span>
          <NDropdown 
            :options="databaseOptions" 
            trigger="click"
            @select="handleDatabaseSelect"
          >
            <div class="flex items-center gap-1.5 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer">
              <div 
                class="w-2 h-2 rounded-full"
                :class="currentDatabase.status === 'connected' ? 'bg-green-500' : 'bg-red-500'"
              />
              <span class="font-medium text-gray-700 dark:text-gray-200">
                {{ currentDatabase.displayName || currentDatabase.name }}
              </span>
              <div class="i-carbon-chevron-down text-gray-400 text-xs" />
            </div>
          </NDropdown>
        </div>
      </template>
    </div>

    <!-- Right: Actions -->
    <div class="flex items-center gap-3">
      <!-- Demo Link -->
      <NButton 
        v-if="!route.path.startsWith('/demo')"
        quaternary 
        size="small"
        @click="goToDemo"
      >
        <template #icon>
          <div class="i-carbon-demo" />
        </template>
        Demo
      </NButton>

      <!-- Theme Toggle -->
      <NTooltip>
        <template #trigger>
          <NButton quaternary circle size="small" @click="appStore.toggleDarkMode">
            <div 
              class="text-lg"
              :class="appStore.isDarkMode ? 'i-carbon-sun' : 'i-carbon-moon'"
            />
          </NButton>
        </template>
        {{ appStore.isDarkMode ? '切换到亮色模式' : '切换到暗色模式' }}
      </NTooltip>

      <!-- Settings -->
      <NTooltip>
        <template #trigger>
          <NButton quaternary circle size="small">
            <div class="i-carbon-settings text-lg" />
          </NButton>
        </template>
        设置
      </NTooltip>
    </div>
  </header>
</template>
