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
    <!-- Header -->
    <header class="bg-white border-b border-gray-200 sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 py-4 flex-between">
        <div class="flex items-center gap-3">
          <div class="i-carbon-data-base text-2xl text-blue-500" />
          <h1 class="text-xl font-bold text-gray-900">LUCID</h1>
          <span class="text-sm text-gray-500">湖基原生 Text-to-SQL</span>
        </div>
        <div class="flex items-center gap-4 text-sm text-gray-600">
          <span class="flex items-center gap-1">
            <span class="i-carbon-cube text-green-500" />
            MariaDB 11.7
          </span>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 py-6">
      <!-- Scenario Tabs -->
      <ScenarioTabs 
        v-model="store.currentScenario" 
        class="mb-6"
      />

      <!-- Content Grid -->
      <div class="grid grid-cols-12 gap-6">
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
