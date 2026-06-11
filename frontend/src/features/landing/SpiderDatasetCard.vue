<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Database } from '@/types'

const props = defineProps<{
  databases: Database[]
}>()

const router = useRouter()

const totalTables = computed(() =>
  props.databases.reduce((sum, db) => sum + (db.tableCount || 0), 0)
)

const totalContext = computed(() =>
  props.databases.reduce((sum, db) => sum + (db.contextCount || 0), 0)
)

const isConnected = computed(() =>
  props.databases.some(db => db.status === 'connected')
)

// Scenario definitions with descriptions
const SCENARIO_META: Record<string, { name: string; icon: string; desc: string }> = {
  tvshow:  { name: 'TV Show',    icon: '📺', desc: 'TV channels, series ratings and cartoon metadata' },
  flight:  { name: 'Flight',     icon: '✈️', desc: 'Airlines, airports and flight route networks' },
  wta:     { name: 'WTA Tennis', icon: '🎾', desc: 'Players, matches, rankings and tournament data' },
}

function matchScenarioKey(dbName: string): string {
  const n = dbName.toLowerCase()
  if (n.includes('tvshow') || n.includes('tv_show')) return 'tvshow'
  if (n.includes('flight')) return 'flight'
  if (n.includes('wta')) return 'wta'
  return 'unknown'
}

const scenarios = computed(() =>
  props.databases.map((db, idx) => {
    const key = matchScenarioKey(db.name)
    const meta = SCENARIO_META[key] || { name: db.name, icon: '📊', desc: '' }
    return {
      id: db.id,
      index: idx + 1,
      ...meta,
      tables: db.tableCount || 0,
      connected: db.status === 'connected',
    }
  })
)

function enterScenario(scenarioId: string) {
  router.push(`/workspace/${scenarioId}`)
}
</script>

<template>
  <div
    class="spider-card group relative overflow-hidden rounded-2xl bg-white/80 backdrop-blur-sm border border-white/60 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:shadow-gray-300/50 transition-all duration-300"
    :class="{ 'opacity-60 grayscale': !isConnected }"
  >
    <!-- Top accent bar -->
    <div class="h-1.5 w-full bg-gradient-to-r from-violet-500 via-purple-500 to-indigo-500" />

    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-5">
        <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-lg shadow-violet-500/30">
          <span class="text-2xl">🕷️</span>
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-800 leading-tight">Spider Dataset</h3>
          <p class="text-xs text-gray-500 mt-0.5">Text-to-SQL Benchmark</p>
        </div>
        <div
          class="w-3 h-3 rounded-full flex-shrink-0 ring-4"
          :class="isConnected
            ? 'bg-green-500 ring-green-500/20 animate-pulse'
            : 'bg-yellow-500 ring-yellow-500/20'"
        />
      </div>

      <!-- Scenario list -->
      <div class="flex-1 space-y-2 mb-4">
        <div
          v-for="s in scenarios"
          :key="s.id"
          class="scenario-row flex items-center gap-3 p-3 rounded-xl border border-transparent cursor-pointer transition-all duration-200"
          :class="s.connected
            ? 'hover:bg-violet-50 hover:border-violet-200 active:scale-[0.99]'
            : 'opacity-50 cursor-not-allowed'"
          @click="s.connected && enterScenario(s.id)"
        >
          <!-- Number badge -->
          <div class="w-7 h-7 rounded-lg bg-violet-100 text-violet-600 flex items-center justify-center text-xs font-extrabold flex-shrink-0">
            {{ s.index }}
          </div>
          <!-- Icon -->
          <span class="text-lg flex-shrink-0">{{ s.icon }}</span>
          <!-- Info -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-bold text-gray-800">{{ s.name }}</span>
              <span class="text-[10px] font-semibold text-gray-400">{{ s.tables }} tables</span>
            </div>
            <p class="text-xs text-gray-500 leading-snug mt-0.5 truncate">{{ s.desc }}</p>
          </div>
          <!-- Arrow -->
          <div class="i-lucide-chevron-right text-gray-300 text-sm flex-shrink-0 group-hover/row:text-violet-400 transition-colors" />
        </div>
      </div>

      <!-- Footer stats -->
      <div class="flex items-center justify-between pt-3 border-t border-gray-100">
        <div class="flex items-center gap-4 text-xs font-semibold text-gray-500">
          <span>{{ totalTables }} tables</span>
          <span class="w-1 h-1 rounded-full bg-gray-300" />
          <span :class="totalContext > 0 ? 'text-violet-600' : ''">{{ totalContext }} context</span>
        </div>
        <span class="px-2 py-0.5 text-[10px] font-bold rounded-md bg-violet-100 text-violet-600 uppercase tracking-wide">
          {{ databases.length }} scenarios
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.spider-card {
  min-height: 280px;
}
</style>
