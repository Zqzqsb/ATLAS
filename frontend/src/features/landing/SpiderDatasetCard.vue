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
    class="spider-card group relative overflow-hidden rounded-lg bg-white border border-gray-200 hover:border-gray-300 transition-colors"
    :class="{ 'opacity-60 grayscale': !isConnected }"
  >
    <!-- Top accent bar -->
    <div class="h-0.5 w-full bg-violet-500" />

    <div class="p-4 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <div class="w-10 h-10 rounded-lg bg-violet-50 flex items-center justify-center">
          <span class="text-xl">🕷️</span>
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-medium text-sm text-gray-800 leading-tight">Spider Dataset</h3>
          <p class="text-xs text-gray-400 mt-0.5">Text-to-SQL Benchmark</p>
        </div>
        <div
          class="w-2 h-2 rounded-full flex-shrink-0"
          :class="isConnected ? 'bg-emerald-500' : 'bg-amber-400'"
        />
      </div>

      <!-- Scenario list -->
      <div class="flex-1 space-y-1.5 mb-3">
        <div
          v-for="s in scenarios"
          :key="s.id"
          class="scenario-row flex items-center gap-2.5 p-2.5 rounded-md border border-transparent cursor-pointer transition-colors"
          :class="s.connected
            ? 'hover:bg-violet-50 hover:border-violet-200'
            : 'opacity-50 cursor-not-allowed'"
          @click="s.connected && enterScenario(s.id)"
        >
          <div class="w-6 h-6 rounded bg-violet-100 text-violet-600 flex items-center justify-center text-xs font-semibold flex-shrink-0">
            {{ s.index }}
          </div>
          <span class="text-base flex-shrink-0">{{ s.icon }}</span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-gray-700">{{ s.name }}</span>
              <span class="text-[10px] text-gray-400">{{ s.tables }} tables</span>
            </div>
            <p class="text-xs text-gray-400 leading-snug mt-0.5 truncate">{{ s.desc }}</p>
          </div>
          <div class="i-lucide-chevron-right text-gray-300 text-sm flex-shrink-0" />
        </div>
      </div>

      <!-- Footer stats -->
      <div class="flex items-center justify-between pt-2.5 border-t border-gray-100">
        <div class="flex items-center gap-3 text-xs text-gray-400">
          <span>{{ totalTables }} tables</span>
          <span class="w-1 h-1 rounded-full bg-gray-300" />
          <span :class="totalContext > 0 ? 'text-violet-600' : ''">{{ totalContext }} context</span>
        </div>
        <span class="px-1.5 py-0.5 text-[10px] font-medium rounded bg-violet-50 text-violet-600 uppercase tracking-wide">
          {{ databases.length }} scenarios
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.spider-card {
  min-height: 220px;
}
</style>
