<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { NStatistic, NCard, NGrid, NGridItem, NDataTable, NTag, NEmpty, NButton } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { databaseApi } from '@/api'

const workspaceStore = useWorkspaceStore()

const stats = ref({
  queryCount: 0,
  avgDuration: 0,
  successRate: 0,
  contextUsageRate: 0
})

const loading = ref(false)

onMounted(async () => {
  if (workspaceStore.currentDatabaseId) {
    loading.value = true
    try {
      stats.value = await databaseApi.getStats(workspaceStore.currentDatabaseId)
    } finally {
      loading.value = false
    }
  }
})

// History columns
const historyColumns = [
  {
    title: '时间',
    key: 'timestamp',
    width: 160,
    render: (row: any) => new Date(row.timestamp).toLocaleString()
  },
  {
    title: '问题',
    key: 'question',
    ellipsis: { tooltip: true }
  },
  {
    title: '耗时',
    key: 'duration',
    width: 100,
    render: (row: any) => `${row.duration.toFixed(2)}s`
  },
  {
    title: '状态',
    key: 'feedback',
    width: 100,
    render: (row: any) => {
      if (row.feedback === 'positive') {
        return h(NTag, { type: 'success', size: 'small' }, { default: () => '正确' })
      }
      if (row.feedback === 'negative') {
        return h(NTag, { type: 'error', size: 'small' }, { default: () => '错误' })
      }
      return h(NTag, { size: 'small' }, { default: () => '未评价' })
    }
  }
]

import { h } from 'vue'
</script>

<template>
  <div class="monitor-page p-6">
    <!-- Stats cards -->
    <NGrid :x-gap="16" :y-gap="16" :cols="4" class="mb-6">
      <NGridItem>
        <NCard>
          <NStatistic label="查询总数" :value="stats.queryCount">
            <template #prefix>
              <div class="i-carbon-query text-blue-500" />
            </template>
          </NStatistic>
        </NCard>
      </NGridItem>
      
      <NGridItem>
        <NCard>
          <NStatistic label="平均耗时" :value="stats.avgDuration" :precision="2">
            <template #prefix>
              <div class="i-carbon-time text-green-500" />
            </template>
            <template #suffix>秒</template>
          </NStatistic>
        </NCard>
      </NGridItem>
      
      <NGridItem>
        <NCard>
          <NStatistic label="成功率" :value="stats.successRate * 100" :precision="1">
            <template #prefix>
              <div class="i-carbon-checkmark-filled text-green-500" />
            </template>
            <template #suffix>%</template>
          </NStatistic>
        </NCard>
      </NGridItem>
      
      <NGridItem>
        <NCard>
          <NStatistic label="Context 使用率" :value="stats.contextUsageRate * 100" :precision="1">
            <template #prefix>
              <div class="i-carbon-magic-wand text-purple-500" />
            </template>
            <template #suffix>%</template>
          </NStatistic>
        </NCard>
      </NGridItem>
    </NGrid>

    <!-- Recent queries -->
    <NCard title="最近查询">
      <template #header-extra>
        <NButton quaternary size="small" @click="workspaceStore.fetchQueryHistory">
          <template #icon>
            <div class="i-carbon-refresh" />
          </template>
          刷新
        </NButton>
      </template>

      <NDataTable
        v-if="workspaceStore.queryHistory.length > 0"
        :columns="historyColumns"
        :data="workspaceStore.queryHistory"
        :loading="workspaceStore.loadingHistory"
        :max-height="400"
        striped
      />
      <NEmpty v-else description="暂无查询记录" />
    </NCard>
  </div>
</template>
