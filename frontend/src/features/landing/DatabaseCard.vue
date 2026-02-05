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
    class="database-card group relative overflow-hidden rounded-lg cursor-pointer bg-white border border-gray-200 shadow-sm hover:shadow-md hover:border-primary-300 transition-all duration-200"
    :class="{ 'opacity-75 grayscale': database.status !== 'connected' }"
    @click="handleEnter"
  >
    <!-- Content -->
    <div class="p-5 flex flex-col h-full">
      <!-- Header with status -->
      <div class="flex items-start justify-between mb-4">
        <div class="flex items-center gap-3">
          <!-- Type icon -->
          <div 
            class="w-12 h-12 rounded-lg flex items-center justify-center border"
            :class="iconBgClass"
          >
            <div :class="typeIcon" class="text-2xl" />
          </div>
          
          <div>
            <h3 class="font-bold text-base text-gray-900 leading-tight group-hover:text-primary-600 transition-colors">
              {{ database.displayName || database.name }}
            </h3>
            <p class="text-xs text-gray-500 mt-0.5">
              {{ database.type.toUpperCase() }}
              <span v-if="database.host" class="text-gray-400 ml-1">· {{ database.host }}</span>
            </p>
          </div>
        </div>

        <!-- Status badge -->
        <NTag 
          :type="statusColor" 
          size="small" 
          :bordered="false"
          class="font-medium"
        >
          <template #icon>
            <div 
              class="w-1.5 h-1.5 rounded-full mr-1"
              :class="{
                'bg-green-500 animate-pulse': database.status === 'connected',
                'bg-yellow-500': database.status === 'disconnected',
                'bg-red-500': database.status === 'error'
              }"
            />
          </template>
          {{ statusText }}
        </NTag>
      </div>

      <!-- Stats bar -->
      <div class="grid grid-cols-2 gap-3 mb-4">
        <div class="p-3 rounded-lg bg-gray-50 border border-gray-100 text-center">
          <div class="text-xs font-medium text-gray-500 mb-1">Tables</div>
          <div class="text-xl font-bold text-gray-900">{{ database.tableCount }}</div>
        </div>
        
        <div class="p-3 rounded-lg bg-gray-50 border border-gray-100 text-center">
          <div class="text-xs font-medium text-gray-500 mb-1">Context</div>
          <div class="text-xl font-bold" :class="database.contextCount > 0 ? 'text-primary-600' : 'text-gray-400'">
            {{ database.contextCount }}
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-auto flex items-center justify-between">
        <div v-if="database.tags?.length" class="flex flex-wrap gap-1.5">
          <span 
            v-for="tag in database.tags" 
            :key="tag"
            class="px-2 py-0.5 text-xs font-medium rounded bg-gray-100 text-gray-600"
          >
            {{ tag }}
          </span>
        </div>
        
        <div v-if="database.status === 'connected'" class="ml-auto text-primary-600 text-sm font-medium flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
          Open <div class="i-carbon-arrow-right" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.database-card {
  height: 240px;
}
</style>
