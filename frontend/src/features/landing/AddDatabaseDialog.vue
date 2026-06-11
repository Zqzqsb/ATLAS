<script setup lang="ts">
import { ref, watch } from 'vue'
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

const defaultForm = (): DatabaseConfig => ({
  name: '',
  type: 'mariadb',
  host: 'localhost',
  port: 3306,
  username: '',
  password: '',
  database: ''
})

const formData = ref<DatabaseConfig>(defaultForm())

const typeOptions = [
  { label: 'MariaDB', value: 'mariadb' },
  { label: 'MySQL', value: 'mysql' },
  { label: 'PostgreSQL', value: 'postgresql' },
  { label: 'SQLite', value: 'sqlite' }
]

// Preset connections
const presetOptions = [
  { label: 'Custom (Manual)', value: 'custom' },
  { label: 'Evolution Demo DB', value: 'lucid_evolution' }
]

const selectedPreset = ref('custom')

function applyPreset(preset: string) {
  if (preset === 'lucid_evolution') {
    formData.value = {
      name: 'lucid_evolution',
      type: 'mariadb',
      host: 'lucid-mariadb',
      port: 3306,
      username: 'root',
      password: '',
      database: 'lucid_evolution'
    }
  } else {
    formData.value = defaultForm()
  }
}

const rules = {
  name: { required: true, message: 'Connection name is required', trigger: ['blur', 'input'] },
  type: { required: true, message: 'Database type is required', trigger: 'change' },
  host: { required: true, message: 'Host address is required', trigger: ['blur', 'input'] },
  database: { required: true, message: 'Database name is required', trigger: ['blur', 'input'] }
}

// Reset form when dialog opens
watch(() => props.show, (val) => {
  if (val) {
    selectedPreset.value = 'custom'
    formData.value = defaultForm()
    // Clear validation on next tick
    setTimeout(() => formRef.value?.restoreValidation(), 0)
  }
})

function handleClose() {
  emit('update:show', false)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch (e) {
    return // Validation failed
  }
  // Emit and let parent handle the async request.
  // Parent will close the dialog via v-model:show when done.
  loading.value = true
  emit('submit', { ...formData.value })
}

function updatePort() {
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
    title="Add Database Connection"
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
      <!-- Preset selector -->
      <NFormItem label="Quick Select" :show-feedback="false">
        <NSelect
          v-model:value="selectedPreset"
          :options="presetOptions"
          @update:value="applyPreset"
        />
      </NFormItem>

      <NFormItem label="Name" path="name">
        <NInput 
          v-model:value="formData.name" 
          placeholder="Give this connection a name"
        />
      </NFormItem>

      <NFormItem label="Type" path="type">
        <NSelect
          v-model:value="formData.type"
          :options="typeOptions"
          @update:value="updatePort"
        />
      </NFormItem>

      <template v-if="formData.type !== 'sqlite'">
        <NFormItem label="Host" path="host">
          <NInput 
            v-model:value="formData.host" 
            placeholder="localhost"
          />
        </NFormItem>

        <NFormItem label="Port" path="port">
          <NInputNumber
            v-model:value="formData.port"
            :min="1"
            :max="65535"
            placeholder="3306"
            class="w-full"
          />
        </NFormItem>

        <NFormItem label="Username" path="username">
          <NInput 
            v-model:value="formData.username" 
            placeholder="root"
          />
        </NFormItem>

        <NFormItem label="Password" path="password">
          <NInput
            v-model:value="formData.password"
            type="password"
            show-password-on="click"
            placeholder="Enter password"
          />
        </NFormItem>
      </template>

      <template v-else>
        <NFormItem label="File Path" path="path">
          <NInput 
            v-model:value="formData.path" 
            placeholder="/path/to/database.db"
          />
        </NFormItem>
      </template>

      <NFormItem label="Database" path="database">
        <NInput 
          v-model:value="formData.database" 
          placeholder="Enter database name"
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <NSpace justify="end">
        <NButton @click="handleClose">Cancel</NButton>
        <NButton type="primary" :loading="loading" @click="handleSubmit">
          Add Connection
        </NButton>
      </NSpace>
    </template>
  </NModal>
</template>
