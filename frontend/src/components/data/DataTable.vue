<script setup lang="ts">
import { computed } from 'vue'
import { NDataTable, NEmpty, NScrollbar } from 'naive-ui'

const props = withDefaults(defineProps<{
  columns: string[]
  data: any[][]
  loading?: boolean
  maxHeight?: string
  striped?: boolean
}>(), {
  loading: false,
  maxHeight: '400px',
  striped: true
})

const tableColumns = computed(() => 
  props.columns.map((col, index) => ({
    title: col,
    key: `col_${index}`,
    ellipsis: {
      tooltip: true
    },
    resizable: true,
    minWidth: 100
  }))
)

const tableData = computed(() => 
  props.data.map((row, rowIndex) => {
    const obj: Record<string, any> = { key: rowIndex }
    props.columns.forEach((_, colIndex) => {
      obj[`col_${colIndex}`] = row[colIndex]
    })
    return obj
  })
)
</script>

<template>
  <div class="data-table rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
    <NScrollbar :style="{ maxHeight }">
      <NDataTable
        v-if="data.length > 0"
        :columns="tableColumns"
        :data="tableData"
        :loading="loading"
        :striped="striped"
        :bordered="false"
        size="small"
        flex-height
      />
      <NEmpty v-else description="暂无数据" class="py-8" />
    </NScrollbar>
    
    <!-- Footer with row count -->
    <div 
      v-if="data.length > 0" 
      class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 text-sm text-gray-500"
    >
      共 {{ data.length }} 行
    </div>
  </div>
</template>

<style scoped>
.data-table :deep(.n-data-table) {
  --n-th-padding: 8px 12px;
  --n-td-padding: 6px 12px;
}

.data-table :deep(.n-data-table-th) {
  background: #f9fafb;
  font-weight: 600;
}

.dark .data-table :deep(.n-data-table-th) {
  background: #1f2937;
}
</style>
