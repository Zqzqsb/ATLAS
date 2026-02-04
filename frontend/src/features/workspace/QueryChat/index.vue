<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { NScrollbar, NEmpty, NButton } from 'naive-ui'
import { useWorkspaceStore } from '@/stores/workspace'
import ChatInput from './ChatInput.vue'
import ChatMessage from './ChatMessage.vue'

const workspaceStore = useWorkspaceStore()

const chatContainerRef = ref<HTMLElement>()

// Auto scroll to bottom on new message
watch(
  () => [workspaceStore.reactSteps.length, workspaceStore.generatedSql],
  async () => {
    await nextTick()
    scrollToBottom()
  }
)

function scrollToBottom() {
  if (chatContainerRef.value) {
    chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
  }
}

// Build message list from current query and history
const hasCurrentQuery = computed(() => 
  workspaceStore.currentQuestion || 
  workspaceStore.isQuerying || 
  workspaceStore.generatedSql
)

function handleHistoryClick(record: any) {
  workspaceStore.currentQuestion = record.question
}
</script>

<template>
  <div class="query-chat flex flex-col h-[calc(100vh-140px)]">
    <!-- Chat messages area -->
    <div 
      ref="chatContainerRef"
      class="flex-1 overflow-auto p-6"
    >
      <!-- Empty state -->
      <div v-if="!hasCurrentQuery && workspaceStore.queryHistory.length === 0" class="h-full flex items-center justify-center">
        <div class="text-center max-w-md">
          <div class="w-20 h-20 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center mx-auto mb-4">
            <div class="i-carbon-chat text-4xl text-blue-500" />
          </div>
          <h3 class="text-lg font-medium text-gray-800 dark:text-gray-200 mb-2">
            开始你的查询
          </h3>
          <p class="text-gray-500 mb-4">
            用自然语言描述你想查询的内容，AI 会帮你生成 SQL
          </p>
          <div class="space-y-2 text-left bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <p class="text-sm text-gray-600 dark:text-gray-400">示例问题：</p>
            <ul class="space-y-1 text-sm">
              <li 
                class="text-blue-500 cursor-pointer hover:underline"
                @click="workspaceStore.currentQuestion = '查询VIP客户的总订单金额'"
              >
                查询VIP客户的总订单金额
              </li>
              <li 
                class="text-blue-500 cursor-pointer hover:underline"
                @click="workspaceStore.currentQuestion = '统计各等级客户数量'"
              >
                统计各等级客户数量
              </li>
              <li 
                class="text-blue-500 cursor-pointer hover:underline"
                @click="workspaceStore.currentQuestion = '查询最近7天的有效订单'"
              >
                查询最近7天的有效订单
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- History messages -->
      <template v-else>
        <!-- Previous queries from history -->
        <template v-for="record in workspaceStore.queryHistory.slice().reverse()" :key="record.id">
          <ChatMessage
            type="user"
            :question="record.question"
          />
          <ChatMessage
            type="assistant"
            :sql="record.sql"
            :used-contexts="record.usedContexts"
            :duration="record.duration * 1000"
          />
        </template>

        <!-- Current query (if different from last history item) -->
        <template v-if="workspaceStore.currentQuestion && (workspaceStore.isQuerying || workspaceStore.generatedSql)">
          <ChatMessage
            type="user"
            :question="workspaceStore.currentQuestion"
          />
          <ChatMessage
            type="assistant"
            :sql="workspaceStore.generatedSql"
            :react-steps="workspaceStore.reactSteps"
            :used-contexts="workspaceStore.usedContexts"
            :grounding-result="workspaceStore.groundingResult"
            :grounding-stage="workspaceStore.groundingStage"
            :loading="workspaceStore.isQuerying"
            :error="workspaceStore.queryError"
            :duration="workspaceStore.queryDuration"
          />
        </template>
      </template>
    </div>

    <!-- Input area -->
    <ChatInput />
  </div>
</template>
