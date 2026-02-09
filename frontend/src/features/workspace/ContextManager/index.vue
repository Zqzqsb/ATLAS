<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { 
  NButton, 
  NInput, 
  NSelect, 
  NEmpty, 
  NSpin, 
  NTag, 
  NCard, 
  NModal,
  NForm,
  NFormItem,
  NSpace,
  NProgress,
  NPopconfirm,
  useMessage
} from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import { useContextGenerationStore } from '@/stores/contextGeneration'
import { databaseApi } from '@/api/database'
import type { RichContext, ContextType } from '@/types'
import GenerateContextConsole from './GenerateContextConsole.vue'

const workspaceStore = useWorkspaceStore()
const ctxGenStore = useContextGenerationStore()
const message = useMessage()
const isPruning = ref(false)

const searchKeyword = ref('')
const filterTable = ref<string | null>(null)
const filterType = ref<ContextType | null>(null)
const generateConsoleRef = ref<any>(null)

// Edit dialog
const showEditDialog = ref(false)
const editingContext = ref<RichContext | null>(null)
const editForm = ref({
  tableName: '',
  columnName: '',
  type: 'description' as ContextType,
  content: ''
})

const typeOptions = [
  { label: 'Description', value: 'description' },
  { label: 'Example', value: 'example' },
  { label: 'Constraint', value: 'constraint' },
  { label: 'Synonym', value: 'synonym' },
  { label: 'Value Mapping', value: 'value_mapping' },
  { label: 'Business Rule', value: 'business_rule' },
  { label: 'Calculation', value: 'calculation' }
]

const tableOptions = computed(() => 
  workspaceStore.tableNames.map(name => ({ label: name, value: name }))
)

const filteredContexts = computed(() => {
  let contexts = workspaceStore.contexts

  if (filterTable.value) {
    contexts = contexts.filter(c => c.tableName === filterTable.value)
  }
  if (filterType.value) {
    contexts = contexts.filter(c => c.type === filterType.value)
  }
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    contexts = contexts.filter(c => 
      c.content.toLowerCase().includes(keyword) ||
      c.tableName.toLowerCase().includes(keyword) ||
      c.columnName?.toLowerCase().includes(keyword)
    )
  }

  return contexts
})

// Group contexts by table for structured display
interface TableContextGroup {
  tableName: string
  tableContext: RichContext | null
  columnContexts: RichContext[]
}

const groupedContexts = computed<TableContextGroup[]>(() => {
  const groups = new Map<string, TableContextGroup>()
  
  for (const ctx of filteredContexts.value) {
    if (!groups.has(ctx.tableName)) {
      groups.set(ctx.tableName, {
        tableName: ctx.tableName,
        tableContext: null,
        columnContexts: []
      })
    }
    
    const group = groups.get(ctx.tableName)!
    if (!ctx.columnName) {
      // Table-level context
      group.tableContext = ctx
    } else {
      // Column-level context
      group.columnContexts.push(ctx)
    }
  }
  
  // Sort column contexts by column name
  for (const group of groups.values()) {
    group.columnContexts.sort((a, b) => (a.columnName || '').localeCompare(b.columnName || ''))
  }
  
  return Array.from(groups.values()).sort((a, b) => a.tableName.localeCompare(b.tableName))
})

// Track expanded tables
const expandedTables = ref<Set<string>>(new Set())

function toggleTable(tableName: string) {
  if (expandedTables.value.has(tableName)) {
    expandedTables.value.delete(tableName)
  } else {
    expandedTables.value.add(tableName)
  }
}

function expandAll() {
  groupedContexts.value.forEach(g => expandedTables.value.add(g.tableName))
}

function collapseAll() {
  expandedTables.value.clear()
}

function openCreateDialog() {
  editingContext.value = null
  editForm.value = {
    tableName: filterTable.value || '',
    columnName: '',
    type: 'description',
    content: ''
  }
  showEditDialog.value = true
}

function openCreateDialogForTable(tableName: string) {
  editingContext.value = null
  editForm.value = {
    tableName,
    columnName: '',
    type: 'description',
    content: ''
  }
  showEditDialog.value = true
}

function openEditDialog(ctx: RichContext) {
  editingContext.value = ctx
  editForm.value = {
    tableName: ctx.tableName,
    columnName: ctx.columnName || '',
    type: ctx.type,
    content: ctx.content
  }
  showEditDialog.value = true
}

async function handleSave() {
  if (!editForm.value.tableName || !editForm.value.content) {
    message.warning('Please fill in all required fields')
    return
  }

  if (editingContext.value) {
    // Update
    await workspaceStore.updateContext(editingContext.value.id, {
      tableName: editForm.value.tableName,
      columnName: editForm.value.columnName || undefined,
      type: editForm.value.type,
      content: editForm.value.content
    })
    message.success('Updated successfully')
  } else {
    // Create
    await workspaceStore.addContext({
      databaseId: workspaceStore.currentDatabaseId!,
      tableId: editForm.value.tableName,
      tableName: editForm.value.tableName,
      columnName: editForm.value.columnName || undefined,
      type: editForm.value.type,
      content: editForm.value.content,
      source: 'manual'
    })
    message.success('Added successfully')
  }

  showEditDialog.value = false
}

async function handleDelete(ctx: RichContext) {
  await workspaceStore.deleteContext(ctx.id)
  message.success('Deleted successfully')
}

function getTypeColor(type: ContextType): string {
  const colors: Record<ContextType, string> = {
    description: 'info',
    example: 'success',
    constraint: 'warning',
    synonym: 'primary',
    value_mapping: 'error',
    business_rule: 'info',
    calculation: 'warning'
  }
  return colors[type] || 'default'
}

// Return Tailwind CSS classes for each context type badge
function getTypeBadgeClasses(type: ContextType): string {
  const map: Record<ContextType, string> = {
    description: 'bg-blue-100 text-blue-700',
    example: 'bg-amber-100 text-amber-700',
    constraint: 'bg-red-100 text-red-700',
    synonym: 'bg-purple-100 text-purple-700',
    value_mapping: 'bg-pink-100 text-pink-700',
    business_rule: 'bg-indigo-100 text-indigo-700',
    calculation: 'bg-orange-100 text-orange-700'
  }
  return map[type] || 'bg-gray-100 text-gray-700'
}

// Open generate console
function openGenerateConsole() {
  if (!workspaceStore.currentDatabaseId) {
    message.warning('Please select a database first')
    return
  }
  ctxGenStore.openConsole(workspaceStore.currentDatabaseId)
}

// Handle generation complete
async function handleGenerateComplete() {
  ctxGenStore.reset()
  // Refresh contexts and schema
  await workspaceStore.fetchContexts()
  await workspaceStore.fetchSchema()
}

// Handle minimize to background
function handleMinimize() {
  // Store handles minimization state
}

// Handle prune all context
async function handlePruneAll() {
  const lakebaseId = workspaceStore.currentDatabase?.metadata?.lakebaseId
  if (!lakebaseId) {
    message.warning('Unable to get datasource ID')
    return
  }

  isPruning.value = true
  try {
    const result = await databaseApi.pruneContext(lakebaseId)
    if (result.success) {
      message.success(result.message || 'All Rich Context cleared')
      // Refresh contexts and schema
      await workspaceStore.fetchContexts()
      await workspaceStore.fetchSchema()
    } else {
      message.error(result.message || 'Failed to clear context')
    }
  } catch (e: any) {
    message.error(`Failed to clear: ${e.message}`)
  } finally {
    isPruning.value = false
  }
}
</script>

<template>
  <div class="context-manager p-6">
    <!-- Toolbar -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <NInput
          v-model:value="searchKeyword"
          placeholder="Search context..."
          clearable
          style="width: 240px"
        >
          <template #prefix>
            <div class="i-carbon-search text-gray-400" />
          </template>
        </NInput>

        <NSelect
          v-model:value="filterTable"
          :options="tableOptions"
          placeholder="Filter table"
          clearable
          style="width: 160px"
        />

        <NSelect
          v-model:value="filterType"
          :options="typeOptions"
          placeholder="Filter type"
          clearable
          style="width: 140px"
        />
      </div>

      <div class="flex items-center gap-2">
        <NButton @click="workspaceStore.fetchContexts">
          <template #icon>
            <div class="i-carbon-refresh" />
          </template>
          Refresh
        </NButton>
        <NPopconfirm
          @positive-click="handlePruneAll"
          positive-text="Confirm"
          negative-text="Cancel"
        >
          <template #trigger>
            <NButton 
              type="error" 
              :loading="isPruning"
              :disabled="filteredContexts.length === 0"
            >
              <template #icon>
                <div class="i-carbon-trash-can" />
              </template>
              Clear All
            </NButton>
          </template>
          <div class="max-w-xs">
            <p class="font-semibold mb-2">Clear all Rich Context?</p>
            <p class="text-sm text-gray-500">This will delete all table descriptions, column descriptions, business terms and their vector embeddings. This action cannot be undone.</p>
          </div>
        </NPopconfirm>
        <NButton 
          type="info" 
          @click="openGenerateConsole"
        >
          <template #icon>
            <div class="i-carbon-machine-learning-model" />
          </template>
          AI Generate
        </NButton>
        <NButton type="primary" @click="openCreateDialog">
          <template #icon>
            <div class="i-carbon-add" />
          </template>
          Add Context
        </NButton>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="workspaceStore.loadingContexts" class="flex justify-center py-16">
      <NSpin size="large" />
    </div>

    <!-- Empty -->
    <NEmpty 
      v-else-if="filteredContexts.length === 0" 
      description="No context yet"
      class="py-16"
    >
      <template #extra>
        <NButton type="primary" @click="openCreateDialog">
          Add First Context
        </NButton>
      </template>
    </NEmpty>

    <!-- Structured Context List -->
    <div v-else class="context-tree">
      <!-- Expand/Collapse All -->
      <div class="flex gap-2 mb-4">
        <NButton size="small" quaternary @click="expandAll">
          <template #icon><div class="i-carbon-expand-all" /></template>
          Expand All
        </NButton>
        <NButton size="small" quaternary @click="collapseAll">
          <template #icon><div class="i-carbon-collapse-all" /></template>
          Collapse All
        </NButton>
        <span class="text-sm text-gray-500 ml-auto">
          {{ groupedContexts.length }} tables, {{ filteredContexts.length }} contexts
        </span>
      </div>

      <!-- Table Groups -->
      <div 
        v-for="group in groupedContexts" 
        :key="group.tableName"
        class="table-group mb-3"
      >
        <!-- Table Header - Modern Collapsible -->
        <div 
          class="table-header flex items-center gap-3 px-4 py-3 bg-white border border-gray-200 cursor-pointer hover:bg-gray-50 transition-all"
          :class="expandedTables.has(group.tableName) ? 'rounded-t-lg border-b-0' : 'rounded-lg'"
          @click="toggleTable(group.tableName)"
        >
          <div class="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center flex-shrink-0">
            <div class="i-carbon-data-table text-lg text-blue-600" />
          </div>
          
          <span class="font-semibold text-gray-900">{{ group.tableName }}</span>
          
          <span class="px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
            {{ group.columnContexts.length + (group.tableContext ? 1 : 0) }} contexts
          </span>

          <div class="ml-auto flex items-center gap-3">
            <button 
              class="w-7 h-7 rounded-md bg-gray-100 hover:bg-gray-200 flex items-center justify-center transition-colors"
              @click.stop="openCreateDialogForTable(group.tableName)"
            >
              <div class="i-carbon-add text-sm text-gray-600" />
            </button>
            
            <!-- Modern expand/collapse indicator -->
            <div 
              class="w-7 h-7 rounded-md bg-gray-100 flex items-center justify-center transition-transform duration-200"
              :class="{ 'rotate-180': expandedTables.has(group.tableName) }"
            >
              <div class="i-carbon-chevron-down text-sm text-gray-500" />
            </div>
          </div>
        </div>

        <!-- Expanded Content - Modern Panel -->
        <div 
          v-if="expandedTables.has(group.tableName)"
          class="table-content bg-gray-50 border border-t-0 border-gray-200 rounded-b-lg p-4"
        >
          <!-- Table-level Context -->
          <div 
            v-if="group.tableContext" 
            class="context-item p-4 mb-3 rounded-lg bg-white border border-amber-200"
          >
            <div class="flex items-center gap-2 mb-2">
              <div class="w-6 h-6 rounded bg-amber-100 flex items-center justify-center">
                <div class="i-carbon-document text-amber-600 text-sm" />
              </div>
              <span class="text-sm font-semibold text-gray-900">Table Description</span>
              <span class="px-2 py-0.5 rounded text-xs font-medium" :class="getTypeBadgeClasses(group.tableContext.type)">{{ group.tableContext.type }}</span>
              <div class="ml-auto flex gap-1">
                <button class="w-7 h-7 rounded hover:bg-gray-100 flex items-center justify-center transition-colors" @click="openEditDialog(group.tableContext!)">
                  <div class="i-carbon-edit text-sm text-gray-500" />
                </button>
                <button class="w-7 h-7 rounded hover:bg-red-50 flex items-center justify-center transition-colors" @click="handleDelete(group.tableContext!)">
                  <div class="i-carbon-trash-can text-sm text-red-500" />
                </button>
              </div>
            </div>
            <p class="text-sm text-gray-600 leading-relaxed">{{ group.tableContext.content }}</p>
          </div>

          <!-- Column Contexts -->
          <div 
            v-for="colCtx in group.columnContexts" 
            :key="colCtx.id"
            class="context-item p-4 mb-3 rounded-lg bg-white border border-emerald-200"
          >
            <div class="flex items-center gap-2 mb-2">
              <div class="w-6 h-6 rounded bg-emerald-100 flex items-center justify-center">
                <div class="i-carbon-column text-emerald-600 text-sm" />
              </div>
              <span class="text-sm font-semibold text-gray-900">{{ colCtx.columnName }}</span>
              <span class="px-2 py-0.5 rounded text-xs font-medium" :class="getTypeBadgeClasses(colCtx.type)">{{ colCtx.type }}</span>
              <div class="ml-auto flex gap-1">
                <button class="w-7 h-7 rounded hover:bg-gray-100 flex items-center justify-center transition-colors" @click="openEditDialog(colCtx)">
                  <div class="i-carbon-edit text-sm text-gray-500" />
                </button>
                <button class="w-7 h-7 rounded hover:bg-red-50 flex items-center justify-center transition-colors" @click="handleDelete(colCtx)">
                  <div class="i-carbon-trash-can text-sm text-red-500" />
                </button>
              </div>
            </div>
            <p class="text-sm text-gray-600 leading-relaxed">{{ colCtx.content }}</p>
          </div>

          <!-- Empty columns hint -->
          <div 
            v-if="group.columnContexts.length === 0 && !group.tableContext"
            class="text-sm text-gray-400 py-4 text-center"
          >
            No context yet. Click the + button above to add.
          </div>
        </div>
      </div>
    </div>

    <!-- Edit/Create Dialog -->
    <NModal
      v-model:show="showEditDialog"
      preset="card"
      :title="editingContext ? 'Edit Context' : 'Add Context'"
      style="width: 500px"
    >
      <NForm :model="editForm" label-placement="left" label-width="80">
        <NFormItem label="Table" required>
          <NSelect
            v-model:value="editForm.tableName"
            :options="tableOptions"
            placeholder="Select table"
          />
        </NFormItem>
        <NFormItem label="Column">
          <NInput
            v-model:value="editForm.columnName"
            placeholder="Optional. Leave blank for table-level context"
          />
        </NFormItem>
        <NFormItem label="Type" required>
          <NSelect
            v-model:value="editForm.type"
            :options="typeOptions"
          />
        </NFormItem>
        <NFormItem label="Content" required>
          <NInput
            v-model:value="editForm.content"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="Enter context content..."
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditDialog = false">Cancel</NButton>
          <NButton type="primary" @click="handleSave">Save</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- Generate Context Console -->
    <GenerateContextConsole
      ref="generateConsoleRef"
      v-model:show="ctxGenStore.showConsole"
      :database-id="workspaceStore.currentDatabaseId || ''"
      @complete="handleGenerateComplete"
      @minimize="handleMinimize"
    />
  </div>
</template>
