<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NButton, NTag, NGrid, NGridItem, NSpin, NCode, NCollapse, NCollapseItem } from 'naive-ui'
import { useDemoStore } from '@/stores/demo'
import type { ComparisonCategory } from '@/types'

const demoStore = useDemoStore()

const categoryLabels: Record<ComparisonCategory, { label: string; color: string }> = {
  dirty_data: { label: '脏数据处理', color: 'warning' },
  complex_schema: { label: '复杂 Schema', color: 'info' },
  business_rule: { label: '业务规则', color: 'success' }
}

const groupedCases = computed(() => {
  const groups: Record<ComparisonCategory, typeof demoStore.comparisonCases> = {
    dirty_data: [],
    complex_schema: [],
    business_rule: []
  }
  
  for (const c of demoStore.comparisonCases) {
    groups[c.category].push(c)
  }
  
  return groups
})

function runCase(caseItem: any) {
  demoStore.runComparison(caseItem)
}
</script>

<template>
  <div class="context-comparison">
    <!-- Intro -->
    <div class="text-center mb-8">
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">
        Rich Context 对比演示
      </h2>
      <p class="text-gray-500 max-w-2xl mx-auto">
        对比有无 Rich Context 时的 SQL 生成效果，展示 Context 对查询准确性的提升
      </p>
    </div>

    <!-- Category tabs -->
    <div class="space-y-6">
      <div v-for="(cases, category) in groupedCases" :key="category">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-3 flex items-center gap-2">
          <NTag :type="categoryLabels[category].color as any" size="small">
            {{ categoryLabels[category].label }}
          </NTag>
        </h3>

        <NGrid :x-gap="16" :y-gap="16" :cols="2">
          <NGridItem v-for="c in cases" :key="c.id">
            <NCard 
              :title="c.name" 
              size="small"
              hoverable
              :class="{ 'border-blue-500': demoStore.selectedCase?.id === c.id }"
            >
              <template #header-extra>
                <NTag v-if="c.difficulty" size="tiny" :bordered="false">
                  {{ c.difficulty }}
                </NTag>
              </template>

              <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">
                {{ c.description }}
              </p>

              <div class="bg-gray-50 dark:bg-gray-800 rounded p-3 mb-3">
                <p class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  "{{ c.question }}"
                </p>
              </div>

              <NButton 
                type="primary" 
                size="small"
                :loading="demoStore.isComparing && demoStore.selectedCase?.id === c.id"
                @click="runCase(c)"
              >
                运行对比
              </NButton>
            </NCard>
          </NGridItem>
        </NGrid>
      </div>
    </div>

    <!-- Comparison result -->
    <div v-if="demoStore.comparisonResult" class="mt-8">
      <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-4">
        对比结果
      </h3>

      <NGrid :x-gap="24" :cols="2">
        <!-- Without Context -->
        <NGridItem>
          <NCard 
            title="无 Rich Context" 
            :class="{ 'border-red-500': !demoStore.comparisonResult.withoutContext.isCorrect }"
          >
            <template #header-extra>
              <NTag 
                :type="demoStore.comparisonResult.withoutContext.isCorrect ? 'success' : 'error'" 
                size="small"
              >
                {{ demoStore.comparisonResult.withoutContext.isCorrect ? '正确' : '错误' }}
              </NTag>
            </template>

            <div class="mb-4">
              <p class="text-sm text-gray-500 mb-2">生成的 SQL:</p>
              <div class="bg-gray-900 rounded overflow-hidden">
                <NCode 
                  :code="demoStore.comparisonResult.withoutContext.sql" 
                  language="sql" 
                  class="p-3 text-sm"
                />
              </div>
            </div>

            <div v-if="demoStore.comparisonResult.withoutContext.errorReason" class="bg-red-50 dark:bg-red-900/20 rounded p-3">
              <p class="text-sm text-red-600 dark:text-red-400">
                <span class="font-medium">错误原因：</span>
                {{ demoStore.comparisonResult.withoutContext.errorReason }}
              </p>
            </div>

            <div class="mt-3 text-sm text-gray-500">
              耗时: {{ demoStore.comparisonResult.withoutContext.duration }}ms
            </div>
          </NCard>
        </NGridItem>

        <!-- With Context -->
        <NGridItem>
          <NCard 
            title="有 Rich Context" 
            :class="{ 'border-green-500': demoStore.comparisonResult.withContext.isCorrect }"
          >
            <template #header-extra>
              <NTag 
                :type="demoStore.comparisonResult.withContext.isCorrect ? 'success' : 'error'" 
                size="small"
              >
                {{ demoStore.comparisonResult.withContext.isCorrect ? '正确' : '错误' }}
              </NTag>
            </template>

            <div class="mb-4">
              <p class="text-sm text-gray-500 mb-2">生成的 SQL:</p>
              <div class="bg-gray-900 rounded overflow-hidden">
                <NCode 
                  :code="demoStore.comparisonResult.withContext.sql" 
                  language="sql" 
                  class="p-3 text-sm"
                />
              </div>
            </div>

            <!-- Used contexts -->
            <NCollapse v-if="demoStore.comparisonResult.withContext.usedContexts?.length" arrow-placement="right">
              <NCollapseItem title="使用的 Context" name="contexts">
                <div class="space-y-2">
                  <div 
                    v-for="ctx in demoStore.comparisonResult.withContext.usedContexts" 
                    :key="ctx.id"
                    class="p-2 bg-blue-50 dark:bg-blue-900/20 rounded text-sm"
                  >
                    <NTag size="tiny" type="info" class="mr-2">{{ ctx.type }}</NTag>
                    {{ ctx.content }}
                  </div>
                </div>
              </NCollapseItem>
            </NCollapse>

            <div v-if="demoStore.comparisonResult.withContext.explanation" class="mt-3 bg-green-50 dark:bg-green-900/20 rounded p-3">
              <p class="text-sm text-green-600 dark:text-green-400">
                {{ demoStore.comparisonResult.withContext.explanation }}
              </p>
            </div>

            <div class="mt-3 text-sm text-gray-500">
              耗时: {{ demoStore.comparisonResult.withContext.duration }}ms
            </div>
          </NCard>
        </NGridItem>
      </NGrid>
    </div>

    <!-- Loading overlay -->
    <div v-if="demoStore.isComparing" class="fixed inset-0 bg-black/20 flex items-center justify-center z-50">
      <NCard class="w-64">
        <div class="flex flex-col items-center">
          <NSpin size="large" class="mb-4" />
          <p class="text-gray-600 dark:text-gray-400">正在对比...</p>
        </div>
      </NCard>
    </div>
  </div>
</template>
