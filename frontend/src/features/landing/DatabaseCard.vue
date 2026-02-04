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
    case 'connected': return '已连接'
    case 'disconnected': return '未连接'
    case 'error': return '连接错误'
    default: return '未知'
  }
})

// Steam-style gradient based on database type
const gradientClass = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'from-blue-600/20 via-cyan-600/20 to-blue-800/20'
    case 'mysql': return 'from-orange-600/20 via-yellow-600/20 to-orange-800/20'
    case 'postgresql': return 'from-blue-700/20 via-indigo-600/20 to-purple-800/20'
    case 'sqlite': return 'from-gray-600/20 via-slate-600/20 to-gray-800/20'
    default: return 'from-gray-600/20 via-slate-600/20 to-gray-800/20'
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
    class="database-card group relative overflow-hidden rounded-xl cursor-pointer transition-all duration-300"
    :class="{ 'opacity-60': database.status !== 'connected' }"
    @click="handleEnter"
  >
    <!-- Gradient background -->
    <div 
      class="absolute inset-0 bg-gradient-to-br opacity-50 transition-opacity duration-300 group-hover:opacity-70"
      :class="gradientClass"
    />
    
    <!-- Overlay gradient -->
    <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-black/20 to-transparent" />
    
    <!-- Content -->
    <div class="relative h-full p-6 flex flex-col backdrop-blur-[2px]">
      <!-- Header with status -->
      <div class="flex items-start justify-between mb-auto">
        <div class="flex items-center gap-3">
          <!-- Type icon with glow -->
          <div 
            class="w-14 h-14 rounded-xl bg-white/10 backdrop-blur-md flex items-center justify-center border border-white/20 shadow-lg group-hover:shadow-xl group-hover:bg-white/15 transition-all duration-300"
          >
            <div :class="typeIcon" class="text-3xl" />
          </div>
          
          <div>
            <h3 class="font-bold text-xl text-white drop-shadow-lg">
              {{ database.displayName || database.name }}
            </h3>
            <p class="text-sm text-white/80">
              {{ database.type.toUpperCase() }}
              <span v-if="database.host" class="text-white/60">· {{ database.host }}</span>
            </p>
          </div>
        </div>

        <!-- Status badge with glow -->
        <NTag 
          :type="statusColor" 
          size="small" 
          round
          class="shadow-lg"
        >
          <template #icon>
            <div 
              class="w-2 h-2 rounded-full mr-1 animate-pulse"
              :class="{
                'bg-green-400': database.status === 'connected',
                'bg-yellow-400': database.status === 'disconnected',
                'bg-red-400': database.status === 'error'
              }"
            />
          </template>
          {{ statusText }}
        </NTag>
      </div>

      <!-- Stats bar with glass effect -->
      <div class="mt-4 p-4 rounded-lg bg-black/30 backdrop-blur-md border border-white/10">
        <div class="flex items-center justify-around text-white/90">
          <div class="flex flex-col items-center">
            <div class="flex items-center gap-1 mb-1">
              <div class="i-carbon-data-table text-lg" />
            </div>
            <span class="text-2xl font-bold">{{ database.tableCount }}</span>
            <span class="text-xs text-white/60">张表</span>
          </div>
          
          <div class="w-px h-12 bg-white/20" />
          
          <div class="flex flex-col items-center">
            <div class="flex items-center gap-1 mb-1">
              <div 
                class="i-carbon-magic-wand text-lg"
                :class="database.hasRichContext ? 'text-blue-400' : 'text-white/40'"
              />
            </div>
            <span class="text-2xl font-bold">{{ database.contextCount }}</span>
            <span class="text-xs text-white/60">条 Context</span>
          </div>
        </div>
      </div>

      <!-- Tags -->
      <div v-if="database.tags?.length" class="flex flex-wrap gap-2 mt-3">
        <span 
          v-for="tag in database.tags" 
          :key="tag"
          class="px-2 py-1 text-xs rounded bg-white/10 text-white/80 backdrop-blur-sm border border-white/20"
        >
          {{ tag }}
        </span>
      </div>

      <!-- Hover overlay with action button -->
      <div 
        class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-end justify-center pb-8"
      >
        <NButton 
          type="primary" 
          size="large"
          :disabled="database.status !== 'connected'"
          class="shadow-2xl scale-95 group-hover:scale-100 transition-transform duration-300"
          ghost
          @click.stop="handleEnter"
        >
          <template #icon>
            <div class="i-carbon-play text-xl" />
          </template>
          进入工作区
        </NButton>
      </div>
    </div>

    <!-- Hover border glow -->
    <div class="absolute inset-0 rounded-xl border-2 border-transparent group-hover:border-white/30 transition-colors duration-300 pointer-events-none" />
  </div>
</template>

<style scoped>
.database-card {
  height: 320px;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  transform: translateY(0);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.database-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
}

.database-card:active {
  transform: translateY(-4px);
}
</style>
