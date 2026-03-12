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
  NCollapse,
  NCollapseItem,
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

// Context type legend — explains what each tag means and its visual identity
const typeLegend: { type: ContextType; label: string; icon: string; desc: string }[] = [
  { type: 'description',   label: 'Description',   icon: 'i-lucide-file-text',      desc: 'Natural-language description of a table or column\'s purpose and content.' },
  { type: 'example',       label: 'Example',       icon: 'i-lucide-list',           desc: 'Representative sample values that help the LLM understand data patterns.' },
  { type: 'synonym',       label: 'Synonym',       icon: 'i-lucide-languages',      desc: 'Alternative names, abbreviations, or domain aliases for a column.' },
  { type: 'value_mapping', label: 'Value Mapping',  icon: 'i-lucide-arrow-left-right', desc: 'Maps coded values to human-readable labels (e.g. "M" → "Male").' },
  { type: 'business_rule', label: 'Business Rule',  icon: 'i-lucide-shield-check',   desc: 'Domain constraints or logic rules that govern data interpretation.' },
  { type: 'constraint',    label: 'Constraint',    icon: 'i-lucide-alert-triangle', desc: 'Data constraints like NOT NULL, UNIQUE, range limits, or foreign keys.' },
  { type: 'calculation',   label: 'Calculation',   icon: 'i-lucide-calculator',     desc: 'Derived metric formulas (e.g. profit = revenue − cost).' }
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
  // Auto-set recommended iterations based on table count
  const tableCount = workspaceStore.currentDatabase?.tableCount ?? 0
  ctxGenStore.updateRecommendedConfig(tableCount)
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
  <div class="context-manager max-w-[1800px] mx-auto p-8">
    <!-- Toolbar -->
    <div class="flex items-center justify-between mb-8 bg-white p-4 rounded-2xl border border-gray-200/80 shadow-sm">
      <div class="flex items-center gap-4">
        <NInput
          v-model:value="searchKeyword"
          placeholder="Search context content..."
          clearable
          class="w-64"
        >
          <template #prefix>
            <div class="i-lucide-search text-gray-400" />
          </template>
        </NInput>

        <div class="w-px h-6 bg-gray-200"></div>

        <NSelect
          v-model:value="filterTable"
          :options="tableOptions"
          placeholder="All Tables"
          clearable
          class="w-48"
        />

        <NSelect
          v-model:value="filterType"
          :options="typeOptions"
          placeholder="All Types"
          clearable
          class="w-40"
        />
      </div>

      <div class="flex items-center gap-3">
        <button
          class="flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-gray-600 bg-white border border-gray-200 hover:bg-gray-50 hover:text-gray-900 transition-colors shadow-sm"
          @click="workspaceStore.fetchContexts"
        >
          <div class="i-lucide-refresh-cw text-sm" />
          Refresh
        </button>
        <NPopconfirm
          @positive-click="handlePruneAll"
          positive-text="Confirm"
          negative-text="Cancel"
        >
          <template #trigger>
            <button 
              class="flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
              :class="filteredContexts.length === 0 ? 'bg-gray-100 text-gray-400 border border-gray-200' : 'bg-red-50 text-red-600 border border-red-200 hover:bg-red-100 hover:border-red-300'"
              :disabled="filteredContexts.length === 0 || isPruning"
            >
              <div v-if="isPruning" class="i-lucide-loader-2 animate-spin text-sm" />
              <div v-else class="i-lucide-trash-2 text-sm" />
              Clear All
            </button>
          </template>
          <div class="max-w-xs p-1">
            <p class="font-bold text-gray-800 mb-2">Clear all Rich Context?</p>
            <p class="text-[13px] text-gray-500 leading-relaxed">This will delete all table descriptions, column descriptions, business terms and their vector embeddings. This action cannot be undone.</p>
          </div>
        </NPopconfirm>
        <button 
          class="flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold text-white bg-indigo-500 hover:bg-indigo-600 border border-indigo-600 shadow-sm shadow-indigo-500/20 transition-all"
          @click="openGenerateConsole"
        >
          <div class="i-lucide-brain text-sm" />
          AI Generate
        </button>
        <button 
          class="flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold text-white bg-primary-600 hover:bg-primary-700 border border-primary-700 shadow-sm shadow-primary-500/20 transition-all"
          @click="openCreateDialog"
        >
          <div class="i-lucide-plus text-sm" />
          Add Context
        </button>
      </div>
    </div>

    <!-- Context Type Legend -->
    <div class="mb-6 bg-white rounded-2xl border border-gray-200/80 shadow-sm overflow-hidden">
      <NCollapse :default-expanded-names="[]" arrow-placement="left">
        <NCollapseItem name="legend">
          <template #header>
            <div class="flex items-center gap-3 py-2.5 px-2">
              <div class="i-lucide-book-open text-[18px] text-primary-500" />
              <span class="text-[15px] font-extrabold text-gray-800 tracking-wide">Context Type Guide</span>
            </div>
          </template>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 p-6 bg-slate-50/50 border-t border-gray-100">
            <div
              v-for="item in typeLegend"
              :key="item.type"
              class="flex flex-col gap-3 p-5 rounded-xl bg-white border border-gray-200/80 hover:border-primary-300 transition-all shadow-sm hover:shadow-md"
            >
              <div class="flex items-center">
                <span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-bold tracking-widest uppercase shadow-sm" :class="getTypeBadgeClasses(item.type)">
                  <div :class="item.icon" class="text-sm" />
                  {{ item.label }}
                </span>
              </div>
              <p class="text-[13px] text-gray-600 leading-relaxed font-medium">{{ item.desc }}</p>
            </div>
          </div>
        </NCollapseItem>
      </NCollapse>
    </div>

    <!-- Loading -->
    <div v-if="workspaceStore.loadingContexts" class="flex flex-col items-center justify-center py-24 gap-4">
      <NSpin size="large" />
      <span class="text-sm font-medium text-gray-500">Loading rich contexts...</span>
    </div>

    <!-- Empty -->
    <div v-else-if="filteredContexts.length === 0" class="flex flex-col items-center justify-center py-24 bg-white rounded-2xl border border-gray-200/60 border-dashed">
      <div class="w-20 h-20 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-center mb-6 shadow-sm">
        <div class="i-lucide-lightbulb text-4xl text-gray-300" />
      </div>
      <h3 class="text-xl font-bold text-gray-800 tracking-tight mb-2">No Context Found</h3>
      <p class="text-sm text-gray-500 max-w-md text-center mb-6">
        Rich Context helps the LLM understand your schema better. You can manually add context or use AI to generate it automatically.
      </p>
      <div class="flex items-center gap-3">
        <button 
          class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-indigo-500 hover:bg-indigo-600 shadow-sm shadow-indigo-500/20 transition-all"
          @click="openGenerateConsole"
        >
          <div class="i-lucide-brain text-base" />
          Auto Generate
        </button>
        <button 
          class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-gray-700 bg-white hover:bg-gray-50 border border-gray-200 shadow-sm transition-all"
          @click="openCreateDialog"
        >
          <div class="i-lucide-plus text-base" />
          Manual Add
        </button>
      </div>
    </div>

    <!-- Structured Context List -->
    <div v-else class="context-tree">
      <!-- Expand/Collapse All -->
      <div class="flex items-center gap-2 mb-6 bg-white px-4 py-3 rounded-xl border border-gray-200/60 shadow-sm">
        <button 
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors"
          @click="expandAll"
        >
          <div class="i-lucide-unfold-vertical text-gray-500" />
          Expand All
        </button>
        <button 
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors"
          @click="collapseAll"
        >
          <div class="i-lucide-fold-vertical text-gray-500" />
          Collapse All
        </button>
        <div class="ml-auto flex items-center gap-2">
          <span class="px-2.5 py-1 rounded-md bg-indigo-50 text-indigo-700 text-[13px] font-bold border border-indigo-100">
            {{ groupedContexts.length }} Tables
          </span>
          <span class="px-2.5 py-1 rounded-md bg-primary-50 text-primary-700 text-[13px] font-bold border border-primary-100">
            {{ filteredContexts.length }} Contexts
          </span>
        </div>
      </div>

      <!-- Table Groups -->
      <div class="space-y-4">
        <div 
          v-for="group in groupedContexts" 
          :key="group.tableName"
          class="table-group bg-white rounded-2xl border border-gray-200/80 shadow-sm overflow-hidden transition-all duration-300"
        >
          <!-- Table Header - Modern Collapsible -->
          <div 
            class="table-header flex items-center gap-4 px-5 py-4 cursor-pointer hover:bg-slate-50/80 transition-all select-none"
            :class="{ 'border-b border-gray-100 bg-slate-50/50': expandedTables.has(group.tableName) }"
            @click="toggleTable(group.tableName)"
          >
            <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center flex-shrink-0 border border-blue-100/50 shadow-sm">
              <div class="i-lucide-table-2 text-xl text-blue-600" />
            </div>
            
            <span class="font-extrabold text-[17px] text-gray-800 tracking-tight">{{ group.tableName }}</span>
            
            <span class="px-2.5 py-0.5 rounded-full text-[11px] font-bold bg-gray-100 text-gray-500 border border-gray-200/80 uppercase tracking-widest">
              {{ group.columnContexts.length + (group.tableContext ? 1 : 0) }} contexts
            </span>

            <div class="ml-auto flex items-center gap-3">
              <button 
                class="w-8 h-8 rounded-lg bg-white border border-gray-200 hover:border-primary-300 hover:text-primary-600 flex items-center justify-center transition-all shadow-sm"
                @click.stop="openCreateDialogForTable(group.tableName)"
                title="Add Context to Table"
              >
                <div class="i-lucide-plus text-sm" />
              </button>
              
              <!-- Modern expand/collapse indicator -->
              <div 
                class="w-8 h-8 rounded-lg bg-gray-100/80 flex items-center justify-center transition-transform duration-300"
                :class="{ 'rotate-180 bg-blue-50 text-blue-600': expandedTables.has(group.tableName) }"
              >
                <div class="i-lucide-chevron-down text-sm" :class="{ 'text-gray-500': !expandedTables.has(group.tableName) }" />
              </div>
            </div>
          </div>

          <!-- Expanded Content - Modern Panel -->
          <div 
            v-if="expandedTables.has(group.tableName)"
            class="table-content bg-slate-50/50 p-5 space-y-4"
          >
            <!-- Table-level Context -->
            <div 
              v-if="group.tableContext" 
              class="context-item p-5 rounded-xl bg-white border border-amber-200/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group"
            >
              <div class="absolute top-0 left-0 w-1 h-full bg-amber-400"></div>
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center border border-amber-100/50">
                    <div class="i-lucide-file-text text-amber-600 text-sm" />
                  </div>
                  <div class="flex flex-col">
                    <span class="text-[13px] font-bold text-gray-400 uppercase tracking-widest mb-0.5">Table Level</span>
                    <span class="text-[11px] font-bold px-2 py-0.5 rounded-md bg-amber-50 text-amber-700 border border-amber-200/60 w-fit">{{ group.tableContext.type }}</span>
                  </div>
                </div>
                <div class="flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button class="w-8 h-8 rounded-lg bg-gray-50 hover:bg-gray-100 border border-gray-200 flex items-center justify-center transition-colors" @click="openEditDialog(group.tableContext!)" title="Edit">
                    <div class="i-lucide-pencil text-xs text-gray-500" />
                  </button>
                  <button class="w-8 h-8 rounded-lg bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center transition-colors" @click="handleDelete(group.tableContext!)" title="Delete">
                    <div class="i-lucide-trash-2 text-xs text-red-500" />
                  </button>
                </div>
              </div>
              <p class="text-[14px] text-gray-600 leading-relaxed font-medium pl-11">{{ group.tableContext.content }}</p>
            </div>

            <!-- Column Contexts -->
            <div 
              v-for="colCtx in group.columnContexts" 
              :key="colCtx.id"
              class="context-item p-5 rounded-xl bg-white border border-emerald-200/60 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group"
            >
              <div class="absolute top-0 left-0 w-1 h-full bg-emerald-400"></div>
              <div class="flex items-start justify-between mb-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center border border-emerald-100/50">
                    <div class="i-lucide-columns-3 text-emerald-600 text-sm" />
                  </div>
                  <div class="flex flex-col">
                    <div class="flex items-center gap-2 mb-0.5">
                      <span class="text-[13px] font-bold text-gray-400 uppercase tracking-widest">Column Level</span>
                      <span class="text-[14px] font-black text-gray-800 tracking-tight">{{ colCtx.columnName }}</span>
                    </div>
                    <span class="text-[11px] font-bold px-2 py-0.5 rounded-md bg-emerald-50 text-emerald-700 border border-emerald-200/60 w-fit">{{ colCtx.type }}</span>
                  </div>
                </div>
                <div class="flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button class="w-8 h-8 rounded-lg bg-gray-50 hover:bg-gray-100 border border-gray-200 flex items-center justify-center transition-colors" @click="openEditDialog(colCtx)" title="Edit">
                    <div class="i-lucide-pencil text-xs text-gray-500" />
                  </button>
                  <button class="w-8 h-8 rounded-lg bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center transition-colors" @click="handleDelete(colCtx)" title="Delete">
                    <div class="i-lucide-trash-2 text-xs text-red-500" />
                  </button>
                </div>
              </div>
              <p class="text-[14px] text-gray-600 leading-relaxed font-medium pl-11">{{ colCtx.content }}</p>
            </div>

            <!-- Empty columns hint -->
            <div 
              v-if="group.columnContexts.length === 0 && !group.tableContext"
              class="text-[13px] font-medium text-gray-400 py-6 text-center border border-dashed border-gray-200 rounded-xl bg-white"
            >
              No context yet. Click the + button above to add.
            </div>
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
      :table-count="workspaceStore.currentDatabase?.tableCount ?? 0"
      @complete="handleGenerateComplete"
      @minimize="handleMinimize"
    />
  </div>
</template>
