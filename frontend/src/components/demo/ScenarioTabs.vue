<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'

type ScenarioKey = 'benchmark' | 'comparison' | 'selfmaintain'

const scenarios: { key: ScenarioKey; label: string; icon: string; badge: string }[] = [
  { key: 'comparison', label: 'Context Comparison', icon: 'i-lucide-columns-2', badge: 'Core' },
  { key: 'selfmaintain', label: 'Self-Maintenance', icon: 'i-lucide-refresh-cw', badge: 'NEW' },
  { key: 'benchmark', label: 'Benchmark', icon: 'i-lucide-trending-up', badge: '' }
]

const model = defineModel<ScenarioKey>()

// Sliding indicator
const tabRefs = ref<HTMLElement[]>([])
const indicatorStyle = ref({ left: '0px', width: '0px' })

const activeIndex = computed(() => 
  scenarios.findIndex(s => s.key === model.value)
)

function updateIndicator() {
  const activeEl = tabRefs.value[activeIndex.value]
  if (activeEl) {
    indicatorStyle.value = {
      left: `${activeEl.offsetLeft}px`,
      width: `${activeEl.offsetWidth}px`
    }
  }
}

watch(activeIndex, () => nextTick(updateIndicator))
onMounted(() => nextTick(updateIndicator))
</script>

<template>
  <div class="relative inline-flex p-1 bg-gray-100 rounded-lg">
    <!-- Sliding indicator -->
    <div 
      class="absolute top-1 bottom-1 bg-white rounded-md shadow-sm transition-all duration-300 ease-out"
      :style="indicatorStyle"
    />
    <!-- Tab buttons -->
    <button
      v-for="(s, index) in scenarios"
      :key="s.key"
      :ref="(el) => { if (el) tabRefs[index] = el as HTMLElement }"
      @click="model = s.key"
      class="relative z-10 px-4 py-2 text-sm font-medium transition-colors duration-200 flex items-center gap-2"
      :class="model === s.key 
        ? 'text-gray-900' 
        : 'text-gray-500 hover:text-gray-700'"
    >
      <span :class="s.icon" class="text-base" />
      <span>{{ s.label }}</span>
      <span 
        v-if="s.badge"
        class="px-1.5 py-0.5 text-xs font-medium rounded"
        :class="s.badge === 'Core' ? 'bg-blue-100 text-blue-600' : 'bg-emerald-100 text-emerald-600'"
      >
        {{ s.badge }}
      </span>
    </button>
  </div>
</template>
