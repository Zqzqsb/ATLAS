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
      prefix: () => h('div', { class: 'i-lucide-table-2 text-blue-500' }),
      suffix: () => h('div', { class: 'flex items-center gap-1' }, [
        table.hasContext && h('div', { class: 'i-lucide-sparkles text-purple-500 text-xs' }),
        h('span', { class: 'text-xs text-gray-400' }, `${table.columns.length} cols`)
      ]),
      isLeaf: false,
      children: table.columns.map(col => ({
        key: `${table.name}.${col.name}`,
        label: col.name,
        prefix: () => h('div', { 
          class: col.isPrimaryKey 
            ? 'i-lucide-key-round text-yellow-500' 
            : col.isForeignKey 
              ? 'i-lucide-link text-green-500'
              : 'i-lucide-columns-3 text-gray-400'
        }),
        suffix: () => h('div', { class: 'flex items-center gap-1 text-xs text-gray-400' }, [
          h('span', {}, col.type),
          col.hasContext && h('div', { class: 'i-lucide-sparkles text-purple-500' })
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
  <div class="schema-browser flex h-[calc(100vh-140px)] bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
    <!-- Left: Table tree -->
    <div class="w-80 border-r border-gray-200 flex flex-col bg-gray-50/50">
      <div class="p-4 border-b border-gray-200">
        <NInput
          v-model:value="searchKeyword"
          placeholder="Search tables or columns..."
          clearable
          class="bg-white"
        >
          <template #prefix>
            <div class="i-lucide-search text-gray-400" />
          </template>
        </NInput>
      </div>

      <!-- Table list -->
      <div class="overflow-auto flex-1">
        <NSpin v-if="workspaceStore.loadingSchema" class="mt-8" />
        <NEmpty v-else-if="treeData.length === 0" description="No data" class="mt-8" />
        <NTree
          v-else
          :data="treeData"
          block-line
          selectable
          expand-on-click
          @update:selected-keys="handleSelect"
        />
      </div>

      <!-- Bottom: ER Diagram toggle -->
      <div class="border-t border-gray-200 p-3">
        <button
          class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors cursor-pointer"
          :class="rightView === 'er'
            ? 'bg-primary-50 text-primary-700 border border-primary-200'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200 border border-transparent'"
          @click="rightView = rightView === 'er' ? 'table' : 'er'"
        >
          <div class="i-lucide-git-branch" />
          ER Diagram
        </button>
      </div>
    </div>

    <!-- Right: Main content area -->
    <div class="flex-1 overflow-auto bg-white">
      <!-- ER Diagram View -->
      <div v-if="rightView === 'er'" class="h-full flex flex-col">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-primary-50 flex items-center justify-center">
              <div class="i-lucide-git-branch text-lg text-primary-600" />
            </div>
            <h2 class="text-xl font-bold text-gray-900">Entity Relationship Diagram</h2>
            <span class="text-sm text-gray-400">
              {{ workspaceStore.schemaCache?.tables?.length || 0 }} tables · {{ workspaceStore.relations.length }} relations
            </span>
          </div>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-100 transition-colors cursor-pointer"
            @click="rightView = 'table'"
          >
            <div class="i-lucide-x" />
            Close
          </button>
        </div>
        <div class="flex-1 overflow-auto p-6">
          <NSpin v-if="!erDiagramSvg && !erError" class="mt-12" />
          <div v-else-if="erError" class="flex flex-col items-center justify-center h-full">
            <div class="text-red-500 text-sm p-6 font-mono bg-red-50 rounded-lg border border-red-200 max-w-xl">
              <div class="font-bold mb-2 flex items-center gap-2">
                <div class="i-lucide-alert-triangle-alt" />
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
      <div v-else class="h-full p-8">
      <template v-if="selectedTable">
        <div class="mb-8 pb-6 border-b border-gray-100">
          <div class="flex items-center gap-3 mb-2">
            <div class="w-10 h-10 rounded-lg bg-primary-50 flex items-center justify-center">
              <div class="i-lucide-table-2 text-xl text-primary-600" />
            </div>
            <h2 class="text-2xl font-bold text-gray-900">
              {{ selectedTable.name }}
            </h2>
          </div>
          
          <p v-if="selectedTable.description" class="text-gray-600 text-lg mb-3">
            {{ selectedTable.description }}
          </p>
          
          <div class="flex items-center gap-4 mt-2 text-sm text-gray-500 font-medium">
            <span class="flex items-center gap-1">
              <div class="i-lucide-columns-3" />
              {{ selectedTable.columns.length }} Columns
            </span>
            <span v-if="selectedTable.rowCount" class="flex items-center gap-1">
              <div class="i-lucide-rows-3" />
              ~{{ selectedTable.rowCount }} Rows
            </span>
            <NTag v-if="selectedTable.hasContext" type="success" size="small" round :bordered="false" class="font-bold">
              <template #icon>
                <div class="i-lucide-sparkles" />
              </template>
              Rich Context Active
            </NTag>
          </div>
        </div>

        <!-- Columns table -->
        <div class="mb-8">
          <h3 class="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
            <div class="i-lucide-list text-primary-600" />
            Schema Definition
          </h3>
          <div class="border border-gray-200 rounded-lg overflow-hidden shadow-sm">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 border-b border-gray-200">
                <tr>
                  <th class="px-4 py-3 text-left font-bold text-gray-600 uppercase tracking-wider text-xs">Column Name</th>
                  <th class="px-4 py-3 text-left font-bold text-gray-600 uppercase tracking-wider text-xs">Type</th>
                  <th class="px-4 py-3 text-left font-bold text-gray-600 uppercase tracking-wider text-xs">Attributes</th>
                  <th class="px-4 py-3 text-left font-bold text-gray-600 uppercase tracking-wider text-xs">Context</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr v-for="col in selectedTable.columns" :key="col.name" class="hover:bg-gray-50 transition-colors">
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-2">
                      <div 
                        :class="col.isPrimaryKey 
                          ? 'i-lucide-key-round text-yellow-500' 
                          : col.isForeignKey 
                            ? 'i-lucide-link text-green-500'
                            : 'i-lucide-columns-3 text-gray-400'"
                      />
                      <span class="font-bold text-gray-700">{{ col.name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-gray-600 font-mono text-xs">{{ col.type }}</td>
                  <td class="px-4 py-3">
                    <div class="flex gap-1">
                      <NTag v-if="col.isPrimaryKey" size="tiny" type="warning" :bordered="false" round>PK</NTag>
                      <NTag v-if="col.isForeignKey" size="tiny" type="success" :bordered="false" round>FK</NTag>
                      <NTag v-if="col.isNullable === false" size="tiny" :bordered="false" round class="bg-gray-200 text-gray-600">NOT NULL</NTag>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div v-if="col.hasContext" class="i-lucide-check-filled text-primary-600" />
                    <div v-else class="i-lucide-minus text-gray-200" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Related contexts -->
        <div v-if="tableContexts.length">
          <h3 class="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
            <div class="i-lucide-lightbulb text-primary-600" />
            Related Context ({{ tableContexts.length }})
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div 
              v-for="ctx in tableContexts" 
              :key="ctx.id"
              class="p-4 bg-gray-50 rounded-lg border border-gray-100 hover:shadow-md transition-shadow hover:border-primary-100 group"
            >
              <div class="flex items-center gap-2 mb-2">
                <NTag size="small" :type="getContextTagType(ctx.type)" :bordered="false" round class="uppercase text-xs font-bold">{{ ctx.type }}</NTag>
                <span v-if="ctx.columnName" class="text-xs font-bold text-gray-500 bg-white px-2 py-0.5 rounded border border-gray-200">
                  {{ ctx.columnName }}
                </span>
              </div>
              <p class="text-sm text-gray-700 font-medium group-hover:text-gray-900">{{ ctx.content }}</p>
            </div>
          </div>
        </div>
      </template>

      <!-- No selection -->
      <div v-else class="h-full flex flex-col items-center justify-center text-gray-400">
        <div class="w-16 h-16 rounded-2xl bg-gray-50 flex items-center justify-center mb-4">
          <div class="i-lucide-table-2 text-3xl opacity-50" />
        </div>
        <p class="font-medium">Select a table to view details</p>
        <button
          class="mt-4 flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-primary-600 bg-primary-50 hover:bg-primary-100 transition-colors cursor-pointer"
          @click="rightView = 'er'"
        >
          <div class="i-lucide-git-branch" />
          View ER Diagram
        </button>
      </div>
      </div>
    </div>
  </div>
</template>
