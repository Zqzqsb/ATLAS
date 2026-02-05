<script setup lang="ts">
type ScenarioKey = 'benchmark' | 'comparison' | 'selfmaintain'

const scenarios: { key: ScenarioKey; label: string; icon: string; badge: string }[] = [
  { key: 'comparison', label: 'Context 效果对比', icon: 'i-carbon-compare', badge: '核心' },
  { key: 'selfmaintain', label: '自维持演示', icon: 'i-carbon-recycle', badge: 'NEW' },
  { key: 'benchmark', label: '基准测试', icon: 'i-carbon-chart-line', badge: '' }
]

const model = defineModel<ScenarioKey>()
</script>

<template>
  <div class="inline-flex p-1 bg-gray-100 rounded-lg">
    <button
      v-for="s in scenarios"
      :key="s.key"
      @click="model = s.key"
      class="px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 flex items-center gap-2"
      :class="model === s.key 
        ? 'bg-white text-gray-900 shadow-sm' 
        : 'text-gray-600 hover:text-gray-900'"
    >
      <span :class="s.icon" class="text-base" />
      <span>{{ s.label }}</span>
      <span 
        v-if="s.badge"
        class="px-1.5 py-0.5 text-xs font-medium rounded"
        :class="s.badge === '核心' ? 'bg-blue-100 text-blue-600' : 'bg-emerald-100 text-emerald-600'"
      >
        {{ s.badge }}
      </span>
    </button>
  </div>
</template>
