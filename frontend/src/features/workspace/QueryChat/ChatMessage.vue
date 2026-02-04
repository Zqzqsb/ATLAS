<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NTag, NCollapse, NCollapseItem, NCode, NScrollbar, NSpin } from 'naive-ui'
import type { ReActStep, RichContext, GroundingResult } from '@/types'

const props = defineProps<{
  type: 'user' | 'assistant'
  content?: string
  question?: string
  sql?: string
  reactSteps?: ReActStep[]
  usedContexts?: RichContext[]
  groundingResult?: GroundingResult | null
  groundingStage?: 'idle' | 'stage1' | 'stage2' | 'done'
  loading?: boolean
  error?: string | null
  duration?: number
}>()

const stepTypeLabel: Record<string, { label: string; color: string; icon: string }> = {
  thought: { label: '思考', color: 'info', icon: 'i-carbon-idea' },
  action: { label: '行动', color: 'warning', icon: 'i-carbon-play' },
  observation: { label: '观察', color: 'success', icon: 'i-carbon-view' },
  answer: { label: '回答', color: 'primary', icon: 'i-carbon-checkmark' },
  error: { label: '错误', color: 'error', icon: 'i-carbon-warning' }
}

const groundingStageText = computed(() => {
  switch (props.groundingStage) {
    case 'stage1': return '表级链接中...'
    case 'stage2': return '列级链接中...'
    case 'done': return '链接完成'
    default: return ''
  }
})
</script>

<template>
  <!-- User message -->
  <div v-if="type === 'user'" class="flex justify-end mb-4">
    <div class="max-w-[80%] bg-blue-500 text-white rounded-2xl rounded-tr-sm px-4 py-2">
      <p class="whitespace-pre-wrap">{{ question || content }}</p>
    </div>
  </div>

  <!-- Assistant message -->
  <div v-else class="mb-6">
    <div class="max-w-full">
      <!-- Loading state -->
      <div v-if="loading && !sql" class="flex items-center gap-3 text-gray-500 mb-4">
        <NSpin size="small" />
        <span>{{ groundingStageText || '正在分析...' }}</span>
      </div>

      <!-- Error state -->
      <div v-if="error" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-4">
        <div class="flex items-center gap-2 text-red-600 dark:text-red-400">
          <div class="i-carbon-warning" />
          <span>{{ error }}</span>
        </div>
      </div>

      <!-- Grounding result -->
      <NCollapse v-if="groundingResult" arrow-placement="right" class="mb-4">
        <NCollapseItem title="Schema Grounding" name="grounding">
          <template #header-extra>
            <NTag type="success" size="small">
              {{ groundingResult.tables.length }} 表 · {{ groundingResult.columns.length }} 列
            </NTag>
          </template>
          
          <div class="space-y-2 text-sm">
            <div v-if="groundingResult.tables.length">
              <span class="text-gray-500">相关表：</span>
              <NTag 
                v-for="t in groundingResult.tables" 
                :key="t.name"
                size="small"
                class="mr-1"
              >
                {{ t.name }}
                <span class="text-gray-400 ml-1">{{ (t.confidence * 100).toFixed(0) }}%</span>
              </NTag>
            </div>
            <div v-if="groundingResult.columns.length">
              <span class="text-gray-500">相关列：</span>
              <NTag 
                v-for="c in groundingResult.columns.slice(0, 10)" 
                :key="`${c.table}.${c.column}`"
                size="small"
                class="mr-1"
              >
                {{ c.table }}.{{ c.column }}
              </NTag>
              <span v-if="groundingResult.columns.length > 10" class="text-gray-400">
                +{{ groundingResult.columns.length - 10 }} more
              </span>
            </div>
          </div>
        </NCollapseItem>
      </NCollapse>

      <!-- ReAct steps -->
      <NCollapse v-if="reactSteps && reactSteps.length" arrow-placement="right" class="mb-4">
        <NCollapseItem title="ReAct 推理过程" name="react">
          <template #header-extra>
            <NTag type="info" size="small">
              {{ reactSteps.length }} 步
            </NTag>
          </template>
          
          <div class="space-y-3">
            <div 
              v-for="step in reactSteps" 
              :key="step.step"
              class="flex gap-3"
            >
              <div class="flex-shrink-0">
                <div 
                  class="w-6 h-6 rounded-full flex items-center justify-center text-sm"
                  :class="{
                    'bg-blue-100 text-blue-600': step.type === 'thought',
                    'bg-yellow-100 text-yellow-600': step.type === 'action',
                    'bg-green-100 text-green-600': step.type === 'observation',
                    'bg-purple-100 text-purple-600': step.type === 'answer',
                    'bg-red-100 text-red-600': step.type === 'error'
                  }"
                >
                  <div :class="stepTypeLabel[step.type]?.icon" />
                </div>
              </div>
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <span class="text-xs font-medium text-gray-500">
                    {{ stepTypeLabel[step.type]?.label }}
                  </span>
                  <span class="text-xs text-gray-400">Step {{ step.step }}</span>
                </div>
                <p class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                  {{ step.content }}
                </p>
              </div>
            </div>
          </div>
        </NCollapseItem>
      </NCollapse>

      <!-- Used contexts -->
      <NCollapse v-if="usedContexts && usedContexts.length" arrow-placement="right" class="mb-4">
        <NCollapseItem title="使用的 Rich Context" name="contexts">
          <template #header-extra>
            <NTag type="primary" size="small">
              {{ usedContexts.length }} 条
            </NTag>
          </template>
          
          <div class="space-y-2">
            <div 
              v-for="ctx in usedContexts" 
              :key="ctx.id"
              class="p-2 bg-gray-50 dark:bg-gray-800 rounded text-sm"
            >
              <div class="flex items-center gap-2 mb-1">
                <NTag size="tiny" :bordered="false">{{ ctx.tableName }}</NTag>
                <NTag v-if="ctx.columnName" size="tiny" type="info" :bordered="false">
                  {{ ctx.columnName }}
                </NTag>
                <NTag size="tiny" type="success" :bordered="false">{{ ctx.type }}</NTag>
              </div>
              <p class="text-gray-600 dark:text-gray-400">{{ ctx.content }}</p>
            </div>
          </div>
        </NCollapseItem>
      </NCollapse>

      <!-- Generated SQL -->
      <div v-if="sql" class="bg-gray-900 rounded-lg overflow-hidden">
        <div class="flex items-center justify-between px-4 py-2 bg-gray-800">
          <span class="text-sm text-gray-400">生成的 SQL</span>
          <div class="flex items-center gap-2">
            <span v-if="duration" class="text-xs text-gray-500">
              {{ (duration / 1000).toFixed(2) }}s
            </span>
            <NTag v-if="loading" type="warning" size="small">执行中</NTag>
          </div>
        </div>
        <NScrollbar style="max-height: 200px">
          <NCode :code="sql" language="sql" class="p-4" />
        </NScrollbar>
      </div>
    </div>
  </div>
</template>
