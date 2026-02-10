<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { useDatabaseStore } from '@/stores/database'
import type { Database } from '@/types'

const props = defineProps<{
  database: Database
}>()

const emit = defineEmits<{
  enter: [id: string]
  test: [id: string]
}>()

const router = useRouter()
const message = useMessage()
const databaseStore = useDatabaseStore()
const deleting = ref(false)

const iconBgClass = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'bg-blue-50 text-blue-600'
    case 'mysql': return 'bg-orange-50 text-orange-600'
    case 'postgresql': return 'bg-indigo-50 text-indigo-600'
    case 'sqlite': return 'bg-gray-100 text-gray-600'
    default: return 'bg-gray-100 text-gray-600'
  }
})

const typeIcon = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'i-logos-mariadb-icon'
    case 'mysql': return 'i-logos-mysql'
    case 'postgresql': return 'i-logos-postgresql'
    case 'sqlite': return 'i-simple-icons-sqlite'
    default: return 'i-lucide-database'
  }
})

function handleEnter() {
  if (props.database.status === 'connected') {
    router.push(`/workspace/${props.database.id}`)
  }
}

async function handleDelete(e: Event) {
  e.stopPropagation()
  const lakebaseId = props.database.metadata?.lakebaseId
  if (!lakebaseId) return
  deleting.value = true
  try {
    const ok = await databaseStore.deleteDatasource(lakebaseId)
    if (ok) {
      message.success('Datasource deleted')
    } else {
      message.error('Failed to delete datasource')
    }
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div 
    class="database-card group relative overflow-hidden rounded-lg cursor-pointer bg-white border border-gray-200 hover:border-gray-300 transition-colors"
    :class="{ 'opacity-60 grayscale': database.status !== 'connected' }"
    @click="handleEnter"
  >
    <!-- Top accent bar -->
    <div 
      class="h-0.5 w-full"
      :class="{
        'bg-emerald-500': database.status === 'connected',
        'bg-amber-400': database.status === 'disconnected',
        'bg-red-400': database.status === 'error'
      }"
    />
    
    <!-- Content -->
    <div class="p-4 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <!-- Type icon -->
        <div 
          class="w-10 h-10 rounded-lg flex items-center justify-center"
          :class="iconBgClass"
        >
          <div :class="typeIcon" class="text-xl" />
        </div>
        
        <div class="flex-1 min-w-0">
          <h3 class="font-medium text-sm text-gray-800 leading-tight truncate group-hover:text-primary-600 transition-colors">
            {{ database.displayName || database.name }}
          </h3>
          <p class="text-xs text-gray-400 mt-0.5 flex items-center gap-1">
            <span class="px-1.5 py-0.5 rounded bg-gray-100 font-medium text-gray-500 text-[10px]">{{ database.type.toUpperCase() }}</span>
            <span v-if="database.host" class="text-gray-400 truncate">{{ database.host }}</span>
          </p>
        </div>

        <!-- Status indicator -->
        <div 
          class="w-2 h-2 rounded-full flex-shrink-0"
          :class="{
            'bg-emerald-500': database.status === 'connected',
            'bg-amber-400': database.status === 'disconnected',
            'bg-red-400': database.status === 'error'
          }"
        />
      </div>

      <!-- Stats -->
      <div class="flex gap-4 mb-4 p-2.5 rounded-lg bg-gray-50 border border-gray-100">
        <div class="flex-1 text-center">
          <div class="text-xl font-semibold text-gray-800">{{ database.tableCount }}</div>
          <div class="text-[10px] font-medium text-gray-400 uppercase tracking-wide">Tables</div>
        </div>
        
        <div class="w-px bg-gray-200" />
        
        <div class="flex-1 text-center">
          <div class="text-xl font-semibold" :class="database.contextCount > 0 ? 'text-primary-600' : 'text-gray-300'">
            {{ database.contextCount }}
          </div>
          <div class="text-[10px] font-medium text-gray-400 uppercase tracking-wide">Context</div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-auto flex items-center justify-between">
        <div class="flex flex-wrap gap-1">
          <span 
            v-for="tag in database.tags" 
            :key="tag"
            class="px-2 py-0.5 text-xs font-medium rounded bg-gray-100 text-gray-500"
          >
            {{ tag }}
          </span>
        </div>
        
        <div class="ml-auto flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
          <!-- Delete button -->
          <button 
            :disabled="deleting"
            class="p-1.5 rounded-md bg-red-50 text-red-500 hover:bg-red-100 transition-colors"
            @click="handleDelete"
          >
            <div v-if="deleting" class="i-lucide-loader-2 animate-spin text-sm" />
            <div v-else class="i-lucide-trash-2 text-sm" />
          </button>

          <!-- Open button -->
          <button 
            v-if="database.status === 'connected'" 
            class="px-3 py-1.5 rounded-md bg-primary-600 text-white text-xs font-medium flex items-center gap-1 hover:bg-primary-700 transition-colors"
          >
            Open <div class="i-lucide-arrow-right text-xs" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.database-card {
  min-height: 220px;
}
</style>
