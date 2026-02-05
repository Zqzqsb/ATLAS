<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDemoStore } from '@/stores/demo'
import ScenarioTabs from '@/components/demo/ScenarioTabs.vue'
import ContextComparison from '@/components/demo/ContextComparison.vue'
import SelfMaintainDemo from '@/components/demo/SelfMaintainDemo.vue'
import BenchmarkDemo from '@/components/demo/BenchmarkDemo.vue'
import LakebaseArchitecture from '@/components/demo/LakebaseArchitecture.vue'

const store = useDemoStore()

onMounted(() => {
  store.loadComparisonCases()
  store.loadMaintenanceLogs()
})
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Page Title -->
    <div class="bg-white border-b border-gray-200 px-6 py-6">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">Live Demo</h1>
          <p class="text-sm text-gray-500 mt-1">Interactive demonstration of LUCID capabilities</p>
        </div>
        <div class="flex items-center gap-4 text-sm">
          <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg border border-gray-200">
            <span class="i-carbon-cube text-primary-600" />
            <span class="font-semibold text-gray-700">MariaDB 12</span>
          </div>
          <div class="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg border border-gray-200">
            <span class="i-carbon-model-alt text-purple-600" />
            <span class="font-semibold text-gray-700">GPT-4o</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-6 py-8">
      <!-- Scenario Tabs -->
      <ScenarioTabs 
        v-model="store.currentScenario" 
        class="mb-8"
      />

      <!-- Content Grid -->
      <div class="grid grid-cols-12 gap-8">
        <!-- Left: Main Demo Area -->
        <div class="col-span-8">
          <!-- Context Comparison (Core) -->
          <ContextComparison 
            v-if="store.currentScenario === 'comparison'"
          />

          <!-- Self-Maintain Demo -->
          <SelfMaintainDemo 
            v-else-if="store.currentScenario === 'selfmaintain'"
          />

          <!-- Benchmark Demo -->
          <BenchmarkDemo 
            v-else
          />
        </div>

        <!-- Right: Architecture & Info -->
        <div class="col-span-4">
          <LakebaseArchitecture />
        </div>
      </div>
    </main>
  </div>
</template>
