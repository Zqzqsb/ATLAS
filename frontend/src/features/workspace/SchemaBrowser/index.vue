<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { NTree, NInput, NEmpty, NSpin, NTag, NScrollbar, NTabs, NTabPane } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import type { TableInfo, ColumnInfo } from '@/types'
import mermaid from 'mermaid'

const workspaceStore = useWorkspaceStore()

const searchKeyword = ref('')
const selectedTable = ref<TableInfo | null>(null)
const activePane = ref('tree')

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
    code += `    ${table.name} {\n`
    for (const col of table.columns.slice(0, 8)) { // Limit columns for readability
      const pkMark = col.isPrimaryKey ? 'PK' : col.isForeignKey ? 'FK' : ''
      const colType = (col.type || 'VARCHAR').replace(/[()]/g, '').substring(0, 10)
      code += `        ${colType} ${col.name}${pkMark ? ' ' + pkMark : ''}\n`
    }
    if (table.columns.length > 8) {
      code += `        ... ${table.columns.length - 8}_more\n`
    }
    code += `    }\n`
  }
  
  // Add relationships
  for (const rel of workspaceStore.relations) {
    const relSymbol = rel.relationType === 'many_to_one' ? '}o--||' : 
                      rel.relationType === 'one_to_many' ? '||--o{' :
                      rel.relationType === 'many_to_many' ? '}o--o{' : '||--||'
    code += `    ${rel.fromTable} ${relSymbol} ${rel.toTable} : "${rel.fromColumn}"\n`
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
  if (activePane.value === 'er') {
    nextTick(renderERDiagram)
  }
})

watch(activePane, (pane) => {
  if (pane === 'er' && !erDiagramSvg.value) {
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
      prefix: () => h('div', { class: 'i-carbon-data-table text-blue-500' }),
      suffix: () => h('div', { class: 'flex items-center gap-1' }, [
        table.hasContext && h('div', { class: 'i-carbon-magic-wand text-purple-500 text-xs' }),
        h('span', { class: 'text-xs text-gray-400' }, `${table.columns.length} 列`)
      ]),
      isLeaf: false,
      children: table.columns.map(col => ({
        key: `${table.name}.${col.name}`,
        label: col.name,
        prefix: () => h('div', { 
          class: col.isPrimaryKey 
            ? 'i-carbon-key text-yellow-500' 
            : col.isForeignKey 
              ? 'i-carbon-link text-green-500'
              : 'i-carbon-column text-gray-400'
        }),
        suffix: () => h('div', { class: 'flex items-center gap-1 text-xs text-gray-400' }, [
          h('span', {}, col.type),
          col.hasContext && h('div', { class: 'i-carbon-magic-wand text-purple-500' })
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
</script>

<template>
  <div class="schema-browser flex h-[calc(100vh-140px)]">
    <!-- Left: Table tree -->
    <div class="w-80 border-r border-gray-200 dark:border-gray-700 flex flex-col">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700">
        <NInput
          v-model:value="searchKeyword"
          placeholder="搜索表或列..."
          clearable
        >
          <template #prefix>
            <div class="i-carbon-search text-gray-400" />
          </template>
        </NInput>
      </div>

      <!-- Tabs: Tree / ER Diagram -->
      <NTabs v-model:value="activePane" type="segment" size="small" class="px-4 pt-2">
        <NTabPane name="tree" tab="表列表">
          <div class="overflow-auto" style="height: calc(100vh - 240px);">
            <NSpin v-if="workspaceStore.loadingSchema" class="mt-8" />
            <NEmpty v-else-if="treeData.length === 0" description="暂无数据" class="mt-8" />
            <NTree
              v-else
              :data="treeData"
              block-line
              selectable
              expand-on-click
              @update:selected-keys="handleSelect"
            />
          </div>
        </NTabPane>
        <NTabPane name="er" tab="ER 图">
          <div class="overflow-auto p-2" style="height: calc(100vh - 240px);">
            <NSpin v-if="!erDiagramSvg && !erError" class="mt-8" />
            <div v-else-if="erError" class="text-red-400 text-sm p-4">
              {{ erError }}
            </div>
            <div 
              v-else 
              class="er-diagram-container"
              v-html="erDiagramSvg"
            />
          </div>
        </NTabPane>
      </NTabs>
    </div>

    <!-- Right: Table detail -->
    <div class="flex-1 p-6 overflow-auto">
      <template v-if="selectedTable">
        <div class="mb-6">
          <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-100 mb-2">
            {{ selectedTable.name }}
          </h2>
          <p v-if="selectedTable.description" class="text-gray-500">
            {{ selectedTable.description }}
          </p>
          <div class="flex items-center gap-4 mt-2 text-sm text-gray-500">
            <span>{{ selectedTable.columns.length }} 列</span>
            <span v-if="selectedTable.rowCount">约 {{ selectedTable.rowCount }} 行</span>
            <NTag v-if="selectedTable.hasContext" type="success" size="small">
              <template #icon>
                <div class="i-carbon-magic-wand" />
              </template>
              有 Rich Context
            </NTag>
          </div>
        </div>

        <!-- Columns table -->
        <div class="mb-6">
          <h3 class="text-lg font-medium text-gray-800 dark:text-gray-100 mb-3">列信息</h3>
          <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-gray-800">
                <tr>
                  <th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">列名</th>
                  <th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">类型</th>
                  <th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">属性</th>
                  <th class="px-4 py-2 text-left font-medium text-gray-600 dark:text-gray-300">Context</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                <tr v-for="col in selectedTable.columns" :key="col.name" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td class="px-4 py-2">
                    <div class="flex items-center gap-2">
                      <div 
                        :class="col.isPrimaryKey 
                          ? 'i-carbon-key text-yellow-500' 
                          : col.isForeignKey 
                            ? 'i-carbon-link text-green-500'
                            : 'i-carbon-column text-gray-400'"
                      />
                      <span class="font-medium">{{ col.name }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-2 text-gray-500">{{ col.type }}</td>
                  <td class="px-4 py-2">
                    <div class="flex gap-1">
                      <NTag v-if="col.isPrimaryKey" size="tiny" type="warning">PK</NTag>
                      <NTag v-if="col.isForeignKey" size="tiny" type="success">FK</NTag>
                      <NTag v-if="col.isNullable === false" size="tiny">NOT NULL</NTag>
                    </div>
                  </td>
                  <td class="px-4 py-2">
                    <div v-if="col.hasContext" class="i-carbon-checkmark text-green-500" />
                    <div v-else class="i-carbon-subtract text-gray-300" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Related contexts -->
        <div v-if="tableContexts.length">
          <h3 class="text-lg font-medium text-gray-800 dark:text-gray-100 mb-3">
            相关 Context ({{ tableContexts.length }})
          </h3>
          <div class="space-y-2">
            <div 
              v-for="ctx in tableContexts" 
              :key="ctx.id"
              class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
            >
              <div class="flex items-center gap-2 mb-1">
                <NTag size="small" type="info">{{ ctx.type }}</NTag>
                <span v-if="ctx.columnName" class="text-sm text-gray-500">
                  {{ ctx.columnName }}
                </span>
              </div>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ ctx.content }}</p>
            </div>
          </div>
        </div>
      </template>

      <!-- No selection -->
      <div v-else class="h-full flex items-center justify-center">
        <NEmpty description="选择左侧的表查看详情" />
      </div>
    </div>
  </div>
</template>
