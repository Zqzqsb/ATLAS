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

    <!-- Context list -->
    <div v-else class="grid gap-4">
      <NCard
        v-for="ctx in filteredContexts"
        :key="ctx.id"
        size="small"
        hoverable
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-2">
              <NTag size="small">{{ ctx.tableName }}</NTag>
              <NTag v-if="ctx.columnName" size="small" type="info">
                {{ ctx.columnName }}
              </NTag>
              <NTag size="small" :type="getTypeColor(ctx.type) as any">
                {{ ctx.type }}
              </NTag>
              <NTag v-if="ctx.source === 'auto'" size="small" :bordered="false">
                <template #icon>
                  <div class="i-carbon-machine-learning" />
                </template>
                自动生成
              </NTag>
              <NTag v-else-if="ctx.source === 'feedback'" size="small" type="success" :bordered="false">
                <template #icon>
                  <div class="i-carbon-user-feedback" />
                </template>
                用户反馈
              </NTag>
            </div>
            <p class="text-gray-700 dark:text-gray-300">{{ ctx.content }}</p>
            <div class="flex items-center gap-4 mt-2 text-xs text-gray-400">
              <span v-if="ctx.usageCount">使用 {{ ctx.usageCount }} 次</span>
              <span v-if="ctx.confidence">置信度 {{ (ctx.confidence * 100).toFixed(0) }}%</span>
              <span>{{ new Date(ctx.createdAt).toLocaleDateString() }}</span>
            </div>
          </div>
          
          <div class="flex items-center gap-1 ml-4">
            <NButton quaternary size="small" @click="openEditDialog(ctx)">
              <div class="i-carbon-edit" />
            </NButton>
            <NButton quaternary size="small" type="error" @click="handleDelete(ctx)">
              <div class="i-carbon-trash-can" />
            </NButton>
          </div>
        </div>
      </NCard>
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
