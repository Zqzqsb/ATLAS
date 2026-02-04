<script setup lang="ts">
import { computed } from 'vue'
import { useDemoStore } from '@/stores/demo'
import type { ComparisonCase } from '@/types'

const store = useDemoStore()

const categories = [
  { key: 'dirty_data', label: '脏数据处理', icon: 'i-carbon-clean', color: 'orange' },
  { key: 'complex_schema', label: '复杂Schema', icon: 'i-carbon-diagram', color: 'blue' },
  { key: 'business_rule', label: '业务规则', icon: 'i-carbon-rule', color: 'purple' }
] as const

const groupedCases = computed(() => {
  const groups: Record<string, ComparisonCase[]> = {}
  for (const c of store.comparisonCases) {
    const cat = c.category
    if (!groups[cat]) groups[cat] = []
    groups[cat]!.push(c)
  }
  return groups
})

function selectCase(c: ComparisonCase) {
  store.runComparison(c)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Title -->
    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-2 flex items-center gap-2">
        <span class="i-carbon-compare text-blue-500" />
        Rich Context 效果对比
      </h2>
      <p class="text-gray-600 text-sm">
        通过 A/B 对比实验，验证 Rich Context 在不同场景下的实际效果。
        选择一个测试用例，同时运行「有 Context」和「无 Context」两种模式，直观对比差异。
      </p>
    </div>

    <!-- Test Cases by Category -->
    <div class="grid grid-cols-3 gap-4">
      <div 
        v-for="cat in categories" 
        :key="cat.key"
        class="card p-4"
      >
        <h3 class="font-medium mb-3 flex items-center gap-2">
          <span :class="cat.icon" class="text-lg" :style="{ color: `var(--${cat.color}-500)` }" />
          {{ cat.label }}
        </h3>
        <div class="space-y-2">
          <button
            v-for="c in groupedCases[cat.key] || []"
            :key="c.id"
            @click="selectCase(c)"
            class="w-full text-left p-3 rounded-lg border transition-all"
            :class="store.selectedCase?.id === c.id 
              ? 'border-blue-500 bg-blue-50' 
              : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'"
          >
            <div class="font-medium text-sm">{{ c.name }}</div>
            <div class="text-xs text-gray-500 mt-1 truncate">{{ c.question }}</div>
          </button>
        </div>
      </div>
    </div>

    <!-- Comparison Result -->
    <div v-if="store.selectedCase" class="card p-6">
      <h3 class="font-medium mb-4 flex items-center gap-2">
        <span class="i-carbon-task text-blue-500" />
        测试用例: {{ store.selectedCase.name }}
      </h3>
      
      <div class="mb-4 p-4 bg-gray-50 rounded-lg">
        <div class="text-sm text-gray-600 mb-2">自然语言问题:</div>
        <div class="font-medium">{{ store.selectedCase.question }}</div>
        <div class="text-sm text-gray-500 mt-2">{{ store.selectedCase.description }}</div>
      </div>

      <!-- Loading -->
      <div v-if="store.isComparing" class="flex-center py-12">
        <div class="animate-spin i-carbon-loading text-3xl text-blue-500" />
        <span class="ml-3 text-gray-600">正在执行对比测试...</span>
      </div>

      <!-- Results -->
      <div v-else-if="store.comparisonResult" class="grid grid-cols-2 gap-6">
        <!-- Without Context -->
        <div class="border rounded-xl overflow-hidden">
          <div class="bg-red-50 px-4 py-3 border-b flex-between">
            <span class="font-medium text-red-700">❌ 无 Rich Context</span>
            <span class="text-sm text-gray-500">{{ store.comparisonResult.withoutContext.duration }}ms</span>
          </div>
          <div class="p-4">
            <div class="text-sm text-gray-600 mb-2">生成的 SQL:</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-sm overflow-x-auto">{{ store.comparisonResult.withoutContext.sql }}</pre>
            <div class="mt-3 flex items-center gap-2">
              <span 
                class="px-2 py-1 rounded text-sm"
                :class="store.comparisonResult.withoutContext.isCorrect 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-red-100 text-red-700'"
              >
                {{ store.comparisonResult.withoutContext.isCorrect ? '✓ 正确' : '✗ 错误' }}
              </span>
            </div>
          </div>
        </div>

        <!-- With Context -->
        <div class="border rounded-xl overflow-hidden border-green-200">
          <div class="bg-green-50 px-4 py-3 border-b flex-between">
            <span class="font-medium text-green-700">✓ 使用 Rich Context</span>
            <span class="text-sm text-gray-500">{{ store.comparisonResult.withContext.duration }}ms</span>
          </div>
          <div class="p-4">
            <div class="text-sm text-gray-600 mb-2">生成的 SQL:</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-sm overflow-x-auto">{{ store.comparisonResult.withContext.sql }}</pre>
            <div class="mt-3 flex items-center gap-2">
              <span 
                class="px-2 py-1 rounded text-sm"
                :class="store.comparisonResult.withContext.isCorrect 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-red-100 text-red-700'"
              >
                {{ store.comparisonResult.withContext.isCorrect ? '✓ 正确' : '✗ 错误' }}
              </span>
            </div>

            <!-- Used Contexts -->
            <div v-if="store.comparisonResult.withContext.usedContexts.length" class="mt-4">
              <div class="text-sm text-gray-600 mb-2">使用的 Context:</div>
              <div class="space-y-2">
                <div 
                  v-for="ctx in store.comparisonResult.withContext.usedContexts"
                  :key="ctx.id"
                  class="p-2 bg-blue-50 rounded text-sm"
                >
                  <span class="text-blue-600 font-medium">{{ ctx.type }}:</span>
                  {{ ctx.content }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Placeholder -->
      <div v-else class="text-center py-12 text-gray-500">
        点击上方测试用例开始对比
      </div>
    </div>

    <!-- No case selected -->
    <div v-else class="card p-12 text-center text-gray-500">
      <span class="i-carbon-touch-1 text-4xl mb-3 block" />
      请选择一个测试用例开始对比
    </div>
  </div>
</template>
