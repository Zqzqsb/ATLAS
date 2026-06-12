<script setup lang="ts">
import { computed } from 'vue'
import VizShell from '../shared/VizShell.vue'
import { useArchAnim } from '../../composables/useArchAnim'

const props = defineProps<{ variant: string }>()

const meta = computed(() => ({
  'store-schema': { accent: 'bg-slate-500', icon: 'i-lucide-table-2', label: 'Schema' },
  'store-ctx': { accent: 'bg-indigo-500', icon: 'i-lucide-book-open', label: 'Context' },
  'store-emb': { accent: 'bg-cyan-500', icon: 'i-lucide-radar', label: 'Vectors' },
  'store-log': { accent: 'bg-amber-500', icon: 'i-lucide-history', label: 'Audit' },
  'store-hnsw': { accent: 'bg-cyan-600', icon: 'i-lucide-network', label: 'HNSW' },
}[props.variant] ?? { accent: 'bg-cyan-500', icon: 'i-lucide-cylinder', label: 'Store' }))

const ctxTypes = ['i-lucide-file-text', 'i-lucide-tag', 'i-lucide-scale', 'i-lucide-languages', 'i-lucide-calculator', 'i-lucide-link']
const graphNodes: [number, number][] = [[50, 30], [20, 70], [80, 70], [50, 90], [35, 50], [65, 50]]
const graphEdgePairs: [number, number][] = [[0, 1], [0, 2], [1, 3], [2, 3], [0, 4], [0, 5], [4, 5], [1, 4], [2, 5]]
const graphPoint = (index: number): [number, number] => graphNodes[index]!
const graphEdges = graphEdgePairs.map(([a, b]) => {
  const [x1, y1] = graphPoint(a)
  const [x2, y2] = graphPoint(b)
  return { x1, y1, x2, y2 }
})

const { step } = useArchAnim(6, 800)
</script>

<template>
  <VizShell :accent="meta.accent" :icon="meta.icon" :label="meta.label">
    <!-- Schema stack -->
    <div v-if="variant === 'store-schema'" class="h-full flex items-end justify-center gap-3 pb-4">
      <div
        v-for="(icon, i) in ['i-lucide-table-2', 'i-lucide-columns-3', 'i-lucide-git-branch']"
        :key="i"
        class="w-20 rounded-t-xl flex items-center justify-center transition-all duration-500"
        :style="{ height: `${80 + i * 36}px` }"
        :class="step >= i ? 'bg-slate-400/40 ring-1 ring-slate-300/50' : 'bg-white/5 opacity-30'"
      >
        <div :class="[icon, 'text-2xl text-slate-200']" />
      </div>
    </div>

    <!-- Context types orbit -->
    <div v-else-if="variant === 'store-ctx'" class="h-full flex items-center justify-center">
      <div class="relative w-64 h-64">
        <div
          v-for="(icon, i) in ctxTypes"
          :key="i"
          class="absolute w-10 h-10 -ml-5 -mt-5 rounded-lg flex items-center justify-center transition-all duration-400"
          :style="{
            left: `${50 + 40 * Math.cos((i / ctxTypes.length) * Math.PI * 2 - Math.PI / 2)}%`,
            top: `${50 + 40 * Math.sin((i / ctxTypes.length) * Math.PI * 2 - Math.PI / 2)}%`,
          }"
          :class="step >= i ? 'bg-indigo-500/60 ring-1 ring-indigo-300 scale-110' : 'bg-white/5 opacity-30'"
        >
          <div :class="[icon, 'text-sm text-white']" />
        </div>
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="i-lucide-book-open text-4xl text-indigo-300 animate-pulse" />
        </div>
      </div>
    </div>

    <!-- HNSW graph -->
    <div v-else-if="variant === 'store-hnsw' || variant === 'store-emb'" class="h-full flex items-center justify-center">
      <svg viewBox="0 0 100 100" class="w-72 h-72">
        <line
          v-for="(edge, i) in graphEdges"
          :key="`e${i}`"
          :x1="edge.x1" :y1="edge.y1" :x2="edge.x2" :y2="edge.y2"
          stroke="rgba(34,211,238,0.4)" stroke-width="0.8"
          :class="step >= i % 6 ? 'opacity-100' : 'opacity-15'" class="transition-opacity"
        />
        <circle
          v-for="(p, i) in graphNodes"
          :key="`n${i}`"
          :cx="p[0]" :cy="p[1]" r="4"
          :fill="step >= i ? '#22d3ee' : 'rgba(255,255,255,0.2)'"
          class="transition-all duration-400"
        />
        <circle v-if="variant === 'store-emb'" cx="50" cy="50" r="18" fill="none" stroke="#22d3ee" stroke-width="0.5" opacity="0.4">
          <animate attributeName="r" values="12;22;12" dur="2s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.6;0.1;0.6" dur="2s" repeatCount="indefinite" />
        </circle>
      </svg>
    </div>

    <!-- Audit timeline -->
    <div v-else class="h-full flex items-center px-6">
      <div class="relative w-full h-2 bg-white/10 rounded-full">
        <div
          v-for="i in 6"
          :key="i"
          class="absolute top-1/2 -translate-y-1/2 w-4 h-4 rounded-full border-2 border-amber-400 transition-all duration-400"
          :style="{ left: `${(i - 1) * 18 + 5}%` }"
          :class="step >= i - 1 ? 'bg-amber-400 scale-110 shadow-lg shadow-amber-500/40' : 'bg-transparent scale-75 opacity-30'"
        />
        <div
          class="absolute top-1/2 -translate-y-1/2 h-0.5 bg-amber-400/60 rounded-full transition-all duration-700"
          :style="{ width: `${Math.min(step, 5) * 18 + 5}%` }"
        />
      </div>
    </div>
  </VizShell>
</template>
