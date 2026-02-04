<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NTimeline, NTimelineItem, NTag, NButton, NCode, NCollapse, NCollapseItem, NEmpty } from 'naive-ui'
import { useDemoStore } from '@/stores/demo'
import type { MaintenanceType, MaintenanceStatus } from '@/types'

const demoStore = useDemoStore()

const typeConfig: Record<MaintenanceType, { label: string; icon: string; color: string }> = {
  error_feedback: { label: '错误反馈', icon: 'i-carbon-warning', color: 'error' },
  user_correction: { label: '用户纠正', icon: 'i-carbon-user-feedback', color: 'info' },
  schema_change: { label: 'Schema变更', icon: 'i-carbon-data-table', color: 'warning' },
  pattern_learning: { label: '模式学习', icon: 'i-carbon-machine-learning', color: 'success' }
}

const statusConfig: Record<MaintenanceStatus, { label: string; type: string }> = {
  pending: { label: '待处理', type: 'default' },
  analyzing: { label: '分析中', type: 'warning' },
  applied: { label: '已应用', type: 'info' },
  verified: { label: '已验证', type: 'success' },
  rejected: { label: '已拒绝', type: 'error' }
}

async function triggerMaintenance(type: string) {
  await demoStore.triggerMaintenance(type)
  await demoStore.loadMaintenanceLogs()
}
</script>

<template>
  <div class="self-maintain-demo">
    <!-- Intro -->
    <div class="text-center mb-8">
      <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-2">
        Rich Context 自维持演示
      </h2>
      <p class="text-gray-500 max-w-2xl mx-auto">
        展示系统如何通过错误反馈、用户纠正、Schema变更等机制自动维护和优化 Rich Context
      </p>
    </div>

    <!-- Trigger buttons -->
    <div class="flex flex-wrap justify-center gap-3 mb-8">
      <NButton
        v-for="(config, type) in typeConfig"
        :key="type"
        :type="config.color as any"
        secondary
        @click="triggerMaintenance(type)"
      >
        <template #icon>
          <div :class="config.icon" />
        </template>
        模拟 {{ config.label }}
      </NButton>
    </div>

    <!-- Maintenance pipeline visualization -->
    <NCard title="自维持流程" class="mb-8">
      <div class="flex items-center justify-between px-8 py-4">
        <!-- Trigger -->
        <div class="flex flex-col items-center">
          <div class="w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mb-2">
            <div class="i-carbon-flash text-2xl text-blue-500" />
          </div>
          <span class="text-sm font-medium">触发</span>
          <span class="text-xs text-gray-400">错误/反馈/变更</span>
        </div>

        <div class="flex-1 h-0.5 bg-gray-200 dark:bg-gray-700 mx-4 relative">
          <div class="absolute top-1/2 left-1/4 transform -translate-y-1/2 -translate-x-1/2">
            <div class="i-carbon-arrow-right text-gray-400" />
          </div>
        </div>

        <!-- Analyze -->
        <div class="flex flex-col items-center">
          <div class="w-16 h-16 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center mb-2">
            <div class="i-carbon-analytics text-2xl text-purple-500" />
          </div>
          <span class="text-sm font-medium">分析</span>
          <span class="text-xs text-gray-400">LLM 推理</span>
        </div>

        <div class="flex-1 h-0.5 bg-gray-200 dark:bg-gray-700 mx-4 relative">
          <div class="absolute top-1/2 left-1/4 transform -translate-y-1/2 -translate-x-1/2">
            <div class="i-carbon-arrow-right text-gray-400" />
          </div>
        </div>

        <!-- Generate -->
        <div class="flex flex-col items-center">
          <div class="w-16 h-16 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center mb-2">
            <div class="i-carbon-document-add text-2xl text-green-500" />
          </div>
          <span class="text-sm font-medium">生成</span>
          <span class="text-xs text-gray-400">Context 更新</span>
        </div>

        <div class="flex-1 h-0.5 bg-gray-200 dark:bg-gray-700 mx-4 relative">
          <div class="absolute top-1/2 left-1/4 transform -translate-y-1/2 -translate-x-1/2">
            <div class="i-carbon-arrow-right text-gray-400" />
          </div>
        </div>

        <!-- Verify -->
        <div class="flex flex-col items-center">
          <div class="w-16 h-16 rounded-full bg-yellow-100 dark:bg-yellow-900/30 flex items-center justify-center mb-2">
            <div class="i-carbon-checkmark-outline text-2xl text-yellow-500" />
          </div>
          <span class="text-sm font-medium">验证</span>
          <span class="text-xs text-gray-400">人工/自动</span>
        </div>

        <div class="flex-1 h-0.5 bg-gray-200 dark:bg-gray-700 mx-4 relative">
          <div class="absolute top-1/2 left-1/4 transform -translate-y-1/2 -translate-x-1/2">
            <div class="i-carbon-arrow-right text-gray-400" />
          </div>
        </div>

        <!-- Apply -->
        <div class="flex flex-col items-center">
          <div class="w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mb-2">
            <div class="i-carbon-data-base text-2xl text-blue-500" />
          </div>
          <span class="text-sm font-medium">存储</span>
          <span class="text-xs text-gray-400">写入 MariaDB</span>
        </div>
      </div>
    </NCard>

    <!-- Maintenance logs -->
    <NCard title="维护日志">
      <template #header-extra>
        <NButton quaternary size="small" @click="demoStore.loadMaintenanceLogs">
          <template #icon>
            <div class="i-carbon-refresh" />
          </template>
          刷新
        </NButton>
      </template>

      <NEmpty v-if="demoStore.maintenanceLogs.length === 0" description="暂无维护记录" />

      <NTimeline v-else>
        <NTimelineItem
          v-for="log in demoStore.maintenanceLogs"
          :key="log.id"
          :type="typeConfig[log.type]?.color as any || 'default'"
        >
          <template #icon>
            <div :class="typeConfig[log.type]?.icon" />
          </template>

          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <NTag :type="typeConfig[log.type]?.color as any" size="small">
                  {{ typeConfig[log.type]?.label }}
                </NTag>
                <NTag :type="statusConfig[log.status]?.type as any" size="small">
                  {{ statusConfig[log.status]?.label }}
                </NTag>
                <span class="text-xs text-gray-400">{{ log.timestamp }}</span>
              </div>

              <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">
                <span class="font-medium">触发：</span>{{ log.trigger }}
              </p>

              <p class="text-sm text-gray-700 dark:text-gray-300">
                <span class="font-medium">行动：</span>{{ log.action }}
              </p>

              <!-- Context changes -->
              <NCollapse v-if="log.contextBefore || log.contextAfter" arrow-placement="right" class="mt-2">
                <NCollapseItem title="Context 变更详情" name="detail">
                  <div class="space-y-2">
                    <div v-if="log.contextBefore" class="p-2 bg-red-50 dark:bg-red-900/20 rounded">
                      <p class="text-xs text-gray-500 mb-1">变更前：</p>
                      <p class="text-sm text-red-600 dark:text-red-400 line-through">
                        {{ log.contextBefore.content }}
                      </p>
                    </div>
                    <div v-if="log.contextAfter" class="p-2 bg-green-50 dark:bg-green-900/20 rounded">
                      <p class="text-xs text-gray-500 mb-1">变更后：</p>
                      <p class="text-sm text-green-600 dark:text-green-400">
                        {{ log.contextAfter.content }}
                      </p>
                    </div>
                  </div>
                </NCollapseItem>
              </NCollapse>
            </div>

            <div v-if="log.status === 'pending'" class="flex gap-1 ml-4">
              <NButton size="tiny" type="success">应用</NButton>
              <NButton size="tiny" type="error">拒绝</NButton>
            </div>
          </div>
        </NTimelineItem>
      </NTimeline>
    </NCard>
  </div>
</template>
