<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NTabs, NTab, NTabPane } from 'naive-ui'
import { useDemoStore } from '@/stores/demo'

// Demo components
import ArchitectureShowcase from './ArchitectureShowcase.vue'
import ContextComparison from './ContextComparison.vue'
import SelfMaintainDemo from './SelfMaintainDemo.vue'

const demoStore = useDemoStore()

type DemoTab = 'architecture' | 'comparison' | 'selfmaintain'
const activeTab = ref<DemoTab>('architecture')

const tabs: { key: DemoTab; label: string; icon: string }[] = [
  { key: 'architecture', label: '架构展示', icon: 'i-carbon-diagram' },
  { key: 'comparison', label: 'Context 对比', icon: 'i-carbon-compare' },
  { key: 'selfmaintain', label: '自维持演示', icon: 'i-carbon-recycle' }
]

onMounted(() => {
  demoStore.loadComparisonCases()
  demoStore.loadMaintenanceLogs()
})
</script>

<template>
  <div class="demo-page min-h-screen bg-gray-50 dark:bg-gray-950">
    <!-- Hero header -->
    <div class="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-12 px-6">
      <div class="max-w-6xl mx-auto">
        <h1 class="text-3xl font-bold mb-2">LUCID Demo</h1>
        <p class="text-blue-100 text-lg">
          湖仓一体的智能 Text-to-SQL 系统核心能力演示
        </p>
      </div>
    </div>

    <!-- Tab navigation -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 sticky top-0 z-10">
      <div class="max-w-6xl mx-auto px-6">
        <NTabs
          v-model:value="activeTab"
          type="line"
          animated
        >
          <NTab
            v-for="tab in tabs"
            :key="tab.key"
            :name="tab.key"
          >
            <div class="flex items-center gap-2">
              <div :class="tab.icon" />
              <span>{{ tab.label }}</span>
            </div>
          </NTab>
        </NTabs>
      </div>
    </div>

    <!-- Tab content -->
    <div class="max-w-6xl mx-auto px-6 py-8">
      <ArchitectureShowcase v-if="activeTab === 'architecture'" />
      <ContextComparison v-else-if="activeTab === 'comparison'" />
      <SelfMaintainDemo v-else-if="activeTab === 'selfmaintain'" />
    </div>
  </div>
</template>
