<script setup lang="ts">
import { useDemoStore } from '@/stores/demo'

const store = useDemoStore()
</script>

<template>
  <div class="space-y-6">
    <!-- Title -->
    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-2 flex items-center gap-2">
        <span class="i-lucide-trending-up text-blue-500" />
        基准测试
      </h2>
      <p class="text-gray-600 text-sm">
        在 Spider 数据集上运行 Text-to-SQL 查询，查看完整的推理过程。
      </p>
    </div>

    <!-- Query Input -->
    <div class="card p-6">
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 mb-2">自然语言问题</label>
        <textarea
          v-model="store.question"
          class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          rows="3"
          placeholder="例如: 查询所有员工的姓名和部门"
        />
      </div>

      <!-- Options -->
      <div class="flex flex-wrap gap-4 mb-4">
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="store.options.useRichContext" class="rounded" />
          <span class="text-sm">使用 Rich Context</span>
        </label>
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="store.options.useReact" class="rounded" />
          <span class="text-sm">使用 ReAct</span>
        </label>
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="store.options.useGrounding" class="rounded" />
          <span class="text-sm">使用语义定位</span>
        </label>
      </div>

      <button
        @click="store.runQuery()"
        :disabled="!store.question || store.isLoading"
        class="btn-primary"
      >
        <span v-if="store.isLoading" class="i-lucide-loader-2 animate-spin mr-2" />
        {{ store.isLoading ? '推理中...' : '开始推理' }}
      </button>
    </div>

    <!-- Results -->
    <div v-if="store.sql" class="card p-6">
      <h3 class="font-medium mb-3">生成的 SQL</h3>
      <pre class="bg-gray-900 text-green-400 p-4 rounded-lg overflow-x-auto">{{ store.sql }}</pre>
      <div class="mt-2 text-sm text-gray-500">
        耗时: {{ store.duration }}ms
      </div>
    </div>

    <!-- ReAct Steps -->
    <div v-if="store.reactSteps.length" class="card p-6">
      <h3 class="font-medium mb-3">推理步骤</h3>
      <div class="space-y-3">
        <div 
          v-for="step in store.reactSteps" 
          :key="step.step"
          class="p-3 rounded-lg"
          :class="{
            'bg-yellow-50': step.type === 'thought',
            'bg-blue-50': step.type === 'action',
            'bg-gray-50': step.type === 'observation',
            'bg-green-50': step.type === 'answer'
          }"
        >
          <div class="text-sm font-medium text-gray-600 mb-1">
            Step {{ step.step }} - {{ step.type }}
          </div>
          <div class="text-sm">{{ step.content }}</div>
        </div>
      </div>
    </div>
  </div>
</template>
