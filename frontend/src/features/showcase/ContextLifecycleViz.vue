<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'

const isVisible = ref(false)
const activePhase = ref(-1) // -1=idle, 0=onboard, 1=query, 2=evolve
const containerRef = ref<HTMLElement>()
const isAutoPlaying = ref(false)

const phases = [
  {
    id: 'onboarding',
    title: 'Onboarding',
    subtitle: 'Generate Rich Context',
    icon: 'i-lucide-sparkles',
    color: 'blue',
    gradientFrom: 'from-blue-500',
    gradientTo: 'to-cyan-500',
    bgColor: 'bg-blue-50',
    borderColor: 'border-blue-300',
    textColor: 'text-blue-700',
    steps: [
      { icon: 'i-lucide-plug', text: 'Connect database' },
      { icon: 'i-lucide-scan', text: 'Extract schema (tables, columns, types)' },
      { icon: 'i-lucide-brain', text: 'LLM generates descriptions & business rules' },
      { icon: 'i-lucide-radar', text: 'Embed into HNSW vector index' },
    ],
    contextExample: {
      table: 'orders',
      column: 'status',
      type: 'value_mapping',
      content: 'status: 0=Pending, 1=Paid, 2=Shipped, 3=Completed, 4=Cancelled',
    }
  },
  {
    id: 'query',
    title: 'Query-Time',
    subtitle: 'Leverage Rich Context',
    icon: 'i-lucide-search',
    color: 'emerald',
    gradientFrom: 'from-emerald-500',
    gradientTo: 'to-teal-500',
    bgColor: 'bg-emerald-50',
    borderColor: 'border-emerald-300',
    textColor: 'text-emerald-700',
    steps: [
      { icon: 'i-lucide-message-square', text: 'User asks natural language question' },
      { icon: 'i-lucide-radar', text: 'Vector search retrieves relevant contexts' },
      { icon: 'i-lucide-syringe', text: 'Inject into LLM prompt as knowledge' },
      { icon: 'i-lucide-code', text: 'Generate accurate SQL with context awareness' },
    ],
    contextExample: {
      table: 'orders',
      column: 'status',
      type: 'value_mapping',
      content: '✅ Used: "active orders" → status IN (1, 2)',
    }
  },
  {
    id: 'evolution',
    title: 'Evolution',
    subtitle: 'Self-Maintaining Update',
    icon: 'i-lucide-refresh-cw',
    color: 'violet',
    gradientFrom: 'from-violet-500',
    gradientTo: 'to-purple-500',
    bgColor: 'bg-violet-50',
    borderColor: 'border-violet-300',
    textColor: 'text-violet-700',
    steps: [
      { icon: 'i-lucide-alert-triangle', text: 'DDL change detected (ALTER TABLE)' },
      { icon: 'i-lucide-clock', text: 'Mark affected contexts as stale' },
      { icon: 'i-lucide-wand-2', text: 'LLM regenerates updated descriptions' },
      { icon: 'i-lucide-check-circle', text: 'Re-embed & verify correctness' },
    ],
    contextExample: {
      table: 'orders',
      column: 'status',
      type: 'value_mapping',
      content: '🔄 Updated: status now includes 5=Refunded (new value)',
    }
  },
]

let autoPlayTimer: number | null = null

function startAutoPlay() {
  if (isAutoPlaying.value) return
  isAutoPlaying.value = true
  let phase = 0
  activePhase.value = phase

  autoPlayTimer = window.setInterval(() => {
    phase = (phase + 1) % 3
    activePhase.value = phase
  }, 4000)
}

function selectPhase(idx: number) {
  if (autoPlayTimer) {
    clearInterval(autoPlayTimer)
    autoPlayTimer = null
    isAutoPlaying.value = false
  }
  activePhase.value = idx
}

let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (entry && entry.isIntersecting && !isVisible.value) {
        isVisible.value = true
        setTimeout(startAutoPlay, 800)
      }
    },
    { threshold: 0.2 }
  )
  if (containerRef.value) observer.observe(containerRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
  if (autoPlayTimer) clearInterval(autoPlayTimer)
})

const currentPhase = computed(() => activePhase.value >= 0 ? phases[activePhase.value] : null)
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- Section header -->
    <div class="text-center mb-12">
      <div class="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-emerald-50 border border-emerald-200 mb-4">
        <div class="i-lucide-rotate-cw text-emerald-600" />
        <span class="text-sm font-semibold text-emerald-700">Innovation #3</span>
      </div>
      <h3 class="text-3xl font-bold text-gray-900 mb-3">Rich Context Lifecycle</h3>
      <p class="text-lg text-gray-500 max-w-2xl mx-auto">
        Context is not static annotation — it has a complete lifecycle of generation, usage, and evolution
      </p>
    </div>

    <!-- Lifecycle ring + detail panel -->
    <div class="max-w-5xl mx-auto grid grid-cols-1 lg:grid-cols-5 gap-8 items-start">
      <!-- Left: Lifecycle ring (3 phases as connected cards) -->
      <div class="lg:col-span-2 flex flex-col items-center gap-4">
        <div
          v-for="(phase, idx) in phases"
          :key="phase.id"
          class="w-full"
        >
          <!-- Phase card -->
          <button
            class="w-full flex items-center gap-4 px-5 py-4 rounded-2xl border-2 transition-all duration-400 text-left group"
            :class="[
              activePhase === idx
                ? `${phase.borderColor} ${phase.bgColor} shadow-lg`
                : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-md',
            ]"
            @click="selectPhase(idx)"
          >
            <div
              class="w-12 h-12 rounded-xl bg-gradient-to-br flex-center text-white shrink-0 transition-transform duration-300"
              :class="[phase.gradientFrom, phase.gradientTo, activePhase === idx ? 'scale-110' : 'group-hover:scale-105']"
            >
              <div :class="phase.icon" class="text-xl" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="font-bold text-gray-800">{{ phase.title }}</div>
              <div class="text-sm text-gray-500">{{ phase.subtitle }}</div>
            </div>
            <div
              v-if="activePhase === idx"
              class="w-2 h-8 rounded-full bg-gradient-to-b"
              :class="[phase.gradientFrom, phase.gradientTo]"
            />
          </button>

          <!-- Connector arrow -->
          <div v-if="idx < phases.length - 1" class="flex justify-center py-1">
            <div class="i-lucide-chevron-down text-gray-300" />
          </div>
        </div>

        <!-- Loop-back arrow -->
        <div class="flex items-center gap-2 text-gray-400">
          <div class="i-lucide-corner-left-up" />
          <span class="text-xs font-medium">Continuous loop</span>
        </div>
      </div>

      <!-- Right: Detail panel -->
      <div class="lg:col-span-3">
        <Transition name="phase-fade" mode="out-in">
          <div
            v-if="currentPhase"
            :key="currentPhase.id"
            class="rounded-2xl border-2 p-6"
            :class="[currentPhase.borderColor, currentPhase.bgColor]"
          >
            <!-- Phase title -->
            <div class="flex items-center gap-3 mb-6">
              <div
                class="w-10 h-10 rounded-xl bg-gradient-to-br flex-center text-white"
                :class="[currentPhase.gradientFrom, currentPhase.gradientTo]"
              >
                <div :class="currentPhase.icon" class="text-lg" />
              </div>
              <div>
                <h4 class="font-bold text-lg text-gray-900">{{ currentPhase.title }}</h4>
                <p class="text-sm text-gray-500">{{ currentPhase.subtitle }}</p>
              </div>
            </div>

            <!-- Step-by-step flow -->
            <div class="space-y-3 mb-6">
              <div
                v-for="(step, idx) in currentPhase.steps"
                :key="idx"
                class="flex items-center gap-3 px-4 py-2.5 rounded-xl bg-white/70 border border-white/90 shadow-sm transition-all duration-500"
                :class="isVisible ? 'opacity-100 translate-x-0' : 'opacity-0 -translate-x-4'"
                :style="{ transitionDelay: `${idx * 150}ms` }"
              >
                <div class="w-6 h-6 rounded-lg bg-white flex-center shadow-sm shrink-0">
                  <div :class="step.icon" class="text-xs" :style="{ color: `var(--un-color-${currentPhase.color}-600, #6366f1)` }" />
                </div>
                <span class="text-sm text-gray-700">{{ step.text }}</span>
                <div v-if="idx < currentPhase.steps.length - 1" class="ml-auto i-lucide-arrow-right text-xs text-gray-300" />
                <div v-else class="ml-auto i-lucide-check text-xs text-emerald-500" />
              </div>
            </div>

            <!-- Context example card -->
            <div class="rounded-xl bg-white/80 border border-white/90 p-4 shadow-sm">
              <div class="flex items-center gap-2 mb-2">
                <div class="i-lucide-file-text text-gray-400 text-sm" />
                <span class="text-xs font-semibold text-gray-500 uppercase tracking-wide">Context Example</span>
              </div>
              <div class="font-mono text-sm space-y-1">
                <div><span class="text-gray-400">table:</span> <span class="text-gray-700">{{ currentPhase.contextExample.table }}</span></div>
                <div><span class="text-gray-400">column:</span> <span class="text-gray-700">{{ currentPhase.contextExample.column }}</span></div>
                <div><span class="text-gray-400">type:</span> <span :class="currentPhase.textColor">{{ currentPhase.contextExample.type }}</span></div>
                <div class="pt-1 border-t border-gray-100">
                  <span class="text-gray-400">content:</span>
                  <span class="text-gray-800 font-medium"> {{ currentPhase.contextExample.content }}</span>
                </div>
              </div>
            </div>
          </div>
        </Transition>

        <!-- Idle state -->
        <div
          v-if="!currentPhase"
          class="rounded-2xl border-2 border-dashed border-gray-300 bg-gray-50/50 p-12 flex-center min-h-[400px]"
        >
          <div class="text-center text-gray-400">
            <div class="i-lucide-mouse-pointer text-3xl mb-3" />
            <p>Select a phase to explore</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.phase-fade-enter-active,
.phase-fade-leave-active {
  transition: all 0.3s ease;
}
.phase-fade-enter-from {
  opacity: 0;
  transform: translateY(12px);
}
.phase-fade-leave-to {
  opacity: 0;
  transform: translateY(-12px);
}
</style>
