<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isVisible = ref(false)
const currentStage = ref(0) // 0=idle, 1=embedding, 2=hnsw, 3=llm, 4=done
const containerRef = ref<HTMLElement>()

// Simulated tables
const allTables = [
  'users', 'orders', 'products', 'categories', 'payments',
  'shipping', 'reviews', 'coupons', 'inventory', 'suppliers',
  'warehouses', 'returns', 'cart_items', 'wishlists', 'addresses',
  'notifications', 'sessions', 'logs', 'configs', 'migrations',
  'analytics', 'campaigns', 'tags', 'comments', 'media',
  'permissions', 'roles', 'audit_log', 'api_keys', 'webhooks',
]

const stage1Results = ref<string[]>([])
const stage2Results = ref<string[]>([])
const queryText = ref('Find total order amount for VIP users last month')

let animTimer: number | null = null
let observer: IntersectionObserver | null = null

function startAnimation() {
  currentStage.value = 1

  // Stage 1: Vector search (fast filter)
  setTimeout(() => {
    currentStage.value = 2
    stage1Results.value = ['users', 'orders', 'payments', 'coupons', 'categories', 'products', 'cart_items', 'shipping']
  }, 1200)

  // Stage 2: LLM refine
  setTimeout(() => {
    currentStage.value = 3
  }, 2800)

  setTimeout(() => {
    currentStage.value = 4
    stage2Results.value = ['users', 'orders', 'payments']
  }, 4200)
}

function isInStage1(table: string) {
  return stage1Results.value.includes(table)
}

function isInStage2(table: string) {
  return stage2Results.value.includes(table)
}

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (entry && entry.isIntersecting && !isVisible.value) {
        isVisible.value = true
        setTimeout(startAnimation, 600)
      }
    },
    { threshold: 0.3 }
  )
  if (containerRef.value) observer.observe(containerRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
  if (animTimer) clearInterval(animTimer)
})
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- Section header -->
    <div class="text-center mb-12">
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-violet-50 border border-violet-200 mb-4">
        <div class="i-lucide-filter text-violet-600" />
        <span class="text-sm font-semibold text-violet-700">Innovation #2</span>
      </div>
      <h3 class="text-3xl font-bold text-gray-900 mb-3">Two-Stage Adaptive Schema Linking</h3>
      <p class="text-lg text-gray-500 max-w-2xl mx-auto">
        HNSW vector recall narrows the search space, then LLM precisely selects the relevant tables
      </p>
    </div>

    <!-- Query input display -->
    <div
      class="max-w-2xl mx-auto mb-10 transition-all duration-700"
      :class="isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
    >
      <div class="flex items-center gap-3 px-5 py-3.5 rounded-2xl bg-white border-2 border-violet-200 shadow-lg shadow-violet-100/50">
        <div class="i-lucide-message-square text-violet-500 text-lg" />
        <span class="text-gray-700 font-medium">{{ queryText }}</span>
        <div
          v-if="currentStage >= 1"
          class="ml-auto flex items-center gap-1.5 text-xs font-semibold"
          :class="currentStage >= 4 ? 'text-emerald-600' : 'text-violet-600'"
        >
          <div class="w-1.5 h-1.5 rounded-full animate-pulse" :class="currentStage >= 4 ? 'bg-emerald-500' : 'bg-violet-500'" />
          {{ currentStage >= 4 ? 'Linked' : 'Linking...' }}
        </div>
      </div>
    </div>

    <!-- Two-stage funnel visualization -->
    <div class="max-w-5xl mx-auto">
      <!-- Stage labels -->
      <div class="grid grid-cols-3 gap-6 mb-6">
        <!-- All tables -->
        <div class="text-center">
          <div
            class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-semibold transition-all duration-300"
            :class="currentStage >= 1 ? 'bg-gray-100 text-gray-700' : 'bg-gray-50 text-gray-400'"
          >
            <div class="i-lucide-layout-grid" />
            All Tables ({{ allTables.length }})
          </div>
        </div>

        <!-- Stage 1 result -->
        <div class="text-center">
          <div
            class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-semibold transition-all duration-500"
            :class="currentStage >= 2 ? 'bg-violet-100 text-violet-700' : 'bg-gray-50 text-gray-400'"
          >
            <div class="i-lucide-radar" />
            Stage 1: HNSW ({{ stage1Results.length || '?' }})
          </div>
        </div>

        <!-- Stage 2 result -->
        <div class="text-center">
          <div
            class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-semibold transition-all duration-500"
            :class="currentStage >= 4 ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-50 text-gray-400'"
          >
            <div class="i-lucide-brain" />
            Stage 2: LLM ({{ stage2Results.length || '?' }})
          </div>
        </div>
      </div>

      <!-- Funnel body -->
      <div class="grid grid-cols-3 gap-6 items-start">
        <!-- Column 1: All tables grid -->
        <div class="rounded-2xl border-2 border-gray-200 bg-white/60 p-4 min-h-[320px]">
          <div class="grid grid-cols-3 gap-1.5">
            <div
              v-for="(table, idx) in allTables"
              :key="table"
              class="px-1.5 py-1 rounded-md text-center font-mono transition-all duration-500"
              :class="[
                currentStage >= 2 && isInStage1(table)
                  ? 'bg-violet-100 text-violet-700 border border-violet-300 scale-105'
                  : currentStage >= 2
                    ? 'bg-gray-50 text-gray-300 border border-gray-100 scale-95 opacity-40'
                    : 'bg-gray-100 text-gray-600 border border-gray-200',
              ]"
              :style="{ fontSize: '10px', transitionDelay: `${idx * 30}ms` }"
            >
              {{ table }}
            </div>
          </div>
        </div>

        <!-- Column 2: Stage 1 results (vector candidates) -->
        <div class="relative">
          <!-- Arrow from col1 -->
          <div
            class="absolute -left-6 top-1/2 -translate-y-1/2 transition-all duration-500"
            :class="currentStage >= 2 ? 'opacity-100' : 'opacity-0'"
          >
            <div class="i-lucide-chevrons-right text-violet-400 text-xl" />
          </div>

          <div
            class="rounded-2xl border-2 p-4 min-h-[320px] flex flex-col transition-all duration-500"
            :class="currentStage >= 2 ? 'border-violet-300 bg-violet-50/50' : 'border-gray-200 bg-white/60'"
          >
            <!-- Vector search animation -->
            <div v-if="currentStage === 1" class="flex-1 flex-center">
              <div class="text-center">
                <div class="i-lucide-radar text-3xl text-violet-400 animate-spin mb-3" style="animation-duration: 2s" />
                <p class="text-sm text-violet-600 font-medium">HNSW Vector Search...</p>
                <p class="text-xs text-violet-400 mt-1">&lt; 100ms</p>
              </div>
            </div>

            <!-- Stage 1 results -->
            <div v-else-if="currentStage >= 2" class="space-y-2">
              <div
                v-for="(table, idx) in stage1Results"
                :key="table"
                class="flex items-center gap-2.5 px-3 py-2 rounded-xl transition-all duration-400"
                :class="[
                  currentStage >= 4 && isInStage2(table)
                    ? 'bg-emerald-100 border border-emerald-300'
                    : currentStage >= 3 && !isInStage2(table)
                      ? 'bg-white/50 border border-gray-200 opacity-40'
                      : 'bg-white border border-violet-200',
                ]"
                :style="{ transitionDelay: `${idx * 80}ms` }"
              >
                <div class="w-6 h-6 rounded flex-center text-xs" :class="[
                  currentStage >= 4 && isInStage2(table)
                    ? 'bg-emerald-500 text-white'
                    : 'bg-violet-100 text-violet-600'
                ]">
                  <div :class="currentStage >= 4 && isInStage2(table) ? 'i-lucide-check' : 'i-lucide-table'" />
                </div>
                <span class="font-mono text-sm" :class="currentStage >= 4 && isInStage2(table) ? 'text-emerald-700 font-bold' : 'text-gray-700'">{{ table }}</span>
                <div v-if="currentStage >= 2 && currentStage < 3" class="ml-auto">
                  <span class="text-xs text-violet-500 font-medium">{{ (0.95 - idx * 0.06).toFixed(2) }}</span>
                </div>
              </div>
            </div>

            <!-- Idle state -->
            <div v-else class="flex-1 flex-center text-gray-300 text-sm">
              Waiting...
            </div>
          </div>
        </div>

        <!-- Column 3: Stage 2 results (LLM refined) -->
        <div class="relative">
          <!-- Arrow from col2 -->
          <div
            class="absolute -left-6 top-1/2 -translate-y-1/2 transition-all duration-500"
            :class="currentStage >= 4 ? 'opacity-100' : 'opacity-0'"
          >
            <div class="i-lucide-chevrons-right text-emerald-400 text-xl" />
          </div>

          <div
            class="rounded-2xl border-2 p-4 min-h-[320px] flex flex-col transition-all duration-500"
            :class="currentStage >= 4 ? 'border-emerald-300 bg-emerald-50/50' : 'border-gray-200 bg-white/60'"
          >
            <!-- LLM analysis animation -->
            <div v-if="currentStage === 3" class="flex-1 flex-center">
              <div class="text-center">
                <div class="i-lucide-brain text-3xl text-violet-500 animate-pulse mb-3" />
                <p class="text-sm text-violet-600 font-medium">LLM Semantic Ranking...</p>
                <p class="text-xs text-violet-400 mt-1">Deep understanding</p>
              </div>
            </div>

            <!-- Final results -->
            <div v-else-if="currentStage >= 4" class="space-y-3">
              <div
                v-for="(table, idx) in stage2Results"
                :key="table"
                class="px-4 py-3 rounded-xl bg-white border-2 border-emerald-300 shadow-sm transition-all duration-500"
                :class="isVisible ? 'opacity-100 translate-x-0' : 'opacity-0 translate-x-4'"
                :style="{ transitionDelay: `${idx * 150}ms` }"
              >
                <div class="flex items-center gap-2.5 mb-1.5">
                  <div class="w-6 h-6 rounded bg-emerald-500 text-white flex-center text-xs">
                    <div class="i-lucide-check" />
                  </div>
                  <span class="font-mono font-bold text-emerald-800">{{ table }}</span>
                </div>
                <p class="text-xs text-gray-500 pl-8.5">
                  {{ table === 'users' ? 'VIP user attributes & membership' : table === 'orders' ? 'Order records with timestamps' : 'Transaction amounts & methods' }}
                </p>
              </div>

              <!-- Summary metrics -->
              <div class="mt-4 pt-3 border-t border-emerald-200/50 grid grid-cols-2 gap-3">
                <div class="text-center">
                  <div class="text-lg font-bold text-emerald-700">30 → 3</div>
                  <div class="text-xs text-gray-500">Tables Filtered</div>
                </div>
                <div class="text-center">
                  <div class="text-lg font-bold text-emerald-700">&lt; 2s</div>
                  <div class="text-xs text-gray-500">Total Latency</div>
                </div>
              </div>
            </div>

            <!-- Idle state -->
            <div v-else class="flex-1 flex-center text-gray-300 text-sm">
              Waiting...
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
