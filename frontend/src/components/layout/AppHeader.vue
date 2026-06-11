<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'
import atlasLogo from '@/assets/atlas-logo.svg'

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
  <header class="h-14 bg-white/80 backdrop-blur-md border-b border-gray-200/80 sticky top-0 z-50 px-6 flex items-center justify-between shadow-sm">
    <!-- Left: Logo & Navigation -->
    <div class="flex items-center gap-5">
      <!-- Logo -->
      <div 
        class="flex items-center gap-2.5 cursor-pointer hover:opacity-80 transition-opacity"
        @click="goHome"
      >
<img :src="atlasLogo" alt="ATLAS" class="w-8 h-8 rounded-lg shadow-sm ring-1 ring-gray-900/5" />
        <div class="flex flex-col leading-none mt-0.5">
<span class="font-extrabold text-[17px] text-gray-900 tracking-tight">ATLAS</span>
        </div>
      </div>

      <!-- Divider -->
      <div class="h-5 w-px bg-gray-200/80" v-if="isInWorkspace && currentDatabase"></div>

      <!-- Database indicator (read-only, switch via homepage) -->
      <div v-if="isInWorkspace && currentDatabase" class="flex items-center gap-2 px-2.5 py-1 rounded-md bg-gray-50/50 border border-gray-100">
        <div 
          class="w-2 h-2 rounded-full relative"
          :class="currentDatabase.status === 'connected' ? 'bg-emerald-500' : 'bg-red-500'"
        >
          <div v-if="currentDatabase.status === 'connected'" class="absolute inset-0 rounded-full bg-emerald-500 animate-ping opacity-20"></div>
        </div>
        <span class="font-semibold text-[13px] text-gray-700 tracking-wide">
          {{ currentDatabase.displayName || currentDatabase.name }}
        </span>
      </div>
    </div>

    <!-- Right: Navigation -->
    <div class="flex items-center gap-3">
      <button
        class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-semibold transition-all"
        :class="route.name === 'Features' ? 'bg-primary-50 text-primary-700 ring-1 ring-primary-200/50' : 'text-gray-500 hover:bg-gray-50 hover:text-gray-900 border border-transparent hover:border-gray-200/80'"
        @click="router.push('/features')"
      >
        <div class="i-atlas-sparkles text-sm" />
        Features
      </button>
    </div>
  </header>
</template>
