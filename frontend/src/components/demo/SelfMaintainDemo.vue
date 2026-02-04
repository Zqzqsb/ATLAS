<script setup lang="ts">
import { ref } from 'vue'
import { useDemoStore } from '@/stores/demo'

const store = useDemoStore()
const activeDemo = ref<'error' | 'feedback' | 'schema' | null>(null)
const demoStatus = ref<'idle' | 'running' | 'done'>('idle')
const demoSteps = ref<{ label: string; status: 'pending' | 'running' | 'done' }[]>([])

async function runDemo(type: 'error' | 'feedback' | 'schema') {
  activeDemo.value = type
  demoStatus.value = 'running'
  
  const steps = {
    error: [
      { label: '检测到 SQL 执行错误', status: 'pending' as const },
      { label: '分析错误模式', status: 'pending' as const },
      { label: '生成修复 Context', status: 'pending' as const },
      { label: '验证修复效果', status: 'pending' as const }
    ],
    feedback: [
      { label: '接收用户反馈', status: 'pending' as const },
      { label: '解析反馈意图', status: 'pending' as const },
      { label: '更新 Rich Context', status: 'pending' as const },
      { label: '验证改进效果', status: 'pending' as const }
    ],
    schema: [
      { label: '检测 Schema 变更', status: 'pending' as const },
      { label: '分析影响范围', status: 'pending' as const },
      { label: '更新相关 Context', status: 'pending' as const },
      { label: '重新计算向量索引', status: 'pending' as const }
    ]
  }
  
  demoSteps.value = steps[type]

  for (let i = 0; i < demoSteps.value.length; i++) {
    const step = demoSteps.value[i]
    if (step) {
      step.status = 'running'
      await new Promise(r => setTimeout(r, 1000))
      step.status = 'done'
    }
  }

  demoStatus.value = 'done'
}
</script>

<template>
  <div class="space-y-6">
    <!-- Title -->
    <div class="card p-6">
      <h2 class="text-lg font-semibold mb-2 flex items-center gap-2">
        <span class="i-carbon-recycle text-green-500" />
        自维持机制演示
      </h2>
      <p class="text-gray-600 text-sm">
        展示 Rich Context 如何通过错误反馈、用户纠正、Schema 变更等触发器自动维护和更新。
      </p>
    </div>

    <!-- Trigger Cards -->
    <div class="grid grid-cols-3 gap-4">
      <button
        @click="runDemo('error')"
        class="card p-5 text-left hover:shadow-md transition-shadow"
        :class="activeDemo === 'error' ? 'ring-2 ring-orange-500' : ''"
      >
        <div class="i-carbon-warning-alt text-2xl text-orange-500 mb-3" />
        <div class="font-medium mb-1">SQL 错误反馈</div>
        <div class="text-sm text-gray-500">当 SQL 执行失败时自动分析并修复</div>
      </button>

      <button
        @click="runDemo('feedback')"
        class="card p-5 text-left hover:shadow-md transition-shadow"
        :class="activeDemo === 'feedback' ? 'ring-2 ring-blue-500' : ''"
      >
        <div class="i-carbon-user-feedback text-2xl text-blue-500 mb-3" />
        <div class="font-medium mb-1">用户纠正</div>
        <div class="text-sm text-gray-500">用户反馈错误时更新业务规则</div>
      </button>

      <button
        @click="runDemo('schema')"
        class="card p-5 text-left hover:shadow-md transition-shadow"
        :class="activeDemo === 'schema' ? 'ring-2 ring-purple-500' : ''"
      >
        <div class="i-carbon-data-table text-2xl text-purple-500 mb-3" />
        <div class="font-medium mb-1">Schema 变更</div>
        <div class="text-sm text-gray-500">数据库结构变化时自动更新</div>
      </button>
    </div>

    <!-- Demo Progress -->
    <div v-if="activeDemo" class="card p-6">
      <h3 class="font-medium mb-4 flex items-center gap-2">
        <span 
          class="text-lg"
          :class="{
            'i-carbon-warning-alt text-orange-500': activeDemo === 'error',
            'i-carbon-user-feedback text-blue-500': activeDemo === 'feedback',
            'i-carbon-data-table text-purple-500': activeDemo === 'schema'
          }"
        />
        {{ activeDemo === 'error' ? 'SQL 错误反馈流程' : activeDemo === 'feedback' ? '用户纠正流程' : 'Schema 变更流程' }}
      </h3>

      <div class="relative">
        <!-- Timeline -->
        <div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" />
        
        <div class="space-y-4">
          <div 
            v-for="(step, i) in demoSteps" 
            :key="i"
            class="relative pl-10"
          >
            <!-- Dot -->
            <div 
              class="absolute left-2.5 w-3 h-3 rounded-full border-2 bg-white"
              :class="{
                'border-gray-300': step.status === 'pending',
                'border-blue-500 animate-pulse': step.status === 'running',
                'border-green-500 bg-green-500': step.status === 'done'
              }"
            />
            
            <!-- Content -->
            <div 
              class="p-3 rounded-lg"
              :class="{
                'bg-gray-50': step.status === 'pending',
                'bg-blue-50': step.status === 'running',
                'bg-green-50': step.status === 'done'
              }"
            >
              <div class="flex items-center gap-2">
                <span v-if="step.status === 'running'" class="i-carbon-loading animate-spin" />
                <span v-else-if="step.status === 'done'" class="i-carbon-checkmark text-green-500" />
                <span class="font-medium">{{ step.label }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Result -->
      <div v-if="demoStatus === 'done'" class="mt-6 p-4 bg-green-50 rounded-lg border border-green-200">
        <div class="flex items-center gap-2 text-green-700 font-medium">
          <span class="i-carbon-checkmark-filled" />
          自维持流程完成
        </div>
        <p class="text-sm text-green-600 mt-1">
          Rich Context 已自动更新，后续查询将使用最新的上下文信息。
        </p>
      </div>
    </div>

    <!-- Maintenance Logs -->
    <div class="card p-6">
      <h3 class="font-medium mb-4 flex items-center gap-2">
        <span class="i-carbon-list text-gray-500" />
        维护日志
      </h3>

      <div class="space-y-3">
        <div 
          v-for="log in store.maintenanceLogs"
          :key="log.id"
          class="p-3 border rounded-lg"
        >
          <div class="flex-between mb-2">
            <span 
              class="px-2 py-0.5 text-xs rounded"
              :class="{
                'bg-orange-100 text-orange-700': log.type === 'error_feedback',
                'bg-blue-100 text-blue-700': log.type === 'user_correction',
                'bg-purple-100 text-purple-700': log.type === 'schema_change'
              }"
            >
              {{ log.type === 'error_feedback' ? '错误反馈' : log.type === 'user_correction' ? '用户纠正' : 'Schema变更' }}
            </span>
            <span class="text-xs text-gray-500">{{ log.timestamp }}</span>
          </div>
          <div class="text-sm">
            <div class="text-gray-600">触发: {{ log.trigger }}</div>
            <div class="text-gray-800 mt-1">操作: {{ log.action }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
