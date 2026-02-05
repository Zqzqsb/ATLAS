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
    case 'mariadb': return 'bg-blue-50 text-blue-600 border-blue-100'
    case 'mysql': return 'bg-orange-50 text-orange-600 border-orange-100'
    case 'postgresql': return 'bg-indigo-50 text-indigo-600 border-indigo-100'
    case 'sqlite': return 'bg-gray-50 text-gray-600 border-gray-100'
    default: return 'bg-gray-50 text-gray-600 border-gray-100'
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
    class="database-card group relative overflow-hidden rounded-xl cursor-pointer bg-white border border-gray-200 shadow-sm hover:shadow-lg hover:border-primary-400 transition-all duration-200"
    :class="{ 'opacity-60 grayscale': database.status !== 'connected' }"
    @click="handleEnter"
  >
    <!-- Top accent bar -->
    <div 
      class="h-1.5 w-full"
      :class="{
        'bg-gradient-to-r from-green-400 to-emerald-500': database.status === 'connected',
        'bg-gradient-to-r from-yellow-400 to-amber-500': database.status === 'disconnected',
        'bg-gradient-to-r from-red-400 to-rose-500': database.status === 'error'
      }"
    />
    
    <!-- Content -->
    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-5">
        <!-- Type icon -->
        <div 
          class="w-11 h-11 rounded-lg flex items-center justify-center border"
          :class="iconBgClass"
        >
          <div :class="typeIcon" class="text-2xl" />
        </div>
        
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-900 leading-tight truncate group-hover:text-primary-600 transition-colors">
            {{ database.displayName || database.name }}
          </h3>
          <p class="text-xs text-gray-500 mt-0.5 flex items-center gap-1.5">
            <span class="font-medium">{{ database.type.toUpperCase() }}</span>
            <span v-if="database.host" class="text-gray-400">· {{ database.host }}</span>
          </p>
        </div>

        <!-- Status indicator -->
        <div 
          class="w-2.5 h-2.5 rounded-full flex-shrink-0"
          :class="{
            'bg-green-500 shadow-sm shadow-green-500/50': database.status === 'connected',
            'bg-yellow-500': database.status === 'disconnected',
            'bg-red-500': database.status === 'error'
          }"
        />
      </div>

      <!-- Stats -->
      <div class="flex gap-4 mb-5">
        <div class="flex-1">
          <div class="text-2xl font-bold text-gray-900">{{ database.tableCount }}</div>
          <div class="text-xs font-medium text-gray-500">Tables</div>
        </div>
        
        <div class="w-px bg-gray-200" />
        
        <div class="flex-1">
          <div class="text-2xl font-bold" :class="database.contextCount > 0 ? 'text-primary-600' : 'text-gray-300'">
            {{ database.contextCount }}
          </div>
          <div class="text-xs font-medium text-gray-500">Context</div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-auto flex items-center justify-between">
        <div v-if="database.tags?.length" class="flex flex-wrap gap-1.5">
          <span 
            v-for="tag in database.tags" 
            :key="tag"
            class="px-2 py-0.5 text-xs font-medium rounded-md bg-gray-100 text-gray-500"
          >
            {{ tag }}
          </span>
        </div>
        
        <button 
          v-if="database.status === 'connected'" 
          class="ml-auto px-3 py-1.5 rounded-md bg-primary-50 text-primary-600 text-xs font-semibold flex items-center gap-1 opacity-0 group-hover:opacity-100 hover:bg-primary-100 transition-all"
        >
          Open <div class="i-carbon-arrow-right" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.database-card {
  height: 246px;
}
</style>
