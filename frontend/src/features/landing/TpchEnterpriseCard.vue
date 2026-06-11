<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Database } from '@/types'

const props = defineProps<{
  database: Database
}>()

const router = useRouter()

const isConnected = computed(() => props.database.status === 'connected')

function handleEnter() {
  if (isConnected.value) {
    router.push(`/workspace/${props.database.id}`)
  }
}

const features = [
  { icon: 'i-atlas-layers', label: 'Vector Retrieval', desc: 'HNSW coarse filtering on 500+ tables' },
  { icon: 'i-atlas-brain', label: 'LLM Linking', desc: 'Semantic precision on candidate set' },
  { icon: 'i-atlas-zap', label: 'Ablation Study', desc: '3-mode comparison experiments' },
]
</script>

<template>
  <div
    class="tpch-card group relative overflow-hidden rounded-2xl cursor-pointer bg-white/80 backdrop-blur-sm border border-white/60 shadow-lg shadow-gray-200/40 hover:shadow-xl hover:shadow-amber-200/40 hover:-translate-y-1 hover:bg-white/95 transition-all duration-300"
    :class="{ 'opacity-60 grayscale': !isConnected }"
    @click="handleEnter"
  >
    <!-- Top accent bar -->
    <div class="h-1.5 w-full bg-gradient-to-r from-amber-400 via-orange-500 to-red-500" />

    <div class="p-5 flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-center gap-3 mb-4">
        <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 flex items-center justify-center shadow-lg shadow-amber-500/30">
          <div class="i-atlas-building-2 text-2xl text-white" />
        </div>
        <div class="flex-1 min-w-0">
          <h3 class="font-bold text-base text-gray-800 leading-tight group-hover:text-amber-600 transition-colors">
            TPC-H Enterprise
          </h3>
          <p class="text-xs text-gray-500 mt-0.5">Two-Stage Adaptive Schema Linking</p>
        </div>
        <div
          class="w-3 h-3 rounded-full flex-shrink-0 ring-4"
          :class="isConnected
            ? 'bg-green-500 ring-green-500/20 animate-pulse'
            : 'bg-yellow-500 ring-yellow-500/20'"
        />
      </div>

      <!-- Scale badge -->
      <div class="flex items-center gap-2 mb-4">
        <span class="px-2.5 py-1 text-[10px] font-bold rounded-lg bg-amber-50 text-amber-700 border border-amber-200 uppercase tracking-wide">
          Large-Scale
        </span>
        <span class="text-xs text-gray-400">
          {{ database.tableCount }} tables · {{ database.contextCount }} context entries
        </span>
      </div>

      <!-- Feature list -->
      <div class="flex-1 space-y-2 mb-4">
        <div
          v-for="f in features"
          :key="f.label"
          class="flex items-center gap-3 p-2.5 rounded-xl bg-gradient-to-r from-amber-50/80 to-orange-50/50 border border-amber-100/60"
        >
          <div class="w-7 h-7 rounded-lg bg-white flex items-center justify-center flex-shrink-0 shadow-sm">
            <div :class="f.icon" class="text-sm text-amber-600" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-xs font-bold text-gray-700">{{ f.label }}</div>
            <div class="text-[10px] text-gray-500 leading-snug">{{ f.desc }}</div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between pt-3 border-t border-gray-100">
        <span class="px-2.5 py-1 text-xs font-semibold rounded-lg bg-gray-100 text-gray-600">
          lakebase
        </span>
        <div class="flex items-center gap-1.5 text-xs font-semibold text-amber-600 opacity-0 group-hover:opacity-100 transition-all duration-300">
          Open <div class="i-atlas-arrow-right text-sm" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tpch-card {
  min-height: 280px;
}
</style>
