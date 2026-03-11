<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'
import lucidLogo from '@/assets/lucid-logo.svg'

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
</script>

<template>
  <header class="h-14 bg-white border-b border-gray-200 sticky top-0 z-50 px-5 flex items-center justify-between">
    <!-- Left: Logo & Navigation -->
    <div class="flex items-center gap-5">
      <!-- Logo -->
      <div 
        class="flex items-center gap-2.5 cursor-pointer hover:opacity-80 transition-opacity"
        @click="goHome"
      >
        <img :src="lucidLogo" alt="LUCID" class="w-8 h-8 rounded-lg" />
        <div class="flex flex-col leading-none">
          <span class="font-semibold text-lg text-gray-900 tracking-tight">LUCID</span>
        </div>
      </div>

      <!-- Divider -->
      <div class="h-5 w-px bg-gray-200" v-if="isInWorkspace && currentDatabase"></div>

      <!-- Database indicator (read-only, switch via homepage) -->
      <div v-if="isInWorkspace && currentDatabase" class="flex items-center gap-2 px-2.5 py-1">
        <div 
          class="w-2 h-2 rounded-full"
          :class="currentDatabase.status === 'connected' ? 'bg-emerald-500' : 'bg-red-500'"
        />
        <span class="font-medium text-sm text-gray-700">
          {{ currentDatabase.displayName || currentDatabase.name }}
        </span>
      </div>
    </div>

    <!-- Right: Navigation -->
    <div class="flex items-center gap-3">
      <button
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-all"
        :class="route.name === 'Features' ? 'bg-primary-50 text-primary-700' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'"
        @click="router.push('/features')"
      >
        <div class="i-lucide-sparkles text-sm" />
        Features
      </button>
    </div>
  </header>
</template>
