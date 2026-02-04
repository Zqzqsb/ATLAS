<script setup lang="ts">
import { ref } from 'vue'
import { 
  NModal, 
  NForm, 
  NFormItem, 
  NInput, 
  NSelect, 
  NInputNumber, 
  NButton,
  NSpace,
  useMessage
} from 'naive-ui'
import type { DatabaseConfig } from '@/types'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  submit: [config: DatabaseConfig]
}>()

const message = useMessage()

const formRef = ref()
const loading = ref(false)

const formData = ref<DatabaseConfig>({
  name: '',
  type: 'mariadb',
  host: 'localhost',
  port: 3306,
  username: '',
  password: '',
  database: ''
})

const typeOptions = [
  { label: 'MariaDB', value: 'mariadb' },
  { label: 'MySQL', value: 'mysql' },
  { label: 'PostgreSQL', value: 'postgresql' },
  { label: 'SQLite', value: 'sqlite' }
]

const rules = {
  name: { required: true, message: '请输入连接名称', trigger: 'blur' },
  type: { required: true, message: '请选择数据库类型', trigger: 'change' },
  host: { required: true, message: '请输入主机地址', trigger: 'blur' },
  database: { required: true, message: '请输入数据库名', trigger: 'blur' }
}

function handleClose() {
  emit('update:show', false)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
    loading.value = true
    
    emit('submit', { ...formData.value })
    
    // Reset form
    formData.value = {
      name: '',
      type: 'mariadb',
      host: 'localhost',
      port: 3306,
      username: '',
      password: '',
      database: ''
    }
    
    handleClose()
  } catch (e) {
    // Validation failed
  } finally {
    loading.value = false
  }
}

function updatePort() {
  // Set default port based on database type
  switch (formData.value.type) {
    case 'mariadb':
    case 'mysql':
      formData.value.port = 3306
      break
    case 'postgresql':
      formData.value.port = 5432
      break
    case 'sqlite':
      formData.value.port = undefined
      break
  }
}
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    title="添加数据库连接"
    style="width: 500px"
    :mask-closable="false"
    @update:show="$emit('update:show', $event)"
  >
    <NForm
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-placement="left"
      label-width="100"
    >
      <NFormItem label="连接名称" path="name">
        <NInput 
          v-model:value="formData.name" 
          placeholder="给这个连接起个名字"
        />
      </NFormItem>

      <NFormItem label="数据库类型" path="type">
        <NSelect
          v-model:value="formData.type"
          :options="typeOptions"
          @update:value="updatePort"
        />
      </NFormItem>

      <template v-if="formData.type !== 'sqlite'">
        <NFormItem label="主机地址" path="host">
          <NInput 
            v-model:value="formData.host" 
            placeholder="localhost"
          />
        </NFormItem>

        <NFormItem label="端口" path="port">
          <NInputNumber
            v-model:value="formData.port"
            :min="1"
            :max="65535"
            placeholder="3306"
            class="w-full"
          />
        </NFormItem>

        <NFormItem label="用户名" path="username">
          <NInput 
            v-model:value="formData.username" 
            placeholder="root"
          />
        </NFormItem>

        <NFormItem label="密码" path="password">
          <NInput
            v-model:value="formData.password"
            type="password"
            show-password-on="click"
            placeholder="请输入密码"
          />
        </NFormItem>
      </template>

      <template v-else>
        <NFormItem label="文件路径" path="path">
          <NInput 
            v-model:value="formData.path" 
            placeholder="/path/to/database.db"
          />
        </NFormItem>
      </template>

      <NFormItem label="数据库名" path="database">
        <NInput 
          v-model:value="formData.database" 
          placeholder="请输入数据库名"
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">取消</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          添加连接
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
