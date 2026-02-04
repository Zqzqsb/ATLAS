<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { NInput, NButton, NDropdown, NTooltip } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'

const workspaceStore = useWorkspaceStore()

const inputRef = ref<HTMLTextAreaElement>()

const optionMenuItems = [
  {
    label: '使用 Rich Context',
    key: 'useRichContext',
    icon: () => h('div', { class: workspaceStore.queryOptions.useRichContext ? 'i-carbon-checkmark text-green-500' : 'i-carbon-close' })
  },
  {
    label: '使用 ReAct 推理',
    key: 'useReact',
    icon: () => h('div', { class: workspaceStore.queryOptions.useReact ? 'i-carbon-checkmark text-green-500' : 'i-carbon-close' })
  },
  {
    label: '使用 Grounding',
    key: 'useGrounding',
    icon: () => h('div', { class: workspaceStore.queryOptions.useGrounding ? 'i-carbon-checkmark text-green-500' : 'i-carbon-close' })
  }
]

import { h } from 'vue'

function handleOptionSelect(key: string) {
  const options = workspaceStore.queryOptions as any
  options[key] = !options[key]
}

function handleSend() {
  if (!workspaceStore.currentQuestion.trim() || workspaceStore.isQuerying) return
  workspaceStore.executeQuery()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}
</script>

<template>
  <div class="chat-input bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 p-4">
    <div class="flex items-end gap-3">
      <!-- Input area -->
      <div class="flex-1 relative">
        <NInput
          ref="inputRef"
          v-model:value="workspaceStore.currentQuestion"
          type="textarea"
          :autosize="{ minRows: 1, maxRows: 4 }"
          placeholder="输入你的问题，例如：查询VIP客户的总订单金额..."
          :disabled="workspaceStore.isQuerying"
          @keydown="handleKeydown"
        />
      </div>

      <!-- Options dropdown -->
      <NDropdown
        trigger="click"
        :options="optionMenuItems"
        @select="handleOptionSelect"
      >
        <NTooltip>
          <template #trigger>
            <NButton quaternary circle :disabled="workspaceStore.isQuerying">
              <div class="i-carbon-settings-adjust text-lg" />
            </NButton>
          </template>
          查询选项
        </NTooltip>
      </NDropdown>

      <!-- Send button -->
      <NButton
        type="primary"
        :loading="workspaceStore.isQuerying"
        :disabled="!workspaceStore.currentQuestion.trim()"
        @click="handleSend"
      >
        <template #icon>
          <div class="i-carbon-send" />
        </template>
        发送
      </NButton>

      <!-- Stop button (when querying) -->
      <NButton
        v-if="workspaceStore.isQuerying"
        type="error"
        secondary
        @click="workspaceStore.abortCurrentQuery"
      >
        <template #icon>
          <div class="i-carbon-stop" />
        </template>
        停止
      </NButton>
    </div>

    <!-- Quick options display -->
    <div class="flex items-center gap-2 mt-2 text-xs text-gray-500">
      <span v-if="workspaceStore.queryOptions.useRichContext" class="flex items-center gap-1">
        <div class="i-carbon-magic-wand text-blue-500" />
        Rich Context
      </span>
      <span v-if="workspaceStore.queryOptions.useReact" class="flex items-center gap-1">
        <div class="i-carbon-flow text-purple-500" />
        ReAct
      </span>
      <span v-if="workspaceStore.queryOptions.useGrounding" class="flex items-center gap-1">
        <div class="i-carbon-connect text-green-500" />
        Grounding
      </span>
    </div>
  </div>
</template>
