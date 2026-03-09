<script setup lang="ts">
import { computed } from 'vue'
import { useDemoStore } from '@/stores/demo'
import type { ComparisonCase } from '@/types'

const store = useDemoStore()

const categories = [
  { key: 'dirty_data', label: 'Dirty Data Handling', icon: 'i-lucide-eraser', color: 'orange' },
  { key: 'complex_schema', label: 'Complex Schema', icon: 'i-lucide-git-branch', color: 'blue' },
  { key: 'business_rule', label: 'Business Rules', icon: 'i-lucide-scale', color: 'purple' }
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
        <span class="i-lucide-columns-2 text-blue-500" />
        Rich Context A/B Comparison
      </h2>
      <p class="text-gray-600 text-sm">
        Validate the real-world impact of Rich Context through A/B comparison experiments.
        Select a test case to run both "With Context" and "Without Context" modes side by side.
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
        <span class="i-lucide-clipboard-check text-blue-500" />
        Test Case: {{ store.selectedCase.name }}
      </h3>
      
      <div class="mb-4 p-4 bg-gray-50 rounded-lg">
        <div class="text-sm text-gray-600 mb-2">Natural Language Question:</div>
        <div class="font-medium">{{ store.selectedCase.question }}</div>
        <div class="text-sm text-gray-500 mt-2">{{ store.selectedCase.description }}</div>
      </div>

      <!-- Loading -->
      <div v-if="store.isComparing" class="flex-center py-12">
        <div class="animate-spin i-lucide-loader-2 text-3xl text-blue-500" />
        <span class="ml-3 text-gray-600">Running comparison test...</span>
      </div>

      <!-- Results -->
      <div v-else-if="store.comparisonResult" class="grid grid-cols-2 gap-6">
        <!-- Without Context -->
        <div class="border rounded-xl overflow-hidden">
          <div class="bg-red-50 px-4 py-3 border-b flex-between">
            <span class="font-medium text-red-700">❌ Without Rich Context</span>
            <span class="text-sm text-gray-500">{{ store.comparisonResult.withoutContext.duration }}ms</span>
          </div>
          <div class="p-4">
            <div class="text-sm text-gray-600 mb-2">Generated SQL:</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-sm overflow-x-auto">{{ store.comparisonResult.withoutContext.sql }}</pre>
            <div class="mt-3 flex items-center gap-2">
              <span 
                class="px-2 py-1 rounded text-sm"
                :class="store.comparisonResult.withoutContext.isCorrect 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-red-100 text-red-700'"
              >
                {{ store.comparisonResult.withoutContext.isCorrect ? '✓ Correct' : '✗ Incorrect' }}
              </span>
            </div>
          </div>
        </div>

        <!-- With Context -->
        <div class="border rounded-xl overflow-hidden border-green-200">
          <div class="bg-green-50 px-4 py-3 border-b flex-between">
            <span class="font-medium text-green-700">✓ With Rich Context</span>
            <span class="text-sm text-gray-500">{{ store.comparisonResult.withContext.duration }}ms</span>
          </div>
          <div class="p-4">
            <div class="text-sm text-gray-600 mb-2">Generated SQL:</div>
            <pre class="bg-gray-900 text-green-400 p-3 rounded text-sm overflow-x-auto">{{ store.comparisonResult.withContext.sql }}</pre>
            <div class="mt-3 flex items-center gap-2">
              <span 
                class="px-2 py-1 rounded text-sm"
                :class="store.comparisonResult.withContext.isCorrect 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-red-100 text-red-700'"
              >
                {{ store.comparisonResult.withContext.isCorrect ? '✓ Correct' : '✗ Incorrect' }}
              </span>
            </div>

            <!-- Used Contexts -->
            <div v-if="store.comparisonResult.withContext.usedContexts.length" class="mt-4">
              <div class="text-sm text-gray-600 mb-2">Used Contexts:</div>
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
        Click a test case above to start comparison
      </div>
    </div>

    <!-- No case selected -->
    <div v-else class="card p-12 text-center text-gray-500">
      <span class="i-lucide-pointer text-4xl mb-3 block" />
      Select a test case to begin comparison
    </div>
  </div>
</template>
