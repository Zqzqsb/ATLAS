<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isVisible = ref(false)
const animStep = ref(0)
const containerRef = ref<HTMLElement>()
const activeLayer = ref(-1)

// LUCID unified layers (abstract, no rc_* table names)
const unifiedLayers = [
  {
    name: 'Schema Metadata',
    icon: 'i-lucide-database',
    desc: 'Tables, columns, types, constraints',
    color: 'from-blue-500 to-blue-600',
    bgLight: 'bg-blue-50',
    borderColor: 'border-blue-200',
    delay: 0,
  },
  {
    name: 'Rich Context',
    icon: 'i-lucide-file-text',
    desc: 'Business descriptions, synonyms, rules',
    color: 'from-indigo-500 to-indigo-600',
    bgLight: 'bg-indigo-50',
    borderColor: 'border-indigo-200',
    delay: 150,
  },
  {
    name: 'Relationship Graph',
    icon: 'i-lucide-git-branch',
    desc: 'Foreign keys, join paths, entity links',
    color: 'from-violet-500 to-violet-600',
    bgLight: 'bg-violet-50',
    borderColor: 'border-violet-200',
    delay: 300,
  },
  {
    name: 'Vector Embeddings',
    icon: 'i-lucide-radar',
    desc: 'HNSW index for semantic retrieval',
    color: 'from-purple-500 to-purple-600',
    bgLight: 'bg-purple-50',
    borderColor: 'border-purple-200',
    delay: 450,
  },
  {
    name: 'Change Audit Log',
    icon: 'i-lucide-history',
    desc: 'DDL changes, version tracking',
    color: 'from-fuchsia-500 to-fuchsia-600',
    bgLight: 'bg-fuchsia-50',
    borderColor: 'border-fuchsia-200',
    delay: 600,
  },
]

// Traditional scattered components
const traditionalComponents = [
  {
    name: 'MySQL',
    role: 'Relational DB',
    icon: 'i-lucide-database',
    color: 'from-amber-400 to-orange-500',
    textColor: 'text-orange-700',
    bgColor: 'bg-orange-50',
    x: 15,
    y: 12,
  },
  {
    name: 'Redis',
    role: 'Cache',
    icon: 'i-lucide-zap',
    color: 'from-red-400 to-red-600',
    textColor: 'text-red-700',
    bgColor: 'bg-red-50',
    x: 58,
    y: 8,
  },
  {
    name: 'Milvus',
    role: 'Vector DB',
    icon: 'i-lucide-radar',
    color: 'from-sky-400 to-blue-600',
    textColor: 'text-blue-700',
    bgColor: 'bg-blue-50',
    x: 10,
    y: 52,
  },
  {
    name: 'Elasticsearch',
    role: 'Search Engine',
    icon: 'i-lucide-search',
    color: 'from-yellow-400 to-amber-500',
    textColor: 'text-amber-700',
    bgColor: 'bg-yellow-50',
    x: 52,
    y: 48,
  },
]

const painPoints = [
  { icon: 'i-lucide-refresh-cw', text: 'Data sync overhead across 4 systems' },
  { icon: 'i-lucide-server-crash', text: '4× deployment & maintenance cost' },
  { icon: 'i-lucide-unplug', text: 'Cross-engine inconsistency risk' },
]

let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (entry && entry.isIntersecting && !isVisible.value) {
        isVisible.value = true
        let step = 0
        const interval = setInterval(() => {
          step++
          animStep.value = step
          if (step >= unifiedLayers.length + 4) clearInterval(interval)
        }, 250)
      }
    },
    { threshold: 0.15 }
  )
  if (containerRef.value) observer.observe(containerRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- Section header -->
    <div class="text-center mb-12">
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-blue-50 border border-blue-200 mb-4">
        <div class="i-lucide-cylinder text-blue-600" />
        <span class="text-sm font-semibold text-blue-700">Innovation #1</span>
      </div>
      <h3 class="text-3xl font-bold text-gray-900 mb-3">Lakebase Unified Storage</h3>
      <p class="text-lg text-gray-500 max-w-2xl mx-auto">
        One database to store them all — Schema, Rich Context, Vectors, and Audit Logs coexist inside a single engine
      </p>
    </div>

    <!-- VS Label -->
    <div class="hidden lg:flex items-center justify-center mb-6">
      <div class="flex-1 h-px bg-gradient-to-r from-transparent via-gray-200 to-gray-300" />
      <span class="mx-4 text-xs font-bold text-gray-400 tracking-widest uppercase">Architecture Comparison</span>
      <div class="flex-1 h-px bg-gradient-to-r from-gray-300 via-gray-200 to-transparent" />
    </div>

    <!-- Main visualization -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 items-stretch">

      <!-- ==================== LEFT: LUCID Unified ==================== -->
      <div class="relative flex flex-col">
        <div class="text-center mb-4">
          <span class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-full bg-emerald-50 text-emerald-700 text-sm font-semibold border border-emerald-200 shadow-sm">
            <div class="i-lucide-check-circle text-emerald-500" />
            LUCID — Single Engine
          </span>
        </div>

        <!-- MariaDB Container with glow -->
        <div class="relative flex-1 rounded-2xl border-2 border-blue-200 bg-gradient-to-br from-white to-blue-50/60 p-6 shadow-lg shadow-blue-100/50 overflow-hidden">
          <!-- Subtle animated glow -->
          <div
            class="absolute -top-20 -right-20 w-40 h-40 rounded-full bg-blue-400/10 blur-3xl transition-opacity duration-1000"
            :class="isVisible ? 'opacity-100' : 'opacity-0'"
          />
          <div
            class="absolute -bottom-16 -left-16 w-32 h-32 rounded-full bg-indigo-400/10 blur-3xl transition-opacity duration-1000"
            :class="isVisible ? 'opacity-100' : 'opacity-0'"
            style="transition-delay: 500ms"
          />

          <!-- MariaDB label -->
          <div class="flex items-center gap-3 mb-6 relative z-10">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-600 to-indigo-600 flex-center text-white shadow-md shadow-blue-500/30">
              <div class="i-lucide-database text-lg" />
            </div>
            <div>
              <span class="font-bold text-gray-800">MariaDB</span>
              <div class="flex items-center gap-1.5 mt-0.5">
                <span class="text-xs px-1.5 py-0.5 rounded bg-blue-100 text-blue-600 font-medium">VECTOR</span>
                <span class="text-xs px-1.5 py-0.5 rounded bg-indigo-100 text-indigo-600 font-medium">HNSW</span>
              </div>
            </div>
          </div>

          <!-- Unified layers -->
          <div class="space-y-2 relative z-10">
            <div
              v-for="(layer, idx) in unifiedLayers"
              :key="layer.name"
              class="group relative flex items-center gap-3 px-4 py-3 rounded-xl border bg-white/90 backdrop-blur-sm shadow-sm cursor-default transition-all duration-500 hover:shadow-md"
              :class="[
                isVisible ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-6',
                layer.borderColor,
                activeLayer === idx ? `${layer.bgLight} shadow-md` : ''
              ]"
              :style="{ transitionDelay: `${layer.delay}ms` }"
              @mouseenter="activeLayer = idx"
              @mouseleave="activeLayer = -1"
            >
              <!-- Icon -->
              <div
                class="w-9 h-9 rounded-lg bg-gradient-to-br flex-center text-white text-sm shrink-0 shadow-sm transition-transform duration-200 group-hover:scale-110"
                :class="layer.color"
              >
                <div :class="layer.icon" />
              </div>

              <!-- Text -->
              <div class="flex-1 min-w-0">
                <div class="font-semibold text-sm text-gray-800">{{ layer.name }}</div>
                <div class="text-xs text-gray-500 leading-tight mt-0.5">{{ layer.desc }}</div>
              </div>

              <!-- Active dot -->
              <div
                class="w-2.5 h-2.5 rounded-full bg-emerald-400 ring-4 ring-emerald-100 transition-all duration-300 shrink-0"
                :class="animStep > idx ? 'scale-100 opacity-100' : 'scale-0 opacity-0'"
              />
            </div>
          </div>

          <!-- Connecting flow lines between layers -->
          <div class="flex justify-center mt-3 relative z-10">
            <div
              class="flex items-center gap-1.5 text-xs text-blue-500/70 font-medium transition-all duration-500"
              :class="isVisible ? 'opacity-100' : 'opacity-0'"
              style="transition-delay: 1s"
            >
              <div class="w-1.5 h-1.5 rounded-full bg-blue-400 animate-pulse" />
              All layers colocated — zero sync overhead
            </div>
          </div>
        </div>

        <!-- Benefit badges -->
        <div
          class="flex flex-wrap justify-center gap-2 mt-4 transition-all duration-500"
          :class="isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
          style="transition-delay: 1.2s"
        >
          <span class="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-emerald-50 text-emerald-700 text-xs font-medium border border-emerald-200">
            <div class="i-lucide-zap text-xs" />
            Single deployment
          </span>
          <span class="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-emerald-50 text-emerald-700 text-xs font-medium border border-emerald-200">
            <div class="i-lucide-shield-check text-xs" />
            ACID consistent
          </span>
          <span class="inline-flex items-center gap-1 px-3 py-1 rounded-full bg-emerald-50 text-emerald-700 text-xs font-medium border border-emerald-200">
            <div class="i-lucide-gauge text-xs" />
            Sub-100ms retrieval
          </span>
        </div>
      </div>

      <!-- ==================== RIGHT: Traditional Scattered ==================== -->
      <div class="relative flex flex-col">
        <div class="text-center mb-4">
          <span class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-full bg-red-50 text-red-700 text-sm font-semibold border border-red-200 shadow-sm">
            <div class="i-lucide-x-circle text-red-400" />
            Traditional — Multi-Component
          </span>
        </div>

        <!-- Scattered container -->
        <div class="relative flex-1 rounded-2xl border-2 border-dashed border-gray-300 bg-gradient-to-br from-gray-50 to-red-50/30 p-6 overflow-hidden">

          <!-- Messy connection lines behind everything -->
          <svg
            class="absolute inset-0 w-full h-full pointer-events-none z-0 transition-opacity duration-700"
            :class="isVisible ? 'opacity-100' : 'opacity-0'"
            style="transition-delay: 1.2s"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
          >
            <!-- Chaotic dashed sync lines between components -->
            <line x1="30" y1="25" x2="72" y2="22" stroke="#fca5a5" stroke-width="0.3" stroke-dasharray="1.5,1.5">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="2s" repeatCount="indefinite" />
            </line>
            <line x1="25" y1="30" x2="25" y2="62" stroke="#fca5a5" stroke-width="0.3" stroke-dasharray="1.5,1.5">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="2.3s" repeatCount="indefinite" />
            </line>
            <line x1="72" y1="28" x2="68" y2="58" stroke="#fca5a5" stroke-width="0.3" stroke-dasharray="1.5,1.5">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="1.8s" repeatCount="indefinite" />
            </line>
            <line x1="28" y1="65" x2="65" y2="62" stroke="#fca5a5" stroke-width="0.3" stroke-dasharray="1.5,1.5">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="2.5s" repeatCount="indefinite" />
            </line>
            <line x1="30" y1="28" x2="65" y2="60" stroke="#fca5a5" stroke-width="0.2" stroke-dasharray="1,2">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="3s" repeatCount="indefinite" />
            </line>
            <line x1="70" y1="25" x2="28" y2="60" stroke="#fca5a5" stroke-width="0.2" stroke-dasharray="1,2">
              <animate attributeName="stroke-dashoffset" from="0" to="3" dur="2.7s" repeatCount="indefinite" />
            </line>
          </svg>

          <!-- Component boxes positioned in a scattered grid -->
          <div class="relative z-10 grid grid-cols-2 gap-x-6 gap-y-5">
            <div
              v-for="(comp, idx) in traditionalComponents"
              :key="comp.name"
              class="relative flex flex-col items-center gap-2 p-4 rounded-xl bg-white border-2 border-dashed shadow-sm transition-all duration-500"
              :class="[
                isVisible ? 'opacity-100 scale-100' : 'opacity-0 scale-90',
              ]"
              :style="{ transitionDelay: `${idx * 200 + 600}ms`, borderColor: `var(--border-${idx})` }"
            >
              <!-- Component icon -->
              <div
                class="w-11 h-11 rounded-xl bg-gradient-to-br flex-center text-white shadow-md transition-transform duration-200"
                :class="comp.color"
              >
                <div :class="comp.icon" class="text-lg" />
              </div>
              <div class="text-center">
                <div class="font-bold text-sm" :class="comp.textColor">{{ comp.name }}</div>
                <div class="text-xs text-gray-400 mt-0.5">{{ comp.role }}</div>
              </div>

              <!-- Tiny "sync" indicator -->
              <div class="absolute -top-1.5 -right-1.5 w-3 h-3 rounded-full bg-red-400 border-2 border-white animate-pulse" />
            </div>
          </div>

          <!-- Sync overhead indicator -->
          <div
            class="relative z-10 mt-5 mx-auto max-w-[240px] flex items-center gap-2 px-3 py-2 rounded-lg bg-red-50 border border-red-200/60"
            :class="isVisible ? 'opacity-100' : 'opacity-0'"
            style="transition: opacity 0.5s; transition-delay: 1.6s"
          >
            <div class="i-lucide-activity text-red-400 text-sm shrink-0 animate-pulse" />
            <span class="text-xs text-red-600 font-medium">Constant cross-system sync...</span>
          </div>
        </div>

        <!-- Pain points -->
        <div class="mt-4 space-y-2">
          <div
            v-for="(point, idx) in painPoints"
            :key="idx"
            class="flex items-center gap-2.5 px-3 py-2 rounded-lg bg-red-50/60 border border-red-100 transition-all duration-500"
            :class="isVisible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-4'"
            :style="{ transitionDelay: `${idx * 150 + 1800}ms` }"
          >
            <div class="w-6 h-6 rounded-md bg-red-100 flex-center shrink-0">
              <div :class="point.icon" class="text-red-500 text-xs" />
            </div>
            <span class="text-sm text-red-700/80 font-medium">{{ point.text }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
div[style*="--border-0"] { border-color: #fed7aa; }
div[style*="--border-1"] { border-color: #fecaca; }
div[style*="--border-2"] { border-color: #bfdbfe; }
div[style*="--border-3"] { border-color: #fef08a; }
</style>
