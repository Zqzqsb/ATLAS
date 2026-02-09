<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NButton } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'

const router = useRouter()
const route = useRoute()
const workspaceStore = useWorkspaceStore()

// Computed
const currentDatabase = computed(() => workspaceStore.currentDatabase)
const isInWorkspace = computed(() => route.name?.toString().startsWith('Workspace'))

// Methods
function goHome() {
  router.push('/')
}

function goToDemo() {
  router.push('/demo')
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

      <!-- Database indicator (read-only, switch via homepage) -->
      <div v-if="isInWorkspace && currentDatabase" class="flex items-center gap-2 px-3 py-1.5">
        <div 
          class="w-2.5 h-2.5 rounded-full shadow-sm"
          :class="currentDatabase.status === 'connected' ? 'bg-green-500' : 'bg-red-500'"
        />
        <span class="font-bold text-gray-700">
          {{ currentDatabase.displayName || currentDatabase.name }}
        </span>
      </div>
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
