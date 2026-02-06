<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NTag } from 'naive-ui'
import type { Database } from '@/types'

const props = defineProps<{
  database: Database
}>()

const emit = defineEmits<{
  enter: [id: string]
  test: [id: string]
}>()

const router = useRouter()

const statusColor = computed(() => {
  switch (props.database.status) {
    case 'connected': return 'success'
    case 'disconnected': return 'warning'
    case 'error': return 'error'
    default: return 'default'
  }
})

const statusText = computed(() => {
  switch (props.database.status) {
    case 'connected': return 'Connected'
    case 'disconnected': return 'Disconnected'
    case 'error': return 'Error'
    default: return 'Unknown'
  }
})

const iconBgClass = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'bg-gradient-to-br from-blue-50 to-blue-100 text-blue-600'
    case 'mysql': return 'bg-gradient-to-br from-orange-50 to-orange-100 text-orange-600'
    case 'postgresql': return 'bg-gradient-to-br from-indigo-50 to-indigo-100 text-indigo-600'
    case 'sqlite': return 'bg-gradient-to-br from-gray-50 to-gray-100 text-gray-600'
    default: return 'bg-gradient-to-br from-gray-50 to-gray-100 text-gray-600'
  }
})

const typeIcon = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'i-logos-mariadb-icon'
    case 'mysql': return 'i-logos-mysql'
    case 'postgresql': return 'i-logos-postgresql'
    case 'sqlite': return 'i-simple-icons-sqlite'
    default: return 'i-carbon-data-base'
  }
})

function handleEnter() {
  if (props.database.status === 'connected') {
    router.push(`/workspace/${props.database.id}`)
  }
}
</script>

<template>
  <div 
    class="database-card group relative overflow-hidden rounded-2xl cursor-pointer bg-white/80 backdrop-blur-sm border border-white/60 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:shadow-gray-300/50 hover:-translate-y-1 hover:bg-white/95 transition-all duration-300"
    :class="{ 'opacity-60 grayscale': database.status !== 'connected' }"
    @click="handleEnter"
  >
    <!-- Top accent bar with glow -->
    <div 
      class="h-1 w-full"
      :class="{
        'bg-gradient-to-r from-green-400 via-emerald-400 to-teal-500': database.status === 'connected',
        'bg-gradient-to-r from-yellow-400 via-amber-400 to-orange-500': database.status === 'disconnected',
        'bg-gradient-to-r from-red-400 via-rose-400 to-pink-500': database.status === 'error'
      }"
    />
    
    <!-- Content -->
    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-5">
        <!-- Type icon with enhanced styling -->
        <div 
          class="w-12 h-12 rounded-xl flex items-center justify-center shadow-md"
          :class="iconBgClass"
        >
          <div :class="typeIcon" class="text-2xl" />
        </div>
        
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-800 leading-tight truncate group-hover:text-primary-600 transition-colors">
            {{ database.displayName || database.name }}
          </h3>
          <p class="text-xs text-gray-500 mt-1 flex items-center gap-1.5">
            <span class="px-1.5 py-0.5 rounded bg-gray-100 font-bold text-gray-600">{{ database.type.toUpperCase() }}</span>
            <span v-if="database.host" class="text-gray-400 truncate">{{ database.host }}</span>
          </p>
        </div>

        <!-- Status indicator with pulse -->
        <div 
          class="w-3 h-3 rounded-full flex-shrink-0 ring-4"
          :class="{
            'bg-green-500 ring-green-500/20 animate-pulse': database.status === 'connected',
            'bg-yellow-500 ring-yellow-500/20': database.status === 'disconnected',
            'bg-red-500 ring-red-500/20': database.status === 'error'
          }"
        />
      </div>

      <!-- Stats with better visual hierarchy -->
      <div class="flex gap-4 mb-5 p-3 rounded-xl bg-gradient-to-br from-gray-50 to-slate-100">
        <div class="flex-1 text-center">
          <div class="text-2xl font-extrabold text-gray-800">{{ database.tableCount }}</div>
          <div class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Tables</div>
        </div>
        
        <div class="w-px bg-gradient-to-b from-transparent via-gray-300 to-transparent" />
        
        <div class="flex-1 text-center">
          <div class="text-2xl font-extrabold" :class="database.contextCount > 0 ? 'text-primary-600' : 'text-gray-300'">
            {{ database.contextCount }}
          </div>
          <div class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Context</div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-auto flex items-center justify-between">
        <div v-if="database.tags?.length" class="flex flex-wrap gap-1.5">
          <span 
            v-for="tag in database.tags" 
            :key="tag"
            class="px-2.5 py-1 text-xs font-semibold rounded-lg bg-gray-100 text-gray-600"
          >
            {{ tag }}
          </span>
        </div>
        
        <button 
          v-if="database.status === 'connected'" 
          class="ml-auto px-4 py-2 rounded-lg bg-gradient-to-r from-primary-500 to-blue-600 text-white text-xs font-bold flex items-center gap-1.5 opacity-0 group-hover:opacity-100 shadow-lg shadow-primary-500/30 hover:shadow-xl transition-all duration-300"
        >
          Open <div class="i-carbon-arrow-right" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.database-card {
  min-height: 280px;
}
</style>
