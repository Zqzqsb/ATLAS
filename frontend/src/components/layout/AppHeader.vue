<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NButton, NDropdown } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { useDatabaseStore } from '@/stores/database'

const router = useRouter()
const route = useRoute()
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
  <header class="h-16 bg-white/80 backdrop-blur-md border-b border-gray-200 sticky top-0 z-50 px-6 flex items-center justify-between transition-all">
    <!-- Left: Logo & Navigation -->
    <div class="flex items-center gap-6">
      <!-- Logo -->
      <div 
        class="flex items-center gap-3 cursor-pointer hover:opacity-80 transition-opacity"
        @click="goHome"
      >
        <div class="w-9 h-9 rounded-lg bg-primary-600 shadow-md flex items-center justify-center">
          <span class="text-white font-serif font-bold text-lg">L</span>
        </div>
        <div class="flex flex-col leading-none">
          <span class="font-bold text-xl text-gray-900 tracking-tight">LUCID</span>
          <span class="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Unified Intelligence</span>
        </div>
      </div>

      <!-- Divider -->
      <div class="h-6 w-px bg-gray-200" v-if="isInWorkspace && currentDatabase"></div>

      <!-- Database indicator (when in workspace) -->
      <template v-if="isInWorkspace && currentDatabase">
        <NDropdown 
          :options="databaseOptions" 
          trigger="click"
          @select="handleDatabaseSelect"
        >
          <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-gray-100 cursor-pointer transition-colors border border-transparent hover:border-gray-200">
            <div 
              class="w-2.5 h-2.5 rounded-full shadow-sm"
              :class="currentDatabase.status === 'connected' ? 'bg-green-500' : 'bg-red-500'"
            />
            <span class="font-bold text-gray-700">
              {{ currentDatabase.displayName || currentDatabase.name }}
            </span>
            <div class="i-carbon-chevron-down text-gray-400 text-xs stroke-2" />
          </div>
        </NDropdown>
      </template>
    </div>

    <!-- Right: Actions -->
    <div class="flex items-center gap-4">
      <!-- Demo Link -->
      <NButton 
        v-if="!route.path.startsWith('/demo')"
        secondary
        strong
        size="medium"
        class="!font-bold"
        @click="goToDemo"
      >
        <template #icon>
          <div class="i-carbon-demo" />
        </template>
        Live Demo
      </NButton>
    </div>
  </header>
</template>
