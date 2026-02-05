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
  <div class="data-table rounded-lg border border-gray-200 overflow-hidden bg-white shadow-sm">
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
      <NEmpty v-else description="暂无数据" class="py-12" />
    </NScrollbar>
    
    <!-- Footer with row count -->
    <div 
      v-if="data.length > 0" 
      class="px-4 py-2 bg-gray-50 border-t border-gray-200 text-xs font-medium text-gray-500 flex justify-end"
    >
      Total {{ data.length }} rows
    </div>
  </div>
</template>

<style scoped>
.data-table :deep(.n-data-table) {
  --n-th-padding: 10px 16px;
  --n-td-padding: 8px 16px;
}

.data-table :deep(.n-data-table-th) {
  background: #f9fafb;
  font-weight: 700;
  color: #374151;
  text-transform: uppercase;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e5e7eb;
}

.data-table :deep(.n-data-table-td) {
  color: #4b5563;
  font-size: 0.875rem;
}
</style>
