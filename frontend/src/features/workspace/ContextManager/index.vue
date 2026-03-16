<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
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

// --- View mode ---
type ViewMode = 'flat' | 'cluster'
const viewMode = ref<ViewMode>('flat')
const isCompact = ref(false)

// Forest cluster data for "By Cluster" view
interface ClusterGroup {
  index: number
  tables: string[]
  tableCount: number
  relationCount: number
  coverageRatio: number
  willSkip: boolean
}
const clusterGroups = ref<ClusterGroup[]>([])
const clusterLoading = ref(false)
const clusterLoaded = ref(false)

// Whether this schema is large enough for forest mode
const isForestSchema = computed(() => (workspaceStore.tableNames.length || 0) > 30)

// Load cluster data when switching to cluster view
async function loadClusterData() {
  if (clusterLoaded.value || clusterLoading.value) return
  const lakebaseId = workspaceStore.currentDatabase?.metadata?.lakebaseId
  if (!lakebaseId) return

  clusterLoading.value = true
  try {
    const resp = await fetch(`/api/v1/lakebase/datasources/${lakebaseId}/generate-context/preview`)
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    const data = await resp.json()
    if (data.mode === 'forest_chunked' && Array.isArray(data.clusters)) {
      clusterGroups.value = data.clusters.map((c: any) => ({
        index: c.index,
        tables: c.tables || [],
        tableCount: c.table_count || 0,
        relationCount: c.relation_count || 0,
        coverageRatio: c.coverage_ratio || 0,
        willSkip: c.will_skip || false,
      }))
    }
    clusterLoaded.value = true
  } catch (e: any) {
    message.error('Failed to load cluster info: ' + e.message)
  } finally {
    clusterLoading.value = false
  }
}

watch(viewMode, (mode) => {
  if (mode === 'cluster' && !clusterLoaded.value) {
    loadClusterData()
  }
})

// Reset cluster data when database changes
watch(() => workspaceStore.currentDatabaseId, () => {
  clusterGroups.value = []
  clusterLoaded.value = false
  viewMode.value = 'flat'
})

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
      group.tableContext = ctx
    } else {
      group.columnContexts.push(ctx)
    }
  }
  
  // Sort column contexts by column name
  for (const group of groups.values()) {
    group.columnContexts.sort((a, b) => (a.columnName || '').localeCompare(b.columnName || ''))
  }
  
  return Array.from(groups.values()).sort((a, b) => a.tableName.localeCompare(b.tableName))
})

// Build a lookup map: tableName → TableContextGroup
const groupedContextsMap = computed(() => {
  const map = new Map<string, TableContextGroup>()
  for (const g of groupedContexts.value) {
    map.set(g.tableName, g)
  }
  return map
})

// Cluster-grouped view: clusters with their table groups inside
interface ClusterContextGroup {
  cluster: ClusterGroup
  tables: TableContextGroup[]
  totalContexts: number
  coveragePct: number
}

const clusterContextGroups = computed<ClusterContextGroup[]>(() => {
  if (clusterGroups.value.length === 0) return []
  const map = groupedContextsMap.value
  return clusterGroups.value.map(cl => {
    const tables: TableContextGroup[] = []
    let total = 0
    for (const tName of cl.tables) {
      const g = map.get(tName)
      if (g) {
        tables.push(g)
        total += g.columnContexts.length + (g.tableContext ? 1 : 0)
      } else {
        // Table exists in cluster but has no contexts yet
        tables.push({ tableName: tName, tableContext: null, columnContexts: [] })
      }
    }
    // Sort tables alphabetically within cluster
    tables.sort((a, b) => a.tableName.localeCompare(b.tableName))
    return {
      cluster: cl,
      tables,
      totalContexts: total,
      coveragePct: Math.round(cl.coverageRatio * 100),
    }
  })
})

// Collect type summary pills for a table group
function typeSummary(group: TableContextGroup): { type: ContextType; count: number }[] {
  const counts = new Map<ContextType, number>()
  if (group.tableContext) {
    counts.set(group.tableContext.type, (counts.get(group.tableContext.type) || 0) + 1)
  }
  for (const c of group.columnContexts) {
    counts.set(c.type, (counts.get(c.type) || 0) + 1)
  }
  return Array.from(counts.entries()).map(([type, count]) => ({ type, count }))
}

// Group column contexts by column name for dense display
interface ColumnContextRow {
  columnName: string
  contexts: RichContext[]  // all context entries for this column
}

function groupByColumn(columnContexts: RichContext[]): ColumnContextRow[] {
  const map = new Map<string, RichContext[]>()
  for (const ctx of columnContexts) {
    const col = ctx.columnName || ''
    if (!map.has(col)) map.set(col, [])
    map.get(col)!.push(ctx)
  }
  return Array.from(map.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([columnName, contexts]) => ({ columnName, contexts }))
}

// Short type label for inline display
function shortType(type: ContextType): string {
  const map: Record<ContextType, string> = {
    description: 'desc',
    example: 'ex',
    constraint: 'cstr',
    synonym: 'syn',
    value_mapping: 'map',
    business_rule: 'rule',
    calculation: 'calc'
  }
  return map[type] || type
}

// Type pill colors (short form for inline display)
function typeDotColor(type: ContextType): string {
  const map: Record<ContextType, string> = {
    description: 'bg-blue-400',
    example: 'bg-amber-400',
    constraint: 'bg-red-400',
    synonym: 'bg-purple-400',
    value_mapping: 'bg-pink-400',
    business_rule: 'bg-indigo-400',
    calculation: 'bg-orange-400'
  }
  return map[type] || 'bg-gray-400'
}

// Track expanded tables and clusters
const expandedTables = ref<Set<string>>(new Set())
const expandedClusters = ref<Set<number>>(new Set())

function toggleTable(tableName: string) {
  if (expandedTables.value.has(tableName)) {
    expandedTables.value.delete(tableName)
  } else {
    expandedTables.value.add(tableName)
  }
}

function toggleCluster(index: number) {
  if (expandedClusters.value.has(index)) {
    expandedClusters.value.delete(index)
  } else {
    expandedClusters.value.add(index)
  }
}

function expandAll() {
  groupedContexts.value.forEach(g => expandedTables.value.add(g.tableName))
  clusterContextGroups.value.forEach(cg => expandedClusters.value.add(cg.cluster.index))
}

function collapseAll() {
  expandedTables.value.clear()
  expandedClusters.value.clear()
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
      <!-- Control Bar: Expand/Collapse, View Mode, Stats -->
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

        <!-- Compact toggle -->
        <div class="w-px h-5 bg-gray-200 mx-1"></div>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
          :class="isCompact ? 'bg-indigo-50 text-indigo-700 border border-indigo-200' : 'text-gray-600 hover:bg-gray-100'"
          @click="isCompact = !isCompact"
          title="Toggle compact view"
        >
          <div :class="isCompact ? 'i-lucide-rows-3' : 'i-lucide-rows-4'" class="text-sm" />
          Compact
        </button>

        <!-- View mode toggle (only for forest-scale schemas) -->
        <template v-if="isForestSchema">
          <div class="w-px h-5 bg-gray-200 mx-1"></div>
          <div class="flex items-center bg-gray-100 rounded-lg p-0.5 border border-gray-200/60">
            <button
              class="flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-bold transition-all"
              :class="viewMode === 'flat' ? 'bg-white text-gray-800 shadow-sm border border-gray-200/60' : 'text-gray-500 hover:text-gray-700'"
              @click="viewMode = 'flat'"
            >
              <div class="i-lucide-list text-xs" />
              A–Z
            </button>
            <button
              class="flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-bold transition-all"
              :class="viewMode === 'cluster' ? 'bg-white text-gray-800 shadow-sm border border-gray-200/60' : 'text-gray-500 hover:text-gray-700'"
              @click="viewMode = 'cluster'"
            >
              <div class="i-lucide-network text-xs" />
              By Cluster
            </button>
          </div>
        </template>

        <div class="ml-auto flex items-center gap-2">
          <span class="px-2.5 py-1 rounded-md bg-indigo-50 text-indigo-700 text-[13px] font-bold border border-indigo-100">
            {{ groupedContexts.length }} Tables
          </span>
          <span class="px-2.5 py-1 rounded-md bg-primary-50 text-primary-700 text-[13px] font-bold border border-primary-100">
            {{ filteredContexts.length }} Contexts
          </span>
          <span v-if="viewMode === 'cluster' && clusterLoaded" class="px-2.5 py-1 rounded-md bg-emerald-50 text-emerald-700 text-[13px] font-bold border border-emerald-100">
            {{ clusterGroups.length }} Clusters
          </span>
        </div>
      </div>

      <!-- ===== CLUSTER VIEW ===== -->
      <div v-if="viewMode === 'cluster'" class="space-y-5">
        <!-- Loading clusters -->
        <div v-if="clusterLoading" class="flex items-center justify-center py-12 gap-3">
          <NSpin size="medium" />
          <span class="text-sm font-medium text-gray-500">Analyzing schema clusters...</span>
        </div>

        <!-- Cluster groups -->
        <div
          v-else
          v-for="cg in clusterContextGroups"
          :key="cg.cluster.index"
          class="cluster-group rounded-2xl border border-slate-200 shadow-sm overflow-hidden bg-white"
        >
          <!-- Cluster Header -->
          <div
            class="flex items-center gap-3 px-5 py-3 cursor-pointer hover:bg-slate-50/80 transition-all select-none"
            :class="{ 'border-b border-slate-200 bg-slate-50/50': expandedClusters.has(cg.cluster.index) }"
            @click="toggleCluster(cg.cluster.index)"
          >
            <div class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 border"
              :class="cg.coveragePct >= 90 ? 'bg-emerald-50 border-emerald-200 text-emerald-700' : 'bg-amber-50 border-amber-200 text-amber-700'"
            >
              <span class="text-xs font-black">#{{ cg.cluster.index + 1 }}</span>
            </div>

            <div class="flex flex-col min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-bold text-sm text-gray-800">Cluster {{ cg.cluster.index + 1 }}</span>
                <span class="text-[11px] font-semibold text-gray-500">{{ cg.cluster.tableCount }} tables · {{ cg.cluster.relationCount }} FK</span>
              </div>
              <div class="flex items-center gap-2 mt-0.5">
                <!-- Mini coverage bar -->
                <div class="w-20 h-1.5 rounded-full bg-gray-200 overflow-hidden">
                  <div class="h-full rounded-full transition-all" :style="{ width: cg.coveragePct + '%' }"
                    :class="cg.coveragePct >= 90 ? 'bg-emerald-500' : cg.coveragePct >= 50 ? 'bg-amber-500' : 'bg-red-400'"
                  />
                </div>
                <span class="text-[10px] font-bold" :class="cg.coveragePct >= 90 ? 'text-emerald-600' : 'text-gray-500'">{{ cg.coveragePct }}%</span>
              </div>
            </div>

            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider"
              :class="cg.coveragePct >= 90 ? 'bg-emerald-100 text-emerald-700 border border-emerald-200' : 'bg-amber-100 text-amber-700 border border-amber-200'"
            >
              {{ cg.totalContexts }} ctx
            </span>

            <!-- Compact table name preview -->
            <div class="flex-1 min-w-0 hidden lg:block">
              <span class="text-[11px] text-gray-400 font-medium truncate block">
                {{ cg.cluster.tables.slice(0, 4).join(', ') }}{{ cg.cluster.tables.length > 4 ? ` … +${cg.cluster.tables.length - 4}` : '' }}
              </span>
            </div>

            <div class="ml-auto flex-shrink-0 w-6 h-6 rounded-md bg-gray-100/80 flex items-center justify-center transition-transform duration-300"
              :class="{ 'rotate-180 bg-indigo-50 text-indigo-600': expandedClusters.has(cg.cluster.index) }"
            >
              <div class="i-lucide-chevron-down text-xs" :class="{ 'text-gray-500': !expandedClusters.has(cg.cluster.index) }" />
            </div>
          </div>

          <!-- Expanded: tables within this cluster -->
          <div v-if="expandedClusters.has(cg.cluster.index)" class="bg-slate-50/30">
            <div
              v-for="group in cg.tables"
              :key="group.tableName"
              class="border-b border-slate-100 last:border-b-0"
            >
              <!-- Table row (inside cluster) -->
              <div
                class="flex items-center gap-3 px-5 pl-10 py-2.5 cursor-pointer hover:bg-white/60 transition-all select-none"
                :class="{ 'bg-white/40': expandedTables.has(group.tableName) }"
                @click="toggleTable(group.tableName)"
              >
                <div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center flex-shrink-0 border border-blue-100/50">
                  <div class="i-lucide-table-2 text-sm text-blue-600" />
                </div>
                <span class="font-bold text-[14px] text-gray-800">{{ group.tableName }}</span>

                <!-- Type summary pills -->
                <div class="flex items-center gap-1">
                  <template v-for="ts in typeSummary(group)" :key="ts.type">
                    <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[9px] font-bold" :class="getTypeBadgeClasses(ts.type)">
                      {{ ts.count }}
                    </span>
                  </template>
                  <span v-if="!group.tableContext && group.columnContexts.length === 0"
                    class="text-[10px] text-gray-400 italic"
                  >no context</span>
                </div>

                <div class="ml-auto flex items-center gap-2">
                  <button 
                    class="w-6 h-6 rounded-md bg-white border border-gray-200 hover:border-primary-300 flex items-center justify-center transition-all"
                    @click.stop="openCreateDialogForTable(group.tableName)"
                    title="Add Context"
                  >
                    <div class="i-lucide-plus text-xs text-gray-500" />
                  </button>
                  <div class="w-5 h-5 rounded-md flex items-center justify-center transition-transform duration-300"
                    :class="{ 'rotate-180 text-blue-600': expandedTables.has(group.tableName) }"
                  >
                    <div class="i-lucide-chevron-down text-xs text-gray-400" />
                  </div>
                </div>
              </div>

              <!-- Expanded table content — dense grouped layout -->
              <div v-if="expandedTables.has(group.tableName)" class="ml-10 border-t border-gray-100">
                <!-- Table-level -->
                <div v-if="group.tableContext" class="flex items-center gap-2 px-4 py-1.5 border-b border-gray-50 bg-amber-50/30 group">
                  <div class="w-0.5 h-4 rounded-full bg-amber-400 flex-shrink-0"></div>
                  <span class="text-[9px] font-extrabold text-amber-700 bg-amber-100 px-1.5 py-0.5 rounded flex-shrink-0 uppercase">TBL</span>
                  <span class="text-[9px] font-bold px-1 py-0.5 rounded bg-amber-50 text-amber-600 border border-amber-200/50 flex-shrink-0">{{ shortType(group.tableContext.type) }}</span>
                  <p class="text-[12px] text-gray-600 font-medium flex-1 truncate">{{ group.tableContext.content }}</p>
                  <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                    <button class="w-5 h-5 rounded bg-white/80 hover:bg-white border border-gray-200 flex items-center justify-center" @click="openEditDialog(group.tableContext!)" title="Edit">
                      <div class="i-lucide-pencil text-[9px] text-gray-500" />
                    </button>
                    <button class="w-5 h-5 rounded bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(group.tableContext!)" title="Delete">
                      <div class="i-lucide-trash-2 text-[9px] text-red-500" />
                    </button>
                  </div>
                </div>
                <!-- Columns grouped -->
                <div v-for="colGroup in groupByColumn(group.columnContexts)" :key="colGroup.columnName" class="flex items-start border-b border-gray-50 last:border-b-0">
                  <div class="flex-shrink-0 w-32 min-w-32 border-r border-gray-50 px-3 py-1.5 bg-white/40">
                    <div class="flex items-center gap-1">
                      <div class="w-0.5 h-3 rounded-full bg-emerald-400 flex-shrink-0"></div>
                      <span class="text-[11px] font-bold text-gray-700 truncate" :title="colGroup.columnName">{{ colGroup.columnName }}</span>
                    </div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <div v-for="ctx in colGroup.contexts" :key="ctx.id" class="flex items-center gap-1.5 px-2 py-1 border-b border-gray-50/50 last:border-b-0 hover:bg-white/60 group">
                      <span class="text-[9px] font-bold px-1 py-0.5 rounded flex-shrink-0" :class="getTypeBadgeClasses(ctx.type)">{{ shortType(ctx.type) }}</span>
                      <p class="text-[12px] text-gray-600 font-medium flex-1 truncate">{{ ctx.content }}</p>
                      <div class="flex gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                        <button class="w-4 h-4 rounded bg-gray-50 hover:bg-gray-100 border border-gray-200 flex items-center justify-center" @click="openEditDialog(ctx)" title="Edit">
                          <div class="i-lucide-pencil text-[8px] text-gray-500" />
                        </button>
                        <button class="w-4 h-4 rounded bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(ctx)" title="Delete">
                          <div class="i-lucide-trash-2 text-[8px] text-red-500" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="group.columnContexts.length === 0 && !group.tableContext" class="text-[10px] font-medium text-gray-400 py-2 text-center">No context.</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ===== FLAT VIEW (A-Z) ===== -->
      <div v-else>
        <!-- Compact flat view: dense table rows -->
        <div v-if="isCompact" class="space-y-1.5">
          <div 
            v-for="group in groupedContexts" 
            :key="group.tableName"
            class="bg-white rounded-xl border border-gray-200/80 shadow-sm overflow-hidden"
          >
            <!-- Compact table row -->
            <div 
              class="flex items-center gap-3 px-4 py-2.5 cursor-pointer hover:bg-slate-50/80 transition-all select-none"
              :class="{ 'border-b border-gray-100 bg-slate-50/50': expandedTables.has(group.tableName) }"
              @click="toggleTable(group.tableName)"
            >
              <div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center flex-shrink-0 border border-blue-100/50">
                <div class="i-lucide-table-2 text-sm text-blue-600" />
              </div>
              <span class="font-bold text-[14px] text-gray-800">{{ group.tableName }}</span>

              <!-- Type summary pills -->
              <div class="flex items-center gap-1">
                <template v-for="ts in typeSummary(group)" :key="ts.type">
                  <span class="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[9px] font-bold" :class="getTypeBadgeClasses(ts.type)">
                    {{ ts.count }}
                  </span>
                </template>
              </div>

              <div class="ml-auto flex items-center gap-2">
                <button 
                  class="w-6 h-6 rounded-md bg-white border border-gray-200 hover:border-primary-300 flex items-center justify-center transition-all"
                  @click.stop="openCreateDialogForTable(group.tableName)"
                  title="Add Context"
                >
                  <div class="i-lucide-plus text-xs text-gray-500" />
                </button>
                <div class="w-5 h-5 rounded-md flex items-center justify-center transition-transform duration-300"
                  :class="{ 'rotate-180 text-blue-600': expandedTables.has(group.tableName) }"
                >
                  <div class="i-lucide-chevron-down text-xs text-gray-400" />
                </div>
              </div>
            </div>

            <!-- Expanded compact content — reuses dense grouped layout -->
            <div v-if="expandedTables.has(group.tableName)" class="border-t border-gray-100">
              <!-- Table-level -->
              <div v-if="group.tableContext" class="flex items-center gap-2 px-4 py-1.5 border-b border-gray-50 bg-amber-50/30 group">
                <div class="w-0.5 h-4 rounded-full bg-amber-400 flex-shrink-0"></div>
                <span class="text-[9px] font-extrabold text-amber-700 bg-amber-100 px-1.5 py-0.5 rounded flex-shrink-0 uppercase">TBL</span>
                <span class="text-[9px] font-bold px-1 py-0.5 rounded bg-amber-50 text-amber-600 border border-amber-200/50 flex-shrink-0">{{ shortType(group.tableContext.type) }}</span>
                <p class="text-[12px] text-gray-600 font-medium flex-1 truncate">{{ group.tableContext.content }}</p>
                <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                  <button class="w-5 h-5 rounded bg-white/80 hover:bg-white border border-gray-200 flex items-center justify-center" @click="openEditDialog(group.tableContext!)" title="Edit">
                    <div class="i-lucide-pencil text-[9px] text-gray-500" />
                  </button>
                  <button class="w-5 h-5 rounded bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(group.tableContext!)" title="Delete">
                    <div class="i-lucide-trash-2 text-[9px] text-red-500" />
                  </button>
                </div>
              </div>
              <!-- Columns grouped -->
              <div v-for="colGroup in groupByColumn(group.columnContexts)" :key="colGroup.columnName" class="flex items-start border-b border-gray-50 last:border-b-0">
                <div class="flex-shrink-0 w-28 min-w-28 border-r border-gray-50 px-3 py-1.5">
                  <div class="flex items-center gap-1">
                    <div class="w-0.5 h-3 rounded-full bg-emerald-400 flex-shrink-0"></div>
                    <span class="text-[11px] font-bold text-gray-700 truncate" :title="colGroup.columnName">{{ colGroup.columnName }}</span>
                  </div>
                </div>
                <div class="flex-1 min-w-0">
                  <div v-for="ctx in colGroup.contexts" :key="ctx.id" class="flex items-center gap-1.5 px-2 py-1 border-b border-gray-50/50 last:border-b-0 hover:bg-white/60 group">
                    <span class="text-[9px] font-bold px-1 py-0.5 rounded flex-shrink-0" :class="getTypeBadgeClasses(ctx.type)">{{ shortType(ctx.type) }}</span>
                    <p class="text-[11px] text-gray-600 font-medium flex-1 truncate">{{ ctx.content }}</p>
                    <div class="flex gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                      <button class="w-4 h-4 rounded bg-gray-50 hover:bg-gray-100 border border-gray-200 flex items-center justify-center" @click="openEditDialog(ctx)" title="Edit">
                        <div class="i-lucide-pencil text-[8px] text-gray-500" />
                      </button>
                      <button class="w-4 h-4 rounded bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(ctx)" title="Delete">
                        <div class="i-lucide-trash-2 text-[8px] text-red-500" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="group.columnContexts.length === 0 && !group.tableContext" class="text-[10px] font-medium text-gray-400 py-2 text-center">No context.</div>
            </div>
          </div>
        </div>

        <!-- Full flat view (original style) -->
        <div v-else class="space-y-4">
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
              
              <!-- Type summary pills (replace the plain "14 contexts" badge) -->
              <div class="flex items-center gap-1">
                <template v-for="ts in typeSummary(group)" :key="ts.type">
                  <span class="inline-flex items-center gap-0.5 px-2 py-0.5 rounded-full text-[10px] font-bold" :class="getTypeBadgeClasses(ts.type)">
                    {{ ts.count }}
                  </span>
                </template>
              </div>

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

            <!-- Expanded Content — Dense Table Layout -->
            <div 
              v-if="expandedTables.has(group.tableName)"
              class="table-content bg-slate-50/30 border-t border-gray-100"
            >
              <!-- Table-level Context (compact header row) -->
              <div v-if="group.tableContext" class="flex items-center gap-3 px-5 py-2.5 border-b border-gray-100 bg-amber-50/40 group">
                <div class="w-0.5 h-5 rounded-full bg-amber-400 flex-shrink-0"></div>
                <span class="text-[11px] font-extrabold text-amber-700 bg-amber-100 px-2 py-0.5 rounded flex-shrink-0 uppercase tracking-wide">TABLE</span>
                <span class="text-[11px] font-bold px-1.5 py-0.5 rounded bg-amber-50 text-amber-600 border border-amber-200/50 flex-shrink-0">{{ shortType(group.tableContext.type) }}</span>
                <p class="text-[13px] text-gray-700 font-medium leading-snug flex-1 truncate">{{ group.tableContext.content }}</p>
                <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                  <button class="w-6 h-6 rounded bg-white/80 hover:bg-white border border-gray-200 flex items-center justify-center" @click="openEditDialog(group.tableContext!)" title="Edit">
                    <div class="i-lucide-pencil text-[10px] text-gray-500" />
                  </button>
                  <button class="w-6 h-6 rounded bg-red-50/80 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(group.tableContext!)" title="Delete">
                    <div class="i-lucide-trash-2 text-[10px] text-red-500" />
                  </button>
                </div>
              </div>

              <!-- Column Contexts — grouped by column, each column = one dense block -->
              <div
                v-for="colGroup in groupByColumn(group.columnContexts)"
                :key="colGroup.columnName"
                class="border-b border-gray-100 last:border-b-0"
              >
                <!-- Column row: column name on left, context entries stacked right -->
                <div class="flex items-start gap-0">
                  <!-- Column name sidebar -->
                  <div class="flex-shrink-0 w-36 min-w-36 border-r border-gray-100 px-4 py-2.5 bg-white/60">
                    <div class="flex items-center gap-1.5">
                      <div class="w-1 h-4 rounded-full bg-emerald-400 flex-shrink-0"></div>
                      <span class="text-[13px] font-bold text-gray-800 truncate" :title="colGroup.columnName">{{ colGroup.columnName }}</span>
                    </div>
                  </div>
                  <!-- Context entries for this column -->
                  <div class="flex-1 min-w-0">
                    <div
                      v-for="ctx in colGroup.contexts"
                      :key="ctx.id"
                      class="flex items-center gap-2 px-3 py-2 border-b border-gray-50 last:border-b-0 hover:bg-white/80 transition-colors group"
                    >
                      <span class="text-[10px] font-bold px-1.5 py-0.5 rounded flex-shrink-0" :class="getTypeBadgeClasses(ctx.type)">{{ shortType(ctx.type) }}</span>
                      <p class="text-[13px] text-gray-600 font-medium leading-snug flex-1 truncate">{{ ctx.content }}</p>
                      <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                        <button class="w-5 h-5 rounded bg-gray-50 hover:bg-gray-100 border border-gray-200 flex items-center justify-center" @click="openEditDialog(ctx)" title="Edit">
                          <div class="i-lucide-pencil text-[9px] text-gray-500" />
                        </button>
                        <button class="w-5 h-5 rounded bg-red-50 hover:bg-red-100 border border-red-200 flex items-center justify-center" @click="handleDelete(ctx)" title="Delete">
                          <div class="i-lucide-trash-2 text-[9px] text-red-500" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Empty hint -->
              <div 
                v-if="group.columnContexts.length === 0 && !group.tableContext"
                class="text-[12px] font-medium text-gray-400 py-4 text-center"
              >
                No context yet. Click + to add.
              </div>
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
