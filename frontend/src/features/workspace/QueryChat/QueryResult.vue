<script setup lang="ts">
import { ref, computed } from 'vue'
import { NCode, NScrollbar, NTag, NButton, NPagination, NDropdown, NTooltip, NInput, NModal, useMessage } from 'naive-ui'

const props = defineProps<{
  sql?: string
  error?: string | null
  duration?: number
  result?: any[] | null
  loading?: boolean
  queryId?: string
}>()

const emit = defineEmits<{
  execute: []
  retry: []
  feedback: [type: 'positive' | 'negative', note?: string]
}>()

const message = useMessage()

// Feedback state
const feedbackGiven = ref<'positive' | 'negative' | null>(null)
const showFeedbackModal = ref(false)
const feedbackNote = ref('')
const feedbackType = ref<'positive' | 'negative'>('negative')

// Pagination
const pageSize = ref(10)
const currentPage = ref(1)

const hasResult = computed(() => props.result && props.result.length > 0)
const resultColumns = computed(() => {
  if (!hasResult.value) return []
  return Object.keys(props.result![0])
})

const totalRows = computed(() => props.result?.length || 0)
const totalPages = computed(() => Math.ceil(totalRows.value / pageSize.value))

const paginatedResult = computed(() => {
  if (!props.result) return []
  const start = (currentPage.value - 1) * pageSize.value
  return props.result.slice(start, start + pageSize.value)
})

// Copy SQL
function copySql() {
  if (!props.sql) return
  navigator.clipboard.writeText(props.sql)
  message.success('SQL copied to clipboard')
}

// Export functions
const exportOptions = [
  { label: 'Export as CSV', key: 'csv' },
  { label: 'Export as JSON', key: 'json' }
]

function handleExport(key: string) {
  if (!props.result?.length) return
  
  if (key === 'csv') {
    const headers = resultColumns.value.join(',')
    const rows = props.result.map(row => 
      resultColumns.value.map(col => {
        const val = row[col]
        // Escape quotes and wrap in quotes if contains comma
        if (typeof val === 'string' && (val.includes(',') || val.includes('"'))) {
          return `"${val.replace(/"/g, '""')}"`
        }
        return val
      }).join(',')
    ).join('\n')
    const csv = `${headers}\n${rows}`
    downloadFile(csv, 'query_result.csv', 'text/csv')
    message.success('Exported as CSV')
  } else if (key === 'json') {
    const json = JSON.stringify(props.result, null, 2)
    downloadFile(json, 'query_result.json', 'application/json')
    message.success('Exported as JSON')
  }
}

function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

// Feedback functions
function handleQuickFeedback(type: 'positive' | 'negative') {
  if (type === 'positive') {
    feedbackGiven.value = 'positive'
    emit('feedback', 'positive')
    message.success('Thanks for your feedback!')
  } else {
    // Show modal for negative feedback to collect more info
    feedbackType.value = 'negative'
    feedbackNote.value = ''
    showFeedbackModal.value = true
  }
}

function submitFeedback() {
  feedbackGiven.value = feedbackType.value
  emit('feedback', feedbackType.value, feedbackNote.value || undefined)
  showFeedbackModal.value = false
  message.success('Feedback submitted. This will help improve the system.')
}
</script>

<template>
  <div class="query-result rounded-xl overflow-hidden bg-white border border-gray-200 shadow-sm">
    <!-- SQL Display -->
    <div v-if="sql || error" class="sql-section">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 bg-gray-50/50">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-green-50 flex items-center justify-center border border-green-100">
            <div class="i-lucide-database text-xl text-green-600" />
          </div>
          <div>
            <h3 class="font-bold text-gray-900">Generated SQL</h3>
            <p v-if="duration" class="text-xs text-gray-500 mt-0.5 font-medium">
              Execution time: {{ (duration / 1000).toFixed(2) }}s
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <!-- Feedback buttons -->
          <div v-if="sql && !error && !loading" class="flex items-center gap-1 mr-2">
            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  :type="feedbackGiven === 'positive' ? 'success' : 'default'"
                  :quaternary="feedbackGiven !== 'positive'"
                  size="small"
                  circle
                  :disabled="feedbackGiven !== null"
                  @click="handleQuickFeedback('positive')"
                >
                  <div class="i-lucide-thumbs-up" />
                </NButton>
              </template>
              SQL is correct
            </NTooltip>
            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  :type="feedbackGiven === 'negative' ? 'error' : 'default'"
                  :quaternary="feedbackGiven !== 'negative'"
                  size="small"
                  circle
                  :disabled="feedbackGiven !== null"
                  @click="handleQuickFeedback('negative')"
                >
                  <div class="i-lucide-thumbs-down" />
                </NButton>
              </template>
              SQL has issues
            </NTooltip>
          </div>

          <NButton
            v-if="sql"
            quaternary
            size="small"
            @click="copySql"
          >
            <template #icon>
              <div class="i-lucide-copy" />
            </template>
            Copy
          </NButton>

          <NTag v-if="loading" type="warning" size="small">
            <template #icon>
              <div class="i-lucide-loader-2 animate-spin" />
            </template>
            Executing...
          </NTag>

          <NTag v-if="feedbackGiven" :type="feedbackGiven === 'positive' ? 'success' : 'warning'" size="small">
            <template #icon>
              <div :class="feedbackGiven === 'positive' ? 'i-lucide-check' : 'i-lucide-alert-triangle'" />
            </template>
            {{ feedbackGiven === 'positive' ? 'Marked Correct' : 'Marked for Review' }}
          </NTag>
        </div>
      </div>

      <!-- Error Display -->
      <div v-if="error" class="p-6 bg-red-50 border-l-4 border-red-500">
        <div class="flex items-start justify-between">
          <div class="flex items-start gap-3">
            <div class="i-lucide-alert-triangle text-xl text-red-500 flex-shrink-0 mt-1" />
            <div>
              <h4 class="text-red-700 font-bold mb-1">Error</h4>
              <p class="text-sm text-red-600">{{ error }}</p>
            </div>
          </div>
          <NButton size="small" type="error" ghost @click="emit('retry')">
            <template #icon>
              <div class="i-lucide-rotate-ccw" />
            </template>
            Retry
          </NButton>
        </div>
      </div>

      <!-- SQL Code -->
      <div v-if="sql" class="bg-gray-50 border-b border-gray-100 p-4">
        <NScrollbar style="max-height: 300px">
          <NCode :code="sql" language="sql" class="text-sm font-mono text-gray-800" />
        </NScrollbar>
      </div>
    </div>

    <!-- Execution Result -->
    <div v-if="hasResult" class="result-section">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 bg-gray-50/50">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center border border-blue-100">
            <div class="i-lucide-table-2 text-xl text-blue-600" />
          </div>
          <div>
            <h3 class="font-bold text-gray-900">Query Result</h3>
            <p class="text-xs text-gray-500 mt-0.5 font-medium">
              {{ totalRows }} rows × {{ resultColumns.length }} columns
            </p>
          </div>
        </div>

        <!-- Export dropdown -->
        <NDropdown
          :options="exportOptions"
          @select="handleExport"
        >
          <NButton quaternary size="small">
            <template #icon>
              <div class="i-lucide-download" />
            </template>
            Export
          </NButton>
        </NDropdown>
      </div>

      <div class="p-6">
        <NScrollbar x-scrollable style="max-height: 400px;">
          <table class="result-table w-full">
            <thead>
              <tr>
                <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 bg-gray-50 border-b border-gray-200 w-12 uppercase tracking-wider">
                  #
                </th>
                <th
                  v-for="col in resultColumns"
                  :key="col"
                  class="px-4 py-3 text-left text-xs font-bold text-gray-500 bg-gray-50 border-b border-gray-200 uppercase tracking-wider"
                >
                  {{ col }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in paginatedResult"
                :key="idx"
                class="border-b border-gray-100 hover:bg-gray-50 transition-colors"
              >
                <td class="px-4 py-3 text-sm text-gray-500 font-medium">
                  {{ (currentPage - 1) * pageSize + idx + 1 }}
                </td>
                <td
                  v-for="col in resultColumns"
                  :key="col"
                  class="px-4 py-3 text-sm text-gray-700"
                >
                  {{ row[col] ?? '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </NScrollbar>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex items-center justify-between mt-4 pt-4 border-t border-gray-100">
          <span class="text-xs text-gray-500 font-medium">
            Showing {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, totalRows) }} of {{ totalRows }}
          </span>
          <NPagination
            v-model:page="currentPage"
            :page-count="totalPages"
            :page-size="pageSize"
            show-quick-jumper
          />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!sql && !error && !hasResult" class="p-12 text-center bg-gray-50/50">
      <div class="w-16 h-16 rounded-2xl bg-white border border-gray-200 shadow-sm flex items-center justify-center mx-auto mb-4">
        <div class="i-lucide-search text-3xl text-gray-400" />
      </div>
      <p class="text-gray-500 font-medium">No query result yet</p>
    </div>

    <!-- Feedback Modal -->
    <NModal v-model:show="showFeedbackModal" preset="card" title="Report Issue" style="width: 500px;">
      <div class="space-y-4">
        <p class="text-sm text-gray-500">
          Please describe the issue with the generated SQL. Your feedback helps improve the system.
        </p>
        
        <div>
          <label class="text-xs font-bold text-gray-700 mb-2 block uppercase tracking-wide">What's wrong?</label>
          <NInput
            v-model:value="feedbackNote"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="e.g., Wrong table selected, missing JOIN condition, incorrect column name..."
          />
        </div>
        
        <div class="p-3 rounded-lg bg-blue-50 border border-blue-100">
          <div class="flex items-start gap-2">
            <div class="i-lucide-info text-blue-600 mt-0.5" />
            <p class="text-xs text-blue-800 font-medium">
              Your feedback will be used to update the Rich Context, helping the system generate better SQL in the future.
            </p>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showFeedbackModal = false">Cancel</NButton>
          <NButton type="primary" @click="submitFeedback">
            <template #icon>
              <div class="i-lucide-send" />
            </template>
            Submit Feedback
          </NButton>
        </div>
      </template>
    </NModal>
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
}

.result-table tbody tr:last-child {
  border-bottom: none;
}
</style>
