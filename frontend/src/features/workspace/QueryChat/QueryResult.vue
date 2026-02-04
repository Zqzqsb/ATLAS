<script setup lang="ts">
import { computed } from 'vue'
import { NCode, NScrollbar, NTag, NButton } from 'naive-ui'

const props = defineProps<{
  sql?: string
  error?: string | null
  duration?: number
  result?: any[] | null
  loading?: boolean
}>()

const emit = defineEmits<{
  execute: []
  copy: []
}>()

const hasResult = computed(() => props.result && props.result.length > 0)
const resultColumns = computed(() => {
  if (!hasResult.value) return []
  return Object.keys(props.result![0])
})
</script>

<template>
  <div class="query-result rounded-2xl overflow-hidden bg-gradient-to-br from-gray-800/50 to-gray-900/50 backdrop-blur-md border border-white/10">
    <!-- SQL Display -->
    <div v-if="sql || error" class="sql-section">
      <div class="flex items-center justify-between px-6 py-4 border-b border-white/10 bg-gradient-to-r from-white/5 to-transparent">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-500/20 to-emerald-500/20 flex items-center justify-center border border-green-500/30">
            <div class="i-carbon-sql text-xl text-green-400" />
          </div>
          <div>
            <h3 class="font-semibold text-white">Generated SQL</h3>
            <p v-if="duration" class="text-xs text-gray-400 mt-0.5">
              Execution time: {{ (duration / 1000).toFixed(2) }}s
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <NButton
            v-if="sql"
            quaternary
            size="small"
            @click="emit('copy')"
          >
            <template #icon>
              <div class="i-carbon-copy" />
            </template>
            Copy
          </NButton>

          <NButton
            v-if="sql && !loading"
            type="primary"
            size="small"
            @click="emit('execute')"
          >
            <template #icon>
              <div class="i-carbon-play" />
            </template>
            Execute
          </NButton>

          <NTag v-if="loading" type="warning" size="small">
            <template #icon>
              <div class="i-carbon-hourglass animate-spin" />
            </template>
            Executing...
          </NTag>
        </div>
      </div>

      <!-- Error Display -->
      <div v-if="error" class="p-6 bg-red-500/10 border-l-4 border-red-500">
        <div class="flex items-start gap-3">
          <div class="i-carbon-warning text-xl text-red-400 flex-shrink-0 mt-1" />
          <div>
            <h4 class="text-red-400 font-semibold mb-1">Error</h4>
            <p class="text-sm text-red-300/80">{{ error }}</p>
          </div>
        </div>
      </div>

      <!-- SQL Code -->
      <div v-if="sql" class="bg-[#1e1e1e]">
        <NScrollbar style="max-height: 300px">
          <NCode :code="sql" language="sql" class="text-sm" />
        </NScrollbar>
      </div>
    </div>

    <!-- Execution Result -->
    <div v-if="hasResult" class="result-section">
      <div class="flex items-center justify-between px-6 py-4 border-b border-white/10 bg-gradient-to-r from-white/5 to-transparent">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center border border-blue-500/30">
            <div class="i-carbon-data-table text-xl text-blue-400" />
          </div>
          <div>
            <h3 class="font-semibold text-white">Query Result</h3>
            <p class="text-xs text-gray-400 mt-0.5">
              {{ result?.length }} rows × {{ resultColumns.length }} columns
            </p>
          </div>
        </div>
      </div>

      <div class="p-6">
        <NScrollbar x-scrollable>
          <table class="result-table w-full">
            <thead>
              <tr>
                <th
                  v-for="col in resultColumns"
                  :key="col"
                  class="px-4 py-3 text-left text-xs font-semibold text-gray-400 bg-white/5 border-b border-white/10"
                >
                  {{ col }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in result"
                :key="idx"
                class="border-b border-white/5 hover:bg-white/5 transition-colors"
              >
                <td
                  v-for="col in resultColumns"
                  :key="col"
                  class="px-4 py-3 text-sm text-gray-300"
                >
                  {{ row[col] }}
                </td>
              </tr>
            </tbody>
          </table>
        </NScrollbar>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!sql && !error && !hasResult" class="p-12 text-center">
      <div class="w-16 h-16 rounded-2xl bg-white/5 flex items-center justify-center mx-auto mb-4">
        <div class="i-carbon-query text-3xl text-gray-600" />
      </div>
      <p class="text-gray-500">No query result yet</p>
    </div>
  </div>
</template>

<style scoped>
.result-table {
  min-width: 100%;
  border-collapse: collapse;
}

.result-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  backdrop-filter: blur(8px);
}

.result-table tbody tr:last-child {
  border-bottom: none;
}
</style>
