<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { NTree, NInput, NEmpty, NSpin, NTag } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import type { TableInfo, ColumnInfo } from '@/types'
import mermaid from 'mermaid'

const workspaceStore = useWorkspaceStore()

const searchKeyword = ref('')
const selectedTable = ref<TableInfo | null>(null)
const rightView = ref<'table' | 'er'>('table')

// Initialize mermaid
mermaid.initialize({
  startOnLoad: false,
  theme: 'dark',
  securityLevel: 'loose',
  er: {
    useMaxWidth: true,
    layoutDirection: 'TB'
  }
})

// Generate Mermaid ER diagram code
const erDiagramCode = computed(() => {
  if (!workspaceStore.schemaCache?.tables?.length) return ''
  
  let code = 'erDiagram\n'
  
  // Add tables with their columns
  for (const table of workspaceStore.schemaCache.tables) {
    // Sanitize table name for mermaid (no spaces/special chars)
    const safeName = table.name.replace(/[^a-zA-Z0-9_]/g, '_')
    code += `    ${safeName} {\n`
    for (const col of table.columns) {
      const pkMark = col.isPrimaryKey ? 'PK' : col.isForeignKey ? 'FK' : ''
      // Clean data type: remove parentheses, commas, quotes and special chars for Mermaid compatibility
      const colType = (col.type || 'VARCHAR')
        .replace(/[(),"']/g, '')  // Remove parens, commas, quotes
        .replace(/\s+/g, '_')     // Replace spaces with underscore
        .replace(/[^a-zA-Z0-9_]/g, '') // Remove any remaining special chars
        .substring(0, 20) || 'unknown'
      // Sanitize column name
      const safeCol = col.name.replace(/[^a-zA-Z0-9_]/g, '_')
      code += `        ${colType} ${safeCol}${pkMark ? ' ' + pkMark : ''}\n`
    }
    code += `    }\n`
  }
  
  // Add relationships
  for (const rel of workspaceStore.relations) {
    const safeFrom = rel.fromTable.replace(/[^a-zA-Z0-9_]/g, '_')
    const safeTo = rel.toTable.replace(/[^a-zA-Z0-9_]/g, '_')
    const safeLabel = rel.fromColumn.replace(/"/g, '')
    const relSymbol = rel.relationType === 'many_to_one' ? '}o--||' : 
                      rel.relationType === 'one_to_many' ? '||--o{' :
                      rel.relationType === 'many_to_many' ? '}o--o{' : '||--||'
    code += `    ${safeFrom} ${relSymbol} ${safeTo} : "${safeLabel}"\n`
  }
  
  return code
})

// Render ER diagram
const erDiagramSvg = ref('')
const erError = ref('')

async function renderERDiagram() {
  if (!erDiagramCode.value) {
    erDiagramSvg.value = ''
    return
  }
  
  try {
    erError.value = ''
    const { svg } = await mermaid.render('er-diagram', erDiagramCode.value)
    erDiagramSvg.value = svg
  } catch (e: any) {
    erError.value = e.message
    console.error('Mermaid render error:', e)
  }
}

watch(erDiagramCode, () => {
  if (rightView.value === 'er') {
    nextTick(renderERDiagram)
  }
})

watch(rightView, (view) => {
  if (view === 'er' && !erDiagramSvg.value) {
    nextTick(renderERDiagram)
  }
})

// Build tree data for NTree
const treeData = computed(() => {
  if (!workspaceStore.schemaCache?.tables) return []

  return workspaceStore.schemaCache.tables
    .filter(table => {
      if (!searchKeyword.value) return true
      return table.name.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
        table.columns.some(col => col.name.toLowerCase().includes(searchKeyword.value.toLowerCase()))
    })
    .map(table => ({
      key: table.name,
      label: table.name,
      prefix: () => h('div', { class: 'i-atlas-table-2 text-blue-500' }),
      suffix: () => h('div', { class: 'flex items-center gap-1' }, [
        table.hasContext && h('div', { class: 'i-atlas-sparkles text-purple-500 text-xs' }),
        h('span', { class: 'text-xs text-gray-400' }, `${table.columns.length} cols`)
      ]),
      isLeaf: false,
      children: table.columns.map(col => ({
        key: `${table.name}.${col.name}`,
        label: col.name,
        prefix: () => h('div', { 
          class: col.isPrimaryKey 
            ? 'i-atlas-key-round text-yellow-500' 
            : col.isForeignKey 
              ? 'i-atlas-link text-green-500'
              : 'i-atlas-columns-3 text-gray-400'
        }),
        suffix: () => h('div', { class: 'flex items-center gap-1 text-xs text-gray-400' }, [
          h('span', {}, col.type),
          col.hasContext && h('div', { class: 'i-atlas-sparkles text-purple-500' })
        ]),
        isLeaf: true
      }))
    }))
})

import { h } from 'vue'

function handleSelect(keys: string[]) {
  const key = keys[0]
  if (key && !key.includes('.')) {
    // Table selected
    selectedTable.value = workspaceStore.schemaCache?.tables.find(t => t.name === key) || null
  }
}

// Get contexts for selected table
const tableContexts = computed(() => {
  if (!selectedTable.value) return []
  return workspaceStore.contextsByTable[selectedTable.value.name] || []
})

// Return NTag type for each context type
function getContextTagType(type: string): 'info' | 'success' | 'warning' | 'error' | 'primary' | 'default' {
  const map: Record<string, 'info' | 'success' | 'warning' | 'error' | 'primary' | 'default'> = {
    description: 'info',
    example: 'warning',
    constraint: 'error',
    synonym: 'primary',
    value_mapping: 'error',
    business_rule: 'info',
    calculation: 'warning'
  }
  return map[type] || 'default'
}
</script>

<template>
  <div class="schema-browser flex h-[calc(100vh-140px)] bg-white rounded-2xl shadow-sm border border-gray-200/60 overflow-hidden">
    <!-- Left: Table tree -->
    <div class="w-80 border-r border-gray-100 flex flex-col bg-slate-50/40">
      <div class="p-4 border-b border-gray-100 bg-white/50 backdrop-blur-sm">
        <NInput
          v-model:value="searchKeyword"
          placeholder="Search tables or columns..."
          clearable
          class="bg-white shadow-sm hover:border-primary-300 focus-within:border-primary-400 focus-within:ring-1 focus-within:ring-primary-100"
        >
          <template #prefix>
            <div class="i-atlas-search text-gray-400" />
          </template>
        </NInput>
      </div>

      <!-- Table list -->
      <div class="overflow-auto flex-1 p-2">
        <div v-if="workspaceStore.loadingSchema" class="flex flex-col items-center justify-center h-full gap-3 text-gray-400">
          <NSpin size="large" />
          <span class="text-sm font-medium">Loading schema...</span>
        </div>
        <div v-else-if="treeData.length === 0" class="flex flex-col items-center justify-center h-full gap-2 text-gray-400">
          <div class="i-atlas-database text-3xl opacity-20 mb-2"></div>
          <span class="text-sm font-medium">No tables found</span>
        </div>
        <NTree
          v-else
          :data="treeData"
          block-line
          selectable
          expand-on-click
          class="bg-transparent"
          @update:selected-keys="handleSelect"
        />
      </div>

      <!-- Bottom: ER Diagram toggle -->
      <div class="border-t border-gray-100 p-3 bg-white/50">
        <button
          class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer"
          :class="rightView === 'er'
            ? 'bg-primary-50 text-primary-700 border border-primary-200'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200 border border-transparent'"
          @click="rightView = rightView === 'er' ? 'table' : 'er'"
        >
          <div class="i-atlas-git-branch" />
          ER Diagram
        </button>
      </div>
    </div>

    <!-- Right: Main content area -->
    <div class="flex-1 overflow-auto bg-white/50">
      <!-- ER Diagram View -->
      <div v-if="rightView === 'er'" class="h-full flex flex-col bg-white">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 bg-slate-50/50 backdrop-blur-sm sticky top-0 z-10">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-xl bg-primary-50/80 flex items-center justify-center border border-primary-100/50 shadow-sm">
              <div class="i-atlas-git-branch text-[17px] text-primary-600" />
            </div>
            <h2 class="text-[17px] font-bold text-gray-800 tracking-tight">Entity Relationship Diagram</h2>
            <div class="w-1.5 h-1.5 rounded-full bg-gray-200 ml-1"></div>
            <span class="text-[13px] font-medium text-gray-500">
              {{ workspaceStore.schemaCache?.tables?.length || 0 }} tables · {{ workspaceStore.relations.length }} relations
            </span>
          </div>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[13px] font-medium text-gray-500 border border-gray-200 hover:border-gray-300 hover:text-gray-700 hover:bg-white shadow-sm transition-all cursor-pointer bg-gray-50/80"
            @click="rightView = 'table'"
          >
            <div class="i-atlas-x text-sm" />
            Close
          </button>
        </div>
        <div class="flex-1 overflow-auto p-6">
          <NSpin v-if="!erDiagramSvg && !erError" class="mt-12" />
          <div v-else-if="erError" class="flex flex-col items-center justify-center h-full">
            <div class="text-red-500 text-sm p-6 font-mono bg-red-50 rounded-lg border border-red-200 max-w-xl">
              <div class="font-bold mb-2 flex items-center gap-2">
                <div class="i-atlas-alert-triangle-alt" />
                Diagram Parse Error
              </div>
              {{ erError }}
            </div>
          </div>
          <div
            v-else
            class="er-diagram-container flex items-center justify-center min-h-full"
            v-html="erDiagramSvg"
          />
        </div>
      </div>

      <!-- Table Detail View -->
      <div v-else class="h-full">
        <div v-if="!selectedTable" class="flex flex-col items-center justify-center h-full text-center p-8">
          <div class="w-20 h-20 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-center mb-6 shadow-sm">
            <div class="i-atlas-layout-grid text-4xl text-gray-300" />
          </div>
          <h3 class="text-xl font-bold text-gray-800 tracking-tight mb-2">Schema Browser</h3>
          <p class="text-sm text-gray-500 max-w-sm">
            Select a table from the left sidebar to view its schema definition, columns, and attached Rich Context.
          </p>
        </div>
      <template v-else>
        <div class="p-8">
          <div class="mb-8 pb-6 border-b border-gray-150">
            <div class="flex items-center gap-3.5 mb-2.5">
              <div class="w-12 h-12 rounded-xl bg-primary-50/80 flex items-center justify-center border border-primary-100 shadow-sm">
                <div class="i-atlas-table-2 text-[22px] text-primary-600" />
              </div>
              <h2 class="text-[26px] font-extrabold text-gray-900 tracking-tight">
                {{ selectedTable.name }}
              </h2>
            </div>
            
            <p v-if="selectedTable.description" class="text-gray-500 text-sm leading-relaxed mb-4 max-w-3xl">
              {{ selectedTable.description }}
            </p>
            
            <div class="flex items-center gap-4 text-sm text-gray-500 font-medium bg-gray-50/50 w-fit px-3 py-1.5 rounded-lg border border-gray-100">
              <span class="flex items-center gap-1.5">
                <div class="i-atlas-columns-3 text-gray-400" />
                {{ selectedTable.columns.length }} Columns
              </span>
              <div class="w-px h-3.5 bg-gray-200"></div>
              <span v-if="selectedTable.rowCount" class="flex items-center gap-1.5">
                <div class="i-atlas-rows-3 text-gray-400" />
                ~{{ selectedTable.rowCount }} Rows
              </span>
              <template v-if="selectedTable.hasContext">
                <div class="w-px h-3.5 bg-gray-200"></div>
                <NTag type="success" size="small" round :bordered="false" class="font-bold px-2.5">
                  <template #icon>
                    <div class="i-atlas-sparkles" />
                  </template>
                  Rich Context Active
                </NTag>
              </template>
            </div>
          </div>

          <!-- Columns table -->
          <div class="mb-10">
            <h3 class="text-[17px] font-bold text-gray-800 tracking-tight mb-4 flex items-center gap-2">
              <div class="i-atlas-list text-primary-500" />
              Schema Definition
            </h3>
            <div class="border border-gray-200/80 rounded-xl overflow-hidden shadow-sm bg-white">
              <table class="w-full text-sm">
                <thead class="bg-slate-50/80 border-b border-gray-200/80">
                  <tr>
                    <th class="px-5 py-3.5 text-left font-bold text-gray-500 uppercase tracking-widest text-[11px]">Column Name</th>
                    <th class="px-5 py-3.5 text-left font-bold text-gray-500 uppercase tracking-widest text-[11px]">Type</th>
                    <th class="px-5 py-3.5 text-left font-bold text-gray-500 uppercase tracking-widest text-[11px]">Attributes</th>
                    <th class="px-5 py-3.5 text-center font-bold text-gray-500 uppercase tracking-widest text-[11px]">Context</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100">
                  <tr v-for="col in selectedTable.columns" :key="col.name" class="hover:bg-slate-50/50 transition-colors group">
                    <td class="px-5 py-3.5">
                      <div class="flex items-center gap-2.5">
                        <div 
                          class="opacity-70 group-hover:opacity-100 transition-opacity"
                          :class="col.isPrimaryKey 
                            ? 'i-atlas-key-round text-amber-500' 
                            : col.isForeignKey 
                              ? 'i-atlas-link text-emerald-500'
                              : 'i-atlas-columns-3 text-gray-300'"
                        />
                        <span class="font-bold text-gray-800">{{ col.name }}</span>
                      </div>
                    </td>
                    <td class="px-5 py-3.5 text-gray-500 font-mono text-xs">{{ col.type }}</td>
                    <td class="px-5 py-3.5">
                      <div class="flex gap-1.5 flex-wrap">
                        <span v-if="col.isPrimaryKey" class="px-2 py-0.5 rounded text-[10px] font-bold bg-amber-50 text-amber-600 border border-amber-100 uppercase tracking-wide">PK</span>
                        <span v-if="col.isForeignKey" class="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-50 text-emerald-600 border border-emerald-100 uppercase tracking-wide">FK</span>
                        <span v-if="col.isNullable === false" class="px-2 py-0.5 rounded text-[10px] font-bold bg-gray-100 text-gray-500 border border-gray-200 uppercase tracking-wide">NOT NULL</span>
                      </div>
                    </td>
                    <td class="px-5 py-3.5 text-center">
                      <div v-if="col.hasContext" class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-primary-50 text-primary-600">
                        <div class="i-atlas-check text-sm" />
                      </div>
                      <div v-else class="i-atlas-minus text-gray-200 mx-auto" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

        <!-- Related contexts -->
        <div v-if="tableContexts.length">
          <h3 class="text-[17px] font-bold text-gray-800 tracking-tight mb-4 flex items-center gap-2">
            <div class="i-atlas-lightbulb text-amber-500" />
            Related Context ({{ tableContexts.length }})
          </h3>
          <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
            <div 
              v-for="ctx in tableContexts" 
              :key="ctx.id"
              class="p-4 bg-white rounded-xl border border-gray-150 hover:shadow-md transition-shadow hover:border-primary-200 group"
            >
              <div class="flex items-center gap-2 mb-2.5">
                <NTag size="small" :type="getContextTagType(ctx.type)" :bordered="false" round class="uppercase tracking-widest text-[10px] font-bold">{{ ctx.type }}</NTag>
                <span v-if="ctx.columnName" class="text-[11px] font-bold text-gray-500 bg-gray-50 px-2 py-0.5 rounded border border-gray-200/80">
                  {{ ctx.columnName }}
                </span>
              </div>
              <p class="text-[13px] text-gray-600 font-medium leading-relaxed group-hover:text-gray-800">{{ ctx.content }}</p>
            </div>
          </div>
        </div>
      </div>
      </template>

      <!-- ER diagram CTA at bottom when nothing selected -->
      <div v-if="!selectedTable" class="flex items-center justify-center pb-12">
        <button
          class="flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-primary-600 bg-primary-50 hover:bg-primary-100 hover:shadow-sm border border-primary-100 transition-all cursor-pointer"
          @click="rightView = 'er'"
        >
          <div class="i-atlas-git-branch text-base" />
          View ER Diagram
        </button>
      </div>
      </div>
    </div>
  </div>
</template>
