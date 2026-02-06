<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Database } from '@/types'

const props = defineProps<{
  databases: Database[]  // Spider 子库列表
}>()

const router = useRouter()

// 计算汇总信息
const totalTables = computed(() => 
  props.databases.reduce((sum, db) => sum + (db.tableCount || 0), 0)
)

const totalContext = computed(() => 
  props.databases.reduce((sum, db) => sum + (db.contextCount || 0), 0)
)

const isConnected = computed(() => 
  props.databases.some(db => db.status === 'connected')
)

// 场景列表
const scenarios = computed(() => 
  props.databases.map(db => ({
    id: db.id,
    name: getScenarioName(db.name),
    icon: getScenarioIcon(db.name),
    tag: getScenarioTag(db.name)
  }))
)

function getScenarioName(dbName: string): string {
  if (dbName.includes('tvshow') || dbName.includes('tv_show')) return 'TV Show'
  if (dbName.includes('flight')) return 'Flight'
  if (dbName.includes('wta')) return 'WTA Tennis'
  return dbName
}

function getScenarioIcon(dbName: string): string {
  if (dbName.includes('tvshow') || dbName.includes('tv_show')) return '📺'
  if (dbName.includes('flight')) return '✈️'
  if (dbName.includes('wta')) return '🎾'
  return '📊'
}

function getScenarioTag(_dbName: string): string {
  // 所有 Spider 库都是脏库场景，不做区分
  return ''
}

function handleEnter() {
  // 进入第一个已连接的 Spider 库
  const connectedDb = props.databases.find(db => db.status === 'connected')
  if (connectedDb) {
    router.push(`/workspace/${connectedDb.id}`)
  }
}
</script>

<template>
  <div 
    class="spider-card group relative overflow-hidden rounded-2xl cursor-pointer bg-white/80 backdrop-blur-sm border border-white/60 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:shadow-gray-300/50 hover:-translate-y-1 hover:bg-white/95 transition-all duration-300"
    :class="{ 'opacity-60 grayscale': !isConnected }"
    @click="handleEnter"
  >
    <!-- Top accent bar - Spider themed gradient -->
    <div class="h-1.5 w-full bg-gradient-to-r from-violet-500 via-purple-500 to-indigo-500" />
    
    <!-- Content -->
    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <!-- Spider icon -->
        <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-lg shadow-violet-500/30">
          <span class="text-2xl">🕷️</span>
        </div>
        
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-800 leading-tight group-hover:text-violet-600 transition-colors">
            Spider Dataset
          </h3>
          <p class="text-xs text-gray-500 mt-1">
            Text-to-SQL Benchmark
          </p>
        </div>

        <!-- Status indicator -->
        <div 
          class="w-3 h-3 rounded-full flex-shrink-0 ring-4"
          :class="isConnected 
            ? 'bg-green-500 ring-green-500/20 animate-pulse' 
            : 'bg-yellow-500 ring-yellow-500/20'"
        />
      </div>

      <!-- Scenarios -->
      <div class="mb-4">
        <div class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">Databases</div>
        <div class="flex flex-wrap gap-2">
          <div 
            v-for="scenario in scenarios" 
            :key="scenario.id"
            class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-gray-50/80 hover:bg-violet-50 transition-colors"
          >
            <span class="text-sm">{{ scenario.icon }}</span>
            <span class="text-xs font-medium text-gray-700">{{ scenario.name }}</span>
          </div>
        </div>
      </div>

      <!-- Stats -->
      <div class="flex gap-4 mb-4 p-3 rounded-xl bg-gradient-to-br from-violet-50 to-purple-50">
        <div class="flex-1 text-center">
          <div class="text-2xl font-extrabold text-gray-800">{{ totalTables }}</div>
          <div class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Tables</div>
        </div>
        
        <div class="w-px bg-gradient-to-b from-transparent via-violet-200 to-transparent" />
        
        <div class="flex-1 text-center">
          <div class="text-2xl font-extrabold" :class="totalContext > 0 ? 'text-violet-600' : 'text-gray-300'">
            {{ totalContext }}
          </div>
          <div class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Context</div>
        </div>
      </div>

      <!-- Footer -->
      <div class="mt-auto flex items-center justify-between">
        <span class="px-2.5 py-1 text-xs font-semibold rounded-lg bg-violet-100 text-violet-700">
          {{ databases.length }} databases
        </span>
        
        <button 
          v-if="isConnected" 
          class="px-4 py-2 rounded-lg bg-gradient-to-r from-violet-500 to-purple-600 text-white text-xs font-bold flex items-center gap-1.5 opacity-0 group-hover:opacity-100 shadow-lg shadow-violet-500/30 hover:shadow-xl transition-all duration-300"
        >
          Explore <div class="i-carbon-arrow-right" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.spider-card {
  min-height: 280px;
}
</style>
