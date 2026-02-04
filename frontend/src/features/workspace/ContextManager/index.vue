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
  useMessage
} from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import type { RichContext, ContextType } from '@/types'
import GenerateContextConsole from './GenerateContextConsole.vue'

const workspaceStore = useWorkspaceStore()
const message = useMessage()

const searchKeyword = ref('')
const filterTable = ref<string | null>(null)
const filterType = ref<ContextType | null>(null)
const showGenerateConsole = ref(false)

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
  { label: '描述', value: 'description' },
  { label: '示例', value: 'example' },
  { label: '约束', value: 'constraint' },
  { label: '同义词', value: 'synonym' },
  { label: '值映射', value: 'value_mapping' },
  { label: '业务规则', value: 'business_rule' },
  { label: '计算规则', value: 'calculation' }
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
    message.warning('请填写完整信息')
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
    message.success('更新成功')
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
    message.success('添加成功')
  }

  showEditDialog.value = false
}

async function handleDelete(ctx: RichContext) {
  await workspaceStore.deleteContext(ctx.id)
  message.success('删除成功')
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

// Open generate console
function openGenerateConsole() {
  if (!workspaceStore.currentDatabaseId) {
    message.warning('请先选择数据库')
    return
  }
  showGenerateConsole.value = true
}

// Handle generation complete
async function handleGenerateComplete() {
  // Refresh contexts and schema
  await workspaceStore.fetchContexts()
  await workspaceStore.fetchSchema()
}
</script>

<template>
  <div class="context-manager p-6">
    <!-- Toolbar -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <NInput
          v-model:value="searchKeyword"
          placeholder="搜索 Context..."
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
          placeholder="筛选表"
          clearable
          style="width: 160px"
        />

        <NSelect
          v-model:value="filterType"
          :options="typeOptions"
          placeholder="筛选类型"
          clearable
          style="width: 140px"
        />
      </div>

      <div class="flex items-center gap-2">
        <NButton @click="workspaceStore.fetchContexts">
          <template #icon>
            <div class="i-carbon-refresh" />
          </template>
          刷新
        </NButton>
        <NButton 
          type="info" 
          @click="openGenerateConsole"
        >
          <template #icon>
            <div class="i-carbon-machine-learning-model" />
          </template>
          AI 自动生成
        </NButton>
        <NButton type="primary" @click="openCreateDialog">
          <template #icon>
            <div class="i-carbon-add" />
          </template>
          添加 Context
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
      description="暂无 Context"
      class="py-16"
    >
      <template #extra>
        <NButton type="primary" @click="openCreateDialog">
          添加第一条 Context
        </NButton>
      </template>
    </NEmpty>

    <!-- Structured Context List -->
    <div v-else class="context-tree">
      <!-- Expand/Collapse All -->
      <div class="flex gap-2 mb-4">
        <NButton size="small" quaternary @click="expandAll">
          <template #icon><div class="i-carbon-expand-all" /></template>
          展开全部
        </NButton>
        <NButton size="small" quaternary @click="collapseAll">
          <template #icon><div class="i-carbon-collapse-all" /></template>
          折叠全部
        </NButton>
        <span class="text-sm text-gray-500 ml-auto">
          {{ groupedContexts.length }} 个表，{{ filteredContexts.length }} 条 Context
        </span>
      </div>

      <!-- Table Groups -->
      <div 
        v-for="group in groupedContexts" 
        :key="group.tableName"
        class="table-group mb-4"
      >
        <!-- Table Header -->
        <div 
          class="table-header flex items-center gap-3 p-3 rounded-lg bg-gradient-to-r from-blue-500/10 to-cyan-500/10 border border-blue-500/20 cursor-pointer hover:from-blue-500/15 hover:to-cyan-500/15 transition-all"
          @click="toggleTable(group.tableName)"
        >
          <div 
            class="expand-icon transition-transform"
            :class="{ 'rotate-90': expandedTables.has(group.tableName) }"
          >
            <div class="i-carbon-chevron-right text-blue-400" />
          </div>
          
          <div class="i-carbon-data-table text-lg text-blue-400" />
          
          <span class="font-semibold text-blue-300">{{ group.tableName }}</span>
          
          <NTag size="small" :bordered="false" class="ml-2">
            {{ group.columnContexts.length + (group.tableContext ? 1 : 0) }} contexts
          </NTag>

          <div class="ml-auto flex items-center gap-2">
            <NButton 
              size="tiny" 
              quaternary 
              @click.stop="openCreateDialogForTable(group.tableName)"
            >
              <div class="i-carbon-add text-xs" />
            </NButton>
          </div>
        </div>

        <!-- Expanded Content -->
        <div 
          v-if="expandedTables.has(group.tableName)"
          class="table-content ml-6 mt-2 border-l-2 border-blue-500/20 pl-4"
        >
          <!-- Table-level Context -->
          <div 
            v-if="group.tableContext" 
            class="context-item p-3 mb-2 rounded-lg bg-yellow-500/5 border border-yellow-500/20"
          >
            <div class="flex items-center gap-2 mb-2">
              <div class="i-carbon-document text-yellow-400" />
              <span class="text-sm font-medium text-yellow-300">Table Description</span>
              <NTag size="tiny" type="warning">{{ group.tableContext.type }}</NTag>
              <div class="ml-auto flex gap-1">
                <NButton size="tiny" quaternary @click="openEditDialog(group.tableContext!)">
                  <div class="i-carbon-edit text-xs" />
                </NButton>
                <NButton size="tiny" quaternary type="error" @click="handleDelete(group.tableContext!)">
                  <div class="i-carbon-trash-can text-xs" />
                </NButton>
              </div>
            </div>
            <p class="text-sm text-gray-300 leading-relaxed">{{ group.tableContext.content }}</p>
          </div>

          <!-- Column Contexts -->
          <div 
            v-for="colCtx in group.columnContexts" 
            :key="colCtx.id"
            class="context-item p-3 mb-2 rounded-lg bg-emerald-500/5 border border-emerald-500/20"
          >
            <div class="flex items-center gap-2 mb-2">
              <div class="i-carbon-column text-emerald-400" />
              <span class="text-sm font-medium text-emerald-300">{{ colCtx.columnName }}</span>
              <NTag size="tiny" type="success">{{ colCtx.type }}</NTag>
              <div class="ml-auto flex gap-1">
                <NButton size="tiny" quaternary @click="openEditDialog(colCtx)">
                  <div class="i-carbon-edit text-xs" />
                </NButton>
                <NButton size="tiny" quaternary type="error" @click="handleDelete(colCtx)">
                  <div class="i-carbon-trash-can text-xs" />
                </NButton>
              </div>
            </div>
            <p class="text-sm text-gray-300 leading-relaxed">{{ colCtx.content }}</p>
          </div>

          <!-- Empty columns hint -->
          <div 
            v-if="group.columnContexts.length === 0 && !group.tableContext"
            class="text-sm text-gray-500 py-2"
          >
            暂无 Context
          </div>
        </div>
      </div>
    </div>

    <!-- Edit/Create Dialog -->
    <NModal
      v-model:show="showEditDialog"
      preset="card"
      :title="editingContext ? '编辑 Context' : '添加 Context'"
      style="width: 500px"
    >
      <NForm :model="editForm" label-placement="left" label-width="80">
        <NFormItem label="表名" required>
          <NSelect
            v-model:value="editForm.tableName"
            :options="tableOptions"
            placeholder="选择表"
          />
        </NFormItem>
        <NFormItem label="列名">
          <NInput
            v-model:value="editForm.columnName"
            placeholder="可选，不填则为表级 Context"
          />
        </NFormItem>
        <NFormItem label="类型" required>
          <NSelect
            v-model:value="editForm.type"
            :options="typeOptions"
          />
        </NFormItem>
        <NFormItem label="内容" required>
          <NInput
            v-model:value="editForm.content"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="输入 Context 内容..."
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditDialog = false">取消</NButton>
          <NButton type="primary" @click="handleSave">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- Generate Context Console -->
    <GenerateContextConsole
      v-model:show="showGenerateConsole"
      :database-id="workspaceStore.currentDatabaseId || ''"
      @complete="handleGenerateComplete"
    />
  </div>
</template>
