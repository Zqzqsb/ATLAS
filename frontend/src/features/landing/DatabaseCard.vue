<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NButton, NTag, NTooltip, NBadge } from 'naive-ui'
import type { Database } from '@/types'

const props = defineProps<{
  database: Database
}>()

const emit = defineEmits<{
  enter: [id: string]
  test: [id: string]
}>()

const router = useRouter()

const statusColor = computed(() => {
  switch (props.database.status) {
    case 'connected': return 'success'
    case 'disconnected': return 'warning'
    case 'error': return 'error'
    default: return 'default'
  }
})

const statusText = computed(() => {
  switch (props.database.status) {
    case 'connected': return '已连接'
    case 'disconnected': return '未连接'
    case 'error': return '连接错误'
    default: return '未知'
  }
})

const typeIcon = computed(() => {
  switch (props.database.type) {
    case 'mariadb': return 'i-logos-mariadb-icon'
    case 'mysql': return 'i-logos-mysql'
    case 'postgresql': return 'i-logos-postgresql'
    case 'sqlite': return 'i-simple-icons-sqlite'
    default: return 'i-carbon-data-base'
  }
})

function handleEnter() {
  if (props.database.status === 'connected') {
    router.push(`/workspace/${props.database.id}`)
  }
}

function handleTest() {
  emit('test', props.database.id)
}
</script>

<template>
  <NCard 
    class="database-card h-full cursor-pointer transition-all duration-200 hover:shadow-lg hover:border-blue-300 dark:hover:border-blue-600"
    :class="{ 'opacity-60': database.status !== 'connected' }"
    hoverable
    @click="handleEnter"
  >
    <div class="flex flex-col h-full">
      <!-- Header -->
      <div class="flex items-start justify-between mb-3">
        <div class="flex items-center gap-3">
          <!-- Database type icon -->
          <div 
            class="w-12 h-12 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center"
          >
            <div :class="typeIcon" class="text-2xl" />
          </div>
          
          <div>
            <h3 class="font-semibold text-lg text-gray-800 dark:text-gray-100">
              {{ database.displayName || database.name }}
            </h3>
            <p class="text-sm text-gray-500">
              {{ database.type.toUpperCase() }}
              <span v-if="database.host">· {{ database.host }}</span>
            </p>
          </div>
        </div>

        <!-- Status badge -->
        <NTag :type="statusColor" size="small" round>
          <template #icon>
            <div 
              class="w-2 h-2 rounded-full mr-1"
              :class="{
                'bg-green-500': database.status === 'connected',
                'bg-yellow-500': database.status === 'disconnected',
                'bg-red-500': database.status === 'error'
              }"
            />
          </template>
          {{ statusText }}
        </NTag>
      </div>

      <!-- Description -->
      <p v-if="database.description" class="text-sm text-gray-600 dark:text-gray-400 mb-3 line-clamp-2">
        {{ database.description }}
      </p>

      <!-- Stats -->
      <div class="flex items-center gap-4 text-sm text-gray-500 mb-4">
        <div class="flex items-center gap-1">
          <div class="i-carbon-data-table" />
          <span>{{ database.tableCount }} 张表</span>
        </div>
        
        <NTooltip v-if="database.hasRichContext">
          <template #trigger>
            <div class="flex items-center gap-1 text-blue-500">
              <div class="i-carbon-magic-wand" />
              <span>{{ database.contextCount }} 条 Context</span>
            </div>
          </template>
          富上下文已启用
        </NTooltip>
        
        <div v-else class="flex items-center gap-1 text-yellow-600">
          <div class="i-carbon-warning" />
          <span>待配置</span>
        </div>
      </div>

      <!-- Tags -->
      <div v-if="database.tags?.length" class="flex flex-wrap gap-1 mb-4">
        <NTag v-for="tag in database.tags" :key="tag" size="tiny" :bordered="false">
          {{ tag }}
        </NTag>
      </div>

      <!-- Actions -->
      <div class="mt-auto flex items-center gap-2">
        <NButton 
          type="primary" 
          :disabled="database.status !== 'connected'"
          class="flex-1"
          @click.stop="handleEnter"
        >
          <template #icon>
            <div class="i-carbon-arrow-right" />
          </template>
          进入工作区
        </NButton>
        
        <NTooltip>
          <template #trigger>
            <NButton 
              quaternary 
              circle
              @click.stop="handleTest"
            >
              <div class="i-carbon-connection-signal" />
            </NButton>
          </template>
          测试连接
        </NTooltip>
      </div>
    </div>
  </NCard>
</template>

<style scoped>
.database-card {
  min-height: 240px;
}
</style>
